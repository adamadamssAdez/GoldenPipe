#!/bin/bash
set -e

# GoldenPipe Build Script
# This script builds and deploys the GoldenPipe microservice

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="goldenpipe"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(go version | awk '{print $3}' 2>/dev/null || echo "unknown")

# Default values
DOCKER_REGISTRY="ghcr.io"
DOCKER_IMAGE="${DOCKER_REGISTRY}/${APP_NAME}"
DOCKER_TAG="${VERSION}"
K8S_NAMESPACE="goldenpipe-system"
BUILD_LOCAL=false
DEPLOY=false
TEST=false
LINT=false

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_help() {
    cat << EOF
GoldenPipe Build Script

Usage: $0 [OPTIONS]

Options:
    -h, --help              Show this help message
    -l, --local             Build for local platform only
    -d, --deploy            Deploy to Kubernetes after building
    -t, --test              Run tests before building
    -c, --lint              Run linter before building
    -r, --registry REGISTRY Docker registry (default: ghcr.io)
    -i, --image IMAGE       Docker image name (default: goldenpipe)
    -n, --namespace NS      Kubernetes namespace (default: goldenpipe-system)
    -v, --version VERSION   Version tag (default: git describe)

Examples:
    $0                      # Build Docker image
    $0 --local              # Build for local platform
    $0 --test --lint        # Run tests and linting
    $0 --deploy             # Build and deploy to Kubernetes
    $0 --registry my-registry.com --image my-goldenpipe

EOF
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Check Go version
    GO_VER=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ $(echo "$GO_VER 1.21" | awk '{print ($1 >= $2)}') -eq 0 ]]; then
        log_error "Go version $GO_VER is too old. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Check if Docker is installed (if not building locally)
    if [[ "$BUILD_LOCAL" == false ]] && ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker."
        exit 1
    fi
    
    # Check if kubectl is installed (if deploying)
    if [[ "$DEPLOY" == true ]] && ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl."
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

run_tests() {
    log_info "Running tests..."
    
    cd microservice
    
    # Download dependencies
    go mod tidy
    go mod download
    
    # Run tests
    if go test -v ./...; then
        log_success "Tests passed"
    else
        log_error "Tests failed"
        exit 1
    fi
    
    cd ..
}

run_lint() {
    log_info "Running linter..."
    
    # Check if golangci-lint is installed
    if ! command -v golangci-lint &> /dev/null; then
        log_warning "golangci-lint is not installed. Installing..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
    fi
    
    cd microservice
    
    if golangci-lint run; then
        log_success "Linting passed"
    else
        log_error "Linting failed"
        exit 1
    fi
    
    cd ..
}

build_application() {
    log_info "Building application..."
    
    cd microservice
    
    # Create bin directory
    mkdir -p bin
    
    # Set build flags
    LDFLAGS="-ldflags -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GoVersion=${GO_VERSION}"
    
    if [[ "$BUILD_LOCAL" == true ]]; then
        log_info "Building for local platform..."
        if go build ${LDFLAGS} -o bin/${APP_NAME} ./cmd/server; then
            log_success "Local build completed"
        else
            log_error "Local build failed"
            exit 1
        fi
    else
        log_info "Building for Linux (Docker)..."
        if CGO_ENABLED=0 GOOS=linux go build ${LDFLAGS} -o bin/${APP_NAME} ./cmd/server; then
            log_success "Linux build completed"
        else
            log_error "Linux build failed"
            exit 1
        fi
    fi
    
    cd ..
}

build_docker_image() {
    if [[ "$BUILD_LOCAL" == true ]]; then
        log_info "Skipping Docker build (local build only)"
        return
    fi
    
    log_info "Building Docker image..."
    
    # Build Docker image
    if docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} -t ${DOCKER_IMAGE}:latest ./microservice; then
        log_success "Docker image built successfully"
        log_info "Image: ${DOCKER_IMAGE}:${DOCKER_TAG}"
        log_info "Image: ${DOCKER_IMAGE}:latest"
    else
        log_error "Docker build failed"
        exit 1
    fi
}

deploy_to_kubernetes() {
    log_info "Deploying to Kubernetes..."
    
    # Check if kubectl can connect to cluster
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi
    
    # Update image tag in deployment
    if [[ "$BUILD_LOCAL" == false ]]; then
        log_info "Updating deployment with new image tag..."
        sed -i.bak "s|image: goldenpipe:latest|image: ${DOCKER_IMAGE}:${DOCKER_TAG}|g" k8s/base/deployment.yaml
    fi
    
    # Apply Kubernetes manifests
    log_info "Applying Kubernetes manifests..."
    if kubectl apply -f k8s/base/; then
        log_success "Kubernetes manifests applied successfully"
    else
        log_error "Failed to apply Kubernetes manifests"
        exit 1
    fi
    
    # Wait for deployment to be ready
    log_info "Waiting for deployment to be ready..."
    if kubectl rollout status deployment/${APP_NAME} -n ${K8S_NAMESPACE} --timeout=300s; then
        log_success "Deployment is ready"
    else
        log_error "Deployment failed to become ready"
        exit 1
    fi
    
    # Show deployment status
    log_info "Deployment status:"
    kubectl get pods -n ${K8S_NAMESPACE}
    kubectl get services -n ${K8S_NAMESPACE}
    
    # Restore original deployment file
    if [[ "$BUILD_LOCAL" == false ]] && [[ -f k8s/base/deployment.yaml.bak ]]; then
        mv k8s/base/deployment.yaml.bak k8s/base/deployment.yaml
    fi
}

show_build_info() {
    log_info "Build Information:"
    echo "  Version: ${VERSION}"
    echo "  Build Time: ${BUILD_TIME}"
    echo "  Go Version: ${GO_VERSION}"
    echo "  Docker Image: ${DOCKER_IMAGE}:${DOCKER_TAG}"
    echo "  Kubernetes Namespace: ${K8S_NAMESPACE}"
    echo "  Local Build: ${BUILD_LOCAL}"
    echo "  Deploy: ${DEPLOY}"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -l|--local)
            BUILD_LOCAL=true
            shift
            ;;
        -d|--deploy)
            DEPLOY=true
            shift
            ;;
        -t|--test)
            TEST=true
            shift
            ;;
        -c|--lint)
            LINT=true
            shift
            ;;
        -r|--registry)
            DOCKER_REGISTRY="$2"
            DOCKER_IMAGE="${DOCKER_REGISTRY}/${APP_NAME}"
            shift 2
            ;;
        -i|--image)
            APP_NAME="$2"
            DOCKER_IMAGE="${DOCKER_REGISTRY}/${APP_NAME}"
            shift 2
            ;;
        -n|--namespace)
            K8S_NAMESPACE="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            DOCKER_TAG="${VERSION}"
            shift 2
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Main execution
main() {
    log_info "Starting GoldenPipe build process..."
    show_build_info
    
    check_prerequisites
    
    if [[ "$TEST" == true ]]; then
        run_tests
    fi
    
    if [[ "$LINT" == true ]]; then
        run_lint
    fi
    
    build_application
    
    if [[ "$BUILD_LOCAL" == false ]]; then
        build_docker_image
    fi
    
    if [[ "$DEPLOY" == true ]]; then
        deploy_to_kubernetes
    fi
    
    log_success "Build process completed successfully!"
    
    if [[ "$DEPLOY" == true ]]; then
        log_info "You can now access GoldenPipe at:"
        echo "  kubectl port-forward service/${APP_NAME} 8080:80 -n ${K8S_NAMESPACE}"
        echo "  curl http://localhost:8080/api/v1/health"
    fi
}

# Run main function
main "$@"
