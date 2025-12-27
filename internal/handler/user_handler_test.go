package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserHandler_UpdateProfile_Unauthorized(t *testing.T) {
	handler := &UserHandler{}

	// Request without user context
	req := createTestRequest(http.MethodPut, "/api/profile", `{"name": "New Name"}`)
	rr := httptest.NewRecorder()

	handler.UpdateProfile(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserHandler_UpdateProfile_InvalidJSON(t *testing.T) {
	handler := &UserHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/profile", "{bad json", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateProfile(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestUserHandler_UpdateProfile_InvalidBirthday(t *testing.T) {
	handler := &UserHandler{}

	// Invalid birthday format
	req := createAuthenticatedRequest(http.MethodPut, "/api/profile", `{"birthday": "invalid-date"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateProfile(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid birthday format")
}
