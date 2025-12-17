package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"
	"sync/atomic"
	"time"
)

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares in order (first middleware wraps the handler first).
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	// Apply in reverse order so first middleware is outermost
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// LoggingMiddleware logs HTTP requests and responses.
func LoggingMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a custom response writer to capture status code
			crw := &captureResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Add request ID to context
			requestID := generateRequestID()
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			r = r.WithContext(ctx)

			// Log request
			logger.Info("request started",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)

			next.ServeHTTP(crw, r)

			// Log response
			duration := time.Since(start)
			logger.Info("request completed",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", crw.statusCode,
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}

// RecoveryMiddleware recovers from panics and returns 500 errors.
func RecoveryMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID, _ := r.Context().Value(requestIDKey).(string)

					logger.Error("panic recovered",
						"request_id", requestID,
						"error", err,
						"stack", string(debug.Stack()),
					)

					writeJSONError(w, "Internal server error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware adds CORS headers.
func CORSMiddleware(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			if len(allowedOrigins) == 0 || (len(allowedOrigins) == 1 && allowedOrigins[0] == "*") {
				allowed = true
				origin = "*"
			} else {
				if slices.Contains(allowedOrigins, origin) {
					allowed = true
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
			}

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// BasicAuthMiddleware implements HTTP Basic Authentication.
func BasicAuthMiddleware(password string, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health endpoint
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			// Extract credentials
			auth := r.Header.Get("Authorization")
			if auth == "" {
				requestID, _ := r.Context().Value(requestIDKey).(string)
				logger.Warn("missing authorization header", "request_id", requestID)
				requireAuth(w)
				return
			}

			// Parse Basic auth
			const prefix = "Basic "
			if !strings.HasPrefix(auth, prefix) {
				logger.Warn("invalid authorization format")
				requireAuth(w)
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
			if err != nil {
				logger.Warn("failed to decode authorization", "error", err)
				requireAuth(w)
				return
			}

			// Format is "username:password"
			credentials := strings.SplitN(string(decoded), ":", 2)
			if len(credentials) != 2 {
				logger.Warn("invalid credentials format")
				requireAuth(w)
				return
			}

			// For this simple implementation, we don't care about username
			if credentials[1] != password {
				requestID, _ := r.Context().Value(requestIDKey).(string)
				logger.Warn("authentication failed", "request_id", requestID)
				requireAuth(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// requireAuth sends a 401 response with WWW-Authenticate header.
func requireAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="fave", charset="UTF-8"`)
	writeJSONError(w, "Authentication required", http.StatusUnauthorized)
}

// captureResponseWriter wraps http.ResponseWriter to capture status code.
type captureResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (crw *captureResponseWriter) WriteHeader(code int) {
	crw.statusCode = code
	crw.ResponseWriter.WriteHeader(code)
}

// Context key for request ID
type contextKey string

const requestIDKey contextKey = "request_id"

// Simple request ID generator
var requestCounter uint64

func generateRequestID() string {
	count := atomic.AddUint64(&requestCounter, 1)
	return fmt.Sprintf("req-%d", count)
}
