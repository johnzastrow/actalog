package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

type ClassSessionRepository struct {
	db *sql.DB
}

func NewClassSessionRepository(db *sql.DB) *ClassSessionRepository {
	return &ClassSessionRepository{db: db}
}

// Create creates a new class session
func (r *ClassSessionRepository) Create(session *domain.ClassSession) error {
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()

	query := rebindQuery(`
		INSERT INTO class_sessions (organization_id, template_id, location_id, name, description,
		                            workout_id, start_time, end_time, capacity, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			session.OrganizationID,
			session.TemplateID,
			session.LocationID,
			session.Name,
			session.Description,
			session.WorkoutID,
			session.StartTime,
			session.EndTime,
			session.Capacity,
			session.Status,
			session.CreatedAt,
			session.UpdatedAt,
		).Scan(&session.ID)
	}

	result, err := r.db.Exec(query,
		session.OrganizationID,
		session.TemplateID,
		session.LocationID,
		session.Name,
		session.Description,
		session.WorkoutID,
		session.StartTime,
		session.EndTime,
		session.Capacity,
		session.Status,
		session.CreatedAt,
		session.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create class session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get class session ID: %w", err)
	}

	session.ID = id
	return nil
}

// GetByID retrieves a class session by ID with related data
func (r *ClassSessionRepository) GetByID(id int64) (*domain.ClassSession, error) {
	query := rebindQuery(`
		SELECT cs.id, cs.organization_id, cs.template_id, cs.location_id, cs.name, cs.description,
		       cs.workout_id, cs.start_time, cs.end_time, cs.capacity, cs.status,
		       cs.cancelled_at, cs.cancelled_reason, cs.completed_at, cs.created_at, cs.updated_at,
		       ct.name as template_name, gl.name as location_name, w.name as workout_name
		FROM class_sessions cs
		LEFT JOIN class_templates ct ON cs.template_id = ct.id
		LEFT JOIN gym_locations gl ON cs.location_id = gl.id
		LEFT JOIN workouts w ON cs.workout_id = w.id
		WHERE cs.id = ?
	`)

	session := &domain.ClassSession{}
	var templateID, locationID, workoutID sql.NullInt64
	var description, cancelledReason sql.NullString
	var cancelledAt, completedAt sql.NullTime
	var templateName, locationName, workoutName sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&session.ID,
		&session.OrganizationID,
		&templateID,
		&locationID,
		&session.Name,
		&description,
		&workoutID,
		&session.StartTime,
		&session.EndTime,
		&session.Capacity,
		&session.Status,
		&cancelledAt,
		&cancelledReason,
		&completedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
		&templateName,
		&locationName,
		&workoutName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get class session: %w", err)
	}

	if templateID.Valid {
		session.TemplateID = &templateID.Int64
	}
	if locationID.Valid {
		session.LocationID = &locationID.Int64
	}
	if workoutID.Valid {
		session.WorkoutID = &workoutID.Int64
	}
	if description.Valid {
		session.Description = &description.String
	}
	if cancelledAt.Valid {
		session.CancelledAt = &cancelledAt.Time
	}
	if cancelledReason.Valid {
		session.CancelledReason = &cancelledReason.String
	}
	if completedAt.Valid {
		session.CompletedAt = &completedAt.Time
	}
	if templateName.Valid {
		session.TemplateName = &templateName.String
	}
	if locationName.Valid {
		session.LocationName = &locationName.String
	}
	if workoutName.Valid {
		session.WorkoutName = &workoutName.String
	}

	// Get reservation count
	count, err := r.GetReservationCount(session.ID)
	if err != nil {
		return nil, err
	}
	session.ReservationCount = count
	session.AvailableSpots = session.Capacity - count

	return session, nil
}

// GetByOrganizationID retrieves all class sessions for an organization within a date range
func (r *ClassSessionRepository) GetByOrganizationID(orgID int64, startDate, endDate time.Time) ([]*domain.ClassSession, error) {
	query := rebindQuery(`
		SELECT cs.id, cs.organization_id, cs.template_id, cs.location_id, cs.name, cs.description,
		       cs.workout_id, cs.start_time, cs.end_time, cs.capacity, cs.status,
		       cs.cancelled_at, cs.cancelled_reason, cs.completed_at, cs.created_at, cs.updated_at,
		       ct.name as template_name, gl.name as location_name, w.name as workout_name,
		       (SELECT COUNT(*) FROM reservations r WHERE r.session_id = cs.id AND r.status IN ('reserved', 'checked_in', 'attended')) as reservation_count
		FROM class_sessions cs
		LEFT JOIN class_templates ct ON cs.template_id = ct.id
		LEFT JOIN gym_locations gl ON cs.location_id = gl.id
		LEFT JOIN workouts w ON cs.workout_id = w.id
		WHERE cs.organization_id = ? AND cs.start_time >= ? AND cs.start_time <= ?
		ORDER BY cs.start_time ASC
	`)

	rows, err := r.db.Query(query, orgID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to list class sessions: %w", err)
	}
	defer rows.Close()

	return r.scanSessions(rows)
}

// GetByOrganizationIDWithUserStatus retrieves sessions with the current user's reservation status
func (r *ClassSessionRepository) GetByOrganizationIDWithUserStatus(orgID, userID int64, startDate, endDate time.Time) ([]*domain.ClassSession, error) {
	query := rebindQuery(`
		SELECT cs.id, cs.organization_id, cs.template_id, cs.location_id, cs.name, cs.description,
		       cs.workout_id, cs.start_time, cs.end_time, cs.capacity, cs.status,
		       cs.cancelled_at, cs.cancelled_reason, cs.completed_at, cs.created_at, cs.updated_at,
		       ct.name as template_name, gl.name as location_name, w.name as workout_name,
		       (SELECT COUNT(*) FROM reservations r WHERE r.session_id = cs.id AND r.status IN ('reserved', 'checked_in', 'attended')) as reservation_count,
		       (SELECT r2.status FROM reservations r2 WHERE r2.session_id = cs.id AND r2.user_id = ? LIMIT 1) as user_status
		FROM class_sessions cs
		LEFT JOIN class_templates ct ON cs.template_id = ct.id
		LEFT JOIN gym_locations gl ON cs.location_id = gl.id
		LEFT JOIN workouts w ON cs.workout_id = w.id
		WHERE cs.organization_id = ? AND cs.start_time >= ? AND cs.start_time <= ?
		ORDER BY cs.start_time ASC
	`)

	rows, err := r.db.Query(query, userID, orgID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to list class sessions with user status: %w", err)
	}
	defer rows.Close()

	return r.scanSessionsWithUserStatus(rows)
}

// GetUpcomingByCoachID retrieves upcoming sessions for a coach
func (r *ClassSessionRepository) GetUpcomingByCoachID(coachUserID int64, limit int) ([]*domain.ClassSession, error) {
	query := rebindQuery(`
		SELECT DISTINCT cs.id, cs.organization_id, cs.template_id, cs.location_id, cs.name, cs.description,
		       cs.workout_id, cs.start_time, cs.end_time, cs.capacity, cs.status,
		       cs.cancelled_at, cs.cancelled_reason, cs.completed_at, cs.created_at, cs.updated_at,
		       ct.name as template_name, gl.name as location_name, w.name as workout_name,
		       (SELECT COUNT(*) FROM reservations r WHERE r.session_id = cs.id AND r.status IN ('reserved', 'checked_in', 'attended')) as reservation_count
		FROM class_sessions cs
		JOIN session_coaches sc ON cs.id = sc.session_id
		LEFT JOIN class_templates ct ON cs.template_id = ct.id
		LEFT JOIN gym_locations gl ON cs.location_id = gl.id
		LEFT JOIN workouts w ON cs.workout_id = w.id
		WHERE sc.user_id = ? AND cs.start_time >= ? AND cs.status IN ('scheduled', 'in_progress')
		ORDER BY cs.start_time ASC
		LIMIT ?
	`)

	rows, err := r.db.Query(query, coachUserID, time.Now(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list coach sessions: %w", err)
	}
	defer rows.Close()

	return r.scanSessions(rows)
}

func (r *ClassSessionRepository) scanSessions(rows *sql.Rows) ([]*domain.ClassSession, error) {
	var sessions []*domain.ClassSession

	for rows.Next() {
		session := &domain.ClassSession{}
		var templateID, locationID, workoutID sql.NullInt64
		var description, cancelledReason sql.NullString
		var cancelledAt, completedAt sql.NullTime
		var templateName, locationName, workoutName sql.NullString
		var reservationCount int

		err := rows.Scan(
			&session.ID,
			&session.OrganizationID,
			&templateID,
			&locationID,
			&session.Name,
			&description,
			&workoutID,
			&session.StartTime,
			&session.EndTime,
			&session.Capacity,
			&session.Status,
			&cancelledAt,
			&cancelledReason,
			&completedAt,
			&session.CreatedAt,
			&session.UpdatedAt,
			&templateName,
			&locationName,
			&workoutName,
			&reservationCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan class session: %w", err)
		}

		if templateID.Valid {
			session.TemplateID = &templateID.Int64
		}
		if locationID.Valid {
			session.LocationID = &locationID.Int64
		}
		if workoutID.Valid {
			session.WorkoutID = &workoutID.Int64
		}
		if description.Valid {
			session.Description = &description.String
		}
		if cancelledAt.Valid {
			session.CancelledAt = &cancelledAt.Time
		}
		if cancelledReason.Valid {
			session.CancelledReason = &cancelledReason.String
		}
		if completedAt.Valid {
			session.CompletedAt = &completedAt.Time
		}
		if templateName.Valid {
			session.TemplateName = &templateName.String
		}
		if locationName.Valid {
			session.LocationName = &locationName.String
		}
		if workoutName.Valid {
			session.WorkoutName = &workoutName.String
		}

		session.ReservationCount = reservationCount
		session.AvailableSpots = session.Capacity - reservationCount

		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

func (r *ClassSessionRepository) scanSessionsWithUserStatus(rows *sql.Rows) ([]*domain.ClassSession, error) {
	var sessions []*domain.ClassSession

	for rows.Next() {
		session := &domain.ClassSession{}
		var templateID, locationID, workoutID sql.NullInt64
		var description, cancelledReason sql.NullString
		var cancelledAt, completedAt sql.NullTime
		var templateName, locationName, workoutName sql.NullString
		var reservationCount int
		var userStatus sql.NullString

		err := rows.Scan(
			&session.ID,
			&session.OrganizationID,
			&templateID,
			&locationID,
			&session.Name,
			&description,
			&workoutID,
			&session.StartTime,
			&session.EndTime,
			&session.Capacity,
			&session.Status,
			&cancelledAt,
			&cancelledReason,
			&completedAt,
			&session.CreatedAt,
			&session.UpdatedAt,
			&templateName,
			&locationName,
			&workoutName,
			&reservationCount,
			&userStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan class session: %w", err)
		}

		if templateID.Valid {
			session.TemplateID = &templateID.Int64
		}
		if locationID.Valid {
			session.LocationID = &locationID.Int64
		}
		if workoutID.Valid {
			session.WorkoutID = &workoutID.Int64
		}
		if description.Valid {
			session.Description = &description.String
		}
		if cancelledAt.Valid {
			session.CancelledAt = &cancelledAt.Time
		}
		if cancelledReason.Valid {
			session.CancelledReason = &cancelledReason.String
		}
		if completedAt.Valid {
			session.CompletedAt = &completedAt.Time
		}
		if templateName.Valid {
			session.TemplateName = &templateName.String
		}
		if locationName.Valid {
			session.LocationName = &locationName.String
		}
		if workoutName.Valid {
			session.WorkoutName = &workoutName.String
		}
		if userStatus.Valid {
			session.CurrentUserStatus = &userStatus.String
		}

		session.ReservationCount = reservationCount
		session.AvailableSpots = session.Capacity - reservationCount

		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// Update updates a class session
func (r *ClassSessionRepository) Update(session *domain.ClassSession) error {
	session.UpdatedAt = time.Now()

	query := rebindQuery(`
		UPDATE class_sessions
		SET template_id = ?, location_id = ?, name = ?, description = ?,
		    workout_id = ?, start_time = ?, end_time = ?, capacity = ?, status = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query,
		session.TemplateID,
		session.LocationID,
		session.Name,
		session.Description,
		session.WorkoutID,
		session.StartTime,
		session.EndTime,
		session.Capacity,
		session.Status,
		session.UpdatedAt,
		session.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update class session: %w", err)
	}

	return nil
}

