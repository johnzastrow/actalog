# ActaLog Implementation Plan: v0.8.0 to v1.0.0

**Created:** 2025-11-22
**Status:** Ready to Execute
**Target Completion:** 12-16 days

---

## Overview

This plan takes ActaLog from v0.7.6-beta to v1.0.0, focusing on:
- PostgreSQL driver migration (breaking changes)
- Complete PR system (time-based + AMRAP)
- Calendar and Timeline views
- Admin metrics and bulk operations
- User engagement features (leaderboards)
- Comprehensive help documentation

---

## Phase 0: Foundation & Infrastructure (v0.8.0-beta)
**Duration:** 2-3 days
**Priority:** CRITICAL - Must complete before other phases

### 1. PostgreSQL Driver Migration (lib/pq → pgx)
**Priority:** HIGH (Breaking Changes Expected)

**Why Now:**
- lib/pq is in maintenance mode (no new features)
- pgx offers better performance and active development
- Breaking change affects all database operations
- Must migrate before adding new features to avoid rework

**Required Information:**
```env
DB_HOST=?
DB_PORT=? (default: 5432)
DB_NAME=?
DB_USER=?
DB_PASSWORD=?

# Performance Tuning
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
```

**Backend Tasks:**
- [ ] Update dependencies in go.mod
  - [ ] Remove `github.com/lib/pq`
  - [ ] Add `github.com/jackc/pgx/v5`
  - [ ] Add `github.com/jackc/pgx/v5/stdlib` (for database/sql compatibility)
- [ ] Update database connection in `internal/repository/database.go`
  - [ ] Replace pq.Open() with pgx connection
  - [ ] Configure connection pooling
  - [ ] Add connection health checks
  - [ ] Update DSN format (pq format → pgx format)
- [ ] Test all repository methods with pgx
- [ ] Update migration runner to use pgx
- [ ] Add environment variable validation
- [ ] Document connection string format change

**Migration Support:**
- [ ] Create migration guide in `docs/MIGRATION_PGX.md`
- [ ] Document breaking changes
- [ ] Provide before/after .env examples
- [ ] Add troubleshooting section

**Testing:**
- [ ] Test with PostgreSQL database
- [ ] Test with SQLite (ensure still works)
- [ ] Test connection pooling under load
- [ ] Test migration rollback if needed

**Success Criteria:**
- All tests pass with pgx driver
- Connection pooling configured and working
- No performance degradation
- Documentation complete

---

### 2. Help Documentation Structure
**Priority:** HIGH (Build now, populate as we develop)

**Goals:**
- Create skeleton for end-user and admin help
- Harvest existing content from .md files
- Set up cross-reference system
- Include screenshot placeholders

#### End-User Help Documentation (docs/help/)

**Directory Structure:**
```
docs/help/
├── README.md (Table of Contents)
├── getting-started/
│   ├── creating-account.md
│   ├── first-workout.md
│   ├── installing-pwa.md
│   └── profile-setup.md
├── logging-workouts/
│   ├── quick-log.md
│   ├── using-templates.md
│   ├── logging-wods.md
│   └── logging-movements.md
├── tracking-progress/
│   ├── personal-records.md
│   ├── performance-charts.md
│   ├── calendar-view.md
│   └── timeline-view.md
├── data-management/
│   ├── importing-data.md
│   ├── exporting-data.md
│   └── wodify-import.md
├── features/
│   ├── leaderboards.md
│   ├── workout-templates.md
│   └── offline-mode.md
├── troubleshooting/
│   ├── common-issues.md
│   ├── sync-problems.md
│   └── login-issues.md
└── faq.md
```

**Screenshot Placeholders Format:**
```markdown
[Screenshot: Dashboard showing recent workouts and PR summary]
- Caption: The main dashboard displays your recent activity
- Location: After logging in, you'll see this view
- Key elements: Recent workouts card, PR summary, Quick Log button
```

**Tasks:**
- [ ] Create directory structure
- [ ] Build Table of Contents (README.md)
- [ ] Create placeholder files for each topic
- [ ] Add screenshot placeholder templates
- [ ] Harvest content from CLAUDE.md, REQUIREMENTS.md, TODO.md
- [ ] Write initial "Getting Started" content
- [ ] Document existing features (Quick Log, Templates, Import/Export)

