package service

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/flowpilotx/libs/grpc-common/pkg/api/worker/v1"
	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/services/flowpilotx-worker/internal/metrics"
	libctx "github.com/flowpilotx/libs/context"
)

// WorkerConfig holds worker-specific configuration
type WorkerConfig struct {
	TaskTimeout        time.Duration `yaml:"task_timeout" json:"task_timeout"`
	MaxConcurrentTasks int          `yaml:"max_concurrent_tasks" json:"max_concurrent_tasks"`
	Retry             RetryConfig   `yaml:"retry" json:"retry"`
}

// RetryConfig holds retry-related configuration
type RetryConfig struct {
	MaxAttempts     int           `yaml:"max_attempts" json:"max_attempts"`
	InitialInterval time.Duration `yaml:"initial_interval" json:"initial_interval"`
	MaxInterval     time.Duration `yaml:"max_interval" json:"max_interval"`
}

// WorkerService implements the gRPC worker service
type WorkerService struct {
	pb.UnimplementedWorkerServiceServer
	logger     logger.LoggerInterface
	metrics    *metrics.Metrics
	executions sync.Map
	config     *config.Config
	workerCfg  *WorkerConfig
}

// NewWorkerService creates a new worker service instance
func NewWorkerService(cfg *config.Config, log logger.LoggerInterface, m *metrics.Metrics) *WorkerService {
	// Default worker configuration
	workerCfg := &WorkerConfig{
		TaskTimeout:        30 * time.Second,
		MaxConcurrentTasks: 100,
		Retry: RetryConfig{
			MaxAttempts:     3,
			InitialInterval: 1 * time.Second,
			MaxInterval:     5 * time.Second,
		},
	}

	// TODO: Load worker config from cfg when available
	
	return &WorkerService{
		logger:     log,
		metrics:    m,
		config:     cfg,
		workerCfg:  workerCfg,
		executions: sync.Map{},
	}
}

