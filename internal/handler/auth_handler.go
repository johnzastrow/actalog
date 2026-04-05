package handler

import (
	"encoding/json"
	"net/http"

	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userService *service.UserService
	logger      *logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *service.UserService, l *logger.Logger) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		logger:      l,
	}
}

// RegisterRequest represents a registration request
// @Description User registration request body
type RegisterRequest struct {
	Name     string `json:"name" example:"John Doe"`
	Email    string `json:"email" example:"john@example.com"`
	Password string `json:"password" example:"securePassword123"`
}

// LoginRequest represents a login request
// @Description User login request body
type LoginRequest struct {
	Email      string `json:"email" example:"john@example.com"`
	Password   string `json:"password" example:"securePassword123"`
	RememberMe bool   `json:"remember_me,omitempty" example:"true"`
}

// AuthResponse represents an authentication response
// @Description Authentication response containing JWT token and user info
type AuthResponse struct {
	Token        string      `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string      `json:"refresh_token,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	User         interface{} `json:"user"`
}

// ErrorResponse represents an error response
// @Description Error response returned when a request fails
type ErrorResponse struct {
	Message string `json:"message" example:"Invalid credentials"`
	Error   string `json:"error,omitempty" example:"additional error details"`
}

// Register handles user registration
// @Summary      Register a new user
// @Description  Create a new user account with name, email, and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Registration details"
// @Success      201 {object} AuthResponse "Successfully registered"
// @Failure      400 {object} ErrorResponse "Invalid request or validation error"
// @Failure      403 {object} ErrorResponse "Registration is closed"
// @Failure      409 {object} ErrorResponse "Email already exists"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorWithDetail(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate input
	if req.Name == "" || req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Name, email, and password are required")
		return
	}

	// Log attempt
	if h.logger != nil {
		h.logger.Info("action=register_attempt email=%s name=%s remote=%s ua=%s", req.Email, req.Name, r.RemoteAddr, r.UserAgent())
	}

	// Register user
	user, token, err := h.userService.Register(req.Name, req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrEmailAlreadyExists:
			if h.logger != nil {
				h.logger.Warn("action=register outcome=failure email=%s reason=email_exists", req.Email)
			}
			respondError(w, http.StatusConflict, "Email already exists")
		case service.ErrRegistrationClosed:
			if h.logger != nil {
				h.logger.Warn("action=register outcome=failure email=%s reason=registration_closed", req.Email)
			}
			respondError(w, http.StatusForbidden, "Registration is closed. Please contact an administrator.")
		default:
			if h.logger != nil {
				h.logger.Error("action=register outcome=failure email=%s error=%v", req.Email, err)
			}
			// Return the actual validation error to help users fix their input
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	if h.logger != nil {
		h.logger.Info("action=register outcome=success user_id=%d email=%s", user.ID, user.Email)
	}

	respondJSON(w, http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login handles user login
// @Summary      User login
// @Description  Authenticate user with email and password. Set remember_me to true for refresh token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login credentials"
// @Success      200 {object} AuthResponse "Successfully authenticated"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      401 {object} ErrorResponse "Invalid email or password"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorWithDetail(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=login_attempt email=%s remote=%s ua=%s", req.Email, r.RemoteAddr, r.UserAgent())
	}

	// Login user
	user, token, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			if h.logger != nil {
				h.logger.Warn("action=login outcome=failure email=%s reason=invalid_credentials", req.Email)
			}
			respondError(w, http.StatusUnauthorized, "Invalid email or password")
		} else {
			if h.logger != nil {
				h.logger.Error("action=login outcome=failure email=%s error=%v", req.Email, err)
			}
			respondErrorWithDetail(w, http.StatusInternalServerError, "Failed to login", err.Error())
		}
		return
	}

	response := AuthResponse{
		Token: token,
		User:  user,
	}

	// Create refresh token if remember_me is true
	if req.RememberMe {
		deviceInfo := r.UserAgent() // Get browser/device info from User-Agent header
		refreshToken, err := h.userService.CreateRefreshToken(user.ID, deviceInfo, req.RememberMe)
		if err != nil {
			// Log error but don't fail the login
			if h.logger != nil {
				h.logger.Warn("action=create_refresh_token outcome=failure user_id=%d email=%s error=%v", user.ID, user.Email, err)
			}
			// User can still use the access token
			respondErrorWithDetail(w, http.StatusInternalServerError, "Warning: Failed to create refresh token", err.Error())
		} else {
			response.RefreshToken = refreshToken
			if h.logger != nil {
				h.logger.Info("action=create_refresh_token outcome=success user_id=%d email=%s remember_me=%v", user.ID, user.Email, req.RememberMe)
			}
		}
	}
	if h.logger != nil {
		h.logger.Info("action=login outcome=success user_id=%d email=%s", user.ID, user.Email)
	}

	respondJSON(w, http.StatusOK, response)
}

// ForgotPasswordRequest represents a forgot password request
// @Description Request to initiate password reset
type ForgotPasswordRequest struct {
	Email string `json:"email" example:"john@example.com"`
}

// ResetPasswordRequest represents a reset password request
// @Description Request to complete password reset with token
type ResetPasswordRequest struct {
	Token       string `json:"token" example:"abc123def456"`
	NewPassword string `json:"new_password" example:"newSecurePassword123"`
}

