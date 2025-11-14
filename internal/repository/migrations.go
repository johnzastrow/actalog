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
				// SQLite: Check if columns exist before adding
				// Query to check if column exists
				var count int
				err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('workout_wods') WHERE name='score_value'`).Scan(&count)
				if err != nil {
					return fmt.Errorf("failed to check for score_value column: %w", err)
				}

				// Only add columns if they don't exist
				if count == 0 {
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
				}
				return nil

			case "postgres":
				// PostgreSQL: Add columns with ALTER TABLE (IF NOT EXISTS supported)
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
				// MySQL: Check if columns exist before adding
				var count int
				err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.COLUMNS
					WHERE TABLE_SCHEMA = DATABASE()
					AND TABLE_NAME = 'workout_wods'
					AND COLUMN_NAME = 'score_value'`).Scan(&count)
				if err != nil {
					return fmt.Errorf("failed to check for score_value column: %w", err)
				}

				// Only add columns if they don't exist
				if count == 0 {
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
	{
		Version:     "0.4.2",
		Description: "Add user_workout_movements and user_workout_wods tables for performance tracking",
		Up: func(db *sql.DB, driver string) error {
			switch driver {
			case "sqlite3":
				queries := []string{
					`CREATE TABLE IF NOT EXISTS user_workout_movements (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						user_workout_id INTEGER NOT NULL,
						movement_id INTEGER NOT NULL,
						sets INTEGER,
						reps INTEGER,
						weight REAL,
						time INTEGER,
						distance REAL,
						notes TEXT,
						order_index INTEGER NOT NULL DEFAULT 0,
						created_at DATETIME NOT NULL,
						updated_at DATETIME NOT NULL,
						FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
						FOREIGN KEY (movement_id) REFERENCES movements(id) ON DELETE RESTRICT
					)`,
					`CREATE INDEX IF NOT EXISTS idx_user_workout_movements_user_workout_id ON user_workout_movements(user_workout_id)`,
					`CREATE INDEX IF NOT EXISTS idx_user_workout_movements_movement_id ON user_workout_movements(movement_id)`,
					`CREATE TABLE IF NOT EXISTS user_workout_wods (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						user_workout_id INTEGER NOT NULL,
						wod_id INTEGER NOT NULL,
						score_type TEXT,
						score_value TEXT,
						time_seconds INTEGER,
						rounds INTEGER,
						reps INTEGER,
						weight REAL,
						notes TEXT,
						order_index INTEGER NOT NULL DEFAULT 0,
						created_at DATETIME NOT NULL,
						updated_at DATETIME NOT NULL,
						FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
						FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT
					)`,
					`CREATE INDEX IF NOT EXISTS idx_user_workout_wods_user_workout_id ON user_workout_wods(user_workout_id)`,
					`CREATE INDEX IF NOT EXISTS idx_user_workout_wods_wod_id ON user_workout_wods(wod_id)`,
				}
				for _, query := range queries {
					if _, err := db.Exec(query); err != nil {
						return fmt.Errorf("failed to execute query: %w", err)
					}
				}
				return nil

			case "postgres":
				queries := []string{
					`CREATE TABLE IF NOT EXISTS user_workout_movements (
						id BIGSERIAL PRIMARY KEY,
						user_workout_id BIGINT NOT NULL,
						movement_id BIGINT NOT NULL,
						sets INTEGER,
						reps INTEGER,
						weight DECIMAL(10,2),
						time INTEGER,
						distance DECIMAL(10,2),
						notes TEXT,
						order_index INTEGER NOT NULL DEFAULT 0,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
						FOREIGN KEY (movement_id) REFERENCES movements(id) ON DELETE RESTRICT
					)`,
					`CREATE INDEX IF NOT EXISTS idx_user_workout_movements_user_workout_id ON user_workout_movements(user_workout_id)`,
					`CREATE INDEX IF NOT EXISTS idx_user_workout_movements_movement_id ON user_workout_movements(movement_id)`,
					`CREATE TABLE IF NOT EXISTS user_workout_wods (
						id BIGSERIAL PRIMARY KEY,
						user_workout_id BIGINT NOT NULL,
						wod_id BIGINT NOT NULL,
						score_type VARCHAR(50),
						score_value TEXT,
						time_seconds INTEGER,
						rounds INTEGER,
						reps INTEGER,
						weight DECIMAL(10,2),
						notes TEXT,
						order_index INTEGER NOT NULL DEFAULT 0,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
						FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT
					)`,
					`CREATE INDEX IF NOT EXISTS idx_user_workout_wods_user_workout_id ON user_workout_wods(user_workout_id)`,
					`CREATE INDEX IF NOT EXISTS idx_user_workout_wods_wod_id ON user_workout_wods(wod_id)`,
				}
				for _, query := range queries {
					if _, err := db.Exec(query); err != nil {
						return fmt.Errorf("failed to execute query: %w", err)
					}
				}
				return nil

			case "mysql":
				queries := []string{
					`CREATE TABLE IF NOT EXISTS user_workout_movements (
						id BIGINT AUTO_INCREMENT PRIMARY KEY,
						user_workout_id BIGINT NOT NULL,
						movement_id BIGINT NOT NULL,
						sets INT,
						reps INT,
						weight DECIMAL(10,2),
						time INT,
						distance DECIMAL(10,2),
						notes TEXT,
						order_index INT NOT NULL DEFAULT 0,
						created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
						FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
						FOREIGN KEY (movement_id) REFERENCES movements(id) ON DELETE RESTRICT,
						INDEX idx_user_workout_movements_user_workout_id (user_workout_id),
						INDEX idx_user_workout_movements_movement_id (movement_id)
					) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
					`CREATE TABLE IF NOT EXISTS user_workout_wods (
						id BIGINT AUTO_INCREMENT PRIMARY KEY,
						user_workout_id BIGINT NOT NULL,
						wod_id BIGINT NOT NULL,
						score_type VARCHAR(50),
						score_value TEXT,
						time_seconds INT,
						rounds INT,
						reps INT,
						weight DECIMAL(10,2),
						notes TEXT,
						order_index INT NOT NULL DEFAULT 0,
						created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
						FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
						FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT,
						INDEX idx_user_workout_wods_user_workout_id (user_workout_id),
						INDEX idx_user_workout_wods_wod_id (wod_id)
					) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
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
			queries := []string{
				`DROP TABLE IF EXISTS user_workout_wods`,
				`DROP TABLE IF EXISTS user_workout_movements`,
			}
			for _, query := range queries {
				if _, err := db.Exec(query); err != nil {
					return fmt.Errorf("failed to execute query: %w", err)
				}
			}
			return nil
		},
	},
	{
		Version:     "0.4.3",
		Description: "Add is_pr column to user_workout_movements and user_workout_wods tables",
		Up: func(db *sql.DB, driver string) error {
			switch driver {
			case "sqlite3":
				// SQLite: Check if column exists before adding
				var count int
				err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('user_workout_movements') WHERE name='is_pr'`).Scan(&count)
				if err != nil {
					return fmt.Errorf("failed to check for is_pr column: %w", err)
				}

				// Only add columns if they don't exist
				if count == 0 {
					queries := []string{
						`ALTER TABLE user_workout_movements ADD COLUMN is_pr INTEGER NOT NULL DEFAULT 0`,
						`CREATE INDEX IF NOT EXISTS idx_user_workout_movements_pr ON user_workout_movements(is_pr)`,
						`ALTER TABLE user_workout_wods ADD COLUMN is_pr INTEGER NOT NULL DEFAULT 0`,
						`CREATE INDEX IF NOT EXISTS idx_user_workout_wods_pr ON user_workout_wods(is_pr)`,
					}
					for _, query := range queries {
						if _, err := db.Exec(query); err != nil {
							return fmt.Errorf("failed to execute query: %w", err)
						}
					}
				}
				return nil

			case "postgres":
				// PostgreSQL: Add columns with ALTER TABLE (IF NOT EXISTS supported)
				queries := []string{
					`ALTER TABLE user_workout_movements ADD COLUMN IF NOT EXISTS is_pr BOOLEAN NOT NULL DEFAULT false`,
					`CREATE INDEX IF NOT EXISTS idx_user_workout_movements_pr ON user_workout_movements(is_pr)`,
					`ALTER TABLE user_workout_wods ADD COLUMN IF NOT EXISTS is_pr BOOLEAN NOT NULL DEFAULT false`,
					`CREATE INDEX IF NOT EXISTS idx_user_workout_wods_pr ON user_workout_wods(is_pr)`,
				}
				for _, query := range queries {
					if _, err := db.Exec(query); err != nil {
						return fmt.Errorf("failed to execute query: %w", err)
					}
				}
				return nil

			case "mysql":
				// MySQL: Check if columns exist before adding
				var count int
				err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.COLUMNS
					WHERE TABLE_SCHEMA = DATABASE()
					AND TABLE_NAME = 'user_workout_movements'
					AND COLUMN_NAME = 'is_pr'`).Scan(&count)
				if err != nil {
					return fmt.Errorf("failed to check for is_pr column: %w", err)
				}

				// Only add columns if they don't exist
				if count == 0 {
					queries := []string{
						`ALTER TABLE user_workout_movements ADD COLUMN is_pr BOOLEAN NOT NULL DEFAULT 0`,
						`CREATE INDEX idx_user_workout_movements_pr ON user_workout_movements(is_pr)`,
						`ALTER TABLE user_workout_wods ADD COLUMN is_pr BOOLEAN NOT NULL DEFAULT 0`,
						`CREATE INDEX idx_user_workout_wods_pr ON user_workout_wods(is_pr)`,
					}
					for _, query := range queries {
						if _, err := db.Exec(query); err != nil {
							return fmt.Errorf("failed to execute query: %w", err)
						}
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
					`DROP INDEX IF EXISTS idx_user_workout_wods_pr`,
					`ALTER TABLE user_workout_wods DROP COLUMN IF EXISTS is_pr`,
					`DROP INDEX IF EXISTS idx_user_workout_movements_pr`,
					`ALTER TABLE user_workout_movements DROP COLUMN IF EXISTS is_pr`,
				}
				for _, query := range queries {
					if _, err := db.Exec(query); err != nil {
						return fmt.Errorf("failed to execute query: %w", err)
					}
				}
				return nil

			case "mysql":
				queries := []string{
					`DROP INDEX idx_user_workout_wods_pr ON user_workout_wods`,
					`ALTER TABLE user_workout_wods DROP COLUMN is_pr`,
					`DROP INDEX idx_user_workout_movements_pr ON user_workout_movements`,
					`ALTER TABLE user_workout_movements DROP COLUMN is_pr`,
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
