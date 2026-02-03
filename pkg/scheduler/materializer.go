package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/logger"
)

// SessionExistenceChecker is an interface for checking session existence
// This extends the domain.ClassSessionRepository for materialization needs
type SessionExistenceChecker interface {
	ExistsByTemplateAndStartTime(templateID int64, startTime time.Time) (bool, error)
}

// MaterializerConfig holds configuration for the session materializer
type MaterializerConfig struct {
	DaysAhead int // Number of days ahead to materialize sessions
}

// DefaultMaterializerConfig returns the default configuration
func DefaultMaterializerConfig() MaterializerConfig {
	return MaterializerConfig{
		DaysAhead: 14, // Default to 2 weeks ahead
	}
}

// ExtendedSessionRepository combines the domain repository with existence checking
type ExtendedSessionRepository interface {
	domain.ClassSessionRepository
	SessionExistenceChecker
}

// Materializer handles generating class sessions from templates and schedule slots
type Materializer struct {
	templateRepo domain.ClassTemplateRepository
	slotRepo     domain.ScheduleSlotRepository
	sessionRepo  ExtendedSessionRepository
	logger       *logger.Logger
	config       MaterializerConfig
}

// NewMaterializer creates a new session materializer
func NewMaterializer(
	templateRepo domain.ClassTemplateRepository,
	slotRepo domain.ScheduleSlotRepository,
	sessionRepo ExtendedSessionRepository,
	logger *logger.Logger,
	config MaterializerConfig,
) *Materializer {
	return &Materializer{
		templateRepo: templateRepo,
		slotRepo:     slotRepo,
		sessionRepo:  sessionRepo,
		logger:       logger,
		config:       config,
	}
}

// MaterializeResult holds the results of a materialization run
type MaterializeResult struct {
	TemplatesProcessed int
	SlotsProcessed     int
	SessionsCreated    int
	SessionsSkipped    int // Already existing
	Errors             []error
}

// MaterializeAll materializes sessions for all organizations
func (m *Materializer) MaterializeAll(ctx context.Context) (*MaterializeResult, error) {
	result := &MaterializeResult{}

	// Calculate date range
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, m.config.DaysAhead)

	m.logger.Info("Starting session materialization: %s to %s (%d days ahead)",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		m.config.DaysAhead,
	)

	// We need to get all active templates
	// Since templates are per-organization, we'll need a method to get all active templates
	// For now, we'll iterate through organizations
	// This could be optimized with a single query if needed

	// Get all templates (we'll add a method for this)
	templates, err := m.getAllActiveTemplates()
	if err != nil {
		return result, fmt.Errorf("failed to get templates: %w", err)
	}

	for _, template := range templates {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		result.TemplatesProcessed++

		// Get slots for this template
		slots, err := m.slotRepo.GetByTemplateID(template.ID, false) // Only active slots
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("template %d: %w", template.ID, err))
			continue
		}

		for _, slot := range slots {
			result.SlotsProcessed++

			created, skipped, err := m.materializeSlot(ctx, template, slot, startDate, endDate)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("slot %d: %w", slot.ID, err))
				continue
			}

			result.SessionsCreated += created
			result.SessionsSkipped += skipped
		}
	}

	m.logger.Info("Materialization complete: %d templates, %d slots, %d sessions created, %d skipped",
		result.TemplatesProcessed,
		result.SlotsProcessed,
		result.SessionsCreated,
		result.SessionsSkipped,
	)

	return result, nil
}

// MaterializeForOrganization materializes sessions for a specific organization
func (m *Materializer) MaterializeForOrganization(ctx context.Context, orgID int64) (*MaterializeResult, error) {
	result := &MaterializeResult{}

	// Calculate date range
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, m.config.DaysAhead)

	m.logger.Info("Materializing sessions for org %d: %s to %s",
		orgID,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)

	templates, err := m.templateRepo.GetByOrganizationID(orgID, false) // Only active templates
	if err != nil {
		return result, fmt.Errorf("failed to get templates: %w", err)
	}

	for _, template := range templates {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		result.TemplatesProcessed++

		slots, err := m.slotRepo.GetByTemplateID(template.ID, false)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("template %d: %w", template.ID, err))
			continue
		}

		for _, slot := range slots {
			result.SlotsProcessed++

			created, skipped, err := m.materializeSlot(ctx, template, slot, startDate, endDate)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("slot %d: %w", slot.ID, err))
				continue
			}

			result.SessionsCreated += created
			result.SessionsSkipped += skipped
		}
	}

	m.logger.Info("Org %d materialization complete: %d templates, %d slots, %d sessions created, %d skipped",
		orgID,
		result.TemplatesProcessed,
		result.SlotsProcessed,
		result.SessionsCreated,
		result.SessionsSkipped,
	)

	return result, nil
}

