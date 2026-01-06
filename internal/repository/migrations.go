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
	{
		Version:     "0.5.2",
		Description: "Add data_change_logs table for audit trail of edits and deletes",
		Up: func(db *sql.DB, driver string) error {
			// Check if data_change_logs table exists
			hasDataChangeLogs, err := checkTableExists(db, driver, "data_change_logs")
			if err != nil {
				return fmt.Errorf("failed to check for data_change_logs table: %w", err)
			}

			if !hasDataChangeLogs {
				var createDataChangeLogsSQL string
				switch driver {
				case "sqlite3":
					createDataChangeLogsSQL = `
					CREATE TABLE IF NOT EXISTS data_change_logs (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						entity_type TEXT NOT NULL,
						entity_id INTEGER NOT NULL,
						entity_name TEXT NOT NULL,
						operation TEXT NOT NULL,
						user_id INTEGER NOT NULL,
						user_email TEXT NOT NULL,
						before_values TEXT,
						after_values TEXT,
						ip_address TEXT,
						user_agent TEXT,
						created_at DATETIME NOT NULL
					);
					CREATE INDEX IF NOT EXISTS idx_data_change_logs_entity ON data_change_logs(entity_type, entity_id);
					CREATE INDEX IF NOT EXISTS idx_data_change_logs_user_id ON data_change_logs(user_id);
					CREATE INDEX IF NOT EXISTS idx_data_change_logs_created_at ON data_change_logs(created_at);
					CREATE INDEX IF NOT EXISTS idx_data_change_logs_operation ON data_change_logs(operation);`
				case "postgres":
					createDataChangeLogsSQL = `
					CREATE TABLE IF NOT EXISTS data_change_logs (
						id BIGSERIAL PRIMARY KEY,
						entity_type VARCHAR(50) NOT NULL,
						entity_id BIGINT NOT NULL,
						entity_name VARCHAR(255) NOT NULL,
						operation VARCHAR(20) NOT NULL,
						user_id BIGINT NOT NULL,
						user_email VARCHAR(255) NOT NULL,
						before_values TEXT,
						after_values TEXT,
						ip_address VARCHAR(45),
						user_agent TEXT,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
					);
					CREATE INDEX IF NOT EXISTS idx_data_change_logs_entity ON data_change_logs(entity_type, entity_id);
					CREATE INDEX IF NOT EXISTS idx_data_change_logs_user_id ON data_change_logs(user_id);
					CREATE INDEX IF NOT EXISTS idx_data_change_logs_created_at ON data_change_logs(created_at);
					CREATE INDEX IF NOT EXISTS idx_data_change_logs_operation ON data_change_logs(operation);`
				case "mysql":
					createDataChangeLogsSQL = `
					CREATE TABLE IF NOT EXISTS data_change_logs (
						id BIGINT AUTO_INCREMENT PRIMARY KEY,
						entity_type VARCHAR(50) NOT NULL,
						entity_id BIGINT NOT NULL,
						entity_name VARCHAR(255) NOT NULL,
						operation VARCHAR(20) NOT NULL,
						user_id BIGINT NOT NULL,
						user_email VARCHAR(255) NOT NULL,
						before_values TEXT,
						after_values TEXT,
						ip_address VARCHAR(45),
						user_agent TEXT,
						created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
						INDEX idx_data_change_logs_entity (entity_type, entity_id),
						INDEX idx_data_change_logs_user_id (user_id),
						INDEX idx_data_change_logs_created_at (created_at),
						INDEX idx_data_change_logs_operation (operation)
					) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
				default:
					return fmt.Errorf("unsupported database driver: %s", driver)
				}

				if _, err := db.Exec(createDataChangeLogsSQL); err != nil {
					return fmt.Errorf("failed to create data_change_logs table: %w", err)
				}
			}

			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			if _, err := db.Exec("DROP TABLE IF EXISTS data_change_logs"); err != nil {
				return err
			}
			return nil
		},
	},
	{
		Version:     "0.6.0",
		Description: "Add organizations table and organization_id to users",
		Up: func(db *sql.DB, driver string) error {
			var createOrgSQL string
			var alterUsersSQL string

			switch driver {
			case "sqlite3":
				createOrgSQL = `
				CREATE TABLE IF NOT EXISTS organizations (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT UNIQUE NOT NULL,
					description TEXT,
					created_at DATETIME NOT NULL,
					updated_at DATETIME NOT NULL
				);
				CREATE INDEX IF NOT EXISTS idx_organizations_name ON organizations(name);`

				alterUsersSQL = `ALTER TABLE users ADD COLUMN organization_id INTEGER REFERENCES organizations(id) ON DELETE SET NULL;`

			case "postgres":
				createOrgSQL = `
				CREATE TABLE IF NOT EXISTS organizations (
					id BIGSERIAL PRIMARY KEY,
					name VARCHAR(255) UNIQUE NOT NULL,
					description TEXT,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
				);
				CREATE INDEX IF NOT EXISTS idx_organizations_name ON organizations(name);`

				alterUsersSQL = `ALTER TABLE users ADD COLUMN organization_id BIGINT REFERENCES organizations(id) ON DELETE SET NULL;`

			case "mysql":
				createOrgSQL = `
				CREATE TABLE IF NOT EXISTS organizations (
					id BIGINT AUTO_INCREMENT PRIMARY KEY,
					name VARCHAR(255) UNIQUE NOT NULL,
					description TEXT,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					INDEX idx_organizations_name (name)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

				alterUsersSQL = `ALTER TABLE users ADD COLUMN organization_id BIGINT,
								ADD FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL;`

			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			// Create organizations table
			if _, err := db.Exec(createOrgSQL); err != nil {
				return fmt.Errorf("failed to create organizations table: %w", err)
			}

			// Check if column already exists before adding
			hasColumn, err := checkColumnExists(db, driver, "users", "organization_id")
			if err != nil {
				return fmt.Errorf("failed to check for organization_id column: %w", err)
			}

			if !hasColumn {
				if _, err := db.Exec(alterUsersSQL); err != nil {
					return fmt.Errorf("failed to add organization_id to users: %w", err)
				}
			}

			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			// Drop organization_id column from users
			var dropColumnSQL string
			switch driver {
			case "sqlite3":
				return fmt.Errorf("SQLite does not support dropping columns easily")
			case "postgres", "mysql":
				dropColumnSQL = "ALTER TABLE users DROP COLUMN organization_id"
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			if _, err := db.Exec(dropColumnSQL); err != nil {
				return err
			}

			// Drop organizations table
			if _, err := db.Exec("DROP TABLE IF EXISTS organizations"); err != nil {
				return err
			}

			return nil
		},
	},
	{
		Version:     "0.13.0",
		Description: "Migrate user-organization relationship from one-to-many to many-to-many",
		Up: func(db *sql.DB, driver string) error {
			// Check if user_organizations table already exists
			hasJunctionTable, err := checkTableExists(db, driver, "user_organizations")
			if err != nil {
				return fmt.Errorf("failed to check for user_organizations table: %w", err)
			}

			if hasJunctionTable {
				// Table already exists, check if we need to drop old column
				hasOldColumn, err := checkColumnExists(db, driver, "users", "organization_id")
				if err != nil {
					return fmt.Errorf("failed to check for organization_id column: %w", err)
				}

				if !hasOldColumn {
					// Migration already complete
					return nil
				}
				// Fall through to drop column
			} else {
				// Step 1: Create user_organizations junction table
				var createJunctionSQL string
				switch driver {
				case "sqlite3":
					createJunctionSQL = `
					CREATE TABLE IF NOT EXISTS user_organizations (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						user_id INTEGER NOT NULL,
						organization_id INTEGER NOT NULL,
						role TEXT DEFAULT 'member',
						joined_at DATETIME NOT NULL,
						created_at DATETIME NOT NULL,
						updated_at DATETIME NOT NULL,
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
						FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT,
						UNIQUE(user_id, organization_id)
					);
					CREATE INDEX IF NOT EXISTS idx_user_orgs_user_id ON user_organizations(user_id);
					CREATE INDEX IF NOT EXISTS idx_user_orgs_org_id ON user_organizations(organization_id);
					CREATE INDEX IF NOT EXISTS idx_user_orgs_lookup ON user_organizations(user_id, organization_id);`

				case "postgres":
					createJunctionSQL = `
					CREATE TABLE IF NOT EXISTS user_organizations (
						id BIGSERIAL PRIMARY KEY,
						user_id BIGINT NOT NULL,
						organization_id BIGINT NOT NULL,
						role VARCHAR(50) DEFAULT 'member',
						joined_at TIMESTAMP NOT NULL,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
						FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT,
						UNIQUE(user_id, organization_id)
					);
					CREATE INDEX IF NOT EXISTS idx_user_orgs_user_id ON user_organizations(user_id);
					CREATE INDEX IF NOT EXISTS idx_user_orgs_org_id ON user_organizations(organization_id);
					CREATE INDEX IF NOT EXISTS idx_user_orgs_lookup ON user_organizations(user_id, organization_id);`

				case "mysql":
					createJunctionSQL = `
					CREATE TABLE IF NOT EXISTS user_organizations (
						id BIGINT AUTO_INCREMENT PRIMARY KEY,
						user_id BIGINT NOT NULL,
						organization_id BIGINT NOT NULL,
						role VARCHAR(50) DEFAULT 'member',
						joined_at DATETIME NOT NULL,
						created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
						INDEX idx_user_orgs_user_id (user_id),
						INDEX idx_user_orgs_org_id (organization_id),
						INDEX idx_user_orgs_lookup (user_id, organization_id),
						UNIQUE KEY unique_user_org (user_id, organization_id),
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
						FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT
					) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

				default:
					return fmt.Errorf("unsupported database driver: %s", driver)
				}

				if _, err := db.Exec(createJunctionSQL); err != nil {
					return fmt.Errorf("failed to create user_organizations table: %w", err)
				}

				// Step 2: Migrate existing data from users.organization_id to junction table
				var migrateQuery string
				now := time.Now()

				switch driver {
				case "sqlite3", "mysql":
					migrateQuery = `
					INSERT INTO user_organizations (user_id, organization_id, role, joined_at, created_at, updated_at)
					SELECT id, organization_id, 'member', created_at, ?, ?
					FROM users
					WHERE organization_id IS NOT NULL`
					_, err = db.Exec(migrateQuery, now, now)

				case "postgres":
					migrateQuery = `
					INSERT INTO user_organizations (user_id, organization_id, role, joined_at, created_at, updated_at)
					SELECT id, organization_id, 'member', created_at, $1, $2
					FROM users
					WHERE organization_id IS NOT NULL`
					_, err = db.Exec(migrateQuery, now, now)

				default:
					return fmt.Errorf("unsupported database driver: %s", driver)
				}

				if err != nil {
					return fmt.Errorf("failed to migrate organization data: %w", err)
				}

				// Step 3: Verify migration
				var countUsers, countJunction int64
				var countQuery string

				switch driver {
				case "sqlite3", "mysql":
					countQuery = "SELECT COUNT(*) FROM users WHERE organization_id IS NOT NULL"
				case "postgres":
					countQuery = "SELECT COUNT(*) FROM users WHERE organization_id IS NOT NULL"
				}

				if err := db.QueryRow(countQuery).Scan(&countUsers); err != nil {
					return fmt.Errorf("failed to count users with organizations: %w", err)
				}

				if err := db.QueryRow("SELECT COUNT(*) FROM user_organizations").Scan(&countJunction); err != nil {
					return fmt.Errorf("failed to count junction records: %w", err)
				}

				if countUsers != countJunction {
					return fmt.Errorf("migration verification failed: expected %d records, got %d", countUsers, countJunction)
				}

				fmt.Printf("✓ Migrated %d organization assignments to junction table\n", countJunction)
			}

			// Step 4: Drop organization_id column from users table
			hasOldColumn, err := checkColumnExists(db, driver, "users", "organization_id")
			if err != nil {
				return fmt.Errorf("failed to check for organization_id column: %w", err)
			}

			if hasOldColumn {
				switch driver {
				case "sqlite3":
					// SQLite doesn't support DROP COLUMN, need to rebuild table
					// Get all columns except organization_id
					rows, err := db.Query("PRAGMA table_info(users)")
					if err != nil {
						return fmt.Errorf("failed to get users table info: %w", err)
					}

					var columns []string
					for rows.Next() {
						var cid int
						var name, colType string
						var notnull, pk int
						var dfltValue sql.NullString

						if err := rows.Scan(&cid, &name, &colType, &notnull, &dfltValue, &pk); err != nil {
							rows.Close()
							return fmt.Errorf("failed to scan column info: %w", err)
						}

						if name != "organization_id" {
							columns = append(columns, name)
						}
					}
					rows.Close()

					// Build column list for new table
					columnList := ""
					for i, col := range columns {
						if i > 0 {
							columnList += ", "
						}
						columnList += col
					}

					// Rebuild table without organization_id
					rebuildSQL := fmt.Sprintf(`
						BEGIN TRANSACTION;

						CREATE TABLE users_new (
							id INTEGER PRIMARY KEY AUTOINCREMENT,
							email TEXT UNIQUE NOT NULL,
							password_hash TEXT NOT NULL,
							name TEXT NOT NULL,
							profile_image TEXT,
							birthday DATE,
							role TEXT NOT NULL DEFAULT 'user',
							email_verified INTEGER NOT NULL DEFAULT 0,
							email_verified_at DATETIME,
							verification_token TEXT,
							verification_token_expires_at DATETIME,
							reset_token TEXT,
							reset_token_expires_at DATETIME,
							failed_login_attempts INTEGER NOT NULL DEFAULT 0,
							locked_at DATETIME,
							locked_until DATETIME,
							account_disabled INTEGER NOT NULL DEFAULT 0,
							disabled_at DATETIME,
							disabled_by_user_id INTEGER,
							disable_reason TEXT,
							created_at DATETIME NOT NULL,
							updated_at DATETIME NOT NULL,
							last_login_at DATETIME,
							FOREIGN KEY (disabled_by_user_id) REFERENCES users(id) ON DELETE SET NULL
						);

						INSERT INTO users_new (%s)
						SELECT %s FROM users;

						DROP TABLE users;

						ALTER TABLE users_new RENAME TO users;

						CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
						CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

						COMMIT;
					`, columnList, columnList)

					if _, err := db.Exec(rebuildSQL); err != nil {
						return fmt.Errorf("failed to rebuild users table: %w", err)
					}

				case "postgres":
					dropColumnSQL := "ALTER TABLE users DROP COLUMN organization_id"
					if _, err := db.Exec(dropColumnSQL); err != nil {
						return fmt.Errorf("failed to drop organization_id column: %w", err)
					}

				case "mysql":
					// MySQL requires dropping the foreign key constraint first
					// Find the foreign key constraint name
					var constraintName sql.NullString
					err := db.QueryRow(`
						SELECT CONSTRAINT_NAME
						FROM information_schema.KEY_COLUMN_USAGE
						WHERE TABLE_SCHEMA = DATABASE()
						AND TABLE_NAME = 'users'
						AND COLUMN_NAME = 'organization_id'
						AND REFERENCED_TABLE_NAME IS NOT NULL
					`).Scan(&constraintName)

					if err != nil && err != sql.ErrNoRows {
						return fmt.Errorf("failed to find foreign key constraint: %w", err)
					}

					// Drop foreign key if it exists
					if constraintName.Valid {
						dropFKSQL := fmt.Sprintf("ALTER TABLE users DROP FOREIGN KEY %s", constraintName.String)
						if _, err := db.Exec(dropFKSQL); err != nil {
							return fmt.Errorf("failed to drop foreign key constraint: %w", err)
						}
					}

					// Now drop the column
					dropColumnSQL := "ALTER TABLE users DROP COLUMN organization_id"
					if _, err := db.Exec(dropColumnSQL); err != nil {
						return fmt.Errorf("failed to drop organization_id column: %w", err)
					}

				default:
					return fmt.Errorf("unsupported database driver: %s", driver)
				}

				fmt.Println("✓ Dropped organization_id column from users table")
			}

			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			// Step 1: Re-add organization_id column to users
			hasColumn, err := checkColumnExists(db, driver, "users", "organization_id")
			if err != nil {
				return fmt.Errorf("failed to check for organization_id column: %w", err)
			}

			if !hasColumn {
				var addColumnSQL string
				switch driver {
				case "sqlite3":
					addColumnSQL = `ALTER TABLE users ADD COLUMN organization_id INTEGER REFERENCES organizations(id) ON DELETE SET NULL`
				case "postgres":
					addColumnSQL = `ALTER TABLE users ADD COLUMN organization_id BIGINT REFERENCES organizations(id) ON DELETE SET NULL`
				case "mysql":
					addColumnSQL = `ALTER TABLE users ADD COLUMN organization_id BIGINT,
									ADD FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL`
				default:
					return fmt.Errorf("unsupported database driver: %s", driver)
				}

				if _, err := db.Exec(addColumnSQL); err != nil {
					return fmt.Errorf("failed to add organization_id column: %w", err)
				}
			}

			// Step 2: Copy data back (takes first organization only)
			// WARNING: This loses data for users in multiple organizations
			var copyDataSQL string
			switch driver {
			case "sqlite3", "mysql":
				copyDataSQL = `
					UPDATE users SET organization_id = (
						SELECT organization_id FROM user_organizations
						WHERE user_organizations.user_id = users.id
						LIMIT 1
					)`
			case "postgres":
				copyDataSQL = `
					UPDATE users SET organization_id = (
						SELECT organization_id FROM user_organizations
						WHERE user_organizations.user_id = users.id
						LIMIT 1
					)`
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			if _, err := db.Exec(copyDataSQL); err != nil {
				return fmt.Errorf("failed to copy organization data back: %w", err)
			}

			// Step 3: Drop junction table
			if _, err := db.Exec("DROP TABLE IF EXISTS user_organizations"); err != nil {
				return fmt.Errorf("failed to drop user_organizations table: %w", err)
			}

			fmt.Println("⚠️  WARNING: Rollback complete but data loss occurred for users in multiple organizations")
			return nil
		},
	},
	{
		Version:     "0.14.0",
		Description: "Add subscription billing system with user and organization subscriptions",
		Up: func(db *sql.DB, driver string) error {
			// Check if tables already exist
			hasUserSubs, err := checkTableExists(db, driver, "user_subscriptions")
			if err != nil {
				return fmt.Errorf("failed to check for user_subscriptions table: %w", err)
			}
			if hasUserSubs {
				fmt.Println("✓ user_subscriptions table already exists, skipping creation")
			}

			hasOrgSubs, err := checkTableExists(db, driver, "organization_subscriptions")
			if err != nil {
				return fmt.Errorf("failed to check for organization_subscriptions table: %w", err)
			}
			if hasOrgSubs {
				fmt.Println("✓ organization_subscriptions table already exists, skipping creation")
			}

			// Step 1: Create user_subscriptions table
			if !hasUserSubs {
				var createUserSubsSQL string
				switch driver {
				case "sqlite3":
					createUserSubsSQL = `
					CREATE TABLE user_subscriptions (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						user_id INTEGER NOT NULL,
						subscription_type TEXT NOT NULL CHECK(subscription_type IN ('free', 'monthly', 'annual')),
						status TEXT NOT NULL CHECK(status IN ('active', 'expired', 'cancelled')),
						is_permanent_free INTEGER NOT NULL DEFAULT 0,
						start_date DATETIME NOT NULL,
						end_date DATETIME,
						last_payment_date DATETIME,
						next_billing_date DATETIME,
						cancelled_at DATETIME,
						cancelled_reason TEXT,
						notes TEXT,
						created_at DATETIME NOT NULL,
						updated_at DATETIME NOT NULL,
						created_by_user_id INTEGER,
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
						FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE SET NULL
					);
					CREATE INDEX idx_user_subscriptions_user_id ON user_subscriptions(user_id);
					CREATE INDEX idx_user_subscriptions_status ON user_subscriptions(status);
					CREATE INDEX idx_user_subscriptions_next_billing ON user_subscriptions(next_billing_date);
					`
				case "postgres":
					createUserSubsSQL = `
					CREATE TABLE user_subscriptions (
						id BIGSERIAL PRIMARY KEY,
						user_id BIGINT NOT NULL,
						subscription_type VARCHAR(20) NOT NULL CHECK(subscription_type IN ('free', 'monthly', 'annual')),
						status VARCHAR(20) NOT NULL CHECK(status IN ('active', 'expired', 'cancelled')),
						is_permanent_free BOOLEAN NOT NULL DEFAULT FALSE,
						start_date TIMESTAMP NOT NULL,
						end_date TIMESTAMP,
						last_payment_date TIMESTAMP,
						next_billing_date TIMESTAMP,
						cancelled_at TIMESTAMP,
						cancelled_reason TEXT,
						notes TEXT,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						created_by_user_id BIGINT,
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
						FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE SET NULL
					);
					CREATE INDEX idx_user_subscriptions_user_id ON user_subscriptions(user_id);
					CREATE INDEX idx_user_subscriptions_status ON user_subscriptions(status);
					CREATE INDEX idx_user_subscriptions_next_billing ON user_subscriptions(next_billing_date);
					`
				case "mysql":
					createUserSubsSQL = `
					CREATE TABLE user_subscriptions (
						id BIGINT AUTO_INCREMENT PRIMARY KEY,
						user_id BIGINT NOT NULL,
						subscription_type VARCHAR(20) NOT NULL,
						status VARCHAR(20) NOT NULL,
						is_permanent_free BOOLEAN NOT NULL DEFAULT FALSE,
						start_date DATETIME NOT NULL,
						end_date DATETIME,
						last_payment_date DATETIME,
						next_billing_date DATETIME,
						cancelled_at DATETIME,
						cancelled_reason TEXT,
						notes TEXT,
						created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
						created_by_user_id BIGINT,
						INDEX idx_user_subscriptions_user_id (user_id),
						INDEX idx_user_subscriptions_status (status),
						INDEX idx_user_subscriptions_next_billing (next_billing_date),
						CONSTRAINT chk_user_sub_type CHECK (subscription_type IN ('free', 'monthly', 'annual')),
						CONSTRAINT chk_user_status CHECK (status IN ('active', 'expired', 'cancelled')),
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
						FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE SET NULL
					) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
					`
				default:
					return fmt.Errorf("unsupported database driver: %s", driver)
				}

				if _, err := db.Exec(createUserSubsSQL); err != nil {
					return fmt.Errorf("failed to create user_subscriptions table: %w", err)
				}
				fmt.Println("✓ Created user_subscriptions table")
			}

			// Step 2: Create organization_subscriptions table
			if !hasOrgSubs {
				var createOrgSubsSQL string
				switch driver {
				case "sqlite3":
					createOrgSubsSQL = `
					CREATE TABLE organization_subscriptions (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						organization_id INTEGER NOT NULL,
						subscription_type TEXT NOT NULL CHECK(subscription_type IN ('free', 'monthly', 'annual')),
						status TEXT NOT NULL CHECK(status IN ('active', 'expired', 'cancelled')),
						is_permanent_free INTEGER NOT NULL DEFAULT 0,
						start_date DATETIME NOT NULL,
						end_date DATETIME,
						last_payment_date DATETIME,
						next_billing_date DATETIME,
						cancelled_at DATETIME,
						cancelled_reason TEXT,
						notes TEXT,
						created_at DATETIME NOT NULL,
						updated_at DATETIME NOT NULL,
						created_by_user_id INTEGER,
						FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
						FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE SET NULL
					);
					CREATE INDEX idx_org_subscriptions_org_id ON organization_subscriptions(organization_id);
					CREATE INDEX idx_org_subscriptions_status ON organization_subscriptions(status);
					CREATE INDEX idx_org_subscriptions_next_billing ON organization_subscriptions(next_billing_date);
					`
				case "postgres":
					createOrgSubsSQL = `
					CREATE TABLE organization_subscriptions (
						id BIGSERIAL PRIMARY KEY,
						organization_id BIGINT NOT NULL,
						subscription_type VARCHAR(20) NOT NULL CHECK(subscription_type IN ('free', 'monthly', 'annual')),
						status VARCHAR(20) NOT NULL CHECK(status IN ('active', 'expired', 'cancelled')),
						is_permanent_free BOOLEAN NOT NULL DEFAULT FALSE,
						start_date TIMESTAMP NOT NULL,
						end_date TIMESTAMP,
						last_payment_date TIMESTAMP,
						next_billing_date TIMESTAMP,
						cancelled_at TIMESTAMP,
						cancelled_reason TEXT,
						notes TEXT,
						created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
						created_by_user_id BIGINT,
						FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
						FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE SET NULL
					);
					CREATE INDEX idx_org_subscriptions_org_id ON organization_subscriptions(organization_id);
					CREATE INDEX idx_org_subscriptions_status ON organization_subscriptions(status);
					CREATE INDEX idx_org_subscriptions_next_billing ON organization_subscriptions(next_billing_date);
					`
				case "mysql":
					createOrgSubsSQL = `
					CREATE TABLE organization_subscriptions (
						id BIGINT AUTO_INCREMENT PRIMARY KEY,
						organization_id BIGINT NOT NULL,
						subscription_type VARCHAR(20) NOT NULL,
						status VARCHAR(20) NOT NULL,
						is_permanent_free BOOLEAN NOT NULL DEFAULT FALSE,
						start_date DATETIME NOT NULL,
						end_date DATETIME,
						last_payment_date DATETIME,
						next_billing_date DATETIME,
						cancelled_at DATETIME,
						cancelled_reason TEXT,
						notes TEXT,
						created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
						updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
						created_by_user_id BIGINT,
						INDEX idx_org_subscriptions_org_id (organization_id),
						INDEX idx_org_subscriptions_status (status),
						INDEX idx_org_subscriptions_next_billing (next_billing_date),
						CONSTRAINT chk_org_sub_type CHECK (subscription_type IN ('free', 'monthly', 'annual')),
						CONSTRAINT chk_org_status CHECK (status IN ('active', 'expired', 'cancelled')),
						FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
						FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE SET NULL
					) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
					`
				default:
					return fmt.Errorf("unsupported database driver: %s", driver)
				}

				if _, err := db.Exec(createOrgSubsSQL); err != nil {
					return fmt.Errorf("failed to create organization_subscriptions table: %w", err)
				}
				fmt.Println("✓ Created organization_subscriptions table")
			}

			// Step 3: Seed existing users with permanent free subscriptions (backward compatibility)
			// This ensures all existing users maintain full access
			var seedQuery string
			now := time.Now()

			switch driver {
			case "sqlite3":
				seedQuery = rebindQuery(`
					INSERT INTO user_subscriptions (user_id, subscription_type, status, is_permanent_free, start_date, created_at, updated_at)
					SELECT id, 'free', 'active', 1, ?, ?, ?
					FROM users
					WHERE NOT EXISTS (
						SELECT 1 FROM user_subscriptions WHERE user_subscriptions.user_id = users.id
					)
				`)
			case "postgres":
				seedQuery = `
					INSERT INTO user_subscriptions (user_id, subscription_type, status, is_permanent_free, start_date, created_at, updated_at)
					SELECT id, 'free', 'active', TRUE, $1, $2, $3
					FROM users
					WHERE NOT EXISTS (
						SELECT 1 FROM user_subscriptions WHERE user_subscriptions.user_id = users.id
					)
				`
			case "mysql":
				seedQuery = rebindQuery(`
					INSERT INTO user_subscriptions (user_id, subscription_type, status, is_permanent_free, start_date, created_at, updated_at)
					SELECT id, 'free', 'active', TRUE, ?, ?, ?
					FROM users
					WHERE NOT EXISTS (
						SELECT 1 FROM user_subscriptions WHERE user_subscriptions.user_id = users.id
					)
				`)
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			result, err := db.Exec(seedQuery, now, now, now)
			if err != nil {
				return fmt.Errorf("failed to seed user subscriptions: %w", err)
			}

			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return fmt.Errorf("failed to get rows affected: %w", err)
			}

			fmt.Printf("✓ Seeded %d existing users with permanent free subscriptions\n", rowsAffected)

			// Step 4: Verify seeding
			var userCount, subCount int64
			if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
				return fmt.Errorf("failed to count users: %w", err)
			}
			if err := db.QueryRow("SELECT COUNT(*) FROM user_subscriptions").Scan(&subCount); err != nil {
				return fmt.Errorf("failed to count subscriptions: %w", err)
			}

			fmt.Printf("✓ Verification: %d users, %d subscriptions\n", userCount, subCount)

			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			// Drop subscription tables (WARNING: Data loss)
			if _, err := db.Exec("DROP TABLE IF EXISTS organization_subscriptions"); err != nil {
				return fmt.Errorf("failed to drop organization_subscriptions: %w", err)
			}
			if _, err := db.Exec("DROP TABLE IF EXISTS user_subscriptions"); err != nil {
				return fmt.Errorf("failed to drop user_subscriptions: %w", err)
			}

			fmt.Println("⚠️  WARNING: Subscription data has been deleted")
			return nil
		},
	},
	{
		Version:     "0.15.0",
		Description: "Add notifications table for gym-wide notification system",
		Up: func(db *sql.DB, driver string) error {
			// Check if notifications table already exists
			hasNotifications, err := checkTableExists(db, driver, "notifications")
			if err != nil {
				return fmt.Errorf("failed to check for notifications table: %w", err)
			}
			if hasNotifications {
				fmt.Println("✓ notifications table already exists, skipping creation")
				return nil
			}

			// Create notifications table
			var createNotificationsSQL string
			switch driver {
			case "sqlite3":
				createNotificationsSQL = `
				CREATE TABLE notifications (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					user_id INTEGER NOT NULL,
					organization_id INTEGER,
					type TEXT NOT NULL,
					title TEXT NOT NULL,
					message TEXT NOT NULL,
					data TEXT,
					read_at DATETIME,
					created_at DATETIME NOT NULL,
					updated_at DATETIME NOT NULL,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
					FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
				);
				CREATE INDEX idx_notifications_user_id ON notifications(user_id);
				CREATE INDEX idx_notifications_user_read ON notifications(user_id, read_at);
				CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
				CREATE INDEX idx_notifications_type ON notifications(type);
				`
			case "postgres":
				createNotificationsSQL = `
				CREATE TABLE notifications (
					id BIGSERIAL PRIMARY KEY,
					user_id BIGINT NOT NULL,
					organization_id BIGINT,
					type VARCHAR(50) NOT NULL,
					title VARCHAR(255) NOT NULL,
					message TEXT NOT NULL,
					data JSONB,
					read_at TIMESTAMP,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
					FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
				);
				CREATE INDEX idx_notifications_user_id ON notifications(user_id);
				CREATE INDEX idx_notifications_user_read ON notifications(user_id, read_at);
				CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
				CREATE INDEX idx_notifications_type ON notifications(type);
				`
			case "mysql":
				createNotificationsSQL = `
				CREATE TABLE notifications (
					id BIGINT AUTO_INCREMENT PRIMARY KEY,
					user_id BIGINT NOT NULL,
					organization_id BIGINT,
					type VARCHAR(50) NOT NULL,
					title VARCHAR(255) NOT NULL,
					message TEXT NOT NULL,
					data JSON,
					read_at DATETIME,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
					INDEX idx_notifications_user_id (user_id),
					INDEX idx_notifications_user_read (user_id, read_at),
					INDEX idx_notifications_created_at (created_at DESC),
					INDEX idx_notifications_type (type),
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
					FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
				`
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			if _, err := db.Exec(createNotificationsSQL); err != nil {
				return fmt.Errorf("failed to create notifications table: %w", err)
			}
			fmt.Println("✓ Created notifications table")

			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			if _, err := db.Exec("DROP TABLE IF EXISTS notifications"); err != nil {
				return fmt.Errorf("failed to drop notifications table: %w", err)
			}
			fmt.Println("⚠️  WARNING: Notifications data has been deleted")
			return nil
		},
	},
	{
		Version:     "0.16.0",
		Description: "Add notification_likes table for liking notifications",
		Up: func(db *sql.DB, driver string) error {
			// Check if notification_likes table already exists
			hasNotificationLikes, err := checkTableExists(db, driver, "notification_likes")
			if err != nil {
				return fmt.Errorf("failed to check for notification_likes table: %w", err)
			}
			if hasNotificationLikes {
				fmt.Println("✓ notification_likes table already exists, skipping creation")
				return nil
			}

			// Create notification_likes table
			var createNotificationLikesSQL string
			switch driver {
			case "sqlite3":
				createNotificationLikesSQL = `
				CREATE TABLE notification_likes (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					notification_id INTEGER NOT NULL,
					user_id INTEGER NOT NULL,
					created_at DATETIME NOT NULL,
					FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
					UNIQUE (notification_id, user_id)
				);
				CREATE INDEX idx_notification_likes_notification_id ON notification_likes(notification_id);
				CREATE INDEX idx_notification_likes_user_id ON notification_likes(user_id);
				`
			case "postgres":
				createNotificationLikesSQL = `
				CREATE TABLE notification_likes (
					id BIGSERIAL PRIMARY KEY,
					notification_id BIGINT NOT NULL,
					user_id BIGINT NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
					UNIQUE (notification_id, user_id)
				);
				CREATE INDEX idx_notification_likes_notification_id ON notification_likes(notification_id);
				CREATE INDEX idx_notification_likes_user_id ON notification_likes(user_id);
				`
			case "mysql":
				createNotificationLikesSQL = `
				CREATE TABLE notification_likes (
					id BIGINT AUTO_INCREMENT PRIMARY KEY,
					notification_id BIGINT NOT NULL,
					user_id BIGINT NOT NULL,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					INDEX idx_notification_likes_notification_id (notification_id),
					INDEX idx_notification_likes_user_id (user_id),
					UNIQUE KEY unique_user_notification_like (notification_id, user_id),
					FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
				`
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			if _, err := db.Exec(createNotificationLikesSQL); err != nil {
				return fmt.Errorf("failed to create notification_likes table: %w", err)
			}
			fmt.Println("✓ Created notification_likes table")

			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			if _, err := db.Exec("DROP TABLE IF EXISTS notification_likes"); err != nil {
				return fmt.Errorf("failed to drop notification_likes table: %w", err)
			}
			fmt.Println("⚠️  WARNING: Notification likes data has been deleted")
			return nil
		},
	},
	{
		Version:     "0.17.0",
		Description: "Add missing indexes to audit_logs table for query optimization",
		Up: func(db *sql.DB, driver string) error {
			// Add index on target_user_id for queries filtering by affected user
			var createTargetUserIDIndex string
			switch driver {
			case "sqlite3":
				createTargetUserIDIndex = `CREATE INDEX IF NOT EXISTS idx_audit_logs_target_user_id ON audit_logs(target_user_id)`
			case "postgres":
				createTargetUserIDIndex = `CREATE INDEX IF NOT EXISTS idx_audit_logs_target_user_id ON audit_logs(target_user_id)`
			case "mysql":
				// MySQL doesn't support IF NOT EXISTS for indexes, need to check first
				var count int
				err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.statistics
					WHERE table_schema = DATABASE() AND table_name = 'audit_logs' AND index_name = 'idx_audit_logs_target_user_id'`).Scan(&count)
				if err != nil {
					return fmt.Errorf("failed to check for existing index: %w", err)
				}
				if count == 0 {
					createTargetUserIDIndex = `CREATE INDEX idx_audit_logs_target_user_id ON audit_logs(target_user_id)`
				}
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			if createTargetUserIDIndex != "" {
				if _, err := db.Exec(createTargetUserIDIndex); err != nil {
					return fmt.Errorf("failed to create idx_audit_logs_target_user_id: %w", err)
				}
				fmt.Println("✓ Created idx_audit_logs_target_user_id index")
			}

			// Add composite index for common query pattern (user_id, event_type, created_at)
			var createCompositeIndex string
			switch driver {
			case "sqlite3":
				createCompositeIndex = `CREATE INDEX IF NOT EXISTS idx_audit_logs_user_event ON audit_logs(user_id, event_type, created_at)`
			case "postgres":
				createCompositeIndex = `CREATE INDEX IF NOT EXISTS idx_audit_logs_user_event ON audit_logs(user_id, event_type, created_at)`
			case "mysql":
				var count int
				err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.statistics
					WHERE table_schema = DATABASE() AND table_name = 'audit_logs' AND index_name = 'idx_audit_logs_user_event'`).Scan(&count)
				if err != nil {
					return fmt.Errorf("failed to check for existing index: %w", err)
				}
				if count == 0 {
					createCompositeIndex = `CREATE INDEX idx_audit_logs_user_event ON audit_logs(user_id, event_type, created_at)`
				}
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			if createCompositeIndex != "" {
				if _, err := db.Exec(createCompositeIndex); err != nil {
					return fmt.Errorf("failed to create idx_audit_logs_user_event: %w", err)
				}
				fmt.Println("✓ Created idx_audit_logs_user_event composite index")
			}

			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			if _, err := db.Exec("DROP INDEX IF EXISTS idx_audit_logs_target_user_id"); err != nil {
				// MySQL syntax differs
				if driver == "mysql" {
					if _, err := db.Exec("DROP INDEX idx_audit_logs_target_user_id ON audit_logs"); err != nil {
						fmt.Printf("⚠️  Warning: could not drop idx_audit_logs_target_user_id: %v\n", err)
					}
				} else {
					fmt.Printf("⚠️  Warning: could not drop idx_audit_logs_target_user_id: %v\n", err)
				}
			}
			if _, err := db.Exec("DROP INDEX IF EXISTS idx_audit_logs_user_event"); err != nil {
				if driver == "mysql" {
					if _, err := db.Exec("DROP INDEX idx_audit_logs_user_event ON audit_logs"); err != nil {
						fmt.Printf("⚠️  Warning: could not drop idx_audit_logs_user_event: %v\n", err)
					}
				} else {
					fmt.Printf("⚠️  Warning: could not drop idx_audit_logs_user_event: %v\n", err)
				}
			}
			fmt.Println("⚠️  Removed audit_logs optimization indexes")
			return nil
		},
	},
	{
		Version:     "0.18.0",
		Description: "Add email_logs table for email audit trail",
		Up: func(db *sql.DB, driver string) error {
			// Check if email_logs table already exists
			hasEmailLogs, err := checkTableExists(db, driver, "email_logs")
			if err != nil {
				return fmt.Errorf("failed to check for email_logs table: %w", err)
			}
			if hasEmailLogs {
				fmt.Println("✓ email_logs table already exists, skipping creation")
				return nil
			}

			// Create email_logs table
			var createEmailLogsSQL string
			switch driver {
			case "sqlite3":
				createEmailLogsSQL = `
				CREATE TABLE email_logs (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					recipient_email TEXT NOT NULL,
					email_type TEXT NOT NULL,
					subject TEXT NOT NULL,
					success INTEGER NOT NULL DEFAULT 0,
					error_message TEXT,
					debug_info TEXT,
					sent_by_user_id INTEGER,
					created_at DATETIME NOT NULL,
					FOREIGN KEY (sent_by_user_id) REFERENCES users(id) ON DELETE SET NULL
				);
				CREATE INDEX idx_email_logs_type ON email_logs(email_type);
				CREATE INDEX idx_email_logs_recipient ON email_logs(recipient_email);
				CREATE INDEX idx_email_logs_created_at ON email_logs(created_at DESC);
				CREATE INDEX idx_email_logs_success ON email_logs(success);
				`
			case "postgres":
				createEmailLogsSQL = `
				CREATE TABLE email_logs (
					id BIGSERIAL PRIMARY KEY,
					recipient_email VARCHAR(255) NOT NULL,
					email_type VARCHAR(50) NOT NULL,
					subject VARCHAR(500) NOT NULL,
					success BOOLEAN NOT NULL DEFAULT FALSE,
					error_message TEXT,
					debug_info TEXT,
					sent_by_user_id BIGINT,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (sent_by_user_id) REFERENCES users(id) ON DELETE SET NULL
				);
				CREATE INDEX idx_email_logs_type ON email_logs(email_type);
				CREATE INDEX idx_email_logs_recipient ON email_logs(recipient_email);
				CREATE INDEX idx_email_logs_created_at ON email_logs(created_at DESC);
				CREATE INDEX idx_email_logs_success ON email_logs(success);
				`
			case "mysql":
				createEmailLogsSQL = `
				CREATE TABLE email_logs (
					id BIGINT AUTO_INCREMENT PRIMARY KEY,
					recipient_email VARCHAR(255) NOT NULL,
					email_type VARCHAR(50) NOT NULL,
					subject VARCHAR(500) NOT NULL,
					success BOOLEAN NOT NULL DEFAULT FALSE,
					error_message TEXT,
					debug_info TEXT,
					sent_by_user_id BIGINT,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					INDEX idx_email_logs_type (email_type),
					INDEX idx_email_logs_recipient (recipient_email),
					INDEX idx_email_logs_created_at (created_at DESC),
					INDEX idx_email_logs_success (success),
					FOREIGN KEY (sent_by_user_id) REFERENCES users(id) ON DELETE SET NULL
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
				`
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			if _, err := db.Exec(createEmailLogsSQL); err != nil {
				return fmt.Errorf("failed to create email_logs table: %w", err)
			}
			fmt.Println("✓ Created email_logs table")

			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			if _, err := db.Exec("DROP TABLE IF EXISTS email_logs"); err != nil {
				return fmt.Errorf("failed to drop email_logs table: %w", err)
			}
			fmt.Println("⚠️  WARNING: Email logs data has been deleted")
			return nil
		},
	},
	{
		Version:     "0.19.0",
		Description: "Add timezone column to user_settings table",
		Up: func(db *sql.DB, driver string) error {
			// Check if timezone column already exists
			hasTimezone, err := checkColumnExists(db, driver, "user_settings", "timezone")
			if err != nil {
				return fmt.Errorf("failed to check for timezone column: %w", err)
			}
			if hasTimezone {
				fmt.Println("✓ timezone column already exists in user_settings, skipping")
				return nil
			}

			// Add timezone column with default value
			var alterSQL string
			switch driver {
			case "sqlite3":
				alterSQL = `ALTER TABLE user_settings ADD COLUMN timezone TEXT DEFAULT 'America/New_York'`
			case "postgres":
				alterSQL = `ALTER TABLE user_settings ADD COLUMN timezone VARCHAR(50) DEFAULT 'America/New_York'`
			case "mysql":
				alterSQL = `ALTER TABLE user_settings ADD COLUMN timezone VARCHAR(50) DEFAULT 'America/New_York'`
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			if _, err := db.Exec(alterSQL); err != nil {
				return fmt.Errorf("failed to add timezone column: %w", err)
			}
			fmt.Println("✓ Added timezone column to user_settings table")

			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			// SQLite doesn't support DROP COLUMN easily, so we skip it
			switch driver {
			case "postgres", "mysql":
				if _, err := db.Exec("ALTER TABLE user_settings DROP COLUMN timezone"); err != nil {
					return fmt.Errorf("failed to drop timezone column: %w", err)
				}
				fmt.Println("✓ Dropped timezone column from user_settings")
			case "sqlite3":
				fmt.Println("⚠️  SQLite does not support DROP COLUMN, timezone column remains")
			}
			return nil
		},
	},
	{
		Version:     "0.20.0",
		Description: "Add admin_user_event_notifications column to user_settings for admin email preferences",
		Up: func(db *sql.DB, driver string) error {
			// Check if column already exists
			hasColumn, err := checkColumnExists(db, driver, "user_settings", "admin_user_event_notifications")
			if err != nil {
				return fmt.Errorf("failed to check for admin_user_event_notifications column: %w", err)
			}
			if hasColumn {
				fmt.Println("✓ admin_user_event_notifications column already exists in user_settings, skipping")
				return nil
			}

			// Add column with default value TRUE (admins receive notifications by default)
			var alterSQL string
			switch driver {
			case "sqlite3":
				alterSQL = `ALTER TABLE user_settings ADD COLUMN admin_user_event_notifications INTEGER NOT NULL DEFAULT 1`
			case "postgres":
				alterSQL = `ALTER TABLE user_settings ADD COLUMN admin_user_event_notifications BOOLEAN NOT NULL DEFAULT TRUE`
			case "mysql":
				alterSQL = `ALTER TABLE user_settings ADD COLUMN admin_user_event_notifications BOOLEAN NOT NULL DEFAULT TRUE`
			default:
				return fmt.Errorf("unsupported database driver: %s", driver)
			}

			if _, err := db.Exec(alterSQL); err != nil {
				return fmt.Errorf("failed to add admin_user_event_notifications column: %w", err)
			}
			fmt.Println("✓ Added admin_user_event_notifications column to user_settings table")

			return nil
		},
		Down: func(db *sql.DB, driver string) error {
			switch driver {
			case "postgres", "mysql":
				if _, err := db.Exec("ALTER TABLE user_settings DROP COLUMN admin_user_event_notifications"); err != nil {
					return fmt.Errorf("failed to drop admin_user_event_notifications column: %w", err)
				}
				fmt.Println("✓ Dropped admin_user_event_notifications column from user_settings")
			case "sqlite3":
				fmt.Println("⚠️  SQLite does not support DROP COLUMN, admin_user_event_notifications column remains")
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
