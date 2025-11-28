package domain

import "time"

// UserSettings represents user preferences and settings
type UserSettings struct {
	ID                      int64     `json:"id"`
	UserID                  int64     `json:"user_id"`
	NotificationPreferences string    `json:"notification_preferences"` // JSON format
	DataExportFormat        string    `json:"data_export_format"`       // JSON, CSV
	Theme                   string    `json:"theme"`                    // light, dark
	WeightUnit              string    `json:"weight_unit"`              // lbs, kg
	DistanceUnit            string    `json:"distance_unit"`            // miles, km
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// UserSettingsRepository defines the interface for user settings data access
type UserSettingsRepository interface {
	// GetByUserID retrieves settings for a specific user
	GetByUserID(userID int64) (*UserSettings, error)

	// Create creates new settings for a user
	Create(settings *UserSettings) error

	// Update updates existing user settings
	Update(settings *UserSettings) error

	// Delete removes user settings
	Delete(userID int64) error
}
