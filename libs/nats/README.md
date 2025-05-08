# NATS Library

A high-level NATS client library supporting pub/sub patterns, JetStream, and request-reply functionality with built-in resilience features.

## Features

- Pub/Sub messaging patterns
- JetStream support for persistent messaging
- Request-Reply pattern implementation
- Automatic reconnection handling
- Message delivery guarantees
- Message filtering and wildcards
- Queue groups support
- Message headers and metadata
- Connection pooling
- Monitoring and metrics

## Installation

```bash
go get github.com/flowpilotx/libs/nats
```

## Configuration

```go
type Config struct {
    // Connection settings
    URLs              []string
    Username          string
    Password          string
    Token             string
    ConnectTimeout    time.Duration
    MaxReconnects     int
    ReconnectWait     time.Duration
    
    // TLS settings
    TLS              *tls.Config
    
    // JetStream settings
    EnableJetStream  bool
    StreamConfig     StreamConfig
    
    // Queue settings
    QueueGroup       string
    
    // Monitoring
    EnableMetrics    bool
}

type StreamConfig struct {
    Name            string
    Subjects        []string
    Retention       RetentionPolicy
    MaxAge          time.Duration
    MaxBytes        int64
    Replicas        int
}
```

## Usage Examples

### Basic Pub/Sub

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/flowpilotx/libs/nats"
)

func main() {
    // Initialize client
    client, err := nats.NewClient(nats.Config{
        URLs: []string{"nats://localhost:4222"},
        ConnectTimeout: time.Second * 10,
        MaxReconnects: 5,
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    // Subscribe to a subject
    subscription, err := client.Subscribe("greetings", func(msg *nats.Message) {
        log.Printf("Received: %s", string(msg.Data))
    })
    if err != nil {
        log.Fatalf("Subscribe failed: %v", err)
    }
    defer subscription.Unsubscribe()
    
    // Publish a message
    err = client.Publish("greetings", []byte("Hello, NATS!"))
    if err != nil {
        log.Fatalf("Publish failed: %v", err)
    }
}
```

### JetStream Usage

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/flowpilotx/libs/nats"
)

func main() {
    client, err := nats.NewClient(nats.Config{
        URLs: []string{"nats://localhost:4222"},
        EnableJetStream: true,
        StreamConfig: nats.StreamConfig{
            Name:     "orders",
            Subjects: []string{"orders.*"},
            Retention: nats.WorkQueuePolicy,
            MaxAge:   time.Hour * 24,
            Replicas: 3,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Create consumer
    consumer, err := client.CreateConsumer("orders", &nats.ConsumerConfig{
        DeliverPolicy: nats.DeliverAll,
        AckPolicy:     nats.AckExplicit,
        MaxDeliver:    3,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Subscribe with JetStream
    subscription, err := consumer.Subscribe("orders.new", func(msg *nats.Message) {
        log.Printf("Received order: %s", string(msg.Data))
        msg.Ack()
    })
    if err != nil {
        log.Fatal(err)
    }
    defer subscription.Unsubscribe()
    
    // Publish with JetStream
    ack, err := client.PublishAsync("orders.new", []byte("new order data"))
    if err != nil {
        log.Fatal(err)
    }
    
    select {
    case <-ack.Ok():
        log.Println("Message stored in JetStream")
    case err := <-ack.Err():
        log.Printf("Store failed: %v", err)
    }
}
```

### Request-Reply Pattern

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/flowpilotx/libs/nats"
)

func main() {
    client, err := nats.NewClient(nats.Config{
        URLs: []string{"nats://localhost:4222"},
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Start reply handler
    subscription, err := client.Subscribe("service.time", func(msg *nats.Message) {
        response := time.Now().Format(time.RFC3339)
        msg.Respond([]byte(response))
    })
    if err != nil {
        log.Fatal(err)
    }
    defer subscription.Unsubscribe()
    
    // Make request
    ctx := context.Background()
    response, err := client.Request(ctx, "service.time", nil, time.Second)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Current time: %s", string(response.Data))
}
```

### Queue Groups

```go
package main

import (
    "log"
    "time"
    
    "github.com/flowpilotx/libs/nats"
)

func main() {
    client, err := nats.NewClient(nats.Config{
        URLs: []string{"nats://localhost:4222"},
        QueueGroup: "workers",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Subscribe to queue group
    subscription, err := client.QueueSubscribe(
        "tasks",
        "workers",
        func(msg *nats.Message) {
            log.Printf("Processing task: %s", string(msg.Data))
        },
    )
    if err != nil {
        log.Fatal(err)
    }
    defer subscription.Unsubscribe()
    
    // Publish tasks
    for i := 0; i < 10; i++ {
        task := fmt.Sprintf("task-%d", i)
        err = client.Publish("tasks", []byte(task))
        if err != nil {
            log.Printf("Failed to publish task: %v", err)
        }
    }
}
```

## Error Handling

```go
switch err := err.(type) {
case *nats.ConnectionError:
    log.Printf("Connection error: %v", err)
case *nats.SubscriptionError:
    log.Printf("Subscription error: %v", err)
case *nats.PublishError:
    log.Printf("Publish error: %v", err)
case *nats.TimeoutError:
    log.Printf("Timeout error: %v", err)
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
# Start NATS server
docker run -d --name nats-server -p 4222:4222 nats

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
