# FlowPilot Worker Service

A generic worker service that supports both gRPC and REST APIs using grpc-gateway.

## Features

- gRPC and REST API support
- Swagger/OpenAPI documentation
- Task management (Create, Read, Update, Delete)
- Task execution
- Support for different task types (HTTP, gRPC, Script, Workflow)
- Pagination support
- Field masks for partial updates
- Metadata and configuration flexibility

## Prerequisites

- Go 1.22 or later
- Protocol Buffers compiler (protoc)
- Required protoc plugins:
  - protoc-gen-go
  - protoc-gen-go-grpc
  - protoc-gen-grpc-gateway
  - protoc-gen-openapiv2

## Installation

1. Install the required protoc plugins:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```

2. Clone the googleapis repository (required for proto imports):

```bash
git clone https://github.com/googleapis/googleapis.git third_party/googleapis
```

3. Generate code from proto files:

```bash
make proto
```

4. Build the service:

```bash
make build
```

## Running the Service

Start the service with default ports:

```bash
./bin/server
```

Or specify custom ports:

```bash
./bin/server --grpc-port=9090 --http-port=8080
```

## API Documentation

The Swagger UI is available at `http://localhost:8080/swagger/` when the service is running.

## API Endpoints

### gRPC

The gRPC server runs on port 9090 by default and implements the following methods:

- CreateTask
- GetTask
- ListTasks
- UpdateTask
- DeleteTask
- ExecuteTask

### REST

The REST API is available on port 8080 by default with the following endpoints:

- POST /v1/tasks - Create a task
- GET /v1/tasks/{task_id} - Get a task
- GET /v1/tasks - List tasks
- PUT /v1/tasks/{task_id} - Update a task
- DELETE /v1/tasks/{task_id} - Delete a task
- POST /v1/tasks/{task_id}/execute - Execute a task

## Example Usage

### Using curl (REST API)

Create a task:
```bash
curl -X POST http://localhost:8080/v1/tasks -d '{
  "name": "example-task",
  "description": "An example task",
  "type": "TASK_TYPE_HTTP",
  "config": {
    "url": "https://example.com",
    "method": "GET"
  }
}'
```

List tasks:
```bash
curl http://localhost:8080/v1/tasks
```

### Using grpcurl (gRPC)

List tasks:
```bash
grpcurl -plaintext localhost:9090 flowpilot.worker.v1.WorkerService/ListTasks
```

## License

This project is licensed under the BSD 3-Clause License - see the LICENSE file for details. 