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

// UserWorkoutHandler handles logging workout instances
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
	WorkoutID   int64   `json:"workout_id"`   // Template ID
	WorkoutDate string  `json:"workout_date"` // YYYY-MM-DD format
	WorkoutType *string `json:"workout_type,omitempty"`
	TotalTime   *int    `json:"total_time,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

// UpdateLoggedWorkoutRequest represents a request to update a logged workout
type UpdateLoggedWorkoutRequest struct {
	WorkoutType *string `json:"workout_type,omitempty"`
	TotalTime   *int    `json:"total_time,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

// UserWorkoutResponse represents a logged workout instance
type UserWorkoutResponse struct {
	ID           int64                           `json:"id"`
	UserID       int64                           `json:"user_id"`
	WorkoutID    int64                           `json:"workout_id"`
	WorkoutName  string                          `json:"workout_name"`
	WorkoutDate  string                          `json:"workout_date"`
	WorkoutType  *string                         `json:"workout_type,omitempty"`
	TotalTime    *int                            `json:"total_time,omitempty"`
	Notes        *string                         `json:"notes,omitempty"`
	CreatedAt    string                          `json:"created_at"`
	UpdatedAt    string                          `json:"updated_at"`
	Movements    []*domain.WorkoutMovement       `json:"movements,omitempty"`
	WODs         []*domain.WorkoutWODWithDetails `json:"wods,omitempty"`
	WorkoutNotes *string                         `json:"workout_notes,omitempty"`
}

// LogWorkout logs a workout instance (user performs a workout template)
func (h *UserWorkoutHandler) LogWorkout(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req LogWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.WorkoutID == 0 {
		respondError(w, http.StatusBadRequest, "Workout ID is required")
		return
	}

	if req.WorkoutDate == "" {
		respondError(w, http.StatusBadRequest, "Workout date is required")
		return
	}

	// Parse workout date
	workoutDate, err := time.Parse("2006-01-02", req.WorkoutDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout date format. Use YYYY-MM-DD")
		return
	}

	// Log workout
	userWorkout, err := h.userWorkoutService.LogWorkout(userID, req.WorkoutID, workoutDate, req.Notes, req.TotalTime, req.WorkoutType)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to log workout: "+err.Error())
		return
	}

	// Retrieve logged workout with details
	logged, err := h.userWorkoutService.GetLoggedWorkout(userWorkout.ID, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve logged workout")
		return
	}

	response := UserWorkoutResponse{
		ID:           logged.ID,
		UserID:       logged.UserID,
		WorkoutID:    logged.WorkoutID,
		WorkoutName:  logged.WorkoutName,
		WorkoutDate:  logged.WorkoutDate.Format("2006-01-02"),
		WorkoutType:  logged.WorkoutType,
		TotalTime:    logged.TotalTime,
		Notes:        logged.Notes,
		CreatedAt:    logged.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    logged.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Movements:    logged.Movements,
		WODs:         logged.WODs,
		WorkoutNotes: logged.WorkoutDescription,
	}

	respondJSON(w, http.StatusCreated, response)
}

// GetLoggedWorkout retrieves a logged workout by ID
func (h *UserWorkoutHandler) GetLoggedWorkout(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	logged, err := h.userWorkoutService.GetLoggedWorkout(id, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve logged workout: "+err.Error())
		return
	}

	response := UserWorkoutResponse{
		ID:           logged.ID,
		UserID:       logged.UserID,
		WorkoutID:    logged.WorkoutID,
		WorkoutName:  logged.WorkoutName,
		WorkoutDate:  logged.WorkoutDate.Format("2006-01-02"),
		WorkoutType:  logged.WorkoutType,
		TotalTime:    logged.TotalTime,
		Notes:        logged.Notes,
		CreatedAt:    logged.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    logged.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Movements:    logged.Movements,
		WODs:         logged.WODs,
		WorkoutNotes: logged.WorkoutDescription,
	}

	respondJSON(w, http.StatusOK, response)
}

