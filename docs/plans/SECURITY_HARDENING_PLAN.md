# Security Hardening Plan — v1.2.2

**Branch**: `security/hardening-2026-04-03`
**Source documents**: `docs/OWASP_AUDIT_2026-04-03.md`, `docs/MATURITY_ASSESSMENT.md`
**Target**: All OWASP failures resolved, maturity Auth category raised from Moderate (2) to Satisfactory (3)

---

## Overview

Two audits run on 2026-04-03 identified the same cluster of issues from different angles.
Cross-referencing them, every item below appears in at least one audit. Items are ordered
so that each step is independently testable and does not block the next.

---

## Status (as of 2026-04-28)

| Step | Title | Shipped in |
|------|-------|-----------|
| 1 | Fix CORS enforcement | **v1.2.3** |
| 2 | Security response headers middleware | **v1.2.4** |
| 3 | Fix npm vulnerabilities (serialize-javascript) | **v1.2.3** |
| 4 | Add DOMPurify to MarkdownRenderer | **v1.2.3** |
| 5 | Avatar upload file type validation | **v1.2.4** |
| 6 | WriteError internal detail exposure | **v1.2.3** |
| 7 | X-Forwarded-For IP parsing | **v1.2.3** |
| 8 | Strengthen password policy | **v1.2.3** |
| 9 | Wire `rate_limit_exceeded` audit event | **v1.2.3** |

All steps complete. Plan retained as a reference for the audit→remediation pattern; future security work should produce its own dated plan.

---

## Step 1 — Fix CORS enforcement

**Source**: OWASP A05, Maturity C1
**File**: `pkg/middleware/cors.go:31-33`
**Effort**: 15 min

The `else` branch in the origin check sets `Access-Control-Allow-Origin` to the requesting
origin even when it is not in the allowlist, making the allowlist inert. Remove the `else`
branch. When an origin is not allowed, set no `Access-Control-Allow-Origin` header at all.

```go
// Before (broken):
if allowed {
    w.Header().Set("Access-Control-Allow-Origin", origin)
} else {
    w.Header().Set("Access-Control-Allow-Origin", origin) // ← delete this branch
}

// After:
if allowed {
    w.Header().Set("Access-Control-Allow-Origin", origin)
}
```

**Test**: Existing `cors_test.go` has `TestCORS_MultipleAllowedOrigins` — add a case that
asserts a disallowed origin receives NO `Access-Control-Allow-Origin` header.

---

## Step 2 — Add security response headers middleware

**Source**: OWASP A05, Maturity C2
**File**: New `pkg/middleware/security_headers.go`, wire into `cmd/actalog/main.go`
**Effort**: 2 hrs

Create a new middleware that sets the following headers on every response:

