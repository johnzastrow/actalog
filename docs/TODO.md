# TODO

## v0.7.0-beta Release - COMPLETE ✅ (2025-11-21)

**Status:** Wodify Performance Import system fully implemented and tested with real-world data.

### Completed ✅
- [x] **Wodify Performance Import System**
  - [x] Domain models for Wodify CSV structure (`internal/domain/wodify_import.go`)
  - [x] Result string parser with 9 regex-based parsers (`internal/service/wodify_parser.go`)
  - [x] CSV import service with preview/confirm workflow (`internal/service/wodify_import_service.go`)
  - [x] Date grouping logic to consolidate performances into workouts
  - [x] Auto-entity creation for missing movements and WODs
  - [x] HTTP handler with multipart form-data support (`internal/handler/wodify_import_handler.go`)
  - [x] API routes: `POST /api/import/wodify/preview`, `POST /api/import/wodify/confirm`
  - [x] Frontend integration in ImportView with Wodify-specific preview UI
  - [x] Workout summary table showing dates, counts, component types, and PR flags
  - [x] New entities preview with chips for movements and WODs to be created
  - [x] Success messaging with detailed import statistics
  - [x] Successfully tested with real Wodify export (293 entries, 6+ years of data)
  - [x] Documentation updated in CLAUDE.md with comprehensive implementation details

### Real-World Test Results
- **Input:** Wodify Performance CSV export (68KB, 557 lines, 293 valid rows)
- **Output:** 189 user workouts, 37 movements, 28 WODs, 293 performances, 62 PRs flagged
- **Date Range:** 2018-05-30 to 2025-10-10 (6+ years of workout history)
- **Data Quality:** 1 invalid row handled gracefully, 99.7% success rate
- **Round-trip:** Import → Export verified working correctly

### Bug Fixes
- [x] **User Workouts Import Persistence Bug - RESOLVED**
  - Issue: Reported that workouts don't persist after import
  - Resolution: Testing confirmed feature working correctly
  - Evidence: Database persistence, API retrieval, and export all functional
  - Conclusion: Bug already fixed or false report

---

## v0.6.0-beta Release - COMPLETE ✅ (2025-11-21)

**Status:** Database Backup/Restore system fully functional.

---

## v0.5.1-beta Release - COMPLETE ✅ (2025-11-21)

**Status:** Import/Export system fully functional. All features tested and working correctly.

### Completed ✅
- [x] **WOD Export to CSV**
  - [x] Export endpoint `GET /api/export/wods`
  - [x] Query params: include_standard, include_custom
  - [x] CSV format with all WOD fields
  - [x] Successfully tested - exports all standard WODs
- [x] **Movement Export to CSV**
  - [x] Export endpoint `GET /api/export/movements`
  - [x] Query params: include_standard, include_custom
  - [x] CSV format with all movement fields
  - [x] Successfully tested - exports all standard movements
- [x] **User Workouts Export to JSON**
  - [x] Export endpoint `GET /api/export/user-workouts`
  - [x] Optional date range filtering
  - [x] JSON format with metadata and nested workout data
  - [x] Successfully tested - proper JSON structure
- [x] **WOD Import with Preview and Validation**
  - [x] Preview endpoint `POST /api/import/wods/preview`
  - [x] Confirm endpoint `POST /api/import/wods/confirm`
  - [x] CSV validation (source, type, regime, score_type enums)
  - [x] Duplicate detection
  - [x] Successfully tested - created custom WOD
- [x] **Movement Import with Preview and Validation**
  - [x] Preview endpoint `POST /api/import/movements/preview`
  - [x] Confirm endpoint `POST /api/import/movements/confirm`
  - [x] CSV validation (type enum)
  - [x] Duplicate detection
  - [x] Successfully tested - created 2 custom movements
- [x] **User Workouts Import Preview**
  - [x] Preview endpoint `POST /api/import/user-workouts/preview`
  - [x] JSON parsing and validation
  - [x] Successfully tested - validates workouts
- [x] **Import/Export Frontend Views**
  - [x] `web/src/views/ExportView.vue` - Full export UI
  - [x] `web/src/views/ImportView.vue` - Full import UI with drag-and-drop
  - [x] Routes registered at `/settings/export` and `/settings/import`
  - [x] Fixed axios import to use authenticated instance
- [x] **Backend Services (1,691 lines)**
  - [x] `internal/service/export_service.go` (385 lines)
  - [x] `internal/service/import_service.go` (829 lines)
  - [x] `internal/handler/export_handler.go` (178 lines)
  - [x] `internal/handler/import_handler.go` (299 lines)

### Known Issues ⚠️
- [x] **User Workouts Import Confirm Persistence Bug** - RESOLVED ✅ (2025-11-21)
  - Endpoint: `POST /api/import/user-workouts/confirm`
  - **Status**: Bug could not be reproduced - feature working correctly
  - **Testing**: Imported workout persists to database, appears in `/api/workouts`, and exports correctly
  - **Evidence**:
    - Database: user_workouts table contains imported records
    - API: GET /api/workouts returns imported workouts with correct data
    - Export: Round-trip import → export works perfectly
  - **Conclusion**: Either already fixed in previous session or false report

### Not Implemented (Deferred to Future Versions)
- [ ] Flattened CSV export for workout performance analysis (v0.6.0)
- [ ] Performance data export per WOD/Movement with PR flags (v0.6.0)
- [ ] Export history tracking (v0.7.0)

---

## v0.4.1-beta Release - COMPLETE (2025-01-14)

**Status:** Bug fix release addressing Quick Log and deployment issues.

### Completed ✅
- [x] **Quick Log Movement Search Fix**
  - [x] Added loading states for movements and WODs in Quick Log dialog
  - [x] Added `:loading` prop to autocomplete components
  - [x] Added `auto-select-first` for better search UX
  - [x] Added search icon (magnify) to match design patterns
  - [x] Dialog now opens immediately and fetches data in background
  - [x] Added console logging for debugging data loading
  - [x] Updated `web/src/views/DashboardView.vue:326-342, 420-436`

