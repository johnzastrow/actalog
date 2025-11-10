package repository

import (
	"database/sql"
	"fmt"

	"github.com/johnzastrow/actalog/internal/domain"
)

// SQLiteRefreshTokenRepository implements RefreshTokenRepository for SQLite
type SQLiteRefreshTokenRepository struct {
	db *sql.DB
}

// NewSQLiteRefreshTokenRepository creates a new SQLite refresh token repository
func NewSQLiteRefreshTokenRepository(db *sql.DB) domain.RefreshTokenRepository {
	return &SQLiteRefreshTokenRepository{db: db}
}

// Create creates a new refresh token
func (r *SQLiteRefreshTokenRepository) Create(token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at, device_info)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
		token.DeviceInfo,
	)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	token.ID = id
	return nil
}

// GetByToken retrieves a refresh token by its token string
func (r *SQLiteRefreshTokenRepository) GetByToken(tokenStr string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, revoked_at, device_info
		FROM refresh_tokens
		WHERE token = ? AND revoked_at IS NULL AND expires_at > datetime('now')
	`

	token := &domain.RefreshToken{}
	err := r.db.QueryRow(query, tokenStr).Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.RevokedAt,
		&token.DeviceInfo,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return token, nil
}

// GetByUserID retrieves all refresh tokens for a user
func (r *SQLiteRefreshTokenRepository) GetByUserID(userID int64) ([]*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, revoked_at, device_info
		FROM refresh_tokens
		WHERE user_id = ? AND revoked_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query refresh tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*domain.RefreshToken
	for rows.Next() {
		token := &domain.RefreshToken{}
		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.Token,
			&token.ExpiresAt,
			&token.CreatedAt,
			&token.RevokedAt,
			&token.DeviceInfo,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan refresh token: %w", err)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// Revoke revokes a specific refresh token
func (r *SQLiteRefreshTokenRepository) Revoke(tokenID int64) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = datetime('now')
		WHERE id = ?
	`

	_, err := r.db.Exec(query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// RevokeAllForUser revokes all refresh tokens for a user
func (r *SQLiteRefreshTokenRepository) RevokeAllForUser(userID int64) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = datetime('now')
		WHERE user_id = ? AND revoked_at IS NULL
	`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all refresh tokens: %w", err)
	}

	return nil
}

// DeleteExpired deletes all expired refresh tokens
func (r *SQLiteRefreshTokenRepository) DeleteExpired() error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < datetime('now')
	`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	return nil
}

// Delete deletes a specific refresh token
func (r *SQLiteRefreshTokenRepository) Delete(tokenID int64) error {
	query := `DELETE FROM refresh_tokens WHERE id = ?`

	_, err := r.db.Exec(query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}
