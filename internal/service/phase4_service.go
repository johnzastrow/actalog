package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

var (
	ErrDocumentNotFound       = errors.New("document not found")
	ErrUserDocumentNotFound   = errors.New("user document not found")
	ErrClassPackageNotFound   = errors.New("class package not found")
	ErrInsufficientCredits    = errors.New("insufficient credits")
	ErrWaitlistEntryNotFound  = errors.New("waitlist entry not found")
	ErrAlreadyOnWaitlist      = errors.New("user is already on the waitlist")
	ErrNotificationNotFound   = errors.New("notification not found")
	ErrDocumentAlreadyTracked = errors.New("document is already being tracked for this user")
)

// Phase4Service handles documents, credits, waitlist, and notifications
type Phase4Service struct {
	documentRepo     domain.DocumentRepository
	userDocumentRepo domain.UserDocumentRepository
	packageRepo      domain.ClassPackageRepository
	creditsRepo      domain.UserClassCreditsRepository
	waitlistRepo     domain.WaitlistRepository
	notificationRepo domain.ClassNotificationRepository
	reservationRepo  domain.ReservationRepository
	sessionRepo      domain.ClassSessionRepository
	orgRepo          domain.OrganizationRepository
	auditLogRepo     domain.AuditLogRepository
}

// NewPhase4Service creates a new Phase 4 service
func NewPhase4Service(
	documentRepo domain.DocumentRepository,
	userDocumentRepo domain.UserDocumentRepository,
	packageRepo domain.ClassPackageRepository,
	creditsRepo domain.UserClassCreditsRepository,
	waitlistRepo domain.WaitlistRepository,
	notificationRepo domain.ClassNotificationRepository,
	reservationRepo domain.ReservationRepository,
	sessionRepo domain.ClassSessionRepository,
	orgRepo domain.OrganizationRepository,
	auditLogRepo domain.AuditLogRepository,
) *Phase4Service {
	return &Phase4Service{
		documentRepo:     documentRepo,
		userDocumentRepo: userDocumentRepo,
		packageRepo:      packageRepo,
		creditsRepo:      creditsRepo,
		waitlistRepo:     waitlistRepo,
		notificationRepo: notificationRepo,
		reservationRepo:  reservationRepo,
		sessionRepo:      sessionRepo,
		orgRepo:          orgRepo,
		auditLogRepo:     auditLogRepo,
	}
}

// ============================================
// DOCUMENT METHODS
// ============================================

// CreateDocument creates a new document type
func (s *Phase4Service) CreateDocument(adminUserID int64, doc *domain.Document) error {
	// Validate organization exists
	org, err := s.orgRepo.GetByID(doc.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to verify organization: %w", err)
	}
	if org == nil {
		return ErrOrganizationNotFound
	}

	doc.IsActive = true
	if err := s.documentRepo.Create(doc); err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	s.logAudit(adminUserID, nil, domain.EventDocumentCreated, map[string]interface{}{
		"document_id":     doc.ID,
		"document_name":   doc.Name,
		"organization_id": doc.OrganizationID,
	})

	return nil
}

// GetDocumentByID retrieves a document by ID
func (s *Phase4Service) GetDocumentByID(id int64) (*domain.Document, error) {
	doc, err := s.documentRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	if doc == nil {
		return nil, ErrDocumentNotFound
	}
	return doc, nil
}

// GetDocumentsByOrganization retrieves all documents for an organization
func (s *Phase4Service) GetDocumentsByOrganization(orgID int64, includeInactive bool) ([]*domain.Document, error) {
	return s.documentRepo.GetByOrganizationID(orgID, includeInactive)
}

// GetRequiredDocuments retrieves all required documents for an organization
func (s *Phase4Service) GetRequiredDocuments(orgID int64) ([]*domain.Document, error) {
	return s.documentRepo.GetRequiredByOrganizationID(orgID)
}

// UpdateDocument updates a document
func (s *Phase4Service) UpdateDocument(adminUserID int64, doc *domain.Document) error {
	existing, err := s.documentRepo.GetByID(doc.ID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}
	if existing == nil {
		return ErrDocumentNotFound
	}

	if err := s.documentRepo.Update(doc); err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	s.logAudit(adminUserID, nil, domain.EventDocumentUpdated, map[string]interface{}{
		"document_id":   doc.ID,
		"document_name": doc.Name,
	})

	return nil
}

