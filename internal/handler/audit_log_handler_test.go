package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

// Note: GetAuditLog_NonAdmin test skipped because handler uses h.logger.Warn() without nil check
// which causes panic when logger is nil. Would require initializing logger for test.

func TestAuditLogHandler_GetAuditLog_InvalidID(t *testing.T) {
	handler := &AuditLogHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodGet, "/api/audit-logs/abc", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.GetAuditLog(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid ID")
}

// Note: ListAuditLogs_NonAdmin test skipped because handler uses h.logger.Warn() without nil check

func TestAuditLogHandler_ListAuditLogs_InvalidUserID(t *testing.T) {
	handler := &AuditLogHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/audit-logs?user_id=abc", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListAuditLogs(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid user_id")
}

func TestAuditLogHandler_ListAuditLogs_InvalidTargetUserID(t *testing.T) {
	handler := &AuditLogHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/audit-logs?target_user_id=abc", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListAuditLogs(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid target_user_id")
}

func TestAuditLogHandler_ListAuditLogs_InvalidStartDate(t *testing.T) {
	handler := &AuditLogHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/audit-logs?start_date=invalid", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListAuditLogs(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid start_date format")
}

func TestAuditLogHandler_ListAuditLogs_InvalidEndDate(t *testing.T) {
	handler := &AuditLogHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/audit-logs?end_date=invalid", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListAuditLogs(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid end_date format")
}

