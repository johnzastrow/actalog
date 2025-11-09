# Changelog

All notable changes to ActaLog will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Multi-database support**: SQLite, PostgreSQL, and MySQL/MariaDB
- **Database migration system** with version tracking and rollback support
- **Workout logging functionality** with complete CRUD operations
- **Movement database** with 82 standard CrossFit movements (auto-seeded)
- **Progress tracking** by movement for PR analysis
- Database-agnostic DSN builder
- Driver-specific schema generation (SQLite, PostgreSQL, MySQL)
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
- Comprehensive DATABASE_SUPPORT.md documentation

### Changed
- Updated migration system to accept driver parameter for database-agnostic migrations
- Improved table existence checking across all database types
- Enhanced schema creation with database-specific SQL dialects

### Planned
- Data import/export (CSV/JSON)
- PR (Personal Record) tracking with automatic detection
- Workout templates and named WOD database
- Charts and graphs for progress visualization

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

**Current Version:** 0.1.0-alpha
**Last Updated:** 2025-11-07
