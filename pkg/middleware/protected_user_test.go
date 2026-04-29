package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
)

// errNoRows is sql.ErrNoRows, used to test forward-compatible not-found handling.
var errNoRows = sql.ErrNoRows

// stubUserRepo implements domain.UserRepository for tests.
// Only GetByID is used by ProtectedUserGuard; other methods panic.
type stubUserRepo struct {
	users map[int64]*domain.User
	err   error // if non-nil, GetByID returns this error
}

func (s *stubUserRepo) GetByID(id int64) (*domain.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	u, ok := s.users[id]
	if !ok {
		return nil, nil // repo contract: (nil, nil) == not found
	}
	return u, nil
}

// Stub out unused interface methods.
func (s *stubUserRepo) Create(u *domain.User) error                            { panic("unused") }
func (s *stubUserRepo) GetByEmail(e string) (*domain.User, error)              { panic("unused") }
func (s *stubUserRepo) GetByResetToken(t string) (*domain.User, error)         { panic("unused") }
func (s *stubUserRepo) GetByVerificationToken(t string) (*domain.User, error)  { panic("unused") }
func (s *stubUserRepo) Update(u *domain.User) error                            { panic("unused") }
func (s *stubUserRepo) UpdatePassword(id int64, hash string) error             { panic("unused") }
func (s *stubUserRepo) Delete(id int64) error                                  { panic("unused") }
func (s *stubUserRepo) List(limit, offset int) ([]*domain.User, error)         { panic("unused") }
func (s *stubUserRepo) Count() (int64, error)                                  { panic("unused") }
func (s *stubUserRepo) ListAdmins() ([]*domain.User, error)                    { panic("unused") }
func (s *stubUserRepo) IncrementFailedAttempts(id int64) error                 { panic("unused") }
func (s *stubUserRepo) ResetFailedAttempts(id int64) error                     { panic("unused") }
func (s *stubUserRepo) LockAccount(id int64, d time.Duration) error            { panic("unused") }
func (s *stubUserRepo) UnlockAccount(id int64) error                           { panic("unused") }
func (s *stubUserRepo) IsAccountLocked(id int64) (bool, *time.Time, error)     { panic("unused") }
func (s *stubUserRepo) DisableAccount(id int64, by int64, reason string) error { panic("unused") }
func (s *stubUserRepo) EnableAccount(id int64) error                           { panic("unused") }
func (s *stubUserRepo) CountNewThisMonth() (int64, error)                      { panic("unused") }
func (s *stubUserRepo) CountDisabled() (int64, error)                          { panic("unused") }
func (s *stubUserRepo) ListWithFilter(f domain.UserListFilter, l, o int) ([]*domain.User, error) {
	panic("unused")
}
func (s *stubUserRepo) CountWithFilter(f domain.UserListFilter) (int64, error) { panic("unused") }

// recordingAuditLogger records LogEvent calls for assertion.
type recordingAuditLogger struct {
	calls []auditCall
}

type auditCall struct {
	eventType    string
	userID       *int64
	targetUserID *int64
	ipAddress    *string
	userAgent    *string
	details      map[string]interface{}
}

func (r *recordingAuditLogger) LogEvent(
	eventType string,
	userID *int64,
	targetUserID *int64,
	ipAddress *string,
	userAgent *string,
	details map[string]interface{},
) error {
	r.calls = append(r.calls, auditCall{
		eventType:    eventType,
		userID:       userID,
		targetUserID: targetUserID,
		ipAddress:    ipAddress,
		userAgent:    userAgent,
		details:      details,
	})
	return nil
}

// recordingLogger records Error calls for assertion.
type recordingLogger struct {
	errors []string
}

func (l *recordingLogger) Error(format string, v ...interface{}) {
	l.errors = append(l.errors, fmt.Sprintf(format, v...))
}

// buildRouter builds a chi router with ProtectedUserGuard mounted under
// /users/{id} and a simple inline handler that echoes the method name.
func buildRouter(repo domain.UserRepository, audit AuditLogger) http.Handler {
	log := &recordingLogger{}
	r := chi.NewRouter()
	r.Route("/users/{id}", func(r chi.Router) {
		r.Use(ProtectedUserGuard(repo, audit, log))
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("GET"))
		})
		r.Head("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Patch("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("PATCH"))
		})
		r.Put("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("PUT"))
		})
		r.Delete("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("DELETE"))
		})
	})
	return r
}

const (
	protectedEmail    = "br8kwall@gmail.com"
	nonProtectedEmail = "alice@example.com"
)

func makeRepo(id int64, email string) *stubUserRepo {
	return &stubUserRepo{
		users: map[int64]*domain.User{
			id: {ID: id, Email: email},
		},
	}
}

// --- Tests ---

func TestProtectedUserGuard_GETPassesThroughAlways(t *testing.T) {
	repo := makeRepo(1, protectedEmail)
	audit := &recordingAuditLogger{}
	router := buildRouter(repo, audit)

	req := httptest.NewRequest(http.MethodGet, "/users/1/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET protected user: got %d, want 200", rec.Code)
	}
	if rec.Body.String() != "GET" {
		t.Errorf("handler not reached: body = %q", rec.Body.String())
	}
	if len(audit.calls) != 0 {
		t.Errorf("GET should not produce audit events, got %d", len(audit.calls))
	}
}

