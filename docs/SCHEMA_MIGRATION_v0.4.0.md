# Schema Migration v0.4.0 - Requirements Compliance

## Overview

This migration transforms the current schema (v0.3.3) to match the data model specified in REQUIREMENTS.md. This is a **major architectural change** that converts user-specific workout instances into reusable workout templates.

## Current State (v0.3.3)

### Current Tables
- **users** - User accounts
- **workouts** - User-specific workout instances (has user_id, workout_date) ❌
- **movements** - Exercise definitions (weightlifting, cardio, gymnastics only) ❌
- **workout_movements** - Links workouts to movements with performance data
- **refresh_tokens** - Remember Me functionality
- **password_resets** - Password reset tokens
- **email_verification_tokens** - Email verification

### Current Data Model
```
User → Workout (1:M direct relationship)
Workout → WorkoutMovement → Movement
```

**Problem**: Workouts are tied to users and dates, making them non-reusable.

## Target State (v0.4.0 - Requirements Compliant)

### New/Modified Tables

#### 1. **workouts** (MODIFIED - Now Templates)
Remove user-specific fields, make reusable:
```sql
-- REMOVE: user_id, workout_date, workout_type, workout_name, total_time
-- KEEP: id, notes, created_at, updated_at
-- ADD: name (workout template name), created_by (user_id FK, nullable)
```

| Column | Type | Description |
|--------|------|-------------|
| id | BIGINT PK | Unique workout template ID |
| name | VARCHAR(255) | Template name (e.g., "Monday Strength") |
| notes | TEXT | General template notes |
| created_by | BIGINT FK | User who created (NULL for standard) |
| created_at | TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | Last update time |

**Design**: Workouts become reusable templates that can be used by multiple users multiple times.

#### 2. **wods** (NEW TABLE)
Separate table for CrossFit WOD definitions:
```sql
CREATE TABLE wods (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(255) UNIQUE NOT NULL,
  source VARCHAR(100),  -- CrossFit, Other Coach, Self-recorded
  type VARCHAR(50),      -- Benchmark, Hero, Girl, Notables, Games, Endurance, Self-created
  regime VARCHAR(50),    -- EMOM, AMRAP, Fastest Time, Slowest Round, Get Stronger, Skills
  score_type VARCHAR(50), -- Time, Rounds+Reps, Max Weight
  description TEXT,
  url VARCHAR(512),
  notes TEXT,
  is_standard BOOLEAN NOT NULL DEFAULT FALSE,
  created_by BIGINT FK,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);
```

#### 3. **strength_movements** (RENAME FROM movements)
Rename `movements` → `strength_movements` (requires data migration):
```sql
-- Keep all current fields, just rename table
ALTER TABLE movements RENAME TO strength_movements;
```

#### 4. **user_workouts** (NEW TABLE - Junction)
Links users to workouts they've logged on specific dates:
```sql
CREATE TABLE user_workouts (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL FK → users(id),
  workout_id BIGINT NOT NULL FK → workouts(id),
  workout_date DATE NOT NULL,
  notes TEXT,  -- User's notes for this specific instance
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  UNIQUE KEY (user_id, workout_id, workout_date)
);
```

#### 5. **workout_wods** (NEW TABLE - Junction)
Links workouts to WODs they contain:
```sql
CREATE TABLE workout_wods (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  workout_id BIGINT NOT NULL FK → workouts(id),
  wod_id BIGINT NOT NULL FK → wods(id),
  score_value VARCHAR(50),  -- Actual score when logged (time, rounds+reps, weight)
  division VARCHAR(20),     -- rx, scaled, beginner
  is_pr BOOLEAN NOT NULL DEFAULT FALSE,
  order_index INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);
```

#### 6. **workout_strength** (RENAME FROM workout_movements)
Rename `workout_movements` → `workout_strength`:
```sql
ALTER TABLE workout_movements RENAME TO workout_strength;
-- Update FK reference: movement_id → strength_id
ALTER TABLE workout_strength CHANGE movement_id strength_id BIGINT NOT NULL;
```

### New Data Model
```
User → UserWorkout ← Workout → WorkoutWOD → WOD
                         ↓
                  WorkoutStrength → StrengthMovement
```

## Migration Steps

### Phase 1: Data Preservation (Before Schema Changes)

**Step 1.1**: Export existing data for backup
```sql
-- Backup current workouts with user associations
CREATE TABLE workouts_backup AS SELECT * FROM workouts;
CREATE TABLE workout_movements_backup AS SELECT * FROM workout_movements;
```

