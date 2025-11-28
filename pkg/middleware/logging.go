package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/pkg/logger"
)

// loggingResponseWriter wraps http.ResponseWriter to capture status code and response size
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
	body   *bytes.Buffer
}

func (rw *loggingResponseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	// Capture response body for error responses
	if rw.status >= 400 && rw.body != nil {
		rw.body.Write(b)
	}
	return size, err
}

// LoggingMiddleware creates a middleware that logs HTTP requests with detailed information
func LoggingMiddleware(log *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Capture request body for non-GET requests (for debugging)
			var requestBody string
			if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Body != nil {
				bodyBytes, err := io.ReadAll(r.Body)
				if err == nil {
					// Restore body for handlers
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

					// Parse and sanitize request body (hide passwords)
					var bodyMap map[string]interface{}
					if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
						// Redact sensitive fields
						if _, ok := bodyMap["password"]; ok {
							bodyMap["password"] = "[REDACTED]"
						}
						if _, ok := bodyMap["token"]; ok {
							bodyMap["token"] = "[REDACTED]"
						}
						sanitized, _ := json.Marshal(bodyMap)
						requestBody = string(sanitized)
					} else {
						// If not JSON, just truncate long bodies
						if len(bodyBytes) > 200 {
							requestBody = string(bodyBytes[:200]) + "... [truncated]"
						} else {
							requestBody = string(bodyBytes)
						}
					}
				}
			}

			// Wrap response writer to capture status and body
			wrapped := &loggingResponseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
				body:           &bytes.Buffer{},
			}

			// Get route pattern from chi router
			routePattern := "-"
			rctx := chi.RouteContext(r.Context())
			if rctx != nil && rctx.RoutePattern() != "" {
				routePattern = rctx.RoutePattern()
			}

			// Extract user info from context if available
			userID := "-"
			userEmail := "-"
			if uid, ok := r.Context().Value("user_id").(int64); ok {
				userID = strconv.FormatInt(uid, 10)
			}
			if email, ok := r.Context().Value("user_email").(string); ok {
				userEmail = email
				_ = userEmail // Will use later for detailed logs
			}

			// Log request start
			log.Debug("=> %s %s user=%s route=%s remote=%s ua=%s",
				r.Method, r.URL.Path, userID, routePattern, r.RemoteAddr, r.UserAgent())

			if requestBody != "" {
				log.Debug("   body=%s", requestBody)
			}

			if len(r.URL.RawQuery) > 0 {
				log.Debug("   query=%s", r.URL.RawQuery)
			}

			// Call the next handler
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Determine log level based on status code
			statusClass := wrapped.status / 100

			// Build base log message
			logMsg := "%s %s status=%d duration=%v size=%d bytes user=%s route=%s"
			logArgs := []interface{}{
				r.Method,
				r.URL.Path,
				wrapped.status,
				duration,
				wrapped.size,
				userID,
				routePattern,
			}

			// Add query params if present
			if len(r.URL.RawQuery) > 0 {
				logMsg += " query=%s"
				logArgs = append(logArgs, r.URL.RawQuery)
			}

			// Add error response body for 4xx/5xx
			if wrapped.status >= 400 && wrapped.body.Len() > 0 {
				errorBody := wrapped.body.String()
				// Truncate long error messages
				if len(errorBody) > 500 {
					errorBody = errorBody[:500] + "... [truncated]"
				}
				logMsg += " error_response=%s"
				logArgs = append(logArgs, strings.TrimSpace(errorBody))
			}

			// Log based on status code
			switch {
			case statusClass == 5:
				log.Error(logMsg, logArgs...)
			case statusClass == 4:
				log.Warn(logMsg, logArgs...)
			case statusClass == 3:
				log.Info(logMsg, logArgs...)
			default:
				log.Info(logMsg, logArgs...)
			}

			// Log slow requests (> 1 second)
			if duration > time.Second {
				log.Warn("SLOW REQUEST: %s %s took %v", r.Method, r.URL.Path, duration)
			}
		})
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(log *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Add to response headers
			w.Header().Set("X-Request-ID", requestID)

			// Log request ID for correlation
			log.Debug("request_id=%s %s %s", requestID, r.Method, r.URL.Path)

			next.ServeHTTP(w, r)
		})
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405.000000")
}
