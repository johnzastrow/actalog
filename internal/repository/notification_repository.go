package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// NotificationRepository implements domain.NotificationRepository
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification
func (r *NotificationRepository) Create(notification *domain.Notification) error {
	query := rebindQuery(`
		INSERT INTO notifications (user_id, organization_id, type, title, message, data, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		err := r.db.QueryRow(
			query,
			notification.UserID,
			notification.OrganizationID,
			notification.Type,
			notification.Title,
			notification.Message,
			notification.Data,
			notification.CreatedAt,
			notification.UpdatedAt,
		).Scan(&notification.ID)
		return err
	}

	result, err := r.db.Exec(
		query,
		notification.UserID,
		notification.OrganizationID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.Data,
		notification.CreatedAt,
		notification.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	notification.ID = id
	return nil
}

// GetByID retrieves a notification by ID
func (r *NotificationRepository) GetByID(id int64) (*domain.Notification, error) {
	query := rebindQuery(`
		SELECT id, user_id, organization_id, type, title, message, data, read_at, created_at, updated_at
		FROM notifications
		WHERE id = ?
	`)

	notification := &domain.Notification{}
	var organizationID sql.NullInt64
	var readAt sql.NullTime
	var data sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&organizationID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&data,
		&readAt,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Handle nullable fields
	if organizationID.Valid {
		notification.OrganizationID = &organizationID.Int64
	}
	if readAt.Valid {
		notification.ReadAt = &readAt.Time
	}
	if data.Valid {
		notification.Data = data.String
	}

	return notification, nil
}

// GetByUserID retrieves notifications for a user with pagination
func (r *NotificationRepository) GetByUserID(userID int64, limit, offset int) ([]*domain.Notification, error) {
	query := rebindQuery(`
		SELECT id, user_id, organization_id, type, title, message, data, read_at, created_at, updated_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`)

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		notification := &domain.Notification{}
		var organizationID sql.NullInt64
		var readAt sql.NullTime
		var data sql.NullString

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&organizationID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&data,
			&readAt,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if organizationID.Valid {
			notification.OrganizationID = &organizationID.Int64
		}
		if readAt.Valid {
			notification.ReadAt = &readAt.Time
		}
		if data.Valid {
			notification.Data = data.String
		}

		notifications = append(notifications, notification)
	}

	return notifications, rows.Err()
}

// GetUnreadByUserID retrieves unread notifications for a user with pagination
func (r *NotificationRepository) GetUnreadByUserID(userID int64, limit, offset int) ([]*domain.Notification, error) {
	query := rebindQuery(`
		SELECT id, user_id, organization_id, type, title, message, data, read_at, created_at, updated_at
		FROM notifications
		WHERE user_id = ? AND read_at IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`)

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		notification := &domain.Notification{}
		var organizationID sql.NullInt64
		var readAt sql.NullTime
		var data sql.NullString

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&organizationID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&data,
			&readAt,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if organizationID.Valid {
			notification.OrganizationID = &organizationID.Int64
		}
		if readAt.Valid {
			notification.ReadAt = &readAt.Time
		}
		if data.Valid {
			notification.Data = data.String
		}

		notifications = append(notifications, notification)
	}

	return notifications, rows.Err()
}

// CountUnreadByUserID counts unread notifications for a user
func (r *NotificationRepository) CountUnreadByUserID(userID int64) (int64, error) {
	query := rebindQuery(`
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = ? AND read_at IS NULL
	`)

	var count int64
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(id int64) error {
	query := rebindQuery(`
		UPDATE notifications
		SET read_at = ?, updated_at = ?
		WHERE id = ?
	`)

	now := time.Now()
	result, err := r.db.Exec(query, now, now, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification %d not found", id)
	}

	return nil
}

// MarkAsUnread marks a notification as unread (sets read_at to NULL)
func (r *NotificationRepository) MarkAsUnread(id int64) error {
	query := rebindQuery(`
		UPDATE notifications
		SET read_at = NULL, updated_at = ?
		WHERE id = ?
	`)

	now := time.Now()
	result, err := r.db.Exec(query, now, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification %d not found", id)
	}

	return nil
}

// MarkAllAsReadForUser marks all unread notifications as read for a user
func (r *NotificationRepository) MarkAllAsReadForUser(userID int64) error {
	query := rebindQuery(`
		UPDATE notifications
		SET read_at = ?, updated_at = ?
		WHERE user_id = ? AND read_at IS NULL
	`)

	now := time.Now()
	_, err := r.db.Exec(query, now, now, userID)
	return err
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(id int64) error {
	query := rebindQuery(`DELETE FROM notifications WHERE id = ?`)

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification %d not found", id)
	}

	return nil
}

// DeleteOlderThan deletes notifications older than the specified cutoff time
func (r *NotificationRepository) DeleteOlderThan(cutoff time.Time) (int64, error) {
	query := rebindQuery(`DELETE FROM notifications WHERE created_at < ?`)

	result, err := r.db.Exec(query, cutoff)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
