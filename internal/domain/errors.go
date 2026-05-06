// Package domain defines the core domain entities, interfaces, and error
// sentinels for the actalog application.
package domain

import "errors"

// ErrProtectedUser is returned when an operation targets a system-reserved
// user account that must not be modified. Handlers should map this to HTTP 403.
var ErrProtectedUser = errors.New("protected user: modifications blocked")

// ErrConflict signals an optimistic-concurrency failure: the caller's
// updated_at did not match the row's current updated_at. Returned from
// admin update services; mapped to HTTP 409 by the handler error pipeline.
//
// Note: this is the domain-layer sentinel. See handler.ErrConflict for
// the HTTP-layer equivalent (a *HTTPError used by the older MapServiceError pipeline).
var ErrConflict = errors.New("resource modified concurrently")

// InvalidInputError is the validation-failure type returned by service-layer
// methods when user-supplied input violates a domain rule (e.g., empty name,
// future birthday, malformed email). Handlers should surface it as HTTP 400
// with the embedded message — it is safe to expose because services construct
// it with human-readable, field-scoped wording.
//
// Use a typed value (not a sentinel) so the handler can extract the exact
// message via errors.As without sniffing strings, and so Field can drive
// future per-field error UIs without changing the wire format.
//
// Example:
//
//	if name == "" {
//	    return nil, &domain.InvalidInputError{Field: "name", Message: "is required"}
//	}
//	// → response body: {"message": "name: is required", "error": "invalid_input"}
type InvalidInputError struct {
	// Field is the name of the offending input (e.g., "birthday"). Optional —
	// when empty, Error() returns Message alone.
	Field string
	// Message is the human-readable explanation. Required.
	Message string
	// Cause is the underlying error, if any (e.g., a parse failure). Optional.
	Cause error
}

func (e *InvalidInputError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}

func (e *InvalidInputError) Unwrap() error { return e.Cause }
