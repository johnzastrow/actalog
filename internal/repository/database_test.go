package repository

import (
	"strings"
	"testing"
)

func TestBuildDSN_SQLite(t *testing.T) {
	dsn := BuildDSN("sqlite3", "", 0, "", "", "/path/to/db.sqlite", "", "")
	if dsn != "/path/to/db.sqlite" {
		t.Errorf("Expected '/path/to/db.sqlite', got '%s'", dsn)
	}
}

func TestBuildDSN_Postgres(t *testing.T) {
	dsn := BuildDSN("postgres", "localhost", 5432, "user", "password", "testdb", "disable", "")
	expected := "postgres://user:password@localhost:5432/testdb?sslmode=disable&connect_timeout=30"
	if dsn != expected {
		t.Errorf("Expected '%s', got '%s'", expected, dsn)
	}
}

func TestBuildDSN_Postgres_WithSchema(t *testing.T) {
	dsn := BuildDSN("postgres", "localhost", 5432, "user", "password", "testdb", "disable", "myschema")
	expected := "postgres://user:password@localhost:5432/testdb?sslmode=disable&connect_timeout=30&search_path=myschema"
	if dsn != expected {
		t.Errorf("Expected '%s', got '%s'", expected, dsn)
	}
}

func TestBuildDSN_Postgres_NoPassword(t *testing.T) {
	dsn := BuildDSN("postgres", "localhost", 5432, "user", "", "testdb", "disable", "")
	expected := "postgres://user@localhost:5432/testdb?sslmode=disable&connect_timeout=30"
	if dsn != expected {
		t.Errorf("Expected '%s', got '%s'", expected, dsn)
	}
}

func TestBuildDSN_MySQL(t *testing.T) {
	dsn := BuildDSN("mysql", "localhost", 3306, "user", "password", "testdb", "", "")
	expected := "user:password@tcp(localhost:3306)/testdb?parseTime=true&multiStatements=true&timeout=30s&readTimeout=30s&writeTimeout=30s"
	if dsn != expected {
		t.Errorf("Expected '%s', got '%s'", expected, dsn)
	}
}

func TestBuildDSN_UnknownDriver(t *testing.T) {
	dsn := BuildDSN("unknown", "", 0, "", "", "somevalue", "", "")
	if dsn != "somevalue" {
		t.Errorf("Expected 'somevalue', got '%s'", dsn)
	}
}

func TestSanitizeDSN_MySQL(t *testing.T) {
	dsn := "user:secretpassword@tcp(localhost:3306)/testdb"
	sanitized := sanitizeDSN("mysql", dsn)
	expected := "user:****@tcp(localhost:3306)/testdb"
	if sanitized != expected {
		t.Errorf("Expected '%s', got '%s'", expected, sanitized)
	}
}

func TestSanitizeDSN_Postgres(t *testing.T) {
	dsn := "postgres://user:secretpassword@localhost:5432/testdb"
	sanitized := sanitizeDSN("postgres", dsn)
	expected := "postgres://user:****@localhost:5432/testdb"
	if sanitized != expected {
		t.Errorf("Expected '%s', got '%s'", expected, sanitized)
	}
}

func TestSanitizeDSN_Postgres_NoPassword(t *testing.T) {
	dsn := "postgres://user@localhost:5432/testdb"
	sanitized := sanitizeDSN("postgres", dsn)
	// Should return original DSN if no password present
	if sanitized != dsn {
		t.Errorf("Expected '%s', got '%s'", dsn, sanitized)
	}
}

func TestSanitizeDSN_SQLite(t *testing.T) {
	dsn := "/path/to/database.sqlite"
	sanitized := sanitizeDSN("sqlite3", dsn)
	// SQLite should return DSN unchanged
	if sanitized != dsn {
		t.Errorf("Expected '%s', got '%s'", dsn, sanitized)
	}
}

func TestSanitizeDSN_MySQL_NoPassword(t *testing.T) {
	dsn := "user@tcp(localhost:3306)/testdb"
	sanitized := sanitizeDSN("mysql", dsn)
	// Should return original DSN if no colon before @
	if sanitized != dsn {
		t.Errorf("Expected '%s', got '%s'", dsn, sanitized)
	}
}

func TestGetBoolValue_SQLite(t *testing.T) {
	trueVal := getBoolValue("sqlite3", true)
	if trueVal != "1" {
		t.Errorf("Expected '1' for true, got '%s'", trueVal)
	}

	falseVal := getBoolValue("sqlite3", false)
	if falseVal != "0" {
		t.Errorf("Expected '0' for false, got '%s'", falseVal)
	}
}

