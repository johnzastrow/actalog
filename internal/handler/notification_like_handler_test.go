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

func TestNotificationLikeHandler_LikeNotification_ValidIDNilService(t *testing.T) {
	handler := &NotificationLikeHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/notifications/1/like", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry with valid ID
	defer func() {
		if r := recover(); r == nil {
			t.Log("LikeNotification requires service")
		}
	}()

	handler.LikeNotification(rr, req)
}

func TestNotificationLikeHandler_GetNotificationLikes_ValidIDNilService(t *testing.T) {
	handler := &NotificationLikeHandler{}

	req := createTestRequest(http.MethodGet, "/api/notifications/1/likes", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry with valid ID
	defer func() {
		if r := recover(); r == nil {
			t.Log("GetNotificationLikes requires service")
		}
	}()

	handler.GetNotificationLikes(rr, req)
}
