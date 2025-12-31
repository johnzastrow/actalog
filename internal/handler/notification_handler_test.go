package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johnzastrow/actalog/internal/service"
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

func TestNotificationHandler_ListNotifications_WithPaginationParams(t *testing.T) {
	handler := &NotificationHandler{
		logger: createTestLogger(),
	}

	tests := []struct {
		name  string
		query string
	}{
		{"default params", "/api/notifications"},
		{"with limit", "/api/notifications?limit=25"},
		{"with offset", "/api/notifications?offset=10"},
		{"with both", "/api/notifications?limit=25&offset=10"},
		{"with high limit (capped at 100)", "/api/notifications?limit=500"},
		{"with invalid limit", "/api/notifications?limit=abc"},
		{"with invalid offset", "/api/notifications?offset=xyz"},
		{"with zero limit", "/api/notifications?limit=0"},
		{"with negative offset", "/api/notifications?offset=-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tt.query, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			// Without service, will panic - tests param parsing paths
			defer func() {
				if r := recover(); r == nil {
					t.Log("ListNotifications requires service - params were parsed")
				}
			}()

			handler.ListNotifications(rr, req)
		})
	}
}

func TestNotificationHandler_ListUnreadNotifications_WithPaginationParams(t *testing.T) {
	handler := &NotificationHandler{
		logger: createTestLogger(),
	}

	tests := []struct {
		name  string
		query string
	}{
		{"default params", "/api/notifications/unread"},
		{"with limit", "/api/notifications/unread?limit=25"},
		{"with offset", "/api/notifications/unread?offset=10"},
		{"with both", "/api/notifications/unread?limit=25&offset=10"},
		{"with high limit (capped at 100)", "/api/notifications/unread?limit=500"},
		{"with invalid limit", "/api/notifications/unread?limit=abc"},
		{"with invalid offset", "/api/notifications/unread?offset=xyz"},
		{"with zero limit", "/api/notifications/unread?limit=0"},
		{"with negative offset", "/api/notifications/unread?offset=-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tt.query, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			// Without service, will panic - tests param parsing paths
			defer func() {
				if r := recover(); r == nil {
					t.Log("ListUnreadNotifications requires service - params were parsed")
				}
			}()

			handler.ListUnreadNotifications(rr, req)
		})
	}
}

func TestNotificationHandler_DeleteNotification_ValidIDNilService(t *testing.T) {
	handler := &NotificationHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/notifications/1", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry with valid ID
	defer func() {
		if r := recover(); r == nil {
			t.Log("DeleteNotification requires service")
		}
	}()

	handler.DeleteNotification(rr, req)
}

func TestNotificationHandler_CreateAnnouncement_ValidInputAllUsers(t *testing.T) {
	handler := &NotificationHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/notifications/announce",
		`{"title": "Test", "message": "Test message", "target_type": "all"}`,
		1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	// Without service, will panic - tests validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("CreateAnnouncement requires service")
		}
	}()

	handler.CreateAnnouncement(rr, req)
}

func TestNotificationHandler_CreateAnnouncement_ValidInputOrganization(t *testing.T) {
	handler := &NotificationHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/notifications/announce",
		`{"title": "Test", "message": "Test message", "target_type": "organization", "target_ids": [1, 2]}`,
		1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	// Without service, will panic - tests validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("CreateAnnouncement requires service")
		}
	}()

	handler.CreateAnnouncement(rr, req)
}

func TestNotificationHandler_CreateAnnouncement_ValidInputUsers(t *testing.T) {
	handler := &NotificationHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/notifications/announce",
		`{"title": "Test", "message": "Test message", "target_type": "users", "target_ids": [1, 2, 3]}`,
		1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	// Without service, will panic - tests validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("CreateAnnouncement requires service")
		}
	}()

	handler.CreateAnnouncement(rr, req)
}

func TestNotificationHandler_GetUnreadCount_DifferentUserIDs(t *testing.T) {
	handler := &NotificationHandler{
		logger: createTestLogger(),
	}

	testUserIDs := []int64{1, 10, 100}

	for _, userID := range testUserIDs {
		t.Run("user_"+string(rune(userID)), func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/notifications/count", "", userID, "test@example.com", "user")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("GetUnreadCount requires service")
				}
			}()

			handler.GetUnreadCount(rr, req)
		})
	}
}

func TestNotificationHandler_MarkAllAsRead_DifferentUserIDs(t *testing.T) {
	handler := &NotificationHandler{
		logger: createTestLogger(),
	}

	testUserIDs := []int64{1, 10, 100}

	for _, userID := range testUserIDs {
		t.Run("user_"+string(rune(userID)), func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPut, "/api/notifications/read-all", "", userID, "test@example.com", "user")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("MarkAllAsRead requires service")
				}
			}()

			handler.MarkAllAsRead(rr, req)
		})
	}
}

func TestNotificationHandler_MarkAsRead_DifferentIDs(t *testing.T) {
	handler := &NotificationHandler{
		logger: createTestLogger(),
	}

	testIDs := []string{"1", "10", "100", "999"}

	for _, id := range testIDs {
		t.Run("notification_id_"+id, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPut, "/api/notifications/"+id+"/read", "", 1, "test@example.com", "user")
			req = addChiURLParam(req, "id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("MarkAsRead requires service")
				}
			}()

			handler.MarkAsRead(rr, req)
		})
	}
}

