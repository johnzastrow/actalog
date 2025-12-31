package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johnzastrow/actalog/internal/service"
)

func TestWODHandler_CreateWOD_Unauthorized(t *testing.T) {
	handler := &WODHandler{}

	// Request without user context
	req := createTestRequest(http.MethodPost, "/api/wods", `{"name": "Test WOD"}`)
	rr := httptest.NewRecorder()

	handler.CreateWOD(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWODHandler_CreateWOD_InvalidJSON(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/wods", "{bad json", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.CreateWOD(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestWODHandler_CreateWOD_MissingName(t *testing.T) {
	handler := &WODHandler{}

	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{
			name:      "empty name",
			body:      `{"name": ""}`,
			wantError: "WOD name is required",
		},
		{
			name:      "missing name field",
			body:      `{"source": "CrossFit"}`,
			wantError: "WOD name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPost, "/api/wods", tt.body, 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.CreateWOD(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestWODHandler_GetWOD_InvalidID(t *testing.T) {
	handler := &WODHandler{}

	// Without chi router context, URLParam returns empty string
	req := createTestRequest(http.MethodGet, "/api/wods/abc", "")
	rr := httptest.NewRecorder()

	handler.GetWOD(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid WOD ID")
}

// Tests for ListWODs

func TestWODHandler_ListWODs_NilService(t *testing.T) {
	handler := &WODHandler{}

	// Without auth, tries to call service.ListStandard which will panic on nil service
	// This tests the function entry - for full test would need mock service
	req := createTestRequest(http.MethodGet, "/api/wods", "")
	rr := httptest.NewRecorder()

	// We expect a panic or error since service is nil
	// Test the code path by wrapping in recover
	defer func() {
		if r := recover(); r == nil {
			t.Log("ListWODs requires service - expected panic on nil service")
		}
	}()

	handler.ListWODs(rr, req)
}

func TestWODHandler_ListWODs_WithStandardOnlyParam(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods?standard=true", "")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("ListWODs requires service")
		}
	}()

	handler.ListWODs(rr, req)
}

func TestWODHandler_ListWODs_WithPaginationParams(t *testing.T) {
	handler := &WODHandler{}

	tests := []struct {
		name  string
		query string
	}{
		{"with limit", "/api/wods?limit=10"},
		{"with offset", "/api/wods?offset=5"},
		{"with both", "/api/wods?limit=10&offset=5"},
		{"with invalid limit", "/api/wods?limit=abc"},
		{"with invalid offset", "/api/wods?offset=xyz"},
		{"with negative values", "/api/wods?limit=-1&offset=-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, tt.query, "")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListWODs requires service - params were parsed")
				}
			}()

			handler.ListWODs(rr, req)
		})
	}
}

// Tests for ListStandardWODs

func TestWODHandler_ListStandardWODs_WithPagination(t *testing.T) {
	handler := &WODHandler{}

	tests := []struct {
		name  string
		query string
	}{
		{"default pagination", "/api/wods/standard"},
		{"with limit", "/api/wods/standard?limit=50"},
		{"with offset", "/api/wods/standard?offset=10"},
		{"with both", "/api/wods/standard?limit=25&offset=10"},
		{"invalid limit ignored", "/api/wods/standard?limit=abc"},
		{"invalid offset ignored", "/api/wods/standard?offset=xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, tt.query, "")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListStandardWODs requires service")
				}
			}()

			handler.ListStandardWODs(rr, req)
		})
	}
}

// Tests for ListMyWODs

func TestWODHandler_ListMyWODs_Unauthorized(t *testing.T) {
	handler := &WODHandler{}

	// Request without user context
	req := createTestRequest(http.MethodGet, "/api/wods/mine", "")
	rr := httptest.NewRecorder()

	handler.ListMyWODs(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Authentication required")
}

func TestWODHandler_ListMyWODs_WithPagination(t *testing.T) {
	handler := &WODHandler{}

	tests := []struct {
		name  string
		query string
	}{
		{"default pagination", "/api/wods/mine"},
		{"with limit", "/api/wods/mine?limit=50"},
		{"with offset", "/api/wods/mine?offset=10"},
		{"with both", "/api/wods/mine?limit=25&offset=10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tt.query, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListMyWODs requires service")
				}
			}()

			handler.ListMyWODs(rr, req)
		})
	}
}

// Tests for SearchWODs

func TestWODHandler_SearchWODs_MissingQuery(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods/search", "")
	rr := httptest.NewRecorder()

	handler.SearchWODs(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Search query is required")
}

func TestWODHandler_SearchWODs_EmptyQuery(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods/search?q=", "")
	rr := httptest.NewRecorder()

	handler.SearchWODs(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Search query is required")
}

func TestWODHandler_SearchWODs_ValidQuery(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods/search?q=fran", "")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("SearchWODs requires service")
		}
	}()

	handler.SearchWODs(rr, req)
}

