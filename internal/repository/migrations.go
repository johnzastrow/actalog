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
// NOTE: As of v0.4.0, all schema changes have been consolidated into the baseline schema in database.go
// Starting fresh databases no longer need incremental migrations.
var migrations = []Migration{
	{
		Version:     "0.4.0",
		Description: "Baseline schema with template-based workouts, WODs, and all features",
		Up: func(db *sql.DB, driver string) error {
			// This migration is handled by createTables in database.go
			// All features are now in the baseline schema:
			// - users (with password reset, email verification, birthday fields)
			// - workouts (template-based with name, created_by)
			// - wods (WOD definitions)
			// - workout_wods (junction table)
			// - user_workouts (junction table for workout instances)
			// - movements
			// - workout_movements (with is_pr field)
			// - refresh_tokens
			// - user_settings
			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			return fmt.Errorf("cannot rollback baseline migration")
		},
	},
	{
		Version:     "0.4.1",
		Description: "Add score_value, division, and is_pr columns to workout_wods table",
		Up: func(db *sql.DB, driver string) error {
			switch driver {
			case "sqlite3":
				// SQLite: Add columns with ALTER TABLE
				queries := []string{
					`ALTER TABLE workout_wods ADD COLUMN score_value TEXT`,
					`ALTER TABLE workout_wods ADD COLUMN division TEXT`,
					`ALTER TABLE workout_wods ADD COLUMN is_pr INTEGER NOT NULL DEFAULT 0`,
				}
				for _, query := range queries {
					if _, err := db.Exec(query); err != nil {
						return fmt.Errorf("failed to execute query: %w", err)
					}
				}
				return nil

			case "postgres":
				// PostgreSQL: Add columns with ALTER TABLE
				queries := []string{
					`ALTER TABLE workout_wods ADD COLUMN IF NOT EXISTS score_value TEXT`,
					`ALTER TABLE workout_wods ADD COLUMN IF NOT EXISTS division TEXT`,
					`ALTER TABLE workout_wods ADD COLUMN IF NOT EXISTS is_pr BOOLEAN NOT NULL DEFAULT false`,
				}
				for _, query := range queries {
					if _, err := db.Exec(query); err != nil {
						return fmt.Errorf("failed to execute query: %w", err)
					}
				}
				return nil

			case "mysql":
				// MySQL: Add columns with ALTER TABLE
				queries := []string{
					`ALTER TABLE workout_wods ADD COLUMN score_value TEXT`,
					`ALTER TABLE workout_wods ADD COLUMN division TEXT`,
					`ALTER TABLE workout_wods ADD COLUMN is_pr BOOLEAN NOT NULL DEFAULT 0`,
				}
				for _, query := range queries {
					if _, err := db.Exec(query); err != nil {
						return fmt.Errorf("failed to execute query: %w", err)
					}
				}
				return nil

			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}
		},
		Down: func(db *sql.DB, driver string) error {
			switch driver {
			case "sqlite3":
				// SQLite doesn't support DROP COLUMN easily, would require table recreation
				return fmt.Errorf("SQLite does not support dropping columns; manual intervention required")

			case "postgres":
				queries := []string{
					`ALTER TABLE workout_wods DROP COLUMN IF EXISTS is_pr`,
					`ALTER TABLE workout_wods DROP COLUMN IF EXISTS division`,
					`ALTER TABLE workout_wods DROP COLUMN IF EXISTS score_value`,
				}
				for _, query := range queries {
					if _, err := db.Exec(query); err != nil {
						return fmt.Errorf("failed to execute query: %w", err)
					}
				}
				return nil

			case "mysql":
				queries := []string{
					`ALTER TABLE workout_wods DROP COLUMN is_pr`,
					`ALTER TABLE workout_wods DROP COLUMN division`,
					`ALTER TABLE workout_wods DROP COLUMN score_value`,
				}
				for _, query := range queries {
					if _, err := db.Exec(query); err != nil {
						return fmt.Errorf("failed to execute query: %w", err)
					}
				}
				return nil

			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}
		},
	},
	// Future migrations for incremental schema changes will be added here
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
