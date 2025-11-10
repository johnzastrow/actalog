# Changelog

All notable changes to ActaLog will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
