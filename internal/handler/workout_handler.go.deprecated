package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// WorkoutHandler handles workout template-related endpoints
type WorkoutHandler struct {
	workoutService *service.WorkoutService
}

// NewWorkoutHandler creates a new workout handler
func NewWorkoutHandler(workoutService *service.WorkoutService) *WorkoutHandler {
	return &WorkoutHandler{
		workoutService: workoutService,
	}
}

// CreateTemplateRequest represents a request to create a workout template
type CreateTemplateRequest struct {
	Name      string                         `json:"name"`
	Notes     *string                        `json:"notes,omitempty"`
	Movements []CreateWorkoutMovementRequest `json:"movements,omitempty"`
	WODs      []CreateWorkoutWODRequest      `json:"wods,omitempty"`
}

// CreateWorkoutMovementRequest represents a movement in a workout template
type CreateWorkoutMovementRequest struct {
	MovementID int64    `json:"movement_id"`
	Weight     *float64 `json:"weight,omitempty"`
	Sets       *int     `json:"sets,omitempty"`
	Reps       *int     `json:"reps,omitempty"`
	Time       *int     `json:"time,omitempty"`
	Distance   *float64 `json:"distance,omitempty"`
	IsRx       bool     `json:"is_rx"`
	Notes      *string  `json:"notes,omitempty"`
}

// CreateWorkoutWODRequest represents a WOD in a workout template
type CreateWorkoutWODRequest struct {
	WODID      int64   `json:"wod_id"`
	ScoreValue *string `json:"score_value,omitempty"`
	Division   *string `json:"division,omitempty"`
}

// UpdateTemplateRequest represents a request to update a workout template
type UpdateTemplateRequest struct {
	Name      string                         `json:"name"`
	Notes     *string                        `json:"notes,omitempty"`
	Movements []CreateWorkoutMovementRequest `json:"movements,omitempty"`
	WODs      []CreateWorkoutWODRequest      `json:"wods,omitempty"`
}

// TemplateResponse represents a workout template
type TemplateResponse struct {
	ID        int64                          `json:"id"`
	Name      string                         `json:"name"`
	Notes     *string                        `json:"notes,omitempty"`
	CreatedBy *int64                         `json:"created_by,omitempty"`
	CreatedAt string                         `json:"created_at"`
	UpdatedAt string                         `json:"updated_at"`
	Movements []*domain.WorkoutMovement      `json:"movements,omitempty"`
	WODs      []*domain.WorkoutWODWithDetails `json:"wods,omitempty"`
}

// CreateTemplate creates a new workout template
func (h *WorkoutHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Template name is required")
		return
	}

	// Build workout template
	workout := &domain.Workout{
		Name:  req.Name,
		Notes: req.Notes,
	}

	// Build movements
	if len(req.Movements) > 0 {
		for _, mv := range req.Movements {
			var notes string
			if mv.Notes != nil {
				notes = *mv.Notes
			}
			workoutMovement := &domain.WorkoutMovement{
				MovementID: mv.MovementID,
				Weight:     mv.Weight,
				Sets:       mv.Sets,
				Reps:       mv.Reps,
				Time:       mv.Time,
				Distance:   mv.Distance,
				IsRx:       mv.IsRx,
				Notes:      notes,
			}
			workout.Movements = append(workout.Movements, workoutMovement)
		}
	}

	// Build WODs
	if len(req.WODs) > 0 {
		for _, wod := range req.WODs {
			workoutWOD := &domain.WorkoutWODWithDetails{
				WorkoutWOD: domain.WorkoutWOD{
					WODID:      wod.WODID,
					ScoreValue: wod.ScoreValue,
					Division:   wod.Division,
				},
			}
			workout.WODs = append(workout.WODs, workoutWOD)
		}
	}

	// Create template
	if err := h.workoutService.CreateTemplate(userID, workout); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create workout template: "+err.Error())
		return
	}

	// Retrieve created template with details
	template, err := h.workoutService.GetTemplate(workout.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve created template")
		return
	}

	response := TemplateResponse{
		ID:        template.ID,
		Name:      template.Name,
		Notes:     template.Notes,
		CreatedBy: template.CreatedBy,
		CreatedAt: template.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: template.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Movements: template.Movements,
		WODs:      template.WODs,
	}

	respondJSON(w, http.StatusCreated, response)
}

// GetTemplate retrieves a workout template by ID
func (h *WorkoutHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	template, err := h.workoutService.GetTemplate(id)
	if err != nil {
		if err == service.ErrWorkoutNotFound {
			respondError(w, http.StatusNotFound, "Template not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to retrieve template")
		}
		return
	}

	response := TemplateResponse{
		ID:        template.ID,
		Name:      template.Name,
		Notes:     template.Notes,
		CreatedBy: template.CreatedBy,
		CreatedAt: template.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: template.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Movements: template.Movements,
		WODs:      template.WODs,
	}

	respondJSON(w, http.StatusOK, response)
}