func TestProtectedUserGuard_HEADPassesThroughAlways(t *testing.T) {
	repo := makeRepo(1, protectedEmail)
	audit := &recordingAuditLogger{}
	router := buildRouter(repo, audit)

	req := httptest.NewRequest(http.MethodHead, "/users/1/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("HEAD protected user: got %d, want 200", rec.Code)
	}
	if len(audit.calls) != 0 {
		t.Errorf("HEAD should not produce audit events, got %d", len(audit.calls))
	}
}

func TestProtectedUserGuard_PATCHOnProtectedReturns403(t *testing.T) {
	repo := makeRepo(1, protectedEmail)
	audit := &recordingAuditLogger{}
	router := buildRouter(repo, audit)

	req := httptest.NewRequest(http.MethodPatch, "/users/1/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("PATCH protected user: got %d, want 403", rec.Code)
	}

	// Verify structured JSON error body uses the package constants.
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if body["error"] != ProtectedUserErrorCode {
		t.Errorf("body[\"error\"] = %q, want %q", body["error"], ProtectedUserErrorCode)
	}
	if body["message"] != ProtectedUserMessage {
		t.Errorf("body[\"message\"] = %q, want %q", body["message"], ProtectedUserMessage)
	}
	if body["documentation_url"] != ProtectedUserDocumentationURL {
		t.Errorf("body[\"documentation_url\"] = %q, want %q", body["documentation_url"], ProtectedUserDocumentationURL)
	}
}