**Step 1.2**: Analyze current data patterns
```sql
-- Count unique workout patterns (same movements, weights, structure)
-- Determine which workouts can be consolidated into templates
SELECT workout_type, workout_name, COUNT(*) as usage_count
FROM workouts
GROUP BY workout_type, workout_name
ORDER BY usage_count DESC;
```

### Phase 2: Create New Tables

**Step 2.1**: Create `wods` table
```sql
-- Schema defined above
```

**Step 2.2**: Create `user_workouts` junction table
```sql
-- Schema defined above
```

**Step 2.3**: Create `workout_wods` junction table
```sql
-- Schema defined above
```

### Phase 3: Rename Existing Tables

**Step 3.1**: Rename `movements` → `strength_movements`
```sql
ALTER TABLE movements RENAME TO strength_movements;
-- Update indexes
```

**Step 3.2**: Rename `workout_movements` → `workout_strength`
```sql
ALTER TABLE workout_movements RENAME TO workout_strength;
ALTER TABLE workout_strength CHANGE COLUMN movement_id strength_id BIGINT NOT NULL;
-- Update foreign keys and indexes
```

### Phase 4: Refactor `workouts` Table

**Step 4.1**: Add new columns to workouts
```sql
ALTER TABLE workouts ADD COLUMN name VARCHAR(255);
ALTER TABLE workouts ADD COLUMN created_by BIGINT;
```

**Step 4.2**: Migrate data - Create templates from existing workouts

**Strategy**: Convert each user's unique workout pattern into a template
```sql
-- For each existing workout:
-- 1. Create a workout template (if pattern doesn't already exist)
-- 2. Create user_workout entry linking user to template with date
-- 3. Update workout_strength references

-- Pseudo-algorithm:
FOR EACH old_workout IN workouts_backup:
  -- Find or create template
  template_id = FIND_OR_CREATE_TEMPLATE(
    name = old_workout.workout_name OR "Custom Workout " + old_workout.id,
    notes = old_workout.notes
  )

  -- Create user_workout link
  INSERT INTO user_workouts (user_id, workout_id, workout_date, notes)
  VALUES (old_workout.user_id, template_id, old_workout.workout_date, old_workout.notes)

  -- workout_strength rows already reference workout_id (now template_id)
  -- No changes needed if workout_id becomes template_id
END FOR
```

**Step 4.3**: Remove old columns from workouts
```sql
ALTER TABLE workouts DROP COLUMN user_id;
ALTER TABLE workouts DROP COLUMN workout_date;
ALTER TABLE workouts DROP COLUMN workout_type;
ALTER TABLE workouts DROP COLUMN total_time;
-- workout_name becomes just 'name'
```

### Phase 5: Seed Standard WODs

**Step 5.1**: Populate `wods` table with standard CrossFit benchmarks
```go
// Example WODs to seed:
- Fran (Girl)
- Grace (Girl)
- Helen (Girl)
- Diane (Girl)
- Karen (Girl)
- Murph (Hero)
- DT (Hero)
- Cindy (Girl - AMRAP)
- Annie (Girl)
// ... etc.
```

### Phase 6: Update Foreign Key Constraints

**Step 6.1**: Update all FK references
```sql
-- workout_strength.strength_id → strength_movements(id)
-- workout_wods.wod_id → wods(id)
-- workout_wods.workout_id → workouts(id)
-- user_workouts.workout_id → workouts(id)
-- user_workouts.user_id → users(id)
```

## Data Migration Complexity

### Challenges

1. **Workout Deduplication**: Multiple users may have logged similar workouts. Need to identify unique patterns and create shared templates.

2. **Preserving History**: User's historical workout data (dates, notes, PRs) must be preserved in `user_workouts`.

3. **Workout Movements Linkage**: Current `workout_movements` directly links to user-specific workouts. After migration, these must link to templates, with user-specific data (dates, user) in `user_workouts`.

4. **Loss of workout_type and total_time**: Current schema has these fields. Need to decide:
   - Store in user_workouts.notes?
   - Recreate from workout_strength data?
   - Accept data loss?

### Migration Script Outline

```go
func MigrateToV040(db *sql.DB) error {
  tx, _ := db.Begin()

  // 1. Create new tables (wods, user_workouts, workout_wods)
  createNewTables(tx)

  // 2. Rename tables (movements→strength_movements, workout_movements→workout_strength)
  renameTables(tx)

  // 3. Backup existing workouts
  backupExistingData(tx)

  // 4. For each user's workout:
  rows, _ := tx.Query("SELECT * FROM workouts_backup")
  for rows.Next() {
    var oldWorkout OldWorkout
    rows.Scan(&oldWorkout...)

    // Find or create template
    templateID := findOrCreateTemplate(tx, oldWorkout)

    // Create user_workout entry
    createUserWorkout(tx, oldWorkout.UserID, templateID, oldWorkout.WorkoutDate, oldWorkout.Notes)
  }

  // 5. Remove old columns from workouts
  refactorWorkoutsTable(tx)

  // 6. Seed standard WODs
  seedStandardWODs(tx)

  tx.Commit()
  return nil
}
```

