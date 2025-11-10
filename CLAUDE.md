# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ActaLog is a mobile-first CrossFit workout tracker built with Go backend (Chi router, SQLite/PostgreSQL) and Vue.js 3 frontend (Vuetify 3). The project follows Clean Architecture principles with strict separation between domain, service, repository, and handler layers.

**Current Version:** 0.2.0-beta

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

# Clean up and rebuild dependencies (troubleshooting)
rm -rf node_modules package-lock.json
npm cache clean --force
npm install
```

**Frontend Troubleshooting:**

If you encounter build issues or dependency problems:

```bash
cd web

# Option 1: Quick reinstall
npm install

# Option 2: Clear cache and reinstall
npm cache clean --force && npm install

# Option 3: Complete cleanup (for corrupted dependencies)
rm -rf node_modules package-lock.json
npm cache clean --force
npm install

# Verify build works
npm run dev
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
- `mysql` - Supported alternative (MySQL or MariaDB)

**How to Switch Databases:**

1. Edit `.env` and change `DB_DRIVER`:
   ```env
   # For SQLite (development)
   DB_DRIVER=sqlite3
   DB_NAME=actalog.db

   # For PostgreSQL (production)
   DB_DRIVER=postgres
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=actalog
   DB_PASSWORD=your_password
   DB_NAME=actalog

   # For MySQL/MariaDB (production)
   DB_DRIVER=mysql
   DB_HOST=localhost
   DB_PORT=3306
   DB_USER=actalog
   DB_PASSWORD=your_password
   DB_NAME=actalog
   ```

2. Restart the application - it will automatically connect to the new database and run migrations

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

## Implemented Features (v0.2.0-beta)

### Workout Management

**Location:** `internal/repository/workout_repository.go`, `internal/service/workout_service.go`, `internal/handler/workout_handler.go`

Complete workout CRUD functionality:
- **Create Workout:** `POST /api/workouts` - Create new workout with movements
- **List Workouts:** `GET /api/workouts` - Fetch user's workouts with movement details
- **Get Workout:** `GET /api/workouts/{id}` - Fetch single workout by ID
- **Update Workout:** `PUT /api/workouts/{id}` - Modify existing workout
- **Delete Workout:** `DELETE /api/workouts/{id}` - Remove workout

**Key Implementation Details:**
- Workouts are user-scoped (users can only access their own workouts)
- Service layer enforces authorization checks
- Movement details are eager-loaded for display (movement names, not just IDs)
- Supports nullable fields (notes, workout_name, total_time)
- Handles workout_movements as a sub-collection

**Frontend Integration:**
- `web/src/views/LogWorkoutView.vue` - Workout creation form with autocomplete
- `web/src/views/WorkoutsView.vue` - List view with movement details
- `web/src/views/DashboardView.vue` - Recent workouts and statistics

### Movement Management

**Location:** `internal/repository/movement_repository.go`

31 standard CrossFit movements are automatically seeded on first run:
- Weightlifting: Back Squat, Deadlift, Bench Press, Clean, Snatch, etc.
- Gymnastics: Pull-ups, Muscle-ups, Handstand Push-ups, Toes-to-Bar, etc.
- Cardio: Running, Rowing, Air Bike, Jump Rope, etc.

**API Endpoints:**
- `GET /api/movements` - List all available movements
- `GET /api/movements/search?q=squat` - Search movements by name

**Seeding Logic:** See `internal/repository/database.go` function `seedStandardMovements()`

### Dashboard & Statistics

**Location:** `web/src/views/DashboardView.vue`

Real-time workout statistics:
- Total workouts count (lifetime)
- Monthly workouts count (current month)
- Recent 5 workouts with movement details
- Quick action button to log new workout

**Data Flow:**
1. Component calls `GET /api/workouts` on mount
2. Processes response to calculate stats in frontend
3. Displays formatted workout cards with date and movements

### Authentication & Authorization

**Location:** `pkg/middleware/auth.go`

JWT middleware protects all workout endpoints:
- Extracts token from `Authorization: Bearer <token>` header
- Validates JWT signature and expiration
- Adds user context (ID, email, role) to request
- Returns 401 for missing/invalid tokens

**Protected Routes (in main.go):**
```go
r.Group(func(r chi.Router) {
    r.Use(middleware.Auth(cfg.JWT.SecretKey))
    r.Post("/workouts", workoutHandler.CreateWorkout)
    r.Get("/workouts", workoutHandler.ListWorkouts)
    r.Get("/workouts/{id}", workoutHandler.GetWorkout)
    // ... other protected routes
})
```

### UI Design System

**Color Palette:**
- Primary Accent: `#00bcd4` (cyan/turquoise)
- Header Background: `#2c3e50` (dark navy)
- Page Background: `#f5f7fa` (light gray)
- Text Primary: `#1a1a1a` (very dark gray)
- Text Secondary: `#666` (medium gray)
- Action Button: `#ffc107` (amber)

**Layout Pattern:**
- Fixed header (v-app-bar) at top with z-index: 10
- Scrollable content area with `margin-top: 56px, margin-bottom: 70px`
- Fixed bottom navigation with elevation: 8
- Content uses `overflow-y: auto` for scrolling

**Common Components:**
- Bottom navigation replicated across all main views
- Autocomplete search with magnify icon for movement selection
- Card-based layout with `elevation="0"` and `rounded="lg"`
- Consistent spacing (mb-3 for sections, pa-3 for padding)

### Autocomplete Search Implementation

**Location:** `web/src/views/LogWorkoutView.vue`, `web/src/views/PerformanceView.vue`

Searchable movement selection using Vuetify v-autocomplete:
```vue
<v-autocomplete
  v-model="selectedMovement"
  :items="movements"
  item-title="title"
  item-value="value"
  :loading="loading"
  placeholder="Type to search movements..."
  clearable
  auto-select-first
>
  <template #prepend-inner>
    <v-icon color="#00bcd4" size="small">mdi-magnify</v-icon>
  </template>
  <template #item="{ props, item }">
    <!-- Custom item template with icon and type -->
  </template>
</v-autocomplete>
```

**Features:**
- Type-ahead search filtering
- Custom item templates showing movement type
- Icon differentiation (cyan for weightlifting, gray for others)
- Auto-select first match
- Clearable selection

### Known Fixes Applied

**Makefile Cache Issue:**
- Added `@mkdir -p $(GO_BUILD_CACHE) $(GO_MOD_CACHE) $(CACHE_DIR)/tmp` to `run` and `dev` targets
- Prevents "stat .cache/tmp: no such file or directory" error

**SQLite Driver Name:**
- Changed default `DB_DRIVER` from `"sqlite"` to `"sqlite3"` in `configs/config.go:66`
- Matches the imported driver name from `github.com/mattn/go-sqlite3`

**Scrolling Issues:**
- Added `overflow-y: auto` to container styles
- Proper margins to account for fixed header (56px) and bottom nav (70px)
- Reduced spacing (mb-4 → mb-3) for tighter mobile layout

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