// DeleteDocument deletes a document
func (s *Phase4Service) DeleteDocument(adminUserID int64, id int64) error {
	doc, err := s.documentRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}
	if doc == nil {
		return ErrDocumentNotFound
	}

	if err := s.documentRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	s.logAudit(adminUserID, nil, domain.EventDocumentDeleted, map[string]interface{}{
		"document_id":   id,
		"document_name": doc.Name,
	})

	return nil
}

// ============================================
// USER DOCUMENT METHODS
// ============================================

// InitializeUserDocuments creates pending user_document records for all required documents
func (s *Phase4Service) InitializeUserDocuments(userID, orgID int64) error {
	requiredDocs, err := s.documentRepo.GetRequiredByOrganizationID(orgID)
	if err != nil {
		return fmt.Errorf("failed to get required documents: %w", err)
	}

	for _, doc := range requiredDocs {
		// Check if already exists
		existing, err := s.userDocumentRepo.GetByUserAndDocument(userID, doc.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check existing user document: %w", err)
		}
		if existing != nil {
			continue // Already exists
		}

		// Create pending record
		userDoc := &domain.UserDocument{
			UserID:     userID,
			DocumentID: doc.ID,
			Status:     domain.DocumentStatusPending,
		}
		if err := s.userDocumentRepo.Create(userDoc); err != nil {
			return fmt.Errorf("failed to create user document record: %w", err)
		}
	}

	return nil
}

// GetUserDocuments retrieves all document records for a user
func (s *Phase4Service) GetUserDocuments(userID int64) ([]*domain.UserDocument, error) {
	return s.userDocumentRepo.GetByUserID(userID)
}

// GetUserDocumentsByOrganization retrieves user's documents for a specific organization
func (s *Phase4Service) GetUserDocumentsByOrganization(userID, orgID int64) ([]*domain.UserDocument, error) {
	return s.userDocumentRepo.GetByUserAndOrganization(userID, orgID)
}

// GetPendingUserDocuments retrieves pending documents for a user in an organization
func (s *Phase4Service) GetPendingUserDocuments(userID, orgID int64) ([]*domain.UserDocument, error) {
	return s.userDocumentRepo.GetPendingByUserAndOrganization(userID, orgID)
}

// MarkDocumentCompleted marks a user document as completed
func (s *Phase4Service) MarkDocumentCompleted(adminUserID int64, userDocID int64) error {
	userDoc, err := s.userDocumentRepo.GetByID(userDocID)
	if err != nil {
		return fmt.Errorf("failed to get user document: %w", err)
	}
	if userDoc == nil {
		return ErrUserDocumentNotFound
	}

	if err := s.userDocumentRepo.MarkCompleted(userDocID, &adminUserID); err != nil {
		return fmt.Errorf("failed to mark document completed: %w", err)
	}

	s.logAudit(adminUserID, &userDoc.UserID, domain.EventUserDocumentCompleted, map[string]interface{}{
		"user_document_id": userDocID,
		"document_id":      userDoc.DocumentID,
	})

	return nil
}

// GetExpiringDocuments retrieves documents expiring within daysAhead days
func (s *Phase4Service) GetExpiringDocuments(daysAhead int) ([]*domain.UserDocument, error) {
	return s.userDocumentRepo.GetExpiringSoon(daysAhead)
}

// ============================================
// CLASS PACKAGE METHODS
// ============================================

// CreatePackage creates a new class package
func (s *Phase4Service) CreatePackage(adminUserID int64, pkg *domain.ClassPackage) error {
	// Validate organization exists
	org, err := s.orgRepo.GetByID(pkg.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to verify organization: %w", err)
	}
	if org == nil {
		return ErrOrganizationNotFound
	}

	pkg.IsActive = true
	if err := s.packageRepo.Create(pkg); err != nil {
		return fmt.Errorf("failed to create class package: %w", err)
	}

	s.logAudit(adminUserID, nil, domain.EventClassPackageCreated, map[string]interface{}{
		"package_id":      pkg.ID,
		"package_name":    pkg.Name,
		"credits":         pkg.Credits,
		"organization_id": pkg.OrganizationID,
	})

	return nil
}

