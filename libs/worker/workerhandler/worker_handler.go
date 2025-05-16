package workerhandler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/libs/mongodb"
	"github.com/flowpilotx/libs/worker/constants"
	"github.com/flowpilotx/libs/worker/model"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// WorkflowWorker handles workflow and activity execution
type WorkflowWorker struct {
	temporalClient client.Client
	mongoClient    *mongodb.Client
	logger         logger.LoggerInterface
	activities     *WorkflowActivities

	workers      []worker.Worker
	workerStatus []model.WorkerStatus
	workerQueues map[int][]string
	activeQueues map[string]bool
	config       model.WorkerConfig

	mu sync.RWMutex
}

// NewWorkflowWorker creates a new workflow worker pool
func NewWorkflowWorker(temporalClient client.Client, mongoClient *mongodb.Client, logger logger.LoggerInterface, config *model.WorkerConfig) *WorkflowWorker {
	defaultConfig := model.WorkerConfig{
		NumWorkers:         5,
		MaxQueuesPerWorker: 5,
		InitialQueues:      []string{"workflow-task-queue"},
		HeartbeatInterval:  30,
	}

	if config != nil {
		defaultConfig = *config
	}

	worker := &WorkflowWorker{
		temporalClient: temporalClient,
		mongoClient:    mongoClient,
		logger:         logger,
		activities:     NewWorkflowActivities(mongoClient, logger),
		workerQueues:   make(map[int][]string),
		activeQueues:   make(map[string]bool),
		config:         defaultConfig,
	}

	worker.initializeWorkerStatus()

	logger.Info("Created new workflow worker pool", map[string]interface{}{
		"num_workers":           defaultConfig.NumWorkers,
		"max_queues_per_worker": defaultConfig.MaxQueuesPerWorker,
		"initial_queues":        defaultConfig.InitialQueues,
	})

	return worker
}

// initializeWorkerStatus initializes the worker status array
func (w *WorkflowWorker) initializeWorkerStatus() {
	w.workerStatus = make([]model.WorkerStatus, w.config.NumWorkers)
	for i := 0; i < w.config.NumWorkers; i++ {
		w.workerStatus[i] = model.WorkerStatus{
			WorkerID:      i,
			ActiveQueues:  make([]string, 0),
			IsActive:      false,
			LastHeartbeat: time.Now(),
		}
	}
}

// Start starts the worker pool
func (w *WorkflowWorker) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.workers = make([]worker.Worker, w.config.NumWorkers)

	// Distribute initial queues
	if err := w.distributeQueues(w.config.InitialQueues); err != nil {
		return fmt.Errorf("failed to distribute queues: %v", err)
	}

	// Start workers
	for workerIndex := 0; workerIndex < w.config.NumWorkers; workerIndex++ {
		queues := w.workerQueues[workerIndex]
		if len(queues) == 0 {
			continue
		}

		if err := w.startWorker(ctx, workerIndex, queues); err != nil {
			return fmt.Errorf("failed to start worker %d: %v", workerIndex, err)
		}
	}

	// Start health check and heartbeat monitoring
	go w.startHealthCheck(ctx)
	go w.startHeartbeatMonitor(ctx)

	return nil
}

// distributeQueues distributes queues among workers
func (w *WorkflowWorker) distributeQueues(queues []string) error {
	if len(queues) == 0 {
		return fmt.Errorf("no queues to distribute")
	}

	workerIndex := 0
	for _, queue := range queues {
		if w.activeQueues[queue] {
			continue
		}

		// Find next worker with capacity
		for len(w.workerQueues[workerIndex]) >= w.config.MaxQueuesPerWorker {
			workerIndex = (workerIndex + 1) % w.config.NumWorkers
		}

		w.workerQueues[workerIndex] = append(w.workerQueues[workerIndex], queue)
		w.activeQueues[queue] = true
		w.workerStatus[workerIndex].ActiveQueues = append(w.workerStatus[workerIndex].ActiveQueues, queue)
	}

	return nil
}

