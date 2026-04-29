package integration

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/johnzastrow/actalog/internal/repository"
)

// mustOpenTestDB opens and fully migrates a test DB using the configured driver and DSN.
// It is registered with t.Cleanup to close when the test ends.
func mustOpenTestDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	db, err := repository.InitDatabase(*testDBDriver, *testDSN, nil)
	if err != nil {
		t.Fatalf("mustOpenTestDB: InitDatabase(%s, %s): %v", *testDBDriver, *testDSN, err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, *testDBDriver
}

// triggerExists reports whether a trigger with the given name exists in the database catalog.
// Per-dialect query:
//   - sqlite3: sqlite_master
//   - postgres: pg_trigger
//   - mysql:   INFORMATION_SCHEMA.TRIGGERS
func triggerExists(t *testing.T, db *sql.DB, driver, name string) bool {
	t.Helper()
	var query string
	var args []interface{}
	switch driver {
	case "sqlite3":
		query = `SELECT 1 FROM sqlite_master WHERE type='trigger' AND name=? AND tbl_name='users'`
		args = []interface{}{name}
	case "postgres":
		query = `SELECT 1 FROM pg_trigger WHERE tgname=$1 AND tgrelid = 'users'::regclass`
		args = []interface{}{name}
	case "mysql":
		query = `SELECT 1 FROM INFORMATION_SCHEMA.TRIGGERS WHERE TRIGGER_NAME=? AND event_object_table='users'`
		args = []interface{}{name}
	default:
		t.Fatalf("triggerExists: unsupported driver %q", driver)
	}

	var dummy int
	err := db.QueryRow(query, args...).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		t.Fatalf("triggerExists(%s): %v", name, err)
	}
	return true
}

// insertProtectedUser inserts a row with email br8kwall@gmail.com into the users table.
// Uses minimal required columns based on the baseline schema.
func insertProtectedUser(t *testing.T, db *sql.DB, driver string) {
	t.Helper()
	var q string
	switch driver {
	case "sqlite3":
		q = `INSERT OR IGNORE INTO users (email, password_hash, name, role, created_at, updated_at)
		     VALUES ('br8kwall@gmail.com', 'hash', 'Protected', 'admin', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	case "postgres":
		q = `INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		     VALUES ('br8kwall@gmail.com', 'hash', 'Protected', 'admin', NOW(), NOW())
		     ON CONFLICT (email) DO NOTHING`
	case "mysql":
		q = `INSERT IGNORE INTO users (email, password_hash, name, role, created_at, updated_at)
		     VALUES ('br8kwall@gmail.com', 'hash', 'Protected', 'admin', NOW(), NOW())`
	default:
		t.Fatalf("insertProtectedUser: unsupported driver %q", driver)
	}
	if _, err := db.Exec(q); err != nil {
		t.Fatalf("insertProtectedUser: %v", err)
	}
}

// insertNonProtectedUser inserts alice@example.com for tests that verify allowed writes.
func insertNonProtectedUser(t *testing.T, db *sql.DB, driver string) {
	t.Helper()
	var q string
	switch driver {
	case "sqlite3":
		q = `INSERT OR IGNORE INTO users (email, password_hash, name, role, created_at, updated_at)
		     VALUES ('alice@example.com', 'hash', 'Alice', 'user', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	case "postgres":
		q = `INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		     VALUES ('alice@example.com', 'hash', 'Alice', 'user', NOW(), NOW())
		     ON CONFLICT (email) DO NOTHING`
	case "mysql":
		q = `INSERT IGNORE INTO users (email, password_hash, name, role, created_at, updated_at)
		     VALUES ('alice@example.com', 'hash', 'Alice', 'user', NOW(), NOW())`
	default:
		t.Fatalf("insertNonProtectedUser: unsupported driver %q", driver)
	}
	if _, err := db.Exec(q); err != nil {
		t.Fatalf("insertNonProtectedUser: %v", err)
	}
}

// TestL3_TriggerExistsAfterMigration verifies that migration 0.35.0 installs both
// protected_users_no_update and protected_users_no_delete triggers in the DB catalog.
func TestL3_TriggerExistsAfterMigration(t *testing.T) {
	db, driver := mustOpenTestDB(t)

	for _, name := range []string{"protected_users_no_update", "protected_users_no_delete"} {
		if !triggerExists(t, db, driver, name) {
			t.Errorf("trigger %q not found in %s catalog after migration", name, driver)
		}
	}
}

// TestL3_UPDATEOnProtectedRowRejectedByDB verifies that the DB-level trigger
// rejects UPDATE on br8kwall@gmail.com with an error.
func TestL3_UPDATEOnProtectedRowRejectedByDB(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)

	_, err := db.Exec(`UPDATE users SET name='hacked' WHERE email='br8kwall@gmail.com'`)
	if err == nil {
		t.Fatal("expected UPDATE on protected user to fail, but it succeeded")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "protected user") {
		t.Errorf("error message does not contain 'protected user'; got: %v", err)
	}
}

