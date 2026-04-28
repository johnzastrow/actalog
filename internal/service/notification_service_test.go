package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// =============================================================================
// Mock Notification Repository
// =============================================================================

type testNotificationRepo struct {
	notifications map[int64]*domain.Notification
	nextID        int64
	createError   error
}

func newTestNotificationRepo() *testNotificationRepo {
	return &testNotificationRepo{
		notifications: make(map[int64]*domain.Notification),
		nextID:        1,
	}
}

func (m *testNotificationRepo) Create(notification *domain.Notification) error {
	if m.createError != nil {
		return m.createError
	}
	notification.ID = m.nextID
	m.nextID++
	m.notifications[notification.ID] = notification
	return nil
}

func (m *testNotificationRepo) GetByID(id int64) (*domain.Notification, error) {
	notification, ok := m.notifications[id]
	if !ok {
		return nil, nil
	}
	return notification, nil
}

func (m *testNotificationRepo) GetByUserID(userID int64, limit, offset int) ([]*domain.Notification, error) {
	var result []*domain.Notification
	for _, n := range m.notifications {
		if n.UserID == userID {
			result = append(result, n)
		}
	}
	return result, nil
}

func (m *testNotificationRepo) GetUnreadByUserID(userID int64, limit, offset int) ([]*domain.Notification, error) {
	var result []*domain.Notification
	for _, n := range m.notifications {
		if n.UserID == userID && n.ReadAt == nil {
			result = append(result, n)
		}
	}
	return result, nil
}

func (m *testNotificationRepo) CountUnreadByUserID(userID int64) (int64, error) {
	count := int64(0)
	for _, n := range m.notifications {
		if n.UserID == userID && n.ReadAt == nil {
			count++
		}
	}
	return count, nil
}

func (m *testNotificationRepo) MarkAsRead(id int64) error {
	if n, ok := m.notifications[id]; ok {
		now := time.Now()
		n.ReadAt = &now
	}
	return nil
}

func (m *testNotificationRepo) MarkAsUnread(id int64) error {
	if n, ok := m.notifications[id]; ok {
		n.ReadAt = nil
	}
	return nil
}

func (m *testNotificationRepo) MarkAllAsReadForUser(userID int64) error {
	now := time.Now()
	for _, n := range m.notifications {
		if n.UserID == userID {
			n.ReadAt = &now
		}
	}
	return nil
}

func (m *testNotificationRepo) Delete(id int64) error {
	delete(m.notifications, id)
	return nil
}

func (m *testNotificationRepo) DeleteOlderThan(cutoff time.Time) (int64, error) {
	count := int64(0)
	for id, n := range m.notifications {
		if n.CreatedAt.Before(cutoff) {
			delete(m.notifications, id)
			count++
		}
	}
	return count, nil
}

// =============================================================================
// Mock User Settings Repository
// =============================================================================

type testUserSettingsRepo struct {
	settings map[int64]*domain.UserSettings
}

func newTestUserSettingsRepo() *testUserSettingsRepo {
	return &testUserSettingsRepo{
		settings: make(map[int64]*domain.UserSettings),
	}
}

func (m *testUserSettingsRepo) GetByUserID(userID int64) (*domain.UserSettings, error) {
	settings, ok := m.settings[userID]
	if !ok {
		return nil, nil
	}
	return settings, nil
}

func (m *testUserSettingsRepo) Create(settings *domain.UserSettings) error {
	m.settings[settings.UserID] = settings
	return nil
}

func (m *testUserSettingsRepo) Update(settings *domain.UserSettings) error {
	m.settings[settings.UserID] = settings
	return nil
}

func (m *testUserSettingsRepo) Delete(userID int64) error {
	delete(m.settings, userID)
	return nil
}

// =============================================================================
// Extended Mock Organization Repository for Notification Tests
// =============================================================================

