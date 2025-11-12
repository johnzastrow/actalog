package service

import (
	"github.com/johnzastrow/actalog/internal/domain"
)

// UserSettingsService handles business logic for user settings
type UserSettingsService struct {
	settingsRepo domain.UserSettingsRepository
}

// NewUserSettingsService creates a new user settings service
func NewUserSettingsService(settingsRepo domain.UserSettingsRepository) *UserSettingsService {
	return &UserSettingsService{
		settingsRepo: settingsRepo,
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
		}

		if err := s.settingsRepo.Create(settings); err != nil {
			return nil, err
		}
	}

	return settings, nil
}

// UpdateSettings updates user settings
func (s *UserSettingsService) UpdateSettings(userID int64, updates *domain.UserSettings) (*domain.UserSettings, error) {
	// Get existing settings or create defaults
	existing, err := s.GetSettings(userID)
	if err != nil {
		return nil, err
	}

	// Update fields (preserve ID and UserID)
	existing.NotificationPreferences = updates.NotificationPreferences
	existing.DataExportFormat = updates.DataExportFormat
	existing.Theme = updates.Theme
	existing.WeightUnit = updates.WeightUnit
	existing.DistanceUnit = updates.DistanceUnit

	if err := s.settingsRepo.Update(existing); err != nil {
		return nil, err
	}

	return existing, nil
}
