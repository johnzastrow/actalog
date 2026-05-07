# Admin User Lifecycle — v1.3.1 Design

**Date:** 2026-05-06
**Target release:** v1.3.1
**Scope partner:** brainstorming session 2026-05-06 (this doc) → writing-plans → implementation

## Goal

Give admins a complete way to bring users into the system and recover broken accounts without leaving the admin UI. Specifically:

1. **Create a new user account** from the admin User Management screen, populating credentials directly so the new user can sign in immediately (no self-registration round trip).
2. **Set a user's password directly** when force-reset-by-email isn't the right tool (small-team recovery, onboarding fix-ups, lockout recovery).

Out of scope (separate v1.3.x spec): break-glass CLI for protected users.

## Use cases

Primary: an admin onboards a new gym member or training partner who hasn't registered themselves. Admin populates the account; user signs in with provided credentials.

Secondary: admin recovers an account where self-registration went wrong, or where the user is locked out and email is unreliable.

Tertiary (acknowledged but not the driver): bulk admin provisioning during demo/test setup.

## Non-goals

- 2FA / TOTP. Doesn't exist yet; the design leaves room for it without committing.
- Forced password change on first login. Considered and rejected — small-team deployment, no real benefit, costs a column + login interceptor + forced-change page.
- Mass / scripted user provisioning. CSV import already covers that.
- Break-glass edits to protected users. Separate spec.

## Architecture

```
                   AdminUsersView (existing)
                         │
              [+ Create User button — new]
                         │
                         ▼
              AdminUserCreateDialog (new)
                         │
                         ▼  POST /api/admin/users
              ┌──────────────────────────┐
              │  AdminUserHandler.Create │
              │  → AdminUserService.Create
              │  → UserRepository.Create │
              │  → audit: admin_user_created
              └──────────────────────────┘

  AdminUserEditView/ProfileTab (existing v1.3.0)
                         │
              [+ Password Management card — new]
                         │
              [Force Password Reset]   [Set Password Directly]
                         │                       │
                         │                       ▼
                         │           AdminSetPasswordDialog (new)
                         │                       │
                         │                       ▼
                         │      POST /api/admin/users/{id}/password
                         │      ┌─────────────────────────────────┐
                         │      │ AdminUserHandler.SetPassword    │
                         │      │ → AdminUserService.SetPassword  │
                         │      │   - hash + store password       │
                         │      │   - clear failed_login_attempts │
                         │      │   - clear locked_at/locked_until│
                         │      │   - revoke all refresh tokens   │
                         │      │ → audit: admin_password_set     │
                         │      └─────────────────────────────────┘
                         ▼
              POST /api/admin/users/{id}/force-password-reset (existing — unchanged)
```

### Reused, not rebuilt

