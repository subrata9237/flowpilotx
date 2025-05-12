package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/flowpilotx/services/flowpilotx-gateway/internal/models"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func DynamicWorkflow(ctx workflow.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Extract input parameters with validation
	workflowName, ok := input["workflow_name"].(string)
	if !ok {
		return nil, fmt.Errorf("workflow_name is required")
	}
	workflowID, ok := input["workflow_id"].(string)
	if !ok {
		return nil, fmt.Errorf("workflow_id is required")
	}
	requestID, ok := input["request_id"].(string)
	if !ok {
		return nil, fmt.Errorf("request_id is required")
	}
	workflowInput, ok := input["input"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("input must be a map")
	}

	// Get workflow definition
	var workflowDef *models.WorkflowSchema
	err := workflow.ExecuteActivity(ctx, "GetWorkflowDefinition", workflowID).Get(ctx, &workflowDef)
	if err != nil {
		workflow.GetLogger(ctx).Error("Failed to get workflow definition",
			"error", err.Error(),
			"workflow_name", workflowName,
			"workflow_id", workflowID)
		workflow.ExecuteActivity(ctx, "UpdateWorkflowRequest", requestID, "FAILED", nil, err)
		return nil, err
	}

	// Log workflow start
	workflow.GetLogger(ctx).Info("Starting dynamic workflow",
		"workflow_name", workflowName,
		"workflow_id", workflowID,
		"request_id", requestID)

	results := make(map[string]interface{})

	// Process each activity based on DAG
	for id, activity := range workflowDef.Activities {
		// Check if dependencies are met
		if deps, ok := workflowDef.DAG[strconv.Itoa(id)]; ok {
			for _, dep := range deps {
				// Wait for dependency results
				if _, exists := results[dep]; !exists {
					workflow.GetLogger(ctx).Info("Waiting for dependency",
						"activity_id", id,
						"dependency", dep)
					workflow.Sleep(ctx, time.Second)
				}
			}
		}

		// Execute the activity with enhanced input
		var result interface{}
		activityInput := map[string]interface{}{
			"workflow_name": workflowName,
			"workflow_id":   workflowID,
			"request_id":    requestID,
			"activity_id":   strconv.Itoa(id),
			"activity_name": activity.Name,
			"activity_type": activity.Type,
			"input":         workflowInput,
			"config":        activity.Config,
		}

		// Execute activity with retry policy if configured
		var activityOptions workflow.ActivityOptions
		if activity.Retry != nil {
			activityOptions = workflow.ActivityOptions{
				RetryPolicy: &temporal.RetryPolicy{
					InitialInterval:    activity.Retry.InitialInterval,
					BackoffCoefficient: activity.Retry.BackoffCoefficient,
					MaximumInterval:    activity.Retry.MaxInterval,
					MaximumAttempts:    int32(activity.Retry.MaxAttempts),
				},
			}
		}

		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, activityOptions),
			activity.Name,
			activityInput,
		).Get(ctx, &result)

		if err != nil {
			workflow.GetLogger(ctx).Error("Activity failed",
				"activity_id", id,
				"activity_name", activity.Name,
				"error", err.Error())
			workflow.ExecuteActivity(ctx, "SaveActivityResult", requestID, id, nil, err)
			workflow.ExecuteActivity(ctx, "UpdateWorkflowRequest", requestID, "FAILED", results, err)
			return nil, err
		}

		// Save activity result
		workflow.ExecuteActivity(ctx, "SaveActivityResult", requestID, strconv.Itoa(id), result, nil)
		results[strconv.Itoa(id)] = result

		// Log activity completion
		workflow.GetLogger(ctx).Info("Activity completed",
			"activity_id", id,
			"activity_name", activity.Name)
	}

	// Update workflow request with final results
	workflow.ExecuteActivity(ctx, "UpdateWorkflowRequest", requestID, "COMPLETED", results, nil)

	// Log workflow completion
	workflow.GetLogger(ctx).Info("Dynamic workflow completed",
		"workflow_name", workflowName,
		"workflow_id", workflowID,
		"request_id", requestID)

	return results, nil
}
