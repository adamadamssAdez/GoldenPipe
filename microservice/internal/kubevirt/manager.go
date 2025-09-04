package kubevirt

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"kubevirt.io/kubevirt v1.1.0"

	"github.com/goldenpipe/microservice/internal/storage"
	"github.com/goldenpipe/microservice/pkg/types"
)

// Manager manages KubeVirt VMs and golden images
type Manager struct {
	client         *Client
	storageManager *storage.Manager
	namespace      string
}

// NewManager creates a new KubeVirt manager
func NewManager(client *Client, storageManager *storage.Manager, namespace string) (*Manager, error) {
	return &Manager{
		client:         client,
		storageManager: storageManager,
		namespace:      namespace,
	}, nil
}

// CreateGoldenImage creates a new golden image VM
func (m *Manager) CreateGoldenImage(ctx context.Context, req *types.CreateImageRequest) (*types.GoldenImage, error) {
	// Generate unique names
	vmName := fmt.Sprintf("golden-image-%s-%s", req.Name, uuid.New().String()[:8])
	pvcName := fmt.Sprintf("golden-image-%s", req.Name)

	// Create PVC for the golden image
	pvc, err := m.storageManager.CreateImagePVC(ctx, pvcName, req.GetDefaultStorageSize())
	if err != nil {
		return nil, fmt.Errorf("failed to create PVC: %w", err)
	}

	// Create VM for image creation
	vm, err := m.createImageCreationVM(ctx, vmName, req, pvcName)
	if err != nil {
		// Clean up PVC if VM creation fails
		m.storageManager.DeletePVC(ctx, pvcName)
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	// Create golden image record
	image := &types.GoldenImage{
		Name:           req.Name,
		OSType:         req.OSType,
		Status:         types.ImageStatusCreating,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Labels:         req.Labels,
		Customizations: req.Customizations,
		PVCName:        pvcName,
		VMName:         vmName,
	}

	// Store image metadata
	err = m.storageManager.StoreImageMetadata(image)
	if err != nil {
		klog.Errorf("Failed to store image metadata: %v", err)
	}

	return image, nil
}

// createImageCreationVM creates a VM for building the golden image
func (m *Manager) createImageCreationVM(ctx context.Context, vmName string, req *types.CreateImageRequest, pvcName string) (*kubevirt.VirtualMachine, error) {
	// Create cloud-init or autounattend configuration
	var userData, networkData string
	var err error

	if req.OSType == types.OSTypeLinux {
		userData, networkData, err = m.createCloudInitConfig(req)
	} else {
		userData, err = m.createAutounattendConfig(req)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create configuration: %w", err)
	}

	// Create VM specification
	vm := &kubevirt.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vmName,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app":                   "goldenpipe",
				"goldenpipe.io/image":   req.Name,
				"goldenpipe.io/os-type": string(req.OSType),
				"goldenpipe.io/purpose": "image-creation",
			},
		},
		Spec: kubevirt.VirtualMachineSpec{
			Running: &[]bool{true}[0],
			Template: &kubevirt.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                   "goldenpipe",
						"goldenpipe.io/image":   req.Name,
						"goldenpipe.io/os-type": string(req.OSType),
						"goldenpipe.io/purpose": "image-creation",
					},
				},
				Spec: kubevirt.VirtualMachineInstanceSpec{
					Domain: kubevirt.DomainSpec{
						CPU: &kubevirt.CPU{
							Cores: uint32(req.GetDefaultCPU()),
						},
						Resources: kubevirt.ResourceRequirements{
							Requests: kubevirt.ResourceList{
								"memory": resource.MustParse(req.GetDefaultMemory()),
							},
						},
						Devices: kubevirt.Devices{
							Disks: []kubevirt.Disk{
								{
									Name: "bootdisk",
									DiskDevice: kubevirt.DiskDevice{
										CDRom: &kubevirt.CDRomTarget{
											Bus: "sata",
										},
									},
								},
								{
									Name: "datavolume",
									DiskDevice: kubevirt.DiskDevice{
										Disk: &kubevirt.DiskTarget{
											Bus: "virtio",
										},
									},
								},
							},
							Interfaces: []kubevirt.Interface{
								{
									Name: "default",
									InterfaceBindingMethod: kubevirt.InterfaceBindingMethod{
										Masquerade: &kubevirt.InterfaceMasquerade{},
									},
								},
							},
						},
					},
					Volumes: []kubevirt.Volume{
						{
							Name: "bootdisk",
							VolumeSource: kubevirt.VolumeSource{
								ContainerDisk: &kubevirt.ContainerDiskSource{
									Image: req.BaseISOURL,
								},
							},
						},
						{
							Name: "datavolume",
							VolumeSource: kubevirt.VolumeSource{
								PersistentVolumeClaim: &kubevirt.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
						{
							Name: "cloudinitdisk",
							VolumeSource: kubevirt.VolumeSource{
								CloudInitNoCloud: &kubevirt.CloudInitNoCloudSource{
									UserData:    userData,
									NetworkData: networkData,
								},
							},
						},
					},
					Networks: []kubevirt.Network{
						{
							Name: "default",
							NetworkSource: kubevirt.NetworkSource{
								Pod: &kubevirt.PodNetwork{},
							},
						},
					},
				},
			},
		},
	}

	// Create the VM
	createdVM, err := m.client.KubeVirtClient.VirtualMachines(m.namespace).Create(ctx, vm, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	return createdVM, nil
}

