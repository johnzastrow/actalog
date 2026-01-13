# ActaLog TODO

> **Last Updated:** 2026-01-13
> **Current Version:** 0.22.0-beta (Build 8)

---

## Claude Instructions

**This is the canonical TODO file for ActaLog.** All task tracking should be done here.

### Guidelines for Claude Code:

1. **Only track TODOs here** - Do not create TODO sections in other documentation files
2. **Active Tasks** - Items currently being worked on (move here when starting)
3. **Backlog** - Planned features, known bugs, and improvements not yet started
4. **Completed Releases** - Keep only the last 5 releases here; older releases are in CHANGELOG.md
5. **Periodic Cleanup** - When starting a new session, clean up completed items and archive old releases
6. **Technical Details** - Include file paths and implementation notes for completed work
7. **Priority Markers** - Use `[HIGH]`, `[MEDIUM]`, `[LOW]` for backlog items
8. **Bug Format** - `[BUG]` prefix for bug reports
9. **Code TODOs** - Periodically scan codebase for TODO/FIXME comments:
   ```bash
   grep -rn "TODO\|FIXME" --include="*.go" --include="*.vue" --include="*.js" .
   ```

### File Relationships:
- `TODO.md` - Active task tracking (this file)
- `CHANGELOG.md` - Release history and change documentation
- `ROADMAP.md` - High-level version planning and feature roadmap
- Do NOT duplicate TODO content in ROADMAP.md or CHANGELOG.md

---

## Active Tasks

*Items currently being worked on. Move items here from Backlog when starting.*

### Test Suite Cleanup *(Completed)*

**Status:** Phase 1 Complete, Phase 2 Complete
**Tracking:** See `docs/TEST_CLEANUP.md` for detailed summary

**Phase 1 - Removed ~31 low-value tests:**
- Struct field assignment tests (19)
- Language feature tests (10)
- Trivial helper tests (2)

**Phase 2 - Removed ~267 panic-expectation tests:**
- All handler test files cleaned (24 files total)
- Panic tests that created nil dependencies removed
- Validation tests and mock-based tests preserved
- All tests pass with proper coverage

### CI/Lint Fixes (Deferred)

**Status:** golangci-lint set to `continue-on-error` in `.github/workflows/ci.yml:37`

The following lint issues need to be resolved to re-enable strict linting:

1. **goconst warnings** - Repeated strings in SQL queries (LIMIT, OFFSET, ORDER BY)
   - Files: `internal/repository/movement_repository.go`, others
   - Decision: Either make constants or add exclusion

2. **Remaining gofmt issues** - Some files may still have formatting issues
   - Run: `gofmt -w ./...` and review changes

3. **Other minor issues** - Review full golangci-lint output for any remaining items

**To re-enable strict linting:**
1. Fix or exclude remaining issues
2. Remove `continue-on-error: true` from `.github/workflows/ci.yml:37`

---

## Backlog

### High Priority

#### Comprehensive Benchmark Endpoint *(Completed v0.22.0)*
- [x] `[HIGH]` **Create `/api/benchmark` endpoint for system-wide performance testing**
  - Exercises database, serialization, business logic, and concurrency through synthetic operations
  - Uses isolated `benchmark_data` table (never touches real user data)
  - Auto-cleanup after runs, user-scoped data isolation
  - JWT authentication required
  - **Backend implementation:**
    - [x] Domain model (`internal/domain/benchmark.go`) - BenchmarkData entity, result structs, repository interface
    - [x] Migration 0.22.0 (`internal/repository/migrations.go`) - benchmark_data table with indexes
    - [x] Migration 0.23.0 - Extended fields for stress testing (large_text, json_blob, uuid)
    - [x] Repository (`internal/repository/benchmark_repository.go`) - CRUD + batch operations
    - [x] Service (`internal/service/benchmark_service.go`) - 18+ benchmark operations
    - [x] Handler (`internal/handler/benchmark_handler.go`) - POST /benchmark, GET /benchmark/status
    - [x] Route registration (`cmd/actalog/main.go`) - wire up handler
  - **Benchmark operations:**
    - Database: Insert, BulkInsert(100), SelectByID, SelectByKey, SelectList, SelectFiltered, Update, Delete
    - Serialization: Marshal/Unmarshal small & large JSON
    - Business Logic: 1RM calculations (1000x), validation, string/date operations
    - Concurrent (optional): 10 parallel reads, 5 writes, mixed operations
  - **Configurable record count:**
    - [x] `?records=N` query parameter (default: 1000, max: 500000)
    - [x] Complex benchmark data (5-10KB text, nested JSON blobs)
    - [x] Server timeout configuration for large record counts
  - **Benchmark tool integration (`actalog-benchmark` v0.7.0):**
    - [x] Add `/api/benchmark` caller (`internal/metrics/benchmark_api.go`)
    - [x] Add BenchmarkAPIResult to types.go
    - [x] Update reporters to include benchmark API results
    - [x] Add `--benchmark-records` flag for configurable record count
    - [x] Comparison reports include server-side benchmark results