// MessageResponse represents a simple message response
// @Description Generic message response
type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// ForgotPassword handles forgot password requests
// @Summary      Request password reset
// @Description  Send a password reset email to the user. Always returns success for security.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body ForgotPasswordRequest true "Email address"
// @Success      200 {object} MessageResponse "Reset email sent if address exists"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      500 {object} ErrorResponse "Failed to process request"
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate email
	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "Email is required")
		return
	}

	// Request password reset (always succeeds for security)
	err := h.userService.RequestPasswordReset(req.Email)
	if err != nil {
		// Log error but don't reveal to user
		// In production, this should use proper logging
		respondError(w, http.StatusInternalServerError, "Failed to process request")
		return
	}

	// Always return success (don't reveal if email exists)
	respondJSON(w, http.StatusOK, MessageResponse{
		Message: "If your email is registered, you will receive a password reset link shortly",
	})
}

// ResetPassword handles password reset requests
// @Summary      Reset password
// @Description  Complete password reset using token from email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body ResetPasswordRequest true "Reset token and new password"
// @Success      200 {object} MessageResponse "Password reset successfully"
// @Failure      400 {object} ErrorResponse "Invalid token, expired token, or weak password"
// @Failure      500 {object} ErrorResponse "Failed to reset password"
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Token == "" || req.NewPassword == "" {
		respondError(w, http.StatusBadRequest, "Token and new password are required")
		return
	}

	// Reset password (policy enforced in service layer)
	err := h.userService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		switch err {
		case service.ErrInvalidResetToken:
			respondError(w, http.StatusBadRequest, "Invalid reset token")
		case service.ErrResetTokenExpired:
			respondError(w, http.StatusBadRequest, "Reset token has expired. Please request a new one")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to reset password")
		}
		return
	}

	respondJSON(w, http.StatusOK, MessageResponse{
		Message: "Password has been reset successfully. You can now login with your new password",
	})
}

// VerifyEmailRequest represents an email verification request
// @Description Email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" example:"verification-token-123"`
}

// VerifyEmail handles email verification requests
// @Summary      Verify email address
// @Description  Verify user's email using token from verification email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        token query string true "Verification token from email"
// @Success      200 {object} MessageResponse "Email verified successfully"
// @Failure      400 {object} ErrorResponse "Invalid or expired token, or already verified"
// @Failure      500 {object} ErrorResponse "Failed to verify email"
// @Router       /auth/verify-email [get]
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "Verification token is required")
		return
	}

	// Verify email
	err := h.userService.VerifyEmail(token)
	if err != nil {
		switch err {
		case service.ErrInvalidVerificationToken:
			respondError(w, http.StatusBadRequest, "Invalid verification token")
		case service.ErrVerificationTokenExpired:
			respondError(w, http.StatusBadRequest, "Verification token has expired. Please request a new one")
		case service.ErrEmailAlreadyVerified:
			respondError(w, http.StatusBadRequest, "Email is already verified")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to verify email")
		}
		return
	}

	respondJSON(w, http.StatusOK, MessageResponse{
		Message: "Email verified successfully. You can now login",
	})
}

// ResendVerificationRequest represents a resend verification email request
// @Description Request to resend verification email
type ResendVerificationRequest struct {
	Email string `json:"email" example:"john@example.com"`
}

// ResendVerification handles resend verification email requests
// @Summary      Resend verification email
// @Description  Resend email verification link. Always returns success for security.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body ResendVerificationRequest true "Email address"
// @Success      200 {object} MessageResponse "Verification email sent if applicable"
// @Failure      400 {object} ErrorResponse "Invalid request or already verified"
// @Failure      500 {object} ErrorResponse "Failed to resend verification email"
// @Router       /auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate email
	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "Email is required")
		return
	}

	// Resend verification email
	err := h.userService.ResendVerificationEmail(req.Email)
	if err != nil {
		if err == service.ErrEmailAlreadyVerified {
			respondError(w, http.StatusBadRequest, "Email is already verified")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to resend verification email")
		return
	}

	// Always return success (don't reveal if email exists)
	respondJSON(w, http.StatusOK, MessageResponse{
		Message: "If your email is registered and not yet verified, you will receive a verification link shortly",
	})
}

// RefreshTokenRequest represents a refresh token request
// @Description Request to refresh access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// RefreshToken handles refresh token requests
// @Summary      Refresh access token
// @Description  Get a new access token using a valid refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RefreshTokenRequest true "Refresh token"
// @Success      200 {object} AuthResponse "New access token"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      401 {object} ErrorResponse "Invalid or expired refresh token"
// @Failure      500 {object} ErrorResponse "Failed to refresh token"
// @Router       /auth/refresh-token [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	// Refresh access token
	user, newAccessToken, err := h.userService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		if err == service.ErrInvalidRefreshToken {
			respondError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to refresh token")
		}
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		Token: newAccessToken,
		User:  user,
	})
}

// RevokeTokenRequest represents a revoke token request
// @Description Request to revoke a refresh token (logout)
type RevokeTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// RevokeToken handles token revocation (logout)
// @Summary      Revoke refresh token
// @Description  Invalidate a refresh token to logout from a specific session
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RevokeTokenRequest true "Refresh token to revoke"
// @Success      200 {object} MessageResponse "Token revoked successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      404 {object} ErrorResponse "Refresh token not found"
// @Failure      500 {object} ErrorResponse "Failed to revoke token"
// @Router       /auth/revoke-token [post]
func (h *AuthHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	var req RevokeTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	// Revoke token
	err := h.userService.RevokeRefreshToken(req.RefreshToken)
	if err != nil {
		if err == service.ErrInvalidRefreshToken {
			respondError(w, http.StatusNotFound, "Refresh token not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to revoke token")
		}
		return
	}

	respondJSON(w, http.StatusOK, MessageResponse{
		Message: "Token revoked successfully",
	})
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Message: message})
}

func respondErrorWithDetail(w http.ResponseWriter, status int, message string, detail string) {
	respondJSON(w, status, ErrorResponse{Message: message, Error: detail})
}
