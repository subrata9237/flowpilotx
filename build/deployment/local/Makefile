# Makefile for local development environment

OS := $(shell uname -s)
ARCH := $(shell uname -m)

# Helm chart configurations
HELM_TIMEOUT := 900s
RELEASE_NAME := temporal
TEMPORAL_NAMESPACE := temporal
DEFAULT_NAMESPACE := default

.PHONY: setup install uninstall clean helm-deps port-forward-all

# Main setup command
setup: minikube-setup helm-setup install port-forward-all

# Install required tools
install-tools:
	@echo "🔧 Installing required tools..."
	@if ! command -v minikube >/dev/null 2>&1; then \
		echo "Installing Minikube..."; \
		brew install minikube; \
	fi
	@if ! command -v kubectl >/dev/null 2>&1; then \
		echo "Installing kubectl..."; \
		brew install kubectl; \
	fi
	@if ! command -v helm >/dev/null 2>&1; then \
		echo "Installing Helm..."; \
		brew install helm; \
	fi

# Minikube setup
minikube-setup: install-tools
	@echo "🚀 Setting up Minikube environment..."
	@echo ""
	@echo "🧹 Step 1: Cleaning up existing cluster..."
	-@minikube delete 2>/dev/null || true
	@docker rm -f minikube 2>/dev/null || true
	@echo ""
	@echo "🚀 Step 2: Starting fresh Minikube cluster..."
	@minikube start --driver=docker --cpus=6 --memory=10240 --disk-size=40g
	@echo ""
	@echo "✅ Minikube setup complete!"

