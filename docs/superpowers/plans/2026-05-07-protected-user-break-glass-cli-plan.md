# Protected-User Break-Glass CLI Implementation Plan (v1.3.2)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `actalog admin force-edit-protected` CLI subcommand for operator-driven recovery of protected user accounts, plus collapse the v1.3.1 admin user lifecycle work into v1.3.2 (single PR / tag / CHANGELOG).

**Architecture:** New `internal/protectedusers/breakglass.go` providing `AdminForceEditProtected` with per-field handlers (password / email / name / role / account_disabled). Reuses existing `pkg/auth.HashPassword`, `security.IsProtectedEmail`, the trigger SQL constants in `internal/repository/protected_triggers_sql.go`, and `VerifyProtectedUserInvariant` for post-edit verification. Five new per-field audit event constants. Wired into the existing admin-dispatch switch in `cmd/actalog/main.go`.

**Tech Stack:** Go (Chi router), bcrypt, SQLite/PostgreSQL/MySQL drivers, `internal/protectedusers/` package.

**Spec:** `/home/jcz/Github/actionlog/.worktrees/feature/admin-user-lifecycle-v1.3.1/docs/superpowers/specs/2026-05-07-protected-user-break-glass-cli-design.md`

---

## File map

**New backend files:**
- `internal/protectedusers/breakglass.go` — `AdminForceEditProtected` + per-field handlers + helpers
- `internal/protectedusers/breakglass_test.go` — unit tests
- `test/integration/protected_users_break_glass_test.go` — multi-DB end-to-end

**Modified backend files:**
- `internal/domain/audit_log.go` — five new event constants
- `cmd/actalog/main.go` — extend admin-dispatch switch with `force-edit-protected` case

**Modified docs/version:**
- `docs/security/PROTECTED_USERS.md` — new "Break-glass CLI" subsection
- `docs/security/PROTECTED_USERS_RECOVERY.md` — point at the CLI
- `docs/security/THREAT_MODEL.md` — shell-access residual-risk row
- `docs/CHANGELOG.md` — collapse v1.3.1 → v1.3.2 with both areas
- `docs/USER_PERMISSIONS.md` — operator CLI note
- `docs/TODO.md` — mark break-glass line complete
- `pkg/version/version.go` — Patch 1 → 2
- `web/package.json` — 1.3.1 → 1.3.2
- `CLAUDE.md` — version + docker examples
- `secrets/reset-password.sh` — header comment updated to point at CLI as preferred path (script kept as fallback)

---

## Conventions

- Paths absolute. Backend tests: `go test ./<pkg>/ -count=1`. Build: `go build ./cmd/actalog/`.
- The implementer ALWAYS dispatches in the worktree at `/home/jcz/Github/actionlog/.worktrees/feature/admin-user-lifecycle-v1.3.1`.
- Existing helpers to call (verified to exist):
  - `pkg/auth.HashPassword(string) (string, error)` — bcrypt cost 12
  - `pkg/security.IsProtectedEmail(string) bool` and `ProtectedEmailsList() []string`
  - `internal/repository.SplitProtectedTriggerSQL(constant)` — splits dialect SQL into individual statements
  - `internal/repository.SQLiteProtectedTriggers`, `PostgresProtectedTriggers`, `MySQLProtectedTriggers` — dialect SQL constants
  - `internal/protectedusers.VerifyProtectedUserInvariant(db, driver) (*InvariantReport, error)`
  - `internal/protectedusers.AdminReapplyProtectedMigrations(db, driver, true, &out)` — full reinstall path used as recovery on identity-field handler errors
  - `repository.NewSQLiteUserRepository`, `NewSQLiteRefreshTokenRepository` — but for direct SQL the CLI uses `db.Exec` with `rebindQuery` since we need to write columns the repo doesn't expose (e.g., `account_disabled`, `disabled_at`)

---

## Task 1: Add five audit event constants

**Files:**
- Modify: `internal/domain/audit_log.go`

- [ ] **Step 1: Add the constants**

In `internal/domain/audit_log.go`, find the existing `EventProtectedUserAttack*` block (around line 165). After `EventProtectedUserAttackDB`, add:

```go
// Break-glass operator CLI events (v1.3.2)
//
// Per-field events (rather than one event with a `field` discriminator)
// to enable clean alert routing and per-field log queries — e.g.
//   SELECT ... WHERE event_type = 'protected_user_break_glass_email'
// is a single index lookup, no JSON-path filter on details.
EventProtectedUserBreakGlassPassword        = "protected_user_break_glass_password"
EventProtectedUserBreakGlassEmail           = "protected_user_break_glass_email"
EventProtectedUserBreakGlassName            = "protected_user_break_glass_name"
EventProtectedUserBreakGlassRole            = "protected_user_break_glass_role"
EventProtectedUserBreakGlassAccountDisabled = "protected_user_break_glass_account_disabled"
```

- [ ] **Step 2: Verify build**

```bash
go build ./internal/domain/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/domain/audit_log.go
git commit -m "feat(audit): break-glass per-field event constants

Five new event types for v1.3.2 protected-user break-glass CLI:
- protected_user_break_glass_password
- protected_user_break_glass_email
- protected_user_break_glass_name
- protected_user_break_glass_role
- protected_user_break_glass_account_disabled"
```

