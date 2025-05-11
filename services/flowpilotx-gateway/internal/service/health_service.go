package service

import (
	"net/http"

	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/models"
)

type HealthService struct {
	log logger.LoggerInterface
	cfg *config.Config
}

func NewHealthService(log logger.LoggerInterface, cfg *config.Config) *HealthService {
	return &HealthService{
		log: log,
		cfg: cfg,
	}
}

// HealthCheck handles the health check request
func (s *HealthService) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := models.HealthResponse{
		Status:  "UP",
		Version: s.cfg.FlowpilotxGateway.Service.Version,
	}

	models.WriteJSON(w, http.StatusOK, response)
}
