package models

import (
	"time"

	"github.com/flowpilotx/libs/logger"
	"github.com/gorilla/websocket"
)

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	ReadBufferSize    int
	WriteBufferSize   int
	HandshakeTimeout  time.Duration
	PingInterval      time.Duration
	PongWait         time.Duration
}

// WebSocketClient represents a WebSocket client connection
type WebSocketClient struct {
	Conn   *websocket.Conn
	Send   chan []byte
	Log    logger.LoggerInterface
	Config WebSocketConfig
} 