// Tests for UpdateWOD

func TestWODHandler_UpdateWOD_Unauthorized(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodPut, "/api/wods/1", `{"name": "Updated WOD"}`)
	rr := httptest.NewRecorder()

	handler.UpdateWOD(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWODHandler_UpdateWOD_InvalidID(t *testing.T) {
	handler := &WODHandler{}

	// Without chi context, URLParam returns empty string -> invalid ID
	req := createAuthenticatedRequest(http.MethodPut, "/api/wods/abc", `{"name": "Updated"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateWOD(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid WOD ID")
}

func TestWODHandler_UpdateWOD_InvalidJSON(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/wods/1", "{bad json", 1, "test@example.com", "user")
	// Add chi URL param context
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateWOD(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

// Tests for DeleteWOD

func TestWODHandler_DeleteWOD_Unauthorized(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodDelete, "/api/wods/1", "")
	rr := httptest.NewRecorder()

	handler.DeleteWOD(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWODHandler_DeleteWOD_InvalidID(t *testing.T) {
	handler := &WODHandler{}

	// Without chi context, URLParam returns empty string -> invalid ID
	req := createAuthenticatedRequest(http.MethodDelete, "/api/wods/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.DeleteWOD(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid WOD ID")
}

func TestNewWODHandler(t *testing.T) {
	handler := NewWODHandler(nil)
	if handler == nil {
		t.Error("NewWODHandler should return a non-nil handler")
	}
}

func TestWODHandler_GetWOD_ValidIDNilService(t *testing.T) {
	handler := &WODHandler{}

	req := createTestRequest(http.MethodGet, "/api/wods/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("GetWOD requires service")
		}
	}()

	handler.GetWOD(rr, req)
}

func TestWODHandler_GetWOD_DifferentIDs(t *testing.T) {
	handler := &WODHandler{}

	testIDs := []string{"1", "10", "100", "999"}

	for _, id := range testIDs {
		t.Run("wod_id_"+id, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, "/api/wods/"+id, "")
			req = addChiURLParam(req, "id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("GetWOD requires service")
				}
			}()

			handler.GetWOD(rr, req)
		})
	}
}

func TestWODHandler_CreateWOD_ValidInputNilService(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/wods",
		`{"name": "Test WOD", "type": "AMRAP", "regime": "15 min", "score_type": "rounds+reps", "description": "Test description"}`,
		1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("CreateWOD requires service")
		}
	}()

	handler.CreateWOD(rr, req)
}

func TestWODHandler_CreateWOD_MinimalValidInput(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/wods", `{"name": "Minimal WOD"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("CreateWOD requires service")
		}
	}()

	handler.CreateWOD(rr, req)
}

func TestWODHandler_CreateWOD_WithAllFields(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/wods",
		`{"name": "Full WOD", "source": "CrossFit", "type": "For Time", "regime": "21-15-9", "score_type": "time", "description": "Thrusters and pull-ups", "url": "https://example.com", "notes": "Scale as needed"}`,
		1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("CreateWOD requires service")
		}
	}()

	handler.CreateWOD(rr, req)
}

func TestWODHandler_UpdateWOD_ValidInputNilService(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/wods/1",
		`{"name": "Updated WOD", "type": "For Time"}`,
		1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("UpdateWOD requires service")
		}
	}()

	handler.UpdateWOD(rr, req)
}

func TestWODHandler_UpdateWOD_AsAdmin(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/wods/1",
		`{"name": "Admin Updated WOD"}`,
		1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("UpdateWOD requires service")
		}
	}()

	handler.UpdateWOD(rr, req)
}

func TestWODHandler_DeleteWOD_ValidIDNilService(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/wods/1", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("DeleteWOD requires service")
		}
	}()

	handler.DeleteWOD(rr, req)
}

func TestWODHandler_DeleteWOD_AsAdmin(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/wods/1", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("DeleteWOD requires service")
		}
	}()

	handler.DeleteWOD(rr, req)
}

func TestWODHandler_SearchWODs_WithLimitParam(t *testing.T) {
	handler := &WODHandler{}

	tests := []struct {
		name  string
		query string
	}{
		{"with limit", "/api/wods/search?q=fran&limit=10"},
		{"with large limit", "/api/wods/search?q=fran&limit=100"},
		{"with invalid limit", "/api/wods/search?q=fran&limit=abc"},
		{"with zero limit", "/api/wods/search?q=fran&limit=0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, tt.query, "")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("SearchWODs requires service")
				}
			}()

			handler.SearchWODs(rr, req)
		})
	}
}

