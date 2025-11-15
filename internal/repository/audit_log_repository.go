package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// AuditLogRepository implements domain.AuditLogRepository
type AuditLogRepository struct {
	db     *sql.DB
	driver string
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *sql.DB, driver string) *AuditLogRepository {
	return &AuditLogRepository{
		db:     db,
		driver: driver,
	}
}

// Create creates a new audit log entry
func (r *AuditLogRepository) Create(log *domain.AuditLog) error {
	now := time.Now()
	log.CreatedAt = now

	var query string
	var err error

	switch r.driver {
	case "sqlite3", "mysql":
		query = `INSERT INTO audit_logs (user_id, target_user_id, event_type, ip_address, user_agent, details, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`
		err = r.db.QueryRow(query, log.UserID, log.TargetUserID, log.EventType, log.IPAddress, log.UserAgent, log.Details, log.CreatedAt).Scan(&log.ID)
		if err != nil {
			// For SQLite/MySQL, INSERT doesn't return ID directly, need to get last insert id
			result, execErr := r.db.Exec(query, log.UserID, log.TargetUserID, log.EventType, log.IPAddress, log.UserAgent, log.Details, log.CreatedAt)
			if execErr != nil {
				return fmt.Errorf("failed to create audit log: %w", execErr)
			}
			id, idErr := result.LastInsertId()
			if idErr != nil {
				return fmt.Errorf("failed to get audit log ID: %w", idErr)
			}
			log.ID = id
		}

	case "postgres":
		query = `INSERT INTO audit_logs (user_id, target_user_id, event_type, ip_address, user_agent, details, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
		err = r.db.QueryRow(query, log.UserID, log.TargetUserID, log.EventType, log.IPAddress, log.UserAgent, log.Details, log.CreatedAt).Scan(&log.ID)
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}

	default:
		return fmt.Errorf("unsupported database driver: %s", r.driver)
	}

	return nil
}

// GetByID retrieves a single audit log by ID
func (r *AuditLogRepository) GetByID(id int64) (*domain.AuditLog, error) {
	query := `
		SELECT
			al.id, al.user_id, al.target_user_id, al.event_type,
			al.ip_address, al.user_agent, al.details, al.created_at,
			u1.email as user_email,
			u2.email as target_user_email
		FROM audit_logs al
		LEFT JOIN users u1 ON al.user_id = u1.id
		LEFT JOIN users u2 ON al.target_user_id = u2.id
		WHERE al.id = ?`

	if r.driver == "postgres" {
		query = strings.Replace(query, "?", "$1", 1)
	}

	log := &domain.AuditLog{}
	err := r.db.QueryRow(query, id).Scan(
		&log.ID, &log.UserID, &log.TargetUserID, &log.EventType,
		&log.IPAddress, &log.UserAgent, &log.Details, &log.CreatedAt,
		&log.UserEmail, &log.TargetUserEmail,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("audit log not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	return log, nil
}

// List retrieves audit logs with pagination and optional filters
func (r *AuditLogRepository) List(filters domain.AuditLogFilters, limit, offset int) ([]*domain.AuditLog, error) {
	query := `
		SELECT
			al.id, al.user_id, al.target_user_id, al.event_type,
			al.ip_address, al.user_agent, al.details, al.created_at,
			u1.email as user_email,
			u2.email as target_user_email
		FROM audit_logs al
		LEFT JOIN users u1 ON al.user_id = u1.id
		LEFT JOIN users u2 ON al.target_user_id = u2.id
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters.UserID != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND al.user_id = $%d", argIndex)
		} else {
			query += " AND al.user_id = ?"
		}
		args = append(args, *filters.UserID)
		argIndex++
	}

	if filters.TargetUserID != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND al.target_user_id = $%d", argIndex)
		} else {
			query += " AND al.target_user_id = ?"
		}
		args = append(args, *filters.TargetUserID)
		argIndex++
	}

	if filters.EventType != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND al.event_type = $%d", argIndex)
		} else {
			query += " AND al.event_type = ?"
		}
		args = append(args, *filters.EventType)
		argIndex++
	}

	if filters.IPAddress != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND al.ip_address = $%d", argIndex)
		} else {
			query += " AND al.ip_address = ?"
		}
		args = append(args, *filters.IPAddress)
		argIndex++
	}

	if filters.StartDate != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND al.created_at >= $%d", argIndex)
		} else {
			query += " AND al.created_at >= ?"
		}
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND al.created_at <= $%d", argIndex)
		} else {
			query += " AND al.created_at <= ?"
		}
		args = append(args, *filters.EndDate)
		argIndex++
	}

	// Order by created_at descending (most recent first)
	query += " ORDER BY al.created_at DESC"

	// Add pagination
	if r.driver == "postgres" {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	} else {
		query += " LIMIT ? OFFSET ?"
	}
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

	logs := []*domain.AuditLog{}
	for rows.Next() {
		log := &domain.AuditLog{}
		err := rows.Scan(
			&log.ID, &log.UserID, &log.TargetUserID, &log.EventType,
			&log.IPAddress, &log.UserAgent, &log.Details, &log.CreatedAt,
			&log.UserEmail, &log.TargetUserEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// Count returns the total number of audit logs matching the filters
func (r *AuditLogRepository) Count(filters domain.AuditLogFilters) (int, error) {
	query := "SELECT COUNT(*) FROM audit_logs WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	// Apply same filters as List
	if filters.UserID != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		} else {
			query += " AND user_id = ?"
		}
		args = append(args, *filters.UserID)
		argIndex++
	}

	if filters.TargetUserID != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND target_user_id = $%d", argIndex)
		} else {
			query += " AND target_user_id = ?"
		}
		args = append(args, *filters.TargetUserID)
		argIndex++
	}

	if filters.EventType != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND event_type = $%d", argIndex)
		} else {
			query += " AND event_type = ?"
		}
		args = append(args, *filters.EventType)
		argIndex++
	}

	if filters.IPAddress != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND ip_address = $%d", argIndex)
		} else {
			query += " AND ip_address = ?"
		}
		args = append(args, *filters.IPAddress)
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
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return count, nil
}

// GetByUserID retrieves all audit logs for a specific user (actions performed BY the user)
func (r *AuditLogRepository) GetByUserID(userID int64, limit, offset int) ([]*domain.AuditLog, error) {
	filters := domain.AuditLogFilters{
		UserID: &userID,
	}
	return r.List(filters, limit, offset)
}

// GetByTargetUserID retrieves all audit logs affecting a specific user
func (r *AuditLogRepository) GetByTargetUserID(targetUserID int64, limit, offset int) ([]*domain.AuditLog, error) {
	filters := domain.AuditLogFilters{
		TargetUserID: &targetUserID,
	}
	return r.List(filters, limit, offset)
}

// DeleteOlderThan deletes audit logs older than the specified time
func (r *AuditLogRepository) DeleteOlderThan(before time.Time) (int, error) {
	var query string
	if r.driver == "postgres" {
		query = "DELETE FROM audit_logs WHERE created_at < $1"
	} else {
		query = "DELETE FROM audit_logs WHERE created_at < ?"
	}

	result, err := r.db.Exec(query, before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}
