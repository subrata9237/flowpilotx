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
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/handlers"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/middleware"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/websocket"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/time/rate"
)

const serviceName = "flowpilotx-gateway"

func main() {
	// Load configuration
	cfg, err := config.LoadConfigAuto("flowpilotx_gateway", "GATEWAY")
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	//print cfg
	// Initialize logger using the library's setup function
	log := logger.NewLogger(logger.ServiceLogConfig{
		Level:       cfg.Logging.Level,
		Format:      cfg.Logging.Format,
		Environment: cfg.Service.Environment,
		ServiceName: serviceName,
	})

	log.Info(context.Background(), "Starting gateway service", map[string]interface{}{
		"version":     cfg.FlowpilotxGateway.Service.Version,
		"environment": cfg.FlowpilotxGateway.Service.Environment,
	})

	// Create main router
	router := mux.NewRouter()

	// Add logging middleware
	router.Use(middleware.LoggingMiddleware(log))

	// Configure CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   cfg.FlowpilotxGateway.CORS.AllowedOrigins,
		AllowedMethods:   append(cfg.FlowpilotxGateway.CORS.AllowedMethods, "GET", "OPTIONS"),
		AllowedHeaders:   append(cfg.FlowpilotxGateway.CORS.AllowedHeaders, "Sec-WebSocket-Protocol", "Sec-WebSocket-Version", "Sec-WebSocket-Key", "Sec-WebSocket-Extensions", "Upgrade", "Connection"),
		MaxAge:           cfg.FlowpilotxGateway.CORS.MaxAge,
		AllowCredentials: true,
		ExposedHeaders:   []string{"Sec-WebSocket-Accept"},
	})
	router.Use(corsHandler.Handler)

	// Configure rate limiter if enabled
	if cfg.FlowpilotxGateway.RateLimit.Enabled {
		limiter := rate.NewLimiter(
			rate.Limit(cfg.FlowpilotxGateway.RateLimit.RequestsPerSecond),
			cfg.FlowpilotxGateway.RateLimit.Burst,
		)
		router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !limiter.Allow() {
					http.Error(w, "Too many requests", http.StatusTooManyRequests)
					return
				}
				next.ServeHTTP(w, r)
			})
		})
		log.Info(context.Background(), "Rate limiter enabled", map[string]interface{}{
			"requests_per_second": cfg.FlowpilotxGateway.RateLimit.RequestsPerSecond,
			"burst":              cfg.FlowpilotxGateway.RateLimit.Burst,
		})
	}

	// Create API subrouter with base path
	api := router.PathPrefix(cfg.FlowpilotxGateway.API.BasePath).Subrouter()

	// Add health check endpoint
	api.HandleFunc(cfg.FlowpilotxGateway.API.HealthCheckPath, handlers.NewHealthHandler(log)).Methods("GET")

	// Add version endpoint
	versionHandler := handlers.NewVersionHandler(
		log,
		cfg.FlowpilotxGateway.Service.Version,
		cfg.FlowpilotxGateway.Service.Environment,
		cfg.FlowpilotxGateway.Service.Name,
	)
	api.HandleFunc(cfg.FlowpilotxGateway.API.VersionPath, versionHandler).Methods("GET")

	// Configure WebSocket handler
	wsPath := cfg.FlowpilotxGateway.WebSocket.Path
	wsConfig := websocket.WebSocketConfig{
		ReadBufferSize:   cfg.FlowpilotxGateway.WebSocket.ReadBufferSize,
		WriteBufferSize:  cfg.FlowpilotxGateway.WebSocket.WriteBufferSize,
		HandshakeTimeout: cfg.FlowpilotxGateway.WebSocket.HandshakeTimeout,
		PingInterval:     cfg.FlowpilotxGateway.WebSocket.PingInterval,
		PongWait:         cfg.FlowpilotxGateway.WebSocket.PongWait,
		WriteTimeout:     cfg.FlowpilotxGateway.Server.WriteTimeout,
	}
	wsHandler := websocket.NewHandler(log, wsConfig)
	router.Handle(wsPath, wsHandler)

	log.Info(context.Background(), "WebSocket configured", map[string]interface{}{
		"path":             wsPath,
		"ping_interval":    wsConfig.PingInterval,
		"pong_wait":        wsConfig.PongWait,
		"read_buffer":      wsConfig.ReadBufferSize,
		"write_buffer":     wsConfig.WriteBufferSize,
		"write_timeout":    wsConfig.WriteTimeout,
		"handshake_timeout": wsConfig.HandshakeTimeout,
	})

	// Create server
	addr := fmt.Sprintf("%s:%d", cfg.FlowpilotxGateway.Server.Host, cfg.FlowpilotxGateway.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.FlowpilotxGateway.Server.ReadTimeout,
		WriteTimeout: cfg.FlowpilotxGateway.Server.WriteTimeout,
		IdleTimeout:  cfg.FlowpilotxGateway.Server.IdleTimeout,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Info(context.Background(), "Server starting", map[string]interface{}{
			"address":        addr,
			"read_timeout":   cfg.FlowpilotxGateway.Server.ReadTimeout,
			"write_timeout":  cfg.FlowpilotxGateway.Server.WriteTimeout,
			"idle_timeout":   cfg.FlowpilotxGateway.Server.IdleTimeout,
		})
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
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