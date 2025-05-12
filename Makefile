# Binary name
BINARY_NAME=httprequest-tool

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GO_VERSION?=1.24.2
NODE_VERSION?=20.11.1
PYTHON_VERSION?=3.12.2
DOWNLOADS_DIR=$(HOME)/.local/flowpilotx/downloads

# Build variables
BUILD_DIR=build
VERSION?=1.0.0
BUILD_TIME=$(shell date +%FT%T%z)

# Gateway service configuration
GATEWAY_HTTP_PORT?=8080
GATEWAY_BASE_PATH?=/api/flowpilotx-gateway
GATEWAY_WS_PATH?=/ws/flowpilotx-gateway
GATEWAY_CLIENT_MODE?=ping
GATEWAY_CLIENT_MSG_COUNT?=10

# Proto related variables
GRPC_COMMON_DIR=libs/grpc-common

# OS specific variables
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
        DETECTED_ARCH := amd64
    else ifeq ($(PROCESSOR_ARCHITECTURE),x86)
        DETECTED_ARCH := 386
    else ifeq ($(PROCESSOR_ARCHITECTURE),ARM64)
        DETECTED_ARCH := arm64
    endif
else
    UNAME_S := $(shell uname -s)
    UNAME_M := $(shell uname -m)
    ifeq ($(UNAME_S),Darwin)
        DETECTED_OS := Darwin
        ifeq ($(UNAME_M),arm64)
            DETECTED_ARCH := arm64
        else
            DETECTED_ARCH := amd64
        endifx
    else
        DETECTED_OS := Linux
        ifeq ($(UNAME_M),aarch64)
            DETECTED_ARCH := arm64
        else
            DETECTED_ARCH := amd64
        endif
    endif
endif

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: all build clean test run docker-build docker-run setup-go-darwin setup-go-windows dev-setup setup-dev-env setup-node setup-python env-setup env-get env-list worker-build worker-run worker-stop worker-clean worker-test worker-lint worker-proto proto proto-clean proto-deps gateway-client gateway-client-ping gateway-client-echo gateway-client-load

all: clean build

## Development Environment Setup:
setup-dev-env: setup-golang setup-node setup-python
	@echo "  >  Development environment setup completed!"
ifeq ($(DETECTED_OS),Windows)
	@echo "  >  Please restart your terminal or run 'refreshenv' to update your environment"
else ifeq ($(DETECTED_OS),Darwin)
	@echo "  >  Please run 'source ~/.zshrc' to update your environment"
else
	@echo "  >  Please run 'source ~/.bashrc' to update your environment"
endif
	@echo "  >  Verify installations with:"
	@echo "     go version"
	@echo "     node --version"
	@echo "     python --version"

setup-golang:
	@echo "  >  Setting up Go $(GO_VERSION)..."
	@echo "  >  Detected OS: $(DETECTED_OS), Architecture: $(DETECTED_ARCH)"
	@mkdir -p $(DOWNLOADS_DIR)
ifeq ($(DETECTED_OS),Windows)
	@echo "  >  Downloading Go $(GO_VERSION) for Windows..."
	@powershell -Command "Invoke-WebRequest -Uri https://go.dev/dl/go$(GO_VERSION).windows-$(DETECTED_ARCH).zip -OutFile $(DOWNLOADS_DIR)\go$(GO_VERSION).windows-$(DETECTED_ARCH).zip"
	@powershell -Command "Expand-Archive -Path $(DOWNLOADS_DIR)\go$(GO_VERSION).windows-$(DETECTED_ARCH).zip -DestinationPath 'C:\Program Files' -Force"
	@echo "  >  Setting up Go environment variables..."
	@powershell -Command "[System.Environment]::SetEnvironmentVariable('GOROOT', 'C:\Program Files\go', 'Machine')"
	@powershell -Command "[System.Environment]::SetEnvironmentVariable('GOPATH', '$$env:USERPROFILE\go', 'User')"
	@powershell -Command "[System.Environment]::SetEnvironmentVariable('Path', [System.Environment]::GetEnvironmentVariable('Path', 'User') + ';C:\Program Files\go\bin;$$env:USERPROFILE\go\bin', 'User')"
