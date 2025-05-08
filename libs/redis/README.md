# Redis Library

A high-level Redis client library supporting various data structures, caching patterns, and cluster mode operations.

## Features

- Connection pooling
- Cluster mode support
- Pub/Sub messaging
- Lua scripting
- Pipeline operations
- Transaction support
- Automatic reconnection
- Key event notifications
- Rate limiting
- Cache patterns
  - Cache-aside
  - Write-through
  - Write-behind
  - Read-through
- Data structures
  - Strings
  - Lists
  - Sets
  - Sorted Sets
  - Hashes
  - Streams
  - Bitmaps
  - HyperLogLog
- Monitoring and metrics

## Installation

```bash
go get github.com/flowpilotx/libs/redis
```

## Configuration

```go
type Config struct {
    // Connection settings
    Addresses        []string
    Password         string
    DB              int
    ConnectTimeout   time.Duration
    ReadTimeout      time.Duration
    WriteTimeout     time.Duration
    
    // Pool settings
    PoolSize        int
    MinIdleConns    int
    MaxConnAge      time.Duration
    PoolTimeout     time.Duration
    IdleTimeout     time.Duration
    
    // Cluster settings
    EnableCluster   bool
    RouteByLatency  bool
    RouteRandomly   bool
    
    // Retry settings
    MaxRetries      int
    MinRetryBackoff time.Duration
    MaxRetryBackoff time.Duration
    
    // TLS settings
    TLS            *tls.Config
    
    // Monitoring
    EnableMetrics   bool
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
    
    "github.com/flowpilotx/libs/redis"
)

func main() {
    // Initialize client
    client, err := redis.NewClient(redis.Config{
        Addresses:      []string{"localhost:6379"},
        ConnectTimeout: time.Second * 5,
        MaxRetries:     3,
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    ctx := context.Background()
    
    // String operations
    err = client.Set(ctx, "key", "value", time.Hour)
    if err != nil {
        log.Fatalf("Set failed: %v", err)
    }
    
    val, err := client.Get(ctx, "key")
    if err != nil {
        log.Fatalf("Get failed: %v", err)
    }
    log.Printf("Value: %s", val)
    
    // Hash operations
    err = client.HSet(ctx, "user:1", map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
        "age":   30,
    })
    if err != nil {
        log.Fatalf("HSet failed: %v", err)
    }
    
    userData, err := client.HGetAll(ctx, "user:1")
    if err != nil {
        log.Fatalf("HGetAll failed: %v", err)
    }
    log.Printf("User data: %v", userData)
}
```

### Pub/Sub

```go
package main

import (
    "context"
    "log"
    
    "github.com/flowpilotx/libs/redis"
)

func main() {
    client, err := redis.NewClient(redis.Config{
        Addresses: []string{"localhost:6379"},
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    ctx := context.Background()
    
    // Subscribe to channel
    subscription := client.Subscribe(ctx, "notifications")
    defer subscription.Close()
    
    // Start message handler
    go func() {
        for msg := range subscription.Channel() {
            log.Printf("Received message: %s", msg.Payload)
        }
    }()
    
    // Publish messages
    err = client.Publish(ctx, "notifications", "Hello Redis!")
    if err != nil {
        log.Printf("Publish failed: %v", err)
    }
}
```

### Pipeline Operations

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/flowpilotx/libs/redis"
)

func main() {
    client, err := redis.NewClient(redis.Config{
        Addresses: []string{"localhost:6379"},
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    ctx := context.Background()
    
    // Create pipeline
    pipe := client.Pipeline()
    
    // Queue commands
    setCmd := pipe.Set(ctx, "key1", "value1", time.Hour)
    incrCmd := pipe.Incr(ctx, "counter")
    getCmd := pipe.Get(ctx, "key1")
    
    // Execute pipeline
    _, err = pipe.Exec(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get command results
    if err := setCmd.Err(); err != nil {
        log.Printf("Set failed: %v", err)
    }
    
    counter, err := incrCmd.Result()
    if err != nil {
        log.Printf("Incr failed: %v", err)
    }
    log.Printf("Counter: %d", counter)
    
    value, err := getCmd.Result()
    if err != nil {
        log.Printf("Get failed: %v", err)
    }
    log.Printf("Value: %s", value)
}
```

### Caching Patterns

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "time"
    
    "github.com/flowpilotx/libs/redis"
)

type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    client, err := redis.NewClient(redis.Config{
        Addresses: []string{"localhost:6379"},
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    ctx := context.Background()
    
    // Cache-aside pattern
    user, err := getUser(ctx, client, "123")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("User: %+v", user)
}

func getUser(ctx context.Context, client *redis.Client, userID string) (*User, error) {
    // Try cache first
    data, err := client.Get(ctx, "user:"+userID)
    if err == nil {
        // Cache hit
        var user User
        err = json.Unmarshal([]byte(data), &user)
        return &user, err
    }
    
    if err != redis.ErrNil {
        return nil, err
    }
    
    // Cache miss - get from database
    user := &User{
        ID:    userID,
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    // Update cache
    userData, err := json.Marshal(user)
    if err != nil {
        return nil, err
    }
    
    err = client.Set(ctx, "user:"+userID, userData, time.Hour)
    if err != nil {
        return nil, err
    }
    
    return user, nil
}
```

## Error Handling

```go
switch err := err.(type) {
case *redis.ConnectionError:
    log.Printf("Connection error: %v", err)
case *redis.CommandError:
    log.Printf("Command error: %v", err)
case *redis.TimeoutError:
    log.Printf("Timeout error: %v", err)
case *redis.CacheError:
    log.Printf("Cache error: %v", err)
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
# Start Redis
docker run -d --name redis -p 6379:6379 redis

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
