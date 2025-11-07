# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ActaLog is a mobile-first CrossFit workout tracker built with Go backend (Chi router, SQLite/PostgreSQL) and Vue.js 3 frontend (Vuetify 3). The project follows Clean Architecture principles with strict separation between domain, service, repository, and handler layers.

**Current Version:** 0.1.0-alpha

## Essential Commands

### Backend Development

```bash
# Build application
make build

# Run application (backend API on :8080)
make run

# Run with auto-reload (requires air)
make dev

# Install development tools (air, goimports, golangci-lint)
make install-tools

# Run all tests with coverage report
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Run linter
make lint

# Format code
make fmt

# Clean build artifacts and cache
make clean

# Download and tidy dependencies
make deps
```

### Frontend Development

```bash
cd web

# Start dev server (on :3000)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Run linter
npm run lint

# Fix linting issues
npm run lint:fix

# Format code with Prettier
npm run format
```

### Docker Operations

```bash
# Start all services (backend, frontend, database)
make docker-up

# Stop all services
make docker-down

# View logs
make docker-logs

# Build Docker image
make docker-build
```

### Database Migrations

```bash
# Create new migration
make migrate-create name=create_users_table

# This creates timestamped up/down migration files in migrations/
```

## Architecture Principles

### Clean Architecture Layers

The codebase strictly follows dependency rules:

```
handlers → services → domain ← repositories
```

**Dependency Flow:**
- **Domain layer** (`internal/domain/`) defines entities and interfaces, has ZERO dependencies
- **Repository layer** (`internal/repository/`) implements data access, depends only on domain
- **Service layer** (`internal/service/`) implements business logic, depends only on domain
- **Handler layer** (`internal/handler/`) handles HTTP, depends on services and domain

### Directory Structure

```
internal/
├── domain/       # Business entities + repository interfaces (no dependencies)
├── repository/   # Data access implementations (depends on domain)
├── service/      # Business logic/use cases (depends on domain)
└── handler/      # HTTP handlers (depends on services)

pkg/              # Public reusable packages
├── auth/         # JWT authentication utilities
├── middleware/   # HTTP middleware (CORS, auth, logging)
└── version/      # Version management

cmd/actalog/      # Application entry point (main.go)
```

### Key Patterns

1. **Dependency Injection:** All dependencies injected via constructors
2. **Interface-Driven:** Domain defines interfaces, implementations in other layers
3. **Repository Pattern:** Data access abstracted through interfaces
4. **No Global State:** Everything passed explicitly through function parameters

## Database Configuration

The application supports three database drivers controlled via `DB_DRIVER` in `.env`:

- `sqlite3` - Default for development, file-based (DB_NAME=actalog.db)
- `postgres` - Recommended for production
- `mysql` - Supported alternative

**Important:** When using SQLite, the driver name in Go code must be `"sqlite3"` (not `"sqlite"`).

## Testing Strategy

### Test Structure

```
test/
├── unit/          # Fast, isolated unit tests
└── integration/   # Tests with real database/external dependencies
```

### Test Conventions

- Use table-driven tests for multiple scenarios
- Mock external dependencies using interfaces
- Run tests in parallel where safe (use `t.Parallel()`)
- Maintain >80% test coverage
- Each test must be isolated (no shared state)

### Running Specific Tests

```bash
# Run tests in specific package
go test -v ./internal/service/...

# Run specific test
go test -v -run TestUserService_Create ./internal/service/...

# Run with race detection
go test -race ./...
```

## Security Practices

1. **Password Hashing:** Always use bcrypt with cost factor ≥12
2. **SQL Injection:** Only parameterized queries (using sqlx or database/sql)
3. **JWT Tokens:** Secret must be changed from default in `.env` for production
4. **Input Validation:** Validate at handler layer before passing to services
5. **CORS:** Configure `CORS_ORIGINS` in `.env` to whitelist allowed origins

## OpenTelemetry Integration

The project includes OpenTelemetry for observability:

