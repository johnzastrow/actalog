package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

type WorkoutTemplateService interface {
	Create(userID int64, userEmail, userRole, name string, introWarmup, notes *string, isStandard bool, movements []domain.WorkoutMovement, wods []domain.WorkoutWOD) (*domain.Workout, error)
	GetByID(id int64) (*domain.Workout, error)
	GetByIDWithDetails(id int64) (*domain.Workout, error)
	ListByUser(userID int64, limit, offset int) ([]*domain.Workout, error)
	ListStandard(limit, offset int) ([]*domain.Workout, error)
	Update(id, userID int64, userEmail, userRole, name string, introWarmup, notes *string, isStandard bool, movements []domain.WorkoutMovement, wods []domain.WorkoutWOD) (*domain.Workout, error)
	Delete(id, userID int64, userEmail, userRole string) error
}

type WorkoutTemplateHandler struct {
	service WorkoutTemplateService
}

func NewWorkoutTemplateHandler(service WorkoutTemplateService) *WorkoutTemplateHandler {
	return &WorkoutTemplateHandler{service: service}
}

// CreateTemplate handles POST /api/templates
// @Summary      Create workout template
// @Description  Create a new reusable workout template with movements and WODs
// @Tags         templates
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object true "Template details with movements and WODs"
// @Success      201 {object} map[string]interface{} "Created template"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /templates [post]
func (h *WorkoutTemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userEmail, _ := middleware.GetUserEmail(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	var req struct {
		Name        string  `json:"name"`
		WorkoutType string  `json:"workout_type"` // Accept but ignore for now
		IntroWarmup *string `json:"intro_warmup"`
		Notes       *string `json:"notes"`
		IsStandard  bool    `json:"is_standard"` // Only admins can create standard workouts
		Movements   []struct {
			MovementID   int64    `json:"movement_id"`
			Sets         *int     `json:"sets"`
			Reps         *int     `json:"reps"`
			Weight       *float64 `json:"weight"`
			Time         *int     `json:"time"`      // Duration in seconds (alternate name)
			WorkTime     *int     `json:"work_time"` // Work duration in seconds
			Distance     *float64 `json:"distance"`
			IsRx         bool     `json:"is_rx"`        // Prescribed standard flag
			IsPR         bool     `json:"is_pr"`        // Personal record flag
			Instructions string   `json:"instructions"` // Markdown instructions
			Notes        string   `json:"notes"`
			OrderIndex   int      `json:"order_index"`
		} `json:"movements"`
		WODs []struct {
			WODID        int64   `json:"wod_id"`
			ScoreValue   *string `json:"score_value"`  // Time, rounds+reps, etc.
			Division     *string `json:"division"`     // rx, scaled, beginner
			IsPR         bool    `json:"is_pr"`        // Personal record flag
			Instructions string  `json:"instructions"` // Markdown instructions
			Notes        string  `json:"notes"`
			OrderIndex   int     `json:"order_index"`
		} `json:"wods"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Template name is required", http.StatusBadRequest)
		return
	}

	// Convert request movements to domain movements
	movements := make([]domain.WorkoutMovement, len(req.Movements))
	for i, m := range req.Movements {
		// Use Time if provided, otherwise fall back to WorkTime
		timeVal := m.Time
		if timeVal == nil {
			timeVal = m.WorkTime
		}
		movements[i] = domain.WorkoutMovement{
			MovementID:   m.MovementID,
			Sets:         m.Sets,
			Reps:         m.Reps,
			Weight:       m.Weight,
			Time:         timeVal,
			Distance:     m.Distance,
			IsRx:         m.IsRx,
			IsPR:         m.IsPR,
			Instructions: m.Instructions,
			Notes:        m.Notes,
			OrderIndex:   m.OrderIndex,
		}
	}

	// Convert request WODs to domain WODs
	wods := make([]domain.WorkoutWOD, len(req.WODs))
	for i, w := range req.WODs {
		wods[i] = domain.WorkoutWOD{
			WODID:        w.WODID,
			ScoreValue:   w.ScoreValue,
			Division:     w.Division,
			IsPR:         w.IsPR,
			Instructions: w.Instructions,
			Notes:        w.Notes,
			OrderIndex:   w.OrderIndex,
		}
	}

	template, err := h.service.Create(userID, userEmail, userRole, req.Name, req.IntroWarmup, req.Notes, req.IsStandard, movements, wods)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"template": template,
	})
}

// GetTemplate handles GET /api/templates/{id}
// @Summary      Get workout template
// @Description  Retrieve a specific workout template by ID with full details
// @Tags         templates
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Template ID"
// @Success      200 {object} map[string]interface{} "Template details"
// @Failure      400 {object} ErrorResponse "Invalid template ID"
// @Failure      404 {object} ErrorResponse "Template not found"
// @Router       /templates/{id} [get]
func (h *WorkoutTemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	template, err := h.service.GetByIDWithDetails(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if template == nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"template": template,
	})
}

// ListMyTemplates handles GET /api/workouts/my-templates
// @Summary      List my workout templates
// @Description  Get all workout templates created by the current user
// @Tags         templates
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Max results (default 100)"
// @Param        offset query int false "Skip N results (default 0)"
// @Success      200 {object} map[string]interface{} "List of user's templates"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /workouts/my-templates [get]
func (h *WorkoutTemplateHandler) ListMyTemplates(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit := 100
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	templates, err := h.service.ListByUser(userID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure we always return an array, not null
	if templates == nil {
		templates = []*domain.Workout{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workouts": templates,
	})
}

// ListStandardTemplates handles GET /api/templates
// @Summary      List standard workout templates
// @Description  Get all standard (system-provided) workout templates
// @Tags         templates
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Max results (default 100)"
// @Param        offset query int false "Skip N results (default 0)"
// @Success      200 {object} map[string]interface{} "List of standard templates"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /templates [get]
func (h *WorkoutTemplateHandler) ListStandardTemplates(w http.ResponseWriter, r *http.Request) {
	limit := 100
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	templates, err := h.service.ListStandard(limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure we always return an array, not null
	if templates == nil {
		templates = []*domain.Workout{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workouts": templates,
	})
}

// UpdateTemplate handles PUT /api/templates/{id}
// @Summary      Update workout template
// @Description  Update an existing workout template (owner only)
// @Tags         templates
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Template ID"
// @Param        request body object true "Updated template details"
// @Success      200 {object} map[string]interface{} "Updated template"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /templates/{id} [put]
func (h *WorkoutTemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userEmail, _ := middleware.GetUserEmail(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name        string  `json:"name"`
		WorkoutType string  `json:"workout_type"` // Accept but ignore for now
		IntroWarmup *string `json:"intro_warmup"`
		Notes       *string `json:"notes"`
		IsStandard  bool    `json:"is_standard"` // Only admins can change this
		Movements   []struct {
			MovementID   int64    `json:"movement_id"`
			Sets         *int     `json:"sets"`
			Reps         *int     `json:"reps"`
			Weight       *float64 `json:"weight"`
			Time         *int     `json:"time"`      // Duration in seconds (alternate name)
			WorkTime     *int     `json:"work_time"` // Work duration in seconds
			Distance     *float64 `json:"distance"`
			IsRx         bool     `json:"is_rx"`        // Prescribed standard flag
			IsPR         bool     `json:"is_pr"`        // Personal record flag
			Instructions string   `json:"instructions"` // Markdown instructions
			Notes        string   `json:"notes"`
			OrderIndex   int      `json:"order_index"`
		} `json:"movements"`
		WODs []struct {
			WODID        int64   `json:"wod_id"`
			ScoreValue   *string `json:"score_value"`  // Time, rounds+reps, etc.
			Division     *string `json:"division"`     // rx, scaled, beginner
			IsPR         bool    `json:"is_pr"`        // Personal record flag
			Instructions string  `json:"instructions"` // Markdown instructions
			Notes        string  `json:"notes"`
			OrderIndex   int     `json:"order_index"`
		} `json:"wods"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Template name is required", http.StatusBadRequest)
		return
	}

	// Convert request movements to domain movements
	movements := make([]domain.WorkoutMovement, len(req.Movements))
	for i, m := range req.Movements {
		// Use Time if provided, otherwise fall back to WorkTime
		timeVal := m.Time
		if timeVal == nil {
			timeVal = m.WorkTime
		}
		movements[i] = domain.WorkoutMovement{
			MovementID:   m.MovementID,
			Sets:         m.Sets,
			Reps:         m.Reps,
			Weight:       m.Weight,
			Time:         timeVal,
			Distance:     m.Distance,
			IsRx:         m.IsRx,
			IsPR:         m.IsPR,
			Instructions: m.Instructions,
			Notes:        m.Notes,
			OrderIndex:   m.OrderIndex,
		}
	}

	// Convert request WODs to domain WODs
	wods := make([]domain.WorkoutWOD, len(req.WODs))
	for i, w := range req.WODs {
		wods[i] = domain.WorkoutWOD{
			WODID:        w.WODID,
			ScoreValue:   w.ScoreValue,
			Division:     w.Division,
			IsPR:         w.IsPR,
			Instructions: w.Instructions,
			Notes:        w.Notes,
			OrderIndex:   w.OrderIndex,
		}
	}

	template, err := h.service.Update(id, userID, userEmail, userRole, req.Name, req.IntroWarmup, req.Notes, req.IsStandard, movements, wods)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"template": template,
	})
}

// DeleteTemplate handles DELETE /api/templates/{id}
// @Summary      Delete workout template
// @Description  Delete a workout template (owner only)
// @Tags         templates
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Template ID"
// @Success      204 "Template deleted"
// @Failure      400 {object} ErrorResponse "Invalid template ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Not the template owner"
// @Router       /templates/{id} [delete]
func (h *WorkoutTemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userEmail, _ := middleware.GetUserEmail(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id, userID, userEmail, userRole); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
