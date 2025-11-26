package domain

import "time"

// DataChangeLog records changes (updates and deletes) to data records
// This provides an audit trail with before/after values for data modifications
type DataChangeLog struct {
	ID           int64     `json:"id" db:"id"`
	EntityType   string    `json:"entity_type" db:"entity_type"`     // wod, movement, workout, user_workout, etc.
	EntityID     int64     `json:"entity_id" db:"entity_id"`         // ID of the record that was changed
	EntityName   string    `json:"entity_name" db:"entity_name"`     // Human-readable name (e.g., WOD name, movement name)
	Operation    string    `json:"operation" db:"operation"`         // update, delete
	UserID       int64     `json:"user_id" db:"user_id"`             // User who made the change
	UserEmail    string    `json:"user_email" db:"user_email"`       // Email for display (denormalized)
	BeforeValues *string   `json:"before_values" db:"before_values"` // JSON of record before change
	AfterValues  *string   `json:"after_values" db:"after_values"`   // JSON of record after change (NULL for deletes)
	IPAddress    *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string   `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Operation types
const (
	OperationUpdate = "update"
	OperationDelete = "delete"
)

// Entity types
const (
	EntityTypeWOD                 = "wod"
	EntityTypeMovement            = "movement"
	EntityTypeWorkout             = "workout"
	EntityTypeUserWorkout         = "user_workout"
	EntityTypeUserWorkoutMovement = "user_workout_movement"
	EntityTypeUserWorkoutWOD      = "user_workout_wod"
	EntityTypeUser                = "user"
)

// DataChangeLogRepository defines the interface for data change log access
type DataChangeLogRepository interface {
	// Create creates a new data change log entry
	Create(log *DataChangeLog) error

	// GetByID retrieves a single data change log by ID
	GetByID(id int64) (*DataChangeLog, error)

	// List retrieves data change logs with pagination and filters
	List(filters DataChangeLogFilters, limit, offset int) ([]*DataChangeLog, error)

	// Count returns the total number of logs matching the filters
	Count(filters DataChangeLogFilters) (int, error)

	// GetByEntityID retrieves all changes for a specific entity
	GetByEntityID(entityType string, entityID int64, limit, offset int) ([]*DataChangeLog, error)

	// GetByUserID retrieves all changes made by a specific user
	GetByUserID(userID int64, limit, offset int) ([]*DataChangeLog, error)

	// DeleteOlderThan deletes logs older than the specified time (for cleanup)
	DeleteOlderThan(before time.Time) (int, error)
}

// DataChangeLogFilters represents filter options for querying data change logs
type DataChangeLogFilters struct {
	EntityType *string    // Filter by entity type
	EntityID   *int64     // Filter by specific entity
	Operation  *string    // Filter by operation type
	UserID     *int64     // Filter by user who made the change
	UserEmail  *string    // Filter by user email (partial match)
	StartDate  *time.Time // Filter logs after this date
	EndDate    *time.Time // Filter logs before this date
}