- [x] **Localhost URL Hardcoding Fixes**
  - [x] Created `web/src/utils/url.js` with dynamic URL resolution utilities
    - [x] `getApiBaseUrl()` - Environment-aware API base URL
    - [x] `getAssetUrl(path)` - Converts relative paths to absolute URLs
    - [x] `getProfileImageUrl(profileImage)` - Handles profile image URLs
  - [x] Updated `web/src/stores/auth.js:22` to use `getProfileImageUrl()`
  - [x] Updated `web/src/views/ProfileView.vue:364` to use `getProfileImageUrl()`
  - [x] Fixed `web/src/views/VerifyEmailView.vue:99` to use relative URLs
  - [x] Added `/uploads` proxy to `web/vite.config.js:164-167`
  - [x] Changed axios baseURL to empty string for Vite proxy in development

- [x] **Documentation & Configuration**
  - [x] Created `web/.env.example` documenting `VITE_API_BASE_URL` variable
  - [x] Updated `CLAUDE.md` with Frontend Configuration section
  - [x] Added deployment guidance for production environments

- [x] **Version Management**
  - [x] Incremented version to 0.4.1-beta in `pkg/version/version.go`
  - [x] Updated version to 0.4.1 in `web/package.json`
  - [x] Updated CHANGELOG.md with v0.4.1-beta release notes
  - [x] Updated README.md version references

### Notes
- Profile pictures and assets now work correctly outside localhost
- Quick Log dialog movement search now functional
- Frontend can be deployed to any domain/IP without hardcoded URLs
- Production deployments should set `VITE_API_BASE_URL` as needed

## v0.4.0-beta Release - PARTIALLY COMPLETE (2025-11-12)

**Status:** Backend complete with seeded data. Frontend stores created. Critical path for testing established.

### Completed (v0.3.1-beta)
- [x] Add `is_pr` column to `workout_movements` table (migration v0.3.0 completed 2025-11-10)
- [x] Multi-database support for `is_pr` field (SQLite, PostgreSQL, MySQL)
- [x] Add `email_verified` and `email_verified_at` columns to `users` table (migration v0.3.1 completed 2025-11-10)
- [x] Create `email_verification_tokens` table with token, user_id, expires_at, used_at

### Schema Changes Required (v0.4.0)
- [ ] Create database migration from v0.3.1 to v0.4.0 - **NEXT PRIORITY**
- [ ] Add `birthday` column to `users` table
- [ ] Create `wods` table with all attributes (name, source, type, regime, score_type, is_standard, etc.)
- [ ] Rename `movements` table to `strength_movements`
- [ ] Add `movement_type` and `is_standard` columns to `strength_movements`
- [ ] Modify `workouts` table (remove user_id, workout_date, workout_type, workout_name, total_time)
- [ ] Create `user_workouts` junction table
- [ ] Rename `workout_movements` to `workout_strength`
- [ ] Create `workout_wods` junction table with `division` and `is_pr` columns
- [ ] Create `user_settings` table
- [ ] Create `audit_logs` table
- [ ] Add `updated_by` columns to all relevant tables
- [ ] Migrate existing data to new schema structure
- [ ] Test migration on development database
- [ ] Create rollback migration script

### Backend Updates for New Schema (v0.4.0) ✅ **COMPLETED 2025-11-10**
- [x] Update domain models for new entities (WOD, Strength, UserWorkout, etc.)
  - [x] `internal/domain/wod.go` - WOD model and WODRepository interface
  - [x] `internal/domain/user_workout.go` - UserWorkout and UserWorkoutWithDetails models
  - [x] `internal/domain/workout_wod.go` - WorkoutWOD junction table model
  - [x] Updated `internal/domain/workout.go` to template-based architecture
- [x] Create repository interfaces and implementations for new entities
  - [x] `internal/repository/wod_repository.go` - WOD data access (Create, Get, List, Update, Delete, Search)
  - [x] `internal/repository/user_workout_repository.go` - User workout instance tracking
  - [x] `internal/repository/workout_wod_repository.go` - Workout-WOD associations
  - [x] Updated `internal/repository/workout_repository.go` for template operations
- [x] Update service layer to work with new schema
  - [x] `internal/service/wod_service.go` - WOD business logic (171 lines)
  - [x] `internal/service/user_workout_service.go` - User workout instance logic (205 lines)
  - [x] `internal/service/workout_wod_service.go` - WOD-workout linking logic (192 lines)
  - [x] Updated `internal/service/workout_service.go` for template operations (382 lines)
- [x] Update API handlers for new data structure
  - [x] `internal/handler/wod_handler.go` - WOD endpoints (247 lines)
  - [x] `internal/handler/user_workout_handler.go` - User workout endpoints (245 lines)
  - [x] `internal/handler/workout_wod_handler.go` - Workout-WOD linking endpoints (219 lines)
  - [x] Deprecated old `workout_handler.go` (incompatible with v0.4.0)
- [x] Wire up new services and handlers in `cmd/actalog/main.go`
  - [x] Repository initialization (userWorkoutRepo, wodRepo, workoutWODRepo)
  - [x] Service initialization (UserWorkoutService, WODService, WorkoutWODService)
  - [x] Handler initialization (userWorkoutHandler, wodHandler, workoutWODHandler)
  - [x] API routes configured for v0.4.0 endpoints
- [x] Add validation for WOD attributes (source, type, regime, score_type)
- [ ] Implement audit logging functionality - **DEFERRED**
- [ ] Create user settings management endpoints - **DEFERRED**

### API Endpoints Implemented (v0.4.0)
**User Workouts** (Log workout instances):
- `POST /api/user-workouts` - Log a workout instance
- `GET /api/user-workouts` - List logged workouts
- `GET /api/user-workouts/{id}` - Get logged workout details
- `PUT /api/user-workouts/{id}` - Update logged workout
- `DELETE /api/user-workouts/{id}` - Delete logged workout
- `GET /api/user-workouts/stats/month` - Monthly workout statistics

**WOD Management**:
- `GET /api/wods` - List all WODs (standard + custom)
- `POST /api/wods` - Create custom WOD
- `GET /api/wods/search` - Search WODs
- `GET /api/wods/{id}` - Get WOD details
- `PUT /api/wods/{id}` - Update custom WOD
- `DELETE /api/wods/{id}` - Delete custom WOD

