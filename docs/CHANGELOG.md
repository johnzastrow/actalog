# Changelog

All notable changes to ActaLog will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.1] - 2026-05-06 — Admin user lifecycle

### Added
- **`POST /api/admin/users`** — admins can create user accounts with email + password + role; new user can sign in immediately
- **`POST /api/admin/users/{id}/password`** — admin sets a specific password directly; bundles lockout-clear + refresh-token revocation in one operation; protected accounts blocked at L1 + L2
- **AdminUserCreateDialog** on the User Management screen — modal with email/password/name/role/email-verified inputs
- **AdminSetPasswordDialog** on the Profile tab Password Management card — sits alongside the existing Force Password Reset button with explanatory copy distinguishing the two
- **`usePasswordInputs`** composable — shared password+confirm state with show/hide toggle, complexity hint, matching validation; used by both new dialogs
- **Three new audit events**: `admin_user_created`, `admin_password_set` (with prior failed-attempts/lockout state captured for forensics), `admin_user_create_rejected_protected`

### Security
- Protected emails are rejected at create with their own audit event — distinct from the `protected_user_attack_*` family which targets modifications of existing protected rows
- L1 `ProtectedUserGuard` covers the new set-password endpoint automatically (inside the `/users/{id}` sub-router)
- L2 service-layer `ensureNotProtected` defensive check on `SetPassword` matches the pattern used by `UpdateProfile` / `ForcePasswordReset`
- Password complexity policy unchanged (12+ chars, upper, lower, digit) — single source of truth in `validatePassword`

### Documentation
- `docs/security/THREAT_MODEL.md` — admin-compromise residual-risk row added
- `docs/USER_PERMISSIONS.md` — new endpoints + UI surfaces

---

## [1.3.0] - 2026-05-03 — Admin user-edit screen + protected-user defense-in-depth

### Added
- **Admin User Edit Screen** at `/admin/users/:id/edit` — tabbed view for editing per-user attributes. v1.3.0 ships the **Profile** tab with name/email/birthday/email-verified field editing, optimistic-concurrency via `updated_at`, and a force-password-reset action that sends a reset email and revokes refresh tokens. Future tabs (Affiliations, Subscriptions, Credits, Preferences, Activity) shown as disabled with version chips.
- `PATCH /api/admin/users/{id}` — partial profile update with `updated_at` precondition; returns 409 on stale `updated_at`
- `POST /api/admin/users/{id}/force-password-reset` — sends password-reset email + revokes all refresh tokens
- Edit pencil-icon button in `AdminUsersView` (disabled for protected users)

### Security — Protected User System (defense-in-depth)
A four-layer defense for system-reserved accounts (currently `br8kwall@gmail.com`):
- **L1 — HTTP middleware** (`pkg/middleware/protected_user.go`): refuses non-GET writes under `/api/admin/users/{id}` with structured 403 + `protected_user_attack_http` audit event
- **L2 — Service guard** (`internal/service/admin_user_service.go`): catches in-process callers (cron, internal APIs) that bypass HTTP. Fires `protected_user_attack_service` audit event
- **L3 — Database trigger** (migration 0.35.0, per-dialect SQL for sqlite3/postgres/mysql): `BEFORE UPDATE` blocks identity-field changes only (`email`, `name`, `role`, `account_disabled`); `BEFORE DELETE` blocks unconditionally. Lifecycle writes (password_hash, last_login_at, email_verified, verification_token, ...) pass through so the protected user can register, log in, and rotate their own password — L1 and L2 remain the primary defenses against admin-screen tampering. Mixed identity+lifecycle UPDATEs are still blocked. See `docs/security/PROTECTED_USERS.md` §3.4 for the full contract
- **L4 — Audit log tagging** (`internal/domain/audit_log.go`): single event per attempt, tagged at the rejecting layer
- **Boot-time invariant**: three sub-checks (triggers exist, triggers fire, protected rows present). Fails closed on hard failure; soft-warn on fresh installs
- **`ACTALOG_SKIP_PROTECTED_INVARIANT=true`** env var: degraded-mode escape hatch. `/health` returns 503; admin user-write endpoints return 503; ERROR-level heartbeat every 60s for alerting
- **Admin CLI**: `actalog admin verify-protected-users [--verbose]` (read-only diagnostic) and `actalog admin reapply-protected-migrations --confirm` (idempotent recovery)
- **Recovery scripts**: `scripts/recover/restore-protected-triggers.sh` + per-dialect SQL files for last-resort recovery when the binary itself is unrunnable

### Documentation
- `docs/security/PROTECTED_USERS.md` — master runbook: purpose, threat model, architecture, list management, verification, recovery, audit forensics, degraded mode
- `docs/security/PROTECTED_USERS_RECOVERY.md` — focused 3-AM-incident playbook
- `docs/security/THREAT_MODEL.md` — app-wide threat model linking each section to its implementation
- `docs/superpowers/specs/2026-04-28-admin-user-edit-design.md` and `docs/superpowers/plans/2026-04-28-admin-user-edit-v1.3.0-plan.md` capture the design and implementation history

### CI
- Generator drift check: `go run ./cmd/gen-protected-emails/` + `git diff --exit-code` keeps Go source ↔ frontend JS in sync
- Lockstep check: migration SQL ↔ standalone recovery scripts can't drift
- Doc code-block lint: `bash -n` validates every shell command in security runbooks
- CODEOWNERS: security-critical paths (`pkg/security/`, migrations, security docs, recovery scripts) require explicit review

### Maintenance
- Migration 0.35.0 is additive (no schema changes, just triggers); existing v1.2.x deployments upgrade by applying the migration
- v1.2.x rollback is safe: triggers stay in the database, security stays on even if the binary rolls back

### Carried into next releases
- v1.3.1: Affiliations tab (gym memberships, coach assignments per gym)
- v1.3.2: Subscriptions, Credits & Documents, Preferences, Activity & Audit tabs

---

## [1.2.4] - 2026-04-28 — Security hardening, part two

Closes the two items deferred from v1.2.3 (`docs/plans/SECURITY_HARDENING_PLAN.md` Steps 2 and 5).

### Security
- **Security response headers middleware** — new `pkg/middleware/SecurityHeaders` sets `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: strict-origin-when-cross-origin`, `Strict-Transport-Security: max-age=31536000; includeSubDomains`, and a project-tuned `Content-Security-Policy` on every response. Wired in `cmd/actalog/main.go` after the CORS middleware so it applies to API, static assets, and `/health`. CSP allows `'unsafe-inline'` for *styles* (Vuetify dynamic theme injection requires it) but keeps scripts strict (`script-src 'self'` only)
- **Avatar upload magic-byte validation** — `internal/handler/user_handler.go` no longer trusts the client-supplied `Content-Type` header. The first 512 bytes of the uploaded file are read and passed to `http.DetectContentType` (WHATWG MIME Sniffing); the file pointer is then rewound for the subsequent copy. A filename-extension allowlist (`.jpg/.jpeg/.png/.gif/.webp`) closes the residual stored-XSS path where image magic bytes could be uploaded under a `.html` filename and later served with a `text/html` content-type by the static file handler
- New negative tests cover the spoofed-Content-Type and disallowed-extension cases

### CI
- **CI Failure Notify workflow now deduplicates and auto-closes.** Previously every CI failure (including each Dependabot rebase) opened a fresh `[CI] Workflow failure: ...` issue; old issues never closed. The workflow now (a) re-uses an existing tracking issue with the same `(workflow, branch)` title instead of spawning duplicates, and (b) on the next *successful* CI run for the same branch, closes the tracking issue with a comment linking the green run. Switched from `peter-evans/create-issue-from-file` to `actions/github-script` to access the issue list/update REST endpoints

### Maintenance
- Eight Dependabot bumps merged: actions/deploy-pages 4→5, docker/setup-buildx-action 3→4, lib/pq 1.12.1→1.12.3, x/crypto 0.49.0→0.50.0, pgx 5.9.1→5.9.2, vue-ecosystem group, go-sqlite3 1.14.38→1.14.42, dev-dependencies group (8 npm updates)
- Two Dependabot PRs closed with policy comments: Vuetify 4.0.5 (held by watch-and-wait policy until 4.1.x), axios 1.14.0 (post supply-chain compromise; close-and-wait until npm reissues clean 1.14.x or 1.15.x ships)

---

## [1.2.3] - 2026-04-03 — Security hardening release

### Security
- **Password policy** — raised minimum from 8 to 12 characters; now requires at least one uppercase letter, one lowercase letter, and one digit. Enforced in registration, password reset, and password change flows (`internal/service/user_service.go`). UI hints and client-side length checks updated in RegisterView, SettingsView, ResetPasswordView, and AdminUserImportExportView
- **CORS allowlist enforcement** — `pkg/middleware/cors.go` previously echoed the request's `Origin` back as `Access-Control-Allow-Origin` for both allowed and disallowed origins, making the `CORS_ORIGINS` allowlist inert. Disallowed origins now receive no `Access-Control-Allow-Origin` header (browser default-deny). Test in `cors_test.go` strengthened to assert header absence
- **Error sanitisation** — `WriteError()` no longer forwards raw internal error strings to HTTP responses; unknown errors return a generic message (`internal/handler/errors.go`)
- **XSS prevention** — DOMPurify 3.3.3 added to `MarkdownRenderer.vue`; all `v-html` output is sanitized before DOM insertion
- **serialize-javascript** — pinned to 7.0.5 via `package.json` overrides, fixing RCE (GHSA-5c6j-r48x-rmvq) and CPU exhaustion DoS (GHSA-qj8w-gfj5-8c6v) in the `vite-plugin-pwa → workbox-build → @rollup/plugin-terser` chain
- **Rate limiter IP extraction** — `X-Forwarded-For` header now correctly takes the leftmost (original client) IP from comma-separated lists; `RemoteAddr` port suffix stripped to prevent per-reconnect bucket bypass
- **Audit logging** — `rate_limit_exceeded` events now emitted via `OnExceeded` callback on both auth and password-reset rate limiters

### Added
- Admin organizations list shows a member count column (`LEFT JOIN user_organizations` in `OrganizationRepository.List`)
- Edit-from-list action in admin organizations view (reuses existing `PUT /api/admin/organizations/{id}`)

### Documentation
- `docs/OWASP_AUDIT_2026-04-03.md` — OWASP Top-10 audit findings
- `docs/MATURITY_ASSESSMENT.md` — Trail of Bits 9-category scorecard
- `docs/plans/SECURITY_HARDENING_PLAN.md` — phased remediation plan (Steps 2 and 5 carried into backlog)

### Maintenance
- golangci-lint upgraded from v1.64.8 to v2.11.4; config migrated to v2 format (`version: "2"`)
- Frontend unit tests added to CI pipeline (`web-test` job in `.github/workflows/ci.yml`); 11 new Vitest suites covering App shell, auth store, axios interceptor, and 8 admin views

### Known gaps (deferred to a future release — tracked in TODO.md)
- Security response headers middleware (CSP, HSTS, X-Frame-Options)
- Avatar upload still trusts client `Content-Type` header rather than detecting magic bytes

---

## [1.2.1] - 2026-04-01

### Added
- Calendar header now displays workout count for the viewed month

### Fixed
- Calendar view now fetches all workouts instead of defaulting to 20 (fixes missing entries for active users)
- Admin breadcrumb link returning 404
- Avatar upload click target too small; removed stale debug logging

### Changed
- Nav bar icons switched to heavier filled MDI variants for improved visual weight

### Security
- Pinned axios to `1.13.5` across frontend to prevent auto-upgrade into compromised versions 1.14.1 / 0.30.4 (supply chain attack via hijacked npm account, 2026-03-31)
- `npm audit fix`: patched brace-expansion (ReDoS), picomatch (method injection / ReDoS), undici (WebSocket memory/smuggling), yaml (stack overflow)
- Updated Go deps: `x/crypto` 0.48→0.49, `pgx/v5` 5.8→5.9, `go-sqlite3` 1.14.34→1.14.38, `lib/pq` 1.11→1.12

### Maintenance
- CI: `setup-go` now uses `go-version-file: go.mod` so toolchain tracks the module automatically
- CI: pinned `golangci-lint-action` to v6 (v9 dropped golangci-lint v1.x support)
- Fixed date-sensitive test `TestUserWorkoutRepository_GetActiveUsersThisMonth` (failed on first day of each month)
- Updated GitHub Actions: `docker/login-action` 3→4, `docker/metadata-action` 5→6, `docker/build-push-action` 6→7, `docker/setup-buildx-action` 3→4, `actions/checkout` 4→6, `github/super-linter` 4→7
- Updated frontend deps: Vue 3.5.28→3.5.31, vue-router 5.0.3→5.0.4, vitest 4.0.18→4.1.0, sass 1.97→1.98, marked 17.0.1→17.0.3

---

## [1.2.0-beta] - 2026-02-19

### Added - Configurable Beta Logo

- **Environment-Variable-Driven Logo Switching**
  - New `LOGO_VARIANT` environment variable (`"logo"` default, `"betalogo"` for red beta logo)
  - Backend: `LogoVariant` field in `AppConfig`, exposed via `/api/version` response
  - Frontend: Auth store fetches variant, computes `logoPath`, dynamically updates favicon and apple-touch-icon
  - All three logo locations updated (App.vue header, LoginView, RegisterView)
  - Full beta icon asset set generated from `betalogo.svg` (SVG, 12 PNGs, favicon.ico, 6 apple-touch-icons)
  - Example `.env` files updated with `LOGO_VARIANT` documentation

### Added - PR Leaderboards & Consistency Achievements

- **Leaderboard search API response parsing fix and header nav icon**

## [1.1.0-beta]

### Added - Coach Role & Role Rename

- **Three-Tier Role System**
  - Renamed `"user"` role to `"athlete"` across backend, frontend, and tests
  - Added `"coach"` as a new middle-tier role: athlete < coach < admin
  - Database migration (0.32.0) renames existing `user` records to `athlete`

