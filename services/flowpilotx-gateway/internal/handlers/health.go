package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/models"
)

// HealthHandler creates a new health check handler with logging
func NewHealthHandler(log logger.LoggerInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := models.HealthResponse{
			Status:    "ok",
			Timestamp: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error(r.Context(), "Failed to encode health response", map[string]interface{}{
				"path":   r.URL.Path,
				"method": r.Method,
				"error":  err.Error(),
			})
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		log.Debug(r.Context(), "Health check completed", map[string]interface{}{
			"status":    response.Status,
			"timestamp": response.Timestamp,
		})
	}
} 