| Header | Value |
|--------|-------|
| `X-Frame-Options` | `DENY` |
| `X-Content-Type-Options` | `nosniff` |
| `Referrer-Policy` | `strict-origin-when-cross-origin` |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` |
| `Content-Security-Policy` | See note below |

**CSP note**: Start permissive and tighten. A starting value that won't break the Vue PWA:
```
default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self'; font-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'
```
`unsafe-inline` for styles is needed for Vuetify's dynamic styles. Do not add `unsafe-eval`
or `unsafe-inline` for scripts.

Wire the middleware in `cmd/actalog/main.go` after the CORS middleware, before routes.

**Test**: Add `security_headers_test.go` that creates a test handler, passes a request through
the middleware, and asserts each header is present with the correct value.

---

## Step 3 — Fix npm vulnerabilities (serialize-javascript)

**Source**: OWASP A06
**Directory**: `web/`
**Effort**: 30 min

```bash
cd web
npm audit fix --force   # downgrades vite-plugin-pwa to 0.19.8
npm run build           # verify build still succeeds
npm run test:run        # verify tests still pass
```

After running, verify the PWA install prompt still appears in a browser test. The PWA
service worker is the main thing to check — `vite-plugin-pwa@0.19.8` has a different
workbox config API.

---

## Step 4 — Add DOMPurify to MarkdownRenderer

**Source**: OWASP A08
**File**: `web/src/components/MarkdownRenderer.vue`
**Effort**: 30 min

```bash
cd web && npm install dompurify
```

Update `MarkdownRenderer.vue` to sanitize the `marked.parse()` output before assigning
to `v-html`. The fallback plain-text path must also be sanitized.

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

**Test**: Add a unit test to `MarkdownRenderer` (or its existing test file if one exists)
that passes `<script>alert(1)</script>` as `content` and asserts the rendered output
contains no `<script>` tag.

---

## Step 5 — Fix avatar upload file type validation

**Source**: OWASP A05
**File**: `internal/handler/user_handler.go:220-223`
**Effort**: 1 hr

Replace `Content-Type` header check with magic bytes detection. Read the first 512 bytes
of the uploaded file and use `http.DetectContentType()` to determine the real MIME type.

```go
// Read first 512 bytes for magic byte detection
buf := make([]byte, 512)
n, err := file.Read(buf)
if err != nil && err != io.EOF {
    respondError(w, http.StatusBadRequest, "Failed to read file")
    return
}
detectedType := http.DetectContentType(buf[:n])
if !strings.HasPrefix(detectedType, "image/") {
    respondError(w, http.StatusBadRequest, "File must be an image")
    return
}
// Seek back to start before copying
if _, err := file.Seek(0, io.SeekStart); err != nil {
    respondError(w, http.StatusInternalServerError, "Failed to process file")
    return
}
```

Also validate the extension against an allowlist:
```go
allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
ext := strings.ToLower(filepath.Ext(header.Filename))
if !allowedExts[ext] {
    respondError(w, http.StatusBadRequest, "Only jpg, png, gif, and webp images are allowed")
    return
}
```

---

## Step 6 — Fix WriteError internal detail exposure

**Source**: OWASP A05, Maturity A05
**File**: `internal/handler/errors.go:117-130`
**Effort**: 30 min

`WriteError()` currently calls `err.Error()` directly, which can expose internal details
for unknown errors. Make it consistent with `HandleServiceError()`:

```go
func WriteError(w http.ResponseWriter, err error) {
    status, known := MapServiceError(err)
    if !known {
        var httpErr *HTTPError
        if errors.As(err, &httpErr) {
            status = httpErr.Status
            // HTTPError.Message is already a safe, human-written string
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(status)
            json.NewEncoder(w).Encode(ErrorResponse{Message: httpErr.Message})
            return
        }
        // Unknown error — do not expose internal detail
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "an internal error occurred"})
        return
    }
    // Known service error — safe to use the sentinel error message
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
}
```

---

## Step 7 — Fix X-Forwarded-For IP parsing

**Source**: Maturity C3
**File**: `pkg/middleware/rate_limit.go:149-153`
**Effort**: 30 min

The current code takes the entire `X-Forwarded-For` value, which can contain a
comma-separated list. A client can prepend a spoofed IP to bypass rate limiting.
When not behind a trusted proxy, use `RemoteAddr` only. When behind a proxy, take
the *last* value in the `X-Forwarded-For` chain (set by the proxy, not the client).

For this single-instance deployment without a reverse proxy, the simplest safe fix is:

```go
func getIP(r *http.Request) string {
    // X-Real-IP is set by trusted proxies (nginx, etc.) and is not client-spoofable
    if xri := r.Header.Get("X-Real-IP"); xri != "" {
        return strings.TrimSpace(xri)
    }
    // Fall back to RemoteAddr — always trustworthy
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr
    }
    return ip
}
```

Remove the `X-Forwarded-For` path entirely unless a reverse proxy is added in future.
Update `rate_limit_test.go` to reflect the new behaviour.

---

## Step 8 — Strengthen password policy

**Source**: OWASP A04, Maturity H1
**File**: `internal/service/user_service.go:111`
**Effort**: 1 hr

Raise the minimum to 12 characters and add basic complexity checks. Apply the same
rules in `ResetPassword` (`auth_handler.go:279`) and any admin password-set paths.

```go
func validatePassword(password string) error {
    if len(password) < 12 {
        return errors.New("password must be at least 12 characters")
    }
    var hasUpper, hasLower, hasDigit bool
    for _, c := range password {
        switch {
        case unicode.IsUpper(c):
            hasUpper = true
        case unicode.IsLower(c):
            hasLower = true
        case unicode.IsDigit(c):
            hasDigit = true
        }
    }
    if !hasUpper || !hasLower || !hasDigit {
        return errors.New("password must contain uppercase, lowercase, and a number")
    }
    return nil
}
```

**Note**: This is a breaking change for existing users only at next password reset/change.
Existing stored hashes are unaffected. Document the new requirement in the UI.

---

## Step 9 — Wire `rate_limit_exceeded` audit event

**Source**: Maturity Auditing gap
**File**: `pkg/middleware/rate_limit.go`, requires passing `auditLogService` into middleware
**Effort**: 1 hr

The `EventRateLimitExceeded` constant is defined in `internal/domain/audit_log.go` but
never emitted. This is lower priority than steps 1–8 but completes the audit coverage.

Because the rate limiter lives in `pkg/` (no service dependencies), the cleanest approach
is to accept an optional callback:

```go
type RateLimiter struct {
    // ...existing fields...
    onExceeded func(ip string) // optional hook, called when rate limit is hit
}
```

Wire from `main.go` to call `auditLogService.LogEvent(EventRateLimitExceeded, ...)`.

---

## Step 10 — Add frontend tests to CI and fix lint gate

**Source**: Maturity H2, H3
**File**: `.github/workflows/ci.yml`
**Effort**: 2 hrs

Two CI fixes in one step:

1. **Add frontend test job**:
```yaml
web-test:
  name: "Web: test"
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v6
    - uses: actions/setup-node@v6
      with:
        node-version: '20'
    - run: cd web && npm ci
    - run: cd web && npm run test:run
