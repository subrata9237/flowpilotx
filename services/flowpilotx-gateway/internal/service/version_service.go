package service

import (
	"encoding/json"
	"net/http"

	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/models"
)

type VersionService struct {
	log logger.LoggerInterface
	cfg *config.Config
}

func NewVersionService(log logger.LoggerInterface, cfg *config.Config) *VersionService {
	return &VersionService{
		log: log,
		cfg: cfg,
	}
}

// GetVersion handles the version information request
func (s *VersionService) GetVersion(w http.ResponseWriter, r *http.Request) {
	response := models.VersionResponse{
		Version:     s.cfg.FlowpilotxGateway.Service.Version,
		Environment: s.cfg.FlowpilotxGateway.Service.Environment,
		ServiceName: s.cfg.FlowpilotxGateway.Service.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.log.ErrorWithCtx(r.Context(), "Failed to encode version response", map[string]interface{}{
			"error":  err.Error(),
			"path":   r.URL.Path,
			"method": r.Method,
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.log.DebugWithCtx(r.Context(), "Version info served", map[string]interface{}{
		"version":      response.Version,
		"environment":  response.Environment,
		"service_name": response.ServiceName,
	})
}
