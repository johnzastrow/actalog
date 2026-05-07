package integration

// admin_user_lifecycle_test.go — Task 8: backend integration tests for the
// v1.3.1 admin user lifecycle (CreateUser + SetPassword).
//
// Pattern matches protected_users_layered_defense_test.go: a real
// AdminUserService is wired against real repositories backed by the matrix-
// managed test DB (sqlite3 by default; postgres/mysql in CI). Service methods
// are called directly and DB state plus audit events are verified.
//
// HTTP-layer behaviour for these endpoints is unit-tested in T3/T6 with stubs;
// the L1 ProtectedUserGuard is covered by TestL3_* / TestDefense_* tests in
// neighbouring files. This file fills the remaining gap: end-to-end service
// behaviour against a real DB.

import (
	"errors"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/auth"
	"github.com/johnzastrow/actalog/pkg/security"
)

// buildAdminUserServiceWithRefreshRepo builds an AdminUserService using a
// real (caller-supplied) RefreshTokenRepository so that SetPassword's
// RevokeAllForUser path actually persists. The sister helper
// buildAdminUserService in protected_users_layered_defense_test.go uses a
// no-op stub, which is fine for L2 tests but not for verifying revocation
// side-effects.
func buildAdminUserServiceWithRefreshRepo(
	userRepo domain.UserRepository,
	refreshRepo domain.RefreshTokenRepository,
	audit service.AdminAuditLogger,
) *service.AdminUserService {
	log := &defenseGuardLogger{}
	return service.NewAdminUserService(
		userRepo, refreshRepo, &defenseEmailSvc{}, audit, log, "http://test.local",
	)
}

