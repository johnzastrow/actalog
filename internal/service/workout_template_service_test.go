package service

import (
	"database/sql"
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestNewWorkoutTemplateService(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()
	auditLogRepo := newMockAuditLogRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, auditLogRepo)

	if svc == nil {
		t.Fatal("NewWorkoutTemplateService returned nil")
	}
}

func TestWorkoutTemplateService_Create(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()
	auditLogRepo := newMockAuditLogRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, auditLogRepo)

	// Create a template
	movements := []domain.WorkoutMovement{
		{MovementID: 1, Sets: intPtr(3), Reps: intPtr(10), Weight: floatPtr(100.0)},
		{MovementID: 2, Sets: intPtr(4), Reps: intPtr(8), Weight: floatPtr(50.0)},
	}
	wods := []domain.WorkoutWOD{
		{WODID: 1},
	}

	result, err := svc.Create(1, "user@example.com", "Test Template", stringPtr("Test notes"), movements, wods)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if result == nil {
		t.Fatal("Create() returned nil")
	}
	if result.Name != "Test Template" {
		t.Errorf("Name = %s, want Test Template", result.Name)
	}
	if result.CreatedBy == nil || *result.CreatedBy != 1 {
		t.Error("CreatedBy should be set to user ID 1")
	}

	// Verify audit log was created
	if len(auditLogRepo.logs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(auditLogRepo.logs))
	}
	if auditLogRepo.logs[0].EventType != domain.EventWorkoutTemplateCreated {
		t.Errorf("EventType = %s, want %s", auditLogRepo.logs[0].EventType, domain.EventWorkoutTemplateCreated)
	}
}

func TestWorkoutTemplateService_Create_WithoutMovementsOrWODs(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()
	auditLogRepo := newMockAuditLogRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, auditLogRepo)

	// Create a template without movements or WODs
	result, err := svc.Create(1, "user@example.com", "Simple Template", nil, nil, nil)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if result == nil {
		t.Fatal("Create() returned nil")
	}
	if result.Name != "Simple Template" {
		t.Errorf("Name = %s, want Simple Template", result.Name)
	}
}

func TestWorkoutTemplateService_Create_WithoutAuditLogRepo(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Should not panic when audit log repo is nil
	result, err := svc.Create(1, "user@example.com", "No Audit Template", nil, nil, nil)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if result == nil {
		t.Fatal("Create() returned nil")
	}
}

func TestWorkoutTemplateService_GetByID(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()
	auditLogRepo := newMockAuditLogRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, auditLogRepo)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "Test Workout", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Get it back
	result, err := svc.GetByID(workout.ID)
	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}
	if result == nil {
		t.Fatal("GetByID() returned nil")
	}
	if result.Name != "Test Workout" {
		t.Errorf("Name = %s, want Test Workout", result.Name)
	}
}

func TestWorkoutTemplateService_GetByID_NotFound(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.getByIDError = sql.ErrNoRows
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, err := svc.GetByID(999)
	if err == nil {
		t.Error("GetByID() should return error when not found")
	}
}

func TestWorkoutTemplateService_GetByIDWithDetails(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "Detailed Workout", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Get with details
	result, err := svc.GetByIDWithDetails(workout.ID)
	if err != nil {
		t.Errorf("GetByIDWithDetails() error = %v", err)
	}
	if result == nil {
		t.Fatal("GetByIDWithDetails() returned nil")
	}
}

func TestWorkoutTemplateService_ListByUser(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create workouts for different users
	user1 := int64(1)
	user2 := int64(2)
	_ = workoutRepo.Create(&domain.Workout{Name: "User1 Workout 1", CreatedBy: &user1})
	_ = workoutRepo.Create(&domain.Workout{Name: "User1 Workout 2", CreatedBy: &user1})
	_ = workoutRepo.Create(&domain.Workout{Name: "User2 Workout", CreatedBy: &user2})

	// List user1's workouts
	result, err := svc.ListByUser(1, 50, 0)
	if err != nil {
		t.Errorf("ListByUser() error = %v", err)
	}
	if len(result) != 2 {
		t.Errorf("ListByUser() returned %d workouts, want 2", len(result))
	}
}

func TestWorkoutTemplateService_ListStandard(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create standard (CreatedBy = nil) and user-created workouts
	userID := int64(1)
	_ = workoutRepo.Create(&domain.Workout{Name: "Standard Workout 1", CreatedBy: nil})
	_ = workoutRepo.Create(&domain.Workout{Name: "Standard Workout 2", CreatedBy: nil})
	_ = workoutRepo.Create(&domain.Workout{Name: "User Workout", CreatedBy: &userID})

	// List standard workouts
	result, err := svc.ListStandard(50, 0)
	if err != nil {
		t.Errorf("ListStandard() error = %v", err)
	}
	if len(result) != 2 {
		t.Errorf("ListStandard() returned %d workouts, want 2", len(result))
	}
}

func TestWorkoutTemplateService_Update(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()
	auditLogRepo := newMockAuditLogRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, auditLogRepo)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "Original Name", Notes: stringPtr("Original notes"), CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Update it
	movements := []domain.WorkoutMovement{
		{MovementID: 1, Sets: intPtr(5), Reps: intPtr(5)},
	}
	result, err := svc.Update(workout.ID, 1, "user@example.com", "Updated Name", stringPtr("Updated notes"), movements, nil)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}
	if result == nil {
		t.Fatal("Update() returned nil")
	}
	if result.Name != "Updated Name" {
		t.Errorf("Name = %s, want Updated Name", result.Name)
	}

	// Verify audit log
	if len(auditLogRepo.logs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(auditLogRepo.logs))
	}
	if auditLogRepo.logs[0].EventType != domain.EventWorkoutTemplateUpdated {
		t.Errorf("EventType = %s, want %s", auditLogRepo.logs[0].EventType, domain.EventWorkoutTemplateUpdated)
	}
}

