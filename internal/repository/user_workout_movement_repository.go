package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// UserWorkoutMovementRepository implements domain.UserWorkoutMovementRepository
type UserWorkoutMovementRepository struct {
	db *sql.DB
}

// NewUserWorkoutMovementRepository creates a new user workout movement repository
func NewUserWorkoutMovementRepository(db *sql.DB) *UserWorkoutMovementRepository {
	return &UserWorkoutMovementRepository{db: db}
}

// Create creates a new user workout movement performance record
func (r *UserWorkoutMovementRepository) Create(uwm *domain.UserWorkoutMovement) error {
	uwm.CreatedAt = time.Now()
	uwm.UpdatedAt = time.Now()

	query := `INSERT INTO user_workout_movements (user_workout_id, movement_id, sets, reps, weight, time, distance, notes, is_pr, order_index, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, uwm.UserWorkoutID, uwm.MovementID, uwm.Sets, uwm.Reps, uwm.Weight, uwm.Time, uwm.Distance, uwm.Notes, uwm.IsPR, uwm.OrderIndex, uwm.CreatedAt, uwm.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user workout movement: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get user workout movement ID: %w", err)
	}

	uwm.ID = id
	return nil
}

// CreateBatch creates multiple user workout movement records at once
func (r *UserWorkoutMovementRepository) CreateBatch(movements []*domain.UserWorkoutMovement) error {
	if len(movements) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO user_workout_movements (user_workout_id, movement_id, sets, reps, weight, time, distance, notes, is_pr, order_index, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, uwm := range movements {
		uwm.CreatedAt = now
		uwm.UpdatedAt = now

		result, err := stmt.Exec(uwm.UserWorkoutID, uwm.MovementID, uwm.Sets, uwm.Reps, uwm.Weight, uwm.Time, uwm.Distance, uwm.Notes, uwm.IsPR, uwm.OrderIndex, uwm.CreatedAt, uwm.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert user workout movement: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get user workout movement ID: %w", err)
		}
		uwm.ID = id
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a user workout movement by ID
func (r *UserWorkoutMovementRepository) GetByID(id int64) (*domain.UserWorkoutMovement, error) {
	query := `SELECT id, user_workout_id, movement_id, sets, reps, weight, time, distance, notes, order_index, created_at, updated_at
	          FROM user_workout_movements WHERE id = ?`

	uwm := &domain.UserWorkoutMovement{}
	var sets sql.NullInt64
	var reps sql.NullInt64
	var weight sql.NullFloat64
	var time sql.NullInt64
	var distance sql.NullFloat64

	err := r.db.QueryRow(query, id).Scan(&uwm.ID, &uwm.UserWorkoutID, &uwm.MovementID, &sets, &reps, &weight, &time, &distance, &uwm.Notes, &uwm.OrderIndex, &uwm.CreatedAt, &uwm.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user workout movement: %w", err)
	}

	if sets.Valid {
		s := int(sets.Int64)
		uwm.Sets = &s
	}
	if reps.Valid {
		r := int(reps.Int64)
		uwm.Reps = &r
	}
	if weight.Valid {
		uwm.Weight = &weight.Float64
	}
	if time.Valid {
		t := int(time.Int64)
		uwm.Time = &t
	}
	if distance.Valid {
		uwm.Distance = &distance.Float64
	}

	return uwm, nil
}

// GetByUserWorkoutID retrieves all movements for a specific logged workout
func (r *UserWorkoutMovementRepository) GetByUserWorkoutID(userWorkoutID int64) ([]*domain.UserWorkoutMovement, error) {
	query := `
		SELECT uwm.id, uwm.user_workout_id, uwm.movement_id, uwm.sets, uwm.reps, uwm.weight, uwm.time, uwm.distance,
		       uwm.notes, uwm.order_index, uwm.created_at, uwm.updated_at,
		       m.id as movement_id, m.name, m.description, m.type, m.is_standard, m.created_by, m.created_at, m.updated_at
		FROM user_workout_movements uwm
		JOIN movements m ON uwm.movement_id = m.id
		WHERE uwm.user_workout_id = ?
		ORDER BY uwm.order_index`

	rows, err := r.db.Query(query, userWorkoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workout movements: %w", err)
	}
	defer rows.Close()

	var movements []*domain.UserWorkoutMovement
	for rows.Next() {
		uwm := &domain.UserWorkoutMovement{
			Movement: &domain.Movement{},
		}
		var sets sql.NullInt64
		var reps sql.NullInt64
		var weight sql.NullFloat64
		var time sql.NullInt64
		var distance sql.NullFloat64
		var createdBy sql.NullInt64

		err := rows.Scan(&uwm.ID, &uwm.UserWorkoutID, &uwm.MovementID, &sets, &reps, &weight, &time, &distance,
			&uwm.Notes, &uwm.OrderIndex, &uwm.CreatedAt, &uwm.UpdatedAt,
			&uwm.Movement.ID, &uwm.Movement.Name, &uwm.Movement.Description, &uwm.Movement.Type, &uwm.Movement.IsStandard, &createdBy, &uwm.Movement.CreatedAt, &uwm.Movement.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user workout movement: %w", err)
		}

		if sets.Valid {
			s := int(sets.Int64)
			uwm.Sets = &s
		}
		if reps.Valid {
			r := int(reps.Int64)
			uwm.Reps = &r
		}
		if weight.Valid {
			uwm.Weight = &weight.Float64
		}
		if time.Valid {
			t := int(time.Int64)
			uwm.Time = &t
		}
		if distance.Valid {
			uwm.Distance = &distance.Float64
		}
		if createdBy.Valid {
			cb := createdBy.Int64
			uwm.Movement.CreatedBy = &cb
		}

		movements = append(movements, uwm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user workout movements: %w", err)
	}

	return movements, nil
}

// Update updates an existing user workout movement
func (r *UserWorkoutMovementRepository) Update(uwm *domain.UserWorkoutMovement) error {
	uwm.UpdatedAt = time.Now()

	query := `UPDATE user_workout_movements
	          SET sets = ?, reps = ?, weight = ?, time = ?, distance = ?, notes = ?, order_index = ?, updated_at = ?
	          WHERE id = ?`

	result, err := r.db.Exec(query, uwm.Sets, uwm.Reps, uwm.Weight, uwm.Time, uwm.Distance, uwm.Notes, uwm.OrderIndex, uwm.UpdatedAt, uwm.ID)
	if err != nil {
		return fmt.Errorf("failed to update user workout movement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user workout movement not found")
	}

	return nil
}

// Delete deletes a user workout movement
func (r *UserWorkoutMovementRepository) Delete(id int64) error {
	query := `DELETE FROM user_workout_movements WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user workout movement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user workout movement not found")
	}

	return nil
}

