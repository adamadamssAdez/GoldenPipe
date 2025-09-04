# API Reference

GoldenPipe provides a REST API for managing golden images and VMs. This document describes all available endpoints.

## Base URL

The API is available at:
- Local development: `http://localhost:8080`
- Kubernetes cluster: `http://goldenpipe.goldenpipe-system.svc.cluster.local`
- Ingress: `http://goldenpipe.local` (if ingress is configured)

## Authentication

Currently, the API does not require authentication. In production environments, you should implement proper authentication and authorization.

## Content Type

All requests and responses use `application/json` content type.

## Endpoints

### Health Check

#### GET /api/v1/health

Check the health status of the GoldenPipe service.

**Response:**
```json
{
  "status": "healthy"
}
```

**Status Codes:**
- `200 OK` - Service is healthy
- `503 Service Unavailable` - Service is unhealthy

---

### Image Management

#### POST /api/v1/images

Create a new golden image.

**Request Body:**
```json
{
  "name": "ubuntu-22.04-golden",
  "os_type": "linux",
  "base_iso_url": "https://releases.ubuntu.com/22.04/ubuntu-22.04.3-desktop-amd64.iso",
  "storage_size": "20Gi",
  "cpu": 2,
  "memory": "4Gi",
  "customizations": {
    "packages": ["docker", "kubectl", "helm"],
    "scripts": ["install-docker.sh", "configure-k8s.sh"],
    "files": {
      "/etc/motd": "Welcome to GoldenPipe!\n"
    },
    "users": [
      {
        "name": "admin",
        "password": "password123",
        "groups": ["sudo"],
        "sudo": true
      }
    ],
    "ssh_keys": [
      "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC..."
    ]
  },
  "labels": {
    "environment": "production",
    "team": "platform"
  }
}
```

**Response:**
```json
{
  "message": "Golden image creation started",
  "image": {
    "name": "ubuntu-22.04-golden",
    "os_type": "linux",
    "status": "creating",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "labels": {
      "environment": "production",
      "team": "platform"
    },
    "customizations": {
      "packages": ["docker", "kubectl", "helm"],
      "scripts": ["install-docker.sh", "configure-k8s.sh"]
    },
    "pvc_name": "golden-image-ubuntu-22.04-golden",
    "vm_name": "golden-image-ubuntu-22.04-golden-abc123"
  }
}
```

**Status Codes:**
- `202 Accepted` - Image creation started
- `400 Bad Request` - Invalid request data
- `500 Internal Server Error` - Creation failed

#### GET /api/v1/images

List all golden images.

**Response:**
```json
{
  "images": [
    {
      "name": "ubuntu-22.04-golden",
      "os_type": "linux",
      "status": "ready",
      "size": "15.2Gi",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T11:45:00Z",
      "labels": {
        "environment": "production"
      },
      "pvc_name": "golden-image-ubuntu-22.04-golden"
    }
  ],
  "count": 1
}
```

**Status Codes:**
- `200 OK` - Success

#### GET /api/v1/images/{name}

Get details of a specific golden image.

