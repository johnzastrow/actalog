package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/email"
)

// User event types for admin notifications
const (
	UserEventRegistration  = "registration"
	UserEventProfileUpdate = "profile_update"
	UserEventRoleChange    = "role_change"
	UserEventDisable       = "disable"
	UserEventEnable        = "enable"
	UserEventUnlock        = "unlock"
	UserEventDelete        = "delete"
)

// UserEventInfo contains information about a user event for admin notification
type UserEventInfo struct {
	EventType    string
	TargetUserID int64
	TargetEmail  string
	TargetName   string
	ActorUserID  *int64            // nil for self-registration
	ActorEmail   string            // empty for self-registration
	ActorName    string            // empty for self-registration
	Changes      map[string][2]string // field -> [old, new]
	Reason       string            // for disable events
	Timestamp    time.Time
}

// AdminNotificationService handles sending email notifications to admins for user events
type AdminNotificationService struct {
	userRepo         domain.UserRepository
	userSettingsRepo domain.UserSettingsRepository
	emailService     email.EmailService
	emailLogService  *EmailLogService
	appURL           string
	logger           *log.Logger
}

// NewAdminNotificationService creates a new admin notification service
func NewAdminNotificationService(
	userRepo domain.UserRepository,
	userSettingsRepo domain.UserSettingsRepository,
	emailService email.EmailService,
	emailLogService *EmailLogService,
	appURL string,
	logger *log.Logger,
) *AdminNotificationService {
	return &AdminNotificationService{
		userRepo:         userRepo,
		userSettingsRepo: userSettingsRepo,
		emailService:     emailService,
		emailLogService:  emailLogService,
		appURL:           appURL,
		logger:           logger,
	}
}

// NotifyAdminsOfUserEvent sends email notifications to all opted-in admins about a user event
// This method runs asynchronously and never blocks the caller
func (s *AdminNotificationService) NotifyAdminsOfUserEvent(event UserEventInfo) {
	go s.notifyAdminsAsync(event)
}

// notifyAdminsAsync performs the actual notification sending
func (s *AdminNotificationService) notifyAdminsAsync(event UserEventInfo) {
	// Skip if email service is not configured
	if s.emailService == nil {
		s.logger.Printf("[INFO] AdminNotificationService: Email service not configured, skipping notification")
		return
	}

	// Get all admin users
	admins, err := s.userRepo.ListAdmins()
	if err != nil {
		s.logger.Printf("[ERROR] AdminNotificationService: Failed to list admins: %v", err)
		return
	}

	if len(admins) == 0 {
		s.logger.Printf("[INFO] AdminNotificationService: No admin users found")
		return
	}

	// Build email content
	subject := s.buildSubject(event)
	htmlBody := s.buildHTMLBody(event)

	// Send to each opted-in admin
	for _, admin := range admins {
		// Skip if this admin is the actor (don't notify about your own actions)
		if event.ActorUserID != nil && admin.ID == *event.ActorUserID {
			continue
		}

		// Skip if admin account is disabled
		if admin.AccountDisabled {
			continue
		}

		// Check if admin has opted out of notifications
		settings, err := s.userSettingsRepo.GetByUserID(admin.ID)
		if err != nil {
			s.logger.Printf("[ERROR] AdminNotificationService: Failed to get settings for admin %d: %v", admin.ID, err)
			continue
		}

		// If settings exist and notifications are disabled, skip
		if settings != nil && !settings.AdminUserEventNotifications {
			s.logger.Printf("[INFO] AdminNotificationService: Admin %s has opted out of notifications", admin.Email)
			continue
		}

		// Send email
		err = s.emailService.SendHTMLEmail(admin.Email, subject, htmlBody)
		if err != nil {
			s.logger.Printf("[ERROR] AdminNotificationService: Failed to send email to %s: %v", admin.Email, err)
			// Log failed attempt if email log service is available
			if s.emailLogService != nil {
				errMsg := err.Error()
				s.emailLogService.LogEmailAttempt(
					admin.Email,
					domain.EmailTypeNotification,
					subject,
					false,
					&errMsg,
					nil,
					nil,
				)
			}
			continue
		}

		s.logger.Printf("[INFO] AdminNotificationService: Notification sent to %s for event %s", admin.Email, event.EventType)

		// Log successful email
		if s.emailLogService != nil {
			s.emailLogService.LogEmailAttempt(
				admin.Email,
				domain.EmailTypeNotification,
				subject,
				true,
				nil,
				nil,
				nil,
			)
		}
	}
}

