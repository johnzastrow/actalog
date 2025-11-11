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

// WODHandler handles WOD (Workout of the Day) endpoints
type WODHandler struct {
	wodService *service.WODService
}

// NewWODHandler creates a new WOD handler
func NewWODHandler(wodService *service.WODService) *WODHandler {
	return &WODHandler{
		wodService: wodService,
	}
}

// CreateWODRequest represents a request to create a custom WOD
type CreateWODRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Source      *string `json:"source,omitempty"`
	Type        *string `json:"type,omitempty"`        // girl, hero, open, games
	Regime      *string `json:"regime,omitempty"`      // time, amrap, emom, tabata, strength, skill
	ScoreType   *string `json:"score_type,omitempty"`  // time, reps, weight, rounds_reps
}

// CreateWOD creates a new custom WOD
func (h *WODHandler) CreateWOD(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	var req CreateWODRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "WOD name is required")
		return
	}

	wod := &domain.WOD{
		Name: req.Name,
	}

	// Handle optional string pointer fields
	if req.Description != nil {
		wod.Description = *req.Description
	}
	if req.Source != nil {
		wod.Source = *req.Source
	}
	if req.Type != nil {
		wod.Type = *req.Type
	}
	if req.Regime != nil {
		wod.Regime = *req.Regime
	}
	if req.ScoreType != nil {
		wod.ScoreType = *req.ScoreType
	}

	err := h.wodService.CreateWOD(userID, wod)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create WOD")
		return
	}

	respondJSON(w, http.StatusCreated, wod)
}

// GetWOD retrieves a WOD by ID
func (h *WODHandler) GetWOD(w http.ResponseWriter, r *http.Request) {
	wodIDStr := chi.URLParam(r, "id")

	wodID, err := strconv.ParseInt(wodIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid WOD ID")
		return
	}

	wod, err := h.wodService.GetWOD(wodID)
	if err != nil {
		switch err {
		case service.ErrWODNotFound:
			respondError(w, http.StatusNotFound, "WOD not found")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to get WOD")
		}
		return
	}

	respondJSON(w, http.StatusOK, wod)
}

// ListWODs lists WODs (standard, user's custom, or all)
func (h *WODHandler) ListWODs(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	// Check query parameter for filtering
	filter := r.URL.Query().Get("filter")

	var wods []*domain.WOD
	var err error

	switch filter {
	case "standard":
		wods, err = h.wodService.ListStandardWODs()
	case "custom":
		wods, err = h.wodService.ListUserWODs(userID)
	default:
		// List all (standard + user's custom)
		wods, err = h.wodService.ListAllWODs(userID)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list WODs")
		return
	}

	respondJSON(w, http.StatusOK, wods)
}

// SearchWODs searches for WODs by name
func (h *WODHandler) SearchWODs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Search query parameter 'q' is required")
		return
	}

	wods, err := h.wodService.SearchWODs(query)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to search WODs")
		return
	}

	respondJSON(w, http.StatusOK, wods)
}

// UpdateWODRequest represents a request to update a WOD
type UpdateWODRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Source      *string `json:"source,omitempty"`
	Type        *string `json:"type,omitempty"`
	Regime      *string `json:"regime,omitempty"`
	ScoreType   *string `json:"score_type,omitempty"`
}

// UpdateWOD updates a custom WOD
func (h *WODHandler) UpdateWOD(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	wodIDStr := chi.URLParam(r, "id")

	wodID, err := strconv.ParseInt(wodIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid WOD ID")
		return
	}

	var req UpdateWODRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "WOD name is required")
		return
	}

	updates := &domain.WOD{
		Name: req.Name,
	}

	// Handle optional string pointer fields
	if req.Description != nil {
		updates.Description = *req.Description
	}
	if req.Source != nil {
		updates.Source = *req.Source
	}
	if req.Type != nil {
		updates.Type = *req.Type
	}
	if req.Regime != nil {
		updates.Regime = *req.Regime
	}
	if req.ScoreType != nil {
		updates.ScoreType = *req.ScoreType
	}

	err = h.wodService.UpdateWOD(wodID, userID, updates)
	if err != nil {
		switch err {
		case service.ErrWODNotFound:
			respondError(w, http.StatusNotFound, "WOD not found")
		case service.ErrUnauthorizedWODAccess:
			respondError(w, http.StatusForbidden, "You can only update your own custom WODs")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to update WOD")
		}
		return
	}

	respondJSON(w, http.StatusOK, MessageResponse{Message: "WOD updated successfully"})
}

// DeleteWOD deletes a custom WOD
func (h *WODHandler) DeleteWOD(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	wodIDStr := chi.URLParam(r, "id")

	wodID, err := strconv.ParseInt(wodIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid WOD ID")
		return
	}

	err = h.wodService.DeleteWOD(wodID, userID)
	if err != nil {
		switch err {
		case service.ErrWODNotFound:
			respondError(w, http.StatusNotFound, "WOD not found")
		case service.ErrUnauthorizedWODAccess:
			respondError(w, http.StatusForbidden, "You can only delete your own custom WODs")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to delete WOD")
		}
		return
	}

	respondJSON(w, http.StatusOK, MessageResponse{Message: "WOD deleted successfully"})
}
