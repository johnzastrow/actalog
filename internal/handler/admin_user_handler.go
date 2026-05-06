package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// adminUserServiceIface is the package-private surface AdminUserHandler depends
// on. The concrete *service.AdminUserService satisfies it; tests use a stub.
// Keep this list narrow: only methods this handler actually invokes.
type adminUserServiceIface interface {
	CreateUser(actorID int64, fields service.CreateUserFields) (*domain.User, error)
	UpdateProfile(actorID, targetID int64, fields service.ProfileUpdateFields, ifMatchUpdatedAt time.Time) (*domain.User, error)
	ForcePasswordReset(actorID, targetID int64) error
}

// AdminUserHandler handles admin user management operations
type AdminUserHandler struct {
	userService      *service.UserService
	adminUserService adminUserServiceIface
	logger           *logger.Logger
}

// NewAdminUserHandler creates a new admin user handler
func NewAdminUserHandler(userService *service.UserService, adminUserService adminUserServiceIface, logger *logger.Logger) *AdminUserHandler {
	return &AdminUserHandler{
		userService:      userService,
		adminUserService: adminUserService,
		logger:           logger,
	}
}

// ListUsers handles GET /api/admin/users
// @Summary      List users (Admin)
// @Description  Get a paginated list of all users
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Max results (default 50)"
// @Param        offset query int false "Skip N results (default 0)"
// @Success      200 {object} map[string]interface{} "Users list with total count"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/users [get]
func (h *AdminUserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get users
	users, total, err := h.userService.ListUsers(limit, offset)
	if err != nil {
		h.logger.Error("Failed to list users: %v", err)
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UnlockUser handles POST /api/admin/users/:id/unlock
// @Summary      Unlock user account (Admin)
// @Description  Unlock a user account that was locked due to failed login attempts
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200 {object} MessageResponse "Success message"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/users/{id}/unlock [post]
func (h *AdminUserHandler) UnlockUser(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get target user ID from URL
	targetUserIDStr := chi.URLParam(r, "id")
	targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Unlock the account
	if err := h.userService.UnlockAccount(adminUserID, targetUserID); err != nil {
		h.logger.Error("Failed to unlock user account: admin_user_id=%d target_user_id=%d error=%v", adminUserID, targetUserID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("User account unlocked: admin_user_id=%d target_user_id=%d", adminUserID, targetUserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Account unlocked successfully",
	})
}

// DisableUser handles POST /api/admin/users/:id/disable
// @Summary      Disable user account (Admin)
// @Description  Disable a user account with optional reason
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Param        request body object false "Optional disable reason"
// @Success      200 {object} MessageResponse "Success message"
// @Failure      400 {object} ErrorResponse "Invalid user ID or cannot disable"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/users/{id}/disable [post]
func (h *AdminUserHandler) DisableUser(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get target user ID from URL
	targetUserIDStr := chi.URLParam(r, "id")
	targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse request body for reason
	var request struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// Reason is optional
		request.Reason = ""
	}

	// Disable the account
	if err := h.userService.DisableAccount(adminUserID, targetUserID, request.Reason); err != nil {
		h.logger.Error("Failed to disable user account: admin_user_id=%d target_user_id=%d error=%v", adminUserID, targetUserID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("User account disabled: admin_user_id=%d target_user_id=%d reason=%s", adminUserID, targetUserID, request.Reason)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Account disabled successfully",
	})
}

// EnableUser handles POST /api/admin/users/:id/enable
// @Summary      Enable user account (Admin)
// @Description  Re-enable a previously disabled user account
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200 {object} MessageResponse "Success message"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/users/{id}/enable [post]
func (h *AdminUserHandler) EnableUser(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get target user ID from URL
	targetUserIDStr := chi.URLParam(r, "id")
	targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Enable the account
	if err := h.userService.EnableAccount(adminUserID, targetUserID); err != nil {
		h.logger.Error("Failed to enable user account: admin_user_id=%d target_user_id=%d error=%v", adminUserID, targetUserID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("User account enabled: admin_user_id=%d target_user_id=%d", adminUserID, targetUserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Account enabled successfully",
	})
}

// ChangeUserRole handles PUT /api/admin/users/:id/role
// @Summary      Change user role (Admin)
// @Description  Change a user's role to 'user' or 'admin'
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Param        request body object true "New role (user or admin)"
// @Success      200 {object} MessageResponse "Success message"
// @Failure      400 {object} ErrorResponse "Invalid user ID or role"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/users/{id}/role [put]
func (h *AdminUserHandler) ChangeUserRole(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get target user ID from URL
	targetUserIDStr := chi.URLParam(r, "id")
	targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var request struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate role
	if request.Role != "athlete" && request.Role != "coach" && request.Role != "admin" {
		http.Error(w, "Role must be 'athlete', 'coach', or 'admin'", http.StatusBadRequest)
		return
	}

	// Change the role
	if err := h.userService.ChangeUserRole(adminUserID, targetUserID, request.Role); err != nil {
		h.logger.Error("Failed to change user role: admin_user_id=%d target_user_id=%d new_role=%s error=%v", adminUserID, targetUserID, request.Role, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("User role changed: admin_user_id=%d target_user_id=%d new_role=%s", adminUserID, targetUserID, request.Role)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User role changed successfully",
	})
}

// ToggleEmailVerification handles POST /api/admin/users/:id/toggle-email-verification
// @Summary      Toggle email verification (Admin)
// @Description  Manually set a user's email verification status
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Param        request body object true "Verified status (true/false)"
// @Success      200 {object} MessageResponse "Success message"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/users/{id}/toggle-email-verification [post]
func (h *AdminUserHandler) ToggleEmailVerification(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get target user ID from URL
	targetUserIDStr := chi.URLParam(r, "id")
	targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var request struct {
		Verified bool `json:"verified"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Toggle email verification
	if err := h.userService.SetEmailVerification(adminUserID, targetUserID, request.Verified); err != nil {
		h.logger.Error("Failed to toggle email verification: admin_user_id=%d target_user_id=%d verified=%v error=%v", adminUserID, targetUserID, request.Verified, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Email verification toggled: admin_user_id=%d target_user_id=%d verified=%v", adminUserID, targetUserID, request.Verified)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Email verification updated successfully",
	})
}

// GetUserDetails handles GET /api/admin/users/:id
// @Summary      Get user details (Admin)
// @Description  Get detailed information about a specific user
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200 {object} map[string]interface{} "User details"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /admin/users/{id} [get]
func (h *AdminUserHandler) GetUserDetails(w http.ResponseWriter, r *http.Request) {
	// Get target user ID from URL
	targetUserIDStr := chi.URLParam(r, "id")
	targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get user with admin details
	user, err := h.userService.GetUserByIDWithAdminDetails(targetUserID)
	if err != nil {
		h.logger.Error("Failed to get user details: target_user_id=%d error=%v", targetUserID, err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// DeleteUser handles DELETE /api/admin/users/:id
// @Summary      Delete user (Admin)
// @Description  Permanently delete a user and all their data
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200 {object} MessageResponse "Success message"
// @Failure      400 {object} ErrorResponse "Invalid user ID or cannot delete"
// @Router       /admin/users/{id} [delete]
func (h *AdminUserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get target user ID from URL
	targetUserIDStr := chi.URLParam(r, "id")
	targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Delete the user
	if err := h.userService.DeleteUser(adminUserID, targetUserID); err != nil {
		h.logger.Error("Failed to delete user: admin_user_id=%d target_user_id=%d error=%v", adminUserID, targetUserID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("User deleted: admin_user_id=%d target_user_id=%d", adminUserID, targetUserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User deleted successfully",
	})
}

// updateProfileRequest is the PATCH /api/admin/users/{id} request body.
// UpdatedAt is required; it is compared against the stored row timestamp for
// optimistic concurrency control. Only present fields are applied.
type updateProfileRequest struct {
	Name          *string    `json:"name,omitempty"`
	Email         *string    `json:"email,omitempty"`
	Birthday      *time.Time `json:"birthday,omitempty"`
	EmailVerified *bool      `json:"email_verified,omitempty"`
	UpdatedAt     time.Time  `json:"updated_at"` // required (precondition)
}

// UpdateProfile handles PATCH /api/admin/users/{id}.
//
//   - Optimistic concurrency: client must send updated_at matching current row.
//   - Partial: only fields present in body are updated.
//   - L2 enforcement: service returns ErrProtectedUser → 403 via WriteError.
//   - On success: 200 with the updated User JSON.
func (h *AdminUserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	targetID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusNotFound, "Not found")
		return
	}
	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.UpdatedAt.IsZero() {
		respondError(w, http.StatusBadRequest, "updated_at must be a valid non-zero timestamp")
		return
	}
	user, err := h.adminUserService.UpdateProfile(actorID, targetID,
		service.ProfileUpdateFields{
			Name:          req.Name,
			Email:         req.Email,
			Birthday:      req.Birthday,
			EmailVerified: req.EmailVerified,
		},
		req.UpdatedAt,
	)
	if err != nil {
		WriteError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, user)
}

// ForcePasswordReset handles POST /api/admin/users/{id}/force-password-reset.
//
//   - Calls service.ForcePasswordReset (sends email + revokes refresh tokens).
//   - L2 enforcement: protected target → 403.
//   - On success: 204 No Content.
func (h *AdminUserHandler) ForcePasswordReset(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	targetID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusNotFound, "Not found")
		return
	}
	if err := h.adminUserService.ForcePasswordReset(actorID, targetID); err != nil {
		WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// createUserRequest is the POST /api/admin/users request body.
type createUserRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	EmailVerified *bool  `json:"email_verified,omitempty"` // nil → default true
}

// CreateUser handles POST /api/admin/users — admin creates a new user account.
//
//   - 201 with the created user JSON on success
//   - 400 invalid_input on validation failure
//   - 409 duplicate_email if email already exists
//   - 401/403 from auth middleware
func (h *AdminUserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	// Default EmailVerified=true when the caller omits the field.
	emailVerified := true
	if req.EmailVerified != nil {
		emailVerified = *req.EmailVerified
	}
	user, err := h.adminUserService.CreateUser(actorID, service.CreateUserFields{
		Email:         req.Email,
		Password:      req.Password,
		Name:          req.Name,
		Role:          req.Role,
		EmailVerified: emailVerified,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, user)
}
