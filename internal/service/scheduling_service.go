package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

var (
	ErrGymLocationNotFound     = errors.New("gym location not found")
	ErrClassTemplateNotFound   = errors.New("class template not found")
	ErrScheduleSlotNotFound    = errors.New("schedule slot not found")
	ErrClassSessionNotFound    = errors.New("class session not found")
	ErrReservationNotFound     = errors.New("reservation not found")
	ErrCoachAssignmentNotFound = errors.New("coach assignment not found")
	ErrSessionFull             = errors.New("session is full")
	ErrAlreadyReserved         = errors.New("user already has a reservation for this session")
	ErrCannotCancelReservation = errors.New("cannot cancel this reservation")
	ErrSessionNotActive        = errors.New("session is not in an active state")
	ErrNotAuthorized           = errors.New("not authorized to perform this action")
	ErrInvalidDayOfWeek        = errors.New("day of week must be 0-6")
	ErrInvalidTimeFormat       = errors.New("time must be in HH:MM format")
)

// MaterializeFunc is a callback function to trigger session materialization for a template
type MaterializeFunc func(templateID int64) error

type SchedulingService struct {
	locationRepo        domain.GymLocationRepository
	templateRepo        domain.ClassTemplateRepository
	templateCoachRepo   domain.TemplateCoachRepository
	slotRepo            domain.ScheduleSlotRepository
	sessionRepo         domain.ClassSessionRepository
	sessionCoachRepo    domain.SessionCoachRepository
	reservationRepo     domain.ReservationRepository
	coachAssignmentRepo domain.CoachAssignmentRepository
	orgRepo             domain.OrganizationRepository
	auditLogRepo        domain.AuditLogRepository
	workoutRepo         domain.WorkoutRepository
	materializeFunc     MaterializeFunc // Optional callback for session materialization
}

func NewSchedulingService(
	locationRepo domain.GymLocationRepository,
	templateRepo domain.ClassTemplateRepository,
	templateCoachRepo domain.TemplateCoachRepository,
	slotRepo domain.ScheduleSlotRepository,
	sessionRepo domain.ClassSessionRepository,
	sessionCoachRepo domain.SessionCoachRepository,
	reservationRepo domain.ReservationRepository,
	coachAssignmentRepo domain.CoachAssignmentRepository,
	orgRepo domain.OrganizationRepository,
	auditLogRepo domain.AuditLogRepository,
	workoutRepo domain.WorkoutRepository,
) *SchedulingService {
	return &SchedulingService{
		locationRepo:        locationRepo,
		templateRepo:        templateRepo,
		templateCoachRepo:   templateCoachRepo,
		slotRepo:            slotRepo,
		sessionRepo:         sessionRepo,
		sessionCoachRepo:    sessionCoachRepo,
		reservationRepo:     reservationRepo,
		coachAssignmentRepo: coachAssignmentRepo,
		orgRepo:             orgRepo,
		auditLogRepo:        auditLogRepo,
		workoutRepo:         workoutRepo,
	}
}

// SetMaterializeFunc sets the callback function for session materialization.
// This should be called after creating the service to enable automatic
// materialization when schedule slots are created or updated.
func (s *SchedulingService) SetMaterializeFunc(fn MaterializeFunc) {
	s.materializeFunc = fn
}

// triggerMaterialization calls the materialize callback if set
func (s *SchedulingService) triggerMaterialization(templateID int64) {
	if s.materializeFunc != nil {
		// Run in background to not block the request
		go func() {
			if err := s.materializeFunc(templateID); err != nil {
				// Log error but don't fail the request
				// The scheduler will pick up any missed sessions
				_ = err // Error is logged in the materializer
			}
		}()
	}
}

// ============================================
// GYM LOCATION METHODS
// ============================================

// CreateLocation creates a new gym location
func (s *SchedulingService) CreateLocation(adminUserID int64, location *domain.GymLocation) error {
	// Validate organization exists
	org, err := s.orgRepo.GetByID(location.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to verify organization: %w", err)
	}
	if org == nil {
		return ErrOrganizationNotFound
	}

	location.IsActive = true
	if err := s.locationRepo.Create(location); err != nil {
		return fmt.Errorf("failed to create gym location: %w", err)
	}

	// Audit log
	s.logAudit(adminUserID, nil, domain.EventGymLocationCreated, map[string]interface{}{
		"location_id":     location.ID,
		"location_name":   location.Name,
		"organization_id": location.OrganizationID,
	})

	return nil
}

