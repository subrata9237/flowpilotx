package main

import (
	"os"

	"github.com/flowpilotx/services/flowpilotx-tools/httprequest-tool/controller"
	_ "github.com/flowpilotx/services/flowpilotx-tools/httprequest-tool/docs" // swagger docs
	"github.com/flowpilotx/services/flowpilotx-tools/httprequest-tool/service"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	log "github.com/sirupsen/logrus"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// @title HTTP Request Tool API
// @version 1.0
// @description A service for handling HTTP requests
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Initialize logger
	logger := log.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&log.JSONFormatter{})

	// Initialize services
	httpService := service.NewHTTPService(logger)

	// Initialize controllers
	httpController := controller.NewHTTPController(httpService, logger)

	// Initialize Echo router
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(middleware.Secure())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
	e.Use(middleware.BodyLimit("100M")) // Limit request body size to 100MB

	// Swagger docs
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// API v1 group
	v1 := e.Group("/api/v1")

	// Health check
	v1.GET("/health", httpController.HealthCheck)

	// HTTP Request endpoints
	v1.POST("/process", httpController.ProcessRequest)
	v1.POST("/process/async", httpController.ProcessAsyncRequest)
	v1.POST("/process/batch", httpController.BatchProcess)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Infof("Server starting on port %s", port)
	if err := e.Start(":" + port); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}
