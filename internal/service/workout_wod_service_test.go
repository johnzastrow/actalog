package service

import (
	"errors"
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestNewWorkoutWODService(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	if svc == nil {
		t.Fatal("NewWorkoutWODService returned nil")
	}
}

func TestWorkoutWODService_AddWODToWorkout(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{
		Name: "Test WOD",
	}
	_ = wodRepo.Create(wod)

	// Add WOD to workout
	division := "RX"
	result, err := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, &division)
	if err != nil {
		t.Errorf("AddWODToWorkout() error = %v", err)
	}
	if result == nil {
		t.Fatal("AddWODToWorkout() returned nil")
	}
	if result.WorkoutID != workout.ID {
		t.Errorf("WorkoutID = %d, want %d", result.WorkoutID, workout.ID)
	}
	if result.WODID != wod.ID {
		t.Errorf("WODID = %d, want %d", result.WODID, wod.ID)
	}
	if result.Division == nil || *result.Division != "RX" {
		t.Error("Division should be RX")
	}
}

func TestWorkoutWODService_AddWODToWorkout_WorkoutNotFound(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Try to add WOD to non-existent workout
	_, err := svc.AddWODToWorkout(999, 1, 1, 0, nil)
	if err == nil {
		t.Error("AddWODToWorkout() should return error when workout not found")
	}
}

func TestWorkoutWODService_AddWODToWorkout_Unauthorized(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Try to add WOD as different user
	_, err := svc.AddWODToWorkout(workout.ID, wod.ID, 2, 0, nil)
	if err != ErrUnauthorized {
		t.Errorf("AddWODToWorkout() error = %v, want ErrUnauthorized", err)
	}
}

func TestWorkoutWODService_AddWODToWorkout_WODNotFound(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Try to add non-existent WOD
	_, err := svc.AddWODToWorkout(workout.ID, 999, userID, 0, nil)
	if err == nil {
		t.Error("AddWODToWorkout() should return error when WOD not found")
	}
}

func TestWorkoutWODService_RemoveWODFromWorkout(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Add WOD to workout
	workoutWOD, _ := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, nil)

	// Remove WOD from workout
	err := svc.RemoveWODFromWorkout(workoutWOD.ID, userID)
	if err != nil {
		t.Errorf("RemoveWODFromWorkout() error = %v", err)
	}

	// Verify it was deleted
	deleted, _ := workoutWODRepo.GetByID(workoutWOD.ID)
	if deleted != nil {
		t.Error("WorkoutWOD should be deleted")
	}
}

func TestWorkoutWODService_RemoveWODFromWorkout_NotFound(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Try to remove non-existent workout WOD
	err := svc.RemoveWODFromWorkout(999, 1)
	if err == nil {
		t.Error("RemoveWODFromWorkout() should return error when not found")
	}
}

func TestWorkoutWODService_RemoveWODFromWorkout_Unauthorized(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Add WOD to workout
	workoutWOD, _ := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, nil)

	// Try to remove as different user
	err := svc.RemoveWODFromWorkout(workoutWOD.ID, 2)
	if err != ErrUnauthorized {
		t.Errorf("RemoveWODFromWorkout() error = %v, want ErrUnauthorized", err)
	}
}

func TestWorkoutWODService_UpdateWorkoutWOD(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Add WOD to workout
	workoutWOD, _ := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, nil)

	// Update the workout WOD
	scoreValue := "10:30"
	division := "Scaled"
	err := svc.UpdateWorkoutWOD(workoutWOD.ID, userID, &scoreValue, &division)
	if err != nil {
		t.Errorf("UpdateWorkoutWOD() error = %v", err)
	}

	// Verify updates
	updated, _ := workoutWODRepo.GetByID(workoutWOD.ID)
	if updated.ScoreValue == nil || *updated.ScoreValue != "10:30" {
		t.Error("ScoreValue should be 10:30")
	}
	if updated.Division == nil || *updated.Division != "Scaled" {
		t.Error("Division should be Scaled")
	}
}

