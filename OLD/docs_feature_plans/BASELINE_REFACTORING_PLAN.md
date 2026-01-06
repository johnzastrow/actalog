# Baseline Refactoring Plan - Requirements Alignment

**Date:** 2025-11-12
**Status:** 🚧 IN PROGRESS
**Goal:** Rebuild ActaLog to match REQUIREMENTS.md specification

## Executive Summary

This document tracks the refactoring effort to align the ActaLog codebase with the original requirements specification. The work involves changing from an instance-based workout model to a template-based model as specified in requirements.

## Current State Analysis

### ✅ What's Already Done

**Domain Models (Complete):**
- ✅ `internal/domain/workout.go` - Workout template model
- ✅ `internal/domain/wod.go` - WOD entity model
- ✅ `internal/domain/workout_wod.go` - WorkoutWOD junction model
- ✅ `internal/domain/user_workout.go` - UserWorkout junction model
- ✅ `internal/domain/movement.go` - Movement model (already correct)
- ✅ `internal/domain/user.go` - User model (already correct)

**Database Backup:**
- ✅ Old database backed up to `actalog.db.backup-YYYYMMDD-HHMMSS`

### ❌ What Needs to Be Built/Changed

**Database Schema (`internal/repository/database.go`):**
1. ❌ `workouts` table - Remove user-specific fields (user_id, workout_date, workout_type, total_time)
2. ❌ `wods` table - Add new table for WOD definitions
3. ❌ `workout_wods` junction table - Link workouts to WODs
4. ❌ `user_workouts` junction table - Link users to workout instances

**Repositories (`internal/repository/`):**
1. ❌ Update `workout_repository.go` - Change to work with templates (no user_id in workouts table)
2. ❌ Create `wod_repository.go` - New repository for WOD operations
3. ❌ Create `workout_wod_repository.go` - New repository for WorkoutWOD operations
4. ❌ Create `user_workout_repository.go` - New repository for UserWorkout operations

**Services (`internal/service/`):**
1. ❌ Update `workout_service.go` - Template-based operations
2. ❌ Create `wod_service.go` - WOD business logic
3. ❌ Create `user_workout_service.go` - Workout logging business logic

**Handlers (`internal/handler/`):**
1. ❌ Update `workout_handler.go` - Template CRUD endpoints
2. ❌ Create `wod_handler.go` - WOD CRUD endpoints
3. ❌ Create `user_workout_handler.go` - Workout logging endpoints

**Main (`cmd/actalog/main.go`):**
1. ❌ Wire up new repositories
2. ❌ Wire up new services
3. ❌ Wire up new handlers
4. ❌ Update route definitions

**Frontend (`web/src/`):**
1. ❌ Update views to work with template-based model
2. ❌ Separate "browse templates" from "log workout" workflows
3. ❌ Add WOD browsing and selection
4. ❌ Update workout logging to reference templates

## Database Schema Changes

### OLD Schema (Instance-Based)

```sql
-- Old workouts table (user-specific instances)
CREATE TABLE workouts (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,          -- ❌ Remove
    workout_date DATE NOT NULL,         -- ❌ Move to user_workouts
    workout_type TEXT NOT NULL,         -- ❌ Move to user_workouts
    workout_name TEXT,                  -- Keep as "name"
    notes TEXT,                         -- Keep
    total_time INTEGER,                 -- ❌ Move to user_workouts
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### NEW Schema (Template-Based)

```sql
-- 1. Workouts table (templates, user-independent)
CREATE TABLE workouts (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,                 -- Template name
    notes TEXT,                         -- Template description
    created_by INTEGER,                 -- NULL for standard templates
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
);

-- 2. WODs table (benchmark WOD definitions)
CREATE TABLE wods (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,          -- "Fran", "Murph", etc.
    source TEXT,                        -- "CrossFit", "Other Coach", "Self-recorded"
    type TEXT,                          -- "Benchmark", "Hero", "Girl", etc.
    regime TEXT,                        -- "EMOM", "AMRAP", "Fastest Time", etc.
    score_type TEXT,                    -- "Time", "Rounds+Reps", "Max Weight"
    description TEXT,                   -- Full WOD description
    url TEXT,                           -- Video/reference URL
    notes TEXT,
    is_standard INTEGER NOT NULL DEFAULT 0,  -- Pre-seeded vs user-created
    created_by INTEGER,                      -- NULL for standard WODs
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
);

-- 3. Workout-WOD junction (workouts can include multiple WODs)
CREATE TABLE workout_wods (
    id INTEGER PRIMARY KEY,
    workout_id INTEGER NOT NULL,
    wod_id INTEGER NOT NULL,
    order_index INTEGER NOT NULL DEFAULT 0,  -- Order in workout
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
    FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT
);

-- 4. User-Workout junction (users log workout instances)
CREATE TABLE user_workouts (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    workout_id INTEGER NOT NULL,        -- References workout template
    workout_date DATE NOT NULL,          -- When it was performed
    workout_type TEXT,                   -- strength, metcon, cardio, mixed
    total_time INTEGER,                  -- Duration in seconds
    notes TEXT,                          -- User's notes for this instance
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE RESTRICT
);

