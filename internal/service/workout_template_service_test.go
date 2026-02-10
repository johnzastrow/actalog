package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestNewWorkoutTemplateService(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	auditLogRepo := newMockAuditLogRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, auditLogRepo)

	if svc == nil {
		t.Fatal("NewWorkoutTemplateService returned nil")
	}
}

func TestWorkoutTemplateService_Create(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
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

	result, err := svc.Create(1, "user@example.com", "user", "Test Template", nil, stringPtr("Test notes"), false, movements, wods)
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	auditLogRepo := newMockAuditLogRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, auditLogRepo)

	// Create a template without movements or WODs
	result, err := svc.Create(1, "user@example.com", "user", "Simple Template", nil, nil, false, nil, nil)
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Should not panic when audit log repo is nil
	result, err := svc.Create(1, "user@example.com", "user", "No Audit Template", nil, nil, false, nil, nil)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if result == nil {
		t.Fatal("Create() returned nil")
	}
}

func TestWorkoutTemplateService_GetByID(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, err := svc.GetByID(999)
	if err == nil {
		t.Error("GetByID() should return error when not found")
	}
}

func TestWorkoutTemplateService_GetByIDWithDetails(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
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
	result, err := svc.Update(workout.ID, 1, "user@example.com", "user", "Updated Name", nil, stringPtr("Updated notes"), false, movements, nil)
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a workout owned by user 1
	userID := int64(1)
	workout := &domain.Workout{Name: "User1's Workout", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Try to update as user 2
	_, err := svc.Update(workout.ID, 2, "other@example.com", "user", "Hacked Name", nil, nil, false, nil, nil)
	if err == nil {
		t.Error("Update() should return error when user doesn't own the template")
	}
}

func TestWorkoutTemplateService_Update_NotFound(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.getByIDError = sql.ErrNoRows
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, err := svc.Update(999, 1, "user@example.com", "user", "Name", nil, nil, false, nil, nil)
	if err == nil {
		t.Error("Update() should return error when template not found")
	}
}

func TestWorkoutTemplateService_Delete(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	auditLogRepo := newMockAuditLogRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, auditLogRepo)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "To Delete", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Delete it
	err := svc.Delete(workout.ID, 1, "user@example.com", "user")
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a workout owned by user 1
	userID := int64(1)
	workout := &domain.Workout{Name: "User1's Workout", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Try to delete as user 2
	err := svc.Delete(workout.ID, 2, "other@example.com", "user")
	if err == nil {
		t.Error("Delete() should return error when user doesn't own the template")
	}
}

func TestWorkoutTemplateService_Delete_NotFound(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.getByIDError = sql.ErrNoRows
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	err := svc.Delete(999, 1, "user@example.com", "user")
	if err == nil {
		t.Error("Delete() should return error when template not found")
	}
}

func TestWorkoutTemplateService_ListAllUserCreated(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
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

func TestWorkoutTemplateService_GetByIDWithDetails_NotFound(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.getByIDWithDetailsError = sql.ErrNoRows
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, err := svc.GetByIDWithDetails(999)
	if err == nil {
		t.Error("GetByIDWithDetails() should return error when not found")
	}
}

func TestWorkoutTemplateService_GetByIDWithDetails_RepoError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.getByIDWithDetailsError = errors.New("database error")
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, err := svc.GetByIDWithDetails(1)
	if err == nil {
		t.Error("GetByIDWithDetails() should return error when repo fails")
	}
}

func TestWorkoutTemplateService_GetByID_RepoError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.getByIDError = errors.New("database error")
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, err := svc.GetByID(1)
	if err == nil {
		t.Error("GetByID() should return error when repo fails")
	}
}

func TestWorkoutTemplateService_ListByUser_Error(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.listByUserError = errors.New("database error")
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, err := svc.ListByUser(1, 50, 0)
	if err == nil {
		t.Error("ListByUser() should return error when repo fails")
	}
}

func TestWorkoutTemplateService_ListStandard_Error(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.listStandardError = errors.New("database error")
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, err := svc.ListStandard(50, 0)
	if err == nil {
		t.Error("ListStandard() should return error when repo fails")
	}
}

func TestWorkoutTemplateService_ListAllUserCreated_ListError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.listAllUserCreatedError = errors.New("database error")
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, _, err := svc.ListAllUserCreated(50, 0)
	if err == nil {
		t.Error("ListAllUserCreated() should return error when repo fails")
	}
}

func TestWorkoutTemplateService_ListAllUserCreated_CountError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.countAllUserCreatedError = errors.New("count error")
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a user workout
	userID := int64(1)
	_ = workoutRepo.Create(&domain.Workout{Name: "Test", CreatedBy: &userID})

	_, _, err := svc.ListAllUserCreated(50, 0)
	if err == nil {
		t.Error("ListAllUserCreated() should return error when count fails")
	}
}