---

## Task 2: Implement `internal/protectedusers/breakglass.go`

**Files:**
- Create: `internal/protectedusers/breakglass.go`
- Create: `internal/protectedusers/breakglass_test.go`

This is the largest task. Single file with `AdminForceEditProtected` plus per-field helpers plus tests. Implementer should follow the spec's per-field rules table closely.

- [ ] **Step 1: Write `breakglass.go`**

Create `internal/protectedusers/breakglass.go`:

```go
package protectedusers

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/auth"
	"github.com/johnzastrow/actalog/pkg/security"
)

// BreakGlassOptions captures every input AdminForceEditProtected needs.
type BreakGlassOptions struct {
	Email    string  // target user's email — must be on the protected list
	Field    string  // password|email|name|role|account_disabled
	Value    string  // empty for password (read from Stdin)
	Stdin    io.Reader // for password reads (use os.Stdin in normal use; bytes.Buffer in tests)
	Stdout   io.Writer // operator UX output
	Confirm  bool    // --confirm flag must be present
	// IdentityConfirmation is the literal string the operator must type when changing
	// identity fields ("BREAK-GLASS"). Pass "" in non-interactive tests for tests that
	// pre-read the prompt.
	IdentityConfirmation string
}

// AdminForceEditProtected is the entry point invoked by the CLI.
// On success it writes the requested change, fires the per-field audit event,
// and re-runs the boot invariant. Returns an error if anything in the chain fails.
//
// For identity fields (email|name|role|account_disabled), this temporarily
// drops the L3 triggers, performs the UPDATE, then reinstalls. If reinstall
// fails the function panics with a recovery message — the binary is the only
// thing that knows the canonical trigger SQL, and a partial state is worse
// than a hard exit.
func AdminForceEditProtected(db *sql.DB, driver string, opts BreakGlassOptions) error {
	// 1. Validate flags + target.
	if !opts.Confirm {
		return errors.New("--confirm flag is required")
	}
	target := strings.ToLower(strings.TrimSpace(opts.Email))
	if !security.IsProtectedEmail(target) {
		return fmt.Errorf("target %q is not on the protected list — use the normal admin path", target)
	}

	// 2. Resolve target user record.
	userRepo := repository.NewSQLiteUserRepository(db)
	user, err := userRepo.GetByEmail(target)
	if err != nil {
		return fmt.Errorf("look up target: %w", err)
	}
	if user == nil {
		return fmt.Errorf("target %q not found in users table", target)
	}

	// 3. Build operator metadata for audit details.
	meta := captureOperatorMetadata()

	// 4. Dispatch by field.
	auditDetails := map[string]interface{}{
		"operator_user":     meta.user,
		"operator_hostname": meta.hostname,
		"operator_tty":      meta.tty,
		"operator_cwd":      meta.cwd,
	}
	var eventType string
	switch opts.Field {
	case "password":
		eventType = domain.EventProtectedUserBreakGlassPassword
		if err := handlePasswordField(db, driver, user, opts, auditDetails); err != nil {
			return err
		}
	case "email":
		eventType = domain.EventProtectedUserBreakGlassEmail
		if err := handleIdentityField(db, driver, user, "email", opts, auditDetails); err != nil {
			return err
		}
	case "name":
		eventType = domain.EventProtectedUserBreakGlassName
		if err := handleIdentityField(db, driver, user, "name", opts, auditDetails); err != nil {
			return err
		}
	case "role":
		eventType = domain.EventProtectedUserBreakGlassRole
		if err := handleIdentityField(db, driver, user, "role", opts, auditDetails); err != nil {
			return err
		}
	case "account_disabled":
		eventType = domain.EventProtectedUserBreakGlassAccountDisabled
		if err := handleIdentityField(db, driver, user, "account_disabled", opts, auditDetails); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown field %q (allowed: password|email|name|role|account_disabled)", opts.Field)
	}

	// 5. Re-run boot invariant.
	if _, invErr := VerifyProtectedUserInvariant(db, driver); invErr != nil {
		return fmt.Errorf("post-edit invariant FAILED: %w  — protected-user defense may be DEGRADED; run: ./bin/actalog admin reapply-protected-migrations --confirm", invErr)
	}

	// 6. Write audit event. Build the minimal audit graph inline (same pattern
	//    AdminReapplyProtectedMigrations uses).
	auditRepo := repository.NewAuditLogRepository(db, driver)
	auditSvc := service.NewAuditLogService(auditRepo)
	targetID := user.ID
	if auditErr := auditSvc.LogEvent(eventType, nil, &targetID, nil, nil, auditDetails); auditErr != nil {
		fmt.Fprintf(opts.Stdout, "warn: failed to write audit event: %v\n", auditErr)
	}

	fmt.Fprintf(opts.Stdout, "Done. Audit event written: %s\n", eventType)
	return nil
}

// handlePasswordField runs the password-change branch. No trigger fiddle —
// password_hash is a lifecycle field and L3 lets it through.
func handlePasswordField(db *sql.DB, driver string, user *domain.User, opts BreakGlassOptions, details map[string]interface{}) error {
	if opts.Value != "" {
		return errors.New("password must be read from stdin, not --value (security)")
	}
	pw, err := readPasswordTwice(opts.Stdin, opts.Stdout)
	if err != nil {
		return err
	}
	if err := validatePasswordPolicy(pw); err != nil {
		return err
	}
	hash, err := auth.HashPassword(pw)
	if err != nil {
		return fmt.Errorf("hash: %w", err)
	}
	priorAttempts := user.FailedLoginAttempts
	priorLocked := user.LockedUntil != nil

	userRepo := repository.NewSQLiteUserRepository(db)
	if err := userRepo.UpdatePassword(user.ID, hash); err != nil {
		return fmt.Errorf("UpdatePassword: %w", err)
	}
	if err := userRepo.UnlockAccount(user.ID); err != nil {
		return fmt.Errorf("UnlockAccount: %w", err)
	}
	tokenRepo := repository.NewSQLiteRefreshTokenRepository(db)
	if err := tokenRepo.RevokeAllForUser(user.ID); err != nil {
		fmt.Fprintf(opts.Stdout, "warn: RevokeAllForUser: %v\n", err)
	}

	details["cleared_failed_login_attempts"] = priorAttempts
	details["cleared_lockout"] = priorLocked
	details["revoked_refresh_tokens"] = true
	details["trigger_dropped"] = false
	return nil
}

// handleIdentityField runs the identity-field branch. Drops L3 triggers,
// performs the column-specific UPDATE, then reinstalls triggers. Reinstall
// failure is fatal — call site panics with a recovery message.
func handleIdentityField(db *sql.DB, driver string, user *domain.User, field string, opts BreakGlassOptions, details map[string]interface{}) error {
	if err := requireIdentityConfirmation(opts); err != nil {
		return err
	}
	if opts.Value == "" {
		return fmt.Errorf("--value is required for field %q", field)
	}

	// Resolve old + new values per field.
	var oldVal, newVal string
	switch field {
	case "email":
		addr, parseErr := mail.ParseAddress(opts.Value)
		if parseErr != nil {
			return fmt.Errorf("invalid email: %w", parseErr)
		}
		newLower := strings.ToLower(addr.Address)
		if security.IsProtectedEmail(newLower) {
			return errors.New("new email is on the protected list — refusing to migrate to a protected address")
		}
		// Uniqueness check.
		userRepo := repository.NewSQLiteUserRepository(db)
		existing, _ := userRepo.GetByEmail(newLower)
		if existing != nil && existing.ID != user.ID {
			return fmt.Errorf("email %q already in use by user id %d", newLower, existing.ID)
		}
		oldVal = user.Email
		newVal = newLower
	case "name":
		newName := strings.TrimSpace(opts.Value)
		if newName == "" || len(newName) > 100 {
			return errors.New("name must be 1–100 characters")
		}
		oldVal = user.Name
		newVal = newName
	case "role":
		if opts.Value != "athlete" && opts.Value != "coach" && opts.Value != "admin" {
			return errors.New("role must be athlete | coach | admin")
		}
		oldVal = user.Role
		newVal = opts.Value
	case "account_disabled":
		boolVal, err := parseBool(opts.Value)
		if err != nil {
			return err
		}
		oldVal = fmt.Sprintf("%t", user.AccountDisabled)
		newVal = fmt.Sprintf("%t", boolVal)
	default:
		return fmt.Errorf("unsupported identity field %q", field)
	}

	// Drop triggers.
	if err := dropProtectedTriggers(db, driver); err != nil {
		return fmt.Errorf("drop triggers: %w", err)
	}

	// Perform the per-field UPDATE.
	updateErr := executeIdentityUpdate(db, driver, user.ID, field, newVal, captureOperatorMetadata().user)

	// Reinstall triggers UNCONDITIONALLY.
	var reinstallOut strings.Builder
	if reinstallErr := AdminReapplyProtectedMigrations(db, driver, true, &reinstallOut); reinstallErr != nil {
		// FATAL — system is now in a degraded protected-user state.
		panic(fmt.Sprintf("FATAL: trigger reinstall failed after edit. Run NOW: ./bin/actalog admin reapply-protected-migrations --confirm. Original update err: %v. Reinstall err: %v. Reinstall output:\n%s", updateErr, reinstallErr, reinstallOut.String()))
	}
	if updateErr != nil {
		return fmt.Errorf("UPDATE %s: %w", field, updateErr)
	}

	details["old_value"] = oldVal
	details["new_value"] = newVal
	details["trigger_dropped"] = true
	return nil
}

// executeIdentityUpdate performs the SQL UPDATE for the given identity field.
// account_disabled also writes disabled_at, disabled_by_user_id (NULL — no admin
// actor), and disable_reason ("break-glass: <op-user>").
func executeIdentityUpdate(db *sql.DB, driver string, userID int64, field string, newVal string, operator string) error {
	now := time.Now().UTC()
	var query string
	var args []interface{}

	switch field {
	case "email", "name", "role":
		// Simple single-column update.
		// Use rebindQuery for postgres (?-style works on sqlite/mysql).
		base := fmt.Sprintf("UPDATE users SET %s = ?, updated_at = ? WHERE id = ?", field)
		query = repository.RebindQuery(base, driver)
		args = []interface{}{newVal, now, userID}
	case "account_disabled":
		boolVal := newVal == "true"
		var disabledAt interface{}
		var disableReason interface{}
		if boolVal {
			disabledAt = now
			disableReason = fmt.Sprintf("break-glass: %s", operator)
		} else {
			disabledAt = nil
			disableReason = nil
		}
		base := "UPDATE users SET account_disabled = ?, disabled_at = ?, disabled_by_user_id = NULL, disable_reason = ?, updated_at = ? WHERE id = ?"
		query = repository.RebindQuery(base, driver)
		// Driver-specific bool: getBoolValue(driver, boolVal) is needed for postgres,
		// but the repository package's existing pattern stores account_disabled as
		// boolean for postgres/mysql and integer for sqlite. Use the existing helper
		// `repository.GetBoolValue(driver, boolVal)` if exposed; otherwise just pass
		// the bool — drivers handle conversion.
		args = []interface{}{boolVal, disabledAt, disableReason, now, userID}
	default:
		return fmt.Errorf("executeIdentityUpdate: unsupported field %q", field)
	}
	_, err := db.Exec(query, args...)
	return err
}

// dropProtectedTriggers drops both protected_users_no_update and
// protected_users_no_delete using per-driver syntax. Postgres requires
// `ON users` suffix; sqlite3 and mysql do not.
func dropProtectedTriggers(db *sql.DB, driver string) error {
	for _, trigger := range []string{"protected_users_no_update", "protected_users_no_delete"} {
		var stmt string
		switch driver {
		case "postgres":
			stmt = fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON users", trigger)
		case "sqlite3", "mysql":
			stmt = fmt.Sprintf("DROP TRIGGER IF EXISTS %s", trigger)
		default:
			return fmt.Errorf("unsupported driver %q", driver)
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("drop %s: %w", trigger, err)
		}
	}
	return nil
}

// readPasswordTwice reads a password from stdin twice and verifies they match.
// In tests, stdin is a bytes.Buffer with two newline-separated lines.
func readPasswordTwice(in io.Reader, out io.Writer) (string, error) {
	if in == nil { in = os.Stdin }
	if out == nil { out = os.Stderr }
	fmt.Fprint(out, "New password: ")
	pw, err := readLine(in)
	if err != nil { return "", err }
	fmt.Fprint(out, "Confirm:      ")
	confirm, err := readLine(in)
	if err != nil { return "", err }
	if pw != confirm {
		return "", errors.New("passwords don't match")
	}
	return pw, nil
}

// readLine reads up to a newline. For terminal stdin, the caller should arrange
// for echo-off via golang.org/x/term in the CLI wrapper; here we keep it simple
// for testability.
func readLine(in io.Reader) (string, error) {
	var sb strings.Builder
	buf := make([]byte, 1)
	for {
		n, err := in.Read(buf)
		if n > 0 {
			if buf[0] == '\n' { return sb.String(), nil }
			if buf[0] != '\r' { sb.WriteByte(buf[0]) }
		}
		if err == io.EOF {
			if sb.Len() == 0 { return "", err }
			return sb.String(), nil
		}
		if err != nil { return "", err }
	}
}

// requireIdentityConfirmation enforces the "type BREAK-GLASS to proceed" gate
// for identity-field changes. If opts.IdentityConfirmation is non-empty, that
// value is used (test path); otherwise reads from opts.Stdin.
func requireIdentityConfirmation(opts BreakGlassOptions) error {
	const required = "BREAK-GLASS"
	if opts.IdentityConfirmation != "" {
		if opts.IdentityConfirmation != required {
			return fmt.Errorf("identity confirmation must be exactly %q", required)
		}
		return nil
	}
	if opts.Stdin == nil { return errors.New("identity confirmation required but no stdin provided") }
	fmt.Fprintf(opts.Stdout, "Type 'BREAK-GLASS' to proceed: ")
	line, err := readLine(opts.Stdin)
	if err != nil { return fmt.Errorf("confirmation read: %w", err) }
	if strings.TrimSpace(line) != required {
		return fmt.Errorf("identity confirmation must be exactly %q", required)
	}
	return nil
}

// validatePasswordPolicy mirrors validatePassword in user_service.go:
//   ≥12 chars, at least one upper, one lower, one digit.
func validatePasswordPolicy(pw string) error {
	if len(pw) < 12 {
		return errors.New("password must be at least 12 characters")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, c := range pw {
		switch {
		case c >= 'A' && c <= 'Z': hasUpper = true
		case c >= 'a' && c <= 'z': hasLower = true
		case c >= '0' && c <= '9': hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("password must include uppercase, lowercase, and a digit")
	}
	return nil
}

// parseBool accepts true|false|1|0|yes|no (case-insensitive).
func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes", "y": return true, nil
	case "false", "0", "no", "n": return false, nil
	}
	return false, fmt.Errorf("could not parse %q as bool (use true|false|1|0|yes|no)", s)
}

type operatorMetadata struct {
	user, hostname, tty, cwd string
}

func captureOperatorMetadata() operatorMetadata {
	m := operatorMetadata{}
	m.user = os.Getenv("USER")
	if m.user == "" { m.user = os.Getenv("LOGNAME") }
	if h, err := os.Hostname(); err == nil { m.hostname = h }
	if cwd, err := os.Getwd(); err == nil { m.cwd = cwd }
	// /dev/tty discovery is platform-specific; safe fallback:
	if tty := os.Getenv("SSH_TTY"); tty != "" {
		m.tty = tty
	}
	return m
}
```

