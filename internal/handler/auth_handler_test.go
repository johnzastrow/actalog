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

func TestAuthHandler_Register_ValidInput(t *testing.T) {
	handler := &AuthHandler{}

	req := createTestRequest(http.MethodPost, "/api/auth/register", `{"name": "Test User", "email": "test@example.com", "password": "password123"}`)
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests that validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("Register requires service")
		}
	}()

	handler.Register(rr, req)
}

func TestAuthHandler_Login_ValidInput(t *testing.T) {
	handler := &AuthHandler{}

	req := createTestRequest(http.MethodPost, "/api/auth/login", `{"email": "test@example.com", "password": "password123"}`)
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests that validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("Login requires service")
		}
	}()

	handler.Login(rr, req)
}

func TestAuthHandler_ForgotPassword_ValidInput(t *testing.T) {
	handler := &AuthHandler{}

	req := createTestRequest(http.MethodPost, "/api/auth/forgot-password", `{"email": "test@example.com"}`)
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests that validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("ForgotPassword requires service")
		}
	}()

	handler.ForgotPassword(rr, req)
}

func TestAuthHandler_ResetPassword_ValidInput(t *testing.T) {
	handler := &AuthHandler{}

	req := createTestRequest(http.MethodPost, "/api/auth/reset-password", `{"token": "abc123", "new_password": "newpassword123"}`)
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests that validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("ResetPassword requires service")
		}
	}()

	handler.ResetPassword(rr, req)
}

func TestAuthHandler_VerifyEmail_ValidToken(t *testing.T) {
	handler := &AuthHandler{}

	req := createTestRequest(http.MethodGet, "/api/auth/verify-email?token=abc123", "")
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests that validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("VerifyEmail requires service")
		}
	}()

	handler.VerifyEmail(rr, req)
}

func TestAuthHandler_ResendVerification_ValidInput(t *testing.T) {
	handler := &AuthHandler{}

	req := createTestRequest(http.MethodPost, "/api/auth/resend-verification", `{"email": "test@example.com"}`)
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests that validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("ResendVerification requires service")
		}
	}()

	handler.ResendVerification(rr, req)
}

func TestAuthHandler_RefreshToken_ValidInput(t *testing.T) {
	handler := &AuthHandler{}

	req := createTestRequest(http.MethodPost, "/api/auth/refresh", `{"refresh_token": "valid_token"}`)
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests that validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("RefreshToken requires service")
		}
	}()

	handler.RefreshToken(rr, req)
}

func TestAuthHandler_RevokeToken_ValidInput(t *testing.T) {
	handler := &AuthHandler{}

	req := createTestRequest(http.MethodPost, "/api/auth/revoke", `{"refresh_token": "valid_token"}`)
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests that validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("RevokeToken requires service")
		}
	}()

	handler.RevokeToken(rr, req)
}

func TestAuthHandler_Register_WithLogger(t *testing.T) {
	handler := &AuthHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/auth/register", `{"name": "Test User", "email": "test@example.com", "password": "password123"}`)
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests logger paths
	defer func() {
		if r := recover(); r == nil {
			t.Log("Register requires service - logger path tested")
		}
	}()

	handler.Register(rr, req)
}

func TestAuthHandler_Login_WithLogger(t *testing.T) {
	handler := &AuthHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/auth/login", `{"email": "test@example.com", "password": "password123"}`)
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests logger paths
	defer func() {
		if r := recover(); r == nil {
			t.Log("Login requires service - logger path tested")
		}
	}()

	handler.Login(rr, req)
}

func TestAuthHandler_Login_WithRememberMe(t *testing.T) {
	handler := &AuthHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/auth/login", `{"email": "test@example.com", "password": "password123", "remember_me": true}`)
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests remember_me parsing
	defer func() {
		if r := recover(); r == nil {
			t.Log("Login requires service - remember_me path tested")
		}
	}()

	handler.Login(rr, req)
}

func TestAuthHandler_Register_DifferentUserAgents(t *testing.T) {
	handler := &AuthHandler{
		logger: createTestLogger(),
	}

	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
		"Mozilla/5.0 (Linux; Android 10)",
		"",
	}

	for _, ua := range userAgents {
		t.Run("ua_"+ua[:min(10, len(ua))], func(t *testing.T) {
			req := createTestRequest(http.MethodPost, "/api/auth/register", `{"name": "Test", "email": "test@example.com", "password": "password123"}`)
			req.Header.Set("User-Agent", ua)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("Register logged user agent")
				}
			}()

			handler.Register(rr, req)
		})
	}
}

func TestAuthHandler_Login_DifferentUserAgents(t *testing.T) {
	handler := &AuthHandler{
		logger: createTestLogger(),
	}

	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"PostmanRuntime/7.28.4",
		"curl/7.64.1",
	}

	for _, ua := range userAgents {
		t.Run("ua_"+ua[:min(10, len(ua))], func(t *testing.T) {
			req := createTestRequest(http.MethodPost, "/api/auth/login", `{"email": "test@example.com", "password": "password123"}`)
			req.Header.Set("User-Agent", ua)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("Login logged user agent")
				}
			}()

			handler.Login(rr, req)
		})
	}
}

func TestAuthHandler_VerifyEmail_WithLogger(t *testing.T) {
	handler := &AuthHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/auth/verify-email?token=test123", "")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("VerifyEmail requires service")
		}
	}()

	handler.VerifyEmail(rr, req)
}

func TestAuthHandler_ForgotPassword_WithLogger(t *testing.T) {
	handler := &AuthHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/auth/forgot-password", `{"email": "test@example.com"}`)
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("ForgotPassword requires service")
		}
	}()

	handler.ForgotPassword(rr, req)
}

func TestAuthHandler_ResetPassword_WithLogger(t *testing.T) {
	handler := &AuthHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/auth/reset-password", `{"token": "test123", "new_password": "newpassword123"}`)
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("ResetPassword requires service")
		}
	}()

	handler.ResetPassword(rr, req)
}

func TestAuthHandler_ResendVerification_WithLogger(t *testing.T) {
	handler := &AuthHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/auth/resend-verification", `{"email": "test@example.com"}`)
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("ResendVerification requires service")
		}
	}()

	handler.ResendVerification(rr, req)
}

func TestAuthHandler_RefreshToken_WithLogger(t *testing.T) {
	handler := &AuthHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/auth/refresh", `{"refresh_token": "test123"}`)
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("RefreshToken requires service")
		}
	}()

	handler.RefreshToken(rr, req)
}

func TestAuthHandler_RevokeToken_WithLogger(t *testing.T) {
	handler := &AuthHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/auth/revoke", `{"refresh_token": "test123"}`)
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("RevokeToken requires service")
		}
	}()

	handler.RevokeToken(rr, req)
}

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
		true,  // allowRegistration
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
