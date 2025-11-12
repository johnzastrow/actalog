package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

type WorkoutWODRepository struct {
	db *sql.DB
}

func NewWorkoutWODRepository(db *sql.DB) *WorkoutWODRepository {
	return &WorkoutWODRepository{db: db}
}

// Create creates a new workout-WOD association
func (r *WorkoutWODRepository) Create(workoutWOD *domain.WorkoutWOD) error {
	workoutWOD.CreatedAt = time.Now()
	workoutWOD.UpdatedAt = time.Now()

	query := `INSERT INTO workout_wods (workout_id, wod_id, score_value, division, is_pr, order_index, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, workoutWOD.WorkoutID, workoutWOD.WODID, workoutWOD.ScoreValue, workoutWOD.Division, workoutWOD.IsPR, workoutWOD.OrderIndex, workoutWOD.CreatedAt, workoutWOD.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create workout-WOD: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get workout-WOD ID: %w", err)
	}

	workoutWOD.ID = id
	return nil
}

// GetByID retrieves a workout-WOD by ID
func (r *WorkoutWODRepository) GetByID(id int64) (*domain.WorkoutWOD, error) {
	query := `SELECT id, workout_id, wod_id, score_value, division, is_pr, order_index, created_at, updated_at FROM workout_wods WHERE id = ?`

	workoutWOD := &domain.WorkoutWOD{}
	var scoreValue sql.NullString
	var division sql.NullString

	err := r.db.QueryRow(query, id).Scan(&workoutWOD.ID, &workoutWOD.WorkoutID, &workoutWOD.WODID, &scoreValue, &division, &workoutWOD.IsPR, &workoutWOD.OrderIndex, &workoutWOD.CreatedAt, &workoutWOD.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workout-WOD: %w", err)
	}

	if scoreValue.Valid {
		workoutWOD.ScoreValue = &scoreValue.String
	}
	if division.Valid {
		workoutWOD.Division = &division.String
	}

	return workoutWOD, nil
}

// ListByWorkout retrieves all WODs associated with a workout template
func (r *WorkoutWODRepository) ListByWorkout(workoutID int64) ([]*domain.WorkoutWOD, error) {
	query := `SELECT id, workout_id, wod_id, score_value, division, is_pr, order_index, created_at, updated_at FROM workout_wods WHERE workout_id = ? ORDER BY order_index`

	rows, err := r.db.Query(query, workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workout WODs: %w", err)
	}
	defer rows.Close()

	return r.scanWorkoutWODs(rows)
}

// ListByWorkoutWithDetails retrieves WODs with full WOD details
func (r *WorkoutWODRepository) ListByWorkoutWithDetails(workoutID int64) ([]*domain.WorkoutWODWithDetails, error) {
	query := `
		SELECT
			ww.id, ww.workout_id, ww.wod_id, ww.score_value, ww.division, ww.is_pr,
			ww.order_index, ww.created_at, ww.updated_at,
			w.name as wod_name, w.type as wod_type, w.regime as wod_regime,
			w.score_type as wod_score_type, w.description as wod_description
		FROM workout_wods ww
		JOIN wods w ON ww.wod_id = w.id
		WHERE ww.workout_id = ?
		ORDER BY ww.order_index`

	rows, err := r.db.Query(query, workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workout WODs with details: %w", err)
	}
	defer rows.Close()

	return r.scanWorkoutWODsWithDetails(rows)
}

// Update updates an existing workout-WOD association
func (r *WorkoutWODRepository) Update(workoutWOD *domain.WorkoutWOD) error {
	workoutWOD.UpdatedAt = time.Now()

	query := `UPDATE workout_wods
	          SET score_value = ?, division = ?, is_pr = ?,
	              order_index = ?, updated_at = ?
	          WHERE id = ?`

	result, err := r.db.Exec(query, workoutWOD.ScoreValue, workoutWOD.Division, workoutWOD.IsPR, workoutWOD.OrderIndex, workoutWOD.UpdatedAt, workoutWOD.ID)
	if err != nil {
		return fmt.Errorf("failed to update workout-WOD: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("workout-WOD not found")
	}

	return nil
}

// Delete deletes a workout-WOD association
func (r *WorkoutWODRepository) Delete(id int64) error {
	query := `DELETE FROM workout_wods WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete workout-WOD: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("workout-WOD not found")
	}

	return nil
}

// GetByWorkoutID retrieves all WODs for a specific workout (alias for ListByWorkout)
func (r *WorkoutWODRepository) GetByWorkoutID(workoutID int64) ([]*domain.WorkoutWOD, error) {
	return r.ListByWorkout(workoutID)
}

