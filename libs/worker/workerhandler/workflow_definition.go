package workerhandler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/flowpilotx/libs/worker/model"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func (w *FlowPilotXWorker) FlowpilotxWorkflow(ctx workflow.Context, input *model.Workflow) (*model.Workflow, error) {
	if input == nil || input.WorkflowSchema == nil {
		return nil, fmt.Errorf("invalid workflow input: workflow or schema is nil")
	}
	input.Status = "running"
	now := time.Now()
	input.CreatedAt = now
	input.UpdatedAt = now
	input.StartTime = &now
	input.EndTime = nil
	input.Duration = nil

	w.logger.Info("Starting workflow with metadata", map[string]interface{}{
		"workflow_id": input.WorkflowSchema.ID.Hex(),
		"name":        input.WorkflowSchema.Name,
		"category":    input.WorkflowSchema.Category,
		"team":        input.WorkflowSchema.Team,
		"priority":    input.WorkflowSchema.Priority,
		"version":     input.WorkflowSchema.Version,
		"parameters":  input.Parameters,
	})

	// Initialize schema resolver
	resolver := NewSchemaResolver(w.logger)
	results := make(map[string]*model.ActivityDefinition)
	executed := make(map[string]bool)

	// Add activity processing metadata log before the loop
	w.logger.Info("Processing workflow activities", map[string]interface{}{
		"total_activities": len(input.WorkflowSchema.Activities),
		"workflow_id":      input.WorkflowSchema.ID.Hex(),
		"dag":              input.WorkflowSchema.DAG,
	})

	// Process activities in DAG order
	for _, activity := range input.WorkflowSchema.Activities {
		activityID := activity.ID

		// Add detailed activity start log
		w.logger.Info("Starting activity processing", map[string]interface{}{
			"activity_id":   activityID,
			"activity_name": activity.Name,
			"activity_type": activity.Type,
			"is_optional":   activity.IsOptional,
			"config":        activity.Config,
			"workflow_id":   input.WorkflowSchema.ID.Hex(),
		})

		// Validate activity configuration
		if activity.Config == nil {
			return nil, fmt.Errorf("activity %s has no configuration", activityID)
		}

		if activity.Name == "" || activity.Type == "" {
			return nil, fmt.Errorf("activity %s missing required fields: name or type", activityID)
		}

		now := time.Now()
		activity.StartTime = &now
		activity.EndTime = nil
		activity.Duration = nil
		activity.Attempt = 0
		activity.Error = ""
		activity.Status = "running"

		// Add queue assignment log
		w.logger.Info("Activity queue assignment", map[string]interface{}{
			"activity_id":    activityID,
			"queue":          activity.Queue,
			"workflow_queue": input.WorkflowSchema.Queue,
		})

		// Configure activity options
		activityOptions := workflow.ActivityOptions{
			StartToCloseTimeout: activity.Config.Timeout.StartToClose,
			HeartbeatTimeout:    activity.Config.Timeout.Heartbeat,
			//TaskQueue:           activity.Queue,
		}

		// Add activity options log
		w.logger.Info("Configuring activity options", map[string]interface{}{
			"activity_id":       activityID,
			"timeout":           activity.Config.Timeout,
			"retry_policy":      activity.Retry,
			"start_to_close":    activity.Config.Timeout.StartToClose,
			"heartbeat_timeout": activity.Config.Timeout.Heartbeat,
		})

		// Validate timeouts
		if activityOptions.StartToCloseTimeout <= 0 {
			return nil, fmt.Errorf("invalid start to close timeout for activity %s", activityID)
		}

		if activity.Retry != nil {
			if activity.Retry.MaxAttempts <= 0 {
				return nil, fmt.Errorf("invalid retry attempts for activity %s", activityID)
			}
			activityOptions.RetryPolicy = &temporal.RetryPolicy{
				InitialInterval:    activity.Retry.InitialInterval,
				BackoffCoefficient: activity.Retry.BackoffCoefficient,
				MaximumInterval:    activity.Retry.MaxInterval,
				MaximumAttempts:    int32(activity.Retry.MaxAttempts),
			}
		}

		activityCtx := workflow.WithActivityOptions(ctx, activityOptions)

		// Log activity context details
		w.logger.Info("Activity context details", map[string]interface{}{
			"activity_id": activityID,
			"options": map[string]interface{}{
				"task_queue":        activityOptions.TaskQueue,
				"start_to_close":    activityOptions.StartToCloseTimeout.String(),
				"heartbeat_timeout": activityOptions.HeartbeatTimeout.String(),
				"retry_policy": map[string]interface{}{
					"initial_interval":    activityOptions.RetryPolicy.InitialInterval.String(),
					"backoff_coefficient": activityOptions.RetryPolicy.BackoffCoefficient,
					"maximum_interval":    activityOptions.RetryPolicy.MaximumInterval.String(),
					"maximum_attempts":    activityOptions.RetryPolicy.MaximumAttempts,
				},
			},
			"workflow_id": input.WorkflowSchema.ID.Hex(),
		})

		// Process dependencies
		if deps, ok := input.WorkflowSchema.DAG[activityID]; ok && len(deps) > 0 {
			for _, depID := range deps {
				if !executed[depID] {
					w.logger.Error("dependency %s not executed for activity %s", map[string]interface{}{
						"dependency_id": depID,
						"activity_id":   activityID,
					})
					return w.prepareWorkflowOutput(input, results), fmt.Errorf("dependency %s not executed for activity %s", depID, activityID)
				}

				depIdx, err := strconv.Atoi(depID)
				if err != nil {
					w.logger.Error("invalid dependency ID %s for activity %s: %w", map[string]interface{}{
						"dependency_id": depID,
						"activity_id":   activityID,
						"error":         err.Error(),
					})
					return w.prepareWorkflowOutput(input, results), fmt.Errorf("invalid dependency ID %s for activity %s: %w", depID, activityID, err)
				}

				if depIdx >= len(input.WorkflowSchema.Activities) {
					w.logger.Error("dependency index out of range for activity %s", map[string]interface{}{
						"dependency_id": depID,
						"activity_id":   activityID,
					})
					return w.prepareWorkflowOutput(input, results), fmt.Errorf("dependency index out of range for activity %s", activityID)
				}

				depActivity := input.WorkflowSchema.Activities[depIdx]
				resolver.activities[depID] = &depActivity
			}

			resolvedInputs, err := resolver.ResolveInputSchema(&activity)
			if err != nil {
				w.logger.Error("failed to resolve input schema for activity %s: %w", map[string]interface{}{
					"activity_id": activityID,
					"error":       err.Error(),
				})
				return w.prepareWorkflowOutput(input, results), fmt.Errorf("failed to resolve input schema for activity %s: %w", activityID, err)
			}
			activity.InputSchema = resolvedInputs
		}

		var result model.ActivityDefinition

		// Execute activity
		if activity.Config.Async {
			w.logger.Error("async activities are not supported yet", map[string]interface{}{
				"activity_id": activityID,
			})
			return w.prepareWorkflowOutput(input, results), fmt.Errorf("async activities are not supported yet")
		}
		// Add pre-execution log
		w.logger.Info("Preparing activity execution", map[string]interface{}{
			"activity_id":     activityID,
			"function_name":   activity.Name,
			"input_schema":    activity.InputSchema,
			"current_attempt": activity.Attempt,
		})
		w.logger.Info("Activity function name", map[string]interface{}{
			"activity_id":   activityID,
			"function_name": activity.Name,
		})
		err := workflow.ExecuteActivity(
			activityCtx,
			activity.Name,
			activity,
		).Get(ctx, &result)

		if err != nil {
			w.logger.Error("failed to execute activity %s: %w", map[string]interface{}{
				"activity_id": activityID,
				"error":       err.Error(),
			})
			return w.prepareWorkflowOutput(input, results), fmt.Errorf("failed to execute activity %s: %w", activityID, err)
		}
		now = time.Now()
		result.EndTime = &now
		duration := now.Sub(*result.StartTime)
		result.Duration = &duration
		result.Status = "completed"
		result.Error = ""
		results[activityID] = &result
		executed[activityID] = true
		activity = result
		resolver.activities[activityID] = &activity

		// Add detailed completion metrics
		w.logger.Info("Activity execution metrics", map[string]interface{}{
			"activity_id": activityID,
			"start_time":  result.StartTime,
			"end_time":    result.EndTime,
			"duration":    result.Duration,
			"status":      result.Status,
			"attempt":     result.Attempt,
		})

		w.logger.Info("Activity completed", map[string]interface{}{
			"activity_id": activityID,
			"type":        activity.Type,
			"name":        activity.Name,
			"result":      result,
		})
	}

	// Add final workflow metrics before completion
	w.logger.Info("Workflow execution metrics", map[string]interface{}{
		"workflow_id":          input.WorkflowSchema.ID.Hex(),
		"total_duration":       input.Duration,
		"activities_completed": len(results),
		"start_time":           input.StartTime,
		"end_time":             input.EndTime,
		"status":               input.Status,
	})

	return w.prepareWorkflowOutput(input, results), nil
}

