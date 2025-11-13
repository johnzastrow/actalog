package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// UserWorkoutHandler handles logging workout instances
type UserWorkoutHandler struct {
	userWorkoutService *service.UserWorkoutService
	logger             *logger.Logger
}

// NewUserWorkoutHandler creates a new user workout handler
func NewUserWorkoutHandler(userWorkoutService *service.UserWorkoutService, l *logger.Logger) *UserWorkoutHandler {
	return &UserWorkoutHandler{
		userWorkoutService: userWorkoutService,
		logger:             l,
	}
}

// LogWorkoutRequest represents a request to log a workout instance
type LogWorkoutRequest struct {
	WorkoutID   *int64  `json:"workout_id,omitempty"`   // Template ID (optional for ad-hoc workouts)
	WorkoutName *string `json:"workout_name,omitempty"` // Name for ad-hoc workouts (required if workout_id is null)
	WorkoutDate string  `json:"workout_date"`           // YYYY-MM-DD format
	WorkoutType *string `json:"workout_type,omitempty"`
	TotalTime   *int    `json:"total_time,omitempty"`
	Notes       *string `json:"notes,omitempty"`
	// Performance data
	Movements []MovementPerformance `json:"movements,omitempty"`
	WODs      []WODPerformance      `json:"wods,omitempty"`
}

// MovementPerformance represents performance data for a single movement
type MovementPerformance struct {
	MovementID int64    `json:"movement_id"`
	Sets       *int     `json:"sets,omitempty"`
	Reps       *int     `json:"reps,omitempty"`
	Weight     *float64 `json:"weight,omitempty"`
	Time       *int     `json:"time,omitempty"`       // in seconds
	Distance   *float64 `json:"distance,omitempty"`
	Notes      string   `json:"notes,omitempty"`
	OrderIndex int      `json:"order_index"`
}

// WODPerformance represents performance data for a single WOD
type WODPerformance struct {
	WODID       int64    `json:"wod_id"`
	ScoreType   *string  `json:"score_type,omitempty"`   // Time, Rounds+Reps, Max Weight
	ScoreValue  *string  `json:"score_value,omitempty"`  // Formatted score
	TimeSeconds *int     `json:"time_seconds,omitempty"` // For time-based WODs
	Rounds      *int     `json:"rounds,omitempty"`       // For AMRAP
	Reps        *int     `json:"reps,omitempty"`         // Remaining reps in AMRAP
	Weight      *float64 `json:"weight,omitempty"`       // For max weight WODs
	Notes       string   `json:"notes,omitempty"`
	OrderIndex  int      `json:"order_index"`
}

// UpdateLoggedWorkoutRequest represents a request to update a logged workout
type UpdateLoggedWorkoutRequest struct {
	WorkoutType *string                `json:"workout_type,omitempty"`
	TotalTime   *int                   `json:"total_time,omitempty"`
	Notes       *string                `json:"notes,omitempty"`
	Movements   []MovementPerformance  `json:"movements,omitempty"`
	WODs        []WODPerformance       `json:"wods,omitempty"`
}

