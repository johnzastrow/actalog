package repository

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestNotificationRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create a user first
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "notifyuser@example.com")

	repo := NewNotificationRepository(db)

	tests := []struct {
		name         string
		notification *domain.Notification
		wantErr      bool
	}{
		{
			name: "create PR achievement notification",
			notification: &domain.Notification{
				UserID:    user.ID,
				Type:      domain.NotificationTypePRAchievement,
				Title:     "New PR!",
				Message:   "You set a new personal record on Back Squat: 315 lbs",
				Data:      `{"movement_name": "Back Squat", "value": "315 lbs"}`,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "create weekly streak notification",
			notification: &domain.Notification{
				UserID:    user.ID,
				Type:      domain.NotificationTypeWeeklyStreak,
				Title:     "5 Week Streak!",
				Message:   "You've worked out for 5 weeks in a row!",
				Data:      `{"count": 5}`,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "create WOD milestone notification",
			notification: &domain.Notification{
				UserID:    user.ID,
				Type:      domain.NotificationTypeWODMilestone,
				Title:     "10 Hero WODs!",
				Message:   "You've completed 10 Hero WODs!",
				Data:      `{"wod_type": "hero", "count": 10}`,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.notification)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.notification.ID == 0 {
				t.Error("Create() should set notification ID")
			}
		})
	}
}

func TestNotificationRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create a user first
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "getnotify@example.com")

	repo := NewNotificationRepository(db)

	// Create a test notification
	notification := &domain.Notification{
		UserID:    user.ID,
		Type:      domain.NotificationTypePRAchievement,
		Title:     "Test Notification",
		Message:   "This is a test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repo.Create(notification); err != nil {
		t.Fatalf("failed to create test notification: %v", err)
	}

	tests := []struct {
		name             string
		id               int64
		wantNotification bool
		wantTitle        string
	}{
		{
			name:             "existing notification",
			id:               notification.ID,
			wantNotification: true,
			wantTitle:        "Test Notification",
		},
		{
			name:             "non-existent notification",
			id:               99999,
			wantNotification: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(tt.id)
			if err != nil {
				t.Errorf("GetByID() error = %v", err)
				return
			}
			if tt.wantNotification && got == nil {
				t.Error("GetByID() expected notification but got nil")
				return
			}
			if !tt.wantNotification && got != nil {
				t.Errorf("GetByID() expected nil but got notification: %+v", got)
				return
			}
			if tt.wantNotification && got.Title != tt.wantTitle {
				t.Errorf("GetByID() title = %v, want %v", got.Title, tt.wantTitle)
			}
		})
	}
}

func TestNotificationRepository_GetByUserID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create users
	userRepo := NewSQLiteUserRepository(db)
	user1 := createTestUser(t, userRepo, "user1notify@example.com")
	user2 := createTestUser(t, userRepo, "user2notify@example.com")

	repo := NewNotificationRepository(db)

	// Create notifications for user1
	for i := 0; i < 5; i++ {
		n := &domain.Notification{
			UserID:    user1.ID,
			Type:      domain.NotificationTypePRAchievement,
			Title:     "User1 Notification " + string(rune('A'+i)),
			Message:   "Test message",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
			UpdatedAt: time.Now(),
		}
		if err := repo.Create(n); err != nil {
			t.Fatalf("failed to create notification: %v", err)
		}
	}

	// Create notifications for user2
	for i := 0; i < 3; i++ {
		n := &domain.Notification{
			UserID:    user2.ID,
			Type:      domain.NotificationTypeWeeklyStreak,
			Title:     "User2 Notification " + string(rune('A'+i)),
			Message:   "Test message",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := repo.Create(n); err != nil {
			t.Fatalf("failed to create notification: %v", err)
		}
	}

	tests := []struct {
		name      string
		userID    int64
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "all user1 notifications",
			userID:    user1.ID,
			limit:     10,
			offset:    0,
			wantCount: 5,
		},
		{
			name:      "all user2 notifications",
			userID:    user2.ID,
			limit:     10,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "user1 with limit",
			userID:    user1.ID,
			limit:     3,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "user1 with offset",
			userID:    user1.ID,
			limit:     10,
			offset:    2,
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifications, err := repo.GetByUserID(tt.userID, tt.limit, tt.offset)
			if err != nil {
				t.Errorf("GetByUserID() error = %v", err)
				return
			}
			if len(notifications) != tt.wantCount {
				t.Errorf("GetByUserID() returned %d notifications, want %d", len(notifications), tt.wantCount)
			}
		})
	}
}