#### Subscription System (Backend Complete - v0.14.0, Frontend Complete - v0.17.0)
- [x] `[HIGH]` **Frontend Subscription Status Display**
  - [x] Add subscription status badge to user profile/settings (SubscriptionStatusBadge.vue in SettingsView)
  - [x] Show expiration date for paid subscriptions
  - [x] Display "Permanent Free" badge for permanent users
  - [x] Show organization subscriptions the user benefits from

- [x] `[HIGH]` **Admin Subscription Management UI**
  - [x] Admin panel for viewing all user subscriptions (AdminSubscriptionsView.vue)
  - [x] Admin panel for viewing all organization subscriptions
  - [x] Create subscription form (CreateSubscriptionDialog.vue)
  - [x] Mark subscription as paid button/action (MarkAsPaidDialog.vue)
  - [x] Cancel subscription with reason input (CancelSubscriptionDialog.vue)
  - [x] View subscription history (SubscriptionDetailDialog.vue)
  - [x] List expiring subscriptions API (backend complete v0.19.0)
  - [x] List overdue/expired subscriptions API (backend complete v0.19.0)
  - [x] Frontend UI for expiring/expired subscription lists (AdminSubscriptionsView.vue tabs)

- [x] `[HIGH]` **Read-Only Mode UI Feedback**
  - [x] Graceful handling of HTTP 402 Payment Required responses (subscription.js store)
  - [x] Show "Subscription Expired" banner when in read-only mode (SubscriptionExpiredBanner.vue in App.vue)
  - [x] Disable/hide create/edit buttons when subscription expired
  - [x] "Renew Subscription" call-to-action in banner
  - [ ] Toast notifications explaining why operations are blocked

#### Front end Improvements
- [x] `[HIGH]` **PWA**
  - [x] Icon saved to home screen is fuzzy - regenerated all icons from SVG with full RGBA color
- [x] `[HIGH]` **Overall**
  - [x] Find more attractive icons for the navigation bar - replaced with Font Awesome icons (fa-house-chimney, fa-arrow-trend-up, fa-person-running, fa-circle-user)
  - [x] Allow notifications to be marked as read/unread (implemented in NotificationsView.vue)
  - [x] Develop a theme called "Sunrise" based on colors from sunset image (added to vuetify.js and theme.js)
  - [x] Remove the View all link on the Dashboard (removed from DashboardView.vue)
  - [x] Add more helper text to form controls throughout the app - Added markdown hints to textarea fields in WODCreateView, WODEditView, MovementCreateView, MovementEditView, WorkoutTemplateEditView, AdminAnnouncementsView (v0.22.0)

