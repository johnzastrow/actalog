package middleware

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/security"
)

const (
	// ProtectedUserErrorCode is the JSON error code returned in 403 responses
	// when a request targets a protected user. Stable contract — frontend
	// axios interceptor (Task 18) keys off this value.
	ProtectedUserErrorCode = "protected_user"

	// ProtectedUserMessage is the user-facing message returned in 403 responses.
	ProtectedUserMessage = "This account is system-reserved and cannot be modified."

	// ProtectedUserDocumentationURL points at the runbook (full content lands in Task 24).
	ProtectedUserDocumentationURL = "/docs/protected-users"
)

// AuditLogger is the minimal interface ProtectedUserGuard depends on.
// It matches the LogEvent method signature of service.AuditLogService exactly,
// allowing the concrete service to satisfy this interface without adaptation.
type AuditLogger interface {
	LogEvent(
		eventType string,
		userID *int64,
		targetUserID *int64,
		ipAddress *string,
		userAgent *string,
		details map[string]interface{},
	) error
}

// GuardLogger is the minimal logger interface ProtectedUserGuard depends on.
// Matches a method on the project's *logger.Logger (e.g., Error(format string, v ...interface{})).
// Named GuardLogger (not Logger) to avoid a name collision with the deprecated
// Logger middleware function already declared in cors.go.
type GuardLogger interface {
	Error(format string, v ...interface{})
}

// ProtectedUserGuard is an HTTP middleware that intercepts non-GET/HEAD requests
// under a route containing the chi URL parameter {id}, and rejects them with
// HTTP 403 if the target user's email is in the protected-user registry.
//
// Behaviour:
//   - GET and HEAD always pass through unchanged (reads are permitted).
//   - Non-integer {id} values → 404 (bad request, not a protected-user event).
//   - Unknown user IDs (repository returns nil, nil) → 404.
//   - Repository errors → 500 (the audit system is not notified; the error is
//     logged via the structured logger so it is visible in application logs).
//   - Protected user targeted by a write → 403 with structured JSON body and
//     an EventProtectedUserAttackHTTP audit event (L1 layer).
//   - Non-protected user → pass through to the next handler.
//
// Audit-write failures do not surface as 500 to the client; the guard logs the
// error via the structured logger and still returns 403 so the attacker receives
// no information about audit-system health.
func ProtectedUserGuard(userRepo domain.UserRepository, audit AuditLogger, log GuardLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// GET and HEAD are safe reads — always pass through.
			if r.Method == http.MethodGet || r.Method == http.MethodHead {
				next.ServeHTTP(w, r)
				return
			}

			// Parse the {id} URL param; a non-integer is a 404, not a 403.
			idStr := chi.URLParam(r, "id")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			// Look up the target user. The repository contract returns (nil, nil)
			// when the user does not exist — that is a 404.
			// Also treat sql.ErrNoRows as not-found for forward-compatibility with
			// future repository refactors that may return that sentinel instead of
			// the (nil, nil) convention.
			user, err := userRepo.GetByID(id)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				log.Error("ProtectedUserGuard: GetByID(%d) error: %v", id, err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if user == nil {
				http.NotFound(w, r)
				return
			}

			// Check whether this user is in the protected-email registry.
			if security.IsProtectedEmail(user.Email) {
				// Capture actor ID from context only when authentication context is
				// present. If the bool is false (unauthenticated or middleware ordering
				// bug), keep actorPtr nil so the audit log records nil rather than 0.
				var actorPtr *int64
				if actorID, ok := GetUserID(r.Context()); ok {
					actorPtr = &actorID
				}

				ua := r.UserAgent()
				ip := r.RemoteAddr

				// Use json.Marshal for the details map to ensure proper escaping of
				// user-controlled values (user-agent, referer, path).
				details := map[string]interface{}{
					"path":       r.URL.Path,
					"method":     r.Method,
					"user_agent": ua,
					"referer":    r.Referer(),
				}

				if auditErr := audit.LogEvent(
					domain.EventProtectedUserAttackHTTP,
					actorPtr,
					&id,
					&ip,
					&ua,
					details,
				); auditErr != nil {
					// Log the failure but do not expose it to the caller.
					log.Error("ProtectedUserGuard: audit.LogEvent error: %v", auditErr)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)

				resp := map[string]string{
					"error":             ProtectedUserErrorCode,
					"message":           ProtectedUserMessage,
					"documentation_url": ProtectedUserDocumentationURL,
				}
				if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
					log.Error("ProtectedUserGuard: json encode error: %v", encErr)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
