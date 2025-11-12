package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
	"github.com/johnzastrow/actalog/pkg/prmath"
)

// PRHandler handles PR (Personal Record) endpoints
type PRHandler struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewPRHandler creates a new PR handler
func NewPRHandler(db *sql.DB, l *logger.Logger) *PRHandler {
	return &PRHandler{
		db:     db,
		logger: l,
	}
}

// PersonalRecord represents a unified PR from either movements or WODs
type PersonalRecord struct {
	Type           string   `json:"type"` // "movement" or "wod"
	ID             int64    `json:"id"`
	UserWorkoutID  int64    `json:"user_workout_id"`
	WorkoutDate    string   `json:"workout_date"`
	Name           string   `json:"name"`          // Movement name or WOD name
	MovementType   *string  `json:"movement_type,omitempty"` // For movements: weightlifting, gymnastics, etc.
	Weight         *float64 `json:"weight,omitempty"`        // For movements (actual weight lifted)
	Sets           *int     `json:"sets,omitempty"`          // For movements
	Reps           *int     `json:"reps,omitempty"`          // For movements
	Time           *int     `json:"time,omitempty"`          // For movements (seconds) OR WOD time
	Distance       *float64 `json:"distance,omitempty"`      // For movements
	Calculated1RM  *float64 `json:"calculated_1rm,omitempty"` // Calculated one-rep max for movements
	Formula        *string  `json:"formula,omitempty"`       // Which formula was used (e.g., "Epley (2-10 reps)")
	ScoreValue     *string  `json:"score_value,omitempty"`   // For WODs
	Division       *string  `json:"division,omitempty"`      // For WODs (rx, scaled, etc.)
	WODType        *string  `json:"wod_type,omitempty"`      // For WODs
	WODScoreType   *string  `json:"wod_score_type,omitempty"` // For WODs
}

// MovementPRSummary represents PR summary for a specific movement
type MovementPRSummary struct {
	MovementID    int64    `json:"movement_id"`
	MovementName  string   `json:"movement_name"`
	MovementType  string   `json:"movement_type"`
	PRCount       int      `json:"pr_count"`
	Best1RM       *float64 `json:"best_1rm,omitempty"`        // Highest calculated 1RM
	BestFormula   *string  `json:"best_formula,omitempty"`    // Formula used for best 1RM
	BestWeight    *float64 `json:"best_weight,omitempty"`     // Actual weight lifted for best PR
	BestSets      *int     `json:"best_sets,omitempty"`       // Sets for best PR
	BestReps      *int     `json:"best_reps,omitempty"`       // Reps for best PR
	LastPRDate    string   `json:"last_pr_date"`
}

