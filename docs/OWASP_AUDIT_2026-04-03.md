# OWASP Top 10 Security Audit — ActaLog v1.2.1

**Date**: 2026-04-03
**Framework**: OWASP Top 10 (2021)
**Auditor**: Claude Code (automated static analysis)
**Deployment Context**: Single-instance, ~3 users

---

## Results Summary

| # | Category | Status | Fixed |
|---|----------|--------|-------|
| A01 | Broken Access Control | ✅ Pass | — |
| A02 | Cryptographic Failures | ✅ Pass | — |
| A03 | Injection | ✅ Pass | — |
| A04 | Insecure Design | ✅ Pass | 2026-04-03 |
| A05 | Security Misconfiguration | ⚠️ Partial | Partial 2026-04-03 |
| A06 | Vulnerable Components | ✅ Pass | 2026-04-03 |
| A07 | Authentication Failures | ✅ Pass | — |
| A08 | Data Integrity / XSS | ✅ Pass | 2026-04-03 |
| A09 | Logging & Monitoring | ✅ Pass | — |
| A10 | SSRF | ✅ Pass | — |

**Original: 3 failures, 1 partial, 6 pass.**
**Current: 0 failures, 1 partial, 9 pass.** (A05 partial — CORS bug and security headers deferred)

### A04 fix (2026-04-03)
Password policy raised to 12-char minimum with uppercase, lowercase, and digit requirements.
Applied in `Register`, `ResetPassword`, and `ChangePassword` (`internal/service/user_service.go`).

### A05 partial fix (2026-04-03)
`WriteError()` internal error leakage fixed (`internal/handler/errors.go`).
CORS `else` branch and security response headers remain deferred (Steps 1 & 2).

### A06 fix (2026-04-03)
`serialize-javascript` pinned to `7.0.5` via `package.json` `overrides`. Fixes RCE
(GHSA-5c6j-r48x-rmvq) and CPU exhaustion DoS (GHSA-qj8w-gfj5-8c6v) without
downgrading `vite-plugin-pwa`.

### A08 fix (2026-04-03)
DOMPurify 3.3.3 added to `MarkdownRenderer.vue`. All `v-html` output is now
sanitized before DOM insertion.

---

## A01: Broken Access Control — PASS with notes

- Workout ownership enforced in service layer: `internal/service/user_workout_service.go:75,296,355` — `ErrUnauthorizedWorkoutAccess` returned if `userID` doesn't match
- `AdminOnly` and `CoachOrAdmin` middleware applied at route level (`pkg/middleware/auth.go`)
- `GetUserID` from JWT context used consistently in all handlers
- `GetLoggedWorkout(id, userID)` passes both resource ID and owner ID — prevents IDOR

**Note:** Integer auto-increment IDs are used (not UUIDs), so workout IDs are enumerable. Protection
relies entirely on the ownership check in the service layer. That check must never be skipped.

---

## A02: Cryptographic Failures — PASS

- bcrypt cost 12 (`pkg/auth/password.go:8`) ✅
- JWT HMAC-SHA256 with algorithm pinning (`pkg/auth/jwt.go:43`) ✅
- Password reset tokens use `crypto/rand` ✅
- `FailedLoginAttempts` excluded from JSON serialization (`json:"-"`) ✅

---

## A03: Injection — PASS

- All SQL uses parameterized queries with `args...` — no user input interpolated
- `fmt.Sprintf` in SQL used only for placeholder generation (`?,?,?`) and driver-specific
  booleans via `getBoolValue()` — verified safe
- No shell exec with user input found
- No template injection risk (Vue auto-escapes by default)

---

## A04: Insecure Design — PARTIAL

**Pass:**
- Rate limiter exists and is applied to auth endpoints ✅
- Account lockout after configurable failed attempts ✅

**Issues:**
- ~~**Weak password policy** — 8-character minimum only, no complexity requirements~~ **Fixed 2026-04-03**: raised to 12-char minimum with uppercase, lowercase, and digit requirements; UI hints updated in all four password-entry views
- **Unresolved authorization TODO** — `internal/service/workout_service.go:396`:
  `"TODO: Add proper authorization through workout template ownership"` — acknowledged gap
- **Subscription expiry not enforced** — `internal/service/subscription_service.go:487`:
  `"TODO: Implement background job to check and expire subscriptions"` — expired subscriptions
  remain active until manually cancelled

---

## A05: Security Misconfiguration — FAIL

### CORS enforcement disabled

`pkg/middleware/cors.go:31-33`: The `else` branch for disallowed origins also sets
`Access-Control-Allow-Origin` to the requesting origin. The allowlist has no effect.

```go
// Current broken code:
if allowed {
    w.Header().Set("Access-Control-Allow-Origin", origin)
} else {
    // "For now, allow all origins in development" ← ships to production
    w.Header().Set("Access-Control-Allow-Origin", origin)
}
```

Fix: remove the `else` branch entirely.

### No security response headers

The following headers are entirely absent from all responses:

| Header | Purpose |
|--------|---------|
| `Content-Security-Policy` | Prevents XSS, controls resource loading |
| `X-Frame-Options: DENY` | Prevents clickjacking |
| `X-Content-Type-Options: nosniff` | Prevents MIME sniffing attacks |
| `Strict-Transport-Security` | Enforces HTTPS |
| `Referrer-Policy` | Controls referrer information leakage |

### Internal error messages exposed