// GetLocationByID retrieves a gym location by ID
func (s *SchedulingService) GetLocationByID(id int64) (*domain.GymLocation, error) {
	location, err := s.locationRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get gym location: %w", err)
	}
	if location == nil {
		return nil, ErrGymLocationNotFound
	}
	return location, nil
}

// GetLocationsByOrganization retrieves all gym locations for an organization
func (s *SchedulingService) GetLocationsByOrganization(orgID int64, includeInactive bool) ([]*domain.GymLocation, error) {
	return s.locationRepo.GetByOrganizationID(orgID, includeInactive)
}

// UpdateLocation updates a gym location
func (s *SchedulingService) UpdateLocation(adminUserID int64, location *domain.GymLocation) error {
	existing, err := s.locationRepo.GetByID(location.ID)
	if err != nil {
		return fmt.Errorf("failed to get gym location: %w", err)
	}
	if existing == nil {
		return ErrGymLocationNotFound
	}

	if err := s.locationRepo.Update(location); err != nil {
		return fmt.Errorf("failed to update gym location: %w", err)
	}

	// Audit log
	s.logAudit(adminUserID, nil, domain.EventGymLocationUpdated, map[string]interface{}{
		"location_id":   location.ID,
		"location_name": location.Name,
	})

	return nil
}

// DeleteLocation deletes a gym location
func (s *SchedulingService) DeleteLocation(adminUserID int64, id int64) error {
	location, err := s.locationRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get gym location: %w", err)
	}
	if location == nil {
		return ErrGymLocationNotFound
	}

	if err := s.locationRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete gym location: %w", err)
	}

	// Audit log
	s.logAudit(adminUserID, nil, domain.EventGymLocationDeleted, map[string]interface{}{
		"location_id":   id,
		"location_name": location.Name,
	})

	return nil
}

// ============================================
// CLASS TEMPLATE METHODS
// ============================================

// CreateTemplate creates a new class template
func (s *SchedulingService) CreateTemplate(adminUserID int64, template *domain.ClassTemplate) error {
	// Validate organization exists
	org, err := s.orgRepo.GetByID(template.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to verify organization: %w", err)
	}
	if org == nil {
		return ErrOrganizationNotFound
	}

	// Set defaults
	template.IsActive = true
	if template.DurationMinutes == 0 {
		template.DurationMinutes = 60
	}
	if template.DefaultCapacity == 0 {
		template.DefaultCapacity = 20
	}
	if template.Color == "" {
		template.Color = "#00bcd4"
	}

	if err := s.templateRepo.Create(template); err != nil {
		return fmt.Errorf("failed to create class template: %w", err)
	}

	// Audit log
	s.logAudit(adminUserID, nil, domain.EventClassTemplateCreated, map[string]interface{}{
		"template_id":     template.ID,
		"template_name":   template.Name,
		"organization_id": template.OrganizationID,
	})

	return nil
}

// GetTemplateByID retrieves a class template by ID
func (s *SchedulingService) GetTemplateByID(id int64) (*domain.ClassTemplate, error) {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get class template: %w", err)
	}
	if template == nil {
		return nil, ErrClassTemplateNotFound
	}

	// Load schedule slots
	slots, err := s.slotRepo.GetByTemplateID(id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule slots: %w", err)
	}
	template.ScheduleSlots = slots

	return template, nil
}

// GetTemplatesByOrganization retrieves all class templates for an organization
// Includes schedule slots and default coaches for each template
func (s *SchedulingService) GetTemplatesByOrganization(orgID int64, includeInactive bool) ([]*domain.ClassTemplate, error) {
	templates, err := s.templateRepo.GetByOrganizationID(orgID, includeInactive)
	if err != nil {
		return nil, err
	}

	// Fetch schedule slots and coaches for each template
	for _, template := range templates {
		// Get schedule slots
		slots, err := s.slotRepo.GetByTemplateID(template.ID, false) // Only active slots
		if err == nil {
			template.ScheduleSlots = slots
		}

		// Get default coaches
		coaches, err := s.templateCoachRepo.GetByTemplateID(template.ID)
		if err == nil {
			template.DefaultCoaches = coaches
		}
	}

	return templates, nil
}

