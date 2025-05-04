# Makefile to setup Minikube on Apple Silicon (M1/M2/M3)

OS := $(shell uname -s)
ARCH := $(shell uname -m)

.PHONY: setup minikube-install minikube-install-f minikube-start minikube-docker-env minikube-status minikube-stop minikube-delete minikube-deploy-mongo minikube-deploy-redis minikube-deploy-nats-jetstream clear redis-forward nats-forward mongo-forward stop-forward forward-status forward-all help install-redis-cli install-mongo-cli install-all-cli redis-ping redis-cli mongo-cli mongo-ping mongo-cli-auth mongo-create-db mongo-list-dbs mongo-cli-db

# Function to check if port is in use
check-port = @if lsof -i :$(1) > /dev/null 2>&1; then \
	echo "Port $(1) is already in use"; \
	exit 0; \
fi

# Function to check if port forwarding is already running
check-forward = @if pgrep -f "port-forward.*$(1)" > /dev/null; then \
	echo "Port forwarding for $(1) is already running"; \
	exit 0; \
fi

setup:
	@echo "🔧 Step 1: Installing tools..."
	@$(MAKE) minikube-install
	@echo ""
	@echo "🚀 Step 2: Starting Minikube..."
	@$(MAKE) minikube-start
	@echo ""
	@echo "🔗 Step 3: Configuring Docker env..."
	@$(MAKE) minikube-docker-env
	@echo ""
	@echo "📊 Step 4: Showing final status..."
	@$(MAKE) minikube-status
	@echo ""
	@echo "🐱‍🏍 Step 5: Deploying MongoDB as StatefulSet..."
	@$(MAKE) minikube-deploy-mongo
	@echo ""
	@echo "🐱‍🏍 Step 6: Deploying Redis as StatefulSet..."
	@$(MAKE) minikube-deploy-redis
	@echo ""
	@echo "🐱‍🏍 Step 7: Deploying NATS JetStream as StatefulSet..."
	@$(MAKE) minikube-deploy-nats-jetstream
	@echo ""
	
	@echo "🔌 Step 8: Setting up port forwarding..."
	@$(MAKE) port-forward-resource-all

# 🔒 Private/internal targets (prefixed with 'minikube-')

minikube-install:
ifeq ($(OS),Darwin)
	@echo "🔍 Checking Minikube installation..."
	@if ! command -v minikube >/dev/null 2>&1; then \
		echo "❌ Minikube not found. Installing..."; \
		brew install minikube; \
	else \
		echo "✅ Minikube already installed."; \
	fi

	@echo "🔍 Checking kubectl installation..."
	@if ! command -v kubectl >/dev/null 2>&1; then \
		echo "❌ kubectl not found. Installing..."; \
		brew install kubectl; \
	else \
		echo "✅ kubectl already installed."; \
	fi
else
	@echo "❌ Unsupported OS: $(OS)"
	exit 1
endif

minikube-install-f:
ifeq ($(OS),Darwin)
	@echo "🧹 Force reinstalling Minikube and kubectl..."
	@echo "🗑️ Deleting existing Minikube cluster if present..."
	@minikube delete || true
	@brew reinstall minikube
	@brew reinstall kubectl
else
	@echo "❌ Unsupported OS: $(OS)"
	exit 1
endif

minikube-start:
	@echo "🚀 Checking Minikube status..."
	@if [ "$$(minikube status --format '{{.Host}}')" = "Running" ]; then \
		echo "✅ Minikube is already running."; \
	else \
		echo "🛠️ Starting Minikube with Docker driver (4 CPUs, 8GB RAM)..."; \
		minikube start --driver=docker --container-runtime=containerd --cpus=4 --memory=8192; \
	fi

minikube-docker-env:
	@echo "🔧 Configuring Docker environment for Minikube..."
	@eval $$(minikube docker-env)
	@echo "✅ Docker environment configured."

minikube-status:
	@echo "📊 Minikube status:"
	@minikube status

minikube-stop:
	@echo "🛑 Stopping Minikube..."
	@minikube stop

minikube-delete:
	@echo "🗑️ Deleting Minikube cluster..."
	@minikube delete

minikube-deploy-mongo:
	@echo "🐱‍🏍 Deploying MongoDB StatefulSet..."
	@kubectl apply -f ./dependency-deployment/mongo-statefulset.yaml
	@kubectl apply -f ./dependency-deployment/mongo-service.yaml
	@echo "✅ MongoDB StatefulSet deployed."

