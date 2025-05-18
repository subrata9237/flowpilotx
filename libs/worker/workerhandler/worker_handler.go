package workerhandler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/libs/mongodb"
	activity "github.com/flowpilotx/libs/worker/activity"
	"github.com/flowpilotx/libs/worker/constants"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type WorkerHandler struct {
	temporalClient    client.Client
	mongoClient       *mongodb.Client
	logger            logger.LoggerInterface
	flowPilotXWorkers []*FlowPilotXWorker
	config            *config.WorkerConfig

	mu sync.RWMutex
}
type FlowPilotXWorker struct {
	worker         worker.Worker
	WorkerID       int
	Name           string
	Queue          string
	IsActive       bool
	LastHeartbeat  time.Time
	HealthStatus   string
	ErrorCount     int
	LastError      string
	RestartCount   int
	mu             sync.RWMutex
	logger         logger.LoggerInterface
	queueConfig    *config.QueueConfig
	temporalClient client.Client
	mongoClient    *mongodb.Client
}

// NewWorkflowWorker creates a new workflow worker pool
func NewWorkerHandler(temporalClient client.Client, mongoClient *mongodb.Client, logger logger.LoggerInterface, cfg *config.WorkerConfig) *WorkerHandler {
	logger.Info("Initializing new worker handler", map[string]interface{}{
		"default_queue":   "flowpilotx-default-queue",
		"config_provided": cfg != nil,
	})

	defaultConfig := &config.WorkerConfig{
		WorkflowQueues: []config.QueueConfig{
			{
				QueueName:                          "flowpilotx-default-queue",
				MaxConcurrentActivityExecutionSize: 100,
				MaxConcurrentWorkflowExecutionSize: 50,
				Weight:                             1,
				WorkerActivitiesPerSecond:          100,
				TaskQueueActivitiesPerSecond:       100,
			},
		},
		HeartbeatInterval: 30,
	}

	if cfg != nil {
		logger.Debug("Using provided config instead of default", map[string]interface{}{
			"workflow_queues_count": len(cfg.WorkflowQueues),
			"activity_queues_count": len(cfg.ActivityQueues),
		})
		defaultConfig = cfg
	} else {
		//print config
		logger.Debug("Using default config", map[string]interface{}{
			"config": cfg,
		})
	}

	workerHandler := &WorkerHandler{
		temporalClient: temporalClient,
		mongoClient:    mongoClient,
		logger:         logger,
		config:         defaultConfig,
	}

	workerHandler.initializeDefaultWorker()

	logger.Info("Created new workflow worker pool", map[string]interface{}{
		"num_of_queues": defaultConfig.WorkflowQueues,
	})

	return workerHandler
}

// initializeWorkerStatus initializes the worker status array
func (w *WorkerHandler) initializeDefaultWorker() {
	w.logger.Info("Starting worker initialization", map[string]interface{}{
		"workflow_queues": len(w.config.WorkflowQueues),
		"activity_queues": len(w.config.ActivityQueues),
	})

	totalQueueConfig := make([]config.QueueConfig, 0)
	totalQueueConfig = append(totalQueueConfig, w.config.WorkflowQueues...)
	totalQueueConfig = append(totalQueueConfig, w.config.ActivityQueues...)

	w.logger.Info("Creating worker instances", map[string]interface{}{
		"total_queues": len(totalQueueConfig),
	})

	w.flowPilotXWorkers = make([]*FlowPilotXWorker, len(totalQueueConfig))

	for i, qConfig := range totalQueueConfig {
		w.logger.Info("Initializing worker", map[string]interface{}{
			"worker_id":                 i,
			"queue_name":                qConfig.QueueName,
			"max_concurrent_activities": qConfig.MaxConcurrentActivityExecutionSize,
			"max_concurrent_workflows":  qConfig.MaxConcurrentWorkflowExecutionSize,
		})

		w.flowPilotXWorkers[i] = &FlowPilotXWorker{
			WorkerID:       i,
			Name:           fmt.Sprintf("flowpilotx-worker-%d-%s", i, qConfig.QueueName),
			Queue:          qConfig.QueueName,
			IsActive:       false,
			LastHeartbeat:  time.Now(),
			HealthStatus:   constants.HealthStatusHealthy,
			ErrorCount:     0,
			LastError:      "",
			RestartCount:   0,
			logger:         w.logger,
			queueConfig:    &qConfig,
			temporalClient: w.temporalClient,
			mongoClient:    w.mongoClient,
		}
	}

	w.logger.Info("Worker initialization completed", map[string]interface{}{
		"total_workers":   len(w.flowPilotXWorkers),
		"workflow_queues": len(w.config.WorkflowQueues),
		"activity_queues": len(w.config.ActivityQueues),
	})
}

// startWorker starts a single worker with its assigned queues
func (w *WorkerHandler) Start(ctx context.Context) error {
	w.logger.Info("Starting all workers", map[string]interface{}{
		"worker_count": len(w.flowPilotXWorkers),
	})

	for i, worker := range w.flowPilotXWorkers {
		w.logger.Debug("Starting individual worker", map[string]interface{}{
			"worker_index": i,
			"worker_id":    worker.WorkerID,
			"queue":        worker.Queue,
		})

		if err := worker.startWorker(ctx); err != nil {
			w.logger.Error("Failed to start worker", map[string]interface{}{
				"worker_index": i,
				"worker_id":    worker.WorkerID,
				"queue":        worker.Queue,
				"error":        err.Error(),
			})
			return err
		}
	}

	w.logger.Info("All workers started successfully", map[string]interface{}{
		"total_workers": len(w.flowPilotXWorkers),
	})
	return nil
}

