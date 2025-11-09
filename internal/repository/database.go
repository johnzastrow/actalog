package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// InitDatabase initializes the database connection and runs migrations
func InitDatabase(driver, dsn string) (*sql.DB, error) {
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
	if err := createInitialTablesIfNotExist(db); err != nil {
		return nil, fmt.Errorf("failed to create initial tables: %w", err)
	}

	// Run migrations to bring schema up to latest version
	fmt.Println("Running database migrations...")
	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Seed standard movements (if not already seeded)
	if err := seedStandardMovements(db); err != nil {
		return nil, fmt.Errorf("failed to seed standard movements: %w", err)
	}

	return db, nil
}

// createInitialTablesIfNotExist creates the initial v0.1.0 schema if tables don't exist
func createInitialTablesIfNotExist(db *sql.DB) error {
	// Check if users table exists
	var tableName string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableName)
	if err == nil {
		// Tables already exist, skip initialization
		return nil
	}

	fmt.Println("Initializing new database with v0.1.0 schema...")
	return createTables(db)
}

// createTables creates all necessary database tables
func createTables(db *sql.DB) error {
	schema := `
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

	_, err := db.Exec(schema)
	return err
}

// seedStandardMovements seeds the database with standard CrossFit movements
func seedStandardMovements(db *sql.DB) error {
	// Check if movements already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM movements WHERE is_standard = 1").Scan(&count)
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
	}

	// Prepare insert statement
	stmt := `
		INSERT INTO movements (name, description, type, is_standard, created_by, created_at, updated_at)
		VALUES (?, ?, ?, 1, NULL, datetime('now'), datetime('now'))
	`

	// Insert each movement
	for _, m := range movements {
		_, err := db.Exec(stmt, m.name, m.description, m.movType)
		if err != nil {
			return fmt.Errorf("failed to seed movement %s: %w", m.name, err)
		}
	}

	return nil
}
