package repository

import (
	"database/sql"
	"errors"

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
		INSERT INTO workout_movements (workout_id, movement_id, weight, sets, reps, time, distance, is_rx, notes, order_index, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

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
		SELECT id, workout_id, movement_id, weight, sets, reps, time, distance, is_rx, notes, order_index,
		       created_at, updated_at
		FROM workout_movements
		WHERE id = ?
	`

	wm := &domain.WorkoutMovement{}
	var weight, distance sql.NullFloat64
	var sets, reps, time sql.NullInt64
	var notes sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&wm.ID,
		&wm.WorkoutID,
		&wm.MovementID,
		&weight,
		&sets,
		&reps,
		&time,
		&distance,
		&wm.IsRx,
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

// GetByWorkoutID retrieves all movements for a workout
func (r *SQLiteWorkoutMovementRepository) GetByWorkoutID(workoutID int64) ([]*domain.WorkoutMovement, error) {
	query := `
		SELECT id, workout_id, movement_id, weight, sets, reps, time, distance, is_rx, notes, order_index,
		       created_at, updated_at
		FROM workout_movements
		WHERE workout_id = ?
		ORDER BY order_index
	`

	rows, err := r.db.Query(query, workoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []*domain.WorkoutMovement
	for rows.Next() {
		wm := &domain.WorkoutMovement{}
		var weight, distance sql.NullFloat64
		var sets, reps, time sql.NullInt64
		var notes sql.NullString

		err := rows.Scan(
			&wm.ID,
			&wm.WorkoutID,
			&wm.MovementID,
			&weight,
			&sets,
			&reps,
			&time,
			&distance,
			&wm.IsRx,
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

		movements = append(movements, wm)
	}

	return movements, rows.Err()
}

// GetByUserIDAndMovementID retrieves workout movements for a user and specific movement
func (r *SQLiteWorkoutMovementRepository) GetByUserIDAndMovementID(userID, movementID int64, limit int) ([]*domain.WorkoutMovement, error) {
	query := `
		SELECT wm.id, wm.workout_id, wm.movement_id, wm.weight, wm.sets, wm.reps, wm.time, wm.distance,
		       wm.is_rx, wm.notes, wm.order_index, wm.created_at, wm.updated_at
		FROM workout_movements wm
		INNER JOIN workouts w ON wm.workout_id = w.id
		WHERE w.user_id = ? AND wm.movement_id = ?
		ORDER BY w.workout_date DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, userID, movementID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []*domain.WorkoutMovement
	for rows.Next() {
		wm := &domain.WorkoutMovement{}
		var weight, distance sql.NullFloat64
		var sets, reps, time sql.NullInt64
		var notes sql.NullString

		err := rows.Scan(
			&wm.ID,
			&wm.WorkoutID,
			&wm.MovementID,
			&weight,
			&sets,
			&reps,
			&time,
			&distance,
			&wm.IsRx,
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

		movements = append(movements, wm)
	}

	return movements, rows.Err()
}

// Update updates a workout movement
func (r *SQLiteWorkoutMovementRepository) Update(wm *domain.WorkoutMovement) error {
	query := `
		UPDATE workout_movements
		SET movement_id = ?, weight = ?, sets = ?, reps = ?, time = ?, distance = ?,
		    is_rx = ?, notes = ?, order_index = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(
		query,
		wm.MovementID,
		wm.Weight,
		wm.Sets,
		wm.Reps,
		wm.Time,
		wm.Distance,
		wm.IsRx,
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

// DeleteByWorkoutID deletes all movements for a workout
func (r *SQLiteWorkoutMovementRepository) DeleteByWorkoutID(workoutID int64) error {
	query := `DELETE FROM workout_movements WHERE workout_id = ?`
	_, err := r.db.Exec(query, workoutID)
	return err
}
