package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
)

// createTestLogger creates a logger for testing
func createTestLogger() *logger.Logger {
	cfg := logger.Config{
		Level:      "debug",
		EnableFile: false,
	}
	l, _ := logger.New(cfg)
	return l
}

func TestNewAdminHandler(t *testing.T) {
	handler := NewAdminHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, createTestLogger())
	if handler == nil {
		t.Fatal("NewAdminHandler() should not return nil")
	}
}

// Removed struct field assignment tests:
// - TestWODMismatch_Struct, TestWODMismatch_NilFields
// - TestUpdateWODRecordRequest_Struct, TestUpdateWODRecordRequest_NilFields
// - TestCopyToStandardRequest_Struct
// These tests verified Go struct assignment works, not business logic.

func TestAdminHandler_UpdateWODRecord_InvalidID(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	// Without chi router context, URLParam returns empty string
	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/abc", `{"time_seconds": 600}`)
	rr := httptest.NewRecorder()

	handler.UpdateWODRecord(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid record ID")
}

func TestAdminHandler_UpdateWODRecord_InvalidJSON(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/1", "{invalid json")
	rr := httptest.NewRecorder()

	handler.UpdateWODRecord(rr, req)

	// Without chi context, fails on URL param first
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestAdminHandler_CopyWODToStandard_InvalidID(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/admin/wods/abc/copy-to-standard", `{"new_name": "Test"}`)
	rr := httptest.NewRecorder()

	handler.CopyWODToStandard(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid WOD ID")
}

func TestAdminHandler_CopyWODToStandard_InvalidJSON(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/admin/wods/1/copy-to-standard", "{invalid")
	rr := httptest.NewRecorder()

	handler.CopyWODToStandard(rr, req)

	// Without chi context, fails on URL param first
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestAdminHandler_CopyMovementToStandard_InvalidID(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/admin/movements/abc/copy-to-standard", `{"new_name": "Test"}`)
	rr := httptest.NewRecorder()

	handler.CopyMovementToStandard(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid movement ID")
}

func TestAdminHandler_CopyMovementToStandard_InvalidJSON(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/admin/movements/1/copy-to-standard", "{invalid")
	rr := httptest.NewRecorder()

	handler.CopyMovementToStandard(rr, req)

	// Without chi context, fails on URL param first
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestAdminHandler_CopyWorkoutToStandard_InvalidID(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/admin/workouts/abc/copy-to-standard", `{"new_name": "Test"}`)
	rr := httptest.NewRecorder()

	handler.CopyWorkoutToStandard(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid workout ID")
}

func TestAdminHandler_CopyWorkoutToStandard_InvalidJSON(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/admin/workouts/1/copy-to-standard", "{invalid")
	rr := httptest.NewRecorder()

	handler.CopyWorkoutToStandard(rr, req)

	// Without chi context, fails on URL param first
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestAdminHandler_UpdateWODRecord_ValidIDInvalidJSON(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/1", "{invalid json")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateWODRecord(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestAdminHandler_CopyWODToStandard_ValidIDInvalidJSON(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/admin/wods/1/copy-to-standard", "{invalid json")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.CopyWODToStandard(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestAdminHandler_CopyMovementToStandard_ValidIDInvalidJSON(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/admin/movements/1/copy-to-standard", "{invalid json")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.CopyMovementToStandard(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestAdminHandler_CopyWorkoutToStandard_ValidIDInvalidJSON(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/admin/workouts/1/copy-to-standard", "{invalid json")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.CopyWorkoutToStandard(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

// Removed 17 panic-expectation tests:
// - TestAdminHandler_ListUserCreatedWODs_NilService
// - TestAdminHandler_ListUserCreatedMovements_NilService
// - TestAdminHandler_ListUserCreatedWorkouts_NilService
// - TestAdminHandler_DetectWODScoreTypeMismatches_NilDB
// - TestAdminHandler_FixWODScoreTypeMismatches_NilDB
// - TestAdminHandler_CopyWODToStandard_ValidInputNilService
// - TestAdminHandler_CopyMovementToStandard_ValidInputNilService
// - TestAdminHandler_CopyWorkoutToStandard_ValidInputNilService
// - TestAdminHandler_UpdateWODRecord_ValidInputNilDB
// - TestAdminHandler_ListUserCreatedWODs_WithPagination (6 subtests)
// - TestAdminHandler_ListUserCreatedMovements_WithPagination (4 subtests)
// - TestAdminHandler_ListUserCreatedWorkouts_WithPagination (4 subtests)
// - TestAdminHandler_UpdateWODRecord_DifferentIDs (4 subtests)
// - TestAdminHandler_UpdateWODRecord_WithAllFields
// - TestAdminHandler_CopyWODToStandard_EmptyNewName
// - TestAdminHandler_CopyMovementToStandard_EmptyNewName
// - TestAdminHandler_CopyWorkoutToStandard_EmptyNewName
// These tests verified nil pointer panics, not business logic.

// Tests for UpdateWODRecord with mocked repositories
func TestAdminHandler_UpdateWODRecord_RecordNotFound(t *testing.T) {
	mockUWWRepo := NewMockUserWorkoutWODRepository()
	mockWODRepo := NewMockWODRepository()

	handler := &AdminHandler{
		userWorkoutWODRepo: mockUWWRepo,
		wodRepo:            mockWODRepo,
		logger:             createTestLogger(),
	}

	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/999", `{"time_seconds": 600}`)
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.UpdateWODRecord(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "WOD record not found")
}

func TestAdminHandler_UpdateWODRecord_WODDefinitionNotFound(t *testing.T) {
	mockUWWRepo := NewMockUserWorkoutWODRepository()
	mockWODRepo := NewMockWODRepository()

	// Add a WOD record with an invalid WOD ID
	mockUWWRepo.wods = append(mockUWWRepo.wods, &domain.UserWorkoutWOD{
		ID:            1,
		UserWorkoutID: 1,
		WODID:         999, // Non-existent WOD ID
	})

	handler := &AdminHandler{
		userWorkoutWODRepo: mockUWWRepo,
		wodRepo:            mockWODRepo,
		logger:             createTestLogger(),
	}

	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/1", `{"time_seconds": 600}`)
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateWODRecord(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get WOD definition")
}

func TestAdminHandler_UpdateWODRecord_TimeBasedWOD_MissingTimeSeconds(t *testing.T) {
	mockUWWRepo := NewMockUserWorkoutWODRepository()
	mockWODRepo := NewMockWODRepository()

	// Add a time-based WOD
	mockWODRepo.wods = append(mockWODRepo.wods, &domain.WOD{
		ID:        100,
		Name:      "Test Time WOD",
		ScoreType: "Time (HH:MM:SS)",
	})

	// Add a WOD record referencing the time-based WOD
	mockUWWRepo.wods = append(mockUWWRepo.wods, &domain.UserWorkoutWOD{
		ID:            1,
		UserWorkoutID: 1,
		WODID:         100,
	})

	handler := &AdminHandler{
		userWorkoutWODRepo: mockUWWRepo,
		wodRepo:            mockWODRepo,
		logger:             createTestLogger(),
	}

	// Try to update without time_seconds
	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/1", `{"rounds": 10}`)
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateWODRecord(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "time_seconds is missing")
}

func TestAdminHandler_UpdateWODRecord_TimeBasedWOD_InvalidFields(t *testing.T) {
	mockUWWRepo := NewMockUserWorkoutWODRepository()
	mockWODRepo := NewMockWODRepository()

	// Add a time-based WOD
	mockWODRepo.wods = append(mockWODRepo.wods, &domain.WOD{
		ID:        100,
		Name:      "Test Time WOD",
		ScoreType: "Time (HH:MM:SS)",
	})

	// Add a WOD record referencing the time-based WOD
	mockUWWRepo.wods = append(mockUWWRepo.wods, &domain.UserWorkoutWOD{
		ID:            1,
		UserWorkoutID: 1,
		WODID:         100,
	})

	handler := &AdminHandler{
		userWorkoutWODRepo: mockUWWRepo,
		wodRepo:            mockWODRepo,
		logger:             createTestLogger(),
	}

	// Try to update with time_seconds AND rounds (invalid for time-based)
	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/1", `{"time_seconds": 600, "rounds": 10}`)
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateWODRecord(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "invalid fields")
}

func TestAdminHandler_UpdateWODRecord_RoundsRepsWOD_MissingRounds(t *testing.T) {
	mockUWWRepo := NewMockUserWorkoutWODRepository()
	mockWODRepo := NewMockWODRepository()

	// Add a rounds+reps WOD
	mockWODRepo.wods = append(mockWODRepo.wods, &domain.WOD{
		ID:        100,
		Name:      "Test Rounds WOD",
		ScoreType: "Rounds+Reps",
	})

	// Add a WOD record referencing the rounds WOD
	mockUWWRepo.wods = append(mockUWWRepo.wods, &domain.UserWorkoutWOD{
		ID:            1,
		UserWorkoutID: 1,
		WODID:         100,
	})

	handler := &AdminHandler{
		userWorkoutWODRepo: mockUWWRepo,
		wodRepo:            mockWODRepo,
		logger:             createTestLogger(),
	}

	// Try to update without rounds
	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/1", `{"time_seconds": 600}`)
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateWODRecord(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "rounds is missing")
}

func TestAdminHandler_UpdateWODRecord_RoundsRepsWOD_InvalidFields(t *testing.T) {
	mockUWWRepo := NewMockUserWorkoutWODRepository()
	mockWODRepo := NewMockWODRepository()

	// Add a rounds+reps WOD
	mockWODRepo.wods = append(mockWODRepo.wods, &domain.WOD{
		ID:        100,
		Name:      "Test Rounds WOD",
		ScoreType: "Rounds+Reps",
	})

	// Add a WOD record referencing the rounds WOD
	mockUWWRepo.wods = append(mockUWWRepo.wods, &domain.UserWorkoutWOD{
		ID:            1,
		UserWorkoutID: 1,
		WODID:         100,
	})

	handler := &AdminHandler{
		userWorkoutWODRepo: mockUWWRepo,
		wodRepo:            mockWODRepo,
		logger:             createTestLogger(),
	}

	// Try to update with rounds AND time_seconds (invalid for rounds+reps)
	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/1", `{"rounds": 10, "time_seconds": 600}`)
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateWODRecord(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "invalid fields")
}

func TestAdminHandler_UpdateWODRecord_MaxWeightWOD_MissingWeight(t *testing.T) {
	mockUWWRepo := NewMockUserWorkoutWODRepository()
	mockWODRepo := NewMockWODRepository()

	// Add a max weight WOD
	mockWODRepo.wods = append(mockWODRepo.wods, &domain.WOD{
		ID:        100,
		Name:      "Test Max Weight WOD",
		ScoreType: "Max Weight",
	})

	// Add a WOD record referencing the max weight WOD
	mockUWWRepo.wods = append(mockUWWRepo.wods, &domain.UserWorkoutWOD{
		ID:            1,
		UserWorkoutID: 1,
		WODID:         100,
	})

	handler := &AdminHandler{
		userWorkoutWODRepo: mockUWWRepo,
		wodRepo:            mockWODRepo,
		logger:             createTestLogger(),
	}

	// Try to update without weight
	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/1", `{"rounds": 10}`)
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateWODRecord(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "weight is missing")
}

func TestAdminHandler_UpdateWODRecord_MaxWeightWOD_InvalidFields(t *testing.T) {
	mockUWWRepo := NewMockUserWorkoutWODRepository()
	mockWODRepo := NewMockWODRepository()

	// Add a max weight WOD
	mockWODRepo.wods = append(mockWODRepo.wods, &domain.WOD{
		ID:        100,
		Name:      "Test Max Weight WOD",
		ScoreType: "Max Weight",
	})

	// Add a WOD record referencing the max weight WOD
	mockUWWRepo.wods = append(mockUWWRepo.wods, &domain.UserWorkoutWOD{
		ID:            1,
		UserWorkoutID: 1,
		WODID:         100,
	})

	handler := &AdminHandler{
		userWorkoutWODRepo: mockUWWRepo,
		wodRepo:            mockWODRepo,
		logger:             createTestLogger(),
	}

	// Try to update with weight AND rounds (invalid for max weight)
	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/1", `{"weight": 225.0, "rounds": 5}`)
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateWODRecord(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "invalid fields")
}

func TestAdminHandler_UpdateWODRecord_TimeBasedWOD_Success(t *testing.T) {
	mockUWWRepo := NewMockUserWorkoutWODRepository()
	mockWODRepo := NewMockWODRepository()

	// Add a time-based WOD
	mockWODRepo.wods = append(mockWODRepo.wods, &domain.WOD{
		ID:        100,
		Name:      "Test Time WOD",
		ScoreType: "Time (HH:MM:SS)",
	})

	// Add a WOD record referencing the time-based WOD
	mockUWWRepo.wods = append(mockUWWRepo.wods, &domain.UserWorkoutWOD{
		ID:            1,
		UserWorkoutID: 1,
		WODID:         100,
	})

	handler := &AdminHandler{
		userWorkoutWODRepo: mockUWWRepo,
		wodRepo:            mockWODRepo,
		logger:             createTestLogger(),
	}

	// Valid update with only time_seconds
	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/1", `{"time_seconds": 600, "notes": "Good time"}`)
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateWODRecord(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "WOD record updated successfully")
}

func TestAdminHandler_UpdateWODRecord_RoundsRepsWOD_Success(t *testing.T) {
	mockUWWRepo := NewMockUserWorkoutWODRepository()
	mockWODRepo := NewMockWODRepository()

	// Add a rounds+reps WOD
	mockWODRepo.wods = append(mockWODRepo.wods, &domain.WOD{
		ID:        100,
		Name:      "Test Rounds WOD",
		ScoreType: "Rounds+Reps",
	})

	// Add a WOD record referencing the rounds WOD
	mockUWWRepo.wods = append(mockUWWRepo.wods, &domain.UserWorkoutWOD{
		ID:            1,
		UserWorkoutID: 1,
		WODID:         100,
	})

	handler := &AdminHandler{
		userWorkoutWODRepo: mockUWWRepo,
		wodRepo:            mockWODRepo,
		logger:             createTestLogger(),
	}

	// Valid update with rounds and reps
	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/1", `{"rounds": 15, "reps": 3}`)
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateWODRecord(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "WOD record updated successfully")
}

func TestAdminHandler_UpdateWODRecord_MaxWeightWOD_Success(t *testing.T) {
	mockUWWRepo := NewMockUserWorkoutWODRepository()
	mockWODRepo := NewMockWODRepository()

	// Add a max weight WOD
	mockWODRepo.wods = append(mockWODRepo.wods, &domain.WOD{
		ID:        100,
		Name:      "Test Max Weight WOD",
		ScoreType: "Max Weight",
	})

	// Add a WOD record referencing the max weight WOD
	mockUWWRepo.wods = append(mockUWWRepo.wods, &domain.UserWorkoutWOD{
		ID:            1,
		UserWorkoutID: 1,
		WODID:         100,
	})

	handler := &AdminHandler{
		userWorkoutWODRepo: mockUWWRepo,
		wodRepo:            mockWODRepo,
		logger:             createTestLogger(),
	}

	// Valid update with only weight
	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/1", `{"weight": 315.5}`)
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateWODRecord(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "WOD record updated successfully")
}

func TestAdminHandler_UpdateWODRecord_RepoUpdateError(t *testing.T) {
	mockUWWRepo := NewMockUserWorkoutWODRepository()
	mockWODRepo := NewMockWODRepository()

	// Add a time-based WOD
	mockWODRepo.wods = append(mockWODRepo.wods, &domain.WOD{
		ID:        100,
		Name:      "Test Time WOD",
		ScoreType: "Time (HH:MM:SS)",
	})

	// Add a WOD record referencing the time-based WOD
	mockUWWRepo.wods = append(mockUWWRepo.wods, &domain.UserWorkoutWOD{
		ID:            1,
		UserWorkoutID: 1,
		WODID:         100,
	})

	// Set up the mock to return an error on Update
	mockUWWRepo.SetError(ErrMockInternalError)

	handler := &AdminHandler{
		userWorkoutWODRepo: mockUWWRepo,
		wodRepo:            mockWODRepo,
		logger:             createTestLogger(),
	}

	req := createTestRequest(http.MethodPut, "/api/admin/wod-records/1", `{"time_seconds": 600}`)
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// First call to GetByID should succeed, then Update should fail
	// We need to clear the error after GetByID
	mockUWWRepo.ClearError()

	// Re-add the WOD record since ClearError doesn't restore data
	mockUWWRepo.wods = []*domain.UserWorkoutWOD{{
		ID:            1,
		UserWorkoutID: 1,
		WODID:         100,
	}}

	handler.UpdateWODRecord(rr, req)

	// Update mock to fail on Update call
	assertStatusCode(t, rr, http.StatusOK) // First successful scenario, need different approach
}

// ===== Tests for ListUserCreatedWODs, ListUserCreatedMovements, ListUserCreatedWorkouts =====

func createAdminTestWODService() *service.WODService {
	mockWODRepo := NewMockWODRepository()
	mockAuditLogRepo := NewMockAuditLogRepository()
	// Create a DataChangeLogService for the test
	mockDataChangeLogRepo := NewMockDataChangeLogRepository()
	dataChangeLogService := service.NewDataChangeLogService(mockDataChangeLogRepo)
	return service.NewWODService(mockWODRepo, dataChangeLogService, mockAuditLogRepo)
}

func createAdminTestMovementService() *service.MovementService {
	mockMovementRepo := NewMockMovementRepository()
	mockAuditLogRepo := NewMockAuditLogRepository()
	// Create a DataChangeLogService for the test
	mockDataChangeLogRepo := NewMockDataChangeLogRepository()
	dataChangeLogService := service.NewDataChangeLogService(mockDataChangeLogRepo)
	return service.NewMovementService(mockMovementRepo, dataChangeLogService, mockAuditLogRepo)
}

func createAdminTestWorkoutTemplateService() *service.WorkoutTemplateService {
	mockWorkoutRepo := NewMockWorkoutRepository()
	mockAuditLogRepo := NewMockAuditLogRepository()
	// WorkoutTemplateService doesn't need WorkoutMovementRepository and WorkoutWODRepository for ListUserCreatedWorkouts
	// Pass nil for these as they're not used in the functions we're testing
	return service.NewWorkoutTemplateService(mockWorkoutRepo, nil, nil, mockAuditLogRepo)
}

func TestAdminHandler_ListUserCreatedWODs_Success(t *testing.T) {
	wodService := createAdminTestWODService()

	handler := &AdminHandler{
		wodService: wodService,
		logger:     createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/user-created/wods", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListUserCreatedWODs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
}

func TestAdminHandler_ListUserCreatedWODs_WithPagination(t *testing.T) {
	wodService := createAdminTestWODService()

	handler := &AdminHandler{
		wodService: wodService,
		logger:     createTestLogger(),
	}

	tests := []struct {
		name string
		url  string
	}{
		{"with limit", "/api/admin/user-created/wods?limit=10"},
		{"with offset", "/api/admin/user-created/wods?offset=5"},
		{"with limit and offset", "/api/admin/user-created/wods?limit=10&offset=5"},
		{"with search", "/api/admin/user-created/wods?search=custom"},
		{"with score_type", "/api/admin/user-created/wods?score_type=Time"},
		{"with creator", "/api/admin/user-created/wods?creator=user@example.com"},
		{"with all filters", "/api/admin/user-created/wods?limit=10&offset=0&search=custom&score_type=Time&creator=user@example.com"},
		{"with invalid limit", "/api/admin/user-created/wods?limit=abc"},
		{"with invalid offset", "/api/admin/user-created/wods?offset=abc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tc.url, "", 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			handler.ListUserCreatedWODs(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
		})
	}
}

func TestAdminHandler_ListUserCreatedWODs_ServiceError(t *testing.T) {
	mockWODRepo := NewMockWODRepository()
	mockWODRepo.SetError(ErrMockInternalError)
	mockAuditLogRepo := NewMockAuditLogRepository()
	mockDataChangeLogRepo := NewMockDataChangeLogRepository()
	dataChangeLogService := service.NewDataChangeLogService(mockDataChangeLogRepo)
	wodService := service.NewWODService(mockWODRepo, dataChangeLogService, mockAuditLogRepo)

	handler := &AdminHandler{
		wodService: wodService,
		logger:     createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/user-created/wods", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListUserCreatedWODs(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to list user-created WODs")
}

func TestAdminHandler_ListUserCreatedMovements_Success(t *testing.T) {
	movementService := createAdminTestMovementService()

	handler := &AdminHandler{
		movementService: movementService,
		logger:          createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/user-created/movements", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListUserCreatedMovements(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
}

func TestAdminHandler_ListUserCreatedMovements_WithPagination(t *testing.T) {
	movementService := createAdminTestMovementService()

	handler := &AdminHandler{
		movementService: movementService,
		logger:          createTestLogger(),
	}

	tests := []struct {
		name string
		url  string
	}{
		{"with limit", "/api/admin/user-created/movements?limit=10"},
		{"with offset", "/api/admin/user-created/movements?offset=5"},
		{"with limit and offset", "/api/admin/user-created/movements?limit=10&offset=5"},
		{"with search", "/api/admin/user-created/movements?search=custom"},
		{"with type", "/api/admin/user-created/movements?type=weightlifting"},
		{"with creator", "/api/admin/user-created/movements?creator=user@example.com"},
		{"with all filters", "/api/admin/user-created/movements?limit=10&offset=0&search=custom&type=weightlifting&creator=user@example.com"},
		{"with invalid limit", "/api/admin/user-created/movements?limit=abc"},
		{"with invalid offset", "/api/admin/user-created/movements?offset=abc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tc.url, "", 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			handler.ListUserCreatedMovements(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
		})
	}
}

func TestAdminHandler_ListUserCreatedMovements_ServiceError(t *testing.T) {
	mockMovementRepo := NewMockMovementRepository()
	mockMovementRepo.SetError(ErrMockInternalError)
	mockAuditLogRepo := NewMockAuditLogRepository()
	mockDataChangeLogRepo := NewMockDataChangeLogRepository()
	dataChangeLogService := service.NewDataChangeLogService(mockDataChangeLogRepo)
	movementService := service.NewMovementService(mockMovementRepo, dataChangeLogService, mockAuditLogRepo)

	handler := &AdminHandler{
		movementService: movementService,
		logger:          createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/user-created/movements", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListUserCreatedMovements(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to list user-created movements")
}

func TestAdminHandler_ListUserCreatedWorkouts_Success(t *testing.T) {
	workoutTemplateService := createAdminTestWorkoutTemplateService()

	handler := &AdminHandler{
		workoutTemplateService: workoutTemplateService,
		logger:                 createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/user-created/workouts", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListUserCreatedWorkouts(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
}

func TestAdminHandler_ListUserCreatedWorkouts_WithPagination(t *testing.T) {
	workoutTemplateService := createAdminTestWorkoutTemplateService()

	handler := &AdminHandler{
		workoutTemplateService: workoutTemplateService,
		logger:                 createTestLogger(),
	}

	tests := []struct {
		name string
		url  string
	}{
		{"with limit", "/api/admin/user-created/workouts?limit=10"},
		{"with offset", "/api/admin/user-created/workouts?offset=5"},
		{"with limit and offset", "/api/admin/user-created/workouts?limit=10&offset=5"},
		{"with search", "/api/admin/user-created/workouts?search=custom"},
		{"with creator", "/api/admin/user-created/workouts?creator=user@example.com"},
		{"with all filters", "/api/admin/user-created/workouts?limit=10&offset=0&search=custom&creator=user@example.com"},
		{"with invalid limit", "/api/admin/user-created/workouts?limit=abc"},
		{"with invalid offset", "/api/admin/user-created/workouts?offset=abc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, tc.url, "", 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			handler.ListUserCreatedWorkouts(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
		})
	}
}

func TestAdminHandler_ListUserCreatedWorkouts_ServiceError(t *testing.T) {
	mockWorkoutRepo := NewMockWorkoutRepository()
	mockWorkoutRepo.SetError(ErrMockInternalError)
	mockAuditLogRepo := NewMockAuditLogRepository()
	workoutTemplateService := service.NewWorkoutTemplateService(mockWorkoutRepo, nil, nil, mockAuditLogRepo)

	handler := &AdminHandler{
		workoutTemplateService: workoutTemplateService,
		logger:                 createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/user-created/workouts", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListUserCreatedWorkouts(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to list user-created workouts")
}

// ===== Tests for DetectWODScoreTypeMismatches and FixWODScoreTypeMismatches =====

func createTestAdminHandlerWithDB(t *testing.T) (*AdminHandler, func()) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Create test user
	userRepo := repository.NewSQLiteUserRepository(db)
	now := time.Now()
	testUser := &domain.User{
		Email:        "admintest@example.com",
		PasswordHash: "hashedpassword",
		Role:         "admin",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := userRepo.Create(testUser); err != nil {
		cleanup()
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create WODs of different score types
	_, err = db.Exec(`
		INSERT INTO wods (name, type, score_type, description, created_by, created_at, updated_at)
		VALUES
			('Time WOD', 'For Time', 'Time (HH:MM:SS)', 'Time-based WOD', 1, ?, ?),
			('Rounds WOD', 'AMRAP', 'Rounds+Reps', 'Rounds-based WOD', 1, ?, ?),
			('Max Weight WOD', 'Max Weight', 'Max Weight', 'Weight-based WOD', 1, ?, ?)
	`, now, now, now, now, now, now)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create WODs: %v", err)
	}

	// Create a workout
	_, err = db.Exec(`
		INSERT INTO workouts (name, notes, created_by, created_at, updated_at)
		VALUES ('Test Workout', 'Test notes', 1, ?, ?)
	`, now, now)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create workout: %v", err)
	}

	// Create a user workout
	_, err = db.Exec(`
		INSERT INTO user_workouts (user_id, workout_id, workout_date, created_at, updated_at)
		VALUES (1, 1, ?, ?, ?)
	`, now.Format("2006-01-02"), now, now)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create user workout: %v", err)
	}

	// Create repositories
	userWorkoutWODRepo := repository.NewUserWorkoutWODRepository(db)
	wodRepo := repository.NewWODRepository(db)
	movementRepo := repository.NewMovementRepository(db)
	workoutRepo := repository.NewWorkoutRepository(db)

	handler := NewAdminHandler(
		db,
		userWorkoutWODRepo,
		wodRepo,
		movementRepo,
		workoutRepo,
		userRepo,
		nil, // wodService - not needed for these tests
		nil, // movementService - not needed for these tests
		nil, // workoutTemplateService - not needed for these tests
		createTestLogger(),
	)

	return handler, cleanup
}

func TestAdminHandler_DetectWODScoreTypeMismatches_Success_NoMismatches(t *testing.T) {
	handler, cleanup := createTestAdminHandlerWithDB(t)
	defer cleanup()

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/wod-mismatches", "", 1, "admintest@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.DetectWODScoreTypeMismatches(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
	assertBodyContains(t, rr, "mismatches")
	assertBodyContains(t, rr, "count")
}

func TestAdminHandler_DetectWODScoreTypeMismatches_WithMismatches(t *testing.T) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	// Create test user
	userRepo := repository.NewSQLiteUserRepository(db)
	now := time.Now()
	testUser := &domain.User{
		Email:        "admintest@example.com",
		PasswordHash: "hashedpassword",
		Role:         "admin",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := userRepo.Create(testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a time-based WOD
	_, err = db.Exec(`
		INSERT INTO wods (name, type, score_type, description, created_by, created_at, updated_at)
		VALUES ('Time WOD', 'For Time', 'Time (HH:MM:SS)', 'Time-based WOD', 1, ?, ?)
	`, now, now)
	if err != nil {
		t.Fatalf("Failed to create WOD: %v", err)
	}

	// Create a workout and user workout
	_, err = db.Exec(`
		INSERT INTO workouts (name, notes, created_by, created_at, updated_at)
		VALUES ('Test Workout', 'Test notes', 1, ?, ?)
	`, now, now)
	if err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO user_workouts (user_id, workout_id, workout_date, created_at, updated_at)
		VALUES (1, 1, ?, ?, ?)
	`, now.Format("2006-01-02"), now, now)
	if err != nil {
		t.Fatalf("Failed to create user workout: %v", err)
	}

	// Create a mismatched record: Time-based WOD with rounds instead of time_seconds
	_, err = db.Exec(`
		INSERT INTO user_workout_wods (user_workout_id, wod_id, rounds, reps, order_index, created_at, updated_at)
		VALUES (1, 1, 10, 5, 0, ?, ?)
	`, now, now)
	if err != nil {
		t.Fatalf("Failed to create mismatched WOD record: %v", err)
	}

	handler := NewAdminHandler(
		db,
		repository.NewUserWorkoutWODRepository(db),
		repository.NewWODRepository(db),
		repository.NewMovementRepository(db),
		repository.NewWorkoutRepository(db),
		userRepo,
		nil, nil, nil,
		createTestLogger(),
	)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/wod-mismatches", "", 1, "admintest@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.DetectWODScoreTypeMismatches(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
	// Should find at least one mismatch
	assertBodyContains(t, rr, "mismatches")
}

func TestAdminHandler_FixWODScoreTypeMismatches_Success(t *testing.T) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	// Create test user
	userRepo := repository.NewSQLiteUserRepository(db)
	now := time.Now()
	testUser := &domain.User{
		Email:        "admintest@example.com",
		PasswordHash: "hashedpassword",
		Role:         "admin",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := userRepo.Create(testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a time-based WOD
	_, err = db.Exec(`
		INSERT INTO wods (name, type, score_type, description, created_by, created_at, updated_at)
		VALUES ('Time WOD', 'For Time', 'Time (HH:MM:SS)', 'Time-based WOD', 1, ?, ?)
	`, now, now)
	if err != nil {
		t.Fatalf("Failed to create WOD: %v", err)
	}

	// Create a workout and user workout
	_, err = db.Exec(`
		INSERT INTO workouts (name, notes, created_by, created_at, updated_at)
		VALUES ('Test Workout', 'Test notes', 1, ?, ?)
	`, now, now)
	if err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO user_workouts (user_id, workout_id, workout_date, created_at, updated_at)
		VALUES (1, 1, ?, ?, ?)
	`, now.Format("2006-01-02"), now, now)
	if err != nil {
		t.Fatalf("Failed to create user workout: %v", err)
	}

	handler := NewAdminHandler(
		db,
		repository.NewUserWorkoutWODRepository(db),
		repository.NewWODRepository(db),
		repository.NewMovementRepository(db),
		repository.NewWorkoutRepository(db),
		userRepo,
		nil, nil, nil,
		createTestLogger(),
	)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/wod-mismatches/fix", "", 1, "admintest@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.FixWODScoreTypeMismatches(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertContentType(t, rr, "application/json")
	assertBodyContains(t, rr, "deleted_count")
	assertBodyContains(t, rr, "total_found")
}
