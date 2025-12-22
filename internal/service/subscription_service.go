package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

var (
	ErrSubscriptionNotFound           = errors.New("subscription not found")
	ErrCannotModifyOwnSubscription    = errors.New("admins cannot modify their own subscriptions")
	ErrActiveSubscriptionExists       = errors.New("user already has an active subscription")
	ErrActiveOrgSubscriptionExists    = errors.New("organization already has an active subscription")
	ErrCannotMarkFreeSubscriptionPaid = errors.New("cannot mark free subscription as paid")
)

// SubscriptionService handles subscription-related business logic
type SubscriptionService struct {
	userSubRepo  domain.UserSubscriptionRepository
	orgSubRepo   domain.OrganizationSubscriptionRepository
	accessRepo   domain.SubscriptionAccessRepository
	auditLogRepo domain.AuditLogRepository
	userRepo     domain.UserRepository
	orgRepo      domain.OrganizationRepository
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(
	userSubRepo domain.UserSubscriptionRepository,
	orgSubRepo domain.OrganizationSubscriptionRepository,
	accessRepo domain.SubscriptionAccessRepository,
	auditLogRepo domain.AuditLogRepository,
	userRepo domain.UserRepository,
	orgRepo domain.OrganizationRepository,
) *SubscriptionService {
	return &SubscriptionService{
		userSubRepo:  userSubRepo,
		orgSubRepo:   orgSubRepo,
		accessRepo:   accessRepo,
		auditLogRepo: auditLogRepo,
		userRepo:     userRepo,
		orgRepo:      orgRepo,
	}
}

// CreateUserSubscription creates a new user subscription (admin only)
func (s *SubscriptionService) CreateUserSubscription(
	adminUserID int64,
	userID int64,
	subscriptionType string,
	isPermanentFree bool,
	notes string,
) (*domain.UserSubscription, error) {
	// Admins cannot modify their own subscriptions
	if adminUserID == userID {
		return nil, ErrCannotModifyOwnSubscription
	}

	// Check if user already has an active subscription
	existing, err := s.userSubRepo.GetActiveByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing subscription: %w", err)
	}
	if existing != nil {
		return nil, ErrActiveSubscriptionExists
	}

	// Validate subscription type
	subType := domain.SubscriptionType(subscriptionType)
	if subType != domain.SubscriptionTypeFree &&
		subType != domain.SubscriptionTypeMonthly &&
		subType != domain.SubscriptionTypeAnnual {
		return nil, fmt.Errorf("invalid subscription type: %s", subscriptionType)
	}

	now := time.Now()
	var endDate *time.Time
	var nextBillingDate *time.Time

	// Calculate end_date and next_billing_date based on type and permanent free status
	if !isPermanentFree {
		switch subType {
		case domain.SubscriptionTypeMonthly:
			// Monthly subscriptions start with 30-day trial
			end := now.AddDate(0, 0, 30)
			endDate = &end
			nextBillingDate = &end
		case domain.SubscriptionTypeAnnual:
			// Annual subscriptions start with 365-day trial
			end := now.AddDate(0, 0, 365)
			endDate = &end
			nextBillingDate = &end
		}
		// Free subscriptions have no end date
	}

	// Create the subscription
	sub := &domain.UserSubscription{
		UserID:           userID,
		SubscriptionType: subType,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  isPermanentFree,
		StartDate:        now,
		EndDate:          endDate,
		NextBillingDate:  nextBillingDate,
		Notes:            &notes,
		CreatedAt:        now,
		UpdatedAt:        now,
		CreatedByUserID:  &adminUserID,
	}

	err = s.userSubRepo.Create(sub)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Log to audit trail
	details := fmt.Sprintf("Created %s subscription for user %d (permanent_free=%v)", subscriptionType, userID, isPermanentFree)
	targetUserID := userID
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:       &adminUserID,
		TargetUserID: &targetUserID,
		EventType:    domain.EventSubscriptionCreated,
		Details:      &details,
		CreatedAt:    now,
	})

	return sub, nil
}

