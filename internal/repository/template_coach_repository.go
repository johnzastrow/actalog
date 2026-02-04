package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

type TemplateCoachRepository struct {
	db *sql.DB
}

func NewTemplateCoachRepository(db *sql.DB) *TemplateCoachRepository {
	return &TemplateCoachRepository{db: db}
}

// Add adds a coach to a template
func (r *TemplateCoachRepository) Add(templateID, userID int64, isLead bool) error {
	query := rebindQuery(`
		INSERT INTO template_coaches (template_id, user_id, is_lead, created_at)
		VALUES (?, ?, ?, ?)
	`)

	_, err := r.db.Exec(query, templateID, userID, isLead, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add template coach: %w", err)
	}

	return nil
}

// Remove removes a coach from a template
func (r *TemplateCoachRepository) Remove(templateID, userID int64) error {
	query := rebindQuery(`DELETE FROM template_coaches WHERE template_id = ? AND user_id = ?`)

	_, err := r.db.Exec(query, templateID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove template coach: %w", err)
	}

	return nil
}

// GetByTemplateID retrieves all coaches for a template
func (r *TemplateCoachRepository) GetByTemplateID(templateID int64) ([]*domain.TemplateCoach, error) {
	query := rebindQuery(`
		SELECT tc.id, tc.template_id, tc.user_id, tc.is_lead, tc.created_at,
		       u.name as user_name, u.email as user_email
		FROM template_coaches tc
		JOIN users u ON tc.user_id = u.id
		WHERE tc.template_id = ?
		ORDER BY tc.is_lead DESC, u.name ASC
	`)

	rows, err := r.db.Query(query, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template coaches: %w", err)
	}
	defer rows.Close()

	var coaches []*domain.TemplateCoach
	for rows.Next() {
		coach := &domain.TemplateCoach{}
		var userName, userEmail sql.NullString

		err := rows.Scan(
			&coach.ID,
			&coach.TemplateID,
			&coach.UserID,
			&coach.IsLead,
			&coach.CreatedAt,
			&userName,
			&userEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template coach: %w", err)
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

// RemoveAllByTemplateID removes all coaches from a template
func (r *TemplateCoachRepository) RemoveAllByTemplateID(templateID int64) error {
	query := rebindQuery(`DELETE FROM template_coaches WHERE template_id = ?`)

	_, err := r.db.Exec(query, templateID)
	if err != nil {
		return fmt.Errorf("failed to remove all template coaches: %w", err)
	}

	return nil
}

// SetLeadCoach sets a coach as lead (and removes lead from others)
func (r *TemplateCoachRepository) SetLeadCoach(templateID, userID int64) error {
	// First, remove lead from all coaches
	query1 := rebindQuery(`UPDATE template_coaches SET is_lead = ? WHERE template_id = ?`)
	if _, err := r.db.Exec(query1, false, templateID); err != nil {
		return fmt.Errorf("failed to reset lead coaches: %w", err)
	}

	// Then set the specified coach as lead
	query2 := rebindQuery(`UPDATE template_coaches SET is_lead = ? WHERE template_id = ? AND user_id = ?`)
	if _, err := r.db.Exec(query2, true, templateID, userID); err != nil {
		return fmt.Errorf("failed to set lead coach: %w", err)
	}

	return nil
}