func TestWorkoutWODService_UpdateWorkoutWOD_NotFound(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Try to update non-existent workout WOD
	scoreValue := "10:30"
	err := svc.UpdateWorkoutWOD(999, 1, &scoreValue, nil)
	if err == nil {
		t.Error("UpdateWorkoutWOD() should return error when not found")
	}
}

func TestWorkoutWODService_UpdateWorkoutWOD_Unauthorized(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Add WOD to workout
	workoutWOD, _ := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, nil)

	// Try to update as different user
	scoreValue := "10:30"
	err := svc.UpdateWorkoutWOD(workoutWOD.ID, 2, &scoreValue, nil)
	if err != ErrUnauthorized {
		t.Errorf("UpdateWorkoutWOD() error = %v, want ErrUnauthorized", err)
	}
}

func TestWorkoutWODService_UpdateWorkoutWOD_PartialUpdate(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Add WOD to workout with initial division
	initialDivision := "RX"
	workoutWOD, _ := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, &initialDivision)

	// Update only score value (division should remain unchanged)
	scoreValue := "10:30"
	err := svc.UpdateWorkoutWOD(workoutWOD.ID, userID, &scoreValue, nil)
	if err != nil {
		t.Errorf("UpdateWorkoutWOD() error = %v", err)
	}

	// Verify updates
	updated, _ := workoutWODRepo.GetByID(workoutWOD.ID)
	if updated.ScoreValue == nil || *updated.ScoreValue != "10:30" {
		t.Error("ScoreValue should be 10:30")
	}
	if updated.Division == nil || *updated.Division != "RX" {
		t.Error("Division should remain RX")
	}
}

func TestWorkoutWODService_ToggleWODPR(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Add WOD to workout (IsPR defaults to false)
	workoutWOD, _ := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, nil)
	if workoutWOD.IsPR {
		t.Error("IsPR should default to false")
	}

	// Toggle PR
	err := svc.ToggleWODPR(workoutWOD.ID, userID)
	if err != nil {
		t.Errorf("ToggleWODPR() error = %v", err)
	}

	// Verify it was toggled
	updated, _ := workoutWODRepo.GetByID(workoutWOD.ID)
	if !updated.IsPR {
		t.Error("IsPR should be true after toggle")
	}

	// Toggle again
	err = svc.ToggleWODPR(workoutWOD.ID, userID)
	if err != nil {
		t.Errorf("ToggleWODPR() error = %v", err)
	}

	// Verify it was toggled back
	updated, _ = workoutWODRepo.GetByID(workoutWOD.ID)
	if updated.IsPR {
		t.Error("IsPR should be false after second toggle")
	}
}

func TestWorkoutWODService_ToggleWODPR_NotFound(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Try to toggle PR on non-existent workout WOD
	err := svc.ToggleWODPR(999, 1)
	if err == nil {
		t.Error("ToggleWODPR() should return error when not found")
	}
}

func TestWorkoutWODService_ToggleWODPR_Unauthorized(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Add WOD to workout
	workoutWOD, _ := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, nil)

	// Try to toggle PR as different user
	err := svc.ToggleWODPR(workoutWOD.ID, 2)
	if err != ErrUnauthorized {
		t.Errorf("ToggleWODPR() error = %v, want ErrUnauthorized", err)
	}
}

func TestWorkoutWODService_ListWODsForWorkout(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create multiple WODs
	wod1 := &domain.WOD{Name: "WOD 1"}
	wod2 := &domain.WOD{Name: "WOD 2"}
	wod3 := &domain.WOD{Name: "WOD 3"}
	_ = wodRepo.Create(wod1)
	_ = wodRepo.Create(wod2)
	_ = wodRepo.Create(wod3)

	// Add WODs to workout
	_, _ = svc.AddWODToWorkout(workout.ID, wod1.ID, userID, 0, nil)
	_, _ = svc.AddWODToWorkout(workout.ID, wod2.ID, userID, 1, nil)
	_, _ = svc.AddWODToWorkout(workout.ID, wod3.ID, userID, 2, nil)

	// List WODs for workout
	wods, err := svc.ListWODsForWorkout(workout.ID)
	if err != nil {
		t.Errorf("ListWODsForWorkout() error = %v", err)
	}
	if len(wods) != 3 {
		t.Errorf("ListWODsForWorkout() returned %d WODs, want 3", len(wods))
	}
}

