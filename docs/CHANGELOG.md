# Changelog

All notable changes to ActaLog will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added - Design Refinements v0.3.0
**Refined design decisions documented** through user consultation:

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

### Added - Database Schema v0.3.0
- **Major schema redesign** based on logical data model requirements
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

### Changed - Database Schema v0.3.0
- **Workouts** are now reusable templates (not user-specific instances)
- Renamed `movements` table to `strength_movements`
- Added `movement_type` to strength_movements (weightlifting, cardio, gymnastics)
- Renamed `workout_movements` to `workout_strength`
- Removed user-specific fields from workouts table (user_id, workout_date, workout_type)
- Updated ERD to reflect many-to-many relationships properly

### Migration Required
- Database migration from v0.2.0 to v0.3.0 needed
- See DATABASE_SCHEMA.md for migration steps
- Backend domain models need updates
- API endpoints need refactoring for new structure

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
- Implement database migration scripts
- Update backend for new schema
- Seed data for standard WODs and movements
- Workout logging functionality (updated for new schema)
- Progress tracking with charts and graphs
- Data import/export (CSV/JSON)
- Push notifications for workout reminders
- Web Share API integration
- Implement all 33 screens defined in screen inventory:
  - Management screens for WODs (List, Create, Edit with CRUD operations)
  - Management screens for Strength Movements (List, Create, Edit with CRUD operations)
  - Management screens for Workout Templates (List, Create, Edit with CRUD operations)
  - Import/Export data screens
  - Settings Menu flyout implementation

## [0.2.0] - 2025-11-08

### Added - PWA Implementation

**Progressive Web App Features**:
- 🎯 Full PWA support with installable app capability
- 📱 Works on iOS, Android, and desktop browsers
- ⚡ Service worker with Workbox for offline functionality
- 💾 IndexedDB for offline data storage and synchronization
- 🔄 Background sync for pending workout operations
- 🚀 Automatic updates with user notification
- 📦 Precaching of static assets for instant loading
- 🎨 Web app manifest with ActaLog branding

**PWA Infrastructure**:
- Added `vite-plugin-pwa` version 0.20.5
- Added `workbox-window` version 7.3.0 for service worker management
- Configured service worker with intelligent caching strategies:
  - CacheFirst for fonts and static assets
  - NetworkFirst for API data with 5-minute fallback
  - Precaching for all build artifacts
- PWA development mode enabled in vite.config.js

**Offline Functionality**:
- Created `offlineStorage.js` utility for IndexedDB operations
- Object stores for workouts, movements, and pending sync queue
- Automatic sync when connection is restored
- Network status detection and handling
- Queue system for offline workout creation

**Configuration**:
- Web app manifest with:
  - Name: "ActaLog - CrossFit Workout Tracker"
  - Theme color: #2c3657 (primary brand color)
  - Background color: #ffffff
  - Standalone display mode
  - Portrait orientation
  - 8 icon sizes (72px to 512px)
- Service worker registration in main.js
- Update notification system for new versions

**Documentation**:
- Created comprehensive DEPLOYMENT.md with:
  - HTTPS setup requirements and SSL certificate guide
  - Traditional server deployment steps
  - Docker deployment option
  - Nginx reverse proxy configuration for PWA
  - PWA-specific deployment checklist
  - Troubleshooting guide
- Updated REQUIREMENTS.md with PWA capabilities
- Updated ARCHITECTURE.md with PWA architecture diagrams
- Updated SETUP.md with PWA development and testing instructions
- Created web/public/icons/README.md with icon generation guide

**Developer Experience**:
- PWA testing enabled in development mode (works on localhost)
- Service worker updates on reload during development
- DevTools integration for PWA debugging
- Lighthouse PWA audit support

### Changed
- Updated package.json version to 0.2.0
- Enhanced vite.config.js with complete PWA configuration
- Modified main.js to register service worker

### Technical Details

**Caching Strategies**:
- Google Fonts: 1-year cache (CacheFirst)
- CDN assets: 1-year cache (CacheFirst)
- API responses: 5-minute cache with network priority (NetworkFirst)
- Static assets: Precached and updated on deployment

**Browser Support**:
- Chrome/Edge: Full support (Desktop & Mobile)
- Safari: Full support iOS 11.3+
- Firefox: Full support (Desktop & Android)
- Samsung Internet: Full support
- Opera: Full support

**Performance**:
- Offline-first architecture for instant loading
- Background sync for seamless data synchronization
- Code splitting and lazy loading ready
- Optimized asset caching

### Security
- HTTPS required in production for PWA features
- Service worker scope limited to application
- Same-origin policy enforced
- Secure service worker update mechanism

### Notes for Deployment
- ⚠️ **HTTPS is required** for PWA features to work in production
- 📦 Generate PWA icons before deployment (see web/public/icons/README.md)
- 🔍 Run Lighthouse PWA audit (target score: 90+)
- ✅ Test offline functionality after deployment
- 📱 Verify "Add to Home Screen" works on mobile devices

## [0.1.0-alpha] - 2025-11-07

### Added
- Initial project structure with Clean Architecture
- Go backend with Chi router
- Vue.js 3 frontend with Vuetify 3
- User registration and login system
- JWT-based authentication
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

**Current Version:** 0.3.0-dev (schema updated, migration pending)
**Last Updated:** 2025-11-09
