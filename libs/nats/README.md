# NATS Shared Library

A production-ready NATS client library built on top of nats.go with support for both core NATS and JetStream functionality.

## Features

- Support for both standalone and cluster modes
- Comprehensive messaging patterns:
  - Publish/Subscribe
  - Request/Reply
  - Queue Groups
  - JetStream support
- Production-ready capabilities:
  - Connection pooling
  - Automatic reconnection
  - Health monitoring
  - Connection draining
  - Message flushing
  - Comprehensive metrics
  - Connection status monitoring
- Security features:
  - TLS support with certificate validation
  - Multiple authentication methods
  - Token-based security
- Operational features:
  - Graceful shutdown
  - Connection draining
  - Message tracking
  - Error monitoring
- Context-aware operations

## Installation

```bash
go get github.com/flowpilotx/sharedlib/nats
```

## Configuration

```go
type Config struct {
    // Core settings
    Mode            string        // "standalone" or "cluster"
    Addresses       []string      // NATS server addresses

    // Authentication
    Username        string        // Authentication username
    Password        string        // Authentication password
    Token           string        // Authentication token

    // TLS Security
    EnableTLS      bool          // Enable TLS security
    TLSCertFile    string        // Client certificate file
    TLSKeyFile     string        // Client key file
    TLSCACertFile  string        // CA certificate file

    // Connection settings
    MaxReconnects   int           // Maximum reconnection attempts (-1 for unlimited)
    ReconnectWait   time.Duration // Wait time between reconnection attempts
    ConnectionName  string        // Name of the connection
    ConnectTimeout  time.Duration // Connection timeout
    PingInterval    time.Duration // Ping interval
    MaxPingOutstand int           // Maximum outstanding pings

    // JetStream
    EnableJetStream bool          // Enable JetStream
    JetStreamConfig []nats.JSOpt  // JetStream options

    // Health monitoring
    HealthCheckInterval time.Duration // Health check interval
    HealthCheckTimeout  time.Duration // Health check timeout
}
```

## Usage Examples

### Initialize Client

```go
config := &nats.Config{
    Mode:      "standalone",
    Addresses: []string{"nats://localhost:4222"},
    // Production settings
    MaxReconnects:      60,
    ReconnectWait:      2 * time.Second,
    ConnectTimeout:     5 * time.Second,
    HealthCheckInterval: 30 * time.Second,
}

client, err := nats.NewClient(config)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### Basic Messaging

```go
// Subscribe
sub, err := client.Subscribe("subject", func(msg *nats.Msg) {
    fmt.Printf("Received: %s\n", string(msg.Data))
})
if err != nil {
    log.Fatal(err)
}
defer sub.Unsubscribe()

// Publish with guaranteed delivery
err = client.Publish("subject", []byte("Hello NATS!"))
if err != nil {
    log.Fatal(err)
}
// Ensure message is sent
err = client.Flush(time.Second)
if err != nil {
    log.Fatal(err)
}
```

### Production Monitoring

```go
// Get connection statistics
stats := client.GetStats()
fmt.Printf("Messages - In: %d, Out: %d\n", stats.InMsgs, stats.OutMsgs)

// Check connection status
status := client.Status()
if status != nats.CONNECTED {
    log.Printf("Connection status: %v", status)
}

// Monitor message counts
inCount := client.InMsgCount()
outCount := client.OutMsgCount()
fmt.Printf("Message counts - In: %d, Out: %d\n", inCount, outCount)

// Check last error
if err := client.LastError(); err != nil {
    log.Printf("Last error: %v", err)
}
```

### Graceful Shutdown

```go
// Drain connection (wait for in-flight messages)
if err := client.DrainConnection(5 * time.Second); err != nil {
    log.Printf("Error draining connection: %v", err)
}

// Close the client
client.Close()
```

### Health Checks

```go
// Manual health check
if err := client.Ping(time.Second); err != nil {
    log.Printf("Health check failed: %v", err)
}
```

### Secure Connection

```go
config := &nats.Config{
    EnableTLS:     true,
    TLSCertFile:   "/path/to/cert.pem",
    TLSKeyFile:    "/path/to/key.pem",
    TLSCACertFile: "/path/to/ca.pem",
    Username:      "secure_user",
    Password:      "secure_password",
}
```

## Error Handling

The library provides specific error types for different scenarios:

```go
switch err {
case nats.ErrClientClosed:
    // Handle closed client
case nats.ErrInvalidConfig:
    // Handle invalid configuration
case nats.ErrJetStreamNotEnabled:
    // Handle JetStream not enabled
default:
    // Handle other errors
}
```

## Best Practices

1. Always configure timeouts:

```go
config := &nats.Config{
    ConnectTimeout: 5 * time.Second,
    ReconnectWait:  2 * time.Second,
}
```

2. Use health checks:

```go
config := &nats.Config{
    HealthCheckInterval: 30 * time.Second,
    HealthCheckTimeout:  2 * time.Second,
}
```

3. Configure reconnection:

```go
config := &nats.Config{
    MaxReconnects: 60,  // Retry for up to 2 minutes with 2s wait
    ReconnectWait: 2 * time.Second,
}
```

4. Implement proper shutdown:

```go
// Graceful shutdown
if err := client.DrainConnection(5 * time.Second); err != nil {
    log.Printf("Drain error: %v", err)
}
defer client.Close()
```

5. Monitor connection health:

```go
go func() {
    for {
        stats := client.GetStats()
        status := client.Status()
        // Log or metric collection
        time.Sleep(time.Minute)
    }
}()
```

## Thread Safety

All operations in this library are thread-safe and can be safely used in concurrent goroutines.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This library is licensed under the MIT License - see the LICENSE file for details.
