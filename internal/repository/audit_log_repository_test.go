package repository

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestAuditLogRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	auditRepo := NewAuditLogRepository(db, "sqlite3")

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

	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"
	details := `{"reason": "test"}`

	tests := []struct {
		name     string
		auditLog *domain.AuditLog
		wantErr  bool
	}{
		{
			name: "Create login success log",
			auditLog: &domain.AuditLog{
				UserID:    &user.ID,
				EventType: domain.EventLoginSuccess,
				IPAddress: &ipAddress,
				UserAgent: &userAgent,
			},
			wantErr: false,
		},
		{
			name: "Create login failed log with target user",
			auditLog: &domain.AuditLog{
				UserID:       &user.ID,
				TargetUserID: &user.ID,
				EventType:    domain.EventLoginFailed,
				IPAddress:    &ipAddress,
				Details:      &details,
			},
			wantErr: false,
		},
		{
			name: "Create system event (no user)",
			auditLog: &domain.AuditLog{
				EventType: domain.EventAccountLockedAuto,
				Details:   &details,
			},
			wantErr: false,
		},
		{
			name: "Create minimal log",
			auditLog: &domain.AuditLog{
				EventType: domain.EventLogout,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditRepo.Create(tt.auditLog)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.auditLog.ID == 0 {
				t.Error("Create() should set ID")
			}
			if !tt.wantErr && tt.auditLog.CreatedAt.IsZero() {
				t.Error("Create() should set CreatedAt")
			}
		})
	}
}

func TestAuditLogRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	auditRepo := NewAuditLogRepository(db, "sqlite3")

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

	// Create audit log
	ipAddress := "192.168.1.1"
	auditLog := &domain.AuditLog{
		UserID:    &user.ID,
		EventType: domain.EventLoginSuccess,
		IPAddress: &ipAddress,
	}
	if err := auditRepo.Create(auditLog); err != nil {
		t.Fatalf("Failed to create audit log: %v", err)
	}

	// Test GetByID
	got, err := auditRepo.GetByID(auditLog.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByID() returned nil")
	}
	if got.ID != auditLog.ID {
		t.Errorf("ID = %d, want %d", got.ID, auditLog.ID)
	}
	if got.EventType != domain.EventLoginSuccess {
		t.Errorf("EventType = %v, want %v", got.EventType, domain.EventLoginSuccess)
	}
	if got.UserEmail == nil || *got.UserEmail != "test@example.com" {
		t.Error("UserEmail should be populated from JOIN")
	}

	// Test non-existent log
	_, err = auditRepo.GetByID(999)
	if err == nil {
		t.Error("GetByID(999) should return error")
	}
}

