package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/grpc-common/pkg/client"
)

const (
	serviceName = "flowpilotx_worker"
	envPrefix   = "WORKER"
)

var (
	// Command-line flags
	serverAddrs      = flag.String("servers", "", "Comma-separated list of server addresses (host:port)")
	mode             = flag.String("mode", "sync", "Execution mode: sync, async, or stream")
	message          = flag.String("message", "Hello Worker!", "Message to send")
	count            = flag.Int("count", 1, "Number of messages to send in stream mode")
	timeout          = flag.Duration("timeout", 30*time.Second, "Operation timeout")
	useLoadBalancing = flag.Bool("lb", false, "Enable load balancing")
	healthCheck      = flag.Bool("health", true, "Enable health checking")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfigAuto(serviceName, envPrefix)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Get server addresses
	var addresses []string
	if *serverAddrs != "" {
		addresses = strings.Split(*serverAddrs, ",")
	} else {
		// Use config address as default
		addresses = []string{fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)}
	}

	// Create client options
	opts := client.DefaultClientOptions().
		WithTimeout(*timeout)

	if len(addresses) > 1 || *useLoadBalancing {
		opts.WithAddresses(addresses).
			WithLoadBalancing(true, "round_robin").
			WithHealthCheck(*healthCheck)
	} else {
		opts.WithAddress(addresses[0])
	}

	// Create worker client
	workerClient, err := client.NewWorkerClient(opts)
	if err != nil {
		log.Fatalf("Failed to create worker client: %v", err)
	}
	defer workerClient.Close()

	log.Printf("Connected to worker service with %d server(s)", len(addresses))
	if opts.LoadBalancing.Enabled {
		log.Printf("Load balancing enabled with policy: %s", opts.LoadBalancing.Policy)
		if opts.LoadBalancing.HealthCheck {
			log.Printf("Health checking enabled")
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()

	// Execute based on mode
	switch *mode {
	case "sync":
		executeSyncRequest(ctx, workerClient)
	case "async":
		executeAsyncRequest(ctx, workerClient)
	case "stream":
		executeStreamRequest(ctx, workerClient)
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}

func executeSyncRequest(ctx context.Context, client *client.WorkerClient) {
	log.Println("Executing synchronous request...")

	resp, err := client.Execute(ctx, []byte(*message))
	if err != nil {
		log.Fatalf("Sync execution failed: %v", err)
	}

	log.Printf("Sync execution response: %+v", resp)
}

func executeAsyncRequest(ctx context.Context, client *client.WorkerClient) {
	log.Println("Executing asynchronous request...")

	resp, err := client.ExecuteAsync(ctx, []byte(*message))
	if err != nil {
		log.Fatalf("Async execution failed: %v", err)
	}
	log.Printf("Async execution started: %+v", resp)

	result, err := client.WaitForAsyncCompletion(ctx, resp, *timeout)
	if err != nil {
		log.Fatalf("Failed waiting for async completion: %v", err)
	}

	log.Printf("Async execution completed: %+v", result)
}

func executeStreamRequest(ctx context.Context, client *client.WorkerClient) {
	log.Println("Executing streaming request...")

	stream, err := client.StreamExecute(ctx)
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}

	// Prepare messages
	messages := make([][]byte, *count)
	for i := 0; i < *count; i++ {
		messages[i] = []byte(fmt.Sprintf("%s (message %d)", *message, i+1))
	}

	// Send messages with rate limiting
	if err := client.SendStreamMessages(stream, messages, time.Second); err != nil {
		log.Fatalf("Failed to send stream messages: %v", err)
	}
}
