package service

import (
	"errors"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestSubscriptionService_CreateUserSubscription(t *testing.T) {
	tests := []struct {
		name             string
		adminUserID      int64
		targetUserID     int64
		subscriptionType string
		isPermanentFree  bool
		notes            string
		setupMock        func(*mockUserSubscriptionRepo, *mockUserRepo)
		expectedError    error
	}{
		{
			name:             "successful free subscription creation",
			adminUserID:      1,
			targetUserID:     2,
			subscriptionType: "free",
			isPermanentFree:  false,
			notes:            "Test subscription",
			setupMock: func(subRepo *mockUserSubscriptionRepo, userRepo *mockUserRepo) {
				userRepo.users[2] = &domain.User{
					ID:    2,
					Email: "user@example.com",
					Name:  "Test User",
				}
			},
			expectedError: nil,
		},
		{
			name:             "successful monthly subscription creation",
			adminUserID:      1,
			targetUserID:     2,
			subscriptionType: "monthly",
			isPermanentFree:  false,
			notes:            "Monthly subscription",
			setupMock: func(subRepo *mockUserSubscriptionRepo, userRepo *mockUserRepo) {
				userRepo.users[2] = &domain.User{
					ID:    2,
					Email: "user@example.com",
					Name:  "Test User",
				}
			},
			expectedError: nil,
		},
		{
			name:             "successful permanent free subscription creation",
			adminUserID:      1,
			targetUserID:     2,
			subscriptionType: "free",
			isPermanentFree:  true,
			notes:            "Permanent free for early adopter",
			setupMock: func(subRepo *mockUserSubscriptionRepo, userRepo *mockUserRepo) {
				userRepo.users[2] = &domain.User{
					ID:    2,
					Email: "user@example.com",
					Name:  "Test User",
				}
			},
			expectedError: nil,
		},
		{
			name:             "admin cannot modify own subscription",
			adminUserID:      1,
			targetUserID:     1,
			subscriptionType: "free",
			isPermanentFree:  false,
			notes:            "",
			setupMock:        func(subRepo *mockUserSubscriptionRepo, userRepo *mockUserRepo) {},
			expectedError:    ErrCannotModifyOwnSubscription,
		},
		{
			name:             "user already has active subscription",
			adminUserID:      1,
			targetUserID:     2,
			subscriptionType: "monthly",
			isPermanentFree:  false,
			notes:            "",
			setupMock: func(subRepo *mockUserSubscriptionRepo, userRepo *mockUserRepo) {
				userRepo.users[2] = &domain.User{
					ID:    2,
					Email: "user@example.com",
					Name:  "Test User",
				}
				endDate := time.Now().Add(30 * 24 * time.Hour)
				subRepo.subscriptions[1] = &domain.UserSubscription{
					ID:               1,
					UserID:           2,
					SubscriptionType: domain.SubscriptionTypeFree,
					Status:           domain.SubscriptionStatusActive,
					StartDate:        time.Now(),
					EndDate:          &endDate,
				}
			},
			expectedError: ErrActiveSubscriptionExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subRepo := newMockUserSubscriptionRepo()
			userRepo := newMockUserRepo()
			orgRepo := newMockOrganizationRepo()
			orgSubRepo := newMockOrganizationSubscriptionRepo()
			accessRepo := newMockSubscriptionAccessRepo()
			auditRepo := newMockAuditLogRepo()

			if tt.setupMock != nil {
				tt.setupMock(subRepo, userRepo)
			}

			service := NewSubscriptionService(
				subRepo,
				orgSubRepo,
				accessRepo,
				auditRepo,
				userRepo,
				orgRepo,
			)

			sub, err := service.CreateUserSubscription(
				tt.adminUserID,
				tt.targetUserID,
				tt.subscriptionType,
				tt.isPermanentFree,
				tt.notes,
			)

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if sub == nil {
				t.Fatal("expected subscription, got nil")
			}

			// Verify subscription properties
			if sub.UserID != tt.targetUserID {
				t.Errorf("expected UserID %d, got %d", tt.targetUserID, sub.UserID)
			}
			if string(sub.SubscriptionType) != tt.subscriptionType {
				t.Errorf("expected SubscriptionType %s, got %s", tt.subscriptionType, sub.SubscriptionType)
			}
			if sub.IsPermanentFree != tt.isPermanentFree {
				t.Errorf("expected IsPermanentFree %v, got %v", tt.isPermanentFree, sub.IsPermanentFree)
			}
			if sub.Status != domain.SubscriptionStatusActive {
				t.Errorf("expected Status active, got %s", sub.Status)
			}
			if sub.CreatedByUserID == nil || *sub.CreatedByUserID != tt.adminUserID {
				t.Errorf("expected CreatedByUserID %d, got %v", tt.adminUserID, sub.CreatedByUserID)
			}

			// Verify end date is set correctly
			if !tt.isPermanentFree && (tt.subscriptionType == "monthly" || tt.subscriptionType == "annual") {
				// Monthly and annual subscriptions should have end date set
				if sub.EndDate == nil {
					t.Error("expected EndDate to be set for monthly/annual subscription")
				} else {
					expectedDuration := 30 * 24 * time.Hour // Default for monthly
					if tt.subscriptionType == "annual" {
						expectedDuration = 365 * 24 * time.Hour
					}
					expectedEndDate := sub.StartDate.Add(expectedDuration)
					if sub.EndDate.Before(expectedEndDate.Add(-1*time.Minute)) || sub.EndDate.After(expectedEndDate.Add(1*time.Minute)) {
						t.Errorf("EndDate not set correctly, expected around %v, got %v", expectedEndDate, sub.EndDate)
					}
				}
			} else {
				// Free subscriptions and permanent free subscriptions should have no end date
				if sub.EndDate != nil {
					t.Error("expected EndDate to be nil for free or permanent free subscription")
				}
			}
		})
	}
}

