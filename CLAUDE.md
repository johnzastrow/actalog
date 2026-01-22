# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Project Overview

ActaLog is a mobile-first CrossFit workout tracker built with:
- **Backend:** Go (Chi router), SQLite/PostgreSQL/MySQL
- **Frontend:** Vue.js 3, Vuetify 3, Pinia
- **Architecture:** Clean Architecture with strict layer separation

**Version:** 1.1.0-beta

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
npm run test:run    # Run tests once
npm run test:coverage # Tests with coverage
npm run lint:fix    # Fix linting issues

# Docker
./docker/scripts/build.sh dev     # Build image with dev tag
docker tag ghcr.io/johnzastrow/actalog:dev ghcr.io/johnzastrow/actalog:latest
docker tag ghcr.io/johnzastrow/actalog:dev ghcr.io/johnzastrow/actalog:<version>
./docker/scripts/push.sh dev      # Push dev tag
./docker/scripts/push.sh latest   # Push latest tag
./docker/scripts/push.sh <version> # Push version tag (e.g., 1.1.0-beta)

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

## Feature Testing Workflow

**Default: Test features using Docker with multiple database backends.**

When testing new features, cycle through all supported databases to ensure compatibility:

| Database | Connection |
|----------|------------|
| SQLite | Local file: `./data/actalog.db` |
| MariaDB | `192.168.1.234:3306` (user: jcz) |
| PostgreSQL | `192.168.1.143:5432` (database: jcz, schema: actalog) |

**Email Testing (SMTP2GO):**
| Setting | Value |
|---------|-------|
| SMTP Server | `mail.smtp2go.com` |
| SMTP Port | `2525` |
| SMTP User | `acta@northredoubt.com` |
| SMTP Password | Set `$SMTP_PASSWORD` env var |

**Docker Testing Commands:**
```bash
# Build Docker image
./docker/scripts/build.sh dev

# Tag and push to registry (always push to dev, latest, and version tags)
docker tag ghcr.io/johnzastrow/actalog:dev ghcr.io/johnzastrow/actalog:latest
docker tag ghcr.io/johnzastrow/actalog:dev ghcr.io/johnzastrow/actalog:1.1.0-beta  # Use current version
./docker/scripts/push.sh dev
./docker/scripts/push.sh latest
./docker/scripts/push.sh 1.1.0-beta  # Use current version

# Test with SQLite (mount local data directory)
docker run -p 8080:8080 -v $(pwd)/data:/app/data \
  -e DB_DRIVER=sqlite3 -e DB_NAME=/app/data/actalog.db \
  ghcr.io/johnzastrow/actalog:dev

# Test with MariaDB (use --network host for external DB access)
docker run --network host \
  -e DB_DRIVER=mysql -e DB_HOST=192.168.1.234 -e DB_PORT=3306 \
  -e DB_USER=jcz -e DB_PASSWORD=$DB_PASSWORD -e DB_NAME=actalog \
  ghcr.io/johnzastrow/actalog:dev

# Test with MariaDB + Email enabled
docker run --network host \
  -e DB_DRIVER=mysql -e DB_HOST=192.168.1.234 -e DB_PORT=3306 \
  -e DB_USER=jcz -e DB_PASSWORD=$DB_PASSWORD -e DB_NAME=actalog \
  -e EMAIL_ENABLED=true -e EMAIL_FROM=acta@northredoubt.com \
  -e SMTP_HOST=mail.smtp2go.com -e SMTP_PORT=2525 \
  -e SMTP_USER=acta@northredoubt.com -e SMTP_PASSWORD=$SMTP_PASSWORD \
  ghcr.io/johnzastrow/actalog:dev

# Test with PostgreSQL (use --network host for external DB access)
docker run --network host \
  -e DB_DRIVER=postgres -e DB_HOST=192.168.1.143 -e DB_PORT=5432 \
  -e DB_USER=jcz -e DB_PASSWORD=$DB_PASSWORD -e DB_NAME=jcz -e DB_SCHEMA=actalog \
  ghcr.io/johnzastrow/actalog:dev
```

