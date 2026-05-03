# ActaLog TODO

> **Last Updated:** 2026-04-02
> **Current Version:** 1.2.1 (Build 44)

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

### Class Scheduling System *(Completed)*

**Status:** Phase 1-4 Complete

**Phase 1-3 (v0.26.0 migration):**
- [x] Gym locations, class templates, schedule slots
- [x] Class sessions with capacity management
- [x] Coach assignments (per-gym role)
- [x] Reservations with check-in flow

**Phase 4 (v0.27.0 migration):**
- [x] Documents - Required documents (waivers, liability forms) per gym
- [x] User Documents - Track document completion status per user
- [x] Class Packages - Credit packages (e.g., "10-Class Pack")
- [x] User Credits - Credit balances with expiration tracking
- [x] Waitlist - Queue management when classes are full
- [x] Class Notifications - Reminders and waitlist promotions

**Frontend Views:**
- [x] `ScheduleView.vue` - Browse sessions, reserve, join/leave waitlist
- [x] `MyCreditsView.vue` - User's credits, documents, waitlist entries
- [x] `MyReservationsView.vue` - User's upcoming reservations
- [x] `AdminPackagesView.vue` - Admin management for packages, documents, user credits
- [x] `AdminSchedulingView.vue` - Admin session/template management

**API Endpoints (Phase 4):**
- Documents: `GET/POST/PUT/DELETE /api/admin/gyms/{id}/documents`
- Packages: `GET/POST/PUT/DELETE /api/admin/gyms/{id}/packages`
- Credits: `POST /api/admin/gyms/{id}/users/{id}/credits`, `GET /api/gyms/{id}/users/me/credits`
- Waitlist: `POST/DELETE /api/sessions/{id}/waitlist`, `GET /api/users/me/waitlist`
- User documents: `GET /api/gyms/{id}/users/me/documents`

**Database Tables (6 new in v0.27.0):**
- `documents`, `user_documents`, `class_packages`, `user_class_credits`, `waitlist_entries`, `class_notifications`

**Bug Fix:** PurchaseCredits now correctly uses package credits when `package_id` is provided without explicit `credits` value.

**Tested on:** SQLite, MariaDB (192.168.1.234), PostgreSQL (192.168.1.143)

---

### Delete Class with Sessions *(Completed)*

**Status:** Complete (Build 31)

- [x] Three delete modes for class templates:
  - `template_only` - Delete template only, sessions orphaned
  - `with_future_sessions` - Delete template + future sessions, keep past
  - `with_all_sessions` - Delete template + all sessions
- [x] Cascade deletion (coaches, reservations, waitlist, notifications)
- [x] Credit refunds for unconfirmed reservations
- [x] User notifications for cancelled sessions
- [x] New UI dialog with mode selection and session counts
- [x] API: `DELETE /api/admin/scheduling/templates/{id}?mode=<mode>`

**Files Modified:**
- `internal/domain/scheduling.go`, `internal/domain/phase4.go` - Interface methods
- `internal/repository/class_session_repository.go` - Bulk delete methods
- `internal/repository/session_coach_repository.go` - DeleteBySessionIDs
- `internal/repository/reservation_repository.go` - GetBySessionIDs, DeleteBySessionIDs
- `internal/repository/waitlist_repository.go` - DeleteBySessionIDs
- `internal/repository/credits_repository.go` - RefundCredit
- `internal/service/scheduling_service.go` - DeleteTemplateWithMode
- `internal/handler/scheduling_handler.go` - Mode parameter handling
- `web/src/views/AdminSchedulingView.vue` - Delete dialog UI

**Tested on:** SQLite, MariaDB, PostgreSQL

---

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


#### Security (closed in v1.2.4 hardening branch)
- [x] `[HIGH]` **Security Response Headers Middleware** *(Completed v1.2.4)* — Added `pkg/middleware/security_headers.go` setting X-Frame-Options, X-Content-Type-Options, Referrer-Policy, HSTS, and a project-tuned CSP. Wired after CORS in `cmd/actalog/main.go`. Tests in `security_headers_test.go` assert per-header values and that script-src stays free of `'unsafe-inline'`/`'unsafe-eval'`.
- [x] `[HIGH]` **Avatar Upload Magic-Byte Validation** *(Completed v1.2.4)* — Replaced trust-the-header check with `http.DetectContentType()` over the first 512 bytes plus a `.jpg/.jpeg/.png/.gif/.webp` extension allowlist. Negative tests cover spoofed-Content-Type and disallowed-extension cases.

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