#### Admin Documentation (docs/admin/)

**Directory Structure:**
```
docs/admin/
├── README.md (Admin Table of Contents)
├── setup/
│   ├── installation.md
│   ├── database-setup.md
│   ├── environment-variables.md
│   └── deployment.md
├── user-management/
│   ├── creating-users.md
│   ├── managing-accounts.md
│   ├── roles-permissions.md
│   └── bulk-operations.md
├── system-operations/
│   ├── database-backups.md
│   ├── restore-procedures.md
│   ├── monitoring.md
│   └── performance-metrics.md
├── security/
│   ├── authentication.md
│   ├── audit-logs.md
│   ├── best-practices.md
│   └── cors-configuration.md
├── data-operations/
│   ├── import-export.md
│   ├── data-migration.md
│   └── cleanup-tools.md
├── troubleshooting/
│   ├── common-admin-issues.md
│   ├── database-problems.md
│   └── performance-tuning.md
└── api-reference/
    ├── admin-endpoints.md
    └── automation-scripts.md
```

**Tasks:**
- [ ] Create directory structure
- [ ] Build Admin Table of Contents
- [ ] Document existing admin features
  - [ ] User management (from v0.7.4)
  - [ ] Database backups (from v0.7.5)
  - [ ] Audit logs
- [ ] Write security best practices
- [ ] Document environment variables (from CLAUDE.md)
- [ ] Create deployment guide
- [ ] Add troubleshooting guides

**Deliverables:**
- Complete help documentation skeleton
- Initial content for existing features
- Screenshot placeholder system
- Cross-reference linking system
- Link from Profile screen to help docs

---

## Phase 1: Time-Based PRs & Core Views (v0.9.0-beta)
**Duration:** 4-5 days
**Priority:** HIGH

### 1. Complete PR System - Time-Based & AMRAP

**Current State:**
- ✅ Weight-based PRs working (movements)
- ❌ Time-based PRs missing (WODs)
- ❌ AMRAP PRs missing (WODs)

**Requirements:**
- Detect fastest time for time-based WODs (e.g., Fran, Murph)
- Detect most rounds+reps for AMRAP WODs
- Auto-flag PRs when logging workouts
- Manual PR toggle support
- Display PR badges in UI

**Backend Tasks:**
- [ ] Update `internal/service/user_workout_service.go`
  - [ ] Add `DetectTimePRs()` method
  - [ ] Add `DetectAMRAPPRs()` method
  - [ ] Integrate into workout logging workflow
- [ ] Update `internal/repository/user_workout_wod_repository.go`
  - [ ] Add `GetBestTimeForWOD(userID, wodID)` method
  - [ ] Add `GetBestRoundsRepsForWOD(userID, wodID)` method
  - [ ] Update PR history queries
- [ ] Add PR comparison logic
  - [ ] Time: Lower seconds wins
  - [ ] AMRAP: Higher rounds+reps wins (rounds * 1000 + reps for comparison)
- [ ] Update PR toggle endpoint to support WODs
- [ ] Add WOD PR statistics to `/api/workouts/prs` endpoint

**Frontend Tasks:**
- [ ] Update WorkoutsView to show WOD PR badges
- [ ] Add WOD PRs to PR History view
- [ ] Update Quick Log to indicate potential PRs
- [ ] Add PR icon (gold trophy) to WOD cards
- [ ] Update Performance view charts with WOD PRs

**Testing:**
- [ ] Test time-based PR detection with various WODs
- [ ] Test AMRAP PR detection with rounds+reps
- [ ] Test PR toggle for WODs
- [ ] Test PR history with mixed movement/WOD PRs

---

### 2. Calendar View

**User Story:** View workout history in calendar format to see patterns and consistency

**Requirements:**
- Full-month calendar display
- Workout indicators on dates (dots/badges)
- Click date to see workouts
- Color coding by workout type
- Navigate between months
- Current day highlight
- Mobile-responsive

