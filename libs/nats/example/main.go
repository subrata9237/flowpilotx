package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/flowpilotx/libs/nats"
)

func main() {
	// Example 1: Basic Pub/Sub
	fmt.Println("\n=== Basic Pub/Sub Example ===")
	if err := basicPubSub(); err != nil {
		log.Printf("Basic pub/sub failed: %v", err)
	}

	// Example 2: JetStream
	fmt.Println("\n=== JetStream Example ===")
	if err := jetStreamExample(); err != nil {
		log.Printf("JetStream example failed: %v", err)
	}

	// Example 3: Request-Reply
	fmt.Println("\n=== Request-Reply Example ===")
	if err := requestReplyExample(); err != nil {
		log.Printf("Request-reply failed: %v", err)
	}

	// Example 4: Queue Groups
	fmt.Println("\n=== Queue Groups Example ===")
	if err := queueGroupExample(); err != nil {
		log.Printf("Queue groups failed: %v", err)
	}

	// Example 5: Error Handling
	fmt.Println("\n=== Error Handling Example ===")
	if err := errorHandlingExample(); err != nil {
		log.Printf("Error handling failed: %v", err)
	}
}

func basicPubSub() error {
	// Create client
	client, err := nats.NewClient(nats.Config{
		URLs:           []string{"nats://localhost:4222"},
		ConnectTimeout: time.Second * 10,
		MaxReconnects:  5,
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	// Create WaitGroup for synchronization
	var wg sync.WaitGroup
	wg.Add(1)

	// Subscribe to messages
	subscription, err := client.Subscribe("greetings", func(msg *nats.Message) {
		defer wg.Done()
		log.Printf("Received message: %s", string(msg.Data))
	})
	if err != nil {
		return fmt.Errorf("subscribe failed: %v", err)
	}
	defer subscription.Unsubscribe()

	// Publish message
	err = client.Publish("greetings", []byte("Hello, NATS!"))
	if err != nil {
		return fmt.Errorf("publish failed: %v", err)
	}

	// Wait for message to be received
	wg.Wait()
	return nil
}

func jetStreamExample() error {
	// Create client with JetStream enabled
	client, err := nats.NewClient(nats.Config{
		URLs:            []string{"nats://localhost:4222"},
		EnableJetStream: true,
		StreamConfig: nats.StreamConfig{
			Name:      "orders",
			Subjects:  []string{"orders.*"},
			Retention: nats.WorkQueuePolicy,
			MaxAge:    time.Hour * 24,
			Replicas:  1,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	// Create consumer
	consumer, err := client.CreateConsumer("orders", &nats.ConsumerConfig{
		DeliverPolicy: nats.DeliverAll,
		AckPolicy:     nats.AckExplicit,
		MaxDeliver:    3,
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer: %v", err)
	}

	// Subscribe to orders
	var wg sync.WaitGroup
	wg.Add(1)

	subscription, err := consumer.Subscribe("orders.new", func(msg *nats.Message) {
		defer wg.Done()
		log.Printf("Received order: %s", string(msg.Data))
		msg.Ack()
	})
	if err != nil {
		return fmt.Errorf("subscribe failed: %v", err)
	}
	defer subscription.Unsubscribe()

	// Publish order
	ack, err := client.PublishAsync("orders.new", []byte("new order data"))
	if err != nil {
		return fmt.Errorf("publish failed: %v", err)
	}

	// Wait for acknowledgment
	select {
	case <-ack.Ok():
		log.Println("Order stored in JetStream")
	case err := <-ack.Err():
		return fmt.Errorf("store failed: %v", err)
	case <-time.After(time.Second * 5):
		return fmt.Errorf("store timeout")
	}

	wg.Wait()
	return nil
}

func requestReplyExample() error {
	client, err := nats.NewClient(nats.Config{
		URLs: []string{"nats://localhost:4222"},
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	// Start reply handler
	subscription, err := client.Subscribe("service.time", func(msg *nats.Message) {
		response := time.Now().Format(time.RFC3339)
		msg.Respond([]byte(response))
	})
	if err != nil {
		return fmt.Errorf("subscribe failed: %v", err)
	}
	defer subscription.Unsubscribe()

	// Make request
	ctx := context.Background()
	response, err := client.Request(ctx, "service.time", nil, time.Second)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	log.Printf("Current time: %s", string(response.Data))
	return nil
}

func queueGroupExample() error {
	// Create multiple workers
	var wg sync.WaitGroup
	workerCount := 3

	// Create signal channel for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			client, err := nats.NewClient(nats.Config{
				URLs:       []string{"nats://localhost:4222"},
				QueueGroup: "workers",
			})
			if err != nil {
				log.Printf("Worker %d failed to create client: %v", workerID, err)
				return
			}
			defer client.Close()

			// Subscribe to queue group
			subscription, err := client.QueueSubscribe(
				"tasks",
				"workers",
				func(msg *nats.Message) {
					log.Printf("Worker %d processing task: %s", workerID, string(msg.Data))
					time.Sleep(time.Millisecond * 100) // Simulate processing
				},
			)
			if err != nil {
				log.Printf("Worker %d failed to subscribe: %v", workerID, err)
				return
			}
			defer subscription.Unsubscribe()

			// Wait for shutdown signal
			<-sigChan
		}(i)
	}

	// Create publisher
	publisher, err := nats.NewClient(nats.Config{
		URLs: []string{"nats://localhost:4222"},
	})
	if err != nil {
		return fmt.Errorf("failed to create publisher: %v", err)
	}
	defer publisher.Close()

	// Publish tasks
	for i := 0; i < 10; i++ {
		task := fmt.Sprintf("task-%d", i)
		err = publisher.Publish("tasks", []byte(task))
		if err != nil {
			log.Printf("Failed to publish task: %v", err)
		}
		time.Sleep(time.Millisecond * 50)
	}

	// Signal workers to shut down
	close(sigChan)

	// Wait for workers to finish
	wg.Wait()
	return nil
}

func errorHandlingExample() error {
	client, err := nats.NewClient(nats.Config{
		URLs:          []string{"nats://invalid:4222"},
		MaxReconnects: 1,
		ReconnectWait: time.Second,
	})

	// Handle connection error
	if err != nil {
		switch err := err.(type) {
		case *nats.ConnectionError:
			log.Printf("Connection error: %v", err)
		default:
			log.Printf("Unknown error: %v", err)
		}
		return err
	}
	defer client.Close()

	// Try invalid subject
	err = client.Publish("", []byte("invalid"))
	if err != nil {
		switch err := err.(type) {
		case *nats.PublishError:
			log.Printf("Publish error: %v", err)
		default:
			log.Printf("Unknown error: %v", err)
		}
	}

	// Try invalid subscription
	_, err = client.Subscribe("", nil)
	if err != nil {
		switch err := err.(type) {
		case *nats.SubscriptionError:
			log.Printf("Subscription error: %v", err)
		default:
			log.Printf("Unknown error: %v", err)
		}
	}

	// Try request with timeout
	ctx := context.Background()
	_, err = client.Request(ctx, "service.nonexistent", nil, time.Millisecond*100)
	if err != nil {
		switch err := err.(type) {
		case *nats.TimeoutError:
			log.Printf("Timeout error: %v", err)
		default:
			log.Printf("Unknown error: %v", err)
		}
	}

	return nil
} 