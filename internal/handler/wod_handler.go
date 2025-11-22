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
	Source      string  `json:"source,omitempty"`
	Type        string  `json:"type,omitempty"`
	Regime      string  `json:"regime,omitempty"`
	ScoreType   string  `json:"score_type,omitempty"`
	Description string  `json:"description,omitempty"`
	URL         *string `json:"url,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

// UpdateWODRequest represents a request to update a WOD
type UpdateWODRequest struct {
	Name        string  `json:"name"`
	Source      string  `json:"source,omitempty"`
	Type        string  `json:"type,omitempty"`
	Regime      string  `json:"regime,omitempty"`
	ScoreType   string  `json:"score_type,omitempty"`
	Description string  `json:"description,omitempty"`
	URL         *string `json:"url,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

// WODResponse represents a WOD
type WODResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Source      string  `json:"source,omitempty"`
	Type        string  `json:"type,omitempty"`
	Regime      string  `json:"regime,omitempty"`
	ScoreType   string  `json:"score_type,omitempty"`
	Description string  `json:"description,omitempty"`
	URL         *string `json:"url,omitempty"`
	Notes       *string `json:"notes,omitempty"`
	IsStandard  bool    `json:"is_standard"`
	CreatedBy   *int64  `json:"created_by,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// CreateWOD creates a new custom WOD
func (h *WODHandler) CreateWOD(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

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

	// Create WOD
	wod := &domain.WOD{
		Name:        req.Name,
		Source:      req.Source,
		Type:        req.Type,
		Regime:      req.Regime,
		ScoreType:   req.ScoreType,
		Description: req.Description,
		URL:         req.URL,
		Notes:       req.Notes,
	}

	if err := h.wodService.Create(wod, userID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create WOD: "+err.Error())
		return
	}

	// Retrieve created WOD
	created, err := h.wodService.GetByID(wod.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve created WOD")
		return
	}

	response := WODResponse{
		ID:          created.ID,
		Name:        created.Name,
		Source:      created.Source,
		Type:        created.Type,
		Regime:      created.Regime,
		ScoreType:   created.ScoreType,
		Description: created.Description,
		URL:         created.URL,
		Notes:       created.Notes,
		IsStandard:  created.IsStandard,
		CreatedBy:   created.CreatedBy,
		CreatedAt:   created.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   created.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	respondJSON(w, http.StatusCreated, response)
}

// GetWOD retrieves a WOD by ID
func (h *WODHandler) GetWOD(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid WOD ID")
		return
	}

	wod, err := h.wodService.GetByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve WOD: "+err.Error())
		return
	}

	response := WODResponse{
		ID:          wod.ID,
		Name:        wod.Name,
		Source:      wod.Source,
		Type:        wod.Type,
		Regime:      wod.Regime,
		ScoreType:   wod.ScoreType,
		Description: wod.Description,
		URL:         wod.URL,
		Notes:       wod.Notes,
		IsStandard:  wod.IsStandard,
		CreatedBy:   wod.CreatedBy,
		CreatedAt:   wod.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   wod.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	respondJSON(w, http.StatusOK, response)
}

// ListWODs retrieves all WODs (standard + user's custom)
func (h *WODHandler) ListWODs(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context (optional)
	userID, ok := middleware.GetUserID(r.Context())

	var wods []*domain.WOD
	var err error

	// Parse pagination parameters
	limit := 100 // default
	offset := 0  // default

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Check if user wants only standard WODs
	standardOnly := r.URL.Query().Get("standard") == "true"

	if standardOnly {
		wods, err = h.wodService.ListStandard(limit, offset)
	} else if ok {
		userIDPtr := &userID
		wods, err = h.wodService.ListAll(userIDPtr, limit, offset)
	} else {
		wods, err = h.wodService.ListStandard(limit, offset)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve WODs")
		return
	}

	// Build response
	var responses []WODResponse
	for _, wod := range wods {
		response := WODResponse{
			ID:          wod.ID,
			Name:        wod.Name,
			Source:      wod.Source,
			Type:        wod.Type,
			Regime:      wod.Regime,
			ScoreType:   wod.ScoreType,
			Description: wod.Description,
			URL:         wod.URL,
			Notes:       wod.Notes,
			IsStandard:  wod.IsStandard,
			CreatedBy:   wod.CreatedBy,
			CreatedAt:   wod.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   wod.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"wods": responses,
	})
}

// ListStandardWODs returns only standard/pre-seeded WODs
func (h *WODHandler) ListStandardWODs(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limit := 100 // default
	offset := 0  // default

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	wods, err := h.wodService.ListStandard(limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve standard WODs")
		return
	}

	// Build response
	var responses []WODResponse
	for _, wod := range wods {
		response := WODResponse{
			ID:          wod.ID,
			Name:        wod.Name,
			Source:      wod.Source,
			Type:        wod.Type,
			Regime:      wod.Regime,
			ScoreType:   wod.ScoreType,
			Description: wod.Description,
			URL:         wod.URL,
			Notes:       wod.Notes,
			IsStandard:  wod.IsStandard,
			CreatedBy:   wod.CreatedBy,
			CreatedAt:   wod.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   wod.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"wods": responses,
	})
}

// ListMyWODs returns only the authenticated user's custom WODs
func (h *WODHandler) ListMyWODs(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse pagination parameters
	limit := 100 // default
	offset := 0  // default

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	wods, err := h.wodService.ListByUser(userID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve custom WODs")
		return
	}

	// Build response
	var responses []WODResponse
	for _, wod := range wods {
		response := WODResponse{
			ID:          wod.ID,
			Name:        wod.Name,
			Source:      wod.Source,
			Type:        wod.Type,
			Regime:      wod.Regime,
			ScoreType:   wod.ScoreType,
			Description: wod.Description,
			URL:         wod.URL,
			Notes:       wod.Notes,
			IsStandard:  wod.IsStandard,
			CreatedBy:   wod.CreatedBy,
			CreatedAt:   wod.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   wod.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"wods": responses,
	})
}

// SearchWODs searches for WODs by name
func (h *WODHandler) SearchWODs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Search query is required")
		return
	}

	wods, err := h.wodService.Search(query, 20)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to search WODs")
		return
	}

	// Build response
	var responses []WODResponse
	for _, wod := range wods {
		response := WODResponse{
			ID:          wod.ID,
			Name:        wod.Name,
			Source:      wod.Source,
			Type:        wod.Type,
			Regime:      wod.Regime,
			ScoreType:   wod.ScoreType,
			Description: wod.Description,
			URL:         wod.URL,
			Notes:       wod.Notes,
			IsStandard:  wod.IsStandard,
			CreatedBy:   wod.CreatedBy,
			CreatedAt:   wod.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   wod.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"wods": responses,
	})
}

// UpdateWOD updates a custom WOD
func (h *WODHandler) UpdateWOD(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid WOD ID")
		return
	}

	var req UpdateWODRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build update
	wod := &domain.WOD{
		ID:          id,
		Name:        req.Name,
		Source:      req.Source,
		Type:        req.Type,
		Regime:      req.Regime,
		ScoreType:   req.ScoreType,
		Description: req.Description,
		URL:         req.URL,
		Notes:       req.Notes,
	}

	if err := h.wodService.Update(wod, userID); err != nil {
		if err == service.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "You don't have permission to update this WOD")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to update WOD: "+err.Error())
		}
		return
	}

	// Retrieve updated WOD
	updated, err := h.wodService.GetByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve updated WOD")
		return
	}

	response := WODResponse{
		ID:          updated.ID,
		Name:        updated.Name,
		Source:      updated.Source,
		Type:        updated.Type,
		Regime:      updated.Regime,
		ScoreType:   updated.ScoreType,
		Description: updated.Description,
		URL:         updated.URL,
		Notes:       updated.Notes,
		IsStandard:  updated.IsStandard,
		CreatedBy:   updated.CreatedBy,
		CreatedAt:   updated.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   updated.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	respondJSON(w, http.StatusOK, response)
}

// DeleteWOD deletes a custom WOD
func (h *WODHandler) DeleteWOD(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid WOD ID")
		return
	}

	if err := h.wodService.Delete(id, userID); err != nil {
		if err == service.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "You don't have permission to delete this WOD")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to delete WOD: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "WOD deleted successfully"})
}