func TestSubscriptionService_MarkUserSubscriptionAsPaid(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID int64
		adminUserID    int64
		paymentDate    time.Time
		durationDays   *int
		setupMock      func(*mockUserSubscriptionRepo)
		expectedError  error
	}{
		{
			name:           "successful payment marking with default duration",
			subscriptionID: 1,
			adminUserID:    1,
			paymentDate:    time.Now(),
			durationDays:   nil,
			setupMock: func(repo *mockUserSubscriptionRepo) {
				endDate := time.Now().Add(10 * 24 * time.Hour) // Expires in 10 days
				repo.subscriptions[1] = &domain.UserSubscription{
					ID:               1,
					UserID:           2,
					SubscriptionType: domain.SubscriptionTypeMonthly,
					Status:           domain.SubscriptionStatusActive,
					IsPermanentFree:  false,
					StartDate:        time.Now().Add(-20 * 24 * time.Hour),
					EndDate:          &endDate,
				}
			},
			expectedError: nil,
		},
		{
			name:           "successful payment marking with custom duration",
			subscriptionID: 1,
			adminUserID:    1,
			paymentDate:    time.Now(),
			durationDays:   intPtr(60), // 2 months
			setupMock: func(repo *mockUserSubscriptionRepo) {
				endDate := time.Now().Add(10 * 24 * time.Hour)
				repo.subscriptions[1] = &domain.UserSubscription{
					ID:               1,
					UserID:           2,
					SubscriptionType: domain.SubscriptionTypeMonthly,
					Status:           domain.SubscriptionStatusActive,
					IsPermanentFree:  false,
					StartDate:        time.Now().Add(-20 * 24 * time.Hour),
					EndDate:          &endDate,
				}
			},
			expectedError: nil,
		},
		{
			name:           "cannot mark free subscription as paid",
			subscriptionID: 1,
			adminUserID:    1,
			paymentDate:    time.Now(),
			durationDays:   nil,
			setupMock: func(repo *mockUserSubscriptionRepo) {
				repo.subscriptions[1] = &domain.UserSubscription{
					ID:               1,
					UserID:           2,
					SubscriptionType: domain.SubscriptionTypeFree,
					Status:           domain.SubscriptionStatusActive,
					IsPermanentFree:  false,
					StartDate:        time.Now(),
					EndDate:          nil,
				}
			},
			expectedError: ErrCannotMarkFreeSubscriptionPaid,
		},
		{
			name:           "subscription not found",
			subscriptionID: 999,
			adminUserID:    1,
			paymentDate:    time.Now(),
			durationDays:   nil,
			setupMock:      func(repo *mockUserSubscriptionRepo) {},
			expectedError:  ErrSubscriptionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subRepo := newMockUserSubscriptionRepo()
			userRepo := newMockUserRepo()
			orgRepo := newMockOrganizationRepo()
			orgSubRepo := newMockOrganizationSubscriptionRepo()
			accessRepo := newMockSubscriptionAccessRepo()
			auditRepo := newMockAuditLogRepo()

			if tt.setupMock != nil {
				tt.setupMock(subRepo)
			}

			service := NewSubscriptionService(
				subRepo,
				orgSubRepo,
				accessRepo,
				auditRepo,
				userRepo,
				orgRepo,
			)

			// Convert time.Time to ISO date string
			paymentDateStr := tt.paymentDate.Format("2006-01-02")

			err := service.MarkUserSubscriptionAsPaid(
				tt.adminUserID,
				tt.subscriptionID,
				&paymentDateStr,
				tt.durationDays,
			)

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify subscription was updated
			sub := subRepo.subscriptions[tt.subscriptionID]
			if sub.LastPaymentDate == nil {
				t.Error("expected LastPaymentDate to be set")
			}
			if sub.EndDate == nil {
				t.Error("expected EndDate to be set")
			}
			if sub.NextBillingDate == nil {
				t.Error("expected NextBillingDate to be set")
			}
		})
	}
}

func TestSubscriptionService_CancelUserSubscription(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID int64
		reason         string
		adminUserID    int64
		setupMock      func(*mockUserSubscriptionRepo)
		expectedError  error
	}{
		{
			name:           "successful cancellation",
			subscriptionID: 1,
			reason:         "User requested cancellation",
			adminUserID:    1,
			setupMock: func(repo *mockUserSubscriptionRepo) {
				endDate := time.Now().Add(30 * 24 * time.Hour)
				repo.subscriptions[1] = &domain.UserSubscription{
					ID:               1,
					UserID:           2,
					SubscriptionType: domain.SubscriptionTypeMonthly,
					Status:           domain.SubscriptionStatusActive,
					IsPermanentFree:  false,
					StartDate:        time.Now(),
					EndDate:          &endDate,
				}
			},
			expectedError: nil,
		},
		{
			name:           "subscription not found",
			subscriptionID: 999,
			reason:         "Cancellation reason",
			adminUserID:    1,
			setupMock:      func(repo *mockUserSubscriptionRepo) {},
			expectedError:  ErrSubscriptionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subRepo := newMockUserSubscriptionRepo()
			userRepo := newMockUserRepo()
			orgRepo := newMockOrganizationRepo()
			orgSubRepo := newMockOrganizationSubscriptionRepo()
			accessRepo := newMockSubscriptionAccessRepo()
			auditRepo := newMockAuditLogRepo()

			if tt.setupMock != nil {
				tt.setupMock(subRepo)
			}

			service := NewSubscriptionService(
				subRepo,
				orgSubRepo,
				accessRepo,
				auditRepo,
				userRepo,
				orgRepo,
			)

			err := service.CancelUserSubscription(
				tt.adminUserID,
				tt.subscriptionID,
				tt.reason,
			)

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify subscription was cancelled
			sub := subRepo.subscriptions[tt.subscriptionID]
			if sub.Status != domain.SubscriptionStatusCancelled {
				t.Errorf("expected Status cancelled, got %s", sub.Status)
			}
			if sub.CancelledAt == nil {
				t.Error("expected CancelledAt to be set")
			}
			if sub.CancelledReason == nil || *sub.CancelledReason != tt.reason {
				t.Errorf("expected CancelledReason %s, got %v", tt.reason, sub.CancelledReason)
			}
		})
	}
}

