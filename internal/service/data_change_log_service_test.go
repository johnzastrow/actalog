package service

import (
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestNewDataChangeLogService(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	if svc == nil {
		t.Fatal("NewDataChangeLogService returned nil")
	}
	if svc.repo != repo {
		t.Error("Service repo not set correctly")
	}
}

func TestDataChangeLogService_LogUpdate(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"name": "Old Name"}
	after := map[string]interface{}{"name": "New Name"}
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	err := svc.LogUpdate("wod", 1, "Test WOD", 42, "user@example.com", before, after, &ipAddress, &userAgent)
	if err != nil {
		t.Errorf("LogUpdate() error = %v", err)
	}

	if len(repo.logs) != 1 {
		t.Fatalf("Expected 1 log, got %d", len(repo.logs))
	}

	log := repo.logs[0]
	if log.EntityType != "wod" {
		t.Errorf("EntityType = %s, want wod", log.EntityType)
	}
	if log.EntityID != 1 {
		t.Errorf("EntityID = %d, want 1", log.EntityID)
	}
	if log.EntityName != "Test WOD" {
		t.Errorf("EntityName = %s, want 'Test WOD'", log.EntityName)
	}
	if log.Operation != domain.OperationUpdate {
		t.Errorf("Operation = %s, want %s", log.Operation, domain.OperationUpdate)
	}
	if log.UserID != 42 {
		t.Errorf("UserID = %d, want 42", log.UserID)
	}
	if log.BeforeValues == nil {
		t.Error("BeforeValues should not be nil")
	}
	if log.AfterValues == nil {
		t.Error("AfterValues should not be nil")
	}
}

func TestDataChangeLogService_LogDelete(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"name": "Deleted WOD", "score_type": "time"}
	ipAddress := "192.168.1.1"

	err := svc.LogDelete("wod", 1, "Deleted WOD", 42, "user@example.com", before, &ipAddress, nil)
	if err != nil {
		t.Errorf("LogDelete() error = %v", err)
	}

	if len(repo.logs) != 1 {
		t.Fatalf("Expected 1 log, got %d", len(repo.logs))
	}

	log := repo.logs[0]
	if log.Operation != domain.OperationDelete {
		t.Errorf("Operation = %s, want %s", log.Operation, domain.OperationDelete)
	}
	if log.BeforeValues == nil {
		t.Error("BeforeValues should not be nil")
	}
	if log.AfterValues != nil {
		t.Error("AfterValues should be nil for delete operations")
	}
}

func TestDataChangeLogService_GetByID(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	// Create a log first
	log := &domain.DataChangeLog{
		EntityType: "wod",
		EntityID:   1,
		EntityName: "Test WOD",
		Operation:  domain.OperationUpdate,
		UserID:     42,
		UserEmail:  "user@example.com",
	}
	_ = repo.Create(log)

	found, err := svc.GetByID(log.ID)
	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByID() returned nil")
	}
	if found.ID != log.ID {
		t.Errorf("GetByID() returned wrong log: got ID %d, want %d", found.ID, log.ID)
	}
}

func TestDataChangeLogService_List(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	// Create some logs
	for i := 0; i < 5; i++ {
		_ = repo.Create(&domain.DataChangeLog{
			EntityType: "wod",
			EntityID:   int64(i + 1),
			EntityName: "Test WOD",
			Operation:  domain.OperationUpdate,
			UserID:     42,
			UserEmail:  "user@example.com",
		})
	}

	tests := []struct {
		name   string
		limit  int
		offset int
	}{
		{"default values", 0, -1},
		{"over max limit", 200, 0},
		{"valid limit", 10, 0},
		{"negative offset", 10, -5},
		{"valid pagination", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, err := svc.List(domain.DataChangeLogFilters{}, tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
			}
			if logs == nil {
				t.Error("List() returned nil")
			}
		})
	}
}

func TestDataChangeLogService_Count(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	for i := 0; i < 3; i++ {
		_ = repo.Create(&domain.DataChangeLog{
			EntityType: "wod",
			EntityID:   int64(i + 1),
			EntityName: "Test WOD",
			Operation:  domain.OperationUpdate,
			UserID:     42,
			UserEmail:  "user@example.com",
		})
	}

	count, err := svc.Count(domain.DataChangeLogFilters{})
	if err != nil {
		t.Errorf("Count() error = %v", err)
	}
	if count != 3 {
		t.Errorf("Count() = %d, want 3", count)
	}
}