// UserWorkoutResponse represents a logged workout instance
type UserWorkoutResponse struct {
	ID                   int64                           `json:"id"`
	UserID               int64                           `json:"user_id"`
	WorkoutID            int64                           `json:"workout_id"`
	WorkoutName          string                          `json:"workout_name"`
	WorkoutDate          string                          `json:"workout_date"`
	WorkoutType          *string                         `json:"workout_type,omitempty"`
	TotalTime            *int                            `json:"total_time,omitempty"`
	Notes                *string                         `json:"notes,omitempty"`
	CreatedAt            string                          `json:"created_at"`
	UpdatedAt            string                          `json:"updated_at"`
	Movements            []*domain.WorkoutMovement       `json:"movements,omitempty"`       // Template movements
	WODs                 []*domain.WorkoutWODWithDetails `json:"wods,omitempty"`            // Template WODs
	PerformanceMovements []*domain.UserWorkoutMovement   `json:"performance_movements,omitempty"` // Actual performance
	PerformanceWODs      []*domain.UserWorkoutWOD        `json:"performance_wods,omitempty"`      // Actual performance
	WorkoutNotes         *string                         `json:"workout_notes,omitempty"`
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
	// Either workout_id (template-based) OR workout_name (ad-hoc) must be provided
	if (req.WorkoutID == nil || *req.WorkoutID == 0) && (req.WorkoutName == nil || *req.WorkoutName == "") {
		if h.logger != nil {
			h.logger.Warn("action=log_workout outcome=failure user_id=%d reason=missing_workout_id_and_name", userID)
		}
		respondError(w, http.StatusBadRequest, "Either workout_id or workout_name is required")
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

	if h.logger != nil {
		h.logger.Info("action=log_workout_attempt user_id=%d workout_id=%d date=%s", userID, req.WorkoutID, req.WorkoutDate)
	}

	// Check if performance data was provided
	var userWorkout *domain.UserWorkout
	if len(req.Movements) > 0 || len(req.WODs) > 0 {
		// Convert request movements to domain movements
		movements := make([]*domain.UserWorkoutMovement, len(req.Movements))
		for i, m := range req.Movements {
			movements[i] = &domain.UserWorkoutMovement{
				MovementID: m.MovementID,
				Sets:       m.Sets,
				Reps:       m.Reps,
				Weight:     m.Weight,
				Time:       m.Time,
				Distance:   m.Distance,
				Notes:      m.Notes,
				OrderIndex: m.OrderIndex,
			}
		}

		// Convert request WODs to domain WODs
		wods := make([]*domain.UserWorkoutWOD, len(req.WODs))
		for i, w := range req.WODs {
			wods[i] = &domain.UserWorkoutWOD{
				WODID:       w.WODID,
				ScoreType:   w.ScoreType,
				ScoreValue:  w.ScoreValue,
				TimeSeconds: w.TimeSeconds,
				Rounds:      w.Rounds,
				Reps:        w.Reps,
				Weight:      w.Weight,
				Notes:       w.Notes,
				OrderIndex:  w.OrderIndex,
			}
		}

		// Log workout with performance data
		userWorkout, err = h.userWorkoutService.LogWorkoutWithPerformance(
			userID, req.WorkoutID, workoutDate,
			req.Notes, req.TotalTime, req.WorkoutType,
			movements, wods,
		)
	} else {
		// Log workout without performance data
		userWorkout, err = h.userWorkoutService.LogWorkout(userID, req.WorkoutID, workoutDate, req.Notes, req.TotalTime, req.WorkoutType)
	}

	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=log_workout outcome=failure user_id=%d workout_id=%d error=%v", userID, req.WorkoutID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to log workout: "+err.Error())
		return
	}

	// Retrieve logged workout with details
	logged, err := h.userWorkoutService.GetLoggedWorkout(userWorkout.ID, userID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=log_workout outcome=failure user_id=%d workout_id=%d error=retrieval_failed %v", userID, req.WorkoutID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve logged workout")
		return
	}

	response := UserWorkoutResponse{
		ID:                   logged.ID,
		UserID:               logged.UserID,
		WorkoutID:            logged.WorkoutID,
		WorkoutName:          logged.WorkoutName,
		WorkoutDate:          logged.WorkoutDate.Format("2006-01-02"),
		WorkoutType:          logged.WorkoutType,
		TotalTime:            logged.TotalTime,
		Notes:                logged.Notes,
		CreatedAt:            logged.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:            logged.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Movements:            logged.Movements,
		WODs:                 logged.WODs,
		PerformanceMovements: logged.PerformanceMovements,
		PerformanceWODs:      logged.PerformanceWODs,
		WorkoutNotes:         logged.WorkoutDescription,
	}

	if h.logger != nil {
		h.logger.Info("action=log_workout outcome=success user_id=%d logged_id=%d", userID, userWorkout.ID)
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

	if h.logger != nil {
		h.logger.Info("action=get_workout user_id=%d workout_id=%d", userID, id)
	}

	logged, err := h.userWorkoutService.GetLoggedWorkout(id, userID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=get_workout outcome=failure user_id=%d workout_id=%d error=%v", userID, id, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve logged workout: "+err.Error())
		return
	}

	response := UserWorkoutResponse{
		ID:                   logged.ID,
		UserID:               logged.UserID,
		WorkoutID:            logged.WorkoutID,
		WorkoutName:          logged.WorkoutName,
		WorkoutDate:          logged.WorkoutDate.Format("2006-01-02"),
		WorkoutType:          logged.WorkoutType,
		TotalTime:            logged.TotalTime,
		Notes:                logged.Notes,
		CreatedAt:            logged.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:            logged.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Movements:            logged.Movements,
		WODs:                 logged.WODs,
		PerformanceMovements: logged.PerformanceMovements,
		PerformanceWODs:      logged.PerformanceWODs,
		WorkoutNotes:         logged.WorkoutDescription,
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
			if h.logger != nil {
				h.logger.Warn("action=list_workouts outcome=failure user_id=%d reason=invalid_date_range start=%s end=%s", userID, startDateStr, endDateStr)
			}
			respondError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
			return
		}

		// Get basic workouts in range
		basicWorkouts, err := h.userWorkoutService.ListLoggedWorkoutsByDateRange(userID, startDate, endDate)
		if err != nil {
			if h.logger != nil {
				h.logger.Error("action=list_workouts outcome=failure user_id=%d error=%v", userID, err)
			}
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
		if h.logger != nil {
			h.logger.Info("action=list_workouts_attempt user_id=%d limit=%d offset=%d", userID, limit, offset)
		}

		workouts, err = h.userWorkoutService.ListLoggedWorkouts(userID, limit, offset)
		if err != nil {
			if h.logger != nil {
				h.logger.Error("action=list_workouts outcome=failure user_id=%d error=%v", userID, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to retrieve logged workouts")
			return
		}
		if h.logger != nil {
			h.logger.Info("action=list_workouts outcome=success user_id=%d returned=%d", userID, len(workouts))
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
	if h.logger != nil {
		h.logger.Info("action=update_workout_attempt user_id=%d workout_id=%d", userID, id)
	}

	if err := h.userWorkoutService.UpdateLoggedWorkout(id, userID, req.Notes, req.TotalTime, req.WorkoutType); err != nil {
		switch err {
		case service.ErrUserWorkoutNotFound:
			if h.logger != nil {
				h.logger.Warn("action=update_workout outcome=failure user_id=%d workout_id=%d reason=not_found", userID, id)
			}
			respondError(w, http.StatusNotFound, "Logged workout not found")
		case service.ErrUnauthorizedWorkoutAccess:
			if h.logger != nil {
				h.logger.Warn("action=update_workout outcome=failure user_id=%d workout_id=%d reason=unauthorized", userID, id)
			}
			respondError(w, http.StatusForbidden, "You don't have permission to update this workout")
		default:
			if h.logger != nil {
				h.logger.Error("action=update_workout outcome=failure user_id=%d workout_id=%d error=%v", userID, id, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to update logged workout")
		}
		return
	}

	// Update movements if provided
	if len(req.Movements) > 0 {
		// Convert request movements to domain movements
		movements := make([]domain.UserWorkoutMovement, len(req.Movements))
		for i, m := range req.Movements {
			movements[i] = domain.UserWorkoutMovement{
				MovementID: m.MovementID,
				Sets:       m.Sets,
				Reps:       m.Reps,
				Weight:     m.Weight,
				Time:       m.Time,
				Distance:   m.Distance,
				Notes:      m.Notes,
				OrderIndex: m.OrderIndex,
			}
		}

		if err := h.userWorkoutService.UpdateWorkoutMovements(id, userID, movements); err != nil {
			if h.logger != nil {
				h.logger.Error("action=update_workout_movements outcome=failure user_id=%d workout_id=%d error=%v", userID, id, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to update workout movements")
			return
		}
	}

	// Update WODs if provided
	if len(req.WODs) > 0 {
		// Convert request WODs to domain WODs
		wods := make([]domain.UserWorkoutWOD, len(req.WODs))
		for i, w := range req.WODs {
			wods[i] = domain.UserWorkoutWOD{
				WODID:       w.WODID,
				ScoreType:   w.ScoreType,
				ScoreValue:  w.ScoreValue,
				TimeSeconds: w.TimeSeconds,
				Rounds:      w.Rounds,
				Reps:        w.Reps,
				Weight:      w.Weight,
				Notes:       w.Notes,
				OrderIndex:  w.OrderIndex,
			}
		}

		if err := h.userWorkoutService.UpdateWorkoutWODs(id, userID, wods); err != nil {
			if h.logger != nil {
				h.logger.Error("action=update_workout_wods outcome=failure user_id=%d workout_id=%d error=%v", userID, id, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to update workout WODs")
			return
		}
	}

	// Retrieve updated logged workout
	logged, err := h.userWorkoutService.GetLoggedWorkout(id, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve updated workout")
		return
	}

	response := UserWorkoutResponse{
		ID:                   logged.ID,
		UserID:               logged.UserID,
		WorkoutID:            logged.WorkoutID,
		WorkoutName:          logged.WorkoutName,
		WorkoutDate:          logged.WorkoutDate.Format("2006-01-02"),
		WorkoutType:          logged.WorkoutType,
		TotalTime:            logged.TotalTime,
		Notes:                logged.Notes,
		CreatedAt:            logged.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:            logged.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Movements:            logged.Movements,
		WODs:                 logged.WODs,
		PerformanceMovements: logged.PerformanceMovements,
		PerformanceWODs:      logged.PerformanceWODs,
		WorkoutNotes:         logged.WorkoutDescription,
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

	if h.logger != nil {
		h.logger.Info("action=delete_workout_attempt user_id=%d workout_id=%d", userID, id)
	}

	if err := h.userWorkoutService.DeleteLoggedWorkout(id, userID); err != nil {
		if err == service.ErrUnauthorized {
			if h.logger != nil {
				h.logger.Warn("action=delete_workout outcome=failure user_id=%d workout_id=%d reason=unauthorized", userID, id)
			}
			respondError(w, http.StatusForbidden, "You don't have permission to delete this workout")
		} else {
			if h.logger != nil {
				h.logger.Error("action=delete_workout outcome=failure user_id=%d workout_id=%d error=%v", userID, id, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to delete logged workout: "+err.Error())
		}
		return
	}

	if h.logger != nil {
		h.logger.Info("action=delete_workout outcome=success user_id=%d workout_id=%d", userID, id)
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
// ErrorResponse represents an error response
