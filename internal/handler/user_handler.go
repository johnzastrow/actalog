package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// UserHandler handles user profile endpoints
type UserHandler struct {
	userService *service.UserService
	logger      *logger.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService, l *logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      l,
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

	if h.logger != nil {
		h.logger.Info("action=update_profile_attempt user_id=%d name=%s email=%s", userID, req.Name, req.Email)
	}

	// Update profile
	user, err := h.userService.UpdateProfile(userID, req.Name, req.Email, birthday)
	if err != nil {
		switch err {
		case service.ErrEmailAlreadyExists:
			if h.logger != nil {
				h.logger.Warn("action=update_profile outcome=failure user_id=%d reason=email_exists email=%s", userID, req.Email)
			}
			respondError(w, http.StatusConflict, "Email already in use")
		case service.ErrUserNotFound:
			if h.logger != nil {
				h.logger.Warn("action=update_profile outcome=failure user_id=%d reason=not_found", userID)
			}
			respondError(w, http.StatusNotFound, "User not found")
		default:
			if h.logger != nil {
				h.logger.Error("action=update_profile outcome=failure user_id=%d error=%v", userID, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to update profile")
		}
		return
	}

	if h.logger != nil {
		h.logger.Info("action=update_profile outcome=success user_id=%d", userID)
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

	if h.logger != nil {
		h.logger.Info("action=get_profile user_id=%d", userID)
	}

	// Get user from service
	user, err := h.userService.GetByID(userID)
	if err != nil {
		if err == service.ErrUserNotFound {
			if h.logger != nil {
				h.logger.Warn("action=get_profile outcome=failure user_id=%d reason=not_found", userID)
			}
			respondError(w, http.StatusNotFound, "User not found")
		} else {
			if h.logger != nil {
				h.logger.Error("action=get_profile outcome=failure user_id=%d error=%v", userID, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to get profile")
		}
		return
	}

	respondJSON(w, http.StatusOK, ProfileResponse{
		User: user,
	})
}
