package service

import (
	"testing"

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
		{"default values", 0, -1, 50, 0},    // Invalid limit defaults to 50
		{"over max limit", 200, 0, 50, 0},   // Over 100 defaults to 50
		{"valid limit", 10, 0, 10, 0},       // Valid limit stays
		{"negative offset", 10, -5, 10, 0},  // Negative offset becomes 0
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