// CreateOrganizationSubscription creates a new organization subscription (admin only)
func (s *SubscriptionService) CreateOrganizationSubscription(
	adminUserID int64,
	organizationID int64,
	subscriptionType string,
	isPermanentFree bool,
	notes string,
) (*domain.OrganizationSubscription, error) {
	// Check if organization already has an active subscription
	existing, err := s.orgSubRepo.GetActiveByOrganizationID(organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing subscription: %w", err)
	}
	if existing != nil {
		return nil, ErrActiveOrgSubscriptionExists
	}

	// Validate subscription type
	subType := domain.SubscriptionType(subscriptionType)
	if subType != domain.SubscriptionTypeFree &&
		subType != domain.SubscriptionTypeMonthly &&
		subType != domain.SubscriptionTypeAnnual {
		return nil, fmt.Errorf("invalid subscription type: %s", subscriptionType)
	}

	now := time.Now()
	var endDate *time.Time
	var nextBillingDate *time.Time

	// Calculate end_date and next_billing_date based on type and permanent free status
	if !isPermanentFree {
		switch subType {
		case domain.SubscriptionTypeMonthly:
			end := now.AddDate(0, 0, 30)
			endDate = &end
			nextBillingDate = &end
		case domain.SubscriptionTypeAnnual:
			end := now.AddDate(0, 0, 365)
			endDate = &end
			nextBillingDate = &end
		}
	}

	// Create the subscription
	sub := &domain.OrganizationSubscription{
		OrganizationID:   organizationID,
		SubscriptionType: subType,
		Status:           domain.SubscriptionStatusActive,
		IsPermanentFree:  isPermanentFree,
		StartDate:        now,
		EndDate:          endDate,
		NextBillingDate:  nextBillingDate,
		Notes:            &notes,
		CreatedAt:        now,
		UpdatedAt:        now,
		CreatedByUserID:  &adminUserID,
	}

	err = s.orgSubRepo.Create(sub)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Log to audit trail
	details := fmt.Sprintf("Created %s subscription for organization %d (permanent_free=%v)", subscriptionType, organizationID, isPermanentFree)
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:    &adminUserID,
		EventType: domain.EventOrgSubscriptionCreated,
		Details:   &details,
		CreatedAt: now,
	})

	return sub, nil
}

// MarkUserSubscriptionAsPaid marks a user subscription as paid (admin only)
// paymentDateStr: optional ISO date string (defaults to now)
// durationDays: optional custom duration in days (defaults to subscription type)
func (s *SubscriptionService) MarkUserSubscriptionAsPaid(adminUserID int64, subscriptionID int64, paymentDateStr *string, durationDays *int) error {
	// Get the subscription
	sub, err := s.userSubRepo.GetByID(subscriptionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSubscriptionNotFound
		}
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if sub == nil {
		return ErrSubscriptionNotFound
	}

	// Admins cannot modify their own subscriptions
	if adminUserID == sub.UserID {
		return ErrCannotModifyOwnSubscription
	}

	// Cannot mark free subscriptions as paid
	if sub.SubscriptionType == domain.SubscriptionTypeFree {
		return ErrCannotMarkFreeSubscriptionPaid
	}

	// Parse payment date or use now
	paymentDate := time.Now()
	if paymentDateStr != nil && *paymentDateStr != "" {
		parsed, err := time.Parse("2006-01-02", *paymentDateStr)
		if err == nil {
			paymentDate = parsed
		}
	}

	err = s.userSubRepo.MarkAsPaid(subscriptionID, paymentDate, adminUserID, durationDays)
	if err != nil {
		return fmt.Errorf("failed to mark subscription as paid: %w", err)
	}

	// Log to audit trail
	durationStr := "default"
	if durationDays != nil {
		durationStr = fmt.Sprintf("%d days", *durationDays)
	}
	details := fmt.Sprintf("Marked subscription %d as paid for user %d (duration: %s)", subscriptionID, sub.UserID, durationStr)
	targetUserID := sub.UserID
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:       &adminUserID,
		TargetUserID: &targetUserID,
		EventType:    domain.EventSubscriptionMarkedPaid,
		Details:      &details,
		CreatedAt:    time.Now(),
	})

	return nil
}