func TestSubscriptionService_GetUserSubscriptionStatus(t *testing.T) {
	tests := []struct {
		name              string
		userID            int64
		setupMock         func(*mockSubscriptionAccessRepo)
		expectedHasAccess bool
		expectedSource    string
	}{
		{
			name:   "user has active personal subscription",
			userID: 1,
			setupMock: func(repo *mockSubscriptionAccessRepo) {
				endDate := time.Now().Add(30 * 24 * time.Hour)
				repo.accessResults[1] = &domain.SubscriptionAccessResult{
					HasAccess: true,
					Source:    "user",
					UserSubscription: &domain.UserSubscription{
						ID:               1,
						UserID:           1,
						SubscriptionType: domain.SubscriptionTypeMonthly,
						Status:           domain.SubscriptionStatusActive,
						EndDate:          &endDate,
					},
					OrgSubscriptions: nil,
				}
			},
			expectedHasAccess: true,
			expectedSource:    "user",
		},
		{
			name:   "user has access through organization",
			userID: 2,
			setupMock: func(repo *mockSubscriptionAccessRepo) {
				endDate := time.Now().Add(30 * 24 * time.Hour)
				repo.accessResults[2] = &domain.SubscriptionAccessResult{
					HasAccess:        true,
					Source:           "organization",
					UserSubscription: nil,
					OrgSubscriptions: []*domain.OrganizationSubscription{
						{
							ID:               1,
							OrganizationID:   1,
							OrganizationName: "Test Org",
							SubscriptionType: domain.SubscriptionTypeAnnual,
							Status:           domain.SubscriptionStatusActive,
							EndDate:          &endDate,
						},
					},
				}
			},
			expectedHasAccess: true,
			expectedSource:    "organization",
		},
		{
			name:   "user has no active subscription",
			userID: 3,
			setupMock: func(repo *mockSubscriptionAccessRepo) {
				repo.accessResults[3] = &domain.SubscriptionAccessResult{
					HasAccess:        false,
					Source:           "none",
					UserSubscription: nil,
					OrgSubscriptions: nil,
				}
			},
			expectedHasAccess: false,
			expectedSource:    "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subRepo := newMockUserSubscriptionRepo()
			userRepo := newMockUserRepo()
			orgRepo := newMockOrganizationRepo()
			orgSubRepo := newMockOrganizationSubscriptionRepo()
			accessRepo := newMockSubscriptionAccessRepo()
			auditRepo := newMockAuditLogRepo()

			if tt.setupMock != nil {
				tt.setupMock(accessRepo)
			}

			service := NewSubscriptionService(
				subRepo,
				orgSubRepo,
				accessRepo,
				auditRepo,
				userRepo,
				orgRepo,
			)

			result, err := service.GetUserSubscriptionStatus(tt.userID)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			if result.HasAccess != tt.expectedHasAccess {
				t.Errorf("expected HasAccess %v, got %v", tt.expectedHasAccess, result.HasAccess)
			}
			if result.Source != tt.expectedSource {
				t.Errorf("expected Source %s, got %s", tt.expectedSource, result.Source)
			}
		})
	}
}

