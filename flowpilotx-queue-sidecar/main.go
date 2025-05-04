package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flowpilotx-queue-sidecar/internal/config"
	"flowpilotx-queue-sidecar/internal/queue"
	"flowpilotx-queue-sidecar/internal/worker"

	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Initialize queue client
	queueClient, err := queue.NewClient(cfg.Queue)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize queue client")
	}
	defer queueClient.Close()

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: time.Duration(cfg.HTTP.TimeoutSeconds) * time.Second,
	}

	// Create worker
	worker := worker.NewWorker(queueClient, httpClient, cfg.HTTP, logger)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker
	go worker.Start(ctx)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down...")
	cancel()
	time.Sleep(2 * time.Second) // Give time for cleanup
}
