package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

type ReservationRepository struct {
	db *sql.DB
}

func NewReservationRepository(db *sql.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

// Create creates a new reservation
func (r *ReservationRepository) Create(reservation *domain.Reservation) error {
	reservation.ReservedAt = time.Now()
	reservation.CreatedAt = time.Now()
	reservation.UpdatedAt = time.Now()

	query := rebindQuery(`
		INSERT INTO reservations (session_id, user_id, status, reserved_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			reservation.SessionID,
			reservation.UserID,
			reservation.Status,
			reservation.ReservedAt,
			reservation.CreatedAt,
			reservation.UpdatedAt,
		).Scan(&reservation.ID)
	}

	result, err := r.db.Exec(query,
		reservation.SessionID,
		reservation.UserID,
		reservation.Status,
		reservation.ReservedAt,
		reservation.CreatedAt,
		reservation.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create reservation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get reservation ID: %w", err)
	}

	reservation.ID = id
	return nil
}

// GetByID retrieves a reservation by ID
func (r *ReservationRepository) GetByID(id int64) (*domain.Reservation, error) {
	query := rebindQuery(`
		SELECT r.id, r.session_id, r.user_id, r.status, r.reserved_at,
		       r.checked_in_at, r.checked_in_by_user_id, r.cancelled_at, r.cancelled_reason,
		       r.no_show_marked_at, r.user_workout_id, r.created_at, r.updated_at,
		       u.name as user_name, u.email as user_email,
		       cb.name as checked_in_by_name, cb.email as checked_in_by_email
		FROM reservations r
		JOIN users u ON r.user_id = u.id
		LEFT JOIN users cb ON r.checked_in_by_user_id = cb.id
		WHERE r.id = ?
	`)

	reservation := &domain.Reservation{}
	var checkedInAt, cancelledAt, noShowMarkedAt sql.NullTime
	var checkedInByUserID, userWorkoutID sql.NullInt64
	var cancelledReason sql.NullString
	var userName, userEmail, checkedInByName, checkedInByEmail sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&reservation.ID,
		&reservation.SessionID,
		&reservation.UserID,
		&reservation.Status,
		&reservation.ReservedAt,
		&checkedInAt,
		&checkedInByUserID,
		&cancelledAt,
		&cancelledReason,
		&noShowMarkedAt,
		&userWorkoutID,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
		&userName,
		&userEmail,
		&checkedInByName,
		&checkedInByEmail,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	r.populateReservationNullableFields(reservation, checkedInAt, checkedInByUserID, cancelledAt,
		cancelledReason, noShowMarkedAt, userWorkoutID, userName, userEmail, checkedInByName, checkedInByEmail)

	return reservation, nil
}

// GetBySessionID retrieves all reservations for a session
func (r *ReservationRepository) GetBySessionID(sessionID int64) ([]*domain.Reservation, error) {
	query := rebindQuery(`
		SELECT r.id, r.session_id, r.user_id, r.status, r.reserved_at,
		       r.checked_in_at, r.checked_in_by_user_id, r.cancelled_at, r.cancelled_reason,
		       r.no_show_marked_at, r.user_workout_id, r.created_at, r.updated_at,
		       u.name as user_name, u.email as user_email,
		       cb.name as checked_in_by_name, cb.email as checked_in_by_email
		FROM reservations r
		JOIN users u ON r.user_id = u.id
		LEFT JOIN users cb ON r.checked_in_by_user_id = cb.id
		WHERE r.session_id = ?
		ORDER BY r.reserved_at ASC
	`)

	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list reservations: %w", err)
	}
	defer rows.Close()

	return r.scanReservations(rows)
}

// GetBySessionIDWithStatus retrieves reservations for a session filtered by statuses
func (r *ReservationRepository) GetBySessionIDWithStatus(sessionID int64, statuses []string) ([]*domain.Reservation, error) {
	if len(statuses) == 0 {
		return r.GetBySessionID(sessionID)
	}

	placeholders := make([]string, len(statuses))
	args := make([]interface{}, 0, len(statuses)+1)
	args = append(args, sessionID)

	for i, status := range statuses {
		placeholders[i] = "?"
		args = append(args, status)
	}

	query := rebindQuery(fmt.Sprintf(`
		SELECT r.id, r.session_id, r.user_id, r.status, r.reserved_at,
		       r.checked_in_at, r.checked_in_by_user_id, r.cancelled_at, r.cancelled_reason,
		       r.no_show_marked_at, r.user_workout_id, r.created_at, r.updated_at,
		       u.name as user_name, u.email as user_email,
		       cb.name as checked_in_by_name, cb.email as checked_in_by_email
		FROM reservations r
		JOIN users u ON r.user_id = u.id
		LEFT JOIN users cb ON r.checked_in_by_user_id = cb.id
		WHERE r.session_id = ? AND r.status IN (%s)
		ORDER BY r.reserved_at ASC
	`, strings.Join(placeholders, ",")))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list reservations with status: %w", err)
	}
	defer rows.Close()

	return r.scanReservations(rows)
}

// GetByUserID retrieves reservations for a user with pagination
func (r *ReservationRepository) GetByUserID(userID int64, limit, offset int) ([]*domain.ReservationWithSession, error) {
	query := rebindQuery(`
		SELECT r.id, r.session_id, r.user_id, r.status, r.reserved_at,
		       r.checked_in_at, r.checked_in_by_user_id, r.cancelled_at, r.cancelled_reason,
		       r.no_show_marked_at, r.user_workout_id, r.created_at, r.updated_at,
		       cs.name as session_name, cs.start_time as session_start_time,
		       cs.end_time as session_end_time, cs.status as session_status,
		       gl.name as location_name, cs.organization_id, o.name as organization_name
		FROM reservations r
		JOIN class_sessions cs ON r.session_id = cs.id
		LEFT JOIN gym_locations gl ON cs.location_id = gl.id
		JOIN organizations o ON cs.organization_id = o.id
		WHERE r.user_id = ?
		ORDER BY cs.start_time DESC
		LIMIT ? OFFSET ?
	`)

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user reservations: %w", err)
	}
	defer rows.Close()

	return r.scanReservationsWithSession(rows)
}

// GetUpcomingByUserID retrieves upcoming reservations for a user
func (r *ReservationRepository) GetUpcomingByUserID(userID int64) ([]*domain.ReservationWithSession, error) {
	query := rebindQuery(`
		SELECT r.id, r.session_id, r.user_id, r.status, r.reserved_at,
		       r.checked_in_at, r.checked_in_by_user_id, r.cancelled_at, r.cancelled_reason,
		       r.no_show_marked_at, r.user_workout_id, r.created_at, r.updated_at,
		       cs.name as session_name, cs.start_time as session_start_time,
		       cs.end_time as session_end_time, cs.status as session_status,
		       gl.name as location_name, cs.organization_id, o.name as organization_name
		FROM reservations r
		JOIN class_sessions cs ON r.session_id = cs.id
		LEFT JOIN gym_locations gl ON cs.location_id = gl.id
		JOIN organizations o ON cs.organization_id = o.id
		WHERE r.user_id = ? AND cs.start_time >= ? AND r.status IN ('reserved', 'checked_in')
		ORDER BY cs.start_time ASC
	`)

	rows, err := r.db.Query(query, userID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to list upcoming user reservations: %w", err)
	}
	defer rows.Close()

	return r.scanReservationsWithSession(rows)
}

// GetBySessionAndUser retrieves a reservation by session and user
func (r *ReservationRepository) GetBySessionAndUser(sessionID, userID int64) (*domain.Reservation, error) {
	query := rebindQuery(`
		SELECT r.id, r.session_id, r.user_id, r.status, r.reserved_at,
		       r.checked_in_at, r.checked_in_by_user_id, r.cancelled_at, r.cancelled_reason,
		       r.no_show_marked_at, r.user_workout_id, r.created_at, r.updated_at,
		       u.name as user_name, u.email as user_email,
		       cb.name as checked_in_by_name, cb.email as checked_in_by_email
		FROM reservations r
		JOIN users u ON r.user_id = u.id
		LEFT JOIN users cb ON r.checked_in_by_user_id = cb.id
		WHERE r.session_id = ? AND r.user_id = ?
	`)

	reservation := &domain.Reservation{}
	var checkedInAt, cancelledAt, noShowMarkedAt sql.NullTime
	var checkedInByUserID, userWorkoutID sql.NullInt64
	var cancelledReason sql.NullString
	var userName, userEmail, checkedInByName, checkedInByEmail sql.NullString

	err := r.db.QueryRow(query, sessionID, userID).Scan(
		&reservation.ID,
		&reservation.SessionID,
		&reservation.UserID,
		&reservation.Status,
		&reservation.ReservedAt,
		&checkedInAt,
		&checkedInByUserID,
		&cancelledAt,
		&cancelledReason,
		&noShowMarkedAt,
		&userWorkoutID,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
		&userName,
		&userEmail,
		&checkedInByName,
		&checkedInByEmail,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get reservation by session and user: %w", err)
	}

	r.populateReservationNullableFields(reservation, checkedInAt, checkedInByUserID, cancelledAt,
		cancelledReason, noShowMarkedAt, userWorkoutID, userName, userEmail, checkedInByName, checkedInByEmail)

	return reservation, nil
}

// Update updates a reservation
func (r *ReservationRepository) Update(reservation *domain.Reservation) error {
	reservation.UpdatedAt = time.Now()

	query := rebindQuery(`
		UPDATE reservations
		SET status = ?, checked_in_at = ?, checked_in_by_user_id = ?,
		    cancelled_at = ?, cancelled_reason = ?, no_show_marked_at = ?,
		    user_workout_id = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query,
		reservation.Status,
		reservation.CheckedInAt,
		reservation.CheckedInByUserID,
		reservation.CancelledAt,
		reservation.CancelledReason,
		reservation.NoShowMarkedAt,
		reservation.UserWorkoutID,
		reservation.UpdatedAt,
		reservation.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update reservation: %w", err)
	}

	return nil
}

// UpdateStatus updates only the status of a reservation
func (r *ReservationRepository) UpdateStatus(id int64, status string) error {
	query := rebindQuery(`
		UPDATE reservations
		SET status = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	return nil
}

// CheckIn marks a reservation as checked in
func (r *ReservationRepository) CheckIn(id int64, checkedInByUserID int64) error {
	now := time.Now()
	query := rebindQuery(`
		UPDATE reservations
		SET status = ?, checked_in_at = ?, checked_in_by_user_id = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query, domain.ReservationStatusCheckedIn, now, checkedInByUserID, now, id)
	if err != nil {
		return fmt.Errorf("failed to check in reservation: %w", err)
	}

	return nil
}

// Cancel cancels a reservation with reason
func (r *ReservationRepository) Cancel(id int64, reason string) error {
	now := time.Now()
	query := rebindQuery(`
		UPDATE reservations
		SET status = ?, cancelled_at = ?, cancelled_reason = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query, domain.ReservationStatusCancelled, now, reason, now, id)
	if err != nil {
		return fmt.Errorf("failed to cancel reservation: %w", err)
	}

	return nil
}

// MarkNoShow marks a reservation as no-show
func (r *ReservationRepository) MarkNoShow(id int64) error {
	now := time.Now()
	query := rebindQuery(`
		UPDATE reservations
		SET status = ?, no_show_marked_at = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query, domain.ReservationStatusNoShow, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark reservation as no-show: %w", err)
	}

	return nil
}

// MarkAttended marks a reservation as attended and links to user workout
func (r *ReservationRepository) MarkAttended(id int64, userWorkoutID int64) error {
	now := time.Now()
	query := rebindQuery(`
		UPDATE reservations
		SET status = ?, user_workout_id = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query, domain.ReservationStatusAttended, userWorkoutID, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark reservation as attended: %w", err)
	}

	return nil
}

// CountActiveForSession counts active reservations for a session
func (r *ReservationRepository) CountActiveForSession(sessionID int64) (int, error) {
	query := rebindQuery(`
		SELECT COUNT(*) FROM reservations
		WHERE session_id = ? AND status IN ('reserved', 'checked_in', 'attended')
	`)

	var count int
	err := r.db.QueryRow(query, sessionID).Scan(&count)
	return count, err
}

func (r *ReservationRepository) scanReservations(rows *sql.Rows) ([]*domain.Reservation, error) {
	var reservations []*domain.Reservation

	for rows.Next() {
		reservation := &domain.Reservation{}
		var checkedInAt, cancelledAt, noShowMarkedAt sql.NullTime
		var checkedInByUserID, userWorkoutID sql.NullInt64
		var cancelledReason sql.NullString
		var userName, userEmail, checkedInByName, checkedInByEmail sql.NullString

		err := rows.Scan(
			&reservation.ID,
			&reservation.SessionID,
			&reservation.UserID,
			&reservation.Status,
			&reservation.ReservedAt,
			&checkedInAt,
			&checkedInByUserID,
			&cancelledAt,
			&cancelledReason,
			&noShowMarkedAt,
			&userWorkoutID,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&userName,
			&userEmail,
			&checkedInByName,
			&checkedInByEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}

		r.populateReservationNullableFields(reservation, checkedInAt, checkedInByUserID, cancelledAt,
			cancelledReason, noShowMarkedAt, userWorkoutID, userName, userEmail, checkedInByName, checkedInByEmail)

		reservations = append(reservations, reservation)
	}

	return reservations, rows.Err()
}

func (r *ReservationRepository) scanReservationsWithSession(rows *sql.Rows) ([]*domain.ReservationWithSession, error) {
	var reservations []*domain.ReservationWithSession

	for rows.Next() {
		res := &domain.ReservationWithSession{}
		var checkedInAt, cancelledAt, noShowMarkedAt sql.NullTime
		var checkedInByUserID, userWorkoutID sql.NullInt64
		var cancelledReason sql.NullString
		var locationName sql.NullString

		err := rows.Scan(
			&res.ID,
			&res.SessionID,
			&res.UserID,
			&res.Status,
			&res.ReservedAt,
			&checkedInAt,
			&checkedInByUserID,
			&cancelledAt,
			&cancelledReason,
			&noShowMarkedAt,
			&userWorkoutID,
			&res.CreatedAt,
			&res.UpdatedAt,
			&res.SessionName,
			&res.SessionStartTime,
			&res.SessionEndTime,
			&res.SessionStatus,
			&locationName,
			&res.OrganizationID,
			&res.OrganizationName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation with session: %w", err)
		}

		if checkedInAt.Valid {
			res.CheckedInAt = &checkedInAt.Time
		}
		if checkedInByUserID.Valid {
			res.CheckedInByUserID = &checkedInByUserID.Int64
		}
		if cancelledAt.Valid {
			res.CancelledAt = &cancelledAt.Time
		}
		if cancelledReason.Valid {
			res.CancelledReason = &cancelledReason.String
		}
		if noShowMarkedAt.Valid {
			res.NoShowMarkedAt = &noShowMarkedAt.Time
		}
		if userWorkoutID.Valid {
			res.UserWorkoutID = &userWorkoutID.Int64
		}
		if locationName.Valid {
			res.LocationName = &locationName.String
		}

		reservations = append(reservations, res)
	}

	return reservations, rows.Err()
}

func (r *ReservationRepository) populateReservationNullableFields(
	reservation *domain.Reservation,
	checkedInAt sql.NullTime,
	checkedInByUserID sql.NullInt64,
	cancelledAt sql.NullTime,
	cancelledReason sql.NullString,
	noShowMarkedAt sql.NullTime,
	userWorkoutID sql.NullInt64,
	userName sql.NullString,
	userEmail sql.NullString,
	checkedInByName sql.NullString,
	checkedInByEmail sql.NullString,
) {
	if checkedInAt.Valid {
		reservation.CheckedInAt = &checkedInAt.Time
	}
	if checkedInByUserID.Valid {
		reservation.CheckedInByUserID = &checkedInByUserID.Int64
	}
	if cancelledAt.Valid {
		reservation.CancelledAt = &cancelledAt.Time
	}
	if cancelledReason.Valid {
		reservation.CancelledReason = &cancelledReason.String
	}
	if noShowMarkedAt.Valid {
		reservation.NoShowMarkedAt = &noShowMarkedAt.Time
	}
	if userWorkoutID.Valid {
		reservation.UserWorkoutID = &userWorkoutID.Int64
	}
	if userName.Valid {
		reservation.UserName = &userName.String
	}
	if userEmail.Valid {
		reservation.UserEmail = &userEmail.String
	}
	if checkedInByName.Valid {
		reservation.CheckedInByName = &checkedInByName.String
	}
	if checkedInByEmail.Valid {
		reservation.CheckedInByEmail = &checkedInByEmail.String
	}
}

// GetBySessionIDs retrieves all reservations for the given sessions
func (r *ReservationRepository) GetBySessionIDs(sessionIDs []int64) ([]*domain.Reservation, error) {
	if len(sessionIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(sessionIDs))
	args := make([]interface{}, len(sessionIDs))
	for i, id := range sessionIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT r.id, r.session_id, r.user_id, r.status, r.reserved_at,
		       r.checked_in_at, r.checked_in_by_user_id, r.cancelled_at, r.cancelled_reason,
		       r.no_show_marked_at, r.user_workout_id, r.created_at, r.updated_at,
		       u.name as user_name, u.email as user_email,
		       cb.name as checked_in_by_name, cb.email as checked_in_by_email
		FROM reservations r
		JOIN users u ON r.user_id = u.id
		LEFT JOIN users cb ON r.checked_in_by_user_id = cb.id
		WHERE r.session_id IN (%s)
		ORDER BY r.reserved_at ASC
	`, strings.Join(placeholders, ","))

	query = rebindQuery(query)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservations by session IDs: %w", err)
	}
	defer rows.Close()

	return r.scanReservations(rows)
}

// DeleteBySessionIDs deletes all reservations for the given sessions
func (r *ReservationRepository) DeleteBySessionIDs(sessionIDs []int64) error {
	if len(sessionIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(sessionIDs))
	args := make([]interface{}, len(sessionIDs))
	for i, id := range sessionIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`DELETE FROM reservations WHERE session_id IN (%s)`, strings.Join(placeholders, ","))
	query = rebindQuery(query)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete reservations: %w", err)
	}

	return nil
}
