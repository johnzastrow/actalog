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
	{
		Version:     "0.5.1",
		Description: "Add missing tables (refresh_tokens, user_settings, user_workout_movements, user_workout_wods) for existing databases",
		Up: func(db *sql.DB, driver string) error {
			// Check if refresh_tokens table exists
			hasRefreshTokens, err := checkTableExists(db, driver, "refresh_tokens")
			if err != nil {
				return fmt.Errorf("failed to check for refresh_tokens table: %w", err)
			}

			if !hasRefreshTokens {
				var createRefreshTokensSQL string
				switch driver {
				case "sqlite3":
					createRefreshTokensSQL = `
					CREATE TABLE IF NOT EXISTS refresh_tokens (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						user_id INTEGER NOT NULL,
						token TEXT UNIQUE NOT NULL,
						expires_at DATETIME NOT NULL,
						created_at DATETIME NOT NULL,
						revoked_at DATETIME,
						device_info TEXT,
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
					);
					CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
					CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);`
				case "postgres":
					createRefreshTokensSQL = `
					CREATE TABLE IF NOT EXISTS refresh_tokens (
						id BIGSERIAL PRIMARY KEY,
						user_id BIGINT NOT NULL,
						token VARCHAR(255) UNIQUE NOT NULL,
						expires_at TIMESTAMP NOT NULL,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						revoked_at TIMESTAMP,
						device_info TEXT,
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
					);
					CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
					CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);`
				case "mysql":
					createRefreshTokensSQL = `
					CREATE TABLE IF NOT EXISTS refresh_tokens (
						id BIGINT AUTO_INCREMENT PRIMARY KEY,
						user_id BIGINT NOT NULL,
						token VARCHAR(255) UNIQUE NOT NULL,
						expires_at DATETIME NOT NULL,
						created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
						revoked_at DATETIME,
						device_info TEXT,
						INDEX idx_refresh_tokens_user_id (user_id),
						INDEX idx_refresh_tokens_token (token),
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
					) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
				default:
					return fmt.Errorf("unsupported database driver: %s", driver)
				}

				if _, err := db.Exec(createRefreshTokensSQL); err != nil {
					return fmt.Errorf("failed to create refresh_tokens table: %w", err)
				}
			}

			// Check if user_settings table exists
			hasUserSettings, err := checkTableExists(db, driver, "user_settings")
			if err != nil {
				return fmt.Errorf("failed to check for user_settings table: %w", err)
			}

			if !hasUserSettings {
				var createUserSettingsSQL string
				switch driver {
				case "sqlite3":
					createUserSettingsSQL = `
					CREATE TABLE IF NOT EXISTS user_settings (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						user_id INTEGER UNIQUE NOT NULL,
						notification_preferences TEXT,
						data_export_format TEXT DEFAULT 'json',
						theme TEXT DEFAULT 'light',
						weight_unit TEXT DEFAULT 'lbs',
						distance_unit TEXT DEFAULT 'meters',
						created_at DATETIME NOT NULL,
						updated_at DATETIME NOT NULL,
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
					);
					CREATE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings(user_id);`
				case "postgres":
					createUserSettingsSQL = `
					CREATE TABLE IF NOT EXISTS user_settings (
						id BIGSERIAL PRIMARY KEY,
						user_id BIGINT UNIQUE NOT NULL,
						notification_preferences TEXT,
						data_export_format VARCHAR(50) DEFAULT 'json',
						theme VARCHAR(50) DEFAULT 'light',
						weight_unit VARCHAR(20) DEFAULT 'lbs',
						distance_unit VARCHAR(20) DEFAULT 'meters',
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
					);
					CREATE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings(user_id);`
				case "mysql":
					createUserSettingsSQL = `
					CREATE TABLE IF NOT EXISTS user_settings (
						id BIGINT AUTO_INCREMENT PRIMARY KEY,
						user_id BIGINT UNIQUE NOT NULL,
						notification_preferences TEXT,
						data_export_format VARCHAR(50) DEFAULT 'json',
						theme VARCHAR(50) DEFAULT 'light',
						weight_unit VARCHAR(20) DEFAULT 'lbs',
						distance_unit VARCHAR(20) DEFAULT 'meters',
						created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
						INDEX idx_user_settings_user_id (user_id),
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
					) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
				default:
					return fmt.Errorf("unsupported database driver: %s", driver)
				}

				if _, err := db.Exec(createUserSettingsSQL); err != nil {
					return fmt.Errorf("failed to create user_settings table: %w", err)
				}
			}

			// Check if user_workout_movements table exists
			hasUserWorkoutMovements, err := checkTableExists(db, driver, "user_workout_movements")
			if err != nil {
				return fmt.Errorf("failed to check for user_workout_movements table: %w", err)
			}

			if !hasUserWorkoutMovements {
				var createUserWorkoutMovementsSQL string
				switch driver {
				case "sqlite3":
					createUserWorkoutMovementsSQL = `
					CREATE TABLE IF NOT EXISTS user_workout_movements (
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
						is_pr INTEGER NOT NULL DEFAULT 0,
						FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
						FOREIGN KEY (movement_id) REFERENCES movements(id) ON DELETE RESTRICT
					);
					CREATE INDEX IF NOT EXISTS idx_user_workout_movements_user_workout_id ON user_workout_movements(user_workout_id);
					CREATE INDEX IF NOT EXISTS idx_user_workout_movements_movement_id ON user_workout_movements(movement_id);`
				case "postgres":
					createUserWorkoutMovementsSQL = `
					CREATE TABLE IF NOT EXISTS user_workout_movements (
						id BIGSERIAL PRIMARY KEY,
						user_workout_id BIGINT NOT NULL,
						movement_id BIGINT NOT NULL,
						sets INTEGER,
						reps INTEGER,
						weight DOUBLE PRECISION,
						time INTEGER,
						distance DOUBLE PRECISION,
						notes TEXT,
						order_index INTEGER NOT NULL DEFAULT 0,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						is_pr BOOLEAN NOT NULL DEFAULT FALSE,
						FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
						FOREIGN KEY (movement_id) REFERENCES movements(id) ON DELETE RESTRICT
					);
					CREATE INDEX IF NOT EXISTS idx_user_workout_movements_user_workout_id ON user_workout_movements(user_workout_id);
					CREATE INDEX IF NOT EXISTS idx_user_workout_movements_movement_id ON user_workout_movements(movement_id);`
				case "mysql":
					createUserWorkoutMovementsSQL = `
					CREATE TABLE IF NOT EXISTS user_workout_movements (
						id BIGINT AUTO_INCREMENT PRIMARY KEY,
						user_workout_id BIGINT NOT NULL,
						movement_id BIGINT NOT NULL,
						sets INTEGER,
						reps INTEGER,
						weight DOUBLE,
						time INTEGER,
						distance DOUBLE,
						notes TEXT,
						order_index INTEGER NOT NULL DEFAULT 0,
						created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
						is_pr BOOLEAN NOT NULL DEFAULT FALSE,
						INDEX idx_user_workout_movements_user_workout_id (user_workout_id),
						INDEX idx_user_workout_movements_movement_id (movement_id),
						FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
						FOREIGN KEY (movement_id) REFERENCES movements(id) ON DELETE RESTRICT
					) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
				default:
					return fmt.Errorf("unsupported database driver: %s", driver)
				}

				if _, err := db.Exec(createUserWorkoutMovementsSQL); err != nil {
					return fmt.Errorf("failed to create user_workout_movements table: %w", err)
				}
			}

			// Check if user_workout_wods table exists
			hasUserWorkoutWODs, err := checkTableExists(db, driver, "user_workout_wods")
			if err != nil {
				return fmt.Errorf("failed to check for user_workout_wods table: %w", err)
			}

			if !hasUserWorkoutWODs {
				var createUserWorkoutWODsSQL string
				switch driver {
				case "sqlite3":
					createUserWorkoutWODsSQL = `
					CREATE TABLE IF NOT EXISTS user_workout_wods (
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
						is_pr INTEGER NOT NULL DEFAULT 0,
						FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
						FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT
					);
					CREATE INDEX IF NOT EXISTS idx_user_workout_wods_user_workout_id ON user_workout_wods(user_workout_id);
					CREATE INDEX IF NOT EXISTS idx_user_workout_wods_wod_id ON user_workout_wods(wod_id);`
				case "postgres":
					createUserWorkoutWODsSQL = `
					CREATE TABLE IF NOT EXISTS user_workout_wods (
						id BIGSERIAL PRIMARY KEY,
						user_workout_id BIGINT NOT NULL,
						wod_id BIGINT NOT NULL,
						score_type VARCHAR(50),
						score_value TEXT,
						time_seconds INTEGER,
						rounds INTEGER,
						reps INTEGER,
						weight DOUBLE PRECISION,
						notes TEXT,
						order_index INTEGER NOT NULL DEFAULT 0,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						is_pr BOOLEAN NOT NULL DEFAULT FALSE,
						FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
						FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT
					);
					CREATE INDEX IF NOT EXISTS idx_user_workout_wods_user_workout_id ON user_workout_wods(user_workout_id);
					CREATE INDEX IF NOT EXISTS idx_user_workout_wods_wod_id ON user_workout_wods(wod_id);`
				case "mysql":
					createUserWorkoutWODsSQL = `
					CREATE TABLE IF NOT EXISTS user_workout_wods (
						id BIGINT AUTO_INCREMENT PRIMARY KEY,
						user_workout_id BIGINT NOT NULL,
						wod_id BIGINT NOT NULL,
						score_type VARCHAR(50),
						score_value TEXT,
						time_seconds INTEGER,
						rounds INTEGER,
						reps INTEGER,
						weight DOUBLE,
						notes TEXT,
						order_index INTEGER NOT NULL DEFAULT 0,
						created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
						is_pr BOOLEAN NOT NULL DEFAULT FALSE,
						INDEX idx_user_workout_wods_user_workout_id (user_workout_id),
						INDEX idx_user_workout_wods_wod_id (wod_id),
						FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
						FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT
					) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
				default:
					return fmt.Errorf("unsupported database driver: %s", driver)
				}

				if _, err := db.Exec(createUserWorkoutWODsSQL); err != nil {
					return fmt.Errorf("failed to create user_workout_wods table: %w", err)
				}
			}

			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			if _, err := db.Exec("DROP TABLE IF EXISTS user_workout_wods"); err != nil {
				return err
			}
			if _, err := db.Exec("DROP TABLE IF EXISTS user_workout_movements"); err != nil {
				return err
			}
			if _, err := db.Exec("DROP TABLE IF EXISTS user_settings"); err != nil {
				return err
			}
			if _, err := db.Exec("DROP TABLE IF EXISTS refresh_tokens"); err != nil {
				return err
			}
			return nil
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