func TestNotificationRepository_GetUnreadByUserID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create user
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "unreadnotify@example.com")

	repo := NewNotificationRepository(db)

	// Create some notifications
	var notifications []*domain.Notification
	for i := 0; i < 5; i++ {
		n := &domain.Notification{
			UserID:    user.ID,
			Type:      domain.NotificationTypePRAchievement,
			Title:     "Notification " + string(rune('A'+i)),
			Message:   "Test message",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := repo.Create(n); err != nil {
			t.Fatalf("failed to create notification: %v", err)
		}
		notifications = append(notifications, n)
	}

	// Mark 2 as read
	if err := repo.MarkAsRead(notifications[0].ID); err != nil {
		t.Fatalf("failed to mark as read: %v", err)
	}
	if err := repo.MarkAsRead(notifications[1].ID); err != nil {
		t.Fatalf("failed to mark as read: %v", err)
	}

	// Get unread - should be 3
	unread, err := repo.GetUnreadByUserID(user.ID, 10, 0)
	if err != nil {
		t.Fatalf("GetUnreadByUserID() error = %v", err)
	}

	if len(unread) != 3 {
		t.Errorf("GetUnreadByUserID() returned %d, want 3", len(unread))
	}

	// Verify all returned are unread
	for _, n := range unread {
		if n.ReadAt != nil {
			t.Errorf("GetUnreadByUserID() returned read notification: %d", n.ID)
		}
	}
}

func TestNotificationRepository_CountUnreadByUserID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create user
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "countunread@example.com")

	repo := NewNotificationRepository(db)

	// Initial count should be 0
	count, err := repo.CountUnreadByUserID(user.ID)
	if err != nil {
		t.Fatalf("CountUnreadByUserID() error = %v", err)
	}
	if count != 0 {
		t.Errorf("CountUnreadByUserID() initial = %d, want 0", count)
	}

	// Create notifications
	var notifications []*domain.Notification
	for i := 0; i < 5; i++ {
		n := &domain.Notification{
			UserID:    user.ID,
			Type:      domain.NotificationTypePRAchievement,
			Title:     "Count Notification " + string(rune('A'+i)),
			Message:   "Test message",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := repo.Create(n); err != nil {
			t.Fatalf("failed to create notification: %v", err)
		}
		notifications = append(notifications, n)
	}

	// Count should be 5
	count, err = repo.CountUnreadByUserID(user.ID)
	if err != nil {
		t.Fatalf("CountUnreadByUserID() error = %v", err)
	}
	if count != 5 {
		t.Errorf("CountUnreadByUserID() = %d, want 5", count)
	}

	// Mark 2 as read
	if err := repo.MarkAsRead(notifications[0].ID); err != nil {
		t.Fatalf("failed to mark as read: %v", err)
	}
	if err := repo.MarkAsRead(notifications[1].ID); err != nil {
		t.Fatalf("failed to mark as read: %v", err)
	}

	// Count should be 3
	count, err = repo.CountUnreadByUserID(user.ID)
	if err != nil {
		t.Fatalf("CountUnreadByUserID() error = %v", err)
	}
	if count != 3 {
		t.Errorf("CountUnreadByUserID() after marking = %d, want 3", count)
	}
}

