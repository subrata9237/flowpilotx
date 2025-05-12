package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkflowSchema struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name        string               `bson:"name" json:"name"`
	Description string               `bson:"description" json:"description"`
	Activities  []ActivityDefinition `bson:"activities" json:"activities"`
	DAG         map[string][]string  `bson:"dag" json:"dag"`
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updated_at"`
	Status      string               `bson:"status" json:"status"`
	Version     int                  `bson:"version" json:"version"`
}

type ActivityDefinition struct {
	ID           string                 `bson:"id" json:"id"`
	Name         string                 `bson:"name" json:"name"`
	Type         string                 `bson:"type" json:"type"`
	InputSchema  map[string]interface{} `bson:"input_schema" json:"input_schema"`
	OutputSchema map[string]interface{} `bson:"output_schema" json:"output_schema"`
	Config       map[string]interface{} `bson:"config" json:"config"`
	Retry        *RetryPolicy           `bson:"retry,omitempty" json:"retry,omitempty"`
}

type RetryPolicy struct {
	MaxAttempts        int           `bson:"max_attempts" json:"max_attempts"`
	InitialInterval    time.Duration `bson:"initial_interval" json:"initial_interval"`
	MaxInterval        time.Duration `bson:"max_interval" json:"max_interval"`
	BackoffCoefficient float64       `bson:"backoff_coefficient" json:"backoff_coefficient"`
}

type WorkflowRequest struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	WorkflowID  primitive.ObjectID     `bson:"workflow_id" json:"workflow_id"`
	Input       map[string]interface{} `bson:"input" json:"input"`
	Status      string                 `bson:"status" json:"status"`
	Result      map[string]interface{} `bson:"result" json:"result"`
	Error       string                 `bson:"error,omitempty" json:"error,omitempty"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
	StartedAt   *time.Time             `bson:"started_at,omitempty" json:"started_at,omitempty"`
	CompletedAt *time.Time             `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

type ActivityResult struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RequestID   primitive.ObjectID `bson:"request_id" json:"request_id"`
	ActivityID  string             `bson:"activity_id" json:"activity_id"`
	Result      interface{}        `bson:"result" json:"result"`
	Error       string             `bson:"error,omitempty" json:"error,omitempty"`
	StartedAt   time.Time          `bson:"started_at" json:"started_at"`
	CompletedAt *time.Time         `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	Attempt     int                `bson:"attempt" json:"attempt"`
}
