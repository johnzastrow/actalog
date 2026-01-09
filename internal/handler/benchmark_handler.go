package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// BenchmarkHandler handles HTTP requests for benchmarks
type BenchmarkHandler struct {
	service *service.BenchmarkService
	logger  *logger.Logger
}

// NewBenchmarkHandler creates a new benchmark handler
func NewBenchmarkHandler(service *service.BenchmarkService, logger *logger.Logger) *BenchmarkHandler {
	return &BenchmarkHandler{
		service: service,
		logger:  logger,
	}
}

// RunBenchmark handles POST /api/benchmark
// @Summary      Run benchmark suite
// @Description  Execute the full benchmark suite to test database, serialization, and business logic performance
// @Tags         benchmark
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        concurrent query bool false "Include concurrent operation tests (default: false)"
// @Param        cleanup query bool false "Skip cleanup after run if set to false (default: true)"
// @Param        records query int false "Number of records to generate for benchmark testing (default: 1000, max: 500000)"
// @Success      200 {object} domain.BenchmarkSuiteResult "Benchmark results"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /benchmark [post]
func (h *BenchmarkHandler) RunBenchmark(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.logger.Warn("Unauthorized benchmark attempt: no user context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse record count from query parameters
	recordCount := domain.DefaultRecordCount
	if recordsStr := r.URL.Query().Get("records"); recordsStr != "" {
		if parsed, err := strconv.Atoi(recordsStr); err == nil && parsed > 0 {
			recordCount = parsed
			if recordCount > domain.MaxRecordCount {
				recordCount = domain.MaxRecordCount
			}
		}
	}

	// Parse options from query parameters
	options := domain.BenchmarkOptions{
		IncludeConcurrent: r.URL.Query().Get("concurrent") == "true",
		Cleanup:           r.URL.Query().Get("cleanup") != "false", // Default to true
		RecordCount:       recordCount,
	}

	h.logger.Info("Running benchmark suite: user_id=%d concurrent=%v cleanup=%v records=%d", userID, options.IncludeConcurrent, options.Cleanup, options.RecordCount)

	result, err := h.service.RunBenchmark(userID, options)
	if err != nil {
		h.logger.Error("Benchmark failed: user_id=%d error=%v", userID, err)
		http.Error(w, "Benchmark failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Benchmark completed: user_id=%d overall=%s duration_ms=%.2f ops=%d success=%d failed=%d records=%d",
		userID, result.Overall, result.TotalDurationMs, result.TotalOperations, result.SuccessfulOperations, result.FailedOperations, result.RecordCount)

	// Buffer the JSON response to set Content-Length header
	// This avoids chunked transfer encoding issues with long-running requests
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(result); err != nil {
		h.logger.Error("Failed to encode benchmark result: error=%v", err)
		http.Error(w, "Failed to encode result", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.Write(buf.Bytes())
}

// GetBenchmarkStatus handles GET /api/benchmark/status
// @Summary      Get benchmark status
// @Description  Quick status check for the benchmark system
// @Tags         benchmark
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "Benchmark status"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /benchmark/status [get]
func (h *BenchmarkHandler) GetBenchmarkStatus(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.logger.Warn("Unauthorized benchmark status request: no user context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	status, err := h.service.GetBenchmarkStatus()
	if err != nil {
		h.logger.Error("Failed to get benchmark status: error=%v", err)
		http.Error(w, "Failed to get status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// CleanupBenchmarkData handles DELETE /api/admin/benchmark/data
// @Summary      Cleanup all benchmark data (Admin)
// @Description  Delete all benchmark data from the database
// @Tags         benchmark
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]string "Cleanup result"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/benchmark/data [delete]
func (h *BenchmarkHandler) CleanupBenchmarkData(w http.ResponseWriter, r *http.Request) {
	userRole, _ := middleware.GetUserRole(r.Context())
	if userRole != "admin" {
		h.logger.Warn("Unauthorized benchmark cleanup attempt: user_role=%s", userRole)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err := h.service.CleanupAllBenchmarkData()
	if err != nil {
		h.logger.Error("Failed to cleanup benchmark data: error=%v", err)
		http.Error(w, "Failed to cleanup: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Benchmark data cleaned up by admin")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "All benchmark data has been deleted",
	})
}