**Response:**
```json
{
  "name": "ubuntu-22.04-golden",
  "os_type": "linux",
  "status": "ready",
  "size": "15.2Gi",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T11:45:00Z",
  "labels": {
    "environment": "production"
  },
  "customizations": {
    "packages": ["docker", "kubectl", "helm"],
    "scripts": ["install-docker.sh", "configure-k8s.sh"]
  },
  "pvc_name": "golden-image-ubuntu-22.04-golden"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Image not found

#### GET /api/v1/images/{name}/status

Get the status of an image creation process.

**Response:**
```json
{
  "name": "ubuntu-22.04-golden",
  "status": "creating",
  "progress": 75,
  "message": "Installing packages...",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T11:15:00Z",
  "steps": [
    {
      "name": "download_iso",
      "status": "completed",
      "message": "ISO downloaded successfully",
      "started_at": "2024-01-15T10:30:00Z",
      "ended_at": "2024-01-15T10:35:00Z"
    },
    {
      "name": "create_vm",
      "status": "completed",
      "message": "VM created successfully",
      "started_at": "2024-01-15T10:35:00Z",
      "ended_at": "2024-01-15T10:40:00Z"
    },
    {
      "name": "install_packages",
      "status": "in_progress",
      "message": "Installing packages...",
      "started_at": "2024-01-15T10:40:00Z"
    }
  ]
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Image not found

#### DELETE /api/v1/images/{name}

Delete a golden image and its associated resources.

**Response:**
```json
{
  "message": "Image deletion started"
}
```

**Status Codes:**
- `200 OK` - Deletion started
- `404 Not Found` - Image not found
- `500 Internal Server Error` - Deletion failed

---

### VM Management

#### GET /api/v1/vms

List all VMs.

**Response:**
```json
{
  "vms": [
    {
      "name": "golden-image-ubuntu-22.04-golden-abc123",
      "image_name": "ubuntu-22.04-golden",
      "status": "running",
      "ip": "10.244.1.5",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z",
      "labels": {
        "goldenpipe.io/purpose": "image-creation"
      },
      "cpu": 2,
      "memory": "4Gi"
    }
  ],
  "count": 1
}
```

**Status Codes:**
- `200 OK` - Success

#### GET /api/v1/vms/{name}

Get details of a specific VM.

**Response:**
```json
{
  "name": "golden-image-ubuntu-22.04-golden-abc123",
  "image_name": "ubuntu-22.04-golden",
  "status": "running",
  "ip": "10.244.1.5",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "labels": {
    "goldenpipe.io/purpose": "image-creation"
  },
  "cpu": 2,
  "memory": "4Gi"
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - VM not found

#### DELETE /api/v1/vms/{name}

Delete a VM.

**Response:**
```json
{
  "message": "VM deletion started"
}
```

**Status Codes:**
- `200 OK` - Deletion started
- `404 Not Found` - VM not found
- `500 Internal Server Error` - Deletion failed

---

### Metrics

#### GET /api/v1/metrics

Get system metrics.

**Response:**
```json
{
  "total_images": 5,
  "active_vms": 2,
  "failed_images": 0,
  "storage_used": "45.2Gi",
  "last_updated": "2024-01-15T12:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Success

---

## Data Types

### OSType
- `linux` - Linux operating system
- `windows` - Windows operating system

### ImageStatus
- `pending` - Image creation not started
- `creating` - Image creation in progress
- `ready` - Image is ready for use
- `failed` - Image creation failed
- `deleting` - Image deletion in progress

### VMStatus
- `pending` - VM not started
- `running` - VM is running
- `stopped` - VM is stopped
- `failed` - VM failed to start
- `deleting` - VM deletion in progress

### ImageCustomizations
```json
{
  "packages": ["string"],           // List of packages to install
  "scripts": ["string"],            // List of scripts to run
  "files": {                        // Files to create
    "path": "content"
  },
  "users": [                        // Users to create
    {
      "name": "string",
      "password": "string",         // Optional
      "groups": ["string"],         // Optional
      "sudo": boolean               // Optional
    }
  ],
  "ssh_keys": ["string"]            // SSH public keys
}
```

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message describing what went wrong"
}
```

**Common Error Codes:**
- `400 Bad Request` - Invalid request data
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Rate Limiting

Currently, there are no rate limits implemented. In production, you should implement appropriate rate limiting.

## Examples

### Create a Linux Golden Image

```bash
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
```

### Create a Windows Golden Image

```bash
curl -X POST http://localhost:8080/api/v1/images \
  -H "Content-Type: application/json" \
  -d '{
    "name": "windows-2022-golden",
    "os_type": "windows",
    "base_iso_url": "https://software-download.microsoft.com/download/sg/444969d5-f34g-4e03-ac9d-1f9786c69161/Win11_22H2_English_x64v1.iso",
    "storage_size": "50Gi",
    "cpu": 4,
    "memory": "8Gi",
    "customizations": {
      "scripts": ["install-docker.ps1", "configure-k8s.ps1"]
    }
  }'
```

### Monitor Image Creation

```bash
# Check status
curl http://localhost:8080/api/v1/images/ubuntu-22.04-golden/status

# List all images
curl http://localhost:8080/api/v1/images

# Get system metrics
curl http://localhost:8080/api/v1/metrics
```
