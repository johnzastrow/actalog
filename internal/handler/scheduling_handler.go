package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

type SchedulingHandler struct {
	schedulingService *service.SchedulingService
	logger            *logger.Logger
}

func NewSchedulingHandler(schedulingService *service.SchedulingService, logger *logger.Logger) *SchedulingHandler {
	return &SchedulingHandler{
		schedulingService: schedulingService,
		logger:            logger,
	}
}

// ============================================
// GYM LOCATION HANDLERS
// ============================================

// CreateLocation handles POST /api/gyms/{gym_id}/locations
func (h *SchedulingHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Address     *string `json:"address"`
		Capacity    int     `json:"capacity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Location name is required")
		return
	}

	location := &domain.GymLocation{
		OrganizationID: gymID,
		Name:           req.Name,
		Description:    req.Description,
		Address:        req.Address,
		Capacity:       req.Capacity,
	}

	if err := h.schedulingService.CreateLocation(userID, location); err != nil {
		if err == service.ErrOrganizationNotFound {
			respondError(w, http.StatusNotFound, "Organization not found")
			return
		}
		h.logger.Error("Failed to create location: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create location")
		return
	}

	respondJSON(w, http.StatusCreated, location)
}

// ListLocations handles GET /api/gyms/{gym_id}/locations
func (h *SchedulingHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	locations, err := h.schedulingService.GetLocationsByOrganization(gymID, includeInactive)
	if err != nil {
		h.logger.Error("Failed to list locations: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list locations")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"locations": locations,
		"count":     len(locations),
	})
}

// GetLocation handles GET /api/gyms/{gym_id}/locations/{id}
func (h *SchedulingHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid location ID")
		return
	}

	location, err := h.schedulingService.GetLocationByID(id)
	if err != nil {
		if err == service.ErrGymLocationNotFound {
			respondError(w, http.StatusNotFound, "Location not found")
			return
		}
		h.logger.Error("Failed to get location: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get location")
		return
	}

	respondJSON(w, http.StatusOK, location)
}

// UpdateLocation handles PUT /api/gyms/{gym_id}/locations/{id}
func (h *SchedulingHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid location ID")
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Address     *string `json:"address"`
		Capacity    int     `json:"capacity"`
		IsActive    bool    `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	location := &domain.GymLocation{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Capacity:    req.Capacity,
		IsActive:    req.IsActive,
	}

	if err := h.schedulingService.UpdateLocation(userID, location); err != nil {
		if err == service.ErrGymLocationNotFound {
			respondError(w, http.StatusNotFound, "Location not found")
			return
		}
		h.logger.Error("Failed to update location: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update location")
		return
	}

	respondJSON(w, http.StatusOK, location)
}

// DeleteLocation handles DELETE /api/gyms/{gym_id}/locations/{id}
func (h *SchedulingHandler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid location ID")
		return
	}

	if err := h.schedulingService.DeleteLocation(userID, id); err != nil {
		if err == service.ErrGymLocationNotFound {
			respondError(w, http.StatusNotFound, "Location not found")
			return
		}
		h.logger.Error("Failed to delete location: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to delete location")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Location deleted successfully"})
}

// ============================================
// CLASS TEMPLATE HANDLERS
// ============================================

