package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSettingsHandler_GetSettings_Unauthorized(t *testing.T) {
	handler := &SettingsHandler{}

	// Request without user context (unauthenticated)
	req := createTestRequest(http.MethodGet, "/api/settings", "")
	rr := httptest.NewRecorder()

	handler.GetSettings(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestSettingsHandler_UpdateSettings_Unauthorized(t *testing.T) {
	handler := &SettingsHandler{}

	// Request without user context (unauthenticated)
	req := createTestRequest(http.MethodPut, "/api/settings", `{"theme": "dark"}`)
	rr := httptest.NewRecorder()

	handler.UpdateSettings(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestSettingsHandler_UpdateSettings_InvalidJSON(t *testing.T) {
	handler := &SettingsHandler{}

	// Authenticated request with invalid JSON
	req := createAuthenticatedRequest(http.MethodPut, "/api/settings", "{bad json", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateSettings(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}
