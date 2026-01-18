package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

type ScheduleSlotRepository struct {
	db *sql.DB
}

func NewScheduleSlotRepository(db *sql.DB) *ScheduleSlotRepository {
	return &ScheduleSlotRepository{db: db}
}

// Create creates a new schedule slot
func (r *ScheduleSlotRepository) Create(slot *domain.ScheduleSlot) error {
	slot.CreatedAt = time.Now()
	slot.UpdatedAt = time.Now()

	query := rebindQuery(`
		INSERT INTO schedule_slots (template_id, location_id, day_of_week, start_time, override_capacity, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			slot.TemplateID,
			slot.LocationID,
			slot.DayOfWeek,
			slot.StartTime,
			slot.OverrideCapacity,
			slot.IsActive,
			slot.CreatedAt,
			slot.UpdatedAt,
		).Scan(&slot.ID)
	}

	result, err := r.db.Exec(query,
		slot.TemplateID,
		slot.LocationID,
		slot.DayOfWeek,
		slot.StartTime,
		slot.OverrideCapacity,
		slot.IsActive,
		slot.CreatedAt,
		slot.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create schedule slot: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get schedule slot ID: %w", err)
	}

	slot.ID = id
	return nil
}

// GetByID retrieves a schedule slot by ID
func (r *ScheduleSlotRepository) GetByID(id int64) (*domain.ScheduleSlot, error) {
	query := rebindQuery(`
		SELECT ss.id, ss.template_id, ss.location_id, ss.day_of_week, ss.start_time,
		       ss.override_capacity, ss.is_active, ss.created_at, ss.updated_at,
		       gl.name as location_name
		FROM schedule_slots ss
		LEFT JOIN gym_locations gl ON ss.location_id = gl.id
		WHERE ss.id = ?
	`)

	slot := &domain.ScheduleSlot{}
	var locationID sql.NullInt64
	var overrideCapacity sql.NullInt64
	var locationName sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&slot.ID,
		&slot.TemplateID,
		&locationID,
		&slot.DayOfWeek,
		&slot.StartTime,
		&overrideCapacity,
		&slot.IsActive,
		&slot.CreatedAt,
		&slot.UpdatedAt,
		&locationName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get schedule slot: %w", err)
	}

	if locationID.Valid {
		slot.LocationID = &locationID.Int64
	}
	if overrideCapacity.Valid {
		cap := int(overrideCapacity.Int64)
		slot.OverrideCapacity = &cap
	}
	if locationName.Valid {
		slot.LocationName = &locationName.String
	}

	return slot, nil
}

// GetByTemplateID retrieves all schedule slots for a template
func (r *ScheduleSlotRepository) GetByTemplateID(templateID int64, includeInactive bool) ([]*domain.ScheduleSlot, error) {
	var query string
	if includeInactive {
		query = rebindQuery(`
			SELECT ss.id, ss.template_id, ss.location_id, ss.day_of_week, ss.start_time,
			       ss.override_capacity, ss.is_active, ss.created_at, ss.updated_at,
			       gl.name as location_name
			FROM schedule_slots ss
			LEFT JOIN gym_locations gl ON ss.location_id = gl.id
			WHERE ss.template_id = ?
			ORDER BY ss.day_of_week ASC, ss.start_time ASC
		`)
	} else {
		query = rebindQuery(`
			SELECT ss.id, ss.template_id, ss.location_id, ss.day_of_week, ss.start_time,
			       ss.override_capacity, ss.is_active, ss.created_at, ss.updated_at,
			       gl.name as location_name
			FROM schedule_slots ss
			LEFT JOIN gym_locations gl ON ss.location_id = gl.id
			WHERE ss.template_id = ? AND ss.is_active = ?
			ORDER BY ss.day_of_week ASC, ss.start_time ASC
		`)
	}

	var rows *sql.Rows
	var err error
	if includeInactive {
		rows, err = r.db.Query(query, templateID)
	} else {
		rows, err = r.db.Query(query, templateID, true)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list schedule slots: %w", err)
	}
	defer rows.Close()

	var slots []*domain.ScheduleSlot
	for rows.Next() {
		slot := &domain.ScheduleSlot{}
		var locationID sql.NullInt64
		var overrideCapacity sql.NullInt64
		var locationName sql.NullString

		err := rows.Scan(
			&slot.ID,
			&slot.TemplateID,
			&locationID,
			&slot.DayOfWeek,
			&slot.StartTime,
			&overrideCapacity,
			&slot.IsActive,
			&slot.CreatedAt,
			&slot.UpdatedAt,
			&locationName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule slot: %w", err)
		}

		if locationID.Valid {
			slot.LocationID = &locationID.Int64
		}
		if overrideCapacity.Valid {
			cap := int(overrideCapacity.Int64)
			slot.OverrideCapacity = &cap
		}
		if locationName.Valid {
			slot.LocationName = &locationName.String
		}

		slots = append(slots, slot)
	}

	return slots, rows.Err()
}

// Update updates a schedule slot
func (r *ScheduleSlotRepository) Update(slot *domain.ScheduleSlot) error {
	slot.UpdatedAt = time.Now()

	query := rebindQuery(`
		UPDATE schedule_slots
		SET location_id = ?, day_of_week = ?, start_time = ?,
		    override_capacity = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query,
		slot.LocationID,
		slot.DayOfWeek,
		slot.StartTime,
		slot.OverrideCapacity,
		slot.IsActive,
		slot.UpdatedAt,
		slot.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update schedule slot: %w", err)
	}

	return nil
}

// Delete deletes a schedule slot
func (r *ScheduleSlotRepository) Delete(id int64) error {
	query := rebindQuery(`DELETE FROM schedule_slots WHERE id = ?`)

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete schedule slot: %w", err)
	}

	return nil
}
