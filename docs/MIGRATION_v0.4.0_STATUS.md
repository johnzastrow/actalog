# Migration v0.4.0 Status Report

**Date:** 2025-11-11
**Migration:** 0.4.0 - Transform schema to match REQUIREMENTS.md
**Status:** ✅ **DATABASE MIGRATED** ⚠️ **CODE NEEDS UPDATES**

## Summary

Migration 0.4.0 has been **successfully applied** to the database. The schema now matches the requirements specification. However, the Go codebase still references the old table names and structures.

## Database Changes Applied

### Tables Created
- ✅ `wods` - WOD (Workout of the Day) definitions
- ✅ `user_workouts` - Junction table linking users to workouts they've performed
- ✅ `workout_wods` - Links workout templates to WODs

### Tables Renamed
- ✅ `movements` → `strength_movements`
- ✅ `workout_movements` → `workout_strength`

### Backup Tables Created
- `workouts_backup_v033` - Backup of workouts table before migration
- `workout_movements_backup_v033` - Backup of workout_movements table

### Data Migration
- ✅ Existing user workouts converted to template + user_workout structure
- ✅ Each old workout became a private template for that user
- ✅ 9 standard WODs seeded (Fran, Grace, Helen, Diane, Karen, Murph, Cindy, Annie, DT)

## Current Database Schema (Post-Migration)

```
users
refresh_tokens
wods                              # NEW
user_workouts                     # NEW
workout_wods                      # NEW
workouts                          # MODIFIED (now templates)
strength_movements                # RENAMED from movements
workout_strength                  # RENAMED from workout_movements
schema_migrations
```

## Required Code Updates

### Critical Updates Needed

The following files **MUST** be updated to work with the new schema:

#### 1. Domain Models (`internal/domain/`)

**File: `movement.go`**
- Currently references `movements` table
- Needs: Update to `strength_movements`
- Impact: All queries will fail

**New file needed: `wod.go`**
```go
type WOD struct {
    ID          int64     `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Source      string    `json:"source" db:"source"`
    Type        string    `json:"type" db:"type"`
    Regime      string    `json:"regime" db:"regime"`
    ScoreType   string    `json:"score_type" db:"score_type"`
    Description string    `json:"description" db:"description"`
    URL         *string   `json:"url" db:"url"`
    Notes       *string   `json:"notes" db:"notes"`
    IsStandard  bool      `json:"is_standard" db:"is_standard"`
    CreatedBy   *int64    `json:"created_by" db:"created_by"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type WorkoutWOD struct {
    ID          int64     `json:"id" db:"id"`
    WorkoutID   int64     `json:"workout_id" db:"workout_id"`
    WODID       int64     `json:"wod_id" db:"wod_id"`
    ScoreValue  *string   `json:"score_value" db:"score_value"`
    Division    *string   `json:"division" db:"division"`
    IsPR        bool      `json:"is_pr" db:"is_pr"`
    OrderIndex  int       `json:"order_index" db:"order_index"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
    WOD         *WOD      `json:"wod,omitempty" db:"-"`
}
```

**File: `workout.go`**
- Update `WorkoutMovement` struct name → `WorkoutStrength` (optional, semantic)
- Add references to new WOD-related structures

**File: `user_workout.go`**
- Already exists but may need updates to match new schema
- Ensure it matches the `user_workouts` table structure

#### 2. Repositories (`internal/repository/`)

**File: `movement_repository.go`** → Rename to `strength_movement_repository.go`
- Update all SQL queries: `movements` → `strength_movements`
- Update table references in Create, Update, Delete, List, Search functions

**Example changes:**
```go
// OLD
query := `SELECT * FROM movements WHERE id = ?`

// NEW
query := `SELECT * FROM strength_movements WHERE id = ?`
```

**File: `workout_movement_repository.go`** → Rename to `workout_strength_repository.go`
- Update all SQL queries: `workout_movements` → `workout_strength`
- Update column references: `movement_id` might stay same (SQLite migration didn't rename it)

**New file needed: `wod_repository.go`**
```go
type WODRepository interface {
    Create(wod *domain.WOD) error
    GetByID(id int64) (*domain.WOD, error)
    GetByName(name string) (*domain.WOD, error)
    ListStandard(limit, offset int) ([]*domain.WOD, error)
    ListByUser(userID int64, limit, offset int) ([]*domain.WOD, error)
    Search(query string, limit int) ([]*domain.WOD, error)
    Update(wod *domain.WOD) error
    Delete(id int64) error
}
```

**New file needed: `workout_wod_repository.go`**
```go
type WorkoutWODRepository interface {
    Create(ww *domain.WorkoutWOD) error
    GetByID(id int64) (*domain.WorkoutWOD, error)
    ListByWorkout(workoutID int64) ([]*domain.WorkoutWOD, error)
    Update(ww *domain.WorkoutWOD) error
    Delete(id int64) error
}
```

**File: `database.go`**
- Update seed functions to use `strength_movements` table
- Already has `seedStandardMovements()` - just needs table name update

#### 3. Services (`internal/service/`)

**File: `workout_template_service.go`**
- Update to use new repository names
- May need adjustments for new schema structure

**File: `workout_service.go`** (if exists)
- Update references to `MovementRepository` → `StrengthMovementRepository`

**File: `user_workout_service.go`**
- Verify it works with new `user_workouts` table structure
- May need updates for workout template references

**New file needed: `wod_service.go`**
```go
type WODService struct {
    wodRepo domain.WODRepository
}

