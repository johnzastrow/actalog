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
	workoutMovementRepo := newMockWorkoutMovementRepo()
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	_, err := svc.Update(999, 1, "user@example.com", "Name", nil, nil, nil)
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
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
	workoutMovementRepo := newMockWorkoutMovementRepo()
	workoutWODRepo := newMockWorkoutWODRepo()

	svc := NewWorkoutTemplateService(workoutRepo, workoutMovementRepo, workoutWODRepo, nil)

	err := svc.Delete(999, 1, "user@example.com")
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

	_, err := svc.Create(1, "user@example.com", "Test", nil, nil, nil)
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
	_, err := svc.Create(1, "user@example.com", "Test", nil, movements, nil)
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
	_, err := svc.Create(1, "user@example.com", "Test", nil, nil, wods)
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

	_, err := svc.Update(workout.ID, 1, "user@example.com", "Updated", nil, nil, nil)
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

	_, err := svc.Update(workout.ID, 1, "user@example.com", "Updated", nil, nil, nil)
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
	_, err := svc.Update(workout.ID, 1, "user@example.com", "Updated", nil, movements, nil)
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

	_, err := svc.Update(workout.ID, 1, "user@example.com", "Updated", nil, nil, nil)
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
	_, err := svc.Update(workout.ID, 1, "user@example.com", "Updated", nil, nil, wods)
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

	_, err := svc.Update(1, 1, "user@example.com", "Updated", nil, nil, nil)
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
	_, err := svc.Update(workout.ID, 1, "user@example.com", "Updated", nil, nil, nil)
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

	err := svc.Delete(workout.ID, 1, "user@example.com")
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

	err := svc.Delete(workout.ID, 1, "user@example.com")
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

	err := svc.Delete(1, 1, "user@example.com")
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
	err := svc.Delete(workout.ID, 1, "user@example.com")
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
	result, err := svc.Update(workout.ID, 1, "user@example.com", "Updated Name", nil, nil, wods)
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

	result, err := svc.Create(1, "user@example.com", "Template With Instructions", stringPtr("Test notes"), movements, wods)
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

	result, err := svc.Update(workout.ID, 1, "user@example.com", "Updated Name", stringPtr("Updated notes"), movements, wods)
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

	result, err := svc.Create(1, "user@example.com", "Full Movement Details", nil, movements, nil)
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

	result, err := svc.Create(1, "user@example.com", "Full WOD Details", nil, nil, wods)
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