// UpdateTemplate updates a class template
func (s *SchedulingService) UpdateTemplate(adminUserID int64, template *domain.ClassTemplate) error {
	existing, err := s.templateRepo.GetByID(template.ID)
	if err != nil {
		return fmt.Errorf("failed to get class template: %w", err)
	}
	if existing == nil {
		return ErrClassTemplateNotFound
	}

	if err := s.templateRepo.Update(template); err != nil {
		return fmt.Errorf("failed to update class template: %w", err)
	}

	// Audit log
	s.logAudit(adminUserID, nil, domain.EventClassTemplateUpdated, map[string]interface{}{
		"template_id":   template.ID,
		"template_name": template.Name,
	})

	return nil
}

// DeleteTemplate deletes a class template
func (s *SchedulingService) DeleteTemplate(adminUserID int64, id int64) error {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get class template: %w", err)
	}
	if template == nil {
		return ErrClassTemplateNotFound
	}

	if err := s.templateRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete class template: %w", err)
	}

	// Audit log
	s.logAudit(adminUserID, nil, domain.EventClassTemplateDeleted, map[string]interface{}{
		"template_id":   id,
		"template_name": template.Name,
	})

	return nil
}

// ============================================
// SCHEDULE SLOT METHODS
// ============================================

// CreateScheduleSlot creates a new schedule slot
func (s *SchedulingService) CreateScheduleSlot(adminUserID int64, slot *domain.ScheduleSlot) error {
	// Validate template exists
	template, err := s.templateRepo.GetByID(slot.TemplateID)
	if err != nil {
		return fmt.Errorf("failed to verify template: %w", err)
	}
	if template == nil {
		return ErrClassTemplateNotFound
	}

	// Validate day of week
	if slot.DayOfWeek < 0 || slot.DayOfWeek > 6 {
		return ErrInvalidDayOfWeek
	}

	// Validate time format (basic check)
	if len(slot.StartTime) < 5 {
		return ErrInvalidTimeFormat
	}

	slot.IsActive = true
	if err := s.slotRepo.Create(slot); err != nil {
		return fmt.Errorf("failed to create schedule slot: %w", err)
	}

	// Audit log
	s.logAudit(adminUserID, nil, domain.EventScheduleSlotCreated, map[string]interface{}{
		"slot_id":     slot.ID,
		"template_id": slot.TemplateID,
		"day_of_week": slot.DayOfWeek,
		"start_time":  slot.StartTime,
	})

	// Trigger session materialization for this template
	s.triggerMaterialization(slot.TemplateID)

	return nil
}

// GetScheduleSlotsByTemplate retrieves all schedule slots for a template
func (s *SchedulingService) GetScheduleSlotsByTemplate(templateID int64, includeInactive bool) ([]*domain.ScheduleSlot, error) {
	return s.slotRepo.GetByTemplateID(templateID, includeInactive)
}

// UpdateScheduleSlot updates a schedule slot
func (s *SchedulingService) UpdateScheduleSlot(adminUserID int64, slot *domain.ScheduleSlot) error {
	existing, err := s.slotRepo.GetByID(slot.ID)
	if err != nil {
		return fmt.Errorf("failed to get schedule slot: %w", err)
	}
	if existing == nil {
		return ErrScheduleSlotNotFound
	}

	// Validate day of week
	if slot.DayOfWeek < 0 || slot.DayOfWeek > 6 {
		return ErrInvalidDayOfWeek
	}

	if err := s.slotRepo.Update(slot); err != nil {
		return fmt.Errorf("failed to update schedule slot: %w", err)
	}

	// Audit log
	s.logAudit(adminUserID, nil, domain.EventScheduleSlotUpdated, map[string]interface{}{
		"slot_id":     slot.ID,
		"day_of_week": slot.DayOfWeek,
		"start_time":  slot.StartTime,
	})

	// Trigger session materialization for this template
	s.triggerMaterialization(existing.TemplateID)

	return nil
}

// DeleteScheduleSlot deletes a schedule slot
func (s *SchedulingService) DeleteScheduleSlot(adminUserID int64, id int64) error {
	slot, err := s.slotRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get schedule slot: %w", err)
	}
	if slot == nil {
		return ErrScheduleSlotNotFound
	}

	if err := s.slotRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete schedule slot: %w", err)
	}

	// Audit log
	s.logAudit(adminUserID, nil, domain.EventScheduleSlotDeleted, map[string]interface{}{
		"slot_id":     id,
		"template_id": slot.TemplateID,
	})

	return nil
}

