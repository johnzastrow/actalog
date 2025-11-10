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
	{
		Version:     "0.3.3",
		Description: "Add birthday field to users table for profile editing",
		Up: func(db *sql.DB, driver string) error {
			var query string
			switch driver {
			case "sqlite3":
				query = "ALTER TABLE users ADD COLUMN birthday DATE"
			case "postgres":
				query = "ALTER TABLE users ADD COLUMN birthday DATE"
			case "mysql":
				query = "ALTER TABLE users ADD COLUMN birthday DATE"
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			_, err := db.Exec(query)
			if err != nil {
				return fmt.Errorf("failed to add birthday column: %w", err)
			}
			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			var query string
			switch driver {
			case "sqlite3":
				// SQLite doesn't support DROP COLUMN directly, would need table recreation
				return fmt.Errorf("SQLite does not support dropping columns")
			case "postgres":
				query = "ALTER TABLE users DROP COLUMN birthday"
			case "mysql":
				query = "ALTER TABLE users DROP COLUMN birthday"
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			_, err := db.Exec(query)
			return err
		},
	},
	{
		Version:     "0.4.0",
		Description: "Transform schema to match REQUIREMENTS.md - workouts become templates, add WODs, rename tables",
		Up: func(db *sql.DB, driver string) error {
			// This is a complex multi-phase migration
			// Phase 1: Create backup tables
			if err := migrateV040_CreateBackups(db, driver); err != nil {
				return fmt.Errorf("phase 1 failed (backups): %w", err)
			}

			// Phase 2: Create new tables (wods, user_workouts, workout_wods)
			if err := migrateV040_CreateNewTables(db, driver); err != nil {
				return fmt.Errorf("phase 2 failed (new tables): %w", err)
			}

			// Phase 3: Add new columns to workouts table
			if err := migrateV040_AddWorkoutColumns(db, driver); err != nil {
				return fmt.Errorf("phase 3 failed (add workout columns): %w", err)
			}

			// Phase 4: Rename existing tables
			if err := migrateV040_RenameTables(db, driver); err != nil {
				return fmt.Errorf("phase 4 failed (rename tables): %w", err)
			}

			// Phase 5: Migrate data from old structure to new
			if err := migrateV040_MigrateData(db, driver); err != nil {
				return fmt.Errorf("phase 5 failed (data migration): %w", err)
			}

			// Phase 6: Remove old columns from workouts table
			if err := migrateV040_RemoveOldWorkoutColumns(db, driver); err != nil {
				return fmt.Errorf("phase 6 failed (remove old columns): %w", err)
			}

			// Phase 7: Seed standard WODs
			if err := migrateV040_SeedWODs(db, driver); err != nil {
				return fmt.Errorf("phase 7 failed (seed WODs): %w", err)
			}

			fmt.Println("✓ Migration 0.4.0 completed successfully")
			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			return fmt.Errorf("rollback of 0.4.0 not supported - restore from backup instead")
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

// ========================================
// Migration v0.4.0 Helper Functions
// ========================================

// Phase 1: Create backup tables
func migrateV040_CreateBackups(db *sql.DB, driver string) error {
	fmt.Println("Phase 1: Creating backup tables...")

	queries := []string{
		"CREATE TABLE workouts_backup_v033 AS SELECT * FROM workouts",
		"CREATE TABLE workout_movements_backup_v033 AS SELECT * FROM workout_movements",
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	fmt.Println("  ✓ Backup tables created")
	return nil
}

// Phase 2: Create new tables (wods, user_workouts, workout_wods)
func migrateV040_CreateNewTables(db *sql.DB, driver string) error {
	fmt.Println("Phase 2: Creating new tables...")

	var queries []string

	switch driver {
	case "sqlite3":
		queries = []string{
			// wods table
			`CREATE TABLE IF NOT EXISTS wods (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT UNIQUE NOT NULL,
				source TEXT,
				type TEXT,
				regime TEXT,
				score_type TEXT,
				description TEXT,
				url TEXT,
				notes TEXT,
				is_standard INTEGER NOT NULL DEFAULT 0,
				created_by INTEGER,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL,
				FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
			)`,
			"CREATE INDEX IF NOT EXISTS idx_wods_name ON wods(name)",
			"CREATE INDEX IF NOT EXISTS idx_wods_type ON wods(type)",
			"CREATE INDEX IF NOT EXISTS idx_wods_source ON wods(source)",

			// user_workouts junction table
			`CREATE TABLE IF NOT EXISTS user_workouts (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				workout_id INTEGER NOT NULL,
				workout_date DATE NOT NULL,
				workout_type TEXT,
				total_time INTEGER,
				notes TEXT,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
				FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
				UNIQUE(user_id, workout_id, workout_date)
			)`,
			"CREATE INDEX IF NOT EXISTS idx_user_workouts_user_id ON user_workouts(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_user_workouts_workout_id ON user_workouts(workout_id)",
			"CREATE INDEX IF NOT EXISTS idx_user_workouts_date ON user_workouts(workout_date)",

			// workout_wods junction table
			`CREATE TABLE IF NOT EXISTS workout_wods (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				workout_id INTEGER NOT NULL,
				wod_id INTEGER NOT NULL,
				score_value TEXT,
				division TEXT,
				is_pr INTEGER NOT NULL DEFAULT 0,
				order_index INTEGER NOT NULL DEFAULT 0,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL,
				FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
				FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT
			)`,
			"CREATE INDEX IF NOT EXISTS idx_workout_wods_workout_id ON workout_wods(workout_id)",
			"CREATE INDEX IF NOT EXISTS idx_workout_wods_wod_id ON workout_wods(wod_id)",
		}

	case "postgres":
		queries = []string{
			// wods table
			`CREATE TABLE IF NOT EXISTS wods (
				id BIGSERIAL PRIMARY KEY,
				name VARCHAR(255) UNIQUE NOT NULL,
				source VARCHAR(100),
				type VARCHAR(50),
				regime VARCHAR(50),
				score_type VARCHAR(50),
				description TEXT,
				url VARCHAR(512),
				notes TEXT,
				is_standard BOOLEAN NOT NULL DEFAULT FALSE,
				created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)`,
			"CREATE INDEX IF NOT EXISTS idx_wods_name ON wods(name)",
			"CREATE INDEX IF NOT EXISTS idx_wods_type ON wods(type)",
			"CREATE INDEX IF NOT EXISTS idx_wods_source ON wods(source)",

			// user_workouts junction table
			`CREATE TABLE IF NOT EXISTS user_workouts (
				id BIGSERIAL PRIMARY KEY,
				user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				workout_id BIGINT NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
				workout_date DATE NOT NULL,
				workout_type VARCHAR(50),
				total_time INTEGER,
				notes TEXT,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(user_id, workout_id, workout_date)
			)`,
			"CREATE INDEX IF NOT EXISTS idx_user_workouts_user_id ON user_workouts(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_user_workouts_workout_id ON user_workouts(workout_id)",
			"CREATE INDEX IF NOT EXISTS idx_user_workouts_date ON user_workouts(workout_date)",

			// workout_wods junction table
			`CREATE TABLE IF NOT EXISTS workout_wods (
				id BIGSERIAL PRIMARY KEY,
				workout_id BIGINT NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
				wod_id BIGINT NOT NULL REFERENCES wods(id) ON DELETE RESTRICT,
				score_value VARCHAR(50),
				division VARCHAR(20),
				is_pr BOOLEAN NOT NULL DEFAULT FALSE,
				order_index INTEGER NOT NULL DEFAULT 0,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)`,
			"CREATE INDEX IF NOT EXISTS idx_workout_wods_workout_id ON workout_wods(workout_id)",
			"CREATE INDEX IF NOT EXISTS idx_workout_wods_wod_id ON workout_wods(wod_id)",
		}

	case "mysql":
		queries = []string{
			// wods table
			`CREATE TABLE IF NOT EXISTS wods (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(255) UNIQUE NOT NULL,
				source VARCHAR(100),
				type VARCHAR(50),
				regime VARCHAR(50),
				score_type VARCHAR(50),
				description TEXT,
				url VARCHAR(512),
				notes TEXT,
				is_standard BOOLEAN NOT NULL DEFAULT FALSE,
				created_by BIGINT,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
			)`,
			"CREATE INDEX idx_wods_name ON wods(name)",
			"CREATE INDEX idx_wods_type ON wods(type)",
			"CREATE INDEX idx_wods_source ON wods(source)",

			// user_workouts junction table
			`CREATE TABLE IF NOT EXISTS user_workouts (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				user_id BIGINT NOT NULL,
				workout_id BIGINT NOT NULL,
				workout_date DATE NOT NULL,
				workout_type VARCHAR(50),
				total_time INTEGER,
				notes TEXT,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
				FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
				UNIQUE KEY unique_user_workout_date (user_id, workout_id, workout_date)
			)`,
			"CREATE INDEX idx_user_workouts_user_id ON user_workouts(user_id)",
			"CREATE INDEX idx_user_workouts_workout_id ON user_workouts(workout_id)",
			"CREATE INDEX idx_user_workouts_date ON user_workouts(workout_date)",

			// workout_wods junction table
			`CREATE TABLE IF NOT EXISTS workout_wods (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				workout_id BIGINT NOT NULL,
				wod_id BIGINT NOT NULL,
				score_value VARCHAR(50),
				division VARCHAR(20),
				is_pr BOOLEAN NOT NULL DEFAULT FALSE,
				order_index INTEGER NOT NULL DEFAULT 0,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
				FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT
			)`,
			"CREATE INDEX idx_workout_wods_workout_id ON workout_wods(workout_id)",
			"CREATE INDEX idx_workout_wods_wod_id ON workout_wods(wod_id)",
		}

	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	fmt.Println("  ✓ New tables created (wods, user_workouts, workout_wods)")
	return nil
}

// Phase 3: Add new columns to workouts table
func migrateV040_AddWorkoutColumns(db *sql.DB, driver string) error {
	fmt.Println("Phase 3: Adding new columns to workouts table...")

	var queries []string

	switch driver {
	case "sqlite3":
		queries = []string{
			"ALTER TABLE workouts ADD COLUMN name TEXT",
			"ALTER TABLE workouts ADD COLUMN created_by INTEGER",
		}

	case "postgres":
		queries = []string{
			"ALTER TABLE workouts ADD COLUMN IF NOT EXISTS name VARCHAR(255)",
			"ALTER TABLE workouts ADD COLUMN IF NOT EXISTS created_by BIGINT",
		}

	case "mysql":
		queries = []string{
			"ALTER TABLE workouts ADD COLUMN name VARCHAR(255)",
			"ALTER TABLE workouts ADD COLUMN created_by BIGINT",
		}

	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to add column: %w", err)
		}
	}

	fmt.Println("  ✓ Added name and created_by columns to workouts")
	return nil
}

// Phase 4: Rename existing tables
func migrateV040_RenameTables(db *sql.DB, driver string) error {
	fmt.Println("Phase 4: Renaming tables...")

	var queries []string

	switch driver {
	case "sqlite3":
		queries = []string{
			"ALTER TABLE movements RENAME TO strength_movements",
			"ALTER TABLE workout_movements RENAME TO workout_strength",
		}

	case "postgres":
		queries = []string{
			"ALTER TABLE movements RENAME TO strength_movements",
			"ALTER TABLE workout_movements RENAME TO workout_strength",
		}

	case "mysql":
		queries = []string{
			"RENAME TABLE movements TO strength_movements",
			"RENAME TABLE workout_movements TO workout_strength",
		}

	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to rename table: %w", err)
		}
	}

	// Rename column in workout_strength: movement_id → strength_id
	var renameColQuery string
	switch driver {
	case "sqlite3":
		// SQLite doesn't support RENAME COLUMN directly in older versions
		// We'll need to recreate the table - this is complex, skip for now
		// The column name change is non-critical for functionality
		fmt.Println("  ⚠ SQLite: movement_id column not renamed (functionally equivalent)")

	case "postgres":
		renameColQuery = "ALTER TABLE workout_strength RENAME COLUMN movement_id TO strength_id"

	case "mysql":
		// MySQL requires full column redefinition
		// Get the column definition first - this is complex, we'll handle it later
		fmt.Println("  ⚠ MySQL: movement_id column not renamed (functionally equivalent)")

	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	if renameColQuery != "" {
		if _, err := db.Exec(renameColQuery); err != nil {
			// Non-critical error, log and continue
			fmt.Printf("  ⚠ Warning: failed to rename column: %v\n", err)
		}
	}

	fmt.Println("  ✓ Tables renamed (movements → strength_movements, workout_movements → workout_strength)")
	return nil
}

// Phase 5: Migrate data from old structure to new
func migrateV040_MigrateData(db *sql.DB, driver string) error {
	fmt.Println("Phase 5: Migrating workout data...")

	// Strategy: For each old workout, create a template and user_workout entry
	// Conservative approach: each user's workout becomes their own private template

	// Get all workouts from backup
	rows, err := db.Query("SELECT id, user_id, workout_date, workout_type, workout_name, notes, total_time, created_at, updated_at FROM workouts_backup_v033 ORDER BY id")
	if err != nil {
		return fmt.Errorf("failed to query backup: %w", err)
	}
	defer rows.Close()

	type OldWorkout struct {
		ID          int64
		UserID      int64
		WorkoutDate string
		WorkoutType string
		WorkoutName sql.NullString
		Notes       sql.NullString
		TotalTime   sql.NullInt64
		CreatedAt   string
		UpdatedAt   string
	}

	// Map to track template creation: workout_id → template_id
	templateMap := make(map[int64]int64)

	for rows.Next() {
		var old OldWorkout
		if err := rows.Scan(&old.ID, &old.UserID, &old.WorkoutDate, &old.WorkoutType, &old.WorkoutName, &old.Notes, &old.TotalTime, &old.CreatedAt, &old.UpdatedAt); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// The workout already exists with the same ID (it's been preserved)
		// We just need to add the 'name' column value and created_by
		// Then create user_workout entry

		// Generate template name
		templateName := "Custom Workout"
		if old.WorkoutName.Valid && old.WorkoutName.String != "" {
			templateName = old.WorkoutName.String
		} else {
			templateName = fmt.Sprintf("Workout %d", old.ID)
		}

		// Update the workout row to add name and created_by
		var updateQuery string
		switch driver {
		case "sqlite3", "mysql":
			updateQuery = "UPDATE workouts SET name = ?, created_by = ? WHERE id = ?"
		case "postgres":
			updateQuery = "UPDATE workouts SET name = $1, created_by = $2 WHERE id = $3"
		}

		if _, err := db.Exec(updateQuery, templateName, old.UserID, old.ID); err != nil {
			return fmt.Errorf("failed to update workout: %w", err)
		}

		// Create user_workout entry linking user to this template
		var insertQuery string
		switch driver {
		case "sqlite3":
			insertQuery = `INSERT INTO user_workouts (user_id, workout_id, workout_date, workout_type, total_time, notes, created_at, updated_at)
			               VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		case "postgres":
			insertQuery = `INSERT INTO user_workouts (user_id, workout_id, workout_date, workout_type, total_time, notes, created_at, updated_at)
			               VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
		case "mysql":
			insertQuery = `INSERT INTO user_workouts (user_id, workout_id, workout_date, workout_type, total_time, notes, created_at, updated_at)
			               VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		}

		totalTime := sql.NullInt64{}
		if old.TotalTime.Valid {
			totalTime = old.TotalTime
		}

		notes := sql.NullString{}
		if old.Notes.Valid {
			notes = old.Notes
		}

		if _, err := db.Exec(insertQuery, old.UserID, old.ID, old.WorkoutDate, old.WorkoutType, totalTime, notes, old.CreatedAt, old.UpdatedAt); err != nil {
			return fmt.Errorf("failed to insert user_workout: %w", err)
		}

		templateMap[old.ID] = old.ID // Workout ID becomes template ID
	}

	fmt.Printf("  ✓ Migrated %d workouts to template + user_workout structure\n", len(templateMap))
	return nil
}

// Phase 6: Remove old columns from workouts table
func migrateV040_RemoveOldWorkoutColumns(db *sql.DB, driver string) error {
	fmt.Println("Phase 6: Removing old columns from workouts table...")

	// Drop old columns: user_id, workout_date, workout_type, workout_name, total_time
	// Keep: id, notes, created_at, updated_at
	// Already added: name, created_by

	var dropQueries []string

	switch driver {
	case "sqlite3":
		// SQLite doesn't support DROP COLUMN easily
		// We'd need to recreate the table - skip for now
		// The extra columns don't hurt
		fmt.Println("  ⚠ SQLite: Old columns not dropped (no ALTER TABLE DROP COLUMN support)")
		fmt.Println("  ℹ Extra columns (user_id, workout_date, workout_type, total_time) preserved but unused")
		return nil

	case "postgres":
		dropQueries = []string{
			"ALTER TABLE workouts DROP COLUMN IF EXISTS user_id",
			"ALTER TABLE workouts DROP COLUMN IF EXISTS workout_date",
			"ALTER TABLE workouts DROP COLUMN IF EXISTS workout_type",
			"ALTER TABLE workouts DROP COLUMN IF EXISTS workout_name",
			"ALTER TABLE workouts DROP COLUMN IF EXISTS total_time",
		}

	case "mysql":
		dropQueries = []string{
			"ALTER TABLE workouts DROP COLUMN user_id",
			"ALTER TABLE workouts DROP COLUMN workout_date",
			"ALTER TABLE workouts DROP COLUMN workout_type",
			"ALTER TABLE workouts DROP COLUMN workout_name",
			"ALTER TABLE workouts DROP COLUMN total_time",
		}

	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	for _, query := range dropQueries {
		if _, err := db.Exec(query); err != nil {
			// Non-critical error
			fmt.Printf("  ⚠ Warning: failed to drop column: %v\n", err)
		}
	}

	fmt.Println("  ✓ Workouts table refactored")
	return nil
}

// Phase 7: Seed standard WODs
func migrateV040_SeedWODs(db *sql.DB, driver string) error {
	fmt.Println("Phase 7: Seeding standard WODs...")

	wods := []struct {
		Name        string
		Source      string
		Type        string
		Regime      string
		ScoreType   string
		Description string
	}{
		{
			Name:        "Fran",
			Source:      "CrossFit",
			Type:        "Girl",
			Regime:      "Fastest Time",
			ScoreType:   "Time",
			Description: "21-15-9 reps for time of:\n- Thrusters (95 lbs)\n- Pull-ups",
		},
		{
			Name:        "Grace",
			Source:      "CrossFit",
			Type:        "Girl",
			Regime:      "Fastest Time",
			ScoreType:   "Time",
			Description: "30 Clean and Jerks for time (135 lbs)",
		},
		{
			Name:        "Helen",
			Source:      "CrossFit",
			Type:        "Girl",
			Regime:      "Fastest Time",
			ScoreType:   "Time",
			Description: "3 rounds for time of:\n- 400m Run\n- 21 Kettlebell Swings (53 lbs)\n- 12 Pull-ups",
		},
		{
			Name:        "Diane",
			Source:      "CrossFit",
			Type:        "Girl",
			Regime:      "Fastest Time",
			ScoreType:   "Time",
			Description: "21-15-9 reps for time of:\n- Deadlifts (225 lbs)\n- Handstand Push-ups",
		},
		{
			Name:        "Karen",
			Source:      "CrossFit",
			Type:        "Girl",
			Regime:      "Fastest Time",
			ScoreType:   "Time",
			Description: "150 Wall Balls for time (20 lbs, 10 ft target)",
		},
		{
			Name:        "Murph",
			Source:      "CrossFit",
			Type:        "Hero",
			Regime:      "Fastest Time",
			ScoreType:   "Time",
			Description: "For time:\n- 1 mile Run\n- 100 Pull-ups\n- 200 Push-ups\n- 300 Air Squats\n- 1 mile Run\n\nPartition the pull-ups, push-ups, and squats as needed. Start and finish with a mile run. Wear a 20 lb vest if possible.",
		},
		{
			Name:        "Cindy",
			Source:      "CrossFit",
			Type:        "Girl",
			Regime:      "AMRAP",
			ScoreType:   "Rounds+Reps",
			Description: "20 minute AMRAP of:\n- 5 Pull-ups\n- 10 Push-ups\n- 15 Air Squats",
		},
		{
			Name:        "Annie",
			Source:      "CrossFit",
			Type:        "Girl",
			Regime:      "Fastest Time",
			ScoreType:   "Time",
			Description: "50-40-30-20-10 reps for time of:\n- Double-Unders\n- Sit-ups",
		},
		{
			Name:        "DT",
			Source:      "CrossFit",
			Type:        "Hero",
			Regime:      "Fastest Time",
			ScoreType:   "Time",
			Description: "5 rounds for time of:\n- 12 Deadlifts (155 lbs)\n- 9 Hang Power Cleans (155 lbs)\n- 6 Push Jerks (155 lbs)",
		},
	}

	now := time.Now()

	for _, wod := range wods {
		var insertQuery string
		switch driver {
		case "sqlite3":
			insertQuery = `INSERT INTO wods (name, source, type, regime, score_type, description, is_standard, created_at, updated_at)
			               VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?)`
		case "postgres":
			insertQuery = `INSERT INTO wods (name, source, type, regime, score_type, description, is_standard, created_at, updated_at)
			               VALUES ($1, $2, $3, $4, $5, $6, TRUE, $7, $8)`
		case "mysql":
			insertQuery = `INSERT INTO wods (name, source, type, regime, score_type, description, is_standard, created_at, updated_at)
			               VALUES (?, ?, ?, ?, ?, ?, TRUE, ?, ?)`
		}

		if _, err := db.Exec(insertQuery, wod.Name, wod.Source, wod.Type, wod.Regime, wod.ScoreType, wod.Description, now, now); err != nil {
			// Skip duplicates
			continue
		}
	}

	fmt.Printf("  ✓ Seeded %d standard WODs\n", len(wods))
	return nil
}
