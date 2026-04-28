# Admin User Edit Screen — Design Spec

**Date:** 2026-04-28
**Target release:** v1.3.0 (Profile tab + framework), v1.3.1 (Affiliations), v1.3.2 (remaining tabs)
**Status:** Approved (brainstorming complete; ready for implementation planning)

---

## 1. Goal

Provide a single admin surface for viewing and editing every attribute associated with a user account. Currently `AdminUsersView.vue` only supports disable/enable/unlock/role-change — there is no path to edit profile fields or manage cross-domain associations (gym memberships, coach assignments, subscriptions, credits, settings).

The motivating example is gym affiliations, but the screen covers six distinct data domains across the codebase, each currently scoped to the authenticated user. Phase 2 wires admin-level write surfaces to all of them inside one tabbed view.

## 2. Architecture

### 2.1 Frontend

- **Route:** `/admin/users/:id/edit` with optional `?tab=<name>` query param for deep linking
- **Page component:** `web/src/views/AdminUserEditView.vue`
- **Tab pattern:** Vuetify `v-tabs` + `v-window` + `v-window-item value="..."`, matching `AdminPackagesView.vue:37-44`
- **Tab components** (one file each in `web/src/components/admin/user-edit/`):
  - `ProfileTab.vue` (v1.3.0)
  - `AffiliationsTab.vue` (v1.3.1)
  - `SubscriptionsTab.vue`, `CreditsDocumentsTab.vue`, `PreferencesTab.vue`, `ActivityAuditTab.vue` (v1.3.2)
- **Shared components:**
  - `ProtectedUserBanner.vue` — shown when target is protected; replaces editable surface entirely
  - `TabFooterActions.vue` — Save / Discard / "Unsaved changes" indicator; reused by every tab
- **Shared composables** (`web/src/components/admin/user-edit/composables/`):
  - `useUserDraft.js` — encapsulates draft state, dirty detection, conflict handling, save/discard
  - `useProtectedUserStatus.js` — single source of truth for the protected-user check on the frontend

### 2.2 Backend

- **Service:** new `internal/service/admin_user_service.go` — admin-scoped wrappers around existing user/org/subscription/credits/settings services. Each method takes `actor int64` and `target int64` so audit logs always have both
- **Handler:** extends existing `internal/handler/admin_user_handler.go`
- **Middleware:** new `pkg/middleware/protected_user.go` (L1)
- **Helper:** new `pkg/security/protected_users.go` (single source of truth for the protected list)
- **Migration:** new `0.35.0_add_protected_user_triggers` (L3)

### 2.3 Edit pattern: draft + save per tab

Editing a field modifies local state. A Save button at the tab footer commits all modified fields for that tab in one PATCH request. Discard resets the working copy to the server's original.

Rationale: cleaner audit log (one event per intentional save), tolerates typos, atomic per-tab unit of work, and matches the bounded-context-per-tab service split.

## 3. Defense-in-Depth: protected user system

Four independent layers guard `br8kwall@gmail.com` (and any future protected accounts). Each layer is independently sufficient — the others exist as defense in depth.

### 3.1 L1 — HTTP middleware

**File:** `pkg/middleware/protected_user.go`

Wraps the `/api/admin/users/{id}/*` subrouter. Short-circuits on `GET`/`HEAD` (reads pass through). For write methods, loads the target user by URL `{id}` param, checks `pkg/security.IsProtectedEmail(user.Email)`, returns `403` with a structured error body and writes an L4 audit event if protected.

```go
r.Route("/users/{id}", func(r chi.Router) {
    r.Use(middleware.ProtectedUserGuard(userRepo, auditLogService))
    r.Patch("/", adminUserHandler.UpdateProfile)
    r.Post("/force-password-reset", adminUserHandler.ForcePasswordReset)
    r.Post("/disable", adminUserHandler.DisableUser)
    // ...all existing user-by-id write endpoints inherit protection automatically
})
```

### 3.2 L2 — Service-layer guard

**Helper:** `internal/service/admin_user_service.go::ensureNotProtected(targetID int64) error`