// materializeSlot creates sessions for a single slot within the date range
func (m *Materializer) materializeSlot(
	ctx context.Context,
	template *domain.ClassTemplate,
	slot *domain.ScheduleSlot,
	startDate, endDate time.Time,
) (created int, skipped int, err error) {
	// Parse the slot's start time
	slotTime, err := time.Parse("15:04", slot.StartTime)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start time format: %w", err)
	}

	// Determine effective start date (slot's effective_start_date or startDate)
	effectiveStart := startDate
	if slot.EffectiveStartDate != nil && slot.EffectiveStartDate.After(startDate) {
		effectiveStart = *slot.EffectiveStartDate
	}

	// Determine effective end date based on recurrence_end_type
	effectiveEnd := endDate
	if slot.RecurrenceEndType == domain.RecurrenceEndDate && slot.RecurrenceEndDate != nil {
		if slot.RecurrenceEndDate.Before(endDate) {
			effectiveEnd = *slot.RecurrenceEndDate
		}
	}

	// Determine recurrence interval (default to 1 = weekly)
	recurrenceInterval := slot.RecurrenceInterval
	if recurrenceInterval <= 0 {
		recurrenceInterval = 1
	}

	// Track occurrence count for RecurrenceEndCount
	occurrenceCount := 0

	// Find the first matching day of week from effectiveStart
	date := effectiveStart
	for !date.After(effectiveEnd) {
		select {
		case <-ctx.Done():
			return created, skipped, ctx.Err()
		default:
		}

		// Check if this day matches the slot's day of week
		if int(date.Weekday()) != slot.DayOfWeek {
			date = date.AddDate(0, 0, 1)
			continue
		}

		// Build the session start time
		sessionStart := time.Date(
			date.Year(), date.Month(), date.Day(),
			slotTime.Hour(), slotTime.Minute(), 0, 0,
			time.Local,
		)

		// Skip if session start is in the past
		if sessionStart.Before(time.Now()) {
			// Move to next occurrence based on recurrence interval
			date = date.AddDate(0, 0, 7*recurrenceInterval)
			continue
		}

		// Check if we've exceeded the occurrence count limit
		if slot.RecurrenceEndType == domain.RecurrenceEndCount && slot.RecurrenceEndCount != nil {
			if occurrenceCount >= *slot.RecurrenceEndCount {
				break
			}
		}

		// Calculate end time based on template duration
		sessionEnd := sessionStart.Add(time.Duration(template.DurationMinutes) * time.Minute)

		// Check if session already exists for this template/slot/time
		exists, err := m.sessionExists(template.ID, sessionStart)
		if err != nil {
			m.logger.Warn("Error checking session existence: %v", err)
			date = date.AddDate(0, 0, 7*recurrenceInterval)
			continue
		}

		if exists {
			skipped++
			occurrenceCount++ // Count existing sessions toward limit
			date = date.AddDate(0, 0, 7*recurrenceInterval)
			continue
		}

		// Determine capacity (slot override or template default)
		capacity := template.DefaultCapacity
		if slot.OverrideCapacity != nil {
			capacity = *slot.OverrideCapacity
		}

		// Create the session
		session := &domain.ClassSession{
			OrganizationID: template.OrganizationID,
			TemplateID:     &template.ID,
			LocationID:     slot.LocationID,
			Name:           template.Name,
			Description:    template.Description,
			WorkoutID:      template.WorkoutID,
			StartTime:      sessionStart,
			EndTime:        sessionEnd,
			Capacity:       capacity,
			Status:         domain.SessionStatusScheduled,
		}

		if err := m.sessionRepo.Create(session); err != nil {
			m.logger.Error("Failed to create session for template %d at %s: %v",
				template.ID, sessionStart.Format("2006-01-02 15:04"), err)
			date = date.AddDate(0, 0, 7*recurrenceInterval)
			continue
		}

		created++
		occurrenceCount++
		m.logger.Debug("Created session: %s at %s (template %d, slot %d, interval %dw)",
			template.Name, sessionStart.Format("2006-01-02 15:04"), template.ID, slot.ID, recurrenceInterval)

		// Move to next occurrence based on recurrence interval
		date = date.AddDate(0, 0, 7*recurrenceInterval)
	}

	return created, skipped, nil
}

// sessionExists checks if a session already exists for the given template and start time
func (m *Materializer) sessionExists(templateID int64, startTime time.Time) (bool, error) {
	return m.sessionRepo.ExistsByTemplateAndStartTime(templateID, startTime)
}

// getAllActiveTemplates gets all active templates across all organizations
func (m *Materializer) getAllActiveTemplates() ([]*domain.ClassTemplate, error) {
	return m.templateRepo.GetAllActive()
}

// CreateMaterializerJob creates a scheduler job for the materializer
func (m *Materializer) CreateMaterializerJob(interval time.Duration) *Job {
	return &Job{
		Name:     "session-materializer",
		Interval: interval,
		RunFunc: func(ctx context.Context) error {
			_, err := m.MaterializeAll(ctx)
			return err
		},
	}
}
