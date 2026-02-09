package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// ClassPackageRepository implements domain.ClassPackageRepository
type ClassPackageRepository struct {
	db *sql.DB
}

// NewClassPackageRepository creates a new class package repository
func NewClassPackageRepository(db *sql.DB) *ClassPackageRepository {
	return &ClassPackageRepository{db: db}
}

// Create creates a new class package
func (r *ClassPackageRepository) Create(pkg *domain.ClassPackage) error {
	now := time.Now()
	pkg.CreatedAt = now
	pkg.UpdatedAt = now

	query := rebindQuery(`INSERT INTO class_packages (organization_id, name, description, credits, price_cents, validity_days, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			pkg.OrganizationID,
			pkg.Name,
			pkg.Description,
			pkg.Credits,
			pkg.PriceCents,
			pkg.ValidityDays,
			pkg.IsActive,
			pkg.CreatedAt,
			pkg.UpdatedAt,
		).Scan(&pkg.ID)
	}

	result, err := r.db.Exec(query,
		pkg.OrganizationID,
		pkg.Name,
		pkg.Description,
		pkg.Credits,
		pkg.PriceCents,
		pkg.ValidityDays,
		pkg.IsActive,
		pkg.CreatedAt,
		pkg.UpdatedAt,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	pkg.ID = id
	return nil
}

// GetByID retrieves a class package by ID
func (r *ClassPackageRepository) GetByID(id int64) (*domain.ClassPackage, error) {
	query := rebindQuery(`SELECT id, organization_id, name, description, credits, price_cents, validity_days, is_active, created_at, updated_at
		FROM class_packages WHERE id = ?`)

	pkg := &domain.ClassPackage{}
	err := r.db.QueryRow(query, id).Scan(
		&pkg.ID,
		&pkg.OrganizationID,
		&pkg.Name,
		&pkg.Description,
		&pkg.Credits,
		&pkg.PriceCents,
		&pkg.ValidityDays,
		&pkg.IsActive,
		&pkg.CreatedAt,
		&pkg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return pkg, nil
}

// GetByOrganizationID retrieves all packages for an organization
func (r *ClassPackageRepository) GetByOrganizationID(orgID int64, includeInactive bool) ([]*domain.ClassPackage, error) {
	var query string
	if includeInactive {
		query = rebindQuery(`SELECT id, organization_id, name, description, credits, price_cents, validity_days, is_active, created_at, updated_at
			FROM class_packages WHERE organization_id = ? ORDER BY credits`)
	} else {
		query = rebindQuery(`SELECT id, organization_id, name, description, credits, price_cents, validity_days, is_active, created_at, updated_at
			FROM class_packages WHERE organization_id = ? AND is_active = true ORDER BY credits`)
	}

	rows, err := r.db.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []*domain.ClassPackage
	for rows.Next() {
		pkg := &domain.ClassPackage{}
		err := rows.Scan(
			&pkg.ID,
			&pkg.OrganizationID,
			&pkg.Name,
			&pkg.Description,
			&pkg.Credits,
			&pkg.PriceCents,
			&pkg.ValidityDays,
			&pkg.IsActive,
			&pkg.CreatedAt,
			&pkg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}
	return packages, rows.Err()
}

// Update updates a class package
func (r *ClassPackageRepository) Update(pkg *domain.ClassPackage) error {
	pkg.UpdatedAt = time.Now()

	query := rebindQuery(`UPDATE class_packages SET name = ?, description = ?, credits = ?, price_cents = ?, validity_days = ?, is_active = ?, updated_at = ?
		WHERE id = ?`)

	_, err := r.db.Exec(query,
		pkg.Name,
		pkg.Description,
		pkg.Credits,
		pkg.PriceCents,
		pkg.ValidityDays,
		pkg.IsActive,
		pkg.UpdatedAt,
		pkg.ID,
	)
	return err
}

// Delete deletes a class package
func (r *ClassPackageRepository) Delete(id int64) error {
	query := rebindQuery(`DELETE FROM class_packages WHERE id = ?`)
	_, err := r.db.Exec(query, id)
	return err
}

// UserClassCreditsRepository implements domain.UserClassCreditsRepository
type UserClassCreditsRepository struct {
	db *sql.DB
}

// NewUserClassCreditsRepository creates a new user class credits repository
func NewUserClassCreditsRepository(db *sql.DB) *UserClassCreditsRepository {
	return &UserClassCreditsRepository{db: db}
}

// Create creates a new user class credits record
func (r *UserClassCreditsRepository) Create(credits *domain.UserClassCredits) error {
	now := time.Now()
	credits.CreatedAt = now
	credits.UpdatedAt = now
	if credits.PurchasedAt.IsZero() {
		credits.PurchasedAt = now
	}

	query := rebindQuery(`INSERT INTO user_class_credits (user_id, organization_id, package_id, credits_total, credits_used, purchased_at, expires_at, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			credits.UserID,
			credits.OrganizationID,
			credits.PackageID,
			credits.CreditsTotal,
			credits.CreditsUsed,
			credits.PurchasedAt,
			credits.ExpiresAt,
			credits.Notes,
			credits.CreatedAt,
			credits.UpdatedAt,
		).Scan(&credits.ID)
	}

	result, err := r.db.Exec(query,
		credits.UserID,
		credits.OrganizationID,
		credits.PackageID,
		credits.CreditsTotal,
		credits.CreditsUsed,
		credits.PurchasedAt,
		credits.ExpiresAt,
		credits.Notes,
		credits.CreatedAt,
		credits.UpdatedAt,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	credits.ID = id
	return nil
}

// GetByID retrieves a credits record by ID
func (r *UserClassCreditsRepository) GetByID(id int64) (*domain.UserClassCredits, error) {
	query := rebindQuery(`SELECT ucc.id, ucc.user_id, ucc.organization_id, ucc.package_id, ucc.credits_total, ucc.credits_used, ucc.purchased_at, ucc.expires_at, ucc.notes, ucc.created_at, ucc.updated_at,
		cp.name as package_name
		FROM user_class_credits ucc
		LEFT JOIN class_packages cp ON ucc.package_id = cp.id
		WHERE ucc.id = ?`)

	credits := &domain.UserClassCredits{}
	err := r.db.QueryRow(query, id).Scan(
		&credits.ID,
		&credits.UserID,
		&credits.OrganizationID,
		&credits.PackageID,
		&credits.CreditsTotal,
		&credits.CreditsUsed,
		&credits.PurchasedAt,
		&credits.ExpiresAt,
		&credits.Notes,
		&credits.CreatedAt,
		&credits.UpdatedAt,
		&credits.PackageName,
	)
	if err != nil {
		return nil, err
	}
	credits.CreditsRemaining = credits.CreditsTotal - credits.CreditsUsed
	return credits, nil
}

// GetByUserID retrieves all credits for a user
func (r *UserClassCreditsRepository) GetByUserID(userID int64) ([]*domain.UserClassCredits, error) {
	query := rebindQuery(`SELECT ucc.id, ucc.user_id, ucc.organization_id, ucc.package_id, ucc.credits_total, ucc.credits_used, ucc.purchased_at, ucc.expires_at, ucc.notes, ucc.created_at, ucc.updated_at,
		cp.name as package_name
		FROM user_class_credits ucc
		LEFT JOIN class_packages cp ON ucc.package_id = cp.id
		WHERE ucc.user_id = ? ORDER BY ucc.purchased_at DESC`)

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creditsList []*domain.UserClassCredits
	for rows.Next() {
		credits := &domain.UserClassCredits{}
		err := rows.Scan(
			&credits.ID,
			&credits.UserID,
			&credits.OrganizationID,
			&credits.PackageID,
			&credits.CreditsTotal,
			&credits.CreditsUsed,
			&credits.PurchasedAt,
			&credits.ExpiresAt,
			&credits.Notes,
			&credits.CreatedAt,
			&credits.UpdatedAt,
			&credits.PackageName,
		)
		if err != nil {
			return nil, err
		}
		credits.CreditsRemaining = credits.CreditsTotal - credits.CreditsUsed
		creditsList = append(creditsList, credits)
	}
	return creditsList, rows.Err()
}

// GetByUserAndOrganization retrieves all credits for a user in an organization
func (r *UserClassCreditsRepository) GetByUserAndOrganization(userID, orgID int64) ([]*domain.UserClassCredits, error) {
	var query string
	if currentDriver == "postgres" {
		query = rebindQuery(`SELECT ucc.id, ucc.user_id, ucc.organization_id, ucc.package_id, ucc.credits_total, ucc.credits_used, ucc.purchased_at, ucc.expires_at, ucc.notes, ucc.created_at, ucc.updated_at,
			cp.name as package_name
			FROM user_class_credits ucc
			LEFT JOIN class_packages cp ON ucc.package_id = cp.id
			WHERE ucc.user_id = ? AND ucc.organization_id = ? ORDER BY ucc.expires_at NULLS LAST, ucc.purchased_at`)
	} else {
		query = rebindQuery(`SELECT ucc.id, ucc.user_id, ucc.organization_id, ucc.package_id, ucc.credits_total, ucc.credits_used, ucc.purchased_at, ucc.expires_at, ucc.notes, ucc.created_at, ucc.updated_at,
			cp.name as package_name
			FROM user_class_credits ucc
			LEFT JOIN class_packages cp ON ucc.package_id = cp.id
			WHERE ucc.user_id = ? AND ucc.organization_id = ? ORDER BY CASE WHEN ucc.expires_at IS NULL THEN 1 ELSE 0 END, ucc.expires_at, ucc.purchased_at`)
	}

	rows, err := r.db.Query(query, userID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creditsList []*domain.UserClassCredits
	for rows.Next() {
		credits := &domain.UserClassCredits{}
		err := rows.Scan(
			&credits.ID,
			&credits.UserID,
			&credits.OrganizationID,
			&credits.PackageID,
			&credits.CreditsTotal,
			&credits.CreditsUsed,
			&credits.PurchasedAt,
			&credits.ExpiresAt,
			&credits.Notes,
			&credits.CreatedAt,
			&credits.UpdatedAt,
			&credits.PackageName,
		)
		if err != nil {
			return nil, err
		}
		credits.CreditsRemaining = credits.CreditsTotal - credits.CreditsUsed
		creditsList = append(creditsList, credits)
	}
	return creditsList, rows.Err()
}

// GetAvailableCredits returns the total available (non-expired, not fully used) credits for a user in an organization
func (r *UserClassCreditsRepository) GetAvailableCredits(userID, orgID int64) (int, error) {
	var query string
	if currentDriver == "postgres" {
		query = rebindQuery(`SELECT COALESCE(SUM(credits_total - credits_used), 0)
			FROM user_class_credits
			WHERE user_id = ? AND organization_id = ?
			AND credits_used < credits_total
			AND (expires_at IS NULL OR expires_at > NOW())`)
	} else {
		query = rebindQuery(`SELECT COALESCE(SUM(credits_total - credits_used), 0)
			FROM user_class_credits
			WHERE user_id = ? AND organization_id = ?
			AND credits_used < credits_total
			AND (expires_at IS NULL OR expires_at > datetime('now'))`)
	}

	var total int
	err := r.db.QueryRow(query, userID, orgID).Scan(&total)
	return total, err
}

// UseCredit deducts 1 credit from the oldest non-expired balance
func (r *UserClassCreditsRepository) UseCredit(userID, orgID int64) error {
	// Find the oldest non-expired credits record with available credits
	var creditID int64
	var query string
	if currentDriver == "postgres" {
		query = rebindQuery(`SELECT id FROM user_class_credits
			WHERE user_id = ? AND organization_id = ?
			AND credits_used < credits_total
			AND (expires_at IS NULL OR expires_at > NOW())
			ORDER BY expires_at NULLS LAST, purchased_at
			LIMIT 1`)
	} else {
		query = rebindQuery(`SELECT id FROM user_class_credits
			WHERE user_id = ? AND organization_id = ?
			AND credits_used < credits_total
			AND (expires_at IS NULL OR expires_at > datetime('now'))
			ORDER BY CASE WHEN expires_at IS NULL THEN 1 ELSE 0 END, expires_at, purchased_at
			LIMIT 1`)
	}

	err := r.db.QueryRow(query, userID, orgID).Scan(&creditID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no available credits for user %d in organization %d", userID, orgID)
	}
	if err != nil {
		return err
	}

	// Increment credits_used
	updateQuery := rebindQuery(`UPDATE user_class_credits SET credits_used = credits_used + 1, updated_at = ? WHERE id = ?`)

	_, err = r.db.Exec(updateQuery, time.Now(), creditID)
	return err
}

// RefundCredit restores 1 credit to the user's balance (opposite of UseCredit)
// It finds the most recently used credits record with credits_used > 0 and decrements it
func (r *UserClassCreditsRepository) RefundCredit(userID, orgID int64) error {
	// Find the most recently updated credits record that has credits_used > 0
	var creditID int64
	query := rebindQuery(`SELECT id FROM user_class_credits
		WHERE user_id = ? AND organization_id = ?
		AND credits_used > 0
		ORDER BY updated_at DESC
		LIMIT 1`)

	err := r.db.QueryRow(query, userID, orgID).Scan(&creditID)
	if err == sql.ErrNoRows {
		// No credits to refund - this is not an error, just nothing to do
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to find credits to refund: %w", err)
	}

	// Decrement credits_used
	updateQuery := rebindQuery(`UPDATE user_class_credits SET credits_used = credits_used - 1, updated_at = ? WHERE id = ?`)

	_, err = r.db.Exec(updateQuery, time.Now(), creditID)
	if err != nil {
		return fmt.Errorf("failed to refund credit: %w", err)
	}

	return nil
}

// GetExpiringSoon retrieves credits expiring within daysAhead days
func (r *UserClassCreditsRepository) GetExpiringSoon(daysAhead int) ([]*domain.UserClassCredits, error) {
	expirationDate := time.Now().AddDate(0, 0, daysAhead)

	var query string
	if currentDriver == "postgres" {
		query = rebindQuery(`SELECT ucc.id, ucc.user_id, ucc.organization_id, ucc.package_id, ucc.credits_total, ucc.credits_used, ucc.purchased_at, ucc.expires_at, ucc.notes, ucc.created_at, ucc.updated_at,
			cp.name as package_name, u.name as user_name, u.email as user_email
			FROM user_class_credits ucc
			LEFT JOIN class_packages cp ON ucc.package_id = cp.id
			JOIN users u ON ucc.user_id = u.id
			WHERE ucc.credits_used < ucc.credits_total
			AND ucc.expires_at IS NOT NULL AND ucc.expires_at <= ? AND ucc.expires_at > NOW()
			ORDER BY ucc.expires_at`)
	} else {
		query = rebindQuery(`SELECT ucc.id, ucc.user_id, ucc.organization_id, ucc.package_id, ucc.credits_total, ucc.credits_used, ucc.purchased_at, ucc.expires_at, ucc.notes, ucc.created_at, ucc.updated_at,
			cp.name as package_name, u.name as user_name, u.email as user_email
			FROM user_class_credits ucc
			LEFT JOIN class_packages cp ON ucc.package_id = cp.id
			JOIN users u ON ucc.user_id = u.id
			WHERE ucc.credits_used < ucc.credits_total
			AND ucc.expires_at IS NOT NULL AND ucc.expires_at <= ? AND ucc.expires_at > datetime('now')
			ORDER BY ucc.expires_at`)
	}

	rows, err := r.db.Query(query, expirationDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creditsList []*domain.UserClassCredits
	for rows.Next() {
		credits := &domain.UserClassCredits{}
		err := rows.Scan(
			&credits.ID,
			&credits.UserID,
			&credits.OrganizationID,
			&credits.PackageID,
			&credits.CreditsTotal,
			&credits.CreditsUsed,
			&credits.PurchasedAt,
			&credits.ExpiresAt,
			&credits.Notes,
			&credits.CreatedAt,
			&credits.UpdatedAt,
			&credits.PackageName,
			&credits.UserName,
			&credits.UserEmail,
		)
		if err != nil {
			return nil, err
		}
		credits.CreditsRemaining = credits.CreditsTotal - credits.CreditsUsed
		creditsList = append(creditsList, credits)
	}
	return creditsList, rows.Err()
}

// Update updates a user class credits record
func (r *UserClassCreditsRepository) Update(credits *domain.UserClassCredits) error {
	credits.UpdatedAt = time.Now()

	query := rebindQuery(`UPDATE user_class_credits SET credits_total = ?, credits_used = ?, expires_at = ?, notes = ?, updated_at = ? WHERE id = ?`)

	_, err := r.db.Exec(query,
		credits.CreditsTotal,
		credits.CreditsUsed,
		credits.ExpiresAt,
		credits.Notes,
		credits.UpdatedAt,
		credits.ID,
	)
	return err
}
