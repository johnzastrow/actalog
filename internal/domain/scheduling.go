// Package domain contains the core business entities for scheduling
package domain

import (
	"time"
)

// Session status constants
const (
	SessionStatusScheduled  = "scheduled"
	SessionStatusInProgress = "in_progress"
	SessionStatusCompleted  = "completed"
	SessionStatusCancelled  = "cancelled"
)

// Reservation status constants
const (
	ReservationStatusReserved  = "reserved"
	ReservationStatusCheckedIn = "checked_in"
	ReservationStatusCancelled = "cancelled"
	ReservationStatusNoShow    = "no_show"
	ReservationStatusAttended  = "attended"
)

// GymLocation represents a physical location within an organization
type GymLocation struct {
	ID             int64     `json:"id" db:"id"`
	OrganizationID int64     `json:"organization_id" db:"organization_id"`
	Name           string    `json:"name" db:"name"`
	Description    *string   `json:"description,omitempty" db:"description"`
	Address        *string   `json:"address,omitempty" db:"address"`
	Capacity       int       `json:"capacity" db:"capacity"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// ClassTemplate represents a reusable class definition
type ClassTemplate struct {
	ID              int64     `json:"id" db:"id"`
	OrganizationID  int64     `json:"organization_id" db:"organization_id"`
	Name            string    `json:"name" db:"name"`
	Description     *string   `json:"description,omitempty" db:"description"`
	WorkoutID       *int64    `json:"workout_id,omitempty" db:"workout_id"`
	DurationMinutes int       `json:"duration_minutes" db:"duration_minutes"`
	DefaultCapacity int       `json:"default_capacity" db:"default_capacity"`
	Color           string    `json:"color" db:"color"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`

	// Populated via JOINs, not stored in class_templates table
	ScheduleSlots []*ScheduleSlot `json:"schedule_slots,omitempty"`
	WorkoutName   *string         `json:"workout_name,omitempty"`
}

// ScheduleSlot represents a recurring time pattern for a class template
type ScheduleSlot struct {
	ID               int64     `json:"id" db:"id"`
	TemplateID       int64     `json:"template_id" db:"template_id"`
	LocationID       *int64    `json:"location_id,omitempty" db:"location_id"`
	DayOfWeek        int       `json:"day_of_week" db:"day_of_week"` // 0=Sunday, 6=Saturday
	StartTime        string    `json:"start_time" db:"start_time"`   // HH:MM format
	OverrideCapacity *int      `json:"override_capacity,omitempty" db:"override_capacity"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`

	// Populated via JOINs
	LocationName *string `json:"location_name,omitempty"`
}

// CoachAssignment represents a per-gym coach role assignment
type CoachAssignment struct {
	ID             int64     `json:"id" db:"id"`
	OrganizationID int64     `json:"organization_id" db:"organization_id"`
	UserID         int64     `json:"user_id" db:"user_id"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	AssignedAt     time.Time `json:"assigned_at" db:"assigned_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`

	// Populated via JOINs
	UserName  *string `json:"user_name,omitempty"`
	UserEmail *string `json:"user_email,omitempty"`
}

// ClassSession represents an actual scheduled class instance
type ClassSession struct {
	ID              int64      `json:"id" db:"id"`
	OrganizationID  int64      `json:"organization_id" db:"organization_id"`
	TemplateID      *int64     `json:"template_id,omitempty" db:"template_id"`
	LocationID      *int64     `json:"location_id,omitempty" db:"location_id"`
	Name            string     `json:"name" db:"name"`
	Description     *string    `json:"description,omitempty" db:"description"`
	WorkoutID       *int64     `json:"workout_id,omitempty" db:"workout_id"`
	StartTime       time.Time  `json:"start_time" db:"start_time"`
	EndTime         time.Time  `json:"end_time" db:"end_time"`
	Capacity        int        `json:"capacity" db:"capacity"`
	Status          string     `json:"status" db:"status"` // scheduled, in_progress, completed, cancelled
	CancelledAt     *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CancelledReason *string    `json:"cancelled_reason,omitempty" db:"cancelled_reason"`
	CompletedAt     *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`

	// Populated via JOINs or computed
	TemplateName       *string        `json:"template_name,omitempty"`
	LocationName       *string        `json:"location_name,omitempty"`
	WorkoutName        *string        `json:"workout_name,omitempty"`
	ReservationCount   int            `json:"reservation_count"`
	AvailableSpots     int            `json:"available_spots"`
	Coaches            []*SessionCoach `json:"coaches,omitempty"`
	CurrentUserStatus  *string        `json:"current_user_status,omitempty"` // For athlete view: reserved/checked_in/etc
}

// SessionCoach represents a coach assigned to a specific session
type SessionCoach struct {
	ID        int64     `json:"id" db:"id"`
	SessionID int64     `json:"session_id" db:"session_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	IsLead    bool      `json:"is_lead" db:"is_lead"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Populated via JOINs
	UserName  *string `json:"user_name,omitempty"`
	UserEmail *string `json:"user_email,omitempty"`
}

