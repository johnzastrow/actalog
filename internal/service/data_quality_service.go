package service

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// EntityType represents the type of entity for duplicate detection
type EntityType string

const (
	EntityTypeMovement    EntityType = "movements"
	EntityTypeWOD         EntityType = "wods"
	EntityTypeUserWorkout EntityType = "user_workouts"
	EntityTypeUser        EntityType = "users"
	EntityTypeWorkout     EntityType = "workouts"
)

// DuplicateRecord represents a single record in a duplicate group
type DuplicateRecord struct {
	ID        int64                  `json:"id"`
	Name      string                 `json:"name,omitempty"`
	Email     string                 `json:"email,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	Details   map[string]interface{} `json:"details,omitempty"`
	FKCount   int                    `json:"fk_count"` // Number of FK references
}

// DuplicateGroup represents a group of duplicate records
type DuplicateGroup struct {
	Key       string             `json:"key"`     // The duplicate key (e.g., lowercase name)
	Count     int                `json:"count"`   // Number of duplicates
	Records   []*DuplicateRecord `json:"records"` // The duplicate records
	Canonical int64              `json:"canonical_id,omitempty"`
}

// DuplicateScanResult contains the results of scanning for duplicates
type DuplicateScanResult struct {
	EntityType EntityType        `json:"entity_type"`
	Groups     []*DuplicateGroup `json:"groups"`
	TotalDups  int               `json:"total_duplicates"`
	ScanTime   string            `json:"scan_time"`
}

// MergePreview shows what will happen when merging duplicates
type MergePreview struct {
	EntityType  EntityType        `json:"entity_type"`
	KeepID      int64             `json:"keep_id"`
	DeleteIDs   []int64           `json:"delete_ids"`
	FKUpdates   map[string]int    `json:"fk_updates"` // table_name -> count of updates
	CanProceed  bool              `json:"can_proceed"`
	Warnings    []string          `json:"warnings,omitempty"`
	AffectedIDs map[string][]int64 `json:"affected_ids,omitempty"` // table_name -> IDs that will be updated
}

// MergeResult contains the result of a merge operation
type MergeResult struct {
	EntityType    EntityType     `json:"entity_type"`
	KeepID        int64          `json:"keep_id"`
	DeletedIDs    []int64        `json:"deleted_ids"`
	DeletedCount  int            `json:"deleted_count"`
	FKUpdates     map[string]int `json:"fk_updates"`
	Success       bool           `json:"success"`
	Error         string         `json:"error,omitempty"`
}

// DataQualityIssue represents a data quality problem
type DataQualityIssue struct {
	IssueType   string                 `json:"issue_type"`
	EntityType  EntityType             `json:"entity_type"`
	RecordID    int64                  `json:"record_id"`
	Field       string                 `json:"field,omitempty"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Severity    string                 `json:"severity"` // "error", "warning", "info"
}

// DataQualityScanResult contains results of a data quality scan
type DataQualityScanResult struct {
	Issues     []*DataQualityIssue `json:"issues"`
	TotalCount int                 `json:"total_count"`
	ByType     map[string]int      `json:"by_type"`     // issue_type -> count
	BySeverity map[string]int      `json:"by_severity"` // severity -> count
	ScanTime   string              `json:"scan_time"`
}

// DataQualityService handles duplicate detection and data quality scanning
type DataQualityService struct {
	db              *sql.DB
	driver          string
	auditLogService *AuditLogService
}

// NewDataQualityService creates a new data quality service
func NewDataQualityService(db *sql.DB, driver string, auditLogService *AuditLogService) *DataQualityService {
	return &DataQualityService{
		db:              db,
		driver:          driver,
		auditLogService: auditLogService,
	}
}

// ScanForDuplicates scans a specific entity type for duplicates
func (s *DataQualityService) ScanForDuplicates(entityType EntityType) (*DuplicateScanResult, error) {
	start := time.Now()

	var groups []*DuplicateGroup
	var err error

	switch entityType {
	case EntityTypeMovement:
		groups, err = s.scanMovementDuplicates()
	case EntityTypeWOD:
		groups, err = s.scanWODDuplicates()
	case EntityTypeUserWorkout:
		groups, err = s.scanUserWorkoutDuplicates()
	case EntityTypeUser:
		groups, err = s.scanUserDuplicates()
	case EntityTypeWorkout:
		groups, err = s.scanWorkoutDuplicates()
	default:
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan %s duplicates: %w", entityType, err)
	}

	totalDups := 0
	for _, g := range groups {
		totalDups += g.Count
	}

	return &DuplicateScanResult{
		EntityType: entityType,
		Groups:     groups,
		TotalDups:  totalDups,
		ScanTime:   time.Since(start).String(),
	}, nil
}

// ScanAllDuplicates scans all entity types for duplicates
func (s *DataQualityService) ScanAllDuplicates() (map[EntityType]*DuplicateScanResult, error) {
	results := make(map[EntityType]*DuplicateScanResult)
	entityTypes := []EntityType{
		EntityTypeMovement,
		EntityTypeWOD,
		EntityTypeUserWorkout,
		EntityTypeUser,
		EntityTypeWorkout,
	}

	for _, et := range entityTypes {
		result, err := s.ScanForDuplicates(et)
		if err != nil {
			return nil, fmt.Errorf("failed to scan %s: %w", et, err)
		}
		results[et] = result
	}

	return results, nil
}

