package routes

import (
	"fmt"
	"net/http"

	"github.com/flowpilotx/libs/worker"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/controllers"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (r *Router) initAPIRoutes() {
	// Create API subrouter with base path
	api := r.router.PathPrefix(r.cfg.FlowpilotxGateway.API.BasePath).Subrouter()

	// Initialize controllers
	healthController := controllers.NewHealthController(r.log, r.cfg)
	versionController := controllers.NewVersionController(r.log, r.cfg)

	// Initialize services for workflow
	workflowService := worker.NewWorkflowService(
		r.temporalClient,
		r.mongoClient,
		r.log,
		r.cfg,
	)
	workflowController := controllers.NewWorkflowController(workflowService, r.log)

	// Health and Version routes
	api.Handle("/health", middleware.Chain(
		healthController.HealthCheck(),
		middleware.LoggingMiddleware(r.log),
	)).Methods("GET")

	api.Handle("/version", middleware.Chain(
		versionController.GetVersion(),
		middleware.LoggingMiddleware(r.log),
	)).Methods("GET")

	// Workflow routes (v1)
	v1 := api.PathPrefix("/v1").Subrouter()
	workflowRouter := v1.PathPrefix("/workflows").Subrouter()

	// Workflow route definitions
	workflowRoutes := []struct {
		path    string
		handler http.Handler
		method  string
	}{
		{"", workflowController.CreateWorkflow(), "POST"},
		{"/{id}", workflowController.GetWorkflow(), "GET"},
		{"/{id}/trigger", workflowController.TriggerWorkflow(), "POST"},
		{"/request/{id}", workflowController.GetWorkflowRequest(), "GET"},
		{"/request/{id}/activities", workflowController.GetActivityResults(), "GET"},
	}

	// Register workflow routes
	for _, route := range workflowRoutes {
		workflowRouter.Handle(route.path, middleware.Chain(
			route.handler,
			middleware.LoggingMiddleware(r.log),
		)).Methods(route.method)
	}

	// Swagger documentation
	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("%s/swagger/doc.json", r.cfg.FlowpilotxGateway.API.BasePath)),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)
	api.PathPrefix("/swagger/").Handler(swaggerHandler)
}
