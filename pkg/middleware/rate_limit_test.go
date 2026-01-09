package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(10, time.Minute)

	if limiter == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
	if limiter.limit != 10 {
		t.Errorf("limit = %d, want 10", limiter.limit)
	}
	if limiter.window != time.Minute {
		t.Errorf("window = %v, want %v", limiter.window, time.Minute)
	}
	if limiter.requests == nil {
		t.Error("requests map should be initialized")
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(3, time.Second)
	ip := "192.168.1.1"

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !limiter.Allow(ip) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if limiter.Allow(ip) {
		t.Error("Request 4 should be denied (rate limit exceeded)")
	}
}

func TestRateLimiter_AllowDifferentIPs(t *testing.T) {
	limiter := NewRateLimiter(2, time.Second)

	// IP1 should have its own limit
	if !limiter.Allow("192.168.1.1") {
		t.Error("First request from IP1 should be allowed")
	}
	if !limiter.Allow("192.168.1.1") {
		t.Error("Second request from IP1 should be allowed")
	}
	if limiter.Allow("192.168.1.1") {
		t.Error("Third request from IP1 should be denied")
	}

	// IP2 should have its own separate limit
	if !limiter.Allow("192.168.1.2") {
		t.Error("First request from IP2 should be allowed")
	}
	if !limiter.Allow("192.168.1.2") {
		t.Error("Second request from IP2 should be allowed")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	// Use a very short window for testing
	limiter := NewRateLimiter(2, 50*time.Millisecond)
	ip := "192.168.1.1"

	// Use up the limit
	limiter.Allow(ip)
	limiter.Allow(ip)

	// Should be denied
	if limiter.Allow(ip) {
		t.Error("Should be denied after limit reached")
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// Should be allowed again
	if !limiter.Allow(ip) {
		t.Error("Should be allowed after window expired")
	}
}

func TestRateLimiter_GetRetryAfter(t *testing.T) {
	limiter := NewRateLimiter(1, time.Second)
	ip := "192.168.1.1"

	// No requests yet - retry after should be 0
	if retryAfter := limiter.GetRetryAfter(ip); retryAfter != 0 {
		t.Errorf("GetRetryAfter before requests = %v, want 0", retryAfter)
	}

	// Make a request
	limiter.Allow(ip)

	// Get retry after - should be around 1 second
	retryAfter := limiter.GetRetryAfter(ip)
	if retryAfter < 900*time.Millisecond || retryAfter > 1100*time.Millisecond {
		t.Errorf("GetRetryAfter = %v, want ~1s", retryAfter)
	}
}

func TestRateLimit_Middleware(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)

	handlerCalled := 0
	handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled++
		w.WriteHeader(http.StatusOK)
	}))

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d status = %d, want %d", i+1, rec.Code, http.StatusOK)
		}
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Request 3 status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}

	// Check rate limit headers
	if rec.Header().Get("Retry-After") == "" {
		t.Error("Retry-After header should be set")
	}
	if rec.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("X-RateLimit-Limit header should be set")
	}
	if rec.Header().Get("X-RateLimit-Window") == "" {
		t.Error("X-RateLimit-Window header should be set")
	}

	// Handler should have been called only twice
	if handlerCalled != 2 {
		t.Errorf("Handler called %d times, want 2", handlerCalled)
	}
}

func TestGetIP(t *testing.T) {
	tests := []struct {
		name          string
		xForwardedFor string
		xRealIP       string
		remoteAddr    string
		expectedIP    string
	}{
		{
			name:       "use RemoteAddr when no headers",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1:12345",
		},
		{
			name:          "prefer X-Forwarded-For",
			xForwardedFor: "10.0.0.1",
			xRealIP:       "172.16.0.1",
			remoteAddr:    "192.168.1.1:12345",
			expectedIP:    "10.0.0.1",
		},
		{
			name:       "use X-Real-IP when no X-Forwarded-For",
			xRealIP:    "172.16.0.1",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "172.16.0.1",
		},
		{
			name:          "X-Forwarded-For with multiple IPs",
			xForwardedFor: "10.0.0.1, 10.0.0.2, 10.0.0.3",
			remoteAddr:    "192.168.1.1:12345",
			expectedIP:    "10.0.0.1, 10.0.0.2, 10.0.0.3", // Returns full header
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := getIP(req)
			if ip != tt.expectedIP {
				t.Errorf("getIP() = %q, want %q", ip, tt.expectedIP)
			}
		})
	}
}

func TestFormatRetryAfter(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{0, "1"},                      // Minimum 1 second
		{500 * time.Millisecond, "1"}, // Less than 1 second rounds to 1
		{1 * time.Second, "1"},
		{5 * time.Second, "5"},
		{60 * time.Second, "60"},
	}

	for _, tt := range tests {
		t.Run(tt.duration.String(), func(t *testing.T) {
			result := formatRetryAfter(tt.duration)
			if result != tt.expected {
				t.Errorf("formatRetryAfter(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

// Removed: TestFormatInt - tested strconv.Itoa wrapper, trivial helper
