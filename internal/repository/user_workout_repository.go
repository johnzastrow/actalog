package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

type UserWorkoutRepository struct {
	db *sql.DB
}

func NewUserWorkoutRepository(db *sql.DB) *UserWorkoutRepository {
	return &UserWorkoutRepository{db: db}
}

// Create creates a new user workout (logs a workout instance)
func (r *UserWorkoutRepository) Create(userWorkout *domain.UserWorkout) error {
	userWorkout.CreatedAt = time.Now()
	userWorkout.UpdatedAt = time.Now()

	query := `INSERT INTO user_workouts (user_id, workout_id, workout_name, workout_date, workout_type, total_time, notes, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, userWorkout.UserID, userWorkout.WorkoutID, userWorkout.WorkoutName, userWorkout.WorkoutDate, userWorkout.WorkoutType, userWorkout.TotalTime, userWorkout.Notes, userWorkout.CreatedAt, userWorkout.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user workout: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get user workout ID: %w", err)
	}

	userWorkout.ID = id
	return nil
}

// GetByID retrieves a user workout by ID
func (r *UserWorkoutRepository) GetByID(id int64) (*domain.UserWorkout, error) {
	query := `SELECT id, user_id, workout_id, workout_name, workout_date, workout_type, total_time, notes, created_at, updated_at FROM user_workouts WHERE id = ?`

	userWorkout := &domain.UserWorkout{}
	var workoutID sql.NullInt64
	var workoutName sql.NullString
	var workoutType sql.NullString
	var totalTime sql.NullInt64
	var notes sql.NullString

	err := r.db.QueryRow(query, id).Scan(&userWorkout.ID, &userWorkout.UserID, &workoutID, &workoutName, &userWorkout.WorkoutDate, &workoutType, &totalTime, &notes, &userWorkout.CreatedAt, &userWorkout.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user workout: %w", err)
	}

	if workoutID.Valid {
		wid := workoutID.Int64
		userWorkout.WorkoutID = &wid
	}
	if workoutName.Valid {
		userWorkout.WorkoutName = &workoutName.String
	}
	if workoutType.Valid {
		userWorkout.WorkoutType = &workoutType.String
	}
	if totalTime.Valid {
		t := int(totalTime.Int64)
		userWorkout.TotalTime = &t
	}
	if notes.Valid {
		userWorkout.Notes = &notes.String
	}

	return userWorkout, nil
}

// GetByIDWithDetails retrieves a user workout with full details (movements, WODs)
func (r *UserWorkoutRepository) GetByIDWithDetails(id int64, userID int64) (*domain.UserWorkoutWithDetails, error) {
	// First get the user workout
	userWorkout, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	if userWorkout == nil {
		return nil, nil
	}

	// Verify the user owns this workout
	if userWorkout.UserID != userID {
		return nil, fmt.Errorf("unauthorized: workout does not belong to user")
	}

	// Get workout name and description
	var workoutName string
	var workoutDescription *string

	if userWorkout.WorkoutID != nil {
		// Template-based workout - get name from template
		var workoutNotes sql.NullString
		query := `SELECT name, notes FROM workouts WHERE id = ?`
		if err := r.db.QueryRow(query, *userWorkout.WorkoutID).Scan(&workoutName, &workoutNotes); err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("workout template not found")
			}
			return nil, fmt.Errorf("failed to get workout template: %w", err)
		}
		if workoutNotes.Valid {
			workoutDescription = &workoutNotes.String
		}
	} else if userWorkout.WorkoutName != nil {
		// Ad-hoc workout - use workout_name from user_workouts table
		workoutName = *userWorkout.WorkoutName
	} else {
		return nil, fmt.Errorf("workout has neither workout_id nor workout_name")
	}

	// Get movements from workout_movements table with movement details (only for template-based workouts)
	var movements []*domain.WorkoutMovement
	if userWorkout.WorkoutID != nil {
		movementsQuery := `
			SELECT ws.id, ws.workout_id, ws.movement_id, ws.weight, ws.sets, ws.reps, ws.time, ws.distance,
				   ws.is_rx, ws.is_pr, ws.notes, ws.order_index, ws.created_at, ws.updated_at,
				   m.name as movement_name, m.type as movement_type
			FROM workout_movements ws
			JOIN movements m ON ws.movement_id = m.id
			WHERE ws.workout_id = ?
			ORDER BY ws.order_index`

		rows, err := r.db.Query(movementsQuery, *userWorkout.WorkoutID)
		if err != nil {
			return nil, fmt.Errorf("failed to get workout movements: %w", err)
		}
		defer rows.Close()

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

			movements = append(movements, wm)
		}
		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("failed to get workout movements: %w", err)
		}
	}

	// Get WODs from workout_wods table with WOD details (only for template-based workouts)
	var wods []*domain.WorkoutWODWithDetails
	if userWorkout.WorkoutID != nil {
		wodsQuery := `
			SELECT ww.id, ww.workout_id, ww.wod_id,
				   ww.order_index, ww.created_at, ww.updated_at,
				   w.name as wod_name, w.type as wod_type, w.regime as wod_regime,
				   w.score_type as wod_score_type, w.description as wod_description
			FROM workout_wods ww
			JOIN wods w ON ww.wod_id = w.id
			WHERE ww.workout_id = ?
			ORDER BY ww.order_index`

		rows, err := r.db.Query(wodsQuery, *userWorkout.WorkoutID)
		if err != nil {
			return nil, fmt.Errorf("failed to get workout WODs: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			wod := &domain.WorkoutWODWithDetails{}

			err := rows.Scan(&wod.ID, &wod.WorkoutID, &wod.WODID,
				&wod.OrderIndex, &wod.CreatedAt, &wod.UpdatedAt,
				&wod.WODName, &wod.WODType, &wod.WODRegime, &wod.WODScoreType, &wod.WODDescription)
			if err != nil {
				return nil, fmt.Errorf("failed to scan workout WOD: %w", err)
			}

			wods = append(wods, wod)
		}
		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("failed to get workout WODs: %w", err)
		}
	}

	// Get actual performance movements from user_workout_movements table
	perfMovementsQuery := `
		SELECT uwm.id, uwm.user_workout_id, uwm.movement_id, uwm.sets, uwm.reps, uwm.weight,
		       uwm.time, uwm.distance, uwm.notes, uwm.order_index, uwm.created_at, uwm.updated_at,
		       m.name as movement_name, m.type as movement_type
		FROM user_workout_movements uwm
		JOIN movements m ON uwm.movement_id = m.id
		WHERE uwm.user_workout_id = ?
		ORDER BY uwm.order_index`

	perfMovRows, err := r.db.Query(perfMovementsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workout movements: %w", err)
	}
	defer perfMovRows.Close()

	var performanceMovements []*domain.UserWorkoutMovement
	for perfMovRows.Next() {
		uwm := &domain.UserWorkoutMovement{}
		var sets sql.NullInt64
		var reps sql.NullInt64
		var weight sql.NullFloat64
		var time sql.NullInt64
		var distance sql.NullFloat64
		var notes sql.NullString
		var movementName string
		var movementType string

		err := perfMovRows.Scan(&uwm.ID, &uwm.UserWorkoutID, &uwm.MovementID, &sets, &reps, &weight,
			&time, &distance, &notes, &uwm.OrderIndex, &uwm.CreatedAt, &uwm.UpdatedAt,
			&movementName, &movementType)
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
		if notes.Valid {
			uwm.Notes = notes.String
		}

		uwm.Movement = &domain.Movement{
			ID:   uwm.MovementID,
			Name: movementName,
			Type: domain.MovementType(movementType),
		}
		uwm.MovementName = movementName
		uwm.MovementType = movementType

		performanceMovements = append(performanceMovements, uwm)
	}
	if err = perfMovRows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get user workout movements: %w", err)
	}

	// Get actual performance WODs from user_workout_wods table
	perfWODsQuery := `
		SELECT uww.id, uww.user_workout_id, uww.wod_id, uww.score_type, uww.score_value,
		       uww.time_seconds, uww.rounds, uww.reps, uww.weight, uww.notes,
		       uww.order_index, uww.created_at, uww.updated_at,
		       w.name as wod_name, w.type as wod_type, w.regime as wod_regime
		FROM user_workout_wods uww
		JOIN wods w ON uww.wod_id = w.id
		WHERE uww.user_workout_id = ?
		ORDER BY uww.order_index`

	perfWODRows, err := r.db.Query(perfWODsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workout WODs: %w", err)
	}
	defer perfWODRows.Close()

	var performanceWODs []*domain.UserWorkoutWOD
	for perfWODRows.Next() {
		uww := &domain.UserWorkoutWOD{}
		var scoreType sql.NullString
		var scoreValue sql.NullString
		var timeSeconds sql.NullInt64
		var rounds sql.NullInt64
		var reps sql.NullInt64
		var weight sql.NullFloat64
		var notes sql.NullString
		var wodName string
		var wodType string
		var wodRegime string

		err := perfWODRows.Scan(&uww.ID, &uww.UserWorkoutID, &uww.WODID, &scoreType, &scoreValue,
			&timeSeconds, &rounds, &reps, &weight, &notes,
			&uww.OrderIndex, &uww.CreatedAt, &uww.UpdatedAt,
			&wodName, &wodType, &wodRegime)
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
		if notes.Valid {
			uww.Notes = notes.String
		}

		uww.WOD = &domain.WOD{
			ID:     uww.WODID,
			Name:   wodName,
			Type:   wodType,
			Regime: wodRegime,
		}
		uww.WODName = wodName

		performanceWODs = append(performanceWODs, uww)
	}
	if err = perfWODRows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get user workout WODs: %w", err)
	}

	// Construct the detailed response
	result := &domain.UserWorkoutWithDetails{
		UserWorkout:          *userWorkout,
		WorkoutName:          workoutName,
		WorkoutDescription:   workoutDescription,
		Movements:            movements,
		WODs:                 wods,
		PerformanceMovements: performanceMovements,
		PerformanceWODs:      performanceWODs,
	}

	return result, nil
}

// ListByUser retrieves all workouts logged by a specific user
func (r *UserWorkoutRepository) ListByUser(userID int64, limit, offset int) ([]*domain.UserWorkout, error) {
	query := `SELECT id, user_id, workout_id, workout_date, workout_type, total_time, notes, created_at, updated_at FROM user_workouts WHERE user_id = ? ORDER BY workout_date DESC, created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user workouts: %w", err)
	}
	defer rows.Close()

	return r.scanUserWorkouts(rows)
}

