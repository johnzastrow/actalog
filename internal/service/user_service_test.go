package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/auth"
)

type mockEmailService struct {
	sentEmails []mockEmail
}

type mockEmail struct {
	to      string
	subject string
	body    string
}

func (m *mockEmailService) SendPasswordResetEmail(to, resetURL string) error {
	m.sentEmails = append(m.sentEmails, mockEmail{
		to:      to,
		subject: "Password Reset",
		body:    resetURL,
	})
	return nil
}
func (m *mockEmailService) SendVerificationEmail(to, verifyURL string) error {
	// For tests we just record as a sent email (reuse subject)
	m.sentEmails = append(m.sentEmails, mockEmail{
		to:      to,
		subject: "Verify Email",
		body:    verifyURL,
	})
	return nil
}

func (m *mockEmailService) SendHTMLEmail(to, subject, htmlBody string) error {
	m.sentEmails = append(m.sentEmails, mockEmail{
		to:      to,
		subject: subject,
		body:    htmlBody,
	})
	return nil
}

func (m *mockEmailService) SendWelcomeEmail(to, userName, appURL string) error {
	m.sentEmails = append(m.sentEmails, mockEmail{
		to:      to,
		subject: "Welcome",
		body:    "Welcome " + userName,
	})
	return nil
}

// Mock refresh token repository
type mockRefreshTokenRepo struct {
	tokens map[string]*domain.RefreshToken
	nextID int64
}

func (m *mockRefreshTokenRepo) Create(token *domain.RefreshToken) error {
	if m.tokens == nil {
		m.tokens = make(map[string]*domain.RefreshToken)
	}
	m.nextID++
	token.ID = m.nextID
	m.tokens[token.Token] = token
	return nil
}

func (m *mockRefreshTokenRepo) GetByToken(token string) (*domain.RefreshToken, error) {
	if t, ok := m.tokens[token]; ok {
		return t, nil
	}
	return nil, nil // Return nil, nil for not found (service checks for nil)
}

func (m *mockRefreshTokenRepo) Revoke(tokenID int64) error {
	// Find token by ID and remove
	for k, t := range m.tokens {
		if t.ID == tokenID {
			delete(m.tokens, k)
			return nil
		}
	}
	return nil
}

func (m *mockRefreshTokenRepo) RevokeAllForUser(userID int64) error {
	for k, t := range m.tokens {
		if t.UserID == userID {
			delete(m.tokens, k)
		}
	}
	return nil
}

func (m *mockRefreshTokenRepo) DeleteExpired() error {
	now := time.Now()
	for k, t := range m.tokens {
		if t.ExpiresAt.Before(now) {
			delete(m.tokens, k)
		}
	}
	return nil
}

func (m *mockRefreshTokenRepo) Delete(tokenID int64) error {
	for k, t := range m.tokens {
		if t.ID == tokenID {
			delete(m.tokens, k)
			return nil
		}
	}
	return nil
}
func (m *mockRefreshTokenRepo) GetByUserID(userID int64) ([]*domain.RefreshToken, error) {
	var out []*domain.RefreshToken
	for _, t := range m.tokens {
		if t.UserID == userID {
			out = append(out, t)
		}
	}
	return out, nil
}

// Helper to create test user service
func newTestUserService(allowRegistration bool) *UserService {
	return newTestUserServiceWithRepo(
		&mockUserRepo{users: make(map[int64]*domain.User), nextID: 0},
		allowRegistration,
	)
}

// Helper to create test user service with a custom mock repo (for error injection)
func newTestUserServiceWithRepo(repo *mockUserRepo, allowRegistration bool) *UserService {
	return NewUserService(
		repo,
		&mockRefreshTokenRepo{tokens: make(map[string]*domain.RefreshToken)},
		nil, // no user subscription repo for tests
		nil, // no audit log service for tests
		"test-secret-key",
		24*time.Hour,
		7*24*time.Hour, // 7 days refresh token duration
		allowRegistration,
		&mockEmailService{},
		"http://localhost:3000",
		false,          // Don't require email verification in tests
		5,              // max login attempts
		15*time.Minute, // lockout duration
	)
}

// Test User Registration
func TestRegister(t *testing.T) {
	service := newTestUserService(true)

	tests := []struct {
		name        string
		userName    string
		email       string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid registration",
			userName:    "John Doe",
			email:       "john@example.com",
			password:    "SecurePass123",
			expectError: false,
		},
		{
			name:        "Empty name",
			userName:    "",
			email:       "test@example.com",
			password:    "SecurePass123",
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name:        "Empty email",
			userName:    "John Doe",
			email:       "",
			password:    "SecurePass123",
			expectError: true,
			errorMsg:    "email is required",
		},
		{
			name:        "Empty password",
			userName:    "John Doe",
			email:       "john@example.com",
			password:    "",
			expectError: true,
			errorMsg:    "password is required",
		},
		{
			name:        "Short password",
			userName:    "John Doe",
			email:       "john@example.com",
			password:    "short",
			expectError: true,
			errorMsg:    "password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, token, err := service.Register(tt.userName, tt.email, tt.password)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if user == nil {
				t.Fatal("Expected user, got nil")
			}

			if token == "" {
				t.Error("Expected JWT token, got empty string")
			}

			if user.Email != tt.email {
				t.Errorf("Expected email %s, got %s", tt.email, user.Email)
			}

			if user.Name != tt.userName {
				t.Errorf("Expected name %s, got %s", tt.userName, user.Name)
			}

			// Verify password is hashed
			if user.PasswordHash == tt.password {
				t.Error("Password should be hashed, not stored in plain text")
			}
		})
	}
}

// Test First User Becomes Admin
func TestFirstUserBecomesAdmin(t *testing.T) {
	service := newTestUserService(true)

	// Register first user
	user1, _, err := service.Register("Admin User", "admin@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register first user: %v", err)
	}

	if user1.Role != "admin" {
		t.Errorf("First user should be admin, got role: %s", user1.Role)
	}

	// Register second user
	user2, _, err := service.Register("Regular User", "user@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register second user: %v", err)
	}

	if user2.Role != "athlete" {
		t.Errorf("Second user should be regular athlete, got role: %s", user2.Role)
	}
}

// Test Duplicate Email Registration
func TestDuplicateEmailRegistration(t *testing.T) {
	service := newTestUserService(true)

	// Register first user
	_, _, err := service.Register("User One", "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register first user: %v", err)
	}

	// Try to register with same email
	_, _, err = service.Register("User Two", "test@example.com", "Password123")
	if err != ErrEmailAlreadyExists {
		t.Errorf("Expected ErrEmailAlreadyExists, got: %v", err)
	}
}