func TestNotificationRepository_MarkAsRead(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create user
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "markread@example.com")

	repo := NewNotificationRepository(db)

	// Create notification
	notification := &domain.Notification{
		UserID:    user.ID,
		Type:      domain.NotificationTypePRAchievement,
		Title:     "To Mark Read",
		Message:   "Test message",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repo.Create(notification); err != nil {
		t.Fatalf("failed to create notification: %v", err)
	}

	// Initially should be unread
	n, err := repo.GetByID(notification.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if n.ReadAt != nil {
		t.Error("notification should be unread initially")
	}

	// Mark as read
	if err := repo.MarkAsRead(notification.ID); err != nil {
		t.Fatalf("MarkAsRead() error = %v", err)
	}

	// Verify it's read
	n, err = repo.GetByID(notification.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if n.ReadAt == nil {
		t.Error("notification should be read after MarkAsRead()")
	}
}

func TestNotificationRepository_MarkAsUnread(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create user
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "markunread@example.com")

	repo := NewNotificationRepository(db)

	// Create notification
	notification := &domain.Notification{
		UserID:    user.ID,
		Type:      domain.NotificationTypePRAchievement,
		Title:     "To Mark Unread",
		Message:   "Test message",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repo.Create(notification); err != nil {
		t.Fatalf("failed to create notification: %v", err)
	}

	// Mark as read first
	if err := repo.MarkAsRead(notification.ID); err != nil {
		t.Fatalf("MarkAsRead() error = %v", err)
	}

	// Verify it's read
	n, err := repo.GetByID(notification.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if n.ReadAt == nil {
		t.Error("notification should be read")
	}

	// Mark as unread
	if err := repo.MarkAsUnread(notification.ID); err != nil {
		t.Fatalf("MarkAsUnread() error = %v", err)
	}

	// Verify it's unread
	n, err = repo.GetByID(notification.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if n.ReadAt != nil {
		t.Error("notification should be unread after MarkAsUnread()")
	}
}

func TestNotificationRepository_MarkAllAsReadForUser(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create users
	userRepo := NewSQLiteUserRepository(db)
	user1 := createTestUser(t, userRepo, "markall1@example.com")
	user2 := createTestUser(t, userRepo, "markall2@example.com")

	repo := NewNotificationRepository(db)

	// Create notifications for user1
	for i := 0; i < 5; i++ {
		n := &domain.Notification{
			UserID:    user1.ID,
			Type:      domain.NotificationTypePRAchievement,
			Title:     "User1 Notification " + string(rune('A'+i)),
			Message:   "Test message",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := repo.Create(n); err != nil {
			t.Fatalf("failed to create notification: %v", err)
		}
	}

	// Create notifications for user2
	for i := 0; i < 3; i++ {
		n := &domain.Notification{
			UserID:    user2.ID,
			Type:      domain.NotificationTypeWeeklyStreak,
			Title:     "User2 Notification " + string(rune('A'+i)),
			Message:   "Test message",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := repo.Create(n); err != nil {
			t.Fatalf("failed to create notification: %v", err)
		}
	}

	// Mark all as read for user1
	if err := repo.MarkAllAsReadForUser(user1.ID); err != nil {
		t.Fatalf("MarkAllAsReadForUser() error = %v", err)
	}

	// Verify user1 has no unread
	count1, err := repo.CountUnreadByUserID(user1.ID)
	if err != nil {
		t.Fatalf("CountUnreadByUserID() error = %v", err)
	}
	if count1 != 0 {
		t.Errorf("user1 should have 0 unread, got %d", count1)
	}

	// Verify user2 still has unread (should not be affected)
	count2, err := repo.CountUnreadByUserID(user2.ID)
	if err != nil {
		t.Fatalf("CountUnreadByUserID() error = %v", err)
	}
	if count2 != 3 {
		t.Errorf("user2 should have 3 unread, got %d", count2)
	}
}

func TestNotificationRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create user
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "deletenotify@example.com")

	repo := NewNotificationRepository(db)

	// Create notification
	notification := &domain.Notification{
		UserID:    user.ID,
		Type:      domain.NotificationTypePRAchievement,
		Title:     "To Delete",
		Message:   "This will be deleted",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repo.Create(notification); err != nil {
		t.Fatalf("failed to create notification: %v", err)
	}

	// Verify it exists
	n, err := repo.GetByID(notification.ID)
	if err != nil || n == nil {
		t.Fatal("notification should exist before delete")
	}

	// Delete
	if err := repo.Delete(notification.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's deleted
	deleted, err := repo.GetByID(notification.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if deleted != nil {
		t.Error("notification should be deleted")
	}
}

func TestNotificationRepository_Delete_NonExistent(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewNotificationRepository(db)

	// Try to delete non-existent notification
	err = repo.Delete(99999)
	if err == nil {
		t.Error("Delete() should fail for non-existent notification")
	}
}

func TestNotificationRepository_DeleteOlderThan(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create user
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "deleteold@example.com")

	repo := NewNotificationRepository(db)

	// Create old notifications
	oldTime := time.Now().Add(-48 * time.Hour)
	for i := 0; i < 3; i++ {
		n := &domain.Notification{
			UserID:    user.ID,
			Type:      domain.NotificationTypePRAchievement,
			Title:     "Old Notification " + string(rune('A'+i)),
			Message:   "This is old",
			CreatedAt: oldTime,
			UpdatedAt: oldTime,
		}
		if err := repo.Create(n); err != nil {
			t.Fatalf("failed to create notification: %v", err)
		}
	}

	// Create new notifications
	newTime := time.Now()
	for i := 0; i < 2; i++ {
		n := &domain.Notification{
			UserID:    user.ID,
			Type:      domain.NotificationTypeWeeklyStreak,
			Title:     "New Notification " + string(rune('A'+i)),
			Message:   "This is new",
			CreatedAt: newTime,
			UpdatedAt: newTime,
		}
		if err := repo.Create(n); err != nil {
			t.Fatalf("failed to create notification: %v", err)
		}
	}

	// Delete notifications older than 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	deleted, err := repo.DeleteOlderThan(cutoff)
	if err != nil {
		t.Fatalf("DeleteOlderThan() error = %v", err)
	}

	if deleted != 3 {
		t.Errorf("DeleteOlderThan() deleted %d, want 3", deleted)
	}

	// Verify remaining notifications
	remaining, err := repo.GetByUserID(user.ID, 10, 0)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}

	if len(remaining) != 2 {
		t.Errorf("remaining notifications = %d, want 2", len(remaining))
	}
}
