package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	pb "github.com/flowpilotx/libs/grpc-common/pkg/api/worker/v1"
)

// WorkerClient wraps the worker service client with additional functionality
type WorkerClient struct {
	client     pb.WorkerServiceClient
	conn       *grpc.ClientConn
	options    *ClientOptions
	health     grpc_health_v1.HealthClient
}

// NewWorkerClient creates a new worker service client
func NewWorkerClient(opts *ClientOptions) (*WorkerClient, error) {
	if opts == nil {
		opts = DefaultClientOptions()
	}

	// Validate options
	if err := validateOptions(opts); err != nil {
		return nil, err
	}

	// Create connection target
	target := buildConnectionTarget(opts.Addresses)

	// Create connection
	conn, err := grpc.Dial(target, opts.BuildDialOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client connection: %w", err)
	}

	// Create client
	client := &WorkerClient{
		client:  pb.NewWorkerServiceClient(conn),
		conn:    conn,
		options: opts,
	}

	// Initialize health client if health checking is enabled
	if opts.LoadBalancing.Enabled && opts.LoadBalancing.HealthCheck {
		client.health = grpc_health_v1.NewHealthClient(conn)
	}

	return client, nil
}

// validateOptions validates the client options
func validateOptions(opts *ClientOptions) error {
	if len(opts.Addresses) == 0 {
		return fmt.Errorf("no server addresses provided")
	}

	if opts.LoadBalancing.Enabled && len(opts.Addresses) == 1 {
		return fmt.Errorf("load balancing enabled but only one address provided")
	}

	return nil
}

// buildConnectionTarget builds the connection target string
func buildConnectionTarget(addresses []string) string {
	// For load balancing, we use dns:///host1:port1,host2:port2 format
	if len(addresses) > 1 {
		return fmt.Sprintf("dns:///%s", strings.Join(addresses, ","))
	}
	return addresses[0]
}

// Close closes the client connection
func (c *WorkerClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CheckHealth checks the health of the service
func (c *WorkerClient) CheckHealth(ctx context.Context) error {
	if c.health == nil {
		return fmt.Errorf("health checking not enabled")
	}

	resp, err := c.health.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: "worker",
	})
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("service unhealthy: %s", resp.Status.String())
	}

	return nil
}

// Execute sends a synchronous execution request with retry support
func (c *WorkerClient) Execute(ctx context.Context, message []byte) (*pb.ExecuteResponse, error) {
	req := &pb.ExecuteRequest{
		RequestId: fmt.Sprintf("sync-%d", time.Now().Unix()),
		EventId:   time.Now().Unix(),
		Message:   message,
	}

	var resp *pb.ExecuteResponse
	var lastErr error

	for attempt := 0; attempt < c.options.RetryPolicy.MaxAttempts; attempt++ {
		// Check health if enabled
		if c.health != nil {
			if err := c.CheckHealth(ctx); err != nil {
				lastErr = err
				continue
			}
		}

		resp, lastErr = c.client.Execute(ctx, req)
		if lastErr == nil {
			return resp, nil
		}

		if !isRetryable(lastErr) {
			return nil, fmt.Errorf("non-retryable error: %w", lastErr)
		}

		backoff := calculateBackoff(attempt, c.options.RetryPolicy)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			continue
		}
	}

	return nil, fmt.Errorf("max retry attempts reached: %w", lastErr)
}

// ExecuteAsync sends an asynchronous execution request
func (c *WorkerClient) ExecuteAsync(ctx context.Context, message []byte) (*pb.ExecuteAsyncResponse, error) {
	req := &pb.ExecuteAsyncRequest{
		RequestId: fmt.Sprintf("async-%d", time.Now().Unix()),
		EventId:   time.Now().Unix(),
		Message:   message,
	}

	return c.client.ExecuteAsync(ctx, req)
}

// GetAsyncResult gets the result of an asynchronous execution
func (c *WorkerClient) GetAsyncResult(ctx context.Context, asyncResp *pb.ExecuteAsyncResponse) (*pb.ExecutionResult, error) {
	return c.client.GetAsyncResult(ctx, asyncResp)
}

// StreamExecute creates a bidirectional streaming connection
func (c *WorkerClient) StreamExecute(ctx context.Context) (pb.WorkerService_StreamExecuteClient, error) {
	return c.client.StreamExecute(ctx)
}

// WaitForAsyncCompletion waits for an async task to complete with timeout
func (c *WorkerClient) WaitForAsyncCompletion(ctx context.Context, asyncResp *pb.ExecuteAsyncResponse, timeout time.Duration) (*pb.ExecutionResult, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			result, err := c.GetAsyncResult(ctx, asyncResp)
			if err != nil {
				return nil, err
			}

			switch result.Status {
			case pb.ExecutionStatus_COMPLETED:
				return result, nil
			case pb.ExecutionStatus_FAILED:
				return result, fmt.Errorf("execution failed: %s", result.Error)
			default:
				continue
			}
		}
	}
}

// SendStreamMessages sends multiple messages over a stream with rate limiting
func (c *WorkerClient) SendStreamMessages(stream pb.WorkerService_StreamExecuteClient, messages [][]byte, rateLimit time.Duration) error {
	for i, msg := range messages {
		req := &pb.ExecuteRequest{
			RequestId: fmt.Sprintf("stream-%d-%d", time.Now().Unix(), i+1),
			EventId:   time.Now().Unix(),
			Message:   msg,
		}

		if err := stream.Send(req); err != nil {
			return fmt.Errorf("failed to send message %d: %w", i+1, err)
		}

		if rateLimit > 0 {
			time.Sleep(rateLimit)
		}
	}

	return stream.CloseSend()
}

// Helper functions

func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	switch st.Code() {
	case codes.Unavailable,
		codes.DeadlineExceeded,
		codes.ResourceExhausted,
		codes.Aborted:
		return true
	default:
		return false
	}
}

func calculateBackoff(attempt int, policy *RetryPolicy) time.Duration {
	backoff := float64(policy.InitialBackoff)
	for i := 0; i < attempt; i++ {
		backoff *= policy.BackoffMultiplier
	}

	if backoff > float64(policy.MaxBackoff) {
		backoff = float64(policy.MaxBackoff)
	}

	return time.Duration(backoff)
} 