// Test Registration Closed
func TestRegistrationClosed(t *testing.T) {
	// First user (admin) can register
	service := newTestUserService(false)

	user1, _, err := service.Register("Admin", "admin@example.com", "Password123")
	if err != nil {
		t.Fatalf("First user should be able to register: %v", err)
	}

	if user1.Role != "admin" {
		t.Error("First user should be admin")
	}

	// Second user cannot register when registration is closed
	_, _, err = service.Register("User", "user@example.com", "Password123")
	if err != ErrRegistrationClosed {
		t.Errorf("Expected ErrRegistrationClosed, got: %v", err)
	}
}

// Test User Login
func TestLogin(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	email := "test@example.com"
	password := "Password123"
	_, _, err := service.Register("Test User", email, password)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	tests := []struct {
		name        string
		email       string
		password    string
		expectError bool
	}{
		{
			name:        "Valid credentials",
			email:       email,
			password:    password,
			expectError: false,
		},
		{
			name:        "Invalid email",
			email:       "wrong@example.com",
			password:    password,
			expectError: true,
		},
		{
			name:        "Invalid password",
			email:       email,
			password:    "WrongPassword",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, token, err := service.Login(tt.email, tt.password)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if user == nil {
				t.Fatal("Expected user, got nil")
			}

			if token == "" {
				t.Error("Expected JWT token, got empty string")
			}

			// Verify last login time was updated
			if user.LastLoginAt == nil {
				t.Error("Last login time should be set")
			}
		})
	}
}

// Test Password Hashing
func TestPasswordHashing(t *testing.T) {
	service := newTestUserService(true)

	password := "TestPassword123"
	user, _, err := service.Register("Test User", "test@example.com", password)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Password should be hashed
	if user.PasswordHash == password {
		t.Error("Password should be hashed")
	}

	// Should be able to verify password
	err = auth.CheckPassword(user.PasswordHash, password)
	if err != nil {
		t.Error("Password verification failed for correct password")
	}

	// Should fail for wrong password
	err = auth.CheckPassword(user.PasswordHash, "WrongPassword")
	if err == nil {
		t.Error("Password verification should fail for wrong password")
	}
}

// Test Generate Password Reset Token
func TestGeneratePasswordResetToken(t *testing.T) {
	service := newTestUserService(true)
	emailService := service.emailService.(*mockEmailService)

	// Register a user
	email := "test@example.com"
	_, _, err := service.Register("Test User", email, "Password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Request password reset
	err = service.RequestPasswordReset(email)
	if err != nil {
		t.Fatalf("Failed to generate reset token: %v", err)
	}

	// Verify at least one email was sent (verification + reset possible)
	if len(emailService.sentEmails) == 0 {
		t.Errorf("Expected at least 1 email sent, got %d", len(emailService.sentEmails))
	}

	// Verify user has reset token
	user, err := service.userRepo.GetByEmail(email)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if user.ResetToken == nil {
		t.Error("User should have reset token")
	}

	if user.ResetTokenExpiresAt == nil {
		t.Error("Reset token should have expiration")
	}

	// Token should expire in the future
	if user.ResetTokenExpiresAt.Before(time.Now()) {
		t.Error("Reset token expiration should be in the future")
	}
}

// Test Reset Password
func TestResetPassword(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	email := "test@example.com"
	oldPassword := "OldPassword123"
	_, _, err := service.Register("Test User", email, oldPassword)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Generate reset token
	err = service.RequestPasswordReset(email)
	if err != nil {
		t.Fatalf("Failed to generate reset token: %v", err)
	}

	// Get the token
	user, _ := service.userRepo.GetByEmail(email)
	token := *user.ResetToken

	// Reset password
	newPassword := "NewPassword123"
	err = service.ResetPassword(token, newPassword)
	if err != nil {
		t.Fatalf("Failed to reset password: %v", err)
	}

	// Try to login with new password
	_, _, err = service.Login(email, newPassword)
	if err != nil {
		t.Error("Should be able to login with new password")
	}

	// Old password should not work
	_, _, err = service.Login(email, oldPassword)
	if err == nil {
		t.Error("Old password should not work after reset")
	}

	// Token should be cleared
	user, _ = service.userRepo.GetByEmail(email)
	if user.ResetToken != nil {
		t.Error("Reset token should be cleared after use")
	}
}

// Test Invalid Reset Token
func TestInvalidResetToken(t *testing.T) {
	service := newTestUserService(true)

	err := service.ResetPassword("invalid-token", "NewPassword123")
	if err != ErrInvalidResetToken {
		t.Errorf("Expected ErrInvalidResetToken, got: %v", err)
	}
}

// Test Expired Reset Token
func TestExpiredResetToken(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	email := "test@example.com"
	_, _, err := service.Register("Test User", email, "Password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Generate reset token and immediately expire it
	user, _ := service.userRepo.GetByEmail(email)
	token := "test-token"
	expiredTime := time.Now().Add(-1 * time.Hour)
	user.ResetToken = &token
	user.ResetTokenExpiresAt = &expiredTime
	service.userRepo.Update(user)

	// Try to reset with expired token
	err = service.ResetPassword(token, "NewPassword123")
	if err != ErrResetTokenExpired {
		t.Errorf("Expected ErrResetTokenExpired, got: %v", err)
	}
}

// Test JWT Token Generation
func TestJWTTokenGeneration(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	user, token, err := service.Register("Test User", "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Token should not be empty
	if token == "" {
		t.Fatal("JWT token should not be empty")
	}

	// Token should be a valid JWT format (3 parts separated by dots)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("JWT should have 3 parts, got %d", len(parts))
	}

	// Verify we can validate the token (basic check)
	claims, err := auth.ValidateToken(token, service.jwtSecretKey)
	if err != nil {
		t.Fatalf("Failed to validate JWT token: %v", err)
	}

	// Verify claims
	if claims.UserID != user.ID {
		t.Errorf("Expected UserID %d, got %d", user.ID, claims.UserID)
	}

	if claims.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, claims.Email)
	}

	if claims.Role != user.Role {
		t.Errorf("Expected Role %s, got %s", user.Role, claims.Role)
	}
}

