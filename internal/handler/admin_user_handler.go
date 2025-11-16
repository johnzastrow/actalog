package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// AdminUserHandler handles admin user management operations
type AdminUserHandler struct {
	userService *service.UserService
	logger      *logger.Logger
}

// NewAdminUserHandler creates a new admin user handler
func NewAdminUserHandler(userService *service.UserService, logger *logger.Logger) *AdminUserHandler {
	return &AdminUserHandler{
		userService: userService,
		logger:      logger,
	}
}

// ListUsers handles GET /api/admin/users
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
		h.logger.Error("Failed to list users",
			"error", err.Error(),
		)
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
		h.logger.Error("Failed to unlock user account",
			"admin_user_id", adminUserID,
			"target_user_id", targetUserID,
			"error", err.Error(),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("User account unlocked",
		"admin_user_id", adminUserID,
		"target_user_id", targetUserID,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Account unlocked successfully",
	})
}

// DisableUser handles POST /api/admin/users/:id/disable
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
		h.logger.Error("Failed to disable user account",
			"admin_user_id", adminUserID,
			"target_user_id", targetUserID,
			"error", err.Error(),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("User account disabled",
		"admin_user_id", adminUserID,
		"target_user_id", targetUserID,
		"reason", request.Reason,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Account disabled successfully",
	})
}

// EnableUser handles POST /api/admin/users/:id/enable
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
		h.logger.Error("Failed to enable user account",
			"admin_user_id", adminUserID,
			"target_user_id", targetUserID,
			"error", err.Error(),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("User account enabled",
		"admin_user_id", adminUserID,
		"target_user_id", targetUserID,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Account enabled successfully",
	})
}

// ChangeUserRole handles PUT /api/admin/users/:id/role
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
	if request.Role != "user" && request.Role != "admin" {
		http.Error(w, "Role must be 'user' or 'admin'", http.StatusBadRequest)
		return
	}

	// Change the role
	if err := h.userService.ChangeUserRole(adminUserID, targetUserID, request.Role); err != nil {
		h.logger.Error("Failed to change user role",
			"admin_user_id", adminUserID,
			"target_user_id", targetUserID,
			"new_role", request.Role,
			"error", err.Error(),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("User role changed",
		"admin_user_id", adminUserID,
		"target_user_id", targetUserID,
		"new_role", request.Role,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User role changed successfully",
	})
}