- `pkg/auth.HashPassword` (bcrypt cost 12)
- Password complexity policy (≥12 chars + upper + lower + digit) — extracted to `pkg/auth.ValidatePasswordComplexity` if not already a single helper
- `pkg/middleware.AdminOnly`
- `pkg/middleware.ProtectedUserGuard` (L1) — protected accounts can't have passwords set via admin endpoint
- `domain.InvalidInputError` (shipped in v1.3.0 hotfix #220) — 400 with structured field+message
- `RefreshTokenRepository.RevokeAllForUser`
- Existing audit-log infrastructure in `internal/service/audit_log_service.go`

## Backend API

### `POST /api/admin/users` — create user

**Auth chain:** auth → AdminOnly → handler.

**Request:**

```json
{
  "email": "newcoach@example.com",
  "password": "SetByAdminAtCreate1",
  "name": "Jamie Coach",
  "role": "athlete",
  "email_verified": true
}
```

| Field | Required | Default | Validation |
|-------|----------|---------|------------|
| `email` | yes | — | RFC 5322 parseable; unique (case-insensitive); not on `security.IsProtectedEmail` list |
| `password` | yes | — | ≥12 chars + upper + lower + digit |
| `name` | yes | — | 1–100 chars after trim |
| `role` | no | `"athlete"` | one of `athlete \| coach \| admin` |
| `email_verified` | no | `true` | bool |

**Responses:**

| Code | Body | Trigger |
|------|------|---------|
| 201 | full User JSON | success |
| 400 | `{error: "invalid_input", message: "<field>: <rule>"}` | validation fail |
| 400 | `{error: "invalid_input", message: "email: is reserved"}` | protected email |
| 409 | `{error: "duplicate_email", message: "..."}` | email already exists |
| 401 / 403 | per existing middleware | unauthenticated / non-admin |

**Side effects:**
- Hash password via existing helper, store via `UserRepository.Create`
- Audit event `admin_user_created` with `details: {role, email_verified, email_domain}`
- No JWT generated (admin isn't logging in as the new user)
- No verification email sent — `email_verified=true` is the path; if the admin un-checked it, the user follows the existing self-verification flow on their own login

### `POST /api/admin/users/{id}/password` — admin-set password

**Auth chain:** auth → AdminOnly → ProtectedUserGuard → handler.

**Request:**

```json
{ "new_password": "..." }
```

**Responses:**

| Code | Body | Trigger |
|------|------|---------|
| 204 | (empty) | success |
| 400 | `{error: "invalid_input", message: "new_password: <rule>"}` | complexity fail |
| 403 | `{error: "protected_user", ...}` | L1 guard fired |
| 404 | `{error: "not_found", message: "user not found"}` | invalid `id` |
| 401 / 403 | per existing middleware | unauthenticated / non-admin |

**Side effects (single transaction):**
- Hash new password, update `password_hash`
- `failed_login_attempts = 0`
- `locked_at = NULL`, `locked_until = NULL`
- `RevokeAllForUser(targetID)` on the refresh-token repo
- Audit event `admin_password_set` with `details: {cleared_failed_login_attempts: <prior int>, cleared_lockout: <bool>, revoked_refresh_tokens: <count>}`

**Explicitly not touched:** `account_disabled`, `disabled_at`, `disabled_by_user_id`, `disable_reason`, `email_verified`, `email_verified_at`. Admin policy lives in those fields; password-set is a credential operation, not a policy override.

### `POST /api/admin/users/{id}/force-password-reset` — unchanged

Already exists. Still sends a reset email + revokes refresh tokens. Frontend will surface it as a separate button alongside the new direct-set button.

### Authorization summary

| Endpoint | Middleware order |
|----------|------------------|
| `POST /api/admin/users` | auth → AdminOnly → handler |
| `POST /api/admin/users/{id}/password` | auth → AdminOnly → ProtectedUserGuard → handler |
| `POST /api/admin/users/{id}/force-password-reset` | auth → AdminOnly → ProtectedUserGuard → handler (existing) |

## Frontend UX

### `AdminUsersView.vue`

Primary-color **Create User** button in the toolbar, next to existing search/filter. Click opens `AdminUserCreateDialog`. Existing route guard already restricts the view to admins.

### `AdminUserCreateDialog.vue` (new)

Modal `v-dialog`. Form fields top-to-bottom:

| Field | Control | Default | Notes |
|-------|---------|---------|-------|
| Email | `v-text-field type="email"` autofocus | — | required |
| Password | `v-text-field type="password"` | — | show/hide toggle |
| Confirm password | `v-text-field type="password"` | — | match validation |
| Name | `v-text-field` | — | trim on submit |
| Role | `v-select` with athlete/coach/admin | `athlete` | required |
| Email verified | `v-checkbox` | checked | tooltip explains the implication |

**Buttons:** Cancel | Create (disabled until form validates).

**Submit flow:**
- Send `POST /api/admin/users`
- 201 → close dialog, refresh user list, toast `"User <name> created"`
- 400 `invalid_input` → inline error under the offending field, parsed from the `field: ...` prefix
- 409 `duplicate_email` → inline error on email field
- 5xx → toast `"Could not create user; check server logs"` (generic; details server-side)

### Profile tab Password Management card

A new card on `ProfileTab.vue` below the existing field grid:

```
┌─ Password Management ─────────────────────────────────┐
│  [Force Password Reset]   [Set Password Directly]      │
│                                                        │
│  Force Reset sends an email link; user picks their     │
│  own password. Set Directly lets you type the          │
│  password yourself — both will sign the user out of    │
│  all current sessions and clear any account lockout.   │
└────────────────────────────────────────────────────────┘
```

The explanatory paragraph is required — the buttons look similar but do different things, and getting them mixed up has UX consequences.

### `AdminSetPasswordDialog.vue` (new)

- Read-only header: `"Set password for <user.email>"`
- New password / Confirm new password (same complexity rules as create dialog)
- Submit → `POST /api/admin/users/{id}/password`
- 204 → close, toast `"Password set. <user.email> signed out of all devices."`
- 400 → inline error
- 403 → inline message `"This account is system-reserved"` (route guard should prevent reaching this dialog, but degrade gracefully)

### Reused / extracted components

`<PasswordInputs>` composable — shared between create dialog and set-password dialog. Encapsulates new+confirm pair, show/hide toggle, matching validation, and the complexity-policy hint text. Single source of complexity rules in the frontend.

### Common dialog behaviour

- `Esc` and outside-click cancel
- Submit button shows loading spinner while in flight
- Form state resets on dialog close
- Client-side validation always replicated server-side

## Audit catalog

### New event constants

| Constant | Value | When |
|----------|-------|------|
| `EventAdminUserCreated` | `admin_user_created` | `POST /api/admin/users` 201 |
| `EventAdminPasswordSet` | `admin_password_set` | `POST /api/admin/users/{id}/password` 204 |
| `EventAdminUserCreateRejectedProtected` | `admin_user_create_rejected_protected` | admin tried to create using a protected email |

### Reused (no change)

- `EventPasswordResetForcedByAdmin` — existing force-password-reset
- `EventProtectedUserAttackHTTP` — L1 guard fires on the new set-password endpoint when target is protected
- `EventProtectedUserAttackService` — L2 guard

### Detail bodies

```
admin_user_created
  details: { role, email_verified, email_domain }
  Email domain only ("gmail.com"), not the full address — full email is in target_id linkage.
  Keeps logs queryable without over-exposing PII.

admin_password_set
  details: {
    cleared_failed_login_attempts: <int>,   // value before reset
    cleared_lockout: <bool>,                // was locked at time of set
    revoked_refresh_tokens: <int>           // forensic count
  }
  NEVER includes the password, hash, or any input typed by the admin.

admin_user_create_rejected_protected
  details: { attempted_email }
  Captures the operator's intent for security review.

password_reset_forced_by_admin (existing — unchanged)
  details: { revoked_refresh_tokens: <int> }
```

### Logging discipline

- Plaintext password is never logged anywhere — not in audit details, not in handler error logs, not in stack traces
- Failed-validation paths log the URL and outcome only; the request body is never written to logs
- Audit-log writes happen *after* DB commit so a failed write doesn't leave a phantom event

### Audit coverage policy

Every administrative action that **changes state** is audited. Validation rejections (e.g., admin types a too-short password, fixes it, retries) are **not** audited — they don't change state and would generate noise that buries the meaningful events.

Authorisation failures (`401`, `403`) are already captured by existing middleware logging. Protected-user blocks fire dedicated audit events at the L1/L2/L3 layer that catches them (existing `protected_user_attack_*` family).

This policy keeps the audit log a high-signal record of "what changed and who changed it" rather than a per-request access log.

## Protected-user interaction matrix

| Admin action | Result |
|--------------|--------|
| `POST /api/admin/users` with body email on protected list | Service rejects with 400 + `admin_user_create_rejected_protected` audit event. Not `protected_user_attack_*` — that family is for *modifications* of protected rows, not attempts to create new ones. |
| `POST /api/admin/users/{id}/password` with `{id}` protected | L1 fires before service, returns 403, fires `protected_user_attack_http` audit event. |
| Profile tab loads for a protected user | UI route guard blocks (v1.3.0 work). The set-password dialog is unreachable. |

## Password security

- bcrypt cost 12 via `pkg/auth.HashPassword` (existing)
- Complexity policy: ≥12 chars + uppercase + lowercase + digit (matches registration). Implementation in `pkg/auth.ValidatePasswordComplexity` — a single source of truth used by registration, change-password, admin-create, and admin-set
- Plaintext never logged; see audit logging discipline above
- Refresh token revocation runs in the same service-layer transaction as the password write, so partial-state windows are minimised. If revocation fails after a successful password write, the response is still **204** (the operator's intent — set the password — succeeded); the failure is captured server-side as a `[WARN]` log line and as a `revoked_refresh_tokens: -1` sentinel in the audit `details`, and ops can grep the warning to know to advise the user to log out everywhere manually

## Rate limiting

Both new endpoints inherit the existing global admin rate limit. No per-endpoint tightening at this point — admin actions are deliberately rare and a per-actor limit can be added later if a compromise scenario justifies it.

## Threat model addendum (`docs/security/THREAT_MODEL.md`)

Add row: *Admin compromise → mass password reset / mass user creation*. Detection via audit log review (`admin_password_set` and `admin_user_created` events per actor over a window); not blocked at the system level. Accepted residual risk for a small-team deployment with admin trust.

## Test surface

| Layer | Tests |
|-------|-------|
| `pkg/auth` | `ValidatePasswordComplexity` unit tests (boundary cases) |
| `internal/service/admin_user_service` | `CreateUser` happy path; rejections for short password, missing upper, missing lower, missing digit, name too long, name empty, bad email, duplicate email, protected email, bad role; `SetPassword` happy + lockout-clear semantics + complexity rejection + 404 + protected (defensive) + revoke-failure |
| `internal/handler/admin_user_handler` | route registration; 201/400/403/404/409 for create; 204/400/403/404 for set-password; audit-call assertions via stub |
| `test/integration` | end-to-end against sqlite3 / postgres / mysql: create user → log in as new user → fetch profile; set-password → verify counters cleared in DB → log in with new password; protected-email rejection (create); protected-user 403 (set-password) |
| Web (`vitest`) | `AdminUserCreateDialog`: form validation cases, submit success, submit failure, error inline placement; `AdminSetPasswordDialog`: same; `<PasswordInputs>` composable matching-validation; `AdminUsersView`: Create button visibility for admin vs non-admin; `ProfileTab`: Password Management card visibility and button state for protected vs non-protected target |

## Files touched

**New:**
- `web/src/components/admin/AdminUserCreateDialog.vue`
- `web/src/components/admin/AdminSetPasswordDialog.vue`
- `web/src/components/admin/composables/usePasswordInputs.js`
- Tests for each (`*.test.js` siblings)

**Modified:**
- `internal/handler/admin_user_handler.go` — add `CreateUser`, `SetPassword` methods
- `internal/service/admin_user_service.go` — add `CreateUser`, `SetPassword` methods
- `internal/domain/audit_log.go` — three new event constants
- `cmd/actalog/main.go` — route registration for the two new endpoints
- `web/src/views/AdminUsersView.vue` — toolbar Create button + dialog wiring
- `web/src/components/admin/user-edit/ProfileTab.vue` — Password Management card + dialog wiring
- `pkg/auth/password.go` (or equivalent) — extract `ValidatePasswordComplexity` if not already separated
- `docs/security/THREAT_MODEL.md` — admin compromise row
- `docs/CHANGELOG.md` — v1.3.1 entry
- `docs/USER_PERMISSIONS.md` — new endpoints + UI
- `docs/TODO.md` — mark items complete

## Open questions for plan-writing phase

These are intentionally deferred to the plan, not the design — they're implementation details:

- Whether `ValidatePasswordComplexity` already exists as a helper or needs extracting from `Register` / `ChangePassword` / `ResetPassword`. Plan should start with a quick code read.
- Exact wire format for the `409 duplicate_email` body (use existing `MapServiceError` mapping or add a new sentinel).
- Whether the ProfileTab "Password Management" card needs feature-flagging behind a property to allow staged rollout (probably no — it's strictly additive).
