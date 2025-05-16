package workerhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/libs/mongodb"
	acitivityWrapper "github.com/flowpilotx/libs/worker/activities"
	"github.com/flowpilotx/libs/worker/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.temporal.io/sdk/activity"
)

type WorkflowActivities struct {
	mongoClient *mongodb.Client
	logger      logger.LoggerInterface
	activities  *acitivityWrapper.Activities
}

func NewWorkflowActivities(mongoClient *mongodb.Client, logger logger.LoggerInterface) *WorkflowActivities {
	workflowActivity := &WorkflowActivities{
		mongoClient: mongoClient,
		logger:      logger,
		activities: &acitivityWrapper.Activities{
			Logger:      logger,
			MongoClient: mongoClient,
		},
	}

	// Initialize activity map
	workflowActivity.activities.RegisterActivities()
	return workflowActivity
}

func (a *WorkflowActivities) ExecuteActivity(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	activityType := strings.ToLower(activityInfo.ActivityType.Name) // Convert to lowercase for matching

	a.logger.InfoWithCtx(ctx, "Looking up activity", map[string]interface{}{
		"activity_type":         activityType,
		"registered_activities": a.activities.GetAvailableActivities(),
	})

	a.logger.InfoWithCtx(ctx, "Starting activity execution", map[string]interface{}{
		"activity_type": activityType,
		"attempt":       activityInfo.Attempt,
		"workflow_id":   input["workflow_id"],
		"request_id":    input["request_id"],
	})
	now := time.Now()
	var activityDefinition model.ActivityDefinition
	err := a.mapToStruct(input, &activityDefinition)
	if err != nil {
		a.logger.ErrorWithCtx(ctx, "Failed to convert input to struct", map[string]interface{}{
			"error": err.Error(),
			"input": input,
		})
	}
	// Convert string ID to ObjectID
	objectID, _ := primitive.ObjectIDFromHex(activityDefinition.ID)

	// Record activity start
	result := &model.ActivityResult{
		RequestID:  objectID,
		ActivityID: activityDefinition.ID,
		StartedAt:  &now,
		Attempt:    int(activityInfo.Attempt),
	}
	a.logger.InfoWithCtx(ctx, "Activity definition", map[string]interface{}{
		"activity_definition": activityDefinition,
	})
	// Get the activity function from the map
	activityFunc, exists := a.activities.ActivityMap[activityDefinition.Name]
	if !exists {
		err := fmt.Errorf("unknown activity type: %s", activityType)
		a.logger.ErrorWithCtx(ctx, "Activity type not found", map[string]interface{}{
			"activity_type":        activityType,
			"available_activities": a.activities.GetAvailableActivities(),
		})
		result.Error = err.Error()
		//a.saveActivityResult(ctx, result)
		return result, err
	}

	a.logger.InfoWithCtx(ctx, "Executing activity function", map[string]interface{}{
		"activity_type": activityType,
		"input":         input,
	})

	// Execute the activity
	resp, err := activityFunc(input)
	if err != nil {
		a.logger.ErrorWithCtx(ctx, "Activity execution failed", map[string]interface{}{
			"activity_type": activityType,
			"error":         err.Error(),
		})
		return nil, err
	}
	// Record activity completion
	finalTime := time.Now()
	result.CompletedAt = &finalTime
	result.Result = resp
	result.Error = ""
	/*if err := a.saveActivityResult(ctx, result); err != nil {
		a.logger.ErrorWithCtx(ctx, "Failed to save activity result", map[string]interface{}{
			"error":         err.Error(),
			"activity_type": activityType,
		})
		return nil, err
	}*/

	a.logger.InfoWithCtx(ctx, "Activity completed successfully", map[string]interface{}{
		"activity_type": activityType,
		"result":        result.Result,
		"duration_ms":   now.Sub(*result.StartedAt).Milliseconds(),
	})

	return a.structToMap(result.Result, nil)
}

func (a *WorkflowActivities) saveActivityResult(ctx context.Context, result *model.ActivityResult) error {
	a.logger.InfoWithCtx(ctx, "Saving activity result", map[string]interface{}{
		"activity_id": result.ActivityID,
		"request_id":  result.RequestID,
		"status":      result.Error == "",
	})

	collection := a.mongoClient.Collection("activity_results")
	_, err := collection.InsertOne(ctx, result)
	if err != nil {
		a.logger.ErrorWithCtx(ctx, "Failed to save activity result", map[string]interface{}{
			"error":       err.Error(),
			"activity_id": result.ActivityID,
			"request_id":  result.RequestID,
		})
		return err
	}

	a.logger.InfoWithCtx(ctx, "Activity result saved successfully", map[string]interface{}{
		"activity_id": result.ActivityID,
		"request_id":  result.RequestID,
	})
	return nil
}

// GetAvailableActivities returns a list of all available activities
func (a *WorkflowActivities) getAvailableActivities() []string {
	return a.activities.GetAvailableActivities()
}

// GetAllActivityNames returns all available activity names
func (a *WorkflowActivities) getAllActivityNames() []string {
	var activityNames []string

	// Get type information of Activities struct
	t := reflect.TypeOf(a)

	// Iterate through all methods
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)

		// Check if method name starts with "execute"
		if strings.HasPrefix(method.Name, "execute") {
			// Convert "ExecuteMultiply" to "Multiply"
			activityName := strings.TrimPrefix(method.Name, "execute")
			activityNames = append(activityNames, activityName)
		}
	}

	return activityNames
}
func (w *WorkflowActivities) structToMap(data interface{}, errInput error) (map[string]interface{}, error) {
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
func (w *WorkflowActivities) mapToStruct(data map[string]interface{}, result interface{}) error {
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