- Propagate `context.Context` through all function calls
- Start spans for significant operations (HTTP requests, DB queries, external calls)
- Attach relevant attributes (user_id, request_id, error messages)
- Use structured JSON logging with trace correlation
- Export to OTLP endpoint (configured via `OTEL_EXPORTER_OTLP_ENDPOINT`)

## Code Style and Conventions

### Go Code

- Follow standard Go formatting (enforced by `make fmt`)
- Run `goimports` to organize imports
- Run `golangci-lint` before committing
- Use descriptive variable names (avoid single-letter except in short scopes)
- Keep functions short and focused (single responsibility)
- Always handle errors explicitly (never ignore with `_`)
- Use wrapped errors for context: `fmt.Errorf("context: %w", err)`

### Vue.js Code

- Use Composition API with `<script setup>`
- Follow Vue 3 best practices
- Run ESLint and Prettier before committing
- Use Vuetify 3 components for UI consistency
- Store state in Pinia stores (not component local state for shared data)

## Configuration Management

All configuration via environment variables (see `.env.example`):

- Copy `.env.example` to `.env` for local development
- **Never commit `.env` file** (it's in `.gitignore`)
- Required for development: `DB_DRIVER`, `DB_NAME`, `JWT_SECRET`
- First registered user automatically becomes admin

## Common Development Tasks

### Adding a New Feature

1. Define domain entities and interfaces in `internal/domain/`
2. Implement repository in `internal/repository/`
3. Implement business logic in `internal/service/`
4. Create HTTP handlers in `internal/handler/`
5. Wire up routes in `cmd/actalog/main.go`
6. Write tests at each layer
7. Update API documentation if adding endpoints

### Creating Database Migrations

```bash
# Create migration files
make migrate-create name=add_user_preferences

# Edit the generated files:
# migrations/YYYYMMDDHHMMSS_add_user_preferences.up.sql
# migrations/YYYYMMDDHHMMSS_add_user_preferences.down.sql
```

### Adding a New Dependency

```bash
# Add Go dependency
go get github.com/some/package
go mod tidy

# Add npm dependency
cd web
npm install some-package
```

## Version Management

Version number is defined in `pkg/version/version.go` and should be incremented following semantic versioning:

- **Patch** (0.1.X): Bug fixes
- **Minor** (0.X.0): New features (backward compatible)
- **Major** (X.0.0): Breaking changes

Update version in both:
- `pkg/version/version.go`
- `web/package.json`

## Development Workflow

1. **Start backend:** `make run` or `make dev` (with auto-reload)
2. **Start frontend:** `cd web && npm run dev`
3. **Access application:**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Health check: http://localhost:8080/health

## Documentation References

Key documentation in `docs/`:
- `ARCHITECTURE.md` - Detailed architecture and design patterns
- `DATABASE_SCHEMA.md` - Complete database schema with ERD
- `AI_INSTRUCTIONS.md` - Development guidelines and best practices
- `REQUIIREMENTS.md` - Project requirements and user stories
- `TODO.md` - Planned features and improvements
- `CHANGELOG.md` - Version history and changes

## Project-Specific Notes

### Clean Architecture Compliance

When adding or modifying code:
- Domain layer must remain dependency-free (no imports except standard library)
- Services must not know about HTTP, handlers, or delivery mechanisms
- Repositories must only implement interfaces defined in domain
- Handlers should be thin, delegating business logic to services

### Standard CrossFit Movements

The app includes pre-seeded movements (Weightlifting, Gymnastics, Bodyweight, Cardio). See `docs/DATABASE_SCHEMA.md` for the complete list. Users can also create custom movements.

### JWT Authentication Flow

1. User logs in via `POST /api/auth/login`
2. Backend validates credentials, returns JWT token
3. Frontend stores token and includes in `Authorization: Bearer <token>` header
4. Middleware validates JWT and extracts user context
5. Handlers receive authenticated user ID from context

### Multi-Database Support

When writing queries:
- Use `?` placeholders for SQLite and MySQL
- Use `$1, $2, ...` placeholders for PostgreSQL
- Consider using sqlx for consistent query handling
- Test migrations against all supported databases
