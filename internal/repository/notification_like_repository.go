package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// NotificationLikeRepository implements the NotificationLikeRepository interface
type NotificationLikeRepository struct {
	db *sql.DB
}

// NewNotificationLikeRepository creates a new notification like repository
func NewNotificationLikeRepository(db *sql.DB) *NotificationLikeRepository {
	return &NotificationLikeRepository{db: db}
}

// Create adds a like to a notification
func (r *NotificationLikeRepository) Create(like *domain.NotificationLike) error {
	query := `
		INSERT INTO notification_likes (notification_id, user_id, created_at)
		VALUES (?, ?, ?)
	`
	result, err := r.db.Exec(query, like.NotificationID, like.UserID, like.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create notification like: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	like.ID = id
	return nil
}

// Delete removes a like
func (r *NotificationLikeRepository) Delete(notificationID, userID int64) error {
	query := `DELETE FROM notification_likes WHERE notification_id = ? AND user_id = ?`
	_, err := r.db.Exec(query, notificationID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete notification like: %w", err)
	}
	return nil
}

// GetByNotificationID returns all likes for a notification with user details
func (r *NotificationLikeRepository) GetByNotificationID(notificationID int64) ([]*domain.NotificationLike, error) {
	query := `
		SELECT nl.id, nl.notification_id, nl.user_id, nl.created_at, u.name, u.email
		FROM notification_likes nl
		JOIN users u ON nl.user_id = u.id
		WHERE nl.notification_id = ?
		ORDER BY nl.created_at DESC
	`

	rows, err := r.db.Query(query, notificationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification likes: %w", err)
	}
	defer rows.Close()

	var likes []*domain.NotificationLike
	for rows.Next() {
		like := &domain.NotificationLike{}
		var createdAt string
		if err := rows.Scan(&like.ID, &like.NotificationID, &like.UserID,
			&createdAt, &like.UserName, &like.UserEmail); err != nil {
			return nil, fmt.Errorf("failed to scan notification like: %w", err)
		}

		// Parse the timestamp
		parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAt)
		if err != nil {
			// Try alternative format
			parsedTime, err = time.Parse(time.RFC3339, createdAt)
			if err != nil {
				return nil, fmt.Errorf("failed to parse created_at: %w", err)
			}
		}
		like.CreatedAt = parsedTime

		likes = append(likes, like)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notification likes: %w", err)
	}

	return likes, nil
}

// GetLikeCount returns the count of likes for a notification
func (r *NotificationLikeRepository) GetLikeCount(notificationID int64) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM notification_likes WHERE notification_id = ?`
	err := r.db.QueryRow(query, notificationID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get like count: %w", err)
	}
	return count, nil
}

// HasUserLiked checks if a user has liked a notification
func (r *NotificationLikeRepository) HasUserLiked(notificationID, userID int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM notification_likes WHERE notification_id = ? AND user_id = ?`
	err := r.db.QueryRow(query, notificationID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if user liked: %w", err)
	}
	return count > 0, nil
}
