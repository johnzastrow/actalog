# Protected-User Break-Glass CLI — v1.3.2 Design

**Date:** 2026-05-07
**Target release:** v1.3.2 (combined with the v1.3.1 admin user lifecycle work — single PR, single tag, single CHANGELOG entry)
**Spec partner:** brainstorming session 2026-05-07 (this doc) → writing-plans → implementation

## Goal

Provide a documented, audited operator escape hatch for editing protected user accounts when the normal admin paths are unavailable. Fixes the operator-recovery scenario currently solved by an ad-hoc `secrets/reset-password.sh` script: directly write to the DB, bypassing L1+L2 (and temporarily L3 for identity fields), with full forensic audit trail.

Fields supported: `password | email | name | role | account_disabled` (full operator recovery, per design call).

## Use cases

Primary: protected user has forgotten their password and email is unavailable; operator needs to set a new credential to restore access.

Secondary:
- Identity-field cleanup after a typo or compromise (rename, role change, account_disabled toggle)
- Email change driven by a user-initiated address migration that the protected-list policy doesn't yet permit through the normal admin screen

## Non-goals

- Editing non-protected users via this CLI. Those go through the normal admin UI / API (`/api/admin/users/{id}`). The CLI **rejects** non-protected target emails.
- Bulk operations. One field per invocation.
- Adding/removing entries from the protected list. That stays a code-change + PR (existing process in `pkg/security/protected_users.go`).

## Architecture

```
operator shell
      │
      ▼
actalog admin force-edit-protected --email X --field F [--value V | --stdin] --confirm
      │
      ▼
cmd/actalog/main.go admin-dispatch  (extends existing switch)
      │
      ▼
internal/protectedusers/breakglass.go  AdminForceEditProtected(...)
      │
      ├── 1. Validate field allowlist + value (per-field rule)
      ├── 2. Confirm target IS protected (security.IsProtectedEmail)
      ├── 3. Capture operator metadata (USER, hostname, tty, cwd)
      ├── 4. Interactive confirmation (always for password; "type BREAK-GLASS" for identity)
      ├── 5. For password: hash + UpdatePassword + UnlockAccount + RevokeAllForUser
      ├── 6. For identity: drop L3 triggers → UPDATE → reinstall L3 triggers
      ├── 7. Re-run boot invariant (3/3 hard checks); abort + recovery message on failure
      └── 8. Write per-field audit event with operator metadata in details
```

### Reused, not rebuilt

- `pkg/auth.HashPassword` (bcrypt cost 12) — same hash function the application uses
- `pkg/security.IsProtectedEmail` — target eligibility check
- `internal/repository/protected_triggers_sql.go` — drop/reinstall trigger SQL constants per dialect (existing from v1.3.0; same SQL the recovery scripts use)
- `internal/protectedusers/invariant.go` `VerifyProtectedUserInvariant` — post-edit verification step
- `internal/repository.NewAuditLogRepository` + `service.NewAuditLogService` — same minimal-graph audit construction the existing `reapply-protected-migrations` CLI uses

### What's new

- `internal/protectedusers/breakglass.go` — the core `AdminForceEditProtected` function
- 5 new audit event constants in `internal/domain/audit_log.go`
- New CLI subcommand wired into the existing admin dispatch in `cmd/actalog/main.go`

## CLI shape

```
actalog admin force-edit-protected \
  --email <target>              # required; must be on the protected list
  --field <field>               # required; one of password|email|name|role|account_disabled
  [--value <new-value>          # required for email|name|role|account_disabled
   | --stdin]                   # required for password (reads + confirms via terminal)
  --confirm                     # required guard flag
```

`--value` and `--stdin` are mutually exclusive. Password ALWAYS reads from stdin (forbidden to pass on `--value`) — avoids `ps`/shell-history exposure of credentials.

## Per-field rules

