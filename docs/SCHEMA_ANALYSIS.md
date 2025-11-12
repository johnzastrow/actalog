# Database Schema Analysis - Requirements vs Implementation

**Date:** 2025-11-11
**Current Version:** 0.2.0-beta
**Database Schema Version:** 0.3.3

## Executive Summary

This document analyzes the **critical conceptual mismatch** between the requirements (REQUIIREMENTS.md) and the current database implementation. Your concern about conceptual misalignment is **100% correct**.

## The Fundamental Mismatch

### What REQUIREMENTS.md Specifies

The requirements document describes a **template-based system** where:

1. **Workouts are reusable templates** - Pre-defined combinations of movements and WODs that can be used by multiple users
2. **User Workouts** - Individual instances where a user performs a workout template on a specific date
3. **WODs** - Separate entities (named CrossFit workouts like Fran, Murph, Helen)
4. **Workout Templates contain** both WODs and Strength Movements

**Requirements Quote (line 800-808):**
```
"Workout = Warmup + WOD(s) + Strength Movement(s)"

"Each user can log multiple workouts each day, and each workout can include
multiple strength movements and WODs. A workout is independent of other
workouts and is not linked to users in the workout definition."

"Each WOD can be linked to zero or more workouts, and each strength movement
can also be linked to zero or more workouts."
```

**Requirements Data Model:**
- **Workout**: Template definition (NOT user-specific, reusable)
- **WOD**: Named workout definition (Fran, Murph, etc.)
- **Strength**: Movement definition (Back Squat, Deadlift, etc.)
- **WorkoutWOD**: Links workouts to WODs (many-to-many)
- **WorkoutStrength**: Links workouts to movements (many-to-many)
- **UserWorkout**: Links users to workout templates they've performed on specific dates

### What's Currently Implemented

The current implementation uses a **direct logging system** where:

1. **Workouts are user-specific logs** - Each workout record IS the actual performed instance
2. **No WODs table** - WOD names stored as strings in workout_name field
3. **No UserWorkout junction table** - user_id is directly in workouts table
4. **Movements** - Correctly implemented as reusable definitions

**Current Schema:**
- users
- workouts (contains user_id, workout_date - these are logged instances, NOT templates)
- movements
- workout_movements (junction table)
- No wods table
- No user_workouts table

## Detailed Comparison

### Table: `workouts`

| Aspect | Requirements (Template Model) | Current Implementation (Log Model) |
|--------|-------------------------------|-------------------------------------|
| **Purpose** | Reusable workout template | User-specific workout log |
| **User Relationship** | None (templates shared) | Direct FK to users table |
| **Date** | Not stored here | Stored as workout_date |
| **Reusability** | Multiple users can use same template | Each record is one user's log |
| **Name** | Template name (required) | workout_name (optional) |

**Current Schema:**
```sql
workouts (
  id,
  user_id FK,              -- ❌ Should NOT be here per requirements
  workout_date DATE,       -- ❌ Should be in user_workouts
  workout_type VARCHAR,    -- ❌ Should be in user_workouts
  workout_name VARCHAR,    -- ⚠️ Should be 'name' and required
  notes TEXT,              -- ✅ Correct (template notes)
  total_time INT,          -- ❌ Should be in user_workouts
  created_at, updated_at
)
```

**Requirements Schema:**
```sql
workouts (
  id,
  name VARCHAR NOT NULL,   -- Template name
  notes TEXT,              -- Template instructions
  created_by INT FK,       -- Who created template (NULL for standard)
  created_at, updated_at
)
```

### Missing Table: `wods`

**Required per REQUIIREMENTS.md (line 801):**