// GetByWODID finds which workouts use this WOD
func (r *WorkoutWODRepository) GetByWODID(wodID int64) ([]*domain.WorkoutWOD, error) {
	query := `SELECT id, workout_id, wod_id, score_value, division, is_pr, order_index, created_at, updated_at
	          FROM workout_wods
	          WHERE wod_id = ?
	          ORDER BY created_at DESC`

	rows, err := r.db.Query(query, wodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workouts by WOD ID: %w", err)
	}
	defer rows.Close()

	return r.scanWorkoutWODs(rows)
}

// DeleteByWorkout deletes all WOD associations for a workout
func (r *WorkoutWODRepository) DeleteByWorkout(workoutID int64) error {
	query := `DELETE FROM workout_wods WHERE workout_id = ?`

	if _, err := r.db.Exec(query, workoutID); err != nil {
		return fmt.Errorf("failed to delete workout WODs: %w", err)
	}

	return nil
}

// DeleteByWorkoutID deletes all WODs for a workout template (alias for DeleteByWorkout)
func (r *WorkoutWODRepository) DeleteByWorkoutID(workoutID int64) error {
	return r.DeleteByWorkout(workoutID)
}

// BatchCreate adds multiple WODs to a workout at once
func (r *WorkoutWODRepository) BatchCreate(workoutID int64, wodIDs []int64) error {
	if len(wodIDs) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO workout_wods (workout_id, wod_id, score_value, division, is_pr, order_index, created_at, updated_at)
	          VALUES (?, ?, NULL, NULL, 0, ?, ?, ?)`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for i, wodID := range wodIDs {
		_, err := stmt.Exec(workoutID, wodID, i, now, now)
		if err != nil {
			return fmt.Errorf("failed to create workout WOD for wod_id %d: %w", wodID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Reorder updates order_index for WODs in a workout
func (r *WorkoutWODRepository) Reorder(workoutID int64, wodIDs []int64) error {
	if len(wodIDs) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `UPDATE workout_wods SET order_index = ?, updated_at = ? WHERE workout_id = ? AND wod_id = ?`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for i, wodID := range wodIDs {
		result, err := stmt.Exec(i, now, workoutID, wodID)
		if err != nil {
			return fmt.Errorf("failed to update order for wod_id %d: %w", wodID, err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rows == 0 {
			return fmt.Errorf("workout WOD not found for workout_id %d, wod_id %d", workoutID, wodID)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TogglePR toggles the PR flag for a workout-WOD
func (r *WorkoutWODRepository) TogglePR(id int64) error {
	query := `UPDATE workout_wods SET is_pr = NOT is_pr, updated_at = ? WHERE id = ?`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to toggle PR: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("workout-WOD not found")
	}

	return nil
}

// scanWorkoutWODs scans multiple workout WOD rows
func (r *WorkoutWODRepository) scanWorkoutWODs(rows *sql.Rows) ([]*domain.WorkoutWOD, error) {
	var workoutWODs []*domain.WorkoutWOD
	for rows.Next() {
		workoutWOD := &domain.WorkoutWOD{}
		var scoreValue sql.NullString
		var division sql.NullString

		err := rows.Scan(&workoutWOD.ID, &workoutWOD.WorkoutID, &workoutWOD.WODID, &scoreValue, &division, &workoutWOD.IsPR, &workoutWOD.OrderIndex, &workoutWOD.CreatedAt, &workoutWOD.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workout WOD: %w", err)
		}

		if scoreValue.Valid {
			workoutWOD.ScoreValue = &scoreValue.String
		}
		if division.Valid {
			workoutWOD.Division = &division.String
		}

		workoutWODs = append(workoutWODs, workoutWOD)
	}

	return workoutWODs, rows.Err()
}

// scanWorkoutWODsWithDetails scans multiple workout WOD rows with WOD details
func (r *WorkoutWODRepository) scanWorkoutWODsWithDetails(rows *sql.Rows) ([]*domain.WorkoutWODWithDetails, error) {
	var workoutWODs []*domain.WorkoutWODWithDetails
	for rows.Next() {
		workoutWOD := &domain.WorkoutWODWithDetails{}
		var scoreValue sql.NullString
		var division sql.NullString

		err := rows.Scan(&workoutWOD.ID, &workoutWOD.WorkoutID, &workoutWOD.WODID, &scoreValue, &division, &workoutWOD.IsPR, &workoutWOD.OrderIndex, &workoutWOD.CreatedAt, &workoutWOD.UpdatedAt,
			&workoutWOD.WODName, &workoutWOD.WODType, &workoutWOD.WODRegime, &workoutWOD.WODScoreType, &workoutWOD.WODDescription)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workout WOD with details: %w", err)
		}

		if scoreValue.Valid {
			workoutWOD.ScoreValue = &scoreValue.String
		}
		if division.Valid {
			workoutWOD.Division = &division.String
		}

		workoutWODs = append(workoutWODs, workoutWOD)
	}

	return workoutWODs, rows.Err()
}
