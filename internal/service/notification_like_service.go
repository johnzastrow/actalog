package service

import (
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// NotificationLikeService handles notification like business logic
type NotificationLikeService struct {
	likeRepo         domain.NotificationLikeRepository
	notificationRepo domain.NotificationRepository
}

// NewNotificationLikeService creates a new notification like service
func NewNotificationLikeService(
	likeRepo domain.NotificationLikeRepository,
	notificationRepo domain.NotificationRepository,
) *NotificationLikeService {
	return &NotificationLikeService{
		likeRepo:         likeRepo,
		notificationRepo: notificationRepo,
	}
}

// LikeNotification adds a like to a notification and marks it unread for the original recipient
func (s *NotificationLikeService) LikeNotification(notificationID, likingUserID int64) error {
	// Check if already liked
	hasLiked, err := s.likeRepo.HasUserLiked(notificationID, likingUserID)
	if err != nil {
		return fmt.Errorf("failed to check existing like: %w", err)
	}
	if hasLiked {
		return fmt.Errorf("user has already liked this notification")
	}

	// Create the like
	like := &domain.NotificationLike{
		NotificationID: notificationID,
		UserID:         likingUserID,
		CreatedAt:      time.Now(),
	}

	if err := s.likeRepo.Create(like); err != nil {
		return fmt.Errorf("failed to create like: %w", err)
	}

	// Mark notification as unread for the original recipient
	// (This ensures they see that someone liked their notification)
	if err := s.notificationRepo.MarkAsUnread(notificationID); err != nil {
		// Log error but don't fail the like operation
		fmt.Printf("Failed to mark notification %d as unread: %v\n", notificationID, err)
	}

	return nil
}

// UnlikeNotification removes a like from a notification
func (s *NotificationLikeService) UnlikeNotification(notificationID, likingUserID int64) error {
	return s.likeRepo.Delete(notificationID, likingUserID)
}

// GetNotificationLikes returns all likes for a notification
func (s *NotificationLikeService) GetNotificationLikes(notificationID int64) ([]*domain.NotificationLike, error) {
	return s.likeRepo.GetByNotificationID(notificationID)
}

// GetLikeCount returns the count of likes for a notification
func (s *NotificationLikeService) GetLikeCount(notificationID int64) (int, error) {
	return s.likeRepo.GetLikeCount(notificationID)
}

// HasUserLiked checks if a user has liked a notification
func (s *NotificationLikeService) HasUserLiked(notificationID, likingUserID int64) (bool, error) {
	return s.likeRepo.HasUserLiked(notificationID, likingUserID)
}
