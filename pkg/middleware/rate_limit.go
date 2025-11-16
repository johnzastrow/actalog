package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimiter implements in-memory rate limiting with sliding window
type RateLimiter struct {
	requests map[string][]time.Time // IP -> request timestamps
	mu       sync.RWMutex
	limit    int           // Max requests allowed
	window   time.Duration // Time window for rate limiting
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine to prevent memory leaks
	go rl.cleanup()

	return rl
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get existing requests for this IP
	requests := rl.requests[ip]

	// Filter out requests outside the current window
	validRequests := []time.Time{}
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if limit exceeded
	if len(validRequests) >= rl.limit {
		// Update the map (even though we're denying, keep accurate count)
		rl.requests[ip] = validRequests
		return false
	}

	// Add current request and update
	validRequests = append(validRequests, now)
	rl.requests[ip] = validRequests

	return true
}

// cleanup periodically removes old entries to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()

		now := time.Now()
		windowStart := now.Add(-rl.window)

		// Remove IPs with no recent requests
		for ip, requests := range rl.requests {
			// Filter to keep only requests within window
			validRequests := []time.Time{}
			for _, reqTime := range requests {
				if reqTime.After(windowStart) {
					validRequests = append(validRequests, reqTime)
				}
			}

			// If no valid requests, remove IP entirely
			if len(validRequests) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = validRequests
			}
		}

		rl.mu.Unlock()
	}
}

// GetRetryAfter returns how long the client should wait before retrying
func (rl *RateLimiter) GetRetryAfter(ip string) time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	requests := rl.requests[ip]
	if len(requests) == 0 {
		return 0
	}

	// Get the oldest request still in the window
	oldestRequest := requests[0]
	retryTime := oldestRequest.Add(rl.window)
	retryAfter := time.Until(retryTime)

	if retryAfter < 0 {
		return 0
	}

	return retryAfter
}

// RateLimit is a middleware that applies rate limiting to HTTP requests
func RateLimit(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract IP address
			ip := getIP(r)

			// Check if request is allowed
			if !limiter.Allow(ip) {
				// Get retry-after duration
				retryAfter := limiter.GetRetryAfter(ip)

				// Set Retry-After header (in seconds)
				w.Header().Set("Retry-After", formatRetryAfter(retryAfter))
				w.Header().Set("X-RateLimit-Limit", formatInt(limiter.limit))
				w.Header().Set("X-RateLimit-Window", limiter.window.String())

				http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
				return
			}

			// Request allowed, proceed
			next.ServeHTTP(w, r)
		})
	}
}

// getIP extracts the client IP address from the request
func getIP(r *http.Request) string {
	// Try X-Forwarded-For header first (for requests behind proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		return xff
	}

	// Try X-Real-IP header (used by some proxies)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// formatRetryAfter formats duration as seconds for Retry-After header
func formatRetryAfter(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return formatInt(seconds)
}

// formatInt converts int to string
func formatInt(n int) string {
	return strconv.Itoa(n)
}
