package main

import (
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
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
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

	// Create done channel for graceful shutdown
	done := make(chan struct{})

	// Setup interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Start read goroutine
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("read error: %v", err)
				}
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	// Handle different test modes
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	switch *testMode {
	case "ping":
		go pingTest(ctx, c)
	case "echo":
		go echoTest(ctx, c)
	case "load":
		go loadTest(ctx, c, *msgCount)
	default:
		log.Fatalf("unknown test mode: %s", *testMode)
	}

	// Wait for interrupt signal
	select {
	case <-interrupt:
		log.Println("interrupt received, closing connection...")
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("write close:", err)
		}
		select {
		case <-done:
			log.Println("connection closed cleanly")
		case <-time.After(time.Second):
			log.Println("timeout waiting for connection to close")
		}
		return
	}
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

// pingTest sends periodic ping messages
func pingTest(ctx context.Context, c *websocket.Conn) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("ping"))
			if err != nil {
				log.Println("write:", err)
				return
			}
			log.Println("sent: ping")
		}
	}
}

// echoTest reads input from stdin and sends it to the server
func echoTest(ctx context.Context, c *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var message string
			fmt.Print("Enter message (or 'quit' to exit): ")
			fmt.Scanln(&message)

			if message == "quit" {
				return
			}

			err := c.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Println("write:", err)
				return
			}
			log.Printf("sent: %s", message)
		}
	}
}

// loadTest sends a specified number of messages as quickly as possible
func loadTest(ctx context.Context, c *websocket.Conn, count int) {
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			message := fmt.Sprintf("load test message %d", i+1)
			err := c.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Printf("write error: %v", err)
				return
			}
			log.Printf("sent: %s", message)
			time.Sleep(100 * time.Millisecond) // Small delay to prevent overwhelming the server
		}
	}
	log.Printf("Load test complete: sent %d messages", count)
} 