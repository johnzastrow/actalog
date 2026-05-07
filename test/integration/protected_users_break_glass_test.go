package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/johnzastrow/actalog/internal/protectedusers"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/pkg/auth"
	"github.com/johnzastrow/actalog/pkg/security"
)

// TestBreakGlass_Password_EndToEnd verifies a password change via the CLI
// entry point: writes the new hash, validates against bcrypt, and the
// post-edit boot invariant still passes.
func TestBreakGlass_Password_EndToEnd(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)
	withTriggerCleanup(t, db, driver)

	stdin := bytes.NewBufferString("BreakGlassPass123\nBreakGlassPass123\n")
	stdout := &bytes.Buffer{}

	err := protectedusers.AdminForceEditProtected(db, driver, protectedusers.BreakGlassOptions{
		Email:   security.ProtectedEmailsList()[0],
		Field:   "password",
		Stdin:   stdin,
		Stdout:  stdout,
		Confirm: true,
	})
	if err != nil {
		t.Fatalf("AdminForceEditProtected password: %v", err)
	}

	repo := repository.NewSQLiteUserRepository(db)
	stored, err := repo.GetByEmail(security.ProtectedEmailsList()[0])
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if stored == nil {
		t.Fatal("protected user disappeared")
	}
	if err := auth.CheckPassword(stored.PasswordHash, "BreakGlassPass123"); err != nil {
		t.Errorf("new password should validate: %v", err)
	}
}

// TestBreakGlass_Role_DropsAndReinstallsTriggers verifies the identity-field
// path: triggers go away briefly, the UPDATE lands, triggers come back, and
// L3 still rejects subsequent identity changes.
func TestBreakGlass_Role_DropsAndReinstallsTriggers(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)
	withTriggerCleanup(t, db, driver)

	err := protectedusers.AdminForceEditProtected(db, driver, protectedusers.BreakGlassOptions{
		Email:                security.ProtectedEmailsList()[0],
		Field:                "role",
		Value:                "athlete",
		Stdout:               &bytes.Buffer{},
		Confirm:              true,
		IdentityConfirmation: "BREAK-GLASS",
	})
	if err != nil {
		t.Fatalf("AdminForceEditProtected role: %v", err)
	}

	repo := repository.NewSQLiteUserRepository(db)
	stored, _ := repo.GetByEmail(security.ProtectedEmailsList()[0])
	if stored.Role != "athlete" {
		t.Errorf("role = %q, want athlete", stored.Role)
	}

	// L3 must still block identity-field changes after the break-glass operation.
	_, attemptErr := db.Exec(`UPDATE users SET name = 'attack' WHERE email = ?`, security.ProtectedEmailsList()[0])
	if attemptErr == nil {
		t.Error("L3 trigger should still block identity-field changes after break-glass")
	}
	if attemptErr != nil && !strings.Contains(attemptErr.Error(), "protected user") {
		t.Errorf("expected protected-user contract error; got %v", attemptErr)
	}
}

// TestBreakGlass_AccountDisabled_TogglePersists verifies the account_disabled
// branch sets the column + disabled_at + disable_reason.
func TestBreakGlass_AccountDisabled_TogglePersists(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)
	withTriggerCleanup(t, db, driver)

	err := protectedusers.AdminForceEditProtected(db, driver, protectedusers.BreakGlassOptions{
		Email:                security.ProtectedEmailsList()[0],
		Field:                "account_disabled",
		Value:                "true",
		Stdout:               &bytes.Buffer{},
		Confirm:              true,
		IdentityConfirmation: "BREAK-GLASS",
	})
	if err != nil {
		t.Fatalf("AdminForceEditProtected account_disabled: %v", err)
	}

	repo := repository.NewSQLiteUserRepository(db)
	stored, _ := repo.GetByEmail(security.ProtectedEmailsList()[0])
	if !stored.AccountDisabled {
		t.Error("AccountDisabled should be true")
	}
	if stored.DisabledAt == nil {
		t.Error("DisabledAt should be set when account_disabled flips to true")
	}
}

// TestBreakGlass_Email_RejectsProtectedTarget verifies that migrating a
// protected account to another protected email is rejected before any DB
// write or trigger drop happens.
func TestBreakGlass_Email_RejectsProtectedTarget(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)
	withTriggerCleanup(t, db, driver)

	err := protectedusers.AdminForceEditProtected(db, driver, protectedusers.BreakGlassOptions{
		Email:                security.ProtectedEmailsList()[0],
		Field:                "email",
		Value:                security.ProtectedEmailsList()[0], // same protected address
		Stdout:               &bytes.Buffer{},
		Confirm:              true,
		IdentityConfirmation: "BREAK-GLASS",
	})
	if err == nil {
		t.Fatal("expected rejection of protected-list migration")
	}
	if !strings.Contains(err.Error(), "protected list") {
		t.Errorf("expected protected-list error; got %v", err)
	}
}

// TestBreakGlass_AuditEventFires verifies an audit row is written for each
// successful break-glass operation, with the per-field event type.
func TestBreakGlass_AuditEventFires(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)
	withTriggerCleanup(t, db, driver)

	err := protectedusers.AdminForceEditProtected(db, driver, protectedusers.BreakGlassOptions{
		Email:                security.ProtectedEmailsList()[0],
		Field:                "name",
		Value:                "Renamed-Via-CLI",
		Stdout:               &bytes.Buffer{},
		Confirm:              true,
		IdentityConfirmation: "BREAK-GLASS",
	})
	if err != nil {
		t.Fatalf("AdminForceEditProtected name: %v", err)
	}

	// Direct audit-log query — the project's repository doesn't expose a
	// ListByEventType helper, so we verify via SQL.
	row := db.QueryRow(
		`SELECT COUNT(*) FROM audit_logs WHERE event_type = 'protected_user_break_glass_name'`,
	)
	var n int
	if err := row.Scan(&n); err != nil {
		t.Fatalf("audit count query: %v", err)
	}
	if n < 1 {
		t.Errorf("expected ≥1 protected_user_break_glass_name event, got %d", n)
	}
}
