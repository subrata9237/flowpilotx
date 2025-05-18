package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/libs/mongodb"
	"github.com/flowpilotx/libs/worker/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

type WorkflowService interface {
	CreateWorkflow(ctx context.Context, workflow *model.WorkflowSchema) error
	GetWorkflow(ctx context.Context, id string) (*model.WorkflowSchema, error)
	TriggerWorkflow(ctx context.Context, workflowID string, input map[string]interface{}) (string, error)
	GetWorkflowRequest(ctx context.Context, requestID string) (*model.Workflow, error)
	GetActivityResults(ctx context.Context, requestID string) ([]*model.ActivityDefinition, error)
}

type workflowService struct {
	temporalClient client.Client
	mongoClient    *mongodb.Client
	logger         logger.LoggerInterface
	config         *config.Config
}

func NewWorkflowService(temporalClient client.Client, mongoClient *mongodb.Client, logger logger.LoggerInterface, config *config.Config) WorkflowService {
	return &workflowService{
		temporalClient: temporalClient,
		mongoClient:    mongoClient,
		logger:         logger,
		config:         config,
	}
}

func (s *workflowService) CreateWorkflow(ctx context.Context, workflow *model.WorkflowSchema) error {
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = workflow.CreatedAt
	workflow.Status = "ACTIVE"
	workflow.Version = 1
	workflow.Queue = s.getTaskQueue(workflow)

	result, err := s.mongoClient.InsertOne(ctx, "workflows", workflow)
	if err != nil {
		s.logger.ErrorWithCtx(ctx, "Failed to create workflow", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("failed to create workflow: %v", err)
	}

	workflow.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (s *workflowService) GetWorkflow(ctx context.Context, id string) (*model.WorkflowSchema, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid workflow ID: %v", err)
	}

	var workflow model.WorkflowSchema
	err = s.mongoClient.FindOne(ctx, "workflows", bson.M{"_id": objectID}, &workflow)
	if err != nil {
		if err == mongodb.ErrNoDocuments {
			return nil, fmt.Errorf("workflow not found")
		}
		return nil, fmt.Errorf("failed to get workflow: %v", err)
	}

	return &workflow, nil
}

func (s *workflowService) TriggerWorkflow(ctx context.Context, workflowID string, input map[string]interface{}) (string, error) {
	workflow, err := s.GetWorkflow(ctx, workflowID)
	if err != nil {
		return "", err
	}
	//print workflow
	s.logger.InfoWithCtx(ctx, "Workflow", map[string]interface{}{
		"workflow": workflow,
	})
	// Validate workflow name
	if workflow.Name == "" {
		return "", fmt.Errorf("workflow name cannot be empty")
	}
	request := &model.Workflow{
		WorkflowSchema: workflow,
		Parameters:     input,
		Status:         "PENDING",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	result, err := s.mongoClient.InsertOne(ctx, "workflows", request)
	if err != nil {
		return "", fmt.Errorf("failed to create workflow request: %v", err)
	}

	request.RequestID = result.InsertedID.(primitive.ObjectID)
	// Get task queue based on workflow name and ID
	if request.WorkflowSchema.Queue == "" {
		return "", fmt.Errorf("workflow queue cannot be empty")
	}
	options := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             request.WorkflowSchema.Queue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
	workflowNameGeneric := "FlowpilotxWorkflow"
	_, err = s.temporalClient.ExecuteWorkflow(ctx, options, workflowNameGeneric, request)
	if err != nil {
		// Log the error with workflow name for better debugging
		s.logger.ErrorWithCtx(ctx, "Failed to start workflow", map[string]interface{}{
			"error":         err.Error(),
			"workflow_name": workflow.Name,
			"workflow_id":   workflowID,
			"request_id":    request.RequestID.Hex(),
		})
		return "", fmt.Errorf("failed to start workflow %s: %v", workflow.Name, err)
	}

	return request.RequestID.Hex(), nil
}

func (s *workflowService) GetWorkflowRequest(ctx context.Context, requestID string) (*model.Workflow, error) {
	objectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return nil, fmt.Errorf("invalid request ID: %v", err)
	}

	var request model.Workflow
	err = s.mongoClient.FindOne(ctx, "workflow_requests", bson.M{"_id": objectID}, &request)
	if err != nil {
		if err == mongodb.ErrNoDocuments {
			return nil, fmt.Errorf("workflow request not found")
		}
		return nil, fmt.Errorf("failed to get workflow request: %v", err)
	}

	return &request, nil
}

func (s *workflowService) GetActivityResults(ctx context.Context, requestID string) ([]*model.ActivityDefinition, error) {
	objectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return nil, fmt.Errorf("invalid request ID: %v", err)
	}

	cursor, err := s.mongoClient.FindMany(ctx, "activity_results", bson.M{"request_id": objectID})
	if err != nil {
		return nil, fmt.Errorf("failed to get activity results: %v", err)
	}
	defer cursor.Close(ctx)

	var results []*model.ActivityDefinition
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode activity results: %v", err)
	}

	return results, nil
}

func (s *workflowService) getTaskQueue(workflow *model.WorkflowSchema) string {
	if workflow.Queue != "" {
		return workflow.Queue
	}
	// Default queue name prefix
	queuePrefix := "flowpilotx"

	// Get environment from config or use default
	env := s.config.Environment

	// Get region from config or use default
	region := s.config.Region

	// Create base queue name with workflow name and ID
	baseQueueName := fmt.Sprintf("%s-%s-%s-%s-%s",
		queuePrefix,
		env,
		region,
		workflow.Name,
		workflow.ID.Hex(),
	)

	// Determine queue type based on workflow properties
	queueType := "default"

	// Check workflow status for priority
	if workflow.Status == "high_priority" {
		queueType = "high-priority"
	} else if workflow.Status == "background" {
		queueType = "background"
	}

	// Check workflow version for version-specific queue
	if workflow.Version > 1 {
		queueType = fmt.Sprintf("%s-v%d", queueType, workflow.Version)
	}

	// Construct final queue name
	queueName := fmt.Sprintf("%s-%s", baseQueueName, queueType)

	// Log queue selection
	s.logger.InfoWithCtx(context.Background(), "Selected task queue", map[string]interface{}{
		"workflow_name": workflow.Name,
		"workflow_id":   workflow.ID.Hex(),
		"queue_name":    queueName,
		"queue_type":    queueType,
	})

	return queueName
}
