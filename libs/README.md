# FlowPilot Libraries

This directory contains a collection of reusable libraries for building microservices and applications. Each library is designed to be independent and follows best practices for its specific domain.

## Available Libraries

### 1. [grpc-common](./grpc-common)
A combined gRPC and REST API server implementation using gRPC-gateway. Supports both protocols with automatic REST endpoint generation from protobuf definitions.

### 2. [objectstorage](./objectstorage)
Object storage abstraction library supporting multiple providers (S3, GCS, etc.) with a unified interface for file operations.

### 3. [httprequest](./httprequest)
HTTP client library with built-in retries, circuit breaking, and observability features.

### 4. [nats](./nats)
NATS messaging library for pub/sub and request-reply patterns with support for JetStream.

### 5. [config](./config)
Configuration management library supporting multiple formats (YAML, JSON, ENV) with dynamic reloading.

### 6. [mongodb](./mongodb)
MongoDB client library with connection pooling, retries, and common operations abstraction.

### 7. [redis](./redis)
Redis client library supporting various data structures, caching patterns, and cluster mode.

## Getting Started

Each library has its own README.md with detailed documentation and examples. To use a library:

1. Navigate to the library's directory
2. Read the README.md for setup instructions
3. Check the examples directory for usage examples
4. Install dependencies if required
5. Run the examples to understand the functionality

## Common Setup

Most libraries require Go 1.19 or later. To get started:

```bash
# Clone the repository
git clone https://github.com/flowpilotx/flowpilotx.git

# Navigate to libs directory
cd flowpilotx/libs

# Choose a library
cd <library-name>

# Install dependencies
go mod download

# Run examples
go run examples/main.go
```

## Contributing

When contributing to these libraries:

1. Follow the existing code style
2. Add tests for new functionality
3. Update documentation
4. Include examples for new features
5. Ensure backward compatibility

## License

Each library is licensed under the MIT License. See individual library directories for specific terms. 