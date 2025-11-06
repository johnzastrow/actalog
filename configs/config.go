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
	Driver   string // sqlite, postgres, mysql
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	SecretKey      string
	ExpirationTime time.Duration
	Issuer         string
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name        string
	Environment string // development, staging, production
	LogLevel    string // debug, info, warn, error
	CORSOrigins []string
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
			Driver:   getEnv("DB_DRIVER", "sqlite"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "actalog"),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_NAME", "actalog.db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			SecretKey:      getEnv("JWT_SECRET", ""), // Must be set in production
			ExpirationTime: getEnvDuration("JWT_EXPIRATION", 24*time.Hour),
			Issuer:         getEnv("JWT_ISSUER", "actalog"),
		},
		App: AppConfig{
			Name:        "ActaLog",
			Environment: getEnv("APP_ENV", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
			CORSOrigins: getEnvSlice("CORS_ORIGINS", []string{"http://localhost:8080", "http://localhost:3000"}),
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
