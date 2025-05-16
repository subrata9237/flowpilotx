# Makefile for local development environment

OS := $(shell uname -s)
ARCH := $(shell uname -m)

# Helm chart configurations
HELM_TIMEOUT := 900s
RELEASE_NAME := temporal
TEMPORAL_NAMESPACE := temporal
DEFAULT_NAMESPACE := default

.PHONY: setup install uninstall clean helm-deps port-forward-all register-namespace delete-temporal-pods delete-temporal-all port-forward-temporal stop-temporal-forward check-resources check-storage-resources check-temporal-statefulsets

# Main setup command
setup: minikube-setup helm-setup install port-forward-all register-namespace

# Register default namespace in Temporal
register-namespace:
	@echo "🔧 Registering default namespace in Temporal..."
	@kubectl exec -n $(TEMPORAL_NAMESPACE) deployment/temporal-admintools -- tctl --namespace default namespace register --retention 1
	@echo "✅ Default namespace registered"

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
	@echo "📦 Checking Helm repositories..."
	@if ! helm repo list | grep -q "temporalio"; then \
		echo "Adding Temporal Helm repository..."; \
		helm repo add temporalio https://go.temporal.io/helm-charts; \
	fi
	@helm repo update

# Install all components
install: verify-env install-temporal install-mongodb

install-temporal: helm-deps
	@echo "🚀 Installing Temporal..."
	@helm install $(RELEASE_NAME) temporalio/temporal \
		--namespace $(TEMPORAL_NAMESPACE) \
		--create-namespace \
		--timeout $(HELM_TIMEOUT) \
		--set server.replicaCount=1 \
		--set cassandra.config.cluster_size=1 \
		--set cassandra.replicaCount=1 \
		--set cassandra.persistence.enabled=true \
		--set cassandra.persistence.size=10Gi \
		--set prometheus.enabled=false \
		--set grafana.enabled=false \
		--set elasticsearch.enabled=true \
		--set elasticsearch.replicas=1 \
		--set elasticsearch.persistence.enabled=true \
		--set elasticsearch.persistence.size=10Gi \
		--set server.image.repository=temporalio/server \
		--set server.image.tag=latest \
		--set server.resources.requests.cpu=100m \
		--set server.resources.requests.memory=128Mi \
		--set server.resources.limits.cpu=200m \
		--set server.resources.limits.memory=256Mi

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
	@echo "Removing Temporal Helm release..."
	@helm uninstall $(RELEASE_NAME) --namespace $(TEMPORAL_NAMESPACE) || true
	@echo "Removing Temporal PVCs..."
	@kubectl delete pvc --all -n $(TEMPORAL_NAMESPACE) || true
	@echo "Removing Temporal ConfigMaps..."
	@kubectl delete configmap --all -n $(TEMPORAL_NAMESPACE) || true
	@echo "Removing Temporal Secrets..."
	@kubectl delete secret --all -n $(TEMPORAL_NAMESPACE) || true
	@echo "Removing MongoDB components..."
	@kubectl delete -f ./dependency-deployment/mongo-statefulset.yaml -n $(DEFAULT_NAMESPACE) || true
	@kubectl delete -f ./dependency-deployment/mongo-service.yaml -n $(DEFAULT_NAMESPACE) || true
	@echo "Removing PostgreSQL components..."
	@kubectl delete -f ./dependency-deployment/postgres-statefulset.yaml -n $(DEFAULT_NAMESPACE) || true
	@kubectl delete -f ./dependency-deployment/postgres-service.yaml -n $(DEFAULT_NAMESPACE) || true
	@kubectl delete -f ./dependency-deployment/postgres-pvc.yaml -n $(DEFAULT_NAMESPACE) || true
	@kubectl delete -f ./dependency-deployment/postgres-init-config.yaml -n $(DEFAULT_NAMESPACE) || true
	@echo "✅ Uninstallation complete"

