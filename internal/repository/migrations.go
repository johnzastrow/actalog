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
	Up          func(*sql.DB) error
	Down        func(*sql.DB) error
}

// migrations holds all database migrations in order
var migrations = []Migration{
	{
		Version:     "0.1.0",
		Description: "Initial schema with users, workouts, movements, workout_movements",
		Up: func(db *sql.DB) error {
			// This migration is handled by createInitialTablesIfNotExist
			return nil
		},
		Down: func(db *sql.DB) error {
			return fmt.Errorf("cannot rollback initial migration")
		},
	},
	// Future migrations will be added here
	// Example:
	// {
	//     Version:     "0.2.0",
	//     Description: "Add email_verified column to users",
	//     Up: func(db *sql.DB) error {
	//         _, err := db.Exec("ALTER TABLE users ADD COLUMN email_verified INTEGER NOT NULL DEFAULT 0")
	//         return err
	//     },
	//     Down: func(db *sql.DB) error {
	//         _, err := db.Exec("ALTER TABLE users DROP COLUMN email_verified")
	//         return err
	//     },
	// },
}

// RunMigrations runs all pending migrations
func RunMigrations(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return err
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
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
		if err := migration.Up(db); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}

		// Record the migration
		if err := recordMigration(db, migration.Version, migration.Description); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		fmt.Printf("✓ Migration %s applied successfully\n", migration.Version)
	}

	return nil
}

// createMigrationsTable creates the schema_migrations table
func createMigrationsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version TEXT UNIQUE NOT NULL,
		description TEXT NOT NULL,
		applied_at DATETIME NOT NULL
	);
	`
	_, err := db.Exec(query)
	return err
}

// getAppliedMigrations returns a list of applied migration versions
func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
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

// recordMigration records a migration as applied
func recordMigration(db *sql.DB, version, description string) error {
	query := `INSERT INTO schema_migrations (version, description, applied_at) VALUES (?, ?, ?)`
	_, err := db.Exec(query, version, description, time.Now())
	return err
}

// RollbackMigration rolls back the last applied migration
func RollbackMigration(db *sql.DB) error {
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
	if err := targetMigration.Down(db); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", version, err)
	}

	// Remove the migration record
	_, err = db.Exec("DELETE FROM schema_migrations WHERE version = ?", version)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	fmt.Printf("✓ Migration %s rolled back successfully\n", version)
	return nil
}
