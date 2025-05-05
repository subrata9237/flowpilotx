# MongoDB Client Wrapper

A production-ready MongoDB client wrapper for Go applications with support for connection management, CRUD operations, and advanced querying capabilities.

## Features

- Connection pooling and management
- Comprehensive CRUD operations
- Transaction support
- Aggregation pipeline support
- Change stream support
- Index management
- Bulk operations
- Error handling
- Configurable timeouts and retries
- Thread-safe operations

## Installation

```bash
go get github.com/flowpilotx/sharedlib/mongodb
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/flowpilotx/sharedlib/mongodb"
)

func main() {
    // Create a new client with default configuration
    client, err := mongodb.NewClient(nil)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Or with custom configuration
    config := &mongodb.Config{
        URI:              "mongodb://localhost:27017",
        Database:         "mydb",
        ConnectTimeout:   10 * time.Second,
        OperationTimeout: 5 * time.Second,
        MaxPoolSize:      100,
        MinPoolSize:      10,
        RetryWrites:      true,
        RetryReads:       true,
    }

    client, err = mongodb.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Basic CRUD operations
    ctx := context.Background()

    // Insert a document
    doc := map[string]interface{}{
        "name": "test",
        "value": 1,
    }
    result, err := client.InsertOne(ctx, "mycollection", doc)
    if err != nil {
        log.Fatal(err)
    }

    // Find a document
    var found map[string]interface{}
    err = client.FindOne(ctx, "mycollection", map[string]interface{}{"name": "test"}, &found)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Configuration Options

| Option           | Description                 | Default                   |
| ---------------- | --------------------------- | ------------------------- |
| URI              | MongoDB connection string   | mongodb://localhost:27017 |
| Database         | Default database name       | -                         |
| Username         | Authentication username     | -                         |
| Password         | Authentication password     | -                         |
| ConnectTimeout   | Initial connection timeout  | 10s                       |
| OperationTimeout | Operation timeout           | 5s                        |
| MaxPoolSize      | Maximum connections in pool | 100                       |
| MinPoolSize      | Minimum connections in pool | 10                        |
| RetryWrites      | Enable retryable writes     | true                      |
| RetryReads       | Enable retryable reads      | true                      |
| Direct           | Enable direct connection    | false                     |

## Advanced Usage

### Transactions

```go
result, err := client.Transaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
    // Perform multiple operations atomically
    _, err := client.InsertOne(sessCtx, "collection", doc1)
    if err != nil {
        return nil, err
    }

    _, err = client.UpdateOne(sessCtx, "collection", filter, update)
    return nil, err
})
```

### Aggregation Pipeline

```go
pipeline := []bson.M{
    {"$match": bson.M{"status": "active"}},
    {"$group": bson.M{
        "_id": "$category",
        "count": bson.M{"$sum": 1},
    }},
}

cursor, err := client.Aggregate(ctx, "collection", pipeline)
```

### Change Streams

```go
pipeline := []bson.M{}
changeStream, err := client.Watch(ctx, "collection", pipeline)
if err != nil {
    log.Fatal(err)
}
defer changeStream.Close(ctx)

for changeStream.Next(ctx) {
    var changeEvent bson.M
    if err := changeStream.Decode(&changeEvent); err != nil {
        log.Printf("Error decoding change event: %v", err)
        continue
    }
    // Handle change event
}
```

### Bulk Operations

```go
models := []mongo.WriteModel{
    mongo.NewInsertOneModel().SetDocument(doc1),
    mongo.NewUpdateOneModel().
        SetFilter(filter).
        SetUpdate(update),
    mongo.NewDeleteOneModel().
        SetFilter(deleteFilter),
}

result, err := client.BulkWrite(ctx, "collection", models)
```

## Error Handling

The library provides several predefined errors:

- `ErrClientClosed`: Returned when attempting to use a closed client
- `ErrInvalidConfig`: Returned when the configuration is invalid
- `ErrNoDocuments`: Returned when no documents are found (from mongo.ErrNoDocuments)

## Testing

To run the tests:

```bash
# Set MongoDB test URI if needed
export MONGODB_TEST_URI="mongodb://localhost:27017"

# Run tests
go test -v ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