// ListByUserWithDetails retrieves all workouts logged by a user with details
func (r *UserWorkoutRepository) ListByUserWithDetails(userID int64, limit, offset int) ([]*domain.UserWorkoutWithDetails, error) {
	// Get user workouts
	userWorkouts, err := r.ListByUser(userID, limit, offset)
	if err != nil {
		return nil, err
	}

	var results []*domain.UserWorkoutWithDetails
	for _, uw := range userWorkouts {
		// Get details for each workout
		details, err := r.GetByIDWithDetails(uw.ID, userID)
		if err != nil {
			return nil, err
		}
		results = append(results, details)
	}

	return results, nil
}

// ListByUserAndDateRange retrieves workouts within a date range
func (r *UserWorkoutRepository) ListByUserAndDateRange(userID int64, startDate, endDate time.Time) ([]*domain.UserWorkout, error) {
	query := `SELECT id, user_id, workout_id, workout_date, workout_type, total_time, notes, created_at, updated_at FROM user_workouts WHERE user_id = ? AND workout_date >= ? AND workout_date <= ? ORDER BY workout_date DESC`

	rows, err := r.db.Query(query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to list user workouts by date range: %w", err)
	}
	defer rows.Close()

	return r.scanUserWorkouts(rows)
}

// Update updates an existing user workout
func (r *UserWorkoutRepository) Update(userWorkout *domain.UserWorkout) error {
	userWorkout.UpdatedAt = time.Now()

	query := `UPDATE user_workouts
	          SET workout_name = ?, workout_date = ?, workout_type = ?, total_time = ?,
	              notes = ?, updated_at = ?
	          WHERE id = ? AND user_id = ?`

	result, err := r.db.Exec(query, userWorkout.WorkoutName, userWorkout.WorkoutDate, userWorkout.WorkoutType, userWorkout.TotalTime, userWorkout.Notes, userWorkout.UpdatedAt, userWorkout.ID, userWorkout.UserID)
	if err != nil {
		return fmt.Errorf("failed to update user workout: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user workout not found or unauthorized")
	}

	return nil
}