**When to build outside Docker:**
- Debugging with IDE/debugger
- Running unit tests (`make test`)
- Quick iteration during development
- When Docker overhead is unnecessary

## Process & Port Management

**Before starting any server, always check and clean up:**
```bash
# Check what's using port 8080
lsof -i :8080 | grep LISTEN

# Kill all actalog processes
pkill -9 -f actalog

# Kill Docker containers on port 8080
docker ps --filter "publish=8080" -q | xargs -r docker stop

# Full cleanup (processes + containers)
pkill -9 -f actalog 2>/dev/null; docker stop $(docker ps -q --filter "publish=8080") 2>/dev/null; sleep 1
```

**Standard startup sequence:**
1. Kill existing processes: `pkill -9 -f actalog`
2. Verify port is free: `lsof -i :8080 | grep LISTEN` (should be empty)
3. Start server
4. Verify running: `pgrep -f actalog` or `docker ps`

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

### Protected Users (DO NOT MODIFY)

**NEVER change any data for the following user accounts:**
- `br8kwall@gmail.com` - Do not modify password, email, role, or any other user data

### Claude Testing Account

**For all testing, use this dedicated account instead of modifying protected users:**
- Email: `claude@there.com`
- Password: `FuzzyBanana38`
- Role: `admin`

Create this account in all testing databases (PostgreSQL, MariaDB, SQLite) as needed.

### Credential Scanning (CRITICAL)

**Before EVERY commit and push, scan for credentials:**

1. **Never commit files containing:**
   - API keys, JWT secrets, or tokens
   - Database passwords or connection strings with credentials
   - Private keys (`.pem`, `.key` files)
   - `.env` files with real credentials (use `.env.example` instead)
   - Hardcoded passwords in source code

2. **Files to always check before committing:**
   - `.env`, `.env.*` (except `.env.example`)
   - `configs/`, `config.*` files
   - Any JSON/YAML with "password", "secret", "key", "token" fields
   - Docker compose files with environment variables

3. **Scan commands to run before pushing:**
   ```bash
   # Check for common credential patterns in staged files
   git diff --cached --name-only | xargs grep -l -E "(password|secret|api_key|token|credential).*=" 2>/dev/null

   # Check for .env files being tracked
   git ls-files | grep -E "^\.env$|\.env\.[^e]"
   ```

4. **If credentials are accidentally committed:**
   - Do NOT push to remote
   - Remove from history with `git reset` or `git filter-branch`
   - Rotate/regenerate the exposed credentials immediately

## UI Design

Colors: Primary `#00bcd4`, Header `#2c3e50`, Background `#f5f7fa`, PR/Action `#ffc107`. Layout: Fixed header 56px, bottom nav 70px.

## Database Version Management

SQLite snapshots in `db_versions/` for migration testing. Format: `actalog_X.Y.Z.db`. See `db_versions/README.md` for procedures.

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

## TODO Management

`docs/TODO.md` is the single source of truth. Use `[HIGH]`/`[MEDIUM]`/`[LOW]` markers for backlog. Keep only last 5 completed releases.

## Documentation Updates

**After implementing new features, update these documents:**

1. **`docs/USER_PERMISSIONS.md`** - Update whenever:
   - Adding new API endpoints
   - Adding new UI screens/routes
   - Changing permission requirements (auth, admin, subscription)
   - Adding new user actions or capabilities

2. **`docs/TODO.md`** - Mark completed items and add new tasks

3. **`docs/CHANGELOG.md`** - Document user-facing changes for releases

4. **`docs/DATABASE_SCHEMA.md`** - Update when adding/modifying tables

## Troubleshooting

**Frontend dependency issues:**
```bash
cd web && rm -rf node_modules package-lock.json && npm install
```

**Makefile cache error:**
The `make run` target creates cache directories automatically.

**First user becomes admin:**
The first registered user is automatically assigned admin role.