// GetImageStatus returns the status of an image creation process
func (m *Manager) GetImageStatus(ctx context.Context, imageName string) (*types.ImageStatusResponse, error) {
	// Get image metadata
	image, err := m.storageManager.GetImageMetadata(imageName)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	// Get VM status
	vm, err := m.client.KubeVirtClient.VirtualMachines(m.namespace).Get(ctx, image.VMName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}

	// Determine status based on VM state
	status := types.ImageStatusPending
	progress := 0
	message := ""

	if vm.Status.Ready {
		status = types.ImageStatusReady
		progress = 100
		message = "Image creation completed successfully"
	} else if vm.Status.Created {
		status = types.ImageStatusCreating
		progress = 50
		message = "VM is running, image creation in progress"
	}

	// Check for failures
	if vm.Status.Conditions != nil {
		for _, condition := range vm.Status.Conditions {
			if condition.Type == kubevirt.VirtualMachineFailure && condition.Status == "True" {
				status = types.ImageStatusFailed
				message = condition.Message
				break
			}
		}
	}

	return &types.ImageStatusResponse{
		Name:      imageName,
		Status:    status,
		Progress:  progress,
		Message:   message,
		CreatedAt: image.CreatedAt,
		UpdatedAt: time.Now(),
	}, nil
}

// DeleteImage deletes a golden image and its associated resources
func (m *Manager) DeleteImage(ctx context.Context, imageName string) error {
	// Get image metadata
	image, err := m.storageManager.GetImageMetadata(imageName)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	// Delete VM if it exists
	if image.VMName != "" {
		err = m.client.KubeVirtClient.VirtualMachines(m.namespace).Delete(ctx, image.VMName, metav1.DeleteOptions{})
		if err != nil {
			klog.Errorf("Failed to delete VM %s: %v", image.VMName, err)
		}
	}

	// Delete PVC
	if image.PVCName != "" {
		err = m.storageManager.DeletePVC(ctx, image.PVCName)
		if err != nil {
			klog.Errorf("Failed to delete PVC %s: %v", image.PVCName, err)
		}
	}

	// Delete image metadata
	err = m.storageManager.DeleteImageMetadata(imageName)
	if err != nil {
		klog.Errorf("Failed to delete image metadata: %v", err)
	}

	return nil
}

// ListVMs lists all VMs in the namespace
func (m *Manager) ListVMs(ctx context.Context) ([]*types.VM, error) {
	vms, err := m.client.KubeVirtClient.VirtualMachines(m.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=goldenpipe",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	var result []*types.VM
	for _, vm := range vms.Items {
		vmType := &types.VM{
			Name:      vm.Name,
			Status:    m.mapVMStatus(vm.Status),
			CreatedAt: vm.CreationTimestamp.Time,
			UpdatedAt: vm.CreationTimestamp.Time,
			Labels:    vm.Labels,
		}

		// Get image name from labels
		if imageName, exists := vm.Labels["goldenpipe.io/image"]; exists {
			vmType.ImageName = imageName
		}

		result = append(result, vmType)
	}

	return result, nil
}

// GetVM gets a specific VM
func (m *Manager) GetVM(ctx context.Context, vmName string) (*types.VM, error) {
	vm, err := m.client.KubeVirtClient.VirtualMachines(m.namespace).Get(ctx, vmName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}

	vmType := &types.VM{
		Name:      vm.Name,
		Status:    m.mapVMStatus(vm.Status),
		CreatedAt: vm.CreationTimestamp.Time,
		UpdatedAt: vm.CreationTimestamp.Time,
		Labels:    vm.Labels,
	}

	// Get image name from labels
	if imageName, exists := vm.Labels["goldenpipe.io/image"]; exists {
		vmType.ImageName = imageName
	}

	return vmType, nil
}

// DeleteVM deletes a VM
func (m *Manager) DeleteVM(ctx context.Context, vmName string) error {
	err := m.client.KubeVirtClient.VirtualMachines(m.namespace).Delete(ctx, vmName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}
	return nil
}

// HealthCheck performs a health check
func (m *Manager) HealthCheck(ctx context.Context) (bool, error) {
	return m.client.HealthCheck(ctx)
}

// GetMetrics returns system metrics
func (m *Manager) GetMetrics(ctx context.Context) (*types.Metrics, error) {
	// List all images
	images, err := m.storageManager.ListImages()
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	// List all VMs
	vms, err := m.ListVMs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	// Calculate metrics
	totalImages := len(images)
	activeVMs := 0
	failedImages := 0

	for _, vm := range vms {
		if vm.Status == types.VMStatusRunning {
			activeVMs++
		}
	}

	for _, image := range images {
		if image.Status == types.ImageStatusFailed {
			failedImages++
		}
	}

	return &types.Metrics{
		TotalImages:  totalImages,
		ActiveVMs:    activeVMs,
		FailedImages: failedImages,
		StorageUsed:  "N/A", // TODO: Calculate actual storage usage
		LastUpdated:  time.Now(),
	}, nil
}

// mapVMStatus maps KubeVirt VM status to our VM status
func (m *Manager) mapVMStatus(status kubevirt.VirtualMachineStatus) types.VMStatus {
	if status.Ready {
		return types.VMStatusRunning
	}
	if status.Created {
		return types.VMStatusPending
	}
	return types.VMStatusStopped
}
