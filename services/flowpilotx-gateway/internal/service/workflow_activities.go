package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.temporal.io/sdk/activity"
)

type WorkflowActivities struct {
	mongoClient *mongo.Client
	logger      logger.LoggerInterface
}

func NewWorkflowActivities(mongoClient *mongo.Client, logger logger.LoggerInterface) *WorkflowActivities {
	return &WorkflowActivities{
		mongoClient: mongoClient,
		logger:      logger,
	}
}

func (a *WorkflowActivities) ExecuteActivity(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)

	// Record activity start
	result := &models.ActivityResult{
		RequestID:  input["request_id"].(primitive.ObjectID),
		ActivityID: activityInfo.ActivityType.Name,
		StartedAt:  time.Now(),
		Attempt:    int(activityInfo.Attempt),
	}

	// Execute activity logic based on activity type
	var output interface{}
	var err error
	switch activityInfo.ActivityType.Name {
	case "add":
		output, err = a.executeAddActivity(input)
	case "multiply":
		output, err = a.executeMultiplyActivity(input)
	default:
		err = fmt.Errorf("unknown activity type: %s", activityInfo.ActivityType.Name)
	}

	if err != nil {
		result.Error = err.Error()
		a.saveActivityResult(ctx, result)
		return nil, err
	}

	// Record activity completion
	now := time.Now()
	result.CompletedAt = &now
	result.Result = output

	if err := a.saveActivityResult(ctx, result); err != nil {
		a.logger.ErrorWithCtx(ctx, "Failed to save activity result", map[string]interface{}{"error": err.Error()})
		return nil, err
	}

	return output, nil
}

func (a *WorkflowActivities) executeAddActivity(input map[string]interface{}) (float64, error) {
	x, ok := input["a"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid input: a must be a number")
	}
	y, ok := input["b"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid input: b must be a number")
	}
	return x + y, nil
}

func (a *WorkflowActivities) executeMultiplyActivity(input map[string]interface{}) (float64, error) {
	x, ok := input["a"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid input: a must be a number")
	}
	y, ok := input["b"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid input: b must be a number")
	}
	return x * y, nil
}

func (a *WorkflowActivities) saveActivityResult(ctx context.Context, result *models.ActivityResult) error {
	collection := a.mongoClient.Database("flowpilotx").Collection("activity_results")
	_, err := collection.InsertOne(ctx, result)
	return err
}
