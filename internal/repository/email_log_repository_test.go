package repository

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestEmailLogRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create a user first for sent_by_user_id
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "emailloguser@example.com")

	repo := NewEmailLogRepository(db, "sqlite3")

	tests := []struct {
		name    string
		log     *domain.EmailLog
		wantErr bool
	}{
		{
			name: "create test email log",
			log: &domain.EmailLog{
				RecipientEmail: "recipient@example.com",
				EmailType:      domain.EmailTypeTest,
				Subject:        "Test Email",
				Success:        true,
				SentByUserID:   &user.ID,
			},
			wantErr: false,
		},
		{
			name: "create password reset email log",
			log: &domain.EmailLog{
				RecipientEmail: "reset@example.com",
				EmailType:      domain.EmailTypePasswordReset,
				Subject:        "Password Reset Request",
				Success:        true,
			},
			wantErr: false,
		},
		{
			name: "create failed email log with error",
			log: &domain.EmailLog{
				RecipientEmail: "failed@example.com",
				EmailType:      domain.EmailTypeNotification,
				Subject:        "Notification",
				Success:        false,
				ErrorMessage:   strPtr("SMTP connection failed"),
				DebugInfo:      strPtr(`{"phase": "connection", "error": "timeout"}`),
			},
			wantErr: false,
		},
		{
			name: "create verification email log",
			log: &domain.EmailLog{
				RecipientEmail: "verify@example.com",
				EmailType:      domain.EmailTypeVerification,
				Subject:        "Verify Your Email",
				Success:        true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.log)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.log.ID == 0 {
				t.Error("Create() should set log ID")
			}
			if !tt.wantErr && tt.log.CreatedAt.IsZero() {
				t.Error("Create() should set CreatedAt")
			}
		})
	}
}

func TestEmailLogRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create a user for sent_by_user_id
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "getbyiduser@example.com")

	repo := NewEmailLogRepository(db, "sqlite3")

	// Create a test email log
	log := &domain.EmailLog{
		RecipientEmail: "test@example.com",
		EmailType:      domain.EmailTypeTest,
		Subject:        "Test Email Subject",
		Success:        true,
		SentByUserID:   &user.ID,
	}
	if err := repo.Create(log); err != nil {
		t.Fatalf("failed to create test email log: %v", err)
	}

	tests := []struct {
		name        string
		id          int64
		wantLog     bool
		wantSubject string
		wantErr     bool
	}{
		{
			name:        "existing email log",
			id:          log.ID,
			wantLog:     true,
			wantSubject: "Test Email Subject",
			wantErr:     false,
		},
		{
			name:    "non-existent email log",
			id:      99999,
			wantLog: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantLog && got == nil {
				t.Error("GetByID() expected log but got nil")
				return
			}
			if !tt.wantLog && got != nil {
				t.Errorf("GetByID() expected nil but got log: %+v", got)
				return
			}
			if tt.wantLog && got.Subject != tt.wantSubject {
				t.Errorf("GetByID() subject = %v, want %v", got.Subject, tt.wantSubject)
			}
			// Check that sent_by_user_email is populated
			if tt.wantLog && got.SentByUserEmail == nil {
				t.Error("GetByID() should populate SentByUserEmail")
			}
		})
	}
}