**Backend Tasks:**
- [ ] Create `GET /api/workouts/calendar` endpoint
  - [ ] Query params: `year`, `month`
  - [ ] Return workouts grouped by date
  - [ ] Include workout counts per day
  - [ ] Include workout types per day
- [ ] Add calendar data aggregation to service layer

**Frontend Tasks:**
- [ ] Create `web/src/views/CalendarView.vue`
  - [ ] Month calendar grid (7 columns × 5-6 rows)
  - [ ] Workout indicators (colored dots or count badges)
  - [ ] Click date to expand day's workouts
  - [ ] Month navigation (prev/next arrows)
  - [ ] Year selector
  - [ ] Color legend for workout types
- [ ] Add route `/calendar` to router
- [ ] Add navigation link in bottom navigation
- [ ] Mobile optimizations
  - [ ] Touch-friendly date cells
  - [ ] Swipe to change months
  - [ ] Compact view on small screens

**Color Coding:**
- Strength workouts: Blue
- Metcon/WODs: Red/Orange
- Mixed workouts: Purple
- Multiple workouts: Stacked dots or count badge

**Success Criteria:**
- Calendar loads quickly (<500ms)
- All workouts visible in month view
- Date selection shows full workout details
- Responsive on mobile devices

---

### 3. Timeline View

**User Story:** View workouts in chronological order with rich details

**Requirements:**
- Chronological workout feed (newest first)
- Rich workout cards with expandable details
- Infinite scroll or pagination
- Date separators (Today, Yesterday, This Week, etc.)
- Filter by date range
- Placement: Dashboard block if fits, otherwise Profile link

**Implementation Decision:**
- **Try Dashboard first**: Add as collapsible section
- **Fallback**: If dashboard gets too crowded, add link from Profile to `/timeline`

**Backend Tasks:**
- [ ] Update `GET /api/workouts` endpoint
  - [ ] Add `timeline=true` query param
  - [ ] Return enriched data (movement names, WOD names, PR flags)
  - [ ] Add pagination support (`limit`, `offset`)
  - [ ] Default sort: newest first

**Frontend Tasks:**
- [ ] Create Timeline component (`web/src/components/TimelineView.vue`)
  - [ ] Workout card with expandable details
  - [ ] Date separators (intelligent grouping)
  - [ ] PR badges visible
  - [ ] Notes preview
  - [ ] Click to expand full details
- [ ] Add to Dashboard as collapsible section
  - [ ] "Recent Activity" or "Timeline" header
  - [ ] Show last 10 workouts
  - [ ] "View All" button if more exist
- [ ] If dashboard too crowded:
  - [ ] Create standalone `/timeline` route
  - [ ] Add link from Profile screen
  - [ ] Full-featured timeline with filtering

**Features:**
- Date separators: "Today", "Yesterday", "2 days ago", "Last Week", specific dates
- Workout cards show:
  - Date and time
  - Workout type (strength, metcon, mixed)
  - Movements/WODs with key metrics
  - PR badges
  - Notes snippet
  - Expand for full details

**Success Criteria:**
- Timeline loads quickly
- Smooth scrolling performance
- Clear date grouping
- Easy to scan and read

---

### 4. Dashboard Enhancements

**Goal:** Make dashboard the central hub with actionable insights

**New Widgets:**

1. **Quick Stats Summary**
   - This week's workouts
   - This month's workouts
   - This year's workouts
   - Total PRs this month

2. **Recent PRs Widget**
   - Last 5 PRs (movement + WOD)
   - Click to view full PR history
   - "New PR!" badge for recent

3. **Workout Streak**
   - Current streak (consecutive days with workouts)
   - Longest streak (all-time record)
   - Visual streak counter

4. **Quick Log Prominence**
   - Large, centered Quick Log button
   - Teal lightning bolt icon
   - "Log Workout" call-to-action

5. **Timeline Block** (if fits)
   - Last 5-10 workouts
   - Collapsible section
   - "View All" link