`internal/handler/errors.go:129`: `WriteError()` passes `err.Error()` directly to HTTP responses
for all errors. `HandleServiceError()` at line 165 correctly sanitizes unknowns but `WriteError()`
does not — inconsistent behavior depending on which helper a handler calls.

### Avatar upload — Content-Type spoofing

`internal/handler/user_handler.go:220-223`: File type validation checks the `Content-Type` header
sent by the client, not the actual file magic bytes. A malicious file with
`Content-Type: image/jpeg` bypasses this check.

---

## A06: Vulnerable Components — FAIL

### npm (frontend) — 4 HIGH severity

```
serialize-javascript — RCE via RegExp.flags / Date.prototype.toISOString()
  GHSA-5c6j-r48x-rmvq
serialize-javascript — CPU exhaustion DoS via crafted array-like objects
  GHSA-qj8w-gfj5-8c6v

Dependency chain:
  vite-plugin-pwa >= 0.20.0
    → workbox-build
      → @rollup/plugin-terser 0.2.0-0.4.4
        → serialize-javascript (vulnerable)

Fix: npm audit fix --force
  (installs vite-plugin-pwa@0.19.8 — breaking change, test PWA install flow after)
```

### Go dependencies — PASS

All Go dependencies updated to latest in v1.2.1 per CHANGELOG.

---

## A07: Authentication Failures — PASS with notes

- Refresh token revocation implemented ✅
- Account lockout after configurable failed attempts ✅
- `ForgotPassword` always returns success — prevents email enumeration ✅
- JWT expiration enforced on validation ✅
- No MFA (acceptable for 3-user personal deployment)

---

## A08: Data Integrity Failures / XSS — FAIL

### Unsanitized markdown rendering (stored XSS vector)

`web/src/components/MarkdownRenderer.vue:25-33` uses `v-html` with `marked.parse()` output but
**no HTML sanitization**. `marked` is a parser, not a sanitizer — it faithfully renders `<script>`
tags and `javascript:` href links embedded in markdown.

```js
// Fixed 2026-04-03 — DOMPurify 3.3.3 added:
const renderedHtml = computed(() => {
  if (!props.content) return ''
  try {
    return DOMPurify.sanitize(marked.parse(props.content))
  } catch (error) {
    return DOMPurify.sanitize(props.content)
  }
})
```

This component is used in 9 views:
- `web/src/views/DashboardView.vue`
- `web/src/views/NotificationsView.vue`
- `web/src/views/AdminAnnouncementsView.vue`
- `web/src/views/WODDetailView.vue`
- `web/src/views/WODsView.vue`
- `web/src/views/WorkoutDetailView.vue`
- `web/src/views/MovementDetailView.vue`
- `web/src/components/SessionDetailDialog.vue`
- `web/src/components/DetailViewDialog.vue`

Any user-supplied content rendered through this component (WOD descriptions, announcements,
notifications) is a stored XSS vector.

**Fix:**
```bash
npm install dompurify
```
```js
import DOMPurify from 'dompurify'

const renderedHtml = computed(() => {
  if (!props.content) return ''
  try {
    return DOMPurify.sanitize(marked.parse(props.content))
  } catch (error) {
    console.error('Failed to parse markdown:', error)
    return DOMPurify.sanitize(props.content)
  }
})
```

---

## A09: Logging & Monitoring — PASS with notes

- Comprehensive structured logging with IP, UA, action, outcome on all auth events ✅
- 80+ typed audit event constants covering all sensitive operations ✅
- Log rotation implemented in `pkg/logger/logger.go` ✅
- No automated monitoring/alerting — accepted risk at current deployment scale (3 users, manual review feasible)

---

## A10: SSRF — PASS

No outbound HTTP calls with user-supplied URLs found in application code. Email is sent via
configured SMTP server (no user-controlled URL). No webhook or proxy features detected.

---

## Prioritised Fix List

| Priority | Issue | File | Effort |
|----------|-------|------|--------|
| 1 | `npm audit fix --force` — 4 high-severity serialize-javascript vulns | `web/` | 30 min |
| 2 | Add DOMPurify to MarkdownRenderer — stored XSS | `web/src/components/MarkdownRenderer.vue` | 30 min |
| 3 | Fix CORS `else` branch — allowlist has no effect | `pkg/middleware/cors.go:31-33` | 15 min |
| 4 | Add security headers middleware | New `pkg/middleware/security_headers.go` | 2 hrs |
| 5 | Fix avatar upload — validate magic bytes not client Content-Type | `internal/handler/user_handler.go:220` | 1 hr |
| 6 | Sanitize `WriteError()` — don't expose `err.Error()` for unknown errors | `internal/handler/errors.go:129` | 30 min |
| 7 | Strengthen password policy — minimum 12 chars + complexity | `internal/service/user_service.go:111` | 1 hr |

---

## Resolution status (added 2026-04-28)

This audit captured the state of `main` on 2026-04-03. The remediation work it kicked off is now complete:

- **v1.2.3** (2026-04-03) — closed CORS allowlist, error sanitization, DOMPurify, serialize-javascript, X-Forwarded-For IP, password policy, rate-limit audit event
- **v1.2.4** (2026-04-28) — closed security response headers middleware and avatar upload magic-byte validation

See `docs/plans/SECURITY_HARDENING_PLAN.md` for the per-step status table, and `docs/CHANGELOG.md` for the per-release Security sections. The audit body above is intentionally preserved as the dated point-in-time snapshot.
