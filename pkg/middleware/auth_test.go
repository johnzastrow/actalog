package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/pkg/auth"
)

const testSecret = "test-jwt-secret-for-testing"

func TestAuth_MissingAuthorizationHeader(t *testing.T) {
	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if !contains(rec.Body.String(), "Missing authorization header") {
		t.Errorf("Expected error message about missing header, got: %s", rec.Body.String())
	}
}

func TestAuth_InvalidAuthorizationFormat(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
	}{
		{"no bearer prefix", "token123"},
		{"wrong prefix", "Basic token123"},
		{"too many parts", "Bearer token123 extra"},
		{"bearer lowercase", "bearer token123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.authHeader)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}
		})
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if !contains(rec.Body.String(), "Invalid or expired token") {
		t.Errorf("Expected error about invalid token, got: %s", rec.Body.String())
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	// Generate a token that expires immediately
	token, err := auth.GenerateToken(1, "test@example.com", "user", testSecret, time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for expired token, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	userID := int64(42)
	email := "test@example.com"
	role := "admin"

	token, err := auth.GenerateToken(userID, email, role, testSecret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	var capturedUserID int64
	var capturedEmail string
	var capturedRole string

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		capturedUserID, ok = GetUserID(r.Context())
		if !ok {
			t.Error("UserID not found in context")
		}
		capturedEmail, ok = GetUserEmail(r.Context())
		if !ok {
			t.Error("Email not found in context")
		}
		capturedRole, ok = GetUserRole(r.Context())
		if !ok {
			t.Error("Role not found in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if capturedUserID != userID {
		t.Errorf("Expected userID %d, got %d", userID, capturedUserID)
	}
	if capturedEmail != email {
		t.Errorf("Expected email %s, got %s", email, capturedEmail)
	}
	if capturedRole != role {
		t.Errorf("Expected role %s, got %s", role, capturedRole)
	}
}

func TestAuth_WrongSecret(t *testing.T) {
	// Generate token with one secret
	token, err := auth.GenerateToken(1, "test@example.com", "user", "secret1", time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Try to validate with different secret
	handler := Auth("secret2")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for wrong secret, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		wantID int64
		wantOK bool
	}{
		{
			name:   "valid userID in context",
			ctx:    context.WithValue(context.Background(), UserIDKey, int64(42)),
			wantID: 42,
			wantOK: true,
		},
		{
			name:   "no userID in context",
			ctx:    context.Background(),
			wantID: 0,
			wantOK: false,
		},
		{
			name:   "wrong type in context",
			ctx:    context.WithValue(context.Background(), UserIDKey, "not-an-int"),
			wantID: 0,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := GetUserID(tt.ctx)
			if ok != tt.wantOK {
				t.Errorf("GetUserID() ok = %v, want %v", ok, tt.wantOK)
			}
			if id != tt.wantID {
				t.Errorf("GetUserID() id = %v, want %v", id, tt.wantID)
			}
		})
	}
}

func TestGetUserEmail(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		wantEmail string
		wantOK    bool
	}{
		{
			name:      "valid email in context",
			ctx:       context.WithValue(context.Background(), UserEmailKey, "test@example.com"),
			wantEmail: "test@example.com",
			wantOK:    true,
		},
		{
			name:      "no email in context",
			ctx:       context.Background(),
			wantEmail: "",
			wantOK:    false,
		},
		{
			name:      "wrong type in context",
			ctx:       context.WithValue(context.Background(), UserEmailKey, 12345),
			wantEmail: "",
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, ok := GetUserEmail(tt.ctx)
			if ok != tt.wantOK {
				t.Errorf("GetUserEmail() ok = %v, want %v", ok, tt.wantOK)
			}
			if email != tt.wantEmail {
				t.Errorf("GetUserEmail() email = %v, want %v", email, tt.wantEmail)
			}
		})
	}
}

func TestGetUserRole(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		wantRole string
		wantOK   bool
	}{
		{
			name:     "valid role in context",
			ctx:      context.WithValue(context.Background(), UserRoleKey, "admin"),
			wantRole: "admin",
			wantOK:   true,
		},
		{
			name:     "no role in context",
			ctx:      context.Background(),
			wantRole: "",
			wantOK:   false,
		},
		{
			name:     "wrong type in context",
			ctx:      context.WithValue(context.Background(), UserRoleKey, 123),
			wantRole: "",
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, ok := GetUserRole(tt.ctx)
			if ok != tt.wantOK {
				t.Errorf("GetUserRole() ok = %v, want %v", ok, tt.wantOK)
			}
			if role != tt.wantRole {
				t.Errorf("GetUserRole() role = %v, want %v", role, tt.wantRole)
			}
		})
	}
}

func TestAdminOnly_NoUserContext(t *testing.T) {
	handler := AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/admin", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if !contains(rec.Body.String(), "no user context found") {
		t.Errorf("Expected error about no user context, got: %s", rec.Body.String())
	}
}

func TestAdminOnly_NonAdminUser(t *testing.T) {
	handler := AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/admin", nil)
	// Add user role to context
	ctx := context.WithValue(req.Context(), UserRoleKey, "user")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if !contains(rec.Body.String(), "admin access required") {
		t.Errorf("Expected error about admin access, got: %s", rec.Body.String())
	}
}

func TestAdminOnly_AdminUser(t *testing.T) {
	handlerCalled := false
	handler := AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/admin", nil)
	// Add admin role to context
	ctx := context.WithValue(req.Context(), UserRoleKey, "admin")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !handlerCalled {
		t.Error("Expected handler to be called for admin user")
	}
}

func TestAuth_Integration_WithAdminOnly(t *testing.T) {
	// Generate admin token
	token, err := auth.GenerateToken(1, "admin@example.com", "admin", testSecret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	handlerCalled := false
	handler := Auth(testSecret)(AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
}

func TestAuth_Integration_NonAdminWithAdminOnly(t *testing.T) {
	// Generate non-admin token
	token, err := auth.GenerateToken(1, "user@example.com", "user", testSecret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	handler := Auth(testSecret)(AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

// helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
