package routes

import (
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/middleware"
	"github.com/rs/cors"
)

func (r *Router) initMiddleware() {
	// Add CORS middleware
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   r.cfg.FlowpilotxGateway.CORS.AllowedOrigins,
		AllowedMethods:   append(r.cfg.FlowpilotxGateway.CORS.AllowedMethods, "GET", "OPTIONS"),
		AllowedHeaders:   append(r.cfg.FlowpilotxGateway.CORS.AllowedHeaders, "Sec-WebSocket-Protocol", "Sec-WebSocket-Version", "Sec-WebSocket-Key", "Sec-WebSocket-Extensions", "Upgrade", "Connection"),
		MaxAge:           r.cfg.FlowpilotxGateway.CORS.MaxAge,
		AllowCredentials: true,
		ExposedHeaders:   []string{"Sec-WebSocket-Accept"},
	})
	r.router.Use(corsHandler.Handler)

	// Add rate limiter if enabled
	if r.cfg.FlowpilotxGateway.RateLimit.Enabled {
		r.router.Use(middleware.RateLimit(r.cfg, r.log))
	}
}
