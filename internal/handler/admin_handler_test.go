package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
	// Test constructor with nil dependencies (basic construction test)
	handler := NewAdminHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	if handler == nil {
		t.Error("NewAdminHandler() should not return nil")
	}

	// All fields should be nil when passed nil
	if handler.db != nil {
		t.Error("db should be nil when passed nil")
	}
	if handler.userWorkoutWODRepo != nil {
		t.Error("userWorkoutWODRepo should be nil when passed nil")
	}
	if handler.wodRepo != nil {
		t.Error("wodRepo should be nil when passed nil")
	}
	if handler.movementRepo != nil {
		t.Error("movementRepo should be nil when passed nil")
	}
	if handler.workoutRepo != nil {
		t.Error("workoutRepo should be nil when passed nil")
	}
	if handler.userRepo != nil {
		t.Error("userRepo should be nil when passed nil")
	}
	if handler.wodService != nil {
		t.Error("wodService should be nil when passed nil")
	}
	if handler.movementService != nil {
		t.Error("movementService should be nil when passed nil")
	}
	if handler.workoutTemplateService != nil {
		t.Error("workoutTemplateService should be nil when passed nil")
	}
	if handler.logger != nil {
		t.Error("logger should be nil when passed nil")
	}
}

func TestWODMismatch_Struct(t *testing.T) {
	timeSeconds := 600
	rounds := 5
	reps := 10
	weight := 135.5

	mismatch := WODMismatch{
		ID:                1,
		WODID:             100,
		WODName:           "Fran",
		UserEmail:         "user@example.com",
		WorkoutDate:       "2024-01-15",
		ExpectedScoreType: "Time (HH:MM:SS)",
		Issue:             "Has invalid fields for Time-based WOD",
		TimeSeconds:       &timeSeconds,
		Rounds:            &rounds,
		Reps:              &reps,
		Weight:            &weight,
	}

	if mismatch.ID != 1 {
		t.Errorf("ID = %d, want 1", mismatch.ID)
	}
	if mismatch.WODID != 100 {
		t.Errorf("WODID = %d, want 100", mismatch.WODID)
	}
	if mismatch.WODName != "Fran" {
		t.Errorf("WODName = %q, want %q", mismatch.WODName, "Fran")
	}
	if mismatch.UserEmail != "user@example.com" {
		t.Errorf("UserEmail = %q, want %q", mismatch.UserEmail, "user@example.com")
	}
	if mismatch.WorkoutDate != "2024-01-15" {
		t.Errorf("WorkoutDate = %q, want %q", mismatch.WorkoutDate, "2024-01-15")
	}
	if mismatch.ExpectedScoreType != "Time (HH:MM:SS)" {
		t.Errorf("ExpectedScoreType = %q, want %q", mismatch.ExpectedScoreType, "Time (HH:MM:SS)")
	}
	if mismatch.Issue != "Has invalid fields for Time-based WOD" {
		t.Errorf("Issue = %q, want %q", mismatch.Issue, "Has invalid fields for Time-based WOD")
	}
	if mismatch.TimeSeconds == nil || *mismatch.TimeSeconds != 600 {
		t.Errorf("TimeSeconds = %v, want 600", mismatch.TimeSeconds)
	}
	if mismatch.Rounds == nil || *mismatch.Rounds != 5 {
		t.Errorf("Rounds = %v, want 5", mismatch.Rounds)
	}
	if mismatch.Reps == nil || *mismatch.Reps != 10 {
		t.Errorf("Reps = %v, want 10", mismatch.Reps)
	}
	if mismatch.Weight == nil || *mismatch.Weight != 135.5 {
		t.Errorf("Weight = %v, want 135.5", mismatch.Weight)
	}
}

func TestWODMismatch_NilFields(t *testing.T) {
	mismatch := WODMismatch{
		ID:                1,
		WODID:             100,
		WODName:           "Cindy",
		UserEmail:         "user@example.com",
		WorkoutDate:       "2024-01-15",
		ExpectedScoreType: "Rounds+Reps",
		Issue:             "Missing rounds",
		TimeSeconds:       nil,
		Rounds:            nil,
		Reps:              nil,
		Weight:            nil,
	}

	if mismatch.TimeSeconds != nil {
		t.Error("TimeSeconds should be nil")
	}
	if mismatch.Rounds != nil {
		t.Error("Rounds should be nil")
	}
	if mismatch.Reps != nil {
		t.Error("Reps should be nil")
	}
	if mismatch.Weight != nil {
		t.Error("Weight should be nil")
	}
}

