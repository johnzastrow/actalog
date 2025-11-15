-- Migration Rollback: v0.4.6 Security Enhancements
-- This migration rolls back security enhancements added in v0.4.6

-- ============================================================================
-- STEP 1: Restore original audit_logs table structure
-- ============================================================================

-- Backup current enhanced audit_logs
CREATE TABLE IF NOT EXISTS audit_logs_enhanced_backup AS SELECT * FROM audit_logs;

-- Drop enhanced audit_logs table
DROP TABLE audit_logs;

-- Restore original v0.3.0 audit_logs structure
CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    action TEXT NOT NULL,
    details TEXT,
    timestamp DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);

-- Migrate enhanced audit logs back to old structure
-- New schema: user_id, target_user_id, event_type, ip_address, user_agent, details, created_at
-- Old schema: user_id, action, details, timestamp
INSERT INTO audit_logs (user_id, action, details, timestamp)
SELECT user_id, event_type, details, created_at
FROM audit_logs_enhanced_backup;

-- ============================================================================
-- STEP 2: Remove security columns from users table
-- ============================================================================

-- Note: SQLite doesn't support DROP COLUMN directly
-- We need to recreate the table without the new columns

-- Create new users table without security columns
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
    updated_by INTEGER,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    last_login_at DATETIME
);

-- Copy data from old users table (excluding new security columns)
INSERT INTO users_new (id, email, password_hash, name, profile_image, birthday, role, email_verified, email_verified_at, updated_by, created_at, updated_at, last_login_at)
SELECT id, email, password_hash, name, profile_image, birthday, role, email_verified, email_verified_at, updated_by, created_at, updated_at, last_login_at
FROM users;

-- Drop old users table
DROP TABLE users;

-- Rename new users table to users
ALTER TABLE users_new RENAME TO users;

-- Recreate indexes on users table
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- ============================================================================
-- Migration rollback complete
-- ============================================================================
-- Schema version: 0.3.2 (reverted from 0.4.6)
-- Removed: Account lockout and disable columns
-- Reverted: Audit logging to original structure