// GetPackageByID retrieves a class package by ID
func (s *Phase4Service) GetPackageByID(id int64) (*domain.ClassPackage, error) {
	pkg, err := s.packageRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get class package: %w", err)
	}
	if pkg == nil {
		return nil, ErrClassPackageNotFound
	}
	return pkg, nil
}

// GetPackagesByOrganization retrieves all packages for an organization
func (s *Phase4Service) GetPackagesByOrganization(orgID int64, includeInactive bool) ([]*domain.ClassPackage, error) {
	return s.packageRepo.GetByOrganizationID(orgID, includeInactive)
}

// UpdatePackage updates a class package
func (s *Phase4Service) UpdatePackage(adminUserID int64, pkg *domain.ClassPackage) error {
	existing, err := s.packageRepo.GetByID(pkg.ID)
	if err != nil {
		return fmt.Errorf("failed to get class package: %w", err)
	}
	if existing == nil {
		return ErrClassPackageNotFound
	}

	if err := s.packageRepo.Update(pkg); err != nil {
		return fmt.Errorf("failed to update class package: %w", err)
	}

	s.logAudit(adminUserID, nil, domain.EventClassPackageUpdated, map[string]interface{}{
		"package_id":   pkg.ID,
		"package_name": pkg.Name,
	})

	return nil
}

// DeletePackage deletes a class package
func (s *Phase4Service) DeletePackage(adminUserID int64, id int64) error {
	pkg, err := s.packageRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get class package: %w", err)
	}
	if pkg == nil {
		return ErrClassPackageNotFound
	}

	if err := s.packageRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete class package: %w", err)
	}

	s.logAudit(adminUserID, nil, domain.EventClassPackageDeleted, map[string]interface{}{
		"package_id":   id,
		"package_name": pkg.Name,
	})

	return nil
}

// ============================================
// USER CREDITS METHODS
// ============================================

// PurchaseCredits adds credits to a user's account
func (s *Phase4Service) PurchaseCredits(adminUserID int64, userID, orgID int64, packageID *int64, credits int, notes string) error {
	// Calculate expiration and get credits from package if provided
	var expiresAt *time.Time
	creditsToAdd := credits
	if packageID != nil {
		pkg, err := s.packageRepo.GetByID(*packageID)
		if err != nil {
			return fmt.Errorf("failed to get package: %w", err)
		}
		if pkg == nil {
			return ErrClassPackageNotFound
		}
		// Use package credits if credits not explicitly provided
		if credits == 0 {
			creditsToAdd = pkg.Credits
		}
		if pkg.ValidityDays != nil {
			exp := time.Now().AddDate(0, 0, *pkg.ValidityDays)
			expiresAt = &exp
		}
	}

	creditRecord := &domain.UserClassCredits{
		UserID:         userID,
		OrganizationID: orgID,
		PackageID:      packageID,
		CreditsTotal:   creditsToAdd,
		CreditsUsed:    0,
		PurchasedAt:    time.Now(),
		ExpiresAt:      expiresAt,
		Notes:          &notes,
	}

	if err := s.creditsRepo.Create(creditRecord); err != nil {
		return fmt.Errorf("failed to create credits record: %w", err)
	}

	s.logAudit(adminUserID, &userID, domain.EventCreditsAdded, map[string]interface{}{
		"credits_id":      creditRecord.ID,
		"credits":         creditsToAdd,
		"organization_id": orgID,
		"package_id":      packageID,
	})

	return nil
}

// GetUserCredits retrieves all credit records for a user
func (s *Phase4Service) GetUserCredits(userID int64) ([]*domain.UserClassCredits, error) {
	return s.creditsRepo.GetByUserID(userID)
}

// GetUserCreditsByOrganization retrieves user credits for a specific organization
func (s *Phase4Service) GetUserCreditsByOrganization(userID, orgID int64) ([]*domain.UserClassCredits, error) {
	return s.creditsRepo.GetByUserAndOrganization(userID, orgID)
}

// GetAvailableCredits returns the total available credits for a user in an organization
func (s *Phase4Service) GetAvailableCredits(userID, orgID int64) (int, error) {
	return s.creditsRepo.GetAvailableCredits(userID, orgID)
}