minikube-deploy-redis:
	@echo "🐱‍🏍 Deploying Redis StatefulSet..."
	@kubectl apply -f ./dependency-deployment/redis-statefulset.yaml
	@kubectl apply -f ./dependency-deployment/redis-service.yaml
	@echo "✅ Redis StatefulSet deployed."

minikube-deploy-nats-jetstream:
	@echo "🐱‍🏍 Deploying NATS JetStream StatefulSet..."
	@kubectl apply -f ./dependency-deployment/nats-jetstream-statefulset.yaml
	@kubectl apply -f ./dependency-deployment/nats-service.yaml
	@echo "✅ NATS JetStream StatefulSet deployed."

# 🔥 Clear all Minikube-related resources
clear:
	@echo "🧹 Clearing everything..."
	@$(MAKE) minikube-stop
	@$(MAKE) minikube-delete
	@echo "⚠️ Resetting Docker environment variables..."
	@unset $(shell minikube docker-env --shell bash | awk '{print $$3}' | tr -d "'")
	@echo "✅ All resources cleared."
# delete all dependencies resources
delete-dependencies:
	-kubectl delete -f ./dependency-deployment/mongo-statefulset.yaml
	-kubectl delete -f ./dependency-deployment/mongo-service.yaml
	-kubectl delete -f ./dependency-deployment/redis-statefulset.yaml
	-kubectl delete -f ./dependency-deployment/redis-service.yaml
	-kubectl delete -f ./dependency-deployment/nats-jetstream-statefulset.yaml
	-kubectl delete -f ./dependency-deployment/nats-service.yaml

# Start all port forwarding in background
port-forward-resource-all:
	@echo "🔌 Starting all port forwarding in background..."
	@echo "🛑 Stopping any existing port forwarding..."
	@-pkill -f "kubectl port-forward"
	@sleep 1
	@nohup kubectl port-forward svc/redis-wf-cache 6379:6379 > /dev/null 2>&1 &
	@nohup kubectl port-forward svc/nats-jetstream 4222:4222 8222:8222 > /dev/null 2>&1 &
	@nohup kubectl port-forward svc/mongo-wf-storage 27017:27017 > /dev/null 2>&1 &
	@echo "✅ All port forwarding started. PIDs:"
	@echo "Redis: $$(pgrep -f "port-forward.*redis-wf-cache")"
	@echo "NATS: $$(pgrep -f "port-forward.*nats-jetstream")"
	@echo "MongoDB: $$(pgrep -f "port-forward.*mongo-wf-storage")"
	@echo "Use 'make forward-status' to check status or 'make stop-forward' to stop all"

# Port forward Redis for external access
redis-forward:
	@echo "🔌 Starting Redis port forwarding in background..."
	@$(call check-forward,redis-wf-cache)
	@$(call check-port,6379)
	@nohup kubectl port-forward svc/redis-wf-cache 6379:6379 > /dev/null 2>&1 &
	@echo "✅ Redis port forwarding started. PID: $$(pgrep -f "port-forward.*redis-wf-cache")"

# Port forward NATS for external access
nats-forward:
	@echo "🔌 Starting NATS port forwarding in background..."
	@$(call check-forward,nats-jetstream)
	@$(call check-port,4222)
	@$(call check-port,8222)
	@nohup kubectl port-forward svc/nats-jetstream 4222:4222 8222:8222 > /dev/null 2>&1 &
	@echo "✅ NATS port forwarding started. PID: $$(pgrep -f "port-forward.*nats-jetstream")"

# Port forward MongoDB for external access
mongo-forward:
	@echo "🔌 Starting MongoDB port forwarding in background..."
	@$(call check-forward,mongo-wf-storage)
	@$(call check-port,27017)
	@nohup kubectl port-forward svc/mongo-wf-storage 27017:27017 > /dev/null 2>&1 &
	@echo "✅ MongoDB port forwarding started. PID: $$(pgrep -f "port-forward.*mongo-wf-storage")"

# Stop all port forwarding
stop-forward:
	@echo "🛑 Stopping all port forwarding..."
	@pkill -f "kubectl port-forward"
	@echo "✅ All port forwarding stopped"

# Show status of port forwarding
forward-status:
	@echo "📊 Port forwarding status:"
	@ps aux | grep "kubectl port-forward" | grep -v grep || echo "No port forwarding running"

