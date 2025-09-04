# Development Guide

## Prerequisites

### Local Development
- Go 1.21+
- Docker
- kubectl
- kind or minikube for local Kubernetes testing

### Kubernetes Cluster Requirements
- Kubernetes 1.20+
- KubeVirt v1.0+
- CDI (Containerized Data Importer) v1.55+
- Rook-Ceph storage class
- At least 4 CPU cores and 8GB RAM for VM operations

## Project Structure

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
└── docs/                # Documentation
```

## Building the Microservice

### Local Build
```bash
cd microservice
go mod tidy
go build -o goldenpipe ./cmd/server
```

### Docker Build
```bash
docker build -t goldenpipe:latest ./microservice
```

### Kubernetes Deployment
```bash
# Install operators (if not already installed)
kubectl apply -f k8s/operators/

# Deploy the application
kubectl apply -f k8s/base/

# For development with local changes
kubectl apply -f k8s/overlays/dev/
```

## Configuration

The microservice uses environment variables for configuration:

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

## API Endpoints

### Create Golden Image
```bash
POST /api/v1/images
Content-Type: application/json

{
  "name": "ubuntu-22.04-golden",
  "os_type": "linux",
  "base_iso_url": "https://releases.ubuntu.com/22.04/ubuntu-22.04.3-desktop-amd64.iso",
  "customizations": {
    "packages": ["docker", "kubectl", "helm"],
    "scripts": ["install-docker.sh", "configure-k8s.sh"]
  }
}
```

### List Images
```bash
GET /api/v1/images
```

### Get Image Status
```bash
GET /api/v1/images/{name}/status
```

### Delete Image
```bash
DELETE /api/v1/images/{name}
```

## Testing

### Unit Tests
```bash
cd microservice
go test ./...
```

### Integration Tests
```bash
# Start local Kubernetes cluster
kind create cluster --config=scripts/kind-config.yaml

# Install operators
kubectl apply -f k8s/operators/

# Run integration tests
go test -tags=integration ./...
```

### End-to-End Tests
```bash
# Deploy to test cluster
kubectl apply -f k8s/overlays/test/

# Run E2E tests
go test -tags=e2e ./tests/e2e/
```

## Development Workflow

1. **Make changes** to the microservice code
2. **Run tests** locally: `go test ./...`
3. **Build and test** container: `docker build -t goldenpipe:dev ./microservice`
4. **Deploy to dev cluster**: `kubectl apply -f k8s/overlays/dev/`
5. **Test API endpoints** using the provided examples
6. **Create PR** with changes and test results

## Debugging

### View Logs
```bash
kubectl logs -f deployment/goldenpipe -n goldenpipe-system
```

### Debug VM Creation
```bash
kubectl get vms -n goldenpipe-system
kubectl describe vm <vm-name> -n goldenpipe-system
kubectl get pvcs -n goldenpipe-system
```

### Access VM Console
```bash
kubectl virt console <vm-name> -n goldenpipe-system
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Run the test suite: `make test`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

## Troubleshooting

### Common Issues

1. **VM fails to start**: Check KubeVirt installation and resource limits
2. **Storage issues**: Verify Rook-Ceph storage class is available
3. **Image download fails**: Check network policies and proxy settings
4. **Windows sysprep fails**: Verify autounattend.xml syntax and Windows version compatibility

### Getting Help

- Check the [troubleshooting guide](docs/troubleshooting.md)
- Open an issue on GitHub
- Join our community discussions