// ============================================
// CLASS SESSION METHODS
// ============================================

// CreateSession creates a new class session
func (s *SchedulingService) CreateSession(adminUserID int64, session *domain.ClassSession) error {
	// Validate organization exists
	org, err := s.orgRepo.GetByID(session.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to verify organization: %w", err)
	}
	if org == nil {
		return ErrOrganizationNotFound
	}

	// Set defaults
	session.Status = domain.SessionStatusScheduled
	if session.Capacity == 0 {
		session.Capacity = 20
	}

	// If template is provided, use its defaults
	if session.TemplateID != nil {
		template, err := s.templateRepo.GetByID(*session.TemplateID)
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}
		if template != nil {
			if session.Name == "" {
				session.Name = template.Name
			}
			if session.Description == nil && template.Description != nil {
				session.Description = template.Description
			}
			if session.WorkoutID == nil && template.WorkoutID != nil {
				session.WorkoutID = template.WorkoutID
			}
			if session.Capacity == 20 { // If still default, use template's
				session.Capacity = template.DefaultCapacity
			}
		}
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return fmt.Errorf("failed to create class session: %w", err)
	}

	// Audit log
	s.logAudit(adminUserID, nil, domain.EventClassSessionCreated, map[string]interface{}{
		"session_id":      session.ID,
		"session_name":    session.Name,
		"organization_id": session.OrganizationID,
		"start_time":      session.StartTime.Format(time.RFC3339),
	})

	return nil
}

// CreateSessionFromTemplate creates a session from a template for a specific date/time
func (s *SchedulingService) CreateSessionFromTemplate(adminUserID int64, templateID int64, startTime time.Time, locationID *int64) (*domain.ClassSession, error) {
	template, err := s.templateRepo.GetByID(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	if template == nil {
		return nil, ErrClassTemplateNotFound
	}

	session := &domain.ClassSession{
		OrganizationID: template.OrganizationID,
		TemplateID:     &templateID,
		LocationID:     locationID,
		Name:           template.Name,
		Description:    template.Description,
		WorkoutID:      template.WorkoutID,
		StartTime:      startTime,
		EndTime:        startTime.Add(time.Duration(template.DurationMinutes) * time.Minute),
		Capacity:       template.DefaultCapacity,
		Status:         domain.SessionStatusScheduled,
	}

	if err := s.CreateSession(adminUserID, session); err != nil {
		return nil, err
	}

	return session, nil
}

// GetSessionByID retrieves a class session by ID
func (s *SchedulingService) GetSessionByID(id int64) (*domain.ClassSession, error) {
	session, err := s.sessionRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get class session: %w", err)
	}
	if session == nil {
		return nil, ErrClassSessionNotFound
	}

	// Load coaches
	coaches, err := s.sessionCoachRepo.GetBySessionID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session coaches: %w", err)
	}
	session.Coaches = coaches

	return session, nil
}

// GetSessionByIDWithWorkout retrieves a class session by ID with full workout details
func (s *SchedulingService) GetSessionByIDWithWorkout(id int64) (*domain.ClassSession, error) {
	session, err := s.sessionRepo.GetByIDWithWorkoutDetails(id, s.workoutRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to get class session with workout: %w", err)
	}
	if session == nil {
		return nil, ErrClassSessionNotFound
	}

	// Load coaches
	coaches, err := s.sessionCoachRepo.GetBySessionID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session coaches: %w", err)
	}
	session.Coaches = coaches

	return session, nil
}

// GetSessionsByOrganization retrieves all sessions for an organization within a date range
func (s *SchedulingService) GetSessionsByOrganization(orgID int64, startDate, endDate time.Time) ([]*domain.ClassSession, error) {
	return s.sessionRepo.GetByOrganizationID(orgID, startDate, endDate)
}

// GetSessionsByOrganizationForUser retrieves sessions with user's reservation status
func (s *SchedulingService) GetSessionsByOrganizationForUser(orgID, userID int64, startDate, endDate time.Time) ([]*domain.ClassSession, error) {
	return s.sessionRepo.GetByOrganizationIDWithUserStatus(orgID, userID, startDate, endDate)
}