func TestWorkoutWODService_ListWODsForWorkout_Empty(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template with no WODs
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Empty Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// List WODs for workout
	wods, err := svc.ListWODsForWorkout(workout.ID)
	if err != nil {
		t.Errorf("ListWODsForWorkout() error = %v", err)
	}
	if len(wods) != 0 {
		t.Errorf("ListWODsForWorkout() returned %d WODs, want 0", len(wods))
	}
}

func TestWorkoutWODService_AddWODToWorkout_NilCreatedBy(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a standard workout template (no CreatedBy)
	workout := &domain.Workout{
		Name:      "Standard Workout",
		CreatedBy: nil,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Try to add WOD - should fail authorization
	_, err := svc.AddWODToWorkout(workout.ID, wod.ID, 1, 0, nil)
	if err != ErrUnauthorized {
		t.Errorf("AddWODToWorkout() error = %v, want ErrUnauthorized for nil CreatedBy", err)
	}
}

// Error path tests

func TestWorkoutWODService_AddWODToWorkout_WorkoutRepoError(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	// Set error on workout repo
	workoutRepo.getByIDError = errors.New("database error")

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	_, err := svc.AddWODToWorkout(1, 1, 1, 0, nil)
	if err == nil {
		t.Error("AddWODToWorkout() should return error when workout repo fails")
	}
}

func TestWorkoutWODService_AddWODToWorkout_WODRepoError(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Set error on WOD repo
	wodRepo.getByIDError = errors.New("database error")

	_, err := svc.AddWODToWorkout(workout.ID, 1, userID, 0, nil)
	if err == nil {
		t.Error("AddWODToWorkout() should return error when WOD repo fails")
	}
}

func TestWorkoutWODService_AddWODToWorkout_CreateError(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Set error on create
	workoutWODRepo.createError = errors.New("database error")

	_, err := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, nil)
	if err == nil {
		t.Error("AddWODToWorkout() should return error when create fails")
	}
}

func TestWorkoutWODService_RemoveWODFromWorkout_GetByIDError(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	// Set error on GetByID
	workoutWODRepo.getByIDError = errors.New("database error")

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	err := svc.RemoveWODFromWorkout(1, 1)
	if err == nil {
		t.Error("RemoveWODFromWorkout() should return error when GetByID fails")
	}
}

func TestWorkoutWODService_RemoveWODFromWorkout_GetWorkoutError(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Add WOD to workout
	workoutWOD, _ := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, nil)

	// Set error on workout repo
	workoutRepo.getByIDError = errors.New("database error")

	err := svc.RemoveWODFromWorkout(workoutWOD.ID, userID)
	if err == nil {
		t.Error("RemoveWODFromWorkout() should return error when get workout fails")
	}
}

func TestWorkoutWODService_RemoveWODFromWorkout_DeleteError(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Add WOD to workout
	workoutWOD, _ := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, nil)

	// Set error on delete
	workoutWODRepo.deleteError = errors.New("database error")

	err := svc.RemoveWODFromWorkout(workoutWOD.ID, userID)
	if err == nil {
		t.Error("RemoveWODFromWorkout() should return error when delete fails")
	}
}

func TestWorkoutWODService_RemoveWODFromWorkout_WorkoutNotFound(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Manually create a workoutWOD with a non-existent workout
	workoutWOD := &domain.WorkoutWOD{
		WorkoutID: 999, // Non-existent
		WODID:     1,
	}
	_ = workoutWODRepo.Create(workoutWOD)

	err := svc.RemoveWODFromWorkout(workoutWOD.ID, 1)
	if err == nil {
		t.Error("RemoveWODFromWorkout() should return error when workout not found")
	}
}

func TestWorkoutWODService_UpdateWorkoutWOD_GetByIDError(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	// Set error on GetByID
	workoutWODRepo.getByIDError = errors.New("database error")

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	score := "10:30"
	err := svc.UpdateWorkoutWOD(1, 1, &score, nil)
	if err == nil {
		t.Error("UpdateWorkoutWOD() should return error when GetByID fails")
	}
}