```sql
wods (
  id,
  name VARCHAR UNIQUE NOT NULL,          -- "Fran", "Murph", "Helen"
  source VARCHAR,                         -- "CrossFit", "Other Coach", "Self-recorded"
  type VARCHAR,                           -- "Benchmark", "Hero", "Girl", "Games", etc.
  regime VARCHAR,                         -- "EMOM", "AMRAP", "Fastest Time", etc.
  score_type VARCHAR,                     -- "Time", "Rounds+Reps", "Max Weight"
  description TEXT,                       -- WOD instructions
  url VARCHAR,                            -- Video/reference URL
  notes TEXT,
  is_standard BOOLEAN,                    -- TRUE for official CrossFit WODs
  created_by INT FK,                      -- NULL for standard WODs
  created_at, updated_at
)
```

**Status:** ❌ **Not implemented** - Migration 0.4.0 prepared but not applied

### Missing Table: `user_workouts`

**Required per REQUIIREMENTS.md (lines 821, 827-834):**

```sql
user_workouts (
  id,
  user_id INT FK,           -- Who performed the workout
  workout_id INT FK,        -- Which template they used
  workout_date DATE,        -- When they did it
  workout_type VARCHAR,     -- Type of session
  total_time INT,           -- How long it took
  notes TEXT,               -- User's notes for this session
  created_at, updated_at,
  UNIQUE(user_id, workout_id, workout_date)  -- One template per user per day
)
```

**Status:** ❌ **Not implemented** - Migration 0.4.0 prepared but not applied

### Missing Table: `workout_wods`

**Required per REQUIIREMENTS.md (line 819):**

```sql
workout_wods (
  id,
  workout_id INT FK,        -- Which workout template
  wod_id INT FK,            -- Which WOD is included
  score_value VARCHAR,      -- User's score (from user_workouts)
  division VARCHAR,         -- "Rx", "Scaled", "Beginner"
  is_pr BOOLEAN,            -- Personal record flag
  order_index INT,          -- Order in workout
  created_at, updated_at
)
```

**Status:** ❌ **Not implemented** - Migration 0.4.0 prepared but not applied

### Table: `movements` (now called `strength_movements` in reqs)

| Aspect | Requirements | Current Implementation |
|--------|--------------|------------------------|
| **Table Name** | strength_movements | movements |
| **Purpose** | Exercise definitions | Exercise definitions |
| **Fields** | ✅ Matches | ✅ Matches |

**Status:** ✅ **Correctly implemented** - Only naming difference

### Table: `workout_movements` (requirements call it `workout_strength`)

| Aspect | Requirements | Current Implementation |
|--------|--------------|------------------------|
| **Table Name** | WorkoutStrength | workout_movements |
| **Purpose** | Link workouts to movements | Link workouts to movements |
| **Fields** | ✅ Mostly matches | ✅ Mostly matches |

**Status:** ✅ **Correctly implemented** - Only naming difference

## Migration 0.4.0 Status

The code contains a **prepared but unapplied migration (0.4.0)** in `internal/repository/migrations.go` that would transform the schema to match requirements:

**What Migration 0.4.0 Would Do:**

1. ✅ Create `wods` table
2. ✅ Create `user_workouts` table
3. ✅ Create `workout_wods` table
4. ✅ Rename `movements` → `strength_movements`
5. ✅ Rename `workout_movements` → `workout_strength`
6. ✅ Transform existing data: Convert user-specific workouts into templates + user_workout entries
7. ✅ Seed 9 standard WODs (Fran, Grace, Helen, Diane, Karen, Murph, Cindy, Annie, DT)

**Why Not Applied:**

Looking at the migration history, only migrations 0.1.0 through 0.3.3 have been applied. Migration 0.4.0 exists in the code but is a major breaking change that hasn't been executed.

## Impact on Current Features

### Workout Template Feature (Currently Being Built)

The frontend `WorkoutTemplateEditView.vue` and handler `WorkoutTemplateHandler` are treating "workouts" as templates, which is **conceptually correct per requirements** but uses the **wrong database structure**.

**Current Code Behavior:**
```
User creates "Upper Body Strength" template
  ↓
Saves to workouts table with user_id=5
  ↓
Result: Template is user-specific, can't be shared
```

