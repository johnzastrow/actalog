package service

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/version"
	_ "github.com/mattn/go-sqlite3" // SQLite driver for database dumps
)

// BackupServiceImpl implements domain.BackupService
type BackupServiceImpl struct {
	db           *sql.DB
	dbDriver     string
	dbName       string
	backupDir    string
	uploadsDir   string
	userRepo     domain.UserRepository
	auditLogRepo domain.AuditLogRepository
}

// NewBackupService creates a new backup service
func NewBackupService(
	db *sql.DB,
	dbDriver string,
	dbName string,
	backupDir string,
	uploadsDir string,
	userRepo domain.UserRepository,
	auditLogRepo domain.AuditLogRepository,
) *BackupServiceImpl {
	return &BackupServiceImpl{
		db:           db,
		dbDriver:     dbDriver,
		dbName:       dbName,
		backupDir:    backupDir,
		uploadsDir:   uploadsDir,
		userRepo:     userRepo,
		auditLogRepo: auditLogRepo,
	}
}

// CreateBackup creates a full database backup and returns the filename
func (s *BackupServiceImpl) CreateBackup(createdByUserID int64) (string, error) {
	// Get user info for metadata
	user, err := s.userRepo.GetByID(createdByUserID)
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("actalog_backup_%s.zip", timestamp)
	filePath := filepath.Join(s.backupDir, filename)

	// Create ZIP file
	zipFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Export all tables to JSON
	backupData, err := s.exportAllTables()
	if err != nil {
		return "", fmt.Errorf("failed to export tables: %w", err)
	}

	// Create metadata
	metadata := domain.BackupMetadata{
		Filename:       filename,
		CreatedAt:      time.Now(),
		CreatedByEmail: user.Email,
		Version:        version.String(),
		DatabaseDriver: s.dbDriver,
		DatabaseName:   s.dbName,
		TotalUsers:     len(backupData.Users),
		TotalWorkouts:  len(backupData.UserWorkouts),
		TotalMovements: len(backupData.Movements),
		TotalWODs:      len(backupData.WODs),
	}
	backupData.Metadata = metadata

	// Write backup data to ZIP
	dataJSON, err := json.MarshalIndent(backupData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal backup data: %w", err)
	}

	dataFile, err := zipWriter.Create("backup_data.json")
	if err != nil {
		return "", fmt.Errorf("failed to create data file in ZIP: %w", err)
	}
	if _, err := dataFile.Write(dataJSON); err != nil {
		return "", fmt.Errorf("failed to write data to ZIP: %w", err)
	}

	// Create SQLite database dump and add to ZIP
	sqlitePath := filepath.Join(s.backupDir, "temp_backup.db")
	defer os.Remove(sqlitePath) // Clean up temporary file

	if err := s.createSQLiteDump(backupData, sqlitePath); err != nil {
		// Log warning but don't fail the backup
		fmt.Printf("Warning: failed to create SQLite dump: %v\n", err)
	} else {
		// Add SQLite database to ZIP
		sqliteFile, err := os.Open(sqlitePath)
		if err != nil {
			fmt.Printf("Warning: failed to open SQLite dump: %v\n", err)
		} else {
			defer sqliteFile.Close()

			// Create file in ZIP named actalog_backup.db
			dbWriter, err := zipWriter.Create("actalog_backup.db")
			if err != nil {
				fmt.Printf("Warning: failed to create db file in ZIP: %v\n", err)
			} else {
				if _, err := io.Copy(dbWriter, sqliteFile); err != nil {
					fmt.Printf("Warning: failed to write db to ZIP: %v\n", err)
				}
			}
		}
	}

	// Add uploaded files to ZIP (profile pictures, etc.)
	if err := s.addUploadsToZip(zipWriter); err != nil {
		return "", fmt.Errorf("failed to add uploads to ZIP: %w", err)
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	// Create audit log
	if err := s.auditLogRepo.Create(&domain.AuditLog{
		UserID:    &createdByUserID,
		EventType: "backup_created",
		Details:   stringPtr(fmt.Sprintf("Created backup: %s (size: %d bytes)", filename, fileInfo.Size())),
		CreatedAt: time.Now(),
	}); err != nil {
		// Log error but don't fail the backup
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	return filename, nil
}

// ListBackups returns metadata for all available backups
func (s *BackupServiceImpl) ListBackups() ([]domain.BackupMetadata, error) {
	files, err := os.ReadDir(s.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []domain.BackupMetadata{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []domain.BackupMetadata
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".zip" {
			continue
		}

		metadata, err := s.GetBackupMetadata(file.Name())
		if err != nil {
			// Skip files with invalid metadata
			continue
		}

		// Get file size
		filePath := filepath.Join(s.backupDir, file.Name())
		fileInfo, err := os.Stat(filePath)
		if err == nil {
			metadata.FileSize = fileInfo.Size()
			metadata.FilePath = filePath
		}

		backups = append(backups, *metadata)
	}

	return backups, nil
}

// GetBackupMetadata reads and returns metadata for a specific backup
func (s *BackupServiceImpl) GetBackupMetadata(filename string) (*domain.BackupMetadata, error) {
	filePath := filepath.Join(s.backupDir, filename)

	// Open ZIP file
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}
	defer zipReader.Close()

	// Find backup_data.json
	var dataFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == "backup_data.json" {
			dataFile = f
			break
		}
	}
	if dataFile == nil {
		return nil, fmt.Errorf("backup_data.json not found in backup file")
	}

	// Read and parse backup data
	rc, err := dataFile.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open backup data file: %w", err)
	}
	defer rc.Close()

	var backupData domain.BackupData
	if err := json.NewDecoder(rc).Decode(&backupData); err != nil {
		return nil, fmt.Errorf("failed to parse backup data: %w", err)
	}

	metadata := backupData.Metadata
	metadata.Filename = filename

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err == nil {
		metadata.FileSize = fileInfo.Size()
		metadata.FilePath = filePath
	}

	return &metadata, nil
}

// DownloadBackup returns the file path for downloading a backup
func (s *BackupServiceImpl) DownloadBackup(filename string) (string, error) {
	filePath := filepath.Join(s.backupDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("backup file not found: %s", filename)
		}
		return "", fmt.Errorf("failed to access backup file: %w", err)
	}

	return filePath, nil
}

// UploadBackup saves an uploaded backup file to the backups directory
func (s *BackupServiceImpl) UploadBackup(file interface{}, originalFilename string, uploadedByUserID int64) (string, error) {
	// Type assert to multipart.File
	uploadedFile, ok := file.(multipart.File)
	if !ok {
		return "", fmt.Errorf("invalid file type")
	}

	// Generate unique filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	ext := filepath.Ext(originalFilename)
	baseFilename := strings.TrimSuffix(filepath.Base(originalFilename), ext)

	// Sanitize filename (remove any path components)
	baseFilename = filepath.Base(baseFilename)

	// Create new filename: original_timestamp.zip
	newFilename := fmt.Sprintf("%s_%s%s", baseFilename, timestamp, ext)
	filePath := filepath.Join(s.backupDir, newFilename)

	// Create destination file
	destFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy uploaded file to destination
	if _, err := io.Copy(destFile, uploadedFile); err != nil {
		os.Remove(filePath) // Clean up on error
		return "", fmt.Errorf("failed to save uploaded file: %w", err)
	}

	// Validate that it's a valid ZIP file
	if _, err := zip.OpenReader(filePath); err != nil {
		os.Remove(filePath) // Clean up invalid file
		return "", fmt.Errorf("uploaded file is not a valid ZIP archive: %w", err)
	}

	// Create audit log
	if err := s.auditLogRepo.Create(&domain.AuditLog{
		UserID:    &uploadedByUserID,
		EventType: "backup_uploaded",
		Details:   stringPtr(fmt.Sprintf("Uploaded backup: %s (original: %s)", newFilename, originalFilename)),
		CreatedAt: time.Now(),
	}); err != nil {
		// Log error but don't fail the upload
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	return newFilename, nil
}

// DeleteBackup removes a backup file and logs the deletion
func (s *BackupServiceImpl) DeleteBackup(filename string, deletedByUserID int64) error {
	filePath := filepath.Join(s.backupDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("backup file not found: %s", filename)
		}
		return fmt.Errorf("failed to access backup file: %w", err)
	}

	// Delete file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete backup file: %w", err)
	}

	// Create audit log
	if err := s.auditLogRepo.Create(&domain.AuditLog{
		UserID:    &deletedByUserID,
		EventType: "backup_deleted",
		Details:   stringPtr(fmt.Sprintf("Deleted backup: %s", filename)),
		CreatedAt: time.Now(),
	}); err != nil {
		// Log error but don't fail the deletion
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	return nil
}

