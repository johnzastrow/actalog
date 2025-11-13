package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// WorkoutMovementRepository implements domain.WorkoutMovementRepository
type WorkoutMovementRepository struct {
	db *sql.DB
}

// NewWorkoutMovementRepository creates a new workout movement repository
func NewWorkoutMovementRepository(db *sql.DB) *WorkoutMovementRepository {
	return &WorkoutMovementRepository{db: db}
}

// Create creates a new workout movement
func (r *WorkoutMovementRepository) Create(wm *domain.WorkoutMovement) error {
	wm.CreatedAt = time.Now()
	wm.UpdatedAt = time.Now()

	query := `INSERT INTO workout_movements (workout_id, movement_id, weight, sets, reps, time, distance, is_rx, is_pr, notes, order_index, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, wm.WorkoutID, wm.MovementID, wm.Weight, wm.Sets, wm.Reps, wm.Time, wm.Distance, wm.IsRx, wm.IsPR, wm.Notes, wm.OrderIndex, wm.CreatedAt, wm.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create workout movement: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get workout movement ID: %w", err)
	}

	wm.ID = id
	return nil
}

// GetByID retrieves a workout movement by ID
func (r *WorkoutMovementRepository) GetByID(id int64) (*domain.WorkoutMovement, error) {
	query := `SELECT id, workout_id, movement_id, weight, sets, reps, time, distance, is_rx, is_pr, notes, order_index, created_at, updated_at FROM workout_movements WHERE id = ?`

	wm := &domain.WorkoutMovement{}
	var weight sql.NullFloat64
	var sets sql.NullInt64
	var reps sql.NullInt64
	var time sql.NullInt64
	var distance sql.NullFloat64
	var notes sql.NullString

	err := r.db.QueryRow(query, id).Scan(&wm.ID, &wm.WorkoutID, &wm.MovementID, &weight, &sets, &reps, &time, &distance, &wm.IsRx, &wm.IsPR, &notes, &wm.OrderIndex, &wm.CreatedAt, &wm.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workout movement: %w", err)
	}

	if weight.Valid {
		wm.Weight = &weight.Float64
	}
	if sets.Valid {
		s := int(sets.Int64)
		wm.Sets = &s
	}
	if reps.Valid {
		r := int(reps.Int64)
		wm.Reps = &r
	}
	if time.Valid {
		t := int(time.Int64)
		wm.Time = &t
	}
	if distance.Valid {
		wm.Distance = &distance.Float64
	}
	if notes.Valid {
		wm.Notes = notes.String
	}

	return wm, nil
}

// GetByWorkoutID retrieves all workout movements for a specific workout template
func (r *WorkoutMovementRepository) GetByWorkoutID(workoutID int64) ([]*domain.WorkoutMovement, error) {
	query := `
		SELECT ws.id, ws.workout_id, ws.movement_id, ws.weight, ws.sets, ws.reps, ws.time, ws.distance,
		       ws.is_rx, ws.is_pr, ws.notes, ws.order_index, ws.created_at, ws.updated_at,
		       m.name as movement_name, m.type as movement_type
		FROM workout_movements ws
		JOIN movements m ON ws.movement_id = m.id
		WHERE ws.workout_id = ?
		ORDER BY ws.order_index`

	rows, err := r.db.Query(query, workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout movements: %w", err)
	}
	defer rows.Close()

	var workoutMovements []*domain.WorkoutMovement
	for rows.Next() {
		wm := &domain.WorkoutMovement{}
		var weight sql.NullFloat64
		var sets sql.NullInt64
		var reps sql.NullInt64
		var time sql.NullInt64
		var distance sql.NullFloat64
		var notes sql.NullString
		var movementName string
		var movementType string

		err := rows.Scan(&wm.ID, &wm.WorkoutID, &wm.MovementID, &weight, &sets, &reps, &time, &distance,
			&wm.IsRx, &wm.IsPR, &notes, &wm.OrderIndex, &wm.CreatedAt, &wm.UpdatedAt,
			&movementName, &movementType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workout movement: %w", err)
		}

		if weight.Valid {
			wm.Weight = &weight.Float64
		}
		if sets.Valid {
			s := int(sets.Int64)
			wm.Sets = &s
		}
		if reps.Valid {
			r := int(reps.Int64)
			wm.Reps = &r
		}
		if time.Valid {
			t := int(time.Int64)
			wm.Time = &t
		}
		if distance.Valid {
			wm.Distance = &distance.Float64
		}
		if notes.Valid {
			wm.Notes = notes.String
		}

		wm.Movement = &domain.Movement{
			ID:   wm.MovementID,
			Name: movementName,
			Type: domain.MovementType(movementType),
		}

		workoutMovements = append(workoutMovements, wm)
	}

	return workoutMovements, rows.Err()
}

// GetByUserIDAndMovementID retrieves workout movements for a user and specific movement
// This now queries through user_workouts junction table since workouts are templates
func (r *WorkoutMovementRepository) GetByUserIDAndMovementID(userID, movementID int64, limit int) ([]*domain.WorkoutMovement, error) {
	query := `
		SELECT ws.id, ws.workout_id, ws.movement_id, ws.weight, ws.sets, ws.reps, ws.time, ws.distance,
		       ws.is_rx, ws.is_pr, ws.notes, ws.order_index, ws.created_at, ws.updated_at
		FROM workout_movements ws
		INNER JOIN user_workouts uw ON ws.workout_id = uw.workout_id
		WHERE uw.user_id = ? AND ws.movement_id = ?
		ORDER BY uw.workout_date DESC, ws.created_at DESC
		LIMIT ?`

	rows, err := r.db.Query(query, userID, movementID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workout movements: %w", err)
	}
	defer rows.Close()

	var workoutMovements []*domain.WorkoutMovement
	for rows.Next() {
		wm := &domain.WorkoutMovement{}
		var weight sql.NullFloat64
		var sets sql.NullInt64
		var reps sql.NullInt64
		var time sql.NullInt64
		var distance sql.NullFloat64
		var notes sql.NullString

		err := rows.Scan(&wm.ID, &wm.WorkoutID, &wm.MovementID, &weight, &sets, &reps, &time, &distance,
			&wm.IsRx, &wm.IsPR, &notes, &wm.OrderIndex, &wm.CreatedAt, &wm.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workout movement: %w", err)
		}

		if weight.Valid {
			wm.Weight = &weight.Float64
		}
		if sets.Valid {
			s := int(sets.Int64)
			wm.Sets = &s
		}
		if reps.Valid {
			r := int(reps.Int64)
			wm.Reps = &r
		}
		if time.Valid {
			t := int(time.Int64)
			wm.Time = &t
		}
		if distance.Valid {
			wm.Distance = &distance.Float64
		}
		if notes.Valid {
			wm.Notes = notes.String
		}

		workoutMovements = append(workoutMovements, wm)
	}

	return workoutMovements, rows.Err()
}

// Update updates a workout movement
func (r *WorkoutMovementRepository) Update(wm *domain.WorkoutMovement) error {
	wm.UpdatedAt = time.Now()

	query := `UPDATE workout_movements
	          SET movement_id = ?, weight = ?, sets = ?, reps = ?,
	              time = ?, distance = ?, is_rx = ?, is_pr = ?,
	              notes = ?, order_index = ?, updated_at = ?
	          WHERE id = ?`

	result, err := r.db.Exec(query, wm.MovementID, wm.Weight, wm.Sets, wm.Reps, wm.Time, wm.Distance, wm.IsRx, wm.IsPR, wm.Notes, wm.OrderIndex, wm.UpdatedAt, wm.ID)
	if err != nil {
		return fmt.Errorf("failed to update workout movement: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("workout movement not found")
	}

	return nil
}

// Delete deletes a workout movement
func (r *WorkoutMovementRepository) Delete(id int64) error {
	query := `DELETE FROM workout_movements WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete workout movement: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("workout movement not found")
	}

	return nil
}

