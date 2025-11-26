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

// ListAllUserCreated retrieves all user-created WODs across all users (for admin view)
func (r *WODRepository) ListAllUserCreated(limit, offset int) ([]*domain.WOD, error) {
	query := `SELECT w.id, w.name, w.source, w.type, w.regime, w.score_type, w.description, w.url, w.notes, w.is_standard, w.created_by, w.created_at, w.updated_at
	          FROM wods w
	          WHERE w.is_standard = 0
	          ORDER BY w.name`

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
		return nil, fmt.Errorf("failed to list all user-created wods: %w", err)
	}
	defer rows.Close()

	return r.scanWODs(rows)
}

// ListAllUserCreatedWithUserInfo retrieves all user-created WODs with creator info (for admin view)
func (r *WODRepository) ListAllUserCreatedWithUserInfo(limit, offset int) ([]*domain.WODWithCreator, error) {
	query := `SELECT w.id, w.name, w.source, w.type, w.regime, w.score_type, w.description, w.url, w.notes, w.is_standard, w.created_by, w.created_at, w.updated_at,
	                 COALESCE(u.email, '') as creator_email, COALESCE(u.name, '') as creator_name
	          FROM wods w
	          LEFT JOIN users u ON w.created_by = u.id
	          WHERE w.is_standard = 0
	          ORDER BY w.name`

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
		return nil, fmt.Errorf("failed to list all user-created wods with user info: %w", err)
	}
	defer rows.Close()

	var wods []*domain.WODWithCreator
	for rows.Next() {
		wod := &domain.WODWithCreator{}
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
			&wod.CreatorEmail,
			&wod.CreatorName,
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

// CountAllUserCreated counts all user-created WODs
func (r *WODRepository) CountAllUserCreated() (int64, error) {
	query := `SELECT COUNT(*) FROM wods WHERE is_standard = 0`
	var count int64
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user-created wods: %w", err)
	}
	return count, nil
}

// ListAllUserCreatedWithUserInfoFiltered retrieves all user-created WODs with creator info and filters (for admin view)
func (r *WODRepository) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, scoreType, creator string) ([]*domain.WODWithCreator, int64, error) {
	baseQuery := `SELECT w.id, w.name, w.source, w.type, w.regime, w.score_type, w.description, w.url, w.notes, w.is_standard, w.created_by, w.created_at, w.updated_at,
	                 COALESCE(u.email, '') as creator_email, COALESCE(u.name, '') as creator_name
	          FROM wods w
	          LEFT JOIN users u ON w.created_by = u.id
	          WHERE w.is_standard = 0`

	countQuery := `SELECT COUNT(*) FROM wods w LEFT JOIN users u ON w.created_by = u.id WHERE w.is_standard = 0`

	var args []interface{}
	var countArgs []interface{}

	// Apply filters
	if search != "" {
		baseQuery += " AND (w.name LIKE ? OR w.description LIKE ?)"
		countQuery += " AND (w.name LIKE ? OR w.description LIKE ?)"
		searchTerm := "%" + strings.TrimSpace(search) + "%"
		args = append(args, searchTerm, searchTerm)
		countArgs = append(countArgs, searchTerm, searchTerm)
	}
	if scoreType != "" {
		baseQuery += " AND w.score_type = ?"
		countQuery += " AND w.score_type = ?"
		args = append(args, scoreType)
		countArgs = append(countArgs, scoreType)
	}
	if creator != "" {
		baseQuery += " AND u.email LIKE ?"
		countQuery += " AND u.email LIKE ?"
		creatorTerm := "%" + strings.TrimSpace(creator) + "%"
		args = append(args, creatorTerm)
		countArgs = append(countArgs, creatorTerm)
	}

	// Get count first
	var count int64
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count filtered wods: %w", err)
	}

	// Add ordering and pagination
	baseQuery += " ORDER BY w.name"
	if limit > 0 {
		baseQuery += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		baseQuery += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list filtered user-created wods: %w", err)
	}
	defer rows.Close()

	var wods []*domain.WODWithCreator
	for rows.Next() {
		wod := &domain.WODWithCreator{}
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
			&wod.CreatorEmail,
			&wod.CreatorName,
		)
		if err != nil {
			return nil, 0, err
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

	return wods, count, rows.Err()
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

// CopyToStandard creates a standard WOD by copying a user-created one
// If the new name matches the source WOD's name, it converts the source to standard
// If a user-created WOD with the new name exists, it deletes it first
func (r *WODRepository) CopyToStandard(id int64, newName string) (*domain.WOD, error) {
	// Get the source WOD
	source, err := r.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get source wod: %w", err)
	}
	if source == nil {
		return nil, fmt.Errorf("source wod not found")
	}

	now := time.Now()

	// If copying with the same name, convert the source WOD to standard
	if newName == source.Name {
		query := `UPDATE wods SET is_standard = 1, created_by = NULL, updated_at = ? WHERE id = ?`
		_, err := r.db.Exec(query, now, id)
		if err != nil {
			return nil, fmt.Errorf("failed to convert wod to standard: %w", err)
		}
		source.IsStandard = true
		source.CreatedBy = nil
		source.UpdatedAt = now
		return source, nil
	}

	// Check if there's an existing user-created WOD with the new name and delete it
	existing, err := r.GetByName(newName)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing wod: %w", err)
	}
	if existing != nil && !existing.IsStandard {
		// Delete the user-created WOD with this name
		_, err := r.db.Exec(`DELETE FROM wods WHERE id = ? AND is_standard = 0`, existing.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to remove existing user wod: %w", err)
		}
	}

	// Create a new standard WOD
	standardWOD := &domain.WOD{
		Name:        newName,
		Source:      source.Source,
		Type:        source.Type,
		Regime:      source.Regime,
		ScoreType:   source.ScoreType,
		Description: source.Description,
		URL:         source.URL,
		Notes:       source.Notes,
		IsStandard:  true,
		CreatedBy:   nil, // Standard WODs have no creator
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query := `INSERT INTO wods (name, source, type, regime, score_type, description, url, notes, is_standard, created_by, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, NULL, ?, ?)`

	result, err := r.db.Exec(query,
		standardWOD.Name,
		standardWOD.Source,
		standardWOD.Type,
		standardWOD.Regime,
		standardWOD.ScoreType,
		standardWOD.Description,
		standardWOD.URL,
		standardWOD.Notes,
		standardWOD.CreatedAt,
		standardWOD.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create standard wod: %w", err)
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get new wod ID: %w", err)
	}

	standardWOD.ID = newID
	return standardWOD, nil
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