| Field | Validation | Side effects | L3 trigger |
|-------|-----------|--------------|------------|
| `password` | ≥12 chars, upper, lower, digit (matches `validatePassword`) | `UpdatePassword` (bcrypt-12 hash) → `UnlockAccount` (clears `failed_login_attempts`, `locked_at`, `locked_until`) → `RevokeAllForUser` | **Not touched** — `password_hash` is a lifecycle field; L3 already lets it through |
| `email` | RFC 5322 parse + lower-case + NOT on protected list (can't migrate-into protected) + unique | UPDATE `users.email` | Drop both triggers → UPDATE → reinstall (lockstep SQL from v1.3.0) |
| `name` | 1–100 chars after trim | UPDATE `users.name` | same drop/reinstall |
| `role` | one of `athlete\|coach\|admin` | UPDATE `users.role` | same drop/reinstall |
| `account_disabled` | parses to bool (`true`/`false`/`1`/`0`/`yes`/`no`) | UPDATE column + `disabled_at` (now or NULL) + `disabled_by_user_id = NULL` + `disable_reason = "break-glass: <op-name>"` | same drop/reinstall |

For identity-field changes: if reinstall fails, the program **panics** with a loud recovery message:

```
FATAL: trigger reinstall failed after edit. Protected-user defense is DEGRADED.
Run NOW:
  ./bin/actalog admin reapply-protected-migrations --confirm
Or:
  scripts/recover/restore-protected-triggers.sh

The audit event was written before the reinstall attempt; review it.
```

## Operator confirmation flow

For `password` (no value on argv, stdin reads):

```
⚠  BREAK-GLASS OPERATION
   Target: br8kwall@gmail.com  (id=1, role=admin)
   Field:  password
   Audit:  protected_user_break_glass_password

New password: ********
Confirm:      ********
[proceeds with hash + write + audit]
```

For identity fields (additional second-confirmation gate):

```
⚠  BREAK-GLASS OPERATION
   Target: br8kwall@gmail.com  (id=1, current role=admin)
   Field:  role
   New:    athlete
   This will drop L3 triggers, UPDATE, reinstall, re-verify boot invariant.

Type 'BREAK-GLASS' to proceed: BREAK-GLASS
[proceeds]
```

The "type BREAK-GLASS" gate is intentionally annoying — this should be hard to fat-finger. Without `--confirm` flag the program rejects with usage text before reaching the prompt.

## Audit catalog

### Five new event constants

```go
EventProtectedUserBreakGlassPassword         = "protected_user_break_glass_password"
EventProtectedUserBreakGlassEmail            = "protected_user_break_glass_email"
EventProtectedUserBreakGlassName             = "protected_user_break_glass_name"
EventProtectedUserBreakGlassRole             = "protected_user_break_glass_role"
EventProtectedUserBreakGlassAccountDisabled  = "protected_user_break_glass_account_disabled"
```

Per-field rather than one event with a `field` discriminator: enables clean alert-routing and per-field log queries without JSON-path filters.

### Audit row shape

| Column | Value |
|--------|-------|
| `event_type` | one of the five constants above |
| `actor_id` | `NULL` (no logged-in user — break-glass runs from a shell) |
| `target_id` | the protected user's `users.id` |
| `details` | structured operator metadata + value diff (see below) |

### `details` body per field

**Password:**

```json
{
  "operator_user":     "jcz",
  "operator_hostname": "daisy.lan",
  "operator_tty":      "/dev/pts/3",
  "operator_cwd":      "/home/jcz/Github/actionlog",
  "cleared_failed_login_attempts": 3,
  "cleared_lockout":               true,
  "revoked_refresh_tokens":        true
}
```

NEVER includes the password, the hash, or any input typed by the operator.

**Identity fields (email/name/role/account_disabled):**

```json
{
  "operator_user":     "jcz",
  "operator_hostname": "daisy.lan",
  "operator_tty":      "/dev/pts/3",
  "operator_cwd":      "/home/jcz/Github/actionlog",
  "old_value":         "admin",
  "new_value":         "athlete",
  "trigger_dropped":   true
}
```

For `email`, `old_value` / `new_value` are full email addresses (forensic clarity outweighs PII concern — this is operator action against an account the operator already controls; the change itself goes into the audit log so partial values don't help).

`trigger_dropped: true` for identity fields documents that L3 was briefly down. `false` for password.

## File map

**New files:**
- `internal/protectedusers/breakglass.go` — `AdminForceEditProtected`, helpers, per-field handlers
- `internal/protectedusers/breakglass_test.go` — unit tests
- `test/integration/protected_users_break_glass_test.go` — multi-DB matrix end-to-end

**Modified files:**
- `internal/domain/audit_log.go` — five new event constants
- `cmd/actalog/main.go` — extend admin-dispatch switch with `force-edit-protected` case
- `docs/security/PROTECTED_USERS.md` — new "Break-glass CLI" subsection under §recovery
- `docs/security/PROTECTED_USERS_RECOVERY.md` — point at the CLI from the 3-AM playbook
- `docs/security/THREAT_MODEL.md` — residual-risk row for shell-access bypass capability
- `docs/CHANGELOG.md` — collapse the v1.3.1 entry into v1.3.2; add break-glass section
- `docs/USER_PERMISSIONS.md` — note the operator-only CLI command
- `docs/TODO.md` — mark the v1.3.1 break-glass line complete
- `pkg/version/version.go` — `Patch = 1` → `Patch = 2`
- `web/package.json` + `web/package-lock.json` — `1.3.1` → `1.3.2`
- `CLAUDE.md` — version line + docker-tag examples
- `secrets/reset-password.sh` — flagged as superseded by the CLI (kept as fallback if the binary itself is broken; comment updated)

## Threat-model addition (`docs/security/THREAT_MODEL.md`)

New row under residual risks:

> **Shell access on the application host** — the operator running `actalog admin force-edit-protected` can bypass the entire L1+L2+L3 stack with a single command. **Accepted** for small-team deployments where shell access already implies DB credential access via `secrets/local-test-credentials.env` or equivalent. Mitigation: every break-glass operation writes an audit event with operator metadata (`USER`/`hostname`/`tty`/`cwd`) — post-incident review can answer who-when-where.

## Test surface

| Layer | Tests |
|-------|-------|
| `internal/protectedusers/breakglass_test.go` | per-field validation (rejects bad input); rejects non-protected target; rejects without --confirm; password reads stdin; identity fields require BREAK-GLASS confirmation |
| `test/integration/protected_users_break_glass_test.go` | multi-DB end-to-end: password change → login works → audit fired; role change → trigger drops + reinstalls + invariant verifies; account_disabled toggle → disabled_at/disable_reason populated; email change rejected if target email is on protected list |
| Existing `TestRecovery_*` and `TestL3_*` tests | must still pass — break-glass piggybacks on the existing trigger drop/reinstall plumbing |

## Release strategy

This is the merged v1.3.1 + break-glass release as **v1.3.2**:

1. The `feature/admin-user-lifecycle-v1.3.1` branch keeps its Git history but the version number is bumped 1.3.1 → 1.3.2 in a single commit.
2. CHANGELOG: collapse the v1.3.1 entry into v1.3.2 with both feature areas documented.
3. PR #222 retitled to v1.3.2; force-push.
4. Net result: one PR, one tag (`v1.3.2`), one image, one CHANGELOG entry.

## Open questions for the plan-writing phase

These are intentionally deferred:

- Whether to also expose `--bypass-confirm-for-password` for scripted use cases. Current call: no — interactive confirmation is mandatory. If automation is needed later, a separate non-interactive command can be added with stricter logging.
- Whether the `account_disabled` value parser should accept additional formats (e.g., `enabled`/`disabled`). Plan can iterate on the small parser if the matrix list becomes useful.
- Whether `secrets/reset-password.sh` should be removed entirely. Plan: keep as fallback (works when the binary itself is broken), update its header comment to point at the CLI as the preferred path.