**Note on `repository.RebindQuery` and `repository.GetBoolValue`:** these helpers may be unexported; check by reading `internal/repository/database.go` (where similar driver-specific helpers live). If they're unexported, either use `db.Exec` with the canonical driver-specific placeholders inline (sqlite3/mysql use `?`, postgres uses `$1, $2, ...`), OR add a small per-driver helper to the breakglass.go file. Don't export the repository helpers if they're currently unexported — keep the boundary clean.

The cleanest fix if helpers are unexported: write `rebindFor(driver, query)` directly inside breakglass.go.

- [ ] **Step 2: Verify build**

```bash
cd /home/jcz/Github/actionlog/.worktrees/feature/admin-user-lifecycle-v1.3.1
go build ./internal/protectedusers/
```

Fix any signature mismatches with the existing repository / service helpers.

- [ ] **Step 3: Write unit tests in `breakglass_test.go`**

Create `internal/protectedusers/breakglass_test.go`:

```go
package protectedusers

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// validatePasswordPolicy
func TestValidatePasswordPolicy(t *testing.T) {
	cases := []struct {
		pw   string
		ok   bool
	}{
		{"ValidPass123A",     true},
		{"short1A",           false}, // <12 chars
		{"alllower1234",      false}, // no upper
		{"ALLUPPER1234",      false}, // no lower
		{"NoDigitsHereXY",    false}, // no digit
	}
	for _, tc := range cases {
		err := validatePasswordPolicy(tc.pw)
		if (err == nil) != tc.ok {
			t.Errorf("validatePasswordPolicy(%q): got err=%v, want ok=%v", tc.pw, err, tc.ok)
		}
	}
}

// readPasswordTwice
func TestReadPasswordTwice(t *testing.T) {
	t.Run("matching", func(t *testing.T) {
		in := bytes.NewBufferString("ValidPass123A\nValidPass123A\n")
		out := &bytes.Buffer{}
		got, err := readPasswordTwice(in, out)
		if err != nil { t.Fatal(err) }
		if got != "ValidPass123A" { t.Errorf("got %q", got) }
	})
	t.Run("mismatch", func(t *testing.T) {
		in := bytes.NewBufferString("Foo123Bar456!\nFOO123Bar456!\n")
		_, err := readPasswordTwice(in, &bytes.Buffer{})
		if err == nil || !strings.Contains(err.Error(), "match") {
			t.Errorf("expected mismatch error, got %v", err)
		}
	})
}

// parseBool
func TestParseBool(t *testing.T) {
	cases := map[string]bool{
		"true": true, "True": true, "1": true, "yes": true, "y": true,
		"false": false, "0": false, "no": false, "n": false,
	}
	for in, want := range cases {
		got, err := parseBool(in)
		if err != nil { t.Errorf("parseBool(%q): %v", in, err); continue }
		if got != want { t.Errorf("parseBool(%q) = %v, want %v", in, got, want) }
	}
	if _, err := parseBool("maybe"); err == nil {
		t.Error("parseBool(\"maybe\") should error")
	}
}

// requireIdentityConfirmation
func TestRequireIdentityConfirmation(t *testing.T) {
	t.Run("explicit match", func(t *testing.T) {
		if err := requireIdentityConfirmation(BreakGlassOptions{IdentityConfirmation: "BREAK-GLASS"}); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("explicit mismatch", func(t *testing.T) {
		if err := requireIdentityConfirmation(BreakGlassOptions{IdentityConfirmation: "yes"}); err == nil {
			t.Error("expected mismatch error")
		}
	})
	t.Run("via stdin", func(t *testing.T) {
		out := &bytes.Buffer{}
		err := requireIdentityConfirmation(BreakGlassOptions{
			Stdin:  bytes.NewBufferString("BREAK-GLASS\n"),
			Stdout: out,
		})
		if err != nil { t.Fatal(err) }
		if !strings.Contains(out.String(), "BREAK-GLASS") {
			t.Errorf("prompt not written; got: %s", out.String())
		}
	})
}

// AdminForceEditProtected — input validation only (no DB)
func TestAdminForceEditProtected_RejectsNoConfirm(t *testing.T) {
	err := AdminForceEditProtected(nil, "sqlite3", BreakGlassOptions{Email: "x@y.z", Field: "password"})
	if err == nil || !strings.Contains(err.Error(), "--confirm") {
		t.Errorf("expected --confirm required error; got %v", err)
	}
}

func TestAdminForceEditProtected_RejectsNonProtectedEmail(t *testing.T) {
	err := AdminForceEditProtected(nil, "sqlite3", BreakGlassOptions{
		Email:  "random@example.com",
		Field:  "password",
		Confirm: true,
	})
	if err == nil || !strings.Contains(err.Error(), "not on the protected list") {
		t.Errorf("expected protected-list rejection; got %v", err)
	}
}

func TestAdminForceEditProtected_RejectsUnknownField(t *testing.T) {
	// Use a real protected email so we get past the protected check.
	if errors.Is(nil, nil) {} // keep imports happy
	err := AdminForceEditProtected(nil, "sqlite3", BreakGlassOptions{
		Email:   "br8kwall@gmail.com",
		Field:   "favorite_color",
		Confirm: true,
	})
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Errorf("expected unknown-field error; got %v", err)
	}
}
```