- [x] `[HIGH]` **User-Customizable Fonts** *(Completed v0.19.0)*
  - Self-hosted web fonts with 10 options (including accessibility fonts)
  - Backend sync via user_settings table + localStorage cache
  - All fonts use SIL OFL 1.1 or Apache 2.0 (free for commercial use, self-hosting allowed)
  - **Font options:**
    - System Default, Inter, Roboto, Lato, Fira Sans, Lexend
    - OpenDyslexic (A11y), Atkinson Hyperlegible (A11y)
    - Source Serif Pro, JetBrains Mono
  - **Backend changes:**
    - [x] Add `font_family` field to `internal/domain/user_settings.go`
    - [x] Add migration 0.21.0 in `internal/repository/migrations.go`
    - [x] Update SQL queries in `internal/repository/user_settings_repository.go`
    - [x] Add default + audit logging in `internal/service/user_settings_service.go`
  - **Frontend changes:**
    - [x] Download fonts to `web/public/fonts/` (woff2 format)
    - [x] Create `web/src/assets/fonts.css` with @font-face declarations
    - [x] Create `web/src/stores/font.js` (font store)
    - [x] Update `web/src/stores/settings.js` (add fontFamily sync)
    - [x] Update `web/src/App.vue` (use CSS variable `--app-font-family`)
    - [x] Add font selector UI in `web/src/views/SettingsView.vue`
  - **Performance:** `font-display: swap`, service worker caching, only selected font loads 


#### Backend Improvements
- [x] `[HIGH]` **Backup and Restore** make sure the backup functions keep up with the database schema changes (tested backup and restore round trips on SQLite, PostgreSQL, and MariaDB).
  - [x] Added schema metadata to backup format for type-aware restoration
  - [x] Implemented three restore modes: `replace`, `merge`, `skip`
  - [x] Natural key matching (users by email, movements by name, etc.)
  - [x] ID remapping for foreign key references during merge/skip
- [x] `[HIGH]` **Duplicate Record Protection**
  - [x] All imports now protect against duplicate records:
    - [x] WOD Import: `skip_duplicates` and `update_duplicates` options
    - [x] Movement Import: `skip_duplicates` and `update_duplicates` options
    - [x] User Workout Import: Added `update_duplicates` option (v0.17.0)
    - [x] Wodify Import: Added `skip_duplicates` and `update_duplicates` options (v0.17.0)
  - [x] System restore supports merge/skip modes with natural key matching
  - [x] API returns detailed results (records created, updated, skipped)
- [x] `[HIGH]` **Check for missing indexes** - 83 indexes defined in migrations (foreign keys and commonly queried fields covered)
- [x] `[HIGH]` **Audit Log Enhancements** - Added 17+ helper methods for comprehensive audit coverage (logout, token refresh, profile updates, user deletion, organization CRUD, subscription lifecycle)
- [x] `[HIGH]` **Database Query Optimization** - Added audit_logs indexes (migration 0.17.0), optimized GetUsageStats (3→1 query), GetActiveUsersThisMonth (3→2 queries), ListByUserWithDetails N+1 fix (6N+1→5 queries)
- [x] `[HIGH]` **Error Handling Consistency** - Centralized error handling in internal/handler/errors.go with HTTPError type and 30+ service error mappings
- [x] `[HIGH]` **Structured Logging** - Added JSON format support to pkg/logger with InfoWithFields, ErrorWithFields, and FieldLogger chaining
- [x] `[HIGH]` **Comprehensive API Documentation** - OpenAPI/Swagger documentation at `/docs/swagger.json` (v0.22.0)
- [x] `[HIGH]` **Implement Rate Limiting** - pkg/middleware/rate_limit.go with sliding window algorithm

## Future Enhancements (Post-MVP)

These features can be added after the core frontend is complete:

1. **Stripe Integration** (v0.15.0+)
   - Replace manual admin control with automated billing webhooks
   - Credit card payment processing
   - Automatic subscription renewal

2. **Email Notifications** (v0.15.0+)
   - Notify users 7/3/1 days before expiration
   - Notify users when subscription expires
   - Notify admins of failed payments

3. **Self-Service Portal** (v0.16.0+)
   - Users can upgrade/downgrade themselves
   - View payment history
   - Update payment method

4. **Usage Limits** (v0.16.0+)
   - Track API usage for free tier
   - Enforce limits on free subscriptions
   - Display usage metrics

5. **Bulk Operations** (v0.17.0+)
   - Admin can bulk-update subscriptions
   - Bulk extend expiration dates
   - Bulk cancel subscriptions

6. **Grace Period Configuration** (v0.17.0+)
   - Make grace period configurable per organization
   - Different grace periods for different subscription tiers


