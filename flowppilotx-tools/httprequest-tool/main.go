package main

import (
	"os"

	"github.com/flowpilotx/flowppilotx-tools/httprequest-tool/controllers"
	"github.com/flowpilotx/flowppilotx-tools/httprequest-tool/services"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Initialize services
	httpService := services.NewHTTPService(logger)

	// Initialize controllers
	httpController := controllers.NewHTTPController(httpService, logger)

	// Initialize Fiber router
	router := fiber.New()

	// Register routes
	router.Get("/health", httpController.HealthCheck)
	router.Post("/process", httpController.ProcessRequest)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.Infof("Server starting on port %s", port)
	if err := router.Listen(":" + port); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}
