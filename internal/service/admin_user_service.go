package service

import (
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/auth"
	"github.com/johnzastrow/actalog/pkg/security"
)

// CreateUser uses the package-level ErrEmailAlreadyExists sentinel declared in
// user_service.go. Handlers should map it to HTTP 409.

// ProfileUpdateFields is a partial-update descriptor for admin profile edits.
// Only non-nil fields are written; absent fields are left unchanged.
type ProfileUpdateFields struct {
	Name          *string
	Email         *string
	Birthday      *time.Time
	EmailVerified *bool
}

// CreateUserFields is the input to AdminUserService.CreateUser. Defaults
// (role="athlete", EmailVerified=true) are applied inside CreateUser, not
// at decode time, so the wire format stays explicit.
type CreateUserFields struct {
	Email         string
	Password      string
	Name          string
	Role          string
	EmailVerified bool
}

// AdminEmailSender is the minimal email-service surface AdminUserService depends on.
// The concrete email.Service satisfies this interface.
// Note: the second parameter is a URL (not a bare token) — the existing email
// service formats the full reset/verify URL before sending.
type AdminEmailSender interface {
	SendPasswordResetEmail(to, resetURL string) error
	SendVerificationEmail(to, verifyURL string) error
}

// AdminAuditLogger is the minimal audit-service surface AdminUserService depends on.
// It matches the LogEvent method signature of service.AuditLogService exactly.
type AdminAuditLogger interface {
	LogEvent(
		eventType string,
		userID *int64,
		targetUserID *int64,
		ipAddress *string,
		userAgent *string,
		details map[string]interface{},
	) error
}

// AdminGuardLogger is the minimal logger interface AdminUserService uses for
// non-fatal operational errors (e.g., audit-write failures). It matches the
// GuardLogger interface already established in pkg/middleware.
type AdminGuardLogger interface {
	Error(format string, v ...interface{})
}

// AdminUserService provides admin-only user-management operations with a
// four-layer protected-user defence:
//
//   - L2 (this file): ensureNotProtected called at the start of every write.
//   - L4 (this file): DB-trigger error pattern caught in UpdateProfile.
//
// The other layers (L1 HTTP middleware, L3 repository trigger) are implemented
// separately and operate independently.
type AdminUserService struct {
	userRepo         domain.UserRepository
	refreshTokenRepo domain.RefreshTokenRepository
	emailService     AdminEmailSender
	auditLogService  AdminAuditLogger
	log              AdminGuardLogger
	appURL           string // Base URL for verification and reset links (e.g. "https://app.example.com")
}

// NewAdminUserService constructs an AdminUserService with all required dependencies.
// appURL is the base application URL used to build clickable links in emails
// (e.g. "https://app.example.com"). It must match the value passed to UserService
// so that admin-triggered emails are indistinguishable from user-self-service ones.
func NewAdminUserService(
	userRepo domain.UserRepository,
	refreshTokenRepo domain.RefreshTokenRepository,
	emailService AdminEmailSender,
	auditLogService AdminAuditLogger,
	log AdminGuardLogger,
	appURL string,
) *AdminUserService {
	return &AdminUserService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		emailService:     emailService,
		auditLogService:  auditLogService,
		log:              log,
		appURL:           appURL,
	}
}

// ensureNotProtected is the L2 defence layer. It loads the target user and
// returns ErrProtectedUser (plus writes a protected_user_attack_service audit
// event) if the target's email is in the protected-user registry.
//
// Every write method must call this before performing any mutation.
func (s *AdminUserService) ensureNotProtected(actorID, targetID int64) error {
	user, err := s.userRepo.GetByID(targetID)
	if err != nil {
		return fmt.Errorf("ensureNotProtected: GetByID(%d): %w", targetID, err)
	}
	if user == nil {
		return fmt.Errorf("ensureNotProtected: user %d: %w", targetID, ErrUserNotFound)
	}
	if security.IsProtectedEmail(user.Email) {
		if auditErr := s.auditLogService.LogEvent(
			domain.EventProtectedUserAttackService,
			&actorID,
			&targetID,
			nil, nil, nil,
		); auditErr != nil {
			s.log.Error("AdminUserService.ensureNotProtected: audit error: %v", auditErr)
		}
		return domain.ErrProtectedUser
	}
	return nil
}

