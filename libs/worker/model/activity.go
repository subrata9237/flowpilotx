package model

import "time"

// Add ActivityProgress struct
type ActivityProgress struct {
	Status         string    `json:"status"`
	Percentage     float64   `json:"percentage"`
	StartTime      time.Time `json:"start_time"`
	LastUpdate     time.Time `json:"last_update"`
	CompletionTime time.Time `json:"completion_time,omitempty"`
	CurrentStage   string    `json:"current_stage,omitempty"`
	Message        string    `json:"message,omitempty"`
	Error          string    `json:"error,omitempty"`
}
