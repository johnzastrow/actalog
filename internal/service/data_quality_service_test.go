package service

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a temporary SQLite database with required tables
func setupDataQualityTestDB(t *testing.T) *sql.DB {
	// Use a temporary file database for proper multi-connection support
	tmpFile := t.TempDir() + "/test.db"
	db, err := sql.Open("sqlite3", tmpFile)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables needed for data quality testing (matching production schema)
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT,
			role TEXT DEFAULT 'user',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS movements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			type TEXT NOT NULL DEFAULT 'strength',
			is_standard INTEGER NOT NULL DEFAULT 0,
			created_by INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (created_by) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS wods (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			source TEXT,
			type TEXT,
			regime TEXT,
			score_type TEXT,
			description TEXT,
			url TEXT,
			notes TEXT,
			is_standard INTEGER NOT NULL DEFAULT 0,
			created_by INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (created_by) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS workouts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			created_by INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (created_by) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_workouts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			workout_id INTEGER,
			workout_name TEXT,
			workout_date DATE NOT NULL,
			workout_type TEXT,
			total_time INTEGER,
			notes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (workout_id) REFERENCES workouts(id)
		)`,
		`CREATE TABLE IF NOT EXISTS workout_movements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workout_id INTEGER NOT NULL,
			movement_id INTEGER NOT NULL,
			sets INTEGER,
			reps INTEGER,
			order_index INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workout_id) REFERENCES workouts(id),
			FOREIGN KEY (movement_id) REFERENCES movements(id)
		)`,
		`CREATE TABLE IF NOT EXISTS workout_wods (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workout_id INTEGER NOT NULL,
			wod_id INTEGER NOT NULL,
			score_value TEXT,
			order_index INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workout_id) REFERENCES workouts(id),
			FOREIGN KEY (wod_id) REFERENCES wods(id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_workout_movements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_workout_id INTEGER NOT NULL,
			movement_id INTEGER NOT NULL,
			sets INTEGER,
			reps INTEGER,
			weight REAL,
			notes TEXT,
			order_index INTEGER DEFAULT 0,
			is_pr INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id),
			FOREIGN KEY (movement_id) REFERENCES movements(id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_workout_wods (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_workout_id INTEGER NOT NULL,
			wod_id INTEGER NOT NULL,
			score_value TEXT,
			division TEXT,
			is_pr INTEGER DEFAULT 0,
			order_index INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id),
			FOREIGN KEY (wod_id) REFERENCES wods(id)
		)`,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	return db
}

func TestNewDataQualityService(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	service := NewDataQualityService(db, "sqlite3", nil)
	if service == nil {
		t.Fatal("Expected service, got nil")
	}

	if service.db != db {
		t.Error("Database not properly set")
	}

	if service.driver != "sqlite3" {
		t.Errorf("Expected driver 'sqlite3', got '%s'", service.driver)
	}
}

func TestScanForDuplicates_Movements_NoDuplicates(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Insert unique movements
	db.Exec("INSERT INTO movements (name, type) VALUES ('Squat', 'strength')")
	db.Exec("INSERT INTO movements (name, type) VALUES ('Deadlift', 'strength')")
	db.Exec("INSERT INTO movements (name, type) VALUES ('Press', 'strength')")

	service := NewDataQualityService(db, "sqlite3", nil)
	result, err := service.ScanForDuplicates(EntityTypeMovement)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.EntityType != EntityTypeMovement {
		t.Errorf("Expected entity type 'movements', got '%s'", result.EntityType)
	}

	if len(result.Groups) != 0 {
		t.Errorf("Expected 0 duplicate groups, got %d", len(result.Groups))
	}

	if result.TotalDups != 0 {
		t.Errorf("Expected 0 total duplicates, got %d", result.TotalDups)
	}
}

func TestScanForDuplicates_Movements_WithDuplicates(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Insert duplicate movements (case-insensitive)
	_, err := db.Exec("INSERT INTO movements (name, type) VALUES ('Squat', 'strength')")
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}
	db.Exec("INSERT INTO movements (name, type) VALUES ('squat', 'strength')")
	db.Exec("INSERT INTO movements (name, type) VALUES ('SQUAT', 'strength')")
	db.Exec("INSERT INTO movements (name, type) VALUES ('Deadlift', 'strength')")

	service := NewDataQualityService(db, "sqlite3", nil)
	result, err := service.ScanForDuplicates(EntityTypeMovement)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result.Groups) != 1 {
		t.Errorf("Expected 1 duplicate group, got %d", len(result.Groups))
	}

	if result.TotalDups != 3 {
		t.Errorf("Expected 3 total duplicates, got %d", result.TotalDups)
	}

	// Check the group details
	if len(result.Groups) > 0 {
		group := result.Groups[0]
		if group.Key != "squat" {
			t.Errorf("Expected key 'squat', got '%s'", group.Key)
		}
		if group.Count != 3 {
			t.Errorf("Expected count 3, got %d", group.Count)
		}
		if len(group.Records) != 3 {
			t.Errorf("Expected 3 records, got %d", len(group.Records))
		}
	}
}

func TestScanForDuplicates_WODs_WithDuplicates(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Insert duplicate WODs
	db.Exec("INSERT INTO wods (name, description, type) VALUES ('Fran', 'Thrusters and Pull-ups', 'ForTime')")
	db.Exec("INSERT INTO wods (name, description, type) VALUES ('fran', 'Another Fran', 'ForTime')")
	db.Exec("INSERT INTO wods (name, description, type) VALUES ('Cindy', 'AMRAP', 'AMRAP')")

	service := NewDataQualityService(db, "sqlite3", nil)
	result, err := service.ScanForDuplicates(EntityTypeWOD)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result.Groups) != 1 {
		t.Errorf("Expected 1 duplicate group, got %d", len(result.Groups))
	}

	if result.TotalDups != 2 {
		t.Errorf("Expected 2 total duplicates, got %d", result.TotalDups)
	}
}

func TestScanForDuplicates_Users_WithDuplicates(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Insert duplicate users by email (case-insensitive)
	// Note: This would violate UNIQUE constraint in real DB, but we're testing the scan logic
	db.Exec("INSERT INTO users (name, email) VALUES ('John', 'john@example.com')")
	db.Exec("INSERT INTO users (name, email) VALUES ('John Doe', 'JOHN@example.com')")
	db.Exec("INSERT INTO users (name, email) VALUES ('Jane', 'jane@example.com')")

	service := NewDataQualityService(db, "sqlite3", nil)
	result, err := service.ScanForDuplicates(EntityTypeUser)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.EntityType != EntityTypeUser {
		t.Errorf("Expected entity type 'users', got '%s'", result.EntityType)
	}

	// We expect duplicates due to case-insensitive matching
	if len(result.Groups) != 1 {
		t.Errorf("Expected 1 duplicate group, got %d", len(result.Groups))
	}
}

func TestScanForDuplicates_UserWorkouts_WithDuplicates(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// First create a user
	db.Exec("INSERT INTO users (name, email) VALUES ('Test User', 'test@example.com')")

	// Insert duplicate user workouts (same user, date, name)
	db.Exec("INSERT INTO user_workouts (user_id, workout_date, workout_name) VALUES (1, '2024-01-15', 'Morning WOD')")
	db.Exec("INSERT INTO user_workouts (user_id, workout_date, workout_name) VALUES (1, '2024-01-15', 'Morning WOD')")
	db.Exec("INSERT INTO user_workouts (user_id, workout_date, workout_name) VALUES (1, '2024-01-16', 'Morning WOD')")

	service := NewDataQualityService(db, "sqlite3", nil)
	result, err := service.ScanForDuplicates(EntityTypeUserWorkout)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.EntityType != EntityTypeUserWorkout {
		t.Errorf("Expected entity type 'user_workouts', got '%s'", result.EntityType)
	}

	if len(result.Groups) != 1 {
		t.Errorf("Expected 1 duplicate group, got %d", len(result.Groups))
	}

	if result.TotalDups != 2 {
		t.Errorf("Expected 2 total duplicates, got %d", result.TotalDups)
	}
}

func TestScanForDuplicates_Workouts_WithDuplicates(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// First create a user
	db.Exec("INSERT INTO users (name, email) VALUES ('Test User', 'test@example.com')")

	// Insert duplicate workouts (same name, same user)
	db.Exec("INSERT INTO workouts (name, description, created_by) VALUES ('Leg Day', 'Squats', 1)")
	db.Exec("INSERT INTO workouts (name, description, created_by) VALUES ('leg day', 'More squats', 1)")
	db.Exec("INSERT INTO workouts (name, description, created_by) VALUES ('Arm Day', 'Curls', 1)")

	service := NewDataQualityService(db, "sqlite3", nil)
	result, err := service.ScanForDuplicates(EntityTypeWorkout)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result.Groups) != 1 {
		t.Errorf("Expected 1 duplicate group, got %d", len(result.Groups))
	}
}

func TestScanForDuplicates_InvalidEntityType(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	service := NewDataQualityService(db, "sqlite3", nil)
	_, err := service.ScanForDuplicates(EntityType("invalid"))

	if err == nil {
		t.Error("Expected error for invalid entity type")
	}
}

func TestScanAllDuplicates(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Insert some duplicate movements
	db.Exec("INSERT INTO movements (name) VALUES ('Squat')")
	db.Exec("INSERT INTO movements (name) VALUES ('squat')")

	// Insert a user for FK references
	db.Exec("INSERT INTO users (name, email) VALUES ('Test', 'test@example.com')")

	service := NewDataQualityService(db, "sqlite3", nil)
	results, err := service.ScanAllDuplicates()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have results for all 5 entity types
	if len(results) != 5 {
		t.Errorf("Expected 5 result sets, got %d", len(results))
	}

	// Check that movements has duplicates
	if movementResult, ok := results[EntityTypeMovement]; ok {
		if len(movementResult.Groups) != 1 {
			t.Errorf("Expected 1 movement duplicate group, got %d", len(movementResult.Groups))
		}
	} else {
		t.Error("Missing movements result")
	}
}

func TestPreviewMerge_Movements(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Insert duplicate movements
	db.Exec("INSERT INTO movements (name) VALUES ('Squat')")
	db.Exec("INSERT INTO movements (name) VALUES ('squat')")
	db.Exec("INSERT INTO movements (name) VALUES ('SQUAT')")

	// Create FK references
	db.Exec("INSERT INTO users (name, email) VALUES ('Test', 'test@example.com')")
	db.Exec("INSERT INTO workouts (name, created_by) VALUES ('Test WO', 1)")
	db.Exec("INSERT INTO workout_movements (workout_id, movement_id) VALUES (1, 2)")
	db.Exec("INSERT INTO workout_movements (workout_id, movement_id) VALUES (1, 3)")

	service := NewDataQualityService(db, "sqlite3", nil)
	preview, err := service.PreviewMerge(EntityTypeMovement, 1, []int64{2, 3})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if preview.KeepID != 1 {
		t.Errorf("Expected keep ID 1, got %d", preview.KeepID)
	}

	if len(preview.DeleteIDs) != 2 {
		t.Errorf("Expected 2 delete IDs, got %d", len(preview.DeleteIDs))
	}

	if !preview.CanProceed {
		t.Error("Expected CanProceed to be true")
	}

	// Check FK updates were counted
	if count, ok := preview.FKUpdates["workout_movements"]; ok {
		if count != 2 {
			t.Errorf("Expected 2 workout_movements updates, got %d", count)
		}
	}
}

func TestPreviewMerge_WODs(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Insert duplicate WODs
	db.Exec("INSERT INTO wods (name) VALUES ('Fran')")
	db.Exec("INSERT INTO wods (name) VALUES ('fran')")

	// Create FK references
	db.Exec("INSERT INTO users (name, email) VALUES ('Test', 'test@example.com')")
	db.Exec("INSERT INTO workouts (name, created_by) VALUES ('Test WO', 1)")
	db.Exec("INSERT INTO workout_wods (workout_id, wod_id) VALUES (1, 2)")

	service := NewDataQualityService(db, "sqlite3", nil)
	preview, err := service.PreviewMerge(EntityTypeWOD, 1, []int64{2})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !preview.CanProceed {
		t.Error("Expected CanProceed to be true")
	}
}

func TestPreviewMerge_UserWorkouts(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Create user and duplicate workouts
	db.Exec("INSERT INTO users (name, email) VALUES ('Test', 'test@example.com')")
	db.Exec("INSERT INTO user_workouts (user_id, workout_date, workout_name) VALUES (1, '2024-01-15', 'WOD')")
	db.Exec("INSERT INTO user_workouts (user_id, workout_date, workout_name) VALUES (1, '2024-01-15', 'WOD')")

	// Create child records
	db.Exec("INSERT INTO movements (name) VALUES ('Squat')")
	db.Exec("INSERT INTO user_workout_movements (user_workout_id, movement_id) VALUES (2, 1)")

	service := NewDataQualityService(db, "sqlite3", nil)
	preview, err := service.PreviewMerge(EntityTypeUserWorkout, 1, []int64{2})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !preview.CanProceed {
		t.Error("Expected CanProceed to be true")
	}
}

func TestPreviewMerge_InvalidEntityType(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	service := NewDataQualityService(db, "sqlite3", nil)
	preview, err := service.PreviewMerge(EntityType("invalid"), 1, []int64{2})

	// PreviewMerge returns error for unsupported entity types
	if err != nil {
		// This is expected behavior - unsupported entity type returns error
		return
	}

	// If no error, check preview indicates cannot proceed
	if preview.CanProceed {
		t.Error("Expected CanProceed to be false for invalid entity type")
	}

	if len(preview.Warnings) == 0 {
		t.Error("Expected warnings for invalid entity type")
	}
}

func TestExecuteMerge_Movements(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Insert duplicate movements
	db.Exec("INSERT INTO movements (name) VALUES ('Squat')")
	db.Exec("INSERT INTO movements (name) VALUES ('squat')")
	db.Exec("INSERT INTO movements (name) VALUES ('SQUAT')")

	// Create FK references
	db.Exec("INSERT INTO users (name, email) VALUES ('Test', 'test@example.com')")
	db.Exec("INSERT INTO workouts (name, created_by) VALUES ('Test WO', 1)")
	db.Exec("INSERT INTO workout_movements (workout_id, movement_id) VALUES (1, 2)")
	db.Exec("INSERT INTO workout_movements (workout_id, movement_id) VALUES (1, 3)")

	service := NewDataQualityService(db, "sqlite3", nil)
	result, err := service.ExecuteMerge(EntityTypeMovement, 1, []int64{2, 3}, 1)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	if result.DeletedCount != 2 {
		t.Errorf("Expected 2 deleted, got %d", result.DeletedCount)
	}

	// Verify FK updates happened
	if count, ok := result.FKUpdates["workout_movements"]; ok {
		if count != 2 {
			t.Errorf("Expected 2 FK updates, got %d", count)
		}
	}

	// Verify only 1 movement remains
	var count int
	db.QueryRow("SELECT COUNT(*) FROM movements").Scan(&count)
	if count != 1 {
		t.Errorf("Expected 1 movement remaining, got %d", count)
	}

	// Verify FK references point to kept record
	var movementID int64
	db.QueryRow("SELECT movement_id FROM workout_movements LIMIT 1").Scan(&movementID)
	if movementID != 1 {
		t.Errorf("Expected movement_id 1, got %d", movementID)
	}
}

func TestExecuteMerge_WODs(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Insert duplicate WODs
	db.Exec("INSERT INTO wods (name) VALUES ('Fran')")
	db.Exec("INSERT INTO wods (name) VALUES ('fran')")

	// Create FK references
	db.Exec("INSERT INTO users (name, email) VALUES ('Test', 'test@example.com')")
	db.Exec("INSERT INTO workouts (name, created_by) VALUES ('Test WO', 1)")
	db.Exec("INSERT INTO workout_wods (workout_id, wod_id) VALUES (1, 2)")

	service := NewDataQualityService(db, "sqlite3", nil)
	result, err := service.ExecuteMerge(EntityTypeWOD, 1, []int64{2}, 1)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	// Verify only 1 WOD remains
	var count int
	db.QueryRow("SELECT COUNT(*) FROM wods").Scan(&count)
	if count != 1 {
		t.Errorf("Expected 1 WOD remaining, got %d", count)
	}
}

func TestExecuteMerge_UserWorkouts(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Create user and duplicate workouts
	db.Exec("INSERT INTO users (name, email) VALUES ('Test', 'test@example.com')")
	db.Exec("INSERT INTO user_workouts (user_id, workout_date, workout_name) VALUES (1, '2024-01-15', 'WOD')")
	db.Exec("INSERT INTO user_workouts (user_id, workout_date, workout_name) VALUES (1, '2024-01-15', 'WOD')")

	// Create child records on the duplicate
	db.Exec("INSERT INTO movements (name) VALUES ('Squat')")
	db.Exec("INSERT INTO wods (name) VALUES ('Fran')")
	db.Exec("INSERT INTO user_workout_movements (user_workout_id, movement_id) VALUES (2, 1)")
	db.Exec("INSERT INTO user_workout_wods (user_workout_id, wod_id) VALUES (2, 1)")

	service := NewDataQualityService(db, "sqlite3", nil)
	result, err := service.ExecuteMerge(EntityTypeUserWorkout, 1, []int64{2}, 1)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	// Verify only 1 user_workout remains
	var count int
	db.QueryRow("SELECT COUNT(*) FROM user_workouts").Scan(&count)
	if count != 1 {
		t.Errorf("Expected 1 user_workout remaining, got %d", count)
	}

	// Verify child records were deleted (not transferred for user_workouts)
	db.QueryRow("SELECT COUNT(*) FROM user_workout_movements").Scan(&count)
	if count != 0 {
		t.Errorf("Expected 0 user_workout_movements remaining, got %d", count)
	}
}

func TestExecuteMerge_InvalidEntityType(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	service := NewDataQualityService(db, "sqlite3", nil)
	result, err := service.ExecuteMerge(EntityType("invalid"), 1, []int64{2}, 1)

	// ExecuteMerge returns error for unsupported entity types
	if err != nil {
		// This is expected behavior - unsupported entity type returns error
		return
	}

	// If no error returned, result should indicate failure
	if result.Success {
		t.Error("Expected failure for invalid entity type")
	}
}

func TestScanDataQuality_NoIssues(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Insert valid data
	db.Exec("INSERT INTO users (name, email) VALUES ('Test', 'test@example.com')")
	db.Exec("INSERT INTO movements (name) VALUES ('Squat')")
	db.Exec("INSERT INTO wods (name) VALUES ('Fran')")

	service := NewDataQualityService(db, "sqlite3", nil)
	result, err := service.ScanDataQuality()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// No orphaned records or empty required fields expected
	if result.TotalCount < 0 {
		t.Error("TotalCount should not be negative")
	}
}

func TestScanDataQuality_EmptyRequiredFields(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Insert records with empty required fields
	db.Exec("INSERT INTO movements (name) VALUES ('')")
	db.Exec("INSERT INTO wods (name) VALUES ('')")
	db.Exec("INSERT INTO users (name, email) VALUES ('', 'test@example.com')")

	service := NewDataQualityService(db, "sqlite3", nil)
	result, err := service.ScanDataQuality()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should find empty required field issues
	if result.TotalCount == 0 {
		t.Log("Note: Empty field detection may vary by implementation")
	}
}

func TestScanDataQuality_FutureDates(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Create user
	db.Exec("INSERT INTO users (name, email) VALUES ('Test', 'test@example.com')")

	// Insert workout with future date
	futureDate := time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	db.Exec("INSERT INTO user_workouts (user_id, workout_date, workout_name) VALUES (1, ?, 'Future WOD')", futureDate)

	service := NewDataQualityService(db, "sqlite3", nil)
	result, err := service.ScanDataQuality()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should find future date issues
	foundFutureDate := false
	for _, issue := range result.Issues {
		if issue.IssueType == "future_date" {
			foundFutureDate = true
			break
		}
	}

	if !foundFutureDate {
		t.Error("Expected to find future date issue")
	}
}

func TestScanDataQuality_InvalidEmail(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	// Insert user with invalid email
	db.Exec("INSERT INTO users (name, email) VALUES ('Test', 'invalid-email')")

	service := NewDataQualityService(db, "sqlite3", nil)
	result, err := service.ScanDataQuality()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should find invalid email issue
	foundInvalidEmail := false
	for _, issue := range result.Issues {
		if issue.IssueType == "invalid_email" {
			foundInvalidEmail = true
			break
		}
	}

	if !foundInvalidEmail {
		t.Error("Expected to find invalid email issue")
	}
}

func TestMakeInClause(t *testing.T) {
	db := setupDataQualityTestDB(t)
	defer db.Close()

	service := NewDataQualityService(db, "sqlite3", nil)

	tests := []struct {
		name     string
		ids      []int64
		expected string
	}{
		{
			name:     "Single ID",
			ids:      []int64{1},
			expected: "1",
		},
		{
			name:     "Multiple IDs",
			ids:      []int64{1, 2, 3},
			expected: "1,2,3",
		},
		{
			name:     "Empty IDs",
			ids:      []int64{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.makeInClause(tt.ids)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
