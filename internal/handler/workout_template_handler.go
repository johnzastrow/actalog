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
	Create(userID int64, name string, notes *string, movements []domain.WorkoutMovement) (*domain.Workout, error)
	GetByID(id int64) (*domain.Workout, error)
	GetByIDWithDetails(id int64) (*domain.Workout, error)
	ListByUser(userID int64, limit, offset int) ([]*domain.Workout, error)
	ListStandard(limit, offset int) ([]*domain.Workout, error)
	Update(id, userID int64, name string, notes *string, movements []domain.WorkoutMovement) (*domain.Workout, error)
	Delete(id, userID int64) error
}

type WorkoutTemplateHandler struct {
	service WorkoutTemplateService
}

func NewWorkoutTemplateHandler(service WorkoutTemplateService) *WorkoutTemplateHandler {
	return &WorkoutTemplateHandler{service: service}
}

// CreateTemplate handles POST /api/templates
func (h *WorkoutTemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name        string                   `json:"name"`
		Description *string                  `json:"description"`
		Movements   []domain.WorkoutMovement `json:"movements"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Template name is required", http.StatusBadRequest)
		return
	}

	template, err := h.service.Create(userID, req.Name, req.Description, req.Movements)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"template": template,
	})
}

// ListMyTemplates handles GET /api/workouts/my-templates
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workouts": templates,
	})
}

// ListStandardTemplates handles GET /api/templates
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": templates,
	})
}

// UpdateTemplate handles PUT /api/templates/{id}
func (h *WorkoutTemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name        string                   `json:"name"`
		Description *string                  `json:"description"`
		Movements   []domain.WorkoutMovement `json:"movements"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Template name is required", http.StatusBadRequest)
		return
	}

	template, err := h.service.Update(id, userID, req.Name, req.Description, req.Movements)
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
func (h *WorkoutTemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id, userID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
