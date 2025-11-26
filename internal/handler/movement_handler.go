package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// MovementHandler handles movement-related endpoints
type MovementHandler struct {
	movementRepo    domain.MovementRepository
	movementService *service.MovementService
	logger          *logger.Logger
}

// NewMovementHandler creates a new movement handler
func NewMovementHandler(movementRepo domain.MovementRepository, movementService *service.MovementService, l *logger.Logger) *MovementHandler {
	return &MovementHandler{
		movementRepo:    movementRepo,
		movementService: movementService,
		logger:          l,
	}
}

// ListAll returns all movements (both standard and custom)
func (h *MovementHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	movements, err := h.movementRepo.ListAll()
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=list_all_movements outcome=failure error=%v", err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve movements")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"movements": movements,
	})
}

// ListStandard returns all standard movements
func (h *MovementHandler) ListStandard(w http.ResponseWriter, r *http.Request) {
	movements, err := h.movementRepo.ListStandard()
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=list_movements outcome=failure error=%v", err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve movements")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"movements": movements,
	})
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

	if h.logger != nil {
		h.logger.Info("action=search_movements query=%s limit=%d", query, limit)
	}

	movements, err := h.movementRepo.Search(query, limit)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=search_movements outcome=failure query=%s error=%v", query, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to search movements")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"movements": movements,
	})
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
		if h.logger != nil {
			h.logger.Error("action=get_movement outcome=failure id=%d error=%v", id, err)
		}
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

	if h.logger != nil {
		h.logger.Info("action=create_movement_attempt name=%s type=%s", req.Name, req.Type)
	}

	if err := h.movementRepo.Create(movement); err != nil {
		if h.logger != nil {
			h.logger.Error("action=create_movement outcome=failure name=%s error=%v", req.Name, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to create movement")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=create_movement outcome=success id=%d name=%s", movement.ID, movement.Name)
	}

	respondJSON(w, http.StatusCreated, movement)
}

// Update updates an existing custom movement
func (h *MovementHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Extract user ID and email from context for audit logging
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userEmail, _ := middleware.GetUserEmail(r.Context())

	// Check if user is admin
	role, _ := middleware.GetUserRole(r.Context())
	isAdmin := role == "admin"

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid movement ID")
		return
	}

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

	movement := &domain.Movement{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Type:        domain.MovementType(req.Type),
	}

	if h.logger != nil {
		h.logger.Info("action=update_movement_attempt id=%d name=%s", id, req.Name)
	}

	// Use admin update if admin, otherwise regular update
	if isAdmin {
		err = h.movementService.UpdateAsAdmin(movement, userID, userEmail)
	} else {
		err = h.movementService.Update(movement, userID, userEmail)
	}
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=update_movement outcome=failure id=%d error=%v", id, err)
		}
		if err == service.ErrMovementUnauthorized {
			respondError(w, http.StatusForbidden, "Cannot modify standard movement")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to update movement: "+err.Error())
		}
		return
	}

	if h.logger != nil {
		h.logger.Info("action=update_movement outcome=success id=%d", id)
	}

	// Retrieve updated movement
	updated, _ := h.movementService.GetByID(id)
	if updated != nil {
		respondJSON(w, http.StatusOK, updated)
	} else {
		respondJSON(w, http.StatusOK, movement)
	}
}

// Delete deletes a custom movement
func (h *MovementHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Extract user ID and email from context for audit logging
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userEmail, _ := middleware.GetUserEmail(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid movement ID")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=delete_movement_attempt id=%d", id)
	}

	if err := h.movementService.Delete(id, userID, userEmail); err != nil {
		if h.logger != nil {
			h.logger.Error("action=delete_movement outcome=failure id=%d error=%v", id, err)
		}
		if err == service.ErrMovementUnauthorized {
			respondError(w, http.StatusForbidden, "Cannot delete standard movement")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to delete movement: "+err.Error())
		}
		return
	}

	if h.logger != nil {
		h.logger.Info("action=delete_movement outcome=success id=%d", id)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Movement deleted successfully",
	})
}