// CreateUser creates a new user account from admin context.
//
// Validates input strictly. On protected-email attempt, fires
// EventAdminUserCreateRejectedProtected and returns *domain.InvalidInputError.
// On success, persists the user and fires EventAdminUserCreated.
//
// actorID is the admin's user ID (used for audit attribution).
func (s *AdminUserService) CreateUser(actorID int64, fields CreateUserFields) (*domain.User, error) {
	// 1. Email syntax + protected check (before duplicate check, so the audit
	//    event is fired for protected-email attempts even if the email also
	//    happens to be a duplicate).
	addr, parseErr := mail.ParseAddress(fields.Email)
	if parseErr != nil {
		return nil, &domain.InvalidInputError{Field: "email", Message: "is not a valid address", Cause: parseErr}
	}
	emailLower := strings.ToLower(addr.Address)
	if security.IsProtectedEmail(emailLower) {
		if auditErr := s.auditLogService.LogEvent(
			domain.EventAdminUserCreateRejectedProtected,
			&actorID,
			nil,
			nil, nil,
			map[string]interface{}{"attempted_email": emailLower},
		); auditErr != nil {
			s.log.Error("AdminUserService.CreateUser: protected-email audit error: %v", auditErr)
		}
		return nil, &domain.InvalidInputError{Field: "email", Message: "is reserved"}
	}

	// 2. Name length.
	name := strings.TrimSpace(fields.Name)
	if name == "" || len(name) > 100 {
		return nil, &domain.InvalidInputError{Field: "name", Message: "must be 1–100 characters"}
	}

	// 3. Password complexity (reuses validatePassword in the same package).
	if err := validatePassword(fields.Password); err != nil {
		return nil, &domain.InvalidInputError{Field: "password", Message: err.Error(), Cause: err}
	}

	// 4. Role allowlist. Default to athlete if empty.
	role := fields.Role
	if role == "" {
		role = "athlete"
	}
	if role != "athlete" && role != "coach" && role != "admin" {
		return nil, &domain.InvalidInputError{Field: "role", Message: "must be athlete, coach, or admin"}
	}

	// 5. Duplicate email check via repo.
	existing, lookupErr := s.userRepo.GetByEmail(emailLower)
	if lookupErr != nil {
		return nil, fmt.Errorf("CreateUser: lookup existing: %w", lookupErr)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	// 6. Hash password and persist.
	hash, hashErr := auth.HashPassword(fields.Password)
	if hashErr != nil {
		return nil, fmt.Errorf("CreateUser: hash password: %w", hashErr)
	}

	now := time.Now().UTC()
	user := &domain.User{
		Email:        emailLower,
		PasswordHash: hash,
		Name:         name,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if createErr := s.userRepo.Create(user); createErr != nil {
		if isUniqueViolation(createErr) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("CreateUser: persist: %w", createErr)
	}

	// 7. If admin set EmailVerified=true, set it via Update.
	if fields.EmailVerified {
		user.EmailVerified = true
		verifiedAt := time.Now().UTC()
		user.EmailVerifiedAt = &verifiedAt
		if updateErr := s.userRepo.Update(user); updateErr != nil {
			s.log.Error("AdminUserService.CreateUser: post-create email_verified set: %v", updateErr)
		}
	}

	// 8. Audit success.
	emailDomain := emailLower
	if at := strings.LastIndex(emailLower, "@"); at >= 0 {
		emailDomain = emailLower[at+1:]
	}
	if auditErr := s.auditLogService.LogEvent(
		domain.EventAdminUserCreated,
		&actorID,
		&user.ID,
		nil, nil,
		map[string]interface{}{
			"role":           role,
			"email_verified": user.EmailVerified,
			"email_domain":   emailDomain,
		},
	); auditErr != nil {
		s.log.Error("AdminUserService.CreateUser: audit: %v", auditErr)
	}

	return user, nil
}

// isUniqueViolation returns true if err is a database UNIQUE constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "unique constraint") ||
		(strings.Contains(s, "unique") && strings.Contains(s, "constraint")) ||
		strings.Contains(s, "duplicate entry") ||
		strings.Contains(s, "duplicate key")
}

// UpdateProfile applies a partial PATCH to the target user's profile fields.
//
// Defence layers invoked:
//   - L2: ensureNotProtected guard on entry.
//   - L4: DB-trigger error substring caught on Update failure.
//
// Optimistic concurrency: ifMatchUpdatedAt must equal the current row's
// UpdatedAt; a mismatch returns ErrConflict.
//
// Email-change side-effects: resets EmailVerified to false, stores a
// verification token on the user row (via Update), and fires
// SendVerificationEmail. Both email_changed and profile_updated audit events
// are written on success.
func (s *AdminUserService) UpdateProfile(
	actorID, targetID int64,
	fields ProfileUpdateFields,
	ifMatchUpdatedAt time.Time,
) (*domain.User, error) {
	// L2 — protected-user check.
	if err := s.ensureNotProtected(actorID, targetID); err != nil {
		return nil, err
	}

	// Load the current state of the target user. ensureNotProtected already
	// fetched the user once, but we re-fetch here to keep concerns separate and
	// avoid subtle aliasing bugs if the stub/repo caches.
	user, err := s.userRepo.GetByID(targetID)
	if err != nil {
		return nil, fmt.Errorf("UpdateProfile: GetByID(%d): %w", targetID, err)
	}
	if user == nil {
		return nil, fmt.Errorf("UpdateProfile: user %d: %w", targetID, ErrUserNotFound)
	}

	// Optimistic concurrency check.
	if !user.UpdatedAt.Equal(ifMatchUpdatedAt) {
		return nil, domain.ErrConflict
	}

	emailChanged := false
	oldEmail := user.Email

	// Apply only the non-nil fields.
	if fields.Name != nil {
		name := strings.TrimSpace(*fields.Name)
		if name == "" || len(name) > 100 {
			return nil, &domain.InvalidInputError{Field: "name", Message: "must be 1–100 characters"}
		}
		user.Name = name
	}

	if fields.Email != nil {
		addr, parseErr := mail.ParseAddress(*fields.Email)
		if parseErr != nil {
			return nil, &domain.InvalidInputError{Field: "email", Message: "is not a valid address", Cause: parseErr}
		}
		if addr.Address != user.Email {
			user.Email = addr.Address
			user.EmailVerified = false
			user.EmailVerifiedAt = nil
			emailChanged = true
		}
	}

	if fields.Birthday != nil {
		if fields.Birthday.After(time.Now()) {
			return nil, &domain.InvalidInputError{Field: "birthday", Message: "must be in the past"}
		}
		user.Birthday = fields.Birthday
	}

	// EmailVerified may be overridden by an admin, but only when the email is
	// not simultaneously changing (in which case we always force-reset it).
	if fields.EmailVerified != nil && !emailChanged {
		user.EmailVerified = *fields.EmailVerified
		if *fields.EmailVerified {
			now := time.Now().UTC()
			user.EmailVerifiedAt = &now
		} else {
			user.EmailVerifiedAt = nil
		}
	}

	user.UpdatedAt = time.Now().UTC()

	if err := s.userRepo.Update(user); err != nil {
		// L4 — catch DB-trigger rejections that slipped past L1/L2.
		if strings.Contains(err.Error(), security.TriggerErrorContract) {
			if auditErr := s.auditLogService.LogEvent(
				domain.EventProtectedUserAttackDB,
				&actorID,
				&targetID,
				nil, nil, nil,
			); auditErr != nil {
				s.log.Error("AdminUserService.UpdateProfile: L4 audit error: %v", auditErr)
			}
			return nil, domain.ErrProtectedUser
		}
		return nil, fmt.Errorf("UpdateProfile: Update: %w", err)
	}

	// Email-change side-effects: store verification token and send email.
	if emailChanged {
		token, genErr := generateVerificationToken()
		if genErr != nil {
			// Non-fatal: log and continue; the profile update already succeeded.
			s.log.Error("AdminUserService.UpdateProfile: generateVerificationToken: %v", genErr)
		} else {
			expiresAt := time.Now().Add(24 * time.Hour)
			user.VerificationToken = &token
			user.VerificationTokenExpiresAt = &expiresAt
			// Best-effort persist of the token; failure is non-fatal.
			if updateErr := s.userRepo.Update(user); updateErr != nil {
				s.log.Error("AdminUserService.UpdateProfile: failed to store verification token: %v", updateErr)
			} else {
				verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.appURL, token)
				if sendErr := s.emailService.SendVerificationEmail(user.Email, verifyURL); sendErr != nil {
					s.log.Error("AdminUserService.UpdateProfile: SendVerificationEmail: %v", sendErr)
				}
			}
		}

		details := map[string]interface{}{"old": oldEmail, "new": user.Email}
		if auditErr := s.auditLogService.LogEvent(
			domain.EventEmailChanged,
			&actorID,
			&targetID,
			nil, nil,
			details,
		); auditErr != nil {
			s.log.Error("AdminUserService.UpdateProfile: email_changed audit: %v", auditErr)
		}
	}

	if auditErr := s.auditLogService.LogEvent(
		domain.EventProfileUpdated,
		&actorID,
		&targetID,
		nil, nil, nil,
	); auditErr != nil {
		s.log.Error("AdminUserService.UpdateProfile: profile_updated audit: %v", auditErr)
	}

	return user, nil
}