(The integration tests against real DB go in Task 4.)

- [ ] **Step 4: Run tests**

```bash
go test ./internal/protectedusers/ -count=1 -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/protectedusers/breakglass.go internal/protectedusers/breakglass_test.go
git commit -m "feat(breakglass): AdminForceEditProtected core + unit tests

Per-field handlers for password|email|name|role|account_disabled.
Password reads from stdin (no --value); identity fields require both
--confirm flag and 'BREAK-GLASS' typed confirmation. Identity changes
drop L3 triggers, UPDATE, reinstall, then re-run boot invariant; reinstall
failure panics with a recovery message. Operator metadata captured for
audit details. Reuses existing helpers — no new repo methods."
```

---

## Task 3: Wire the CLI subcommand into `cmd/actalog/main.go`

**Files:**
- Modify: `cmd/actalog/main.go`

- [ ] **Step 1: Extend the admin-dispatch switch**

Find the switch in `cmd/actalog/main.go` (around line 214 — currently has `verify-protected-users` and `reapply-protected-migrations`). Add a new case before the `default`:

```go
case "force-edit-protected":
	opts := protectedusers.BreakGlassOptions{
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Confirm: hasFlag(os.Args[3:], "--confirm"),
	}
	opts.Email = flagValue(os.Args[3:], "--email")
	opts.Field = flagValue(os.Args[3:], "--field")
	opts.Value = flagValue(os.Args[3:], "--value")
	if opts.Email == "" || opts.Field == "" {
		fmt.Fprintln(os.Stderr, "usage: actalog admin force-edit-protected --email <target> --field <field> [--value <v>] --confirm")
		os.Exit(2)
	}
	if err := protectedusers.AdminForceEditProtected(db, cfg.Database.Driver, opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
```

