package service

import (
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestNewWorkoutService(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	movementRepo := newMockMovementRepo()

	svc := NewWorkoutService(workoutRepo, workoutMovementRepo, workoutWODRepo, movementRepo)

	if svc == nil {
		t.Fatal("NewWorkoutService returned nil")
	}
}

func TestWorkoutService_CreateTemplate(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	movementRepo := newMockMovementRepo()

	svc := NewWorkoutService(workoutRepo, workoutMovementRepo, workoutWODRepo, movementRepo)

	workout := &domain.Workout{
		Name:  "Test Template",
		Notes: stringPtr("Test notes"),
	}

	err := svc.CreateTemplate(1, workout)
	if err != nil {
		t.Errorf("CreateTemplate() error = %v", err)
	}

	// Verify workout was created
	if workout.ID == 0 {
		t.Error("CreateTemplate() did not set workout ID")
	}
	if workout.CreatedBy == nil || *workout.CreatedBy != 1 {
		t.Error("CreateTemplate() did not set CreatedBy")
	}
}

func TestWorkoutService_CreateTemplate_WithMovements(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	movementRepo := newMockMovementRepo()

	svc := NewWorkoutService(workoutRepo, workoutMovementRepo, workoutWODRepo, movementRepo)

	workout := &domain.Workout{
		Name: "Template with Movements",
		Movements: []*domain.WorkoutMovement{
			{MovementID: 1, Sets: intPtr(3), Reps: intPtr(10)},
			{MovementID: 2, Sets: intPtr(4), Reps: intPtr(8)},
		},
	}

	err := svc.CreateTemplate(1, workout)
	if err != nil {
		t.Errorf("CreateTemplate() error = %v", err)
	}

	// Verify movements were created
	if len(workoutMovementRepo.movements) != 2 {
		t.Errorf("CreateTemplate() created %d movements, want 2", len(workoutMovementRepo.movements))
	}
}

func TestWorkoutService_CreateTemplate_WithWODs(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	movementRepo := newMockMovementRepo()

	svc := NewWorkoutService(workoutRepo, workoutMovementRepo, workoutWODRepo, movementRepo)

	workout := &domain.Workout{
		Name: "Template with WODs",
		WODs: []*domain.WorkoutWODWithDetails{
			{WorkoutWOD: domain.WorkoutWOD{WODID: 1}},
			{WorkoutWOD: domain.WorkoutWOD{WODID: 2}},
		},
	}

	err := svc.CreateTemplate(1, workout)
	if err != nil {
		t.Errorf("CreateTemplate() error = %v", err)
	}

	// Verify WODs were created
	if len(workoutWODRepo.workoutWODs) != 2 {
		t.Errorf("CreateTemplate() created %d WODs, want 2", len(workoutWODRepo.workoutWODs))
	}
}

func TestWorkoutService_GetTemplate(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	movementRepo := newMockMovementRepo()

	svc := NewWorkoutService(workoutRepo, workoutMovementRepo, workoutWODRepo, movementRepo)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Get it back
	result, err := svc.GetTemplate(workout.ID)
	if err != nil {
		t.Errorf("GetTemplate() error = %v", err)
	}
	if result == nil {
		t.Fatal("GetTemplate() returned nil")
	}
	if result.Name != "Test Workout" {
		t.Errorf("GetTemplate() Name = %s, want Test Workout", result.Name)
	}
}

func TestWorkoutService_GetTemplate_NotFound(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	svc := NewWorkoutService(workoutRepo, nil, nil, nil)

	_, err := svc.GetTemplate(999)
	if err == nil {
		t.Error("GetTemplate() should return error for non-existent template")
	}
	if err != ErrWorkoutNotFound {
		t.Errorf("GetTemplate() error = %v, want ErrWorkoutNotFound", err)
	}
}

func TestWorkoutService_ListTemplates_StandardOnly(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	svc := NewWorkoutService(workoutRepo, nil, nil, nil)

	// Create standard templates
	_ = workoutRepo.Create(&domain.Workout{Name: "Standard 1", CreatedBy: nil})
	_ = workoutRepo.Create(&domain.Workout{Name: "Standard 2", CreatedBy: nil})

	// Create user template
	userID := int64(1)
	_ = workoutRepo.Create(&domain.Workout{Name: "User Template", CreatedBy: &userID})

	// List without user ID - should only get standard
	result, err := svc.ListTemplates(nil, 50, 0)
	if err != nil {
		t.Errorf("ListTemplates() error = %v", err)
	}
	if len(result) != 2 {
		t.Errorf("ListTemplates() returned %d templates, want 2 (standard only)", len(result))
	}
}

func TestWorkoutService_ListTemplates_WithUser(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	svc := NewWorkoutService(workoutRepo, nil, nil, nil)

	// Create standard templates
	_ = workoutRepo.Create(&domain.Workout{Name: "Standard 1", CreatedBy: nil})

	// Create user templates
	userID := int64(1)
	_ = workoutRepo.Create(&domain.Workout{Name: "User Template 1", CreatedBy: &userID})
	_ = workoutRepo.Create(&domain.Workout{Name: "User Template 2", CreatedBy: &userID})

	// List with user ID - should get standard + user's
	result, err := svc.ListTemplates(&userID, 50, 0)
	if err != nil {
		t.Errorf("ListTemplates() error = %v", err)
	}
	if len(result) != 3 {
		t.Errorf("ListTemplates() returned %d templates, want 3 (1 standard + 2 user)", len(result))
	}
}

func TestWorkoutService_UpdateTemplate(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	svc := NewWorkoutService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "Original Name", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Update it
	updates := &domain.Workout{
		Name:  "Updated Name",
		Notes: stringPtr("Updated notes"),
	}

	err := svc.UpdateTemplate(workout.ID, userID, updates)
	if err != nil {
		t.Errorf("UpdateTemplate() error = %v", err)
	}

	// Verify update
	updated, _ := workoutRepo.GetByID(workout.ID)
	if updated.Name != "Updated Name" {
		t.Errorf("UpdateTemplate() Name = %s, want Updated Name", updated.Name)
	}
}

func TestWorkoutService_UpdateTemplate_NotFound(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	svc := NewWorkoutService(workoutRepo, nil, nil, nil)

	err := svc.UpdateTemplate(999, 1, &domain.Workout{Name: "Test"})
	if err == nil {
		t.Error("UpdateTemplate() should return error for non-existent template")
	}
	if err != ErrWorkoutNotFound {
		t.Errorf("UpdateTemplate() error = %v, want ErrWorkoutNotFound", err)
	}
}

func TestWorkoutService_UpdateTemplate_Unauthorized(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	svc := NewWorkoutService(workoutRepo, nil, nil, nil)

	// Create a workout owned by user 1
	userID := int64(1)
	workout := &domain.Workout{Name: "User 1's Workout", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Try to update as user 2
	err := svc.UpdateTemplate(workout.ID, 2, &domain.Workout{Name: "Hacked"})
	if err == nil {
		t.Error("UpdateTemplate() should return error for unauthorized user")
	}
	if err != ErrUnauthorized {
		t.Errorf("UpdateTemplate() error = %v, want ErrUnauthorized", err)
	}
}

func TestWorkoutService_UpdateTemplate_WithMovements(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	svc := NewWorkoutService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a workout with movements
	userID := int64(1)
	workout := &domain.Workout{Name: "Template", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)
	_ = workoutMovementRepo.Create(&domain.WorkoutMovement{WorkoutID: workout.ID, MovementID: 1})

	// Update with new movements
	updates := &domain.Workout{
		Name: "Updated",
		Movements: []*domain.WorkoutMovement{
			{MovementID: 2},
			{MovementID: 3},
		},
	}

	err := svc.UpdateTemplate(workout.ID, userID, updates)
	if err != nil {
		t.Errorf("UpdateTemplate() error = %v", err)
	}

	// Verify old movements deleted and new ones created
	movements, _ := workoutMovementRepo.GetByWorkoutID(workout.ID)
	if len(movements) != 2 {
		t.Errorf("UpdateTemplate() resulted in %d movements, want 2", len(movements))
	}
}

func TestWorkoutService_DeleteTemplate(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	svc := NewWorkoutService(workoutRepo, nil, nil, nil)

	// Create a workout
	userID := int64(1)
	workout := &domain.Workout{Name: "To Delete", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Delete it
	err := svc.DeleteTemplate(workout.ID, userID)
	if err != nil {
		t.Errorf("DeleteTemplate() error = %v", err)
	}

	// Verify deleted
	if _, exists := workoutRepo.workouts[workout.ID]; exists {
		t.Error("DeleteTemplate() did not delete the workout")
	}
}

func TestWorkoutService_DeleteTemplate_NotFound(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	svc := NewWorkoutService(workoutRepo, nil, nil, nil)

	err := svc.DeleteTemplate(999, 1)
	if err == nil {
		t.Error("DeleteTemplate() should return error for non-existent template")
	}
	if err != ErrWorkoutNotFound {
		t.Errorf("DeleteTemplate() error = %v, want ErrWorkoutNotFound", err)
	}
}

func TestWorkoutService_DeleteTemplate_Unauthorized(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	svc := NewWorkoutService(workoutRepo, nil, nil, nil)

	// Create a workout owned by user 1
	userID := int64(1)
	workout := &domain.Workout{Name: "User 1's Workout", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Try to delete as user 2
	err := svc.DeleteTemplate(workout.ID, 2)
	if err == nil {
		t.Error("DeleteTemplate() should return error for unauthorized user")
	}
	if err != ErrUnauthorized {
		t.Errorf("DeleteTemplate() error = %v, want ErrUnauthorized", err)
	}
}

func TestWorkoutService_GetTemplateUsageStats(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	svc := NewWorkoutService(workoutRepo, nil, nil, nil)

	// Create a workout
	userID := int64(1)
	workout := &domain.Workout{Name: "Test", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Get usage stats
	stats, err := svc.GetTemplateUsageStats(workout.ID)
	if err != nil {
		t.Errorf("GetTemplateUsageStats() error = %v", err)
	}
	if stats == nil {
		t.Fatal("GetTemplateUsageStats() returned nil")
	}
}

func TestWorkoutService_GetTemplateUsageStats_NotFound(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	svc := NewWorkoutService(workoutRepo, nil, nil, nil)

	_, err := svc.GetTemplateUsageStats(999)
	if err == nil {
		t.Error("GetTemplateUsageStats() should return error for non-existent template")
	}
}

func TestWorkoutService_AddMovementToTemplate(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	movementRepo := newMockMovementRepo()
	svc := NewWorkoutService(workoutRepo, workoutMovementRepo, nil, movementRepo)

	// Create a workout
	userID := int64(1)
	workout := &domain.Workout{Name: "Template", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Create a movement
	movement := &domain.Movement{Name: "Back Squat", IsStandard: true}
	_ = movementRepo.Create(movement)

	// Add movement to template
	wm := &domain.WorkoutMovement{Sets: intPtr(3), Reps: intPtr(10)}
	err := svc.AddMovementToTemplate(workout.ID, movement.ID, userID, wm)
	if err != nil {
		t.Errorf("AddMovementToTemplate() error = %v", err)
	}

	// Verify movement was added
	if len(workoutMovementRepo.movements) != 1 {
		t.Errorf("AddMovementToTemplate() created %d movements, want 1", len(workoutMovementRepo.movements))
	}
}

func TestWorkoutService_AddMovementToTemplate_TemplateNotFound(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	svc := NewWorkoutService(workoutRepo, nil, nil, nil)

	err := svc.AddMovementToTemplate(999, 1, 1, &domain.WorkoutMovement{})
	if err == nil {
		t.Error("AddMovementToTemplate() should return error for non-existent template")
	}
}

func TestWorkoutService_AddMovementToTemplate_Unauthorized(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	svc := NewWorkoutService(workoutRepo, nil, nil, nil)

	// Create a workout owned by user 1
	userID := int64(1)
	workout := &domain.Workout{Name: "Template", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Try to add movement as user 2
	err := svc.AddMovementToTemplate(workout.ID, 1, 2, &domain.WorkoutMovement{})
	if err == nil {
		t.Error("AddMovementToTemplate() should return error for unauthorized user")
	}
	if err != ErrUnauthorized {
		t.Errorf("AddMovementToTemplate() error = %v, want ErrUnauthorized", err)
	}
}

func TestWorkoutService_AddMovementToTemplate_MovementNotFound(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	movementRepo := newMockMovementRepo()
	svc := NewWorkoutService(workoutRepo, workoutMovementRepo, nil, movementRepo)

	// Create a workout
	userID := int64(1)
	workout := &domain.Workout{Name: "Template", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Try to add non-existent movement
	err := svc.AddMovementToTemplate(workout.ID, 999, userID, &domain.WorkoutMovement{})
	if err == nil {
		t.Error("AddMovementToTemplate() should return error for non-existent movement")
	}
}

func TestWorkoutService_AddWODToTemplate(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	svc := NewWorkoutService(workoutRepo, nil, workoutWODRepo, nil)

	// Create a workout
	userID := int64(1)
	workout := &domain.Workout{Name: "Template", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Add WOD to template
	wod := &domain.WorkoutWOD{}
	err := svc.AddWODToTemplate(workout.ID, 1, userID, wod)
	if err != nil {
		t.Errorf("AddWODToTemplate() error = %v", err)
	}

	// Verify WOD was added
	if len(workoutWODRepo.workoutWODs) != 1 {
		t.Errorf("AddWODToTemplate() created %d WODs, want 1", len(workoutWODRepo.workoutWODs))
	}
}

func TestWorkoutService_AddWODToTemplate_TemplateNotFound(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	svc := NewWorkoutService(workoutRepo, nil, nil, nil)

	err := svc.AddWODToTemplate(999, 1, 1, &domain.WorkoutWOD{})
	if err == nil {
		t.Error("AddWODToTemplate() should return error for non-existent template")
	}
}

func TestWorkoutService_AddWODToTemplate_Unauthorized(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	svc := NewWorkoutService(workoutRepo, nil, nil, nil)

	// Create a workout owned by user 1
	userID := int64(1)
	workout := &domain.Workout{Name: "Template", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Try to add WOD as user 2
	err := svc.AddWODToTemplate(workout.ID, 1, 2, &domain.WorkoutWOD{})
	if err == nil {
		t.Error("AddWODToTemplate() should return error for unauthorized user")
	}
	if err != ErrUnauthorized {
		t.Errorf("AddWODToTemplate() error = %v, want ErrUnauthorized", err)
	}
}

func TestWorkoutService_ListMovements(t *testing.T) {
	movementRepo := newMockMovementRepo()
	svc := NewWorkoutService(nil, nil, nil, movementRepo)

	// Create standard movements
	_ = movementRepo.Create(&domain.Movement{Name: "Back Squat", IsStandard: true})
	_ = movementRepo.Create(&domain.Movement{Name: "Deadlift", IsStandard: true})

	// Create user movement
	userID := int64(1)
	_ = movementRepo.Create(&domain.Movement{Name: "Custom Movement", CreatedBy: &userID})

	// List movements for user
	result, err := svc.ListMovements(userID)
	if err != nil {
		t.Errorf("ListMovements() error = %v", err)
	}
	if len(result) != 3 {
		t.Errorf("ListMovements() returned %d movements, want 3 (2 standard + 1 user)", len(result))
	}
}

func TestWorkoutService_DetectAndFlagPRs_NewPR(t *testing.T) {
	workoutMovementRepo := newMockWorkoutMovementRepo()
	svc := NewWorkoutService(nil, workoutMovementRepo, nil, nil)

	// Create movement with weight
	weight := 315.0
	movements := []*domain.WorkoutMovement{
		{MovementID: 1, Weight: &weight},
	}

	// No previous max weight - should flag as PR
	err := svc.DetectAndFlagPRs(1, movements)
	if err != nil {
		t.Errorf("DetectAndFlagPRs() error = %v", err)
	}

	if !movements[0].IsPR {
		t.Error("DetectAndFlagPRs() should flag first lift as PR")
	}
}

func TestWorkoutService_DetectAndFlagPRs_ExceedsPrevious(t *testing.T) {
	workoutMovementRepo := newMockWorkoutMovementRepo()
	svc := NewWorkoutService(nil, workoutMovementRepo, nil, nil)

	// Set up previous max weight
	userID := int64(1)
	movementID := int64(1)
	previousMax := 300.0
	workoutMovementRepo.maxWeights[userID] = make(map[int64]*float64)
	workoutMovementRepo.maxWeights[userID][movementID] = &previousMax

	// Create movement with weight exceeding previous max
	newWeight := 315.0
	movements := []*domain.WorkoutMovement{
		{MovementID: movementID, Weight: &newWeight},
	}

	err := svc.DetectAndFlagPRs(userID, movements)
	if err != nil {
		t.Errorf("DetectAndFlagPRs() error = %v", err)
	}

	if !movements[0].IsPR {
		t.Error("DetectAndFlagPRs() should flag lift that exceeds previous max as PR")
	}
}

func TestWorkoutService_DetectAndFlagPRs_NotPR(t *testing.T) {
	workoutMovementRepo := newMockWorkoutMovementRepo()
	svc := NewWorkoutService(nil, workoutMovementRepo, nil, nil)

	// Set up previous max weight
	userID := int64(1)
	movementID := int64(1)
	previousMax := 315.0
	workoutMovementRepo.maxWeights[userID] = make(map[int64]*float64)
	workoutMovementRepo.maxWeights[userID][movementID] = &previousMax

	// Create movement with weight below previous max
	newWeight := 300.0
	movements := []*domain.WorkoutMovement{
		{MovementID: movementID, Weight: &newWeight},
	}

	err := svc.DetectAndFlagPRs(userID, movements)
	if err != nil {
		t.Errorf("DetectAndFlagPRs() error = %v", err)
	}

	if movements[0].IsPR {
		t.Error("DetectAndFlagPRs() should not flag lift below previous max as PR")
	}
}

func TestWorkoutService_DetectAndFlagPRs_NoWeight(t *testing.T) {
	workoutMovementRepo := newMockWorkoutMovementRepo()
	svc := NewWorkoutService(nil, workoutMovementRepo, nil, nil)

	// Create movement without weight
	movements := []*domain.WorkoutMovement{
		{MovementID: 1, Weight: nil},
	}

	err := svc.DetectAndFlagPRs(1, movements)
	if err != nil {
		t.Errorf("DetectAndFlagPRs() error = %v", err)
	}

	if movements[0].IsPR {
		t.Error("DetectAndFlagPRs() should not flag movement without weight as PR")
	}
}

func TestWorkoutService_GetPersonalRecords(t *testing.T) {
	workoutMovementRepo := newMockWorkoutMovementRepo()
	svc := NewWorkoutService(nil, workoutMovementRepo, nil, nil)

	records, err := svc.GetPersonalRecords(1)
	if err != nil {
		t.Errorf("GetPersonalRecords() error = %v", err)
	}
	if records == nil {
		t.Error("GetPersonalRecords() returned nil")
	}
}

func TestWorkoutService_GetPRMovements(t *testing.T) {
	workoutMovementRepo := newMockWorkoutMovementRepo()
	svc := NewWorkoutService(nil, workoutMovementRepo, nil, nil)

	// Create some movements with PR flags
	weight := 315.0
	_ = workoutMovementRepo.Create(&domain.WorkoutMovement{MovementID: 1, Weight: &weight, IsPR: true})
	_ = workoutMovementRepo.Create(&domain.WorkoutMovement{MovementID: 2, Weight: &weight, IsPR: false})

	movements, err := svc.GetPRMovements(1, 10)
	if err != nil {
		t.Errorf("GetPRMovements() error = %v", err)
	}
	if len(movements) != 1 {
		t.Errorf("GetPRMovements() returned %d movements, want 1 (PR only)", len(movements))
	}
}

func TestWorkoutService_TogglePRFlag(t *testing.T) {
	workoutMovementRepo := newMockWorkoutMovementRepo()
	svc := NewWorkoutService(nil, workoutMovementRepo, nil, nil)

	// Create a movement
	weight := 315.0
	wm := &domain.WorkoutMovement{MovementID: 1, Weight: &weight, IsPR: false}
	_ = workoutMovementRepo.Create(wm)

	// Toggle PR flag
	err := svc.TogglePRFlag(wm.ID, 1)
	if err != nil {
		t.Errorf("TogglePRFlag() error = %v", err)
	}

	// Verify toggled
	updated, _ := workoutMovementRepo.GetByID(wm.ID)
	if !updated.IsPR {
		t.Error("TogglePRFlag() did not toggle IsPR to true")
	}

	// Toggle again
	err = svc.TogglePRFlag(wm.ID, 1)
	if err != nil {
		t.Errorf("TogglePRFlag() error = %v", err)
	}

	// Verify toggled back
	updated, _ = workoutMovementRepo.GetByID(wm.ID)
	if updated.IsPR {
		t.Error("TogglePRFlag() did not toggle IsPR back to false")
	}
}

func TestWorkoutService_TogglePRFlag_NotFound(t *testing.T) {
	workoutMovementRepo := newMockWorkoutMovementRepo()
	svc := NewWorkoutService(nil, workoutMovementRepo, nil, nil)

	err := svc.TogglePRFlag(999, 1)
	if err == nil {
		t.Error("TogglePRFlag() should return error for non-existent movement")
	}
}
