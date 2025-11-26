package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
)

// AdminHandler handles admin-only operations
type AdminHandler struct {
	db                      *sql.DB
	userWorkoutWODRepo      domain.UserWorkoutWODRepository
	wodRepo                 domain.WODRepository
	movementRepo            domain.MovementRepository
	workoutRepo             domain.WorkoutRepository
	userRepo                domain.UserRepository
	wodService              *service.WODService
	movementService         *service.MovementService
	workoutTemplateService  *service.WorkoutTemplateService
	logger                  *logger.Logger
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	db *sql.DB,
	userWorkoutWODRepo domain.UserWorkoutWODRepository,
	wodRepo domain.WODRepository,
	movementRepo domain.MovementRepository,
	workoutRepo domain.WorkoutRepository,
	userRepo domain.UserRepository,
	wodService *service.WODService,
	movementService *service.MovementService,
	workoutTemplateService *service.WorkoutTemplateService,
	logger *logger.Logger,
) *AdminHandler {
	return &AdminHandler{
		db:                      db,
		userWorkoutWODRepo:      userWorkoutWODRepo,
		wodRepo:                 wodRepo,
		movementRepo:            movementRepo,
		workoutRepo:             workoutRepo,
		userRepo:                userRepo,
		wodService:              wodService,
		movementService:         movementService,
		workoutTemplateService:  workoutTemplateService,
		logger:                  logger,
	}
}

// WODMismatch represents a WOD score_type mismatch
type WODMismatch struct {
	ID               int64   `json:"id"`
	WODID            int64   `json:"wod_id"`
	WODName          string  `json:"wod_name"`
	UserEmail        string  `json:"user_email"`
	WorkoutDate      string  `json:"workout_date"`
	ExpectedScoreType string `json:"expected_score_type"`
	Issue            string  `json:"issue"`
	TimeSeconds      *int    `json:"time_seconds,omitempty"`
	Rounds           *int    `json:"rounds,omitempty"`
	Reps             *int    `json:"reps,omitempty"`
	Weight           *float64 `json:"weight,omitempty"`
}

