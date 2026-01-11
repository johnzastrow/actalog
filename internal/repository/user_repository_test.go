package repository

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestUserRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewSQLiteUserRepository(db)

	tests := []struct {
		name    string
		user    *domain.User
		wantErr bool
	}{
		{
			name: "create valid user",
			user: &domain.User{
				Email:        "test@example.com",
				PasswordHash: "hashed_password_123",
				Name:         "Test User",
				Role:         "user",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			wantErr: false,
		},
		{
			name: "create admin user",
			user: &domain.User{
				Email:        "admin@example.com",
				PasswordHash: "hashed_password_456",
				Name:         "Admin User",
				Role:         "admin",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			wantErr: false,
		},
		{
			name: "duplicate email should fail",
			user: &domain.User{
				Email:        "test@example.com", // Already exists
				PasswordHash: "another_hash",
				Name:         "Another User",
				Role:         "user",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.user.ID == 0 {
				t.Error("Create() should set user ID")
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewSQLiteUserRepository(db)

	// Create a test user
	user := &domain.User{
		Email:        "getbyid@example.com",
		PasswordHash: "hash123",
		Name:         "Get By ID User",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name      string
		id        int64
		wantUser  bool
		wantEmail string
	}{
		{
			name:      "existing user",
			id:        user.ID,
			wantUser:  true,
			wantEmail: "getbyid@example.com",
		},
		{
			name:     "non-existent user",
			id:       99999,
			wantUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(tt.id)
			if err != nil {
				t.Errorf("GetByID() unexpected error: %v", err)
				return
			}
			if tt.wantUser && got == nil {
				t.Error("GetByID() expected user but got nil")
				return
			}
			if !tt.wantUser && got != nil {
				t.Errorf("GetByID() expected nil but got user: %+v", got)
				return
			}
			if tt.wantUser && got.Email != tt.wantEmail {
				t.Errorf("GetByID() email = %v, want %v", got.Email, tt.wantEmail)
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewSQLiteUserRepository(db)

	// Create a test user
	user := &domain.User{
		Email:        "byemail@example.com",
		PasswordHash: "hash123",
		Name:         "Get By Email User",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name     string
		email    string
		wantUser bool
		wantName string
	}{
		{
			name:     "existing email",
			email:    "byemail@example.com",
			wantUser: true,
			wantName: "Get By Email User",
		},
		{
			name:     "non-existent email",
			email:    "notfound@example.com",
			wantUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByEmail(tt.email)
			if err != nil {
				t.Errorf("GetByEmail() unexpected error: %v", err)
				return
			}
			if tt.wantUser && got == nil {
				t.Error("GetByEmail() expected user but got nil")
				return
			}
			if !tt.wantUser && got != nil {
				t.Errorf("GetByEmail() expected nil but got user: %+v", got)
				return
			}
			if tt.wantUser && got.Name != tt.wantName {
				t.Errorf("GetByEmail() name = %v, want %v", got.Name, tt.wantName)
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewSQLiteUserRepository(db)

	// Create a test user
	user := &domain.User{
		Email:        "update@example.com",
		PasswordHash: "hash123",
		Name:         "Original Name",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Update the user
	user.Name = "Updated Name"
	user.Role = "admin"
	if err := repo.Update(user); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify the update
	updated, err := repo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("GetByID() after update error = %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("Update() name = %v, want Updated Name", updated.Name)
	}
	if updated.Role != "admin" {
		t.Errorf("Update() role = %v, want admin", updated.Role)
	}
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewSQLiteUserRepository(db)

	// Create a test user
	user := &domain.User{
		Email:        "password@example.com",
		PasswordHash: "old_hash",
		Name:         "Password User",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Update the password
	newHash := "new_secure_hash"
	if err := repo.UpdatePassword(user.ID, newHash); err != nil {
		t.Fatalf("UpdatePassword() error = %v", err)
	}

	// Verify the password update
	updated, err := repo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("GetByID() after password update error = %v", err)
	}
	if updated.PasswordHash != newHash {
		t.Errorf("UpdatePassword() hash = %v, want %v", updated.PasswordHash, newHash)
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewSQLiteUserRepository(db)

	// Create a test user
	user := &domain.User{
		Email:        "delete@example.com",
		PasswordHash: "hash123",
		Name:         "Delete User",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Verify user exists
	existing, err := repo.GetByID(user.ID)
	if err != nil || existing == nil {
		t.Fatal("user should exist before delete")
	}

	// Delete the user
	if err := repo.Delete(user.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify user is deleted
	deleted, err := repo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("GetByID() after delete error = %v", err)
	}
	if deleted != nil {
		t.Error("Delete() user still exists after deletion")
	}
}

func TestUserRepository_List(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewSQLiteUserRepository(db)

	// Create multiple test users
	for i := 0; i < 5; i++ {
		user := &domain.User{
			Email:        "list" + string(rune('0'+i)) + "@example.com",
			PasswordHash: "hash123",
			Name:         "List User " + string(rune('0'+i)),
			Role:         "user",
			CreatedAt:    time.Now().Add(time.Duration(i) * time.Second),
			UpdatedAt:    time.Now(),
		}
		if err := repo.Create(user); err != nil {
			t.Fatalf("failed to create test user %d: %v", i, err)
		}
	}

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "list all users",
			limit:     10,
			offset:    0,
			wantCount: 5,
		},
		{
			name:      "list with limit",
			limit:     3,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "list with offset",
			limit:     10,
			offset:    2,
			wantCount: 3,
		},
		{
			name:      "offset beyond count",
			limit:     10,
			offset:    10,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := repo.List(tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}
			if len(users) != tt.wantCount {
				t.Errorf("List() returned %d users, want %d", len(users), tt.wantCount)
			}
		})
	}
}

func TestUserRepository_Count(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewSQLiteUserRepository(db)

	// Initial count should be 0
	count, err := repo.Count()
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Count() initial = %d, want 0", count)
	}

	// Create users
	for i := 0; i < 3; i++ {
		user := &domain.User{
			Email:        "count" + string(rune('0'+i)) + "@example.com",
			PasswordHash: "hash123",
			Name:         "Count User " + string(rune('0'+i)),
			Role:         "user",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := repo.Create(user); err != nil {
			t.Fatalf("failed to create test user %d: %v", i, err)
		}
	}

	// Count after creation
	count, err = repo.Count()
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 3 {
		t.Errorf("Count() after creation = %d, want 3", count)
	}
}

func TestUserRepository_AccountLocking(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewSQLiteUserRepository(db)

	// Create a test user
	user := &domain.User{
		Email:        "locktest@example.com",
		PasswordHash: "hash123",
		Name:         "Lock Test User",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Test IncrementFailedAttempts
	t.Run("IncrementFailedAttempts", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			if err := repo.IncrementFailedAttempts(user.ID); err != nil {
				t.Fatalf("IncrementFailedAttempts() error = %v", err)
			}
		}

		updated, _ := repo.GetByID(user.ID)
		if updated.FailedLoginAttempts != 3 {
			t.Errorf("FailedLoginAttempts = %d, want 3", updated.FailedLoginAttempts)
		}
	})

	// Test ResetFailedAttempts
	t.Run("ResetFailedAttempts", func(t *testing.T) {
		if err := repo.ResetFailedAttempts(user.ID); err != nil {
			t.Fatalf("ResetFailedAttempts() error = %v", err)
		}

		updated, _ := repo.GetByID(user.ID)
		if updated.FailedLoginAttempts != 0 {
			t.Errorf("FailedLoginAttempts after reset = %d, want 0", updated.FailedLoginAttempts)
		}
	})

	// Test LockAccount
	t.Run("LockAccount", func(t *testing.T) {
		lockDuration := 15 * time.Minute
		if err := repo.LockAccount(user.ID, lockDuration); err != nil {
			t.Fatalf("LockAccount() error = %v", err)
		}

		locked, unlockTime, err := repo.IsAccountLocked(user.ID)
		if err != nil {
			t.Fatalf("IsAccountLocked() error = %v", err)
		}
		if !locked {
			t.Error("IsAccountLocked() = false, want true")
		}
		if unlockTime == nil {
			t.Error("unlockTime should not be nil")
		}
	})

	// Test UnlockAccount
	t.Run("UnlockAccount", func(t *testing.T) {
		if err := repo.UnlockAccount(user.ID); err != nil {
			t.Fatalf("UnlockAccount() error = %v", err)
		}

		locked, _, err := repo.IsAccountLocked(user.ID)
		if err != nil {
			t.Fatalf("IsAccountLocked() error = %v", err)
		}
		if locked {
			t.Error("IsAccountLocked() = true after unlock, want false")
		}
	})
}

func TestUserRepository_AccountDisabling(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewSQLiteUserRepository(db)

	// Create admin and regular user
	admin := &domain.User{
		Email:        "admin@example.com",
		PasswordHash: "hash123",
		Name:         "Admin User",
		Role:         "admin",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := repo.Create(admin); err != nil {
		t.Fatalf("failed to create admin user: %v", err)
	}

	user := &domain.User{
		Email:        "disable@example.com",
		PasswordHash: "hash123",
		Name:         "Disable Test User",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Test DisableAccount
	t.Run("DisableAccount", func(t *testing.T) {
		reason := "Violated terms of service"
		if err := repo.DisableAccount(user.ID, admin.ID, reason); err != nil {
			t.Fatalf("DisableAccount() error = %v", err)
		}

		updated, _ := repo.GetByID(user.ID)
		if !updated.AccountDisabled {
			t.Error("AccountDisabled = false, want true")
		}
		if updated.DisabledByUserID == nil || *updated.DisabledByUserID != admin.ID {
			t.Error("DisabledByUserID should be admin.ID")
		}
		if updated.DisableReason == nil || *updated.DisableReason != reason {
			t.Errorf("DisableReason = %v, want %v", updated.DisableReason, reason)
		}
	})

	// Test EnableAccount
	t.Run("EnableAccount", func(t *testing.T) {
		if err := repo.EnableAccount(user.ID); err != nil {
			t.Fatalf("EnableAccount() error = %v", err)
		}

		updated, _ := repo.GetByID(user.ID)
		if updated.AccountDisabled {
			t.Error("AccountDisabled = true after enable, want false")
		}
		if updated.DisabledByUserID != nil {
			t.Error("DisabledByUserID should be nil after enable")
		}
		if updated.DisableReason != nil {
			t.Error("DisableReason should be nil after enable")
		}
	})
}

// NOTE: GetByResetToken and GetByVerificationToken are stub implementations
// that always return nil - they use separate token tables in the actual implementation.
// Tests for these would just verify they return nil, so we skip testing them.

func TestUserRepository_ListAdmins(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewSQLiteUserRepository(db)

	// Initially no admins
	admins, err := repo.ListAdmins()
	if err != nil {
		t.Fatalf("ListAdmins() error = %v", err)
	}
	if len(admins) != 0 {
		t.Errorf("ListAdmins() should return empty initially, got %d", len(admins))
	}

	// Create some users with different roles
	users := []struct {
		email string
		role  string
	}{
		{"admin1@example.com", "admin"},
		{"admin2@example.com", "admin"},
		{"user1@example.com", "user"},
		{"user2@example.com", "user"},
		{"admin3@example.com", "admin"},
	}

	for _, u := range users {
		user := &domain.User{
			Email:        u.email,
			PasswordHash: "hash123",
			Name:         u.email,
			Role:         u.role,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := repo.Create(user); err != nil {
			t.Fatalf("failed to create user %s: %v", u.email, err)
		}
	}

	// List admins - should only get admins
	admins, err = repo.ListAdmins()
	if err != nil {
		t.Fatalf("ListAdmins() error = %v", err)
	}
	if len(admins) != 3 {
		t.Errorf("ListAdmins() returned %d admins, want 3", len(admins))
	}

	// Verify all returned users are admins
	for _, admin := range admins {
		if admin.Role != "admin" {
			t.Errorf("ListAdmins() returned non-admin user: %s with role %s", admin.Email, admin.Role)
		}
	}
}