func TestSubscriptionService_CreateOrganizationSubscription(t *testing.T) {
	tests := []struct {
		name             string
		adminUserID      int64
		organizationID   int64
		subscriptionType string
		isPermanentFree  bool
		notes            string
		setupMock        func(*mockOrganizationSubscriptionRepo, *mockOrganizationRepo)
		expectedError    error
	}{
		{
			name:             "successful organization subscription creation",
			adminUserID:      1,
			organizationID:   1,
			subscriptionType: "annual",
			isPermanentFree:  false,
			notes:            "Annual subscription",
			setupMock: func(subRepo *mockOrganizationSubscriptionRepo, orgRepo *mockOrganizationRepo) {
				orgRepo.organizations[1] = &domain.Organization{
					ID:   1,
					Name: "Test Organization",
				}
			},
			expectedError: nil,
		},
		{
			name:             "organization already has active subscription",
			adminUserID:      1,
			organizationID:   1,
			subscriptionType: "annual",
			isPermanentFree:  false,
			notes:            "",
			setupMock: func(subRepo *mockOrganizationSubscriptionRepo, orgRepo *mockOrganizationRepo) {
				orgRepo.organizations[1] = &domain.Organization{
					ID:   1,
					Name: "Test Organization",
				}
				endDate := time.Now().Add(365 * 24 * time.Hour)
				subRepo.subscriptions[1] = &domain.OrganizationSubscription{
					ID:               1,
					OrganizationID:   1,
					SubscriptionType: domain.SubscriptionTypeAnnual,
					Status:           domain.SubscriptionStatusActive,
					StartDate:        time.Now(),
					EndDate:          &endDate,
				}
			},
			expectedError: ErrActiveOrgSubscriptionExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subRepo := newMockUserSubscriptionRepo()
			userRepo := newMockUserRepo()
			orgSubRepo := newMockOrganizationSubscriptionRepo()
			orgRepo := newMockOrganizationRepo()
			accessRepo := newMockSubscriptionAccessRepo()
			auditRepo := newMockAuditLogRepo()

			if tt.setupMock != nil {
				tt.setupMock(orgSubRepo, orgRepo)
			}

			service := NewSubscriptionService(
				subRepo,
				orgSubRepo,
				accessRepo,
				auditRepo,
				userRepo,
				orgRepo,
			)

			sub, err := service.CreateOrganizationSubscription(
				tt.adminUserID,
				tt.organizationID,
				tt.subscriptionType,
				tt.isPermanentFree,
				tt.notes,
			)

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if sub == nil {
				t.Fatal("expected subscription, got nil")
			}

			// Verify subscription properties
			if sub.OrganizationID != tt.organizationID {
				t.Errorf("expected OrganizationID %d, got %d", tt.organizationID, sub.OrganizationID)
			}
			if string(sub.SubscriptionType) != tt.subscriptionType {
				t.Errorf("expected SubscriptionType %s, got %s", tt.subscriptionType, sub.SubscriptionType)
			}
			if sub.IsPermanentFree != tt.isPermanentFree {
				t.Errorf("expected IsPermanentFree %v, got %v", tt.isPermanentFree, sub.IsPermanentFree)
			}
			if sub.Status != domain.SubscriptionStatusActive {
				t.Errorf("expected Status active, got %s", sub.Status)
			}
		})
	}
}

// Tests for MarkOrganizationSubscriptionAsPaid
func TestSubscriptionService_MarkOrganizationSubscriptionAsPaid(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID int64
		adminUserID    int64
		paymentDate    *string
		durationDays   *int
		setupMock      func(*mockOrganizationSubscriptionRepo)
		expectedError  error
	}{
		{
			name:           "successful payment marking",
			subscriptionID: 1,
			adminUserID:    1,
			paymentDate:    nil,
			durationDays:   nil,
			setupMock: func(repo *mockOrganizationSubscriptionRepo) {
				endDate := time.Now().Add(10 * 24 * time.Hour)
				repo.subscriptions[1] = &domain.OrganizationSubscription{
					ID:               1,
					OrganizationID:   10,
					SubscriptionType: domain.SubscriptionTypeMonthly,
					Status:           domain.SubscriptionStatusActive,
					IsPermanentFree:  false,
					StartDate:        time.Now().Add(-20 * 24 * time.Hour),
					EndDate:          &endDate,
				}
			},
			expectedError: nil,
		},
		{
			name:           "with custom duration",
			subscriptionID: 1,
			adminUserID:    1,
			paymentDate:    stringPtr("2025-01-15"),
			durationDays:   intPtr(90),
			setupMock: func(repo *mockOrganizationSubscriptionRepo) {
				endDate := time.Now().Add(10 * 24 * time.Hour)
				repo.subscriptions[1] = &domain.OrganizationSubscription{
					ID:               1,
					OrganizationID:   10,
					SubscriptionType: domain.SubscriptionTypeAnnual,
					Status:           domain.SubscriptionStatusActive,
					StartDate:        time.Now(),
					EndDate:          &endDate,
				}
			},
			expectedError: nil,
		},
		{
			name:           "cannot mark free subscription as paid",
			subscriptionID: 1,
			adminUserID:    1,
			paymentDate:    nil,
			durationDays:   nil,
			setupMock: func(repo *mockOrganizationSubscriptionRepo) {
				repo.subscriptions[1] = &domain.OrganizationSubscription{
					ID:               1,
					OrganizationID:   10,
					SubscriptionType: domain.SubscriptionTypeFree,
					Status:           domain.SubscriptionStatusActive,
				}
			},
			expectedError: ErrCannotMarkFreeSubscriptionPaid,
		},
		{
			name:           "subscription not found",
			subscriptionID: 999,
			adminUserID:    1,
			paymentDate:    nil,
			durationDays:   nil,
			setupMock:      func(repo *mockOrganizationSubscriptionRepo) {},
			expectedError:  nil, // Will check for any error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subRepo := newMockUserSubscriptionRepo()
			userRepo := newMockUserRepo()
			orgRepo := newMockOrganizationRepo()
			orgSubRepo := newMockOrganizationSubscriptionRepo()
			accessRepo := newMockSubscriptionAccessRepo()
			auditRepo := newMockAuditLogRepo()

			if tt.setupMock != nil {
				tt.setupMock(orgSubRepo)
			}

			service := NewSubscriptionService(
				subRepo,
				orgSubRepo,
				accessRepo,
				auditRepo,
				userRepo,
				orgRepo,
			)

			err := service.MarkOrganizationSubscriptionAsPaid(
				tt.adminUserID,
				tt.subscriptionID,
				tt.paymentDate,
				tt.durationDays,
			)

			// Special case for "subscription not found" - just expect an error
			if tt.name == "subscription not found" {
				if err == nil {
					t.Error("expected error for subscription not found")
				}
				return
			}

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Tests for SetUserSubscriptionPermanent
func TestSubscriptionService_SetUserSubscriptionPermanent(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID int64
		adminUserID    int64
		isPermanent    bool
		setupMock      func(*mockUserSubscriptionRepo)
		expectedError  error
	}{
		{
			name:           "set permanent true",
			subscriptionID: 1,
			adminUserID:    1,
			isPermanent:    true,
			setupMock: func(repo *mockUserSubscriptionRepo) {
				repo.subscriptions[1] = &domain.UserSubscription{
					ID:              1,
					UserID:          2,
					IsPermanentFree: false,
				}
			},
			expectedError: nil,
		},
		{
			name:           "set permanent false",
			subscriptionID: 1,
			adminUserID:    1,
			isPermanent:    false,
			setupMock: func(repo *mockUserSubscriptionRepo) {
				repo.subscriptions[1] = &domain.UserSubscription{
					ID:              1,
					UserID:          2,
					IsPermanentFree: true,
				}
			},
			expectedError: nil,
		},
		{
			name:           "cannot modify own subscription",
			subscriptionID: 1,
			adminUserID:    2,
			isPermanent:    true,
			setupMock: func(repo *mockUserSubscriptionRepo) {
				repo.subscriptions[1] = &domain.UserSubscription{
					ID:     1,
					UserID: 2,
				}
			},
			expectedError: ErrCannotModifyOwnSubscription,
		},
		{
			name:           "subscription not found",
			subscriptionID: 999,
			adminUserID:    1,
			isPermanent:    true,
			setupMock:      func(repo *mockUserSubscriptionRepo) {},
			expectedError:  nil, // Will check for any error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subRepo := newMockUserSubscriptionRepo()
			userRepo := newMockUserRepo()
			orgRepo := newMockOrganizationRepo()
			orgSubRepo := newMockOrganizationSubscriptionRepo()
			accessRepo := newMockSubscriptionAccessRepo()
			auditRepo := newMockAuditLogRepo()

			if tt.setupMock != nil {
				tt.setupMock(subRepo)
			}

			service := NewSubscriptionService(
				subRepo,
				orgSubRepo,
				accessRepo,
				auditRepo,
				userRepo,
				orgRepo,
			)

			err := service.SetUserSubscriptionPermanent(
				tt.adminUserID,
				tt.subscriptionID,
				tt.isPermanent,
			)

			// Special case for "subscription not found" - just expect an error
			if tt.name == "subscription not found" {
				if err == nil {
					t.Error("expected error for subscription not found")
				}
				return
			}

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify subscription was updated
			sub := subRepo.subscriptions[tt.subscriptionID]
			if sub.IsPermanentFree != tt.isPermanent {
				t.Errorf("expected IsPermanentFree %v, got %v", tt.isPermanent, sub.IsPermanentFree)
			}
		})
	}
}