// DetectWODScoreTypeMismatches detects WOD records that don't match their score_type
func (h *AdminHandler) DetectWODScoreTypeMismatches(w http.ResponseWriter, r *http.Request) {
	// Get all WOD performance records with WOD definitions
	// Note: This query needs to be run across all users
	query := `
		SELECT uww.id, uww.wod_id, uww.time_seconds, uww.rounds, uww.reps, uww.weight,
		       w.name, w.score_type,
		       u.email,
		       uw.workout_date
		FROM user_workout_wods uww
		JOIN wods w ON uww.wod_id = w.id
		JOIN user_workouts uw ON uww.user_workout_id = uw.id
		JOIN users u ON uw.user_id = u.id
		ORDER BY uw.workout_date DESC`

	rows, err := h.db.Query(query)
	if err != nil {
		h.logger.Error("Failed to query WOD records", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Failed to query WOD records"})
		return
	}
	defer rows.Close()

	var mismatches []WODMismatch

	for rows.Next() {
		var (
			id          int64
			wodID       int64
			timeSeconds *int
			rounds      *int
			reps        *int
			weight      *float64
			wodName     string
			scoreType   string
			userEmail   string
			workoutDate string
		)

		err := rows.Scan(&id, &wodID, &timeSeconds, &rounds, &reps, &weight, &wodName, &scoreType, &userEmail, &workoutDate)
		if err != nil {
			h.logger.Error("Failed to scan WOD record", "error", err)
			continue
		}

		// Check for mismatches based on score_type
		var issue string

		if scoreType == "Time (HH:MM:SS)" {
			// Must have time_seconds, must NOT have rounds/reps/weight
			if timeSeconds == nil {
				issue = "Missing time_seconds for Time-based WOD"
			} else if rounds != nil || reps != nil || weight != nil {
				issue = "Has invalid fields (rounds/reps/weight) for Time-based WOD"
			}
		} else if scoreType == "Rounds+Reps" {
			// Must have rounds, must NOT have time_seconds/weight
			if rounds == nil {
				issue = "Missing rounds for Rounds+Reps WOD"
			} else if timeSeconds != nil || weight != nil {
				issue = "Has invalid fields (time_seconds/weight) for Rounds+Reps WOD"
			}
		} else if scoreType == "Max Weight" {
			// Must have weight, must NOT have time_seconds/rounds/reps
			if weight == nil {
				issue = "Missing weight for Max Weight WOD"
			} else if timeSeconds != nil || rounds != nil || reps != nil {
				issue = "Has invalid fields (time_seconds/rounds/reps) for Max Weight WOD"
			}
		}

		// If there's an issue, add to mismatches
		if issue != "" {
			mismatches = append(mismatches, WODMismatch{
				ID:                id,
				WODID:             wodID,
				WODName:           wodName,
				UserEmail:         userEmail,
				WorkoutDate:       workoutDate,
				ExpectedScoreType: scoreType,
				Issue:             issue,
				TimeSeconds:       timeSeconds,
				Rounds:            rounds,
				Reps:              reps,
				Weight:            weight,
			})
		}
	}

	if err := rows.Err(); err != nil {
		h.logger.Error("Error iterating WOD records", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Error processing WOD records"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mismatches": mismatches,
		"count":      len(mismatches),
	})
}

// FixWODScoreTypeMismatches deletes WOD records that don't match their score_type
func (h *AdminHandler) FixWODScoreTypeMismatches(w http.ResponseWriter, r *http.Request) {
	// First, get all mismatches
	query := `
		SELECT uww.id, uww.wod_id, uww.time_seconds, uww.rounds, uww.reps, uww.weight,
		       w.score_type
		FROM user_workout_wods uww
		JOIN wods w ON uww.wod_id = w.id`

	rows, err := h.db.Query(query)
	if err != nil {
		h.logger.Error("Failed to query WOD records", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Failed to query WOD records"})
		return
	}
	defer rows.Close()

	var idsToDelete []int64

	for rows.Next() {
		var (
			id          int64
			wodID       int64
			timeSeconds *int
			rounds      *int
			reps        *int
			weight      *float64
			scoreType   string
		)

		err := rows.Scan(&id, &wodID, &timeSeconds, &rounds, &reps, &weight, &scoreType)
		if err != nil {
			h.logger.Error("Failed to scan WOD record", "error", err)
			continue
		}

		// Check for mismatches based on score_type
		isMismatch := false

		if scoreType == "Time (HH:MM:SS)" {
			if timeSeconds == nil || rounds != nil || reps != nil || weight != nil {
				isMismatch = true
			}
		} else if scoreType == "Rounds+Reps" {
			if rounds == nil || timeSeconds != nil || weight != nil {
				isMismatch = true
			}
		} else if scoreType == "Max Weight" {
			if weight == nil || timeSeconds != nil || rounds != nil || reps != nil {
				isMismatch = true
			}
		}

		if isMismatch {
			idsToDelete = append(idsToDelete, id)
		}
	}

	if err := rows.Err(); err != nil {
		h.logger.Error("Error iterating WOD records", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Error processing WOD records"})
		return
	}

	// Delete mismatched records
	deletedCount := 0
	for _, id := range idsToDelete {
		err := h.userWorkoutWODRepo.Delete(id)
		if err != nil {
			h.logger.Error("Failed to delete WOD record", "id", id, "error", err)
			continue
		}
		deletedCount++
	}

	h.logger.Info("Deleted mismatched WOD records", "count", deletedCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted_count": deletedCount,
		"total_found":   len(idsToDelete),
	})
}

// UpdateWODRecordRequest represents the request payload for updating a WOD record
type UpdateWODRecordRequest struct {
	TimeSeconds *int     `json:"time_seconds"`
	Rounds      *int     `json:"rounds"`
	Reps        *int     `json:"reps"`
	Weight      *float64 `json:"weight"`
	Notes       string   `json:"notes"`
}

// UpdateWODRecord updates an individual WOD record
func (h *AdminHandler) UpdateWODRecord(w http.ResponseWriter, r *http.Request) {
	// Get record ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Error("Invalid record ID", "id", idStr, "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid record ID"})
		return
	}

	// Parse request body
	var req UpdateWODRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request body"})
		return
	}

	// Get the existing record to find the WOD ID
	existingRecord, err := h.userWorkoutWODRepo.GetByID(id)
	if err != nil {
		h.logger.Error("Failed to get existing WOD record", "id", id, "error", err)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "WOD record not found"})
		return
	}

	// Get the WOD definition to validate score_type
	wod, err := h.wodRepo.GetByID(existingRecord.WODID)
	if err != nil {
		h.logger.Error("Failed to get WOD definition", "wod_id", existingRecord.WODID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Failed to get WOD definition"})
		return
	}

	// Validate that the update matches the score_type
	scoreType := wod.ScoreType
	if scoreType == "Time (HH:MM:SS)" {
		if req.TimeSeconds == nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("WOD '%s' has score_type '%s' but time_seconds is missing", wod.Name, scoreType),
			})
			return
		}
		if req.Rounds != nil || req.Reps != nil || req.Weight != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("WOD '%s' has score_type '%s' but contains invalid fields (rounds/reps/weight)", wod.Name, scoreType),
			})
			return
		}
	} else if scoreType == "Rounds+Reps" {
		if req.Rounds == nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("WOD '%s' has score_type '%s' but rounds is missing", wod.Name, scoreType),
			})
			return
		}
		if req.TimeSeconds != nil || req.Weight != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("WOD '%s' has score_type '%s' but contains invalid fields (time_seconds/weight)", wod.Name, scoreType),
			})
			return
		}
	} else if scoreType == "Max Weight" {
		if req.Weight == nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("WOD '%s' has score_type '%s' but weight is missing", wod.Name, scoreType),
			})
			return
		}
		if req.TimeSeconds != nil || req.Rounds != nil || req.Reps != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("WOD '%s' has score_type '%s' but contains invalid fields (time_seconds/rounds/reps)", wod.Name, scoreType),
			})
			return
		}
	}

	// Update the record
	updatedRecord := &domain.UserWorkoutWOD{
		ID:            id,
		UserWorkoutID: existingRecord.UserWorkoutID,
		WODID:         existingRecord.WODID,
		TimeSeconds:   req.TimeSeconds,
		Rounds:        req.Rounds,
		Reps:          req.Reps,
		Weight:        req.Weight,
		Notes:         req.Notes,
		IsPR:          existingRecord.IsPR,
		OrderIndex:    existingRecord.OrderIndex,
	}

	if err := h.userWorkoutWODRepo.Update(updatedRecord); err != nil {
		h.logger.Error("Failed to update WOD record", "id", id, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Failed to update WOD record"})
		return
	}

	h.logger.Info("Updated WOD record", "id", id, "wod_name", wod.Name, "score_type", scoreType)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "WOD record updated successfully",
		"id":      id,
	})
}

