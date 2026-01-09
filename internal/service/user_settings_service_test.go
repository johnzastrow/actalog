package service

import (
	"errors"
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestNewUserSettingsService(t *testing.T) {
	settingsRepo := newMockUserSettingsRepo()
	auditLogRepo := newMockAuditLogRepo()
	svc := NewUserSettingsService(settingsRepo, auditLogRepo)

	if svc == nil {
		t.Fatal("NewUserSettingsService returned nil")
	}
}

func TestUserSettingsService_GetSettings_ExistingSettings(t *testing.T) {
	settingsRepo := newMockUserSettingsRepo()
	auditLogRepo := newMockAuditLogRepo()
	svc := NewUserSettingsService(settingsRepo, auditLogRepo)

	// Pre-create settings
	existingSettings := &domain.UserSettings{
		UserID:                  1,
		NotificationPreferences: `{"pr_achievements": {"enabled": true}}`,
		DataExportFormat:        "CSV",
		Theme:                   "dark",
		WeightUnit:              "kg",
		DistanceUnit:            "km",
	}
	_ = settingsRepo.Create(existingSettings)

	// Get settings
	settings, err := svc.GetSettings(1)
	if err != nil {
		t.Errorf("GetSettings() error = %v", err)
	}
	if settings == nil {
		t.Fatal("GetSettings() returned nil")
	}

	// Verify existing settings are returned
	if settings.Theme != "dark" {
		t.Errorf("Theme = %s, want dark", settings.Theme)
	}
	if settings.WeightUnit != "kg" {
		t.Errorf("WeightUnit = %s, want kg", settings.WeightUnit)
	}
	if settings.DataExportFormat != "CSV" {
		t.Errorf("DataExportFormat = %s, want CSV", settings.DataExportFormat)
	}
}

func TestUserSettingsService_GetSettings_CreatesDefaults(t *testing.T) {
	settingsRepo := newMockUserSettingsRepo()
	auditLogRepo := newMockAuditLogRepo()
	svc := NewUserSettingsService(settingsRepo, auditLogRepo)

	// Get settings for user with no existing settings
	settings, err := svc.GetSettings(42)
	if err != nil {
		t.Errorf("GetSettings() error = %v", err)
	}
	if settings == nil {
		t.Fatal("GetSettings() returned nil")
	}

	// Verify default values
	if settings.UserID != 42 {
		t.Errorf("UserID = %d, want 42", settings.UserID)
	}
	if settings.NotificationPreferences != "{}" {
		t.Errorf("NotificationPreferences = %s, want {}", settings.NotificationPreferences)
	}
	if settings.DataExportFormat != "JSON" {
		t.Errorf("DataExportFormat = %s, want JSON", settings.DataExportFormat)
	}
	if settings.Theme != "light" {
		t.Errorf("Theme = %s, want light", settings.Theme)
	}
	if settings.WeightUnit != "lbs" {
		t.Errorf("WeightUnit = %s, want lbs", settings.WeightUnit)
	}
	if settings.DistanceUnit != "miles" {
		t.Errorf("DistanceUnit = %s, want miles", settings.DistanceUnit)
	}

	// Verify settings were persisted
	storedSettings, _ := settingsRepo.GetByUserID(42)
	if storedSettings == nil {
		t.Error("Default settings were not persisted")
	}
}

func TestUserSettingsService_GetSettings_CreateError(t *testing.T) {
	settingsRepo := newMockUserSettingsRepo()
	settingsRepo.createError = errors.New("database error")
	auditLogRepo := newMockAuditLogRepo()
	svc := NewUserSettingsService(settingsRepo, auditLogRepo)

	// Get settings should fail when create fails
	_, err := svc.GetSettings(1)
	if err == nil {
		t.Error("GetSettings() should return error when create fails")
	}
}

func TestUserSettingsService_UpdateSettings(t *testing.T) {
	settingsRepo := newMockUserSettingsRepo()
	auditLogRepo := newMockAuditLogRepo()
	svc := NewUserSettingsService(settingsRepo, auditLogRepo)

	// Create initial settings
	_ = settingsRepo.Create(&domain.UserSettings{
		UserID:                  1,
		NotificationPreferences: "{}",
		DataExportFormat:        "JSON",
		Theme:                   "light",
		WeightUnit:              "lbs",
		DistanceUnit:            "miles",
	})

	// Update settings
	updates := &domain.UserSettings{
		NotificationPreferences: `{"pr_achievements": {"enabled": false}}`,
		DataExportFormat:        "CSV",
		Theme:                   "dark",
		WeightUnit:              "kg",
		DistanceUnit:            "km",
	}

	result, err := svc.UpdateSettings(1, "user@example.com", updates)
	if err != nil {
		t.Errorf("UpdateSettings() error = %v", err)
	}
	if result == nil {
		t.Fatal("UpdateSettings() returned nil")
	}

	// Verify updates were applied
	if result.Theme != "dark" {
		t.Errorf("Theme = %s, want dark", result.Theme)
	}
	if result.WeightUnit != "kg" {
		t.Errorf("WeightUnit = %s, want kg", result.WeightUnit)
	}
	if result.DistanceUnit != "km" {
		t.Errorf("DistanceUnit = %s, want km", result.DistanceUnit)
	}
	if result.DataExportFormat != "CSV" {
		t.Errorf("DataExportFormat = %s, want CSV", result.DataExportFormat)
	}

	// Verify audit log was created
	if len(auditLogRepo.logs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(auditLogRepo.logs))
	}
	if auditLogRepo.logs[0].EventType != domain.EventUserSettingsUpdate {
		t.Errorf("EventType = %s, want %s", auditLogRepo.logs[0].EventType, domain.EventUserSettingsUpdate)
	}
}

