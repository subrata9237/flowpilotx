package controller

import (
	"encoding/base64"
	"net/http"

	"github.com/flowpilotx/services/flowpilotx-tools/httprequest-tool/model"
	"github.com/flowpilotx/services/flowpilotx-tools/httprequest-tool/service"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// HTTPController handles HTTP endpoints
type HTTPController struct {
	httpService *service.HTTPService
	logger      *log.Logger
}

// NewHTTPController creates a new HTTP controller
func NewHTTPController(httpService *service.HTTPService, logger *log.Logger) *HTTPController {
	return &HTTPController{
		httpService: httpService,
		logger:      logger,
	}
}

// HealthCheck godoc
// @Summary Check service health
// @Description Get the health status of the service
// @Tags health
// @Accept application/json
// @Produce application/json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (c *HTTPController) HealthCheck(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// ProcessRequest godoc
// @Summary Process HTTP request
// @Description Process an HTTP request with the given parameters
// @Tags request
// @Accept application/json
// @Produce application/json
// @Produce application/xml
// @Produce text/plain
// @Produce application/octet-stream
// @Param request body model.HTTPRequest true "HTTP request parameters"
// @Success 200 {object} model.HTTPResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /process [post]
func (c *HTTPController) ProcessRequest(ctx echo.Context) error {
	req := new(model.HTTPRequest)
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format: " + err.Error(),
		})
	}

	response, err := c.httpService.ProcessRequest(req)
	if err != nil {
		c.logger.WithError(err).Error("Failed to process request")
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process request: " + err.Error(),
		})
	}

	return c.sendResponse(ctx, response)
}

// ProcessAsyncRequest godoc
// @Summary Process HTTP request asynchronously
// @Description Process an HTTP request asynchronously and return a request ID
// @Tags request
// @Accept application/json
// @Produce application/json
// @Param request body model.HTTPRequest true "HTTP request parameters"
// @Success 202 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /process/async [post]
func (c *HTTPController) ProcessAsyncRequest(ctx echo.Context) error {
	req := new(model.HTTPRequest)
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format: " + err.Error(),
		})
	}

	requestID, err := c.httpService.ProcessAsyncRequest(req)
	if err != nil {
		c.logger.WithError(err).Error("Failed to process async request")
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process async request: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusAccepted, map[string]string{
		"request_id": requestID,
		"status":     "processing",
	})
}

// BatchProcess godoc
// @Summary Process multiple HTTP requests
// @Description Process multiple HTTP requests in batch
// @Tags request
// @Accept application/json
// @Produce application/json
// @Produce application/xml
// @Produce text/plain
// @Produce application/octet-stream
// @Param requests body []model.HTTPRequest true "Array of HTTP request parameters"
// @Success 200 {array} model.HTTPResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /process/batch [post]
func (c *HTTPController) BatchProcess(ctx echo.Context) error {
	var requests []model.HTTPRequest
	if err := ctx.Bind(&requests); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format: " + err.Error(),
		})
	}

	if len(requests) == 0 {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Batch request cannot be empty",
		})
	}

	responses, err := c.httpService.BatchProcess(requests)
	if err != nil {
		c.logger.WithError(err).Error("Failed to process batch request")
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process batch request: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, responses)
}

// sendResponse handles sending the response with the appropriate content type
func (c *HTTPController) sendResponse(ctx echo.Context, response *model.HTTPResponse) error {
	// Set response headers
	for key, value := range response.Headers {
		ctx.Response().Header().Set(key, value)
	}

	// Set content type
	ctx.Response().Header().Set("Content-Type", response.ContentType)

	// Handle different response formats
	switch response.Format {
	case "json":
		return ctx.JSON(response.StatusCode, response.Body)
	case "xml":
		return ctx.XML(response.StatusCode, response.Body)
	case "text":
		return ctx.String(response.StatusCode, response.Body.(string))
	case "binary", "image", "audio", "video", "pdf", "zip", "executable":
		if response.IsBase64 {
			data, err := base64.StdEncoding.DecodeString(response.Body.(string))
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to decode base64 response",
				})
			}
			return ctx.Blob(response.StatusCode, response.ContentType, data)
		}
		return ctx.Blob(response.StatusCode, response.ContentType, response.RawBody)
	default:
		// Default to JSON if format is unknown
		return ctx.JSON(response.StatusCode, response)
	}
}
