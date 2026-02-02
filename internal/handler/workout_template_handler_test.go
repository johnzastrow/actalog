package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johnzastrow/actalog/internal/service"
)

func TestWorkoutTemplateHandler_CreateTemplate_Unauthorized(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createTestRequest(http.MethodPost, "/api/templates", `{"name": "Test Template"}`)
	rr := httptest.NewRecorder()

	handler.CreateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWorkoutTemplateHandler_CreateTemplate_InvalidJSON(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/templates", "{bad json", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.CreateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestWorkoutTemplateHandler_CreateTemplate_MissingName(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/templates", `{"description": "Test"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.CreateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Template name is required")
}

func TestWorkoutTemplateHandler_GetTemplate_InvalidID(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	// chi.URLParam returns empty string without router context
	req := createTestRequest(http.MethodGet, "/api/templates/abc", "")
	rr := httptest.NewRecorder()

	handler.GetTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid template ID")
}

func TestWorkoutTemplateHandler_ListMyTemplates_Unauthorized(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createTestRequest(http.MethodGet, "/api/workouts/my-templates", "")
	rr := httptest.NewRecorder()

	handler.ListMyTemplates(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWorkoutTemplateHandler_UpdateTemplate_Unauthorized(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createTestRequest(http.MethodPut, "/api/templates/1", `{"name": "Updated"}`)
	rr := httptest.NewRecorder()

	handler.UpdateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWorkoutTemplateHandler_UpdateTemplate_InvalidID(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodPut, "/api/templates/abc", `{"name": "Updated"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid template ID")
}

func TestWorkoutTemplateHandler_UpdateTemplate_InvalidJSON(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/templates/1", "{bad json", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.UpdateTemplate(rr, req)

	// Without chi router context, URL param is empty -> Invalid template ID
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestWorkoutTemplateHandler_DeleteTemplate_Unauthorized(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createTestRequest(http.MethodDelete, "/api/templates/1", "")
	rr := httptest.NewRecorder()

	handler.DeleteTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestWorkoutTemplateHandler_DeleteTemplate_InvalidID(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodDelete, "/api/templates/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.DeleteTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid template ID")
}

// Test NewWorkoutTemplateHandler

func TestNewWorkoutTemplateHandler(t *testing.T) {
	handler := NewWorkoutTemplateHandler(nil)
	if handler == nil {
		t.Error("NewWorkoutTemplateHandler should return a non-nil handler")
	}
}

// Test UpdateTemplate with valid ID and JSON

func TestWorkoutTemplateHandler_UpdateTemplate_ValidIDInvalidJSON(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/templates/1", "{bad json", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestWorkoutTemplateHandler_UpdateTemplate_MissingName(t *testing.T) {
	handler := &WorkoutTemplateHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/templates/1", `{"description": "Test"}`, 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Template name is required")
}

// Removed 9 panic-expectation tests:
// - TestWorkoutTemplateHandler_ListStandardTemplates_NilService
// - TestWorkoutTemplateHandler_ListMyTemplates_NilService
// - TestWorkoutTemplateHandler_GetTemplate_ValidIDNilService
// - TestWorkoutTemplateHandler_DeleteTemplate_ValidIDNilService
// - TestWorkoutTemplateHandler_CreateTemplate_ValidInputNilService
// - TestWorkoutTemplateHandler_UpdateTemplate_ValidInputNilService
// - TestWorkoutTemplateHandler_ListStandardTemplates_WithPagination (4 subtests)
// - TestWorkoutTemplateHandler_ListMyTemplates_WithPagination (3 subtests)
// - TestWorkoutTemplateHandler_GetTemplate_DifferentIDs (3 subtests)
// These tests verified nil pointer panics, not business logic.

// ===== Tests with real service for WorkoutTemplateHandler =====

func createWorkoutTemplateTestService() *service.WorkoutTemplateService {
	mockWorkoutRepo := NewMockWorkoutRepository()
	mockAuditLogRepo := NewMockAuditLogRepository()
	return service.NewWorkoutTemplateService(mockWorkoutRepo, nil, nil, mockAuditLogRepo)
}

func TestWorkoutTemplateHandler_ListStandardTemplates_Success(t *testing.T) {
	svc := createWorkoutTemplateTestService()
	handler := NewWorkoutTemplateHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/templates", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ListStandardTemplates(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
	assertBodyContains(t, rr, "workouts")
}

func TestWorkoutTemplateHandler_ListStandardTemplates_WithPagination(t *testing.T) {
	svc := createWorkoutTemplateTestService()
	handler := NewWorkoutTemplateHandler(svc)

	tests := []struct {
		name string
		url  string
	}{
		{"with limit", "/api/templates?limit=10"},
		{"with offset", "/api/templates?offset=5"},
		{"with both", "/api/templates?limit=10&offset=5"},
		{"with invalid limit", "/api/templates?limit=abc"},
		{"with invalid offset", "/api/templates?offset=abc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tc.url, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.ListStandardTemplates(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
			assertBodyContains(t, rr, "workouts")
		})
	}
}

func TestWorkoutTemplateHandler_ListStandardTemplates_ServiceError(t *testing.T) {
	mockWorkoutRepo := NewMockWorkoutRepository()
	mockWorkoutRepo.SetError(ErrMockInternalError)
	mockAuditLogRepo := NewMockAuditLogRepository()
	svc := service.NewWorkoutTemplateService(mockWorkoutRepo, nil, nil, mockAuditLogRepo)
	handler := NewWorkoutTemplateHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/templates", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ListStandardTemplates(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestWorkoutTemplateHandler_ListMyTemplates_Success(t *testing.T) {
	svc := createWorkoutTemplateTestService()
	handler := NewWorkoutTemplateHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/workouts/my-templates", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ListMyTemplates(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
	assertBodyContains(t, rr, "workouts")
}

func TestWorkoutTemplateHandler_ListMyTemplates_WithPagination(t *testing.T) {
	svc := createWorkoutTemplateTestService()
	handler := NewWorkoutTemplateHandler(svc)

	tests := []struct {
		name string
		url  string
	}{
		{"with limit", "/api/workouts/my-templates?limit=10"},
		{"with offset", "/api/workouts/my-templates?offset=5"},
		{"with both", "/api/workouts/my-templates?limit=10&offset=5"},
		{"with invalid limit", "/api/workouts/my-templates?limit=abc"},
		{"with invalid offset", "/api/workouts/my-templates?offset=abc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tc.url, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.ListMyTemplates(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
			assertBodyContains(t, rr, "workouts")
		})
	}
}

func TestWorkoutTemplateHandler_ListMyTemplates_ServiceError(t *testing.T) {
	mockWorkoutRepo := NewMockWorkoutRepository()
	mockWorkoutRepo.SetError(ErrMockInternalError)
	mockAuditLogRepo := NewMockAuditLogRepository()
	svc := service.NewWorkoutTemplateService(mockWorkoutRepo, nil, nil, mockAuditLogRepo)
	handler := NewWorkoutTemplateHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/workouts/my-templates", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ListMyTemplates(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestWorkoutTemplateHandler_GetTemplate_Success(t *testing.T) {
	svc := createWorkoutTemplateTestService()
	handler := NewWorkoutTemplateHandler(svc)

	// ID 3 exists in the mock - "My Custom Workout"
	req := createAuthenticatedRequest(http.MethodGet, "/api/templates/3", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "3")
	rr := httptest.NewRecorder()

	handler.GetTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
}

func TestWorkoutTemplateHandler_GetTemplate_NotFound(t *testing.T) {
	svc := createWorkoutTemplateTestService()
	handler := NewWorkoutTemplateHandler(svc)

	// ID 999 doesn't exist
	req := createAuthenticatedRequest(http.MethodGet, "/api/templates/999", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.GetTemplate(rr, req)

	// Returns 404 with error message when not found
	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "not found")
}

func TestWorkoutTemplateHandler_CreateTemplate_Success(t *testing.T) {
	svc := createWorkoutTemplateTestService()
	handler := NewWorkoutTemplateHandler(svc)

	body := `{"name": "New Template", "intro_warmup": "Warmup instructions", "notes": "Test notes"}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/templates", body, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.CreateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusCreated)
	assertContentType(t, rr, "application/json")
}

func TestWorkoutTemplateHandler_CreateTemplate_ServiceError(t *testing.T) {
	mockWorkoutRepo := NewMockWorkoutRepository()
	mockWorkoutRepo.SetError(ErrMockInternalError)
	mockAuditLogRepo := NewMockAuditLogRepository()
	svc := service.NewWorkoutTemplateService(mockWorkoutRepo, nil, nil, mockAuditLogRepo)
	handler := NewWorkoutTemplateHandler(svc)

	body := `{"name": "New Template", "intro_warmup": "Warmup instructions", "notes": "Test notes"}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/templates", body, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.CreateTemplate(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}
