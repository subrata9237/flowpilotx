package controllers

import (
	"net/http"

	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/service"
)

type HealthController struct {
	log     logger.LoggerInterface
	cfg     *config.Config
	service *service.HealthService
}

func NewHealthController(log logger.LoggerInterface, cfg *config.Config) *HealthController {
	return &HealthController{
		log:     log,
		cfg:     cfg,
		service: service.NewHealthService(log, cfg),
	}
}

// HealthCheck godoc
// @Summary      Health check endpoint (v1)
// @Description  Get service health status
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.HealthResponse
// @Router       /v1/health [get]
func (c *HealthController) HealthCheck() http.HandlerFunc {
	return c.service.HealthCheck
}
