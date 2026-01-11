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

// Removed 4 panic-expectation tests:
// - TestAuditLogHandler_GetMyAuditLogs_WithPaginationParams (8 subtests)
// - TestAuditLogHandler_GetAuditLog_ValidID
// - TestAuditLogHandler_ListAuditLogs_WithAllParams (7 subtests)
// - TestAuditLogHandler_CleanupOldLogs_ValidRetentionDays
// These tests verified nil pointer panics, not business logic.

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

func TestAuditLogHandler_GetMyAuditLogs_WithPagination(t *testing.T) {
	mockService := createTestAuditLogService()
	handler := NewAuditLogHandler(mockService, createTestLogger())

	tests := []struct {
		name     string
		query    string
		wantCode int
	}{
		{"with valid limit", "?limit=10", http.StatusOK},
		{"with valid offset", "?offset=5", http.StatusOK},
		{"with both params", "?limit=10&offset=5", http.StatusOK},
		{"with invalid limit", "?limit=abc", http.StatusOK},
		{"with negative limit", "?limit=-1", http.StatusOK},
		{"with invalid offset", "?offset=abc", http.StatusOK},
		{"with negative offset", "?offset=-1", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/users/me/audit-logs"+tt.query, "", 1, "user@example.com", "user")
			rr := httptest.NewRecorder()

			handler.GetMyAuditLogs(rr, req)

			assertStatusCode(t, rr, tt.wantCode)
		})
	}
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
