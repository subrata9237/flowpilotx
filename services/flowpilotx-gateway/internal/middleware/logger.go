package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/models"
)

// LoggingMiddleware creates a new middleware handler for logging HTTP requests
func LoggingMiddleware(log logger.LoggerInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Check if this is a WebSocket upgrade request
			if isWebSocketUpgrade(r) {
				// For WebSocket requests, don't wrap the response writer
				next.ServeHTTP(w, r)
				duration := time.Since(start)
				log.InfoWithCtx(r.Context(), "WebSocket Upgrade Request", map[string]interface{}{
					"method":      r.Method,
					"path":        r.URL.Path,
					"remote_addr": r.RemoteAddr,
					"duration":    duration,
					"user_agent":  r.UserAgent(),
				})
				return
			}

			// Log request start
			log.InfoWithCtx(r.Context(), "Request started", map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
				"query":       r.URL.RawQuery,
			})

			// Create a custom response writer to capture the status code
			rw := &models.ResponseWriter{
				ResponseWriter: w,
				Status:         http.StatusOK,
			}

			// Process request
			next.ServeHTTP(rw, r)

			// Log request completion
			duration := time.Since(start)
			log.InfoWithCtx(r.Context(), "Request completed", map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
				"status":      rw.Status,
				"duration":    duration,
				"user_agent":  r.UserAgent(),
				"query":       r.URL.RawQuery,
			})
		})
	}
}

// isWebSocketUpgrade checks if the request is a WebSocket upgrade request
func isWebSocketUpgrade(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Upgrade")) == "websocket" &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}
