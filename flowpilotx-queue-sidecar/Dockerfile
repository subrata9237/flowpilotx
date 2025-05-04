FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o queue-sidecar

# Use a minimal alpine image
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/queue-sidecar .
COPY --from=builder /app/config/config.yaml ./config/

# Expose port (if needed for health checks)
EXPOSE 8080

# Run the application
CMD ["./queue-sidecar"]
