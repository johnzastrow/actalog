package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// LeaderboardHandler handles leaderboard endpoints
type LeaderboardHandler struct {
	service *service.LeaderboardService
	logger  *logger.Logger
}

// NewLeaderboardHandler creates a new leaderboard handler
func NewLeaderboardHandler(service *service.LeaderboardService, logger *logger.Logger) *LeaderboardHandler {
	return &LeaderboardHandler{service: service, logger: logger}
}

// GetMovementLeaderboard returns the leaderboard for a movement
func (h *LeaderboardHandler) GetMovementLeaderboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	movementID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid movement ID")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "all_time"
	}

	limit := 50
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && o >= 0 {
		offset = o
	}

	result, err := h.service.GetMovementLeaderboard(userID, movementID, period, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get movement leaderboard: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get leaderboard")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetWODLeaderboard returns the leaderboard for a WOD
func (h *LeaderboardHandler) GetWODLeaderboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	wodID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid WOD ID")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "all_time"
	}

	limit := 50
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && o >= 0 {
		offset = o
	}

	result, err := h.service.GetWODLeaderboard(userID, wodID, period, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get WOD leaderboard: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get leaderboard")
		return
	}

	respondJSON(w, http.StatusOK, result)
}
