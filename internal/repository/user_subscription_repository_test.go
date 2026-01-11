package repository

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestUserSubscriptionRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	subRepo := NewSQLiteUserSubscriptionRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	now := time.Now()
	endDate := now.AddDate(0, 1, 0)
	sub := &domain.UserSubscription{
		UserID:           user.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now,
		EndDate:          &endDate,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err = subRepo.Create(sub)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if sub.ID == 0 {
		t.Error("Create() should set ID")
	}
}

func TestUserSubscriptionRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	subRepo := NewSQLiteUserSubscriptionRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	now := time.Now()
	endDate := now.AddDate(0, 1, 0)
	notes := "Test subscription"
	sub := &domain.UserSubscription{
		UserID:           user.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now,
		EndDate:          &endDate,
		Notes:            &notes,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := subRepo.Create(sub); err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Get by ID
	got, err := subRepo.GetByID(sub.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByID() returned nil")
	}
	if got.UserID != user.ID {
		t.Errorf("UserID = %d, want %d", got.UserID, user.ID)
	}
	if got.SubscriptionType != domain.SubscriptionTypeMonthly {
		t.Errorf("SubscriptionType = %v, want monthly", got.SubscriptionType)
	}
	if got.Status != domain.SubscriptionStatusActive {
		t.Errorf("Status = %v, want active", got.Status)
	}
	if got.Notes == nil || *got.Notes != notes {
		t.Errorf("Notes = %v, want %v", got.Notes, notes)
	}

	// Non-existent
	got, err = subRepo.GetByID(999)
	if err != nil {
		t.Fatalf("GetByID(999) error = %v", err)
	}
	if got != nil {
		t.Error("GetByID(999) should return nil")
	}
}

func TestUserSubscriptionRepository_GetActiveByUserID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	subRepo := NewSQLiteUserSubscriptionRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test getting non-existent subscription
	got, err := subRepo.GetActiveByUserID(user.ID)
	if err != nil {
		t.Fatalf("GetActiveByUserID() error = %v", err)
	}
	if got != nil {
		t.Error("GetActiveByUserID() should return nil for user without subscription")
	}

	// Create active subscription
	now := time.Now()
	endDate := now.AddDate(0, 1, 0)
	activeSub := &domain.UserSubscription{
		UserID:           user.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now,
		EndDate:          &endDate,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := subRepo.Create(activeSub); err != nil {
		t.Fatalf("Failed to create active subscription: %v", err)
	}

	// Get active subscription
	got, err = subRepo.GetActiveByUserID(user.ID)
	if err != nil {
		t.Fatalf("GetActiveByUserID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetActiveByUserID() returned nil for user with active subscription")
	}
	if got.ID != activeSub.ID {
		t.Errorf("ID = %d, want %d", got.ID, activeSub.ID)
	}
}

func TestUserSubscriptionRepository_GetByUserID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	subRepo := NewSQLiteUserSubscriptionRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	now := time.Now()

	// Create multiple subscriptions (history)
	statuses := []domain.SubscriptionStatus{
		domain.SubscriptionStatusCancelled,
		domain.SubscriptionStatusExpired,
		domain.SubscriptionStatusActive,
	}

	for i, status := range statuses {
		endDate := now.AddDate(0, i+1, 0)
		sub := &domain.UserSubscription{
			UserID:           user.ID,
			SubscriptionType: domain.SubscriptionTypeMonthly,
			Status:           status,
			StartDate:        now.AddDate(0, -i, 0),
			EndDate:          &endDate,
			CreatedAt:        now.Add(time.Duration(i) * time.Hour),
			UpdatedAt:        now.Add(time.Duration(i) * time.Hour),
		}
		if err := subRepo.Create(sub); err != nil {
			t.Fatalf("Failed to create subscription: %v", err)
		}
	}

	// Get all subscriptions for user
	subs, err := subRepo.GetByUserID(user.ID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if len(subs) != 3 {
		t.Errorf("GetByUserID() returned %d subscriptions, want 3", len(subs))
	}

	// Non-existent user
	subs, err = subRepo.GetByUserID(999)
	if err != nil {
		t.Fatalf("GetByUserID(999) error = %v", err)
	}
	if len(subs) != 0 {
		t.Errorf("GetByUserID(999) returned %d subscriptions, want 0", len(subs))
	}
}

