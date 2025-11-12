# Requirements vs Implementation Analysis

**Date:** 2025-11-12
**Current Version:** v0.2.0-beta
**Reviewed By:** Claude Code

## Executive Summary

This document identifies **critical architectural discrepancies** between the requirements specified in `REQUIREMENTS.md` and the current implementation. The most significant issue is that the **data model architecture is fundamentally different** from what was specified in requirements.

**Status:** 🔴 **MAJOR MISALIGNMENT** - Requires architectural decision

## Critical Discrepancies

### 1. Workout Architecture (CRITICAL)

#### Requirements Specification (REQUIREMENTS.md lines 800-808, 818-821)

**Conceptual Model:**
```
Workout = Warmup + WOD(s) + Strength Movement(s)
```

**Requirements state:**
- "A workout is created before logging"
- "A workout is independent of other workouts and is not linked to users in the workout definition"
- "Each user can log multiple workouts each day"
- Workouts are **templates** that users reference when logging

**Required Entities:**
- `Workout`: Template definition (id, user_id FK, date, notes)
- `UserWorkout`: Junction table linking users to workout instances (id, user_id FK, workout_id FK, date_performed)
- **Many-to-Many relationship** between Users and Workouts via UserWorkout

#### Current Implementation (database.go lines 180-191, DATABASE_SCHEMA.md lines 165-193)

**Actual Database Schema:**
```sql
CREATE TABLE IF NOT EXISTS workouts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,  -- WORKOUTS ARE USER-SPECIFIC
    workout_date DATE NOT NULL,
    workout_type TEXT NOT NULL,
    workout_name TEXT,
    notes TEXT,
    total_time INTEGER,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

**Implementation Reality:**
- Workouts **ARE** user-specific instances (user_id NOT NULL)
- No separate workout templates table exists
- No UserWorkout junction table exists
- Workouts represent **actual performed workouts**, not templates
- **One-to-Many relationship** (User has many Workouts)

**Impact:** 🔴 **FUNDAMENTAL ARCHITECTURAL DIFFERENCE**

This is not a minor schema difference - it's a completely different conceptual model:
- Requirements: Template-based system (like recipe books - create template once, reference many times)
- Implementation: Instance-based system (like diary entries - each workout is a unique user-owned record)

---

### 2. WOD Entity (CRITICAL)

#### Requirements Specification (REQUIREMENTS.md lines 800-801, 816, 819, 838-841)

**Required Entity:**
```
WOD: id, name, source, type, regime, score_type, description, url, notes,
     created_at, updated_at, updated_by
```

**Required Fields:**
- **Name**: e.g., "Fran", "Murph", "Cindy"
- **Source**: Crossfit named workout, Other Coach, Self-recorded
- **Type**: Benchmark, Hero, Girl, Notables, Games, Endurance, Self-created
- **Regime**: EMOM, AMRAP, Fastest Time, Slowest Round, Get Stronger, Skills
- **Score Type**: Time [HH:MM:SS], Rounds and Reps [Rounds:Reps], Max Weight [Decimal]
- **Description**: WOD description text
- **URL**: Video or online resource link
- **Notes**: Additional notes

**Required Junction Table:**
```
WorkoutWOD: id, workout_id (FK), wod_id (FK), score_value, created_at, updated_at
```

**Requirements state:**
- "A Workout can include multiple WODs (Many-to-Many via WorkoutWOD)"
- "A WOD can be included in multiple Workouts"
- "Each WOD in a UserWorkout should allow a score to be recorded"

#### Current Implementation

**Reality:** ❌ **WOD entity does NOT exist**

No `wods` table exists in:
- `internal/repository/database.go`
- `docs/DATABASE_SCHEMA.md`

No `workout_wods` junction table exists.

**Impact:** 🔴 **MISSING CORE FEATURE**

WODs are a fundamental part of CrossFit training (Fran, Murph, Cindy, etc.). The requirements explicitly call for this, but it's completely missing from the implementation.

---

### 3. Strength/Movement Naming (MINOR)

#### Requirements Specification (REQUIREMENTS.md line 817, 820)

**Required Entity:**
```
Strength: id, name, movement_type (weightlifting/cardio/gymnastics),
          created_at, updated_at, updated_by
```

**Required Junction Table:**
```
WorkoutStrength: id, workout_id (FK), strength_id (FK), weight, reps, sets,
                 created_at, updated_at
