package service

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func createTestAdminNotificationService() *AdminNotificationService {
	logger := log.New(os.Stdout, "", 0)
	return &AdminNotificationService{
		appURL: "https://example.com",
		logger: logger,
	}
}

func TestNewAdminNotificationService(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)

	svc := NewAdminNotificationService(
		nil, // userRepo
		nil, // userSettingsRepo
		nil, // notificationRepo
		nil, // emailService
		nil, // emailLogService
		"https://example.com",
		logger,
	)

	if svc == nil {
		t.Fatal("NewAdminNotificationService() returned nil")
	}

	if svc.appURL != "https://example.com" {
		t.Errorf("appURL = %q, want %q", svc.appURL, "https://example.com")
	}

	if svc.logger != logger {
		t.Error("logger not set correctly")
	}
}

func TestBuildSubject(t *testing.T) {
	svc := createTestAdminNotificationService()

	tests := []struct {
		name     string
		event    UserEventInfo
		contains string
	}{
		{
			name: "registration event",
			event: UserEventInfo{
				EventType:   UserEventRegistration,
				TargetEmail: "newuser@example.com",
			},
			contains: "New User: newuser@example.com",
		},
		{
			name: "profile update event",
			event: UserEventInfo{
				EventType:   UserEventProfileUpdate,
				TargetEmail: "user@example.com",
			},
			contains: "Profile Updated: user@example.com",
		},
		{
			name: "role change event",
			event: UserEventInfo{
				EventType:   UserEventRoleChange,
				TargetEmail: "user@example.com",
				Changes: map[string][2]string{
					"role": {"athlete", "admin"},
				},
			},
			contains: "Role Changed: user@example.com (athlete → admin)",
		},
		{
			name: "disable event",
			event: UserEventInfo{
				EventType:   UserEventDisable,
				TargetEmail: "user@example.com",
			},
			contains: "Account Disabled: user@example.com",
		},
		{
			name: "enable event",
			event: UserEventInfo{
				EventType:   UserEventEnable,
				TargetEmail: "user@example.com",
			},
			contains: "Account Enabled: user@example.com",
		},
		{
			name: "unlock event",
			event: UserEventInfo{
				EventType:   UserEventUnlock,
				TargetEmail: "user@example.com",
			},
			contains: "Account Unlocked: user@example.com",
		},
		{
			name: "delete event",
			event: UserEventInfo{
				EventType:   UserEventDelete,
				TargetEmail: "user@example.com",
			},
			contains: "User Deleted: user@example.com",
		},
		{
			name: "unknown event type",
			event: UserEventInfo{
				EventType:   "unknown",
				TargetEmail: "user@example.com",
			},
			contains: "User Event: user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.buildSubject(tt.event)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("buildSubject() = %q, want to contain %q", result, tt.contains)
			}
			// All subjects should have [ActaLog] prefix
			if !strings.HasPrefix(result, "[ActaLog]") {
				t.Errorf("buildSubject() = %q, should start with [ActaLog]", result)
			}
		})
	}
}

func TestGetEventTitle(t *testing.T) {
	svc := createTestAdminNotificationService()

	tests := []struct {
		eventType string
		expected  string
	}{
		{UserEventRegistration, "New User Registration"},
		{UserEventProfileUpdate, "Profile Updated"},
		{UserEventRoleChange, "Role Changed"},
		{UserEventDisable, "Account Disabled"},
		{UserEventEnable, "Account Enabled"},
		{UserEventUnlock, "Account Unlocked"},
		{UserEventDelete, "User Deleted"},
		{"unknown", "User Event"},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			result := svc.getEventTitle(tt.eventType)
			if result != tt.expected {
				t.Errorf("getEventTitle(%q) = %q, want %q", tt.eventType, result, tt.expected)
			}
		})
	}
}

func TestGetEventDescription(t *testing.T) {
	svc := createTestAdminNotificationService()

	tests := []struct {
		name     string
		event    UserEventInfo
		contains string
	}{
		{
			name:     "registration",
			event:    UserEventInfo{EventType: UserEventRegistration},
			contains: "new user has registered",
		},
		{
			name:     "profile update",
			event:    UserEventInfo{EventType: UserEventProfileUpdate},
			contains: "updated their profile",
		},
		{
			name:     "role change",
			event:    UserEventInfo{EventType: UserEventRoleChange},
			contains: "role has been changed",
		},
		{
			name:     "disable without reason",
			event:    UserEventInfo{EventType: UserEventDisable},
			contains: "account has been disabled",
		},
		{
			name:     "disable with reason",
			event:    UserEventInfo{EventType: UserEventDisable, Reason: "Policy violation"},
			contains: "Policy violation",
		},
		{
			name:     "enable",
			event:    UserEventInfo{EventType: UserEventEnable},
			contains: "re-enabled",
		},
		{
			name:     "unlock",
			event:    UserEventInfo{EventType: UserEventUnlock},
			contains: "unlocked",
		},
		{
			name:     "delete",
			event:    UserEventInfo{EventType: UserEventDelete},
			contains: "deleted from the system",
		},
		{
			name:     "unknown",
			event:    UserEventInfo{EventType: "unknown"},
			contains: "event has occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.getEventDescription(tt.event)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("getEventDescription() = %q, want to contain %q", result, tt.contains)
			}
		})
	}
}

