package repository

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestOrganizationSubscriptionRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	orgRepo := NewOrganizationRepository(db)
	subRepo := NewSQLiteOrganizationSubscriptionRepository(db)

	// Create test organization
	org := &domain.Organization{
		Name:      "Test Org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	now := time.Now()
	endDate := now.AddDate(0, 1, 0)
	sub := &domain.OrganizationSubscription{
		OrganizationID:   org.ID,
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

func TestOrganizationSubscriptionRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	orgRepo := NewOrganizationRepository(db)
	subRepo := NewSQLiteOrganizationSubscriptionRepository(db)

	// Create test organization
	org := &domain.Organization{
		Name:      "Test Org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	now := time.Now()
	endDate := now.AddDate(0, 1, 0)
	notes := "Test subscription"
	sub := &domain.OrganizationSubscription{
		OrganizationID:   org.ID,
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
	if got.OrganizationID != org.ID {
		t.Errorf("OrganizationID = %d, want %d", got.OrganizationID, org.ID)
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

func TestOrganizationSubscriptionRepository_GetActiveByOrganizationID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	orgRepo := NewOrganizationRepository(db)
	subRepo := NewSQLiteOrganizationSubscriptionRepository(db)

	// Create test organization
	org := &domain.Organization{
		Name:      "Test Org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	// Test getting non-existent subscription
	got, err := subRepo.GetActiveByOrganizationID(org.ID)
	if err != nil {
		t.Fatalf("GetActiveByOrganizationID() error = %v", err)
	}
	if got != nil {
		t.Error("GetActiveByOrganizationID() should return nil for org without subscription")
	}

	// Create active subscription
	now := time.Now()
	endDate := now.AddDate(0, 1, 0)
	activeSub := &domain.OrganizationSubscription{
		OrganizationID:   org.ID,
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
	got, err = subRepo.GetActiveByOrganizationID(org.ID)
	if err != nil {
		t.Fatalf("GetActiveByOrganizationID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetActiveByOrganizationID() returned nil for org with active subscription")
	}
	if got.ID != activeSub.ID {
		t.Errorf("ID = %d, want %d", got.ID, activeSub.ID)
	}
}

func TestOrganizationSubscriptionRepository_GetByOrganizationID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	orgRepo := NewOrganizationRepository(db)
	subRepo := NewSQLiteOrganizationSubscriptionRepository(db)

	// Create test organization
	org := &domain.Organization{
		Name:      "Test Org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
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
		sub := &domain.OrganizationSubscription{
			OrganizationID:   org.ID,
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

	// Get all subscriptions for organization
	subs, err := subRepo.GetByOrganizationID(org.ID)
	if err != nil {
		t.Fatalf("GetByOrganizationID() error = %v", err)
	}
	if len(subs) != 3 {
		t.Errorf("GetByOrganizationID() returned %d subscriptions, want 3", len(subs))
	}

	// Non-existent organization
	subs, err = subRepo.GetByOrganizationID(999)
	if err != nil {
		t.Fatalf("GetByOrganizationID(999) error = %v", err)
	}
	if len(subs) != 0 {
		t.Errorf("GetByOrganizationID(999) returned %d subscriptions, want 0", len(subs))
	}
}

func TestOrganizationSubscriptionRepository_Update(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	orgRepo := NewOrganizationRepository(db)
	subRepo := NewSQLiteOrganizationSubscriptionRepository(db)

	// Create test organization
	org := &domain.Organization{
		Name:      "Test Org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	now := time.Now()
	endDate := now.AddDate(0, 1, 0)
	sub := &domain.OrganizationSubscription{
		OrganizationID:   org.ID,
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

func TestOrganizationSubscriptionRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	orgRepo := NewOrganizationRepository(db)
	subRepo := NewSQLiteOrganizationSubscriptionRepository(db)

	// Create test organization
	org := &domain.Organization{
		Name:      "Test Org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	now := time.Now()
	sub := &domain.OrganizationSubscription{
		OrganizationID:   org.ID,
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

func TestOrganizationSubscriptionRepository_MarkAsExpired(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	orgRepo := NewOrganizationRepository(db)
	subRepo := NewSQLiteOrganizationSubscriptionRepository(db)

	// Create test organization
	org := &domain.Organization{
		Name:      "Test Org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	now := time.Now()
	sub := &domain.OrganizationSubscription{
		OrganizationID:   org.ID,
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

func TestOrganizationSubscriptionRepository_Cancel(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	orgRepo := NewOrganizationRepository(db)
	subRepo := NewSQLiteOrganizationSubscriptionRepository(db)

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

	// Create test organization
	org := &domain.Organization{
		Name:      "Test Org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	now := time.Now()
	sub := &domain.OrganizationSubscription{
		OrganizationID:   org.ID,
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
	reason := "Organization requested cancellation"
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

func TestOrganizationSubscriptionRepository_MarkAsPaid(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	orgRepo := NewOrganizationRepository(db)
	subRepo := NewSQLiteOrganizationSubscriptionRepository(db)

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

	// Create test organization
	org := &domain.Organization{
		Name:      "Test Org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	now := time.Now()
	sub := &domain.OrganizationSubscription{
		OrganizationID:   org.ID,
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

func TestOrganizationSubscriptionRepository_PermanentFree(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	orgRepo := NewOrganizationRepository(db)
	subRepo := NewSQLiteOrganizationSubscriptionRepository(db)

	// Create test organization
	org := &domain.Organization{
		Name:      "Test Org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	now := time.Now()
	sub := &domain.OrganizationSubscription{
		OrganizationID:   org.ID,
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
	active, _ := subRepo.GetActiveByOrganizationID(org.ID)
	if active == nil {
		t.Error("Permanent free subscription should be returned as active")
	}
}
