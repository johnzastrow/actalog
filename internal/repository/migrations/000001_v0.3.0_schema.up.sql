-- Migration: v0.3.0 Schema Update
-- This migration transforms the database from v0.2.0 to v0.3.0
-- Key changes: Workouts become templates, add WODs, PR tracking, leaderboards

-- ============================================================================
-- STEP 1: Add new columns to existing tables
-- ============================================================================

-- Add new columns to users table
ALTER TABLE users ADD COLUMN birthday DATE;
ALTER TABLE users ADD COLUMN email_verified INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN email_verified_at DATETIME;
ALTER TABLE users ADD COLUMN updated_by INTEGER;

-- ============================================================================
-- STEP 2: Rename tables
-- ============================================================================

-- Rename movements to strength_movements
ALTER TABLE movements RENAME TO strength_movements;
-- Rename workout_movements to workout_strength
ALTER TABLE workout_movements RENAME TO workout_strength;

-- ============================================================================
-- STEP 3: Backup old workouts data before transformation
-- ============================================================================

-- Create backup table for old workouts structure
CREATE TABLE IF NOT EXISTS workouts_old AS SELECT * FROM workouts;

-- ============================================================================
-- STEP 4: Create new tables
-- ============================================================================

-- WODs table (predefined CrossFit workouts)
CREATE TABLE IF NOT EXISTS wods (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    source TEXT,  -- CrossFit, Other Coach, Self-recorded
    type TEXT,    -- Benchmark, Hero, Girl, Notables, Games, Endurance, Self-created
    regime TEXT,  -- EMOM, AMRAP, Fastest Time, Slowest Round, Get Stronger, Skills
    score_type TEXT,  -- Time, Rounds and Reps, Max Weight
    description TEXT,
    url TEXT,
    notes TEXT,
    is_standard INTEGER NOT NULL DEFAULT 0,
    created_by INTEGER,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    updated_by INTEGER,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_wods_name ON wods(name);
CREATE INDEX IF NOT EXISTS idx_wods_type ON wods(type);
CREATE INDEX IF NOT EXISTS idx_wods_source ON wods(source);
CREATE INDEX IF NOT EXISTS idx_wods_standard ON wods(is_standard);
CREATE INDEX IF NOT EXISTS idx_wods_created_by ON wods(created_by);

-- User Settings table
CREATE TABLE IF NOT EXISTS user_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    notification_preferences TEXT,  -- JSON format
    data_export_format TEXT DEFAULT 'JSON',
    theme TEXT DEFAULT 'light',
    weight_unit TEXT DEFAULT 'lbs',  -- lbs or kg
    distance_unit TEXT DEFAULT 'miles',  -- miles or km
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings(user_id);

-- Audit Logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    action TEXT NOT NULL,
    details TEXT,  -- JSON format
    timestamp DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);

-- ============================================================================
-- STEP 5: Transform workouts table to templates
-- ============================================================================

-- Drop the old workouts table
DROP TABLE workouts;

-- Create new workouts table (templates, not user-specific instances)
CREATE TABLE IF NOT EXISTS workouts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    notes TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    updated_by INTEGER,
    FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_workouts_name ON workouts(name);

-- ============================================================================
-- STEP 6: Create junction tables
-- ============================================================================

-- User Workouts junction (links users to workouts on specific dates)
CREATE TABLE IF NOT EXISTS user_workouts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    workout_id INTEGER NOT NULL,
    workout_date DATE NOT NULL,
    notes TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_workouts_user_id ON user_workouts(user_id);
CREATE INDEX IF NOT EXISTS idx_user_workouts_workout_id ON user_workouts(workout_id);
CREATE INDEX IF NOT EXISTS idx_user_workouts_date ON user_workouts(workout_date);
CREATE INDEX IF NOT EXISTS idx_user_workouts_user_date ON user_workouts(user_id, workout_date DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_workouts_unique ON user_workouts(user_id, workout_id, workout_date);

-- Workout WODs junction (links workouts to WODs with scoring)
CREATE TABLE IF NOT EXISTS workout_wods (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workout_id INTEGER NOT NULL,
    wod_id INTEGER NOT NULL,
    score_value TEXT,  -- Time (HH:MM:SS), Rounds+Reps (R:R), or Weight
    division TEXT,  -- rx, scaled, beginner
    is_pr INTEGER NOT NULL DEFAULT 0,
    order_index INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
    FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_workout_wods_workout_id ON workout_wods(workout_id);
CREATE INDEX IF NOT EXISTS idx_workout_wods_wod_id ON workout_wods(wod_id);
CREATE INDEX IF NOT EXISTS idx_workout_wods_order ON workout_wods(workout_id, order_index);
CREATE INDEX IF NOT EXISTS idx_workout_wods_pr ON workout_wods(is_pr);
CREATE INDEX IF NOT EXISTS idx_workout_wods_division ON workout_wods(division);

-- ============================================================================
-- STEP 7: Add new columns to workout_strength (formerly workout_movements)
-- ============================================================================

ALTER TABLE workout_strength ADD COLUMN is_pr INTEGER NOT NULL DEFAULT 0;

-- Add index for PR tracking
CREATE INDEX IF NOT EXISTS idx_workout_strength_pr ON workout_strength(is_pr);

-- ============================================================================
-- STEP 8: Add updated_by to strength_movements (formerly movements)
-- ============================================================================

ALTER TABLE strength_movements ADD COLUMN updated_by INTEGER;

-- ============================================================================
-- STEP 9: Migrate old workout data to new structure
-- ============================================================================

-- For each old workout, create:
-- 1. A new workout template
-- 2. A user_workout linking the user to that template on the workout_date
-- 3. Keep existing workout_strength records (they already reference workout_id)

INSERT INTO workouts (id, name, notes, created_at, updated_at, updated_by)
SELECT id, workout_name, notes, created_at, updated_at, NULL
FROM workouts_old;

INSERT INTO user_workouts (user_id, workout_id, workout_date, notes, created_at, updated_at)
SELECT user_id, id, workout_date, NULL, created_at, updated_at
FROM workouts_old;

-- Note: workout_strength table still has correct workout_id references
-- No need to update those

-- ============================================================================
-- STEP 10: Create default user settings for existing users
-- ============================================================================

INSERT INTO user_settings (user_id, notification_preferences, data_export_format, theme, weight_unit, distance_unit, created_at, updated_at)
SELECT id, NULL, 'JSON', 'light', 'lbs', 'miles', datetime('now'), datetime('now')
FROM users
WHERE id NOT IN (SELECT user_id FROM user_settings);

-- ============================================================================
-- Migration complete
-- ============================================================================
-- Schema version: 0.3.0
-- Old workouts backed up in workouts_old table (can be dropped after verification)