// GetPersonalRecords retrieves all PRs for the authenticated user
func (h *PRHandler) GetPersonalRecords(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse optional limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 200 {
				limit = 200 // cap at 200
			}
		}
	}

	if h.logger != nil {
		h.logger.Info("action=get_prs user_id=%d limit=%d", userID, limit)
	}

	// Query to get PRs from workout_movements
	movementQuery := `
		SELECT
			'movement' as type,
			wm.id,
			uw.id as user_workout_id,
			uw.workout_date,
			m.name,
			m.type as movement_type,
			wm.weight,
			wm.sets,
			wm.reps,
			wm.time,
			wm.distance,
			NULL as score_value,
			NULL as division,
			NULL as wod_type,
			NULL as wod_score_type
		FROM workout_movements wm
		JOIN workouts w ON wm.workout_id = w.id
		JOIN user_workouts uw ON uw.workout_id = w.id
		JOIN movements m ON wm.movement_id = m.id
		WHERE uw.user_id = ? AND wm.is_pr = 1
		ORDER BY uw.workout_date DESC
		LIMIT ?
	`

	// Query to get PRs from workout_wods
	wodQuery := `
		SELECT
			'wod' as type,
			ww.id,
			uw.id as user_workout_id,
			uw.workout_date,
			wod.name,
			NULL as movement_type,
			NULL as weight,
			NULL as sets,
			NULL as reps,
			NULL as time,
			NULL as distance,
			ww.score_value,
			ww.division,
			wod.type as wod_type,
			wod.score_type as wod_score_type
		FROM workout_wods ww
		JOIN workouts w ON ww.workout_id = w.id
		JOIN user_workouts uw ON uw.workout_id = w.id
		JOIN wods wod ON ww.wod_id = wod.id
		WHERE uw.user_id = ? AND ww.is_pr = 1
		ORDER BY uw.workout_date DESC
		LIMIT ?
	`

	// Execute both queries
	var prs []PersonalRecord

	// Get movement PRs
	movementRows, err := h.db.Query(movementQuery, userID, limit)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=get_prs outcome=failure user_id=%d error=query_movements: %v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve PRs")
		return
	}
	defer movementRows.Close()

	for movementRows.Next() {
		var pr PersonalRecord
		err := movementRows.Scan(
			&pr.Type,
			&pr.ID,
			&pr.UserWorkoutID,
			&pr.WorkoutDate,
			&pr.Name,
			&pr.MovementType,
			&pr.Weight,
			&pr.Sets,
			&pr.Reps,
			&pr.Time,
			&pr.Distance,
			&pr.ScoreValue,
			&pr.Division,
			&pr.WODType,
			&pr.WODScoreType,
		)
		if err != nil {
			if h.logger != nil {
				h.logger.Error("action=get_prs outcome=failure user_id=%d error=scan_movement: %v", userID, err)
			}
			continue
		}

		// Calculate 1RM for movements with weight and reps
		if pr.Weight != nil && pr.Reps != nil && *pr.Weight > 0 && *pr.Reps > 0 {
			oneRM, formula := prmath.Calculate1RM(*pr.Weight, *pr.Reps)
			pr.Calculated1RM = &oneRM
			formulaStr := string(formula)
			pr.Formula = &formulaStr
		}

		prs = append(prs, pr)
	}

	// Get WOD PRs
	wodRows, err := h.db.Query(wodQuery, userID, limit)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=get_prs outcome=failure user_id=%d error=query_wods: %v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve PRs")
		return
	}
	defer wodRows.Close()

	for wodRows.Next() {
		var pr PersonalRecord
		err := wodRows.Scan(
			&pr.Type,
			&pr.ID,
			&pr.UserWorkoutID,
			&pr.WorkoutDate,
			&pr.Name,
			&pr.MovementType,
			&pr.Weight,
			&pr.Sets,
			&pr.Reps,
			&pr.Time,
			&pr.Distance,
			&pr.ScoreValue,
			&pr.Division,
			&pr.WODType,
			&pr.WODScoreType,
		)
		if err != nil {
			if h.logger != nil {
				h.logger.Error("action=get_prs outcome=failure user_id=%d error=scan_wod: %v", userID, err)
			}
			continue
		}
		prs = append(prs, pr)
	}

	if h.logger != nil {
		h.logger.Info("action=get_prs outcome=success user_id=%d count=%d", userID, len(prs))
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"prs": prs,
	})
}