// Tests for CancelOrganizationSubscription
func TestSubscriptionService_CancelOrganizationSubscription(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID int64
		adminUserID    int64
		reason         string
		setupMock      func(*mockOrganizationSubscriptionRepo)
		expectedError  error
	}{
		{
			name:           "successful cancellation",
			subscriptionID: 1,
			adminUserID:    1,
			reason:         "Organization closed",
			setupMock: func(repo *mockOrganizationSubscriptionRepo) {
				endDate := time.Now().Add(30 * 24 * time.Hour)
				repo.subscriptions[1] = &domain.OrganizationSubscription{
					ID:               1,
					OrganizationID:   10,
					SubscriptionType: domain.SubscriptionTypeAnnual,
					Status:           domain.SubscriptionStatusActive,
					StartDate:        time.Now(),
					EndDate:          &endDate,
				}
			},
			expectedError: nil,
		},
		{
			name:           "subscription not found",
			subscriptionID: 999,
			adminUserID:    1,
			reason:         "Test",
			setupMock:      func(repo *mockOrganizationSubscriptionRepo) {},
			expectedError:  nil, // Will check for any error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subRepo := newMockUserSubscriptionRepo()
			userRepo := newMockUserRepo()
			orgRepo := newMockOrganizationRepo()
			orgSubRepo := newMockOrganizationSubscriptionRepo()
			accessRepo := newMockSubscriptionAccessRepo()
			auditRepo := newMockAuditLogRepo()

			if tt.setupMock != nil {
				tt.setupMock(orgSubRepo)
			}

			service := NewSubscriptionService(
				subRepo,
				orgSubRepo,
				accessRepo,
				auditRepo,
				userRepo,
				orgRepo,
			)

			err := service.CancelOrganizationSubscription(
				tt.adminUserID,
				tt.subscriptionID,
				tt.reason,
			)

			// Special case for "subscription not found" - just expect an error
			if tt.name == "subscription not found" {
				if err == nil {
					t.Error("expected error for subscription not found")
				}
				return
			}

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify subscription was cancelled
			sub := orgSubRepo.subscriptions[tt.subscriptionID]
			if sub.Status != domain.SubscriptionStatusCancelled {
				t.Errorf("expected Status cancelled, got %s", sub.Status)
			}
		})
	}
}