else ifeq ($(DETECTED_OS),Darwin)
	@echo "  >  Downloading Go $(GO_VERSION) for macOS..."
	@curl -L -o $(DOWNLOADS_DIR)/go$(GO_VERSION).darwin-$(DETECTED_ARCH).tar.gz https://go.dev/dl/go$(GO_VERSION).darwin-$(DETECTED_ARCH).tar.gz
	@sudo rm -rf /usr/local/go
	@sudo tar -C /usr/local -xzf $(DOWNLOADS_DIR)/go$(GO_VERSION).darwin-$(DETECTED_ARCH).tar.gz
	@echo "  >  Setting up Go environment variables..."
	@echo "export GOROOT=/usr/local/go" >> ~/.zshrc
	@echo "export GOPATH=\$$HOME/go" >> ~/.zshrc
	@echo "export PATH=\$$GOROOT/bin:\$$GOPATH/bin:\$$PATH" >> ~/.zshrc
else
	@echo "  >  Downloading Go $(GO_VERSION) for Linux..."
	@curl -L -o $(DOWNLOADS_DIR)/go$(GO_VERSION).linux-$(DETECTED_ARCH).tar.gz https://go.dev/dl/go$(GO_VERSION).linux-$(DETECTED_ARCH).tar.gz
	@sudo rm -rf /usr/local/go
	@sudo tar -C /usr/local -xzf $(DOWNLOADS_DIR)/go$(GO_VERSION).linux-$(DETECTED_ARCH).tar.gz
	@echo "  >  Setting up Go environment variables..."
	@echo "export GOROOT=/usr/local/go" >> ~/.bashrc
	@echo "export GOPATH=\$$HOME/go" >> ~/.bashrc
	@echo "export PATH=\$$GOROOT/bin:\$$GOPATH/bin:\$$PATH" >> ~/.bashrc
endif
	@echo "  >  Installing Go development tools..."
	@go install golang.org/x/tools/gopls@latest
	@go install github.com/go-delve/delve/cmd/dlv@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install github.com/swaggo/swag/v2/cmd/swag@latest
	@echo "  >  Go $(GO_VERSION) setup completed!"

setup-node:
	@echo "  >  Setting up Node.js $(NODE_VERSION)..."
ifeq ($(DETECTED_OS),Windows)
	@echo "  >  Installing NVM for Windows..."
	@powershell -Command "Invoke-WebRequest -Uri https://github.com/coreybutler/nvm-windows/releases/download/1.1.12/nvm-setup.exe -OutFile $(DOWNLOADS_DIR)\nvm-setup.exe"
	@$(DOWNLOADS_DIR)\nvm-setup.exe /SILENT
	@nvm install $(NODE_VERSION)
	@nvm use $(NODE_VERSION)
	@npm install -g yarn typescript ts-node
else ifeq ($(DETECTED_OS),Darwin)
	@if [ ! -f /opt/homebrew/bin/brew ] && [ ! -f /usr/local/bin/brew ]; then \
		/bin/bash -c "$$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"; \
	fi
	@if [ ! -f /opt/homebrew/bin/nvm ] && [ ! -f /usr/local/bin/nvm ]; then \
		brew install nvm; \
		mkdir -p ~/.nvm; \
		echo "export NVM_DIR=\$$HOME/.nvm" >> ~/.zshrc; \
		echo "[ -s \"/opt/homebrew/opt/nvm/nvm.sh\" ] && \. \"/opt/homebrew/opt/nvm/nvm.sh\"" >> ~/.zshrc; \
		echo "[ -s \"/opt/homebrew/opt/nvm/etc/bash_completion.d/nvm\" ] && \. \"/opt/homebrew/opt/nvm/etc/bash_completion.d/nvm\"" >> ~/.zshrc; \
	fi
	@. "/opt/homebrew/opt/nvm/nvm.sh"
	@nvm install $(NODE_VERSION)
	@nvm use $(NODE_VERSION)
	@nvm alias default $(NODE_VERSION)
	@npm install -g yarn typescript ts-node