// GetPRMovements retrieves movements with PR counts for the authenticated user
func (h *PRHandler) GetPRMovements(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse optional limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 20 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100 // cap at 100
			}
		}
	}

	if h.logger != nil {
		h.logger.Info("action=get_pr_movements user_id=%d limit=%d", userID, limit)
	}

	query := `
		SELECT
			m.id as movement_id,
			m.name as movement_name,
			m.type as movement_type,
			COUNT(wm.id) as pr_count,
			MAX(wm.weight) as best_weight,
			MAX(wm.sets) as best_sets,
			MAX(wm.reps) as best_reps,
			MAX(uw.workout_date) as last_pr_date
		FROM workout_movements wm
		JOIN workouts w ON wm.workout_id = w.id
		JOIN user_workouts uw ON uw.workout_id = w.id
		JOIN movements m ON wm.movement_id = m.id
		WHERE uw.user_id = ? AND wm.is_pr = 1
		GROUP BY m.id, m.name, m.type
		ORDER BY pr_count DESC, last_pr_date DESC
		LIMIT ?
	`

	rows, err := h.db.Query(query, userID, limit)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=get_pr_movements outcome=failure user_id=%d error=%v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve PR movements")
		return
	}
	defer rows.Close()

	var movements []MovementPRSummary
	for rows.Next() {
		var m MovementPRSummary
		err := rows.Scan(
			&m.MovementID,
			&m.MovementName,
			&m.MovementType,
			&m.PRCount,
			&m.BestWeight,
			&m.BestSets,
			&m.BestReps,
			&m.LastPRDate,
		)
		if err != nil {
			if h.logger != nil {
				h.logger.Error("action=get_pr_movements outcome=failure user_id=%d error=scan: %v", userID, err)
			}
			continue
		}

		// Calculate 1RM for the best lift (this is simplified - ideally we'd check all PRs for this movement)
		if m.BestWeight != nil && m.BestReps != nil && *m.BestWeight > 0 && *m.BestReps > 0 {
			oneRM, formula := prmath.Calculate1RM(*m.BestWeight, *m.BestReps)
			m.Best1RM = &oneRM
			formulaStr := string(formula)
			m.BestFormula = &formulaStr
		}

		movements = append(movements, m)
	}

	if h.logger != nil {
		h.logger.Info("action=get_pr_movements outcome=success user_id=%d count=%d", userID, len(movements))
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"movements": movements,
	})
}

// ToggleMovementPR toggles the PR flag for a specific workout_movement
func (h *PRHandler) ToggleMovementPR(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get workout_movement ID from URL parameter
	movementIDStr := r.URL.Query().Get("id")
	if movementIDStr == "" {
		respondError(w, http.StatusBadRequest, "Missing movement ID")
		return
	}

	movementID, err := strconv.ParseInt(movementIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid movement ID")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=toggle_movement_pr_attempt user_id=%d movement_id=%d", userID, movementID)
	}

	// Verify user owns this workout
	verifyQuery := `
		SELECT 1
		FROM workout_movements wm
		JOIN workouts w ON wm.workout_id = w.id
		JOIN user_workouts uw ON uw.workout_id = w.id
		WHERE wm.id = ? AND uw.user_id = ?
	`

	var exists int
	err = h.db.QueryRow(verifyQuery, movementID, userID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			if h.logger != nil {
				h.logger.Warn("action=toggle_movement_pr outcome=failure user_id=%d movement_id=%d reason=not_found", userID, movementID)
			}
			respondError(w, http.StatusNotFound, "Movement not found")
			return
		}
		if h.logger != nil {
			h.logger.Error("action=toggle_movement_pr outcome=failure user_id=%d movement_id=%d error=%v", userID, movementID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify ownership")
		return
	}

	// Toggle the PR flag
	toggleQuery := `
		UPDATE workout_movements
		SET is_pr = CASE WHEN is_pr = 1 THEN 0 ELSE 1 END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := h.db.Exec(toggleQuery, movementID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=toggle_movement_pr outcome=failure user_id=%d movement_id=%d error=%v", userID, movementID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to toggle PR flag")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "Movement not found")
		return
	}

	// Get the new state
	var newState bool
	err = h.db.QueryRow("SELECT is_pr FROM workout_movements WHERE id = ?", movementID).Scan(&newState)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=toggle_movement_pr outcome=failure user_id=%d movement_id=%d error=get_state: %v", userID, movementID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to get PR state")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=toggle_movement_pr outcome=success user_id=%d movement_id=%d new_state=%t", userID, movementID, newState)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "PR flag toggled successfully",
		"is_pr":   newState,
	})
}
