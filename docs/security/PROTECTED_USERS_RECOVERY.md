# Protected Users — Incident Response Playbook

**USE THIS WHEN** you are paged about a protected-user invariant failure and need to act NOW.

For design rationale, architecture details, and the full recovery reference, see
`docs/security/PROTECTED_USERS.md`.

---

## QUICK REFERENCE — error message to command

| If you see ... | Run this first |
|---|---|
| `"FATAL: protected-user invariant failed"` on boot | `./bin/actalog admin verify-protected-users --verbose` |
| `trigger "protected_users_no_update" missing` | `./bin/actalog admin reapply-protected-migrations --confirm` |
| `trigger fires but rejects with wrong message` | `./bin/actalog admin reapply-protected-migrations --confirm` |
| `/health returns 503` + `degraded` in logs | See [Degraded Mode](#degraded-mode) below |
| `binary won't even start the CLI` | `./scripts/recover/restore-protected-triggers.sh --driver=... --conn=...` |
| `protected user "br8kwall@gmail.com" missing while N other users exist` | Restore from backup — see [Backup Restoration](#backup-restoration) |

---

## Annotated boot-failure output

When the binary refuses to start you will see this on stderr:

```text
[ERROR] protected-user invariant failed: invariant: trigger "protected_users_no_update" missing — recover with: ./bin/actalog admin reapply-protected-migrations --confirm
         ↑ entry point — read the failure reason on this line
[FATAL] Refusing to start. To recover (most → least disruptive):
  1) ./bin/actalog admin reapply-protected-migrations --confirm
     ↑ try this first — idempotent, transactional
  2) ./bin/actalog admin verify-protected-users --verbose
     ↑ read-only diagnostic — safe to run while a previous instance is still up
  3) scripts/recover/restore-protected-triggers.sh
     ↑ last resort — use only when the binary itself is unrunnable

Full runbook: docs/security/PROTECTED_USERS.md#recovery
```

The failure reason on the `[ERROR]` line tells you which of the three scenarios below applies.

---

## Decision flowchart

```text
Can the binary boot?
├── YES → run verify first
│   │   ./bin/actalog admin verify-protected-users --verbose
│   ├── 3/3 ✓ → done (false alarm or already fixed)
│   └── any ✗ → ./bin/actalog admin reapply-protected-migrations --confirm
│       └── still failing? → ./scripts/recover/restore-protected-triggers.sh
│
├── NO (boot fails — FATAL on stderr) → read the [ERROR] line above FATAL
│   ├── "trigger missing" → ./bin/actalog admin reapply-protected-migrations --confirm
│   ├── "protected row ... missing while N other users exist" → restore from backup
│   └── "invariant query failed" / DB connection error → fix DB connectivity first
│
└── NO (binary unrunnable — crashloop, no shell) → degraded mode escape hatch
    │   ACTALOG_SKIP_PROTECTED_INVARIANT=true ./bin/actalog
    └── once running → ./bin/actalog admin reapply-protected-migrations --confirm
        └── verify passes? → restart WITHOUT the env var
```

---

## Scenario A — trigger missing or broken

**Symptoms:** boot fails with `trigger "protected_users_no_update" missing` or
`simulated UPDATE ... did not reject as expected`.

**Step 1 — diagnose (read-only, safe on live DB):**

```bash
./bin/actalog admin verify-protected-users --verbose
```

**Step 2 — reapply triggers:**

```bash
./bin/actalog admin reapply-protected-migrations --confirm
```

The command drops any existing protected-user triggers and recreates them from
the compiled-in SQL. Runs in a transaction — partial failure leaves the DB
unchanged. Re-runs the invariant check internally and exits non-zero if still
broken.

**Step 3 — restart and verify:**

```bash
pkill -9 -f actalog
./bin/actalog
./bin/actalog admin verify-protected-users
```

Expected final output:

```text
Protected-user invariant:
  Check 1/3 — L3 triggers exist        ✓
  Check 2/3 — triggers fire correctly  ✓
  Check 3/3 — protected rows exist     ✓

✓ all checks passed
```

---

## Scenario B — binary unrunnable (crashloop / no CLI access)

Use degraded mode to get the app serving traffic while you repair:

```bash
ACTALOG_SKIP_PROTECTED_INVARIANT=true ./bin/actalog
```

Then repair from inside the running process:

```bash
./bin/actalog admin reapply-protected-migrations --confirm
```

Then restart without the env var. Exit criteria: `verify-protected-users` shows
3/3 ✓ and `/health` returns `200 {"status":"healthy"}`.

---

## Degraded mode

Entered when `ACTALOG_SKIP_PROTECTED_INVARIANT=true` is set **or** when the
binary crashes past the boot invariant check.

**What still works:** all reads, login, workout logging, admin reads.

**What is blocked:** every non-GET/HEAD under `/api/admin/*` returns `503`
with body `{"error":"protected_invariant_degraded","message":"..."}`.

**Monitoring signals while degraded:**

- `/health` → `503 {"status":"degraded","version":"..."}`
- Log line every 60 s: `[ERROR] protected_invariant_degraded heartbeat: ...`
- `audit_logs` row: `event_type = "protected_invariant_degraded"`

**Exit criteria:**

```bash
# 1. Fix the invariant (run reapply, or restore from backup)
./bin/actalog admin reapply-protected-migrations --confirm

# 2. Remove ACTALOG_SKIP_PROTECTED_INVARIANT from env / .env

# 3. Restart
pkill -9 -f actalog && ./bin/actalog

# 4. Confirm
curl -s http://localhost:8080/health
# expect: {"status":"healthy","version":"..."}
./bin/actalog admin verify-protected-users
# expect: ✓ all checks passed
```

---

## Per-dialect paste-ready commands

### reapply-protected-migrations (preferred — uses compiled-in SQL)

The binary reads `DB_DRIVER` and the connection string from `.env` automatically.
No driver flag needed:

```bash
./bin/actalog admin reapply-protected-migrations --confirm
```

### restore-protected-triggers.sh (last resort — binary unrunnable)

**SQLite:**

```bash
./scripts/recover/restore-protected-triggers.sh \
  --driver=sqlite3 \
  --conn=./data/actalog.db
```

**PostgreSQL:**

```bash
./scripts/recover/restore-protected-triggers.sh \
  --driver=postgres \
  --conn='postgres://USER:PASSWORD@HOST:5432/DB?sslmode=require'
```

**MySQL / MariaDB:**

```bash
./scripts/recover/restore-protected-triggers.sh \
  --driver=mysql \
  --conn='USER:PASSWORD@tcp(HOST:3306)/DB'
```

After either script completes, verify and restart:

```bash
./bin/actalog admin verify-protected-users --verbose
pkill -9 -f actalog && ./bin/actalog
```

---

## Backup restoration

**When to use:** boot fails with `"protected user ... missing while N other users exist"`.
`reapply-protected-migrations` does NOT restore deleted user rows — only the backup can.

**30-second version:**

1. Find the most recent backup that contains the row.
2. Extract the row and insert it into the live DB.
3. Restart; verify shows 3/3 ✓.

For the full step-by-step procedure including per-dialect SQL (SQLite `INSERT OR IGNORE`,
PostgreSQL `pg_dump --where`, MySQL `mysqldump --where`), see
`docs/security/PROTECTED_USERS.md` §8.4.

After restoring the row, reapply triggers to be safe, then restart:

```bash
./bin/actalog admin reapply-protected-migrations --confirm
pkill -9 -f actalog && ./bin/actalog
./bin/actalog admin verify-protected-users
```

---

## If none of these work

Contact the security team. Full background and additional scenarios (rogue migrations,
pre-trigger backup restores, CI drift) are in `docs/security/PROTECTED_USERS.md`.
