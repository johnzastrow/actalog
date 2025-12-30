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

// NotificationHandler handles notification endpoints
type NotificationHandler struct {
	notificationService *service.NotificationService
	logger              *logger.Logger
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService *service.NotificationService, l *logger.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		logger:              l,
	}
}

// ListNotifications handles GET /api/notifications
// @Summary      List notifications
// @Description  Retrieve all notifications for the authenticated user
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Max results (default 20, max 100)"
// @Param        offset query int false "Skip N results (default 0)"
// @Success      200 {object} map[string]interface{} "List of notifications"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /notifications [get]
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // Default limit
	offset := 0 // Default offset

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100 // Max limit
			}
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get notifications
	notifications, err := h.notificationService.GetNotifications(userID, limit, offset)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=list_notifications outcome=failure user_id=%d error=%v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve notifications")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=list_notifications outcome=success user_id=%d count=%d", userID, len(notifications))
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": notifications,
		"limit":         limit,
		"offset":        offset,
	})
}

// ListUnreadNotifications handles GET /api/notifications/unread
// @Summary      List unread notifications
// @Description  Retrieve only unread notifications for the authenticated user
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Max results (default 20, max 100)"
// @Param        offset query int false "Skip N results (default 0)"
// @Success      200 {object} map[string]interface{} "List of unread notifications"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /notifications/unread [get]
func (h *NotificationHandler) ListUnreadNotifications(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // Default limit
	offset := 0 // Default offset

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100 // Max limit
			}
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get unread notifications
	notifications, err := h.notificationService.GetUnreadNotifications(userID, limit, offset)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=list_unread_notifications outcome=failure user_id=%d error=%v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve unread notifications")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=list_unread_notifications outcome=success user_id=%d count=%d", userID, len(notifications))
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": notifications,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetUnreadCount handles GET /api/notifications/count
// @Summary      Get unread notification count
// @Description  Get the count of unread notifications (for badge display)
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "Unread count"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /notifications/count [get]
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get unread count
	count, err := h.notificationService.GetUnreadCount(userID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=get_unread_count outcome=failure user_id=%d error=%v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to get unread count")
		return
	}

	if h.logger != nil {
		h.logger.Debug("action=get_unread_count outcome=success user_id=%d count=%d", userID, count)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"count": count,
	})
}

// MarkAsRead handles PUT /api/notifications/{id}/read
// @Summary      Mark notification as read
// @Description  Mark a specific notification as read
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Notification ID"
// @Success      200 {object} MessageResponse "Marked as read"
// @Failure      400 {object} ErrorResponse "Invalid notification ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get notification ID from URL
	notificationIDStr := chi.URLParam(r, "id")
	notificationID, err := strconv.ParseInt(notificationIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	// Mark as read
	if err := h.notificationService.MarkAsRead(notificationID); err != nil {
		if h.logger != nil {
			h.logger.Error("action=mark_as_read outcome=failure user_id=%d notification_id=%d error=%v", userID, notificationID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to mark notification as read")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=mark_as_read outcome=success user_id=%d notification_id=%d", userID, notificationID)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Notification marked as read",
	})
}

// MarkAllAsRead handles PUT /api/notifications/read-all
// @Summary      Mark all notifications as read
// @Description  Mark all notifications for the user as read
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} MessageResponse "All marked as read"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Mark all as read
	if err := h.notificationService.MarkAllAsRead(userID); err != nil {
		if h.logger != nil {
			h.logger.Error("action=mark_all_as_read outcome=failure user_id=%d error=%v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to mark all notifications as read")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=mark_all_as_read outcome=success user_id=%d", userID)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "All notifications marked as read",
	})
}

// DeleteNotification handles DELETE /api/notifications/{id}
// @Summary      Delete notification
// @Description  Delete a specific notification
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Notification ID"
// @Success      200 {object} MessageResponse "Notification deleted"
// @Failure      400 {object} ErrorResponse "Invalid notification ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get notification ID from URL
	notificationIDStr := chi.URLParam(r, "id")
	notificationID, err := strconv.ParseInt(notificationIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	// Delete notification
	if err := h.notificationService.DeleteNotification(notificationID); err != nil {
		if h.logger != nil {
			h.logger.Error("action=delete_notification outcome=failure user_id=%d notification_id=%d error=%v", userID, notificationID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete notification")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=delete_notification outcome=success user_id=%d notification_id=%d", userID, notificationID)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Notification deleted",
	})
}

// CreateAnnouncementRequest represents an announcement creation request
// @Description Admin announcement request
type CreateAnnouncementRequest struct {
	Title          string  `json:"title" example:"System Update"`
	Message        string  `json:"message" example:"Scheduled maintenance tonight"`
	TargetType     string  `json:"target_type" example:"all"` // "all", "organization", or "users"
	TargetIDs      []int64 `json:"target_ids"`                // org IDs or user IDs
	OrganizationID *int64  `json:"organization_id,omitempty"`
}

// CreateAnnouncement handles POST /api/admin/notifications/announce (admin only)
// @Summary      Create announcement (Admin)
// @Description  Send an announcement to all users, organization members, or specific users
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateAnnouncementRequest true "Announcement details"
// @Success      201 {object} MessageResponse "Announcement created"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      403 {object} ErrorResponse "Admin access required"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/notifications/announce [post]
func (h *NotificationHandler) CreateAnnouncement(w http.ResponseWriter, r *http.Request) {
	// Get user role from context (set by auth middleware)
	userRole, ok := middleware.GetUserRole(r.Context())
	if !ok || userRole != "admin" {
		respondError(w, http.StatusForbidden, "Admin access required")
		return
	}

	// Parse request body
	var req struct {
		Title          string  `json:"title"`
		Message        string  `json:"message"`
		TargetType     string  `json:"target_type"` // "all", "organization", or "users"
		TargetIDs      []int64 `json:"target_ids"`  // org IDs or user IDs
		OrganizationID *int64  `json:"organization_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "Title is required")
		return
	}
	if req.Message == "" {
		respondError(w, http.StatusBadRequest, "Message is required")
		return
	}
	if req.TargetType == "" {
		respondError(w, http.StatusBadRequest, "Target type is required")
		return
	}

	// Validate target type
	if req.TargetType != "all" && req.TargetType != "organization" && req.TargetType != "users" {
		respondError(w, http.StatusBadRequest, "Invalid target type (must be 'all', 'organization', or 'users')")
		return
	}

	// Validate target IDs for non-"all" target types
	if req.TargetType != "all" && len(req.TargetIDs) == 0 {
		respondError(w, http.StatusBadRequest, "Target IDs required for non-'all' target types")
		return
	}

	// Create announcement
	if err := h.notificationService.CreateAnnouncement(req.Title, req.Message, req.TargetType, req.TargetIDs, req.OrganizationID); err != nil {
		if h.logger != nil {
			h.logger.Error("action=create_announcement outcome=failure error=%v", err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to create announcement")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=create_announcement outcome=success target_type=%s title=%s", req.TargetType, req.Title)
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Announcement created successfully",
	})
}
