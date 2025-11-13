-- Migration: v0.3.2 Add support for ad-hoc workouts (Quick Log)
-- This migration allows users to log workouts without creating a template
-- Key changes: Make workout_id nullable, add workout_name field

-- ============================================================================
-- STEP 1: Create new user_workouts table with nullable workout_id
-- ============================================================================

-- SQLite doesn't support ALTER COLUMN, so we need to recreate the table
-- First, create new table with updated schema
CREATE TABLE user_workouts_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    workout_id INTEGER,  -- Now nullable to support ad-hoc workouts
    workout_name TEXT,   -- Name for ad-hoc workouts (when workout_id is NULL)
    workout_date DATE NOT NULL,
    workout_type TEXT,
    total_time INTEGER,
    notes TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE RESTRICT
);

-- ============================================================================
-- STEP 2: Copy existing data to new table
-- ============================================================================

INSERT INTO user_workouts_new (id, user_id, workout_id, workout_name, workout_date, workout_type, total_time, notes, created_at, updated_at)
SELECT id, user_id, workout_id, NULL, workout_date, workout_type, total_time, notes, created_at, updated_at
FROM user_workouts;

-- ============================================================================
-- STEP 3: Drop old table and rename new table
-- ============================================================================

DROP TABLE user_workouts;
ALTER TABLE user_workouts_new RENAME TO user_workouts;

-- ============================================================================
-- STEP 4: Recreate indexes
-- ============================================================================

CREATE INDEX idx_user_workouts_user_id ON user_workouts(user_id);
CREATE INDEX idx_user_workouts_workout_id ON user_workouts(workout_id);
CREATE INDEX idx_user_workouts_workout_date ON user_workouts(workout_date);
CREATE INDEX idx_user_workouts_user_date ON user_workouts(user_id, workout_date DESC);

-- Add constraint: either workout_id or workout_name must be set
-- This is enforced at application level, not database level for SQLite

-- ============================================================================
-- Migration complete
-- ============================================================================
-- Schema version: 0.3.2
-- Ad-hoc workouts now supported via workout_name field