Every admin write service method calls this as its first line. Catches: cron jobs, scheduled tasks, internal callers, future routes added outside the protected subrouter, bugs in middleware. Returns typed `ErrProtectedUser`, mapped to HTTP 403 by `internal/handler/errors.go`.

### 3.3 L3 — Database trigger

**Migration:** `0.35.0_add_protected_user_triggers` in `internal/repository/migrations.go`

Per-dialect SQL — `BEFORE UPDATE` and `BEFORE DELETE` triggers on `users` that raise an error when `OLD.email` matches the protected set:

- **SQLite:** `CREATE TRIGGER ... BEGIN SELECT RAISE(ABORT, 'protected user: writes blocked at db layer'); END;`
- **PostgreSQL:** PL/pgSQL function + `CREATE TRIGGER ... EXECUTE FUNCTION block_protected_users()` raising via `RAISE EXCEPTION`
- **MySQL/MariaDB:** `CREATE TRIGGER ... BEGIN IF OLD.email = '...' THEN SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '...'; END IF; END;`

Catches: raw SQL via shell, SQL injection that bypasses the Go layer, rogue migrations, anyone with database credentials but no app access.

The error message text is part of the contract — L4's database-level catch grep-matches it. A change to the trigger SQL must be matched by a change to the L4 pattern, enforced by `TestL3_TriggerErrorMessageMatchesContract`.

### 3.4 L4 — Audit + structured error log

New audit event types in `internal/domain/audit_log.go`:
- `EventProtectedUserAttackHTTP = "protected_user_attack_http"` — L1 blocked
- `EventProtectedUserAttackService = "protected_user_attack_service"` — L2 blocked
- `EventProtectedUserAttackDB = "protected_user_attack_db"` — L3 caught (recognized by error-message pattern in service-layer error wrapping)

Event tagged by *which layer rejected*, not by which layer was traversed. A request that passes through L1 and L2 and is rejected at L3 fires exactly one event tagged `_db`.

Each event also produces a structured `[ERROR]` log line with `actor_id`, `target_id`, `path`, `user_agent`, `referer`, `had_session_cookie` — alertable by any log aggregator.

### 3.5 Boot-time invariant

`cmd/actalog/main.go` calls a startup check before starting the HTTP server. Three sub-checks:

| # | Check | Severity |
|---|-------|----------|
| 1 | L3 triggers exist in DB catalog | HARD — refuse to boot |
| 2 | Simulated UPDATE inside `BEGIN; ROLLBACK` rejects | HARD — refuse to boot |
| 3 | Protected user row exists in `users` | SOFT (WARN) on fresh install; HARD if any users exist |

Check 3's split severity handles the bootstrap problem: a fresh DB has no users yet, and the owner registers via the normal flow. Hard-failing on a fresh install would brick the binary.

If any HARD check fails, the binary prints a diagnostic naming the specific failure and the exact recovery command, then exits non-zero.

### 3.6 Where the protected list lives

**Single source of truth:** `pkg/security/protected_users.go` — Go constant.

The migration's trigger SQL embeds the same list. The frontend gets the list via an auto-generated `web/src/utils/protectedUsers.js` produced by `cmd/gen-protected-emails/`. Three artifacts; one source.

CI checks for drift: any PR that modifies `pkg/security/protected_users.go` without also updating the migration AND regenerating the frontend file fails the build.

### 3.7 Operations: how to change the protected list

**No runtime path.** The only way to add or remove a protected user is:

1. Open a PR that touches `pkg/security/protected_users.go` AND adds a new migration `0.NN.0_update_protected_users.sql`. CI requires both.
2. Migration drops and recreates the L3 triggers with the updated email list.
3. Run the generator to update the frontend artifact.
4. CODEOWNERS gate on `pkg/security/**` requires a security-team reviewer.
5. Boot-time invariant catches any drift on the next start.

Removing follows the same procedure. Audit trail is the git commit message.

**Why no runtime path:** anything you can change at runtime is something an attacker can change. The friction of a release+migration is the security boundary.