# Setup Helm and create namespaces
helm-setup:
	@echo "🔧 Setting up Helm and namespaces..."
	@echo "Adding Temporal Helm repository..."
	@helm repo add temporal https://go.temporal.io/helm-charts
	@helm repo update
	@echo "Creating temporal namespace..."
	@kubectl create namespace $(TEMPORAL_NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -

# Verify tools and environment
verify-env:
	@echo "🔍 Verifying environment..."
	@if ! command -v minikube >/dev/null 2>&1; then \
		echo "❌ Minikube not found"; \
		exit 1; \
	fi
	@if ! command -v kubectl >/dev/null 2>&1; then \
		echo "❌ kubectl not found"; \
		exit 1; \
	fi
	@if ! command -v helm >/dev/null 2>&1; then \
		echo "❌ Helm not found"; \
		exit 1; \
	fi
	@if ! minikube status > /dev/null 2>&1; then \
		echo "❌ Minikube is not running"; \
		exit 1; \
	fi
	@echo "✅ Environment verified"

# Helm dependencies
helm-deps:
	@echo "📦 Adding Helm repositories..."
	@helm repo add temporal https://go.temporal.io/helm-charts
	@helm repo update

# Install all components
install: verify-env install-temporal install-mongodb

install-temporal:
	@echo "🚀 Installing Temporal..."
	@helm install $(RELEASE_NAME) temporal/temporal \
		--namespace $(TEMPORAL_NAMESPACE) \
		--create-namespace \
		--timeout $(HELM_TIMEOUT) \
		--set server.replicaCount=1 \
		--set cassandra.config.cluster_size=1 \
		--set prometheus.enabled=false \
		--set grafana.enabled=false \
		--set elasticsearch.enabled=true \
		--set elasticsearch.replicas=1 \
		--set server.image.repository=temporalio/server \
		--set server.image.tag=latest \
		--set server.resources.requests.cpu=100m \
		--set server.resources.requests.memory=128Mi \
		--set server.resources.limits.cpu=200m \
		--set server.resources.limits.memory=256Mi \

install-mongodb:
	@echo "🚀 Installing MongoDB..."
	@kubectl apply -f ./dependency-deployment/mongo-statefulset.yaml -n $(DEFAULT_NAMESPACE)
	@kubectl apply -f ./dependency-deployment/mongo-service.yaml -n $(DEFAULT_NAMESPACE)

install-postgres:
	@echo "🚀 Installing PostgreSQL..."
	@kubectl apply -f ./dependency-deployment/postgres-init-config.yaml -n $(DEFAULT_NAMESPACE)
	@kubectl apply -f ./dependency-deployment/postgres-pvc.yaml -n $(DEFAULT_NAMESPACE)
	@kubectl apply -f ./dependency-deployment/postgres-statefulset.yaml -n $(DEFAULT_NAMESPACE)
	@kubectl apply -f ./dependency-deployment/postgres-service.yaml -n $(DEFAULT_NAMESPACE)

# Uninstall all components
uninstall:
	@echo "🗑️ Uninstalling all components..."
	@helm uninstall $(RELEASE_NAME) --namespace $(TEMPORAL_NAMESPACE) || true
	@kubectl delete -f ./dependency-deployment/mongo-statefulset.yaml -n $(DEFAULT_NAMESPACE) || true
	@kubectl delete -f ./dependency-deployment/mongo-service.yaml -n $(DEFAULT_NAMESPACE) || true
	@kubectl delete -f ./dependency-deployment/postgres-statefulset.yaml -n $(DEFAULT_NAMESPACE) || true
	@kubectl delete -f ./dependency-deployment/postgres-service.yaml -n $(DEFAULT_NAMESPACE) || true
	@kubectl delete -f ./dependency-deployment/postgres-pvc.yaml -n $(DEFAULT_NAMESPACE) || true
	@kubectl delete -f ./dependency-deployment/postgres-init-config.yaml -n $(DEFAULT_NAMESPACE) || true

# Port forwarding
port-forward-all:
	@echo "🔌 Setting up port forwarding..."
	@echo "Starting port forwards in background..."
	@-pkill -f "kubectl port-forward" || true
	@sleep 1
	@nohup kubectl port-forward -n $(TEMPORAL_NAMESPACE) svc/$(RELEASE_NAME)-web 8080:8080 > /dev/null 2>&1 &
	@nohup kubectl port-forward -n $(TEMPORAL_NAMESPACE) svc/$(RELEASE_NAME)-frontend-headless 7233:7233 > /dev/null 2>&1 &
	@nohup kubectl port-forward -n $(DEFAULT_NAMESPACE) svc/mongo-wf-storage 27017:27017 > /dev/null 2>&1 &
	@nohup kubectl port-forward -n $(DEFAULT_NAMESPACE) svc/postgres 5432:5432 > /dev/null 2>&1 &
	@echo "✅ Port forwarding started"
	@echo "📊 Services available at:"
	@echo "  - Temporal Web UI: http://localhost:8080"
	@echo "  - Temporal Frontend: localhost:7233"
	@echo "  - MongoDB: localhost:27017 (default namespace)"
	@echo "  - PostgreSQL: localhost:5432 (default namespace)"

# Stop port forwarding
stop-forward:
	@echo "🛑 Stopping all port forwarding..."
	@pkill -f "kubectl port-forward" || true
	@echo "✅ All port forwarding stopped"

# Clean everything
clean: uninstall
	@echo "🧹 Cleaning up everything..."
	@minikube delete || true
	@echo "✅ Cleanup complete"

# Show status
status:
	@echo "🔄 Checking component status..."
	@echo "\n🔄 Minikube status:"
	@minikube status
	@echo "\n🔄 Helm releases in temporal namespace:"
	@helm list --namespace $(TEMPORAL_NAMESPACE)
	@echo "\n🔄 Temporal namespace resources:"
	@kubectl get all --namespace $(TEMPORAL_NAMESPACE)
	@echo "\n🔄 MongoDB and PostgreSQL resources in default namespace:"
	@kubectl get all --namespace $(DEFAULT_NAMESPACE) -l app=mongo-wf-storage
	@kubectl get all --namespace $(DEFAULT_NAMESPACE) -l app=postgres
	@echo "\n🔄 Port forwards:"
	@ps aux | grep "kubectl port-forward" | grep -v grep || echo "No port forwarding running"
