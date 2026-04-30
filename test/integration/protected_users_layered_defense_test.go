package integration

// protected_users_layered_defense_test.go — Task 14: adversarial bypass tests.
//
// Each test deliberately skips one or more outer layers and confirms the next
// layer catches the attempt independently. Audit-event tagging is the rejecting
// layer only — exactly one event fires per attempt.
//
// Test matrix (4 tests):
//   1. TestDefense_BypassL1_ServiceCatchesIt        — L2 catches when L1 is bypassed.
//   2. TestDefense_BypassL1AndL2_DBTriggerCatchesIt — L3 catches when L1+L2 are bypassed.
//   3. TestDefense_AllLayersFireCorrectAuditEvents  — Table-driven audit-event correctness.
//   4. TestDefense_LayerOrderingPropagation         — L4 tags EventProtectedUserAttackDB when
//                                                     L2 is fooled but L3 catches.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/middleware"
	"github.com/johnzastrow/actalog/pkg/security"
)

// ---------------------------------------------------------------------------
// Local stubs — kept in this file to avoid polluting other test files.
// ---------------------------------------------------------------------------

// defenseAuditLogger records LogEvent calls and can be inspected per-test.
type defenseAuditLogger struct {
	calls []defenseAuditCall
}

type defenseAuditCall struct {
	eventType    string
	userID       *int64
	targetUserID *int64
}

func (a *defenseAuditLogger) LogEvent(
	eventType string,
	userID *int64,
	targetUserID *int64,
	_ *string,
	_ *string,
	_ map[string]interface{},
) error {
	a.calls = append(a.calls, defenseAuditCall{eventType, userID, targetUserID})
	return nil
}

func (a *defenseAuditLogger) countByType(eventType string) int {
	n := 0
	for _, c := range a.calls {
		if c.eventType == eventType {
			n++
		}
	}
	return n
}

func (a *defenseAuditLogger) countProtectedUserAttacks() int {
	return a.countByType(domain.EventProtectedUserAttackHTTP) +
		a.countByType(domain.EventProtectedUserAttackService) +
		a.countByType(domain.EventProtectedUserAttackDB)
}

// defenseGuardLogger satisfies service.AdminGuardLogger and middleware.GuardLogger.
type defenseGuardLogger struct {
	errors []string
}

func (l *defenseGuardLogger) Error(format string, v ...interface{}) {
	l.errors = append(l.errors, fmt.Sprintf(format, v...))
}

// defenseEmailSvc is a no-op email sender that satisfies service.AdminEmailSender.
type defenseEmailSvc struct{}

func (e *defenseEmailSvc) SendPasswordResetEmail(_, _ string) error { return nil }
func (e *defenseEmailSvc) SendVerificationEmail(_, _ string) error  { return nil }

// defenseRefreshTokenRepo is a no-op that satisfies domain.RefreshTokenRepository.
type defenseRefreshTokenRepo struct{}

func (r *defenseRefreshTokenRepo) Create(_ *domain.RefreshToken) error               { return nil }
func (r *defenseRefreshTokenRepo) GetByToken(_ string) (*domain.RefreshToken, error) { return nil, nil }
func (r *defenseRefreshTokenRepo) GetByUserID(_ int64) ([]*domain.RefreshToken, error) {
	return nil, nil
}
func (r *defenseRefreshTokenRepo) Revoke(_ int64) error           { return nil }
func (r *defenseRefreshTokenRepo) RevokeAllForUser(_ int64) error { return nil }
func (r *defenseRefreshTokenRepo) DeleteExpired() error           { return nil }
func (r *defenseRefreshTokenRepo) Delete(_ int64) error           { return nil }

// ---------------------------------------------------------------------------
// fetchProtectedUserID looks up the ID of the protected row that was inserted
// by insertProtectedUser. Needed to call service/repo methods by ID.
// ---------------------------------------------------------------------------

