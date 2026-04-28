# Admin User Edit Screen v1.3.0 — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship v1.3.0: a tabbed admin user-edit screen at `/admin/users/:id/edit` with a working Profile tab, plus the four-layer defense-in-depth system (L1 middleware, L2 service guard, L3 DB trigger, L4 audit log) that protects `br8kwall@gmail.com` from any modification — including from shells, scripts, or compromised admin accounts.

**Architecture:** Vue 3 SPA with `v-tabs` + `v-window` per existing admin views. Go backend with Chi router, existing clean-architecture layers (handler → service → repository → domain). New `pkg/security/protected_users.go` is the single source of truth for the protected list; a code generator produces the matching frontend file. The L3 trigger in the database makes rolling back the binary keep security on. Boot-time invariant fails closed if any defense layer is broken.

**Tech Stack:** Go 1.25+, Chi router, sqlite3/postgres/mysql via per-dialect SQL. Vue 3 + Vuetify 3 + Pinia + Vitest. Per-DB CI matrix already exists (`Integration tests (DB matrix)` job).

**Source spec:** `docs/superpowers/specs/2026-04-28-admin-user-edit-design.md` — read this first; the plan references its sections rather than duplicating detail.

---

## Test helpers introduced by this plan

Several tasks need shared helpers that wrap the project's `os/exec` machinery for invoking the compiled `actalog` binary from tests. Add these once to `test/integration/cli_test_helpers.go` (or extend the existing helpers file):

- `runActalogCLI(t, args ...string) (stdout, stderr string, exit int)` — wraps the project's existing process-runner, used to invoke `actalog admin ...` subcommands from integration tests.
- `bootActalogServer(t, opts ...BootOpt) *RunningServer` — starts the binary in the background with given env, returns a handle with `.Stop()`, `.Logs()`, `.HTTP("/path")`.
- `triggerExists(t, db, name) bool` — per-dialect catalog query (sqlite3 → `sqlite_master`, postgres → `pg_trigger`, mysql → `INFORMATION_SCHEMA.TRIGGERS`).
- `mustExec(t, db, sql)` and `nowSQL(db)` — driver-aware SQL helpers.

These exist as conventional patterns; check `test/integration/` for whatever's already there before re-implementing.

---

## File structure

### Backend (Go)

| File | Status | Responsibility |
|------|--------|----------------|
| `pkg/security/protected_users.go` | new | Protected-emails registry + `IsProtectedEmail()` |
| `pkg/security/protected_users_test.go` | new | Helper unit tests |
| `pkg/middleware/protected_user.go` | new | L1 — HTTP middleware |
| `pkg/middleware/protected_user_test.go` | new | L1 unit tests |
| `internal/service/admin_user_service.go` | new | L2 + `UpdateProfile` + `ForcePasswordReset` |
| `internal/service/admin_user_service_test.go` | new | L2 unit tests |
| `internal/handler/admin_user_handler.go` | modify | New `UpdateProfile` and `ForcePasswordReset` methods |
| `internal/handler/admin_user_handler_test.go` | modify | Tests for new methods |
| `internal/handler/errors.go` | modify | Map `ErrProtectedUser`→403, `ErrConflict`→409 |
| `internal/domain/audit_log.go` | modify | Add `EventProtectedUserAttack*` and `EventPasswordResetForcedByAdmin` |
| `internal/domain/errors.go` | new | `ErrProtectedUser`, `ErrConflict` sentinels |
| `internal/repository/migrations.go` | modify | Migration 0.35.0 |
| `internal/repository/protected_triggers_sql.go` | new | Per-dialect trigger SQL constants |
| `cmd/actalog/main.go` | modify | Wire L1, boot invariant, env-var degraded mode |
| `cmd/actalog/boot_invariant.go` | new | Three-check boot-time invariant |
| `cmd/actalog/boot_invariant_test.go` | new | Boot invariant unit tests |
| `cmd/actalog/admin_cli.go` | new | `verify-protected-users` + `reapply-protected-migrations` |
| `cmd/actalog/admin_cli_test.go` | new | CLI tests |
| `cmd/gen-protected-emails/main.go` | new | Frontend-file generator |
| `cmd/gen-protected-emails/main_test.go` | new | Generator tests |

### Database / recovery

| File | Status | Responsibility |
|------|--------|----------------|
| `scripts/recover/sql/sqlite/protected_users.sql` | new | Stand-alone SQL (lockstep with migration) |
| `scripts/recover/sql/postgres/protected_users.sql` | new | Same, PostgreSQL |
| `scripts/recover/sql/mysql/protected_users.sql` | new | Same, MySQL/MariaDB |
| `scripts/recover/restore-protected-triggers.sh` | new | Last-resort recovery wrapper |
| `scripts/recover/lockstep_test.go` | new | CI lockstep check |
| `scripts/recover/README.md` | new | When to use; per-driver examples |

### Frontend

| File | Status | Responsibility |
|------|--------|----------------|
| `web/src/utils/protectedUsers.js` | new (auto-generated) | Generated mirror of Go constant |
| `web/src/utils/protectedUsers.test.js` | new | Generated-module tests |
| `web/src/utils/axios.js` | modify | 403 protected_user + 503 degraded interceptors |
| `web/src/utils/axios.test.js` | modify | New cases |
| `web/src/components/admin/user-edit/composables/useUserDraft.js` | new | Draft+save state |
| `web/src/components/admin/user-edit/composables/useUserDraft.test.js` | new | Composable tests |
| `web/src/components/admin/user-edit/composables/useProtectedUserStatus.js` | new | Reactive isProtected wrapper |
| `web/src/components/admin/user-edit/ProtectedUserBanner.vue` | new | Read-only banner |
| `web/src/components/admin/user-edit/TabFooterActions.vue` | new | Save/Discard/dirty indicator |
| `web/src/components/admin/user-edit/ProfileTab.vue` | new | The v1.3.0 tab |
| `web/src/components/admin/user-edit/ProfileTab.test.js` | new | Tab tests |
| `web/src/views/AdminUserEditView.vue` | new | Page shell |
| `web/src/views/AdminUserEditView.test.js` | new | Shell tests |
| `web/src/views/AdminUsersView.vue` | modify | Edit button + isProtected guard |
| `web/src/router/index.js` | modify | Register new route |

### Integration tests

| File | Status | Responsibility |
|------|--------|----------------|
| `test/integration/protected_users_test.go` | new | L3 trigger correctness, per-DB matrix |
| `test/integration/protected_users_recovery_test.go` | new | Recovery matrix per spec §6.4 |
| `test/integration/protected_users_layered_defense_test.go` | new | Adversarial bypass tests per spec §6.5 |
| `test/integration/cli_test_helpers.go` | new or extend | The shared CLI/server helpers listed at top |

### CI & build

| File | Status | Responsibility |
|------|--------|----------------|
| `.github/CODEOWNERS` | new or modify | Require security review for `pkg/security/**`, migrations, security docs |
| `.github/workflows/ci.yml` | modify | Generator drift check; lockstep check; doc code-block lint |
| `Makefile` | modify | `gen-protected-emails` target |

### Documentation

| File | Status | Responsibility |
|------|--------|----------------|
| `docs/security/PROTECTED_USERS.md` | new | Master runbook (spec §7.1) |
| `docs/security/PROTECTED_USERS_RECOVERY.md` | new | 3-AM incident playbook |
| `docs/security/THREAT_MODEL.md` | new | App-wide threat model |
| `docs/security/PROTECTED_USERS_LIST.md` | new (auto-generated) | Table view of who's protected |
| `docs/USER_PERMISSIONS.md` | modify | New endpoints + protected-user behavior |
| `docs/ARCHITECTURE.md` | modify | "Defense-in-Depth" subsection |
| `docs/DATABASE_SCHEMA.md` | modify | "Triggers" subsection; bump schema version |
| `docs/TESTING.md` | modify | "Security Tests" subsection |
| `docs/CHANGELOG.md` | modify | v1.3.0 entry |
| `docs/TODO.md` | modify | Mark in-progress; v1.3.1/v1.3.2 entries |
| `CLAUDE.md` | modify | Replace inline protected-users paragraph with doc pointer |
| `README.md` | modify | Version stamp |

### Release artifacts

| File | Status | Responsibility |
|------|--------|----------------|
| `pkg/version/version.go` | modify | Bump to 1.3.0, build 50 |
| `web/package.json` | modify | Bump to 1.3.0 |

