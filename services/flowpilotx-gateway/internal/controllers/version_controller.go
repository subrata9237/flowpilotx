package controllers

import (
	"net/http"

	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/service"
)

type VersionController struct {
	log     logger.LoggerInterface
	cfg     *config.Config
	service *service.VersionService
}

func NewVersionController(log logger.LoggerInterface, cfg *config.Config) *VersionController {
	return &VersionController{
		log:     log,
		cfg:     cfg,
		service: service.NewVersionService(log, cfg),
	}
}

// GetVersion godoc
// @Summary      Version information endpoint
// @Description  Get service version information
// @Tags         Version
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.VersionResponse
// @Router       /version [get]
func (c *VersionController) GetVersion() http.HandlerFunc {
	return c.service.GetVersion
}
