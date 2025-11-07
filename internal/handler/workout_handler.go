package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// WorkoutHandler handles workout endpoints
type WorkoutHandler struct {
	workoutService *service.WorkoutService
}

// NewWorkoutHandler creates a new workout handler
func NewWorkoutHandler(workoutService *service.WorkoutService) *WorkoutHandler {
	return &WorkoutHandler{
		workoutService: workoutService,
	}
}

// CreateWorkoutRequest represents a request to create a workout
type CreateWorkoutRequest struct {
	WorkoutDate string                      `json:"workout_date"` // YYYY-MM-DD format
	WorkoutType string                      `json:"workout_type"` // named_wod or custom
	WorkoutName string                      `json:"workout_name,omitempty"`
	Notes       string                      `json:"notes,omitempty"`
	TotalTime   *int                        `json:"total_time,omitempty"` // in seconds
	Movements   []CreateWorkoutMovementRequest `json:"movements,omitempty"`
}

// CreateWorkoutMovementRequest represents a movement in a workout
type CreateWorkoutMovementRequest struct {
	MovementID int64    `json:"movement_id"`
	Weight     *float64 `json:"weight,omitempty"`
	Sets       *int     `json:"sets,omitempty"`
	Reps       *int     `json:"reps,omitempty"`
	Time       *int     `json:"time,omitempty"`
	Distance   *float64 `json:"distance,omitempty"`
	IsRx       bool     `json:"is_rx"`
	Notes      string   `json:"notes,omitempty"`
}

// CreateWorkout handles POST /api/workouts
func (h *WorkoutHandler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.WorkoutDate == "" || req.WorkoutType == "" {
		respondError(w, http.StatusBadRequest, "Workout date and type are required")
		return
	}

	// Parse workout date
	workoutDate, err := time.Parse("2006-01-02", req.WorkoutDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout date format. Use YYYY-MM-DD")
		return
	}

	// Validate workout type
	workoutType := domain.WorkoutType(req.WorkoutType)
	if workoutType != domain.WorkoutTypeNamedWOD && workoutType != domain.WorkoutTypeCustom {
		respondError(w, http.StatusBadRequest, "Invalid workout type. Use 'named_wod' or 'custom'")
		return
	}

	// Create workout domain object
	workout := &domain.Workout{
		WorkoutDate: workoutDate,
		WorkoutType: workoutType,
		WorkoutName: req.WorkoutName,
		Notes:       req.Notes,
		TotalTime:   req.TotalTime,
	}

	// Convert movements
	if len(req.Movements) > 0 {
		workout.Movements = make([]*domain.WorkoutMovement, len(req.Movements))
		now := time.Now()
		for i, m := range req.Movements {
			workout.Movements[i] = &domain.WorkoutMovement{
				MovementID: m.MovementID,
				Weight:     m.Weight,
				Sets:       m.Sets,
				Reps:       m.Reps,
				Time:       m.Time,
				Distance:   m.Distance,
				IsRx:       m.IsRx,
				Notes:      m.Notes,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
		}
	}

	// Create workout
	err = h.workoutService.CreateWorkout(userID, workout)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create workout")
		return
	}

	respondJSON(w, http.StatusCreated, workout)
}

// GetWorkout handles GET /api/workouts/{id}
func (h *WorkoutHandler) GetWorkout(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get workout ID from URL
	workoutIDStr := chi.URLParam(r, "id")
	workoutID, err := strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	// Get workout
	workout, err := h.workoutService.GetWorkout(workoutID, userID)
	if err != nil {
		switch err {
		case service.ErrWorkoutNotFound:
			respondError(w, http.StatusNotFound, "Workout not found")
		case service.ErrUnauthorized:
			respondError(w, http.StatusForbidden, "Access denied")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to get workout")
		}
		return
	}

	respondJSON(w, http.StatusOK, workout)
}

// ListWorkouts handles GET /api/workouts
func (h *WorkoutHandler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get workouts
	workouts, err := h.workoutService.ListUserWorkouts(userID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list workouts")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"workouts": workouts,
		"limit":    limit,
		"offset":   offset,
	})
}

// UpdateWorkout handles PUT /api/workouts/{id}
func (h *WorkoutHandler) UpdateWorkout(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get workout ID from URL
	workoutIDStr := chi.URLParam(r, "id")
	workoutID, err := strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	var req CreateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse workout date
	workoutDate, err := time.Parse("2006-01-02", req.WorkoutDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout date format. Use YYYY-MM-DD")
		return
	}

	// Create workout domain object
	workout := &domain.Workout{
		WorkoutDate: workoutDate,
		WorkoutType: domain.WorkoutType(req.WorkoutType),
		WorkoutName: req.WorkoutName,
		Notes:       req.Notes,
		TotalTime:   req.TotalTime,
	}

	// Convert movements
	if len(req.Movements) > 0 {
		workout.Movements = make([]*domain.WorkoutMovement, len(req.Movements))
		now := time.Now()
		for i, m := range req.Movements {
			workout.Movements[i] = &domain.WorkoutMovement{
				MovementID: m.MovementID,
				Weight:     m.Weight,
				Sets:       m.Sets,
				Reps:       m.Reps,
				Time:       m.Time,
				Distance:   m.Distance,
				IsRx:       m.IsRx,
				Notes:      m.Notes,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
		}
	}

	// Update workout
	err = h.workoutService.UpdateWorkout(workoutID, userID, workout)
	if err != nil {
		switch err {
		case service.ErrWorkoutNotFound:
			respondError(w, http.StatusNotFound, "Workout not found")
		case service.ErrUnauthorized:
			respondError(w, http.StatusForbidden, "Access denied")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to update workout")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Workout updated successfully"})
}

// DeleteWorkout handles DELETE /api/workouts/{id}
func (h *WorkoutHandler) DeleteWorkout(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get workout ID from URL
	workoutIDStr := chi.URLParam(r, "id")
	workoutID, err := strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	// Delete workout
	err = h.workoutService.DeleteWorkout(workoutID, userID)
	if err != nil {
		switch err {
		case service.ErrWorkoutNotFound:
			respondError(w, http.StatusNotFound, "Workout not found")
		case service.ErrUnauthorized:
			respondError(w, http.StatusForbidden, "Access denied")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to delete workout")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Workout deleted successfully"})
}

// ListMovements handles GET /api/movements
func (h *WorkoutHandler) ListMovements(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get movements
	movements, err := h.workoutService.ListMovements(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list movements")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"movements": movements,
	})
}