## Rollback Strategy

**If migration fails:**
1. Drop new tables (wods, user_workouts, workout_wods)
2. Rename tables back (strength_movements→movements, workout_strength→workout_movements)
3. Restore from backups (workouts_backup, workout_movements_backup)
4. Drop migration v0.4.0 from schema_migrations

**Point of No Return**: Once old columns are dropped from `workouts` table.

## Impact on Application Code

### Backend Changes Required

1. **Domain Models** (`internal/domain/`):
   - Add `WOD` struct
   - Add `UserWorkout` struct
   - Rename `Movement` → `StrengthMovement`
   - Rename `WorkoutMovement` → `WorkoutStrength`
   - Update `Workout` struct (remove user_id, workout_date)

2. **Repositories** (`internal/repository/`):
   - Create `WODRepository`
   - Create `UserWorkoutRepository`
   - Rename `MovementRepository` → `StrengthMovementRepository`
   - Update `WorkoutRepository` (now returns templates)
   - Update `WorkoutMovementRepository` → `WorkoutStrengthRepository`

3. **Services** (`internal/service/`):
   - Create `WODService`
   - Update `WorkoutService` (template CRUD + logging via user_workouts)
   - Update all references to movements/workout_movements

4. **Handlers** (`internal/handler/`):
   - Create `WODHandler`
   - Update `WorkoutHandler` (separate template management from logging)
   - Update all API endpoints

### Frontend Changes Required

1. **API Calls** (`web/src/utils/axios.js`, stores):
   - Update all workout-related API calls
   - Add WOD-related API calls
   - Update data structures

2. **Views**:
   - **LogWorkoutView**: Select template + log instance
   - **WorkoutsView**: Show user's logged workouts (via user_workouts)
   - Add **WODLibraryView** (browse/search WODs)
   - Add **WorkoutTemplatesView** (manage templates)
   - Update **PerformanceView** (query user_workouts, not workouts directly)

3. **Stores** (`web/src/stores/`):
   - Add `wods` store
   - Update `workouts` store (templates vs. instances)

## Testing Strategy

### Unit Tests
- Test each repository's CRUD operations with new schema
- Test foreign key constraints
- Test cascade deletes

### Integration Tests
- Test full workout logging flow (select template → log → retrieve)
- Test WOD creation and usage
- Test user_workout queries (date filtering, user-specific data)

### Migration Tests
- Test migration with sample data
- Test rollback
- Verify data integrity after migration
- Performance test with large datasets

## Timeline Estimate

- **Phase 1 (Planning & Design)**: 2 hours
- **Phase 2 (Backend Schema Migration)**: 4 hours
- **Phase 3 (Domain Models & Repositories)**: 4 hours
- **Phase 4 (Services & Handlers)**: 3 hours
- **Phase 5 (Frontend Updates)**: 4 hours
- **Phase 6 (Testing & Validation)**: 3 hours
- **Total**: ~20 hours

## Risk Assessment

**High Risk**:
- Data loss during migration if not properly backed up
- Downtime during migration (requires maintenance window)
- Breaking existing user workflows

**Medium Risk**:
- Foreign key constraint violations
- Performance degradation with new junction tables
- Frontend/backend sync issues during rollout

**Mitigation**:
- Comprehensive backups before migration
- Extensive testing in development environment
- Rollback plan documented and tested
- Gradual feature rollout (v0.4.0-alpha, beta, stable)

## Success Criteria

✓ All existing user workout data preserved
✓ Users can create and reuse workout templates
✓ WODs can be selected from library
✓ PR tracking still works
✓ Performance is equivalent or better
✓ All tests pass
✓ Documentation updated

## Open Questions

1. **Workout Deduplication Logic**: How aggressive should template consolidation be?
   - Conservative: Each user's workout becomes a unique template
   - Aggressive: Identical workout patterns share templates across users

2. **workout_type Field**: How to preserve this data?
   - Option A: Store in user_workouts.notes as JSON
   - Option B: Add workout_type to user_workouts table
   - Option C: Derive from workout contents

3. **total_time Field**: Keep or remove?
   - If keep, where? (user_workouts or workouts template?)

4. **Backward Compatibility**: Support old API endpoints during transition?
   - Dual API during migration period?
   - Immediate cutover?
