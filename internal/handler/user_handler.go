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
	userService       *service.UserService
	schedulingService *service.SchedulingService
	logger            *logger.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService, schedulingService *service.SchedulingService, l *logger.Logger) *UserHandler {
	return &UserHandler{
		userService:       userService,
		schedulingService: schedulingService,
		logger:            l,
	}
}

// UpdateProfileRequest represents a profile update request
// @Description User profile update request
type UpdateProfileRequest struct {
	Name     string `json:"name,omitempty" example:"John Doe"`
	Email    string `json:"email,omitempty" example:"john@example.com"`
	Birthday string `json:"birthday,omitempty" example:"1990-05-15"` // Format: "YYYY-MM-DD" or empty
}

// ProfileResponse represents a profile response
// @Description User profile data
type ProfileResponse struct {
	User interface{} `json:"user"`
}

// ChangePasswordRequest represents a password change request
// @Description Password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" example:"currentPassword123"`
	NewPassword string `json:"new_password" example:"newSecurePassword456"`
}

// UpdateProfile handles profile update requests
// @Summary      Update user profile
// @Description  Update the authenticated user's profile information
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body UpdateProfileRequest true "Profile updates"
// @Success      200 {object} ProfileResponse "Profile updated successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body or birthday format"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "User not found"
// @Failure      409 {object} ErrorResponse "Email already in use"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /users/profile [put]
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
// @Summary      Get user profile
// @Description  Retrieve the authenticated user's profile information
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} ProfileResponse "User profile"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "User not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /users/profile [get]
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

	// Check if user is a coach (for any organization)
	isCoach := false
	if h.schedulingService != nil {
		isCoach, _ = h.schedulingService.IsUserCoach(userID)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user":     user,
		"is_coach": isCoach,
	})
}

// UploadAvatar handles avatar image uploads
// @Summary      Upload avatar image
// @Description  Upload a new avatar image for the authenticated user (max 5MB)
// @Tags         users
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        avatar formData file true "Avatar image file"
// @Success      200 {object} ProfileResponse "Avatar uploaded successfully"
// @Failure      400 {object} ErrorResponse "No file provided, file too large, or invalid file type"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Failed to save avatar"
// @Router       /users/avatar [post]
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

	// Validate file type by sniffing the actual bytes — never trust the
	// client-supplied Content-Type header. http.DetectContentType implements
	// the WHATWG MIME Sniffing algorithm and reads up to 512 bytes.
	sniff := make([]byte, 512)
	n, err := file.Read(sniff)
	if err != nil && err != io.EOF {
		respondError(w, http.StatusBadRequest, "Failed to read uploaded file")
		return
	}
	detectedType := http.DetectContentType(sniff[:n])
	if !strings.HasPrefix(detectedType, "image/") {
		respondError(w, http.StatusBadRequest, "File must be an image")
		return
	}
	// Rewind so the subsequent io.Copy writes the whole file, not just the
	// bytes after the sniff buffer.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to process file")
		return
	}

	// Validate extension against an allowlist. Magic-byte sniffing alone does
	// not stop a JPEG payload uploaded under a .html filename, which the
	// static file server would later serve with a text/html content-type.
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[strings.ToLower(filepath.Ext(header.Filename))] {
		respondError(w, http.StatusBadRequest, "Only jpg, png, gif, and webp images are allowed")
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
// @Summary      Delete avatar image
// @Description  Remove the authenticated user's avatar image
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} ProfileResponse "Avatar deleted successfully"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Failed to delete avatar"
// @Router       /users/avatar [delete]
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
// @Summary      Change password
// @Description  Change the authenticated user's password
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body ChangePasswordRequest true "Current and new password"
// @Success      200 {object} MessageResponse "Password changed successfully"
// @Failure      400 {object} ErrorResponse "Invalid request or weak password"
// @Failure      401 {object} ErrorResponse "Unauthorized or incorrect current password"
// @Failure      500 {object} ErrorResponse "Failed to change password"
// @Router       /users/change-password [post]
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
