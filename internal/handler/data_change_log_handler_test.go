package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDataChangeLogHandler_GetDataChangeLog_NonAdmin(t *testing.T) {
	handler := &DataChangeLogHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs/1", "", 1, "user@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetDataChangeLog(rr, req)

	assertStatusCode(t, rr, http.StatusForbidden)
	assertBodyContains(t, rr, "Forbidden")
}

func TestDataChangeLogHandler_GetDataChangeLog_InvalidID(t *testing.T) {
	handler := &DataChangeLogHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs/abc", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.GetDataChangeLog(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid ID")
}

func TestDataChangeLogHandler_ListDataChangeLogs_NonAdmin(t *testing.T) {
	handler := &DataChangeLogHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs", "", 1, "user@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ListDataChangeLogs(rr, req)

	assertStatusCode(t, rr, http.StatusForbidden)
	assertBodyContains(t, rr, "Forbidden")
}

func TestDataChangeLogHandler_ListDataChangeLogs_InvalidEntityID(t *testing.T) {
	handler := &DataChangeLogHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs?entity_id=abc", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListDataChangeLogs(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid entity_id")
}

func TestDataChangeLogHandler_ListDataChangeLogs_InvalidUserID(t *testing.T) {
	handler := &DataChangeLogHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs?user_id=abc", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListDataChangeLogs(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid user_id")
}

func TestDataChangeLogHandler_ListDataChangeLogs_InvalidStartDate(t *testing.T) {
	handler := &DataChangeLogHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs?start_date=invalid", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListDataChangeLogs(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid start_date format")
}

func TestDataChangeLogHandler_ListDataChangeLogs_InvalidEndDate(t *testing.T) {
	handler := &DataChangeLogHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs?end_date=invalid", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListDataChangeLogs(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid end_date format")
}

func TestDataChangeLogHandler_GetEntityHistory_NonAdmin(t *testing.T) {
	handler := &DataChangeLogHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs/entity/user/1", "", 1, "user@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetEntityHistory(rr, req)

	assertStatusCode(t, rr, http.StatusForbidden)
	assertBodyContains(t, rr, "Forbidden")
}

func TestDataChangeLogHandler_GetEntityHistory_InvalidEntityID(t *testing.T) {
	handler := &DataChangeLogHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs/entity/user/abc", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.GetEntityHistory(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid entity_id")
}

func TestDataChangeLogHandler_CleanupOldLogs_NonAdmin(t *testing.T) {
	handler := &DataChangeLogHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/data-change-logs/cleanup", `{"retention_days": 30}`, 1, "user@example.com", "user")
	rr := httptest.NewRecorder()

	handler.CleanupOldLogs(rr, req)

	assertStatusCode(t, rr, http.StatusForbidden)
	assertBodyContains(t, rr, "Forbidden")
}

func TestDataChangeLogHandler_CleanupOldLogs_InvalidJSON(t *testing.T) {
	handler := &DataChangeLogHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/data-change-logs/cleanup", "{bad json", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CleanupOldLogs(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestDataChangeLogHandler_CleanupOldLogs_InvalidRetentionDays(t *testing.T) {
	handler := &DataChangeLogHandler{}

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
			req := createAuthenticatedRequest(http.MethodPost, "/api/admin/data-change-logs/cleanup", tt.body, 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			handler.CleanupOldLogs(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestNewDataChangeLogHandler(t *testing.T) {
	handler := NewDataChangeLogHandler(nil, nil)
	if handler == nil {
		t.Error("NewDataChangeLogHandler should return a non-nil handler")
	}
}

// Removed 11 panic-expectation tests:
// - TestDataChangeLogHandler_GetDataChangeLog_ValidIDNilService
// - TestDataChangeLogHandler_ListDataChangeLogs_WithFilters (7 subtests)
// - TestDataChangeLogHandler_ListDataChangeLogs_ValidDateFormats (5 subtests)
// - TestDataChangeLogHandler_ListDataChangeLogs_ValidEntityID
// - TestDataChangeLogHandler_ListDataChangeLogs_ValidUserID
// - TestDataChangeLogHandler_GetEntityHistory_ValidInputNilService
// - TestDataChangeLogHandler_GetEntityHistory_WithPagination (4 subtests)
// - TestDataChangeLogHandler_GetEntityHistory_DifferentEntityTypes (4 subtests)
// - TestDataChangeLogHandler_CleanupOldLogs_ValidInputNilService
// - TestDataChangeLogHandler_CleanupOldLogs_DifferentRetentionDays (4 subtests)
// - TestDataChangeLogHandler_GetDataChangeLog_DifferentIDs (3 subtests)
// These tests verified nil pointer panics, not business logic.

// Test for GetDataChangeLog with logger

func TestDataChangeLogHandler_GetDataChangeLog_WithLogger(t *testing.T) {
	handler := &DataChangeLogHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs/1", "", 1, "user@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetDataChangeLog(rr, req)

	assertStatusCode(t, rr, http.StatusForbidden)
}

// Test for ListDataChangeLogs with logger

func TestDataChangeLogHandler_ListDataChangeLogs_WithLogger(t *testing.T) {
	handler := &DataChangeLogHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs", "", 1, "user@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ListDataChangeLogs(rr, req)

	assertStatusCode(t, rr, http.StatusForbidden)
}
