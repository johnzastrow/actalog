package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

// currentDriver stores the database driver being used
var currentDriver string

// BuildDSN constructs a database connection string based on the driver type
func BuildDSN(driver, host string, port int, user, password, database, sslMode, schema string) string {
	switch driver {
	case "sqlite3":
		// For SQLite, database is the file path
		return database

	case "postgres":
		// PostgreSQL connection string (pgx format)
		// Format: postgres://user:password@host:port/database?sslmode=disable&search_path=schema
		dsn := fmt.Sprintf("postgres://%s", user)
		if password != "" {
			dsn = fmt.Sprintf("postgres://%s:%s", user, password)
		}
		dsn = fmt.Sprintf("%s@%s:%d/%s?sslmode=%s", dsn, host, port, database, sslMode)
		if schema != "" && schema != "public" {
			dsn = fmt.Sprintf("%s&search_path=%s", dsn, schema)
		}
		return dsn

	case "mysql":
		// MySQL/MariaDB connection string
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
			user, password, host, port, database)
		return dsn

	default:
		// Fallback: return database as-is
		return database
	}
}

// InitDatabase initializes the database connection and runs migrations
func InitDatabase(driver, dsn string, dbConfig interface{}) (*sql.DB, error) {
	// Store driver for later use
	currentDriver = driver

	// Use pgx driver name for PostgreSQL
	driverName := driver
	if driver == "postgres" {
		driverName = "pgx"
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pooling for PostgreSQL and MySQL
	if driver == "postgres" || driver == "mysql" {
		// Type assert to get database config
		// This is safe because the caller passes configs.DatabaseConfig
		if cfg, ok := dbConfig.(interface {
			GetMaxOpenConns() int
			GetMaxIdleConns() int
			GetConnMaxLifetime() time.Duration
		}); ok {
			db.SetMaxOpenConns(cfg.GetMaxOpenConns())
			db.SetMaxIdleConns(cfg.GetMaxIdleConns())
			db.SetConnMaxLifetime(cfg.GetConnMaxLifetime())
		} else {
			// Fallback to default values if type assertion fails
			db.SetMaxOpenConns(25)
			db.SetMaxIdleConns(5)
			db.SetConnMaxLifetime(5 * time.Minute)
		}
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// For new databases, create initial tables (v0.1.0 schema)
	// This ensures the database is initialized before running migrations
	if err := createInitialTablesIfNotExist(db, driver); err != nil {
		return nil, fmt.Errorf("failed to create initial tables: %w", err)
	}

	// Run migrations to bring schema up to latest version
	fmt.Println("Running database migrations...")
	if err := RunMigrations(db, driver); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Seed standard movements (if not already seeded)
	if err := seedStandardMovements(db); err != nil {
		return nil, fmt.Errorf("failed to seed standard movements: %w", err)
	}

	// Seed standard WODs (if not already seeded)
	if err := seedStandardWODs(db); err != nil {
		return nil, fmt.Errorf("failed to seed standard WODs: %w", err)
	}

	// Seed workout templates (if not already seeded)
	if err := seedWorkoutTemplates(db); err != nil {
		return nil, fmt.Errorf("failed to seed workout templates: %w", err)
	}

	return db, nil
}

// createInitialTablesIfNotExist creates the initial v0.1.0 schema if tables don't exist
func createInitialTablesIfNotExist(db *sql.DB, driver string) error {
	// Check if users table exists using driver-specific query
	tableExists, err := checkTableExists(db, driver, "users")
	if err != nil {
		return fmt.Errorf("failed to check if users table exists: %w", err)
	}

	if tableExists {
		// Tables already exist, skip initialization
		return nil
	}

	fmt.Println("Initializing new database with v0.1.0 schema...")
	return createTables(db, driver)
}

// checkTableExists checks if a table exists in the database
func checkTableExists(db *sql.DB, driver, tableName string) (bool, error) {
	var query string
	var result interface{}

	switch driver {
	case "sqlite3":
		query = "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
		var name string
		result = &name

	case "postgres":
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_name=$1"
		var name string
		result = &name

	case "mysql":
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?"
		var name string
		result = &name

	default:
		return false, fmt.Errorf("unsupported database driver: %s", driver)
	}

	err := db.QueryRow(query, tableName).Scan(result)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// createTables creates all necessary database tables using driver-specific SQL
func createTables(db *sql.DB, driver string) error {
	var schema string

	switch driver {
	case "sqlite3":
		schema = getSQLiteSchema()
	case "postgres":
		schema = getPostgreSQLSchema()
	case "mysql":
		schema = getMySQLSchema()
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	// Split schema into individual statements for better error reporting
	statements := strings.Split(schema, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute schema statement: %w\nStatement: %s", err, stmt)
		}
	}

	return nil
}

// getSQLiteSchema returns the SQLite-specific schema
func getSQLiteSchema() string {
	return `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		name TEXT NOT NULL,
		profile_image TEXT,
		birthday DATE,
		role TEXT NOT NULL DEFAULT 'user',
		email_verified INTEGER NOT NULL DEFAULT 0,
		email_verified_at DATETIME,
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

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

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
	CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);

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

	CREATE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings(user_id);

	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		target_user_id INTEGER,
		event_type TEXT NOT NULL,
		ip_address TEXT,
		user_agent TEXT,
		details TEXT,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
		FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_event_type ON audit_logs(event_type);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);

	CREATE TABLE IF NOT EXISTS workouts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		notes TEXT,
		created_by INTEGER,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_workouts_created_by ON workouts(created_by);
	CREATE INDEX IF NOT EXISTS idx_workouts_name ON workouts(name);

	CREATE TABLE IF NOT EXISTS movements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		description TEXT,
		type TEXT NOT NULL,
		is_standard INTEGER NOT NULL DEFAULT 0,
		created_by INTEGER,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_movements_name ON movements(name);
	CREATE INDEX IF NOT EXISTS idx_movements_type ON movements(type);
	CREATE INDEX IF NOT EXISTS idx_movements_standard ON movements(is_standard);

	CREATE TABLE IF NOT EXISTS wods (
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
	);

	CREATE INDEX IF NOT EXISTS idx_wods_name ON wods(name);
	CREATE INDEX IF NOT EXISTS idx_wods_type ON wods(type);
	CREATE INDEX IF NOT EXISTS idx_wods_is_standard ON wods(is_standard);

	CREATE TABLE IF NOT EXISTS workout_wods (
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
	);

	CREATE INDEX IF NOT EXISTS idx_workout_wods_workout_id ON workout_wods(workout_id);
	CREATE INDEX IF NOT EXISTS idx_workout_wods_wod_id ON workout_wods(wod_id);

	CREATE TABLE IF NOT EXISTS user_workouts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		workout_id INTEGER,
		workout_name TEXT,
		workout_date DATE NOT NULL,
		workout_type TEXT,
		total_time INTEGER,
		notes TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE RESTRICT
	);

	CREATE INDEX IF NOT EXISTS idx_user_workouts_user_id ON user_workouts(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_workouts_workout_date ON user_workouts(workout_date);
	CREATE INDEX IF NOT EXISTS idx_user_workouts_user_date ON user_workouts(user_id, workout_date DESC);

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
	CREATE INDEX IF NOT EXISTS idx_user_workout_movements_movement_id ON user_workout_movements(movement_id);

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
	CREATE INDEX IF NOT EXISTS idx_user_workout_wods_wod_id ON user_workout_wods(wod_id);

	CREATE TABLE IF NOT EXISTS workout_movements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workout_id INTEGER NOT NULL,
		movement_id INTEGER NOT NULL,
		weight REAL,
		sets INTEGER,
		reps INTEGER,
		time INTEGER,
		distance REAL,
		is_rx INTEGER NOT NULL DEFAULT 0,
		is_pr INTEGER NOT NULL DEFAULT 0,
		notes TEXT,
		order_index INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (movement_id) REFERENCES movements(id) ON DELETE RESTRICT
	);

	CREATE INDEX IF NOT EXISTS idx_wm_workout_id ON workout_movements(workout_id);
	CREATE INDEX IF NOT EXISTS idx_wm_movement_id ON workout_movements(movement_id);
	CREATE INDEX IF NOT EXISTS idx_wm_workout_order ON workout_movements(workout_id, order_index);
	`
}

// getPostgreSQLSchema returns the PostgreSQL-specific schema
func getPostgreSQLSchema() string {
	return `
	CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		profile_image TEXT,
		birthday DATE,
		role VARCHAR(50) NOT NULL DEFAULT 'user',
		email_verified BOOLEAN NOT NULL DEFAULT FALSE,
		email_verified_at TIMESTAMP,
		failed_login_attempts INTEGER NOT NULL DEFAULT 0,
		locked_at TIMESTAMP,
		locked_until TIMESTAMP,
		account_disabled BOOLEAN NOT NULL DEFAULT FALSE,
		disabled_at TIMESTAMP,
		disabled_by_user_id BIGINT,
		disable_reason TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_login_at TIMESTAMP,
		FOREIGN KEY (disabled_by_user_id) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

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
	CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);

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

	CREATE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings(user_id);

	CREATE TABLE IF NOT EXISTS audit_logs (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT,
		target_user_id BIGINT,
		event_type VARCHAR(100) NOT NULL,
		ip_address VARCHAR(50),
		user_agent TEXT,
		details TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
		FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_event_type ON audit_logs(event_type);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);

	CREATE TABLE IF NOT EXISTS workouts (
		id BIGSERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		notes TEXT,
		created_by BIGINT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_workouts_created_by ON workouts(created_by);
	CREATE INDEX IF NOT EXISTS idx_workouts_name ON workouts(name);

	CREATE TABLE IF NOT EXISTS movements (
		id BIGSERIAL PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		description TEXT,
		type VARCHAR(50) NOT NULL,
		is_standard BOOLEAN NOT NULL DEFAULT FALSE,
		created_by BIGINT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_movements_name ON movements(name);
	CREATE INDEX IF NOT EXISTS idx_movements_type ON movements(type);
	CREATE INDEX IF NOT EXISTS idx_movements_standard ON movements(is_standard);

	CREATE TABLE IF NOT EXISTS wods (
		id BIGSERIAL PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		source VARCHAR(255),
		type VARCHAR(255),
		regime VARCHAR(255),
		score_type VARCHAR(255),
		description TEXT,
		url TEXT,
		notes TEXT,
		is_standard BOOLEAN NOT NULL DEFAULT FALSE,
		created_by BIGINT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_wods_name ON wods(name);
	CREATE INDEX IF NOT EXISTS idx_wods_type ON wods(type);
	CREATE INDEX IF NOT EXISTS idx_wods_is_standard ON wods(is_standard);

	CREATE TABLE IF NOT EXISTS workout_wods (
		id BIGSERIAL PRIMARY KEY,
		workout_id BIGINT NOT NULL,
		wod_id BIGINT NOT NULL,
		score_value TEXT,
		division TEXT,
		is_pr BOOLEAN NOT NULL DEFAULT false,
		order_index INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT
	);

	CREATE INDEX IF NOT EXISTS idx_workout_wods_workout_id ON workout_wods(workout_id);
	CREATE INDEX IF NOT EXISTS idx_workout_wods_wod_id ON workout_wods(wod_id);

	CREATE TABLE IF NOT EXISTS user_workouts (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		workout_id BIGINT,
		workout_name TEXT,
		workout_date DATE NOT NULL,
		workout_type VARCHAR(255),
		total_time INTEGER,
		notes TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE RESTRICT
	);

	CREATE INDEX IF NOT EXISTS idx_user_workouts_user_id ON user_workouts(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_workouts_workout_date ON user_workouts(workout_date);
	CREATE INDEX IF NOT EXISTS idx_user_workouts_user_date ON user_workouts(user_id, workout_date DESC);

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
	CREATE INDEX IF NOT EXISTS idx_user_workout_movements_movement_id ON user_workout_movements(movement_id);

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
	CREATE INDEX IF NOT EXISTS idx_user_workout_wods_wod_id ON user_workout_wods(wod_id);

	CREATE TABLE IF NOT EXISTS workout_movements (
		id BIGSERIAL PRIMARY KEY,
		workout_id BIGINT NOT NULL,
		movement_id BIGINT NOT NULL,
		weight DOUBLE PRECISION,
		sets INTEGER,
		reps INTEGER,
		time INTEGER,
		distance DOUBLE PRECISION,
		is_rx BOOLEAN NOT NULL DEFAULT FALSE,
		is_pr BOOLEAN NOT NULL DEFAULT FALSE,
		notes TEXT,
		order_index INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (movement_id) REFERENCES movements(id) ON DELETE RESTRICT
	);

	CREATE INDEX IF NOT EXISTS idx_wm_workout_id ON workout_movements(workout_id);
	CREATE INDEX IF NOT EXISTS idx_wm_movement_id ON workout_movements(movement_id);
	CREATE INDEX IF NOT EXISTS idx_wm_workout_order ON workout_movements(workout_id, order_index);
	`
}

// getMySQLSchema returns the MySQL/MariaDB-specific schema
func getMySQLSchema() string {
	return `
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		profile_image TEXT,
		birthday DATE,
		role VARCHAR(50) NOT NULL DEFAULT 'user',
		email_verified BOOLEAN NOT NULL DEFAULT FALSE,
		email_verified_at DATETIME,
		failed_login_attempts INTEGER NOT NULL DEFAULT 0,
		locked_at DATETIME,
		locked_until DATETIME,
		account_disabled BOOLEAN NOT NULL DEFAULT FALSE,
		disabled_at DATETIME,
		disabled_by_user_id BIGINT,
		disable_reason TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		last_login_at DATETIME,
		INDEX idx_users_email (email),
		INDEX idx_users_role (role),
		FOREIGN KEY (disabled_by_user_id) REFERENCES users(id) ON DELETE SET NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS audit_logs (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT,
		target_user_id BIGINT,
		event_type VARCHAR(100) NOT NULL,
		ip_address VARCHAR(50),
		user_agent TEXT,
		details TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_audit_logs_user_id (user_id),
		INDEX idx_audit_logs_event_type (event_type),
		INDEX idx_audit_logs_created_at (created_at DESC),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
		FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE SET NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS workouts (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		notes TEXT,
		created_by BIGINT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
		INDEX idx_workouts_created_by (created_by),
		INDEX idx_workouts_name (name)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS movements (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		description TEXT,
		type VARCHAR(50) NOT NULL,
		is_standard BOOLEAN NOT NULL DEFAULT FALSE,
		created_by BIGINT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
		INDEX idx_movements_name (name),
		INDEX idx_movements_type (type),
		INDEX idx_movements_standard (is_standard)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS wods (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		source VARCHAR(255),
		type VARCHAR(255),
		regime VARCHAR(255),
		score_type VARCHAR(255),
		description TEXT,
		url TEXT,
		notes TEXT,
		is_standard BOOLEAN NOT NULL DEFAULT FALSE,
		created_by BIGINT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
		INDEX idx_wods_name (name),
		INDEX idx_wods_type (type),
		INDEX idx_wods_is_standard (is_standard)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS workout_wods (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		workout_id BIGINT NOT NULL,
		wod_id BIGINT NOT NULL,
		score_value TEXT,
		division TEXT,
		is_pr BOOLEAN NOT NULL DEFAULT 0,
		order_index INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT,
		INDEX idx_workout_wods_workout_id (workout_id),
		INDEX idx_workout_wods_wod_id (wod_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS user_workouts (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT NOT NULL,
		workout_id BIGINT,
		workout_name TEXT,
		workout_date DATE NOT NULL,
		workout_type VARCHAR(255),
		total_time INTEGER,
		notes TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE RESTRICT,
		INDEX idx_user_workouts_user_id (user_id),
		INDEX idx_user_workouts_workout_date (workout_date),
		INDEX idx_user_workouts_user_date (user_id, workout_date DESC)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
		FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (movement_id) REFERENCES movements(id) ON DELETE RESTRICT,
		INDEX idx_user_workout_movements_user_workout_id (user_workout_id),
		INDEX idx_user_workout_movements_movement_id (movement_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
		FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT,
		INDEX idx_user_workout_wods_user_workout_id (user_workout_id),
		INDEX idx_user_workout_wods_wod_id (wod_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS workout_movements (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		workout_id BIGINT NOT NULL,
		movement_id BIGINT NOT NULL,
		weight DOUBLE,
		sets INTEGER,
		reps INTEGER,
		time INTEGER,
		distance DOUBLE,
		is_rx BOOLEAN NOT NULL DEFAULT FALSE,
		is_pr BOOLEAN NOT NULL DEFAULT FALSE,
		notes TEXT,
		order_index INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (movement_id) REFERENCES movements(id) ON DELETE RESTRICT,
		INDEX idx_wm_workout_id (workout_id),
		INDEX idx_wm_movement_id (movement_id),
		INDEX idx_wm_workout_order (workout_id, order_index)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
}

// getBoolValue returns database-specific boolean value for WHERE clauses
func getBoolValue(driver string, value bool) string {
	if driver == "sqlite3" {
		if value {
			return "1"
		}
		return "0"
	}
	// PostgreSQL and MySQL use TRUE/FALSE
	if value {
		return "TRUE"
	}
	return "FALSE"
}

// getPlaceholders returns database-specific placeholder syntax
// count is the number of placeholders needed
// Returns a slice of placeholder strings like ["?", "?"] or ["$1", "$2"]
func getPlaceholders(driver string, count int) []string {
	placeholders := make([]string, count)
	if driver == "postgres" {
		// PostgreSQL uses $1, $2, $3, etc.
		for i := 0; i < count; i++ {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
	} else {
		// SQLite and MySQL use ?
		for i := 0; i < count; i++ {
			placeholders[i] = "?"
		}
	}
	return placeholders
}

// getTimestampFunc returns the database-specific function for current timestamp
func getTimestampFunc() string {
	switch currentDriver {
	case "sqlite3":
		return "datetime('now')"
	case "postgres":
		return "CURRENT_TIMESTAMP"
	case "mysql":
		return "NOW()"
	default:
		return "CURRENT_TIMESTAMP"
	}
}

// rebindQuery converts ? placeholders to $N for PostgreSQL
// This allows queries written with ? placeholders to work across all databases
func rebindQuery(query string) string {
	if currentDriver != "postgres" {
		return query
	}

	// Convert ? to $1, $2, etc for PostgreSQL
	result := make([]byte, 0, len(query))
	paramNum := 1
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			result = append(result, fmt.Sprintf("$%d", paramNum)...)
			paramNum++
		} else {
			result = append(result, query[i])
		}
	}
	return string(result)
}

// seedStandardMovements seeds the database with standard CrossFit movements
func seedStandardMovements(db *sql.DB) error {
	// Determine target table before querying (migrations may rename it)
	targetTable := "movements"
	if ok, _ := checkTableExists(db, currentDriver, "movements"); !ok {
		if ok2, _ := checkTableExists(db, currentDriver, "strength_movements"); ok2 {
			targetTable = "strength_movements"
		} else {
			// No movements table found; nothing to seed
			return nil
		}
	}

	// Check if movements already exist in the target table
	var count int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE is_standard = %s", targetTable, getBoolValue(currentDriver, true))
	err := db.QueryRow(countQuery).Scan(&count)
	if err != nil {
		return err
	}

	// If movements already seeded, skip
	if count > 0 {
		return nil
	}

	// Standard movements to seed
	movements := []struct {
		name        string
		description string
		movType     string
	}{
		// Weightlifting
		{"Back Squat", "Barbell back squat", "weightlifting"},
		{"Front Squat", "Barbell front squat", "weightlifting"},
		{"Overhead Squat", "Barbell overhead squat", "weightlifting"},
		{"Deadlift", "Conventional deadlift", "weightlifting"},
		{"Sumo Deadlift High Pull", "SDHP with barbell", "weightlifting"},
		{"Clean", "Full clean", "weightlifting"},
		{"Power Clean", "Power clean (squat above parallel)", "weightlifting"},
		{"Hang Clean", "Clean from hang position", "weightlifting"},
		{"Snatch", "Full snatch", "weightlifting"},
		{"Power Snatch", "Power snatch (squat above parallel)", "weightlifting"},
		{"Clean and Jerk", "Full clean and jerk", "weightlifting"},
		{"Thruster", "Front squat to push press", "weightlifting"},
		{"Push Press", "Barbell push press", "weightlifting"},
		{"Push Jerk", "Barbell push jerk", "weightlifting"},
		{"Split Jerk", "Barbell split jerk", "weightlifting"},

		// Gymnastics
		{"Pull-up", "Strict or kipping pull-up", "gymnastics"},
		{"Chest-to-Bar Pull-up", "Pull-up with chest touching bar", "gymnastics"},
		{"Muscle-up", "Ring or bar muscle-up", "gymnastics"},
		{"Handstand Push-up", "HSPU against wall or freestanding", "gymnastics"},
		{"Dip", "Ring or bar dip", "gymnastics"},
		{"Toes-to-Bar", "Hanging toes to bar", "gymnastics"},
		{"Knees-to-Elbow", "Hanging knees to elbows", "gymnastics"},

		// Bodyweight
		{"Push-up", "Standard push-up", "bodyweight"},
		{"Sit-up", "Abdominal sit-up", "bodyweight"},
		{"Air Squat", "Bodyweight squat", "bodyweight"},
		{"Burpee", "Full burpee", "bodyweight"},
		{"Box Jump", "Jump onto box", "bodyweight"},

		// Cardio
		{"Row", "Rowing machine (meters or calories)", "cardio"},
		{"Run", "Running (meters or miles)", "cardio"},
		{"Bike", "Assault bike or stationary bike", "cardio"},
		{"Ski Erg", "Ski erg machine", "cardio"},

		// Other
		{"Kettlebell Swing", "Kettlebell swing", "weightlifting"},
	}

	// Get database-specific timestamp function
	var timestampFunc string
	switch currentDriver {
	case "sqlite3":
		timestampFunc = "datetime('now')"
	case "postgres":
		timestampFunc = "CURRENT_TIMESTAMP"
	case "mysql":
		timestampFunc = "NOW()"
	default:
		timestampFunc = "CURRENT_TIMESTAMP"
	}

	// Prepare insert statement with database-specific placeholders, timestamp and boolean
	ph := getPlaceholders(currentDriver, 3) // 3 parameters: name, description, type
	stmt := fmt.Sprintf(`
		INSERT INTO %s (name, description, type, is_standard, created_by, created_at, updated_at)
		VALUES (%s, %s, %s, %s, NULL, %s, %s)
	`, targetTable, ph[0], ph[1], ph[2], getBoolValue(currentDriver, true), timestampFunc, timestampFunc)

	// Insert each movement
	for _, m := range movements {
		_, err := db.Exec(stmt, m.name, m.description, m.movType)
		if err != nil {
			return fmt.Errorf("failed to seed movement %s: %w", m.name, err)
		}
	}

	return nil
}

// seedStandardWODs seeds the database with famous CrossFit benchmark WODs
func seedStandardWODs(db *sql.DB) error {
	// Check if WODs already seeded (check for "Fran" - a very famous benchmark WOD)
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM wods WHERE name = 'Fran' AND is_standard = %s", getBoolValue(currentDriver, true))
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for existing WODs: %w", err)
	}
	
	if count > 0 {
		return nil // Already seeded
	}

	// List of standard CrossFit WODs (Girls, Heroes, and other benchmarks)
	wods := []struct {
		name        string
		source      string
		wodType     string
		regime      string
		scoreType   string
		description string
		url         string
	}{
		// Girls - Classic CrossFit Benchmark WODs
		{
			name:        "Fran",
			source:      "CrossFit",
			wodType:     "Girl",
			regime:      "Fastest Time",
			scoreType:   "Time (HH:MM:SS)",
			description: "21-15-9 reps for time of: Thrusters (95/65 lb), Pull-ups",
			url:         "https://www.crossfit.com/workout/fran",
		},
		{
			name:        "Helen",
			source:      "CrossFit",
			wodType:     "Girl",
			regime:      "Fastest Time",
			scoreType:   "Time (HH:MM:SS)",
			description: "3 rounds for time of: 400m Run, 21 Kettlebell Swings (53/35 lb), 12 Pull-ups",
			url:         "https://www.crossfit.com/workout/helen",
		},
		{
			name:        "Cindy",
			source:      "CrossFit",
			wodType:     "Girl",
			regime:      "AMRAP",
			scoreType:   "Rounds+Reps",
			description: "20 min AMRAP of: 5 Pull-ups, 10 Push-ups, 15 Air Squats",
			url:         "https://www.crossfit.com/workout/cindy",
		},
		{
			name:        "Grace",
			source:      "CrossFit",
			wodType:     "Girl",
			regime:      "Fastest Time",
			scoreType:   "Time (HH:MM:SS)",
			description: "30 Clean and Jerks for time (135/95 lb)",
			url:         "https://www.crossfit.com/workout/grace",
		},
		{
			name:        "Annie",
			source:      "CrossFit",
			wodType:     "Girl",
			regime:      "Fastest Time",
			scoreType:   "Time (HH:MM:SS)",
			description: "50-40-30-20-10 reps for time of: Double-Unders, Sit-ups",
			url:         "https://www.crossfit.com/workout/annie",
		},
		{
			name:        "Karen",
			source:      "CrossFit",
			wodType:     "Girl",
			regime:      "Fastest Time",
			scoreType:   "Time (HH:MM:SS)",
			description: "150 Wall Ball Shots for time (20/14 lb, 10/9 ft)",
			url:         "https://www.crossfit.com/workout/karen",
		},
		{
			name:        "Diane",
			source:      "CrossFit",
			wodType:     "Girl",
			regime:      "Fastest Time",
			scoreType:   "Time (HH:MM:SS)",
			description: "21-15-9 reps for time of: Deadlifts (225/155 lb), Handstand Push-ups",
			url:         "https://www.crossfit.com/workout/diane",
		},
		{
			name:        "Elizabeth",
			source:      "CrossFit",
			wodType:     "Girl",
			regime:      "Fastest Time",
			scoreType:   "Time (HH:MM:SS)",
			description: "21-15-9 reps for time of: Cleans (135/95 lb), Dips",
			url:         "https://www.crossfit.com/workout/elizabeth",
		},
		// Hero WODs - Named after fallen military/first responders
		{
			name:        "Murph",
			source:      "CrossFit",
			wodType:     "Hero",
			regime:      "Fastest Time",
			scoreType:   "Time (HH:MM:SS)",
			description: "For time: 1 mile Run, 100 Pull-ups, 200 Push-ups, 300 Air Squats, 1 mile Run (wear 20 lb vest if possible)",
			url:         "https://www.crossfit.com/workout/murph",
		},
		{
			name:        "DT",
			source:      "CrossFit",
			wodType:     "Hero",
			regime:      "Fastest Time",
			scoreType:   "Time (HH:MM:SS)",
			description: "5 rounds for time of: 12 Deadlifts (155/105 lb), 9 Hang Power Cleans (155/105 lb), 6 Push Jerks (155/105 lb)",
			url:         "https://www.crossfit.com/workout/dt",
		},
	}

	// Determine database-specific timestamp function
	var timestampFunc string
	switch currentDriver {
	case "sqlite3":
		timestampFunc = "datetime('now')"
	case "postgres":
		timestampFunc = "CURRENT_TIMESTAMP"
	case "mysql":
		timestampFunc = "NOW()"
	default:
		timestampFunc = "CURRENT_TIMESTAMP"
	}

	// Prepare insert statement with database-specific placeholders, timestamp and boolean
	ph := getPlaceholders(currentDriver, 7) // 7 parameters: name, source, type, regime, score_type, description, url
	insertQuery := fmt.Sprintf(`INSERT INTO wods (name, source, type, regime, score_type, description, url, is_standard, created_by, created_at, updated_at)
	          VALUES (%s, %s, %s, %s, %s, %s, %s, %s, NULL, %s, %s)`,
		ph[0], ph[1], ph[2], ph[3], ph[4], ph[5], ph[6], getBoolValue(currentDriver, true), timestampFunc, timestampFunc)

	for _, wod := range wods {
		_, err := db.Exec(insertQuery, wod.name, wod.source, wod.wodType, wod.regime, wod.scoreType, wod.description, wod.url)
		if err != nil {
			return fmt.Errorf("failed to seed WOD %s: %w", wod.name, err)
		}
	}

	return nil
}

// seedWorkoutTemplates seeds the database with sample workout templates
// This demonstrates the template-based system with movements and WODs
func seedWorkoutTemplates(db *sql.DB) error {
	// Check if workout templates already seeded (check for "Strength Training - Back Squat Focus")
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM workouts WHERE name = 'Strength Training - Back Squat Focus'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for existing workout templates: %w", err)
	}
	
	if count > 0 {
		return nil // Already seeded
	}

	// Template 1: Strength Training - Back Squat Focus
	workoutID, err := createWorkout(db, "Strength Training - Back Squat Focus", "5x5 progressive overload program")
	if err != nil {
		return err
	}
	
	// Get movement IDs
	backSquatID, err := getMovementIDByName(db, "Back Squat")
	if err != nil {
		return err
	}
	
	// Add movements to workout
	if err := addWorkoutMovement(db, workoutID, backSquatID, 225.0, 5, 5, 0); err != nil {
		return err
	}

	// Template 2: Olympic Lifting - Clean & Jerk Practice
	workoutID, err = createWorkout(db, "Olympic Lifting - Clean & Jerk Practice", "Technical practice with moderate weight")
	if err != nil {
		return err
	}
	
	cleanID, err := getMovementIDByName(db, "Clean")
	if err != nil {
		return err
	}
	jerkID, err := getMovementIDByName(db, "Push Jerk")
	if err != nil {
		return err
	}
	
	if err := addWorkoutMovement(db, workoutID, cleanID, 135.0, 5, 3, 0); err != nil {
		return err
	}
	if err := addWorkoutMovement(db, workoutID, jerkID, 135.0, 5, 3, 1); err != nil {
		return err
	}

	// Template 3: Gymnastics Strength
	workoutID, err = createWorkout(db, "Gymnastics Strength", "Bodyweight strength and skill work")
	if err != nil {
		return err
	}
	
	pullupID, err := getMovementIDByName(db, "Pull-up")
	if err != nil {
		return err
	}
	dipID, err := getMovementIDByName(db, "Dip")
	if err != nil {
		return err
	}
	hspuID, err := getMovementIDByName(db, "Handstand Push-up")
	if err != nil {
		return err
	}
	
	if err := addWorkoutMovement(db, workoutID, pullupID, 0, 5, 10, 0); err != nil {
		return err
	}
	if err := addWorkoutMovement(db, workoutID, dipID, 0, 5, 10, 1); err != nil {
		return err
	}
	if err := addWorkoutMovement(db, workoutID, hspuID, 0, 5, 5, 2); err != nil {
		return err
	}

	// Template 4: Cardio Endurance
	workoutID, err = createWorkout(db, "Cardio Endurance", "Mixed cardio modalities")
	if err != nil {
		return err
	}
	
	runID, err := getMovementIDByName(db, "Run")
	if err != nil {
		return err
	}
	rowID, err := getMovementIDByName(db, "Row")
	if err != nil {
		return err
	}
	bikeID, err := getMovementIDByName(db, "Bike")
	if err != nil {
		return err
	}
	
	// Time in seconds (20 minutes = 1200 seconds)
	if err := addWorkoutMovementWithTime(db, workoutID, runID, 1200, 0); err != nil {
		return err
	}
	if err := addWorkoutMovementWithTime(db, workoutID, rowID, 1200, 1); err != nil {
		return err
	}
	if err := addWorkoutMovementWithTime(db, workoutID, bikeID, 1200, 2); err != nil {
		return err
	}

	// Template 5: Fran (linking to WOD)
	workoutID, err = createWorkout(db, "Fran - Classic Girl WOD", "21-15-9 Thrusters and Pull-ups")
	if err != nil {
		return err
	}
	
	// Get Fran WOD ID
	franWODID, err := getWODIDByName(db, "Fran")
	if err != nil {
		return err
	}
	
	// Link WOD to workout
	if err := addWorkoutWOD(db, workoutID, franWODID, 0); err != nil {
		return err
	}
	
	// Add movements for Fran
	thrusterID, err := getMovementIDByName(db, "Thruster")
	if err != nil {
		return err
	}
	
	// Fran: 21-15-9 (we'll represent as 3 rounds, 7 reps average for simplicity)
	if err := addWorkoutMovement(db, workoutID, thrusterID, 95.0, 3, 15, 0); err != nil {
		return err
	}
	if err := addWorkoutMovement(db, workoutID, pullupID, 0, 3, 15, 1); err != nil {
		return err
	}

	// Template 6: Helen (linking to WOD)
	workoutID, err = createWorkout(db, "Helen - Classic Girl WOD", "3 rounds: 400m run, 21 KB swings, 12 pull-ups")
	if err != nil {
		return err
	}
	
	helenWODID, err := getWODIDByName(db, "Helen")
	if err != nil {
		return err
	}
	
	if err := addWorkoutWOD(db, workoutID, helenWODID, 0); err != nil {
		return err
	}
	
	kbSwingID, err := getMovementIDByName(db, "Kettlebell Swing")
	if err != nil {
		return err
	}
	
	// Helen movements
	if err := addWorkoutMovementWithDistance(db, workoutID, runID, 400.0, 3, 0); err != nil {
		return err
	}
	if err := addWorkoutMovement(db, workoutID, kbSwingID, 53.0, 3, 21, 1); err != nil {
		return err
	}
	if err := addWorkoutMovement(db, workoutID, pullupID, 0, 3, 12, 2); err != nil {
		return err
	}

	return nil
}

// Helper functions for workout template seeding

func createWorkout(db *sql.DB, name, notes string) (int64, error) {
	// Get database-specific timestamp function
	var timestampFunc string
	switch currentDriver {
	case "sqlite3":
		timestampFunc = "datetime('now')"
	case "postgres":
		timestampFunc = "CURRENT_TIMESTAMP"
	case "mysql":
		timestampFunc = "NOW()"
	default:
		timestampFunc = "CURRENT_TIMESTAMP"
	}

	// Get database-specific placeholders
	ph := getPlaceholders(currentDriver, 2) // name, notes

	// PostgreSQL uses RETURNING clause instead of LastInsertId
	if currentDriver == "postgres" {
		query := fmt.Sprintf(`INSERT INTO workouts (name, notes, created_by, created_at, updated_at)
		          VALUES (%s, %s, NULL, %s, %s) RETURNING id`, ph[0], ph[1], timestampFunc, timestampFunc)
		var id int64
		err := db.QueryRow(query, name, notes).Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("failed to create workout %s: %w", name, err)
		}
		return id, nil
	}

	// SQLite and MySQL use LastInsertId
	query := fmt.Sprintf(`INSERT INTO workouts (name, notes, created_by, created_at, updated_at)
	          VALUES (%s, %s, NULL, %s, %s)`, ph[0], ph[1], timestampFunc, timestampFunc)
	result, err := db.Exec(query, name, notes)
	if err != nil {
		return 0, fmt.Errorf("failed to create workout %s: %w", name, err)
	}
	return result.LastInsertId()
}

func addWorkoutMovement(db *sql.DB, workoutID, movementID int64, weight float64, sets, reps, orderIndex int) error {
	// Get database-specific timestamp and boolean functions
	var timestampFunc string
	switch currentDriver {
	case "sqlite3":
		timestampFunc = "datetime('now')"
	case "postgres":
		timestampFunc = "CURRENT_TIMESTAMP"
	case "mysql":
		timestampFunc = "NOW()"
	default:
		timestampFunc = "CURRENT_TIMESTAMP"
	}

	// Get database-specific placeholders and boolean values
	ph := getPlaceholders(currentDriver, 6) // workoutID, movementID, weight, sets, reps, orderIndex
	boolFalse := getBoolValue(currentDriver, false)
	query := fmt.Sprintf(`INSERT INTO workout_movements (workout_id, movement_id, weight, sets, reps, time, distance, is_rx, is_pr, order_index, created_at, updated_at)
	          VALUES (%s, %s, %s, %s, %s, NULL, NULL, %s, %s, %s, %s, %s)`,
		ph[0], ph[1], ph[2], ph[3], ph[4], boolFalse, boolFalse, ph[5], timestampFunc, timestampFunc)
	_, err := db.Exec(query, workoutID, movementID, weight, sets, reps, orderIndex)
	return err
}

func addWorkoutMovementWithTime(db *sql.DB, workoutID, movementID int64, timeSeconds, orderIndex int) error {
	// Get database-specific timestamp and boolean functions
	var timestampFunc string
	switch currentDriver {
	case "sqlite3":
		timestampFunc = "datetime('now')"
	case "postgres":
		timestampFunc = "CURRENT_TIMESTAMP"
	case "mysql":
		timestampFunc = "NOW()"
	default:
		timestampFunc = "CURRENT_TIMESTAMP"
	}

	// Get database-specific placeholders and boolean values
	ph := getPlaceholders(currentDriver, 4) // workoutID, movementID, timeSeconds, orderIndex
	boolFalse := getBoolValue(currentDriver, false)
	query := fmt.Sprintf(`INSERT INTO workout_movements (workout_id, movement_id, weight, sets, reps, time, distance, is_rx, is_pr, order_index, created_at, updated_at)
	          VALUES (%s, %s, NULL, NULL, NULL, %s, NULL, %s, %s, %s, %s, %s)`,
		ph[0], ph[1], ph[2], boolFalse, boolFalse, ph[3], timestampFunc, timestampFunc)
	_, err := db.Exec(query, workoutID, movementID, timeSeconds, orderIndex)
	return err
}

func addWorkoutMovementWithDistance(db *sql.DB, workoutID, movementID int64, distance float64, rounds, orderIndex int) error {
	// Get database-specific timestamp function
	var timestampFunc string
	switch currentDriver {
	case "sqlite3":
		timestampFunc = "datetime('now')"
	case "postgres":
		timestampFunc = "CURRENT_TIMESTAMP"
	case "mysql":
		timestampFunc = "NOW()"
	default:
		timestampFunc = "CURRENT_TIMESTAMP"
	}

	// Get database-specific placeholders and boolean values
	ph := getPlaceholders(currentDriver, 5) // workoutID, movementID, rounds, distance, orderIndex
	boolFalse := getBoolValue(currentDriver, false)
	query := fmt.Sprintf(`INSERT INTO workout_movements (workout_id, movement_id, weight, sets, reps, time, distance, is_rx, is_pr, order_index, created_at, updated_at)
	          VALUES (%s, %s, NULL, %s, NULL, NULL, %s, %s, %s, %s, %s, %s)`,
		ph[0], ph[1], ph[2], ph[3], boolFalse, boolFalse, ph[4], timestampFunc, timestampFunc)
	_, err := db.Exec(query, workoutID, movementID, rounds, distance, orderIndex)
	return err
}

func addWorkoutWOD(db *sql.DB, workoutID, wodID int64, orderIndex int) error {
	// Get database-specific timestamp function
	var timestampFunc string
	switch currentDriver {
	case "sqlite3":
		timestampFunc = "datetime('now')"
	case "postgres":
		timestampFunc = "CURRENT_TIMESTAMP"
	case "mysql":
		timestampFunc = "NOW()"
	default:
		timestampFunc = "CURRENT_TIMESTAMP"
	}

	// Get database-specific placeholders
	ph := getPlaceholders(currentDriver, 3) // workoutID, wodID, orderIndex
	query := fmt.Sprintf(`INSERT INTO workout_wods (workout_id, wod_id, order_index, created_at, updated_at)
	          VALUES (%s, %s, %s, %s, %s)`, ph[0], ph[1], ph[2], timestampFunc, timestampFunc)
	_, err := db.Exec(query, workoutID, wodID, orderIndex)
	return err
}

func getMovementIDByName(db *sql.DB, name string) (int64, error) {
	var id int64
	ph := getPlaceholders(currentDriver, 1)
	query := fmt.Sprintf("SELECT id FROM movements WHERE name = %s", ph[0])
	err := db.QueryRow(query, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to find movement %s: %w", name, err)
	}
	return id, nil
}

func getWODIDByName(db *sql.DB, name string) (int64, error) {
	var id int64
	ph := getPlaceholders(currentDriver, 1)
	query := fmt.Sprintf("SELECT id FROM wods WHERE name = %s", ph[0])
	err := db.QueryRow(query, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to find WOD %s: %w", name, err)
	}
	return id, nil
}
