# ActaLog Development Roadmap

**Current Version:** 0.5.1-beta
**Last Updated:** 2025-11-21
**Overall Completion:** ~75% of core requirements

---

## Executive Summary

ActaLog is a mobile-first Progressive Web App (PWA) for CrossFit workout tracking. The application is **functional for personal use** with core features implemented including user authentication, workout logging, performance tracking, and import/export capabilities. However, several critical features are needed before multi-user production deployment.

---

## Version History & Status

### v0.5.1-beta (Current - In Progress)
**Status:** Import/Export implementation complete, one critical bug pending

**Completed:**
- ✅ WOD export to CSV
- ✅ Movement export to CSV
- ✅ User Workouts export to JSON
- ✅ WOD import with preview and validation
- ✅ Movement import with preview and validation
- ✅ User Workouts import preview
- ⚠️ User Workouts import confirm (has persistence bug)

**Remaining:**
- 🔴 Fix User Workouts import persistence bug
- 🟡 Flattened CSV export for spreadsheet analysis
- 🟡 Performance data export with PR flags
- 🟢 Export history tracking (future)

### v0.5.0-beta (Released)
**Status:** Core workout system complete

**Features:**
- ✅ Template-based workout architecture
- ✅ WOD management (create, list, search, update, delete)
- ✅ User workout logging with movements and WODs
- ✅ Monthly workout statistics
- ✅ Performance tracking and search

### v0.4.6-beta (Released)
**Features:**
- ✅ Session management (list, revoke, revoke-all)
- ✅ Admin user management (CRUD operations)
- ✅ Account unlock/disable/enable by admins
- ✅ User role management
- ✅ Email verification toggle

### v0.3.1-beta (Released)
**Features:**
- ✅ Email verification system
- ✅ Verification email sending on registration
- ✅ Resend verification functionality

### v0.3.0-beta (Released)
**Features:**
- ✅ Personal Records (PR) tracking for movements
- ✅ PR auto-detection and manual toggle
- ✅ Password reset flow
- ✅ Profile picture upload

### v0.2.0-beta (Released)
**Features:**
- ✅ PWA implementation with service worker
- ✅ Offline support with IndexedDB
- ✅ Background sync queue
- ✅ Auto-update notifications

---

## Completion Analysis

### ✅ Completed Features (75%)

#### Authentication & User Management
- User registration with email verification
- Login/logout with JWT tokens and refresh tokens
- Password reset flow
- Profile management with avatar upload
- Session management
- Admin user management (full CRUD)
- Account lockout after failed login attempts

#### Workout System
- WOD management (10 standard WODs seeded)
- Movement library (31 standard movements seeded)
- Custom WOD and movement creation
- Workout templates (create, list, update, delete)
- User workout logging (movements + WODs)
- Workout history with filtering
- Monthly workout statistics

#### Performance Tracking
- PR tracking for movements (weight-based)
- PR auto-detection when logging workouts
- Manual PR flag toggle
- PR history view
- Retroactive PR flagging
- Performance search and charts

#### Import/Export System
- WOD export/import (CSV)
- Movement export/import (CSV)
- User Workouts export (JSON)
- Import preview with validation
- Duplicate detection

#### PWA Features
- Service worker with caching strategies
- Web app manifest
- Offline support with IndexedDB
- Background sync queue
- Auto-update notifications
- Installable on mobile/desktop

#### Admin Features
- Admin-only routes with middleware
- Data cleanup tools
- Audit log viewing and cleanup
- User account management UI

#### Security
- JWT authentication with refresh tokens
- Password hashing with bcrypt
- CORS configuration
- Rate limiting on auth endpoints
- SQL injection protection
- Account lockout protection

### ⚠️ Incomplete Features (25%)

#### High Priority (Blockers for Production)
1. **User Workouts Import Bug** - Import confirm doesn't persist data
2. **Database Backup/Restore** - Critical for production deployment
3. **Calendar/Timeline Views** - Core user story requirement
4. **Visual Progress Charts** - Core user story requirement
5. **WOD PR Tracking** - Only weight-based PRs work, missing time/AMRAP
6. **Test Coverage** - Currently 68%, need 80%+

#### Medium Priority (Enhanced Features)
7. Leaderboard system (by division: rx, scaled, beginner)
8. User settings management UI
9. Workout history filters (by type, movement, date range)
10. Flattened CSV export for spreadsheet analysis
11. PWA icon generation (all sizes)
12. Custom install prompt
13. Admin reporting (activity, performance)

#### Low Priority (Future Enhancements)
14. Workout scheduling (future dates)
15. Push notifications
16. Periodic background sync
17. Web Share API
18. PostgreSQL driver migration (lib/pq → pgx)
19. Accessibility compliance (WCAG 2.1 AA)
20. Redis session storage