// CreateTemplate handles POST /api/gyms/{gym_id}/templates
func (h *SchedulingHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	var req struct {
		Name              string  `json:"name"`
		Description       *string `json:"description"`
		WorkoutID         *int64  `json:"workout_id"`
		DurationMinutes   int     `json:"duration_minutes"`
		DefaultCapacity   int     `json:"default_capacity"`
		DefaultLocationID *int64  `json:"default_location_id"`
		Color             string  `json:"color"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Template name is required")
		return
	}

	template := &domain.ClassTemplate{
		OrganizationID:    gymID,
		Name:              req.Name,
		Description:       req.Description,
		WorkoutID:         req.WorkoutID,
		DurationMinutes:   req.DurationMinutes,
		DefaultCapacity:   req.DefaultCapacity,
		DefaultLocationID: req.DefaultLocationID,
		Color:             req.Color,
	}

	if err := h.schedulingService.CreateTemplate(userID, template); err != nil {
		if err == service.ErrOrganizationNotFound {
			respondError(w, http.StatusNotFound, "Organization not found")
			return
		}
		h.logger.Error("Failed to create template: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create template")
		return
	}

	respondJSON(w, http.StatusCreated, template)
}

// ListTemplates handles GET /api/gyms/{gym_id}/templates
func (h *SchedulingHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	templates, err := h.schedulingService.GetTemplatesByOrganization(gymID, includeInactive)
	if err != nil {
		h.logger.Error("Failed to list templates: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list templates")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

// GetTemplate handles GET /api/templates/{id}
func (h *SchedulingHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	template, err := h.schedulingService.GetTemplateByID(id)
	if err != nil {
		if err == service.ErrClassTemplateNotFound {
			respondError(w, http.StatusNotFound, "Template not found")
			return
		}
		h.logger.Error("Failed to get template: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get template")
		return
	}

	respondJSON(w, http.StatusOK, template)
}

// UpdateTemplate handles PUT /api/templates/{id}
func (h *SchedulingHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	var req struct {
		Name              string  `json:"name"`
		Description       *string `json:"description"`
		WorkoutID         *int64  `json:"workout_id"`
		DurationMinutes   int     `json:"duration_minutes"`
		DefaultCapacity   int     `json:"default_capacity"`
		DefaultLocationID *int64  `json:"default_location_id"`
		Color             string  `json:"color"`
		IsActive          bool    `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	template := &domain.ClassTemplate{
		ID:                id,
		Name:              req.Name,
		Description:       req.Description,
		WorkoutID:         req.WorkoutID,
		DurationMinutes:   req.DurationMinutes,
		DefaultCapacity:   req.DefaultCapacity,
		DefaultLocationID: req.DefaultLocationID,
		Color:             req.Color,
		IsActive:          req.IsActive,
	}

	if err := h.schedulingService.UpdateTemplate(userID, template); err != nil {
		if err == service.ErrClassTemplateNotFound {
			respondError(w, http.StatusNotFound, "Template not found")
			return
		}
		h.logger.Error("Failed to update template: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update template")
		return
	}

	respondJSON(w, http.StatusOK, template)
}

// DeleteTemplate handles DELETE /api/templates/{id}
// Query params:
//   - mode: template_only (default), with_all_sessions, with_future_sessions
func (h *SchedulingHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	// Check for mode parameter
	mode := r.URL.Query().Get("mode")
	if mode != "" {
		// Use the new DeleteTemplateWithMode method
		deleteMode := service.DeleteMode(mode)
		switch deleteMode {
		case service.DeleteModeTemplateOnly, service.DeleteModeWithAllSessions, service.DeleteModeWithFutureSessions:
			// Valid mode
		default:
			respondError(w, http.StatusBadRequest, "Invalid mode. Must be: template_only, with_all_sessions, or with_future_sessions")
			return
		}

		result, err := h.schedulingService.DeleteTemplateWithMode(userID, id, deleteMode)
		if err != nil {
			if err == service.ErrClassTemplateNotFound {
				respondError(w, http.StatusNotFound, "Template not found")
				return
			}
			h.logger.Error("Failed to delete template with mode: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to delete template")
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"message":            "Template deleted successfully",
			"sessions_deleted":   result.SessionsDeleted,
			"notifications_sent": result.NotificationsSent,
			"credits_refunded":   result.CreditsRefunded,
			"mode":               string(result.Mode),
		})
		return
	}

	// Default behavior: just delete the template
	if err := h.schedulingService.DeleteTemplate(userID, id); err != nil {
		if err == service.ErrClassTemplateNotFound {
			respondError(w, http.StatusNotFound, "Template not found")
			return
		}
		h.logger.Error("Failed to delete template: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to delete template")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Template deleted successfully"})
}

// ============================================
// SCHEDULE PREVIEW HANDLERS
// ============================================

// PreviewSchedule handles GET /api/templates/{id}/preview-schedule
func (h *SchedulingHandler) PreviewSchedule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	// Parse months parameter (default 3)
	monthsStr := r.URL.Query().Get("months")
	months := 3
	if monthsStr != "" {
		if m, err := strconv.Atoi(monthsStr); err == nil && m > 0 && m <= 12 {
			months = m
		}
	}

	dates, err := h.schedulingService.PreviewScheduleDates(id, months)
	if err != nil {
		if err == service.ErrClassTemplateNotFound {
			respondError(w, http.StatusNotFound, "Template not found")
			return
		}
		h.logger.Error("Failed to preview schedule: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to preview schedule")
		return
	}

	// Convert dates to strings for JSON response
	dateStrings := make([]string, len(dates))
	for i, d := range dates {
		dateStrings[i] = d.Format(time.RFC3339)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"dates": dateStrings,
		"count": len(dates),
	})
}

