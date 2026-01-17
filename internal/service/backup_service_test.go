package service

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
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

// TestBackupService_GetTableSchema tests schema retrieval for tables
func TestBackupService_GetTableSchema(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := &BackupServiceImpl{
		db:       db,
		dbDriver: "sqlite3",
	}

	schema, err := svc.getTableSchema("users")
	if err != nil {
		t.Fatalf("getTableSchema error: %v", err)
	}

	// Verify columns are returned
	if len(schema.Columns) == 0 {
		t.Error("expected columns in schema")
	}

	// Check for expected columns
	expectedColumns := map[string]bool{
		"id":    false,
		"email": false,
		"name":  false,
		"role":  false,
	}

	for _, col := range schema.Columns {
		if _, ok := expectedColumns[col.Name]; ok {
			expectedColumns[col.Name] = true
		}
	}

	for col, found := range expectedColumns {
		if !found {
			t.Errorf("expected column %q not found in schema", col)
		}
	}
}

// TestBackupService_GetTableSchema_UnsupportedDriver tests unsupported driver error
func TestBackupService_GetTableSchema_UnsupportedDriver(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := &BackupServiceImpl{
		db:       db,
		dbDriver: "unsupported_driver",
	}

	_, err := svc.getTableSchema("users")
	if err == nil {
		t.Error("expected error for unsupported driver")
	}
	if !strings.Contains(err.Error(), "unsupported database driver") {
		t.Errorf("expected unsupported driver error, got: %v", err)
	}
}

// TestBackupService_CreateSchemaFromData tests schema generation from data
func TestBackupService_CreateSchemaFromData(t *testing.T) {
	svc := &BackupServiceImpl{}

	backupData := &domain.BackupData{
		Users: []map[string]interface{}{
			{"id": int64(1), "email": "test@example.com"},
		},
	}

	schema, err := svc.createSchemaFromData(backupData)
	if err != nil {
		t.Fatalf("createSchemaFromData error: %v", err)
	}

	// Verify schema contains CREATE TABLE statements
	if !strings.Contains(schema, "CREATE TABLE") {
		t.Error("expected CREATE TABLE in schema")
	}

	// Verify key tables are present
	requiredTables := []string{"users", "movements", "workouts", "wods", "user_workouts", "notifications"}
	for _, table := range requiredTables {
		if !strings.Contains(schema, table) {
			t.Errorf("expected table %q in schema", table)
		}
	}
}

// TestBackupService_RestoreTable tests basic table restore
func TestBackupService_RestoreTable(t *testing.T) {
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

	// Test data to restore
	data := []map[string]interface{}{
		{
			"id":            int64(10),
			"email":         "restore1@example.com",
			"password_hash": "hash1",
			"name":          "Restore User 1",
			"role":          "user",
			"created_at":    "2025-01-01 00:00:00",
			"updated_at":    "2025-01-01 00:00:00",
		},
		{
			"id":            int64(11),
			"email":         "restore2@example.com",
			"password_hash": "hash2",
			"name":          "Restore User 2",
			"role":          "admin",
			"created_at":    "2025-01-01 00:00:00",
			"updated_at":    "2025-01-01 00:00:00",
		},
	}

	sourceSchema := domain.TableSchema{
		Columns: []domain.ColumnSchema{
			{Name: "id", Type: "integer", Nullable: false},
			{Name: "email", Type: "string", Nullable: false},
			{Name: "password_hash", Type: "string", Nullable: false},
			{Name: "name", Type: "string", Nullable: false},
			{Name: "role", Type: "string", Nullable: false},
			{Name: "created_at", Type: "datetime", Nullable: false},
			{Name: "updated_at", Type: "datetime", Nullable: false},
		},
	}

	err = svc.restoreTable(tx, "users", data, sourceSchema)
	if err != nil {
		t.Fatalf("restoreTable error: %v", err)
	}

	// Verify data was inserted
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count users: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 users, got %d", count)
	}
}

// TestBackupService_RestoreTable_EmptyData tests restoring empty data
func TestBackupService_RestoreTable_EmptyData(t *testing.T) {
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

	err = svc.restoreTable(tx, "users", []map[string]interface{}{}, domain.TableSchema{})
	if err != nil {
		t.Errorf("restoreTable with empty data should not error: %v", err)
	}
}

// TestBackupService_RestoreTable_NonExistentTable tests restoring to non-existent table
func TestBackupService_RestoreTable_NonExistentTable(t *testing.T) {
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

	data := []map[string]interface{}{
		{"id": int64(1), "name": "test"},
	}

	// Should not error - just skip the table
	err = svc.restoreTable(tx, "nonexistent_table_xyz", data, domain.TableSchema{})
	if err != nil {
		t.Errorf("restoreTable to non-existent table should not error: %v", err)
	}
}

// TestBackupService_InsertRecord tests record insertion
func TestBackupService_InsertRecord(t *testing.T) {
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

	row := map[string]interface{}{
		"email":         "insert@example.com",
		"password_hash": "hash",
		"name":          "Insert User",
		"role":          "user",
		"created_at":    "2025-01-01 00:00:00",
		"updated_at":    "2025-01-01 00:00:00",
	}

	targetColumns := []string{"id", "email", "password_hash", "name", "role", "created_at", "updated_at"}
	columnTypes := map[string]string{
		"email":         "string",
		"password_hash": "string",
		"name":          "string",
		"role":          "string",
		"created_at":    "datetime",
		"updated_at":    "datetime",
	}

	newID, err := svc.insertRecord(tx, "users", row, targetColumns, columnTypes)
	if err != nil {
		t.Fatalf("insertRecord error: %v", err)
	}

	if newID == 0 {
		t.Error("expected non-zero new ID")
	}

	// Verify record was inserted
	var email string
	err = tx.QueryRow("SELECT email FROM users WHERE id = ?", newID).Scan(&email)
	if err != nil {
		t.Fatalf("failed to query inserted record: %v", err)
	}
	if email != "insert@example.com" {
		t.Errorf("expected email 'insert@example.com', got %q", email)
	}
}

// TestBackupService_InsertRecord_EmptyRow tests inserting empty row
func TestBackupService_InsertRecord_EmptyRow(t *testing.T) {
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

	newID, err := svc.insertRecord(tx, "users", map[string]interface{}{"id": int64(1)}, []string{"id"}, map[string]string{})
	if err != nil {
		t.Fatalf("insertRecord with only ID should not error: %v", err)
	}
	// ID column is skipped, so no columns means no insert
	if newID != 0 {
		t.Errorf("expected 0 ID for empty column insert, got %d", newID)
	}
}

// TestBackupService_UpdateRecord tests record update
func TestBackupService_UpdateRecord(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert a record first
	_, err := db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('update@example.com', 'hash', 'Original Name', 'user', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// Get the inserted ID
	var userID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = 'update@example.com'").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to get user ID: %v", err)
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

	row := map[string]interface{}{
		"id":   userID,
		"name": "Updated Name",
		"role": "admin",
	}

	targetColumns := []string{"id", "email", "password_hash", "name", "role", "created_at", "updated_at"}
	columnTypes := map[string]string{
		"name": "string",
		"role": "string",
	}

	err = svc.updateRecord(tx, "users", userID, row, targetColumns, columnTypes)
	if err != nil {
		t.Fatalf("updateRecord error: %v", err)
	}

	// Verify record was updated
	var name, role string
	err = tx.QueryRow("SELECT name, role FROM users WHERE id = ?", userID).Scan(&name, &role)
	if err != nil {
		t.Fatalf("failed to query updated record: %v", err)
	}
	if name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %q", name)
	}
	if role != "admin" {
		t.Errorf("expected role 'admin', got %q", role)
	}
}