func TestEmailLogRepository_List(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create a user for sent_by_user_id
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "listuser@example.com")

	repo := NewEmailLogRepository(db, "sqlite3")

	// Create test email logs
	testLogs := []*domain.EmailLog{
		{
			RecipientEmail: "user1@example.com",
			EmailType:      domain.EmailTypeTest,
			Subject:        "Test Email 1",
			Success:        true,
			SentByUserID:   &user.ID,
		},
		{
			RecipientEmail: "user2@example.com",
			EmailType:      domain.EmailTypePasswordReset,
			Subject:        "Password Reset",
			Success:        true,
		},
		{
			RecipientEmail: "user3@example.com",
			EmailType:      domain.EmailTypeNotification,
			Subject:        "Notification",
			Success:        false,
			ErrorMessage:   strPtr("Failed to send"),
		},
		{
			RecipientEmail: "user1@example.com",
			EmailType:      domain.EmailTypeVerification,
			Subject:        "Verify Email",
			Success:        true,
		},
	}

	for _, log := range testLogs {
		if err := repo.Create(log); err != nil {
			t.Fatalf("failed to create test email log: %v", err)
		}
		// Add small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	tests := []struct {
		name      string
		filters   domain.EmailLogFilters
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "list all with no filters",
			filters:   domain.EmailLogFilters{},
			limit:     10,
			offset:    0,
			wantCount: 4,
		},
		{
			name: "filter by email type",
			filters: domain.EmailLogFilters{
				EmailType: strPtr(domain.EmailTypeTest),
			},
			limit:     10,
			offset:    0,
			wantCount: 1,
		},
		{
			name: "filter by success status true",
			filters: domain.EmailLogFilters{
				Success: boolPtr(true),
			},
			limit:     10,
			offset:    0,
			wantCount: 3,
		},
		{
			name: "filter by success status false",
			filters: domain.EmailLogFilters{
				Success: boolPtr(false),
			},
			limit:     10,
			offset:    0,
			wantCount: 1,
		},
		{
			name: "filter by recipient email partial match",
			filters: domain.EmailLogFilters{
				RecipientEmail: strPtr("user1"),
			},
			limit:     10,
			offset:    0,
			wantCount: 2,
		},
		{
			name: "filter by sent_by_user_id",
			filters: domain.EmailLogFilters{
				SentByUserID: &user.ID,
			},
			limit:     10,
			offset:    0,
			wantCount: 1,
		},
		{
			name:      "pagination - first page",
			filters:   domain.EmailLogFilters{},
			limit:     2,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "pagination - second page",
			filters:   domain.EmailLogFilters{},
			limit:     2,
			offset:    2,
			wantCount: 2,
		},
		{
			name:      "pagination - beyond data",
			filters:   domain.EmailLogFilters{},
			limit:     10,
			offset:    100,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, err := repo.List(tt.filters, tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}
			if len(logs) != tt.wantCount {
				t.Errorf("List() returned %d logs, want %d", len(logs), tt.wantCount)
			}
		})
	}
}

