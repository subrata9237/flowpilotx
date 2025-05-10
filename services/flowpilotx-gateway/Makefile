# Build variables
BUILD_DIR=../../build
BINARY_NAME=flowpilotx-gateway
BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)
CONFIG_PATH=../../config.yaml

# Server configuration
HTTP_PORT?=8080
BASE_PATH?=/api/flowpilotx-gateway
WS_PATH?=/ws/flowpilotx-gateway

# Client configuration
CLIENT_MODE?=ping
CLIENT_MSG_COUNT?=10

# Development settings
GO_ENV?=development
CGO_ENABLED?=0
GO_BUILD_FLAGS?=-v
GO_TEST_FLAGS?=-v -race -count=1

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: all build clean test run stop help dev lint run-client build-client

## Build Commands
all: clean build

build:
	@echo "  >  Building gateway service..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) GO_ENV=$(GO_ENV) go build $(GO_BUILD_FLAGS) \
		-o $(BINARY_PATH) ./cmd/server/main.go
	@echo "  >  Build complete: $(BINARY_PATH)"

clean:
	@echo "  >  Cleaning gateway service..."
	@rm -f $(BINARY_PATH)
	@go clean -cache -testcache -modcache
	@echo "  >  Clean complete"

## Development Commands
dev: build run

## Test Commands
test:
	@echo "  >  Running tests..."
	@GO_ENV=test go test $(GO_TEST_FLAGS) ./...
	@echo "  >  Tests complete"

test-coverage:
	@echo "  >  Running tests with coverage..."
	@GO_ENV=test go test $(GO_TEST_FLAGS) -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "  >  Coverage report generated: coverage.html"

## Run Commands
run: check-port
	@echo "  >  Starting gateway service..."
	@HTTP_PORT=$(HTTP_PORT) \
	BASE_PATH=$(BASE_PATH) \
	GO_ENV=$(GO_ENV) \
	$(BINARY_PATH)

run-dev: build run

stop:
	@echo "  >  Stopping worker service..."
	@-pkill -f "$(BINARY_PATH)" || true
	@echo "  >  Service stopped"

## Utility Commands
check-port:
	@echo "  >  Checking port..."
	@if lsof -i :$(HTTP_PORT) >/dev/null 2>&1; then \
		echo "  >  Error: HTTP Port $(HTTP_PORT) is already in use"; \
		exit 1; \
	fi
	@echo "  >  Port is available"

## Docker Commands
docker-build:
	@echo "  >  Building Docker image..."
	@docker build -t $(BINARY_NAME):latest .
	@echo "  >  Docker build complete"

docker-run:
	@echo "  >  Running Docker container..."
	@docker run -p $(HTTP_PORT):$(HTTP_PORT) $(BINARY_NAME):latest

## Client Commands
build-client:
	@echo "  >  Building test client..."
	@CGO_ENABLED=$(CGO_ENABLED) GO_ENV=$(GO_ENV) go build $(GO_BUILD_FLAGS) \
		-o $(BUILD_DIR)/test-client ./cmd/client/main.go
	@echo "  >  Client build complete"

run-client: build-client
	@echo "  >  Running test client..."
	@$(BUILD_DIR)/test-client \
		-addr localhost:$(HTTP_PORT) \
		-base $(BASE_PATH) \
		-ws $(WS_PATH) \
		-mode $(CLIENT_MODE) \
		-n $(CLIENT_MSG_COUNT)

run-client-ping: CLIENT_MODE=ping
run-client-ping: run-client

run-client-echo: CLIENT_MODE=echo
run-client-echo: run-client

run-client-load: CLIENT_MODE=load
run-client-load: run-client

## Help Command
help:
	@echo "Gateway Service Commands:"
	@echo ""
	@echo "  Build Commands:"
	@echo "    make build          - Build the gateway service"
	@echo "    make clean          - Clean build artifacts"
	@echo ""
	@echo "  Development Commands:"
	@echo "    make dev            - Run service in development mode"
	@echo "    make lint           - Run linters"
	@echo ""
	@echo "  Test Commands:"
	@echo "    make test           - Run tests"
	@echo "    make test-coverage  - Run tests with coverage report"
	@echo ""
	@echo "  Run Commands:"
	@echo "    make run            - Run the service"
	@echo "    make run-dev        - Build and run in development mode"
	@echo "    make stop           - Stop the service"
	@echo ""
	@echo "  Client Commands:"
	@echo "    make run-client     - Run test client with custom settings"
	@echo "    make run-client-ping - Run client in ping mode"
	@echo "    make run-client-echo - Run client in echo mode"
	@echo "    make run-client-load - Run client in load test mode"
	@echo ""
	@echo "  Docker Commands:"
	@echo "    make docker-build   - Build Docker image"
	@echo "    make docker-run     - Run Docker container"
	@echo ""
	@echo "  Environment Variables:"
	@echo "    HTTP_PORT          - HTTP port (default: $(HTTP_PORT))"
	@echo "    BASE_PATH          - API base path (default: $(BASE_PATH))"
	@echo "    WS_PATH            - WebSocket path (default: $(WS_PATH))"
	@echo "    CLIENT_MODE        - Test client mode (default: $(CLIENT_MODE))"
	@echo "    CLIENT_MSG_COUNT   - Number of messages for load test (default: $(CLIENT_MSG_COUNT))"
	@echo "    GO_ENV             - Go environment (default: $(GO_ENV))"
	@echo "    CGO_ENABLED        - CGO enabled flag (default: $(CGO_ENABLED))"
	@echo "    GO_BUILD_FLAGS     - Additional build flags"
	@echo "    GO_TEST_FLAGS      - Additional test flags" 