else
	@curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
	@. ~/.nvm/nvm.sh
	@nvm install $(NODE_VERSION)
	@nvm use $(NODE_VERSION)
	@nvm alias default $(NODE_VERSION)
	@npm install -g yarn typescript ts-node
endif 
	@echo "  >  Node.js $(NODE_VERSION) setup completed!"

setup-python:
	@echo "  >  Setting up Python $(PYTHON_VERSION)..."
ifeq ($(DETECTED_OS),Windows)
	@echo "  >  Installing Python $(PYTHON_VERSION) for Windows..."
	@powershell -Command "Invoke-WebRequest -Uri https://www.python.org/ftp/python/$(PYTHON_VERSION)/python-$(PYTHON_VERSION)-amd64.exe -OutFile $(DOWNLOADS_DIR)\python-installer.exe"
	@$(DOWNLOADS_DIR)\python-installer.exe /quiet InstallAllUsers=1 PrependPath=1
	@python -m pip install --upgrade pip
	@pip install poetry black mypy pylint pytest
else ifeq ($(DETECTED_OS),Darwin)
	@if [ ! -f /opt/homebrew/bin/brew ] && [ ! -f /usr/local/bin/brew ]; then \
		/bin/bash -c "$$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"; \
	fi
	@if [ ! -f /opt/homebrew/bin/pyenv ] && [ ! -f /usr/local/bin/pyenv ]; then \
		brew install pyenv; \
		echo 'export PYENV_ROOT="$$HOME/.pyenv"' >> ~/.zshrc; \
		echo 'command -v pyenv >/dev/null || export PATH="$$PYENV_ROOT/bin:$$PATH"' >> ~/.zshrc; \
		echo 'eval "$$(pyenv init -)"' >> ~/.zshrc; \
	fi
	@eval "$$(pyenv init -)"
	@pyenv install $(PYTHON_VERSION)
	@pyenv global $(PYTHON_VERSION)
	@pip install --upgrade pip
	@pip install poetry black mypy pylint pytest
else
	@curl -L https://raw.githubusercontent.com/pyenv/pyenv-installer/master/bin/pyenv-installer | bash
	@echo 'export PYENV_ROOT="$$HOME/.pyenv"' >> ~/.bashrc
	@echo 'command -v pyenv >/dev/null || export PATH="$$PYENV_ROOT/bin:$$PATH"' >> ~/.bashrc
	@echo 'eval "$$(pyenv init -)"' >> ~/.bashrc
	@eval "$$(pyenv init -)"
	@pyenv install $(PYTHON_VERSION)
	@pyenv global $(PYTHON_VERSION)
	@pip install --upgrade pip
	@pip install poetry black mypy pylint pytest
endif
	@echo "  >  Python $(PYTHON_VERSION) setup completed!"

## Environment Setup:
env-setup:
	@echo "  >  Creating setup_env.sh from .devenv..."
ifeq ($(DETECTED_OS),Windows)
	@powershell -Command "Get-Content .devenv | Where-Object { $$_ -match '^export' } | ForEach-Object { $$_.Split('#')[0].Trim() } | Out-File -FilePath setup_env.sh -Encoding UTF8"
else
	@echo '#!/bin/sh' > setup_env.sh
	@while IFS= read -r line; do \
		case "$$line" in \
			export*) \
				clean_line=$$(echo "$$line" | sed 's/#.*$$//' | sed 's/[[:space:]]*$$//'); \
				echo "$$clean_line" >> setup_env.sh; \
		esac \
	done < .devenv
	@chmod +x setup_env.sh
endif
	@echo "  >  setup_env.sh has been created"
	@echo "  >  To set environment variables, run:"
	@echo "  >  source setup_env.sh"

## Environment Commands:
env-get:
ifndef VAR
	@echo "  >  Usage: make env-get VAR=VARIABLE_NAME"
	@echo "  >  Example: make env-get VAR=FLOWPILOT_MONGODBURI"
else
ifeq ($(DETECTED_OS),Windows)
	@powershell -Command "Write-Host '  >  $(VAR)=' -NoNewline; Write-Host $$env:$(VAR)"
