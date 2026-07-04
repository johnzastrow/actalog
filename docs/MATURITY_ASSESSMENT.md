# ActaLog — Code Maturity Assessment

**Framework**: Trail of Bits Code Maturity Evaluation v0.1.0 (adapted for web application)
**Date**: 2026-04-03 (base assessment) · **Last reviewed**: 2026-07-04 (v1.3.3)
**Version Assessed**: 1.2.1 (base) — see "Remediation status" below for changes through 1.3.3
**Deployment Context**: Single-instance, ~3 users

---

## Remediation status — updated 2026-07-04 (v1.3.3)

The base assessment below reflects v1.2.1 (2026-04-03). Several of its top gaps have since
been closed; this section tracks the delta so the scorecard is not read as current state.

**Resolved since the base assessment:**
- ✅ **CORS enforcement** — allowlist now actually enforced (v1.2.3).
- ✅ **Security response headers** — CSP, HSTS, X-Frame-Options, X-Content-Type-Options, Referrer-Policy middleware added (v1.2.4).
- ✅ **Password policy** — 12-char minimum + complexity (v1.2.3).
- ✅ **File upload hardening** — avatar uploads validated by magic bytes, not client `Content-Type` (v1.2.4).
- ✅ **CI toolchain drift** — Node 24 + npm pinned to 11.16.0 in CI and Docker so `npm ci` stays in lockstep with the lockfile (v1.3.3).

**In progress (this release — CI/supply-chain hardening):**
- 🔶 Go lint regressions now blocking on PRs (`only-new-issues`); `govulncheck`, `npm audit`, and a Trivy image scan added as blocking CI gates; GitHub Actions pinned to commit SHAs and Docker base images pinned to digests.

**Still open (tracked, not yet addressed):**
- ⬜ **Repository-layer `context.Context`** — 0 of 373 repo functions accept a context, so DB calls have no timeout/cancellation (`noctx` lint disabled with 851 violations noted). Needs a dedicated refactor branch.
- ⬜ **Single JWT signing secret** — no key rotation (`kid` + multi-key) path.
- ⬜ **Large files** — `backup_service.go` (2.2k), `scheduling_handler.go` (1.8k), `data_quality_service.go` (1.5k) remain decomposition candidates.
- ⬜ **Accumulated lint debt** — 1,803 golangci-lint findings (errcheck 1045, revive 445, gosec 142, …) grandfathered; blocked from growing via `only-new-issues`.

---

## Executive Summary

**Overall Maturity Score: 2.8 / 4.0 (Moderate–Satisfactory)**

ActaLog is a well-structured, production-deployed CrossFit tracking PWA with strong architectural
discipline, impressive audit trail coverage, and solid test coverage for a project of its age. The
main gaps are concentrated in security headers, CORS enforcement, password policy, and documentation
drift.

**Top 3 Strengths:**
1. Comprehensive audit event taxonomy — 80+ typed events, full CRUD coverage across all entities
2. Strong CI matrix — unit + integration tests against SQLite, PostgreSQL, and MariaDB on every push
3. Clean Architecture with zero cross-layer violations — domain layer has no dependencies

**Top 3 Critical Gaps (at time of assessment — 2026-04-03):**
1. CORS enforcement is intentionally disabled (`pkg/middleware/cors.go:31-33` — allows any origin even when not in the allowed list) *(deferred)*
2. No security response headers (no CSP, HSTS, X-Frame-Options, X-Content-Type-Options) *(deferred)*
3. Password minimum was only 8 characters with no complexity requirements — **fixed 2026-04-03** (raised to 12 chars + complexity)

---

## Maturity Scorecard