// scanMovementDuplicates scans for duplicate movements by case-insensitive name
func (s *DataQualityService) scanMovementDuplicates() ([]*DuplicateGroup, error) {
	// Query for duplicate movements by lowercase name
	var query string
	if s.driver == "postgres" {
		query = `
			SELECT LOWER(name) as key_name, COUNT(*) as cnt
			FROM movements
			GROUP BY LOWER(name)
			HAVING COUNT(*) > 1
			ORDER BY cnt DESC`
	} else {
		query = `
			SELECT LOWER(name) as key_name, COUNT(*) as cnt
			FROM movements
			GROUP BY LOWER(name)
			HAVING COUNT(*) > 1
			ORDER BY cnt DESC`
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query duplicates: %w", err)
	}
	defer rows.Close()

	var groups []*DuplicateGroup
	for rows.Next() {
		var keyName string
		var count int
		if err := rows.Scan(&keyName, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Get the actual records for this duplicate group
		records, err := s.getMovementRecords(keyName)
		if err != nil {
			return nil, err
		}

		groups = append(groups, &DuplicateGroup{
			Key:     keyName,
			Count:   count,
			Records: records,
		})
	}

	return groups, rows.Err()
}

// getMovementRecords retrieves movement records for a duplicate key
func (s *DataQualityService) getMovementRecords(keyName string) ([]*DuplicateRecord, error) {
	var query string
	if s.driver == "postgres" {
		query = `
			SELECT m.id, m.name, m.type, m.is_standard, m.created_at,
				(SELECT COUNT(*) FROM workout_movements wm WHERE wm.movement_id = m.id) +
				(SELECT COUNT(*) FROM user_workout_movements uwm WHERE uwm.movement_id = m.id) as fk_count
			FROM movements m
			WHERE LOWER(m.name) = $1
			ORDER BY m.created_at ASC`
	} else {
		query = `
			SELECT m.id, m.name, m.type, m.is_standard, m.created_at,
				(SELECT COUNT(*) FROM workout_movements wm WHERE wm.movement_id = m.id) +
				(SELECT COUNT(*) FROM user_workout_movements uwm WHERE uwm.movement_id = m.id) as fk_count
			FROM movements m
			WHERE LOWER(m.name) = ?
			ORDER BY m.created_at ASC`
	}

	rows, err := s.db.Query(query, keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to query movement records: %w", err)
	}
	defer rows.Close()

	var records []*DuplicateRecord
	for rows.Next() {
		var id int64
		var name, movType string
		var isStandard bool
		var createdAt time.Time
		var fkCount int

		if err := rows.Scan(&id, &name, &movType, &isStandard, &createdAt, &fkCount); err != nil {
			return nil, fmt.Errorf("failed to scan movement record: %w", err)
		}

		records = append(records, &DuplicateRecord{
			ID:        id,
			Name:      name,
			CreatedAt: createdAt,
			FKCount:   fkCount,
			Details: map[string]interface{}{
				"type":        movType,
				"is_standard": isStandard,
			},
		})
	}

	return records, rows.Err()
}

// scanWODDuplicates scans for duplicate WODs by case-insensitive name
func (s *DataQualityService) scanWODDuplicates() ([]*DuplicateGroup, error) {
	var query string
	if s.driver == "postgres" {
		query = `
			SELECT LOWER(name) as key_name, COUNT(*) as cnt
			FROM wods
			GROUP BY LOWER(name)
			HAVING COUNT(*) > 1
			ORDER BY cnt DESC`
	} else {
		query = `
			SELECT LOWER(name) as key_name, COUNT(*) as cnt
			FROM wods
			GROUP BY LOWER(name)
			HAVING COUNT(*) > 1
			ORDER BY cnt DESC`
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query WOD duplicates: %w", err)
	}
	defer rows.Close()

	var groups []*DuplicateGroup
	for rows.Next() {
		var keyName string
		var count int
		if err := rows.Scan(&keyName, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		records, err := s.getWODRecords(keyName)
		if err != nil {
			return nil, err
		}

		groups = append(groups, &DuplicateGroup{
			Key:     keyName,
			Count:   count,
			Records: records,
		})
	}

	return groups, rows.Err()
}

// getWODRecords retrieves WOD records for a duplicate key
func (s *DataQualityService) getWODRecords(keyName string) ([]*DuplicateRecord, error) {
	var query string
	if s.driver == "postgres" {
		query = `
			SELECT w.id, w.name, w.source, w.type, w.is_standard, w.created_at,
				(SELECT COUNT(*) FROM workout_wods ww WHERE ww.wod_id = w.id) +
				(SELECT COUNT(*) FROM user_workout_wods uww WHERE uww.wod_id = w.id) as fk_count
			FROM wods w
			WHERE LOWER(w.name) = $1
			ORDER BY w.created_at ASC`
	} else {
		query = `
			SELECT w.id, w.name, w.source, w.type, w.is_standard, w.created_at,
				(SELECT COUNT(*) FROM workout_wods ww WHERE ww.wod_id = w.id) +
				(SELECT COUNT(*) FROM user_workout_wods uww WHERE uww.wod_id = w.id) as fk_count
			FROM wods w
			WHERE LOWER(w.name) = ?
			ORDER BY w.created_at ASC`
	}

	rows, err := s.db.Query(query, keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to query WOD records: %w", err)
	}
	defer rows.Close()

	var records []*DuplicateRecord
	for rows.Next() {
		var id int64
		var name string
		var source, wodType sql.NullString
		var isStandard bool
		var createdAt time.Time
		var fkCount int

		if err := rows.Scan(&id, &name, &source, &wodType, &isStandard, &createdAt, &fkCount); err != nil {
			return nil, fmt.Errorf("failed to scan WOD record: %w", err)
		}

		records = append(records, &DuplicateRecord{
			ID:        id,
			Name:      name,
			CreatedAt: createdAt,
			FKCount:   fkCount,
			Details: map[string]interface{}{
				"source":      source.String,
				"type":        wodType.String,
				"is_standard": isStandard,
			},
		})
	}

	return records, rows.Err()
}

// scanUserWorkoutDuplicates scans for duplicate user workouts
func (s *DataQualityService) scanUserWorkoutDuplicates() ([]*DuplicateGroup, error) {
	var query string
	if s.driver == "postgres" {
		query = `
			SELECT user_id, workout_date, COALESCE(workout_name, ''), COUNT(*) as cnt
			FROM user_workouts
			GROUP BY user_id, workout_date, COALESCE(workout_name, '')
			HAVING COUNT(*) > 1
			ORDER BY cnt DESC`
	} else {
		query = `
			SELECT user_id, workout_date, COALESCE(workout_name, ''), COUNT(*) as cnt
			FROM user_workouts
			GROUP BY user_id, workout_date, COALESCE(workout_name, '')
			HAVING COUNT(*) > 1
			ORDER BY cnt DESC`
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query user workout duplicates: %w", err)
	}
	defer rows.Close()

	var groups []*DuplicateGroup
	for rows.Next() {
		var userID int64
		var workoutDate time.Time
		var workoutName string
		var count int
		if err := rows.Scan(&userID, &workoutDate, &workoutName, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		key := fmt.Sprintf("%d|%s|%s", userID, workoutDate.Format("2006-01-02"), workoutName)
		records, err := s.getUserWorkoutRecords(userID, workoutDate, workoutName)
		if err != nil {
			return nil, err
		}

		groups = append(groups, &DuplicateGroup{
			Key:     key,
			Count:   count,
			Records: records,
		})
	}

	return groups, rows.Err()
}

// getUserWorkoutRecords retrieves user workout records for a duplicate key
func (s *DataQualityService) getUserWorkoutRecords(userID int64, workoutDate time.Time, workoutName string) ([]*DuplicateRecord, error) {
	var query string
	if s.driver == "postgres" {
		query = `
			SELECT uw.id, uw.workout_name, uw.workout_type, uw.created_at,
				(SELECT COUNT(*) FROM user_workout_movements uwm WHERE uwm.user_workout_id = uw.id) +
				(SELECT COUNT(*) FROM user_workout_wods uww WHERE uww.user_workout_id = uw.id) as fk_count
			FROM user_workouts uw
			WHERE uw.user_id = $1 AND uw.workout_date = $2 AND COALESCE(uw.workout_name, '') = $3
			ORDER BY uw.created_at ASC`
	} else {
		query = `
			SELECT uw.id, uw.workout_name, uw.workout_type, uw.created_at,
				(SELECT COUNT(*) FROM user_workout_movements uwm WHERE uwm.user_workout_id = uw.id) +
				(SELECT COUNT(*) FROM user_workout_wods uww WHERE uww.user_workout_id = uw.id) as fk_count
			FROM user_workouts uw
			WHERE uw.user_id = ? AND uw.workout_date = ? AND COALESCE(uw.workout_name, '') = ?
			ORDER BY uw.created_at ASC`
	}

	rows, err := s.db.Query(query, userID, workoutDate, workoutName)
	if err != nil {
		return nil, fmt.Errorf("failed to query user workout records: %w", err)
	}
	defer rows.Close()

	var records []*DuplicateRecord
	for rows.Next() {
		var id int64
		var name, workoutType sql.NullString
		var createdAt time.Time
		var fkCount int

		if err := rows.Scan(&id, &name, &workoutType, &createdAt, &fkCount); err != nil {
			return nil, fmt.Errorf("failed to scan user workout record: %w", err)
		}

		records = append(records, &DuplicateRecord{
			ID:        id,
			Name:      name.String,
			CreatedAt: createdAt,
			FKCount:   fkCount,
			Details: map[string]interface{}{
				"user_id":      userID,
				"workout_date": workoutDate.Format("2006-01-02"),
				"workout_type": workoutType.String,
			},
		})
	}

	return records, rows.Err()
}

// scanUserDuplicates scans for duplicate users by case-insensitive email
func (s *DataQualityService) scanUserDuplicates() ([]*DuplicateGroup, error) {
	var query string
	if s.driver == "postgres" {
		query = `
			SELECT LOWER(email) as key_email, COUNT(*) as cnt
			FROM users
			GROUP BY LOWER(email)
			HAVING COUNT(*) > 1
			ORDER BY cnt DESC`
	} else {
		query = `
			SELECT LOWER(email) as key_email, COUNT(*) as cnt
			FROM users
			GROUP BY LOWER(email)
			HAVING COUNT(*) > 1
			ORDER BY cnt DESC`
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query user duplicates: %w", err)
	}
	defer rows.Close()

	var groups []*DuplicateGroup
	for rows.Next() {
		var keyEmail string
		var count int
		if err := rows.Scan(&keyEmail, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		records, err := s.getUserRecords(keyEmail)
		if err != nil {
			return nil, err
		}

		groups = append(groups, &DuplicateGroup{
			Key:     keyEmail,
			Count:   count,
			Records: records,
		})
	}

	return groups, rows.Err()
}

// getUserRecords retrieves user records for a duplicate key
func (s *DataQualityService) getUserRecords(keyEmail string) ([]*DuplicateRecord, error) {
	var query string
	if s.driver == "postgres" {
		query = `
			SELECT u.id, u.email, u.name, u.role, u.created_at,
				(SELECT COUNT(*) FROM user_workouts uw WHERE uw.user_id = u.id) as fk_count
			FROM users u
			WHERE LOWER(u.email) = $1
			ORDER BY u.created_at ASC`
	} else {
		query = `
			SELECT u.id, u.email, u.name, u.role, u.created_at,
				(SELECT COUNT(*) FROM user_workouts uw WHERE uw.user_id = u.id) as fk_count
			FROM users u
			WHERE LOWER(u.email) = ?
			ORDER BY u.created_at ASC`
	}

	rows, err := s.db.Query(query, keyEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to query user records: %w", err)
	}
	defer rows.Close()

	var records []*DuplicateRecord
	for rows.Next() {
		var id int64
		var email, name, role string
		var createdAt time.Time
		var fkCount int

		if err := rows.Scan(&id, &email, &name, &role, &createdAt, &fkCount); err != nil {
			return nil, fmt.Errorf("failed to scan user record: %w", err)
		}

		records = append(records, &DuplicateRecord{
			ID:        id,
			Name:      name,
			Email:     email,
			CreatedAt: createdAt,
			FKCount:   fkCount,
			Details: map[string]interface{}{
				"role": role,
			},
		})
	}

	return records, rows.Err()
}

// scanWorkoutDuplicates scans for duplicate workout templates
func (s *DataQualityService) scanWorkoutDuplicates() ([]*DuplicateGroup, error) {
	var query string
	if s.driver == "postgres" {
		query = `
			SELECT LOWER(name) as key_name, created_by, COUNT(*) as cnt
			FROM workouts
			GROUP BY LOWER(name), created_by
			HAVING COUNT(*) > 1
			ORDER BY cnt DESC`
	} else {
		query = `
			SELECT LOWER(name) as key_name, created_by, COUNT(*) as cnt
			FROM workouts
			GROUP BY LOWER(name), created_by
			HAVING COUNT(*) > 1
			ORDER BY cnt DESC`
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query workout duplicates: %w", err)
	}
	defer rows.Close()

	var groups []*DuplicateGroup
	for rows.Next() {
		var keyName string
		var createdBy sql.NullInt64
		var count int
		if err := rows.Scan(&keyName, &createdBy, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		key := keyName
		if createdBy.Valid {
			key = fmt.Sprintf("%s|user:%d", keyName, createdBy.Int64)
		}

		records, err := s.getWorkoutRecords(keyName, createdBy)
		if err != nil {
			return nil, err
		}

		groups = append(groups, &DuplicateGroup{
			Key:     key,
			Count:   count,
			Records: records,
		})
	}

	return groups, rows.Err()
}

// getWorkoutRecords retrieves workout records for a duplicate key
func (s *DataQualityService) getWorkoutRecords(keyName string, createdBy sql.NullInt64) ([]*DuplicateRecord, error) {
	var query string
	var args []interface{}

	if createdBy.Valid {
		if s.driver == "postgres" {
			query = `
				SELECT w.id, w.name, w.created_at,
					(SELECT COUNT(*) FROM workout_movements wm WHERE wm.workout_id = w.id) +
					(SELECT COUNT(*) FROM workout_wods ww WHERE ww.workout_id = w.id) +
					(SELECT COUNT(*) FROM user_workouts uw WHERE uw.workout_id = w.id) as fk_count
				FROM workouts w
				WHERE LOWER(w.name) = $1 AND w.created_by = $2
				ORDER BY w.created_at ASC`
		} else {
			query = `
				SELECT w.id, w.name, w.created_at,
					(SELECT COUNT(*) FROM workout_movements wm WHERE wm.workout_id = w.id) +
					(SELECT COUNT(*) FROM workout_wods ww WHERE ww.workout_id = w.id) +
					(SELECT COUNT(*) FROM user_workouts uw WHERE uw.workout_id = w.id) as fk_count
				FROM workouts w
				WHERE LOWER(w.name) = ? AND w.created_by = ?
				ORDER BY w.created_at ASC`
		}
		args = []interface{}{keyName, createdBy.Int64}
	} else {
		if s.driver == "postgres" {
			query = `
				SELECT w.id, w.name, w.created_at,
					(SELECT COUNT(*) FROM workout_movements wm WHERE wm.workout_id = w.id) +
					(SELECT COUNT(*) FROM workout_wods ww WHERE ww.workout_id = w.id) +
					(SELECT COUNT(*) FROM user_workouts uw WHERE uw.workout_id = w.id) as fk_count
				FROM workouts w
				WHERE LOWER(w.name) = $1 AND w.created_by IS NULL
				ORDER BY w.created_at ASC`
		} else {
			query = `
				SELECT w.id, w.name, w.created_at,
					(SELECT COUNT(*) FROM workout_movements wm WHERE wm.workout_id = w.id) +
					(SELECT COUNT(*) FROM workout_wods ww WHERE ww.workout_id = w.id) +
					(SELECT COUNT(*) FROM user_workouts uw WHERE uw.workout_id = w.id) as fk_count
				FROM workouts w
				WHERE LOWER(w.name) = ? AND w.created_by IS NULL
				ORDER BY w.created_at ASC`
		}
		args = []interface{}{keyName}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query workout records: %w", err)
	}
	defer rows.Close()

	var records []*DuplicateRecord
	for rows.Next() {
		var id int64
		var name string
		var createdAt time.Time
		var fkCount int

		if err := rows.Scan(&id, &name, &createdAt, &fkCount); err != nil {
			return nil, fmt.Errorf("failed to scan workout record: %w", err)
		}

		records = append(records, &DuplicateRecord{
			ID:        id,
			Name:      name,
			CreatedAt: createdAt,
			FKCount:   fkCount,
		})
	}

	return records, rows.Err()
}

// PreviewMerge shows what will happen when merging duplicates
func (s *DataQualityService) PreviewMerge(entityType EntityType, keepID int64, deleteIDs []int64) (*MergePreview, error) {
	preview := &MergePreview{
		EntityType:  entityType,
		KeepID:      keepID,
		DeleteIDs:   deleteIDs,
		FKUpdates:   make(map[string]int),
		AffectedIDs: make(map[string][]int64),
		CanProceed:  true,
	}

	// Validate that keepID and deleteIDs don't overlap
	for _, id := range deleteIDs {
		if id == keepID {
			return nil, fmt.Errorf("keepID (%d) cannot be in deleteIDs", keepID)
		}
	}

	switch entityType {
	case EntityTypeMovement:
		return s.previewMovementMerge(preview)
	case EntityTypeWOD:
		return s.previewWODMerge(preview)
	case EntityTypeUserWorkout:
		return s.previewUserWorkoutMerge(preview)
	case EntityTypeUser:
		preview.CanProceed = false
		preview.Warnings = append(preview.Warnings, "User merging is not supported - too many dependencies. Delete users manually after reviewing.")
		return preview, nil
	case EntityTypeWorkout:
		return s.previewWorkoutMerge(preview)
	default:
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}
}

// previewMovementMerge previews merging movements
func (s *DataQualityService) previewMovementMerge(preview *MergePreview) (*MergePreview, error) {
	// Count FK references in workout_movements
	for _, id := range preview.DeleteIDs {
		var count int
		var query string
		if s.driver == "postgres" {
			query = "SELECT COUNT(*) FROM workout_movements WHERE movement_id = $1"
		} else {
			query = "SELECT COUNT(*) FROM workout_movements WHERE movement_id = ?"
		}
		if err := s.db.QueryRow(query, id).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to count workout_movements: %w", err)
		}
		preview.FKUpdates["workout_movements"] += count
	}

	// Count FK references in user_workout_movements
	for _, id := range preview.DeleteIDs {
		var count int
		var query string
		if s.driver == "postgres" {
			query = "SELECT COUNT(*) FROM user_workout_movements WHERE movement_id = $1"
		} else {
			query = "SELECT COUNT(*) FROM user_workout_movements WHERE movement_id = ?"
		}
		if err := s.db.QueryRow(query, id).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to count user_workout_movements: %w", err)
		}
		preview.FKUpdates["user_workout_movements"] += count
	}

	return preview, nil
}

// previewWODMerge previews merging WODs
func (s *DataQualityService) previewWODMerge(preview *MergePreview) (*MergePreview, error) {
	// Count FK references in workout_wods
	for _, id := range preview.DeleteIDs {
		var count int
		var query string
		if s.driver == "postgres" {
			query = "SELECT COUNT(*) FROM workout_wods WHERE wod_id = $1"
		} else {
			query = "SELECT COUNT(*) FROM workout_wods WHERE wod_id = ?"
		}
		if err := s.db.QueryRow(query, id).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to count workout_wods: %w", err)
		}
		preview.FKUpdates["workout_wods"] += count
	}

	// Count FK references in user_workout_wods
	for _, id := range preview.DeleteIDs {
		var count int
		var query string
		if s.driver == "postgres" {
			query = "SELECT COUNT(*) FROM user_workout_wods WHERE wod_id = $1"
		} else {
			query = "SELECT COUNT(*) FROM user_workout_wods WHERE wod_id = ?"
		}
		if err := s.db.QueryRow(query, id).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to count user_workout_wods: %w", err)
		}
		preview.FKUpdates["user_workout_wods"] += count
	}

	return preview, nil
}

// previewUserWorkoutMerge previews merging user workouts
func (s *DataQualityService) previewUserWorkoutMerge(preview *MergePreview) (*MergePreview, error) {
	// Count FK references in user_workout_movements
	for _, id := range preview.DeleteIDs {
		var count int
		var query string
		if s.driver == "postgres" {
			query = "SELECT COUNT(*) FROM user_workout_movements WHERE user_workout_id = $1"
		} else {
			query = "SELECT COUNT(*) FROM user_workout_movements WHERE user_workout_id = ?"
		}
		if err := s.db.QueryRow(query, id).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to count user_workout_movements: %w", err)
		}
		preview.FKUpdates["user_workout_movements"] += count
	}

	// Count FK references in user_workout_wods
	for _, id := range preview.DeleteIDs {
		var count int
		var query string
		if s.driver == "postgres" {
			query = "SELECT COUNT(*) FROM user_workout_wods WHERE user_workout_id = $1"
		} else {
			query = "SELECT COUNT(*) FROM user_workout_wods WHERE user_workout_id = ?"
		}
		if err := s.db.QueryRow(query, id).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to count user_workout_wods: %w", err)
		}
		preview.FKUpdates["user_workout_wods"] += count
	}

	// For user workouts, we delete the child records rather than update FKs
	preview.Warnings = append(preview.Warnings, "Merging user workouts will DELETE child movement/WOD records from duplicates")

	return preview, nil
}

