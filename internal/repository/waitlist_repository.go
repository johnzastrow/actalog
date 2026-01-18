package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// WaitlistRepository implements domain.WaitlistRepository
type WaitlistRepository struct {
	db *sql.DB
}

// NewWaitlistRepository creates a new waitlist repository
func NewWaitlistRepository(db *sql.DB) *WaitlistRepository {
	return &WaitlistRepository{db: db}
}

// Create creates a new waitlist entry
func (r *WaitlistRepository) Create(entry *domain.WaitlistEntry) error {
	now := time.Now()
	entry.CreatedAt = now
	entry.UpdatedAt = now
	if entry.JoinedAt.IsZero() {
		entry.JoinedAt = now
	}

	// Get the next position for this session
	var maxPosition int
	posQuery := rebindQuery(`SELECT COALESCE(MAX(position), 0) FROM waitlist_entries WHERE session_id = ?`)
	err := r.db.QueryRow(posQuery, entry.SessionID).Scan(&maxPosition)
	if err != nil {
		return fmt.Errorf("failed to get max position: %w", err)
	}
	entry.Position = maxPosition + 1

	query := rebindQuery(`INSERT INTO waitlist_entries (session_id, user_id, position, status, joined_at, promoted_at, expired_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			entry.SessionID,
			entry.UserID,
			entry.Position,
			entry.Status,
			entry.JoinedAt,
			entry.PromotedAt,
			entry.ExpiredAt,
			entry.CreatedAt,
			entry.UpdatedAt,
		).Scan(&entry.ID)
	}

	result, err := r.db.Exec(query,
		entry.SessionID,
		entry.UserID,
		entry.Position,
		entry.Status,
		entry.JoinedAt,
		entry.PromotedAt,
		entry.ExpiredAt,
		entry.CreatedAt,
		entry.UpdatedAt,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	entry.ID = id
	return nil
}

// GetByID retrieves a waitlist entry by ID
func (r *WaitlistRepository) GetByID(id int64) (*domain.WaitlistEntry, error) {
	query := rebindQuery(`SELECT we.id, we.session_id, we.user_id, we.position, we.status, we.joined_at, we.promoted_at, we.expired_at, we.created_at, we.updated_at,
		u.name as user_name, u.email as user_email, cs.name as session_name
		FROM waitlist_entries we
		JOIN users u ON we.user_id = u.id
		JOIN class_sessions cs ON we.session_id = cs.id
		WHERE we.id = ?`)

	entry := &domain.WaitlistEntry{}
	err := r.db.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.SessionID,
		&entry.UserID,
		&entry.Position,
		&entry.Status,
		&entry.JoinedAt,
		&entry.PromotedAt,
		&entry.ExpiredAt,
		&entry.CreatedAt,
		&entry.UpdatedAt,
		&entry.UserName,
		&entry.UserEmail,
		&entry.SessionName,
	)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// GetBySessionID retrieves all waitlist entries for a session
func (r *WaitlistRepository) GetBySessionID(sessionID int64) ([]*domain.WaitlistEntry, error) {
	query := rebindQuery(`SELECT we.id, we.session_id, we.user_id, we.position, we.status, we.joined_at, we.promoted_at, we.expired_at, we.created_at, we.updated_at,
		u.name as user_name, u.email as user_email
		FROM waitlist_entries we
		JOIN users u ON we.user_id = u.id
		WHERE we.session_id = ? AND we.status = 'waiting'
		ORDER BY we.position`)

	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*domain.WaitlistEntry
	for rows.Next() {
		entry := &domain.WaitlistEntry{}
		err := rows.Scan(
			&entry.ID,
			&entry.SessionID,
			&entry.UserID,
			&entry.Position,
			&entry.Status,
			&entry.JoinedAt,
			&entry.PromotedAt,
			&entry.ExpiredAt,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&entry.UserName,
			&entry.UserEmail,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

// GetByUserID retrieves all waitlist entries for a user
func (r *WaitlistRepository) GetByUserID(userID int64) ([]*domain.WaitlistEntry, error) {
	query := rebindQuery(`SELECT we.id, we.session_id, we.user_id, we.position, we.status, we.joined_at, we.promoted_at, we.expired_at, we.created_at, we.updated_at,
		cs.name as session_name
		FROM waitlist_entries we
		JOIN class_sessions cs ON we.session_id = cs.id
		WHERE we.user_id = ? AND we.status = 'waiting'
		ORDER BY cs.start_time`)

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*domain.WaitlistEntry
	for rows.Next() {
		entry := &domain.WaitlistEntry{}
		err := rows.Scan(
			&entry.ID,
			&entry.SessionID,
			&entry.UserID,
			&entry.Position,
			&entry.Status,
			&entry.JoinedAt,
			&entry.PromotedAt,
			&entry.ExpiredAt,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&entry.SessionName,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

// GetBySessionAndUser retrieves a waitlist entry for a specific session and user
func (r *WaitlistRepository) GetBySessionAndUser(sessionID, userID int64) (*domain.WaitlistEntry, error) {
	query := rebindQuery(`SELECT id, session_id, user_id, position, status, joined_at, promoted_at, expired_at, created_at, updated_at
		FROM waitlist_entries WHERE session_id = ? AND user_id = ?`)

	entry := &domain.WaitlistEntry{}
	err := r.db.QueryRow(query, sessionID, userID).Scan(
		&entry.ID,
		&entry.SessionID,
		&entry.UserID,
		&entry.Position,
		&entry.Status,
		&entry.JoinedAt,
		&entry.PromotedAt,
		&entry.ExpiredAt,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// GetNextInLine retrieves the first person waiting in line for a session
func (r *WaitlistRepository) GetNextInLine(sessionID int64) (*domain.WaitlistEntry, error) {
	query := rebindQuery(`SELECT we.id, we.session_id, we.user_id, we.position, we.status, we.joined_at, we.promoted_at, we.expired_at, we.created_at, we.updated_at,
		u.name as user_name, u.email as user_email
		FROM waitlist_entries we
		JOIN users u ON we.user_id = u.id
		WHERE we.session_id = ? AND we.status = 'waiting'
		ORDER BY we.position
		LIMIT 1`)

	entry := &domain.WaitlistEntry{}
	err := r.db.QueryRow(query, sessionID).Scan(
		&entry.ID,
		&entry.SessionID,
		&entry.UserID,
		&entry.Position,
		&entry.Status,
		&entry.JoinedAt,
		&entry.PromotedAt,
		&entry.ExpiredAt,
		&entry.CreatedAt,
		&entry.UpdatedAt,
		&entry.UserName,
		&entry.UserEmail,
	)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// GetPosition returns the position of a user on a session's waitlist
func (r *WaitlistRepository) GetPosition(sessionID, userID int64) (int, error) {
	query := rebindQuery(`SELECT position FROM waitlist_entries WHERE session_id = ? AND user_id = ? AND status = 'waiting'`)

	var position int
	err := r.db.QueryRow(query, sessionID, userID).Scan(&position)
	return position, err
}

// Promote marks a waitlist entry as promoted (given a spot)
func (r *WaitlistRepository) Promote(id int64) error {
	now := time.Now()

	query := rebindQuery(`UPDATE waitlist_entries SET status = 'promoted', promoted_at = ?, updated_at = ? WHERE id = ?`)

	_, err := r.db.Exec(query, now, now, id)
	return err
}

// Cancel cancels a waitlist entry
func (r *WaitlistRepository) Cancel(id int64) error {
	now := time.Now()

	query := rebindQuery(`UPDATE waitlist_entries SET status = 'cancelled', updated_at = ? WHERE id = ?`)

	_, err := r.db.Exec(query, now, id)
	return err
}

// ExpireOldEntries expires all waitlist entries for sessions that have ended
func (r *WaitlistRepository) ExpireOldEntries() (int, error) {
	now := time.Now()

	query := rebindQuery(`UPDATE waitlist_entries SET status = 'expired', expired_at = ?, updated_at = ?
		WHERE status = 'waiting' AND session_id IN (
			SELECT id FROM class_sessions WHERE start_time < ?
		)`)

	result, err := r.db.Exec(query, now, now, now)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	return int(count), err
}

// ReorderPositions reorders positions after a cancellation or promotion
func (r *WaitlistRepository) ReorderPositions(sessionID int64) error {
	// Get all waiting entries ordered by position
	query := rebindQuery(`SELECT id FROM waitlist_entries WHERE session_id = ? AND status = 'waiting' ORDER BY position`)

	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Update positions sequentially
	now := time.Now()
	updateQuery := rebindQuery(`UPDATE waitlist_entries SET position = ?, updated_at = ? WHERE id = ?`)
	for i, id := range ids {
		if _, err := r.db.Exec(updateQuery, i+1, now, id); err != nil {
			return err
		}
	}
	return nil
}

// Delete deletes a waitlist entry
func (r *WaitlistRepository) Delete(id int64) error {
	query := rebindQuery(`DELETE FROM waitlist_entries WHERE id = ?`)
	_, err := r.db.Exec(query, id)
	return err
}

// ClassNotificationRepository implements domain.ClassNotificationRepository
type ClassNotificationRepository struct {
	db *sql.DB
}

// NewClassNotificationRepository creates a new notification repository
func NewClassNotificationRepository(db *sql.DB) *ClassNotificationRepository {
	return &ClassNotificationRepository{db: db}
}

// Create creates a new notification
func (r *ClassNotificationRepository) Create(notification *domain.ClassNotification) error {
	now := time.Now()
	notification.CreatedAt = now
	notification.UpdatedAt = now

	query := rebindQuery(`INSERT INTO class_notifications (user_id, session_id, reservation_id, waitlist_entry_id, notification_type, status, scheduled_for, sent_at, error_message, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			notification.UserID,
			notification.SessionID,
			notification.ReservationID,
			notification.WaitlistEntryID,
			notification.NotificationType,
			notification.Status,
			notification.ScheduledFor,
			notification.SentAt,
			notification.ErrorMessage,
			notification.CreatedAt,
			notification.UpdatedAt,
		).Scan(&notification.ID)
	}

	result, err := r.db.Exec(query,
		notification.UserID,
		notification.SessionID,
		notification.ReservationID,
		notification.WaitlistEntryID,
		notification.NotificationType,
		notification.Status,
		notification.ScheduledFor,
		notification.SentAt,
		notification.ErrorMessage,
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
func (r *ClassNotificationRepository) GetByID(id int64) (*domain.ClassNotification, error) {
	query := rebindQuery(`SELECT cn.id, cn.user_id, cn.session_id, cn.reservation_id, cn.waitlist_entry_id, cn.notification_type, cn.status, cn.scheduled_for, cn.sent_at, cn.error_message, cn.created_at, cn.updated_at,
		u.name as user_name, u.email as user_email, cs.name as session_name
		FROM class_notifications cn
		JOIN users u ON cn.user_id = u.id
		LEFT JOIN class_sessions cs ON cn.session_id = cs.id
		WHERE cn.id = ?`)

	notification := &domain.ClassNotification{}
	err := r.db.QueryRow(query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.SessionID,
		&notification.ReservationID,
		&notification.WaitlistEntryID,
		&notification.NotificationType,
		&notification.Status,
		&notification.ScheduledFor,
		&notification.SentAt,
		&notification.ErrorMessage,
		&notification.CreatedAt,
		&notification.UpdatedAt,
		&notification.UserName,
		&notification.UserEmail,
		&notification.SessionName,
	)
	if err != nil {
		return nil, err
	}
	return notification, nil
}

// GetPendingByScheduledTime retrieves all pending notifications scheduled before a given time
func (r *ClassNotificationRepository) GetPendingByScheduledTime(before time.Time) ([]*domain.ClassNotification, error) {
	query := rebindQuery(`SELECT cn.id, cn.user_id, cn.session_id, cn.reservation_id, cn.waitlist_entry_id, cn.notification_type, cn.status, cn.scheduled_for, cn.sent_at, cn.error_message, cn.created_at, cn.updated_at,
		u.name as user_name, u.email as user_email, cs.name as session_name
		FROM class_notifications cn
		JOIN users u ON cn.user_id = u.id
		LEFT JOIN class_sessions cs ON cn.session_id = cs.id
		WHERE cn.status = 'pending' AND cn.scheduled_for <= ?
		ORDER BY cn.scheduled_for`)

	rows, err := r.db.Query(query, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*domain.ClassNotification
	for rows.Next() {
		notification := &domain.ClassNotification{}
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.SessionID,
			&notification.ReservationID,
			&notification.WaitlistEntryID,
			&notification.NotificationType,
			&notification.Status,
			&notification.ScheduledFor,
			&notification.SentAt,
			&notification.ErrorMessage,
			&notification.CreatedAt,
			&notification.UpdatedAt,
			&notification.UserName,
			&notification.UserEmail,
			&notification.SessionName,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	return notifications, rows.Err()
}

// GetByUserID retrieves recent notifications for a user
func (r *ClassNotificationRepository) GetByUserID(userID int64, limit int) ([]*domain.ClassNotification, error) {
	query := rebindQuery(`SELECT cn.id, cn.user_id, cn.session_id, cn.reservation_id, cn.waitlist_entry_id, cn.notification_type, cn.status, cn.scheduled_for, cn.sent_at, cn.error_message, cn.created_at, cn.updated_at,
		cs.name as session_name
		FROM class_notifications cn
		LEFT JOIN class_sessions cs ON cn.session_id = cs.id
		WHERE cn.user_id = ?
		ORDER BY cn.created_at DESC
		LIMIT ?`)

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*domain.ClassNotification
	for rows.Next() {
		notification := &domain.ClassNotification{}
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.SessionID,
			&notification.ReservationID,
			&notification.WaitlistEntryID,
			&notification.NotificationType,
			&notification.Status,
			&notification.ScheduledFor,
			&notification.SentAt,
			&notification.ErrorMessage,
			&notification.CreatedAt,
			&notification.UpdatedAt,
			&notification.SessionName,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	return notifications, rows.Err()
}

// MarkSent marks a notification as sent
func (r *ClassNotificationRepository) MarkSent(id int64) error {
	now := time.Now()

	query := rebindQuery(`UPDATE class_notifications SET status = 'sent', sent_at = ?, updated_at = ? WHERE id = ?`)

	_, err := r.db.Exec(query, now, now, id)
	return err
}

// MarkFailed marks a notification as failed
func (r *ClassNotificationRepository) MarkFailed(id int64, errorMessage string) error {
	now := time.Now()

	query := rebindQuery(`UPDATE class_notifications SET status = 'failed', error_message = ?, updated_at = ? WHERE id = ?`)

	_, err := r.db.Exec(query, errorMessage, now, id)
	return err
}

// DeleteBySessionID deletes all notifications for a session
func (r *ClassNotificationRepository) DeleteBySessionID(sessionID int64) error {
	query := rebindQuery(`DELETE FROM class_notifications WHERE session_id = ?`)
	_, err := r.db.Exec(query, sessionID)
	return err
}

// DeleteByReservationID deletes all notifications for a reservation
func (r *ClassNotificationRepository) DeleteByReservationID(reservationID int64) error {
	query := rebindQuery(`DELETE FROM class_notifications WHERE reservation_id = ?`)
	_, err := r.db.Exec(query, reservationID)
	return err
}