#### Testing Coverage (81.6% Service Layer)
- [x] `[HIGH]` **Subscription Service Tests** - Comprehensive test suite (internal/service/subscription_service_test.go)
- [x] `[HIGH]` **Backup Service Tests** - 10 test functions with full restore mode coverage (internal/service/backup_service_test.go)
- [x] `[HIGH]` **User Workout Service Tests** - Complete coverage with PR detection
- [x] `[HIGH]` **Wodify Import Service Tests** - Duplicate handling and error recovery
- [x] `[HIGH]` **Most Service Tests Complete** - 13 services at 100% coverage
- [x] `[HIGH]` **Add handler unit tests** - Handler coverage at 72.6%
  - [x] benchmark_handler_test.go (17 tests, 53-100% coverage per method)
  - [x] auth_handler_test.go (63-100% coverage, comprehensive with mock service)
  - [x] user_workout_handler_test.go (32-100% coverage)
  - [x] movement_handler_test.go (59-100% coverage)
  - [x] wod_handler_test.go (84-100% coverage)
  - [x] subscription_handler_test.go (28-69% coverage)
  - [x] All other handlers have tests: settings, admin, performance, data_change_log, organization, etc.
- [ ] `[MEDIUM]` **Improve user_service coverage** - Currently 61.9%, target 80%+
- [ ] `[MEDIUM]` **Improve import_service coverage** - Currently 60.8%, target 80%+
- [ ] `[LOW]` **Add repository unit tests** - All repository implementations

#### Admin Features
- [x] `[HIGH]` **User Import/Export System** (Admin only) *(Completed v0.23.0)*
  - [x] Export users to CSV format (email, name)
  - [x] Import users from CSV (email, name, password) with preview/confirm workflow
  - [x] Preview workflow with validation
  - [x] Duplicate detection by email
  - [x] Batch password reset emails - select users from filterable list
  - **Backend files:** `internal/service/user_import_service.go`, `internal/handler/user_import_handler.go`
  - **Frontend:** `web/src/views/AdminUserImportExportView.vue` with 3 tabs (Import, Export, Password Reset)
  - **API endpoints:**
    - `POST /api/admin/users/import/preview` - Preview CSV import
    - `POST /api/admin/users/import/confirm` - Execute import
    - `GET /api/admin/users/export` - Download CSV
    - `GET /api/admin/users/filter` - List users with search/date filters
    - `POST /api/admin/users/batch-password-reset` - Send reset emails to selected users

- [x] `[HIGH]` **Improved Duplicate Record Detection During Imports** (v0.17.0)
  - [x] Trap duplicate movements during import (check by name)
  - [x] Trap duplicate WODs during import (check by name)
  - [x] Trap duplicate user workouts during import (check by date/user/name)
  - [x] Trap duplicate Wodify workouts during import (check by date)
  - [x] Import options for handling duplicates:
    - [x] Skip duplicates (keep existing) - `skip_duplicates=true`
    - [x] Update duplicates (overwrite existing) - `update_duplicates=true`
  - [x] Detailed import result with created/updated/skipped counts
  - [ ] Frontend UI to review and select action for each duplicate
  - [ ] Batch duplicate resolution UI (apply same action to all)