// Test GetByID
func TestGetByID(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	user, _, err := service.Register("Test User", "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Get user by ID
	retrieved, err := service.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if retrieved.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, retrieved.Email)
	}

	// Password hash should be cleared
	if retrieved.PasswordHash != "" {
		t.Error("Password hash should be cleared")
	}

	// Get non-existent user
	_, err = service.GetByID(99999)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// Test ValidateToken
func TestValidateToken(t *testing.T) {
	service := newTestUserService(true)

	// Register a user and get token
	user, token, err := service.Register("Test User", "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Validate valid token
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected UserID %d, got %d", user.ID, claims.UserID)
	}

	// Validate invalid token
	_, err = service.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

// Test VerifyEmail
func TestVerifyEmail(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	_, _, err := service.Register("Test User", "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Set up a verification token manually
	user, _ := service.userRepo.GetByEmail("test@example.com")
	token := "test-verification-token"
	expiresAt := time.Now().Add(24 * time.Hour)
	user.VerificationToken = &token
	user.VerificationTokenExpiresAt = &expiresAt
	user.EmailVerified = false
	service.userRepo.Update(user)

	// Verify email
	err = service.VerifyEmail(token)
	if err != nil {
		t.Fatalf("Failed to verify email: %v", err)
	}

	// Check user is verified
	user, _ = service.userRepo.GetByEmail("test@example.com")
	if !user.EmailVerified {
		t.Error("Email should be verified")
	}

	// Try to verify again (already verified)
	user.VerificationToken = &token
	user.VerificationTokenExpiresAt = &expiresAt
	service.userRepo.Update(user)
	err = service.VerifyEmail(token)
	if err != ErrEmailAlreadyVerified {
		t.Errorf("Expected ErrEmailAlreadyVerified, got: %v", err)
	}

	// Test invalid token
	err = service.VerifyEmail("invalid-token")
	if err != ErrInvalidVerificationToken {
		t.Errorf("Expected ErrInvalidVerificationToken, got: %v", err)
	}

	// Test expired token
	user.EmailVerified = false
	expiredTime := time.Now().Add(-1 * time.Hour)
	user.VerificationTokenExpiresAt = &expiredTime
	service.userRepo.Update(user)
	err = service.VerifyEmail(token)
	if err != ErrVerificationTokenExpired {
		t.Errorf("Expected ErrVerificationTokenExpired, got: %v", err)
	}
}

// Test ChangePassword
func TestChangePassword(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	user, _, err := service.Register("Test User", "test@example.com", "OldPassword123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Change password
	err = service.ChangePassword(user.ID, "OldPassword123", "NewPassword123")
	if err != nil {
		t.Fatalf("Failed to change password: %v", err)
	}

	// Login with new password
	_, _, err = service.Login("test@example.com", "NewPassword123")
	if err != nil {
		t.Error("Should be able to login with new password")
	}

	// Old password should not work
	_, _, err = service.Login("test@example.com", "OldPassword123")
	if err == nil {
		t.Error("Old password should not work")
	}

	// Wrong old password should fail
	err = service.ChangePassword(user.ID, "WrongPassword", "AnotherPassword")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got: %v", err)
	}

	// Non-existent user should fail
	err = service.ChangePassword(99999, "OldPassword", "NewPassword")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// Test ResendVerificationEmail
func TestResendVerificationEmail(t *testing.T) {
	service := newTestUserService(true)
	emailService := service.emailService.(*mockEmailService)

	// Register a user
	_, _, err := service.Register("Test User", "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Set up as unverified
	user, _ := service.userRepo.GetByEmail("test@example.com")
	user.EmailVerified = false
	service.userRepo.Update(user)

	initialEmailCount := len(emailService.sentEmails)

	// Resend verification email
	err = service.ResendVerificationEmail("test@example.com")
	if err != nil {
		t.Fatalf("Failed to resend verification email: %v", err)
	}

	// Check email was sent
	if len(emailService.sentEmails) <= initialEmailCount {
		t.Error("Verification email should have been sent")
	}

	// Non-existent email should silently succeed (security)
	err = service.ResendVerificationEmail("nonexistent@example.com")
	if err != nil {
		t.Errorf("Should silently succeed for non-existent email, got: %v", err)
	}

	// Already verified should fail
	user.EmailVerified = true
	service.userRepo.Update(user)
	err = service.ResendVerificationEmail("test@example.com")
	if err != ErrEmailAlreadyVerified {
		t.Errorf("Expected ErrEmailAlreadyVerified, got: %v", err)
	}
}

// Test CreateRefreshToken
func TestCreateRefreshToken(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	user, _, err := service.Register("Test User", "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Create refresh token
	tokenStr, err := service.CreateRefreshToken(user.ID, "Test Device", false)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	if tokenStr == "" {
		t.Error("Refresh token should not be empty")
	}

	// Create with rememberMe
	tokenStr2, err := service.CreateRefreshToken(user.ID, "Test Device 2", true)
	if err != nil {
		t.Fatalf("Failed to create refresh token with rememberMe: %v", err)
	}

	if tokenStr2 == "" {
		t.Error("Refresh token should not be empty")
	}
}

// Test RefreshAccessToken
func TestRefreshAccessToken(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	user, _, err := service.Register("Test User", "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Create refresh token
	refreshTokenStr, err := service.CreateRefreshToken(user.ID, "Test Device", false)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Refresh access token
	returnedUser, newToken, err := service.RefreshAccessToken(refreshTokenStr)
	if err != nil {
		t.Fatalf("Failed to refresh access token: %v", err)
	}

	if returnedUser == nil {
		t.Fatal("Expected user, got nil")
	}

	if newToken == "" {
		t.Error("Expected new JWT token")
	}

	// Invalid refresh token
	_, _, err = service.RefreshAccessToken("invalid-token")
	if err != ErrInvalidRefreshToken {
		t.Errorf("Expected ErrInvalidRefreshToken, got: %v", err)
	}
}

// Test RevokeRefreshToken
func TestRevokeRefreshToken(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	user, _, err := service.Register("Test User", "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Create refresh token
	refreshTokenStr, err := service.CreateRefreshToken(user.ID, "Test Device", false)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Revoke the token
	err = service.RevokeRefreshToken(refreshTokenStr)
	if err != nil {
		t.Fatalf("Failed to revoke refresh token: %v", err)
	}

	// Token should no longer work
	_, _, err = service.RefreshAccessToken(refreshTokenStr)
	if err == nil {
		t.Error("Revoked token should not work")
	}

	// Invalid token
	err = service.RevokeRefreshToken("invalid-token")
	if err != ErrInvalidRefreshToken {
		t.Errorf("Expected ErrInvalidRefreshToken, got: %v", err)
	}
}

// Test RevokeAllRefreshTokens
func TestRevokeAllRefreshTokens(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	user, _, err := service.Register("Test User", "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Create multiple refresh tokens
	token1, _ := service.CreateRefreshToken(user.ID, "Device 1", false)
	token2, _ := service.CreateRefreshToken(user.ID, "Device 2", false)

	// Revoke all
	err = service.RevokeAllRefreshTokens(user.ID)
	if err != nil {
		t.Fatalf("Failed to revoke all refresh tokens: %v", err)
	}

	// Both tokens should no longer work
	_, _, err = service.RefreshAccessToken(token1)
	if err == nil {
		t.Error("First token should be revoked")
	}

	_, _, err = service.RefreshAccessToken(token2)
	if err == nil {
		t.Error("Second token should be revoked")
	}
}

// Test GetUserRefreshTokens
func TestGetUserRefreshTokens(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	user, _, err := service.Register("Test User", "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Create multiple refresh tokens
	service.CreateRefreshToken(user.ID, "Device 1", false)
	service.CreateRefreshToken(user.ID, "Device 2", false)

	// Get tokens
	tokens, err := service.GetUserRefreshTokens(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user refresh tokens: %v", err)
	}

	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(tokens))
	}
}

// Test UpdateProfile
func TestUpdateProfile(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	user, _, err := service.Register("Test User", "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Update name only
	updated, err := service.UpdateProfile(user.ID, "New Name", "", nil)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("Expected name 'New Name', got '%s'", updated.Name)
	}

	// Update with birthday
	birthday := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)
	updated, err = service.UpdateProfile(user.ID, "", "", &birthday)
	if err != nil {
		t.Fatalf("Failed to update profile with birthday: %v", err)
	}

	if updated.Birthday == nil || !updated.Birthday.Equal(birthday) {
		t.Error("Birthday should be updated")
	}

	// Update email
	updated, err = service.UpdateProfile(user.ID, "", "newemail@example.com", nil)
	if err != nil {
		t.Fatalf("Failed to update email: %v", err)
	}

	if updated.Email != "newemail@example.com" {
		t.Errorf("Expected email 'newemail@example.com', got '%s'", updated.Email)
	}

	// Email should be unverified after change
	if updated.EmailVerified {
		t.Error("Email should be unverified after change")
	}

	// Non-existent user
	_, err = service.UpdateProfile(99999, "Name", "", nil)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}

	// Duplicate email
	service.Register("Other User", "other@example.com", "Password123")
	_, err = service.UpdateProfile(user.ID, "", "other@example.com", nil)
	if err != ErrEmailAlreadyExists {
		t.Errorf("Expected ErrEmailAlreadyExists, got: %v", err)
	}
}

// Test UpdateAvatar
func TestUpdateAvatar(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	user, _, err := service.Register("Test User", "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Update avatar
	err = service.UpdateAvatar(user.ID, "https://example.com/avatar.jpg")
	if err != nil {
		t.Fatalf("Failed to update avatar: %v", err)
	}

	// Verify avatar is set
	updated, _ := service.GetByID(user.ID)
	if updated.ProfileImage == nil || *updated.ProfileImage != "https://example.com/avatar.jpg" {
		t.Error("Avatar should be set")
	}

	// Clear avatar
	err = service.UpdateAvatar(user.ID, "")
	if err != nil {
		t.Fatalf("Failed to clear avatar: %v", err)
	}

	updated, _ = service.GetByID(user.ID)
	if updated.ProfileImage != nil {
		t.Error("Avatar should be cleared")
	}

	// Non-existent user
	err = service.UpdateAvatar(99999, "https://example.com/avatar.jpg")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// Test ListUsers
func TestListUsers(t *testing.T) {
	service := newTestUserService(true)

	// Register multiple users
	service.Register("User 1", "user1@example.com", "Password123")
	service.Register("User 2", "user2@example.com", "Password123")
	service.Register("User 3", "user3@example.com", "Password123")

	// List users
	users, count, err := service.ListUsers(10, 0)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// Verify password hashes are cleared
	for _, u := range users {
		if u.PasswordHash != "" {
			t.Error("Password hash should be cleared")
		}
	}

	// Test pagination validation (negative offset)
	users, _, err = service.ListUsers(10, -1)
	if err != nil {
		t.Fatalf("Should handle negative offset: %v", err)
	}

	// Test limit validation (too high)
	users, _, err = service.ListUsers(1000, 0)
	if err != nil {
		t.Fatalf("Should handle high limit: %v", err)
	}
}

// Test Admin Account Operations
func TestUnlockAccount(t *testing.T) {
	service := newTestUserService(true)

	// Register admin and regular user
	admin, _, _ := service.Register("Admin", "admin@example.com", "Password123")
	user, _, _ := service.Register("User", "user@example.com", "Password123")

	// Lock the user account manually
	userFromRepo, _ := service.userRepo.GetByEmail("user@example.com")
	service.userRepo.LockAccount(userFromRepo.ID, 15*time.Minute)

	// Verify account is locked
	isLocked, _, _ := service.userRepo.IsAccountLocked(user.ID)
	if !isLocked {
		t.Error("Account should be locked")
	}

	// Unlock as admin
	err := service.UnlockAccount(admin.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to unlock account: %v", err)
	}

	// Verify account is unlocked
	isLocked, _, _ = service.userRepo.IsAccountLocked(user.ID)
	if isLocked {
		t.Error("Account should be unlocked")
	}
}

// Test DisableAccount
func TestDisableAccount(t *testing.T) {
	service := newTestUserService(true)

	// Register admin and regular user
	admin, _, _ := service.Register("Admin", "admin@example.com", "Password123")
	user, _, _ := service.Register("User", "user@example.com", "Password123")

	// Disable user account
	err := service.DisableAccount(admin.ID, user.ID, "Test reason")
	if err != nil {
		t.Fatalf("Failed to disable account: %v", err)
	}

	// Verify account is disabled
	userFromRepo, _ := service.userRepo.GetByID(user.ID)
	if !userFromRepo.AccountDisabled {
		t.Error("Account should be disabled")
	}

	// Try to login - should fail
	_, _, err = service.Login("user@example.com", "Password123")
	if err != ErrAccountDisabled {
		t.Errorf("Expected ErrAccountDisabled, got: %v", err)
	}

	// Cannot disable self
	err = service.DisableAccount(admin.ID, admin.ID, "Self disable")
	if err == nil {
		t.Error("Should not be able to disable own account")
	}
}

// Test EnableAccount
func TestEnableAccount(t *testing.T) {
	service := newTestUserService(true)

	// Register admin and regular user
	admin, _, _ := service.Register("Admin", "admin@example.com", "Password123")
	user, _, _ := service.Register("User", "user@example.com", "Password123")

	// Disable then enable
	service.DisableAccount(admin.ID, user.ID, "Test")
	err := service.EnableAccount(admin.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to enable account: %v", err)
	}

	// Verify account is enabled
	userFromRepo, _ := service.userRepo.GetByID(user.ID)
	if userFromRepo.AccountDisabled {
		t.Error("Account should be enabled")
	}

	// Can login again
	_, _, err = service.Login("user@example.com", "Password123")
	if err != nil {
		t.Error("Should be able to login after enabling")
	}
}

// Test ChangeUserRole
func TestChangeUserRole(t *testing.T) {
	service := newTestUserService(true)

	// Register admin and regular user
	admin, _, _ := service.Register("Admin", "admin@example.com", "Password123")
	user, _, _ := service.Register("User", "user@example.com", "Password123")

	// Change user to admin
	err := service.ChangeUserRole(admin.ID, user.ID, "admin")
	if err != nil {
		t.Fatalf("Failed to change role: %v", err)
	}

	// Verify role changed
	userFromRepo, _ := service.userRepo.GetByID(user.ID)
	if userFromRepo.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", userFromRepo.Role)
	}

	// Change back to athlete
	err = service.ChangeUserRole(admin.ID, user.ID, "athlete")
	if err != nil {
		t.Fatalf("Failed to change role back: %v", err)
	}

	userFromRepo, _ = service.userRepo.GetByID(user.ID)
	if userFromRepo.Role != "athlete" {
		t.Errorf("Expected role 'athlete', got '%s'", userFromRepo.Role)
	}

	// Invalid role
	err = service.ChangeUserRole(admin.ID, user.ID, "invalid")
	if err == nil {
		t.Error("Should fail for invalid role")
	}

	// Cannot change own role
	err = service.ChangeUserRole(admin.ID, admin.ID, "athlete")
	if err == nil {
		t.Error("Should not be able to change own role")
	}
}

// Test SetEmailVerification
func TestSetEmailVerification(t *testing.T) {
	service := newTestUserService(true)

	// Register admin and regular user
	admin, _, _ := service.Register("Admin", "admin@example.com", "Password123")
	user, _, _ := service.Register("User", "user@example.com", "Password123")

	// Set as unverified
	userFromRepo, _ := service.userRepo.GetByEmail("user@example.com")
	userFromRepo.EmailVerified = false
	service.userRepo.Update(userFromRepo)

	// Set verified
	err := service.SetEmailVerification(admin.ID, user.ID, true)
	if err != nil {
		t.Fatalf("Failed to set email verification: %v", err)
	}

	userFromRepo, _ = service.userRepo.GetByID(user.ID)
	if !userFromRepo.EmailVerified {
		t.Error("Email should be verified")
	}
	if userFromRepo.EmailVerifiedAt == nil {
		t.Error("EmailVerifiedAt should be set")
	}

	// Set unverified
	err = service.SetEmailVerification(admin.ID, user.ID, false)
	if err != nil {
		t.Fatalf("Failed to unset email verification: %v", err)
	}

	userFromRepo, _ = service.userRepo.GetByID(user.ID)
	if userFromRepo.EmailVerified {
		t.Error("Email should be unverified")
	}
	if userFromRepo.EmailVerifiedAt != nil {
		t.Error("EmailVerifiedAt should be nil")
	}
}

// Test GetUserByIDWithAdminDetails
func TestGetUserByIDWithAdminDetails(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	user, _, _ := service.Register("Test User", "test@example.com", "Password123")

	// Get with admin details
	retrieved, err := service.GetUserByIDWithAdminDetails(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user with admin details: %v", err)
	}

	if retrieved.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, retrieved.Email)
	}

	// Password hash should still be cleared
	if retrieved.PasswordHash != "" {
		t.Error("Password hash should be cleared")
	}

	// Non-existent user
	_, err = service.GetUserByIDWithAdminDetails(99999)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// Test DeleteUser
func TestDeleteUser(t *testing.T) {
	service := newTestUserService(true)

	// Register admin and regular user
	admin, _, _ := service.Register("Admin", "admin@example.com", "Password123")
	user, _, _ := service.Register("User", "user@example.com", "Password123")

	// Delete user
	err := service.DeleteUser(admin.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// User should no longer exist
	_, err = service.GetByID(user.ID)
	if err != ErrUserNotFound {
		t.Error("User should be deleted")
	}

	// Cannot delete self
	err = service.DeleteUser(admin.ID, admin.ID)
	if err == nil {
		t.Error("Should not be able to delete own account")
	}

	// Non-admin cannot delete
	regularUser, _, _ := service.Register("Regular", "regular@example.com", "Password123")
	anotherUser, _, _ := service.Register("Another", "another@example.com", "Password123")
	err = service.DeleteUser(regularUser.ID, anotherUser.ID)
	if err == nil {
		t.Error("Non-admin should not be able to delete users")
	}

	// Delete non-existent user
	err = service.DeleteUser(admin.ID, 99999)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// Test GetActiveSessions
func TestGetActiveSessions(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	user, _, _ := service.Register("Test User", "test@example.com", "Password123")

	// Create some sessions
	service.CreateRefreshToken(user.ID, "Device 1", false)
	service.CreateRefreshToken(user.ID, "Device 2", false)

	// Get active sessions
	sessions, err := service.GetActiveSessions(user.ID)
	if err != nil {
		t.Fatalf("Failed to get active sessions: %v", err)
	}

	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}

	// Non-existent user
	_, err = service.GetActiveSessions(99999)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// Test RevokeSession
func TestRevokeSession(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	user, _, _ := service.Register("Test User", "test@example.com", "Password123")

	// Create a session
	service.CreateRefreshToken(user.ID, "Device 1", false)

	// Get the session
	sessions, _ := service.GetActiveSessions(user.ID)
	if len(sessions) == 0 {
		t.Fatal("Expected at least one session")
	}

	// Revoke the session
	err := service.RevokeSession(user.ID, sessions[0].ID)
	if err != nil {
		t.Fatalf("Failed to revoke session: %v", err)
	}

	// Session should be revoked
	sessions, _ = service.GetActiveSessions(user.ID)
	if len(sessions) != 0 {
		t.Error("Session should be revoked")
	}

	// Non-existent user
	err = service.RevokeSession(99999, 1)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}

	// Non-existent session
	service.CreateRefreshToken(user.ID, "Device 1", false)
	err = service.RevokeSession(user.ID, 99999)
	if err == nil {
		t.Error("Should fail for non-existent session")
	}
}

// Test RevokeAllSessions
func TestRevokeAllSessions(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	user, _, _ := service.Register("Test User", "test@example.com", "Password123")

	// Create multiple sessions
	service.CreateRefreshToken(user.ID, "Device 1", false)
	service.CreateRefreshToken(user.ID, "Device 2", false)
	service.CreateRefreshToken(user.ID, "Device 3", false)

	// Get one session to keep
	sessions, _ := service.GetActiveSessions(user.ID)
	keepID := sessions[0].ID

	// Revoke all except one
	err := service.RevokeAllSessions(user.ID, &keepID)
	if err != nil {
		t.Fatalf("Failed to revoke all sessions: %v", err)
	}

	// Only one session should remain
	sessions, _ = service.GetActiveSessions(user.ID)
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessions))
	}

	// Create more sessions
	service.CreateRefreshToken(user.ID, "Device 4", false)
	service.CreateRefreshToken(user.ID, "Device 5", false)

	// Revoke all (no exception)
	err = service.RevokeAllSessions(user.ID, nil)
	if err != nil {
		t.Fatalf("Failed to revoke all sessions: %v", err)
	}

	sessions, _ = service.GetActiveSessions(user.ID)
	if len(sessions) != 0 {
		t.Error("All sessions should be revoked")
	}

	// Non-existent user
	err = service.RevokeAllSessions(99999, nil)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// Test Login with locked account
func TestLoginAccountLocked(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	_, _, _ = service.Register("Test User", "test@example.com", "Password123")

	// Lock the account
	user, _ := service.userRepo.GetByEmail("test@example.com")
	service.userRepo.LockAccount(user.ID, 15*time.Minute)

	// Try to login
	_, _, err := service.Login("test@example.com", "Password123")
	if err != ErrAccountLocked {
		t.Errorf("Expected ErrAccountLocked, got: %v", err)
	}
}

// Test password reset for non-existent email
func TestPasswordResetNonExistentEmail(t *testing.T) {
	service := newTestUserService(true)

	// Should silently succeed (security best practice)
	err := service.RequestPasswordReset("nonexistent@example.com")
	if err != nil {
		t.Errorf("Should silently succeed for non-existent email, got: %v", err)
	}
}

// ==================== Edge Case Tests ====================

// Test Register - GetByEmail returns database error
func TestRegisterGetByEmailError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	repo.getByEmailError = errors.New("database connection lost")
	service := newTestUserServiceWithRepo(repo, true)

	_, _, err := service.Register("Test User", "test@example.com", "Password123")
	if err == nil {
		t.Fatal("Expected error when GetByEmail fails")
	}
	if !strings.Contains(err.Error(), "failed to check existing user") {
		t.Errorf("Expected 'failed to check existing user' error, got: %v", err)
	}
}

// Test Register - Count returns database error
func TestRegisterCountError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	repo.countError = errors.New("count query failed")
	service := newTestUserServiceWithRepo(repo, true)

	_, _, err := service.Register("Test User", "test@example.com", "Password123")
	if err == nil {
		t.Fatal("Expected error when Count fails")
	}
	if !strings.Contains(err.Error(), "failed to count users") {
		t.Errorf("Expected 'failed to count users' error, got: %v", err)
	}
}

// Test Register - Create returns database error
func TestRegisterCreateError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	repo.createError = errors.New("insert failed")
	service := newTestUserServiceWithRepo(repo, true)

	_, _, err := service.Register("Test User", "test@example.com", "Password123")
	if err == nil {
		t.Fatal("Expected error when Create fails")
	}
	if !strings.Contains(err.Error(), "failed to create user") {
		t.Errorf("Expected 'failed to create user' error, got: %v", err)
	}
}

// Test Register - Update returns error during auto-verification
func TestRegisterAutoVerifyUpdateError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	// First register succeeds (to set up the first user)
	_, _, err := service.Register("Admin", "admin@example.com", "Password123")
	if err != nil {
		t.Fatalf("First register should succeed: %v", err)
	}

	// Now set update error - the second register will fail during auto-verify update
	repo.updateError = errors.New("update failed")
	_, _, err = service.Register("User", "user@example.com", "Password123")
	if err == nil {
		t.Fatal("Expected error when Update fails during auto-verification")
	}
	if !strings.Contains(err.Error(), "failed to update user") {
		t.Errorf("Expected 'failed to update user' error, got: %v", err)
	}
}