// startWorker starts a single worker with its assigned queues
func (w *WorkflowWorker) startWorker(ctx context.Context, workerIndex int, queues []string) error {
	primaryQueue := queues[0]
	w.logger.InfoWithCtx(ctx, "Starting worker", map[string]interface{}{
		"worker_index":  workerIndex,
		"primary_queue": primaryQueue,
		"all_queues":    queues,
	})

	workerInstance := worker.New(w.temporalClient, primaryQueue, worker.Options{})

	// Register workflow and activities
	workerInstance.RegisterWorkflow(w.DynamicWorkflow)
	workerInstance.RegisterActivity(w.ExecuteActivity)
	workerInstance.RegisterActivity(w.ExecuteAsyncActivity)

	// Start worker
	if err := workerInstance.Start(); err != nil {
		w.workerStatus[workerIndex].HealthStatus = constants.HealthStatusUnhealthy
		w.workerStatus[workerIndex].LastError = err.Error()
		w.workerStatus[workerIndex].ErrorCount++
		return fmt.Errorf("failed to start worker: %v", err)
	}

	w.workers[workerIndex] = workerInstance
	w.workerStatus[workerIndex].IsActive = true
	w.workerStatus[workerIndex].LastHeartbeat = time.Now()
	w.workerStatus[workerIndex].HealthStatus = constants.HealthStatusHealthy

	return nil
}

// startHeartbeatMonitor starts monitoring worker heartbeats
func (w *WorkflowWorker) startHeartbeatMonitor(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(w.config.HeartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.updateHeartbeats()
		}
	}
}

// startHealthCheck starts the health check routine
func (w *WorkflowWorker) startHealthCheck(ctx context.Context) {
	w.logger.Info("Starting health check monitor")

	ticker := time.NewTicker(constants.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.checkWorkersHealth(ctx)
		}
	}
}

// checkWorkersHealth checks the health of all workers and restarts unhealthy ones
func (w *WorkflowWorker) checkWorkersHealth(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	for i, workerStatus := range w.workerStatus {
		if !workerStatus.IsActive {
			continue
		}

		isHealthy := true
		var reason string

		// Check heartbeat timeout
		if now.Sub(workerStatus.LastHeartbeat) > constants.HeartbeatTimeout {
			isHealthy = false
			reason = "heartbeat timeout"
		}

		// Check error count
		if workerStatus.ErrorCount >= constants.MaxErrorCount {
			isHealthy = false
			reason = "max errors exceeded"
		}

		if !isHealthy {
			w.logger.WarnWithCtx(ctx, "Unhealthy worker detected", map[string]interface{}{
				"worker_id":      workerStatus.WorkerID,
				"reason":         reason,
				"last_heartbeat": workerStatus.LastHeartbeat,
				"error_count":    workerStatus.ErrorCount,
			})

			// Attempt to restart the worker
			if err := w.restartWorker(ctx, i); err != nil {
				w.logger.ErrorWithCtx(ctx, "Failed to restart worker", map[string]interface{}{
					"error":     err.Error(),
					"worker_id": workerStatus.WorkerID,
				})
			}
		}
	}
}

// restartWorker restarts a specific worker
func (w *WorkflowWorker) restartWorker(ctx context.Context, workerIndex int) error {
	w.logger.InfoWithCtx(ctx, "Attempting to restart worker", map[string]interface{}{
		"worker_id": workerIndex,
	})

	// Stop the existing worker if it's still running
	if w.workers[workerIndex] != nil {
		w.workers[workerIndex].Stop()
	}

	// Update worker status
	w.workerStatus[workerIndex].HealthStatus = constants.HealthStatusRestarting
	w.workerStatus[workerIndex].RestartCount++

	// Get the queues that were assigned to this worker
	queues := w.workerQueues[workerIndex]
	if len(queues) == 0 {
		w.logger.WarnWithCtx(ctx, "No queues found for worker", map[string]interface{}{
			"worker_id": workerIndex,
		})
		return nil
	}

	// Start a new worker instance
	if err := w.startWorker(ctx, workerIndex, queues); err != nil {
		w.workerStatus[workerIndex].HealthStatus = constants.HealthStatusUnhealthy
		w.workerStatus[workerIndex].LastError = err.Error()
		w.workerStatus[workerIndex].ErrorCount++
		return fmt.Errorf("failed to restart worker %d: %v", workerIndex, err)
	}

	// Reset error count and update status
	w.workerStatus[workerIndex].HealthStatus = constants.HealthStatusHealthy
	w.workerStatus[workerIndex].ErrorCount = 0
	w.workerStatus[workerIndex].LastError = ""

	w.logger.InfoWithCtx(ctx, "Successfully restarted worker", map[string]interface{}{
		"worker_id":     workerIndex,
		"queues":        queues,
		"restart_count": w.workerStatus[workerIndex].RestartCount,
	})

	return nil
}