// UseCredit deducts one credit from a user's balance
func (s *Phase4Service) UseCredit(userID, orgID int64) error {
	return s.creditsRepo.UseCredit(userID, orgID)
}

// GetExpiringCredits retrieves credits expiring within daysAhead days
func (s *Phase4Service) GetExpiringCredits(daysAhead int) ([]*domain.UserClassCredits, error) {
	return s.creditsRepo.GetExpiringSoon(daysAhead)
}

// ============================================
// WAITLIST METHODS
// ============================================

// JoinWaitlist adds a user to a session's waitlist
func (s *Phase4Service) JoinWaitlist(userID, sessionID int64) (*domain.WaitlistEntry, error) {
	// Check if session exists
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return nil, ErrClassSessionNotFound
	}

	// Check if already on waitlist
	existing, err := s.waitlistRepo.GetBySessionAndUser(sessionID, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing waitlist entry: %w", err)
	}
	if existing != nil && existing.Status == domain.WaitlistStatusWaiting {
		return nil, ErrAlreadyOnWaitlist
	}

	entry := &domain.WaitlistEntry{
		SessionID: sessionID,
		UserID:    userID,
		Status:    domain.WaitlistStatusWaiting,
	}

	if err := s.waitlistRepo.Create(entry); err != nil {
		return nil, fmt.Errorf("failed to create waitlist entry: %w", err)
	}

	s.logAudit(userID, nil, domain.EventWaitlistJoined, map[string]interface{}{
		"waitlist_entry_id": entry.ID,
		"session_id":        sessionID,
		"position":          entry.Position,
	})

	return entry, nil
}

// GetWaitlistPosition returns the user's position on the waitlist
func (s *Phase4Service) GetWaitlistPosition(sessionID, userID int64) (int, error) {
	return s.waitlistRepo.GetPosition(sessionID, userID)
}

// GetSessionWaitlist retrieves the waitlist for a session
func (s *Phase4Service) GetSessionWaitlist(sessionID int64) ([]*domain.WaitlistEntry, error) {
	return s.waitlistRepo.GetBySessionID(sessionID)
}

// GetUserWaitlistEntries retrieves all waitlist entries for a user
func (s *Phase4Service) GetUserWaitlistEntries(userID int64) ([]*domain.WaitlistEntry, error) {
	return s.waitlistRepo.GetByUserID(userID)
}

// CancelWaitlistEntry removes a user from the waitlist by entry ID
func (s *Phase4Service) CancelWaitlistEntry(userID int64, entryID int64) error {
	entry, err := s.waitlistRepo.GetByID(entryID)
	if err != nil {
		return fmt.Errorf("failed to get waitlist entry: %w", err)
	}
	if entry == nil {
		return ErrWaitlistEntryNotFound
	}

	// Only the user or admin can cancel
	if entry.UserID != userID {
		return ErrNotAuthorized
	}

	if err := s.waitlistRepo.Cancel(entryID); err != nil {
		return fmt.Errorf("failed to cancel waitlist entry: %w", err)
	}

	// Reorder positions
	if err := s.waitlistRepo.ReorderPositions(entry.SessionID); err != nil {
		return fmt.Errorf("failed to reorder waitlist positions: %w", err)
	}

	s.logAudit(userID, nil, domain.EventWaitlistCancelled, map[string]interface{}{
		"waitlist_entry_id": entryID,
		"session_id":        entry.SessionID,
	})

	return nil
}

// CancelWaitlistBySession removes a user from a session's waitlist
func (s *Phase4Service) CancelWaitlistBySession(userID, sessionID int64) error {
	entry, err := s.waitlistRepo.GetBySessionAndUser(sessionID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWaitlistEntryNotFound
		}
		return fmt.Errorf("failed to get waitlist entry: %w", err)
	}
	if entry == nil {
		return ErrWaitlistEntryNotFound
	}

	return s.CancelWaitlistEntry(userID, entry.ID)
}