**Requirements Behavior:**
```
User creates "Upper Body Strength" template
  ↓
Saves to workouts table with created_by=5, NO user_id
  ↓
Template is reusable by anyone
  ↓
When user performs it: Creates user_workout entry linking user+template+date
```

### Field: `rest_seconds` vs `work_time`

We just renamed `rest_seconds` to `work_time` in the handler request structs. This maps to the `time` field in `workout_movements`:

**Current Mapping:**
- Frontend sends: `work_time` (seconds)
- Handler receives: `WorkTime` field
- Maps to domain: `Time` field
- Database column: `time` (INT, seconds)

**Purpose:** Work duration for time-based movements (e.g., "Hold plank for 60 seconds")

This is **semantically correct** regardless of schema version.

## Recommendations

### Option 1: Apply Migration 0.4.0 (Align with Requirements)

**Pros:**
- Implements the requirements as specified
- Enables workout template sharing
- Supports WODs as first-class entities
- Matches the logical data model in REQUIIREMENTS.md

**Cons:**
- Breaking change requiring data migration
- Frontend needs updates to use new endpoints
- All existing code assumes current schema

**Effort:** High (2-3 days of development + testing)

### Option 2: Update Requirements to Match Implementation

**Pros:**
- No migration needed
- Current code continues working
- Simpler mental model (workouts = logs)

**Cons:**
- Can't share workout templates between users
- No WOD library functionality
- Doesn't match original vision in REQUIIREMENTS.md

**Effort:** Low (update documentation only)

### Option 3: Hybrid Approach (Incremental Migration)

**Phase 1:** Keep current schema but clarify naming
- Rename `workouts` → `workout_logs` in documentation
- Update frontend to match current schema
- Document that templates aren't currently supported

**Phase 2:** Add WODs table (future)
- Add wods table
- Add workout_wods table
- Keep current workout_logs for user data

**Phase 3:** Add Templates (future)
- Add workout_templates table
- Add user_workouts junction
- Migrate data

**Effort:** Medium (spread over multiple versions)

## Data Dictionary

### Current Schema (v0.3.3)

#### users
| Field | Type | Purpose | Human Explanation |
|-------|------|---------|-------------------|
| id | BIGINT | Primary key | Unique identifier for each user account |
| email | VARCHAR(255) | Login identifier | User's email address, must be unique |
| password_hash | VARCHAR(255) | Authentication | Bcrypt-hashed password (never stored in plain text) |
| name | VARCHAR(255) | Display name | User's full name or nickname |
| birthday | DATE | Profile info | User's birth date (added v0.3.3) |
| profile_image | TEXT | Avatar URL | Link to user's profile picture |
| role | VARCHAR(50) | Authorization | "user" or "admin" - determines permissions |
| email_verified | BOOLEAN | Email confirmation | Whether user has verified their email |
| email_verified_at | TIMESTAMP | Verification time | When email was verified |
| verification_token | VARCHAR(255) | Email verification | Token sent in verification email |
| verification_token_expires_at | TIMESTAMP | Token expiry | When verification token becomes invalid |
| reset_token | VARCHAR(255) | Password reset | Token for password reset flow |
| reset_token_expires_at | TIMESTAMP | Token expiry | When reset token expires (1 hour) |
| created_at | TIMESTAMP | Account creation | When user registered |
| updated_at | TIMESTAMP | Last modification | When profile was last updated |
| last_login_at | TIMESTAMP | Last session | When user last logged in |

#### workouts (Currently: User-Specific Logs)
| Field | Type | Purpose | Human Explanation |
|-------|------|---------|-------------------|
| id | BIGINT | Primary key | Unique identifier for this workout log |
| user_id | BIGINT | Owner | **Who performed this workout** (FK to users) |
| workout_date | DATE | When performed | **Date user did this workout** |
| workout_type | VARCHAR(50) | Category | Type: "strength", "metcon", "wod", "mixed" |
| workout_name | VARCHAR(255) | Optional title | Name of workout (e.g., "Fran", "Leg Day") |
| notes | TEXT | User notes | Comments about how it went, modifications, etc. |
| total_time | INT | Duration | Total workout time in seconds |
| created_at | TIMESTAMP | Log creation | When this workout was logged |
| updated_at | TIMESTAMP | Last edit | When this log was last modified |

