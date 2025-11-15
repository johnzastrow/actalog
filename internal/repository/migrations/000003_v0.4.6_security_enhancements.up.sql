-- Migration: v0.4.6 Security Enhancements
-- This migration adds security features: account lockout, manual disable, enhanced audit logging

-- ============================================================================
-- STEP 1: Add security columns to users table
-- ============================================================================

-- Account lockout fields (automatic, time-based)
ALTER TABLE users ADD COLUMN failed_login_attempts INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN locked_at DATETIME;
ALTER TABLE users ADD COLUMN locked_until DATETIME;

-- Manual account disable fields (admin-controlled, permanent until re-enabled)
ALTER TABLE users ADD COLUMN account_disabled INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN disabled_at DATETIME;
ALTER TABLE users ADD COLUMN disabled_by_user_id INTEGER;

-- Add foreign key for disabled_by_user_id (admin who disabled the account)
-- Note: SQLite doesn't support adding FK constraints to existing tables,
-- so this will be enforced in application logic

-- ============================================================================
-- STEP 2: Enhance audit_logs table
-- ============================================================================

-- The audit_logs table already exists from v0.3.0 migration, but we need to enhance it
-- We'll rename the old table and create a new one with enhanced fields

-- Backup existing audit logs
CREATE TABLE IF NOT EXISTS audit_logs_backup AS SELECT * FROM audit_logs;

-- Drop old audit_logs table
DROP TABLE audit_logs;

-- Create enhanced audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,                    -- user who performed the action (NULL for system actions)
    target_user_id INTEGER,             -- user affected by the action (e.g., admin unlocking another user)
    event_type TEXT NOT NULL,           -- event type (login_success, login_failed, account_locked, etc.)
    ip_address TEXT,                    -- IP address of the request
    user_agent TEXT,                    -- User-Agent header from request
    details TEXT,                       -- JSON object with additional details
    created_at DATETIME NOT NULL,       -- when the event occurred

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_target_user_id ON audit_logs(target_user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- Migrate old audit logs to new structure (map old fields to new)
-- Old schema: user_id, action, details, timestamp
-- New schema: user_id, target_user_id, event_type, ip_address, user_agent, details, created_at
INSERT INTO audit_logs (user_id, target_user_id, event_type, ip_address, user_agent, details, created_at)
SELECT user_id, NULL, action, NULL, NULL, details, timestamp
FROM audit_logs_backup;

-- Keep backup table for safety (can be dropped later if migration is successful)
-- DROP TABLE audit_logs_backup;

-- ============================================================================
-- Migration complete
-- ============================================================================
-- Schema version: 0.4.6
-- Added: Account lockout and disable functionality
-- Enhanced: Audit logging with more detailed tracking