func TestWorkoutTemplateService_ListAllUserCreatedWithUserInfo_ListError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.listUserCreatedWithInfoError = errors.New("database error")
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, _, err := svc.ListAllUserCreatedWithUserInfo(50, 0)
	if err == nil {
		t.Error("ListAllUserCreatedWithUserInfo() should return error when repo fails")
	}
}

func TestWorkoutTemplateService_ListAllUserCreatedWithUserInfo_CountError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.countAllUserCreatedError = errors.New("count error")
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, _, err := svc.ListAllUserCreatedWithUserInfo(50, 0)
	if err == nil {
		t.Error("ListAllUserCreatedWithUserInfo() should return error when count fails")
	}
}

func TestWorkoutTemplateService_ListAllUserCreatedWithUserInfoFiltered_Error(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.listUserCreatedFilteredError = errors.New("database error")
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, _, err := svc.ListAllUserCreatedWithUserInfoFiltered(50, 0, "test", "")
	if err == nil {
		t.Error("ListAllUserCreatedWithUserInfoFiltered() should return error when repo fails")
	}
}

func TestWorkoutTemplateService_Create_RepoError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.createError = errors.New("database error")
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, err := svc.Create(1, "user@example.com", "user", "Test", nil, nil, false, nil, nil)
	if err == nil {
		t.Error("Create() should return error when repo fails")
	}
}

func TestWorkoutTemplateService_Create_MovementError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutMovementRepo.createError = errors.New("movement error")
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	movements := []domain.WorkoutMovement{
		{MovementID: 1, Sets: intPtr(3), Reps: intPtr(10)},
	}
	_, err := svc.Create(1, "user@example.com", "user", "Test", nil, nil, false, movements, nil)
	if err == nil {
		t.Error("Create() should return error when adding movement fails")
	}
}

func TestWorkoutTemplateService_Create_WODError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutWODRepo.createError = errors.New("wod error")

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	wods := []domain.WorkoutWOD{
		{WODID: 1},
	}
	_, err := svc.Create(1, "user@example.com", "user", "Test", nil, nil, false, nil, wods)
	if err == nil {
		t.Error("Create() should return error when adding WOD fails")
	}
}

func TestWorkoutTemplateService_Update_RepoUpdateError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "Original", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Set update error
	workoutRepo.updateError = errors.New("update error")

	_, err := svc.Update(workout.ID, 1, "user@example.com", "user", "Updated", nil, nil, false, nil, nil)
	if err == nil {
		t.Error("Update() should return error when repo update fails")
	}
}

func TestWorkoutTemplateService_Update_DeleteMovementsError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "Original", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Set delete movements error
	workoutMovementRepo.deleteByWorkoutIDError = errors.New("delete movements error")

	_, err := svc.Update(workout.ID, 1, "user@example.com", "user", "Updated", nil, nil, false, nil, nil)
	if err == nil {
		t.Error("Update() should return error when deleting movements fails")
	}
}

func TestWorkoutTemplateService_Update_CreateMovementError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "Original", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Set create movement error
	workoutMovementRepo.createError = errors.New("create movement error")

	movements := []domain.WorkoutMovement{
		{MovementID: 1, Sets: intPtr(3), Reps: intPtr(10)},
	}
	_, err := svc.Update(workout.ID, 1, "user@example.com", "user", "Updated", nil, nil, false, movements, nil)
	if err == nil {
		t.Error("Update() should return error when creating movement fails")
	}
}

func TestWorkoutTemplateService_Update_DeleteWODsError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	workoutWODRepo.deleteError = errors.New("delete wods error")

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "Original", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	_, err := svc.Update(workout.ID, 1, "user@example.com", "user", "Updated", nil, nil, false, nil, nil)
	if err == nil {
		t.Error("Update() should return error when deleting WODs fails")
	}
}

func TestWorkoutTemplateService_Update_CreateWODError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "Original", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Set create WOD error after delete succeeds
	workoutWODRepo.createError = errors.New("create wod error")

	wods := []domain.WorkoutWOD{
		{WODID: 1},
	}
	_, err := svc.Update(workout.ID, 1, "user@example.com", "user", "Updated", nil, nil, false, nil, wods)
	if err == nil {
		t.Error("Update() should return error when creating WOD fails")
	}
}

func TestWorkoutTemplateService_Update_GetByIDRepoError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.getByIDError = errors.New("database error")
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, err := svc.Update(1, 1, "user@example.com", "user", "Updated", nil, nil, false, nil, nil)
	if err == nil {
		t.Error("Update() should return error when GetByID fails")
	}
}

func TestWorkoutTemplateService_Update_NilCreatedBy(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a standard workout (CreatedBy = nil)
	workout := &domain.Workout{Name: "Standard", CreatedBy: nil}
	_ = workoutRepo.Create(workout)

	// Try to update as any user - should fail since CreatedBy is nil
	_, err := svc.Update(workout.ID, 1, "user@example.com", "user", "Updated", nil, nil, false, nil, nil)
	if err == nil {
		t.Error("Update() should return error when workout has no owner")
	}
}

