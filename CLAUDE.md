# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ActaLog is a mobile-first CrossFit workout tracker built with Go backend (Chi router, SQLite/PostgreSQL) and Vue.js 3 frontend (Vuetify 3). The project follows Clean Architecture principles with strict separation between domain, service, repository, and handler layers.

**Current Version:** 0.4.4-beta

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

### Backend Configuration

All configuration via environment variables (see `.env.example`):

- Copy `.env.example` to `.env` for local development
- **Never commit `.env` file** (it's in `.gitignore`)
- Required for development: `DB_DRIVER`, `DB_NAME`, `JWT_SECRET`
- First registered user automatically becomes admin

### Frontend Configuration

The frontend uses environment variables prefixed with `VITE_` (see `web/.env.example`):

**Development (localhost):**
- Uses Vite proxy for `/api` and `/uploads` routes
- No environment variables needed
- Proxy automatically forwards requests to `localhost:8080`

**Production Deployment:**
- Set `VITE_API_BASE_URL` to your backend URL if on different domain/port
- Examples:
  - Same domain, different port: `http://your-domain.com:8080`
  - Different domain: `https://api.your-domain.com`
  - Local network: `http://192.168.1.100:8080`

**URL Utilities (`web/src/utils/url.js`):**
- `getApiBaseUrl()` - Returns environment-aware API base URL
- `getAssetUrl(path)` - Converts relative paths to absolute URLs
- `getProfileImageUrl(profileImage)` - Handles profile image URLs

These utilities automatically:
- Use Vite proxy in development
- Use relative URLs when possible
- Fall back to `window.location` in production
- Support `VITE_API_BASE_URL` environment variable override

**Important for Deployment:**
- Backend `.env` must set `CORS_ORIGINS` to include frontend URL
- Backend `.env` must set `APP_URL` to frontend URL (for email links)

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

### Build Number Auto-Increment

The build number is automatically incremented with each build:
- **Build number** is stored in `pkg/version/version.go` (Build constant)
- **Automatic increment** happens when you run `make build`
- The script `scripts/increment-build.sh` handles the increment
- Format: `0.4.1-beta+build.3`

**How it works:**
1. Running `make build` calls `scripts/increment-build.sh`
2. Script extracts current build number from `pkg/version/version.go`
3. Increments the build number by 1
4. Updates the file with new build number
5. Builds the application

**Version Display:**
- Backend exposes version via `/api/version` endpoint (public, no auth required)
- Returns: `version`, `build`, `fullVersion`, `app` fields
- Frontend displays in Profile screen (top card)
- Shows: "Version: 0.4.1-beta+build.4" and "Build: #4"

**Manual Version Updates:**

When releasing a new version, update:
1. `pkg/version/version.go` - Major, Minor, Patch, PreRelease constants
2. `web/package.json` - version field
3. Build number is auto-incremented, no manual update needed

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

## Implemented Features

### Personal Records (PR) Tracking (v0.3.0-beta)

**Location:** `internal/domain/movement.go`, `internal/repository/workout_movement_repository.go`, `internal/service/workout_service.go`, `internal/handler/workout_handler.go`

Complete PR tracking system with automatic detection and manual control:
- **Auto-Detection:** Automatically flags PRs when weight exceeds previous max for a movement
- **Get Personal Records:** `GET /api/workouts/prs` - Aggregated max weight, reps, time per movement
- **Get PR Movements:** `GET /api/workouts/pr-movements?limit=5` - Recent PR-flagged movements
- **Toggle PR Flag:** `POST /api/workouts/movements/:id/toggle-pr` - Manual PR flag control

**Key Implementation Details:**
- Database migration v0.3.0 adds `is_pr` BOOLEAN field to `workout_movements` table
- Multi-database support (SQLite: INTEGER, PostgreSQL/MySQL: BOOLEAN)
- PR detection integrated into workout creation workflow via `DetectAndFlagPRs()` service method
- Compares current weight against `GetMaxWeightForMovement()` for the user
- Authorization checks ensure users can only toggle PRs on their own workouts
- Repository methods aggregate data: `GetPersonalRecords()` returns MAX(weight), MAX(reps), MIN(time)

**Frontend Integration:**
- `web/src/views/PRHistoryView.vue` - Dedicated PR history page at `/prs` route
- `web/src/components/RecentWorkoutsCards.vue` - Gold PR chip badges on workout cards
- `web/src/views/WorkoutsView.vue` - Gold trophy icons (mdi-trophy) next to PR movements
- Visual design: Gold/amber color scheme (#ffc107) for PR indicators

**Domain Models:**
```go
type WorkoutMovement struct {
  // ... existing fields
  IsPR bool `json:"is_pr" db:"is_pr"`
}

type PersonalRecord struct {
  MovementID   int64
  MovementName string
  MaxWeight    *float64
  MaxReps      *int
  BestTime     *int
  WorkoutID    int64
  WorkoutDate  time.Time
}
```

### Retroactive PR Detection (v0.4.4-beta)

**Location:** `internal/service/user_workout_service.go`, `internal/repository/user_workout_movement_repository.go`, `internal/repository/user_workout_wod_repository.go`, `scripts/retroactive_prs.go`

System to analyze historical workouts and flag PRs retroactively:
- **Service Method:** `RetroactivelyFlagPRs(userID)` - Processes all workouts chronologically
- **API Endpoint:** `POST /api/workouts/retroactive-flag-prs` - Authenticated endpoint for retroactive PR flagging
- **CLI Script:** `scripts/retroactive_prs.go` - Direct database script to flag PRs for a user

**Key Implementation Details:**
- Processes workouts in chronological order (by workout_date)
- Tracks max weights per movement_id in-memory during processing
- Tracks best times and best rounds+reps per wod_id in-memory
- Repository methods: `UpdatePRFlag(id, isPR)` for movements and WODs
- Flags PRs when new max is achieved chronologically
- Returns count of movement PRs and WOD PRs flagged
- Multi-database support (SQLite, PostgreSQL, MySQL)

**How It Works:**
1. Fetches all user workouts ordered by date (chronologically)
2. For each workout, gets all movements and WODs
3. For movements with weight: compares to historical max, flags if new max
4. For WODs with time: compares to historical best time (lower is better), flags if faster
5. For WODs with rounds+reps: compares to historical best, flags if more rounds/reps
6. Updates database with `is_pr = true/false` for each performance record

**Usage:**
```bash
# Via CLI script
go run scripts/retroactive_prs.go

# Via API (requires authentication)
curl -X POST http://localhost:8080/api/workouts/retroactive-flag-prs \
  -H "Authorization: Bearer <token>"
```

**Use Cases:**
- Fixing historical data after PR system implementation
- Recalculating PRs after data corrections
- Initial PR flagging for existing users migrating to new system
- Batch processing for multiple users via script modification

### Wodify Performance Import (v0.7.0-beta)

**Location:** `internal/domain/wodify_import.go`, `internal/service/wodify_parser.go`, `internal/service/wodify_import_service.go`, `internal/handler/wodify_import_handler.go`

Complete system to import workout history from Wodify Performance CSV exports:
- **Preview Import:** `POST /api/import/wodify/preview` - Analyzes CSV and shows what will be imported
- **Confirm Import:** `POST /api/import/wodify/confirm` - Executes the import with auto-entity creation
- **Frontend Integration:** Import page with Wodify-specific preview showing workout summaries

**Key Implementation Details:**
- **CSV Parser:** Handles 19-column Wodify performance export format with multi-line field support
- **Result String Parsing:** Regex-based parser for 9 different result types (Weight, Time, AMRAP variations, Calories, Distance, etc.)
  - `"3 x 10 @ 85 lbs"` → sets: 3, reps: 10, weight: 85
  - `"7 + 3"` → rounds: 7, reps: 3 (AMRAP)
  - `"5:30"` → time_seconds: 330
- **Date Grouping:** Groups performances by date to create cohesive UserWorkout entries
- **Auto-Entity Creation:** Automatically creates movements and WODs if they don't exist in database
- **PR Preservation:** Maintains "Is Personal Record" flags from Wodify export
- **Preview System:** Shows total rows, unique workout dates, new entities to be created, and workout summaries

**Result Type Parsing:**
- `Weight` - Parses weightlifting results like "3 x 10 @ 85 lbs"
- `Time` - Parses time results like "5:30" (MM:SS) or "1:05:30" (HH:MM:SS)
- `AMRAP - Rounds and Reps` - Parses "7 + 3" format
- `AMRAP - Reps` - Parses "50 Reps"
- `AMRAP - Rounds` - Parses "5 Rounds"
- `Max reps` - Parses "3 x 8" (sets x reps)
- `Calories` - Parses "133 Calories"
- `Distance` - Parses "500 m"
- `Each Round` - Parses "175 Total Reps"

**Frontend Integration:**
- `web/src/views/ImportView.vue` - Added "Wodify Performance" import type
- **Wodify-Specific Preview:**
  - Summary stats: total rows, valid rows, workout dates, entities to create
  - New Entities card: chips showing movements and WODs to be auto-created
  - Workout Summary table: shows date, movement count, WOD count, component types, PR flags
- **Success Message:** Displays workouts created, performances, movements/WODs auto-created, PRs flagged

**Domain Models:**
```go
type WodifyPerformanceRow struct {
  Date                  string // MM/DD/YYYY
  ComponentType         string // "Weightlifting", "Metcon", "Gymnastics"
  ComponentName         string
  PerformanceResultType string // "Weight", "Time", "AMRAP - Rounds and Reps", etc.
  FullyFormattedResult  string
  IsPersonalRecord      bool
  Comment               string
  // ... 12 more fields
}

type WodifyImportPreview struct {
  TotalRows             int
  ValidRows             int
  UniqueWorkoutDates    int
  MovementsToCreate     int
  WODsToCreate          int
  UserWorkoutsToCreate  int
  PerformancesToCreate  int
  WorkoutSummary        []WodifyWorkoutSummary
  NewMovements          []string
  NewWODs               []string
}

type WodifyImportResult struct {
  WorkoutsCreated     int
  MovementsCreated    int
  WODsCreated         int
  PerformancesCreated int
  PRsFlagged          int
}
```

**Usage Example:**
```bash
# Preview Wodify CSV import
curl -X POST http://localhost:8080/api/import/wodify/preview \
  -H "Authorization: Bearer <token>" \
  -F "file=@Performance_fromWodify.csv"

# Confirm import
curl -X POST http://localhost:8080/api/import/wodify/confirm \
  -H "Authorization: Bearer <token>" \
  -F "file=@Performance_fromWodify.csv"
```

**Real-World Test Results:**
- Successfully imported 293 performance entries from 6+ years of data (2018-2025)
- Created 189 user workouts (grouped by date)
- Auto-created 37 new movements and 28 new WODs
- Automatically flagged 62 PRs from Wodify data
- 1 invalid row handled gracefully (missing component type/name)

### 1RM (One-Rep Max) Calculation and Display (v0.7.2-beta)

**Location:** `internal/handler/performance_handler.go`, `pkg/prmath/one_rm.go`, `web/src/views/PerformanceView.vue`

Complete system for calculating and displaying estimated one-rep maximums for strength movements:
- **Backend API Enhancement:** `GET /api/performance/movements/{id}` now calculates and returns 1RM data
- **Frontend Display:** Performance screen shows best 1RM, trend chart, and per-record estimates
- **Formula Selection:** Automatic formula selection based on rep range (Epley, Wathan, or Actual)

**Key Implementation Details:**
- **Backend (`internal/handler/performance_handler.go`):**
  - New response type: `MovementPerformanceWithRM` with `calculated_1rm` and `formula` fields
  - Returns `best_1rm` - overall best estimated 1RM across all performances
  - Returns `best_formula` - formula used for best estimate (e.g., "Epley (2-10 reps)")
  - Calculation performed on-the-fly (no database storage required)

- **Calculation Package (`pkg/prmath/one_rm.go`):**
  - Hybrid formula approach: `Calculate1RM(weight, reps) → (oneRM, formula)`
  - 1 rep: Returns actual weight as "Actual 1RM"
  - 2-10 reps: Epley formula `1RM = weight × (1 + reps/30)`
  - 11+ reps: Wathan formula `1RM = (100 × weight) / (48.8 + 53.8 × e^(-0.075 × reps))`
  - Also includes `CalculateAllFormulas()` for comparing multiple estimates

- **Multi-database Support:** Works with SQLite, PostgreSQL, and MySQL (no schema changes)

**Frontend Integration (`web/src/views/PerformanceView.vue`):**

1. **Best 1RM Stat Card:**
   - Gold-colored display (#ffc107) with arm-flex icon (mdi-arm-flex)
   - Shows rounded estimated 1RM value
   - Displays formula chip indicating calculation method
   - Only appears when weight/reps data is available

2. **Performance History Enhancement:**
   - Each movement entry shows "Est. 1RM: XXX lbs" in gold text
   - Appears alongside workout date and notes in subtitle
   - Provides quick reference for strength progression

3. **Dual-Line Performance Chart:**
   - **Solid dark line (#2c3e50):** Actual weight lifted
   - **Dashed gold line (#ffc107):** Estimated 1RM trend
   - Legend automatically displays when 1RM data exists
   - Y-axis labeled "Weight (lbs)" for clarity

4. **Enhanced Chart Tooltips:**
   - Weight line: Shows actual weight, reps, and PR flag
   - 1RM line: Shows calculated 1RM with formula (e.g., "Est. 1RM: 225 lbs (Epley)")
   - Null value filtering prevents gaps in visualization

**API Response Example:**
```json
{
  "performances": [
    {
      "id": 123,
      "weight": 185,
      "reps": 5,
      "calculated_1rm": 215.5,
      "formula": "Epley (2-10 reps)",
      "is_pr": true,
      "workout_date": "2025-01-22"
    }
  ],
  "best_1rm": 225.0,
  "best_formula": "Epley (2-10 reps)",
  "count": 15
}
```

**Use Cases:**
- Track estimated strength progression over time without testing true 1RM
- Compare performance across different rep schemes
- Identify when to increase working weights
- Visualize strength trends alongside actual workout data

### Password Reset System (v0.3.0-beta)

**Location:** `internal/repository/password_reset_repository.go`, `internal/service/user_service.go`, `internal/handler/auth_handler.go`, `web/src/views/ForgotPasswordView.vue`, `web/src/views/ResetPasswordView.vue`

Complete password reset flow with email delivery:
- **Forgot Password:** `POST /api/auth/forgot-password` - Generate reset token and send email
- **Reset Password:** `POST /api/auth/reset-password` - Validate token and update password
- **Frontend Routes:** `/forgot-password` and `/reset-password/:token`

**Key Implementation Details:**
- Database migration adds `password_resets` table with token, user_id, expires_at, used_at
- Secure token generation using crypto/rand (32 bytes, hex-encoded)
- Token expiration (configurable, default 1 hour)
- Email delivery via SMTP with configurable templates
- Single-use tokens (marked as used after successful password reset)
- Authorization: token validation ensures only valid, unexpired, unused tokens work

**Frontend Integration:**
- `web/src/views/ForgotPasswordView.vue` - Email input form for password reset request
- `web/src/views/ResetPasswordView.vue` - New password form with token validation
- `web/src/views/LoginView.vue` - "Forgot password?" link between sign-in and register
- Router guards prevent access to reset flows when already authenticated
- Success/error messaging with user-friendly feedback

### Email Verification System (v0.3.1-beta)

**Location:** `internal/repository/email_verification_repository.go`, `internal/service/user_service.go`, `internal/handler/auth_handler.go`, `web/src/views/VerifyEmailView.vue`, `web/src/views/ResendVerificationView.vue`

Complete email verification flow with automated email delivery:
- **Verify Email:** `GET /api/auth/verify-email?token=...` - Validate token and mark email as verified
- **Resend Verification:** `POST /api/auth/resend-verification` - Send new verification email
- **Frontend Routes:** `/verify-email?token=...` and `/resend-verification`

**Key Implementation Details:**
- Database migration v0.3.1 adds `email_verified` and `email_verified_at` columns to users table
- Creates `email_verification_tokens` table with token, user_id, expires_at, used_at
- Secure token generation using crypto/rand (32 bytes, hex-encoded)
- Token expiration (24 hours for email verification)
- SMTP email delivery with styled HTML templates
- Single-use tokens (marked as used after successful verification)
- Verification email sent automatically on user registration
- Authorization: users can only resend verification for their own email

**Email Service:**
- SMTP configuration via environment variables (EMAIL_FROM, SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS)
- HTML email templates with inline styles for cross-client compatibility
- Verification link format: `https://domain.com/verify-email?token={token}`
- SendVerificationEmail() method constructs and sends styled HTML emails

**Repository Layer:**
- `CreateVerificationToken()` - Generates and stores verification token
- `GetVerificationToken()` - Retrieves token with validation (expiration, used status)
- `MarkTokenAsUsed()` - Marks token as used after successful verification
- `UpdateEmailVerified()` - Sets email_verified=true and email_verified_at timestamp

**Service Layer:**
- `SendVerificationEmail()` - Creates token and sends verification email
- `VerifyEmailWithToken()` - Validates token and updates user email_verified status
- `ResendVerificationEmail()` - Creates new token and resends verification email
- Authorization checks ensure users can only verify their own emails

**Frontend Integration:**
- `web/src/views/VerifyEmailView.vue` - Email verification page with token validation
  - Automatic verification on page load using query parameter token
  - Three states: Loading, Success, Error
  - Updates auth store user object on successful verification
  - Handles expired, invalid, and already-used tokens with appropriate error messages
- `web/src/views/ResendVerificationView.vue` - Resend verification email page
  - Email input form to request new verification email
  - Success confirmation displaying the email address
  - Error handling for 404 (user not found) and 400 (already verified)
- `web/src/views/RegisterView.vue` - Updated to show verification success message
  - No longer auto-redirects to dashboard after registration
  - Displays sent email address and 24-hour expiration notice
  - Link to resend verification if email not received
- `web/src/views/DashboardView.vue` - Verification status banner
  - Warning alert for users with unverified emails
  - Prominent "Resend Email" button
  - Closable alert for better UX
- Router guards allow `/verify-email` access for both authenticated and unauthenticated users

**Domain Models:**
```go
type User struct {
  // ... existing fields
  EmailVerified   bool       `json:"email_verified" db:"email_verified"`
  EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
}

type EmailVerificationToken struct {
  ID        int64      `json:"id" db:"id"`
  UserID    int64      `json:"user_id" db:"user_id"`
  Token     string     `json:"token" db:"token"`
  ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
  UsedAt    *time.Time `json:"used_at,omitempty" db:"used_at"`
  CreatedAt time.Time  `json:"created_at" db:"created_at"`
}
```

### Session Management System (v0.4.6-beta)

**Location:** `internal/repository/refresh_token_repository.go`, `internal/service/user_service.go`, `internal/handler/session_handler.go`

Complete session (refresh token) management for security and user control:
- **List Sessions:** `GET /api/sessions` - Get all active sessions for authenticated user
- **Revoke Session:** `DELETE /api/sessions/{id}` - Revoke a specific session
- **Revoke All Sessions:** `POST /api/sessions/revoke-all` - Revoke all user sessions

**Key Implementation Details:**
- Session ownership validation (users can only manage their own sessions)
- Service layer methods: `GetActiveSessions()`, `RevokeSession()`, `RevokeAllSessions()`
- Audit logging for all session operations (security trail)
- Sessions are refresh tokens stored in `refresh_tokens` table
- Each session has: id, user_id, token, expires_at, created_at, revoked_at, device_info
- Authorization checks prevent users from revoking other users' sessions

**Security Features:**
- All endpoints require authentication
- Token ownership verified before any operation
- Audit trail logged with user_email and token_id
- Supports "revoke all except current" for security scenarios
- Expired tokens automatically excluded from active session list

**Use Cases:**
- User wants to log out from all devices
- User suspects unauthorized access and wants to revoke all sessions
- User manages active sessions from security settings
- Admin monitoring via audit logs

**Domain Models:**
```go
type RefreshToken struct {
  ID         int64      `json:"id" db:"id"`
  UserID     int64      `json:"user_id" db:"user_id"`
  Token      string     `json:"token" db:"token"`
  ExpiresAt  time.Time  `json:"expires_at" db:"expires_at"`
  CreatedAt  time.Time  `json:"created_at" db:"created_at"`
  RevokedAt  *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
  DeviceInfo string     `json:"device_info,omitempty" db:"device_info"`
}
```

### Admin User Management (v0.4.6-beta)

**Location:** `internal/service/user_service.go`, `internal/handler/admin_user_handler.go`, `web/src/views/AdminUsersView.vue`

Complete administrative control over user accounts:
- **List Users:** `GET /api/admin/users` - Paginated list with all user details
- **Get User Details:** `GET /api/admin/users/{id}` - Full user information with admin fields
- **Unlock Account:** `POST /api/admin/users/{id}/unlock` - Remove temporary account lock
- **Disable Account:** `POST /api/admin/users/{id}/disable` - Permanently disable with reason
- **Enable Account:** `POST /api/admin/users/{id}/enable` - Re-enable disabled account
- **Change User Role:** `PUT /api/admin/users/{id}/role` - Change between "user" and "admin"
- **Toggle Email Verification:** `POST /api/admin/users/{id}/toggle-email-verification` - Manual verification control
- **Delete User:** `DELETE /api/admin/users/{id}` - Permanently delete user and all data

**Key Implementation Details:**
- All operations restricted to admin role via `middleware.AdminOnly`
- Service layer validates admin permissions before each operation
- Self-modification prevention (admin cannot delete/disable themselves)
- Audit logging for all administrative actions
- Cascading deletes configured for user-related data
- Fixed `List()` method to include all admin-relevant fields

**User Repository Fields:**
- email_verified, email_verified_at - Email verification status
- failed_login_attempts, locked_at, locked_until - Account lockout
- account_disabled, disabled_at, disabled_by_user_id, disable_reason - Permanent disable
- All fields properly handled with NULL-safe SQL queries

**Frontend Integration:**
- `web/src/views/AdminUsersView.vue` - Full admin user management UI
- Dynamic icons showing current state (lock, verification, enabled, role)
- Color-coded states: green (positive), red (negative), purple (admin), blue (user)
- Enhanced tooltips displaying current state explicitly
- Confirmation dialogs for destructive operations (disable, delete)
- Delete dialog shows what will be removed: profile, workouts, PRs, performance history

**Security Features:**
- Admin-only access enforced by middleware
- Authorization checks in service layer
- Self-modification prevention
- Audit trail for accountability
- Confirmation required for destructive actions

### Workout Management (v0.2.0-beta)

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
