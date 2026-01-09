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

// Removed 2 panic-expectation tests:
// - TestSettingsHandler_GetSettings_NilService
// - TestSettingsHandler_UpdateSettings_NilService
// These tests verified nil pointer panics, not business logic.

// Tests with mock service

func TestSettingsHandler_GetSettings_Success(t *testing.T) {
	settingsService := createTestUserSettingsService()
	handler := NewSettingsHandler(settingsService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/settings", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.GetSettings(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	// Should contain default settings fields
	assertBodyContains(t, rr, "theme")
}

func TestSettingsHandler_UpdateSettings_Success(t *testing.T) {
	settingsService := createTestUserSettingsService()
	handler := NewSettingsHandler(settingsService, createTestLogger())

	body := `{"theme": "dark", "weight_unit": "kg", "distance_unit": "km", "data_export_format": "CSV", "notification_preferences": "{}"}`
	req := createAuthenticatedRequest(http.MethodPut, "/api/settings", body, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.UpdateSettings(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "dark")
}

func TestSettingsHandler_UpdateSettings_EmptyJSON(t *testing.T) {
	settingsService := createTestUserSettingsService()
	handler := NewSettingsHandler(settingsService, createTestLogger())

	body := `{}`
	req := createAuthenticatedRequest(http.MethodPut, "/api/settings", body, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.UpdateSettings(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestSettingsHandler_GetSettings_CreatesDefaults(t *testing.T) {
	settingsService := createTestUserSettingsService()
	handler := NewSettingsHandler(settingsService, createTestLogger())

	// First call should create default settings
	req := createAuthenticatedRequest(http.MethodGet, "/api/settings", "", 2, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetSettings(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "light") // default theme
	assertBodyContains(t, rr, "lbs")   // default weight unit
}
