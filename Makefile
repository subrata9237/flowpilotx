.PHONY: build run clean

# Variables
TOOL_DIR=flowppilotx-tools/httprequest-tool
BINARY_NAME=httprequest-tool
PORT=8080

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	cd $(TOOL_DIR) && \
	CGO_ENABLED=0 GOOS=linux go build -o $(BINARY_NAME)

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	cd $(TOOL_DIR) && \
	PORT=$(PORT) ./$(BINARY_NAME)

# Build and run in development mode
dev: build run

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -f $(TOOL_DIR)/$(BINARY_NAME)

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t httprequest-tool $(TOOL_DIR)

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p $(PORT):8080 httprequest-tool

# Build and run Docker container
docker: docker-build docker-run

# Help command
help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make dev          - Build and run in development mode"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo "  make docker       - Build and run Docker container"
	@echo "  make help         - Show this help message"
