package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// mockSubscriptionService implements SubscriptionService for testing
type mockSubscriptionService struct {
	accessResult *domain.SubscriptionAccessResult
	err          error
}

func (m *mockSubscriptionService) CheckUserAccess(userID int64) (*domain.SubscriptionAccessResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.accessResult, nil
}

func TestRequireActiveSubscription_NoUserContext(t *testing.T) {
	svc := &mockSubscriptionService{
		accessResult: &domain.SubscriptionAccessResult{HasAccess: true},
	}

	handler := RequireActiveSubscription(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(rec.Body.String(), "no user context found") {
		t.Errorf("Expected error about no user context, got: %s", rec.Body.String())
	}
}

func TestRequireActiveSubscription_ServiceError(t *testing.T) {
	svc := &mockSubscriptionService{
		err: context.DeadlineExceeded, // Some error
	}

	handler := RequireActiveSubscription(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	// Add user context
	ctx := context.WithValue(req.Context(), UserIDKey, int64(1))
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestRequireActiveSubscription_ActiveSubscription(t *testing.T) {
	svc := &mockSubscriptionService{
		accessResult: &domain.SubscriptionAccessResult{
			HasAccess: true,
			Source:    "user",
		},
	}

	handlerCalled := false
	handler := RequireActiveSubscription(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// Verify subscription status is in context
		status, ok := GetSubscriptionStatus(r.Context())
		if !ok {
			t.Error("Subscription status not found in context")
		}
		if !status.HasAccess {
			t.Error("Expected HasAccess to be true")
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Test all HTTP methods with active subscription
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			handlerCalled = false
			req := httptest.NewRequest(method, "/test", nil)
			ctx := context.WithValue(req.Context(), UserIDKey, int64(1))
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("%s status = %d, want %d", method, rec.Code, http.StatusOK)
			}
			if !handlerCalled {
				t.Errorf("%s: handler should be called with active subscription", method)
			}
		})
	}
}

func TestRequireActiveSubscription_ExpiredSubscription_ReadOnly(t *testing.T) {
	expiresAt := time.Now().Add(-24 * time.Hour)
	svc := &mockSubscriptionService{
		accessResult: &domain.SubscriptionAccessResult{
			HasAccess: false,
			Source:    "expired",
			ExpiresAt: &expiresAt,
		},
	}

	handlerCalled := false
	handler := RequireActiveSubscription(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	// Read operations should be allowed
	readMethods := []string{"GET", "HEAD", "OPTIONS"}
	for _, method := range readMethods {
		t.Run("allowed_"+method, func(t *testing.T) {
			handlerCalled = false
			req := httptest.NewRequest(method, "/test", nil)
			ctx := context.WithValue(req.Context(), UserIDKey, int64(1))
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("%s status = %d, want %d", method, rec.Code, http.StatusOK)
			}
			if !handlerCalled {
				t.Errorf("%s: handler should be called for read operation", method)
			}
		})
	}

	// Write operations should be blocked
	writeMethods := []string{"POST", "PUT", "PATCH", "DELETE"}
	for _, method := range writeMethods {
		t.Run("blocked_"+method, func(t *testing.T) {
			handlerCalled = false
			req := httptest.NewRequest(method, "/test", nil)
			ctx := context.WithValue(req.Context(), UserIDKey, int64(1))
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusPaymentRequired {
				t.Errorf("%s status = %d, want %d", method, rec.Code, http.StatusPaymentRequired)
			}
			if handlerCalled {
				t.Errorf("%s: handler should NOT be called for write operation with expired subscription", method)
			}
			if !strings.Contains(rec.Body.String(), "Subscription required") {
				t.Errorf("Expected subscription required error, got: %s", rec.Body.String())
			}
		})
	}
}

func TestGetSubscriptionStatus(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		wantOK bool
	}{
		{
			name: "valid status in context",
			ctx: context.WithValue(context.Background(), SubscriptionStatusKey, &domain.SubscriptionAccessResult{
				HasAccess: true,
				Source:    "user",
			}),
			wantOK: true,
		},
		{
			name:   "no status in context",
			ctx:    context.Background(),
			wantOK: false,
		},
		{
			name:   "wrong type in context",
			ctx:    context.WithValue(context.Background(), SubscriptionStatusKey, "not-a-subscription-result"),
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, ok := GetSubscriptionStatus(tt.ctx)
			if ok != tt.wantOK {
				t.Errorf("GetSubscriptionStatus() ok = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantOK && status == nil {
				t.Error("Expected non-nil status when ok is true")
			}
		})
	}
}

