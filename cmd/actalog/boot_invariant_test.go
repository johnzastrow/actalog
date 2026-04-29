package main

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/johnzastrow/actalog/internal/repository"
)

// openMigratedTestDB opens an in-memory SQLite3 database, runs all migrations
// (including 0.35.0 which installs the protected-user triggers), and registers
// a cleanup to close the DB when the test ends.
func openMigratedTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := repository.InitDatabase("sqlite3", ":memory:", nil)
	if err != nil {
		t.Fatalf("openMigratedTestDB: InitDatabase: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// mustExecTest executes a SQL statement against db, failing the test on error.
func mustExecTest(t *testing.T, db *sql.DB, query string, args ...interface{}) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("mustExecTest(%q): %v", query, err)
	}
}

// TestBootInvariant_FreshInstallNoUsersYet verifies that on a fresh install with
// zero users, Check 3 is a soft WARN and VerifyProtectedUserInvariant returns nil.
func TestBootInvariant_FreshInstallNoUsersYet(t *testing.T) {
	db := openMigratedTestDB(t)

	report, err := VerifyProtectedUserInvariant(db, "sqlite3")
	if err != nil {
		t.Fatalf("expected no error on fresh install, got: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if !report.Check1TriggersExist {
		t.Error("Check1TriggersExist should be true after migration 0.35.0")
	}
	if !report.Check2TriggersFire {
		t.Error("Check2TriggersFire should be true — triggers should reject the probe UPDATE")
	}
	if report.Check3ProtectedRowsExist {
		t.Error("Check3ProtectedRowsExist should be false on fresh install with no users")
	}
	if len(report.SoftWarnings) == 0 {
		t.Error("expected at least one SoftWarning for missing protected user on fresh install")
	}
}

// TestBootInvariant_HardFailWhenTriggerMissing verifies that if either required
// trigger is absent from the DB catalog, VerifyProtectedUserInvariant returns a
// non-nil error that names the missing trigger.
func TestBootInvariant_HardFailWhenTriggerMissing(t *testing.T) {
	db := openMigratedTestDB(t)

	// Drop both triggers to simulate a corrupt / partial migration.
	mustExecTest(t, db, `DROP TRIGGER IF EXISTS protected_users_no_update`)
	mustExecTest(t, db, `DROP TRIGGER IF EXISTS protected_users_no_delete`)

	_, err := VerifyProtectedUserInvariant(db, "sqlite3")
	if err == nil {
		t.Fatal("expected hard error when triggers are missing, got nil")
	}
	// The error must name which trigger is missing.
	if !strings.Contains(err.Error(), "protected_users_no_update") &&
		!strings.Contains(err.Error(), "protected_users_no_delete") {
		t.Errorf("error message does not name the missing trigger: %v", err)
	}
}

// TestBootInvariant_HardFailWhenProtectedRowMissingAndOtherUsersExist verifies
// that when at least one non-protected user exists but the protected user row is
// absent, VerifyProtectedUserInvariant returns a hard (non-nil) error.
func TestBootInvariant_HardFailWhenProtectedRowMissingAndOtherUsersExist(t *testing.T) {
	db := openMigratedTestDB(t)

	// Insert a non-protected user — the protected user is NOT inserted.
	mustExecTest(t, db,
		`INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		 VALUES ('alice@example.com', 'hash', 'Alice', 'athlete', datetime('now'), datetime('now'))`)

	_, err := VerifyProtectedUserInvariant(db, "sqlite3")
	if err == nil {
		t.Fatal("expected hard error when protected user missing and other users exist, got nil")
	}
}

// TestBootInvariant_PassesWithProtectedRowAndTriggers verifies the happy path:
// with the protected user row present and all triggers installed, all three
// checks pass and VerifyProtectedUserInvariant returns nil error.
func TestBootInvariant_PassesWithProtectedRowAndTriggers(t *testing.T) {
	db := openMigratedTestDB(t)

	// Insert the protected user row.
	mustExecTest(t, db,
		`INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		 VALUES ('br8kwall@gmail.com', 'hash', 'Protected', 'admin', datetime('now'), datetime('now'))`)

	report, err := VerifyProtectedUserInvariant(db, "sqlite3")
	if err != nil {
		t.Fatalf("expected no error on fully-provisioned DB, got: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if !report.Check1TriggersExist {
		t.Error("Check1TriggersExist should be true")
	}
	if !report.Check2TriggersFire {
		t.Error("Check2TriggersFire should be true")
	}
	if !report.Check3ProtectedRowsExist {
		t.Error("Check3ProtectedRowsExist should be true")
	}
	if len(report.SoftWarnings) != 0 {
		t.Errorf("expected no SoftWarnings on fully-provisioned DB, got: %v", report.SoftWarnings)
	}
}

// TestBootInvariant_Check2_HardFailWhenTriggerRaisesWrongMessage verifies that
// if the trigger exists but raises a different RAISE message (contract drift),
// VerifyProtectedUserInvariant returns a hard failure mentioning the expected
// contract substring.
func TestBootInvariant_Check2_HardFailWhenTriggerRaisesWrongMessage(t *testing.T) {
	db := openMigratedTestDB(t)

	// Replace the existing trigger with one that raises the wrong message.
	mustExecTest(t, db, `DROP TRIGGER protected_users_no_update`)
	mustExecTest(t, db, `
		CREATE TRIGGER protected_users_no_update
		BEFORE UPDATE ON users
		FOR EACH ROW
		WHEN OLD.email IN ('br8kwall@gmail.com')
		BEGIN
			SELECT RAISE(ABORT, 'wrong message');
		END;
	`)
	// Insert the protected user so Check 3 doesn't interfere.
	mustExecTest(t, db,
		`INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		 VALUES ('br8kwall@gmail.com', 'x', 'Owner', 'admin', datetime('now'), datetime('now'))`)

	_, err := VerifyProtectedUserInvariant(db, "sqlite3")
	if err == nil {
		t.Fatal("expected hard failure when trigger raises wrong message")
	}
	if !strings.Contains(err.Error(), "protected user: writes blocked at db layer") {
		t.Errorf("error must mention contract substring: %v", err)
	}
}