func TestWorkoutTemplateService_Update_NotOwner(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a workout owned by user 1
	userID := int64(1)
	workout := &domain.Workout{Name: "User1's Workout", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Try to update as user 2
	_, err := svc.Update(workout.ID, 2, "other@example.com", "Hacked Name", nil, nil, nil)
	if err == nil {
		t.Error("Update() should return error when user doesn't own the template")
	}
}

func TestWorkoutTemplateService_Update_NotFound(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.getByIDError = sql.ErrNoRows
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, err := svc.Update(999, 1, "user@example.com", "Name", nil, nil, nil)
	if err == nil {
		t.Error("Update() should return error when template not found")
	}
}

func TestWorkoutTemplateService_Delete(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()
	auditLogRepo := newMockAuditLogRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, auditLogRepo)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "To Delete", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Delete it
	err := svc.Delete(workout.ID, 1, "user@example.com")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify it's deleted
	if _, exists := workoutRepo.workouts[workout.ID]; exists {
		t.Error("Workout should be deleted")
	}

	// Verify audit log
	if len(auditLogRepo.logs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(auditLogRepo.logs))
	}
	if auditLogRepo.logs[0].EventType != domain.EventWorkoutTemplateDeleted {
		t.Errorf("EventType = %s, want %s", auditLogRepo.logs[0].EventType, domain.EventWorkoutTemplateDeleted)
	}
}

func TestWorkoutTemplateService_Delete_NotOwner(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a workout owned by user 1
	userID := int64(1)
	workout := &domain.Workout{Name: "User1's Workout", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Try to delete as user 2
	err := svc.Delete(workout.ID, 2, "other@example.com")
	if err == nil {
		t.Error("Delete() should return error when user doesn't own the template")
	}
}

func TestWorkoutTemplateService_Delete_NotFound(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.getByIDError = sql.ErrNoRows
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	err := svc.Delete(999, 1, "user@example.com")
	if err == nil {
		t.Error("Delete() should return error when template not found")
	}
}

func TestWorkoutTemplateService_ListAllUserCreated(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create standard and user-created workouts
	user1 := int64(1)
	user2 := int64(2)
	_ = workoutRepo.Create(&domain.Workout{Name: "Standard", CreatedBy: nil})
	_ = workoutRepo.Create(&domain.Workout{Name: "User1 Workout", CreatedBy: &user1})
	_ = workoutRepo.Create(&domain.Workout{Name: "User2 Workout", CreatedBy: &user2})

	// List all user-created (admin function)
	result, count, err := svc.ListAllUserCreated(50, 0)
	if err != nil {
		t.Errorf("ListAllUserCreated() error = %v", err)
	}
	if len(result) != 2 {
		t.Errorf("ListAllUserCreated() returned %d workouts, want 2", len(result))
	}
	if count != 2 {
		t.Errorf("ListAllUserCreated() count = %d, want 2", count)
	}
}

func TestWorkoutTemplateService_ListAllUserCreatedWithUserInfo(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// This returns WorkoutWithCreator structs
	result, count, err := svc.ListAllUserCreatedWithUserInfo(50, 0)
	if err != nil {
		t.Errorf("ListAllUserCreatedWithUserInfo() error = %v", err)
	}
	// Mock returns empty slice
	if result == nil {
		t.Error("ListAllUserCreatedWithUserInfo() returned nil")
	}
	if count < 0 {
		t.Errorf("ListAllUserCreatedWithUserInfo() count = %d, should be >= 0", count)
	}
}

func TestWorkoutTemplateService_ListAllUserCreatedWithUserInfoFiltered(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	result, count, err := svc.ListAllUserCreatedWithUserInfoFiltered(50, 0, "test", "creator")
	if err != nil {
		t.Errorf("ListAllUserCreatedWithUserInfoFiltered() error = %v", err)
	}
	if result == nil {
		t.Error("ListAllUserCreatedWithUserInfoFiltered() returned nil")
	}
	if count < 0 {
		t.Errorf("ListAllUserCreatedWithUserInfoFiltered() count = %d, should be >= 0", count)
	}
}

func TestWorkoutTemplateService_CopyToStandard(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a user workout to copy
	userID := int64(1)
	workout := &domain.Workout{Name: "User's Custom Workout", Notes: stringPtr("Custom notes"), CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Copy to standard
	result, err := svc.CopyToStandard(workout.ID, "Standard Copy")
	if err != nil {
		t.Errorf("CopyToStandard() error = %v", err)
	}
	if result == nil {
		t.Fatal("CopyToStandard() returned nil")
	}
	if result.Name != "Standard Copy" {
		t.Errorf("Name = %s, want Standard Copy", result.Name)
	}
	if result.CreatedBy != nil {
		t.Error("Copied workout should be standard (CreatedBy = nil)")
	}
}

func TestWorkoutTemplateService_CopyToStandard_EmptyName(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a user workout
	userID := int64(1)
	workout := &domain.Workout{Name: "Original", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Try to copy with empty name
	_, err := svc.CopyToStandard(workout.ID, "")
	if err == nil {
		t.Error("CopyToStandard() should return error when name is empty")
	}
}

func TestWorkoutTemplateService_CopyToStandard_NotFound(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := &mockWorkoutMovementRepo{}
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, err := svc.CopyToStandard(999, "New Name")
	if err == nil {
		t.Error("CopyToStandard() should return error when workout not found")
	}
}

// Helper function for float pointers
func floatPtr(f float64) *float64 {
	return &f
}