**Layout Strategy:**
- Grid layout for stats cards
- Prioritize mobile-first design
- Collapsible sections to reduce clutter
- Scroll if needed (don't cram everything above fold)

**Tasks:**
- [ ] Update `web/src/views/DashboardView.vue`
- [ ] Create reusable stat card component
- [ ] Add streak calculation to backend
- [ ] Update dashboard API to return new data
- [ ] Add skeleton loading states
- [ ] Test on various screen sizes

---

## Phase 2: Admin & Data Management (v0.9.5-beta)
**Duration:** 3-4 days
**Priority:** HIGH

### 1. Application Performance Metrics Dashboard

**Goal:** Give admins visibility into app health and usage

**Metrics to Display:**

**User Metrics:**
- Total users (all time)
- Active users (last 7 days)
- Active users (last 30 days)
- New users this month
- User growth chart (monthly)

**Workout Metrics:**
- Total workouts logged (all time)
- Workouts this month
- Workouts this week
- Average workouts per user
- Workout frequency chart (daily)

**Content Metrics:**
- Top 10 movements logged
- Top 10 WODs logged
- Total PRs flagged
- PR frequency (PRs per day chart)

**System Health:**
- Database size
- Database growth rate
- Backup count and last backup date
- Total uploads (profile images)
- Audit log entries count

**Backend Tasks:**
- [ ] Create `internal/service/metrics_service.go`
  - [ ] `GetUserMetrics()` - user statistics
  - [ ] `GetWorkoutMetrics()` - workout statistics
  - [ ] `GetContentMetrics()` - movements, WODs, PRs
  - [ ] `GetSystemMetrics()` - database, storage, logs
- [ ] Create `internal/handler/metrics_handler.go`
  - [ ] `GET /api/admin/metrics` - all metrics combined
  - [ ] `GET /api/admin/metrics/users` - user stats only
  - [ ] `GET /api/admin/metrics/workouts` - workout stats only
  - [ ] Admin-only authorization
- [ ] Wire up routes in main.go

**Frontend Tasks:**
- [ ] Create `web/src/views/AdminMetricsView.vue`
  - [ ] Stats cards for key numbers
  - [ ] Charts (line charts for growth, bar charts for top items)
  - [ ] Date range selector
  - [ ] Export metrics to CSV button
  - [ ] Auto-refresh option
- [ ] Add route `/admin/metrics` to router
- [ ] Add navigation from AdminView
- [ ] Use Chart.js for visualizations

**Charts:**
- User growth (line chart, monthly)
- Workout frequency (line chart, daily/weekly)
- Top movements (bar chart)
- Top WODs (bar chart)
- PR frequency (line chart)

**Export:**
- Export all metrics to CSV
- Include timestamp and date range

---

### 2. User Data Import/Export (Bulk Operations)

**Goal:** Enable bulk user management for onboarding and migrations

**Requirements:**
- Export all users to CSV
- Import users from CSV (bulk creation)
- Preview before import
- Auto-generate temporary passwords
- Send welcome emails
- Admin-only feature

**CSV Schema:**
```csv
email,name,role,email_verified,send_welcome_email
user1@example.com,John Doe,user,true,true
user2@example.com,Jane Smith,admin,false,false
```

**Backend Tasks:**
- [ ] Update `internal/service/export_service.go`
  - [ ] `ExportUsersToCSV(adminUserID)` method
  - [ ] Include all user fields except password_hash
  - [ ] Include timestamps (created_at, last_login_at, etc.)
- [ ] Update `internal/service/import_service.go`
  - [ ] `ImportUsersFromCSV(adminUserID, csvData, options)` method
  - [ ] Preview mode (validate without creating)
  - [ ] Confirm mode (create users)
  - [ ] Auto-generate passwords (16 char random)
  - [ ] Send welcome emails with password reset links
  - [ ] Duplicate detection (by email)
  - [ ] Validation (email format, role enum)
- [ ] Update handlers for user import/export
- [ ] Add audit logging for bulk operations

**Frontend Tasks:**
- [ ] Update `web/src/views/ExportView.vue`
  - [ ] Add "Users" export option (admin only)
- [ ] Update `web/src/views/ImportView.vue`
  - [ ] Add "Users" import option (admin only)
  - [ ] Preview table showing users to be created
  - [ ] Options: send_welcome_email checkbox
  - [ ] Success summary: users created, emails sent
- [ ] Admin-only route guards

**Email Template:**
```
Subject: Welcome to ActaLog

Hi {name},

An administrator has created an account for you on ActaLog.

Email: {email}

Please reset your password using this link:
{reset_link}

This link expires in 1 hour.

Welcome to ActaLog!
```

---

### 3. User Activity Monitoring

**Goal:** Track user engagement and identify inactive users

**Metrics per User:**
- Last login date
- Total workouts logged
- Workouts this month
- Last workout date
- Days since last activity
- Total PRs
- Account status

**Backend Tasks:**
- [ ] Create activity monitoring queries
- [ ] Add `GET /api/admin/activity` endpoint
  - [ ] Return user activity stats
  - [ ] Filter by active/inactive
  - [ ] Sort by last activity
  - [ ] Pagination support
- [ ] Add "inactive users" detector (e.g., no activity in 30+ days)

**Frontend Tasks:**
- [ ] Create `web/src/views/AdminActivityView.vue`
  - [ ] User activity table
  - [ ] Filter: active (last 7 days), inactive (30+ days), all
  - [ ] Sort by: last login, last workout, total workouts
  - [ ] Search by email/name
  - [ ] Click user to see details
- [ ] Add route `/admin/activity`
- [ ] Add navigation from AdminView

**Features:**
- Identify inactive users for engagement campaigns
- Monitor new user onboarding (are they logging workouts?)
- Track power users (high workout frequency)

---

## Phase 3: User Engagement Features (v1.0.0-beta)
**Duration:** 3-4 days
**Priority:** MEDIUM

### 1. Leaderboard System

**Goal:** Drive competition and engagement through friendly leaderboards

**Requirements:**
- Per-WOD leaderboards
- Division support (Rx, Scaled, Beginner)
- No email verification required to appear
- Filter by date range (optional)
- User rank display
- Admin verification for top entries (future)

**Backend Tasks:**
- [ ] Add `division` field to `user_workout_wods` table
  - [ ] Migration: Add division ENUM ('rx', 'scaled', 'beginner')
  - [ ] Default to 'scaled' if NULL
- [ ] Create leaderboard queries
  - [ ] Time-based WODs: ORDER BY time_seconds ASC
  - [ ] AMRAP WODs: ORDER BY (rounds * 1000 + reps) DESC
- [ ] Create `GET /api/leaderboards/wods/{id}` endpoint
  - [ ] Query params: division, date_from, date_to, limit
  - [ ] Return ranked list with user info
  - [ ] Include current user's rank
- [ ] Add division field to workout logging endpoints

**Frontend Tasks:**
- [ ] Update Quick Log to include division selector
- [ ] Update workout logging forms with division
- [ ] Create `web/src/components/LeaderboardCard.vue`
  - [ ] Show top 10 entries
  - [ ] Division tabs (Rx, Scaled, Beginner)
  - [ ] User's rank highlighted
  - [ ] Date range filter
- [ ] Add leaderboard to WOD detail pages
- [ ] Create standalone `/leaderboards` view
  - [ ] List of WODs with leaderboard links
  - [ ] Filter by WOD type

**Leaderboard Display:**
```
Fran - Rx Division
┌─────┬──────────────────┬──────────┬──────────┐
│ Rank│ User             │ Score    │ Date     │
├─────┼──────────────────┼──────────┼──────────┤
│  1  │ John D.          │ 2:45     │ Nov 20   │
│  2  │ Jane S.          │ 3:12     │ Nov 18   │
│  3  │ You ⭐           │ 3:45     │ Nov 22   │
└─────┴──────────────────┴──────────┴──────────┘
```

**Privacy:**
- User names displayed (not emails)
- Users can see all leaderboards regardless of verification status
- No opt-out for now (future feature)

---

### 2. Additional Dashboard Metrics

**Goal:** Provide insights into consistency and progress

**New Metrics:**

1. **Workout Frequency Chart**
   - Bar/line chart showing workouts per week
   - Last 12 weeks
   - Average workouts per week indicator

2. **Consistency Indicators**
   - Longest streak (all-time)
   - Current streak
   - Streaks this month
   - Average days between workouts

3. **Month-to-Month Comparison**
   - This month vs last month
   - Workout count change (+/- %)
   - PR count change
   - New movements tried

4. **Year-to-Year Comparison**
   - This year vs last year
   - Total workouts comparison
   - PR comparison
   - Most improved movement

5. **Volume Trends**
   - Total weight lifted (all movements)
   - Chart showing volume over time
   - Month-by-month breakdown

**Backend Tasks:**
- [ ] Create aggregation queries for new metrics
- [ ] Add to dashboard endpoint
- [ ] Optimize queries for performance

**Frontend Tasks:**
- [ ] Update DashboardView with new widgets
- [ ] Create charts for visualizations
- [ ] Add comparison displays
- [ ] Ensure mobile responsive

---

### 3. Workout History Advanced Filters

**Goal:** Help users find specific workouts quickly

**Filter Options:**
- By movement (dropdown or autocomplete)
- By WOD (dropdown or autocomplete)
- By workout type (strength, metcon, mixed)
- By date range (date pickers)
- By PR status (show only PRs)
- By notes (text search)

**Backend Tasks:**
- [ ] Update `GET /api/workouts` endpoint
  - [ ] Add filter query params
  - [ ] Build dynamic WHERE clause
  - [ ] Maintain pagination support
- [ ] Optimize filter queries with indexes

**Frontend Tasks:**
- [ ] Update `web/src/views/WorkoutsView.vue`
  - [ ] Add filter panel (collapsible)
  - [ ] Movement autocomplete
  - [ ] WOD autocomplete
  - [ ] Type checkboxes
  - [ ] Date range pickers
  - [ ] PR filter toggle
  - [ ] "Clear Filters" button
- [ ] Save filter preferences in localStorage
- [ ] Display active filters as chips

**UX:**
- Filters apply immediately (no "Apply" button)
- Active filters shown as removable chips
- Filter panel collapsible to save space
- Mobile-friendly filter drawer

---

## Documentation Updates Throughout

As we implement each phase, we'll populate help docs:

**Phase 0:**
- PostgreSQL migration guide
- Updated deployment docs

**Phase 1:**
- Calendar view user guide
- Timeline view user guide
- PR tracking documentation (complete)
- Dashboard overview

**Phase 2:**
- Admin metrics guide
- User import/export guide
- Activity monitoring guide

**Phase 3:**
- Leaderboard user guide
- Advanced filtering guide
- Analytics interpretation guide

---

## Success Criteria for v1.0.0

### Functional Requirements:
- ✅ All PR types working (weight, time, AMRAP)
- ✅ Calendar and timeline views functional
- ✅ Admin metrics dashboard operational
- ✅ User bulk operations working
- ✅ Leaderboards displaying correctly
- ✅ PostgreSQL driver migrated successfully

### Performance Requirements:
- Dashboard loads in <1 second
- Calendar view <500ms
- Leaderboards <1 second
- No noticeable performance degradation from lib/pq → pgx

### Documentation Requirements:
- Complete end-user help docs
- Complete admin help docs
- All screenshots placeholders in place
- Migration guides complete

### User Experience:
- Mobile-responsive on all new views
- Intuitive navigation
- Clear visual hierarchy
- Helpful error messages

---

## Ready to Execute

**Next Immediate Steps:**
1. ✅ Plan approved
2. ⏳ Receive PostgreSQL connection details
3. ⏳ Start Phase 0 (PostgreSQL migration + Help docs)

**PostgreSQL Connection Details Needed:**
```env
DB_HOST=?
DB_PORT=?
DB_NAME=?
DB_USER=?
DB_PASSWORD=?
```

**Questions Answered:**
- ✅ Screenshot placeholders: YES, include now
- ✅ Timeline placement: Dashboard block if fits, else Profile link
- ✅ Email verification for leaderboards: NO, not required
- ✅ Admin features priority: Metrics and bulk ops are high priority
- ✅ Test coverage: MEDIUM priority (defer for now)
- ✅ Help docs: Build skeleton now, populate as we develop

---

**Let's build ActaLog v1.0! 🚀**