func fetchProtectedUserID(t *testing.T, userRepo domain.UserRepository) int64 {
	t.Helper()
	emails := security.ProtectedEmailsList()
	if len(emails) == 0 {
		t.Fatal("security.ProtectedEmailsList returned empty slice")
	}
	u, err := userRepo.GetByEmail(emails[0])
	if err != nil {
		t.Fatalf("fetchProtectedUserID: GetByEmail(%q): %v", emails[0], err)
	}
	if u == nil {
		t.Fatalf("fetchProtectedUserID: protected user not found in DB (did you call insertProtectedUser first?)")
	}
	return u.ID
}

// fetchNonProtectedUserID looks up the ID of the non-protected user inserted
// by insertNonProtectedUser.
func fetchNonProtectedUserID(t *testing.T, userRepo domain.UserRepository) int64 {
	t.Helper()
	u, err := userRepo.GetByEmail("alice@example.com")
	if err != nil {
		t.Fatalf("fetchNonProtectedUserID: GetByEmail: %v", err)
	}
	if u == nil {
		t.Fatalf("fetchNonProtectedUserID: alice not found (call insertNonProtectedUser first)")
	}
	return u.ID
}

// ---------------------------------------------------------------------------
// Helper: build a real AdminUserService backed by the given userRepo.
// ---------------------------------------------------------------------------

func buildAdminUserService(
	userRepo domain.UserRepository,
	audit service.AdminAuditLogger,
) *service.AdminUserService {
	log := &defenseGuardLogger{}
	return service.NewAdminUserService(
		userRepo,
		&defenseRefreshTokenRepo{},
		&defenseEmailSvc{},
		audit,
		log,
		"http://test.local",
	)
}

// ---------------------------------------------------------------------------
// Test 1 — Bypass L1 (HTTP middleware); L2 service must catch.
// ---------------------------------------------------------------------------

// TestDefense_BypassL1_ServiceCatchesIt constructs a real AdminUserService
// directly (no chi router, no ProtectedUserGuard) and calls UpdateProfile on
// a protected target. L2 (ensureNotProtected) must:
//   - return domain.ErrProtectedUser
//   - write exactly one EventProtectedUserAttackService audit event
//   - NOT write any EventProtectedUserAttackHTTP event (L1 was never invoked)
func TestDefense_BypassL1_ServiceCatchesIt(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)

	userRepo := repository.NewSQLiteUserRepository(db)
	audit := &defenseAuditLogger{}
	svc := buildAdminUserService(userRepo, audit)

	protectedID := fetchProtectedUserID(t, userRepo)
	actorID := int64(999) // a fake admin actor who bypassed L1

	fields := service.ProfileUpdateFields{Name: strPtr("hacked-name")}

	// We need a non-zero UpdatedAt to avoid confusing ErrConflict with ErrProtectedUser.
	// The L2 check fires before the optimistic-lock check, so this value doesn't matter —
	// but let's use a plausible time.
	_, err := svc.UpdateProfile(actorID, protectedID, fields, time.Now().Add(-1*time.Hour))

	// L2 must have caught it.
	if !errors.Is(err, domain.ErrProtectedUser) {
		t.Errorf("expected ErrProtectedUser from L2; got: %v", err)
	}

	// Exactly one service-layer audit event.
	if n := audit.countByType(domain.EventProtectedUserAttackService); n != 1 {
		t.Errorf("expected exactly 1 EventProtectedUserAttackService, got %d", n)
	}

	// No HTTP-layer event (L1 was bypassed entirely).
	if n := audit.countByType(domain.EventProtectedUserAttackHTTP); n != 0 {
		t.Errorf("expected 0 EventProtectedUserAttackHTTP (L1 bypassed), got %d", n)
	}

	// Total protected-user attack events: exactly 1.
	if n := audit.countProtectedUserAttacks(); n != 1 {
		t.Errorf("expected exactly 1 total protected-user attack event, got %d", n)
	}
}

// ---------------------------------------------------------------------------
// Test 2 — Bypass L1 AND L2 (call userRepo.Update directly); L3 must catch.
// ---------------------------------------------------------------------------

