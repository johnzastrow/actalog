package main

import (
	"strings"
	"testing"
)

// TestAdminVerifyCLI_PassesOnHealthyDB verifies that AdminVerifyProtectedUsers
// prints a full 3/3 pass report and returns exit code 0 when the DB is healthy.
func TestAdminVerifyCLI_PassesOnHealthyDB(t *testing.T) {
	db := openMigratedTestDB(t)

	// Insert the protected user so Check 3 passes too.
	mustExecTest(t, db,
		`INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		 VALUES ('br8kwall@gmail.com', 'hash', 'Protected', 'admin', datetime('now'), datetime('now'))`)

	var buf strings.Builder
	code := AdminVerifyProtectedUsers(db, "sqlite3", true, &buf)

	if code != 0 {
		t.Errorf("expected exit code 0 on healthy DB, got %d; output:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Check 1/3") {
		t.Errorf("output missing 'Check 1/3': %s", out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("output missing '✓': %s", out)
	}
}

// TestAdminVerifyCLI_FailsAndNamesCheck verifies that dropping a trigger causes
// AdminVerifyProtectedUsers to return exit code 1 and name the missing trigger.
func TestAdminVerifyCLI_FailsAndNamesCheck(t *testing.T) {
	db := openMigratedTestDB(t)

	// Drop one trigger to simulate corruption.
	mustExecTest(t, db, `DROP TRIGGER IF EXISTS protected_users_no_update`)

	var buf strings.Builder
	code := AdminVerifyProtectedUsers(db, "sqlite3", false, &buf)

	if code != 1 {
		t.Errorf("expected exit code 1 when trigger missing, got %d; output:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "protected_users_no_update") {
		t.Errorf("output must name the missing trigger; got:\n%s", out)
	}
}

// TestAdminReapplyCLI_RestoresDroppedTrigger verifies that AdminReapplyProtectedMigrations
// drops and recreates the triggers and that a subsequent invariant check passes —
// and that a real UPDATE on the protected email is rejected.
func TestAdminReapplyCLI_RestoresDroppedTrigger(t *testing.T) {
	db := openMigratedTestDB(t)

	// Insert the protected user row so Check 3 doesn't produce a hard error.
	mustExecTest(t, db,
		`INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		 VALUES ('br8kwall@gmail.com', 'hash', 'Protected', 'admin', datetime('now'), datetime('now'))`)

	// Simulate trigger loss.
	mustExecTest(t, db, `DROP TRIGGER IF EXISTS protected_users_no_update`)
	mustExecTest(t, db, `DROP TRIGGER IF EXISTS protected_users_no_delete`)

	// Verify that invariant now fails.
	_, errBefore := VerifyProtectedUserInvariant(db, "sqlite3")
	if errBefore == nil {
		t.Fatal("expected invariant to fail after dropping triggers")
	}

	var buf strings.Builder
	if err := AdminReapplyProtectedMigrations(db, "sqlite3", true, &buf); err != nil {
		t.Fatalf("AdminReapplyProtectedMigrations returned error: %v; output:\n%s", err, buf.String())
	}

	// Post-reapply invariant must pass.
	if _, err := VerifyProtectedUserInvariant(db, "sqlite3"); err != nil {
		t.Fatalf("invariant still failing after reapply: %v", err)
	}

	// Confirm that a real UPDATE on the protected row is actually rejected.
	_, updateErr := db.Exec(
		`UPDATE users SET name = '__should_be_blocked__' WHERE email = 'br8kwall@gmail.com'`)
	if updateErr == nil {
		t.Error("expected UPDATE on protected row to be rejected by trigger after reapply, but it succeeded")
	}
	if !strings.Contains(updateErr.Error(), "protected user: writes blocked at db layer") {
		t.Errorf("UPDATE rejected but missing contract message: %v", updateErr)
	}
}

// TestAdminReapplyCLI_RequiresConfirm verifies that without --confirm the function
// returns a non-nil error and makes no changes.
func TestAdminReapplyCLI_RequiresConfirm(t *testing.T) {
	db := openMigratedTestDB(t)

	var buf strings.Builder
	err := AdminReapplyProtectedMigrations(db, "sqlite3", false, &buf)
	if err == nil {
		t.Fatal("expected error when confirm=false, got nil")
	}
	if !strings.Contains(err.Error(), "--confirm") {
		t.Errorf("error message should mention --confirm; got: %v", err)
	}
	// The triggers must still be intact (we made no changes).
	if _, verifyErr := VerifyProtectedUserInvariant(db, "sqlite3"); verifyErr != nil {
		t.Errorf("triggers should be untouched when confirm=false, but invariant failed: %v", verifyErr)
	}
}
