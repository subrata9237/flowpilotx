# Queue Sidecar

A Go service that listens to a RabbitMQ queue and forwards messages to a local HTTP endpoint.

## Features

- Listens to a RabbitMQ queue for messages
- Forwards messages to a configurable HTTP endpoint
- Implements retry mechanism for failed requests
- Graceful shutdown handling
- Configurable through YAML file

## Configuration

The service can be configured through the `config/config.yaml` file:

```yaml
queue:
  host: "localhost"
  port: 5672
  username: "guest"
  password: "guest"
  queue_name: "http_requests"
  exchange: ""
  routing_key: "http_requests"

http:
  target_url: "http://localhost:8080/process"
  timeout_seconds: 5
  retry_attempts: 3
  retry_delay_seconds: 1

logging:
  level: "info"
  format: "json"
```

## Building and Running

### Local Development

1. Install dependencies:

```bash
go mod download
```

2. Build the application:

```bash
go build -o queue-sidecar
```

3. Run the application:

```bash
./queue-sidecar
```

### Docker

1. Build the Docker image:

```bash
docker build -t queue-sidecar .
```

2. Run the container:

```bash
docker run -d --name queue-sidecar queue-sidecar
```

## Environment Variables

The following environment variables can be used to override configuration:

- `RABBITMQ_HOST`
- `RABBITMQ_PORT`
- `RABBITMQ_USERNAME`
- `RABBITMQ_PASSWORD`
- `RABBITMQ_QUEUE_NAME`
- `RABBITMQ_EXCHANGE`
- `RABBITMQ_ROUTING_KEY`
- `HTTP_TARGET_URL`
- `HTTP_TIMEOUT_SECONDS`
- `HTTP_RETRY_ATTEMPTS`
- `HTTP_RETRY_DELAY_SECONDS`

## Logging

The service uses structured logging with JSON format. Log levels can be configured through the configuration file.

## Health Checks

The service exposes a health check endpoint at `/health` that can be used to monitor its status.
