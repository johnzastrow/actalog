package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
)

// WorkoutHandler handles workout-related endpoints
type WorkoutHandler struct {
	workoutRepo         domain.WorkoutRepository
	workoutMovementRepo domain.WorkoutMovementRepository
}

// NewWorkoutHandler creates a new workout handler
func NewWorkoutHandler(workoutRepo domain.WorkoutRepository, workoutMovementRepo domain.WorkoutMovementRepository) *WorkoutHandler {
	return &WorkoutHandler{
		workoutRepo:         workoutRepo,
		workoutMovementRepo: workoutMovementRepo,
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
	WorkoutDate time.Time                   `json:"workout_date"`
	WorkoutType string                      `json:"workout_type"`
	WorkoutName string                      `json:"workout_name,omitempty"`
	Notes       string                      `json:"notes,omitempty"`
	TotalTime   *int                        `json:"total_time,omitempty"`
	Movements   []CreateWorkoutMovementData `json:"movements"`
}

// CreateWorkoutMovementData represents movement data in workout creation
type CreateWorkoutMovementData struct {
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
	ID          int64                      `json:"id"`
	UserID      int64                      `json:"user_id"`
	WorkoutDate time.Time                  `json:"workout_date"`
	WorkoutType string                     `json:"workout_type"`
	WorkoutName string                     `json:"workout_name,omitempty"`
	Notes       string                     `json:"notes,omitempty"`
	TotalTime   *int                       `json:"total_time,omitempty"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`
	Movements   []*domain.WorkoutMovement  `json:"movements,omitempty"`
}

// Create creates a new workout with movements
func (h *WorkoutHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	// Validate required fields
	if req.WorkoutType == "" {
		respondError(w, http.StatusBadRequest, "Workout type is required")
		return
	}

	if req.WorkoutDate.IsZero() {
		req.WorkoutDate = time.Now()
	}

	// TODO: Extract user ID from JWT token in context
	// For now, using a placeholder - this will be replaced with actual auth middleware
	userID := int64(1) // This should come from authenticated user context

	// Create workout
	workout := &domain.Workout{
		UserID:      userID,
		WorkoutDate: req.WorkoutDate,
		WorkoutType: domain.WorkoutType(req.WorkoutType),
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

	if err := h.workoutRepo.Create(workout); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create workout")
		return
	}

	// Create workout movements
	var movements []*domain.WorkoutMovement
	for i, movData := range req.Movements {
		wm := &domain.WorkoutMovement{
			WorkoutID:  workout.ID,
			MovementID: movData.MovementID,
			Weight:     movData.Weight,
			Sets:       movData.Sets,
			Reps:       movData.Reps,
			Time:       movData.Time,
			Distance:   movData.Distance,
			IsRx:       movData.IsRx,
			Notes:      movData.Notes,
			OrderIndex: i,
		}

		if err := h.workoutMovementRepo.Create(wm); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create workout movement")
			return
		}

		movements = append(movements, wm)
	}

	// Build response
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

// GetByID retrieves a workout by ID with its movements
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
	movements, err := h.workoutMovementRepo.GetByWorkoutID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve workout movements")
		return
	}

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

// ListByUser retrieves workouts for a user with pagination
func (h *WorkoutHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Extract user ID from JWT token in context
	userID := int64(1) // Placeholder

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
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

	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Check for date range filter
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var workouts []*domain.Workout
	var err error

	if startDateStr != "" && endDateStr != "" {
		startDate, err1 := time.Parse("2006-01-02", startDateStr)
		endDate, err2 := time.Parse("2006-01-02", endDateStr)

		if err1 != nil || err2 != nil {
			respondError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
			return
		}

		workouts, err = h.workoutRepo.GetByUserIDAndDateRange(userID, startDate, endDate)
	} else {
		workouts, err = h.workoutRepo.GetByUserID(userID, limit, offset)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve workouts")
		return
	}

	// Get total count
	count, err := h.workoutRepo.Count(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to count workouts")
		return
	}

	// Build response with movements for each workout
	var responses []WorkoutResponse
	for _, workout := range workouts {
		movements, err := h.workoutMovementRepo.GetByWorkoutID(workout.ID)
		if err != nil {
			// Log error but continue
			movements = []*domain.WorkoutMovement{}
		}

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

	result := map[string]interface{}{
		"workouts": responses,
		"total":    count,
		"limit":    limit,
		"offset":   offset,
	}

	respondJSON(w, http.StatusOK, result)
}

// Update updates a workout (not including movements)
func (h *WorkoutHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
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

	var req UpdateWorkoutRequest
	var req CreateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if workout exists
	workout, err := h.workoutRepo.GetByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve workout")
		return
	}

	if workout == nil {
		respondError(w, http.StatusNotFound, "Workout not found")
		return
	}

	// Update workout fields
	workout.WorkoutDate = req.WorkoutDate
	workout.WorkoutType = domain.WorkoutType(req.WorkoutType)
	workout.WorkoutName = req.WorkoutName
	workout.Notes = req.Notes
	workout.TotalTime = req.TotalTime

	if err := h.workoutRepo.Update(workout); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update workout")
		return
	}

	// Get updated workout with movements
	movements, err := h.workoutMovementRepo.GetByWorkoutID(id)
	if err != nil {
		movements = []*domain.WorkoutMovement{}
	}

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

// Delete deletes a workout and its movements
func (h *WorkoutHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	// Check if workout exists
	workout, err := h.workoutRepo.GetByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve workout")
		return
	}

	if workout == nil {
		respondError(w, http.StatusNotFound, "Workout not found")
		return
	}

	// Delete workout (movements will be cascade deleted by foreign key constraint)
	if err := h.workoutRepo.Delete(id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete workout")
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

// GetProgressByMovement retrieves progress data for a specific movement
func (h *WorkoutHandler) GetProgressByMovement(w http.ResponseWriter, r *http.Request) {
	movementIDStr := chi.URLParam(r, "movement_id")
	movementID, err := strconv.ParseInt(movementIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid movement ID")
		return
	}

	// TODO: Extract user ID from JWT token in context
	userID := int64(1) // Placeholder

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
		"movement_id": movementID,
		"history":     workoutMovements,
		"movements": movements,
	})
}
