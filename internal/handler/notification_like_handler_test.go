package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotificationLikeHandler_LikeNotification_Unauthorized(t *testing.T) {
	handler := &NotificationLikeHandler{}

	req := createTestRequest(http.MethodPost, "/api/notifications/1/like", "")
	rr := httptest.NewRecorder()

	handler.LikeNotification(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestNotificationLikeHandler_LikeNotification_InvalidID(t *testing.T) {
	handler := &NotificationLikeHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodPost, "/api/notifications/abc/like", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.LikeNotification(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid notification ID")
}

func TestNotificationLikeHandler_UnlikeNotification_Unauthorized(t *testing.T) {
	handler := &NotificationLikeHandler{}

	req := createTestRequest(http.MethodDelete, "/api/notifications/1/like", "")
	rr := httptest.NewRecorder()

	handler.UnlikeNotification(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestNotificationLikeHandler_UnlikeNotification_InvalidID(t *testing.T) {
	handler := &NotificationLikeHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodDelete, "/api/notifications/abc/like", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UnlikeNotification(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid notification ID")
}

func TestNotificationLikeHandler_GetNotificationLikes_InvalidID(t *testing.T) {
	handler := &NotificationLikeHandler{}

	// chi.URLParam returns empty string without router context
	req := createTestRequest(http.MethodGet, "/api/notifications/abc/likes", "")
	rr := httptest.NewRecorder()

	handler.GetNotificationLikes(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid notification ID")
}

func TestNewNotificationLikeHandler(t *testing.T) {
	handler := NewNotificationLikeHandler(nil)
	if handler == nil {
		t.Error("NewNotificationLikeHandler should return a non-nil handler")
	}
}

// Tests with mock service

func TestNotificationLikeHandler_LikeNotification_Success(t *testing.T) {
	likeService := createTestNotificationLikeService()
	handler := NewNotificationLikeHandler(likeService)

	req := createAuthenticatedRequest(http.MethodPost, "/api/notifications/1/like", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.LikeNotification(rr, req)

	assertStatusCode(t, rr, http.StatusCreated)
	assertBodyContains(t, rr, "liked")
}

func TestNotificationLikeHandler_LikeNotification_AlreadyLiked(t *testing.T) {
	likeService := createTestNotificationLikeService()
	handler := NewNotificationLikeHandler(likeService)

	// First like
	req1 := createAuthenticatedRequest(http.MethodPost, "/api/notifications/1/like", "", 1, "admin@example.com", "admin")
	req1 = addChiURLParam(req1, "id", "1")
	rr1 := httptest.NewRecorder()
	handler.LikeNotification(rr1, req1)
	assertStatusCode(t, rr1, http.StatusCreated)

	// Second like - should fail with conflict
	req2 := createAuthenticatedRequest(http.MethodPost, "/api/notifications/1/like", "", 1, "admin@example.com", "admin")
	req2 = addChiURLParam(req2, "id", "1")
	rr2 := httptest.NewRecorder()
	handler.LikeNotification(rr2, req2)

	assertStatusCode(t, rr2, http.StatusConflict)
	assertBodyContains(t, rr2, "already liked")
}

func TestNotificationLikeHandler_UnlikeNotification_Success(t *testing.T) {
	likeService := createTestNotificationLikeService()
	handler := NewNotificationLikeHandler(likeService)

	// First like
	req1 := createAuthenticatedRequest(http.MethodPost, "/api/notifications/1/like", "", 1, "admin@example.com", "admin")
	req1 = addChiURLParam(req1, "id", "1")
	rr1 := httptest.NewRecorder()
	handler.LikeNotification(rr1, req1)

	// Then unlike
	req2 := createAuthenticatedRequest(http.MethodDelete, "/api/notifications/1/like", "", 1, "admin@example.com", "admin")
	req2 = addChiURLParam(req2, "id", "1")
	rr2 := httptest.NewRecorder()
	handler.UnlikeNotification(rr2, req2)

	assertStatusCode(t, rr2, http.StatusNoContent)
}

func TestNotificationLikeHandler_GetNotificationLikes_Success(t *testing.T) {
	likeService := createTestNotificationLikeService()
	handler := NewNotificationLikeHandler(likeService)

	req := createTestRequest(http.MethodGet, "/api/notifications/1/likes", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetNotificationLikes(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "likes")
	assertBodyContains(t, rr, "count")
}

// Removed 3 panic-expectation tests:
// - TestNotificationLikeHandler_LikeNotification_ValidIDNilService
// - TestNotificationLikeHandler_GetNotificationLikes_ValidIDNilService
// - TestNotificationLikeHandler_UnlikeNotification_ValidIDNilService
// These tests verified nil pointer panics, not business logic.