func TestBuildChangesHTML(t *testing.T) {
	svc := createTestAdminNotificationService()

	tests := []struct {
		name     string
		event    UserEventInfo
		contains []string
		isEmpty  bool
	}{
		{
			name:    "no changes",
			event:   UserEventInfo{Changes: nil},
			isEmpty: true,
		},
		{
			name:    "empty changes map",
			event:   UserEventInfo{Changes: map[string][2]string{}},
			isEmpty: true,
		},
		{
			name: "single change",
			event: UserEventInfo{
				Changes: map[string][2]string{
					"name": {"Old Name", "New Name"},
				},
			},
			contains: []string{"Name", "Old Name", "New Name", "Changes"},
		},
		{
			name: "multiple changes",
			event: UserEventInfo{
				Changes: map[string][2]string{
					"name":  {"Old", "New"},
					"email": {"old@example.com", "new@example.com"},
				},
			},
			contains: []string{"Name", "Email", "changes-table"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.buildChangesHTML(tt.event)
			if tt.isEmpty {
				if result != "" {
					t.Errorf("buildChangesHTML() = %q, want empty string", result)
				}
				return
			}
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("buildChangesHTML() missing %q", expected)
				}
			}
		})
	}
}

func TestFormatFieldName(t *testing.T) {
	svc := createTestAdminNotificationService()

	tests := []struct {
		input    string
		expected string
	}{
		{"name", "Name"},
		{"email", "Email"},
		{"role", "Role"},
		{"birthday", "Birthday"},
		{"profile_image", "Profile Image"},
		{"custom_field", "Custom Field"},
		{"single", "Single"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := svc.formatFieldName(tt.input)
			if result != tt.expected {
				t.Errorf("formatFieldName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildActorInfo(t *testing.T) {
	svc := createTestAdminNotificationService()
	actorID := int64(123)

	tests := []struct {
		name     string
		event    UserEventInfo
		contains string
		isEmpty  bool
	}{
		{
			name: "self registration",
			event: UserEventInfo{
				EventType:   UserEventRegistration,
				ActorUserID: nil,
			},
			contains: "registered themselves",
		},
		{
			name: "other event without actor",
			event: UserEventInfo{
				EventType:   UserEventProfileUpdate,
				ActorUserID: nil,
			},
			isEmpty: true,
		},
		{
			name: "action by admin",
			event: UserEventInfo{
				EventType:   UserEventDisable,
				ActorUserID: &actorID,
				ActorName:   "Admin User",
				ActorEmail:  "admin@example.com",
			},
			contains: "Admin User (admin@example.com)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.buildActorInfo(tt.event)
			if tt.isEmpty {
				if result != "" {
					t.Errorf("buildActorInfo() = %q, want empty string", result)
				}
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("buildActorInfo() = %q, want to contain %q", result, tt.contains)
			}
		})
	}
}

func TestBuildNotificationTitle(t *testing.T) {
	svc := createTestAdminNotificationService()

	tests := []struct {
		eventType string
		expected  string
	}{
		{UserEventRegistration, "New User Registration"},
		{UserEventProfileUpdate, "User Profile Updated"},
		{UserEventRoleChange, "User Role Changed"},
		{UserEventDisable, "User Account Disabled"},
		{UserEventEnable, "User Account Enabled"},
		{UserEventUnlock, "User Account Unlocked"},
		{UserEventDelete, "User Deleted"},
		{"unknown", "User Event"},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			event := UserEventInfo{EventType: tt.eventType}
			result := svc.buildNotificationTitle(event)
			if result != tt.expected {
				t.Errorf("buildNotificationTitle() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildNotificationMessage(t *testing.T) {
	svc := createTestAdminNotificationService()
	actorID := int64(1)

	tests := []struct {
		name     string
		event    UserEventInfo
		contains []string
	}{
		{
			name: "registration",
			event: UserEventInfo{
				EventType:   UserEventRegistration,
				TargetName:  "New User",
				TargetEmail: "new@example.com",
			},
			contains: []string{"New User", "new@example.com", "registered"},
		},
		{
			name: "profile update without changes",
			event: UserEventInfo{
				EventType:   UserEventProfileUpdate,
				TargetEmail: "user@example.com",
			},
			contains: []string{"user@example.com", "updated their profile"},
		},
		{
			name: "profile update with changes",
			event: UserEventInfo{
				EventType:   UserEventProfileUpdate,
				TargetEmail: "user@example.com",
				Changes: map[string][2]string{
					"name": {"Old", "New"},
				},
			},
			contains: []string{"user@example.com", "Name changed"},
		},
		{
			name: "role change with details",
			event: UserEventInfo{
				EventType:   UserEventRoleChange,
				TargetEmail: "user@example.com",
				ActorUserID: &actorID,
				ActorEmail:  "admin@example.com",
				Changes: map[string][2]string{
					"role": {"athlete", "admin"},
				},
			},
			contains: []string{"user@example.com", "athlete", "admin"},
		},
		{
			name: "role change without details",
			event: UserEventInfo{
				EventType:   UserEventRoleChange,
				TargetEmail: "user@example.com",
				ActorUserID: &actorID,
				ActorEmail:  "admin@example.com",
			},
			contains: []string{"user@example.com", "admin@example.com"},
		},
		{
			name: "disable without reason",
			event: UserEventInfo{
				EventType:   UserEventDisable,
				TargetEmail: "user@example.com",
				ActorUserID: &actorID,
				ActorEmail:  "admin@example.com",
			},
			contains: []string{"user@example.com", "disabled"},
		},
		{
			name: "disable with reason",
			event: UserEventInfo{
				EventType:   UserEventDisable,
				TargetEmail: "user@example.com",
				ActorUserID: &actorID,
				ActorEmail:  "admin@example.com",
				Reason:      "Policy violation",
			},
			contains: []string{"user@example.com", "disabled", "Policy violation"},
		},
		{
			name: "enable",
			event: UserEventInfo{
				EventType:   UserEventEnable,
				TargetEmail: "user@example.com",
				ActorUserID: &actorID,
				ActorEmail:  "admin@example.com",
			},
			contains: []string{"user@example.com", "re-enabled"},
		},
		{
			name: "unlock",
			event: UserEventInfo{
				EventType:   UserEventUnlock,
				TargetEmail: "user@example.com",
				ActorUserID: &actorID,
				ActorEmail:  "admin@example.com",
			},
			contains: []string{"user@example.com", "unlocked"},
		},
		{
			name: "delete",
			event: UserEventInfo{
				EventType:   UserEventDelete,
				TargetEmail: "user@example.com",
				ActorUserID: &actorID,
				ActorEmail:  "admin@example.com",
			},
			contains: []string{"user@example.com", "deleted"},
		},
		{
			name: "unknown event",
			event: UserEventInfo{
				EventType:   "unknown",
				TargetEmail: "user@example.com",
			},
			contains: []string{"user@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.buildNotificationMessage(tt.event)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("buildNotificationMessage() = %q, want to contain %q", result, expected)
				}
			}
		})
	}
}

func TestFormatChangesForMessage(t *testing.T) {
	svc := createTestAdminNotificationService()

	tests := []struct {
		name     string
		changes  map[string][2]string
		contains []string
		isEmpty  bool
	}{
		{
			name:    "nil changes",
			changes: nil,
			isEmpty: true,
		},
		{
			name:    "empty changes",
			changes: map[string][2]string{},
			isEmpty: true,
		},
		{
			name: "single change",
			changes: map[string][2]string{
				"name": {"Old", "New"},
			},
			contains: []string{"Name changed"},
		},
		{
			name: "multiple changes",
			changes: map[string][2]string{
				"name":  {"Old", "New"},
				"email": {"old@example.com", "new@example.com"},
			},
			contains: []string{"changed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.formatChangesForMessage(tt.changes)
			if tt.isEmpty {
				if result != "" {
					t.Errorf("formatChangesForMessage() = %q, want empty string", result)
				}
				return
			}
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("formatChangesForMessage() = %q, want to contain %q", result, expected)
				}
			}
		})
	}
}

func TestBuildNotificationData(t *testing.T) {
	svc := createTestAdminNotificationService()
	actorID := int64(123)

	tests := []struct {
		name       string
		event      UserEventInfo
		checkKeys  []string
		checkActor bool
	}{
		{
			name: "basic event",
			event: UserEventInfo{
				EventType:    UserEventRegistration,
				TargetUserID: 456,
				TargetEmail:  "user@example.com",
				TargetName:   "Test User",
			},
			checkKeys:  []string{"event_type", "target_user_id", "target_email", "target_name"},
			checkActor: false,
		},
		{
			name: "event with actor",
			event: UserEventInfo{
				EventType:    UserEventDisable,
				TargetUserID: 456,
				TargetEmail:  "user@example.com",
				TargetName:   "Test User",
				ActorUserID:  &actorID,
				ActorEmail:   "admin@example.com",
				ActorName:    "Admin",
			},
			checkKeys:  []string{"event_type", "target_user_id", "actor_user_id", "actor_email"},
			checkActor: true,
		},
		{
			name: "event with changes",
			event: UserEventInfo{
				EventType:    UserEventProfileUpdate,
				TargetUserID: 456,
				TargetEmail:  "user@example.com",
				TargetName:   "Test User",
				Changes: map[string][2]string{
					"name": {"Old", "New"},
				},
			},
			checkKeys: []string{"event_type", "changes"},
		},
		{
			name: "event with reason",
			event: UserEventInfo{
				EventType:    UserEventDisable,
				TargetUserID: 456,
				TargetEmail:  "user@example.com",
				TargetName:   "Test User",
				ActorUserID:  &actorID,
				ActorEmail:   "admin@example.com",
				Reason:       "Policy violation",
			},
			checkKeys: []string{"event_type", "reason"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.buildNotificationData(tt.event)

			// Should be valid JSON
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(result), &data); err != nil {
				t.Fatalf("buildNotificationData() returned invalid JSON: %v", err)
			}

			// Check for expected keys
			for _, key := range tt.checkKeys {
				if _, ok := data[key]; !ok {
					t.Errorf("buildNotificationData() missing key %q", key)
				}
			}

			// Verify actor info if expected
			if tt.checkActor {
				if _, ok := data["actor_user_id"]; !ok {
					t.Error("buildNotificationData() missing actor_user_id")
				}
			}
		})
	}
}

func TestBuildHTMLBody(t *testing.T) {
	svc := createTestAdminNotificationService()
	actorID := int64(1)

	event := UserEventInfo{
		EventType:    UserEventRegistration,
		TargetUserID: 123,
		TargetEmail:  "newuser@example.com",
		TargetName:   "New User",
		ActorUserID:  &actorID,
		ActorEmail:   "admin@example.com",
		ActorName:    "Admin",
		Timestamp:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	result := svc.buildHTMLBody(event)

	// Check that HTML contains expected elements
	expectedContents := []string{
		"<!DOCTYPE html>",
		"ActaLog Admin Notification",
		"newuser@example.com",
		"New User",
		"123", // User ID
		"https://example.com/admin/users",
		"2024-01-15",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(result, expected) {
			t.Errorf("buildHTMLBody() missing %q", expected)
		}
	}
}

func TestBuildHTMLBody_AllEventTypes(t *testing.T) {
	svc := createTestAdminNotificationService()

	eventTypes := []string{
		UserEventRegistration,
		UserEventProfileUpdate,
		UserEventRoleChange,
		UserEventDisable,
		UserEventEnable,
		UserEventUnlock,
		UserEventDelete,
	}

	for _, eventType := range eventTypes {
		t.Run(eventType, func(t *testing.T) {
			event := UserEventInfo{
				EventType:    eventType,
				TargetUserID: 123,
				TargetEmail:  "user@example.com",
				TargetName:   "Test User",
				Timestamp:    time.Now(),
			}

			result := svc.buildHTMLBody(event)

			// All event types should produce valid HTML
			if !strings.Contains(result, "<!DOCTYPE html>") {
				t.Error("buildHTMLBody() should produce valid HTML")
			}
			if !strings.Contains(result, "ActaLog Admin Notification") {
				t.Error("buildHTMLBody() should contain header")
			}
			if !strings.Contains(result, "user@example.com") {
				t.Error("buildHTMLBody() should contain target email")
			}
		})
	}
}

func TestUserEventConstants(t *testing.T) {
	// Verify constants are defined correctly
	constants := map[string]string{
		"UserEventRegistration":  UserEventRegistration,
		"UserEventProfileUpdate": UserEventProfileUpdate,
		"UserEventRoleChange":    UserEventRoleChange,
		"UserEventDisable":       UserEventDisable,
		"UserEventEnable":        UserEventEnable,
		"UserEventUnlock":        UserEventUnlock,
		"UserEventDelete":        UserEventDelete,
	}

	for name, value := range constants {
		if value == "" {
			t.Errorf("Constant %s should not be empty", name)
		}
	}
}