// Reservation represents a user's booking for a class session
type Reservation struct {
	ID                int64      `json:"id" db:"id"`
	SessionID         int64      `json:"session_id" db:"session_id"`
	UserID            int64      `json:"user_id" db:"user_id"`
	Status            string     `json:"status" db:"status"` // reserved, checked_in, cancelled, no_show, attended
	ReservedAt        time.Time  `json:"reserved_at" db:"reserved_at"`
	CheckedInAt       *time.Time `json:"checked_in_at,omitempty" db:"checked_in_at"`
	CheckedInByUserID *int64     `json:"checked_in_by_user_id,omitempty" db:"checked_in_by_user_id"`
	CancelledAt       *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CancelledReason   *string    `json:"cancelled_reason,omitempty" db:"cancelled_reason"`
	NoShowMarkedAt    *time.Time `json:"no_show_marked_at,omitempty" db:"no_show_marked_at"`
	UserWorkoutID     *int64     `json:"user_workout_id,omitempty" db:"user_workout_id"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`

	// Populated via JOINs
	UserName         *string `json:"user_name,omitempty"`
	UserEmail        *string `json:"user_email,omitempty"`
	CheckedInByName  *string `json:"checked_in_by_name,omitempty"`
	CheckedInByEmail *string `json:"checked_in_by_email,omitempty"`
}

// ReservationWithSession extends Reservation with session details for user's view
type ReservationWithSession struct {
	Reservation
	SessionName      string    `json:"session_name"`
	SessionStartTime time.Time `json:"session_start_time"`
	SessionEndTime   time.Time `json:"session_end_time"`
	SessionStatus    string    `json:"session_status"`
	LocationName     *string   `json:"location_name,omitempty"`
	OrganizationID   int64     `json:"organization_id"`
	OrganizationName string    `json:"organization_name"`
}

// GymLocationRepository defines the interface for gym location data access
type GymLocationRepository interface {
	Create(location *GymLocation) error
	GetByID(id int64) (*GymLocation, error)
	GetByOrganizationID(orgID int64, includeInactive bool) ([]*GymLocation, error)
	Update(location *GymLocation) error
	Delete(id int64) error
}

// ClassTemplateRepository defines the interface for class template data access
type ClassTemplateRepository interface {
	Create(template *ClassTemplate) error
	GetByID(id int64) (*ClassTemplate, error)
	GetByOrganizationID(orgID int64, includeInactive bool) ([]*ClassTemplate, error)
	GetAllActive() ([]*ClassTemplate, error)
	Update(template *ClassTemplate) error
	Delete(id int64) error
}

// ScheduleSlotRepository defines the interface for schedule slot data access
type ScheduleSlotRepository interface {
	Create(slot *ScheduleSlot) error
	GetByID(id int64) (*ScheduleSlot, error)
	GetByTemplateID(templateID int64, includeInactive bool) ([]*ScheduleSlot, error)
	Update(slot *ScheduleSlot) error
	Delete(id int64) error
}

// CoachAssignmentRepository defines the interface for coach assignment data access
type CoachAssignmentRepository interface {
	Create(assignment *CoachAssignment) error
	GetByID(id int64) (*CoachAssignment, error)
	GetByOrganizationID(orgID int64, includeInactive bool) ([]*CoachAssignment, error)
	GetByUserID(userID int64) ([]*CoachAssignment, error)
	IsCoachForOrganization(userID, orgID int64) (bool, error)
	Update(assignment *CoachAssignment) error
	Delete(id int64) error
}

// ClassSessionRepository defines the interface for class session data access
type ClassSessionRepository interface {
	Create(session *ClassSession) error
	GetByID(id int64) (*ClassSession, error)
	GetByOrganizationID(orgID int64, startDate, endDate time.Time) ([]*ClassSession, error)
	GetByOrganizationIDWithUserStatus(orgID, userID int64, startDate, endDate time.Time) ([]*ClassSession, error)
	GetUpcomingByCoachID(coachUserID int64, limit int) ([]*ClassSession, error)
	Update(session *ClassSession) error
	UpdateStatus(id int64, status string) error
	Cancel(id int64, reason string) error
	Complete(id int64) error
	GetReservationCount(sessionID int64) (int, error)
	ExistsByTemplateAndStartTime(templateID int64, startTime time.Time) (bool, error)
}

// SessionCoachRepository defines the interface for session coach data access
type SessionCoachRepository interface {
	Create(coach *SessionCoach) error
	GetBySessionID(sessionID int64) ([]*SessionCoach, error)
	Delete(id int64) error
	DeleteBySessionAndUser(sessionID, userID int64) error
}

// ReservationRepository defines the interface for reservation data access
type ReservationRepository interface {
	Create(reservation *Reservation) error
	GetByID(id int64) (*Reservation, error)
	GetBySessionID(sessionID int64) ([]*Reservation, error)
	GetBySessionIDWithStatus(sessionID int64, statuses []string) ([]*Reservation, error)
	GetByUserID(userID int64, limit, offset int) ([]*ReservationWithSession, error)
	GetUpcomingByUserID(userID int64) ([]*ReservationWithSession, error)
	GetBySessionAndUser(sessionID, userID int64) (*Reservation, error)
	Update(reservation *Reservation) error
	UpdateStatus(id int64, status string) error
	CheckIn(id int64, checkedInByUserID int64) error
	Cancel(id int64, reason string) error
	MarkNoShow(id int64) error
	MarkAttended(id int64, userWorkoutID int64) error
	CountActiveForSession(sessionID int64) (int, error)
}
