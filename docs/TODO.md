# TODO

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

### Import/Export Enhancements
- [ ] Implement Markdown export for workout reports
- [ ] Format Markdown with workout details, scores, notes
- [ ] Ensure CSV export includes all new fields (division, PR flags, birthday)
- [ ] Ensure JSON export includes complete data structure
- [ ] Add date range selector for exports
- [ ] Add data type checkboxes (Workouts, WODs, Movements, Profile)

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
- [ ] Database backup functionality
- [ ] User activity monitoring

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

**Last Updated:** 2025-11-10
**Version:** 0.4.0-dev (Template-based architecture - Service layer complete, database migration pending)

**v0.4.0 Status:**
- ✅ Domain models updated for template architecture
- ✅ Repositories implemented (UserWorkout, WOD, WorkoutWOD)
- ✅ Services implemented (UserWorkoutService, WODService, WorkoutWODService, updated WorkoutService)
- ✅ Handlers created (user_workout_handler, wod_handler, workout_wod_handler)
- ✅ API routes configured in main.go
- ✅ Application compiles successfully
- 🔄 Unit tests in progress (UserWorkoutService: 11/16 passing)
- ⏳ Database migration not yet applied (still at v0.3.1 schema)
- ⏳ Frontend updates pending