type testNotificationOrgRepo struct {
	*mockOrganizationRepo
	orgUsers map[int64][]*domain.User
}

func newTestNotificationOrgRepo() *testNotificationOrgRepo {
	return &testNotificationOrgRepo{
		mockOrganizationRepo: newMockOrganizationRepo(),
		orgUsers:             make(map[int64][]*domain.User),
	}
}

func (m *testNotificationOrgRepo) GetOrganizationUsers(orgID int64) ([]*domain.User, error) {
	users, ok := m.orgUsers[orgID]
	if !ok {
		return []*domain.User{}, nil
	}
	return users, nil
}

// =============================================================================
// Notification Service Tests
// =============================================================================

func TestNotificationService_GetNotifications(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	service := NewNotificationService(notifRepo, nil, nil, nil, nil)

	userID := int64(1)

	// Create some notifications
	for i := 0; i < 5; i++ {
		notifRepo.Create(&domain.Notification{
			UserID:    userID,
			Type:      domain.NotificationTypePRAchievement,
			Title:     "Test",
			Message:   "Test message",
			CreatedAt: time.Now(),
		})
	}

	// Create notifications for another user
	for i := 0; i < 3; i++ {
		notifRepo.Create(&domain.Notification{
			UserID:    2,
			Type:      domain.NotificationTypePRAchievement,
			Title:     "Test",
			Message:   "Test message",
			CreatedAt: time.Now(),
		})
	}

	notifications, err := service.GetNotifications(userID, 10, 0)
	if err != nil {
		t.Errorf("GetNotifications() error = %v", err)
		return
	}

	if len(notifications) != 5 {
		t.Errorf("GetNotifications() returned %d notifications, want 5", len(notifications))
	}
}

func TestNotificationService_GetUnreadNotifications(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	service := NewNotificationService(notifRepo, nil, nil, nil, nil)

	userID := int64(1)

	// Create some notifications
	for i := 0; i < 5; i++ {
		notifRepo.Create(&domain.Notification{
			UserID:    userID,
			Type:      domain.NotificationTypePRAchievement,
			Title:     "Test",
			Message:   "Test message",
			CreatedAt: time.Now(),
		})
	}

	// Mark some as read
	notifRepo.MarkAsRead(1)
	notifRepo.MarkAsRead(2)

	unread, err := service.GetUnreadNotifications(userID, 10, 0)
	if err != nil {
		t.Errorf("GetUnreadNotifications() error = %v", err)
		return
	}

	if len(unread) != 3 {
		t.Errorf("GetUnreadNotifications() returned %d, want 3", len(unread))
	}
}

func TestNotificationService_GetUnreadCount(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	service := NewNotificationService(notifRepo, nil, nil, nil, nil)

	userID := int64(1)

	// Create some notifications
	for i := 0; i < 5; i++ {
		notifRepo.Create(&domain.Notification{
			UserID:    userID,
			Type:      domain.NotificationTypePRAchievement,
			Title:     "Test",
			Message:   "Test message",
			CreatedAt: time.Now(),
		})
	}

	// Mark some as read
	notifRepo.MarkAsRead(1)
	notifRepo.MarkAsRead(2)

	count, err := service.GetUnreadCount(userID)
	if err != nil {
		t.Errorf("GetUnreadCount() error = %v", err)
		return
	}

	if count != 3 {
		t.Errorf("GetUnreadCount() = %d, want 3", count)
	}
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	service := NewNotificationService(notifRepo, nil, nil, nil, nil)

	// Create a notification
	notifRepo.Create(&domain.Notification{
		UserID:    1,
		Type:      domain.NotificationTypePRAchievement,
		Title:     "Test",
		Message:   "Test message",
		CreatedAt: time.Now(),
	})

	err := service.MarkAsRead(1)
	if err != nil {
		t.Errorf("MarkAsRead() error = %v", err)
		return
	}

	// Verify it's marked as read
	notification, _ := notifRepo.GetByID(1)
	if notification.ReadAt == nil {
		t.Error("MarkAsRead() did not set ReadAt")
	}
}