**Workout-WOD Linking**:
- `POST /api/templates/{id}/wods` - Add WOD to template
- `GET /api/templates/{id}/wods` - List WODs in template
- `PUT /api/templates/{id}/wods/{wod_id}` - Update WOD in template
- `DELETE /api/templates/{id}/wods/{wod_id}` - Remove WOD from template
- `POST /api/templates/{id}/wods/{wod_id}/toggle-pr` - Toggle PR flag

### Seed Data
- [ ] Create seed data for standard CrossFit WODs (Fran, Grace, Helen, Diane, Karen, Murph, DT, etc.)
- [ ] Mark standard WODs with `is_standard = TRUE`
- [ ] Create seed data for standard strength movements
- [ ] Mark standard movements with `is_standard = TRUE`
- [ ] Categorize movements by type (weightlifting, cardio, gymnastics)
- [ ] Add descriptions and URLs for standard WODs

## Design Refinements - HIGH PRIORITY

### Email Verification System

**Status:** ✅ **Completed in v0.3.1-beta** (2025-11-10)

- [x] Implement email verification token generation (crypto/rand, 32 bytes hex)
- [x] Create email verification endpoint (`GET /api/auth/verify-email?token=...`)
- [x] Send verification email on user registration (SMTP with HTML template)
- [x] Add "Resend verification email" functionality (`POST /api/auth/resend-verification`)
- [x] Add verification status indicator in UI (Dashboard warning banner)
- [x] Frontend views: VerifyEmailView, ResendVerificationView
- [x] Updated RegisterView to show verification success message
- [x] Router updates for `/verify-email` and `/resend-verification` routes
- [x] Database migration v0.3.1 with email_verified fields
- [x] Repository methods: `CreateVerificationToken()`, `GetVerificationToken()`, `MarkTokenAsUsed()`
- [x] Service methods: `SendVerificationEmail()`, `VerifyEmailWithToken()`, `ResendVerificationEmail()`
- [ ] Update login to check verification status - Future enhancement (currently soft check)
- [ ] Lock leaderboard participation until verified - Future enhancement
- [ ] Lock data export until verified - Future enhancement

### Personal Records (PR) Tracking

**Status:** ✅ **Completed in v0.3.0-beta** (2025-11-10)

- [x] Implement auto-detection algorithm for PRs:
  - [x] Highest weight for strength movements (per user per movement)
  - [ ] Fastest time for time-based WODs (per user per WOD) - Future enhancement
  - [ ] Most rounds+reps for AMRAP WODs (per user per WOD) - Future enhancement
- [x] Add manual PR flag/unflag endpoints (`POST /api/workouts/movements/:id/toggle-pr`)
- [x] Display PR badges on workout cards in dashboard (gold trophy icons)
- [x] Show PR indicators (🏆) in movement history
- [x] Add PR history view at `/prs` route showing recent PRs and all-time records
- [x] Update PR status when new workout logged (integrated into CreateWorkout workflow)
- [x] API endpoints: `GET /api/workouts/prs`, `GET /api/workouts/pr-movements`
- [x] Repository methods: `GetPersonalRecords()`, `GetMaxWeightForMovement()`, `GetPRMovements()`
- [x] Service layer: `DetectAndFlagPRs()`, authorization checks, PR aggregation

### Leaderboard System - Scaled Divisions
- [ ] Create `leaderboard_entries` table (optional - could query from workout_wods)
- [ ] Implement leaderboard query for each standard WOD
- [ ] Separate leaderboards by division (rx, scaled, beginner)
- [ ] Add division selector when logging WOD scores
- [ ] Display leaderboards on WOD detail screens
- [ ] Implement leaderboard ranking algorithm
- [ ] Add user's rank display on their workouts
- [ ] Filter leaderboards by date range (optional)
- [ ] Admin verification for top entries (future)

### Hybrid Template System
- [ ] Allow users to create custom workout templates
- [ ] Allow users to reuse existing templates when logging
- [ ] Display both standard and custom templates in selectors
- [ ] Add "Save as Template" option when logging workout
- [ ] Implement template management UI (create, edit, delete)
- [ ] Track template usage count
- [ ] Filter templates by custom vs. standard

### Workout Scheduling
- [ ] Add scheduled workout indicator in user_workouts
- [ ] Allow users to schedule workouts for future dates
- [ ] Display scheduled vs. completed workouts differently on calendar
- [ ] Add "Complete Scheduled Workout" flow
- [ ] Prevent scheduling in the past (validation)

### Import/Export System (v0.5.1-beta) - MOSTLY COMPLETE ✅

**Priority:** HIGH
**Status:** Implementation complete, one critical bug pending
**Target Version:** v0.5.1-beta
**Testing Status:** 5.5/6 features working (92%)

#### Requirements Summary
From REQUIREMENTS.md lines 864-870, the import/export system must support:
1. **Round-trip CSV** for WODs and Movements (import what you export)
2. **User Workouts** with full context (JSON format with nested data)
3. **Flattened CSV** for workout performance analysis in spreadsheets
4. **Performance data** export for each WOD and Movement with PR flags
5. **Admin vs User permissions** - admins can export global data, users only their own
6. **Import preview** before confirmation
7. **Date range selector** for partial exports
8. **Export history** tracking

#### Phase 1: CSV Export/Import for WODs and Movements ✅ COMPLETE

**CSV Schema Design:**

**WODs CSV Format:**
```csv
name,source,type,regime,score_type,description,url,notes,is_standard,created_by_email
Fran,CrossFit,Girl,Fastest Time,Time (HH:MM:SS),"21-15-9 reps for time of: Thrusters (95/65 lb), Pull-ups",https://www.crossfit.com/workout/fran,,true,
```

**Movements CSV Format:**
```csv
name,type,description,is_standard,created_by_email
Back Squat,weightlifting,Barbell back squat,true,
Custom Movement,weightlifting,My custom exercise,false,user@example.com
```

**Backend Tasks:**
- [x] Create `internal/service/export_service.go` - CSV export business logic ✅ (385 lines)
  - [x] `ExportWODsToCSV(userID, isAdmin)` - Export WODs with permission checks ✅
  - [x] `ExportMovementsToCSV(userID, isAdmin)` - Export movements with permission checks ✅
  - [x] Handle standard vs custom WODs/movements filtering ✅
  - [x] Include created_by_email for custom entries ✅
  - [x] Support admin export (all data) vs user export (own data only) ✅

