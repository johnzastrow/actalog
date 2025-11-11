package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// WorkoutWODHandler handles workout-WOD linking endpoints
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

// AddWODToWorkout adds a WOD to a workout template
func (h *WorkoutWODHandler) AddWODToWorkout(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	workoutIDStr := chi.URLParam(r, "id")

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

	workoutWOD, err := h.workoutWODService.AddWODToWorkout(
		workoutID,
		req.WODID,
		userID,
		req.OrderIndex,
		req.Division,
	)
	if err != nil {
		switch err {
		case service.ErrWorkoutNotFound:
			respondError(w, http.StatusNotFound, "Workout template not found")
		case service.ErrWODNotFound:
			respondError(w, http.StatusNotFound, "WOD not found")
		case service.ErrUnauthorized:
			respondError(w, http.StatusForbidden, "You can only modify your own workout templates")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to add WOD to workout")
		}
		return
	}

	respondJSON(w, http.StatusCreated, workoutWOD)
}

// RemoveWODFromWorkout removes a WOD from a workout template
func (h *WorkoutWODHandler) RemoveWODFromWorkout(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	workoutIDStr := chi.URLParam(r, "id")
	wodIDStr := chi.URLParam(r, "wod_id")

	workoutWODID, err := strconv.ParseInt(wodIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout WOD ID")
		return
	}

	// workoutIDStr is not used directly but validates URL structure
	_, err = strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	err = h.workoutWODService.RemoveWODFromWorkout(workoutWODID, userID)
	if err != nil {
		switch err {
		case service.ErrWorkoutNotFound:
			respondError(w, http.StatusNotFound, "Workout template not found")
		case service.ErrUnauthorized:
			respondError(w, http.StatusForbidden, "You can only modify your own workout templates")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to remove WOD from workout")
		}
		return
	}

	respondJSON(w, http.StatusOK, MessageResponse{Message: "WOD removed from workout successfully"})
}

// UpdateWorkoutWODRequest represents a request to update a WOD in a workout
type UpdateWorkoutWODRequest struct {
	ScoreValue *string `json:"score_value,omitempty"`
	Division   *string `json:"division,omitempty"`
}

// UpdateWorkoutWOD updates a WOD in a workout template
func (h *WorkoutWODHandler) UpdateWorkoutWOD(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	workoutIDStr := chi.URLParam(r, "id")
	wodIDStr := chi.URLParam(r, "wod_id")

	workoutWODID, err := strconv.ParseInt(wodIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout WOD ID")
		return
	}

	// workoutIDStr validates URL structure
	_, err = strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	var req UpdateWorkoutWODRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.workoutWODService.UpdateWorkoutWOD(
		workoutWODID,
		userID,
		req.ScoreValue,
		req.Division,
	)
	if err != nil {
		switch err {
		case service.ErrWorkoutNotFound:
			respondError(w, http.StatusNotFound, "Workout template not found")
		case service.ErrUnauthorized:
			respondError(w, http.StatusForbidden, "You can only modify your own workout templates")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to update workout WOD")
		}
		return
	}

	respondJSON(w, http.StatusOK, MessageResponse{Message: "Workout WOD updated successfully"})
}

// ToggleWODPR toggles the PR flag on a WOD in a workout template
func (h *WorkoutWODHandler) ToggleWODPR(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	workoutIDStr := chi.URLParam(r, "id")
	wodIDStr := chi.URLParam(r, "wod_id")

	workoutWODID, err := strconv.ParseInt(wodIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout WOD ID")
		return
	}

	// workoutIDStr validates URL structure
	_, err = strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	err = h.workoutWODService.ToggleWODPR(workoutWODID, userID)
	if err != nil {
		switch err {
		case service.ErrWorkoutNotFound:
			respondError(w, http.StatusNotFound, "Workout template not found")
		case service.ErrUnauthorized:
			respondError(w, http.StatusForbidden, "You can only modify your own workout templates")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to toggle WOD PR")
		}
		return
	}

	respondJSON(w, http.StatusOK, MessageResponse{Message: "WOD PR flag toggled successfully"})
}

// ListWODsForWorkout lists all WODs in a workout template
func (h *WorkoutWODHandler) ListWODsForWorkout(w http.ResponseWriter, r *http.Request) {
	workoutIDStr := chi.URLParam(r, "id")

	workoutID, err := strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	wods, err := h.workoutWODService.ListWODsForWorkout(workoutID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list WODs for workout")
		return
	}

	respondJSON(w, http.StatusOK, wods)
}