| # | Category | Rating | Score | Key Finding |
|---|----------|--------|-------|-------------|
| 1 | Arithmetic | Satisfactory | 3 | 1RM formulas well-tested; guard clauses on all zero inputs |
| 2 | Auditing | Satisfactory | 3 | 80+ typed events, full service integration; no monitoring/alerting defined |
| 3 | Auth / Access Controls | Moderate | 2 | Lockout implemented; CORS broken; no security headers; password policy fixed 2026-04-03 |
| 4 | Complexity Management | Moderate | 2 | Clean Architecture excellent; several files >1000 lines; lint upgraded to v2 2026-04-03 |
| 5 | Decentralization (Single Points of Failure) | Moderate | 2 | Single JWT secret, admin-only unlocking, no secret rotation docs |
| 6 | Documentation | Satisfactory | 3 | Extensive docs; testing doc version drift fixed 2026-04-03; swagger present |
| 7 | Transaction Ordering / Concurrency Risks | Moderate | 2 | Rate limiter in-memory only (not distributed); no request deduplication |
| 8 | Low-Level / Unsafe Code | Strong | 4 | Only `syscall` for signal handling; all SQL uses parameterized queries |
| 9 | Testing & Verification | Satisfactory | 3 | 81.6% service coverage; CI matrix across 3 DBs; no fuzz tests; frontend tests added to CI 2026-04-03 |

---

## Detailed Analysis

### 1. ARITHMETIC — Satisfactory (3/4)

**Evidence:**
- `pkg/prmath/one_rm.go`: Hybrid Epley/Wathan formula selection by rep range with named formula constants
- `pkg/prmath/one_rm_test.go`: Table-driven tests covering all formula branches, zero-input guards, and boundary at reps=1/10/11
- Guard clauses: `if weight <= 0 || reps <= 0 { return 0, "" }` prevents division-by-zero
- `CalculateAllFormulas`: Brzycki protected against `reps >= 37` which would produce division-by-zero (`36/(37-reps)`)
- `CompareToBaseline`: Zero-baseline guard present

**Gaps:**
- No overflow protection for extremely large weight values (float64 is safe in practice but undocumented)
- Repository arithmetic (counting queries, percentages in `benchmark_service.go`) is untested at boundary conditions
- No precision rounding documented — floating point results passed directly to users without rounding spec

---

### 2. AUDITING — Satisfactory (3/4)

**Evidence:**
- `internal/domain/audit_log.go`: ~80 typed event constants spanning auth, CRUD, scheduling, subscriptions
- `internal/service/audit_log_service.go`: Dedicated service with `LogEvent()`, `LogLoginSuccess()`, etc.
- Auth handler (`internal/handler/auth_handler.go:82-184`): Login attempts, failures, and outcomes all logged with IP+UA
- AuditLog injected as dependency in `movement_service`, `wod_service`, `subscription_service`, `organization_service`
- `AuditLogFilters` supports date range, user, event type — good for incident investigation
- `audit_log_service.go` at 100% test coverage

**Gaps:**
- No alerting, dashboards, or automated anomaly detection (accepted risk at current deployment scale)
- `notification_service.go` only 82.6% covered — notification audit events may have gaps
- No documented retention policy, though `DeleteOlderThan` exists in the repository
- `rate_limit_exceeded` event constant defined in domain but not verified to be emitted by the rate limiter middleware

---

### 3. AUTHENTICATION / ACCESS CONTROLS — Moderate (2/4)

**Strengths:**
- `pkg/auth/password.go`: bcrypt at cost 12 ✅
- `pkg/auth/jwt.go`: HMAC-SHA256, explicit algorithm check (`if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok`)
- `pkg/middleware/auth.go`: `AdminOnly` and `CoachOrAdmin` middleware for role separation
- Account lockout: `internal/service/user_service.go:297` — locks after `maxLoginAttempts` failures with configurable duration
- `FailedLoginAttempts` field not exposed in JSON responses (`json:"-"`)
- Refresh token revocation implemented (`RevokeToken` handler)
- Forgot password always returns success regardless of email existence (prevents user enumeration)

**Critical Issues:**

