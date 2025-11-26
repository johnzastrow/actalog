package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// WorkoutRepository implements domain.WorkoutRepository for workout templates
type WorkoutRepository struct {
	db *sql.DB
}

// NewWorkoutRepository creates a new workout repository
func NewWorkoutRepository(db *sql.DB) *WorkoutRepository {
	return &WorkoutRepository{db: db}
}

// Create creates a new workout template
func (r *WorkoutRepository) Create(workout *domain.Workout) error {
	workout.CreatedAt = time.Now()
	workout.UpdatedAt = time.Now()

	query := `INSERT INTO workouts (name, notes, created_by, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, workout.Name, workout.Notes, workout.CreatedBy, workout.CreatedAt, workout.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create workout: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get workout ID: %w", err)
	}

	workout.ID = id
	return nil
}

// GetByID retrieves a workout template by ID
func (r *WorkoutRepository) GetByID(id int64) (*domain.Workout, error) {
	query := `SELECT id, name, notes, created_by, created_at, updated_at FROM workouts WHERE id = ?`

	workout := &domain.Workout{}
	var createdBy sql.NullInt64
	var notes sql.NullString

	err := r.db.QueryRow(query, id).Scan(&workout.ID, &workout.Name, &notes, &createdBy, &workout.CreatedAt, &workout.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workout: %w", err)
	}

	if notes.Valid {
		workout.Notes = &notes.String
	}
	if createdBy.Valid {
		workout.CreatedBy = &createdBy.Int64
	}

	return workout, nil
}

// GetByIDWithDetails retrieves a workout with movements and WODs
func (r *WorkoutRepository) GetByIDWithDetails(id int64) (*domain.Workout, error) {
	// Get the workout template
	workout, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	if workout == nil {
		return nil, nil
	}

	// Get movements from workout_movements table with movement details
	movementsQuery := `
		SELECT ws.id, ws.workout_id, ws.movement_id, ws.weight, ws.sets, ws.reps, ws.time, ws.distance,
		       ws.is_rx, ws.is_pr, ws.notes, ws.order_index, ws.created_at, ws.updated_at,
		       m.name as movement_name, m.type as movement_type
		FROM workout_movements ws
		JOIN movements m ON ws.movement_id = m.id
		WHERE ws.workout_id = ?
		ORDER BY ws.order_index`

	rows, err := r.db.Query(movementsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout movements: %w", err)
	}
	defer rows.Close()

	var movements []*domain.WorkoutMovement
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

	workout.Movements = movements

	// Get WODs from workout_wods table with WOD details
	wodsQuery := `
		SELECT ww.id, ww.workout_id, ww.wod_id,
		       ww.order_index, ww.created_at, ww.updated_at,
		       w.name as wod_name, w.type as wod_type, w.regime as wod_regime,
		       w.score_type as wod_score_type, w.description as wod_description
		FROM workout_wods ww
		JOIN wods w ON ww.wod_id = w.id
		WHERE ww.workout_id = ?
		ORDER BY ww.order_index`

	rows, err = r.db.Query(wodsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout WODs: %w", err)
	}
	defer rows.Close()

	var wods []*domain.WorkoutWODWithDetails
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

	workout.WODs = wods

	return workout, nil
}

// List retrieves all workout templates with optional filtering
func (r *WorkoutRepository) List(filters map[string]interface{}, limit, offset int) ([]*domain.Workout, error) {
	query := `SELECT id, name, notes, created_by, created_at, updated_at FROM workouts WHERE 1=1`
	args := []interface{}{}

	// Apply filters if provided
	if name, ok := filters["name"].(string); ok && name != "" {
		query += ` AND name LIKE ?`
		args = append(args, "%"+name+"%")
	}

	if createdBy, ok := filters["created_by"].(int64); ok && createdBy > 0 {
		query += ` AND created_by = ?`
		args = append(args, createdBy)
	}

	query += ` ORDER BY name LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list workouts: %w", err)
	}
	defer rows.Close()

	return r.scanWorkouts(rows)
}

