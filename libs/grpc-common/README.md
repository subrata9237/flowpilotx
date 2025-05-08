# gRPC Common Library

A combined gRPC and REST API server implementation that demonstrates how to build a service supporting both protocols using gRPC-gateway.

## Features

- Dual protocol support (gRPC and REST)
- Automatic REST endpoint generation from protobuf definitions
- Streaming support (gRPC only)
- Health check and version endpoints
- Synchronous and asynchronous execution modes
- Comprehensive process management
- Base path configuration

## Prerequisites

- Go 1.19 or later
- Protocol Buffers compiler (protoc)
- Make

## Installation

1. Install the required Go tools:
```bash
make deps
```

This will install:
- protoc-gen-go
- protoc-gen-go-grpc
- protoc-gen-grpc-gateway
- protoc-gen-openapiv2

2. Generate the protocol buffer code:
```bash
make proto
```

## Project Structure

```
libs/grpc-common/
├── proto/                  # Protocol buffer definitions
├── pkg/api/                # Generated Go code
├── example/
│   ├── server/            # Server implementation
│   └── client/            # Client implementation
├── api/swagger/           # Generated OpenAPI/Swagger specs
└── Makefile              # Build and management commands
```

## Available Endpoints

### REST Endpoints (Default: http://localhost:8080)

- `GET /example-grpc/health` - Health check
- `GET /example-grpc/version` - Version information
- `POST /example-grpc/v1/worker/execute` - Synchronous execution
- `POST /example-grpc/v1/worker/execute/async` - Asynchronous execution
- `GET /example-grpc/v1/worker/execute/async/{execution_id}` - Get async result

### gRPC Endpoints (Default: localhost:50051)

- `Execute` - Synchronous execution
- `ExecuteAsync` - Asynchronous execution
- `GetAsyncResult` - Get async execution result
- `StreamExecute` - Streaming execution (gRPC only)

## Running the Server

1. Kill any existing processes using the required ports:
```bash
make kill-ports
```

2. Start the server:
```bash
make run-server
```

The server will start and listen on:
- gRPC: port 50051
- HTTP: port 8080

## Running the Client

### REST Client
```bash
make run-client-rest
```

### gRPC Client
```bash
make run-client-grpc
```

## Testing with curl

1. Health Check:
```bash
curl http://localhost:8080/example-grpc/health
```

2. Version Info:
```bash
curl http://localhost:8080/example-grpc/version
```

3. Synchronous Execution:
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"request_id":"123","event_id":1,"message":"dGVzdCBtZXNzYWdl"}' \
  http://localhost:8080/example-grpc/v1/worker/execute
```

4. Asynchronous Execution:
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"request_id":"456","event_id":2,"message":"dGVzdCBtZXNzYWdl"}' \
  http://localhost:8080/example-grpc/v1/worker/execute/async
```

5. Get Async Result:
```bash
curl http://localhost:8080/example-grpc/v1/worker/execute/async/{execution_id}
```

## Makefile Commands

- `make deps` - Install required dependencies
- `make proto` - Generate protocol buffer code
- `make build` - Build server and client binaries
- `make run-server` - Start the combined server
- `make run-client-rest` - Run REST client example
- `make run-client-grpc` - Run gRPC client example
- `make kill-ports` - Kill processes using server ports
- `make check-ports` - Check if ports are available
- `make clean` - Clean generated files
- `make help-rest` - Show REST API usage examples

## Troubleshooting

1. Port Already in Use
```bash
make kill-ports
```

2. Missing Dependencies
```bash
make deps
```

3. Clean and Rebuild
```bash
make clean
make proto
make build
```

4. Check Server Status
```bash
curl http://localhost:8080/example-grpc/health
```

## Message Format

### Request Message
```json
{
  "request_id": "string",
  "event_id": 1,
  "message": "base64_encoded_string"
}
```

### Response Message
```json
{
  "request_id": "string",
  "event_id": 1,
  "status": "EXECUTE_STATUS_SUCCESS",
  "message": "base64_encoded_string"
}
```

## Development

1. Modify proto files in `proto/` directory
2. Generate new code:
```bash
make proto
```
3. Implement new endpoints in `example/server/main.go`
4. Update client code in `example/client/main.go`
5. Test changes:
```bash
make run-server
make run-client-rest  # or make run-client-grpc
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 