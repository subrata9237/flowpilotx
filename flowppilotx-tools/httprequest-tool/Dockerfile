FROM golang:1.24.2-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o httprequest-tool

# Use a minimal alpine image
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/httprequest-tool .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./httprequest-tool"] 