---

## Phases

| Phase | Tasks | Outcome |
|-------|-------|---------|
| A — Foundations | 1–3 | Protected list, error sentinels, generator |
| B — Database (L3) + boot check | 4–9 | Migration, recovery scripts, boot invariant, admin CLI, recovery test matrix |
| C — Application (L1+L2) | 10–14 | Middleware, service methods, handlers, wired in main.go, adversarial tests |
| D — Frontend foundation | 15–18 | Generated module, composables, axios interceptors |
| E — Frontend UI | 19–22 | Banner, footer, ProfileTab, AdminUserEditView, AdminUsersView entry |
| F — Recovery + docs | 23–28 | CODEOWNERS, drift checks, all docs, doc CI lint |
| G — Release | 29–30 | Version bump, smoke test, PR |

---

## Phase A — Foundations

### Task 1: Protected emails registry + `IsProtectedEmail()`

**Files:** `pkg/security/protected_users.go` (new), `pkg/security/protected_users_test.go` (new)

- [ ] **Step 1: Failing test.** Create `pkg/security/protected_users_test.go` with these named cases (test-table or `t.Run` per case):
  - `ExactMatch` — `IsProtectedEmail("br8kwall@gmail.com")` is true
  - `CaseInsensitive` — three variants (`BR8KWALL@gmail.com`, mixed-case, all-uppercase domain) all true
  - `TrimsWhitespace` — `"  br8kwall@gmail.com  "` is true
  - `DoesNotMatchPlusAddressing` — `br8kwall+test@gmail.com` is false (plus-addressing routes to same mailbox but is a different identity)
  - `NonProtectedReturnsFalse` — `alice@example.com` is false
  - `EmptyReturnsFalse` — empty string is false
  - `LowercaseInvariant` — every entry in `ProtectedEmails` map is lowercase (so the lowercase-on-input strategy can't miss a typo)

- [ ] **Step 2:** `go test -count=1 ./pkg/security/` — expect FAIL (package doesn't exist).

- [ ] **Step 3: Implementation.** Create `pkg/security/protected_users.go`:
  - `ProtectedEmails map[string]struct{}` with one entry: `"br8kwall@gmail.com": {}`
  - Doc comment that calls out the **three coupled artifacts** (this Go map, the migration trigger SQL, the auto-generated `web/src/utils/protectedUsers.js`) — adding/removing requires updating all three in one PR (see `docs/security/PROTECTED_USERS.md`)
  - `IsProtectedEmail(email string) bool` — `strings.ToLower(strings.TrimSpace(email))` → map lookup

- [ ] **Step 4:** `go test -count=1 -v ./pkg/security/` — expect PASS.

- [ ] **Step 5:** Verify 100% coverage: `go test -count=1 -coverprofile=/tmp/sec.cov ./pkg/security/ && go tool cover -func=/tmp/sec.cov` shows `IsProtectedEmail` at 100%.

- [ ] **Step 6:** Commit:
```
feat(security): add ProtectedEmails registry and IsProtectedEmail helper

Single source of truth for the protected-user list. Coupled artifacts
(migration SQL, frontend JS) follow in subsequent commits.
```

### Task 2: Audit-event constants + error sentinels + handler mapping

**Files:** `internal/domain/audit_log.go` (modify), `internal/domain/audit_log_protected_test.go` (new), `internal/domain/errors.go` (new), `internal/handler/errors.go` (modify), `internal/handler/errors_test.go` (modify or new)

- [ ] **Step 1: Failing test for event constants.** Create `internal/domain/audit_log_protected_test.go` asserting that `EventProtectedUserAttackHTTP`, `EventProtectedUserAttackService`, `EventProtectedUserAttackDB`, and `EventPasswordResetForcedByAdmin` are defined and have distinct string values.

- [ ] **Step 2:** Run — expect FAIL (undefined identifiers).

- [ ] **Step 3:** Append constants to the existing event-type `const (...)` block in `internal/domain/audit_log.go`. Values:
  - `"protected_user_attack_http"` (L1 blocked)
  - `"protected_user_attack_service"` (L2 blocked)
  - `"protected_user_attack_db"` (L3 caught via service-layer error pattern matching)
  - `"password_reset_forced_by_admin"`

- [ ] **Step 4:** Run constants test — expect PASS.

- [ ] **Step 5:** Create `internal/domain/errors.go` with two `errors.New(...)` sentinels:
  - `ErrProtectedUser = errors.New("protected user: modifications blocked")`
  - `ErrConflict = errors.New("resource modified concurrently")`

- [ ] **Step 6:** Modify `internal/handler/errors.go` so the existing handler error mapper checks `errors.Is(err, domain.ErrProtectedUser)` first and writes a structured 403 body with these JSON fields:
  - `error: "protected_user"`
  - `message: "This account is system-reserved and cannot be modified."`
  - `documentation_url: "/docs/protected-users"`
  - `support_contact: ""` (configurable later, omitted if empty)

  Same shape for `ErrConflict` → 409 with `error: "conflict"`. If `ErrorResponse` doesn't have `Error`/`DocumentationURL` fields, extend it. Add a `WriteJSONError(w, status, body)` helper if missing.

- [ ] **Step 7:** Add tests to `internal/handler/errors_test.go`:
  - `TestWriteError_MapsErrProtectedUserTo403` — asserts status 403 and body contains `"error":"protected_user"` + non-empty `documentation_url`
  - `TestWriteError_MapsErrConflictTo409`

- [ ] **Step 8:** `go test -count=1 ./internal/handler/ ./internal/domain/` — expect PASS.

- [ ] **Step 9:** Commit:
```
feat(domain): add protected-user audit events and ErrProtectedUser/ErrConflict sentinels
```

### Task 3: Frontend-file generator `cmd/gen-protected-emails`

**Files:** `cmd/gen-protected-emails/main.go` (new), `cmd/gen-protected-emails/main_test.go` (new), `Makefile` (modify), `web/src/utils/protectedUsers.js` (new, auto-generated)

- [ ] **Step 1: Failing test.** `cmd/gen-protected-emails/main_test.go` covering:
  - `Generate_ProducesEs6ModuleWithSetAndHelper` — the output string contains the expected header comment, `export const PROTECTED_EMAILS = new Set([...])`, the seeded emails, and an `export function isProtectedEmail(email)` body that calls `String(email).trim().toLowerCase()`
  - `Generate_DeterministicOrder` — calling `generate()` 20 times with the same map yields byte-identical output (sort by email)

- [ ] **Step 2:** `go test -count=1 ./cmd/gen-protected-emails/` — expect FAIL.

- [ ] **Step 3: Implementation.** `cmd/gen-protected-emails/main.go`:
  - `generate(map[string]struct{}) string` — sorts keys, emits the ES6 module text with header comment naming the source path and a "do not edit by hand; run `make gen-protected-emails`" warning
  - `main()` — reads `security.ProtectedEmails`, writes to `web/src/utils/protectedUsers.js`
  - The `isProtectedEmail` body in the emitted JS: `if (!email) return false; return PROTECTED_EMAILS.has(String(email).trim().toLowerCase())`

- [ ] **Step 4:** `go test -count=1 ./cmd/gen-protected-emails/` — expect PASS.

- [ ] **Step 5:** Run `go run ./cmd/gen-protected-emails/` — produces `web/src/utils/protectedUsers.js`. Inspect it: header comment, Set with `'br8kwall@gmail.com'`, `isProtectedEmail` function.

- [ ] **Step 6:** Add `Makefile` target:
```make
.PHONY: gen-protected-emails
gen-protected-emails:
	@go run ./cmd/gen-protected-emails/
```
If a `build` target exists, add `gen-protected-emails` as a prerequisite.

- [ ] **Step 7:** Commit:
```
feat(build): add gen-protected-emails generator + emit web/src/utils/protectedUsers.js
```

---

## Phase B — Database (L3) + boot check

### Task 4: Migration `0.35.0_add_protected_user_triggers` (per-dialect)

**Files:** `internal/repository/migrations.go` (modify), `internal/repository/protected_triggers_sql.go` (new), `test/integration/protected_users_test.go` (new), `test/integration/cli_test_helpers.go` (new or extend), `docs/DATABASE_SCHEMA.md` (modify)

The codebase already has a per-dialect migration mechanism. Read existing migrations (e.g. `0.34.0`) before writing new code so the new one matches the pattern (struct fields, `Up`/`Down` shape, multi-statement helper).

- [ ] **Step 1: Failing integration tests.** Create `test/integration/protected_users_test.go` with these named tests (each must run on sqlite3, postgres, mysql via the matrix):
  - `TestL3_TriggerExistsAfterMigration` — both `protected_users_no_update` and `protected_users_no_delete` exist in the DB catalog
  - `TestL3_UPDATEOnProtectedRowRejectedByDB` — insert `br8kwall@gmail.com`, then `UPDATE users SET name='hacked' WHERE email='br8kwall@gmail.com'` returns an error containing `"protected user"`
  - `TestL3_DELETEOnProtectedRowRejectedByDB` — same shape with DELETE
  - `TestL3_UPDATEOnNonProtectedRowSucceeds` — UPDATE of `alice@example.com` succeeds
  - `TestL3_TriggerErrorMessageMatchesContract` — error contains the exact substring `protected user: writes blocked at db layer` (this string is the contract that L4's pattern matcher relies on)
  - `TestL3_TriggerSurvivesRollback` — a transaction containing one allowed UPDATE and one blocked UPDATE rolls back cleanly without leaving the trigger in a broken state

- [ ] **Step 2:** Run on each driver (`DB_DRIVER=sqlite3 go test -count=1 -run TestL3_ ./test/integration/` etc.) — expect FAIL (triggers don't exist).

- [ ] **Step 3: Add migration.** Create `internal/repository/protected_triggers_sql.go` holding three per-dialect const strings (`SQLiteProtectedTriggers`, `PostgresProtectedTriggers`, `MySQLProtectedTriggers`). Each wraps a body marked with `-- LOCKSTEP-START <dialect>` and `-- LOCKSTEP-END <dialect>` comment markers (used by the lockstep test in Task 7).

  Per-dialect SQL (full text in spec §3.3):
  - **SQLite:** `CREATE TRIGGER ... BEFORE UPDATE ... WHEN OLD.email IN ('br8kwall@gmail.com') BEGIN SELECT RAISE(ABORT, 'protected user: writes blocked at db layer'); END;` + same for DELETE.
  - **Postgres:** `CREATE OR REPLACE FUNCTION block_protected_users() ... RAISE EXCEPTION` + `CREATE TRIGGER ... EXECUTE FUNCTION block_protected_users()` for both UPDATE and DELETE.
  - **MySQL:** `CREATE TRIGGER ... BEFORE UPDATE ... IF OLD.email = 'br8kwall@gmail.com' THEN SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'protected user: writes blocked at db layer'; END IF; END;` + same for DELETE.

  Append the migration to `internal/repository/migrations.go` registry: `Version: "0.35.0"`, `Description: "Add protected user triggers (L3 defense layer)"`, `Up` dispatches per `driver` to the correct const, `Down` drops both triggers (forbidden by policy in production but provided for dev rollback).

- [ ] **Step 4:** Run integration tests on all three dialects. Expect PASS.

- [ ] **Step 5:** Update `docs/DATABASE_SCHEMA.md`: bump schema-version mention to `0.35.0`; add a one-line note pointing at `docs/security/PROTECTED_USERS.md`.

- [ ] **Step 6:** Commit:
```
feat(db): add migration 0.35.0 with protected-user triggers (L3)

Per-dialect BEFORE UPDATE / BEFORE DELETE triggers. Error message text
is contract-locked — see L4 service-error pattern matcher. Tests verify
trigger existence and rejection on sqlite3, postgres, mysql.
```

### Task 5: Boot-time invariant function

**Files:** `cmd/actalog/boot_invariant.go` (new), `cmd/actalog/boot_invariant_test.go` (new)

Three sub-checks per spec §3.5:
1. L3 triggers exist in DB catalog (HARD)
2. Simulated UPDATE inside `BEGIN; ... ROLLBACK` rejects (HARD)
3. Protected user row exists (SOFT on fresh install with zero rows; HARD when other users exist)

- [ ] **Step 1: Failing tests.** `cmd/actalog/boot_invariant_test.go`:
  - `TestBootInvariant_FreshInstallNoUsersYet` — fresh DB, no users; check 3 is soft-WARN; function returns nil error
  - `TestBootInvariant_HardFailWhenTriggerMissing` — drop trigger, function returns non-nil error naming the missing trigger
  - `TestBootInvariant_HardFailWhenProtectedRowMissingAndOtherUsersExist` — insert non-protected user but no protected row; function returns hard error
  - `TestBootInvariant_PassesWithProtectedRowAndTriggers` — insert `br8kwall@gmail.com`; all three checks pass

- [ ] **Step 2:** `go test -count=1 -run TestBootInvariant ./cmd/actalog/` — expect FAIL.

- [ ] **Step 3: Implementation.** `cmd/actalog/boot_invariant.go`:
  - `InvariantReport` struct with `Check1TriggersExist`, `Check2TriggersFire`, `Check3ProtectedRowsExist`, `SoftWarnings []string`
  - `VerifyProtectedUserInvariant(db *sql.DB, driver string) (*InvariantReport, error)` runs all three checks
  - Constant `triggerErrorContract = "protected user: writes blocked at db layer"` — the matchable substring
  - Hard-failure error messages name the specific failing check AND include the recovery command literal: `"./bin/actalog admin reapply-protected-migrations --confirm"`
  - Per-dialect `triggerExistsForDriver(db, driver, name) (bool, error)` queries the catalog: sqlite3 → `sqlite_master`, postgres → `pg_trigger`, mysql → `INFORMATION_SCHEMA.TRIGGERS`

  The simulated UPDATE for Check 2 runs inside a transaction that's always rolled back. To fire the trigger without affecting real data: `BEGIN; INSERT a probe row with a non-protected email; UPDATE the probe to set its email TO the protected value (this fires the trigger because OLD.email was non-protected but NEW row matches); assert the UPDATE returned an error containing the contract substring; ROLLBACK.` (Adjust if a simpler approach exists for a given dialect.)

- [ ] **Step 4:** Run boot-invariant tests — expect PASS.

- [ ] **Step 5:** Commit:
```
feat(boot): add VerifyProtectedUserInvariant with three boot-time checks
```

### Task 6: Wire boot invariant + ACTALOG_SKIP_PROTECTED_INVARIANT into main.go

**Files:** `cmd/actalog/main.go` (modify), `cmd/actalog/degraded_middleware.go` (new)

- [ ] **Step 1:** Locate the post-migration startup spot (right after the existing `Migrations completed` log line).

- [ ] **Step 2:** Add the invariant call. On hard failure with no env var: `appLogger.Fatal(...)` printing the diagnostic and recovery options block from spec §2.6. With env var: log ERROR, set `degraded := true`, write audit event `protected_invariant_degraded`, start a 60-second-tick goroutine that re-logs at ERROR level for alerting.

- [ ] **Step 3:** Update the `/health` handler to set `status: "degraded"` and HTTP 503 when `degraded` is true (vs. `"healthy"` and 200 otherwise).

- [ ] **Step 4:** Create `cmd/actalog/degraded_middleware.go` with `degradedAdminWriteGuard(degraded *bool) func(http.Handler) http.Handler` — for non-GET/HEAD requests, returns 503 with body `{"error":"protected_invariant_degraded","message":"..."}`.

- [ ] **Step 5:** Manual smoke test: build, boot, curl `/health` (healthy), break trigger via direct SQL, restart (refuses to boot), set env var and restart (boots degraded; `/health` returns 503).

- [ ] **Step 6:** Commit:
```
feat(boot): wire protected-user invariant + ACTALOG_SKIP_PROTECTED_INVARIANT degraded mode
```

### Task 7: Recovery SQL scripts + lockstep CI check

**Files:** `scripts/recover/sql/{sqlite,postgres,mysql}/protected_users.sql` (new), `scripts/recover/restore-protected-triggers.sh` (new), `scripts/recover/lockstep_test.go` (new), `scripts/recover/README.md` (new), `.github/workflows/ci.yml` (modify)

- [ ] **Step 1: Failing test.** `scripts/recover/lockstep_test.go` — `TestLockstep_RecoveryScriptsMatchMigration` for each dialect:
  - Reads the migration source's per-dialect const, extracts the body between `-- LOCKSTEP-START` and `-- LOCKSTEP-END` markers
  - Reads the recovery script file
  - Normalises whitespace + comments (collapse runs of whitespace, strip blank lines and SQL comments) and compares
  - Fails if the bodies differ

- [ ] **Step 2:** Run — expect FAIL (recovery scripts don't exist yet).

- [ ] **Step 3:** Create the per-dialect recovery script files. Body is the exact SQL between the lockstep markers from `internal/repository/protected_triggers_sql.go`, minus the marker comments themselves.

- [ ] **Step 4:** Run lockstep test — expect PASS.

- [ ] **Step 5:** Create `scripts/recover/restore-protected-triggers.sh` — a small shell wrapper that takes `--driver=<sqlite3|postgres|mysql>` and `--conn=<...>`, dispatches to the correct CLI (`sqlite3`, `psql`, `mysql`), and applies the matching `.sql` file. Print a final hint pointing at `actalog admin verify-protected-users` to confirm. `chmod +x`.

- [ ] **Step 6:** Write `scripts/recover/README.md` — when to use vs. the in-binary CLI; per-driver invocation; safety warnings.

- [ ] **Step 7:** Add CI step to `.github/workflows/ci.yml`:
```yaml
      - name: Lockstep check (recovery scripts ↔ migration SQL)
        run: go test -count=1 ./scripts/recover/...
```

- [ ] **Step 8:** Commit:
```
feat(recover): raw-SQL recovery scripts + CI lockstep check
```

### Task 8: Admin CLI subcommands `verify-protected-users` and `reapply-protected-migrations`

**Files:** `cmd/actalog/admin_cli.go` (new), `cmd/actalog/admin_cli_test.go` (new), `cmd/actalog/main.go` (modify)

- [ ] **Step 1: Failing tests.** `cmd/actalog/admin_cli_test.go`:
  - `TestAdminVerifyCLI_PassesOnHealthyDB` — sets up a migrated DB with `br8kwall@gmail.com`, calls `AdminVerifyProtectedUsers(db, driver, verbose=true, &stdout)`, asserts exit 0 and `stdout` contains "Check 1/3" and "✓"
  - `TestAdminVerifyCLI_FailsAndNamesCheck` — drops a trigger, asserts exit non-zero and stdout names `protected_users_no_update`
  - `TestAdminReapplyCLI_RestoresDroppedTrigger` — drops trigger, calls `AdminReapplyProtectedMigrations(db, driver, confirm=true, &stdout)`, asserts subsequent `VerifyProtectedUserInvariant` passes AND a real UPDATE on a freshly-inserted protected row is rejected
  - `TestAdminReapplyCLI_RequiresConfirm` — `confirm=false` returns error without modifying anything

- [ ] **Step 2:** Run — expect FAIL.

- [ ] **Step 3: Implementation.** `cmd/actalog/admin_cli.go`:
  - `AdminVerifyProtectedUsers(db, driver, verbose bool, out io.Writer) int` — calls the boot invariant, prints per-check report with checkmark/✗ glyphs, returns shell-style exit code (0 healthy, 1 hard failure)
  - `AdminReapplyProtectedMigrations(db, driver, confirm bool, out io.Writer) error` — refuses without confirm; otherwise: drop existing triggers, re-execute the per-dialect migration SQL, run the invariant to prove success, return error with diagnostic if anything fails

  Capitalise the trigger-SQL constants in `internal/repository/protected_triggers_sql.go` (e.g., `SQLiteProtectedTriggers`) to make them importable from `cmd/actalog/`.

- [ ] **Step 4:** Wire dispatch in `cmd/actalog/main.go`. Before the existing HTTP-server bring-up, check `os.Args[1] == "admin"` and dispatch to subcommands. Open the DB and run migrations *without* starting the server. Honor `--verbose` and `--confirm` flag presence (a tiny `hasFlag(args, name) bool` helper is enough).

- [ ] **Step 5:** Run all CLI tests — expect PASS.

- [ ] **Step 6:** Manual smoke: build, run `./bin/actalog admin verify-protected-users --verbose` (3/3 ✓ exit 0), break trigger via direct SQL, run again (✗ exit 1), run reapply (recovers), verify (clean).

- [ ] **Step 7:** Commit:
```
feat(cli): 'actalog admin verify-protected-users' and 'reapply-protected-migrations'
```

### Task 9: Recovery test matrix (per-failure-mode)

**Files:** `test/integration/protected_users_recovery_test.go` (new)

Use the `runActalogCLI` and `bootActalogServer` helpers introduced at the top of this plan.

- [ ] **Step 1:** Write tests covering the matrix in spec §6.4. Each test name maps directly to a row in that matrix:
  - `TestRecovery_FreshInstall_NoUsersYet`
  - `TestRecovery_TriggerDroppedBetweenBoots`
  - `TestRecovery_ReapplyCLIRestoresTriggers`
  - `TestRecovery_VerifyCLIDetectsAllFailureModes` (table-driven across "missing trigger" and "protected row deleted with others present")
  - `TestRecovery_DegradedMode_BootsWithBadTrigger`
  - `TestRecovery_DegradedMode_NormalRoutesStillWork`
  - `TestRecovery_DegradedMode_LogLineIsAlertable` (run binary 90s, scan logs, assert ≥1 `protected_invariant_degraded` line per 60s window)
  - `TestRecovery_GoFrontendDriftCheck` — hand-edit `web/src/utils/protectedUsers.js`, re-run the generator, assert the file was overwritten and `git diff --exit-code` would have flagged the drift
  - `TestRecovery_BackupRestoreScenario` — take a backup with triggers in place, restore the backup into a fresh DB, assert triggers are present and the boot invariant passes

  Each test uses `runActalogCLI` for CLI invocations, direct SQL for damage simulation, and `bootActalogServer` for full-binary boot scenarios.

- [ ] **Step 2:** Run on each dialect — expect PASS.

- [ ] **Step 3:** Commit:
```
test(integration): per-failure-mode recovery matrix
```

---

## Phase C — Application layer (L1+L2)

### Task 10: ProtectedUserGuard middleware (L1)

**Files:** `pkg/middleware/protected_user.go` (new), `pkg/middleware/protected_user_test.go` (new)

- [ ] **Step 1: Failing tests.** Eight cases per spec §6.2:
  - `GETPassesThroughAlways` — GET on protected user reaches the handler
  - `HEADPassesThroughAlways` — HEAD too
  - `PATCHOnProtectedReturns403` — structured JSON body with `error: "protected_user"`
  - `PATCHOnProtectedWritesAuditEvent` — recording stub captures one `protected_user_attack_http`
  - `PATCHOnNonProtectedPassesThrough` — alice@example.com reaches the handler
  - `DELETEOnProtectedReturns403` — covers all write methods
  - `MalformedIDParamReturns404` — non-integer `{id}` is 404, not 403, not 500
  - `UnknownUserReturns404` — repository ErrNoRows is 404

  Use a stub `UserRepository` and a recording stub `AuditLogger`. Mount the middleware on a chi `Route("/users/{id}", ...)` block with simple inline handlers that return 200 with method-named body for assertion.

- [ ] **Step 2:** `go test -count=1 ./pkg/middleware/` — expect FAIL.

- [ ] **Step 3: Implementation.** `pkg/middleware/protected_user.go`:
  - Define a minimal `AuditLogger` interface matching the existing `AuditLogService.LogEvent` signature (`(eventType string, userID, targetUserID *int64, ip, userAgent, details *string)`)
  - `ProtectedUserGuard(userRepo domain.UserRepository, audit AuditLogger) func(http.Handler) http.Handler`
  - Short-circuit on GET/HEAD
  - Parse `chi.URLParam(r, "id")` as int64; on parse error → 404
  - `userRepo.GetByID(id)`; on `sql.ErrNoRows` → 404
  - `security.IsProtectedEmail(user.Email)`: if true, write audit event with details containing `path`, `method`, `user_agent`, `referer`; return 403 with structured body via `json.Encoder`

- [ ] **Step 4:** Run middleware tests — expect PASS.

- [ ] **Step 5:** Verify 100% branch coverage on `protected_user.go`.

- [ ] **Step 6:** Commit:
```
feat(middleware): add L1 ProtectedUserGuard
```

### Task 11: AdminUserService — `ensureNotProtected`, `UpdateProfile`, `ForcePasswordReset`

**Files:** `internal/service/admin_user_service.go` (new), `internal/service/admin_user_service_test.go` (new)

Build stubs for the dependencies: `domain.UserRepository`, `domain.RefreshTokenRepository`, an `EmailSender` interface (`SendPasswordResetEmail` + `SendVerificationEmail`), and an `AuditLogger` interface (matches the recording stub from Task 10).

- [ ] **Step 1: Failing tests** — name and assertion per spec §6.1 (`internal/service/admin_user_service_test.go`):
  - `UpdateProfile_RejectsProtectedTarget` — protected target → returns `domain.ErrProtectedUser`
  - `UpdateProfile_OnlyUpdatesProvidedFields` — only `Name` set in `ProfileUpdateFields`; verify `Email` unchanged
  - `UpdateProfile_EmailChangeResetsVerification` — change email; verify `EmailVerified=false` and `SendVerificationEmail` called
  - `UpdateProfile_StaleUpdatedAtReturnsErrConflict` — pass an old `updated_at` and verify `ErrConflict`
  - `UpdateProfile_DBTriggerErrorIsTaggedAsAttackDB` — make the repository return an error containing the trigger contract substring; verify the audit event is `EventProtectedUserAttackDB` and the error returned is `ErrProtectedUser`
  - `ForcePasswordReset_RejectsProtectedTarget`
  - `ForcePasswordReset_SendsEmailAndRevokesTokens` — verify the stub email service was called with `SendPasswordResetEmail` AND the refresh-token stub's `RevokeAllForUser` was called AND audit event `EventPasswordResetForcedByAdmin` written

  Audit-coverage tests (separate `_audit_test.go` is fine, or integrate): every successful write emits an audit event with both `user_id` (actor) and `target_user_id` populated. Email change carries `old`/`new` in the `details` JSON.

- [ ] **Step 2:** Run — expect FAIL.

- [ ] **Step 3: Implementation.** `internal/service/admin_user_service.go`:
  - `ProfileUpdateFields` struct with pointer fields (`Name *string`, `Email *string`, `Birthday *time.Time`, `EmailVerified *bool`) — pointer means "unset", absent ≠ empty
  - `AdminUserService` constructor takes the four dependencies
  - `ensureNotProtected(targetID int64) error` — load user, check `IsProtectedEmail`, write `EventProtectedUserAttackService` audit, return `ErrProtectedUser`
  - `UpdateProfile(actorID, targetID int64, fields ProfileUpdateFields, ifMatchUpdatedAt time.Time) (*domain.User, error)`:
    - L2 check first
    - Reload user; compare `user.UpdatedAt` against `ifMatchUpdatedAt` (use `.Equal()`); mismatch → `ErrConflict`
    - Apply provided fields with validation per spec §4.3 (name 1-100, email via `mail.ParseAddress`, birthday past-or-null)
    - Email change branch: reset `EmailVerified=false`, generate verification token, call `SendVerificationEmail`, audit `email_changed` with old/new in details
    - `userRepo.Update(user)`: if the error message contains the L3 contract substring, write `EventProtectedUserAttackDB` audit and return `ErrProtectedUser` (this is L4)
    - On success: audit `profile_updated`; return updated user
  - `ForcePasswordReset(actorID, targetID int64) error`:
    - L2 check
    - Generate token, set on user (use existing repository helper or extend interface), call `SendPasswordResetEmail`, call `refreshTokenRepo.RevokeAllForUser(user.ID)`, audit `EventPasswordResetForcedByAdmin`

  If `UserRepository` lacks `SetVerificationToken` / `SetResetToken`, extend the interface and the concrete implementation (use existing patterns).

- [ ] **Step 4:** Run service tests — expect PASS.

- [ ] **Step 5:** Commit:
```
feat(service): add AdminUserService with L2 guard, UpdateProfile, ForcePasswordReset
```

### Task 12: Admin handler — PATCH and force-password-reset endpoints

**Files:** `internal/handler/admin_user_handler.go` (modify), `internal/handler/admin_user_handler_test.go` (modify)

- [ ] **Step 1: Failing tests.** Add to `admin_user_handler_test.go`:
  - `UpdateProfile_HappyPath` — PATCH with `{"name":"New Name","updated_at":"<current>"}` → 200 with body containing the new name
  - `UpdateProfile_StaleUpdatedAtReturns409`
  - `UpdateProfile_MissingUpdatedAtReturns400`
  - `ForcePasswordReset_Returns204`

  Use the existing `chiURLParam` helper (or add one to test helpers) to inject the `{id}` URL param into the request context.

- [ ] **Step 2:** Run — expect FAIL.

- [ ] **Step 3: Implementation.** Append to `admin_user_handler.go`:
  - Define `updateProfileRequest` struct with the same pointer fields as `ProfileUpdateFields` plus `UpdatedAt time.Time` (required)
  - `UpdateProfile(w, r)` handler: extract actorID via `middleware.GetUserID`, parse `{id}` via `chi.URLParam`, decode body, validate `UpdatedAt` is non-zero, call `adminUserService.UpdateProfile(...)`, route errors through `WriteError` (which maps `ErrProtectedUser`→403 and `ErrConflict`→409 from Task 2), JSON-encode the returned user on success
  - `ForcePasswordReset(w, r)` handler: same prelude; call `adminUserService.ForcePasswordReset(actorID, targetID)`; on success `w.WriteHeader(204)`

  Add `adminUserService *service.AdminUserService` field to `AdminUserHandler` struct; update `NewAdminUserHandler` constructor; wire in `cmd/actalog/main.go`.

- [ ] **Step 4:** Run handler tests — expect PASS.

- [ ] **Step 5:** Commit:
```
feat(handler): PATCH /admin/users/{id} and POST /admin/users/{id}/force-password-reset
```

### Task 13: Wire L1 middleware on the admin/users/{id} subrouter

**Files:** `cmd/actalog/main.go` (modify)

- [ ] **Step 1:** Locate the existing flat admin user routes (around `r.Get("/users", adminUserHandler.ListUsers)` and the `/users/{id}/*` group around line 905 of main.go).

- [ ] **Step 2:** Restructure as a chi sub-router. The list endpoint stays at `/users`. Everything `/users/{id}/*` moves into a `Route("/{id}", ...)` block that mounts `middleware.ProtectedUserGuard(userRepo, auditLogService)` AND `degradedAdminWriteGuard(&degraded)` (from Task 6) as middleware. Existing endpoints inside the block: `GetUserDetails` (GET — passes through L1), `UpdateProfile` (PATCH, new), `DeleteUser` (DELETE), `DisableUser`/`EnableUser`/`UnlockUser`/`ChangeUserRole`/`ToggleEmailVerification`, plus the new `ForcePasswordReset` (POST).

- [ ] **Step 3:** Manual smoke test: build, log in as admin, `curl -X PATCH .../api/admin/users/<protected-id>` → expect 403 with `error: "protected_user"`.

- [ ] **Step 4:** Commit:
```
feat(routing): mount L1 ProtectedUserGuard on /admin/users/{id} subrouter
```

### Task 14: Adversarial bypass tests

**Files:** `test/integration/protected_users_layered_defense_test.go` (new)

Per spec §6.5. These tests deliberately skip outer layers to prove the next layer catches independently.

- [ ] **Step 1: Tests.**
  - `TestDefense_BypassL1_ServiceCatchesIt` — wire only handler+service+repo (no middleware); call `AdminUserSvc.UpdateProfile` directly with a protected target → returns `domain.ErrProtectedUser`; audit log contains `EventProtectedUserAttackService`
  - `TestDefense_BypassL1AndL2_DBTriggerCatchesIt` — call `userRepo.Update(...)` directly on a protected user; expect DB error containing the contract substring
  - `TestDefense_AllLayersFireCorrectAuditEvents` — table-driven: HTTP attempt → exactly one `protected_user_attack_http`; service-direct → exactly one `protected_user_attack_service`; repo-direct → at most one `protected_user_attack_db` (the service-layer wrapping turns the DB error into the `_db` event when the call goes via the service; if calling repo directly, no audit fires automatically — that test asserts the DB error itself contains the contract substring instead)
  - `TestDefense_LayerOrderingPropagation` — normal request through L1+L2; assert exactly one audit event written (not three duplicates) and tagged at the rejecting layer

- [ ] **Step 2:** Run on each dialect — expect PASS.

- [ ] **Step 3:** Commit:
```
test(integration): adversarial bypass tests prove each layer is independently sufficient
```

---

## Phase D — Frontend foundation

### Task 15: protectedUsers.js consumer tests

**Files:** `web/src/utils/protectedUsers.test.js` (new)

- [ ] **Step 1: Tests** for the auto-generated module from Task 3:
  - exports a non-empty `Set` named `PROTECTED_EMAILS`
  - `isProtectedEmail('br8kwall@gmail.com')` is true
  - case-insensitive (uppercase, mixed)
  - trims whitespace
  - does NOT match plus-addressed variants
  - returns false for non-protected, empty, null, undefined

- [ ] **Step 2:** `cd web && npm run test:run -- src/utils/protectedUsers.test.js` — expect PASS (file already generated in Task 3).

- [ ] **Step 3:** Commit:
```
test(frontend): Vitest coverage for auto-generated protectedUsers module
```

### Task 16: `useUserDraft` composable

**Files:** `web/src/components/admin/user-edit/composables/useUserDraft.js` (new), `useUserDraft.test.js` (new)

- [ ] **Step 1: Failing tests** per spec §6.6:
  - `starts clean after fetchOriginal resolves`
  - `detects modified fields when working differs from original`
  - `save() sends only modified fields plus updated_at`
  - `save() refreshes original and clears dirty on success`
  - `save() surfaces 409 as conflict` (sets `conflict=true`)
  - `discard() resets working to original`

  Mock `fetchOriginal` and `saveFields` with `vi.fn()`. Use `vi.fn()`'s `toHaveBeenCalledWith(expect.objectContaining({...}))` to assert payload shape.

- [ ] **Step 2:** Run — expect FAIL.

- [ ] **Step 3: Implementation.** `useUserDraft.js`:
  - Inputs: `{ fetchOriginal, saveFields, fields }`
  - Reactive state: `original` (ref), `working` (reactive), `saving` (ref), `error` (ref), `conflict` (ref), `fieldErrors` (reactive)
  - Computed: `modified` (Set of fields where `working[f] !== original.value[f]`), `isDirty` (modified.size > 0)
  - Methods:
    - `load()` — calls `fetchOriginal()`, populates `original` and `working` from `res.data`
    - `save()` — early-return if not dirty; build payload as `{ updated_at: original.value.updated_at, ...modified-fields-only }`; call `saveFields(payload)`; on success refresh `original` and re-sync `working`; on 409 set `conflict=true`; rethrow
    - `discard()` — copy `original` back into `working`
  - Return all reactive refs and methods

- [ ] **Step 4:** Run tests — expect PASS.

- [ ] **Step 5:** Commit:
```
feat(frontend): useUserDraft composable
```

### Task 17: `useProtectedUserStatus` composable

**Files:** `web/src/components/admin/user-edit/composables/useProtectedUserStatus.js` (new)

Trivial wrapper. ~6 lines.

- [ ] **Step 1: Implementation.** Imports `isProtectedEmail` from `@/utils/protectedUsers`. Exports `useProtectedUserStatus(userRef)` returning a `computed(() => userRef.value ? isProtectedEmail(userRef.value.email) : false)`.

- [ ] **Step 2:** No dedicated test — Task 15 already covers `isProtectedEmail`; component-level tests in Tasks 20-22 exercise the composable in context.

- [ ] **Step 3:** Commit:
```
feat(frontend): useProtectedUserStatus composable
```

### Task 18: axios interceptors for 403 protected_user and 503 degraded

**Files:** `web/src/utils/axios.js` (modify), `web/src/utils/axios.test.js` (modify)

- [ ] **Step 1: Failing tests.** Add to the existing `axios.test.js`:
  - `catches 403 protected_user and dispatches actalog-snackbar event` — mock a 403 response with body `{ error: "protected_user", message: "...", documentation_url: "..." }`; assert a `CustomEvent('actalog-snackbar', ...)` was fired with `type: 'warning'` and the message text
  - `catches 503 protected_invariant_degraded and dispatches admin-degraded snackbar` — same shape with body `{ error: "protected_invariant_degraded" }`; assert `type: 'error'`

  Use the harness already in `axios.test.js` (likely `axios-mock-adapter` or fetch mocks).

- [ ] **Step 2:** Run — expect FAIL.

- [ ] **Step 3: Implementation.** In the existing response-error interceptor in `web/src/utils/axios.js`, add two new branches BEFORE the existing 402 branch:
  - `if (status === 403 && code === 'protected_user') { dispatch event with { type: 'warning', text: data.message, helpUrl: data.documentation_url } }`
  - `if (status === 503 && code === 'protected_invariant_degraded') { dispatch event with { type: 'error', text: 'Admin user actions are temporarily unavailable. Operator: see logs.' } }`

  Use the project's existing snackbar dispatch mechanism (a Pinia store, an event bus, or `window.dispatchEvent(new CustomEvent('actalog-snackbar', { detail }))` — pick the one already in use).

- [ ] **Step 4:** Run tests — expect PASS.

- [ ] **Step 5:** Commit:
```
feat(frontend): axios interceptor for protected_user 403 and degraded 503
```

---

## Phase E — Frontend UI

### Task 19: `ProtectedUserBanner` and `TabFooterActions` shared components

**Files:** `web/src/components/admin/user-edit/ProtectedUserBanner.vue` (new), `TabFooterActions.vue` (new)

These are presentational components reused by every tab — building them once in v1.3.0 pays off in v1.3.1+.

- [ ] **Step 1: ProtectedUserBanner.vue.** Vuetify `v-alert` with `type="info"`, `prominent`, `variant="tonal"`. Icon `mdi-shield-account`. Bold heading "This is a system-reserved account." Body explains all writes are disabled by design. Footer button linking to `/docs/protected-users` (target=_blank).

- [ ] **Step 2: TabFooterActions.vue.** `<script setup>` with `defineProps({ draft: { type: Object, required: true } })`. Template shows `· Unsaved changes` text (small caption) when `draft.isDirty.value`, plus `Discard` and `Save changes` buttons disabled when not dirty or while saving. Calls `draft.discard()` and `draft.save()` directly. (For tabs that need pre-save confirmation — e.g. ProfileTab's email-change flow — they can wrap the Save call by hiding `TabFooterActions` and providing their own action row.)

- [ ] **Step 3:** No dedicated tests for these (their logic is trivial; covered by ProfileTab tests in Task 20).

- [ ] **Step 4:** Commit:
```
feat(frontend): shared ProtectedUserBanner + TabFooterActions components
```

### Task 20: `ProfileTab.vue`

**Files:** `web/src/components/admin/user-edit/ProfileTab.vue` (new), `ProfileTab.test.js` (new)

- [ ] **Step 1: Failing tests** per spec §6.6:
  - `renders profile fields when target is not protected` — mount with mocked GET returning a non-protected user; assert form fields visible with the user's data; ProtectedUserBanner is NOT rendered
  - `shows ProtectedUserBanner for protected user` — mount with `email: 'br8kwall@gmail.com'`; assert banner is present and Save buttons are absent
  - `save sends only modified fields with updated_at` — mount, change name field, click Save; assert PATCH was called with payload containing the new name and the `updated_at` from the original load

  Mock `axios` via `vi.mock('@/utils/axios')`. Build a Vuetify instance for mounting.

- [ ] **Step 2:** Run — expect FAIL.

- [ ] **Step 3: Implementation.** `ProfileTab.vue` `<script setup>`:
  - Props: `userId: Number`
  - Use `useUserDraft({ fetchOriginal: () => axios.get(`/api/admin/users/${userId}`), saveFields: (f) => axios.patch(...), fields: ['name','email','birthday','email_verified'] })`
  - Use `useProtectedUserStatus(draft.original)` for the protected-banner branch
  - Call `draft.load()` on setup
  - Local state: `sendingReset`, `showEmailChangeConfirm`
  - Method `forcePasswordReset()` — POST to `/api/admin/users/${userId}/force-password-reset`; on success dispatch a success snackbar
  - Method `attemptSave()` — if `draft.modified.value.has('email')`, set `showEmailChangeConfirm=true` and return; otherwise call `draft.save()`
  - Method `confirmEmailChange()` — close dialog, call `draft.save()`

  Template:
  - `<ProtectedUserBanner v-if="isProtected" />`
  - `<v-form v-else>` with: `v-text-field` for name, email, birthday (type="date"); `v-switch` for email_verified
  - "Account actions" section with "Force password reset" button (variant=tonal, color=warning, prepend-icon=mdi-key-alert)
  - `<TabFooterActions :draft="draft" />` — but ProfileTab needs a custom save handler for the email confirmation, so either: (a) emit `click:save` from `TabFooterActions` and intercept here, OR (b) inline a custom action row in ProfileTab and skip `TabFooterActions`. Pick (a) for reusability.
  - `v-dialog` for the email-change confirmation

- [ ] **Step 4:** Run tests — expect PASS.

- [ ] **Step 5:** Commit:
```
feat(frontend): ProfileTab with draft+save, force-reset action, email-change confirmation
```

### Task 21: `AdminUserEditView.vue` shell + Vue Router registration

**Files:** `web/src/views/AdminUserEditView.vue` (new), `web/src/views/AdminUserEditView.test.js` (new), `web/src/router/index.js` (modify)

- [ ] **Step 1: Implementation of the shell.** `<script setup>`:
  - `useRoute()`, `useRouter()` for the URL
  - `userId = computed(() => Number(route.params.id))`
  - `activeTab = ref(route.query.tab || 'profile')`; `watch(activeTab, ...)` syncs back to the URL via `router.replace({ query: { ...route.query, tab: t } })` — enables deep links

  Template:
  - `AdminHeader` with breadcrumbs `[{title:'Users',to:'/admin/users'},{title:'Edit',to:route.fullPath}]`
  - `v-card` containing `v-tabs v-model="activeTab"` with one tab per future domain. Profile is enabled; Affiliations/Subscriptions/Credits/Preferences/Activity are `disabled` and show a `v-chip size="x-small"` with the version label (e.g. `v1.3.1`) so users see what's coming
  - `v-window v-model="activeTab"` with one `v-window-item value="profile"` containing `<ProfileTab :user-id="userId" />`. The other tabs are stubs that render a "Coming in v1.3.x" alert.

- [ ] **Step 2: Vue Router registration.** In `web/src/router/index.js`, add inside the admin routes block:
  - path: `/admin/users/:id/edit`, name: `admin-user-edit`, component: lazy `() => import('@/views/AdminUserEditView.vue')`, meta: `{ requiresAuth: true, requiresAdmin: true }`

- [ ] **Step 3: Tests.** `AdminUserEditView.test.js` mounts the view with a mock router and stubbed `ProfileTab`; asserts: tabs render with Profile selected by default; switching tabs updates the URL; deep-linking with `?tab=affiliations` shows the disabled tab as active.

- [ ] **Step 4:** Commit:
```
feat(frontend): AdminUserEditView shell with tabs (Profile active in v1.3.0)
```

### Task 22: AdminUsersView Edit button + protected-user disable

**Files:** `web/src/views/AdminUsersView.vue` (modify), `web/src/views/AdminUsersView.test.js` (modify)

- [ ] **Step 1: Failing tests.** Add to `AdminUsersView.test.js`:
  - `renders Edit pencil-icon button per row`
  - `Edit button is disabled for protected user rows` (assert `:disabled` attribute)
  - `clicking Edit navigates to /admin/users/:id/edit` (use a mock router and assert `push` was called)

- [ ] **Step 2:** Run — expect FAIL.

- [ ] **Step 3: Implementation.** In `AdminUsersView.vue`:
  - In `<script setup>`: import `isProtectedEmail` from `@/utils/protectedUsers`; define `function isProtected(user) { return isProtectedEmail(user.email) }`
  - In the actions column template: add a `v-btn icon size="small" variant="text" title="Edit user" :disabled="isProtected(item)" @click="$router.push('/admin/users/'+item.id+'/edit')"` with `<v-icon color="warning">mdi-pencil</v-icon>`
  - Optional: render a small shield icon next to the email text for protected rows so they're visually distinct

- [ ] **Step 4:** Run tests — expect PASS.

- [ ] **Step 5:** Commit:
```
feat(frontend): Edit button on AdminUsersView (disabled for protected users)
```

---

## Phase F — Recovery + docs

### Task 23: CODEOWNERS + generator drift CI check

**Files:** `.github/CODEOWNERS` (new or modify), `.github/workflows/ci.yml` (modify)

- [ ] **Step 1:** Add CODEOWNERS entries (replace `@johnzastrow` with whatever team the project uses):
```
pkg/security/                          @johnzastrow
internal/repository/protected_*        @johnzastrow
docs/security/                         @johnzastrow
scripts/recover/                       @johnzastrow
.github/workflows/ci-failure-notify.yml @johnzastrow
```

- [ ] **Step 2:** Add generator drift check as a step in the existing `Web: build` job (or its own short job):
```yaml
      - name: Generator drift check (protectedUsers.js)
        run: |
          go run ./cmd/gen-protected-emails/
          if ! git diff --exit-code -- web/src/utils/protectedUsers.js; then
            echo '::error::web/src/utils/protectedUsers.js is out of sync with pkg/security/protected_users.go.'
            echo '::error::Run `make gen-protected-emails` and commit the result.'
            exit 1
          fi
```

- [ ] **Step 3:** Commit:
```
ci: CODEOWNERS for security paths + generator drift check
```

### Task 24: `docs/security/PROTECTED_USERS.md` master runbook

**Files:** `docs/security/PROTECTED_USERS.md` (new)

Comprehensive ~600-800 line runbook covering all 10 sections from spec §7.1.

- [ ] **Step 1:** Write the file. Section structure (each must be present, sized to its content):
  1. Purpose & scope — what this is, what it isn't
  2. Threat model — what protects against (admin error; compromised admin; hostile insider with shell; SQL injection; rogue migration); what does NOT protect against (root on the DB server, physical disk access, supply-chain compromise of Go modules)
  3. Architecture — diagram of L1/L2/L3/L4 and how a request flows through them
  4. The protected list — current contents; one paragraph per entry explaining who/why
  5. How to add a protected user — PR checklist with exact file paths, the CODEOWNERS gate, the boot check that validates the change
  6. How to remove a protected user — same procedure
  7. How to verify (`actalog admin verify-protected-users`) — example output for healthy AND failure cases; explanation of each check
  8. Recovery playbook — disaster matrix from spec §2.6 expanded with paste-ready shell commands per scenario
  9. Audit-log forensics — query patterns (e.g. SQL example for "find all `protected_user_attack_*` events from non-browser User-Agents in the last 7 days"); event-type semantics; UI vs script classification at query time
  10. Degraded mode — when to use `ACTALOG_SKIP_PROTECTED_INVARIANT=true`; what works/doesn't; monitoring; exit criteria

  Code blocks tagged with language fences (` ```bash `, ` ```sql `) for the doc-lint step in Task 28.

- [ ] **Step 2:** Commit:
```
docs(security): master runbook for the protected-user system
```

### Task 25: `docs/security/PROTECTED_USERS_RECOVERY.md` incident playbook

**Files:** `docs/security/PROTECTED_USERS_RECOVERY.md` (new)

Pure how-to. ~200-300 lines. The 3-AM-incident-response context.

- [ ] **Step 1:** Write the file:
  - Symptom-to-command table: each row is a boot-time error message → exact command to run
  - Annotated example boot-failure output (the diagnostic block from Task 6) with arrows pointing at the actionable parts
  - Decision flowchart in ASCII: try verify CLI first → if that fails, try reapply → if that fails, try the raw-SQL recovery script → if that fails, restore from backup
  - Per-DB-driver paste-ready commands

- [ ] **Step 2:** Commit:
```
docs(security): focused incident-response playbook for boot-time recovery
```

### Task 26: `docs/security/THREAT_MODEL.md`

**Files:** `docs/security/THREAT_MODEL.md` (new)

App-wide threat model. Each section links out to the implementation it documents.

- [ ] **Step 1:** Write the file:
  - Overview + scope
  - Per-section: auth/sessions, rate limiting, input validation, SQL injection prevention, file uploads (link to v1.2.4 magic-byte work in `internal/handler/user_handler.go`), CORS (link to `pkg/middleware/cors.go`), CSP (link to `pkg/middleware/security_headers.go` from v1.2.4), password policy (link to `internal/service/user_service.go`), protected users (link to `docs/security/PROTECTED_USERS.md`)
  - Each section states the threats it addresses, the mitigations in place, and the residual risks

- [ ] **Step 2:** Commit:
```
docs(security): app-wide threat model
```

### Task 27: Update existing docs

**Files:** `docs/USER_PERMISSIONS.md`, `docs/ARCHITECTURE.md`, `docs/DATABASE_SCHEMA.md`, `docs/TESTING.md`, `docs/CHANGELOG.md`, `docs/TODO.md`, `CLAUDE.md`, `README.md` (all modified)

Per spec §7.2:

- [ ] **USER_PERMISSIONS.md** — add rows for `PATCH /admin/users/{id}` and `POST /admin/users/{id}/force-password-reset`; add a top-of-section note that all `/admin/users/{id}/*` writes are blocked for protected users
- [ ] **ARCHITECTURE.md** — new "Defense-in-Depth" subsection documenting the L1-L4 pattern as a reference architecture for future security work
- [ ] **DATABASE_SCHEMA.md** — new "Triggers" subsection listing the protected-user triggers; bump schema-version stamp to `0.35.0`
- [ ] **TESTING.md** — new "Security Tests" subsection pointing at `test/integration/protected_users_*_test.go` and explaining the layered-defense test pattern as a model for future security tests
- [ ] **CHANGELOG.md** — full v1.3.0 entry: Security (defense-in-depth, all four layers, recovery tooling), Added (admin user-edit screen with Profile tab + force password reset), Documentation (the four new security docs), Maintenance (release engineering bits)
- [ ] **TODO.md** — mark "Comprehensive User Edit Screen" as in-progress; add `[HIGH] User Edit Screen — Affiliations tab (v1.3.1)` and `[HIGH] User Edit Screen — Subscriptions / Credits / Preferences / Activity tabs (v1.3.2)` entries
- [ ] **CLAUDE.md** — replace inline "Protected Users (DO NOT MODIFY)" paragraph with a one-line pointer to `docs/security/PROTECTED_USERS.md` (the runbook is now load-bearing enough to deserve its own doc)
- [ ] **README.md** — version stamp bump (will be 1.3.0 in Task 29); one-line mention in the security section pointing at the new runbook

- [ ] Commit:
```
docs: catch up existing docs with v1.3.0 protected-user system
```

### Task 28: Doc code-block CI lint

**Files:** `.github/workflows/ci.yml` (modify)

- [ ] **Step 1:** Add a new step:
```yaml
      - name: Doc code-block syntax check
        run: |
          set -e
          for f in docs/security/PROTECTED_USERS.md docs/security/PROTECTED_USERS_RECOVERY.md; do
            awk '/^```bash$/,/^```$/' "$f" | sed '/^```/d' > /tmp/check.sh
            bash -n /tmp/check.sh || { echo "syntax error in bash blocks of $f"; exit 1; }
          done
          for f in docs/security/PROTECTED_USERS.md; do
            awk '/^```sql$/,/^```$/' "$f" | sed '/^```/d' > /tmp/check.sql
            sqlite3 :memory: ".read /tmp/check.sql" || true
          done
```

- [ ] **Step 2:** Commit:
```
ci: lint bash and sql code blocks in security docs
```

---

## Phase G — Release

### Task 29: Version bump and release notes

**Files:** `pkg/version/version.go`, `web/package.json`, `README.md`, `CLAUDE.md`, `docs/CHANGELOG.md`, `docs/TODO.md`

- [ ] **Step 1:** Bump versions:
  - `pkg/version/version.go`: `Minor = 2 → 3`, `Patch = 4 → 0`, `Build = 49 → 50`
  - `web/package.json`: `"version": "1.3.0"`
  - `README.md` and `CLAUDE.md`: `1.2.4` → `1.3.0` (replace_all)

- [ ] **Step 2:** Finalize the CHANGELOG v1.3.0 entry from Task 27 — review once with fresh eyes, add the release date.

- [ ] **Step 3:** Update TODO entries (already drafted in Task 27): mark Comprehensive User Edit Screen as **partially shipped (v1.3.0: Profile tab + framework)**; ensure the v1.3.1 and v1.3.2 follow-on entries are present and tagged `[HIGH]`.

- [ ] **Step 4:** Commit:
```
release: v1.3.0 — admin user-edit screen + protected-user defense-in-depth

First user-edit tab (Profile) + the full L1-L4 defense framework + boot
invariant + recovery tooling. Affiliations and remaining tabs follow in
v1.3.1 and v1.3.2.
```

### Task 30: Pre-PR smoke test, push, create PR

- [ ] **Step 1: Full test suite.**
```bash
go test -count=1 ./...
DB_DRIVER=sqlite3 go test -count=1 ./test/integration/
DB_DRIVER=postgres go test -count=1 ./test/integration/
DB_DRIVER=mysql    go test -count=1 ./test/integration/
cd web && npm run test:run
```
All must pass.

- [ ] **Step 2: Local Docker smoke test** (per the deploy-before-merge memory):
```bash
./docker/scripts/build.sh dev
docker run -d --name actalog-v1.3.0-test -p 8080:8080 -v $(pwd)/data:/app/data \
  -e DB_DRIVER=sqlite3 -e DB_NAME=/app/data/actalog.db \
  ghcr.io/johnzastrow/actalog:dev
```

In the browser, exercise the full flow:
- Register first user → becomes admin automatically
- Visit `/admin/users`; create a non-admin user
- Click the new pencil-icon Edit button → User Edit page loads with Profile tab active
- Edit name → "Unsaved changes" appears → click Save → reload page, name persists
- Edit email → confirmation dialog appears; cancel works
- Click "Force password reset" → success snackbar; check email logs admin view to confirm
- Try to navigate by URL to `/admin/users/<owner-id>/edit` → ProtectedUserBanner shown, no Save buttons
- `curl -X PATCH .../api/admin/users/<owner-id>` (with admin token) → 403 with `error: "protected_user"` body
- Inside the running container: `sqlite3 .../actalog.db "DROP TRIGGER protected_users_no_update"` → restart container → expect refusal-to-start (boot logs name the failing check)
- Re-start with `ACTALOG_SKIP_PROTECTED_INVARIANT=true` → boots degraded; `/health` returns 503
- Run `actalog admin verify-protected-users --verbose` from the container → see the failure
- Run `actalog admin reapply-protected-migrations --confirm` → recovery; then verify clean

- [ ] **Step 3:** Push branch and open PR:
```bash
git push -u origin <branch-name>
gh pr create --base main --title "v1.3.0 — admin user-edit screen + protected-user defense-in-depth"
```

PR body should include the v1.3.0 acceptance checklist from spec §8.6 with each box checked.

---

## Self-review

**1. Spec coverage:**
- Spec §1 Goal — Tasks 19–22 (frontend) + 11–13 (backend)
- Spec §2 Architecture — Tasks 21 (route+shell), 19–22 (tabs)
- Spec §3 Defense — Tasks 1, 4–11, 14
- Spec §4 Backend API surface — Tasks 11, 12
- Spec §5 Frontend — Tasks 15–22
- Spec §6 Test apparatus — Tasks 4, 9, 14, 23, 28; coverage gates in 23
- Spec §7 Documentation — Tasks 24–27
- Spec §8 Migration & rollout — Tasks 4, 6, 29, 30
- Spec §9 Out of scope — explicit in v1.3.0 release notes (Task 29)
- Spec §10 Decision log — preserved verbatim in spec; not duplicated in plan

**2. Placeholder scan:** no `TBD` / `TODO` / `fill in later` content. The few "(adjust if a simpler approach exists for a given dialect)" notes refer to integration with project-specific helpers — these are decisions the executing engineer makes by reading 1-2 existing files, not punted requirements.

**3. Type / signature consistency:** `ProfileUpdateFields`, `useUserDraft`, `ProtectedUserGuard`, `IsProtectedEmail`, `VerifyProtectedUserInvariant`, `AdminUserService`, `AuditLogger`, `EmailSender` — all referenced consistently across tasks. Audit-event constants `EventProtectedUserAttackHTTP/Service/DB` and `EventPasswordResetForcedByAdmin` are defined in Task 2 and used in Tasks 10, 11, 14.

**4. Scope:** focused on v1.3.0 only. v1.3.1 (Affiliations) and v1.3.2 (remaining tabs) explicitly out of scope and tracked as TODO entries created in Task 29.

---

*Plan generated via `superpowers:writing-plans` skill on 2026-04-28 from spec `docs/superpowers/specs/2026-04-28-admin-user-edit-design.md`.*
