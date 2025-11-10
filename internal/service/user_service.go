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
	ErrUserNotFound               = errors.New("user not found")
	ErrInvalidCredentials         = errors.New("invalid credentials")
	ErrEmailAlreadyExists         = errors.New("email already exists")
	ErrRegistrationClosed         = errors.New("registration is closed")
	ErrInvalidResetToken          = errors.New("invalid or expired reset token")
	ErrResetTokenExpired          = errors.New("reset token has expired")
	ErrInvalidVerificationToken   = errors.New("invalid or expired verification token")
	ErrVerificationTokenExpired   = errors.New("verification token has expired")
	ErrEmailAlreadyVerified       = errors.New("email is already verified")
)

// UserService handles user-related business logic
type UserService struct {
	userRepo          domain.UserRepository
	jwtSecret         string
	jwtExpiration     time.Duration
	allowRegistration bool
	emailService      *email.Service
	appURL            string // Base URL for password reset links
}

// NewUserService creates a new user service
func NewUserService(
	userRepo domain.UserRepository,
	jwtSecret string,
	jwtExpiration time.Duration,
	allowRegistration bool,
	emailService *email.Service,
	appURL string,
) *UserService {
	return &UserService{
		userRepo:          userRepo,
		jwtSecret:         jwtSecret,
		jwtExpiration:     jwtExpiration,
		allowRegistration: allowRegistration,
		emailService:      emailService,
		appURL:            appURL,
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
