# Migration v0.4.0 Progress Report

**Date:** 2025-11-10
**Status:** ~50% Complete (Repository Layer Finished, Service Layer Next)
**Branch:** `claude/setup-docs-review-011CUsVWbDWKPRCHnsHotMyr`

## Overview

This migration transforms the ActaLog database schema from v0.3.3 (user-specific workout instances) to v0.4.0 (template-based architecture with WODs) to match REQUIREMENTS.md specifications.

## Architecture Change

### Before (v0.3.3)
- **Workouts**: User-specific instances with `user_id`, `workout_date`
- Each user's workout was a unique row
- No concept of reusable templates or WODs

### After (v0.4.0)
- **Workouts**: Reusable templates (shared or user-specific)
- **UserWorkouts**: Junction table linking users to templates with dates
- **WODs**: CrossFit benchmark workouts (Fran, Murph, etc.)
- **WorkoutWODs**: Links WODs to workout templates
- **Renamed tables**: `movements` → `strength_movements`, `workout_movements` → `workout_strength`

## Completed Work (50%)

### ✅ 1. Migration Code (migrations.go)
**File:** `internal/repository/migrations.go`
**Status:** Complete (631 lines added)

- **7 migration phases implemented:**
  1. Create backup tables
  2. Create new tables (wods, user_workouts, workout_wods)
  3. Add columns to workouts (name, created_by)
  4. Rename tables (movements → strength_movements, workout_movements → workout_strength)
  5. Migrate data (conservative approach: each user's workout becomes their own template)
  6. Remove old columns (marked as unused due to SQLite limitations)
  7. Seed 9 standard WODs (Fran, Grace, Helen, Diane, Karen, Murph, Cindy, Annie, DT)

- **Multi-database support:** SQLite, PostgreSQL, MySQL

### ✅ 2. Domain Models (5 files)
**Status:** Complete

#### New Files:
- **`internal/domain/wod.go`** (105 lines)
  - WOD entity with source, type, regime, score_type fields
  - WODRepository interface (9 methods)

- **`internal/domain/user_workout.go`** (57 lines)
  - UserWorkout entity (junction: user + workout template + date)
  - UserWorkoutWithDetails view model
  - UserWorkoutRepository interface (10 methods)

- **`internal/domain/workout_wod.go`** (55 lines)
  - WorkoutWOD entity (junction: workout template + WOD)
  - WorkoutWODWithDetails view model
  - WorkoutWODRepository interface (8 methods)

#### Updated Files:
- **`internal/domain/workout.go`**
  - Removed: UserID, WorkoutDate, WorkoutType, WorkoutName, TotalTime
  - Added: Name, CreatedBy, WODs collection
  - Now represents reusable templates
  - WorkoutWithUsageStats added (tracks how many times template used)

- **`internal/domain/movement.go`**
  - Added clarifying comments about table name changes
  - Updated PersonalRecord to reference user_workouts.id

### ✅ 3. Repository Layer (6 files)
**Status:** Complete (all using standard database/sql, no sqlx)

#### New Repositories:
- **`internal/repository/wod_repository.go`** (246 lines)
  - Full CRUD for WODs
  - Methods: Create, GetByID, GetByName, List, ListStandard, ListByUser, Update, Delete, Search

- **`internal/repository/user_workout_repository.go`** (292 lines)
  - Manages user workout instances
  - GetByIDWithDetails: joins with workout templates, movements, WODs
  - Methods: Create, GetByID, GetByIDWithDetails, ListByUser, ListByUserWithDetails, ListByUserAndDateRange, Update, Delete, GetByUserWorkoutDate

- **`internal/repository/workout_wod_repository.go`** (220 lines)
  - Links WODs to workout templates
  - Methods: Create, GetByID, ListByWorkout, ListByWorkoutWithDetails, Update, Delete, DeleteByWorkout, TogglePR

#### Updated Repositories:
- **`internal/repository/workout_repository.go`** (325 lines)
  - Complete rewrite for template operations
  - GetByIDWithDetails: loads movements and WODs
  - GetUsageStats: counts how many times template has been used
  - Methods: Create, GetByID, GetByIDWithDetails, List, ListByUser, ListStandard, Update, Delete, Search, Count, GetUsageStats

- **`internal/repository/movement_repository.go`** (181 lines)
  - Updated to query `strength_movements` table (renamed from `movements`)
  - All queries changed from `movements` to `strength_movements`

- **`internal/repository/workout_movement_repository.go`** (302 lines)
  - Updated to query `workout_strength` table (renamed from `workout_movements`)
  - PR tracking queries updated to join through `user_workouts` junction table
  - Methods: GetByUserIDAndMovementID, GetPersonalRecords, GetMaxWeightForMovement, GetPRMovements all updated for new schema

## Remaining Work (50%)

### ⏳ 4. Service Layer
**Status:** Pending (compilation errors need fixing)

**Current Issues:**
- `internal/service/workout_service.go` still uses v0.3.3 schema
- References `workout.UserID`, `workout.WorkoutDate`, `workout.WorkoutType` (removed fields)
- Calls `workoutRepo.GetByUserID()` (method no longer exists)

**Required Changes:**

#### A. Update WorkoutService (Workout Templates)
**File:** `internal/service/workout_service.go`

New methods needed:
- `CreateTemplate(userID int64, workout *domain.Workout) error` - Create reusable template
- `GetTemplate(templateID int64) (*domain.Workout, error)` - Get template with movements/WODs
- `ListTemplates(userID *int64, limit, offset int) ([]*domain.Workout, error)` - List templates (user or standard)
- `UpdateTemplate(templateID, userID int64, updates *domain.Workout) error` - Update user's template
- `DeleteTemplate(templateID, userID int64) error` - Delete user's template
- `GetTemplateUsageStats(templateID int64) (*domain.WorkoutWithUsageStats, error)` - How many times template used
- `AddMovementToTemplate(templateID, movementID int64, userID int64, wm *domain.WorkoutMovement) error`
- `AddWODToTemplate(templateID, wodID int64, userID int64, wod *domain.WorkoutWOD) error`

Remove methods:
- `CreateWorkout()`, `GetWorkout()`, `ListUserWorkouts()`, `UpdateWorkout()`, `DeleteWorkout()`, `GetWorkoutsByDateRange()` - These now belong in UserWorkoutService

Keep methods:
- `ListMovements()`, `DetectAndFlagPRs()`, `GetPersonalRecords()`, `GetPRMovements()`, `TogglePRFlag()` - Movement-related methods stay

#### B. Create UserWorkoutService (Log Workout Instances)
**File:** `internal/service/user_workout_service.go` (NEW)

Purpose: Log when a user performs a workout (creates user_workouts entry)

Dependencies:
```go
type UserWorkoutService struct {
    userWorkoutRepo     domain.UserWorkoutRepository
    workoutRepo         domain.WorkoutRepository
    workoutMovementRepo domain.WorkoutMovementRepository
}
```

Methods needed:
- `LogWorkout(userID, templateID int64, date time.Time, notes *string, totalTime *int, workoutType *string) (*domain.UserWorkout, error)` - Log that user did a workout template
- `GetLoggedWorkout(userWorkoutID, userID int64) (*domain.UserWorkoutWithDetails, error)` - Get logged workout with full details
- `ListLoggedWorkouts(userID int64, limit, offset int) ([]*domain.UserWorkoutWithDetails, error)` - User's workout history
- `ListLoggedWorkoutsByDateRange(userID int64, startDate, endDate time.Time) ([]*domain.UserWorkout, error)` - History by date range
- `UpdateLoggedWorkout(userWorkoutID, userID int64, updates *domain.UserWorkout) error` - Update notes/time on logged workout
- `DeleteLoggedWorkout(userWorkoutID, userID int64) error` - Delete workout log entry
- `GetWorkoutStatsForMonth(userID int64, year, month int) (int, error)` - Count workouts in month (for dashboard)

#### C. Create WODService
**File:** `internal/service/wod_service.go` (NEW)

Purpose: Manage WOD (Workout of the Day) benchmark workouts

Dependencies:
```go
type WODService struct {
    wodRepo domain.WODRepository
}
```

Methods needed:
- `CreateWOD(userID int64, wod *domain.WOD) error` - Create custom WOD
- `GetWOD(wodID int64) (*domain.WOD, error)` - Get WOD by ID
- `GetWODByName(name string) (*domain.WOD, error)` - Get WOD by name (e.g., "Fran")
- `ListStandardWODs() ([]*domain.WOD, error)` - List standard CrossFit benchmarks
- `ListUserWODs(userID int64) ([]*domain.WOD, error)` - List user's custom WODs
- `ListAllWODs(userID int64) ([]*domain.WOD, error)` - Combined list (standard + user)
- `SearchWODs(query string) ([]*domain.WOD, error)` - Search by name
- `UpdateWOD(wodID, userID int64, updates *domain.WOD) error` - Update custom WOD
- `DeleteWOD(wodID, userID int64) error` - Delete custom WOD

#### D. Create WorkoutWODService
**File:** `internal/service/workout_wod_service.go` (NEW)

Purpose: Link WODs to workout templates

Dependencies:
```go
type WorkoutWODService struct {
    workoutWODRepo domain.WorkoutWODRepository
    workoutRepo    domain.WorkoutRepository
    wodRepo        domain.WODRepository
}
```

Methods needed:
- `AddWODToWorkout(workoutID, wodID int64, userID int64, orderIndex int, division *string) (*domain.WorkoutWOD, error)` - Add WOD to template
- `RemoveWODFromWorkout(workoutWODID, userID int64) error` - Remove WOD from template
- `UpdateWorkoutWOD(workoutWODID, userID int64, scoreValue, division *string) error` - Update WOD in workout
- `ToggleWODPR(workoutWODID, userID int64) error` - Toggle PR flag on WOD
- `ListWODsForWorkout(workoutID int64) ([]*domain.WorkoutWODWithDetails, error)` - Get all WODs in a workout template

### ⏳ 5. API Handler Layer
**Status:** Pending

**Files to Update/Create:**
- `internal/handler/workout_handler.go` - Update for template operations
- `internal/handler/user_workout_handler.go` - NEW for logging workouts
- `internal/handler/wod_handler.go` - NEW for WOD management
- `internal/handler/workout_wod_handler.go` - NEW for linking WODs to workouts

**Routes to Add:** (in `cmd/actalog/main.go`)
```go
// Workout Templates
POST   /api/templates             - Create template
GET    /api/templates             - List templates (standard + user)
GET    /api/templates/:id         - Get template details
PUT    /api/templates/:id         - Update template
DELETE /api/templates/:id         - Delete template
GET    /api/templates/:id/stats   - Get usage stats

// Logging Workouts (User Workouts)
POST   /api/workouts              - Log a workout instance
GET    /api/workouts              - List logged workouts
GET    /api/workouts/:id          - Get logged workout details
PUT    /api/workouts/:id          - Update logged workout
DELETE /api/workouts/:id          - Delete logged workout

// WODs
POST   /api/wods                  - Create custom WOD
GET    /api/wods                  - List WODs (standard + user)
GET    /api/wods/:id              - Get WOD details
PUT    /api/wods/:id              - Update WOD
DELETE /api/wods/:id              - Delete WOD
GET    /api/wods/search?q=Fran    - Search WODs

// Link WODs to Templates
POST   /api/templates/:id/wods    - Add WOD to template
DELETE /api/templates/:id/wods/:wod_id - Remove WOD from template
```

### ⏳ 6. Frontend Layer
**Status:** Pending

**Files to Update:**
- `web/src/views/LogWorkoutView.vue` - Update to select from templates, then log
- `web/src/views/WorkoutsView.vue` - Show logged workouts (not templates)
- `web/src/views/DashboardView.vue` - Update stats to use user_workouts
- `web/src/components/*` - Update components to work with new API

**New Views Needed:**
- `web/src/views/TemplatesView.vue` - Browse/create workout templates
- `web/src/views/WODsView.vue` - Browse/create WODs
- `web/src/views/TemplateDetailView.vue` - View/edit template with movements and WODs

### ⏳ 7. Testing
**Status:** Pending

- Test migration with sample data
- Test all API endpoints
- Test frontend workflows
- Integration tests for PR tracking with new schema

## Technical Debt Addressed

1. **No more sqlx dependency**: All repositories use standard `database/sql`
2. **Consistent error handling**: All repos use `fmt.Errorf("context: %w", err)` pattern
3. **Null handling**: Proper use of sql.Null* types for nullable database columns
4. **Row scanning helpers**: Each repository has `scanX()` methods for code reuse

## Key Design Decisions

### 1. Conservative Data Migration
**Decision:** Each user's workout becomes their own private template
**Rationale:** Preserves all historical data without requiring user intervention
**Impact:** Users can later consolidate templates if desired

### 2. Standard vs Custom
**Decision:** Both workouts and WODs can be standard (system) or custom (user-created)
**Implementation:** `created_by IS NULL` for standard, `created_by = user_id` for custom
**Benefit:** Users can create custom variations of standard workouts

### 3. Template Usage Tracking
**Decision:** Track how many times each template has been used
**Implementation:** `WorkoutRepository.GetUsageStats()` counts user_workouts entries
**Benefit:** Popular templates can be identified and promoted

### 4. Renamed Tables
**Decision:** `movements` → `strength_movements`, `workout_movements` → `workout_strength`
**Rationale:** Matches REQUIREMENTS.md naming convention
**Note:** SQLite doesn't support column rename easily, so `movement_id` column kept in workout_strength

## Next Steps (Priority Order)

1. **Service Layer** (~4-5 hours)
   - Update WorkoutService for templates
   - Create UserWorkoutService for logging
   - Create WODService
   - Create WorkoutWODService

2. **Handler Layer** (~2-3 hours)
   - Update workout_handler.go
   - Create user_workout_handler.go
   - Create wod_handler.go
   - Create workout_wod_handler.go
   - Update routes in main.go

3. **Frontend Layer** (~4-5 hours)
   - Update existing views for new API
   - Create template management views
   - Create WOD management views
   - Update dashboard for user_workouts stats

4. **Testing** (~2-3 hours)
   - Run migration on test data
   - Test all API endpoints
   - Test frontend workflows
   - Verify PR tracking still works

**Total Estimated Time Remaining:** 12-16 hours

## Files Changed Summary

### New Files (10)
- `internal/domain/wod.go`
- `internal/domain/user_workout.go`
- `internal/domain/workout_wod.go`
- `internal/repository/wod_repository.go`
- `internal/repository/user_workout_repository.go`
- `internal/repository/workout_wod_repository.go`
- `docs/SCHEMA_MIGRATION_v0.4.0.md`
- `docs/MIGRATION_v0.4.0_PROGRESS.md` (this file)
- (4 more service files needed)
- (3 more handler files needed)

### Modified Files (8)
- `internal/domain/workout.go` - Template architecture
- `internal/domain/movement.go` - Comments and PersonalRecord updates
- `internal/repository/migrations.go` - Added v0.4.0 migration (631 lines)
- `internal/repository/workout_repository.go` - Complete rewrite for templates
- `internal/repository/movement_repository.go` - Updated for strength_movements table
- `internal/repository/workout_movement_repository.go` - Updated for workout_strength table
- `docs/DATABASE_SCHEMA.md` - Updated to reflect v0.3.3 reality
- (Services, handlers, frontend pending)

## Compilation Status

**Current:** ❌ Build fails in service layer
**Error:** WorkoutService using removed fields (UserID, WorkoutDate, etc.)
**Fix:** Complete service layer refactoring per section 4 above

**After Services Fixed:** Backend should compile ✅
**After Handlers Fixed:** API should work ✅
**After Frontend Fixed:** Full application functional ✅

## Questions to Resolve

1. **Dashboard Stats:** Should dashboard show stats from user_workouts or keep aggregating from workouts table?
   - **Answer:** Should use user_workouts (logged instances)

2. **Template Visibility:** Should users be able to browse other users' public templates?
   - **Current:** No, only standard (system) and own templates
   - **Future Enhancement:** Could add `is_public` flag to workouts table

3. **WOD Scores:** Should scores be stored in workout_wods or user_workouts?
   - **Current:** `workout_wods.score_value` stores the score
   - **Note:** This is per-template, not per-logged-instance
   - **Consider:** Might need `user_workout_wods` junction table for instance-specific scores

## Migration Safety

- **Backup tables created:** `workouts_backup`, `movements_backup`, `workout_movements_backup`
- **Rollback:** If migration fails, restore from backup tables
- **Data preservation:** All historical data migrated to user_workouts table
- **Testing recommended:** Run on copy of production database first

## Resources

- Original requirements: `docs/REQUIREMENTS.md`
- Schema design: `docs/SCHEMA_MIGRATION_v0.4.0.md`
- Database schema: `docs/DATABASE_SCHEMA.md`
- Migration code: `internal/repository/migrations.go` (search for "v0.4.0")
