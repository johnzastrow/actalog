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

// Removed 25 panic-expectation tests:
// - TestWODHandler_GetWOD_ValidIDNilService, TestWODHandler_GetWOD_DifferentIDs
// - TestWODHandler_CreateWOD_ValidInputNilService, TestWODHandler_CreateWOD_MinimalValidInput
// - TestWODHandler_CreateWOD_WithAllFields, TestWODHandler_CreateWOD_WithSource
// - TestWODHandler_UpdateWOD_ValidInputNilService, TestWODHandler_UpdateWOD_AsAdmin
// - TestWODHandler_UpdateWOD_MissingName, TestWODHandler_UpdateWOD_DifferentIDs
// - TestWODHandler_DeleteWOD_ValidIDNilService, TestWODHandler_DeleteWOD_AsAdmin
// - TestWODHandler_DeleteWOD_DifferentIDs
// - TestWODHandler_ListWODs_NilService, TestWODHandler_ListWODs_WithStandardOnlyParam
// - TestWODHandler_ListWODs_WithPaginationParams, TestWODHandler_ListWODs_WithUserAuth
// - TestWODHandler_ListStandardWODs_WithPagination, TestWODHandler_ListStandardWODs_DifferentLimits
// - TestWODHandler_ListMyWODs_WithPagination, TestWODHandler_ListMyWODs_DifferentUserIDs
// - TestWODHandler_SearchWODs_ValidQuery, TestWODHandler_SearchWODs_WithLimitParam
// - TestWODHandler_SearchWODs_VariousQueries, TestWODHandler_SearchWODs_WithAuth
// These tests verified nil pointer panics, not business logic.

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
