package main

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/johnzastrow/actalog/pkg/security"
)

// InvariantReport summarises the three boot-time checks.
type InvariantReport struct {
	Check1TriggersExist      bool
	Check2TriggersFire       bool
	Check3ProtectedRowsExist bool
	SoftWarnings             []string
}

// emailPlaceholder returns the per-driver placeholder for a single email
// parameter ("?" for sqlite3 and mysql, "$1" for postgres).
// Postgres rejects "?" literally — must be "$1".
func emailPlaceholder(driver string) string {
	if driver == "postgres" {
		return "$1"
	}
	return "?"
}

// VerifyProtectedUserInvariant runs three boot-time checks against the DB.
//
// Check 1 (HARD): required triggers exist in the DB catalog.
// Check 2 (HARD): a simulated UPDATE inside BEGIN/ROLLBACK is rejected by the trigger.
// Check 3 (SOFT on fresh install / HARD otherwise): protected user row exists.
//
// On HARD failure it returns a non-nil error naming the specific failing check and
// the recovery command literal. On SOFT failure (fresh install with no users yet)
// it appends to report.SoftWarnings and returns nil.
func VerifyProtectedUserInvariant(db *sql.DB, driver string) (*InvariantReport, error) {
	r := &InvariantReport{}

	// Check 1: required triggers exist in the catalog.
	required := []string{"protected_users_no_update", "protected_users_no_delete"}
	for _, name := range required {
		exists, err := triggerExistsForDriver(db, driver, name)
		if err != nil {
			return r, fmt.Errorf("invariant: catalog query for %q: %w", name, err)
		}
		if !exists {
			return r, fmt.Errorf(
				"invariant: trigger %q missing — recover with: "+
					"./bin/actalog admin reapply-protected-migrations --confirm",
				name)
		}
	}
	r.Check1TriggersExist = true

	// Check 2: simulated UPDATE inside a rolled-back transaction must be rejected.
	// We test one protected email at a time; for current scale (1 entry) this is sufficient.
	for _, protectedEmail := range security.ProtectedEmailsList() {
		if err := simulateProtectedUpdate(db, driver, protectedEmail); err != nil {
			return r, fmt.Errorf(
				"invariant: simulated UPDATE for %q did not reject as expected (trigger broken or absent): %w. "+
					"Recover with: ./bin/actalog admin reapply-protected-migrations --confirm",
				protectedEmail, err)
		}
	}
	r.Check2TriggersFire = true

	// Check 3: protected user row exists. Soft on fresh install only.
	for _, protectedEmail := range security.ProtectedEmailsList() {
		var n int
		ph := emailPlaceholder(driver)
		if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = "+ph, protectedEmail).Scan(&n); err != nil {
			return r, fmt.Errorf("invariant: count protected: %w", err)
		}
		if n == 0 {
			var totalUsers int
			if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers); err != nil {
				return r, fmt.Errorf("invariant: count users: %w", err)
			}
			if totalUsers == 0 {
				r.SoftWarnings = append(r.SoftWarnings,
					fmt.Sprintf("protected user %q not yet present (expected on fresh install)", protectedEmail))
			} else {
				return r, fmt.Errorf(
					"invariant: protected user %q missing while %d other users exist — likely deletion. "+
						"Restore from backup or remove from protected list via the operator runbook "+
						"(docs/security/PROTECTED_USERS.md)",
					protectedEmail, totalUsers)
			}
		}
	}
	r.Check3ProtectedRowsExist = len(r.SoftWarnings) == 0
	return r, nil
}

// triggerExistsForDriver executes a per-dialect catalog query to check whether a
// trigger named `name` exists on the users table.
func triggerExistsForDriver(db *sql.DB, driver, name string) (bool, error) {
	var q string
	switch driver {
	case "sqlite3":
		q = `SELECT 1 FROM sqlite_master WHERE type='trigger' AND name=? AND tbl_name='users'`
	case "postgres":
		q = `SELECT 1 FROM pg_trigger WHERE tgname=$1 AND tgrelid = 'users'::regclass`
	case "mysql":
		q = `SELECT 1 FROM INFORMATION_SCHEMA.TRIGGERS WHERE trigger_name=? AND event_object_table='users'`
	default:
		return false, fmt.Errorf("unknown driver %q", driver)
	}
	var x int
	err := db.QueryRow(q, name).Scan(&x)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// simulateProtectedUpdate inserts a row with the protected email inside a
// transaction, then attempts to UPDATE it (changing only the name column).
// The BEFORE UPDATE trigger fires because OLD.email matches the protected
// address and should reject the UPDATE. The transaction is always rolled back
// so the database is left unchanged whether the trigger fires or not.
//
// Returns nil if the trigger correctly rejected with the contract substring.
// Returns a non-nil error if the trigger did NOT reject (broken or absent), the
// contract substring is missing from the error, or any unexpected DB error occurs.
func simulateProtectedUpdate(db *sql.DB, driver, protectedEmail string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback on commit never called; no-op on already-rolled-back tx

	nowExpr := "datetime('now')"
	if driver != "sqlite3" {
		nowExpr = "NOW()"
	}

	// Insert a row with the protected email. Use INSERT OR IGNORE / ON CONFLICT
	// so that if the protected user already exists (happy-path test), the
	// statement succeeds without violating the UNIQUE constraint.
	//
	// nowExpr is a compile-time-known database expression (datetime('now') or
	// NOW()), not user input. fmt.Sprintf is safe here; no injection risk.
	var insertSQL string
	switch driver {
	case "sqlite3":
		insertSQL = fmt.Sprintf(
			"INSERT OR IGNORE INTO users (email, password_hash, name, role, created_at, updated_at) "+
				"VALUES (?, 'x', 'probe', 'athlete', %s, %s)",
			nowExpr, nowExpr)
	case "postgres":
		insertSQL = fmt.Sprintf(
			"INSERT INTO users (email, password_hash, name, role, created_at, updated_at) "+
				"VALUES ($1, 'x', 'probe', 'athlete', %s, %s) ON CONFLICT (email) DO NOTHING",
			nowExpr, nowExpr)
	case "mysql":
		insertSQL = fmt.Sprintf(
			"INSERT IGNORE INTO users (email, password_hash, name, role, created_at, updated_at) "+
				"VALUES (?, 'x', 'probe', 'athlete', %s, %s)",
			nowExpr, nowExpr)
	default:
		return fmt.Errorf("unsupported driver %q", driver)
	}
	if _, err := tx.Exec(insertSQL, protectedEmail); err != nil {
		return fmt.Errorf("probe insert: %w", err)
	}

	// Now attempt to UPDATE the protected row. This fires the BEFORE UPDATE
	// trigger which should raise an error with the contract message.
	updateSQL := "UPDATE users SET name = '__probe__' WHERE email = ?"
	if driver == "postgres" {
		updateSQL = "UPDATE users SET name = '__probe__' WHERE email = $1"
	}
	_, err = tx.Exec(updateSQL, protectedEmail)
	if err == nil {
		return fmt.Errorf("UPDATE was not rejected (trigger absent or broken)")
	}
	if !strings.Contains(err.Error(), security.TriggerErrorContract) {
		return fmt.Errorf("UPDATE rejected, but error message missing contract %q: %v", security.TriggerErrorContract, err.Error())
	}
	return nil
}
