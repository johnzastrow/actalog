package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// WODRepository implements domain.WODRepository
type WODRepository struct {
	db *sql.DB
}

// NewWODRepository creates a new WOD repository
func NewWODRepository(db *sql.DB) *WODRepository {
	return &WODRepository{db: db}
}

// Create creates a new custom WOD
func (r *WODRepository) Create(wod *domain.WOD) error {
	wod.CreatedAt = time.Now()
	wod.UpdatedAt = time.Now()

	query := `INSERT INTO wods (name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		wod.Name,
		wod.Source,
		wod.Type,
		wod.Regime,
		wod.ScoreType,
		wod.Description,
		wod.URL,
		wod.Notes,
		wod.IsStandard,
		wod.CreatedBy,
		wod.CreatedAt,
		wod.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create wod: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get wod ID: %w", err)
	}

	wod.ID = id
	return nil
}

// GetByID retrieves a WOD by ID
func (r *WODRepository) GetByID(id int64) (*domain.WOD, error) {
	query := `SELECT id, name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at
	          FROM wods WHERE id = ?`

	wod := &domain.WOD{}
	var url, notes sql.NullString
	var createdBy sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&wod.ID,
		&wod.Name,
		&wod.Source,
		&wod.Type,
		&wod.Regime,
		&wod.ScoreType,
		&wod.Description,
		&url,
		&notes,
		&wod.IsStandard,
		&createdBy,
		&wod.CreatedAt,
		&wod.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get wod: %w", err)
	}

	if url.Valid {
		wod.URL = &url.String
	}
	if notes.Valid {
		wod.Notes = &notes.String
	}
	if createdBy.Valid {
		wod.CreatedBy = &createdBy.Int64
	}

	return wod, nil
}

// GetByName retrieves a WOD by name
func (r *WODRepository) GetByName(name string) (*domain.WOD, error) {
	query := `SELECT id, name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at
	          FROM wods WHERE name = ?`

	wod := &domain.WOD{}
	var url, notes sql.NullString
	var createdBy sql.NullInt64

	err := r.db.QueryRow(query, name).Scan(
		&wod.ID,
		&wod.Name,
		&wod.Source,
		&wod.Type,
		&wod.Regime,
		&wod.ScoreType,
		&wod.Description,
		&url,
		&notes,
		&wod.IsStandard,
		&createdBy,
		&wod.CreatedAt,
		&wod.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get wod by name: %w", err)
	}

	if url.Valid {
		wod.URL = &url.String
	}
	if notes.Valid {
		wod.Notes = &notes.String
	}
	if createdBy.Valid {
		wod.CreatedBy = &createdBy.Int64
	}

	return wod, nil
}

// List retrieves WODs with optional filtering, limit, and offset
func (r *WODRepository) List(filters map[string]interface{}, limit, offset int) ([]*domain.WOD, error) {
	query := `SELECT id, name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at
	          FROM wods WHERE 1=1`

	var args []interface{}

	// Build dynamic WHERE clause based on filters
	if filters != nil {
		if source, ok := filters["source"].(string); ok && source != "" {
			query += " AND source = ?"
			args = append(args, source)
		}
		if wodType, ok := filters["type"].(string); ok && wodType != "" {
			query += " AND type = ?"
			args = append(args, wodType)
		}
		if regime, ok := filters["regime"].(string); ok && regime != "" {
			query += " AND regime = ?"
			args = append(args, regime)
		}
		if scoreType, ok := filters["score_type"].(string); ok && scoreType != "" {
			query += " AND score_type = ?"
			args = append(args, scoreType)
		}
		if isStandard, ok := filters["is_standard"].(bool); ok {
			query += " AND is_standard = ?"
			args = append(args, isStandard)
		}
		if createdBy, ok := filters["created_by"].(int64); ok {
			query += " AND created_by = ?"
			args = append(args, createdBy)
		}
	}

	query += " ORDER BY is_standard DESC, name"

	// Add pagination
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list wods: %w", err)
	}
	defer rows.Close()

	return r.scanWODs(rows)
}

// ListStandard retrieves all standard (pre-seeded) WODs
func (r *WODRepository) ListStandard(limit, offset int) ([]*domain.WOD, error) {
	query := `SELECT id, name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at
	          FROM wods WHERE is_standard = 1 ORDER BY name`

	var args []interface{}

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list standard wods: %w", err)
	}
	defer rows.Close()

	return r.scanWODs(rows)
}

// ListByUser retrieves all custom WODs created by a specific user
func (r *WODRepository) ListByUser(userID int64, limit, offset int) ([]*domain.WOD, error) {
	query := `SELECT id, name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at
	          FROM wods WHERE created_by = ? ORDER BY name`

	var args []interface{}
	args = append(args, userID)

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list user wods: %w", err)
	}
	defer rows.Close()

	return r.scanWODs(rows)
}

// Update updates an existing WOD (only for user-created WODs)
func (r *WODRepository) Update(wod *domain.WOD) error {
	wod.UpdatedAt = time.Now()

	query := `UPDATE wods
	          SET name = ?, source = ?, type = ?, regime = ?, score_type = ?, description = ?, url = ?, notes = ?, updated_at = ?
	          WHERE id = ? AND is_standard = 0`

	result, err := r.db.Exec(query,
		wod.Name,
		wod.Source,
		wod.Type,
		wod.Regime,
		wod.ScoreType,
		wod.Description,
		wod.URL,
		wod.Notes,
		wod.UpdatedAt,
		wod.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update wod: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("wod not found or is a standard wod (cannot update)")
	}

	return nil
}

// UpdateStandard updates an existing standard WOD (for admin import)
func (r *WODRepository) UpdateStandard(wod *domain.WOD) error {
	wod.UpdatedAt = time.Now()

	query := `UPDATE wods
	          SET name = ?, source = ?, type = ?, regime = ?, score_type = ?, description = ?, url = ?, notes = ?, updated_at = ?
	          WHERE id = ? AND is_standard = 1`

	result, err := r.db.Exec(query,
		wod.Name,
		wod.Source,
		wod.Type,
		wod.Regime,
		wod.ScoreType,
		wod.Description,
		wod.URL,
		wod.Notes,
		wod.UpdatedAt,
		wod.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update standard wod: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("standard wod not found")
	}

	return nil
}

// Delete deletes a WOD (only for user-created WODs)
func (r *WODRepository) Delete(id int64) error {
	query := `DELETE FROM wods WHERE id = ? AND is_standard = 0`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete wod: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("wod not found or is a standard wod (cannot delete)")
	}

	return nil
}

// Search searches for WODs by name (partial match)
func (r *WODRepository) Search(query string, limit int) ([]*domain.WOD, error) {
	searchQuery := `SELECT id, name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at
	                FROM wods
	                WHERE name LIKE ?
	                ORDER BY is_standard DESC, name`

	var args []interface{}
	args = append(args, "%"+strings.TrimSpace(query)+"%")

	if limit > 0 {
		searchQuery += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := r.db.Query(searchQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search wods: %w", err)
	}
	defer rows.Close()

	return r.scanWODs(rows)
}

// Count returns the total count of WODs, optionally filtered by user
func (r *WODRepository) Count(userID *int64) (int64, error) {
	var query string
	var args []interface{}

	if userID != nil {
		query = `SELECT COUNT(*) FROM wods WHERE created_by = ?`
		args = append(args, *userID)
	} else {
		query = `SELECT COUNT(*) FROM wods`
	}

	var count int64
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count wods: %w", err)
	}

	return count, nil
}

// scanWODs scans multiple WOD rows
func (r *WODRepository) scanWODs(rows *sql.Rows) ([]*domain.WOD, error) {
	var wods []*domain.WOD
	for rows.Next() {
		wod := &domain.WOD{}
		var url, notes sql.NullString
		var createdBy sql.NullInt64

		err := rows.Scan(
			&wod.ID,
			&wod.Name,
			&wod.Source,
			&wod.Type,
			&wod.Regime,
			&wod.ScoreType,
			&wod.Description,
			&url,
			&notes,
			&wod.IsStandard,
			&createdBy,
			&wod.CreatedAt,
			&wod.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if url.Valid {
			wod.URL = &url.String
		}
		if notes.Valid {
			wod.Notes = &notes.String
		}
		if createdBy.Valid {
			wod.CreatedBy = &createdBy.Int64
		}

		wods = append(wods, wod)
	}

	return wods, rows.Err()
}