func TestGetBoolValue_Postgres(t *testing.T) {
	trueVal := getBoolValue("postgres", true)
	if trueVal != "TRUE" {
		t.Errorf("Expected 'TRUE' for true, got '%s'", trueVal)
	}

	falseVal := getBoolValue("postgres", false)
	if falseVal != "FALSE" {
		t.Errorf("Expected 'FALSE' for false, got '%s'", falseVal)
	}
}

func TestGetBoolValue_MySQL(t *testing.T) {
	trueVal := getBoolValue("mysql", true)
	if trueVal != "TRUE" {
		t.Errorf("Expected 'TRUE' for true, got '%s'", trueVal)
	}

	falseVal := getBoolValue("mysql", false)
	if falseVal != "FALSE" {
		t.Errorf("Expected 'FALSE' for false, got '%s'", falseVal)
	}
}

func TestGetPlaceholders_SQLite(t *testing.T) {
	placeholders := getPlaceholders("sqlite3", 3)
	if len(placeholders) != 3 {
		t.Errorf("Expected 3 placeholders, got %d", len(placeholders))
	}
	for _, p := range placeholders {
		if p != "?" {
			t.Errorf("Expected '?', got '%s'", p)
		}
	}
}

func TestGetPlaceholders_Postgres(t *testing.T) {
	placeholders := getPlaceholders("postgres", 3)
	if len(placeholders) != 3 {
		t.Errorf("Expected 3 placeholders, got %d", len(placeholders))
	}
	expected := []string{"$1", "$2", "$3"}
	for i, p := range placeholders {
		if p != expected[i] {
			t.Errorf("Expected '%s', got '%s'", expected[i], p)
		}
	}
}

func TestGetPlaceholders_MySQL(t *testing.T) {
	placeholders := getPlaceholders("mysql", 2)
	if len(placeholders) != 2 {
		t.Errorf("Expected 2 placeholders, got %d", len(placeholders))
	}
	for _, p := range placeholders {
		if p != "?" {
			t.Errorf("Expected '?', got '%s'", p)
		}
	}
}

func TestGetTimestampFunc_SQLite(t *testing.T) {
	// Save and restore original driver
	originalDriver := currentDriver
	defer func() { currentDriver = originalDriver }()

	currentDriver = "sqlite3"
	result := getTimestampFunc()
	if result != "datetime('now')" {
		t.Errorf("Expected \"datetime('now')\", got '%s'", result)
	}
}

func TestGetTimestampFunc_Postgres(t *testing.T) {
	originalDriver := currentDriver
	defer func() { currentDriver = originalDriver }()

	currentDriver = "postgres"
	result := getTimestampFunc()
	if result != "CURRENT_TIMESTAMP" {
		t.Errorf("Expected 'CURRENT_TIMESTAMP', got '%s'", result)
	}
}

func TestGetTimestampFunc_MySQL(t *testing.T) {
	originalDriver := currentDriver
	defer func() { currentDriver = originalDriver }()

	currentDriver = "mysql"
	result := getTimestampFunc()
	if result != "NOW()" {
		t.Errorf("Expected 'NOW()', got '%s'", result)
	}
}

func TestGetTimestampFunc_UnknownDriver(t *testing.T) {
	originalDriver := currentDriver
	defer func() { currentDriver = originalDriver }()

	currentDriver = "unknown"
	result := getTimestampFunc()
	if result != "CURRENT_TIMESTAMP" {
		t.Errorf("Expected 'CURRENT_TIMESTAMP' as default, got '%s'", result)
	}
}

func TestRebindQuery_SQLite(t *testing.T) {
	originalDriver := currentDriver
	defer func() { currentDriver = originalDriver }()

	currentDriver = "sqlite3"
	query := "SELECT * FROM users WHERE id = ? AND name = ?"
	result := rebindQuery(query)
	// For non-postgres, query should be unchanged
	if result != query {
		t.Errorf("Expected unchanged query for SQLite, got '%s'", result)
	}
}