// updateHeartbeats updates heartbeats for all active workers
func (w *WorkflowWorker) updateHeartbeats() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i, worker := range w.workers {
		if worker != nil {
			w.workerStatus[i].LastHeartbeat = time.Now()

			// Save heartbeat to MongoDB
			collection := w.mongoClient.Collection("worker_heartbeats")
			opts := options.Update().SetUpsert(true)
			_, err := collection.UpdateOne(
				context.Background(),
				map[string]interface{}{"worker_id": i},
				map[string]interface{}{
					"$set": w.workerStatus[i],
				},
				opts,
			)

			if err != nil {
				w.logger.Error("Failed to update heartbeat", map[string]interface{}{
					"error":     err.Error(),
					"worker_id": i,
				})
				w.workerStatus[i].ErrorCount++
				w.workerStatus[i].LastError = err.Error()
			}
		}
	}
}

// AddQueue adds a new queue to the worker pool
func (w *WorkflowWorker) AddQueue(ctx context.Context, queue string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.logger.InfoWithCtx(ctx, "Starting queue addition process", map[string]interface{}{
		"queue":                 queue,
		"current_active_queues": w.activeQueues,
		"num_workers":           w.config.NumWorkers,
		"max_queues_per_worker": w.config.MaxQueuesPerWorker,
	})

	if w.activeQueues[queue] {
		w.logger.WarnWithCtx(ctx, "Queue already exists", map[string]interface{}{
			"queue": queue,
		})
		return fmt.Errorf("queue %s already exists", queue)
	}

	// Step 1: Try to find a worker under max capacity
	targetWorker := -1
	w.logger.InfoWithCtx(ctx, "Starting worker selection - Phase 1", map[string]interface{}{
		"phase": "under_capacity_search",
	})

	// First try: Look for workers under max capacity
	for i := 0; i < w.config.NumWorkers; i++ {
		currentQueueCount := len(w.workerQueues[i])
		w.logger.DebugWithCtx(ctx, "Checking worker capacity", map[string]interface{}{
			"worker_index":      i,
			"current_queues":    currentQueueCount,
			"max_queues":        w.config.MaxQueuesPerWorker,
			"is_under_capacity": currentQueueCount < w.config.MaxQueuesPerWorker,
		})

		if currentQueueCount < w.config.MaxQueuesPerWorker {
			targetWorker = i
			w.logger.InfoWithCtx(ctx, "Found worker under capacity", map[string]interface{}{
				"worker_index":       targetWorker,
				"current_queues":     currentQueueCount,
				"available_capacity": w.config.MaxQueuesPerWorker - currentQueueCount,
			})
			break
		}
	}

	// Step 2: If no worker under capacity, find least loaded worker
	if targetWorker == -1 {
		w.logger.InfoWithCtx(ctx, "No workers under capacity, starting Phase 2", map[string]interface{}{
			"phase": "least_loaded_search",
		})

		minQueues := len(w.workerQueues[0])
		targetWorker = 0

		// Find the least loaded worker
		for i := 1; i < w.config.NumWorkers; i++ {
			currentQueueCount := len(w.workerQueues[i])
			w.logger.DebugWithCtx(ctx, "Comparing worker loads", map[string]interface{}{
				"worker_index":      i,
				"current_queues":    currentQueueCount,
				"min_queues_so_far": minQueues,
			})

			if currentQueueCount < minQueues {
				minQueues = currentQueueCount
				targetWorker = i
				w.logger.DebugWithCtx(ctx, "Found new least loaded worker", map[string]interface{}{
					"new_target_worker": i,
					"queue_count":       currentQueueCount,
				})
			}
		}

		w.logger.InfoWithCtx(ctx, "Selected least loaded worker", map[string]interface{}{
			"worker_index":    targetWorker,
			"current_queues":  minQueues,
			"will_exceed_max": (minQueues + 1) > w.config.MaxQueuesPerWorker,
		})
	}

	// Step 3: Add queue to selected worker
	w.logger.InfoWithCtx(ctx, "Adding queue to worker", map[string]interface{}{
		"phase":         "queue_assignment",
		"worker_index":  targetWorker,
		"queue":         queue,
		"worker_queues": w.workerQueues[targetWorker],
	})

	w.workerQueues[targetWorker] = append(w.workerQueues[targetWorker], queue)
	w.activeQueues[queue] = true
	w.workerStatus[targetWorker].ActiveQueues = append(w.workerStatus[targetWorker].ActiveQueues, queue)
	w.workerStatus[targetWorker].LastHeartbeat = time.Now()

	// Step 4: Restart worker with all queues
	w.logger.InfoWithCtx(ctx, "Initiating worker restart", map[string]interface{}{
		"phase":        "worker_restart",
		"worker_index": targetWorker,
		"total_queues": len(w.workerQueues[targetWorker]),
		"all_queues":   w.workerQueues[targetWorker],
	})

	if err := w.restartWorker(ctx, targetWorker); err != nil {
		w.logger.ErrorWithCtx(ctx, "Worker restart failed", map[string]interface{}{
			"error":        err.Error(),
			"worker_index": targetWorker,
			"new_queue":    queue,
			"all_queues":   w.workerQueues[targetWorker],
		})
		return fmt.Errorf("failed to restart worker after adding queue: %v", err)
	}

	// Step 5: Calculate and log final metrics
	newQueueCount := len(w.workerQueues[targetWorker])
	workerLoadPercentage := float64(newQueueCount) / float64(w.config.MaxQueuesPerWorker) * 100
	distribution := w.getQueueDistribution()

	w.logger.InfoWithCtx(ctx, "Queue addition completed successfully", map[string]interface{}{
		"phase":              "completion",
		"queue":              queue,
		"worker_index":       targetWorker,
		"final_queues":       w.workerQueues[targetWorker],
		"queue_count":        newQueueCount,
		"load_percentage":    workerLoadPercentage,
		"worker_status":      w.workerStatus[targetWorker],
		"queue_distribution": distribution,
	})

	return nil
}

