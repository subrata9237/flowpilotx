package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/flowpilotx/flowppilotx-tools/httprequest-tool/models"
	"github.com/sirupsen/logrus"
)

// HTTPService handles HTTP request processing
type HTTPService struct {
	logger *logrus.Logger
}

// NewHTTPService creates a new HTTP service
func NewHTTPService(logger *logrus.Logger) *HTTPService {
	return &HTTPService{
		logger: logger,
	}
}

// ProcessRequest processes an HTTP request
func (s *HTTPService) ProcessRequest(req *models.HTTPRequest) (*models.HTTPResponse, error) {
	// Log the incoming request
	s.logger.WithFields(logrus.Fields{
		"method":      req.Method,
		"url":         req.URL,
		"headers":     req.Headers,
		"queryParams": req.QueryParams,
	}).Info("Processing HTTP request")

	// Create HTTP client with timeout
	timeout := 30 * time.Second // default timeout
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}
	client := &http.Client{
		Timeout: timeout,
	}

	// Prepare request body
	bodyReader, err := s.prepareRequestBody(req)
	if err != nil {
		return nil, err
	}

	// Create request
	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create HTTP request")
		return nil, err
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Add query parameters
	if len(req.QueryParams) > 0 {
		q := httpReq.URL.Query()
		for key, value := range req.QueryParams {
			q.Add(key, value)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	// Record start time
	startTime := time.Now()

	// Send request
	resp, err := client.Do(httpReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to send HTTP request")
		return nil, err
	}
	defer resp.Body.Close()

	// Calculate time elapsed
	timeElapsed := time.Since(startTime).Seconds()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.WithError(err).Error("Failed to read response body")
		return nil, err
	}

	// Try to parse JSON response
	var responseBody interface{}
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		if err := json.Unmarshal(bodyBytes, &responseBody); err != nil {
			// If not JSON, use raw body
			responseBody = string(bodyBytes)
		}
	} else {
		responseBody = string(bodyBytes)
	}

	// Prepare response
	response := &models.HTTPResponse{
		Status:      resp.StatusCode,
		Headers:     make(map[string]string),
		Body:        responseBody,
		TimeElapsed: timeElapsed,
	}

	// Copy headers
	for key, values := range resp.Header {
		if len(values) > 0 {
			response.Headers[key] = values[0]
		}
	}

	return response, nil
}

// prepareRequestBody prepares the request body based on content type
func (s *HTTPService) prepareRequestBody(req *models.HTTPRequest) (io.Reader, error) {
	if req.Body == nil {
		return nil, nil
	}

	contentType := req.Headers["Content-Type"]
	switch {
	case strings.Contains(contentType, "application/json"):
		jsonData, err := json.Marshal(req.Body)
		if err != nil {
			s.logger.WithError(err).Error("Failed to marshal JSON body")
			return nil, err
		}
		return bytes.NewBuffer(jsonData), nil

	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		if formData, ok := req.Body.(map[string]interface{}); ok {
			form := make(url.Values)
			for k, v := range formData {
				form.Set(k, fmt.Sprintf("%v", v))
			}
			return strings.NewReader(form.Encode()), nil
		}

	case strings.Contains(contentType, "multipart/form-data"):
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		if formData, ok := req.Body.(map[string]interface{}); ok {
			for k, v := range formData {
				if file, ok := v.(map[string]interface{}); ok {
					// Handle file upload
					if fileName, ok := file["filename"].(string); ok {
						if fileContent, ok := file["content"].(string); ok {
							part, err := writer.CreateFormFile(k, fileName)
							if err != nil {
								s.logger.WithError(err).Error("Failed to create form file")
								continue
							}
							part.Write([]byte(fileContent))
						}
					}
				} else {
					// Handle regular form field
					writer.WriteField(k, fmt.Sprintf("%v", v))
				}
			}
			writer.Close()
			req.Headers["Content-Type"] = writer.FormDataContentType()
			return body, nil
		}

	default:
		// Handle raw body
		if rawBody, ok := req.Body.(string); ok {
			return strings.NewReader(rawBody), nil
		}
	}

	return nil, fmt.Errorf("unsupported content type: %s", contentType)
}