// PromoteFromWaitlist promotes the next person in line and creates a reservation
func (s *Phase4Service) PromoteFromWaitlist(sessionID int64) (*domain.WaitlistEntry, error) {
	// Get next in line
	entry, err := s.waitlistRepo.GetNextInLine(sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No one on waitlist
		}
		return nil, fmt.Errorf("failed to get next in line: %w", err)
	}

	// Promote the entry
	if err := s.waitlistRepo.Promote(entry.ID); err != nil {
		return nil, fmt.Errorf("failed to promote waitlist entry: %w", err)
	}

	// Create a reservation for the user
	reservation := &domain.Reservation{
		SessionID: sessionID,
		UserID:    entry.UserID,
		Status:    domain.ReservationStatusReserved,
	}
	if err := s.reservationRepo.Create(reservation); err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	// Reorder remaining positions
	if err := s.waitlistRepo.ReorderPositions(sessionID); err != nil {
		return nil, fmt.Errorf("failed to reorder positions: %w", err)
	}

	// Create notification for promoted user
	s.createNotification(entry.UserID, &sessionID, nil, &entry.ID, domain.NotificationTypeWaitlistPromotion, time.Now())

	s.logAudit(entry.UserID, nil, domain.EventWaitlistPromoted, map[string]interface{}{
		"waitlist_entry_id": entry.ID,
		"session_id":        sessionID,
		"reservation_id":    reservation.ID,
	})

	return entry, nil
}

// ExpireOldWaitlistEntries expires waitlist entries for past sessions
func (s *Phase4Service) ExpireOldWaitlistEntries() (int, error) {
	return s.waitlistRepo.ExpireOldEntries()
}

// ============================================
// NOTIFICATION METHODS
// ============================================

// createNotification creates a new notification (internal helper)
func (s *Phase4Service) createNotification(userID int64, sessionID, reservationID, waitlistEntryID *int64, notificationType string, scheduledFor time.Time) error {
	notification := &domain.ClassNotification{
		UserID:           userID,
		SessionID:        sessionID,
		ReservationID:    reservationID,
		WaitlistEntryID:  waitlistEntryID,
		NotificationType: notificationType,
		Status:           domain.NotificationStatusPending,
		ScheduledFor:     scheduledFor,
	}
	return s.notificationRepo.Create(notification)
}

// ScheduleClassReminder schedules a reminder notification for a class
func (s *Phase4Service) ScheduleClassReminder(userID, sessionID, reservationID int64, reminderTime time.Time) error {
	return s.createNotification(userID, &sessionID, &reservationID, nil, domain.NotificationTypeClassReminder, reminderTime)
}

// GetPendingNotifications retrieves notifications ready to be sent
func (s *Phase4Service) GetPendingNotifications() ([]*domain.ClassNotification, error) {
	return s.notificationRepo.GetPendingByScheduledTime(time.Now())
}

// GetUserNotifications retrieves recent notifications for a user
func (s *Phase4Service) GetUserNotifications(userID int64, limit int) ([]*domain.ClassNotification, error) {
	return s.notificationRepo.GetByUserID(userID, limit)
}

// MarkNotificationSent marks a notification as sent
func (s *Phase4Service) MarkNotificationSent(id int64) error {
	return s.notificationRepo.MarkSent(id)
}

// MarkNotificationFailed marks a notification as failed
func (s *Phase4Service) MarkNotificationFailed(id int64, errorMsg string) error {
	return s.notificationRepo.MarkFailed(id, errorMsg)
}

// DeleteSessionNotifications deletes all notifications for a cancelled session
func (s *Phase4Service) DeleteSessionNotifications(sessionID int64) error {
	return s.notificationRepo.DeleteBySessionID(sessionID)
}

// DeleteReservationNotifications deletes notifications for a cancelled reservation
func (s *Phase4Service) DeleteReservationNotifications(reservationID int64) error {
	return s.notificationRepo.DeleteByReservationID(reservationID)
}

// ============================================
// AUDIT LOGGING HELPER
// ============================================

func (s *Phase4Service) logAudit(actorUserID int64, targetUserID *int64, eventType string, metadata map[string]interface{}) {
	if s.auditLogRepo == nil {
		return
	}

	metadataJSON, _ := json.Marshal(metadata)
	details := string(metadataJSON)

	log := &domain.AuditLog{
		UserID:       &actorUserID,
		TargetUserID: targetUserID,
		EventType:    eventType,
		Details:      &details,
		CreatedAt:    time.Now(),
	}

	_ = s.auditLogRepo.Create(log)
}
