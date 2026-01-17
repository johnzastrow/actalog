package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// NotificationLikeHandler handles notification like endpoints
type NotificationLikeHandler struct {
	likeService *service.NotificationLikeService
}

// NewNotificationLikeHandler creates a new notification like handler
func NewNotificationLikeHandler(likeService *service.NotificationLikeService) *NotificationLikeHandler {
	return &NotificationLikeHandler{
		likeService: likeService,
	}
}

// LikeNotification handles POST /api/notifications/{id}/like
// @Summary      Like notification
// @Description  Like a notification (e.g., PR announcement)
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Notification ID"
// @Success      201 {object} MessageResponse "Notification liked"
// @Failure      400 {object} ErrorResponse "Invalid notification ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      409 {object} ErrorResponse "Already liked"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /notifications/{id}/like [post]
func (h *NotificationLikeHandler) LikeNotification(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get notification ID from URL parameter
	notificationIDStr := chi.URLParam(r, "id")
	notificationID, err := strconv.ParseInt(notificationIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	// Create the like
	if err := h.likeService.LikeNotification(notificationID, userID); err != nil {
		if err.Error() == "user has already liked this notification" {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to like notification")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Notification liked"})
}

// UnlikeNotification handles DELETE /api/notifications/{id}/like
// @Summary      Unlike notification
// @Description  Remove like from a notification
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Notification ID"
// @Success      204 "Like removed"
// @Failure      400 {object} ErrorResponse "Invalid notification ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /notifications/{id}/like [delete]
func (h *NotificationLikeHandler) UnlikeNotification(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get notification ID from URL parameter
	notificationIDStr := chi.URLParam(r, "id")
	notificationID, err := strconv.ParseInt(notificationIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	// Remove the like
	if err := h.likeService.UnlikeNotification(notificationID, userID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to unlike notification")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetNotificationLikes handles GET /api/notifications/{id}/likes
// @Summary      Get notification likes
// @Description  Get all likes for a notification
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Notification ID"
// @Success      200 {object} map[string]interface{} "Likes list with count"
// @Failure      400 {object} ErrorResponse "Invalid notification ID"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /notifications/{id}/likes [get]
func (h *NotificationLikeHandler) GetNotificationLikes(w http.ResponseWriter, r *http.Request) {
	// Get notification ID from URL parameter
	notificationIDStr := chi.URLParam(r, "id")
	notificationID, err := strconv.ParseInt(notificationIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	// Get likes
	likes, err := h.likeService.GetNotificationLikes(notificationID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get likes")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"likes": likes,
		"count": len(likes),
	})
}
