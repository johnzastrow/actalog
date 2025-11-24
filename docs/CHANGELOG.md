# Changelog

All notable changes to ActaLog will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

**Current Version:** 0.10.0-beta
**Schema Version:** v0.10.0 (Docker deployment with seed data)
**Last Updated:** 2025-01-23
