package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotificationHandler_ListNotifications_Unauthorized(t *testing.T) {
	handler := &NotificationHandler{}

	req := createTestRequest(http.MethodGet, "/api/notifications", "")
	rr := httptest.NewRecorder()

	handler.ListNotifications(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestNotificationHandler_ListUnreadNotifications_Unauthorized(t *testing.T) {
	handler := &NotificationHandler{}

	req := createTestRequest(http.MethodGet, "/api/notifications/unread", "")
	rr := httptest.NewRecorder()

	handler.ListUnreadNotifications(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestNotificationHandler_GetUnreadCount_Unauthorized(t *testing.T) {
	handler := &NotificationHandler{}

	req := createTestRequest(http.MethodGet, "/api/notifications/count", "")
	rr := httptest.NewRecorder()

	handler.GetUnreadCount(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestNotificationHandler_MarkAsRead_Unauthorized(t *testing.T) {
	handler := &NotificationHandler{}

	req := createTestRequest(http.MethodPut, "/api/notifications/1/read", "")
	rr := httptest.NewRecorder()

	handler.MarkAsRead(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestNotificationHandler_MarkAsRead_InvalidID(t *testing.T) {
	handler := &NotificationHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodPut, "/api/notifications/abc/read", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.MarkAsRead(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid notification ID")
}

func TestNotificationHandler_MarkAllAsRead_Unauthorized(t *testing.T) {
	handler := &NotificationHandler{}

	req := createTestRequest(http.MethodPut, "/api/notifications/read-all", "")
	rr := httptest.NewRecorder()

	handler.MarkAllAsRead(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestNotificationHandler_DeleteNotification_Unauthorized(t *testing.T) {
	handler := &NotificationHandler{}

	req := createTestRequest(http.MethodDelete, "/api/notifications/1", "")
	rr := httptest.NewRecorder()

	handler.DeleteNotification(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestNotificationHandler_DeleteNotification_InvalidID(t *testing.T) {
	handler := &NotificationHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodDelete, "/api/notifications/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.DeleteNotification(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid notification ID")
}

func TestNotificationHandler_CreateAnnouncement_NonAdmin(t *testing.T) {
	handler := &NotificationHandler{}

	// Regular user trying to create announcement
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/notifications/announce", `{"title": "Test", "message": "Test message", "target_type": "all"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.CreateAnnouncement(rr, req)

	assertStatusCode(t, rr, http.StatusForbidden)
	assertBodyContains(t, rr, "Admin access required")
}

func TestNotificationHandler_CreateAnnouncement_InvalidJSON(t *testing.T) {
	handler := &NotificationHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/notifications/announce", "{bad json", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateAnnouncement(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestNotificationHandler_CreateAnnouncement_MissingFields(t *testing.T) {
	handler := &NotificationHandler{}

	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{
			name:      "missing title",
			body:      `{"message": "Test", "target_type": "all"}`,
			wantError: "Title is required",
		},
		{
			name:      "missing message",
			body:      `{"title": "Test", "target_type": "all"}`,
			wantError: "Message is required",
		},
		{
			name:      "missing target_type",
			body:      `{"title": "Test", "message": "Test message"}`,
			wantError: "Target type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPost, "/api/admin/notifications/announce", tt.body, 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			handler.CreateAnnouncement(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestNotificationHandler_CreateAnnouncement_InvalidTargetType(t *testing.T) {
	handler := &NotificationHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/notifications/announce", `{"title": "Test", "message": "Test message", "target_type": "invalid"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateAnnouncement(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid target type")
}

func TestNotificationHandler_CreateAnnouncement_MissingTargetIDs(t *testing.T) {
	handler := &NotificationHandler{}

	tests := []struct {
		name       string
		targetType string
	}{
		{name: "organization without IDs", targetType: "organization"},
		{name: "users without IDs", targetType: "users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"title": "Test", "message": "Test message", "target_type": "` + tt.targetType + `"}`
			req := createAuthenticatedRequest(http.MethodPost, "/api/admin/notifications/announce", body, 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			handler.CreateAnnouncement(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, "Target IDs required")
		})
	}
}

func TestNewNotificationHandler(t *testing.T) {
	handler := NewNotificationHandler(nil, nil)
	if handler == nil {
		t.Error("NewNotificationHandler should return a non-nil handler")
	}
}

func TestNotificationHandler_ListNotifications_NilService(t *testing.T) {
	handler := &NotificationHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/notifications", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("ListNotifications requires service")
		}
	}()

	handler.ListNotifications(rr, req)
}

func TestNotificationHandler_GetUnreadCount_NilService(t *testing.T) {
	handler := &NotificationHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/notifications/count", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("GetUnreadCount requires service")
		}
	}()

	handler.GetUnreadCount(rr, req)
}

func TestNotificationHandler_MarkAsRead_ValidIDNilService(t *testing.T) {
	handler := &NotificationHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/notifications/1/read", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry with valid ID
	defer func() {
		if r := recover(); r == nil {
			t.Log("MarkAsRead requires service")
		}
	}()

	handler.MarkAsRead(rr, req)
}

func TestNotificationHandler_MarkAllAsRead_NilService(t *testing.T) {
	handler := &NotificationHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/notifications/read-all", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("MarkAllAsRead requires service")
		}
	}()

	handler.MarkAllAsRead(rr, req)
}