### 3.8 Operations: explaining blocks to humans

Four UX layers, preventive to corrective:

**Layer A — Don't let them click.** `AdminUsersView.vue` shows a shield icon on protected rows; Edit/Disable/Role/Delete actions render disabled with a tooltip pointing at the runbook.

**Layer B — Read-only at the route.** Deep-linking to `/admin/users/{protectedId}/edit` loads the page in read-only mode with a banner explaining why. Save/Discard buttons are absent (not disabled — absent), form fields are `:readonly`.

**Layer C — Structured 403 body.** Server response shape:
```json
{
  "error": "protected_user",
  "message": "This account is system-reserved and cannot be modified.",
  "documentation_url": "/docs/protected-users",
  "support_contact": "<configured admin email>"
}
```

**Layer D — Frontend interceptor.** `web/src/utils/axios.js` catches `403 + error="protected_user"` and shows a Vuetify snackbar: "This is a system-protected account — no changes were saved. [ Why? ]". Click [Why?] opens a dialog with the full explanation and runbook link. No raw error text shown.

### 3.9 Operations: recovery from boot-time failure

Four escape hatches by severity:

| Tool | Use when | Risk |
|------|----------|------|
| `actalog admin verify-protected-users` | Diagnose without restarting | Read-only, always safe |
| `actalog admin reapply-protected-migrations --confirm` | Triggers were dropped/corrupted; want to restore from current source | Idempotent; writes one audit event |
| `ACTALOG_SKIP_PROTECTED_INVARIANT=true` | Binary won't start AND CLI subcommands aren't reachable (crashloop, no shell) | Boots in degraded mode: admin-write endpoints return 503; `/health=degraded`; ERROR log every 60s |
| `scripts/recover/restore-protected-triggers.sh` | Binary itself is unrunnable | Last resort; raw SQL + shell |

Disaster matrix in `docs/security/PROTECTED_USERS.md` maps each failure mode to a first/second/third-choice recovery. Recovery itself is tested per-DB in `test/integration/protected_users_recovery_test.go`.

### 3.10 Rollback property: security stays on

Rolling back the binary from v1.3.0 to v1.2.4 leaves the L3 triggers in place. v1.2.4 doesn't know about them but doesn't conflict — they keep rejecting writes against protected rows. **Rolling back the binary does not roll back the security.** This is a feature, not an accident: the strongest layer lives in the most stable place.

## 4. v1.3.0 backend API surface

### 4.1 New endpoints

```
PATCH  /api/admin/users/{id}                    Profile field updates
POST   /api/admin/users/{id}/force-password-reset   Trigger reset + revoke refresh tokens
```

Both inherit L1 protection from the shared subrouter middleware.

### 4.2 PATCH body

All editable fields optional; only present fields update. Atomic per-request — either all listed fields apply or none do.

```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "birthday": "1990-04-15",
  "email_verified": true,
  "updated_at": "2026-04-28T15:30:00Z"
}
```

**`updated_at` is a precondition, not an editable field.** Required on every PATCH. The server uses it to detect concurrent edits (Section 4.4) and never writes it from the request body — `updated_at` is set server-side on every successful UPDATE.

### 4.3 Field rules (server-enforced)

| Field | Validation |
|-------|------------|
| `name` | trimmed; length 1–100; empty rejected |
| `email` | RFC-5322-ish via `mail.ParseAddress`; on change forces `email_verified=false`, generates new verification token, sends verification email |
| `birthday` | ISO-8601 `YYYY-MM-DD` or RFC-3339; past or null only |
| `email_verified` | admin override; if sent alongside `email` change, the email-change rule wins |

### 4.4 Conflict handling

Server compares request's `updated_at` (sent in the body) against current row's `updated_at`. Mismatch → `409 Conflict`. Frontend composable surfaces a "User changed elsewhere — discard your edits and reload?" dialog.

### 4.5 Force password reset side effects

