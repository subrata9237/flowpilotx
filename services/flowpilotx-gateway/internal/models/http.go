package models

import (
	"bufio"
	"net"
	"net/http"
	"time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// VersionResponse represents the version information response
type VersionResponse struct {
	Version     string `json:"version"`
	Environment string `json:"environment"`
	ServiceName string `json:"service_name"`
}

// ResponseWriter is a custom response writer that captures the status code
type ResponseWriter struct {
	http.ResponseWriter
	Status int
}

// WriteHeader captures the status code before writing it
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.Status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the 200 status code if WriteHeader hasn't been called
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if rw.Status == 0 {
		rw.Status = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

// Hijack implements the http.Hijacker interface to allow WebSocket upgrades
func (rw *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
} 