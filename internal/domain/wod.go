package domain

import "time"

// WOD represents a Workout of the Day (CrossFit benchmark workout)
// WODs are predefined workouts like "Fran", "Murph", "Helen", etc.
// Standard WODs are pre-seeded, users can also create custom WODs
type WOD struct {
	ID          int64      `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Source      string     `json:"source,omitempty" db:"source"`           // CrossFit, Other Coach, Self-recorded
	Type        string     `json:"type,omitempty" db:"type"`               // Benchmark, Hero, Girl, Notables, Games, Endurance, Self-created
	Regime      string     `json:"regime,omitempty" db:"regime"`           // EMOM, AMRAP, Fastest Time, Slowest Round, Get Stronger, Skills
	ScoreType   string     `json:"score_type,omitempty" db:"score_type"`   // Time (HH:MM:SS), Rounds+Reps, Max Weight
	Description string     `json:"description,omitempty" db:"description"` // Full WOD description/instructions
	URL         *string    `json:"url,omitempty" db:"url"`                 // Optional video or reference URL
	Notes       *string    `json:"notes,omitempty" db:"notes"`             // Additional notes
	IsStandard  bool       `json:"is_standard" db:"is_standard"`           // TRUE for pre-seeded WODs, FALSE for user-created
	CreatedBy   *int64     `json:"created_by,omitempty" db:"created_by"`   // User ID if custom WOD (NULL for standard)
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// WODRepository defines the interface for WOD data access
type WODRepository interface {
	// Create creates a new custom WOD
	Create(wod *WOD) error

	// GetByID retrieves a WOD by ID
	GetByID(id int64) (*WOD, error)

	// GetByName retrieves a WOD by name
	GetByName(name string) (*WOD, error)

	// List retrieves all WODs with optional filtering
	List(filters map[string]interface{}) ([]*WOD, error)

	// ListStandard retrieves all standard (pre-seeded) WODs
	ListStandard() ([]*WOD, error)

	// ListByUser retrieves all custom WODs created by a specific user
	ListByUser(userID int64) ([]*WOD, error)

	// Update updates an existing WOD (only for user-created WODs)
	Update(wod *WOD) error

	// Delete deletes a WOD (only for user-created WODs)
	Delete(id int64) error

	// Search searches WODs by name (partial match)
	Search(query string) ([]*WOD, error)
}
