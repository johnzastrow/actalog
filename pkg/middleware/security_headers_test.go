package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecurityHeaders_SetsBaselineHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	cases := []struct {
		header string
		want   string
	}{
		{"X-Frame-Options", "DENY"},
		{"X-Content-Type-Options", "nosniff"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
		{"Strict-Transport-Security", "max-age=31536000; includeSubDomains"},
	}
	for _, tc := range cases {
		if got := rec.Header().Get(tc.header); got != tc.want {
			t.Errorf("%s = %q, want %q", tc.header, got, tc.want)
		}
	}
}

func TestSecurityHeaders_CSPIncludesRequiredDirectives(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatal("Content-Security-Policy header is empty")
	}

	required := []string{
		"default-src 'self'",
		"script-src 'self'", // strict — no 'unsafe-inline' or 'unsafe-eval'
		"frame-ancestors 'none'",
		"base-uri 'self'",
		"form-action 'self'",
	}
	for _, directive := range required {
		if !strings.Contains(csp, directive) {
			t.Errorf("CSP missing required directive %q\nfull CSP: %s", directive, csp)
		}
	}

	// Script source must NOT carry unsafe-inline or unsafe-eval. Check the
	// script-src directive specifically; 'unsafe-inline' is allowed for styles.
	scriptDirective := extractDirective(csp, "script-src")
	for _, forbidden := range []string{"'unsafe-inline'", "'unsafe-eval'"} {
		if strings.Contains(scriptDirective, forbidden) {
			t.Errorf("script-src must not contain %s; got %q", forbidden, scriptDirective)
		}
	}
}

func TestSecurityHeaders_PassesThroughHandlerResponse(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("hello"))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusTeapot)
	}
	if rec.Body.String() != "hello" {
		t.Errorf("Body = %q, want %q", rec.Body.String(), "hello")
	}
}

// extractDirective returns the value portion of a single CSP directive, or "" if
// the directive is not present. Used so tests can assert per-directive content
// rather than substring-matching the entire CSP string (which would falsely
// match 'unsafe-inline' on style-src when checking script-src).
func extractDirective(csp, name string) string {
	for _, part := range strings.Split(csp, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, name+" ") || part == name {
			return strings.TrimPrefix(part, name)
		}
	}
	return ""
}