5. **Documentation Website**
   - Build a dedicated user and admin guide website from Markdown documents
   - Use a static site generator (e.g., MkDocs, Docusaurus, VitePress, Hugo) to convert existing Markdown docs into an attractive HTML site with navigation, search, and theming
   - Separate user guide (workout logging, schedules, PRs, leaderboards) and admin guide (user management, scheduling config, backups, subscriptions)
   - Host alongside the app or as a standalone site

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
- [x] **Comprehensive User Edit Screen — Profile tab + framework (v1.3.0)**
  - Shipped: Profile tab, four-layer defense-in-depth, recovery tooling, full documentation
  - Plan: `docs/superpowers/plans/2026-04-28-admin-user-edit-v1.3.0-plan.md`

- [ ] `[HIGH]` **User Edit Screen — Affiliations tab (v1.3.1)** — gym memberships, coach assignments per gym; add/remove org membership; manage `CoachAssignment` per `GymLocation` (assign/revoke coach role per gym); view `TemplateCoach` and `SessionCoach` rows that reference this user
- [ ] `[HIGH]` **User Edit Screen — Subscriptions / Credits / Preferences / Activity tabs (v1.3.2)** — complete the remaining tabs deferred from v1.3.0: Subscriptions, Class Credits & Documents, Preferences (UserSettings), and read-only Activity & Audit summary

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

- [x] `[HIGH]` **Database Duplicate Detection and Cleaning Procedure** *(Completed v0.24.0)*
  - [x] Create admin tool to scan for existing duplicates in database
  - [x] Detect duplicate movements (case-insensitive name matching)
  - [x] Detect duplicate WODs (case-insensitive name matching)
  - [x] Detect duplicate user workouts (same user, same date, same name)
  - [x] Detect duplicate users (case-insensitive email matching)
  - [x] Detect duplicate workout templates (same name, same user)
  - [x] Generate duplicate report with:
    - Count of duplicates per entity type
    - List of duplicate groups with record IDs
    - FK reference counts for each record
  - [x] Safe merge/cleanup procedure:
    - Preview which records will be kept vs deleted
    - Automatically update foreign key references
    - Delete child records from duplicates (movements/WODs)
    - Preserve data integrity with transaction support
  - [x] Admin UI to review and approve cleanup (AdminDataQualityView.vue)
  - [x] Preview mode shows FK impact before confirming
  - [x] Audit log of all merge operations (duplicate_merge event type)
  - [x] **Data Quality Issue Detection** - 4 quality checks:
    - Orphaned FK references (error severity)
    - Empty required fields (error/warning severity)
    - Future workout dates (warning severity)
    - Invalid email formats (warning severity)
  - **Backend files:** `internal/service/data_quality_service.go`, `internal/handler/data_quality_handler.go`
  - **Frontend:** `web/src/views/AdminDataQualityView.vue` with 3 tabs (Overview, Duplicates, Data Issues)
  - **API endpoints:**
    - `GET /api/admin/data-quality/duplicates` - Scan all entities
    - `GET /api/admin/data-quality/duplicates/summary` - Quick summary
    - `GET /api/admin/data-quality/duplicates/{entity}` - Scan specific entity
    - `POST /api/admin/data-quality/duplicates/merge/preview` - Preview merge
    - `POST /api/admin/data-quality/duplicates/merge/confirm` - Execute merge
    - `GET /api/admin/data-quality/issues` - Scan for data quality issues

### High Priority