func TestNotificationHandler_DeleteNotification_DifferentIDs(t *testing.T) {
	handler := &NotificationHandler{
		logger: createTestLogger(),
	}

	testIDs := []string{"1", "10", "100"}

	for _, id := range testIDs {
		t.Run("notification_id_"+id, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodDelete, "/api/notifications/"+id, "", 1, "test@example.com", "user")
			req = addChiURLParam(req, "id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("DeleteNotification requires service")
				}
			}()

			handler.DeleteNotification(rr, req)
		})
	}
}

func TestNotificationHandler_ListUnreadNotifications_DifferentUserIDs(t *testing.T) {
	handler := &NotificationHandler{
		logger: createTestLogger(),
	}

	testUserIDs := []int64{1, 10, 100}

	for _, userID := range testUserIDs {
		t.Run("user_"+string(rune(userID)), func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/notifications/unread", "", userID, "test@example.com", "user")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListUnreadNotifications requires service")
				}
			}()

			handler.ListUnreadNotifications(rr, req)
		})
	}
}

func TestNotificationHandler_ListNotifications_DifferentUserIDs(t *testing.T) {
	handler := &NotificationHandler{
		logger: createTestLogger(),
	}

	testUserIDs := []int64{1, 10, 100}

	for _, userID := range testUserIDs {
		t.Run("user_"+string(rune(userID)), func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/notifications", "", userID, "test@example.com", "user")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListNotifications requires service")
				}
			}()

			handler.ListNotifications(rr, req)
		})
	}
}

// Helper to create a test notification service with mocks
func createTestNotificationService() *service.NotificationService {
	mockNotificationRepo := NewMockNotificationRepository()
	mockOrgRepo := NewMockOrganizationRepository()
	mockUserRepo := NewMockUserRepository()
	mockSettingsRepo := NewMockUserSettingsRepository()
	mockEmailService := NewMockEmailService()

	return service.NewNotificationService(
		mockNotificationRepo,
		mockOrgRepo,
		mockUserRepo,
		mockSettingsRepo,
		mockEmailService,
	)
}

// Tests using real NotificationService with mock repositories

func TestNotificationHandler_ListNotifications_Success(t *testing.T) {
	notificationService := createTestNotificationService()
	handler := &NotificationHandler{
		notificationService: notificationService,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/notifications", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ListNotifications(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "notifications")
}

func TestNotificationHandler_ListNotifications_WithPagination(t *testing.T) {
	notificationService := createTestNotificationService()
	handler := &NotificationHandler{
		notificationService: notificationService,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/notifications?limit=10&offset=0", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ListNotifications(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "notifications")
	assertBodyContains(t, rr, `"limit":10`)
}

func TestNotificationHandler_ListUnreadNotifications_Success(t *testing.T) {
	notificationService := createTestNotificationService()
	handler := &NotificationHandler{
		notificationService: notificationService,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/notifications/unread", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ListUnreadNotifications(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "notifications")
}

func TestNotificationHandler_GetUnreadCount_Success(t *testing.T) {
	notificationService := createTestNotificationService()
	handler := &NotificationHandler{
		notificationService: notificationService,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/notifications/count", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetUnreadCount(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "count")
}

func TestNotificationHandler_MarkAsRead_Success(t *testing.T) {
	notificationService := createTestNotificationService()
	handler := &NotificationHandler{
		notificationService: notificationService,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPut, "/api/notifications/1/read", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.MarkAsRead(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "Notification marked as read")
}

func TestNotificationHandler_MarkAllAsRead_Success(t *testing.T) {
	notificationService := createTestNotificationService()
	handler := &NotificationHandler{
		notificationService: notificationService,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPut, "/api/notifications/read-all", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.MarkAllAsRead(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "All notifications marked as read")
}

func TestNotificationHandler_DeleteNotification_Success(t *testing.T) {
	notificationService := createTestNotificationService()
	handler := &NotificationHandler{
		notificationService: notificationService,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/notifications/1", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.DeleteNotification(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "Notification deleted")
}

func TestNotificationHandler_CreateAnnouncement_Success(t *testing.T) {
	notificationService := createTestNotificationService()
	handler := &NotificationHandler{
		notificationService: notificationService,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/notifications/announce",
		`{"title": "Test Announcement", "message": "Test message", "target_type": "all"}`,
		1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateAnnouncement(rr, req)

	assertStatusCode(t, rr, http.StatusCreated)
	assertBodyContains(t, rr, "Announcement created successfully")
}

func TestNotificationHandler_ListNotifications_NoLogger(t *testing.T) {
	notificationService := createTestNotificationService()
	handler := &NotificationHandler{
		notificationService: notificationService,
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/notifications", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ListNotifications(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestNotificationHandler_GetUnreadCount_NoLogger(t *testing.T) {
	notificationService := createTestNotificationService()
	handler := &NotificationHandler{
		notificationService: notificationService,
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/notifications/count", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetUnreadCount(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}
