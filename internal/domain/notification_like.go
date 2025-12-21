package domain

import "time"

// NotificationLike represents a like on a notification
type NotificationLike struct {
	ID             int64     `json:"id"`
	NotificationID int64     `json:"notification_id"`
	UserID         int64     `json:"user_id"`
	CreatedAt      time.Time `json:"created_at"`

	// Populated by joins
	UserName  string `json:"user_name,omitempty"`
	UserEmail string `json:"user_email,omitempty"`
}

// NotificationLikeRepository defines the interface for notification like data access
type NotificationLikeRepository interface {
	// Create adds a like to a notification
	Create(like *NotificationLike) error

	// Delete removes a like
	Delete(notificationID, userID int64) error

	// GetByNotificationID returns all likes for a notification with user details
	GetByNotificationID(notificationID int64) ([]*NotificationLike, error)

	// GetLikeCount returns the count of likes for a notification
	GetLikeCount(notificationID int64) (int, error)

	// HasUserLiked checks if a user has liked a notification
	HasUserLiked(notificationID, userID int64) (bool, error)
}
