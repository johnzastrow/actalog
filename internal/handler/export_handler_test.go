package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/johnzastrow/actalog/internal/service"
)

func createTestExportService() *service.ExportService {
	mockWODRepo := NewMockWODRepository()
	mockMovementRepo := NewMockMovementRepository()
	mockUserRepo := NewMockUserRepository()
	mockUserWorkoutRepo := NewMockUserWorkoutRepository()
	return service.NewExportService(mockWODRepo, mockMovementRepo, mockUserRepo, mockUserWorkoutRepo)
}

func TestExportHandler_ExportWODs_Unauthorized(t *testing.T) {
	handler := &ExportHandler{}

	req := createTestRequest(http.MethodGet, "/api/export/wods", "")
	rr := httptest.NewRecorder()

	handler.ExportWODs(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestExportHandler_ExportMovements_Unauthorized(t *testing.T) {
	handler := &ExportHandler{}

	req := createTestRequest(http.MethodGet, "/api/export/movements", "")
	rr := httptest.NewRecorder()

	handler.ExportMovements(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestExportHandler_ExportUserWorkouts_Unauthorized(t *testing.T) {
	handler := &ExportHandler{}

	req := createTestRequest(http.MethodGet, "/api/export/user-workouts", "")
	rr := httptest.NewRecorder()

	handler.ExportUserWorkouts(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestExportHandler_ExportUserWorkouts_InvalidStartDate(t *testing.T) {
	handler := &ExportHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/export/user-workouts?start_date=invalid&end_date=2024-12-31", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ExportUserWorkouts(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid start_date format")
}

func TestExportHandler_ExportUserWorkouts_InvalidEndDate(t *testing.T) {
	handler := &ExportHandler{}

	req := createAuthenticatedRequest(http.MethodGet, "/api/export/user-workouts?start_date=2024-01-01&end_date=invalid", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ExportUserWorkouts(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid end_date format")
}

func TestExportHandler_ExportUserWorkouts_MissingOneDateParam(t *testing.T) {
	handler := &ExportHandler{}

	tests := []struct {
		name      string
		query     string
		wantError string
	}{
		{
			name:      "only start_date",
			query:     "?start_date=2024-01-01",
			wantError: "Both start_date and end_date must be provided together",
		},
		{
			name:      "only end_date",
			query:     "?end_date=2024-12-31",
			wantError: "Both start_date and end_date must be provided together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/export/user-workouts"+tt.query, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.ExportUserWorkouts(rr, req)

			assertStatusCode(t, rr, http.StatusBadRequest)
			assertBodyContains(t, rr, tt.wantError)
		})
	}
}

func TestParseBoolParam(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		defaultValue bool
		want         bool
	}{
		{name: "empty with true default", value: "", defaultValue: true, want: true},
		{name: "empty with false default", value: "", defaultValue: false, want: false},
		{name: "true", value: "true", defaultValue: false, want: true},
		{name: "false", value: "false", defaultValue: true, want: false},
		{name: "1", value: "1", defaultValue: false, want: true},
		{name: "0", value: "0", defaultValue: true, want: false},
		{name: "invalid", value: "invalid", defaultValue: true, want: true},
		{name: "invalid with false default", value: "invalid", defaultValue: false, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBoolParam(tt.value, tt.defaultValue)
			if got != tt.want {
				t.Errorf("parseBoolParam(%q, %v) = %v, want %v", tt.value, tt.defaultValue, got, tt.want)
			}
		})
	}
}

func TestNewExportHandler(t *testing.T) {
	handler := NewExportHandler(nil)
	if handler == nil {
		t.Error("NewExportHandler should return a non-nil handler")
	}
}

func TestExportHandler_ExportWODs_MissingRole(t *testing.T) {
	handler := &ExportHandler{}

	// Create request with userID but no role context
	req := createUserIDOnlyRequest(http.MethodGet, "/api/export/wods", "", 1)
	rr := httptest.NewRecorder()

	handler.ExportWODs(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestExportHandler_ExportMovements_MissingRole(t *testing.T) {
	handler := &ExportHandler{}

	// Create request with userID but no role context
	req := createUserIDOnlyRequest(http.MethodGet, "/api/export/movements", "", 1)
	rr := httptest.NewRecorder()

	handler.ExportMovements(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

// Removed 5 panic-expectation tests:
// - TestExportHandler_ExportWODs_WithQueryParams (6 subtests)
// - TestExportHandler_ExportWODs_AsAdmin
// - TestExportHandler_ExportMovements_WithQueryParams (6 subtests)
// - TestExportHandler_ExportMovements_AsAdmin
// - TestExportHandler_ExportUserWorkouts_WithFormatParams (5 subtests)
// These tests verified nil pointer panics, not business logic.

// ===== Success path tests with real service =====

func TestExportHandler_ExportWODs_Success_CSV(t *testing.T) {
	svc := createTestExportService()
	handler := NewExportHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/export/wods", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ExportWODs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	// Check for CSV content type
	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/csv") {
		t.Errorf("Expected Content-Type text/csv, got %s", contentType)
	}
	// Check for Content-Disposition header
	disposition := rr.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, "wods_export.csv") {
		t.Errorf("Expected Content-Disposition to contain wods_export.csv, got %s", disposition)
	}
}

func TestExportHandler_ExportWODs_Success_JSON(t *testing.T) {
	svc := createTestExportService()
	handler := NewExportHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/export/wods?format=json", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ExportWODs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	// Check for JSON content type
	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestExportHandler_ExportWODs_WithQueryParams(t *testing.T) {
	svc := createTestExportService()
	handler := NewExportHandler(svc)

	tests := []struct {
		name  string
		query string
	}{
		{"include_standard_true", "?include_standard=true"},
		{"include_standard_false", "?include_standard=false"},
		{"include_custom_true", "?include_custom=true"},
		{"include_custom_false", "?include_custom=false"},
		{"all_params", "?include_standard=true&include_custom=true&format=csv"},
		{"json_format", "?format=json"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/export/wods"+tc.query, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.ExportWODs(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
		})
	}
}

func TestExportHandler_ExportWODs_AsAdmin(t *testing.T) {
	svc := createTestExportService()
	handler := NewExportHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/export/wods", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ExportWODs(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestExportHandler_ExportMovements_Success_CSV(t *testing.T) {
	svc := createTestExportService()
	handler := NewExportHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/export/movements", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ExportMovements(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	// Check for CSV content type
	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/csv") {
		t.Errorf("Expected Content-Type text/csv, got %s", contentType)
	}
	// Check for Content-Disposition header
	disposition := rr.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, "movements_export.csv") {
		t.Errorf("Expected Content-Disposition to contain movements_export.csv, got %s", disposition)
	}
}

func TestExportHandler_ExportMovements_Success_JSON(t *testing.T) {
	svc := createTestExportService()
	handler := NewExportHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/export/movements?format=json", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ExportMovements(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	// Check for JSON content type
	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestExportHandler_ExportMovements_WithQueryParams(t *testing.T) {
	svc := createTestExportService()
	handler := NewExportHandler(svc)

	tests := []struct {
		name  string
		query string
	}{
		{"include_standard_true", "?include_standard=true"},
		{"include_standard_false", "?include_standard=false"},
		{"include_custom_true", "?include_custom=true"},
		{"include_custom_false", "?include_custom=false"},
		{"all_params", "?include_standard=true&include_custom=true&format=csv"},
		{"json_format", "?format=json"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/export/movements"+tc.query, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.ExportMovements(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
		})
	}
}

func TestExportHandler_ExportMovements_AsAdmin(t *testing.T) {
	svc := createTestExportService()
	handler := NewExportHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/export/movements", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ExportMovements(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestExportHandler_ExportUserWorkouts_Success_JSON(t *testing.T) {
	svc := createTestExportService()
	handler := NewExportHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/export/user-workouts", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ExportUserWorkouts(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	// Check for JSON content type (default format)
	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestExportHandler_ExportUserWorkouts_Success_CSV(t *testing.T) {
	svc := createTestExportService()
	handler := NewExportHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/export/user-workouts?format=csv", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ExportUserWorkouts(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	// Check for CSV content type
	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/csv") {
		t.Errorf("Expected Content-Type text/csv, got %s", contentType)
	}
}

func TestExportHandler_ExportUserWorkouts_WithDateRange(t *testing.T) {
	svc := createTestExportService()
	handler := NewExportHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/export/user-workouts?start_date=2024-01-01&end_date=2024-12-31", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ExportUserWorkouts(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestExportHandler_ExportUserWorkouts_WithAllParams(t *testing.T) {
	svc := createTestExportService()
	handler := NewExportHandler(svc)

	tests := []struct {
		name  string
		query string
	}{
		{"csv_format", "?format=csv"},
		{"json_format", "?format=json"},
		{"with_date_range_csv", "?start_date=2024-01-01&end_date=2024-12-31&format=csv"},
		{"with_date_range_json", "?start_date=2024-01-01&end_date=2024-12-31&format=json"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodGet, "/api/export/user-workouts"+tc.query, "", 1, "test@example.com", "user")
			rr := httptest.NewRecorder()

			handler.ExportUserWorkouts(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
		})
	}
}

func TestExportHandler_ExportWODs_ServiceError(t *testing.T) {
	mockWODRepo := NewMockWODRepository()
	mockWODRepo.SetError(ErrMockInternalError)
	mockMovementRepo := NewMockMovementRepository()
	mockUserRepo := NewMockUserRepository()
	mockUserWorkoutRepo := NewMockUserWorkoutRepository()
	svc := service.NewExportService(mockWODRepo, mockMovementRepo, mockUserRepo, mockUserWorkoutRepo)
	handler := NewExportHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/export/wods", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ExportWODs(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestExportHandler_ExportMovements_ServiceError(t *testing.T) {
	mockWODRepo := NewMockWODRepository()
	mockMovementRepo := NewMockMovementRepository()
	mockMovementRepo.SetError(ErrMockInternalError)
	mockUserRepo := NewMockUserRepository()
	mockUserWorkoutRepo := NewMockUserWorkoutRepository()
	svc := service.NewExportService(mockWODRepo, mockMovementRepo, mockUserRepo, mockUserWorkoutRepo)
	handler := NewExportHandler(svc)

	req := createAuthenticatedRequest(http.MethodGet, "/api/export/movements", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.ExportMovements(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}
