package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

type WODRepository struct {
	db *sql.DB
}

func NewWODRepository(db *sql.DB) *WODRepository {
	return &WODRepository{db: db}
}

// Create creates a new custom WOD
func (r *WODRepository) Create(wod *domain.WOD) error {
	wod.CreatedAt = time.Now()
	wod.UpdatedAt = time.Now()

	query := `INSERT INTO wods (name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, wod.Name, wod.Source, wod.Type, wod.Regime, wod.ScoreType, wod.Description, wod.URL, wod.Notes, wod.IsStandard, wod.CreatedBy, wod.CreatedAt, wod.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create WOD: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get WOD ID: %w", err)
	}

	wod.ID = id
	return nil
}

// GetByID retrieves a WOD by ID
func (r *WODRepository) GetByID(id int64) (*domain.WOD, error) {
	query := `SELECT id, name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at FROM wods WHERE id = ?`

	wod := &domain.WOD{}
	var url sql.NullString
	var notes sql.NullString
	var createdBy sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(&wod.ID, &wod.Name, &wod.Source, &wod.Type, &wod.Regime, &wod.ScoreType, &wod.Description, &url, &notes, &wod.IsStandard, &createdBy, &wod.CreatedAt, &wod.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get WOD: %w", err)
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
	query := `SELECT id, name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at FROM wods WHERE name = ?`

	wod := &domain.WOD{}
	var url sql.NullString
	var notes sql.NullString
	var createdBy sql.NullInt64

	err := r.db.QueryRow(query, name).Scan(&wod.ID, &wod.Name, &wod.Source, &wod.Type, &wod.Regime, &wod.ScoreType, &wod.Description, &url, &notes, &wod.IsStandard, &createdBy, &wod.CreatedAt, &wod.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get WOD by name: %w", err)
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

// List retrieves all WODs with optional filtering
func (r *WODRepository) List(filters map[string]interface{}) ([]*domain.WOD, error) {
	query := `SELECT id, name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at FROM wods WHERE 1=1`
	args := []interface{}{}

	if source, ok := filters["source"].(string); ok && source != "" {
		query += ` AND source = ?`
		args = append(args, source)
	}

	if wodType, ok := filters["type"].(string); ok && wodType != "" {
		query += ` AND type = ?`
		args = append(args, wodType)
	}

	if regime, ok := filters["regime"].(string); ok && regime != "" {
		query += ` AND regime = ?`
		args = append(args, regime)
	}

	query += ` ORDER BY name`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list WODs: %w", err)
	}
	defer rows.Close()

	return r.scanWODs(rows)
}

// ListStandard retrieves all standard (pre-seeded) WODs
func (r *WODRepository) ListStandard() ([]*domain.WOD, error) {
	query := `SELECT id, name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at FROM wods WHERE is_standard = 1 ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list standard WODs: %w", err)
	}
	defer rows.Close()

	return r.scanWODs(rows)
}

// ListByUser retrieves all custom WODs created by a specific user
func (r *WODRepository) ListByUser(userID int64) ([]*domain.WOD, error) {
	query := `SELECT id, name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at FROM wods WHERE created_by = ? ORDER BY name`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user WODs: %w", err)
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

	result, err := r.db.Exec(query, wod.Name, wod.Source, wod.Type, wod.Regime, wod.ScoreType, wod.Description, wod.URL, wod.Notes, wod.UpdatedAt, wod.ID)
	if err != nil {
		return fmt.Errorf("failed to update WOD: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("WOD not found or is a standard WOD (cannot update)")
	}

	return nil
}

// Delete deletes a WOD (only for user-created WODs)
func (r *WODRepository) Delete(id int64) error {
	query := `DELETE FROM wods WHERE id = ? AND is_standard = 0`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete WOD: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("WOD not found or is a standard WOD (cannot delete)")
	}

	return nil
}

// Search searches WODs by name (partial match)
func (r *WODRepository) Search(query string) ([]*domain.WOD, error) {
	searchQuery := `SELECT id, name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at FROM wods WHERE name LIKE ? ORDER BY name LIMIT 20`

	rows, err := r.db.Query(searchQuery, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search WODs: %w", err)
	}
	defer rows.Close()

	return r.scanWODs(rows)
}

// scanWODs scans multiple WOD rows
func (r *WODRepository) scanWODs(rows *sql.Rows) ([]*domain.WOD, error) {
	var wods []*domain.WOD
	for rows.Next() {
		wod := &domain.WOD{}
		var url sql.NullString
		var notes sql.NullString
		var createdBy sql.NullInt64

		err := rows.Scan(&wod.ID, &wod.Name, &wod.Source, &wod.Type, &wod.Regime, &wod.ScoreType, &wod.Description, &url, &notes, &wod.IsStandard, &createdBy, &wod.CreatedAt, &wod.UpdatedAt)
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
