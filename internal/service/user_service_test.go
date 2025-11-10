package service

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/auth"
)

// Mock user repository
type mockUserRepo struct {
	users map[int64]*domain.User
	nextID int64
}

func (m *mockUserRepo) Create(user *domain.User) error {
	m.nextID++
	user.ID = m.nextID
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(id int64) (*domain.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return user, nil
}

func (m *mockUserRepo) GetByEmail(email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockUserRepo) GetByResetToken(token string) (*domain.User, error) {
	for _, user := range m.users {
		if user.ResetToken != nil && *user.ResetToken == token {
			return user, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockUserRepo) Update(user *domain.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return sql.ErrNoRows
	}
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) Delete(id int64) error {
	if _, ok := m.users[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) List(limit, offset int) ([]*domain.User, error) {
	var users []*domain.User
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *mockUserRepo) Count() (int64, error) {
	return int64(len(m.users)), nil
}

// Mock email service
type mockEmailService struct {
	sentEmails []mockEmail
}

type mockEmail struct {
	to      string
	subject string
	body    string
}

func (m *mockEmailService) SendPasswordResetEmail(to, token, resetURL string) error {
	m.sentEmails = append(m.sentEmails, mockEmail{
		to:      to,
		subject: "Password Reset",
		body:    resetURL,
	})
	return nil
}

// Helper to create test user service
func newTestUserService(allowRegistration bool) *UserService {
	return NewUserService(
		&mockUserRepo{users: make(map[int64]*domain.User), nextID: 0},
		"test-secret-key",
		24*time.Hour,
		allowRegistration,
		&mockEmailService{},
		"http://localhost:3000",
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

	if user2.Role != "user" {
		t.Errorf("Second user should be regular user, got role: %s", user2.Role)
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
	err = service.GeneratePasswordResetToken(email)
	if err != nil {
		t.Fatalf("Failed to generate reset token: %v", err)
	}

	// Verify email was sent
	if len(emailService.sentEmails) != 1 {
		t.Errorf("Expected 1 email sent, got %d", len(emailService.sentEmails))
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
	err = service.GeneratePasswordResetToken(email)
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

	// Verify we can parse the token (basic check)
	claims, err := auth.ParseToken(token, service.jwtSecretKey)
	if err != nil {
		t.Fatalf("Failed to parse JWT token: %v", err)
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
