package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
)

// MovementHandler handles movement-related endpoints
type MovementHandler struct {
	movementRepo domain.MovementRepository
}

// NewMovementHandler creates a new movement handler
func NewMovementHandler(movementRepo domain.MovementRepository) *MovementHandler {
	return &MovementHandler{
		movementRepo: movementRepo,
	}
}

// ListStandard returns all standard movements
func (h *MovementHandler) ListStandard(w http.ResponseWriter, r *http.Request) {
	movements, err := h.movementRepo.ListStandard()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve movements")
		return
	}

	respondJSON(w, http.StatusOK, movements)
}

// Search searches for movements by name
func (h *MovementHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	movements, err := h.movementRepo.Search(query, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to search movements")
		return
	}

	respondJSON(w, http.StatusOK, movements)
}

// GetByID returns a single movement by ID
func (h *MovementHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid movement ID")
		return
	}

	movement, err := h.movementRepo.GetByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve movement")
		return
	}

	if movement == nil {
		respondError(w, http.StatusNotFound, "Movement not found")
		return
	}

	respondJSON(w, http.StatusOK, movement)
}

// Create creates a new custom movement
func (h *MovementHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Type == "" {
		respondError(w, http.StatusBadRequest, "Name and type are required")
		return
	}

	// TODO: Get user ID from context when auth middleware is added
	// For now, custom movements without user ID
	movement := &domain.Movement{
		Name:        req.Name,
		Description: req.Description,
		Type:        domain.MovementType(req.Type),
		IsStandard:  false,
	}

	if err := h.movementRepo.Create(movement); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create movement")
		return
	}

	respondJSON(w, http.StatusCreated, movement)
}
