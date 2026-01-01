package service

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
	_ "github.com/mattn/go-sqlite3"
)

// TestRestoreModeConstants verifies the RestoreMode constants are defined correctly
func TestRestoreModeConstants(t *testing.T) {
	tests := []struct {
		name     string
		mode     domain.RestoreMode
		expected string
	}{
		{"Replace mode", domain.RestoreModeReplace, "replace"},
		{"Merge mode", domain.RestoreModeMerge, "merge"},
		{"Skip mode", domain.RestoreModeSkip, "skip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mode) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.mode))
			}
		})
	}
}

// TestIDMappings verifies the idMappings struct initialization
func TestIDMappings(t *testing.T) {
	mappings := newIDMappings()

	if mappings == nil {
		t.Fatal("expected non-nil mappings")
	}

	// Verify all maps are initialized
	if mappings.users == nil {
		t.Error("users map is nil")
	}
	if mappings.movements == nil {
		t.Error("movements map is nil")
	}
	if mappings.wods == nil {
		t.Error("wods map is nil")
	}
	if mappings.workouts == nil {
		t.Error("workouts map is nil")
	}
	if mappings.organizations == nil {
		t.Error("organizations map is nil")
	}
	if mappings.userWorkouts == nil {
		t.Error("userWorkouts map is nil")
	}
}