func TestNotificationService_MarkAllAsRead(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	service := NewNotificationService(notifRepo, nil, nil, nil, nil)

	userID := int64(1)

	// Create some notifications
	for i := 0; i < 5; i++ {
		notifRepo.Create(&domain.Notification{
			UserID:    userID,
			Type:      domain.NotificationTypePRAchievement,
			Title:     "Test",
			Message:   "Test message",
			CreatedAt: time.Now(),
		})
	}

	err := service.MarkAllAsRead(userID)
	if err != nil {
		t.Errorf("MarkAllAsRead() error = %v", err)
		return
	}

	// Verify all are marked as read
	unread, _ := service.GetUnreadNotifications(userID, 10, 0)
	if len(unread) != 0 {
		t.Errorf("MarkAllAsRead() left %d unread notifications", len(unread))
	}
}

func TestNotificationService_DeleteNotification(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	service := NewNotificationService(notifRepo, nil, nil, nil, nil)

	// Create a notification
	notifRepo.Create(&domain.Notification{
		UserID:    1,
		Type:      domain.NotificationTypePRAchievement,
		Title:     "Test",
		Message:   "Test message",
		CreatedAt: time.Now(),
	})

	err := service.DeleteNotification(1)
	if err != nil {
		t.Errorf("DeleteNotification() error = %v", err)
		return
	}

	// Verify it's deleted
	notification, _ := notifRepo.GetByID(1)
	if notification != nil {
		t.Error("DeleteNotification() did not delete the notification")
	}
}

func TestNotificationService_CleanupOldNotifications(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	service := NewNotificationService(notifRepo, nil, nil, nil, nil)

	// Create some old notifications
	oldTime := time.Now().AddDate(0, 0, -100) // 100 days ago
	for i := 0; i < 3; i++ {
		notif := &domain.Notification{
			UserID:    1,
			Type:      domain.NotificationTypePRAchievement,
			Title:     "Old",
			Message:   "Old message",
			CreatedAt: oldTime,
		}
		notifRepo.Create(notif)
		// Override the created_at for testing
		notifRepo.notifications[notif.ID].CreatedAt = oldTime
	}

	// Create some recent notifications
	for i := 0; i < 2; i++ {
		notifRepo.Create(&domain.Notification{
			UserID:    1,
			Type:      domain.NotificationTypePRAchievement,
			Title:     "Recent",
			Message:   "Recent message",
			CreatedAt: time.Now(),
		})
	}

	deleted, err := service.CleanupOldNotifications()
	if err != nil {
		t.Errorf("CleanupOldNotifications() error = %v", err)
		return
	}

	if deleted != 3 {
		t.Errorf("CleanupOldNotifications() deleted %d, want 3", deleted)
	}

	// Verify only recent notifications remain
	all, _ := service.GetNotifications(1, 10, 0)
	if len(all) != 2 {
		t.Errorf("CleanupOldNotifications() left %d notifications, want 2", len(all))
	}
}

func TestNotificationService_CreateNotification_NoOrganization(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	service := NewNotificationService(notifRepo, nil, nil, nil, nil)

	// Creating notification with nil organization should still notify the achieving user
	err := service.CreateNotification(1, nil, domain.NotificationTypePRAchievement, "Test", "Message", nil)
	if err != nil {
		t.Errorf("CreateNotification() with nil org error = %v, want nil", err)
	}

	// Notification should be created for the achieving user (solo user case)
	notifications, _ := notifRepo.GetByUserID(1, 10, 0)
	if len(notifications) != 1 {
		t.Errorf("CreateNotification() with nil org created %d notifications, want 1", len(notifications))
	}
}

