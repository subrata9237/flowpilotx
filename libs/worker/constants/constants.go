package constants

import "time"

const (
	HealthStatusHealthy    = "healthy"
	HealthStatusUnhealthy  = "unhealthy"
	HealthStatusRestarting = "restarting"
	MaxErrorCount          = 3
	HealthCheckInterval    = 30 * time.Second
	HeartbeatTimeout       = 60 * time.Second
)