// ============================================
// SCHEDULE SLOT HANDLERS
// ============================================

// CreateScheduleSlot handles POST /api/templates/{id}/slots
func (h *SchedulingHandler) CreateScheduleSlot(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	templateIDStr := chi.URLParam(r, "id")
	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	var req struct {
		LocationID         *int64  `json:"location_id"`
		DayOfWeek          int     `json:"day_of_week"`
		DaysOfWeek         []int   `json:"days_of_week"`
		StartTime          string  `json:"start_time"`
		OverrideCapacity   *int    `json:"override_capacity"`
		RecurrenceInterval int     `json:"recurrence_interval"`
		RecurrenceEndType  string  `json:"recurrence_end_type"`
		RecurrenceEndCount *int    `json:"recurrence_end_count"`
		RecurrenceEndDate  *string `json:"recurrence_end_date"`
		EffectiveStartDate *string `json:"effective_start_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.StartTime == "" {
		respondError(w, http.StatusBadRequest, "Start time is required")
		return
	}

	// Use DaysOfWeek if provided, otherwise fall back to DayOfWeek
	daysOfWeek := req.DaysOfWeek
	dayOfWeek := req.DayOfWeek
	if len(daysOfWeek) > 0 {
		dayOfWeek = daysOfWeek[0] // Set legacy field to first day
	} else if dayOfWeek >= 0 {
		daysOfWeek = []int{dayOfWeek}
	}

	slot := &domain.ScheduleSlot{
		TemplateID:         templateID,
		LocationID:         req.LocationID,
		DayOfWeek:          dayOfWeek,
		DaysOfWeek:         daysOfWeek,
		StartTime:          req.StartTime,
		OverrideCapacity:   req.OverrideCapacity,
		RecurrenceInterval: req.RecurrenceInterval,
		RecurrenceEndType:  req.RecurrenceEndType,
		RecurrenceEndCount: req.RecurrenceEndCount,
	}

	// Parse recurrence end date
	if req.RecurrenceEndDate != nil && *req.RecurrenceEndDate != "" {
		t, err := time.Parse("2006-01-02", *req.RecurrenceEndDate)
		if err == nil {
			slot.RecurrenceEndDate = &t
		}
	}

	// Parse effective start date
	if req.EffectiveStartDate != nil && *req.EffectiveStartDate != "" {
		t, err := time.Parse("2006-01-02", *req.EffectiveStartDate)
		if err == nil {
			slot.EffectiveStartDate = &t
		}
	}

	if err := h.schedulingService.CreateScheduleSlot(userID, slot); err != nil {
		if err == service.ErrClassTemplateNotFound {
			respondError(w, http.StatusNotFound, "Template not found")
			return
		}
		if err == service.ErrInvalidDayOfWeek {
			respondError(w, http.StatusBadRequest, "Day of week must be 0-6 (Sunday-Saturday)")
			return
		}
		if err == service.ErrInvalidTimeFormat {
			respondError(w, http.StatusBadRequest, "Time must be in HH:MM format")
			return
		}
		h.logger.Error("Failed to create schedule slot: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create schedule slot")
		return
	}

	respondJSON(w, http.StatusCreated, slot)
}

// ListScheduleSlots handles GET /api/templates/{id}/slots
func (h *SchedulingHandler) ListScheduleSlots(w http.ResponseWriter, r *http.Request) {
	templateIDStr := chi.URLParam(r, "id")
	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	slots, err := h.schedulingService.GetScheduleSlotsByTemplate(templateID, includeInactive)
	if err != nil {
		h.logger.Error("Failed to list schedule slots: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list schedule slots")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"slots": slots,
		"count": len(slots),
	})
}

// UpdateScheduleSlot handles PUT /api/templates/{template_id}/slots/{id}
func (h *SchedulingHandler) UpdateScheduleSlot(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "slot_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid slot ID")
		return
	}

	var req struct {
		LocationID         *int64  `json:"location_id"`
		DayOfWeek          int     `json:"day_of_week"`
		DaysOfWeek         []int   `json:"days_of_week"`
		StartTime          string  `json:"start_time"`
		OverrideCapacity   *int    `json:"override_capacity"`
		IsActive           bool    `json:"is_active"`
		RecurrenceInterval int     `json:"recurrence_interval"`
		RecurrenceEndType  string  `json:"recurrence_end_type"`
		RecurrenceEndCount *int    `json:"recurrence_end_count"`
		RecurrenceEndDate  *string `json:"recurrence_end_date"`
		EffectiveStartDate *string `json:"effective_start_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Use DaysOfWeek if provided, otherwise fall back to DayOfWeek
	daysOfWeek := req.DaysOfWeek
	dayOfWeek := req.DayOfWeek
	if len(daysOfWeek) > 0 {
		dayOfWeek = daysOfWeek[0] // Set legacy field to first day
	} else if dayOfWeek >= 0 {
		daysOfWeek = []int{dayOfWeek}
	}

	slot := &domain.ScheduleSlot{
		ID:                 id,
		LocationID:         req.LocationID,
		DayOfWeek:          dayOfWeek,
		DaysOfWeek:         daysOfWeek,
		StartTime:          req.StartTime,
		OverrideCapacity:   req.OverrideCapacity,
		IsActive:           req.IsActive,
		RecurrenceInterval: req.RecurrenceInterval,
		RecurrenceEndType:  req.RecurrenceEndType,
		RecurrenceEndCount: req.RecurrenceEndCount,
	}

	// Parse recurrence end date
	if req.RecurrenceEndDate != nil && *req.RecurrenceEndDate != "" {
		t, err := time.Parse("2006-01-02", *req.RecurrenceEndDate)
		if err == nil {
			slot.RecurrenceEndDate = &t
		}
	}

	// Parse effective start date
	if req.EffectiveStartDate != nil && *req.EffectiveStartDate != "" {
		t, err := time.Parse("2006-01-02", *req.EffectiveStartDate)
		if err == nil {
			slot.EffectiveStartDate = &t
		}
	}

	if err := h.schedulingService.UpdateScheduleSlot(userID, slot); err != nil {
		if err == service.ErrScheduleSlotNotFound {
			respondError(w, http.StatusNotFound, "Schedule slot not found")
			return
		}
		if err == service.ErrInvalidDayOfWeek {
			respondError(w, http.StatusBadRequest, "Day of week must be 0-6 (Sunday-Saturday)")
			return
		}
		h.logger.Error("Failed to update schedule slot: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update schedule slot")
		return
	}

	respondJSON(w, http.StatusOK, slot)
}

// DeleteScheduleSlot handles DELETE /api/templates/{template_id}/slots/{id}
func (h *SchedulingHandler) DeleteScheduleSlot(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "slot_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid slot ID")
		return
	}

	if err := h.schedulingService.DeleteScheduleSlot(userID, id); err != nil {
		if err == service.ErrScheduleSlotNotFound {
			respondError(w, http.StatusNotFound, "Schedule slot not found")
			return
		}
		h.logger.Error("Failed to delete schedule slot: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to delete schedule slot")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Schedule slot deleted successfully"})
}

