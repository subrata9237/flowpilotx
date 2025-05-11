package routes

import (
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/controllers"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (r *Router) initAPIRoutes() {
	// Create API controllers
	healthController := controllers.NewHealthController(r.log, r.cfg)
	versionController := controllers.NewVersionController(r.log, r.cfg)

	// Create API subrouter
	api := r.router.PathPrefix(r.cfg.FlowpilotxGateway.API.BasePath).Subrouter()

	// Health routes
	api.Handle("/health", middleware.Chain(
		healthController.HealthCheck(),
		middleware.LoggingMiddleware(r.log),
	)).Methods("GET")

	// Version routes
	api.Handle("/version", middleware.Chain(
		versionController.GetVersion(),
		middleware.LoggingMiddleware(r.log),
	)).Methods("GET")

	// Swagger documentation
	api.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL(r.cfg.FlowpilotxGateway.API.BasePath+"/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))
}
