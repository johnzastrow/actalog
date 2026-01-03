package service

import (
	"encoding/json"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// UserSettingsService handles business logic for user settings
type UserSettingsService struct {
	settingsRepo domain.UserSettingsRepository
	auditLogRepo domain.AuditLogRepository
}

// NewUserSettingsService creates a new user settings service
func NewUserSettingsService(settingsRepo domain.UserSettingsRepository, auditLogRepo domain.AuditLogRepository) *UserSettingsService {
	return &UserSettingsService{
		settingsRepo: settingsRepo,
		auditLogRepo: auditLogRepo,
	}
}

// GetSettings retrieves settings for a user, creating defaults if none exist
func (s *UserSettingsService) GetSettings(userID int64) (*domain.UserSettings, error) {
	settings, err := s.settingsRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	// If no settings exist, create defaults
	if settings == nil {
		settings = &domain.UserSettings{
			UserID:                  userID,
			NotificationPreferences: "{}",
			DataExportFormat:        "JSON",
			Theme:                   "light",
			WeightUnit:              "lbs",
			DistanceUnit:            "miles",
			Timezone:                "America/New_York",
		}

		if err := s.settingsRepo.Create(settings); err != nil {
			return nil, err
		}
	}

	return settings, nil
}

// UpdateSettings updates user settings
func (s *UserSettingsService) UpdateSettings(userID int64, userEmail string, updates *domain.UserSettings) (*domain.UserSettings, error) {
	// Get existing settings or create defaults
	existing, err := s.GetSettings(userID)
	if err != nil {
		return nil, err
	}

	// Store old values for audit logging
	oldNotificationPrefs := existing.NotificationPreferences
	oldDataExportFormat := existing.DataExportFormat
	oldTheme := existing.Theme
	oldWeightUnit := existing.WeightUnit
	oldDistanceUnit := existing.DistanceUnit
	oldTimezone := existing.Timezone

	// Update fields (preserve ID and UserID)
	existing.NotificationPreferences = updates.NotificationPreferences
	existing.DataExportFormat = updates.DataExportFormat
	existing.Theme = updates.Theme
	existing.WeightUnit = updates.WeightUnit
	existing.DistanceUnit = updates.DistanceUnit
	existing.Timezone = updates.Timezone

	if err := s.settingsRepo.Update(existing); err != nil {
		return nil, err
	}

	// Audit log the settings update
	if s.auditLogRepo != nil {
		details, _ := json.Marshal(map[string]interface{}{
			"user_id":    userID,
			"updated_by": userEmail,
			"changes": map[string]interface{}{
				"notification_preferences_old": oldNotificationPrefs,
				"notification_preferences_new": existing.NotificationPreferences,
				"data_export_format_old":       oldDataExportFormat,
				"data_export_format_new":       existing.DataExportFormat,
				"theme_old":                    oldTheme,
				"theme_new":                    existing.Theme,
				"weight_unit_old":              oldWeightUnit,
				"weight_unit_new":              existing.WeightUnit,
				"distance_unit_old":            oldDistanceUnit,
				"distance_unit_new":            existing.DistanceUnit,
				"timezone_old":                 oldTimezone,
				"timezone_new":                 existing.Timezone,
			},
		})
		detailsStr := string(details)
		targetUserID := userID
		_ = s.auditLogRepo.Create(&domain.AuditLog{
			UserID:       &targetUserID,
			TargetUserID: &targetUserID,
			EventType:    domain.EventUserSettingsUpdate,
			Details:      &detailsStr,
			CreatedAt:    time.Now(),
		})
	}

	return existing, nil
}