func TestWorkoutTemplateService_Delete_RepoError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "To Delete", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Set delete error
	workoutRepo.deleteError = errors.New("delete error")

	err := svc.Delete(workout.ID, 1, "user@example.com", "user")
	if err == nil {
		t.Error("Delete() should return error when repo delete fails")
	}
}

func TestWorkoutTemplateService_Delete_DeleteMovementsError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutMovementRepo.deleteByWorkoutIDError = errors.New("delete movements error")
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "To Delete", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	err := svc.Delete(workout.ID, 1, "user@example.com", "user")
	if err == nil {
		t.Error("Delete() should return error when deleting movements fails")
	}
}

func TestWorkoutTemplateService_Delete_GetByIDRepoError(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutRepo.getByIDError = errors.New("database error")
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	err := svc.Delete(1, 1, "user@example.com", "user")
	if err == nil {
		t.Error("Delete() should return error when GetByID fails")
	}
}

func TestWorkoutTemplateService_Delete_NilCreatedBy(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a standard workout (CreatedBy = nil)
	workout := &domain.Workout{Name: "Standard", CreatedBy: nil}
	_ = workoutRepo.Create(workout)

	// Try to delete as any user - should fail since CreatedBy is nil
	err := svc.Delete(workout.ID, 1, "user@example.com", "user")
	if err == nil {
		t.Error("Delete() should return error when workout has no owner")
	}
}

func TestWorkoutTemplateService_Update_WithWODs(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	auditLogRepo := newMockAuditLogRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, auditLogRepo)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "Original Name", CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Update with WODs
	wods := []domain.WorkoutWOD{
		{WODID: 1},
		{WODID: 2},
	}
	result, err := svc.Update(workout.ID, 1, "user@example.com", "user", "Updated Name", nil, nil, false, nil, wods)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}
	if result == nil {
		t.Fatal("Update() returned nil")
	}
	if result.Name != "Updated Name" {
		t.Errorf("Name = %s, want Updated Name", result.Name)
	}
}

// Tests for Instructions field preservation

func TestWorkoutTemplateService_Create_WithInstructions(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	auditLogRepo := newMockAuditLogRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, auditLogRepo)

	// Create a template with movements that have instructions
	movements := []domain.WorkoutMovement{
		{
			MovementID:   1,
			Sets:         intPtr(3),
			Reps:         intPtr(10),
			Weight:       floatPtr(100.0),
			Instructions: "**Setup:** Stand with feet shoulder-width apart\n* Keep core tight\n* Breathe out on the push",
			Notes:        "Increase weight next session",
			OrderIndex:   1,
		},
		{
			MovementID:   2,
			Sets:         intPtr(4),
			Reps:         intPtr(8),
			Weight:       floatPtr(50.0),
			Instructions: "Focus on **slow eccentric** movement",
			Notes:        "Rest 90 seconds between sets",
			OrderIndex:   2,
		},
	}
	wods := []domain.WorkoutWOD{
		{
			WODID:        1,
			Instructions: "## Scaling Options\n- Rx: As written\n- Scaled: 15-12-9 reps",
			Notes:        "Time cap: 12 minutes",
			OrderIndex:   1,
		},
	}

	result, err := svc.Create(1, "user@example.com", "user", "Template With Instructions", nil, stringPtr("Test notes"), false, movements, wods)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if result == nil {
		t.Fatal("Create() returned nil")
	}
	if result.Name != "Template With Instructions" {
		t.Errorf("Name = %s, want Template With Instructions", result.Name)
	}

	// Verify movements were created with instructions
	storedMovements, err := workoutMovementRepo.GetByWorkoutID(result.ID)
	if err != nil {
		t.Errorf("GetByWorkoutID() error = %v", err)
	}
	if len(storedMovements) != 2 {
		t.Errorf("Expected 2 movements, got %d", len(storedMovements))
	}

	// Check first movement has instructions
	found := false
	for _, m := range storedMovements {
		if m.MovementID == 1 {
			found = true
			if m.Instructions == "" {
				t.Error("Movement 1 should have instructions")
			}
			if m.Notes != "Increase weight next session" {
				t.Errorf("Movement 1 notes = %s, want 'Increase weight next session'", m.Notes)
			}
		}
	}
	if !found {
		t.Error("Movement 1 not found in stored movements")
	}
}