// Tests for SetOrganizationSubscriptionPermanent
func TestSubscriptionService_SetOrganizationSubscriptionPermanent(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID int64
		adminUserID    int64
		isPermanent    bool
		setupMock      func(*mockOrganizationSubscriptionRepo)
		expectedError  error
	}{
		{
			name:           "set permanent true",
			subscriptionID: 1,
			adminUserID:    1,
			isPermanent:    true,
			setupMock: func(repo *mockOrganizationSubscriptionRepo) {
				repo.subscriptions[1] = &domain.OrganizationSubscription{
					ID:              1,
					OrganizationID:  10,
					IsPermanentFree: false,
				}
			},
			expectedError: nil,
		},
		{
			name:           "set permanent false",
			subscriptionID: 1,
			adminUserID:    1,
			isPermanent:    false,
			setupMock: func(repo *mockOrganizationSubscriptionRepo) {
				repo.subscriptions[1] = &domain.OrganizationSubscription{
					ID:              1,
					OrganizationID:  10,
					IsPermanentFree: true,
				}
			},
			expectedError: nil,
		},
		{
			name:           "subscription not found",
			subscriptionID: 999,
			adminUserID:    1,
			isPermanent:    true,
			setupMock:      func(repo *mockOrganizationSubscriptionRepo) {},
			expectedError:  nil, // Will check for any error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subRepo := newMockUserSubscriptionRepo()
			userRepo := newMockUserRepo()
			orgRepo := newMockOrganizationRepo()
			orgSubRepo := newMockOrganizationSubscriptionRepo()
			accessRepo := newMockSubscriptionAccessRepo()
			auditRepo := newMockAuditLogRepo()

			if tt.setupMock != nil {
				tt.setupMock(orgSubRepo)
			}

			service := NewSubscriptionService(
				subRepo,
				orgSubRepo,
				accessRepo,
				auditRepo,
				userRepo,
				orgRepo,
			)

			err := service.SetOrganizationSubscriptionPermanent(
				tt.adminUserID,
				tt.subscriptionID,
				tt.isPermanent,
			)

			// Special case for "subscription not found" - just expect an error
			if tt.name == "subscription not found" {
				if err == nil {
					t.Error("expected error for subscription not found")
				}
				return
			}

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify subscription was updated
			sub := orgSubRepo.subscriptions[tt.subscriptionID]
			if sub.IsPermanentFree != tt.isPermanent {
				t.Errorf("expected IsPermanentFree %v, got %v", tt.isPermanent, sub.IsPermanentFree)
			}
		})
	}
}

// Tests for CheckUserAccess
func TestSubscriptionService_CheckUserAccess(t *testing.T) {
	tests := []struct {
		name              string
		userID            int64
		setupMock         func(*mockSubscriptionAccessRepo)
		expectedHasAccess bool
	}{
		{
			name:   "user has access",
			userID: 1,
			setupMock: func(repo *mockSubscriptionAccessRepo) {
				repo.accessResults[1] = &domain.SubscriptionAccessResult{
					HasAccess: true,
					Source:    "user",
				}
			},
			expectedHasAccess: true,
		},
		{
			name:   "user has no access",
			userID: 2,
			setupMock: func(repo *mockSubscriptionAccessRepo) {
				repo.accessResults[2] = &domain.SubscriptionAccessResult{
					HasAccess: false,
					Source:    "none",
				}
			},
			expectedHasAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subRepo := newMockUserSubscriptionRepo()
			userRepo := newMockUserRepo()
			orgRepo := newMockOrganizationRepo()
			orgSubRepo := newMockOrganizationSubscriptionRepo()
			accessRepo := newMockSubscriptionAccessRepo()
			auditRepo := newMockAuditLogRepo()

			if tt.setupMock != nil {
				tt.setupMock(accessRepo)
			}

			service := NewSubscriptionService(
				subRepo,
				orgSubRepo,
				accessRepo,
				auditRepo,
				userRepo,
				orgRepo,
			)

			result, err := service.CheckUserAccess(tt.userID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.HasAccess != tt.expectedHasAccess {
				t.Errorf("expected HasAccess %v, got %v", tt.expectedHasAccess, result.HasAccess)
			}
		})
	}
}