# Show help for all commands
help:
	@echo "📚 Available commands:"
	@echo ""
	@echo "🔧 Setup and Management:"
	@echo "  make setup                    - Complete setup (install tools, start minikube, deploy dependencies)"
	@echo "  make minikube-install         - Install Minikube and kubectl"
	@echo "  make minikube-install-f       - Force reinstall Minikube and kubectl"
	@echo "  make minikube-start           - Start Minikube cluster"
	@echo "  make minikube-docker-env      - Configure Docker environment for Minikube"
	@echo "  make minikube-status          - Show Minikube status"
	@echo "  make minikube-stop            - Stop Minikube cluster"
	@echo "  make minikube-delete          - Delete Minikube cluster"
	@echo ""
	@echo "🚀 Deployment:"
	@echo "  make minikube-deploy-mongo    - Deploy MongoDB StatefulSet"
	@echo "  make minikube-deploy-redis    - Deploy Redis StatefulSet"
	@echo "  make minikube-deploy-nats-jetstream - Deploy NATS JetStream StatefulSet"
	@echo ""
	@echo "🔌 Port Forwarding:"
	@echo "  make port-forward-resource-all - Start all port forwarding (Redis, NATS, MongoDB)"
	@echo "  make redis-forward            - Start Redis port forwarding"
	@echo "  make nats-forward             - Start NATS port forwarding"
	@echo "  make mongo-forward            - Start MongoDB port forwarding"
	@echo "  make forward-status           - Show status of all port forwarding"
	@echo "  make stop-forward             - Stop all port forwarding"
	@echo ""
	@echo "🧹 Cleanup:"
	@echo "  make clear                    - Clear all Minikube resources"
	@echo "  make delete-dependencies      - Delete all dependency resources"
	@echo ""
	@echo "❓ Help:"
	@echo "  make help                     - Show this help message"
#install  redis-cli ad mongo clii of not exist
install-redis-cli:
	@if ! command -v redis-cli >/dev/null 2>&1; then \
		echo "❌ redis-cli not found. Installing..."; \
		brew install redis; \
	fi

install-mongo-cli:
	@if ! command -v mongosh >/dev/null 2>&1; then \
		echo "❌ mongosh not found. Installing..."; \
		brew install mongosh; \
	fi

install-all-cli:
	@$(MAKE) install-redis-cli
	@$(MAKE) install-mongo-cli

# Redis CLI commands
redis-ping:
	@echo "🔍 Checking if redis-cli is installed..."
	@if ! command -v redis-cli >/dev/null 2>&1; then \
		echo "❌ redis-cli not found. Installing..."; \
		brew install redis; \
	fi
	@echo "🏓 Checking Redis connection..."
	@if ! redis-cli -h localhost -p 6379 ping > /dev/null 2>&1; then \
		echo "❌ Redis is not accessible. Starting port forwarding..."; \
		$(MAKE) stop-forward; \
		sleep 1; \
		$(MAKE) redis-forward; \
		sleep 2; \
	fi
	@echo "✅ Redis is accessible. Pinging..."
	@redis-cli -h localhost -p 6379 ping

redis-cli:
	@echo "🔍 Checking if redis-cli is installed..."
	@if ! command -v redis-cli >/dev/null 2>&1; then \
		echo "❌ redis-cli not found. Installing..."; \
		brew install redis; \
	fi
	@echo "🔌 Opening Redis CLI..."
	@if ! redis-cli -h localhost -p 6379 ping > /dev/null 2>&1; then \
		echo "❌ Redis is not accessible. Starting port forwarding..."; \
		$(MAKE) stop-forward; \
		sleep 1; \
		$(MAKE) redis-forward; \
		sleep 2; \
	fi
	@echo "✅ Redis is accessible. Opening CLI..."
	@redis-cli -h localhost -p 6379

# MongoDB CLI commands
mongo-cli:
	@echo "🔍 Checking if mongosh is installed..."
	@if ! command -v mongosh >/dev/null 2>&1; then \
		echo "❌ mongosh not found. Installing..."; \
		brew install mongosh; \
	fi
	@echo "🔌 Opening MongoDB shell..."
	@if ! mongosh --eval "db.version()" mongodb://subratasuperadmin:password123@localhost:27017/admin > /dev/null 2>&1; then \
		echo "❌ MongoDB is not accessible. Starting port forwarding..."; \
		$(MAKE) stop-forward; \
		sleep 1; \
		$(MAKE) mongo-forward; \
		sleep 2; \
	fi
	@echo "✅ MongoDB is accessible. Opening shell..."
	@mongosh mongodb://subratasuperadmin:password123@localhost:27017/admin

