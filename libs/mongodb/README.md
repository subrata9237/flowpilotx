# MongoDB Library

A high-level MongoDB client library with connection pooling, automatic retries, and common operations abstraction.

## Features

- Connection pooling and management
- Automatic retries with backoff
- Transaction support
- Bulk operations
- Query builder
- Index management
- Change streams
- Aggregation pipeline builder
- GridFS support
- Monitoring and metrics
- Type-safe operations
- Context support

## Installation

```bash
go get github.com/flowpilotx/libs/mongodb
```

## Configuration

```go
type Config struct {
    // Connection settings
    URI                string
    Database          string
    ConnectTimeout    time.Duration
    OperationTimeout  time.Duration
    
    // Authentication
    Username          string
    Password          string
    AuthSource        string
    
    // Pool settings
    MaxPoolSize       uint64
    MinPoolSize       uint64
    MaxConnIdleTime   time.Duration
    
    // Retry settings
    MaxRetries        int
    RetryInterval     time.Duration
    
    // Write concern
    WriteConcern      *WriteConcern
    
    // Read preference
    ReadPreference    *ReadPreference
    
    // Monitoring
    EnableMetrics     bool
}

type WriteConcern struct {
    W                 interface{} // Number of nodes or "majority"
    J                 bool        // Journal sync
    WTimeout          time.Duration
}

type ReadPreference struct {
    Mode             string // "primary", "secondary", etc.
    MaxStaleness     time.Duration
    TagSets         []map[string]string
}
```

## Usage Examples

### Basic Operations

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/flowpilotx/libs/mongodb"
)

func main() {
    // Initialize client
    client, err := mongodb.NewClient(mongodb.Config{
        URI:            "mongodb://localhost:27017",
        Database:       "myapp",
        ConnectTimeout: time.Second * 10,
        MaxRetries:     3,
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    ctx := context.Background()
    
    // Insert document
    doc := map[string]interface{}{
        "name": "John Doe",
        "email": "john@example.com",
        "created_at": time.Now(),
    }
    
    result, err := client.Collection("users").InsertOne(ctx, doc)
    if err != nil {
        log.Fatalf("Insert failed: %v", err)
    }
    log.Printf("Inserted ID: %v", result.InsertedID)
    
    // Find document
    var user map[string]interface{}
    err = client.Collection("users").FindOne(ctx, map[string]interface{}{
        "email": "john@example.com",
    }).Decode(&user)
    if err != nil {
        log.Fatalf("Find failed: %v", err)
    }
    log.Printf("Found user: %v", user)
}
```

### Transactions

```go
package main

import (
    "context"
    "log"
    
    "github.com/flowpilotx/libs/mongodb"
)

func main() {
    client, err := mongodb.NewClient(mongodb.Config{
        URI:      "mongodb://localhost:27017",
        Database: "myapp",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    ctx := context.Background()
    
    // Start transaction
    err = client.WithTransaction(ctx, func(sessCtx context.Context) error {
        // Insert order
        order := map[string]interface{}{
            "user_id": "123",
            "amount":  100.50,
            "status": "pending",
        }
        _, err := client.Collection("orders").InsertOne(sessCtx, order)
        if err != nil {
            return err
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
        return err
    })
    
    if err != nil {
        log.Printf("Transaction failed: %v", err)
    }
}
```

### Bulk Operations

```go
package main

import (
    "context"
    "log"
    
    "github.com/flowpilotx/libs/mongodb"
)

func main() {
    client, err := mongodb.NewClient(mongodb.Config{
        URI:      "mongodb://localhost:27017",
        Database: "myapp",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    ctx := context.Background()
    
    // Create bulk operation
    bulk := client.Collection("users").Bulk()
    
    // Add operations
    bulk.InsertOne(map[string]interface{}{
        "name": "User 1",
        "role": "admin",
    })
    
    bulk.InsertOne(map[string]interface{}{
        "name": "User 2",
        "role": "user",
    })
    
    bulk.UpdateOne(
        map[string]interface{}{"name": "User 3"},
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
        log.Fatal(err)
    }
    
    log.Printf("Inserted: %d, Modified: %d, Deleted: %d",
        result.InsertedCount,
        result.ModifiedCount,
        result.DeletedCount,
    )
}
```

### Change Streams

```go
package main

import (
    "context"
    "log"
    
    "github.com/flowpilotx/libs/mongodb"
)

func main() {
    client, err := mongodb.NewClient(mongodb.Config{
        URI:      "mongodb://localhost:27017",
        Database: "myapp",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    ctx := context.Background()
    
    // Watch collection changes
    stream, err := client.Collection("users").Watch(ctx, mongodb.Pipeline{
        {{"$match": map[string]interface{}{
            "operationType": "insert",
        }}},
    })
    if err != nil {
        log.Fatal(err)
    }
    defer stream.Close()
    
    // Process changes
    for stream.Next(ctx) {
        var change map[string]interface{}
        if err := stream.Decode(&change); err != nil {
            log.Printf("Decode error: %v", err)
            continue
        }
        log.Printf("Change detected: %v", change)
    }
    
    if err := stream.Err(); err != nil {
        log.Printf("Stream error: %v", err)
    }
}
```

## Error Handling

```go
switch err := err.(type) {
case *mongodb.ConnectionError:
    log.Printf("Connection error: %v", err)
case *mongodb.WriteError:
    log.Printf("Write error: %v", err)
case *mongodb.ReadError:
    log.Printf("Read error: %v", err)
case *mongodb.TimeoutError:
    log.Printf("Timeout error: %v", err)
case *mongodb.ValidationError:
    log.Printf("Validation error: %v", err)
default:
    log.Printf("Unknown error: %v", err)
}
```

## Development

### Running Tests

```bash
go test ./...
```

### Running Examples

```bash
# Start MongoDB
docker run -d --name mongodb -p 27017:27017 mongo

# Run example
go run example/main.go
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Add tests for new functionality
4. Update documentation
5. Submit a pull request

## License

This library is licensed under the MIT License.
