package service

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/security"
)

// ---------------------------------------------------------------------------
// Test stubs — all stubs are local to this _test.go file.
// ---------------------------------------------------------------------------

// stubUserRepo implements domain.UserRepository for testing.
type stubUserRepo struct {
	users            map[int64]*domain.User
	updateErr        error
	updateErrOnce    bool // only fail the first Update call
	updateErrOnCallN int  // if > 0, fail the Nth Update call (1-based); 0 = disabled
	updateCallCount  int
}

func newStubUserRepo() *stubUserRepo {
	return &stubUserRepo{users: make(map[int64]*domain.User)}
}

func (r *stubUserRepo) addUser(u *domain.User) {
	r.users[u.ID] = u
}

func (r *stubUserRepo) GetByID(id int64) (*domain.User, error) {
	u := r.users[id]
	if u == nil {
		return nil, nil
	}
	// Return a shallow copy so tests can compare original vs saved state.
	cp := *u
	return &cp, nil
}

func (r *stubUserRepo) GetByEmail(email string) (*domain.User, error) {
	for _, u := range r.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return nil, nil
}

func (r *stubUserRepo) Update(user *domain.User) error {
	r.updateCallCount++
	// updateErrOnCallN: fail exactly on the Nth call (1-based); all other calls succeed.
	if r.updateErrOnCallN > 0 {
		if r.updateCallCount == r.updateErrOnCallN {
			return r.updateErr
		}
		// Not the target call — persist and return nil regardless of updateErr.
		cp := *user
		r.users[user.ID] = &cp
		return nil
	}
	if r.updateErr != nil {
		if r.updateErrOnce {
			// Consume the error after first use.
			err := r.updateErr
			r.updateErr = nil
			return err
		}
		return r.updateErr
	}
	// Persist in the map so subsequent GetByID sees the update.
	cp := *user
	r.users[user.ID] = &cp
	return nil
}

// Create assigns an ID and stores the user. Used by CreateUser tests; older
// tests that never call Create are unaffected.
func (r *stubUserRepo) Create(user *domain.User) error {
	if user.ID == 0 {
		var maxID int64
		for id := range r.users {
			if id > maxID {
				maxID = id
			}
		}
		user.ID = maxID + 1
	}
	cp := *user
	r.users[user.ID] = &cp
	return nil
}
func (r *stubUserRepo) GetByResetToken(token string) (*domain.User, error)        { return nil, nil }
func (r *stubUserRepo) GetByVerificationToken(token string) (*domain.User, error) { return nil, nil }
func (r *stubUserRepo) UpdatePassword(userID int64, hash string) error {
	if u, ok := r.users[userID]; ok {
		u.PasswordHash = hash
	}
	return nil
}
func (r *stubUserRepo) Delete(id int64) error                          { return nil }
func (r *stubUserRepo) List(limit, offset int) ([]*domain.User, error) { return nil, nil }
func (r *stubUserRepo) Count() (int64, error)                          { return 0, nil }
func (r *stubUserRepo) ListAdmins() ([]*domain.User, error)            { return nil, nil }
func (r *stubUserRepo) IncrementFailedAttempts(userID int64) error     { return nil }
func (r *stubUserRepo) ResetFailedAttempts(userID int64) error         { return nil }
func (r *stubUserRepo) LockAccount(userID int64, d time.Duration) error {
	return nil
}
func (r *stubUserRepo) UnlockAccount(userID int64) error {
	if u, ok := r.users[userID]; ok {
		u.FailedLoginAttempts = 0
		u.LockedAt = nil
		u.LockedUntil = nil
	}
	return nil
}
func (r *stubUserRepo) IsAccountLocked(userID int64) (bool, *time.Time, error) {
	return false, nil, nil
}
func (r *stubUserRepo) DisableAccount(userID int64, disabledBy int64, reason string) error {
	return nil
}
func (r *stubUserRepo) EnableAccount(userID int64) error  { return nil }
func (r *stubUserRepo) CountNewThisMonth() (int64, error) { return 0, nil }
func (r *stubUserRepo) CountDisabled() (int64, error)     { return 0, nil }
func (r *stubUserRepo) ListWithFilter(filter domain.UserListFilter, limit, offset int) ([]*domain.User, error) {
	return nil, nil
}
func (r *stubUserRepo) CountWithFilter(filter domain.UserListFilter) (int64, error) { return 0, nil }

// stubRefreshTokenRepo implements domain.RefreshTokenRepository.
type stubRefreshTokenRepo struct {
	revokeAllErr           error
	revokeAllCallCount     int
	revokeAllForUserCalled bool
	revokeAllForUserID     int64
}

func (r *stubRefreshTokenRepo) Create(token *domain.RefreshToken) error { return nil }
func (r *stubRefreshTokenRepo) GetByToken(token string) (*domain.RefreshToken, error) {
	return nil, nil
}
func (r *stubRefreshTokenRepo) GetByUserID(userID int64) ([]*domain.RefreshToken, error) {
	return nil, nil
}
func (r *stubRefreshTokenRepo) Revoke(tokenID int64) error { return nil }
func (r *stubRefreshTokenRepo) RevokeAllForUser(userID int64) error {
	r.revokeAllCallCount++
	r.revokeAllForUserCalled = true
	r.revokeAllForUserID = userID
	return r.revokeAllErr
}
func (r *stubRefreshTokenRepo) DeleteExpired() error       { return nil }
func (r *stubRefreshTokenRepo) Delete(tokenID int64) error { return nil }