- [ ] `[HIGH]` **Database Duplicate Detection and Cleaning Procedure**
  - [ ] Create admin tool to scan for existing duplicates in database
  - [ ] Detect duplicate movements (same name, same user)
  - [ ] Detect duplicate WODs (same name, same source)
  - [ ] Detect duplicate user workouts (same user, same date, same template)
  - [ ] Generate duplicate report with:
    - Count of duplicates per table
    - List of duplicate records with IDs
    - References/dependencies (what's using each record)
  - [ ] Safe merge/cleanup procedure:
    - Preview which records will be kept vs deleted
    - Automatically update foreign key references
    - Preserve data integrity (don't break relationships)
  - [ ] Admin UI to review and approve cleanup
  - [ ] Dry-run mode (show what would happen without doing it)
  - [ ] Backup database before cleanup operation
  - [ ] Audit log of all cleanup operations

### High Priority

#### Documentation & Marketing
- [ ] `[High]` **Enhance App Documentation with Screenshots**
- [x] `[High]` Static site deployed to GitHub Pages (https://johnzastrow.github.io/actalog/)
  - [x] Home page with feature highlights and theme gallery
  - [x] For Admins page with subscription and announcement features
  - [x] For Developers page with tech stack, architecture, and deployment guide
  - [x] Performance & Benchmarks section added (v0.22.0) - minimum resources, load test results
  - [x] FAQ page
  - [x] Auto-deploy via GitHub Actions on push to main
  - [ ] Add more screenshots demonstrating all available themes
  - [ ] Show desktop vs mobile responsive views
  - [ ] Document key user flows with visual guides


#### Performance & Analytics
- [x] `[MEDIUM]` **Calendar View** - WorkoutCalendarView.vue with workout dots and month navigation
- [x] `[MEDIUM]` **Timeline View** - WorkoutTimelineView.vue with chronological history
- [x] `[MEDIUM]` **Progress Charts** - WeightProgressChart.vue and WorkoutFrequencyChart.vue components (not yet integrated into views)
- [x] `[MEDIUM]` **Admin Metrics Dashboard** - User stats, workout counts, system health (v0.22.0)
- [ ] `[MEDIUM]` **PR Leaderboards** - Opt-in community leaderboards

#### Testing
- [x] `[MEDIUM]` **Add backup_service tests** - backup_service_test.go with 10 test functions covering create/restore/export
- [x] `[MEDIUM]` **Add wodify_import_service tests** - wodify_import_service_test.go with duplicate handling
- [ ] `[MEDIUM]` **Add export/import_service tests** - Currently at 60-83% coverage
- [ ] `[MEDIUM]` **Add admin_handler tests**

#### PWA Enhancements
- [ ] `[MEDIUM]` Run Lighthouse PWA audit
- [ ] `[MEDIUM]` Optimize service worker cache size
- [ ] `[MEDIUM]` Test offline sync end-to-end on mobile devices

### Low Priority

#### Testing
- [x] `[LOW]` **Add audit_log_service tests** - 100% coverage achieved
- [x] `[LOW]` **Add wodify_import_service tests** - 100% coverage achieved

#### Features
- [ ] `[LOW]` **Push Notifications** - Workout reminders
- [ ] `[LOW]` **Data Visualization** - Charts for PR progression
- [ ] `[LOW]` **Social Features** - Share workouts (opt-in)

---

## Known Bugs

*Report bugs here with reproduction steps.*

*(none currently)*

---

## Code TODOs

*TODOs found in source code comments. These should be addressed or promoted to Backlog.*

### Backend (Go)

| File | Line | Description |
|------|------|-------------|
| `internal/service/workout_service.go` | 396 | Add proper authorization through workout template ownership |
| ~~`internal/service/import_service.go`~~ | ~~587, 701~~ | ~~Add duplicate detection using userWorkoutRepo.ListByUserAndDateRange~~ **DONE (v0.17.0)** |
| `internal/handler/movement_handler.go` | 141 | Get user ID from context when auth middleware is added |

### Frontend (Vue)

| File | Line | Description |
|------|------|-------------|
| `web/src/views/WorkoutsView.vue` | 372 | Navigate to template detail page |

*Last scanned: 2026-01-09*

**Note:** Calendar view (WorkoutCalendarView.vue) and PR notification system are implemented.

---

## Completed Releases

### v0.22.0-beta (2026-01-09)

**Status:** Comprehensive benchmark endpoint with configurable stress testing.

**Completed:**
- [x] **Benchmark API Endpoint**
  - [x] `POST /api/benchmark` - Full system benchmark with 18+ operations
  - [x] `GET /api/benchmark/status` - Quick status check
  - [x] `DELETE /api/admin/benchmark/data` - Admin cleanup
  - [x] Isolated `benchmark_data` table (never touches production data)
  - [x] Auto-cleanup after runs

- [x] **Configurable Record Count**
  - [x] `?records=N` query parameter (default: 1000, max: 500000)
  - [x] Complex benchmark data (5-10KB text, nested JSON, UUIDs)
  - [x] Migration 0.23.0 for extended benchmark fields

- [x] **Server Timeout Configuration**
  - [x] Fixed EOF errors on long-running benchmarks
  - [x] Updated default SERVER_WRITE_TIMEOUT to 60s
  - [x] Buffered JSON response with explicit Content-Length

- [x] **Benchmark Tool Integration (actalog-benchmark v0.7.0)**
  - [x] `--benchmark-records` flag for configurable record count
  - [x] Server-side benchmark comparison in reports
  - [x] Enhanced error display with word wrapping

- [x] **Static Site Updates**
  - [x] Performance & Benchmarks section on For Developers page
  - [x] Updated version to 0.22.0-beta
  - [x] Added Myst theme (7 total themes)

**Files Created:** `internal/domain/benchmark.go`, `internal/repository/benchmark_repository.go`, `internal/service/benchmark_service.go`, `internal/handler/benchmark_handler.go`

**Files Modified:** `internal/repository/migrations.go` (0.22.0, 0.23.0), `configs/config.go`, `.env.example`, `site/tech.html`, `site/index.html`

---

### v0.16.0-beta (2025-12-20)

**Status:** Notification likes, profile stats fixes, time filters, and CI pipeline fixes.

**Completed:**
- [x] **Notification Likes Feature**
  - [x] Domain layer (NotificationLike entity and repository interface)
  - [x] Repository layer with JOIN queries for user details
  - [x] Service layer with LikeNotification marking notifications as unread
  - [x] Handler layer with like/unlike/get likes endpoints
  - [x] Frontend NotificationLikes.vue component
  - [x] Integration into NotificationsView.vue
  - [x] Database migration 0.16.0 with CASCADE DELETE
  - [x] Multi-database support (SQLite, PostgreSQL, MariaDB)

- [x] **Social Engagement Features**
  - [x] Users can like any notification (PR achievements, announcements, streaks, milestones)
  - [x] Only original recipient sees like count and list of likers
  - [x] Liking marks notification as unread for recipient
  - [x] Users CAN like their own notifications
  - [x] Display: Thumbs up icon with count, comma-separated liker names
  - [x] "Liked by: " prefix for liker names
  - [x] CASCADE DELETE when notification is deleted

- [x] **Profile View Workout Summary Fixes**
  - [x] Fixed Personal Records count (was looking for wrong field)
  - [x] Fixed streak calculation algorithm (consecutive days logic error)
  - [x] Added null safety for API responses
  - [x] Time period filters: This Week, This Month, This Year, All Time
  - [x] Default period: This Month
  - [x] Reactive updates with Vue watch

- [x] **CI Pipeline Fixes**
  - [x] Fixed mockEmailService missing SendHTMLEmail method
  - [x] Fixed NewUserWorkoutService calls missing 4 parameters in tests
  - [x] Added mockMovementRepo with 13 methods
  - [x] Added GetAllPerformancesForMovement to mockUserWorkoutMovementRepo
  - [x] All service tests now compile successfully
  - [x] Integration tests compile successfully

**Files Created:** `internal/domain/notification_like.go`, `internal/repository/notification_like_repository.go`, `internal/service/notification_like_service.go`, `internal/handler/notification_like_handler.go`, `web/src/components/NotificationLikes.vue`

**Files Modified:**
- Notification likes: `internal/repository/migrations.go`, `internal/domain/notification.go`, `internal/repository/notification_repository.go`, `cmd/actalog/main.go`, `web/src/views/NotificationsView.vue`, `pkg/version/version.go`
- Profile stats: `web/src/views/ProfileView.vue`
- CI fixes: `internal/service/user_service_test.go`, `internal/service/user_workout_service_test.go`, `internal/service/test_helpers.go`
- Documentation: `docs/CHANGELOG.md`, `docs/TODO.md`, `docs/ROADMAP.md`, `docs/DATABASE_SCHEMA.md`, `CLAUDE.md`, `web/package.json`

**API Endpoints:**
- `POST /api/notifications/{id}/like` - Like a notification
- `DELETE /api/notifications/{id}/like` - Unlike a notification
- `GET /api/notifications/{id}/likes` - Get all likes with user details

---

### v0.15.0-beta (2025-12-19)

**Status:** Admin announcement system for gym-wide notifications.

**Completed:**
- [x] **Admin Announcement Feature**
  - [x] Admin-only endpoint for creating announcements
  - [x] Sends notification to all users in the system
  - [x] Flexible notification system (PR achievements, announcements, etc.)
  - [x] Audit trail for all announcement creation

**Files Created:** Admin announcement handler endpoint

**Files Modified:** `internal/handler/notification_handler.go` (admin announcement endpoint), `cmd/actalog/main.go` (route wiring)

**API Endpoints:**
- `POST /api/admin/notifications/announce` - Create announcement for all users (admin only)

---

### v0.14.0-beta (2024-12-16)

**Status:** Subscription billing system implementation with dual-level (user + organization) billing.

**Completed:**
- [x] **Subscription Billing System Backend**
  - [x] Domain entities (UserSubscription, OrganizationSubscription, SubscriptionAccessResult)
  - [x] Repository layer (3 repositories: UserSubscription, OrganizationSubscription, SubscriptionAccess)
  - [x] Service layer (SubscriptionService with admin operations)
  - [x] Middleware layer (RequireActiveSubscription - read-only enforcement)
  - [x] Handler layer (10 API endpoints: 8 admin, 2 user)
  - [x] Route configuration (wired into main.go)

- [x] **Migration 0.14.0**
  - [x] Create user_subscriptions table (15 columns, 4 indexes)
  - [x] Create organization_subscriptions table (15 columns, 4 indexes)
  - [x] Seed existing users with permanent free subscriptions
  - [x] Multi-database support (SQLite, PostgreSQL, MariaDB)

- [x] **Database Version Management System**
  - [x] SQLite snapshot: `db_versions/actalog_0.14.0.db` (564 KB)
  - [x] PostgreSQL schema: `actalog_0_14_0` on 192.168.1.143
  - [x] MariaDB database: `actalog_0_14_0` on 192.168.1.234
  - [x] Automation scripts: `create-db-snapshot.sh`, `verify-version-databases.sh`
  - [x] Documentation: `VERSION_DATABASES.md`, `MIGRATION_TEST_0.14.0.md`

- [x] **Subscription Features**
  - [x] Three subscription types: Free, Monthly, Annual
  - [x] Permanent Free option (never expires, for founders/staff)
  - [x] Dual-level billing (user-level AND organization-level)
  - [x] Flexible access (access if EITHER personal OR org subscription active)
  - [x] Manual admin payment control (mark as paid/unpaid)
  - [x] Immediate read-only mode when expired (no grace period)
  - [x] HTTP 402 Payment Required for blocked operations
  - [x] Complete audit trail for all subscription operations

- [x] **Backward Compatibility**
  - [x] All existing users seeded with permanent free subscriptions
  - [x] Zero downtime deployment verified
  - [x] Migration tested on all 3 database engines

**Files Created:** `internal/domain/subscription.go`, `internal/repository/user_subscription_repository.go`, `internal/repository/organization_subscription_repository.go`, `internal/repository/subscription_access_repository.go`, `internal/service/subscription_service.go`, `pkg/middleware/subscription.go`, `internal/handler/subscription_handler.go`, `db_versions/README.md`, `db_versions/VERSION_DATABASES.md`, `db_versions/MIGRATION_TEST_0.14.0.md`, `scripts/create-db-snapshot.sh`, `scripts/verify-version-databases.sh`

**Files Modified:** `internal/repository/migrations.go` (migration 0.14.0), `pkg/version/version.go` (v0.14.0), `cmd/actalog/main.go` (wiring), `internal/domain/audit_log.go` (subscription events), `CLAUDE.md` (version management)

**API Endpoints:**
- User: `GET /api/subscriptions/status`
- Admin: `POST /api/admin/subscriptions/user`, `GET /api/admin/subscriptions/user/{user_id}`, `POST /api/admin/subscriptions/user/{id}/mark-paid`, `POST /api/admin/subscriptions/user/{id}/cancel`
- Admin Org: `POST /api/admin/subscriptions/organization`, `GET /api/admin/subscriptions/organization/{org_id}`, `POST /api/admin/subscriptions/organization/{id}/mark-paid`, `POST /api/admin/subscriptions/organization/{id}/cancel`

---

### v0.12.2-beta (2025-11-28)

**Status:** PWA offline functionality fix and user-controlled updates.

**Completed:**
- [x] **PWA Offline Workout Recording**
  - [x] Fixed service worker API caching pattern
  - [x] Added robust offline detection (Network Error, ERR_NETWORK, navigator.onLine, timeout)
  - [x] Extended offline handling to support PUT requests
  - [x] Added 24-hour cache for API responses when offline

- [x] **User-Controlled PWA Updates**
  - [x] Replaced silent auto-reload with user prompt
  - [x] New `UpdatePrompt.vue` component with "Later" and "Update Now" buttons
  - [x] New `pwa.js` Pinia store for PWA state management

- [x] **Offline Save Notification**
  - [x] Added "Saved Offline" snackbar notification
  - [x] Custom `offline-save` event for UI notification

- [x] **Unit Test Fixes**
  - [x] Fixed `mockWODRepo.GetByName()` return value
  - [x] Updated WOD tests with correct error types
  - [x] Added required fields to Update tests

**Files:** `web/src/components/UpdatePrompt.vue`, `web/src/stores/pwa.js`, `web/vite.config.js`, `web/src/utils/axios.js`, `web/src/App.vue`, `web/src/main.js`, `internal/service/test_helpers.go`, `internal/service/wod_service_test.go`

---

### v0.12.1-beta (2025-11-28)

**Status:** MySQL/MariaDB compatibility fix and Docker troubleshooting.

**Completed:**
- [x] Fixed database-agnostic timestamp functions for MySQL/MariaDB
- [x] Fixed hardcoded SQLite `datetime('now')` in refresh token repository
- [x] Added `getTimestampFunc()` helper for cross-database timestamp support
- [x] Enhanced Docker host database troubleshooting documentation

**Files:** `internal/repository/database.go`, `internal/repository/refresh_token_repository.go`, `docker/DOCKER.md`, `docker/DATABASE_DEPLOYMENT.md`

---

### v0.12.0-beta (2025-11-26)

**Status:** Mobile PWA stability and Docker metadata improvements.

**Completed:**
- [x] Mobile PWA overflow fix across 27 view files
- [x] `.mobile-view-wrapper` CSS pattern for consistent mobile layouts
- [x] OCI-compliant labels added to Docker build scripts
- [x] Admin User Content view Actions column moved to first position
- [x] iOS PWA safe-area handling enhanced

---

### v0.11.0-beta (2025-11-26)

**Status:** Data Change Audit Logging system.

**Completed:**
- [x] Complete audit trail for data modifications (WOD, Movement)
- [x] Before/after values stored as JSON
- [x] Admin UI for viewing and filtering data change logs
- [x] Multi-database support (SQLite, PostgreSQL, MariaDB)

**Files:** `internal/domain/data_change_log.go`, `internal/repository/data_change_log_repository.go`, `internal/service/data_change_log_service.go`, `internal/handler/data_change_log_handler.go`, `web/src/views/AdminDataChangeLogsView.vue`

---

### v0.10.0-beta (2025-11-24)

**Status:** Docker deployment infrastructure with automatic seed import.

**Completed:**
- [x] Multi-stage Dockerfile with optimized build
- [x] Three docker-compose configurations (SQLite, PostgreSQL, MariaDB)
- [x] GitHub Actions CI/CD for automated image building
- [x] Automatic seed data import (182 movements, 314 WODs)
- [x] GitHub Container Registry integration (ghcr.io)

---

*For releases prior to v0.10.0, see [CHANGELOG.md](./CHANGELOG.md)*

---

## Future Considerations

These are longer-term ideas that may or may not be implemented:

- **Kubernetes manifests** - For larger deployments
- **Mobile native apps** - React Native or Flutter wrappers
- **API rate limiting** - For multi-tenant deployments
- **Webhook integrations** - Connect to external services
- **Multi-language support** - i18n framework

---

## Technical Debt

Items to address when time permits:

- [ ] Refactor large view components (DashboardView, PerformanceView) into smaller sub-components
- [ ] Add comprehensive API documentation (OpenAPI/Swagger)
- [x] Improve error handling consistency across handlers (centralized in internal/handler/errors.go)
- [x] Add structured logging throughout the codebase (JSON format support in pkg/logger)
- [x] Review and optimize database queries with EXPLAIN (audit_logs indexes, N+1 fixes)

---

*This file is maintained by Claude Code. See CHANGELOG.md for complete release history.*