func TestUserSettingsService_UpdateSettings_CreatesDefaultsFirst(t *testing.T) {
	settingsRepo := newMockUserSettingsRepo()
	auditLogRepo := newMockAuditLogRepo()
	svc := NewUserSettingsService(settingsRepo, auditLogRepo)

	// Update settings for user with no existing settings
	updates := &domain.UserSettings{
		NotificationPreferences: `{"test": true}`,
		DataExportFormat:        "CSV",
		Theme:                   "dark",
		WeightUnit:              "kg",
		DistanceUnit:            "km",
	}

	result, err := svc.UpdateSettings(99, "newuser@example.com", updates)
	if err != nil {
		t.Errorf("UpdateSettings() error = %v", err)
	}
	if result == nil {
		t.Fatal("UpdateSettings() returned nil")
	}

	// Verify UserID is preserved
	if result.UserID != 99 {
		t.Errorf("UserID = %d, want 99", result.UserID)
	}

	// Verify updates were applied
	if result.Theme != "dark" {
		t.Errorf("Theme = %s, want dark", result.Theme)
	}
}

func TestUserSettingsService_UpdateSettings_UpdateError(t *testing.T) {
	settingsRepo := newMockUserSettingsRepo()
	auditLogRepo := newMockAuditLogRepo()
	svc := NewUserSettingsService(settingsRepo, auditLogRepo)

	// Create initial settings
	_ = settingsRepo.Create(&domain.UserSettings{
		UserID:                  1,
		NotificationPreferences: "{}",
		DataExportFormat:        "JSON",
		Theme:                   "light",
		WeightUnit:              "lbs",
		DistanceUnit:            "miles",
	})

	// Set update error
	settingsRepo.updateError = errors.New("update failed")

	updates := &domain.UserSettings{
		Theme: "dark",
	}

	_, err := svc.UpdateSettings(1, "user@example.com", updates)
	if err == nil {
		t.Error("UpdateSettings() should return error when update fails")
	}

	// Verify no audit log was created
	if len(auditLogRepo.logs) != 0 {
		t.Errorf("Expected 0 audit logs on failure, got %d", len(auditLogRepo.logs))
	}
}

func TestUserSettingsService_UpdateSettings_WithoutAuditLogRepo(t *testing.T) {
	settingsRepo := newMockUserSettingsRepo()
	// Pass nil for audit log repo
	svc := NewUserSettingsService(settingsRepo, nil)

	// Create initial settings
	_ = settingsRepo.Create(&domain.UserSettings{
		UserID:                  1,
		NotificationPreferences: "{}",
		DataExportFormat:        "JSON",
		Theme:                   "light",
		WeightUnit:              "lbs",
		DistanceUnit:            "miles",
	})

	updates := &domain.UserSettings{
		Theme: "dark",
	}

	// Should not panic when audit log repo is nil
	result, err := svc.UpdateSettings(1, "user@example.com", updates)
	if err != nil {
		t.Errorf("UpdateSettings() error = %v", err)
	}
	if result == nil {
		t.Fatal("UpdateSettings() returned nil")
	}
	if result.Theme != "dark" {
		t.Errorf("Theme = %s, want dark", result.Theme)
	}
}

func TestUserSettingsService_UpdateSettings_AllFields(t *testing.T) {
	settingsRepo := newMockUserSettingsRepo()
	auditLogRepo := newMockAuditLogRepo()
	svc := NewUserSettingsService(settingsRepo, auditLogRepo)

	// Create initial settings
	_ = settingsRepo.Create(&domain.UserSettings{
		UserID:                  1,
		NotificationPreferences: `{"old": true}`,
		DataExportFormat:        "JSON",
		Theme:                   "light",
		WeightUnit:              "lbs",
		DistanceUnit:            "miles",
	})

	// Update all fields
	updates := &domain.UserSettings{
		NotificationPreferences: `{"pr_achievements": {"enabled": true, "email": false}}`,
		DataExportFormat:        "CSV",
		Theme:                   "dark",
		WeightUnit:              "kg",
		DistanceUnit:            "km",
	}

	result, err := svc.UpdateSettings(1, "user@example.com", updates)
	if err != nil {
		t.Errorf("UpdateSettings() error = %v", err)
	}

	// Verify all fields were updated
	if result.NotificationPreferences != updates.NotificationPreferences {
		t.Errorf("NotificationPreferences = %s, want %s", result.NotificationPreferences, updates.NotificationPreferences)
	}
	if result.DataExportFormat != updates.DataExportFormat {
		t.Errorf("DataExportFormat = %s, want %s", result.DataExportFormat, updates.DataExportFormat)
	}
	if result.Theme != updates.Theme {
		t.Errorf("Theme = %s, want %s", result.Theme, updates.Theme)
	}
	if result.WeightUnit != updates.WeightUnit {
		t.Errorf("WeightUnit = %s, want %s", result.WeightUnit, updates.WeightUnit)
	}
	if result.DistanceUnit != updates.DistanceUnit {
		t.Errorf("DistanceUnit = %s, want %s", result.DistanceUnit, updates.DistanceUnit)
	}

	// Verify audit log contains details
	if auditLogRepo.logs[0].Details == nil {
		t.Error("Audit log should contain details")
	}
}
