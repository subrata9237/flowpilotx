package websocket

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/flowpilotx/libs/logger"
	"github.com/gorilla/websocket"
)

// Handler manages WebSocket connections
type Handler struct {
	log      logger.LoggerInterface
	config   WebSocketConfig
	upgrader websocket.Upgrader
	// Track active connections
	connections sync.Map
}

// NewHandler creates a new WebSocket handler
func NewHandler(log logger.LoggerInterface, config WebSocketConfig) http.Handler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for now
		},
		HandshakeTimeout: config.HandshakeTimeout,
	}

	return &Handler{
		log:      log,
		config:   config,
		upgrader: upgrader,
	}
}

// ServeHTTP handles WebSocket connections
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Create a context with cancel for this connection
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Log connection attempt
	h.log.Info(ctx, "WebSocket connection attempt", map[string]interface{}{
		"remote_addr": r.RemoteAddr,
		"user_agent": r.UserAgent(),
	})

	// Upgrade connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error(ctx, "Failed to upgrade connection", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Ensure connection is closed when we're done
	defer func() {
		conn.Close()
		h.connections.Delete(conn.RemoteAddr().String())
		h.log.Info(ctx, "WebSocket connection closed", map[string]interface{}{
			"remote_addr": conn.RemoteAddr().String(),
		})
	}()

	// Store connection
	h.connections.Store(conn.RemoteAddr().String(), conn)

	// Log successful connection
	h.log.Info(ctx, "WebSocket connection established", map[string]interface{}{
		"remote_addr": conn.RemoteAddr().String(),
	})

	// Create error channel for goroutine communication
	errChan := make(chan error, 1)

	// Setup connection parameters
	conn.SetReadLimit(int64(h.config.ReadBufferSize))
	
	// Setup ping handler
	conn.SetPingHandler(func(message string) error {
		err := conn.SetReadDeadline(time.Now().Add(h.config.PongWait))
		if err != nil {
			return err
		}
		return conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(h.config.WriteTimeout))
	})

	// Setup pong handler
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(h.config.PongWait))
	})

	// Start ping ticker in a separate goroutine
	go func() {
		pingTicker := time.NewTicker(h.config.PingInterval)
		defer pingTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-pingTicker.C:
				// Set write deadline for ping
				if err := conn.SetWriteDeadline(time.Now().Add(h.config.WriteTimeout)); err != nil {
					errChan <- err
					return
				}
				// Send ping message
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					errChan <- err
					return
				}
			}
		}
	}()

	// Start read loop in a separate goroutine
	go func() {
		defer func() {
			close(errChan)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Set read deadline
				if err := conn.SetReadDeadline(time.Now().Add(h.config.PongWait)); err != nil {
					errChan <- err
					return
				}

				messageType, message, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						h.log.Error(ctx, "WebSocket read error", map[string]interface{}{
							"error": err.Error(),
						})
					}
					errChan <- err
					return
				}

				// Log received message
				h.log.Info(ctx, "Received message", map[string]interface{}{
					"message": string(message),
					"type":    messageType,
				})

				// Set write deadline for response
				if err := conn.SetWriteDeadline(time.Now().Add(h.config.WriteTimeout)); err != nil {
					errChan <- err
					return
				}

				// Echo the message back
				if err := conn.WriteMessage(messageType, message); err != nil {
					errChan <- err
					return
				}

				// Log sent message
				h.log.Info(ctx, "Sent message", map[string]interface{}{
					"message": string(message),
					"type":    messageType,
				})
			}
		}
	}()

	// Wait for any errors
	select {
	case err := <-errChan:
		if err != nil {
			h.log.Error(ctx, "WebSocket error", map[string]interface{}{
				"error": err.Error(),
				"remote_addr": conn.RemoteAddr().String(),
			})
		}
	case <-ctx.Done():
		h.log.Info(ctx, "WebSocket connection context cancelled", map[string]interface{}{
			"remote_addr": conn.RemoteAddr().String(),
		})
	}
} 