package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewUserHandler(t *testing.T) {
	handler := NewUserHandler(nil, nil, createTestLogger())
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

func TestUserHandler_GetProfile_Unauthorized(t *testing.T) {
	handler := &UserHandler{}

	req := createTestRequest(http.MethodGet, "/api/profile", "")
	rr := httptest.NewRecorder()

	handler.GetProfile(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
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

func TestUserHandler_ChangePassword_EmptyStrings(t *testing.T) {
	handler := &UserHandler{}

	// Empty string passwords
	req := createAuthenticatedRequest(http.MethodPost, "/api/profile/password", `{"old_password": "", "new_password": ""}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Both old_password and new_password are required")
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

func TestUserHandler_UploadAvatar_SpoofedContentType(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	// Filename and Content-Type both claim image/png, but bytes are HTML.
	// Magic-byte sniffing must override the client-supplied header.
	fileContent := []byte("<script>alert('xss')</script>")
	req := createMultipartRequest(http.MethodPost, "/api/profile/avatar", "avatar", "evil.png", "image/png", fileContent, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UploadAvatar(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "File must be an image")
}

func TestUserHandler_UploadAvatar_DisallowedExtensionWithImageBytes(t *testing.T) {
	handler := &UserHandler{
		logger: createTestLogger(),
	}

	// Valid PNG magic bytes but a .html filename. Magic-byte sniffing alone
	// would let this through; the extension allowlist is what stops the
	// stored-XSS attack vector via a later /uploads/... navigation.
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	pngBody := append(pngMagic, []byte("rest of png data ...")...)
	req := createMultipartRequest(http.MethodPost, "/api/profile/avatar", "avatar", "evil.html", "image/png", pngBody, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UploadAvatar(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Only jpg, png, gif, and webp images are allowed")
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

// Removed 11 panic-expectation tests:
// - TestUserHandler_UpdateProfile_NilService
// - TestUserHandler_GetProfile_NilService
// - TestUserHandler_DeleteAvatar_NilService
// - TestUserHandler_ChangePassword_NilService
// - TestUserHandler_UpdateProfile_WithValidBirthday
// - TestUserHandler_UpdateProfile_EmptyBirthday
// - TestUserHandler_UpdateProfile_OnlyName
// - TestUserHandler_UpdateProfile_OnlyEmail
// - TestUserHandler_UpdateProfile_AllFields
// - TestUserHandler_UploadAvatar_ValidImageNilService
// - TestUserHandler_UploadAvatar_ValidJPEGNilService
// These tests verified nil pointer panics, not business logic.

// Tests with mock service

func TestUserHandler_GetProfile_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, nil, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/profile", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.GetProfile(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "email")
}

func TestUserHandler_GetProfile_NotFound(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, nil, createTestLogger())

	// User ID 999 doesn't exist in mock
	req := createAuthenticatedRequest(http.MethodGet, "/api/profile", "", 999, "unknown@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetProfile(rr, req)

	// Handler returns 500 for user not found
	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestUserHandler_UpdateProfile_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, nil, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPut, "/api/profile", `{"name": "Updated Name"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.UpdateProfile(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestUserHandler_UpdateProfile_NotFound(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, nil, createTestLogger())

	// User ID 999 doesn't exist in mock
	req := createAuthenticatedRequest(http.MethodPut, "/api/profile", `{"name": "Updated Name"}`, 999, "unknown@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateProfile(rr, req)

	// Handler returns 500 for user not found
	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestUserHandler_DeleteAvatar_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, nil, createTestLogger())

	req := createAuthenticatedRequest(http.MethodDelete, "/api/profile/avatar", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.DeleteAvatar(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestUserHandler_DeleteAvatar_NotFound(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, nil, createTestLogger())

	// User ID 999 doesn't exist in mock
	req := createAuthenticatedRequest(http.MethodDelete, "/api/profile/avatar", "", 999, "unknown@example.com", "user")
	rr := httptest.NewRecorder()

	handler.DeleteAvatar(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestUserHandler_ChangePassword_WrongOldPassword(t *testing.T) {
	userService := createTestUserService()
	handler := NewUserHandler(userService, nil, createTestLogger())

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
	handler := NewUserHandler(userService, nil, createTestLogger())

	// User ID 999 doesn't exist in mock
	req := createAuthenticatedRequest(http.MethodPost, "/api/profile/password",
		`{"old_password": "oldpassword", "new_password": "newpassword123"}`,
		999, "unknown@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}
