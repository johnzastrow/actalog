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
// RESTRICT behavior: Prevents deletion if users are assigned
func (r *OrganizationRepository) Delete(id int64) error {
	// Check if any users are assigned to this organization
	users, err := r.GetOrganizationUsers(id)
	if err != nil {
		return fmt.Errorf("failed to check organization users: %w", err)
	}

	if len(users) > 0 {
		return fmt.Errorf("cannot delete organization: %d user(s) are still assigned", len(users))
	}

	query := rebindQuery(`DELETE FROM organizations WHERE id = ?`)

	_, err = r.db.Exec(query, id)
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

// AddUserToOrganization creates a user-organization relationship
func (r *OrganizationRepository) AddUserToOrganization(userID, orgID int64, role string) error {
	now := time.Now()

	query := rebindQuery(`
		INSERT INTO user_organizations (user_id, organization_id, role, joined_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`)

	_, err := r.db.Exec(query, userID, orgID, role, now, now, now)
	if err != nil {
		return fmt.Errorf("failed to add user to organization: %w", err)
	}

	return nil
}

// RemoveUserFromOrganization removes a user-organization relationship
func (r *OrganizationRepository) RemoveUserFromOrganization(userID, orgID int64) error {
	query := rebindQuery(`DELETE FROM user_organizations WHERE user_id = ? AND organization_id = ?`)

	_, err := r.db.Exec(query, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to remove user from organization: %w", err)
	}

	return nil
}

// GetUserOrganizations retrieves all organizations for a user
func (r *OrganizationRepository) GetUserOrganizations(userID int64) ([]*domain.Organization, error) {
	query := rebindQuery(`
		SELECT o.id, o.name, o.description, o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN user_organizations uo ON o.id = uo.organization_id
		WHERE uo.user_id = ?
		ORDER BY o.name ASC
	`)

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}
	defer rows.Close()

	var orgs []*domain.Organization
	for rows.Next() {
		org := &domain.Organization{}
		var description sql.NullString

		err := rows.Scan(&org.ID, &org.Name, &description, &org.CreatedAt, &org.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}

		if description.Valid {
			org.Description = &description.String
		}

		orgs = append(orgs, org)
	}

	return orgs, rows.Err()
}

// GetOrganizationUsers retrieves all users in an organization
func (r *OrganizationRepository) GetOrganizationUsers(orgID int64) ([]*domain.User, error) {
	query := rebindQuery(`
		SELECT u.id, u.email, u.name, u.profile_image, u.role,
		       u.email_verified, u.email_verified_at,
		       u.created_at, u.updated_at, u.last_login_at
		FROM users u
		INNER JOIN user_organizations uo ON u.id = uo.user_id
		WHERE uo.organization_id = ?
		ORDER BY u.name ASC
	`)

	rows, err := r.db.Query(query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		var profileImage sql.NullString
		var emailVerifiedAt, lastLoginAt sql.NullTime

		err := rows.Scan(
			&user.ID, &user.Email, &user.Name, &profileImage, &user.Role,
			&user.EmailVerified, &emailVerifiedAt,
			&user.CreatedAt, &user.UpdatedAt, &lastLoginAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if profileImage.Valid {
			user.ProfileImage = &profileImage.String
		}
		if emailVerifiedAt.Valid {
			user.EmailVerifiedAt = &emailVerifiedAt.Time
		}
		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}

		users = append(users, user)
	}

	return users, rows.Err()
}

// IsUserInOrganization checks if a user belongs to an organization
func (r *OrganizationRepository) IsUserInOrganization(userID, orgID int64) (bool, error) {
	query := rebindQuery(`
		SELECT EXISTS(
			SELECT 1 FROM user_organizations
			WHERE user_id = ? AND organization_id = ?
		)
	`)

	var exists bool
	err := r.db.QueryRow(query, userID, orgID).Scan(&exists)
	return exists, err
}

// GetUserOrganizationIDs returns all organization IDs for a user
func (r *OrganizationRepository) GetUserOrganizationIDs(userID int64) ([]int64, error) {
	query := rebindQuery(`
		SELECT organization_id
		FROM user_organizations
		WHERE user_id = ?
		ORDER BY joined_at ASC
	`)

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organization IDs: %w", err)
	}
	defer rows.Close()

	var orgIDs []int64
	for rows.Next() {
		var orgID int64
		if err := rows.Scan(&orgID); err != nil {
			return nil, fmt.Errorf("failed to scan organization ID: %w", err)
		}
		orgIDs = append(orgIDs, orgID)
	}

	return orgIDs, rows.Err()
}
