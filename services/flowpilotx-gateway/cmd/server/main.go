package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/libs/mongodb"
	"github.com/flowpilotx/libs/worker/workerhandler"
	_ "github.com/flowpilotx/services/flowpilotx-gateway/docs" // swagger docs
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/routes"
	"go.temporal.io/sdk/client"
)

// Swagger documentation annotations
// @title           FlowPilotX Gateway API
// @version         1.0
// @description     REST API for FlowPilotX Gateway service
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.flowpilotx.io/support
// @contact.email  support@flowpilotx.io
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8080
// @BasePath  /api/flowpilotx-gateway
// @schemes   http
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @tag.name Health
// @tag.description Health check endpoints
// @tag.name Version
// @tag.description Version information endpoints
// @tag.name Workflows
// @tag.description Workflow management endpoints

const (
	serviceName = "flowpilotx-gateway"
)

// main is the entry point of the application
func main() {
	// Load configuration
	cfg, err := config.LoadConfigAuto(serviceName, "GATEWAY")
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := initializeLogger(cfg)
	log.Info("Starting FlowPilotX Gateway service", map[string]interface{}{
		"service": serviceName,
		"env":     cfg.Service.Environment,
	})

	// Initialize MongoDB client
	mongoClient, err := initializeMongoDB(cfg, log)
	if err != nil {
		log.Error("Failed to initialize MongoDB", map[string]interface{}{
			"error": err.Error(),
			"uri":   cfg.MongoDB.URI,
		})
		os.Exit(1)
	}
	defer func() {
		if err := mongoClient.Close(); err != nil {
			log.Error("Failed to close MongoDB connection", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()
	log.Info("MongoDB connection established", map[string]interface{}{
		"database": cfg.MongoDB.Database,
	})
	// Initialize Temporal client
	temporalClient, err := initializeTemporalClient(cfg, log)
	if err != nil {
		log.Error("Failed to initialize Temporal client", map[string]interface{}{
			"error":     err.Error(),
			"hostPort":  cfg.FlowpilotxGateway.Temporal.HostPort,
			"namespace": cfg.FlowpilotxGateway.Temporal.Namespace,
		})
		os.Exit(1)
	}
	defer func() {
		temporalClient.Close()
		log.Info("Temporal client connection closed", map[string]interface{}{})
	}()
	log.Info("Temporal client initialized", map[string]interface{}{
		"namespace": cfg.FlowpilotxGateway.Temporal.Namespace,
	})

	// Initialize workflow worker
	workflowWorker := workerhandler.NewWorkerHandler(temporalClient, mongoClient, log, &cfg.FlowpilotxWorker.Worker)

	// Start workflow worker in a goroutine
	workerCtx, workerCancel := context.WithCancel(context.Background())
	go func() {
		if err := workflowWorker.Start(workerCtx); err != nil {
			log.Error("Workflow worker failed", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()
	defer workflowWorker.Stop()
	defer workerCancel()
	// Initialize and configure router
	router := routes.New(cfg, log, temporalClient, mongoClient)
	muxRouter := router.Init()
	log.Info("Router initialized")

	// Create server
	srv := createServer(cfg, muxRouter)

	// Start server and handle shutdown
	startServer(srv, cfg, log)
}

// startServer starts the HTTP server and handles graceful shutdown
func startServer(srv *http.Server, cfg *config.Config, log logger.LoggerInterface) {
	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the service listening for requests
	go func() {
		log.Info("Server starting", map[string]interface{}{
			"address":       srv.Addr,
			"read_timeout":  cfg.FlowpilotxGateway.Server.ReadTimeout,
			"write_timeout": cfg.FlowpilotxGateway.Server.WriteTimeout,
			"idle_timeout":  cfg.FlowpilotxGateway.Server.IdleTimeout,
		})
		log.Info("Swagger Documentation", map[string]interface{}{
			"url": fmt.Sprintf("http://localhost:%d%s/swagger/index.html",
				cfg.FlowpilotxGateway.Server.Port,
				cfg.FlowpilotxGateway.API.BasePath),
		})
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown
	select {
	case err := <-serverErrors:
		log.Error("Server error", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	case sig := <-shutdown:
		log.Info("Shutdown signal received", map[string]interface{}{
			"signal": sig.String(),
		})

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		// Gracefully shutdown the server
		if err := srv.Shutdown(ctx); err != nil {
			log.ErrorWithCtx(ctx, "Server shutdown error", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
		log.InfoWithCtx(ctx, "Server shutdown complete")
	}
}

// initializeLogger creates and returns a new logger instance
func initializeLogger(cfg *config.Config) logger.LoggerInterface {
	log := logger.NewLogger(logger.ServiceLogConfig{
		Level:       cfg.Logging.Level,
		Format:      cfg.Logging.Format,
		Environment: cfg.Service.Environment,
		ServiceName: serviceName,
	})
	return log
}

// initializeMongoDB creates and returns a new MongoDB client
func initializeMongoDB(cfg *config.Config, log logger.LoggerInterface) (*mongodb.Client, error) {
	mongoConfig := &mongodb.Config{
		URI:              cfg.MongoDB.URI,
		Database:         cfg.MongoDB.Database,
		Username:         cfg.MongoDB.Username,
		Password:         cfg.MongoDB.Password,
		ConnectTimeout:   time.Duration(cfg.MongoDB.Options.ConnectTimeoutMS) * time.Millisecond,
		OperationTimeout: time.Duration(cfg.MongoDB.Options.SocketTimeoutMS) * time.Millisecond,
		MaxPoolSize:      uint64(cfg.MongoDB.Options.MaxPoolSize),
		MinPoolSize:      uint64(cfg.MongoDB.Options.MinPoolSize),
		RetryWrites:      cfg.MongoDB.Options.RetryWrites,
		RetryReads:       cfg.MongoDB.Options.RetryReads,
		Direct:           false,
	}
	return mongodb.NewClient(mongoConfig)
}

// initializeTemporalClient creates and returns a new Temporal client
func initializeTemporalClient(cfg *config.Config, log logger.LoggerInterface) (client.Client, error) {
	options := client.Options{
		HostPort:  cfg.FlowpilotxGateway.Temporal.HostPort,
		Namespace: cfg.FlowpilotxGateway.Temporal.Namespace,
		Identity:  cfg.FlowpilotxGateway.Temporal.Client.Identity,
	}
	log.Info("Initializing Temporal client", map[string]interface{}{
		"options": options,
	})
	return client.Dial(options)
}

// createServer creates and returns a new HTTP server
func createServer(cfg *config.Config, router http.Handler) *http.Server {
	addr := fmt.Sprintf("%s:%d", cfg.FlowpilotxGateway.Server.Host, cfg.FlowpilotxGateway.Server.Port)
	return &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.FlowpilotxGateway.Server.ReadTimeout,
		WriteTimeout: cfg.FlowpilotxGateway.Server.WriteTimeout,
		IdleTimeout:  cfg.FlowpilotxGateway.Server.IdleTimeout,
	}
}