If `flagValue` doesn't exist, add it next to the existing `hasFlag` helper:

```go
// flagValue returns the value of --name in args, or "" if absent.
// Supports both "--name=value" and "--name value" forms.
func flagValue(args []string, name string) string {
	for i, a := range args {
		if a == name && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(a, name+"=") {
			return strings.TrimPrefix(a, name+"=")
		}
	}
	return ""
}
```

Also update the usage banner near line 211 to include the new subcommand.

- [ ] **Step 2: Build**

```bash
go build ./cmd/actalog/
```

- [ ] **Step 3: Commit**

```bash
git add cmd/actalog/main.go
git commit -m "feat(cli): wire 'actalog admin force-edit-protected' subcommand

New admin-dispatch case calls protectedusers.AdminForceEditProtected.
Adds flagValue helper for --name <value> / --name=value parsing.
Updates the usage banner."
```

---

## Task 4: Multi-DB integration tests

**Files:**
- Create: `test/integration/protected_users_break_glass_test.go`

- [ ] **Step 1: Write the test file**

Create `test/integration/protected_users_break_glass_test.go`:

```go
package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/protectedusers"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/pkg/auth"
)

// TestBreakGlass_Password verifies password change works end-to-end against the
// matrix DB and produces the expected audit event.
func TestBreakGlass_Password(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)

	stdin := bytes.NewBufferString("BreakGlassPass123\nBreakGlassPass123\n")
	stdout := &bytes.Buffer{}

	err := protectedusers.AdminForceEditProtected(db, driver, protectedusers.BreakGlassOptions{
		Email:   "br8kwall@gmail.com",
		Field:   "password",
		Stdin:   stdin,
		Stdout:  stdout,
		Confirm: true,
	})
	if err != nil { t.Fatalf("force-edit password: %v", err) }

	// Verify the new password validates against the stored hash.
	repo := repository.NewSQLiteUserRepository(db)
	stored, _ := repo.GetByEmail("br8kwall@gmail.com")
	if err := auth.CheckPassword(stored.PasswordHash, "BreakGlassPass123"); err != nil {
		t.Errorf("new password does not validate: %v", err)
	}
}

// TestBreakGlass_Role drops triggers, changes role, reinstalls, verifies invariant.
func TestBreakGlass_Role(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)

	err := protectedusers.AdminForceEditProtected(db, driver, protectedusers.BreakGlassOptions{
		Email:                "br8kwall@gmail.com",
		Field:                "role",
		Value:                "athlete",
		Stdout:               &bytes.Buffer{},
		Confirm:              true,
		IdentityConfirmation: "BREAK-GLASS",
	})
	if err != nil { t.Fatalf("force-edit role: %v", err) }

	repo := repository.NewSQLiteUserRepository(db)
	stored, _ := repo.GetByEmail("br8kwall@gmail.com")
	if stored.Role != "athlete" {
		t.Errorf("role = %q, want athlete", stored.Role)
	}

	// Verify L3 triggers reinstalled (write a name change attempt — should be blocked).
	_, attemptErr := db.Exec(`UPDATE users SET name = 'attack' WHERE email = 'br8kwall@gmail.com'`)
	if attemptErr == nil {
		t.Error("L3 trigger should still block identity-field changes after break-glass")
	}
}

// TestBreakGlass_AccountDisabled
func TestBreakGlass_AccountDisabled(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)

	err := protectedusers.AdminForceEditProtected(db, driver, protectedusers.BreakGlassOptions{
		Email:                "br8kwall@gmail.com",
		Field:                "account_disabled",
		Value:                "true",
		Stdout:               &bytes.Buffer{},
		Confirm:              true,
		IdentityConfirmation: "BREAK-GLASS",
	})
	if err != nil { t.Fatalf("force-edit account_disabled: %v", err) }

	repo := repository.NewSQLiteUserRepository(db)
	stored, _ := repo.GetByEmail("br8kwall@gmail.com")
	if !stored.AccountDisabled {
		t.Error("AccountDisabled should be true")
	}
}

// TestBreakGlass_Email_RejectsProtectedTarget verifies migrating to a protected
// address is rejected.
func TestBreakGlass_Email_RejectsProtectedTarget(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)

	err := protectedusers.AdminForceEditProtected(db, driver, protectedusers.BreakGlassOptions{
		Email:                "br8kwall@gmail.com",
		Field:                "email",
		Value:                "br8kwall@gmail.com",  // same protected address
		Stdout:               &bytes.Buffer{},
		Confirm:              true,
		IdentityConfirmation: "BREAK-GLASS",
	})
	if err == nil {
		t.Fatal("expected rejection of protected-list migration")
	}
	if !strings.Contains(err.Error(), "protected list") {
		t.Errorf("expected protected-list error; got: %v", err)
	}
}

// TestBreakGlass_AuditEventsFire — at least one event of the right type fires per call.
func TestBreakGlass_AuditEventsFire(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)

	err := protectedusers.AdminForceEditProtected(db, driver, protectedusers.BreakGlassOptions{
		Email:                "br8kwall@gmail.com",
		Field:                "name",
		Value:                "Renamed",
		Stdout:               &bytes.Buffer{},
		Confirm:              true,
		IdentityConfirmation: "BREAK-GLASS",
	})
	if err != nil { t.Fatalf("force-edit: %v", err) }

	auditRepo := repository.NewAuditLogRepository(db, driver)
	events, _ := auditRepo.ListByEventType(domain.EventProtectedUserBreakGlassName, 10, 0)
	if len(events) < 1 {
		t.Errorf("expected at least 1 EventProtectedUserBreakGlassName, got %d", len(events))
	}
}
```

