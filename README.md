# ActaLog

> A mobile-first fitness tracker for CrossFit enthusiasts to log workouts, track progress, and analyze performance.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Vue.js](https://img.shields.io/badge/Vue.js-3.x-4FC08D?style=flat&logo=vue.js)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/johnzastrow/actalog/actions/workflows/ci.yml/badge.svg)](https://github.com/johnzastrow/actalog/actions/workflows/ci.yml)

## Roadmap (current)

This roadmap is synchronized with the repository `CHANGELOG.md` and `TODO.md`. Recent releases (v0.4.x-beta) delivered significant features; below is the current status grouped by priority.

### Completed (high level)
- Authentication: registration, login, JWT, password reset, email verification (v0.3.1+)
- Personal Records (PR) tracking, history, and retroactive PR detection (v0.3.0 → v0.4.4)
- Workout CRUD and logging with multi-movement support (v0.2.0+)
- Movement database with 31 seeded CrossFit movements
- WOD (Workout of the Day) management system with CRUD and seeded WODs (v0.4.0)
- Workout Template system and workout–WOD linking (v0.4.0)
- Pinia frontend stores for WODs and templates
- Admin tools: data cleanup, WOD record edits, audit logging, and admin user management (v0.4.5 → v0.4.6)
- Session management (list/revoke sessions) and session revocation endpoints (v0.4.6)
- CI, multi-database integration tests (sqlite, Postgres, MariaDB) and Docker Compose support
- PWA build + service worker (VitePWA) and frontend build pipeline

These completed items are documented in `CHANGELOG.md` (see latest entries v0.4.6-beta) and reflected in `TODO.md` completed sections.

### In Progress / High Priority
- Workout detail view and edit workflow — partially implemented, needs polish (TODO high-priority)
- Edit/delete workout with confirmation (TODO)
- Performance charts (movement progress visualization) (TODO)
- Template Library browsing and template-based logging integration (TODO)
- UI/UX improvements: loading states, notifications, skeleton loaders, improved time input UX (TODO)
- Backend: pagination and workout search/filter endpoints (TODO)

### Planned / Medium–Low Priority
- Add custom movements from the UI
- Workout sharing, photo uploads, rest timers, calendar view, dark mode
- Data export (CSV/JSON) and bulk admin cleanup tools
- Monitoring + observability (Prometheus/Grafana), production deployment guide
- Mobile apps (future)

Notes
For the full, authoritative list of completed changes look at `CHANGELOG.md` (recent releases up to v0.4.6-beta). The `TODO.md` file contains the prioritized backlog and `docs/` contains design and deployment notes.

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

1. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

1. **Install Go dependencies**

   ```bash
   go mod download
   ```

1. **Install frontend dependencies**

   ```bash
   cd web
   npm install
   cd ..
   ```

1. **Run the backend**

   ```bash
   # Terminal 1
   make run
   # Or: go run cmd/actalog/main.go
   ```

1. **Run the frontend**

   ```bash
   # Terminal 2
   cd web
   npm run dev
   ```

Local dev using an example subdomain

If you want to run the frontend so it is served from a public-style hostname (useful when testing PWA behavior, cookies, or integration with a reverse proxy), map an example subdomain to your local machine. Replace `subdomain.example.com` below with the hostname you prefer. On Windows, edit the hosts file as administrator:

```text
# Add this line to C:\Windows\System32\drivers\etc\hosts
127.0.0.1 subdomain.example.com
```

Notes:

- The Vite dev server can be configured to listen on `subdomain.example.com:3000` and HMR will expect that host. Ensure the hosts file entry points to the machine running the dev server.
- If you need HTTPS locally (for Service Worker/PWA testing), you will need to provision a local certificate for `subdomain.example.com` and configure your browser to trust it — Vite's default dev server uses HTTP. Be cautious when trusting self-signed certs.
- The production PWA manifest and built assets may be configured to use `https://subdomain.example.com/` as the base URL; building the frontend (`npm run build`) will produce assets and a manifest that assume that origin if the build `base`/`manifest` are set that way.

1. **Access the application**

   - Frontend: `http://localhost:3000`
   - Backend API: `http://localhost:8080`
   - API Health: `http://localhost:8080/health`

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

```text
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
See the top-level Roadmap section for current status and next priorities (keeps a single authoritative roadmap in this README).