// getQueueDistribution returns the current distribution of queues across workers
func (w *WorkflowWorker) getQueueDistribution() map[int]struct {
	QueueCount   int      `json:"queue_count"`
	Queues       []string `json:"queues"`
	IsOverMax    bool     `json:"is_over_max"`
	LoadPercent  float64  `json:"load_percent"`
	ExcessQueues int      `json:"excess_queues,omitempty"`
} {
	dist := make(map[int]struct {
		QueueCount   int      `json:"queue_count"`
		Queues       []string `json:"queues"`
		IsOverMax    bool     `json:"is_over_max"`
		LoadPercent  float64  `json:"load_percent"`
		ExcessQueues int      `json:"excess_queues,omitempty"`
	})

	for workerID, queues := range w.workerQueues {
		queueCount := len(queues)
		isOverMax := queueCount > w.config.MaxQueuesPerWorker

		dist[workerID] = struct {
			QueueCount   int      `json:"queue_count"`
			Queues       []string `json:"queues"`
			IsOverMax    bool     `json:"is_over_max"`
			LoadPercent  float64  `json:"load_percent"`
			ExcessQueues int      `json:"excess_queues,omitempty"`
		}{
			QueueCount:   queueCount,
			Queues:       queues,
			IsOverMax:    isOverMax,
			LoadPercent:  float64(queueCount) / float64(w.config.MaxQueuesPerWorker) * 100,
			ExcessQueues: max(0, queueCount-w.config.MaxQueuesPerWorker),
		}
	}

	return dist
}

// RemoveQueue removes a queue from the worker pool
func (w *WorkflowWorker) RemoveQueue(ctx context.Context, queue string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.activeQueues[queue] {
		return fmt.Errorf("queue %s does not exist", queue)
	}

	// Find worker handling this queue
	for workerIndex, queues := range w.workerQueues {
		for i, q := range queues {
			if q == queue {
				// Remove queue
				w.workerQueues[workerIndex] = append(queues[:i], queues[i+1:]...)
				delete(w.activeQueues, queue)

				// Update worker status
				for i, q := range w.workerStatus[workerIndex].ActiveQueues {
					if q == queue {
						w.workerStatus[workerIndex].ActiveQueues = append(
							w.workerStatus[workerIndex].ActiveQueues[:i],
							w.workerStatus[workerIndex].ActiveQueues[i+1:]...,
						)
						break
					}
				}

				w.logger.InfoWithCtx(ctx, "Removed queue", map[string]interface{}{
					"queue":        queue,
					"worker_index": workerIndex,
				})
				return nil
			}
		}
	}

	return fmt.Errorf("queue %s not found in any worker", queue)
}