1. `auth_service.SendPasswordResetEmail()` — same flow as `/api/auth/forgot-password`; reuses existing email template
2. `refresh_token_repository.RevokeAllForUser(targetID)` — invalidates current sessions
3. Audit event `password_reset_forced_by_admin` with both `user_id` (admin) and `target_user_id`

Returns `204 No Content`. Why a separate endpoint instead of a PATCH field: PATCH is idempotent; force-reset has irreversible side effects (email sent, tokens revoked) that are wrong to repeat on a retry.

## 5. v1.3.0 frontend

### 5.1 `useUserDraft.js` composable

```js
const draft = useUserDraft({
  fetchOriginal: () => axios.get(`/api/admin/users/${userId}`),
  saveFields:    (fields) => axios.patch(`/api/admin/users/${userId}`, fields),
  fields: ['name', 'email', 'birthday', 'email_verified'],
})

// reactive state: original, working, modified (Set), isDirty, saving, error, fieldErrors
// methods: save(), discard()
```

Key properties:
- **Only modified fields are sent.** Unchanged fields are excluded from the PATCH body — keeps the audit log clean and avoids re-triggering email verification on save.
- Save disabled while clean (`!isDirty || saving`).
- `409 Conflict` triggers conflict dialog.
- Email-change branch: if `email` is in the modified set, `save()` shows a confirmation dialog before sending — UX guard, not security guard.

### 5.2 ProfileTab UI

```
┌────────────────────────────────────────────┐
│  Admin > Users > jane@example.com          │
├────────────────────────────────────────────┤
│ [Profile] [Affiliations] [Subs] [Credits]  │
├────────────────────────────────────────────┤
│  Name:  [ Jane Doe              ]          │
│  Email: [ jane@example.com      ]          │
│  Birthday: [ 1990-04-15  ]                 │
│  ☑ Email verified                          │
│                                            │
│  Account actions:                          │
│  [ Force password reset ]                  │
│                                            │
│  · Unsaved changes                         │
│  [ Discard ]   [ Save changes ]            │
└────────────────────────────────────────────┘
```

For protected users, the entire form is replaced by `ProtectedUserBanner.vue`.

### 5.3 Entry point in AdminUsersView

New pencil icon in the actions column, disabled (with tooltip) for protected users:
```html
<v-btn icon size="small" variant="text"
       title="Edit user"
       :disabled="isProtected(item)"
       @click="$router.push(`/admin/users/${item.id}/edit`)">
  <v-icon color="warning">mdi-pencil</v-icon>
</v-btn>
```

`isProtected()` reads from the auto-generated `web/src/utils/protectedUsers.js`.

## 6. Test apparatus

### 6.1 Unit tests

| File | Scope |
|------|-------|
| `pkg/security/protected_users_test.go` | `IsProtectedEmail()` case-insensitivity, exact match (no plus-addressing match), trim handling |
| `internal/service/admin_user_service_test.go` | Each write method returns `ErrProtectedUser` for protected target; email-change resets verification; conflict on stale `updated_at` |
| `internal/service/admin_user_service_audit_test.go` | Every write emits audit event with both `user_id` and `target_user_id`; email-change carries old/new in `details` JSON |

### 6.2 Middleware tests (L1)

`pkg/middleware/protected_user_test.go` — eight cases covering: GET/HEAD pass-through, all write methods returning structured 403, audit-event emission, non-protected pass-through, malformed `{id}` returning 404 (not 403, not 500), JSON-schema check on the response body.

### 6.3 Integration tests — per-DB matrix (L3)

`test/integration/protected_users_test.go` runs on sqlite/postgres/mysql:
- Triggers exist in DB catalog post-migration
- UPDATE on protected row rejected by DB
- DELETE on protected row rejected by DB
- UPDATE on non-protected row succeeds
- Trigger error message matches L4's pattern contract
- Trigger survives transaction rollback cleanly

### 6.4 Recovery tests — per-failure-mode matrix

`test/integration/protected_users_recovery_test.go`:

