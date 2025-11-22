package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
	"github.com/johnzastrow/actalog/pkg/prmath"
)

// PerformanceHandler handles performance tracking endpoints
type PerformanceHandler struct {
	movementRepo            *repository.MovementRepository
	wodRepo                 *repository.WODRepository
	userWorkoutMovementRepo *repository.UserWorkoutMovementRepository
	userWorkoutWODRepo      *repository.UserWorkoutWODRepository
	logger                  *logger.Logger
}

// MovementPerformanceWithRM extends UserWorkoutMovement with calculated 1RM
type MovementPerformanceWithRM struct {
	*domain.UserWorkoutMovement
	Calculated1RM *float64 `json:"calculated_1rm,omitempty"`
	Formula       *string  `json:"formula,omitempty"`
}

// NewPerformanceHandler creates a new performance handler
func NewPerformanceHandler(
	movementRepo *repository.MovementRepository,
	wodRepo *repository.WODRepository,
	userWorkoutMovementRepo *repository.UserWorkoutMovementRepository,
	userWorkoutWODRepo      *repository.UserWorkoutWODRepository,
	logger *logger.Logger,
) *PerformanceHandler {
	return &PerformanceHandler{
		movementRepo:            movementRepo,
		wodRepo:                 wodRepo,
		userWorkoutMovementRepo: userWorkoutMovementRepo,
		userWorkoutWODRepo:      userWorkoutWODRepo,
		logger:                  logger,
	}
}

// UnifiedSearch searches both movements and WODs
func (h *PerformanceHandler) UnifiedSearch(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (for authorization)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Search query is required")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if h.logger != nil {
		h.logger.Info("action=unified_search user_id=%d query=%s limit=%d", userID, query, limit)
	}

	// Search movements
	movements, err := h.movementRepo.Search(query, limit)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=unified_search outcome=failure user_id=%d error=%v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to search movements")
		return
	}

	// Search WODs
	wods, err := h.wodRepo.Search(query, limit)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=unified_search outcome=failure user_id=%d error=%v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to search WODs")
		return
	}

	// Format results
	type SearchResult struct {
		Type string      `json:"type"` // "movement" or "wod"
		ID   int64       `json:"id"`
		Name string      `json:"name"`
		Data interface{} `json:"data"` // Full movement or WOD object
	}

	results := []SearchResult{}

	for _, m := range movements {
		results = append(results, SearchResult{
			Type: "movement",
			ID:   m.ID,
			Name: m.Name,
			Data: m,
		})
	}

	for _, w := range wods {
		results = append(results, SearchResult{
			Type: "wod",
			ID:   w.ID,
			Name: w.Name,
			Data: w,
		})
	}

	if h.logger != nil {
		h.logger.Info("action=unified_search outcome=success user_id=%d movements=%d wods=%d", userID, len(movements), len(wods))
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"count":   len(results),
	})
}

// GetMovementPerformance retrieves all performance history for a specific movement with calculated 1RM
func (h *PerformanceHandler) GetMovementPerformance(w http.ResponseWriter, r *http.Request) {
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

	if h.logger != nil {
		h.logger.Info("action=get_movement_performance user_id=%d movement_id=%d", userID, movementID)
	}

	// Get all performance records for this movement
	performances, err := h.userWorkoutMovementRepo.GetByUserIDAndMovementID(userID, movementID, 1000)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=get_movement_performance outcome=failure user_id=%d error=%v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve movement performance")
		return
	}

	// Calculate 1RM for each performance and find the best
	performancesWithRM := make([]MovementPerformanceWithRM, 0, len(performances))
	var best1RM *float64
	var bestFormula *string

	for _, perf := range performances {
		perfWithRM := MovementPerformanceWithRM{
			UserWorkoutMovement: perf,
		}

		// Calculate 1RM if weight and reps are present
		if perf.Weight != nil && perf.Reps != nil && *perf.Weight > 0 && *perf.Reps > 0 {
			oneRM, formula := prmath.Calculate1RM(*perf.Weight, *perf.Reps)
			perfWithRM.Calculated1RM = &oneRM
			formulaStr := string(formula)
			perfWithRM.Formula = &formulaStr

			// Track best 1RM
			if best1RM == nil || oneRM > *best1RM {
				best1RM = &oneRM
				bestFormula = &formulaStr
			}
		}

		performancesWithRM = append(performancesWithRM, perfWithRM)
	}

	if h.logger != nil {
		h.logger.Info("action=get_movement_performance outcome=success user_id=%d movement_id=%d records=%d", userID, movementID, len(performances))
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"performances": performancesWithRM,
		"count":        len(performancesWithRM),
		"best_1rm":     best1RM,
		"best_formula": bestFormula,
	})
}

// GetWODPerformance retrieves all performance history for a specific WOD
func (h *PerformanceHandler) GetWODPerformance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	wodIDStr := chi.URLParam(r, "id")
	wodID, err := strconv.ParseInt(wodIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid WOD ID")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=get_wod_performance user_id=%d wod_id=%d", userID, wodID)
	}

	// Get all performance records for this WOD
	performances, err := h.userWorkoutWODRepo.GetByUserIDAndWODID(userID, wodID, 1000)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=get_wod_performance outcome=failure user_id=%d error=%v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve WOD performance")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=get_wod_performance outcome=success user_id=%d wod_id=%d records=%d", userID, wodID, len(performances))
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"performances": performances,
		"count":        len(performances),
	})
}