// ForcePasswordReset triggers a password-reset flow for a target user:
//
//  1. L2 protected-user check.
//  2. Generate a reset token and persist it on the user row (via Update).
//  3. Send the password-reset email.
//  4. Revoke all the user's refresh tokens.
//  5. Write a password_reset_forced_by_admin audit event.
func (s *AdminUserService) ForcePasswordReset(actorID, targetID int64) error {
	// L2 — protected-user check.
	if err := s.ensureNotProtected(actorID, targetID); err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(targetID)
	if err != nil {
		return fmt.Errorf("ForcePasswordReset: GetByID(%d): %w", targetID, err)
	}
	if user == nil {
		return fmt.Errorf("ForcePasswordReset: user %d: %w", targetID, ErrUserNotFound)
	}

	// Generate and persist the reset token.
	token, err := generateResetToken()
	if err != nil {
		return fmt.Errorf("ForcePasswordReset: generateResetToken: %w", err)
	}
	expiresAt := time.Now().Add(1 * time.Hour)
	user.ResetToken = &token
	user.ResetTokenExpiresAt = &expiresAt
	user.UpdatedAt = time.Now().UTC()

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("ForcePasswordReset: save reset token: %w", err)
	}

	// Build the full reset URL — matches the format used by UserService.RequestPasswordReset
	// so admin-triggered reset emails are indistinguishable from user-self-service ones.
	resetURL := fmt.Sprintf("%s/reset-password/%s", s.appURL, token)
	if err := s.emailService.SendPasswordResetEmail(user.Email, resetURL); err != nil {
		return fmt.Errorf("ForcePasswordReset: SendPasswordResetEmail: %w", err)
	}

	// Revoke all refresh tokens so the user is signed out on all devices.
	if err := s.refreshTokenRepo.RevokeAllForUser(user.ID); err != nil {
		return fmt.Errorf("ForcePasswordReset: RevokeAllForUser: %w", err)
	}

	// Audit the forced reset.
	if auditErr := s.auditLogService.LogEvent(
		domain.EventPasswordResetForcedByAdmin,
		&actorID,
		&targetID,
		nil, nil, nil,
	); auditErr != nil {
		s.log.Error("AdminUserService.ForcePasswordReset: audit error: %v", auditErr)
	}

	return nil
}