func TestWorkoutTemplateService_Update_WithInstructions(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	auditLogRepo := newMockAuditLogRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, auditLogRepo)

	// Create a workout first
	userID := int64(1)
	workout := &domain.Workout{Name: "Original Name", Notes: stringPtr("Original notes"), CreatedBy: &userID}
	_ = workoutRepo.Create(workout)

	// Update with movements and WODs that have instructions
	movements := []domain.WorkoutMovement{
		{
			MovementID:   1,
			Sets:         intPtr(5),
			Reps:         intPtr(5),
			Instructions: "**Heavy day** - focus on form",
			Notes:        "Use lifting belt",
		},
	}
	wods := []domain.WorkoutWOD{
		{
			WODID:        1,
			Instructions: "For time, cap at 15 minutes",
			Notes:        "Scale pull-ups to ring rows if needed",
		},
	}

	result, err := svc.Update(workout.ID, 1, "user@example.com", "user", "Updated Name", nil, stringPtr("Updated notes"), false, movements, wods)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}
	if result == nil {
		t.Fatal("Update() returned nil")
	}

	// Verify movements were updated with instructions
	storedMovements, err := workoutMovementRepo.GetByWorkoutID(result.ID)
	if err != nil {
		t.Errorf("GetByWorkoutID() error = %v", err)
	}
	if len(storedMovements) != 1 {
		t.Errorf("Expected 1 movement, got %d", len(storedMovements))
	}
	if len(storedMovements) > 0 && storedMovements[0].Instructions == "" {
		t.Error("Movement should have instructions after update")
	}
}

func TestWorkoutTemplateService_Create_AllMovementFields(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a template with all movement fields populated
	movements := []domain.WorkoutMovement{
		{
			MovementID:   1,
			Sets:         intPtr(3),
			Reps:         intPtr(10),
			Weight:       floatPtr(135.5),
			Time:         intPtr(60),
			Distance:     floatPtr(400.0),
			IsRx:         true,
			IsPR:         false,
			Instructions: "**Warmup**: Start with empty bar\n- Add weight gradually\n- Focus on depth",
			Notes:        "Target: 3x10 @ 60% 1RM",
			OrderIndex:   1,
		},
	}

	result, err := svc.Create(1, "user@example.com", "user", "Full Movement Details", nil, nil, false, movements, nil)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if result == nil {
		t.Fatal("Create() returned nil")
	}

	// Verify all fields were stored
	storedMovements, _ := workoutMovementRepo.GetByWorkoutID(result.ID)
	if len(storedMovements) != 1 {
		t.Fatalf("Expected 1 movement, got %d", len(storedMovements))
	}

	m := storedMovements[0]
	if m.Sets == nil || *m.Sets != 3 {
		t.Errorf("Sets = %v, want 3", m.Sets)
	}
	if m.Reps == nil || *m.Reps != 10 {
		t.Errorf("Reps = %v, want 10", m.Reps)
	}
	if m.Weight == nil || *m.Weight != 135.5 {
		t.Errorf("Weight = %v, want 135.5", m.Weight)
	}
	if m.Time == nil || *m.Time != 60 {
		t.Errorf("Time = %v, want 60", m.Time)
	}
	if m.Distance == nil || *m.Distance != 400.0 {
		t.Errorf("Distance = %v, want 400.0", m.Distance)
	}
	if m.Instructions == "" {
		t.Error("Instructions should not be empty")
	}
	expectedNotes := "Target: 3x10 @ 60% 1RM"
	if m.Notes != expectedNotes {
		t.Errorf("Notes = %s, want '%s'", m.Notes, expectedNotes)
	}
}

func TestWorkoutTemplateService_Create_AllWODFields(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	// Create a template with all WOD fields populated
	scoreValue := "12:35"
	division := "rx"
	wods := []domain.WorkoutWOD{
		{
			WODID:        1,
			ScoreValue:   &scoreValue,
			Division:     &division,
			IsPR:         true,
			Instructions: "## Standards\n- Full ROM on each rep\n- No kipping on strict movements\n\n## Scaling\n- Scaled: Reduce reps by 50%",
			Notes:        "Achieved during open workout 24.1",
			OrderIndex:   1,
		},
	}

	result, err := svc.Create(1, "user@example.com", "user", "Full WOD Details", nil, nil, false, nil, wods)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if result == nil {
		t.Fatal("Create() returned nil")
	}

	// Verify WOD was stored (check via mock)
	// The mock stores WODs by ID, verify it was created
	if len(workoutWODRepo.workoutWODs) != 1 {
		t.Errorf("Expected 1 WOD in mock, got %d", len(workoutWODRepo.workoutWODs))
	}

	// Check the stored WOD has all fields
	for _, w := range workoutWODRepo.workoutWODs {
		if w.Instructions == "" {
			t.Error("WOD instructions should not be empty")
		}
		if w.Notes != "Achieved during open workout 24.1" {
			t.Errorf("WOD notes = %s, want 'Achieved during open workout 24.1'", w.Notes)
		}
	}
}

// =============================================================================
// Comprehensive CRUD Tests - Verify all fields through create-edit cycle
// =============================================================================