func TestNotificationService_NotifyPRAchievement(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	orgRepo := newTestNotificationOrgRepo()
	userRepo := newMockUserRepo()
	settingsRepo := newTestUserSettingsRepo()

	service := NewNotificationService(notifRepo, orgRepo, userRepo, settingsRepo, nil)

	// Create users - note that mockUserRepo increments ID after setting
	user1 := &domain.User{Email: "user1@example.com", Name: "User One"}
	user2 := &domain.User{Email: "user2@example.com", Name: "User Two"}
	userRepo.Create(user1) // Gets ID 2
	userRepo.Create(user2) // Gets ID 3

	orgID := int64(1)
	orgRepo.orgUsers[orgID] = []*domain.User{user1, user2}

	movementName := "Back Squat"
	err := service.NotifyPRAchievement(user1.ID, "User One", &movementName, nil, "315 lbs", &orgID)
	if err != nil {
		t.Errorf("NotifyPRAchievement() error = %v", err)
		return
	}

	// Both users should receive notifications
	notif1, _ := notifRepo.GetByUserID(user1.ID, 10, 0)
	notif2, _ := notifRepo.GetByUserID(user2.ID, 10, 0)

	if len(notif1) == 0 {
		t.Error("NotifyPRAchievement() did not create notification for achieving user")
	}
	if len(notif2) == 0 {
		t.Error("NotifyPRAchievement() did not create notification for gym mate")
	}
}

func TestNotificationService_NotifyPRAchievement_WithWODName(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	orgRepo := newTestNotificationOrgRepo()
	userRepo := newMockUserRepo()
	settingsRepo := newTestUserSettingsRepo()

	service := NewNotificationService(notifRepo, orgRepo, userRepo, settingsRepo, nil)

	user := &domain.User{Email: "user@example.com", Name: "Test User"}
	userRepo.Create(user)

	orgID := int64(1)
	orgRepo.orgUsers[orgID] = []*domain.User{user}

	wodName := "Fran"
	err := service.NotifyPRAchievement(user.ID, "Test User", nil, &wodName, "3:45", &orgID)
	if err != nil {
		t.Errorf("NotifyPRAchievement() with WOD name error = %v", err)
		return
	}

	notifications, _ := notifRepo.GetByUserID(user.ID, 10, 0)
	if len(notifications) == 0 {
		t.Error("NotifyPRAchievement() with WOD name did not create notification")
	}
}

func TestNotificationService_NotifyPRAchievement_NoMovementOrWOD(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	orgRepo := newTestNotificationOrgRepo()
	userRepo := newMockUserRepo()
	settingsRepo := newTestUserSettingsRepo()

	service := NewNotificationService(notifRepo, orgRepo, userRepo, settingsRepo, nil)

	user := &domain.User{Email: "user@example.com", Name: "Test User"}
	userRepo.Create(user)

	orgID := int64(1)
	orgRepo.orgUsers[orgID] = []*domain.User{user}

	// Test with neither movement nor WOD name
	err := service.NotifyPRAchievement(user.ID, "Test User", nil, nil, "New PR!", &orgID)
	if err != nil {
		t.Errorf("NotifyPRAchievement() without movement/WOD error = %v", err)
		return
	}

	notifications, _ := notifRepo.GetByUserID(user.ID, 10, 0)
	if len(notifications) == 0 {
		t.Error("NotifyPRAchievement() without movement/WOD did not create notification")
	}
}

func TestNotificationService_NotifyWeeklyStreak(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	orgRepo := newTestNotificationOrgRepo()
	userRepo := newMockUserRepo()
	settingsRepo := newTestUserSettingsRepo()

	service := NewNotificationService(notifRepo, orgRepo, userRepo, settingsRepo, nil)

	// Create user in organization
	user := &domain.User{Email: "user@example.com", Name: "Test User"}
	userRepo.Create(user)

	orgID := int64(1)
	orgRepo.orgUsers[orgID] = []*domain.User{user}

	err := service.NotifyWeeklyStreak(user.ID, "Test User", 7, &orgID)
	if err != nil {
		t.Errorf("NotifyWeeklyStreak() error = %v", err)
		return
	}

	notifications, _ := notifRepo.GetByUserID(user.ID, 10, 0)
	if len(notifications) == 0 {
		t.Error("NotifyWeeklyStreak() did not create notification")
	}
}

