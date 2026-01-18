package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

type GymLocationRepository struct {
	db *sql.DB
}

func NewGymLocationRepository(db *sql.DB) *GymLocationRepository {
	return &GymLocationRepository{db: db}
}

// Create creates a new gym location
func (r *GymLocationRepository) Create(location *domain.GymLocation) error {
	location.CreatedAt = time.Now()
	location.UpdatedAt = time.Now()

	query := rebindQuery(`
		INSERT INTO gym_locations (organization_id, name, description, address, capacity, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			location.OrganizationID,
			location.Name,
			location.Description,
			location.Address,
			location.Capacity,
			location.IsActive,
			location.CreatedAt,
			location.UpdatedAt,
		).Scan(&location.ID)
	}

	result, err := r.db.Exec(query,
		location.OrganizationID,
		location.Name,
		location.Description,
		location.Address,
		location.Capacity,
		location.IsActive,
		location.CreatedAt,
		location.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create gym location: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get gym location ID: %w", err)
	}

	location.ID = id
	return nil
}

// GetByID retrieves a gym location by ID
func (r *GymLocationRepository) GetByID(id int64) (*domain.GymLocation, error) {
	query := rebindQuery(`
		SELECT id, organization_id, name, description, address, capacity, is_active, created_at, updated_at
		FROM gym_locations
		WHERE id = ?
	`)

	location := &domain.GymLocation{}
	var description, address sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&location.ID,
		&location.OrganizationID,
		&location.Name,
		&description,
		&address,
		&location.Capacity,
		&location.IsActive,
		&location.CreatedAt,
		&location.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get gym location: %w", err)
	}

	if description.Valid {
		location.Description = &description.String
	}
	if address.Valid {
		location.Address = &address.String
	}

	return location, nil
}

// GetByOrganizationID retrieves all gym locations for an organization
func (r *GymLocationRepository) GetByOrganizationID(orgID int64, includeInactive bool) ([]*domain.GymLocation, error) {
	var query string
	if includeInactive {
		query = rebindQuery(`
			SELECT id, organization_id, name, description, address, capacity, is_active, created_at, updated_at
			FROM gym_locations
			WHERE organization_id = ?
			ORDER BY name ASC
		`)
	} else {
		query = rebindQuery(`
			SELECT id, organization_id, name, description, address, capacity, is_active, created_at, updated_at
			FROM gym_locations
			WHERE organization_id = ? AND is_active = ?
			ORDER BY name ASC
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
		return nil, fmt.Errorf("failed to list gym locations: %w", err)
	}
	defer rows.Close()

	var locations []*domain.GymLocation
	for rows.Next() {
		location := &domain.GymLocation{}
		var description, address sql.NullString

		err := rows.Scan(
			&location.ID,
			&location.OrganizationID,
			&location.Name,
			&description,
			&address,
			&location.Capacity,
			&location.IsActive,
			&location.CreatedAt,
			&location.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan gym location: %w", err)
		}

		if description.Valid {
			location.Description = &description.String
		}
		if address.Valid {
			location.Address = &address.String
		}

		locations = append(locations, location)
	}

	return locations, rows.Err()
}

// Update updates a gym location
func (r *GymLocationRepository) Update(location *domain.GymLocation) error {
	location.UpdatedAt = time.Now()

	query := rebindQuery(`
		UPDATE gym_locations
		SET name = ?, description = ?, address = ?, capacity = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query,
		location.Name,
		location.Description,
		location.Address,
		location.Capacity,
		location.IsActive,
		location.UpdatedAt,
		location.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update gym location: %w", err)
	}

	return nil
}

// Delete deletes a gym location
func (r *GymLocationRepository) Delete(id int64) error {
	query := rebindQuery(`DELETE FROM gym_locations WHERE id = ?`)

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete gym location: %w", err)
	}

	return nil
}