// ListTemplates retrieves all workout templates (standard + user's)
func (h *WorkoutHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context (optional for browsing standard templates)
	userID, ok := middleware.GetUserID(r.Context())

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var templates []*domain.Workout
	var err error

	if ok {
		templates, err = h.workoutService.ListTemplates(&userID, limit, offset)
	} else {
		templates, err = h.workoutService.ListTemplates(nil, limit, offset)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve templates")
		return
	}

	// Build response
	var responses []TemplateResponse
	for _, template := range templates {
		response := TemplateResponse{
			ID:        template.ID,
			Name:      template.Name,
			Notes:     template.Notes,
			CreatedBy: template.CreatedBy,
			CreatedAt: template.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: template.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Movements: template.Movements,
			WODs:      template.WODs,
		}
		responses = append(responses, response)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"templates": responses,
		"limit":     limit,
		"offset":    offset,
	})
}

// UpdateTemplate updates a workout template
func (h *WorkoutHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
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

	var req UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build update
	workout := &domain.Workout{
		Name:  req.Name,
		Notes: req.Notes,
	}

	// Build movements if provided
	if req.Movements != nil {
		for _, mv := range req.Movements {
			var notes string
			if mv.Notes != nil {
				notes = *mv.Notes
			}
			workoutMovement := &domain.WorkoutMovement{
				MovementID: mv.MovementID,
				Weight:     mv.Weight,
				Sets:       mv.Sets,
				Reps:       mv.Reps,
				Time:       mv.Time,
				Distance:   mv.Distance,
				IsRx:       mv.IsRx,
				Notes:      notes,
			}
			workout.Movements = append(workout.Movements, workoutMovement)
		}
	}

	// Build WODs if provided
	if req.WODs != nil {
		for _, wod := range req.WODs {
			workoutWOD := &domain.WorkoutWODWithDetails{
				WorkoutWOD: domain.WorkoutWOD{
					WODID:      wod.WODID,
					ScoreValue: wod.ScoreValue,
					Division:   wod.Division,
				},
			}
			workout.WODs = append(workout.WODs, workoutWOD)
		}
	}

	// Update template
	if err := h.workoutService.UpdateTemplate(id, userID, workout); err != nil {
		if err == service.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "You don't have permission to update this template")
		} else if err == service.ErrWorkoutNotFound {
			respondError(w, http.StatusNotFound, "Template not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to update template: "+err.Error())
		}
		return
	}

	// Retrieve updated template
	template, err := h.workoutService.GetTemplate(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve updated template")
		return
	}

	response := TemplateResponse{
		ID:        template.ID,
		Name:      template.Name,
		Notes:     template.Notes,
		CreatedBy: template.CreatedBy,
		CreatedAt: template.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: template.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Movements: template.Movements,
		WODs:      template.WODs,
	}

	respondJSON(w, http.StatusOK, response)
}

// DeleteTemplate deletes a workout template
func (h *WorkoutHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
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

	if err := h.workoutService.DeleteTemplate(id, userID); err != nil {
		if err == service.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "You don't have permission to delete this template")
		} else if err == service.ErrWorkoutNotFound {
			respondError(w, http.StatusNotFound, "Template not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to delete template")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Template deleted successfully"})
}

// GetTemplateStats retrieves usage statistics for a template
func (h *WorkoutHandler) GetTemplateStats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	stats, err := h.workoutService.GetTemplateUsageStats(id)
	if err != nil {
		if err == service.ErrWorkoutNotFound {
			respondError(w, http.StatusNotFound, "Template not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to retrieve template stats")
		}
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// GetPersonalRecords retrieves all personal records for a user
func (h *WorkoutHandler) GetPersonalRecords(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	records, err := h.workoutService.GetPersonalRecords(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve personal records")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"records": records,
	})
}

// GetPRMovements retrieves recent PR-flagged movements for a user
func (h *WorkoutHandler) GetPRMovements(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse limit
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	movements, err := h.workoutService.GetPRMovements(userID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve PR movements")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pr_movements": movements,
	})
}

// TogglePRFlag manually toggles the PR flag on a workout movement
func (h *WorkoutHandler) TogglePRFlag(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	movementIDStr := chi.URLParam(r, "id")
	movementID, err := strconv.ParseInt(movementIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid movement ID")
		return
	}

	err = h.workoutService.TogglePRFlag(movementID, userID)
	if err != nil {
		if err == service.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "Unauthorized")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to toggle PR flag")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "PR flag toggled successfully"})
}