func TestNotificationService_NotifyWODMilestone(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	orgRepo := newTestNotificationOrgRepo()
	userRepo := newMockUserRepo()
	settingsRepo := newTestUserSettingsRepo()

	service := NewNotificationService(notifRepo, orgRepo, userRepo, settingsRepo, nil)

	// Create user in organization
	user := &domain.User{Email: "user@example.com", Name: "Test User"}
	userRepo.Create(user)

	orgID := int64(1)
	orgRepo.orgUsers[orgID] = []*domain.User{user}

	err := service.NotifyWODMilestone(user.ID, "Test User", "hero", 10, &orgID)
	if err != nil {
		t.Errorf("NotifyWODMilestone() error = %v", err)
		return
	}

	notifications, _ := notifRepo.GetByUserID(user.ID, 10, 0)
	if len(notifications) == 0 {
		t.Error("NotifyWODMilestone() did not create notification")
	}
}

func TestNotificationService_CreateAnnouncement(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	orgRepo := newTestNotificationOrgRepo()
	userRepo := newMockUserRepo()

	service := NewNotificationService(notifRepo, orgRepo, userRepo, nil, nil)

	// Create some users - store their IDs
	var userIDs []int64
	for i := 0; i < 5; i++ {
		user := &domain.User{
			Email: "user@example.com",
			Name:  "User",
		}
		userRepo.Create(user)
		userIDs = append(userIDs, user.ID)
	}

	tests := []struct {
		name       string
		title      string
		message    string
		targetType string
		targetIDs  []int64
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "announce to specific users",
			title:      "Test Announcement",
			message:    "This is a test",
			targetType: "users",
			targetIDs:  userIDs[:3], // First 3 users
			wantCount:  3,
			wantErr:    false,
		},
		{
			name:       "invalid target type",
			title:      "Test",
			message:    "Test",
			targetType: "invalid",
			targetIDs:  userIDs[:1],
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear notifications
			notifRepo.notifications = make(map[int64]*domain.Notification)
			notifRepo.nextID = 1

			err := service.CreateAnnouncement(tt.title, tt.message, tt.targetType, tt.targetIDs, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAnnouncement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				count := len(notifRepo.notifications)
				if count != tt.wantCount {
					t.Errorf("CreateAnnouncement() created %d notifications, want %d", count, tt.wantCount)
				}
			}
		})
	}
}

