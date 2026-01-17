package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// EmailLogRepository implements domain.EmailLogRepository
type EmailLogRepository struct {
	db     *sql.DB
	driver string
}

// NewEmailLogRepository creates a new email log repository
func NewEmailLogRepository(db *sql.DB, driver string) *EmailLogRepository {
	return &EmailLogRepository{
		db:     db,
		driver: driver,
	}
}

// Create creates a new email log entry
func (r *EmailLogRepository) Create(log *domain.EmailLog) error {
	now := time.Now()
	log.CreatedAt = now

	var query string
	var err error

	switch r.driver {
	case "sqlite3", "mysql":
		query = `INSERT INTO email_logs (recipient_email, email_type, subject, success, error_message, debug_info, sent_by_user_id, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		result, execErr := r.db.Exec(query, log.RecipientEmail, log.EmailType, log.Subject, log.Success, log.ErrorMessage, log.DebugInfo, log.SentByUserID, log.CreatedAt)
		if execErr != nil {
			return fmt.Errorf("failed to create email log: %w", execErr)
		}
		id, idErr := result.LastInsertId()
		if idErr != nil {
			return fmt.Errorf("failed to get email log ID: %w", idErr)
		}
		log.ID = id

	case "postgres":
		query = `INSERT INTO email_logs (recipient_email, email_type, subject, success, error_message, debug_info, sent_by_user_id, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
		err = r.db.QueryRow(query, log.RecipientEmail, log.EmailType, log.Subject, log.Success, log.ErrorMessage, log.DebugInfo, log.SentByUserID, log.CreatedAt).Scan(&log.ID)
		if err != nil {
			return fmt.Errorf("failed to create email log: %w", err)
		}

	default:
		return fmt.Errorf("unsupported database driver: %s", r.driver)
	}

	return nil
}

// GetByID retrieves a single email log by ID
func (r *EmailLogRepository) GetByID(id int64) (*domain.EmailLog, error) {
	query := `
		SELECT
			el.id, el.recipient_email, el.email_type, el.subject,
			el.success, el.error_message, el.debug_info, el.sent_by_user_id, el.created_at,
			u.email as sent_by_user_email
		FROM email_logs el
		LEFT JOIN users u ON el.sent_by_user_id = u.id
		WHERE el.id = ?`

	if r.driver == "postgres" {
		query = strings.Replace(query, "?", "$1", 1)
	}

	log := &domain.EmailLog{}
	err := r.db.QueryRow(query, id).Scan(
		&log.ID, &log.RecipientEmail, &log.EmailType, &log.Subject,
		&log.Success, &log.ErrorMessage, &log.DebugInfo, &log.SentByUserID, &log.CreatedAt,
		&log.SentByUserEmail,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("email log not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get email log: %w", err)
	}

	return log, nil
}

// List retrieves email logs with pagination and optional filters
func (r *EmailLogRepository) List(filters domain.EmailLogFilters, limit, offset int) ([]*domain.EmailLog, error) {
	query := `
		SELECT
			el.id, el.recipient_email, el.email_type, el.subject,
			el.success, el.error_message, el.debug_info, el.sent_by_user_id, el.created_at,
			u.email as sent_by_user_email
		FROM email_logs el
		LEFT JOIN users u ON el.sent_by_user_id = u.id
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters.EmailType != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND el.email_type = $%d", argIndex)
		} else {
			query += " AND el.email_type = ?"
		}
		args = append(args, *filters.EmailType)
		argIndex++
	}

	if filters.RecipientEmail != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND el.recipient_email ILIKE $%d", argIndex)
		} else {
			query += " AND el.recipient_email LIKE ?"
		}
		args = append(args, "%"+*filters.RecipientEmail+"%")
		argIndex++
	}

	if filters.Success != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND el.success = $%d", argIndex)
		} else {
			query += " AND el.success = ?"
		}
		args = append(args, *filters.Success)
		argIndex++
	}

	if filters.SentByUserID != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND el.sent_by_user_id = $%d", argIndex)
		} else {
			query += " AND el.sent_by_user_id = ?"
		}
		args = append(args, *filters.SentByUserID)
		argIndex++
	}

	if filters.StartDate != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND el.created_at >= $%d", argIndex)
		} else {
			query += " AND el.created_at >= ?"
		}
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND el.created_at <= $%d", argIndex)
		} else {
			query += " AND el.created_at <= ?"
		}
		args = append(args, *filters.EndDate)
		argIndex++
	}

	// Order by created_at descending (most recent first)
	query += " ORDER BY el.created_at DESC"

	// Add pagination
	if r.driver == "postgres" {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	} else {
		query += " LIMIT ? OFFSET ?"
	}
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list email logs: %w", err)
	}
	defer rows.Close()

	logs := []*domain.EmailLog{}
	for rows.Next() {
		log := &domain.EmailLog{}
		err := rows.Scan(
			&log.ID, &log.RecipientEmail, &log.EmailType, &log.Subject,
			&log.Success, &log.ErrorMessage, &log.DebugInfo, &log.SentByUserID, &log.CreatedAt,
			&log.SentByUserEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// Count returns the total number of email logs matching the filters
func (r *EmailLogRepository) Count(filters domain.EmailLogFilters) (int, error) {
	query := "SELECT COUNT(*) FROM email_logs WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	// Apply same filters as List
	if filters.EmailType != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND email_type = $%d", argIndex)
		} else {
			query += " AND email_type = ?"
		}
		args = append(args, *filters.EmailType)
		argIndex++
	}

	if filters.RecipientEmail != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND recipient_email ILIKE $%d", argIndex)
		} else {
			query += " AND recipient_email LIKE ?"
		}
		args = append(args, "%"+*filters.RecipientEmail+"%")
		argIndex++
	}

	if filters.Success != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND success = $%d", argIndex)
		} else {
			query += " AND success = ?"
		}
		args = append(args, *filters.Success)
		argIndex++
	}

	if filters.SentByUserID != nil {
		if r.driver == "postgres" {
			query += fmt.Sprintf(" AND sent_by_user_id = $%d", argIndex)
		} else {
			query += " AND sent_by_user_id = ?"
		}
		args = append(args, *filters.SentByUserID)
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
		return 0, fmt.Errorf("failed to count email logs: %w", err)
	}

	return count, nil
}

// DeleteOlderThan deletes email logs older than the specified time
func (r *EmailLogRepository) DeleteOlderThan(before time.Time) (int64, error) {
	var query string
	if r.driver == "postgres" {
		query = "DELETE FROM email_logs WHERE created_at < $1"
	} else {
		query = "DELETE FROM email_logs WHERE created_at < ?"
	}

	result, err := r.db.Exec(query, before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old email logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
