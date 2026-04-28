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
