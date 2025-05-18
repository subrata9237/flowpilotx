package model

import (
	"time"
)

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
