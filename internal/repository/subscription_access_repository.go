package repository

import (
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// SubscriptionAccessRepository provides fast access checking
type SubscriptionAccessRepository struct {
	userSubRepo *SQLiteUserSubscriptionRepository
	orgSubRepo  *SQLiteOrganizationSubscriptionRepository
	orgRepo     *OrganizationRepository
}

// NewSubscriptionAccessRepository creates a new subscription access repository
func NewSubscriptionAccessRepository(
	userSubRepo *SQLiteUserSubscriptionRepository,
	orgSubRepo *SQLiteOrganizationSubscriptionRepository,
	orgRepo *OrganizationRepository,
) *SubscriptionAccessRepository {
	return &SubscriptionAccessRepository{
		userSubRepo: userSubRepo,
		orgSubRepo:  orgSubRepo,
		orgRepo:     orgRepo,
	}
}

// CheckUserAccess returns whether user has active subscription
// Checks both user-level and org-level subscriptions
// This is CRITICAL PERFORMANCE PATH - called on every authenticated request
func (r *SubscriptionAccessRepository) CheckUserAccess(userID int64) (*domain.SubscriptionAccessResult, error) {
	result := &domain.SubscriptionAccessResult{
		HasAccess: false,
		Source:    "none",
	}

	// Step 1: Check user-level subscription (fastest - single indexed query)
	userSub, err := r.userSubRepo.GetActiveByUserID(userID)
	if err != nil {
		return nil, err
	}

	if userSub != nil && isActiveSubscription(userSub.Status, userSub.IsPermanentFree, userSub.EndDate) {
		result.HasAccess = true
		result.Source = "user"
		result.UserSubscription = userSub
		result.ExpiresAt = calculateExpirationDate(userSub.IsPermanentFree, userSub.EndDate)
		return result, nil
	}

	// Step 2: Check organization subscriptions (requires join)
	orgIDs, err := r.orgRepo.GetUserOrganizationIDs(userID)
	if err != nil {
		return nil, err
	}

	// If user has organization subscriptions, check them
	var activeOrgSubs []*domain.OrganizationSubscription
	var earliestExpiration *time.Time

	for _, orgID := range orgIDs {
		orgSub, err := r.orgSubRepo.GetActiveByOrganizationID(orgID)
		if err != nil {
			return nil, err
		}

		if orgSub != nil && isActiveSubscription(orgSub.Status, orgSub.IsPermanentFree, orgSub.EndDate) {
			activeOrgSubs = append(activeOrgSubs, orgSub)

			// Track earliest expiration date among active org subscriptions
			expires := calculateExpirationDate(orgSub.IsPermanentFree, orgSub.EndDate)
			if expires != nil {
				if earliestExpiration == nil || expires.Before(*earliestExpiration) {
					earliestExpiration = expires
				}
			}
		}
	}

	// If we found active organization subscriptions, user has access
	if len(activeOrgSubs) > 0 {
		result.HasAccess = true
		result.Source = "organization"
		result.OrgSubscriptions = activeOrgSubs
		result.ExpiresAt = earliestExpiration

		// If user also had a subscription (even if inactive), include it in result
		if userSub != nil {
			result.UserSubscription = userSub
			if result.Source == "organization" && userSub != nil {
				result.Source = "both"
			}
		}

		return result, nil
	}

	// No active subscriptions found - return no access
	// Include inactive user subscription if it exists for status display
	if userSub != nil {
		result.UserSubscription = userSub
	}

	return result, nil
}

// isActiveSubscription checks if a subscription is currently active
func isActiveSubscription(status domain.SubscriptionStatus, isPermanentFree bool, endDate *time.Time) bool {
	// Must have active status
	if status != domain.SubscriptionStatusActive {
		return false
	}

	// Permanent free subscriptions never expire
	if isPermanentFree {
		return true
	}

	// If no end date, treat as active (shouldn't happen for non-permanent-free, but be safe)
	if endDate == nil {
		return true
	}

	// Check if end date is in the future
	return endDate.After(time.Now())
}

// calculateExpirationDate returns the expiration date for display
// Returns nil for permanent free subscriptions (never expire)
func calculateExpirationDate(isPermanentFree bool, endDate *time.Time) *time.Time {
	if isPermanentFree {
		return nil // Never expires
	}
	return endDate
}