// UpdateStatus updates only the status of a session
func (r *ClassSessionRepository) UpdateStatus(id int64, status string) error {
	query := rebindQuery(`
		UPDATE class_sessions
		SET status = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update class session status: %w", err)
	}

	return nil
}

// Cancel cancels a session with reason
func (r *ClassSessionRepository) Cancel(id int64, reason string) error {
	now := time.Now()
	query := rebindQuery(`
		UPDATE class_sessions
		SET status = ?, cancelled_at = ?, cancelled_reason = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query, domain.SessionStatusCancelled, now, reason, now, id)
	if err != nil {
		return fmt.Errorf("failed to cancel class session: %w", err)
	}

	return nil
}

// Complete marks a session as completed
func (r *ClassSessionRepository) Complete(id int64) error {
	now := time.Now()
	query := rebindQuery(`
		UPDATE class_sessions
		SET status = ?, completed_at = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query, domain.SessionStatusCompleted, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to complete class session: %w", err)
	}

	return nil
}

// GetReservationCount returns the count of active reservations for a session
func (r *ClassSessionRepository) GetReservationCount(sessionID int64) (int, error) {
	query := rebindQuery(`
		SELECT COUNT(*) FROM reservations
		WHERE session_id = ? AND status IN ('reserved', 'checked_in', 'attended')
	`)

	var count int
	err := r.db.QueryRow(query, sessionID).Scan(&count)
	return count, err
}

// ExistsByTemplateAndStartTime checks if a session already exists for a given template and start time
func (r *ClassSessionRepository) ExistsByTemplateAndStartTime(templateID int64, startTime time.Time) (bool, error) {
	query := rebindQuery(`
		SELECT COUNT(*) FROM class_sessions
		WHERE template_id = ? AND start_time = ? AND status != 'cancelled'
	`)

	var count int
	err := r.db.QueryRow(query, templateID, startTime).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