// previewWorkoutMerge previews merging workout templates
func (s *DataQualityService) previewWorkoutMerge(preview *MergePreview) (*MergePreview, error) {
	// Count FK references in workout_movements
	for _, id := range preview.DeleteIDs {
		var count int
		var query string
		if s.driver == "postgres" {
			query = "SELECT COUNT(*) FROM workout_movements WHERE workout_id = $1"
		} else {
			query = "SELECT COUNT(*) FROM workout_movements WHERE workout_id = ?"
		}
		if err := s.db.QueryRow(query, id).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to count workout_movements: %w", err)
		}
		preview.FKUpdates["workout_movements"] += count
	}

	// Count FK references in workout_wods
	for _, id := range preview.DeleteIDs {
		var count int
		var query string
		if s.driver == "postgres" {
			query = "SELECT COUNT(*) FROM workout_wods WHERE workout_id = $1"
		} else {
			query = "SELECT COUNT(*) FROM workout_wods WHERE workout_id = ?"
		}
		if err := s.db.QueryRow(query, id).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to count workout_wods: %w", err)
		}
		preview.FKUpdates["workout_wods"] += count
	}

	// Count FK references in user_workouts
	for _, id := range preview.DeleteIDs {
		var count int
		var query string
		if s.driver == "postgres" {
			query = "SELECT COUNT(*) FROM user_workouts WHERE workout_id = $1"
		} else {
			query = "SELECT COUNT(*) FROM user_workouts WHERE workout_id = ?"
		}
		if err := s.db.QueryRow(query, id).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to count user_workouts: %w", err)
		}
		preview.FKUpdates["user_workouts"] += count
	}

	return preview, nil
}