// stubEmailSvc implements AdminEmailSender.
type stubEmailSvc struct {
	resetCalls    []string // to addresses for password reset
	resetURLs     []string // URL arguments for password reset calls
	verifyCalls   []string // to addresses for verification
	verifyURLs    []string // URL arguments for verification calls
	sendResetErr  error
	sendVerifyErr error
}

func (s *stubEmailSvc) SendPasswordResetEmail(to, resetURL string) error {
	s.resetCalls = append(s.resetCalls, to)
	s.resetURLs = append(s.resetURLs, resetURL)
	return s.sendResetErr
}

func (s *stubEmailSvc) SendVerificationEmail(to, verifyURL string) error {
	s.verifyCalls = append(s.verifyCalls, to)
	s.verifyURLs = append(s.verifyURLs, verifyURL)
	return s.sendVerifyErr
}

// stubAuditLogger implements AdminAuditLogger.
type stubAuditLogger struct {
	events []auditEvent
	logErr error
}

type auditEvent struct {
	eventType    string
	userID       *int64
	targetUserID *int64
	details      map[string]interface{}
}

func (s *stubAuditLogger) LogEvent(
	eventType string,
	userID *int64,
	targetUserID *int64,
	ipAddress *string,
	userAgent *string,
	details map[string]interface{},
) error {
	s.events = append(s.events, auditEvent{eventType, userID, targetUserID, details})
	return s.logErr
}

func (s *stubAuditLogger) hasEvent(t string) bool {
	for _, e := range s.events {
		if e.eventType == t {
			return true
		}
	}
	return false
}

func (s *stubAuditLogger) lastEvent() *auditEvent {
	if len(s.events) == 0 {
		return nil
	}
	e := s.events[len(s.events)-1]
	return &e
}

// stubGuardLogger implements AdminGuardLogger.
type stubGuardLogger struct {
	errors []string
}

func (l *stubGuardLogger) Error(format string, v ...interface{}) {
	l.errors = append(l.errors, fmt.Sprintf(format, v...))
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeNormalUser(id int64, email, name string) *domain.User {
	now := time.Now().UTC().Truncate(time.Second)
	return &domain.User{
		ID:            id,
		Email:         email,
		Name:          name,
		EmailVerified: true,
		UpdatedAt:     now,
	}
}

func makeProtectedUser(id int64) *domain.User {
	// Use the first entry from the protected list.
	emails := security.ProtectedEmailsList()
	email := emails[0]
	return makeNormalUser(id, email, "Protected")
}

func newService(userRepo domain.UserRepository, rtRepo domain.RefreshTokenRepository, email AdminEmailSender, audit AdminAuditLogger) *AdminUserService {
	log := &stubGuardLogger{}
	return NewAdminUserService(userRepo, rtRepo, email, audit, log, "http://test.local")
}

// ---------------------------------------------------------------------------
// UpdateProfile tests
// ---------------------------------------------------------------------------

func TestAdminUserService_UpdateProfile_RejectsProtectedTarget(t *testing.T) {
	repo := newStubUserRepo()
	protected := makeProtectedUser(99)
	repo.addUser(protected)

	audit := &stubAuditLogger{}
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, audit)

	fields := ProfileUpdateFields{Name: strPtr("NewName")}
	_, err := svc.UpdateProfile(1, 99, fields, protected.UpdatedAt)

	if !errors.Is(err, domain.ErrProtectedUser) {
		t.Errorf("expected ErrProtectedUser, got %v", err)
	}
	if !audit.hasEvent(domain.EventProtectedUserAttackService) {
		t.Errorf("expected %s audit event to be fired", domain.EventProtectedUserAttackService)
	}
	// Actor and target IDs must be populated in the audit event.
	if e := audit.lastEvent(); e != nil {
		if e.userID == nil || *e.userID != 1 {
			t.Errorf("expected actorID=1, got %v", e.userID)
		}
		if e.targetUserID == nil || *e.targetUserID != 99 {
			t.Errorf("expected targetUserID=99, got %v", e.targetUserID)
		}
	}
}

func TestAdminUserService_UpdateProfile_OnlyUpdatesProvidedFields(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(42, "alice@example.com", "Alice")
	repo.addUser(user)

	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	// Only set Name; Email is NOT in fields.
	newName := "Alice Updated"
	fields := ProfileUpdateFields{Name: &newName}

	updated, err := svc.UpdateProfile(1, 42, fields, user.UpdatedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Alice Updated" {
		t.Errorf("expected Name=%q, got %q", "Alice Updated", updated.Name)
	}
	if updated.Email != "alice@example.com" {
		t.Errorf("Email must not change when not specified, got %q", updated.Email)
	}
}

