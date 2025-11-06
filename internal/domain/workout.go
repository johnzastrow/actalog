package domain

import (
	"time"
)

// WorkoutType represents the type of workout
type WorkoutType string

const (
	WorkoutTypeNamedWOD WorkoutType = "named_wod"
	WorkoutTypeCustom   WorkoutType = "custom"
)

// Workout represents a logged workout session
type Workout struct {
	ID          int64       `json:"id" db:"id"`
	UserID      int64       `json:"user_id" db:"user_id"`
	WorkoutDate time.Time   `json:"workout_date" db:"workout_date"`
	WorkoutType WorkoutType `json:"workout_type" db:"workout_type"`
	WorkoutName string      `json:"workout_name,omitempty" db:"workout_name"` // For named WODs
	Notes       string      `json:"notes,omitempty" db:"notes"`
	TotalTime   *int        `json:"total_time,omitempty" db:"total_time"` // in seconds, for time-based workouts
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`

	// Related data (not stored directly in workout table)
	Movements []*WorkoutMovement `json:"movements,omitempty" db:"-"`
}

// WorkoutRepository defines the interface for workout data access
type WorkoutRepository interface {
	Create(workout *Workout) error
	GetByID(id int64) (*Workout, error)
	GetByUserID(userID int64, limit, offset int) ([]*Workout, error)
	GetByUserIDAndDateRange(userID int64, startDate, endDate time.Time) ([]*Workout, error)
	Update(workout *Workout) error
	Delete(id int64) error
	Count(userID int64) (int64, error)
}
