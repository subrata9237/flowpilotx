# Build variables
BUILD_DIR=../../build
BINARY_NAME=flowpilotx-worker
BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)
CONFIG_PATH=../../config.yaml

# Server configuration
HTTP_PORT?=8082
GRPC_PORT?=9002
BASE_PATH?=/worker/v1

# Development settings
GO_ENV?=development
CGO_ENABLED?=0
GO_BUILD_FLAGS?=-v
GO_TEST_FLAGS?=-v -race -count=1

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: all build clean test run stop help dev lint

## Build Commands
all: clean build

build:
	@echo "  >  Building worker service..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) GO_ENV=$(GO_ENV) go build $(GO_BUILD_FLAGS) \
		-o $(BINARY_PATH) ./cmd/main.go
	@echo "  >  Build complete: $(BINARY_PATH)"

clean:
	@echo "  >  Cleaning worker service..."
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
run: check-ports
	@echo "  >  Starting worker service..."
	@HTTP_PORT=$(HTTP_PORT) \
	GRPC_PORT=$(GRPC_PORT) \
	BASE_PATH=$(BASE_PATH) \
	GO_ENV=$(GO_ENV) \
	$(BINARY_PATH)

run-dev: build run

stop:
	@echo "  >  Stopping worker service..."
	@-pkill -f "$(BINARY_PATH)" || true
	@echo "  >  Service stopped"

## Utility Commands
check-ports:
	@echo "  >  Checking ports..."
	@if lsof -i :$(HTTP_PORT) >/dev/null 2>&1; then \
		echo "  >  Error: HTTP Port $(HTTP_PORT) is already in use"; \
		exit 1; \
	fi
	@if lsof -i :$(GRPC_PORT) >/dev/null 2>&1; then \
		echo "  >  Error: gRPC Port $(GRPC_PORT) is already in use"; \
		exit 1; \
	fi
	@echo "  >  Ports are available"

## Help Command
help:
	@echo "Worker Service Commands:"
	@echo ""
	@echo "  Build Commands:"
	@echo "    make build          - Build the worker service"
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
	@echo "  Environment Variables:"
	@echo "    HTTP_PORT          - HTTP port (default: $(HTTP_PORT))"
	@echo "    GRPC_PORT          - gRPC port (default: $(GRPC_PORT))"
	@echo "    BASE_PATH          - API base path (default: $(BASE_PATH))"
	@echo "    GO_ENV             - Go environment (default: $(GO_ENV))"
	@echo "    CGO_ENABLED        - CGO enabled flag (default: $(CGO_ENABLED))"
	@echo "    GO_BUILD_FLAGS     - Additional build flags"
	@echo "    GO_TEST_FLAGS      - Additional test flags" 