func TestAdminUserService_UpdateProfile_EmailChangeResetsVerification(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(10, "bob@example.com", "Bob")
	user.EmailVerified = true
	repo.addUser(user)

	emailSvc := &stubEmailSvc{}
	audit := &stubAuditLogger{}
	svc := newService(repo, &stubRefreshTokenRepo{}, emailSvc, audit)

	newEmail := "bob-new@example.com"
	fields := ProfileUpdateFields{Email: &newEmail}

	updated, err := svc.UpdateProfile(1, 10, fields, user.UpdatedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.EmailVerified {
		t.Error("EmailVerified must be reset to false when email changes")
	}
	if len(emailSvc.verifyCalls) == 0 {
		t.Error("SendVerificationEmail must be called when email changes")
	}
	if emailSvc.verifyCalls[0] != "bob-new@example.com" {
		t.Errorf("SendVerificationEmail called with wrong address: %q", emailSvc.verifyCalls[0])
	}
	// The second argument must be a full URL, not a bare token.
	if len(emailSvc.verifyURLs) > 0 {
		url := emailSvc.verifyURLs[0]
		if !strings.HasPrefix(url, "http://test.local/verify-email?token=") {
			t.Errorf("SendVerificationEmail URL must be a full URL, got %q", url)
		}
	}
	// email_changed audit event must fire.
	if !audit.hasEvent(domain.EventEmailChanged) {
		t.Error("expected email_changed audit event")
	}
	// profile_updated audit event must also fire.
	if !audit.hasEvent(domain.EventProfileUpdated) {
		t.Error("expected profile_updated audit event")
	}
}

func TestAdminUserService_UpdateProfile_VerificationTokenStorageFailure_LogsAndContinues(t *testing.T) {
	// Scenario: the first Update (profile save) succeeds; the second Update
	// (token storage) fails. The service must:
	//   - return nil (profile update succeeded)
	//   - return the user with the new email and EmailVerified == false
	//   - fire EventEmailChanged audit event
	//   - NOT call SendVerificationEmail (never reached because token storage failed)
	//   - log the storage failure via the guard logger
	repo := newStubUserRepo()
	user := makeNormalUser(50, "old@example.com", "TestUser")
	user.EmailVerified = true
	repo.addUser(user)

	// Fail on the 2nd Update call (token storage); first call (profile) succeeds.
	repo.updateErr = fmt.Errorf("db write error on token storage")
	repo.updateErrOnCallN = 2

	emailSvc := &stubEmailSvc{}
	audit := &stubAuditLogger{}
	guardLog := &stubGuardLogger{}
	svc := NewAdminUserService(repo, &stubRefreshTokenRepo{}, emailSvc, audit, guardLog, "http://test.local")

	newEmail := "new@example.com"
	fields := ProfileUpdateFields{Email: &newEmail}

	updated, err := svc.UpdateProfile(1, 50, fields, user.UpdatedAt)

	// Profile update succeeded — must return nil error.
	if err != nil {
		t.Fatalf("UpdateProfile must return nil when only token storage fails, got: %v", err)
	}
	// Returned user must have the new email.
	if updated.Email != "new@example.com" {
		t.Errorf("expected Email=%q, got %q", "new@example.com", updated.Email)
	}
	// EmailVerified must be false (reset by email change logic).
	if updated.EmailVerified {
		t.Error("EmailVerified must be false after email change")
	}
	// EventEmailChanged must have fired (it runs after the second Update attempt).
	if !audit.hasEvent(domain.EventEmailChanged) {
		t.Error("expected email_changed audit event to fire even when token storage fails")
	}
	// SendVerificationEmail must NOT have been called — we never got there.
	if len(emailSvc.verifyCalls) != 0 {
		t.Errorf("SendVerificationEmail must not be called when token storage fails, got %d calls", len(emailSvc.verifyCalls))
	}
	// Guard logger must have recorded the storage failure.
	if len(guardLog.errors) == 0 {
		t.Error("expected guard logger to record the token storage failure")
	}
	foundStorageErr := false
	for _, msg := range guardLog.errors {
		if strings.Contains(msg, "failed to store verification token") {
			foundStorageErr = true
			break
		}
	}
	if !foundStorageErr {
		t.Errorf("guard logger must contain 'failed to store verification token', got: %v", guardLog.errors)
	}
}

func TestAdminUserService_UpdateProfile_StaleUpdatedAtReturnsErrConflict(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(5, "charlie@example.com", "Charlie")
	repo.addUser(user)

	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	staleTime := user.UpdatedAt.Add(-5 * time.Minute)
	fields := ProfileUpdateFields{Name: strPtr("StaleUpdate")}
	_, err := svc.UpdateProfile(1, 5, fields, staleTime)

	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("expected ErrConflict for stale updated_at, got %v", err)
	}
}

func TestAdminUserService_UpdateProfile_DBTriggerErrorIsTaggedAsAttackDB(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(7, "dave@example.com", "Dave")
	repo.addUser(user)

	// Inject the trigger-contract error on the Update call.
	repo.updateErr = fmt.Errorf("something: %s", security.TriggerErrorContract)

	audit := &stubAuditLogger{}
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, audit)

	fields := ProfileUpdateFields{Name: strPtr("Trigger")}
	_, err := svc.UpdateProfile(1, 7, fields, user.UpdatedAt)

	if !errors.Is(err, domain.ErrProtectedUser) {
		t.Errorf("expected ErrProtectedUser from DB trigger catch, got %v", err)
	}
	if !audit.hasEvent(domain.EventProtectedUserAttackDB) {
		t.Errorf("expected %s audit event for L4 DB trigger catch", domain.EventProtectedUserAttackDB)
	}
	// Actor and target must be populated.
	if e := audit.lastEvent(); e != nil {
		if e.userID == nil {
			t.Error("actorID must be populated in L4 DB audit event")
		}
		if e.targetUserID == nil {
			t.Error("targetUserID must be populated in L4 DB audit event")
		}
	}
}

func TestAdminUserService_UpdateProfile_AuditEventHasActorAndTarget(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(20, "eve@example.com", "Eve")
	repo.addUser(user)

	audit := &stubAuditLogger{}
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, audit)

	actorID := int64(1)
	targetID := int64(20)
	fields := ProfileUpdateFields{Name: strPtr("Eve Updated")}
	_, err := svc.UpdateProfile(actorID, targetID, fields, user.UpdatedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The profile_updated audit event must carry both IDs.
	for _, e := range audit.events {
		if e.eventType == domain.EventProfileUpdated {
			if e.userID == nil || *e.userID != actorID {
				t.Errorf("profile_updated: actorID want %d, got %v", actorID, e.userID)
			}
			if e.targetUserID == nil || *e.targetUserID != targetID {
				t.Errorf("profile_updated: targetUserID want %d, got %v", targetID, e.targetUserID)
			}
			return
		}
	}
	t.Error("profile_updated audit event not found")
}