// ExecuteMerge performs the actual merge operation
func (s *DataQualityService) ExecuteMerge(entityType EntityType, keepID int64, deleteIDs []int64, adminUserID int64) (*MergeResult, error) {
	result := &MergeResult{
		EntityType: entityType,
		KeepID:     keepID,
		DeletedIDs: deleteIDs,
		FKUpdates:  make(map[string]int),
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		result.Error = fmt.Sprintf("failed to begin transaction: %v", err)
		return result, err
	}
	defer tx.Rollback()

	switch entityType {
	case EntityTypeMovement:
		err = s.executeMovementMerge(tx, result, keepID, deleteIDs)
	case EntityTypeWOD:
		err = s.executeWODMerge(tx, result, keepID, deleteIDs)
	case EntityTypeUserWorkout:
		err = s.executeUserWorkoutMerge(tx, result, keepID, deleteIDs)
	case EntityTypeWorkout:
		err = s.executeWorkoutMerge(tx, result, keepID, deleteIDs)
	default:
		return nil, fmt.Errorf("unsupported entity type for merge: %s", entityType)
	}

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		result.Error = fmt.Sprintf("failed to commit transaction: %v", err)
		return result, err
	}

	result.Success = true
	result.DeletedCount = len(deleteIDs)

	// Log audit entry
	if s.auditLogService != nil {
		auditDetails := map[string]interface{}{
			"entity":      entityType,
			"keep_id":     keepID,
			"deleted_ids": deleteIDs,
			"fk_updates":  result.FKUpdates,
		}
		s.auditLogService.LogEvent("duplicate_merge", &adminUserID, nil, nil, nil, auditDetails)
	}

	return result, nil
}

