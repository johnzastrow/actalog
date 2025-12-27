package auth

import (
	"testing"
	"time"
)

// =============================================================================
// Password Tests
// =============================================================================

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "securePassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt allows empty passwords
		},
		{
			name:     "very long password",
			password: "a" + string(make([]byte, 71)), // 72 bytes is bcrypt limit
			wantErr:  false,
		},
		{
			name:     "password with special characters",
			password: "P@$$w0rd!#%^&*()",
			wantErr:  false,
		},
		{
			name:     "unicode password",
			password: "密码テスト🔐",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && hash == "" {
				t.Error("HashPassword() returned empty hash")
			}
			if !tt.wantErr && hash == tt.password {
				t.Error("HashPassword() returned unhashed password")
			}
		})
	}
}

func TestHashPassword_UniqueHashes(t *testing.T) {
	password := "testPassword123"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash1 == hash2 {
		t.Error("HashPassword() should generate unique hashes for the same password (salt)")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testPassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		wantErr        bool
	}{
		{
			name:           "correct password",
			hashedPassword: hash,
			password:       password,
			wantErr:        false,
		},
		{
			name:           "incorrect password",
			hashedPassword: hash,
			password:       "wrongPassword",
			wantErr:        true,
		},
		{
			name:           "empty password",
			hashedPassword: hash,
			password:       "",
			wantErr:        true,
		},
		{
			name:           "case sensitive",
			hashedPassword: hash,
			password:       "TESTPASSWORD123",
			wantErr:        true,
		},
		{
			name:           "invalid hash format",
			hashedPassword: "notavalidhash",
			password:       password,
			wantErr:        true,
		},
		{
			name:           "empty hash",
			hashedPassword: "",
			password:       password,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPassword(tt.hashedPassword, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckPassword_SpecialCharacters(t *testing.T) {
	specialPasswords := []string{
		"P@$$w0rd!#%^&*()",
		"pass word with spaces",
		"tab\tand\nnewline",
		"密码テスト🔐",
		"<script>alert('xss')</script>",
	}

	for _, password := range specialPasswords {
		t.Run(password, func(t *testing.T) {
			hash, err := HashPassword(password)
			if err != nil {
				t.Fatalf("HashPassword() error = %v", err)
			}

			err = CheckPassword(hash, password)
			if err != nil {
				t.Errorf("CheckPassword() failed for special password: %v", err)
			}
		})
	}
}

// =============================================================================
// JWT Tests
// =============================================================================

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		email      string
		role       string
		secret     string
		expiration time.Duration
		wantErr    bool
	}{
		{
			name:       "valid token",
			userID:     1,
			email:      "test@example.com",
			role:       "user",
			secret:     "testsecret123",
			expiration: time.Hour,
			wantErr:    false,
		},
		{
			name:       "admin role",
			userID:     2,
			email:      "admin@example.com",
			role:       "admin",
			secret:     "testsecret123",
			expiration: time.Hour * 24,
			wantErr:    false,
		},
		{
			name:       "short expiration",
			userID:     3,
			email:      "user@example.com",
			role:       "user",
			secret:     "testsecret123",
			expiration: time.Second,
			wantErr:    false,
		},
		{
			name:       "empty secret",
			userID:     4,
			email:      "user@example.com",
			role:       "user",
			secret:     "",
			expiration: time.Hour,
			wantErr:    false, // JWT library allows empty secrets
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.email, tt.role, tt.secret, tt.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	secret := "testsecret123"
	userID := int64(1)
	email := "test@example.com"
	role := "user"

	validToken, err := GenerateToken(userID, email, role, secret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	tests := []struct {
		name    string
		token   string
		secret  string
		wantErr bool
		errType error
	}{
		{
			name:    "valid token",
			token:   validToken,
			secret:  secret,
			wantErr: false,
		},
		{
			name:    "wrong secret",
			token:   validToken,
			secret:  "wrongsecret",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			secret:  secret,
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "not.a.valid.token",
			secret:  secret,
			wantErr: true,
		},
		{
			name:    "token with invalid signature",
			token:   validToken + "tampered",
			secret:  secret,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if claims.UserID != userID {
					t.Errorf("ValidateToken() userID = %v, want %v", claims.UserID, userID)
				}
				if claims.Email != email {
					t.Errorf("ValidateToken() email = %v, want %v", claims.Email, email)
				}
				if claims.Role != role {
					t.Errorf("ValidateToken() role = %v, want %v", claims.Role, role)
				}
			}
		})
	}
}

func TestValidateToken_Expired(t *testing.T) {
	secret := "testsecret123"

	// Generate a token with very short expiration
	token, err := GenerateToken(1, "test@example.com", "user", secret, time.Millisecond)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	_, err = ValidateToken(token, secret)
	if err == nil {
		t.Error("ValidateToken() should return error for expired token")
	}
}

func TestGenerateAndValidate_Roundtrip(t *testing.T) {
	testCases := []struct {
		userID int64
		email  string
		role   string
	}{
		{1, "user@example.com", "user"},
		{2, "admin@example.com", "admin"},
		{999999, "bigid@example.com", "user"},
		{1, "special+email@sub.domain.com", "user"},
	}

	secret := "testsecret123"
	expiration := time.Hour

	for _, tc := range testCases {
		t.Run(tc.email, func(t *testing.T) {
			token, err := GenerateToken(tc.userID, tc.email, tc.role, secret, expiration)
			if err != nil {
				t.Fatalf("GenerateToken() error = %v", err)
			}

			claims, err := ValidateToken(token, secret)
			if err != nil {
				t.Fatalf("ValidateToken() error = %v", err)
			}

			if claims.UserID != tc.userID {
				t.Errorf("UserID = %v, want %v", claims.UserID, tc.userID)
			}
			if claims.Email != tc.email {
				t.Errorf("Email = %v, want %v", claims.Email, tc.email)
			}
			if claims.Role != tc.role {
				t.Errorf("Role = %v, want %v", claims.Role, tc.role)
			}
		})
	}
}

func TestValidateToken_DifferentSecrets(t *testing.T) {
	token, err := GenerateToken(1, "test@example.com", "user", "secret1", time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Should fail with different secret
	_, err = ValidateToken(token, "secret2")
	if err == nil {
		t.Error("ValidateToken() should fail with different secret")
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkPassword123"
	for i := 0; i < b.N; i++ {
		HashPassword(password)
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	password := "benchmarkPassword123"
	hash, _ := HashPassword(password)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckPassword(hash, password)
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateToken(1, "test@example.com", "user", "secret123", time.Hour)
	}
}

func BenchmarkValidateToken(b *testing.B) {
	token, _ := GenerateToken(1, "test@example.com", "user", "secret123", time.Hour)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateToken(token, "secret123")
	}
}
