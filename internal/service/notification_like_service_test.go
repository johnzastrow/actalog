package service

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestNewNotificationLikeService(t *testing.T) {
	likeRepo := newMockNotificationLikeRepo()
	notificationRepo := newMockNotificationRepo()
	svc := NewNotificationLikeService(likeRepo, notificationRepo)

	if svc == nil {
		t.Fatal("NewNotificationLikeService returned nil")
	}
}

func TestNotificationLikeService_LikeNotification(t *testing.T) {
	likeRepo := newMockNotificationLikeRepo()
	notificationRepo := newMockNotificationRepo()
	svc := NewNotificationLikeService(likeRepo, notificationRepo)

	// Create a notification first (mark it as read initially)
	now := time.Now()
	notification := &domain.Notification{
		UserID:  1,
		Type:    "pr_achieved",
		Title:   "New PR!",
		Message: "You achieved a new personal record",
		ReadAt:  &now, // Already read
	}
	_ = notificationRepo.Create(notification)

	// Like the notification
	err := svc.LikeNotification(notification.ID, 2)
	if err != nil {
		t.Errorf("LikeNotification() error = %v", err)
	}

	// Verify like was created
	if len(likeRepo.likes) != 1 {
		t.Fatalf("Expected 1 like, got %d", len(likeRepo.likes))
	}

	like := likeRepo.likes[0]
	if like.NotificationID != notification.ID {
		t.Errorf("NotificationID = %d, want %d", like.NotificationID, notification.ID)
	}
	if like.UserID != 2 {
		t.Errorf("UserID = %d, want 2", like.UserID)
	}

	// Verify notification was marked as unread (ReadAt should be nil)
	updatedNotification, _ := notificationRepo.GetByID(notification.ID)
	if updatedNotification.ReadAt != nil {
		t.Error("Notification should be marked as unread after being liked")
	}
}

func TestNotificationLikeService_LikeNotification_AlreadyLiked(t *testing.T) {
	likeRepo := newMockNotificationLikeRepo()
	notificationRepo := newMockNotificationRepo()
	svc := NewNotificationLikeService(likeRepo, notificationRepo)

	// Create a notification
	notification := &domain.Notification{
		UserID:  1,
		Type:    "pr_achieved",
		Title:   "New PR!",
		Message: "You achieved a new personal record",
	}
	_ = notificationRepo.Create(notification)

	// Like the notification
	_ = svc.LikeNotification(notification.ID, 2)

	// Try to like again - should fail
	err := svc.LikeNotification(notification.ID, 2)
	if err == nil {
		t.Error("LikeNotification() should return error when already liked")
	}
}

func TestNotificationLikeService_UnlikeNotification(t *testing.T) {
	likeRepo := newMockNotificationLikeRepo()
	notificationRepo := newMockNotificationRepo()
	svc := NewNotificationLikeService(likeRepo, notificationRepo)

	// Create a notification and like it
	notification := &domain.Notification{
		UserID:  1,
		Type:    "pr_achieved",
		Title:   "New PR!",
		Message: "You achieved a new personal record",
	}
	_ = notificationRepo.Create(notification)
	_ = svc.LikeNotification(notification.ID, 2)

	// Unlike the notification
	err := svc.UnlikeNotification(notification.ID, 2)
	if err != nil {
		t.Errorf("UnlikeNotification() error = %v", err)
	}

	// Verify like was removed
	if len(likeRepo.likes) != 0 {
		t.Errorf("Expected 0 likes, got %d", len(likeRepo.likes))
	}
}

func TestNotificationLikeService_UnlikeNotification_NotLiked(t *testing.T) {
	likeRepo := newMockNotificationLikeRepo()
	notificationRepo := newMockNotificationRepo()
	svc := NewNotificationLikeService(likeRepo, notificationRepo)

	// Try to unlike without liking first
	err := svc.UnlikeNotification(999, 2)
	if err == nil {
		t.Error("UnlikeNotification() should return error when not liked")
	}
}

func TestNotificationLikeService_GetNotificationLikes(t *testing.T) {
	likeRepo := newMockNotificationLikeRepo()
	notificationRepo := newMockNotificationRepo()
	svc := NewNotificationLikeService(likeRepo, notificationRepo)

	// Create a notification
	notification := &domain.Notification{
		UserID:  1,
		Type:    "pr_achieved",
		Title:   "New PR!",
		Message: "You achieved a new personal record",
	}
	_ = notificationRepo.Create(notification)

	// Add multiple likes
	_ = svc.LikeNotification(notification.ID, 2)
	_ = svc.LikeNotification(notification.ID, 3)
	_ = svc.LikeNotification(notification.ID, 4)

	// Get likes
	likes, err := svc.GetNotificationLikes(notification.ID)
	if err != nil {
		t.Errorf("GetNotificationLikes() error = %v", err)
	}
	if len(likes) != 3 {
		t.Errorf("GetNotificationLikes() returned %d likes, want 3", len(likes))
	}
}