// executeMovementMerge performs movement merge within a transaction
func (s *DataQualityService) executeMovementMerge(tx *sql.Tx, result *MergeResult, keepID int64, deleteIDs []int64) error {
	deleteIDsStr := s.makeInClause(deleteIDs)

	// Update workout_movements
	var query string
	if s.driver == "postgres" {
		query = fmt.Sprintf("UPDATE workout_movements SET movement_id = $1 WHERE movement_id IN (%s)", deleteIDsStr)
	} else {
		query = fmt.Sprintf("UPDATE workout_movements SET movement_id = ? WHERE movement_id IN (%s)", deleteIDsStr)
	}
	res, err := tx.Exec(query, keepID)
	if err != nil {
		return fmt.Errorf("failed to update workout_movements: %w", err)
	}
	affected, _ := res.RowsAffected()
	result.FKUpdates["workout_movements"] = int(affected)

	// Update user_workout_movements
	if s.driver == "postgres" {
		query = fmt.Sprintf("UPDATE user_workout_movements SET movement_id = $1 WHERE movement_id IN (%s)", deleteIDsStr)
	} else {
		query = fmt.Sprintf("UPDATE user_workout_movements SET movement_id = ? WHERE movement_id IN (%s)", deleteIDsStr)
	}
	res, err = tx.Exec(query, keepID)
	if err != nil {
		return fmt.Errorf("failed to update user_workout_movements: %w", err)
	}
	affected, _ = res.RowsAffected()
	result.FKUpdates["user_workout_movements"] = int(affected)

	// Delete duplicates
	query = fmt.Sprintf("DELETE FROM movements WHERE id IN (%s)", deleteIDsStr)
	_, err = tx.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete movements: %w", err)
	}

	return nil
}