// TestAdminCreateUser_ThenAuthenticate verifies the create round-trip: admin
// creates a user via AdminUserService.CreateUser, the user is persisted, and
// the stored bcrypt hash validates against the original plaintext password.
func TestAdminCreateUser_ThenAuthenticate(t *testing.T) {
	db, _ := mustOpenTestDB(t)

	userRepo := repository.NewSQLiteUserRepository(db)
	refreshRepo := repository.NewSQLiteRefreshTokenRepository(db)
	audit := &defenseAuditLogger{}
	svc := buildAdminUserServiceWithRefreshRepo(userRepo, refreshRepo, audit)

	created, err := svc.CreateUser(1, service.CreateUserFields{
		Email:         "newathlete@example.com",
		Password:      "NewAthletePass123",
		Name:          "New Athlete",
		Role:          "athlete",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected created user to have an assigned ID")
	}

	// Re-load via the repo and verify the stored hash matches the plaintext.
	stored, err := userRepo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored == nil {
		t.Fatal("stored user is nil")
	}
	if cmpErr := auth.CheckPassword(stored.PasswordHash, "NewAthletePass123"); cmpErr != nil {
		t.Errorf("bcrypt CheckPassword failed against stored hash: %v", cmpErr)
	}
	if !stored.EmailVerified {
		t.Error("EmailVerified should be true (admin-vouched)")
	}

	// Audit verification.
	if got := audit.countByType(domain.EventAdminUserCreated); got != 1 {
		t.Errorf("expected exactly 1 EventAdminUserCreated, got %d (events: %v)", got, audit.calls)
	}
}

// TestAdminCreateUser_ProtectedEmailRejected ensures a protected email cannot
// be used to create a new account, and that the rejection produces the
// expected audit event.
func TestAdminCreateUser_ProtectedEmailRejected(t *testing.T) {
	db, _ := mustOpenTestDB(t)

	userRepo := repository.NewSQLiteUserRepository(db)
	refreshRepo := repository.NewSQLiteRefreshTokenRepository(db)
	audit := &defenseAuditLogger{}
	svc := buildAdminUserServiceWithRefreshRepo(userRepo, refreshRepo, audit)

	emails := security.ProtectedEmailsList()
	if len(emails) == 0 {
		t.Fatal("security.ProtectedEmailsList returned empty slice")
	}
	protectedEmail := emails[0]

	_, err := svc.CreateUser(1, service.CreateUserFields{
		Email:    protectedEmail,
		Password: "ValidPass123Long",
		Name:     "Imp",
		Role:     "athlete",
	})
	if err == nil {
		t.Fatal("expected protected-email rejection, got nil")
	}

	var invErr *domain.InvalidInputError
	if !errors.As(err, &invErr) {
		t.Fatalf("expected *domain.InvalidInputError, got %T: %v", err, err)
	}
	if invErr.Field != "email" {
		t.Errorf("InvalidInputError.Field = %q, want %q", invErr.Field, "email")
	}

	if got := audit.countByType(domain.EventAdminUserCreateRejectedProtected); got != 1 {
		t.Errorf("expected exactly 1 EventAdminUserCreateRejectedProtected, got %d (events: %v)",
			got, audit.calls)
	}
}

// TestAdminSetPassword_ClearsLockoutAndAuthenticates locks a user, calls
// AdminUserService.SetPassword, then verifies:
//   - failed_login_attempts is reset to 0
//   - locked_until is NULL
//   - the new password validates via bcrypt
//   - the old password no longer validates
//   - exactly one EventAdminPasswordSet audit event fires
func TestAdminSetPassword_ClearsLockoutAndAuthenticates(t *testing.T) {
	db, _ := mustOpenTestDB(t)

	userRepo := repository.NewSQLiteUserRepository(db)
	refreshRepo := repository.NewSQLiteRefreshTokenRepository(db)
	audit := &defenseAuditLogger{}
	svc := buildAdminUserServiceWithRefreshRepo(userRepo, refreshRepo, audit)

	// Create the target via the admin endpoint so we get a real bcrypt hash.
	created, err := svc.CreateUser(1, service.CreateUserFields{
		Email:         "victim@example.com",
		Password:      "OriginalPass123",
		Name:          "Victim",
		Role:          "athlete",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// Lock the target via the repo helper and bump failed_login_attempts to 5
	// using the existing IncrementFailedAttempts helper (no raw SQL needed).
	if err := userRepo.LockAccount(created.ID, 15*time.Minute); err != nil {
		t.Fatalf("LockAccount: %v", err)
	}
	for i := 0; i < 5; i++ {
		if err := userRepo.IncrementFailedAttempts(created.ID); err != nil {
			t.Fatalf("IncrementFailedAttempts (i=%d): %v", i, err)
		}
	}

	// Sanity: the target really is locked with 5 failed attempts before SetPassword.
	preState, err := userRepo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID (pre): %v", err)
	}
	if preState.FailedLoginAttempts != 5 {
		t.Fatalf("pre-state FailedLoginAttempts = %d, want 5", preState.FailedLoginAttempts)
	}
	if preState.LockedUntil == nil {
		t.Fatal("pre-state LockedUntil should be non-nil after LockAccount")
	}

	// Admin sets a new password.
	if err := svc.SetPassword(1, created.ID, "AdminFixedPass45A"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	// Verify lockout cleared.
	stored, err := userRepo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID (post): %v", err)
	}
	if stored.FailedLoginAttempts != 0 {
		t.Errorf("FailedLoginAttempts = %d, want 0 (lockout should be cleared)", stored.FailedLoginAttempts)
	}
	if stored.LockedUntil != nil {
		t.Errorf("LockedUntil should be nil after SetPassword; got %v", stored.LockedUntil)
	}
	// New password validates.
	if cmpErr := auth.CheckPassword(stored.PasswordHash, "AdminFixedPass45A"); cmpErr != nil {
		t.Errorf("new password should validate via bcrypt; got error: %v", cmpErr)
	}
	// Old password no longer validates.
	if cmpErr := auth.CheckPassword(stored.PasswordHash, "OriginalPass123"); cmpErr == nil {
		t.Error("old password should NOT validate against the new hash")
	}

	// Audit: exactly one EventAdminPasswordSet (the create event also fires,
	// so we assert per-type counts rather than total).
	if got := audit.countByType(domain.EventAdminPasswordSet); got != 1 {
		t.Errorf("expected exactly 1 EventAdminPasswordSet, got %d (events: %v)", got, audit.calls)
	}
}

// TestAdminSetPassword_RejectsProtectedTarget verifies the L2 service-layer
// defence fires when an admin tries to set a password on a protected user.
// Real DB, real service.
func TestAdminSetPassword_RejectsProtectedTarget(t *testing.T) {
	db, driver := mustOpenTestDB(t)

	userRepo := repository.NewSQLiteUserRepository(db)
	refreshRepo := repository.NewSQLiteRefreshTokenRepository(db)
	audit := &defenseAuditLogger{}
	svc := buildAdminUserServiceWithRefreshRepo(userRepo, refreshRepo, audit)

	// Insert the protected user via the helper from protected_users_test.go.
	insertProtectedUser(t, db, driver)
	protectedID := fetchProtectedUserID(t, userRepo)

	err := svc.SetPassword(1, protectedID, "ValidPass123Long")
	if err == nil {
		t.Fatal("expected ErrProtectedUser, got nil")
	}
	if !errors.Is(err, domain.ErrProtectedUser) {
		t.Errorf("expected ErrProtectedUser, got: %v", err)
	}
	if got := audit.countByType(domain.EventProtectedUserAttackService); got != 1 {
		t.Errorf("expected exactly 1 EventProtectedUserAttackService, got %d (events: %v)",
			got, audit.calls)
	}
}
