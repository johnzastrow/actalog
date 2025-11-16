package repository

import (
	"database/sql"
	"errors"
	"fmt"
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
		       created_at, updated_at, last_login_at, email_verified, email_verified_at,
		       failed_login_attempts, locked_at, locked_until,
		       account_disabled, disabled_at, disabled_by_user_id
		FROM users
		WHERE id = ?
	`

	user := &domain.User{}
	var lastLoginAt, emailVerifiedAt, lockedAt, lockedUntil, disabledAt sql.NullTime
	var disabledByUserID sql.NullInt64

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
		&user.EmailVerified,
		&emailVerifiedAt,
		&user.FailedLoginAttempts,
		&lockedAt,
		&lockedUntil,
		&user.AccountDisabled,
		&disabledAt,
		&disabledByUserID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Handle nullable fields
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = &emailVerifiedAt.Time
	}
	if lockedAt.Valid {
		user.LockedAt = &lockedAt.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}
	if disabledAt.Valid {
		user.DisabledAt = &disabledAt.Time
	}
	if disabledByUserID.Valid {
		user.DisabledByUserID = &disabledByUserID.Int64
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *SQLiteUserRepository) GetByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, profile_image, role,
		       created_at, updated_at, last_login_at, email_verified, email_verified_at,
		       failed_login_attempts, locked_at, locked_until,
		       account_disabled, disabled_at, disabled_by_user_id
		FROM users
		WHERE email = ?
	`

	user := &domain.User{}
	var lastLoginAt, emailVerifiedAt, lockedAt, lockedUntil, disabledAt sql.NullTime
	var disabledByUserID sql.NullInt64

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
		&user.EmailVerified,
		&emailVerifiedAt,
		&user.FailedLoginAttempts,
		&lockedAt,
		&lockedUntil,
		&user.AccountDisabled,
		&disabledAt,
		&disabledByUserID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Handle nullable fields
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = &emailVerifiedAt.Time
	}
	if lockedAt.Valid {
		user.LockedAt = &lockedAt.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}
	if disabledAt.Valid {
		user.DisabledAt = &disabledAt.Time
	}
	if disabledByUserID.Valid {
		user.DisabledByUserID = &disabledByUserID.Int64
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

	var profileImage interface{}
	if user.ProfileImage != nil {
		profileImage = *user.ProfileImage
	}

	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		user.Email,
		user.Name,
		profileImage,
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

// Account Security Methods

// IncrementFailedAttempts increments the failed login attempts counter
func (r *SQLiteUserRepository) IncrementFailedAttempts(userID int64) error {
	query := `UPDATE users SET failed_login_attempts = failed_login_attempts + 1, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to increment failed attempts: %w", err)
	}
	return nil
}

// ResetFailedAttempts resets the failed login attempts counter to zero
func (r *SQLiteUserRepository) ResetFailedAttempts(userID int64) error {
	query := `UPDATE users SET failed_login_attempts = 0, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to reset failed attempts: %w", err)
	}
	return nil
}

// LockAccount locks a user account for a specified duration
func (r *SQLiteUserRepository) LockAccount(userID int64, lockDuration time.Duration) error {
	now := time.Now()
	lockedUntil := now.Add(lockDuration)

	query := `UPDATE users
		SET locked_at = ?, locked_until = ?, updated_at = ?
		WHERE id = ?`

	_, err := r.db.Exec(query, now, lockedUntil, now, userID)
	if err != nil {
		return fmt.Errorf("failed to lock account: %w", err)
	}
	return nil
}

// UnlockAccount unlocks a user account and resets failed attempts
func (r *SQLiteUserRepository) UnlockAccount(userID int64) error {
	query := `UPDATE users
		SET locked_at = NULL, locked_until = NULL, failed_login_attempts = 0, updated_at = ?
		WHERE id = ?`

	_, err := r.db.Exec(query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to unlock account: %w", err)
	}
	return nil
}

// IsAccountLocked checks if an account is currently locked
// Returns: locked (bool), unlock time (*time.Time), error
func (r *SQLiteUserRepository) IsAccountLocked(userID int64) (bool, *time.Time, error) {
	query := `SELECT locked_at, locked_until FROM users WHERE id = ?`

	var lockedAt, lockedUntil *time.Time
	err := r.db.QueryRow(query, userID).Scan(&lockedAt, &lockedUntil)
	if err == sql.ErrNoRows {
		return false, nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return false, nil, fmt.Errorf("failed to check lock status: %w", err)
	}

	// If locked_at is NULL, account is not locked
	if lockedAt == nil || lockedUntil == nil {
		return false, nil, nil
	}

	// Check if lock has expired
	if time.Now().After(*lockedUntil) {
		// Lock has expired, auto-unlock
		err = r.UnlockAccount(userID)
		if err != nil {
			return false, nil, fmt.Errorf("failed to auto-unlock account: %w", err)
		}
		return false, nil, nil
	}

	// Account is still locked
	return true, lockedUntil, nil
}

// DisableAccount permanently disables a user account (admin action)
func (r *SQLiteUserRepository) DisableAccount(userID int64, disabledBy int64) error {
	now := time.Now()
	query := `UPDATE users
		SET account_disabled = 1, disabled_at = ?, disabled_by_user_id = ?, updated_at = ?
		WHERE id = ?`

	_, err := r.db.Exec(query, now, disabledBy, now, userID)
	if err != nil {
		return fmt.Errorf("failed to disable account: %w", err)
	}
	return nil
}

// EnableAccount re-enables a disabled user account (admin action)
func (r *SQLiteUserRepository) EnableAccount(userID int64) error {
	query := `UPDATE users
		SET account_disabled = 0, disabled_at = NULL, disabled_by_user_id = NULL, updated_at = ?
		WHERE id = ?`

	_, err := r.db.Exec(query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to enable account: %w", err)
	}
	return nil
}