// TestBackupService_UpdateRecord_NoValidColumns tests updating with no valid columns
func TestBackupService_UpdateRecord_NoValidColumns(t *testing.T) {
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

	// Only ID in row, no columns to update
	row := map[string]interface{}{"id": int64(1)}

	err = svc.updateRecord(tx, "users", 1, row, []string{"id"}, map[string]string{})
	if err != nil {
		t.Errorf("updateRecord with only ID should not error: %v", err)
	}
}

// TestBackupService_RestoreTableToSQLite tests SQLite table restore
func TestBackupService_RestoreTableToSQLite(t *testing.T) {
	// Create an in-memory SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open SQLite database: %v", err)
	}
	defer db.Close()

	// Create a simple table
	_, err = db.Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			value INTEGER
		)
	`)
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	svc := &BackupServiceImpl{dbDriver: "sqlite3"}

	data := []map[string]interface{}{
		{"id": int64(1), "name": "Test 1", "value": int64(100)},
		{"id": int64(2), "name": "Test 2", "value": int64(200)},
	}

	err = svc.restoreTableToSQLite(tx, "test_table", data)
	if err != nil {
		t.Fatalf("restoreTableToSQLite error: %v", err)
	}

	// Verify data was inserted
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 rows, got %d", count)
	}
}

// TestBackupService_RestoreTableToSQLite_EmptyData tests empty data handling
func TestBackupService_RestoreTableToSQLite_EmptyData(t *testing.T) {
	svc := &BackupServiceImpl{dbDriver: "sqlite3"}

	err := svc.restoreTableToSQLite(nil, "test_table", []map[string]interface{}{})
	if err != nil {
		t.Errorf("restoreTableToSQLite with empty data should not error: %v", err)
	}
}

// TestBackupService_ExtractSQLiteSchema tests schema extraction from SQLite
func TestBackupService_ExtractSQLiteSchema(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := &BackupServiceImpl{
		db:       db,
		dbDriver: "sqlite3",
	}

	schema, err := svc.extractSQLiteSchema()
	if err != nil {
		t.Fatalf("extractSQLiteSchema error: %v", err)
	}

	// Verify schema contains CREATE TABLE statements
	if !strings.Contains(schema, "CREATE TABLE") {
		t.Error("expected CREATE TABLE in schema")
	}

	// Verify users table is present
	if !strings.Contains(schema, "users") {
		t.Error("expected 'users' table in schema")
	}
}

// TestBackupService_FindByNaturalKey_UserWorkouts tests user_workouts natural key matching
func TestBackupService_FindByNaturalKey_UserWorkouts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test user
	_, err := db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('test@example.com', 'hash', 'Test User', 'user', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// Get user ID
	var userID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = 'test@example.com'").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to get user ID: %v", err)
	}

	// Insert test user workout
	_, err = db.Exec(`
		INSERT INTO user_workouts (user_id, workout_date, workout_name, created_at, updated_at)
		VALUES (?, '2025-01-15', 'Test Workout', datetime('now'), datetime('now'))
	`, userID)
	if err != nil {
		t.Fatalf("failed to insert test user workout: %v", err)
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
		row         map[string]interface{}
		expectFound bool
	}{
		{
			name: "find with workout_name",
			row: map[string]interface{}{
				"user_id":      userID,
				"workout_date": "2025-01-15",
				"workout_name": "Test Workout",
			},
			expectFound: true,
		},
		{
			name: "not found with different date",
			row: map[string]interface{}{
				"user_id":      userID,
				"workout_date": "2025-01-16",
				"workout_name": "Test Workout",
			},
			expectFound: false,
		},
		{
			name: "missing user_id returns not found",
			row: map[string]interface{}{
				"workout_date": "2025-01-15",
				"workout_name": "Test Workout",
			},
			expectFound: false,
		},
		{
			name: "missing workout_date returns not found",
			row: map[string]interface{}{
				"user_id":      userID,
				"workout_name": "Test Workout",
			},
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, found, err := svc.findByNaturalKey(tx, "user_workouts", tt.row, mappings)
			if err != nil {
				t.Fatalf("findByNaturalKey error: %v", err)
			}
			if found != tt.expectFound {
				t.Errorf("findByNaturalKey found = %v, expected %v", found, tt.expectFound)
			}
		})
	}
}

// TestBackupService_FindByNaturalKey_UserSettings tests user_settings natural key matching
func TestBackupService_FindByNaturalKey_UserSettings(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Add user_settings table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER UNIQUE NOT NULL,
			theme TEXT DEFAULT 'light',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("failed to create user_settings table: %v", err)
	}

	// Insert test user
	_, err = db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('settings@example.com', 'hash', 'Settings User', 'user', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	var userID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = 'settings@example.com'").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to get user ID: %v", err)
	}

	// Insert user settings
	_, err = db.Exec(`
		INSERT INTO user_settings (user_id, theme, created_at, updated_at)
		VALUES (?, 'dark', datetime('now'), datetime('now'))
	`, userID)
	if err != nil {
		t.Fatalf("failed to insert user settings: %v", err)
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

	// Test finding by user_id
	id, found, err := svc.findByNaturalKey(tx, "user_settings", map[string]interface{}{"user_id": userID}, mappings)
	if err != nil {
		t.Fatalf("findByNaturalKey error: %v", err)
	}
	if !found {
		t.Error("expected to find user_settings")
	}
	if id == 0 {
		t.Error("expected non-zero ID")
	}

	// Test not found with missing user_id
	_, found, err = svc.findByNaturalKey(tx, "user_settings", map[string]interface{}{}, mappings)
	if err != nil {
		t.Fatalf("findByNaturalKey error: %v", err)
	}
	if found {
		t.Error("expected not found for missing user_id")
	}
}

// TestBackupService_RestoreTableWithMode tests the mode-based restore function
func TestBackupService_RestoreTableWithMode(t *testing.T) {
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

	data := []map[string]interface{}{
		{
			"id":            int64(100),
			"email":         "mode@example.com",
			"password_hash": "hash",
			"name":          "Mode User",
			"role":          "user",
			"created_at":    "2025-01-01 00:00:00",
			"updated_at":    "2025-01-01 00:00:00",
		},
	}

	sourceSchema := domain.TableSchema{
		Columns: []domain.ColumnSchema{
			{Name: "id", Type: "integer"},
			{Name: "email", Type: "string"},
			{Name: "password_hash", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "role", Type: "string"},
			{Name: "created_at", Type: "datetime"},
			{Name: "updated_at", Type: "datetime"},
		},
	}

	mappings := newIDMappings()
	result := &domain.RestoreResult{Errors: []string{}}

	// Test replace mode
	err = svc.restoreTableWithMode(tx, "users", data, sourceSchema, domain.RestoreModeReplace, mappings, result)
	if err != nil {
		t.Fatalf("restoreTableWithMode error: %v", err)
	}

	// Verify data was inserted
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count users: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 user, got %d", count)
	}
}

// TestBackupService_RestoreTableWithMode_EmptyData tests empty data handling
func TestBackupService_RestoreTableWithMode_EmptyData(t *testing.T) {
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

	mappings := newIDMappings()
	result := &domain.RestoreResult{Errors: []string{}}

	err = svc.restoreTableWithMode(tx, "users", []map[string]interface{}{}, domain.TableSchema{}, domain.RestoreModeReplace, mappings, result)
	if err != nil {
		t.Errorf("restoreTableWithMode with empty data should not error: %v", err)
	}
}

// TestBackupService_MergeTable_MergeMode tests merge mode restore
func TestBackupService_MergeTable_MergeMode(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert existing record
	_, err := db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('merge@example.com', 'oldhash', 'Old Name', 'user', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert existing user: %v", err)
	}

	var existingID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = 'merge@example.com'").Scan(&existingID)
	if err != nil {
		t.Fatalf("failed to get existing user ID: %v", err)
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

	// Data to merge - should update existing record
	data := []map[string]interface{}{
		{
			"id":            int64(999), // Old ID from backup
			"email":         "merge@example.com",
			"password_hash": "newhash",
			"name":          "Updated Name",
			"role":          "admin",
			"created_at":    "2025-01-01 00:00:00",
			"updated_at":    "2025-01-01 00:00:00",
		},
	}

	sourceSchema := domain.TableSchema{
		Columns: []domain.ColumnSchema{
			{Name: "id", Type: "integer"},
			{Name: "email", Type: "string"},
			{Name: "password_hash", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "role", Type: "string"},
			{Name: "created_at", Type: "datetime"},
			{Name: "updated_at", Type: "datetime"},
		},
	}

	mappings := newIDMappings()
	result := &domain.RestoreResult{Errors: []string{}}

	err = svc.mergeTable(tx, "users", data, sourceSchema, domain.RestoreModeMerge, mappings, result)
	if err != nil {
		t.Fatalf("mergeTable error: %v", err)
	}

	// Verify record was updated
	if result.RecordsUpdated != 1 {
		t.Errorf("expected 1 record updated, got %d", result.RecordsUpdated)
	}

	// Verify ID mapping was recorded
	if mappings.users[999] != existingID {
		t.Errorf("expected ID mapping 999 -> %d, got %d", existingID, mappings.users[999])
	}

	// Verify data was updated
	var name string
	err = tx.QueryRow("SELECT name FROM users WHERE id = ?", existingID).Scan(&name)
	if err != nil {
		t.Fatalf("failed to query updated record: %v", err)
	}
	if name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %q", name)
	}
}

// TestBackupService_MergeTable_SkipMode tests skip mode restore
func TestBackupService_MergeTable_SkipMode(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert existing record
	_, err := db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('skip@example.com', 'oldhash', 'Old Name', 'user', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert existing user: %v", err)
	}

	var existingID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = 'skip@example.com'").Scan(&existingID)
	if err != nil {
		t.Fatalf("failed to get existing user ID: %v", err)
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

	// Data to skip - should NOT update existing record
	data := []map[string]interface{}{
		{
			"id":            int64(888),
			"email":         "skip@example.com",
			"password_hash": "newhash",
			"name":          "Should Not Update",
			"role":          "admin",
			"created_at":    "2025-01-01 00:00:00",
			"updated_at":    "2025-01-01 00:00:00",
		},
	}

	sourceSchema := domain.TableSchema{
		Columns: []domain.ColumnSchema{
			{Name: "id", Type: "integer"},
			{Name: "email", Type: "string"},
			{Name: "password_hash", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "role", Type: "string"},
			{Name: "created_at", Type: "datetime"},
			{Name: "updated_at", Type: "datetime"},
		},
	}

	mappings := newIDMappings()
	result := &domain.RestoreResult{Errors: []string{}}

	err = svc.mergeTable(tx, "users", data, sourceSchema, domain.RestoreModeSkip, mappings, result)
	if err != nil {
		t.Fatalf("mergeTable error: %v", err)
	}

	// Verify record was skipped
	if result.RecordsSkipped != 1 {
		t.Errorf("expected 1 record skipped, got %d", result.RecordsSkipped)
	}

	// Verify ID mapping was still recorded
	if mappings.users[888] != existingID {
		t.Errorf("expected ID mapping 888 -> %d, got %d", existingID, mappings.users[888])
	}

	// Verify data was NOT updated
	var name string
	err = tx.QueryRow("SELECT name FROM users WHERE id = ?", existingID).Scan(&name)
	if err != nil {
		t.Fatalf("failed to query record: %v", err)
	}
	if name != "Old Name" {
		t.Errorf("expected name 'Old Name' (unchanged), got %q", name)
	}
}

// TestBackupService_MergeTable_InsertNew tests inserting new records in merge mode
func TestBackupService_MergeTable_InsertNew(t *testing.T) {
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

	// Data to insert - no existing record
	data := []map[string]interface{}{
		{
			"id":            int64(777),
			"email":         "new@example.com",
			"password_hash": "hash",
			"name":          "New User",
			"role":          "user",
			"created_at":    "2025-01-01 00:00:00",
			"updated_at":    "2025-01-01 00:00:00",
		},
	}

	sourceSchema := domain.TableSchema{
		Columns: []domain.ColumnSchema{
			{Name: "id", Type: "integer"},
			{Name: "email", Type: "string"},
			{Name: "password_hash", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "role", Type: "string"},
			{Name: "created_at", Type: "datetime"},
			{Name: "updated_at", Type: "datetime"},
		},
	}

	mappings := newIDMappings()
	result := &domain.RestoreResult{Errors: []string{}}

	err = svc.mergeTable(tx, "users", data, sourceSchema, domain.RestoreModeMerge, mappings, result)
	if err != nil {
		t.Fatalf("mergeTable error: %v", err)
	}

	// Verify record was created
	if result.RecordsCreated != 1 {
		t.Errorf("expected 1 record created, got %d", result.RecordsCreated)
	}

	// Verify ID mapping was recorded
	if mappings.users[777] == 0 {
		t.Error("expected ID mapping for new record")
	}

	// Verify data was inserted
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM users WHERE email = 'new@example.com'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count users: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 user, got %d", count)
	}
}

// TestBackupService_MergeTable_NonExistentTable tests merging to non-existent table
func TestBackupService_MergeTable_NonExistentTable(t *testing.T) {
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

	data := []map[string]interface{}{
		{"id": int64(1), "name": "test"},
	}

	mappings := newIDMappings()
	result := &domain.RestoreResult{Errors: []string{}}

	// Should not error for non-existent table
	err = svc.mergeTable(tx, "nonexistent_table_xyz", data, domain.TableSchema{}, domain.RestoreModeMerge, mappings, result)
	if err != nil {
		t.Errorf("mergeTable to non-existent table should not error: %v", err)
	}
}

// TestBackupService_GetBackupMetadata tests metadata extraction from backup files
func TestBackupService_GetBackupMetadata(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "backup_metadata_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid backup ZIP file
	backupFilePath := filepath.Join(tmpDir, "test_backup.zip")
	zipFile, err := os.Create(backupFilePath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}

	zipWriter := zip.NewWriter(zipFile)

	// Create backup_data.json with metadata
	backupData := domain.BackupData{
		Metadata: domain.BackupMetadata{
			Filename:       "test_backup.zip",
			Version:        "1.0.0",
			DatabaseDriver: "sqlite3",
			DatabaseName:   "test.db",
			TotalUsers:     5,
			TotalWorkouts:  10,
			TotalMovements: 20,
			TotalWODs:      15,
		},
		Users: []map[string]interface{}{
			{"id": int64(1), "email": "test@example.com"},
		},
	}

	dataJSON, err := json.Marshal(backupData)
	if err != nil {
		t.Fatalf("failed to marshal backup data: %v", err)
	}

	dataFile, err := zipWriter.Create("backup_data.json")
	if err != nil {
		t.Fatalf("failed to create data file in ZIP: %v", err)
	}
	if _, err := dataFile.Write(dataJSON); err != nil {
		t.Fatalf("failed to write data to ZIP: %v", err)
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("failed to close ZIP writer: %v", err)
	}
	zipFile.Close()

	svc := &BackupServiceImpl{
		backupDir: tmpDir,
	}

	metadata, err := svc.GetBackupMetadata("test_backup.zip")
	if err != nil {
		t.Fatalf("GetBackupMetadata error: %v", err)
	}

	if metadata.Filename != "test_backup.zip" {
		t.Errorf("expected filename 'test_backup.zip', got %q", metadata.Filename)
	}
	if metadata.TotalUsers != 5 {
		t.Errorf("expected 5 users, got %d", metadata.TotalUsers)
	}
	if metadata.TotalWorkouts != 10 {
		t.Errorf("expected 10 workouts, got %d", metadata.TotalWorkouts)
	}
}

// TestBackupService_GetBackupMetadata_NoDataFile tests error when backup_data.json is missing
func TestBackupService_GetBackupMetadata_NoDataFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "backup_metadata_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a ZIP without backup_data.json
	backupFilePath := filepath.Join(tmpDir, "empty_backup.zip")
	zipFile, err := os.Create(backupFilePath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}

	zipWriter := zip.NewWriter(zipFile)
	// Add a different file
	otherFile, _ := zipWriter.Create("other.txt")
	otherFile.Write([]byte("test"))
	zipWriter.Close()
	zipFile.Close()

	svc := &BackupServiceImpl{
		backupDir: tmpDir,
	}

	_, err = svc.GetBackupMetadata("empty_backup.zip")
	if err == nil {
		t.Error("expected error for missing backup_data.json")
	}
	if !strings.Contains(err.Error(), "backup_data.json not found") {
		t.Errorf("expected 'backup_data.json not found' error, got: %v", err)
	}
}

// TestBackupService_GetBackupMetadata_InvalidZip tests error for invalid ZIP file
func TestBackupService_GetBackupMetadata_InvalidZip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "backup_metadata_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an invalid ZIP file
	invalidPath := filepath.Join(tmpDir, "invalid.zip")
	if err := os.WriteFile(invalidPath, []byte("not a zip file"), 0644); err != nil {
		t.Fatalf("failed to create invalid file: %v", err)
	}

	svc := &BackupServiceImpl{
		backupDir: tmpDir,
	}

	_, err = svc.GetBackupMetadata("invalid.zip")
	if err == nil {
		t.Error("expected error for invalid ZIP file")
	}
}

// TestBackupService_AddUploadsToZip tests adding uploads to ZIP
func TestBackupService_AddUploadsToZip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "uploads_zip_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	uploadsDir := filepath.Join(tmpDir, "uploads")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		t.Fatalf("failed to create uploads dir: %v", err)
	}

	// Create test upload files
	testFile1 := filepath.Join(uploadsDir, "avatar.jpg")
	if err := os.WriteFile(testFile1, []byte("fake image data"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	testFile2 := filepath.Join(uploadsDir, "profile.png")
	if err := os.WriteFile(testFile2, []byte("another fake image"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create ZIP file
	zipPath := filepath.Join(tmpDir, "backup.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)

	svc := &BackupServiceImpl{
		uploadsDir: uploadsDir,
	}

	err = svc.addUploadsToZip(zipWriter)
	if err != nil {
		t.Fatalf("addUploadsToZip error: %v", err)
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("failed to close ZIP writer: %v", err)
	}
	zipFile.Close()

	// Verify files were added to ZIP
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open ZIP: %v", err)
	}
	defer zipReader.Close()

	uploadFiles := 0
	for _, f := range zipReader.File {
		if strings.HasPrefix(f.Name, "uploads/") {
			uploadFiles++
		}
	}

	if uploadFiles != 2 {
		t.Errorf("expected 2 upload files in ZIP, got %d", uploadFiles)
	}
}

// TestBackupService_AddUploadsToZip_NoUploadsDir tests handling missing uploads directory
func TestBackupService_AddUploadsToZip_NoUploadsDir(t *testing.T) {
	svc := &BackupServiceImpl{
		uploadsDir: "/nonexistent/uploads/directory",
	}

	// Create a dummy ZIP writer
	var buf strings.Builder
	zipWriter := zip.NewWriter(&buf)
	defer zipWriter.Close()

	// Should not error for non-existent uploads directory
	err := svc.addUploadsToZip(zipWriter)
	if err != nil {
		t.Errorf("addUploadsToZip should not error for non-existent dir: %v", err)
	}
}

// TestBackupService_RestoreUploadsFromZip tests restoring uploads from ZIP
func TestBackupService_RestoreUploadsFromZip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "restore_uploads_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a ZIP with upload files
	zipPath := filepath.Join(tmpDir, "backup_with_uploads.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}

	zipWriter := zip.NewWriter(zipFile)

	// Add upload files to ZIP
	uploadFile1, _ := zipWriter.Create("uploads/avatar.jpg")
	uploadFile1.Write([]byte("fake image data"))

	uploadFile2, _ := zipWriter.Create("uploads/profile.png")
	uploadFile2.Write([]byte("another fake image"))

	// Add non-upload file (should be ignored)
	otherFile, _ := zipWriter.Create("backup_data.json")
	otherFile.Write([]byte("{}"))

	zipWriter.Close()
	zipFile.Close()

	// Open the ZIP for reading
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open ZIP: %v", err)
	}
	defer zipReader.Close()

	uploadsDir := filepath.Join(tmpDir, "restored_uploads")
	svc := &BackupServiceImpl{
		uploadsDir: uploadsDir,
	}

	err = svc.restoreUploadsFromZip(zipReader)
	if err != nil {
		t.Fatalf("restoreUploadsFromZip error: %v", err)
	}

	// Verify files were restored
	files, err := os.ReadDir(uploadsDir)
	if err != nil {
		t.Fatalf("failed to read uploads dir: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("expected 2 restored files, got %d", len(files))
	}

	// Verify file contents
	content, err := os.ReadFile(filepath.Join(uploadsDir, "avatar.jpg"))
	if err != nil {
		t.Fatalf("failed to read restored file: %v", err)
	}
	if string(content) != "fake image data" {
		t.Errorf("unexpected file content: %s", string(content))
	}
}

// mockMultipartFile implements multipart.File for testing
type mockMultipartFile struct {
	*strings.Reader
}

func (m *mockMultipartFile) Close() error {
	return nil
}

// TestBackupService_UploadBackup tests backup file upload
func TestBackupService_UploadBackup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "upload_backup_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid ZIP file content
	var zipBuffer strings.Builder
	zipWriter := zip.NewWriter(&zipBuffer)
	dataFile, _ := zipWriter.Create("backup_data.json")
	dataFile.Write([]byte("{}"))
	zipWriter.Close()

	svc := &BackupServiceImpl{
		backupDir:    tmpDir,
		auditLogRepo: &mockAuditLogRepo{},
	}

	// Create mock file
	mockFile := &mockMultipartFile{
		Reader: strings.NewReader(zipBuffer.String()),
	}

	filename, err := svc.UploadBackup(mockFile, "original_backup.zip", 1)
	if err != nil {
		t.Fatalf("UploadBackup error: %v", err)
	}

	if filename == "" {
		t.Error("expected non-empty filename")
	}

	// Verify file was saved
	savedPath := filepath.Join(tmpDir, filename)
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		t.Error("expected file to be saved")
	}
}

// TestBackupService_UploadBackup_InvalidFileType tests invalid file type error
func TestBackupService_UploadBackup_InvalidFileType(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "upload_backup_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	svc := &BackupServiceImpl{
		backupDir:    tmpDir,
		auditLogRepo: &mockAuditLogRepo{},
	}

	// Pass invalid file type (string instead of multipart.File)
	_, err = svc.UploadBackup("not a file", "test.zip", 1)
	if err == nil {
		t.Error("expected error for invalid file type")
	}
	if !strings.Contains(err.Error(), "invalid file type") {
		t.Errorf("expected 'invalid file type' error, got: %v", err)
	}
}

// TestBackupService_UploadBackup_InvalidZip tests invalid ZIP file rejection
func TestBackupService_UploadBackup_InvalidZip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "upload_backup_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	svc := &BackupServiceImpl{
		backupDir:    tmpDir,
		auditLogRepo: &mockAuditLogRepo{},
	}

	// Create mock file with invalid ZIP content
	mockFile := &mockMultipartFile{
		Reader: strings.NewReader("not a zip file content"),
	}

	_, err = svc.UploadBackup(mockFile, "invalid.zip", 1)
	if err == nil {
		t.Error("expected error for invalid ZIP file")
	}
	if !strings.Contains(err.Error(), "not a valid ZIP archive") {
		t.Errorf("expected 'not a valid ZIP archive' error, got: %v", err)
	}
}

// TestBackupService_ListBackups_WithValidBackups tests listing backups with valid files
func TestBackupService_ListBackups_WithValidBackups(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "list_backups_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid backup file
	backupPath := filepath.Join(tmpDir, "actalog_backup_20250101_120000.zip")
	zipFile, err := os.Create(backupPath)
	if err != nil {
		t.Fatalf("failed to create backup file: %v", err)
	}

	zipWriter := zip.NewWriter(zipFile)
	backupData := domain.BackupData{
		Metadata: domain.BackupMetadata{
			Filename:   "actalog_backup_20250101_120000.zip",
			TotalUsers: 5,
			Version:    "1.0.0",
		},
	}
	dataJSON, _ := json.Marshal(backupData)
	dataFile, _ := zipWriter.Create("backup_data.json")
	dataFile.Write(dataJSON)
	zipWriter.Close()
	zipFile.Close()

	// Create a non-zip file (should be skipped)
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("test"), 0644)

	// Create a directory (should be skipped)
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)

	svc := &BackupServiceImpl{
		backupDir: tmpDir,
	}

	backups, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups error: %v", err)
	}

	if len(backups) != 1 {
		t.Errorf("expected 1 backup, got %d", len(backups))
	}

	if len(backups) > 0 && backups[0].TotalUsers != 5 {
		t.Errorf("expected 5 users, got %d", backups[0].TotalUsers)
	}
}

// TestBackupService_FindByNaturalKey_MoreTables tests natural key matching for additional tables
func TestBackupService_FindByNaturalKey_MoreTables(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create additional tables for testing
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_subscriptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER UNIQUE NOT NULL,
			subscription_type TEXT NOT NULL,
			status TEXT NOT NULL,
			start_date DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("failed to create user_subscriptions table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS organization_subscriptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			organization_id INTEGER UNIQUE NOT NULL,
			subscription_type TEXT NOT NULL,
			status TEXT NOT NULL,
			start_date DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("failed to create organization_subscriptions table: %v", err)
	}

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('sub@example.com', 'hash', 'Sub User', 'user', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	var userID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = 'sub@example.com'").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to get user ID: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO user_subscriptions (user_id, subscription_type, status, start_date, created_at, updated_at)
		VALUES (?, 'premium', 'active', datetime('now'), datetime('now'), datetime('now'))
	`, userID)
	if err != nil {
		t.Fatalf("failed to insert user subscription: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO organizations (name, created_at, updated_at)
		VALUES ('Test Org Sub', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert organization: %v", err)
	}

	var orgID int64
	err = db.QueryRow("SELECT id FROM organizations WHERE name = 'Test Org Sub'").Scan(&orgID)
	if err != nil {
		t.Fatalf("failed to get org ID: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO organization_subscriptions (organization_id, subscription_type, status, start_date, created_at, updated_at)
		VALUES (?, 'enterprise', 'active', datetime('now'), datetime('now'), datetime('now'))
	`, orgID)
	if err != nil {
		t.Fatalf("failed to insert org subscription: %v", err)
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

	// Test user_subscriptions
	id, found, err := svc.findByNaturalKey(tx, "user_subscriptions", map[string]interface{}{"user_id": userID}, mappings)
	if err != nil {
		t.Fatalf("findByNaturalKey error for user_subscriptions: %v", err)
	}
	if !found {
		t.Error("expected to find user_subscriptions")
	}
	if id == 0 {
		t.Error("expected non-zero ID for user_subscriptions")
	}

	// Test user_subscriptions not found
	_, found, err = svc.findByNaturalKey(tx, "user_subscriptions", map[string]interface{}{}, mappings)
	if err != nil {
		t.Fatalf("findByNaturalKey error: %v", err)
	}
	if found {
		t.Error("expected not found for missing user_id")
	}

	// Test organization_subscriptions
	id, found, err = svc.findByNaturalKey(tx, "organization_subscriptions", map[string]interface{}{"organization_id": orgID}, mappings)
	if err != nil {
		t.Fatalf("findByNaturalKey error for organization_subscriptions: %v", err)
	}
	if !found {
		t.Error("expected to find organization_subscriptions")
	}
	if id == 0 {
		t.Error("expected non-zero ID for organization_subscriptions")
	}

	// Test organization_subscriptions not found
	_, found, err = svc.findByNaturalKey(tx, "organization_subscriptions", map[string]interface{}{}, mappings)
	if err != nil {
		t.Fatalf("findByNaturalKey error: %v", err)
	}
	if found {
		t.Error("expected not found for missing organization_id")
	}
}

// TestBackupService_FindByNaturalKey_UserOrganizations tests user_organizations natural key
func TestBackupService_FindByNaturalKey_UserOrganizations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create user_organizations table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_organizations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			organization_id INTEGER NOT NULL,
			role TEXT DEFAULT 'member',
			joined_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			UNIQUE (user_id, organization_id)
		)
	`)
	if err != nil {
		t.Fatalf("failed to create user_organizations table: %v", err)
	}

	// Insert test user
	_, err = db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('orguser@example.com', 'hash', 'Org User', 'user', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	var userID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = 'orguser@example.com'").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to get user ID: %v", err)
	}

	// Insert test organization
	_, err = db.Exec(`
		INSERT INTO organizations (name, created_at, updated_at)
		VALUES ('User Org', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test organization: %v", err)
	}

	var orgID int64
	err = db.QueryRow("SELECT id FROM organizations WHERE name = 'User Org'").Scan(&orgID)
	if err != nil {
		t.Fatalf("failed to get org ID: %v", err)
	}

	// Insert user_organization
	_, err = db.Exec(`
		INSERT INTO user_organizations (user_id, organization_id, role, joined_at, created_at, updated_at)
		VALUES (?, ?, 'admin', datetime('now'), datetime('now'), datetime('now'))
	`, userID, orgID)
	if err != nil {
		t.Fatalf("failed to insert user_organization: %v", err)
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

	// Test finding by user_id and organization_id
	id, found, err := svc.findByNaturalKey(tx, "user_organizations", map[string]interface{}{
		"user_id":         userID,
		"organization_id": orgID,
	}, mappings)
	if err != nil {
		t.Fatalf("findByNaturalKey error: %v", err)
	}
	if !found {
		t.Error("expected to find user_organization")
	}
	if id == 0 {
		t.Error("expected non-zero ID")
	}

	// Test not found with missing user_id
	_, found, err = svc.findByNaturalKey(tx, "user_organizations", map[string]interface{}{"organization_id": orgID}, mappings)
	if err != nil {
		t.Fatalf("findByNaturalKey error: %v", err)
	}
	if found {
		t.Error("expected not found for missing user_id")
	}

	// Test not found with missing organization_id
	_, found, err = svc.findByNaturalKey(tx, "user_organizations", map[string]interface{}{"user_id": userID}, mappings)
	if err != nil {
		t.Fatalf("findByNaturalKey error: %v", err)
	}
	if found {
		t.Error("expected not found for missing organization_id")
	}
}

// TestBackupService_FindByNaturalKey_UserWorkoutsNullName tests user_workouts with null workout_name
func TestBackupService_FindByNaturalKey_UserWorkoutsNullName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test user
	_, err := db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('nullname@example.com', 'hash', 'Null Name User', 'user', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	var userID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = 'nullname@example.com'").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to get user ID: %v", err)
	}

	// Insert user workout with NULL workout_name
	_, err = db.Exec(`
		INSERT INTO user_workouts (user_id, workout_date, workout_name, created_at, updated_at)
		VALUES (?, '2025-01-20', NULL, datetime('now'), datetime('now'))
	`, userID)
	if err != nil {
		t.Fatalf("failed to insert user workout: %v", err)
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

	// Test finding with empty workout_name (should match NULL)
	_, found, err := svc.findByNaturalKey(tx, "user_workouts", map[string]interface{}{
		"user_id":      userID,
		"workout_date": "2025-01-20",
		"workout_name": "",
	}, mappings)
	if err != nil {
		t.Fatalf("findByNaturalKey error: %v", err)
	}
	if !found {
		t.Error("expected to find user_workout with NULL workout_name")
	}
}

// setupFullTestDB creates an in-memory SQLite database with all tables needed for backup testing
func setupFullTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create complete schema for testing backups
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
			updated_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS wods (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			score_type TEXT,
			is_standard INTEGER NOT NULL DEFAULT 0,
			created_by INTEGER,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS workouts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			notes TEXT,
			created_by INTEGER,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
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
			updated_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS workout_movements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workout_id INTEGER NOT NULL,
			movement_id INTEGER NOT NULL,
			sets INTEGER,
			reps INTEGER,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS workout_wods (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workout_id INTEGER NOT NULL,
			wod_id INTEGER NOT NULL,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS user_workout_movements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_workout_id INTEGER NOT NULL,
			movement_id INTEGER NOT NULL,
			sets INTEGER,
			reps INTEGER,
			weight REAL,
			is_pr INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS user_workout_wods (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_workout_id INTEGER NOT NULL,
			wod_id INTEGER NOT NULL,
			score TEXT,
			is_pr INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS password_resets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS email_verification_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			action TEXT NOT NULL,
			entity_type TEXT,
			entity_id INTEGER,
			details TEXT,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS user_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER UNIQUE NOT NULL,
			theme TEXT DEFAULT 'light',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS data_change_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			table_name TEXT NOT NULL,
			record_id INTEGER,
			action TEXT NOT NULL,
			old_data TEXT,
			new_data TEXT,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS user_organizations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			organization_id INTEGER NOT NULL,
			role TEXT DEFAULT 'member',
			created_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS user_subscriptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			plan TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS organization_subscriptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			organization_id INTEGER NOT NULL,
			plan TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS notifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			type TEXT NOT NULL,
			title TEXT NOT NULL,
			message TEXT,
			is_read INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS notification_likes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			notification_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			created_at DATETIME NOT NULL
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

// TestBackupService_ExportAllTables tests exporting all database tables
func TestBackupService_ExportAllTables(t *testing.T) {
	db := setupFullTestDB(t)
	defer db.Close()

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('test@example.com', 'hash123', 'Test User', 'user', '2025-01-01 00:00:00', '2025-01-01 00:00:00')
	`)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO movements (name, type, is_standard, created_at, updated_at)
		VALUES ('Squat', 'weightlifting', 1, '2025-01-01 00:00:00', '2025-01-01 00:00:00')
	`)
	if err != nil {
		t.Fatalf("failed to insert test movement: %v", err)
	}

	svc := &BackupServiceImpl{
		db:       db,
		dbDriver: "sqlite3",
	}

	data, err := svc.exportAllTables()
	if err != nil {
		t.Fatalf("exportAllTables error: %v", err)
	}

	if data == nil {
		t.Fatal("expected non-nil backup data")
	}

	// Verify users exported
	if len(data.Users) != 1 {
		t.Errorf("expected 1 user, got %d", len(data.Users))
	}

	// Verify movements exported
	if len(data.Movements) != 1 {
		t.Errorf("expected 1 movement, got %d", len(data.Movements))
	}

	// Verify schema metadata captured
	if data.Schema.Tables == nil {
		t.Error("expected schema tables to be initialized")
	}
	if _, ok := data.Schema.Tables["users"]; !ok {
		t.Error("expected users table in schema metadata")
	}
}

// TestBackupService_ExportAllTables_EmptyDatabase tests export with no data
func TestBackupService_ExportAllTables_EmptyDatabase(t *testing.T) {
	db := setupFullTestDB(t)
	defer db.Close()

	svc := &BackupServiceImpl{
		db:       db,
		dbDriver: "sqlite3",
	}

	data, err := svc.exportAllTables()
	if err != nil {
		t.Fatalf("exportAllTables error: %v", err)
	}

	if len(data.Users) != 0 {
		t.Errorf("expected 0 users, got %d", len(data.Users))
	}
	if len(data.Movements) != 0 {
		t.Errorf("expected 0 movements, got %d", len(data.Movements))
	}
}

// TestBackupService_CreateBackup tests creating a full backup
func TestBackupService_CreateBackup(t *testing.T) {
	db := setupFullTestDB(t)
	defer db.Close()

	// Create temp directories
	backupDir, err := os.MkdirTemp("", "backup_test")
	if err != nil {
		t.Fatalf("failed to create backup temp dir: %v", err)
	}
	defer os.RemoveAll(backupDir)

	uploadsDir, err := os.MkdirTemp("", "uploads_test")
	if err != nil {
		t.Fatalf("failed to create uploads temp dir: %v", err)
	}
	defer os.RemoveAll(uploadsDir)

	// Insert test user
	_, err = db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('admin@example.com', 'hash123', 'Admin User', 'admin', '2025-01-01 00:00:00', '2025-01-01 00:00:00')
	`)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// Create mock user repo
	userRepo := newMockUserRepo()
	userRepo.users[1] = &domain.User{
		ID:    1,
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  "admin",
	}

	svc := NewBackupService(db, "sqlite3", "test.db", backupDir, uploadsDir, userRepo, &mockAuditLogRepo{})

	filename, err := svc.CreateBackup(1)
	if err != nil {
		t.Fatalf("CreateBackup error: %v", err)
	}

	if filename == "" {
		t.Error("expected non-empty filename")
	}

	// Verify file was created
	backupPath := filepath.Join(backupDir, filename)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("backup file not found at %s", backupPath)
	}

	// Verify ZIP contents
	zipReader, err := zip.OpenReader(backupPath)
	if err != nil {
		t.Fatalf("failed to open backup ZIP: %v", err)
	}
	defer zipReader.Close()

	foundDataJSON := false
	for _, f := range zipReader.File {
		if f.Name == "backup_data.json" {
			foundDataJSON = true
			break
		}
	}
	if !foundDataJSON {
		t.Error("backup_data.json not found in ZIP")
	}
}

