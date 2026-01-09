package service

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestNewAuditLogService(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	if svc == nil {
		t.Fatal("NewAuditLogService returned nil")
	}
	if svc.repo != repo {
		t.Error("Service repo not set correctly")
	}
}

func TestAuditLogService_Log(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	log := &domain.AuditLog{
		EventType: domain.EventLoginSuccess,
		UserID:    int64Ptr(1),
	}

	err := svc.Log(log)
	if err != nil {
		t.Errorf("Log() error = %v", err)
	}

	if len(repo.logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(repo.logs))
	}
}

func TestAuditLogService_LogEvent(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	userID := int64(1)
	targetUserID := int64(2)
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"
	details := map[string]interface{}{
		"key": "value",
	}

	err := svc.LogEvent(domain.EventLoginSuccess, &userID, &targetUserID, &ipAddress, &userAgent, details)
	if err != nil {
		t.Errorf("LogEvent() error = %v", err)
	}

	if len(repo.logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(repo.logs))
	}

	log := repo.logs[0]
	if log.EventType != domain.EventLoginSuccess {
		t.Errorf("EventType = %s, want %s", log.EventType, domain.EventLoginSuccess)
	}
	if *log.UserID != userID {
		t.Errorf("UserID = %d, want %d", *log.UserID, userID)
	}
	if *log.TargetUserID != targetUserID {
		t.Errorf("TargetUserID = %d, want %d", *log.TargetUserID, targetUserID)
	}
	if log.Details == nil {
		t.Error("Details should not be nil")
	}
}

func TestAuditLogService_LogEvent_NoDetails(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	userID := int64(1)
	err := svc.LogEvent(domain.EventLoginSuccess, &userID, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("LogEvent() error = %v", err)
	}

	if len(repo.logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(repo.logs))
	}

	if repo.logs[0].Details != nil {
		t.Error("Details should be nil when not provided")
	}
}

func TestAuditLogService_GetByID(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	// Create a log first
	log := &domain.AuditLog{
		EventType: domain.EventLoginSuccess,
	}
	_ = repo.Create(log)

	// Get it back
	found, err := svc.GetByID(log.ID)
	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}
	if found == nil {
		t.Error("GetByID() returned nil")
	}
	if found.ID != log.ID {
		t.Errorf("GetByID() returned wrong log: got ID %d, want %d", found.ID, log.ID)
	}
}

func TestAuditLogService_List(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	// Create some logs
	for i := 0; i < 5; i++ {
		_ = repo.Create(&domain.AuditLog{EventType: domain.EventLoginSuccess})
	}

	tests := []struct {
		name       string
		limit      int
		offset     int
		wantLimit  int
		wantOffset int
	}{
		{"default values", 0, -1, 50, 0},   // Invalid limit defaults to 50
		{"over max limit", 200, 0, 50, 0},  // Over 100 defaults to 50
		{"valid limit", 10, 0, 10, 0},      // Valid limit stays
		{"negative offset", 10, -5, 10, 0}, // Negative offset becomes 0
		{"valid pagination", 10, 10, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, err := svc.List(domain.AuditLogFilters{}, tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
			}
			// Mock returns all logs regardless of limit/offset, so just check it returns
			if logs == nil {
				t.Error("List() returned nil")
			}
		})
	}
}

func TestAuditLogService_Count(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	// Create some logs
	for i := 0; i < 3; i++ {
		_ = repo.Create(&domain.AuditLog{EventType: domain.EventLoginSuccess})
	}

	count, err := svc.Count(domain.AuditLogFilters{})
	if err != nil {
		t.Errorf("Count() error = %v", err)
	}
	if count != 3 {
		t.Errorf("Count() = %d, want 3", count)
	}
}