func TestRequireActiveSubscription_SubscriptionStatusInContext(t *testing.T) {
	expiresAt := time.Now().Add(24 * time.Hour)
	svc := &mockSubscriptionService{
		accessResult: &domain.SubscriptionAccessResult{
			HasAccess: true,
			Source:    "organization",
			ExpiresAt: &expiresAt,
		},
	}

	handler := RequireActiveSubscription(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status, ok := GetSubscriptionStatus(r.Context())
		if !ok {
			t.Error("Subscription status not in context")
			return
		}
		if status.Source != "organization" {
			t.Errorf("Source = %s, want organization", status.Source)
		}
		if !status.HasAccess {
			t.Error("HasAccess should be true")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, int64(1))
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRequireActiveSubscription_AdminBypass(t *testing.T) {
	// Admin users should bypass subscription checks entirely
	// Even with expired/no subscription, admin write operations should succeed
	expiresAt := time.Now().Add(-24 * time.Hour)
	svc := &mockSubscriptionService{
		accessResult: &domain.SubscriptionAccessResult{
			HasAccess: false, // No access - but admin should bypass
			Source:    "expired",
			ExpiresAt: &expiresAt,
		},
	}

	handlerCalled := false
	handler := RequireActiveSubscription(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	// All methods should be allowed for admin users, even write operations
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	for _, method := range methods {
		t.Run("admin_"+method, func(t *testing.T) {
			handlerCalled = false
			req := httptest.NewRequest(method, "/test", nil)
			// Add both user ID and admin role to context
			ctx := context.WithValue(req.Context(), UserIDKey, int64(1))
			ctx = context.WithValue(ctx, UserRoleKey, "admin")
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Admin %s status = %d, want %d", method, rec.Code, http.StatusOK)
			}
			if !handlerCalled {
				t.Errorf("Admin %s: handler should be called (admin bypasses subscription)", method)
			}
		})
	}
}

func TestRequireActiveSubscription_CoachBypass(t *testing.T) {
	// Coach users should bypass subscription checks entirely, just like admins
	expiresAt := time.Now().Add(-24 * time.Hour)
	svc := &mockSubscriptionService{
		accessResult: &domain.SubscriptionAccessResult{
			HasAccess: false,
			Source:    "expired",
			ExpiresAt: &expiresAt,
		},
	}

	handlerCalled := false
	handler := RequireActiveSubscription(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	// All methods should be allowed for coach users, even write operations
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	for _, method := range methods {
		t.Run("coach_"+method, func(t *testing.T) {
			handlerCalled = false
			req := httptest.NewRequest(method, "/test", nil)
			ctx := context.WithValue(req.Context(), UserIDKey, int64(1))
			ctx = context.WithValue(ctx, UserRoleKey, "coach")
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Coach %s status = %d, want %d", method, rec.Code, http.StatusOK)
			}
			if !handlerCalled {
				t.Errorf("Coach %s: handler should be called (coach bypasses subscription)", method)
			}
		})
	}
}

func TestRequireActiveSubscription_AthleteStillRequiresSubscription(t *testing.T) {
	// Athlete users should NOT bypass subscription checks
	expiresAt := time.Now().Add(-24 * time.Hour)
	svc := &mockSubscriptionService{
		accessResult: &domain.SubscriptionAccessResult{
			HasAccess: false,
			Source:    "expired",
			ExpiresAt: &expiresAt,
		},
	}

	handlerCalled := false
	handler := RequireActiveSubscription(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	// Write operations should be blocked for athletes with expired subscriptions
	t.Run("athlete_POST", func(t *testing.T) {
		handlerCalled = false
		req := httptest.NewRequest("POST", "/test", nil)
		ctx := context.WithValue(req.Context(), UserIDKey, int64(1))
		ctx = context.WithValue(ctx, UserRoleKey, "athlete")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusPaymentRequired {
			t.Errorf("Athlete POST status = %d, want %d", rec.Code, http.StatusPaymentRequired)
		}
		if handlerCalled {
			t.Error("Athlete: handler should NOT be called for write operation with expired subscription")
		}
	})

	// Read operations should still be allowed
	t.Run("athlete_GET", func(t *testing.T) {
		handlerCalled = false
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := context.WithValue(req.Context(), UserIDKey, int64(1))
		ctx = context.WithValue(ctx, UserRoleKey, "athlete")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Athlete GET status = %d, want %d", rec.Code, http.StatusOK)
		}
		if !handlerCalled {
			t.Error("Athlete: handler should be called for read operation")
		}
	})
}

func TestRequireActiveSubscription_NonAdminRoleStillRequiresSubscription(t *testing.T) {
	// Non-admin users (role="user" or empty) should still require subscription
	expiresAt := time.Now().Add(-24 * time.Hour)
	svc := &mockSubscriptionService{
		accessResult: &domain.SubscriptionAccessResult{
			HasAccess: false,
			Source:    "expired",
			ExpiresAt: &expiresAt,
		},
	}

	handlerCalled := false
	handler := RequireActiveSubscription(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	// Test with role="user" - should NOT bypass
	roles := []string{"user", "", "member", "subscriber"}
	for _, role := range roles {
		t.Run("role_"+role+"_POST", func(t *testing.T) {
			handlerCalled = false
			req := httptest.NewRequest("POST", "/test", nil)
			ctx := context.WithValue(req.Context(), UserIDKey, int64(1))
			if role != "" {
				ctx = context.WithValue(ctx, UserRoleKey, role)
			}
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			// Write operations should be blocked for non-admin expired subscriptions
			if rec.Code != http.StatusPaymentRequired {
				t.Errorf("Non-admin (role=%q) POST status = %d, want %d", role, rec.Code, http.StatusPaymentRequired)
			}
			if handlerCalled {
				t.Errorf("Non-admin (role=%q): handler should NOT be called for write operation", role)
			}
		})

		// Read operations should still be allowed
		t.Run("role_"+role+"_GET", func(t *testing.T) {
			handlerCalled = false
			req := httptest.NewRequest("GET", "/test", nil)
			ctx := context.WithValue(req.Context(), UserIDKey, int64(1))
			if role != "" {
				ctx = context.WithValue(ctx, UserRoleKey, role)
			}
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Non-admin (role=%q) GET status = %d, want %d", role, rec.Code, http.StatusOK)
			}
			if !handlerCalled {
				t.Errorf("Non-admin (role=%q): handler should be called for read operation", role)
			}
		})
	}
}
