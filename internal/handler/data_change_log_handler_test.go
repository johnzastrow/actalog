package handler

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
)

func createTestDataChangeLogService() *service.DataChangeLogService {
	mockDataChangeLogRepo := NewMockDataChangeLogRepository()
	return service.NewDataChangeLogService(mockDataChangeLogRepo)
}

func createTestDataChangeLogServiceWithData() (*service.DataChangeLogService, *MockDataChangeLogRepository) {
	mockRepo := NewMockDataChangeLogRepository()
	now := time.Now()
	afterVal1 := `{"email":"test@example.com"}`
	mockRepo.AddLog(&domain.DataChangeLog{
		ID:          1,
		EntityType:  "user",
		EntityID:    1,
		EntityName:  "test@example.com",
		Operation:   "update",
		UserID:      1,
		UserEmail:   "admin@example.com",
		AfterValues: &afterVal1,
		CreatedAt:   now,
	})
	beforeVal2 := `{"name":"Old Name"}`
	afterVal2 := `{"name":"New Name"}`
	mockRepo.AddLog(&domain.DataChangeLog{
		ID:           2,
		EntityType:   "workout",
		EntityID:     1,
		EntityName:   "Test Workout",
		Operation:    "update",
		UserID:       1,
		UserEmail:    "admin@example.com",
		BeforeValues: &beforeVal2,
		AfterValues:  &afterVal2,
		CreatedAt:    now,
	})
	return service.NewDataChangeLogService(mockRepo), mockRepo
}

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

// Success path tests

func TestDataChangeLogHandler_GetDataChangeLog_Success(t *testing.T) {
	svc, _ := createTestDataChangeLogServiceWithData()
	handler := NewDataChangeLogHandler(svc, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs/1", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetDataChangeLog(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")

	body := rr.Body.String()
	if !strings.Contains(body, "user") {
		t.Error("Response should contain entity_type 'user'")
	}
	if !strings.Contains(body, "create") {
		t.Error("Response should contain operation 'create'")
	}
}

func TestDataChangeLogHandler_GetDataChangeLog_NotFound(t *testing.T) {
	svc, _ := createTestDataChangeLogServiceWithData()
	handler := NewDataChangeLogHandler(svc, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs/999", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.GetDataChangeLog(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "not found")
}

func TestDataChangeLogHandler_ListDataChangeLogs_Success(t *testing.T) {
	svc, _ := createTestDataChangeLogServiceWithData()
	handler := NewDataChangeLogHandler(svc, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListDataChangeLogs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")

	body := rr.Body.String()
	if !strings.Contains(body, "logs") {
		t.Error("Response should contain 'logs' field")
	}
	if !strings.Contains(body, "total") {
		t.Error("Response should contain 'total' field")
	}
}

func TestDataChangeLogHandler_ListDataChangeLogs_WithFilters(t *testing.T) {
	svc, _ := createTestDataChangeLogServiceWithData()
	handler := NewDataChangeLogHandler(svc, createTestLogger())

	tests := []struct {
		name  string
		query string
	}{
		{"entity_type filter", "?entity_type=user"},
		{"entity_id filter", "?entity_id=1"},
		{"operation filter", "?operation=create"},
		{"user_id filter", "?user_id=1"},
		{"user_email filter", "?user_email=test@example.com"},
		{"pagination", "?limit=10&offset=0"},
		{"combined filters", "?entity_type=user&operation=create&limit=5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs"+tt.query, "", 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			handler.ListDataChangeLogs(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
		})
	}
}

func TestDataChangeLogHandler_ListDataChangeLogs_ValidDateFormats(t *testing.T) {
	svc, _ := createTestDataChangeLogServiceWithData()
	handler := NewDataChangeLogHandler(svc, createTestLogger())

	tests := []struct {
		name  string
		query string
	}{
		{"RFC3339 start_date", "?start_date=2025-01-01T00:00:00Z"},
		{"simple start_date", "?start_date=2025-01-01"},
		{"RFC3339 end_date", "?end_date=2025-12-31T23:59:59Z"},
		{"simple end_date", "?end_date=2025-12-31"},
		{"date range", "?start_date=2025-01-01&end_date=2025-12-31"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs"+tt.query, "", 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			handler.ListDataChangeLogs(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
		})
	}
}

func TestDataChangeLogHandler_GetEntityHistory_Success(t *testing.T) {
	svc, _ := createTestDataChangeLogServiceWithData()
	handler := NewDataChangeLogHandler(svc, createTestLogger())

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs/entity/user/1", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "entity_type", "user")
	req = addChiURLParam(req, "entity_id", "1")
	rr := httptest.NewRecorder()

	handler.GetEntityHistory(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")

	body := rr.Body.String()
	if !strings.Contains(body, "entity_type") {
		t.Error("Response should contain 'entity_type' field")
	}
	if !strings.Contains(body, "entity_id") {
		t.Error("Response should contain 'entity_id' field")
	}
	if !strings.Contains(body, "logs") {
		t.Error("Response should contain 'logs' field")
	}
}

func TestDataChangeLogHandler_GetEntityHistory_WithPagination(t *testing.T) {
	svc, _ := createTestDataChangeLogServiceWithData()
	handler := NewDataChangeLogHandler(svc, createTestLogger())

	tests := []struct {
		name  string
		query string
	}{
		{"with limit", "?limit=10"},
		{"with offset", "?offset=0"},
		{"with pagination", "?limit=10&offset=0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs/entity/user/1"+tt.query, "", 1, "admin@example.com", "admin")
			req = addChiURLParam(req, "entity_type", "user")
			req = addChiURLParam(req, "entity_id", "1")
			rr := httptest.NewRecorder()

			handler.GetEntityHistory(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
		})
	}
}

func TestDataChangeLogHandler_GetEntityHistory_DifferentEntityTypes(t *testing.T) {
	svc, _ := createTestDataChangeLogServiceWithData()
	handler := NewDataChangeLogHandler(svc, createTestLogger())

	entityTypes := []string{"user", "workout", "wod", "movement"}

	for _, entityType := range entityTypes {
		t.Run(entityType, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs/entity/"+entityType+"/1", "", 1, "admin@example.com", "admin")
			req = addChiURLParam(req, "entity_type", entityType)
			req = addChiURLParam(req, "entity_id", "1")
			rr := httptest.NewRecorder()

			handler.GetEntityHistory(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
		})
	}
}

func TestDataChangeLogHandler_CleanupOldLogs_Success(t *testing.T) {
	svc, _ := createTestDataChangeLogServiceWithData()
	handler := NewDataChangeLogHandler(svc, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/data-change-logs/cleanup", `{"retention_days": 30}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CleanupOldLogs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")

	body := rr.Body.String()
	if !strings.Contains(body, "deleted_count") {
		t.Error("Response should contain 'deleted_count' field")
	}
	if !strings.Contains(body, "message") {
		t.Error("Response should contain 'message' field")
	}
}

func TestDataChangeLogHandler_CleanupOldLogs_DifferentRetentionDays(t *testing.T) {
	svc, _ := createTestDataChangeLogServiceWithData()
	handler := NewDataChangeLogHandler(svc, createTestLogger())

	tests := []struct {
		days int
		name string
	}{
		{7, "7_days"},
		{30, "30_days"},
		{60, "60_days"},
		{90, "90_days"},
		{365, "365_days"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"retention_days": ` + strconv.Itoa(tt.days) + `}`
			req := createAuthenticatedRequest(http.MethodPost, "/api/admin/data-change-logs/cleanup", body, 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			handler.CleanupOldLogs(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
		})
	}
}
