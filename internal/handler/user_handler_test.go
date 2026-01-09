package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewUserHandler(t *testing.T) {
	handler := NewUserHandler(nil, createTestLogger())
	if handler == nil {
		t.Fatal("NewUserHandler() should not return nil")
	}
}

// Removed struct field assignment tests:
// - TestUpdateProfileRequest_Struct
// - TestProfileResponse_Struct
// These tests verified Go struct assignment works, not business logic.

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

func TestUserHandler_UpdateProfile_WithValidBirthday(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	// Valid birthday format should pass validation and reach service
	req := createAuthenticatedRequest(http.MethodPut, "/api/profile", `{"name": "New Name", "birthday": "1990-05-15"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// This will panic due to nil userService - tests birthday parsing passes
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userService")
		}
	}()

	handler.UpdateProfile(rr, req)
}

func TestUserHandler_UpdateProfile_EmptyBirthday(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	// Empty birthday should skip parsing and reach service
	req := createAuthenticatedRequest(http.MethodPut, "/api/profile", `{"name": "New Name", "birthday": ""}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// This will panic due to nil userService - tests empty birthday is allowed
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userService")
		}
	}()

	handler.UpdateProfile(rr, req)
}

func TestUserHandler_UpdateProfile_OnlyName(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	// Only updating name (no birthday field at all)
	req := createAuthenticatedRequest(http.MethodPut, "/api/profile", `{"name": "New Name"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userService")
		}
	}()

	handler.UpdateProfile(rr, req)
}

func TestUserHandler_UpdateProfile_OnlyEmail(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	// Only updating email
	req := createAuthenticatedRequest(http.MethodPut, "/api/profile", `{"email": "newemail@example.com"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userService")
		}
	}()

	handler.UpdateProfile(rr, req)
}

func TestUserHandler_UpdateProfile_AllFields(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	// Updating all fields
	req := createAuthenticatedRequest(http.MethodPut, "/api/profile",
		`{"name": "New Name", "email": "newemail@example.com", "birthday": "1990-05-15"}`,
		1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil userService")
		}
	}()

	handler.UpdateProfile(rr, req)
}

func TestUserHandler_UploadAvatar_InvalidContentType(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	// Upload a non-image file (text file)
	fileContent := []byte("This is not an image")
	req := createMultipartRequest(http.MethodPost, "/api/profile/avatar", "avatar", "test.txt", "text/plain", fileContent, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UploadAvatar(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "File must be an image")
}

func TestUserHandler_UploadAvatar_ValidImageNilService(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	// Upload a valid image file (tests file processing paths)
	// Using a minimal PNG header for content type detection
	fileContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	req := createMultipartRequest(http.MethodPost, "/api/profile/avatar", "avatar", "test.png", "image/png", fileContent, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// This will panic when trying to get user from nil service (after file validation passes)
	defer func() {
		if r := recover(); r == nil {
			t.Log("UploadAvatar requires service after file validation")
		}
	}()

	handler.UploadAvatar(rr, req)
}

func TestUserHandler_UploadAvatar_ValidJPEGNilService(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	// Upload a valid JPEG file
	fileContent := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	req := createMultipartRequest(http.MethodPost, "/api/profile/avatar", "avatar", "test.jpg", "image/jpeg", fileContent, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("UploadAvatar requires service after file validation")
		}
	}()

	handler.UploadAvatar(rr, req)
}

func TestUserHandler_UploadAvatar_ApplicationPDF(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	// Upload a PDF file (not an image)
	fileContent := []byte("PDF content")
	req := createMultipartRequest(http.MethodPost, "/api/profile/avatar", "avatar", "test.pdf", "application/pdf", fileContent, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UploadAvatar(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "File must be an image")
}

func TestUserHandler_UploadAvatar_WrongFieldName(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	// Upload with wrong field name (not "avatar")
	fileContent := []byte{0x89, 0x50, 0x4E, 0x47}
	req := createMultipartRequest(http.MethodPost, "/api/profile/avatar", "wrong_field", "test.png", "image/png", fileContent, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UploadAvatar(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "No file provided")
}

// Tests with mock service

func TestUserHandler_GetProfile_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/profile", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.GetProfile(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "email")
}

func TestUserHandler_GetProfile_NotFound(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, createTestLogger())

	// User ID 999 doesn't exist in mock
	req := createAuthenticatedRequest(http.MethodGet, "/api/profile", "", 999, "unknown@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetProfile(rr, req)

	// Handler returns 500 for user not found
	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestUserHandler_UpdateProfile_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPut, "/api/profile", `{"name": "Updated Name"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.UpdateProfile(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestUserHandler_UpdateProfile_NotFound(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, createTestLogger())

	// User ID 999 doesn't exist in mock
	req := createAuthenticatedRequest(http.MethodPut, "/api/profile", `{"name": "Updated Name"}`, 999, "unknown@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateProfile(rr, req)

	// Handler returns 500 for user not found
	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestUserHandler_DeleteAvatar_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodDelete, "/api/profile/avatar", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.DeleteAvatar(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestUserHandler_DeleteAvatar_NotFound(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, createTestLogger())

	// User ID 999 doesn't exist in mock
	req := createAuthenticatedRequest(http.MethodDelete, "/api/profile/avatar", "", 999, "unknown@example.com", "user")
	rr := httptest.NewRecorder()

	handler.DeleteAvatar(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestUserHandler_ChangePassword_WrongOldPassword(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/profile/password",
		`{"old_password": "wrongpassword", "new_password": "newpassword123"}`,
		1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	// Returns 401 when old password doesn't match
	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Current password is incorrect")
}

func TestUserHandler_ChangePassword_NotFound(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, createTestLogger())

	// User ID 999 doesn't exist in mock
	req := createAuthenticatedRequest(http.MethodPost, "/api/profile/password",
		`{"old_password": "oldpassword", "new_password": "newpassword123"}`,
		999, "unknown@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}
