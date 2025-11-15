# TODO

This document tracks planned features, improvements, and known issues for ActaLog.

## Current Status (v0.4.5-beta)

### Completed Features
- [x] User authentication (registration, login, JWT)
- [x] Workout CRUD operations
- [x] Movement management with 31 seeded CrossFit movements
- [x] Workout logging functionality
- [x] Workouts list with movement details
- [x] Dashboard with live statistics
- [x] Searchable movement selection (autocomplete)
- [x] Modern mobile-first UI design
- [x] Bottom navigation
- [x] Profile view with logout
- [x] Clean Architecture implementation
- [x] API authentication middleware
- [x] Database seeding for movements
- [x] Responsive scrolling on all views
- [x] Personal Records (PR) tracking system
- [x] Password reset system with email delivery
- [x] Email verification system
- [x] WOD (Workout of the Day) management system
- [x] Workout template system
- [x] Pinia state management (stores for WODs and templates)
- [x] WOD Library view with filtering
- [x] Admin data cleanup tools (detect/fix WOD score_type mismatches)
- [x] WOD score_type constraint enforcement across all forms
- [x] HH:MM:SS time format support for Time-based WODs
- [x] Admin edit functionality for mismatched WOD records

## High Priority

### Features
- [ ] Workout detail view (individual workout page) - **Partially implemented, needs enhancement**
- [ ] Edit workout functionality - **Partially implemented, needs enhancement**
- [ ] Delete workout with confirmation
- [ ] Performance charts for movement progress over time
- [ ] Add custom movements from the UI
- [ ] Workout history filtering by date range
- [ ] Search workouts by movement or date
- [ ] Template Library browsing view
- [ ] Template-based workout logging integration

### UI/UX Improvements
- [ ] Loading states for all API calls
- [ ] Success notifications after save operations
- [ ] Error boundary for better error handling
- [ ] Pull-to-refresh on mobile
- [ ] Skeleton loaders for better perceived performance
- [ ] Improve time input UX (consider time picker component)

### Backend
- [ ] Pagination for workout lists
- [ ] Workout search and filtering endpoints
- [x] Movement statistics endpoint (PR tracking) - **Implemented in v0.3.0**
- [ ] User preferences/settings storage
- [ ] Data export functionality (CSV, JSON)
- [x] Enable support for MariaDB - **Implemented in CI/CD**

## Medium Priority

### Features
- [x] Multiple movements per workout - **Implemented**
- [x] Timed workouts (AMRAP, EMOM, For Time) - **Implemented via WOD regimes**
- [x] Workout templates for common WODs - **Implemented in v0.4.0**
- [x] Personal records (PR) tracking and display - **Implemented in v0.3.0**
- [ ] Workout sharing between users
- [x] Comments/notes on specific movements - **Implemented via notes field**
- [ ] Photo upload for form checks
- [ ] Rest timer integration
- [ ] Bulk data cleanup tools for admins (additional cleanup operations)

### Data & Analytics
- [ ] Weekly/monthly workout summaries
- [ ] Volume tracking (total weight lifted)
- [ ] Movement frequency analysis
- [ ] Progress streak tracking
- [ ] Goal setting and tracking

### UI/UX
- [ ] Dark mode support
- [ ] Workout calendar view
- [ ] Movement video demonstrations
- [ ] Onboarding tutorial for new users
- [ ] Settings page (units, preferences)

## Low Priority

### Features
- [ ] Social features (friends, leaderboards)
- [ ] Gym/box integration
- [ ] Coach accounts with athlete management
- [ ] Workout planning and scheduling
- [ ] Integration with wearable devices
- [ ] Mobile app (React Native or Flutter)
- [ ] Offline support with sync

### Backend
- [ ] Redis caching layer
- [ ] Background jobs for data aggregation
- [x] Email notifications - **Implemented for password reset and email verification**
- [ ] API rate limiting
- [ ] GraphQL API option
- [ ] Multi-tenant support for gyms
- [x] Admin-only endpoints with role-based access control - **Implemented in v0.4.4**

### DevOps
- [x] Docker Compose for development - **Implemented**
- [x] CI/CD pipeline (GitHub Actions) - **Implemented with multi-database testing**
- [x] Automated testing (unit, integration) - **Implemented, e2e pending**
- [x] Database migration system - **Implemented**
- [ ] Monitoring and logging (Prometheus, Grafana)
- [ ] Production deployment guide

## Continuous Integration

A GitHub Actions workflow is present at `.github/workflows/ci.yml`. It performs linting, unit tests, an integration test matrix (sqlite3, Postgres, MariaDB using Actions service containers), and a frontend build.

Notes:

- The CI integration jobs pass `-db` and `-dsn` to the integration tests so they can run against different database backends.
- The Go test job sets `CGO_ENABLED=1` when running sqlite3-based tests to support the `github.com/mattn/go-sqlite3` driver. If you want to avoid CGO, consider switching to a pure-Go SQLite driver or running only Postgres/MariaDB in CI.

Run tests locally:

```bash
# Run all tests
go test ./... -v

# Run integration tests (default: sqlite in-memory)
go test ./test/integration -run Test -v
```




## Known Issues

### Bugs
- [x] Date picker may show timezone offset issues - **Fixed with local date formatting**
- [ ] Console warning about iterable (third-party library)
- [ ] Movement type filter not yet implemented

### Technical Debt
- [ ] Add comprehensive test coverage (improving, needs more coverage)
- [ ] Implement proper error logging system
- [ ] Add API request/response validation with schemas
- [x] Implement database migrations instead of manual schema - **Implemented**
- [ ] Add API versioning strategy
- [ ] Refactor duplicate bottom navigation code into component
- [ ] Consider time picker component for better UX (currently using 3 separate number inputs)

## Future Considerations

### Architecture
- [ ] Microservices architecture for scaling
- [ ] Event-driven architecture for analytics
- [ ] CQRS pattern for complex queries
- [ ] WebSocket support for real-time features

### Business
- [ ] Subscription/pricing model
- [ ] White-label solution for gyms
- [ ] Mobile app monetization strategy
- [ ] API for third-party integrations

---

**Last Updated**: 2025-11-14
**Version**: 0.4.5-beta
