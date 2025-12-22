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

// CreateAnnouncement handles POST /api/admin/notifications/announce (admin only)
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