// ============================================
// CLASS SESSION HANDLERS
// ============================================

// CreateSession handles POST /api/gyms/{gym_id}/sessions
func (h *SchedulingHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	var req struct {
		TemplateID  *int64  `json:"template_id"`
		LocationID  *int64  `json:"location_id"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
		WorkoutID   *int64  `json:"workout_id"`
		StartTime   string  `json:"start_time"`
		EndTime     string  `json:"end_time"`
		Capacity    int     `json:"capacity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.StartTime == "" {
		respondError(w, http.StatusBadRequest, "Start time is required")
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid start time format (use RFC3339)")
		return
	}

	var endTime time.Time
	if req.EndTime != "" {
		endTime, err = time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid end time format (use RFC3339)")
			return
		}
	} else {
		// Default to 1 hour session
		endTime = startTime.Add(time.Hour)
	}

	session := &domain.ClassSession{
		OrganizationID: gymID,
		TemplateID:     req.TemplateID,
		LocationID:     req.LocationID,
		Name:           req.Name,
		Description:    req.Description,
		WorkoutID:      req.WorkoutID,
		StartTime:      startTime,
		EndTime:        endTime,
		Capacity:       req.Capacity,
	}

	if err := h.schedulingService.CreateSession(userID, session); err != nil {
		if err == service.ErrOrganizationNotFound {
			respondError(w, http.StatusNotFound, "Organization not found")
			return
		}
		h.logger.Error("Failed to create session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	respondJSON(w, http.StatusCreated, session)
}

// ListSessions handles GET /api/gyms/{gym_id}/sessions
func (h *SchedulingHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	// Parse date range from query params
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid start_date format (use YYYY-MM-DD)")
			return
		}
	} else {
		// Default to start of today
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid end_date format (use YYYY-MM-DD)")
			return
		}
		// Add 1 day to include the end date
		endDate = endDate.Add(24*time.Hour - time.Second)
	} else {
		// Default to 7 days from start
		endDate = startDate.Add(7*24*time.Hour - time.Second)
	}

	// Check if user is authenticated (for user status)
	var sessions []*domain.ClassSession
	userID, hasUser := middleware.GetUserID(r.Context())

	if hasUser {
		sessions, err = h.schedulingService.GetSessionsByOrganizationForUser(gymID, userID, startDate, endDate)
	} else {
		sessions, err = h.schedulingService.GetSessionsByOrganization(gymID, startDate, endDate)
	}

	if err != nil {
		h.logger.Error("Failed to list sessions: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list sessions")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"sessions":   sessions,
		"count":      len(sessions),
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	})
}