// DeleteByWorkoutID deletes all movements for a workout template
func (r *WorkoutMovementRepository) DeleteByWorkoutID(workoutID int64) error {
	query := `DELETE FROM workout_movements WHERE workout_id = ?`

	if _, err := r.db.Exec(query, workoutID); err != nil {
		return fmt.Errorf("failed to delete workout movements: %w", err)
	}

	return nil
}

// GetPersonalRecords retrieves all personal records for a user
// Updated to work with new schema where user_workouts is the junction table
func (r *WorkoutMovementRepository) GetPersonalRecords(userID int64) ([]*domain.PersonalRecord, error) {
	query := `
		SELECT
			m.id as movement_id,
			m.name as movement_name,
			MAX(ws.weight) as max_weight,
			MAX(ws.reps) as max_reps,
			MIN(ws.time) as best_time,
			uw.id as user_workout_id,
			ws.workout_id,
			uw.workout_date
		FROM workout_movements ws
		INNER JOIN user_workouts uw ON ws.workout_id = uw.workout_id
		INNER JOIN movements m ON ws.movement_id = m.id
		WHERE uw.user_id = ?
		GROUP BY m.id, m.name
		ORDER BY m.name`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get personal records: %w", err)
	}
	defer rows.Close()

	var records []*domain.PersonalRecord
	for rows.Next() {
		pr := &domain.PersonalRecord{}
		var maxWeight sql.NullFloat64
		var maxReps sql.NullInt64
		var bestTime sql.NullInt64

		err := rows.Scan(
			&pr.MovementID,
			&pr.MovementName,
			&maxWeight,
			&maxReps,
			&bestTime,
			&pr.UserWorkoutID,
			&pr.WorkoutID,
			&pr.WorkoutDate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan personal record: %w", err)
		}

		if maxWeight.Valid {
			pr.MaxWeight = &maxWeight.Float64
		}
		if maxReps.Valid {
			r := int(maxReps.Int64)
			pr.MaxReps = &r
		}
		if bestTime.Valid {
			t := int(bestTime.Int64)
			pr.BestTime = &t
		}

		records = append(records, pr)
	}

	return records, rows.Err()
}