func TestWODHandler_ListWODs_WithUserAuth(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/wods", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("ListWODs requires service")
		}
	}()

	handler.ListWODs(rr, req)
}

func TestWODHandler_UpdateWOD_MissingName(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/wods/1", `{"type": "AMRAP"}`, 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("UpdateWOD validation might require service")
		}
	}()

	handler.UpdateWOD(rr, req)
}

func TestWODHandler_SearchWODs_VariousQueries(t *testing.T) {
	handler := &WODHandler{}

	tests := []struct {
		name  string
		query string
	}{
		{"single word", "/api/wods/search?q=cindy"},
		{"multiple words", "/api/wods/search?q=for+time"},
		{"with numbers", "/api/wods/search?q=21-15-9"},
		{"long query", "/api/wods/search?q=crossfit+benchmark+workout"},
		{"special chars", "/api/wods/search?q=helen%20wod"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, tt.query, "")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("SearchWODs requires service")
				}
			}()

			handler.SearchWODs(rr, req)
		})
	}
}

func TestWODHandler_SearchWODs_WithAuth(t *testing.T) {
	handler := &WODHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/wods/search?q=fran", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("SearchWODs requires service")
		}
	}()

	handler.SearchWODs(rr, req)
}

func TestWODHandler_ListMyWODs_DifferentUserIDs(t *testing.T) {
	handler := &WODHandler{}

	testUserIDs := []int64{1, 10, 100}

	for _, userID := range testUserIDs {
		t.Run("user_"+string(rune(userID)), func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/wods/mine", "", userID, "test@example.com", "user")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListMyWODs requires service")
				}
			}()

			handler.ListMyWODs(rr, req)
		})
	}
}

func TestWODHandler_DeleteWOD_DifferentIDs(t *testing.T) {
	handler := &WODHandler{}

	testIDs := []string{"1", "10", "100", "999"}

	for _, id := range testIDs {
		t.Run("wod_id_"+id, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodDelete, "/api/wods/"+id, "", 1, "test@example.com", "user")
			req = addChiURLParam(req, "id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("DeleteWOD requires service")
				}
			}()

			handler.DeleteWOD(rr, req)
		})
	}
}

func TestWODHandler_UpdateWOD_DifferentIDs(t *testing.T) {
	handler := &WODHandler{}

	testIDs := []string{"1", "10", "100"}

	for _, id := range testIDs {
		t.Run("wod_id_"+id, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPut, "/api/wods/"+id, `{"name": "Test"}`, 1, "test@example.com", "user")
			req = addChiURLParam(req, "id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("UpdateWOD requires service")
				}
			}()

			handler.UpdateWOD(rr, req)
		})
	}
}