If `auditRepo.ListByEventType` doesn't exist, query directly via `db.Query("SELECT event_type FROM audit_logs WHERE event_type = ? AND target_user_id = ?", ...)`.

- [ ] **Step 2: Run tests**

```bash
go test ./test/integration/ -run TestBreakGlass -count=1 -v
```

- [ ] **Step 3: Commit**

```bash
git add test/integration/protected_users_break_glass_test.go
git commit -m "test(integration): break-glass CLI end-to-end (multi-DB)

Five scenarios: password change validates via bcrypt; role change drops
triggers + reinstalls + L3 still blocks unauthorized name changes after;
account_disabled toggle persists; protected-list migration rejected;
audit event fires per call."
```

---

## Task 5: Documentation + version bump + CHANGELOG collapse

**Files:**
- Modify: `docs/security/PROTECTED_USERS.md` — new "Break-glass CLI" subsection under §recovery (or as its own §)
- Modify: `docs/security/PROTECTED_USERS_RECOVERY.md` — point at the CLI from the 3-AM playbook
- Modify: `docs/security/THREAT_MODEL.md` — shell-access residual-risk row
- Modify: `docs/CHANGELOG.md` — collapse v1.3.1 entry into v1.3.2 with both feature areas
- Modify: `docs/USER_PERMISSIONS.md` — operator-CLI note in the admin-only section
- Modify: `docs/TODO.md` — mark "Admin: break-glass CLI for protected users (v1.3.1)" complete and re-tag as v1.3.2
- Modify: `pkg/version/version.go` — `Patch = 1` → `Patch = 2`
- Modify: `web/package.json` — `"version": "1.3.1"` → `"version": "1.3.2"`
- Modify: `CLAUDE.md` — version line, docker-tag examples (`1.3.1` → `1.3.2`)
- Modify: `secrets/reset-password.sh` — header comment now points to the CLI as preferred path; script kept as fallback if binary itself is broken
- Run `cd web && npm install` to refresh `package-lock.json`