// GetSession handles GET /api/gyms/{gym_id}/sessions/{id}
func (h *SchedulingHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Load session with full workout details
	session, err := h.schedulingService.GetSessionByIDWithWorkout(id)
	if err != nil {
		if err == service.ErrClassSessionNotFound {
			respondError(w, http.StatusNotFound, "Session not found")
			return
		}
		h.logger.Error("Failed to get session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get session")
		return
	}

	respondJSON(w, http.StatusOK, session)
}

// UpdateSession handles PUT /api/gyms/{gym_id}/sessions/{id}
func (h *SchedulingHandler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		WorkoutID   *int64  `json:"workout_id"`
		StartTime   string  `json:"start_time"`
		EndTime     string  `json:"end_time"`
		Capacity    int     `json:"capacity"`
		LocationID  *int64  `json:"location_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil && req.StartTime != "" {
		respondError(w, http.StatusBadRequest, "Invalid start time format")
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil && req.EndTime != "" {
		respondError(w, http.StatusBadRequest, "Invalid end time format")
		return
	}

	session := &domain.ClassSession{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		WorkoutID:   req.WorkoutID,
		StartTime:   startTime,
		EndTime:     endTime,
		Capacity:    req.Capacity,
		LocationID:  req.LocationID,
	}

	if err := h.schedulingService.UpdateSession(userID, session); err != nil {
		if err == service.ErrClassSessionNotFound {
			respondError(w, http.StatusNotFound, "Session not found")
			return
		}
		h.logger.Error("Failed to update session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update session")
		return
	}

	respondJSON(w, http.StatusOK, session)
}

// CancelSession handles POST /api/gyms/{gym_id}/sessions/{id}/cancel
func (h *SchedulingHandler) CancelSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.schedulingService.CancelSession(userID, id, req.Reason); err != nil {
		if err == service.ErrClassSessionNotFound {
			respondError(w, http.StatusNotFound, "Session not found")
			return
		}
		if err == service.ErrSessionNotActive {
			respondError(w, http.StatusBadRequest, "Session cannot be cancelled")
			return
		}
		h.logger.Error("Failed to cancel session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to cancel session")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Session cancelled successfully"})
}

// ============================================
// RESERVATION HANDLERS
// ============================================

// CreateReservation handles POST /api/sessions/{session_id}/reserve
func (h *SchedulingHandler) CreateReservation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	reservation, err := h.schedulingService.CreateReservation(userID, sessionID)
	if err != nil {
		if err == service.ErrClassSessionNotFound {
			respondError(w, http.StatusNotFound, "Session not found")
			return
		}
		if err == service.ErrSessionNotActive {
			respondError(w, http.StatusBadRequest, "Session is not available for reservations")
			return
		}
		if err == service.ErrAlreadyReserved {
			respondError(w, http.StatusConflict, "You already have a reservation for this session")
			return
		}
		if err == service.ErrSessionFull {
			respondError(w, http.StatusConflict, "Session is full")
			return
		}
		h.logger.Error("Failed to create reservation: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create reservation")
		return
	}

	respondJSON(w, http.StatusCreated, reservation)
}

// CancelReservation handles DELETE /api/sessions/{session_id}/reserve
func (h *SchedulingHandler) CancelReservation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	// Reason is optional
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.schedulingService.CancelReservation(userID, sessionID, req.Reason); err != nil {
		if err == service.ErrReservationNotFound {
			respondError(w, http.StatusNotFound, "Reservation not found")
			return
		}
		if err == service.ErrCannotCancelReservation {
			respondError(w, http.StatusBadRequest, "This reservation cannot be cancelled")
			return
		}
		h.logger.Error("Failed to cancel reservation: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to cancel reservation")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Reservation cancelled successfully"})
}

// GetUserUpcomingReservations handles GET /api/users/me/reservations/upcoming
func (h *SchedulingHandler) GetUserUpcomingReservations(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	reservations, err := h.schedulingService.GetUpcomingUserReservations(userID)
	if err != nil {
		h.logger.Error("Failed to get upcoming reservations: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get reservations")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"reservations": reservations,
		"count":        len(reservations),
	})
}

// GetSessionRoster handles GET /api/sessions/{session_id}/roster
func (h *SchedulingHandler) GetSessionRoster(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	reservations, err := h.schedulingService.GetReservationRoster(sessionID)
	if err != nil {
		h.logger.Error("Failed to get session roster: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get roster")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"roster": reservations,
		"count":  len(reservations),
	})
}

// CheckInReservation handles POST /api/sessions/{session_id}/check-in/{user_id}
func (h *SchedulingHandler) CheckInReservation(w http.ResponseWriter, r *http.Request) {
	coachUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	reservationIDStr := chi.URLParam(r, "reservation_id")
	reservationID, err := strconv.ParseInt(reservationIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid reservation ID")
		return
	}

	if err := h.schedulingService.CheckInReservation(coachUserID, reservationID); err != nil {
		if err == service.ErrReservationNotFound {
			respondError(w, http.StatusNotFound, "Reservation not found")
			return
		}
		h.logger.Error("Failed to check in reservation: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to check in")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Checked in successfully"})
}

// ============================================
// COACH ASSIGNMENT HANDLERS
// ============================================

// AssignCoach handles POST /api/gyms/{gym_id}/coaches
func (h *SchedulingHandler) AssignCoach(w http.ResponseWriter, r *http.Request) {
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	var req struct {
		UserID int64 `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == 0 {
		respondError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	if err := h.schedulingService.AssignCoach(adminUserID, gymID, req.UserID); err != nil {
		h.logger.Error("Failed to assign coach: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to assign coach")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"message": "Coach assigned successfully"})
}

