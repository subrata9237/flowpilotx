package models

import "time"

// WebSocketConfig holds configuration for WebSocket connections
type WebSocketConfig struct {
	ReadBufferSize   int           `json:"readBufferSize" example:"1024"`
	WriteBufferSize  int           `json:"writeBufferSize" example:"1024"`
	HandshakeTimeout time.Duration `json:"handshakeTimeout" example:"10s"`
	PingInterval     time.Duration `json:"pingInterval" example:"30s"`
	PongWait         time.Duration `json:"pongWait" example:"60s"`
	WriteTimeout     time.Duration `json:"writeTimeout" example:"10s"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	EventID string      `json:"event_id"`
	Message interface{} `json:"message"`
}

// WebSocketResponse represents a WebSocket response
type WebSocketResponse struct {
	EventID    string      `json:"event_id"`
	Success    bool        `json:"success"`
	Message    interface{} `json:"message,omitempty"`
	Error      string      `json:"error,omitempty"`
	StatusCode int         `json:"status_code,omitempty"`
}
