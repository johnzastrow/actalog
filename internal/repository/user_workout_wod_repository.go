package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// UserWorkoutWODRepository implements domain.UserWorkoutWODRepository
type UserWorkoutWODRepository struct {
	db *sql.DB
}

// NewUserWorkoutWODRepository creates a new user workout WOD repository
func NewUserWorkoutWODRepository(db *sql.DB) *UserWorkoutWODRepository {
	return &UserWorkoutWODRepository{db: db}
}

// Create creates a new user workout WOD performance record
func (r *UserWorkoutWODRepository) Create(uww *domain.UserWorkoutWOD) error {
	uww.CreatedAt = time.Now()
	uww.UpdatedAt = time.Now()

	query := `INSERT INTO user_workout_wods (user_workout_id, wod_id, score_type, score_value, time_seconds, rounds, reps, weight, notes, order_index, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, uww.UserWorkoutID, uww.WODID, uww.ScoreType, uww.ScoreValue, uww.TimeSeconds, uww.Rounds, uww.Reps, uww.Weight, uww.Notes, uww.OrderIndex, uww.CreatedAt, uww.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user workout WOD: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get user workout WOD ID: %w", err)
	}

	uww.ID = id
	return nil
}

// CreateBatch creates multiple user workout WOD records at once
func (r *UserWorkoutWODRepository) CreateBatch(wods []*domain.UserWorkoutWOD) error {
	if len(wods) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO user_workout_wods (user_workout_id, wod_id, score_type, score_value, time_seconds, rounds, reps, weight, notes, order_index, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, uww := range wods {
		uww.CreatedAt = now
		uww.UpdatedAt = now

		result, err := stmt.Exec(uww.UserWorkoutID, uww.WODID, uww.ScoreType, uww.ScoreValue, uww.TimeSeconds, uww.Rounds, uww.Reps, uww.Weight, uww.Notes, uww.OrderIndex, uww.CreatedAt, uww.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert user workout WOD: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get user workout WOD ID: %w", err)
		}
		uww.ID = id
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a user workout WOD by ID
func (r *UserWorkoutWODRepository) GetByID(id int64) (*domain.UserWorkoutWOD, error) {
	query := `SELECT id, user_workout_id, wod_id, score_type, score_value, time_seconds, rounds, reps, weight, notes, order_index, created_at, updated_at
	          FROM user_workout_wods WHERE id = ?`

	uww := &domain.UserWorkoutWOD{}
	var scoreType sql.NullString
	var scoreValue sql.NullString
	var timeSeconds sql.NullInt64
	var rounds sql.NullInt64
	var reps sql.NullInt64
	var weight sql.NullFloat64

	err := r.db.QueryRow(query, id).Scan(&uww.ID, &uww.UserWorkoutID, &uww.WODID, &scoreType, &scoreValue, &timeSeconds, &rounds, &reps, &weight, &uww.Notes, &uww.OrderIndex, &uww.CreatedAt, &uww.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user workout WOD: %w", err)
	}

	if scoreType.Valid {
		uww.ScoreType = &scoreType.String
	}
	if scoreValue.Valid {
		uww.ScoreValue = &scoreValue.String
	}
	if timeSeconds.Valid {
		t := int(timeSeconds.Int64)
		uww.TimeSeconds = &t
	}
	if rounds.Valid {
		r := int(rounds.Int64)
		uww.Rounds = &r
	}
	if reps.Valid {
		r := int(reps.Int64)
		uww.Reps = &r
	}
	if weight.Valid {
		uww.Weight = &weight.Float64
	}

	return uww, nil
}

// GetByUserWorkoutID retrieves all WODs for a specific logged workout
func (r *UserWorkoutWODRepository) GetByUserWorkoutID(userWorkoutID int64) ([]*domain.UserWorkoutWOD, error) {
	query := `
		SELECT uww.id, uww.user_workout_id, uww.wod_id, uww.score_type, uww.score_value, uww.time_seconds, uww.rounds, uww.reps, uww.weight,
		       uww.notes, uww.order_index, uww.created_at, uww.updated_at,
		       w.id as wod_id, w.name, w.source, w.type, w.regime, w.score_type as wod_score_type, w.description, w.url, w.notes as wod_notes, w.is_standard, w.created_by, w.created_at, w.updated_at
		FROM user_workout_wods uww
		JOIN wods w ON uww.wod_id = w.id
		WHERE uww.user_workout_id = ?
		ORDER BY uww.order_index`

	rows, err := r.db.Query(query, userWorkoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workout WODs: %w", err)
	}
	defer rows.Close()

	var wods []*domain.UserWorkoutWOD
	for rows.Next() {
		uww := &domain.UserWorkoutWOD{
			WOD: &domain.WOD{},
		}
		var scoreType sql.NullString
		var scoreValue sql.NullString
		var timeSeconds sql.NullInt64
		var rounds sql.NullInt64
		var reps sql.NullInt64
		var weight sql.NullFloat64
		var wodURL sql.NullString
		var wodNotes sql.NullString
		var createdBy sql.NullInt64

		err := rows.Scan(&uww.ID, &uww.UserWorkoutID, &uww.WODID, &scoreType, &scoreValue, &timeSeconds, &rounds, &reps, &weight,
			&uww.Notes, &uww.OrderIndex, &uww.CreatedAt, &uww.UpdatedAt,
			&uww.WOD.ID, &uww.WOD.Name, &uww.WOD.Source, &uww.WOD.Type, &uww.WOD.Regime, &uww.WOD.ScoreType, &uww.WOD.Description, &wodURL, &wodNotes, &uww.WOD.IsStandard, &createdBy, &uww.WOD.CreatedAt, &uww.WOD.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user workout WOD: %w", err)
		}

		if scoreType.Valid {
			uww.ScoreType = &scoreType.String
		}
		if scoreValue.Valid {
			uww.ScoreValue = &scoreValue.String
		}
		if timeSeconds.Valid {
			t := int(timeSeconds.Int64)
			uww.TimeSeconds = &t
		}
		if rounds.Valid {
			r := int(rounds.Int64)
			uww.Rounds = &r
		}
		if reps.Valid {
			r := int(reps.Int64)
			uww.Reps = &r
		}
		if weight.Valid {
			uww.Weight = &weight.Float64
		}
		if wodURL.Valid {
			uww.WOD.URL = &wodURL.String
		}
		if wodNotes.Valid {
			uww.WOD.Notes = &wodNotes.String
		}
		if createdBy.Valid {
			cb := createdBy.Int64
			uww.WOD.CreatedBy = &cb
		}

		wods = append(wods, uww)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user workout WODs: %w", err)
	}

	return wods, nil
}

// Update updates an existing user workout WOD
func (r *UserWorkoutWODRepository) Update(uww *domain.UserWorkoutWOD) error {
	uww.UpdatedAt = time.Now()

	query := `UPDATE user_workout_wods
	          SET score_type = ?, score_value = ?, time_seconds = ?, rounds = ?, reps = ?, weight = ?, notes = ?, order_index = ?, updated_at = ?
	          WHERE id = ?`

	result, err := r.db.Exec(query, uww.ScoreType, uww.ScoreValue, uww.TimeSeconds, uww.Rounds, uww.Reps, uww.Weight, uww.Notes, uww.OrderIndex, uww.UpdatedAt, uww.ID)
	if err != nil {
		return fmt.Errorf("failed to update user workout WOD: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user workout WOD not found")
	}

	return nil
}

// Delete deletes a user workout WOD
func (r *UserWorkoutWODRepository) Delete(id int64) error {
	query := `DELETE FROM user_workout_wods WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user workout WOD: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user workout WOD not found")
	}

	return nil
}

// DeleteByUserWorkoutID deletes all WODs for a logged workout
func (r *UserWorkoutWODRepository) DeleteByUserWorkoutID(userWorkoutID int64) error {
	query := `DELETE FROM user_workout_wods WHERE user_workout_id = ?`

	_, err := r.db.Exec(query, userWorkoutID)
	if err != nil {
		return fmt.Errorf("failed to delete user workout WODs: %w", err)
	}

	return nil
}