The CHANGELOG collapse: take the existing `## [1.3.1]` block (added by Task 14 of the v1.3.1 plan) and:
1. Change the header to `## [1.3.2] - YYYY-MM-DD — Admin user lifecycle + protected-user break-glass CLI`
2. Add a new "### Added" subsection above the existing one for the break-glass CLI:
   ```markdown
   ### Added — Operator break-glass CLI
   - **`actalog admin force-edit-protected`** — operator escape hatch for editing protected accounts when admin paths are unavailable. Five fields: `password | email | name | role | account_disabled`. Per-field audit events with operator metadata (USER, hostname, tty, cwd) for forensics.
   - Identity-field changes briefly drop L3 triggers, UPDATE, then reinstall + re-run boot invariant. Reinstall failure panics with a recovery message — system never silently lands in a half-protected state.
   - Password reads from stdin (no shell history exposure); identity changes additionally require typing `BREAK-GLASS` to proceed.
   ```
3. Ensure the existing v1.3.1 admin user lifecycle "Added" / "Security" / "Documentation" subsections stay below.

Each commit can be its own targeted change; the implementer can group docs into one commit and the version bump into another.

Suggested commits for this task:

```bash
# 1) Version bump + lockfile refresh
git add pkg/version/version.go web/package.json web/package-lock.json CLAUDE.md
git commit -m "chore(version): 1.3.1 → 1.3.2"

# 2) Doc updates
git add docs/CHANGELOG.md docs/security/PROTECTED_USERS.md docs/security/PROTECTED_USERS_RECOVERY.md docs/security/THREAT_MODEL.md docs/USER_PERMISSIONS.md docs/TODO.md secrets/reset-password.sh
git commit -m "docs: v1.3.2 — break-glass CLI + collapse v1.3.1 entry into v1.3.2"
```

