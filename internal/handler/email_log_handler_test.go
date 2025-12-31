package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johnzastrow/actalog/internal/service"
)

func TestNewEmailLogHandler(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	if handler == nil {
		t.Error("NewEmailLogHandler returned nil")
	}
}

func TestEmailLogHandler_ListEmailLogs_Unauthorized(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createTestRequest(http.MethodGet, "/api/admin/email-logs", "")
	rr := httptest.NewRecorder()

	handler.ListEmailLogs(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
}

func TestEmailLogHandler_ListEmailLogs_Success(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/email-logs", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListEmailLogs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, `"logs"`)
	assertBodyContains(t, rr, `"count"`)
}

func TestEmailLogHandler_ListEmailLogs_WithFilters(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/email-logs?type=test&success=true&limit=10", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListEmailLogs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, `"logs"`)
}

func TestEmailLogHandler_ListEmailLogs_WithDateFilters(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/email-logs?start_date=2024-01-01&end_date=2025-12-31", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListEmailLogs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestEmailLogHandler_ListEmailLogs_Pagination(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/email-logs?limit=1&offset=0", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListEmailLogs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, `"limit":1`)
	assertBodyContains(t, rr, `"offset":0`)
}

func TestEmailLogHandler_GetEmailLog_Unauthorized(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createTestRequest(http.MethodGet, "/api/admin/email-logs/1", "")
	rr := httptest.NewRecorder()

	handler.GetEmailLog(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
}

func TestEmailLogHandler_GetEmailLog_InvalidID(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/email-logs/abc", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	// Simulate chi URL param
	req = addChiURLParam(req, "id", "abc")

	handler.GetEmailLog(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid email log ID")
}

func TestEmailLogHandler_GetEmailLog_NotFound(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/email-logs/999", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	// Simulate chi URL param
	req = addChiURLParam(req, "id", "999")

	handler.GetEmailLog(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "Email log not found")
}

func TestEmailLogHandler_GetEmailLog_Success(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/email-logs/1", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	// Simulate chi URL param
	req = addChiURLParam(req, "id", "1")

	handler.GetEmailLog(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, `"recipient_email":"test@example.com"`)
}

func TestEmailLogHandler_GetEmailStats_Unauthorized(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createTestRequest(http.MethodGet, "/api/admin/email-logs/stats", "")
	rr := httptest.NewRecorder()

	handler.GetEmailStats(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
}

func TestEmailLogHandler_GetEmailStats_Success(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/email-logs/stats", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.GetEmailStats(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestEmailLogHandler_CleanupEmailLogs_Unauthorized(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createTestRequest(http.MethodPost, "/api/admin/email-logs/cleanup", "")
	rr := httptest.NewRecorder()

	handler.CleanupEmailLogs(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
}

func TestEmailLogHandler_CleanupEmailLogs_Success(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/email-logs/cleanup", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CleanupEmailLogs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, `"message":"Email logs cleaned up successfully"`)
	assertBodyContains(t, rr, `"deleted_count"`)
	assertBodyContains(t, rr, `"retention_days"`)
}

func TestEmailLogHandler_GetRecentFailures_Unauthorized(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createTestRequest(http.MethodGet, "/api/admin/email-logs/failures", "")
	rr := httptest.NewRecorder()

	handler.GetRecentFailures(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
}

func TestEmailLogHandler_GetRecentFailures_Success(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/email-logs/failures", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.GetRecentFailures(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, `"failures"`)
	assertBodyContains(t, rr, `"count"`)
}

func TestEmailLogHandler_GetRecentFailures_WithLimit(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailLogHandler(emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/email-logs/failures?limit=5", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.GetRecentFailures(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}
