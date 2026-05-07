package protectedusers

// breakglass.go — operator escape-hatch CLI for editing protected users.
//
// Why this exists: when self-service flows aren't viable (forgotten password +
// email gateway down) and the admin UI is locked out for protected users by
// design (L1+L2), the operator needs a documented, audited path to recover.
// This is that path. It is intended to be invoked manually by someone with
// shell access to the application host; there is no JWT, no admin role check,
// no rate limit. The audit log is the only forensic trail.
//
// Field allowlist: password | email | name | role | account_disabled.
// All five are user-visible identity/credential fields; per-field audit events
// (rather than one event with a discriminator) enable clean alert routing.
//
// Trigger interaction:
//   - password: lifecycle field. L3 already lets it through; no trigger fiddle.
//   - email | name | role | account_disabled: identity fields blocked by L3.
//     The CLI drops both protected_users_no_update and protected_users_no_delete
//     triggers, performs the UPDATE, then unconditionally reinstalls via
//     AdminReapplyProtectedMigrations. Reinstall failure panics with a recovery
//     message — a partial trigger state is worse than a hard exit.
//
// Operator metadata (USER, hostname, tty, cwd) is captured for audit details so
// post-incident review can answer who-when-where.

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
	Email   string    // target user's email — must be on the protected list
	Field   string    // password|email|name|role|account_disabled
	Value   string    // empty for password (read from Stdin); required otherwise
	Stdin   io.Reader // for password reads + identity confirmation; defaults to os.Stdin
	Stdout  io.Writer // operator UX output; defaults to os.Stderr
	Confirm bool      // --confirm flag must be present

	// IdentityConfirmation, if non-empty, is taken as the operator's response to
	// the "Type 'BREAK-GLASS' to proceed" gate. Used by tests to bypass stdin.
	IdentityConfirmation string
}

