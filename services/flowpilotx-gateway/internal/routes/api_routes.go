package routes

import (
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/controllers"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/middleware"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/service"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (r *Router) initAPIRoutes() {
	// Create API controllers
	healthController := controllers.NewHealthController(r.log, r.cfg)
	versionController := controllers.NewVersionController(r.log, r.cfg)

	// Initialize services for workflow
	workflowService := service.NewWorkflowService(
		r.temporalClient,
		r.mongoClient,
		r.log,
		r.cfg,
	)

	// Create workflow controller
	workflowController := controllers.NewWorkflowController(workflowService, r.log)

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

	// Workflow routes
	workflowRouter := api.PathPrefix(r.cfg.FlowpilotxGateway.API.BasePath + "/v1/workflows").Subrouter()
	workflowRouter.Handle("", middleware.Chain(
		workflowController.CreateWorkflow(),
		middleware.LoggingMiddleware(r.log),
	)).Methods("POST")

	workflowRouter.Handle("/{id}", middleware.Chain(
		workflowController.GetWorkflow(),
		middleware.LoggingMiddleware(r.log),
	)).Methods("GET")

	workflowRouter.Handle("/{id}/trigger", middleware.Chain(
		workflowController.TriggerWorkflow(),
		middleware.LoggingMiddleware(r.log),
	)).Methods("POST")

	workflowRouter.Handle("/request/{id}", middleware.Chain(
		workflowController.GetWorkflowRequest(),
		middleware.LoggingMiddleware(r.log),
	)).Methods("GET")

	workflowRouter.Handle("/request/{id}/activities", middleware.Chain(
		workflowController.GetActivityResults(),
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