- [x] Create `internal/service/import_service.go` - CSV import business logic with validation ✅ (829 lines)
  - [x] `ImportWODsFromCSV(userID, isAdmin, csvData)` - Import WODs with moderate validation ✅
    - [x] Parse CSV and validate headers ✅
    - [x] Validate required fields (name, source, type, regime, score_type) ✅
    - [x] Check for duplicate names (skip or update based on flag) ✅
    - [x] Validate enum values (source, type, regime, score_type against domain constants) ✅
    - [x] Foreign key validation (created_by_email exists) ✅
    - [x] Data type validation (booleans, strings) ✅
    - [x] Return preview data before actual import ✅
  - [x] `ImportMovementsFromCSV(userID, isAdmin, csvData)` - Import movements with moderate validation ✅
    - [x] Parse CSV and validate headers ✅
    - [x] Validate required fields (name, type) ✅
    - [x] Check for duplicate names (skip or update) ✅
    - [x] Validate type enum (weightlifting, cardio, gymnastics, bodyweight) ✅
    - [x] Foreign key validation (created_by_email exists) ✅
    - [x] Return preview data before actual import ✅
  - [x] `ValidateCSVFormat(csvData, entityType)` - Common validation logic ✅
  - [x] Permission checks (users cannot import as standard, admins can) ✅

- [x] Create `internal/handler/export_handler.go` - HTTP handlers for export ✅ (178 lines)
  - [x] `GET /api/export/wods` - Export WODs to CSV ✅
    - Query params: `include_standard=true`, `include_custom=true` ✅
    - Response: CSV file download with Content-Disposition header ✅
    - Authorization: All users can export (filtered by ownership) ✅
  - [x] `GET /api/export/movements` - Export movements to CSV ✅
    - Query params: `include_standard=true`, `include_custom=true` ✅
    - Response: CSV file download ✅
    - Authorization: All users can export (filtered by ownership) ✅

- [x] Create `internal/handler/import_handler.go` - HTTP handlers for import ✅ (299 lines)
  - [x] `POST /api/import/wods/preview` - Preview WOD import before committing ✅
    - Request: multipart/form-data with CSV file ✅
    - Response: JSON with parsed data, validation errors, counts (total, valid, invalid, duplicates) ✅
  - [x] `POST /api/import/wods/confirm` - Commit WOD import after preview approval ✅
    - Request: multipart/form-data with file and options (skip_duplicates, update_duplicates) ✅
    - Response: JSON with import result (created, updated, skipped counts) ✅
  - [x] `POST /api/import/movements/preview` - Preview movement import ✅
  - [x] `POST /api/import/movements/confirm` - Commit movement import ✅
  - [x] Rate limiting for import endpoints (prevent abuse) ✅
  - [x] File size limits (max 10MB CSV) ✅

- [x] Update `cmd/actalog/main.go` - Wire up new services and routes ✅
  - [x] Initialize ExportService and ImportService ✅ (lines 180-181)
  - [x] Initialize ExportHandler and ImportHandler ✅ (lines 198-199)
  - [x] Register export routes (GET /api/export/wods, /api/export/movements) ✅ (lines 335-337)
  - [x] Register import routes (POST /api/import/{entity}/preview, /api/import/{entity}/confirm) ✅ (lines 340-345)

**Frontend Tasks:**
- [x] Create `web/src/views/ExportView.vue` - Export data screen ✅
  - [x] Data type selector (WODs, Movements, User Workouts) ✅
  - [x] Format handling (CSV for WODs/Movements, JSON for User Workouts) ✅
  - [x] Options: Include standard items, Include custom items ✅
  - [x] Export button triggers download ✅
  - [ ] Export history section (future - tracks past exports)
  - [x] Route: `/settings/export` ✅

- [x] Create `web/src/views/ImportView.vue` - Import data screen ✅
  - [x] File upload dropzone (drag & drop support) ✅
  - [x] Supported formats info (CSV, JSON) ✅
  - [x] Preview table showing parsed data with validation status ✅
  - [x] Validation errors display (red highlights for invalid rows) ✅
  - [x] Import statistics (total, valid, invalid, duplicates) ✅
  - [x] Import options: Skip duplicates, Update duplicates, Create only new ✅
  - [x] Confirm import button (after preview) ✅
  - [x] Cancel button ✅
  - [x] Route: `/settings/import` ✅

- [x] Update `web/src/router/index.js` - Add new routes ✅
  - [x] `/settings/import` → ImportView ✅ (line 157)
  - [x] `/settings/export` → ExportView ✅ (line 152)

- [x] Navigation accessible from Profile menu ✅
  - [x] "Import Data" accessible ✅
  - [x] "Export Data" accessible ✅

**Testing Tasks:**
- [x] Manual testing for ExportService ✅
  - [x] Test WOD export with standard and custom ✅
  - [x] Test Movement export with standard and custom ✅
  - [x] Test User Workouts export ✅
  - [x] Test authenticated endpoints ✅

- [x] Manual testing for ImportService ✅
  - [x] Test CSV parsing with valid data ✅
  - [x] Test CSV parsing with invalid headers ✅
  - [x] Test duplicate detection ✅
  - [x] Test enum validation for WOD fields ✅
  - [x] Test WOD import (created "My Custom Fran") ✅
  - [x] Test Movement import (created 2 custom movements) ✅

- [x] Manual end-to-end testing for export/import ✅
  - [x] Test WOD round-trip (export CSV, import same CSV) ✅
  - [x] Test Movement round-trip ✅
  - [x] Test import with validation errors ✅
  - [x] Test import preview workflow ✅
  - [x] Verified imported data persists in database ✅

**Migration:**
- [x] No database changes needed for Phase 1 (using existing v0.5.0 schema) ✅

#### Phase 2: User Workouts Export/Import (JSON) - PARTIAL ⚠️

**Requirements:**
- [x] Export user workouts with full nested data (workouts, movements, WODs, scores) ✅
- [x] Include all performance data (weights, reps, times, PR flags) ✅
- [x] Support date range filtering ✅
- [x] JSON format for complete data structure ✅
- [x] Import preview with validation ✅
- [~] Import confirm with conflict resolution ⚠️ HAS PERSISTENCE BUG