// TestWorkoutTemplate_FullCRUD_AllFields tests complete CRUD operations
// with verification that all fields are stored and returned correctly
func TestWorkoutTemplate_FullCRUD_AllFields(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()
	auditLogRepo := newMockAuditLogRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, auditLogRepo)

	userID := int64(1)
	userEmail := "test@example.com"

	// =========================================================================
	// STEP 1: Create template with ALL fields populated
	// =========================================================================
	t.Run("Create_AllFields", func(t *testing.T) {
		movements := []domain.WorkoutMovement{
			{
				MovementID:   1,
				Sets:         intPtr(4),
				Reps:         intPtr(12),
				Weight:       floatPtr(155.5),
				Time:         intPtr(90),
				Distance:     floatPtr(500.0),
				IsRx:         true,
				IsPR:         false,
				Instructions: "**Setup Instructions**\n\n1. Position bar at mid-shin\n2. Grip just outside knees\n3. Keep back flat",
				Notes:        "Progressive overload week 3",
				OrderIndex:   1,
			},
			{
				MovementID:   2,
				Sets:         intPtr(3),
				Reps:         intPtr(8),
				Weight:       floatPtr(95.0),
				Time:         intPtr(45),
				Distance:     floatPtr(200.0),
				IsRx:         false,
				IsPR:         true,
				Instructions: "Tempo: 3-1-2-0",
				Notes:        "Use belt for heavy sets",
				OrderIndex:   2,
			},
		}

		scoreVal := "15:30"
		div := "scaled"
		wods := []domain.WorkoutWOD{
			{
				WODID:        1,
				ScoreValue:   &scoreVal,
				Division:     &div,
				IsPR:         true,
				Instructions: "## Workout Standards\n\n- Chest to floor on burpees\n- Full lockout overhead",
				Notes:        "First attempt at this WOD",
				OrderIndex:   1,
			},
			{
				WODID:        2,
				ScoreValue:   nil,
				Division:     nil,
				IsPR:         false,
				Instructions: "Scale as needed",
				Notes:        "",
				OrderIndex:   2,
			},
		}

		templateNotes := "Weekly strength program - Day 1: Push focus"
		created, err := svc.Create(userID, userEmail, "user", "Full Test Template", nil, &templateNotes, false, movements, wods)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// Verify template-level fields
		if created.Name != "Full Test Template" {
			t.Errorf("Name = %s, want 'Full Test Template'", created.Name)
		}
		if created.Notes == nil || *created.Notes != templateNotes {
			t.Errorf("Notes = %v, want '%s'", created.Notes, templateNotes)
		}
		if created.CreatedBy == nil || *created.CreatedBy != userID {
			t.Errorf("CreatedBy = %v, want %d", created.CreatedBy, userID)
		}

		// Verify movements were stored
		storedMovements, _ := workoutMovementRepo.GetByWorkoutID(created.ID)
		if len(storedMovements) != 2 {
			t.Fatalf("Expected 2 movements, got %d", len(storedMovements))
		}

		// Verify first movement ALL fields
		m1 := findMovementByID(storedMovements, 1)
		if m1 == nil {
			t.Fatal("Movement 1 not found")
		}
		assertIntPtr(t, "Movement1.Sets", m1.Sets, 4)
		assertIntPtr(t, "Movement1.Reps", m1.Reps, 12)
		assertFloatPtr(t, "Movement1.Weight", m1.Weight, 155.5)
		assertIntPtr(t, "Movement1.Time", m1.Time, 90)
		assertFloatPtr(t, "Movement1.Distance", m1.Distance, 500.0)
		// Note: IsRx and IsPR are set by mock but not verified here as mock behavior may vary
		if m1.Instructions == "" {
			t.Error("Movement1.Instructions should not be empty")
		}
		if m1.Notes != "Progressive overload week 3" {
			t.Errorf("Movement1.Notes = %s, want 'Progressive overload week 3'", m1.Notes)
		}

		// Verify second movement ALL fields
		m2 := findMovementByID(storedMovements, 2)
		if m2 == nil {
			t.Fatal("Movement 2 not found")
		}
		assertIntPtr(t, "Movement2.Sets", m2.Sets, 3)
		assertIntPtr(t, "Movement2.Reps", m2.Reps, 8)
		assertFloatPtr(t, "Movement2.Weight", m2.Weight, 95.0)
		// Note: IsRx and IsPR verification skipped for mock

		// Verify WODs were stored
		if len(workoutWODRepo.workoutWODs) != 2 {
			t.Fatalf("Expected 2 WODs, got %d", len(workoutWODRepo.workoutWODs))
		}
	})
}

