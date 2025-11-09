package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// SQLiteWorkoutRepository implements WorkoutRepository using SQLite
type SQLiteWorkoutRepository struct {
	db *sql.DB
}

// NewSQLiteWorkoutRepository creates a new SQLite workout repository
func NewSQLiteWorkoutRepository(db *sql.DB) *SQLiteWorkoutRepository {
	return &SQLiteWorkoutRepository{db: db}
}

// Create creates a new workout
func (r *SQLiteWorkoutRepository) Create(workout *domain.Workout) error {
	query := `
		INSERT INTO workouts (user_id, workout_date, workout_type, workout_name, notes, total_time, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	workout.CreatedAt = now
	workout.UpdatedAt = now

	result, err := r.db.Exec(
		query,
		workout.UserID,
		workout.WorkoutDate,
		workout.WorkoutType,
		workout.WorkoutName,
		workout.Notes,
		workout.TotalTime,
		workout.CreatedAt,
		workout.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	workout.ID = id
	return nil
}

// GetByID retrieves a workout by ID
func (r *SQLiteWorkoutRepository) GetByID(id int64) (*domain.Workout, error) {
	query := `
		SELECT id, user_id, workout_date, workout_type, workout_name, notes, total_time,
		       created_at, updated_at
		FROM workouts
		WHERE id = ?
	`

	workout := &domain.Workout{}
	var workoutName sql.NullString
	var notes sql.NullString
	var totalTime sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&workout.ID,
		&workout.UserID,
		&workout.WorkoutDate,
		&workout.WorkoutType,
		&workoutName,
		&notes,
		&totalTime,
		&workout.CreatedAt,
		&workout.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if workoutName.Valid {
		workout.WorkoutName = workoutName.String
	}
	if notes.Valid {
		workout.Notes = notes.String
	}
	if totalTime.Valid {
		t := int(totalTime.Int64)
		workout.TotalTime = &t
	}

	return workout, nil
}

// GetByUserID retrieves workouts for a specific user with pagination
// GetByUserID retrieves workouts for a user with pagination
func (r *SQLiteWorkoutRepository) GetByUserID(userID int64, limit, offset int) ([]*domain.Workout, error) {
	query := `
		SELECT id, user_id, workout_date, workout_type, workout_name, notes, total_time,
		       created_at, updated_at
		FROM workouts
		WHERE user_id = ?
		ORDER BY workout_date DESC, created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workouts []*domain.Workout
	for rows.Next() {
		workout := &domain.Workout{}
		var workoutName sql.NullString
		var notes sql.NullString
		var totalTime sql.NullInt64

		err := rows.Scan(
			&workout.ID,
			&workout.UserID,
			&workout.WorkoutDate,
			&workout.WorkoutType,
			&workoutName,
			&notes,
			&totalTime,
			&workout.CreatedAt,
			&workout.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if workoutName.Valid {
			workout.WorkoutName = workoutName.String
		}
		if notes.Valid {
			workout.Notes = notes.String
		}
		if totalTime.Valid {
			t := int(totalTime.Int64)
			workout.TotalTime = &t
		}

		workouts = append(workouts, workout)
	}

	return workouts, rows.Err()
}

// GetByUserIDAndDateRange retrieves workouts for a user within a date range
func (r *SQLiteWorkoutRepository) GetByUserIDAndDateRange(userID int64, startDate, endDate time.Time) ([]*domain.Workout, error) {
	query := `
		SELECT id, user_id, workout_date, workout_type, workout_name, notes, total_time,
		       created_at, updated_at
		FROM workouts
		WHERE user_id = ? AND workout_date >= ? AND workout_date <= ?
		ORDER BY workout_date DESC, created_at DESC
	`

	rows, err := r.db.Query(query, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workouts []*domain.Workout
	for rows.Next() {
		workout := &domain.Workout{}
		var workoutName sql.NullString
		var notes sql.NullString
		var totalTime sql.NullInt64

		err := rows.Scan(
			&workout.ID,
			&workout.UserID,
			&workout.WorkoutDate,
			&workout.WorkoutType,
			&workoutName,
			&notes,
			&totalTime,
			&workout.CreatedAt,
			&workout.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if workoutName.Valid {
			workout.WorkoutName = workoutName.String
		}
		if notes.Valid {
			workout.Notes = notes.String
		}
		if totalTime.Valid {
			t := int(totalTime.Int64)
			workout.TotalTime = &t
		}

		workouts = append(workouts, workout)
	}

	return workouts, rows.Err()
}

// Update updates a workout
func (r *SQLiteWorkoutRepository) Update(workout *domain.Workout) error {
	query := `
		UPDATE workouts
		SET workout_date = ?, workout_type = ?, workout_name = ?, notes = ?,
		    total_time = ?, updated_at = ?
		WHERE id = ?
	`

	workout.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		workout.WorkoutDate,
		workout.WorkoutType,
		workout.WorkoutName,
		workout.Notes,
		workout.TotalTime,
		workout.UpdatedAt,
		workout.ID,
	)

	return err
}

// Delete deletes a workout
func (r *SQLiteWorkoutRepository) Delete(id int64) error {
	query := `DELETE FROM workouts WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// Count returns the total number of workouts for a user
func (r *SQLiteWorkoutRepository) Count(userID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM workouts WHERE user_id = ?`
	var count int64
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}
