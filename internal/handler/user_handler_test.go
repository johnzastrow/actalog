package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewUserHandler(t *testing.T) {
	// Test constructor with nil dependencies
	handler := NewUserHandler(nil, nil)

	if handler == nil {
		t.Error("NewUserHandler() should not return nil")
	}

	if handler.userService != nil {
		t.Error("userService should be nil when passed nil")
	}

	if handler.logger != nil {
		t.Error("logger should be nil when passed nil")
	}
}

func TestUpdateProfileRequest_Struct(t *testing.T) {
	req := UpdateProfileRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Birthday: "1990-05-15",
	}

	if req.Name != "Test User" {
		t.Errorf("Name = %q, want %q", req.Name, "Test User")
	}
	if req.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", req.Email, "test@example.com")
	}
	if req.Birthday != "1990-05-15" {
		t.Errorf("Birthday = %q, want %q", req.Birthday, "1990-05-15")
	}

	// Test empty struct
	req2 := UpdateProfileRequest{}
	if req2.Name != "" || req2.Email != "" || req2.Birthday != "" {
		t.Error("Empty UpdateProfileRequest should have all fields empty")
	}
}

func TestProfileResponse_Struct(t *testing.T) {
	user := map[string]interface{}{
		"id":    1,
		"email": "test@example.com",
	}
	resp := ProfileResponse{
		User: user,
	}

	if resp.User == nil {
		t.Error("User should not be nil")
	}

	// Test nil user
	resp2 := ProfileResponse{User: nil}
	if resp2.User != nil {
		t.Error("User should be nil")
	}
}

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

func TestUserHandler_UpdateProfile_InvalidBirthdayFormats(t *testing.T) {
	handler := &UserHandler{}

	testCases := []struct {
		name     string
		birthday string
	}{
		{"wrong format", "15-05-1990"},
		{"partial date", "1990-05"},
		{"text date", "May 15, 1990"},
		{"slash format", "1990/05/15"},
		{"invalid month", "1990-13-01"},
		{"invalid day", "1990-01-32"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"birthday": "` + tc.birthday + `"}`
			req := createAuthenticatedRequest(http.MethodPut, "/api/profile", body, 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.UpdateProfile(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, "Invalid birthday format")
		})
	}
}

func TestUserHandler_UpdateProfile_NilService(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	// Valid request with valid birthday but nil service
	req := createAuthenticatedRequest(http.MethodPut, "/api/profile", `{"name": "New Name"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// This will panic due to nil userService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userService")
		}
	}()

	handler.UpdateProfile(rr, req)
}

func TestUserHandler_GetProfile_Unauthorized(t *testing.T) {
	handler := &UserHandler{}

	req := createTestRequest(http.MethodGet, "/api/profile", "")
	rr := httptest.NewRecorder()

	handler.GetProfile(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserHandler_GetProfile_NilService(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/profile", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// This will panic due to nil userService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userService")
		}
	}()

	handler.GetProfile(rr, req)
}

func TestUserHandler_UploadAvatar_Unauthorized(t *testing.T) {
	handler := &UserHandler{}

	req := createTestRequest(http.MethodPost, "/api/profile/avatar", "")
	rr := httptest.NewRecorder()

	handler.UploadAvatar(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserHandler_UploadAvatar_NoFile(t *testing.T) {
	handler := &UserHandler{}

	// Request with no multipart form
	req := createAuthenticatedRequest(http.MethodPost, "/api/profile/avatar", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UploadAvatar(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestUserHandler_DeleteAvatar_Unauthorized(t *testing.T) {
	handler := &UserHandler{}

	req := createTestRequest(http.MethodDelete, "/api/profile/avatar", "")
	rr := httptest.NewRecorder()

	handler.DeleteAvatar(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserHandler_DeleteAvatar_NilService(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/profile/avatar", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// This will panic due to nil userService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userService")
		}
	}()

	handler.DeleteAvatar(rr, req)
}

func TestUserHandler_ChangePassword_Unauthorized(t *testing.T) {
	handler := &UserHandler{}

	req := createTestRequest(http.MethodPost, "/api/profile/password", `{"old_password": "old", "new_password": "newpassword123"}`)
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestUserHandler_ChangePassword_InvalidJSON(t *testing.T) {
	handler := &UserHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/profile/password", "{bad json", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestUserHandler_ChangePassword_MissingOldPassword(t *testing.T) {
	handler := &UserHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/profile/password", `{"new_password": "newpassword123"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Both old_password and new_password are required")
}

func TestUserHandler_ChangePassword_MissingNewPassword(t *testing.T) {
	handler := &UserHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/profile/password", `{"old_password": "oldpassword"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Both old_password and new_password are required")
}

func TestUserHandler_ChangePassword_MissingBothPasswords(t *testing.T) {
	handler := &UserHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/profile/password", `{}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Both old_password and new_password are required")
}

func TestUserHandler_ChangePassword_PasswordTooShort(t *testing.T) {
	handler := &UserHandler{}

	// Password less than 8 characters
	req := createAuthenticatedRequest(http.MethodPost, "/api/profile/password", `{"old_password": "oldpass", "new_password": "short"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "New password must be at least 8 characters")
}

func TestUserHandler_ChangePassword_PasswordExactly7Chars(t *testing.T) {
	handler := &UserHandler{}

	// Password exactly 7 characters (boundary test)
	req := createAuthenticatedRequest(http.MethodPost, "/api/profile/password", `{"old_password": "oldpass", "new_password": "1234567"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "New password must be at least 8 characters")
}

func TestUserHandler_ChangePassword_NilService(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	// Valid request with 8-char password but nil service
	req := createAuthenticatedRequest(http.MethodPost, "/api/profile/password", `{"old_password": "oldpassword", "new_password": "12345678"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// This will panic due to nil userService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userService")
		}
	}()

	handler.ChangePassword(rr, req)
}

func TestUserHandler_ChangePassword_EmptyStrings(t *testing.T) {
	handler := &UserHandler{}

	// Empty string passwords
	req := createAuthenticatedRequest(http.MethodPost, "/api/profile/password", `{"old_password": "", "new_password": ""}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Both old_password and new_password are required")
}
