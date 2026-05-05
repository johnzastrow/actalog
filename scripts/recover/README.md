# scripts/recover — Last-Resort Protected-User Trigger Recovery

## When to use this

**Prefer the in-binary CLI first:**

```bash
./bin/actalog admin reapply-protected-migrations --confirm
```

Use the tools in *this* directory only when the actalog binary itself is
unrunnable (e.g. corrupt binary, missing dependencies, boot crash). In that
case you can apply the SQL directly via a database CLI.

See [docs/security/PROTECTED_USERS.md](../../docs/security/PROTECTED_USERS.md)
for deeper context on the protected-user security model.

---

## Per-driver invocation examples

### SQLite

```bash
sqlite3 ./data/actalog.db < scripts/recover/sql/sqlite/protected_users.sql
```

Or using the wrapper:

```bash
./scripts/recover/restore-protected-triggers.sh \
  --driver=sqlite3 \
  --conn=./data/actalog.db
```

### PostgreSQL

```bash
psql 'postgres://user:pass@host:5432/db?sslmode=require' \
  -f scripts/recover/sql/postgres/protected_users.sql
```

Or using the wrapper:

```bash
./scripts/recover/restore-protected-triggers.sh \
  --driver=postgres \
  --conn='postgres://user:pass@host:5432/db?sslmode=require'
```

### MySQL / MariaDB

```bash
mysql -h host -u user -p db_name \
  < scripts/recover/sql/mysql/protected_users.sql
```

Or using the wrapper (see MySQL caveat below):

```bash
./scripts/recover/restore-protected-triggers.sh \
  --driver=mysql \
  --conn='user:pass@tcp(host:3306)/db'
```

---

## MySQL connection-string caveat

The shell wrapper accepts Go's standard DSN format (`user:pass@tcp(host:port)/db`)
and attempts to parse it via `sed`/`cut` into `mysql` CLI flags. This parsing
is best-effort and may fail for passwords that contain special characters
(`@`, `:`, `/`). If the wrapper fails for MySQL, invoke the `mysql` client
directly as shown above — it is safer and equally effective.

---

## Where the SQL files come from

The three files under `sql/` are **not hand-edited**. They are extracted
byte-for-byte from the LOCKSTEP-bracketed bodies inside:

```
internal/repository/protected_triggers_sql.go
```

Each dialect's block is bounded by:

```sql
-- LOCKSTEP-START <dialect>
...
-- LOCKSTEP-END <dialect>
```

A CI lockstep test (`scripts/recover/lockstep_test.go`) verifies that the
recovery scripts and the migration constants stay in sync. **Never edit the
`.sql` files directly** — edit the Go constants and re-extract instead.

To re-extract manually:

1. Open `internal/repository/protected_triggers_sql.go`.
2. Copy the text between `-- LOCKSTEP-START <dialect>` and `-- LOCKSTEP-END <dialect>` (excluding the marker lines).
3. Overwrite `scripts/recover/sql/<dialect>/protected_users.sql`.
4. Run `go test -count=1 ./scripts/recover/...` — must PASS before committing.

---

## Safety warnings

- These scripts use `CREATE OR REPLACE` / `DROP TRIGGER IF EXISTS` — they are
  idempotent. Running them twice is safe.
- Always restart actalog after applying and verify:
  ```bash
  ./bin/actalog admin verify-protected-users --verbose
  ```
- Do **not** apply these scripts on a live production database without
  confirming there is no active migration in progress.
