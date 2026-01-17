package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/johnzastrow/actalog/internal/domain"
)

const (
	// SubscriptionStatusKey is the context key for subscription status
	SubscriptionStatusKey ContextKey = "subscriptionStatus"
)

// SubscriptionService is an interface for checking subscription access
// This avoids circular dependency by defining the interface here
type SubscriptionService interface {
	CheckUserAccess(userID int64) (*domain.SubscriptionAccessResult, error)
}

// RequireActiveSubscription enforces read-only mode when subscription is expired
// This middleware should be applied AFTER Auth middleware
// Admin users bypass subscription checks entirely
func RequireActiveSubscription(subscriptionService SubscriptionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user ID from context (set by Auth middleware)
			userID, ok := GetUserID(r.Context())
			if !ok {
				http.Error(w, `{"error":"Unauthorized: no user context found"}`, http.StatusUnauthorized)
				return
			}

			// Admin users bypass subscription checks entirely
			role, _ := GetUserRole(r.Context())
			if role == "admin" {
				next.ServeHTTP(w, r)
				return
			}

			// Check subscription status
			accessResult, err := subscriptionService.CheckUserAccess(userID)
			if err != nil {
				http.Error(w, `{"error":"Failed to check subscription status"}`, http.StatusInternalServerError)
				return
			}

			// Add subscription status to context for handlers to use
			ctx := context.WithValue(r.Context(), SubscriptionStatusKey, accessResult)

			// If user has active subscription, allow all operations
			if accessResult.HasAccess {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// No active subscription - enforce read-only mode
			method := r.Method

			// Allow read operations (GET, HEAD, OPTIONS)
			if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Block write operations (POST, PUT, PATCH, DELETE)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired) // 402 Payment Required

			response := map[string]interface{}{
				"error":   "Subscription required",
				"message": "Your subscription has expired. You can view data but cannot create or modify content.",
				"subscription_status": map[string]interface{}{
					"has_access": accessResult.HasAccess,
					"source":     accessResult.Source,
					"expires_at": accessResult.ExpiresAt,
				},
			}

			json.NewEncoder(w).Encode(response)
		})
	}
}

// GetSubscriptionStatus extracts subscription status from context
func GetSubscriptionStatus(ctx context.Context) (*domain.SubscriptionAccessResult, bool) {
	status, ok := ctx.Value(SubscriptionStatusKey).(*domain.SubscriptionAccessResult)
	return status, ok
}
