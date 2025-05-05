# HTTP Request Tool

A Go-based HTTP request processing tool similar to n8n, designed to handle and process all kinds of HTTP requests with ease.

## Features

- Support for all HTTP methods (GET, POST, PUT, DELETE, PATCH, etc.)
- Multiple content type support:
  - JSON
  - Form data (application/x-www-form-urlencoded)
  - Multipart form data (file uploads)
  - Raw text
- Query parameter support
- Customizable request timeouts
- Response time tracking
- Structured logging
- Health check endpoint
- Docker support
- VS Code dev container support

## Prerequisites

- Go 1.24.2
- Docker (optional)
- VS Code with Remote Containers extension (optional)

## Getting Started

### Using Dev Container

1. Open the project in VS Code
2. When prompted, click "Reopen in Container"
3. The development environment will be set up automatically

### Local Development

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Run the application:
   ```bash
   go run main.go
   ```

## API Endpoints

### Health Check
```
GET /health
```
Returns the health status of the service.

### Process HTTP Request
```
POST /process
```
Process an HTTP request with the following JSON body:

```json
{
  "method": "POST",
  "url": "https://api.example.com/data",
  "headers": {
    "Content-Type": "application/json",
    "Authorization": "Bearer token"
  },
  "body": {
    "key": "value"
  },
  "query_params": {
    "param1": "value1",
    "param2": "value2"
  },
  "timeout": 30
}
```

#### Content Type Examples

1. JSON Request:
```json
{
  "method": "POST",
  "url": "https://api.example.com/data",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "name": "John Doe",
    "age": 30
  }
}
```

2. Form Data:
```json
{
  "method": "POST",
  "url": "https://api.example.com/form",
  "headers": {
    "Content-Type": "application/x-www-form-urlencoded"
  },
  "body": {
    "username": "johndoe",
    "password": "secret"
  }
}
```

3. Multipart Form Data (File Upload):
```json
{
  "method": "POST",
  "url": "https://api.example.com/upload",
  "headers": {
    "Content-Type": "multipart/form-data"
  },
  "body": {
    "file": {
      "filename": "test.txt",
      "content": "Hello, World!"
    },
    "description": "Test file"
  }
}
```

4. Raw Text:
```json
{
  "method": "POST",
  "url": "https://api.example.com/raw",
  "headers": {
    "Content-Type": "text/plain"
  },
  "body": "Hello, World!"
}
```

Response:
```json
{
  "status": 200,
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "data": "response data"
  },
  "time_elapsed": 0.123
}
```

## Development

### Building

```bash
go build -o httprequest-tool
```

### Running Tests

```bash
go test ./...
```

### Docker Build

```bash
docker build -t httprequest-tool .
```

## License

[MIT License](LICENSE) 