# Port forwarding
port-forward-all:
	@echo "🔌 Setting up port forwarding..."
	@echo "Starting port forwards in background..."
	@-pkill -f "kubectl port-forward" || true
	@sleep 1
	@nohup kubectl port-forward -n $(TEMPORAL_NAMESPACE) svc/$(RELEASE_NAME)-web 8080:8080 > /dev/null 2>&1 &
	@nohup kubectl port-forward -n $(TEMPORAL_NAMESPACE) svc/$(RELEASE_NAME)-frontend-headless 7233:7233 > /dev/null 2>&1 &
	@nohup kubectl port-forward -n $(TEMPORAL_NAMESPACE) svc/temporal-cassandra 9042:9042 > /dev/null 2>&1 &
	@nohup kubectl port-forward -n $(DEFAULT_NAMESPACE) svc/mongo-wf-storage 27017:27017 > /dev/null 2>&1 &
	@nohup kubectl port-forward -n $(DEFAULT_NAMESPACE) svc/postgres 5432:5432 > /dev/null 2>&1 &
	@echo "✅ Port forwarding started"
	@echo "📊 Services available at:"
	@echo "  - Temporal Web UI: http://localhost:8080"
	@echo "  - Temporal Frontend: localhost:7233"
	@echo "  - Cassandra: localhost:9042"
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

# Delete all pods in temporal namespace
delete-temporal-pods:
	@echo "🗑️ Deleting all pods in temporal namespace..."
	@kubectl delete pods --all -n $(TEMPORAL_NAMESPACE)
	@echo "✅ All pods in temporal namespace deleted"

# Delete all resources in temporal namespace
delete-temporal-all:
	@echo "🗑️ Deleting all resources in temporal namespace..."
	@kubectl delete all --all -n $(TEMPORAL_NAMESPACE)
	@kubectl delete pvc --all -n $(TEMPORAL_NAMESPACE)
	@kubectl delete configmap --all -n $(TEMPORAL_NAMESPACE)
	@kubectl delete secret --all -n $(TEMPORAL_NAMESPACE)
	@echo "✅ All resources in temporal namespace deleted"

# Port forwarding for Temporal
port-forward-temporal:
	@echo "🔌 Setting up Temporal port forwarding..."
	@echo "Starting Temporal port forwards in background..."
	@-pkill -f "kubectl port-forward.*temporal" || true
	@sleep 1
	@nohup kubectl port-forward -n $(TEMPORAL_NAMESPACE) svc/$(RELEASE_NAME)-web 8088:8080 > /dev/null 2>&1 &
	@nohup kubectl port-forward -n $(TEMPORAL_NAMESPACE) svc/$(RELEASE_NAME)-frontend-headless 7233:7233 > /dev/null 2>&1 &
	@echo "✅ Temporal port forwarding started"
	@echo "📊 Temporal services available at:"
	@echo "  - Temporal Web UI: http://localhost:8088"
	@echo "  - Temporal Frontend: localhost:7233"

# Stop Temporal port forwarding
stop-temporal-forward:
	@echo "🛑 Stopping Temporal port forwarding..."
	@pkill -f "kubectl port-forward.*temporal" || true
	@echo "✅ Temporal port forwarding stopped"

# Check deployment resources
check-temporal-resources:
	@echo "📊 Checking deployment resources..."
	@echo "\n🔄 Temporal namespace deployments:"
	@kubectl get deployments -n $(TEMPORAL_NAMESPACE) -o custom-columns="NAME:.metadata.name,CPU_REQUEST:.spec.template.spec.containers[0].resources.requests.cpu,CPU_LIMIT:.spec.template.spec.containers[0].resources.limits.cpu,MEMORY_REQUEST:.spec.template.spec.containers[0].resources.requests.memory,MEMORY_LIMIT:.spec.template.spec.containers[0].resources.limits.memory"
	@echo "\n🔄 Temporal namespace statefulsets:"
	@kubectl get statefulsets -n $(TEMPORAL_NAMESPACE) -o custom-columns="NAME:.metadata.name,CPU_REQUEST:.spec.template.spec.containers[0].resources.requests.cpu,CPU_LIMIT:.spec.template.spec.containers[0].resources.limits.cpu,MEMORY_REQUEST:.spec.template.spec.containers[0].resources.requests.memory,MEMORY_LIMIT:.spec.template.spec.containers[0].resources.limits.memory"
	@echo "\n🔄 Default namespace deployments:"
	@kubectl get deployments -n $(DEFAULT_NAMESPACE) -o custom-columns="NAME:.metadata.name,CPU_REQUEST:.spec.template.spec.containers[0].resources.requests.cpu,CPU_LIMIT:.spec.template.spec.containers[0].resources.limits.cpu,MEMORY_REQUEST:.spec.template.spec.containers[0].resources.requests.memory,MEMORY_LIMIT:.spec.template.spec.containers[0].resources.limits.memory"