func TestEmailLogRepository_List_DateFilters(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewEmailLogRepository(db, "sqlite3")

	// Create logs at different times
	now := time.Now()

	log1 := &domain.EmailLog{
		RecipientEmail: "old@example.com",
		EmailType:      domain.EmailTypeTest,
		Subject:        "Old Email",
		Success:        true,
	}
	if err := repo.Create(log1); err != nil {
		t.Fatalf("failed to create log1: %v", err)
	}

	// Test start date filter (logs after a certain date)
	startDate := now.Add(-1 * time.Hour)
	filters := domain.EmailLogFilters{
		StartDate: &startDate,
	}
	logs, err := repo.List(filters, 10, 0)
	if err != nil {
		t.Errorf("List with StartDate filter error = %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("List with StartDate filter returned %d logs, want 1", len(logs))
	}

	// Test end date filter (logs before a certain date)
	endDate := now.Add(1 * time.Hour)
	filters = domain.EmailLogFilters{
		EndDate: &endDate,
	}
	logs, err = repo.List(filters, 10, 0)
	if err != nil {
		t.Errorf("List with EndDate filter error = %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("List with EndDate filter returned %d logs, want 1", len(logs))
	}

	// Test combined date filters
	filters = domain.EmailLogFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	logs, err = repo.List(filters, 10, 0)
	if err != nil {
		t.Errorf("List with both date filters error = %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("List with both date filters returned %d logs, want 1", len(logs))
	}
}

func TestEmailLogRepository_Count(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	// Create a user for sent_by_user_id
	userRepo := NewSQLiteUserRepository(db)
	user := createTestUser(t, userRepo, "countuser@example.com")

	repo := NewEmailLogRepository(db, "sqlite3")

	// Create test email logs
	testLogs := []*domain.EmailLog{
		{
			RecipientEmail: "count1@example.com",
			EmailType:      domain.EmailTypeTest,
			Subject:        "Test 1",
			Success:        true,
			SentByUserID:   &user.ID,
		},
		{
			RecipientEmail: "count2@example.com",
			EmailType:      domain.EmailTypePasswordReset,
			Subject:        "Password Reset",
			Success:        true,
		},
		{
			RecipientEmail: "count3@example.com",
			EmailType:      domain.EmailTypeTest,
			Subject:        "Test 2",
			Success:        false,
		},
	}

	for _, log := range testLogs {
		if err := repo.Create(log); err != nil {
			t.Fatalf("failed to create test email log: %v", err)
		}
	}

	tests := []struct {
		name      string
		filters   domain.EmailLogFilters
		wantCount int
	}{
		{
			name:      "count all",
			filters:   domain.EmailLogFilters{},
			wantCount: 3,
		},
		{
			name: "count by email type",
			filters: domain.EmailLogFilters{
				EmailType: strPtr(domain.EmailTypeTest),
			},
			wantCount: 2,
		},
		{
			name: "count successful only",
			filters: domain.EmailLogFilters{
				Success: boolPtr(true),
			},
			wantCount: 2,
		},
		{
			name: "count failed only",
			filters: domain.EmailLogFilters{
				Success: boolPtr(false),
			},
			wantCount: 1,
		},
		{
			name: "count by recipient email",
			filters: domain.EmailLogFilters{
				RecipientEmail: strPtr("count1"),
			},
			wantCount: 1,
		},
		{
			name: "count by sent_by_user_id",
			filters: domain.EmailLogFilters{
				SentByUserID: &user.ID,
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := repo.Count(tt.filters)
			if err != nil {
				t.Errorf("Count() error = %v", err)
				return
			}
			if count != tt.wantCount {
				t.Errorf("Count() = %d, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestEmailLogRepository_Count_DateFilters(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewEmailLogRepository(db, "sqlite3")

	// Create a log
	log := &domain.EmailLog{
		RecipientEmail: "datecount@example.com",
		EmailType:      domain.EmailTypeTest,
		Subject:        "Date Count Test",
		Success:        true,
	}
	if err := repo.Create(log); err != nil {
		t.Fatalf("failed to create log: %v", err)
	}

	now := time.Now()

	// Test start date filter
	startDate := now.Add(-1 * time.Hour)
	filters := domain.EmailLogFilters{
		StartDate: &startDate,
	}
	count, err := repo.Count(filters)
	if err != nil {
		t.Errorf("Count with StartDate filter error = %v", err)
	}
	if count != 1 {
		t.Errorf("Count with StartDate filter = %d, want 1", count)
	}

	// Test end date filter
	endDate := now.Add(1 * time.Hour)
	filters = domain.EmailLogFilters{
		EndDate: &endDate,
	}
	count, err = repo.Count(filters)
	if err != nil {
		t.Errorf("Count with EndDate filter error = %v", err)
	}
	if count != 1 {
		t.Errorf("Count with EndDate filter = %d, want 1", count)
	}
}

func TestEmailLogRepository_DeleteOlderThan(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewEmailLogRepository(db, "sqlite3")

	// Create test email logs
	log1 := &domain.EmailLog{
		RecipientEmail: "old1@example.com",
		EmailType:      domain.EmailTypeTest,
		Subject:        "Old Email 1",
		Success:        true,
	}
	if err := repo.Create(log1); err != nil {
		t.Fatalf("failed to create log1: %v", err)
	}

	log2 := &domain.EmailLog{
		RecipientEmail: "old2@example.com",
		EmailType:      domain.EmailTypePasswordReset,
		Subject:        "Old Email 2",
		Success:        true,
	}
	if err := repo.Create(log2); err != nil {
		t.Fatalf("failed to create log2: %v", err)
	}

	// Verify we have 2 logs
	count, err := repo.Count(domain.EmailLogFilters{})
	if err != nil {
		t.Fatalf("failed to count logs: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 logs, got %d", count)
	}

	// Delete logs older than 1 hour from now (should delete nothing)
	futureTime := time.Now().Add(-1 * time.Hour)
	deleted, err := repo.DeleteOlderThan(futureTime)
	if err != nil {
		t.Errorf("DeleteOlderThan() error = %v", err)
	}
	if deleted != 0 {
		t.Errorf("DeleteOlderThan() deleted %d logs, want 0", deleted)
	}

	// Verify no logs were deleted
	count, _ = repo.Count(domain.EmailLogFilters{})
	if count != 2 {
		t.Errorf("expected 2 logs after delete, got %d", count)
	}

	// Delete logs older than 1 hour in the future (should delete all)
	pastTime := time.Now().Add(1 * time.Hour)
	deleted, err = repo.DeleteOlderThan(pastTime)
	if err != nil {
		t.Errorf("DeleteOlderThan() error = %v", err)
	}
	if deleted != 2 {
		t.Errorf("DeleteOlderThan() deleted %d logs, want 2", deleted)
	}

	// Verify all logs were deleted
	count, _ = repo.Count(domain.EmailLogFilters{})
	if count != 0 {
		t.Errorf("expected 0 logs after delete, got %d", count)
	}
}

func TestNewEmailLogRepository(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	repo := NewEmailLogRepository(db, "sqlite3")
	if repo == nil {
		t.Error("NewEmailLogRepository() returned nil")
	}
	if repo.db != db {
		t.Error("NewEmailLogRepository() db not set correctly")
	}
	if repo.driver != "sqlite3" {
		t.Errorf("NewEmailLogRepository() driver = %v, want sqlite3", repo.driver)
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
