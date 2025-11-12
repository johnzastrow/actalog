package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// UploadAvatar handles avatar image uploads
func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form (max 5MB)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "File too large (max 5MB)")
		return
	}

	// Get file from form
	file, header, err := r.FormFile("avatar")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		respondError(w, http.StatusBadRequest, "File must be an image")
		return
	}

	// Create avatars directory if it doesn't exist
	avatarDir := "uploads/avatars"
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		if h.logger != nil {
			h.logger.Error("action=upload_avatar outcome=failure user_id=%d error=failed_to_create_directory: %v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to save avatar")
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("user_%d_%d%s", userID, time.Now().Unix(), ext)
	filePath := filepath.Join(avatarDir, filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=upload_avatar outcome=failure user_id=%d error=failed_to_create_file: %v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to save avatar")
		return
	}
	defer dst.Close()

	// Copy file data
	if _, err := io.Copy(dst, file); err != nil {
		if h.logger != nil {
			h.logger.Error("action=upload_avatar outcome=failure user_id=%d error=failed_to_copy_file: %v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to save avatar")
		return
	}

	// Get current user to check for old avatar
	user, err := h.userService.GetByID(userID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=upload_avatar outcome=failure user_id=%d error=failed_to_get_user: %v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	// Delete old avatar if it exists
	if user.ProfileImage != nil && *user.ProfileImage != "" {
		oldPath := *user.ProfileImage
		if strings.HasPrefix(oldPath, "/uploads/") {
			oldPath = "." + oldPath
		}
		if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
			if h.logger != nil {
				h.logger.Warn("action=upload_avatar outcome=warning user_id=%d error=failed_to_delete_old_avatar: %v", userID, err)
			}
		}
	}

	// Update user profile with new avatar URL
	avatarURL := "/uploads/avatars/" + filename
	user.ProfileImage = &avatarURL

	if err := h.userService.UpdateAvatar(userID, avatarURL); err != nil {
		if h.logger != nil {
			h.logger.Error("action=upload_avatar outcome=failure user_id=%d error=failed_to_update_avatar: %v", userID, err)
		}
		// Try to clean up uploaded file
		os.Remove(filePath)
		respondError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=upload_avatar outcome=success user_id=%d avatar_url=%s", userID, avatarURL)
	}

	respondJSON(w, http.StatusOK, ProfileResponse{
		User: user,
	})
}

// DeleteAvatar handles avatar image deletion
func (h *UserHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get current user
	user, err := h.userService.GetByID(userID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=delete_avatar outcome=failure user_id=%d error=failed_to_get_user: %v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete avatar")
		return
	}

	// Delete avatar file if it exists
	if user.ProfileImage != nil && *user.ProfileImage != "" {
		oldPath := *user.ProfileImage
		if strings.HasPrefix(oldPath, "/uploads/") {
			oldPath = "." + oldPath
		}
		if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
			if h.logger != nil {
				h.logger.Warn("action=delete_avatar outcome=warning user_id=%d error=failed_to_delete_file: %v", userID, err)
			}
		}
	}

	// Update user profile to remove avatar
	if err := h.userService.UpdateAvatar(userID, ""); err != nil {
		if h.logger != nil {
			h.logger.Error("action=delete_avatar outcome=failure user_id=%d error=failed_to_update_profile: %v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete avatar")
		return
	}

	user.ProfileImage = nil

	if h.logger != nil {
		h.logger.Info("action=delete_avatar outcome=success user_id=%d", userID)
	}

	respondJSON(w, http.StatusOK, ProfileResponse{
		User: user,
	})
}

// ChangePassword handles password change requests
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		respondError(w, http.StatusBadRequest, "Both old_password and new_password are required")
		return
	}

	if len(req.NewPassword) < 8 {
		respondError(w, http.StatusBadRequest, "New password must be at least 8 characters")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=change_password_attempt user_id=%d", userID)
	}

	if err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		if err == service.ErrInvalidCredentials {
			if h.logger != nil {
				h.logger.Warn("action=change_password outcome=failure user_id=%d reason=invalid_old_password", userID)
			}
			respondError(w, http.StatusUnauthorized, "Current password is incorrect")
			return
		}
		if h.logger != nil {
			h.logger.Error("action=change_password outcome=failure user_id=%d error=%v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to change password")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=change_password outcome=success user_id=%d", userID)
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}
