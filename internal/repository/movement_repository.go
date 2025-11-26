package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// MovementRepository implements domain.MovementRepository
// Note: After v0.4.0 migration, this accesses the 'movements' table
type MovementRepository struct {
	db *sql.DB
}

// NewMovementRepository creates a new movement repository
func NewMovementRepository(db *sql.DB) *MovementRepository {
	return &MovementRepository{db: db}
}

// Create creates a new movement
func (r *MovementRepository) Create(movement *domain.Movement) error {
	movement.CreatedAt = time.Now()
	movement.UpdatedAt = time.Now()

	query := `INSERT INTO movements (name, description, type, is_standard, created_by, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, movement.Name, movement.Description, movement.Type, movement.IsStandard, movement.CreatedBy, movement.CreatedAt, movement.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get movement ID: %w", err)
	}

	movement.ID = id
	return nil
}

// GetByID retrieves a movement by ID
func (r *MovementRepository) GetByID(id int64) (*domain.Movement, error) {
	query := `SELECT id, name, description, type, is_standard, created_by, created_at, updated_at FROM movements WHERE id = ?`

	movement := &domain.Movement{}
	var createdBy sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(&movement.ID, &movement.Name, &movement.Description, &movement.Type, &movement.IsStandard, &createdBy, &movement.CreatedAt, &movement.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get movement: %w", err)
	}

	if createdBy.Valid {
		movement.CreatedBy = &createdBy.Int64
	}

	return movement, nil
}

// GetByName retrieves a movement by name
func (r *MovementRepository) GetByName(name string) (*domain.Movement, error) {
	query := `SELECT id, name, description, type, is_standard, created_by, created_at, updated_at FROM movements WHERE name = ?`

	movement := &domain.Movement{}
	var createdBy sql.NullInt64

	err := r.db.QueryRow(query, name).Scan(&movement.ID, &movement.Name, &movement.Description, &movement.Type, &movement.IsStandard, &createdBy, &movement.CreatedAt, &movement.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get movement by name: %w", err)
	}

	if createdBy.Valid {
		movement.CreatedBy = &createdBy.Int64
	}

	return movement, nil
}

// ListStandard retrieves all standard movements
func (r *MovementRepository) ListStandard() ([]*domain.Movement, error) {
	query := `SELECT id, name, description, type, is_standard, created_by, created_at, updated_at FROM movements WHERE is_standard = 1 ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list standard movements: %w", err)
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

// ListAll retrieves all movements (both standard and custom)
func (r *MovementRepository) ListAll() ([]*domain.Movement, error) {
	query := `SELECT id, name, description, type, is_standard, created_by, created_at, updated_at FROM movements ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all movements: %w", err)
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

// ListByUser retrieves movements created by a user
func (r *MovementRepository) ListByUser(userID int64) ([]*domain.Movement, error) {
	query := `SELECT id, name, description, type, is_standard, created_by, created_at, updated_at FROM movements WHERE created_by = ? ORDER BY name`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user movements: %w", err)
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

// Update updates a movement (only for user-created movements)
func (r *MovementRepository) Update(movement *domain.Movement) error {
	movement.UpdatedAt = time.Now()

	query := `UPDATE movements
	          SET name = ?, description = ?, type = ?, updated_at = ?
	          WHERE id = ? AND is_standard = 0`

	result, err := r.db.Exec(query, movement.Name, movement.Description, movement.Type, movement.UpdatedAt, movement.ID)
	if err != nil {
		return fmt.Errorf("failed to update movement: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("movement not found or is a standard movement (cannot update)")
	}

	return nil
}

// Delete deletes a movement (only for user-created movements)
func (r *MovementRepository) Delete(id int64) error {
	query := `DELETE FROM movements WHERE id = ? AND is_standard = 0`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete movement: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("movement not found or is a standard movement (cannot delete)")
	}

	return nil
}

// ListAllUserCreated retrieves all user-created movements across all users (for admin view)
func (r *MovementRepository) ListAllUserCreated() ([]*domain.Movement, error) {
	query := `SELECT id, name, description, type, is_standard, created_by, created_at, updated_at
	          FROM movements WHERE is_standard = 0 ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all user-created movements: %w", err)
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

// ListAllUserCreatedWithUserInfo retrieves all user-created movements with creator info (for admin view)
func (r *MovementRepository) ListAllUserCreatedWithUserInfo() ([]*domain.MovementWithCreator, error) {
	query := `SELECT m.id, m.name, m.description, m.type, m.is_standard, m.created_by, m.created_at, m.updated_at,
	                 COALESCE(u.email, '') as creator_email, COALESCE(u.name, '') as creator_name
	          FROM movements m
	          LEFT JOIN users u ON m.created_by = u.id
	          WHERE m.is_standard = 0
	          ORDER BY m.name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all user-created movements with user info: %w", err)
	}
	defer rows.Close()

	var movements []*domain.MovementWithCreator
	for rows.Next() {
		m := &domain.MovementWithCreator{}
		var createdBy sql.NullInt64

		err := rows.Scan(
			&m.ID,
			&m.Name,
			&m.Description,
			&m.Type,
			&m.IsStandard,
			&createdBy,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.CreatorEmail,
			&m.CreatorName,
		)
		if err != nil {
			return nil, err
		}

		if createdBy.Valid {
			m.CreatedBy = &createdBy.Int64
		}

		movements = append(movements, m)
	}

	return movements, rows.Err()
}

// CountAllUserCreated counts all user-created movements
func (r *MovementRepository) CountAllUserCreated() (int64, error) {
	query := `SELECT COUNT(*) FROM movements WHERE is_standard = 0`
	var count int64
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user-created movements: %w", err)
	}
	return count, nil
}

// ListAllUserCreatedWithUserInfoFiltered retrieves all user-created movements with creator info and filters (for admin view)
func (r *MovementRepository) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, movementType, creator string) ([]*domain.MovementWithCreator, int64, error) {
	baseQuery := `SELECT m.id, m.name, m.description, m.type, m.is_standard, m.created_by, m.created_at, m.updated_at,
	                 COALESCE(u.email, '') as creator_email, COALESCE(u.name, '') as creator_name
	          FROM movements m
	          LEFT JOIN users u ON m.created_by = u.id
	          WHERE m.is_standard = 0`

	countQuery := `SELECT COUNT(*) FROM movements m LEFT JOIN users u ON m.created_by = u.id WHERE m.is_standard = 0`

	var args []interface{}
	var countArgs []interface{}

	// Apply filters
	if search != "" {
		baseQuery += " AND (m.name LIKE ? OR m.description LIKE ?)"
		countQuery += " AND (m.name LIKE ? OR m.description LIKE ?)"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
		countArgs = append(countArgs, searchTerm, searchTerm)
	}
	if movementType != "" {
		baseQuery += " AND m.type = ?"
		countQuery += " AND m.type = ?"
		args = append(args, movementType)
		countArgs = append(countArgs, movementType)
	}
	if creator != "" {
		baseQuery += " AND u.email LIKE ?"
		countQuery += " AND u.email LIKE ?"
		creatorTerm := "%" + creator + "%"
		args = append(args, creatorTerm)
		countArgs = append(countArgs, creatorTerm)
	}

	// Get count first
	var count int64
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count filtered movements: %w", err)
	}

	// Add ordering and pagination
	baseQuery += " ORDER BY m.name"
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
		return nil, 0, fmt.Errorf("failed to list filtered user-created movements: %w", err)
	}
	defer rows.Close()

	var movements []*domain.MovementWithCreator
	for rows.Next() {
		m := &domain.MovementWithCreator{}
		var createdBy sql.NullInt64

		err := rows.Scan(
			&m.ID,
			&m.Name,
			&m.Description,
			&m.Type,
			&m.IsStandard,
			&createdBy,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.CreatorEmail,
			&m.CreatorName,
		)
		if err != nil {
			return nil, 0, err
		}

		if createdBy.Valid {
			m.CreatedBy = &createdBy.Int64
		}

		movements = append(movements, m)
	}

	return movements, count, rows.Err()
}

// Search searches for movements by name
func (r *MovementRepository) Search(query string, limit int) ([]*domain.Movement, error) {
	searchQuery := `SELECT id, name, description, type, is_standard, created_by, created_at, updated_at FROM movements
	                WHERE name LIKE ?
	                ORDER BY is_standard DESC, name
	                LIMIT ?`

	rows, err := r.db.Query(searchQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search movements: %w", err)
	}
	defer rows.Close()

	return r.scanMovements(rows)
}

// CopyToStandard creates a standard movement by copying a user-created one
func (r *MovementRepository) CopyToStandard(id int64, newName string) (*domain.Movement, error) {
	// Get the source movement
	source, err := r.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get source movement: %w", err)
	}
	if source == nil {
		return nil, fmt.Errorf("source movement not found")
	}

	// Create a new standard movement
	now := time.Now()
	standardMovement := &domain.Movement{
		Name:        newName,
		Description: source.Description,
		Type:        source.Type,
		IsStandard:  true,
		CreatedBy:   nil, // Standard movements have no creator
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query := `INSERT INTO movements (name, description, type, is_standard, created_by, created_at, updated_at)
	          VALUES (?, ?, ?, 1, NULL, ?, ?)`

	result, err := r.db.Exec(query,
		standardMovement.Name,
		standardMovement.Description,
		standardMovement.Type,
		standardMovement.CreatedAt,
		standardMovement.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create standard movement: %w", err)
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get new movement ID: %w", err)
	}

	standardMovement.ID = newID
	return standardMovement, nil
}

// scanMovements scans multiple movement rows
func (r *MovementRepository) scanMovements(rows *sql.Rows) ([]*domain.Movement, error) {
	var movements []*domain.Movement
	for rows.Next() {
		movement := &domain.Movement{}
		var createdBy sql.NullInt64

		err := rows.Scan(&movement.ID, &movement.Name, &movement.Description, &movement.Type, &movement.IsStandard, &createdBy, &movement.CreatedAt, &movement.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if createdBy.Valid {
			movement.CreatedBy = &createdBy.Int64
		}

		movements = append(movements, movement)
	}

	return movements, rows.Err()
}
