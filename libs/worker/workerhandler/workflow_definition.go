package workerhandler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/flowpilotx/libs/worker/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func (w *WorkflowWorker) DynamicWorkflow(ctx workflow.Context, input map[string]interface{}) (map[string]interface{}, error) {
	w.logger.Info("Starting dynamic workflow", map[string]interface{}{
		"input": input,
	})

	response := &model.WorkflowResponse{}

	// Validation logs
	w.logger.Info("Validating workflow inputs", nil)
	//print input
	w.logger.Info("Input", map[string]interface{}{
		"input": input,
	})
	// Extract input parameters with validation
	workflowName, ok := input["name"].(string)
	if !ok {
		w.logger.Error("Invalid workflow name", map[string]interface{}{
			"input":       input,
			"actual_type": fmt.Sprintf("%T", input["name"]),
		})
		response.Status = "FAILED"
		response.Error = "workflow_name is required"
		result, err := w.structToMap(response, fmt.Errorf("workflow_name is required"))
		return result, err
	}
	workflowID, ok := input["id"].(string)
	if !ok {
		w.logger.Error("Invalid workflow ID", map[string]interface{}{
			"input":       input,
			"actual_type": fmt.Sprintf("%T", input["id"]),
		})
		response.Status = "FAILED"
		response.Error = "workflow_id is required"
		result, err := w.structToMap(response, fmt.Errorf("workflow_id is required"))
		return result, err
	}

	// Convert string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(workflowID)
	if err != nil {
		w.logger.Error("Invalid workflow ID format", map[string]interface{}{
			"workflow_id": workflowID,
			"error":       err.Error(),
		})
		response.Status = "FAILED"
		response.Error = "invalid workflow_id format"
		result, err := w.structToMap(response, fmt.Errorf("invalid workflow_id format: %v", err))
		return result, err
	}
	response.ID = objectID

	requestID, ok := input["request_id"].(string)
	if !ok {
		w.logger.Error("Invalid request ID", map[string]interface{}{
			"input":       input,
			"actual_type": fmt.Sprintf("%T", input["request_id"]),
		})
		response.Status = "FAILED"
		response.Error = "request_id is required"
		result, err := w.structToMap(response, fmt.Errorf("request_id is required"))
		return result, err
	}

	// Convert string to ObjectID
	requestObjectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		w.logger.Error("Invalid request ID format", map[string]interface{}{
			"request_id": requestID,
			"error":      err.Error(),
		})
		response.Status = "FAILED"
		response.Error = "invalid request_id format"
		result, err := w.structToMap(response, fmt.Errorf("invalid request_id format: %v", err))
		return result, err
	}
	response.RequestID = requestObjectID

	var workflowDef *model.WorkflowSchema
	err = w.mapToStruct(input, &workflowDef)
	if err != nil {
		w.logger.Error("Failed to unmarshal workflow input", map[string]interface{}{
			"workflow_input": input,
			"error":          err.Error(),
		})
		response.Status = "FAILED"
		response.Error = fmt.Sprintf("failed to unmarshal workflow input: %v", err)
		result, err := w.structToMap(response, fmt.Errorf("failed to unmarshal workflow input: %v", err))
		return result, err
	}
	//print workflowDef
	w.logger.Info("Workflow definition", map[string]interface{}{
		"workflow_def": workflowDef,
	})
	results := make(map[string]interface{})
	response.Activities = make([]model.ActivityResult, 0)
	response.Status = "STARTED"
	response.CreatedAt = workflow.Now(ctx)
	response.UpdatedAt = workflow.Now(ctx)
	response.Version = workflowDef.Version
	response.DAG = workflowDef.DAG

	// Activity execution logs
	w.logger.Info("Starting activity processing", map[string]interface{}{
		"total_activities": len(workflowDef.Activities),
		"dag":              workflowDef.DAG,
	})

	// Process each activity based on DAG
	for id, activity := range workflowDef.Activities {
		activity.Workflow = workflowDef
		w.logger.Info("Processing activity", map[string]interface{}{
			"activity_id":   id,
			"activity_name": activity.Name,
			"activity_type": activity.Type,
			"is_async":      activity.Config != nil && activity.Config["async"] == true,
		})

		// Dependency check logs
		if deps, ok := workflowDef.DAG[strconv.Itoa(id)]; ok && len(deps) > 0 {
			w.logger.Info("Checking activity dependencies", map[string]interface{}{
				"activity_id":  id,
				"dependencies": deps,
			})
		}

		// Check dependencies
		if deps, ok := workflowDef.DAG[strconv.Itoa(id)]; ok {
			for _, dep := range deps {
				if _, exists := results[dep]; !exists {
					w.logger.Info("Waiting for dependency", map[string]interface{}{
						"activity_id": id,
						"dependency":  dep,
					})
					workflow.Sleep(ctx, time.Second)
				}
			}
		}
		activityInput, err := w.structToMap(activity, nil)
		if err != nil {
			w.logger.Error("Failed to convert activity to map", map[string]interface{}{
				"error":    err.Error(),
				"activity": activity,
			})
		}
		// Configure activity options
		activityOptions := workflow.ActivityOptions{
			StartToCloseTimeout: 30 * time.Minute, // Default timeout
			HeartbeatTimeout:    30 * time.Second, // For async activities
		}

		// Add retry policy if configured
		if activity.Retry != nil {
			activityOptions.RetryPolicy = &temporal.RetryPolicy{
				InitialInterval:    activity.Retry.InitialInterval,
				BackoffCoefficient: activity.Retry.BackoffCoefficient,
				MaximumInterval:    activity.Retry.MaxInterval,
				MaximumAttempts:    int32(activity.Retry.MaxAttempts),
			}
		}

		// Determine if activity is async
		isAsync := activity.Config != nil && activity.Config["async"] == true

		// Execute activity based on type
		var result interface{}
		activityCtx := workflow.WithActivityOptions(ctx, activityOptions)
		now := workflow.Now(ctx)
		if isAsync {
			w.logger.Info("Executing async activity", map[string]interface{}{
				"activity_id":   id,
				"activity_name": activity.Name,
				"retry_policy":  activity.Retry != nil,
			})
			// Execute async activity with progress monitoring
			future := workflow.ExecuteActivity(activityCtx, w.ExecuteAsyncActivity, activity.Type, activityInput)

			// Monitor progress while waiting for completion
			for {
				if future.IsReady() {
					break
				}

				var progress ActivityProgress
				if err := future.Get(ctx, &progress); err == nil {
					w.logger.Info("Async activity progress", map[string]interface{}{
						"activity_id":   id,
						"activity_name": activity.Name,
						"status":        progress.Status,
						"percentage":    progress.Percentage,
						"stage":         progress.CurrentStage,
					})
				}

				workflow.Sleep(ctx, 3*time.Second)
			}

			// Get final result
			if err := future.Get(ctx, &result); err != nil {
				w.logger.Error("Async activity failed", map[string]interface{}{
					"activity_id":   id,
					"activity_name": activity.Name,
					"error":         err.Error(),
					"activity_type": activity.Type,
				})
				finalTime := workflow.Now(ctx)
				activityResult := model.ActivityResult{}
				activityResult.Error = err.Error()
				activityResult.Result = nil
				activityResult.StartedAt = &now
				activityResult.CompletedAt = &finalTime
				activityResult.Attempt = 1
				response.Activities = append(response.Activities, activityResult)
				response.Status = "FAILED"
				result, err := w.structToMap(response, err)
				return result, err
			}
		} else {
			w.logger.Info("Executing sync activity", map[string]interface{}{
				"activity_id":   id,
				"activity_name": activity.Name,
				"retry_policy":  activity.Retry != nil,
			})
			activityResult := model.ActivityResult{}
			// Execute sync activity
			err := workflow.ExecuteActivity(
				activityCtx,
				w.ExecuteActivity,
				activity.Type,
				activityInput,
			).Get(ctx, &result)

			if err != nil {
				w.logger.Error("Sync activity failed", map[string]interface{}{
					"activity_id":   id,
					"activity_name": activity.Name,
					"error":         err.Error(),
					"activity_type": activity.Type,
				})
				finalTime := workflow.Now(ctx)
				activityResult.Error = err.Error()
				activityResult.Result = nil
				activityResult.StartedAt = &now
				activityResult.CompletedAt = &finalTime
				activityResult.Attempt = 1
				response.Activities = append(response.Activities, activityResult)
				response.Status = "FAILED"
				result, err := w.structToMap(response, err)
				return result, err
			}
		}

		// Save activity result
		//workflow.ExecuteActivity(ctx, "SaveActivityResult", requestID, strconv.Itoa(id), result, nil)
		results[strconv.Itoa(id)] = result

		w.logger.Info("Activity completed successfully", map[string]interface{}{
			"activity_id":    id,
			"activity_name":  activity.Name,
			"execution_time": time.Since(now),
			"result_size":    len(fmt.Sprintf("%v", result)),
		})
		finalTime := workflow.Now(ctx)
		// For each activity completion, add to activityResults
		activityResult := model.ActivityResult{
			ActivityID:  strconv.Itoa(id),
			Result:      results[strconv.Itoa(id)].(map[string]interface{}),
			StartedAt:   &now,
			CompletedAt: &finalTime,
			Attempt:     1, // Set appropriate attempt count
		}
		response.Activities = append(response.Activities, activityResult)
	}
	response.Status = "COMPLETED"
	// Update workflow request with final results
	//workflow.ExecuteActivity(ctx, "UpdateWorkflowRequest", requestID, "COMPLETED", results, nil)

	// Final workflow completion log
	w.logger.Info("Dynamic workflow completed", map[string]interface{}{
		"workflow_name":        workflowName,
		"workflow_id":          workflowID,
		"request_id":           requestID,
		"total_activities":     len(workflowDef.Activities),
		"completed_activities": len(response.Activities),
		"status":               response.Status,
		"has_error":            response.Error != "",
	})

	result, err := w.structToMap(response, nil)
	if err != nil {
		w.logger.Error("Failed to convert final response to map", map[string]interface{}{
			"error":    err.Error(),
			"response": fmt.Sprintf("%+v", response),
		})
	}
	return result, err
}

