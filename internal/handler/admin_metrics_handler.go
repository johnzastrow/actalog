package handler

import (
	"net/http"

	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
)

// AdminMetricsHandler handles admin metrics dashboard requests
type AdminMetricsHandler struct {
	metricsService *service.AdminMetricsService
	logger         *logger.Logger
}

// NewAdminMetricsHandler creates a new admin metrics handler
func NewAdminMetricsHandler(metricsService *service.AdminMetricsService, logger *logger.Logger) *AdminMetricsHandler {
	return &AdminMetricsHandler{
		metricsService: metricsService,
		logger:         logger,
	}
}

// GetAdminMetrics handles GET /api/admin/metrics
// Returns all admin dashboard metrics in a single response
func (h *AdminMetricsHandler) GetAdminMetrics(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("action=get_admin_metrics")

	metrics, err := h.metricsService.GetMetrics()
	if err != nil {
		h.logger.Error("action=get_admin_metrics outcome=failure error=%v", err)
		respondError(w, http.StatusInternalServerError, "Failed to retrieve metrics")
		return
	}

	h.logger.Info("action=get_admin_metrics outcome=success")
	respondJSON(w, http.StatusOK, metrics)
}
