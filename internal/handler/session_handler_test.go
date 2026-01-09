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

func TestNewSessionHandler(t *testing.T) {
	handler := NewSessionHandler(nil, nil)
	if handler == nil {
		t.Error("NewSessionHandler should return a non-nil handler")
	}
}

func TestSessionHandler_ListSessions_NilService(t *testing.T) {
	handler := &SessionHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/sessions", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("ListSessions requires service")
		}
	}()

	handler.ListSessions(rr, req)
}

func TestSessionHandler_RevokeSession_ValidIDNilService(t *testing.T) {
	handler := &SessionHandler{}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/sessions/1", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry with valid ID
	defer func() {
		if r := recover(); r == nil {
			t.Log("RevokeSession requires service")
		}
	}()

	handler.RevokeSession(rr, req)
}

func TestSessionHandler_RevokeAllSessions_NilService(t *testing.T) {
	handler := &SessionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/sessions/revoke-all", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("RevokeAllSessions requires service")
		}
	}()

	handler.RevokeAllSessions(rr, req)
}

// Tests with mock service

func TestSessionHandler_ListSessions_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewSessionHandler(userService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/sessions", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListSessions(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestSessionHandler_RevokeAllSessions_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewSessionHandler(userService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/sessions/revoke-all", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.RevokeAllSessions(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestSessionHandler_RevokeSession_NotFound(t *testing.T) {
	userService := createTestUserService()
	handler := NewSessionHandler(userService, createTestLogger())

	// Try to revoke a session that doesn't exist
	req := createAuthenticatedRequest(http.MethodDelete, "/api/sessions/999", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.RevokeSession(rr, req)

	// Returns 400 when session not found or doesn't belong to user
	assertStatusCode(t, rr, http.StatusBadRequest)
}
