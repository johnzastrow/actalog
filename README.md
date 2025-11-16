# ActaLog

> A mobile-first fitness tracker for CrossFit enthusiasts to log workouts, track progress, and analyze performance.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Vue.js](https://img.shields.io/badge/Vue.js-3.x-4FC08D?style=flat&logo=vue.js)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/johnzastrow/actalog/actions/workflows/ci.yml/badge.svg)](https://github.com/johnzastrow/actalog/actions/workflows/ci.yml)

## Overview

ActaLog is an open source fitness tracker app made for individuals or small groups. It is capable of logging CrossFit (or functional fitness) workouts and tracking history for weights, reps, and named weightlifting lifts, but can be used to track other types of workouts as well. Built with a Go backend and Vue.js/Vuetify frontend, it provides a clean, mobile-first interface for tracking your fitness journey.

**Version:** 0.4.3-beta

## Features

### Current Features (v0.4.2-beta)

- ✅ **User Authentication**: Secure registration and login with JWT tokens
- ✅ **Email Verification**: Verify email addresses with secure tokens and automated emails
- ✅ **Password Reset**: Complete forgot password flow with email delivery
- ✅ **Workout Logging**: Track workouts with movements, weights, sets, and reps
- ✅ **Movement Database**: 31 pre-seeded standard CrossFit movements
- ✅ **Searchable Movements**: Autocomplete search for quick movement selection
- ✅ **Workout History**: View all logged workouts with movement details
- ✅ **Personal Records (PR) Tracking**: Automatic PR detection and gold trophy badges
- ✅ **PR History Page**: Dedicated view showing recent PRs and all-time records
- ✅ **Dashboard**: Real-time statistics showing total and monthly workout counts
- ✅ **Recent Activity**: Quick view of your last 5 workouts with PR indicators
- ✅ **Mobile-First Design**: Responsive UI optimized for mobile devices
- ✅ **Modern UI**: Clean design with cyan accents and dark navy headers
- ✅ **Rx/Scaled Tracking**: Mark movements as Rx or Scaled
- ✅ **Workout Notes**: Add personal notes to each workout
- ✅ **Secure API**: Protected endpoints with JWT authentication
- 🔒 **Security**: bcrypt password hashing, parameterized SQL queries

### Coming Soon

- 📊 **Performance Charts**: Visual progress tracking for movements over time
- ✏️ **Edit Workouts**: Modify existing workout entries
- 🗑️ **Delete Workouts**: Remove workouts with confirmation
- ➕ **Custom Movements**: Add your own movements from the UI
- 🔍 **Workout Filtering**: Search and filter by date, movement, or type
- 📤 **Data Export**: Download your workout data (CSV, JSON)

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Router**: Chi
- **Database**: SQLite (dev), PostgreSQL (prod), MariaDB (supported)
- **Authentication**: JWT with golang-jwt/jwt
- **ORM**: sqlx
- **Testing**: testify

### Frontend
- **Framework**: Vue.js 3
- **UI Library**: Vuetify 3
- **State Management**: Pinia
- **Build Tool**: Vite
- **Charts**: Chart.js with vue-chartjs

### Infrastructure
- **Containerization**: Docker + Docker Compose
- **Database Migrations**: golang-migrate
- **Reverse Proxy**: Nginx (optional)

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Node.js 18+ and npm
- Docker and Docker Compose (optional)

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/johnzastrow/actalog.git
   cd actalog
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Install Go dependencies**
   ```bash
   go mod download
   ```

4. **Install frontend dependencies**
   ```bash
   cd web
   npm install
   cd ..
   ```

5. **Run the backend**
   ```bash
   # Terminal 1
   make run
   # Or: go run cmd/actalog/main.go
   ```

6. **Run the frontend**
   ```bash
   # Terminal 2
   cd web
   npm run dev
   ```

7. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - API Health: http://localhost:8080/health

### Using Docker

```bash
# Start all services
make docker-up

# Stop all services
make docker-down

# View logs
make docker-logs
```

## Project Structure

```
actalog/
├── cmd/actalog/          # Application entry point
├── internal/             # Private application code
│   ├── domain/          # Business entities and interfaces
│   ├── repository/      # Data access layer
│   ├── service/         # Business logic layer
│   └── handler/         # HTTP handlers
├── pkg/                 # Public packages
│   ├── auth/           # Authentication utilities
│   ├── middleware/     # HTTP middleware
│   ├── utils/          # Helper functions
│   └── version/        # Version management
├── api/                 # API definitions
├── configs/            # Configuration
├── test/               # Tests
├── web/                # Frontend Vue.js app
├── docs/               # Documentation
├── design/             # Design assets
└── migrations/         # Database migrations
```

## Available Commands

### Backend (Makefile)

```bash
make help              # Show all available commands
make build             # Build the application
make run               # Run the application
make test              # Run all tests with coverage
make test-unit         # Run unit tests only
make lint              # Run linters
make fmt               # Format code
make clean             # Clean build artifacts
make install-tools     # Install development tools
```

### Frontend

```bash
npm run dev            # Start development server
npm run build          # Build for production
npm run preview        # Preview production build
npm run lint           # Run ESLint
npm run format         # Format code with Prettier
```

## Documentation

Comprehensive documentation is available in the `docs/` directory:

- [Architecture](docs/ARCHITECTURE.md) - System architecture and design patterns
- [Database Schema](docs/DATABASE_SCHEMA.md) - Database structure and ERD
- [Database Support](docs/DATABASE_SUPPORT.md) - Multi-database setup (SQLite, PostgreSQL, MySQL/MariaDB)
- [Logging Guide](docs/LOGGING.md) - Logging configuration and best practices
- [Requirements](docs/REQUIIREMENTS.md) - Project requirements and user stories
- [AI Instructions](docs/AI_INSTRUCTIONS.md) - Development guidelines

## Configuration

Configuration is managed through environment variables. See [.env.example](.env.example) for all available options.

Key configuration:
- `APP_ENV`: Environment (development, staging, production)
- `DB_DRIVER`: Database driver (sqlite, postgres, mysql)
- `JWT_SECRET`: Secret key for JWT tokens (MUST change in production!)
- `SERVER_PORT`: Server port (default: 8080)

## Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# View coverage report
make coverage
```

## CI and Integration Tests

We run CI using GitHub Actions. The primary workflow is `.github/workflows/ci.yml` and performs linting, unit tests, integration tests (matrix: sqlite3, postgres, mariadb), and a frontend build.

Integration tests accept flags and environment variables:

- Flag `-db` (default: `sqlite3`) — driver name passed to tests
- Flag `-dsn` (default: `:memory:`) — DSN used by repository.InitDatabase
- Environment variables `DB_DRIVER` and `DB_DSN` can also be used to override flags in CI or local runs.

Examples:

```bash
# Run integration tests against in-memory SQLite (default)
go test ./test/integration -run Test -v

# Run against a local Postgres container
docker run -d --name actalog-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=actalog_test -p 5432:5432 postgres:15
go test ./test/integration -run Test -v -args -db=postgres -dsn="host=127.0.0.1 port=5432 user=postgres password=postgres dbname=actalog_test sslmode=disable"

# Run against a local MariaDB container
docker run -d --name actalog-mariadb -e MYSQL_ROOT_PASSWORD=example -e MYSQL_DATABASE=actalog_test -p 3306:3306 mariadb:10.11
go test ./test/integration -run Test -v -args -db=mysql -dsn="root:example@tcp(127.0.0.1:3306)/actalog_test?parseTime=true&multiStatements=true"
```

Local CI note

If you want to run the same checks locally as CI does, run the unit tests, linters and the web build:

```bash
# run linters and unit tests
make lint
make test

# build frontend
cd web && npm run build
```


## Security

- **Passwords**: Hashed with bcrypt (cost factor 12+)
- **Authentication**: JWT with secure secret keys
- **SQL Injection**: Parameterized queries only
- **CORS**: Configurable allowed origins
- **TLS/SSL**: Required in production

⚠️ **Important**: Change `JWT_SECRET` before deploying to production!

## Contributing

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for development guidelines.

1. Follow Clean Architecture principles
2. Write tests for new features
3. Run linters before committing
4. Follow Go and Vue.js best practices

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or feature requests, please open an issue on GitHub.

## Roadmap

### Completed (v0.3.1-beta)
- [x] User authentication and registration
- [x] Email verification system
- [x] Password reset functionality (forgot password flow)
- [x] Workout logging functionality
- [x] Movement database with 31 standard CrossFit movements
- [x] Workout history viewing
- [x] Dashboard with statistics
- [x] Mobile-responsive design
- [x] Searchable movement selection
- [x] Personal records (PR) tracking with auto-detection
- [x] PR history page with all-time records

### In Progress
- [ ] Performance tracking with charts
- [ ] Edit/delete workout functionality
- [ ] Custom movement creation
- [ ] Workout filtering and search

### Planned
- [ ] Data import/export (CSV, JSON) with PR flags
- [ ] Workout templates for common WODs
- [ ] Timed workouts (AMRAP, EMOM, For Time)
- [ ] PWA support for offline access
- [ ] Dark mode
- [ ] Profile management and settings
- [ ] Mobile apps (iOS/Android)
- [ ] Social features and leaderboards

---

