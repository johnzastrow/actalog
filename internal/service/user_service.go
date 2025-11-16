package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/auth"
	"github.com/johnzastrow/actalog/pkg/email"
)

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrEmailAlreadyExists       = errors.New("email already exists")
	ErrRegistrationClosed       = errors.New("registration is closed")
	ErrInvalidResetToken        = errors.New("invalid or expired reset token")
	ErrResetTokenExpired        = errors.New("reset token has expired")
	ErrInvalidVerificationToken = errors.New("invalid or expired verification token")
	ErrVerificationTokenExpired = errors.New("verification token has expired")
	ErrEmailAlreadyVerified     = errors.New("email is already verified")
	ErrInvalidRefreshToken      = errors.New("invalid or expired refresh token")
	ErrAccountLocked            = errors.New("account locked due to too many failed login attempts")
	ErrAccountDisabled          = errors.New("account has been disabled by an administrator")
)

// UserService handles user-related business logic
type UserService struct {
	userRepo             domain.UserRepository
	refreshTokenRepo     domain.RefreshTokenRepository
	auditLogService      *AuditLogService
	jwtSecret            string
	jwtExpiration        time.Duration
	refreshTokenDuration time.Duration
	allowRegistration    bool
	emailService         email.EmailService
	jwtSecretKey         string
	appURL               string // Base URL for password reset links
	requireVerification  bool   // Require email verification for new users

	// Security configuration
	maxLoginAttempts  int
	lockoutDuration   time.Duration
}

// NewUserService creates a new user service
func NewUserService(
	userRepo domain.UserRepository,
	refreshTokenRepo domain.RefreshTokenRepository,
	auditLogService *AuditLogService,
	jwtSecret string,
	jwtExpiration time.Duration,
	refreshTokenDuration time.Duration,
	allowRegistration bool,
	emailService email.EmailService,
	appURL string,
	requireVerification bool,
	maxLoginAttempts int,
	lockoutDuration time.Duration,
) *UserService {
	return &UserService{
		userRepo:             userRepo,
		refreshTokenRepo:     refreshTokenRepo,
		auditLogService:      auditLogService,
		jwtSecretKey:         jwtSecret,
		jwtExpiration:        jwtExpiration,
		refreshTokenDuration: refreshTokenDuration,
		allowRegistration:    allowRegistration,
		emailService:         emailService,
		appURL:               appURL,
		requireVerification:  requireVerification,
		maxLoginAttempts:     maxLoginAttempts,
		lockoutDuration:      lockoutDuration,
	}
}

