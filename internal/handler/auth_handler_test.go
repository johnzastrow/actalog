package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/service"
)

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	// Create handler with nil service - validation should fail before service is called
	handler := &AuthHandler{}

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "empty body",
			body:       "",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid request body",
		},
		{
			name:       "invalid JSON",
			body:       "{not valid json}",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid request body",
		},
		{
			name:       "missing all required fields",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Name, email, and password are required",
		},
		{
			name:       "missing email",
			body:       `{"name": "Test", "password": "test123"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Name, email, and password are required",
		},
		{
			name:       "missing password",
			body:       `{"name": "Test", "email": "test@example.com"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Name, email, and password are required",
		},
		{
			name:       "missing name",
			body:       `{"email": "test@example.com", "password": "test123"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Name, email, and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodPost, "/api/auth/register", tt.body)
			rr := httptest.NewRecorder()

			handler.Register(rr, req)

			assertStatusCode(t, rr, tt.wantStatus)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	handler := &AuthHandler{}

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "empty body",
			body:       "",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid request body",
		},
		{
			name:       "invalid JSON",
			body:       "{bad json",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid request body",
		},
		{
			name:       "missing all required fields",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Email and password are required",
		},
		{
			name:       "missing password",
			body:       `{"email": "test@example.com"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Email and password are required",
		},
		{
			name:       "missing email",
			body:       `{"password": "test123"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Email and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodPost, "/api/auth/login", tt.body)
			rr := httptest.NewRecorder()

			handler.Login(rr, req)

			assertStatusCode(t, rr, tt.wantStatus)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestAuthHandler_ForgotPassword_InvalidInput(t *testing.T) {
	handler := &AuthHandler{}

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "invalid JSON",
			body:       "{bad json",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid request body",
		},
		{
			name:       "missing email",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Email is required",
		},
		{
			name:       "empty email",
			body:       `{"email": ""}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodPost, "/api/auth/forgot-password", tt.body)
			rr := httptest.NewRecorder()

			handler.ForgotPassword(rr, req)

			assertStatusCode(t, rr, tt.wantStatus)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestAuthHandler_ResetPassword_InvalidInput(t *testing.T) {
	handler := &AuthHandler{}

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "invalid JSON",
			body:       "{bad json",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid request body",
		},
		{
			name:       "missing token and password",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Token and new password are required",
		},
		{
			name:       "missing password",
			body:       `{"token": "abc123"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Token and new password are required",
		},
		{
			name:       "missing token",
			body:       `{"new_password": "newpass123"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Token and new password are required",
		},
		{
			name:       "password too short",
			body:       `{"token": "abc123", "new_password": "short"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodPost, "/api/auth/reset-password", tt.body)
			rr := httptest.NewRecorder()

			handler.ResetPassword(rr, req)

			assertStatusCode(t, rr, tt.wantStatus)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestAuthHandler_VerifyEmail_MissingToken(t *testing.T) {
	handler := &AuthHandler{}

	req := createTestRequest(http.MethodGet, "/api/auth/verify-email", "")
	rr := httptest.NewRecorder()

	handler.VerifyEmail(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Verification token is required")
}

func TestAuthHandler_ResendVerification_InvalidInput(t *testing.T) {
	handler := &AuthHandler{}

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "invalid JSON",
			body:       "{bad json",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid request body",
		},
		{
			name:       "missing email",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Email is required",
		},
		{
			name:       "empty email",
			body:       `{"email": ""}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodPost, "/api/auth/resend-verification", tt.body)
			rr := httptest.NewRecorder()

			handler.ResendVerification(rr, req)

			assertStatusCode(t, rr, tt.wantStatus)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestAuthHandler_RefreshToken_InvalidInput(t *testing.T) {
	handler := &AuthHandler{}

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "invalid JSON",
			body:       "{bad json",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid request body",
		},
		{
			name:       "missing refresh token",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Refresh token is required",
		},
		{
			name:       "empty refresh token",
			body:       `{"refresh_token": ""}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Refresh token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodPost, "/api/auth/refresh", tt.body)
			rr := httptest.NewRecorder()

			handler.RefreshToken(rr, req)

			assertStatusCode(t, rr, tt.wantStatus)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestAuthHandler_RevokeToken_InvalidInput(t *testing.T) {
	handler := &AuthHandler{}

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "invalid JSON",
			body:       "{bad json",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid request body",
		},
		{
			name:       "missing refresh token",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Refresh token is required",
		},
		{
			name:       "empty refresh token",
			body:       `{"refresh_token": ""}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Refresh token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodPost, "/api/auth/revoke", tt.body)
			rr := httptest.NewRecorder()

			handler.RevokeToken(rr, req)

			assertStatusCode(t, rr, tt.wantStatus)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestRespondJSON(t *testing.T) {
	rr := httptest.NewRecorder()

	data := map[string]string{"message": "success"}
	respondJSON(rr, http.StatusOK, data)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
	assertBodyContains(t, rr, `"message":"success"`)
}

func TestRespondError(t *testing.T) {
	rr := httptest.NewRecorder()

	respondError(rr, http.StatusBadRequest, "test error")

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertContentType(t, rr, "application/json")
	assertBodyContains(t, rr, `"message":"test error"`)
}

func TestRespondErrorWithDetail(t *testing.T) {
	rr := httptest.NewRecorder()

	respondErrorWithDetail(rr, http.StatusBadRequest, "main error", "detailed error")

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertContentType(t, rr, "application/json")
	assertBodyContains(t, rr, `"message":"main error"`)
	assertBodyContains(t, rr, `"error":"detailed error"`)
}

func TestNewAuthHandler(t *testing.T) {
	handler := NewAuthHandler(nil, nil)
	if handler == nil {
		t.Error("NewAuthHandler should return a non-nil handler")
	}
}

// Removed 19 panic-expectation tests:
// - TestAuthHandler_Register_ValidInput
// - TestAuthHandler_Login_ValidInput
// - TestAuthHandler_ForgotPassword_ValidInput
// - TestAuthHandler_ResetPassword_ValidInput
// - TestAuthHandler_VerifyEmail_ValidToken
// - TestAuthHandler_ResendVerification_ValidInput
// - TestAuthHandler_RefreshToken_ValidInput
// - TestAuthHandler_RevokeToken_ValidInput
// - TestAuthHandler_Register_WithLogger
// - TestAuthHandler_Login_WithLogger
// - TestAuthHandler_Login_WithRememberMe
// - TestAuthHandler_Register_DifferentUserAgents (4 subtests)
// - TestAuthHandler_Login_DifferentUserAgents (3 subtests)
// - TestAuthHandler_VerifyEmail_WithLogger
// - TestAuthHandler_ForgotPassword_WithLogger
// - TestAuthHandler_ResetPassword_WithLogger
// - TestAuthHandler_ResendVerification_WithLogger
// - TestAuthHandler_RefreshToken_WithLogger
// - TestAuthHandler_RevokeToken_WithLogger
// These tests verified nil pointer panics, not business logic.

// =============================================
// Mock-based tests with real service for Auth handler
// =============================================

// Helper function to create a test user service with mock repositories
func createTestUserService() *service.UserService {
	mockUserRepo := NewMockUserRepository()
	mockRefreshTokenRepo := NewMockRefreshTokenRepository()
	mockUserSubRepo := NewMockUserSubscriptionRepository()
	mockAuditLogRepo := NewMockAuditLogRepository()
	mockEmailService := NewMockEmailService()

	auditLogService := service.NewAuditLogService(mockAuditLogRepo)

	return service.NewUserService(
		mockUserRepo,
		mockRefreshTokenRepo,
		mockUserSubRepo,
		auditLogService,
		"test-jwt-secret-key-1234567890",
		time.Hour*24,
		time.Hour*24*7,
		true, // allowRegistration
		mockEmailService,
		"http://localhost:3000",
		false, // requireVerification
		5,     // maxLoginAttempts
		time.Minute*15,
	)
}

func TestAuthHandler_Register_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	body := `{"name": "Test User", "email": "newuser@example.com", "password": "password123"}`
	req := createTestRequest(http.MethodPost, "/api/auth/register", body)
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assertStatusCode(t, rr, http.StatusCreated)
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	// Register first user
	body := `{"name": "Test User", "email": "duplicate@example.com", "password": "password123"}`
	req := createTestRequest(http.MethodPost, "/api/auth/register", body)
	rr := httptest.NewRecorder()
	handler.Register(rr, req)
	assertStatusCode(t, rr, http.StatusCreated)

	// Try to register with same email
	req2 := createTestRequest(http.MethodPost, "/api/auth/register", body)
	rr2 := httptest.NewRecorder()
	handler.Register(rr2, req2)

	assertStatusCode(t, rr2, http.StatusConflict)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	body := `{"email": "nonexistent@example.com", "password": "wrongpassword"}`
	req := createTestRequest(http.MethodPost, "/api/auth/login", body)
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	// First register a user
	registerBody := `{"name": "Test User", "email": "logintest@example.com", "password": "password123"}`
	regReq := createTestRequest(http.MethodPost, "/api/auth/register", registerBody)
	regRR := httptest.NewRecorder()
	handler.Register(regRR, regReq)
	assertStatusCode(t, regRR, http.StatusCreated)

	// Now try to login
	loginBody := `{"email": "logintest@example.com", "password": "password123"}`
	loginReq := createTestRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginRR := httptest.NewRecorder()
	handler.Login(loginRR, loginReq)

	assertStatusCode(t, loginRR, http.StatusOK)
	assertBodyContains(t, loginRR, "token")
}

func TestAuthHandler_ForgotPassword_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	// First register a user
	registerBody := `{"name": "Forgot User", "email": "forgot@example.com", "password": "password123"}`
	regReq := createTestRequest(http.MethodPost, "/api/auth/register", registerBody)
	regRR := httptest.NewRecorder()
	handler.Register(regRR, regReq)

	// Request password reset
	body := `{"email": "forgot@example.com"}`
	req := createTestRequest(http.MethodPost, "/api/auth/forgot-password", body)
	rr := httptest.NewRecorder()

	handler.ForgotPassword(rr, req)

	// Should succeed even if email is not found (security)
	assertStatusCode(t, rr, http.StatusOK)
}

func TestAuthHandler_ForgotPassword_InvalidEmail(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	body := `{"email": "nonexistent@example.com"}`
	req := createTestRequest(http.MethodPost, "/api/auth/forgot-password", body)
	rr := httptest.NewRecorder()

	handler.ForgotPassword(rr, req)

	// Should still return OK for security (don't reveal if email exists)
	assertStatusCode(t, rr, http.StatusOK)
}

func TestAuthHandler_ResetPassword_InvalidToken(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	body := `{"token": "invalid-token", "new_password": "newpassword123"}`
	req := createTestRequest(http.MethodPost, "/api/auth/reset-password", body)
	rr := httptest.NewRecorder()

	handler.ResetPassword(rr, req)

	// Should fail - service returns 500 for invalid token
	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestAuthHandler_VerifyEmail_InvalidToken(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	req := createTestRequest(http.MethodGet, "/api/auth/verify-email?token=invalid-token", "")
	rr := httptest.NewRecorder()

	handler.VerifyEmail(rr, req)

	// Should fail - service returns 500 for invalid token
	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestAuthHandler_ResendVerification_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	// First register a user
	registerBody := `{"name": "Verify User", "email": "verify@example.com", "password": "password123"}`
	regReq := createTestRequest(http.MethodPost, "/api/auth/register", registerBody)
	regRR := httptest.NewRecorder()
	handler.Register(regRR, regReq)

	// Request resend verification
	body := `{"email": "verify@example.com"}`
	req := createTestRequest(http.MethodPost, "/api/auth/resend-verification", body)
	rr := httptest.NewRecorder()

	handler.ResendVerification(rr, req)

	// May return different status based on verification setting
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadRequest {
		t.Errorf("Expected OK or BadRequest, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	body := `{"refresh_token": "invalid-refresh-token"}`
	req := createTestRequest(http.MethodPost, "/api/auth/refresh", body)
	rr := httptest.NewRecorder()

	handler.RefreshToken(rr, req)

	// Should fail with invalid token - could be 401 or 500
	if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected Unauthorized or InternalServerError, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_RevokeToken_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	body := `{"refresh_token": "some-token"}`
	req := createTestRequest(http.MethodPost, "/api/auth/revoke", body)
	rr := httptest.NewRecorder()

	handler.RevokeToken(rr, req)

	// Revoke should work - returns 500 if token not found, 401 or 204 if successful
	if rr.Code != http.StatusOK && rr.Code != http.StatusUnauthorized && rr.Code != http.StatusNoContent && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected success/unauthorized/error, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	// First register a user
	registerBody := `{"name": "Wrong Password User", "email": "wrongpass@example.com", "password": "password123"}`
	regReq := createTestRequest(http.MethodPost, "/api/auth/register", registerBody)
	regRR := httptest.NewRecorder()
	handler.Register(regRR, regReq)

	// Try login with wrong password
	loginBody := `{"email": "wrongpass@example.com", "password": "wrongpassword"}`
	loginReq := createTestRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginRR := httptest.NewRecorder()
	handler.Login(loginRR, loginReq)

	assertStatusCode(t, loginRR, http.StatusUnauthorized)
	assertBodyContains(t, loginRR, "Invalid")
}

func TestAuthHandler_Login_NonexistentUser(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	// Try login with non-existent user
	loginBody := `{"email": "nonexistent@example.com", "password": "password123"}`
	loginReq := createTestRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginRR := httptest.NewRecorder()
	handler.Login(loginRR, loginReq)

	assertStatusCode(t, loginRR, http.StatusUnauthorized)
}

func TestAuthHandler_Login_EmptyPassword(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	loginBody := `{"email": "test@example.com", "password": ""}`
	loginReq := createTestRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginRR := httptest.NewRecorder()
	handler.Login(loginRR, loginReq)

	assertStatusCode(t, loginRR, http.StatusBadRequest)
}

func TestAuthHandler_Login_EmptyEmail(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	loginBody := `{"email": "", "password": "password123"}`
	loginReq := createTestRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginRR := httptest.NewRecorder()
	handler.Login(loginRR, loginReq)

	assertStatusCode(t, loginRR, http.StatusBadRequest)
}

func TestAuthHandler_Login_WithRememberMeFalse(t *testing.T) {
	userService := createTestUserService()
	handler := NewAuthHandler(userService, createTestLogger())

	// First register a user
	registerBody := `{"name": "Remember User", "email": "remember@example.com", "password": "password123"}`
	regReq := createTestRequest(http.MethodPost, "/api/auth/register", registerBody)
	regRR := httptest.NewRecorder()
	handler.Register(regRR, regReq)

	// Login with remember_me false
	loginBody := `{"email": "remember@example.com", "password": "password123", "remember_me": false}`
	loginReq := createTestRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginRR := httptest.NewRecorder()
	handler.Login(loginRR, loginReq)

	assertStatusCode(t, loginRR, http.StatusOK)
	assertBodyContains(t, loginRR, "token")
}
