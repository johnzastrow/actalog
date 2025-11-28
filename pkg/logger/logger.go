// Package logger provides structured logging with optional file output
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents a log level
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

// Logger handles application logging
type Logger struct {
	level      Level
	stdout     *log.Logger
	file       *log.Logger
	fileHandle *os.File
	mu         sync.Mutex
	maxSize    int64  // Max file size in bytes before rotation
	logPath    string // Path to log file
}

// Config holds logger configuration
type Config struct {
	Level      string // "debug", "info", "warn", "error"
	EnableFile bool   // Enable file logging
	FilePath   string // Path to log file (default: ./logs/actalog.log)
	MaxSizeMB  int    // Max log file size in MB before rotation (default: 100)
	MaxBackups int    // Number of old log files to keep (default: 3)
}

// New creates a new logger with the given configuration
func New(cfg Config) (*Logger, error) {
	l := &Logger{
		level:   parseLevel(cfg.Level),
		maxSize: int64(cfg.MaxSizeMB) * 1024 * 1024, // Convert MB to bytes
	}

	// Always set up stdout logger
	l.stdout = log.New(os.Stdout, "", 0)

	// Set up file logger if enabled
	if cfg.EnableFile {
		// Use provided path or default
		logPath := cfg.FilePath
		if logPath == "" {
			// Default to logs/actalog.log in the directory where the binary runs
			execDir, err := os.Executable()
			if err != nil {
				execDir = "."
			} else {
				execDir = filepath.Dir(execDir)
			}
			logPath = filepath.Join(execDir, "logs", "actalog.log")
		}

		l.logPath = logPath

		// Create log directory if it doesn't exist
		logDir := filepath.Dir(logPath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		l.fileHandle = f
		l.file = log.New(f, "", 0)

		// Set default max size if not specified
		if l.maxSize == 0 {
			l.maxSize = 100 * 1024 * 1024 // 100MB default
		}
	}

	return l, nil
}

// parseLevel converts a string to a log level
func parseLevel(level string) Level {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// Close closes the log file if it's open
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.fileHandle != nil {
		return l.fileHandle.Close()
	}
	return nil
}

// log writes a log message at the specified level
func (l *Logger) log(level Level, format string, v ...interface{}) {
	if level < l.level {
		return // Skip if below configured level
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Format: 2024-11-09 15:04:05 [INFO] message
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := levelNames[level]
	message := fmt.Sprintf(format, v...)
	logLine := fmt.Sprintf("%s [%s] %s", timestamp, levelName, message)

	// Write to stdout
	l.stdout.Println(logLine)

	// Write to file if enabled
	if l.file != nil {
		l.file.Println(logLine)

		// Check if rotation is needed
		if err := l.rotateIfNeeded(); err != nil {
			// Log rotation error to stdout only
			l.stdout.Printf("%s [ERROR] Failed to rotate log file: %v", timestamp, err)
		}
	}
}

// rotateIfNeeded rotates the log file if it exceeds maxSize
// Must be called with mutex locked
func (l *Logger) rotateIfNeeded() error {
	if l.fileHandle == nil || l.logPath == "" {
		return nil
	}

	// Check file size
	fi, err := l.fileHandle.Stat()
	if err != nil {
		return err
	}

	if fi.Size() < l.maxSize {
		return nil // No rotation needed
	}

	// Close current file
	if err := l.fileHandle.Close(); err != nil {
		return err
	}

	// Rename current file to backup with timestamp
	backupPath := fmt.Sprintf("%s.%s", l.logPath, time.Now().Format("20060102-150405"))
	if err := os.Rename(l.logPath, backupPath); err != nil {
		return err
	}

	// Open new file
	f, err := os.OpenFile(l.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.fileHandle = f
	l.file = log.New(f, "", 0)

	// Clean up old backups (keep last 3)
	go l.cleanupOldBackups()

	return nil
}

// cleanupOldBackups removes old backup files, keeping only the most recent ones
func (l *Logger) cleanupOldBackups() {
	logDir := filepath.Dir(l.logPath)
	baseName := filepath.Base(l.logPath)
	pattern := fmt.Sprintf("%s.*", baseName)

	matches, err := filepath.Glob(filepath.Join(logDir, pattern))
	if err != nil {
		return
	}

	// Keep only the 3 most recent backups
	if len(matches) <= 3 {
		return
	}

	// Sort by modification time and delete oldest
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	var files []fileInfo
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		files = append(files, fileInfo{path: match, modTime: info.ModTime()})
	}

	// Simple bubble sort (good enough for small number of files)
	for i := 0; i < len(files)-1; i++ {
		for j := 0; j < len(files)-i-1; j++ {
			if files[j].modTime.After(files[j+1].modTime) {
				files[j], files[j+1] = files[j+1], files[j]
			}
		}
	}

	// Delete all but the 3 most recent
	for i := 0; i < len(files)-3; i++ {
		os.Remove(files[i].path)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(DEBUG, format, v...)
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.log(INFO, format, v...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(WARN, format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ERROR, format, v...)
}

// Fatal logs an error message and exits the program
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(ERROR, format, v...)
	os.Exit(1)
}

// Printf provides compatibility with standard logger
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Info(format, v...)
}

// Println provides compatibility with standard logger
func (l *Logger) Println(v ...interface{}) {
	l.Info("%s", fmt.Sprint(v...))
}

// Writer returns an io.Writer for the logger at INFO level
func (l *Logger) Writer() io.Writer {
	return &logWriter{logger: l}
}

// logWriter implements io.Writer for the logger
type logWriter struct {
	logger *Logger
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.logger.Info("%s", string(p))
	return len(p), nil
}
