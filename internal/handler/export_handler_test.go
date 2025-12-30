package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