// TestDefense_BypassL1AndL2_DBTriggerCatchesIt calls userRepo.Update directly
// on the protected user, skipping both the HTTP middleware (L1) and the service
// guard (L2). The database trigger (L3) must fire and return an error containing
// the contract substring.
func TestDefense_BypassL1AndL2_DBTriggerCatchesIt(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)

	userRepo := repository.NewSQLiteUserRepository(db)
	protectedID := fetchProtectedUserID(t, userRepo)

	// Fetch the current state so we have a valid User struct to pass to Update.
	user, err := userRepo.GetByID(protectedID)
	if err != nil || user == nil {
		t.Fatalf("GetByID(%d): user=%v err=%v", protectedID, user, err)
	}

	// Mutate a field and call Update directly — no middleware, no service guard.
	user.Name = "bypassed-l1-l2"
	user.UpdatedAt = time.Now().UTC()

	updateErr := userRepo.Update(user)

	// The DB trigger must have blocked it.
	if updateErr == nil {
		t.Fatal("expected userRepo.Update to fail for protected user (L3 trigger), but it succeeded")
	}

	// The error must contain the L3 contract substring.
	if !strings.Contains(updateErr.Error(), security.TriggerErrorContract) {
		t.Errorf("L3 trigger error missing contract substring %q; got: %v",
			security.TriggerErrorContract, updateErr)
	}
}

// ---------------------------------------------------------------------------
// Test 3 — Table-driven: each layer fires the correct (and only) audit event.
// ---------------------------------------------------------------------------

