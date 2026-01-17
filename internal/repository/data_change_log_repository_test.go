package repository

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestDataChangeLogRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewDataChangeLogRepository(db, "sqlite3")

	beforeValues := `{"name": "Old Name"}`
	afterValues := `{"name": "New Name"}`
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	tests := []struct {
		name    string
		log     *domain.DataChangeLog
		wantErr bool
	}{
		{
			name: "Create update log with all fields",
			log: &domain.DataChangeLog{
				EntityType:   domain.EntityTypeWOD,
				EntityID:     1,
				EntityName:   "Fran",
				Operation:    domain.OperationUpdate,
				UserID:       1,
				UserEmail:    "admin@example.com",
				BeforeValues: &beforeValues,
				AfterValues:  &afterValues,
				IPAddress:    &ipAddress,
				UserAgent:    &userAgent,
			},
			wantErr: false,
		},
		{
			name: "Create delete log (no after values)",
			log: &domain.DataChangeLog{
				EntityType:   domain.EntityTypeMovement,
				EntityID:     5,
				EntityName:   "Deleted Movement",
				Operation:    domain.OperationDelete,
				UserID:       2,
				UserEmail:    "user@example.com",
				BeforeValues: &beforeValues,
				AfterValues:  nil,
			},
			wantErr: false,
		},
		{
			name: "Create log with minimal fields",
			log: &domain.DataChangeLog{
				EntityType: domain.EntityTypeUserWorkout,
				EntityID:   10,
				EntityName: "Morning Workout",
				Operation:  domain.OperationUpdate,
				UserID:     1,
				UserEmail:  "test@example.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.log)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.log.ID == 0 {
				t.Error("Create() should set ID")
			}
			if !tt.wantErr && tt.log.CreatedAt.IsZero() {
				t.Error("Create() should set CreatedAt")
			}
		})
	}
}

func TestDataChangeLogRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewDataChangeLogRepository(db, "sqlite3")

	// Create a log entry
	beforeValues := `{"name": "Old"}`
	afterValues := `{"name": "New"}`
	log := &domain.DataChangeLog{
		EntityType:   domain.EntityTypeWOD,
		EntityID:     1,
		EntityName:   "Test WOD",
		Operation:    domain.OperationUpdate,
		UserID:       1,
		UserEmail:    "test@example.com",
		BeforeValues: &beforeValues,
		AfterValues:  &afterValues,
	}
	if err := repo.Create(log); err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	// Test GetByID
	got, err := repo.GetByID(log.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByID() returned nil")
	}
	if got.ID != log.ID {
		t.Errorf("ID = %d, want %d", got.ID, log.ID)
	}
	if got.EntityType != domain.EntityTypeWOD {
		t.Errorf("EntityType = %v, want %v", got.EntityType, domain.EntityTypeWOD)
	}
	if got.EntityName != "Test WOD" {
		t.Errorf("EntityName = %v, want 'Test WOD'", got.EntityName)
	}
	if got.Operation != domain.OperationUpdate {
		t.Errorf("Operation = %v, want %v", got.Operation, domain.OperationUpdate)
	}
	if got.BeforeValues == nil || *got.BeforeValues != beforeValues {
		t.Error("BeforeValues should match")
	}

	// Test non-existent log
	_, err = repo.GetByID(999)
	if err == nil {
		t.Error("GetByID(999) should return error")
	}
}

