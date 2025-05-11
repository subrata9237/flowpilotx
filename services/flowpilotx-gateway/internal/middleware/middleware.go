package middleware

import (
	"net/http"

	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/logger"
	"golang.org/x/time/rate"
)

// Middleware represents a middleware function
type Middleware func(http.Handler) http.Handler

// ApplyMiddleware applies logging middleware to a handler
func ApplyMiddleware(h http.Handler, log logger.LoggerInterface) http.Handler {
	return LoggingMiddleware(log)(h)
}

// Chain chains multiple middleware functions together
func Chain(handler http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

// RateLimit creates a rate limiting middleware
func RateLimit(cfg *config.Config, log logger.LoggerInterface) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(
		rate.Limit(cfg.FlowpilotxGateway.RateLimit.RequestsPerSecond),
		cfg.FlowpilotxGateway.RateLimit.Burst,
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