-- Indexes for user_workouts
CREATE INDEX idx_user_workouts_user_id ON user_workouts(user_id);
CREATE INDEX idx_user_workouts_workout_date ON user_workouts(workout_date);
CREATE INDEX idx_user_workouts_user_date ON user_workouts(user_id, workout_date DESC);
```

## API Endpoint Changes

### Current (Broken) Endpoints

These currently exist but work with the wrong data model:
- `POST /api/workouts` - Creates workout instance (should create template)
- `GET /api/workouts` - Lists user's workouts (should list templates)

### NEW Endpoint Structure

**Workout Templates:**
- `POST /api/templates` - Create workout template
- `GET /api/templates` - List all templates (standard + user's custom)
- `GET /api/templates/{id}` - Get template details (with movements, WODs)
- `PUT /api/templates/{id}` - Update template (user's only)
- `DELETE /api/templates/{id}` - Delete template (user's only)
- `GET /api/templates/my` - List user's custom templates
- `GET /api/templates/standard` - List standard templates

**WODs:**
- `POST /api/wods` - Create custom WOD
- `GET /api/wods` - List all WODs (standard + user's custom)
- `GET /api/wods/{id}` - Get WOD details
- `PUT /api/wods/{id}` - Update WOD (user's only)
- `DELETE /api/wods/{id}` - Delete WOD (user's only)
- `GET /api/wods/search?q=fran` - Search WODs
- `GET /api/wods/standard` - List standard WODs (Fran, Murph, etc.)

**User Workouts (Logging):**
- `POST /api/workouts` - Log a workout (references template)
- `GET /api/workouts` - List user's logged workouts
- `GET /api/workouts/{id}` - Get logged workout details
- `PUT /api/workouts/{id}` - Update logged workout
- `DELETE /api/workouts/{id}` - Delete logged workout
- `GET /api/workouts/date/{YYYY-MM-DD}` - Get workouts for specific date

**Statistics/PRs (Keep as-is):**
- `GET /api/prs` - Get personal records
- `GET /api/pr-movements` - Get recent PR movements

## Implementation Order

### Phase 1: Database Schema ✅
1. ✅ Backup existing database
2. ❌ Update `database.go` with new schema
3. ❌ Delete old database file
4. ❌ Test app starts and creates new schema

### Phase 2: Repositories
1. ❌ Create `wod_repository.go`
2. ❌ Create `workout_wod_repository.go`
3. ❌ Create `user_workout_repository.go`
4. ❌ Update `workout_repository.go` (remove user-specific logic)

### Phase 3: Services
1. ❌ Create `wod_service.go`
2. ❌ Create `user_workout_service.go`
3. ❌ Update `workout_service.go` (template logic)

### Phase 4: Handlers
1. ❌ Create `wod_handler.go`
2. ❌ Create `user_workout_handler.go`
3. ❌ Update `workout_handler.go` (template endpoints)

### Phase 5: Main Wiring
1. ❌ Update `main.go` to wire up new components
2. ❌ Update route definitions
3. ❌ Test backend API

### Phase 6: Frontend Updates
1. ❌ Update `LogWorkoutView.vue` to reference templates
2. ❌ Update `WorkoutsView.vue` to show templates
3. ❌ Create `WODsView.vue` for browsing benchmark WODs
4. ❌ Update `DashboardView.vue` to show logged workouts
5. ❌ Update router and navigation

### Phase 7: Seed Data
1. ❌ Create seed data for standard workout templates
2. ❌ Create seed data for standard WODs (Fran, Murph, Cindy, etc.)

### Phase 8: Testing & Documentation
1. ❌ Test full workflow (create template → log workout → view history)
2. ❌ Update DATABASE_SCHEMA.md
3. ❌ Update API documentation
4. ❌ Update CHANGELOG.md

## Breaking Changes

This is a **complete data model change**. The following are incompatible:

1. All existing workout data will be lost (database reset)
2. All API endpoints change behavior
3. All frontend components need updates
4. Any external integrations will break

## Rollback Plan

If issues arise:
1. Restore from backup: `cp actalog.db.backup-YYYYMMDD-HHMMSS actalog.db`
2. Revert code changes
3. Rebuild application

## Success Criteria

The refactoring is complete when:
- ✅ Database schema matches requirements specification
- ✅ Can create workout templates (independent of users)
- ✅ Can create/browse standard WODs (Fran, Murph, etc.)
- ✅ Can log workouts by selecting a template
- ✅ Multiple users can reference the same template
- ✅ Can track workout history per user
- ✅ PRs still work correctly
- ✅ Frontend shows templates vs logged workouts correctly
- ✅ All tests pass

## Related Documents

- [REQUIREMENTS_VS_IMPLEMENTATION.md](./REQUIREMENTS_VS_IMPLEMENTATION.md) - Analysis that triggered this refactoring
- [REQUIREMENTS.md](./REQUIREMENTS.md) - Original requirements specification
- [DATABASE_SCHEMA.md](./DATABASE_SCHEMA.md) - Will be updated to reflect new schema

---

**Status:** Phase 1 in progress
**Last Updated:** 2025-11-12
**Estimated Completion:** TBD (depends on scope decisions)