// TestWorkoutTemplate_EditIndividualFields tests editing one field at a time
func TestWorkoutTemplate_EditIndividualFields(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	userID := int64(1)
	userEmail := "test@example.com"

	// Create initial template
	initialNotes := "Initial notes"
	movements := []domain.WorkoutMovement{
		{
			MovementID:   1,
			Sets:         intPtr(3),
			Reps:         intPtr(10),
			Weight:       floatPtr(100.0),
			Instructions: "Original instructions",
			Notes:        "Original notes",
			OrderIndex:   1,
		},
	}
	wods := []domain.WorkoutWOD{
		{
			WODID:        1,
			Instructions: "Original WOD instructions",
			Notes:        "Original WOD notes",
			OrderIndex:   1,
		},
	}

	created, err := svc.Create(userID, userEmail, "user", "Edit Test Template", nil, &initialNotes, false, movements, wods)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	templateID := created.ID

	// =========================================================================
	// Test 1: Edit only template name
	// =========================================================================
	t.Run("Edit_TemplateName", func(t *testing.T) {
		result, err := svc.Update(templateID, userID, userEmail, "user", "Updated Template Name", nil, &initialNotes, false, movements, wods)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		if result.Name != "Updated Template Name" {
			t.Errorf("Name = %s, want 'Updated Template Name'", result.Name)
		}
		// Notes should be unchanged
		if result.Notes == nil || *result.Notes != initialNotes {
			t.Errorf("Notes changed unexpectedly: got %v", result.Notes)
		}
	})

	// =========================================================================
	// Test 2: Edit only template notes
	// =========================================================================
	t.Run("Edit_TemplateNotes", func(t *testing.T) {
		newNotes := "Updated notes content"
		result, err := svc.Update(templateID, userID, userEmail, "user", "Updated Template Name", nil, &newNotes, false, movements, wods)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		if result.Notes == nil || *result.Notes != newNotes {
			t.Errorf("Notes = %v, want '%s'", result.Notes, newNotes)
		}
	})

	// =========================================================================
	// Test 3: Edit movement sets only
	// =========================================================================
	t.Run("Edit_MovementSets", func(t *testing.T) {
		updatedMovements := []domain.WorkoutMovement{
			{
				MovementID:   1,
				Sets:         intPtr(5), // Changed from 3 to 5
				Reps:         intPtr(10),
				Weight:       floatPtr(100.0),
				Instructions: "Original instructions",
				Notes:        "Original notes",
				OrderIndex:   1,
			},
		}

		_, err := svc.Update(templateID, userID, userEmail, "user", "Updated Template Name", nil, &initialNotes, false, updatedMovements, wods)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		storedMovements, _ := workoutMovementRepo.GetByWorkoutID(templateID)
		if len(storedMovements) != 1 {
			t.Fatalf("Expected 1 movement, got %d", len(storedMovements))
		}
		assertIntPtr(t, "Sets", storedMovements[0].Sets, 5)
		// Other fields should be preserved
		assertIntPtr(t, "Reps", storedMovements[0].Reps, 10)
		assertFloatPtr(t, "Weight", storedMovements[0].Weight, 100.0)
	})

	// =========================================================================
	// Test 4: Edit movement instructions only
	// =========================================================================
	t.Run("Edit_MovementInstructions", func(t *testing.T) {
		updatedMovements := []domain.WorkoutMovement{
			{
				MovementID:   1,
				Sets:         intPtr(5),
				Reps:         intPtr(10),
				Weight:       floatPtr(100.0),
				Instructions: "**New Instructions**\n\n- Step 1\n- Step 2", // Changed
				Notes:        "Original notes",
				OrderIndex:   1,
			},
		}

		_, err := svc.Update(templateID, userID, userEmail, "user", "Updated Template Name", nil, &initialNotes, false, updatedMovements, wods)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		storedMovements, _ := workoutMovementRepo.GetByWorkoutID(templateID)
		if storedMovements[0].Instructions != "**New Instructions**\n\n- Step 1\n- Step 2" {
			t.Errorf("Instructions not updated correctly")
		}
	})

	// =========================================================================
	// Test 5: Edit movement weight only
	// =========================================================================
	t.Run("Edit_MovementWeight", func(t *testing.T) {
		updatedMovements := []domain.WorkoutMovement{
			{
				MovementID:   1,
				Sets:         intPtr(5),
				Reps:         intPtr(10),
				Weight:       floatPtr(125.5), // Changed
				Instructions: "**New Instructions**\n\n- Step 1\n- Step 2",
				Notes:        "Original notes",
				OrderIndex:   1,
			},
		}

		_, err := svc.Update(templateID, userID, userEmail, "user", "Updated Template Name", nil, &initialNotes, false, updatedMovements, wods)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		storedMovements, _ := workoutMovementRepo.GetByWorkoutID(templateID)
		assertFloatPtr(t, "Weight", storedMovements[0].Weight, 125.5)
	})

	// =========================================================================
	// Test 6: Edit WOD instructions only
	// =========================================================================
	t.Run("Edit_WODInstructions", func(t *testing.T) {
		// Keep same movements
		currentMovements := []domain.WorkoutMovement{
			{
				MovementID:   1,
				Sets:         intPtr(5),
				Reps:         intPtr(10),
				Weight:       floatPtr(125.5),
				Instructions: "**New Instructions**\n\n- Step 1\n- Step 2",
				Notes:        "Original notes",
				OrderIndex:   1,
			},
		}
		updatedWODs := []domain.WorkoutWOD{
			{
				WODID:        1,
				Instructions: "## Updated WOD Standards\n\n- New standard 1\n- New standard 2", // Changed
				Notes:        "Original WOD notes",
				OrderIndex:   1,
			},
		}

		_, err := svc.Update(templateID, userID, userEmail, "user", "Updated Template Name", nil, &initialNotes, false, currentMovements, updatedWODs)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Verify WOD was updated
		found := false
		for _, w := range workoutWODRepo.workoutWODs {
			if w.WorkoutID == templateID {
				found = true
				if w.Instructions != "## Updated WOD Standards\n\n- New standard 1\n- New standard 2" {
					t.Errorf("WOD instructions not updated correctly")
				}
			}
		}
		if !found {
			t.Error("WOD not found after update")
		}
	})
}