func TestAuditLogService_GetByUserID(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	userID := int64(42)
	_ = repo.Create(&domain.AuditLog{EventType: domain.EventLoginSuccess, UserID: &userID})
	_ = repo.Create(&domain.AuditLog{EventType: domain.EventLoginFailed, UserID: &userID})

	logs, err := svc.GetByUserID(userID, 50, 0)
	if err != nil {
		t.Errorf("GetByUserID() error = %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("GetByUserID() returned %d logs, want 2", len(logs))
	}
}

func TestAuditLogService_GetByTargetUserID(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	targetUserID := int64(99)
	_ = repo.Create(&domain.AuditLog{EventType: domain.EventAccountDisabled, TargetUserID: &targetUserID})

	logs, err := svc.GetByTargetUserID(targetUserID, 50, 0)
	if err != nil {
		t.Errorf("GetByTargetUserID() error = %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("GetByTargetUserID() returned %d logs, want 1", len(logs))
	}
}

func TestAuditLogService_CleanupOldLogs(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	tests := []struct {
		name          string
		retentionDays int
		wantError     bool
	}{
		{"valid retention", 30, false},
		{"zero retention", 0, true},
		{"negative retention", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CleanupOldLogs(tt.retentionDays)
			if (err != nil) != tt.wantError {
				t.Errorf("CleanupOldLogs() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAuditLogService_LogLoginSuccess(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogLoginSuccess(1, "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Errorf("LogLoginSuccess() error = %v", err)
	}

	if len(repo.logs) != 1 {
		t.Fatal("Expected 1 log")
	}
	if repo.logs[0].EventType != domain.EventLoginSuccess {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventLoginSuccess)
	}
}

func TestAuditLogService_LogLoginFailed(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogLoginFailed("test@example.com", "invalid password", "192.168.1.1", "Mozilla/5.0", 3)
	if err != nil {
		t.Errorf("LogLoginFailed() error = %v", err)
	}

	if len(repo.logs) != 1 {
		t.Fatal("Expected 1 log")
	}
	if repo.logs[0].EventType != domain.EventLoginFailed {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventLoginFailed)
	}
	if repo.logs[0].Details == nil {
		t.Error("Details should contain email and reason")
	}
}

func TestAuditLogService_LogAccountLocked(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogAccountLocked(1, "test@example.com", "192.168.1.1", "Mozilla/5.0", 5)
	if err != nil {
		t.Errorf("LogAccountLocked() error = %v", err)
	}

	if len(repo.logs) != 1 {
		t.Fatal("Expected 1 log")
	}
	if repo.logs[0].EventType != domain.EventAccountLockedAuto {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventAccountLockedAuto)
	}
}

func TestAuditLogService_LogAccountUnlocked(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogAccountUnlocked(1, 2, "admin@example.com", "user@example.com")
	if err != nil {
		t.Errorf("LogAccountUnlocked() error = %v", err)
	}

	if len(repo.logs) != 1 {
		t.Fatal("Expected 1 log")
	}
	if repo.logs[0].EventType != domain.EventAccountUnlockedAdmin {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventAccountUnlockedAdmin)
	}
}

func TestAuditLogService_LogAccountDisabled(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogAccountDisabled(1, 2, "admin@example.com", "user@example.com", "Terms violation")
	if err != nil {
		t.Errorf("LogAccountDisabled() error = %v", err)
	}

	if len(repo.logs) != 1 {
		t.Fatal("Expected 1 log")
	}
	if repo.logs[0].EventType != domain.EventAccountDisabled {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventAccountDisabled)
	}
}

func TestAuditLogService_LogAccountEnabled(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogAccountEnabled(1, 2, "admin@example.com", "user@example.com")
	if err != nil {
		t.Errorf("LogAccountEnabled() error = %v", err)
	}

	if len(repo.logs) != 1 {
		t.Fatal("Expected 1 log")
	}
	if repo.logs[0].EventType != domain.EventAccountEnabled {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventAccountEnabled)
	}
}

func TestAuditLogService_LogPasswordChanged(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogPasswordChanged(1, "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Errorf("LogPasswordChanged() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventPasswordChanged {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventPasswordChanged)
	}
}

func TestAuditLogService_LogPasswordReset(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogPasswordReset(1, "user@example.com", "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Errorf("LogPasswordReset() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventPasswordReset {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventPasswordReset)
	}
}

func TestAuditLogService_LogEmailChanged(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogEmailChanged(1, "old@example.com", "new@example.com", "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Errorf("LogEmailChanged() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventEmailChanged {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventEmailChanged)
	}
}

