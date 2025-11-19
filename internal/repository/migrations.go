package repository

import (
	"database/sql"
	"fmt"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     string
	Description string
	Up          func(*sql.DB, string) error // Takes db and driver
	Down        func(*sql.DB, string) error // Takes db and driver
}

// migrations holds all database migrations in order
// NOTE: As of v0.5.0 (2025-11-19), the schema has been rebaselined with all features consolidated.
// The baseline schema includes all tables and features in database.go.
// No migrations from versions prior to 0.5.0 are supported.
var migrations = []Migration{
	{
		Version:     "0.5.0",
		Description: "Baseline schema (2025-11-19) - Complete ActaLog schema with all features",
		Up: func(db *sql.DB, driver string) error {
			// This migration is handled by createTables in database.go
			// The baseline schema includes:
			// - users (auth, email verification, account lockout, birthday)
			// - refresh_tokens (session management)
			// - password_resets (password recovery)
			// - email_verification_tokens (email verification)
			// - user_settings (preferences, units, theme)
			// - audit_logs (security audit trail with enhanced fields)
			// - movements (exercise definitions)
			// - workouts (workout templates)
			// - wods (WOD definitions with metadata)
			// - workout_movements (template movements with PR tracking)
			// - workout_wods (template WODs with scoring)
			// - user_workouts (workout instances - supports templates OR ad-hoc via workout_name)
			// - user_workout_movements (performance tracking for movements)
			// - user_workout_wods (performance tracking for WODs with PR flags)
			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			return fmt.Errorf("cannot rollback baseline migration")
		},
	},
	// Future incremental migrations will be added here
}

// RunMigrations runs all pending migrations
func RunMigrations(db *sql.DB, driver string) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db, driver); err != nil {
		return err
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db, driver)
	if err != nil {
		return err
	}

	// Run pending migrations
	for _, migration := range migrations {
		if isApplied(migration.Version, appliedMigrations) {
			continue
		}

		fmt.Printf("Applying migration %s: %s\n", migration.Version, migration.Description)

		// Run the migration
		if err := migration.Up(db, driver); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}

		// Record the migration
		if err := recordMigration(db, driver, migration.Version, migration.Description); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		fmt.Printf("✓ Migration %s applied successfully\n", migration.Version)
	}

	return nil
}

// createMigrationsTable creates the schema_migrations table with database-specific syntax
func createMigrationsTable(db *sql.DB, driver string) error {
	var query string

	switch driver {
	case "sqlite3":
		query = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version TEXT UNIQUE NOT NULL,
			description TEXT NOT NULL,
			applied_at DATETIME NOT NULL
		)`

	case "postgres":
		query = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id BIGSERIAL PRIMARY KEY,
			version VARCHAR(50) UNIQUE NOT NULL,
			description TEXT NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`

	case "mysql":
		query = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			version VARCHAR(50) UNIQUE NOT NULL,
			description TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// getAppliedMigrations returns a list of applied migration versions
func getAppliedMigrations(db *sql.DB, driver string) (map[string]bool, error) {
	query := `SELECT version FROM schema_migrations ORDER BY applied_at`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// isApplied checks if a migration version has been applied
func isApplied(version string, appliedMigrations map[string]bool) bool {
	return appliedMigrations[version]
}

// recordMigration records a migration as applied with database-specific syntax
func recordMigration(db *sql.DB, driver, version, description string) error {
	var query string

	switch driver {
	case "sqlite3", "mysql":
		query = `INSERT INTO schema_migrations (version, description, applied_at) VALUES (?, ?, ?)`
		_, err := db.Exec(query, version, description, time.Now())
		return err

	case "postgres":
		query = `INSERT INTO schema_migrations (version, description, applied_at) VALUES ($1, $2, $3)`
		_, err := db.Exec(query, version, description, time.Now())
		return err

	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}
}

// RollbackMigration rolls back the last applied migration
func RollbackMigration(db *sql.DB, driver string) error {
	// Get the last applied migration
	var version, description string
	query := `SELECT version, description FROM schema_migrations ORDER BY applied_at DESC LIMIT 1`
	err := db.QueryRow(query).Scan(&version, &description)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no migrations to rollback")
		}
		return err
	}

	// Find the migration
	var targetMigration *Migration
	for i := range migrations {
		if migrations[i].Version == version {
			targetMigration = &migrations[i]
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %s not found", version)
	}

	fmt.Printf("Rolling back migration %s: %s\n", version, description)

	// Run the down migration
	if err := targetMigration.Down(db, driver); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", version, err)
	}

	// Remove the migration record with database-specific parameter syntax
	var deleteQuery string
	switch driver {
	case "sqlite3", "mysql":
		deleteQuery = "DELETE FROM schema_migrations WHERE version = ?"
	case "postgres":
		deleteQuery = "DELETE FROM schema_migrations WHERE version = $1"
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err = db.Exec(deleteQuery, version)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	fmt.Printf("✓ Migration %s rolled back successfully\n", version)
	return nil
}
