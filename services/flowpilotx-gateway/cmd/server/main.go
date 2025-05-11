package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/logger"
	_ "github.com/flowpilotx/services/flowpilotx-gateway/docs" // swagger docs
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/routes"
)

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
// @schemes   http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @tag.name Health
// @tag.description Health check endpoints

// @tag.name Version
// @tag.description Version information endpoints

const (
	serviceName = "flowpilotx-gateway"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfigAuto(serviceName, "GATEWAY")
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := initializeLogger(cfg)

	// Initialize and configure router
	router := routes.New(cfg, log)
	muxRouter := router.Init()

	// Create server
	srv := createServer(cfg, muxRouter)

	// Start server and handle shutdown
	startServer(srv, cfg, log)
}

func initializeLogger(cfg *config.Config) logger.LoggerInterface {
	return logger.NewLogger(logger.ServiceLogConfig{
		Level:       cfg.Logging.Level,
		Format:      cfg.Logging.Format,
		Environment: cfg.Service.Environment,
		ServiceName: serviceName,
	})
}

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

func startServer(srv *http.Server, cfg *config.Config, log logger.LoggerInterface) {
	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the service listening for requests
	go func() {
		log.Info(context.Background(), "Server starting", map[string]interface{}{
			"address":       srv.Addr,
			"read_timeout":  cfg.FlowpilotxGateway.Server.ReadTimeout,
			"write_timeout": cfg.FlowpilotxGateway.Server.WriteTimeout,
			"idle_timeout":  cfg.FlowpilotxGateway.Server.IdleTimeout,
		})
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown
	select {
	case err := <-serverErrors:
		log.Error(context.Background(), "Server error", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	case sig := <-shutdown:
		log.Info(context.Background(), "Shutdown signal received", map[string]interface{}{
			"signal": sig.String(),
		})

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		// Gracefully shutdown the server
		if err := srv.Shutdown(ctx); err != nil {
			log.Error(ctx, "Server shutdown error", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
	}
}