// DeleteByUserWorkoutID deletes all movements for a logged workout
func (r *UserWorkoutMovementRepository) DeleteByUserWorkoutID(userWorkoutID int64) error {
	query := `DELETE FROM user_workout_movements WHERE user_workout_id = ?`

	_, err := r.db.Exec(query, userWorkoutID)
	if err != nil {
		return fmt.Errorf("failed to delete user workout movements: %w", err)
	}

	return nil
}

// GetMaxWeightForMovement retrieves the maximum weight for a specific movement for a user
func (r *UserWorkoutMovementRepository) GetMaxWeightForMovement(userID, movementID int64) (*float64, error) {
	query := `
		SELECT MAX(uwm.weight)
		FROM user_workout_movements uwm
		INNER JOIN user_workouts uw ON uwm.user_workout_id = uw.id
		WHERE uw.user_id = ? AND uwm.movement_id = ? AND uwm.weight IS NOT NULL`

	var maxWeight sql.NullFloat64
	err := r.db.QueryRow(query, userID, movementID).Scan(&maxWeight)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get max weight: %w", err)
	}

	if !maxWeight.Valid {
		return nil, nil
	}

	return &maxWeight.Float64, nil
}

// GetPRMovements retrieves recent PR-flagged movements for a user
func (r *UserWorkoutMovementRepository) GetPRMovements(userID int64, limit int) ([]*domain.UserWorkoutMovement, error) {
	query := `
		SELECT uwm.id, uwm.user_workout_id, uwm.movement_id, uwm.sets, uwm.reps, uwm.weight, uwm.time, uwm.distance,
		       uwm.notes, uwm.is_pr, uwm.order_index, uwm.created_at, uwm.updated_at,
		       m.name, m.type,
		       uw.workout_date
		FROM user_workout_movements uwm
		JOIN movements m ON uwm.movement_id = m.id
		JOIN user_workouts uw ON uwm.user_workout_id = uw.id
		WHERE uw.user_id = ? AND uwm.is_pr = 1
		ORDER BY uw.workout_date DESC, uwm.created_at DESC
		LIMIT ?`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR movements: %w", err)
	}
	defer rows.Close()

	var movements []*domain.UserWorkoutMovement
	for rows.Next() {
		uwm := &domain.UserWorkoutMovement{}
		var sets sql.NullInt64
		var reps sql.NullInt64
		var weight sql.NullFloat64
		var timeVal sql.NullInt64
		var distance sql.NullFloat64
		var workoutDate time.Time

		err := rows.Scan(&uwm.ID, &uwm.UserWorkoutID, &uwm.MovementID, &sets, &reps, &weight, &timeVal, &distance,
			&uwm.Notes, &uwm.IsPR, &uwm.OrderIndex, &uwm.CreatedAt, &uwm.UpdatedAt,
			&uwm.MovementName, &uwm.MovementType, &workoutDate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan PR movement: %w", err)
		}

		if sets.Valid {
			s := int(sets.Int64)
			uwm.Sets = &s
		}
		if reps.Valid {
			r := int(reps.Int64)
			uwm.Reps = &r
		}
		if weight.Valid {
			uwm.Weight = &weight.Float64
		}
		if timeVal.Valid {
			t := int(timeVal.Int64)
			uwm.Time = &t
		}
		if distance.Valid {
			uwm.Distance = &distance.Float64
		}

		movements = append(movements, uwm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate PR movements: %w", err)
	}

	return movements, nil
}

