# GoldenPipe Makefile

# Variables
APP_NAME := goldenpipe
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | awk '{print $$3}')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GoVersion=$(GO_VERSION)"

# Docker variables
DOCKER_REGISTRY := ghcr.io
DOCKER_IMAGE := $(DOCKER_REGISTRY)/$(APP_NAME)
DOCKER_TAG := $(VERSION)

# Kubernetes variables
K8S_NAMESPACE := goldenpipe-system
K8S_CONTEXT := $(shell kubectl config current-context 2>/dev/null || echo "default")

# Default target
.PHONY: help
help: ## Show this help message
	@echo "GoldenPipe - VM Golden Image Automation Microservice"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
.PHONY: deps
deps: ## Install dependencies
	cd microservice && go mod tidy && go mod download

.PHONY: test
test: ## Run tests
	cd microservice && go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	cd microservice && go test -v -coverprofile=coverage.out ./...
	cd microservice && go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint: ## Run linter
	cd microservice && golangci-lint run

.PHONY: fmt
fmt: ## Format code
	cd microservice && go fmt ./...

.PHONY: vet
vet: ## Run go vet
	cd microservice && go vet ./...

# Build targets
.PHONY: build
build: deps ## Build the application
	cd microservice && CGO_ENABLED=0 GOOS=linux go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/server

.PHONY: build-local
build-local: deps ## Build for local platform
	cd microservice && go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/server

.PHONY: build-all
build-all: deps ## Build for all platforms
	cd microservice && \
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux-amd64 ./cmd/server && \
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-windows-amd64.exe ./cmd/server && \
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-amd64 ./cmd/server && \
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-arm64 ./cmd/server

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -t $(DOCKER_IMAGE):latest ./microservice

.PHONY: docker-push
docker-push: ## Push Docker image
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

.PHONY: docker-run
docker-run: ## Run Docker container locally
	docker run -p 8080:8080 --rm $(DOCKER_IMAGE):latest

# Kubernetes targets
.PHONY: k8s-apply
k8s-apply: ## Apply Kubernetes manifests
	kubectl apply -f k8s/base/

.PHONY: k8s-delete
k8s-delete: ## Delete Kubernetes resources
	kubectl delete -f k8s/base/ --ignore-not-found=true

.PHONY: k8s-status
k8s-status: ## Check Kubernetes deployment status
	kubectl get pods -n $(K8S_NAMESPACE)
	kubectl get services -n $(K8S_NAMESPACE)
	kubectl get ingress -n $(K8S_NAMESPACE)

.PHONY: k8s-logs
k8s-logs: ## View application logs
	kubectl logs -f deployment/$(APP_NAME) -n $(K8S_NAMESPACE)

.PHONY: k8s-port-forward
k8s-port-forward: ## Port forward to the service
	kubectl port-forward service/$(APP_NAME) 8080:80 -n $(K8S_NAMESPACE)

# Operator installation targets
.PHONY: install-kubevirt
install-kubevirt: ## Install KubeVirt operator
	kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/v1.1.0/kubevirt-operator.yaml
	kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/v1.1.0/kubevirt-cr.yaml

.PHONY: install-cdi
install-cdi: ## Install CDI operator
	kubectl apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/v1.55.0/cdi-operator.yaml
	kubectl apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/v1.55.0/cdi-cr.yaml

.PHONY: install-rook-ceph
install-rook-ceph: ## Install Rook-Ceph operator
	kubectl apply -f https://raw.githubusercontent.com/rook/rook/v1.12.0/deploy/examples/crds.yaml
	kubectl apply -f https://raw.githubusercontent.com/rook/rook/v1.12.0/deploy/examples/common.yaml
	kubectl apply -f https://raw.githubusercontent.com/rook/rook/v1.12.0/deploy/examples/operator.yaml
	kubectl apply -f https://raw.githubusercontent.com/rook/rook/v1.12.0/deploy/examples/cluster.yaml

.PHONY: install-operators
install-operators: install-kubevirt install-cdi install-rook-ceph ## Install all required operators

# Development environment targets
.PHONY: dev-setup
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	@echo "Installing required tools..."
	@which kind > /dev/null || (echo "Installing kind..." && go install sigs.k8s.io/kind@v0.20.0)
	@which kubectl > /dev/null || (echo "Please install kubectl: https://kubernetes.io/docs/tasks/tools/")
	@which docker > /dev/null || (echo "Please install Docker: https://docs.docker.com/get-docker/")
	@echo "Development environment setup complete!"

.PHONY: dev-cluster
dev-cluster: ## Create local development cluster
	kind create cluster --name goldenpipe-dev --config scripts/kind-config.yaml
	kubectl cluster-info --context kind-goldenpipe-dev

.PHONY: dev-deploy
dev-deploy: docker-build ## Deploy to development cluster
	kind load docker-image $(DOCKER_IMAGE):latest --name goldenpipe-dev
	kubectl apply -f k8s/overlays/dev/
	kubectl rollout status deployment/$(APP_NAME) -n $(K8S_NAMESPACE) --timeout=300s

.PHONY: dev-cleanup
dev-cleanup: ## Clean up development cluster
	kind delete cluster --name goldenpipe-dev

# API testing targets
.PHONY: test-api
test-api: ## Test API endpoints
	@echo "Testing API endpoints..."
	@curl -s http://localhost:8080/api/v1/health | jq .
	@curl -s http://localhost:8080/api/v1/images | jq .

.PHONY: create-test-image
create-test-image: ## Create a test golden image
	@echo "Creating test golden image..."
	@curl -X POST http://localhost:8080/api/v1/images \
		-H "Content-Type: application/json" \
		-d '{"name":"test-ubuntu","os_type":"linux","base_iso_url":"https://releases.ubuntu.com/22.04/ubuntu-22.04.3-desktop-amd64.iso"}' | jq .

# Cleanup targets
.PHONY: clean
clean: ## Clean build artifacts
	rm -rf microservice/bin/
	rm -rf microservice/coverage.out
	rm -rf microservice/coverage.html

.PHONY: clean-docker
clean-docker: ## Clean Docker images
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest 2>/dev/null || true

# Release targets
.PHONY: release
release: test lint build docker-build docker-push ## Create a release

.PHONY: version
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "Docker Image: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Default target
.DEFAULT_GOAL := help
