package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

type SessionCoachRepository struct {
	db *sql.DB
}

func NewSessionCoachRepository(db *sql.DB) *SessionCoachRepository {
	return &SessionCoachRepository{db: db}
}

// Create creates a new session coach assignment
func (r *SessionCoachRepository) Create(coach *domain.SessionCoach) error {
	coach.CreatedAt = time.Now()

	query := rebindQuery(`
		INSERT INTO session_coaches (session_id, user_id, is_lead, created_at)
		VALUES (?, ?, ?, ?)
	`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			coach.SessionID,
			coach.UserID,
			coach.IsLead,
			coach.CreatedAt,
		).Scan(&coach.ID)
	}

	result, err := r.db.Exec(query,
		coach.SessionID,
		coach.UserID,
		coach.IsLead,
		coach.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create session coach: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get session coach ID: %w", err)
	}

	coach.ID = id
	return nil
}

// GetBySessionID retrieves all coaches for a session
func (r *SessionCoachRepository) GetBySessionID(sessionID int64) ([]*domain.SessionCoach, error) {
	query := rebindQuery(`
		SELECT sc.id, sc.session_id, sc.user_id, sc.is_lead, sc.created_at,
		       u.name as user_name, u.email as user_email
		FROM session_coaches sc
		JOIN users u ON sc.user_id = u.id
		WHERE sc.session_id = ?
		ORDER BY sc.is_lead DESC, u.name ASC
	`)

	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list session coaches: %w", err)
	}
	defer rows.Close()

	var coaches []*domain.SessionCoach
	for rows.Next() {
		coach := &domain.SessionCoach{}
		var userName, userEmail sql.NullString

		err := rows.Scan(
			&coach.ID,
			&coach.SessionID,
			&coach.UserID,
			&coach.IsLead,
			&coach.CreatedAt,
			&userName,
			&userEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session coach: %w", err)
		}

		if userName.Valid {
			coach.UserName = &userName.String
		}
		if userEmail.Valid {
			coach.UserEmail = &userEmail.String
		}

		coaches = append(coaches, coach)
	}

	return coaches, rows.Err()
}

// Delete deletes a session coach assignment by ID
func (r *SessionCoachRepository) Delete(id int64) error {
	query := rebindQuery(`DELETE FROM session_coaches WHERE id = ?`)

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session coach: %w", err)
	}

	return nil
}

// DeleteBySessionAndUser deletes a session coach by session and user IDs
func (r *SessionCoachRepository) DeleteBySessionAndUser(sessionID, userID int64) error {
	query := rebindQuery(`DELETE FROM session_coaches WHERE session_id = ? AND user_id = ?`)

	_, err := r.db.Exec(query, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete session coach: %w", err)
	}

	return nil
}

// SetLeadCoach sets the specified coach as lead and removes lead from others
func (r *SessionCoachRepository) SetLeadCoach(sessionID, userID int64) error {
	// First, remove lead from all coaches for this session
	query1 := rebindQuery(`UPDATE session_coaches SET is_lead = ? WHERE session_id = ?`)
	if _, err := r.db.Exec(query1, false, sessionID); err != nil {
		return fmt.Errorf("failed to reset lead coaches: %w", err)
	}

	// Then set the specified coach as lead
	query2 := rebindQuery(`UPDATE session_coaches SET is_lead = ? WHERE session_id = ? AND user_id = ?`)
	if _, err := r.db.Exec(query2, true, sessionID, userID); err != nil {
		return fmt.Errorf("failed to set lead coach: %w", err)
	}

	return nil
}

// DeleteBySessionIDs deletes all coach assignments for the given sessions
func (r *SessionCoachRepository) DeleteBySessionIDs(sessionIDs []int64) error {
	if len(sessionIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(sessionIDs))
	args := make([]interface{}, len(sessionIDs))
	for i, id := range sessionIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`DELETE FROM session_coaches WHERE session_id IN (%s)`, strings.Join(placeholders, ","))
	query = rebindQuery(query)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete session coaches: %w", err)
	}

	return nil
}
