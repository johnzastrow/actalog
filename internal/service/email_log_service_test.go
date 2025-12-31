package service

import (
	"errors"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/email"
)

// mockEmailLogRepo is a mock implementation of EmailLogRepository
type mockEmailLogRepo struct {
	logs        []*domain.EmailLog
	createErr   error
	getByIDErr  error
	listErr     error
	countErr    error
	deleteErr   error
	nextID      int64
	countResult int
}

func newMockEmailLogRepo() *mockEmailLogRepo {
	return &mockEmailLogRepo{
		logs:   make([]*domain.EmailLog, 0),
		nextID: 1,
	}
}

func (m *mockEmailLogRepo) Create(log *domain.EmailLog) error {
	if m.createErr != nil {
		return m.createErr
	}
	log.ID = m.nextID
	log.CreatedAt = time.Now()
	m.nextID++
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockEmailLogRepo) GetByID(id int64) (*domain.EmailLog, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, log := range m.logs {
		if log.ID == id {
			return log, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockEmailLogRepo) List(filters domain.EmailLogFilters, limit, offset int) ([]*domain.EmailLog, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}

	// Apply filters
	var result []*domain.EmailLog
	for _, log := range m.logs {
		if filters.EmailType != nil && log.EmailType != *filters.EmailType {
			continue
		}
		if filters.Success != nil && log.Success != *filters.Success {
			continue
		}
		if filters.RecipientEmail != nil && log.RecipientEmail != *filters.RecipientEmail {
			continue
		}
		result = append(result, log)
	}

	// Apply pagination
	if offset >= len(result) {
		return []*domain.EmailLog{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (m *mockEmailLogRepo) Count(filters domain.EmailLogFilters) (int, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	if m.countResult > 0 {
		return m.countResult, nil
	}

	count := 0
	for _, log := range m.logs {
		if filters.EmailType != nil && log.EmailType != *filters.EmailType {
			continue
		}
		if filters.Success != nil && log.Success != *filters.Success {
			continue
		}
		count++
	}
	return count, nil
}

func (m *mockEmailLogRepo) DeleteOlderThan(before time.Time) (int64, error) {
	if m.deleteErr != nil {
		return 0, m.deleteErr
	}

	var remaining []*domain.EmailLog
	var deleted int64
	for _, log := range m.logs {
		if log.CreatedAt.Before(before) {
			deleted++
		} else {
			remaining = append(remaining, log)
		}
	}
	m.logs = remaining
	return deleted, nil
}

func TestNewEmailLogService(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	if service == nil {
		t.Error("NewEmailLogService should not return nil")
	}
	if service.repo != repo {
		t.Error("NewEmailLogService should set the repo")
	}
}

func TestEmailLogService_LogEmailAttempt_Success(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	userID := int64(1)
	err := service.LogEmailAttempt(
		"test@example.com",
		domain.EmailTypeTest,
		"Test Subject",
		true,
		nil,
		nil,
		&userID,
	)

	if err != nil {
		t.Errorf("LogEmailAttempt should not return error: %v", err)
	}
	if len(repo.logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(repo.logs))
	}
	if repo.logs[0].RecipientEmail != "test@example.com" {
		t.Errorf("Expected recipient test@example.com, got %s", repo.logs[0].RecipientEmail)
	}
	if repo.logs[0].EmailType != domain.EmailTypeTest {
		t.Errorf("Expected type %s, got %s", domain.EmailTypeTest, repo.logs[0].EmailType)
	}
	if !repo.logs[0].Success {
		t.Error("Expected success to be true")
	}
}

func TestEmailLogService_LogEmailAttempt_WithDebugInfo(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	debugInfo := &email.SMTPDebugInfo{
		ConnectionHost: "smtp.example.com",
		ConnectionPort: 587,
		ConnectionTLS:  "STARTTLS",
		AuthMethod:     "PLAIN",
	}

	err := service.LogEmailAttempt(
		"test@example.com",
		domain.EmailTypeTest,
		"Test Subject",
		true,
		nil,
		debugInfo,
		nil,
	)

	if err != nil {
		t.Errorf("LogEmailAttempt should not return error: %v", err)
	}
	if repo.logs[0].DebugInfo == nil {
		t.Error("Expected debug info to be set")
	}
}

func TestEmailLogService_LogEmailAttempt_WithError(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	errMsg := "SMTP connection failed"
	err := service.LogEmailAttempt(
		"test@example.com",
		domain.EmailTypeTest,
		"Test Subject",
		false,
		&errMsg,
		nil,
		nil,
	)

	if err != nil {
		t.Errorf("LogEmailAttempt should not return error: %v", err)
	}
	if !repo.logs[0].Success != true {
		t.Error("Expected success to be false")
	}
	if repo.logs[0].ErrorMessage == nil || *repo.logs[0].ErrorMessage != errMsg {
		t.Error("Expected error message to be set")
	}
}

func TestEmailLogService_LogEmailAttempt_RepoError(t *testing.T) {
	repo := newMockEmailLogRepo()
	repo.createErr = errors.New("database error")
	service := NewEmailLogService(repo)

	err := service.LogEmailAttempt(
		"test@example.com",
		domain.EmailTypeTest,
		"Test Subject",
		true,
		nil,
		nil,
		nil,
	)

	if err == nil {
		t.Error("LogEmailAttempt should return error when repo fails")
	}
}

func TestEmailLogService_LogTestEmail_Success(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	result := email.SendResult{
		Success: true,
	}

	err := service.LogTestEmail("test@example.com", "Test Email", result, 1)

	if err != nil {
		t.Errorf("LogTestEmail should not return error: %v", err)
	}
	if len(repo.logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(repo.logs))
	}
	if repo.logs[0].EmailType != domain.EmailTypeTest {
		t.Errorf("Expected type %s, got %s", domain.EmailTypeTest, repo.logs[0].EmailType)
	}
}

func TestEmailLogService_LogTestEmail_WithError(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	result := email.SendResult{
		Success: false,
		Error:   errors.New("SMTP failed"),
	}

	err := service.LogTestEmail("test@example.com", "Test Email", result, 1)

	if err != nil {
		t.Errorf("LogTestEmail should not return error: %v", err)
	}
	if repo.logs[0].Success {
		t.Error("Expected success to be false")
	}
	if repo.logs[0].ErrorMessage == nil {
		t.Error("Expected error message to be set")
	}
}

func TestEmailLogService_LogPasswordResetEmail_Success(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	result := email.SendResult{
		Success: true,
	}

	err := service.LogPasswordResetEmail("user@example.com", "Password Reset", result)

	if err != nil {
		t.Errorf("LogPasswordResetEmail should not return error: %v", err)
	}
	if repo.logs[0].EmailType != domain.EmailTypePasswordReset {
		t.Errorf("Expected type %s, got %s", domain.EmailTypePasswordReset, repo.logs[0].EmailType)
	}
}

func TestEmailLogService_LogPasswordResetEmail_WithError(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	result := email.SendResult{
		Success: false,
		Error:   errors.New("SMTP failed"),
	}

	err := service.LogPasswordResetEmail("user@example.com", "Password Reset", result)

	if err != nil {
		t.Errorf("LogPasswordResetEmail should not return error: %v", err)
	}
	if repo.logs[0].ErrorMessage == nil {
		t.Error("Expected error message to be set")
	}
}

func TestEmailLogService_LogVerificationEmail_Success(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	result := email.SendResult{
		Success: true,
	}

	err := service.LogVerificationEmail("user@example.com", "Verify Email", result)

	if err != nil {
		t.Errorf("LogVerificationEmail should not return error: %v", err)
	}
	if repo.logs[0].EmailType != domain.EmailTypeVerification {
		t.Errorf("Expected type %s, got %s", domain.EmailTypeVerification, repo.logs[0].EmailType)
	}
}

func TestEmailLogService_LogVerificationEmail_WithError(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	result := email.SendResult{
		Success: false,
		Error:   errors.New("SMTP failed"),
	}

	err := service.LogVerificationEmail("user@example.com", "Verify Email", result)

	if err != nil {
		t.Errorf("LogVerificationEmail should not return error: %v", err)
	}
	if repo.logs[0].ErrorMessage == nil {
		t.Error("Expected error message to be set")
	}
}

func TestEmailLogService_LogNotificationEmail_Success(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	result := email.SendResult{
		Success: true,
	}

	err := service.LogNotificationEmail("user@example.com", "New PR!", result)

	if err != nil {
		t.Errorf("LogNotificationEmail should not return error: %v", err)
	}
	if repo.logs[0].EmailType != domain.EmailTypeNotification {
		t.Errorf("Expected type %s, got %s", domain.EmailTypeNotification, repo.logs[0].EmailType)
	}
}

func TestEmailLogService_LogNotificationEmail_WithError(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	result := email.SendResult{
		Success: false,
		Error:   errors.New("SMTP failed"),
	}

	err := service.LogNotificationEmail("user@example.com", "New PR!", result)

	if err != nil {
		t.Errorf("LogNotificationEmail should not return error: %v", err)
	}
	if repo.logs[0].ErrorMessage == nil {
		t.Error("Expected error message to be set")
	}
}

func TestEmailLogService_GetByID_Success(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Create a log first
	_ = service.LogEmailAttempt("test@example.com", domain.EmailTypeTest, "Test", true, nil, nil, nil)

	log, err := service.GetByID(1)

	if err != nil {
		t.Errorf("GetByID should not return error: %v", err)
	}
	if log == nil {
		t.Error("GetByID should return a log")
	}
	if log.ID != 1 {
		t.Errorf("Expected ID 1, got %d", log.ID)
	}
}

func TestEmailLogService_GetByID_NotFound(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	_, err := service.GetByID(999)

	if err == nil {
		t.Error("GetByID should return error when not found")
	}
}

func TestEmailLogService_List_Success(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Create some logs
	_ = service.LogEmailAttempt("test1@example.com", domain.EmailTypeTest, "Test 1", true, nil, nil, nil)
	_ = service.LogEmailAttempt("test2@example.com", domain.EmailTypePasswordReset, "Test 2", true, nil, nil, nil)
	_ = service.LogEmailAttempt("test3@example.com", domain.EmailTypeTest, "Test 3", false, nil, nil, nil)

	logs, err := service.List(domain.EmailLogFilters{}, 10, 0)

	if err != nil {
		t.Errorf("List should not return error: %v", err)
	}
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(logs))
	}
}

