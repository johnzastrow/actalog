package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

type CoachAssignmentRepository struct {
	db *sql.DB
}

func NewCoachAssignmentRepository(db *sql.DB) *CoachAssignmentRepository {
	return &CoachAssignmentRepository{db: db}
}

// Create creates a new coach assignment
func (r *CoachAssignmentRepository) Create(assignment *domain.CoachAssignment) error {
	assignment.AssignedAt = time.Now()
	assignment.CreatedAt = time.Now()
	assignment.UpdatedAt = time.Now()

	query := rebindQuery(`
		INSERT INTO coach_assignments (organization_id, user_id, is_active, assigned_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			assignment.OrganizationID,
			assignment.UserID,
			assignment.IsActive,
			assignment.AssignedAt,
			assignment.CreatedAt,
			assignment.UpdatedAt,
		).Scan(&assignment.ID)
	}

	result, err := r.db.Exec(query,
		assignment.OrganizationID,
		assignment.UserID,
		assignment.IsActive,
		assignment.AssignedAt,
		assignment.CreatedAt,
		assignment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create coach assignment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get coach assignment ID: %w", err)
	}

	assignment.ID = id
	return nil
}

// GetByID retrieves a coach assignment by ID
func (r *CoachAssignmentRepository) GetByID(id int64) (*domain.CoachAssignment, error) {
	query := rebindQuery(`
		SELECT ca.id, ca.organization_id, ca.user_id, ca.is_active,
		       ca.assigned_at, ca.created_at, ca.updated_at,
		       u.name as user_name, u.email as user_email
		FROM coach_assignments ca
		JOIN users u ON ca.user_id = u.id
		WHERE ca.id = ?
	`)

	assignment := &domain.CoachAssignment{}
	var userName, userEmail sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&assignment.ID,
		&assignment.OrganizationID,
		&assignment.UserID,
		&assignment.IsActive,
		&assignment.AssignedAt,
		&assignment.CreatedAt,
		&assignment.UpdatedAt,
		&userName,
		&userEmail,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get coach assignment: %w", err)
	}

	if userName.Valid {
		assignment.UserName = &userName.String
	}
	if userEmail.Valid {
		assignment.UserEmail = &userEmail.String
	}

	return assignment, nil
}

// GetByOrganizationID retrieves all coach assignments for an organization
func (r *CoachAssignmentRepository) GetByOrganizationID(orgID int64, includeInactive bool) ([]*domain.CoachAssignment, error) {
	var query string
	if includeInactive {
		query = rebindQuery(`
			SELECT ca.id, ca.organization_id, ca.user_id, ca.is_active,
			       ca.assigned_at, ca.created_at, ca.updated_at,
			       u.name as user_name, u.email as user_email
			FROM coach_assignments ca
			JOIN users u ON ca.user_id = u.id
			WHERE ca.organization_id = ?
			ORDER BY u.name ASC
		`)
	} else {
		query = rebindQuery(`
			SELECT ca.id, ca.organization_id, ca.user_id, ca.is_active,
			       ca.assigned_at, ca.created_at, ca.updated_at,
			       u.name as user_name, u.email as user_email
			FROM coach_assignments ca
			JOIN users u ON ca.user_id = u.id
			WHERE ca.organization_id = ? AND ca.is_active = ?
			ORDER BY u.name ASC
		`)
	}

	var rows *sql.Rows
	var err error
	if includeInactive {
		rows, err = r.db.Query(query, orgID)
	} else {
		rows, err = r.db.Query(query, orgID, true)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list coach assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*domain.CoachAssignment
	for rows.Next() {
		assignment := &domain.CoachAssignment{}
		var userName, userEmail sql.NullString

		err := rows.Scan(
			&assignment.ID,
			&assignment.OrganizationID,
			&assignment.UserID,
			&assignment.IsActive,
			&assignment.AssignedAt,
			&assignment.CreatedAt,
			&assignment.UpdatedAt,
			&userName,
			&userEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coach assignment: %w", err)
		}

		if userName.Valid {
			assignment.UserName = &userName.String
		}
		if userEmail.Valid {
			assignment.UserEmail = &userEmail.String
		}

		assignments = append(assignments, assignment)
	}

	return assignments, rows.Err()
}

// GetByUserID retrieves all coach assignments for a user
func (r *CoachAssignmentRepository) GetByUserID(userID int64) ([]*domain.CoachAssignment, error) {
	query := rebindQuery(`
		SELECT ca.id, ca.organization_id, ca.user_id, ca.is_active,
		       ca.assigned_at, ca.created_at, ca.updated_at,
		       u.name as user_name, u.email as user_email
		FROM coach_assignments ca
		JOIN users u ON ca.user_id = u.id
		WHERE ca.user_id = ? AND ca.is_active = ?
		ORDER BY ca.assigned_at DESC
	`)

	rows, err := r.db.Query(query, userID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to list coach assignments by user: %w", err)
	}
	defer rows.Close()

	var assignments []*domain.CoachAssignment
	for rows.Next() {
		assignment := &domain.CoachAssignment{}
		var userName, userEmail sql.NullString

		err := rows.Scan(
			&assignment.ID,
			&assignment.OrganizationID,
			&assignment.UserID,
			&assignment.IsActive,
			&assignment.AssignedAt,
			&assignment.CreatedAt,
			&assignment.UpdatedAt,
			&userName,
			&userEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coach assignment: %w", err)
		}

		if userName.Valid {
			assignment.UserName = &userName.String
		}
		if userEmail.Valid {
			assignment.UserEmail = &userEmail.String
		}

		assignments = append(assignments, assignment)
	}

	return assignments, rows.Err()
}

// IsCoachForOrganization checks if a user is an active coach for an organization
func (r *CoachAssignmentRepository) IsCoachForOrganization(userID, orgID int64) (bool, error) {
	query := rebindQuery(`
		SELECT EXISTS(
			SELECT 1 FROM coach_assignments
			WHERE user_id = ? AND organization_id = ? AND is_active = ?
		)
	`)

	var exists bool
	err := r.db.QueryRow(query, userID, orgID, true).Scan(&exists)
	return exists, err
}

// Update updates a coach assignment
func (r *CoachAssignmentRepository) Update(assignment *domain.CoachAssignment) error {
	assignment.UpdatedAt = time.Now()

	query := rebindQuery(`
		UPDATE coach_assignments
		SET is_active = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query,
		assignment.IsActive,
		assignment.UpdatedAt,
		assignment.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update coach assignment: %w", err)
	}

	return nil
}

// Delete deletes a coach assignment
func (r *CoachAssignmentRepository) Delete(id int64) error {
	query := rebindQuery(`DELETE FROM coach_assignments WHERE id = ?`)

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete coach assignment: %w", err)
	}

	return nil
}