// ListByUser retrieves all workout templates created by a specific user
func (r *WorkoutRepository) ListByUser(userID int64, limit, offset int) ([]*domain.Workout, error) {
	query := `SELECT id, name, notes, created_by, created_at, updated_at
	          FROM workouts
	          WHERE created_by = ?
	          ORDER BY name
	          LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user workouts: %w", err)
	}
	defer rows.Close()

	return r.scanWorkouts(rows)
}

// ListStandard retrieves all standard (system) workout templates
func (r *WorkoutRepository) ListStandard(limit, offset int) ([]*domain.Workout, error) {
	query := `SELECT id, name, notes, created_by, created_at, updated_at
	          FROM workouts
	          WHERE created_by IS NULL
	          ORDER BY name
	          LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list standard workouts: %w", err)
	}
	defer rows.Close()

	return r.scanWorkouts(rows)
}

// Update updates an existing workout template
func (r *WorkoutRepository) Update(workout *domain.Workout) error {
	workout.UpdatedAt = time.Now()

	query := `UPDATE workouts
	          SET name = ?, notes = ?, updated_at = ?
	          WHERE id = ?`

	result, err := r.db.Exec(query, workout.Name, workout.Notes, workout.UpdatedAt, workout.ID)
	if err != nil {
		return fmt.Errorf("failed to update workout: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("workout not found")
	}

	return nil
}

// Delete deletes a workout template
func (r *WorkoutRepository) Delete(id int64) error {
	query := `DELETE FROM workouts WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete workout: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("workout not found")
	}

	return nil
}

// ListAllUserCreated retrieves all user-created workout templates (for admin view)
func (r *WorkoutRepository) ListAllUserCreated(limit, offset int) ([]*domain.Workout, error) {
	query := `SELECT id, name, notes, created_by, created_at, updated_at
	          FROM workouts
	          WHERE created_by IS NOT NULL
	          ORDER BY name
	          LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list all user-created workouts: %w", err)
	}
	defer rows.Close()

	return r.scanWorkouts(rows)
}