func TestRebindQuery_Postgres(t *testing.T) {
	originalDriver := currentDriver
	defer func() { currentDriver = originalDriver }()

	currentDriver = "postgres"
	query := "SELECT * FROM users WHERE id = ? AND name = ?"
	result := rebindQuery(query)
	expected := "SELECT * FROM users WHERE id = $1 AND name = $2"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestRebindQuery_Postgres_NoPlaceholders(t *testing.T) {
	originalDriver := currentDriver
	defer func() { currentDriver = originalDriver }()

	currentDriver = "postgres"
	query := "SELECT * FROM users"
	result := rebindQuery(query)
	if result != query {
		t.Errorf("Expected unchanged query, got '%s'", result)
	}
}

func TestRebindQuery_Postgres_MultiplePlaceholders(t *testing.T) {
	originalDriver := currentDriver
	defer func() { currentDriver = originalDriver }()

	currentDriver = "postgres"
	query := "INSERT INTO users (a, b, c, d, e) VALUES (?, ?, ?, ?, ?)"
	result := rebindQuery(query)
	expected := "INSERT INTO users (a, b, c, d, e) VALUES ($1, $2, $3, $4, $5)"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestRebindQuery_MySQL(t *testing.T) {
	originalDriver := currentDriver
	defer func() { currentDriver = originalDriver }()

	currentDriver = "mysql"
	query := "SELECT * FROM users WHERE id = ? AND name = ?"
	result := rebindQuery(query)
	// For MySQL, query should be unchanged
	if result != query {
		t.Errorf("Expected unchanged query for MySQL, got '%s'", result)
	}
}

// Test logging functions - just ensure they don't panic
func TestLogInfo(t *testing.T) {
	// Just verify it doesn't panic
	logInfo("TEST", "Test message %s", "arg")
}

func TestLogSuccess(t *testing.T) {
	// Just verify it doesn't panic
	logSuccess("TEST", "Test message %s", "arg")
}

func TestLogError(t *testing.T) {
	// Just verify it doesn't panic
	logError("TEST", "Test message %s", "arg")
}

func TestGetSQLiteSchema(t *testing.T) {
	schema := getSQLiteSchema()

	// Verify schema is not empty
	if schema == "" {
		t.Error("getSQLiteSchema() returned empty string")
	}

	// Verify schema contains expected tables
	expectedTables := []string{
		"CREATE TABLE IF NOT EXISTS users",
		"CREATE TABLE IF NOT EXISTS refresh_tokens",
		"CREATE TABLE IF NOT EXISTS user_settings",
		"CREATE TABLE IF NOT EXISTS audit_logs",
		"CREATE TABLE IF NOT EXISTS workouts",
		"CREATE TABLE IF NOT EXISTS movements",
		"CREATE TABLE IF NOT EXISTS workout_movements",
	}

	for _, table := range expectedTables {
		if !strings.Contains(schema, table) {
			t.Errorf("getSQLiteSchema() missing expected table: %s", table)
		}
	}
}

func TestGetPostgreSQLSchema(t *testing.T) {
	schema := getPostgreSQLSchema()

	// Verify schema is not empty
	if schema == "" {
		t.Error("getPostgreSQLSchema() returned empty string")
	}

	// Verify schema contains expected tables
	expectedTables := []string{
		"CREATE TABLE IF NOT EXISTS users",
		"CREATE TABLE IF NOT EXISTS refresh_tokens",
		"CREATE TABLE IF NOT EXISTS user_settings",
		"CREATE TABLE IF NOT EXISTS audit_logs",
		"CREATE TABLE IF NOT EXISTS workouts",
		"CREATE TABLE IF NOT EXISTS movements",
		"CREATE TABLE IF NOT EXISTS workout_movements",
	}

	for _, table := range expectedTables {
		if !strings.Contains(schema, table) {
			t.Errorf("getPostgreSQLSchema() missing expected table: %s", table)
		}
	}

	// Verify PostgreSQL-specific syntax
	if !strings.Contains(schema, "BIGSERIAL PRIMARY KEY") {
		t.Error("getPostgreSQLSchema() should use BIGSERIAL for auto-increment")
	}
	if !strings.Contains(schema, "TIMESTAMP") {
		t.Error("getPostgreSQLSchema() should use TIMESTAMP for datetime fields")
	}
}

func TestGetMySQLSchema(t *testing.T) {
	schema := getMySQLSchema()

	// Verify schema is not empty
	if schema == "" {
		t.Error("getMySQLSchema() returned empty string")
	}

	// Verify schema contains expected tables
	expectedTables := []string{
		"CREATE TABLE IF NOT EXISTS users",
		"CREATE TABLE IF NOT EXISTS refresh_tokens",
		"CREATE TABLE IF NOT EXISTS user_settings",
		"CREATE TABLE IF NOT EXISTS audit_logs",
		"CREATE TABLE IF NOT EXISTS workouts",
		"CREATE TABLE IF NOT EXISTS movements",
		"CREATE TABLE IF NOT EXISTS workout_movements",
	}

	for _, table := range expectedTables {
		if !strings.Contains(schema, table) {
			t.Errorf("getMySQLSchema() missing expected table: %s", table)
		}
	}

	// Verify MySQL-specific syntax
	if !strings.Contains(schema, "AUTO_INCREMENT") {
		t.Error("getMySQLSchema() should use AUTO_INCREMENT for auto-increment")
	}
	if !strings.Contains(schema, "ENGINE=InnoDB") {
		t.Error("getMySQLSchema() should use InnoDB engine")
	}
	if !strings.Contains(schema, "utf8mb4") {
		t.Error("getMySQLSchema() should use utf8mb4 charset")
	}
}