| Test | Setup | Action | Expectation |
|------|-------|--------|-------------|
| `TestRecovery_FreshInstall_NoUsersYet` | fresh DB | boot | boots with WARN; `/health=healthy` |
| `TestRecovery_TriggerDroppedBetweenBoots` | drop trigger | boot | refuses with diagnostic naming the specific failure |
| `TestRecovery_ReapplyCLIRestoresTriggers` | drop trigger | reapply CLI | triggers restored; subsequent UPDATE rejected |
| `TestRecovery_VerifyCLIDetectsAllFailureModes` | each failure mode | verify CLI | exits 1 with per-check report |
| `TestRecovery_DegradedMode_BootsWithBadTrigger` | bad trigger + env var | boot | degraded mode; `/health=degraded`; admin write 503 |
| `TestRecovery_DegradedMode_NormalRoutesStillWork` | same | non-admin GET | 200 |
| `TestRecovery_DegradedMode_LogLineIsAlertable` | same, run 90s | tail logs | ≥1 ERROR line per 60s window |
| `TestRecovery_GoFrontendDriftCheck` | hand-edit JS file | run generator | exits non-zero |
| `TestRecovery_BackupRestoreScenario` | backup with triggers; restore | boot | triggers preserved; boot succeeds |

### 6.5 Adversarial tests — prove each layer is independently sufficient

`test/integration/protected_users_layered_defense_test.go`:

| Test | Bypasses | Asserts |
|------|----------|---------|
| `TestDefense_BypassL1_ServiceCatchesIt` | middleware | service returns `ErrProtectedUser` |
| `TestDefense_BypassL1AndL2_DBTriggerCatchesIt` | middleware + service | DB returns trigger error |
| `TestDefense_AllLayersFireAuditEvents` | each bypass | correct event_type tagging |
| `TestDefense_LayerOrderingPropagation` | normal request | exactly one audit event, tagged at the rejecting layer (not all traversed layers) |

### 6.6 Frontend tests (Vitest)

| File | Scope |
|------|-------|
| `web/src/views/AdminUserEditView.test.js` | route load, tab switching, deep-link via `?tab=` |
| `web/src/components/admin/user-edit/ProfileTab.test.js` | draft+save state machine, validation errors, protected banner |
| `web/src/components/admin/user-edit/composables/useUserDraft.test.js` | only-modified-fields-sent assertion, conflict 409 handling, save/discard |
| `web/src/utils/axios.test.js` (extends existing) | `403 + error="protected_user"` snackbar; `503 + protected_invariant_degraded` handling |
| `web/src/utils/protectedUsers.test.js` | generated module behavior; never edited by hand |

### 6.7 Build-time / CI checks

- Generator drift: `make build` runs the generator, `git diff --exit-code` on the output
- Migration ↔ recovery script lockstep: CI diff between migration SQL and `scripts/recover/sql/<dialect>/`
- CODEOWNERS for `pkg/security/**` and `migrations/*protected_users*`
- Boot-invariant check in CI: integration tests boot a real binary, assert `/health=healthy` only after migrations apply

### 6.8 Coverage targets (CI-enforced gates)

- `pkg/security/` — 100% line coverage
- `pkg/middleware/protected_user.go` — 100% branch coverage
- `internal/service/admin_user_service.go` — 90%+ line coverage

CI fails the build if any file in `pkg/security/` drops below 100%.

## 7. Documentation deliverables

### 7.1 New documents

| File | Purpose |
|------|---------|
| `docs/security/PROTECTED_USERS.md` | Master runbook: purpose, threat model, architecture, list management, verification, recovery, audit forensics, degraded mode |
| `docs/security/PROTECTED_USERS_RECOVERY.md` | Focused 3-AM-incident-response playbook: error message → exact command, decision flowchart, paste-ready commands per DB |
| `docs/security/THREAT_MODEL.md` | Broader app-wide threat model linking out to existing implementations (CORS, CSP, password policy, magic-byte avatar validation) and the new protected-user surface |
| `scripts/recover/README.md` | When and how to use raw SQL recovery scripts; per-DB-driver invocation; safety warnings |
| `docs/superpowers/specs/2026-04-28-admin-user-edit-design.md` | This document |