**CORS disabled** (`pkg/middleware/cors.go:31-33`):
The `else` branch for disallowed origins *also* sets `Access-Control-Allow-Origin` to the
requesting origin. This nullifies the allowed origins list entirely. The code correctly detects
whether an origin is allowed, then ignores its own result:

```go
// pkg/middleware/cors.go:25-33
// Always set CORS headers (for debugging - will improve later)
if origin != "" {
    if allowed {
        w.Header().Set("Access-Control-Allow-Origin", origin)
    } else {
        // For now, allow all origins in development  ← THIS SHIPS TO PRODUCTION
        w.Header().Set("Access-Control-Allow-Origin", origin)
    }
}
```

**No security response headers**: Missing `Content-Security-Policy`, `X-Frame-Options`,
`X-Content-Type-Options`, `Strict-Transport-Security`, `Referrer-Policy`. These are
browser-enforced protections against XSS and clickjacking regardless of user count.

**Weak password policy** *(fixed 2026-04-03)*: Was 8-character minimum only.
Now enforces 12-character minimum with uppercase, lowercase, and digit requirements in
`Register`, `ResetPassword`, and `ChangePassword`. UI hints updated across all four views.

**X-Forwarded-For unsanitized** *(fixed 2026-04-03)*: Was taking the full header value
without splitting on comma. Now takes the leftmost (original client) IP from
`X-Forwarded-For` and strips the port from `RemoteAddr` via `net.SplitHostPort`
(`pkg/middleware/rate_limit.go`).

---

### 4. COMPLEXITY MANAGEMENT — Moderate (2/4)

**Strengths:**
- Clean Architecture: domain → service → handler with zero upward dependencies
- 21 domain entities, each in its own file
- Handler structs inject only what they need (constructor injection)
- `internal/service/test_helpers.go` and `internal/handler/mocks_test.go` centralize test infrastructure

**Issues:**
- `cmd/actalog/main.go`: **1,118 lines** — route registration, middleware stack, static file serving, migration runner, and scheduler startup all in one file
- `internal/service/backup_service.go`: **2,249 lines** / 38 functions
- `internal/service/scheduling_service.go`: **1,450 lines** / 55 functions
- `internal/service/data_quality_service.go`: **1,554 lines**
- `golangci-lint` runs with `continue-on-error: true` in CI (`.github/workflows/ci.yml:35`) — lint failures are silently ignored and never block merges
- `interface{}` / `any` used extensively in audit log details — loses type safety in structured event data

---

### 5. DECENTRALIZATION / SINGLE POINTS OF FAILURE — Moderate (2/4)

*(For a web app: single-point admin controls, key/secret management, centralized state, upgrade risk.)*

**Strengths:**
- Database abstraction allows swapping SQLite/Postgres/MySQL without code changes
- Docker multi-database testing matrix validates portability
- `db_versions/` snapshots enable rollback testing

**Issues:**
- **Single JWT secret**: Compromise of `JWT_SECRET` invalidates all active sessions — no key rotation mechanism or multiple-key support documented
- **Admin account bootstrap**: First registered user becomes admin with no documented recovery procedure if that account is compromised or deleted
- **In-memory rate limiter**: State lost on restart; acceptable for single instance but fragile
- **Backup service**: `backup_service.go` stores backups locally — no documented offsite backup strategy
- **No secret rotation runbook**: `JWT_SECRET`, `DB_PASSWORD`, `SMTP_PASSWORD` have no rotation procedure in docs

---

### 6. DOCUMENTATION — Satisfactory (3/4)

**Strengths:**
- `docs/` contains: ARCHITECTURE, CHANGELOG, DATABASE_SCHEMA, DEPLOYMENT, DOCKER, REQUIREMENTS, TESTING, USER_PERMISSIONS, ROADMAP, LOGGING, GDPR compliance docs
- Swagger/OpenAPI generated from handler annotations (`docs/swagger.yaml`, `docs/swagger.json`)
- `CLAUDE.md` has excellent operational runbook (port cleanup, DB testing matrix, credential scanning)
- `CHANGELOG.md` follows Keep-a-Changelog format with dedicated Security section (supply chain incident documented in v1.2.1)
- Architecture diagram uses Mermaid (machine-readable, renderable in GitHub)