**JSON Schema Design:**
```json
{
  "export_metadata": {
    "user_email": "user@example.com",
    "export_date": "2025-11-21T10:00:00Z",
    "date_range": {"start": "2025-01-01", "end": "2025-11-21"},
    "version": "0.5.1"
  },
  "user_workouts": [
    {
      "workout_date": "2025-11-20",
      "workout_type": "metcon",
      "notes": "Felt great today",
      "wods": [
        {
          "wod_name": "Fran",
          "score_type": "Time (HH:MM:SS)",
          "score_value": "00:05:43",
          "is_pr": true
        }
      ],
      "movements": [
        {
          "movement_name": "Back Squat",
          "sets": 5,
          "reps": 5,
          "weight": 225,
          "is_pr": false
        }
      ]
    }
  ]
}
```

**Backend Tasks:**
- [x] `ExportUserWorkoutsToJSON(userID, startDate, endDate)` - Export workouts with nested data ✅
- [x] `ImportUserWorkoutsFromJSON(userID, jsonData)` - Import with conflict resolution ✅
- [x] Handle nested relationships (user_workouts → user_workout_movements, user_workout_wods) ✅
- [x] Validate foreign key references (WOD names, movement names must exist or be created) ✅
- [x] Duplicate detection by workout_date ✅
- [ ] **Fix persistence bug in import confirm** - CRITICAL

**Frontend Tasks:**
- [x] Update ExportView with date range picker ✅
- [x] Update ExportView with JSON format option ✅
- [x] Update ImportView to handle JSON uploads ✅
- [x] Display nested data preview for JSON imports ✅
- [x] Conflict resolution UI (skip duplicates option) ✅

#### Phase 3: Performance Analytics Export (CSV) - FUTURE

**Requirements:**
- [ ] Export flattened CSV for spreadsheet analysis
- [ ] One row per performance record (movement or WOD)
- [ ] Include PR flags, user info, workout date
- [ ] Suitable for pivot tables and charts in Excel/Sheets

**CSV Schema:**
```csv
date,workout_type,entity_type,entity_name,metric_type,metric_value,is_pr,notes
2025-11-20,strength,movement,Back Squat,weight,225,false,5x5 progressive
2025-11-20,metcon,wod,Fran,time,00:05:43,true,New PR!
```

**Backend Tasks:**
- [ ] `ExportPerformanceDataToCSV(userID, startDate, endDate, entityType)` - Flatten performance
- [ ] Support filtering by WOD or Movement
- [ ] Include calculated metrics (volume = weight × reps × sets)

#### Phase 4: Markdown Export for Workout Reports - FUTURE

**Requirements:**
- [ ] Format workouts as readable Markdown documents
- [ ] Include workout details, scores, notes
- [ ] Support rich formatting (bold, lists, links)
- [ ] Suitable for blogs, social sharing, personal journaling

**Markdown Template:**
```markdown
# Workout - November 20, 2025

## WODs
### Fran ⭐ PR
**Score:** 5:43 (Time)
**Notes:** New personal record!

## Strength Work
### Back Squat
**Weight:** 225 lbs
**Sets:** 5 × 5 reps

## Workout Notes
Felt great today. Progressive overload working well.
```

**Backend Tasks:**
- [ ] `ExportWorkoutToMarkdown(workoutID)` - Format single workout
- [ ] `ExportWorkoutsToMarkdown(userID, startDate, endDate)` - Format multiple workouts
- [ ] Template system for customizable formatting

#### Phase 5: Export History Tracking - FUTURE

**Requirements:**
- [ ] Track all exports in `export_history` table
- [ ] Display export history in UI with download links
- [ ] Auto-cleanup old exports after 30 days
- [ ] Support re-download of previous exports

**Database Changes:**
- [ ] Create `export_history` table (id, user_id, entity_type, format, file_path, created_at)

**Backend Tasks:**
- [ ] Save export files to disk or S3
- [ ] Create export history records
- [ ] Auto-cleanup job (cron)

**Frontend Tasks:**
- [ ] Display export history list in ExportView
- [ ] Download button for past exports
- [ ] Delete button for manual cleanup

---

**Current Focus:** Phase 1 (CSV export/import for WODs and Movements)
**Target Completion:** v0.5.1-beta
**Estimated Effort:** 2-3 days (backend + frontend + testing)

**Dependencies:**
- All v0.5.0 features must be stable
- Existing WOD and Movement repositories functional
- Frontend router and Settings menu accessible

## PWA Features (v0.2.0)

### Completed ✅
- [x] Configure vite-plugin-pwa
- [x] Create web app manifest
- [x] Set up service worker with Workbox
- [x] Implement IndexedDB offline storage
- [x] Add background sync queue
- [x] Service worker registration
- [x] Auto-update notification system
- [x] PWA documentation (DEPLOYMENT.md)
- [x] PWA development setup in SETUP.md

### Remaining PWA Tasks
- [ ] Generate all PWA icon sizes (72px - 512px)
- [ ] Create apple-touch-icon.png for iOS
- [ ] Test offline workout creation
- [ ] Test background sync functionality
- [ ] Implement offline indicator in UI
- [ ] Add sync status indicator
- [ ] Test install prompt on all platforms
- [ ] Run Lighthouse PWA audit
- [ ] Optimize service worker cache size

## High Priority

### Authentication & User Management
- [x] Implement password reset functionality ✅ **Completed in v0.3.0-beta** (Parts 1-3: DB, backend, frontend)
- [x] Add email verification for new users ✅ **Completed in v0.3.1-beta** (see Design Refinements section)
- [x] Add profile picture upload ✅ **Completed in v0.3.2-beta** (Avatar upload with initials fallback)
- [ ] Implement "Remember Me" functionality - **NEXT PRIORITY**
- [ ] Add user profile editing with birthday field
- [ ] Add user management for admins:
  - [ ] Admin dashboard for user management
  - [ ] List all users with pagination and search
  - [ ] View user details (profile, stats, activity)
  - [ ] Edit user roles (admin, user)
  - [ ] Disable/enable user accounts
  - [ ] Delete user accounts (with confirmation)
  - [ ] View user workout history
  - [ ] Reset user passwords (admin action)
  - [ ] Manage user permissions
  - [ ] Bulk user actions (export, disable, etc.)