### 7.2 Existing documents to update

| File | Update |
|------|--------|
| `docs/USER_PERMISSIONS.md` | New `PATCH /admin/users/{id}` and `POST /admin/users/{id}/force-password-reset` rows; note that all `/admin/users/{id}/*` writes are blocked for protected users |
| `docs/ARCHITECTURE.md` | New "Defense-in-Depth" subsection documenting L1-L4 as a reference architecture |
| `docs/DATABASE_SCHEMA.md` | New "Triggers" section listing protected-user triggers; note v1.3.0 migration ships triggers without table changes |
| `docs/TESTING.md` | New "Security Tests" subsection documenting the layered-defense test pattern |
| `docs/CHANGELOG.md` | v1.3.0 Security and Added sections |
| `docs/TODO.md` | Mark User Edit Screen in-progress; add v1.3.1 (Affiliations) and v1.3.2 entries |
| `CLAUDE.md` | Replace inline "Protected Users (DO NOT MODIFY)" paragraph with a pointer to `docs/security/PROTECTED_USERS.md` |
| `README.md` | Version stamp; one-line mention in security section |

### 7.3 Auto-generated documentation

- `web/src/utils/protectedUsers.js` — generated from `pkg/security/protected_users.go` by `cmd/gen-protected-emails/`
- `docs/security/PROTECTED_USERS_LIST.md` — table view of who's protected, generated by the same tool
- OpenAPI spec `/docs/swagger.json` — new endpoints get docstrings (`@Summary`, `@Param`, `@Failure 403 protected_user`)

### 7.4 Documentation tested by CI

Code blocks in `docs/security/PROTECTED_USERS.md` and `docs/security/PROTECTED_USERS_RECOVERY.md` are tagged (` ```bash `, ` ```sql `) and validated by a new CI step:
- `bash` blocks → `bash -n` syntax check
- `sql` blocks → parsed by per-dialect CLIs

Catches typos in the runbook before they reach a 3-AM operator.

## 8. Migration & rollout

### 8.1 Migration

Single new migration `0.35.0_add_protected_user_triggers`. Per-dialect SQL creates `protected_users_no_update` and `protected_users_no_delete` triggers. `down` drops both — provided for completeness but **forbidden by policy** for production rollback.

No data backfill, no downtime beyond a normal restart. Customers upgrading from v1.2.x apply this single migration on next start.

### 8.2 Boot sequence

```
1. Load config + .env
2. Open DB connection
3. Run migrations          ← 0.35.0 lands here
4. Verify protected-user invariant   ← NEW
5. Seed standard movements + WODs
6. Wire handlers and middleware
7. Start HTTP server
```

Step 4 hard-failure → exit non-zero with diagnostic. Soft-failure (no protected user row on fresh install) → log WARN, continue.

### 8.3 Day-of release sequence

Same shape as v1.2.3 and v1.2.4:

1. Cross-DB CI matrix green
2. Local Docker deploy + browser smoke test (per the deploy-before-merge memory). Smoke includes: protected-user UX (banner + structured 403), recovery CLI flow (drop trigger → verify → reapply → real UPDATE rejected)
3. Squash-merge PR → main
4. Tag `v1.3.0`, push tag
5. Docker rebuild from main → tag `:dev`, `:latest`, `:1.3.0`, push all three
6. CHANGELOG and version-stamped docs catch-up commit

### 8.4 Production monitoring

| Signal | Source | Healthy state | Investigate when |
|--------|--------|---------------|------------------|
| `/health` body | HTTP probe | `{"status":"healthy",...}` | Status flips to `degraded` |
| Boot logs | Process logs | "protected-user invariant: 3/3 checks passed" | Anything else |
| `protected_user_attack_*` rate | `audit_logs` table | 0–few/week | Spike, especially `_db` events |
| Audit event source classification | `audit_logs.details.user_agent` | Mostly browser UAs | curl/python/empty UAs → check `actor_id`'s session |
| `protected_invariant_degraded` log lines | Process logs | Absent | Any occurrence — page/alert |
| Dependabot PRs touching `pkg/security/**` | GitHub | Don't auto-merge | Always require human review (CODEOWNERS) |