func TestNotificationService_ShouldNotifyUser(t *testing.T) {
	service := NewNotificationService(nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		settings         *domain.UserSettings
		notificationType string
		isAchievingUser  bool
		want             bool
	}{
		{
			name:             "nil settings - should default to true",
			settings:         nil,
			notificationType: domain.NotificationTypePRAchievement,
			isAchievingUser:  true,
			want:             true,
		},
		{
			name:             "empty preferences - should default to true",
			settings:         &domain.UserSettings{NotificationPreferences: ""},
			notificationType: domain.NotificationTypePRAchievement,
			isAchievingUser:  true,
			want:             true,
		},
		{
			name:             "invalid JSON - should default to true",
			settings:         &domain.UserSettings{NotificationPreferences: "not valid json"},
			notificationType: domain.NotificationTypePRAchievement,
			isAchievingUser:  true,
			want:             true,
		},
		{
			name: "pr_achievements disabled",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"pr_achievements":{"enabled":false}}`,
			},
			notificationType: domain.NotificationTypePRAchievement,
			isAchievingUser:  true,
			want:             false,
		},
		{
			name: "pr_achievements enabled with in_app true",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"pr_achievements":{"enabled":true,"in_app":true}}`,
			},
			notificationType: domain.NotificationTypePRAchievement,
			isAchievingUser:  true,
			want:             true,
		},
		{
			name: "gym_mate_prs disabled",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"gym_mate_prs":{"enabled":false}}`,
			},
			notificationType: domain.NotificationTypePRAchievement,
			isAchievingUser:  false,
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.shouldNotifyUser(tt.settings, tt.notificationType, tt.isAchievingUser)
			if got != tt.want {
				t.Errorf("shouldNotifyUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotificationService_ShouldEmailUser(t *testing.T) {
	service := NewNotificationService(nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		settings         *domain.UserSettings
		notificationType string
		isAchievingUser  bool
		want             bool
	}{
		{
			name:             "nil settings - own achievement defaults to true",
			settings:         nil,
			notificationType: domain.NotificationTypePRAchievement,
			isAchievingUser:  true,
			want:             true,
		},
		{
			name:             "nil settings - gym mate defaults to false",
			settings:         nil,
			notificationType: domain.NotificationTypePRAchievement,
			isAchievingUser:  false,
			want:             false,
		},
		{
			name: "email explicitly enabled",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"pr_achievements":{"email":true}}`,
			},
			notificationType: domain.NotificationTypePRAchievement,
			isAchievingUser:  true,
			want:             true,
		},
		{
			name: "email explicitly disabled",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"pr_achievements":{"email":false}}`,
			},
			notificationType: domain.NotificationTypePRAchievement,
			isAchievingUser:  true,
			want:             false,
		},
		{
			name: "weekly_streak email enabled for own achievement",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"weekly_streak":{"email":true}}`,
			},
			notificationType: domain.NotificationTypeWeeklyStreak,
			isAchievingUser:  true,
			want:             true,
		},
		{
			name: "wod_milestones email disabled",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"wod_milestones":{"email":false}}`,
			},
			notificationType: domain.NotificationTypeWODMilestone,
			isAchievingUser:  true,
			want:             false,
		},
		{
			name: "gym_mate_streaks email enabled",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"gym_mate_streaks":{"email":true}}`,
			},
			notificationType: domain.NotificationTypeWeeklyStreak,
			isAchievingUser:  false,
			want:             true,
		},
		{
			name: "gym_mate_milestones email enabled",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"gym_mate_milestones":{"email":true}}`,
			},
			notificationType: domain.NotificationTypeWODMilestone,
			isAchievingUser:  false,
			want:             true,
		},
		{
			name:             "invalid JSON defaults based on isAchievingUser",
			settings:         &domain.UserSettings{NotificationPreferences: "not valid json"},
			notificationType: domain.NotificationTypePRAchievement,
			isAchievingUser:  true,
			want:             true,
		},
		{
			name:             "unknown type for own achievement",
			settings:         &domain.UserSettings{NotificationPreferences: `{}`},
			notificationType: "unknown_type",
			isAchievingUser:  true,
			want:             true,
		},
		{
			name:             "unknown type for gym mate",
			settings:         &domain.UserSettings{NotificationPreferences: `{}`},
			notificationType: "unknown_type",
			isAchievingUser:  false,
			want:             false,
		},
		{
			name: "preference exists but email not specified - defaults based on isAchievingUser",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"pr_achievements":{"enabled":true}}`,
			},
			notificationType: domain.NotificationTypePRAchievement,
			isAchievingUser:  true,
			want:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.shouldEmailUser(tt.settings, tt.notificationType, tt.isAchievingUser)
			if got != tt.want {
				t.Errorf("shouldEmailUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotificationService_ShouldNotifyUser_MoreCases(t *testing.T) {
	service := NewNotificationService(nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		settings         *domain.UserSettings
		notificationType string
		isAchievingUser  bool
		want             bool
	}{
		{
			name: "weekly_streak enabled",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"weekly_streak":{"enabled":true,"in_app":true}}`,
			},
			notificationType: domain.NotificationTypeWeeklyStreak,
			isAchievingUser:  true,
			want:             true,
		},
		{
			name: "wod_milestones disabled",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"wod_milestones":{"enabled":false}}`,
			},
			notificationType: domain.NotificationTypeWODMilestone,
			isAchievingUser:  true,
			want:             false,
		},
		{
			name: "gym_mate_streaks enabled",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"gym_mate_streaks":{"enabled":true,"in_app":true}}`,
			},
			notificationType: domain.NotificationTypeWeeklyStreak,
			isAchievingUser:  false,
			want:             true,
		},
		{
			name: "gym_mate_milestones disabled",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"gym_mate_milestones":{"enabled":false}}`,
			},
			notificationType: domain.NotificationTypeWODMilestone,
			isAchievingUser:  false,
			want:             false,
		},
		{
			name: "enabled true but in_app false",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"pr_achievements":{"enabled":true,"in_app":false}}`,
			},
			notificationType: domain.NotificationTypePRAchievement,
			isAchievingUser:  true,
			want:             false,
		},
		{
			name: "enabled true but in_app not specified",
			settings: &domain.UserSettings{
				NotificationPreferences: `{"pr_achievements":{"enabled":true}}`,
			},
			notificationType: domain.NotificationTypePRAchievement,
			isAchievingUser:  true,
			want:             true,
		},
		{
			name:             "unknown type for own user defaults to true",
			settings:         &domain.UserSettings{NotificationPreferences: `{}`},
			notificationType: "unknown_type",
			isAchievingUser:  true,
			want:             true,
		},
		{
			name:             "unknown type for gym mate defaults to true",
			settings:         &domain.UserSettings{NotificationPreferences: `{}`},
			notificationType: "unknown_type",
			isAchievingUser:  false,
			want:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.shouldNotifyUser(tt.settings, tt.notificationType, tt.isAchievingUser)
			if got != tt.want {
				t.Errorf("shouldNotifyUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotificationService_CreateNotification_WithData(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	orgRepo := newTestNotificationOrgRepo()
	userRepo := newMockUserRepo()
	settingsRepo := newTestUserSettingsRepo()

	service := NewNotificationService(notifRepo, orgRepo, userRepo, settingsRepo, nil)

	// Create a user
	user := &domain.User{Email: "user@example.com", Name: "Test User"}
	userRepo.Create(user)

	orgID := int64(1)
	orgRepo.orgUsers[orgID] = []*domain.User{user}

	movementName := "Back Squat"
	value := "315 lbs"
	data := &domain.NotificationData{
		MovementName: &movementName,
		Value:        &value,
	}

	err := service.CreateNotification(user.ID, &orgID, domain.NotificationTypePRAchievement, "Test", "Message", data)
	if err != nil {
		t.Errorf("CreateNotification() with data error = %v", err)
		return
	}

	// Verify notification was created with data
	notifications, _ := notifRepo.GetByUserID(user.ID, 10, 0)
	if len(notifications) != 1 {
		t.Errorf("CreateNotification() created %d notifications, want 1", len(notifications))
		return
	}

	if notifications[0].Data == "" {
		t.Error("CreateNotification() should have set data field")
	}
}

func TestNotificationService_CreateNotification_CreateError(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	notifRepo.createError = fmt.Errorf("database error")

	orgRepo := newTestNotificationOrgRepo()
	userRepo := newMockUserRepo()
	settingsRepo := newTestUserSettingsRepo()

	service := NewNotificationService(notifRepo, orgRepo, userRepo, settingsRepo, nil)

	// Create a user
	user := &domain.User{Email: "user@example.com", Name: "Test User"}
	userRepo.Create(user)

	orgID := int64(1)
	orgRepo.orgUsers[orgID] = []*domain.User{user}

	// Should not return error even if notification creation fails (just logs)
	err := service.CreateNotification(user.ID, &orgID, domain.NotificationTypePRAchievement, "Test", "Message", nil)
	if err != nil {
		t.Errorf("CreateNotification() should not return error on create failure, got: %v", err)
	}
}

func TestNotificationService_CreateNotification_UserSkipsByPreference(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	orgRepo := newTestNotificationOrgRepo()
	userRepo := newMockUserRepo()
	settingsRepo := newTestUserSettingsRepo()

	service := NewNotificationService(notifRepo, orgRepo, userRepo, settingsRepo, nil)

	// Create a user with disabled notifications
	user := &domain.User{Email: "user@example.com", Name: "Test User"}
	userRepo.Create(user)

	// Set user preferences to disable PR notifications
	settingsRepo.settings[user.ID] = &domain.UserSettings{
		UserID:                  user.ID,
		NotificationPreferences: `{"pr_achievements":{"enabled":false}}`,
	}

	orgID := int64(1)
	orgRepo.orgUsers[orgID] = []*domain.User{user}

	err := service.CreateNotification(user.ID, &orgID, domain.NotificationTypePRAchievement, "Test", "Message", nil)
	if err != nil {
		t.Errorf("CreateNotification() error = %v", err)
		return
	}

	// Verify no notifications were created (user opted out)
	notifications, _ := notifRepo.GetByUserID(user.ID, 10, 0)
	if len(notifications) != 0 {
		t.Errorf("CreateNotification() created %d notifications for opted-out user, want 0", len(notifications))
	}
}

func TestNotificationService_CreateAnnouncement_AllUsers(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	orgRepo := newTestNotificationOrgRepo()
	userRepo := newMockUserRepoWithList()

	service := NewNotificationService(notifRepo, orgRepo, userRepo, nil, nil)

	// Create some users
	for i := 0; i < 5; i++ {
		user := &domain.User{Email: "user@example.com", Name: "User"}
		userRepo.Create(user)
	}

	err := service.CreateAnnouncement("Test", "Message", "all", nil, nil)
	if err != nil {
		t.Errorf("CreateAnnouncement() for all users error = %v", err)
		return
	}

	// All users should receive notifications
	if len(notifRepo.notifications) != 5 {
		t.Errorf("CreateAnnouncement() created %d notifications, want 5", len(notifRepo.notifications))
	}
}

func TestNotificationService_CreateAnnouncement_ToOrganization(t *testing.T) {
	notifRepo := newTestNotificationRepo()
	orgRepo := newTestNotificationOrgRepo()
	userRepo := newMockUserRepo()

	service := NewNotificationService(notifRepo, orgRepo, userRepo, nil, nil)

	// Create users in organization
	user1 := &domain.User{Email: "user1@example.com", Name: "User One"}
	user2 := &domain.User{Email: "user2@example.com", Name: "User Two"}
	userRepo.Create(user1)
	userRepo.Create(user2)

	orgID := int64(1)
	orgRepo.orgUsers[orgID] = []*domain.User{user1, user2}

	err := service.CreateAnnouncement("Test", "Message", "organization", []int64{orgID}, nil)
	if err != nil {
		t.Errorf("CreateAnnouncement() to organization error = %v", err)
		return
	}

	// Both users in org should receive notifications
	if len(notifRepo.notifications) != 2 {
		t.Errorf("CreateAnnouncement() created %d notifications, want 2", len(notifRepo.notifications))
	}
}

// mockUserRepoWithList extends mockUserRepo with List functionality
type mockUserRepoWithList struct {
	*mockUserRepo
}

func newMockUserRepoWithList() *mockUserRepoWithList {
	return &mockUserRepoWithList{
		mockUserRepo: newMockUserRepo(),
	}
}

func (m *mockUserRepoWithList) List(limit, offset int) ([]*domain.User, error) {
	var users []*domain.User
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, nil
}
