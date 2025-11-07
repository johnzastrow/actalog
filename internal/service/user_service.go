package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/auth"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrRegistrationClosed  = errors.New("registration is closed")
)

// UserService handles user-related business logic
type UserService struct {
	userRepo         domain.UserRepository
	jwtSecret        string
	jwtExpiration    time.Duration
	allowRegistration bool
}

// NewUserService creates a new user service
func NewUserService(
	userRepo domain.UserRepository,
	jwtSecret string,
	jwtExpiration time.Duration,
	allowRegistration bool,
) *UserService {
	return &UserService{
		userRepo:         userRepo,
		jwtSecret:        jwtSecret,
		jwtExpiration:    jwtExpiration,
		allowRegistration: allowRegistration,
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
