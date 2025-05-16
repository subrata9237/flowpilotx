package model

import (
	"time"
)

// WorkerConfig defines configuration for a worker pool
type WorkerConfig struct {
	NumWorkers         int      `json:"num_workers"`
	MaxQueuesPerWorker int      `json:"max_queues_per_worker"`
	InitialQueues      []string `json:"initial_queues"`
	HeartbeatInterval  int      `json:"heartbeat_interval"`

	// New configuration options
	QueueBalancingStrategy string            `json:"queue_balancing_strategy"` // "least_loaded", "round_robin", "weighted"
	WorkerStartupDelay     time.Duration     `json:"worker_startup_delay"`     // Delay between starting workers
	GracefulShutdownTime   time.Duration     `json:"graceful_shutdown_time"`   // Time to wait for graceful shutdown
	QueueRebalanceInterval time.Duration     `json:"queue_rebalance_interval"` // Interval for queue rebalancing
	HealthCheck            HealthCheckConfig `json:"health_check"`
}

// HealthCheckConfig defines health check configuration
type HealthCheckConfig struct {
	Enabled          bool          `json:"enabled"`
	Interval         time.Duration `json:"interval"`
	Timeout          time.Duration `json:"timeout"`
	FailureThreshold int           `json:"failure_threshold"`
	SuccessThreshold int           `json:"success_threshold"`
}

// WorkerStatus represents the current status of a worker
type WorkerStatus struct {
	WorkerID      int       `json:"worker_id"`
	ActiveQueues  []string  `json:"active_queues"`
	IsActive      bool      `json:"is_active"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	HealthStatus  string    `json:"health_status"`
	ErrorCount    int       `json:"error_count"`
	LastError     string    `json:"last_error"`
	RestartCount  int       `json:"restart_count"`
}
