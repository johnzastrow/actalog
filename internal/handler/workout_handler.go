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

// WorkoutHandler handles workout-related endpoints
type WorkoutHandler struct {
	workoutRepo         domain.WorkoutRepository
	workoutMovementRepo domain.WorkoutMovementRepository
	workoutService      *service.WorkoutService
}

// NewWorkoutHandler creates a new workout handler
func NewWorkoutHandler(workoutRepo domain.WorkoutRepository, workoutMovementRepo domain.WorkoutMovementRepository, workoutService *service.WorkoutService) *WorkoutHandler {
	return &WorkoutHandler{
		workoutRepo:         workoutRepo,
		workoutMovementRepo: workoutMovementRepo,
		workoutService:      workoutService,
	}
}

// CreateWorkoutRequest represents a request to create a workout
type CreateWorkoutRequest struct {
	WorkoutDate time.Time                      `json:"workout_date"`
	WorkoutType string                         `json:"workout_type"`
	WorkoutName string                         `json:"workout_name,omitempty"`
	Notes       string                         `json:"notes,omitempty"`
	TotalTime   *int                           `json:"total_time,omitempty"`
	Movements   []CreateWorkoutMovementRequest `json:"movements"`
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

// UpdateWorkoutRequest represents a request to update a workout
type UpdateWorkoutRequest struct {
	WorkoutDate time.Time `json:"workout_date"`
	WorkoutType string    `json:"workout_type"`
	WorkoutName string    `json:"workout_name,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	TotalTime   *int      `json:"total_time,omitempty"`
}

// WorkoutResponse represents a workout with its movements
type WorkoutResponse struct {
	ID          int64                     `json:"id"`
	UserID      int64                     `json:"user_id"`
	WorkoutDate time.Time                 `json:"workout_date"`
	WorkoutType string                    `json:"workout_type"`
	WorkoutName string                    `json:"workout_name,omitempty"`
	Notes       string                    `json:"notes,omitempty"`
	TotalTime   *int                      `json:"total_time,omitempty"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
	Movements   []*domain.WorkoutMovement `json:"movements,omitempty"`
}

// Create creates a new workout with movements
func (h *WorkoutHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
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

	// Validate required fields
	if req.WorkoutType == "" {
		respondError(w, http.StatusBadRequest, "Workout type is required")
		return
	}

	if req.WorkoutDate.IsZero() {
		req.WorkoutDate = time.Now()
	}

	// Create workout
	workout := &domain.Workout{
		UserID:      userID,
		WorkoutDate: req.WorkoutDate,
		WorkoutType: domain.WorkoutType(req.WorkoutType),
		WorkoutName: req.WorkoutName,
		Notes:       req.Notes,
		TotalTime:   req.TotalTime,
	}

	if err := h.workoutRepo.Create(workout); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create workout")
		return
	}

	// Create workout movements
	for i, mv := range req.Movements {
		workoutMovement := &domain.WorkoutMovement{
			WorkoutID:  workout.ID,
			MovementID: mv.MovementID,
			Weight:     mv.Weight,
			Sets:       mv.Sets,
			Reps:       mv.Reps,
			Time:       mv.Time,
			Distance:   mv.Distance,
			IsRx:       mv.IsRx,
			Notes:      mv.Notes,
			OrderIndex: i,
		}

		if err := h.workoutMovementRepo.Create(workoutMovement); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create workout movement")
			return
		}
	}

	// Retrieve created workout with movements
	movements, _ := h.workoutMovementRepo.GetByWorkoutID(workout.ID)

	response := WorkoutResponse{
		ID:          workout.ID,
		UserID:      workout.UserID,
		WorkoutDate: workout.WorkoutDate,
		WorkoutType: string(workout.WorkoutType),
		WorkoutName: workout.WorkoutName,
		Notes:       workout.Notes,
		TotalTime:   workout.TotalTime,
		CreatedAt:   workout.CreatedAt,
		UpdatedAt:   workout.UpdatedAt,
		Movements:   movements,
	}

	respondJSON(w, http.StatusCreated, response)
}

// GetByID retrieves a workout by ID
func (h *WorkoutHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	workout, err := h.workoutRepo.GetByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve workout")
		return
	}

	if workout == nil {
		respondError(w, http.StatusNotFound, "Workout not found")
		return
	}

	// Get workout movements
	movements, _ := h.workoutMovementRepo.GetByWorkoutID(workout.ID)

	response := WorkoutResponse{
		ID:          workout.ID,
		UserID:      workout.UserID,
		WorkoutDate: workout.WorkoutDate,
		WorkoutType: string(workout.WorkoutType),
		WorkoutName: workout.WorkoutName,
		Notes:       workout.Notes,
		TotalTime:   workout.TotalTime,
		CreatedAt:   workout.CreatedAt,
		UpdatedAt:   workout.UpdatedAt,
		Movements:   movements,
	}

	respondJSON(w, http.StatusOK, response)
}

