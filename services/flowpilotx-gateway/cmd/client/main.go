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
	testMode = flag.String("mode", "echo", "test mode: ping, echo, load, or workflow")
	msgCount = flag.Int("n", 10, "number of messages to send in load test")
)

// Response structures for API endpoints
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type VersionResponse struct {
	ServiceName string `json:"service_name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
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
		log.Printf("Health Status: %s, Version: %s", healthResp.Status, healthResp.Version)
	}

	// Call Version API
	versionResp, err := callVersionAPI(httpClient, *addr, *basePath)
	if err != nil {
		log.Printf("Version API error: %v", err)
	} else {
		log.Printf("Service: %s, Version: %s, Environment: %s",
			versionResp.ServiceName, versionResp.Version, versionResp.Environment)
	}

	// Handle different test modes
	switch *testMode {
	case "workflow":
		runWorkflowTest(httpClient, *addr, *basePath)
	case "ping", "echo", "load":
		runWebSocketTest(*addr, *wsPath, *testMode, *msgCount)
	default:
		log.Printf("Unknown test mode: %s", *testMode)
		log.Printf("Available modes: workflow, ping, echo, load")
		return
	}
}

func runWorkflowTest(client *http.Client, addr, basePath string) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nWorkflow Test Menu:\n")
		fmt.Print("1. Create Workflow\n")
		fmt.Print("2. Get Workflow\n")
		fmt.Print("3. Trigger Workflow\n")
		fmt.Print("4. Get Workflow Request\n")
		fmt.Print("5. Get Activity Results\n")
		fmt.Print("q. Quit\n")
		fmt.Print("Enter choice: ")

		if scanner.Scan() {
			choice := scanner.Text()
			switch choice {
			case "1":
				createWorkflow(client, addr, basePath)
			case "2":
				getWorkflow(client, addr, basePath)
			case "3":
				triggerWorkflow(client, addr, basePath)
			case "4":
				getWorkflowRequest(client, addr, basePath)
			case "5":
				getActivityResults(client, addr, basePath)
			case "q":
				return
			default:
				log.Printf("Invalid choice: %s", choice)
			}
		}
	}
}

func createWorkflow(client *http.Client, addr, basePath string) {
	// Example workflow schema
	workflow := models.WorkflowSchema{
		Name:        "test-workflow",
		Description: "Test workflow created by client",
		Activities: []models.ActivityDefinition{
			{
				ID:   "test-activity",
				Name: "test-activity",
				Type: "test",
				Retry: &models.RetryPolicy{
					MaxAttempts:        3,
					InitialInterval:    5 * time.Second,
					MaxInterval:        30 * time.Second,
					BackoffCoefficient: 2.0,
				},
			},
		},
		DAG: map[string][]string{
			"test-activity": {},
		},
	}

	url := fmt.Sprintf("http://%s%s/v1/workflows", addr, basePath)
	body, err := json.Marshal(workflow)
	if err != nil {
		log.Printf("Failed to marshal workflow: %v", err)
		return
	}

	resp, err := client.Post(url, "application/json", strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Failed to create workflow: %v", err)
		return
	}
	defer resp.Body.Close()

	var response models.WorkflowSchema
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return
	}

	log.Printf("Created workflow: ID=%s, Name=%s", response.ID.Hex(), response.Name)
}

func getWorkflow(client *http.Client, addr, basePath string) {
	fmt.Print("Enter workflow ID: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}
	id := scanner.Text()

	url := fmt.Sprintf("http://%s%s/v1/workflows/%s", addr, basePath, id)
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Failed to get workflow: %v", err)
		return
	}
	defer resp.Body.Close()

	var workflow models.WorkflowSchema
	if err := json.NewDecoder(resp.Body).Decode(&workflow); err != nil {
		log.Printf("Failed to decode workflow: %v", err)
		return
	}

	log.Printf("Workflow: %+v", workflow)
}

func triggerWorkflow(client *http.Client, addr, basePath string) {
	fmt.Print("Enter workflow ID: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}
	id := scanner.Text()

	url := fmt.Sprintf("http://%s%s/v1/workflows/%s/trigger", addr, basePath, id)
	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		log.Printf("Failed to trigger workflow: %v", err)
		return
	}
	defer resp.Body.Close()

	var response models.WorkflowRequest
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return
	}

	log.Printf("Triggered workflow: RequestID=%s", response.ID.Hex())
}

func getWorkflowRequest(client *http.Client, addr, basePath string) {
	fmt.Print("Enter request ID: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}
	id := scanner.Text()

	url := fmt.Sprintf("http://%s%s/v1/workflows/request/%s", addr, basePath, id)
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Failed to get workflow request: %v", err)
		return
	}
	defer resp.Body.Close()

	var request models.WorkflowRequest
	if err := json.NewDecoder(resp.Body).Decode(&request); err != nil {
		log.Printf("Failed to decode request: %v", err)
		return
	}

	log.Printf("Workflow request: %+v", request)
}

func getActivityResults(client *http.Client, addr, basePath string) {
	fmt.Print("Enter request ID: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}
	id := scanner.Text()

	url := fmt.Sprintf("http://%s%s/v1/workflows/request/%s/activities", addr, basePath, id)
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Failed to get activity results: %v", err)
		return
	}
	defer resp.Body.Close()

	var results []models.ActivityResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		log.Printf("Failed to decode results: %v", err)
		return
	}

	log.Printf("Activity results: %+v", results)
}

func runWebSocketTest(addr, wsPath, mode string, count int) {
	url := fmt.Sprintf("ws://%s%s", addr, wsPath)
	log.Printf("Connecting to WebSocket at %s", url)

	dialer := websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  45 * time.Second,
		EnableCompression: true,
	}

	headers := http.Header{}
	headers.Add("User-Agent", "FlowPilotX-Client/1.0")
	headers.Add("Sec-WebSocket-Protocol", "flowpilotx-v1")

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

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
	mode = "echo"
	switch mode {
	case "ping":
		go runPingTest(ctx, c, interrupt)
	case "echo":
		go runEchoTest(ctx, c, interrupt)
	case "load":
		go runLoadTest(ctx, c, interrupt, count)
	}

	select {
	case <-interrupt:
		log.Println("Received interrupt signal")
	case <-done:
		log.Println("Connection closed")
	}

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

				wsMessage := models.WebSocketMessage{
					EventID: uuid.New().String(),
					Message: message,
				}

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
			wsMessage := models.WebSocketMessage{
				EventID: uuid.New().String(),
				Message: fmt.Sprintf("load test message %d", i+1),
			}

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

	var healthResp HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
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

	var versionResp VersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&versionResp); err != nil {
		return nil, err
	}

	return &versionResp, nil
}
