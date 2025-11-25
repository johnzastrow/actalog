# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Project Overview

ActaLog is a mobile-first CrossFit workout tracker built with:
- **Backend:** Go (Chi router), SQLite/PostgreSQL/MySQL
- **Frontend:** Vue.js 3, Vuetify 3, Pinia
- **Architecture:** Clean Architecture with strict layer separation

**Version:** 0.10.0-beta

## Quick Reference

```bash
# Backend
make build          # Build (auto-increments build number)
make run            # Run on :8080
make dev            # Run with hot reload (requires air)
make test           # Run all tests
make lint           # Run linter
make fmt            # Format code

# Frontend (from web/)
npm run dev         # Dev server on :3000
npm run build       # Production build
npm run lint:fix    # Fix linting issues

# Docker
./docker/scripts/build.sh <tag>   # Build image
./docker/scripts/push.sh <tag>    # Push to ghcr.io

# Migrations
make migrate-create name=add_feature
```

## Architecture

### Clean Architecture Layers

```
handlers → services → domain ← repositories
```

| Layer | Location | Depends On | Responsibility |
|-------|----------|------------|----------------|
| Domain | `internal/domain/` | Nothing | Entities, interfaces |
| Repository | `internal/repository/` | Domain | Data access |
| Service | `internal/service/` | Domain | Business logic |
| Handler | `internal/handler/` | Services, Domain | HTTP handling |

### Directory Structure

```
internal/
├── domain/       # Entities + repository interfaces (ZERO dependencies)
├── repository/   # Data access implementations
├── service/      # Business logic/use cases
└── handler/      # HTTP handlers

pkg/
├── auth/         # JWT utilities
├── middleware/   # HTTP middleware
├── prmath/       # 1RM calculation formulas
└── version/      # Version management

cmd/actalog/      # Application entry point
web/              # Vue.js frontend
migrations/       # Database migrations
```

### Key Patterns

1. **Dependency Injection** - All dependencies via constructors
2. **Interface-Driven** - Domain defines interfaces, others implement
3. **Repository Pattern** - Data access abstracted through interfaces
4. **No Global State** - Everything passed explicitly

## Adding Features

1. Define entities/interfaces in `internal/domain/`
2. Implement repository in `internal/repository/`
3. Implement business logic in `internal/service/`
4. Create HTTP handlers in `internal/handler/`
5. Wire routes in `cmd/actalog/main.go`
6. Write tests at each layer

## Database

**Drivers** (set `DB_DRIVER` in `.env`):
- `sqlite3` - Development default
- `postgres` - Production recommended
- `mysql` - MySQL/MariaDB

**Query Placeholders:**
- SQLite/MySQL: `?`
- PostgreSQL: `$1, $2, ...`

**Important:** SQLite driver name must be `"sqlite3"` (not `"sqlite"`).

## Development Workflow

**Local Development:**
1. Terminal 1: `make run` (backend on :8080)
2. Terminal 2: `cd web && npm run dev` (frontend on :3000)
3. Vite proxy forwards `/api` and `/uploads` to backend

**Production (Docker):**
- Single port :8080 serves both API and static frontend
- No separate Node.js process
- `cmd/actalog/main.go:418-436` handles static file serving

## Code Style

### Go
- `make fmt` and `make lint` before committing
- Always handle errors explicitly (never `_`)
- Wrap errors: `fmt.Errorf("context: %w", err)`
- Keep functions focused (single responsibility)

### Vue.js
- Composition API with `<script setup>`
- Vuetify 3 components for UI
- Pinia for shared state
- Run ESLint and Prettier before committing

## Security

- Bcrypt with cost ≥12 for passwords
- Parameterized queries only (no string concatenation)
- JWT secret must be changed from default in production
- Validate input at handler layer
- Configure `CORS_ORIGINS` in `.env`

## UI Design

**Colors:**
- Primary: `#00bcd4` (cyan)
- Header: `#2c3e50` (dark navy)
- Background: `#f5f7fa`
- PR/Action: `#ffc107` (gold/amber)

**Layout:**
- Fixed header (56px), fixed bottom nav (70px)
- Content: `margin-top: 56px, margin-bottom: 70px, overflow-y: auto`

## Testing

```bash
go test -v ./internal/service/...           # Specific package
go test -v -run TestName ./...              # Specific test
go test -race ./...                         # Race detection
```

- Table-driven tests for multiple scenarios
- Mock dependencies using interfaces
- Tests must be isolated (no shared state)

## Configuration

**Backend** (`.env`):
- `DB_DRIVER`, `DB_NAME` - Database settings
- `JWT_SECRET` - Must change for production
- `CORS_ORIGINS` - Allowed frontend origins
- `EMAIL_*`, `SMTP_*` - Email configuration

**Frontend** (`web/.env`):
- `VITE_API_BASE_URL` - Backend URL (only needed if different domain)

## Key Files

| Purpose | Location |
|---------|----------|
| Entry point | `cmd/actalog/main.go` |
| Routes | `cmd/actalog/main.go:350-450` |
| Config | `configs/config.go` |
| Version | `pkg/version/version.go` |
| Auth middleware | `pkg/middleware/auth.go` |
| DB setup | `internal/repository/database.go` |

## API Patterns

**Authentication:**
- `POST /api/auth/login` → Returns JWT token
- Include `Authorization: Bearer <token>` header
- Middleware extracts user context (ID, email, role)

**Resource Ownership:**
- All user data is user-scoped
- Service layer enforces authorization
- Admin routes use `middleware.AdminOnly`

**Response Format:**
```json
{"error": "message"}           // Errors
{"data": [...], "count": N}    // Lists
{...entity fields...}          // Single items
```

## Documentation

Additional docs in `docs/`:
- `ARCHITECTURE.md` - Detailed design patterns
- `DATABASE_SCHEMA.md` - Complete schema with ERD
- `CHANGELOG.md` - Version history
- `TODO.md` - Planned features

## Troubleshooting

**Frontend dependency issues:**
```bash
cd web && rm -rf node_modules package-lock.json && npm install
```

**Makefile cache error:**
The `make run` target creates cache directories automatically.

**First user becomes admin:**
The first registered user is automatically assigned admin role.