**Gaps:**
- `docs/TESTING.md` says version `0.17.0-beta` but current codebase is `1.2.1` — significant version drift in the key testing document
- Coverage targets stated in TESTING.md (`>90%` for service layer) are not met for `user_service.go` (61.9%), `import_service.go` (60.8%) with no tracking issue
- No incident response playbook (what to do when JWT secret is compromised, account breach, etc.)
- API docs not verified to be in sync with actual routes — no CI step validates swagger accuracy against live routes

---

### 7. TRANSACTION ORDERING / CONCURRENCY RISKS — Moderate (2/4)

*(Adapted from MEV: covers race conditions, idempotency, concurrent request handling.)*

**Strengths:**
- Rate limiter uses `sync.RWMutex` correctly (`pkg/middleware/rate_limit.go:33`)
- Logger uses mutex for concurrent file writes (`pkg/logger/logger.go:45`)
- `go test -race` available via `make test`

**Issues:**
- **No request idempotency keys** for mutations (POST workout, POST reservation) — duplicate submissions possible on network retry, though low risk at 3-user scale
- `consistency_repository.go` runs health checks but no distributed locking for concurrent migrations (benign at single-instance scale)
- `user_workout_repository.go` UPSERT-style operations not verified to be atomic across all 3 DB backends in tests

*Note: Distributed rate limiting and scheduler deduplication gaps are not applicable for single-instance deployment.*

---

### 8. LOW-LEVEL / UNSAFE CODE — Strong (4/4)

**Evidence:**
- Only `syscall` usage is `signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)` — standard graceful shutdown pattern
- No `unsafe` package usage anywhere in application code
- All SQL uses parameterized queries (`db.Query(query, args...)`, `db.Exec(query, args...)`)
- `fmt.Sprintf` in SQL is used only for: placeholder generation (`?,?,?`), driver-specific boolean expressions via `getBoolValue()`, and static table names — never for user input
- The `whereClause` builder in `internal/repository/user_repository.go:699` appends string constants (e.g., `" AND name LIKE ?"`) and passes values via the `args` slice — correct parameterized pattern
- The `DELETE IN (?,?,?)` pattern in `internal/repository/class_session_repository.go:700` builds placeholder count from `[]int64` length — correct batch delete pattern
- Multi-DB placeholder rebinding via `rebindQuery()` correctly handles `?` vs `$1` difference
- No CGO in application code (sqlite3 driver requires it, but that is a dependency, not owned code)

---

### 9. TESTING & VERIFICATION — Satisfactory (3/4)

**Strengths:**
- **81.6% overall service coverage** documented in `docs/TESTING.md`
- 12 services at **100% coverage**, 3 more at >80%
- CI runs unit tests + integration tests against all 3 DB backends (SQLite, PostgreSQL, MariaDB) on every push and PR
- `ci-failure-notify.yml` auto-creates GitHub issues on CI failure — good operational hygiene
- 54 frontend test files (Vue components tested with Vitest)
- Handler-layer tests present (74 test files in `internal/handler/`)
- `internal/service/test_helpers.go` (2,443 lines) — substantial mock infrastructure

**Gaps:**
- **Lint gate** *(partially fixed 2026-04-03)*: golangci-lint upgraded to v2.11.4; `continue-on-error: true` remains (851 `noctx` violations require architectural refactor before enabling hard gate)
- **No fuzz testing** despite arithmetic operations that would benefit (1RM with extreme inputs)
- **No coverage enforcement** in CI — coverage can silently drop without failing the build
- `user_service.go` at **61.9%** — the most security-critical service has the lowest coverage
- **Frontend tests added to CI 2026-04-03** — `web-test` job now runs `npm run test:run` on every push (was build-only)
- Repository layer coverage listed as "pending" in TESTING.md
- No contract/schema validation tests for REST API responses