// executeWODMerge performs WOD merge within a transaction
func (s *DataQualityService) executeWODMerge(tx *sql.Tx, result *MergeResult, keepID int64, deleteIDs []int64) error {
	deleteIDsStr := s.makeInClause(deleteIDs)

	// Update workout_wods
	var query string
	if s.driver == "postgres" {
		query = fmt.Sprintf("UPDATE workout_wods SET wod_id = $1 WHERE wod_id IN (%s)", deleteIDsStr)
	} else {
		query = fmt.Sprintf("UPDATE workout_wods SET wod_id = ? WHERE wod_id IN (%s)", deleteIDsStr)
	}
	res, err := tx.Exec(query, keepID)
	if err != nil {
		return fmt.Errorf("failed to update workout_wods: %w", err)
	}
	affected, _ := res.RowsAffected()
	result.FKUpdates["workout_wods"] = int(affected)

	// Update user_workout_wods
	if s.driver == "postgres" {
		query = fmt.Sprintf("UPDATE user_workout_wods SET wod_id = $1 WHERE wod_id IN (%s)", deleteIDsStr)
	} else {
		query = fmt.Sprintf("UPDATE user_workout_wods SET wod_id = ? WHERE wod_id IN (%s)", deleteIDsStr)
	}
	res, err = tx.Exec(query, keepID)
	if err != nil {
		return fmt.Errorf("failed to update user_workout_wods: %w", err)
	}
	affected, _ = res.RowsAffected()
	result.FKUpdates["user_workout_wods"] = int(affected)

	// Delete duplicates
	query = fmt.Sprintf("DELETE FROM wods WHERE id IN (%s)", deleteIDsStr)
	_, err = tx.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete wods: %w", err)
	}

	return nil
}