// UpdateSession updates a class session
func (s *SchedulingService) UpdateSession(adminUserID int64, session *domain.ClassSession) error {
	existing, err := s.sessionRepo.GetByID(session.ID)
	if err != nil {
		return fmt.Errorf("failed to get class session: %w", err)
	}
	if existing == nil {
		return ErrClassSessionNotFound
	}

	// Preserve fields that shouldn't be changed via update
	session.Status = existing.Status
	session.OrganizationID = existing.OrganizationID
	session.TemplateID = existing.TemplateID
	session.CreatedAt = existing.CreatedAt

	if err := s.sessionRepo.Update(session); err != nil {
		return fmt.Errorf("failed to update class session: %w", err)
	}

	// Audit log
	s.logAudit(adminUserID, nil, domain.EventClassSessionUpdated, map[string]interface{}{
		"session_id":   session.ID,
		"session_name": session.Name,
	})

	return nil
}

// CancelSession cancels a session with reason
func (s *SchedulingService) CancelSession(adminUserID int64, sessionID int64, reason string) error {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get class session: %w", err)
	}
	if session == nil {
		return ErrClassSessionNotFound
	}

	if session.Status != domain.SessionStatusScheduled && session.Status != domain.SessionStatusInProgress {
		return ErrSessionNotActive
	}

	if err := s.sessionRepo.Cancel(sessionID, reason); err != nil {
		return fmt.Errorf("failed to cancel class session: %w", err)
	}

	// Audit log
	s.logAudit(adminUserID, nil, domain.EventClassSessionCancelled, map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Name,
		"reason":       reason,
	})

	return nil
}

// CompleteSession marks a session as completed
func (s *SchedulingService) CompleteSession(adminUserID int64, sessionID int64) error {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get class session: %w", err)
	}
	if session == nil {
		return ErrClassSessionNotFound
	}

	if err := s.sessionRepo.Complete(sessionID); err != nil {
		return fmt.Errorf("failed to complete class session: %w", err)
	}

	// Audit log
	s.logAudit(adminUserID, nil, domain.EventClassSessionCompleted, map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Name,
	})

	return nil
}

// ============================================
// COACH ASSIGNMENT METHODS
// ============================================

// AssignCoach assigns a user as coach for an organization
func (s *SchedulingService) AssignCoach(adminUserID int64, orgID, userID int64) error {
	// Check if already assigned
	isCoach, err := s.coachAssignmentRepo.IsCoachForOrganization(userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to check coach status: %w", err)
	}
	if isCoach {
		return fmt.Errorf("user is already a coach for this organization")
	}

	assignment := &domain.CoachAssignment{
		OrganizationID: orgID,
		UserID:         userID,
		IsActive:       true,
	}

	if err := s.coachAssignmentRepo.Create(assignment); err != nil {
		return fmt.Errorf("failed to create coach assignment: %w", err)
	}

	// Audit log
	targetUserID := userID
	s.logAudit(adminUserID, &targetUserID, domain.EventCoachAssigned, map[string]interface{}{
		"organization_id": orgID,
		"user_id":         userID,
	})

	return nil
}

// UnassignCoach removes a coach assignment
func (s *SchedulingService) UnassignCoach(adminUserID int64, assignmentID int64) error {
	assignment, err := s.coachAssignmentRepo.GetByID(assignmentID)
	if err != nil {
		return fmt.Errorf("failed to get coach assignment: %w", err)
	}
	if assignment == nil {
		return ErrCoachAssignmentNotFound
	}

	if err := s.coachAssignmentRepo.Delete(assignmentID); err != nil {
		return fmt.Errorf("failed to delete coach assignment: %w", err)
	}

	// Audit log
	targetUserID := assignment.UserID
	s.logAudit(adminUserID, &targetUserID, domain.EventCoachUnassigned, map[string]interface{}{
		"organization_id": assignment.OrganizationID,
		"user_id":         assignment.UserID,
	})

	return nil
}

// GetCoachesByOrganization retrieves all coaches for an organization
func (s *SchedulingService) GetCoachesByOrganization(orgID int64, includeInactive bool) ([]*domain.CoachAssignment, error) {
	return s.coachAssignmentRepo.GetByOrganizationID(orgID, includeInactive)
}

// IsUserCoachForOrganization checks if a user is a coach for an organization
func (s *SchedulingService) IsUserCoachForOrganization(userID, orgID int64) (bool, error) {
	return s.coachAssignmentRepo.IsCoachForOrganization(userID, orgID)
}

// IsUserCoach checks if a user is a coach for any organization
func (s *SchedulingService) IsUserCoach(userID int64) (bool, error) {
	assignments, err := s.coachAssignmentRepo.GetByUserID(userID)
	if err != nil {
		return false, err
	}
	return len(assignments) > 0, nil
}