---

## Task 6: Smoke test against MariaDB + push + retitle PR

**Files:** none changed.

- [ ] **Step 1: Pre-flight cleanup (if container from earlier sessions still running)**

```bash
docker ps -q --filter "name=actalog-test-mariadb" | xargs -r docker stop
docker ps -q --filter "name=actalog-test-mariadb" | xargs -r docker rm
```

- [ ] **Step 2: Rebuild docker image with v1.3.2 binary**

```bash
cd /home/jcz/Github/actionlog/.worktrees/feature/admin-user-lifecycle-v1.3.1
./docker/scripts/build.sh dev
```

- [ ] **Step 3: Start container against MariaDB**

```bash
set -a; source /home/jcz/Github/actionlog/secrets/local-test-credentials.env; set +a
docker run -d --name actalog-test-mariadb --network host \
  -e DB_DRIVER=mysql -e DB_HOST="$MARIADB_HOST" -e DB_PORT="$MARIADB_PORT" \
  -e DB_USER="$MARIADB_USER" -e DB_PASSWORD="$MARIADB_PASSWORD" -e DB_NAME="$MARIADB_DB" \
  -e JWT_SECRET="$LOCAL_JWT_SECRET" \
  ghcr.io/johnzastrow/actalog:dev
sleep 4
curl -s -o /dev/null -w "health: %{http_code}\n" http://localhost:8080/health
```

- [ ] **Step 4: Exercise the CLI from inside the container**

```bash
# Password reset via CLI (use printf to feed stdin)
printf 'NewBreakGlass123\nNewBreakGlass123\n' | docker exec -i actalog-test-mariadb /app/actalog admin force-edit-protected --email br8kwall@gmail.com --field password --confirm

# Verify by logging in
curl -s -X POST http://localhost:8080/api/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"br8kwall@gmail.com","password":"NewBreakGlass123"}' \
  -w "\nHTTP: %{http_code}\n" | tail -1

# Identity field change (role)
printf 'BREAK-GLASS\n' | docker exec -i actalog-test-mariadb /app/actalog admin force-edit-protected --email br8kwall@gmail.com --field role --value athlete --confirm

# Verify role change
mysql -h "$MARIADB_HOST" -u "$MARIADB_USER" -p"$MARIADB_PASSWORD" "$MARIADB_DB" -e "SELECT email, role FROM users WHERE email='br8kwall@gmail.com'"

# Verify boot invariant still passes
docker exec actalog-test-mariadb /app/actalog admin verify-protected-users --verbose

# Audit events visible
mysql -h "$MARIADB_HOST" -u "$MARIADB_USER" -p"$MARIADB_PASSWORD" "$MARIADB_DB" -e "SELECT event_type, COUNT(*) FROM audit_logs WHERE event_type LIKE 'protected_user_break_glass_%' GROUP BY event_type"

# Restore role to admin to leave system in known good state
printf 'BREAK-GLASS\n' | docker exec -i actalog-test-mariadb /app/actalog admin force-edit-protected --email br8kwall@gmail.com --field role --value admin --confirm
```

- [ ] **Step 5: Tear down + push**

```bash
docker stop actalog-test-mariadb && docker rm actalog-test-mariadb
git push -f origin feature/admin-user-lifecycle-v1.3.1
```

- [ ] **Step 6: Retitle PR #222 to v1.3.2**

```bash
gh pr edit 222 --title "v1.3.2: admin user lifecycle + protected-user break-glass CLI"
```

Update the PR body via `gh pr edit 222 --body "$(cat <<'EOF' ... )"` to mention both feature areas. Reuse the v1.3.1 body and prepend a section on break-glass.

---

## End of plan

After PR #222 merges:
- Tag `v1.3.2` (annotated)
- Push docker images: `1.3.2`, `dev`, `latest`
- Delete the feature branch + worktree
- Per `feedback_release_cadence` memory: still no GitHub release page — the next batched release page covers v1.3.2.