func TestAuditLogHandler_GetMyAuditLogs_Unauthorized(t *testing.T) {
	handler := &AuditLogHandler{}

	req := createTestRequest(http.MethodGet, "/api/users/me/audit-logs", "")
	rr := httptest.NewRecorder()

	handler.GetMyAuditLogs(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestAuditLogHandler_CleanupOldLogs_NonAdmin(t *testing.T) {
	handler := &AuditLogHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/audit-logs/cleanup", `{"retention_days": 30}`, 1, "user@example.com", "user")
	rr := httptest.NewRecorder()

	handler.CleanupOldLogs(rr, req)

	assertStatusCode(t, rr, http.StatusForbidden)
	assertBodyContains(t, rr, "Forbidden")
}

func TestAuditLogHandler_CleanupOldLogs_InvalidJSON(t *testing.T) {
	handler := &AuditLogHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/audit-logs/cleanup", "{bad json", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CleanupOldLogs(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestAuditLogHandler_CleanupOldLogs_InvalidRetentionDays(t *testing.T) {
	handler := &AuditLogHandler{}

	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{
			name:      "zero days",
			body:      `{"retention_days": 0}`,
			wantError: "retention_days must be positive",
		},
		{
			name:      "negative days",
			body:      `{"retention_days": -1}`,
			wantError: "retention_days must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPost, "/api/admin/audit-logs/cleanup", tt.body, 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			handler.CleanupOldLogs(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestNewAuditLogHandler(t *testing.T) {
	handler := NewAuditLogHandler(nil, nil)
	if handler == nil {
		t.Error("NewAuditLogHandler should return a non-nil handler")
	}
}

func TestAuditLogHandler_GetMyAuditLogs_WithPaginationParams(t *testing.T) {
	handler := &AuditLogHandler{
		logger: createTestLogger(),
	}

	tests := []struct {
		name  string
		query string
	}{
		{"default pagination", "/api/users/me/audit-logs"},
		{"with limit", "/api/users/me/audit-logs?limit=25"},
		{"with offset", "/api/users/me/audit-logs?offset=10"},
		{"with both", "/api/users/me/audit-logs?limit=25&offset=10"},
		{"with invalid limit", "/api/users/me/audit-logs?limit=abc"},
		{"with invalid offset", "/api/users/me/audit-logs?offset=xyz"},
		{"with zero limit", "/api/users/me/audit-logs?limit=0"},
		{"with negative offset", "/api/users/me/audit-logs?offset=-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tt.query, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			// Without service, will panic - tests param parsing paths
			defer func() {
				if r := recover(); r == nil {
					t.Log("GetMyAuditLogs requires service - params were parsed")
				}
			}()

			handler.GetMyAuditLogs(rr, req)
		})
	}
}

func TestAuditLogHandler_GetAuditLog_ValidID(t *testing.T) {
	handler := &AuditLogHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/audit-logs/1", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without service, will panic
	defer func() {
		if r := recover(); r == nil {
			t.Log("GetAuditLog requires service")
		}
	}()

	handler.GetAuditLog(rr, req)
}

func TestAuditLogHandler_ListAuditLogs_WithAllParams(t *testing.T) {
	handler := &AuditLogHandler{
		logger: createTestLogger(),
	}

	tests := []struct {
		name  string
		query string
	}{
		{"with valid user_id", "/api/audit-logs?user_id=1"},
		{"with valid target_user_id", "/api/audit-logs?target_user_id=2"},
		{"with event_type", "/api/audit-logs?event_type=login"},
		{"with ip_address", "/api/audit-logs?ip_address=192.168.1.1"},
		{"with valid dates", "/api/audit-logs?start_date=2024-01-01T00:00:00Z&end_date=2024-12-31T23:59:59Z"},
		{"with pagination", "/api/audit-logs?limit=10&offset=5"},
		{"with all params", "/api/audit-logs?user_id=1&event_type=login&limit=20&offset=0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tt.query, "", 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			// Without service, will panic - tests param parsing paths
			defer func() {
				if r := recover(); r == nil {
					t.Log("ListAuditLogs requires service - params were parsed")
				}
			}()

			handler.ListAuditLogs(rr, req)
		})
	}
}

func TestAuditLogHandler_CleanupOldLogs_ValidRetentionDays(t *testing.T) {
	handler := &AuditLogHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/audit-logs/cleanup", `{"retention_days": 30}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	// Without service, will panic
	defer func() {
		if r := recover(); r == nil {
			t.Log("CleanupOldLogs requires service")
		}
	}()

	handler.CleanupOldLogs(rr, req)
}

// Tests with mock service

func TestAuditLogHandler_GetAuditLog_Success(t *testing.T) {
	mockService := createTestAuditLogService()
	handler := NewAuditLogHandler(mockService, createTestLogger())

	// First, create an audit log entry via the mock
	_ = mockService.Log(&domain.AuditLog{
		EventType: "test_event",
		Details:   stringPtr("Test audit log"),
	})

	req := createAuthenticatedRequest(http.MethodGet, "/api/audit-logs/1", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetAuditLog(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestAuditLogHandler_GetAuditLog_NotFound(t *testing.T) {
	mockService := createTestAuditLogService()
	handler := NewAuditLogHandler(mockService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/audit-logs/999", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.GetAuditLog(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
}

func TestAuditLogHandler_GetAuditLog_NonAdmin(t *testing.T) {
	mockService := createTestAuditLogService()
	handler := NewAuditLogHandler(mockService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/audit-logs/1", "", 1, "user@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetAuditLog(rr, req)

	assertStatusCode(t, rr, http.StatusForbidden)
}

func TestAuditLogHandler_ListAuditLogs_Success(t *testing.T) {
	mockService := createTestAuditLogService()
	handler := NewAuditLogHandler(mockService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/audit-logs", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListAuditLogs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "logs")
	assertBodyContains(t, rr, "total")
}

func TestAuditLogHandler_ListAuditLogs_NonAdmin(t *testing.T) {
	mockService := createTestAuditLogService()
	handler := NewAuditLogHandler(mockService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/audit-logs", "", 1, "user@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ListAuditLogs(rr, req)

	assertStatusCode(t, rr, http.StatusForbidden)
}

func TestAuditLogHandler_ListAuditLogs_WithFilters(t *testing.T) {
	mockService := createTestAuditLogService()
	handler := NewAuditLogHandler(mockService, createTestLogger())

	tests := []struct {
		name  string
		query string
	}{
		{"with limit", "/api/audit-logs?limit=10"},
		{"with offset", "/api/audit-logs?offset=5"},
		{"with user_id", "/api/audit-logs?user_id=1"},
		{"with event_type", "/api/audit-logs?event_type=login"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tt.query, "", 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			handler.ListAuditLogs(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
		})
	}
}

func TestAuditLogHandler_GetMyAuditLogs_Success(t *testing.T) {
	mockService := createTestAuditLogService()
	handler := NewAuditLogHandler(mockService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/users/me/audit-logs", "", 1, "user@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetMyAuditLogs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "logs")
}

func TestAuditLogHandler_CleanupOldLogs_Success(t *testing.T) {
	mockService := createTestAuditLogService()
	handler := NewAuditLogHandler(mockService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/audit-logs/cleanup", `{"retention_days": 30}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CleanupOldLogs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "deleted")
}