// startWorker starts a single worker with its assigned queues
func (w *FlowPilotXWorker) startWorker(ctx context.Context) error {
	if ctx.Err() != nil {
		w.logger.Error("Context cancelled before starting worker", map[string]interface{}{
			"worker_id": w.WorkerID,
			"queue":     w.Queue,
			"error":     ctx.Err().Error(),
		})
		return fmt.Errorf("context cancelled before starting worker: %v", ctx.Err())
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.logger.Info("Configuring worker", map[string]interface{}{
		"worker_id":                 w.WorkerID,
		"queue":                     w.Queue,
		"max_concurrent_activities": w.queueConfig.MaxConcurrentActivityExecutionSize,
		"max_concurrent_workflows":  w.queueConfig.MaxConcurrentWorkflowExecutionSize,
		"activities_per_second":     w.queueConfig.WorkerActivitiesPerSecond,
	})

	// Create worker options with context-aware settings
	options := worker.Options{
		MaxConcurrentActivityExecutionSize:     w.queueConfig.MaxConcurrentActivityExecutionSize,
		MaxConcurrentWorkflowTaskExecutionSize: w.queueConfig.MaxConcurrentWorkflowExecutionSize,
		WorkerActivitiesPerSecond:              float64(w.queueConfig.WorkerActivitiesPerSecond),
		TaskQueueActivitiesPerSecond:           float64(w.queueConfig.TaskQueueActivitiesPerSecond),
		EnableLoggingInReplay:                  true,
		BackgroundActivityContext:              ctx, // Use the provided context for background activities
	}

	// Create new worker
	w.worker = worker.New(w.temporalClient, w.Queue, options)

	// Register workflow and activities
	w.worker.RegisterWorkflow(w.FlowpilotxWorkflow)
	w.logger.Debug("Registered workflow", map[string]interface{}{
		"worker_id": w.WorkerID,
		"queue":     w.Queue,
	})

	activities := &activity.Activity{
		Logger:         w.logger,
		MongoClient:    w.mongoClient,
		TemporalClient: w.temporalClient,
	}
	w.worker.RegisterActivity(activities)
	w.logger.Debug("Registered activities", map[string]interface{}{
		"worker_id": w.WorkerID,
		"queue":     w.Queue,
	})

	// Start worker with context monitoring
	errChan := make(chan error, 1)
	go func() {
		if err := w.worker.Start(); err != nil {
			errChan <- err
		}
	}()

	// Monitor for context cancellation or worker start error
	select {
	case <-ctx.Done():
		w.worker.Stop()
		return fmt.Errorf("worker stopped due to context cancellation: %v", ctx.Err())
	case err := <-errChan:
		if err != nil {
			w.IsActive = false
			w.HealthStatus = constants.HealthStatusUnhealthy
			w.LastError = err.Error()
			w.ErrorCount++

			w.logger.Error("Failed to start worker", map[string]interface{}{
				"worker_id": w.WorkerID,
				"queue":     w.Queue,
				"error":     err.Error(),
				"context":   ctx.Value("request_id"),
			})
			return fmt.Errorf("failed to start worker: %v", err)
		}
	}

	w.IsActive = true
	w.HealthStatus = constants.HealthStatusHealthy
	w.LastHeartbeat = time.Now()

	w.logger.Info("Worker started successfully", map[string]interface{}{
		"worker_id": w.WorkerID,
		"queue":     w.Queue,
		"status":    w.HealthStatus,
		"context":   ctx.Value("request_id"),
	})

	w.logger.Debug("Starting worker execution", map[string]interface{}{
		"worker_id": w.WorkerID,
		"queue":     w.Queue,
	})

	return nil
}

// Stop stops all workers in the pool
func (w *WorkerHandler) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.logger.Info("Initiating worker pool shutdown", map[string]interface{}{
		"total_workers": len(w.flowPilotXWorkers),
	})

	for i, workerInstance := range w.flowPilotXWorkers {
		if workerInstance != nil {
			w.logger.Debug("Stopping worker", map[string]interface{}{
				"worker_index": i,
				"worker_id":    workerInstance.WorkerID,
				"queue":        workerInstance.Queue,
			})
			workerInstance.stop()
		}
	}

	w.logger.Info("Worker pool shutdown completed", map[string]interface{}{
		"stopped_workers": len(w.flowPilotXWorkers),
	})
}

// Stop stops the worker safely
func (w *FlowPilotXWorker) stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.logger.Info("Initiating worker shutdown", map[string]interface{}{
		"worker_id":     w.WorkerID,
		"queue":         w.Queue,
		"is_active":     w.IsActive,
		"health_status": w.HealthStatus,
	})

	if w.worker != nil && w.IsActive {
		defer func() {
			if r := recover(); r != nil {
				w.logger.Error("Worker stop panic recovered", map[string]interface{}{
					"worker_id": w.WorkerID,
					"queue":     w.Queue,
					"panic":     r,
				})
			}
		}()

		w.worker.Stop()
		w.IsActive = false
		w.worker = nil

		w.logger.Info("Worker stopped successfully", map[string]interface{}{
			"worker_id": w.WorkerID,
			"queue":     w.Queue,
		})
	} else {
		w.logger.Debug("Worker already stopped", map[string]interface{}{
			"worker_id":  w.WorkerID,
			"queue":      w.Queue,
			"was_active": w.IsActive,
		})
	}

	w.logger.Info("Worker shutdown process completed", map[string]interface{}{
		"worker_id":    w.WorkerID,
		"queue":        w.Queue,
		"final_status": w.HealthStatus,
	})
}