func TestUserSubscriptionRepository_Update(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	subRepo := NewSQLiteUserSubscriptionRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	now := time.Now()
	endDate := now.AddDate(0, 1, 0)
	sub := &domain.UserSubscription{
		UserID:           user.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now,
		EndDate:          &endDate,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := subRepo.Create(sub); err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Update
	sub.SubscriptionType = domain.SubscriptionTypeAnnual
	newEndDate := now.AddDate(1, 0, 0)
	sub.EndDate = &newEndDate
	notes := "Updated notes"
	sub.Notes = &notes

	err = subRepo.Update(sub)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify
	got, _ := subRepo.GetByID(sub.ID)
	if got.SubscriptionType != domain.SubscriptionTypeAnnual {
		t.Errorf("SubscriptionType = %v, want annual", got.SubscriptionType)
	}
	if got.Notes == nil || *got.Notes != notes {
		t.Errorf("Notes = %v, want %v", got.Notes, notes)
	}
}

func TestUserSubscriptionRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	subRepo := NewSQLiteUserSubscriptionRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	now := time.Now()
	sub := &domain.UserSubscription{
		UserID:           user.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := subRepo.Create(sub); err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Delete
	err = subRepo.Delete(sub.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	got, _ := subRepo.GetByID(sub.ID)
	if got != nil {
		t.Error("Subscription should be deleted")
	}
}

func TestUserSubscriptionRepository_MarkAsExpired(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	subRepo := NewSQLiteUserSubscriptionRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	now := time.Now()
	sub := &domain.UserSubscription{
		UserID:           user.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := subRepo.Create(sub); err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Mark as expired
	err = subRepo.MarkAsExpired(sub.ID)
	if err != nil {
		t.Fatalf("MarkAsExpired() error = %v", err)
	}

	// Verify
	got, _ := subRepo.GetByID(sub.ID)
	if got.Status != domain.SubscriptionStatusExpired {
		t.Errorf("Status = %v, want expired", got.Status)
	}
}

func TestUserSubscriptionRepository_Cancel(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	subRepo := NewSQLiteUserSubscriptionRepository(db)

	// Create admin user for cancellation
	admin := &domain.User{
		Email:        "admin@example.com",
		PasswordHash: "hash",
		Name:         "Admin",
		Role:         "admin",
	}
	if err := userRepo.Create(admin); err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	now := time.Now()
	sub := &domain.UserSubscription{
		UserID:           user.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := subRepo.Create(sub); err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Cancel subscription
	reason := "User requested cancellation"
	err = subRepo.Cancel(sub.ID, reason, admin.ID)
	if err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}

	// Verify
	got, _ := subRepo.GetByID(sub.ID)
	if got.Status != domain.SubscriptionStatusCancelled {
		t.Errorf("Status = %v, want cancelled", got.Status)
	}
	if got.CancelledReason == nil || *got.CancelledReason != reason {
		t.Errorf("CancelledReason = %v, want %v", got.CancelledReason, reason)
	}
	if got.CancelledAt == nil {
		t.Error("CancelledAt should be set")
	}
}

func TestUserSubscriptionRepository_MarkAsPaid(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	subRepo := NewSQLiteUserSubscriptionRepository(db)

	// Create admin user
	admin := &domain.User{
		Email:        "admin@example.com",
		PasswordHash: "hash",
		Name:         "Admin",
		Role:         "admin",
	}
	if err := userRepo.Create(admin); err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	now := time.Now()
	sub := &domain.UserSubscription{
		UserID:           user.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusExpired,
		StartDate:        now.AddDate(0, -1, 0),
		CreatedAt:        now.AddDate(0, -1, 0),
		UpdatedAt:        now.AddDate(0, -1, 0),
	}
	if err := subRepo.Create(sub); err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Mark as paid (should reactivate and extend 30 days)
	paymentDate := now
	err = subRepo.MarkAsPaid(sub.ID, paymentDate, admin.ID, nil)
	if err != nil {
		t.Fatalf("MarkAsPaid() error = %v", err)
	}

	// Verify
	got, _ := subRepo.GetByID(sub.ID)
	if got.Status != domain.SubscriptionStatusActive {
		t.Errorf("Status = %v, want active", got.Status)
	}
	if got.LastPaymentDate == nil {
		t.Error("LastPaymentDate should be set")
	}
	if got.EndDate == nil {
		t.Error("EndDate should be set")
	}
	// Check that end date is approximately 30 days from payment
	expectedEnd := paymentDate.AddDate(0, 0, 30)
	if got.EndDate != nil {
		diff := got.EndDate.Sub(expectedEnd)
		if diff < -time.Minute || diff > time.Minute {
			t.Errorf("EndDate = %v, want approximately %v", got.EndDate, expectedEnd)
		}
	}
}

func TestUserSubscriptionRepository_MarkAsPaid_CustomDuration(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	subRepo := NewSQLiteUserSubscriptionRepository(db)

	// Create admin and test user
	admin := &domain.User{Email: "admin@example.com", PasswordHash: "hash", Name: "Admin", Role: "admin"}
	user := &domain.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test User", Role: "user"}
	userRepo.Create(admin)
	userRepo.Create(user)

	now := time.Now()
	sub := &domain.UserSubscription{
		UserID:           user.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusExpired,
		StartDate:        now.AddDate(0, -1, 0),
		CreatedAt:        now.AddDate(0, -1, 0),
		UpdatedAt:        now.AddDate(0, -1, 0),
	}
	subRepo.Create(sub)

	// Mark as paid with custom duration (45 days)
	paymentDate := now
	customDays := 45
	err = subRepo.MarkAsPaid(sub.ID, paymentDate, admin.ID, &customDays)
	if err != nil {
		t.Fatalf("MarkAsPaid() error = %v", err)
	}

	// Verify custom duration
	got, _ := subRepo.GetByID(sub.ID)
	expectedEnd := paymentDate.AddDate(0, 0, 45)
	if got.EndDate != nil {
		diff := got.EndDate.Sub(expectedEnd)
		if diff < -time.Minute || diff > time.Minute {
			t.Errorf("EndDate = %v, want approximately %v", got.EndDate, expectedEnd)
		}
	}
}

func TestUserSubscriptionRepository_PermanentFree(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	subRepo := NewSQLiteUserSubscriptionRepository(db)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "user",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	now := time.Now()
	sub := &domain.UserSubscription{
		UserID:           user.ID,
		SubscriptionType: domain.SubscriptionTypeFree,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  true,
		StartDate:        now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := subRepo.Create(sub); err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Verify permanent free subscription
	got, _ := subRepo.GetByID(sub.ID)
	if !got.IsPermanentFree {
		t.Error("IsPermanentFree should be true")
	}
	if got.EndDate != nil {
		t.Error("EndDate should be nil for permanent free subscription")
	}

	// Verify it's returned as active
	active, _ := subRepo.GetActiveByUserID(user.ID)
	if active == nil {
		t.Error("Permanent free subscription should be returned as active")
	}
}

func TestUserSubscriptionRepository_ListExpiring(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	subRepo := NewSQLiteUserSubscriptionRepository(db)

	// Create test users
	user1 := &domain.User{Email: "user1@example.com", PasswordHash: "hash", Name: "User 1", Role: "user"}
	user2 := &domain.User{Email: "user2@example.com", PasswordHash: "hash", Name: "User 2", Role: "user"}
	user3 := &domain.User{Email: "user3@example.com", PasswordHash: "hash", Name: "User 3", Role: "user"}
	user4 := &domain.User{Email: "user4@example.com", PasswordHash: "hash", Name: "User 4", Role: "user"}
	userRepo.Create(user1)
	userRepo.Create(user2)
	userRepo.Create(user3)
	userRepo.Create(user4)

	now := time.Now()

	// Subscription expiring in 10 days (should be included for 30 days query)
	endDate10 := now.AddDate(0, 0, 10)
	sub1 := &domain.UserSubscription{
		UserID:           user1.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now.AddDate(0, -1, 0),
		EndDate:          &endDate10,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	subRepo.Create(sub1)

	// Subscription expiring in 45 days (should NOT be included for 30 days query)
	endDate45 := now.AddDate(0, 0, 45)
	sub2 := &domain.UserSubscription{
		UserID:           user2.ID,
		SubscriptionType: domain.SubscriptionTypeAnnual,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now.AddDate(0, -11, 0),
		EndDate:          &endDate45,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	subRepo.Create(sub2)

	// Permanent free subscription (should NOT be included)
	sub3 := &domain.UserSubscription{
		UserID:           user3.ID,
		SubscriptionType: domain.SubscriptionTypeFree,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  true,
		StartDate:        now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	subRepo.Create(sub3)

	// Already expired subscription (should NOT be included)
	endDatePast := now.AddDate(0, 0, -5)
	sub4 := &domain.UserSubscription{
		UserID:           user4.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now.AddDate(0, -2, 0),
		EndDate:          &endDatePast,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	subRepo.Create(sub4)

	// Test ListExpiring with 30 days
	expiring, err := subRepo.ListExpiring(30)
	if err != nil {
		t.Fatalf("ListExpiring() error = %v", err)
	}

	if len(expiring) != 1 {
		t.Errorf("ListExpiring(30) returned %d subscriptions, want 1", len(expiring))
	}

	if len(expiring) > 0 && expiring[0].UserID != user1.ID {
		t.Errorf("Expected subscription for user1, got user ID %d", expiring[0].UserID)
	}

	// Test ListExpiring with 60 days (should include both sub1 and sub2)
	expiring60, err := subRepo.ListExpiring(60)
	if err != nil {
		t.Fatalf("ListExpiring(60) error = %v", err)
	}

	if len(expiring60) != 2 {
		t.Errorf("ListExpiring(60) returned %d subscriptions, want 2", len(expiring60))
	}
}

func TestUserSubscriptionRepository_ListAll(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	subRepo := NewSQLiteUserSubscriptionRepository(db)

	// Initially should have no subscriptions
	subs, err := subRepo.ListAll()
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(subs) != 0 {
		t.Errorf("ListAll() should return empty list initially, got %d", len(subs))
	}

	// Create test users
	user1 := &domain.User{Email: "user1@example.com", PasswordHash: "hash", Name: "User 1", Role: "user"}
	user2 := &domain.User{Email: "user2@example.com", PasswordHash: "hash", Name: "User 2", Role: "user"}
	user3 := &domain.User{Email: "user3@example.com", PasswordHash: "hash", Name: "User 3", Role: "user"}
	userRepo.Create(user1)
	userRepo.Create(user2)
	userRepo.Create(user3)

	now := time.Now()

	// Create subscriptions for different users with different types
	endDate1 := now.AddDate(0, 1, 0)
	sub1 := &domain.UserSubscription{
		UserID:           user1.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now,
		EndDate:          &endDate1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	subRepo.Create(sub1)

	endDate2 := now.AddDate(1, 0, 0)
	sub2 := &domain.UserSubscription{
		UserID:           user2.ID,
		SubscriptionType: domain.SubscriptionTypeAnnual,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now,
		EndDate:          &endDate2,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	subRepo.Create(sub2)

	sub3 := &domain.UserSubscription{
		UserID:           user3.ID,
		SubscriptionType: domain.SubscriptionTypeFree,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  true,
		StartDate:        now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	subRepo.Create(sub3)

	// Test ListAll returns all subscriptions
	subs, err = subRepo.ListAll()
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(subs) != 3 {
		t.Errorf("ListAll() returned %d subscriptions, want 3", len(subs))
	}

	// Verify all subscriptions have user email populated (from JOIN)
	for _, sub := range subs {
		if sub.UserEmail == "" {
			t.Errorf("ListAll() subscription %d should have UserEmail populated", sub.ID)
		}
	}
}

func TestUserSubscriptionRepository_ListExpired(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	subRepo := NewSQLiteUserSubscriptionRepository(db)

	// Create test users
	user1 := &domain.User{Email: "user1@example.com", PasswordHash: "hash", Name: "User 1", Role: "user"}
	user2 := &domain.User{Email: "user2@example.com", PasswordHash: "hash", Name: "User 2", Role: "user"}
	user3 := &domain.User{Email: "user3@example.com", PasswordHash: "hash", Name: "User 3", Role: "user"}
	userRepo.Create(user1)
	userRepo.Create(user2)
	userRepo.Create(user3)

	now := time.Now()

	// Subscription with status 'expired' (should be included)
	endDatePast := now.AddDate(0, 0, -10)
	sub1 := &domain.UserSubscription{
		UserID:           user1.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusExpired,
		IsPermanentFree:  false,
		StartDate:        now.AddDate(0, -2, 0),
		EndDate:          &endDatePast,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	subRepo.Create(sub1)

	// Active subscription with past end date (overdue, should be included)
	endDatePast2 := now.AddDate(0, 0, -5)
	sub2 := &domain.UserSubscription{
		UserID:           user2.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now.AddDate(0, -1, 0),
		EndDate:          &endDatePast2,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	subRepo.Create(sub2)

	// Active subscription with future end date (should NOT be included)
	endDateFuture := now.AddDate(0, 0, 30)
	sub3 := &domain.UserSubscription{
		UserID:           user3.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now,
		EndDate:          &endDateFuture,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	subRepo.Create(sub3)

	// Test ListExpired
	expired, err := subRepo.ListExpired()
	if err != nil {
		t.Fatalf("ListExpired() error = %v", err)
	}

	if len(expired) != 2 {
		t.Errorf("ListExpired() returned %d subscriptions, want 2", len(expired))
	}

	// Verify the correct subscriptions were returned
	foundUser1 := false
	foundUser2 := false
	for _, sub := range expired {
		if sub.UserID == user1.ID {
			foundUser1 = true
		}
		if sub.UserID == user2.ID {
			foundUser2 = true
		}
	}

	if !foundUser1 {
		t.Error("ListExpired() should include subscription with expired status")
	}
	if !foundUser2 {
		t.Error("ListExpired() should include active subscription with past end date")
	}
}