// Tests for GetUserSubscriptions
func TestSubscriptionService_GetUserSubscriptions(t *testing.T) {
	subRepo := newMockUserSubscriptionRepo()
	userRepo := newMockUserRepo()
	orgRepo := newMockOrganizationRepo()
	orgSubRepo := newMockOrganizationSubscriptionRepo()
	accessRepo := newMockSubscriptionAccessRepo()
	auditRepo := newMockAuditLogRepo()

	// Add some subscriptions
	subRepo.subscriptions[1] = &domain.UserSubscription{ID: 1, UserID: 5}
	subRepo.subscriptions[2] = &domain.UserSubscription{ID: 2, UserID: 5}
	subRepo.subscriptions[3] = &domain.UserSubscription{ID: 3, UserID: 6}

	service := NewSubscriptionService(
		subRepo,
		orgSubRepo,
		accessRepo,
		auditRepo,
		userRepo,
		orgRepo,
	)

	subs, err := service.GetUserSubscriptions(5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(subs) != 2 {
		t.Errorf("expected 2 subscriptions, got %d", len(subs))
	}
}

// Tests for GetOrganizationSubscriptions
func TestSubscriptionService_GetOrganizationSubscriptions(t *testing.T) {
	subRepo := newMockUserSubscriptionRepo()
	userRepo := newMockUserRepo()
	orgRepo := newMockOrganizationRepo()
	orgSubRepo := newMockOrganizationSubscriptionRepo()
	accessRepo := newMockSubscriptionAccessRepo()
	auditRepo := newMockAuditLogRepo()

	// Add some subscriptions
	orgSubRepo.subscriptions[1] = &domain.OrganizationSubscription{ID: 1, OrganizationID: 10}
	orgSubRepo.subscriptions[2] = &domain.OrganizationSubscription{ID: 2, OrganizationID: 10}
	orgSubRepo.subscriptions[3] = &domain.OrganizationSubscription{ID: 3, OrganizationID: 20}

	service := NewSubscriptionService(
		subRepo,
		orgSubRepo,
		accessRepo,
		auditRepo,
		userRepo,
		orgRepo,
	)

	subs, err := service.GetOrganizationSubscriptions(10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(subs) != 2 {
		t.Errorf("expected 2 subscriptions, got %d", len(subs))
	}
}

// Tests for ExpireOverdueSubscriptions
func TestSubscriptionService_ExpireOverdueSubscriptions(t *testing.T) {
	subRepo := newMockUserSubscriptionRepo()
	userRepo := newMockUserRepo()
	orgRepo := newMockOrganizationRepo()
	orgSubRepo := newMockOrganizationSubscriptionRepo()
	accessRepo := newMockSubscriptionAccessRepo()
	auditRepo := newMockAuditLogRepo()

	service := NewSubscriptionService(
		subRepo,
		orgSubRepo,
		accessRepo,
		auditRepo,
		userRepo,
		orgRepo,
	)

	count, err := service.ExpireOverdueSubscriptions()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// Currently not implemented, should return 0
	if count != 0 {
		t.Errorf("expected 0 expired subscriptions, got %d", count)
	}
}

// Tests for ListAllUserSubscriptions
func TestSubscriptionService_ListAllUserSubscriptions(t *testing.T) {
	subRepo := newMockUserSubscriptionRepo()
	userRepo := newMockUserRepo()
	orgRepo := newMockOrganizationRepo()
	orgSubRepo := newMockOrganizationSubscriptionRepo()
	accessRepo := newMockSubscriptionAccessRepo()
	auditRepo := newMockAuditLogRepo()

	// Add some subscriptions
	subRepo.subscriptions[1] = &domain.UserSubscription{ID: 1, UserID: 1}
	subRepo.subscriptions[2] = &domain.UserSubscription{ID: 2, UserID: 2}
	subRepo.subscriptions[3] = &domain.UserSubscription{ID: 3, UserID: 3}

	service := NewSubscriptionService(
		subRepo,
		orgSubRepo,
		accessRepo,
		auditRepo,
		userRepo,
		orgRepo,
	)

	subs, err := service.ListAllUserSubscriptions()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(subs) != 3 {
		t.Errorf("expected 3 subscriptions, got %d", len(subs))
	}
}

// Tests for ListAllOrganizationSubscriptions
func TestSubscriptionService_ListAllOrganizationSubscriptions(t *testing.T) {
	subRepo := newMockUserSubscriptionRepo()
	userRepo := newMockUserRepo()
	orgRepo := newMockOrganizationRepo()
	orgSubRepo := newMockOrganizationSubscriptionRepo()
	accessRepo := newMockSubscriptionAccessRepo()
	auditRepo := newMockAuditLogRepo()

	// Add some subscriptions
	orgSubRepo.subscriptions[1] = &domain.OrganizationSubscription{ID: 1, OrganizationID: 10}
	orgSubRepo.subscriptions[2] = &domain.OrganizationSubscription{ID: 2, OrganizationID: 20}

	service := NewSubscriptionService(
		subRepo,
		orgSubRepo,
		accessRepo,
		auditRepo,
		userRepo,
		orgRepo,
	)

	subs, err := service.ListAllOrganizationSubscriptions()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(subs) != 2 {
		t.Errorf("expected 2 subscriptions, got %d", len(subs))
	}
}

// Tests for ListExpiringUserSubscriptions
func TestSubscriptionService_ListExpiringUserSubscriptions(t *testing.T) {
	subRepo := newMockUserSubscriptionRepo()
	userRepo := newMockUserRepo()
	orgRepo := newMockOrganizationRepo()
	orgSubRepo := newMockOrganizationSubscriptionRepo()
	accessRepo := newMockSubscriptionAccessRepo()
	auditRepo := newMockAuditLogRepo()

	now := time.Now()

	// Subscription expiring in 10 days (should be included)
	endDate10 := now.AddDate(0, 0, 10)
	subRepo.subscriptions[1] = &domain.UserSubscription{
		ID:               1,
		UserID:           1,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now.AddDate(0, -1, 0),
		EndDate:          &endDate10,
	}

	// Subscription expiring in 45 days (should NOT be included for 30 days)
	endDate45 := now.AddDate(0, 0, 45)
	subRepo.subscriptions[2] = &domain.UserSubscription{
		ID:               2,
		UserID:           2,
		SubscriptionType: domain.SubscriptionTypeAnnual,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now.AddDate(0, -10, 0),
		EndDate:          &endDate45,
	}

	// Permanent free (should NOT be included)
	subRepo.subscriptions[3] = &domain.UserSubscription{
		ID:               3,
		UserID:           3,
		SubscriptionType: domain.SubscriptionTypeFree,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  true,
		StartDate:        now,
	}

	service := NewSubscriptionService(
		subRepo,
		orgSubRepo,
		accessRepo,
		auditRepo,
		userRepo,
		orgRepo,
	)

	// Test with 30 days
	subs, err := service.ListExpiringUserSubscriptions(30)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(subs) != 1 {
		t.Errorf("expected 1 subscription, got %d", len(subs))
	}

	// Test with 0 days (should default to 30)
	subs2, err := service.ListExpiringUserSubscriptions(0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(subs2) != 1 {
		t.Errorf("expected 1 subscription with default days, got %d", len(subs2))
	}

	// Test with 60 days (should include both)
	subs3, err := service.ListExpiringUserSubscriptions(60)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(subs3) != 2 {
		t.Errorf("expected 2 subscriptions with 60 days, got %d", len(subs3))
	}
}

// Tests for ListExpiredUserSubscriptions
func TestSubscriptionService_ListExpiredUserSubscriptions(t *testing.T) {
	subRepo := newMockUserSubscriptionRepo()
	userRepo := newMockUserRepo()
	orgRepo := newMockOrganizationRepo()
	orgSubRepo := newMockOrganizationSubscriptionRepo()
	accessRepo := newMockSubscriptionAccessRepo()
	auditRepo := newMockAuditLogRepo()

	now := time.Now()

	// Expired subscription (status = expired)
	endDatePast := now.AddDate(0, 0, -10)
	subRepo.subscriptions[1] = &domain.UserSubscription{
		ID:               1,
		UserID:           1,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusExpired,
		IsPermanentFree:  false,
		StartDate:        now.AddDate(0, -2, 0),
		EndDate:          &endDatePast,
	}

	// Overdue subscription (active but past end date)
	endDatePast2 := now.AddDate(0, 0, -5)
	subRepo.subscriptions[2] = &domain.UserSubscription{
		ID:               2,
		UserID:           2,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now.AddDate(0, -1, 0),
		EndDate:          &endDatePast2,
	}

	// Active subscription with future end date (should NOT be included)
	endDateFuture := now.AddDate(0, 0, 30)
	subRepo.subscriptions[3] = &domain.UserSubscription{
		ID:               3,
		UserID:           3,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now,
		EndDate:          &endDateFuture,
	}

	service := NewSubscriptionService(
		subRepo,
		orgSubRepo,
		accessRepo,
		auditRepo,
		userRepo,
		orgRepo,
	)

	subs, err := service.ListExpiredUserSubscriptions()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(subs) != 2 {
		t.Errorf("expected 2 subscriptions, got %d", len(subs))
	}
}

// Tests for ListExpiringOrganizationSubscriptions
func TestSubscriptionService_ListExpiringOrganizationSubscriptions(t *testing.T) {
	subRepo := newMockUserSubscriptionRepo()
	userRepo := newMockUserRepo()
	orgRepo := newMockOrganizationRepo()
	orgSubRepo := newMockOrganizationSubscriptionRepo()
	accessRepo := newMockSubscriptionAccessRepo()
	auditRepo := newMockAuditLogRepo()

	now := time.Now()

	// Subscription expiring in 15 days
	endDate15 := now.AddDate(0, 0, 15)
	orgSubRepo.subscriptions[1] = &domain.OrganizationSubscription{
		ID:               1,
		OrganizationID:   10,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now.AddDate(0, -1, 0),
		EndDate:          &endDate15,
	}

	// Subscription expiring in 50 days
	endDate50 := now.AddDate(0, 0, 50)
	orgSubRepo.subscriptions[2] = &domain.OrganizationSubscription{
		ID:               2,
		OrganizationID:   20,
		SubscriptionType: domain.SubscriptionTypeAnnual,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now.AddDate(-1, 0, 0),
		EndDate:          &endDate50,
	}

	service := NewSubscriptionService(
		subRepo,
		orgSubRepo,
		accessRepo,
		auditRepo,
		userRepo,
		orgRepo,
	)

	// Test with 30 days
	subs, err := service.ListExpiringOrganizationSubscriptions(30)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(subs) != 1 {
		t.Errorf("expected 1 subscription, got %d", len(subs))
	}

	// Test with 60 days
	subs2, err := service.ListExpiringOrganizationSubscriptions(60)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(subs2) != 2 {
		t.Errorf("expected 2 subscriptions, got %d", len(subs2))
	}
}

// Tests for ListExpiredOrganizationSubscriptions
func TestSubscriptionService_ListExpiredOrganizationSubscriptions(t *testing.T) {
	subRepo := newMockUserSubscriptionRepo()
	userRepo := newMockUserRepo()
	orgRepo := newMockOrganizationRepo()
	orgSubRepo := newMockOrganizationSubscriptionRepo()
	accessRepo := newMockSubscriptionAccessRepo()
	auditRepo := newMockAuditLogRepo()

	now := time.Now()

	// Expired subscription
	endDatePast := now.AddDate(0, 0, -10)
	orgSubRepo.subscriptions[1] = &domain.OrganizationSubscription{
		ID:               1,
		OrganizationID:   10,
		SubscriptionType: domain.SubscriptionTypeMonthly,
		Status:           domain.SubscriptionStatusExpired,
		IsPermanentFree:  false,
		StartDate:        now.AddDate(0, -2, 0),
		EndDate:          &endDatePast,
	}

	// Overdue subscription
	endDatePast2 := now.AddDate(0, 0, -3)
	orgSubRepo.subscriptions[2] = &domain.OrganizationSubscription{
		ID:               2,
		OrganizationID:   20,
		SubscriptionType: domain.SubscriptionTypeAnnual,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now.AddDate(-1, 0, 0),
		EndDate:          &endDatePast2,
	}

	// Active subscription (should NOT be included)
	endDateFuture := now.AddDate(0, 0, 60)
	orgSubRepo.subscriptions[3] = &domain.OrganizationSubscription{
		ID:               3,
		OrganizationID:   30,
		SubscriptionType: domain.SubscriptionTypeAnnual,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  false,
		StartDate:        now,
		EndDate:          &endDateFuture,
	}

	service := NewSubscriptionService(
		subRepo,
		orgSubRepo,
		accessRepo,
		auditRepo,
		userRepo,
		orgRepo,
	)

	subs, err := service.ListExpiredOrganizationSubscriptions()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(subs) != 2 {
		t.Errorf("expected 2 subscriptions, got %d", len(subs))
	}
}