// TestBackupService_CreateBackup_UserNotFound tests backup creation with invalid user
func TestBackupService_CreateBackup_UserNotFound(t *testing.T) {
	db := setupFullTestDB(t)
	defer db.Close()

	backupDir, err := os.MkdirTemp("", "backup_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(backupDir)

	userRepo := newMockUserRepo()
	userRepo.getByIDError = fmt.Errorf("user not found") // Return error for non-existent users

	svc := NewBackupService(db, "sqlite3", "test.db", backupDir, "/tmp", userRepo, &mockAuditLogRepo{})

	_, err = svc.CreateBackup(999)
	if err == nil {
		t.Error("expected error for non-existent user")
	}
	if !strings.Contains(err.Error(), "user info") {
		t.Errorf("expected user error, got: %v", err)
	}
}

// TestBackupService_CreateBackup_WithUploads tests backup with upload files
func TestBackupService_CreateBackup_WithUploads(t *testing.T) {
	db := setupFullTestDB(t)
	defer db.Close()

	// Create temp directories
	backupDir, err := os.MkdirTemp("", "backup_test")
	if err != nil {
		t.Fatalf("failed to create backup temp dir: %v", err)
	}
	defer os.RemoveAll(backupDir)

	uploadsDir, err := os.MkdirTemp("", "uploads_test")
	if err != nil {
		t.Fatalf("failed to create uploads temp dir: %v", err)
	}
	defer os.RemoveAll(uploadsDir)

	// Create a test upload file
	testFile := filepath.Join(uploadsDir, "profile_1.jpg")
	if err := os.WriteFile(testFile, []byte("fake image data"), 0644); err != nil {
		t.Fatalf("failed to create test upload file: %v", err)
	}

	// Insert test user
	_, err = db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('admin@example.com', 'hash123', 'Admin User', 'admin', '2025-01-01 00:00:00', '2025-01-01 00:00:00')
	`)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	userRepo := newMockUserRepo()
	userRepo.users[1] = &domain.User{
		ID:    1,
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  "admin",
	}

	svc := NewBackupService(db, "sqlite3", "test.db", backupDir, uploadsDir, userRepo, &mockAuditLogRepo{})

	filename, err := svc.CreateBackup(1)
	if err != nil {
		t.Fatalf("CreateBackup error: %v", err)
	}

	// Verify ZIP contains upload file
	backupPath := filepath.Join(backupDir, filename)
	zipReader, err := zip.OpenReader(backupPath)
	if err != nil {
		t.Fatalf("failed to open backup ZIP: %v", err)
	}
	defer zipReader.Close()

	foundUpload := false
	for _, f := range zipReader.File {
		if strings.Contains(f.Name, "profile_1.jpg") {
			foundUpload = true
			break
		}
	}
	if !foundUpload {
		t.Error("upload file not found in backup ZIP")
	}
}

// TestBackupService_CreateSQLiteDump tests SQLite dump creation
func TestBackupService_CreateSQLiteDump(t *testing.T) {
	db := setupFullTestDB(t)
	defer db.Close()

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('test@example.com', 'hash123', 'Test User', 'user', '2025-01-01 00:00:00', '2025-01-01 00:00:00')
	`)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	svc := &BackupServiceImpl{
		db:       db,
		dbDriver: "sqlite3",
	}

	// Export data
	data, err := svc.exportAllTables()
	if err != nil {
		t.Fatalf("exportAllTables error: %v", err)
	}

	// Create temp file for SQLite dump
	tmpFile, err := os.CreateTemp("", "sqlitedump_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Create SQLite dump
	err = svc.createSQLiteDump(data, tmpPath)
	if err != nil {
		t.Fatalf("createSQLiteDump error: %v", err)
	}

	// Verify dump file exists
	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		t.Error("SQLite dump file not created")
	}

	// Verify dump contains data
	dumpDB, err := sql.Open("sqlite3", tmpPath)
	if err != nil {
		t.Fatalf("failed to open dump database: %v", err)
	}
	defer dumpDB.Close()

	var count int
	err = dumpDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query dump database: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 user in dump, got %d", count)
	}
}