// idMappings holds old->new ID mappings for foreign key remapping during merge/skip restore
type idMappings struct {
	users         map[int64]int64
	movements     map[int64]int64
	wods          map[int64]int64
	workouts      map[int64]int64
	organizations map[int64]int64
	userWorkouts  map[int64]int64
}

func newIDMappings() *idMappings {
	return &idMappings{
		users:         make(map[int64]int64),
		movements:     make(map[int64]int64),
		wods:          make(map[int64]int64),
		workouts:      make(map[int64]int64),
		organizations: make(map[int64]int64),
		userWorkouts:  make(map[int64]int64),
	}
}

// RestoreBackup restores database from a backup file
func (s *BackupServiceImpl) RestoreBackup(filename string, restoredByUserID int64, mode domain.RestoreMode) (*domain.RestoreResult, error) {
	// Default to replace mode if not specified
	if mode == "" {
		mode = domain.RestoreModeReplace
	}

	result := &domain.RestoreResult{
		Mode:   mode,
		Errors: []string{},
	}

	filePath := filepath.Join(s.backupDir, filename)

	// Open ZIP file
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}
	defer zipReader.Close()

	// Find backup_data.json
	var dataFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == "backup_data.json" {
			dataFile = f
			break
		}
	}
	if dataFile == nil {
		return nil, fmt.Errorf("backup_data.json not found in backup file")
	}

	// Read and parse backup data
	rc, err := dataFile.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open backup data file: %w", err)
	}
	defer rc.Close()

	var backupData domain.BackupData
	if err := json.NewDecoder(rc).Decode(&backupData); err != nil {
		return nil, fmt.Errorf("failed to parse backup data: %w", err)
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// For replace mode, delete all existing data first
	if mode == domain.RestoreModeReplace {
		tables := []string{
			"notification_likes", "notifications", "organization_subscriptions",
			"user_subscriptions", "user_organizations", "data_change_logs",
			"user_workout_wods", "user_workout_movements", "workout_wods",
			"workout_movements", "user_workouts", "workouts", "wods",
			"movements", "organizations", "email_verification_tokens",
			"password_resets", "refresh_tokens", "user_settings", "audit_logs", "users",
		}

		for _, table := range tables {
			exists, err := s.tableExists(tx, table)
			if err != nil {
				return nil, fmt.Errorf("failed to check if table %s exists: %w", table, err)
			}
			if exists {
				if _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s", table)); err != nil {
					return nil, fmt.Errorf("failed to clear table %s: %w", table, err)
				}
			}
		}
	}

	// Initialize ID mappings for merge/skip modes
	mappings := newIDMappings()

	// Restore data in correct order for foreign keys
	// 1. Core tables with no foreign key dependencies
	if err := s.restoreTableWithMode(tx, "users", backupData.Users, backupData.Schema.Tables["users"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore users: %w", err)
	}
	if err := s.restoreTableWithMode(tx, "movements", backupData.Movements, backupData.Schema.Tables["movements"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore movements: %w", err)
	}
	if err := s.restoreTableWithMode(tx, "wods", backupData.WODs, backupData.Schema.Tables["wods"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore wods: %w", err)
	}
	if err := s.restoreTableWithMode(tx, "workouts", backupData.Workouts, backupData.Schema.Tables["workouts"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore workouts: %w", err)
	}
	if err := s.restoreTableWithMode(tx, "organizations", backupData.Organizations, backupData.Schema.Tables["organizations"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore organizations: %w", err)
	}
	result.TablesRestored = 5

	// 2. Tables depending on core tables
	if err := s.restoreTableWithMode(tx, "user_workouts", backupData.UserWorkouts, backupData.Schema.Tables["user_workouts"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore user_workouts: %w", err)
	}
	if err := s.restoreTableWithMode(tx, "workout_movements", backupData.WorkoutMovements, backupData.Schema.Tables["workout_movements"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore workout_movements: %w", err)
	}
	if err := s.restoreTableWithMode(tx, "workout_wods", backupData.WorkoutWODs, backupData.Schema.Tables["workout_wods"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore workout_wods: %w", err)
	}
	if err := s.restoreTableWithMode(tx, "user_workout_movements", backupData.UserWorkoutMovements, backupData.Schema.Tables["user_workout_movements"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore user_workout_movements: %w", err)
	}
	if err := s.restoreTableWithMode(tx, "user_workout_wods", backupData.UserWorkoutWODs, backupData.Schema.Tables["user_workout_wods"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore user_workout_wods: %w", err)
	}
	result.TablesRestored = 10

	// 3. User-related tables (skip security tokens in merge/skip mode)
	if mode == domain.RestoreModeReplace {
		if err := s.restoreTableWithMode(tx, "refresh_tokens", backupData.RefreshTokens, backupData.Schema.Tables["refresh_tokens"], mode, mappings, result); err != nil {
			return nil, fmt.Errorf("failed to restore refresh_tokens: %w", err)
		}
		if err := s.restoreTableWithMode(tx, "password_resets", backupData.PasswordResets, backupData.Schema.Tables["password_resets"], mode, mappings, result); err != nil {
			return nil, fmt.Errorf("failed to restore password_resets: %w", err)
		}
		if err := s.restoreTableWithMode(tx, "email_verification_tokens", backupData.EmailVerificationTokens, backupData.Schema.Tables["email_verification_tokens"], mode, mappings, result); err != nil {
			return nil, fmt.Errorf("failed to restore email_verification_tokens: %w", err)
		}
	}
	if err := s.restoreTableWithMode(tx, "user_settings", backupData.UserSettings, backupData.Schema.Tables["user_settings"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore user_settings: %w", err)
	}
	// Skip audit_logs and data_change_logs in merge/skip mode (append-only)
	if mode == domain.RestoreModeReplace {
		if err := s.restoreTableWithMode(tx, "audit_logs", backupData.AuditLogs, backupData.Schema.Tables["audit_logs"], mode, mappings, result); err != nil {
			return nil, fmt.Errorf("failed to restore audit_logs: %w", err)
		}
		if err := s.restoreTableWithMode(tx, "data_change_logs", backupData.DataChangeLogs, backupData.Schema.Tables["data_change_logs"], mode, mappings, result); err != nil {
			return nil, fmt.Errorf("failed to restore data_change_logs: %w", err)
		}
	}
	result.TablesRestored = 16

	// 4. Organization-related tables
	if err := s.restoreTableWithMode(tx, "user_organizations", backupData.UserOrganizations, backupData.Schema.Tables["user_organizations"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore user_organizations: %w", err)
	}
	if err := s.restoreTableWithMode(tx, "user_subscriptions", backupData.UserSubscriptions, backupData.Schema.Tables["user_subscriptions"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore user_subscriptions: %w", err)
	}
	if err := s.restoreTableWithMode(tx, "organization_subscriptions", backupData.OrganizationSubscriptions, backupData.Schema.Tables["organization_subscriptions"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore organization_subscriptions: %w", err)
	}
	result.TablesRestored = 19

	// 5. Notification tables
	if err := s.restoreTableWithMode(tx, "notifications", backupData.Notifications, backupData.Schema.Tables["notifications"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore notifications: %w", err)
	}
	if err := s.restoreTableWithMode(tx, "notification_likes", backupData.NotificationLikes, backupData.Schema.Tables["notification_likes"], mode, mappings, result); err != nil {
		return nil, fmt.Errorf("failed to restore notification_likes: %w", err)
	}
	result.TablesRestored = 21

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Restore uploaded files
	if err := s.restoreUploadsFromZip(zipReader); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Warning: failed to restore uploads: %v", err))
	}

	// Create audit log
	details := fmt.Sprintf("Restored backup: %s (mode: %s, created: %d, updated: %d, skipped: %d)",
		filename, mode, result.RecordsCreated, result.RecordsUpdated, result.RecordsSkipped)
	if err := s.auditLogRepo.Create(&domain.AuditLog{
		UserID:    &restoredByUserID,
		EventType: "backup_restored",
		Details:   stringPtr(details),
		CreatedAt: time.Now(),
	}); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Warning: failed to create audit log: %v", err))
	}

	return result, nil
}

// restoreTableWithMode restores a table with the specified mode (replace/merge/skip)
func (s *BackupServiceImpl) restoreTableWithMode(tx *sql.Tx, tableName string, data []map[string]interface{}, sourceSchema domain.TableSchema, mode domain.RestoreMode, mappings *idMappings, result *domain.RestoreResult) error {
	if len(data) == 0 {
		return nil
	}

	// For replace mode, use the original simple insert
	if mode == domain.RestoreModeReplace {
		return s.restoreTable(tx, tableName, data, sourceSchema)
	}

	// For merge/skip modes, use natural key matching
	return s.mergeTable(tx, tableName, data, sourceSchema, mode, mappings, result)
}

// mergeTable handles merge/skip restore with natural key matching and ID remapping
func (s *BackupServiceImpl) mergeTable(tx *sql.Tx, tableName string, data []map[string]interface{}, sourceSchema domain.TableSchema, mode domain.RestoreMode, mappings *idMappings, result *domain.RestoreResult) error {
	// Check if table exists
	exists, err := s.tableExists(tx, tableName)
	if err != nil {
		return fmt.Errorf("failed to check if table %s exists: %w", tableName, err)
	}
	if !exists {
		fmt.Printf("Warning: table %s does not exist, skipping\n", tableName)
		return nil
	}

	// Get target columns
	targetColumns, err := s.getTableColumns(tx, tableName)
	if err != nil {
		return fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
	}

	// Build column type map from source schema
	columnTypes := make(map[string]string)
	for _, col := range sourceSchema.Columns {
		columnTypes[col.Name] = col.Type
	}

	// Process each row
	for _, row := range data {
		// Remap foreign key IDs
		s.remapForeignKeys(tableName, row, mappings)

		// Find existing record by natural key
		existingID, found, err := s.findByNaturalKey(tx, tableName, row, mappings)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Error finding %s: %v", tableName, err))
			continue
		}

		oldID := s.getInt64FromRow(row, "id")

		if found {
			if mode == domain.RestoreModeSkip {
				// Skip mode: just record the ID mapping and skip
				s.recordIDMapping(tableName, oldID, existingID, mappings)
				result.RecordsSkipped++
				continue
			}
			// Merge mode: update existing record
			if err := s.updateRecord(tx, tableName, existingID, row, targetColumns, columnTypes); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error updating %s: %v", tableName, err))
				continue
			}
			s.recordIDMapping(tableName, oldID, existingID, mappings)
			result.RecordsUpdated++
		} else {
			// Insert new record
			newID, err := s.insertRecord(tx, tableName, row, targetColumns, columnTypes)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error inserting %s: %v", tableName, err))
				continue
			}
			s.recordIDMapping(tableName, oldID, newID, mappings)
			result.RecordsCreated++
		}
	}

	// Reset sequence for PostgreSQL
	if err := s.resetSequence(tx, tableName); err != nil {
		fmt.Printf("Warning: sequence reset failed for %s: %v\n", tableName, err)
	}

	return nil
}

// remapForeignKeys updates foreign key references in a row based on ID mappings
func (s *BackupServiceImpl) remapForeignKeys(tableName string, row map[string]interface{}, mappings *idMappings) {
	switch tableName {
	case "user_workouts", "user_settings", "user_subscriptions", "notifications":
		if userID := s.getInt64FromRow(row, "user_id"); userID > 0 {
			if newID, ok := mappings.users[userID]; ok {
				row["user_id"] = newID
			}
		}
	case "workout_movements", "workout_wods":
		if workoutID := s.getInt64FromRow(row, "workout_id"); workoutID > 0 {
			if newID, ok := mappings.workouts[workoutID]; ok {
				row["workout_id"] = newID
			}
		}
		if movementID := s.getInt64FromRow(row, "movement_id"); movementID > 0 {
			if newID, ok := mappings.movements[movementID]; ok {
				row["movement_id"] = newID
			}
		}
		if wodID := s.getInt64FromRow(row, "wod_id"); wodID > 0 {
			if newID, ok := mappings.wods[wodID]; ok {
				row["wod_id"] = newID
			}
		}
	case "user_workout_movements":
		if uwID := s.getInt64FromRow(row, "user_workout_id"); uwID > 0 {
			if newID, ok := mappings.userWorkouts[uwID]; ok {
				row["user_workout_id"] = newID
			}
		}
		if movementID := s.getInt64FromRow(row, "movement_id"); movementID > 0 {
			if newID, ok := mappings.movements[movementID]; ok {
				row["movement_id"] = newID
			}
		}
	case "user_workout_wods":
		if uwID := s.getInt64FromRow(row, "user_workout_id"); uwID > 0 {
			if newID, ok := mappings.userWorkouts[uwID]; ok {
				row["user_workout_id"] = newID
			}
		}
		if wodID := s.getInt64FromRow(row, "wod_id"); wodID > 0 {
			if newID, ok := mappings.wods[wodID]; ok {
				row["wod_id"] = newID
			}
		}
	case "user_organizations":
		if userID := s.getInt64FromRow(row, "user_id"); userID > 0 {
			if newID, ok := mappings.users[userID]; ok {
				row["user_id"] = newID
			}
		}
		if orgID := s.getInt64FromRow(row, "organization_id"); orgID > 0 {
			if newID, ok := mappings.organizations[orgID]; ok {
				row["organization_id"] = newID
			}
		}
	case "organization_subscriptions":
		if orgID := s.getInt64FromRow(row, "organization_id"); orgID > 0 {
			if newID, ok := mappings.organizations[orgID]; ok {
				row["organization_id"] = newID
			}
		}
	case "notification_likes":
		if userID := s.getInt64FromRow(row, "user_id"); userID > 0 {
			if newID, ok := mappings.users[userID]; ok {
				row["user_id"] = newID
			}
		}
	}

	// Remap created_by for various tables
	if createdBy := s.getInt64FromRow(row, "created_by"); createdBy > 0 {
		if newID, ok := mappings.users[createdBy]; ok {
			row["created_by"] = newID
		}
	}
}

// findByNaturalKey finds an existing record by its natural key
func (s *BackupServiceImpl) findByNaturalKey(tx *sql.Tx, tableName string, row map[string]interface{}, mappings *idMappings) (int64, bool, error) {
	var query string
	var args []interface{}

	switch tableName {
	case "users":
		email, _ := row["email"].(string)
		if email == "" {
			return 0, false, nil
		}
		query = s.buildQuery("SELECT id FROM users WHERE email = ?", 1)
		args = []interface{}{email}

	case "movements":
		name, _ := row["name"].(string)
		if name == "" {
			return 0, false, nil
		}
		query = s.buildQuery("SELECT id FROM movements WHERE name = ?", 1)
		args = []interface{}{name}

	case "wods":
		name, _ := row["name"].(string)
		if name == "" {
			return 0, false, nil
		}
		query = s.buildQuery("SELECT id FROM wods WHERE name = ?", 1)
		args = []interface{}{name}

	case "workouts":
		name, _ := row["name"].(string)
		if name == "" {
			return 0, false, nil
		}
		query = s.buildQuery("SELECT id FROM workouts WHERE name = ?", 1)
		args = []interface{}{name}

	case "organizations":
		name, _ := row["name"].(string)
		if name == "" {
			return 0, false, nil
		}
		query = s.buildQuery("SELECT id FROM organizations WHERE name = ?", 1)
		args = []interface{}{name}

	case "user_workouts":
		userID := s.getInt64FromRow(row, "user_id")
		workoutDate, _ := row["workout_date"].(string)
		workoutName, _ := row["workout_name"].(string)
		if userID == 0 || workoutDate == "" {
			return 0, false, nil
		}
		if workoutName == "" {
			query = s.buildQuery("SELECT id FROM user_workouts WHERE user_id = ? AND workout_date = ? AND workout_name IS NULL", 2)
			args = []interface{}{userID, workoutDate}
		} else {
			query = s.buildQuery("SELECT id FROM user_workouts WHERE user_id = ? AND workout_date = ? AND workout_name = ?", 3)
			args = []interface{}{userID, workoutDate, workoutName}
		}

	case "user_settings":
		userID := s.getInt64FromRow(row, "user_id")
		if userID == 0 {
			return 0, false, nil
		}
		query = s.buildQuery("SELECT id FROM user_settings WHERE user_id = ?", 1)
		args = []interface{}{userID}

	case "user_organizations":
		userID := s.getInt64FromRow(row, "user_id")
		orgID := s.getInt64FromRow(row, "organization_id")
		if userID == 0 || orgID == 0 {
			return 0, false, nil
		}
		query = s.buildQuery("SELECT id FROM user_organizations WHERE user_id = ? AND organization_id = ?", 2)
		args = []interface{}{userID, orgID}

	case "user_subscriptions":
		userID := s.getInt64FromRow(row, "user_id")
		if userID == 0 {
			return 0, false, nil
		}
		query = s.buildQuery("SELECT id FROM user_subscriptions WHERE user_id = ?", 1)
		args = []interface{}{userID}

	case "organization_subscriptions":
		orgID := s.getInt64FromRow(row, "organization_id")
		if orgID == 0 {
			return 0, false, nil
		}
		query = s.buildQuery("SELECT id FROM organization_subscriptions WHERE organization_id = ?", 1)
		args = []interface{}{orgID}

	default:
		// For tables without natural keys, always insert new
		return 0, false, nil
	}

	var id int64
	err := tx.QueryRow(query, args...).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

// buildQuery converts ? placeholders to $1, $2, etc. for PostgreSQL
func (s *BackupServiceImpl) buildQuery(query string, paramCount int) string {
	if s.dbDriver != "postgres" {
		return query
	}
	for i := 1; i <= paramCount; i++ {
		query = strings.Replace(query, "?", fmt.Sprintf("$%d", i), 1)
	}
	return query
}

// updateRecord updates an existing record
func (s *BackupServiceImpl) updateRecord(tx *sql.Tx, tableName string, id int64, row map[string]interface{}, targetColumns []string, columnTypes map[string]string) error {
	var setClauses []string
	var values []interface{}
	paramIndex := 1

	for col, val := range row {
		if col == "id" {
			continue // Don't update ID
		}
		if !containsString(targetColumns, col) {
			continue
		}

		colType := columnTypes[col]
		convertedValue := s.convertValue(val, col, colType)

		if s.dbDriver == "postgres" {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, paramIndex))
		} else {
			setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
		}
		values = append(values, convertedValue)
		paramIndex++
	}

	if len(setClauses) == 0 {
		return nil
	}

	values = append(values, id)
	var query string
	if s.dbDriver == "postgres" {
		query = fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", tableName, joinStrings(setClauses, ", "), paramIndex)
	} else {
		query = fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", tableName, joinStrings(setClauses, ", "))
	}

	_, err := tx.Exec(query, values...)
	return err
}

// insertRecord inserts a new record and returns the new ID
func (s *BackupServiceImpl) insertRecord(tx *sql.Tx, tableName string, row map[string]interface{}, targetColumns []string, columnTypes map[string]string) (int64, error) {
	var columns []string
	var placeholders []string
	var values []interface{}
	paramIndex := 1

	for col, val := range row {
		if col == "id" {
			continue // Let database generate ID
		}
		if !containsString(targetColumns, col) {
			continue
		}

		columns = append(columns, col)
		colType := columnTypes[col]
		convertedValue := s.convertValue(val, col, colType)

		if s.dbDriver == "postgres" {
			placeholders = append(placeholders, fmt.Sprintf("$%d", paramIndex))
		} else {
			placeholders = append(placeholders, "?")
		}
		values = append(values, convertedValue)
		paramIndex++
	}

	if len(columns) == 0 {
		return 0, nil
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName, joinStrings(columns, ", "), joinStrings(placeholders, ", "))

	// Get the new ID
	if s.dbDriver == "postgres" {
		query += " RETURNING id"
		var newID int64
		err := tx.QueryRow(query, values...).Scan(&newID)
		return newID, err
	}

	result, err := tx.Exec(query, values...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// recordIDMapping stores the old->new ID mapping for a table
func (s *BackupServiceImpl) recordIDMapping(tableName string, oldID, newID int64, mappings *idMappings) {
	if oldID == 0 {
		return
	}
	switch tableName {
	case "users":
		mappings.users[oldID] = newID
	case "movements":
		mappings.movements[oldID] = newID
	case "wods":
		mappings.wods[oldID] = newID
	case "workouts":
		mappings.workouts[oldID] = newID
	case "organizations":
		mappings.organizations[oldID] = newID
	case "user_workouts":
		mappings.userWorkouts[oldID] = newID
	}
}

// getInt64FromRow extracts an int64 value from a row map
func (s *BackupServiceImpl) getInt64FromRow(row map[string]interface{}, key string) int64 {
	val, ok := row[key]
	if !ok || val == nil {
		return 0
	}
	switch v := val.(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	case int:
		return int64(v)
	}
	return 0
}

// exportAllTables exports all database tables to JSON with schema metadata
func (s *BackupServiceImpl) exportAllTables() (*domain.BackupData, error) {
	data := &domain.BackupData{}

	// Initialize schema metadata
	data.Schema = domain.SchemaMetadata{
		Version: version.Version(),
		Tables:  make(map[string]domain.TableSchema),
	}

	tables := []struct {
		name   string
		target *[]map[string]interface{}
	}{
		{"users", &data.Users},
		{"movements", &data.Movements},
		{"wods", &data.WODs},
		{"workouts", &data.Workouts},
		{"user_workouts", &data.UserWorkouts},
		{"workout_movements", &data.WorkoutMovements},
		{"workout_wods", &data.WorkoutWODs},
		{"user_workout_movements", &data.UserWorkoutMovements},
		{"user_workout_wods", &data.UserWorkoutWODs},
		{"refresh_tokens", &data.RefreshTokens},
		{"password_resets", &data.PasswordResets},
		{"email_verification_tokens", &data.EmailVerificationTokens},
		{"audit_logs", &data.AuditLogs},
		{"user_settings", &data.UserSettings},
		{"data_change_logs", &data.DataChangeLogs},
		{"organizations", &data.Organizations},
		{"user_organizations", &data.UserOrganizations},
		{"user_subscriptions", &data.UserSubscriptions},
		{"organization_subscriptions", &data.OrganizationSubscriptions},
		{"notifications", &data.Notifications},
		{"notification_likes", &data.NotificationLikes},
	}

	for _, table := range tables {
		rows, err := s.db.Query(fmt.Sprintf("SELECT * FROM %s", table.name))
		if err != nil {
			// Skip tables that don't exist (they may not have been migrated yet)
			if isTableNotExistsError(err) {
				fmt.Printf("Skipping table %s (does not exist)\n", table.name)
				*table.target = []map[string]interface{}{}
				continue
			}
			return nil, fmt.Errorf("failed to query table %s: %w", table.name, err)
		}

		tableData, err := s.rowsToMaps(rows)
		rows.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to convert rows for table %s: %w", table.name, err)
		}

		*table.target = tableData

		// Capture schema metadata for this table
		tableSchema, err := s.getTableSchema(table.name)
		if err != nil {
			fmt.Printf("Warning: failed to get schema for table %s: %v\n", table.name, err)
		} else {
			tableSchema.RowCount = len(tableData)
			data.Schema.Tables[table.name] = tableSchema
		}
	}

	return data, nil
}

// getTableSchema retrieves column metadata for a table
func (s *BackupServiceImpl) getTableSchema(tableName string) (domain.TableSchema, error) {
	schema := domain.TableSchema{
		Columns: []domain.ColumnSchema{},
	}

	var query string
	switch s.dbDriver {
	case "sqlite3":
		query = fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	case "postgres":
		query = `SELECT column_name, data_type, is_nullable
			FROM information_schema.columns
			WHERE table_schema = current_schema() AND table_name = $1
			ORDER BY ordinal_position`
	case "mysql":
		query = `SELECT column_name, data_type, is_nullable
			FROM information_schema.columns
			WHERE table_schema = DATABASE() AND table_name = ?
			ORDER BY ordinal_position`
	default:
		return schema, fmt.Errorf("unsupported database driver: %s", s.dbDriver)
	}

	var rows *sql.Rows
	var err error
	if s.dbDriver == "sqlite3" {
		rows, err = s.db.Query(query)
	} else {
		rows, err = s.db.Query(query, tableName)
	}
	if err != nil {
		return schema, fmt.Errorf("failed to query table schema: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var col domain.ColumnSchema

		if s.dbDriver == "sqlite3" {
			// SQLite PRAGMA returns: cid, name, type, notnull, dflt_value, pk
			var cid int
			var colType string
			var notNull, pk int
			var dfltValue sql.NullString
			if err := rows.Scan(&cid, &col.Name, &colType, &notNull, &dfltValue, &pk); err != nil {
				return schema, fmt.Errorf("failed to scan column info: %w", err)
			}
			col.Type = s.normalizeColumnType(colType)
			col.Nullable = notNull == 0
		} else {
			// PostgreSQL and MySQL return: column_name, data_type, is_nullable
			var dataType, isNullable string
			if err := rows.Scan(&col.Name, &dataType, &isNullable); err != nil {
				return schema, fmt.Errorf("failed to scan column info: %w", err)
			}
			col.Type = s.normalizeColumnType(dataType)
			col.Nullable = strings.ToUpper(isNullable) == "YES"
		}

		schema.Columns = append(schema.Columns, col)
	}

	return schema, rows.Err()
}

// normalizeColumnType converts database-specific type names to generic types
func (s *BackupServiceImpl) normalizeColumnType(dbType string) string {
	dbType = strings.ToLower(dbType)

	// Boolean types - MUST check before integer types since tinyint(1) contains "int"
	if strings.Contains(dbType, "bool") || strings.HasPrefix(dbType, "tinyint(1)") {
		return "boolean"
	}

	// Integer types
	if strings.Contains(dbType, "int") || strings.Contains(dbType, "serial") {
		return "integer"
	}

	// Float/decimal types
	if strings.Contains(dbType, "float") || strings.Contains(dbType, "double") ||
		strings.Contains(dbType, "decimal") || strings.Contains(dbType, "numeric") ||
		strings.Contains(dbType, "real") {
		return "float"
	}

	// Date types (without time)
	if dbType == "date" {
		return "date"
	}

	// Datetime types
	if strings.Contains(dbType, "datetime") || strings.Contains(dbType, "timestamp") {
		return "datetime"
	}

	// Text/string types
	if strings.Contains(dbType, "char") || strings.Contains(dbType, "text") ||
		strings.Contains(dbType, "varchar") || strings.Contains(dbType, "clob") {
		return "string"
	}

	// Default to string for unknown types
	return "string"
}

// isTableNotExistsError checks if the error is a "table does not exist" error
func isTableNotExistsError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	// Check for different database drivers' "table does not exist" errors
	// SQLite: "no such table: tablename"
	// PostgreSQL: "relation \"tablename\" does not exist"
	// MySQL/MariaDB: "Error 1146 (42S02): Table 'db.tablename' doesn't exist"

	// Check for MySQL error code 1146 (most reliable for MySQL/MariaDB)
	if strings.Contains(errMsg, "1146") {
		return true
	}
	// Check for SQLite error
	if strings.Contains(errMsg, "no such table") {
		return true
	}
	// Check for PostgreSQL error
	if strings.Contains(errMsg, "does not exist") {
		return true
	}
	// Check for various "doesn't exist" patterns (handle different apostrophe encodings)
	if strings.Contains(errMsg, "doesn't exist") || strings.Contains(errMsg, "doesnt exist") {
		return true
	}
	return false
}

// rowsToMaps converts SQL rows to slice of maps
func (s *BackupServiceImpl) rowsToMaps(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Convert []byte to string for JSON serialization
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return results, rows.Err()
}

// restoreTable restores a single table from backup data with schema evolution support
func (s *BackupServiceImpl) restoreTable(tx *sql.Tx, tableName string, data []map[string]interface{}, sourceSchema domain.TableSchema) error {
	if len(data) == 0 {
		return nil
	}

	// Check if table exists before trying to restore (database-agnostic)
	exists, err := s.tableExists(tx, tableName)
	if err != nil {
		return fmt.Errorf("failed to check if table %s exists: %w", tableName, err)
	}

	// Skip if table doesn't exist (forward compatibility)
	if !exists {
		fmt.Printf("Warning: table %s does not exist in target schema, skipping restore for this table\n", tableName)
		return nil
	}

	// Get actual columns in target table for schema evolution support
	targetColumns, err := s.getTableColumns(tx, tableName)
	if err != nil {
		return fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
	}

	// Build column type map from source schema for type-aware conversion
	columnTypes := make(map[string]string)
	for _, col := range sourceSchema.Columns {
		columnTypes[col.Name] = col.Type
	}

	// Restore each row
	for rowIdx, row := range data {
		// Filter backup data to only include columns that exist in target schema
		columns := make([]string, 0, len(row))
		placeholders := make([]string, 0, len(row))
		values := make([]interface{}, 0, len(row))

		i := 1
		skippedColumns := 0
		for col, val := range row {
			// Only include column if it exists in target schema (schema evolution support)
			if !containsString(targetColumns, col) {
				skippedColumns++
				continue
			}

			columns = append(columns, col)
			if s.dbDriver == "postgres" {
				placeholders = append(placeholders, fmt.Sprintf("$%d", i))
			} else {
				placeholders = append(placeholders, "?")
			}

			// Convert value for database compatibility using schema metadata
			colType := columnTypes[col]
			convertedValue := s.convertValue(val, col, colType)
			values = append(values, convertedValue)
			i++
		}

		// Log schema evolution warnings
		if skippedColumns > 0 && rowIdx == 0 {
			fmt.Printf("Info: table %s - skipped %d column(s) not present in target schema (schema evolution)\n", tableName, skippedColumns)
		}

		// Skip row if no valid columns remain
		if len(columns) == 0 {
			continue
		}

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			tableName,
			joinStrings(columns, ", "),
			joinStrings(placeholders, ", "),
		)

		if _, err := tx.Exec(query, values...); err != nil {
			return fmt.Errorf("failed to insert row into %s: %w", tableName, err)
		}
	}

	// Reset auto-increment sequence for PostgreSQL
	if err := s.resetSequence(tx, tableName); err != nil {
		// Log warning but don't fail restore
		fmt.Printf("Warning: sequence reset failed for %s: %v\n", tableName, err)
	}

	return nil
}

// addUploadsToZip adds uploaded files to the backup ZIP
func (s *BackupServiceImpl) addUploadsToZip(zipWriter *zip.Writer) error {
	if _, err := os.Stat(s.uploadsDir); os.IsNotExist(err) {
		return nil // No uploads directory, nothing to backup
	}

	return filepath.Walk(s.uploadsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(s.uploadsDir, path)
		if err != nil {
			return err
		}

		// Create file in ZIP
		zipPath := filepath.Join("uploads", relPath)
		writer, err := zipWriter.Create(zipPath)
		if err != nil {
			return err
		}

		// Copy file contents
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(writer, file); err != nil {
			return err
		}

		return nil
	})
}

// restoreUploadsFromZip restores uploaded files from the backup ZIP
func (s *BackupServiceImpl) restoreUploadsFromZip(zipReader *zip.ReadCloser) error {
	// Ensure uploads directory exists
	if err := os.MkdirAll(s.uploadsDir, 0755); err != nil {
		return fmt.Errorf("failed to create uploads directory: %w", err)
	}

	for _, f := range zipReader.File {
		if !isUploadFile(f.Name) {
			continue
		}

		// Get destination path
		destPath := filepath.Join(s.uploadsDir, filepath.Base(f.Name))

		// Create destination file
		destFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", destPath, err)
		}

		// Open source file
		srcFile, err := f.Open()
		if err != nil {
			destFile.Close()
			return fmt.Errorf("failed to open file %s from ZIP: %w", f.Name, err)
		}

		// Copy contents
		if _, err := io.Copy(destFile, srcFile); err != nil {
			srcFile.Close()
			destFile.Close()
			return fmt.Errorf("failed to copy file %s: %w", f.Name, err)
		}

		srcFile.Close()
		destFile.Close()
	}

	return nil
}

// createSQLiteDump creates a SQLite database file from backup data
func (s *BackupServiceImpl) createSQLiteDump(backupData *domain.BackupData, outputPath string) error {
	// Remove existing file if it exists
	os.Remove(outputPath)

	// Create new SQLite database
	sqliteDB, err := sql.Open("sqlite3", outputPath)
	if err != nil {
		return fmt.Errorf("failed to create SQLite database: %w", err)
	}
	defer sqliteDB.Close()

	// Get schema from source database if it's SQLite, otherwise use basic schema
	var schema string
	if s.dbDriver == "sqlite3" {
		// Extract schema directly from source SQLite database
		schema, err = s.extractSQLiteSchema()
		if err != nil {
			return fmt.Errorf("failed to extract schema: %w", err)
		}
	} else {
		// For PostgreSQL/MySQL, create schema dynamically from data structure
		schema, err = s.createSchemaFromData(backupData)
		if err != nil {
			return fmt.Errorf("failed to create schema from data: %w", err)
		}
	}

	// Execute schema
	if _, err := sqliteDB.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Insert data using a transaction
	tx, err := sqliteDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Restore data in correct order (respecting foreign key constraints)
	// 1. Core tables with no foreign keys first
	if err := s.restoreTableToSQLite(tx, "users", backupData.Users); err != nil {
		return fmt.Errorf("failed to restore users: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "movements", backupData.Movements); err != nil {
		return fmt.Errorf("failed to restore movements: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "wods", backupData.WODs); err != nil {
		return fmt.Errorf("failed to restore wods: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "workouts", backupData.Workouts); err != nil {
		return fmt.Errorf("failed to restore workouts: %w", err)
	}
	// 2. Organizations (depends on users)
	if err := s.restoreTableToSQLite(tx, "organizations", backupData.Organizations); err != nil {
		return fmt.Errorf("failed to restore organizations: %w", err)
	}
	// 3. User-related tables
	if err := s.restoreTableToSQLite(tx, "user_workouts", backupData.UserWorkouts); err != nil {
		return fmt.Errorf("failed to restore user_workouts: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "workout_movements", backupData.WorkoutMovements); err != nil {
		return fmt.Errorf("failed to restore workout_movements: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "workout_wods", backupData.WorkoutWODs); err != nil {
		return fmt.Errorf("failed to restore workout_wods: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "user_workout_movements", backupData.UserWorkoutMovements); err != nil {
		return fmt.Errorf("failed to restore user_workout_movements: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "user_workout_wods", backupData.UserWorkoutWODs); err != nil {
		return fmt.Errorf("failed to restore user_workout_wods: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "refresh_tokens", backupData.RefreshTokens); err != nil {
		return fmt.Errorf("failed to restore refresh_tokens: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "password_resets", backupData.PasswordResets); err != nil {
		return fmt.Errorf("failed to restore password_resets: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "email_verification_tokens", backupData.EmailVerificationTokens); err != nil {
		return fmt.Errorf("failed to restore email_verification_tokens: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "user_settings", backupData.UserSettings); err != nil {
		return fmt.Errorf("failed to restore user_settings: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "audit_logs", backupData.AuditLogs); err != nil {
		return fmt.Errorf("failed to restore audit_logs: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "data_change_logs", backupData.DataChangeLogs); err != nil {
		return fmt.Errorf("failed to restore data_change_logs: %w", err)
	}
	// 4. Organization-related tables
	if err := s.restoreTableToSQLite(tx, "user_organizations", backupData.UserOrganizations); err != nil {
		return fmt.Errorf("failed to restore user_organizations: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "user_subscriptions", backupData.UserSubscriptions); err != nil {
		return fmt.Errorf("failed to restore user_subscriptions: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "organization_subscriptions", backupData.OrganizationSubscriptions); err != nil {
		return fmt.Errorf("failed to restore organization_subscriptions: %w", err)
	}
	// 5. Notifications (depends on users and organizations)
	if err := s.restoreTableToSQLite(tx, "notifications", backupData.Notifications); err != nil {
		return fmt.Errorf("failed to restore notifications: %w", err)
	}
	if err := s.restoreTableToSQLite(tx, "notification_likes", backupData.NotificationLikes); err != nil {
		return fmt.Errorf("failed to restore notification_likes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// extractSQLiteSchema extracts the complete schema from a SQLite database
func (s *BackupServiceImpl) extractSQLiteSchema() (string, error) {
	// Query sqlite_master for all CREATE statements
	query := `SELECT sql FROM sqlite_master WHERE type IN ('table', 'index', 'trigger', 'view') AND name NOT LIKE 'sqlite_%' ORDER BY type DESC, name`

	rows, err := s.db.Query(query)
	if err != nil {
		return "", fmt.Errorf("failed to query schema: %w", err)
	}
	defer rows.Close()

	var schemaParts []string
	for rows.Next() {
		var sql sql.NullString
		if err := rows.Scan(&sql); err != nil {
			return "", fmt.Errorf("failed to scan schema: %w", err)
		}
		if sql.Valid && sql.String != "" {
			schemaParts = append(schemaParts, sql.String+";")
		}
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating schema: %w", err)
	}

	return joinStrings(schemaParts, "\n"), nil
}

// createSchemaFromData creates SQLite schema by analyzing the backup data structure
func (s *BackupServiceImpl) createSchemaFromData(backupData *domain.BackupData) (string, error) {
	// Complete SQLite schema matching the actual database structure
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		name TEXT NOT NULL,
		profile_image TEXT,
		birthday DATE,
		role TEXT NOT NULL DEFAULT 'user',
		email_verified INTEGER NOT NULL DEFAULT 0,
		email_verified_at DATETIME,
		failed_login_attempts INTEGER NOT NULL DEFAULT 0,
		locked_at DATETIME,
		locked_until DATETIME,
		account_disabled INTEGER NOT NULL DEFAULT 0,
		disabled_at DATETIME,
		disabled_by_user_id INTEGER,
		disable_reason TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		last_login_at DATETIME,
		FOREIGN KEY (disabled_by_user_id) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL,
		revoked_at DATETIME,
		device_info TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS user_settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER UNIQUE NOT NULL,
		notification_preferences TEXT,
		data_export_format TEXT DEFAULT 'json',
		theme TEXT DEFAULT 'light',
		font_family TEXT DEFAULT 'system',
		weight_unit TEXT DEFAULT 'lbs',
		distance_unit TEXT DEFAULT 'meters',
		timezone TEXT DEFAULT 'America/New_York',
		admin_user_event_notifications INTEGER DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		target_user_id INTEGER,
		event_type TEXT NOT NULL,
		ip_address TEXT,
		user_agent TEXT,
		details TEXT,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
		FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS workouts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		notes TEXT,
		created_by INTEGER,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS movements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		description TEXT,
		type TEXT NOT NULL,
		is_standard INTEGER NOT NULL DEFAULT 0,
		created_by INTEGER,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS wods (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		source TEXT,
		type TEXT,
		regime TEXT,
		score_type TEXT,
		description TEXT,
		url TEXT,
		notes TEXT,
		is_standard INTEGER NOT NULL DEFAULT 0,
		created_by INTEGER,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS workout_wods (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workout_id INTEGER NOT NULL,
		wod_id INTEGER NOT NULL,
		score_value TEXT,
		division TEXT,
		is_pr INTEGER NOT NULL DEFAULT 0,
		order_index INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT
	);

	CREATE TABLE IF NOT EXISTS user_workouts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		workout_id INTEGER,
		workout_name TEXT,
		workout_date DATE NOT NULL,
		workout_type TEXT,
		total_time INTEGER,
		notes TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE RESTRICT
	);

	CREATE TABLE IF NOT EXISTS user_workout_movements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_workout_id INTEGER NOT NULL,
		movement_id INTEGER NOT NULL,
		sets INTEGER,
		reps INTEGER,
		weight REAL,
		time INTEGER,
		distance REAL,
		notes TEXT,
		order_index INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		is_pr INTEGER NOT NULL DEFAULT 0,
		rpe INTEGER,
		FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (movement_id) REFERENCES movements(id) ON DELETE RESTRICT
	);

	CREATE TABLE IF NOT EXISTS user_workout_wods (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_workout_id INTEGER NOT NULL,
		wod_id INTEGER NOT NULL,
		score_type TEXT,
		score_value TEXT,
		time_seconds INTEGER,
		rounds INTEGER,
		reps INTEGER,
		weight REAL,
		notes TEXT,
		order_index INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		is_pr INTEGER NOT NULL DEFAULT 0,
		rpe INTEGER,
		FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE RESTRICT
	);

	CREATE TABLE IF NOT EXISTS workout_movements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workout_id INTEGER NOT NULL,
		movement_id INTEGER NOT NULL,
		weight REAL,
		sets INTEGER,
		reps INTEGER,
		time INTEGER,
		distance REAL,
		is_rx INTEGER NOT NULL DEFAULT 0,
		is_pr INTEGER NOT NULL DEFAULT 0,
		notes TEXT,
		order_index INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (movement_id) REFERENCES movements(id) ON DELETE RESTRICT
	);

	CREATE TABLE IF NOT EXISTS password_resets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		used_at DATETIME,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS email_verification_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		used_at DATETIME,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS data_change_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		entity_type TEXT NOT NULL,
		entity_id INTEGER NOT NULL,
		entity_name TEXT NOT NULL,
		operation TEXT NOT NULL,
		user_id INTEGER NOT NULL,
		user_email TEXT NOT NULL,
		before_values TEXT,
		after_values TEXT,
		ip_address TEXT,
		user_agent TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS organizations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS user_organizations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		organization_id INTEGER NOT NULL,
		role TEXT DEFAULT 'member',
		joined_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
		UNIQUE (user_id, organization_id)
	);

	CREATE TABLE IF NOT EXISTS user_subscriptions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		subscription_type TEXT NOT NULL,
		status TEXT NOT NULL,
		is_permanent_free INTEGER NOT NULL DEFAULT 0,
		start_date DATETIME NOT NULL,
		end_date DATETIME,
		last_payment_date DATETIME,
		next_billing_date DATETIME,
		cancelled_at DATETIME,
		cancelled_reason TEXT,
		notes TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_by_user_id INTEGER,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS organization_subscriptions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		organization_id INTEGER NOT NULL,
		subscription_type TEXT NOT NULL,
		status TEXT NOT NULL,
		is_permanent_free INTEGER NOT NULL DEFAULT 0,
		start_date DATETIME NOT NULL,
		end_date DATETIME,
		last_payment_date DATETIME,
		next_billing_date DATETIME,
		cancelled_at DATETIME,
		cancelled_reason TEXT,
		notes TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_by_user_id INTEGER,
		FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
		FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		organization_id INTEGER,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		message TEXT NOT NULL,
		data TEXT,
		read_at DATETIME,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS notification_likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		notification_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE (notification_id, user_id)
	);
	`

	return schema, nil
}

// restoreTableToSQLite restores a single table to SQLite database (always uses ? placeholders)
func (s *BackupServiceImpl) restoreTableToSQLite(tx *sql.Tx, tableName string, data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	for _, row := range data {
		// Build INSERT statement
		columns := make([]string, 0, len(row))
		placeholders := make([]string, 0, len(row))
		values := make([]interface{}, 0, len(row))

		for col, val := range row {
			columns = append(columns, col)
			placeholders = append(placeholders, "?")
			values = append(values, val)
		}

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			tableName,
			joinStrings(columns, ", "),
			joinStrings(placeholders, ", "),
		)

		if _, err := tx.Exec(query, values...); err != nil {
			return fmt.Errorf("failed to insert row into %s: %w", tableName, err)
		}
	}

	return nil
}

// Helper functions

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func isUploadFile(name string) bool {
	return len(name) > 8 && name[:8] == "uploads/"
}

// tableExists checks if a table exists in the database (database-agnostic)
func (s *BackupServiceImpl) tableExists(tx *sql.Tx, tableName string) (bool, error) {
	var exists bool
	var query string

	switch s.dbDriver {
	case "sqlite3":
		// SQLite uses sqlite_master
		query = "SELECT COUNT(*) > 0 FROM sqlite_master WHERE type='table' AND name=?"
	case "postgres":
		// PostgreSQL uses information_schema
		query = "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1)"
	case "mysql":
		// MySQL uses information_schema
		query = "SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?"
	default:
		return false, fmt.Errorf("unsupported database driver: %s", s.dbDriver)
	}

	err := tx.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}

	return exists, nil
}

// getTableColumns returns the list of column names for a table (database-agnostic)
func (s *BackupServiceImpl) getTableColumns(tx *sql.Tx, tableName string) ([]string, error) {
	var query string
	var args []interface{}

	switch s.dbDriver {
	case "sqlite3":
		// SQLite uses PRAGMA
		query = fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	case "postgres":
		// PostgreSQL uses information_schema
		query = "SELECT column_name FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = $1 ORDER BY ordinal_position"
		args = []interface{}{tableName}
	case "mysql":
		// MySQL uses information_schema
		query = "SELECT column_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? ORDER BY ordinal_position"
		args = []interface{}{tableName}
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", s.dbDriver)
	}

	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query table columns: %w", err)
	}
	defer rows.Close()

	var columns []string
	if s.dbDriver == "sqlite3" {
		// SQLite PRAGMA returns: cid, name, type, notnull, dflt_value, pk
		for rows.Next() {
			var cid int
			var name, colType string
			var notNull, pk int
			var dfltValue sql.NullString
			if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
				return nil, fmt.Errorf("failed to scan column info: %w", err)
			}
			columns = append(columns, name)
		}
	} else {
		// PostgreSQL and MySQL return column names directly
		for rows.Next() {
			var columnName string
			if err := rows.Scan(&columnName); err != nil {
				return nil, fmt.Errorf("failed to scan column name: %w", err)
			}
			columns = append(columns, columnName)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading table columns: %w", err)
	}

	return columns, nil
}

// resetSequence resets the auto-increment sequence for a table (PostgreSQL only)
func (s *BackupServiceImpl) resetSequence(tx *sql.Tx, tableName string) error {
	if s.dbDriver != "postgres" {
		return nil // Only needed for PostgreSQL
	}

	// List of tables with auto-increment id columns
	tablesWithSequences := map[string]bool{
		"users":                      true,
		"movements":                  true,
		"wods":                       true,
		"workouts":                   true,
		"user_workouts":              true,
		"workout_movements":          true,
		"workout_wods":               true,
		"user_workout_movements":     true,
		"user_workout_wods":          true,
		"refresh_tokens":             true,
		"password_resets":            true,
		"email_verification_tokens":  true,
		"audit_logs":                 true,
		"user_settings":              true,
		"data_change_logs":           true,
		"organizations":              true,
		"user_organizations":         true,
		"user_subscriptions":         true,
		"organization_subscriptions": true,
		"notifications":              true,
		"notification_likes":         true,
	}

	if !tablesWithSequences[tableName] {
		return nil // Table doesn't have a sequence
	}

	// Reset the sequence to max(id) + 1
	query := fmt.Sprintf(`
		SELECT setval(pg_get_serial_sequence('%s', 'id'),
		              COALESCE((SELECT MAX(id) FROM %s), 1),
		              true)
	`, tableName, tableName)

	if _, err := tx.Exec(query); err != nil {
		// Log warning but don't fail the restore
		fmt.Printf("Warning: failed to reset sequence for %s: %v\n", tableName, err)
	}

	return nil
}

// convertValue converts a value from source database type to target database type
// colType is the normalized type from schema metadata ("boolean", "datetime", "date", "integer", "float", "string")
// If colType is empty, falls back to column name-based detection for backward compatibility
func (s *BackupServiceImpl) convertValue(val interface{}, columnName string, colType string) interface{} {
	if val == nil {
		return nil
	}

	// Determine if this is a boolean column
	// Known boolean columns - these take precedence over schema metadata
	// (fixes backups created with buggy normalizeColumnType that marked tinyint(1) as integer)
	booleanColumns := map[string]bool{
		"is_pr":                            true,
		"is_rx":                            true,
		"is_template":                      true,
		"is_standard":                      true,
		"is_permanent_free":                true,
		"email_verified":                   true,
		"account_disabled":                 true,
		"notifications_enabled":            true,
		"admin_user_event_notifications":   true,
	}
	isBoolean := booleanColumns[columnName] || colType == "boolean"

	// Determine if this is a datetime/date column (prefer schema type, fallback to column name)
	isDatetime := colType == "datetime" || colType == "date"
	if !isDatetime && colType == "" {
		// Fallback for older backups without schema metadata
		datetimeColumns := map[string]bool{
			"created_at":        true,
			"updated_at":        true,
			"last_login_at":     true,
			"email_verified_at": true,
			"locked_at":         true,
			"locked_until":      true,
			"disabled_at":       true,
			"expires_at":        true,
			"used_at":           true,
			"revoked_at":        true,
			"workout_date":      true,
			"birthday":          true,
			"joined_at":         true,
			"start_date":        true,
			"end_date":          true,
			"last_payment_date": true,
			"next_billing_date": true,
			"cancelled_at":      true,
			"read_at":           true,
		}
		isDatetime = datetimeColumns[columnName]
	}

	// Convert datetime values for MySQL (ISO 8601 to MySQL format)
	if isDatetime && s.dbDriver == "mysql" {
		if strVal, ok := val.(string); ok && strVal != "" {
			converted := s.convertDatetimeForMySQL(strVal)
			if converted != "" {
				return converted
			}
		}
	}

	if isBoolean {
		// Convert between different boolean representations
		switch v := val.(type) {
		case int:
			// Plain int
			if s.dbDriver == "postgres" {
				return v != 0
			}
			return int64(v)
		case int8:
			if s.dbDriver == "postgres" {
				return v != 0
			}
			return int64(v)
		case int16:
			if s.dbDriver == "postgres" {
				return v != 0
			}
			return int64(v)
		case int32:
			if s.dbDriver == "postgres" {
				return v != 0
			}
			return int64(v)
		case int64:
			// Convert integer to boolean for PostgreSQL
			if s.dbDriver == "postgres" {
				return v != 0
			}
			return v
		case uint:
			if s.dbDriver == "postgres" {
				return v != 0
			}
			return int64(v)
		case uint8:
			if s.dbDriver == "postgres" {
				return v != 0
			}
			return int64(v)
		case uint16:
			if s.dbDriver == "postgres" {
				return v != 0
			}
			return int64(v)
		case uint32:
			if s.dbDriver == "postgres" {
				return v != 0
			}
			return int64(v)
		case uint64:
			if s.dbDriver == "postgres" {
				return v != 0
			}
			return int64(v)
		case bool:
			// Convert boolean to integer for SQLite/MySQL
			if s.dbDriver == "sqlite3" || s.dbDriver == "mysql" {
				if v {
					return int64(1)
				}
				return int64(0)
			}
			return v
		case float32:
			if s.dbDriver == "postgres" {
				return v != 0
			}
			if v != 0 {
				return int64(1)
			}
			return int64(0)
		case float64:
			// JSON unmarshaling might give us float64 for numbers
			if s.dbDriver == "postgres" {
				return v != 0
			}
			if v != 0 {
				return int64(1)
			}
			return int64(0)
		case string:
			// Handle string representations of booleans
			if s.dbDriver == "postgres" {
				return v == "1" || v == "true" || v == "t" || v == "yes"
			}
			if v == "1" || v == "true" || v == "t" || v == "yes" {
				return int64(1)
			}
			return int64(0)
		default:
			// Last resort: try to convert via fmt to catch any other numeric types
			str := fmt.Sprintf("%v", v)
			if s.dbDriver == "postgres" {
				return str == "1" || str == "true" || str == "t" || str == "yes"
			}
			if str == "1" || str == "true" || str == "t" || str == "yes" {
				return int64(1)
			}
			return int64(0)
		}
	}

	return val
}

// convertDatetimeForMySQL converts various datetime formats to MySQL-compatible format
func (s *BackupServiceImpl) convertDatetimeForMySQL(dateStr string) string {
	// Try parsing various formats
	formats := []string{
		time.RFC3339Nano, // 2025-11-26T16:19:14.008192051Z
		time.RFC3339,     // 2025-11-26T16:19:14Z
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05.999999999", // SQLite format with microseconds
		"2006-01-02 15:04:05",           // Standard MySQL format
		"2006-01-02",                    // Date only
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			// Return MySQL-compatible format
			if format == "2006-01-02" {
				return t.Format("2006-01-02")
			}
			return t.Format("2006-01-02 15:04:05")
		}
	}

	// If we can't parse it, return as-is and let MySQL handle it
	return dateStr
}

// containsString checks if a slice contains a string
func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}
