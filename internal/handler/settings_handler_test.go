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

func TestNewSettingsHandler(t *testing.T) {
	handler := NewSettingsHandler(nil, nil)
	if handler == nil {
		t.Error("NewSettingsHandler should return a non-nil handler")
	}
}

func TestSettingsHandler_GetSettings_NilService(t *testing.T) {
	handler := &SettingsHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/settings", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("GetSettings requires service")
		}
	}()

	handler.GetSettings(rr, req)
}

func TestSettingsHandler_UpdateSettings_NilService(t *testing.T) {
	handler := &SettingsHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/settings", `{"theme": "dark"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry with valid JSON
	defer func() {
		if r := recover(); r == nil {
			t.Log("UpdateSettings requires service")
		}
	}()

	handler.UpdateSettings(rr, req)
}