func TestDataChangeLogRepository_List(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewDataChangeLogRepository(db, "sqlite3")

	// Create test data using direct SQL to have consistent data
	now := time.Now()
	logsData := []struct {
		entityType string
		entityID   int64
		entityName string
		operation  string
		userID     int64
		userEmail  string
	}{
		{domain.EntityTypeWOD, 1, "Fran", domain.OperationUpdate, 1, "admin@example.com"},
		{domain.EntityTypeWOD, 2, "Grace", domain.OperationUpdate, 1, "admin@example.com"},
		{domain.EntityTypeWOD, 3, "Helen", domain.OperationDelete, 1, "admin@example.com"},
		{domain.EntityTypeMovement, 1, "Back Squat", domain.OperationUpdate, 2, "user@example.com"},
		{domain.EntityTypeMovement, 2, "Front Squat", domain.OperationDelete, 2, "user@example.com"},
	}

	for _, l := range logsData {
		_, err := db.Exec(`INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			l.entityType, l.entityID, l.entityName, l.operation, l.userID, l.userEmail, now)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	entityTypeWOD := domain.EntityTypeWOD
	entityTypeMovement := domain.EntityTypeMovement
	operationUpdate := domain.OperationUpdate
	operationDelete := domain.OperationDelete
	userID1 := int64(1)
	userID2 := int64(2)
	userEmail := "admin"

	tests := []struct {
		name      string
		filters   domain.DataChangeLogFilters
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "List all",
			filters:   domain.DataChangeLogFilters{},
			limit:     100,
			offset:    0,
			wantCount: 5,
		},
		{
			name:      "Filter by entity type - WOD",
			filters:   domain.DataChangeLogFilters{EntityType: &entityTypeWOD},
			limit:     100,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "Filter by entity type - Movement",
			filters:   domain.DataChangeLogFilters{EntityType: &entityTypeMovement},
			limit:     100,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "Filter by operation - Update",
			filters:   domain.DataChangeLogFilters{Operation: &operationUpdate},
			limit:     100,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "Filter by operation - Delete",
			filters:   domain.DataChangeLogFilters{Operation: &operationDelete},
			limit:     100,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "Filter by user ID",
			filters:   domain.DataChangeLogFilters{UserID: &userID1},
			limit:     100,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "Filter by user ID 2",
			filters:   domain.DataChangeLogFilters{UserID: &userID2},
			limit:     100,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "Filter by user email (partial)",
			filters:   domain.DataChangeLogFilters{UserEmail: &userEmail},
			limit:     100,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "With limit",
			filters:   domain.DataChangeLogFilters{},
			limit:     2,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "With offset",
			filters:   domain.DataChangeLogFilters{},
			limit:     100,
			offset:    3,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.List(tt.filters, tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("List() returned %d logs, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestDataChangeLogRepository_List_DateFilters(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewDataChangeLogRepository(db, "sqlite3")

	// Create logs at different times
	now := time.Now()
	times := []time.Time{
		now.AddDate(0, 0, -10), // 10 days ago
		now.AddDate(0, 0, -5),  // 5 days ago
		now.AddDate(0, 0, -3),  // 3 days ago
		now.AddDate(0, 0, -1),  // 1 day ago
		now,                    // today
	}

	for i, ts := range times {
		_, err := db.Exec(`INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			domain.EntityTypeWOD, int64(i+1), "Test WOD", domain.OperationUpdate, 1, "test@example.com", ts)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Test date range filter
	startDate := now.AddDate(0, 0, -4) // 4 days ago
	endDate := now.AddDate(0, 0, 1)    // tomorrow

	filters := domain.DataChangeLogFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	got, err := repo.List(filters, 100, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Should get logs from -3, -1, and today (3 logs)
	if len(got) != 3 {
		t.Errorf("List() with date filter returned %d logs, want 3", len(got))
	}
}

func TestDataChangeLogRepository_List_EntityIDFilter(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewDataChangeLogRepository(db, "sqlite3")
	now := time.Now()

	// Create multiple logs for different entities
	for i := 1; i <= 3; i++ {
		for j := 0; j < i; j++ { // Entity 1 has 1 log, Entity 2 has 2 logs, Entity 3 has 3 logs
			_, err := db.Exec(`INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
				domain.EntityTypeWOD, int64(i), "Test WOD", domain.OperationUpdate, 1, "test@example.com", now)
			if err != nil {
				t.Fatalf("Failed to create log: %v", err)
			}
		}
	}

	entityID := int64(2)
	filters := domain.DataChangeLogFilters{EntityID: &entityID}

	got, err := repo.List(filters, 100, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(got) != 2 {
		t.Errorf("List() with entity ID filter returned %d logs, want 2", len(got))
	}
}

func TestDataChangeLogRepository_Count(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewDataChangeLogRepository(db, "sqlite3")
	now := time.Now()

	// Create test data
	logsData := []struct {
		entityType string
		operation  string
		userID     int64
	}{
		{domain.EntityTypeWOD, domain.OperationUpdate, 1},
		{domain.EntityTypeWOD, domain.OperationUpdate, 1},
		{domain.EntityTypeWOD, domain.OperationDelete, 1},
		{domain.EntityTypeMovement, domain.OperationUpdate, 2},
		{domain.EntityTypeMovement, domain.OperationDelete, 2},
	}

	for i, l := range logsData {
		_, err := db.Exec(`INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			l.entityType, int64(i+1), "Test", l.operation, l.userID, "test@example.com", now)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	entityTypeWOD := domain.EntityTypeWOD
	operationUpdate := domain.OperationUpdate
	userID1 := int64(1)

	tests := []struct {
		name      string
		filters   domain.DataChangeLogFilters
		wantCount int
	}{
		{
			name:      "Count all",
			filters:   domain.DataChangeLogFilters{},
			wantCount: 5,
		},
		{
			name:      "Count by entity type",
			filters:   domain.DataChangeLogFilters{EntityType: &entityTypeWOD},
			wantCount: 3,
		},
		{
			name:      "Count by operation",
			filters:   domain.DataChangeLogFilters{Operation: &operationUpdate},
			wantCount: 3,
		},
		{
			name:      "Count by user ID",
			filters:   domain.DataChangeLogFilters{UserID: &userID1},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.Count(tt.filters)
			if err != nil {
				t.Errorf("Count() error = %v", err)
				return
			}
			if got != tt.wantCount {
				t.Errorf("Count() = %d, want %d", got, tt.wantCount)
			}
		})
	}
}

func TestDataChangeLogRepository_GetByEntityID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewDataChangeLogRepository(db, "sqlite3")
	now := time.Now()

	// Create logs for different entities
	// WOD ID 1: 3 changes
	for i := 0; i < 3; i++ {
		_, err := db.Exec(`INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			domain.EntityTypeWOD, 1, "Fran", domain.OperationUpdate, 1, "test@example.com", now)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Movement ID 1: 2 changes
	for i := 0; i < 2; i++ {
		_, err := db.Exec(`INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			domain.EntityTypeMovement, 1, "Back Squat", domain.OperationUpdate, 1, "test@example.com", now)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// WOD ID 2: 1 change
	_, err = db.Exec(`INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		domain.EntityTypeWOD, 2, "Grace", domain.OperationUpdate, 1, "test@example.com", now)
	if err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	// Test GetByEntityID for WOD 1
	got, err := repo.GetByEntityID(domain.EntityTypeWOD, 1, 100, 0)
	if err != nil {
		t.Fatalf("GetByEntityID() error = %v", err)
	}
	if len(got) != 3 {
		t.Errorf("GetByEntityID(wod, 1) returned %d logs, want 3", len(got))
	}

	// Test GetByEntityID for Movement 1
	got, err = repo.GetByEntityID(domain.EntityTypeMovement, 1, 100, 0)
	if err != nil {
		t.Fatalf("GetByEntityID() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetByEntityID(movement, 1) returned %d logs, want 2", len(got))
	}

	// Test with pagination
	got, err = repo.GetByEntityID(domain.EntityTypeWOD, 1, 2, 0)
	if err != nil {
		t.Fatalf("GetByEntityID() with limit error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetByEntityID() with limit returned %d logs, want 2", len(got))
	}

	// Test non-existent entity
	got, err = repo.GetByEntityID(domain.EntityTypeWOD, 999, 100, 0)
	if err != nil {
		t.Fatalf("GetByEntityID(999) error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("GetByEntityID(999) returned %d logs, want 0", len(got))
	}
}

func TestDataChangeLogRepository_GetByUserID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewDataChangeLogRepository(db, "sqlite3")
	now := time.Now()

	// Create logs for user 1 (5 changes)
	for i := 0; i < 5; i++ {
		_, err := db.Exec(`INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			domain.EntityTypeWOD, int64(i+1), "Test", domain.OperationUpdate, 1, "user1@example.com", now)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Create logs for user 2 (3 changes)
	for i := 0; i < 3; i++ {
		_, err := db.Exec(`INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			domain.EntityTypeMovement, int64(i+1), "Test", domain.OperationUpdate, 2, "user2@example.com", now)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Test GetByUserID for user 1
	got, err := repo.GetByUserID(1, 100, 0)
	if err != nil {
		t.Fatalf("GetByUserID(1) error = %v", err)
	}
	if len(got) != 5 {
		t.Errorf("GetByUserID(1) returned %d logs, want 5", len(got))
	}

	// Test GetByUserID for user 2
	got, err = repo.GetByUserID(2, 100, 0)
	if err != nil {
		t.Fatalf("GetByUserID(2) error = %v", err)
	}
	if len(got) != 3 {
		t.Errorf("GetByUserID(2) returned %d logs, want 3", len(got))
	}

	// Test with limit
	got, err = repo.GetByUserID(1, 2, 0)
	if err != nil {
		t.Fatalf("GetByUserID() with limit error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetByUserID() with limit returned %d logs, want 2", len(got))
	}

	// Test non-existent user
	got, err = repo.GetByUserID(999, 100, 0)
	if err != nil {
		t.Fatalf("GetByUserID(999) error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("GetByUserID(999) returned %d logs, want 0", len(got))
	}
}

func TestDataChangeLogRepository_DeleteOlderThan(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewDataChangeLogRepository(db, "sqlite3")

	// Create logs at different times
	now := time.Now()
	times := []time.Time{
		now.AddDate(0, 0, -30), // 30 days ago (old)
		now.AddDate(0, 0, -20), // 20 days ago (old)
		now.AddDate(0, 0, -10), // 10 days ago (old)
		now.AddDate(0, 0, -5),  // 5 days ago (recent)
		now,                    // today (recent)
	}

	for i, ts := range times {
		_, err := db.Exec(`INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			domain.EntityTypeWOD, int64(i+1), "Test", domain.OperationUpdate, 1, "test@example.com", ts)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Verify we have 5 logs
	count, err := repo.Count(domain.DataChangeLogFilters{})
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 5 {
		t.Fatalf("Expected 5 logs, got %d", count)
	}

	// Delete logs older than 7 days ago
	cutoff := now.AddDate(0, 0, -7)
	deleted, err := repo.DeleteOlderThan(cutoff)
	if err != nil {
		t.Fatalf("DeleteOlderThan() error = %v", err)
	}
	if deleted != 3 {
		t.Errorf("DeleteOlderThan() deleted %d logs, want 3", deleted)
	}

	// Verify remaining count
	count, err = repo.Count(domain.DataChangeLogFilters{})
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 2 {
		t.Errorf("After deletion, expected 2 logs, got %d", count)
	}

	// Delete with no matches
	deleted, err = repo.DeleteOlderThan(now.AddDate(-1, 0, 0)) // 1 year ago
	if err != nil {
		t.Fatalf("DeleteOlderThan() error = %v", err)
	}
	if deleted != 0 {
		t.Errorf("DeleteOlderThan() with no matches deleted %d logs, want 0", deleted)
	}
}

func TestDataChangeLogRepository_ListOrderByCreatedAt(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewDataChangeLogRepository(db, "sqlite3")

	// Create logs at different times
	now := time.Now()
	for i := 0; i < 5; i++ {
		timestamp := now.Add(time.Duration(-i) * time.Hour)
		_, err := db.Exec(`INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			domain.EntityTypeWOD, int64(i+1), "Test", domain.OperationUpdate, 1, "test@example.com", timestamp)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// List should return most recent first
	got, err := repo.List(domain.DataChangeLogFilters{}, 100, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Verify order (most recent first)
	for i := 1; i < len(got); i++ {
		if got[i].CreatedAt.After(got[i-1].CreatedAt) {
			t.Error("List() should return logs in descending order by created_at")
		}
	}
}

func TestDataChangeLogRepository_CombinedFilters(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewDataChangeLogRepository(db, "sqlite3")
	now := time.Now()

	// Create diverse test data
	testData := []struct {
		entityType string
		operation  string
		userID     int64
		userEmail  string
	}{
		{domain.EntityTypeWOD, domain.OperationUpdate, 1, "admin@example.com"},
		{domain.EntityTypeWOD, domain.OperationDelete, 1, "admin@example.com"},
		{domain.EntityTypeWOD, domain.OperationUpdate, 2, "user@example.com"},
		{domain.EntityTypeMovement, domain.OperationUpdate, 1, "admin@example.com"},
		{domain.EntityTypeMovement, domain.OperationDelete, 2, "user@example.com"},
	}

	for i, td := range testData {
		_, err := db.Exec(`INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			td.entityType, int64(i+1), "Test", td.operation, td.userID, td.userEmail, now)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Test combined filters: WOD + Update
	entityTypeWOD := domain.EntityTypeWOD
	operationUpdate := domain.OperationUpdate
	filters := domain.DataChangeLogFilters{
		EntityType: &entityTypeWOD,
		Operation:  &operationUpdate,
	}

	got, err := repo.List(filters, 100, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("List() with combined filters returned %d logs, want 2", len(got))
	}

	// Test combined filters: WOD + User 1
	userID1 := int64(1)
	filters = domain.DataChangeLogFilters{
		EntityType: &entityTypeWOD,
		UserID:     &userID1,
	}

	got, err = repo.List(filters, 100, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("List() with WOD + User 1 returned %d logs, want 2", len(got))
	}
}

func TestDataChangeLogRepository_Create_UnsupportedDriver(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	// Create repo with unsupported driver
	repo := NewDataChangeLogRepository(db, "unsupported_driver")

	log := &domain.DataChangeLog{
		EntityType: domain.EntityTypeWOD,
		EntityID:   1,
		EntityName: "Test",
		Operation:  domain.OperationUpdate,
		UserID:     1,
		UserEmail:  "test@example.com",
	}

	err = repo.Create(log)
	if err == nil {
		t.Error("Create() with unsupported driver should return error")
	}
	if err != nil && !containsString(err.Error(), "unsupported database driver") {
		t.Errorf("Create() error = %v, want error containing 'unsupported database driver'", err)
	}
}

func TestDataChangeLogRepository_Count_AllFilters(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewDataChangeLogRepository(db, "sqlite3")
	now := time.Now()

	// Create test data with specific timestamps for date filtering
	testData := []struct {
		entityType string
		entityID   int64
		operation  string
		userID     int64
		userEmail  string
		createdAt  time.Time
	}{
		{domain.EntityTypeWOD, 1, domain.OperationUpdate, 1, "admin@example.com", now.AddDate(0, 0, -10)},
		{domain.EntityTypeWOD, 1, domain.OperationUpdate, 1, "admin@example.com", now.AddDate(0, 0, -5)},
		{domain.EntityTypeWOD, 2, domain.OperationDelete, 2, "user@example.com", now.AddDate(0, 0, -3)},
		{domain.EntityTypeMovement, 3, domain.OperationUpdate, 1, "admin@example.com", now.AddDate(0, 0, -1)},
		{domain.EntityTypeMovement, 4, domain.OperationDelete, 2, "user@example.com", now},
	}

	for _, td := range testData {
		_, err := db.Exec(`INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			td.entityType, td.entityID, "Test", td.operation, td.userID, td.userEmail, td.createdAt)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	t.Run("Count by entity ID", func(t *testing.T) {
		entityID := int64(1)
		filters := domain.DataChangeLogFilters{EntityID: &entityID}
		count, err := repo.Count(filters)
		if err != nil {
			t.Fatalf("Count() error = %v", err)
		}
		if count != 2 {
			t.Errorf("Count() = %d, want 2", count)
		}
	})

	t.Run("Count by user email", func(t *testing.T) {
		userEmail := "admin"
		filters := domain.DataChangeLogFilters{UserEmail: &userEmail}
		count, err := repo.Count(filters)
		if err != nil {
			t.Fatalf("Count() error = %v", err)
		}
		if count != 3 {
			t.Errorf("Count() = %d, want 3", count)
		}
	})

	t.Run("Count by date range", func(t *testing.T) {
		startDate := now.AddDate(0, 0, -4)
		endDate := now.AddDate(0, 0, 1)
		filters := domain.DataChangeLogFilters{
			StartDate: &startDate,
			EndDate:   &endDate,
		}
		count, err := repo.Count(filters)
		if err != nil {
			t.Fatalf("Count() error = %v", err)
		}
		// Should get logs from -3, -1, and today (3 logs)
		if count != 3 {
			t.Errorf("Count() = %d, want 3", count)
		}
	})

	t.Run("Count with multiple filters", func(t *testing.T) {
		entityType := domain.EntityTypeWOD
		entityID := int64(1)
		userID := int64(1)
		filters := domain.DataChangeLogFilters{
			EntityType: &entityType,
			EntityID:   &entityID,
			UserID:     &userID,
		}
		count, err := repo.Count(filters)
		if err != nil {
			t.Fatalf("Count() error = %v", err)
		}
		if count != 2 {
			t.Errorf("Count() = %d, want 2", count)
		}
	})
}

// containsString is a helper to check if s contains substr
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
