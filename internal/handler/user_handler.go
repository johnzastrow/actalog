package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// UserHandler handles user profile endpoints
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Birthday string `json:"birthday,omitempty"` // Format: "YYYY-MM-DD" or empty
}

// ProfileResponse represents a profile response
type ProfileResponse struct {
	User interface{} `json:"user"`
}

// UpdateProfile handles profile update requests
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse birthday if provided
	var birthday *time.Time
	if req.Birthday != "" {
		parsedBirthday, err := time.Parse("2006-01-02", req.Birthday)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid birthday format. Use YYYY-MM-DD")
			return
		}
		birthday = &parsedBirthday
	}

	// Update profile
	user, err := h.userService.UpdateProfile(userID, req.Name, req.Email, birthday)
	if err != nil {
		switch err {
		case service.ErrEmailAlreadyExists:
			respondError(w, http.StatusConflict, "Email already in use")
		case service.ErrUserNotFound:
			respondError(w, http.StatusNotFound, "User not found")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to update profile")
		}
		return
	}

	respondJSON(w, http.StatusOK, ProfileResponse{
		User: user,
	})
}

// GetProfile retrieves the current user's profile
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user from service
	user, err := h.userService.GetByID(userID)
	if err != nil {
		if err == service.ErrUserNotFound {
			respondError(w, http.StatusNotFound, "User not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get profile")
		}
		return
	}

	respondJSON(w, http.StatusOK, ProfileResponse{
		User: user,
	})
}