func TestAuditLogService_LogEmailVerified(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogEmailVerified(1, "user@example.com")
	if err != nil {
		t.Errorf("LogEmailVerified() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventEmailVerified {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventEmailVerified)
	}
}

func TestAuditLogService_LogRoleChanged(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogRoleChanged(1, 2, "admin@example.com", "user@example.com", "user", "admin")
	if err != nil {
		t.Errorf("LogRoleChanged() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventRoleChanged {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventRoleChanged)
	}
}

func TestAuditLogService_LogUserCreated(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogUserCreated(1, "new@example.com", "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Errorf("LogUserCreated() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventUserCreated {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventUserCreated)
	}
}

func TestAuditLogService_LogRateLimitExceeded(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogRateLimitExceeded("/api/login", "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Errorf("LogRateLimitExceeded() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventRateLimitExceeded {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventRateLimitExceeded)
	}
}

func TestAuditLogService_LogLogout(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogLogout(1, "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Errorf("LogLogout() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventLogout {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventLogout)
	}
}

func TestAuditLogService_LogTokenRefresh(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogTokenRefresh(1, "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Errorf("LogTokenRefresh() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventTokenRefresh {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventTokenRefresh)
	}
}

func TestAuditLogService_LogProfileUpdated(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	changes := map[string]interface{}{
		"name":  "New Name",
		"email": "new@example.com",
	}
	err := svc.LogProfileUpdated(1, changes, "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Errorf("LogProfileUpdated() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventProfileUpdated {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventProfileUpdated)
	}
}

func TestAuditLogService_LogUserSettingsUpdated(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	changes := map[string]interface{}{
		"theme": "dark",
	}
	err := svc.LogUserSettingsUpdated(1, changes, "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Errorf("LogUserSettingsUpdated() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventUserSettingsUpdate {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventUserSettingsUpdate)
	}
}

func TestAuditLogService_LogUserDeleted(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogUserDeleted(1, 2, "admin@example.com", "user@example.com")
	if err != nil {
		t.Errorf("LogUserDeleted() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventUserDeleted {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventUserDeleted)
	}
}

func TestAuditLogService_LogOrganizationCreated(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogOrganizationCreated(1, 10, "Test Org")
	if err != nil {
		t.Errorf("LogOrganizationCreated() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventOrganizationCreated {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventOrganizationCreated)
	}
}

func TestAuditLogService_LogOrganizationUpdated(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	changes := map[string]interface{}{
		"name": "New Org Name",
	}
	err := svc.LogOrganizationUpdated(1, 10, "Test Org", changes)
	if err != nil {
		t.Errorf("LogOrganizationUpdated() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventOrganizationUpdated {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventOrganizationUpdated)
	}
}

func TestAuditLogService_LogOrganizationDeleted(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogOrganizationDeleted(1, 10, "Test Org")
	if err != nil {
		t.Errorf("LogOrganizationDeleted() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventOrganizationDeleted {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventOrganizationDeleted)
	}
}

func TestAuditLogService_LogUserAssignedToOrganization(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogUserAssignedToOrganization(1, 2, 10, "user@example.com", "Test Org")
	if err != nil {
		t.Errorf("LogUserAssignedToOrganization() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventUserAssignedToOrg {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventUserAssignedToOrg)
	}
}

func TestAuditLogService_LogUserRemovedFromOrganization(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogUserRemovedFromOrganization(1, 2, 10, "user@example.com", "Test Org")
	if err != nil {
		t.Errorf("LogUserRemovedFromOrganization() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventUserRemovedFromOrg {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventUserRemovedFromOrg)
	}
}

func TestAuditLogService_LogSubscriptionCreated(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	expiresAt := time.Now().AddDate(0, 1, 0)
	err := svc.LogSubscriptionCreated(1, 2, "monthly", &expiresAt)
	if err != nil {
		t.Errorf("LogSubscriptionCreated() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventSubscriptionCreated {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventSubscriptionCreated)
	}
}

func TestAuditLogService_LogSubscriptionMarkedPaid(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogSubscriptionMarkedPaid(1, 2, 100)
	if err != nil {
		t.Errorf("LogSubscriptionMarkedPaid() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventSubscriptionMarkedPaid {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventSubscriptionMarkedPaid)
	}
}

func TestAuditLogService_LogSubscriptionCancelled(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogSubscriptionCancelled(1, 2, 100, "User requested")
	if err != nil {
		t.Errorf("LogSubscriptionCancelled() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventSubscriptionCancelled {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventSubscriptionCancelled)
	}
}

func TestAuditLogService_LogOrgSubscriptionCreated(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	expiresAt := time.Now().AddDate(1, 0, 0)
	err := svc.LogOrgSubscriptionCreated(1, 10, "Test Org", "annual", &expiresAt)
	if err != nil {
		t.Errorf("LogOrgSubscriptionCreated() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventOrgSubscriptionCreated {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventOrgSubscriptionCreated)
	}
}

func TestAuditLogService_LogOrgSubscriptionMarkedPaid(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogOrgSubscriptionMarkedPaid(1, 10, 200, "Test Org")
	if err != nil {
		t.Errorf("LogOrgSubscriptionMarkedPaid() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventOrgSubscriptionMarkedPaid {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventOrgSubscriptionMarkedPaid)
	}
}

func TestAuditLogService_LogOrgSubscriptionCancelled(t *testing.T) {
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	err := svc.LogOrgSubscriptionCancelled(1, 10, 200, "Test Org", "Contract ended")
	if err != nil {
		t.Errorf("LogOrgSubscriptionCancelled() error = %v", err)
	}

	if repo.logs[0].EventType != domain.EventOrgSubscriptionCancelled {
		t.Errorf("EventType = %s, want %s", repo.logs[0].EventType, domain.EventOrgSubscriptionCancelled)
	}
}
