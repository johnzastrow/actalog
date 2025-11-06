package domain

import (
	"time"
)

// MovementType represents the type of movement
type MovementType string

const (
	MovementTypeWeightlifting MovementType = "weightlifting"
	MovementTypeBodyweight    MovementType = "bodyweight"
	MovementTypeCardio        MovementType = "cardio"
	MovementTypeGymnastics    MovementType = "gymnastics"
)

// Movement represents a specific exercise or movement
type Movement struct {
	ID          int64        `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description,omitempty" db:"description"`
	Type        MovementType `json:"type" db:"type"`
	IsStandard  bool         `json:"is_standard" db:"is_standard"`         // True for predefined movements
	CreatedBy   *int64       `json:"created_by,omitempty" db:"created_by"` // User ID if custom
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}

// WorkoutMovement represents a movement performed in a workout
type WorkoutMovement struct {
	ID         int64     `json:"id" db:"id"`
	WorkoutID  int64     `json:"workout_id" db:"workout_id"`
	MovementID int64     `json:"movement_id" db:"movement_id"`
	Weight     *float64  `json:"weight,omitempty" db:"weight"` // in lbs or kg
	Sets       *int      `json:"sets,omitempty" db:"sets"`
	Reps       *int      `json:"reps,omitempty" db:"reps"`
	Time       *int      `json:"time,omitempty" db:"time"`         // in seconds
	Distance   *float64  `json:"distance,omitempty" db:"distance"` // in meters or miles
	IsRx       bool      `json:"is_rx" db:"is_rx"`                 // Prescribed weight/standard
	Notes      string    `json:"notes,omitempty" db:"notes"`
	OrderIndex int       `json:"order_index" db:"order_index"` // Order in the workout
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`

	// Related data
	Movement *Movement `json:"movement,omitempty" db:"-"`
}

// MovementRepository defines the interface for movement data access
type MovementRepository interface {
	Create(movement *Movement) error
	GetByID(id int64) (*Movement, error)
	GetByName(name string) (*Movement, error)
	ListStandard() ([]*Movement, error)
	ListByUser(userID int64) ([]*Movement, error)
	Update(movement *Movement) error
	Delete(id int64) error
	Search(query string, limit int) ([]*Movement, error)
}

// WorkoutMovementRepository defines the interface for workout movement data access
type WorkoutMovementRepository interface {
	Create(wm *WorkoutMovement) error
	GetByID(id int64) (*WorkoutMovement, error)
	GetByWorkoutID(workoutID int64) ([]*WorkoutMovement, error)
	GetByUserIDAndMovementID(userID, movementID int64, limit int) ([]*WorkoutMovement, error)
	Update(wm *WorkoutMovement) error
	Delete(id int64) error
	DeleteByWorkoutID(workoutID int64) error
}