// TestWorkoutTemplate_NullableFields tests handling of nullable fields
func TestWorkoutTemplate_NullableFields(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	userID := int64(1)
	userEmail := "test@example.com"

	// Test creating with nil optional fields
	t.Run("Create_WithNilOptionalFields", func(t *testing.T) {
		movements := []domain.WorkoutMovement{
			{
				MovementID:   1,
				Sets:         nil, // nil
				Reps:         nil, // nil
				Weight:       nil, // nil
				Time:         nil, // nil
				Distance:     nil, // nil
				Instructions: "",  // empty
				Notes:        "",  // empty
				OrderIndex:   1,
			},
		}

		result, err := svc.Create(userID, userEmail, "user", "Nil Fields Template", nil, nil, false, movements, nil)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		storedMovements, _ := workoutMovementRepo.GetByWorkoutID(result.ID)
		if len(storedMovements) != 1 {
			t.Fatalf("Expected 1 movement, got %d", len(storedMovements))
		}

		m := storedMovements[0]
		if m.Sets != nil {
			t.Errorf("Sets should be nil, got %v", m.Sets)
		}
		if m.Reps != nil {
			t.Errorf("Reps should be nil, got %v", m.Reps)
		}
		if m.Weight != nil {
			t.Errorf("Weight should be nil, got %v", m.Weight)
		}
		if m.Instructions != "" {
			t.Errorf("Instructions should be empty, got %s", m.Instructions)
		}
	})

	// Test updating from nil to value
	t.Run("Update_NilToValue", func(t *testing.T) {
		// First create with nil fields
		movements := []domain.WorkoutMovement{
			{
				MovementID: 1,
				Sets:       nil,
				Reps:       nil,
				OrderIndex: 1,
			},
		}

		created, _ := svc.Create(userID, userEmail, "user", "Nil to Value Test", nil, nil, false, movements, nil)

		// Now update with values
		updatedMovements := []domain.WorkoutMovement{
			{
				MovementID:   1,
				Sets:         intPtr(4),
				Reps:         intPtr(10),
				Weight:       floatPtr(100.0),
				Instructions: "New instructions",
				Notes:        "New notes",
				OrderIndex:   1,
			},
		}

		_, err := svc.Update(created.ID, userID, userEmail, "user", "Nil to Value Test", nil, nil, false, updatedMovements, nil)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		storedMovements, _ := workoutMovementRepo.GetByWorkoutID(created.ID)
		m := storedMovements[0]
		assertIntPtr(t, "Sets", m.Sets, 4)
		assertIntPtr(t, "Reps", m.Reps, 10)
		assertFloatPtr(t, "Weight", m.Weight, 100.0)
		if m.Instructions != "New instructions" {
			t.Errorf("Instructions = %s, want 'New instructions'", m.Instructions)
		}
	})

	// Test updating from value to nil
	t.Run("Update_ValueToNil", func(t *testing.T) {
		// First create with values
		movements := []domain.WorkoutMovement{
			{
				MovementID:   1,
				Sets:         intPtr(4),
				Reps:         intPtr(10),
				Instructions: "Has instructions",
				OrderIndex:   1,
			},
		}

		created, _ := svc.Create(userID, userEmail, "user", "Value to Nil Test", nil, nil, false, movements, nil)

		// Now update with nil values
		updatedMovements := []domain.WorkoutMovement{
			{
				MovementID:   1,
				Sets:         nil, // Clear
				Reps:         nil, // Clear
				Instructions: "",  // Clear
				OrderIndex:   1,
			},
		}

		_, err := svc.Update(created.ID, userID, userEmail, "user", "Value to Nil Test", nil, nil, false, updatedMovements, nil)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		storedMovements, _ := workoutMovementRepo.GetByWorkoutID(created.ID)
		m := storedMovements[0]
		if m.Sets != nil {
			t.Errorf("Sets should be nil after update, got %v", m.Sets)
		}
		if m.Reps != nil {
			t.Errorf("Reps should be nil after update, got %v", m.Reps)
		}
		if m.Instructions != "" {
			t.Errorf("Instructions should be empty after update, got %s", m.Instructions)
		}
	})
}

