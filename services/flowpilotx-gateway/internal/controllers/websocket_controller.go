package controllers

import (
	"net/http"

	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/service"
)

type WebSocketController struct {
	log     logger.LoggerInterface
	cfg     *config.Config
	service *service.WebSocketService
}

func NewWebSocketController(log logger.LoggerInterface, cfg *config.Config) *WebSocketController {
	return &WebSocketController{
		log:     log,
		cfg:     cfg,
		service: service.NewWebSocketService(cfg, log),
	}
}

func (c *WebSocketController) HandleWebSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Log incoming WebSocket request
		c.log.InfoWithCtx(r.Context(), "Incoming WebSocket connection", map[string]interface{}{
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
			"path":        r.URL.Path,
		})

		// Handle the WebSocket connection
		c.service.HandleConnection(w, r)
	}
}