- **CoachOrAdmin Middleware**
  - New `CoachOrAdmin` middleware in `pkg/middleware/auth.go` for coach-accessible routes
  - Coaches bypass subscription checks (same as admins) in `RequireActiveSubscription`
  - 14 comprehensive middleware tests covering all role permutations

- **Dedicated Coach API Routes**
  - `GET /api/coaches/me/sessions` - Coach's upcoming sessions (admins see all sessions across all gyms)
  - `GET /api/coaches/sessions/{id}/roster` - Session roster
  - `POST /api/coaches/sessions/{id}/check-in/{rid}` - Check in athlete
  - `POST /api/coaches/sessions/{id}/no-show/{rid}` - Mark no-show
  - `POST /api/coaches/sessions/{id}/complete` - Complete session
  - Organization-level access verification for coach actions

- **Frontend Updates**
  - Coach nav button in bottom navigation (visible for coach/admin when scheduling enabled)
  - Coach Dashboard updated to use `/api/coaches/` routes instead of admin routes
  - Admin Users panel shows three role options: Athlete, Coach, Admin
  - Router guard for `requiresCoach` meta on coach routes

- **Admin All-Sessions Visibility**
  - Admins see ALL upcoming sessions on Coach Dashboard (not limited by coach assignments)
  - New `GetAllUpcoming` repository method and `GetAllUpcomingSessions` service method

### Added - Delete Class with Sessions

- **Enhanced Class Template Deletion**
  - Three delete modes via `?mode=` query parameter:
    - `template_only` (default): Delete template only, sessions become orphaned
    - `with_future_sessions`: Delete template and future sessions, preserve past sessions
    - `with_all_sessions`: Delete template and all sessions (past and future)
  - Cascade deletion of related data (coaches, reservations, waitlist, notifications)
  - Credit refunds for unconfirmed reservations (status != checked_in/attended)
  - User notifications for cancelled sessions with session details

- **New UI Delete Dialog**
  - Radio button selection for delete mode
  - Session counts displayed for each option
  - Warning about cascading deletes
  - Information about notifications and credit refunds

- **API Enhancement**
  - `DELETE /api/admin/scheduling/templates/{id}?mode=<mode>`
  - Response includes: `sessions_deleted`, `notifications_sent`, `credits_refunded`

- **Repository Methods**
  - `GetSessionIDsByTemplateID`, `GetFutureSessionIDsByTemplateID`, `DeleteByIDs` on ClassSessionRepository
  - `DeleteBySessionIDs` on SessionCoachRepository, ReservationRepository, WaitlistRepository
  - `GetBySessionIDs` on ReservationRepository
  - `RefundCredit` on UserClassCreditsRepository

- **Tested on:** SQLite, MariaDB (192.168.1.234), PostgreSQL (192.168.1.143)

### Added - Class Scheduling System (Phase 4 Complete)

- **Documents Management**
  - Gym-specific document types (waivers, liability forms, health forms)
  - Track required vs optional documents per organization
  - Document expiration support with configurable validity periods
  - Admin CRUD for document types

- **User Documents Tracking**
  - Per-user document completion status (pending, completed, expired)
  - Expiration date tracking for time-limited documents
  - User view of pending/completed documents
  - Admin can mark documents as completed

- **Class Packages (Credit System)**
  - Flexible credit packages (e.g., "10-Class Pack", "Monthly Unlimited")
  - Configurable credits per package with validity periods
  - Price tracking for package purchases
  - Active/inactive package status

- **User Credits**
  - Credit balance tracking per user per organization
  - Credits automatically populated from package when purchasing
  - Expiration tracking for time-limited credit packages
  - Available credits calculation excluding expired records

- **Waitlist System**
  - Join waitlist when class is full
  - Automatic position tracking
  - Leave waitlist functionality
  - Waitlist status indicators (waiting, promoted, expired, cancelled)

- **Class Notifications**
  - Foundation for class reminders and waitlist promotions
  - Notification types: reminder, waitlist_promoted, class_cancelled, class_updated

- **Frontend Views**
  - `MyCreditsView.vue` - User's credits, documents, and waitlist entries
  - `AdminPackagesView.vue` - Admin management with 3 tabs (Packages, Documents, User Credits)
  - `ScheduleView.vue` - Added waitlist join/leave buttons and position indicators

- **API Endpoints (Phase 4)**
  - `GET/POST/PUT/DELETE /api/admin/gyms/{id}/documents` - Document type management
  - `GET/POST/PUT/DELETE /api/admin/gyms/{id}/packages` - Package management
  - `POST /api/admin/gyms/{id}/users/{id}/credits` - Add credits to user
  - `GET /api/gyms/{id}/users/me/credits` - Get user's credit balance
  - `GET /api/gyms/{id}/users/me/documents` - Get user's document status
  - `POST/DELETE /api/sessions/{id}/waitlist` - Join/leave waitlist
  - `GET /api/users/me/waitlist` - Get user's waitlist entries

- **Database Migration (0.27.0)**
  - `documents` - Document types per organization
  - `user_documents` - User document completion tracking
  - `class_packages` - Credit package definitions
  - `user_class_credits` - User credit balances
  - `waitlist_entries` - Session waitlist queue
  - `class_notifications` - Class-related notifications

### Fixed

- **PurchaseCredits Bug** - Now correctly uses package credits when `package_id` is provided without explicit `credits` value

### Added - Workout Template Instructions Field

- **Instructions Field for Template Movements**
  - New `instructions` column on `workout_template_movements` table
  - Supports markdown-formatted coaching cues, setup notes, and technique reminders
  - Optional field - movements work without instructions
  - Preserved through full CRUD lifecycle

- **Instructions Field for Template WODs**
  - New `instructions` column on `workout_template_wods` table
  - Allows WOD-specific scaling options, standards, and modifications
  - Consistent behavior with movement instructions

- **Frontend Updates**
  - Instructions textarea in `WorkoutTemplateEditView.vue` for both movements and WODs
  - Instructions display in `WorkoutTemplateDetailDialog.vue` with markdown rendering
  - Collapsible sections for cleaner UI when instructions are present

- **Backend Updates**
  - Domain models updated with Instructions field
  - Repository queries updated to include instructions in all operations
  - Handler validation for instructions field

- **Comprehensive Test Coverage**
  - Backend unit tests for service and handler layers
  - Frontend unit tests in `WorkoutTemplateEditView.test.js` (573 lines)
  - Integration tests in `workout_template_test.go` covering:
    - Full CRUD lifecycle for instructions
    - Special characters (markdown, unicode, HTML-like)
    - Multiple movements/WODs with different instructions
    - Optional field behavior (empty/null handling)

### Fixed

- **Repository Query Bug** - Fixed `GetByIDWithDetails` in `workout_repository.go` to include `instructions` column in SELECT queries for both movements and WODs

## [0.24.0-beta] - 2026-01-14

### Added - Data Quality & Duplicate Detection System

- **Admin Data Quality Dashboard**
  - New `AdminDataQualityView.vue` with 3 tabs (Overview, Duplicates, Data Issues)
  - Full database scan with summary cards for duplicates and issues
  - Duplicates by entity type breakdown with clickable View buttons
  - Data quality checks display with icons, descriptions, and counts
  - Zero-state display with success checkmarks when no issues found

- **Duplicate Detection Engine**
  - Scan 5 entity types: movements, WODs, user_workouts, users, workouts
  - Case-insensitive name matching for movements, WODs, workouts
  - Case-insensitive email matching for users
  - Composite key matching for user_workouts (user_id + date + name)
  - Shows duplicate group count and record count per entity type
  - FK reference counts displayed for informed merge decisions

- **Duplicate Merge Functionality**
  - Preview merge operation showing FK impact before execution
  - Safe merge with FK updates in database transaction
  - Automatic child record handling (movements/WODs from duplicates)
  - Audit logging of all merge operations (`duplicate_merge` event type)
  - Per-entity type merge with keep/delete ID selection

- **Data Quality Issue Detection**
  - 4 quality check types with severity levels:
    - Orphaned FK references (error) - FK pointing to deleted records
    - Empty required fields (error/warning) - Missing user names, movement names, WOD names
    - Future workout dates (warning) - Workouts scheduled in the future
    - Invalid email formats (warning) - Malformed email addresses
  - Filter chips for issue type filtering in Data Issues tab
  - Clickable cards to navigate from Overview to filtered issue lists

- **API Endpoints**
  - `GET /api/admin/data-quality/duplicates` - Scan all entities
  - `GET /api/admin/data-quality/duplicates/summary` - Quick summary
  - `GET /api/admin/data-quality/duplicates/{entity}` - Scan specific entity
  - `POST /api/admin/data-quality/duplicates/merge/preview` - Preview merge
  - `POST /api/admin/data-quality/duplicates/merge/confirm` - Execute merge
  - `GET /api/admin/data-quality/issues` - Scan for quality issues

- **Demo Mode Configuration**
  - Added DEMO MODE section to `.env.example`
  - Documented implemented vs planned demo features
  - Current workaround instructions for basic demo setup
  - Planned settings: DEMO_MODE, DEMO_USERS, DEMO_RESET_ON_STARTUP, etc.

### Added - Audit Logging Enhancements (2026-01-13)

- **Password Change Audit Logging**
  - New `password_changed` event logged when users change their password
  - Captures user email in event details
  - Works for both browser and API password changes

- **Password Reset Audit Logging**
  - New `password_reset` event logged when password is reset via token
  - Tracks which user completed the password reset

- **User Deletion Audit Logging Fix**
  - Fixed `user_deleted` event to properly log admin deletions
  - Stores target user info (id, email, name) in details JSON
  - Handles FK constraint by using nil target_user_id (user already deleted)

- **Duplicate Audit Log Bug Fix**
  - Fixed bug where SQLite/MySQL audit logs were being inserted twice
  - Repository Create() was calling both QueryRow and Exec for same INSERT
  - Now correctly uses Exec with LastInsertId for non-RETURNING databases

### Added - Admin Metrics Dashboard (2026-01-13)

- **System Metrics Dashboard**
  - New `GET /api/admin/metrics` endpoint returning all dashboard metrics
  - Real-time statistics for admin users
  - Single API call for efficient frontend consumption

- **User Statistics**
  - Total users count
  - Active users this month (users who logged workouts)
  - New users this month
  - Disabled users count

- **Workout Statistics**
  - Total workouts logged across all users
  - Workouts logged this month
  - Average workouts per user

- **Content Statistics**
  - Total movements (standard + user-created)
  - Total WODs (standard + user-created)
  - Total workout templates (standard + user-created)
  - Breakdown of user-created vs standard content

- **Subscription Statistics**
  - Active subscriptions count
  - Expiring soon (within 7 days)
  - Expired subscriptions count

- **System Health**
  - Recent audit events (24h)
  - Email success rate
  - Total emails sent
  - Failed emails count

- **Workout Trends**
  - 30-day workout activity chart
  - Daily workout counts with Chart.js visualization

- **Frontend Components**
  - `StatCard.vue` - Reusable stat card with icon, value, label, subtitle
  - `WorkoutTrendsChart.vue` - Bar chart for workout trends
  - `AdminMetricsDashboardView.vue` - Main dashboard view
  - Updated `AdminView.vue` with System Metrics card

- **Backend Infrastructure**
  - Cross-database date helpers for SQLite/PostgreSQL/MySQL
  - New repository methods: `CountNewThisMonth()`, `CountDisabled()`, `CountActive()`, `CountExpired()`, `CountExpiringSoon()`, `GetWorkoutTrends()`
  - `AdminMetricsService` aggregating all metrics
  - `AdminMetricsHandler` for HTTP endpoint

### Added - User Import/Export System (2026-01-13)

- **Bulk User Import from CSV**
  - New `POST /api/admin/user-management/import/preview` endpoint for CSV validation
  - New `POST /api/admin/user-management/import/confirm` endpoint to execute import
  - Two-phase import: preview with validation → confirm to create users
  - CSV format: `email,name,password` columns required
  - Duplicate email detection with skip option
  - Password validation (minimum 12 characters with complexity requirements)
  - All imported users get role `user` and permanent free subscription
  - Passwords hashed with bcrypt (cost 12)

- **User Export to CSV**
  - New `GET /api/admin/user-management/export` endpoint
  - Downloads CSV with all user emails and names
  - Passwords NOT included for security
  - Can be used as template for bulk imports

- **Batch Password Reset**
  - New `GET /api/admin/user-management/filter` endpoint for filtered user listing
  - New `POST /api/admin/user-management/batch-password-reset` endpoint
  - Filter users by name/email search and creation date range
  - Select multiple users and send password reset emails in bulk
  - Uses existing password reset token system (32-byte tokens, 1-hour expiry)
  - Maximum 100 users per batch for safety

- **Frontend Components**
  - `AdminUserImportExportView.vue` - Tabbed interface for import/export/reset
  - Drag-and-drop CSV upload zone
  - Import preview table with validation status per row
  - User selection table with filters and pagination
  - Updated `AdminView.vue` with User Import/Export navigation card
  - New route `/admin/users/import-export`

- **Backend Infrastructure**
  - `UserListFilter` type in domain layer for flexible querying
  - `ListWithFilter()` and `CountWithFilter()` repository methods
  - `UserImportService` with preview/confirm/export/filter/batch-reset methods
  - `UserImportHandler` with multipart form handling (max 10MB)
  - Automatic subscription creation for imported users

- **Documentation**
  - Screenshots for all tabs (import, export, password reset, preview)
  - Updated admin README with User Import/Export section
  - API endpoint documentation with examples

## [0.22.0-beta] - 2026-01-09

### Added - Benchmark API Endpoint