// Test Login - GetByEmail returns database error
func TestLoginGetByEmailError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	// Register a user first
	_, _, err := service.Register("Test User", "test@example.com", "Password123")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Now set the error for subsequent GetByEmail calls
	repo.getByEmailError = errors.New("database timeout")
	_, _, err = service.Login("test@example.com", "Password123")
	if err == nil {
		t.Fatal("Expected error when GetByEmail fails")
	}
	if !strings.Contains(err.Error(), "failed to get user") {
		t.Errorf("Expected 'failed to get user' error, got: %v", err)
	}
}

// Test Login - IsAccountLocked returns database error
func TestLoginIsLockedError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	// Register a user
	_, _, _ = service.Register("Test User", "test@example.com", "Password123")

	// Set IsAccountLocked to return error
	repo.isLockedError = errors.New("lock check failed")
	_, _, err := service.Login("test@example.com", "Password123")
	if err == nil {
		t.Fatal("Expected error when IsAccountLocked fails")
	}
	if !strings.Contains(err.Error(), "failed to check lock status") {
		t.Errorf("Expected 'failed to check lock status' error, got: %v", err)
	}
}

// Test Login - account lockout after max failed attempts
func TestLoginAccountLockoutAfterMaxAttempts(t *testing.T) {
	service := newTestUserService(true)

	// Register a user
	_, _, _ = service.Register("Test User", "test@example.com", "Password123")

	// Attempt wrong password 5 times (maxLoginAttempts = 5)
	for i := 0; i < 5; i++ {
		_, _, err := service.Login("test@example.com", "WrongPassword")
		if err != ErrInvalidCredentials {
			t.Errorf("Attempt %d: Expected ErrInvalidCredentials, got: %v", i+1, err)
		}
	}

	// Next attempt should show account locked
	_, _, err := service.Login("test@example.com", "Password123")
	if err != ErrAccountLocked {
		t.Errorf("Expected ErrAccountLocked after max attempts, got: %v", err)
	}
}