// GetMaxWeightForMovement retrieves the maximum weight for a specific movement for a user
// Updated to work with new schema where user_workouts is the junction table
func (r *WorkoutMovementRepository) GetMaxWeightForMovement(userID, movementID int64) (*float64, error) {
	query := `
		SELECT MAX(ws.weight)
		FROM workout_movements ws
		INNER JOIN user_workouts uw ON ws.workout_id = uw.workout_id
		WHERE uw.user_id = ? AND ws.movement_id = ? AND ws.weight IS NOT NULL`

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
// Updated to work with new schema where user_workouts is the junction table
func (r *WorkoutMovementRepository) GetPRMovements(userID int64, limit int) ([]*domain.WorkoutMovement, error) {
	query := `
		SELECT ws.id, ws.workout_id, ws.movement_id, ws.weight, ws.sets, ws.reps, ws.time, ws.distance,
		       ws.is_rx, ws.is_pr, ws.notes, ws.order_index, ws.created_at, ws.updated_at,
		       m.name as movement_name, m.type as movement_type, m.description as movement_description
		FROM workout_movements ws
		INNER JOIN user_workouts uw ON ws.workout_id = uw.workout_id
		INNER JOIN movements m ON ws.movement_id = m.id
		WHERE uw.user_id = ? AND ws.is_pr = 1
		ORDER BY uw.workout_date DESC, ws.created_at DESC
		LIMIT ?`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR movements: %w", err)
	}
	defer rows.Close()

	var workoutMovements []*domain.WorkoutMovement
	for rows.Next() {
		wm := &domain.WorkoutMovement{}
		var weight sql.NullFloat64
		var sets sql.NullInt64
		var reps sql.NullInt64
		var time sql.NullInt64
		var distance sql.NullFloat64
		var notes sql.NullString
		var movementName string
		var movementType string
		var movementDescription sql.NullString

		err := rows.Scan(&wm.ID, &wm.WorkoutID, &wm.MovementID, &weight, &sets, &reps, &time, &distance,
			&wm.IsRx, &wm.IsPR, &notes, &wm.OrderIndex, &wm.CreatedAt, &wm.UpdatedAt,
			&movementName, &movementType, &movementDescription)
		if err != nil {
			return nil, fmt.Errorf("failed to scan PR movement: %w", err)
		}

		if weight.Valid {
			wm.Weight = &weight.Float64
		}
		if sets.Valid {
			s := int(sets.Int64)
			wm.Sets = &s
		}
		if reps.Valid {
			r := int(reps.Int64)
			wm.Reps = &r
		}
		if time.Valid {
			t := int(time.Int64)
			wm.Time = &t
		}
		if distance.Valid {
			wm.Distance = &distance.Float64
		}
		if notes.Valid {
			wm.Notes = notes.String
		}

		wm.Movement = &domain.Movement{
			ID:   wm.MovementID,
			Name: movementName,
			Type: domain.MovementType(movementType),
		}
		if movementDescription.Valid {
			wm.Movement.Description = movementDescription.String
		}

		workoutMovements = append(workoutMovements, wm)
	}

	return workoutMovements, rows.Err()
}