// executeUserWorkoutMerge performs user workout merge within a transaction
func (s *DataQualityService) executeUserWorkoutMerge(tx *sql.Tx, result *MergeResult, keepID int64, deleteIDs []int64) error {
	deleteIDsStr := s.makeInClause(deleteIDs)

	// For user workouts, we delete child records (cascade would handle this, but let's be explicit)
	// Delete user_workout_movements for duplicates
	query := fmt.Sprintf("DELETE FROM user_workout_movements WHERE user_workout_id IN (%s)", deleteIDsStr)
	res, err := tx.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete user_workout_movements: %w", err)
	}
	affected, _ := res.RowsAffected()
	result.FKUpdates["user_workout_movements_deleted"] = int(affected)

	// Delete user_workout_wods for duplicates
	query = fmt.Sprintf("DELETE FROM user_workout_wods WHERE user_workout_id IN (%s)", deleteIDsStr)
	res, err = tx.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete user_workout_wods: %w", err)
	}
	affected, _ = res.RowsAffected()
	result.FKUpdates["user_workout_wods_deleted"] = int(affected)

	// Delete duplicate user_workouts
	query = fmt.Sprintf("DELETE FROM user_workouts WHERE id IN (%s)", deleteIDsStr)
	_, err = tx.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete user_workouts: %w", err)
	}

	return nil
}

// executeWorkoutMerge performs workout template merge within a transaction
func (s *DataQualityService) executeWorkoutMerge(tx *sql.Tx, result *MergeResult, keepID int64, deleteIDs []int64) error {
	deleteIDsStr := s.makeInClause(deleteIDs)

	// Update user_workouts to point to kept workout
	var query string
	if s.driver == "postgres" {
		query = fmt.Sprintf("UPDATE user_workouts SET workout_id = $1 WHERE workout_id IN (%s)", deleteIDsStr)
	} else {
		query = fmt.Sprintf("UPDATE user_workouts SET workout_id = ? WHERE workout_id IN (%s)", deleteIDsStr)
	}
	res, err := tx.Exec(query, keepID)
	if err != nil {
		return fmt.Errorf("failed to update user_workouts: %w", err)
	}
	affected, _ := res.RowsAffected()
	result.FKUpdates["user_workouts"] = int(affected)

	// Delete workout_movements for duplicates (CASCADE)
	query = fmt.Sprintf("DELETE FROM workout_movements WHERE workout_id IN (%s)", deleteIDsStr)
	res, err = tx.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete workout_movements: %w", err)
	}
	affected, _ = res.RowsAffected()
	result.FKUpdates["workout_movements_deleted"] = int(affected)

	// Delete workout_wods for duplicates (CASCADE)
	query = fmt.Sprintf("DELETE FROM workout_wods WHERE workout_id IN (%s)", deleteIDsStr)
	res, err = tx.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete workout_wods: %w", err)
	}
	affected, _ = res.RowsAffected()
	result.FKUpdates["workout_wods_deleted"] = int(affected)

	// Delete duplicate workouts
	query = fmt.Sprintf("DELETE FROM workouts WHERE id IN (%s)", deleteIDsStr)
	_, err = tx.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete workouts: %w", err)
	}

	return nil
}

// makeInClause creates a comma-separated list of IDs for IN clause
func (s *DataQualityService) makeInClause(ids []int64) string {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(strs, ",")
}

// ScanDataQuality performs a comprehensive data quality scan
func (s *DataQualityService) ScanDataQuality() (*DataQualityScanResult, error) {
	start := time.Now()

	result := &DataQualityScanResult{
		Issues:     []*DataQualityIssue{},
		ByType:     make(map[string]int),
		BySeverity: make(map[string]int),
	}

	// Check for orphaned records
	orphanIssues, err := s.checkOrphanedRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to check orphaned records: %w", err)
	}
	result.Issues = append(result.Issues, orphanIssues...)

	// Check for empty required fields
	emptyFieldIssues, err := s.checkEmptyRequiredFields()
	if err != nil {
		return nil, fmt.Errorf("failed to check empty fields: %w", err)
	}
	result.Issues = append(result.Issues, emptyFieldIssues...)

	// Check for future dates
	futureDateIssues, err := s.checkFutureDates()
	if err != nil {
		return nil, fmt.Errorf("failed to check future dates: %w", err)
	}
	result.Issues = append(result.Issues, futureDateIssues...)

	// Check for invalid email formats
	invalidEmailIssues, err := s.checkInvalidEmails()
	if err != nil {
		return nil, fmt.Errorf("failed to check invalid emails: %w", err)
	}
	result.Issues = append(result.Issues, invalidEmailIssues...)

	// Aggregate counts
	for _, issue := range result.Issues {
		result.ByType[issue.IssueType]++
		result.BySeverity[issue.Severity]++
	}
	result.TotalCount = len(result.Issues)
	result.ScanTime = time.Since(start).String()

	return result, nil
}