// Test UpdateProfile - GetByID returns error
func TestUpdateProfileGetByIDError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	repo.getByIDError = errors.New("database error")
	service := newTestUserServiceWithRepo(repo, true)

	_, err := service.UpdateProfile(99999, "New Name", "", nil)
	if err == nil {
		t.Fatal("Expected error when GetByID fails")
	}
	if !strings.Contains(err.Error(), "failed to get user") {
		t.Errorf("Expected 'failed to get user' error, got: %v", err)
	}
}

// Test UpdateProfile - GetByEmail returns error when checking new email
func TestUpdateProfileEmailCheckError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	// Register a user
	user, _, _ := service.Register("Test User", "test@example.com", "Password123")

	// Now set GetByEmail to return error
	repo.getByEmailError = errors.New("email check failed")

	_, err := service.UpdateProfile(user.ID, "", "newemail@example.com", nil)
	if err == nil {
		t.Fatal("Expected error when email check fails")
	}
	if !strings.Contains(err.Error(), "failed to check email") {
		t.Errorf("Expected 'failed to check email' error, got: %v", err)
	}
}

// Test UpdateProfile - Update returns error
func TestUpdateProfileUpdateError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	// Register a user
	user, _, _ := service.Register("Test User", "test@example.com", "Password123")

	// Now set update to fail
	repo.updateError = errors.New("update failed")
	_, err := service.UpdateProfile(user.ID, "New Name", "", nil)
	if err == nil {
		t.Fatal("Expected error when Update fails")
	}
	if !strings.Contains(err.Error(), "failed to update user") {
		t.Errorf("Expected 'failed to update user' error, got: %v", err)
	}
}

