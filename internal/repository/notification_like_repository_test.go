package repository

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestNotificationLikeRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	notifRepo := NewNotificationRepository(db)
	likeRepo := NewNotificationLikeRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
		Role:         "athlete",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create test notification
	notif := &domain.Notification{
		UserID:  user.ID,
		Type:    "pr_achieved",
		Title:   "New PR",
		Message: "You achieved a new PR!",
	}
	if err := notifRepo.Create(notif); err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	// Create like
	like := &domain.NotificationLike{
		NotificationID: notif.ID,
		UserID:         user.ID,
		CreatedAt:      time.Now(),
	}

	err = likeRepo.Create(like)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if like.ID == 0 {
		t.Error("Create() should set ID")
	}
}

func TestNotificationLikeRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	notifRepo := NewNotificationRepository(db)
	likeRepo := NewNotificationLikeRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
		Role:         "athlete",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create test notification
	notif := &domain.Notification{
		UserID:  user.ID,
		Type:    "pr_achieved",
		Title:   "New PR",
		Message: "You achieved a new PR!",
	}
	if err := notifRepo.Create(notif); err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	// Create like
	like := &domain.NotificationLike{
		NotificationID: notif.ID,
		UserID:         user.ID,
		CreatedAt:      time.Now(),
	}
	if err := likeRepo.Create(like); err != nil {
		t.Fatalf("Failed to create like: %v", err)
	}

	// Verify like exists
	hasLiked, err := likeRepo.HasUserLiked(notif.ID, user.ID)
	if err != nil {
		t.Fatalf("HasUserLiked() error = %v", err)
	}
	if !hasLiked {
		t.Fatal("Like should exist")
	}

	// Delete like
	err = likeRepo.Delete(notif.ID, user.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify like is deleted
	hasLiked, err = likeRepo.HasUserLiked(notif.ID, user.ID)
	if err != nil {
		t.Fatalf("HasUserLiked() error = %v", err)
	}
	if hasLiked {
		t.Error("Like should be deleted")
	}
}

func TestNotificationLikeRepository_GetByNotificationID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	notifRepo := NewNotificationRepository(db)
	likeRepo := NewNotificationLikeRepository(db)

	// Create test users
	users := []*domain.User{
		{Email: "user1@example.com", PasswordHash: "hash", Name: "User 1", Role: "athlete"},
		{Email: "user2@example.com", PasswordHash: "hash", Name: "User 2", Role: "athlete"},
		{Email: "user3@example.com", PasswordHash: "hash", Name: "User 3", Role: "athlete"},
	}
	for _, u := range users {
		if err := userRepo.Create(u); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// Create test notification
	notif := &domain.Notification{
		UserID:  users[0].ID,
		Type:    "pr_achieved",
		Title:   "New PR",
		Message: "You achieved a new PR!",
	}
	if err := notifRepo.Create(notif); err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	// Create likes from multiple users
	for _, u := range users {
		like := &domain.NotificationLike{
			NotificationID: notif.ID,
			UserID:         u.ID,
			CreatedAt:      time.Now(),
		}
		if err := likeRepo.Create(like); err != nil {
			t.Fatalf("Failed to create like: %v", err)
		}
	}

	// Get likes
	likes, err := likeRepo.GetByNotificationID(notif.ID)
	if err != nil {
		t.Fatalf("GetByNotificationID() error = %v", err)
	}
	if len(likes) != 3 {
		t.Errorf("GetByNotificationID() returned %d likes, want 3", len(likes))
	}

	// Verify user info is populated
	for _, like := range likes {
		if like.UserName == "" {
			t.Error("UserName should be populated from JOIN")
		}
		if like.UserEmail == "" {
			t.Error("UserEmail should be populated from JOIN")
		}
	}
}

func TestNotificationLikeRepository_GetLikeCount(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	notifRepo := NewNotificationRepository(db)
	likeRepo := NewNotificationLikeRepository(db)

	// Create test users
	users := make([]*domain.User, 5)
	for i := 0; i < 5; i++ {
		users[i] = &domain.User{
			Email:        "user" + string(rune('0'+i)) + "@example.com",
			PasswordHash: "hash",
			Name:         "User " + string(rune('0'+i)),
			Role:         "athlete",
		}
		if err := userRepo.Create(users[i]); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// Create test notification
	notif := &domain.Notification{
		UserID:  users[0].ID,
		Type:    "pr_achieved",
		Title:   "New PR",
		Message: "You achieved a new PR!",
	}
	if err := notifRepo.Create(notif); err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	// Initially count should be 0
	count, err := likeRepo.GetLikeCount(notif.ID)
	if err != nil {
		t.Fatalf("GetLikeCount() error = %v", err)
	}
	if count != 0 {
		t.Errorf("GetLikeCount() = %d, want 0", count)
	}

	// Add likes
	for i := 0; i < 3; i++ {
		like := &domain.NotificationLike{
			NotificationID: notif.ID,
			UserID:         users[i].ID,
			CreatedAt:      time.Now(),
		}
		if err := likeRepo.Create(like); err != nil {
			t.Fatalf("Failed to create like: %v", err)
		}
	}

	// Count should be 3
	count, err = likeRepo.GetLikeCount(notif.ID)
	if err != nil {
		t.Fatalf("GetLikeCount() error = %v", err)
	}
	if count != 3 {
		t.Errorf("GetLikeCount() = %d, want 3", count)
	}
}

func TestNotificationLikeRepository_HasUserLiked(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	notifRepo := NewNotificationRepository(db)
	likeRepo := NewNotificationLikeRepository(db)

	// Create test users
	user1 := &domain.User{Email: "user1@example.com", PasswordHash: "hash", Name: "User 1", Role: "athlete"}
	user2 := &domain.User{Email: "user2@example.com", PasswordHash: "hash", Name: "User 2", Role: "athlete"}
	if err := userRepo.Create(user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	if err := userRepo.Create(user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Create test notification
	notif := &domain.Notification{
		UserID:  user1.ID,
		Type:    "pr_achieved",
		Title:   "New PR",
		Message: "You achieved a new PR!",
	}
	if err := notifRepo.Create(notif); err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	// User 1 has not liked yet
	hasLiked, err := likeRepo.HasUserLiked(notif.ID, user1.ID)
	if err != nil {
		t.Fatalf("HasUserLiked() error = %v", err)
	}
	if hasLiked {
		t.Error("User 1 should not have liked yet")
	}

	// User 1 likes
	like := &domain.NotificationLike{
		NotificationID: notif.ID,
		UserID:         user1.ID,
		CreatedAt:      time.Now(),
	}
	if err := likeRepo.Create(like); err != nil {
		t.Fatalf("Failed to create like: %v", err)
	}

	// User 1 has liked
	hasLiked, err = likeRepo.HasUserLiked(notif.ID, user1.ID)
	if err != nil {
		t.Fatalf("HasUserLiked() error = %v", err)
	}
	if !hasLiked {
		t.Error("User 1 should have liked")
	}

	// User 2 has not liked
	hasLiked, err = likeRepo.HasUserLiked(notif.ID, user2.ID)
	if err != nil {
		t.Fatalf("HasUserLiked() error = %v", err)
	}
	if hasLiked {
		t.Error("User 2 should not have liked")
	}
}