func TestNotificationLikeService_GetLikeCount(t *testing.T) {
	likeRepo := newMockNotificationLikeRepo()
	notificationRepo := newMockNotificationRepo()
	svc := NewNotificationLikeService(likeRepo, notificationRepo)

	// Create a notification
	notification := &domain.Notification{
		UserID:  1,
		Type:    "pr_achieved",
		Title:   "New PR!",
		Message: "You achieved a new personal record",
	}
	_ = notificationRepo.Create(notification)

	// Check count before any likes
	count, err := svc.GetLikeCount(notification.ID)
	if err != nil {
		t.Errorf("GetLikeCount() error = %v", err)
	}
	if count != 0 {
		t.Errorf("GetLikeCount() = %d, want 0", count)
	}

	// Add likes
	_ = svc.LikeNotification(notification.ID, 2)
	_ = svc.LikeNotification(notification.ID, 3)

	// Check count after likes
	count, err = svc.GetLikeCount(notification.ID)
	if err != nil {
		t.Errorf("GetLikeCount() error = %v", err)
	}
	if count != 2 {
		t.Errorf("GetLikeCount() = %d, want 2", count)
	}
}

func TestNotificationLikeService_HasUserLiked(t *testing.T) {
	likeRepo := newMockNotificationLikeRepo()
	notificationRepo := newMockNotificationRepo()
	svc := NewNotificationLikeService(likeRepo, notificationRepo)

	// Create a notification
	notification := &domain.Notification{
		UserID:  1,
		Type:    "pr_achieved",
		Title:   "New PR!",
		Message: "You achieved a new personal record",
	}
	_ = notificationRepo.Create(notification)

	// Check before liking
	hasLiked, err := svc.HasUserLiked(notification.ID, 2)
	if err != nil {
		t.Errorf("HasUserLiked() error = %v", err)
	}
	if hasLiked {
		t.Error("HasUserLiked() = true, want false (before liking)")
	}

	// Like the notification
	_ = svc.LikeNotification(notification.ID, 2)

	// Check after liking
	hasLiked, err = svc.HasUserLiked(notification.ID, 2)
	if err != nil {
		t.Errorf("HasUserLiked() error = %v", err)
	}
	if !hasLiked {
		t.Error("HasUserLiked() = false, want true (after liking)")
	}

	// Check for a different user
	hasLiked, err = svc.HasUserLiked(notification.ID, 3)
	if err != nil {
		t.Errorf("HasUserLiked() error = %v", err)
	}
	if hasLiked {
		t.Error("HasUserLiked() = true, want false (different user)")
	}
}

func TestNotificationLikeService_MultipleLikesFromDifferentUsers(t *testing.T) {
	likeRepo := newMockNotificationLikeRepo()
	notificationRepo := newMockNotificationRepo()
	svc := NewNotificationLikeService(likeRepo, notificationRepo)

	// Create a notification
	notification := &domain.Notification{
		UserID:  1,
		Type:    "pr_achieved",
		Title:   "New PR!",
		Message: "You achieved a new personal record",
	}
	_ = notificationRepo.Create(notification)

	// Multiple users like the same notification
	users := []int64{2, 3, 4, 5, 6}
	for _, userID := range users {
		err := svc.LikeNotification(notification.ID, userID)
		if err != nil {
			t.Errorf("LikeNotification() for user %d error = %v", userID, err)
		}
	}

	// Verify count
	count, err := svc.GetLikeCount(notification.ID)
	if err != nil {
		t.Errorf("GetLikeCount() error = %v", err)
	}
	if count != len(users) {
		t.Errorf("GetLikeCount() = %d, want %d", count, len(users))
	}

	// Verify each user has liked
	for _, userID := range users {
		hasLiked, err := svc.HasUserLiked(notification.ID, userID)
		if err != nil {
			t.Errorf("HasUserLiked() for user %d error = %v", userID, err)
		}
		if !hasLiked {
			t.Errorf("HasUserLiked() for user %d = false, want true", userID)
		}
	}
}

func TestNotificationLikeService_LikesDifferentNotifications(t *testing.T) {
	likeRepo := newMockNotificationLikeRepo()
	notificationRepo := newMockNotificationRepo()
	svc := NewNotificationLikeService(likeRepo, notificationRepo)

	// Create multiple notifications
	notification1 := &domain.Notification{UserID: 1, Type: "pr_achieved", Title: "PR 1", Message: "Message 1"}
	notification2 := &domain.Notification{UserID: 1, Type: "pr_achieved", Title: "PR 2", Message: "Message 2"}
	_ = notificationRepo.Create(notification1)
	_ = notificationRepo.Create(notification2)

	// User likes both notifications
	_ = svc.LikeNotification(notification1.ID, 2)
	_ = svc.LikeNotification(notification2.ID, 2)

	// Verify likes are independent
	count1, _ := svc.GetLikeCount(notification1.ID)
	count2, _ := svc.GetLikeCount(notification2.ID)

	if count1 != 1 {
		t.Errorf("Notification 1 like count = %d, want 1", count1)
	}
	if count2 != 1 {
		t.Errorf("Notification 2 like count = %d, want 1", count2)
	}
}