// Register creates a new user account
// First user automatically becomes admin
// After that, registration requires allowRegistration to be true
func (s *UserService) Register(name, email, password string) (*domain.User, string, error) {
	// Basic input validation
	if name == "" {
		return nil, "", fmt.Errorf("name is required")
	}
	if email == "" {
		return nil, "", fmt.Errorf("email is required")
	}
	if password == "" {
		return nil, "", fmt.Errorf("password is required")
	}
	if len(password) < 8 {
		return nil, "", fmt.Errorf("password must be at least 8 characters")
	}
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, "", ErrEmailAlreadyExists
	}

	// Check if this is the first user
	count, err := s.userRepo.Count()
	if err != nil {
		return nil, "", fmt.Errorf("failed to count users: %w", err)
	}

	// If not the first user and registration is closed, deny
	if count > 0 && !s.allowRegistration {
		return nil, "", ErrRegistrationClosed
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	now := time.Now()
	user := &domain.User{
		Email:        email,
		PasswordHash: hashedPassword,
		Name:         name,
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// First user is admin
	if count == 0 {
		user.Role = "admin"
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Generate verification token if email verification is required
	if s.requireVerification && s.emailService != nil {
		verificationToken, err := generateVerificationToken()
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate verification token: %w", err)
		}

		// Set token expiration (24 hours from now)
		expiresAt := time.Now().Add(24 * time.Hour)
		user.VerificationToken = &verificationToken
		user.VerificationTokenExpiresAt = &expiresAt
		user.EmailVerified = false

		// Update user with verification token
		err = s.userRepo.Update(user)
		if err != nil {
			return nil, "", fmt.Errorf("failed to save verification token: %w", err)
		}

		// Send verification email
		verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.appURL, verificationToken)
		err = s.emailService.SendVerificationEmail(user.Email, verifyURL)
		if err != nil {
			// Log error but don't fail registration
			fmt.Printf("warning: failed to send verification email: %v\n", err)
		}
	} else {
		// If email verification is not required, auto-verify the user
		user.EmailVerified = true
		verifiedAt := time.Now()
		user.EmailVerifiedAt = &verifiedAt
		err = s.userRepo.Update(user)
		if err != nil {
			return nil, "", fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email, user.Role, s.jwtSecretKey, s.jwtExpiration)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Note: return user with PasswordHash set for tests that validate hashing

	return user, token, nil
}

// Login authenticates a user and returns a JWT token
func (s *UserService) Login(email, password string) (*domain.User, string, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, "", ErrInvalidCredentials
	}

	// Check if account is manually disabled
	if user.AccountDisabled {
		return nil, "", ErrAccountDisabled
	}

	// Check if account is locked (also handles auto-unlock if expired)
	isLocked, _, err := s.userRepo.IsAccountLocked(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check lock status: %w", err)
	}
	if isLocked {
		return nil, "", ErrAccountLocked
	}

	// Check password
	err = auth.CheckPassword(user.PasswordHash, password)
	if err != nil {
		// Password is incorrect - increment failed attempts
		if incrementErr := s.userRepo.IncrementFailedAttempts(user.ID); incrementErr != nil {
			fmt.Printf("warning: failed to increment failed attempts: %v\n", incrementErr)
		}

		// Get updated user to check attempts
		user, getErr := s.userRepo.GetByEmail(email)
		if getErr == nil && user != nil {
			// Check if we should lock the account
			if user.FailedLoginAttempts >= s.maxLoginAttempts {
				// Lock the account
				if lockErr := s.userRepo.LockAccount(user.ID, s.lockoutDuration); lockErr != nil {
					fmt.Printf("warning: failed to lock account: %v\n", lockErr)
				} else {
					// Log the account lockout event
					if s.auditLogService != nil {
						s.auditLogService.LogAccountLocked(user.ID, user.Email, "", "", user.FailedLoginAttempts)
					}
				}
			} else {
				// Log failed login attempt
				if s.auditLogService != nil {
					attemptsRemaining := s.maxLoginAttempts - user.FailedLoginAttempts
					s.auditLogService.LogLoginFailed(email, "invalid_password", "", "", attemptsRemaining)
				}
			}
		}

		return nil, "", ErrInvalidCredentials
	}

	// Password is correct - reset failed attempts
	if resetErr := s.userRepo.ResetFailedAttempts(user.ID); resetErr != nil {
		fmt.Printf("warning: failed to reset failed attempts: %v\n", resetErr)
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	err = s.userRepo.Update(user)
	if err != nil {
		// Log error but don't fail login
		fmt.Printf("warning: failed to update last login: %v\n", err)
	}

	// Log successful login
	if s.auditLogService != nil {
		s.auditLogService.LogLoginSuccess(user.ID, "", "")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email, user.Role, s.jwtSecretKey, s.jwtExpiration)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Don't return password hash or sensitive security fields
	user.PasswordHash = ""
	user.FailedLoginAttempts = 0
	user.LockedAt = nil
	user.LockedUntil = nil

	return user, token, nil
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(id int64) (*domain.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Don't return password hash
	user.PasswordHash = ""

	return user, nil
}

// ValidateToken validates a JWT token and returns user info
func (s *UserService) ValidateToken(tokenString string) (*auth.Claims, error) {
	claims, err := auth.ValidateToken(tokenString, s.jwtSecretKey)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// RequestPasswordReset generates a reset token and sends reset email
func (s *UserService) RequestPasswordReset(email string) error {
	// Get user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Silently succeed if user doesn't exist (security best practice)
	// Don't reveal whether email exists in database
	if user == nil {
		return nil
	}

	// Generate secure random token
	token, err := generateResetToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Set token expiration (1 hour from now)
	expiresAt := time.Now().Add(1 * time.Hour)

	// Update user with reset token
	user.ResetToken = &token
	user.ResetTokenExpiresAt = &expiresAt

	err = s.userRepo.Update(user)
	if err != nil {
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	// Send password reset email
	if s.emailService != nil {
		resetURL := fmt.Sprintf("%s/reset-password/%s", s.appURL, token)
		err = s.emailService.SendPasswordResetEmail(user.Email, resetURL)
		if err != nil {
			return fmt.Errorf("failed to send reset email: %w", err)
		}
	}

	return nil
}

// ResetPassword validates reset token and updates password
func (s *UserService) ResetPassword(token, newPassword string) error {
	// Get user by reset token
	user, err := s.userRepo.GetByResetToken(token)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrInvalidResetToken
	}

	// Check if token is expired
	if user.ResetTokenExpiresAt == nil || time.Now().After(*user.ResetTokenExpiresAt) {
		return ErrResetTokenExpired
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password and clear reset token
	user.PasswordHash = hashedPassword
	user.ResetToken = nil
	user.ResetTokenExpiresAt = nil

	err = s.userRepo.Update(user)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// generateResetToken generates a cryptographically secure random token
func generateResetToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateVerificationToken generates a cryptographically secure random token for email verification
func generateVerificationToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// VerifyEmail validates verification token and marks email as verified
func (s *UserService) VerifyEmail(token string) error {
	// Get user by verification token
	user, err := s.userRepo.GetByVerificationToken(token)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrInvalidVerificationToken
	}

	// Check if already verified
	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}

	// Check if token is expired
	if user.VerificationTokenExpiresAt == nil || time.Now().After(*user.VerificationTokenExpiresAt) {
		return ErrVerificationTokenExpired
	}

	// Mark email as verified and clear verification token
	user.EmailVerified = true
	now := time.Now()
	user.EmailVerifiedAt = &now
	user.VerificationToken = nil
	user.VerificationTokenExpiresAt = nil

	err = s.userRepo.Update(user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// ChangePassword changes a user's password after validating the old password
func (s *UserService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Validate old password
	if err := auth.CheckPassword(user.PasswordHash, oldPassword); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	return s.userRepo.UpdatePassword(userID, hashedPassword)
}

// ResendVerificationEmail resends verification email to a user
func (s *UserService) ResendVerificationEmail(email string) error {
	// Get user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Silently succeed if user doesn't exist (security best practice)
	if user == nil {
		return nil
	}

	// Return error if already verified
	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}

	// Generate new verification token
	verificationToken, err := generateVerificationToken()
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	// Set token expiration (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour)
	user.VerificationToken = &verificationToken
	user.VerificationTokenExpiresAt = &expiresAt

	// Update user with new verification token
	err = s.userRepo.Update(user)
	if err != nil {
		return fmt.Errorf("failed to save verification token: %w", err)
	}

	// Send verification email
	if s.emailService != nil {
		verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.appURL, verificationToken)
		err = s.emailService.SendVerificationEmail(user.Email, verifyURL)
		if err != nil {
			return fmt.Errorf("failed to send verification email: %w", err)
		}
	}

	return nil
}

// CreateRefreshToken creates a new refresh token for a user
func (s *UserService) CreateRefreshToken(userID int64, deviceInfo string) (string, error) {
	// Generate secure random token
	tokenStr, err := generateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Create refresh token record
	refreshToken := &domain.RefreshToken{
		UserID:     userID,
		Token:      tokenStr,
		ExpiresAt:  time.Now().Add(s.refreshTokenDuration),
		CreatedAt:  time.Now(),
		DeviceInfo: deviceInfo,
	}

	err = s.refreshTokenRepo.Create(refreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return tokenStr, nil
}

// RefreshAccessToken validates refresh token and generates new access token
func (s *UserService) RefreshAccessToken(refreshTokenStr string) (*domain.User, string, error) {
	// Get refresh token from database
	refreshToken, err := s.refreshTokenRepo.GetByToken(refreshTokenStr)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get refresh token: %w", err)
	}
	if refreshToken == nil {
		return nil, "", ErrInvalidRefreshToken
	}

	// Get user
	user, err := s.userRepo.GetByID(refreshToken.UserID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, "", ErrUserNotFound
	}

	// Generate new JWT access token
	token, err := auth.GenerateToken(user.ID, user.Email, user.Role, s.jwtSecretKey, s.jwtExpiration)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	err = s.userRepo.Update(user)
	if err != nil {
		// Non-critical error, log but don't fail
		fmt.Printf("Warning: failed to update last login: %v\n", err)
	}

	return user, token, nil
}

// RevokeRefreshToken revokes a specific refresh token
func (s *UserService) RevokeRefreshToken(tokenStr string) error {
	refreshToken, err := s.refreshTokenRepo.GetByToken(tokenStr)
	if err != nil {
		return fmt.Errorf("failed to get refresh token: %w", err)
	}
	if refreshToken == nil {
		return ErrInvalidRefreshToken
	}

	err = s.refreshTokenRepo.Revoke(refreshToken.ID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// RevokeAllRefreshTokens revokes all refresh tokens for a user (logout all devices)
func (s *UserService) RevokeAllRefreshTokens(userID int64) error {
	err := s.refreshTokenRepo.RevokeAllForUser(userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all refresh tokens: %w", err)
	}
	return nil
}

// GetUserRefreshTokens gets all active refresh tokens for a user
func (s *UserService) GetUserRefreshTokens(userID int64) ([]*domain.RefreshToken, error) {
	tokens, err := s.refreshTokenRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user refresh tokens: %w", err)
	}
	return tokens, nil
}

// UpdateProfile updates user profile information
func (s *UserService) UpdateProfile(userID int64, name, email string, birthday *time.Time) (*domain.User, error) {
	// Get current user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Check if email is being changed
	if email != "" && email != user.Email {
		// Check if new email already exists
		existingUser, err := s.userRepo.GetByEmail(email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		if existingUser != nil && existingUser.ID != userID {
			return nil, ErrEmailAlreadyExists
		}

		// Update email and mark as unverified
		user.Email = email
		user.EmailVerified = false
		user.EmailVerifiedAt = nil

		// Send new verification email
		if s.emailService != nil {
			err = s.ResendVerificationEmail(email)
			if err != nil {
				// Log error but don't fail the update
				fmt.Printf("Warning: failed to send verification email: %v\n", err)
			}
		}
	}

	// Update name if provided
	if name != "" {
		user.Name = name
	}

	// Update birthday if provided
	user.Birthday = birthday

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Save changes
	err = s.userRepo.Update(user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Don't return password hash
	user.PasswordHash = ""

	return user, nil
}

// UpdateAvatar updates the user's avatar URL
func (s *UserService) UpdateAvatar(userID int64, avatarURL string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if avatarURL == "" {
		user.ProfileImage = nil
	} else {
		user.ProfileImage = &avatarURL
	}

	user.UpdatedAt = time.Now()
	return s.userRepo.Update(user)
}

// generateRefreshToken generates a cryptographically secure random token
func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Admin Account Security Methods

// UnlockAccount unlocks a user account (admin operation)
func (s *UserService) UnlockAccount(adminUserID, targetUserID int64) error {
	// Get both users for audit logging
	admin, err := s.userRepo.GetByID(adminUserID)
	if err != nil {
		return fmt.Errorf("failed to get admin user: %w", err)
	}

	target, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return fmt.Errorf("failed to get target user: %w", err)
	}

	// Unlock the account
	if err := s.userRepo.UnlockAccount(targetUserID); err != nil {
		return fmt.Errorf("failed to unlock account: %w", err)
	}

	// Log the unlock event
	if s.auditLogService != nil {
		s.auditLogService.LogAccountUnlocked(adminUserID, targetUserID, admin.Email, target.Email)
	}

	return nil
}

// DisableAccount permanently disables a user account (admin operation)
func (s *UserService) DisableAccount(adminUserID, targetUserID int64, reason string) error {
	// Get both users for audit logging
	admin, err := s.userRepo.GetByID(adminUserID)
	if err != nil {
		return fmt.Errorf("failed to get admin user: %w", err)
	}

	target, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return fmt.Errorf("failed to get target user: %w", err)
	}

	// Don't allow disabling self
	if adminUserID == targetUserID {
		return fmt.Errorf("cannot disable your own account")
	}

	// Disable the account with reason
	if err := s.userRepo.DisableAccount(targetUserID, adminUserID, reason); err != nil {
		return fmt.Errorf("failed to disable account: %w", err)
	}

	// Log the disable event
	if s.auditLogService != nil {
		s.auditLogService.LogAccountDisabled(adminUserID, targetUserID, admin.Email, target.Email, reason)
	}

	return nil
}

// EnableAccount re-enables a disabled user account (admin operation)
func (s *UserService) EnableAccount(adminUserID, targetUserID int64) error {
	// Get both users for audit logging
	admin, err := s.userRepo.GetByID(adminUserID)
	if err != nil {
		return fmt.Errorf("failed to get admin user: %w", err)
	}

	target, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return fmt.Errorf("failed to get target user: %w", err)
	}

	// Enable the account
	if err := s.userRepo.EnableAccount(targetUserID); err != nil {
		return fmt.Errorf("failed to enable account: %w", err)
	}

	// Log the enable event
	if s.auditLogService != nil {
		s.auditLogService.LogAccountEnabled(adminUserID, targetUserID, admin.Email, target.Email)
	}

	return nil
}

// ChangeUserRole changes a user's role (admin operation)
func (s *UserService) ChangeUserRole(adminUserID, targetUserID int64, newRole string) error {
	// Validate role
	if newRole != "user" && newRole != "admin" {
		return fmt.Errorf("invalid role: must be 'user' or 'admin'")
	}

	// Get both users for audit logging
	admin, err := s.userRepo.GetByID(adminUserID)
	if err != nil {
		return fmt.Errorf("failed to get admin user: %w", err)
	}

	target, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return fmt.Errorf("failed to get target user: %w", err)
	}

	oldRole := target.Role

	// Don't allow changing your own role
	if adminUserID == targetUserID {
		return fmt.Errorf("cannot change your own role")
	}

	// Update the role
	target.Role = newRole
	target.UpdatedAt = time.Now()
	if err := s.userRepo.Update(target); err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	// Log the role change
	if s.auditLogService != nil {
		s.auditLogService.LogRoleChanged(adminUserID, targetUserID, admin.Email, target.Email, oldRole, newRole)
	}

	return nil
}

// ListUsers returns a paginated list of all users (admin operation)
func (s *UserService) ListUsers(limit, offset int) ([]*domain.User, int64, error) {
	// Validate pagination
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	users, err := s.userRepo.List(limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	count, err := s.userRepo.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Don't return password hashes
	for _, user := range users {
		user.PasswordHash = ""
	}

	return users, count, nil
}

// SetEmailVerification sets email verification status (admin operation)
func (s *UserService) SetEmailVerification(adminUserID, targetUserID int64, verified bool) error {
	// Get both users for audit logging
	admin, err := s.userRepo.GetByID(adminUserID)
	if err != nil {
		return fmt.Errorf("failed to get admin user: %w", err)
	}

	target, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return fmt.Errorf("failed to get target user: %w", err)
	}

	// Update email verification status
	if verified {
		now := time.Now()
		target.EmailVerified = true
		target.EmailVerifiedAt = &now
	} else {
		target.EmailVerified = false
		target.EmailVerifiedAt = nil
	}

	target.UpdatedAt = time.Now()
	if err := s.userRepo.Update(target); err != nil {
		return fmt.Errorf("failed to update email verification: %w", err)
	}

	// Log the verification change
	if s.auditLogService != nil {
		eventType := "email_verified"
		if !verified {
			eventType = "email_unverified"
		}
		details := map[string]interface{}{
			"admin_email":  admin.Email,
			"target_email": target.Email,
			"verified":     verified,
		}
		s.auditLogService.LogEvent(eventType, &adminUserID, &targetUserID, nil, nil, details)
	}

	return nil
}

// GetUserByIDWithAdminDetails returns a user with all admin-visible details
func (s *UserService) GetUserByIDWithAdminDetails(userID int64) (*domain.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Don't return password hash
	user.PasswordHash = ""

	return user, nil
}

// DeleteUser permanently deletes a user account (admin action)
func (s *UserService) DeleteUser(adminUserID int64, targetUserID int64) error {
	// Verify the admin user exists and is an admin
	adminUser, err := s.userRepo.GetByID(adminUserID)
	if err != nil {
		return fmt.Errorf("failed to get admin user: %w", err)
	}
	if adminUser == nil {
		return ErrUserNotFound
	}
	if adminUser.Role != "admin" {
		return fmt.Errorf("only admins can delete users")
	}

	// Get the target user
	targetUser, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return fmt.Errorf("failed to get target user: %w", err)
	}
	if targetUser == nil {
		return ErrUserNotFound
	}

	// Prevent admin from deleting themselves
	if adminUserID == targetUserID {
		return fmt.Errorf("cannot delete your own account")
	}

	// Delete the user (this will cascade delete related data based on DB constraints)
	if err := s.userRepo.Delete(targetUserID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Log the deletion
	if s.auditLogService != nil {
		details := map[string]interface{}{
			"admin_email":  adminUser.Email,
			"target_email": targetUser.Email,
			"target_id":    targetUserID,
		}
		s.auditLogService.LogEvent("user_deleted", &adminUserID, &targetUserID, nil, nil, details)
	}

	return nil
}