func TestWorkoutWODService_UpdateWorkoutWOD_GetWorkoutError(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Add WOD to workout
	workoutWOD, _ := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, nil)

	// Set error on workout repo
	workoutRepo.getByIDError = errors.New("database error")

	score := "10:30"
	err := svc.UpdateWorkoutWOD(workoutWOD.ID, userID, &score, nil)
	if err == nil {
		t.Error("UpdateWorkoutWOD() should return error when get workout fails")
	}
}

func TestWorkoutWODService_UpdateWorkoutWOD_WorkoutNotFound(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Manually create a workoutWOD with a non-existent workout
	workoutWOD := &domain.WorkoutWOD{
		WorkoutID: 999, // Non-existent
		WODID:     1,
	}
	_ = workoutWODRepo.Create(workoutWOD)

	score := "10:30"
	err := svc.UpdateWorkoutWOD(workoutWOD.ID, 1, &score, nil)
	if err == nil {
		t.Error("UpdateWorkoutWOD() should return error when workout not found")
	}
}

func TestWorkoutWODService_UpdateWorkoutWOD_UpdateError(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Add WOD to workout
	workoutWOD, _ := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, nil)

	// Set error on update
	workoutWODRepo.updateError = errors.New("database error")

	score := "10:30"
	err := svc.UpdateWorkoutWOD(workoutWOD.ID, userID, &score, nil)
	if err == nil {
		t.Error("UpdateWorkoutWOD() should return error when update fails")
	}
}

func TestWorkoutWODService_ToggleWODPR_GetByIDError(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	// Set error on GetByID
	workoutWODRepo.getByIDError = errors.New("database error")

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	err := svc.ToggleWODPR(1, 1)
	if err == nil {
		t.Error("ToggleWODPR() should return error when GetByID fails")
	}
}

func TestWorkoutWODService_ToggleWODPR_GetWorkoutError(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Add WOD to workout
	workoutWOD, _ := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, nil)

	// Set error on workout repo
	workoutRepo.getByIDError = errors.New("database error")

	err := svc.ToggleWODPR(workoutWOD.ID, userID)
	if err == nil {
		t.Error("ToggleWODPR() should return error when get workout fails")
	}
}

func TestWorkoutWODService_ToggleWODPR_WorkoutNotFound(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Manually create a workoutWOD with a non-existent workout
	workoutWOD := &domain.WorkoutWOD{
		WorkoutID: 999, // Non-existent
		WODID:     1,
	}
	_ = workoutWODRepo.Create(workoutWOD)

	err := svc.ToggleWODPR(workoutWOD.ID, 1)
	if err == nil {
		t.Error("ToggleWODPR() should return error when workout not found")
	}
}

func TestWorkoutWODService_ToggleWODPR_ToggleError(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	// Create a workout template owned by user 1
	userID := int64(1)
	workout := &domain.Workout{
		Name:      "Test Workout",
		CreatedBy: &userID,
	}
	_ = workoutRepo.Create(workout)

	// Create a WOD
	wod := &domain.WOD{Name: "Test WOD"}
	_ = wodRepo.Create(wod)

	// Add WOD to workout
	workoutWOD, _ := svc.AddWODToWorkout(workout.ID, wod.ID, userID, 0, nil)

	// Set error on toggle
	workoutWODRepo.togglePRError = errors.New("database error")

	err := svc.ToggleWODPR(workoutWOD.ID, userID)
	if err == nil {
		t.Error("ToggleWODPR() should return error when toggle fails")
	}
}

func TestWorkoutWODService_ListWODsForWorkout_Error(t *testing.T) {
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutRepo := newMockWorkoutRepo()
	wodRepo := newMockWODRepo()

	// Set error on list
	workoutWODRepo.listByWorkoutWithDetailsError = errors.New("database error")

	svc := NewWorkoutWODService(workoutWODRepo, workoutRepo, wodRepo)

	_, err := svc.ListWODsForWorkout(1)
	if err == nil {
		t.Error("ListWODsForWorkout() should return error when list fails")
	}
}