// Test ChangePassword - UpdatePassword returns error
func TestChangePasswordUpdatePasswordError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	// Register a user
	user, _, _ := service.Register("Test User", "test@example.com", "OldPassword123")

	// Set UpdatePassword to fail
	repo.updatePwdError = errors.New("password update failed")
	err := service.ChangePassword(user.ID, "OldPassword123", "NewPassword123")
	if err == nil {
		t.Fatal("Expected error when UpdatePassword fails")
	}
	if !strings.Contains(err.Error(), "password update failed") {
		t.Errorf("Expected 'password update failed' error, got: %v", err)
	}
}

// Test UnlockAccount - admin user not found (GetByID error)
func TestUnlockAccountAdminGetByIDError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	repo.getByIDError = errors.New("db connection error")
	service := newTestUserServiceWithRepo(repo, true)

	err := service.UnlockAccount(99999, 88888)
	if err == nil {
		t.Fatal("Expected error when admin GetByID fails")
	}
	if !strings.Contains(err.Error(), "failed to get admin user") {
		t.Errorf("Expected 'failed to get admin user' error, got: %v", err)
	}
}

// Test UnlockAccount - UnlockAccount repo method fails
func TestUnlockAccountRepoError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	// Register admin and user
	admin, _, _ := service.Register("Admin", "admin@example.com", "Password123")
	user, _, _ := service.Register("User", "user@example.com", "Password123")

	// Lock user
	service.userRepo.LockAccount(user.ID, 15*time.Minute)

	// Set unlock to fail
	repo.unlockError = errors.New("unlock query failed")
	err := service.UnlockAccount(admin.ID, user.ID)
	if err == nil {
		t.Fatal("Expected error when UnlockAccount repo fails")
	}
	if !strings.Contains(err.Error(), "failed to unlock account") {
		t.Errorf("Expected 'failed to unlock account' error, got: %v", err)
	}
}