**⚠️ Conceptual Issue:** Per requirements, this table should store **reusable templates**, not user-specific logs. The fields `user_id`, `workout_date`, and `total_time` should be in a separate `user_workouts` table.

#### movements (Per Requirements: Should be `strength_movements`)
| Field | Type | Purpose | Human Explanation |
|-------|------|---------|-------------------|
| id | BIGINT | Primary key | Unique identifier for this movement/exercise |
| name | VARCHAR(255) | Movement name | Exercise name (e.g., "Back Squat", "Pull-up") |
| description | TEXT | Instructions | How to perform the movement |
| type | VARCHAR(50) | Category | "weightlifting", "cardio", "gymnastics", "bodyweight" |
| is_standard | BOOLEAN | System vs custom | TRUE = pre-seeded by system, FALSE = user-created |
| created_by | BIGINT | Creator | User who created (NULL for standard movements) |
| created_at | TIMESTAMP | Creation time | When movement was added |
| updated_at | TIMESTAMP | Last modification | When definition was updated |

#### workout_movements (Per Requirements: Should be `workout_strength`)
| Field | Type | Purpose | Human Explanation |
|-------|------|---------|-------------------|
| id | BIGINT | Primary key | Unique identifier for this movement in workout |
| workout_id | BIGINT | Parent workout | Which workout this movement belongs to |
| movement_id | BIGINT | Movement definition | Which exercise (FK to movements) |
| weight | DECIMAL(10,2) | Load | Weight used in lbs or kg (NULL if bodyweight) |
| sets | INT | Volume | Number of sets performed |
| reps | INT | Volume | Reps per set or total reps |
| time | INT | Duration | **Work time in seconds** (e.g., 60s plank hold) |
| distance | DECIMAL(10,2) | Distance | For cardio: meters, miles, etc. |
| is_rx | BOOLEAN | Prescribed | TRUE if done as prescribed, FALSE if scaled |
| is_pr | BOOLEAN | Personal record | TRUE if this is a PR for this movement |
| notes | TEXT | Movement notes | Specific notes (e.g., "Tempo: 3-1-1-0") |
| order_index | INT | Sequence | Order this movement appears in workout |
| created_at | TIMESTAMP | Creation time | When added to workout |
| updated_at | TIMESTAMP | Last modification | When details were updated |

