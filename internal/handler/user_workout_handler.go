package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// UserWorkoutHandler handles user workout instance endpoints (logged workouts)
type UserWorkoutHandler struct {
	userWorkoutService *service.UserWorkoutService
}

// NewUserWorkoutHandler creates a new user workout handler
func NewUserWorkoutHandler(userWorkoutService *service.UserWorkoutService) *UserWorkoutHandler {
	return &UserWorkoutHandler{
		userWorkoutService: userWorkoutService,
	}
}

// LogWorkoutRequest represents a request to log a workout instance
type LogWorkoutRequest struct {
	WorkoutID   int64   `json:"workout_id"`
	WorkoutDate string  `json:"workout_date"` // ISO 8601 format
	Notes       *string `json:"notes,omitempty"`
	TotalTime   *int    `json:"total_time,omitempty"`
	WorkoutType *string `json:"workout_type,omitempty"`
}

// LogWorkout logs that a user performed a workout template
func (h *UserWorkoutHandler) LogWorkout(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	var req LogWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.WorkoutID == 0 || req.WorkoutDate == "" {
		respondError(w, http.StatusBadRequest, "Workout ID and date are required")
		return
	}

	// Parse workout date
	workoutDate, err := time.Parse("2006-01-02", req.WorkoutDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid date format, expected YYYY-MM-DD")
		return
	}

	// Log workout
	userWorkout, err := h.userWorkoutService.LogWorkout(
		userID,
		req.WorkoutID,
		workoutDate,
		req.Notes,
		req.TotalTime,
		req.WorkoutType,
	)
	if err != nil {
		switch err {
		case service.ErrWorkoutNotFound:
			respondError(w, http.StatusNotFound, "Workout template not found")
		case service.ErrUnauthorizedWorkoutAccess:
			respondError(w, http.StatusForbidden, "You don't have access to this workout template")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to log workout")
		}
		return
	}

	respondJSON(w, http.StatusCreated, userWorkout)
}

// GetLoggedWorkout retrieves a logged workout by ID with full details
func (h *UserWorkoutHandler) GetLoggedWorkout(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	workoutIDStr := chi.URLParam(r, "id")

	workoutID, err := strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	userWorkout, err := h.userWorkoutService.GetLoggedWorkout(workoutID, userID)
	if err != nil {
		switch err {
		case service.ErrUserWorkoutNotFound:
			respondError(w, http.StatusNotFound, "Logged workout not found")
		case service.ErrUnauthorizedWorkoutAccess:
			respondError(w, http.StatusForbidden, "Unauthorized access")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to get logged workout")
		}
		return
	}

	respondJSON(w, http.StatusOK, userWorkout)
}

// ListLoggedWorkouts retrieves all logged workouts for the authenticated user
func (h *UserWorkoutHandler) ListLoggedWorkouts(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	// Parse pagination parameters
	limit := 50
	offset := 0
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

	workouts, err := h.userWorkoutService.ListLoggedWorkouts(userID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list logged workouts")
		return
	}

	respondJSON(w, http.StatusOK, workouts)
}

// UpdateLoggedWorkoutRequest represents a request to update a logged workout
type UpdateLoggedWorkoutRequest struct {
	Notes       *string `json:"notes,omitempty"`
	TotalTime   *int    `json:"total_time,omitempty"`
	WorkoutType *string `json:"workout_type,omitempty"`
}

// UpdateLoggedWorkout updates a logged workout (notes, time, type)
func (h *UserWorkoutHandler) UpdateLoggedWorkout(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	workoutIDStr := chi.URLParam(r, "id")

	workoutID, err := strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	var req UpdateLoggedWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.userWorkoutService.UpdateLoggedWorkout(
		workoutID,
		userID,
		req.Notes,
		req.TotalTime,
		req.WorkoutType,
	)
	if err != nil {
		switch err {
		case service.ErrUserWorkoutNotFound:
			respondError(w, http.StatusNotFound, "Logged workout not found")
		case service.ErrUnauthorizedWorkoutAccess:
			respondError(w, http.StatusForbidden, "Unauthorized access")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to update logged workout")
		}
		return
	}

	respondJSON(w, http.StatusOK, MessageResponse{Message: "Logged workout updated successfully"})
}

// DeleteLoggedWorkout deletes a logged workout
func (h *UserWorkoutHandler) DeleteLoggedWorkout(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	workoutIDStr := chi.URLParam(r, "id")

	workoutID, err := strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	err = h.userWorkoutService.DeleteLoggedWorkout(workoutID, userID)
	if err != nil {
		switch err {
		case service.ErrUserWorkoutNotFound:
			respondError(w, http.StatusNotFound, "Logged workout not found")
		case service.ErrUnauthorizedWorkoutAccess:
			respondError(w, http.StatusForbidden, "Unauthorized access")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to delete logged workout")
		}
		return
	}

	respondJSON(w, http.StatusOK, MessageResponse{Message: "Logged workout deleted successfully"})
}

// GetWorkoutStatsForMonth returns workout count for a specific month
func (h *UserWorkoutHandler) GetWorkoutStatsForMonth(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	if yearStr == "" || monthStr == "" {
		respondError(w, http.StatusBadRequest, "Year and month parameters are required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid year format")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		respondError(w, http.StatusBadRequest, "Invalid month format")
		return
	}

	count, err := h.userWorkoutService.GetWorkoutStatsForMonth(userID, year, month)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get workout stats")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"year":  year,
		"month": month,
		"count": count,
	})
}