```

2. **Remove `continue-on-error: true`** from the golangci-lint step. Before doing this,
run `golangci-lint run ./...` locally and fix all reported issues so the gate doesn't
immediately fail.

---

## Step 11 — Update TESTING.md version

**Source**: Maturity H6
**File**: `docs/TESTING.md`
**Effort**: 10 min

Update the version header from `0.17.0-beta` to `1.2.2` and update the
"Last Updated" date to today. Review coverage table for any services that have
changed since the table was last written.

---

## Completion Checklist

- [ ] Step 1: CORS `else` branch removed + test added for rejected origin *(deferred)*
- [ ] Step 2: Security headers middleware created, wired, tested *(deferred)*
- [x] Step 3: serialize-javascript pinned to 7.0.5 via `overrides` (no vite-plugin-pwa downgrade needed)
- [x] Step 4: DOMPurify added to MarkdownRenderer; fallback path also sanitized
- [ ] Step 5: Avatar upload uses magic bytes detection + extension allowlist *(deferred)*
- [x] Step 6: `WriteError()` no longer exposes unknown error detail
- [x] Step 7: `X-Forwarded-For` split on comma, leftmost IP taken; `RemoteAddr` port stripped
- [x] Step 8: Password policy raised to 12 chars + complexity, applied in Register/ResetPassword/ChangePassword
- [x] Step 9: `rate_limit_exceeded` audit event wired via `OnExceeded` callback
- [x] Step 10: golangci-lint upgraded to v2.11.4; frontend test job added to CI
- [x] Step 11: TESTING.md version updated to 1.2.2
- [x] All existing tests pass: `make test`
- [x] All existing frontend tests pass: `cd web && npm run test:run`
- [x] `npm audit` shows 0 high/critical
- [x] CHANGELOG.md updated with Security section entries
- [x] Docs updated: `docs/OWASP_AUDIT_2026-04-03.md` result table updated to reflect fixes

---

## Expected Outcome

After this branch merges:

| Audit | Before | After |
|-------|--------|-------|
| OWASP failures | 3 (A05, A06, A08) | 0 |
| OWASP partials | 1 (A04) | 0 |
| Maturity Auth/Access | Moderate (2) | Satisfactory (3) |
| Maturity Testing | Satisfactory (3) | Satisfactory→Strong (3+) |
| Overall maturity score | 2.8 / 4.0 | ~3.2 / 4.0 |