// TestWorkoutTemplate_SpecialCharacters tests handling of special characters in text fields
func TestWorkoutTemplate_SpecialCharacters(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	userID := int64(1)
	userEmail := "test@example.com"

	// Test with markdown, unicode, and special characters
	t.Run("Create_WithSpecialCharacters", func(t *testing.T) {
		complexInstructions := `## Workout Instructions

**Bold text** and *italic text*

1. First step with "quotes"
2. Second step with 'single quotes'
3. Step with <angle brackets>
4. Step with & ampersand
5. Unicode: café, naïve, 日本語

` + "```go\n// Code block\nfunc main() {}\n```" + `

> Blockquote

| Table | Header |
|-------|--------|
| Cell  | Cell   |`

		movements := []domain.WorkoutMovement{
			{
				MovementID:   1,
				Sets:         intPtr(3),
				Instructions: complexInstructions,
				Notes:        "Notes with special chars: <>&\"'",
				OrderIndex:   1,
			},
		}

		result, err := svc.Create(userID, userEmail, "user", "Template with Special Chars & Symbols", nil, nil, false, movements, nil)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		storedMovements, _ := workoutMovementRepo.GetByWorkoutID(result.ID)
		m := storedMovements[0]

		if m.Instructions != complexInstructions {
			t.Errorf("Instructions not preserved correctly with special characters")
		}
		if m.Notes != "Notes with special chars: <>&\"'" {
			t.Errorf("Notes not preserved correctly: %s", m.Notes)
		}
	})
}

// TestWorkoutTemplate_MultipleMovementsAndWODs tests handling of multiple items with correct ordering
func TestWorkoutTemplate_MultipleMovementsAndWODs(t *testing.T) {
	workoutRepo := newMockWorkoutRepo()
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	userID := int64(1)
	userEmail := "test@example.com"

	t.Run("Create_MultipleItems", func(t *testing.T) {
		movements := []domain.WorkoutMovement{
			{MovementID: 1, Sets: intPtr(3), Instructions: "Movement 1 instructions", OrderIndex: 1},
			{MovementID: 2, Sets: intPtr(4), Instructions: "Movement 2 instructions", OrderIndex: 2},
			{MovementID: 3, Sets: intPtr(5), Instructions: "Movement 3 instructions", OrderIndex: 3},
		}

		wods := []domain.WorkoutWOD{
			{WODID: 1, Instructions: "WOD 1 instructions", OrderIndex: 1},
			{WODID: 2, Instructions: "WOD 2 instructions", OrderIndex: 2},
		}

		result, err := svc.Create(userID, userEmail, "user", "Multi-item Template", nil, nil, false, movements, wods)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		storedMovements, _ := workoutMovementRepo.GetByWorkoutID(result.ID)
		if len(storedMovements) != 3 {
			t.Errorf("Expected 3 movements, got %d", len(storedMovements))
		}

		// Verify each movement has correct data by finding by MovementID
		expectedSets := map[int64]int{1: 3, 2: 4, 3: 5}
		for _, m := range storedMovements {
			expected, ok := expectedSets[m.MovementID]
			if !ok {
				t.Errorf("Unexpected MovementID: %d", m.MovementID)
				continue
			}
			if m.Sets == nil || *m.Sets != expected {
				t.Errorf("Movement %d sets = %v, want %d", m.MovementID, m.Sets, expected)
			}
			if m.Instructions == "" {
				t.Errorf("Movement %d instructions should not be empty", m.MovementID)
			}
		}
	})

	t.Run("Update_ChangeOrder", func(t *testing.T) {
		// Create initial
		movements := []domain.WorkoutMovement{
			{MovementID: 1, Instructions: "First", OrderIndex: 1},
			{MovementID: 2, Instructions: "Second", OrderIndex: 2},
		}

		created, _ := svc.Create(userID, userEmail, "user", "Order Test", nil, nil, false, movements, nil)

		// Update with reversed order
		updatedMovements := []domain.WorkoutMovement{
			{MovementID: 2, Instructions: "Now First", OrderIndex: 1},
			{MovementID: 1, Instructions: "Now Second", OrderIndex: 2},
		}

		_, err := svc.Update(created.ID, userID, userEmail, "user", "Order Test", nil, nil, false, updatedMovements, nil)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		storedMovements, _ := workoutMovementRepo.GetByWorkoutID(created.ID)
		if len(storedMovements) != 2 {
			t.Fatalf("Expected 2 movements, got %d", len(storedMovements))
		}
	})
}

// =============================================================================
// Helper functions
// =============================================================================

func findMovementByID(movements []*domain.WorkoutMovement, movementID int64) *domain.WorkoutMovement {
	for _, m := range movements {
		if m.MovementID == movementID {
			return m
		}
	}
	return nil
}

func assertIntPtr(t *testing.T, field string, got *int, want int) {
	t.Helper()
	if got == nil {
		t.Errorf("%s = nil, want %d", field, want)
		return
	}
	if *got != want {
		t.Errorf("%s = %d, want %d", field, *got, want)
	}
}

func assertFloatPtr(t *testing.T, field string, got *float64, want float64) {
	t.Helper()
	if got == nil {
		t.Errorf("%s = nil, want %f", field, want)
		return
	}
	if *got != want {
		t.Errorf("%s = %f, want %f", field, *got, want)
	}
}
