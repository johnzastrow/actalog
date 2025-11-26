package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// DataChangeLogRepository implements domain.DataChangeLogRepository
type DataChangeLogRepository struct {
	db     *sql.DB
	driver string
}

// NewDataChangeLogRepository creates a new data change log repository
func NewDataChangeLogRepository(db *sql.DB, driver string) *DataChangeLogRepository {
	return &DataChangeLogRepository{
		db:     db,
		driver: driver,
	}
}

// Create creates a new data change log entry
func (r *DataChangeLogRepository) Create(log *domain.DataChangeLog) error {
	now := time.Now()
	log.CreatedAt = now

	var query string

	switch r.driver {
	case "sqlite3", "mysql":
		query = `INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, before_values, after_values, ip_address, user_agent, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		result, err := r.db.Exec(query, log.EntityType, log.EntityID, log.EntityName, log.Operation, log.UserID, log.UserEmail, log.BeforeValues, log.AfterValues, log.IPAddress, log.UserAgent, log.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to create data change log: %w", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get data change log ID: %w", err)
		}
		log.ID = id

	case "postgres":
		query = `INSERT INTO data_change_logs (entity_type, entity_id, entity_name, operation, user_id, user_email, before_values, after_values, ip_address, user_agent, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`
		err := r.db.QueryRow(query, log.EntityType, log.EntityID, log.EntityName, log.Operation, log.UserID, log.UserEmail, log.BeforeValues, log.AfterValues, log.IPAddress, log.UserAgent, log.CreatedAt).Scan(&log.ID)
		if err != nil {
			return fmt.Errorf("failed to create data change log: %w", err)
		}

	default:
		return fmt.Errorf("unsupported database driver: %s", r.driver)
	}

	return nil
}

// GetByID retrieves a single data change log by ID
func (r *DataChangeLogRepository) GetByID(id int64) (*domain.DataChangeLog, error) {
	query := `
		SELECT id, entity_type, entity_id, entity_name, operation, user_id, user_email,
			before_values, after_values, ip_address, user_agent, created_at
		FROM data_change_logs
		WHERE id = ?`

	if r.driver == "postgres" {
		query = `
		SELECT id, entity_type, entity_id, entity_name, operation, user_id, user_email,
			before_values, after_values, ip_address, user_agent, created_at
		FROM data_change_logs
		WHERE id = $1`
	}

	log := &domain.DataChangeLog{}
	err := r.db.QueryRow(query, id).Scan(
		&log.ID, &log.EntityType, &log.EntityID, &log.EntityName, &log.Operation,
		&log.UserID, &log.UserEmail, &log.BeforeValues, &log.AfterValues,
		&log.IPAddress, &log.UserAgent, &log.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("data change log not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get data change log: %w", err)
	}

	return log, nil
}

// List retrieves data change logs with pagination and optional filters
func (r *DataChangeLogRepository) List(filters domain.DataChangeLogFilters, limit, offset int) ([]*domain.DataChangeLog, error) {
	query := `
		SELECT id, entity_type, entity_id, entity_name, operation, user_id, user_email,
			before_values, after_values, ip_address, user_agent, created_at
		FROM data_change_logs
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters.EntityType != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND entity_type = $%d", argIndex)
		} else {
			query += " AND entity_type = ?"
		}
		args = append(args, *filters.EntityType)
		argIndex++
	}

	if filters.EntityID != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND entity_id = $%d", argIndex)
		} else {
			query += " AND entity_id = ?"
		}
		args = append(args, *filters.EntityID)
		argIndex++
	}

	if filters.Operation != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND operation = $%d", argIndex)
		} else {
			query += " AND operation = ?"
		}
		args = append(args, *filters.Operation)
		argIndex++
	}

	if filters.UserID != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		} else {
			query += " AND user_id = ?"
		}
		args = append(args, *filters.UserID)
		argIndex++
	}

	if filters.UserEmail != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND user_email LIKE $%d", argIndex)
		} else {
			query += " AND user_email LIKE ?"
		}
		args = append(args, "%"+*filters.UserEmail+"%")
		argIndex++
	}

	if filters.StartDate != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		} else {
			query += " AND created_at >= ?"
		}
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		} else {
			query += " AND created_at <= ?"
		}
		args = append(args, *filters.EndDate)
		argIndex++
	}

	// Order by created_at descending (most recent first)
	query += " ORDER BY created_at DESC"

	// Add pagination
	if r.driver == "postgres" {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	} else {
		query += " LIMIT ? OFFSET ?"
	}
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list data change logs: %w", err)
	}
	defer rows.Close()

	logs := []*domain.DataChangeLog{}
	for rows.Next() {
		log := &domain.DataChangeLog{}
		err := rows.Scan(
			&log.ID, &log.EntityType, &log.EntityID, &log.EntityName, &log.Operation,
			&log.UserID, &log.UserEmail, &log.BeforeValues, &log.AfterValues,
			&log.IPAddress, &log.UserAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan data change log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// Count returns the total number of data change logs matching the filters
func (r *DataChangeLogRepository) Count(filters domain.DataChangeLogFilters) (int, error) {
	query := "SELECT COUNT(*) FROM data_change_logs WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	// Apply same filters as List
	if filters.EntityType != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND entity_type = $%d", argIndex)
		} else {
			query += " AND entity_type = ?"
		}
		args = append(args, *filters.EntityType)
		argIndex++
	}

	if filters.EntityID != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND entity_id = $%d", argIndex)
		} else {
			query += " AND entity_id = ?"
		}
		args = append(args, *filters.EntityID)
		argIndex++
	}

	if filters.Operation != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND operation = $%d", argIndex)
		} else {
			query += " AND operation = ?"
		}
		args = append(args, *filters.Operation)
		argIndex++
	}

	if filters.UserID != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		} else {
			query += " AND user_id = ?"
		}
		args = append(args, *filters.UserID)
		argIndex++
	}

	if filters.UserEmail != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND user_email LIKE $%d", argIndex)
		} else {
			query += " AND user_email LIKE ?"
		}
		args = append(args, "%"+*filters.UserEmail+"%")
		argIndex++
	}

	if filters.StartDate != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		} else {
			query += " AND created_at >= ?"
		}
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		} else {
			query += " AND created_at <= ?"
		}
		args = append(args, *filters.EndDate)
		argIndex++
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count data change logs: %w", err)
	}

	return count, nil
}

// GetByEntityID retrieves all changes for a specific entity
func (r *DataChangeLogRepository) GetByEntityID(entityType string, entityID int64, limit, offset int) ([]*domain.DataChangeLog, error) {
	filters := domain.DataChangeLogFilters{
		EntityType: &entityType,
		EntityID:   &entityID,
	}
	return r.List(filters, limit, offset)
}

// GetByUserID retrieves all changes made by a specific user
func (r *DataChangeLogRepository) GetByUserID(userID int64, limit, offset int) ([]*domain.DataChangeLog, error) {
	filters := domain.DataChangeLogFilters{
		UserID: &userID,
	}
	return r.List(filters, limit, offset)
}

// DeleteOlderThan deletes logs older than the specified time (for cleanup)
func (r *DataChangeLogRepository) DeleteOlderThan(before time.Time) (int, error) {
	var query string
	if r.driver == "postgres" {
		query = "DELETE FROM data_change_logs WHERE created_at < $1"
	} else {
		query = "DELETE FROM data_change_logs WHERE created_at < ?"
	}

	result, err := r.db.Exec(query, before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old data change logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}
