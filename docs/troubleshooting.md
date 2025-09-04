# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with GoldenPipe.

## Common Issues

### 1. VM Fails to Start

**Symptoms:**
- VM remains in `pending` or `failed` status
- No IP address assigned to VM
- VM pod shows errors

**Diagnosis:**
```bash
# Check VM status
kubectl get vms -n goldenpipe-system
kubectl describe vm <vm-name> -n goldenpipe-system

# Check VM pod
kubectl get pods -n goldenpipe-system -l kubevirt.io=virt-launcher
kubectl logs <vm-pod> -n goldenpipe-system

# Check events
kubectl get events -n goldenpipe-system --sort-by='.lastTimestamp'
```

**Common Causes & Solutions:**

1. **Insufficient Resources**
   ```bash
   # Check node resources
   kubectl describe nodes
   
   # Solution: Increase VM resource requests or add more nodes
   ```

2. **Storage Issues**
   ```bash
   # Check PVC status
   kubectl get pvcs -n goldenpipe-system
   kubectl describe pvc <pvc-name> -n goldenpipe-system
   
   # Solution: Ensure storage class is available and has sufficient capacity
   ```

3. **KubeVirt Not Ready**
   ```bash
   # Check KubeVirt installation
   kubectl get pods -n kubevirt
   kubectl get crd | grep kubevirt
   
   # Solution: Reinstall KubeVirt if necessary
   ```

### 2. Image Download Fails

**Symptoms:**
- DataVolume stuck in `ImportScheduled` or `ImportInProgress`
- Error messages about network connectivity
- ISO download timeout

**Diagnosis:**
```bash
# Check DataVolume status
kubectl get datavolumes -n goldenpipe-system
kubectl describe datavolume <dv-name> -n goldenpipe-system

# Check CDI pods
kubectl get pods -n cdi
kubectl logs -n cdi deployment/cdi-operator
```

**Common Causes & Solutions:**

1. **Network Connectivity**
   ```bash
   # Test connectivity from CDI pod
   kubectl exec -n cdi deployment/cdi-operator -- curl -I <iso-url>
   
   # Solution: Check firewall rules and proxy settings
   ```

2. **Invalid ISO URL**
   ```bash
   # Verify URL is accessible
   curl -I <iso-url>
   
   # Solution: Use a valid, publicly accessible ISO URL
   ```

3. **CDI Not Installed**
   ```bash
   # Check CDI installation
   kubectl get crd | grep cdi
   
   # Solution: Install CDI
   kubectl apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/v1.55.0/cdi-operator.yaml
   ```

### 3. Storage Issues

**Symptoms:**
- PVC remains in `Pending` status
- Storage class not found errors
- Insufficient storage capacity

**Diagnosis:**
```bash
# Check storage classes
kubectl get storageclass
kubectl describe storageclass rook-ceph-block

# Check PVC status
kubectl get pvcs -n goldenpipe-system
kubectl describe pvc <pvc-name> -n goldenpipe-system

# Check persistent volumes
kubectl get pv
```

**Common Causes & Solutions:**

1. **Storage Class Not Available**
   ```bash
   # List available storage classes
   kubectl get storageclass
   
   # Solution: Install Rook-Ceph or configure your preferred storage class
   ```

2. **Insufficient Storage**
   ```bash
   # Check cluster storage capacity
   kubectl top nodes
   
   # Solution: Add more storage or reduce image size requirements
   ```

3. **Rook-Ceph Issues**
   ```bash
   # Check Rook-Ceph status
   kubectl get pods -n rook-ceph
   kubectl logs -n rook-ceph deployment/rook-ceph-operator
   
   # Solution: Troubleshoot Rook-Ceph installation
   ```

### 4. GoldenPipe API Issues

**Symptoms:**
- API endpoints return 500 errors
- Service not responding
- Authentication failures

**Diagnosis:**
```bash
# Check GoldenPipe pods
kubectl get pods -n goldenpipe-system
kubectl logs -f deployment/goldenpipe -n goldenpipe-system

# Check service
kubectl get services -n goldenpipe-system
kubectl describe service goldenpipe -n goldenpipe-system

# Test API connectivity
kubectl port-forward service/goldenpipe 8080:80 -n goldenpipe-system
curl http://localhost:8080/api/v1/health
```

**Common Causes & Solutions:**

1. **Configuration Issues**
   ```bash
   # Check environment variables
   kubectl describe deployment goldenpipe -n goldenpipe-system
   
   # Solution: Verify configuration in deployment.yaml
   ```

2. **Kubernetes API Access**
   ```bash
   # Check RBAC permissions
   kubectl auth can-i create virtualmachines --as=system:serviceaccount:goldenpipe-system:goldenpipe
   
   # Solution: Verify RBAC configuration in namespace.yaml
   ```

