package repository

import (
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestUserSettingsRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	settingsRepo := NewSQLiteUserSettingsRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
		Role:         "athlete",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	settings := &domain.UserSettings{
		UserID:                  user.ID,
		NotificationPreferences: `{"email": true, "push": false}`,
		DataExportFormat:        "JSON",
		Theme:                   "dark",
		WeightUnit:              "lbs",
		DistanceUnit:            "miles",
	}

	err = settingsRepo.Create(settings)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if settings.ID == 0 {
		t.Error("Create() should set ID")
	}
	if settings.CreatedAt.IsZero() {
		t.Error("Create() should set CreatedAt")
	}
	if settings.UpdatedAt.IsZero() {
		t.Error("Create() should set UpdatedAt")
	}
}

func TestUserSettingsRepository_GetByUserID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	settingsRepo := NewSQLiteUserSettingsRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
		Role:         "athlete",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test getting non-existent settings
	got, err := settingsRepo.GetByUserID(user.ID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if got != nil {
		t.Error("GetByUserID() should return nil for user without settings")
	}

	// Create settings
	settings := &domain.UserSettings{
		UserID:                  user.ID,
		NotificationPreferences: `{"email": true}`,
		DataExportFormat:        "CSV",
		Theme:                   "light",
		WeightUnit:              "kg",
		DistanceUnit:            "km",
	}
	if err := settingsRepo.Create(settings); err != nil {
		t.Fatalf("Failed to create settings: %v", err)
	}

	// Get settings
	got, err = settingsRepo.GetByUserID(user.ID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByUserID() returned nil")
	}
	if got.UserID != user.ID {
		t.Errorf("UserID = %d, want %d", got.UserID, user.ID)
	}
	if got.Theme != "light" {
		t.Errorf("Theme = %v, want light", got.Theme)
	}
	if got.WeightUnit != "kg" {
		t.Errorf("WeightUnit = %v, want kg", got.WeightUnit)
	}
	if got.DistanceUnit != "km" {
		t.Errorf("DistanceUnit = %v, want km", got.DistanceUnit)
	}
}

func TestUserSettingsRepository_Update(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	settingsRepo := NewSQLiteUserSettingsRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
		Role:         "athlete",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create settings
	settings := &domain.UserSettings{
		UserID:                  user.ID,
		NotificationPreferences: `{"email": true}`,
		DataExportFormat:        "JSON",
		Theme:                   "light",
		WeightUnit:              "lbs",
		DistanceUnit:            "miles",
	}
	if err := settingsRepo.Create(settings); err != nil {
		t.Fatalf("Failed to create settings: %v", err)
	}

	originalUpdatedAt := settings.UpdatedAt

	// Update settings
	settings.Theme = "dark"
	settings.WeightUnit = "kg"
	err = settingsRepo.Update(settings)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	got, err := settingsRepo.GetByUserID(user.ID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if got.Theme != "dark" {
		t.Errorf("Theme = %v, want dark", got.Theme)
	}
	if got.WeightUnit != "kg" {
		t.Errorf("WeightUnit = %v, want kg", got.WeightUnit)
	}
	if !got.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestUserSettingsRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	settingsRepo := NewSQLiteUserSettingsRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
		Role:         "athlete",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create settings
	settings := &domain.UserSettings{
		UserID:                  user.ID,
		NotificationPreferences: `{}`,
		DataExportFormat:        "JSON",
		Theme:                   "light",
		WeightUnit:              "lbs",
		DistanceUnit:            "miles",
	}
	if err := settingsRepo.Create(settings); err != nil {
		t.Fatalf("Failed to create settings: %v", err)
	}

	// Delete settings
	err = settingsRepo.Delete(user.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion
	got, err := settingsRepo.GetByUserID(user.ID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if got != nil {
		t.Error("Settings should be deleted")
	}
}

func TestUserSettingsRepository_MultipleUsers(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	settingsRepo := NewSQLiteUserSettingsRepository(db)

	// Create multiple users with settings
	users := []*domain.User{
		{Email: "user1@example.com", PasswordHash: "hash", Name: "User 1", Role: "athlete"},
		{Email: "user2@example.com", PasswordHash: "hash", Name: "User 2", Role: "athlete"},
	}

	for _, u := range users {
		if err := userRepo.Create(u); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// Create different settings for each user
	settings1 := &domain.UserSettings{
		UserID:                  users[0].ID,
		NotificationPreferences: `{}`,
		DataExportFormat:        "JSON",
		Theme:                   "dark",
		WeightUnit:              "lbs",
		DistanceUnit:            "miles",
	}
	settings2 := &domain.UserSettings{
		UserID:                  users[1].ID,
		NotificationPreferences: `{}`,
		DataExportFormat:        "CSV",
		Theme:                   "light",
		WeightUnit:              "kg",
		DistanceUnit:            "km",
	}

	if err := settingsRepo.Create(settings1); err != nil {
		t.Fatalf("Failed to create settings1: %v", err)
	}
	if err := settingsRepo.Create(settings2); err != nil {
		t.Fatalf("Failed to create settings2: %v", err)
	}

	// Verify each user gets their own settings
	got1, err := settingsRepo.GetByUserID(users[0].ID)
	if err != nil {
		t.Fatalf("GetByUserID(user1) error = %v", err)
	}
	if got1.Theme != "dark" {
		t.Errorf("User 1 Theme = %v, want dark", got1.Theme)
	}

	got2, err := settingsRepo.GetByUserID(users[1].ID)
	if err != nil {
		t.Fatalf("GetByUserID(user2) error = %v", err)
	}
	if got2.Theme != "light" {
		t.Errorf("User 2 Theme = %v, want light", got2.Theme)
	}
}
