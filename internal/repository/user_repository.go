package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// SQLiteUserRepository implements UserRepository using SQLite
type SQLiteUserRepository struct {
	db *sql.DB
}

// NewSQLiteUserRepository creates a new SQLite user repository
func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

// Create creates a new user
func (r *SQLiteUserRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(
		query,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	return nil
}

// GetByID retrieves a user by ID
func (r *SQLiteUserRepository) GetByID(id int64) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, profile_image, role,
		       created_at, updated_at, last_login_at
		FROM users
		WHERE id = ?
	`

	user := &domain.User{}
	var lastLoginAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.ProfileImage,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *SQLiteUserRepository) GetByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, profile_image, role,
		       created_at, updated_at, last_login_at
		FROM users
		WHERE email = ?
	`

	user := &domain.User{}
	var lastLoginAt sql.NullTime

	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.ProfileImage,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// GetByResetToken retrieves a user by reset token
// NOTE: This method is currently not functional as reset_token columns don't exist in the database.
// The system should use the separate password_resets table instead.
func (r *SQLiteUserRepository) GetByResetToken(token string) (*domain.User, error) {
	// This functionality is not implemented - reset tokens should use a separate table
	return nil, nil
}

// GetByVerificationToken retrieves a user by verification token
// NOTE: This method is currently not functional as verification_token columns don't exist in the database.
// The system should use the separate email_verification_tokens table instead.
func (r *SQLiteUserRepository) GetByVerificationToken(token string) (*domain.User, error) {
	// This functionality is not implemented - verification tokens should use a separate table
	return nil, nil
}

// Update updates a user
func (r *SQLiteUserRepository) Update(user *domain.User) error {
	query := `
		UPDATE users
		SET email = ?, name = ?, profile_image = ?, role = ?,
		    updated_at = ?, last_login_at = ?, password_hash = ?,
		    email_verified = ?, email_verified_at = ?
		WHERE id = ?
	`

	var lastLoginAt interface{}
	if user.LastLoginAt != nil {
		lastLoginAt = *user.LastLoginAt
	}

	var emailVerifiedAt interface{}
	if user.EmailVerifiedAt != nil {
		emailVerifiedAt = *user.EmailVerifiedAt
	}

	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		user.Email,
		user.Name,
		user.ProfileImage,
		user.Role,
		user.UpdatedAt,
		lastLoginAt,
		user.PasswordHash,
		user.EmailVerified,
		emailVerifiedAt,
		user.ID,
	)

	return err
}

// UpdatePassword updates only the password for a user
func (r *SQLiteUserRepository) UpdatePassword(userID int64, hashedPassword string) error {
	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, hashedPassword, time.Now(), userID)
	return err
}

// Delete deletes a user
func (r *SQLiteUserRepository) Delete(id int64) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// List retrieves a list of users with pagination
func (r *SQLiteUserRepository) List(limit, offset int) ([]*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, profile_image, role,
		       created_at, updated_at, last_login_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		var lastLoginAt sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.Name,
			&user.ProfileImage,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
			&lastLoginAt,
		)
		if err != nil {
			return nil, err
		}

		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}

		users = append(users, user)
	}

	return users, rows.Err()
}

// Count returns the total number of users
func (r *SQLiteUserRepository) Count() (int64, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int64
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}