// MarkOrganizationSubscriptionAsPaid marks an organization subscription as paid (admin only)
// paymentDateStr: optional ISO date string (defaults to now)
// durationDays: optional custom duration in days (defaults to subscription type)
func (s *SubscriptionService) MarkOrganizationSubscriptionAsPaid(adminUserID int64, subscriptionID int64, paymentDateStr *string, durationDays *int) error {
	// Get the subscription
	sub, err := s.orgSubRepo.GetByID(subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if sub == nil {
		return ErrSubscriptionNotFound
	}

	// Cannot mark free subscriptions as paid
	if sub.SubscriptionType == domain.SubscriptionTypeFree {
		return ErrCannotMarkFreeSubscriptionPaid
	}

	// Parse payment date or use now
	paymentDate := time.Now()
	if paymentDateStr != nil && *paymentDateStr != "" {
		parsed, err := time.Parse("2006-01-02", *paymentDateStr)
		if err == nil {
			paymentDate = parsed
		}
	}

	err = s.orgSubRepo.MarkAsPaid(subscriptionID, paymentDate, adminUserID, durationDays)
	if err != nil {
		return fmt.Errorf("failed to mark subscription as paid: %w", err)
	}

	// Log to audit trail
	durationStr := "default"
	if durationDays != nil {
		durationStr = fmt.Sprintf("%d days", *durationDays)
	}
	details := fmt.Sprintf("Marked subscription %d as paid for organization %d (duration: %s)", subscriptionID, sub.OrganizationID, durationStr)
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:    &adminUserID,
		EventType: domain.EventOrgSubscriptionMarkedPaid,
		Details:   &details,
		CreatedAt: time.Now(),
	})

	return nil
}

// CancelUserSubscription cancels a user subscription (admin only)
func (s *SubscriptionService) CancelUserSubscription(adminUserID int64, subscriptionID int64, reason string) error {
	// Get the subscription
	sub, err := s.userSubRepo.GetByID(subscriptionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSubscriptionNotFound
		}
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if sub == nil {
		return ErrSubscriptionNotFound
	}

	// Admins cannot modify their own subscriptions
	if adminUserID == sub.UserID {
		return ErrCannotModifyOwnSubscription
	}

	err = s.userSubRepo.Cancel(subscriptionID, reason, adminUserID)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	// Log to audit trail
	now := time.Now()
	details := fmt.Sprintf("Cancelled subscription %d for user %d (reason: %s)", subscriptionID, sub.UserID, reason)
	targetUserID := sub.UserID
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:       &adminUserID,
		TargetUserID: &targetUserID,
		EventType:    domain.EventSubscriptionCancelled,
		Details:      &details,
		CreatedAt:    now,
	})

	return nil
}

// SetUserSubscriptionPermanent sets the permanent free status of a user subscription (admin only)
func (s *SubscriptionService) SetUserSubscriptionPermanent(adminUserID int64, subscriptionID int64, isPermanent bool) error {
	// Get the subscription
	sub, err := s.userSubRepo.GetByID(subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if sub == nil {
		return ErrSubscriptionNotFound
	}

	// Admins cannot modify their own subscriptions
	if adminUserID == sub.UserID {
		return ErrCannotModifyOwnSubscription
	}

	// Update the permanent status
	sub.IsPermanentFree = isPermanent
	sub.UpdatedAt = time.Now()

	err = s.userSubRepo.Update(sub)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Log to audit trail
	status := "removed permanent status from"
	if isPermanent {
		status = "set permanent status for"
	}
	details := fmt.Sprintf("Admin %s subscription %d for user %d", status, subscriptionID, sub.UserID)
	targetUserID := sub.UserID
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:       &adminUserID,
		TargetUserID: &targetUserID,
		EventType:    domain.EventSubscriptionUpdated,
		Details:      &details,
		CreatedAt:    time.Now(),
	})

	return nil
}

