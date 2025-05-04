package controllers

import (
	"github.com/flowpilotx/flowppilotx-tools/httprequest-tool/models"
	"github.com/flowpilotx/flowppilotx-tools/httprequest-tool/services"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// HTTPController handles HTTP endpoints
type HTTPController struct {
	service *services.HTTPService
	logger  *logrus.Logger
}

// NewHTTPController creates a new HTTP controller
func NewHTTPController(service *services.HTTPService, logger *logrus.Logger) *HTTPController {
	return &HTTPController{
		service: service,
		logger:  logger,
	}
}

// HealthCheck handles the health check endpoint
func (c *HTTPController) HealthCheck(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{"status": "healthy"})
}

// ProcessRequest handles the HTTP request processing endpoint
func (c *HTTPController) ProcessRequest(ctx *fiber.Ctx) error {
	var req models.HTTPRequest
	if err := ctx.BodyParser(&req); err != nil {
		c.logger.WithError(err).Error("Invalid request format")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	response, err := c.service.ProcessRequest(&req)
	if err != nil {
		c.logger.WithError(err).Error("Failed to process request")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(response)
}
