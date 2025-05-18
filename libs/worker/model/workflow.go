package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkflowSchema struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Queue       string               `bson:"queue" json:"queue"`
	Name        string               `bson:"name" json:"name"`
	Category    string               `bson:"category" json:"category"`
	Team        string               `bson:"team" json:"team"`
	Priority    string               `bson:"priority" json:"priority"`
	Type        string               `bson:"type" json:"type"`
	Description string               `bson:"description" json:"description"`
	Activities  []ActivityDefinition `bson:"activities" json:"activities"`
	DAG         map[string][]string  `bson:"dag" json:"dag"`
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updated_at"`
	Status      string               `bson:"status" json:"status"`
	Version     int                  `bson:"version" json:"version"`
}

type ActivityDefinition struct {
	ID   string `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
	Type string `json:"type" bson:"type"`
	//make queue as omitempty
	Queue             string                 `json:"queue,omitempty" bson:"queue,omitempty"`
	InputSchema       map[string]interface{} `json:"input_schema" bson:"input_schema"`
	OutputSchema      map[string]interface{} `json:"output_schema" bson:"output_schema"`
	Config            *ActivityConfig        `json:"config" bson:"config"`
	Retry             *RetryPolicy           `json:"retry,omitempty" bson:"retry,omitempty"`
	WorkflowID        *primitive.ObjectID    `bson:"workflow_id,omitempty" json:"workflow_id,omitempty"`
	WorkFlowRequestID *primitive.ObjectID    `json:"work_flow_request_id,omitempty" bson:"work_flow_request_id,omitempty"`
	WorkflowName      string                 `bson:"workflow_name,omitempty" json:"workflow_name,omitempty"`
	StartTime         *time.Time             `json:"start_time,omitempty" bson:"start_time,omitempty"`
	EndTime           *time.Time             `json:"end_time,omitempty" bson:"end_time,omitempty"`
	Duration          *time.Duration         `json:"duration,omitempty" bson:"duration,omitempty"`
	Attempt           int                    `json:"attempt" bson:"attempt"`
	Error             string                 `json:"error,omitempty" bson:"error,omitempty"`
	Status            string                 `json:"status" bson:"status"`
	IsOptional        bool                   `json:"is_optional" bson:"is_optional,omitempty"`
}
type ActivityConfig struct {
	Async   bool            `json:"async" bson:"async"`
	Timeout ActivityTimeout `json:"timeout" bson:"timeout"`
}

type ActivityTimeout struct {
	StartToClose    time.Duration `json:"start_to_close" bson:"start_to_close"`
	ScheduleToStart time.Duration `json:"schedule_to_start" bson:"schedule_to_start"`
	ScheduleToClose time.Duration `json:"schedule_to_close" bson:"schedule_to_close"`
	Heartbeat       time.Duration `json:"heartbeat" bson:"heartbeat"`
}

type RetryPolicy struct {
	MaxAttempts        int           `bson:"max_attempts" json:"max_attempts"`
	InitialInterval    time.Duration `bson:"initial_interval" json:"initial_interval"`
	MaxInterval        time.Duration `bson:"max_interval" json:"max_interval"`
	BackoffCoefficient float64       `bson:"backoff_coefficient" json:"backoff_coefficient"`
}

type Workflow struct {
	WorkflowSchema *WorkflowSchema    `json:"workflow_schema" bson:"workflow_schema"`
	RequestID      primitive.ObjectID `json:"request_id" bson:"request_id"`
	//Workflow status meas wheasther workflwo succesfully compleyted or not
	Status string `json:"status" bson:"status"`
	// Timestamps
	CreatedAt time.Time      `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" bson:"updated_at"`
	StartTime *time.Time     `json:"start_time,omitempty" bson:"start_time,omitempty"`
	EndTime   *time.Time     `json:"end_time,omitempty" bson:"end_time,omitempty"`
	Duration  *time.Duration `json:"duration,omitempty" bson:"duration,omitempty"`
	// Custom input parameters
	Parameters map[string]interface{} `json:"parameters" bson:"parameters"`
}

// Add reference types
type Reference struct {
	Ref string `json:"$ref"`
}