- **Comprehensive Benchmark Endpoint**
  - New `POST /api/benchmark` endpoint for API performance testing
  - Exercises database, serialization, and business logic operations
  - Uses isolated `benchmark_data` table (no production data affected)
  - Auto-cleanup of benchmark data after each run

- **Database Benchmarks** (9 operations)
  - Single insert, bulk insert (100 records)
  - Select by ID, select by key (index lookup)
  - List with pagination, filtered queries
  - Update and delete operations
  - User-scoped cleanup

- **Serialization Benchmarks** (4 operations)
  - JSON marshal/unmarshal for small payloads (1 object)
  - JSON marshal/unmarshal for large payloads (100 objects)

- **Business Logic Benchmarks** (5 operations)
  - 1RM calculations using prmath package (1000 iterations)
  - Intensity calculations (1000 iterations)
  - Input validation (100 iterations)
  - String operations (1000 iterations)
  - Date operations with timezone handling (1000 iterations)

- **Concurrent Benchmarks** (3 operations, optional)
  - Parallel reads (10 goroutines)
  - Parallel writes (5 goroutines)
  - Mixed read/write operations (10 goroutines)
  - Enable with `?concurrent=true` query parameter

- **Configurable Record Count**
  - New `records` query parameter for `/api/benchmark` endpoint
  - Default: 1,000 records, Maximum: 500,000 records
  - Example: `POST /api/benchmark?records=10000`
  - Scales database operations to stress-test the system

- **Complex Benchmark Data**
  - Large text fields (5-10KB random text with paragraphs)
  - Nested JSON blobs (5-level deep structures with arrays)
  - UUID test keys for uniqueness
  - Random numeric values (floats, integers, booleans)
  - Realistic data to push serialization and storage

- **Supporting Endpoints**
  - `GET /api/benchmark/status` - Quick status check
  - `DELETE /api/admin/benchmark/data` - Admin cleanup of all benchmark data

### Added - PWA 1.2.0 Features

- **Improved Update Flow**
  - Changed from `autoUpdate` to `prompt` registerType
  - User-controlled updates via UpdatePrompt component
  - Better state management in PWA store

- **PWA Assets Generator**
  - New `pwa-assets.config.js` for automatic icon generation
  - Generate all icons from single source image
  - New npm script: `npm run generate-pwa-icons`

- **Enhanced Caching**
  - Added caching for local fonts (1 year expiration)
  - Added caching for uploaded images (7 days, StaleWhileRevalidate)
  - Extended glob patterns for font files (woff, ttf)

### Changed

- **Database Migration 0.22.0**
  - Added `benchmark_data` table for safe read/write testing
  - Indexes on `test_key` and `created_by` columns
  - Foreign key to users table with CASCADE delete

- **Server Timeout Defaults**
  - Updated default SERVER_READ_TIMEOUT from 15s to 30s
  - Updated default SERVER_WRITE_TIMEOUT from 15s to 60s
  - Supports long-running benchmark API requests without EOF errors

### Fixed

- **EOF Errors on Long-Running Benchmark Requests**
  - Fixed HTTP client receiving EOF when benchmark takes >15 seconds
  - Root cause: SERVER_WRITE_TIMEOUT was set too low for large record counts
  - Added explicit Content-Length header to avoid chunked encoding issues
  - For 100k+ records, recommend SERVER_WRITE_TIMEOUT=120s or higher

## [0.21.0-beta] - 2026-01-09

### Added - Frontend Testing Framework

- **Vitest Testing Setup**
  - Installed Vitest with Vue Test Utils, jsdom, and v8 coverage
  - Created `vitest.config.js` with Vue/Vuetify support
  - Added test setup file with jsdom mocks (ResizeObserver, matchMedia, IntersectionObserver)
  - New npm scripts: `test`, `test:run`, `test:coverage`

- **Example Tests**
  - `src/utils/timezone.test.js` - 20 tests for timezone utilities
  - `src/components/UserAvatar.test.js` - 11 tests for component rendering

### Fixed - CI/CD and Code Quality

- **Linting Issues Resolved**
  - Fixed gofmt formatting across all Go files
  - Updated `.golangci.yml` with relaxed thresholds for complex query builders
  - Added `font_family` column to base database schema for all drivers

### Updated - Dependencies

- **Go Dependencies**
  - `github.com/jackc/pgx/v5` 5.7.4 → 5.8.0
  - `github.com/mattn/go-sqlite3` 1.14.32 → 1.14.33

- **Frontend Dependencies**
  - `vuetify` 3.10.11 → 3.11.6
  - `eslint-plugin-vue` 9.33.0 → 10.6.2
  - `vite-plugin-pwa` 0.21.2 → 1.2.0
  - Updated dev-dependencies group (4 packages)

## [0.20.0-beta] - 2026-01-08

### Changed - UI Visual Refresh

- **Form Input Style Update**
  - Changed default form input variant from `outlined` to `solo` (filled background, no border)
  - Cleaner, more modern appearance for text fields, selects, textareas, autocomplete, and combobox
  - Updated 43 Vue files to remove explicit `variant="outlined"` attributes

- **Card Border Removal**
  - Removed global card borders for a cleaner look
  - Cards now rely on elevation/shadow and background contrast
  - Removed inline border styles from all components

### Added - Myst Grayscale Theme

- **New Theme: Myst**
  - Elegant grayscale theme with light-gray background
  - Dark text on light backgrounds for readability
  - Inverted buttons (dark buttons with white text)
  - Professional, colorless aesthetic
  - Added to theme selector with fog icon

## [0.19.0-beta] - 2026-01-07

### Added - User-Customizable Fonts

