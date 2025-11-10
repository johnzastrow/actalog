package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/johnzastrow/actalog/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userService *service.UserService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me,omitempty"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token,omitempty"`
	User         interface{} `json:"user"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Register handles user registration
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

	// Register user
	user, token, err := h.userService.Register(req.Name, req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrEmailAlreadyExists:
			respondError(w, http.StatusConflict, "Email already exists")
		case service.ErrRegistrationClosed:
			respondError(w, http.StatusForbidden, "Registration is closed. Please contact an administrator.")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to register user")
		}
		return
	}

	respondJSON(w, http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login handles user login
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

	// Login user
	user, token, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			respondError(w, http.StatusUnauthorized, "Invalid email or password")
		} else {
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
		refreshToken, err := h.userService.CreateRefreshToken(user.ID, deviceInfo)
		if err != nil {
			// Log error but don't fail the login
			// User can still use the access token
			respondErrorWithDetail(w, http.StatusInternalServerError, "Warning: Failed to create refresh token", err.Error())
		} else {
			response.RefreshToken = refreshToken
		}
	}

	respondJSON(w, http.StatusOK, response)
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest represents a reset password request
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// ForgotPassword handles forgot password requests
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

	// Validate password strength (minimum 8 characters)
	if len(req.NewPassword) < 8 {
		respondError(w, http.StatusBadRequest, "Password must be at least 8 characters long")
		return
	}

	// Reset password
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
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// VerifyEmail handles email verification requests
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
type ResendVerificationRequest struct {
	Email string `json:"email"`
}

// ResendVerification handles resend verification email requests
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
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken handles refresh token requests
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
type RevokeTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RevokeToken handles token revocation (logout)
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