---

## Development Plan

### Phase 1: v0.5.1-beta Completion (2-4 hours)
**Goal:** Fix critical import bug

**Tasks:**
- [ ] Debug User Workouts import confirm endpoint
- [ ] Investigate transaction/database constraint issues
- [ ] Add detailed error logging
- [ ] Test import with various scenarios
- [ ] Verify data persistence after import

**Success Criteria:**
- User Workouts import successfully creates workouts
- Data visible in `/api/workouts` and export endpoints
- No silent failures or transaction rollbacks

---

### Phase 2: v0.6.0-beta - Database Backup/Restore (2-3 days)
**Goal:** Production-critical data protection

**Backend Tasks:**
- [ ] Create `internal/service/backup_service.go`
  - [ ] `CreateBackup()` - Export all tables to JSON + files to ZIP
  - [ ] `ListBackups()` - Return backup metadata
  - [ ] `GetBackupMetadata()` - Read metadata from backup file
  - [ ] `DeleteBackup()` - Remove backup with audit log
  - [ ] `RestoreBackup()` - Full restore from backup
- [ ] Create `internal/handler/backup_handler.go`
  - [ ] `POST /api/admin/backups` - Create backup
  - [ ] `GET /api/admin/backups` - List backups
  - [ ] `GET /api/admin/backups/{filename}` - Download backup
  - [ ] `GET /api/admin/backups/{filename}/metadata` - Get metadata
  - [ ] `DELETE /api/admin/backups/{filename}` - Delete backup
  - [ ] `POST /api/admin/backups/{filename}/restore` - Restore backup
- [ ] Wire up routes in `cmd/actalog/main.go`
- [ ] Create `backups/` directory with .gitignore

**Frontend Tasks:**
- [ ] Create `web/src/views/AdminBackupsView.vue`
  - [ ] Backup list table with metadata
  - [ ] Create backup button with progress
  - [ ] Download/delete actions per backup
  - [ ] Restore with strong confirmation dialog
  - [ ] Empty state for no backups
- [ ] Add route `/admin/backups` to router
- [ ] Add navigation link in admin menu

**Testing:**
- [ ] Unit tests for BackupService
- [ ] Integration tests for backup/restore workflow
- [ ] Test with all database drivers (SQLite, PostgreSQL, MySQL)
- [ ] Manual testing: create, download, restore, delete

**Success Criteria:**
- Admin can create full database backups
- Backups include all data + uploaded files
- Restore successfully rebuilds database
- Works across different database drivers

---

### Phase 3: v0.7.0-beta - Enhanced UX (3-4 days)
**Goal:** Calendar views and visual charts

#### 3.1 Calendar/Timeline Views (1-2 days)
- [ ] Create `web/src/views/WorkoutCalendarView.vue`
  - [ ] Calendar component with workout dots
  - [ ] Click date to view workouts
  - [ ] Month/week navigation
  - [ ] Color coding by workout type
- [ ] Create `web/src/views/WorkoutTimelineView.vue`
  - [ ] Chronological timeline with cards
  - [ ] Infinite scroll or pagination
  - [ ] Filters by type, movement, date range
- [ ] Add routes to router
- [ ] Add navigation from dashboard/profile
- [ ] API endpoints (may need new query params)

#### 3.2 Visual Progress Charts (1-2 days)
- [ ] Install chart library (Chart.js or similar)
- [ ] Create `web/src/components/charts/WeightProgressChart.vue`
- [ ] Create `web/src/components/charts/WorkoutFrequencyChart.vue`
- [ ] Create `web/src/components/charts/MovementVolumeChart.vue`
- [ ] Integrate charts into dashboard
- [ ] Add date range selector
- [ ] API endpoints for aggregated data

#### 3.3 WOD PR Tracking (1 day)
- [ ] Update `DetectAndFlagPRs()` service method
  - [ ] Add time-based PR detection (faster time)
  - [ ] Add AMRAP PR detection (more rounds+reps)
- [ ] Update repository methods
  - [ ] `GetBestTimeForWOD()`
  - [ ] `GetBestRoundsRepsForWOD()`
- [ ] Test retroactive PR flagging for WODs
- [ ] Update frontend to show WOD PRs

**Success Criteria:**
- Users can view workouts in calendar format
- Timeline view shows chronological history
- Charts visualize progress over time
- WOD PRs detected for time and AMRAP workouts

---

### Phase 4: v0.8.0-beta - Testing & Polish (2-3 days)
**Goal:** Production stability

- [ ] Increase unit test coverage to 80%+
  - [ ] WODService tests
  - [ ] WorkoutWODService tests
  - [ ] BackupService tests
  - [ ] Repository tests