// TestDefense_AllLayersFireCorrectAuditEvents exercises three attack paths and
// asserts that:
//   - The correct audit event fires (tagged at the catching layer).
//   - No duplicate events are produced (exactly 1 event total per attempt).
//
// Cases:
//
//	"L1_catches_http_attempt"   — real HTTP request through chi + ProtectedUserGuard
//	"L2_catches_service_call"   — direct service call (L1 bypassed)
//	"L3_catches_repo_call"      — direct repo.Update call (L1+L2 bypassed); NO audit event fires
func TestDefense_AllLayersFireCorrectAuditEvents(t *testing.T) {
	cases := []struct {
		name               string
		run                func(t *testing.T, userRepo domain.UserRepository, audit *defenseAuditLogger, protectedID int64) error
		wantEventType      string // "" means no event expected
		wantEventCount     int    // total protected-user-attack events expected
		wantErrSubstring   string // if non-empty, the error must contain this
		wantErrIsProtected bool   // if true, errors.Is(err, ErrProtectedUser) must hold
	}{
		{
			name: "L1_catches_http_attempt",
			run: func(t *testing.T, userRepo domain.UserRepository, audit *defenseAuditLogger, protectedID int64) error {
				// Build a chi router with ProtectedUserGuard; the stub handler is
				// never reached because L1 fires 403 first.
				guardLog := &defenseGuardLogger{}
				router := chi.NewRouter()
				router.Route("/admin/users/{id}", func(r chi.Router) {
					r.Use(middleware.ProtectedUserGuard(userRepo, audit, guardLog))
					r.Patch("/", func(w http.ResponseWriter, req *http.Request) {
						w.WriteHeader(http.StatusOK)
					})
				})
				req := httptest.NewRequest(http.MethodPatch,
					fmt.Sprintf("/admin/users/%d/", protectedID), nil)
				// Inject an authenticated actor via context.
				actorID := int64(1)
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, actorID)
				req = req.WithContext(ctx)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)

				if rec.Code != http.StatusForbidden {
					t.Errorf("expected 403 from L1 guard, got %d", rec.Code)
				}
				return nil
			},
			wantEventType:  domain.EventProtectedUserAttackHTTP,
			wantEventCount: 1,
		},
		{
			name: "L2_catches_service_call",
			run: func(t *testing.T, userRepo domain.UserRepository, audit *defenseAuditLogger, protectedID int64) error {
				svc := buildAdminUserService(userRepo, audit)
				actorID := int64(1)
				fields := service.ProfileUpdateFields{Name: strPtr("hacked")}
				_, err := svc.UpdateProfile(actorID, protectedID, fields, time.Now().Add(-time.Hour))
				return err
			},
			wantEventType:      domain.EventProtectedUserAttackService,
			wantEventCount:     1,
			wantErrIsProtected: true,
		},
		{
			name: "L3_catches_repo_call",
			run: func(t *testing.T, userRepo domain.UserRepository, audit *defenseAuditLogger, protectedID int64) error {
				user, err := userRepo.GetByID(protectedID)
				if err != nil || user == nil {
					t.Fatalf("GetByID: user=%v err=%v", user, err)
				}
				user.Name = "bypassed"
				user.UpdatedAt = time.Now().UTC()
				return userRepo.Update(user)
			},
			wantEventType:    "", // no audit event from repo.Update itself
			wantEventCount:   0,
			wantErrSubstring: security.TriggerErrorContract,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db, driver := mustOpenTestDB(t)
			insertProtectedUser(t, db, driver)

			userRepo := repository.NewSQLiteUserRepository(db)
			audit := &defenseAuditLogger{}

			protectedID := fetchProtectedUserID(t, userRepo)

			gotErr := tc.run(t, userRepo, audit, protectedID)

			// Error assertions.
			if tc.wantErrIsProtected {
				if !errors.Is(gotErr, domain.ErrProtectedUser) {
					t.Errorf("[%s] expected ErrProtectedUser, got %v", tc.name, gotErr)
				}
			}
			if tc.wantErrSubstring != "" {
				if gotErr == nil || !strings.Contains(gotErr.Error(), tc.wantErrSubstring) {
					t.Errorf("[%s] expected error containing %q, got %v",
						tc.name, tc.wantErrSubstring, gotErr)
				}
			}

			// Audit-event count assertions.
			if got := audit.countProtectedUserAttacks(); got != tc.wantEventCount {
				t.Errorf("[%s] expected %d total protected-user-attack events, got %d (events: %v)",
					tc.name, tc.wantEventCount, got, audit.calls)
			}
			if tc.wantEventType != "" {
				if n := audit.countByType(tc.wantEventType); n != 1 {
					t.Errorf("[%s] expected exactly 1 %s event, got %d",
						tc.name, tc.wantEventType, n)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test 4 — Layer ordering: when L2 is fooled but L3 catches, the audit event
//           is tagged EventProtectedUserAttackDB (the rejecting layer).
// ---------------------------------------------------------------------------

// TestDefense_LayerOrderingPropagation proves that when:
//   - L1 is bypassed (no HTTP middleware)
//   - L2 is fooled (sees a non-protected user for GetByID)
//   - L3 fires (the DB trigger catches the real protected row being mutated)
//
// ...the service's L4 wrapping tags the audit event as EventProtectedUserAttackDB
// (the layer that actually rejected), NOT EventProtectedUserAttackService (which
// traversed but did not reject).
//
// Implementation: construct a fullyLyingRepo that returns a fake non-protected
// user from GetByID, but overrides Update via mockUpdateRepo to return a
// trigger-contract error. This directly tests the L4 audit-event tagging path:
// when Update returns the contract error, L4 tags EventProtectedUserAttackDB.
func TestDefense_LayerOrderingPropagation(t *testing.T) {
	db, driver := mustOpenTestDB(t)
	insertProtectedUser(t, db, driver)
	insertNonProtectedUser(t, db, driver)

	realRepo := repository.NewSQLiteUserRepository(db)

	// The fake non-protected user that L2 will see when it calls GetByID.
	nonProtectedID := fetchNonProtectedUserID(t, realRepo)
	fakeUser, err := realRepo.GetByID(nonProtectedID)
	if err != nil || fakeUser == nil {
		t.Fatalf("GetByID(non-protected): user=%v err=%v", fakeUser, err)
	}

	protectedID := fetchProtectedUserID(t, realRepo)

	// The lying repo: GetByID always returns the non-protected fake user, but
	// UpdateProfile will eventually call Update with the PROTECTED user's data
	// (because the service re-fetches via GetByID and then patches the struct).
	// To trigger L3 from UpdateProfile, we need the Update call to target the
	// protected row. Since GetByID is overridden to return the fake non-protected
	// user (including its ID), UpdateProfile will attempt to Update(fakeUser) — which
	// is non-protected. That won't hit L3.
	//
	// A cleaner approach: configure the lying repo's Update to return a
	// trigger-contract error, which exercises the exact L4 error-wrapping path
	// in UpdateProfile without needing a real DB trigger interaction. This directly
	// tests: "when Update returns the contract error, does L4 tag EventProtectedUserAttackDB?"
	//
	// This is option (b) from the plan — mock the Update error — and it tests
	// what matters: the L4 audit-event tagging is EventProtectedUserAttackDB, not
	// EventProtectedUserAttackService. The lying-repo approach would also work,
	// but the mocked-Update variant is simpler and unambiguous.
	triggerLikeErr := fmt.Errorf("SQL trigger: %s", security.TriggerErrorContract)

	// mockUpdateRepo wraps realRepo but returns a trigger-contract error from Update.
	mockUpdate := &mockUpdateRepo{
		UserRepository: realRepo,
		updateErr:      triggerLikeErr,
	}

	// Override GetByID to return the fake non-protected user (so L2 passes),
	// but Update returns the trigger-contract error (so L4 catches it).
	liar := &fullyLyingRepo{
		UserRepository: mockUpdate,
		fakeUser:       fakeUser,
	}

	audit := &defenseAuditLogger{}
	svc := buildAdminUserService(liar, audit)

	actorID := int64(1)
	fields := service.ProfileUpdateFields{Name: strPtr("layer-order-test")}

	// Use fakeUser.UpdatedAt so the optimistic-lock check passes (the liar's
	// GetByID returns fakeUser, so the re-fetch inside UpdateProfile sees fakeUser).
	_, updateErr := svc.UpdateProfile(actorID, protectedID, fields, fakeUser.UpdatedAt)

	// The error returned to the caller must be ErrProtectedUser (L4 wraps it).
	if !errors.Is(updateErr, domain.ErrProtectedUser) {
		t.Errorf("expected ErrProtectedUser from L4 wrapping, got %v", updateErr)
	}

	// L4 must have tagged EventProtectedUserAttackDB (the DB layer caught it),
	// NOT EventProtectedUserAttackService (which traversed but did not reject).
	if n := audit.countByType(domain.EventProtectedUserAttackDB); n != 1 {
		t.Errorf("expected exactly 1 EventProtectedUserAttackDB, got %d", n)
	}
	if n := audit.countByType(domain.EventProtectedUserAttackService); n != 0 {
		t.Errorf("expected 0 EventProtectedUserAttackService (L2 was fooled, not the catcher), got %d", n)
	}

	// Total: exactly 1 protected-user-attack event.
	if n := audit.countProtectedUserAttacks(); n != 1 {
		t.Errorf("expected exactly 1 total protected-user-attack event, got %d (events: %v)",
			n, audit.calls)
	}
}

// mockUpdateRepo wraps a domain.UserRepository and overrides Update to return
// a configurable error. All other methods delegate to the embedded repo.
type mockUpdateRepo struct {
	domain.UserRepository
	updateErr error
}

func (m *mockUpdateRepo) Update(_ *domain.User) error {
	return m.updateErr
}

// fullyLyingRepo wraps a domain.UserRepository and overrides GetByID to always
// return fakeUser. Update and all other methods delegate to the embedded repo.
type fullyLyingRepo struct {
	domain.UserRepository
	fakeUser *domain.User
}

func (f *fullyLyingRepo) GetByID(_ int64) (*domain.User, error) {
	cp := *f.fakeUser
	return &cp, nil
}

// ---------------------------------------------------------------------------
// Utility
// ---------------------------------------------------------------------------

// strPtr returns a pointer to the given string value.
// grep confirms this is the only definition in test/integration/*.go; if another
// test file in this package ever needs strPtr, move this declaration to a shared
// helpers file (e.g. test/integration/helpers_test.go) to avoid a build conflict.
func strPtr(s string) *string { return &s }
