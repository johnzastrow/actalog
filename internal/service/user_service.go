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
)

// UserService handles user-related business logic
type UserService struct {
	userRepo             domain.UserRepository
	refreshTokenRepo     domain.RefreshTokenRepository
	jwtSecret            string
	jwtExpiration        time.Duration
	refreshTokenDuration time.Duration
	allowRegistration    bool
	emailService         *email.Service
	appURL               string // Base URL for password reset links
}

// NewUserService creates a new user service
func NewUserService(
	userRepo domain.UserRepository,
	refreshTokenRepo domain.RefreshTokenRepository,
	jwtSecret string,
	jwtExpiration time.Duration,
	refreshTokenDuration time.Duration,
	allowRegistration bool,
	emailService *email.Service,
	appURL string,
) *UserService {
	return &UserService{
		userRepo:             userRepo,
		refreshTokenRepo:     refreshTokenRepo,
		jwtSecret:            jwtSecret,
		jwtExpiration:        jwtExpiration,
		refreshTokenDuration: refreshTokenDuration,
		allowRegistration:    allowRegistration,
		emailService:         emailService,
		appURL:               appURL,
	}
}

// Register creates a new user account
// First user automatically becomes admin
// After that, registration requires allowRegistration to be true
func (s *UserService) Register(name, email, password string) (*domain.User, string, error) {
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

	// Generate verification token if email service is enabled
	if s.emailService != nil {
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
		// If email service is not enabled, auto-verify the user
		user.EmailVerified = true
		verifiedAt := time.Now()
		user.EmailVerifiedAt = &verifiedAt
		err = s.userRepo.Update(user)
		if err != nil {
			return nil, "", fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email, user.Role, s.jwtSecret, s.jwtExpiration)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Don't return password hash
	user.PasswordHash = ""

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

	// Check password
	err = auth.CheckPassword(user.PasswordHash, password)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	err = s.userRepo.Update(user)
	if err != nil {
		// Log error but don't fail login
		fmt.Printf("warning: failed to update last login: %v\n", err)
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email, user.Role, s.jwtSecret, s.jwtExpiration)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Don't return password hash
	user.PasswordHash = ""

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
	claims, err := auth.ValidateToken(tokenString, s.jwtSecret)
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
	token, err := auth.GenerateToken(user.ID, user.Email, user.Role, s.jwtSecret, s.jwtExpiration)
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

// generateRefreshToken generates a cryptographically secure random token
func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
