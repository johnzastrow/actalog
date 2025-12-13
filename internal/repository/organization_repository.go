package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

type OrganizationRepository struct {
	db *sql.DB
}

func NewOrganizationRepository(db *sql.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// Create creates a new organization
func (r *OrganizationRepository) Create(org *domain.Organization) error {
	org.CreatedAt = time.Now()
	org.UpdatedAt = time.Now()

	query := rebindQuery(`
		INSERT INTO organizations (name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query, org.Name, org.Description, org.CreatedAt, org.UpdatedAt).Scan(&org.ID)
	}

	result, err := r.db.Exec(query, org.Name, org.Description, org.CreatedAt, org.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get organization ID: %w", err)
	}

	org.ID = id
	return nil
}

// GetByID retrieves an organization by ID
func (r *OrganizationRepository) GetByID(id int64) (*domain.Organization, error) {
	query := rebindQuery(`
		SELECT id, name, description, created_at, updated_at
		FROM organizations
		WHERE id = ?
	`)

	org := &domain.Organization{}
	var description sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&org.ID,
		&org.Name,
		&description,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if description.Valid {
		org.Description = &description.String
	}

	return org, nil
}

// GetByName retrieves an organization by name
func (r *OrganizationRepository) GetByName(name string) (*domain.Organization, error) {
	query := rebindQuery(`
		SELECT id, name, description, created_at, updated_at
		FROM organizations
		WHERE name = ?
	`)

	org := &domain.Organization{}
	var description sql.NullString

	err := r.db.QueryRow(query, name).Scan(
		&org.ID,
		&org.Name,
		&description,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if description.Valid {
		org.Description = &description.String
	}

	return org, nil
}

// List retrieves organizations with pagination
func (r *OrganizationRepository) List(limit, offset int) ([]*domain.Organization, int64, error) {
	query := rebindQuery(`
		SELECT id, name, description, created_at, updated_at
		FROM organizations
		ORDER BY name ASC
		LIMIT ? OFFSET ?
	`)

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list organizations: %w", err)
	}
	defer rows.Close()

	var orgs []*domain.Organization
	for rows.Next() {
		org := &domain.Organization{}
		var description sql.NullString

		err := rows.Scan(&org.ID, &org.Name, &description, &org.CreatedAt, &org.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan organization: %w", err)
		}

		if description.Valid {
			org.Description = &description.String
		}

		orgs = append(orgs, org)
	}

	// Get total count
	count, err := r.Count()
	if err != nil {
		return nil, 0, err
	}

	return orgs, count, rows.Err()
}

// Update updates an organization
func (r *OrganizationRepository) Update(org *domain.Organization) error {
	org.UpdatedAt = time.Now()

	query := rebindQuery(`
		UPDATE organizations
		SET name = ?, description = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query, org.Name, org.Description, org.UpdatedAt, org.ID)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	return nil
}

// Delete deletes an organization
func (r *OrganizationRepository) Delete(id int64) error {
	query := rebindQuery(`DELETE FROM organizations WHERE id = ?`)

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return nil
}

// Count returns total number of organizations
func (r *OrganizationRepository) Count() (int64, error) {
	query := `SELECT COUNT(*) FROM organizations`

	var count int64
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}