// ListCoaches handles GET /api/gyms/{gym_id}/coaches
func (h *SchedulingHandler) ListCoaches(w http.ResponseWriter, r *http.Request) {
	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	coaches, err := h.schedulingService.GetCoachesByOrganization(gymID, includeInactive)
	if err != nil {
		h.logger.Error("Failed to list coaches: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list coaches")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"coaches": coaches,
		"count":   len(coaches),
	})
}

// UnassignCoach handles DELETE /api/gyms/{gym_id}/coaches/{id}
func (h *SchedulingHandler) UnassignCoach(w http.ResponseWriter, r *http.Request) {
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid coach assignment ID")
		return
	}

	if err := h.schedulingService.UnassignCoach(adminUserID, id); err != nil {
		if err == service.ErrCoachAssignmentNotFound {
			respondError(w, http.StatusNotFound, "Coach assignment not found")
			return
		}
		h.logger.Error("Failed to unassign coach: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to unassign coach")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Coach unassigned successfully"})
}

// ============================================
// COACH PORTAL HANDLERS
// ============================================

// GetCoachSessions handles GET /api/coaches/me/sessions
func (h *SchedulingHandler) GetCoachSessions(w http.ResponseWriter, r *http.Request) {
	coachUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	sessions, err := h.schedulingService.GetCoachUpcomingSessions(coachUserID, limit)
	if err != nil {
		h.logger.Error("Failed to get coach sessions: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get sessions")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// CompleteSession handles POST /api/sessions/{session_id}/complete
func (h *SchedulingHandler) CompleteSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	if err := h.schedulingService.CompleteSession(userID, sessionID); err != nil {
		if err == service.ErrClassSessionNotFound {
			respondError(w, http.StatusNotFound, "Session not found")
			return
		}
		if err == service.ErrSessionNotActive {
			respondError(w, http.StatusBadRequest, "Session cannot be completed")
			return
		}
		h.logger.Error("Failed to complete session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to complete session")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Session completed successfully"})
}

// AddSessionCoach handles POST /api/sessions/{session_id}/coaches
func (h *SchedulingHandler) AddSessionCoach(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var req struct {
		UserID int64 `json:"user_id"`
		IsLead bool  `json:"is_lead"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == 0 {
		respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if err := h.schedulingService.AddCoachToSession(userID, sessionID, req.UserID, req.IsLead); err != nil {
		if err == service.ErrClassSessionNotFound {
			respondError(w, http.StatusNotFound, "Session not found")
			return
		}
		h.logger.Error("Failed to add coach to session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to add coach to session")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"message": "Coach added to session"})
}

// RemoveSessionCoach handles DELETE /api/sessions/{session_id}/coaches/{user_id}
func (h *SchedulingHandler) RemoveSessionCoach(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	coachUserIDStr := chi.URLParam(r, "user_id")
	coachUserID, err := strconv.ParseInt(coachUserIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.schedulingService.RemoveCoachFromSession(userID, sessionID, coachUserID); err != nil {
		h.logger.Error("Failed to remove coach from session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to remove coach from session")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Coach removed from session"})
}

// GetSessionCoaches handles GET /api/sessions/{session_id}/coaches
func (h *SchedulingHandler) GetSessionCoaches(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	coaches, err := h.schedulingService.GetSessionCoaches(sessionID)
	if err != nil {
		h.logger.Error("Failed to get session coaches: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get session coaches")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"coaches": coaches,
		"count":   len(coaches),
	})
}

// MarkNoShow handles POST /api/sessions/{session_id}/no-show/{reservation_id}
func (h *SchedulingHandler) MarkNoShow(w http.ResponseWriter, r *http.Request) {
	coachUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	reservationIDStr := chi.URLParam(r, "reservation_id")
	reservationID, err := strconv.ParseInt(reservationIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid reservation ID")
		return
	}

	if err := h.schedulingService.MarkNoShow(coachUserID, reservationID); err != nil {
		if err == service.ErrReservationNotFound {
			respondError(w, http.StatusNotFound, "Reservation not found")
			return
		}
		h.logger.Error("Failed to mark no-show: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to mark no-show")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Marked as no-show successfully"})
}

// GetTemplateCoaches handles GET /api/templates/{id}/coaches
func (h *SchedulingHandler) GetTemplateCoaches(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	coaches, err := h.schedulingService.GetTemplateCoaches(id)
	if err != nil {
		h.logger.Error("Failed to get template coaches: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get template coaches")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"coaches": coaches,
	})
}

// AddTemplateCoach handles POST /api/templates/{id}/coaches
func (h *SchedulingHandler) AddTemplateCoach(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	templateID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	var req struct {
		UserID int64 `json:"user_id"`
		IsLead bool  `json:"is_lead"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == 0 {
		respondError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	if err := h.schedulingService.AddTemplateCoach(userID, templateID, req.UserID, req.IsLead); err != nil {
		h.logger.Error("Failed to add template coach: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to add template coach")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"message": "Coach added successfully"})
}

// RemoveTemplateCoach handles DELETE /api/templates/{id}/coaches/{user_id}
func (h *SchedulingHandler) RemoveTemplateCoach(w http.ResponseWriter, r *http.Request) {
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	templateIDStr := chi.URLParam(r, "id")
	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	userIDStr := chi.URLParam(r, "user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.schedulingService.RemoveTemplateCoach(adminUserID, templateID, userID); err != nil {
		h.logger.Error("Failed to remove template coach: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to remove template coach")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Coach removed successfully"})
}

// BatchUpdateSessionWorkout handles PUT /api/admin/gyms/{gym_id}/sessions/batch-workout
func (h *SchedulingHandler) BatchUpdateSessionWorkout(w http.ResponseWriter, r *http.Request) {
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	var req struct {
		SessionIDs []int64 `json:"session_ids"`
		WorkoutID  *int64  `json:"workout_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.SessionIDs) == 0 {
		respondError(w, http.StatusBadRequest, "session_ids is required")
		return
	}

	count, err := h.schedulingService.BatchUpdateSessionWorkout(adminUserID, gymID, req.SessionIDs, req.WorkoutID)
	if err != nil {
		h.logger.Error("Failed to batch update session workouts: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"updated_count": count,
	})
}

// ============================================
// COACH-SPECIFIC HANDLERS
// These handlers verify org-level coach access before delegating
// to existing service methods. Used by /api/coaches/ routes.
// ============================================

// CoachGetSessionRoster handles GET /api/coaches/sessions/{session_id}/roster
func (h *SchedulingHandler) CoachGetSessionRoster(w http.ResponseWriter, r *http.Request) {
	coachUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Admins bypass org check; coaches must be assigned to the session's org
	role, _ := middleware.GetUserRole(r.Context())
	if role != "admin" {
		if err := h.schedulingService.VerifyCoachAccessToSession(coachUserID, sessionID); err != nil {
			if err == service.ErrNotAuthorized {
				respondError(w, http.StatusForbidden, "Not a coach for this organization")
				return
			}
			if err == service.ErrClassSessionNotFound {
				respondError(w, http.StatusNotFound, "Session not found")
				return
			}
			h.logger.Error("Failed to verify coach access: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to verify access")
			return
		}
	}

	reservations, err := h.schedulingService.GetReservationRoster(sessionID)
	if err != nil {
		h.logger.Error("Failed to get session roster: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get roster")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"roster": reservations,
		"count":  len(reservations),
	})
}

// CoachCheckInReservation handles POST /api/coaches/sessions/{session_id}/check-in/{reservation_id}
func (h *SchedulingHandler) CoachCheckInReservation(w http.ResponseWriter, r *http.Request) {
	coachUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	reservationIDStr := chi.URLParam(r, "reservation_id")
	reservationID, err := strconv.ParseInt(reservationIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid reservation ID")
		return
	}

	role, _ := middleware.GetUserRole(r.Context())
	if role != "admin" {
		if err := h.schedulingService.VerifyCoachAccessToSession(coachUserID, sessionID); err != nil {
			if err == service.ErrNotAuthorized {
				respondError(w, http.StatusForbidden, "Not a coach for this organization")
				return
			}
			if err == service.ErrClassSessionNotFound {
				respondError(w, http.StatusNotFound, "Session not found")
				return
			}
			h.logger.Error("Failed to verify coach access: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to verify access")
			return
		}
	}

	if err := h.schedulingService.CheckInReservation(coachUserID, reservationID); err != nil {
		if err == service.ErrReservationNotFound {
			respondError(w, http.StatusNotFound, "Reservation not found")
			return
		}
		h.logger.Error("Failed to check in reservation: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to check in")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Checked in successfully"})
}

// CoachMarkNoShow handles POST /api/coaches/sessions/{session_id}/no-show/{reservation_id}
func (h *SchedulingHandler) CoachMarkNoShow(w http.ResponseWriter, r *http.Request) {
	coachUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	reservationIDStr := chi.URLParam(r, "reservation_id")
	reservationID, err := strconv.ParseInt(reservationIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid reservation ID")
		return
	}

	role, _ := middleware.GetUserRole(r.Context())
	if role != "admin" {
		if err := h.schedulingService.VerifyCoachAccessToSession(coachUserID, sessionID); err != nil {
			if err == service.ErrNotAuthorized {
				respondError(w, http.StatusForbidden, "Not a coach for this organization")
				return
			}
			if err == service.ErrClassSessionNotFound {
				respondError(w, http.StatusNotFound, "Session not found")
				return
			}
			h.logger.Error("Failed to verify coach access: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to verify access")
			return
		}
	}

	if err := h.schedulingService.MarkNoShow(coachUserID, reservationID); err != nil {
		if err == service.ErrReservationNotFound {
			respondError(w, http.StatusNotFound, "Reservation not found")
			return
		}
		h.logger.Error("Failed to mark no-show: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to mark no-show")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Marked as no-show successfully"})
}

// CoachCompleteSession handles POST /api/coaches/sessions/{session_id}/complete
func (h *SchedulingHandler) CoachCompleteSession(w http.ResponseWriter, r *http.Request) {
	coachUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	role, _ := middleware.GetUserRole(r.Context())
	if role != "admin" {
		if err := h.schedulingService.VerifyCoachAccessToSession(coachUserID, sessionID); err != nil {
			if err == service.ErrNotAuthorized {
				respondError(w, http.StatusForbidden, "Not a coach for this organization")
				return
			}
			if err == service.ErrClassSessionNotFound {
				respondError(w, http.StatusNotFound, "Session not found")
				return
			}
			h.logger.Error("Failed to verify coach access: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to verify access")
			return
		}
	}

	if err := h.schedulingService.CompleteSession(coachUserID, sessionID); err != nil {
		if err == service.ErrClassSessionNotFound {
			respondError(w, http.StatusNotFound, "Session not found")
			return
		}
		if err == service.ErrSessionNotActive {
			respondError(w, http.StatusBadRequest, "Session cannot be completed")
			return
		}
		h.logger.Error("Failed to complete session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to complete session")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Session completed successfully"})
}
