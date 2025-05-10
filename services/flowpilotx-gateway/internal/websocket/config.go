package websocket

import "time"

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	ReadBufferSize   int           // Size of the read buffer
	WriteBufferSize  int           // Size of the write buffer
	HandshakeTimeout time.Duration // Timeout for WebSocket handshake
	PingInterval     time.Duration // Interval between ping messages
	PongWait         time.Duration // How long to wait for pong response
	WriteTimeout     time.Duration // Timeout for write operations
} 