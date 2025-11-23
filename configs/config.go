// Package configs provides configuration management for ActaLog
package configs

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	App      AppConfig
	Logging  LoggingConfig
	Email    EmailConfig
	Security SecurityConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Driver   string // sqlite3, postgres, mysql
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	Schema   string // PostgreSQL schema (default: public)

	// Connection pooling (PostgreSQL and MySQL only)
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum connection lifetime
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	SecretKey            string
	ExpirationTime       time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name              string
	Environment       string   // development, staging, production
	LogLevel          string   // debug, info, warn, error
	CORSOrigins       []string
	AllowRegistration bool // Allow new user registration after first user
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string // debug, info, warn, error
	EnableFile bool   // Enable file logging
	FilePath   string // Path to log file
	MaxSizeMB  int    // Max file size in MB before rotation
	MaxBackups int    // Number of old log files to keep
}

// EmailConfig holds email/SMTP configuration
type EmailConfig struct {
	SMTPHost              string // SMTP server host
	SMTPPort              int    // SMTP server port (587 for STARTTLS, 465 for TLS, 25 for plain)
	SMTPUser              string // SMTP username
	SMTPPassword          string // SMTP password
	FromAddress           string // From email address
	FromName              string // From name
	Enabled               bool   // Enable email sending
	RequireVerification   bool   // Require email verification for new users
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	MaxLoginAttempts       int           // Max failed login attempts before lockout
	AccountLockoutDuration time.Duration // How long to lock account after max attempts

	// Password strength (for future use)
	MinPasswordLength       int  // Minimum password length
	RequirePasswordUppercase bool // Require at least one uppercase letter
	RequirePasswordLowercase bool // Require at least one lowercase letter
	RequirePasswordNumber    bool // Require at least one number
	RequirePasswordSpecial   bool // Require at least one special character
}

// Load loads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Driver:          getEnv("DB_DRIVER", "sqlite3"),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "actalog"),
			Password:        getEnv("DB_PASSWORD", ""),
			Database:        getEnv("DB_NAME", "actalog.db"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			Schema:          getEnv("DB_SCHEMA", "public"), // PostgreSQL schema
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		JWT: JWTConfig{
			SecretKey:            getEnv("JWT_SECRET", ""), // Must be set in production
			ExpirationTime:       getEnvDuration("JWT_EXPIRATION", 24*time.Hour),
			RefreshTokenDuration: getEnvDuration("JWT_REFRESH_DURATION", 30*24*time.Hour), // 30 days
			Issuer:               getEnv("JWT_ISSUER", "actalog"),
		},
		App: AppConfig{
			Name:              "ActaLog",
			Environment:       getEnv("APP_ENV", "development"),
			LogLevel:          getEnv("LOG_LEVEL", "info"),
			CORSOrigins:       getEnvSlice("CORS_ORIGINS", []string{"http://localhost:8080", "http://localhost:3000"}),
			AllowRegistration: getEnvBool("ALLOW_REGISTRATION", true), // Allow by default in development
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			EnableFile: getEnvBool("LOG_FILE_ENABLED", false),
			FilePath:   getEnv("LOG_FILE_PATH", ""),    // Empty = auto-detect (./logs/actalog.log)
			MaxSizeMB:  getEnvInt("LOG_MAX_SIZE_MB", 100), // 100MB default
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 3),   // Keep 3 old files
		},
		Email: EmailConfig{
			SMTPHost:            getEnv("SMTP_HOST", ""),
			SMTPPort:            getEnvInt("SMTP_PORT", 587), // Default to STARTTLS port
			SMTPUser:            getEnv("SMTP_USER", ""),
			SMTPPassword:        getEnv("SMTP_PASSWORD", ""),
			FromAddress:         getEnv("EMAIL_FROM", ""),
			FromName:            getEnv("EMAIL_FROM_NAME", "ActaLog"),
			Enabled:             getEnvBool("EMAIL_ENABLED", false), // Disabled by default
			RequireVerification: getEnvBool("REQUIRE_EMAIL_VERIFICATION", false), // Disabled by default for testing
		},
		Security: SecurityConfig{
			MaxLoginAttempts:        getEnvInt("MAX_LOGIN_ATTEMPTS", 5),
			AccountLockoutDuration:  getEnvDuration("ACCOUNT_LOCKOUT_DURATION", 15*time.Minute),

			// Password strength (future use)
			MinPasswordLength:        getEnvInt("MIN_PASSWORD_LENGTH", 8),
			RequirePasswordUppercase: getEnvBool("REQUIRE_PASSWORD_UPPERCASE", false),
			RequirePasswordLowercase: getEnvBool("REQUIRE_PASSWORD_LOWERCASE", false),
			RequirePasswordNumber:    getEnvBool("REQUIRE_PASSWORD_NUMBER", false),
			RequirePasswordSpecial:   getEnvBool("REQUIRE_PASSWORD_SPECIAL", false),
		},
	}

	// Validate critical configuration
	if cfg.App.Environment == "production" && cfg.JWT.SecretKey == "" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production environment")
	}

	return cfg, nil
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		return []string{value}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}
