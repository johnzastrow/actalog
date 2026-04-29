package main

import (
	"net/http"
)

// degradedAdminWriteGuard returns a middleware that short-circuits non-GET/HEAD
// requests with HTTP 503 + a structured body when the binary is running in
// protected-invariant-degraded mode (ACTALOG_SKIP_PROTECTED_INVARIANT=true
// with a failed boot check). The pointer indirection allows the value to be
// updated by the boot path even after the middleware is wired.
func degradedAdminWriteGuard(degraded *bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if *degraded && r.Method != http.MethodGet && r.Method != http.MethodHead {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(`{"error":"protected_invariant_degraded","message":"Admin user-write endpoints are temporarily disabled while the protected-user invariant is broken. Operator: see logs."}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