// Test DisableAccount - target user GetByID error
func TestDisableAccountTargetGetByIDError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	// Register admin
	admin, _, _ := service.Register("Admin", "admin@example.com", "Password123")

	// Set getByIDError - will affect GetByID for non-existing target
	repo.getByIDError = errors.New("db error")
	err := service.DisableAccount(admin.ID, 99999, "test reason")
	if err == nil {
		t.Fatal("Expected error when target GetByID fails")
	}
	if !strings.Contains(err.Error(), "failed to get target user") {
		t.Errorf("Expected 'failed to get target user' error, got: %v", err)
	}
}

// Test DisableAccount - DisableAccount repo method fails
func TestDisableAccountRepoError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	admin, _, _ := service.Register("Admin", "admin@example.com", "Password123")
	user, _, _ := service.Register("User", "user@example.com", "Password123")

	repo.disableError = errors.New("disable query failed")
	err := service.DisableAccount(admin.ID, user.ID, "test reason")
	if err == nil {
		t.Fatal("Expected error when DisableAccount repo fails")
	}
	if !strings.Contains(err.Error(), "failed to disable account") {
		t.Errorf("Expected 'failed to disable account' error, got: %v", err)
	}
}

// Test EnableAccount - admin user not found
func TestEnableAccountAdminGetByIDError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	repo.getByIDError = errors.New("db error")
	service := newTestUserServiceWithRepo(repo, true)

	err := service.EnableAccount(99999, 88888)
	if err == nil {
		t.Fatal("Expected error when admin GetByID fails")
	}
	if !strings.Contains(err.Error(), "failed to get admin user") {
		t.Errorf("Expected 'failed to get admin user' error, got: %v", err)
	}
}

// Test EnableAccount - EnableAccount repo method fails
func TestEnableAccountRepoError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	admin, _, _ := service.Register("Admin", "admin@example.com", "Password123")
	user, _, _ := service.Register("User", "user@example.com", "Password123")

	repo.enableError = errors.New("enable query failed")
	err := service.EnableAccount(admin.ID, user.ID)
	if err == nil {
		t.Fatal("Expected error when EnableAccount repo fails")
	}
	if !strings.Contains(err.Error(), "failed to enable account") {
		t.Errorf("Expected 'failed to enable account' error, got: %v", err)
	}
}

