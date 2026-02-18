package handler

import (
	"net/http"

	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// ConsistencyHandler handles consistency stats endpoints
type ConsistencyHandler struct {
	service *service.ConsistencyService
	logger  *logger.Logger
}

// NewConsistencyHandler creates a new consistency handler
func NewConsistencyHandler(service *service.ConsistencyService, logger *logger.Logger) *ConsistencyHandler {
	return &ConsistencyHandler{service: service, logger: logger}
}

// GetConsistencyStats returns the current user's consistency stats
func (h *ConsistencyHandler) GetConsistencyStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stats, err := h.service.GetUserStats(userID)
	if err != nil {
		h.logger.Error("Failed to get consistency stats: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get consistency stats")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}
