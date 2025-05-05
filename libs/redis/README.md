# Redis Shared Library

A production-ready Redis client library built on top of go-redis/v9 with advanced features and comprehensive configuration options.

## Features

- Connection pooling with configurable pool size
- Support for standalone and cluster modes
- Comprehensive Redis operations support:
  - Strings
  - Hashes
  - Lists
  - Sets
  - Sorted Sets
- Built-in health checks
- Automatic connection management
- TLS support
- Authentication options
- Configurable timeouts and retries

## Installation

```bash
go get github.com/flowpilotx/sharedlib/redis
```

## Configuration

```go
type Config struct {
    Mode            string        // "standalone" or "cluster"
    Addresses       []string      // Redis server addresses
    Username        string        // Redis username (optional)
    Password        string        // Redis password (optional)
    DB             int           // Database number
    PoolSize       int           // Connection pool size
    MinIdleConns   int           // Minimum number of idle connections
    MaxRetries     int           // Maximum number of retries
    ReadTimeout    time.Duration // Read timeout
    WriteTimeout   time.Duration // Write timeout
    ConnectTimeout time.Duration // Connection timeout
    TLSConfig      *tls.Config   // TLS configuration (optional)
}
```

## Usage Examples

### Initialize Client

```go
config := &redis.Config{
    Mode:      "standalone",
    Addresses: []string{"localhost:6379"},
    PoolSize:  10,
}

client, err := redis.NewClient(config)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### Basic Operations

```go
// String operations
err := client.Set("key", "value", 0)
value, err := client.Get("key")

// Hash operations
err := client.HSet("hash", "field", "value")
value, err := client.HGet("hash", "field")

// List operations
err := client.LPush("list", "value")
values, err := client.LRange("list", 0, -1)

// Set operations
err := client.SAdd("set", "member")
members, err := client.SMembers("set")

// Sorted Set operations
err := client.ZAdd("sorted", &redis.Z{Score: 1, Member: "member"})
members, err := client.ZRange("sorted", 0, -1)
```

### Health Check

```go
err := client.Ping()
if err != nil {
    log.Printf("Redis health check failed: %v", err)
}
```

## Error Handling

The library provides comprehensive error handling with specific error types:

- `ErrClientClosed`: Returned when attempting operations on a closed client
- `ErrInvalidConfig`: Returned when configuration is invalid
- `ErrConnectionFailed`: Returned when connection fails
- `ErrOperationTimeout`: Returned when an operation times out

## Thread Safety

All operations in this library are thread-safe and can be safely used in concurrent goroutines.

## Best Practices

1. Always close the client when done:

```go
defer client.Close()
```

2. Configure appropriate timeouts:

```go
config := &redis.Config{
    ReadTimeout:    5 * time.Second,
    WriteTimeout:   5 * time.Second,
    ConnectTimeout: 10 * time.Second,
}
```

3. Use appropriate pool size based on your application needs:

```go
config := &redis.Config{
    PoolSize:     50,
    MinIdleConns: 10,
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This library is licensed under the MIT License - see the LICENSE file for details.
