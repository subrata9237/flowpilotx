package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	pb "github.com/flowpilotx/libs/grpc-common/pkg/api/worker/v1"
)

var (
	grpcPort = flag.Int("grpc-port", 50051, "The gRPC server port")
)

// server implements the WorkerService
type server struct {
	pb.UnimplementedWorkerServiceServer
	executions sync.Map // stores async execution results
}

// Execute handles a single event execution request
func (s *server) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	log.Printf("Processing request %s", req.GetRequestId())
	
	// Simulate processing
	time.Sleep(100 * time.Millisecond)
	processedMessage := append(req.GetMessage(), []byte(" - Processed")...)
	
	return &pb.ExecuteResponse{
		RequestId: req.GetRequestId(),
		EventId:   req.GetEventId(),
		Status:    pb.ExecuteStatus_EXECUTE_STATUS_SUCCESS,
		Message:   processedMessage,
	}, nil
}

// ExecuteAsync handles asynchronous event execution
func (s *server) ExecuteAsync(ctx context.Context, req *pb.ExecuteAsyncRequest) (*pb.ExecuteAsyncResponse, error) {
	executionID := fmt.Sprintf("exec-%s-%d", req.GetRequestId(), time.Now().Unix())
	log.Printf("Starting async execution %s", executionID)

	// Start async processing in a goroutine
	go func() {
		time.Sleep(2 * time.Second) // Simulate async processing
		result := &pb.ExecuteAsyncResult{
			RequestId:   req.GetRequestId(),
			ExecutionId: executionID,
			EventId:    req.GetEventId(),
			Message:    append(req.GetMessage(), []byte(" - Async Processed")...),
			Status:     pb.ExecuteStatus_EXECUTE_STATUS_SUCCESS,
		}
		s.executions.Store(executionID, result)
	}()

	return &pb.ExecuteAsyncResponse{
		RequestId:   req.GetRequestId(),
		ExecutionId: executionID,
		EventId:    req.GetEventId(),
		Message:    req.GetMessage(),
		Status:     pb.ExecuteStatus_EXECUTE_STATUS_IN_PROGRESS,
	}, nil
}

// GetAsyncResult retrieves the result of an async execution
func (s *server) GetAsyncResult(ctx context.Context, req *pb.ExecuteAsyncResponse) (*pb.ExecuteAsyncResult, error) {
	if result, ok := s.executions.Load(req.GetExecutionId()); ok {
		return result.(*pb.ExecuteAsyncResult), nil
	}
	
	return &pb.ExecuteAsyncResult{
		ExecutionId: req.GetExecutionId(),
		Status:     pb.ExecuteStatus_EXECUTE_STATUS_IN_PROGRESS,
	}, nil
}

// StreamExecute handles streaming event execution requests
func (s *server) StreamExecute(stream pb.WorkerService_StreamExecuteServer) error {
	log.Printf("Started streaming execution")
	
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Client closed the stream")
			return nil
		}
		if err != nil {
			log.Printf("Error receiving request: %v", err)
			return fmt.Errorf("failed to receive request: %v", err)
		}
		
		log.Printf("Processing stream request %s", req.GetRequestId())
		processedMessage := append(req.GetMessage(), []byte(" - Stream Processed")...)
		
		response := &pb.ExecuteResponse{
			RequestId: req.GetRequestId(),
			EventId:   req.GetEventId(),
			Message:   processedMessage,
			Status:    pb.ExecuteStatus_EXECUTE_STATUS_SUCCESS,
		}
		
		if err := stream.Send(response); err != nil {
			log.Printf("Error sending response: %v", err)
			return fmt.Errorf("failed to send response: %v", err)
		}
		log.Printf("Sent response for request: %v", req.GetRequestId())
	}
}

func main() {
	flag.Parse()

	// Create listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Configure server options
	serverOpts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:             1 * time.Minute,
			Timeout:         20 * time.Second,
		}),
	}

	// Create gRPC server
	s := grpc.NewServer(serverOpts...)
	pb.RegisterWorkerServiceServer(s, &server{})
	reflection.Register(s)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting gRPC server on port %d", *grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	// Gracefully stop the server
	log.Println("Shutting down server...")
	s.GracefulStop()
} 