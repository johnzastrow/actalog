package domain

import "time"

// BackupMetadata represents metadata about a database backup
type BackupMetadata struct {
	Filename       string    `json:"filename"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedByEmail string    `json:"created_by_email"`
	Version        string    `json:"version"`
	DatabaseDriver string    `json:"database_driver"`
	DatabaseName   string    `json:"database_name"`
	TotalUsers     int       `json:"total_users"`
	TotalWorkouts  int       `json:"total_workouts"`
	TotalMovements int       `json:"total_movements"`
	TotalWODs      int       `json:"total_wods"`
	FileSize       int64     `json:"file_size"` // Size in bytes
	FilePath       string    `json:"-"`         // Not exported to JSON
}

// BackupData represents the complete backup data structure
type BackupData struct {
	Metadata                BackupMetadata               `json:"metadata"`
	Users                   []map[string]interface{}     `json:"users"`
	Movements               []map[string]interface{}     `json:"movements"`
	WODs                    []map[string]interface{}     `json:"wods"`
	Workouts                []map[string]interface{}     `json:"workouts"`
	UserWorkouts            []map[string]interface{}     `json:"user_workouts"`
	WorkoutMovements        []map[string]interface{}     `json:"workout_movements"`
	WorkoutWODs             []map[string]interface{}     `json:"workout_wods"`
	UserWorkoutMovements    []map[string]interface{}     `json:"user_workout_movements"`
	UserWorkoutWODs         []map[string]interface{}     `json:"user_workout_wods"`
	RefreshTokens           []map[string]interface{}     `json:"refresh_tokens"`
	PasswordResets          []map[string]interface{}     `json:"password_resets"`
	EmailVerificationTokens []map[string]interface{}     `json:"email_verification_tokens"`
	AuditLogs               []map[string]interface{}     `json:"audit_logs"`
	UserSettings            []map[string]interface{}     `json:"user_settings"`
}

// BackupService defines the interface for backup/restore operations
type BackupService interface {
	// CreateBackup creates a full database backup and returns the filename
	CreateBackup(createdByUserID int64) (string, error)

	// ListBackups returns metadata for all available backups
	ListBackups() ([]BackupMetadata, error)

	// GetBackupMetadata reads and returns metadata for a specific backup
	GetBackupMetadata(filename string) (*BackupMetadata, error)

	// DownloadBackup returns the file path for downloading a backup
	DownloadBackup(filename string) (string, error)

	// UploadBackup saves an uploaded backup file to the backups directory
	UploadBackup(file interface{}, filename string, uploadedByUserID int64) (string, error)

	// DeleteBackup removes a backup file and logs the deletion
	DeleteBackup(filename string, deletedByUserID int64) error

	// RestoreBackup restores database from a backup file
	RestoreBackup(filename string, restoredByUserID int64) error
}