func (w *FlowPilotXWorker) prepareWorkflowOutput(input *model.Workflow, results map[string]*model.ActivityDefinition) *model.Workflow {
	// Add preparation start log
	w.logger.Info("Preparing workflow output", map[string]interface{}{
		"workflow_id":    input.WorkflowSchema.ID.Hex(),
		"num_results":    len(results),
		"num_activities": len(input.WorkflowSchema.Activities),
	})

	// Update workflow status and timing
	now := time.Now()
	input.Status = "completed"
	input.EndTime = &now
	input.UpdatedAt = now
	duration := now.Sub(input.CreatedAt)
	input.Duration = &duration

	// Create activities map for easier lookup
	activitiesMap := make(map[string]*model.ActivityDefinition)
	for _, activity := range input.WorkflowSchema.Activities {
		id := activity.ID
		activitiesMap[id] = &activity
	}

	// Update activities with their results
	for id, result := range results {
		if activity, exists := activitiesMap[id]; exists {
			*activity = *result
		} else {
			w.logger.Error("Activity not found", map[string]interface{}{
				"activity_id": id,
			})
		}
	}

	// Add activity results summary
	w.logger.Info("Activity results summary", map[string]interface{}{
		"workflow_id":    input.WorkflowSchema.ID.Hex(),
		"results_map":    results,
		"activities_map": activitiesMap,
	})

	w.logger.Info("Workflow completed", map[string]interface{}{
		"workflow_id": input.WorkflowSchema.ID.Hex(),
		"duration":    duration.String(),
		"activities":  len(results),
	})

	return input
}
