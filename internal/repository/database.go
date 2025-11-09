package repository

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// currentDriver stores the database driver being used
var currentDriver string

// BuildDSN constructs a database connection string based on the driver type
func BuildDSN(driver, host string, port int, user, password, database, sslMode string) string {
	switch driver {
	case "sqlite3":
		// For SQLite, database is the file path
		return database

	case "postgres":
		// PostgreSQL connection string
		dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
			host, port, user, database, sslMode)
		if password != "" {
			dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
				host, port, user, password, database, sslMode)
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
func InitDatabase(driver, dsn string) (*sql.DB, error) {
	// Store driver for later use
	currentDriver = driver

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
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
		role TEXT NOT NULL DEFAULT 'user',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		last_login_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

	CREATE TABLE IF NOT EXISTS workouts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		workout_date DATE NOT NULL,
		workout_type TEXT NOT NULL,
		workout_name TEXT,
		notes TEXT,
		total_time INTEGER,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_workouts_user_id ON workouts(user_id);
	CREATE INDEX IF NOT EXISTS idx_workouts_workout_date ON workouts(workout_date);
	CREATE INDEX IF NOT EXISTS idx_workouts_user_date ON workouts(user_id, workout_date DESC);

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
		role VARCHAR(50) NOT NULL DEFAULT 'user',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_login_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

	CREATE TABLE IF NOT EXISTS workouts (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		workout_date DATE NOT NULL,
		workout_type VARCHAR(50) NOT NULL,
		workout_name VARCHAR(255),
		notes TEXT,
		total_time INTEGER,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_workouts_user_id ON workouts(user_id);
	CREATE INDEX IF NOT EXISTS idx_workouts_workout_date ON workouts(workout_date);
	CREATE INDEX IF NOT EXISTS idx_workouts_user_date ON workouts(user_id, workout_date DESC);

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
		role VARCHAR(50) NOT NULL DEFAULT 'user',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		last_login_at DATETIME,
		INDEX idx_users_email (email),
		INDEX idx_users_role (role)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS workouts (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT NOT NULL,
		workout_date DATE NOT NULL,
		workout_type VARCHAR(50) NOT NULL,
		workout_name VARCHAR(255),
		notes TEXT,
		total_time INTEGER,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		INDEX idx_workouts_user_id (user_id),
		INDEX idx_workouts_workout_date (workout_date),
		INDEX idx_workouts_user_date (user_id, workout_date DESC)
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