// AddCoachToSession adds a coach to a specific session or updates their lead status if already assigned
func (s *SchedulingService) AddCoachToSession(adminUserID int64, sessionID, coachUserID int64, isLead bool) error {
	// Verify session exists
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return ErrClassSessionNotFound
	}

	// Verify user is a coach for this organization
	isCoach, err := s.coachAssignmentRepo.IsCoachForOrganization(coachUserID, session.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to verify coach: %w", err)
	}
	if !isCoach {
		return fmt.Errorf("user is not a coach for this organization")
	}

	// Check if coach is already assigned to this session
	existingCoaches, err := s.sessionCoachRepo.GetBySessionID(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get existing coaches: %w", err)
	}

	for _, coach := range existingCoaches {
		if coach.UserID == coachUserID {
			// Coach already assigned - just update lead status if requested
			if isLead {
				if err := s.sessionCoachRepo.SetLeadCoach(sessionID, coachUserID); err != nil {
					return fmt.Errorf("failed to set lead coach: %w", err)
				}
			}
			return nil
		}
	}

	// Coach not yet assigned - create new assignment
	sessionCoach := &domain.SessionCoach{
		SessionID: sessionID,
		UserID:    coachUserID,
		IsLead:    isLead,
	}

	if err := s.sessionCoachRepo.Create(sessionCoach); err != nil {
		return fmt.Errorf("failed to add coach to session: %w", err)
	}

	return nil
}

// RemoveCoachFromSession removes a coach from a session
func (s *SchedulingService) RemoveCoachFromSession(adminUserID int64, sessionID, coachUserID int64) error {
	if err := s.sessionCoachRepo.DeleteBySessionAndUser(sessionID, coachUserID); err != nil {
		return fmt.Errorf("failed to remove coach from session: %w", err)
	}
	return nil
}

// GetSessionCoaches retrieves all coaches assigned to a session
func (s *SchedulingService) GetSessionCoaches(sessionID int64) ([]*domain.SessionCoach, error) {
	return s.sessionCoachRepo.GetBySessionID(sessionID)
}

// ============================================
// RESERVATION METHODS
// ============================================

// CreateReservation creates a new reservation for a user
func (s *SchedulingService) CreateReservation(userID int64, sessionID int64) (*domain.Reservation, error) {
	// Get session
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return nil, ErrClassSessionNotFound
	}

	// Check if session is active
	if session.Status != domain.SessionStatusScheduled {
		return nil, ErrSessionNotActive
	}

	// Check if user already has a reservation
	existing, err := s.reservationRepo.GetBySessionAndUser(sessionID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing reservation: %w", err)
	}
	if existing != nil && existing.Status != domain.ReservationStatusCancelled {
		return nil, ErrAlreadyReserved
	}

	// Check capacity
	count, err := s.reservationRepo.CountActiveForSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to count reservations: %w", err)
	}
	if count >= session.Capacity {
		return nil, ErrSessionFull
	}

	reservation := &domain.Reservation{
		SessionID: sessionID,
		UserID:    userID,
		Status:    domain.ReservationStatusReserved,
	}

	if err := s.reservationRepo.Create(reservation); err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	// Audit log
	s.logAudit(userID, nil, domain.EventReservationCreated, map[string]interface{}{
		"reservation_id": reservation.ID,
		"session_id":     sessionID,
		"session_name":   session.Name,
	})

	return reservation, nil
}

// CancelReservation cancels a user's reservation
func (s *SchedulingService) CancelReservation(userID int64, sessionID int64, reason string) error {
	reservation, err := s.reservationRepo.GetBySessionAndUser(sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}
	if reservation == nil {
		return ErrReservationNotFound
	}

	// Can only cancel if reserved or checked_in
	if reservation.Status != domain.ReservationStatusReserved && reservation.Status != domain.ReservationStatusCheckedIn {
		return ErrCannotCancelReservation
	}

	if err := s.reservationRepo.Cancel(reservation.ID, reason); err != nil {
		return fmt.Errorf("failed to cancel reservation: %w", err)
	}

	// Audit log
	s.logAudit(userID, nil, domain.EventReservationCancelled, map[string]interface{}{
		"reservation_id": reservation.ID,
		"session_id":     sessionID,
		"reason":         reason,
	})

	return nil
}

