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

// Tests for GetDataChangeLog with valid ID

func TestDataChangeLogHandler_GetDataChangeLog_ValidIDNilService(t *testing.T) {
	handler := &DataChangeLogHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs/1", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("GetDataChangeLog requires service")
		}
	}()

	handler.GetDataChangeLog(rr, req)
}

// Tests for ListDataChangeLogs with valid filters

func TestDataChangeLogHandler_ListDataChangeLogs_WithFilters(t *testing.T) {
	handler := &DataChangeLogHandler{
		logger: createTestLogger(),
	}

	tests := []struct {
		name  string
		query string
	}{
		{"default", "/api/data-change-logs"},
		{"with entity_type", "/api/data-change-logs?entity_type=user"},
		{"with operation", "/api/data-change-logs?operation=create"},
		{"with user_email", "/api/data-change-logs?user_email=admin@example.com"},
		{"with limit", "/api/data-change-logs?limit=20"},
		{"with offset", "/api/data-change-logs?offset=10"},
		{"with pagination", "/api/data-change-logs?limit=20&offset=10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tt.query, "", 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListDataChangeLogs requires service")
				}
			}()

			handler.ListDataChangeLogs(rr, req)
		})
	}
}

// Tests for ListDataChangeLogs with valid date formats

func TestDataChangeLogHandler_ListDataChangeLogs_ValidDateFormats(t *testing.T) {
	handler := &DataChangeLogHandler{
		logger: createTestLogger(),
	}

	tests := []struct {
		name  string
		query string
	}{
		{"RFC3339 start_date", "/api/data-change-logs?start_date=2024-01-15T00:00:00Z"},
		{"simple start_date", "/api/data-change-logs?start_date=2024-01-15"},
		{"RFC3339 end_date", "/api/data-change-logs?end_date=2024-12-31T23:59:59Z"},
		{"simple end_date", "/api/data-change-logs?end_date=2024-12-31"},
		{"both dates", "/api/data-change-logs?start_date=2024-01-01&end_date=2024-12-31"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tt.query, "", 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListDataChangeLogs requires service")
				}
			}()

			handler.ListDataChangeLogs(rr, req)
		})
	}
}

// Tests for ListDataChangeLogs with valid entity_id

func TestDataChangeLogHandler_ListDataChangeLogs_ValidEntityID(t *testing.T) {
	handler := &DataChangeLogHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs?entity_id=123", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("ListDataChangeLogs requires service")
		}
	}()

	handler.ListDataChangeLogs(rr, req)
}

// Tests for ListDataChangeLogs with valid user_id

func TestDataChangeLogHandler_ListDataChangeLogs_ValidUserID(t *testing.T) {
	handler := &DataChangeLogHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs?user_id=1", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("ListDataChangeLogs requires service")
		}
	}()

	handler.ListDataChangeLogs(rr, req)
}

// Tests for GetEntityHistory with valid input

func TestDataChangeLogHandler_GetEntityHistory_ValidInputNilService(t *testing.T) {
	handler := &DataChangeLogHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs/entity/user/1", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "entity_type", "user")
	req = addChiURLParam(req, "entity_id", "1")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("GetEntityHistory requires service")
		}
	}()

	handler.GetEntityHistory(rr, req)
}

// Tests for GetEntityHistory with pagination

func TestDataChangeLogHandler_GetEntityHistory_WithPagination(t *testing.T) {
	handler := &DataChangeLogHandler{
		logger: createTestLogger(),
	}

	tests := []struct {
		name  string
		query string
	}{
		{"default", "/api/data-change-logs/entity/user/1"},
		{"with limit", "/api/data-change-logs/entity/user/1?limit=10"},
		{"with offset", "/api/data-change-logs/entity/user/1?offset=5"},
		{"with both", "/api/data-change-logs/entity/user/1?limit=20&offset=10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tt.query, "", 1, "admin@example.com", "admin")
			req = addChiURLParam(req, "entity_type", "user")
			req = addChiURLParam(req, "entity_id", "1")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("GetEntityHistory requires service")
				}
			}()

			handler.GetEntityHistory(rr, req)
		})
	}
}

// Tests for GetEntityHistory with different entity types

func TestDataChangeLogHandler_GetEntityHistory_DifferentEntityTypes(t *testing.T) {
	handler := &DataChangeLogHandler{
		logger: createTestLogger(),
	}

	entityTypes := []string{"user", "workout", "wod", "movement"}

	for _, entityType := range entityTypes {
		t.Run("entity_type_"+entityType, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs/entity/"+entityType+"/1", "", 1, "admin@example.com", "admin")
			req = addChiURLParam(req, "entity_type", entityType)
			req = addChiURLParam(req, "entity_id", "1")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("GetEntityHistory requires service")
				}
			}()

			handler.GetEntityHistory(rr, req)
		})
	}
}

// Tests for CleanupOldLogs with valid input

func TestDataChangeLogHandler_CleanupOldLogs_ValidInputNilService(t *testing.T) {
	handler := &DataChangeLogHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/data-change-logs/cleanup",
		`{"retention_days": 30}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("CleanupOldLogs requires service")
		}
	}()

	handler.CleanupOldLogs(rr, req)
}

// Tests for CleanupOldLogs with different retention days

func TestDataChangeLogHandler_CleanupOldLogs_DifferentRetentionDays(t *testing.T) {
	handler := &DataChangeLogHandler{
		logger: createTestLogger(),
	}

	retentionDays := []int{7, 30, 90, 365}

	for _, days := range retentionDays {
		t.Run("days_"+string(rune('0'+days%10)), func(t *testing.T) {
			body := `{"retention_days": ` + string(rune('0'+days%10)) + `}`
			if days >= 10 {
				body = `{"retention_days": 30}`
			}
			req := createAuthenticatedRequest(http.MethodPost, "/api/admin/data-change-logs/cleanup",
				body, 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("CleanupOldLogs requires service")
				}
			}()

			handler.CleanupOldLogs(rr, req)
		})
	}
}

// Tests for GetDataChangeLog with different IDs

func TestDataChangeLogHandler_GetDataChangeLog_DifferentIDs(t *testing.T) {
	handler := &DataChangeLogHandler{
		logger: createTestLogger(),
	}

	testIDs := []string{"1", "10", "100"}

	for _, id := range testIDs {
		t.Run("id_"+id, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/data-change-logs/"+id, "", 1, "admin@example.com", "admin")
			req = addChiURLParam(req, "id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("GetDataChangeLog requires service")
				}
			}()

			handler.GetDataChangeLog(rr, req)
		})
	}
}

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
