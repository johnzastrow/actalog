# Changelog

All notable changes to ActaLog will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.0-beta] - 2025-11-21

### Added
- **Wodify Performance Import System**
  - **Backend Components**
    - **Domain Models** (`internal/domain/wodify_import.go`)
      - `WodifyPerformanceRow` - Represents CSV row from Wodify export (19 columns)
      - `WodifyImportPreview` - Preview statistics and validation results
      - `WodifyImportResult` - Import completion statistics
      - `ParsedPerformanceResult` - Structured performance data after parsing
    - **Result Parser** (`internal/service/wodify_parser.go` - 273 lines)
      - 9 regex-based parsers for different result types:
        - `Weight` - Parses "3 x 10 @ 85 lbs" → sets, reps, weight
        - `Time` - Parses "5:30" (MM:SS) or "1:05:30" (HH:MM:SS) → seconds
        - `AMRAP - Rounds and Reps` - Parses "7 + 3" → rounds, reps
        - `AMRAP - Reps` - Parses "50 Reps"
        - `AMRAP - Rounds` - Parses "5 Rounds"
        - `Max reps` - Parses "3 x 8" (sets x reps)
        - `Calories` - Parses "133 Calories"
        - `Distance` - Parses "500 m"
        - `Each Round` - Parses "175 Total Reps"
      - `ParseDate()` - Handles MM/DD/YYYY and MM/DD/YY formats
      - `DetermineMovementType()` - Maps component type to movement type
      - `DetermineWODScoreType()` - Maps result type to WOD score type
    - **Import Service** (`internal/service/wodify_import_service.go` - 582 lines)
      - `PreviewImport()` - Analyzes CSV and returns preview without database changes
      - `ConfirmImport()` - Executes import with entity auto-creation
      - `parseCSV()` - Handles 19-column CSV with multi-line field support
      - `groupByDate()` - Groups performances by workout date
      - `getOrCreateMovement()` - Auto-creates missing movements
      - `getOrCreateWOD()` - Auto-creates missing WODs
      - `importWorkout()` - Creates UserWorkout with linked performances
      - PR preservation from Wodify export
    - **HTTP Handler** (`internal/handler/wodify_import_handler.go` - 107 lines)
      - `POST /api/import/wodify/preview` - Preview Wodify CSV import
      - `POST /api/import/wodify/confirm` - Execute Wodify CSV import
      - File size limit: 10MB
      - Multipart form-data with "file" field
  - **Frontend Integration** (`web/src/views/ImportView.vue`)
    - Added "Wodify Performance" import type with file-chart icon
    - **Wodify-Specific Preview UI:**
      - Summary stats: total rows, valid rows, workout dates, entities to create
      - New Entities card with chips showing movements and WODs to auto-create
      - Workout Summary table: date, movement count, WOD count, component types, PR flags
      - Gold trophy icons for workouts containing PRs
    - **Success Message:**
      - Displays: workouts created, performances, movements/WODs auto-created, PRs flagged
      - Format: "Workouts Created: 189 | Performances: 293 | Movements Created: 37 | WODs Created: 28 | PRs Flagged: 62"
  - **Documentation**
    - Updated `CLAUDE.md` with comprehensive Wodify import documentation
    - Real-world test results with 6+ years of data
    - Code examples for API usage
    - Domain model definitions

### Technical
- Clean Architecture maintained: domain → service → handler pattern
- CSV parsing with LazyQuotes and TrimLeadingSpace for robust handling
- Regex-based result string parsing for data extraction
- Date grouping logic to create cohesive UserWorkout entries
- Auto-entity creation reduces manual data entry
- PR flag preservation from source data
- Build number incremented: 27 → 28

### Testing
- ✅ Preview import: Analyzed 293 performance entries, 189 unique dates
- ✅ Confirm import: Successfully imported 6+ years of workout history (2018-2025)
  - 189 user workouts created (grouped by date)
  - 37 new movements auto-created
  - 28 new WODs auto-created
  - 293 performance entries
  - 62 PRs automatically flagged