Any single `_db` event means L1+L2 were bypassed somehow — alert rule is `count >= 1`, not `count > N`.

### 8.5 Backout

| Failure | Backout |
|---------|---------|
| Migration fails on a customer's DB | Pre-deployment matrix CI catches; if it slips, hotfix v1.3.0.1; customer rolls back to v1.2.4 in the meantime |
| Boot invariant has a false-positive bug | Customer sets `ACTALOG_SKIP_PROTECTED_INVARIANT=true` (degraded mode), files incident, we ship v1.3.0.1 |
| Profile tab breaks for some user | New surface; doesn't affect existing admin user-list. Hotfix in v1.3.0.1 |
| L3 trigger fires unexpectedly on customer data | Trigger only matches the protected email; no cross-row interaction possible |

Customer rollback procedure: set the env var, roll binary back to v1.2.4. Triggers stay in the DB and keep enforcing protection at L3 even on the older binary. **Rolling back the binary does not roll back security** — by design.

### 8.6 v1.3.0 acceptance checklist

- [ ] CI matrix green on sqlite3 / postgres / mysql
- [ ] All recovery tests pass on all three DBs
- [ ] Adversarial bypass tests prove each layer catches independently
- [ ] Generator drift check passes
- [ ] Migration ↔ recovery script lockstep check passes
- [ ] CODEOWNERS file updated
- [ ] Local Docker smoke test: edit protected user via UI shows banner; via curl returns structured 403
- [ ] Local Docker smoke test: drop trigger → verify CLI fails → reapply CLI restores → real UPDATE rejected
- [ ] All four new docs created; existing docs updated; doc code-block CI lint passes
- [ ] OpenAPI spec includes new endpoints with 403 documented
- [ ] Coverage gate: 100% on `pkg/security/`, 100% branch on `pkg/middleware/protected_user.go`, 90%+ on `internal/service/admin_user_service.go`

## 9. Out of scope for v1.3.0

Explicitly deferred to keep v1.3.0 small and shippable:

- Affiliations, Subscriptions, Credits & Documents, Preferences, Activity & Audit tabs (later releases)
- Coach assignment per gym (v1.3.1)
- Bulk user actions (post-1.3 feature)
- Inline avatar editing inside the edit screen — use existing `/profile` flow
- Self-service user-facing portal (not an admin feature)
- Email-change re-verification customization beyond the existing template

## 10. Decision log

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Layout | Tabs | Matches `AdminPackagesView.vue` convention; supports deep-linking; clean lazy-load per tab |
| Route shape | Dedicated route `/admin/users/:id/edit` | Bookmarkable, shareable, browser back/forward; mirrors `AdminOrganizationDetailView` |
| Edit pattern | Draft + Save per tab | Cleaner audit log; tolerates typos; per-tab atomic unit matches per-tab service split |
| Defense depth | L1+L2+L3+L4 + boot check | Each layer independently sufficient; rollback property ("security stays on") falls out of L3 living in DB |
| Guard placement | Middleware (L1) + service helper (L2) + DB trigger (L3) + audit (L4) | Belt + suspenders + DB enforcement + alarm; no single layer is the only protection |
| Protected list location | Hardcoded Go constant + migration-embedded SQL | No runtime path to change = no attacker path to disable |
| Protected list change procedure | PR with code change + new migration + CODEOWNERS + boot-time invariant verification | Friction is the security boundary |
| Tab ship order | v1.3.0: Profile + framework; v1.3.1: Affiliations; v1.3.2: rest | Smallest tab first as POC of the whole pattern; motivating example second; bulk later |
| Frontend protected-list source | Auto-generated from Go | Single source of truth; CI fails on drift |

---

*Spec produced via `superpowers:brainstorming` skill on 2026-04-28. Implementation plan to follow via `superpowers:writing-plans`.*
