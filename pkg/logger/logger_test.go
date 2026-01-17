package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew_DefaultConfig(t *testing.T) {
	cfg := Config{
		Level:      "info",
		EnableFile: false,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer l.Close()

	if l.level != INFO {
		t.Errorf("logger level = %v, want %v", l.level, INFO)
	}

	if l.stdout == nil {
		t.Error("stdout logger should not be nil")
	}

	if l.file != nil {
		t.Error("file logger should be nil when EnableFile is false")
	}
}

func TestNew_WithFileLogging(t *testing.T) {
	// Create temp directory for log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "debug",
		EnableFile: true,
		FilePath:   logPath,
		MaxSizeMB:  1,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer l.Close()

	if l.level != DEBUG {
		t.Errorf("logger level = %v, want %v", l.level, DEBUG)
	}

	if l.file == nil {
		t.Error("file logger should not be nil when EnableFile is true")
	}

	if l.logPath != logPath {
		t.Errorf("logPath = %q, want %q", l.logPath, logPath)
	}

	// Verify log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("log file should have been created")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", DEBUG},
		{"info", INFO},
		{"warn", WARN},
		{"error", ERROR},
		{"DEBUG", INFO},   // Invalid case - defaults to INFO
		{"unknown", INFO}, // Invalid level - defaults to INFO
		{"", INFO},        // Empty - defaults to INFO
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseLevel(tt.input); got != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	// Create logger with WARN level
	cfg := Config{
		Level:      "warn",
		EnableFile: false,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer l.Close()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Debug and Info should not be logged
	l.Debug("debug message")
	l.Info("info message")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()

	// Debug and Info should be filtered out at WARN level
	if strings.Contains(output, "debug message") {
		t.Error("DEBUG message should be filtered at WARN level")
	}
	if strings.Contains(output, "info message") {
		t.Error("INFO message should be filtered at WARN level")
	}
}

func TestLogger_LogMethods(t *testing.T) {
	// Create temp directory for log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "debug",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer l.Close()

	// Test all log methods
	l.Debug("debug %s", "test")
	l.Info("info %s", "test")
	l.Warn("warn %s", "test")
	l.Error("error %s", "test")

	// Close to flush
	l.Close()

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify all messages were logged
	tests := []struct {
		level   string
		message string
	}{
		{"DEBUG", "debug test"},
		{"INFO", "info test"},
		{"WARN", "warn test"},
		{"ERROR", "error test"},
	}

	for _, tt := range tests {
		if !strings.Contains(logContent, tt.level) {
			t.Errorf("Log file should contain [%s]", tt.level)
		}
		if !strings.Contains(logContent, tt.message) {
			t.Errorf("Log file should contain %q", tt.message)
		}
	}
}

func TestLogger_Printf_Println(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "debug",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	l.Printf("printf %s %d", "test", 123)
	l.Println("println test")

	l.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	if !strings.Contains(logContent, "printf test 123") {
		t.Error("Printf should log message")
	}
	if !strings.Contains(logContent, "println test") {
		t.Error("Println should log message")
	}
}

func TestLogger_Writer(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "debug",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	writer := l.Writer()
	if writer == nil {
		t.Fatal("Writer() should return non-nil io.Writer")
	}

	// Write to the writer
	n, err := writer.Write([]byte("writer test message"))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != 19 { // len("writer test message")
		t.Errorf("Write() returned %d, want %d", n, 19)
	}

	l.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "writer test message") {
		t.Error("Writer should log message")
	}
}

func TestLogger_Close(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "info",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Close should not error
	if err := l.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Close again should not error (file already closed, returns nil)
	if err := l.Close(); err != nil {
		// This might error because file is already closed, which is acceptable
		t.Logf("Second Close() returned: %v (acceptable)", err)
	}
}

func TestLogger_Close_NoFile(t *testing.T) {
	cfg := Config{
		Level:      "info",
		EnableFile: false,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Close should not error when no file is open
	if err := l.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestLevelNames(t *testing.T) {
	// Verify all levels have names
	levels := []Level{DEBUG, INFO, WARN, ERROR}
	expectedNames := []string{"DEBUG", "INFO", "WARN", "ERROR"}

	for i, level := range levels {
		name, ok := levelNames[level]
		if !ok {
			t.Errorf("levelNames[%v] not found", level)
			continue
		}
		if name != expectedNames[i] {
			t.Errorf("levelNames[%v] = %q, want %q", level, name, expectedNames[i])
		}
	}
}

func TestNew_InvalidPath(t *testing.T) {
	// Try to create log in a path that doesn't exist and can't be created
	cfg := Config{
		Level:      "info",
		EnableFile: true,
		FilePath:   "/nonexistent/path/that/cannot/be/created/test.log",
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("New() should error when log directory cannot be created")
	}
}

func TestLogger_LogFormat(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "info",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	l.Info("test message")
	l.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logLine := string(content)

	// Log format should be: "2024-11-09 15:04:05 [INFO] message"
	// Check for date format (YYYY-MM-DD)
	if !strings.Contains(logLine, "-") {
		t.Error("Log line should contain date separator")
	}

	// Check for level marker
	if !strings.Contains(logLine, "[INFO]") {
		t.Error("Log line should contain [INFO] level marker")
	}

	// Check for message
	if !strings.Contains(logLine, "test message") {
		t.Error("Log line should contain the message")
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
	}{
		{"json", FormatJSON},
		{"text", FormatText},
		{"JSON", FormatText}, // Case-sensitive, defaults to text
		{"", FormatText},     // Empty - defaults to text
		{"invalid", FormatText},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseFormat(tt.input); got != tt.expected {
				t.Errorf("parseFormat(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLogger_JSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "info",
		Format:     "json",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	l.Info("test message with %s", "args")
	l.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logLine := strings.TrimSpace(string(content))

	// Verify it's valid JSON
	var entry LogEntry
	if err := json.Unmarshal([]byte(logLine), &entry); err != nil {
		t.Fatalf("Log line should be valid JSON: %v\nGot: %s", err, logLine)
	}

	// Verify fields
	if entry.Level != "INFO" {
		t.Errorf("Level = %q, want %q", entry.Level, "INFO")
	}
	if entry.Message != "test message with args" {
		t.Errorf("Message = %q, want %q", entry.Message, "test message with args")
	}
	if entry.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
}

func TestLogger_JSONFormatWithFields(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "info",
		Format:     "json",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	fields := Fields{
		"user_id": 123,
		"action":  "login",
		"ip":      "192.168.1.1",
	}
	l.InfoWithFields(fields, "user logged in")
	l.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logLine := strings.TrimSpace(string(content))

	// Verify it's valid JSON
	var entry LogEntry
	if err := json.Unmarshal([]byte(logLine), &entry); err != nil {
		t.Fatalf("Log line should be valid JSON: %v\nGot: %s", err, logLine)
	}

	// Verify message
	if entry.Message != "user logged in" {
		t.Errorf("Message = %q, want %q", entry.Message, "user logged in")
	}

	// Verify fields
	if entry.Fields == nil {
		t.Fatal("Fields should not be nil")
	}
	if entry.Fields["action"] != "login" {
		t.Errorf("Fields[action] = %v, want %q", entry.Fields["action"], "login")
	}
	if entry.Fields["ip"] != "192.168.1.1" {
		t.Errorf("Fields[ip] = %v, want %q", entry.Fields["ip"], "192.168.1.1")
	}
	// user_id may be float64 due to JSON unmarshaling
	if userID, ok := entry.Fields["user_id"].(float64); !ok || userID != 123 {
		t.Errorf("Fields[user_id] = %v, want %v", entry.Fields["user_id"], 123)
	}
}

func TestLogger_TextFormatWithFields(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "info",
		Format:     "text", // Explicitly text format
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	fields := Fields{
		"user_id": 123,
		"action":  "login",
	}
	l.InfoWithFields(fields, "user logged in")
	l.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logLine := string(content)

	// Should contain the message
	if !strings.Contains(logLine, "user logged in") {
		t.Error("Log line should contain the message")
	}

	// Should contain fields as key=value pairs
	if !strings.Contains(logLine, "user_id=123") {
		t.Error("Log line should contain user_id=123")
	}
	if !strings.Contains(logLine, "action=login") {
		t.Error("Log line should contain action=login")
	}
}

func TestLogger_WithChaining(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "info",
		Format:     "json",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Create a logger with base fields
	requestLogger := l.With(Fields{"request_id": "abc123"})

	// Add more fields
	userLogger := requestLogger.With(Fields{"user_id": 456})

	// Log with the chained logger
	userLogger.Info("processing request")
	l.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logLine := strings.TrimSpace(string(content))

	var entry LogEntry
	if err := json.Unmarshal([]byte(logLine), &entry); err != nil {
		t.Fatalf("Log line should be valid JSON: %v", err)
	}

	// Should have both fields from chaining
	if entry.Fields["request_id"] != "abc123" {
		t.Errorf("Fields[request_id] = %v, want %q", entry.Fields["request_id"], "abc123")
	}
	if userID, ok := entry.Fields["user_id"].(float64); !ok || userID != 456 {
		t.Errorf("Fields[user_id] = %v, want %v", entry.Fields["user_id"], 456)
	}
}

func TestLogger_AllLevelsWithFields(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "debug",
		Format:     "json",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	fields := Fields{"test": "value"}
	l.DebugWithFields(fields, "debug message")
	l.InfoWithFields(fields, "info message")
	l.WarnWithFields(fields, "warn message")
	l.ErrorWithFields(fields, "error message")
	l.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 4 {
		t.Fatalf("Expected 4 log lines, got %d", len(lines))
	}

	expectedLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	for i, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("Line %d should be valid JSON: %v", i, err)
		}
		if entry.Level != expectedLevels[i] {
			t.Errorf("Line %d: Level = %q, want %q", i, entry.Level, expectedLevels[i])
		}
		if entry.Fields["test"] != "value" {
			t.Errorf("Line %d: Fields[test] = %v, want %q", i, entry.Fields["test"], "value")
		}
	}
}

func TestFieldLogger_AllLevels(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "debug",
		Format:     "json",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	fl := l.With(Fields{"component": "test"})
	fl.Debug("debug via FieldLogger")
	fl.Info("info via FieldLogger")
	fl.Warn("warn via FieldLogger")
	fl.Error("error via FieldLogger")
	l.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 4 {
		t.Fatalf("Expected 4 log lines, got %d", len(lines))
	}

	for i, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("Line %d should be valid JSON: %v", i, err)
		}
		if entry.Fields["component"] != "test" {
			t.Errorf("Line %d: Fields[component] = %v, want %q", i, entry.Fields["component"], "test")
		}
	}
}

func TestLogger_RotateIfNeeded_NoFile(t *testing.T) {
	// Logger without file logging should not error on rotation
	cfg := Config{
		Level:      "info",
		EnableFile: false,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer l.Close()

	// rotateIfNeeded should return nil when no file is configured
	if err := l.rotateIfNeeded(); err != nil {
		t.Errorf("rotateIfNeeded() error = %v, want nil", err)
	}
}

func TestLogger_RotateIfNeeded_UnderSize(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "info",
		EnableFile: true,
		FilePath:   logPath,
		MaxSizeMB:  1, // 1 MB
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer l.Close()

	// Write a small amount of data
	l.Info("small message")

	// rotateIfNeeded should not rotate (file is under max size)
	if err := l.rotateIfNeeded(); err != nil {
		t.Errorf("rotateIfNeeded() error = %v", err)
	}

	// Verify no backup file was created
	matches, _ := filepath.Glob(filepath.Join(tmpDir, "test.log.*"))
	if len(matches) > 0 {
		t.Errorf("No backup files should exist, found %d", len(matches))
	}
}

func TestLogger_RotateIfNeeded_ExceedsSize(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Set a very small max size to trigger rotation easily
	cfg := Config{
		Level:      "info",
		EnableFile: true,
		FilePath:   logPath,
		MaxSizeMB:  0, // 0 MB means any write triggers rotation (in bytes: 0)
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer l.Close()

	// Write some data to the file first
	l.Info("first message")

	// Now set maxSize to something very small to trigger rotation
	l.maxSize = 10 // 10 bytes

	// Write more to exceed the size
	l.Info("another message that will exceed the size limit")

	// Force a rotation check
	if err := l.rotateIfNeeded(); err != nil {
		t.Errorf("rotateIfNeeded() error = %v", err)
	}

	// Give goroutine time to run cleanup
	// Not strictly necessary but helps ensure cleanup runs
	// The backup should exist
	matches, _ := filepath.Glob(filepath.Join(tmpDir, "test.log.*"))
	if len(matches) == 0 {
		t.Error("Expected backup file to be created after rotation")
	}

	// Original file should be recreated (empty or with new content)
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file should be recreated after rotation")
	}
}

func TestLogger_CleanupOldBackups_KeepsRecent(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create the main log file
	if err := os.WriteFile(logPath, []byte("main log"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create 3 backup files (should all be kept)
	for i := 1; i <= 3; i++ {
		backupPath := filepath.Join(tmpDir, "test.log."+strings.Repeat("0", i))
		if err := os.WriteFile(backupPath, []byte("backup"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	l := &Logger{
		logPath: logPath,
	}

	l.cleanupOldBackups()

	// All 3 backups should still exist
	matches, _ := filepath.Glob(filepath.Join(tmpDir, "test.log.*"))
	if len(matches) != 3 {
		t.Errorf("Expected 3 backup files, found %d", len(matches))
	}
}

func TestLogger_CleanupOldBackups_RemovesOldest(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create the main log file
	if err := os.WriteFile(logPath, []byte("main log"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create 5 backup files with different timestamps
	backups := []string{
		"test.log.20240101-100000", // oldest
		"test.log.20240102-100000",
		"test.log.20240103-100000",
		"test.log.20240104-100000",
		"test.log.20240105-100000", // newest
	}

	for _, backup := range backups {
		backupPath := filepath.Join(tmpDir, backup)
		if err := os.WriteFile(backupPath, []byte("backup content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	l := &Logger{
		logPath: logPath,
	}

	l.cleanupOldBackups()

	// Should have only 3 backup files remaining
	matches, _ := filepath.Glob(filepath.Join(tmpDir, "test.log.*"))
	if len(matches) != 3 {
		t.Errorf("Expected 3 backup files after cleanup, found %d", len(matches))
	}
}

func TestLogger_CleanupOldBackups_NoBackups(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create only the main log file (no backups)
	if err := os.WriteFile(logPath, []byte("main log"), 0644); err != nil {
		t.Fatal(err)
	}

	l := &Logger{
		logPath: logPath,
	}

	// Should not panic or error
	l.cleanupOldBackups()

	// Main log should still exist
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Main log file should still exist")
	}
}

func TestLogger_CleanupOldBackups_InvalidPattern(t *testing.T) {
	l := &Logger{
		logPath: "/nonexistent/path/test.log",
	}

	// Should not panic when directory doesn't exist
	l.cleanupOldBackups()
}

func TestLogger_FormatJSON_WithComplexFields(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "info",
		Format:     "json",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Log with complex field types
	fields := Fields{
		"string": "value",
		"int":    123,
		"float":  3.14,
		"bool":   true,
		"nested": map[string]interface{}{"key": "val"},
		"array":  []int{1, 2, 3},
	}
	l.InfoWithFields(fields, "complex fields test")
	l.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logLine := strings.TrimSpace(string(content))

	var entry LogEntry
	if err := json.Unmarshal([]byte(logLine), &entry); err != nil {
		t.Fatalf("Log line should be valid JSON: %v\nGot: %s", err, logLine)
	}

	// Verify fields are preserved
	if entry.Fields["string"] != "value" {
		t.Errorf("string field = %v, want 'value'", entry.Fields["string"])
	}
	if entry.Fields["bool"] != true {
		t.Errorf("bool field = %v, want true", entry.Fields["bool"])
	}
}

func TestLogger_New_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a path with nested directories that don't exist
	logPath := filepath.Join(tmpDir, "subdir1", "subdir2", "test.log")

	cfg := Config{
		Level:      "info",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() should create directories: %v", err)
	}
	defer l.Close()

	// Verify the directory was created
	dir := filepath.Dir(logPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Directory should have been created")
	}

	// Verify the file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file should have been created")
	}
}

func TestLogger_New_DefaultMaxSize(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "info",
		EnableFile: true,
		FilePath:   logPath,
		// MaxSizeMB not specified, should default to 100MB
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer l.Close()

	// Default max size is 100MB when file logging is enabled
	expectedSize := int64(100 * 1024 * 1024)
	if l.maxSize != expectedSize {
		t.Errorf("maxSize = %d, want %d", l.maxSize, expectedSize)
	}
}

func TestFieldLogger_With(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "info",
		Format:     "json",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Create a FieldLogger and chain With
	fl1 := l.With(Fields{"a": 1})
	fl2 := fl1.With(Fields{"b": 2})
	fl3 := fl2.With(Fields{"c": 3})

	fl3.Info("chained fields")
	l.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	var entry LogEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(content))), &entry); err != nil {
		t.Fatalf("Log line should be valid JSON: %v", err)
	}

	// All chained fields should be present
	if entry.Fields["a"] != float64(1) {
		t.Errorf("Fields[a] = %v, want 1", entry.Fields["a"])
	}
	if entry.Fields["b"] != float64(2) {
		t.Errorf("Fields[b] = %v, want 2", entry.Fields["b"])
	}
	if entry.Fields["c"] != float64(3) {
		t.Errorf("Fields[c] = %v, want 3", entry.Fields["c"])
	}
}

func TestLogger_FormatText_WithFields(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "info",
		Format:     "text",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Log with fields containing special characters
	fields := Fields{
		"path":  "/api/test?q=hello",
		"empty": "",
	}
	l.InfoWithFields(fields, "test with special chars")
	l.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logLine := string(content)

	// Should contain the message
	if !strings.Contains(logLine, "test with special chars") {
		t.Error("Log line should contain the message")
	}

	// Should contain field with special chars
	if !strings.Contains(logLine, "path=/api/test?q=hello") {
		t.Error("Log line should contain path field")
	}
}

func TestLogger_LogWithFields_NilFields(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "info",
		Format:     "json",
		EnableFile: true,
		FilePath:   logPath,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Log with nil fields should not panic
	l.InfoWithFields(nil, "no fields message")
	l.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	var entry LogEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(content))), &entry); err != nil {
		t.Fatalf("Log line should be valid JSON: %v", err)
	}

	if entry.Message != "no fields message" {
		t.Errorf("Message = %q, want %q", entry.Message, "no fields message")
	}
}
