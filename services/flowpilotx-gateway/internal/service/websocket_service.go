package service

import (
	"context" // alias for standard log package
	"io"

	// alias for standard log package
	"net/http"
	"sync"
	"time"

	stdlog "log"

	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/logger"
	"github.com/gorilla/websocket"
)

// Message represents a WebSocket message to be processed
type Message struct {
	Type    int
	Payload []byte
	Conn    *websocket.Conn
}

// WebSocketService handles WebSocket connections and message processing
type WebSocketService struct {
	cfg          *config.Config
	log          logger.LoggerInterface
	upgrader     websocket.Upgrader
	connections  sync.Map
	msgChan      chan Message
	workerCount  int
	done         chan struct{}
	connMutex    sync.Mutex
	msgProcessor *MessageProcessor
}

// NewWebSocketService creates a new WebSocket service instance
func NewWebSocketService(cfg *config.Config, log logger.LoggerInterface) *WebSocketService {
	workerCount := 5 // Number of worker goroutines
	stdlog.SetOutput(io.Discard)
	// Create upgrader with compression
	upgrader := websocket.Upgrader{
		ReadBufferSize:    cfg.FlowpilotxGateway.WebSocket.ReadBufferSize,
		WriteBufferSize:   cfg.FlowpilotxGateway.WebSocket.WriteBufferSize,
		HandshakeTimeout:  cfg.FlowpilotxGateway.WebSocket.HandshakeTimeout,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}
	return &WebSocketService{
		cfg:          cfg,
		log:          log,
		msgChan:      make(chan Message, 1000),
		workerCount:  workerCount,
		done:         make(chan struct{}),
		msgProcessor: NewMessageProcessor(log),
		upgrader:     upgrader,
	}
}

// HandleConnection handles incoming WebSocket connections
func (s *WebSocketService) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Create a context with cancel for this connection
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Log connection attempt
	s.log.InfoWithCtx(ctx, "WebSocket connection attempt", map[string]interface{}{
		"remote_addr": r.RemoteAddr,
		"user_agent":  r.UserAgent(),
	})

	// Upgrade connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.ErrorWithCtx(ctx, "Failed to upgrade connection", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Enable compression if supported
	conn.EnableWriteCompression(true)

	// Store connection
	connID := conn.RemoteAddr().String()
	s.connections.Store(connID, conn)

	// Create error channel for goroutine communication
	errChan := make(chan error, 1)
	closeChan := make(chan struct{})

	// Setup connection parameters
	conn.SetReadLimit(int64(s.cfg.FlowpilotxGateway.WebSocket.ReadBufferSize))
	conn.SetReadDeadline(time.Now().Add(s.cfg.FlowpilotxGateway.WebSocket.PongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(s.cfg.FlowpilotxGateway.WebSocket.PongWait))
	})

	// Start read loop in a separate goroutine
	go s.readHandler(ctx, conn, errChan, closeChan)

	// Start message processor workers if not already started
	s.startMessageProcessors(ctx)

	// Start ping ticker
	ticker := time.NewTicker(s.cfg.FlowpilotxGateway.WebSocket.PingInterval)
	defer ticker.Stop()

	// Handle connection lifecycle
	for {
		select {
		case <-ctx.Done():
			s.cleanupConnection(conn, connID, ctx)
			return
		case <-s.done:
			s.cleanupConnection(conn, connID, ctx)
			return
		case <-closeChan:
			s.cleanupConnection(conn, connID, ctx)
			return
		case err := <-errChan:
			if err != nil {
				if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					s.log.ErrorWithCtx(ctx, "WebSocket error", map[string]interface{}{
						"error":       err.Error(),
						"remote_addr": connID,
					})
				}
			}
			s.cleanupConnection(conn, connID, ctx)
			return
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
				s.log.ErrorWithCtx(ctx, "Failed to send ping", map[string]interface{}{
					"error": err.Error(),
				})
				return
			}
		}
	}
}

// startMessageProcessors starts the worker goroutines for processing messages
func (s *WebSocketService) startMessageProcessors(ctx context.Context) {
	for i := 0; i < s.workerCount; i++ {
		go s.messageProcessor(ctx)
	}
}

// messageProcessor processes messages from the message channel
func (s *WebSocketService) messageProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		case msg := <-s.msgChan:
			if err := s.msgProcessor.ProcessMessage(ctx, msg.Type, msg.Payload, msg.Conn); err != nil {
				s.log.ErrorWithCtx(ctx, "Failed to process message", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}
}

// readHandler handles reading messages from the WebSocket connection
func (s *WebSocketService) readHandler(ctx context.Context, conn *websocket.Conn, errChan chan<- error, closeChan chan<- struct{}) {
	defer func() {
		if r := recover(); r != nil {
			s.log.ErrorWithCtx(ctx, "Recovered from panic in readHandler", map[string]interface{}{
				"error": r,
			})
		}
		closeChan <- struct{}{} // Signal connection should be closed
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		default:

			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					s.log.ErrorWithCtx(ctx, "WebSocket read error", map[string]interface{}{
						"error": err.Error(),
					})
				}
				errChan <- err
				return
			}

			// Send message to channel for processing
			select {
			case s.msgChan <- Message{Type: messageType, Payload: message, Conn: conn}:
			case <-ctx.Done():
				return
			case <-s.done:
				return
			default:
				s.log.ErrorWithCtx(ctx, "Message channel full, dropping message", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}
}

// GetConnectionCount returns the number of active connections
func (s *WebSocketService) GetConnectionCount() int {
	count := 0
	s.connections.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// cleanupConnection handles cleaning up a WebSocket connection
func (s *WebSocketService) cleanupConnection(conn *websocket.Conn, connID string, ctx context.Context) {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	// Remove from connections map
	s.connections.Delete(connID)

	// Send close message
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	err := conn.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(time.Second))
	if err != nil {
		s.log.ErrorWithCtx(ctx, "Failed to send close message", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Close the connection
	conn.Close()

	// Log connection closure
	s.log.InfoWithCtx(ctx, "WebSocket connection closed", map[string]interface{}{
		"remote_addr": connID,
	})
}

// Shutdown gracefully shuts down the WebSocket service
func (s *WebSocketService) Shutdown(ctx context.Context) {
	close(s.done) // Signal all goroutines to stop

	// Close all active connections
	s.connections.Range(func(key, value interface{}) bool {
		if conn, ok := value.(*websocket.Conn); ok {
			s.cleanupConnection(conn, key.(string), ctx)
		}
		return true
	})

	// Wait for message processing to complete
	close(s.msgChan)
}