// buildSubject generates the email subject based on event type
func (s *AdminNotificationService) buildSubject(event UserEventInfo) string {
	switch event.EventType {
	case UserEventRegistration:
		return fmt.Sprintf("[ActaLog] New User: %s", event.TargetEmail)
	case UserEventProfileUpdate:
		return fmt.Sprintf("[ActaLog] Profile Updated: %s", event.TargetEmail)
	case UserEventRoleChange:
		oldRole := ""
		newRole := ""
		if changes, ok := event.Changes["role"]; ok {
			oldRole = changes[0]
			newRole = changes[1]
		}
		return fmt.Sprintf("[ActaLog] Role Changed: %s (%s → %s)", event.TargetEmail, oldRole, newRole)
	case UserEventDisable:
		return fmt.Sprintf("[ActaLog] Account Disabled: %s", event.TargetEmail)
	case UserEventEnable:
		return fmt.Sprintf("[ActaLog] Account Enabled: %s", event.TargetEmail)
	case UserEventUnlock:
		return fmt.Sprintf("[ActaLog] Account Unlocked: %s", event.TargetEmail)
	case UserEventDelete:
		return fmt.Sprintf("[ActaLog] User Deleted: %s", event.TargetEmail)
	default:
		return fmt.Sprintf("[ActaLog] User Event: %s", event.TargetEmail)
	}
}