// Delete deletes a user workout
func (r *UserWorkoutRepository) Delete(id int64, userID int64) error {
	query := `DELETE FROM user_workouts WHERE id = ? AND user_id = ?`

	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user workout: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user workout not found or unauthorized")
	}

	return nil
}

// GetByUserWorkoutDate checks if a user has already logged a specific workout on a date
func (r *UserWorkoutRepository) GetByUserWorkoutDate(userID, workoutID int64, date time.Time) (*domain.UserWorkout, error) {
	query := `SELECT id, user_id, workout_id, workout_date, workout_type, total_time, notes, created_at, updated_at FROM user_workouts WHERE user_id = ? AND workout_id = ? AND DATE(workout_date) = DATE(?)`

	userWorkout := &domain.UserWorkout{}
	var workoutType sql.NullString
	var totalTime sql.NullInt64
	var notes sql.NullString

	err := r.db.QueryRow(query, userID, workoutID, date).Scan(&userWorkout.ID, &userWorkout.UserID, &userWorkout.WorkoutID, &userWorkout.WorkoutDate, &workoutType, &totalTime, &notes, &userWorkout.CreatedAt, &userWorkout.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user workout by date: %w", err)
	}

	if workoutType.Valid {
		userWorkout.WorkoutType = &workoutType.String
	}
	if totalTime.Valid {
		t := int(totalTime.Int64)
		userWorkout.TotalTime = &t
	}
	if notes.Valid {
		userWorkout.Notes = &notes.String
	}

	return userWorkout, nil
}

