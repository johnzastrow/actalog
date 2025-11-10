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
var migrations = []Migration{
	{
		Version:     "0.1.0",
		Description: "Initial schema with users, workouts, movements, workout_movements",
		Up: func(db *sql.DB, driver string) error {
			// This migration is handled by createInitialTablesIfNotExist
			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			return fmt.Errorf("cannot rollback initial migration")
		},
	},
	// Future migrations will be added here
	{
		Version:     "0.2.0",
		Description: "Add password reset fields to users table",
		Up: func(db *sql.DB, driver string) error {
			var queries []string
			switch driver {
			case "sqlite3":
				queries = []string{
					"ALTER TABLE users ADD COLUMN reset_token TEXT",
					"ALTER TABLE users ADD COLUMN reset_token_expires_at DATETIME",
				}
			case "postgres":
				queries = []string{
					"ALTER TABLE users ADD COLUMN IF NOT EXISTS reset_token VARCHAR(255)",
					"ALTER TABLE users ADD COLUMN IF NOT EXISTS reset_token_expires_at TIMESTAMP",
				}
			case "mysql":
				queries = []string{
					"ALTER TABLE users ADD COLUMN reset_token VARCHAR(255)",
					"ALTER TABLE users ADD COLUMN reset_token_expires_at DATETIME",
				}
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			for _, query := range queries {
				if _, err := db.Exec(query); err != nil {
					return fmt.Errorf("failed to execute query '%s': %w", query, err)
				}
			}
			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			var queries []string
			switch driver {
			case "sqlite3":
				// SQLite doesn't support DROP COLUMN directly in older versions
				return fmt.Errorf("rollback not supported for SQLite")
			case "postgres":
				queries = []string{
					"ALTER TABLE users DROP COLUMN IF EXISTS reset_token",
					"ALTER TABLE users DROP COLUMN IF EXISTS reset_token_expires_at",
				}
			case "mysql":
				queries = []string{
					"ALTER TABLE users DROP COLUMN reset_token",
					"ALTER TABLE users DROP COLUMN reset_token_expires_at",
				}
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			for _, query := range queries {
				if _, err := db.Exec(query); err != nil {
					return fmt.Errorf("failed to execute query '%s': %w", query, err)
				}
			}
			return nil
		},
	},
	{
		Version:     "0.3.0",
		Description: "Add PR (Personal Record) tracking to workout_movements",
		Up: func(db *sql.DB, driver string) error {
			var queries []string
			switch driver {
			case "sqlite3":
				queries = []string{
					"ALTER TABLE workout_movements ADD COLUMN is_pr INTEGER NOT NULL DEFAULT 0",
				}
			case "postgres":
				queries = []string{
					"ALTER TABLE workout_movements ADD COLUMN IF NOT EXISTS is_pr BOOLEAN NOT NULL DEFAULT FALSE",
				}
			case "mysql":
				queries = []string{
					"ALTER TABLE workout_movements ADD COLUMN is_pr BOOLEAN NOT NULL DEFAULT FALSE",
				}
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			for _, query := range queries {
				if _, err := db.Exec(query); err != nil {
					return fmt.Errorf("failed to execute query '%s': %w", query, err)
				}
			}
			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			var queries []string
			switch driver {
			case "sqlite3":
				// SQLite doesn't support DROP COLUMN directly in older versions
				return fmt.Errorf("rollback not supported for SQLite")
			case "postgres":
				queries = []string{
					"ALTER TABLE workout_movements DROP COLUMN IF EXISTS is_pr",
				}
			case "mysql":
				queries = []string{
					"ALTER TABLE workout_movements DROP COLUMN is_pr",
				}
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			for _, query := range queries {
				if _, err := db.Exec(query); err != nil {
					return fmt.Errorf("failed to execute query '%s': %w", query, err)
				}
			}
			return nil
		},
	},
	{
		Version:     "0.3.1",
		Description: "Add email verification fields to users table",
		Up: func(db *sql.DB, driver string) error {
			var queries []string
			switch driver {
			case "sqlite3":
				queries = []string{
					"ALTER TABLE users ADD COLUMN email_verified INTEGER NOT NULL DEFAULT 0",
					"ALTER TABLE users ADD COLUMN email_verified_at DATETIME",
					"ALTER TABLE users ADD COLUMN verification_token TEXT",
					"ALTER TABLE users ADD COLUMN verification_token_expires_at DATETIME",
				}
			case "postgres":
				queries = []string{
					"ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE",
					"ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMP",
					"ALTER TABLE users ADD COLUMN IF NOT EXISTS verification_token VARCHAR(255)",
					"ALTER TABLE users ADD COLUMN IF NOT EXISTS verification_token_expires_at TIMESTAMP",
				}
			case "mysql":
				queries = []string{
					"ALTER TABLE users ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE",
					"ALTER TABLE users ADD COLUMN email_verified_at DATETIME",
					"ALTER TABLE users ADD COLUMN verification_token VARCHAR(255)",
					"ALTER TABLE users ADD COLUMN verification_token_expires_at DATETIME",
				}
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			for _, query := range queries {
				if _, err := db.Exec(query); err != nil {
					return fmt.Errorf("failed to execute query '%s': %w", query, err)
				}
			}
			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			var queries []string
			switch driver {
			case "sqlite3":
				// SQLite doesn't support DROP COLUMN directly in older versions
				return fmt.Errorf("rollback not supported for SQLite")
			case "postgres":
				queries = []string{
					"ALTER TABLE users DROP COLUMN IF EXISTS email_verified",
					"ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at",
					"ALTER TABLE users DROP COLUMN IF EXISTS verification_token",
					"ALTER TABLE users DROP COLUMN IF EXISTS verification_token_expires_at",
				}
			case "mysql":
				queries = []string{
					"ALTER TABLE users DROP COLUMN email_verified",
					"ALTER TABLE users DROP COLUMN email_verified_at",
					"ALTER TABLE users DROP COLUMN verification_token",
					"ALTER TABLE users DROP COLUMN verification_token_expires_at",
				}
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			for _, query := range queries {
				if _, err := db.Exec(query); err != nil {
					return fmt.Errorf("failed to execute query '%s': %w", query, err)
				}
			}
			return nil
		},
	},
	{
		Version:     "0.3.2",
		Description: "Add refresh_tokens table for Remember Me functionality",
		Up: func(db *sql.DB, driver string) error {
			var queries []string
			switch driver {
			case "sqlite3":
				queries = []string{
					`CREATE TABLE IF NOT EXISTS refresh_tokens (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						user_id INTEGER NOT NULL,
						token TEXT UNIQUE NOT NULL,
						expires_at DATETIME NOT NULL,
						created_at DATETIME NOT NULL,
						revoked_at DATETIME,
						device_info TEXT,
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
					)`,
					"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)",
					"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token)",
					"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at)",
				}
			case "postgres":
				queries = []string{
					`CREATE TABLE IF NOT EXISTS refresh_tokens (
						id SERIAL PRIMARY KEY,
						user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
						token VARCHAR(255) UNIQUE NOT NULL,
						expires_at TIMESTAMP NOT NULL,
						created_at TIMESTAMP NOT NULL,
						revoked_at TIMESTAMP,
						device_info TEXT
					)`,
					"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)",
					"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token)",
					"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at)",
				}
			case "mysql":
				queries = []string{
					`CREATE TABLE IF NOT EXISTS refresh_tokens (
						id INT AUTO_INCREMENT PRIMARY KEY,
						user_id INT NOT NULL,
						token VARCHAR(255) UNIQUE NOT NULL,
						expires_at DATETIME NOT NULL,
						created_at DATETIME NOT NULL,
						revoked_at DATETIME,
						device_info TEXT,
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
					)`,
					"CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id)",
					"CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token)",
					"CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at)",
				}
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			for _, query := range queries {
				if _, err := db.Exec(query); err != nil {
					return fmt.Errorf("failed to execute query '%s': %w", query, err)
				}
			}
			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			_, err := db.Exec("DROP TABLE IF EXISTS refresh_tokens")
			return err
		},
	},
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