# Check Elasticsearch and Cassandra resources
check-temporal-storage-resources:
	@echo "📊 Checking storage resources..."
	@echo "\n🔄 Elasticsearch resources:"
	@kubectl get statefulset -n $(TEMPORAL_NAMESPACE) -l app.kubernetes.io/name=elasticsearch -o custom-columns="NAME:.metadata.name,CPU_REQUEST:.spec.template.spec.containers[0].resources.requests.cpu,CPU_LIMIT:.spec.template.spec.containers[0].resources.limits.cpu,MEMORY_REQUEST:.spec.template.spec.containers[0].resources.requests.memory,MEMORY_LIMIT:.spec.template.spec.containers[0].resources.limits.memory"
	@echo "\n🔄 Cassandra resources:"
	@kubectl get statefulset -n $(TEMPORAL_NAMESPACE) -l app.kubernetes.io/name=cassandra -o custom-columns="NAME:.metadata.name,CPU_REQUEST:.spec.template.spec.containers[0].resources.requests.cpu,CPU_LIMIT:.spec.template.spec.containers[0].resources.limits.cpu,MEMORY_REQUEST:.spec.template.spec.containers[0].resources.requests.memory,MEMORY_LIMIT:.spec.template.spec.containers[0].resources.limits.memory"
	@echo "\n🔄 Storage PVCs:"
	@kubectl get pvc -n $(TEMPORAL_NAMESPACE) -o custom-columns="NAME:.metadata.name,STATUS:.status.phase,VOLUME:.spec.volumeName,CAPACITY:.spec.resources.requests.storage"

# Check all statefulset resources in temporal namespace
check-temporal-statefulsets:
	@echo "📊 Checking all statefulset resources in temporal namespace..."
	@echo "\n🔄 All StatefulSets:"
	@kubectl get statefulset -n $(TEMPORAL_NAMESPACE) -o custom-columns="NAME:.metadata.name,REPLICAS:.spec.replicas,CPU_REQUEST:.spec.template.spec.containers[0].resources.requests.cpu,CPU_LIMIT:.spec.template.spec.containers[0].resources.limits.cpu,MEMORY_REQUEST:.spec.template.spec.containers[0].resources.requests.memory,MEMORY_LIMIT:.spec.template.spec.containers[0].resources.limits.memory"
	@echo "\n🔄 StatefulSet Pods Status:"
	@kubectl get pods -n $(TEMPORAL_NAMESPACE) -l app.kubernetes.io/component=statefulset -o custom-columns="NAME:.metadata.name,STATUS:.status.phase,READY:.status.containerStatuses[0].ready,RESTARTS:.status.containerStatuses[0].restartCount"
	@echo "\n🔄 StatefulSet PVCs:"
	@kubectl get pvc -n $(TEMPORAL_NAMESPACE) -o custom-columns="NAME:.metadata.name,STATUS:.status.phase,VOLUME:.spec.volumeName,CAPACITY:.spec.resources.requests.storage"

# Uninstall only Temporal components
uninstall-temporal:
	@echo "🗑️ Uninstalling Temporal components..."
	@echo "Removing Temporal Helm release..."
	@helm uninstall $(RELEASE_NAME) --namespace $(TEMPORAL_NAMESPACE) || true
	@echo "Removing Temporal PVCs..."
	@kubectl delete pvc --all -n $(TEMPORAL_NAMESPACE) || true
	@echo "Removing Temporal ConfigMaps..."
	@kubectl delete configmap --all -n $(TEMPORAL_NAMESPACE) || true
	@echo "Removing Temporal Secrets..."
	@kubectl delete secret --all -n $(TEMPORAL_NAMESPACE) || true
	@echo "✅ Temporal uninstallation complete"

# Port forwarding for MongoDB
port-forward-mongo:
	@echo "🔌 Setting up MongoDB port forwarding..."
	@echo "Starting MongoDB port forward in background..."
	@-pkill -f "kubectl port-forward.*mongo" || true
	@sleep 1
	@nohup kubectl port-forward -n $(DEFAULT_NAMESPACE) svc/mongo-wf-storage 27017:27017 > /dev/null 2>&1 &
	@echo "✅ MongoDB port forwarding started"
	@echo "📊 MongoDB available at:"
	@echo "  - MongoDB: localhost:27017 (default namespace)"

# Stop MongoDB port forwarding
stop-mongo-forward:
	@echo "🛑 Stopping MongoDB port forwarding..."
	@pkill -f "kubectl port-forward.*mongo" || true
	@echo "✅ MongoDB port forwarding stopped"