// Stop stops all workers in the pool
func (w *WorkflowWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.logger.Info("Stopping worker pool", map[string]interface{}{
		"active_queues": w.activeQueues,
	})

	for i, workerInstance := range w.workers {
		if workerInstance != nil {
			workerInstance.Stop()
			w.workers[i] = nil
			w.workerStatus[i].IsActive = false
		}
	}

	w.activeQueues = make(map[string]bool)
	w.workerQueues = make(map[int][]string)

	w.logger.Info("Worker pool stopped successfully")
}

// GetWorkerStatus returns the status of all workers
func (w *WorkflowWorker) GetWorkerStatus() []model.WorkerStatus {
	w.mu.RLock()
	defer w.mu.RUnlock()

	status := make([]model.WorkerStatus, len(w.workerStatus))
	copy(status, w.workerStatus)
	return status
}

// GetActiveQueues returns all active queues
func (w *WorkflowWorker) GetActiveQueues() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	queues := make([]string, 0, len(w.activeQueues))
	for queue := range w.activeQueues {
		queues = append(queues, queue)
	}
	return queues
}

// ExecuteActivity executes a single activity
func (w *WorkflowWorker) ExecuteActivity(ctx context.Context, activityType string, input map[string]interface{}) (interface{}, error) {
	w.logger.InfoWithCtx(ctx, "Executing activity", map[string]interface{}{
		"activity_type": activityType,
		"workflow_id":   input["workflow_id"],
		"request_id":    input["request_id"],
	})

	result, err := w.activities.ExecuteActivity(ctx, input)
	if err != nil {
		w.logger.ErrorWithCtx(ctx, "Activity execution failed", map[string]interface{}{
			"error":         err.Error(),
			"activity_type": activityType,
			"workflow_id":   input["workflow_id"],
		})
		return nil, fmt.Errorf("failed to execute activity: %v", err)
	}

	return result, nil
}

// ExecuteAsyncActivity adds a new ExecuteAsyncActivity method
func (w *WorkflowWorker) ExecuteAsyncActivity(ctx context.Context, activityType string, input map[string]interface{}) error {
	info := activity.GetInfo(ctx)
	asyncToken := info.TaskToken

	w.logger.InfoWithCtx(ctx, "Starting async activity execution", map[string]interface{}{
		"activity_type": activityType,
		"workflow_id":   input["workflow_id"],
		"request_id":    input["request_id"],
	})
	// Start async processing
	go func() {
		asyncCtx := context.Background()

		// Initialize progress tracking
		progress := &model.ActivityProgress{
			Status:     "running",
			Percentage: 0,
			StartTime:  time.Now(),
			LastUpdate: time.Now(),
		}

		// Record initial heartbeat
		w.temporalClient.RecordActivityHeartbeat(asyncCtx, asyncToken, progress)

		// Execute the actual activity
		result, err := w.activities.ExecuteActivity(asyncCtx, input)

		if err != nil {
			w.logger.ErrorWithCtx(asyncCtx, "Async activity execution failed", map[string]interface{}{
				"error":         err.Error(),
				"activity_type": activityType,
				"workflow_id":   input["workflow_id"],
			})

			// Complete activity with error
			w.temporalClient.CompleteActivity(asyncCtx, asyncToken, nil, err)
			return
		}

		// Update final progress
		progress.Status = "completed"
		progress.Percentage = 100
		progress.CompletionTime = time.Now()
		w.temporalClient.RecordActivityHeartbeat(asyncCtx, asyncToken, progress)

		// Complete activity with result
		w.temporalClient.CompleteActivity(asyncCtx, asyncToken, result, nil)

		w.logger.InfoWithCtx(asyncCtx, "Async activity completed successfully", map[string]interface{}{
			"activity_type": activityType,
			"workflow_id":   input["workflow_id"],
			"duration":      progress.CompletionTime.Sub(progress.StartTime).String(),
		})
	}()

	return activity.ErrResultPending
}
