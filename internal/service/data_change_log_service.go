package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// DataChangeLogService handles data change logging business logic
type DataChangeLogService struct {
	repo domain.DataChangeLogRepository
}

// NewDataChangeLogService creates a new data change log service
func NewDataChangeLogService(repo domain.DataChangeLogRepository) *DataChangeLogService {
	return &DataChangeLogService{
		repo: repo,
	}
}

// LogUpdate logs an update operation with before/after values
func (s *DataChangeLogService) LogUpdate(entityType string, entityID int64, entityName string, userID int64, userEmail string, before, after interface{}, ipAddress, userAgent *string) error {
	// Serialize before value
	beforeJSON, err := json.Marshal(before)
	if err != nil {
		return fmt.Errorf("failed to marshal before values: %w", err)
	}
	beforeStr := string(beforeJSON)

	// Serialize after value
	afterJSON, err := json.Marshal(after)
	if err != nil {
		return fmt.Errorf("failed to marshal after values: %w", err)
	}
	afterStr := string(afterJSON)

	log := &domain.DataChangeLog{
		EntityType:   entityType,
		EntityID:     entityID,
		EntityName:   entityName,
		Operation:    domain.OperationUpdate,
		UserID:       userID,
		UserEmail:    userEmail,
		BeforeValues: &beforeStr,
		AfterValues:  &afterStr,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	return s.repo.Create(log)
}

// LogDelete logs a delete operation with before values (no after values)
func (s *DataChangeLogService) LogDelete(entityType string, entityID int64, entityName string, userID int64, userEmail string, before interface{}, ipAddress, userAgent *string) error {
	// Serialize before value
	beforeJSON, err := json.Marshal(before)
	if err != nil {
		return fmt.Errorf("failed to marshal before values: %w", err)
	}
	beforeStr := string(beforeJSON)

	log := &domain.DataChangeLog{
		EntityType:   entityType,
		EntityID:     entityID,
		EntityName:   entityName,
		Operation:    domain.OperationDelete,
		UserID:       userID,
		UserEmail:    userEmail,
		BeforeValues: &beforeStr,
		AfterValues:  nil, // No after values for delete
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	return s.repo.Create(log)
}

// GetByID retrieves a single data change log by ID
func (s *DataChangeLogService) GetByID(id int64) (*domain.DataChangeLog, error) {
	return s.repo.GetByID(id)
}

// List retrieves data change logs with pagination and filters
func (s *DataChangeLogService) List(filters domain.DataChangeLogFilters, limit, offset int) ([]*domain.DataChangeLog, error) {
	// Validate limit
	if limit <= 0 || limit > 100 {
		limit = 50 // Default
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(filters, limit, offset)
}

// Count returns the total number of data change logs matching the filters
func (s *DataChangeLogService) Count(filters domain.DataChangeLogFilters) (int, error) {
	return s.repo.Count(filters)
}

// GetByEntityID retrieves all changes for a specific entity
func (s *DataChangeLogService) GetByEntityID(entityType string, entityID int64, limit, offset int) ([]*domain.DataChangeLog, error) {
	// Validate limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetByEntityID(entityType, entityID, limit, offset)
}

// GetByUserID retrieves all changes made by a specific user
func (s *DataChangeLogService) GetByUserID(userID int64, limit, offset int) ([]*domain.DataChangeLog, error) {
	// Validate limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetByUserID(userID, limit, offset)
}

// CleanupOldLogs deletes data change logs older than the specified duration
func (s *DataChangeLogService) CleanupOldLogs(retentionDays int) (int, error) {
	if retentionDays <= 0 {
		return 0, fmt.Errorf("retention days must be positive")
	}

	before := time.Now().AddDate(0, 0, -retentionDays)
	return s.repo.DeleteOlderThan(before)
}

// Helper functions for common data change log scenarios

// LogWODUpdate logs a WOD update
func (s *DataChangeLogService) LogWODUpdate(wodID int64, wodName string, userID int64, userEmail string, before, after interface{}, ipAddress, userAgent *string) error {
	return s.LogUpdate(domain.EntityTypeWOD, wodID, wodName, userID, userEmail, before, after, ipAddress, userAgent)
}

// LogWODDelete logs a WOD delete
func (s *DataChangeLogService) LogWODDelete(wodID int64, wodName string, userID int64, userEmail string, before interface{}, ipAddress, userAgent *string) error {
	return s.LogDelete(domain.EntityTypeWOD, wodID, wodName, userID, userEmail, before, ipAddress, userAgent)
}

// LogMovementUpdate logs a Movement update
func (s *DataChangeLogService) LogMovementUpdate(movementID int64, movementName string, userID int64, userEmail string, before, after interface{}, ipAddress, userAgent *string) error {
	return s.LogUpdate(domain.EntityTypeMovement, movementID, movementName, userID, userEmail, before, after, ipAddress, userAgent)
}

// LogMovementDelete logs a Movement delete
func (s *DataChangeLogService) LogMovementDelete(movementID int64, movementName string, userID int64, userEmail string, before interface{}, ipAddress, userAgent *string) error {
	return s.LogDelete(domain.EntityTypeMovement, movementID, movementName, userID, userEmail, before, ipAddress, userAgent)
}

// LogWorkoutUpdate logs a Workout update
func (s *DataChangeLogService) LogWorkoutUpdate(workoutID int64, workoutName string, userID int64, userEmail string, before, after interface{}, ipAddress, userAgent *string) error {
	return s.LogUpdate(domain.EntityTypeWorkout, workoutID, workoutName, userID, userEmail, before, after, ipAddress, userAgent)
}

// LogWorkoutDelete logs a Workout delete
func (s *DataChangeLogService) LogWorkoutDelete(workoutID int64, workoutName string, userID int64, userEmail string, before interface{}, ipAddress, userAgent *string) error {
	return s.LogDelete(domain.EntityTypeWorkout, workoutID, workoutName, userID, userEmail, before, ipAddress, userAgent)
}

// LogUserWorkoutUpdate logs a UserWorkout update
func (s *DataChangeLogService) LogUserWorkoutUpdate(userWorkoutID int64, workoutName string, userID int64, userEmail string, before, after interface{}, ipAddress, userAgent *string) error {
	return s.LogUpdate(domain.EntityTypeUserWorkout, userWorkoutID, workoutName, userID, userEmail, before, after, ipAddress, userAgent)
}

// LogUserWorkoutDelete logs a UserWorkout delete
func (s *DataChangeLogService) LogUserWorkoutDelete(userWorkoutID int64, workoutName string, userID int64, userEmail string, before interface{}, ipAddress, userAgent *string) error {
	return s.LogDelete(domain.EntityTypeUserWorkout, userWorkoutID, workoutName, userID, userEmail, before, ipAddress, userAgent)
}

// LogUserWorkoutMovementUpdate logs a UserWorkoutMovement update
func (s *DataChangeLogService) LogUserWorkoutMovementUpdate(id int64, movementName string, userID int64, userEmail string, before, after interface{}, ipAddress, userAgent *string) error {
	return s.LogUpdate(domain.EntityTypeUserWorkoutMovement, id, movementName, userID, userEmail, before, after, ipAddress, userAgent)
}

// LogUserWorkoutMovementDelete logs a UserWorkoutMovement delete
func (s *DataChangeLogService) LogUserWorkoutMovementDelete(id int64, movementName string, userID int64, userEmail string, before interface{}, ipAddress, userAgent *string) error {
	return s.LogDelete(domain.EntityTypeUserWorkoutMovement, id, movementName, userID, userEmail, before, ipAddress, userAgent)
}

// LogUserWorkoutWODUpdate logs a UserWorkoutWOD update
func (s *DataChangeLogService) LogUserWorkoutWODUpdate(id int64, wodName string, userID int64, userEmail string, before, after interface{}, ipAddress, userAgent *string) error {
	return s.LogUpdate(domain.EntityTypeUserWorkoutWOD, id, wodName, userID, userEmail, before, after, ipAddress, userAgent)
}

// LogUserWorkoutWODDelete logs a UserWorkoutWOD delete
func (s *DataChangeLogService) LogUserWorkoutWODDelete(id int64, wodName string, userID int64, userEmail string, before interface{}, ipAddress, userAgent *string) error {
	return s.LogDelete(domain.EntityTypeUserWorkoutWOD, id, wodName, userID, userEmail, before, ipAddress, userAgent)
}

// LogUserUpdate logs a User update
func (s *DataChangeLogService) LogUserUpdate(targetUserID int64, targetUserEmail string, userID int64, userEmail string, before, after interface{}, ipAddress, userAgent *string) error {
	return s.LogUpdate(domain.EntityTypeUser, targetUserID, targetUserEmail, userID, userEmail, before, after, ipAddress, userAgent)
}

// LogUserDelete logs a User delete
func (s *DataChangeLogService) LogUserDelete(targetUserID int64, targetUserEmail string, userID int64, userEmail string, before interface{}, ipAddress, userAgent *string) error {
	return s.LogDelete(domain.EntityTypeUser, targetUserID, targetUserEmail, userID, userEmail, before, ipAddress, userAgent)
}
