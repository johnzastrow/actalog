package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johnzastrow/actalog/internal/service"
)

func TestMovementHandler_Search_MissingQuery(t *testing.T) {
	handler := &MovementHandler{}

	req := createTestRequest(http.MethodGet, "/api/movements/search", "")
	rr := httptest.NewRecorder()

	handler.Search(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Query parameter 'q' is required")
}

func TestMovementHandler_Create_Unauthorized(t *testing.T) {
	handler := &MovementHandler{}

	// Request without user context
	req := createTestRequest(http.MethodPost, "/api/movements", `{"name": "Test", "type": "strength"}`)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestMovementHandler_Create_InvalidJSON(t *testing.T) {
	handler := &MovementHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/movements", "{bad json", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestMovementHandler_Create_MissingFields(t *testing.T) {
	handler := &MovementHandler{}

	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{
			name:      "missing name",
			body:      `{"type": "strength"}`,
			wantError: "Name and type are required",
		},
		{
			name:      "missing type",
			body:      `{"name": "Test Movement"}`,
			wantError: "Name and type are required",
		},
		{
			name:      "empty name",
			body:      `{"name": "", "type": "strength"}`,
			wantError: "Name and type are required",
		},
		{
			name:      "empty type",
			body:      `{"name": "Test", "type": ""}`,
			wantError: "Name and type are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPost, "/api/movements", tt.body, 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.Create(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestMovementHandler_Update_Unauthorized(t *testing.T) {
	handler := &MovementHandler{}

	// Request without user context
	req := createTestRequest(http.MethodPut, "/api/movements/1", `{"name": "Test", "type": "strength"}`)
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

// Note: Update tests that require chi.URLParam need router context
// Testing ID parsing requires chi router setup which is complex for unit tests
// These validation tests verify the input is checked before service calls

func TestMovementHandler_Update_InvalidID(t *testing.T) {
	handler := &MovementHandler{}

	// chi.URLParam returns empty string without router context, which fails int parsing
	req := createAuthenticatedRequest(http.MethodPut, "/api/movements/abc", `{"name": "Test", "type": "strength"}`, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	// Without chi router context, URLParam returns empty string -> invalid ID
	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid movement ID")
}

func TestMovementHandler_Delete_Unauthorized(t *testing.T) {
	handler := &MovementHandler{}

	// Request without user context
	req := createTestRequest(http.MethodDelete, "/api/movements/1", "")
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestMovementHandler_Delete_InvalidID(t *testing.T) {
	handler := &MovementHandler{}

	// Without chi router context, URLParam returns empty string -> invalid ID
	req := createAuthenticatedRequest(http.MethodDelete, "/api/movements/abc", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid movement ID")
}

// Tests for GetByID

func TestMovementHandler_GetByID_InvalidID(t *testing.T) {
	handler := &MovementHandler{}

	// Without chi router context, URLParam returns empty string -> invalid ID
	req := createTestRequest(http.MethodGet, "/api/movements/abc", "")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid movement ID")
}

// Tests for Update with valid ID

func TestMovementHandler_Update_InvalidJSON(t *testing.T) {
	handler := &MovementHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/movements/1", "{bad json", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestMovementHandler_Update_MissingFields(t *testing.T) {
	handler := &MovementHandler{}

	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{
			name:      "missing name",
			body:      `{"type": "strength"}`,
			wantError: "Name and type are required",
		},
		{
			name:      "missing type",
			body:      `{"name": "Test Movement"}`,
			wantError: "Name and type are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPut, "/api/movements/1", tt.body, 1, "test@example.com", "user")
			req = addChiURLParam(req, "id", "1")
			rr := httptest.NewRecorder()

			handler.Update(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

// Test NewMovementHandler

func TestNewMovementHandler(t *testing.T) {
	handler := NewMovementHandler(nil, nil, nil)
	if handler == nil {
		t.Error("NewMovementHandler should return a non-nil handler")
	}
}

// Removed 17 panic-expectation tests:
// - TestMovementHandler_ListAll_NilRepo
// - TestMovementHandler_ListStandard_NilRepo
// - TestMovementHandler_GetByID_ValidIDNilRepo
// - TestMovementHandler_Search_WithLimit (4 subtests)
// - TestMovementHandler_Create_ValidInputNilService
// - TestMovementHandler_Update_ValidInputNilService
// - TestMovementHandler_Update_AsAdminNilService
// - TestMovementHandler_Delete_ValidIDNilService
// - TestMovementHandler_Search_ValidQueryWithLogger (4 subtests)
// - TestMovementHandler_ListAll_WithLogger
// - TestMovementHandler_ListStandard_WithLogger
// - TestMovementHandler_GetByID_WithLogger
// - TestMovementHandler_GetByID_DifferentIDs (4 subtests)
// - TestMovementHandler_Delete_DifferentIDs (3 subtests)
// - TestMovementHandler_Update_DifferentIDs (3 subtests)
// - TestMovementHandler_Search_DifferentQueries (5 subtests)
// - TestMovementHandler_Create_AllFields
// - TestMovementHandler_Create_DifferentTypes (5 subtests)
// - TestMovementHandler_Update_AllFields
// These tests verified nil pointer panics, not business logic.

// Tests using MockMovementRepository for full code path coverage

func TestMovementHandler_ListAll_Success(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements", "")
	rr := httptest.NewRecorder()

	handler.ListAll(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "movements")
	assertBodyContains(t, rr, "Back Squat")
}

func TestMovementHandler_ListAll_Error(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	mockRepo.SetError(ErrMockInternalError)
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements", "")
	rr := httptest.NewRecorder()

	handler.ListAll(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to retrieve movements")
}

func TestMovementHandler_ListStandard_Success(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/standard", "")
	rr := httptest.NewRecorder()

	handler.ListStandard(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "movements")
}

func TestMovementHandler_ListStandard_Error(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	mockRepo.SetError(ErrMockInternalError)
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/standard", "")
	rr := httptest.NewRecorder()

	handler.ListStandard(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to retrieve movements")
}

func TestMovementHandler_Search_Success(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/search?q=squat", "")
	rr := httptest.NewRecorder()

	handler.Search(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "movements")
}

func TestMovementHandler_Search_Error(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	mockRepo.SetError(ErrMockInternalError)
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/search?q=squat", "")
	rr := httptest.NewRecorder()

	handler.Search(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to search movements")
}

func TestMovementHandler_Search_WithValidLimit(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/search?q=test&limit=5", "")
	rr := httptest.NewRecorder()

	handler.Search(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "movements")
}

func TestMovementHandler_GetByID_Success(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "Back Squat")
}

func TestMovementHandler_GetByID_NotFound(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/999", "")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestMovementHandler_GetByID_Error(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	mockRepo.SetError(ErrMockInternalError)
	handler := &MovementHandler{
		movementRepo: mockRepo,
		logger:       createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/movements/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to retrieve movement")
}

func TestMovementHandler_ListAll_NoLogger(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
	}

	req := createTestRequest(http.MethodGet, "/api/movements", "")
	rr := httptest.NewRecorder()

	handler.ListAll(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestMovementHandler_ListAll_ErrorNoLogger(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	mockRepo.SetError(ErrMockInternalError)
	handler := &MovementHandler{
		movementRepo: mockRepo,
	}

	req := createTestRequest(http.MethodGet, "/api/movements", "")
	rr := httptest.NewRecorder()

	handler.ListAll(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestMovementHandler_ListStandard_NoLogger(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
	}

	req := createTestRequest(http.MethodGet, "/api/movements/standard", "")
	rr := httptest.NewRecorder()

	handler.ListStandard(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestMovementHandler_Search_NoLogger(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
	}

	req := createTestRequest(http.MethodGet, "/api/movements/search?q=test", "")
	rr := httptest.NewRecorder()

	handler.Search(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestMovementHandler_GetByID_NoLogger(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	handler := &MovementHandler{
		movementRepo: mockRepo,
	}

	req := createTestRequest(http.MethodGet, "/api/movements/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

// =============================================
// Service-based tests for Create, Update, Delete
// =============================================

// createTestMovementService creates a MovementService with mock repositories
func createTestMovementService() *service.MovementService {
	mockMovementRepo := NewMockMovementRepository()
	mockDataChangeLogRepo := NewMockDataChangeLogRepository()
	mockAuditLogRepo := NewMockAuditLogRepository()
	dataChangeLogService := service.NewDataChangeLogService(mockDataChangeLogRepo)
	return service.NewMovementService(mockMovementRepo, dataChangeLogService, mockAuditLogRepo)
}

func TestMovementHandler_Create_Success(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	movementService := createTestMovementService()
	handler := &MovementHandler{
		movementRepo:    mockRepo,
		movementService: movementService,
		logger:          createTestLogger(),
	}

	body := `{"name": "New Movement", "type": "weightlifting", "description": "A new test movement"}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/movements", body, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assertStatusCode(t, rr, http.StatusCreated)
	assertBodyContains(t, rr, "New Movement")
}

func TestMovementHandler_Create_WithAllFields(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	movementService := createTestMovementService()
	handler := &MovementHandler{
		movementRepo:    mockRepo,
		movementService: movementService,
		logger:          createTestLogger(),
	}

	body := `{
		"name": "Complex Movement",
		"type": "gymnastics",
		"description": "A complex gymnastics movement",
		"equipment": "Rings",
		"video_url": "https://example.com/video"
	}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/movements", body, 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assertStatusCode(t, rr, http.StatusCreated)
}

func TestMovementHandler_Update_Success(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	movementService := createTestMovementService()
	handler := &MovementHandler{
		movementRepo:    mockRepo,
		movementService: movementService,
		logger:          createTestLogger(),
	}

	// Update a user-created movement (ID 4 is created by user 1)
	body := `{"name": "Updated Movement", "type": "strength", "description": "Updated description"}`
	req := createAuthenticatedRequest(http.MethodPut, "/api/movements/4", body, 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "4")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestMovementHandler_Update_AsAdmin(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	movementService := createTestMovementService()
	handler := &MovementHandler{
		movementRepo:    mockRepo,
		movementService: movementService,
		logger:          createTestLogger(),
	}

	// Admin can update standard movements
	body := `{"name": "Admin Updated", "type": "weightlifting", "description": "Admin update"}`
	req := createAuthenticatedRequest(http.MethodPut, "/api/movements/1", body, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestMovementHandler_Update_StandardMovementAsUser(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	movementService := createTestMovementService()
	handler := &MovementHandler{
		movementRepo:    mockRepo,
		movementService: movementService,
		logger:          createTestLogger(),
	}

	// Non-admin trying to update standard movement (ID 1)
	body := `{"name": "Unauthorized Update", "type": "strength"}`
	req := createAuthenticatedRequest(http.MethodPut, "/api/movements/1", body, 1, "user@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	// Should fail - can't update standard movements as non-admin
	assertStatusCode(t, rr, http.StatusForbidden)
	assertBodyContains(t, rr, "Cannot modify standard movement")
}

func TestMovementHandler_Delete_Success(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	movementService := createTestMovementService()
	handler := &MovementHandler{
		movementRepo:    mockRepo,
		movementService: movementService,
		logger:          createTestLogger(),
	}

	// Delete user-created movement (ID 4)
	req := createAuthenticatedRequest(http.MethodDelete, "/api/movements/4", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "4")
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	// Should succeed
	if rr.Code != http.StatusOK && rr.Code != http.StatusNoContent {
		t.Errorf("Expected success, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestMovementHandler_Delete_AsAdmin(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	movementService := createTestMovementService()
	handler := &MovementHandler{
		movementRepo:    mockRepo,
		movementService: movementService,
		logger:          createTestLogger(),
	}

	// Admin deleting a user-created movement (ID 4)
	req := createAuthenticatedRequest(http.MethodDelete, "/api/movements/4", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "4")
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	// Should succeed
	if rr.Code != http.StatusOK && rr.Code != http.StatusNoContent {
		t.Errorf("Expected success, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestMovementHandler_Delete_NotFound(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	movementService := createTestMovementService()
	handler := &MovementHandler{
		movementRepo:    mockRepo,
		movementService: movementService,
		logger:          createTestLogger(),
	}

	// Try to delete non-existent movement (ID 999)
	req := createAuthenticatedRequest(http.MethodDelete, "/api/movements/999", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	// Should fail - movement not found
	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestMovementHandler_Delete_StandardMovement(t *testing.T) {
	mockRepo := NewMockMovementRepository()
	movementService := createTestMovementService()
	handler := &MovementHandler{
		movementRepo:    mockRepo,
		movementService: movementService,
		logger:          createTestLogger(),
	}

	// Non-admin trying to delete standard movement (ID 1)
	req := createAuthenticatedRequest(http.MethodDelete, "/api/movements/1", "", 1, "test@example.com", "user")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	// Should fail - can't delete standard movements as non-admin
	assertStatusCode(t, rr, http.StatusForbidden)
	assertBodyContains(t, rr, "Cannot delete standard movement")
}
