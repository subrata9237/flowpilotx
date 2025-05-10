package middleware

import (
	"net/http"

	"github.com/flowpilotx/libs/logger"
)

// Middleware represents a middleware function
type Middleware func(http.Handler) http.Handler

// ApplyMiddleware applies logging middleware to a handler
func ApplyMiddleware(h http.Handler, log logger.LoggerInterface) http.Handler {
	return LoggingMiddleware(log)(h)
} 