// Package handler provides HTTP error handling utilities
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
)

// ErrorResponse represents an error response
// @Description Error response returned when a request fails
type ErrorResponse struct {
	Message          string `json:"message" example:"Invalid credentials"`
	Error            string `json:"error,omitempty" example:"additional error details"`
	DocumentationURL string `json:"documentation_url,omitempty" example:"/docs/protected-users"`
}

// HTTPError represents an error with an associated HTTP status code
type HTTPError struct {
	Status  int
	Message string
	Err     error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError creates a new HTTPError
func NewHTTPError(status int, message string, err error) *HTTPError {
	return &HTTPError{
		Status:  status,
		Message: message,
		Err:     err,
	}
}

// Common HTTP errors
var (
	ErrBadRequest          = &HTTPError{Status: http.StatusBadRequest, Message: "bad request"}
	ErrUnauthorized        = &HTTPError{Status: http.StatusUnauthorized, Message: "unauthorized"}
	ErrForbidden           = &HTTPError{Status: http.StatusForbidden, Message: "forbidden"}
	ErrNotFound            = &HTTPError{Status: http.StatusNotFound, Message: "not found"}
	ErrConflict            = &HTTPError{Status: http.StatusConflict, Message: "conflict"}
	ErrPaymentRequired     = &HTTPError{Status: http.StatusPaymentRequired, Message: "subscription required"}
	ErrInternalServerError = &HTTPError{Status: http.StatusInternalServerError, Message: "internal server error"}
)

// errorStatusMap maps service errors to HTTP status codes
var errorStatusMap = map[error]int{
	// User/Auth errors
	service.ErrUserNotFound:             http.StatusNotFound,
	service.ErrInvalidCredentials:       http.StatusUnauthorized,
	service.ErrEmailAlreadyExists:       http.StatusConflict,
	service.ErrRegistrationClosed:       http.StatusForbidden,
	service.ErrInvalidResetToken:        http.StatusBadRequest,
	service.ErrResetTokenExpired:        http.StatusBadRequest,
	service.ErrInvalidVerificationToken: http.StatusBadRequest,
	service.ErrVerificationTokenExpired: http.StatusBadRequest,
	service.ErrEmailAlreadyVerified:     http.StatusConflict,
	service.ErrInvalidRefreshToken:      http.StatusUnauthorized,
	service.ErrAccountLocked:            http.StatusForbidden,
	service.ErrAccountDisabled:          http.StatusForbidden,

	// Movement errors
	service.ErrMovementNotFound:     http.StatusNotFound,
	service.ErrMovementUnauthorized: http.StatusForbidden,
	service.ErrMovementNameRequired: http.StatusBadRequest,
	service.ErrMovementTypeRequired: http.StatusBadRequest,

	// WOD errors
	service.ErrWODNotFound:       http.StatusNotFound,
	service.ErrWODUnauthorized:   http.StatusForbidden,
	service.ErrWODOwnership:      http.StatusForbidden,
	service.ErrWODNameRequired:   http.StatusBadRequest,
	service.ErrWODSourceRequired: http.StatusBadRequest,
	service.ErrWODTypeRequired:   http.StatusBadRequest,
	service.ErrWODDuplicateName:  http.StatusConflict,

	// Workout errors
	service.ErrWorkoutNotFound: http.StatusNotFound,
	service.ErrUnauthorized:    http.StatusForbidden,

	// User workout errors
	service.ErrUserWorkoutNotFound:       http.StatusNotFound,
	service.ErrUnauthorizedWorkoutAccess: http.StatusForbidden,

	// Organization errors
	service.ErrOrganizationNotFound:   http.StatusNotFound,
	service.ErrOrganizationNameExists: http.StatusConflict,
	service.ErrOrganizationHasUsers:   http.StatusConflict,

	// Subscription errors
	service.ErrSubscriptionNotFound:           http.StatusNotFound,
	service.ErrCannotModifyOwnSubscription:    http.StatusForbidden,
	service.ErrActiveSubscriptionExists:       http.StatusConflict,
	service.ErrActiveOrgSubscriptionExists:    http.StatusConflict,
	service.ErrCannotMarkFreeSubscriptionPaid: http.StatusBadRequest,
}

// MapServiceError maps a service error to an HTTP status code
// Returns the status code and whether the error was found in the map
func MapServiceError(err error) (int, bool) {
	for serviceErr, status := range errorStatusMap {
		if errors.Is(err, serviceErr) {
			return status, true
		}
	}
	return http.StatusInternalServerError, false
}

// WriteError writes an error response with the appropriate status code
// It automatically maps known service errors to HTTP status codes.
// Domain sentinels (ErrProtectedUser, ErrConflict) are checked first and
// produce structured responses with machine-readable error codes.
func WriteError(w http.ResponseWriter, err error) {
	// Check domain sentinels before the generic service error map so that
	// these well-known errors always produce the correct structured body.
	if errors.Is(err, domain.ErrProtectedUser) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:            "protected_user",
			Message:          "This account is system-reserved and cannot be modified.",
			DocumentationURL: "/docs/protected-users",
		})
		return
	}
	if errors.Is(err, domain.ErrConflict) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "conflict",
			Message: "The resource was modified concurrently. Please reload and retry.",
		})
		return
	}
	// Validation errors: services return *domain.InvalidInputError when user
	// input fails a domain rule (empty name, future birthday, malformed email).
	// These are safe to surface verbatim — services construct them with
	// human-readable, field-scoped wording. Without this branch, validation
	// errors fall through to the generic 500 path and the user gets
	// "an internal error occurred" instead of "birthday: must be in the past".
	var invErr *domain.InvalidInputError
	if errors.As(err, &invErr) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_input",
			Message: invErr.Error(),
		})
		return
	}
	// Service-layer duplicate-email sentinel: map to 409 with a stable error
	// code the frontend can act on (e.g., highlight the email input). Placed
	// before MapServiceError so the structured body wins over the generic
	// {message: "..."} mapping that errorStatusMap would otherwise produce.
	if errors.Is(err, service.ErrEmailAlreadyExists) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "duplicate_email",
			Message: "A user with that email already exists.",
		})
		return
	}

	status, known := MapServiceError(err)
	var message string
	if known {
		// Known service error — sentinel message is safe to expose
		message = err.Error()
	} else {
		var httpErr *HTTPError
		if errors.As(err, &httpErr) {
			// HTTPError.Message is a human-written safe string
			status = httpErr.Status
			message = httpErr.Message
		} else {
			// Unknown internal error — do not expose raw Go error detail
			message = "an internal error occurred"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}

// WriteErrorWithStatus writes an error response with a specific status code
func WriteErrorWithStatus(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}

// WriteErrorWithDetail writes an error response with additional detail
func WriteErrorWithDetail(w http.ResponseWriter, status int, message, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message, Error: detail})
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// HandleServiceError handles a service error by writing an appropriate HTTP response
// Returns true if an error was handled, false if err is nil
func HandleServiceError(w http.ResponseWriter, err error, context string) bool {
	if err == nil {
		return false
	}

	status, known := MapServiceError(err)
	if known {
		WriteErrorWithStatus(w, status, err.Error())
	} else {
		// For unknown errors, don't expose internal details
		WriteErrorWithStatus(w, http.StatusInternalServerError, "an internal error occurred")
	}
	return true
}
