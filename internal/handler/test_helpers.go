package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// Test helper functions for handler tests

// createTestRequest creates an HTTP request with optional body
func createTestRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// createAuthenticatedRequest creates a request with user context
func createAuthenticatedRequest(method, path, body string, userID int64, email, role string) *http.Request {
	req := createTestRequest(method, path, body)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.UserEmailKey, email)
	ctx = context.WithValue(ctx, middleware.UserRoleKey, role)
	return req.WithContext(ctx)
}

// parseJSONResponse parses a JSON response body into the target
func parseJSONResponse(rr *httptest.ResponseRecorder, target interface{}) error {
	return json.Unmarshal(rr.Body.Bytes(), target)
}

// assertStatusCode checks if response has expected status code
func assertStatusCode(t interface{ Errorf(string, ...interface{}) }, rr *httptest.ResponseRecorder, expected int) {
	if rr.Code != expected {
		t.Errorf("handler returned wrong status code: got %v want %v, body: %s",
			rr.Code, expected, rr.Body.String())
	}
}

// assertContentType checks if response has expected content type
func assertContentType(t interface{ Errorf(string, ...interface{}) }, rr *httptest.ResponseRecorder, expected string) {
	ct := rr.Header().Get("Content-Type")
	if ct != expected {
		t.Errorf("handler returned wrong content type: got %v want %v", ct, expected)
	}
}

// assertBodyContains checks if response body contains expected string
func assertBodyContains(t interface{ Errorf(string, ...interface{}) }, rr *httptest.ResponseRecorder, expected string) {
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler body does not contain %q: got %s", expected, rr.Body.String())
	}
}

// addChiURLParam adds a chi URL parameter to a request context
func addChiURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}