---

## Improvement Roadmap

### CRITICAL — Fix before next release (30 min – 2 hrs each)

| # | Issue | Location | Status |
|---|-------|----------|--------|
| C1 | Fix CORS middleware — remove `else` branch that grants any origin | `pkg/middleware/cors.go:31-33` | Deferred |
| C2 | Add security response headers middleware (CSP, X-Frame-Options, HSTS, X-Content-Type-Options) | New `pkg/middleware/security_headers.go` | Deferred |
| C3 | Fix X-Forwarded-For IP parsing — split on comma | `pkg/middleware/rate_limit.go` | ✅ Fixed 2026-04-03 |

### HIGH — Within 1–2 months

| # | Issue | Location | Status |
|---|-------|----------|--------|
| H1 | Strengthen password policy to ≥12 chars + complexity | `internal/service/user_service.go` | ✅ Fixed 2026-04-03 |
| H2 | Enable golangci-lint as hard gate (remove `continue-on-error`) | `.github/workflows/ci.yml` | Deferred (851 noctx violations) |
| H3 | Add frontend tests to CI (`npm run test:run`) | `.github/workflows/ci.yml` | ✅ Fixed 2026-04-03 |
| H4 | Increase `user_service.go` coverage from 61.9% to ≥85% | `internal/service/user_service_test.go` | Open |
| H5 | Document JWT secret rotation runbook | `docs/DEPLOYMENT.md` | Open |
| H6 | Update `docs/TESTING.md` version from `0.17.0-beta` to `1.2.1` | `docs/TESTING.md` | ✅ Fixed 2026-04-03 |

### MEDIUM — Within 2–4 months

| # | Issue | Location | Effort |
|---|-------|----------|--------|
| M1 | Break up `cmd/actalog/main.go` (1,118 lines) into route sub-packages | `cmd/actalog/` | 1 week |
| M2 | Add coverage threshold enforcement to CI | CI pipeline | 3 hrs |
| M3 | Add fuzz tests for 1RM arithmetic edge cases | `pkg/prmath/one_rm_test.go` | 4 hrs |
| M4 | Document offsite backup strategy | `docs/DEPLOYMENT.md` | 2 hrs |
| M5 | Add swagger accuracy validation step to CI | `.github/workflows/ci.yml` | 3 hrs |

### LOW / NOT APPLICABLE at current scale

| # | Item | Reason |
|---|------|--------|
| - | Distributed rate limiting | Single instance — in-memory is sufficient |
| - | Idempotency keys | 3 users, low concurrent write probability |
| - | Scheduler deduplication | Single process, no risk |
| - | Incident response playbook | Manual audit log review is feasible at this scale |
| - | Fuzz testing (high-pri) | No external attackers supplying arithmetic inputs |

---

## Notes on Methodology

This assessment used the Trail of Bits 9-category framework adapted for a traditional web
application. Categories originally targeting smart contract / DeFi concerns (MEV, decentralization,
low-level EVM manipulation) were reinterpreted as:

- **MEV / Transaction Ordering** → concurrency, race conditions, request idempotency
- **Decentralization** → single points of failure, key management, admin bootstrap
- **Low-Level Manipulation** → unsafe Go code, raw SQL construction, CGO usage

Ratings follow the Trail of Bits scale:
- **Missing (0)**: Not present
- **Weak (1)**: Several significant improvements needed
- **Moderate (2)**: Adequate, can be improved
- **Satisfactory (3)**: Above average, minor improvements possible
- **Strong (4)**: Exceptional

---

## Resolution status (added 2026-04-28)

This scorecard reflects `main` on 2026-04-03. The "Auth Lifecycle" Moderate (2) rating was driven primarily by password policy and CORS issues; both shipped fixes in v1.2.3, and the response-header coverage gap closed in v1.2.4. The body above is preserved as a dated snapshot — re-run the assessment after the next major batch of changes for an updated rating rather than editing this document in place.