// ActivityProgress struct for tracking async activity progress
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

func (w *WorkflowWorker) structToMap(data interface{}, errInput error) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		w.logger.Error("Failed to marshal struct to JSON", map[string]interface{}{
			"error": err.Error(),
			"data":  fmt.Sprintf("%+v", data),
		})
		return nil, fmt.Errorf("failed to marshal struct to JSON: %v", err)
	}
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		w.logger.Error("Failed to unmarshal JSON to map", map[string]interface{}{
			"error": err.Error(),
			"json":  string(jsonBytes),
		})
	}
	return result, errInput
}

// Convert map to struct using JSON as intermediate
func (w *WorkflowWorker) mapToStruct(data map[string]interface{}, result interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		w.logger.Error("Failed to marshal map to JSON", map[string]interface{}{
			"error": err.Error(),
			"data":  fmt.Sprintf("%+v", data),
		})
		return fmt.Errorf("failed to marshal map to JSON: %v", err)
	}

	if err := json.Unmarshal(jsonBytes, result); err != nil {
		w.logger.Error("Failed to unmarshal JSON to struct", map[string]interface{}{
			"error": err.Error(),
			"json":  string(jsonBytes),
		})
		return fmt.Errorf("failed to unmarshal JSON to struct: %v", err)
	}

	return nil
}