// TestBackupService_RestoreBackup tests restoring from a backup file
func TestBackupService_RestoreBackup(t *testing.T) {
	// Create source database with data
	sourceDB := setupFullTestDB(t)
	defer sourceDB.Close()

	_, err := sourceDB.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('source@example.com', 'hash123', 'Source User', 'user', '2025-01-01 00:00:00', '2025-01-01 00:00:00')
	`)
	if err != nil {
		t.Fatalf("failed to insert source user: %v", err)
	}

	_, err = sourceDB.Exec(`
		INSERT INTO movements (name, type, is_standard, created_at, updated_at)
		VALUES ('Deadlift', 'weightlifting', 1, '2025-01-01 00:00:00', '2025-01-01 00:00:00')
	`)
	if err != nil {
		t.Fatalf("failed to insert source movement: %v", err)
	}

	// Create backup
	backupDir, err := os.MkdirTemp("", "backup_test")
	if err != nil {
		t.Fatalf("failed to create backup dir: %v", err)
	}
	defer os.RemoveAll(backupDir)

	uploadsDir, err := os.MkdirTemp("", "uploads_test")
	if err != nil {
		t.Fatalf("failed to create uploads dir: %v", err)
	}
	defer os.RemoveAll(uploadsDir)

	userRepo := newMockUserRepo()
	userRepo.users[1] = &domain.User{
		ID:    1,
		Email: "source@example.com",
		Name:  "Source User",
	}

	sourceSvc := NewBackupService(sourceDB, "sqlite3", "source.db", backupDir, uploadsDir, userRepo, &mockAuditLogRepo{})
	filename, err := sourceSvc.CreateBackup(1)
	if err != nil {
		t.Fatalf("CreateBackup error: %v", err)
	}

	// Create target database (empty)
	targetDB := setupFullTestDB(t)
	defer targetDB.Close()

	targetSvc := NewBackupService(targetDB, "sqlite3", "target.db", backupDir, uploadsDir, newMockUserRepo(), &mockAuditLogRepo{})

	// Restore backup
	result, err := targetSvc.RestoreBackup(filename, 1, domain.RestoreModeReplace)
	if err != nil {
		t.Fatalf("RestoreBackup error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil restore result")
	}

	// Verify data was restored
	var userCount int
	err = targetDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		t.Fatalf("failed to query target database: %v", err)
	}
	if userCount != 1 {
		t.Errorf("expected 1 user in target, got %d", userCount)
	}

	var movementCount int
	err = targetDB.QueryRow("SELECT COUNT(*) FROM movements").Scan(&movementCount)
	if err != nil {
		t.Fatalf("failed to query movements: %v", err)
	}
	if movementCount != 1 {
		t.Errorf("expected 1 movement in target, got %d", movementCount)
	}
}

// TestBackupService_RestoreBackup_FileNotFound tests restore with missing file
func TestBackupService_RestoreBackup_FileNotFound(t *testing.T) {
	db := setupFullTestDB(t)
	defer db.Close()

	svc := &BackupServiceImpl{
		db:        db,
		dbDriver:  "sqlite3",
		backupDir: "/tmp/nonexistent",
	}

	_, err := svc.RestoreBackup("nonexistent.zip", 1, domain.RestoreModeReplace)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

// TestBackupService_RestoreBackup_MergeMode tests restore with merge mode
func TestBackupService_RestoreBackup_MergeMode(t *testing.T) {
	// Create source database
	sourceDB := setupFullTestDB(t)
	defer sourceDB.Close()

	_, err := sourceDB.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('source@example.com', 'hash123', 'Source User', 'user', '2025-01-01 00:00:00', '2025-01-01 00:00:00')
	`)
	if err != nil {
		t.Fatalf("failed to insert source user: %v", err)
	}

	// Create backup
	backupDir, err := os.MkdirTemp("", "backup_test")
	if err != nil {
		t.Fatalf("failed to create backup dir: %v", err)
	}
	defer os.RemoveAll(backupDir)

	uploadsDir, err := os.MkdirTemp("", "uploads_test")
	if err != nil {
		t.Fatalf("failed to create uploads dir: %v", err)
	}
	defer os.RemoveAll(uploadsDir)

	userRepo := newMockUserRepo()
	userRepo.users[1] = &domain.User{
		ID:    1,
		Email: "source@example.com",
		Name:  "Source User",
	}

	sourceSvc := NewBackupService(sourceDB, "sqlite3", "source.db", backupDir, uploadsDir, userRepo, &mockAuditLogRepo{})
	filename, err := sourceSvc.CreateBackup(1)
	if err != nil {
		t.Fatalf("CreateBackup error: %v", err)
	}

	// Create target database with existing data
	targetDB := setupFullTestDB(t)
	defer targetDB.Close()

	_, err = targetDB.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('existing@example.com', 'hash456', 'Existing User', 'user', '2025-01-01 00:00:00', '2025-01-01 00:00:00')
	`)
	if err != nil {
		t.Fatalf("failed to insert existing user: %v", err)
	}

	targetSvc := NewBackupService(targetDB, "sqlite3", "target.db", backupDir, uploadsDir, newMockUserRepo(), &mockAuditLogRepo{})

	// Restore with merge mode
	result, err := targetSvc.RestoreBackup(filename, 1, domain.RestoreModeMerge)
	if err != nil {
		t.Fatalf("RestoreBackup merge error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil restore result")
	}

	// Verify both users exist (merged)
	var userCount int
	err = targetDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		t.Fatalf("failed to query target database: %v", err)
	}
	if userCount != 2 {
		t.Errorf("expected 2 users after merge, got %d", userCount)
	}
}

// TestBackupService_RestoreBackup_SkipMode tests restore with skip mode
func TestBackupService_RestoreBackup_SkipMode(t *testing.T) {
	// Create source database
	sourceDB := setupFullTestDB(t)
	defer sourceDB.Close()

	_, err := sourceDB.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('test@example.com', 'hash123', 'Source User', 'user', '2025-01-01 00:00:00', '2025-01-01 00:00:00')
	`)
	if err != nil {
		t.Fatalf("failed to insert source user: %v", err)
	}

	// Create backup
	backupDir, err := os.MkdirTemp("", "backup_test")
	if err != nil {
		t.Fatalf("failed to create backup dir: %v", err)
	}
	defer os.RemoveAll(backupDir)

	uploadsDir, err := os.MkdirTemp("", "uploads_test")
	if err != nil {
		t.Fatalf("failed to create uploads dir: %v", err)
	}
	defer os.RemoveAll(uploadsDir)

	userRepo := newMockUserRepo()
	userRepo.users[1] = &domain.User{
		ID:    1,
		Email: "test@example.com",
		Name:  "Source User",
	}

	sourceSvc := NewBackupService(sourceDB, "sqlite3", "source.db", backupDir, uploadsDir, userRepo, &mockAuditLogRepo{})
	filename, err := sourceSvc.CreateBackup(1)
	if err != nil {
		t.Fatalf("CreateBackup error: %v", err)
	}

	// Create target database with same email user (conflict)
	targetDB := setupFullTestDB(t)
	defer targetDB.Close()

	_, err = targetDB.Exec(`
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ('test@example.com', 'hash456', 'Existing User', 'user', '2025-01-01 00:00:00', '2025-01-01 00:00:00')
	`)
	if err != nil {
		t.Fatalf("failed to insert existing user: %v", err)
	}

	targetSvc := NewBackupService(targetDB, "sqlite3", "target.db", backupDir, uploadsDir, newMockUserRepo(), &mockAuditLogRepo{})

	// Restore with skip mode
	result, err := targetSvc.RestoreBackup(filename, 1, domain.RestoreModeSkip)
	if err != nil {
		t.Fatalf("RestoreBackup skip error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil restore result")
	}

	// Verify only 1 user (existing was kept, incoming was skipped)
	var userCount int
	err = targetDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		t.Fatalf("failed to query target database: %v", err)
	}
	if userCount != 1 {
		t.Errorf("expected 1 user after skip, got %d", userCount)
	}

	// Verify it's still the existing user
	var name string
	err = targetDB.QueryRow("SELECT name FROM users WHERE email = 'test@example.com'").Scan(&name)
	if err != nil {
		t.Fatalf("failed to query user name: %v", err)
	}
	if name != "Existing User" {
		t.Errorf("expected 'Existing User', got '%s'", name)
	}
}
