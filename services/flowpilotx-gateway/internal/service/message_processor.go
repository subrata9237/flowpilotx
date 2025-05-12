package service

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/models"
	"github.com/gorilla/websocket"
)

// MessageProcessor handles WebSocket message processing
type MessageProcessor struct {
	log logger.LoggerInterface
}

// NewMessageProcessor creates a new message processor instance
func NewMessageProcessor(log logger.LoggerInterface) *MessageProcessor {
	return &MessageProcessor{
		log: log,
	}
}

// ProcessMessage handles incoming WebSocket messages
func (p *MessageProcessor) ProcessMessage(ctx context.Context, messageType int, payload []byte, conn *websocket.Conn) error {

	p.log.InfoWithCtx(ctx, "Processing message", map[string]interface{}{
		"type":    messageType,
		"payload": string(payload),
	})

	var response *models.WebSocketResponse

	switch messageType {
	case websocket.TextMessage:
		response = p.handleTextMessage(ctx, payload)
	case websocket.BinaryMessage:
		response = p.handleBinaryMessage(ctx, payload)
	default:
		response = &models.WebSocketResponse{
			Success:    false,
			Message:    "Unsupported message type",
			StatusCode: http.StatusBadRequest,
		}
	}

	return p.sendResponse(ctx, conn, response)
}

// handleTextMessage processes text messages
func (p *MessageProcessor) handleTextMessage(ctx context.Context, payload []byte) *models.WebSocketResponse {
	message := string(payload)

	// Handle ping message
	if message == "ping" {
		return &models.WebSocketResponse{
			Success:    true,
			Message:    "pong",
			StatusCode: http.StatusOK,
		}
	}

	// Handle other text messages
	return &models.WebSocketResponse{
		Success:    true,
		Message:    message,
		StatusCode: http.StatusOK,
	}
}

// handleBinaryMessage processes binary messages
func (p *MessageProcessor) handleBinaryMessage(ctx context.Context, payload []byte) *models.WebSocketResponse {
	var wsMessage models.WebSocketMessage
	if err := json.Unmarshal(payload, &wsMessage); err != nil {
		p.log.ErrorWithCtx(ctx, "Failed to parse binary message", map[string]interface{}{
			"error": err.Error(),
		})
		//print wsMessage use p.log methods
		p.log.DebugWithCtx(ctx, "Invalid message format", map[string]interface{}{
			"error":   err.Error(),
			"message": wsMessage,
		})
		return &models.WebSocketResponse{
			Success:    false,
			Message:    "Invalid message format",
			Error:      err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	// Process the message
	if wsMessage.Message == "ping" {
		return &models.WebSocketResponse{
			EventID:    wsMessage.EventID,
			Success:    true,
			Message:    "pong",
			StatusCode: http.StatusOK,
		}
	}

	// Handle other binary messages
	return &models.WebSocketResponse{
		EventID:    wsMessage.EventID,
		Success:    true,
		Message:    wsMessage.Message,
		StatusCode: http.StatusOK,
	}
}

// sendResponse sends the response back through the WebSocket connection
func (p *MessageProcessor) sendResponse(ctx context.Context, conn *websocket.Conn, response *models.WebSocketResponse) error {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		p.log.ErrorWithCtx(ctx, "Failed to marshal response", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Set write deadline
	if err := conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		p.log.ErrorWithCtx(ctx, "Failed to set write deadline", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Send the response
	if err := conn.WriteMessage(websocket.BinaryMessage, responseBytes); err != nil {
		p.log.ErrorWithCtx(ctx, "Failed to send response", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Log the response
	p.log.DebugWithCtx(ctx, "Response sent", map[string]interface{}{
		"response": response,
	})

	return nil
}