// TestGetInt64FromRow tests the getInt64FromRow helper function
func TestGetInt64FromRow(t *testing.T) {
	// Create a minimal BackupServiceImpl for testing helper methods
	svc := &BackupServiceImpl{}

	tests := []struct {
		name     string
		row      map[string]interface{}
		key      string
		expected int64
	}{
		{
			name:     "int64 value",
			row:      map[string]interface{}{"id": int64(42)},
			key:      "id",
			expected: 42,
		},
		{
			name:     "float64 value",
			row:      map[string]interface{}{"id": float64(42.0)},
			key:      "id",
			expected: 42,
		},
		{
			name:     "int value",
			row:      map[string]interface{}{"id": int(42)},
			key:      "id",
			expected: 42,
		},
		{
			name:     "missing key",
			row:      map[string]interface{}{"other": int64(42)},
			key:      "id",
			expected: 0,
		},
		{
			name:     "nil value",
			row:      map[string]interface{}{"id": nil},
			key:      "id",
			expected: 0,
		},
		{
			name:     "string value (unsupported)",
			row:      map[string]interface{}{"id": "42"},
			key:      "id",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.getInt64FromRow(tt.row, tt.key)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestRecordIDMapping tests the recordIDMapping function
func TestRecordIDMapping(t *testing.T) {
	svc := &BackupServiceImpl{}

	tests := []struct {
		name      string
		tableName string
		oldID     int64
		newID     int64
		checkMap  func(*idMappings) map[int64]int64
	}{
		{
			name:      "users mapping",
			tableName: "users",
			oldID:     1,
			newID:     100,
			checkMap:  func(m *idMappings) map[int64]int64 { return m.users },
		},
		{
			name:      "movements mapping",
			tableName: "movements",
			oldID:     2,
			newID:     200,
			checkMap:  func(m *idMappings) map[int64]int64 { return m.movements },
		},
		{
			name:      "wods mapping",
			tableName: "wods",
			oldID:     3,
			newID:     300,
			checkMap:  func(m *idMappings) map[int64]int64 { return m.wods },
		},
		{
			name:      "workouts mapping",
			tableName: "workouts",
			oldID:     4,
			newID:     400,
			checkMap:  func(m *idMappings) map[int64]int64 { return m.workouts },
		},
		{
			name:      "organizations mapping",
			tableName: "organizations",
			oldID:     5,
			newID:     500,
			checkMap:  func(m *idMappings) map[int64]int64 { return m.organizations },
		},
		{
			name:      "user_workouts mapping",
			tableName: "user_workouts",
			oldID:     6,
			newID:     600,
			checkMap:  func(m *idMappings) map[int64]int64 { return m.userWorkouts },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mappings := newIDMappings()
			svc.recordIDMapping(tt.tableName, tt.oldID, tt.newID, mappings)

			targetMap := tt.checkMap(mappings)
			if got, ok := targetMap[tt.oldID]; !ok || got != tt.newID {
				t.Errorf("expected mapping %d -> %d, got %d (found: %v)", tt.oldID, tt.newID, got, ok)
			}
		})
	}
}

// TestRecordIDMapping_ZeroOldID verifies that zero oldID is not recorded
func TestRecordIDMapping_ZeroOldID(t *testing.T) {
	svc := &BackupServiceImpl{}
	mappings := newIDMappings()

	svc.recordIDMapping("users", 0, 100, mappings)

	if len(mappings.users) != 0 {
		t.Errorf("expected empty users map, got %d entries", len(mappings.users))
	}
}

// TestRemapForeignKeys tests the foreign key remapping logic
func TestRemapForeignKeys(t *testing.T) {
	svc := &BackupServiceImpl{}

	tests := []struct {
		name      string
		tableName string
		row       map[string]interface{}
		mappings  *idMappings
		checkKey  string
		expected  int64
	}{
		{
			name:      "user_workouts user_id remapping",
			tableName: "user_workouts",
			row:       map[string]interface{}{"user_id": int64(1)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.users[1] = 100
				return m
			}(),
			checkKey: "user_id",
			expected: 100,
		},
		{
			name:      "workout_movements workout_id remapping",
			tableName: "workout_movements",
			row:       map[string]interface{}{"workout_id": int64(2)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.workouts[2] = 200
				return m
			}(),
			checkKey: "workout_id",
			expected: 200,
		},
		{
			name:      "workout_movements movement_id remapping",
			tableName: "workout_movements",
			row:       map[string]interface{}{"movement_id": int64(3)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.movements[3] = 300
				return m
			}(),
			checkKey: "movement_id",
			expected: 300,
		},
		{
			name:      "user_workout_movements user_workout_id remapping",
			tableName: "user_workout_movements",
			row:       map[string]interface{}{"user_workout_id": int64(4)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.userWorkouts[4] = 400
				return m
			}(),
			checkKey: "user_workout_id",
			expected: 400,
		},
		{
			name:      "user_organizations organization_id remapping",
			tableName: "user_organizations",
			row:       map[string]interface{}{"organization_id": int64(5)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.organizations[5] = 500
				return m
			}(),
			checkKey: "organization_id",
			expected: 500,
		},
		{
			name:      "created_by remapping",
			tableName: "movements",
			row:       map[string]interface{}{"created_by": int64(6)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.users[6] = 600
				return m
			}(),
			checkKey: "created_by",
			expected: 600,
		},
		{
			name:      "no remapping when not in mappings",
			tableName: "user_workouts",
			row:       map[string]interface{}{"user_id": int64(99)},
			mappings:  newIDMappings(), // Empty mappings
			checkKey:  "user_id",
			expected:  99, // Should remain unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc.remapForeignKeys(tt.tableName, tt.row, tt.mappings)

			got := svc.getInt64FromRow(tt.row, tt.checkKey)
			if got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

// TestBuildQuery tests PostgreSQL placeholder conversion
func TestBuildQuery(t *testing.T) {
	tests := []struct {
		name       string
		dbDriver   string
		query      string
		paramCount int
		expected   string
	}{
		{
			name:       "SQLite query unchanged",
			dbDriver:   "sqlite3",
			query:      "SELECT * FROM users WHERE id = ? AND email = ?",
			paramCount: 2,
			expected:   "SELECT * FROM users WHERE id = ? AND email = ?",
		},
		{
			name:       "MySQL query unchanged",
			dbDriver:   "mysql",
			query:      "SELECT * FROM users WHERE id = ? AND email = ?",
			paramCount: 2,
			expected:   "SELECT * FROM users WHERE id = ? AND email = ?",
		},
		{
			name:       "PostgreSQL query converted",
			dbDriver:   "postgres",
			query:      "SELECT * FROM users WHERE id = ? AND email = ?",
			paramCount: 2,
			expected:   "SELECT * FROM users WHERE id = $1 AND email = $2",
		},
		{
			name:       "PostgreSQL single param",
			dbDriver:   "postgres",
			query:      "SELECT id FROM users WHERE email = ?",
			paramCount: 1,
			expected:   "SELECT id FROM users WHERE email = $1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &BackupServiceImpl{dbDriver: tt.dbDriver}
			result := svc.buildQuery(tt.query, tt.paramCount)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestContainsString tests the containsString helper
func TestContainsString(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		str      string
		expected bool
	}{
		{
			name:     "string found",
			slice:    []string{"a", "b", "c"},
			str:      "b",
			expected: true,
		},
		{
			name:     "string not found",
			slice:    []string{"a", "b", "c"},
			str:      "d",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			str:      "a",
			expected: false,
		},
		{
			name:     "empty string search",
			slice:    []string{"a", "b", ""},
			str:      "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsString(tt.slice, tt.str)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestJoinStrings tests the joinStrings helper
func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name     string
		strs     []string
		sep      string
		expected string
	}{
		{
			name:     "join with comma",
			strs:     []string{"a", "b", "c"},
			sep:      ", ",
			expected: "a, b, c",
		},
		{
			name:     "single string",
			strs:     []string{"only"},
			sep:      ", ",
			expected: "only",
		},
		{
			name:     "empty slice",
			strs:     []string{},
			sep:      ", ",
			expected: "",
		},
		{
			name:     "join with AND",
			strs:     []string{"x = 1", "y = 2"},
			sep:      " AND ",
			expected: "x = 1 AND y = 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinStrings(tt.strs, tt.sep)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestIsUploadFile tests the isUploadFile helper
func TestIsUploadFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "valid upload file",
			filename: "uploads/profile.jpg",
			expected: true,
		},
		{
			name:     "upload file in subdirectory",
			filename: "uploads/images/avatar.png",
			expected: true,
		},
		{
			name:     "non-upload file",
			filename: "backup_data.json",
			expected: false,
		},
		{
			name:     "short filename",
			filename: "uploads",
			expected: false,
		},
		{
			name:     "similar prefix but not uploads/",
			filename: "uploaded/file.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUploadFile(tt.filename)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestNormalizeColumnType tests the column type normalization
func TestNormalizeColumnType(t *testing.T) {
	svc := &BackupServiceImpl{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Integer types
		{"INTEGER", "INTEGER", "integer"},
		{"int", "int", "integer"},
		{"BIGINT", "BIGINT", "integer"},
		{"smallint", "smallint", "integer"},
		{"serial", "serial", "integer"},

		// Boolean types
		{"boolean", "boolean", "boolean"},
		{"BOOL", "BOOL", "boolean"},
		// Note: tinyint(1) is often used as boolean in MySQL but contains "int" so matches integer first
		{"tinyint(1)", "tinyint(1)", "integer"},

		// Float types
		{"float", "float", "float"},
		{"double", "double", "float"},
		{"decimal", "decimal", "float"},
		{"numeric", "numeric", "float"},
		{"real", "real", "float"},

		// Date types
		{"date", "date", "date"},

		// Datetime types
		{"datetime", "datetime", "datetime"},
		{"timestamp", "timestamp", "datetime"},
		{"TIMESTAMP", "TIMESTAMP", "datetime"},

		// String types
		{"varchar(255)", "varchar(255)", "string"},
		{"text", "text", "string"},
		{"char(10)", "char(10)", "string"},
		{"TEXT", "TEXT", "string"},

		// Unknown defaults to string
		{"unknown_type", "unknown_type", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.normalizeColumnType(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeColumnType(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestConvertDatetimeForMySQL tests MySQL datetime format conversion
func TestConvertDatetimeForMySQL(t *testing.T) {
	svc := &BackupServiceImpl{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "RFC3339 format",
			input:    "2025-01-15T10:30:00Z",
			expected: "2025-01-15 10:30:00",
		},
		{
			name:     "RFC3339Nano format",
			input:    "2025-01-15T10:30:00.123456789Z",
			expected: "2025-01-15 10:30:00",
		},
		{
			name:     "Date only",
			input:    "2025-01-15",
			expected: "2025-01-15",
		},
		{
			name:     "MySQL format unchanged",
			input:    "2025-01-15 10:30:00",
			expected: "2025-01-15 10:30:00",
		},
		{
			name:     "Invalid format returns as-is",
			input:    "invalid-date",
			expected: "invalid-date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.convertDatetimeForMySQL(tt.input)
			if result != tt.expected {
				t.Errorf("convertDatetimeForMySQL(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestConvertValue tests value conversion for different database drivers
func TestConvertValue(t *testing.T) {
	tests := []struct {
		name       string
		dbDriver   string
		value      interface{}
		columnName string
		colType    string
		expected   interface{}
	}{
		{
			name:       "nil value",
			dbDriver:   "sqlite3",
			value:      nil,
			columnName: "is_pr",
			colType:    "boolean",
			expected:   nil,
		},
		{
			name:       "boolean true to SQLite",
			dbDriver:   "sqlite3",
			value:      true,
			columnName: "is_pr",
			colType:    "boolean",
			expected:   int64(1),
		},
		{
			name:       "boolean false to SQLite",
			dbDriver:   "sqlite3",
			value:      false,
			columnName: "is_pr",
			colType:    "boolean",
			expected:   int64(0),
		},
		{
			name:       "int64 1 to PostgreSQL boolean",
			dbDriver:   "postgres",
			value:      int64(1),
			columnName: "is_pr",
			colType:    "boolean",
			expected:   true,
		},
		{
			name:       "int64 0 to PostgreSQL boolean",
			dbDriver:   "postgres",
			value:      int64(0),
			columnName: "is_pr",
			colType:    "boolean",
			expected:   false,
		},
		{
			name:       "float64 1 to PostgreSQL boolean",
			dbDriver:   "postgres",
			value:      float64(1),
			columnName: "is_pr",
			colType:    "boolean",
			expected:   true,
		},
		{
			name:       "non-boolean value unchanged",
			dbDriver:   "sqlite3",
			value:      "test string",
			columnName: "name",
			colType:    "string",
			expected:   "test string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &BackupServiceImpl{dbDriver: tt.dbDriver}
			result := svc.convertValue(tt.value, tt.columnName, tt.colType)
			if result != tt.expected {
				t.Errorf("convertValue(%v, %q, %q) = %v (%T), expected %v (%T)",
					tt.value, tt.columnName, tt.colType, result, result, tt.expected, tt.expected)
			}
		})
	}
}

// TestRestoreResult tests RestoreResult struct initialization
func TestRestoreResult(t *testing.T) {
	result := &domain.RestoreResult{
		Mode:           domain.RestoreModeMerge,
		TablesRestored: 10,
		RecordsCreated: 100,
		RecordsUpdated: 50,
		RecordsSkipped: 25,
		Errors:         []string{"error 1", "error 2"},
	}

	if result.Mode != domain.RestoreModeMerge {
		t.Errorf("expected mode %q, got %q", domain.RestoreModeMerge, result.Mode)
	}
	if result.TablesRestored != 10 {
		t.Errorf("expected 10 tables, got %d", result.TablesRestored)
	}
	if result.RecordsCreated != 100 {
		t.Errorf("expected 100 created, got %d", result.RecordsCreated)
	}
	if result.RecordsUpdated != 50 {
		t.Errorf("expected 50 updated, got %d", result.RecordsUpdated)
	}
	if result.RecordsSkipped != 25 {
		t.Errorf("expected 25 skipped, got %d", result.RecordsSkipped)
	}
	if len(result.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(result.Errors))
	}
}

// Integration test helpers

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create minimal schema for testing
	schema := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS movements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			type TEXT NOT NULL,
			is_standard INTEGER NOT NULL DEFAULT 0,
			created_by INTEGER,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
		);

		CREATE TABLE IF NOT EXISTS wods (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			score_type TEXT,
			is_standard INTEGER NOT NULL DEFAULT 0,
			created_by INTEGER,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
		);

		CREATE TABLE IF NOT EXISTS workouts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			notes TEXT,
			created_by INTEGER,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
		);

		CREATE TABLE IF NOT EXISTS organizations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS user_workouts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			workout_id INTEGER,
			workout_name TEXT,
			workout_date DATE NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE RESTRICT
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

// TestBackupServiceImpl_TableExists tests table existence checking
func TestBackupServiceImpl_TableExists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := &BackupServiceImpl{
		db:       db,
		dbDriver: "sqlite3",
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	tests := []struct {
		name      string
		tableName string
		expected  bool
	}{
		{"users table exists", "users", true},
		{"movements table exists", "movements", true},
		{"nonexistent table", "nonexistent_table", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := svc.tableExists(tx, tt.tableName)
			if err != nil {
				t.Fatalf("tableExists error: %v", err)
			}
			if exists != tt.expected {
				t.Errorf("tableExists(%q) = %v, expected %v", tt.tableName, exists, tt.expected)
			}
		})
	}
}

// TestBackupServiceImpl_GetTableColumns tests column listing
func TestBackupServiceImpl_GetTableColumns(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := &BackupServiceImpl{
		db:       db,
		dbDriver: "sqlite3",
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	columns, err := svc.getTableColumns(tx, "users")
	if err != nil {
		t.Fatalf("getTableColumns error: %v", err)
	}

	expectedColumns := []string{"id", "email", "password_hash", "name", "role", "created_at", "updated_at"}
	if len(columns) != len(expectedColumns) {
		t.Errorf("expected %d columns, got %d", len(expectedColumns), len(columns))
	}

	for _, expected := range expectedColumns {
		found := false
		for _, col := range columns {
			if col == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected column %q not found in %v", expected, columns)
		}
	}
}

// TestBackupServiceImpl_FindByNaturalKey tests natural key matching
func TestBackupServiceImpl_FindByNaturalKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('test@example.com', 'hash', 'Test User', 'user', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO movements (name, description, type, is_standard, created_at, updated_at)
		VALUES ('Back Squat', 'Barbell back squat', 'strength', 1, datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test movement: %v", err)
	}

	svc := &BackupServiceImpl{
		db:       db,
		dbDriver: "sqlite3",
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	mappings := newIDMappings()

	tests := []struct {
		name        string
		tableName   string
		row         map[string]interface{}
		expectFound bool
	}{
		{
			name:        "find user by email",
			tableName:   "users",
			row:         map[string]interface{}{"email": "test@example.com"},
			expectFound: true,
		},
		{
			name:        "user not found",
			tableName:   "users",
			row:         map[string]interface{}{"email": "notfound@example.com"},
			expectFound: false,
		},
		{
			name:        "find movement by name",
			tableName:   "movements",
			row:         map[string]interface{}{"name": "Back Squat"},
			expectFound: true,
		},
		{
			name:        "movement not found",
			tableName:   "movements",
			row:         map[string]interface{}{"name": "Nonexistent Movement"},
			expectFound: false,
		},
		{
			name:        "empty email returns not found",
			tableName:   "users",
			row:         map[string]interface{}{"email": ""},
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, found, err := svc.findByNaturalKey(tx, tt.tableName, tt.row, mappings)
			if err != nil {
				t.Fatalf("findByNaturalKey error: %v", err)
			}
			if found != tt.expectFound {
				t.Errorf("findByNaturalKey found = %v, expected %v (id=%d)", found, tt.expectFound, id)
			}
			if found && id == 0 {
				t.Error("found = true but id = 0")
			}
		})
	}
}

// TestIsTableNotExistsError tests error detection for missing tables
func TestIsTableNotExistsError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "SQLite no such table",
			err:      sql.ErrNoRows, // This won't match, but let's test the function
			expected: false,
		},
		{
			name:     "SQLite no such table error string",
			err:      fmt.Errorf("no such table: users"),
			expected: true,
		},
		{
			name:     "PostgreSQL does not exist error",
			err:      fmt.Errorf("relation \"users\" does not exist"),
			expected: true,
		},
		{
			name:     "MySQL error code 1146",
			err:      fmt.Errorf("Error 1146 (42S02): Table 'db.users' doesn't exist"),
			expected: true,
		},
		{
			name:     "MySQL doesn't exist pattern",
			err:      fmt.Errorf("Table 'test.users' doesn't exist"),
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      fmt.Errorf("connection refused"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTableNotExistsError(tt.err)
			if result != tt.expected {
				t.Errorf("isTableNotExistsError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

// TestNewBackupService tests backup service creation
func TestNewBackupService(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := &mockUserRepo{users: make(map[int64]*domain.User)}
	auditLogRepo := &mockAuditLogRepo{}

	svc := NewBackupService(db, "sqlite3", "test.db", "/tmp/backups", "/tmp/uploads", userRepo, auditLogRepo)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.db != db {
		t.Error("db not set correctly")
	}
	if svc.dbDriver != "sqlite3" {
		t.Errorf("expected dbDriver 'sqlite3', got %s", svc.dbDriver)
	}
	if svc.dbName != "test.db" {
		t.Errorf("expected dbName 'test.db', got %s", svc.dbName)
	}
	if svc.backupDir != "/tmp/backups" {
		t.Errorf("expected backupDir '/tmp/backups', got %s", svc.backupDir)
	}
	if svc.uploadsDir != "/tmp/uploads" {
		t.Errorf("expected uploadsDir '/tmp/uploads', got %s", svc.uploadsDir)
	}
}

// TestBackupService_ListBackups_EmptyDir tests listing backups from empty directory
func TestBackupService_ListBackups_EmptyDir(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "backup_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	svc := &BackupServiceImpl{
		db:        db,
		dbDriver:  "sqlite3",
		backupDir: tmpDir,
	}

	backups, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups error: %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("expected 0 backups, got %d", len(backups))
	}
}

// TestBackupService_ListBackups_NonExistentDir tests listing backups from non-existent directory
func TestBackupService_ListBackups_NonExistentDir(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := &BackupServiceImpl{
		db:        db,
		dbDriver:  "sqlite3",
		backupDir: "/nonexistent/path/that/does/not/exist",
	}

	backups, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups should return empty list for non-existent dir, got error: %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("expected 0 backups, got %d", len(backups))
	}
}

// TestBackupService_DownloadBackup_NotFound tests downloading non-existent backup
func TestBackupService_DownloadBackup_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tmpDir, err := os.MkdirTemp("", "backup_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	svc := &BackupServiceImpl{
		db:        db,
		dbDriver:  "sqlite3",
		backupDir: tmpDir,
	}

	_, err = svc.DownloadBackup("nonexistent.zip")
	if err == nil {
		t.Error("expected error for non-existent backup")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

// TestBackupService_DownloadBackup_Exists tests downloading existing backup
func TestBackupService_DownloadBackup_Exists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tmpDir, err := os.MkdirTemp("", "backup_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFilePath := filepath.Join(tmpDir, "test_backup.zip")
	if err := os.WriteFile(testFilePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	svc := &BackupServiceImpl{
		db:        db,
		dbDriver:  "sqlite3",
		backupDir: tmpDir,
	}

	path, err := svc.DownloadBackup("test_backup.zip")
	if err != nil {
		t.Fatalf("DownloadBackup error: %v", err)
	}
	if path != testFilePath {
		t.Errorf("expected path %s, got %s", testFilePath, path)
	}
}

// TestBackupService_DeleteBackup_NotFound tests deleting non-existent backup
func TestBackupService_DeleteBackup_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tmpDir, err := os.MkdirTemp("", "backup_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	svc := &BackupServiceImpl{
		db:           db,
		dbDriver:     "sqlite3",
		backupDir:    tmpDir,
		auditLogRepo: &mockAuditLogRepo{},
	}

	err = svc.DeleteBackup("nonexistent.zip", 1)
	if err == nil {
		t.Error("expected error for non-existent backup")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

// TestBackupService_DeleteBackup_Success tests successful backup deletion
func TestBackupService_DeleteBackup_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tmpDir, err := os.MkdirTemp("", "backup_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFilePath := filepath.Join(tmpDir, "test_backup.zip")
	if err := os.WriteFile(testFilePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	svc := &BackupServiceImpl{
		db:           db,
		dbDriver:     "sqlite3",
		backupDir:    tmpDir,
		auditLogRepo: &mockAuditLogRepo{},
	}

	err = svc.DeleteBackup("test_backup.zip", 1)
	if err != nil {
		t.Fatalf("DeleteBackup error: %v", err)
	}

	// Verify file is deleted
	if _, err := os.Stat(testFilePath); !os.IsNotExist(err) {
		t.Error("expected file to be deleted")
	}
}

// TestBackupService_RowsToMaps tests SQL rows to maps conversion
func TestBackupService_RowsToMaps(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('test1@example.com', 'hash1', 'Test User 1', 'user', datetime('now'), datetime('now')),
		       ('test2@example.com', 'hash2', 'Test User 2', 'admin', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test users: %v", err)
	}

	svc := &BackupServiceImpl{
		db:       db,
		dbDriver: "sqlite3",
	}

	rows, err := db.Query("SELECT id, email, name, role FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("failed to query users: %v", err)
	}
	defer rows.Close()

	maps, err := svc.rowsToMaps(rows)
	if err != nil {
		t.Fatalf("rowsToMaps error: %v", err)
	}

	if len(maps) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(maps))
	}

	// Check first row
	if maps[0]["email"] != "test1@example.com" {
		t.Errorf("expected email 'test1@example.com', got %v", maps[0]["email"])
	}
	if maps[0]["name"] != "Test User 1" {
		t.Errorf("expected name 'Test User 1', got %v", maps[0]["name"])
	}
}

// TestBackupService_FindByNaturalKey_AllTables tests natural key matching for all supported tables
func TestBackupService_FindByNaturalKey_AllTables(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('test@example.com', 'hash', 'Test User', 'user', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO wods (name, score_type, is_standard, created_at, updated_at)
		VALUES ('Fran', 'time', 1, datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test wod: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO workouts (name, created_at, updated_at)
		VALUES ('Test Workout', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test workout: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO organizations (name, created_at, updated_at)
		VALUES ('Test Org', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test organization: %v", err)
	}

	svc := &BackupServiceImpl{
		db:       db,
		dbDriver: "sqlite3",
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	mappings := newIDMappings()

	tests := []struct {
		name        string
		tableName   string
		row         map[string]interface{}
		expectFound bool
	}{
		{
			name:        "find wod by name",
			tableName:   "wods",
			row:         map[string]interface{}{"name": "Fran"},
			expectFound: true,
		},
		{
			name:        "wod not found",
			tableName:   "wods",
			row:         map[string]interface{}{"name": "Murph"},
			expectFound: false,
		},
		{
			name:        "empty wod name returns not found",
			tableName:   "wods",
			row:         map[string]interface{}{"name": ""},
			expectFound: false,
		},
		{
			name:        "find workout by name",
			tableName:   "workouts",
			row:         map[string]interface{}{"name": "Test Workout"},
			expectFound: true,
		},
		{
			name:        "workout not found",
			tableName:   "workouts",
			row:         map[string]interface{}{"name": "Other Workout"},
			expectFound: false,
		},
		{
			name:        "empty workout name returns not found",
			tableName:   "workouts",
			row:         map[string]interface{}{"name": ""},
			expectFound: false,
		},
		{
			name:        "find organization by name",
			tableName:   "organizations",
			row:         map[string]interface{}{"name": "Test Org"},
			expectFound: true,
		},
		{
			name:        "organization not found",
			tableName:   "organizations",
			row:         map[string]interface{}{"name": "Other Org"},
			expectFound: false,
		},
		{
			name:        "empty organization name returns not found",
			tableName:   "organizations",
			row:         map[string]interface{}{"name": ""},
			expectFound: false,
		},
		{
			name:        "unknown table returns not found",
			tableName:   "unknown_table",
			row:         map[string]interface{}{"name": "test"},
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, found, err := svc.findByNaturalKey(tx, tt.tableName, tt.row, mappings)
			if err != nil {
				t.Fatalf("findByNaturalKey error: %v", err)
			}
			if found != tt.expectFound {
				t.Errorf("findByNaturalKey found = %v, expected %v (id=%d)", found, tt.expectFound, id)
			}
		})
	}
}

// TestRemapForeignKeys_AllCases tests all remapping cases
func TestRemapForeignKeys_AllCases(t *testing.T) {
	svc := &BackupServiceImpl{}

	tests := []struct {
		name       string
		tableName  string
		row        map[string]interface{}
		mappings   *idMappings
		checkKey   string
		expected   int64
		shouldHave bool
	}{
		{
			name:      "user_settings user_id remapping",
			tableName: "user_settings",
			row:       map[string]interface{}{"user_id": int64(1)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.users[1] = 100
				return m
			}(),
			checkKey:   "user_id",
			expected:   100,
			shouldHave: true,
		},
		{
			name:      "notifications user_id remapping",
			tableName: "notifications",
			row:       map[string]interface{}{"user_id": int64(2)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.users[2] = 200
				return m
			}(),
			checkKey:   "user_id",
			expected:   200,
			shouldHave: true,
		},
		{
			name:      "workout_wods wod_id remapping",
			tableName: "workout_wods",
			row:       map[string]interface{}{"wod_id": int64(3)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.wods[3] = 300
				return m
			}(),
			checkKey:   "wod_id",
			expected:   300,
			shouldHave: true,
		},
		{
			name:      "user_workout_wods user_workout_id remapping",
			tableName: "user_workout_wods",
			row:       map[string]interface{}{"user_workout_id": int64(4)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.userWorkouts[4] = 400
				return m
			}(),
			checkKey:   "user_workout_id",
			expected:   400,
			shouldHave: true,
		},
		{
			name:      "user_workout_wods wod_id remapping",
			tableName: "user_workout_wods",
			row:       map[string]interface{}{"wod_id": int64(5)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.wods[5] = 500
				return m
			}(),
			checkKey:   "wod_id",
			expected:   500,
			shouldHave: true,
		},
		{
			name:      "organization_subscriptions organization_id remapping",
			tableName: "organization_subscriptions",
			row:       map[string]interface{}{"organization_id": int64(6)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.organizations[6] = 600
				return m
			}(),
			checkKey:   "organization_id",
			expected:   600,
			shouldHave: true,
		},
		{
			name:      "notification_likes user_id remapping",
			tableName: "notification_likes",
			row:       map[string]interface{}{"user_id": int64(7)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.users[7] = 700
				return m
			}(),
			checkKey:   "user_id",
			expected:   700,
			shouldHave: true,
		},
		{
			name:      "user_subscriptions user_id remapping",
			tableName: "user_subscriptions",
			row:       map[string]interface{}{"user_id": int64(8)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.users[8] = 800
				return m
			}(),
			checkKey:   "user_id",
			expected:   800,
			shouldHave: true,
		},
		{
			name:      "user_workout_movements movement_id remapping",
			tableName: "user_workout_movements",
			row:       map[string]interface{}{"movement_id": int64(9)},
			mappings: func() *idMappings {
				m := newIDMappings()
				m.movements[9] = 900
				return m
			}(),
			checkKey:   "movement_id",
			expected:   900,
			shouldHave: true,
		},
		{
			name:       "no mapping for zero ID",
			tableName:  "user_workouts",
			row:        map[string]interface{}{"user_id": int64(0)},
			mappings:   newIDMappings(),
			checkKey:   "user_id",
			expected:   0,
			shouldHave: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc.remapForeignKeys(tt.tableName, tt.row, tt.mappings)

			got := svc.getInt64FromRow(tt.row, tt.checkKey)
			if got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

// TestBackupService_ResetSequence_NonPostgres tests resetSequence for non-postgres drivers
func TestBackupService_ResetSequence_NonPostgres(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	tests := []struct {
		name     string
		dbDriver string
	}{
		{"sqlite3 driver", "sqlite3"},
		{"mysql driver", "mysql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &BackupServiceImpl{dbDriver: tt.dbDriver}
			err := svc.resetSequence(tx, "users")
			if err != nil {
				t.Errorf("resetSequence should not error for %s: %v", tt.dbDriver, err)
			}
		})
	}
}

// TestBackupService_ConvertValue_DatetimeMySQL tests datetime conversion for MySQL
func TestBackupService_ConvertValue_DatetimeMySQL(t *testing.T) {
	svc := &BackupServiceImpl{dbDriver: "mysql"}

	tests := []struct {
		name       string
		value      interface{}
		columnName string
		colType    string
		expected   interface{}
	}{
		{
			name:       "RFC3339 datetime",
			value:      "2025-01-15T10:30:00Z",
			columnName: "created_at",
			colType:    "datetime",
			expected:   "2025-01-15 10:30:00",
		},
		{
			name:       "date only column",
			value:      "2025-01-15",
			columnName: "birthday",
			colType:    "date",
			expected:   "2025-01-15",
		},
		{
			name:       "nil datetime",
			value:      nil,
			columnName: "updated_at",
			colType:    "datetime",
			expected:   nil,
		},
		{
			name:       "empty datetime string",
			value:      "",
			columnName: "locked_at",
			colType:    "datetime",
			expected:   "", // Empty string should remain unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.convertValue(tt.value, tt.columnName, tt.colType)
			if result != tt.expected {
				t.Errorf("convertValue(%v, %q, %q) = %v, expected %v", tt.value, tt.columnName, tt.colType, result, tt.expected)
			}
		})
	}
}

// TestBackupService_ConvertValue_BooleanFallback tests boolean conversion with fallback column names
func TestBackupService_ConvertValue_BooleanFallback(t *testing.T) {
	tests := []struct {
		name       string
		dbDriver   string
		value      interface{}
		columnName string
		colType    string // Empty to test fallback
		expected   interface{}
	}{
		{
			name:       "is_rx fallback to SQLite",
			dbDriver:   "sqlite3",
			value:      true,
			columnName: "is_rx",
			colType:    "", // No schema type, use column name fallback
			expected:   int64(1),
		},
		{
			name:       "is_template fallback to PostgreSQL",
			dbDriver:   "postgres",
			value:      int64(1),
			columnName: "is_template",
			colType:    "", // No schema type, use column name fallback
			expected:   true,
		},
		{
			name:       "email_verified fallback to MySQL",
			dbDriver:   "mysql",
			value:      true,
			columnName: "email_verified",
			colType:    "",
			expected:   int64(1),
		},
		{
			name:       "account_disabled fallback",
			dbDriver:   "sqlite3",
			value:      false,
			columnName: "account_disabled",
			colType:    "",
			expected:   int64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &BackupServiceImpl{dbDriver: tt.dbDriver}
			result := svc.convertValue(tt.value, tt.columnName, tt.colType)
			if result != tt.expected {
				t.Errorf("convertValue(%v, %q, %q) = %v (%T), expected %v (%T)",
					tt.value, tt.columnName, tt.colType, result, result, tt.expected, tt.expected)
			}
		})
	}
}

// TestBackupService_ConvertValue_DatetimeFallback tests datetime conversion with fallback column names
func TestBackupService_ConvertValue_DatetimeFallback(t *testing.T) {
	svc := &BackupServiceImpl{dbDriver: "mysql"}

	tests := []struct {
		name       string
		value      interface{}
		columnName string
		expected   interface{}
	}{
		{
			name:       "created_at fallback",
			value:      "2025-01-15T10:30:00Z",
			columnName: "created_at",
			expected:   "2025-01-15 10:30:00",
		},
		{
			name:       "updated_at fallback",
			value:      "2025-01-15T10:30:00Z",
			columnName: "updated_at",
			expected:   "2025-01-15 10:30:00",
		},
		{
			name:       "last_login_at fallback",
			value:      "2025-01-15T10:30:00Z",
			columnName: "last_login_at",
			expected:   "2025-01-15 10:30:00",
		},
		{
			name:       "workout_date fallback",
			value:      "2025-01-15",
			columnName: "workout_date",
			expected:   "2025-01-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.convertValue(tt.value, tt.columnName, "")
			if result != tt.expected {
				t.Errorf("convertValue(%v, %q, \"\") = %v, expected %v", tt.value, tt.columnName, result, tt.expected)
			}
		})
	}
}

// TestBackupService_TableExists_UnsupportedDriver tests unsupported driver error
func TestBackupService_TableExists_UnsupportedDriver(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	svc := &BackupServiceImpl{
		db:       db,
		dbDriver: "unsupported",
	}

	_, err = svc.tableExists(tx, "users")
	if err == nil {
		t.Error("expected error for unsupported driver")
	}
	if !strings.Contains(err.Error(), "unsupported database driver") {
		t.Errorf("expected 'unsupported database driver' error, got: %v", err)
	}
}

// TestBackupService_GetTableColumns_UnsupportedDriver tests unsupported driver error
func TestBackupService_GetTableColumns_UnsupportedDriver(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	svc := &BackupServiceImpl{
		db:       db,
		dbDriver: "unsupported",
	}

	_, err = svc.getTableColumns(tx, "users")
	if err == nil {
		t.Error("expected error for unsupported driver")
	}
	if !strings.Contains(err.Error(), "unsupported database driver") {
		t.Errorf("expected 'unsupported database driver' error, got: %v", err)
	}
}

// mockAuditLogRepo is defined in test_helpers.go