```

#### Current Implementation (database.go lines 197-207, DATABASE_SCHEMA.md lines 195-221)

**Actual Tables:**
```
movements (not "strength")
workout_movements (not "workout_strength")
```

**Impact:** 🟡 **MINOR - NAMING CONVENTION DIFFERENCE**

The functionality is essentially the same, just different naming:
- Requirements: "Strength" entity
- Implementation: "movements" entity

The implementation is actually better (more generic - supports cardio/gymnastics/bodyweight, not just strength movements).

---

### 4. Entity Field Discrepancies

#### Missing Fields in Current Implementation

**From Requirements (line 815, 816, 817, 844):**
All entities should have:
- `updated_by` field (references user who last updated)

**Current Implementation:**
- ❌ No `updated_by` fields exist on any tables
- ✅ `created_at` and `updated_at` exist on all tables

**Impact:** 🟡 **MINOR - AUDIT TRAIL INCOMPLETE**

---

### 5. Additional Missing Entities

#### From Requirements (lines 822-824)

**Required but NOT Implemented:**

1. **UserSetting**: User preferences (notification_preferences, data_export_format)
2. **AuditLog**: Action logging (user_id, action, timestamp, details)
3. **Backup**: Backup tracking (backup_date, status, file_location)

**Current Status:**
- ❌ None of these tables exist
- ❌ No domain models exist
- ❌ No repository/service/handler layers exist

**Impact:** 🟢 **LOW - FUTURE FEATURES**

These were marked as "Additional Considerations" and may not be v1.0 requirements.

---

## What IS Correctly Implemented

### ✅ User Entity

Matches requirements closely with additions:
- ✅ Basic fields: id, email, password_hash, name, created_at, updated_at
- ✅ Profile: profile_image (called profile_picture_url in requirements)
- ➕ **Bonus:** role, last_login_at, birthday (not in requirements)
- ➕ **Bonus:** Email verification fields (email_verified, verification_token, etc.)
- ➕ **Bonus:** Password reset fields (reset_token, reset_token_expires_at)

### ✅ Movement/Strength Entity (Functionally Correct)

Despite naming difference, implementation is solid:
- ✅ id, name, type (movement_type in requirements)
- ✅ created_by (foreign key to users)
- ✅ is_standard flag (distinguishes pre-seeded vs custom)
- ✅ 31 standard movements pre-seeded
- ➕ **Bonus:** description field (not in requirements)

### ✅ Workout_Movements Junction Table (Functionally Correct)

Despite naming (workout_strength in requirements), implementation is complete:
- ✅ workout_id, movement_id foreign keys
- ✅ weight, reps, sets
- ➕ **Bonus:** time, distance, is_rx, is_pr, notes, order_index (not in requirements)

---

## Comparison Table

| Feature | Requirements | Implementation | Status |
|---------|-------------|----------------|--------|
| **Workout Model** | Templates (user-independent) | Instances (user-owned) | 🔴 DIFFERENT |
| **Workout-User Relationship** | Many-to-Many via UserWorkout | One-to-Many (direct FK) | 🔴 DIFFERENT |
| **UserWorkout Junction Table** | Required | ❌ Does not exist | 🔴 MISSING |
| **WOD Entity** | Required with 10+ fields | ❌ Does not exist | 🔴 MISSING |
| **WorkoutWOD Junction** | Required | ❌ Does not exist | 🔴 MISSING |
| **Movement/Strength Entity** | "Strength" | "movements" | 🟡 RENAMED |
| **WorkoutStrength Junction** | "workout_strength" | "workout_movements" | 🟡 RENAMED |
| **updated_by Fields** | Required on all entities | ❌ Missing | 🟡 MINOR |
| **UserSetting Entity** | Mentioned | ❌ Missing | 🟢 FUTURE |
| **AuditLog Entity** | Mentioned | ❌ Missing | 🟢 FUTURE |
| **Backup Entity** | Mentioned | ❌ Missing | 🟢 FUTURE |

---

## Functional Impact Analysis

### What Works Today (v0.2.0-beta)

The current implementation **does work** and supports:
- ✅ User registration and authentication
- ✅ Logging workouts with movements
- ✅ Tracking weight, sets, reps for each movement
- ✅ Personal record (PR) tracking
- ✅ Movement library (31 standard + custom)
- ✅ Workout history and statistics
- ✅ Mobile-first Vue.js frontend

### What Cannot Be Built with Current Architecture

Due to the architectural differences, these features are **blocked or difficult**:

1. **Workout Templates** ❌
   - Cannot create a workout template once and reference it multiple times
   - Every workout log creates a new database record with duplicate data
   - No way to say "I did Fran today" by referencing a template

2. **Benchmark WOD Tracking** ❌
   - Cannot pre-define famous WODs (Fran, Murph, Cindy, etc.)
   - Cannot track scores for standard WODs
   - Cannot compare performance on the same WOD over time easily

3. **WOD Score Types** ❌
   - Cannot capture time-based scores (HH:MM:SS for "Fran")
   - Cannot capture rounds+reps (for AMRAP workouts)
   - No structured way to record WOD performance

4. **Community/Social Features** ⚠️
   - Hard to share "workout templates" between users
   - Each user must recreate workouts manually
   - Cannot have gym-wide templates that users reference

---

## Architectural Decision Required

### Option 1: Update Requirements to Match Implementation

**Pros:**
- No code changes needed
- Current system works well for basic use case
- Simpler architecture (fewer tables)
- Already tested and deployed

**Cons:**
- Cannot build benchmark WOD tracking features
- More data duplication (same workout logged multiple times stores full details each time)
- Harder to add template sharing later

**Recommendation:** ✅ **Best for v1.0 if WODs are not critical**

### Option 2: Refactor Implementation to Match Requirements

**Pros:**
- Enables full WOD tracking (Fran, Murph, etc.)
- Workout templates reduce data duplication
- Aligns with original vision
- Better for social/sharing features later

**Cons:**
- Requires significant refactoring (database schema, repositories, services, handlers, frontend)
- Need to migrate existing data
- Higher complexity
- Breaks current API contracts (frontend will need updates)

**Recommendation:** ⚠️ **Only if WOD tracking is a must-have for v1.0**

### Option 3: Hybrid Approach

Keep current instance-based workouts, but add WOD support:

1. Keep `workouts` table as-is (user-owned instances)
2. Add `wods` table (pre-defined WODs like Fran, Murph)
3. Add `workout_wods` junction table
4. Add optional `workout_template_id` field to workouts (for later)

**Pros:**
- Minimal disruption to current system
- Adds WOD tracking capability
- Leaves door open for templates later
- Incremental improvement

**Cons:**
- Still not fully aligned with requirements
- Hybrid model may be confusing

---

## Recommendations

### Immediate (This Sprint)

1. **Document the architectural decision** - Add to ARCHITECTURE.md explaining why instance-based model was chosen
2. **Update REQUIREMENTS.md** - Either:
   - Mark WOD entity as "v2.0 feature" if going with Option 1
   - Or commit to Option 2 refactoring with timeline

### Short Term (Next Sprint)

If staying with current model (Option 1):
- Update DATABASE_SCHEMA.md to call out "Workouts are instances, not templates"
- Update README to clarify "template" refers to pre-configured workout names, not database templates

If refactoring (Option 2):
- Create migration plan and timeline
- Design new schema with backwards compatibility
- Plan API versioning strategy

### Long Term (v2.0+)

Consider adding:
- `workout_templates` table for reusable workout definitions
- `wods` table for benchmark WOD tracking
- Template sharing between users
- Social features (leaderboards for specific WODs)

---

## Files Reviewed

- `docs/REQUIREMENTS.md` (lines 800-864)
- `docs/DATABASE_SCHEMA.md` (full document)
- `internal/repository/database.go` (lines 160-310)
- `internal/domain/workout.go`
- `internal/repository/workout_repository.go`
- `internal/service/workout_service.go`
- `internal/handler/workout_handler.go`
- `cmd/actalog/main.go` (route definitions)

---

## Related Documents

- [REQUIREMENTS.md](./REQUIREMENTS.md) - Original requirements specification
- [DATABASE_SCHEMA.md](./DATABASE_SCHEMA.md) - Current database schema documentation
- [ARCHITECTURE.md](./ARCHITECTURE.md) - Architecture patterns and design decisions
- [CHANGELOG.md](./CHANGELOG.md) - Version history

---

**Prepared by:** Claude Code
**Review Date:** 2025-11-12
**Status:** Awaiting architectural decision from project owner
