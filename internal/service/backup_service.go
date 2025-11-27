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
	db             *sql.DB
	dbDriver       string
	dbName         string
	backupDir      string
	uploadsDir     string
	userRepo       domain.UserRepository
	auditLogRepo   domain.AuditLogRepository
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

// RestoreBackup restores database from a backup file
func (s *BackupServiceImpl) RestoreBackup(filename string, restoredByUserID int64) error {
	filePath := filepath.Join(s.backupDir, filename)

	// Open ZIP file
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
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
		return fmt.Errorf("backup_data.json not found in backup file")
	}

	// Read and parse backup data
	rc, err := dataFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open backup data file: %w", err)
	}
	defer rc.Close()

	var backupData domain.BackupData
	if err := json.NewDecoder(rc).Decode(&backupData); err != nil {
		return fmt.Errorf("failed to parse backup data: %w", err)
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete all existing data (in reverse order of foreign keys)
	tables := []string{
		"user_workout_wods",
		"user_workout_movements",
		"workout_wods",
		"workout_movements",
		"user_workouts",
		"workouts",
		"wods",
		"movements",
		"email_verification_tokens",
		"password_resets",
		"refresh_tokens",
		"user_settings",
		"audit_logs",
		"users",
	}

	for _, table := range tables {
		// Check if table exists before trying to delete (database-agnostic)
		exists, err := s.tableExists(tx, table)
		if err != nil {
			return fmt.Errorf("failed to check if table %s exists: %w", table, err)
		}

		// Only delete if table exists
		if exists {
			if _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s", table)); err != nil {
				return fmt.Errorf("failed to clear table %s: %w", table, err)
			}
		}
	}

	// Restore data (in correct order for foreign keys)
	if err := s.restoreTable(tx, "users", backupData.Users); err != nil {
		return fmt.Errorf("failed to restore users: %w", err)
	}
	if err := s.restoreTable(tx, "movements", backupData.Movements); err != nil {
		return fmt.Errorf("failed to restore movements: %w", err)
	}
	if err := s.restoreTable(tx, "wods", backupData.WODs); err != nil {
		return fmt.Errorf("failed to restore wods: %w", err)
	}
	if err := s.restoreTable(tx, "workouts", backupData.Workouts); err != nil {
		return fmt.Errorf("failed to restore workouts: %w", err)
	}
	if err := s.restoreTable(tx, "user_workouts", backupData.UserWorkouts); err != nil {
		return fmt.Errorf("failed to restore user_workouts: %w", err)
	}
	if err := s.restoreTable(tx, "workout_movements", backupData.WorkoutMovements); err != nil {
		return fmt.Errorf("failed to restore workout_movements: %w", err)
	}
	if err := s.restoreTable(tx, "workout_wods", backupData.WorkoutWODs); err != nil {
		return fmt.Errorf("failed to restore workout_wods: %w", err)
	}
	if err := s.restoreTable(tx, "user_workout_movements", backupData.UserWorkoutMovements); err != nil {
		return fmt.Errorf("failed to restore user_workout_movements: %w", err)
	}
	if err := s.restoreTable(tx, "user_workout_wods", backupData.UserWorkoutWODs); err != nil {
		return fmt.Errorf("failed to restore user_workout_wods: %w", err)
	}
	if err := s.restoreTable(tx, "refresh_tokens", backupData.RefreshTokens); err != nil {
		return fmt.Errorf("failed to restore refresh_tokens: %w", err)
	}
	if err := s.restoreTable(tx, "password_resets", backupData.PasswordResets); err != nil {
		return fmt.Errorf("failed to restore password_resets: %w", err)
	}
	if err := s.restoreTable(tx, "email_verification_tokens", backupData.EmailVerificationTokens); err != nil {
		return fmt.Errorf("failed to restore email_verification_tokens: %w", err)
	}
	if err := s.restoreTable(tx, "user_settings", backupData.UserSettings); err != nil {
		return fmt.Errorf("failed to restore user_settings: %w", err)
	}
	if err := s.restoreTable(tx, "audit_logs", backupData.AuditLogs); err != nil {
		return fmt.Errorf("failed to restore audit_logs: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Restore uploaded files
	if err := s.restoreUploadsFromZip(zipReader); err != nil {
		// Log error but don't fail the restore
		fmt.Printf("Warning: failed to restore uploads: %v\n", err)
	}

	// Create audit log (after restore, so it's in the new database)
	details := fmt.Sprintf("Restored backup: %s (users: %d, workouts: %d, movements: %d, WODs: %d)",
		filename,
		backupData.Metadata.TotalUsers,
		backupData.Metadata.TotalWorkouts,
		backupData.Metadata.TotalMovements,
		backupData.Metadata.TotalWODs,
	)
	if err := s.auditLogRepo.Create(&domain.AuditLog{
		UserID:    &restoredByUserID,
		EventType: "backup_restored",
		Details:   stringPtr(details),
		CreatedAt: time.Now(),
	}); err != nil {
		// Log error but don't fail the restore
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	return nil
}

// exportAllTables exports all database tables to JSON
func (s *BackupServiceImpl) exportAllTables() (*domain.BackupData, error) {
	data := &domain.BackupData{}

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
	}

	return data, nil
}

// isTableNotExistsError checks if the error is a "table does not exist" error
func isTableNotExistsError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// Check for different database drivers' "table does not exist" errors
	return (err.Error() == "no such table" ||
		err.Error() == "relation does not exist" ||
		err.Error() == "Table doesn't exist") ||
		(len(errMsg) > 0 && (
			errMsg[:13] == "no such table" ||
			errMsg[:22] == "relation does not exist" ||
			errMsg[:20] == "Table doesn't exist"))
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
func (s *BackupServiceImpl) restoreTable(tx *sql.Tx, tableName string, data []map[string]interface{}) error {
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

			// Convert value for database compatibility (e.g., boolean conversion)
			convertedValue := s.convertValue(val, col)
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

	// Restore data in correct order
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
	// For PostgreSQL/MySQL, create a basic schema
	// This is a simplified approach - in production, you might want to query information_schema
	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		birthday DATE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_login_at TIMESTAMP,
		email_verified INTEGER NOT NULL DEFAULT 0,
		email_verified_at TIMESTAMP,
		failed_login_attempts INTEGER NOT NULL DEFAULT 0,
		locked_at TIMESTAMP,
		locked_until TIMESTAMP,
		account_disabled INTEGER NOT NULL DEFAULT 0,
		disabled_at TIMESTAMP,
		disabled_by_user_id INTEGER,
		disable_reason TEXT,
		profile_image TEXT
	);

	CREATE TABLE movements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE wods (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT,
		score_type TEXT,
		description TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE workouts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		created_by_user_id INTEGER,
		is_template INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE TABLE user_workouts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		workout_date DATE NOT NULL,
		notes TEXT,
		total_time INTEGER,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE workout_movements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workout_id INTEGER NOT NULL,
		movement_id INTEGER NOT NULL,
		sets INTEGER,
		reps INTEGER,
		weight REAL,
		time INTEGER,
		distance REAL,
		notes TEXT,
		order_index INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (movement_id) REFERENCES movements(id) ON DELETE CASCADE
	);

	CREATE TABLE workout_wods (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workout_id INTEGER NOT NULL,
		wod_id INTEGER NOT NULL,
		score_type TEXT,
		score_value TEXT,
		time_seconds INTEGER,
		rounds INTEGER,
		reps INTEGER,
		weight REAL,
		notes TEXT,
		order_index INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE CASCADE
	);

	CREATE TABLE user_workout_movements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_workout_id INTEGER NOT NULL,
		movement_id INTEGER NOT NULL,
		sets INTEGER,
		reps INTEGER,
		weight REAL,
		time INTEGER,
		distance REAL,
		notes TEXT,
		is_pr INTEGER NOT NULL DEFAULT 0,
		order_index INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (movement_id) REFERENCES movements(id) ON DELETE CASCADE
	);

	CREATE TABLE user_workout_wods (
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
		is_pr INTEGER NOT NULL DEFAULT 0,
		order_index INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_workout_id) REFERENCES user_workouts(id) ON DELETE CASCADE,
		FOREIGN KEY (wod_id) REFERENCES wods(id) ON DELETE CASCADE
	);

	CREATE TABLE refresh_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		revoked_at TIMESTAMP,
		device_info TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE password_resets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at TIMESTAMP NOT NULL,
		used_at TIMESTAMP,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE email_verification_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at TIMESTAMP NOT NULL,
		used_at TIMESTAMP,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		event_type TEXT NOT NULL,
		details TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE TABLE user_settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL UNIQUE,
		theme TEXT NOT NULL DEFAULT 'light',
		notifications_enabled INTEGER NOT NULL DEFAULT 1,
		units_system TEXT NOT NULL DEFAULT 'imperial',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
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
func (s *BackupServiceImpl) convertValue(val interface{}, columnName string) interface{} {
	if val == nil {
		return nil
	}

	// Boolean conversion for specific columns
	booleanColumns := map[string]bool{
		"is_pr":                true,
		"is_template":          true,
		"is_standard":          true,
		"email_verified":       true,
		"account_disabled":     true,
		"notifications_enabled": true,
	}

	if booleanColumns[columnName] {
		// Convert between different boolean representations
		switch v := val.(type) {
		case int64:
			// Convert integer to boolean for PostgreSQL
			if s.dbDriver == "postgres" {
				return v != 0
			}
			return v
		case bool:
			// Convert boolean to integer for SQLite/MySQL
			if s.dbDriver == "sqlite3" || s.dbDriver == "mysql" {
				if v {
					return int64(1)
				}
				return int64(0)
			}
			return v
		case float64:
			// JSON unmarshaling might give us float64 for numbers
			if s.dbDriver == "postgres" {
				return v != 0
			}
			if v != 0 {
				return int64(1)
			}
			return int64(0)
		}
	}

	return val
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