// AdminForceEditProtected is the entry point invoked by the CLI subcommand.
//
// Flow:
//  1. Validate flags + that target IS on the protected list (non-protected
//     targets must use the regular admin path).
//  2. Resolve target user from the DB.
//  3. Capture operator metadata.
//  4. Dispatch by field — password takes the lifecycle path; identity fields
//     drop+UPDATE+reinstall L3 triggers.
//  5. Re-run the boot invariant (3/3 hard checks). If it fails, abort with
//     a recovery message; the audit event was written before the invariant
//     check so the change is documented even if the system is now degraded.
//  6. Write the per-field audit event with operator metadata.
func AdminForceEditProtected(db *sql.DB, driver string, opts BreakGlassOptions) error {
	if opts.Stdout == nil {
		opts.Stdout = os.Stderr
	}

	// 1. Flag validation.
	if !opts.Confirm {
		return errors.New("--confirm flag is required")
	}
	target := strings.ToLower(strings.TrimSpace(opts.Email))
	if target == "" {
		return errors.New("--email is required")
	}
	if !security.IsProtectedEmail(target) {
		return fmt.Errorf("target %q is not on the protected list — use the normal admin path", target)
	}

	// 2. Resolve target user.
	userRepo := repository.NewSQLiteUserRepository(db)
	user, err := userRepo.GetByEmail(target)
	if err != nil {
		return fmt.Errorf("look up target: %w", err)
	}
	if user == nil {
		return fmt.Errorf("target %q not found in users table", target)
	}

	// 3. Operator metadata.
	meta := captureOperatorMetadata()
	auditDetails := map[string]interface{}{
		"operator_user":     meta.user,
		"operator_hostname": meta.hostname,
		"operator_tty":      meta.tty,
		"operator_cwd":      meta.cwd,
	}

	// 4. Dispatch.
	var eventType string
	switch opts.Field {
	case "password":
		eventType = domain.EventProtectedUserBreakGlassPassword
		if err := handlePasswordField(db, user, opts, auditDetails); err != nil {
			return err
		}
	case "email":
		eventType = domain.EventProtectedUserBreakGlassEmail
		if err := handleIdentityField(db, driver, user, "email", opts, auditDetails, meta.user); err != nil {
			return err
		}
	case "name":
		eventType = domain.EventProtectedUserBreakGlassName
		if err := handleIdentityField(db, driver, user, "name", opts, auditDetails, meta.user); err != nil {
			return err
		}
	case "role":
		eventType = domain.EventProtectedUserBreakGlassRole
		if err := handleIdentityField(db, driver, user, "role", opts, auditDetails, meta.user); err != nil {
			return err
		}
	case "account_disabled":
		eventType = domain.EventProtectedUserBreakGlassAccountDisabled
		if err := handleIdentityField(db, driver, user, "account_disabled", opts, auditDetails, meta.user); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown field %q (allowed: password|email|name|role|account_disabled)", opts.Field)
	}

	// 5. Re-run boot invariant.
	if _, invErr := VerifyProtectedUserInvariant(db, driver); invErr != nil {
		return fmt.Errorf("post-edit invariant FAILED: %w  — protected-user defense may be DEGRADED; run: ./bin/actalog admin reapply-protected-migrations --confirm", invErr)
	}

	// 6. Audit. Build the minimal audit graph inline (same pattern
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
// password_hash is a lifecycle field; the L3 trigger already lets it through.
func handlePasswordField(db *sql.DB, user *domain.User, opts BreakGlassOptions, details map[string]interface{}) error {
	if opts.Value != "" {
		return errors.New("password must be read from stdin, not --value (avoids shell-history exposure)")
	}
	stdin := opts.Stdin
	if stdin == nil {
		stdin = os.Stdin
	}
	pw, err := readPasswordTwice(stdin, opts.Stdout)
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
	revokeOK := true
	if err := tokenRepo.RevokeAllForUser(user.ID); err != nil {
		fmt.Fprintf(opts.Stdout, "warn: RevokeAllForUser: %v\n", err)
		revokeOK = false
	}

	details["cleared_failed_login_attempts"] = priorAttempts
	details["cleared_lockout"] = priorLocked
	details["revoked_refresh_tokens"] = revokeOK
	details["trigger_dropped"] = false
	return nil
}

// handleIdentityField runs the identity-field branch. Drops L3 triggers,
// performs the column-specific UPDATE, then reinstalls triggers. Reinstall
// failure is FATAL (panic with recovery message).
func handleIdentityField(db *sql.DB, driver string, user *domain.User, field string, opts BreakGlassOptions, details map[string]interface{}, opUser string) error {
	if err := requireIdentityConfirmation(opts); err != nil {
		return err
	}
	if opts.Value == "" {
		return fmt.Errorf("--value is required for field %q", field)
	}

	// Per-field validation + old/new value resolution.
	var oldVal, newVal string
	var newAcctDisabled bool
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
		newAcctDisabled = boolVal
	default:
		return fmt.Errorf("unsupported identity field %q", field)
	}

	// Drop both triggers so the UPDATE can land.
	if err := dropProtectedTriggers(db, driver); err != nil {
		return fmt.Errorf("drop triggers: %w", err)
	}

	// Perform the UPDATE.
	updateErr := executeIdentityUpdate(db, driver, user.ID, field, newVal, newAcctDisabled, opUser)

	// Reinstall triggers UNCONDITIONALLY. Reinstall failure is fatal — the
	// system is now in an unprotected state and only the operator can resolve.
	var reinstallOut strings.Builder
	if reinstallErr := AdminReapplyProtectedMigrations(db, driver, true, &reinstallOut); reinstallErr != nil {
		panic(fmt.Sprintf(
			"FATAL: trigger reinstall failed after edit. Run NOW: ./bin/actalog admin reapply-protected-migrations --confirm.\n"+
				"Original UPDATE err: %v\nReinstall err: %v\nReinstall output:\n%s",
			updateErr, reinstallErr, reinstallOut.String(),
		))
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
// Uses ?-style placeholders; postgres doesn't accept those so we do a small
// inline rebind for the postgres path.
func executeIdentityUpdate(db *sql.DB, driver string, userID int64, field string, newVal string, newAcctDisabled bool, operator string) error {
	now := time.Now().UTC()

	var query string
	var args []interface{}

	switch field {
	case "email", "name", "role":
		query = fmt.Sprintf("UPDATE users SET %s = ?, updated_at = ? WHERE id = ?", field)
		args = []interface{}{newVal, now, userID}
	case "account_disabled":
		var disabledAt interface{}
		var disableReason interface{}
		if newAcctDisabled {
			disabledAt = now
			disableReason = "break-glass: " + operator
		} else {
			disabledAt = nil
			disableReason = nil
		}
		query = "UPDATE users SET account_disabled = ?, disabled_at = ?, disabled_by_user_id = NULL, disable_reason = ?, updated_at = ? WHERE id = ?"
		args = []interface{}{newAcctDisabled, disabledAt, disableReason, now, userID}
	default:
		return fmt.Errorf("executeIdentityUpdate: unsupported field %q", field)
	}

	if driver == "postgres" {
		query = rebindToPostgres(query)
	}

	_, err := db.Exec(query, args...)
	return err
}

// rebindToPostgres converts ? placeholders to $1, $2, ... for postgres.
// Inline implementation rather than depending on repository.RebindQuery
// (which reads a package-level currentDriver variable set at InitDatabase
// time — not thread-safe to depend on).
func rebindToPostgres(query string) string {
	var b strings.Builder
	b.Grow(len(query) + 8)
	n := 1
	for _, c := range query {
		if c == '?' {
			fmt.Fprintf(&b, "$%d", n)
			n++
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}

// dropProtectedTriggers drops both protected_users_no_update and
// protected_users_no_delete using per-driver syntax.
//
// Postgres requires "DROP TRIGGER IF EXISTS name ON tablename"; sqlite3 and
// mysql accept just "DROP TRIGGER IF EXISTS name". Same per-driver split as
// scripts/recover/lockstep_test.go encodes.
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

// readPasswordTwice reads a password from in twice and verifies they match.
// In tests, in is a bytes.Buffer with two newline-separated lines; in normal
// use it is os.Stdin (the CLI wrapper is responsible for terminal echo-off if
// desired — kept simple here for testability).
func readPasswordTwice(in io.Reader, out io.Writer) (string, error) {
	if out != nil {
		fmt.Fprint(out, "New password: ")
	}
	pw, err := readLine(in)
	if err != nil {
		return "", fmt.Errorf("read new password: %w", err)
	}
	if out != nil {
		fmt.Fprintln(out)
		fmt.Fprint(out, "Confirm:      ")
	}
	confirm, err := readLine(in)
	if err != nil {
		return "", fmt.Errorf("read confirm: %w", err)
	}
	if out != nil {
		fmt.Fprintln(out)
	}
	if pw != confirm {
		return "", errors.New("passwords don't match")
	}
	return pw, nil
}

// readLine reads up to a newline or EOF.
func readLine(in io.Reader) (string, error) {
	var sb strings.Builder
	buf := make([]byte, 1)
	for {
		n, err := in.Read(buf)
		if n > 0 {
			if buf[0] == '\n' {
				return sb.String(), nil
			}
			if buf[0] != '\r' {
				sb.WriteByte(buf[0])
			}
		}
		if err == io.EOF {
			if sb.Len() == 0 {
				return "", err
			}
			return sb.String(), nil
		}
		if err != nil {
			return "", err
		}
	}
}

// requireIdentityConfirmation enforces the "type BREAK-GLASS to proceed" gate.
// If opts.IdentityConfirmation is non-empty, that value is used directly (test
// path); otherwise reads a line from opts.Stdin.
func requireIdentityConfirmation(opts BreakGlassOptions) error {
	const required = "BREAK-GLASS"

	var typed string
	if opts.IdentityConfirmation != "" {
		typed = strings.TrimSpace(opts.IdentityConfirmation)
	} else {
		stdin := opts.Stdin
		if stdin == nil {
			return errors.New("identity confirmation required but no stdin provided")
		}
		if opts.Stdout != nil {
			fmt.Fprint(opts.Stdout, "Type 'BREAK-GLASS' to proceed: ")
		}
		line, err := readLine(stdin)
		if err != nil {
			return fmt.Errorf("confirmation read: %w", err)
		}
		typed = strings.TrimSpace(line)
	}

	if typed != required {
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
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("password must include uppercase, lowercase, and a digit")
	}
	return nil
}

// parseBool accepts true|false|1|0|yes|no|y|n (case-insensitive).
func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes", "y":
		return true, nil
	case "false", "0", "no", "n":
		return false, nil
	}
	return false, fmt.Errorf("could not parse %q as bool (use true|false|1|0|yes|no)", s)
}

type operatorMetadata struct {
	user, hostname, tty, cwd string
}

// captureOperatorMetadata reads ambient env values used to attribute the
// break-glass operation in the audit log. Best-effort — missing fields are
// recorded as empty strings rather than failing the operation.
func captureOperatorMetadata() operatorMetadata {
	m := operatorMetadata{}
	m.user = os.Getenv("USER")
	if m.user == "" {
		m.user = os.Getenv("LOGNAME")
	}
	if h, err := os.Hostname(); err == nil {
		m.hostname = h
	}
	if cwd, err := os.Getwd(); err == nil {
		m.cwd = cwd
	}
	if tty := os.Getenv("SSH_TTY"); tty != "" {
		m.tty = tty
	}
	return m
}
