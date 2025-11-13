-- Rollback migration: v0.3.2
-- This removes support for ad-hoc workouts

-- ============================================================================
-- STEP 1: Create old user_workouts table (workout_id NOT NULL, no workout_name)
-- ============================================================================

CREATE TABLE user_workouts_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    workout_id INTEGER NOT NULL,
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
-- STEP 2: Copy data back (only template-based workouts, skip ad-hoc)
-- ============================================================================

INSERT INTO user_workouts_old (id, user_id, workout_id, workout_date, workout_type, total_time, notes, created_at, updated_at)
SELECT id, user_id, workout_id, workout_date, workout_type, total_time, notes, created_at, updated_at
FROM user_workouts
WHERE workout_id IS NOT NULL;

-- Note: Ad-hoc workouts (where workout_id IS NULL) are lost in rollback

-- ============================================================================
-- STEP 3: Drop new table and rename old table
-- ============================================================================

DROP TABLE user_workouts;
ALTER TABLE user_workouts_old RENAME TO user_workouts;

-- ============================================================================
-- STEP 4: Recreate indexes
-- ============================================================================

CREATE INDEX idx_user_workouts_user_id ON user_workouts(user_id);
CREATE INDEX idx_user_workouts_workout_date ON user_workouts(workout_date);
CREATE INDEX idx_user_workouts_user_date ON user_workouts(user_id, workout_date DESC);

-- ============================================================================
-- Rollback complete
-- ============================================================================
