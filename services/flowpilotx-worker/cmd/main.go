package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/flowpilotx/libs/config"
	pb "github.com/flowpilotx/libs/grpc-common/pkg/api/worker/v1"
	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/services/flowpilotx-worker/internal/metrics"
	"github.com/flowpilotx/services/flowpilotx-worker/internal/service"
)

const (
	serviceName = "flowpilotx_worker"
	envPrefix   = "WORKER"
)

func main() {
	ctx := context.Background()

	// Load configuration using auto-discovery
	cfg, err := config.LoadConfigAuto(serviceName, envPrefix)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger using the library's setup function
	log, err := logger.SetupServiceLogger(logger.ServiceLogConfig{
		Level:       cfg.Logging.Level,
		Format:      cfg.Logging.Format,
		Environment: cfg.Service.Environment,
		ServiceName: serviceName,
	})
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Log service startup information
	log.Info(ctx, "Starting service", map[string]interface{}{
		"service":     cfg.Service.Name,
		"version":     cfg.Service.Version,
		"env":         cfg.Service.Environment,
		"log_level":   cfg.Logging.Level,
		"log_format":  cfg.Logging.Format,
		"server_host": cfg.Server.Host,
		"server_port": cfg.Server.Port,
	})

	// Initialize metrics
	m := metrics.New(log)
	if err := m.StartServer(&cfg.Metrics); err != nil {
		log.Error(ctx, "Failed to start metrics server", map[string]interface{}{
			"error":        err.Error(),
			"metrics_port": cfg.Metrics.Port,
		})
		os.Exit(1)
	}

	// Create gRPC server with options
	serverOpts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: cfg.Server.MaxConnectionIdle,
			Time:              cfg.Server.KeepAliveTime,
			Timeout:           cfg.Server.KeepAliveTimeout,
		}),
	}

	grpcServer := grpc.NewServer(serverOpts...)

	// Create worker service with the same logger instance
	workerService := service.NewWorkerService(cfg, log, m)
	pb.RegisterWorkerServiceServer(grpcServer, workerService)
	reflection.Register(grpcServer)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	if err != nil {
		log.Error(ctx, "Failed to listen", map[string]interface{}{
			"error":   err.Error(),
			"address": fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		})
		os.Exit(1)
	}

	go func() {
		log.Info(ctx, "Starting gRPC server", map[string]interface{}{
			"address": fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		})
		if err := grpcServer.Serve(lis); err != nil {
			log.Error(ctx, "Failed to serve", map[string]interface{}{
				"error":   err.Error(),
				"address": fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			})
			os.Exit(1)
		}
	}()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh

	log.Info(ctx, "Received shutdown signal", map[string]interface{}{
		"signal": sig.String(),
	})

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
	defer cancel()

	// Stop metrics server (if running)
	if cfg.Metrics.Enabled {
		if err := m.StopServer(shutdownCtx); err != nil {
			log.Error(ctx, "Failed to stop metrics server", map[string]interface{}{
				"error":   err.Error(),
				"timeout": cfg.Server.ShutdownTimeout.String(),
			})
		}
	}

	// Gracefully stop gRPC server
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-shutdownCtx.Done():
		log.Warn(ctx, "Shutdown timeout exceeded, forcing stop", map[string]interface{}{
			"timeout": cfg.Server.ShutdownTimeout.String(),
		})
		grpcServer.Stop()
	case <-stopped:
		log.Info(ctx, "Server stopped gracefully", map[string]interface{}{
			"shutdown_duration": time.Since(time.Now()).String(),
		})
	}
}
