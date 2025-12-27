package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionHandler_ListSessions_Unauthorized(t *testing.T) {
	handler := &SessionHandler{}

	req := createTestRequest(http.MethodGet, "/api/sessions", "")
	rr := httptest.NewRecorder()

	handler.ListSessions(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestSessionHandler_RevokeSession_Unauthorized(t *testing.T) {
	handler := &SessionHandler{}

	req := createTestRequest(http.MethodDelete, "/api/sessions/1", "")
	rr := httptest.NewRecorder()

	handler.RevokeSession(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestSessionHandler_RevokeSession_InvalidID(t *testing.T) {
	handler := &SessionHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodDelete, "/api/sessions/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.RevokeSession(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid session ID")
}

func TestSessionHandler_RevokeAllSessions_Unauthorized(t *testing.T) {
	handler := &SessionHandler{}

	req := createTestRequest(http.MethodPost, "/api/sessions/revoke-all", "")
	rr := httptest.NewRecorder()

	handler.RevokeAllSessions(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}
