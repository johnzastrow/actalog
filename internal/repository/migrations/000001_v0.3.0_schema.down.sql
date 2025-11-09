-- Rollback Migration: v0.3.0 to v0.2.0
-- This migration rolls back the v0.3.0 schema changes

-- ============================================================================
-- WARNING: This rollback will lose data for new v0.3.0 features:
-- - WODs
-- - User settings
-- - Audit logs
-- - PR flags
-- - Division information
-- ============================================================================

-- Drop new tables
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS workout_wods;
DROP TABLE IF EXISTS wods;
DROP TABLE IF EXISTS user_workouts;

-- Remove new columns from users
ALTER TABLE users DROP COLUMN IF EXISTS updated_by;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
ALTER TABLE users DROP COLUMN IF EXISTS birthday;

-- Remove new columns from workout_strength
-- Note: SQLite doesn't support DROP COLUMN before version 3.35.0
-- We'll need to recreate the table

-- Backup workout_strength
CREATE TABLE IF NOT EXISTS workout_strength_backup AS SELECT
    id, workout_id, movement_id, weight, sets, reps, time, distance, is_rx, notes, order_index, created_at, updated_at
FROM workout_strength;

DROP TABLE workout_strength;

-- Recreate workout_movements (old name) without is_pr
CREATE TABLE workout_movements (
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

INSERT INTO workout_movements SELECT
    id, workout_id, movement_id, weight, sets, reps, time, distance, is_rx, notes, order_index, created_at, updated_at
FROM workout_strength_backup;

DROP TABLE workout_strength_backup;

CREATE INDEX IF NOT EXISTS idx_wm_workout_id ON workout_movements(workout_id);
CREATE INDEX IF NOT EXISTS idx_wm_movement_id ON workout_movements(movement_id);
CREATE INDEX IF NOT EXISTS idx_wm_workout_order ON workout_movements(workout_id, order_index);

-- Remove updated_by from strength_movements and rename back to movements
CREATE TABLE movements_backup AS SELECT
    id, name, description, type, is_standard, created_by, created_at, updated_at
FROM strength_movements;

DROP TABLE strength_movements;

CREATE TABLE movements (
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

INSERT INTO movements SELECT * FROM movements_backup;

DROP TABLE movements_backup;

CREATE INDEX IF NOT EXISTS idx_movements_name ON movements(name);
CREATE INDEX IF NOT EXISTS idx_movements_type ON movements(type);
CREATE INDEX IF NOT EXISTS idx_movements_standard ON movements(is_standard);

-- Restore old workouts table from backup
DROP TABLE IF EXISTS workouts;

CREATE TABLE workouts (
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

-- Restore data from backup if it exists
INSERT OR IGNORE INTO workouts SELECT * FROM workouts_old WHERE EXISTS (SELECT 1 FROM workouts_old LIMIT 1);

CREATE INDEX IF NOT EXISTS idx_workouts_user_id ON workouts(user_id);
CREATE INDEX IF NOT EXISTS idx_workouts_workout_date ON workouts(workout_date);
CREATE INDEX IF NOT EXISTS idx_workouts_user_date ON workouts(user_id, workout_date DESC);

-- Drop backup table
DROP TABLE IF EXISTS workouts_old;

-- ============================================================================
-- Rollback complete - Schema reverted to v0.2.0
-- ============================================================================
