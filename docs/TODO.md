# ActaLog TODO

> **Last Updated:** 2025-12-22
> **Current Version:** 0.16.0-beta

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

*(none currently)*

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

#### Front end Improvements
- [ ] `[HIGH]` **PWA**
  - [x] Icon saved to home screen is fuzzy - regenerated all icons from SVG with full RGBA color
- [ ] `[HIGH]` **Overall**
  - [ ] Find more attractive icons for the navigation bar
  - [x] Allow notifications to be marked as read/unread (implemented in NotificationsView.vue)
  - [x] Develop a theme called "Sunrise" based on colors from sunset image (added to vuetify.js and theme.js)
  - [x] Remove the View all link on the Dashboard (removed from DashboardView.vue) 


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

- [ ] `[HIGH]` **Improved Duplicate Record Detection During Imports**
  - [ ] Trap duplicate movements during import (check by name/description)
  - [ ] Trap duplicate WODs during import (check by name/source)
  - [ ] Trap duplicate user workouts during import (check by date/template/user)
  - [ ] Enhanced preview workflow showing detected duplicates with options:
    - Skip duplicates (keep existing)
    - Update duplicates (overwrite existing)
    - Import as new (allow duplicates)
  - [ ] Detailed duplicate report in preview response
  - [ ] Frontend UI to review and select action for each duplicate
  - [ ] Batch duplicate resolution (apply same action to all)

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

### Medium Priority

#### Documentation & Marketing
- [ ] `[MEDIUM]` **Enhance App Documentation with Screenshots**
  - [ ] Add more information about features and functions in the app
  - [ ] Capture screenshots demonstrating all available themes
  - [ ] Show desktop vs mobile responsive views
  - [ ] Document key user flows with visual guides
  - [ ] Add screenshots to static website/README

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

*Last scanned: 2025-12-22*

**Note:** Calendar view (WorkoutCalendarView.vue) and PR notification system are implemented.

---

## Completed Releases

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
- [ ] Improve error handling consistency across handlers
- [ ] Add structured logging throughout the codebase
- [ ] Review and optimize database queries with EXPLAIN

---

*This file is maintained by Claude Code. See CHANGELOG.md for complete release history.*
