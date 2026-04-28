package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
)

func TestHTTPError(t *testing.T) {
	t.Run("Error with message only", func(t *testing.T) {
		err := NewHTTPError(http.StatusNotFound, "resource not found", nil)
		if err.Error() != "resource not found" {
			t.Errorf("Error() = %q, want %q", err.Error(), "resource not found")
		}
		if err.Status != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", err.Status, http.StatusNotFound)
		}
	})

	t.Run("Error with wrapped error", func(t *testing.T) {
		wrapped := errors.New("original error")
		err := NewHTTPError(http.StatusBadRequest, "bad request", wrapped)
		if err.Error() != "original error" {
			t.Errorf("Error() = %q, want %q", err.Error(), "original error")
		}
		if err.Unwrap() != wrapped {
			t.Error("Unwrap() should return the wrapped error")
		}
	})
}

func TestMapServiceError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedKnown  bool
	}{
		{"user not found", service.ErrUserNotFound, http.StatusNotFound, true},
		{"invalid credentials", service.ErrInvalidCredentials, http.StatusUnauthorized, true},
		{"email exists", service.ErrEmailAlreadyExists, http.StatusConflict, true},
		{"registration closed", service.ErrRegistrationClosed, http.StatusForbidden, true},
		{"movement not found", service.ErrMovementNotFound, http.StatusNotFound, true},
		{"movement unauthorized", service.ErrMovementUnauthorized, http.StatusForbidden, true},
		{"wod not found", service.ErrWODNotFound, http.StatusNotFound, true},
		{"wod duplicate name", service.ErrWODDuplicateName, http.StatusConflict, true},
		{"workout not found", service.ErrWorkoutNotFound, http.StatusNotFound, true},
		{"user workout not found", service.ErrUserWorkoutNotFound, http.StatusNotFound, true},
		{"org not found", service.ErrOrganizationNotFound, http.StatusNotFound, true},
		{"subscription not found", service.ErrSubscriptionNotFound, http.StatusNotFound, true},
		{"unknown error", errors.New("unknown error"), http.StatusInternalServerError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, known := MapServiceError(tt.err)
			if status != tt.expectedStatus {
				t.Errorf("MapServiceError() status = %d, want %d", status, tt.expectedStatus)
			}
			if known != tt.expectedKnown {
				t.Errorf("MapServiceError() known = %v, want %v", known, tt.expectedKnown)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	t.Run("known service error", func(t *testing.T) {
		w := httptest.NewRecorder()
		WriteError(w, service.ErrUserNotFound)

		if w.Code != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
		}

		var resp ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if resp.Message != "user not found" {
			t.Errorf("Message = %q, want %q", resp.Message, "user not found")
		}
	})

	t.Run("HTTPError", func(t *testing.T) {
		w := httptest.NewRecorder()
		httpErr := NewHTTPError(http.StatusTeapot, "I'm a teapot", nil)
		WriteError(w, httpErr)

		if w.Code != http.StatusTeapot {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusTeapot)
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		w := httptest.NewRecorder()
		WriteError(w, errors.New("some random error"))

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestWriteErrorWithStatus(t *testing.T) {
	w := httptest.NewRecorder()
	WriteErrorWithStatus(w, http.StatusBadRequest, "invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Content-Type should be application/json")
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if resp.Message != "invalid input" {
		t.Errorf("Message = %q, want %q", resp.Message, "invalid input")
	}
}

func TestWriteErrorWithDetail(t *testing.T) {
	w := httptest.NewRecorder()
	WriteErrorWithDetail(w, http.StatusBadRequest, "validation failed", "email is required")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if resp.Message != "validation failed" {
		t.Errorf("Message = %q, want %q", resp.Message, "validation failed")
	}
	if resp.Error != "email is required" {
		t.Errorf("Error = %q, want %q", resp.Error, "email is required")
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"status": "ok"}
	WriteJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Content-Type should be application/json")
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want %q", resp["status"], "ok")
	}
}

func TestHandleServiceError(t *testing.T) {
	t.Run("nil error returns false", func(t *testing.T) {
		w := httptest.NewRecorder()
		handled := HandleServiceError(w, nil, "test")
		if handled {
			t.Error("HandleServiceError(nil) should return false")
		}
		if w.Code != http.StatusOK {
			t.Errorf("Status should not be set for nil error, got %d", w.Code)
		}
	})

	t.Run("known error", func(t *testing.T) {
		w := httptest.NewRecorder()
		handled := HandleServiceError(w, service.ErrUserNotFound, "getting user")
		if !handled {
			t.Error("HandleServiceError should return true for non-nil error")
		}
		if w.Code != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unknown error hides details", func(t *testing.T) {
		w := httptest.NewRecorder()
		handled := HandleServiceError(w, errors.New("internal database error"), "processing")
		if !handled {
			t.Error("HandleServiceError should return true for non-nil error")
		}
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
		}

		var resp ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		// Should not expose internal error details
		if resp.Message == "internal database error" {
			t.Error("Should not expose internal error details")
		}
	})
}

// TestWriteError_MapsErrProtectedUserTo403 verifies that domain.ErrProtectedUser
// is mapped to HTTP 403 with a structured body containing the machine-readable
// error code "protected_user" and a non-empty documentation_url.
func TestWriteError_MapsErrProtectedUserTo403(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, domain.ErrProtectedUser)

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusForbidden)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if resp.Error != "protected_user" {
		t.Errorf("Error = %q, want %q", resp.Error, "protected_user")
	}
	if resp.DocumentationURL == "" {
		t.Error("DocumentationURL must not be empty for ErrProtectedUser")
	}
	if resp.Message == "" {
		t.Error("Message must not be empty for ErrProtectedUser")
	}
}

// TestWriteError_MapsErrProtectedUserWrappedTo403 verifies that a wrapped
// domain.ErrProtectedUser is still correctly mapped via errors.Is semantics.
func TestWriteError_MapsErrProtectedUserWrappedTo403(t *testing.T) {
	wrapped := errors.Join(errors.New("outer context"), domain.ErrProtectedUser)
	w := httptest.NewRecorder()
	WriteError(w, wrapped)

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d (wrapped ErrProtectedUser)", w.Code, http.StatusForbidden)
	}
}

// TestWriteError_MapsErrConflictTo409 verifies that domain.ErrConflict is
// mapped to HTTP 409 with a structured body containing error code "conflict".
func TestWriteError_MapsErrConflictTo409(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, domain.ErrConflict)

	if w.Code != http.StatusConflict {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusConflict)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if resp.Error != "conflict" {
		t.Errorf("Error = %q, want %q", resp.Error, "conflict")
	}
	if resp.Message == "" {
		t.Error("Message must not be empty for ErrConflict")
	}
}

func TestCommonHTTPErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            *HTTPError
		expectedStatus int
	}{
		{"BadRequest", ErrBadRequest, http.StatusBadRequest},
		{"Unauthorized", ErrUnauthorized, http.StatusUnauthorized},
		{"Forbidden", ErrForbidden, http.StatusForbidden},
		{"NotFound", ErrNotFound, http.StatusNotFound},
		{"Conflict", ErrConflict, http.StatusConflict},
		{"PaymentRequired", ErrPaymentRequired, http.StatusPaymentRequired},
		{"InternalServerError", ErrInternalServerError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Status != tt.expectedStatus {
				t.Errorf("%s.Status = %d, want %d", tt.name, tt.err.Status, tt.expectedStatus)
			}
		})
	}
}