func TestAuditLogRepository_List(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	auditRepo := NewAuditLogRepository(db, "sqlite3")

	// Create test users
	user1 := &domain.User{
		Email:        "user1@example.com",
		PasswordHash: "hashedpassword",
		Name:         "User 1",
		Role:         "athlete",
	}
	if err := userRepo.Create(user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2 := &domain.User{
		Email:        "user2@example.com",
		PasswordHash: "hashedpassword",
		Name:         "User 2",
		Role:         "admin",
	}
	if err := userRepo.Create(user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Create audit logs using direct SQL to avoid double-insert issue in Create()
	ip1 := "192.168.1.1"
	ip2 := "10.0.0.1"
	now := time.Now()

	logsData := []struct {
		userID       *int64
		targetUserID *int64
		eventType    string
		ipAddress    *string
	}{
		{&user1.ID, nil, domain.EventLoginSuccess, &ip1},
		{&user1.ID, nil, domain.EventLoginSuccess, &ip1},
		{&user1.ID, nil, domain.EventLogout, &ip1},
		{&user2.ID, &user1.ID, domain.EventAccountDisabled, &ip2},
		{nil, nil, domain.EventAccountLockedAuto, nil},
	}

	for _, l := range logsData {
		_, err := db.Exec(`INSERT INTO audit_logs (user_id, target_user_id, event_type, ip_address, created_at) VALUES (?, ?, ?, ?, ?)`,
			l.userID, l.targetUserID, l.eventType, l.ipAddress, now)
		if err != nil {
			t.Fatalf("Failed to create audit log: %v", err)
		}
	}

	eventTypeLogin := domain.EventLoginSuccess
	eventTypeLogout := domain.EventLogout

	tests := []struct {
		name      string
		filters   domain.AuditLogFilters
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "List all",
			filters:   domain.AuditLogFilters{},
			limit:     100,
			offset:    0,
			wantCount: 5,
		},
		{
			name:      "Filter by user ID",
			filters:   domain.AuditLogFilters{UserID: &user1.ID},
			limit:     100,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "Filter by target user ID",
			filters:   domain.AuditLogFilters{TargetUserID: &user1.ID},
			limit:     100,
			offset:    0,
			wantCount: 1,
		},
		{
			name:      "Filter by event type - login",
			filters:   domain.AuditLogFilters{EventType: &eventTypeLogin},
			limit:     100,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "Filter by event type - logout",
			filters:   domain.AuditLogFilters{EventType: &eventTypeLogout},
			limit:     100,
			offset:    0,
			wantCount: 1,
		},
		{
			name:      "Filter by IP address",
			filters:   domain.AuditLogFilters{IPAddress: &ip1},
			limit:     100,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "With limit",
			filters:   domain.AuditLogFilters{},
			limit:     2,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "With offset",
			filters:   domain.AuditLogFilters{},
			limit:     100,
			offset:    3,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := auditRepo.List(tt.filters, tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("List() returned %d logs, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestAuditLogRepository_List_DateFilters(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	auditRepo := NewAuditLogRepository(db, "sqlite3")

	// Create logs at different times using direct SQL to control timestamps
	now := time.Now()
	times := []time.Time{
		now.AddDate(0, 0, -10), // 10 days ago
		now.AddDate(0, 0, -5),  // 5 days ago
		now.AddDate(0, 0, -3),  // 3 days ago
		now.AddDate(0, 0, -1),  // 1 day ago
		now,                    // today
	}

	for _, t := range times {
		_, err := db.Exec(`INSERT INTO audit_logs (event_type, created_at) VALUES (?, ?)`,
			domain.EventLoginSuccess, t)
		if err != nil {
			panic("Failed to create audit log: " + err.Error())
		}
	}

	// Test date range filter
	startDate := now.AddDate(0, 0, -4) // 4 days ago
	endDate := now.AddDate(0, 0, 1)    // tomorrow

	filters := domain.AuditLogFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	got, err := auditRepo.List(filters, 100, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Should get logs from -3, -1, and today (3 logs)
	if len(got) != 3 {
		t.Errorf("List() with date filter returned %d logs, want 3", len(got))
	}
}

func TestAuditLogRepository_Count(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	auditRepo := NewAuditLogRepository(db, "sqlite3")

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

	// Create audit logs using direct SQL to avoid double-insert issue
	eventTypes := []string{
		domain.EventLoginSuccess,
		domain.EventLoginSuccess,
		domain.EventLoginFailed,
		domain.EventLogout,
		domain.EventPasswordChanged,
	}
	now := time.Now()

	for _, et := range eventTypes {
		_, err := db.Exec(`INSERT INTO audit_logs (user_id, event_type, created_at) VALUES (?, ?, ?)`,
			user.ID, et, now)
		if err != nil {
			t.Fatalf("Failed to create audit log: %v", err)
		}
	}

	eventTypeLogin := domain.EventLoginSuccess

	tests := []struct {
		name      string
		filters   domain.AuditLogFilters
		wantCount int
	}{
		{
			name:      "Count all",
			filters:   domain.AuditLogFilters{},
			wantCount: 5,
		},
		{
			name:      "Count by user ID",
			filters:   domain.AuditLogFilters{UserID: &user.ID},
			wantCount: 5,
		},
		{
			name:      "Count by event type",
			filters:   domain.AuditLogFilters{EventType: &eventTypeLogin},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := auditRepo.Count(tt.filters)
			if err != nil {
				t.Errorf("Count() error = %v", err)
				return
			}
			if got != tt.wantCount {
				t.Errorf("Count() = %d, want %d", got, tt.wantCount)
			}
		})
	}
}

func TestAuditLogRepository_GetByUserID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	auditRepo := NewAuditLogRepository(db, "sqlite3")

	// Create test users
	user1 := &domain.User{
		Email:        "user1@example.com",
		PasswordHash: "hashedpassword",
		Name:         "User 1",
		Role:         "athlete",
	}
	if err := userRepo.Create(user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2 := &domain.User{
		Email:        "user2@example.com",
		PasswordHash: "hashedpassword",
		Name:         "User 2",
		Role:         "athlete",
	}
	if err := userRepo.Create(user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	now := time.Now()

	// Create logs for user1 using direct SQL to avoid double-insert issue
	for i := 0; i < 5; i++ {
		_, err := db.Exec(`INSERT INTO audit_logs (user_id, event_type, created_at) VALUES (?, ?, ?)`,
			user1.ID, domain.EventLoginSuccess, now)
		if err != nil {
			t.Fatalf("Failed to create audit log: %v", err)
		}
	}

	// Create logs for user2
	for i := 0; i < 3; i++ {
		_, err := db.Exec(`INSERT INTO audit_logs (user_id, event_type, created_at) VALUES (?, ?, ?)`,
			user2.ID, domain.EventLoginSuccess, now)
		if err != nil {
			t.Fatalf("Failed to create audit log: %v", err)
		}
	}

	// Test GetByUserID
	got, err := auditRepo.GetByUserID(user1.ID, 100, 0)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if len(got) != 5 {
		t.Errorf("GetByUserID() returned %d logs, want 5", len(got))
	}

	// Test with limit
	got, err = auditRepo.GetByUserID(user1.ID, 2, 0)
	if err != nil {
		t.Fatalf("GetByUserID() with limit error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetByUserID() with limit returned %d logs, want 2", len(got))
	}

	// Test non-existent user
	got, err = auditRepo.GetByUserID(999, 100, 0)
	if err != nil {
		t.Fatalf("GetByUserID(999) error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("GetByUserID(999) returned %d logs, want 0", len(got))
	}
}

func TestAuditLogRepository_GetByTargetUserID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	auditRepo := NewAuditLogRepository(db, "sqlite3")

	// Create test users
	admin := &domain.User{
		Email:        "admin@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Admin",
		Role:         "admin",
	}
	if err := userRepo.Create(admin); err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	targetUser := &domain.User{
		Email:        "target@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Target User",
		Role:         "athlete",
	}
	if err := userRepo.Create(targetUser); err != nil {
		t.Fatalf("Failed to create target user: %v", err)
	}

	now := time.Now()

	// Create logs affecting target user using direct SQL to avoid double-insert issue
	affectingEvents := []string{
		domain.EventAccountDisabled,
		domain.EventAccountEnabled,
		domain.EventPasswordReset,
	}

	for _, et := range affectingEvents {
		_, err := db.Exec(`INSERT INTO audit_logs (user_id, target_user_id, event_type, created_at) VALUES (?, ?, ?, ?)`,
			admin.ID, targetUser.ID, et, now)
		if err != nil {
			t.Fatalf("Failed to create audit log: %v", err)
		}
	}

	// Create log not affecting target user
	_, err = db.Exec(`INSERT INTO audit_logs (user_id, event_type, created_at) VALUES (?, ?, ?)`,
		admin.ID, domain.EventLoginSuccess, now)
	if err != nil {
		t.Fatalf("Failed to create audit log: %v", err)
	}

	// Test GetByTargetUserID
	got, err := auditRepo.GetByTargetUserID(targetUser.ID, 100, 0)
	if err != nil {
		t.Fatalf("GetByTargetUserID() error = %v", err)
	}
	if len(got) != 3 {
		t.Errorf("GetByTargetUserID() returned %d logs, want 3", len(got))
	}

	// Verify all logs have correct target user
	for _, l := range got {
		if l.TargetUserID == nil || *l.TargetUserID != targetUser.ID {
			t.Error("GetByTargetUserID() returned log with wrong target user")
		}
	}
}

func TestAuditLogRepository_DeleteOlderThan(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	auditRepo := NewAuditLogRepository(db, "sqlite3")

	// Create logs at different times using direct SQL
	now := time.Now()
	times := []time.Time{
		now.AddDate(0, 0, -30), // 30 days ago (old)
		now.AddDate(0, 0, -20), // 20 days ago (old)
		now.AddDate(0, 0, -10), // 10 days ago (old)
		now.AddDate(0, 0, -5),  // 5 days ago (recent)
		now,                    // today (recent)
	}

	for _, t := range times {
		_, err := db.Exec(`INSERT INTO audit_logs (event_type, created_at) VALUES (?, ?)`,
			domain.EventLoginSuccess, t)
		if err != nil {
			panic("Failed to create audit log: " + err.Error())
		}
	}

	// Verify we have 5 logs
	count, err := auditRepo.Count(domain.AuditLogFilters{})
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 5 {
		t.Fatalf("Expected 5 logs, got %d", count)
	}

	// Delete logs older than 7 days ago
	cutoff := now.AddDate(0, 0, -7)
	deleted, err := auditRepo.DeleteOlderThan(cutoff)
	if err != nil {
		t.Fatalf("DeleteOlderThan() error = %v", err)
	}
	if deleted != 3 {
		t.Errorf("DeleteOlderThan() deleted %d logs, want 3", deleted)
	}

	// Verify remaining count
	count, err = auditRepo.Count(domain.AuditLogFilters{})
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 2 {
		t.Errorf("After deletion, expected 2 logs, got %d", count)
	}

	// Delete with no matches
	deleted, err = auditRepo.DeleteOlderThan(now.AddDate(-1, 0, 0)) // 1 year ago
	if err != nil {
		t.Fatalf("DeleteOlderThan() error = %v", err)
	}
	if deleted != 0 {
		t.Errorf("DeleteOlderThan() with no matches deleted %d logs, want 0", deleted)
	}
}

func TestAuditLogRepository_ListOrderByCreatedAt(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	auditRepo := NewAuditLogRepository(db, "sqlite3")

	// Create logs at different times
	now := time.Now()
	for i := 0; i < 5; i++ {
		timestamp := now.Add(time.Duration(-i) * time.Hour)
		_, err := db.Exec(`INSERT INTO audit_logs (event_type, created_at) VALUES (?, ?)`,
			domain.EventLoginSuccess, timestamp)
		if err != nil {
			t.Fatalf("Failed to create audit log: %v", err)
		}
	}

	// List should return most recent first
	got, err := auditRepo.List(domain.AuditLogFilters{}, 100, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Verify order (most recent first)
	for i := 1; i < len(got); i++ {
		if got[i].CreatedAt.After(got[i-1].CreatedAt) {
			t.Error("List() should return logs in descending order by created_at")
		}
	}
}

func TestAuditLogRepository_Create_UnsupportedDriver(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	// Create repo with unsupported driver
	auditRepo := NewAuditLogRepository(db, "unsupported_driver")

	log := &domain.AuditLog{
		EventType: domain.EventLoginSuccess,
	}

	err = auditRepo.Create(log)
	if err == nil {
		t.Error("Create() with unsupported driver should return error")
	}
	if err != nil && !contains(err.Error(), "unsupported database driver") {
		t.Errorf("Create() error = %v, want error containing 'unsupported database driver'", err)
	}
}

func TestAuditLogRepository_Count_AllFilters(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	auditRepo := NewAuditLogRepository(db, "sqlite3")

	// Create test users
	user1 := &domain.User{
		Email:        "user1@example.com",
		PasswordHash: "hashedpassword",
		Name:         "User 1",
		Role:         "athlete",
	}
	if err := userRepo.Create(user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2 := &domain.User{
		Email:        "user2@example.com",
		PasswordHash: "hashedpassword",
		Name:         "User 2",
		Role:         "admin",
	}
	if err := userRepo.Create(user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	ip1 := "192.168.1.1"
	ip2 := "10.0.0.1"

	// Create audit logs with various combinations
	logsData := []struct {
		userID       *int64
		targetUserID *int64
		eventType    string
		ipAddress    *string
		createdAt    time.Time
	}{
		{&user1.ID, nil, domain.EventLoginSuccess, &ip1, now},
		{&user1.ID, nil, domain.EventLoginSuccess, &ip1, now},
		{&user1.ID, nil, domain.EventLogout, &ip1, yesterday},
		{&user2.ID, &user1.ID, domain.EventAccountDisabled, &ip2, now},
		{nil, nil, domain.EventAccountLockedAuto, nil, now},
	}

	for _, l := range logsData {
		_, err := db.Exec(`INSERT INTO audit_logs (user_id, target_user_id, event_type, ip_address, created_at) VALUES (?, ?, ?, ?, ?)`,
			l.userID, l.targetUserID, l.eventType, l.ipAddress, l.createdAt)
		if err != nil {
			t.Fatalf("Failed to create audit log: %v", err)
		}
	}

	eventTypeLogin := domain.EventLoginSuccess

	tests := []struct {
		name      string
		filters   domain.AuditLogFilters
		wantCount int
	}{
		{
			name:      "Count by target user ID",
			filters:   domain.AuditLogFilters{TargetUserID: &user1.ID},
			wantCount: 1,
		},
		{
			name:      "Count by IP address",
			filters:   domain.AuditLogFilters{IPAddress: &ip1},
			wantCount: 3,
		},
		{
			name:      "Count by date range",
			filters:   domain.AuditLogFilters{StartDate: &yesterday, EndDate: &tomorrow},
			wantCount: 5,
		},
		{
			name:      "Count with multiple filters",
			filters:   domain.AuditLogFilters{UserID: &user1.ID, EventType: &eventTypeLogin, IPAddress: &ip1},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := auditRepo.Count(tt.filters)
			if err != nil {
				t.Errorf("Count() error = %v", err)
				return
			}
			if got != tt.wantCount {
				t.Errorf("Count() = %d, want %d", got, tt.wantCount)
			}
		})
	}
}

func TestAuditLogRepository_List_AllFilters(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	auditRepo := NewAuditLogRepository(db, "sqlite3")

	// Create test users
	user1 := &domain.User{
		Email:        "user1@example.com",
		PasswordHash: "hashedpassword",
		Name:         "User 1",
		Role:         "athlete",
	}
	if err := userRepo.Create(user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2 := &domain.User{
		Email:        "user2@example.com",
		PasswordHash: "hashedpassword",
		Name:         "User 2",
		Role:         "admin",
	}
	if err := userRepo.Create(user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	ip1 := "192.168.1.1"

	// Create audit log with all fields
	_, err = db.Exec(`INSERT INTO audit_logs (user_id, target_user_id, event_type, ip_address, created_at) VALUES (?, ?, ?, ?, ?)`,
		user1.ID, user2.ID, domain.EventLoginSuccess, ip1, now)
	if err != nil {
		t.Fatalf("Failed to create audit log: %v", err)
	}

	// Test combined filters
	filters := domain.AuditLogFilters{
		UserID:       &user1.ID,
		TargetUserID: &user2.ID,
		IPAddress:    &ip1,
		StartDate:    &yesterday,
		EndDate:      &tomorrow,
	}

	got, err := auditRepo.List(filters, 100, 0)
	if err != nil {
		t.Fatalf("List() with all filters error = %v", err)
	}
	if len(got) != 1 {
		t.Errorf("List() with all filters returned %d logs, want 1", len(got))
	}
}

// helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