// ListAllUserCreatedWithUserInfo retrieves all user-created workout templates with creator info (for admin view)
func (r *WorkoutRepository) ListAllUserCreatedWithUserInfo(limit, offset int) ([]*domain.WorkoutWithCreator, error) {
	query := `SELECT w.id, w.name, w.notes, w.created_by, w.created_at, w.updated_at,
	                 COALESCE(u.email, '') as creator_email, COALESCE(u.name, '') as creator_name
	          FROM workouts w
	          LEFT JOIN users u ON w.created_by = u.id
	          WHERE w.created_by IS NOT NULL
	          ORDER BY w.name
	          LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list all user-created workouts with user info: %w", err)
	}
	defer rows.Close()

	var workouts []*domain.WorkoutWithCreator
	for rows.Next() {
		w := &domain.WorkoutWithCreator{}
		var createdBy sql.NullInt64
		var notes sql.NullString

		err := rows.Scan(
			&w.ID,
			&w.Name,
			&notes,
			&createdBy,
			&w.CreatedAt,
			&w.UpdatedAt,
			&w.CreatorEmail,
			&w.CreatorName,
		)
		if err != nil {
			return nil, err
		}

		if notes.Valid {
			w.Notes = &notes.String
		}
		if createdBy.Valid {
			w.CreatedBy = &createdBy.Int64
		}

		workouts = append(workouts, w)
	}

	return workouts, rows.Err()
}

// CountAllUserCreated counts all user-created workout templates
func (r *WorkoutRepository) CountAllUserCreated() (int64, error) {
	query := `SELECT COUNT(*) FROM workouts WHERE created_by IS NOT NULL`
	var count int64
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user-created workouts: %w", err)
	}
	return count, nil
}

// ListAllUserCreatedWithUserInfoFiltered retrieves all user-created workout templates with creator info and filters (for admin view)
func (r *WorkoutRepository) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, creator string) ([]*domain.WorkoutWithCreator, int64, error) {
	baseQuery := `SELECT w.id, w.name, w.notes, w.created_by, w.created_at, w.updated_at,
	                 COALESCE(u.email, '') as creator_email, COALESCE(u.name, '') as creator_name
	          FROM workouts w
	          LEFT JOIN users u ON w.created_by = u.id
	          WHERE w.created_by IS NOT NULL`

	countQuery := `SELECT COUNT(*) FROM workouts w LEFT JOIN users u ON w.created_by = u.id WHERE w.created_by IS NOT NULL`

	var args []interface{}
	var countArgs []interface{}

	// Apply filters
	if search != "" {
		baseQuery += " AND (w.name LIKE ? OR w.notes LIKE ?)"
		countQuery += " AND (w.name LIKE ? OR w.notes LIKE ?)"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
		countArgs = append(countArgs, searchTerm, searchTerm)
	}
	if creator != "" {
		baseQuery += " AND u.email LIKE ?"
		countQuery += " AND u.email LIKE ?"
		creatorTerm := "%" + creator + "%"
		args = append(args, creatorTerm)
		countArgs = append(countArgs, creatorTerm)
	}

	// Get count first
	var count int64
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count filtered workouts: %w", err)
	}

	// Add ordering and pagination
	baseQuery += " ORDER BY w.name"
	if limit > 0 {
		baseQuery += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		baseQuery += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list filtered user-created workouts: %w", err)
	}
	defer rows.Close()

	var workouts []*domain.WorkoutWithCreator
	for rows.Next() {
		w := &domain.WorkoutWithCreator{}
		var notes sql.NullString
		var createdBy sql.NullInt64

		err := rows.Scan(
			&w.ID,
			&w.Name,
			&notes,
			&createdBy,
			&w.CreatedAt,
			&w.UpdatedAt,
			&w.CreatorEmail,
			&w.CreatorName,
		)
		if err != nil {
			return nil, 0, err
		}

		if notes.Valid {
			w.Notes = &notes.String
		}
		if createdBy.Valid {
			w.CreatedBy = &createdBy.Int64
		}

		workouts = append(workouts, w)
	}

	return workouts, count, rows.Err()
}

// Search searches workout templates by name
func (r *WorkoutRepository) Search(query string, limit int) ([]*domain.Workout, error) {
	searchQuery := `SELECT id, name, notes, created_by, created_at, updated_at
	                FROM workouts
	                WHERE name LIKE ?
	                ORDER BY name
	                LIMIT ?`

	rows, err := r.db.Query(searchQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search workouts: %w", err)
	}
	defer rows.Close()

	return r.scanWorkouts(rows)
}

// Count counts total workout templates (optionally filtered by user)
func (r *WorkoutRepository) Count(userID *int64) (int64, error) {
	var count int64
	var query string

	if userID != nil {
		query = `SELECT COUNT(*) FROM workouts WHERE created_by = ?`
		err := r.db.QueryRow(query, *userID).Scan(&count)
		if err != nil {
			return 0, fmt.Errorf("failed to count workouts: %w", err)
		}
	} else {
		query = `SELECT COUNT(*) FROM workouts`
		err := r.db.QueryRow(query).Scan(&count)
		if err != nil {
			return 0, fmt.Errorf("failed to count workouts: %w", err)
		}
	}

	return count, nil
}

// GetUsageStats gets usage statistics for a template
func (r *WorkoutRepository) GetUsageStats(workoutID int64) (*domain.WorkoutWithUsageStats, error) {
	// Get the workout template
	workout, err := r.GetByID(workoutID)
	if err != nil {
		return nil, err
	}
	if workout == nil {
		return nil, nil
	}

	// Count how many times this template has been used
	var timesUsed int64
	countQuery := `SELECT COUNT(*) FROM user_workouts WHERE workout_id = ?`
	if err := r.db.QueryRow(countQuery, workoutID).Scan(&timesUsed); err != nil {
		return nil, fmt.Errorf("failed to count usage: %w", err)
	}

	// Get the most recent usage date
	var lastUsedAt *time.Time
	lastUsedQuery := `SELECT MAX(workout_date) FROM user_workouts WHERE workout_id = ?`
	var nullableLastUsed sql.NullTime
	if err := r.db.QueryRow(lastUsedQuery, workoutID).Scan(&nullableLastUsed); err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get last usage: %w", err)
	}
	if nullableLastUsed.Valid {
		lastUsedAt = &nullableLastUsed.Time
	}

	// Construct the stats response
	result := &domain.WorkoutWithUsageStats{
		Workout:    *workout,
		TimesUsed:  int(timesUsed),
		LastUsedAt: lastUsedAt,
	}

	return result, nil
}

// CopyToStandard creates a standard workout template by copying a user-created one (including movements and WODs)
func (r *WorkoutRepository) CopyToStandard(id int64, newName string) (*domain.Workout, error) {
	// Get the source workout with details
	source, err := r.GetByIDWithDetails(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get source workout: %w", err)
	}
	if source == nil {
		return nil, fmt.Errorf("source workout not found")
	}

	// Create a new standard workout
	now := time.Now()
	standardWorkout := &domain.Workout{
		Name:      newName,
		Notes:     source.Notes,
		CreatedBy: nil, // Standard workouts have no creator
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `INSERT INTO workouts (name, notes, created_by, created_at, updated_at)
	          VALUES (?, ?, NULL, ?, ?)`

	result, err := r.db.Exec(query,
		standardWorkout.Name,
		standardWorkout.Notes,
		standardWorkout.CreatedAt,
		standardWorkout.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create standard workout: %w", err)
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get new workout ID: %w", err)
	}
	standardWorkout.ID = newID

	// Copy associated movements
	if len(source.Movements) > 0 {
		movementQuery := `INSERT INTO workout_movements (workout_id, movement_id, weight, sets, reps, time, distance, is_rx, is_pr, notes, order_index, created_at, updated_at)
		                  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		for _, m := range source.Movements {
			_, err := r.db.Exec(movementQuery,
				newID,
				m.MovementID,
				m.Weight,
				m.Sets,
				m.Reps,
				m.Time,
				m.Distance,
				m.IsRx,
				false, // is_pr - reset for standard template
				m.Notes,
				m.OrderIndex,
				now,
				now,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to copy workout movement: %w", err)
			}
		}
	}

	// Copy associated WODs
	if len(source.WODs) > 0 {
		wodQuery := `INSERT INTO workout_wods (workout_id, wod_id, order_index, created_at, updated_at)
		             VALUES (?, ?, ?, ?, ?)`
		for _, w := range source.WODs {
			_, err := r.db.Exec(wodQuery,
				newID,
				w.WODID,
				w.OrderIndex,
				now,
				now,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to copy workout wod: %w", err)
			}
		}
	}

	return standardWorkout, nil
}

// scanWorkouts scans multiple workout rows
func (r *WorkoutRepository) scanWorkouts(rows *sql.Rows) ([]*domain.Workout, error) {
	var workouts []*domain.Workout
	for rows.Next() {
		workout := &domain.Workout{}
		var createdBy sql.NullInt64
		var notes sql.NullString

		err := rows.Scan(&workout.ID, &workout.Name, &notes, &createdBy, &workout.CreatedAt, &workout.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if notes.Valid {
			workout.Notes = &notes.String
		}
		if createdBy.Valid {
			workout.CreatedBy = &createdBy.Int64
		}

		workouts = append(workouts, workout)
	}

	return workouts, rows.Err()
}