- **10 Font Options for User Preference**
  - System Default (device's native font stack)
  - Inter, Roboto, Lato, Fira Sans (modern UI fonts)
  - Lexend (optimized for reading fluency)
  - OpenDyslexic, Atkinson Hyperlegible (accessibility fonts)
  - Source Serif Pro (classic serif)
  - JetBrains Mono (developer monospace)

- **Self-Hosted Web Fonts**
  - All fonts bundled as woff2 files in `web/public/fonts/`
  - No external CDN dependencies
  - All fonts licensed under SIL OFL 1.1 or Apache 2.0

- **Backend Integration**
  - New `font_family` field in user_settings domain model
  - Database migration v0.21.0 adds font_family column
  - Syncs font preference across devices

- **Frontend Implementation**
  - New `font.js` Pinia store for font state management
  - CSS variable `--app-font-family` for dynamic switching
  - Font selector UI in Settings view
  - localStorage caching for instant load on return visits
  - Vuetify component overrides for consistent font application

### Added - Markdown Support for Descriptions

- **Description Fields Support Markdown Formatting**
  - Movements, WODs, and workout templates now render Markdown
  - Added "Format text using Markdown" hint to all description textareas
  - MovementDetailView updated to use MarkdownRenderer component

### Added - Admin Notification System for User Events

- **Email Notifications for Administrators**
  - Notify admins when users are created, modified, or deleted
  - Events: registration, profile updates, role changes, disable/enable/unlock/delete
  - Per-admin opt-out setting in user_settings (`admin_user_event_notifications`)
  - Email includes before/after values for changes
  - Actor (admin who performed action) excluded from notifications

- **In-App Notifications for Administrators**
  - Real-time in-app notifications for all admin user events
  - Notification types: `admin_user_created`, `admin_user_updated`, `admin_user_role_changed`, etc.
  - Links directly to user management in admin panel
  - Notifications sent to all admins except the actor

- **New Service: AdminNotificationService**
  - Centralized admin notification logic
  - Async email sending (non-blocking)
  - Graceful degradation when email disabled
  - Files: `internal/service/admin_notification_service.go`

- **Database Migration v0.20.0**
  - Added `admin_user_event_notifications` column to `user_settings` table

### Added - Expiring/Expired Subscription Management

- **Admin Dashboard Subscription Views**
  - List subscriptions expiring within N days (7, 14, 30, 60, 90)
  - List all expired (overdue) subscriptions
  - Separate views for user and organization subscriptions

- **New API Endpoints**
  - `GET /api/admin/subscriptions/users/expiring?days=N` - Expiring user subscriptions
  - `GET /api/admin/subscriptions/users/expired` - Expired user subscriptions
  - `GET /api/admin/subscriptions/organizations/expiring?days=N` - Expiring org subscriptions
  - `GET /api/admin/subscriptions/organizations/expired` - Expired org subscriptions

- **Repository Methods**
  - Added `ListExpiring(days int)` and `ListExpired()` to both subscription repositories
  - Multi-database support (SQLite, PostgreSQL, MySQL) with database-native date arithmetic

### Added - User Timezone Support

- **Per-User Timezone Settings**
  - Users can set their preferred timezone in settings
  - Dates displayed in user's local timezone throughout the application
  - Database migration v0.19.0 adds `timezone` column to `user_settings`

- **Frontend Integration**
  - Timezone selector in Settings view
  - Utility functions for timezone-aware date formatting
  - Pinia store for timezone state management

### Fixed - PostgreSQL Compatibility

- **GetPersonalRecords Query Fix**
  - Added aggregate functions (MAX) to all non-grouped SELECT columns
  - Fixes PostgreSQL strict GROUP BY enforcement error
  - Scan workout_date as string and parse to time.Time (MAX returns string in SQLite)
  - All databases (SQLite, PostgreSQL, MySQL) now work correctly

- **Base Schema Sync**
  - Added `admin_user_event_notifications` column to base schema for all database types
  - Ensures fresh database installations have all columns

### Fixed - Test Suite Improvements

- **Test Coverage for Subscription Features**
  - Repository tests for `ListExpiring` and `ListExpired` methods
  - Service tests for all 4 new subscription service methods
  - Handler tests with mock implementations

- **Pre-existing Test Fixes**
  - Fixed backup handler test case sensitivity ("Filename" → "filename")
  - Updated WOD service test for removed source validation constraint
  - Integration test now outputs response body on failure for debugging

### Added - Merge/Upsert for Database Restore and Import Duplicate Handling

- **Database Restore Modes**
  - Added three restore modes to `RestoreBackup` API:
    - `replace` - Delete all data, insert from backup (original behavior)
    - `merge` - Update existing records by natural key, insert new ones
    - `skip` - Only insert records that don't exist
  - Natural key matching for record identification:
    - Users matched by email
    - Movements matched by name
    - WODs matched by name
    - Organizations matched by name
    - User workouts matched by user_id + workout_date + workout_name
  - ID remapping for foreign key references during merge/skip restore
  - Returns detailed `RestoreResult` with statistics (records created, updated, skipped)

- **Import Duplicate Handling Enhancements**
  - **User Workout Import**: Added `updateDuplicates` parameter
    - Duplicate detection using user_id + workout_date + workout_name
    - Can now skip OR update existing workouts during import
  - **Wodify Import**: Added `skipDuplicates` and `updateDuplicates` parameters
    - `skipDuplicates=true`: Skip workouts that already exist for a date
    - `updateDuplicates=true`: Replace existing workout data for a date
    - Added `WorkoutsSkipped` to import result statistics

- **API Changes**
  - `POST /api/admin/backups/{filename}/restore` now accepts `mode` parameter:
    ```json
    {"confirm": true, "mode": "merge"}  // or "skip" or "replace"
    ```
  - `POST /api/import/user-workouts/confirm` now accepts `update_duplicates` form field
  - `POST /api/import/wodify/confirm` now accepts `skip_duplicates` and `update_duplicates` form fields

**Files Modified:**
- `internal/domain/backup.go` - Added RestoreMode type and RestoreResult struct
- `internal/domain/wodify_import.go` - Added WorkoutsSkipped field
- `internal/service/backup_service.go` - Implemented merge/upsert with ID remapping (646+ lines)
- `internal/service/import_service.go` - Added updateDuplicates to user workout import
- `internal/service/wodify_import_service.go` - Added skip/update duplicate handling
- `internal/handler/backup_handler.go` - Added mode parameter parsing
- `internal/handler/import_handler.go` - Added update_duplicates form field
- `internal/handler/wodify_import_handler.go` - Added duplicate handling form fields

### Fixed - PostgreSQL Schema Migration Support
- **Critical Fix: PostgreSQL Custom Schema Support**
  - Fixed `checkTableExists()` to use `CURRENT_SCHEMA()` instead of hardcoded 'public'
  - Fixed `checkColumnExists()` to filter by `CURRENT_SCHEMA()` to avoid false positives
  - **Impact**: Migrations now work correctly with custom PostgreSQL schemas (e.g., `DB_SCHEMA=actalog`)
  - **Root Cause**: Previous code hardcoded `table_schema='public'` which failed for non-public schemas
  - **Behavior Before Fix**:
    - Migration checks would fail to find existing tables in custom schemas
    - Would attempt to re-create tables, causing "table already exists" errors
    - Required manual database rebuild when using custom PostgreSQL schemas
  - **Behavior After Fix**:
    - Automatic schema detection using active `search_path`
    - Migrations work seamlessly across all database engines and schemas
    - No manual intervention needed for schema migrations
  - Files Modified: `internal/repository/database.go` (lines 137-200)
  - Build: #60

### Added - Comprehensive Audit Logging
- **Complete Audit Trail for All Data Operations**
  - Added audit logging to MovementService (Create, Update, UpdateAsAdmin, Delete)
  - Added audit logging to WODService (Create, Update, UpdateAsAdmin, Delete)
  - Added audit logging to WorkoutTemplateService (Create, Update, Delete)
  - Added audit logging to UserWorkoutService (LogWorkout, LogWorkoutWithPerformance, UpdateLoggedWorkout, DeleteLoggedWorkout)
  - Added audit logging to UserService (UpdateProfile)
  - Added audit logging to UserSettingsService (UpdateSettings)

- **Audit Log Event Types**
  - Movement events: `movement_created`, `movement_updated`, `movement_deleted`
  - WOD events: `wod_created`, `wod_updated`, `wod_deleted`
  - Workout template events: `workout_template_created`, `workout_template_updated`, `workout_template_deleted`
  - User workout events: `user_workout_logged`, `user_workout_updated`, `user_workout_deleted`
  - User management events: `profile_updated`, `user_settings_updated`

- **Change Tracking Features**
  - Before/after values stored for all update operations
  - JSON-encoded details with full context
  - User attribution (UserID for performer, TargetUserID for affected user)
  - Admin operation flags for administrative updates
  - Timestamp tracking for all operations

### Changed
- **Service Method Signatures**
  - All CRUD operations now accept `userID int64` and `userEmail string` parameters
  - MovementService.Create signature: `Create(movement *domain.Movement, userID int64, userEmail string) error`
  - WODService methods updated to include userEmail parameter
  - WorkoutTemplateService methods updated to include userEmail parameter
  - UserWorkoutService methods updated to include userEmail parameter
  - UserSettingsService.UpdateSettings signature: `UpdateSettings(userID int64, userEmail string, updates *domain.UserSettings) (*domain.UserSettings, error)`

- **Handler Implementations**
  - All handlers extract userEmail from JWT context using `middleware.GetUserEmail()`
  - Handlers pass both userID and userEmail to service methods
  - Updated: movement_handler.go, wod_handler.go, workout_template_handler.go, user_workout_handler.go, settings_handler.go

- **Service Initialization**
  - All service constructors updated to accept auditLogRepo parameter
  - Main.go updated with comprehensive audit log repository injection
  - MovementService, WODService, WorkoutTemplateService, UserWorkoutService, UserSettingsService all initialized with audit logging

### Technical Details
- **Build**: #40
- **Files Modified**:
  - `internal/domain/audit_log.go` - Added comprehensive event constants
  - `internal/service/movement_service.go` - Audit logging for all operations
  - `internal/service/wod_service.go` - Audit logging for all operations
  - `internal/service/workout_template_service.go` - Audit logging for all operations
  - `internal/service/user_workout_service.go` - Audit logging for all operations
  - `internal/service/user_service.go` - Added audit logging to UpdateProfile
  - `internal/service/user_settings_service.go` - Added audit logging to UpdateSettings
  - `internal/handler/movement_handler.go` - Extract and pass userEmail
  - `internal/handler/wod_handler.go` - Extract and pass userEmail
  - `internal/handler/workout_template_handler.go` - Updated interface and handlers
  - `internal/handler/user_workout_handler.go` - Extract and pass userEmail
  - `internal/handler/settings_handler.go` - Extract and pass userEmail
  - `cmd/actalog/main.go` - Updated service initialization with auditLogRepo

### Audit Logging Patterns
- **Conditional Logging**: All audit logging uses nil checks (`if s.auditLogRepo != nil`)
- **JSON Encoding**: Audit details stored as JSON for structured querying
- **Error Handling**: Audit log failures don't block primary operations (fire-and-forget pattern)
- **Entity Context**: Each log includes entity type, entity ID, entity name, and user context
- **Admin Tracking**: Admin operations marked with `admin_update: true` flag

### Coverage
- ✅ Movement CRUD operations fully logged
- ✅ WOD CRUD operations fully logged
- ✅ Workout Template CRUD operations fully logged
- ✅ User Workout operations fully logged
- ✅ User Profile updates fully logged
- ✅ User Settings updates fully logged
- ✅ Organization operations logged (from v0.14.0)
- ✅ Subscription operations logged (from v0.14.0)
- ✅ User associations logged (from v0.14.0)

### Added - Comprehensive Test Coverage Improvements

- **Service Layer Test Coverage: 81.6% Overall**
  - 13 services now at 100% coverage
  - Added comprehensive backup_service tests (10 test functions)
  - Added wodify_import_service tests with duplicate handling
  - Fixed mock initialization patterns across test files
  - Added error injection support to mockUserRepo

- **Backup Service Tests** (`backup_service_test.go`)
  - `TestBackupService_CreateBackup` - Full backup creation with ZIP archive
  - `TestBackupService_CreateBackup_UserNotFound` - Error handling for missing user
  - `TestBackupService_CreateBackup_WithUploads` - Backup with file attachments
  - `TestBackupService_RestoreBackup` - Replace mode restoration
  - `TestBackupService_RestoreBackup_MergeMode` - Merge mode with ID remapping
  - `TestBackupService_RestoreBackup_SkipMode` - Skip mode preserving existing
  - `TestBackupService_CreateSQLiteDump` - SQLite dump generation
  - `TestBackupService_ExportAllTables` - JSON export of all tables
  - Added `setupFullTestDB` helper with complete 20-table schema

- **Test Infrastructure Improvements** (`test_helpers.go`)
  - Added `getByIDError` field to mockUserRepo for error injection
  - Fixed mock constructor patterns for proper map initialization
  - All mocks now use `newMockX()` constructor functions

- **Files Modified:**
  - `internal/service/test_helpers.go` - Enhanced mockUserRepo with error injection
  - `internal/service/backup_service_test.go` - 10 new test functions
  - `internal/service/wodify_import_service_test.go` - Fixed mock initialization
  - `docs/TESTING.md` - Complete rewrite with current coverage stats

---

## [0.16.0-beta] - 2024-12-20

### Added - Notification Likes Feature

**Social Engagement for Notifications:**
- Users can "like" any notification (PR achievements, announcements, weekly streaks, WOD milestones)
- Like count displayed with thumbs up icon next to each notification
- Names of users who liked displayed as comma-separated list ("Liked by: jcz, John, Mary")
- Users can like their own notifications
- **Automatic unread marking**: When someone likes a notification, it marks as unread for the original recipient
- CASCADE DELETE: Deleting a notification automatically removes all associated likes

**Backend Implementation:**
- New `notification_likes` table with proper indexes and foreign keys
- Unique constraint prevents duplicate likes (409 Conflict error returned)
- Repository layer with JOIN query to fetch likes with user details
- Service layer handles like/unlike operations and automatic unread marking
- New `MarkAsUnread` method added to NotificationRepository interface

**API Endpoints:**
- `POST /api/notifications/{id}/like` - Like a notification
- `DELETE /api/notifications/{id}/like` - Unlike a notification
- `GET /api/notifications/{id}/likes` - Get all likes with user details (returns count and array of likes)

**Frontend Component:**
- `NotificationLikes.vue` component with thumbs up icon
- Toggle between filled (liked) and outline (not liked) states
- Real-time like count updates
- Displays "Liked by: name1, name2, name3" in small grey text
- Integrates seamlessly into existing NotificationsView
- Proper authentication using configured axios instance

**Database Migration:**
- Migration v0.16.0 creates `notification_likes` table
- Multi-database support: SQLite, PostgreSQL, MySQL
- Indexes on `notification_id` and `user_id` for performance
- Unique constraint on `(notification_id, user_id)` pair

**Technical Details:**
- **Build**: #59
- **Version**: 0.16.0-beta
- **Files Created**:
  - `internal/domain/notification_like.go` - Domain model and repository interface
  - `internal/repository/notification_like_repository.go` - Data access layer
  - `internal/service/notification_like_service.go` - Business logic layer
  - `internal/handler/notification_like_handler.go` - HTTP handlers
  - `web/src/components/NotificationLikes.vue` - Frontend component
- **Files Modified**:
  - `internal/repository/migrations.go` - Added v0.16.0 migration
  - `internal/domain/notification.go` - Added `MarkAsUnread` method to interface
  - `internal/repository/notification_repository.go` - Implemented `MarkAsUnread`
  - `cmd/actalog/main.go` - Wired up dependencies and routes
  - `web/src/views/NotificationsView.vue` - Integrated NotificationLikes component
  - `pkg/version/version.go` - Updated to v0.16.0-beta

**Testing:**
- ✅ Like/unlike operations verified
- ✅ Like count updates correctly
- ✅ User details populated via JOIN query
- ✅ Notifications marked as unread when liked
- ✅ CASCADE DELETE verified
- ✅ Duplicate like prevention (409 Conflict)

### Fixed - Profile View Workout Summary Statistics

**Bug Fixes:**
- Fixed Personal Records count showing 0 (was looking for `personal_records` field, API returns `movements`)
- Now correctly sums `pr_count` from each movement for total PR count
- Fixed Current Streak calculation with incorrect consecutive day logic
- Rewrote streak algorithm to properly check for consecutive workout days
- Added null safety for API responses when no PRs exist

**Streak Calculation Improvements:**
- Uses unique workout dates (handles multiple workouts per day)
- Checks if most recent workout is today or yesterday (or streak is broken)
- Counts backwards through consecutive days correctly
- Stops at first gap in workout dates

**Files Modified:**
- `web/src/views/ProfileView.vue:604-673` - Stats fetching and calculation logic

### Added - Workout Summary Time Period Filters

**Time Filter Options:**
- This Week (Sunday through today)
- This Month (1st of month through today)
- This Year (January 1st through today)
- All Time (no filtering)

**Filtering Behavior:**
- Total Workouts: Filtered by selected period
- Personal Records: Filtered by `last_pr_date` from PR movements
- Current Streak: Always all-time (continuous metric)
- Custom Templates: Always all-time (cumulative creations)

**UX Features:**
- Chip-based filter selector above stats
- Real-time updates when switching periods
- Default period: "This Month"
- Reactive updates using Vue watch

**Files Modified:**
- `web/src/views/ProfileView.vue:198-212, 578, 603-709, 887-890` - Time filter UI and logic

### Fixed - CI Pipeline Compilation Errors

**Test Compilation Fixes:**
- Fixed `mockEmailService` missing `SendHTMLEmail(to, subject, htmlBody)` method
- Fixed `NewUserWorkoutService` calls missing 4 parameters in all 5 test cases
- Added `mockMovementRepo` with 13 methods implementing `MovementRepository` interface
- Added `GetAllPerformancesForMovement` to `mockUserWorkoutMovementRepo`

**Impact:**
- ✅ All service tests now compile successfully
- ✅ Integration tests compile successfully
- ✅ CI pipeline unblocked for Go compilation step
- ✅ Database matrix tests (SQLite, PostgreSQL, MySQL) can now run

**Files Modified:**
- `internal/service/user_service_test.go` - Added SendHTMLEmail mock method
- `internal/service/user_workout_service_test.go` - Fixed all NewUserWorkoutService calls
- `internal/service/test_helpers.go` - Added mockMovementRepo and GetAllPerformancesForMovement

---

## [0.14.0-beta] - 2024-12-16

### Added - Subscription Billing System

**Dual-Level Subscription Management:**
- User-level subscriptions (individual billing)
- Organization-level subscriptions (gym/team billing)
- Flexible access model: users have access if EITHER personal OR any organization subscription is active
- Three subscription types: Free, Monthly, Annual
- Permanent Free subscriptions for founders/staff (never expire, never need payment)
- Immediate read-only mode enforcement when subscriptions expire (no grace period)

**Admin Controls:**
- Manual payment management (admins mark subscriptions as paid/unpaid)
- Subscription creation, cancellation, and history viewing
- Complete audit trail for all subscription operations
- Admin API endpoints for subscription management

**Read-Only Mode:**
- Users with expired subscriptions can view all data, export data, access dashboard/analytics
- Write operations blocked (POST/PUT/PATCH/DELETE) return HTTP 402 Payment Required
- Read operations allowed (GET/HEAD/OPTIONS) for viewing and exporting

**Backward Compatibility:**
- Migration 0.14.0 automatically seeds all existing users with permanent free subscriptions
- Zero downtime deployment - no existing users lose access
- All current users receive `is_permanent_free = TRUE` status

**Database Support:**
- Multi-database migration tested on SQLite, PostgreSQL, and MariaDB
- Version snapshots created for all three database engines
- Database versioning system for testing future migrations

### Technical Implementation

**New Domain Entities:**
- `UserSubscription` - Individual user subscriptions with 15 fields
- `OrganizationSubscription` - Organization-level subscriptions
- `SubscriptionAccessResult` - Access check result with source indication

**Repository Layer:**
- `UserSubscriptionRepository` - CRUD operations for user subscriptions
- `OrganizationSubscriptionRepository` - CRUD for organization subscriptions
- `SubscriptionAccessRepository` - Performance-optimized access checking (< 10ms target)

**Service Layer:**
- `SubscriptionService` - Business logic for subscription management
- Admin operations: Create, MarkAsPaid, Cancel
- User operations: CheckAccess, GetStatus

**Middleware:**
- `RequireActiveSubscription` - HTTP method-based enforcement
- Allows GET requests when expired (view/export)
- Blocks POST/PUT/PATCH/DELETE when expired

**API Endpoints:**
- `GET /api/subscriptions/status` - User subscription status
- `POST /api/admin/subscriptions/user` - Create user subscription
- `POST /api/admin/subscriptions/user/{id}/mark-paid` - Mark as paid
- `POST /api/admin/subscriptions/user/{id}/cancel` - Cancel subscription
- `GET /api/admin/subscriptions/user/{user_id}` - View subscription history
- Organization subscription endpoints (parallel structure)

**Database Version Management:**
- `db_versions/actalog_0.14.0.db` - SQLite snapshot (564 KB with production-like data)
- PostgreSQL schema `actalog_0_14_0` on 192.168.1.143
- MariaDB database `actalog_0_14_0` on 192.168.1.234
- All version databases contain identical test data with 4 test users
- Automated scripts: `create-db-snapshot.sh`, `verify-version-databases.sh`

### Files Created (14 files)
- `internal/domain/subscription.go` (106 lines) - Domain entities and interfaces
- `internal/repository/user_subscription_repository.go` (441 lines)
- `internal/repository/organization_subscription_repository.go` (441 lines)
- `internal/repository/subscription_access_repository.go` (145 lines)
- `internal/service/subscription_service.go` (368 lines)
- `pkg/middleware/subscription.go` (81 lines)
- `internal/handler/subscription_handler.go` (476 lines)
- `db_versions/README.md` - Version management overview
- `db_versions/VERSION_DATABASES.md` - Multi-database access guide (284 lines)
- `db_versions/MIGRATION_TEST_0.14.0.md` - Migration test report
- `db_versions/actalog_0.14.0.db` - SQLite version snapshot
- `scripts/create-db-snapshot.sh` - Automated snapshot creation
- `scripts/verify-version-databases.sh` - Multi-database verification
- PostgreSQL schema: `actalog_0_14_0` with test data
- MariaDB database: `actalog_0_14_0` with test data

### Files Modified (5 files)
- `internal/repository/migrations.go` - Added migration 0.14.0 (lines 885-1162)
- `pkg/version/version.go` - Updated to 0.14.0, build 24
- `cmd/actalog/main.go` - Wired repositories, service, handler, routes
- `internal/domain/audit_log.go` - Added subscription event types
- `CLAUDE.md` - Added database version management section

### Technical Details
- **Build**: #24
- **Version**: 0.12.2-beta → 0.14.0-beta
- **Migration**: 0.14.0 creates `user_subscriptions` and `organization_subscriptions` tables
- **Schema**: 15 columns per table, 4 indexes per table for performance
- **Constraints**: CHECK constraints on subscription_type and status, CASCADE/SET NULL foreign keys
- **Performance**: Access check optimized for < 10ms per authenticated request
- **Audit**: All subscription operations logged with admin user ID and details

---

## [0.12.2-beta] - 2025-11-28

### Fixed - PWA Offline Functionality

**Offline Workout Recording:**
- Fixed service worker API caching pattern (was `/api/.*/*.json`, now correctly caches `/api/workouts`, `/api/movements`, `/api/wods`, `/api/templates`)
- Added robust offline detection in axios interceptor (checks multiple error indicators: `Network Error`, `ERR_NETWORK`, `navigator.onLine`, timeout)
- Extended offline handling to support PUT requests (workout updates) in addition to POST
- Fixed request data parsing for offline sync queue

**User-Controlled PWA Updates:**
- Replaced disruptive silent auto-reload with user-controlled updates
- New `UpdatePrompt.vue` component shows "Update Available" snackbar with "Later" and "Update Now" buttons
- New `pwa.js` Pinia store manages PWA update state
- Users can choose when to apply updates, preventing data loss during form entry

**Offline Save Notification:**
- Added "Saved Offline" snackbar notification when workouts are saved locally
- Dispatches custom `offline-save` event for UI notification
- Automatically increments pending sync count in network store

### Fixed - Unit Tests

**WOD Service Tests:**
- Fixed `mockWODRepo.GetByName()` to return `nil, nil` instead of `sql.ErrNoRows` when WOD not found
- Updated test cases to use correct error types (`ErrWODOwnership`, `ErrWODUnauthorized`)
- Added required fields (Source, Type, Regime, ScoreType) to Update tests
- Fixed Search test expectation for empty query (service returns empty by design)

### Technical Details
- **Build**: #1 (reset for new patch version)
- **Version**: 0.12.1-beta → 0.12.2-beta
- **Files Created**:
  - `web/src/components/UpdatePrompt.vue` (PWA update notification component)
  - `web/src/stores/pwa.js` (PWA state management)
- **Files Modified**:
  - `web/vite.config.js` (service worker caching patterns)
  - `web/src/utils/axios.js` (offline detection and sync)
  - `web/src/App.vue` (offline save notification, update prompt)
  - `web/src/main.js` (PWA store integration)
  - `internal/service/test_helpers.go` (mock fixes)
  - `internal/service/wod_service_test.go` (test corrections)

---

## [0.12.1-beta] - 2025-11-28

### Fixed - MySQL/MariaDB Compatibility

**Database-Agnostic Timestamp Functions:**
- Fixed `addWorkoutMovementWithDistance()` to use database-specific timestamp syntax
- Fixed `refresh_token_repository.go` functions with hardcoded SQLite `datetime('now')`:
  - `GetByToken()` - Token expiration check
  - `Revoke()` - Token revocation timestamp
  - `RevokeAllForUser()` - Bulk revocation timestamp
  - `DeleteExpired()` - Expired token cleanup
- Added `getTimestampFunc()` helper for database-agnostic timestamp generation
- Supports SQLite (`datetime('now')`), PostgreSQL (`CURRENT_TIMESTAMP`), MySQL/MariaDB (`NOW()`)

### Added - Docker Host Database Documentation

**Comprehensive Troubleshooting Guide for External Databases:**
- `docker/DOCKER.md` - New "Connecting to Host Database" troubleshooting section
- `docker/DATABASE_DEPLOYMENT.md` - Enhanced external database troubleshooting:
  - UFW firewall configuration for Docker network access
  - MariaDB/MySQL bind-address configuration steps
  - PostgreSQL listen_addresses and pg_hba.conf setup
  - Database user permission grants for Docker network (172.17.0.0/16)
  - Linux `extra_hosts: host.docker.internal:host-gateway` configuration
  - Step-by-step connection refused debugging

### Technical Details
- **Build**: #4 → #5
- **Version**: 0.12.0-beta → 0.12.1-beta
- **Files Modified**:
  - `internal/repository/database.go` (timestamp helper + fix)
  - `internal/repository/refresh_token_repository.go` (4 function fixes)
  - `docker/DOCKER.md` (host database troubleshooting)
  - `docker/DATABASE_DEPLOYMENT.md` (firewall + connection troubleshooting)

---

## [0.12.0-beta] - 2025-11-26

### Fixed - Mobile PWA Layout

**Comprehensive Mobile Overflow Fix:**
- Systematically fixed horizontal/vertical overflow issues across 27 view files
- Implemented `.mobile-view-wrapper` CSS pattern for consistent mobile-safe layouts
- Updated `main.css` with global mobile-safe styles preventing body overflow
- Enhanced `App.vue` with dynamic safe-area handling for iOS PWA
- Views updated include:
  - Dashboard, Profile, Performance, Workouts
  - WOD/Movement Libraries and Detail views
  - Log Workout, Quick Log dialogs
  - Admin views (Users, Backups, Data Cleanup, Data Change Logs, User Content)
  - Import/Export, PR History, Templates
  - Auth flows (Login, Register, Forgot/Reset Password)
- Prevents content from extending beyond viewport on mobile devices
- Ensures smooth scrolling within content areas
- Maintains fixed header (56px) and bottom navigation (70px) positioning

### Added - Docker Image Metadata

**OCI Labels for Docker Images:**
- Added comprehensive OCI-compliant labels to Docker build scripts
- User-editable metadata section at top of `docker/scripts/build.sh`
- Labels include:
  - `org.opencontainers.image.title` - Application name
  - `org.opencontainers.image.description` - Project description
  - `org.opencontainers.image.vendor` - Organization/vendor name
  - `org.opencontainers.image.authors` - Author contact information
  - `org.opencontainers.image.source` - Repository URL
  - `org.opencontainers.image.documentation` - Documentation URL
  - `org.opencontainers.image.licenses` - License type
  - `org.opencontainers.image.version` - Build version (from tag)
  - `org.opencontainers.image.created` - Build timestamp (auto-generated)
- Improves container registry display and image discoverability

### Changed - Admin UI Improvements

**Admin User Content View Enhancement:**
- Moved Actions column from last to first position in data table
- Improves mobile usability by placing action buttons immediately visible
- Consistent with other admin table patterns
- Delete buttons now accessible without horizontal scrolling

### Technical Details
- **Build**: #71 → #1 (reset for new minor version)
- **Version**: 0.11.0-beta → 0.12.0-beta
- **Files Modified**: 27 Vue view files, App.vue, main.css, build.sh, AdminUserContentView.vue

---

## [0.11.0-beta] - 2025-11-26

### Added - Data Change Audit Logs

**Data Change Logging System:**
- Complete audit trail for data modifications (updates and deletes)
- Before/after values stored as JSON for full change history
- Tracks entity type, entity ID, entity name, operation, user, timestamp
- Optional IP address and user agent capture
- Automatic integration with WOD and Movement services

**Admin UI for Data Change Logs:**
- New admin view at `/admin/data-change-logs` for browsing change history
- Filterable by entity type (WOD, Movement, Workout, etc.)
- Filterable by operation (Update, Delete)
- Filterable by user email (partial match)
- Paginated data table with 50 logs per page
- Details dialog showing:
  - Full change metadata (timestamp, entity info, user)
  - Before/After JSON values in formatted display
  - Changed fields diff table for updates
- Color-coded operations (warning for update, error for delete)
- Color-coded entity types for easy identification

**Navigation Integration:**
- "Data Change Logs" card added to Admin dashboard
- Link added to Profile page Administration section for admins
- Consistent styling with existing admin tools

**Database Migration:**
- New `data_change_logs` table (migration 0.5.2)
- Indexes on entity_type, entity_id, user_id, created_at, operation
- Support for all database drivers (SQLite, PostgreSQL, MySQL)

**API Endpoints:**
- `GET /api/admin/data-change-logs` - List with filters and pagination
- `GET /api/admin/data-change-logs/:id` - Get single log entry
- `GET /api/admin/data-change-logs/entity/:type/:id` - Entity history
- `POST /api/admin/data-change-logs/cleanup` - Delete old logs

### Technical Details
- **Build**: #62 → #63
- **New Files**:
  - `internal/domain/data_change_log.go` - Domain model and interfaces
  - `internal/repository/data_change_log_repository.go` - Data access layer
  - `internal/service/data_change_log_service.go` - Business logic with helper methods
  - `internal/handler/data_change_log_handler.go` - HTTP handlers
  - `web/src/views/AdminDataChangeLogsView.vue` - Admin UI component
- **Modified Files**:
  - `internal/service/wod_service.go` - Added data change logging
  - `internal/service/movement_service.go` - Created service with data change logging
  - `internal/handler/wod_handler.go` - Pass user email to service
  - `internal/handler/movement_handler.go` - Use new MovementService
  - `web/src/views/AdminView.vue` - Added Data Change Logs card
  - `web/src/views/ProfileView.vue` - Added admin link to Data Change Logs
  - `web/src/router/index.js` - Added route for data-change-logs

### Use Cases Enabled
- Track who changed what and when across all data types
- Review before/after values for any modification
- Audit trail for compliance and accountability
- Debug data issues by reviewing change history
- Admin cleanup of old logs to manage storage

---

## [0.10.0-beta] - 2025-01-23

### Added - Docker Deployment with Automatic Seed Import

**Docker Infrastructure:**
- Multi-stage Dockerfile with optimized build process
- Three docker-compose configurations:
  - `docker-compose.yml` - SQLite (default, single-server deployments)
  - `docker-compose.postgres.yml` - PostgreSQL (production recommended)
  - `docker-compose.mariadb.yml` - MariaDB/MySQL (production alternative)
- GitHub Actions CI/CD workflow for automated image building
- Helper scripts for building and pushing Docker images
- Health checks for container monitoring

**Automatic Seed Data Import:**
- Optional automatic import of CSV seed data on first deployment
- Environment-based configuration (ADMIN_EMAIL, ADMIN_PASSWORD)
- Entrypoint script orchestrating app startup and seed import
- Imports 182 movements and 314 WODs automatically
- One-time execution using marker file pattern
- Graceful degradation when credentials not provided

**Comprehensive Documentation:**
- `DOCKER.md` - Complete Docker deployment guide with examples
- `DATABASE_DEPLOYMENT.md` - Multi-database deployment guide
- `TEST.md` - Testing guide for Docker deployments
- Environment configuration templates for all databases
- Migration guides between database types

**Seed Data:**
- 182 CrossFit movements (all standard movements including Girl/Hero WOD movements)
- 314 benchmark WODs (all Girl and Hero WODs)
- CSV format for easy import and modification

### Technical Details
- **Build**: #62 → #63 (build number auto-incremented)
- **New Files**:
  - `docker/Dockerfile` - Multi-stage build (frontend, backend, runtime)
  - `docker/scripts/entrypoint.sh` - Startup orchestration
  - `docker/scripts/init-seeds.sh` - Seed import script
  - `docker/scripts/build.sh` - Docker build helper
  - `docker/scripts/push.sh` - GitHub Container Registry push helper
  - `.github/workflows/docker-build.yml` - CI/CD automation
- **Modified**: All environment template files (.env.example, .env.postgres, .env.mariadb)
- **Documentation**: Added comprehensive Docker and database deployment guides

### Deployment Features
- GitHub Container Registry (ghcr.io) integration
- Automatic image builds on push to main branch
- Tag-based versioning (latest, version-specific tags)
- Health check endpoints for monitoring
- Volume management for persistent data
- Network isolation with bridge networks
- Non-root container user for security

## [0.9.0-beta] - 2025-01-23

### Added - Full Offline Support & PWA Enhancements

**iOS PWA Support:**
- iOS-specific meta tags for full PWA capabilities
- Apple touch icon configuration
- Black-translucent status bar styling

**Network Status Management:**
- Pinia network store for centralized online/offline state
- Real-time status chip in app bar (Offline/Syncing indicators)
- Automatic network event detection
- Pending sync operation counter

**User Notifications:**
- Persistent offline notification with explanation
- 3-second online notification when reconnected
- Sync complete confirmation notification
- All notifications dismissible by user

**Offline Data Storage:**
- IndexedDB integration with axios interceptors
- Automatic request queuing for failed network calls
- Offline workout creation with background sync
- Movement list caching for offline access

**Auto-Sync Mechanism:**
- Automatic sync when connection restored
- Visual sync status feedback
- Error handling with retry logic
- Manual and automatic sync triggers

**Offline-Capable Data Fetching:**
- `useOfflineData` composable for network-aware loading
- Cache-first strategy with API fallback
- Generic `fetchWithCache` pattern
- Movement caching implementation

**PWA Install Prompt:**
- Custom branded install UI
- Smart timing (1 minute delay)
- 7-day dismissal memory
- Installation state detection

### Changed
- Enhanced axios interceptors for offline request handling
- Updated service worker runtime caching configuration
- App.vue now includes network notifications and install prompt

### Technical Details
- **Build**: #61
- **New Files**: network.js store, InstallPrompt.vue, useOfflineData.js composable
- **Modified**: index.html, App.vue, axios.js, offlineStorage.js

## [0.8.2-beta] - 2025-01-23

### Fixed
- **Quick Log Template Selection**: Fixed crash when selecting workout templates from Quick Log dialog
  - Removed conflicting `item-value` property from v-autocomplete (was causing null object errors)
  - Added optional chaining (`?.`) throughout `submitQuickLog()` for defensive coding
  - Added error alert when template data is invalid
- **Template WOD Display**: Fixed WOD names not showing when logging from template
  - Updated `getWODName()` to handle both nested (`wod.name`) and flattened (`wod_name`) API formats
  - Updated `initializePerformanceArrays()` to handle both WOD data formats
  - Fixed score type mapping to use full format (`'Time (HH:MM:SS)'` instead of `'Time'`)
  - Added missing `time_hours` field to WOD performance initialization

### Enhanced
- **UI Consistency**: Updated Log Workout page styling to match Quick Log aesthetic
  - Removed excessive rounded corners from form elements (changed `rounded` to `border-radius: 8px`)
  - Made card styles more compact and consistent
- **Quick Log UX**: Improved template selection workflow
  - Hidden "Browse Templates" button when arriving from Quick Log with pre-selected template
  - Added prominent warning message explaining data preservation behavior
  - Changed template info box to orange warning theme with information icon
  - Clear message: "Only the date will be preserved. Notes, workout name, and total time entered here will not be carried over."

### Technical
- Enhanced `onTemplateSelected()` function to properly initialize performance arrays after loading template
- Added template to list if not already present (handles Quick Log navigation scenario)
- Improved data format compatibility between different API response structures

## [0.8.1-beta] - 2025-01-22

### Added
- **Cross-Database Backup/Restore**: Complete database-agnostic backup and restore system
  - Database-agnostic table existence checks using `information_schema` (PostgreSQL/MySQL) and `sqlite_master` (SQLite)
  - Table column introspection for schema evolution support
  - Automatic detection of schema differences between backup and target database
  - Column filtering: Only restores columns that exist in target schema (handles removed columns gracefully)
  - New columns use DEFAULT values from schema (handles added columns)
  - **Full cross-database migration support**:
    - ✅ MariaDB → PostgreSQL
    - ✅ SQLite → PostgreSQL
    - ✅ MySQL → MariaDB
    - ✅ Any combination of supported databases

- **PostgreSQL Sequence Reset**: Automatic sequence management after restore
  - Resets auto-increment sequences to `MAX(id) + 1` for all tables
  - Prevents "duplicate key violation" errors on subsequent inserts
  - Uses `pg_get_serial_sequence()` and `setval()` for proper sequence handling
  - Only applies to PostgreSQL (SQLite/MySQL handle sequences differently)

- **Data Type Conversion**: Automatic type conversion between databases
  - Boolean conversion: `0/1` (SQLite/MySQL) ↔ `false/true` (PostgreSQL)
  - Handles columns: `is_pr`, `is_template`, `is_standard`, `email_verified`, `account_disabled`, `notifications_enabled`
  - JSON unmarshaling safety: Handles `float64` → boolean conversion
  - Preserves data integrity across different database type systems

- **Schema Evolution Support**: Forward and backward compatibility for version migrations
  - Backup from v0.6.0 can be restored to v0.8.1 (handles missing tables/columns)
  - Backup from v0.8.1 can be restored to v0.6.0 (newer columns gracefully ignored)
  - Informative logging: "skipped N column(s) not present in target schema"
  - No manual SQL intervention required for schema differences

### Enhanced
- **RestoreBackup Function**: Complete rewrite for database compatibility
  - Replaced SQLite-specific `sqlite_master` queries with database-agnostic `tableExists()`
  - Added column introspection before each table restore
  - Integrated automatic sequence reset for PostgreSQL
  - Enhanced error messages with specific table and column information
  - Graceful handling of missing tables (forward compatibility)

- **restoreTable Function**: Full schema evolution and type conversion support
  - Column filtering based on actual target schema
  - Value conversion for database compatibility
  - Per-column type conversion using `convertValue()`
  - Informative progress logging during restore
  - Automatic sequence reset after table population

### Functions Added
- `tableExists(tx, tableName)`: Database-agnostic table existence check
- `getTableColumns(tx, tableName)`: Query actual schema columns
- `resetSequence(tx, tableName)`: PostgreSQL sequence management
- `convertValue(val, columnName)`: Cross-database type conversion
- `containsString(slice, str)`: Helper for column filtering

### Use Cases Enabled
- **Production Database Migration**: Migrate from SQLite (development) to PostgreSQL (production) using backup/restore
- **Cross-Database Replication**: Copy data between different database systems without manual export/import
- **Version Upgrades**: Restore old backups to newer application versions seamlessly
- **Multi-Tenant Migration**: Migrate from single-tenant to multi-tenant PostgreSQL using schema parameter
- **Disaster Recovery**: Restore backups to different database types in emergency scenarios
- **Development → Production**: Test with SQLite, deploy with PostgreSQL using same backup files

### Technical Details
- **Build Number**: #58
- **Files Modified**:
  - `internal/service/backup_service.go`: Added 190+ lines of new database-agnostic helper functions
  - Updated `RestoreBackup()` and `restoreTable()` with full schema evolution support
- **Backward Compatibility**: 100% backward compatible - same-database restores work identically
- **Testing**: Builds successfully, ready for cross-database testing

### Migration Example
```bash
# On MariaDB v0.7.x instance
POST /api/admin/backups
Download actalog_backup_20250122.zip

# On PostgreSQL v0.8.1 instance
DB_DRIVER=postgres
make migrate  # Creates PostgreSQL schema
POST /api/admin/backups/upload  # Upload MariaDB backup
POST /api/admin/backups/{filename}/restore  # Data migrated!
```

## [0.8.0-beta] - 2025-11-22

### Changed
- **PostgreSQL Driver Migration (BREAKING for PostgreSQL users)**: Migrated from `lib/pq` to `pgx/v5` driver
  - **Dependency Change**: Removed `github.com/lib/pq v1.10.9`, added `github.com/jackc/pgx/v5 v5.7.6`
  - **Performance**: 10-30% faster for most PostgreSQL workloads
  - **Active Development**: pgx is actively maintained vs lib/pq in maintenance mode
  - **Better Features**: Improved support for PostgreSQL-specific features (LISTEN/NOTIFY, COPY, binary protocol)
  - **Context Support**: Better cancellation and timeout handling
  - **Backward Compatibility**: SQLite and MySQL/MariaDB unaffected, full backward compatibility maintained

### Added
- **PostgreSQL Schema Support**: Added `DB_SCHEMA` environment variable for schema isolation
  - Enables multi-tenant PostgreSQL deployments using database schemas
  - DSN now includes `search_path` parameter for schema targeting
  - Default schema: `public` (standard PostgreSQL behavior)
  - Example: `DB_SCHEMA=actalog` routes all operations to the actalog schema

- **Connection Pooling Configuration**: Fine-grained control over database connection pools
  - `DB_MAX_OPEN_CONNS`: Maximum simultaneous database connections (default: 25)
  - `DB_MAX_IDLE_CONNS`: Maximum idle connections kept ready (default: 5)
  - `DB_CONN_MAX_LIFETIME`: Maximum connection lifetime before recycling (default: 5m)
  - Applies to PostgreSQL and MySQL/MariaDB only (SQLite uses single connection)
  - Configurable per deployment for optimal resource usage
  - Updated `.env.example` with connection pooling examples and tuning guidance

- **Multi-Database Testing**: Comprehensive verification across all supported databases
  - ✅ SQLite (sqlite3): Fully backward compatible, all features tested
  - ✅ PostgreSQL (pgx/v5): Full migration verified with real database at 192.168.1.143
  - ✅ MariaDB/MySQL (mysql): Compatibility verified with real database at 192.168.1.234
  - All three databases: schema creation, migrations, seeding, and operations verified
  - Real-world connection pooling and schema isolation tested

- **Documentation**: Created comprehensive migration guide
  - `docs/POSTGRESQL_MIGRATION.md`: Complete migration guide for PostgreSQL users
  - Step-by-step migration instructions for existing lib/pq users
  - New PostgreSQL deployment instructions from scratch
  - Schema isolation configuration examples
  - Connection pooling tuning guidelines
  - Troubleshooting section for common issues
  - Performance comparison (pgx vs lib/pq)
  - Rollback instructions if needed
  - Test results for all three databases with real connection details

- **Docker Deployment Planning**: Added comprehensive Docker deployment roadmap to TODO.md
  - 50+ sub-tasks for complete Docker deployment solution
  - Documentation planning: `DOCKER_DEPLOYMENT.md`, `DOCKER_BUILD.md`, README updates
  - Implementation tasks: Dockerfile (multi-stage), docker-compose files for all 3 databases
  - GitHub Actions workflow for automated builds and publishing to ghcr.io
  - Multi-architecture support (amd64, arm64)
  - Testing across all deployment scenarios
  - Target: One-command deployment with `docker-compose up -d`
  - Target version: v0.9.0-beta

### Enhanced
- **Database Abstraction Layer**: Improved database compatibility handling
  - New helper functions: `getBoolValue()`, `getPlaceholders()` for database-agnostic SQL
  - SQL placeholders: SQLite/MySQL use `?`, PostgreSQL uses `$1, $2, $3`
  - Boolean values: SQLite uses `0/1`, PostgreSQL/MySQL use `TRUE/FALSE`
  - Timestamp functions: SQLite uses `datetime('now')`, PostgreSQL uses `CURRENT_TIMESTAMP`, MySQL uses `NOW()`
  - Insert ID retrieval: PostgreSQL uses `RETURNING id` clause instead of `LastInsertId()`
  - All seeding functions updated: `seedStandardMovements()`, `seedStandardWODs()`, `seedWorkoutTemplates()`
  - All helper functions updated: `createWorkout()`, `addWorkoutMovement()`, `addWorkoutMovementWithTime()`, `addWorkoutWOD()`, `getMovementIDByName()`, `getWODIDByName()`

- **DSN Format**: Updated PostgreSQL connection string format for pgx compatibility
  - Old format (lib/pq): `host=localhost port=5432 user=actalog dbname=actalog sslmode=disable`
  - New format (pgx): `postgres://user:password@host:port/database?sslmode=disable&search_path=schema`
  - Automatic schema path inclusion when `DB_SCHEMA` is configured
  - Full compatibility with PostgreSQL URIs

- **Configuration Files**: Updated all configuration examples
  - `.env.example`: Added DB_SCHEMA, connection pooling parameters, and tuning guidelines
  - `configs/config.go`: New DatabaseConfig fields (Schema, MaxOpenConns, MaxIdleConns, ConnMaxLifetime)
  - Environment loading functions: `getEnvInt()`, `getEnvDuration()` for typed config values
  - Default values optimized for production use

### Fixed
- **MariaDB Compatibility**: Fixed SQL syntax issues for MariaDB/MySQL
  - Fixed `addWorkoutWOD()` to use database-specific timestamp functions
  - Fixed `getMovementIDByName()` to use database-specific placeholders
  - Fixed `getWODIDByName()` to use database-specific placeholders
  - All helper functions now properly handle MySQL/MariaDB-specific SQL

### Technical Details
- **Build Number Range**: #47-56 (10 builds during migration)
- **Files Modified**:
  - Core: `go.mod`, `configs/config.go`, `internal/repository/database.go`
  - Commands: `cmd/actalog/main.go`, `cmd/migrate/main.go`, `cmd/check-schema/main.go`
  - Documentation: `.env.example`, `docs/POSTGRESQL_MIGRATION.md`
- **Breaking Changes**:
  - PostgreSQL users must update from lib/pq to pgx (see migration guide)
  - No breaking changes for SQLite or MySQL users
  - Database schemas and data remain fully compatible
- **Migration Path**: Existing PostgreSQL databases work without changes (DSN format updated automatically)

## [0.7.6-beta] - 2025-11-22

### Added
- **Backup Upload for Migration**: Added ability to upload external backup ZIP files from other systems
  - New upload button in AdminBackupsView with file picker for .zip files
  - `POST /api/admin/backups/upload` endpoint for multipart file upload
  - `UploadBackup()` service method with filename validation and ZIP verification
  - Timestamp-based renaming to prevent filename conflicts
  - Audit logging for all backup uploads with original filename tracking
  - Enables data migration between different ActaLog installations
  - Supports cross-database migrations (e.g., PostgreSQL backup restored to SQLite system)

- **Documentation Planning**: Comprehensive planning for future documentation systems added to TODO.md
  - **End-User Help Documentation System**: Multi-document help system with Markdown, screenshots, and Mermaid diagrams
    - Planned GitHub storage (docs/help/ directory) with links from Profile screen
    - Table of Contents, FAQ section, and "How do I..." tutorials
    - 8 tutorial topics covering key features (logging, PRs, templates, performance, imports, PWA)
    - 4 workflow diagrams (workout logging, PR detection, import/export, authentication)
    - Image placeholders with descriptive captions for future screenshots
    - Cross-referenced topics and troubleshooting section
  - **Administrator Documentation System**: Comprehensive admin guide for system operators
    - Planned GitHub storage (docs/admin/ directory) with admin-only access from Profile screen
    - 12 administrative tutorials (user management, backups, audit logs, security, troubleshooting)
    - 5 admin workflow diagrams (user lifecycle, backup/restore, security/audit, permissions, lockout process)
    - Security best practices section (password policy, JWT management, CORS, email, monitoring)
    - System configuration guide (environment variables, database setup, SMTP, PWA, deployment)
    - API endpoint reference for automation
  - **Test Coverage Planning**: Comprehensive testing strategy for both backend and frontend
    - Backend testing: 13 tasks covering unit tests, integration tests, mocking, transactions
    - Frontend testing: 10 tasks covering components, E2E flows, PWA functionality, routing
    - Testing infrastructure: 7 tasks including CI/CD, coverage reporting, E2E framework, performance testing
    - Documentation: 4 tasks for testing patterns, guidelines, data setup, and CI documentation
  - **Scheduled Remote Backups**: Future enhancement planning for automatic cloud backups
    - Support for 6 cloud providers (AWS S3, Google Cloud Storage, Azure, Dropbox, Google Drive, SFTP/FTP)
    - Configurable schedules (hourly, daily, weekly, monthly)
    - Retention policies, verification, notifications, bandwidth throttling
  - **Expanded Seed Data**: Planning for extracting additional WODs and Movements from import files
    - Parse PDFs and crossfit_wods.csv to expand standard movement and WOD library
    - Automated extraction and conversion to seed CSV format

### Enhanced
- **Audit Logging for Backup Operations**: Comprehensive audit trail for all backup activities
  - `backup_downloaded` event now logs file size in bytes asynchronously
  - `backup_restored` event now includes detailed statistics:
    - Total users, workouts, movements, and WODs restored
    - Provides visibility into restore scope and impact
  - All audit logs include user email and timestamp
  - Enables security monitoring and compliance tracking for data operations

### Fixed
- **Cross-Version Restore Compatibility**: Backup restore now handles schema version differences gracefully
  - Added table existence checks before DELETE and INSERT operations using `sqlite_master` queries
  - Tables missing in current schema are skipped with warnings instead of causing fatal errors
  - `restoreTable()` method validates table existence before attempting data restore
  - Enables restoring backups from different ActaLog versions (forward and backward compatibility)
  - Warning messages logged for skipped tables to aid troubleshooting
  - Prevents 500 errors when restoring backups created on different schema versions

## [0.7.5-beta] - 2025-11-22

### Added
- **Admin User Management - Complete Integration**: Fully activated admin user management dashboard
  - Activated "User Management" card in AdminView (`/admin`) - now clickable and navigates to user management
  - Removed "Coming Soon" placeholder status
  - Backend API endpoints from v0.4.6-beta now fully integrated with frontend UI

- **"Remember Me" Functionality**: Extended session duration for better user experience
  - Added checkbox to login form: "Remember me for 30 days"
  - Backend: Extended refresh token duration from 7 days to 30 days when Remember Me is checked
  - Modified `CreateRefreshToken()` method to accept `rememberMe` parameter (`internal/service/user_service.go:517`)
  - Updated login handler to pass Remember Me flag to service (`internal/handler/auth_handler.go:147`)
  - Frontend: Auth store already configured to send `remember_me` flag to API
  - Refresh tokens stored in localStorage for automatic session restoration
  - Audit logging for Remember Me token creation
  - Users who don't check "Remember Me" still get 7-day refresh tokens (default behavior)

- **Database Backup and Restore System - Complete Activation**: Full disaster recovery and data migration capability
  - Activated "Database Backups" card in AdminView (`/admin`) with orange database-export icon
  - Complete backup/restore functionality previously implemented but not activated
  - Backend fully implemented and wired up:
    - `internal/service/backup_service.go` - Full database export to ZIP with JSON data
    - `internal/handler/backup_handler.go` - All CRUD endpoints for backup management
    - Routes active in `cmd/actalog/main.go` under `/api/admin/backups`
  - Frontend fully implemented:
    - `AdminBackupsView.vue` - Complete backup management interface at `/admin/backups`
    - Create new backups with metadata (version, user counts, workout counts)
    - List all backups with creation date, creator email, stats, and file size
    - Download backups as ZIP files
    - Delete backups with confirmation dialog
    - Restore backups with strong warning dialog and confirmation requirement
    - Empty state for first-time use
  - API Endpoints (Admin-only):
    - `POST /api/admin/backups` - Create new backup
    - `GET /api/admin/backups` - List all backups
    - `GET /api/admin/backups/{filename}` - Download backup file
    - `GET /api/admin/backups/{filename}/metadata` - Get backup metadata
    - `DELETE /api/admin/backups/{filename}` - Delete backup
    - `POST /api/admin/backups/{filename}/restore` - Restore from backup
  - ZIP backup structure includes all database tables exported as JSON
  - **SQLite Database Dump**: All backups now include a portable SQLite database file (`actalog_backup.db`)
    - If running SQLite: Direct copy of production database included in ZIP
    - If running PostgreSQL/MySQL: New SQLite database created from exported data and included in ZIP
    - Provides universal, portable database format that can be opened with any SQLite tool
    - Enables easy data inspection and migration between database systems
  - Supports all database drivers (SQLite, PostgreSQL, MySQL)
  - Audit logging for all backup operations
  - Security: Filename validation prevents directory traversal attacks

### Changed
- **Mobile-Friendly Admin Table**: Restructured user management table for better mobile accessibility
  - Split single "Actions" column into 6 individual labeled columns (Details, Lock, Enable, Email, Change Role, Delete)
  - Clear column headers eliminate need for hover tooltips on mobile devices
  - Improved visual clarity with centered icons and consistent sizing
  - Better touch targets for mobile users

### Fixed
- **PR History Date Display**: Fixed PR history to show workout date instead of record creation date
  - Backend: Added `WorkoutDate` assignment in `GetPRMovements()` and `GetPRWODs()` repository methods
  - Frontend: Changed `PRHistoryView.vue` to display `workout_date` instead of `created_at`
  - PR dates now show when the workout was performed (e.g., "Fri, Oct 10, 2024")
  - Previously showed database record creation timestamp which was incorrect for imported data
  - Affects both movement PRs and WOD PRs

- **AdminBackupsView Layout**: Fixed Vuetify bottom navigation layout error
  - Restructured container hierarchy (outer `<v-container>` → `<div>`)
  - Eliminated "Could not find layout item 'bottom-navigation'" console error
  - Fixed scroll behavior with proper overflow handling
  - Resolved layout conflicts when navigating from ProfileView
  - Applied same container pattern from AdminUsersView, WODLibraryView, MovementsLibraryView

- **AdminUsersView Layout**: Fixed Vuetify bottom navigation layout error
  - Restructured container hierarchy (outer `<v-container>` → `<div>`)
  - Eliminated "Could not find layout item 'bottom-navigation'" console error
  - Fixed scroll behavior with proper overflow handling
  - Resolved layout conflicts when navigating from ProfileView

## [0.7.4-beta] - 2025-11-22

### Added
- **Quick Log Buttons on Library Cards**: Added Quick Log functionality directly to WOD and Movement library card views
  - Teal lightning bolt icon buttons on each WOD card in WOD Library view
  - Teal lightning bolt icon buttons on each Movement card in Movements Library view
  - Quick Log dialog opens directly from cards without navigating to detail pages
  - Pre-populated forms with selected WOD or Movement data
  - Streamlined workout logging workflow from library browsing
- **Quick Log Buttons on Detail Pages**: Enhanced WOD and Movement detail screens
  - Added prominent Quick Log buttons to WODDetailView and MovementDetailView
  - Pre-populated Quick Log dialogs with current item being viewed
  - Consistent user experience across all viewing contexts
- **Admin User Management Dashboard**: Complete administrative control system for user accounts
  - Activated Admin User Management card in AdminView (`/admin`)
  - Full-featured user management UI at `/admin/users` route
  - User list table with pagination (50 users per page) and real-time search
  - Mobile-optimized table with individual labeled columns (no hover tooltips required):
    - **Details**: View comprehensive user information dialog
    - **Lock**: Unlock temporarily locked accounts from failed login attempts
    - **Enable**: Enable/disable user accounts with optional reason tracking
    - **Email**: Toggle email verification status manually
    - **Change Role**: Switch between "user" and "admin" roles
    - **Delete**: Permanently delete users with confirmation dialog
  - User details dialog showing all account information:
    - Email verification status with visual badges
    - Account status (Active/Disabled) with color coding
    - Role display with chips
    - Timestamps: created_at, last_login_at, email_verified_at, disabled_at
    - Disable reason display when applicable
  - Color-coded status indicators throughout (green/success, red/error, purple/admin, blue/user)
  - Confirmation dialogs for destructive actions (disable, delete)
  - Success/error messaging for all operations
  - Backend API endpoints from v0.4.6-beta now fully integrated with frontend

### Changed
- **Icon Consistency**: Unified Quick Log iconography across the entire application
  - All Quick Log buttons now use `mdi-lightning-bolt` icon (teal color)
  - Replaced `mdi-play-circle` icons in WorkoutsView template cards with lightning bolt
  - Consistent visual language for Quick Log feature throughout the app
  - Tooltips added to all Quick Log buttons for clarity
- **User Management Table Structure**: Improved accessibility and mobile usability
  - Restructured "Actions" column into 6 individual labeled columns
  - Clear column headers eliminate need for hover tooltips on mobile devices
  - Centered icon buttons with consistent sizing and colors
  - Enhanced visual feedback for current state (locked/unlocked, enabled/disabled, verified/unverified)

### Fixed
- **Vuetify Layout Issues**: Fixed bottom navigation layout conflicts
  - Restructured WODLibraryView, MovementsLibraryView, and AdminUsersView container hierarchies
  - Changed outer containers from `<v-container>` to `<div>` to prevent layout system conflicts
  - Moved bottom navigation outside scrollable content containers
  - Eliminated "Could not find layout item 'bottom-navigation'" console errors
  - Fixed scroll behavior with proper `overflow-y: auto` and `max-height` constraints
- **Template Deletion Bug**: Fixed custom template deletion endpoint error
  - Corrected API endpoint from `DELETE /api/workouts/{id}` to `DELETE /api/templates/{id}`
  - Resolved 500 Internal Server Error and "unauthorized workout access" issue
  - Custom workout templates now delete successfully

## [0.7.3-beta] - 2025-01-22

### Added
- **Quick Log on Performance Screen**: Complete Quick Log dialog integration on Performance view
  - Quick Log button now opens dialog directly on Performance screen (no navigation to Dashboard)
  - Pre-populates with the movement or WOD currently being viewed
  - User can change selection within dialog if needed
  - Automatically refreshes performance data after successful submission
  - Maintains user context and viewing state

### Fixed
- **Performance Chart Sorting**: Fixed chronological ordering for workouts on the same date
  - Implemented two-level sorting: primary by `workout_date`, secondary by `created_at` timestamp or `id`
  - Ensures newest entries appear on the right side of charts (chronological order)
  - Prevents multiple same-day workouts from appearing in database order
  - Applied to both movement and WOD performance charts

## [0.7.2-beta] - 2025-01-22

### Added
- **1RM (One-Rep Max) Calculation and Display**: Complete system for tracking estimated strength maximums
  - **Backend**: Enhanced `/api/performance/movements/{id}` endpoint to calculate and return 1RM data
    - Added `MovementPerformanceWithRM` response type with `calculated_1rm` and `formula` fields
    - Returns `best_1rm` and `best_formula` for overall best performance
    - Uses hybrid formula approach from `pkg/prmath/one_rm.go`:
      - 1 rep = Actual weight
      - 2-10 reps = Epley formula: `1RM = weight × (1 + reps/30)`
      - 11+ reps = Wathan formula: `1RM = (100 × weight) / (48.8 + 53.8 × e^(-0.075 × reps))`
  - **Frontend - Best 1RM Card**: New stat card displaying estimated 1RM
    - Prominent gold-colored display (#ffc107) with arm-flex icon
    - Shows rounded 1RM value with "lbs (estimated)" label
    - Displays formula chip indicating calculation method
    - Only appears when weight/reps data is available
  - **Frontend - Performance History**: Enhanced history entries with 1RM estimates
    - Shows "Est. 1RM: XXX lbs" in gold text for each performance record
    - Appears alongside date and notes in subtitle line
  - **Frontend - Chart Enhancements**: Dual-line performance chart
    - Added dashed gold line showing estimated 1RM trend
    - Original solid dark line shows actual weight lifted
    - Legend automatically displays when 1RM data exists
    - Enhanced tooltips showing both actual weight and estimated 1RM with formula
    - Null value filtering prevents gaps in chart display
  - **Y-Axis Labels**: Added clear axis labels to all performance charts
    - Movement charts: "Weight (lbs)"
    - WOD charts: Dynamic labels based on score type (Time/Rounds/Weight)

### Fixed
- **WOD Chart Rendering Issue**: Fixed canvas element not rendering on initial load
  - Moved `loadingPerformance = false` before chart rendering to ensure DOM updates
  - Added proper `await nextTick()` sequencing for canvas availability
  - Resolved null reference errors in WOD performance charts

### Changed
- **Code Cleanup**: Removed debug console.log statements from PerformanceView
  - Removed WOD debug logging throughout fetchPerformanceData and renderWODChart
  - Cleaner console output in production

## [0.7.1-beta] - 2025-01-22

### Fixed
- **Wodify Import Date Issue**: Fixed performance charts showing all imported workouts as "today" instead of actual workout dates
  - Backend: Added `WorkoutDate` field to `UserWorkoutMovement` and `UserWorkoutWOD` domain models
  - Backend: Updated repositories to populate `workout_date` from `user_workouts.workout_date` join
  - Frontend: Changed PerformanceView to use `workout_date` instead of `created_at` for all date displays
  - Charts now correctly show historical dates (e.g., Jul 30, 2018) from Wodify CSV imports
  - History grouping and sorting now use actual workout dates

### Changed
- **Performance Chart Date Display**: Charts now display full dates with year
  - X-axis labels show year: "Jul 30, 2018" instead of "Jul 30"
  - Hover tooltips display full date with year in title
  - Applied to both Movement and WOD performance charts
- **Rep Scheme Filter Enhancement**: Improved dropdown filter in Performance view
  - Changed "All Reps" to simplified "All" option
  - "All" displays all weighted records regardless of rep scheme, sets, or other factors
  - Cleaner, more intuitive filtering experience

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

---

## [0.4.3-beta] - 2025-01-14

### Changed
- **UI Spacing Improvements**: Reduced whitespace and padding throughout the application
  - Reduced top margin from 36px to 5px on all main views (Dashboard, Profile, Performance, Workouts)
  - Reduced card padding: `pa-4` → `pa-2`, `pa-3` → `pa-2`, `pa-2` → `pa-1`
  - Reduced section margins: `mb-3` → `mb-2`, `mb-2` → `mb-1`
  - Reduced form field spacing from `mb-2` to `mb-1`
  - Removed top padding from main containers (`pt-0`)
  - Changed border radius from `rounded="lg"` to `rounded` for tighter appearance
  - Applied changes across Dashboard, Profile, Performance, and Workouts views
  - Result: More compact, efficient use of screen space on mobile devices

## [0.4.2-beta] - 2025-01-14

### Added
- **Version and Build Display**: Added version and build number display in Profile screen
  - Created new version display card at top of Profile screen
  - Shows full version (e.g., "0.4.2-beta+build.1") and build number
  - Backend exposes `/api/version` endpoint (public, no auth required)
  - Returns `version`, `build`, `fullVersion`, and `app` fields
  - Frontend fetches version info on Profile page load
- **Automatic Build Number Increment System**
  - Created `scripts/increment-build.sh` for automatic build number management
  - Updated `Makefile` to auto-increment build on every `make build`
  - Build number stored in `pkg/version/version.go` as `Build` constant
  - Added `FullVersion()`, `BuildNumber()`, and `FullString()` functions
  - Format: `Major.Minor.Patch-PreRelease+build.N` (e.g., "0.4.2-beta+build.1")

### Changed
- Version endpoint moved from `/version` to `/api/version` for Vite proxy compatibility
- Updated `CLAUDE.md` with comprehensive Build Number Auto-Increment documentation

## [0.4.1-beta] - 2025-01-14

### Fixed
- **Quick Log movement search**: Fixed autocomplete not displaying movements in Quick Log dialog
  - Added loading states for movements and WODs
  - Added auto-select-first for better search UX
  - Added search icon to match design patterns
  - Added console logging for debugging data loading
- **Localhost hardcoded URLs**: Fixed profile pictures and assets not working outside of localhost
  - Created `web/src/utils/url.js` with dynamic URL resolution utilities
  - `getApiBaseUrl()` - Environment-aware API base URL
  - `getAssetUrl()` - Converts relative paths to absolute URLs
  - `getProfileImageUrl()` - Specifically handles profile image URLs
  - Updated `web/src/stores/auth.js` to use new URL utilities
  - Updated `web/src/views/ProfileView.vue` to use new URL utilities
  - Fixed `web/src/views/VerifyEmailView.vue` to use relative URLs
  - Added `/uploads` proxy to Vite dev server configuration
- **Axios configuration**: Changed baseURL to use relative URLs to leverage Vite proxy in development

### Added
- Created `web/.env.example` documenting `VITE_API_BASE_URL` environment variable
- Added comprehensive URL utility functions for production deployments

### Changed
- Quick Log dialog now opens immediately and fetches data in background for better UX
- Updated Vite proxy configuration to handle both `/api` and `/uploads` routes

## [0.4.0-beta] - 2025-01-13

### Added - Multi-Database Support
- **Multi-database support**: SQLite, PostgreSQL, and MySQL/MariaDB
- **Database migration system** with version tracking and rollback support
- Database-agnostic DSN builder
- Driver-specific schema generation (SQLite, PostgreSQL, MySQL)
- Comprehensive DATABASE_SUPPORT.md documentation
- Support for database-agnostic migrations with driver parameter

### Added - Workout Logging (Backend Complete)
- **Workout logging functionality** with complete CRUD operations
- **Movement database** with 82 standard CrossFit movements (auto-seeded)
- **Progress tracking** by movement for PR analysis
- API endpoints for workout management:
  - POST /api/workouts - Create workout with movements
  - GET /api/workouts - List workouts with pagination and date filtering
  - GET /api/workouts/{id} - Get workout details
  - PUT /api/workouts/{id} - Update workout
  - DELETE /api/workouts/{id} - Delete workout (cascade deletes movements)
  - GET /api/progress/movements/{movement_id} - Track performance history
- Movement management API endpoints:
  - GET /api/movements - List standard movements
  - GET /api/movements/search - Search movements by name
  - GET /api/movements/{id} - Get movement details
  - POST /api/movements - Create custom movement

### Added - Design Refinements (Planned for v0.3.0)
**Refined design decisions documented** through user consultation (not yet implemented):

**Email Verification System:**
- Optional email verification with feature unlock approach
- Users can immediately use core features without verification
- Email verification unlocks leaderboard participation and data export
- Verification email sent on registration with resend capability
- Added `email_verified` and `email_verified_at` fields to users table

**Personal Records (PR) Tracking:**
- Auto-detection system for PRs:
  - Highest weight for strength movements (per user, per movement)
  - Fastest time for time-based WODs (per user, per WOD)
  - Most rounds+reps for AMRAP WODs (per user, per WOD)
- Manual PR flag/unflag capability for user corrections
- PR badges displayed on workout cards in dashboard and history
- PR indicators (⭐) shown in movement history lists
- Added `is_pr` field to `workout_wods` and `workout_strength` tables

**Leaderboard System with Scaled Divisions:**
- Three-division leaderboard system:
  - **Rx (As Prescribed)**: Workout performed exactly as specified
  - **Scaled**: Modified workout (lighter weight, fewer reps, substitute movements)
  - **Beginner**: Simplified version for newer athletes
- Users self-select division when logging WOD scores
- Separate leaderboards for each division to ensure fair comparisons
- Global leaderboards for standard benchmark WODs
- Email verification required for leaderboard participation
- Added `division` field to `workout_wods` table

**Hybrid Workout Template System:**
- Users can use pre-defined WODs and admin-created templates
- Users can create and save their own custom workout templates
- "Save as Template" option when logging workouts
- Template management UI for create, edit, delete operations
- Both standard and custom content searchable and filterable

**Hybrid Movement/WOD Libraries:**
- Pre-defined library of standard CrossFit movements and WODs
- Users can add custom movements and WODs
- `is_standard` flag distinguishes pre-defined vs. user-created content
- Standard content cannot be edited by regular users
- Added `is_standard` field to `wods` and `strength_movements` tables

**Workout Scheduling:**
- Users can schedule workouts for future dates
- Calendar view distinguishes scheduled vs. completed workouts
- "Complete Scheduled Workout" flow for pre-planned training
- No push notifications initially (infrastructure ready for future)

**Performance Analytics:**
- Weight progression charts for strength movements
- Workout frequency heatmap showing consistency and streaks
- WOD leaderboards with division filters
- Focus on three primary visualizations

**Import/Export Enhancements:**
- Support for three formats: CSV, JSON, and Markdown
- CSV for spreadsheet compatibility and data analysis
- JSON for complete structured backup/restore
- Markdown for formatted workout reports
- Date range selection for partial exports
- Data type selection (Workouts, WODs, Movements, Profile)

**Data Sync Strategy:**
- "Last write wins" conflict resolution for offline sync
- Most recent timestamp takes precedence
- Suitable for single-user workout logging scenarios
- Sync status indicator for pending operations

**User Roles:**
- Simple two-tier system: regular users and admins
- First user becomes admin automatically
- No coach or gym owner roles in initial version

### Added - Database Schema Design (Planned for v0.3.0)
- **Major schema redesign** based on logical data model requirements (documented but not yet implemented)
- New `wods` table for predefined CrossFit workouts with comprehensive attributes:
  - Source (CrossFit, Other Coach, Self-recorded)
  - Type (Benchmark, Hero, Girl, Notables, Games, Endurance, Self-created)
  - Regime (EMOM, AMRAP, Fastest Time, etc.)
  - Score Type (Time, Rounds+Reps, Max Weight)
  - Description, URL, and notes fields
- New `user_workouts` junction table linking users to workout instances on specific dates
- New `workout_wods` junction table linking workouts to WODs with scoring
- New `user_settings` table for user preferences (theme, notifications, export format)
- New `audit_logs` table for audit trail and accountability
- Added `updated_by` tracking to all entities for audit purposes

### Changed - Database Schema Design (Planned for v0.3.0)
- **Workouts** are now reusable templates (not user-specific instances)
- Renamed `movements` table to `strength_movements`
- Added `movement_type` to strength_movements (weightlifting, cardio, gymnastics)
- Renamed `workout_movements` to `workout_strength`
- Removed user-specific fields from workouts table (user_id, workout_date, workout_type)
- Updated ERD to reflect many-to-many relationships properly

### Changed - Multi-Database Support
- Updated migration system to accept driver parameter for database-agnostic migrations
- Improved table existence checking across all database types
- Enhanced schema creation with database-specific SQL dialects

### Migration Required (Future Work)
- Database migration from v0.1.0 to v0.3.0 will be needed when implementing v0.3.0
- See DATABASE_SCHEMA.md for planned migration steps
- Backend domain models will need updates
- API endpoints will need refactoring for new structure

### UI Updates - Dashboard Redesign
- New Dashboard UI matching design specifications
- Calendar component showing monthly workout activity
- Recent workouts cards with grouped display
- Top app bar with ActaLog logo and current date
- Unified bottom navigation across all authenticated views
- Avatar support for user profile icon
- Workout badge for Personal Records (PRs)
- Complete Dashboard redesign with calendar view
- Moved header and bottom navigation to App.vue for consistency
- Updated color scheme to match brand guidelines
- Improved mobile-first responsive design
- Enhanced bottom navigation with better iconography

### Documentation
- **Reorganized app navigation structure** - Settings Menu as central hub
- Added comprehensive "Screens & Navigation Flow" section to REQUIREMENTS.md
  - **33 core screens** defined with routes, purposes, and components
  - Settings Menu flyout accessed from user avatar
  - Management screens for WODs, Strength Movements, and Workout Templates with full CRUD operations
  - Import/Export data screens
  - App Preferences screen
  - Navigation flow diagrams
  - Screen interaction patterns
  - PWA-specific screens (install prompt, offline indicator)
- Added `birthday` field to User profile

### Planned
- Implement database migration scripts for v0.3.0 schema
- Update backend domain models for new schema
- Seed data for standard WODs and movements
- Connect frontend to workout logging APIs
- Workout templates and named WOD database
- Charts and graphs for progress visualization
- Push notifications for workout reminders
- Web Share API integration
- Implement all 33 screens defined in screen inventory:
  - Management screens for WODs (List, Create, Edit with CRUD operations)
  - Management screens for Strength Movements (List, Create, Edit with CRUD operations)
  - Management screens for Workout Templates (List, Create, Edit with CRUD operations)
  - Import/Export data screens
  - Settings Menu flyout implementation
- First-user-as-admin logic
- Configurable registration control (ALLOW_REGISTRATION)
- SQLite database with auto-initialization
- PostgreSQL and MariaDB support
- Database schema with users, workouts, movements, and workout_movements tables
- Bcrypt password hashing (cost factor 12)
- CORS middleware with configurable origins
- Request logging middleware
- Health check endpoint (`/health`)
- Version endpoint (`/version`)
- Docker and docker-compose configuration
- Makefile for development workflow
- Windows batch script (`build.bat`) for Windows users
- Comprehensive documentation:
  - README.md with quick start guide
  - ARCHITECTURE.md with Clean Architecture patterns
  - DATABASE_SCHEMA.md with ERD diagrams
  - SETUP.md for local and Docker development
  - REQUIREMENTS.md with user stories
  - AI_INSTRUCTIONS.md for development guidelines
- Frontend views:
  - Login and registration pages
  - Dashboard with bottom navigation
  - Workout logging form (matching design)
  - Workouts history view
  - Performance tracking view
  - Profile and settings views
  - 404 error page
- Vue Router with authentication guards
- Pinia state management for auth
- Axios HTTP client with interceptors
- Custom ActaLog theme with design colors
- Mobile-first responsive design
- ESLint 9 with flat config format
- Prettier code formatting
- golangci-lint configuration
- Version management system (v0.1.0-alpha)

### Fixed
- Windows build permission issues (uses project-local cache)
- SQLite driver name corrected from 'sqlite' to 'sqlite3'
- npm dependency deprecation warnings
- esbuild security vulnerability
- ESLint 8 to ESLint 9 migration
- CORS configuration for development

### Security
- JWT token generation and validation
- Password hashing with bcrypt
- SQL injection prevention via parameterized queries
- CORS origin whitelisting
- Secure defaults in configuration
- No sensitive data in error responses

### Changed
- Updated all npm dependencies to latest versions
- Migrated from ESLint 8 to ESLint 9
- Updated Vite to version 6
- Updated Vue.js to version 3.5
- Updated Vuetify to version 3.7

### Developer Experience
- Hot reload support for frontend (Vite)
- Clean build artifacts with `make clean`
- Formatted code with `make fmt`
- Linting with `make lint`
- Testing support with `make test`
- Docker support for easy deployment
- Cross-platform build scripts (Makefile + build.bat)

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

---

## Version History Format

### [Version] - YYYY-MM-DD

#### Added
New features that have been added to the project.

#### Changed
Changes in existing functionality.

#### Deprecated
Soon-to-be removed features.

#### Removed
Features that have been removed.

#### Fixed
Bug fixes.

#### Security
Security-related changes or fixes.

---

**Current Version:** 1.2.1
**Schema Version:** v0.32.0 (Coach Role & Role Rename)
**Last Updated:** 2026-04-01