// ListByUser retrieves workouts for a specific user
func (h *WorkoutHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // default
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var workouts []*domain.Workout
	var err error

	// Check for date range filtering
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr != "" && endDateStr != "" {
		startDate, err1 := time.Parse("2006-01-02", startDateStr)
		endDate, err2 := time.Parse("2006-01-02", endDateStr)

		if err1 == nil && err2 == nil {
			workouts, err = h.workoutRepo.GetByUserIDAndDateRange(userID, startDate, endDate)
		} else {
			respondError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
			return
		}
	} else {
		workouts, err = h.workoutRepo.GetByUserID(userID, limit, offset)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve workouts")
		return
	}

	// Build response with movements
	var responses []WorkoutResponse
	for _, workout := range workouts {
		movements, _ := h.workoutMovementRepo.GetByWorkoutID(workout.ID)

		response := WorkoutResponse{
			ID:          workout.ID,
			UserID:      workout.UserID,
			WorkoutDate: workout.WorkoutDate,
			WorkoutType: string(workout.WorkoutType),
			WorkoutName: workout.WorkoutName,
			Notes:       workout.Notes,
			TotalTime:   workout.TotalTime,
			CreatedAt:   workout.CreatedAt,
			UpdatedAt:   workout.UpdatedAt,
			Movements:   movements,
		}
		responses = append(responses, response)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"workouts": responses,
		"limit":    limit,
		"offset":   offset,
	})
}

// Update updates a workout
func (h *WorkoutHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	// Retrieve existing workout
	workout, err := h.workoutRepo.GetByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve workout")
		return
	}

	if workout == nil {
		respondError(w, http.StatusNotFound, "Workout not found")
		return
	}

	// Parse update request
	var req UpdateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update workout fields
	if !req.WorkoutDate.IsZero() {
		workout.WorkoutDate = req.WorkoutDate
	}
	if req.WorkoutType != "" {
		workout.WorkoutType = domain.WorkoutType(req.WorkoutType)
	}
	workout.WorkoutName = req.WorkoutName
	workout.Notes = req.Notes
	workout.TotalTime = req.TotalTime

	if err := h.workoutRepo.Update(workout); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update workout")
		return
	}

	// Get workout movements
	movements, _ := h.workoutMovementRepo.GetByWorkoutID(workout.ID)

	response := WorkoutResponse{
		ID:          workout.ID,
		UserID:      workout.UserID,
		WorkoutDate: workout.WorkoutDate,
		WorkoutType: string(workout.WorkoutType),
		WorkoutName: workout.WorkoutName,
		Notes:       workout.Notes,
		TotalTime:   workout.TotalTime,
		CreatedAt:   workout.CreatedAt,
		UpdatedAt:   workout.UpdatedAt,
		Movements:   movements,
	}

	respondJSON(w, http.StatusOK, response)
}

// Delete deletes a workout
func (h *WorkoutHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	// Delete workout movements first (cascade should handle this, but being explicit)
	if err := h.workoutMovementRepo.DeleteByWorkoutID(id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete workout movements")
		return
	}

	// Delete workout
	if err := h.workoutRepo.Delete(id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete workout")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Workout deleted successfully"})
}

// GetProgressByMovement retrieves progress data for a specific movement
func (h *WorkoutHandler) GetProgressByMovement(w http.ResponseWriter, r *http.Request) {
	movementIDStr := chi.URLParam(r, "movement_id")
	movementID, err := strconv.ParseInt(movementIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid movement ID")
		return
	}

	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse limit
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get workout movements for this user and movement
	workoutMovements, err := h.workoutMovementRepo.GetByUserIDAndMovementID(userID, movementID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve progress data")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"movement_id": movementID,
		"history":     workoutMovements,
	})
}

// GetPersonalRecords retrieves all personal records for a user
func (h *WorkoutHandler) GetPersonalRecords(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	records, err := h.workoutService.GetPersonalRecords(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve personal records")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"records": records,
	})
}

// GetPRMovements retrieves recent PR-flagged movements for a user
func (h *WorkoutHandler) GetPRMovements(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse limit
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	movements, err := h.workoutService.GetPRMovements(userID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve PR movements")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pr_movements": movements,
	})
}

// TogglePRFlag manually toggles the PR flag on a workout movement
func (h *WorkoutHandler) TogglePRFlag(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
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

	err = h.workoutService.TogglePRFlag(movementID, userID)
	if err != nil {
		if err == service.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "Unauthorized")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to toggle PR flag")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "PR flag toggled successfully"})
}
