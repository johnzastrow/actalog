package repository

import (
	"database/sql"
	"encoding/json"
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

	// Set defaults for new recurrence fields
	if slot.RecurrenceInterval == 0 {
		slot.RecurrenceInterval = 1
	}
	if slot.RecurrenceEndType == "" {
		slot.RecurrenceEndType = domain.RecurrenceEndNone
	}

	// Serialize DaysOfWeek to JSON
	var daysOfWeekJSON *string
	if len(slot.DaysOfWeek) > 0 {
		jsonBytes, err := json.Marshal(slot.DaysOfWeek)
		if err != nil {
			return fmt.Errorf("failed to serialize days_of_week: %w", err)
		}
		jsonStr := string(jsonBytes)
		daysOfWeekJSON = &jsonStr
		// Also set DayOfWeek to first day for backwards compatibility
		slot.DayOfWeek = slot.DaysOfWeek[0]
	} else if slot.DayOfWeek >= 0 {
		// If only DayOfWeek is set, create single-element array
		jsonStr := fmt.Sprintf("[%d]", slot.DayOfWeek)
		daysOfWeekJSON = &jsonStr
	}

	query := rebindQuery(`
		INSERT INTO schedule_slots (template_id, location_id, day_of_week, days_of_week_json, start_time, override_capacity, is_active,
			recurrence_interval, recurrence_end_type, recurrence_end_count, recurrence_end_date, effective_start_date,
			created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)

	if currentDriver == "postgres" {
		query += " RETURNING id"
		return r.db.QueryRow(query,
			slot.TemplateID,
			slot.LocationID,
			slot.DayOfWeek,
			daysOfWeekJSON,
			slot.StartTime,
			slot.OverrideCapacity,
			slot.IsActive,
			slot.RecurrenceInterval,
			slot.RecurrenceEndType,
			slot.RecurrenceEndCount,
			slot.RecurrenceEndDate,
			slot.EffectiveStartDate,
			slot.CreatedAt,
			slot.UpdatedAt,
		).Scan(&slot.ID)
	}

	result, err := r.db.Exec(query,
		slot.TemplateID,
		slot.LocationID,
		slot.DayOfWeek,
		daysOfWeekJSON,
		slot.StartTime,
		slot.OverrideCapacity,
		slot.IsActive,
		slot.RecurrenceInterval,
		slot.RecurrenceEndType,
		slot.RecurrenceEndCount,
		slot.RecurrenceEndDate,
		slot.EffectiveStartDate,
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
		SELECT ss.id, ss.template_id, ss.location_id, ss.day_of_week, ss.days_of_week_json, ss.start_time,
		       ss.override_capacity, ss.is_active, ss.created_at, ss.updated_at,
		       ss.recurrence_interval, ss.recurrence_end_type, ss.recurrence_end_count,
		       ss.recurrence_end_date, ss.effective_start_date,
		       gl.name as location_name
		FROM schedule_slots ss
		LEFT JOIN gym_locations gl ON ss.location_id = gl.id
		WHERE ss.id = ?
	`)

	slot := &domain.ScheduleSlot{}
	var locationID sql.NullInt64
	var daysOfWeekJSON sql.NullString
	var overrideCapacity sql.NullInt64
	var locationName sql.NullString
	var recurrenceInterval sql.NullInt64
	var recurrenceEndType sql.NullString
	var recurrenceEndCount sql.NullInt64
	var recurrenceEndDateStr sql.NullString
	var effectiveStartDateStr sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&slot.ID,
		&slot.TemplateID,
		&locationID,
		&slot.DayOfWeek,
		&daysOfWeekJSON,
		&slot.StartTime,
		&overrideCapacity,
		&slot.IsActive,
		&slot.CreatedAt,
		&slot.UpdatedAt,
		&recurrenceInterval,
		&recurrenceEndType,
		&recurrenceEndCount,
		&recurrenceEndDateStr,
		&effectiveStartDateStr,
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
	// Parse days_of_week_json
	if daysOfWeekJSON.Valid && daysOfWeekJSON.String != "" {
		var days []int
		if err := json.Unmarshal([]byte(daysOfWeekJSON.String), &days); err == nil {
			slot.DaysOfWeek = days
		}
	}
	// Fallback to single day if no multi-day data
	if len(slot.DaysOfWeek) == 0 {
		slot.DaysOfWeek = []int{slot.DayOfWeek}
	}
	if overrideCapacity.Valid {
		cap := int(overrideCapacity.Int64)
		slot.OverrideCapacity = &cap
	}
	if locationName.Valid {
		slot.LocationName = &locationName.String
	}
	if recurrenceInterval.Valid {
		slot.RecurrenceInterval = int(recurrenceInterval.Int64)
	} else {
		slot.RecurrenceInterval = 1 // Default
	}
	if recurrenceEndType.Valid {
		slot.RecurrenceEndType = recurrenceEndType.String
	} else {
		slot.RecurrenceEndType = domain.RecurrenceEndNone
	}
	if recurrenceEndCount.Valid {
		count := int(recurrenceEndCount.Int64)
		slot.RecurrenceEndCount = &count
	}
	if recurrenceEndDateStr.Valid && recurrenceEndDateStr.String != "" {
		if t, err := parseDateString(recurrenceEndDateStr.String); err == nil {
			slot.RecurrenceEndDate = &t
		}
	}
	if effectiveStartDateStr.Valid && effectiveStartDateStr.String != "" {
		if t, err := parseDateString(effectiveStartDateStr.String); err == nil {
			slot.EffectiveStartDate = &t
		}
	}

	return slot, nil
}