else
	@eval "$$(grep '^export $(VAR)=' .devenv)" && echo "  >  $(VAR)=$$$(VAR)"
endif
endif

env-list:
	@echo "  >  Listing all environment variables from setup_env.sh:"
	@cat setup_env.sh | grep "^export" | sed 's/^export //'

check-ports:
	@echo "  >  Checking if ports are available..."
	@if lsof -i :$(GRPC_PORT) >/dev/null 2>&1; then \
		echo "  >  Error: Port $(GRPC_PORT) is already in use"; \
		exit 1; \
	fi
	@if lsof -i :$(HTTP_PORT) >/dev/null 2>&1; then \
		echo "  >  Error: Port $(HTTP_PORT) is already in use"; \
		exit 1; \
	fi
	@echo "  >  Ports are available"

## Help:
help:
	@echo "FlowPilotX Commands:"
	@echo ""
	@echo "Service Commands:"
	@echo "  Gateway Service:"
	@echo "    make gateway-build    - Build gateway service"
	@echo "    make gateway-run      - Run gateway service"
	@echo "    make gateway-stop     - Stop gateway service"
	@echo "    make gateway-clean    - Clean gateway service"
	@echo "    make gateway-test     - Run gateway tests"
	@echo "    make gateway-lint     - Run gateway linters"
	@echo ""
	@echo "  Worker Service:"
	@echo "    make worker-build     - Build worker service"
	@echo "    make worker-run       - Run worker service"
	@echo "    make worker-stop      - Stop worker service"
	@echo "    make worker-clean     - Clean worker service"
	@echo "    make worker-test      - Run worker tests"
	@echo "    make worker-lint      - Run worker linters"
	@echo ""
	@echo "Combined Commands:"
	@echo "    make build            - Build all services"
	@echo "    make run-services     - Run all services"
	@echo "    make stop-services    - Stop all services"
	@echo "    make clean-services   - Clean all services"
	@echo "    make test-services    - Test all services"
	@echo ""
	@echo "Environment Variables:"
	@echo "  Gateway Service:"
	@echo "    GATEWAY_HTTP_PORT    - HTTP port (default: $(GATEWAY_HTTP_PORT))"
	@echo "    GATEWAY_PATH         - API base path (default: $(GATEWAY_PATH))"
	@echo ""
	@echo "  Worker Service:"
	@echo "    WORKER_HTTP_PORT     - HTTP port (default: $(WORKER_HTTP_PORT))"
	@echo "    WORKER_GRPC_PORT     - gRPC port (default: $(WORKER_GRPC_PORT))"
	@echo "    WORKER_PATH          - API base path (default: $(WORKER_PATH))"
	@echo "  Gateway Client Commands:"
	@echo "    make gateway-client       - Run gateway test client with custom settings"
	@echo "    make gateway-client-ping  - Run gateway client in ping mode"
	@echo "    make gateway-client-echo  - Run gateway client in echo mode"
	@echo "    make gateway-client-load  - Run gateway client in load test mode"
	@echo ""
	@echo "  Environment Variables:"
	@echo "    GATEWAY_BASE_PATH        - Gateway API base path (default: $(GATEWAY_BASE_PATH))"
	@echo "    GATEWAY_WS_PATH          - Gateway WebSocket path (default: $(GATEWAY_WS_PATH))"
	@echo "    GATEWAY_CLIENT_MODE      - Gateway test client mode (default: $(GATEWAY_CLIENT_MODE))"
	@echo "    GATEWAY_CLIENT_MSG_COUNT - Number of messages for load test (default: $(GATEWAY_CLIENT_MSG_COUNT))"

# Add to the service directories section
GATEWAY_DIR=services/flowpilotx-gateway
WORKER_DIR=services/flowpilotx-worker

# Add to the binary names section
GATEWAY_BIN=$(BUILD_DIR)/flowpilotx-gateway
WORKER_BIN=$(BUILD_DIR)/flowpilotx-worker

# Add to the ports section
GATEWAY_HTTP_PORT?=8080
WORKER_HTTP_PORT?=8082
WORKER_GRPC_PORT?=9002