func TestDataChangeLogService_GetByEntityID(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	// Create logs for different entities
	_ = repo.Create(&domain.DataChangeLog{EntityType: "wod", EntityID: 1, EntityName: "WOD 1", Operation: domain.OperationUpdate, UserID: 1, UserEmail: "a@a.com"})
	_ = repo.Create(&domain.DataChangeLog{EntityType: "wod", EntityID: 1, EntityName: "WOD 1", Operation: domain.OperationUpdate, UserID: 1, UserEmail: "a@a.com"})
	_ = repo.Create(&domain.DataChangeLog{EntityType: "wod", EntityID: 2, EntityName: "WOD 2", Operation: domain.OperationUpdate, UserID: 1, UserEmail: "a@a.com"})
	_ = repo.Create(&domain.DataChangeLog{EntityType: "movement", EntityID: 1, EntityName: "Movement 1", Operation: domain.OperationUpdate, UserID: 1, UserEmail: "a@a.com"})

	logs, err := svc.GetByEntityID("wod", 1, 50, 0)
	if err != nil {
		t.Errorf("GetByEntityID() error = %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("GetByEntityID() returned %d logs, want 2", len(logs))
	}
}

func TestDataChangeLogService_GetByUserID(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	_ = repo.Create(&domain.DataChangeLog{EntityType: "wod", EntityID: 1, EntityName: "WOD 1", Operation: domain.OperationUpdate, UserID: 1, UserEmail: "a@a.com"})
	_ = repo.Create(&domain.DataChangeLog{EntityType: "wod", EntityID: 2, EntityName: "WOD 2", Operation: domain.OperationUpdate, UserID: 1, UserEmail: "a@a.com"})
	_ = repo.Create(&domain.DataChangeLog{EntityType: "wod", EntityID: 3, EntityName: "WOD 3", Operation: domain.OperationUpdate, UserID: 2, UserEmail: "b@b.com"})

	logs, err := svc.GetByUserID(1, 50, 0)
	if err != nil {
		t.Errorf("GetByUserID() error = %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("GetByUserID() returned %d logs, want 2", len(logs))
	}
}

func TestDataChangeLogService_CleanupOldLogs(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	tests := []struct {
		name          string
		retentionDays int
		wantError     bool
	}{
		{"valid retention", 30, false},
		{"zero retention", 0, true},
		{"negative retention", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CleanupOldLogs(tt.retentionDays)
			if (err != nil) != tt.wantError {
				t.Errorf("CleanupOldLogs() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestDataChangeLogService_LogWODUpdate(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"name": "Old"}
	after := map[string]interface{}{"name": "New"}

	err := svc.LogWODUpdate(1, "Test WOD", 42, "user@example.com", before, after, nil, nil)
	if err != nil {
		t.Errorf("LogWODUpdate() error = %v", err)
	}

	if repo.logs[0].EntityType != domain.EntityTypeWOD {
		t.Errorf("EntityType = %s, want %s", repo.logs[0].EntityType, domain.EntityTypeWOD)
	}
}

func TestDataChangeLogService_LogWODDelete(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"name": "Deleted WOD"}

	err := svc.LogWODDelete(1, "Test WOD", 42, "user@example.com", before, nil, nil)
	if err != nil {
		t.Errorf("LogWODDelete() error = %v", err)
	}

	if repo.logs[0].EntityType != domain.EntityTypeWOD {
		t.Errorf("EntityType = %s, want %s", repo.logs[0].EntityType, domain.EntityTypeWOD)
	}
	if repo.logs[0].Operation != domain.OperationDelete {
		t.Errorf("Operation = %s, want %s", repo.logs[0].Operation, domain.OperationDelete)
	}
}

func TestDataChangeLogService_LogMovementUpdate(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"name": "Old"}
	after := map[string]interface{}{"name": "New"}

	err := svc.LogMovementUpdate(1, "Back Squat", 42, "user@example.com", before, after, nil, nil)
	if err != nil {
		t.Errorf("LogMovementUpdate() error = %v", err)
	}

	if repo.logs[0].EntityType != domain.EntityTypeMovement {
		t.Errorf("EntityType = %s, want %s", repo.logs[0].EntityType, domain.EntityTypeMovement)
	}
}

func TestDataChangeLogService_LogMovementDelete(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"name": "Deleted Movement"}

	err := svc.LogMovementDelete(1, "Back Squat", 42, "user@example.com", before, nil, nil)
	if err != nil {
		t.Errorf("LogMovementDelete() error = %v", err)
	}

	if repo.logs[0].EntityType != domain.EntityTypeMovement {
		t.Errorf("EntityType = %s, want %s", repo.logs[0].EntityType, domain.EntityTypeMovement)
	}
}

func TestDataChangeLogService_LogWorkoutUpdate(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"name": "Old"}
	after := map[string]interface{}{"name": "New"}

	err := svc.LogWorkoutUpdate(1, "Workout Template", 42, "user@example.com", before, after, nil, nil)
	if err != nil {
		t.Errorf("LogWorkoutUpdate() error = %v", err)
	}

	if repo.logs[0].EntityType != domain.EntityTypeWorkout {
		t.Errorf("EntityType = %s, want %s", repo.logs[0].EntityType, domain.EntityTypeWorkout)
	}
}

func TestDataChangeLogService_LogWorkoutDelete(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"name": "Deleted Workout"}

	err := svc.LogWorkoutDelete(1, "Workout Template", 42, "user@example.com", before, nil, nil)
	if err != nil {
		t.Errorf("LogWorkoutDelete() error = %v", err)
	}

	if repo.logs[0].EntityType != domain.EntityTypeWorkout {
		t.Errorf("EntityType = %s, want %s", repo.logs[0].EntityType, domain.EntityTypeWorkout)
	}
}

func TestDataChangeLogService_LogUserWorkoutUpdate(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"notes": "Old notes"}
	after := map[string]interface{}{"notes": "New notes"}

	err := svc.LogUserWorkoutUpdate(1, "Monday Workout", 42, "user@example.com", before, after, nil, nil)
	if err != nil {
		t.Errorf("LogUserWorkoutUpdate() error = %v", err)
	}

	if repo.logs[0].EntityType != domain.EntityTypeUserWorkout {
		t.Errorf("EntityType = %s, want %s", repo.logs[0].EntityType, domain.EntityTypeUserWorkout)
	}
}

