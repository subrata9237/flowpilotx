package models

// HTTPRequest represents an HTTP request
type HTTPRequest struct {
	Method      string            `json:"method" binding:"required"`
	URL         string            `json:"url" binding:"required"`
	Headers     map[string]string `json:"headers"`
	Body        interface{}       `json:"body"`
	QueryParams map[string]string `json:"query_params"`
	Timeout     int               `json:"timeout"` // in seconds
}

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	Status      int               `json:"status"`
	Headers     map[string]string `json:"headers"`
	Body        interface{}       `json:"body"`
	TimeElapsed float64           `json:"time_elapsed"` // in seconds
} 