// ListLoggedWorkouts retrieves all workouts logged by the user
func (h *UserWorkoutHandler) ListLoggedWorkouts(w http.ResponseWriter, r *http.Request) {
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

	// Check for date range filtering
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var workouts []*domain.UserWorkoutWithDetails
	var err error

	if startDateStr != "" && endDateStr != "" {
		startDate, err1 := time.Parse("2006-01-02", startDateStr)
		endDate, err2 := time.Parse("2006-01-02", endDateStr)

		if err1 != nil || err2 != nil {
			respondError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
			return
		}

		// Get basic workouts in range
		basicWorkouts, err := h.userWorkoutService.ListLoggedWorkoutsByDateRange(userID, startDate, endDate)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to retrieve logged workouts")
			return
		}

		// Get details for each
		for _, uw := range basicWorkouts {
			detailed, err := h.userWorkoutService.GetLoggedWorkout(uw.ID, userID)
			if err != nil {
				continue // Skip if error getting details
			}
			workouts = append(workouts, detailed)
		}
	} else {
		workouts, err = h.userWorkoutService.ListLoggedWorkouts(userID, limit, offset)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to retrieve logged workouts")
			return
		}
	}

	// Build response
	var responses []UserWorkoutResponse
	for _, logged := range workouts {
		response := UserWorkoutResponse{
			ID:           logged.ID,
			UserID:       logged.UserID,
			WorkoutID:    logged.WorkoutID,
			WorkoutName:  logged.WorkoutName,
			WorkoutDate:  logged.WorkoutDate.Format("2006-01-02"),
			WorkoutType:  logged.WorkoutType,
			TotalTime:    logged.TotalTime,
			Notes:        logged.Notes,
			CreatedAt:    logged.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:    logged.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Movements:    logged.Movements,
			WODs:         logged.WODs,
			WorkoutNotes: logged.WorkoutDescription,
		}
		responses = append(responses, response)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"workouts": responses,
		"limit":    limit,
		"offset":   offset,
	})
}

// UpdateLoggedWorkout updates a logged workout
func (h *UserWorkoutHandler) UpdateLoggedWorkout(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	var req UpdateLoggedWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update logged workout with individual fields
	if err := h.userWorkoutService.UpdateLoggedWorkout(id, userID, req.Notes, req.TotalTime, req.WorkoutType); err != nil {
		switch err {
		case service.ErrUserWorkoutNotFound:
			respondError(w, http.StatusNotFound, "Logged workout not found")
		case service.ErrUnauthorizedWorkoutAccess:
			respondError(w, http.StatusForbidden, "You don't have permission to update this workout")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to update logged workout")
		}
		return
	}

	// Retrieve updated logged workout
	logged, err := h.userWorkoutService.GetLoggedWorkout(id, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve updated workout")
		return
	}

	response := UserWorkoutResponse{
		ID:           logged.ID,
		UserID:       logged.UserID,
		WorkoutID:    logged.WorkoutID,
		WorkoutName:  logged.WorkoutName,
		WorkoutDate:  logged.WorkoutDate.Format("2006-01-02"),
		WorkoutType:  logged.WorkoutType,
		TotalTime:    logged.TotalTime,
		Notes:        logged.Notes,
		CreatedAt:    logged.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    logged.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Movements:    logged.Movements,
		WODs:         logged.WODs,
		WorkoutNotes: logged.WorkoutDescription,
	}

	respondJSON(w, http.StatusOK, response)
}

// DeleteLoggedWorkout deletes a logged workout
func (h *UserWorkoutHandler) DeleteLoggedWorkout(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workout ID")
		return
	}

	if err := h.userWorkoutService.DeleteLoggedWorkout(id, userID); err != nil {
		if err == service.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "You don't have permission to delete this workout")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to delete logged workout: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Logged workout deleted successfully"})
}

// GetMonthlyStats retrieves workout count for a specific month
func (h *UserWorkoutHandler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse year and month
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	if yearStr == "" || monthStr == "" {
		respondError(w, http.StatusBadRequest, "Year and month are required")
		return
	}

	year, err1 := strconv.Atoi(yearStr)
	month, err2 := strconv.Atoi(monthStr)

	if err1 != nil || err2 != nil || month < 1 || month > 12 {
		respondError(w, http.StatusBadRequest, "Invalid year or month")
		return
	}

	count, err := h.userWorkoutService.GetWorkoutStatsForMonth(userID, year, month)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve monthly stats")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"year":          year,
		"month":         month,
		"workout_count": count,
	})
}