// Execute handles synchronous execution requests
func (s *WorkerService) Execute(c context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	start := time.Now()
	s.metrics.IncInFlightRequests()
	defer s.metrics.DecInFlightRequests()

	// Get request ID from context
	requestID, err := libctx.GetRequestID(c)
	if err != nil {
		s.metrics.IncRequests("execute", "error")
		s.logger.Error(c, "Failed to get request ID from context", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, status.Error(codes.Internal, "request ID not found in context")
	}

	// Add execution time to context
	c = libctx.WithExecutionTime(c, time.Since(start))

	// Get request fields for logging
	fields := map[string]interface{}{
		"request_id": requestID,
		"event_id":   req.GetEventId(),
	}
	s.logger.Info(c, "Processing execution request", fields)

	// Validate request
	if err := validateExecuteRequest(req); err != nil {
		s.metrics.IncRequests("execute", "error")
		s.logger.Error(c, "Invalid request", map[string]interface{}{
			"error": err.Error(),
			"fields": fields,
		})
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Create context with timeout from config
	timeoutCtx, cancel := context.WithTimeout(c, s.workerCfg.TaskTimeout)
	defer cancel()

	// Process request with timeout
	var processedMessage []byte
	done := make(chan error, 1)
	go func() {
		processedMessage = append(req.GetMessage(), []byte(" - Processed")...)
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			s.metrics.IncRequests("execute", "error")
			s.logger.Error(c, "Processing failed", map[string]interface{}{
				"error": err.Error(),
				"fields": fields,
			})
			return nil, status.Error(codes.Internal, "processing failed")
		}
	case <-timeoutCtx.Done():
		s.metrics.IncRequests("execute", "timeout")
		s.logger.Error(c, "Processing timeout", map[string]interface{}{
			"error": timeoutCtx.Err().Error(),
			"fields": fields,
		})
		return nil, status.Error(codes.DeadlineExceeded, "processing timeout")
	}

	// Record metrics
	s.metrics.IncRequests("execute", "success")
	s.metrics.ObserveRequestDuration("execute", time.Since(start).Seconds())

	// Log success with execution time
	s.logger.Info(c, "Request processed successfully", map[string]interface{}{
		"request_id": requestID,
		"event_id":   req.GetEventId(),
		"duration":   time.Since(start).String(),
	})

	return &pb.ExecuteResponse{
		EventId:   req.GetEventId(),
		Status:    pb.ExecuteStatus_EXECUTE_STATUS_SUCCESS,
		Message:   processedMessage,
	}, nil
}

// ExecuteAsync handles asynchronous execution requests
func (s *WorkerService) ExecuteAsync(c context.Context, req *pb.ExecuteAsyncRequest) (*pb.ExecuteAsyncResponse, error) {
	s.metrics.IncInFlightRequests()
	defer s.metrics.DecInFlightRequests()

	// Get request ID from context
	requestID, err := libctx.GetRequestID(c)
	if err != nil {
		s.metrics.IncRequests("execute_async", "error")
		s.logger.Error(c, "Failed to get request ID from context", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, status.Error(codes.Internal, "request ID not found in context")
	}

	executionID := fmt.Sprintf("exec-%s-%d", requestID, time.Now().Unix())
	fields := map[string]interface{}{
		"request_id":   requestID,
		"execution_id": executionID,
		"event_id":     req.GetEventId(),
	}

	// Check if we can accept more concurrent tasks
	var count int
	s.executions.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	if count >= s.workerCfg.MaxConcurrentTasks {
		s.metrics.IncRequests("execute_async", "rejected")
		s.logger.Error(c, "Max concurrent tasks reached", map[string]interface{}{
			"fields": fields,
		})
		return nil, status.Error(codes.ResourceExhausted, "max concurrent tasks reached")
	}

	// Start async processing with retry logic
	go func() {
		var attempt int
		var err error
		backoff := s.workerCfg.Retry.InitialInterval

		for attempt < s.workerCfg.Retry.MaxAttempts {
			attempt++
			
			// Create context with task timeout
			timeoutCtx, cancel := context.WithTimeout(c, s.workerCfg.TaskTimeout)
			
			// Simulate async processing
			done := make(chan error, 1)
			go func() {
				time.Sleep(2 * time.Second)
				done <- nil
			}()

			// Wait for processing or timeout
			select {
			case err = <-done:
				if err == nil {
					result := &pb.ExecuteAsyncResult{
						EventId:     req.GetEventId(),
						Message:     append(req.GetMessage(), []byte(" - Async Processed")...),
						Status:      pb.ExecuteStatus_EXECUTE_STATUS_SUCCESS,
					}
					s.executions.Store(executionID, result)
					s.logger.Info(c, "Async processing completed", map[string]interface{}{
						"fields": fields,
					})
					cancel()
					return
				}
			case <-timeoutCtx.Done():
				err = timeoutCtx.Err()
			}
			
			cancel()
			
			if attempt < s.workerCfg.Retry.MaxAttempts {
				s.logger.Warn(c, "Retrying async processing", map[string]interface{}{
					"attempt": attempt,
					"backoff": backoff.String(),
					"error":   err.Error(),
					"fields":  fields,
				})
				time.Sleep(backoff)
				
				// Increase backoff up to max interval
				if backoff < s.workerCfg.Retry.MaxInterval {
					backoff *= 2
					if backoff > s.workerCfg.Retry.MaxInterval {
						backoff = s.workerCfg.Retry.MaxInterval
					}
				}
			}
		}

		// All retries failed
		s.metrics.IncRequests("execute_async", "error")
		s.logger.Error(c, "All async processing attempts failed", map[string]interface{}{
			"attempts": attempt,
			"error":    err.Error(),
			"fields":   fields,
		})

		result := &pb.ExecuteAsyncResult{
			EventId:     req.GetEventId(),
			Status:      pb.ExecuteStatus_EXECUTE_STATUS_FAILED,
			Message:     []byte(fmt.Sprintf("Failed after %d attempts: %v", attempt, err)),
		}
		s.executions.Store(executionID, result)
	}()

	return &pb.ExecuteAsyncResponse{
		Status:      pb.ExecuteStatus_EXECUTE_STATUS_IN_PROGRESS,
	}, nil
}

// GetAsyncResult retrieves the result of an async execution
func (s *WorkerService) GetAsyncResult(c context.Context, req *pb.ExecuteAsyncResponse) (*pb.ExecuteAsyncResult, error) {
	s.metrics.IncInFlightRequests()
	defer s.metrics.DecInFlightRequests()

	// Get request ID from context
	requestID, err := libctx.GetRequestID(c)
	if err != nil {
		s.metrics.IncRequests("get_async_result", "error")
		s.logger.Error(c, "Failed to get request ID from context", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, status.Error(codes.Internal, "request ID not found in context")
	}

	// Get execution ID from context
	executionID, err := libctx.GetExecutionID(c)
	if err != nil {
		s.metrics.IncRequests("get_async_result", "error")
		s.logger.Error(c, "Failed to get execution ID from context", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, status.Error(codes.Internal, "execution ID not found in context")
	}

	fields := map[string]interface{}{
		"request_id":   requestID,
		"execution_id": executionID,
	}
	s.logger.Info(c, "Retrieving async result", fields)

	// Get result from storage
	if result, ok := s.executions.Load(executionID); ok {
		s.metrics.IncRequests("get_async_result", "success")
		return result.(*pb.ExecuteAsyncResult), nil
	}

	// Return in-progress if not found
	s.metrics.IncRequests("get_async_result", "in_progress")
	return &pb.ExecuteAsyncResult{
		EventId: req.GetEventId(),  // Use execution ID from context
		Status: pb.ExecuteStatus_EXECUTE_STATUS_IN_PROGRESS,
	}, nil
}

// StreamExecute handles streaming execution requests
func (s *WorkerService) StreamExecute(stream pb.WorkerService_StreamExecuteServer) error {
	s.logger.Info(stream.Context(), "Started streaming execution", nil)

	// Create context with timeout from config
	c := stream.Context()
	streamCtx, cancel := context.WithTimeout(c, s.workerCfg.TaskTimeout)
	defer cancel()

	for {
		select {
		case <-streamCtx.Done():
			s.logger.Warn(c, "Stream timeout reached", nil)
			return status.Error(codes.DeadlineExceeded, "stream timeout")
		default:
			req, err := stream.Recv()
			if err == io.EOF {
				s.logger.Info(c, "Client closed stream", nil)
				return nil
			}
			if err != nil {
				s.logger.Error(c, "Stream receive error", map[string]interface{}{
					"error": err.Error(),
				})
				return status.Error(codes.Internal, "failed to receive request")
			}

			start := time.Now()
			s.metrics.IncInFlightRequests()

			// Get request ID from context
			requestID, err := libctx.GetRequestID(c)
			if err != nil {
				s.metrics.DecInFlightRequests()
				s.metrics.IncRequests("stream_execute", "error")
				s.logger.Error(c, "Failed to get request ID from context", map[string]interface{}{
					"error": err.Error(),
				})
				return status.Error(codes.Internal, "request ID not found in context")
			}

			fields := map[string]interface{}{
				"request_id": requestID,
				"event_id":   req.GetEventId(),
			}
			s.logger.Info(c, "Processing stream request", fields)

			// Check concurrent tasks limit
			var count int
			s.executions.Range(func(key, value interface{}) bool {
				count++
				return true
			})
			if count >= s.workerCfg.MaxConcurrentTasks {
				s.metrics.DecInFlightRequests()
				s.metrics.IncRequests("stream_execute", "rejected")
				s.logger.Error(c, "Max concurrent tasks reached", map[string]interface{}{
					"fields": fields,
				})
				return status.Error(codes.ResourceExhausted, "max concurrent tasks reached")
			}

			// Process request with timeout
			processCtx, processCancel := context.WithTimeout(c, s.workerCfg.TaskTimeout)
			done := make(chan error, 1)
			var processedMessage []byte

			go func() {
				processedMessage = append(req.GetMessage(), []byte(" - Stream Processed")...)
				done <- nil
			}()

			select {
			case err := <-done:
				if err != nil {
					processCancel()
					s.metrics.DecInFlightRequests()
					s.metrics.IncRequests("stream_execute", "error")
					s.logger.Error(c, "Processing failed", map[string]interface{}{
						"error": err.Error(),
						"fields": fields,
					})
					return status.Error(codes.Internal, "processing failed")
				}
			case <-processCtx.Done():
				processCancel()
				s.metrics.DecInFlightRequests()
				s.metrics.IncRequests("stream_execute", "timeout")
				s.logger.Error(c, "Processing timeout", map[string]interface{}{
					"error": processCtx.Err().Error(),
					"fields": fields,
				})
				return status.Error(codes.DeadlineExceeded, "processing timeout")
			}

			processCancel()
			response := &pb.ExecuteResponse{
				EventId:   req.GetEventId(),
				Message:   processedMessage,
				Status:    pb.ExecuteStatus_EXECUTE_STATUS_SUCCESS,
			}

			if err := stream.Send(response); err != nil {
				s.metrics.DecInFlightRequests()
				s.metrics.IncRequests("stream_execute", "error")
				s.logger.Error(c, "Stream send error", map[string]interface{}{
					"error": err.Error(),
					"fields": fields,
				})
				return status.Error(codes.Internal, "failed to send response")
			}

			s.metrics.DecInFlightRequests()
			s.metrics.IncRequests("stream_execute", "success")
			s.metrics.ObserveRequestDuration("stream_execute", time.Since(start).Seconds())
			s.logger.Info(c, "Stream response sent", map[string]interface{}{
				"fields": fields,
			})
		}
	}
}

func validateExecuteRequest(req *pb.ExecuteRequest) error {
	if len(req.GetMessage()) == 0 {
		return fmt.Errorf("message is required")
	}
	return nil
}

func validateAsyncRequest(req *pb.ExecuteAsyncRequest) error {
	if len(req.GetMessage()) == 0 {
		return fmt.Errorf("message is required")
	}
	return nil
} 
