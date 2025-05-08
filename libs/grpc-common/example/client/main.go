package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/flowpilotx/libs/grpc-common/pkg/api/worker/v1"
)

var (
	// Server connection settings
	grpcAddr = flag.String("grpc-addr", "localhost:50051", "The gRPC server address")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	// Set up gRPC connection
	conn, err := grpc.Dial(*grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create gRPC client
	client := pb.NewWorkerServiceClient(conn)

	log.Printf("Connected to gRPC server at %s", *grpcAddr)

	// Test synchronous execution
	log.Println("Testing synchronous execution...")
	syncResp, err := client.Execute(ctx, &pb.ExecuteRequest{
		RequestId: "sync-123",
		EventId:   1,
		Message:   []byte("Hello gRPC!"),
	})
	if err != nil {
		log.Fatalf("Sync execution failed: %v", err)
	}
	log.Printf("Sync execution response: %+v", syncResp)

	// Test asynchronous execution
	log.Println("Testing asynchronous execution...")
	asyncResp, err := client.ExecuteAsync(ctx, &pb.ExecuteAsyncRequest{
		RequestId: "async-456",
		EventId:   2,
		Message:   []byte("Hello Async gRPC!"),
	})
	if err != nil {
		log.Fatalf("Async execution failed: %v", err)
	}
	log.Printf("Async execution response: %+v", asyncResp)

	// Get async result
	log.Println("Getting async result...")
	resultResp, err := client.GetAsyncResult(ctx, asyncResp)
	if err != nil {
		log.Fatalf("Get async result failed: %v", err)
	}
	log.Printf("Async result response: %+v", resultResp)

	// Test streaming
	log.Println("Testing streaming...")
	stream, err := client.StreamExecute(ctx)
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}

	// Send a few messages
	for i := 1; i <= 3; i++ {
		req := &pb.ExecuteRequest{
			RequestId: fmt.Sprintf("stream-%d", i),
			EventId:   int64(i),
			Message:   []byte(fmt.Sprintf("Stream message %d", i)),
		}
		
		if err := stream.Send(req); err != nil {
			log.Fatalf("Failed to send message: %v", err)
		}
		
		resp, err := stream.Recv()
		if err != nil {
			log.Fatalf("Failed to receive response: %v", err)
		}
		log.Printf("Stream response %d: %+v", i, resp)
		time.Sleep(time.Second) // Wait between messages
	}

	if err := stream.CloseSend(); err != nil {
		log.Printf("Error closing stream: %v", err)
	}
}