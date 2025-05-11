package models

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// VersionResponse represents the version information response
type VersionResponse struct {
	Version     string `json:"version"`
	Environment string `json:"environment"`
	ServiceName string `json:"service_name"`
}
