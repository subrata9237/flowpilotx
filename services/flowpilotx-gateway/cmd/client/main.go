package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/flowpilotx/services/flowpilotx-gateway/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	addr     = flag.String("addr", "localhost:8080", "gateway server address")
	basePath = flag.String("base", "/api/flowpilotx-gateway", "base path for all endpoints")
	wsPath   = flag.String("ws", "/ws/flowpilotx-gateway", "websocket path")
	testMode = flag.String("mode", "echo", "test mode: ping, echo, or load")
	msgCount = flag.Int("n", 10, "number of messages to send in load test")
)

// Response structures for API endpoints
type HealthResponse struct {
	Status string `json:"status"`
}

type VersionResponse struct {
	Version string `json:"version"`
}

func main() {
	flag.Parse()
	*testMode = "echo"
	// Ensure base path starts with / and doesn't end with /
	*basePath = "/" + strings.Trim(*basePath, "/")

	// Create HTTP client for REST APIs
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	// Call Health API
	healthResp, err := callHealthAPI(httpClient, *addr, *basePath)
	if err != nil {
		log.Printf("Health API error: %v", err)
	} else {
		log.Printf("Health Status: %s", healthResp.Status)
	}

	// Call Version API
	versionResp, err := callVersionAPI(httpClient, *addr, *basePath)
	if err != nil {
		log.Printf("Version API error: %v", err)
	} else {
		log.Printf("Gateway Version: %s", versionResp.Version)
	}

	// WebSocket Connection
	url := fmt.Sprintf("ws://%s%s", *addr, *wsPath)
	log.Printf("Connecting to WebSocket at %s", url)

	// Set up WebSocket dialer with custom headers
	dialer := websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  45 * time.Second,
		EnableCompression: true,
	}

	// Add custom headers
	headers := http.Header{}
	headers.Add("User-Agent", "FlowPilotX-Client/1.0")
	headers.Add("Sec-WebSocket-Protocol", "flowpilotx-v1")

	// Connect to WebSocket server
	c, resp, err := dialer.Dial(url, headers)
	if err != nil {
		if resp != nil {
			log.Printf("WebSocket dial failed with status: %d", resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			log.Printf("Response body: %s", string(body))
			resp.Body.Close()
		}
		log.Fatal("WebSocket dial:", err)
	}
	defer c.Close()

	// Log successful connection
	log.Printf("Connected to WebSocket at %s", url)

	// Create context for the entire client session
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create done channel for graceful shutdown
	done := make(chan struct{})

	// Setup interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Start read goroutine
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("read error: %v", err)
					}
					return
				}
				// Parse the response
				var response models.WebSocketResponse
				if err := json.Unmarshal(message, &response); err != nil {
					log.Printf("Failed to parse response: %v", err)
					continue
				}
				log.Printf("recv: EventID=%s, Success=%v, Data=%v, Error=%s, StatusCode=%d",
					response.EventID, response.Success, response.Message, response.Error, response.StatusCode)
			}
		}
	}()

	// Start test mode goroutine based on the selected mode
	switch *testMode {
	case "gg":
		go runPingTest(ctx, c, interrupt)
	case "echo":
		go runEchoTest(ctx, c, interrupt)
	case "load":
		go runLoadTest(ctx, c, interrupt, *msgCount)
	default:
		log.Printf("Unknown test mode: %s", *testMode)
		return
	}

	// Wait for interrupt signal
	select {
	case <-interrupt:
		log.Println("Received interrupt signal")
	case <-done:
		log.Println("Connection closed")
	}

	// Cleanup
	log.Println("Closing connection...")
	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Printf("write close error: %v", err)
	}
	<-done
	log.Println("Connection closed")
}

func runPingTest(ctx context.Context, c *websocket.Conn, interrupt chan os.Signal) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-interrupt:
			return
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("ping"))
			if err != nil {
				log.Printf("write error: %v", err)
				return
			}
			log.Printf("sent: EventID=%s, Message=ping", uuid.New().String())
		}
	}
}

func runEchoTest(ctx context.Context, c *websocket.Conn, interrupt chan os.Signal) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		case <-interrupt:
			return
		default:
			fmt.Print("Enter message (or 'quit' to exit): ")
			if scanner.Scan() {
				message := scanner.Text()
				if message == "quit" {
					return
				}

				// Create WebSocket message
				wsMessage := models.WebSocketMessage{
					EventID: uuid.New().String(),
					Message: message,
				}

				// Convert to binary
				messageBytes, err := json.Marshal(wsMessage)
				if err != nil {
					log.Printf("Failed to marshal message: %v", err)
					continue
				}

				err = c.WriteMessage(websocket.BinaryMessage, messageBytes)
				if err != nil {
					log.Printf("write error: %v", err)
					return
				}
				log.Printf("sent: EventID=%s, Payload=%s", wsMessage.EventID, wsMessage.Message)
			}
		}
	}
}

func runLoadTest(ctx context.Context, c *websocket.Conn, interrupt chan os.Signal, count int) {
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return
		case <-interrupt:
			return
		default:
			// Create test message
			wsMessage := models.WebSocketMessage{
				EventID: uuid.New().String(),
				Message: fmt.Sprintf("load test message %d", i+1),
			}

			// Convert to binary
			messageBytes, err := json.Marshal(wsMessage)
			if err != nil {
				log.Printf("Failed to marshal message: %v", err)
				continue
			}

			err = c.WriteMessage(websocket.BinaryMessage, messageBytes)
			if err != nil {
				log.Printf("write error: %v", err)
				return
			}
			log.Printf("sent: EventID=%s, Message=%s", wsMessage.EventID, wsMessage.Message)

			// Add a small delay between messages
			time.Sleep(100 * time.Millisecond)
		}
	}
	log.Printf("Load test completed: sent %d messages", count)
}

func callHealthAPI(client *http.Client, addr, basePath string) (*HealthResponse, error) {
	url := fmt.Sprintf("http://%s%s/health", addr, basePath)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var healthResp HealthResponse
	if err := json.Unmarshal(body, &healthResp); err != nil {
		return nil, err
	}

	return &healthResp, nil
}

func callVersionAPI(client *http.Client, addr, basePath string) (*VersionResponse, error) {
	url := fmt.Sprintf("http://%s%s/version", addr, basePath)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var versionResp VersionResponse
	if err := json.Unmarshal(body, &versionResp); err != nil {
		return nil, err
	}

	return &versionResp, nil
}
