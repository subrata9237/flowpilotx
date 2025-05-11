package routes

import (
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/controllers"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/middleware"
)

func (r *Router) initWebSocketRoutes() {
	wsController := controllers.NewWebSocketController(r.log, r.cfg)

	// WebSocket route with logging and rate limiting middleware
	r.router.Handle(r.cfg.FlowpilotxGateway.WebSocket.Path,
		middleware.Chain(
			wsController.HandleWebSocket(),
			middleware.LoggingMiddleware(r.log),
			middleware.RateLimit(r.cfg, r.log),
		),
	)
}