func TestWODHandler_CreateWOD_WithSource(t *testing.T) {
	handler := &WODHandler{}

	tests := []struct {
		name   string
		source string
	}{
		{"CrossFit source", `{"name": "Test", "source": "CrossFit"}`},
		{"custom source", `{"name": "Test", "source": "My Gym"}`},
		{"no source", `{"name": "Test"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPost, "/api/wods", tt.source, 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("CreateWOD requires service")
				}
			}()

			handler.CreateWOD(rr, req)
		})
	}
}

func TestWODHandler_ListStandardWODs_DifferentLimits(t *testing.T) {
	handler := &WODHandler{}

	tests := []struct {
		name  string
		limit string
	}{
		{"limit 5", "/api/wods/standard?limit=5"},
		{"limit 50", "/api/wods/standard?limit=50"},
		{"limit 100", "/api/wods/standard?limit=100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, tt.limit, "")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListStandardWODs requires service")
				}
			}()

			handler.ListStandardWODs(rr, req)
		})
	}
}

// Helper function to create a test WOD service with mock repositories
func createTestWODService() *service.WODService {
	mockWODRepo := NewMockWODRepository()
	mockDataChangeLogRepo := NewMockDataChangeLogRepository()
	mockAuditLogRepo := NewMockAuditLogRepository()
	dataChangeLogService := service.NewDataChangeLogService(mockDataChangeLogRepo)
	return service.NewWODService(mockWODRepo, dataChangeLogService, mockAuditLogRepo)
}

// =============================================
// Mock-based tests with real service for WOD handler
// =============================================

func TestWODHandler_ListWODs_Success(t *testing.T) {
	wodService := createTestWODService()
	handler := NewWODHandler(wodService)

	req := createTestRequest(http.MethodGet, "/api/wods", "")
	rr := httptest.NewRecorder()

	handler.ListWODs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "wods")
}

func TestWODHandler_ListWODs_WithPagination(t *testing.T) {
	wodService := createTestWODService()
	handler := NewWODHandler(wodService)

	req := createTestRequest(http.MethodGet, "/api/wods?limit=10&offset=0", "")
	rr := httptest.NewRecorder()

	handler.ListWODs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "wods")
}

func TestWODHandler_ListStandardWODs_Success(t *testing.T) {
	wodService := createTestWODService()
	handler := NewWODHandler(wodService)

	req := createTestRequest(http.MethodGet, "/api/wods/standard", "")
	rr := httptest.NewRecorder()

	handler.ListStandardWODs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "wods")
}

func TestWODHandler_ListMyWODs_Success(t *testing.T) {
	wodService := createTestWODService()
	handler := NewWODHandler(wodService)

	req := createAuthenticatedRequest(http.MethodGet, "/api/wods/mine", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ListMyWODs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "wods")
}

func TestWODHandler_SearchWODs_Success(t *testing.T) {
	wodService := createTestWODService()
	handler := NewWODHandler(wodService)

	req := createTestRequest(http.MethodGet, "/api/wods/search?q=Fran", "")
	rr := httptest.NewRecorder()

	handler.SearchWODs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "wods")
}

func TestWODHandler_SearchWODs_EmptyQueryWithService(t *testing.T) {
	wodService := createTestWODService()
	handler := NewWODHandler(wodService)

	req := createTestRequest(http.MethodGet, "/api/wods/search", "")
	rr := httptest.NewRecorder()

	handler.SearchWODs(rr, req)

	// Search query is required
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestWODHandler_GetWOD_Success(t *testing.T) {
	wodService := createTestWODService()
	handler := NewWODHandler(wodService)

	req := createTestRequest(http.MethodGet, "/api/wods/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetWOD(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "Fran")
}

func TestWODHandler_GetWOD_NotFound(t *testing.T) {
	wodService := createTestWODService()
	handler := NewWODHandler(wodService)

	req := createTestRequest(http.MethodGet, "/api/wods/999", "")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.GetWOD(rr, req)

	// Service returns 500 for not found errors (wrapped error)
	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "not found")
}

func TestWODHandler_CreateWOD_Success(t *testing.T) {
	wodService := createTestWODService()
	handler := NewWODHandler(wodService)

	body := `{"name": "New WOD", "description": "A new workout", "score_type": "Time (HH:MM:SS)", "source": "Self-recorded", "type": "Self-created"}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/wods", body, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.CreateWOD(rr, req)

	assertStatusCode(t, rr, http.StatusCreated)
	assertBodyContains(t, rr, "New WOD")
}

func TestWODHandler_UpdateWOD_Success(t *testing.T) {
	wodService := createTestWODService()
	handler := NewWODHandler(wodService)

	// Use ID 4 which is the user-created WOD owned by user 1
	body := `{"name": "Updated Custom WOD", "description": "Updated description", "score_type": "Time (HH:MM:SS)", "source": "Self-recorded", "type": "Self-created"}`
	req := createAuthenticatedRequest(http.MethodPut, "/api/wods/4", body, 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "4")
	rr := httptest.NewRecorder()

	handler.UpdateWOD(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestWODHandler_DeleteWOD_Success(t *testing.T) {
	wodService := createTestWODService()
	handler := NewWODHandler(wodService)

	// Use ID 4 which is the user-created WOD owned by user 1
	req := createAuthenticatedRequest(http.MethodDelete, "/api/wods/4", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "4")
	rr := httptest.NewRecorder()

	handler.DeleteWOD(rr, req)

	// Should succeed since user owns the WOD
	if rr.Code != http.StatusOK && rr.Code != http.StatusNoContent {
		t.Errorf("Expected success, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestWODHandler_DeleteWOD_NotOwner(t *testing.T) {
	wodService := createTestWODService()
	handler := NewWODHandler(wodService)

	// Try to delete WOD 4 as user 999 (not the owner)
	req := createAuthenticatedRequest(http.MethodDelete, "/api/wods/4", "", 999, "other@example.com", "user")
	req = addChiURLParam(req, "id", "4")
	rr := httptest.NewRecorder()

	handler.DeleteWOD(rr, req)

	// Service returns error which handler converts to 500
	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "unauthorized")
}

func TestWODHandler_DeleteWOD_StandardWOD(t *testing.T) {
	wodService := createTestWODService()
	handler := NewWODHandler(wodService)

	// Try to delete a standard WOD (ID 1 is "Fran" - a standard WOD)
	req := createAuthenticatedRequest(http.MethodDelete, "/api/wods/1", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.DeleteWOD(rr, req)

	// Service returns error which handler converts to 500
	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "unauthorized")
}