- ✅ Data persistence verified in database
- ✅ Data appears correctly in GET /api/workouts
- ✅ Round-trip import → export verified working
- ✅ Graceful handling of invalid rows (1 row with missing component type/name)

### Bug Fixes
- ✅ Investigated and resolved reported User Workouts Import bug
  - Testing confirmed feature working correctly
  - Data persists to database, appears in API, exports correctly
  - Bug report was false or already fixed in previous session

---

## [0.6.0-beta] - 2025-11-21

### Added
- **Database Backup/Restore System**
  - **Backend Service** (`internal/service/backup_service.go`)
    - `CreateBackup()` - Exports all tables to JSON + uploaded files to ZIP
    - `ListBackups()` - Returns metadata for all available backups
    - `GetBackupMetadata()` - Reads metadata from backup file
    - `DeleteBackup()` - Removes backup file with audit logging
    - `RestoreBackup()` - Full database restore from backup
    - Automatic table detection (skips tables that don't exist)
    - Multi-database support (SQLite, PostgreSQL, MySQL)
  - **API Endpoints** (`internal/handler/backup_handler.go`)
    - `POST /api/admin/backups` - Create new backup
    - `GET /api/admin/backups` - List all backups
    - `GET /api/admin/backups/{filename}` - Download backup file
    - `GET /api/admin/backups/{filename}/metadata` - Get backup metadata
    - `DELETE /api/admin/backups/{filename}` - Delete backup
    - `POST /api/admin/backups/{filename}/restore` - Restore from backup
    - All endpoints admin-only with authorization checks
    - Filename validation to prevent directory traversal attacks
  - **Frontend View** (`web/src/views/AdminBackupsView.vue`)
    - Backup list table with metadata (users, workouts, movements, WODs)
    - Create backup button with progress indicator
    - Download/delete/restore actions for each backup
    - Strong confirmation dialog for restore (warns about data loss)
    - Empty state for no backups
    - File size formatting and date/time display
  - **Backup Structure**
    - ZIP file containing `backup_data.json` with all table data
    - Includes uploaded files (profile pictures, etc.) in `uploads/` folder
    - Metadata: version, database driver, user counts, file size, created by
    - Stored in `/backups/` directory with .gitignore

- **Documentation**
  - Updated ProfileView with "Database Backups" admin navigation link
  - Added `/admin/backups` route to router
  - Created `backups/.gitignore` to prevent backup files from being committed

### Technical
- Clean Architecture maintained: domain → service → handler pattern
- Audit logging for backup creation, deletion, and restore operations
- Transaction-based restore to ensure data consistency
- Automatic cleanup of existing data before restore
- Security: Admin-only access, filename validation, token-based auth
- Build number incremented: 24 → 25

### Testing
- ✅ Create backup: Successfully generates 1.7MB ZIP file
- ✅ List backups: Returns metadata with correct statistics
- ✅ Download backup: Serves ZIP file for download
- ✅ Delete backup: Removes file and logs action
- ⚠️ Restore backup: Tested manually (destructive operation)

---

## [0.5.1-beta] - 2025-11-21

### Added
- **Import/Export System (Phase 1 & 2 - COMPLETE)**
  - **WOD Export** (`GET /api/export/wods`)
    - CSV format with all WOD fields
    - Query parameters for filtering: `include_standard`, `include_custom`
    - Successfully tested with all standard WODs
  - **Movement Export** (`GET /api/export/movements`)
    - CSV format with all movement fields
    - Query parameters for filtering: `include_standard`, `include_custom`
    - Successfully tested with all standard movements
  - **User Workouts Export** (`GET /api/export/user-workouts`)
    - JSON format with full nested data structure
    - Optional date range filtering: `start_date`, `end_date`
    - Includes metadata: user_email, export_date, version, total_count
    - Nested workout data: movements, WODs, performance metrics
  - **WOD Import** (`POST /api/import/wods/preview`, `POST /api/import/wods/confirm`)
    - Preview endpoint with validation before committing
    - CSV validation: source, type, regime, score_type enums
    - Duplicate detection and handling
    - Options: skip_duplicates, update_duplicates
    - Successfully tested: created custom WOD
  - **Movement Import** (`POST /api/import/movements/preview`, `POST /api/import/movements/confirm`)
    - Preview endpoint with validation
    - CSV validation: type enum (weightlifting, cardio, gymnastics, bodyweight)
    - Duplicate detection
    - Successfully tested: created 2 custom movements
  - **User Workouts Import** (`POST /api/import/user-workouts/preview`, `POST /api/import/user-workouts/confirm`)
    - Preview endpoint working correctly ✅
    - Confirm endpoint working correctly ✅
    - JSON parsing and validation
    - Nested data handling (movements, WODs)
    - Auto-creation of missing movements and WODs
    - Default workout_name generation for ad-hoc workouts

- **Frontend Views**
  - `ExportView.vue` at `/settings/export`
    - Data type selector (WODs, Movements, User Workouts)
    - Format handling (CSV for WODs/Movements, JSON for User Workouts)
    - Options: Include standard items, include custom items
    - Date range picker for User Workouts
    - Export button triggers file download
  - `ImportView.vue` at `/settings/import`
    - File upload with drag-and-drop support
    - Supported formats info (CSV, JSON)
    - Preview table showing parsed data with validation status
    - Validation errors display (red highlights for invalid rows)
    - Import statistics (total, valid, invalid, duplicates)
    - Import options: Skip duplicates, Update duplicates
    - Confirm and Cancel buttons
  - Fixed axios import to use authenticated instance (was causing 401 errors)

- **Backend Services (1,691 lines total)**
  - `internal/service/export_service.go` (385 lines)
  - `internal/service/import_service.go` (829 lines)
  - `internal/handler/export_handler.go` (178 lines)
  - `internal/handler/import_handler.go` (299 lines)
  - All routes wired up in `cmd/actalog/main.go`

- **Documentation**
  - Created `docs/ROADMAP.md` with detailed development plan
  - Updated `docs/TODO.md` with completion status
  - Testing results: 6/6 features working (100%)

### Fixed
- **User Workouts Import Persistence Bug** (CRITICAL - Build 22)
  - **Location:** `internal/service/import_service.go:760-776`
  - **Issue:** Import reported success but workouts didn't appear in API responses
  - **Root Cause 1:** Missing `WorkoutType` field when creating UserWorkout struct
    - Field was present in JSON import data but not being set on the domain object
    - Caused workout_type column to be NULL in database
  - **Root Cause 2:** Missing `workout_name` default value for ad-hoc workouts
    - Ad-hoc workouts (without workout_id) require workout_name to be queryable
    - `GetByIDWithDetails()` throws error when both workout_id and workout_name are NULL
    - This caused API to return empty array even though workouts existed in database
  - **Fix Applied:**
    - Added `WorkoutType: workoutData.WorkoutType` to UserWorkout struct creation
    - Added default workout_name generation: `fmt.Sprintf("Workout %s", workoutDate.Format("2006-01-02"))`
    - Ensures all ad-hoc workouts have a valid workout_name for retrieval
  - **Testing Results:**
    - Before: Database had workouts but API returned 0 ❌
    - After: Database has workouts AND API returns all workouts ✅
    - Verified via database query and `/api/workouts` endpoint

### Changed
- Version remains 0.5.0-beta (will bump to 0.5.1-beta on release)
- Build number incremented: 20 → 22
- Import/Export system is now 100% functional (6/6 features working)

### Testing
- ✅ WOD export/import round-trip tested successfully
- ✅ Movement export/import round-trip tested successfully
- ✅ User Workouts export tested successfully
- ✅ User Workouts import preview tested successfully
- ✅ User Workouts import confirm tested successfully (bug fixed)

### Technical
- Clean Architecture maintained throughout implementation
- Multi-database support (SQLite, PostgreSQL, MySQL)
- CSV parsing with validation
- JSON nested data handling
- Duplicate detection algorithms
- Authorization checks (users can only import/export their own data)
- Rate limiting on import endpoints
- File size limits (max 10MB)

---

## [0.4.6-beta] - 2025-11-15

### Added
- **Admin User Management Enhancements**
  - Delete user functionality with confirmation dialog (`DELETE /api/admin/users/{id}`)
  - Prevents admin from deleting their own account
  - Displays what will be deleted: profile, workouts, PRs, performance history
  - Audit logging for all user deletion operations
  - Service layer validation with authorization checks

- **Session Management System**
  - List active sessions endpoint (`GET /api/sessions`)
  - Revoke specific session endpoint (`DELETE /api/sessions/{id}`)
  - Revoke all sessions endpoint (`POST /api/sessions/revoke-all`)
  - Session ownership validation (users can only manage their own sessions)
  - Audit logging for session revocation events
  - Service layer: `GetActiveSessions()`, `RevokeSession()`, `RevokeAllSessions()`
  - Handler layer: `SessionHandler` with list, revoke, and revoke-all operations
  - All endpoints require authentication

- **User Repository Enhancements**
  - Fixed `List()` method to include all admin-relevant fields
  - Now properly selects: email_verified, account_disabled, locked_at, locked_until, disable_reason
  - Proper NULL handling for all nullable timestamp and string fields

### Fixed
- **Admin Users View**
  - Icons now correctly display current user state (verified, locked, enabled, role)
  - Dynamic icon shapes and colors based on state
  - Enhanced tooltips showing current state explicitly
  - All toggles (verify email, lock, enable, role, delete) now work correctly

### Changed
- Version bumped to 0.4.6-beta
- Build number reset to 1 for new minor version
- Admin panel now has full CRUD capabilities for user management

### Technical
- Clean Architecture maintained: service → repository pattern
- Audit trail for all administrative actions
- Service layer performs authorization checks before operations
- CASCADE delete configured for related user data
- Security: token ownership validation prevents unauthorized access

## [0.4.5-beta] - 2025-11-14

### Added
- **Admin Data Cleanup: Edit WOD Records**
  - New API endpoint: `PUT /api/admin/data-cleanup/wod-record/{id}` for updating individual WOD records
  - Backend validation ensures updates match WOD score_type requirements
  - Clickable mismatch cards in admin cleanup view open edit dialog
  - Edit dialog with score_type-specific form fields (only shows relevant fields)
  - Hours, Minutes, Seconds input for Time-based WODs
  - Rounds and Reps input for Rounds+Reps WODs
  - Weight input for Max Weight WODs

### Fixed
- **Quick Log Form (Dashboard)**
  - Fixed Quick Log to respect WOD score_type constraints
  - Score type now auto-populates from selected WOD (read-only)
  - Only shows fields relevant to the WOD's score_type
  - Time-based WODs now support HH:MM:SS format (was only showing seconds)
  - Added reactive watchers to auto-calculate total seconds from HH:MM:SS inputs
- **Log Workout Form**
  - Fixed score_type check from `'Time'` to `'Time (HH:MM:SS)'`
  - Added hours field to time inputs (now fully supports HH:MM:SS)
  - Updated time calculation logic to include hours in total seconds and score_value formatting
- **Admin Cleanup View**
  - Removed duplicate bottom navigation bar from admin cleanup view

### Changed
- All workout entry and edit forms now enforce WOD score_type constraints
- Time-based WODs consistently use HH:MM:SS format across all forms
- Version bumped to 0.4.5-beta build 1

### Technical
- Frontend conditional rendering prevents invalid field combinations based on score_type
- Backend validation in `UpdateWODRecord` ensures data integrity
- Multi-layer constraint enforcement: frontend UX + backend validation
- Clean Architecture maintained: handler → service → repository layers

## [0.4.4-beta] - 2025-11-14

### Added
- **Retroactive PR Detection System**
  - Service method `RetroactivelyFlagPRs()` to analyze all historical workouts chronologically
  - Automatically flags PRs based on historical max values for movements and WODs
  - Processes workouts in chronological order, tracking max weights, best times, and best rounds+reps
  - Repository methods: `UpdatePRFlag()` for both movements and WODs
  - API endpoint: `POST /api/workouts/retroactive-flag-prs` (authenticated)
  - Command-line script `scripts/retroactive_prs.go` for direct database PR flagging
  - Returns count of movement PRs and WOD PRs flagged

### Fixed
- PR detection now works for historical workouts logged before PR system was implemented
- Personal Records view now displays PRs from all workouts, not just newly logged ones
- Resolved issue where existing workouts had `is_pr = 0` even when they contained record performances

### Technical
- Chronological processing ensures PRs are correctly identified based on order of performance
- In-memory tracking of max values during processing to avoid multiple database queries
- Multi-database support (SQLite, PostgreSQL, MySQL) for PR flag updates
- Clean Architecture maintained: domain interfaces → repository implementation → service logic → handler/script

### Changed
- Version bumped to 0.4.4-beta across all version files (pkg/version/version.go, web/package.json)

## [0.4.0-beta] - 2025-11-12

### Added
- **WOD (Workout of the Day) Management System**
  - Database migration v0.4.0 adding `wods` table with complete schema
  - WOD entity with fields: name, source, type, regime, score_type, description, standards, url, time_cap
  - Seeded 10 standard WODs: 8 Girl WODs (Fran, Helen, Cindy, Grace, Annie, Karen, Diane, Elizabeth) + 2 Hero WODs (Murph, DT)
  - Repository layer: `WODRepository` with CRUD operations, search, and filtering
  - Service layer: `WODService` with validation, authorization, and business logic
  - Handler layer: `WODHandler` with RESTful API endpoints
  - API endpoints: `GET /api/wods`, `GET /api/wods/{id}`, `GET /api/wods/search`, `POST /api/wods`, `PUT /api/wods/{id}`, `DELETE /api/wods/{id}`
  - Support for both standard (pre-seeded) and custom (user-created) WODs
  - WOD types: Benchmark, Hero, Girl, Notables, Games, Endurance, Self-created
  - WOD sources: CrossFit, Other Coach, Self-recorded
  - WOD regimes: EMOM, AMRAP, Fastest Time, Slowest Round, Get Stronger, Skills
  - Score types: Time (HH:MM:SS), Rounds+Reps, Max Weight

- **Workout Template System**
  - Database migration v0.4.0 adding `workout_wods` linking table
  - Workout templates can now include WODs (many-to-many relationship)
  - Repository layer: `WorkoutWODRepository` for managing workout-WOD associations
  - Service layer: `WorkoutWODService` with business logic for linking WODs to templates
  - API endpoints: `POST /api/templates/{id}/wods`, `GET /api/templates/{id}/wods`, `PUT /api/templates/wods/{id}`, `DELETE /api/templates/wods/{id}`, `POST /api/templates/wods/{id}/toggle-pr`
  - Seeded 3 workout templates with movements: Back Squat Focus, Olympic Lifting, Gymnastics Strength
  - Templates can combine movements and WODs in single workout plan

- **Frontend State Management (Pinia Stores)**
  - `useWodsStore` - Complete WOD state management with CRUD operations
    - Actions: fetchWods(), fetchWodById(), searchWods(), createWod(), updateWod(), deleteWod()
    - Filters: filterByType(), filterBySource(), getStandardWods(), getCustomWods()
  - `useTemplatesStore` - Complete template state management
    - Actions: fetchTemplates(), fetchTemplateById(), fetchMyTemplates(), createTemplate(), updateTemplate(), deleteTemplate()
    - WOD linking: fetchTemplateWods(), addWodToTemplate(), removeWodFromTemplate(), toggleWodPR()
    - Filters: getStandardTemplates(), getCustomTemplates(), getTemplatesWithMovementCount()

- **WOD Library View**
  - Updated `/wods` route to use new Pinia store (useWodsStore)
  - Browse all standard WODs with filtering by type (Benchmark, Girl, Hero, Games)
  - Search WODs by name/description
  - View WOD details with regime, score type, time cap, standards
  - Create/edit custom WODs (authenticated users only)
  - Selection mode for linking WODs to workout templates

### Changed
- Workout templates now support WODs in addition to movements
- Updated WOD Library view to use Pinia state management instead of direct axios calls
- Database schema extended to support workout-WOD relationships
- Version bumped to 0.4.0-beta across all version files

### Technical
- Multi-table seeding: movements, WODs, workout templates, workout_movements, workout_wods
- Clean Architecture maintained: domain → repository → service → handler → store → view
- Idempotent seeding with sentinel checks to prevent duplicate data
- WOD validation includes enum validation for source, type, regime, score_type
- Authorization checks: only WOD creators can modify/delete custom WODs
- Frontend stores follow Pinia Composition API pattern with proper error handling
- Multi-database support (SQLite, PostgreSQL, MySQL) for all new tables

### In Progress
- Dashboard view integration with templates and WODs
- Template Library browsing view
- Template-based workout logging in LogWorkoutView

## [0.3.1-beta] - 2025-11-10

### Added
- **Email Verification System (Complete)**
  - Database migration v0.3.1 adding `email_verified` and `email_verified_at` columns to users table
  - Backend API endpoints: `GET /api/auth/verify-email`, `POST /api/auth/resend-verification`
  - Email service with SMTP integration for sending verification emails
  - Styled HTML email templates with verification links
  - 24-hour token expiration with secure token generation (crypto/rand)
  - Single-use verification tokens (marked as used after verification)
  - Repository layer: `CreateVerificationToken()`, `GetVerificationToken()`, `MarkTokenAsUsed()`
  - Service layer: `SendVerificationEmail()`, `VerifyEmailWithToken()`, `ResendVerificationEmail()`
  - Handler layer: `VerifyEmail()`, `ResendVerification()` with proper error handling

- **Email Verification Frontend**
  - VerifyEmailView component at `/verify-email?token=...` route
    - Automatic email verification on page load
    - Loading, success, and error states with appropriate messaging
    - Handles expired, invalid, and already-used tokens
    - Updates auth store user object on successful verification
  - ResendVerificationView component at `/resend-verification` route
    - Email input form to request new verification email
    - Success confirmation displaying the email address
    - Comprehensive error handling (404, 400, network errors)
  - Updated RegisterView to show verification success message
    - No longer auto-redirects to dashboard after registration
    - Displays sent email address and 24-hour expiration notice
    - Link to resend verification if email not received
  - Dashboard verification status banner
    - Warning alert for users with unverified emails
    - Prominent "Resend Email" button
    - Closable alert for better UX

### Changed
- User registration flow now includes email verification step
- Users receive verification email immediately after registration
- Dashboard shows verification reminder until email is verified
- Router updated with `/verify-email` and `/resend-verification` routes
- Navigation guards allow verify-email access for both authenticated and unauthenticated users
- Version bumped to 0.3.1-beta across all version files

### Technical
- Email verification tokens stored in `email_verification_tokens` table
- Tokens generated using crypto/rand (32 bytes hex-encoded) for security
- SMTP configuration via environment variables (EMAIL_FROM, SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS)
- HTML email template with inline styles for cross-client compatibility
- Authorization checks ensure users can only resend verification for their own email
- Frontend build: 618 modules, 47 PWA cache entries
- Multi-database support (SQLite, PostgreSQL, MySQL) for email_verified field

## [0.3.0-beta] - 2025-11-10

### Added
- **Personal Records (PR) Tracking System**
  - Automatic PR detection when logging workouts (weight-based comparison)
  - Manual PR flag toggle via API endpoint
  - Database migration v0.3.0 adding `is_pr` column to workout_movements
  - Multi-database support (SQLite, PostgreSQL, MySQL) for PR field
  - New domain models: `PersonalRecord` struct and `IsPR` field in `WorkoutMovement`
  - Repository methods: `GetPersonalRecords()`, `GetMaxWeightForMovement()`, `GetPRMovements()`
  - Service layer methods: `DetectAndFlagPRs()`, `GetPersonalRecords()`, `TogglePRFlag()`
  - API endpoints: `GET /api/workouts/prs`, `GET /api/workouts/pr-movements`, `POST /api/workouts/movements/:id/toggle-pr`
  - Gold trophy badges (mdi-trophy) on workout cards containing PRs
  - Individual PR indicators next to movements in workout lists
  - Dedicated PR History page at `/prs` route showing recent PRs and all-time records
  - Visual distinction with gold/amber color scheme (#ffc107) for PR indicators

- **Password Reset Frontend (Part 3/3)**
  - Forgot Password view with email submission form
  - Reset Password view with token validation and new password form
  - Router configuration for `/forgot-password` and `/reset-password/:token` routes
  - "Forgot password?" link added to Login view
  - Integration with backend password reset API endpoints
  - Success/error messaging for user feedback

### Changed
- Integrated PR detection into workout creation workflow
- Updated RecentWorkoutsCards component to display PR badges
- Updated WorkoutsView to show PR indicators on individual movements
- Enhanced router with authentication guards for password reset routes
- Version bumped to 0.3.0-beta across all version files

### Technical
- PR auto-detection algorithm: compares current weight against previous max for each movement
- Authorization checks on PR flag toggle to ensure workout ownership
- Backward-compatible database migration with DEFAULT values
- Clean Architecture maintained: domain → repository → service → handler layers
- All PR queries include proper user scoping for security

## [0.2.0-beta] - 2025-11-06

### Added
- Complete workout CRUD functionality with RESTful API endpoints
- Workout repository layer for database operations
- Movement repository with 31 seeded standard CrossFit movements
- Workout movement repository for linking movements to workouts
- Workout service layer with business logic and authorization
- JWT authentication middleware for protected routes
- Dashboard with real-time workout statistics (total workouts, monthly count)
- Recent workouts display on dashboard (last 5 workouts)
- Workout saving functionality from Log Workout screen
- Workouts list view with movement details
- Autocomplete/search functionality for movement selection
- Custom movement item templates showing type and icons
- Modern UI design with cyan accent color (#00bcd4)
- Dark navy header (#2c3e50) across all views
- Responsive scrolling with fixed header and footer navigation

### Changed
- Updated LogWorkoutView with functional save button and API integration
- Updated WorkoutsView to fetch and display real workout data
- Updated DashboardView to show live statistics from API
- Updated PerformanceView with searchable movement dropdown
- Improved font readability with darker colors (#1a1a1a)
- Reduced vertical spacing for better mobile fit
- Changed v-select components to v-autocomplete for better UX
- Enhanced workout responses to include full movement details

### Fixed
- Cache directory creation issue in Makefile (mkdir -p added to run/dev targets)
- SQLite driver name changed from "sqlite" to "sqlite3" in config
- Workout save button now properly calls API endpoint
- Vertical scrolling enabled on all views
- Content no longer runs off bottom of screen
- Movement names now display correctly in workout lists

### Technical
- Implemented Clean Architecture pattern (domain → repository → service → handler)
- Added dependency injection for repositories and services
- Integrated JWT token validation in middleware
- Database seeding for standard movements on first run
- Proper error handling and validation in API endpoints
- User authorization checks in workout service layer

## [0.1.0-alpha] - 2025-11-05

### Added
- Initial project setup with Go backend and Vue.js frontend
- User authentication with JWT tokens
- Basic user registration and login endpoints
- Database schema for users, workouts, movements, and workout_movements
- SQLite and PostgreSQL database support
- Vue.js frontend with Vuetify 3 UI framework
- Vue Router setup with authentication guards
- Pinia store for state management
- Basic view scaffolding (Dashboard, Performance, Workouts, Profile, Login, Register)
- Bottom navigation with mobile-first design
- Clean Architecture folder structure
- Configuration management with environment variables
- Makefile for common development tasks
- Documentation (README.md, ARCHITECTURE.md, AI_INSTRUCTIONS.md, DATABASE_SCHEMA.md)

### Technical
- Go 1.24+ with Chi router
- Vue 3 with Composition API
- Vuetify 3 for UI components
- Axios for HTTP requests
- bcrypt for password hashing
- JWT for authentication
- SQLite3 driver integration
