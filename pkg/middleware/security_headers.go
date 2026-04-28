package middleware

import "net/http"

// defaultCSP is a starting Content-Security-Policy that does not break the Vue 3
// + Vuetify 3 PWA. `'unsafe-inline'` is needed only for styles because Vuetify
// injects inline <style> blocks at runtime for theme switching; scripts stay
// strict. Tighten this further as nonces are introduced.
const defaultCSP = "default-src 'self'; " +
	"script-src 'self'; " +
	"style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data: blob:; " +
	"connect-src 'self'; " +
	"font-src 'self'; " +
	"frame-ancestors 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'"

// SecurityHeaders sets a baseline of OWASP-recommended security response
// headers on every response. Values are conservative defaults; CSP in particular
// is project-specific and should be re-evaluated as the frontend evolves.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		h.Set("Content-Security-Policy", defaultCSP)

		next.ServeHTTP(w, r)
	})
}
