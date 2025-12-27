package repository

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestRefreshTokenRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	tokenRepo := NewSQLiteRefreshTokenRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	token := &domain.RefreshToken{
		UserID:     user.ID,
		Token:      "test-refresh-token-abc123",
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:  time.Now(),
		DeviceInfo: "Chrome on Windows",
	}

	err = tokenRepo.Create(token)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if token.ID == 0 {
		t.Error("Create() should set ID")
	}
}

func TestRefreshTokenRepository_GetByToken(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	tokenRepo := NewSQLiteRefreshTokenRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create valid token
	validToken := &domain.RefreshToken{
		UserID:     user.ID,
		Token:      "valid-token-123",
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:  time.Now(),
		DeviceInfo: "Chrome",
	}
	if err := tokenRepo.Create(validToken); err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Create expired token
	expiredToken := &domain.RefreshToken{
		UserID:     user.ID,
		Token:      "expired-token-456",
		ExpiresAt:  time.Now().Add(-1 * time.Hour), // Already expired
		CreatedAt:  time.Now().Add(-2 * time.Hour),
		DeviceInfo: "Firefox",
	}
	if err := tokenRepo.Create(expiredToken); err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	// Test getting valid token
	got, err := tokenRepo.GetByToken("valid-token-123")
	if err != nil {
		t.Fatalf("GetByToken() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByToken() returned nil for valid token")
	}
	if got.Token != "valid-token-123" {
		t.Errorf("Token = %v, want valid-token-123", got.Token)
	}
	if got.UserID != user.ID {
		t.Errorf("UserID = %v, want %v", got.UserID, user.ID)
	}

	// Test getting expired token (should return nil)
	got, err = tokenRepo.GetByToken("expired-token-456")
	if err != nil {
		t.Fatalf("GetByToken(expired) error = %v", err)
	}
	if got != nil {
		t.Error("GetByToken() should return nil for expired token")
	}

	// Test non-existent token
	got, err = tokenRepo.GetByToken("non-existent-token")
	if err != nil {
		t.Fatalf("GetByToken(non-existent) error = %v", err)
	}
	if got != nil {
		t.Error("GetByToken() should return nil for non-existent token")
	}
}

