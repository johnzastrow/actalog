# Protected Users — Master Runbook

**Version:** v1.3.0
**Last updated:** 2026-04-28
**Owner:** Security / ops on-call

---

## Contents

1. [Purpose and scope](#1-purpose-and-scope)
2. [Threat model](#2-threat-model)
3. [Architecture](#3-architecture)
4. [The protected list](#4-the-protected-list)
5. [How to add a protected user](#5-how-to-add-a-protected-user)
6. [How to remove a protected user](#6-how-to-remove-a-protected-user)
7. [How to verify the system is healthy](#7-how-to-verify-the-system-is-healthy)
8. [Recovery playbook](#8-recovery-playbook)
9. [Audit-log forensics](#9-audit-log-forensics)
10. [Degraded mode](#10-degraded-mode)

---

## 1. Purpose and scope

The protected-user system prevents any admin operation — whether it comes from the
browser UI, a direct API call, a background job, a rogue migration, or raw SQL — from
modifying or deleting a designated set of system-critical accounts.

**What this runbook covers:**

- The four defense layers (L1 middleware, L2 service guard, L3 database trigger, L4
  audit log) and how they fit together
- The boot-time invariant that verifies the layers are intact on every start
- Day-to-day operations: listing protected users, adding/removing entries, verifying
  health
- Recovery procedures when any layer is broken
- Audit-log query patterns for detecting and investigating attacks
- Degraded mode: when to use it, what it disables, and how to exit it

**What this runbook does not cover:**

- Broader application security (see `docs/security/THREAT_MODEL.md`)
- The 3-AM quick-response playbook (see `docs/security/PROTECTED_USERS_RECOVERY.md`)
- Scripts under `scripts/recover/` (see `scripts/recover/README.md`)

---

## 2. Threat model

### What the system protects against

| Threat | Blocked by |
|--------|------------|
| Admin uses the UI to edit/disable a protected account | L1 middleware (HTTP 403) + UI guard (shield icon + disabled actions) |
| Authenticated admin calls the API directly (curl, script) | L1 middleware; L2 service guard |
| Cron job or internal caller reaches the service layer | L2 service guard |
| SQL injection that bypasses the Go layer | L3 database trigger |
| Rogue migration targeting protected rows | L3 database trigger |
| Someone with DB credentials but no app access | L3 database trigger |
| Silent rollback of the binary removes L1 and L2 | L3 trigger lives in the DB; survives binary rollback |
| Trigger silently dropped between deployments | Boot-time invariant (Check 1 + Check 2) refuses to start |
| Protected user row deleted | Boot-time invariant (Check 3) refuses to start (hard) |
| Any of the above succeeds silently | L4 audit log records which layer caught it |

### What the system does NOT protect against

- **Root access on the database server.** An operator who can directly connect to
  the database server and issue DDL (`DROP TRIGGER`) followed by DML can defeat L3.
  The boot invariant will detect the missing trigger on the next binary start.
- **Physical disk access / volume snapshots.** Out-of-band restoration of a database
  from before the protected-user migration leaves no triggers in place.
- **Supply-chain compromise of Go modules.** A malicious dependency that replaces
  `pkg/security/IsProtectedEmail` with a no-op would defeat L1 and L2. Use
  `go.sum` lockfile pinning and Dependabot alerts.
- **Hot-patching a running binary.** If the running binary is patched in memory,
  L1 and L2 may be disabled for the lifetime of that process. L3 still holds.
- **Plus-addressing or sub-addressing aliasing.** `br8kwall+anything@gmail.com` is
  a different address and is NOT protected. The check is exact-match (normalized to
  lowercase) — intentional; do not add wildcards.

---

## 3. Architecture

### 3.1 Overview — four independent layers

A write request targeting a protected user traverses these layers in order. The first
layer that fires rejects the request; layers below it are never reached. Each layer
emits exactly one audit event tagged with its own layer code — NOT all traversed
layers. If L1 fires, the event is tagged `_http`. If the request somehow reached L3,
the event is tagged `_db`.

```text
HTTP request
     │
     ▼
┌────────────────────────────────────────────────────────┐
│ L1: pkg/middleware/protected_user.go                   │
│     ProtectedUserGuard                                 │
│     Checks security.IsProtectedEmail(user.Email)       │
│     → HTTP 403 + JSON body + EventProtectedUserAttackHTTP audit  │
└────────────────────────────────────────────────────────┘
     │ (non-protected: pass through)
     ▼
┌────────────────────────────────────────────────────────┐
│ L2: internal/service/admin_user_service.go             │
│     ensureNotProtected(actorID, targetID)              │
│     → domain.ErrProtectedUser + EventProtectedUserAttackService  │
└────────────────────────────────────────────────────────┘
     │ (non-protected: continue)
     ▼
┌────────────────────────────────────────────────────────┐
│ L3: Database triggers (migration 0.35.0)               │
│     protected_users_no_update / protected_users_no_delete        │
│     BEFORE UPDATE / BEFORE DELETE on users             │
│     → DB error containing TriggerErrorContract string  │
│     → service layer catches, fires EventProtectedUserAttackDB    │
└────────────────────────────────────────────────────────┘
     │ (non-protected: execute DML)
     ▼
┌────────────────────────────────────────────────────────┐
│ L4: Audit log                                          │
│     internal/domain/audit_log.go event constants      │
│     Every blocked attempt is recorded with actor_id,  │
│     target_id, path, method, user_agent, referer.      │
└────────────────────────────────────────────────────────┘
```

### 3.2 The boot-time invariant

Before the HTTP server starts, `cmd/actalog/main.go` calls
`internal/protectedusers.VerifyProtectedUserInvariant(db, driver)` (via `boot_invariant.go`).
This runs three checks:

| Check | What it tests | Failure mode |
|-------|---------------|--------------|
| **1 — triggers exist** | `protected_users_no_update` and `protected_users_no_delete` appear in the DB catalog | **HARD** — binary refuses to start |
| **2 — triggers fire** | Simulated UPDATE inside `BEGIN; ROLLBACK` is rejected with the contract error message | **HARD** — binary refuses to start |
| **3 — protected rows exist** | Each protected email has a row in `users` | **SOFT** if zero users total (fresh install), **HARD** if other users exist |

On any HARD failure, the binary prints the failing check and the exact recovery command,
then exits non-zero. On SOFT failure (fresh install only), it logs a WARN and starts
normally — the owner will register through the normal flow.

#### Boot log — all checks passed

```text
protected-user invariant: 3/3 hard checks passed (warnings: 0)
```

#### Boot log — fresh install (soft only)

```text
protected-user invariant: 2/3 hard checks passed (warnings: 1)
warn: protected user "br8kwall@gmail.com" not yet present (expected on fresh install)
```

#### Boot log — HARD failure (trigger missing)

```text
[ERROR] protected-user invariant failed: invariant: trigger "protected_users_no_update" missing — recover with: ./bin/actalog admin reapply-protected-migrations --confirm
[FATAL] Refusing to start. To recover (most → least disruptive):
  1) ./bin/actalog admin reapply-protected-migrations --confirm
  2) ./bin/actalog admin verify-protected-users --verbose
  3) scripts/recover/restore-protected-triggers.sh

Full runbook: docs/security/PROTECTED_USERS.md#recovery
```

### 3.3 The three artifacts that must stay in sync

The protected email list lives in three places. All three are updated in the same PR
by `make gen-protected-emails`:

| Artifact | File | Role |
|----------|------|------|
| Go registry (single source of truth) | `pkg/security/protected_users.go` | Runtime check at L1 and L2 |
| Migration trigger SQL | `internal/repository/protected_triggers_sql.go` (constants), applied by migration `0.35.0_add_protected_user_triggers` | Runtime check at L3 (DB layer) |
| Frontend guard | `web/src/utils/protectedUsers.js` (auto-generated) | UI guard — disables Edit/Disable/Delete actions |

CI enforces that all three are consistent on every PR (`make gen-protected-emails &&
git diff --exit-code`). Any drift fails the build.

### 3.4 Rollback property — security stays on

Rolling back the binary from v1.3.0 to v1.2.4 leaves the L3 triggers in place.
v1.2.4 does not know about them, but they keep rejecting writes at the database layer.
**Rolling back the binary does not roll back security.** This is intentional: the
strongest layer lives in the most stable place (the DB).

---

## 4. The protected list

The current list has one entry. Add entries only via the PR procedure in §5.

### `br8kwall@gmail.com`

This is the project owner's primary account. It was the first account registered in
production, so it holds the `admin` role by the "first user becomes admin" bootstrap
rule. It also controls the application's GitHub Actions secrets, Docker registry
credentials, and DNS records. Losing admin access to this account would require
out-of-band database repair to restore it.

The account is protected for availability, not secrecy: the email address appears in
the repository's CLAUDE.md, CI config, and commit history. The goal is to ensure that
no admin-level mistake (accidental disable, role demotion, email change, password
reset) can lock the owner out of their own system.

---

## 5. How to add a protected user

> Adding a protected user requires a code change, a database migration, and a
> CODEOWNERS-gated PR review. There is no runtime path — by design. See §2 (threat
> model) for why.

### Step-by-step checklist

**Before you start:** verify you understand why this account needs protection. The bar
is high — protection is a permanent, ops-heavy commitment. Consider whether a strong
password + MFA + admin role restriction is sufficient instead.

#### Step 1 — Edit the Go registry

File: `pkg/security/protected_users.go`

Add a new lowercase entry to the `protectedEmails` map:

```go
var protectedEmails = map[string]struct{}{
    "br8kwall@gmail.com":  {},
    "newprotected@example.com": {}, // added: <reason>, <date>
}
```

All keys must be lowercase. `IsProtectedEmail` lowercases its input before lookup, so
an uppercase key would silently never match.

#### Step 2 — Regenerate the frontend guard

```bash
make gen-protected-emails
```

This regenerates `web/src/utils/protectedUsers.js`. Commit the generated file
alongside the Go source change. Do not edit the generated file by hand.

#### Step 3 — Create a new migration

The migration must drop and recreate the L3 triggers with the updated email list. Use
the next available sequence number:

```bash
make migrate-create name=update_protected_users
```

The migration SQL for each dialect must use the same `WHEN/IF` pattern as
`internal/repository/protected_triggers_sql.go`. Also update the constants in that
file — they are the source for `AdminReapplyProtectedMigrations` and must stay in
lockstep with the migration.

For SQLite, the `WHEN` clause becomes a parenthesized `IN` list:

```sql
WHEN OLD.email IN ('br8kwall@gmail.com', 'newprotected@example.com')
```

For PostgreSQL, the `ANY(ARRAY[...])` check grows:

```sql
IF OLD.email = ANY(ARRAY['br8kwall@gmail.com', 'newprotected@example.com']) THEN
```

For MySQL/MariaDB, add a disjunction:

```sql
IF OLD.email = 'br8kwall@gmail.com' OR OLD.email = 'newprotected@example.com' THEN
```

Also update the constants in `internal/repository/protected_triggers_sql.go`
(`SQLiteProtectedTriggers`, `PostgresProtectedTriggers`, `MySQLProtectedTriggers`).
These are used by `AdminReapplyProtectedMigrations` — they must match the migration.

#### Step 4 — Verify locally

```bash
# Build the binary and run with your dev DB
make build
./bin/actalog admin verify-protected-users --verbose
```

Expected output (all checks pass):

```text
Protected-user invariant:
  Check 1/3 — L3 triggers exist        ✓
  Check 2/3 — triggers fire correctly  ✓
  Check 3/3 — protected rows exist     ✓

✓ all checks passed
```

#### Step 5 — Open the PR

The PR must touch all three artifacts:

- `pkg/security/protected_users.go`
- `web/src/utils/protectedUsers.js` (auto-generated, but must be re-generated)
- `internal/repository/protected_triggers_sql.go`
- A new migration file under `migrations/`

CI will fail if any of these are missing or inconsistent.

The `CODEOWNERS` file gates all changes to `pkg/security/**` on a security-team
reviewer. The PR cannot be merged without that approval.

#### Step 6 — After merge

The migration runs automatically on the next binary start. Verify with:

```bash
./bin/actalog admin verify-protected-users
```

---

## 6. How to remove a protected user

Removal follows the same PR procedure as adding. The checklist is identical except:

- Remove the entry from `protectedEmails` in `pkg/security/protected_users.go`
- Re-run `make gen-protected-emails` to update the frontend guard
- Add a migration that drops and recreates the triggers with the shorter list
- Update the SQL constants in `internal/repository/protected_triggers_sql.go`

After merge, the next binary start will no longer refuse writes to that account.
Existing `protected_user_attack_*` audit events for that email address remain in the
audit log — they are historical facts and are not deleted.

The git commit message for a removal is the audit trail for why protection was lifted.
Include a reason in the commit body.

---

## 7. How to verify the system is healthy

### 7.1 The verify CLI command

```bash
./bin/actalog admin verify-protected-users
./bin/actalog admin verify-protected-users --verbose
```

This command is **read-only and safe to run at any time**, including while the server
is running against the same DB. It does not modify any data.

The command calls `internal/protectedusers.AdminVerifyProtectedUsers` which runs
`VerifyProtectedUserInvariant` and prints a formatted report. Exit code: `0` = all
HARD checks pass; `1` = any HARD failure.

### 7.2 Healthy output

```text
Protected-user invariant:
  Check 1/3 — L3 triggers exist        ✓
  Check 2/3 — triggers fire correctly  ✓
  Check 3/3 — protected rows exist     ✓

✓ all checks passed
```

### 7.3 Trigger missing (Check 1 failure)

```text
Protected-user invariant:
  Check 1/3 — L3 triggers exist        ✗
  Check 2/3 — triggers fire correctly  ✗
  Check 3/3 — protected rows exist     ✓

Failure: invariant: trigger "protected_users_no_update" missing — recover with: ./bin/actalog admin reapply-protected-migrations --confirm

Recovery: ./bin/actalog admin reapply-protected-migrations --confirm
```

### 7.4 Trigger exists but does not fire (Check 2 failure)

This is the dangerous case — the trigger is in the catalog but its SQL body is wrong
(e.g., wrong email list, wrong error message, replaced by a no-op):

```text
Protected-user invariant:
  Check 1/3 — L3 triggers exist        ✓
  Check 2/3 — triggers fire correctly  ✗
  Check 3/3 — protected rows exist     ✓

Failure: invariant: simulated UPDATE for "br8kwall@gmail.com" did not reject as expected (trigger broken or absent): UPDATE was not rejected (trigger absent or broken). Recover with: ./bin/actalog admin reapply-protected-migrations --confirm

Recovery: ./bin/actalog admin reapply-protected-migrations --confirm
```

### 7.5 Protected user row missing while other users exist (Check 3 failure)

```text
Protected-user invariant:
  Check 1/3 — L3 triggers exist        ✓
  Check 2/3 — triggers fire correctly  ✓
  Check 3/3 — protected rows exist     ✗

Failure: invariant: protected user "br8kwall@gmail.com" missing while 47 other users exist — likely deletion. Restore from backup or remove from protected list via the operator runbook (docs/security/PROTECTED_USERS.md)

Recovery: ./bin/actalog admin reapply-protected-migrations --confirm
```

Note: reapplying migrations does not restore a deleted user row. See §8.4 (User row
deleted scenario) for the correct recovery path in this case.

### 7.6 Fresh install (soft warning only, not a failure)

```text
Protected-user invariant:
  Check 1/3 — L3 triggers exist        ✓
  Check 2/3 — triggers fire correctly  ✓
  Check 3/3 — protected rows exist     ⚠ (soft, fresh install)

✓ all checks passed
  warn: protected user "br8kwall@gmail.com" not yet present (expected on fresh install)
```

---

## 8. Recovery playbook

> **3-AM operator:** for the quick-reference panic playbook with nothing but
> error message → paste command, see `docs/security/PROTECTED_USERS_RECOVERY.md`.
> This section is the long-form reference with context and alternatives.

### 8.1 Decision matrix

| Failure symptom | First choice | Second choice | Third choice |
|----------------|--------------|---------------|--------------|
| Binary refuses to start — trigger missing | `reapply-protected-migrations` CLI | `scripts/recover/restore-protected-triggers.sh` | Restore DB from backup |
| Binary refuses to start — trigger fires wrong message | `reapply-protected-migrations` CLI | Manual trigger recreation in psql/mysql/sqlite3 | Restore DB from backup |
| Binary refuses to start — protected row missing | Restore row from backup snapshot; then restart | Add user via app registration if fresh env | Remove email from protected list (last resort) |
| Binary won't start AND CLI subcommands are unreachable | `ACTALOG_SKIP_PROTECTED_INVARIANT=true` (degraded mode) to get app running; then use verify/reapply CLI | `scripts/recover/restore-protected-triggers.sh` | — |
| Triggers dropped by a rogue migration in a running app | `reapply-protected-migrations` CLI (then restart binary so boot invariant re-passes) | — | — |
| Drift between Go source and migration SQL (CI failure only; not a runtime emergency) | Re-run `make gen-protected-emails`, update SQL constants, open a new migration | — | — |
| DB restored from a pre-trigger backup (no triggers) | `reapply-protected-migrations` CLI immediately after restore, before any traffic | — | — |

### 8.2 Scenario: trigger dropped or corrupted

**Symptoms:** Binary refuses to start with Check 1 or Check 2 failure. Existing
running instance is unaffected (triggers only checked at boot).

**Step 1 — Diagnose (read-only, safe on running DB):**

```bash
./bin/actalog admin verify-protected-users --verbose
```

**Step 2 — Reapply triggers (idempotent):**

```bash
./bin/actalog admin reapply-protected-migrations --confirm
```

This drops any existing protected-user triggers and recreates them from the compiled-in
SQL constants in `internal/repository/protected_triggers_sql.go`. The operation runs
in a transaction; a partial failure leaves the DB unchanged. After reapply it
re-runs the invariant check — if the invariant still fails, the command exits non-zero
with the reason.

**Step 3 — Restart the binary:**

```bash
pkill -9 -f actalog
./bin/actalog
```

**Step 4 — Verify:**

```bash
./bin/actalog admin verify-protected-users
```

Expected: all three checks pass.

### 8.3 Scenario: migration rolled back (triggers dropped in a running app)

If a migration rollback removes the L3 triggers while the app is running, L3 is
silently absent. L1 and L2 still protect at the application layer. L3 can be restored
without restarting the app:

**Step 1 — Restore the triggers while the app runs:**

```bash
./bin/actalog admin reapply-protected-migrations --confirm
```

**Step 2 — Schedule a restart to clear the boot invariant:**

The running binary's in-memory state does not need the trigger for L1/L2 to work.
However, the next restart will fail if the boot invariant checks trigger existence
and fails. Reapplying the triggers above is sufficient to pass that check. You can
restart at any convenient maintenance window.

**Step 3 — Confirm after restart:**

```bash
./bin/actalog admin verify-protected-users
```

### 8.4 Scenario: protected user row deleted

If the protected user row was deleted from the `users` table (Check 3 HARD failure),
the binary will refuse to start. The triggers are still present but protect a row that
no longer exists. The correct recovery is to restore the row from a backup, not to
reapply the migration.

**Step 1 — Locate the most recent backup that has the row (SQLite example):**

```bash
sqlite3 db_versions/actalog_latest_backup.db "SELECT id, email, role FROM users WHERE email = 'br8kwall@gmail.com';"
```

**Step 2 — Restore the row (pick the method that matches your DB driver):**

For SQLite — copy the value out of the backup and insert into the live DB:

```bash
sqlite3 data/actalog.db <<'SQL'
INSERT OR IGNORE INTO users (id, email, password_hash, name, role, is_active, created_at, updated_at)
SELECT id, email, password_hash, name, role, is_active, created_at, updated_at
FROM (
    -- substitute the actual values from your backup here
    SELECT
        1 AS id,
        'br8kwall@gmail.com' AS email,
        '<bcrypt_hash_from_backup>' AS password_hash,
        '<name_from_backup>' AS name,
        'admin' AS role,
        1 AS is_active,
        '2024-01-01T00:00:00Z' AS created_at,
        datetime('now') AS updated_at
)
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'br8kwall@gmail.com');
SQL
```

For PostgreSQL — use `pg_dump` to extract the row from a backup and restore it:

```bash
pg_dump --no-privileges --no-owner \
        --table=users \
        --where="email='br8kwall@gmail.com'" \
        -d jcz --schema=actalog \
        -h 192.168.1.143 -U jcz \
        -f /tmp/protected_user_row.sql

# Review the file, then apply to the live DB
psql -d jcz -h 192.168.1.143 -U jcz -f /tmp/protected_user_row.sql
```

For MySQL/MariaDB — similar approach using `mysqldump` with `--where`:

```bash
mysqldump --no-create-info --skip-triggers \
          --where="email='br8kwall@gmail.com'" \
          -h 192.168.1.234 -u jcz -p actalog users \
          > /tmp/protected_user_row.sql

mysql -h 192.168.1.234 -u jcz -p actalog < /tmp/protected_user_row.sql
```

**Step 3 — Verify and restart:**

```bash
./bin/actalog admin verify-protected-users --verbose
./bin/actalog
```

**Step 4 — Audit who deleted the row:**

```sql
SELECT * FROM audit_logs
WHERE target_user_id = (SELECT id FROM users WHERE email = 'br8kwall@gmail.com')
  AND event_type IN ('user_deleted', 'protected_user_attack_db')
ORDER BY created_at DESC
LIMIT 20;
```

### 8.5 Scenario: DB restored from a pre-trigger backup (v1.2.x era)

A backup taken before migration `0.35.0` was applied has no protected-user triggers.
Restoring it leaves the database in a v1.2.x state even though the binary is v1.3.0+.

**Step 1 — Restore the backup (normal procedure for your DB driver)**

**Step 2 — Immediately reapply the triggers before any traffic reaches the binary:**

```bash
./bin/actalog admin reapply-protected-migrations --confirm
```

If the binary refuses to start (boot invariant fails before accepting the CLI
subcommand), use the `ACTALOG_SKIP_PROTECTED_INVARIANT=true` escape hatch to reach
the CLI:

```bash
ACTALOG_SKIP_PROTECTED_INVARIANT=true ./bin/actalog admin reapply-protected-migrations --confirm
```

**Step 3 — Restart the binary normally (without the env var):**

```bash
./bin/actalog
```

The boot invariant should now pass.

**Step 4 — Verify:**

```bash
./bin/actalog admin verify-protected-users
```

### 8.6 Scenario: binary won't start AND CLI subcommands are unreachable

In a crashloop (e.g., the binary panics before reaching the admin-CLI dispatch,
or you cannot get a shell on the container), the only path is degraded mode:

```bash
ACTALOG_SKIP_PROTECTED_INVARIANT=true ./bin/actalog
```

This starts the server with the boot invariant bypassed. The trade-off:

- **Admin user-write endpoints return HTTP 503** (`degradedAdminWriteGuard` middleware)
- **`/health` returns `503 {"status":"degraded"}`**
- **ERROR log every 60 seconds** (`protected_invariant_degraded heartbeat: ...`)
- L3 triggers may be absent — the only active protection is L1 and L2

Once the app is running, use the CLI to repair:

```bash
./bin/actalog admin reapply-protected-migrations --confirm
```

Then restart without the env var. See §10 for full details on degraded mode.

### 8.7 Scenario: drift between Go source and migration SQL (CI failure)

This is not a runtime emergency — the build fails in CI but no running system is
affected.

**Step 1 — Regenerate the frontend guard and check which file(s) are out of date:**

```bash
cd /path/to/repo
make gen-protected-emails
git diff
```

**Step 2 — If the Go source changed without a matching migration, create the migration:**

```bash
make migrate-create name=update_protected_users
# Edit the new migration file with the updated trigger SQL
```

**Step 3 — If the SQL constants in `protected_triggers_sql.go` are stale, update them** to
match the migration SQL, then re-run the generator.

**Step 4 — Verify the build passes:**

```bash
make build
make gen-protected-emails
git diff --exit-code
```

---

## 9. Audit-log forensics

### 9.1 Event types

Three event types cover protected-user attack attempts. Each is tagged by the *rejecting
layer*, not by all layers traversed:

| Event type constant | Value | Layer | Meaning |
|---------------------|-------|-------|---------|
| `EventProtectedUserAttackHTTP` | `protected_user_attack_http` | L1 | Blocked by the HTTP middleware. Normal path for browser and script callers. |
| `EventProtectedUserAttackService` | `protected_user_attack_service` | L2 | Blocked by the service guard. Indicates the request reached the service layer without going through the normal admin subrouter (e.g., a cron job, internal call, or route not protected by the subrouter middleware). |
| `EventProtectedUserAttackDB` | `protected_user_attack_db` | L3 | Rejected by the database trigger. **Any single `_db` event means L1 and L2 were both bypassed.** Treat this as a high-severity incident. |

The `EventPasswordResetForcedByAdmin` (`password_reset_forced_by_admin`) event is not
an attack event — it is the normal audit record when an admin legitimately triggers a
password reset for another (non-protected) user.

### 9.2 Event structure

Each event in the `audit_logs` table has:

```text
id            — auto-increment
user_id       — actor (the admin who attempted the write), NULL if unauthenticated
target_user_id — the protected user's ID
event_type    — one of the three values above
ip_address    — remote address of the HTTP request
user_agent    — User-Agent header from the HTTP request (NULL for L2/L3 catches)
details       — JSON string; for L1 events includes "path", "method", "user_agent", "referer"
created_at    — UTC timestamp
```

L1 events always have `user_agent` and `details.path` populated (from the HTTP request).
L2 events have NULL `user_agent` and NULL `ip_address` — the write was attempted from
inside the Go process, not from an HTTP request. L3 events follow the same shape as L2
(caught by the service layer when the DB trigger fires).

You can distinguish browser-originated requests from scripted callers at query time
using the `user_agent` or `details.user_agent` column.

### 9.3 Query patterns

#### Find all protected-user attack events in the last 7 days

```sql
SELECT
    al.id,
    al.created_at,
    al.event_type,
    al.user_id       AS actor_id,
    al.target_user_id,
    al.ip_address,
    al.user_agent,
    al.details
FROM audit_logs al
WHERE al.event_type IN (
    'protected_user_attack_http',
    'protected_user_attack_service',
    'protected_user_attack_db'
)
  AND al.created_at >= NOW() - INTERVAL '7 days'
ORDER BY al.created_at DESC;
```

For SQLite (no `INTERVAL` syntax):

```sql
SELECT
    id,
    created_at,
    event_type,
    user_id,
    target_user_id,
    ip_address,
    user_agent,
    details
FROM audit_logs
WHERE event_type IN (
    'protected_user_attack_http',
    'protected_user_attack_service',
    'protected_user_attack_db'
)
  AND created_at >= datetime('now', '-7 days')
ORDER BY created_at DESC;
```

#### Find `_db` events only — these are high-severity (L1+L2 bypassed)

```sql
SELECT
    al.id,
    al.created_at,
    al.user_id AS actor_id,
    u.email    AS actor_email,
    al.ip_address,
    al.details
FROM audit_logs al
LEFT JOIN users u ON u.id = al.user_id
WHERE al.event_type = 'protected_user_attack_db'
ORDER BY al.created_at DESC;
```

#### Find non-browser callers (scripted attacks) in the last 7 days

The `user_agent` column is NULL for L2/L3 catches. For L1 catches from scripts, the
User-Agent is typically `curl/*`, `python-requests/*`, `Go-http-client/*`, or empty.
This query surfaces any L1 event where the user-agent is absent or non-browser:

```sql
SELECT
    al.id,
    al.created_at,
    al.user_id,
    al.ip_address,
    al.user_agent,
    al.details
FROM audit_logs al
WHERE al.event_type = 'protected_user_attack_http'
  AND al.created_at >= NOW() - INTERVAL '7 days'
  AND (
      al.user_agent IS NULL
      OR al.user_agent NOT LIKE 'Mozilla/%'
  )
ORDER BY al.created_at DESC;
```

SQLite version:

```sql
SELECT
    id,
    created_at,
    user_id,
    ip_address,
    user_agent,
    details
FROM audit_logs
WHERE event_type = 'protected_user_attack_http'
  AND created_at >= datetime('now', '-7 days')
  AND (
      user_agent IS NULL
      OR user_agent NOT LIKE 'Mozilla/%'
  )
ORDER BY created_at DESC;
```

#### Correlate an actor across all their protected-user events

Replace `<actor_id>` with the numeric `user_id` from the event you are investigating:

```sql
SELECT
    al.id,
    al.created_at,
    al.event_type,
    al.ip_address,
    al.user_agent,
    al.details
FROM audit_logs al
WHERE al.user_id = <actor_id>
  AND al.event_type IN (
      'protected_user_attack_http',
      'protected_user_attack_service',
      'protected_user_attack_db'
  )
ORDER BY al.created_at DESC;
```

#### Count events by type (monitoring / alerting baseline)

```sql
SELECT
    event_type,
    COUNT(*) AS occurrences,
    MIN(created_at) AS first_seen,
    MAX(created_at) AS last_seen
FROM audit_logs
WHERE event_type IN (
    'protected_user_attack_http',
    'protected_user_attack_service',
    'protected_user_attack_db'
)
GROUP BY event_type
ORDER BY occurrences DESC;
```

### 9.4 Alert thresholds

| Event type | Alert rule |
|------------|------------|
| `protected_user_attack_db` | **Alert on `count >= 1`** — a single DB-layer event means L1 and L2 were bypassed |
| `protected_user_attack_service` | Alert on unexpected spike; occasional events from automation are low-risk |
| `protected_user_attack_http` | Baseline expected; alert on unusual spike or non-browser user-agent |
| `protected_invariant_degraded` | Alert on any occurrence — means the binary is running in degraded mode |

### 9.5 Distinguishing UI vs script at query time

The details JSON field for L1 events stores the original User-Agent string and
referer. You can parse it at query time:

**PostgreSQL** — extract user_agent from the JSON details column:

```sql
SELECT
    id,
    created_at,
    details::json->>'user_agent' AS ua,
    details::json->>'referer'   AS referer,
    details::json->>'path'      AS path
FROM audit_logs
WHERE event_type = 'protected_user_attack_http'
ORDER BY created_at DESC
LIMIT 50;
```

**SQLite** — use `json_extract`:

```sql
SELECT
    id,
    created_at,
    json_extract(details, '$.user_agent') AS ua,
    json_extract(details, '$.referer')   AS referer,
    json_extract(details, '$.path')      AS path
FROM audit_logs
WHERE event_type = 'protected_user_attack_http'
ORDER BY created_at DESC
LIMIT 50;
```

**MySQL/MariaDB** — use `JSON_UNQUOTE(JSON_EXTRACT(...))`:

```sql
SELECT
    id,
    created_at,
    JSON_UNQUOTE(JSON_EXTRACT(details, '$.user_agent')) AS ua,
    JSON_UNQUOTE(JSON_EXTRACT(details, '$.referer'))   AS referer,
    JSON_UNQUOTE(JSON_EXTRACT(details, '$.path'))      AS path
FROM audit_logs
WHERE event_type = 'protected_user_attack_http'
ORDER BY created_at DESC
LIMIT 50;
```

A row with a `Mozilla/5.0 ...` user-agent and a `referer` of `https://yourdomain.com/admin/users`
is consistent with an admin clicking the (expected-to-be-disabled) UI. Anything else
warrants investigation.

---

## 10. Degraded mode

### 10.1 What it is

Degraded mode is the escape hatch when the binary cannot start because the boot
invariant fails (HARD check on L3 trigger state), but you need the app to serve
traffic while you repair the invariant. It is enabled by a single environment variable:

```bash
ACTALOG_SKIP_PROTECTED_INVARIANT=true ./bin/actalog
```

Or in a `.env` file:

```yaml
ACTALOG_SKIP_PROTECTED_INVARIANT=true
```

### 10.2 What works in degraded mode

- All read endpoints (`GET`, `HEAD`) work normally
- Non-admin write endpoints (user registration, login, workout logging, etc.) work
  normally
- Admin read endpoints work normally
- The `/health` endpoint responds `503 {"status":"degraded","version":"..."}`

### 10.3 What does NOT work in degraded mode

- **All admin user-write endpoints return HTTP 503.** This includes:
  `PATCH /api/admin/users/{id}`, `DELETE /api/admin/users/{id}`,
  `POST /api/admin/users/{id}/disable`, `POST /api/admin/users/{id}/unlock`,
  `POST /api/admin/users/{id}/force-password-reset`, and any other non-GET/HEAD
  method under `/api/admin/*`.
- This is enforced by `degradedAdminWriteGuard` middleware, which short-circuits
  the request before it reaches L1 or L2. L3 may be absent — write operations are
  blocked at the HTTP layer instead.

The response body for a blocked admin write in degraded mode is:

```json
{
  "error": "protected_invariant_degraded",
  "message": "Admin user-write endpoints are temporarily disabled while the protected-user invariant is broken. Operator: see logs."
}
```

### 10.4 Monitoring while in degraded mode

**Log heartbeat** — the binary logs an ERROR every 60 seconds:

```text
[ERROR] protected_invariant_degraded heartbeat: invariant: trigger "protected_users_no_update" missing ...
```

This is designed to be alertable by any log aggregator that watches for `[ERROR]`
lines or the substring `protected_invariant_degraded`.

**Health probe** — `/health` returns `503`:

```bash
curl -s http://localhost:8080/health
# {"status":"degraded","version":"1.3.0"}
```

**Audit event** — on startup in degraded mode, the binary writes a
`protected_invariant_degraded` event to `audit_logs` with the failure reason in the
`details` JSON. This creates a persistent record of when and why the system entered
degraded mode.

### 10.5 When to use degraded mode

Use degraded mode only in these situations:

1. **The binary refuses to start and you need the app serving traffic** (e.g., the
   L3 trigger was dropped by a migration rollback and the triggers need to be
   restored, but the app cannot wait for a maintenance window).
2. **The boot invariant has a false-positive bug** (i.e., the trigger exists and
   works correctly but the invariant check is reporting a failure). In this case,
   file an incident report and wait for a fix release.

Do NOT leave degraded mode enabled permanently. It disables all admin user-write
operations, which includes legitimate ones.

### 10.6 Exit criteria

Exit degraded mode by:

1. Repairing the invariant failure (see §8 — recovery playbook)
2. Removing `ACTALOG_SKIP_PROTECTED_INVARIANT=true` from the environment
3. Restarting the binary

Confirm the exit:

```bash
curl -s http://localhost:8080/health
# {"status":"healthy","version":"1.3.0"}

./bin/actalog admin verify-protected-users
# ✓ all checks passed
```

If the health endpoint returns `"status":"healthy"` and the verify CLI shows all three
checks passing, degraded mode is no longer active and admin write endpoints are
restored.

---

*This document is the single source of truth for the protected-user system. For the
quick-reference 3-AM incident-response playbook, see `docs/security/PROTECTED_USERS_RECOVERY.md`.*
