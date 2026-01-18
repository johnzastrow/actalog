package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

type ClassTemplateRepository struct {
	db *sql.DB
}

func NewClassTemplateRepository(db *sql.DB) *ClassTemplateRepository {
	return &ClassTemplateRepository{db: db}
}

// Create creates a new class template
func (r *ClassTemplateRepository) Create(template *domain.ClassTemplate) error {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	query := rebindQuery(`
		INSERT INTO class_templates (organization_id, name, description, workout_id, duration_minutes, default_capacity, color, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			template.OrganizationID,
			template.Name,
			template.Description,
			template.WorkoutID,
			template.DurationMinutes,
			template.DefaultCapacity,
			template.Color,
			template.IsActive,
			template.CreatedAt,
			template.UpdatedAt,
		).Scan(&template.ID)
	}

	result, err := r.db.Exec(query,
		template.OrganizationID,
		template.Name,
		template.Description,
		template.WorkoutID,
		template.DurationMinutes,
		template.DefaultCapacity,
		template.Color,
		template.IsActive,
		template.CreatedAt,
		template.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create class template: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get class template ID: %w", err)
	}

	template.ID = id
	return nil
}

// GetByID retrieves a class template by ID with optional workout name
func (r *ClassTemplateRepository) GetByID(id int64) (*domain.ClassTemplate, error) {
	query := rebindQuery(`
		SELECT ct.id, ct.organization_id, ct.name, ct.description, ct.workout_id,
		       ct.duration_minutes, ct.default_capacity, ct.color, ct.is_active,
		       ct.created_at, ct.updated_at, w.name as workout_name
		FROM class_templates ct
		LEFT JOIN workouts w ON ct.workout_id = w.id
		WHERE ct.id = ?
	`)

	template := &domain.ClassTemplate{}
	var description sql.NullString
	var workoutID sql.NullInt64
	var workoutName sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&template.ID,
		&template.OrganizationID,
		&template.Name,
		&description,
		&workoutID,
		&template.DurationMinutes,
		&template.DefaultCapacity,
		&template.Color,
		&template.IsActive,
		&template.CreatedAt,
		&template.UpdatedAt,
		&workoutName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get class template: %w", err)
	}

	if description.Valid {
		template.Description = &description.String
	}
	if workoutID.Valid {
		template.WorkoutID = &workoutID.Int64
	}
	if workoutName.Valid {
		template.WorkoutName = &workoutName.String
	}

	return template, nil
}

// GetByOrganizationID retrieves all class templates for an organization
func (r *ClassTemplateRepository) GetByOrganizationID(orgID int64, includeInactive bool) ([]*domain.ClassTemplate, error) {
	var query string
	if includeInactive {
		query = rebindQuery(`
			SELECT ct.id, ct.organization_id, ct.name, ct.description, ct.workout_id,
			       ct.duration_minutes, ct.default_capacity, ct.color, ct.is_active,
			       ct.created_at, ct.updated_at, w.name as workout_name
			FROM class_templates ct
			LEFT JOIN workouts w ON ct.workout_id = w.id
			WHERE ct.organization_id = ?
			ORDER BY ct.name ASC
		`)
	} else {
		query = rebindQuery(`
			SELECT ct.id, ct.organization_id, ct.name, ct.description, ct.workout_id,
			       ct.duration_minutes, ct.default_capacity, ct.color, ct.is_active,
			       ct.created_at, ct.updated_at, w.name as workout_name
			FROM class_templates ct
			LEFT JOIN workouts w ON ct.workout_id = w.id
			WHERE ct.organization_id = ? AND ct.is_active = ?
			ORDER BY ct.name ASC
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
		return nil, fmt.Errorf("failed to list class templates: %w", err)
	}
	defer rows.Close()

	var templates []*domain.ClassTemplate
	for rows.Next() {
		template := &domain.ClassTemplate{}
		var description sql.NullString
		var workoutID sql.NullInt64
		var workoutName sql.NullString

		err := rows.Scan(
			&template.ID,
			&template.OrganizationID,
			&template.Name,
			&description,
			&workoutID,
			&template.DurationMinutes,
			&template.DefaultCapacity,
			&template.Color,
			&template.IsActive,
			&template.CreatedAt,
			&template.UpdatedAt,
			&workoutName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan class template: %w", err)
		}

		if description.Valid {
			template.Description = &description.String
		}
		if workoutID.Valid {
			template.WorkoutID = &workoutID.Int64
		}
		if workoutName.Valid {
			template.WorkoutName = &workoutName.String
		}

		templates = append(templates, template)
	}

	return templates, rows.Err()
}

// Update updates a class template
func (r *ClassTemplateRepository) Update(template *domain.ClassTemplate) error {
	template.UpdatedAt = time.Now()

	query := rebindQuery(`
		UPDATE class_templates
		SET name = ?, description = ?, workout_id = ?, duration_minutes = ?,
		    default_capacity = ?, color = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query,
		template.Name,
		template.Description,
		template.WorkoutID,
		template.DurationMinutes,
		template.DefaultCapacity,
		template.Color,
		template.IsActive,
		template.UpdatedAt,
		template.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update class template: %w", err)
	}

	return nil
}

// Delete deletes a class template
func (r *ClassTemplateRepository) Delete(id int64) error {
	query := rebindQuery(`DELETE FROM class_templates WHERE id = ?`)

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete class template: %w", err)
	}

	return nil
}