// TestL3_DELETEOnProtectedRowRejectedByDB verifies that the DB-level trigger
// rejects DELETE on br8kwall@gmail.com with an error.
func TestL3_DELETEOnProtectedRowRejectedByDB(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)

	_, err := db.Exec(`DELETE FROM users WHERE email='br8kwall@gmail.com'`)
	if err == nil {
		t.Fatal("expected DELETE on protected user to fail, but it succeeded")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "protected user") {
		t.Errorf("error message does not contain 'protected user'; got: %v", err)
	}
}

// TestL3_UPDATEOnNonProtectedRowSucceeds verifies that the trigger does NOT
// block normal UPDATE operations on non-protected users.
func TestL3_UPDATEOnNonProtectedRowSucceeds(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertNonProtectedUser(t, db, driver)

	_, err := db.Exec(`UPDATE users SET name='Alice Updated' WHERE email='alice@example.com'`)
	if err != nil {
		t.Errorf("expected UPDATE on non-protected user to succeed, but got error: %v", err)
	}
}

// TestL3_TriggerErrorMessageMatchesContract verifies that the exact L4 contract
// substring "protected user: writes blocked at db layer" appears in the error
// returned by an attempted write on the protected user.
// This string is pattern-matched by the L4 service-layer error wrapper (Task 11).
func TestL3_TriggerErrorMessageMatchesContract(t *testing.T) {
	const contractMsg = "protected user: writes blocked at db layer"

	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)

	_, err := db.Exec(`UPDATE users SET name='hacked' WHERE email='br8kwall@gmail.com'`)
	if err == nil {
		t.Fatal("expected UPDATE on protected user to fail, but it succeeded")
	}
	if !strings.Contains(err.Error(), contractMsg) {
		t.Errorf("trigger error message does not match L4 contract.\nExpected substring: %q\nActual error:       %v", contractMsg, err)
	}
}

// TestL3_TriggerSurvivesRollback verifies that a transaction containing one
// allowed UPDATE and one blocked UPDATE rolls back cleanly, and that the trigger
// continues to function correctly after the rollback.
func TestL3_TriggerSurvivesRollback(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)
	insertNonProtectedUser(t, db, driver)

	// Begin a transaction: update alice (allowed), then attempt to update protected user (blocked).
	// The whole transaction should roll back.
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}

	// First statement: allowed update — should succeed within the transaction.
	if _, err := tx.Exec(`UPDATE users SET name='Alice Txn' WHERE email='alice@example.com'`); err != nil {
		_ = tx.Rollback()
		t.Fatalf("allowed UPDATE within tx failed unexpectedly: %v", err)
	}

	// Second statement: blocked update — should return an error.
	_, blockedErr := tx.Exec(`UPDATE users SET name='hacked' WHERE email='br8kwall@gmail.com'`)
	if blockedErr == nil {
		_ = tx.Rollback()
		t.Fatal("expected blocked UPDATE to fail inside tx, but it succeeded")
	}

	// Roll back the transaction.
	if rbErr := tx.Rollback(); rbErr != nil {
		// SQLite returns ErrTxDone after a statement error in some drivers; that is acceptable.
		if rbErr.Error() != "sql: transaction has already been committed or rolled back" {
			t.Fatalf("Rollback error: %v", rbErr)
		}
	}

	// After rollback: trigger must still block writes on the protected user.
	_, err = db.Exec(`UPDATE users SET name='hacked again' WHERE email='br8kwall@gmail.com'`)
	if err == nil {
		t.Error("trigger should still block UPDATE on protected user after rollback, but it succeeded")
	}

	// After rollback: alice's name must NOT have been changed (tx was rolled back).
	var aliceName string
	if qErr := db.QueryRow(`SELECT name FROM users WHERE email='alice@example.com'`).Scan(&aliceName); qErr != nil {
		t.Fatalf("SELECT alice name: %v", qErr)
	}
	if aliceName == "Alice Txn" {
		t.Error("alice's name was committed despite rollback — transaction did not roll back correctly")
	}
}
