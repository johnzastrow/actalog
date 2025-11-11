# TODO

This document tracks planned features, improvements, and known issues for ActaLog.

## Current Status (v0.2.0-beta)

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

## High Priority

### Features
- [ ] Workout detail view (individual workout page)
- [ ] Edit workout functionality
- [ ] Delete workout with confirmation
- [ ] Performance charts for movement progress over time
- [ ] Add custom movements from the UI
- [ ] Workout history filtering by date range
- [ ] Search workouts by movement or date

### UI/UX Improvements
- [ ] Loading states for all API calls
- [ ] Success notifications after save operations
- [ ] Error boundary for better error handling
- [ ] Pull-to-refresh on mobile
- [ ] Skeleton loaders for better perceived performance

### Backend
- [ ] Pagination for workout lists
- [ ] Workout search and filtering endpoints
- [ ] Movement statistics endpoint (PR tracking)
- [ ] User preferences/settings storage
- [ ] Data export functionality (CSV, JSON)
- [ ] Enable support for Mariadb

## Medium Priority

### Features
- [ ] Multiple movements per workout (currently supports one)
- [ ] Timed workouts (AMRAP, EMOM, For Time)
- [ ] Workout templates for common WODs
- [ ] Personal records (PR) tracking and display
- [ ] Workout sharing between users
- [ ] Comments/notes on specific movements
- [ ] Photo upload for form checks
- [ ] Rest timer integration

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
- [ ] Email notifications
- [ ] API rate limiting
- [ ] GraphQL API option
- [ ] Multi-tenant support for gyms

### DevOps
- [ ] Docker Compose for development
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Automated testing (unit, integration, e2e)
- [ ] Database migration system
- [ ] Monitoring and logging (Prometheus, Grafana)
- [ ] Production deployment guide

## Continuous Integration

- A GitHub Actions workflow has been added at `.github/workflows/ci.yml` to run:
	- Go vet and golangci-lint
	- Go unit and integration tests (with CGO enabled for sqlite3 where necessary)
	- Web build for the `web/` frontend (npm install and build)

Usage notes:

- The workflow runs on push and pull_request against `main`.
- The Go job enables `CGO_ENABLED=1` when running `go test ./...` to support the `github.com/mattn/go-sqlite3` driver used by integration tests. If you want CI to avoid CGO, replace sqlite-based integration tests or use a different driver.

To run tests locally from the repository root:

```bash
go test ./... -v
```

Integration test matrix (CI)

- The GitHub Actions CI runs integration tests in a matrix over three databases: sqlite3 (in-memory), Postgres, and MariaDB.

- In CI the Postgres and MariaDB jobs use Actions service containers. The test job runs the integration package with the following test flags passed via `-args`:
	- `-db` — the database driver name (`sqlite3`, `postgres`, `mysql`)
	- `-dsn` — the DSN/connection string for the target DB

Local examples

- Run integration tests against SQLite (default):

```bash
go test ./test/integration -run Test -v
```

- Run integration tests against a local Postgres container:

```bash
docker run -d --name actalog-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=actalog_test -p 5432:5432 postgres:15
# wait for pg to be ready, then:
go test ./test/integration -run Test -v -args -db=postgres -dsn="host=127.0.0.1 port=5432 user=postgres password=postgres dbname=actalog_test sslmode=disable"
docker rm -f actalog-postgres
```

- Run integration tests against a local MariaDB container:

```bash
docker run -d --name actalog-mariadb -e MYSQL_ROOT_PASSWORD=example -e MYSQL_DATABASE=actalog_test -p 3306:3306 mariadb:10.11
# wait for mysql to be ready, then:
go test ./test/integration -run Test -v -args -db=mysql -dsn="root:example@tcp(127.0.0.1:3306)/actalog_test?parseTime=true&multiStatements=true"
docker rm -f actalog-mariadb
```




## Known Issues

### Bugs
- [ ] Date picker may show timezone offset issues
- [ ] Console warning about iterable (third-party library)
- [ ] Movement type filter not yet implemented

### Technical Debt
- [ ] Add comprehensive test coverage (currently minimal)
- [ ] Implement proper error logging system
- [ ] Add API request/response validation with schemas
- [ ] Implement database migrations instead of manual schema
- [ ] Add API versioning strategy
- [ ] Refactor duplicate bottom navigation code into component

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

**Last Updated**: 2025-11-06
**Version**: 0.2.0-beta
