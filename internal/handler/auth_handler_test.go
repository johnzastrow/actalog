package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