**Field Explanation: `time` vs `rest_seconds` vs `work_time`:**
- Database column: `time` (INT, seconds)
- Frontend previously called it: `rest_seconds` (incorrect semantic meaning)
- Now correctly called: `work_time` (how long to perform the movement)
- **Purpose:** Duration for time-based movements (hold plank 60s, row for 120s)
- **NOT** rest time between sets (that's not tracked)

#### refresh_tokens
| Field | Type | Purpose | Human Explanation |
|-------|------|---------|-------------------|
| id | BIGINT | Primary key | Unique token identifier |
| user_id | BIGINT | Token owner | Which user this token belongs to |
| token | VARCHAR(255) | Token value | Cryptographically secure random string |
| expires_at | TIMESTAMP | Expiration | When token becomes invalid (30 days) |
| created_at | TIMESTAMP | Issue time | When token was created |
| revoked_at | TIMESTAMP | Revocation | When token was revoked (logout) |
| device_info | TEXT | Device tracking | Browser/device information |

### Required but Missing Tables

#### wods (Status: ❌ Not Implemented)
| Field | Type | Purpose | Human Explanation |
|-------|------|---------|-------------------|
| id | BIGINT | Primary key | Unique WOD identifier |
| name | VARCHAR(255) | WOD name | "Fran", "Murph", "Helen", etc. |
| source | VARCHAR(100) | Origin | "CrossFit", "Other Coach", "Self-recorded" |
| type | VARCHAR(50) | Category | "Benchmark", "Hero", "Girl", "Games", etc. |
| regime | VARCHAR(50) | Format | "EMOM", "AMRAP", "Fastest Time", "Slowest Round" |
| score_type | VARCHAR(50) | Scoring method | "Time" (MM:SS), "Rounds+Reps", "Max Weight" |
| description | TEXT | Instructions | Full WOD description with movements and scheme |
| url | VARCHAR(512) | Reference link | Video demo or instructions URL |
| notes | TEXT | Additional info | Tips, history, or modifications |
| is_standard | BOOLEAN | System vs custom | TRUE for official CrossFit WODs |
| created_by | BIGINT | Creator | User who created (NULL for standard) |
| created_at | TIMESTAMP | Creation time | When WOD was added |
| updated_at | TIMESTAMP | Last modification | When WOD was updated |

**Example WOD:**
```
Name: "Fran"
Source: "CrossFit"
Type: "Girl"
Regime: "Fastest Time"
Score Type: "Time"
Description: "21-15-9 reps for time of:\n- Thrusters (95 lbs)\n- Pull-ups"
```

#### user_workouts (Status: ❌ Not Implemented)
| Field | Type | Purpose | Human Explanation |
|-------|------|---------|-------------------|
| id | BIGINT | Primary key | Unique identifier for this workout session |
| user_id | BIGINT | Who performed | Which user did this workout (FK to users) |
| workout_id | BIGINT | Template used | Which workout template (FK to workouts) |
| workout_date | DATE | When performed | Date user completed this workout |
| workout_type | VARCHAR(50) | Session type | "strength", "metcon", "wod", "mixed" |
| total_time | INT | Duration | Total time in seconds |
| notes | TEXT | Session notes | How user felt, modifications, RPE, etc. |
| created_at | TIMESTAMP | Log time | When user logged this session |
| updated_at | TIMESTAMP | Last edit | When session log was updated |

**Purpose:** Junction table linking users to workout templates they've performed on specific dates. This allows:
- Same template used by multiple users
- Same user using same template on different dates
- Tracking workout frequency and consistency

#### workout_wods (Status: ❌ Not Implemented)
| Field | Type | Purpose | Human Explanation |
|-------|------|---------|-------------------|
| id | BIGINT | Primary key | Unique identifier |
| workout_id | BIGINT | Parent workout | Which workout template includes this WOD |
| wod_id | BIGINT | WOD definition | Which WOD (FK to wods) |
| score_value | VARCHAR(50) | User's result | Time (12:34), rounds+reps (5+12), weight (225) |
| division | VARCHAR(20) | Scaling level | "Rx", "Scaled", "Beginner" |
| is_pr | BOOLEAN | Personal record | TRUE if this is user's best score |
| order_index | INT | Sequence | Order WOD appears in workout |
| created_at | TIMESTAMP | Creation time | When WOD was added to template |
| updated_at | TIMESTAMP | Last modification | When updated |

**Purpose:** Links workout templates to WODs. A workout can include multiple WODs (e.g., "Fran" + "Helen" in one session).

## Conclusion

Your concern about conceptual mismatch is **absolutely correct**. The current implementation (v0.3.3) uses a **direct logging model** where workouts are user-specific logs, while REQUIIREMENTS.md specifies a **template-based model** where workouts are reusable templates linked to users via a junction table.

The good news is that Migration 0.4.0 has already been written to fix this, but it's a major breaking change that hasn't been applied yet.

**Key Decisions Needed:**

1. **Apply Migration 0.4.0?** - Align implementation with requirements (recommended)
2. **Update Requirements?** - Change requirements to match current simpler implementation
3. **Hybrid Approach?** - Gradually migrate over multiple versions

The field naming issue (`rest_seconds` → `work_time`) has been correctly fixed and is orthogonal to the larger schema design question.
