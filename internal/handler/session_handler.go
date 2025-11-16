package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// SessionHandler handles session management operations
type SessionHandler struct {
	userService *service.UserService
	logger      *logger.Logger
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(userService *service.UserService, logger *logger.Logger) *SessionHandler {
	return &SessionHandler{
		userService: userService,
		logger:      logger,
	}
}

// ListSessions handles GET /api/sessions
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all active sessions
	sessions, err := h.userService.GetActiveSessions(userID)
	if err != nil {
		h.logger.Error("Failed to get sessions",
			"user_id", userID,
			"error", err.Error(),
		)
		http.Error(w, "Failed to retrieve sessions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
	})
}

// RevokeSession handles DELETE /api/sessions/:id
func (h *SessionHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get session ID from URL
	sessionIDStr := chi.URLParam(r, "id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Revoke the session
	if err := h.userService.RevokeSession(userID, sessionID); err != nil {
		h.logger.Error("Failed to revoke session",
			"user_id", userID,
			"session_id", sessionID,
			"error", err.Error(),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("Session revoked",
		"user_id", userID,
		"session_id", sessionID,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Session revoked successfully",
	})
}

// RevokeAllSessions handles POST /api/sessions/revoke-all
func (h *SessionHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body for optional except_token_id
	var request struct {
		ExceptCurrentSession bool `json:"except_current_session"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// If body is empty or invalid, treat as revoke all
		request.ExceptCurrentSession = false
	}

	// Revoke all sessions
	var exceptTokenID *int64
	// Note: In a real implementation, you'd track the current session's token ID
	// For now, we'll just support revoking all
	if err := h.userService.RevokeAllSessions(userID, exceptTokenID); err != nil {
		h.logger.Error("Failed to revoke all sessions",
			"user_id", userID,
			"error", err.Error(),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("All sessions revoked",
		"user_id", userID,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "All sessions revoked successfully",
	})
}
