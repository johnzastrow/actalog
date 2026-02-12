package repository

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestSubscriptionAccessRepository_CheckUserAccess_NoSubscription(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userSubRepo := NewSQLiteUserSubscriptionRepository(db)
	orgSubRepo := NewSQLiteOrganizationSubscriptionRepository(db)
	orgRepo := NewOrganizationRepository(db)
	accessRepo := NewSubscriptionAccessRepository(userSubRepo, orgSubRepo, orgRepo)

	// Create test user without subscription
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "athlete",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Check access
	result, err := accessRepo.CheckUserAccess(user.ID)
	if err != nil {
		t.Fatalf("CheckUserAccess() error = %v", err)
	}
	if result == nil {
		t.Fatal("CheckUserAccess() returned nil")
	}
	if result.HasAccess {
		t.Error("HasAccess should be false for user without subscription")
	}
	if result.Source != "none" {
		t.Errorf("Source = %v, want 'none'", result.Source)
	}
}

func TestSubscriptionAccessRepository_CheckUserAccess_UserSubscription(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userSubRepo := NewSQLiteUserSubscriptionRepository(db)
	orgSubRepo := NewSQLiteOrganizationSubscriptionRepository(db)
	orgRepo := NewOrganizationRepository(db)
	accessRepo := NewSubscriptionAccessRepository(userSubRepo, orgSubRepo, orgRepo)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "athlete",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create active subscription
	now := time.Now()
	endDate := now.AddDate(0, 1, 0) // 1 month from now
	sub := &domain.UserSubscription{
		UserID:           user.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now,
		EndDate:          &endDate,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := userSubRepo.Create(sub); err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Check access
	result, err := accessRepo.CheckUserAccess(user.ID)
	if err != nil {
		t.Fatalf("CheckUserAccess() error = %v", err)
	}
	if !result.HasAccess {
		t.Error("HasAccess should be true for user with active subscription")
	}
	if result.Source != "user" {
		t.Errorf("Source = %v, want 'user'", result.Source)
	}
	if result.UserSubscription == nil {
		t.Error("UserSubscription should not be nil")
	}
	if result.ExpiresAt == nil {
		t.Error("ExpiresAt should be set")
	}
}

func TestSubscriptionAccessRepository_CheckUserAccess_PermanentFree(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userSubRepo := NewSQLiteUserSubscriptionRepository(db)
	orgSubRepo := NewSQLiteOrganizationSubscriptionRepository(db)
	orgRepo := NewOrganizationRepository(db)
	accessRepo := NewSubscriptionAccessRepository(userSubRepo, orgSubRepo, orgRepo)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "athlete",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create permanent free subscription
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
	if err := userSubRepo.Create(sub); err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Check access
	result, err := accessRepo.CheckUserAccess(user.ID)
	if err != nil {
		t.Fatalf("CheckUserAccess() error = %v", err)
	}
	if !result.HasAccess {
		t.Error("HasAccess should be true for permanent free subscription")
	}
	if result.Source != "user" {
		t.Errorf("Source = %v, want 'user'", result.Source)
	}
	if result.ExpiresAt != nil {
		t.Error("ExpiresAt should be nil for permanent free subscription")
	}
}

func TestSubscriptionAccessRepository_CheckUserAccess_ExpiredSubscription(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userSubRepo := NewSQLiteUserSubscriptionRepository(db)
	orgSubRepo := NewSQLiteOrganizationSubscriptionRepository(db)
	orgRepo := NewOrganizationRepository(db)
	accessRepo := NewSubscriptionAccessRepository(userSubRepo, orgSubRepo, orgRepo)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "athlete",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create expired subscription (active status but past end date)
	now := time.Now()
	endDate := now.AddDate(0, 0, -1) // 1 day ago
	sub := &domain.UserSubscription{
		UserID:           user.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now.AddDate(0, -1, -1),
		EndDate:          &endDate,
		CreatedAt:        now.AddDate(0, -1, -1),
		UpdatedAt:        now.AddDate(0, -1, -1),
	}
	if err := userSubRepo.Create(sub); err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Check access
	result, err := accessRepo.CheckUserAccess(user.ID)
	if err != nil {
		t.Fatalf("CheckUserAccess() error = %v", err)
	}
	if result.HasAccess {
		t.Error("HasAccess should be false for expired subscription")
	}
	// The subscription should still be included in result for display purposes
	if result.UserSubscription == nil {
		t.Error("UserSubscription should not be nil (for status display)")
	}
}

func TestSubscriptionAccessRepository_CheckUserAccess_OrganizationSubscription(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userSubRepo := NewSQLiteUserSubscriptionRepository(db)
	orgSubRepo := NewSQLiteOrganizationSubscriptionRepository(db)
	orgRepo := NewOrganizationRepository(db)
	accessRepo := NewSubscriptionAccessRepository(userSubRepo, orgSubRepo, orgRepo)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "athlete",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create organization
	org := &domain.Organization{
		Name:      "Test Org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := orgRepo.Create(org); err != nil {
		t.Fatalf("Failed to create organization: %v", err)
	}

	// Add user to organization
	if err := orgRepo.AddUserToOrganization(user.ID, org.ID, "member"); err != nil {
		t.Fatalf("Failed to add user to organization: %v", err)
	}

	// Create organization subscription
	now := time.Now()
	endDate := now.AddDate(0, 1, 0) // 1 month from now
	orgSub := &domain.OrganizationSubscription{
		OrganizationID:   org.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now,
		EndDate:          &endDate,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := orgSubRepo.Create(orgSub); err != nil {
		t.Fatalf("Failed to create organization subscription: %v", err)
	}

	// Check access
	result, err := accessRepo.CheckUserAccess(user.ID)
	if err != nil {
		t.Fatalf("CheckUserAccess() error = %v", err)
	}
	if !result.HasAccess {
		t.Error("HasAccess should be true for user in org with active subscription")
	}
	if result.Source != "organization" {
		t.Errorf("Source = %v, want 'organization'", result.Source)
	}
	if len(result.OrgSubscriptions) != 1 {
		t.Errorf("OrgSubscriptions length = %d, want 1", len(result.OrgSubscriptions))
	}
}

func TestSubscriptionAccessRepository_CheckUserAccess_BothSubscriptions(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userSubRepo := NewSQLiteUserSubscriptionRepository(db)
	orgSubRepo := NewSQLiteOrganizationSubscriptionRepository(db)
	orgRepo := NewOrganizationRepository(db)
	accessRepo := NewSubscriptionAccessRepository(userSubRepo, orgSubRepo, orgRepo)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "athlete",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create active user subscription
	now := time.Now()
	endDate := now.AddDate(0, 1, 0)
	userSub := &domain.UserSubscription{
		UserID:           user.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now,
		EndDate:          &endDate,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := userSubRepo.Create(userSub); err != nil {
		t.Fatalf("Failed to create user subscription: %v", err)
	}

	// Check access - should return "user" source since user sub is checked first
	result, err := accessRepo.CheckUserAccess(user.ID)
	if err != nil {
		t.Fatalf("CheckUserAccess() error = %v", err)
	}
	if !result.HasAccess {
		t.Error("HasAccess should be true")
	}
	// User subscription is checked first, so it should be the source
	if result.Source != "user" {
		t.Errorf("Source = %v, want 'user'", result.Source)
	}
	if result.UserSubscription == nil {
		t.Error("UserSubscription should not be nil")
	}
}

func TestSubscriptionAccessRepository_CheckUserAccess_MultipleOrganizations(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userSubRepo := NewSQLiteUserSubscriptionRepository(db)
	orgSubRepo := NewSQLiteOrganizationSubscriptionRepository(db)
	orgRepo := NewOrganizationRepository(db)
	accessRepo := NewSubscriptionAccessRepository(userSubRepo, orgSubRepo, orgRepo)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "athlete",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	now := time.Now()

	// Create two organizations
	org1 := &domain.Organization{Name: "Org 1", CreatedAt: now, UpdatedAt: now}
	org2 := &domain.Organization{Name: "Org 2", CreatedAt: now, UpdatedAt: now}
	orgRepo.Create(org1)
	orgRepo.Create(org2)

	// Add user to both organizations
	orgRepo.AddUserToOrganization(user.ID, org1.ID, "member")
	orgRepo.AddUserToOrganization(user.ID, org2.ID, "member")

	// Create subscriptions with different end dates
	endDate1 := now.AddDate(0, 1, 0) // 1 month
	endDate2 := now.AddDate(0, 2, 0) // 2 months

	orgSub1 := &domain.OrganizationSubscription{
		OrganizationID:   org1.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now,
		EndDate:          &endDate1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	orgSub2 := &domain.OrganizationSubscription{
		OrganizationID:   org2.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now,
		EndDate:          &endDate2,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	orgSubRepo.Create(orgSub1)
	orgSubRepo.Create(orgSub2)

	// Check access
	result, err := accessRepo.CheckUserAccess(user.ID)
	if err != nil {
		t.Fatalf("CheckUserAccess() error = %v", err)
	}
	if !result.HasAccess {
		t.Error("HasAccess should be true")
	}
	if result.Source != "organization" {
		t.Errorf("Source = %v, want 'organization'", result.Source)
	}
	if len(result.OrgSubscriptions) != 2 {
		t.Errorf("OrgSubscriptions length = %d, want 2", len(result.OrgSubscriptions))
	}
	// ExpiresAt should be the earliest expiration date
	if result.ExpiresAt == nil {
		t.Error("ExpiresAt should be set")
	}
	expectedExpires := endDate1
	if result.ExpiresAt != nil {
		diff := result.ExpiresAt.Sub(expectedExpires)
		if diff < -time.Minute || diff > time.Minute {
			t.Errorf("ExpiresAt = %v, want approximately %v (earliest org sub)", result.ExpiresAt, expectedExpires)
		}
	}
}

func TestSubscriptionAccessRepository_CheckUserAccess_InactiveUserSubWithActiveOrgSub(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	userSubRepo := NewSQLiteUserSubscriptionRepository(db)
	orgSubRepo := NewSQLiteOrganizationSubscriptionRepository(db)
	orgRepo := NewOrganizationRepository(db)
	accessRepo := NewSubscriptionAccessRepository(userSubRepo, orgSubRepo, orgRepo)

	// Create test user
	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "athlete",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	now := time.Now()

	// Create expired user subscription
	expiredEnd := now.AddDate(0, 0, -1)
	userSub := &domain.UserSubscription{
		UserID:           user.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now.AddDate(0, -1, -1),
		EndDate:          &expiredEnd,
		CreatedAt:        now.AddDate(0, -1, -1),
		UpdatedAt:        now.AddDate(0, -1, -1),
	}
	userSubRepo.Create(userSub)

	// Create organization and subscription
	org := &domain.Organization{Name: "Test Org", CreatedAt: now, UpdatedAt: now}
	orgRepo.Create(org)
	orgRepo.AddUserToOrganization(user.ID, org.ID, "member")

	endDate := now.AddDate(0, 1, 0)
	orgSub := &domain.OrganizationSubscription{
		OrganizationID:   org.ID,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        now,
		EndDate:          &endDate,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	orgSubRepo.Create(orgSub)

	// Check access
	result, err := accessRepo.CheckUserAccess(user.ID)
	if err != nil {
		t.Fatalf("CheckUserAccess() error = %v", err)
	}
	if !result.HasAccess {
		t.Error("HasAccess should be true (from org subscription)")
	}
	// When user has inactive user sub but active org sub, source should be "both"
	// because the code includes the user subscription in the result
	if result.Source != "both" {
		t.Errorf("Source = %v, want 'both'", result.Source)
	}
	if result.UserSubscription == nil {
		t.Error("UserSubscription should not be nil (even if inactive)")
	}
	if len(result.OrgSubscriptions) != 1 {
		t.Errorf("OrgSubscriptions length = %d, want 1", len(result.OrgSubscriptions))
	}
}
