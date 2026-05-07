# ActaLog — Application Threat Model

**Version:** v1.3.0
**Last updated:** 2026-04-28
**Owner:** Security / engineering lead

---

## Contents

1. [Overview and scope](#1-overview-and-scope)
2. [Threat actors considered](#2-threat-actors-considered)
3. [Authentication and sessions](#3-authentication-and-sessions)
4. [Rate limiting](#4-rate-limiting)
5. [Input validation](#5-input-validation)
6. [SQL injection prevention](#6-sql-injection-prevention)
7. [File upload security](#7-file-upload-security)
8. [CORS](#8-cors)
9. [Security headers and CSP](#9-security-headers-and-csp)
10. [Password policy](#10-password-policy)
11. [Protected users](#11-protected-users)
12. [Audit logging](#12-audit-logging)
13. [Cross-references](#13-cross-references)

---

## 1. Overview and scope

This document is the entry point for a security review of the ActaLog application. It
maps known attack classes to the mitigations in place, identifies residual risks, and
links directly to the implementation files that implement each control.

**What this document covers:**

- Application-level security controls: authentication, authorization, input handling,
  transport-layer hardening, audit trail, and the protected-user defense system
- The threat actors the application is designed to resist
- Residual risks the team has accepted (acknowledged, not ignored)

**What this document does not cover:**

- Deployment infrastructure hardening (OS, container runtime, firewall rules, TLS
  termination outside the application)
- Database server security (root access, file-system permissions on the DB host)
- Supply-chain attacks against Go module dependencies (out-of-band processes: `go.sum`
  pinning, Dependabot, manual review)
- Physical access to hardware or cloud-provider console access

For a historical OWASP assessment see [`docs/OWASP_AUDIT_2026-04-03.md`](../OWASP_AUDIT_2026-04-03.md).
For the application maturity scorecard see [`docs/MATURITY_ASSESSMENT.md`](../MATURITY_ASSESSMENT.md).

---

## 2. Threat actors considered

| Actor | Description | Primary concern |
|-------|-------------|-----------------|
| **Unauthenticated external attacker** | Anyone who can reach the HTTP port without a valid JWT | Credential stuffing, brute-force, injection via public endpoints |
| **Authenticated regular user** | Holds a valid JWT with `user` role | Horizontal privilege escalation (accessing another user's data), vertical escalation (reaching admin endpoints) |
| **Authenticated admin** | Holds a valid JWT with `admin` role | Accidental or coerced modification of protected system accounts; insider misuse of admin API |
| **Compromised admin account** | Admin credentials obtained by an attacker | Same as above; the protected-user system is specifically designed to bound the damage radius here |
| **Hostile insider with shell access** | Someone who can execute `psql`/`sqlite3`/`mysql` but not necessarily the application binary | Direct SQL DML that bypasses the Go application layer |
| **SQL injection that escapes the Go layer** | A flaw in query construction that sends unescaped user input to the DB engine | Arbitrary data read or write, privilege escalation within the database |

**Out of scope:**

- Database server root / OS-level root on the server host
- Supply-chain compromise of Go modules or npm packages
- Hot-patching a running binary in memory
- Cloud-provider console access

---

## 3. Authentication and sessions

### Threats addressed

- Credential theft via weak secrets or token forgery
- Session fixation and stolen-token replay
- Account takeover via brute-force login

### Mitigations in place

**JWT issuance and validation**

Access tokens are short-lived JWTs signed with HMAC-SHA256. The secret is configurable
via the `JWT_SECRET` environment variable and must be changed from the placeholder
default before production deployment. Validation is enforced on every protected route.

```text
Refs:
  pkg/middleware/auth.go        — Auth middleware, token extraction and validation
  pkg/auth/                     — JWT sign/validate helpers
  internal/service/user_service.go — token issuance on login/register
```

**Refresh tokens**

Long-lived refresh tokens are stored server-side in the `refresh_tokens` table and are
rotated on each use. A client that presents a used (rotated) refresh token causes the
entire family to be invalidated, providing detection of token theft.

```text
Refs:
  internal/service/user_service.go — RefreshToken, rotateRefreshToken
  internal/domain/                  — RefreshTokenRepository interface
```

**Account lockout**

After a configurable number of consecutive failed login attempts (`MAX_LOGIN_ATTEMPTS`,
default 5) the account is hard-locked for a configurable duration (`LOCKOUT_DURATION`).
An `account_locked_auto` audit event is emitted. An admin can unlock the account via the
admin API, which emits `account_unlocked_admin`.

```text
Refs:
  internal/service/user_service.go — Login, LockAccount path (~line 321)
  internal/domain/audit_log.go     — EventAccountLockedAuto, EventAccountUnlockedAdmin
```

**Email verification**

When `REQUIRE_VERIFICATION=true`, new accounts receive a 24-hour verification token via
email before they can log in. The token is a 32-byte random hex string generated with
`crypto/rand`.

### Residual risks

- JWTs are not revocable before expiry (no server-side blocklist). If a token is stolen
  the attacker retains access until expiry. Accepted: the access-token lifetime is short
  (configurable; production should use ≤15 min). Refresh token rotation provides the
  revocation path for longer-term sessions.
- If `JWT_SECRET` is left at the default placeholder in a production deployment the
  token signing key is effectively public. Mitigation: deployment checklist must include
  rotating this value; the default value is documented as a non-secret.

---

## 4. Rate limiting

### Threats addressed

- Brute-force credential stuffing against `/api/auth/login` and `/api/auth/register`
- Abuse of the password-reset email flow (spam/enumeration)

### Mitigations in place

Rate limiting uses an in-memory sliding-window limiter keyed by client IP.

| Endpoint group | Limit | Window |
|----------------|-------|--------|
| `/auth/login`, `/auth/register`, `/auth/resend-verification` | 5 requests | 15 minutes |
| `/auth/forgot-password`, `/auth/reset-password` | 3 requests | 1 hour |

When the limit is exceeded:
- HTTP 429 is returned with `Retry-After` and `X-RateLimit-Limit` headers.
- A `rate_limit_exceeded` audit event is emitted with the client IP.

IP extraction is proxy-aware: `X-Forwarded-For` (leftmost entry), then `X-Real-IP`,
then `RemoteAddr` (with port stripped).

```text
Refs:
  pkg/middleware/rate_limit.go              — RateLimiter, RateLimit middleware
  cmd/actalog/main.go (~line 643-654)       — limiter instantiation and wiring
  internal/domain/audit_log.go              — EventRateLimitExceeded
```

### Residual risks

- The limiter is in-memory and per-process. In a multi-replica deployment each replica
  has an independent counter; a distributed attacker can spread requests across replicas
  to stay under per-replica limits. Accepted: ActaLog is currently a single-binary
  deployment. If horizontal scaling is added, replace with a Redis-backed limiter.
- Rate limiting does not protect API endpoints that require a valid JWT (those are
  protected by the authentication layer instead).

---

## 5. Input validation

### Threats addressed

- Injection attacks via unvalidated user-supplied strings
- Server-side request forgery via unvalidated URL fields
- Business-logic violations from out-of-range or malformed values

### Mitigations in place

Handler-layer validation is the first line of defense. Handlers in `internal/handler/`
reject requests with missing required fields, out-of-range values, and obviously
malformed inputs before the data reaches the service or repository layer.

Parameterized queries (see §6) are the second line for any data that does reach the
database.

```text
Refs:
  internal/handler/           — all handler files; input checks before service calls
  internal/service/user_service.go (~line 101-121) — validatePassword helper
```

### Residual risks

- There is no centralized validation library (e.g., `go-playground/validator`).
  Validation logic is duplicated across handlers. Gaps in individual handlers could
  allow unexpected values to reach the service layer. Mitigations: code-review policy
  and the database constraints that back-stop the service layer.

---

## 6. SQL injection prevention

### Threats addressed

- Arbitrary SQL execution via user-supplied strings interpolated into queries

### Mitigations in place

All database queries use Go's `database/sql` parameterized query interface. User-
controlled values are always passed as bind parameters (`?` for SQLite/MySQL, `$N` for
PostgreSQL), never string-concatenated into query text. This holds across all three
supported drivers.

```go
// Example pattern used throughout the repository layer:
db.QueryContext(ctx, "SELECT * FROM users WHERE email = ?", email)
```

```text
Refs:
  internal/repository/   — all *_repository.go files; QueryContext / ExecContext
                            with bind parameters throughout
```

### Residual risks

- The codebase does not use a query builder or ORM that would statically prevent raw
  query construction. A future developer who introduces string formatting into a query
  (`fmt.Sprintf("... WHERE id = %d", id)`) would reintroduce risk. Mitigation: lint
  rule (`go-sec`, `gosec`) in CI flags SQL string construction patterns.

---

## 7. File upload security

### Threats addressed

- Stored XSS via file type confusion (uploading HTML/SVG/JS disguised as an image)
- Content-type spoofing (client-supplied `Content-Type` header trusted)
- Path traversal via crafted filenames

### Mitigations in place

Avatar uploads in `internal/handler/user_handler.go` use a two-layer file type check
introduced in v1.2.4:

1. **Magic-byte detection** — the first 512 bytes of the uploaded file are read and
   passed to `http.DetectContentType` (WHATWG MIME Sniffing algorithm). The client-
   supplied `Content-Type` header is ignored. If the detected type is not `image/*`,
   the request is rejected.

2. **Extension allowlist** — even if magic bytes pass, the filename extension must be
   one of `.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`. This closes the path where a JPEG
   payload uploaded with a `.html` extension would later be served by the static file
   handler with `text/html` content-type.

Saved filenames are generated deterministically from the user ID and a Unix timestamp
(`user_<id>_<ts><ext>`) — no user-controlled path component reaches the filesystem.

```go
// From internal/handler/user_handler.go (~line 222):
detectedType := http.DetectContentType(sniff[:n])
if !strings.HasPrefix(detectedType, "image/") {
    respondError(w, http.StatusBadRequest, "File must be an image")
    return
}
```

```text
Refs:
  internal/handler/user_handler.go (~line 210-262) — UploadAvatar handler
```

### Residual risks

- `http.DetectContentType` detects MIME type based on the first 512 bytes. A polyglot
  file (e.g., a valid JPEG header followed by script content) would pass this check.
  The extension allowlist is the main guard against serving such files with an
  executable content-type. Mitigation: the static file server uses the file extension to
  set `Content-Type`, so a `.jpg` extension gets `image/jpeg` regardless of payload.
- File size is not currently limited at the Go layer (relying on reverse-proxy or
  OS limits). Accepted for v1.3.0; a `MaxBytesReader` guard should be added.

---

## 8. CORS

### Threats addressed

- Cross-site request forgery (CSRF) via cross-origin API calls from attacker-controlled
  pages
- Origin spoofing (pre-v1.2.3 behavior where any origin was reflected back)

### Mitigations in place

The CORS middleware in `pkg/middleware/cors.go` maintains an explicit allowlist sourced
from the `CORS_ORIGINS` environment variable. A request from an origin not on the
allowlist receives **no** `Access-Control-Allow-Origin` header — the browser's default-
deny behaviour applies. This was a breaking fix in v1.2.3: the previous implementation
echoed every origin back, making the allowlist inert.

Preflight (`OPTIONS`) requests from allowed origins are handled and return HTTP 204.

```go
// From pkg/middleware/cors.go:
if origin != "" && allowed {
    w.Header().Set("Access-Control-Allow-Origin", origin)
}
// Disallowed origins: header is never set — browser rejects the response.
```

```text
Refs:
  pkg/middleware/cors.go         — CORS middleware implementation
  pkg/middleware/cors_test.go    — asserts header absence for disallowed origins
  cmd/actalog/main.go            — r.Use(middleware.CORS(cfg.App.CORSOrigins))
```

### Residual risks

- CORS relies on the browser enforcing the policy. It does not protect non-browser
  clients (curl, scripts). API authentication (JWT) is the control for those callers.
- Setting `CORS_ORIGINS=*` in `.env` would restore wildcard behaviour. Deployment
  policy should prevent this in production.

---

## 9. Security headers and CSP

### Threats addressed

- Clickjacking (framing attacks)
- MIME-type confusion (`nosniff`)
- Cross-site scripting via injected scripts (CSP `script-src 'self'`)
- Protocol downgrade (HSTS)
- Referrer leakage

### Mitigations in place

The `SecurityHeaders` middleware (`pkg/middleware/security_headers.go`) is wired as a
global middleware in `cmd/actalog/main.go` after CORS, so it applies to every response
including API, static assets, and `/health`. It sets the following headers on every
response:

| Header | Value | Threat blocked |
|--------|-------|----------------|
| `X-Frame-Options` | `DENY` | Clickjacking |
| `X-Content-Type-Options` | `nosniff` | MIME confusion |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Referrer leakage |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Protocol downgrade |
| `Content-Security-Policy` | See below | XSS, resource injection |

**CSP policy (v1.3.0):**

```text
default-src 'self';
script-src 'self';
style-src 'self' 'unsafe-inline';
img-src 'self' data: blob:;
connect-src 'self';
font-src 'self';
frame-ancestors 'none';
base-uri 'self';
form-action 'self'
```

`style-src 'unsafe-inline'` is intentional: Vuetify 3 injects inline `<style>` blocks
for dynamic theme switching; removing `'unsafe-inline'` breaks theming. Scripts remain
strict (`script-src 'self'` only). Nonce-based style policy is the recommended future
direction.

```text
Refs:
  pkg/middleware/security_headers.go     — SecurityHeaders middleware + defaultCSP constant
  pkg/middleware/security_headers_test.go
  cmd/actalog/main.go                    — r.Use(middleware.SecurityHeaders)
```

### Residual risks

- HSTS is set by the application but is only effective when the service is behind TLS.
  In a plain-HTTP deployment the header is sent but ignored by browsers. Mitigation:
  TLS must be terminated at the reverse proxy or the binary itself; plain-HTTP
  production deployments are unsupported.
- `style-src 'unsafe-inline'` weakens CSP against style injection attacks. Accepted
  pending nonce support in Vuetify.

---

## 10. Password policy

### Threats addressed

- Weak passwords susceptible to dictionary or brute-force attacks
- Trivially guessable passwords (common words, short strings)

### Mitigations in place

The `validatePassword` helper in `internal/service/user_service.go` enforces:

- Minimum 12 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one digit

This policy is applied at all three password entry points: registration, password reset,
and password change. Password hashing uses bcrypt at cost 12 (`pkg/auth/password.go`
`defaultCost = 12`).

```go
// From internal/service/user_service.go (~line 101):
func validatePassword(password string) error {
    if len(password) < 12 {
        return fmt.Errorf("password must be at least 12 characters")
    }
    // ... uppercase, lowercase, digit checks
}
```

Policy was raised from 8 to 12 characters and complexity requirements added in v1.2.3.

```text
Refs:
  internal/service/user_service.go (~line 101-122) — validatePassword
  pkg/auth/password.go                              — HashPassword (bcrypt cost 12)
```

### Residual risks

- There is no check against common-password lists (e.g., `Have I Been Pwned` corpus or
  a local wordlist). A 12-character password that meets complexity requirements but
  consists of a dictionary word + digit is accepted. Accepted for v1.3.0.
- bcrypt has a maximum effective input length of 72 bytes. Passwords longer than 72
  bytes are silently truncated. No explicit limit is enforced at the application layer.
  Accepted: the policy minimum (12 chars) is well below the truncation threshold.

---

## 11. Protected users

### Threats addressed

- Accidental or coerced modification/deletion of system-critical user accounts by an
  admin operator
- Direct SQL DML from a hostile actor with database credentials
- Rogue migrations that target protected rows
- Silent rollback of the binary removing application-layer guards

### Mitigations in place

The protected-user system implements four independent, overlapping defense layers. Each
layer can independently block a modification attempt; all four must be simultaneously
defeated for an attack to succeed:

| Layer | Location | Mechanism |
|-------|----------|-----------|
| **L1** — HTTP middleware | `pkg/middleware/protected_user.go` | HTTP 403 before the request reaches any handler |
| **L2** — Service guard | `internal/service/admin_user_service.go` | `ensureNotProtected()` check before any write |
| **L3** — Database trigger | `internal/repository/protected_triggers_sql.go` | `BEFORE UPDATE` blocks identity-field changes (email/name/role/account_disabled); `BEFORE DELETE` is unconditional. Survives binary rollback. See PROTECTED_USERS.md §3.4 for the narrowed contract. |
| **L4** — Audit log | `internal/domain/audit_log.go` | One event per block, tagged by the catching layer |

The protected email registry lives in `pkg/security/protected_users.go`. The frontend
guard (`web/src/utils/protectedUsers.js`) is auto-generated from the same source and
disables Edit/Disable/Delete UI actions for protected accounts.

A **boot-time invariant** (`cmd/actalog/main.go` → `internal/protectedusers/boot_invariant.go`)
runs three checks before the HTTP server starts:

```text
Check 1 — triggers exist in the DB catalog    (HARD failure: binary refuses to start)
Check 2 — triggers actually fire correctly    (HARD failure: binary refuses to start)
Check 3 — protected user rows exist in users  (HARD if other users exist; SOFT if fresh install)
```

This ensures that a deployment with missing or broken L3 triggers cannot silently
process writes to protected accounts.

```text
Refs:
  pkg/security/protected_users.go               — IsProtectedEmail registry
  pkg/middleware/protected_user.go              — L1 HTTP guard
  internal/service/admin_user_service.go        — L2 service guard
  internal/repository/protected_triggers_sql.go — L3 trigger SQL constants (all three drivers)
  internal/domain/audit_log.go                  — EventProtectedUserAttackHTTP/Service/DB constants
  docs/security/PROTECTED_USERS.md              — master runbook (add/remove/verify/recover)
  docs/security/PROTECTED_USERS_RECOVERY.md     — 3-AM incident-response playbook
  scripts/recover/README.md                     — last-resort recovery scripts
```

### Residual risks

See [`docs/security/PROTECTED_USERS.md` §2](PROTECTED_USERS.md#2-threat-model) for the
full list of what the system does NOT protect against. Key accepted residuals:

- Root access on the database server (DDL can drop triggers directly)
- Physical disk access / volume snapshot restoration of a pre-trigger backup
- Supply-chain compromise that replaces `IsProtectedEmail` with a no-op
- **Admin compromise → mass user creation / password reset** — Audit log review (`admin_user_created` and `admin_password_set` events per actor over a window). Not blocked at the system level — accepted residual risk for a small-team deployment with admin trust.
- **Shell access on the application host (v1.3.2 break-glass CLI)** — the operator running `actalog admin force-edit-protected` can bypass the entire L1+L2+L3 stack for protected users with a single command. **Accepted** for small-team deployments where shell access already implies DB-credential access via `secrets/local-test-credentials.env` or equivalent. Mitigation: every break-glass operation writes a per-field audit event (`protected_user_break_glass_*`) with operator metadata (`USER`/`hostname`/`tty`/`cwd`) so post-incident review can answer who-when-where.

---

## 12. Audit logging

### Threats addressed

- Undetectable privilege abuse by authenticated admins
- Post-incident forensics gaps (no record of who did what)
- Silent protected-user attack attempts

### Mitigations in place

Every security-relevant action is recorded in the `audit_logs` table. The schema is
defined in `internal/domain/audit_log.go` and populated by `internal/service/audit_log_service.go`.

Each record captures:

```text
id              — auto-increment
user_id         — actor (NULL for system events)
target_user_id  — affected user (NULL if N/A)
event_type      — string constant from audit_log.go
ip_address      — client IP from HTTP request (NULL for internal events)
user_agent      — User-Agent header (NULL for internal events)
details         — JSON string with event-specific fields
created_at      — UTC timestamp
```

Key event categories:

| Category | Example event types |
|----------|---------------------|
| Authentication | `login_success`, `login_failed`, `logout`, `token_refresh` |
| Account security | `account_locked_auto`, `account_unlocked_admin`, `account_disabled` |
| Password | `password_changed`, `password_reset`, `password_reset_forced_by_admin` |
| User management | `user_created`, `user_updated`, `user_deleted`, `role_changed` |
| Protected user attacks | `protected_user_attack_http`, `protected_user_attack_service`, `protected_user_attack_db` |
| Rate limiting | `rate_limit_exceeded` |

**Retention** — `AuditLogRepository.DeleteOlderThan` provides a pruning interface.
Retention policy is configurable at the operator level.

**Protected-user attack events** — a `_db` event means L1 and L2 were both bypassed
and is unconditionally high-severity. See
[`PROTECTED_USERS.md` §9](PROTECTED_USERS.md#9-audit-log-forensics) for query patterns
and alert thresholds.

```text
Refs:
  internal/domain/audit_log.go            — AuditLog struct, all event type constants
  internal/service/audit_log_service.go   — AuditLogService.LogEvent
  internal/repository/                    — audit_log_repository.go (all three DB drivers)
```

### Residual risks

- Audit logs are stored in the same database as application data. An attacker who
  obtains write access to the database can delete audit records before they are read.
  Accepted for v1.3.0; out-of-band log shipping to an immutable store is the
  recommended mitigation for high-security deployments.
- There is no real-time alerting integration. Operators must query or monitor the
  `audit_logs` table. A log aggregator or SIEM integration is the recommended path.

---

## 13. Cross-references

| Document | Purpose |
|----------|---------|
| [`docs/OWASP_AUDIT_2026-04-03.md`](../OWASP_AUDIT_2026-04-03.md) | Full OWASP Top-10 audit history from 2026-04-03 |
| [`docs/MATURITY_ASSESSMENT.md`](../MATURITY_ASSESSMENT.md) | Security maturity scorecard |
| [`docs/security/PROTECTED_USERS.md`](PROTECTED_USERS.md) | Protected-user system master runbook |
| [`docs/security/PROTECTED_USERS_RECOVERY.md`](PROTECTED_USERS_RECOVERY.md) | 3-AM incident-response playbook |
| [`scripts/recover/README.md`](../../scripts/recover/README.md) | Last-resort DB-level recovery scripts |

---

*This is the primary entry point for application security review. For the protected-user
subsystem specifically, start with `docs/security/PROTECTED_USERS.md`.*
