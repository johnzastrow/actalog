# ActaLog TODO

> **Last Updated:** 2025-11-28
> **Current Version:** 0.12.2-beta

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

*(none)*

---

## Backlog

### High Priority

#### Testing Coverage
- [ ] `[HIGH]` **Add handler unit tests** - auth_handler, user_workout_handler, movement_handler, wod_handler
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
- [ ] `[MEDIUM]` **Calendar View** - Monthly view with workout dots
- [ ] `[MEDIUM]` **Timeline View** - Chronological workout history
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

*Last scanned: 2025-11-28*

---

## Completed Releases

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
