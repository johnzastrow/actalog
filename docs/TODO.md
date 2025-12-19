# ActaLog TODO

> **Last Updated:** 2025-12-19
> **Current Version:** 0.14.0-beta

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

### CI/CD Workflow Failures (In Progress)

**Status:** Multiple workflows failing due to test compilation errors and deprecated actions

#### GitHub Actions Status (as of 2025-12-19 17:09 UTC)
- ❌ **CI Workflow** - Failing (Run #20377062570)
- ❌ **Publish Site** - Failing (Run #20377062579)
- ✅ **Docker Build** - Passing
- ❌ **CI Failure Notify** - Failing (dependency on CI)

#### 1. Integration Tests Compilation Errors (`test/integration/api_test.go`)
**Status:** Blocking CI pipeline

**Errors:**
- Line 149-162: `NewUserService` missing `userSubRepo` parameter
  - Need to add `nil` for userSubRepo before audit log service parameter
- Line 177-184: `NewUserWorkoutService` missing `auditLogRepo` parameter
  - Need to add `nil` as 7th parameter

**Impact:** All 3 DB matrix tests failing (SQLite, PostgreSQL, MySQL)

#### 2. GitHub Pages Publish Workflow Failure
**Status:** Using deprecated GitHub Actions

**Error:**
```
This request has been automatically failed because it uses a
deprecated version of `actions/upload-artifact: v3`
```

**Files to Update:**
- `.github/workflows/publish-site.yml` line 25:
  - Current: `actions/upload-pages-artifact@v1`
  - Need: `actions/upload-pages-artifact@v3` (or latest)
- Check if `actions/deploy-pages@v1` also needs update

**Reference:** https://github.blog/changelog/2024-04-16-deprecation-notice-v3-of-the-artifact-actions/

#### 3. Unit Test Assertion Failures (Non-blocking)

**Subscription Service Tests** (`internal/service/subscription_service_test.go`)
**Compilation:** ✅ Fixed
**Running:** ⚠️ 10/14 tests passing

Failing tests:
1. **successful_free_subscription_creation** (line 173)
   - Issue: Test expects EndDate to be set for non-permanent free subscriptions
   - Location: `subscription_service_test.go:173`

2. **subscription_not_found scenarios** (lines 307, 400)
   - Issue: Tests expect exact error "subscription not found" but get wrapped error "failed to get subscription: sql: no rows in result set"
   - Locations: `subscription_service_test.go:307, :400`

3. **organization_already_has_active_subscription** (line 621)
   - Issue: Test expects wrong error message (says "user already has..." instead of "organization already has...")
   - Location: `subscription_service_test.go:621`

**Mock Repository Fixes Completed:**
- ✅ Fixed `mockAuditLogRepo.DeleteOlderThan` signature (time.Time param)
- ✅ Fixed `mockAuditLogRepo.List` signature (AuditLogFilters param)
- ✅ Added `mockAuditLogRepo.GetByTargetUserID` method
- ✅ Added `mockAuditLogRepo.GetByUserID` method
- ✅ Fixed `mockUserRepo.IsAccountLocked` signature (int64 param, returns bool, *time.Time, error)
- ✅ Fixed `mockUserRepo.LockAccount` signature (int64, time.Duration params)
- ✅ Added `mockOrganizationRepo.GetUserOrganizationIDs` method
- ✅ Fixed `mockOrganizationRepo.List` signature (returns count)
- ✅ Added `mockUserWorkoutRepo.GetActiveUsersThisMonth` method
- ✅ Fixed WOD service test calls (added userEmail parameter)
- ✅ Fixed UserWorkout service test calls (added userEmail parameter)
- ✅ Fixed UserService test (added userSubRepo parameter)

**Files Modified:**
- `internal/service/test_helpers.go` - All mock repository fixes
- `internal/service/subscription_service_test.go` - Comprehensive test suite (created)
- `internal/service/wod_service_test.go` - Added audit log repo parameter
- `internal/service/user_workout_service_test.go` - Added audit log repo, userEmail parameters
- `internal/service/user_service_test.go` - Added userSubRepo parameter

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

#### Subscription System (Backend Complete - v0.14.0)
- [ ] `[HIGH]` **Frontend Subscription Status Display**
  - [ ] Add subscription status badge to user profile/settings
  - [ ] Show expiration date for paid subscriptions
  - [ ] Display "Permanent Free" badge for permanent users
  - [ ] Show organization subscriptions the user benefits from

- [ ] `[HIGH]` **Admin Subscription Management UI**
  - [ ] Admin panel for viewing all user subscriptions
  - [ ] Admin panel for viewing all organization subscriptions
  - [ ] Create subscription form (user/org selection, type, permanent free option)
  - [ ] Mark subscription as paid button/action
  - [ ] Cancel subscription with reason input
  - [ ] View subscription history for users and organizations
  - [ ] List expiring subscriptions (next 30 days)
  - [ ] List overdue/expired subscriptions

- [ ] `[HIGH]` **Read-Only Mode UI Feedback**
  - [ ] Graceful handling of HTTP 402 Payment Required responses
  - [ ] Show "Subscription Expired" banner when in read-only mode
  - [ ] Disable/hide create/edit buttons when subscription expired
  - [ ] "Renew Subscription" call-to-action in banner
  - [ ] Toast notifications explaining why operations are blocked


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


#### Testing Coverage
- [x] `[HIGH]` **Subscription Service Tests** - Comprehensive test suite created (14 test cases)
  - Files: `internal/service/subscription_service_test.go`
  - Status: 10/14 passing, minor assertion fixes needed (see Active Tasks)
- [ ] `[HIGH]` **Add handler unit tests** - auth_handler, user_workout_handler, movement_handler, wod_handler, subscription_handler
- [ ] `[HIGH]` **Add service tests** - movement_service, workout_service, workout_template_service
- [ ] `[HIGH]` **Add repository unit tests** - All repository implementations

#### Admin Features
- [ ] `[HIGH]` **User Import/Export System** (Admin only)
  - [ ] Export users to CSV format
  - [ ] Import users from CSV (bulk user creation)
  - [ ] Preview workflow with validation
  - [ ] Duplicate detection by email
  - [ ] Welcome emails with password reset

### Medium Priority

#### Performance & Analytics
- [X] `[MEDIUM]` **Calendar View** - Monthly view with workout dots
- [X] `[MEDIUM]` **Timeline View** - Chronological workout history
- [ ] `[MEDIUM]` **Admin Metrics Dashboard** - User stats, workout counts, system health
- [ ] `[MEDIUM]` **PR Leaderboards** - Opt-in community leaderboards

#### Testing
- [ ] `[MEDIUM]` **Add backup_service tests**
- [ ] `[MEDIUM]` **Add export/import_service tests**
- [ ] `[MEDIUM]` **Add admin_handler tests**

#### PWA Enhancements
- [ ] `[MEDIUM]` Run Lighthouse PWA audit
- [ ] `[MEDIUM]` Optimize service worker cache size
- [ ] `[MEDIUM]` Test offline sync end-to-end on mobile devices

### Low Priority

#### Testing
- [ ] `[LOW]` **Add audit_log_service tests**
- [ ] `[LOW]` **Add wodify_import_service tests**

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
| `internal/service/import_service.go` | 587, 701 | Add duplicate detection using userWorkoutRepo.ListByUserAndDateRange |
| `internal/handler/movement_handler.go` | 141 | Get user ID from context when auth middleware is added |

### Frontend (Vue)

| File | Line | Description |
|------|------|-------------|
| `web/src/views/WorkoutDetailView.vue` | 409 | Implement edit workout functionality |
| `web/src/views/SettingsView.vue` | 511 | Apply theme change (dark mode toggle) |
| `web/src/views/SettingsView.vue` | 540 | Implement import functionality |
| `web/src/views/WorkoutsView.vue` | 372 | Navigate to template detail page |

Add scrolling month calendar screen from the profile menu that shows dots for workouts that can be clicked to see workout details. More than one workout per day.


Add notifications: notify all users in the gym on the following events: 
1. What user gets a PR. Show the date, word/movement and Congratulations text
2. When a user gets 4 or more workouts in a week. Show date, user, workout count and congratulations 
3. What user Completes a rolling 10 of any type of WOD (hero, girl, etc). For example, every 10 hero wods 

*Last scanned: 2025-11-28*

---

## Completed Releases

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
- [ ] Improve error handling consistency across handlers
- [ ] Add structured logging throughout the codebase
- [ ] Review and optimize database queries with EXPLAIN

---

*This file is maintained by Claude Code. See CHANGELOG.md for complete release history.*
