# Quick Start Guide

This guide will help you get GoldenPipe up and running quickly on your Kubernetes cluster.

## Prerequisites

### System Requirements
- Kubernetes cluster (1.20+)
- At least 4 CPU cores and 8GB RAM
- 50GB+ free storage for golden images
- kubectl configured to access your cluster

### Required Operators
Before deploying GoldenPipe, you need to install the following operators:

1. **KubeVirt** - For VM lifecycle management
2. **CDI (Containerized Data Importer)** - For image handling
3. **Rook-Ceph** - For persistent storage (or use your preferred storage class)

## Installation

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

### 2. Deploy GoldenPipe

```bash
# Clone the repository
git clone https://github.com/your-org/goldenpipe.git
cd goldenpipe

# Deploy GoldenPipe
kubectl apply -f k8s/base/

# Wait for deployment to be ready
kubectl rollout status deployment/goldenpipe -n goldenpipe-system --timeout=300s
```

### 3. Verify Installation

```bash
# Check if GoldenPipe is running
kubectl get pods -n goldenpipe-system

# Check services
kubectl get services -n goldenpipe-system

# Test the API
kubectl port-forward service/goldenpipe 8080:80 -n goldenpipe-system
curl http://localhost:8080/api/v1/health
```

## Creating Your First Golden Image

### Using the API

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

### Using GitHub Actions

1. Go to your repository's Actions tab
2. Select "Create Golden Image" workflow
3. Click "Run workflow"
4. Fill in the required parameters:
   - **Image Name**: `ubuntu-22.04-golden`
   - **OS Type**: `linux`
   - **Base ISO URL**: `https://releases.ubuntu.com/22.04/ubuntu-22.04.3-desktop-amd64.iso`
   - **Packages**: `docker,kubectl,helm`
   - **Scripts**: `install-docker.sh,configure-k8s.sh`

## Monitoring Image Creation

```bash
# List all images
curl http://localhost:8080/api/v1/images

# Get specific image status
curl http://localhost:8080/api/v1/images/ubuntu-22.04-golden/status

# View VM logs
kubectl logs -f vm/golden-image-ubuntu-22.04-golden-abc123 -n goldenpipe-system
```

## Using Your Golden Image

Once your golden image is ready, you can use it to create VMs:

```bash
# Create a VM from the golden image
kubectl apply -f - <<EOF
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: my-vm
  namespace: goldenpipe-system
spec:
  running: true
  template:
    spec:
      domain:
        resources:
          requests:
            memory: 2Gi
            cpu: 1
        devices:
          disks:
          - name: disk0
            disk:
              bus: virtio
      volumes:
      - name: disk0
        persistentVolumeClaim:
          claimName: golden-image-ubuntu-22.04-golden
EOF
```

## Troubleshooting

### Common Issues

1. **VM fails to start**
   ```bash
   kubectl describe vm <vm-name> -n goldenpipe-system
   kubectl logs <vm-pod> -n goldenpipe-system
   ```

2. **Storage issues**
   ```bash
   kubectl get pvcs -n goldenpipe-system
   kubectl describe pvc <pvc-name> -n goldenpipe-system
   ```

3. **Image download fails**
   ```bash
   kubectl get datavolumes -n goldenpipe-system
   kubectl describe datavolume <dv-name> -n goldenpipe-system
   ```

### Getting Help

- Check the [troubleshooting guide](troubleshooting.md)
- View application logs: `kubectl logs -f deployment/goldenpipe -n goldenpipe-system`
- Open an issue on GitHub

## Next Steps

- Read the [API documentation](api-reference.md)
- Learn about [customization options](customization.md)
- Explore [advanced configurations](advanced-config.md)
- Check out [examples](examples.md)
