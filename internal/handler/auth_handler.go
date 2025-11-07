package handler

import (
	"encoding/json"
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
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
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
		respondError(w, http.StatusBadRequest, "Invalid request body")
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
		respondError(w, http.StatusBadRequest, "Invalid request body")
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
			respondError(w, http.StatusInternalServerError, "Failed to login")
		}
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
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