// Test SetEmailVerification - admin not found
func TestSetEmailVerificationAdminNotFound(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	repo.getByIDError = errors.New("db error")
	service := newTestUserServiceWithRepo(repo, true)

	err := service.SetEmailVerification(99999, 88888, true)
	if err == nil {
		t.Fatal("Expected error when admin not found")
	}
	if !strings.Contains(err.Error(), "failed to get admin user") {
		t.Errorf("Expected 'failed to get admin user' error, got: %v", err)
	}
}

// Test SetEmailVerification - target not found
func TestSetEmailVerificationTargetNotFound(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	admin, _, _ := service.Register("Admin", "admin@example.com", "Password123")

	// Target doesn't exist - getByIDError not set so it returns nil, nil
	// But the service doesn't check for nil on these Get calls - it just continues
	// So let's set getByIDError after admin is created
	repo.getByIDError = errors.New("db error")
	err := service.SetEmailVerification(admin.ID, 99999, true)
	if err == nil {
		t.Fatal("Expected error when target not found")
	}
	if !strings.Contains(err.Error(), "failed to get target user") {
		t.Errorf("Expected 'failed to get target user' error, got: %v", err)
	}
}

// Test SetEmailVerification - Update fails
func TestSetEmailVerificationUpdateError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	admin, _, _ := service.Register("Admin", "admin@example.com", "Password123")
	user, _, _ := service.Register("User", "user@example.com", "Password123")

	repo.updateError = errors.New("update failed")
	err := service.SetEmailVerification(admin.ID, user.ID, true)
	if err == nil {
		t.Fatal("Expected error when Update fails")
	}
	if !strings.Contains(err.Error(), "failed to update email verification") {
		t.Errorf("Expected 'failed to update email verification' error, got: %v", err)
	}
}

// Test DeleteUser - admin GetByID returns error
func TestDeleteUserAdminGetByIDError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	repo.getByIDError = errors.New("db error")
	service := newTestUserServiceWithRepo(repo, true)

	err := service.DeleteUser(99999, 88888)
	if err == nil {
		t.Fatal("Expected error when admin GetByID fails")
	}
	if !strings.Contains(err.Error(), "failed to get admin user") {
		t.Errorf("Expected 'failed to get admin user' error, got: %v", err)
	}
}

// Test DeleteUser - repo Delete fails
func TestDeleteUserRepoDeleteError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	admin, _, _ := service.Register("Admin", "admin@example.com", "Password123")
	user, _, _ := service.Register("User", "user@example.com", "Password123")

	repo.deleteError = errors.New("delete constraint violation")
	err := service.DeleteUser(admin.ID, user.ID)
	if err == nil {
		t.Fatal("Expected error when Delete repo fails")
	}
	if !strings.Contains(err.Error(), "failed to delete user") {
		t.Errorf("Expected 'failed to delete user' error, got: %v", err)
	}
}

// Test ChangeUserRole - admin GetByID error
func TestChangeUserRoleAdminGetByIDError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	repo.getByIDError = errors.New("db error")
	service := newTestUserServiceWithRepo(repo, true)

	err := service.ChangeUserRole(99999, 88888, "admin")
	if err == nil {
		t.Fatal("Expected error when admin GetByID fails")
	}
	if !strings.Contains(err.Error(), "failed to get admin user") {
		t.Errorf("Expected 'failed to get admin user' error, got: %v", err)
	}
}

// Test ChangeUserRole - Update fails
func TestChangeUserRoleUpdateError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	admin, _, _ := service.Register("Admin", "admin@example.com", "Password123")
	user, _, _ := service.Register("User", "user@example.com", "Password123")

	repo.updateError = errors.New("update role failed")
	err := service.ChangeUserRole(admin.ID, user.ID, "admin")
	if err == nil {
		t.Fatal("Expected error when Update fails")
	}
	if !strings.Contains(err.Error(), "failed to update user role") {
		t.Errorf("Expected 'failed to update user role' error, got: %v", err)
	}
}

// Test GetActiveSessions - GetByID returns error
func TestGetActiveSessionsGetByIDError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	repo.getByIDError = errors.New("db error")
	service := newTestUserServiceWithRepo(repo, true)

	_, err := service.GetActiveSessions(99999)
	if err == nil {
		t.Fatal("Expected error when GetByID fails")
	}
	if !strings.Contains(err.Error(), "failed to get user") {
		t.Errorf("Expected 'failed to get user' error, got: %v", err)
	}
}

// Test RevokeSession - GetByID returns error
func TestRevokeSessionGetByIDError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	repo.getByIDError = errors.New("db error")
	service := newTestUserServiceWithRepo(repo, true)

	err := service.RevokeSession(99999, 1)
	if err == nil {
		t.Fatal("Expected error when GetByID fails")
	}
	if !strings.Contains(err.Error(), "failed to get user") {
		t.Errorf("Expected 'failed to get user' error, got: %v", err)
	}
}

// Test RevokeAllSessions - GetByID returns error
func TestRevokeAllSessionsGetByIDError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	repo.getByIDError = errors.New("db error")
	service := newTestUserServiceWithRepo(repo, true)

	err := service.RevokeAllSessions(99999, nil)
	if err == nil {
		t.Fatal("Expected error when GetByID fails")
	}
	if !strings.Contains(err.Error(), "failed to get user") {
		t.Errorf("Expected 'failed to get user' error, got: %v", err)
	}
}

// Test GetUserByIDWithAdminDetails - GetByID returns error
func TestGetUserByIDWithAdminDetailsGetByIDError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	repo.getByIDError = errors.New("db error")
	service := newTestUserServiceWithRepo(repo, true)

	_, err := service.GetUserByIDWithAdminDetails(99999)
	if err == nil {
		t.Fatal("Expected error when GetByID fails")
	}
	if !strings.Contains(err.Error(), "failed to get user") {
		t.Errorf("Expected 'failed to get user' error, got: %v", err)
	}
}

// Test ListUsers - List returns error
func TestListUsersCountError(t *testing.T) {
	repo := &mockUserRepo{users: make(map[int64]*domain.User), nextID: 0}
	service := newTestUserServiceWithRepo(repo, true)

	// Register a user to populate the repo
	service.Register("Test", "test@example.com", "Password123")

	// Set count error
	repo.countError = errors.New("count query failed")
	_, _, err := service.ListUsers(10, 0)
	if err == nil {
		t.Fatal("Expected error when Count fails")
	}
	if !strings.Contains(err.Error(), "failed to count users") {
		t.Errorf("Expected 'failed to count users' error, got: %v", err)
	}
}