// CheckInReservation checks in a reservation (coach/admin action)
func (s *SchedulingService) CheckInReservation(coachUserID int64, reservationID int64) error {
	reservation, err := s.reservationRepo.GetByID(reservationID)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}
	if reservation == nil {
		return ErrReservationNotFound
	}

	if reservation.Status != domain.ReservationStatusReserved {
		return fmt.Errorf("can only check in reservations with 'reserved' status")
	}

	if err := s.reservationRepo.CheckIn(reservationID, coachUserID); err != nil {
		return fmt.Errorf("failed to check in reservation: %w", err)
	}

	// Audit log
	targetUserID := reservation.UserID
	s.logAudit(coachUserID, &targetUserID, domain.EventReservationCheckedIn, map[string]interface{}{
		"reservation_id": reservationID,
		"session_id":     reservation.SessionID,
	})

	return nil
}

// MarkNoShow marks a reservation as no-show (coach/admin action)
func (s *SchedulingService) MarkNoShow(coachUserID int64, reservationID int64) error {
	reservation, err := s.reservationRepo.GetByID(reservationID)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}
	if reservation == nil {
		return ErrReservationNotFound
	}

	if err := s.reservationRepo.MarkNoShow(reservationID); err != nil {
		return fmt.Errorf("failed to mark reservation as no-show: %w", err)
	}

	// Audit log
	targetUserID := reservation.UserID
	s.logAudit(coachUserID, &targetUserID, domain.EventReservationNoShow, map[string]interface{}{
		"reservation_id": reservationID,
		"session_id":     reservation.SessionID,
	})

	return nil
}

// GetReservationsBySession retrieves all reservations for a session
func (s *SchedulingService) GetReservationsBySession(sessionID int64) ([]*domain.Reservation, error) {
	return s.reservationRepo.GetBySessionID(sessionID)
}

// GetReservationRoster retrieves active reservations for a session (for coach view)
func (s *SchedulingService) GetReservationRoster(sessionID int64) ([]*domain.Reservation, error) {
	return s.reservationRepo.GetBySessionIDWithStatus(sessionID, []string{
		domain.ReservationStatusReserved,
		domain.ReservationStatusCheckedIn,
		domain.ReservationStatusAttended,
	})
}

// GetUpcomingUserReservations retrieves a user's upcoming reservations
func (s *SchedulingService) GetUpcomingUserReservations(userID int64) ([]*domain.ReservationWithSession, error) {
	return s.reservationRepo.GetUpcomingByUserID(userID)
}

// GetUserReservationHistory retrieves a user's reservation history
func (s *SchedulingService) GetUserReservationHistory(userID int64, limit, offset int) ([]*domain.ReservationWithSession, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.reservationRepo.GetByUserID(userID, limit, offset)
}

// ============================================
// COACH PORTAL METHODS
// ============================================

// GetCoachUpcomingSessions retrieves upcoming sessions where the user is assigned as a coach
func (s *SchedulingService) GetCoachUpcomingSessions(coachUserID int64, limit int) ([]*domain.ClassSession, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.sessionRepo.GetUpcomingByCoachID(coachUserID, limit)
}

// ============================================
// SCHEDULE PREVIEW METHODS
// ============================================