3. **Resource Limits**
   ```bash
   # Check resource usage
   kubectl top pods -n goldenpipe-system
   
   # Solution: Increase resource limits in deployment.yaml
   ```

### 5. Image Creation Hangs

**Symptoms:**
- Image creation stuck at a specific percentage
- No progress updates for extended periods
- VM appears to be running but image not ready

**Diagnosis:**
```bash
# Check VM status
kubectl get vms -n goldenpipe-system
kubectl describe vm <vm-name> -n goldenpipe-system

# Check VM console
kubectl virt console <vm-name> -n goldenpipe-system

# Check VM logs
kubectl logs -f <vm-pod> -n goldenpipe-system
```

**Common Causes & Solutions:**

1. **Cloud-Init Issues**
   ```bash
   # Check cloud-init logs in VM
   kubectl virt console <vm-name> -n goldenpipe-system
   # Then inside VM: sudo journalctl -u cloud-init
   
   # Solution: Verify cloud-init configuration in KubeVirt manager
   ```

2. **Package Installation Hanging**
   ```bash
   # Check VM console for package installation progress
   kubectl virt console <vm-name> -n goldenpipe-system
   
   # Solution: Reduce number of packages or use simpler package names
   ```

3. **Script Execution Issues**
   ```bash
   # Check script execution in VM
   kubectl virt console <vm-name> -n goldenpipe-system
   
   # Solution: Verify script syntax and dependencies
   ```

## Debugging Commands

### General Debugging

```bash
# Get all resources in namespace
kubectl get all -n goldenpipe-system

# Check events
kubectl get events -n goldenpipe-system --sort-by='.lastTimestamp'

# Check resource usage
kubectl top pods -n goldenpipe-system
kubectl top nodes

# Check logs
kubectl logs -f deployment/goldenpipe -n goldenpipe-system
```

### KubeVirt Debugging

```bash
# Check KubeVirt installation
kubectl get pods -n kubevirt
kubectl get crd | grep kubevirt

# Check VM resources
kubectl get vms -n goldenpipe-system
kubectl get vmis -n goldenpipe-system

# Access VM console
kubectl virt console <vm-name> -n goldenpipe-system

# Check VM status
kubectl virt vnc <vm-name> -n goldenpipe-system
```

### Storage Debugging

```bash
# Check storage classes
kubectl get storageclass
kubectl describe storageclass <storage-class-name>

# Check PVCs
kubectl get pvcs -n goldenpipe-system
kubectl describe pvc <pvc-name> -n goldenpipe-system

# Check PVs
kubectl get pv
kubectl describe pv <pv-name>
```

### CDI Debugging

```bash
# Check CDI installation
kubectl get pods -n cdi
kubectl get crd | grep cdi

# Check DataVolumes
kubectl get datavolumes -n goldenpipe-system
kubectl describe datavolume <dv-name> -n goldenpipe-system

# Check CDI logs
kubectl logs -n cdi deployment/cdi-operator
kubectl logs -n cdi deployment/cdi-apiserver
```

## Performance Optimization

### Resource Tuning

```bash
# Increase VM resource limits
kubectl patch deployment goldenpipe -n goldenpipe-system -p '{"spec":{"template":{"spec":{"containers":[{"name":"goldenpipe","resources":{"limits":{"memory":"1Gi","cpu":"1000m"}}}]}}}}'

# Increase concurrent VM limit
kubectl set env deployment/goldenpipe MAX_CONCURRENT_VMS=10 -n goldenpipe-system
```

### Storage Optimization

```bash
# Use faster storage class
kubectl patch storageclass rook-ceph-block -p '{"allowVolumeExpansion":true}'

# Enable volume snapshots
kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/v6.3.0/client/config/crd/snapshot.storage.k8s.io_volumesnapshots.yaml
```

## Getting Help

### Log Collection

When reporting issues, collect the following logs:

```bash
# GoldenPipe logs
kubectl logs deployment/goldenpipe -n goldenpipe-system > goldenpipe.log

# KubeVirt logs
kubectl logs -n kubevirt deployment/virt-operator > kubevirt-operator.log
kubectl logs -n kubevirt deployment/virt-api > kubevirt-api.log

# CDI logs
kubectl logs -n cdi deployment/cdi-operator > cdi-operator.log

# System events
kubectl get events -n goldenpipe-system --sort-by='.lastTimestamp' > events.log

# Resource status
kubectl get all -n goldenpipe-system > resources.log
```

### Community Support

- **GitHub Issues**: Report bugs and feature requests
- **Discussions**: Ask questions and share experiences
- **Documentation**: Check the latest documentation
- **Examples**: Look at example configurations

### Professional Support

For enterprise support and consulting:
- Contact the GoldenPipe team
- Schedule a consultation
- Request custom features