// UpdatePRFlag updates the is_pr flag for a user workout movement
func (r *UserWorkoutMovementRepository) UpdatePRFlag(id int64, isPR bool) error {
	query := `UPDATE user_workout_movements SET is_pr = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	result, err := r.db.Exec(query, isPR, id)
	if err != nil {
		return fmt.Errorf("failed to update PR flag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user workout movement found with id %d", id)
	}

	return nil
}

// GetByUserIDAndMovementID retrieves all movement performance records for a specific user and movement
func (r *UserWorkoutMovementRepository) GetByUserIDAndMovementID(userID, movementID int64, limit int) ([]*domain.UserWorkoutMovement, error) {
	query := `
		SELECT uwm.id, uwm.user_workout_id, uwm.movement_id, uwm.weight, uwm.reps,
		       uwm.sets, uwm.time, uwm.notes, uwm.is_pr, uwm.order_index,
		       uwm.created_at, uwm.updated_at,
		       m.name, m.type,
		       uw.workout_date
		FROM user_workout_movements uwm
		JOIN movements m ON uwm.movement_id = m.id
		JOIN user_workouts uw ON uwm.user_workout_id = uw.id
		WHERE uw.user_id = ? AND uwm.movement_id = ?
		ORDER BY uw.workout_date DESC, uwm.created_at DESC
		LIMIT ?`

	rows, err := r.db.Query(query, userID, movementID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query movement performances: %w", err)
	}
	defer rows.Close()

	var movements []*domain.UserWorkoutMovement
	for rows.Next() {
		uwm := &domain.UserWorkoutMovement{}
		var weight sql.NullFloat64
		var reps sql.NullInt64
		var sets sql.NullInt64
		var timeVal sql.NullInt64
		var workoutDate time.Time

		err := rows.Scan(&uwm.ID, &uwm.UserWorkoutID, &uwm.MovementID, &weight, &reps,
			&sets, &timeVal, &uwm.Notes, &uwm.IsPR, &uwm.OrderIndex,
			&uwm.CreatedAt, &uwm.UpdatedAt,
			&uwm.MovementName, &uwm.MovementType, &workoutDate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan movement performance: %w", err)
		}

		if weight.Valid {
			uwm.Weight = &weight.Float64
		}
		if reps.Valid {
			r := int(reps.Int64)
			uwm.Reps = &r
		}
		if sets.Valid {
			s := int(sets.Int64)
			uwm.Sets = &s
		}
		if timeVal.Valid {
			t := int(timeVal.Int64)
			uwm.Time = &t
		}

		// Store workout date from user_workouts table
		uwm.WorkoutDate = workoutDate

		movements = append(movements, uwm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating movement performances: %w", err)
	}

	return movements, nil
}
