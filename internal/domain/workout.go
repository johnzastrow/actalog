package domain

import (
	"time"
)

// Workout represents a reusable workout template
// Templates can be used by multiple users multiple times
// User-specific workout instances are tracked in UserWorkout
type Workout struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`                       // Template name (e.g., "Monday Strength", "Hero WOD")
	Notes     *string   `json:"notes,omitempty" db:"notes"`           // General template notes/description
	CreatedBy *int64    `json:"created_by,omitempty" db:"created_by"` // User who created (NULL for standard templates)
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Related data (not stored directly in workout table, loaded via joins)
	Movements []*WorkoutMovement       `json:"movements,omitempty" db:"-"` // Strength movements in this template
	WODs      []*WorkoutWODWithDetails `json:"wods,omitempty" db:"-"`      // WODs in this template
}

// WorkoutWithCreator extends Workout with creator information (for admin views)
type WorkoutWithCreator struct {
	Workout
	CreatorEmail string `json:"creator_email,omitempty"`
	CreatorName  string `json:"creator_name,omitempty"`
}

// WorkoutWithUsageStats includes usage statistics for a template
type WorkoutWithUsageStats struct {
	Workout
	TimesUsed  int        `json:"times_used"`             // How many times this template has been logged
	LastUsedAt *time.Time `json:"last_used_at,omitempty"` // When it was last logged
}

// WorkoutRepository defines the interface for workout template data access
type WorkoutRepository interface {
	// Create creates a new workout template
	Create(workout *Workout) error

	// GetByID retrieves a workout template by ID
	GetByID(id int64) (*Workout, error)

	// GetByIDWithDetails retrieves a workout with movements and WODs
	GetByIDWithDetails(id int64) (*Workout, error)

	// List retrieves all workout templates with optional filtering
	List(filters map[string]interface{}, limit, offset int) ([]*Workout, error)

	// ListByUser retrieves all workout templates created by a specific user
	ListByUser(userID int64, limit, offset int) ([]*Workout, error)

	// ListStandard retrieves all standard (system) workout templates
	ListStandard(limit, offset int) ([]*Workout, error)

	// ListAllUserCreated retrieves all user-created workout templates (for admin)
	ListAllUserCreated(limit, offset int) ([]*Workout, error)

	// ListAllUserCreatedWithUserInfo retrieves all user-created workouts with creator info (for admin)
	ListAllUserCreatedWithUserInfo(limit, offset int) ([]*WorkoutWithCreator, error)

	// ListAllUserCreatedWithUserInfoFiltered retrieves all user-created workouts with filters (for admin)
	ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, creator string) ([]*WorkoutWithCreator, int64, error)

	// CountAllUserCreated counts all user-created workout templates
	CountAllUserCreated() (int64, error)

	// Update updates an existing workout template
	Update(workout *Workout) error

	// Delete deletes a workout template
	Delete(id int64) error

	// Search searches workout templates by name
	Search(query string, limit int) ([]*Workout, error)

	// Count counts total workout templates (optionally filtered by user)
	Count(userID *int64) (int64, error)

	// GetUsageStats gets usage statistics for a template
	GetUsageStats(workoutID int64) (*WorkoutWithUsageStats, error)

	// CopyToStandard creates a standard workout by copying a user-created one
	CopyToStandard(id int64, newName string) (*Workout, error)
}
