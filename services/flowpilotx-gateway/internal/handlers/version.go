package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/models"
)

// NewVersionHandler creates a new version handler with service information
func NewVersionHandler(log logger.LoggerInterface, version, environment, serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := models.VersionResponse{
			Version:     version,
			Environment: environment,
			ServiceName: serviceName,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error(r.Context(), "Failed to encode version response", map[string]interface{}{
				"error":  err.Error(),
				"path":   r.URL.Path,
				"method": r.Method,
			})
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		log.Debug(r.Context(), "Version info served", map[string]interface{}{
			"version":      version,
			"environment": environment,
			"service_name": serviceName,
		})
	}
} 