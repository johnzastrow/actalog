package repository

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// SetupTestDB creates an in-memory SQLite database for testing
// Returns the database connection and a cleanup function
func SetupTestDB() (*sql.DB, func(), error) {
	// Use in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, nil, err
	}

	// Set the current driver for rebindQuery to work correctly
	currentDriver = "sqlite3"

	// Create all tables using the SQLite schema
	schema := getSQLiteSchema()
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, nil, err
	}

	// Add additional tables from migrations that tests depend on
	additionalSchema := getTestAdditionalSchema()
	if _, err := db.Exec(additionalSchema); err != nil {
		db.Close()
		return nil, nil, err
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup, nil
}

// getTestAdditionalSchema returns additional table schemas from migrations needed for testing
func getTestAdditionalSchema() string {
	return `
	-- Notifications table (from migration v0.12.0)
	CREATE TABLE IF NOT EXISTS notifications (
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
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
	CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, read_at);
	CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);

	-- Notification likes table (from migration v0.12.0)
	CREATE TABLE IF NOT EXISTS notification_likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		notification_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		UNIQUE(notification_id, user_id),
		FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_notification_likes_notification_id ON notification_likes(notification_id);
	CREATE INDEX IF NOT EXISTS idx_notification_likes_user_id ON notification_likes(user_id);

	-- Organizations table (from migration v0.5.0)
	CREATE TABLE IF NOT EXISTS organizations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		description TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_organizations_name ON organizations(name);

	-- User organizations junction table (from migration v0.5.1)
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
	CREATE INDEX IF NOT EXISTS idx_user_orgs_lookup ON user_organizations(user_id, organization_id);
	`
}