### Workout Logging (Planned for v0.3.0 Schema - Not Yet Implemented)
- [ ] Implement workout template creation API endpoints
- [ ] Implement user_workout logging endpoints (link user to workout on specific date)
- [ ] Add WOD creation/editing for custom WODs
- [ ] Add strength movement creation/editing for custom movements
- [ ] Implement workout_wod association endpoints
- [ ] Implement workout_strength association endpoints with weight/sets/reps
- [ ] Implement workout history retrieval (via user_workouts)
- [ ] Add workout template editing and deletion
- [ ] Add workout search and filtering
- [ ] Implement PR (Personal Record) tracking across user_workouts
- [ ] Add scoring for WODs (time, rounds+reps, max weight)

### Movement Database
- [ ] Seed database with standard CrossFit movements
- [ ] Add movement categories (Weightlifting, Gymnastics, etc.)
- [ ] Implement movement search functionality
- [ ] Add movement details and instructions
- [ ] Support for custom movements per user

### Progress Tracking
- [ ] Implement data aggregation for charts
- [ ] Add progress by movement endpoint
- [ ] Add progress by date range endpoint
- [ ] Calculate and display PRs
- [ ] Add workout frequency analytics

## Medium Priority

### Data Import/Export
- [ ] Implement CSV export for workouts
- [ ] Implement JSON export for workouts
- [ ] Add CSV import functionality
- [ ] Add JSON import functionality
- [ ] Validate imported data

### Admin Features
- [ ] Admin dashboard
- [ ] User management interface
- [ ] System settings management
- [ ] User activity monitoring

### Database Backup and Restore System (v0.6.0-beta) - HIGH PRIORITY

**Status:** Not started
**Priority:** High - Critical for disaster recovery and data migration
**Target Version:** v0.6.0-beta

#### Overview
Complete backup and restore system allowing administrators to create full database backups and restore them when needed. Supports disaster recovery, data migration between installations, and changing database technologies.

#### Phase 1: Full Backup and Restore (v0.6.0)

**Backend Implementation:**

- [ ] Create `internal/service/backup_service.go`
  - [ ] `CreateBackup(adminUserID int64)` - Creates full backup
    - [ ] Export all database tables to JSON format
    - [ ] Include metadata (date, version, admin, record counts, date range)
    - [ ] Copy all uploaded files from `uploads/` directory
    - [ ] Copy `.env` configuration (sanitized, no secrets in plain text)
    - [ ] Package everything into `.zip` file
    - [ ] Store in `backups/` directory with timestamp filename
  - [ ] `CreateSQLiteBackup(adminUserID int64)` - Direct SQLite file copy (if active database is SQLite)
  - [ ] `ListBackups()` - Returns list of available backups with metadata
    - [ ] Read all `.zip` files from `backups/` directory
    - [ ] Extract and parse `metadata.json` from each backup
    - [ ] Return sorted list (newest first) with file size
  - [ ] `GetBackupMetadata(filename string)` - Reads metadata from specific backup
  - [ ] `DeleteBackup(filename string, adminUserID int64)` - Deletes backup file
    - [ ] Audit log entry for deletion
    - [ ] Validate file exists before deletion
  - [ ] `RestoreBackup(filename string, adminUserID int64)` - Full restore from backup
    - [ ] Validate backup file integrity
    - [ ] Parse metadata and check compatibility (schema version)
    - [ ] Wipe all existing database tables (with confirmation)
    - [ ] Parse JSON and insert all data into current database driver
    - [ ] Restore uploaded files to `uploads/` directory
    - [ ] Audit log entry for restore
    - [ ] Return summary (records restored, files restored)

- [ ] Create `internal/handler/backup_handler.go`
  - [ ] `POST /api/admin/backups` - Create new backup
    - [ ] Authorization: Admin only
    - [ ] Async operation with progress tracking (optional)
    - [ ] Returns backup filename and metadata
  - [ ] `GET /api/admin/backups` - List all backups
    - [ ] Authorization: Admin only
    - [ ] Returns array of backup metadata
  - [ ] `GET /api/admin/backups/{filename}` - Download specific backup
    - [ ] Authorization: Admin only
    - [ ] Streams `.zip` file to client
    - [ ] Content-Disposition: attachment header
  - [ ] `GET /api/admin/backups/{filename}/metadata` - Get backup metadata only
    - [ ] Authorization: Admin only
  - [ ] `DELETE /api/admin/backups/{filename}` - Delete backup
    - [ ] Authorization: Admin only
    - [ ] Audit log entry
  - [ ] `POST /api/admin/backups/{filename}/restore` - Restore from backup
    - [ ] Authorization: Admin only
    - [ ] Requires confirmation token (prevent accidental wipe)
    - [ ] Returns restore summary

- [ ] Update `cmd/actalog/main.go`
  - [ ] Initialize BackupService with all repository dependencies
  - [ ] Initialize BackupHandler
  - [ ] Register backup routes under `/api/admin` with AdminOnly middleware