// Count counts total user workouts for a specific user
func (r *UserWorkoutRepository) Count(userID int64) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM user_workouts WHERE user_id = ?`

	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user workouts: %w", err)
	}

	return count, nil
}

// GetRecentForUser retrieves recent user workouts with details (for dashboard/activity feed)
func (r *UserWorkoutRepository) GetRecentForUser(userID int64, limit int) ([]*domain.UserWorkoutWithDetails, error) {
	query := `SELECT uw.id, uw.user_id, uw.workout_id, uw.workout_date, uw.workout_type, uw.total_time,
	                 uw.notes, uw.created_at, uw.updated_at,
	                 w.name as workout_name, w.notes as workout_description
	          FROM user_workouts uw
	          JOIN workouts w ON uw.workout_id = w.id
	          WHERE uw.user_id = ?
	          ORDER BY uw.workout_date DESC, uw.created_at DESC
	          LIMIT ?`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent user workouts: %w", err)
	}
	defer rows.Close()

	var workouts []*domain.UserWorkoutWithDetails
	for rows.Next() {
		details := &domain.UserWorkoutWithDetails{}
		var workoutType sql.NullString
		var totalTime sql.NullInt64
		var notes sql.NullString
		var workoutDescription sql.NullString

		err := rows.Scan(
			&details.ID,
			&details.UserID,
			&details.WorkoutID,
			&details.WorkoutDate,
			&workoutType,
			&totalTime,
			&notes,
			&details.CreatedAt,
			&details.UpdatedAt,
			&details.WorkoutName,
			&workoutDescription)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recent user workout: %w", err)
		}

		if workoutType.Valid {
			details.WorkoutType = &workoutType.String
		}
		if totalTime.Valid {
			t := int(totalTime.Int64)
			details.TotalTime = &t
		}
		if notes.Valid {
			details.Notes = &notes.String
		}
		if workoutDescription.Valid {
			details.WorkoutDescription = &workoutDescription.String
		}

		workouts = append(workouts, details)
	}

	return workouts, rows.Err()
}

// scanUserWorkouts scans multiple user workout rows
func (r *UserWorkoutRepository) scanUserWorkouts(rows *sql.Rows) ([]*domain.UserWorkout, error) {
	var userWorkouts []*domain.UserWorkout
	for rows.Next() {
		userWorkout := &domain.UserWorkout{}
		var workoutType sql.NullString
		var totalTime sql.NullInt64
		var notes sql.NullString

		err := rows.Scan(&userWorkout.ID, &userWorkout.UserID, &userWorkout.WorkoutID, &userWorkout.WorkoutDate, &workoutType, &totalTime, &notes, &userWorkout.CreatedAt, &userWorkout.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if workoutType.Valid {
			userWorkout.WorkoutType = &workoutType.String
		}
		if totalTime.Valid {
			t := int(totalTime.Int64)
			userWorkout.TotalTime = &t
		}
		if notes.Valid {
			userWorkout.Notes = &notes.String
		}

		userWorkouts = append(userWorkouts, userWorkout)
	}

	return userWorkouts, rows.Err()
}
