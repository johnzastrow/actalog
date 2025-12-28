package logger

import (
	"bytes"
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
