package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// WorkoutWODHandler handles linking WODs to workout templates
type WorkoutWODHandler struct {
	workoutWODService *service.WorkoutWODService
}

// NewWorkoutWODHandler creates a new workout WOD handler
func NewWorkoutWODHandler(workoutWODService *service.WorkoutWODService) *WorkoutWODHandler {
	return &WorkoutWODHandler{
		workoutWODService: workoutWODService,
	}
}

// AddWODToWorkoutRequest represents a request to add a WOD to a workout template
type AddWODToWorkoutRequest struct {
	WODID      int64   `json:"wod_id"`
	OrderIndex int     `json:"order_index"`
	Division   *string `json:"division,omitempty"`
}

// UpdateWorkoutWODRequest represents a request to update a WOD in a workout
type UpdateWorkoutWODRequest struct {
	ScoreValue *string `json:"score_value,omitempty"`
	Division   *string `json:"division,omitempty"`
}

// AddWODToWorkout adds a WOD to a workout template
func (h *WorkoutWODHandler) AddWODToWorkout(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get workout ID from URL
	workoutIDStr := chi.URLParam(r, "workout_id")
	workoutID, err := strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	var req AddWODToWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.WODID == 0 {
		respondError(w, http.StatusBadRequest, "WOD ID is required")
		return
	}

	// Add WOD to workout
	workoutWOD, err := h.workoutWODService.AddWODToWorkout(workoutID, req.WODID, userID, req.OrderIndex, req.Division)
	if err != nil {
		if err == service.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "You don't have permission to modify this workout")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to add WOD to workout: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusCreated, workoutWOD)
}

// RemoveWODFromWorkout removes a WOD from a workout template
func (h *WorkoutWODHandler) RemoveWODFromWorkout(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get workout WOD ID from URL
	workoutWODIDStr := chi.URLParam(r, "workout_wod_id")
	workoutWODID, err := strconv.ParseInt(workoutWODIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout WOD ID")
		return
	}

	// Remove WOD from workout
	if err := h.workoutWODService.RemoveWODFromWorkout(workoutWODID, userID); err != nil {
		if err == service.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "You don't have permission to modify this workout")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to remove WOD from workout: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "WOD removed from workout successfully"})
}

// UpdateWorkoutWOD updates a WOD in a workout template
func (h *WorkoutWODHandler) UpdateWorkoutWOD(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get workout WOD ID from URL
	workoutWODIDStr := chi.URLParam(r, "workout_wod_id")
	workoutWODID, err := strconv.ParseInt(workoutWODIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout WOD ID")
		return
	}

	var req UpdateWorkoutWODRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update workout WOD
	if err := h.workoutWODService.UpdateWorkoutWOD(workoutWODID, userID, req.ScoreValue, req.Division); err != nil {
		if err == service.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "You don't have permission to modify this workout")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to update workout WOD: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Workout WOD updated successfully"})
}

// ToggleWODPR toggles the PR flag on a WOD in a workout
func (h *WorkoutWODHandler) ToggleWODPR(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get workout WOD ID from URL
	workoutWODIDStr := chi.URLParam(r, "workout_wod_id")
	workoutWODID, err := strconv.ParseInt(workoutWODIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout WOD ID")
		return
	}

	// Toggle PR flag
	if err := h.workoutWODService.ToggleWODPR(workoutWODID, userID); err != nil {
		if err == service.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "You don't have permission to modify this workout")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to toggle WOD PR: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "WOD PR flag toggled successfully"})
}

// ListWODsForWorkout retrieves all WODs for a workout template
func (h *WorkoutWODHandler) ListWODsForWorkout(w http.ResponseWriter, r *http.Request) {
	// Get workout ID from URL
	workoutIDStr := chi.URLParam(r, "workout_id")
	workoutID, err := strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	wods, err := h.workoutWODService.ListWODsForWorkout(workoutID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve WODs for workout")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"wods": wods,
	})
}