mongo-ping:
	@echo "🔍 Checking if mongosh is installed..."
	@if ! command -v mongosh >/dev/null 2>&1; then \
		echo "❌ mongosh not found. Installing..."; \
		brew install mongosh; \
	fi
	@echo "🏓 Checking MongoDB connection..."
	@if ! mongosh --eval "db.version()" mongodb://subratasuperadmin:password123@localhost:27017/admin > /dev/null 2>&1; then \
		echo "❌ MongoDB is not accessible. Starting port forwarding..."; \
		$(MAKE) stop-forward; \
		sleep 1; \
		$(MAKE) mongo-forward; \
		sleep 2; \
	fi
	@echo "✅ MongoDB is accessible. Testing connection..."
	@mongosh --eval "db.version()" mongodb://subratasuperadmin:password123@localhost:27017/admin

# MongoDB CLI with custom credentials
mongo-cli-auth:
	@echo "🔍 Checking if mongosh is installed..."
	@if ! command -v mongosh >/dev/null 2>&1; then \
		echo "❌ mongosh not found. Installing..."; \
		brew install mongosh; \
	fi
	@echo "🔌 Opening MongoDB shell with custom credentials..."
	@read -p "Enter username: " username; \
	read -s -p "Enter password: " password; \
	echo ""; \
	if ! mongosh --eval "db.version()" mongodb://$$username:$$password@localhost:27017/admin > /dev/null 2>&1; then \
		echo "❌ MongoDB is not accessible. Starting port forwarding..."; \
		$(MAKE) stop-forward; \
		sleep 1; \
		$(MAKE) mongo-forward; \
		sleep 2; \
	fi; \
	echo "✅ MongoDB is accessible. Opening shell..."; \
	mongosh mongodb://$$username:$$password@localhost:27017/admin

# MongoDB Database commands
mongo-create-db:
	@echo "🔍 Checking if mongosh is installed..."
	@if ! command -v mongosh >/dev/null 2>&1; then \
		echo "❌ mongosh not found. Installing..."; \
		brew install mongosh; \
	fi
	@echo "📝 Creating MongoDB database..."
	@read -p "Enter database name: " dbname; \
	if ! mongosh --eval "db.version()" mongodb://subratasuperadmin:password123@localhost:27017/admin > /dev/null 2>&1; then \
		echo "❌ MongoDB is not accessible. Starting port forwarding..."; \
		$(MAKE) stop-forward; \
		sleep 1; \
		$(MAKE) mongo-forward; \
		sleep 2; \
	fi; \
	echo "Creating database $$dbname..."; \
	mongosh --eval "db = db.getSiblingDB('$$dbname'); db.createCollection('init'); db.init.drop();" mongodb://subratasuperadmin:password123@localhost:27017/admin; \
	echo "✅ Database $$dbname created successfully"

mongo-list-dbs:
	@echo "🔍 Checking if mongosh is installed..."
	@if ! command -v mongosh >/dev/null 2>&1; then \
		echo "❌ mongosh not found. Installing..."; \
		brew install mongosh; \
	fi
	@echo "📋 Listing all databases..."
	@if ! mongosh --eval "db.version()" mongodb://subratasuperadmin:password123@localhost:27017/admin > /dev/null 2>&1; then \
		echo "❌ MongoDB is not accessible. Starting port forwarding..."; \
		$(MAKE) stop-forward; \
		sleep 1; \
		$(MAKE) mongo-forward; \
		sleep 2; \
	fi; \
	mongosh --eval "show dbs" mongodb://subratasuperadmin:password123@localhost:27017/admin

mongo-cli-db:
	@echo "🔍 Checking if mongosh is installed..."
	@if ! command -v mongosh >/dev/null 2>&1; then \
		echo "❌ mongosh not found. Installing..."; \
		brew install mongosh; \
	fi
	@echo "🔌 Opening MongoDB shell for specific database..."
	@read -p "Enter database name: " dbname; \
	if ! mongosh --eval "db.version()" mongodb://subratasuperadmin:password123@localhost:27017/admin > /dev/null 2>&1; then \
		echo "❌ MongoDB is not accessible. Starting port forwarding..."; \
		$(MAKE) stop-forward; \
		sleep 1; \
		$(MAKE) mongo-forward; \
		sleep 2; \
	fi; \
	echo "✅ Opening shell for database $$dbname..."; \
	mongosh mongodb://subratasuperadmin:password123@localhost:27017/$$dbname
