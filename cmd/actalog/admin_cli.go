package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"

	"github.com/johnzastrow/actalog/internal/repository"
)

// AdminVerifyProtectedUsers prints a per-check report to out and returns a
// shell-style exit code: 0 = all HARD checks pass, 1 = any HARD failure.
// verbose=true also prints soft warnings.
// This function is read-only and safe to run in any environment.
func AdminVerifyProtectedUsers(db *sql.DB, driver string, verbose bool, out io.Writer) int {
	report, err := VerifyProtectedUserInvariant(db, driver)
	pass := err == nil

	fmt.Fprintln(out, "Protected-user invariant:")
	fmt.Fprintf(out, "  Check 1/3 — L3 triggers exist        %s\n", checkmark(report != nil && report.Check1TriggersExist))
	fmt.Fprintf(out, "  Check 2/3 — triggers fire correctly  %s\n", checkmark(report != nil && report.Check2TriggersFire))
	fmt.Fprintf(out, "  Check 3/3 — protected rows exist     %s\n", softCheckmark(report))

	if !pass {
		fmt.Fprintf(out, "\nFailure: %v\n", err)
		fmt.Fprintln(out, "\nRecovery: ./bin/actalog admin reapply-protected-migrations --confirm")
		return 1
	}

	if verbose && report != nil {
		for _, w := range report.SoftWarnings {
			fmt.Fprintf(out, "  warn: %s\n", w)
		}
	}

	fmt.Fprintln(out, "\n✓ all checks passed")
	return 0
}

// AdminReapplyProtectedMigrations drops the L3 triggers and recreates them
// from the current source's per-dialect SQL constant. The operation is:
//   - Idempotent (safe to run repeatedly)
//   - Wrapped in a transaction so a partial reapply never leaves the DB half-applied
//   - Guarded by confirm=true; returns an error immediately if confirm=false
//
// After reapplying, re-runs VerifyProtectedUserInvariant and returns an error
// if the post-reapply invariant still fails. On success, writes a confirmation
// line to out.
func AdminReapplyProtectedMigrations(db *sql.DB, driver string, confirm bool, out io.Writer) error {
	if !confirm {
		return errors.New("refusing to reapply without --confirm")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Drop any existing protected-user triggers first (idempotent).
	var drops []string
	switch driver {
	case "sqlite3", "mysql":
		drops = []string{
			"DROP TRIGGER IF EXISTS protected_users_no_update",
			"DROP TRIGGER IF EXISTS protected_users_no_delete",
		}
	case "postgres":
		drops = []string{
			"DROP TRIGGER IF EXISTS protected_users_no_update ON users",
			"DROP TRIGGER IF EXISTS protected_users_no_delete ON users",
			"DROP FUNCTION IF EXISTS block_protected_users()",
		}
	default:
		return fmt.Errorf("unknown driver %q", driver)
	}

	for _, s := range drops {
		if _, err := tx.Exec(s); err != nil {
			return fmt.Errorf("drop existing triggers: %w", err)
		}
	}

	// Re-execute the per-dialect trigger SQL via the splitter.
	var sqlConst string
	switch driver {
	case "sqlite3":
		sqlConst = repository.SQLiteProtectedTriggers
	case "postgres":
		sqlConst = repository.PostgresProtectedTriggers
	case "mysql":
		sqlConst = repository.MySQLProtectedTriggers
	default:
		return fmt.Errorf("unknown driver %q", driver)
	}

	for _, stmt := range repository.SplitProtectedTriggerSQL(sqlConst) {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("apply protected-trigger SQL %q: %w", truncateForLog(stmt, 80), err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// Re-verify the invariant against the live DB (outside the transaction).
	if _, err := VerifyProtectedUserInvariant(db, driver); err != nil {
		return fmt.Errorf("triggers reapplied successfully, but invariant still reports failure (likely a Check 3 issue — protected user row missing or other non-trigger problem): %w", err)
	}

	fmt.Fprintln(out, "✓ protected triggers reapplied; invariant now passes")
	return nil
}

// checkmark returns "✓" when ok, "✗" otherwise.
func checkmark(ok bool) string {
	if ok {
		return "✓"
	}
	return "✗"
}

// softCheckmark returns "✓", "⚠ (soft, fresh install)", or "✗" based on the
// report's Check3 result and whether any soft warnings are present.
func softCheckmark(r *InvariantReport) string {
	if r == nil {
		return "✗"
	}
	if r.Check3ProtectedRowsExist {
		return "✓"
	}
	if len(r.SoftWarnings) > 0 {
		return "⚠ (soft, fresh install)"
	}
	return "✗"
}

// truncateForLog returns s if its byte length is <= max; otherwise returns
// the first max BYTES followed by "...". Used for log-truncation of SQL
// statements in error messages. Multi-byte runes split at the boundary
// will produce mojibake but won't panic — acceptable for ASCII SQL.
func truncateForLog(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
