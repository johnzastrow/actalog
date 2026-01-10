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

	-- Data change logs table (from migration v0.11.1)
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
	CREATE INDEX IF NOT EXISTS idx_data_change_logs_operation ON data_change_logs(operation);

	-- User subscriptions table (from migration v0.14.0)
	CREATE TABLE IF NOT EXISTS user_subscriptions (
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
	CREATE INDEX IF NOT EXISTS idx_user_subscriptions_user_id ON user_subscriptions(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_subscriptions_status ON user_subscriptions(status);
	CREATE INDEX IF NOT EXISTS idx_user_subscriptions_next_billing ON user_subscriptions(next_billing_date);

	-- Organization subscriptions table (from migration v0.14.0)
	CREATE TABLE IF NOT EXISTS organization_subscriptions (
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
	CREATE INDEX IF NOT EXISTS idx_org_subscriptions_org_id ON organization_subscriptions(organization_id);
	CREATE INDEX IF NOT EXISTS idx_org_subscriptions_status ON organization_subscriptions(status);
	CREATE INDEX IF NOT EXISTS idx_org_subscriptions_next_billing ON organization_subscriptions(next_billing_date);

	-- Email logs table (from migration v0.18.0)
	CREATE TABLE IF NOT EXISTS email_logs (
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
	CREATE INDEX IF NOT EXISTS idx_email_logs_type ON email_logs(email_type);
	CREATE INDEX IF NOT EXISTS idx_email_logs_recipient ON email_logs(recipient_email);
	CREATE INDEX IF NOT EXISTS idx_email_logs_created_at ON email_logs(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_email_logs_success ON email_logs(success);

	-- Benchmark data table (from migration v0.22.0)
	CREATE TABLE IF NOT EXISTS benchmark_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		test_key TEXT NOT NULL,
		test_value TEXT,
		num_value REAL DEFAULT 0,
		int_value INTEGER DEFAULT 0,
		bool_value INTEGER DEFAULT 0,
		large_text TEXT,
		json_blob TEXT,
		extra_float REAL DEFAULT 0,
		extra_int INTEGER DEFAULT 0,
		category TEXT,
		priority INTEGER DEFAULT 0,
		created_by INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_benchmark_data_test_key ON benchmark_data(test_key);
	CREATE INDEX IF NOT EXISTS idx_benchmark_data_created_by ON benchmark_data(created_by);
	`
}
