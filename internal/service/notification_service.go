package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/email"
)

// NotificationService handles notification-related business logic
type NotificationService struct {
	notificationRepo domain.NotificationRepository
	orgRepo          domain.OrganizationRepository
	userRepo         domain.UserRepository
	settingsRepo     domain.UserSettingsRepository
	emailService     email.EmailService
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	notificationRepo domain.NotificationRepository,
	orgRepo domain.OrganizationRepository,
	userRepo domain.UserRepository,
	settingsRepo domain.UserSettingsRepository,
	emailService email.EmailService,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		orgRepo:          orgRepo,
		userRepo:         userRepo,
		settingsRepo:     settingsRepo,
		emailService:     emailService,
	}
}

// isNilEmailService checks if the email service interface contains a nil pointer
// This is necessary because in Go, an interface can be non-nil but contain a nil concrete value
func (s *NotificationService) isNilEmailService() bool {
	if s.emailService == nil {
		return true
	}
	// Type assert to check if the concrete value is nil
	if svc, ok := s.emailService.(*email.Service); ok {
		return svc == nil
	}
	return false
}

// CreateNotification creates a notification for all users in an organization
// based on their preferences. If no organization is specified, the notification
// is created only for the achieving user (solo user case).
func (s *NotificationService) CreateNotification(
	achievingUserID int64,
	organizationID *int64,
	notificationType string,
	title string,
	message string,
	data *domain.NotificationData,
) error {
	// If no organization specified, create notification only for the achieving user
	if organizationID == nil {
		return s.createNotificationForUser(achievingUserID, achievingUserID, organizationID, notificationType, title, message, data)
	}

	// Get all users in the organization
	orgUsers, err := s.orgRepo.GetOrganizationUsers(*organizationID)
	if err != nil {
		return fmt.Errorf("failed to get organization users: %w", err)
	}

	// Create notifications for each user based on their preferences
	for _, user := range orgUsers {
		if err := s.createNotificationForUser(achievingUserID, user.ID, organizationID, notificationType, title, message, data); err != nil {
			fmt.Printf("Failed to create notification for user %d: %v\n", user.ID, err)
		}
	}

	return nil
}

// createNotificationForUser creates a single notification for one user, respecting their preferences.
// achievingUserID is used to determine whether this recipient is the one who achieved the PR (vs. a gym mate).
func (s *NotificationService) createNotificationForUser(
	achievingUserID int64,
	recipientUserID int64,
	organizationID *int64,
	notificationType string,
	title string,
	message string,
	data *domain.NotificationData,
) error {
	var settings *domain.UserSettings
	if s.settingsRepo != nil {
		if s, err := s.settingsRepo.GetByUserID(recipientUserID); err == nil {
			settings = s
		}
	}

	if !s.shouldNotifyUser(settings, notificationType, recipientUserID == achievingUserID) {
		return nil
	}

	var dataJSON string
	if data != nil {
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal notification data: %w", err)
		}
		dataJSON = string(dataBytes)
	}

	now := time.Now()
	notification := &domain.Notification{
		UserID:         recipientUserID,
		OrganizationID: organizationID,
		Type:           notificationType,
		Title:          title,
		Message:        message,
		Data:           dataJSON,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return s.notificationRepo.Create(notification)
}

// shouldNotifyUser checks if a user should receive a notification based on their preferences
func (s *NotificationService) shouldNotifyUser(settings *domain.UserSettings, notificationType string, isAchievingUser bool) bool {
	if settings == nil || settings.NotificationPreferences == "" {
		// Default to all notifications enabled
		return true
	}

	var prefs map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(settings.NotificationPreferences), &prefs); err != nil {
		// If preferences can't be parsed, default to enabled
		return true
	}

	// Determine preference key based on notification type and whether it's the user's own achievement
	var prefKey string
	if isAchievingUser {
		// Own achievements
		switch notificationType {
		case domain.NotificationTypePRAchievement:
			prefKey = "pr_achievements"
		case domain.NotificationTypeWeeklyStreak:
			prefKey = "weekly_streak"
		case domain.NotificationTypeWODMilestone:
			prefKey = "wod_milestones"
		case domain.NotificationTypeConsistencyMilestone:
			prefKey = "consistency_milestones"
		case domain.NotificationTypeConsistencyInactivity:
			prefKey = "consistency_inactivity"
		default:
			return true
		}
	} else {
		// Gym mate achievements
		switch notificationType {
		case domain.NotificationTypePRAchievement:
			prefKey = "gym_mate_prs"
		case domain.NotificationTypeWeeklyStreak:
			prefKey = "gym_mate_streaks"
		case domain.NotificationTypeWODMilestone:
			prefKey = "gym_mate_milestones"
		case domain.NotificationTypeConsistencyMilestone:
			prefKey = "gym_mate_milestones"
		case domain.NotificationTypeConsistencyInactivity:
			return false // Don't notify others about inactivity
		default:
			return true
		}
	}

	// Check if in-app notifications are enabled for this type
	if pref, ok := prefs[prefKey]; ok {
		if enabled, ok := pref["enabled"].(bool); ok {
			if !enabled {
				return false
			}
			// Check specifically for in_app preference
			if inApp, ok := pref["in_app"].(bool); ok {
				return inApp
			}
			// Default to true if in_app not specified
			return true
		}
	}

	// Default to enabled if preference not found
	return true
}