func TestEmailLogService_List_WithFilter(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Create some logs
	_ = service.LogEmailAttempt("test1@example.com", domain.EmailTypeTest, "Test 1", true, nil, nil, nil)
	_ = service.LogEmailAttempt("test2@example.com", domain.EmailTypePasswordReset, "Test 2", true, nil, nil, nil)

	emailType := domain.EmailTypeTest
	logs, err := service.List(domain.EmailLogFilters{EmailType: &emailType}, 10, 0)

	if err != nil {
		t.Errorf("List should not return error: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}
}

func TestEmailLogService_List_DefaultLimit(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Test that invalid limits are corrected
	logs, err := service.List(domain.EmailLogFilters{}, 0, 0)

	if err != nil {
		t.Errorf("List should not return error: %v", err)
	}
	// Should use default limit of 50
	if logs == nil {
		t.Error("List should return empty slice, not nil")
	}
}

func TestEmailLogService_List_InvalidLimit(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Test that limit > 100 is corrected
	_, err := service.List(domain.EmailLogFilters{}, 200, 0)

	if err != nil {
		t.Errorf("List should not return error: %v", err)
	}
}

func TestEmailLogService_List_NegativeOffset(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Test that negative offset is corrected
	_, err := service.List(domain.EmailLogFilters{}, 10, -5)

	if err != nil {
		t.Errorf("List should not return error: %v", err)
	}
}

func TestEmailLogService_Count_Success(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Create some logs
	_ = service.LogEmailAttempt("test1@example.com", domain.EmailTypeTest, "Test 1", true, nil, nil, nil)
	_ = service.LogEmailAttempt("test2@example.com", domain.EmailTypeTest, "Test 2", true, nil, nil, nil)

	count, err := service.Count(domain.EmailLogFilters{})

	if err != nil {
		t.Errorf("Count should not return error: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestEmailLogService_Count_WithFilter(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Create some logs
	_ = service.LogEmailAttempt("test1@example.com", domain.EmailTypeTest, "Test 1", true, nil, nil, nil)
	_ = service.LogEmailAttempt("test2@example.com", domain.EmailTypePasswordReset, "Test 2", true, nil, nil, nil)

	emailType := domain.EmailTypeTest
	count, err := service.Count(domain.EmailLogFilters{EmailType: &emailType})

	if err != nil {
		t.Errorf("Count should not return error: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestEmailLogService_CleanupOldLogs_Success(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Create a log with old timestamp
	log := &domain.EmailLog{
		RecipientEmail: "old@example.com",
		EmailType:      domain.EmailTypeTest,
		Subject:        "Old Email",
		Success:        true,
		CreatedAt:      time.Now().AddDate(0, 0, -100), // 100 days ago
	}
	log.ID = repo.nextID
	repo.nextID++
	repo.logs = append(repo.logs, log)

	// Create a recent log
	_ = service.LogEmailAttempt("new@example.com", domain.EmailTypeTest, "New Email", true, nil, nil, nil)

	deleted, err := service.CleanupOldLogs(90)

	if err != nil {
		t.Errorf("CleanupOldLogs should not return error: %v", err)
	}
	if deleted != 1 {
		t.Errorf("Expected 1 deleted, got %d", deleted)
	}
	if len(repo.logs) != 1 {
		t.Errorf("Expected 1 remaining log, got %d", len(repo.logs))
	}
}

func TestEmailLogService_CleanupOldLogs_DefaultRetention(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Test that 0 or negative retention uses default
	_, err := service.CleanupOldLogs(0)

	if err != nil {
		t.Errorf("CleanupOldLogs should not return error: %v", err)
	}
}

func TestEmailLogService_CleanupOldLogs_NegativeRetention(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Test that negative retention uses default
	_, err := service.CleanupOldLogs(-10)

	if err != nil {
		t.Errorf("CleanupOldLogs should not return error: %v", err)
	}
}

func TestEmailLogService_GetRecentFailures_Success(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Create some logs
	_ = service.LogEmailAttempt("success@example.com", domain.EmailTypeTest, "Success", true, nil, nil, nil)
	errMsg := "failed"
	_ = service.LogEmailAttempt("fail1@example.com", domain.EmailTypeTest, "Fail 1", false, &errMsg, nil, nil)
	_ = service.LogEmailAttempt("fail2@example.com", domain.EmailTypeTest, "Fail 2", false, &errMsg, nil, nil)

	failures, err := service.GetRecentFailures(10)

	if err != nil {
		t.Errorf("GetRecentFailures should not return error: %v", err)
	}
	if len(failures) != 2 {
		t.Errorf("Expected 2 failures, got %d", len(failures))
	}
}

func TestEmailLogService_GetRecentFailures_DefaultLimit(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Test that 0 limit uses default
	_, err := service.GetRecentFailures(0)

	if err != nil {
		t.Errorf("GetRecentFailures should not return error: %v", err)
	}
}

func TestEmailLogService_GetRecentFailures_MaxLimit(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Test that limit > 100 is corrected
	_, err := service.GetRecentFailures(200)

	if err != nil {
		t.Errorf("GetRecentFailures should not return error: %v", err)
	}
}

func TestEmailLogService_GetEmailStats_Success(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Create some logs
	_ = service.LogEmailAttempt("test1@example.com", domain.EmailTypeTest, "Test 1", true, nil, nil, nil)
	_ = service.LogEmailAttempt("test2@example.com", domain.EmailTypePasswordReset, "Test 2", true, nil, nil, nil)
	errMsg := "failed"
	_ = service.LogEmailAttempt("test3@example.com", domain.EmailTypeTest, "Test 3", false, &errMsg, nil, nil)

	stats, err := service.GetEmailStats()

	if err != nil {
		t.Errorf("GetEmailStats should not return error: %v", err)
	}
	if stats == nil {
		t.Error("GetEmailStats should return stats")
	}
	if stats.TotalCount != 3 {
		t.Errorf("Expected total count 3, got %d", stats.TotalCount)
	}
	if stats.SuccessCount != 2 {
		t.Errorf("Expected success count 2, got %d", stats.SuccessCount)
	}
	if stats.FailedCount != 1 {
		t.Errorf("Expected failed count 1, got %d", stats.FailedCount)
	}
}

func TestEmailLogService_GetEmailStats_SuccessRate(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	// Create logs: 3 success, 1 failure = 75% success rate
	_ = service.LogEmailAttempt("test1@example.com", domain.EmailTypeTest, "Test 1", true, nil, nil, nil)
	_ = service.LogEmailAttempt("test2@example.com", domain.EmailTypeTest, "Test 2", true, nil, nil, nil)
	_ = service.LogEmailAttempt("test3@example.com", domain.EmailTypeTest, "Test 3", true, nil, nil, nil)
	errMsg := "failed"
	_ = service.LogEmailAttempt("test4@example.com", domain.EmailTypeTest, "Test 4", false, &errMsg, nil, nil)

	stats, err := service.GetEmailStats()

	if err != nil {
		t.Errorf("GetEmailStats should not return error: %v", err)
	}
	if stats.SuccessRate != 75.0 {
		t.Errorf("Expected success rate 75.0, got %f", stats.SuccessRate)
	}
}

func TestEmailLogService_GetEmailStats_Empty(t *testing.T) {
	repo := newMockEmailLogRepo()
	service := NewEmailLogService(repo)

	stats, err := service.GetEmailStats()

	if err != nil {
		t.Errorf("GetEmailStats should not return error: %v", err)
	}
	if stats.TotalCount != 0 {
		t.Errorf("Expected total count 0, got %d", stats.TotalCount)
	}
	if stats.SuccessRate != 0.0 {
		t.Errorf("Expected success rate 0.0, got %f", stats.SuccessRate)
	}
}

func TestEmailLogService_GetEmailStats_CountError(t *testing.T) {
	repo := newMockEmailLogRepo()
	repo.countErr = errors.New("database error")
	service := NewEmailLogService(repo)

	_, err := service.GetEmailStats()

	if err == nil {
		t.Error("GetEmailStats should return error when count fails")
	}
}

func TestNewEmailStats(t *testing.T) {
	stats := NewEmailStats()

	if stats == nil {
		t.Error("NewEmailStats should not return nil")
	}
	if stats.CountByType == nil {
		t.Error("CountByType should be initialized")
	}
}

func TestEmailStats_Struct(t *testing.T) {
	stats := &EmailStats{
		TotalCount:   100,
		SuccessCount: 90,
		FailedCount:  10,
		SuccessRate:  90.0,
		CountByType:  map[string]int{"test": 50, "password_reset": 50},
	}

	if stats.TotalCount != 100 {
		t.Errorf("Expected TotalCount 100, got %d", stats.TotalCount)
	}
	if stats.SuccessRate != 90.0 {
		t.Errorf("Expected SuccessRate 90.0, got %f", stats.SuccessRate)
	}
	if stats.CountByType["test"] != 50 {
		t.Errorf("Expected test count 50, got %d", stats.CountByType["test"])
	}
}