func TestDataChangeLogService_LogUserWorkoutDelete(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"notes": "Deleted workout"}

	err := svc.LogUserWorkoutDelete(1, "Monday Workout", 42, "user@example.com", before, nil, nil)
	if err != nil {
		t.Errorf("LogUserWorkoutDelete() error = %v", err)
	}

	if repo.logs[0].EntityType != domain.EntityTypeUserWorkout {
		t.Errorf("EntityType = %s, want %s", repo.logs[0].EntityType, domain.EntityTypeUserWorkout)
	}
}

func TestDataChangeLogService_LogUserWorkoutMovementUpdate(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"weight": 100}
	after := map[string]interface{}{"weight": 110}

	err := svc.LogUserWorkoutMovementUpdate(1, "Back Squat", 42, "user@example.com", before, after, nil, nil)
	if err != nil {
		t.Errorf("LogUserWorkoutMovementUpdate() error = %v", err)
	}

	if repo.logs[0].EntityType != domain.EntityTypeUserWorkoutMovement {
		t.Errorf("EntityType = %s, want %s", repo.logs[0].EntityType, domain.EntityTypeUserWorkoutMovement)
	}
}

func TestDataChangeLogService_LogUserWorkoutMovementDelete(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"weight": 100}

	err := svc.LogUserWorkoutMovementDelete(1, "Back Squat", 42, "user@example.com", before, nil, nil)
	if err != nil {
		t.Errorf("LogUserWorkoutMovementDelete() error = %v", err)
	}

	if repo.logs[0].EntityType != domain.EntityTypeUserWorkoutMovement {
		t.Errorf("EntityType = %s, want %s", repo.logs[0].EntityType, domain.EntityTypeUserWorkoutMovement)
	}
}

func TestDataChangeLogService_LogUserWorkoutWODUpdate(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"time_seconds": 300}
	after := map[string]interface{}{"time_seconds": 280}

	err := svc.LogUserWorkoutWODUpdate(1, "Fran", 42, "user@example.com", before, after, nil, nil)
	if err != nil {
		t.Errorf("LogUserWorkoutWODUpdate() error = %v", err)
	}

	if repo.logs[0].EntityType != domain.EntityTypeUserWorkoutWOD {
		t.Errorf("EntityType = %s, want %s", repo.logs[0].EntityType, domain.EntityTypeUserWorkoutWOD)
	}
}

func TestDataChangeLogService_LogUserWorkoutWODDelete(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"time_seconds": 300}

	err := svc.LogUserWorkoutWODDelete(1, "Fran", 42, "user@example.com", before, nil, nil)
	if err != nil {
		t.Errorf("LogUserWorkoutWODDelete() error = %v", err)
	}

	if repo.logs[0].EntityType != domain.EntityTypeUserWorkoutWOD {
		t.Errorf("EntityType = %s, want %s", repo.logs[0].EntityType, domain.EntityTypeUserWorkoutWOD)
	}
}

func TestDataChangeLogService_LogUserUpdate(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"name": "Old Name"}
	after := map[string]interface{}{"name": "New Name"}

	err := svc.LogUserUpdate(2, "target@example.com", 1, "admin@example.com", before, after, nil, nil)
	if err != nil {
		t.Errorf("LogUserUpdate() error = %v", err)
	}

	if repo.logs[0].EntityType != domain.EntityTypeUser {
		t.Errorf("EntityType = %s, want %s", repo.logs[0].EntityType, domain.EntityTypeUser)
	}
}

func TestDataChangeLogService_LogUserDelete(t *testing.T) {
	repo := newMockDataChangeLogRepo()
	svc := NewDataChangeLogService(repo)

	before := map[string]interface{}{"email": "deleted@example.com"}

	err := svc.LogUserDelete(2, "deleted@example.com", 1, "admin@example.com", before, nil, nil)
	if err != nil {
		t.Errorf("LogUserDelete() error = %v", err)
	}

	if repo.logs[0].EntityType != domain.EntityTypeUser {
		t.Errorf("EntityType = %s, want %s", repo.logs[0].EntityType, domain.EntityTypeUser)
	}
	if repo.logs[0].Operation != domain.OperationDelete {
		t.Errorf("Operation = %s, want %s", repo.logs[0].Operation, domain.OperationDelete)
	}
}
