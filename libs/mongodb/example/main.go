package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/flowpilotx/libs/mongodb"
)

func main() {
	// Example 1: Basic Operations
	fmt.Println("\n=== Basic Operations Example ===")
	if err := basicOperations(); err != nil {
		log.Printf("Basic operations failed: %v", err)
	}

	// Example 2: Transactions
	fmt.Println("\n=== Transactions Example ===")
	if err := transactionExample(); err != nil {
		log.Printf("Transactions failed: %v", err)
	}

	// Example 3: Bulk Operations
	fmt.Println("\n=== Bulk Operations Example ===")
	if err := bulkOperations(); err != nil {
		log.Printf("Bulk operations failed: %v", err)
	}

	// Example 4: Change Streams
	fmt.Println("\n=== Change Streams Example ===")
	if err := changeStreams(); err != nil {
		log.Printf("Change streams failed: %v", err)
	}

	// Example 5: Error Handling
	fmt.Println("\n=== Error Handling Example ===")
	if err := errorHandling(); err != nil {
		log.Printf("Error handling failed: %v", err)
	}
}

func basicOperations() error {
	// Initialize client
	client, err := mongodb.NewClient(mongodb.Config{
		URI:            "mongodb://localhost:27017",
		Database:       "example",
		ConnectTimeout: time.Second * 10,
		MaxRetries:     3,
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Insert document
	user := map[string]interface{}{
		"name":       "John Doe",
		"email":      "john@example.com",
		"age":        30,
		"created_at": time.Now(),
	}

	result, err := client.Collection("users").InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("insert failed: %v", err)
	}
	log.Printf("Inserted ID: %v", result.InsertedID)

	// Find document
	var foundUser map[string]interface{}
	err = client.Collection("users").FindOne(ctx, map[string]interface{}{
		"email": "john@example.com",
	}).Decode(&foundUser)
	if err != nil {
		return fmt.Errorf("find failed: %v", err)
	}
	log.Printf("Found user: %v", foundUser)

	// Update document
	updateResult, err := client.Collection("users").UpdateOne(
		ctx,
		map[string]interface{}{"email": "john@example.com"},
		map[string]interface{}{
			"$set": map[string]interface{}{
				"age": 31,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("update failed: %v", err)
	}
	log.Printf("Updated %d document(s)", updateResult.ModifiedCount)

	// Delete document
	deleteResult, err := client.Collection("users").DeleteOne(
		ctx,
		map[string]interface{}{"email": "john@example.com"},
	)
	if err != nil {
		return fmt.Errorf("delete failed: %v", err)
	}
	log.Printf("Deleted %d document(s)", deleteResult.DeletedCount)

	return nil
}

func transactionExample() error {
	client, err := mongodb.NewClient(mongodb.Config{
		URI:      "mongodb://localhost:27017",
		Database: "example",
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Start transaction
	err = client.WithTransaction(ctx, func(sessCtx context.Context) error {
		// Insert order
		order := map[string]interface{}{
			"user_id":    "123",
			"amount":     100.50,
			"status":     "pending",
			"created_at": time.Now(),
		}
		_, err := client.Collection("orders").InsertOne(sessCtx, order)
		if err != nil {
			return fmt.Errorf("failed to insert order: %v", err)
		}

		// Update user balance
		_, err = client.Collection("users").UpdateOne(
			sessCtx,
			map[string]interface{}{"_id": "123"},
			map[string]interface{}{
				"$inc": map[string]interface{}{
					"balance": -100.50,
				},
			},
		)
		if err != nil {
			return fmt.Errorf("failed to update balance: %v", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("transaction failed: %v", err)
	}

	log.Println("Transaction completed successfully")
	return nil
}

func bulkOperations() error {
	client, err := mongodb.NewClient(mongodb.Config{
		URI:      "mongodb://localhost:27017",
		Database: "example",
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Create bulk operation
	bulk := client.Collection("users").Bulk()

	// Add operations
	bulk.InsertOne(map[string]interface{}{
		"name":  "User 1",
		"email": "user1@example.com",
		"role":  "admin",
	})

	bulk.InsertOne(map[string]interface{}{
		"name":  "User 2",
		"email": "user2@example.com",
		"role":  "user",
	})

	bulk.UpdateOne(
		map[string]interface{}{"email": "user3@example.com"},
		map[string]interface{}{
			"$set": map[string]interface{}{
				"role": "manager",
			},
		},
		mongodb.UpsertOption(),
	)

	// Execute bulk operation
	result, err := bulk.Execute(ctx)
	if err != nil {
		return fmt.Errorf("bulk operation failed: %v", err)
	}

	log.Printf("Bulk operation results - Inserted: %d, Modified: %d, Deleted: %d",
		result.InsertedCount,
		result.ModifiedCount,
		result.DeletedCount,
	)

	return nil
}

func changeStreams() error {
	client, err := mongodb.NewClient(mongodb.Config{
		URI:      "mongodb://localhost:27017",
		Database: "example",
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Start change stream
	stream, err := client.Collection("users").Watch(ctx, mongodb.Pipeline{
		{{"$match": map[string]interface{}{
			"operationType": "insert",
		}}},
	})
	if err != nil {
		return fmt.Errorf("failed to create change stream: %v", err)
	}
	defer stream.Close()

	// Start goroutine to process changes
	go func() {
		for stream.Next(ctx) {
			var change map[string]interface{}
			if err := stream.Decode(&change); err != nil {
				log.Printf("Failed to decode change: %v", err)
				continue
			}
			log.Printf("Change detected: %v", change)
		}

		if err := stream.Err(); err != nil {
			log.Printf("Stream error: %v", err)
		}
	}()

	// Insert some documents to trigger changes
	for i := 0; i < 3; i++ {
		_, err := client.Collection("users").InsertOne(ctx, map[string]interface{}{
			"name":       fmt.Sprintf("User %d", i),
			"created_at": time.Now(),
		})
		if err != nil {
			log.Printf("Failed to insert document: %v", err)
		}
		time.Sleep(time.Second)
	}

	return nil
}

func errorHandling() error {
	client, err := mongodb.NewClient(mongodb.Config{
		URI:          "mongodb://invalid:27017",
		MaxRetries:   1,
		Database:     "example",
	})

	// Handle connection error
	if err != nil {
		switch err := err.(type) {
		case *mongodb.ConnectionError:
			log.Printf("Connection error: %v", err)
		default:
			log.Printf("Unknown error: %v", err)
		}
		return err
	}
	defer client.Close()

	ctx := context.Background()

	// Try invalid operation
	_, err = client.Collection("users").InsertOne(ctx, nil)
	if err != nil {
		switch err := err.(type) {
		case *mongodb.WriteError:
			log.Printf("Write error: %v", err)
		default:
			log.Printf("Unknown error: %v", err)
		}
	}

	// Try reading non-existent document
	var doc map[string]interface{}
	err = client.Collection("users").FindOne(ctx, map[string]interface{}{
		"_id": "nonexistent",
	}).Decode(&doc)
	if err != nil {
		switch err := err.(type) {
		case *mongodb.ReadError:
			log.Printf("Read error: %v", err)
		default:
			log.Printf("Unknown error: %v", err)
		}
	}

	return nil
} 