func NewWODService(wodRepo domain.WODRepository) *WODService {
    return &WODService{wodRepo: wodRepo}
}

// Implement WOD business logic
```

#### 4. Handlers (`internal/handler/`)

**File: `movement_handler.go`** → Can stay named same, but update internals
- Update to use `StrengthMovementRepository`
- All SQL queries automatically handled by repository

**File: `workout_template_handler.go`**
- Verify it works with new workouts table structure
- Templates now don't have user_id, workout_date, etc.
- Those fields are in user_workouts table

**New file needed: `wod_handler.go`**
```go
type WODHandler struct {
    service *service.WODService
}

// Implement HTTP handlers for WOD CRUD operations
```

#### 5. Main Application (`cmd/actalog/main.go`)

**Updates needed:**
- Initialize `WODRepository`
- Initialize `WorkoutWODRepository`
- Rename `movementRepo` → `strengthMovementRepo` (optional, for clarity)
- Initialize `WODService`
- Initialize `WorkoutWODService`
- Initialize `WODHandler`
- Wire up WOD routes

**New routes to add:**
```go
// WOD routes (already documented in requirements)
r.Get("/wods", wodHandler.ListWODs)
r.Get("/wods/search", wodHandler.SearchWODs)
r.Get("/wods/{id}", wodHandler.GetWOD)

// Protected WOD management
r.Post("/wods", wodHandler.CreateWOD)
r.Put("/wods/{id}", wodHandler.UpdateWOD)
r.Delete("/wods/{id}", wodHandler.DeleteWOD)

// Workout-WOD linking
r.Post("/templates/{workout_id}/wods", workoutWODHandler.AddWODToWorkout)
r.Get("/templates/{workout_id}/wods", workoutWODHandler.ListWODsForWorkout)
r.Put("/templates/wods/{workout_wod_id}", workoutWODHandler.UpdateWorkoutWOD)
r.Delete("/templates/wods/{workout_wod_id}", workoutWODHandler.RemoveWODFromWorkout)
r.Post("/templates/wods/{workout_wod_id}/toggle-pr", workoutWODHandler.ToggleWODPR)
```

## Breaking Changes

### API Endpoints
- `/api/movements` still works (no change needed, just internal table name)
- `/api/workouts` behavior changes:
  - **Before:** Returns user-specific workout logs
  - **After:** Returns workout templates (user_workouts table stores instances)

### Data Model
- **Workouts are now templates**, not user-specific logs
- User-specific workout data moved to `user_workouts` table
- Need new endpoints for logging workouts vs managing templates

## Recommended Next Steps

### Option 1: Complete the Migration (Recommended)

1. **Create WOD domain models and repositories** (`wod.go`, `wod_repository.go`)
2. **Rename movement files** to reflect "strength_movements"
3. **Update all SQL queries** in repositories to use new table names
4. **Update services** to use renamed repositories
5. **Update handlers** to use updated services
6. **Update main.go** to wire everything together
7. **Test all endpoints** to ensure they work
8. **Update frontend** to use new API structure

**Estimated Effort:** 4-6 hours

### Option 2: Rollback the Migration

If this is too disruptive:

```bash
# Restore backup
cp actalog.db.backup_before_v040_* actalog.db

# Remove migration 0.4.0 from code
# Edit internal/repository/migrations.go and comment out migration 0.4.0
```

**Note:** You'll lose the 9 seeded WODs and the new table structure.

### Option 3: Hybrid Approach

Keep the database as-is but create compatibility layer:
- Create views or aliases for old table names
- Add wrapper repositories that translate between old and new

**Not recommended** - adds complexity without solving the root issue.

## Files That Reference Old Table Names

Run these commands to find all references:

```bash
# Find "movements" table references
grep -r "FROM movements" internal/
grep -r "INTO movements" internal/
grep -r "movements WHERE" internal/

# Find "workout_movements" table references
grep -r "FROM workout_movements" internal/
grep -r "INTO workout_movements" internal/
```

## Testing Strategy

After code updates:

1. **Unit Tests:** Update table names in test fixtures
2. **Integration Tests:** May need updates for new schema
3. **Manual Testing:**
   - Create a workout template
   - Log a workout using that template
   - Create a WOD
   - Link WOD to a workout template
   - Verify all CRUD operations work

## Conclusion

The database migration was successful! The schema now matches the requirements specification perfectly. The code updates are straightforward but numerous - primarily find-and-replace operations on table names, plus creating the new WOD-related code.

**Current State:**
- ✅ Database schema: Aligned with requirements
- ⚠️ Go codebase: Still using old table names
- ❌ Application: Will not start/work until code is updated

**Recommendation:** Complete the code updates to fully align with requirements. This is the right direction for the project.