// buildHTMLBody generates the HTML email body
func (s *AdminNotificationService) buildHTMLBody(event UserEventInfo) string {
	eventTitle := s.getEventTitle(event.EventType)
	eventDescription := s.getEventDescription(event)
	changesHTML := s.buildChangesHTML(event)
	actorInfo := s.buildActorInfo(event)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #00bcd4; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f5f7fa; }
        .event-badge { display: inline-block; padding: 4px 12px; border-radius: 4px; font-weight: bold; margin-bottom: 10px; }
        .event-registration { background-color: #4caf50; color: white; }
        .event-profile_update { background-color: #2196f3; color: white; }
        .event-role_change { background-color: #ff9800; color: white; }
        .event-disable { background-color: #f44336; color: white; }
        .event-enable { background-color: #4caf50; color: white; }
        .event-unlock { background-color: #9c27b0; color: white; }
        .event-delete { background-color: #f44336; color: white; }
        .user-info { background-color: white; padding: 15px; border-radius: 8px; margin: 15px 0; }
        .user-info h3 { margin-top: 0; color: #2c3e50; }
        .changes-table { width: 100%%; border-collapse: collapse; margin: 15px 0; }
        .changes-table th, .changes-table td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        .changes-table th { background-color: #f0f0f0; }
        .old-value { color: #888; text-decoration: line-through; }
        .new-value { color: #4caf50; font-weight: bold; }
        .actor-info { background-color: #e8f4fd; padding: 10px; border-radius: 4px; margin: 15px 0; font-size: 14px; }
        .timestamp { color: #666; font-size: 12px; margin-top: 15px; }
        .button { display: inline-block; padding: 10px 20px; background-color: #00bcd4; color: white; text-decoration: none; border-radius: 4px; margin-top: 15px; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
        .opt-out { margin-top: 10px; color: #999; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ActaLog Admin Notification</h1>
        </div>
        <div class="content">
            <span class="event-badge event-%s">%s</span>

            <div class="user-info">
                <h3>User Details</h3>
                <p><strong>Email:</strong> %s</p>
                <p><strong>Name:</strong> %s</p>
                <p><strong>User ID:</strong> %d</p>
            </div>

            %s

            %s

            %s

            <p class="timestamp">Event Time: %s (UTC)</p>

            <a href="%s/admin/users" class="button">View in Admin Panel</a>
        </div>
        <div class="footer">
            <p>You received this notification because you are an administrator of ActaLog.</p>
            <p class="opt-out">To opt out of these notifications, update your settings in the ActaLog app.</p>
        </div>
    </div>
</body>
</html>
`,
		event.EventType,
		eventTitle,
		event.TargetEmail,
		event.TargetName,
		event.TargetUserID,
		eventDescription,
		changesHTML,
		actorInfo,
		event.Timestamp.UTC().Format("2006-01-02 15:04:05"),
		s.appURL,
	)
}

// getEventTitle returns a human-readable title for the event type
func (s *AdminNotificationService) getEventTitle(eventType string) string {
	switch eventType {
	case UserEventRegistration:
		return "New User Registration"
	case UserEventProfileUpdate:
		return "Profile Updated"
	case UserEventRoleChange:
		return "Role Changed"
	case UserEventDisable:
		return "Account Disabled"
	case UserEventEnable:
		return "Account Enabled"
	case UserEventUnlock:
		return "Account Unlocked"
	case UserEventDelete:
		return "User Deleted"
	default:
		return "User Event"
	}
}

// getEventDescription returns a description paragraph for the event
func (s *AdminNotificationService) getEventDescription(event UserEventInfo) string {
	switch event.EventType {
	case UserEventRegistration:
		return "<p>A new user has registered on ActaLog.</p>"
	case UserEventProfileUpdate:
		return "<p>A user has updated their profile information.</p>"
	case UserEventRoleChange:
		return "<p>A user's role has been changed by an administrator.</p>"
	case UserEventDisable:
		reason := ""
		if event.Reason != "" {
			reason = fmt.Sprintf(" Reason: %s", event.Reason)
		}
		return fmt.Sprintf("<p>A user's account has been disabled.%s</p>", reason)
	case UserEventEnable:
		return "<p>A user's account has been re-enabled.</p>"
	case UserEventUnlock:
		return "<p>A user's account has been unlocked.</p>"
	case UserEventDelete:
		return "<p>A user account has been deleted from the system.</p>"
	default:
		return "<p>A user event has occurred.</p>"
	}
}

// buildChangesHTML generates HTML for the changes table
func (s *AdminNotificationService) buildChangesHTML(event UserEventInfo) string {
	if len(event.Changes) == 0 {
		return ""
	}

	var rows strings.Builder
	for field, values := range event.Changes {
		fieldName := s.formatFieldName(field)
		rows.WriteString(fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td class="old-value">%s</td>
                <td class="new-value">%s</td>
            </tr>
`, fieldName, values[0], values[1]))
	}

	return fmt.Sprintf(`
            <h4>Changes</h4>
            <table class="changes-table">
                <tr>
                    <th>Field</th>
                    <th>Before</th>
                    <th>After</th>
                </tr>
                %s
            </table>
`, rows.String())
}

// formatFieldName converts field names to human-readable format
func (s *AdminNotificationService) formatFieldName(field string) string {
	switch field {
	case "name":
		return "Name"
	case "email":
		return "Email"
	case "role":
		return "Role"
	case "birthday":
		return "Birthday"
	default:
		// Convert snake_case to Title Case
		words := strings.Split(field, "_")
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(string(word[0])) + word[1:]
			}
		}
		return strings.Join(words, " ")
	}
}

// buildActorInfo generates HTML for who performed the action
func (s *AdminNotificationService) buildActorInfo(event UserEventInfo) string {
	if event.ActorUserID == nil {
		// Self-registration or automated action
		if event.EventType == UserEventRegistration {
			return `<div class="actor-info">This user registered themselves.</div>`
		}
		return ""
	}

	return fmt.Sprintf(`
            <div class="actor-info">
                <strong>Action performed by:</strong> %s (%s)
            </div>
`, event.ActorName, event.ActorEmail)
}
