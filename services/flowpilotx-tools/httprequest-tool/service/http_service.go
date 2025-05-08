package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/flowpilotx/services/flowpilotx-tools/httprequest-tool/model"
	"github.com/gabriel-vasile/mimetype"
	logrus "github.com/sirupsen/logrus"
)

// HTTPService handles HTTP request processing
type HTTPService struct {
	logger *logrus.Logger
	client *http.Client
}

// NewHTTPService creates a new HTTP service
func NewHTTPService(logger *logrus.Logger) *HTTPService {
	return &HTTPService{
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessRequest processes an HTTP request
func (s *HTTPService) ProcessRequest(req *model.HTTPRequest) (*model.HTTPResponse, error) {
	// Set custom timeout if specified
	if req.Timeout > 0 {
		s.client.Timeout = time.Duration(req.Timeout) * time.Second
	}

	// Prepare request body based on content type
	bodyReader, contentType, err := s.prepareRequestBody(req)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request body: %w", err)
	}

	// Create request
	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set content type if not already set
	if contentType != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	// Add query parameters
	if len(req.QueryParams) > 0 {
		q := httpReq.URL.Query()
		for key, value := range req.QueryParams {
			q.Add(key, value)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	// Make the request with retry logic
	var resp *http.Response
	var lastErr error
	retries := req.RetryCount
	if retries <= 0 {
		retries = 1
	}

	for i := 0; i < retries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(req.RetryDelay) * time.Second)
			s.logger.Infof("Retrying request (attempt %d/%d)", i+1, retries)
		}

		startTime := time.Now()
		resp, lastErr = s.client.Do(httpReq)
		if lastErr == nil {
			defer resp.Body.Close()
			return s.processResponse(resp, time.Since(startTime).Seconds())
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", retries, lastErr)
}

func (s *HTTPService) prepareRequestBody(req *model.HTTPRequest) (io.Reader, string, error) {
	if req.Body == nil && len(req.FormData) == 0 && len(req.Files) == 0 {
		return nil, "", nil
	}

	contentType := req.Headers["Content-Type"]
	switch {
	case strings.Contains(contentType, "application/json"):
		jsonData, err := json.Marshal(req.Body)
		if err != nil {
			return nil, "", err
		}
		return bytes.NewBuffer(jsonData), "application/json", nil

	case strings.Contains(contentType, "application/xml"):
		xmlData, err := xml.Marshal(req.Body)
		if err != nil {
			return nil, "", err
		}
		return bytes.NewBuffer(xmlData), "application/xml", nil

	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		formData := url.Values{}
		for key, value := range req.FormData {
			formData.Set(key, fmt.Sprintf("%v", value))
		}
		return strings.NewReader(formData.Encode()), "application/x-www-form-urlencoded", nil

	case strings.Contains(contentType, "multipart/form-data"):
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add form fields
		for key, value := range req.FormData {
			if err := writer.WriteField(key, fmt.Sprintf("%v", value)); err != nil {
				return nil, "", err
			}
		}

		// Add files
		for _, file := range req.Files {
			part, err := writer.CreateFormFile(file.FieldName, file.FileName)
			if err != nil {
				return nil, "", err
			}
			if _, err := part.Write(file.Content); err != nil {
				return nil, "", err
			}
		}

		if err := writer.Close(); err != nil {
			return nil, "", err
		}
		return body, writer.FormDataContentType(), nil

	default:
		// Try JSON as default
		if req.Body != nil {
			jsonData, err := json.Marshal(req.Body)
				if err != nil {
					return nil, "", err
				}
				return bytes.NewBuffer(jsonData), "application/json", nil
		}
	}

	return nil, "", fmt.Errorf("unsupported content type or body format")
}

func (s *HTTPService) processResponse(resp *http.Response, elapsed float64) (*model.HTTPResponse, error) {
	// Read raw response body
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Convert headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Initialize response
	response := &model.HTTPResponse{
		StatusCode:   resp.StatusCode,
		Headers:      headers,
		RawBody:      rawBody,
		ContentType:  resp.Header.Get("Content-Type"),
		ResponseTime: elapsed,
		Size:         int64(len(rawBody)),
	}

	// Detect MIME type if not provided or ambiguous
	detectedMIME := mimetype.Detect(rawBody)
	if response.ContentType == "" || strings.Contains(response.ContentType, "application/octet-stream") {
		response.ContentType = detectedMIME.String()
	}

	// Parse content type
	mediaType, params, err := mime.ParseMediaType(response.ContentType)
	if err == nil && params["charset"] != "" {
		response.TextEncoding = params["charset"]
	}

	// Process body based on content type
	switch {
	case strings.HasPrefix(mediaType, "application/json"):
		response.Format = "json"
		var jsonBody interface{}
		if err := json.Unmarshal(rawBody, &jsonBody); err == nil {
			response.Body = jsonBody
		} else {
			response.Body = string(rawBody)
		}

	case strings.HasPrefix(mediaType, "application/xml"):
		response.Format = "xml"
		var xmlBody interface{}
		if err := xml.Unmarshal(rawBody, &xmlBody); err == nil {
			response.Body = xmlBody
		} else {
				response.Body = string(rawBody)
		}

	case strings.HasPrefix(mediaType, "text/"):
		response.Format = "text"
		if utf8.Valid(rawBody) {
			response.Body = string(rawBody)
		} else {
			response.Body = base64.StdEncoding.EncodeToString(rawBody)
			response.IsBase64 = true
		}

	default:
		// Binary data
		response.Format = "binary"
		response.Body = base64.StdEncoding.EncodeToString(rawBody)
		response.IsBase64 = true
			
			// Try to detect file type
			ext := detectedMIME.Extension()
			if ext != "" {
				response.FileExtension = strings.TrimPrefix(ext, ".")
		}
	}

	return response, nil
}

// ProcessAsyncRequest processes an HTTP request asynchronously
func (s *HTTPService) ProcessAsyncRequest(req *model.HTTPRequest) (string, error) {
	requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())
	
	go func() {
		response, err := s.ProcessRequest(req)
		if err != nil {
			s.logger.WithError(err).WithField("request_id", requestID).Error("Async request failed")
			return
		}
		s.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"status":     response.StatusCode,
		}).Info("Async request completed")
	}()

	return requestID, nil
}

// BatchProcess processes multiple HTTP requests concurrently
func (s *HTTPService) BatchProcess(requests []model.HTTPRequest) ([]*model.HTTPResponse, error) {
	if len(requests) == 0 {
		return nil, fmt.Errorf("no requests to process")
	}

	responses := make([]*model.HTTPResponse, len(requests))
	errors := make([]error, len(requests))
	done := make(chan bool, len(requests))

	// Process requests concurrently
	for i, req := range requests {
		go func(index int, request model.HTTPRequest) {
			resp, err := s.ProcessRequest(&request)
			responses[index] = resp
			errors[index] = err
			done <- true
		}(i, req)
	}

	// Wait for all requests to complete
	for i := 0; i < len(requests); i++ {
		<-done
	}

	// Check for errors
	var errorMsgs []string
	for i, err := range errors {
		if err != nil {
			errorMsgs = append(errorMsgs, fmt.Sprintf("request %d: %v", i+1, err))
		}
	}

	if len(errorMsgs) > 0 {
		return responses, fmt.Errorf("batch processing had errors: %s", strings.Join(errorMsgs, "; "))
	}

	return responses, nil
}