- [ ] Create `backups/` directory structure
  - [ ] Add `.gitignore` entry for `backups/` (don't commit backups to git)
  - [ ] Create directory on server startup if not exists
  - [ ] Add permission checks (writable by application)

**Frontend Implementation:**

- [ ] Create `web/src/views/AdminBackupsView.vue`
  - [ ] Route: `/admin/backups`
  - [ ] **Create Backup Section:**
    - [ ] "Create Backup" button
    - [ ] Loading indicator during backup creation
    - [ ] Success message with backup filename
    - [ ] Error handling with user-friendly messages
  - [ ] **Backup List Section:**
    - [ ] Table displaying all backups
    - [ ] Columns: Date, Size, Data Range, Created By, Actions
    - [ ] Sort by date (newest first)
    - [ ] **Actions per row:**
      - [ ] Download button (downloads .zip file)
      - [ ] Delete button (with confirmation dialog)
      - [ ] Restore button (with strong confirmation dialog)
  - [ ] **Restore Confirmation Dialog:**
    - [ ] Warning message: "This will DELETE all current data"
    - [ ] Checkbox: "I understand this action cannot be undone"
    - [ ] Text input: Type "RESTORE" to confirm
    - [ ] Cancel and Confirm buttons
  - [ ] **Delete Confirmation Dialog:**
    - [ ] Warning message: "Delete backup permanently?"
    - [ ] Cancel and Confirm buttons
  - [ ] **Empty State:**
    - [ ] Message: "No backups found"
    - [ ] "Create your first backup" button

- [ ] Update `web/src/router/index.js`
  - [ ] Add route: `/admin/backups` → AdminBackupsView
  - [ ] Requires authentication and admin role

- [ ] Update navigation (Profile or Admin section)
  - [ ] Add "Database Backups" link for admins
  - [ ] Icon: mdi-database-export or mdi-backup-restore

**Testing:**

- [ ] Unit tests for BackupService
  - [ ] Test backup creation (all tables, files, metadata)
  - [ ] Test SQLite file backup
  - [ ] Test backup listing and sorting
  - [ ] Test backup deletion with audit log
  - [ ] Test metadata extraction
  - [ ] Test restore validation (schema version check)
  - [ ] Test restore with data integrity checks

- [ ] Integration tests for backup/restore workflow
  - [ ] Create backup, verify file exists
  - [ ] Download backup, verify ZIP structure
  - [ ] Restore backup, verify all data restored correctly
  - [ ] Test restore with different database driver (SQLite → PostgreSQL)
  - [ ] Test error scenarios (corrupted backup, insufficient permissions)

- [ ] Manual testing checklist:
  - [ ] Create backup with real data
  - [ ] Download and inspect ZIP contents
  - [ ] Restore backup on clean database
  - [ ] Verify all users, workouts, movements, WODs restored
  - [ ] Verify uploaded files (avatars) restored
  - [ ] Test delete backup
  - [ ] Test admin-only access control

**Database Changes:**
- [ ] No new tables needed (uses existing data)
- [ ] Ensure `backups/` directory has proper file permissions

**JSON Backup Structure:**
```json
{
  "metadata": {
    "backup_version": "1.0",
    "app_version": "0.6.0",
    "created_at": "2025-01-21T10:30:00Z",
    "database_driver": "sqlite3",
    "created_by": "admin@example.com",
    "data_date_range": {
      "earliest_workout": "2024-01-01",
      "latest_workout": "2025-01-21"
    },
    "record_counts": {
      "users": 150,
      "workouts": 3420,
      "movements": 45,
      "wods": 120,
      "audit_logs": 8950
    },
    "schema_version": "0.5.0"
  },
  "users": [...],
  "movements": [...],
  "wods": [...],
  "workouts": [...],
  "user_workouts": [...],
  "user_workout_movements": [...],
  "user_workout_wods": [...],
  "audit_logs": [...],
  "user_settings": [...]
}
```

**ZIP File Structure:**
```
actalog_backup_2025-01-21_10-30-00.zip
├── metadata.json
├── database.json (all tables)
├── database.db (optional, if SQLite)
├── config/
│   └── .env.backup
└── uploads/
    ├── avatars/
    └── attachments/
```

**Dependencies:**
- Go standard library: `archive/zip` for ZIP compression
- All existing repositories for data access
- File system utilities

**Future Enhancements (v0.7.0+):**
- [ ] Selective restore (choose specific tables or date ranges)
- [ ] Merge mode restore (import without wiping existing data)
- [ ] Automated scheduled backups (daily, weekly, monthly)
- [ ] Backup retention policy (auto-delete old backups)
- [ ] Encryption (password-protected backups)
- [ ] Remote storage (S3, Google Drive, Dropbox)
- [ ] Point-in-time recovery with transaction logs
- [ ] CLI tool for backup/restore operations
- [ ] Email notifications for backup completion/failure
- [ ] Incremental backups (only changed data since last full backup)

**Notes:**
- Phase 1 focuses on full restore only (wipe and replace)
- All duplicate handling uses overwrite strategy in Phase 1
- No preview functionality in Phase 1 (too complex)
- Manual backups only (no scheduling in Phase 1)
- Admin-only feature (regular users cannot access)
- Works with all supported database drivers (SQLite, PostgreSQL, MySQL)

### Frontend Enhancements
- [ ] Connect all views to backend APIs
- [ ] Add loading states and error handling
- [ ] Implement data caching with Pinia and IndexedDB
- [x] Add offline support (PWA) - v0.2.0
- [ ] Add pull-to-refresh on mobile (can use PWA techniques)
- [ ] Integrate offline storage with workout forms
- [ ] Show network status indicator
- [ ] Display sync status for pending workouts

### Testing (v0.4.0) - IN PROGRESS

**Status:** Unit test infrastructure created. UserWorkoutService tests completed (68% pass rate). Additional service tests in progress.

#### Completed ✅
- [x] Create shared test helpers (`internal/service/test_helpers.go`)
  - [x] Mock UserWorkoutRepository with full interface implementation
  - [x] Mock WorkoutRepository with full interface implementation
  - [x] Mock WorkoutMovementRepository with full interface implementation
  - [x] Helper functions for pointer types (stringPtr, intPtr, int64Ptr)
- [x] UserWorkoutService unit tests (`internal/service/user_workout_service_test.go`)
  - [x] TestUserWorkoutService_LogWorkout (4 test cases) - 4/4 passing ✅
  - [x] TestUserWorkoutService_GetLoggedWorkout (3 test cases) - 1/3 passing (error wrapping issue)
  - [x] TestUserWorkoutService_UpdateLoggedWorkout (3 test cases) - 2/3 passing (error wrapping issue)
  - [x] TestUserWorkoutService_DeleteLoggedWorkout (3 test cases) - 2/3 passing (error wrapping issue)
  - [x] TestUserWorkoutService_GetWorkoutStatsForMonth (2 test cases) - 2/2 passing ✅
  - **Overall: 11/16 tests passing (68%)**
  - Known issue: Error comparison needs `errors.Is()` for wrapped errors

#### In Progress 🔄
- [ ] WODService unit tests
  - [ ] TestWODService_CreateWOD
  - [ ] TestWODService_GetWOD
  - [ ] TestWODService_ListWODs
  - [ ] TestWODService_UpdateWOD
  - [ ] TestWODService_DeleteWOD
  - [ ] TestWODService_SearchWODs
- [ ] WorkoutWODService unit tests
  - [ ] TestWorkoutWODService_AddWODToWorkout
  - [ ] TestWorkoutWODService_RemoveWODFromWorkout
  - [ ] TestWorkoutWODService_UpdateWorkoutWOD
  - [ ] TestWorkoutWODService_ToggleWODPR
  - [ ] TestWorkoutWODService_ListWODsForWorkout
- [ ] WorkoutService template operation tests
  - [ ] TestWorkoutService_CreateTemplate
  - [ ] TestWorkoutService_GetTemplate
  - [ ] TestWorkoutService_ListTemplates
  - [ ] TestWorkoutService_UpdateTemplate
  - [ ] TestWorkoutService_DeleteTemplate

#### Pending ⏳
- [ ] Fix error wrapping in existing tests (use `errors.Is()` instead of direct comparison)
- [ ] Write unit tests for repositories
- [ ] Write integration tests for v0.4.0 API endpoints
  - [ ] user_workout_handler integration tests
  - [ ] wod_handler integration tests
  - [ ] workout_wod_handler integration tests
- [ ] Add frontend component tests
- [ ] Set up CI/CD pipeline
- [ ] Achieve >80% test coverage target

#### Test Files Created
- `internal/service/test_helpers.go` - Shared mock repositories (334 lines)
- `internal/service/user_workout_service_test.go` - UserWorkoutService tests (483 lines)
- `internal/service/workout_service_test.go.old` - Deprecated v0.3.x tests (renamed)

#### Technical Notes
- Tests use table-driven test pattern for multiple scenarios
- Mock repositories fully implement domain interfaces
- Authorization checks tested (user ownership, standard vs custom resources)
- Edge cases covered (not found, unauthorized, validation failures)
- Error handling paths tested for all service methods

## Low Priority

### Performance
- [ ] Add database query optimization
- [x] Implement PWA caching (service worker) - v0.2.0
- [ ] Add Redis for session storage
- [ ] Optimize frontend bundle size
- [ ] Add lazy loading for images
- [x] Precache static assets - v0.2.0
- [x] Implement code splitting preparation - v0.2.0

### Social Features
- [ ] Add workout sharing (Web Share API)
- [x] Add leaderboards (moved to HIGH PRIORITY - Design Refinements)
- [ ] Add workout comments (future)
- [ ] Add friend system (future - not in current scope)
- [x] Add workout templates (moved to HIGH PRIORITY - Hybrid Template System)

### Notifications
- [ ] Implement email notifications
- [ ] Add in-app notifications
- [ ] Add workout reminders via push notifications (PWA)
- [ ] Add achievement notifications
- [ ] Implement Web Push API for PWA notifications
- [ ] Add notification preferences in settings

### Documentation
- [ ] Complete API documentation
- [ ] Add user guide
- [ ] Create developer setup guide
- [ ] Add deployment guide
- [ ] Create video tutorials

## Future Considerations

- [x] Progressive Web App (completed v0.2.0)
- [ ] Advanced PWA features:
  - [ ] Periodic background sync for data refresh
  - [ ] Web Share API for workout sharing
  - [ ] File System Access API for bulk operations
  - [ ] Badging API for unsynced notifications
- [ ] Mobile native apps (iOS/Android) - may not be needed with PWA
- [ ] Apple Watch integration
- [ ] Wearable device sync
- [ ] Nutrition tracking
- [ ] Workout planning/programming
- [ ] Coach/athlete relationship features
- [ ] Gym/box management features
- [ ] Payment/subscription system
- [ ] Multi-language support

## Technical Debt

### Database & Performance
- [ ] **Migrate from lib/pq to pgx for PostgreSQL support** - HIGH PRIORITY
  - Current: Using `github.com/lib/pq` (maintenance mode, no new features)
  - Target: Migrate to `github.com/jackc/pgx/v5` (actively maintained, better performance)
  - Benefits:
    - Better connection pooling
    - Native support for PostgreSQL types
    - Improved performance (binary protocol)
    - Better prepared statement caching
    - Active maintenance and security updates
  - Migration Steps:
    1. Add pgx/v5 dependency: `go get github.com/jackc/pgx/v5`
    2. Update database connection string format
    3. Replace `database/sql` + `lib/pq` with `pgx.Pool`
    4. Update repository implementations for pgx-specific APIs
    5. Test all database operations
    6. Update connection pooling configuration
    7. Performance benchmark before/after
- [ ] Add comprehensive error handling
- [ ] Improve logging with structured logging
- [ ] Add request rate limiting
- [ ] Implement API versioning
- [ ] Add database migrations system
- [ ] Set up monitoring and alerting
- [ ] Add security headers
- [ ] Implement CSRF protection
- [ ] Clean up old service worker caches
- [ ] Implement PWA update strategy testing

## Deployment Tasks

- [ ] Set up production HTTPS (Let's Encrypt)
- [ ] Configure Nginx for PWA (see DEPLOYMENT.md)
- [ ] Generate production PWA icons
- [ ] Test PWA install on all platforms
- [ ] Set up automated backups
- [ ] Configure monitoring and alerting
- [ ] Set up SSL auto-renewal
- [ ] Performance testing and optimization
- [ ] Security audit

---

**Last Updated:** 2025-01-14
**Version:** 0.4.1-beta

**v0.4.1-beta Status:**
- ✅ Quick Log movement search fixed
- ✅ Localhost URL hardcoding resolved
- ✅ Production deployment support added
- ✅ All documentation updated

**v0.4.0 Status (Previous Release):**
- ✅ Domain models updated for template architecture
- ✅ Repositories implemented (UserWorkout, WOD, WorkoutWOD)
- ✅ Services implemented (UserWorkoutService, WODService, WorkoutWODService, updated WorkoutService)
- ✅ Handlers created (user_workout_handler, wod_handler, workout_wod_handler)
- ✅ API routes configured in main.go
- ✅ Application compiles successfully
- 🔄 Unit tests in progress (UserWorkoutService: 11/16 passing)
- ⏳ Database migration not yet applied (still at v0.3.1 schema)
- ⏳ Frontend updates pending
