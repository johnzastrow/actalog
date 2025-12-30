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
// @Description Request to create a new WOD
type CreateWODRequest struct {
	Name        string  `json:"name" example:"My Custom WOD"`
	Source      string  `json:"source,omitempty" example:"custom"`
	Type        string  `json:"type,omitempty" example:"AMRAP"`
	Regime      string  `json:"regime,omitempty" example:"15 min"`
	ScoreType   string  `json:"score_type,omitempty" example:"rounds+reps"`
	Description string  `json:"description,omitempty" example:"15 min AMRAP of 10 KB swings, 15 box jumps"`
	URL         *string `json:"url,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

// UpdateWODRequest represents a request to update a WOD
// @Description Request to update an existing WOD
type UpdateWODRequest struct {
	Name        string  `json:"name" example:"Updated WOD Name"`
	Source      string  `json:"source,omitempty" example:"custom"`
	Type        string  `json:"type,omitempty" example:"For Time"`
	Regime      string  `json:"regime,omitempty" example:"21-15-9"`
	ScoreType   string  `json:"score_type,omitempty" example:"time"`
	Description string  `json:"description,omitempty" example:"21-15-9 thrusters and pull-ups"`
	URL         *string `json:"url,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

// WODResponse represents a WOD
// @Description WOD details response
type WODResponse struct {
	ID          int64   `json:"id" example:"1"`
	Name        string  `json:"name" example:"Fran"`
	Source      string  `json:"source,omitempty" example:"CrossFit"`
	Type        string  `json:"type,omitempty" example:"For Time"`
	Regime      string  `json:"regime,omitempty" example:"21-15-9"`
	ScoreType   string  `json:"score_type,omitempty" example:"time"`
	Description string  `json:"description,omitempty" example:"21-15-9 thrusters and pull-ups"`
	URL         *string `json:"url,omitempty"`
	Notes       *string `json:"notes,omitempty"`
	IsStandard  bool    `json:"is_standard" example:"true"`
	CreatedBy   *int64  `json:"created_by,omitempty"`
	CreatedAt   string  `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt   string  `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// WODListResponse represents a list of WODs
// @Description List of WODs
type WODListResponse struct {
	WODs []WODResponse `json:"wods"`
}

// CreateWOD creates a new custom WOD
// @Summary      Create a custom WOD
// @Description  Create a new custom WOD (Workout of the Day)
// @Tags         wods
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateWODRequest true "WOD details"
// @Success      201 {object} WODResponse "Created WOD"
// @Failure      400 {object} ErrorResponse "Invalid request body or missing name"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /wods [post]
func (h *WODHandler) CreateWOD(w http.ResponseWriter, r *http.Request) {
	// Extract user ID and email from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userEmail, _ := middleware.GetUserEmail(r.Context())

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

	if err := h.wodService.Create(wod, userID, userEmail); err != nil {
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
// @Summary      Get WOD by ID
// @Description  Retrieve a single WOD by its ID
// @Tags         wods
// @Accept       json
// @Produce      json
// @Param        id path int true "WOD ID"
// @Success      200 {object} WODResponse "WOD details"
// @Failure      400 {object} ErrorResponse "Invalid WOD ID"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /wods/{id} [get]
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
// @Summary      List all WODs
// @Description  Retrieve all WODs including standard and user's custom WODs
// @Tags         wods
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Max results (default 100)"
// @Param        offset query int false "Skip N results (default 0)"
// @Param        standard query bool false "Return only standard WODs"
// @Success      200 {object} WODListResponse "List of WODs"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /wods [get]
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
// @Summary      List standard WODs
// @Description  Retrieve only standard (built-in) WODs
// @Tags         wods
// @Accept       json
// @Produce      json
// @Param        limit query int false "Max results (default 100)"
// @Param        offset query int false "Skip N results (default 0)"
// @Success      200 {object} WODListResponse "List of standard WODs"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /wods/standard [get]
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
// @Summary      List my custom WODs
// @Description  Retrieve only the authenticated user's custom WODs
// @Tags         wods
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Max results (default 100)"
// @Param        offset query int false "Skip N results (default 0)"
// @Success      200 {object} WODListResponse "List of user's custom WODs"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /wods/mine [get]
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
// @Summary      Search WODs
// @Description  Search for WODs by name
// @Tags         wods
// @Accept       json
// @Produce      json
// @Param        q query string true "Search query"
// @Success      200 {object} WODListResponse "Matching WODs"
// @Failure      400 {object} ErrorResponse "Missing search query"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /wods/search [get]
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

// UpdateWOD updates a custom WOD (or any WOD if admin)
// @Summary      Update a WOD
// @Description  Update an existing WOD (admin can update any, users only their own custom WODs)
// @Tags         wods
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "WOD ID"
// @Param        request body UpdateWODRequest true "Updated WOD details"
// @Success      200 {object} WODResponse "Updated WOD"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden - not your WOD"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /wods/{id} [put]
func (h *WODHandler) UpdateWOD(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract user email for audit logging
	userEmail, _ := middleware.GetUserEmail(r.Context())

	// Check if user is admin
	role, _ := middleware.GetUserRole(r.Context())
	isAdmin := role == "admin"

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

	// Use admin update if user is admin, otherwise regular update
	if isAdmin {
		err = h.wodService.UpdateAsAdmin(wod, userID, userEmail)
	} else {
		err = h.wodService.Update(wod, userID, userEmail)
	}
	if err != nil {
		if err == service.ErrUnauthorized || err == service.ErrWODUnauthorized || err == service.ErrWODOwnership {
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
// @Summary      Delete a WOD
// @Description  Delete a custom WOD (cannot delete standard WODs)
// @Tags         wods
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "WOD ID"
// @Success      200 {object} MessageResponse "WOD deleted successfully"
// @Failure      400 {object} ErrorResponse "Invalid WOD ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden - cannot delete this WOD"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /wods/{id} [delete]
func (h *WODHandler) DeleteWOD(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract user email for audit logging
	userEmail, _ := middleware.GetUserEmail(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid WOD ID")
		return
	}

	if err := h.wodService.Delete(id, userID, userEmail); err != nil {
		if err == service.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "You don't have permission to delete this WOD")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to delete WOD: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "WOD deleted successfully"})
}
