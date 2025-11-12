package repository

import (
	"database/sql"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// SQLiteUserSettingsRepository implements UserSettingsRepository for SQLite
type SQLiteUserSettingsRepository struct {
	db *sql.DB
}

// NewSQLiteUserSettingsRepository creates a new user settings repository
func NewSQLiteUserSettingsRepository(db *sql.DB) domain.UserSettingsRepository {
	return &SQLiteUserSettingsRepository{db: db}
}

// GetByUserID retrieves settings for a specific user
func (r *SQLiteUserSettingsRepository) GetByUserID(userID int64) (*domain.UserSettings, error) {
	query := `
		SELECT id, user_id, notification_preferences, data_export_format, theme,
		       weight_unit, distance_unit, created_at, updated_at
		FROM user_settings
		WHERE user_id = ?
	`

	settings := &domain.UserSettings{}
	err := r.db.QueryRow(query, userID).Scan(
		&settings.ID,
		&settings.UserID,
		&settings.NotificationPreferences,
		&settings.DataExportFormat,
		&settings.Theme,
		&settings.WeightUnit,
		&settings.DistanceUnit,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return settings, nil
}

// Create creates new settings for a user
func (r *SQLiteUserSettingsRepository) Create(settings *domain.UserSettings) error {
	query := `
		INSERT INTO user_settings (
			user_id, notification_preferences, data_export_format, theme,
			weight_unit, distance_unit, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	settings.CreatedAt = now
	settings.UpdatedAt = now

	result, err := r.db.Exec(
		query,
		settings.UserID,
		settings.NotificationPreferences,
		settings.DataExportFormat,
		settings.Theme,
		settings.WeightUnit,
		settings.DistanceUnit,
		settings.CreatedAt,
		settings.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	settings.ID = id
	return nil
}

// Update updates existing user settings
func (r *SQLiteUserSettingsRepository) Update(settings *domain.UserSettings) error {
	query := `
		UPDATE user_settings
		SET notification_preferences = ?, data_export_format = ?, theme = ?,
		    weight_unit = ?, distance_unit = ?, updated_at = ?
		WHERE user_id = ?
	`

	settings.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		settings.NotificationPreferences,
		settings.DataExportFormat,
		settings.Theme,
		settings.WeightUnit,
		settings.DistanceUnit,
		settings.UpdatedAt,
		settings.UserID,
	)

	return err
}

// Delete removes user settings
func (r *SQLiteUserSettingsRepository) Delete(userID int64) error {
	query := `DELETE FROM user_settings WHERE user_id = ?`
	_, err := r.db.Exec(query, userID)
	return err
}
