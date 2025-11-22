# Changelog

All notable changes to ActaLog will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

**Current Version:** 0.4.2-beta
**Schema Version:** 0.1.0 (v0.3.0 schema designed, migration not yet implemented)
**Last Updated:** 2025-01-14