- [ ] Integration tests for all v0.5.0+ endpoints
- [ ] Frontend component tests (Vitest)
- [ ] End-to-end testing (Playwright or Cypress)
- [ ] Performance testing
- [ ] Security audit
- [ ] Accessibility audit (basic)

**Success Criteria:**
- >80% test coverage
- All critical paths have integration tests
- No high-severity security vulnerabilities
- Basic accessibility compliance

---

### Phase 5: v1.0.0 - Production Release (1-2 days)
**Goal:** Production-ready deployment

- [ ] Generate all PWA icons (72px - 512px)
- [ ] Create apple-touch-icon.png
- [ ] Production HTTPS setup documentation
- [ ] Nginx configuration guide
- [ ] Database migration guide
- [ ] Backup/restore documentation
- [ ] User guide
- [ ] API documentation
- [ ] Deployment checklist
- [ ] Version bump to 1.0.0

**Success Criteria:**
- Application passes production readiness checklist
- All documentation complete
- PWA installable on all platforms
- Zero critical bugs

---

## Post-1.0 Roadmap (v1.1 - v1.3)

### v1.1.0 - Leaderboards (2-3 days)
- Leaderboard system by division (rx, scaled, beginner)
- Ranking algorithm
- User rank display
- Division selector when logging WOD scores

### v1.2.0 - User Settings & Filters (2-3 days)
- Settings management UI
- Notification preferences
- Export format preferences
- Workout history filters (type, movement, date)
- Flattened CSV export for spreadsheets

### v1.3.0 - Admin & Reports (2-3 days)
- Admin dashboard with statistics
- User activity reports
- Performance monitoring
- Global data export
- User data import for migration

---

## Long-Term Vision (v2.0+)

### Future Enhancements
- Workout scheduling (plan future workouts)
- Push notifications for reminders
- Web Share API for workout sharing
- Social features (friend system, comments)
- Coach/athlete relationship features
- Gym/box management
- Nutrition tracking integration
- Wearable device sync (Apple Watch, Garland, etc.)
- Mobile native apps (if PWA limitations encountered)

### Infrastructure Improvements
- PostgreSQL driver migration (lib/pq → pgx)
- Redis for session storage
- Horizontal scaling support
- Advanced caching strategies
- CDN integration
- Multi-region deployment
- Real-time sync with WebSockets

### Enterprise Features
- Multi-tenant support
- Custom branding
- Advanced analytics
- Data warehouse integration
- API for third-party integrations
- Webhooks for events

---

## Estimated Timeline

### Critical Path to v1.0 (Production Ready)
- **Phase 1:** Fix import bug - 2-4 hours
- **Phase 2:** Backup/restore - 2-3 days
- **Phase 3:** UX enhancements - 3-4 days
- **Phase 4:** Testing & polish - 2-3 days
- **Phase 5:** Production prep - 1-2 days

**Total: 10-14 days** of focused development

### Enhanced Features (v1.1 - v1.3)
- **v1.1:** Leaderboards - 2-3 days
- **v1.2:** Settings & filters - 2-3 days
- **v1.3:** Admin features - 2-3 days

**Total: Additional 7-12 days**

### Grand Total
**17-26 days** (3-5 weeks) to feature-complete v1.3

---

## Success Metrics

### v1.0 Release Criteria
- ✅ All high-priority features implemented
- ✅ >80% test coverage
- ✅ Zero critical bugs
- ✅ Database backup/restore functional
- ✅ PWA fully installable on iOS/Android/Desktop
- ✅ Production deployment documentation complete
- ✅ User guide published
- ✅ Security audit passed

### v1.3 Release Criteria
- ✅ All medium-priority features implemented
- ✅ Leaderboards functional
- ✅ User settings configurable
- ✅ Admin reporting operational
- ✅ >85% test coverage
- ✅ Performance benchmarks met
- ✅ Accessibility compliance (basic)

---

## Risk Assessment

### Critical Risks
1. **Database backup/restore complexity** - Multi-database support may have edge cases
   - Mitigation: Thorough testing with all database drivers
2. **Import bug may be schema-related** - Could indicate deeper issues
   - Mitigation: Detailed debugging and transaction logging
3. **Chart library performance** - May slow down on large datasets
   - Mitigation: Data pagination and lazy loading

### Medium Risks
4. **PWA icon generation** - Design consistency across sizes
   - Mitigation: Use automated tools and test on real devices
5. **Test coverage target** - May uncover hidden bugs
   - Mitigation: Gradual increase with regression testing

---

## Notes

- This roadmap is based on current codebase analysis (v0.5.1-beta)
- Timeline assumes single developer working full-time
- Priorities may shift based on user feedback
- Production deployment assumes self-hosted infrastructure
- Cloud deployment (AWS/GCP/Azure) would require additional configuration

**Last Review:** 2025-11-21
**Next Review:** After v0.5.1-beta completion
