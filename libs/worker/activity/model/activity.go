package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

type ActivityResult struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	RequestID  primitive.ObjectID     `bson:"request_id" json:"request_id"`
	ActivityID string                 `bson:"activity_id" json:"activity_id"`
	Status     string                 `bson:"status" json:"status"`
	Output     map[string]interface{} `bson:"output" json:"output"`
	Error      string                 `bson:"error,omitempty" json:"error,omitempty"`
	StartTime  *time.Time             `bson:"start_time" json:"start_time"`
	EndTime    *time.Time             `bson:"end_time" json:"end_time"`
	Duration   *time.Duration         `bson:"duration" json:"duration"`
}