#### Documentation & Marketing
- [x] `[High]` **Enhance App Documentation with Screenshots**
- [x] `[High]` Static site deployed to GitHub Pages (https://johnzastrow.github.io/actalog/)
  - [x] Home page with feature highlights and theme gallery
  - [x] For Admins page with subscription and announcement features
  - [x] For Developers page with tech stack, architecture, and deployment guide
  - [x] Performance & Benchmarks section added (v0.22.0) - minimum resources, load test results
  - [x] FAQ page
  - [x] Auto-deploy via GitHub Actions on push to main
  - [x] All 7 themes documented with screenshots (Myst added v0.24.0)
  - [x] Responsive design & PWA benefits explained in "Works Everywhere" section
  - [x] User flow guide added ("Your Workout Journey" 3-step visual guide)


#### Performance & Analytics
- [x] `[MEDIUM]` **Calendar View** - WorkoutCalendarView.vue with workout dots and month navigation
- [x] `[MEDIUM]` **Timeline View** - WorkoutTimelineView.vue with chronological history
- [x] `[MEDIUM]` **Progress Charts** - WeightProgressChart.vue and WorkoutFrequencyChart.vue components (not yet integrated into views)
- [x] `[MEDIUM]` **Admin Metrics Dashboard** - User stats, workout counts, system health (v0.22.0)
- [x] `[MEDIUM]` **1RM Percentage Display** - Performance page 1RM percentage tops out at 95%; needs to also show 100% of the maximum weight lifted
- [x] `[MEDIUM]` **PR Leaderboards** - Opt-in community leaderboards (migration 0.33.0, leaderboard_opt_in setting, movement/WOD leaderboard API endpoints, LeaderboardView.vue)

#### Testing
- [x] `[MEDIUM]` **Add backup_service tests** - backup_service_test.go with 10 test functions covering create/restore/export
- [x] `[MEDIUM]` **Add wodify_import_service tests** - wodify_import_service_test.go with duplicate handling
- [ ] `[MEDIUM]` **Add export/import_service tests** - Currently at 60-83% coverage
- [ ] `[MEDIUM]` **Add admin_handler tests**

#### PWA Enhancements *(Audited v0.24.0)*
- [x] `[MEDIUM]` **Lighthouse PWA audit completed** (Jan 2026)
  - Performance: 56/100 (throttled network), Best Practices: 100/100, SEO: 92/100
  - PWA features verified: service worker, offline caching, installable manifest
  - Issues found: color contrast, meta viewport user-scalable, touch target sizes
- [x] `[MEDIUM]` **Service worker cache analysis completed**
  - Total precache: 6.9MB (150+ entries)
  - Material Design Icons: 3.5MB (4 formats) → **Optimization: remove ttf/eot, keep woff2 only (394KB)**
  - Self-hosted accessibility fonts: 848KB (9 families, 22 files)
  - JavaScript: 1.5MB (properly chunked), CSS: 848KB
  - Runtime caching: API responses (NetworkFirst), fonts (CacheFirst), uploads (StaleWhileRevalidate)
- [x] `[MEDIUM]` **Offline sync implementation reviewed**
  - IndexedDB storage for workouts, movements, pending sync queue
  - Axios interceptor saves POST/PUT /api/workouts offline when network unavailable
  - syncWithServer() replays pending operations when back online
  - User-controlled updates via UpdatePrompt.vue and pwa.js store
  - **Status: Fully implemented, works for workout logging**

**Remaining PWA optimizations (deferred):**
- [x] `[LOW]` **Remove unused MDI font formats** - Custom Vite plugin `mdiWoff2Only()` strips ttf/eot/woff
  - Dist size: 6.9MB → 3.8MB (45% reduction)
  - Precache: 6.9MB → 3.4MB (50% reduction)
  - Only woff2 (394KB) shipped to production
- [x] `[LOW]` **Lazy-load accessibility fonts** - Fonts loaded on-demand when user selects them
  - Split fonts.css into 9 separate font family CSS files
  - Font store dynamically imports CSS via Vite's dynamic import
  - Fonts excluded from precache via globIgnores
  - Runtime caching (CacheFirst) caches fonts on first use
  - Precache: 3.4MB → 2.6MB (additional 24% reduction)
- [x] `[LOW]` Fix accessibility issues (color contrast, touch targets) - Fixed label contrast (#666666 for 5.74:1 ratio), touch target sizes (24x24px min)

#### Roles & Permissions
- [x] `[HIGH]` **Coach Role** *(Completed Build 36)* - Added dedicated `coach` role (migration 0.32.0). Renamed `"user"` to `"athlete"`. Three-tier role system: athlete < coach < admin. `CoachOrAdmin` middleware protects `/api/coaches/` routes. Coaches bypass subscription checks. Coach Dashboard uses dedicated `/api/coaches/sessions/` endpoints with org-level access verification. Bottom nav shows Coach button for coach/admin roles when scheduling enabled.
  - **Admin all-sessions visibility:** Admins see ALL upcoming sessions across all gyms on Coach Dashboard (not limited to coach assignments). Added `GetAllUpcoming` repo method and `GetAllUpcomingSessions` service method; handler branches on role.
  - **Comprehensive role/access middleware tests:** 14 new test functions in `pkg/middleware/auth_test.go` and `pkg/middleware/subscription_test.go` covering all role permutations for `AdminOnly`, `CoachOrAdmin`, and `RequireActiveSubscription` middleware. Includes full JWT→middleware integration tests.
  - **Files modified:** `internal/domain/scheduling.go`, `internal/repository/class_session_repository.go`, `internal/service/scheduling_service.go`, `internal/handler/scheduling_handler.go`, `pkg/middleware/auth_test.go`, `pkg/middleware/subscription_test.go`
  - **Docker images pushed:** `ghcr.io/johnzastrow/actalog:dev`, `:latest`, `:1.1.0-beta` (Build 36)

### Low Priority

#### Testing
- [x] `[LOW]` **Add audit_log_service tests** - 100% coverage achieved
- [x] `[LOW]` **Add wodify_import_service tests** - 100% coverage achieved

#### Features
- [x] `[MEDIUM]` **Admin Screen Consolidation (Partial)** - Grouped Profile screen's 16 admin links into 5 labeled categories (Users & Access, Scheduling, Communication, Data & System, Organizations) using `v-list-subheader` dividers. Removed disabled "System Reports" placeholder. Full tab consolidation of admin screens into complex tabbed views remains for future work.
- [x] `[MEDIUM]` **Consistency Achievement Notifications** - Daily scheduler checks all users for milestones (50/100/200/300 workouts, 4+/week, day streaks 4/7/14/21/30, inactivity warnings). Migration 0.34.0 consistency_achievements table, dedup via UNIQUE constraint, notification icons in NotificationsView.
- [ ] `[LOW]` **Push Notifications** - Workout reminders
- [ ] `[LOW]` **Data Visualization** - Charts for PR progression
- [ ] `[LOW]` **Social Features** - Share workouts (opt-in)
- [x] `[MEDIUM]` **UI Styling Consistency Review (Partial)** - Replaced hardcoded `#2c3e50` inline color styles with Vuetify theme classes (`text-h6`, `text-subtitle-1`) in AdminDataQualityView (4 occurrences), AdminDataCleanupView (3), and AdminView (1) for dark mode compatibility. Remaining screens (WorkoutCalendarView, WorkoutTimelineView fixed headers; PerformanceView Chart.js config) flagged for future review.
- [x] `[MEDIUM]` **Admin Breadcrumbs & Navigation Consistency (Partial)** - Fixed AdminSchedulingView to use shared AdminHeader component instead of custom gradient header, adding breadcrumb navigation (Admin > Scheduling) consistent with all other 17 admin views. Full audit of remaining admin views for navigation consistency remains for future work.

#### Technical Debt
- [ ] `[LOW]` **Multiple Save Issue** - TemplateEditDialog save() is being called 3x instead of 1x; investigate root cause (possibly related to Vue reactivity with Schedule tab slots)

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

### v1.2.4 (2026-04-28 — Security Hardening, part two)

**Status:** Released from `security/headers-and-uploads-2026-04-28`. Closes the two items deferred from v1.2.3.

**Completed:**
- [x] **Security** — Security response headers middleware (`pkg/middleware/security_headers.go`) sets X-Frame-Options, X-Content-Type-Options, Referrer-Policy, HSTS, and a project-tuned CSP on every response; wired after CORS in `cmd/actalog/main.go`
- [x] **Security** — Avatar upload no longer trusts client `Content-Type`; uses `http.DetectContentType` over first 512 bytes plus extension allowlist (`internal/handler/user_handler.go`)
- [x] **CI** — CI Failure Notify workflow now deduplicates issues per (workflow, branch) and auto-closes when a subsequent run on the same branch succeeds (`.github/workflows/ci-failure-notify.yml`)
- [x] **Maintenance** — 8 Dependabot bumps merged (deploy-pages, setup-buildx, lib/pq, x/crypto, pgx, vue-ecosystem, go-sqlite3, dev-dependencies group); 2 closed with policy comments (Vuetify 4.0.5 watch-and-wait, axios 1.14.0 supply-chain)

---

### v1.2.3 (2026-04-03 — Security Hardening)

**Status:** Released from `security/hardening-2026-04-03`.

**Completed:**
- [x] **Security** — Password policy raised to 12-char min + uppercase/lowercase/digit (`internal/service/user_service.go`); UI hints and client-side validation updated in RegisterView, SettingsView, ResetPasswordView, AdminUserImportExportView
- [x] **Security** — `WriteError()` no longer leaks raw internal error strings to HTTP responses
- [x] **Security** — DOMPurify 3.3.3 added to `MarkdownRenderer.vue`; all `v-html` output sanitized
- [x] **Security** — `serialize-javascript` pinned to 7.0.5 via `package.json` overrides (RCE + DoS CVEs)
- [x] **Security** — Rate limiter IP extraction fixed: leftmost XFF IP taken, RemoteAddr port stripped
- [x] **Security** — `rate_limit_exceeded` audit events now emitted from auth and password-reset limiters
- [x] **Security** — CORS allowlist now actually enforced — `pkg/middleware/cors.go` no longer echoes disallowed origins back as `Access-Control-Allow-Origin`; test asserts header absence for disallowed origins
- [x] **Feature** — Admin organizations list shows member count and supports edit-from-list
- [x] **CI** — golangci-lint upgraded to v2.11.4; config migrated to v2 format
- [x] **CI** — Frontend unit tests added to CI pipeline; 11 new Vitest suites covering App shell, auth store, axios interceptor, and admin views
- [x] **Docs** — `docs/OWASP_AUDIT_2026-04-03.md`, `docs/MATURITY_ASSESSMENT.md`, `docs/plans/SECURITY_HARDENING_PLAN.md` capture the audit and remediation plan

---

### v1.2.1 (2026-04-01)

**Status:** Patch release — security fixes, dependency maintenance, CI hardening.

**Completed:**
- [x] **Security** — Pinned axios to 1.13.5 (supply chain attack mitigation, 2026-03-31 incident)
- [x] **Security** — `npm audit fix`: patched brace-expansion, picomatch, undici, yaml
- [x] **Fix** — Calendar view now fetches all workouts (was capped at 20, causing missing entries)
- [x] **Fix** — Admin breadcrumb returning 404
- [x] **Fix** — Avatar upload click target; removed stale debug logging
- [x] **Feature** — Calendar header shows workout count for viewed month
- [x] **UI** — Nav bar icons switched to heavier filled MDI variants
- [x] **CI** — `setup-go` uses `go-version-file: go.mod` (auto-tracks toolchain version)
- [x] **CI** — Pinned `golangci-lint-action` to v6 (v9 dropped v1.x support)
- [x] **CI** — Fixed date-sensitive test failure on first day of each month
- [x] **Deps** — Go: x/crypto 0.48→0.49, pgx/v5 5.8→5.9, go-sqlite3 1.14.34→1.14.38, lib/pq 1.11→1.12
- [x] **Deps** — Frontend: Vue 3.5.28→3.5.31, vue-router 5.0.3→5.0.4, vitest 4.0→4.1, sass 1.97→1.98
- [x] **Deps** — GH Actions: docker/login-action 3→4, docker/metadata-action 5→6, docker/build-push-action 6→7, actions/checkout 4→6, super-linter 4→7

**Known issues:**
- `vite-plugin-pwa` dependency chain (serialize-javascript CVE) — fix requires major downgrade to 0.19.8; deferred to next release

---

### v1.2.0-beta (2026-02-19)

**Status:** PR Leaderboards, configurable beta logo, nav and UI polish.

**Completed:**
- [x] PR leaderboards and consistency achievement notifications
- [x] Configurable beta logo via `LOGO_VARIANT` env variable
- [x] Leaderboard search API response parsing fix and header nav icon
- [x] Improved avatar upload UX

---

### v1.1.0-beta (2026-01-22)

**Status:** Coach role, three-tier permissions, class deletion modes, scheduling UX.

**Completed:**
- [x] Three-tier role system: athlete < coach < admin (migration 0.32.0)
- [x] `CoachOrAdmin` middleware; dedicated `/api/coaches/` routes
- [x] Delete class template with configurable cascade (template only / future sessions / all sessions)
- [x] Credit refunds and notifications on session cancellation
- [x] Coach Dashboard — admins see all sessions across all gyms

---

### v0.24.0-beta (2026-01-14)

**Status:** Data Quality & Duplicate Detection system with merge functionality.

**Completed:**
- [x] **Data Quality Admin Dashboard**
  - [x] `AdminDataQualityView.vue` with 3 tabs (Overview, Duplicates, Data Issues)
  - [x] Full database scan with summary cards
  - [x] Duplicates by entity type breakdown
  - [x] Data quality checks with issue counts

- [x] **Duplicate Detection & Merge**
  - [x] Scan 5 entity types: movements, WODs, user_workouts, users, workouts
  - [x] Case-insensitive name/email matching
  - [x] Composite key matching for user_workouts (user_id + date + name)
  - [x] Preview merge with FK reference counts
  - [x] Safe merge with FK updates in transaction
  - [x] Audit logging of all merge operations

- [x] **Data Quality Issue Detection**
  - [x] Orphaned FK references (error severity)
  - [x] Empty required fields (error/warning severity)
  - [x] Future workout dates (warning severity)
  - [x] Invalid email formats (warning severity)
  - [x] Filter chips for issue type filtering

- [x] **Frontend UI Improvements**
  - [x] Quality check type cards with icons, descriptions, counts
  - [x] Zero-state display with success checkmarks
  - [x] Clickable cards to navigate to filtered views
  - [x] Data Quality link added to admin menu in ProfileView

- [x] **Demo Mode Configuration**
  - [x] Added DEMO MODE section to .env.example
  - [x] Documented implemented vs planned features
  - [x] Current workaround instructions for basic demo setup

**Files Created:** `internal/service/data_quality_service.go`, `internal/handler/data_quality_handler.go`, `web/src/views/AdminDataQualityView.vue`

**Files Modified:** `cmd/actalog/main.go` (routes), `web/src/router/index.js` (route), `web/src/views/ProfileView.vue` (admin menu link), `.env.example` (demo mode)

**API Endpoints:**
- `GET /api/admin/data-quality/duplicates` - Scan all entities for duplicates
- `GET /api/admin/data-quality/duplicates/summary` - Quick duplicate summary
- `GET /api/admin/data-quality/duplicates/{entity}` - Scan specific entity type
- `POST /api/admin/data-quality/duplicates/merge/preview` - Preview merge operation
- `POST /api/admin/data-quality/duplicates/merge/confirm` - Execute merge
- `GET /api/admin/data-quality/issues` - Scan for data quality issues

---

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

*For releases prior to v0.24.0, see [CHANGELOG.md](./CHANGELOG.md)*

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

## Dependency Upgrade Watch

### Vuetify 4 Migration — Watch & Wait

**Current:** Vuetify `^3.12.1` (pinned after accidental bump to 4.0.0 in Dependabot commit `039084d`)

**Trigger conditions** — upgrade when ANY of these are true:
- Vuetify 4.1.x or later is released (indicates post-launch stabilization)
- Vuetify 3.x enters security-only or end-of-life maintenance
- A feature we need is only available in Vuetify 4

**Watch:**
- Vuetify GitHub releases: `https://github.com/vuetifyjs/vuetify/releases`
- Vuetify 3 EOL announcement (none as of 2026-04-03)

**Known breaking changes for this app** (catalogued 2026-04-03):

| Area | Change | Fix Required |
|------|--------|-------------|
| CSS cascade | All Vuetify 4 CSS is in `@layer vuetify-components` — unlayered app CSS wins | Wrap `main.css` reset in `@layer reset { }` and remove `p, span, div { font-size: 14px }` from App.vue |
| `v-main` padding | CSS reset `* { padding: 0 }` kills layout top/bottom padding | Keep existing `paddingTop`/`paddingBottom` in `mainStyle` (already fixed) |
| `fill-height` | No longer sets `display: flex` or `align-items: center` — only `height: 100%` | Add `d-flex align-center` to 47 views using `v-container class="fill-height"` |
| `app` prop | Removed from `v-app-bar`, `v-bottom-navigation` — layout registration is now automatic | Remove `app` prop from both (harmless but dead code) |
| `v-bottom-navigation` | `v-model` controls selected tab only; `active` prop controls visibility separately | Audit any code that passed `v-model` expecting visibility control |

**Full plan:** See memory file `vuetify4-upgrade-plan.md`

---

*This file is maintained by Claude Code. See CHANGELOG.md for complete release history.*