// ---------------------------------------------------------------------------
// ForcePasswordReset tests
// ---------------------------------------------------------------------------

func TestAdminUserService_ForcePasswordReset_RejectsProtectedTarget(t *testing.T) {
	repo := newStubUserRepo()
	protected := makeProtectedUser(88)
	repo.addUser(protected)

	audit := &stubAuditLogger{}
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, audit)

	err := svc.ForcePasswordReset(1, 88)

	if !errors.Is(err, domain.ErrProtectedUser) {
		t.Errorf("expected ErrProtectedUser, got %v", err)
	}
	if !audit.hasEvent(domain.EventProtectedUserAttackService) {
		t.Errorf("expected %s audit event", domain.EventProtectedUserAttackService)
	}
	// Verify actor/target populated.
	if e := audit.lastEvent(); e != nil {
		if e.userID == nil || *e.userID != 1 {
			t.Errorf("actorID want 1, got %v", e.userID)
		}
		if e.targetUserID == nil || *e.targetUserID != 88 {
			t.Errorf("targetUserID want 88, got %v", e.targetUserID)
		}
	}
}

func TestAdminUserService_ForcePasswordReset_SendsEmailAndRevokesTokens(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(15, "frank@example.com", "Frank")
	repo.addUser(user)

	emailSvc := &stubEmailSvc{}
	rtRepo := &stubRefreshTokenRepo{}
	audit := &stubAuditLogger{}
	svc := newService(repo, rtRepo, emailSvc, audit)

	err := svc.ForcePasswordReset(1, 15)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Email must be sent.
	if len(emailSvc.resetCalls) == 0 {
		t.Error("SendPasswordResetEmail must be called")
	}
	if emailSvc.resetCalls[0] != "frank@example.com" {
		t.Errorf("reset email sent to wrong address: %q", emailSvc.resetCalls[0])
	}
	// The second argument must be a full URL, not a bare token.
	if len(emailSvc.resetURLs) > 0 {
		url := emailSvc.resetURLs[0]
		if !strings.HasPrefix(url, "http://test.local/reset-password/") {
			t.Errorf("SendPasswordResetEmail URL must be a full URL, got %q", url)
		}
	}

	// Refresh tokens must be revoked.
	if rtRepo.revokeAllCallCount != 1 {
		t.Errorf("RevokeAllForUser must be called once, called %d times", rtRepo.revokeAllCallCount)
	}

	// Audit event must fire with both IDs.
	if !audit.hasEvent(domain.EventPasswordResetForcedByAdmin) {
		t.Error("expected password_reset_forced_by_admin audit event")
	}
	for _, e := range audit.events {
		if e.eventType == domain.EventPasswordResetForcedByAdmin {
			if e.userID == nil || *e.userID != 1 {
				t.Errorf("actorID want 1, got %v", e.userID)
			}
			if e.targetUserID == nil || *e.targetUserID != 15 {
				t.Errorf("targetUserID want 15, got %v", e.targetUserID)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Additional branch-coverage tests
// ---------------------------------------------------------------------------

func TestAdminUserService_UpdateProfile_UserNotFound(t *testing.T) {
	repo := newStubUserRepo()
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	// Target ID 999 does not exist.
	_, err := svc.UpdateProfile(1, 999, ProfileUpdateFields{}, time.Now())
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

func TestAdminUserService_UpdateProfile_EmptyNameRejected(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(3, "grace@example.com", "Grace")
	repo.addUser(user)

	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	emptyName := "   "
	fields := ProfileUpdateFields{Name: &emptyName}
	_, err := svc.UpdateProfile(1, 3, fields, user.UpdatedAt)
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestAdminUserService_UpdateProfile_LongNameRejected(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(4, "hank@example.com", "Hank")
	repo.addUser(user)

	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	longName := strings.Repeat("x", 101)
	fields := ProfileUpdateFields{Name: &longName}
	_, err := svc.UpdateProfile(1, 4, fields, user.UpdatedAt)
	if err == nil {
		t.Fatal("expected error for name >100 chars, got nil")
	}
}

func TestAdminUserService_UpdateProfile_InvalidEmailRejected(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(6, "ivan@example.com", "Ivan")
	repo.addUser(user)

	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	badEmail := "not-an-email"
	fields := ProfileUpdateFields{Email: &badEmail}
	_, err := svc.UpdateProfile(1, 6, fields, user.UpdatedAt)
	if err == nil {
		t.Fatal("expected error for invalid email, got nil")
	}
}

func TestAdminUserService_UpdateProfile_FutureBirthdayRejected(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(8, "julia@example.com", "Julia")
	repo.addUser(user)

	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	future := time.Now().Add(24 * time.Hour)
	fields := ProfileUpdateFields{Birthday: &future}
	_, err := svc.UpdateProfile(1, 8, fields, user.UpdatedAt)
	if err == nil {
		t.Fatal("expected error for future birthday, got nil")
	}
}

func TestAdminUserService_UpdateProfile_SetBirthdaySuccess(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(9, "kyle@example.com", "Kyle")
	repo.addUser(user)

	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	bday := time.Now().Add(-365 * 24 * time.Hour) // one year ago
	fields := ProfileUpdateFields{Birthday: &bday}
	updated, err := svc.UpdateProfile(1, 9, fields, user.UpdatedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Birthday == nil {
		t.Error("expected Birthday to be set")
	}
}

func TestAdminUserService_UpdateProfile_AdminOverrideEmailVerified(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(11, "lena@example.com", "Lena")
	user.EmailVerified = false
	repo.addUser(user)

	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	verified := true
	fields := ProfileUpdateFields{EmailVerified: &verified}
	updated, err := svc.UpdateProfile(1, 11, fields, user.UpdatedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated.EmailVerified {
		t.Error("expected EmailVerified=true after admin override")
	}
}

func TestAdminUserService_UpdateProfile_SameEmailNoVerificationReset(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(12, "mark@example.com", "Mark")
	user.EmailVerified = true
	repo.addUser(user)

	emailSvc := &stubEmailSvc{}
	svc := newService(repo, &stubRefreshTokenRepo{}, emailSvc, &stubAuditLogger{})

	// Set email to the SAME value — should be a no-op for email change logic.
	sameEmail := "mark@example.com"
	fields := ProfileUpdateFields{Email: &sameEmail}
	updated, err := svc.UpdateProfile(1, 12, fields, user.UpdatedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated.EmailVerified {
		t.Error("EmailVerified must remain true when email didn't actually change")
	}
	if len(emailSvc.verifyCalls) != 0 {
		t.Error("SendVerificationEmail must not be called when email is unchanged")
	}
}

func TestAdminUserService_UpdateProfile_DBErrorPropagated(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(13, "nina@example.com", "Nina")
	repo.addUser(user)
	repo.updateErr = fmt.Errorf("connection lost")

	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	fields := ProfileUpdateFields{Name: strPtr("Nina2")}
	_, err := svc.UpdateProfile(1, 13, fields, user.UpdatedAt)
	if err == nil {
		t.Fatal("expected error from DB, got nil")
	}
	if errors.Is(err, domain.ErrProtectedUser) {
		t.Error("plain DB error must not be re-tagged as ErrProtectedUser")
	}
}

func TestAdminUserService_ForcePasswordReset_UserNotFound(t *testing.T) {
	repo := newStubUserRepo()
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	err := svc.ForcePasswordReset(1, 999)
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

func TestAdminUserService_ForcePasswordReset_EmailErrorPropagated(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(16, "oscar@example.com", "Oscar")
	repo.addUser(user)

	emailSvc := &stubEmailSvc{sendResetErr: fmt.Errorf("smtp down")}
	svc := newService(repo, &stubRefreshTokenRepo{}, emailSvc, &stubAuditLogger{})

	err := svc.ForcePasswordReset(1, 16)
	if err == nil {
		t.Fatal("expected error when email send fails")
	}
}

func TestAdminUserService_ForcePasswordReset_RevokeErrorPropagated(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(17, "petra@example.com", "Petra")
	repo.addUser(user)

	rtRepo := &stubRefreshTokenRepo{revokeAllErr: fmt.Errorf("revoke failed")}
	svc := newService(repo, rtRepo, &stubEmailSvc{}, &stubAuditLogger{})

	err := svc.ForcePasswordReset(1, 17)
	if err == nil {
		t.Fatal("expected error when RevokeAllForUser fails")
	}
}

func TestAdminUserService_AuditLogFailureSilentForNonFatalEvents(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(18, "quinn@example.com", "Quinn")
	repo.addUser(user)

	// Make all LogEvent calls fail.
	audit := &stubAuditLogger{logErr: fmt.Errorf("audit db down")}
	guardLog := &stubGuardLogger{}
	svc := NewAdminUserService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, audit, guardLog, "http://test.local")

	fields := ProfileUpdateFields{Name: strPtr("Quinn Updated")}
	// Must not return an error — audit failure on profile_updated is non-fatal.
	_, err := svc.UpdateProfile(1, 18, fields, user.UpdatedAt)
	if err != nil {
		t.Errorf("audit failure on non-fatal event must not surface as error, got: %v", err)
	}
	// The guard logger should have recorded the audit failure.
	if len(guardLog.errors) == 0 {
		t.Error("expected guard logger to record audit failure")
	}
}

func TestAdminUserService_UpdateProfile_GetByIDErrorAfterGuard(t *testing.T) {
	// ensureNotProtected succeeds (first GetByID returns a normal user),
	// but the second GetByID (in UpdateProfile body) fails.
	repo := &getByIDErrorOnSecondCallRepo{
		first: makeNormalUser(30, "second@example.com", "Second"),
	}
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	_, err := svc.UpdateProfile(1, 30, ProfileUpdateFields{}, time.Now())
	if err == nil {
		t.Fatal("expected error from second GetByID, got nil")
	}
}

func TestAdminUserService_ForcePasswordReset_UpdateErrorPropagated(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(31, "rex@example.com", "Rex")
	repo.addUser(user)
	repo.updateErr = fmt.Errorf("db write error")

	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	err := svc.ForcePasswordReset(1, 31)
	if err == nil {
		t.Fatal("expected error when Update fails during ForcePasswordReset")
	}
}

func TestAdminUserService_ForcePasswordReset_AuditFailureLogged(t *testing.T) {
	repo := newStubUserRepo()
	user := makeNormalUser(32, "steve@example.com", "Steve")
	repo.addUser(user)

	audit := &stubAuditLogger{logErr: fmt.Errorf("audit db down")}
	guardLog := &stubGuardLogger{}
	svc := NewAdminUserService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, audit, guardLog, "http://test.local")

	// Should succeed — audit failure on password_reset_forced_by_admin is non-fatal.
	err := svc.ForcePasswordReset(1, 32)
	if err != nil {
		t.Errorf("audit failure must not surface as error, got: %v", err)
	}
	if len(guardLog.errors) == 0 {
		t.Error("expected guard logger to record the audit failure")
	}
}

func TestAdminUserService_EnsureNotProtected_GetByIDError(t *testing.T) {
	repo := &alwaysErrorRepo{}
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	_, err := svc.UpdateProfile(1, 99, ProfileUpdateFields{}, time.Now())
	if err == nil {
		t.Fatal("expected error from GetByID failure in ensureNotProtected")
	}
}

// alwaysErrorRepo — GetByID always returns an error.
type alwaysErrorRepo struct{ stubUserRepo }

func (r *alwaysErrorRepo) GetByID(id int64) (*domain.User, error) {
	return nil, fmt.Errorf("simulated db error")
}

// getByIDErrorOnSecondCallRepo — first call succeeds, second returns error.
type getByIDErrorOnSecondCallRepo struct {
	callCount int
	first     *domain.User
}

func (r *getByIDErrorOnSecondCallRepo) GetByID(id int64) (*domain.User, error) {
	r.callCount++
	if r.callCount == 1 {
		cp := *r.first
		return &cp, nil
	}
	return nil, fmt.Errorf("simulated second GetByID error")
}

func (r *getByIDErrorOnSecondCallRepo) GetByEmail(email string) (*domain.User, error) {
	return nil, nil
}
func (r *getByIDErrorOnSecondCallRepo) Update(user *domain.User) error { return nil }
func (r *getByIDErrorOnSecondCallRepo) Create(user *domain.User) error { return nil }
func (r *getByIDErrorOnSecondCallRepo) GetByResetToken(token string) (*domain.User, error) {
	return nil, nil
}
func (r *getByIDErrorOnSecondCallRepo) GetByVerificationToken(token string) (*domain.User, error) {
	return nil, nil
}
func (r *getByIDErrorOnSecondCallRepo) UpdatePassword(userID int64, hash string) error { return nil }
func (r *getByIDErrorOnSecondCallRepo) Delete(id int64) error                          { return nil }
func (r *getByIDErrorOnSecondCallRepo) List(limit, offset int) ([]*domain.User, error) {
	return nil, nil
}
func (r *getByIDErrorOnSecondCallRepo) Count() (int64, error) { return 0, nil }
func (r *getByIDErrorOnSecondCallRepo) ListAdmins() ([]*domain.User, error) {
	return nil, nil
}
func (r *getByIDErrorOnSecondCallRepo) IncrementFailedAttempts(userID int64) error { return nil }
func (r *getByIDErrorOnSecondCallRepo) ResetFailedAttempts(userID int64) error     { return nil }
func (r *getByIDErrorOnSecondCallRepo) LockAccount(userID int64, d time.Duration) error {
	return nil
}
func (r *getByIDErrorOnSecondCallRepo) UnlockAccount(userID int64) error { return nil }
func (r *getByIDErrorOnSecondCallRepo) IsAccountLocked(userID int64) (bool, *time.Time, error) {
	return false, nil, nil
}
func (r *getByIDErrorOnSecondCallRepo) DisableAccount(userID int64, disabledBy int64, reason string) error {
	return nil
}
func (r *getByIDErrorOnSecondCallRepo) EnableAccount(userID int64) error  { return nil }
func (r *getByIDErrorOnSecondCallRepo) CountNewThisMonth() (int64, error) { return 0, nil }
func (r *getByIDErrorOnSecondCallRepo) CountDisabled() (int64, error)     { return 0, nil }
func (r *getByIDErrorOnSecondCallRepo) ListWithFilter(filter domain.UserListFilter, limit, offset int) ([]*domain.User, error) {
	return nil, nil
}
func (r *getByIDErrorOnSecondCallRepo) CountWithFilter(filter domain.UserListFilter) (int64, error) {
	return 0, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func strPtr(s string) *string { return &s }

// TestAdminUserService_UpdateProfile_ValidationErrorsAreTyped guards the
// contract that validation failures return *domain.InvalidInputError so the
// HTTP layer can surface them as 400 with a helpful message rather than the
// generic 500 path. Test 12 in the v1.3.0 manual-test plan caught this — a
// future-dated birthday returned a plain `errors.New(...)` that fell through
// to `WriteError`'s "an internal error occurred" branch.
func TestAdminUserService_UpdateProfile_ValidationErrorsAreTyped(t *testing.T) {
	user := makeNormalUser(20, "validator@example.com", "Vera")
	emptyName := ""
	tooLong := strings.Repeat("x", 101)
	badEmail := "not-an-email"
	future := time.Now().Add(24 * time.Hour)

	cases := []struct {
		name      string
		fields    ProfileUpdateFields
		wantField string
	}{
		{"empty name", ProfileUpdateFields{Name: &emptyName}, "name"},
		{"name over 100 chars", ProfileUpdateFields{Name: &tooLong}, "name"},
		{"unparseable email", ProfileUpdateFields{Email: &badEmail}, "email"},
		{"future birthday", ProfileUpdateFields{Birthday: &future}, "birthday"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newStubUserRepo()
			repo.addUser(user)
			svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

			_, err := svc.UpdateProfile(1, 20, tc.fields, user.UpdatedAt)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			var invErr *domain.InvalidInputError
			if !errors.As(err, &invErr) {
				t.Fatalf("expected *domain.InvalidInputError, got %T: %v", err, err)
			}
			if invErr.Field != tc.wantField {
				t.Errorf("Field = %q, want %q", invErr.Field, tc.wantField)
			}
			if invErr.Message == "" {
				t.Error("Message must be non-empty so the user knows what went wrong")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CreateUser tests
// ---------------------------------------------------------------------------

// TestAdminUserService_CreateUser_HappyPath verifies the admin can create a
// user account that the new user can immediately authenticate with.
func TestAdminUserService_CreateUser_HappyPath(t *testing.T) {
	repo := newStubUserRepo()
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	created, err := svc.CreateUser(1, CreateUserFields{
		Email:         "newathlete@example.com",
		Password:      "SetByAdminAtCreate1",
		Name:          "New Athlete",
		Role:          "athlete",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected user to have an assigned ID after Create")
	}
	if created.Email != "newathlete@example.com" {
		t.Errorf("Email = %q", created.Email)
	}
	if !created.EmailVerified {
		t.Error("EmailVerified should be true (admin-vouched)")
	}
	if created.PasswordHash == "" {
		t.Error("PasswordHash must be set")
	}
	if created.PasswordHash == "SetByAdminAtCreate1" {
		t.Error("PasswordHash must be hashed, not the raw password")
	}
}

// TestAdminUserService_CreateUser_ValidationRejections covers every input
// rule with a single table-driven test. Each case must return a typed
// *domain.InvalidInputError so the handler surfaces 400.
func TestAdminUserService_CreateUser_ValidationRejections(t *testing.T) {
	cases := []struct {
		name      string
		fields    CreateUserFields
		wantField string
	}{
		{"empty name", CreateUserFields{Email: "a@b.com", Password: "ValidPass123Long", Name: "", Role: "athlete"}, "name"},
		{"name over 100 chars", CreateUserFields{Email: "a@b.com", Password: "ValidPass123Long", Name: strings.Repeat("x", 101), Role: "athlete"}, "name"},
		{"malformed email", CreateUserFields{Email: "not-an-email", Password: "ValidPass123Long", Name: "X", Role: "athlete"}, "email"},
		{"password too short", CreateUserFields{Email: "a@b.com", Password: "Short1", Name: "X", Role: "athlete"}, "password"},
		{"password missing digit", CreateUserFields{Email: "a@b.com", Password: "NoDigitsHereXYZ", Name: "X", Role: "athlete"}, "password"},
		{"password missing upper", CreateUserFields{Email: "a@b.com", Password: "alllower123long", Name: "X", Role: "athlete"}, "password"},
		{"password missing lower", CreateUserFields{Email: "a@b.com", Password: "ALLUPPER123LONG", Name: "X", Role: "athlete"}, "password"},
		{"invalid role", CreateUserFields{Email: "a@b.com", Password: "ValidPass123Long", Name: "X", Role: "wizard"}, "role"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newStubUserRepo()
			svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

			_, err := svc.CreateUser(1, tc.fields)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			var invErr *domain.InvalidInputError
			if !errors.As(err, &invErr) {
				t.Fatalf("expected *domain.InvalidInputError, got %T: %v", err, err)
			}
			if invErr.Field != tc.wantField {
				t.Errorf("Field = %q, want %q", invErr.Field, tc.wantField)
			}
		})
	}
}

// TestAdminUserService_CreateUser_DuplicateEmail verifies the service rejects
// a second create with an email that already exists. Returns a sentinel the
// handler can map to HTTP 409.
func TestAdminUserService_CreateUser_DuplicateEmail(t *testing.T) {
	repo := newStubUserRepo()
	repo.addUser(makeNormalUser(7, "taken@example.com", "Existing"))
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	_, err := svc.CreateUser(1, CreateUserFields{
		Email: "taken@example.com", Password: "ValidPass123Long", Name: "Dup", Role: "athlete",
	})
	if err == nil {
		t.Fatal("expected duplicate-email error, got nil")
	}
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got: %v", err)
	}
}

// TestAdminUserService_CreateUser_ProtectedEmailRejected verifies that an
// admin cannot create a new user using an email on the protected-users list.
// This must fire EventAdminUserCreateRejectedProtected, NOT the
// protected_user_attack_* family.
func TestAdminUserService_CreateUser_ProtectedEmailRejected(t *testing.T) {
	protectedEmail := security.ProtectedEmailsList()[0]

	repo := newStubUserRepo()
	auditLog := &stubAuditLogger{}
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, auditLog)

	_, err := svc.CreateUser(1, CreateUserFields{
		Email: protectedEmail, Password: "ValidPass123Long", Name: "Imposter", Role: "admin",
	})
	if err == nil {
		t.Fatal("expected protected-email rejection, got nil")
	}
	var invErr *domain.InvalidInputError
	if !errors.As(err, &invErr) {
		t.Fatalf("expected *domain.InvalidInputError, got %T", err)
	}
	if invErr.Field != "email" {
		t.Errorf("Field = %q, want %q", invErr.Field, "email")
	}
	if !strings.Contains(invErr.Message, "reserved") {
		t.Errorf("Message should mention 'reserved'; got %q", invErr.Message)
	}
	if len(auditLog.events) != 1 {
		t.Fatalf("expected exactly 1 audit event, got %d: %v", len(auditLog.events), auditLog.events)
	}
	if auditLog.events[0].eventType != domain.EventAdminUserCreateRejectedProtected {
		t.Errorf("event = %q, want %q", auditLog.events[0].eventType, domain.EventAdminUserCreateRejectedProtected)
	}
}

// TestAdminUserService_CreateUser_AuditOnSuccess verifies the success path
// fires EventAdminUserCreated with the role and email_verified in the details.
func TestAdminUserService_CreateUser_AuditOnSuccess(t *testing.T) {
	repo := newStubUserRepo()
	auditLog := &stubAuditLogger{}
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, auditLog)

	_, err := svc.CreateUser(1, CreateUserFields{
		Email: "auditme@example.com", Password: "ValidPass123Long", Name: "Audit", Role: "coach", EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(auditLog.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditLog.events))
	}
	ev := auditLog.events[0]
	if ev.eventType != domain.EventAdminUserCreated {
		t.Errorf("eventType = %q, want %q", ev.eventType, domain.EventAdminUserCreated)
	}
	if ev.details["role"] != "coach" {
		t.Errorf("details.role = %v, want %q", ev.details["role"], "coach")
	}
	if ev.details["email_verified"] != true {
		t.Errorf("details.email_verified = %v, want true", ev.details["email_verified"])
	}
	if ev.details["email_domain"] != "example.com" {
		t.Errorf("details.email_domain = %v, want %q", ev.details["email_domain"], "example.com")
	}
}

// ---------------------------------------------------------------------------
// SetPassword tests
// ---------------------------------------------------------------------------

// TestAdminUserService_SetPassword_HappyPath verifies password is hashed +
// stored, lockout state cleared, refresh tokens revoked, and audit fired.
func TestAdminUserService_SetPassword_HappyPath(t *testing.T) {
	repo := newStubUserRepo()
	target := makeNormalUser(5, "lockedout@example.com", "Locked")
	target.FailedLoginAttempts = 3
	lockUntil := time.Now().Add(15 * time.Minute)
	target.LockedUntil = &lockUntil
	repo.addUser(target)

	tokenRepo := &stubRefreshTokenRepo{}
	auditLog := &stubAuditLogger{}
	svc := newService(repo, tokenRepo, &stubEmailSvc{}, auditLog)

	if err := svc.SetPassword(1, 5, "NewValidPass123A"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	stored := repo.users[5]
	if stored.PasswordHash == "" {
		t.Error("PasswordHash empty after SetPassword")
	}
	if stored.PasswordHash == "NewValidPass123A" {
		t.Error("PasswordHash equals plaintext — not hashed")
	}
	if stored.FailedLoginAttempts != 0 {
		t.Errorf("FailedLoginAttempts = %d, want 0", stored.FailedLoginAttempts)
	}
	if stored.LockedUntil != nil {
		t.Error("LockedUntil should be nil after SetPassword")
	}
	if !tokenRepo.revokeAllForUserCalled {
		t.Error("RevokeAllForUser not called")
	}
	if tokenRepo.revokeAllForUserID != 5 {
		t.Errorf("revoke for userID = %d, want 5", tokenRepo.revokeAllForUserID)
	}

	if len(auditLog.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditLog.events))
	}
	ev := auditLog.events[0]
	if ev.eventType != domain.EventAdminPasswordSet {
		t.Errorf("eventType = %q, want %q", ev.eventType, domain.EventAdminPasswordSet)
	}
	if ev.details["cleared_failed_login_attempts"] != 3 {
		t.Errorf("details.cleared_failed_login_attempts = %v, want 3", ev.details["cleared_failed_login_attempts"])
	}
	if ev.details["cleared_lockout"] != true {
		t.Errorf("details.cleared_lockout = %v, want true", ev.details["cleared_lockout"])
	}
}

// TestAdminUserService_SetPassword_404OnUnknownUser verifies a non-existent
// target returns ErrUserNotFound (handler maps to HTTP 404).
func TestAdminUserService_SetPassword_404OnUnknownUser(t *testing.T) {
	repo := newStubUserRepo()
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	err := svc.SetPassword(1, 99, "NewValidPass123A")
	if err == nil {
		t.Fatal("expected ErrUserNotFound, got nil")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}
}

// TestAdminUserService_SetPassword_RejectsWeakPassword verifies validation
// runs BEFORE any DB lookup and returns a typed *domain.InvalidInputError
// keyed on "new_password".
func TestAdminUserService_SetPassword_RejectsWeakPassword(t *testing.T) {
	repo := newStubUserRepo()
	target := makeNormalUser(5, "u@example.com", "U")
	repo.addUser(target)
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	err := svc.SetPassword(1, 5, "tooshort")
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var invErr *domain.InvalidInputError
	if !errors.As(err, &invErr) {
		t.Fatalf("expected *domain.InvalidInputError, got %T", err)
	}
	if invErr.Field != "new_password" {
		t.Errorf("Field = %q, want %q", invErr.Field, "new_password")
	}
}

// TestAdminUserService_SetPassword_NoLockoutPriorState verifies the audit
// records cleared_lockout=false when the user wasn't locked to begin with.
func TestAdminUserService_SetPassword_NoLockoutPriorState(t *testing.T) {
	repo := newStubUserRepo()
	target := makeNormalUser(5, "calm@example.com", "Calm")
	target.FailedLoginAttempts = 0
	target.LockedUntil = nil
	repo.addUser(target)

	auditLog := &stubAuditLogger{}
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, auditLog)

	if err := svc.SetPassword(1, 5, "NewValidPass123A"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	if auditLog.events[0].details["cleared_lockout"] != false {
		t.Errorf("cleared_lockout should be false; got %v", auditLog.events[0].details["cleared_lockout"])
	}
	if auditLog.events[0].details["cleared_failed_login_attempts"] != 0 {
		t.Errorf("cleared_failed_login_attempts should be 0; got %v", auditLog.events[0].details["cleared_failed_login_attempts"])
	}
}