// shouldEmailUser checks if a user should receive email notifications
func (s *NotificationService) shouldEmailUser(settings *domain.UserSettings, notificationType string, isAchievingUser bool) bool {
	if settings == nil || settings.NotificationPreferences == "" {
		// Default to email enabled for own achievements, disabled for gym mates
		return isAchievingUser
	}

	var prefs map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(settings.NotificationPreferences), &prefs); err != nil {
		// If preferences can't be parsed, default based on whether it's own achievement
		return isAchievingUser
	}

	// Determine preference key
	var prefKey string
	if isAchievingUser {
		switch notificationType {
		case domain.NotificationTypePRAchievement:
			prefKey = "pr_achievements"
		case domain.NotificationTypeWeeklyStreak:
			prefKey = "weekly_streak"
		case domain.NotificationTypeWODMilestone:
			prefKey = "wod_milestones"
		default:
			return isAchievingUser
		}
	} else {
		switch notificationType {
		case domain.NotificationTypePRAchievement:
			prefKey = "gym_mate_prs"
		case domain.NotificationTypeWeeklyStreak:
			prefKey = "gym_mate_streaks"
		case domain.NotificationTypeWODMilestone:
			prefKey = "gym_mate_milestones"
		default:
			return false
		}
	}

	// Check if email notifications are enabled for this type
	if pref, ok := prefs[prefKey]; ok {
		if email, ok := pref["email"].(bool); ok {
			return email
		}
	}

	// Default based on whether it's own achievement
	return isAchievingUser
}

