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

	query := `INSERT INTO user_workout_wods (user_workout_id, wod_id, score_type, score_value, time_seconds, rounds, reps, weight, notes, is_pr, order_index, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, uww.UserWorkoutID, uww.WODID, uww.ScoreType, uww.ScoreValue, uww.TimeSeconds, uww.Rounds, uww.Reps, uww.Weight, uww.Notes, uww.IsPR, uww.OrderIndex, uww.CreatedAt, uww.UpdatedAt)
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

	query := `INSERT INTO user_workout_wods (user_workout_id, wod_id, score_type, score_value, time_seconds, rounds, reps, weight, notes, is_pr, order_index, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, uww := range wods {
		uww.CreatedAt = now
		uww.UpdatedAt = now

		result, err := stmt.Exec(uww.UserWorkoutID, uww.WODID, uww.ScoreType, uww.ScoreValue, uww.TimeSeconds, uww.Rounds, uww.Reps, uww.Weight, uww.Notes, uww.IsPR, uww.OrderIndex, uww.CreatedAt, uww.UpdatedAt)
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

// GetBestTimeForWOD retrieves the fastest time for a specific WOD for a user
func (r *UserWorkoutWODRepository) GetBestTimeForWOD(userID, wodID int64) (*int, error) {
	query := `
		SELECT MIN(uww.time_seconds)
		FROM user_workout_wods uww
		INNER JOIN user_workouts uw ON uww.user_workout_id = uw.id
		WHERE uw.user_id = ? AND uww.wod_id = ? AND uww.time_seconds IS NOT NULL`

	var bestTime sql.NullInt64
	err := r.db.QueryRow(query, userID, wodID).Scan(&bestTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get best time: %w", err)
	}

	if !bestTime.Valid {
		return nil, nil
	}

	t := int(bestTime.Int64)
	return &t, nil
}

// GetBestRoundsRepsForWOD retrieves the best rounds+reps for a specific WOD for a user
// Returns the most rounds, and if tied, the most reps
func (r *UserWorkoutWODRepository) GetBestRoundsRepsForWOD(userID, wodID int64) (rounds *int, reps *int, err error) {
	query := `
		SELECT uww.rounds, uww.reps
		FROM user_workout_wods uww
		INNER JOIN user_workouts uw ON uww.user_workout_id = uw.id
		WHERE uw.user_id = ? AND uww.wod_id = ? AND uww.rounds IS NOT NULL
		ORDER BY uww.rounds DESC, uww.reps DESC
		LIMIT 1`

	var roundsVal sql.NullInt64
	var repsVal sql.NullInt64
	err = r.db.QueryRow(query, userID, wodID).Scan(&roundsVal, &repsVal)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to get best rounds+reps: %w", err)
	}

	if roundsVal.Valid {
		r := int(roundsVal.Int64)
		rounds = &r
	}
	if repsVal.Valid {
		rp := int(repsVal.Int64)
		reps = &rp
	}

	return rounds, reps, nil
}

// GetPRWODs retrieves recent PR-flagged WODs for a user
func (r *UserWorkoutWODRepository) GetPRWODs(userID int64, limit int) ([]*domain.UserWorkoutWOD, error) {
	query := `
		SELECT uww.id, uww.user_workout_id, uww.wod_id, uww.score_type, uww.score_value, uww.time_seconds, uww.rounds, uww.reps, uww.weight,
		       uww.notes, uww.is_pr, uww.order_index, uww.created_at, uww.updated_at,
		       w.name,
		       uw.workout_date
		FROM user_workout_wods uww
		JOIN wods w ON uww.wod_id = w.id
		JOIN user_workouts uw ON uww.user_workout_id = uw.id
		WHERE uw.user_id = ? AND uww.is_pr = 1
		ORDER BY uw.workout_date DESC, uww.created_at DESC
		LIMIT ?`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR WODs: %w", err)
	}
	defer rows.Close()

	var wods []*domain.UserWorkoutWOD
	for rows.Next() {
		uww := &domain.UserWorkoutWOD{}
		var scoreType sql.NullString
		var scoreValue sql.NullString
		var timeSeconds sql.NullInt64
		var rounds sql.NullInt64
		var reps sql.NullInt64
		var weight sql.NullFloat64
		var workoutDate time.Time

		err := rows.Scan(&uww.ID, &uww.UserWorkoutID, &uww.WODID, &scoreType, &scoreValue, &timeSeconds, &rounds, &reps, &weight,
			&uww.Notes, &uww.IsPR, &uww.OrderIndex, &uww.CreatedAt, &uww.UpdatedAt,
			&uww.WODName, &workoutDate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan PR WOD: %w", err)
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

		wods = append(wods, uww)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate PR WODs: %w", err)
	}

	return wods, nil
}

// UpdatePRFlag updates the is_pr flag for a user workout WOD
func (r *UserWorkoutWODRepository) UpdatePRFlag(id int64, isPR bool) error {
	query := `UPDATE user_workout_wods SET is_pr = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	result, err := r.db.Exec(query, isPR, id)
	if err != nil {
		return fmt.Errorf("failed to update PR flag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user workout WOD found with id %d", id)
	}

	return nil
}

// GetByUserIDAndWODID retrieves all WOD performance records for a specific user and WOD
func (r *UserWorkoutWODRepository) GetByUserIDAndWODID(userID, wodID int64, limit int) ([]*domain.UserWorkoutWOD, error) {
	query := `
		SELECT uww.id, uww.user_workout_id, uww.wod_id, uww.score_type, uww.score_value,
		       uww.time_seconds, uww.rounds, uww.reps, uww.weight, uww.notes, uww.is_pr,
		       uww.order_index, uww.created_at, uww.updated_at,
		       w.name, w.type, w.score_type,
		       uw.workout_date
		FROM user_workout_wods uww
		JOIN wods w ON uww.wod_id = w.id
		JOIN user_workouts uw ON uww.user_workout_id = uw.id
		WHERE uw.user_id = ? AND uww.wod_id = ?
		ORDER BY uw.workout_date DESC, uww.created_at DESC
		LIMIT ?`

	rows, err := r.db.Query(query, userID, wodID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query WOD performances: %w", err)
	}
	defer rows.Close()

	var wods []*domain.UserWorkoutWOD
	for rows.Next() {
		uww := &domain.UserWorkoutWOD{}
		var scoreType sql.NullString
		var scoreValue sql.NullString
		var timeSeconds sql.NullInt64
		var rounds sql.NullInt64
		var reps sql.NullInt64
		var weight sql.NullFloat64
		var workoutDate time.Time

		err := rows.Scan(&uww.ID, &uww.UserWorkoutID, &uww.WODID, &scoreType, &scoreValue,
			&timeSeconds, &rounds, &reps, &weight, &uww.Notes, &uww.IsPR,
			&uww.OrderIndex, &uww.CreatedAt, &uww.UpdatedAt,
			&uww.WODName, &uww.WODType, &uww.WODScoreType, &workoutDate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan WOD performance: %w", err)
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

		// Store workout date from user_workouts table
		uww.WorkoutDate = workoutDate

		wods = append(wods, uww)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate WOD performances: %w", err)
	}

	return wods, nil
}
