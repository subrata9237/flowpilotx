package models

// HTTPRequest represents an HTTP request with all its parameters
type HTTPRequest struct {
	Method      string            `json:"method" validate:"required,oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"`
	URL         string            `json:"url" validate:"required,url"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        interface{}       `json:"body,omitempty"`
	FormData    map[string]interface{} `json:"form_data,omitempty"`    // For form data
	Files       []FileUpload     `json:"files,omitempty"`      // For file uploads
	QueryParams map[string]string `json:"query_params,omitempty"`
	Timeout     int              `json:"timeout,omitempty"`    // timeout in seconds
	RetryCount  int              `json:"retry_count,omitempty"`
	RetryDelay  int              `json:"retry_delay,omitempty"` // delay in seconds
}

// FileUpload represents a file to be uploaded
type FileUpload struct {
	FieldName   string `json:"field_name"`
	FileName    string `json:"file_name"`
	Content     []byte `json:"content"`
	ContentType string `json:"content_type,omitempty"`
}

// HTTPResponse represents an HTTP response with all its details
type HTTPResponse struct {
	StatusCode    int               `json:"status_code"`
	Headers       map[string]string `json:"headers"`
	Body          interface{}       `json:"body"`            // Processed body (parsed JSON, XML, etc.)
	RawBody       []byte           `json:"raw_body"`        // Original raw response
	ContentType   string           `json:"content_type"`    // Response content type
	ResponseTime  float64          `json:"response_time"`   // in seconds
	Error         string           `json:"error,omitempty"` // Error message if any
	Format        string           `json:"format"`          // Detected format (json, xml, yaml, text, binary, etc.)
	IsBase64      bool             `json:"is_base64"`       // Whether the body is base64 encoded
	FileExtension string           `json:"file_extension,omitempty"` // Detected file extension
	TextEncoding  string           `json:"text_encoding,omitempty"`  // Text encoding if detected
	Size          int64            `json:"size"`            // Size of the response in bytes
}

// APIResponse represents a standardized API response
type APIResponse struct {
	Status  string      `json:"status"`           // "success" or "error"
	Message string      `json:"message,omitempty"` // Optional message
	Data    interface{} `json:"data,omitempty"`   // Response data
	Error   string     `json:"error,omitempty"`   // Error message if any
} 