// sendNotificationEmail sends a notification email to a user
func (s *NotificationService) sendNotificationEmail(to, title, message string) error {
	if s.emailService == nil {
		return nil // Email service not configured, skip silently
	}

	subject := fmt.Sprintf("ActaLog - %s", title)
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<div style="background-color: #00bcd4; color: white; padding: 20px; text-align: center;">
				<h1 style="margin: 0;">ActaLog</h1>
			</div>
			<div style="padding: 30px; background-color: white;">
				<h2 style="color: #2c3e50;">%s</h2>
				<p style="font-size: 16px; line-height: 1.6; color: #555;">%s</p>
				<div style="margin-top: 30px; text-align: center;">
					<a href="https://actalog.app/notifications"
					   style="background-color: #ffc107; color: #2c3e50; padding: 12px 30px; text-decoration: none; border-radius: 5px; font-weight: bold;">
						View in ActaLog
					</a>
				</div>
			</div>
			<div style="padding: 20px; background-color: #f5f7fa; text-align: center; font-size: 12px; color: #777;">
				<p>Manage your notification preferences in <a href="https://actalog.app/settings" style="color: #00bcd4;">settings</a>.</p>
			</div>
		</div>
	`, title, message)

	return s.emailService.SendHTMLEmail(to, subject, body)
}

// NotifyPRAchievement creates a notification for a personal record achievement
func (s *NotificationService) NotifyPRAchievement(
	userID int64,
	userName string,
	movementName *string,
	wodName *string,
	value string,
	organizationID *int64,
) error {
	var title, message string

	if movementName != nil {
		title = fmt.Sprintf("%s achieved a PR in %s!", userName, *movementName)
		message = fmt.Sprintf("%s set a new personal record in %s with %s", userName, *movementName, value)
	} else if wodName != nil {
		title = fmt.Sprintf("%s achieved a PR in %s!", userName, *wodName)
		message = fmt.Sprintf("%s set a new personal record in %s with %s", userName, *wodName, value)
	} else {
		title = fmt.Sprintf("%s achieved a PR!", userName)
		message = fmt.Sprintf("%s set a new personal record with %s", userName, value)
	}

	data := &domain.NotificationData{
		AchievingUserID:   userID,
		AchievingUserName: userName,
		MovementName:      movementName,
		WODName:           wodName,
		Value:             &value,
	}

	return s.CreateNotification(userID, organizationID, domain.NotificationTypePRAchievement, title, message, data)
}

// NotifyWeeklyStreak creates a notification for a weekly workout streak
func (s *NotificationService) NotifyWeeklyStreak(
	userID int64,
	userName string,
	count int,
	organizationID *int64,
) error {
	title := fmt.Sprintf("%s hit a %d-workout weekly streak!", userName, count)
	message := fmt.Sprintf("%s completed %d workouts in the last 7 days. Keep it up!", userName, count)

	data := &domain.NotificationData{
		AchievingUserID:   userID,
		AchievingUserName: userName,
		Count:             &count,
	}

	return s.CreateNotification(userID, organizationID, domain.NotificationTypeWeeklyStreak, title, message, data)
}

// NotifyWODMilestone creates a notification for a WOD milestone achievement
func (s *NotificationService) NotifyWODMilestone(
	userID int64,
	userName string,
	wodType string,
	count int,
	organizationID *int64,
) error {
	title := fmt.Sprintf("%s completed %d %s WODs!", userName, count, wodType)
	message := fmt.Sprintf("%s reached a milestone by completing %d %s WODs. Congratulations!", userName, count, wodType)

	data := &domain.NotificationData{
		AchievingUserID:   userID,
		AchievingUserName: userName,
		WODType:           &wodType,
		Count:             &count,
	}

	return s.CreateNotification(userID, organizationID, domain.NotificationTypeWODMilestone, title, message, data)
}

// NotifyUserWelcome sends a welcome notification and email to a new user
func (s *NotificationService) NotifyUserWelcome(userID int64, userName, userEmail, appURL string) error {
	title := "Welcome to ActaLog! 👋"
	message := "Thanks for joining! Log your first workout by tapping Quick Log on the dashboard. Track your progress in the Performance tab and explore benchmark WODs. Need help? Contact your administrator."

	now := time.Now()

	// Create in-app notification directly for the user (not organization-based)
	notification := &domain.Notification{
		UserID:         userID,
		OrganizationID: nil, // Welcome notifications are user-specific, not org-specific
		Type:           domain.NotificationTypeWelcome,
		Title:          title,
		Message:        message,
		Data:           "{}",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.notificationRepo.Create(notification); err != nil {
		fmt.Printf("Failed to create welcome notification for user %d: %v\n", userID, err)
		// Continue to send email even if notification creation fails
	}

	// Send welcome email if email service is available and properly initialized
	if s.emailService != nil && !s.isNilEmailService() {
		if err := s.emailService.SendWelcomeEmail(userEmail, userName, appURL); err != nil {
			fmt.Printf("Failed to send welcome email to %s: %v\n", userEmail, err)
			// Don't return error - email failure shouldn't break registration
		}
	}

	return nil
}

// GetNotifications retrieves notifications for a user with pagination
func (s *NotificationService) GetNotifications(userID int64, limit, offset int) ([]*domain.Notification, error) {
	return s.notificationRepo.GetByUserID(userID, limit, offset)
}

// GetUnreadNotifications retrieves unread notifications for a user with pagination
func (s *NotificationService) GetUnreadNotifications(userID int64, limit, offset int) ([]*domain.Notification, error) {
	return s.notificationRepo.GetUnreadByUserID(userID, limit, offset)
}

// GetUnreadCount gets the count of unread notifications for a user
func (s *NotificationService) GetUnreadCount(userID int64) (int64, error) {
	return s.notificationRepo.CountUnreadByUserID(userID)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID int64) error {
	return s.notificationRepo.MarkAsRead(notificationID)
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(userID int64) error {
	return s.notificationRepo.MarkAllAsReadForUser(userID)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(notificationID int64) error {
	return s.notificationRepo.Delete(notificationID)
}

// CleanupOldNotifications deletes notifications older than 90 days
func (s *NotificationService) CleanupOldNotifications() (int64, error) {
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	return s.notificationRepo.DeleteOlderThan(ninetyDaysAgo)
}

// CreateAnnouncement creates an announcement notification for selected users
// targetType can be "all", "organization", or "users"
// targetIDs is a list of organization IDs or user IDs depending on targetType
func (s *NotificationService) CreateAnnouncement(title, message, targetType string, targetIDs []int64, organizationID *int64) error {
	var targetUsers []*domain.User
	var err error

	switch targetType {
	case "all":
		// Get all users (using large limit as workaround for no ListAll method)
		targetUsers, err = s.userRepo.List(10000, 0)
		if err != nil {
			return fmt.Errorf("failed to get all users: %w", err)
		}

	case "organization":
		// Get all users in specified organizations
		for _, orgID := range targetIDs {
			orgUsers, err := s.orgRepo.GetOrganizationUsers(orgID)
			if err != nil {
				return fmt.Errorf("failed to get users for organization %d: %w", orgID, err)
			}
			targetUsers = append(targetUsers, orgUsers...)
		}

	case "users":
		// Get specific users by ID
		for _, userID := range targetIDs {
			user, err := s.userRepo.GetByID(userID)
			if err != nil {
				continue // Skip users that don't exist
			}
			if user != nil {
				targetUsers = append(targetUsers, user)
			}
		}

	default:
		return fmt.Errorf("invalid target type: %s", targetType)
	}

	// Create notification for each target user
	for _, user := range targetUsers {
		notification := &domain.Notification{
			UserID:         user.ID,
			OrganizationID: organizationID,
			Type:           "announcement",
			Title:          title,
			Message:        message,
			Data:           "{}", // Empty JSON object to satisfy CHECK constraint
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := s.notificationRepo.Create(notification); err != nil {
			// Log error but continue with other users
			fmt.Printf("Failed to create announcement for user %d: %v\n", user.ID, err)
		}
	}

	return nil
}