// ========== User-Created Content Management ==========

// CopyToStandardRequest represents the request to copy an item to standard
type CopyToStandardRequest struct {
	NewName string `json:"new_name"`
}

// ListUserCreatedWODs lists all user-created WODs with creator info
func (h *AdminHandler) ListUserCreatedWODs(w http.ResponseWriter, r *http.Request) {
	// Parse pagination params
	limit := 100
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Parse filter params
	search := r.URL.Query().Get("search")
	scoreType := r.URL.Query().Get("score_type")
	creator := r.URL.Query().Get("creator")

	wods, count, err := h.wodService.ListAllUserCreatedWithUserInfoFiltered(limit, offset, search, scoreType, creator)
	if err != nil {
		h.logger.Error("Failed to list user-created WODs", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list user-created WODs"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  wods,
		"count": count,
	})
}

// CopyWODToStandard converts a user-created WOD to become a standard WOD
func (h *AdminHandler) CopyWODToStandard(w http.ResponseWriter, r *http.Request) {
	// Get WOD ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Error("Invalid WOD ID", "id", idStr, "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid WOD ID"})
		return
	}

	// Parse request body
	var req CopyToStandardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Copy to standard
	wod, err := h.wodService.CopyToStandard(id, req.NewName)
	if err != nil {
		h.logger.Error("Failed to copy WOD to standard", "id", id, "error", err)
		// Check if it's a duplicate name error
		if strings.Contains(err.Error(), "already exists") {
			w.WriteHeader(http.StatusConflict)
		} else if strings.Contains(err.Error(), "not found") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	h.logger.Info("Converted WOD to standard", "source_id", id, "new_id", wod.ID, "name", wod.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(wod)
}

// ListUserCreatedMovements lists all user-created movements with creator info
func (h *AdminHandler) ListUserCreatedMovements(w http.ResponseWriter, r *http.Request) {
	// Parse pagination params
	limit := 100
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Parse filter params
	search := r.URL.Query().Get("search")
	movementType := r.URL.Query().Get("type")
	creator := r.URL.Query().Get("creator")

	movements, count, err := h.movementService.ListAllUserCreatedWithUserInfoFiltered(limit, offset, search, movementType, creator)
	if err != nil {
		h.logger.Error("Failed to list user-created movements", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list user-created movements"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  movements,
		"count": count,
	})
}

// CopyMovementToStandard converts a user-created movement to become a standard movement
func (h *AdminHandler) CopyMovementToStandard(w http.ResponseWriter, r *http.Request) {
	// Get movement ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Error("Invalid movement ID", "id", idStr, "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid movement ID"})
		return
	}

	// Parse request body
	var req CopyToStandardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Copy to standard
	movement, err := h.movementService.CopyToStandard(id, req.NewName)
	if err != nil {
		h.logger.Error("Failed to copy movement to standard", "id", id, "error", err)
		// Check if it's a duplicate name error
		if strings.Contains(err.Error(), "already exists") {
			w.WriteHeader(http.StatusConflict)
		} else if strings.Contains(err.Error(), "not found") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	h.logger.Info("Converted movement to standard", "source_id", id, "new_id", movement.ID, "name", movement.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(movement)
}

// ListUserCreatedWorkouts lists all user-created workout templates with creator info
func (h *AdminHandler) ListUserCreatedWorkouts(w http.ResponseWriter, r *http.Request) {
	// Parse pagination params
	limit := 100
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Parse filter params
	search := r.URL.Query().Get("search")
	creator := r.URL.Query().Get("creator")

	workouts, count, err := h.workoutTemplateService.ListAllUserCreatedWithUserInfoFiltered(limit, offset, search, creator)
	if err != nil {
		h.logger.Error("Failed to list user-created workouts", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list user-created workouts"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  workouts,
		"count": count,
	})
}

// CopyWorkoutToStandard converts a user-created workout to become a standard workout
func (h *AdminHandler) CopyWorkoutToStandard(w http.ResponseWriter, r *http.Request) {
	// Get workout ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Error("Invalid workout ID", "id", idStr, "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid workout ID"})
		return
	}

	// Parse request body
	var req CopyToStandardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Copy to standard
	workout, err := h.workoutTemplateService.CopyToStandard(id, req.NewName)
	if err != nil {
		h.logger.Error("Failed to copy workout to standard", "id", id, "error", err)
		// Check if it's a duplicate name error
		if strings.Contains(err.Error(), "already exists") {
			w.WriteHeader(http.StatusConflict)
		} else if strings.Contains(err.Error(), "not found") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	h.logger.Info("Converted workout to standard", "source_id", id, "new_id", workout.ID, "name", workout.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(workout)
}