func TestProtectedUserGuard_PATCHOnProtectedWritesAuditEvent(t *testing.T) {
	repo := makeRepo(42, protectedEmail)
	audit := &recordingAuditLogger{}
	log := &recordingLogger{}
	guard := ProtectedUserGuard(repo, audit, log)

	r := chi.NewRouter()
	r.Route("/users/{id}", func(r chi.Router) {
		r.Use(guard)
		r.Patch("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/42/", nil)
	// Inject an authenticated actor so the actorID ok=true branch is exercised.
	actorID := int64(99)
	ctx := context.WithValue(req.Context(), UserIDKey, actorID)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	if len(audit.calls) != 1 {
		t.Fatalf("expected 1 audit call, got %d", len(audit.calls))
	}

	call := audit.calls[0]
	if call.eventType != domain.EventProtectedUserAttackHTTP {
		t.Errorf("eventType = %q, want %q", call.eventType, domain.EventProtectedUserAttackHTTP)
	}
	if call.targetUserID == nil || *call.targetUserID != 42 {
		t.Errorf("targetUserID = %v, want &42", call.targetUserID)
	}
	// The actor ID should be populated when auth context is present.
	if call.userID == nil || *call.userID != 99 {
		t.Errorf("userID = %v, want &99", call.userID)
	}
}

func TestProtectedUserGuard_PATCHOnNonProtectedPassesThrough(t *testing.T) {
	repo := makeRepo(7, nonProtectedEmail)
	audit := &recordingAuditLogger{}
	router := buildRouter(repo, audit)

	req := httptest.NewRequest(http.MethodPatch, "/users/7/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("PATCH non-protected: got %d, want 200", rec.Code)
	}
	if rec.Body.String() != "PATCH" {
		t.Errorf("handler not reached: body = %q", rec.Body.String())
	}
	if len(audit.calls) != 0 {
		t.Errorf("non-protected write should not produce audit events, got %d", len(audit.calls))
	}
}

func TestProtectedUserGuard_DELETEOnProtectedReturns403(t *testing.T) {
	repo := makeRepo(1, protectedEmail)
	audit := &recordingAuditLogger{}
	router := buildRouter(repo, audit)

	req := httptest.NewRequest(http.MethodDelete, "/users/1/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("DELETE protected user: got %d, want 403", rec.Code)
	}
}

func TestProtectedUserGuard_MalformedIDParamReturns404(t *testing.T) {
	repo := &stubUserRepo{users: map[int64]*domain.User{}}
	audit := &recordingAuditLogger{}
	router := buildRouter(repo, audit)

	req := httptest.NewRequest(http.MethodPatch, "/users/not-an-int/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("malformed ID: got %d, want 404", rec.Code)
	}
	if len(audit.calls) != 0 {
		t.Errorf("malformed ID should not produce audit events, got %d", len(audit.calls))
	}
}

func TestProtectedUserGuard_UnknownUserReturns404(t *testing.T) {
	// repo returns (nil, nil) for unknown IDs — matches actual repository contract
	repo := &stubUserRepo{users: map[int64]*domain.User{}}
	audit := &recordingAuditLogger{}
	router := buildRouter(repo, audit)

	req := httptest.NewRequest(http.MethodPatch, "/users/999/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("unknown user: got %d, want 404", rec.Code)
	}
	if len(audit.calls) != 0 {
		t.Errorf("unknown user should not produce audit events, got %d", len(audit.calls))
	}
}

// TestProtectedUserGuard_RepoErrorReturns500 ensures internal errors don't leak
// as 403 or 404, but tests the 500 code path for coverage.
func TestProtectedUserGuard_RepoErrorReturns500(t *testing.T) {
	repo := &stubUserRepo{err: fmt.Errorf("db connection lost")}
	audit := &recordingAuditLogger{}
	log := &recordingLogger{}
	guard := ProtectedUserGuard(repo, audit, log)

	r := chi.NewRouter()
	r.Route("/users/{id}", func(r chi.Router) {
		r.Use(guard)
		r.Patch("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/1/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("repo error: got %d, want 500", rec.Code)
	}
	if len(log.errors) == 0 {
		t.Error("expected structured logger to capture the repo error, got none")
	}
}

// TestProtectedUserGuard_AuditLogFailureStill403 verifies that an audit.LogEvent
// error does not prevent the 403 from being returned to the caller (the guard
// logs the failure via the structured logger and continues).
func TestProtectedUserGuard_AuditLogFailureStill403(t *testing.T) {
	repo := makeRepo(1, protectedEmail)
	audit := &failingAuditLogger{}
	log := &recordingLogger{}
	guard := ProtectedUserGuard(repo, audit, log)

	r := chi.NewRouter()
	r.Route("/users/{id}", func(r chi.Router) {
		r.Use(guard)
		r.Patch("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/1/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("audit failure should not change HTTP status: got %d, want 403", rec.Code)
	}
	if len(log.errors) == 0 {
		t.Error("expected structured logger to capture the audit error, got none")
	}
}

// failingAuditLogger always returns an error from LogEvent, simulating an
// audit-system outage. Used to exercise the error-handling branch in
// ProtectedUserGuard that must not propagate to the HTTP response.
type failingAuditLogger struct{}

func (f *failingAuditLogger) LogEvent(
	_ string,
	_ *int64,
	_ *int64,
	_ *string,
	_ *string,
	_ map[string]interface{},
) error {
	return fmt.Errorf("audit system unavailable")
}

// TestProtectedUserGuard_BrokenWriterStill403 exercises the json.Encode error
// branch by using a ResponseWriter whose Write always fails after the header
// is flushed. The guard must not panic and must have attempted to write 403.
func TestProtectedUserGuard_BrokenWriterStill403(t *testing.T) {
	repo := makeRepo(1, protectedEmail)
	audit := &recordingAuditLogger{}
	log := &recordingLogger{}
	guard := ProtectedUserGuard(repo, audit, log)

	handler := guard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPatch, "/users/1", nil)
	// Inject chi route context manually.
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	bw := &brokenWriter{ResponseRecorder: httptest.NewRecorder()}
	handler.ServeHTTP(bw, req)

	// Header should be 403 even though body write failed.
	if bw.Code != http.StatusForbidden {
		t.Errorf("broken writer: got %d, want 403", bw.Code)
	}
}

// brokenWriter wraps ResponseRecorder but makes all Write calls fail after
// WriteHeader is called with a 4xx/5xx status, to exercise the encode-error path.
type brokenWriter struct {
	*httptest.ResponseRecorder
}

func (b *brokenWriter) Write(_ []byte) (int, error) {
	if b.Code >= 400 {
		return 0, fmt.Errorf("write: broken pipe")
	}
	return b.ResponseRecorder.Write(nil)
}

// TestProtectedUserGuard_SqlErrNoRowsTreatedAs404 verifies forward-compatibility:
// if a future repository refactor returns sql.ErrNoRows instead of (nil, nil),
// the guard treats it as 404 rather than 500.
func TestProtectedUserGuard_SqlErrNoRowsTreatedAs404(t *testing.T) {
	repo := &stubUserRepo{err: errNoRows}
	audit := &recordingAuditLogger{}
	log := &recordingLogger{}
	guard := ProtectedUserGuard(repo, audit, log)

	r := chi.NewRouter()
	r.Route("/users/{id}", func(r chi.Router) {
		r.Use(guard)
		r.Patch("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/1/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("sql.ErrNoRows: got %d, want 404", rec.Code)
	}
	if len(log.errors) != 0 {
		t.Errorf("sql.ErrNoRows should not log an error, got: %v", log.errors)
	}
}

// TestProtectedUserGuard_NilActorWhenUnauthenticated verifies that when no
// authenticated context is present, the audit event's userID is nil (not 0).
func TestProtectedUserGuard_NilActorWhenUnauthenticated(t *testing.T) {
	repo := makeRepo(1, protectedEmail)
	audit := &recordingAuditLogger{}
	log := &recordingLogger{}
	guard := ProtectedUserGuard(repo, audit, log)

	// Build handler directly without any auth context — simulates an
	// unauthenticated request reaching the middleware.
	handler := guard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPatch, "/users/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	// Deliberately do NOT set a user ID in the context.
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	if len(audit.calls) != 1 {
		t.Fatalf("expected 1 audit call, got %d", len(audit.calls))
	}
	if audit.calls[0].userID != nil {
		t.Errorf("unauthenticated request: audit userID = %v, want nil", audit.calls[0].userID)
	}
}
