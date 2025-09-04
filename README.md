# GoldenPipe - VM Golden Image Automation Microservice

A Kubernetes-native microservice that automates the creation of custom VM golden images for both Linux and Windows. Built with open-source technologies including KubeVirt, CDI, Rook-Ceph, and GitHub Actions.

## 🚀 Features

- **Automated VM Creation**: Downloads and customizes base ISOs for Linux and Windows
- **Unattended Setup**: Injects cloud-init scripts for Linux and autounattend.xml for Windows
- **Software Preloading**: Preloads software and internal tools during image creation
- **Windows Sysprep**: Automatically syspreps and seals Windows images
- **Persistent Storage**: Stores golden images in Rook-Ceph persistent volumes
- **GitHub Actions Integration**: Trigger workflows via GitHub Actions for 5-minute automation
- **Multi-Cluster Support**: Reuse golden images across different Kubernetes clusters

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  GitHub Actions │───▶│  GoldenPipe API  │───▶│   KubeVirt      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │      CDI         │───▶│   Rook-Ceph     │
                       │ (Data Importer)  │    │  (Storage)      │
                       └──────────────────┘    └─────────────────┘
```

## ⚡ Quick Start

### Prerequisites
- Kubernetes cluster (1.20+)
- KubeVirt installed
- CDI (Containerized Data Importer) installed
- Rook-Ceph storage class available

### 1. Install Required Operators

```bash
# Install KubeVirt
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/v1.1.0/kubevirt-operator.yaml
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/v1.1.0/kubevirt-cr.yaml

# Install CDI
kubectl apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/v1.55.0/cdi-operator.yaml
kubectl apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/v1.55.0/cdi-cr.yaml

# Install Rook-Ceph (optional - use your preferred storage class)
kubectl apply -f https://raw.githubusercontent.com/rook/rook/v1.12.0/deploy/examples/crds.yaml
kubectl apply -f https://raw.githubusercontent.com/rook/rook/v1.12.0/deploy/examples/common.yaml
kubectl apply -f https://raw.githubusercontent.com/rook/rook/v1.12.0/deploy/examples/operator.yaml
kubectl apply -f https://raw.githubusercontent.com/rook/rook/v1.12.0/deploy/examples/cluster.yaml
```

### 2. Build and Deploy GoldenPipe

```bash
# Clone the repository
git clone https://github.com/your-org/goldenpipe.git
cd goldenpipe

# Build and deploy (recommended)
./build.sh --test --lint --deploy

# Or build locally only
./build.sh --local --test --lint

# Or use Make
make install-operators
make build docker-build
make k8s-apply
```

### 3. Verify Installation

```bash
# Check if GoldenPipe is running
kubectl get pods -n goldenpipe-system

# Test the API
kubectl port-forward service/goldenpipe 8080:80 -n goldenpipe-system
curl http://localhost:8080/api/v1/health
```

### 4. Create Your First Golden Image

```bash
# Create a Linux golden image
curl -X POST http://localhost:8080/api/v1/images \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ubuntu-22.04-golden",
    "os_type": "linux",
    "base_iso_url": "https://releases.ubuntu.com/22.04/ubuntu-22.04.3-desktop-amd64.iso",
    "customizations": {
      "packages": ["docker", "kubectl", "helm"],
      "scripts": ["install-docker.sh", "configure-k8s.sh"]
    }
  }'

# Check the status
curl http://localhost:8080/api/v1/images/ubuntu-22.04-golden/status
```

## 🛠️ Build Instructions for Cursor

### Using the Build Script (Recommended)

```bash
# Make the build script executable
chmod +x build.sh

# Build with tests and linting
./build.sh --test --lint

# Build and deploy to Kubernetes
./build.sh --test --lint --deploy

# Build for local development
./build.sh --local --test --lint

# Build with custom registry
./build.sh --registry my-registry.com --image my-goldenpipe --deploy
```

### Using Make

```bash
# Install dependencies
make deps

# Run tests
make test

# Run linting
make lint

# Build application
make build

# Build Docker image
make docker-build

# Deploy to Kubernetes
make k8s-apply

# Check deployment status
make k8s-status

# View logs
make k8s-logs
```

### Manual Build Steps

```bash
# 1. Install dependencies
cd microservice
go mod tidy
go mod download

# 2. Run tests
go test -v ./...

# 3. Build application
CGO_ENABLED=0 GOOS=linux go build -o bin/goldenpipe ./cmd/server

# 4. Build Docker image
docker build -t goldenpipe:latest ./microservice

# 5. Deploy to Kubernetes
kubectl apply -f k8s/base/

# 6. Check deployment
kubectl rollout status deployment/goldenpipe -n goldenpipe-system
```

## 📁 Project Structure

```
GoldenPipe/
├── microservice/           # Main Go application
│   ├── cmd/               # Application entry points
│   ├── internal/          # Internal packages
│   │   ├── api/          # REST API handlers
│   │   ├── kubevirt/     # KubeVirt integration
│   │   ├── cdi/          # CDI integration
│   │   ├── storage/      # Storage management
│   │   └── vm/           # VM lifecycle management
│   ├── pkg/              # Public packages
│   ├── configs/          # Configuration files
│   └── Dockerfile        # Container image
├── k8s/                  # Kubernetes manifests
│   ├── base/             # Base configurations
│   ├── overlays/         # Environment-specific overlays
│   └── operators/        # Operator installations
├── scripts/              # Automation scripts
│   ├── linux/           # Linux-specific scripts
│   └── windows/         # Windows-specific scripts
├── .github/             # GitHub Actions
│   └── workflows/       # CI/CD workflows
├── docs/                # Documentation
├── build.sh             # Build script
├── Makefile             # Make targets
└── README.md            # This file
```

## 🔧 Development

### Local Development Setup

```bash
# Set up development environment
make dev-setup

# Create local Kubernetes cluster
make dev-cluster

# Deploy to development cluster
make dev-deploy

# Test API endpoints
make test-api
```

### Environment Variables

```bash
# Required
KUBECONFIG_PATH=/path/to/kubeconfig
STORAGE_CLASS=rook-ceph-block
NAMESPACE=goldenpipe-system

# Optional
LOG_LEVEL=info
API_PORT=8080
MAX_CONCURRENT_VMS=5
```

## 📚 Documentation

- [Quick Start Guide](docs/quick-start.md)
- [API Reference](docs/api-reference.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Development Guide](DEVELOPMENT.md)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Run the test suite: `make test`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🆘 Support

- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: Check the [docs](docs/) directory
- **Examples**: Look at the [scripts](scripts/) directory for automation examples