# Add to the base paths section
GATEWAY_PATH?=/gateway/v1
WORKER_PATH?=/worker/v1

# Add to the build target
build: gateway-build worker-build

# Add these new gateway commands
gateway-build:
	@echo "  >  Building gateway service..."
	@$(MAKE) -C $(GATEWAY_DIR) build

gateway-run:
	@echo "  >  Starting gateway service..."
	@HTTP_PORT=$(GATEWAY_HTTP_PORT) \
	BASE_PATH=$(GATEWAY_PATH) \
	$(MAKE) -C $(GATEWAY_DIR) run

gateway-stop:
	@echo "  >  Stopping gateway service..."
	@$(MAKE) -C $(GATEWAY_DIR) stop

gateway-clean:
	@echo "  >  Cleaning gateway service..."
	@$(MAKE) -C $(GATEWAY_DIR) clean

gateway-test:
	@echo "  >  Running gateway service tests..."
	@$(MAKE) -C $(GATEWAY_DIR) test

gateway-lint:
	@echo "  >  Running gateway service linters..."
	@$(MAKE) -C $(GATEWAY_DIR) lint

# Combined service commands
run-services: gateway-run worker-run
	@echo "  >  All services started"

stop-services: gateway-stop worker-stop
	@echo "  >  All services stopped"

clean-services: gateway-clean worker-clean
	@echo "  >  All services cleaned"

test-services: gateway-test worker-test
	@echo "  >  All service tests completed"

# Add these new worker commands
worker-build:
	@echo "  >  Building worker service..."
	@$(MAKE) -C $(WORKER_DIR) build

worker-run:
	@echo "  >  Starting worker service..."
	@HTTP_PORT=$(WORKER_HTTP_PORT) \
	GRPC_PORT=$(WORKER_GRPC_PORT) \
	BASE_PATH=$(WORKER_PATH) \
	$(MAKE) -C $(WORKER_DIR) run

worker-stop:
	@echo "  >  Stopping worker service..."
	@$(MAKE) -C $(WORKER_DIR) stop

worker-clean:
	@echo "  >  Cleaning worker service..."
	@$(MAKE) -C $(WORKER_DIR) clean

worker-test:
	@echo "  >  Running worker service tests..."
	@$(MAKE) -C $(WORKER_DIR) test

worker-lint:
	@echo "  >  Running worker service linters..."
	@$(MAKE) -C $(WORKER_DIR) lint

worker-proto:
	@echo "  >  Generating worker service protobuf code..."
	@$(MAKE) -C $(WORKER_DIR) proto

## Proto Commands
proto-deps:
	@echo "  >  Installing proto dependencies..."
	@cd $(GRPC_COMMON_DIR) && $(MAKE) deps

proto-clean:
	@echo "  >  Cleaning proto generated files..."
	@cd $(GRPC_COMMON_DIR) && $(MAKE) clean

proto: proto-deps
	@echo "  >  Generating proto files..."
	@cd $(GRPC_COMMON_DIR) && $(MAKE) proto
	@echo "  >  Proto generation complete"

## Gateway Client Commands
gateway-client-build:
	@echo "  >  Building gateway test client..."
	@cd services/flowpilotx-gateway && \
	CGO_ENABLED=0 go build -v -o ../../$(BUILD_DIR)/gateway-test-client ./cmd/client/main.go
	@echo "  >  Gateway client build complete"

gateway-client: gateway-client-build
	@echo "  >  Running gateway test client..."
	@$(BUILD_DIR)/gateway-test-client \
		-addr localhost:$(GATEWAY_HTTP_PORT) \
		-base $(GATEWAY_BASE_PATH) \
		-ws $(GATEWAY_WS_PATH) \
		-mode $(GATEWAY_CLIENT_MODE) \
		-n $(GATEWAY_CLIENT_MSG_COUNT)

gateway-client-ping: GATEWAY_CLIENT_MODE=ping
gateway-client-ping: gateway-client

gateway-client-echo: GATEWAY_CLIENT_MODE=echo
gateway-client-echo: gateway-client

gateway-client-load: GATEWAY_CLIENT_MODE=load
gateway-client-load: gateway-client
