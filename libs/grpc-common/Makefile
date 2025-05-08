.PHONY: proto clean deps check-deps install-deps run-server run-client test-grpc stop-server stop-client stop-all check-port kill-port

# Variables
PROTO_DIR = proto
API_DIR = pkg/api
EXAMPLE_DIR = example
GRPC_PORT = 9000

# Tools
PROTOC = protoc
PROTOC_GEN_GO = protoc-gen-go
PROTOC_GEN_GRPC = protoc-gen-go-grpc

# Go commands
GO = go
GOGET = $(GO) install

# Check if tools are installed
HAS_PROTOC := $(shell which $(PROTOC))
HAS_PROTOC_GEN_GO := $(shell which $(PROTOC_GEN_GO))
HAS_PROTOC_GEN_GRPC := $(shell which $(PROTOC_GEN_GRPC))

# Process management
check-port:
	@echo "Checking if port is in use..."
	@if lsof -i :$(GRPC_PORT) > /dev/null; then \
		echo "Port $(GRPC_PORT) is in use"; \
		exit 1; \
	fi
	@echo "Port is available"

kill-port:
	@echo "Attempting to kill processes using port..."
	@-lsof -ti :$(GRPC_PORT) | xargs kill -9 2>/dev/null || true
	@echo "Any processes using port have been terminated"

# Build and run targets
build:
	@echo "Building server and client..."
	$(GO) build -o bin/server $(EXAMPLE_DIR)/server/main.go
	$(GO) build -o bin/client $(EXAMPLE_DIR)/client/main.go

run-server: kill-port
	@echo "Starting gRPC server on port $(GRPC_PORT)..."
	$(GO) run $(EXAMPLE_DIR)/server/main.go -grpc-port $(GRPC_PORT)

run-client:
	@echo "Running client with gRPC..."
	$(GO) run $(EXAMPLE_DIR)/client/main.go -grpc-addr "localhost:$(GRPC_PORT)" -grpc

# Stop targets
stop-server:
	@echo "Stopping server..."
	@-pkill -f "$(EXAMPLE_DIR)/server/main" || echo "No server process found"
	@-lsof -ti :$(GRPC_PORT) | xargs kill -9 2>/dev/null || true
	@echo "Server stopped"

stop-client:
	@echo "Stopping client..."
	@-pkill -f "$(EXAMPLE_DIR)/client/main.go" || echo "No client process found"

stop-all: stop-client stop-server
	@echo "All processes stopped"

# Proto targets
check-deps:
	@echo "Checking dependencies..."
	@if [ -z "$(HAS_PROTOC)" ]; then \
		echo "protoc is not installed. Please install it first: https://grpc.io/docs/protoc-installation/"; \
		exit 1; \
	fi
	@if [ -z "$(HAS_PROTOC_GEN_GO)" ]; then \
		echo "protoc-gen-go is not installed. Will install it."; \
		$(GOGET) google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
	fi
	@if [ -z "$(HAS_PROTOC_GEN_GRPC)" ]; then \
		echo "protoc-gen-go-grpc is not installed. Will install it."; \
		$(GOGET) google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest; \
	fi
	@echo "All dependencies are installed."

install-deps:
	@echo "Installing dependencies..."
	$(GOGET) google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOGET) google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Dependencies installed successfully."

install-grpc-deps:
	@echo "Installing gRPC dependencies..."
	$(GO) get -u google.golang.org/grpc
	$(GO) get -u google.golang.org/protobuf

deps: check-deps install-grpc-deps

proto: deps clean
	mkdir -p $(API_DIR)
	$(PROTOC) -I $(PROTO_DIR) \
		--go_out=$(API_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(API_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/worker/v1/*.proto
	$(GO) mod tidy

clean:
	rm -rf $(API_DIR)/*
	rm -rf bin/*

# Test targets
test-complete:
	@echo "Please run these commands in separate terminals:"
	@echo "1. make run-server     # Start the gRPC server"
	@echo "2. make run-client     # Run the gRPC client"
	@echo "\nTo stop all processes, run: make stop-all"

.DEFAULT_GOAL := test-complete 