// checkOrphanedRecords checks for FK references to non-existent records
func (s *DataQualityService) checkOrphanedRecords() ([]*DataQualityIssue, error) {
	var issues []*DataQualityIssue

	// Check workout_movements referencing non-existent movements
	query := `
		SELECT wm.id, wm.movement_id
		FROM workout_movements wm
		LEFT JOIN movements m ON wm.movement_id = m.id
		WHERE m.id IS NULL`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, movementID int64
		if err := rows.Scan(&id, &movementID); err != nil {
			return nil, err
		}
		issues = append(issues, &DataQualityIssue{
			IssueType:   "orphaned_fk",
			EntityType:  "workout_movements",
			RecordID:    id,
			Field:       "movement_id",
			Description: fmt.Sprintf("References non-existent movement ID %d", movementID),
			Severity:    "error",
		})
	}

	// Check user_workout_movements referencing non-existent movements
	query = `
		SELECT uwm.id, uwm.movement_id
		FROM user_workout_movements uwm
		LEFT JOIN movements m ON uwm.movement_id = m.id
		WHERE m.id IS NULL`
	rows, err = s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, movementID int64
		if err := rows.Scan(&id, &movementID); err != nil {
			return nil, err
		}
		issues = append(issues, &DataQualityIssue{
			IssueType:   "orphaned_fk",
			EntityType:  "user_workout_movements",
			RecordID:    id,
			Field:       "movement_id",
			Description: fmt.Sprintf("References non-existent movement ID %d", movementID),
			Severity:    "error",
		})
	}

	// Check workout_wods referencing non-existent WODs
	query = `
		SELECT ww.id, ww.wod_id
		FROM workout_wods ww
		LEFT JOIN wods w ON ww.wod_id = w.id
		WHERE w.id IS NULL`
	rows, err = s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, wodID int64
		if err := rows.Scan(&id, &wodID); err != nil {
			return nil, err
		}
		issues = append(issues, &DataQualityIssue{
			IssueType:   "orphaned_fk",
			EntityType:  "workout_wods",
			RecordID:    id,
			Field:       "wod_id",
			Description: fmt.Sprintf("References non-existent WOD ID %d", wodID),
			Severity:    "error",
		})
	}

	// Check user_workout_wods referencing non-existent WODs
	query = `
		SELECT uww.id, uww.wod_id
		FROM user_workout_wods uww
		LEFT JOIN wods w ON uww.wod_id = w.id
		WHERE w.id IS NULL`
	rows, err = s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, wodID int64
		if err := rows.Scan(&id, &wodID); err != nil {
			return nil, err
		}
		issues = append(issues, &DataQualityIssue{
			IssueType:   "orphaned_fk",
			EntityType:  "user_workout_wods",
			RecordID:    id,
			Field:       "wod_id",
			Description: fmt.Sprintf("References non-existent WOD ID %d", wodID),
			Severity:    "error",
		})
	}

	// Check user_workouts referencing non-existent users
	query = `
		SELECT uw.id, uw.user_id
		FROM user_workouts uw
		LEFT JOIN users u ON uw.user_id = u.id
		WHERE u.id IS NULL`
	rows, err = s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, userID int64
		if err := rows.Scan(&id, &userID); err != nil {
			return nil, err
		}
		issues = append(issues, &DataQualityIssue{
			IssueType:   "orphaned_fk",
			EntityType:  "user_workouts",
			RecordID:    id,
			Field:       "user_id",
			Description: fmt.Sprintf("References non-existent user ID %d", userID),
			Severity:    "error",
		})
	}

	return issues, nil
}

// checkEmptyRequiredFields checks for empty required fields
func (s *DataQualityService) checkEmptyRequiredFields() ([]*DataQualityIssue, error) {
	var issues []*DataQualityIssue

	// Check users with empty names
	query := "SELECT id, email FROM users WHERE name IS NULL OR TRIM(name) = ''"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var email string
		if err := rows.Scan(&id, &email); err != nil {
			return nil, err
		}
		issues = append(issues, &DataQualityIssue{
			IssueType:   "empty_required_field",
			EntityType:  EntityTypeUser,
			RecordID:    id,
			Field:       "name",
			Description: fmt.Sprintf("User %s has empty name", email),
			Severity:    "warning",
		})
	}

	// Check movements with empty names
	query = "SELECT id FROM movements WHERE name IS NULL OR TRIM(name) = ''"
	rows, err = s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		issues = append(issues, &DataQualityIssue{
			IssueType:   "empty_required_field",
			EntityType:  EntityTypeMovement,
			RecordID:    id,
			Field:       "name",
			Description: "Movement has empty name",
			Severity:    "error",
		})
	}

	// Check WODs with empty names
	query = "SELECT id FROM wods WHERE name IS NULL OR TRIM(name) = ''"
	rows, err = s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		issues = append(issues, &DataQualityIssue{
			IssueType:   "empty_required_field",
			EntityType:  EntityTypeWOD,
			RecordID:    id,
			Field:       "name",
			Description: "WOD has empty name",
			Severity:    "error",
		})
	}

	return issues, nil
}

// checkFutureDates checks for records with dates in the future
func (s *DataQualityService) checkFutureDates() ([]*DataQualityIssue, error) {
	var issues []*DataQualityIssue

	// Check user_workouts with future workout dates
	var query string
	if s.driver == "postgres" {
		query = "SELECT id, workout_date, user_id FROM user_workouts WHERE workout_date > CURRENT_DATE"
	} else if s.driver == "mysql" {
		query = "SELECT id, workout_date, user_id FROM user_workouts WHERE workout_date > CURDATE()"
	} else {
		query = "SELECT id, workout_date, user_id FROM user_workouts WHERE workout_date > date('now')"
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, userID int64
		var workoutDate time.Time
		if err := rows.Scan(&id, &workoutDate, &userID); err != nil {
			return nil, err
		}
		issues = append(issues, &DataQualityIssue{
			IssueType:   "future_date",
			EntityType:  EntityTypeUserWorkout,
			RecordID:    id,
			Field:       "workout_date",
			Description: fmt.Sprintf("Workout date %s is in the future", workoutDate.Format("2006-01-02")),
			Severity:    "warning",
			Details: map[string]interface{}{
				"user_id":      userID,
				"workout_date": workoutDate.Format("2006-01-02"),
			},
		})
	}

	return issues, nil
}

// checkInvalidEmails checks for users with invalid email formats
func (s *DataQualityService) checkInvalidEmails() ([]*DataQualityIssue, error) {
	var issues []*DataQualityIssue

	// Simple check: emails without @ symbol
	query := "SELECT id, email FROM users WHERE email NOT LIKE '%@%.%'"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var email string
		if err := rows.Scan(&id, &email); err != nil {
			return nil, err
		}
		issues = append(issues, &DataQualityIssue{
			IssueType:   "invalid_email",
			EntityType:  EntityTypeUser,
			RecordID:    id,
			Field:       "email",
			Description: fmt.Sprintf("Invalid email format: %s", email),
			Severity:    "warning",
		})
	}

	return issues, nil
}
