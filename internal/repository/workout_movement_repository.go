package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// SQLiteWorkoutMovementRepository implements WorkoutMovementRepository using SQLite
type SQLiteWorkoutMovementRepository struct {
	db *sql.DB
}

// NewSQLiteWorkoutMovementRepository creates a new SQLite workout movement repository
func NewSQLiteWorkoutMovementRepository(db *sql.DB) *SQLiteWorkoutMovementRepository {
	return &SQLiteWorkoutMovementRepository{db: db}
}

// Create creates a new workout movement
func (r *SQLiteWorkoutMovementRepository) Create(wm *domain.WorkoutMovement) error {
	query := `
		INSERT INTO workout_movements (workout_id, movement_id, weight, sets, reps, time, distance, is_rx, is_pr, notes, order_index, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	wm.CreatedAt = now
	wm.UpdatedAt = now

	result, err := r.db.Exec(
		query,
		wm.WorkoutID,
		wm.MovementID,
		wm.Weight,
		wm.Sets,
		wm.Reps,
		wm.Time,
		wm.Distance,
		wm.IsRx,
		wm.IsPR,
		wm.Notes,
		wm.OrderIndex,
		wm.CreatedAt,
		wm.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	wm.ID = id
	return nil
}

// GetByID retrieves a workout movement by ID
func (r *SQLiteWorkoutMovementRepository) GetByID(id int64) (*domain.WorkoutMovement, error) {
	query := `
		SELECT id, workout_id, movement_id, weight, sets, reps, time, distance, is_rx, is_pr, notes, order_index,
		       created_at, updated_at
		FROM workout_movements
		WHERE id = ?
	`

	wm := &domain.WorkoutMovement{}
	var weight sql.NullFloat64
	var sets sql.NullInt64
	var reps sql.NullInt64
	var timeVal sql.NullInt64
	var distance sql.NullFloat64
	var notes sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&wm.ID,
		&wm.WorkoutID,
		&wm.MovementID,
		&weight,
		&sets,
		&reps,
		&timeVal,
		&distance,
		&wm.IsRx,
		&wm.IsPR,
		&notes,
		&wm.OrderIndex,
		&wm.CreatedAt,
		&wm.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
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
	if timeVal.Valid {
		t := int(timeVal.Int64)
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

// GetByWorkoutID retrieves all workout movements for a specific workout
func (r *SQLiteWorkoutMovementRepository) GetByWorkoutID(workoutID int64) ([]*domain.WorkoutMovement, error) {
	query := `
		SELECT wm.id, wm.workout_id, wm.movement_id, wm.weight, wm.sets, wm.reps, wm.time, wm.distance,
		       wm.is_rx, wm.is_pr, wm.notes, wm.order_index, wm.created_at, wm.updated_at,
		       m.id, m.name, m.description, m.type, m.is_standard, m.created_by, m.created_at, m.updated_at
		FROM workout_movements wm
		JOIN movements m ON wm.movement_id = m.id
		WHERE wm.workout_id = ?
		ORDER BY wm.order_index
	`

	rows, err := r.db.Query(query, workoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workoutMovements []*domain.WorkoutMovement
	for rows.Next() {
		wm := &domain.WorkoutMovement{}
		movement := &domain.Movement{}
		var weight sql.NullFloat64
		var sets sql.NullInt64
		var reps sql.NullInt64
		var timeVal sql.NullInt64
		var distance sql.NullFloat64
		var wmNotes sql.NullString
		var movementDesc sql.NullString
		var movementCreatedBy sql.NullInt64

		err := rows.Scan(
			&wm.ID,
			&wm.WorkoutID,
			&wm.MovementID,
			&weight,
			&sets,
			&reps,
			&timeVal,
			&distance,
			&wm.IsRx,
			&wm.IsPR,
			&wmNotes,
			&wm.OrderIndex,
			&wm.CreatedAt,
			&wm.UpdatedAt,
			&movement.ID,
			&movement.Name,
			&movementDesc,
			&movement.Type,
			&movement.IsStandard,
			&movementCreatedBy,
			&movement.CreatedAt,
			&movement.UpdatedAt,
		)
		if err != nil {
			return nil, err
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
		if timeVal.Valid {
			t := int(timeVal.Int64)
			wm.Time = &t
		}
		if distance.Valid {
			wm.Distance = &distance.Float64
		}
		if wmNotes.Valid {
			wm.Notes = wmNotes.String
		}
		if movementDesc.Valid {
			movement.Description = movementDesc.String
		}
		if movementCreatedBy.Valid {
			movement.CreatedBy = &movementCreatedBy.Int64
		}

		wm.Movement = movement
		workoutMovements = append(workoutMovements, wm)
	}

	return workoutMovements, rows.Err()
}

// GetByUserIDAndMovementID retrieves workout movements for a user and specific movement
func (r *SQLiteWorkoutMovementRepository) GetByUserIDAndMovementID(userID, movementID int64, limit int) ([]*domain.WorkoutMovement, error) {
	query := `
		SELECT wm.id, wm.workout_id, wm.movement_id, wm.weight, wm.sets, wm.reps, wm.time, wm.distance,
		       wm.is_rx, wm.is_pr, wm.notes, wm.order_index, wm.created_at, wm.updated_at
		FROM workout_movements wm
		INNER JOIN workouts w ON wm.workout_id = w.id
		WHERE w.user_id = ? AND wm.movement_id = ?
		ORDER BY w.workout_date DESC, wm.created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, userID, movementID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workoutMovements []*domain.WorkoutMovement
	for rows.Next() {
		wm := &domain.WorkoutMovement{}
		var weight sql.NullFloat64
		var sets sql.NullInt64
		var reps sql.NullInt64
		var timeVal sql.NullInt64
		var distance sql.NullFloat64
		var notes sql.NullString

		err := rows.Scan(
			&wm.ID,
			&wm.WorkoutID,
			&wm.MovementID,
			&weight,
			&sets,
			&reps,
			&timeVal,
			&distance,
			&wm.IsRx,
			&wm.IsPR,
			&notes,
			&wm.OrderIndex,
			&wm.CreatedAt,
			&wm.UpdatedAt,
		)
		if err != nil {
			return nil, err
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
		if timeVal.Valid {
			t := int(timeVal.Int64)
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
func (r *SQLiteWorkoutMovementRepository) Update(wm *domain.WorkoutMovement) error {
	query := `
		UPDATE workout_movements
		SET movement_id = ?, weight = ?, sets = ?, reps = ?, time = ?, distance = ?,
		    is_rx = ?, is_pr = ?, notes = ?, order_index = ?, updated_at = ?
		WHERE id = ?
	`

	wm.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		wm.MovementID,
		wm.Weight,
		wm.Sets,
		wm.Reps,
		wm.Time,
		wm.Distance,
		wm.IsRx,
		wm.IsPR,
		wm.Notes,
		wm.OrderIndex,
		wm.UpdatedAt,
		wm.ID,
	)

	return err
}

// Delete deletes a workout movement
func (r *SQLiteWorkoutMovementRepository) Delete(id int64) error {
	query := `DELETE FROM workout_movements WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteByWorkoutID deletes all workout movements for a specific workout
// DeleteByWorkoutID deletes all movements for a workout
func (r *SQLiteWorkoutMovementRepository) DeleteByWorkoutID(workoutID int64) error {
	query := `DELETE FROM workout_movements WHERE workout_id = ?`
	_, err := r.db.Exec(query, workoutID)
	return err
}

// GetPersonalRecords retrieves all personal records for a user
func (r *SQLiteWorkoutMovementRepository) GetPersonalRecords(userID int64) ([]*domain.PersonalRecord, error) {
	query := `
		SELECT
			m.id as movement_id,
			m.name as movement_name,
			MAX(wm.weight) as max_weight,
			MAX(wm.reps) as max_reps,
			MIN(wm.time) as best_time,
			wm.workout_id,
			w.workout_date
		FROM workout_movements wm
		INNER JOIN workouts w ON wm.workout_id = w.id
		INNER JOIN movements m ON wm.movement_id = m.id
		WHERE w.user_id = ?
		GROUP BY m.id, m.name
		ORDER BY m.name
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
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
			&pr.WorkoutID,
			&pr.WorkoutDate,
		)
		if err != nil {
			return nil, err
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
func (r *SQLiteWorkoutMovementRepository) GetMaxWeightForMovement(userID, movementID int64) (*float64, error) {
	query := `
		SELECT MAX(wm.weight)
		FROM workout_movements wm
		INNER JOIN workouts w ON wm.workout_id = w.id
		WHERE w.user_id = ? AND wm.movement_id = ? AND wm.weight IS NOT NULL
	`

	var maxWeight sql.NullFloat64
	err := r.db.QueryRow(query, userID, movementID).Scan(&maxWeight)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if !maxWeight.Valid {
		return nil, nil
	}

	return &maxWeight.Float64, nil
}

// GetPRMovements retrieves recent PR-flagged movements for a user
func (r *SQLiteWorkoutMovementRepository) GetPRMovements(userID int64, limit int) ([]*domain.WorkoutMovement, error) {
	query := `
		SELECT wm.id, wm.workout_id, wm.movement_id, wm.weight, wm.sets, wm.reps, wm.time, wm.distance,
		       wm.is_rx, wm.is_pr, wm.notes, wm.order_index, wm.created_at, wm.updated_at,
		       m.id, m.name, m.description, m.type, m.is_standard, m.created_by, m.created_at, m.updated_at
		FROM workout_movements wm
		INNER JOIN workouts w ON wm.workout_id = w.id
		INNER JOIN movements m ON wm.movement_id = m.id
		WHERE w.user_id = ? AND wm.is_pr = 1
		ORDER BY w.workout_date DESC, wm.created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workoutMovements []*domain.WorkoutMovement
	for rows.Next() {
		wm := &domain.WorkoutMovement{}
		movement := &domain.Movement{}
		var weight sql.NullFloat64
		var sets sql.NullInt64
		var reps sql.NullInt64
		var timeVal sql.NullInt64
		var distance sql.NullFloat64
		var wmNotes sql.NullString
		var movementDesc sql.NullString
		var movementCreatedBy sql.NullInt64

		err := rows.Scan(
			&wm.ID,
			&wm.WorkoutID,
			&wm.MovementID,
			&weight,
			&sets,
			&reps,
			&timeVal,
			&distance,
			&wm.IsRx,
			&wm.IsPR,
			&wmNotes,
			&wm.OrderIndex,
			&wm.CreatedAt,
			&wm.UpdatedAt,
			&movement.ID,
			&movement.Name,
			&movementDesc,
			&movement.Type,
			&movement.IsStandard,
			&movementCreatedBy,
			&movement.CreatedAt,
			&movement.UpdatedAt,
		)
		if err != nil {
			return nil, err
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
		if timeVal.Valid {
			t := int(timeVal.Int64)
			wm.Time = &t
		}
		if distance.Valid {
			wm.Distance = &distance.Float64
		}
		if wmNotes.Valid {
			wm.Notes = wmNotes.String
		}
		if movementDesc.Valid {
			movement.Description = movementDesc.String
		}
		if movementCreatedBy.Valid {
			movement.CreatedBy = &movementCreatedBy.Int64
		}

		wm.Movement = movement
		workoutMovements = append(workoutMovements, wm)
	}

	return workoutMovements, rows.Err()
}