// parseDateString parses various date string formats from different database drivers
func parseDateString(s string) (time.Time, error) {
	// Try common formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05+00:00",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}

// GetByTemplateID retrieves all schedule slots for a template
func (r *ScheduleSlotRepository) GetByTemplateID(templateID int64, includeInactive bool) ([]*domain.ScheduleSlot, error) {
	var query string
	if includeInactive {
		query = rebindQuery(`
			SELECT ss.id, ss.template_id, ss.location_id, ss.day_of_week, ss.days_of_week_json, ss.start_time,
			       ss.override_capacity, ss.is_active, ss.created_at, ss.updated_at,
			       ss.recurrence_interval, ss.recurrence_end_type, ss.recurrence_end_count,
			       ss.recurrence_end_date, ss.effective_start_date,
			       gl.name as location_name
			FROM schedule_slots ss
			LEFT JOIN gym_locations gl ON ss.location_id = gl.id
			WHERE ss.template_id = ?
			ORDER BY ss.day_of_week ASC, ss.start_time ASC
		`)
	} else {
		query = rebindQuery(`
			SELECT ss.id, ss.template_id, ss.location_id, ss.day_of_week, ss.days_of_week_json, ss.start_time,
			       ss.override_capacity, ss.is_active, ss.created_at, ss.updated_at,
			       ss.recurrence_interval, ss.recurrence_end_type, ss.recurrence_end_count,
			       ss.recurrence_end_date, ss.effective_start_date,
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
		var daysOfWeekJSON sql.NullString
		var overrideCapacity sql.NullInt64
		var locationName sql.NullString
		var recurrenceInterval sql.NullInt64
		var recurrenceEndType sql.NullString
		var recurrenceEndCount sql.NullInt64
		var recurrenceEndDateStr sql.NullString
		var effectiveStartDateStr sql.NullString

		err := rows.Scan(
			&slot.ID,
			&slot.TemplateID,
			&locationID,
			&slot.DayOfWeek,
			&daysOfWeekJSON,
			&slot.StartTime,
			&overrideCapacity,
			&slot.IsActive,
			&slot.CreatedAt,
			&slot.UpdatedAt,
			&recurrenceInterval,
			&recurrenceEndType,
			&recurrenceEndCount,
			&recurrenceEndDateStr,
			&effectiveStartDateStr,
			&locationName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule slot: %w", err)
		}

		if locationID.Valid {
			slot.LocationID = &locationID.Int64
		}
		// Parse days_of_week_json
		if daysOfWeekJSON.Valid && daysOfWeekJSON.String != "" {
			var days []int
			if err := json.Unmarshal([]byte(daysOfWeekJSON.String), &days); err == nil {
				slot.DaysOfWeek = days
			}
		}
		// Fallback to single day if no multi-day data
		if len(slot.DaysOfWeek) == 0 {
			slot.DaysOfWeek = []int{slot.DayOfWeek}
		}
		if overrideCapacity.Valid {
			cap := int(overrideCapacity.Int64)
			slot.OverrideCapacity = &cap
		}
		if locationName.Valid {
			slot.LocationName = &locationName.String
		}
		if recurrenceInterval.Valid {
			slot.RecurrenceInterval = int(recurrenceInterval.Int64)
		} else {
			slot.RecurrenceInterval = 1 // Default
		}
		if recurrenceEndType.Valid {
			slot.RecurrenceEndType = recurrenceEndType.String
		} else {
			slot.RecurrenceEndType = domain.RecurrenceEndNone
		}
		if recurrenceEndCount.Valid {
			count := int(recurrenceEndCount.Int64)
			slot.RecurrenceEndCount = &count
		}
		if recurrenceEndDateStr.Valid && recurrenceEndDateStr.String != "" {
			if t, err := parseDateString(recurrenceEndDateStr.String); err == nil {
				slot.RecurrenceEndDate = &t
			}
		}
		if effectiveStartDateStr.Valid && effectiveStartDateStr.String != "" {
			if t, err := parseDateString(effectiveStartDateStr.String); err == nil {
				slot.EffectiveStartDate = &t
			}
		}

		slots = append(slots, slot)
	}

	return slots, rows.Err()
}

// Update updates a schedule slot
func (r *ScheduleSlotRepository) Update(slot *domain.ScheduleSlot) error {
	slot.UpdatedAt = time.Now()

	// Set defaults for recurrence fields
	if slot.RecurrenceInterval == 0 {
		slot.RecurrenceInterval = 1
	}
	if slot.RecurrenceEndType == "" {
		slot.RecurrenceEndType = domain.RecurrenceEndNone
	}

	// Serialize DaysOfWeek to JSON
	var daysOfWeekJSON *string
	if len(slot.DaysOfWeek) > 0 {
		jsonBytes, err := json.Marshal(slot.DaysOfWeek)
		if err != nil {
			return fmt.Errorf("failed to serialize days_of_week: %w", err)
		}
		jsonStr := string(jsonBytes)
		daysOfWeekJSON = &jsonStr
		// Also set DayOfWeek to first day for backwards compatibility
		slot.DayOfWeek = slot.DaysOfWeek[0]
	} else if slot.DayOfWeek >= 0 {
		// If only DayOfWeek is set, create single-element array
		jsonStr := fmt.Sprintf("[%d]", slot.DayOfWeek)
		daysOfWeekJSON = &jsonStr
	}

	query := rebindQuery(`
		UPDATE schedule_slots
		SET location_id = ?, day_of_week = ?, days_of_week_json = ?, start_time = ?,
		    override_capacity = ?, is_active = ?,
		    recurrence_interval = ?, recurrence_end_type = ?,
		    recurrence_end_count = ?, recurrence_end_date = ?,
		    effective_start_date = ?, updated_at = ?
		WHERE id = ?
	`)

	_, err := r.db.Exec(query,
		slot.LocationID,
		slot.DayOfWeek,
		daysOfWeekJSON,
		slot.StartTime,
		slot.OverrideCapacity,
		slot.IsActive,
		slot.RecurrenceInterval,
		slot.RecurrenceEndType,
		slot.RecurrenceEndCount,
		slot.RecurrenceEndDate,
		slot.EffectiveStartDate,
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