// CancelOrganizationSubscription cancels an organization subscription (admin only)
func (s *SubscriptionService) CancelOrganizationSubscription(adminUserID int64, subscriptionID int64, reason string) error {
	// Get the subscription
	sub, err := s.orgSubRepo.GetByID(subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if sub == nil {
		return ErrSubscriptionNotFound
	}

	err = s.orgSubRepo.Cancel(subscriptionID, reason, adminUserID)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	// Log to audit trail
	now := time.Now()
	details := fmt.Sprintf("Cancelled subscription %d for organization %d (reason: %s)", subscriptionID, sub.OrganizationID, reason)
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:    &adminUserID,
		EventType: domain.EventOrgSubscriptionCancelled,
		Details:   &details,
		CreatedAt: now,
	})

	return nil
}

// SetOrganizationSubscriptionPermanent sets the permanent free status of an organization subscription (admin only)
func (s *SubscriptionService) SetOrganizationSubscriptionPermanent(adminUserID int64, subscriptionID int64, isPermanent bool) error {
	// Get the subscription
	sub, err := s.orgSubRepo.GetByID(subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if sub == nil {
		return ErrSubscriptionNotFound
	}

	// Update the permanent status
	sub.IsPermanentFree = isPermanent
	sub.UpdatedAt = time.Now()

	err = s.orgSubRepo.Update(sub)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Log to audit trail
	status := "removed permanent status from"
	if isPermanent {
		status = "set permanent status for"
	}
	details := fmt.Sprintf("Admin %s subscription %d for organization %d", status, subscriptionID, sub.OrganizationID)
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:    &adminUserID,
		EventType: domain.EventOrgSubscriptionUpdated,
		Details:   &details,
		CreatedAt: time.Now(),
	})

	return nil
}

// CheckUserAccess checks if user has active subscription (any user)
// This is the CRITICAL PERFORMANCE PATH - called on every authenticated request
func (s *SubscriptionService) CheckUserAccess(userID int64) (*domain.SubscriptionAccessResult, error) {
	return s.accessRepo.CheckUserAccess(userID)
}

// GetUserSubscriptionStatus gets full subscription status for a user (user or admin)
func (s *SubscriptionService) GetUserSubscriptionStatus(userID int64) (*domain.SubscriptionAccessResult, error) {
	return s.accessRepo.CheckUserAccess(userID)
}

// GetUserSubscriptions gets all subscriptions for a user (admin only)
func (s *SubscriptionService) GetUserSubscriptions(userID int64) ([]*domain.UserSubscription, error) {
	return s.userSubRepo.GetByUserID(userID)
}

// GetOrganizationSubscriptions gets all subscriptions for an organization (admin only)
func (s *SubscriptionService) GetOrganizationSubscriptions(orgID int64) ([]*domain.OrganizationSubscription, error) {
	return s.orgSubRepo.GetByOrganizationID(orgID)
}

// ExpireOverdueSubscriptions checks all subscriptions and marks expired ones (background job)
func (s *SubscriptionService) ExpireOverdueSubscriptions() (int, error) {
	// This would be called by a background job/cron
	// For now, we'll return 0 since this is not implemented yet
	// TODO: Implement background job to check and expire subscriptions
	return 0, nil
}

// ListAllUserSubscriptions returns all user subscriptions (admin only)
func (s *SubscriptionService) ListAllUserSubscriptions() ([]*domain.UserSubscription, error) {
	return s.userSubRepo.ListAll()
}

// ListAllOrganizationSubscriptions returns all organization subscriptions (admin only)
func (s *SubscriptionService) ListAllOrganizationSubscriptions() ([]*domain.OrganizationSubscription, error) {
	return s.orgSubRepo.ListAll()
}
