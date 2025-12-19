package service

import (
	"errors"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestSubscriptionService_CreateUserSubscription(t *testing.T) {
	tests := []struct {
		name              string
		adminUserID       int64
		targetUserID      int64
		subscriptionType  string
		isPermanentFree   bool
		notes             string
		setupMock         func(*mockUserSubscriptionRepo, *mockUserRepo)
		expectedError     error
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
			if !tt.isPermanentFree {
				if sub.EndDate == nil {
					t.Error("expected EndDate to be set for non-permanent subscription")
				} else {
					expectedDuration := 30 * 24 * time.Hour // Default for free/monthly
					if tt.subscriptionType == "annual" {
						expectedDuration = 365 * 24 * time.Hour
					}
					expectedEndDate := sub.StartDate.Add(expectedDuration)
					if sub.EndDate.Before(expectedEndDate.Add(-1*time.Minute)) || sub.EndDate.After(expectedEndDate.Add(1*time.Minute)) {
						t.Errorf("EndDate not set correctly, expected around %v, got %v", expectedEndDate, sub.EndDate)
					}
				}
			} else {
				if sub.EndDate != nil {
					t.Error("expected EndDate to be nil for permanent free subscription")
				}
			}
		})
	}
}

func TestSubscriptionService_MarkUserSubscriptionAsPaid(t *testing.T) {
	tests := []struct {
		name          string
		subscriptionID int64
		adminUserID   int64
		paymentDate   time.Time
		durationDays  *int
		setupMock     func(*mockUserSubscriptionRepo)
		expectedError error
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
		name             string
		userID           int64
		setupMock        func(*mockSubscriptionAccessRepo)
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
			expectedError: ErrActiveSubscriptionExists,
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