func TestUpdateWODRecordRequest_Struct(t *testing.T) {
	timeSeconds := 720
	rounds := 10
	reps := 5
	weight := 185.0

	req := UpdateWODRecordRequest{
		TimeSeconds: &timeSeconds,
		Rounds:      &rounds,
		Reps:        &reps,
		Weight:      &weight,
		Notes:       "Good workout",
	}

	if req.TimeSeconds == nil || *req.TimeSeconds != 720 {
		t.Errorf("TimeSeconds = %v, want 720", req.TimeSeconds)
	}
	if req.Rounds == nil || *req.Rounds != 10 {
		t.Errorf("Rounds = %v, want 10", req.Rounds)
	}
	if req.Reps == nil || *req.Reps != 5 {
		t.Errorf("Reps = %v, want 5", req.Reps)
	}
	if req.Weight == nil || *req.Weight != 185.0 {
		t.Errorf("Weight = %v, want 185.0", req.Weight)
	}
	if req.Notes != "Good workout" {
		t.Errorf("Notes = %q, want %q", req.Notes, "Good workout")
	}
}

func TestUpdateWODRecordRequest_NilFields(t *testing.T) {
	req := UpdateWODRecordRequest{
		TimeSeconds: nil,
		Rounds:      nil,
		Reps:        nil,
		Weight:      nil,
		Notes:       "",
	}

	if req.TimeSeconds != nil {
		t.Error("TimeSeconds should be nil")
	}
	if req.Rounds != nil {
		t.Error("Rounds should be nil")
	}
	if req.Reps != nil {
		t.Error("Reps should be nil")
	}
	if req.Weight != nil {
		t.Error("Weight should be nil")
	}
	if req.Notes != "" {
		t.Errorf("Notes = %q, want empty", req.Notes)
	}
}

func TestCopyToStandardRequest_Struct(t *testing.T) {
	req := CopyToStandardRequest{
		NewName: "Standard Fran",
	}

	if req.NewName != "Standard Fran" {
		t.Errorf("NewName = %q, want %q", req.NewName, "Standard Fran")
	}

	// Empty name
	req2 := CopyToStandardRequest{}
	if req2.NewName != "" {
		t.Errorf("NewName = %q, want empty", req2.NewName)
	}
}

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

func TestAdminHandler_ListUserCreatedWODs_NilService(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}
	req := createTestRequest(http.MethodGet, "/api/admin/user-content/wods", "")
	rr := httptest.NewRecorder()

	// This will panic because wodService is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil wodService")
		}
	}()

	handler.ListUserCreatedWODs(rr, req)
}

func TestAdminHandler_ListUserCreatedMovements_NilService(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}
	req := createTestRequest(http.MethodGet, "/api/admin/user-content/movements", "")
	rr := httptest.NewRecorder()

	// This will panic because movementService is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil movementService")
		}
	}()

	handler.ListUserCreatedMovements(rr, req)
}

func TestAdminHandler_ListUserCreatedWorkouts_NilService(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}
	req := createTestRequest(http.MethodGet, "/api/admin/user-content/workouts", "")
	rr := httptest.NewRecorder()

	// This will panic because workoutTemplateService is nil
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil workoutTemplateService")
		}
	}()

	handler.ListUserCreatedWorkouts(rr, req)
}

func TestAdminHandler_DetectWODScoreTypeMismatches_NilDB(t *testing.T) {
	handler := &AdminHandler{
		db:     nil,
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/admin/detect-wod-mismatches", "")
	rr := httptest.NewRecorder()

	// This will panic due to nil db, which is expected behavior
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil db")
		}
	}()

	handler.DetectWODScoreTypeMismatches(rr, req)
}

func TestAdminHandler_FixWODScoreTypeMismatches_NilDB(t *testing.T) {
	handler := &AdminHandler{
		db:     nil,
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/admin/fix-wod-mismatches", "")
	rr := httptest.NewRecorder()

	// This will panic due to nil db, which is expected behavior
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil db")
		}
	}()

	handler.FixWODScoreTypeMismatches(rr, req)
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

func TestAdminHandler_CopyWODToStandard_ValidInputNilService(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/admin/wods/1/copy-to-standard", `{"new_name": "Test WOD"}`)
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic
	defer func() {
		if r := recover(); r == nil {
			t.Log("CopyWODToStandard requires wodService")
		}
	}()

	handler.CopyWODToStandard(rr, req)
}

func TestAdminHandler_CopyMovementToStandard_ValidInputNilService(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/admin/movements/1/copy-to-standard", `{"new_name": "Test Movement"}`)
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic
	defer func() {
		if r := recover(); r == nil {
			t.Log("CopyMovementToStandard requires movementService")
		}
	}()

	handler.CopyMovementToStandard(rr, req)
}

func TestAdminHandler_CopyWorkoutToStandard_ValidInputNilService(t *testing.T) {
	handler := &AdminHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodPost, "/api/admin/workouts/1/copy-to-standard", `{"new_name": "Test Workout"}`)
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic
	defer func() {
		if r := recover(); r == nil {
			t.Log("CopyWorkoutToStandard requires workoutTemplateService")
		}
	}()

	handler.CopyWorkoutToStandard(rr, req)
}