// PreviewScheduleDates generates a list of dates when sessions would occur for a template
func (s *SchedulingService) PreviewScheduleDates(templateID int64, months int) ([]time.Time, error) {
	// Get template
	template, err := s.templateRepo.GetByID(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	if template == nil {
		return nil, ErrClassTemplateNotFound
	}

	// Get slots for this template
	slots, err := s.slotRepo.GetByTemplateID(templateID, false) // Only active slots
	if err != nil {
		return nil, fmt.Errorf("failed to get slots: %w", err)
	}

	// Calculate date range
	startDate := time.Now()
	endDate := startDate.AddDate(0, months, 0)

	var dates []time.Time

	for _, slot := range slots {
		// Parse slot time - handle both "15:04" and "15:04:05" formats
		var slotTime time.Time
		var err error
		slotTime, err = time.Parse("15:04", slot.StartTime)
		if err != nil {
			// Try with seconds (MariaDB/MySQL format)
			slotTime, err = time.Parse("15:04:05", slot.StartTime)
			if err != nil {
				continue
			}
		}

		// Determine effective start date
		effectiveStart := startDate
		if slot.EffectiveStartDate != nil && slot.EffectiveStartDate.After(startDate) {
			effectiveStart = *slot.EffectiveStartDate
		}

		// Determine effective end date
		effectiveEnd := endDate
		if slot.RecurrenceEndType == domain.RecurrenceEndDate && slot.RecurrenceEndDate != nil {
			if slot.RecurrenceEndDate.Before(endDate) {
				effectiveEnd = *slot.RecurrenceEndDate
			}
		}

		// Determine recurrence interval
		recurrenceInterval := slot.RecurrenceInterval
		if recurrenceInterval <= 0 {
			recurrenceInterval = 1
		}

		// Get list of days to generate sessions for
		// Use DaysOfWeek array if available, otherwise fall back to single DayOfWeek
		daysOfWeek := slot.DaysOfWeek
		if len(daysOfWeek) == 0 {
			daysOfWeek = []int{slot.DayOfWeek}
		}

		// Track occurrence count (total across all days)
		occurrenceCount := 0

		// Find the start of the week containing effectiveStart (Sunday)
		weekStart := effectiveStart
		for weekStart.Weekday() != time.Sunday {
			weekStart = weekStart.AddDate(0, 0, -1)
		}

		// Iterate through weeks
		for !weekStart.After(effectiveEnd) {
			// For each day of week in the slot
			for _, dayOfWeek := range daysOfWeek {
				// Calculate the date for this day of week in the current week
				date := weekStart.AddDate(0, 0, dayOfWeek)

				// Skip if before effective start or after effective end
				if date.Before(effectiveStart) || date.After(effectiveEnd) {
					continue
				}

				// Build the session start time
				sessionStart := time.Date(
					date.Year(), date.Month(), date.Day(),
					slotTime.Hour(), slotTime.Minute(), 0, 0,
					time.Local,
				)

				// Check occurrence count limit
				if slot.RecurrenceEndType == domain.RecurrenceEndCount && slot.RecurrenceEndCount != nil {
					if occurrenceCount >= *slot.RecurrenceEndCount {
						break
					}
				}

				dates = append(dates, sessionStart)
				occurrenceCount++
			}

			// Check if we've hit the occurrence limit
			if slot.RecurrenceEndType == domain.RecurrenceEndCount && slot.RecurrenceEndCount != nil {
				if occurrenceCount >= *slot.RecurrenceEndCount {
					break
				}
			}

			// Move to next week (respecting recurrence interval)
			weekStart = weekStart.AddDate(0, 0, 7*recurrenceInterval)
		}
	}

	// Sort dates chronologically
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	return dates, nil
}

// ============================================
// HELPER METHODS
// ============================================

func (s *SchedulingService) logAudit(userID int64, targetUserID *int64, eventType string, details map[string]interface{}) {
	detailsJSON, _ := json.Marshal(details)
	detailsStr := string(detailsJSON)

	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:       &userID,
		TargetUserID: targetUserID,
		EventType:    eventType,
		Details:      &detailsStr,
		CreatedAt:    time.Now(),
	})
}

// ============================================
// TEMPLATE COACH METHODS
// ============================================

// GetTemplateCoaches retrieves all coaches for a template
func (s *SchedulingService) GetTemplateCoaches(templateID int64) ([]*domain.TemplateCoach, error) {
	return s.templateCoachRepo.GetByTemplateID(templateID)
}

// AddTemplateCoach adds a coach to a template
func (s *SchedulingService) AddTemplateCoach(adminUserID, templateID, coachUserID int64, isLead bool) error {
	if err := s.templateCoachRepo.Add(templateID, coachUserID, isLead); err != nil {
		return fmt.Errorf("failed to add template coach: %w", err)
	}

	s.logAudit(adminUserID, nil, domain.EventTemplateCoachAdded, map[string]interface{}{
		"template_id":   templateID,
		"coach_user_id": coachUserID,
		"is_lead":       isLead,
	})

	return nil
}

// RemoveTemplateCoach removes a coach from a template
func (s *SchedulingService) RemoveTemplateCoach(adminUserID, templateID, coachUserID int64) error {
	if err := s.templateCoachRepo.Remove(templateID, coachUserID); err != nil {
		return fmt.Errorf("failed to remove template coach: %w", err)
	}

	s.logAudit(adminUserID, nil, domain.EventTemplateCoachRemoved, map[string]interface{}{
		"template_id":   templateID,
		"coach_user_id": coachUserID,
	})

	return nil
}