func TestRefreshTokenRepository_GetByUserID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	tokenRepo := NewSQLiteRefreshTokenRepository(db)

	// Create test users
	user1 := &domain.User{Email: "user1@example.com", PasswordHash: "hash", Name: "User 1", Role: "user"}
	user2 := &domain.User{Email: "user2@example.com", PasswordHash: "hash", Name: "User 2", Role: "user"}
	if err := userRepo.Create(user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	if err := userRepo.Create(user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Create tokens for user1
	for i := 0; i < 3; i++ {
		token := &domain.RefreshToken{
			UserID:     user1.ID,
			Token:      "user1-token-" + string(rune('0'+i)),
			ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
			CreatedAt:  time.Now(),
			DeviceInfo: "Device " + string(rune('0'+i)),
		}
		if err := tokenRepo.Create(token); err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}
	}

	// Create token for user2
	token := &domain.RefreshToken{
		UserID:     user2.ID,
		Token:      "user2-token",
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:  time.Now(),
		DeviceInfo: "Chrome",
	}
	if err := tokenRepo.Create(token); err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Get tokens for user1
	tokens, err := tokenRepo.GetByUserID(user1.ID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if len(tokens) != 3 {
		t.Errorf("GetByUserID() returned %d tokens, want 3", len(tokens))
	}

	// Get tokens for user2
	tokens, err = tokenRepo.GetByUserID(user2.ID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if len(tokens) != 1 {
		t.Errorf("GetByUserID() returned %d tokens, want 1", len(tokens))
	}

	// Get tokens for non-existent user
	tokens, err = tokenRepo.GetByUserID(999)
	if err != nil {
		t.Fatalf("GetByUserID(999) error = %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("GetByUserID(999) returned %d tokens, want 0", len(tokens))
	}
}

func TestRefreshTokenRepository_Revoke(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	tokenRepo := NewSQLiteRefreshTokenRepository(db)

	// Create test user
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create token
	token := &domain.RefreshToken{
		UserID:     user.ID,
		Token:      "test-token",
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:  time.Now(),
		DeviceInfo: "Chrome",
	}
	if err := tokenRepo.Create(token); err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Verify token is valid
	got, err := tokenRepo.GetByToken("test-token")
	if err != nil {
		t.Fatalf("GetByToken() error = %v", err)
	}
	if got == nil {
		t.Fatal("Token should exist")
	}

	// Revoke token
	err = tokenRepo.Revoke(token.ID)
	if err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}

	// Token should no longer be valid (revoked)
	got, err = tokenRepo.GetByToken("test-token")
	if err != nil {
		t.Fatalf("GetByToken() after revoke error = %v", err)
	}
	if got != nil {
		t.Error("Revoked token should not be returned by GetByToken")
	}
}

func TestRefreshTokenRepository_RevokeAllForUser(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	tokenRepo := NewSQLiteRefreshTokenRepository(db)

	// Create test user
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create multiple tokens
	for i := 0; i < 3; i++ {
		token := &domain.RefreshToken{
			UserID:     user.ID,
			Token:      "token-" + string(rune('0'+i)),
			ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
			CreatedAt:  time.Now(),
			DeviceInfo: "Device",
		}
		if err := tokenRepo.Create(token); err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}
	}

	// Verify tokens exist
	tokens, err := tokenRepo.GetByUserID(user.ID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if len(tokens) != 3 {
		t.Fatalf("Expected 3 tokens, got %d", len(tokens))
	}

	// Revoke all tokens
	err = tokenRepo.RevokeAllForUser(user.ID)
	if err != nil {
		t.Fatalf("RevokeAllForUser() error = %v", err)
	}

	// Verify all tokens are revoked
	tokens, err = tokenRepo.GetByUserID(user.ID)
	if err != nil {
		t.Fatalf("GetByUserID() after revoke error = %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("After RevokeAllForUser, expected 0 tokens, got %d", len(tokens))
	}
}

func TestRefreshTokenRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	tokenRepo := NewSQLiteRefreshTokenRepository(db)

	// Create test user
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create token
	token := &domain.RefreshToken{
		UserID:     user.ID,
		Token:      "test-token",
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:  time.Now(),
		DeviceInfo: "Chrome",
	}
	if err := tokenRepo.Create(token); err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Delete token
	err = tokenRepo.Delete(token.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify token is deleted
	got, err := tokenRepo.GetByToken("test-token")
	if err != nil {
		t.Fatalf("GetByToken() error = %v", err)
	}
	if got != nil {
		t.Error("Deleted token should not be returned")
	}
}

func TestRefreshTokenRepository_DeleteExpired(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	tokenRepo := NewSQLiteRefreshTokenRepository(db)

	// Create test user
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test", Role: "user"}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create valid token
	validToken := &domain.RefreshToken{
		UserID:     user.ID,
		Token:      "valid-token",
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:  time.Now(),
		DeviceInfo: "Chrome",
	}
	if err := tokenRepo.Create(validToken); err != nil {
		t.Fatalf("Failed to create valid token: %v", err)
	}

	// Create expired tokens using direct SQL
	for i := 0; i < 3; i++ {
		_, err := db.Exec(`INSERT INTO refresh_tokens (user_id, token, expires_at, created_at, device_info) VALUES (?, ?, ?, ?, ?)`,
			user.ID, "expired-token-"+string(rune('0'+i)), time.Now().Add(-1*time.Hour), time.Now().Add(-2*time.Hour), "Device")
		if err != nil {
			t.Fatalf("Failed to create expired token: %v", err)
		}
	}

	// Count all tokens
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM refresh_tokens WHERE user_id = ?", user.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count tokens: %v", err)
	}
	if count != 4 {
		t.Fatalf("Expected 4 tokens, got %d", count)
	}

	// Delete expired tokens
	err = tokenRepo.DeleteExpired()
	if err != nil {
		t.Fatalf("DeleteExpired() error = %v", err)
	}

	// Verify only valid token remains
	err = db.QueryRow("SELECT COUNT(*) FROM refresh_tokens WHERE user_id = ?", user.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count tokens after delete: %v", err)
	}
	if count != 1 {
		t.Errorf("After DeleteExpired, expected 1 token, got %d", count)
	}
}
