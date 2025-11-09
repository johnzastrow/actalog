# Changelog

All notable changes to ActaLog will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- New Dashboard UI matching design specifications
- Calendar component showing monthly workout activity
- Recent workouts cards with grouped display
- Top app bar with ActaLog logo and current date
- Unified bottom navigation across all authenticated views
- Avatar support for user profile icon
- Workout badge for Personal Records (PRs)

### Changed
- Complete Dashboard redesign with calendar view
- Moved header and bottom navigation to App.vue for consistency
- Updated color scheme to match brand guidelines
- Improved mobile-first responsive design
- Enhanced bottom navigation with better iconography

### Planned
- Workout logging functionality
- Movement database with standard CrossFit movements
- Progress tracking with charts and graphs
- Data import/export (CSV/JSON)
- Push notifications for workout reminders
- Web Share API integration

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

**Current Version:** 0.2.0
**Last Updated:** 2025-11-08
