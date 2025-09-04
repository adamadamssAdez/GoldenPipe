package kubevirt

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"k8s.io/klog/v2"

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
	_, err := m.storageManager.CreateImagePVC(ctx, pvcName, req.GetDefaultStorageSize())
	if err != nil {
		return nil, fmt.Errorf("failed to create PVC: %w", err)
	}

	// TODO: Create VM for image creation when KubeVirt client is available
	// vm, err := m.createImageCreationVM(ctx, vmName, req, pvcName)
	// if err != nil {
	//     // Clean up PVC if VM creation fails
	//     m.storageManager.DeletePVC(ctx, pvcName)
	//     return nil, fmt.Errorf("failed to create VM: %w", err)
	// }

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

// GetImageStatus returns the status of an image creation process
func (m *Manager) GetImageStatus(ctx context.Context, imageName string) (*types.ImageStatusResponse, error) {
	// Get image metadata
	image, err := m.storageManager.GetImageMetadata(imageName)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	// TODO: Get VM status when KubeVirt client is available
	// vm, err := m.client.KubeVirtClient.VirtualMachines(m.namespace).Get(ctx, image.VMName, metav1.GetOptions{})
	// if err != nil {
	//     return nil, fmt.Errorf("failed to get VM: %w", err)
	// }

	// For now, return a mock status
	status := types.ImageStatusCreating
	progress := 25
	message := "Image creation in progress (KubeVirt integration pending)"

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

	// TODO: Delete VM if it exists when KubeVirt client is available
	// if image.VMName != "" {
	//     err = m.client.KubeVirtClient.VirtualMachines(m.namespace).Delete(ctx, image.VMName, metav1.DeleteOptions{})
	//     if err != nil {
	//         klog.Errorf("Failed to delete VM %s: %v", image.VMName, err)
	//     }
	// }

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
	// TODO: List VMs when KubeVirt client is available
	// vms, err := m.client.KubeVirtClient.VirtualMachines(m.namespace).List(ctx, metav1.ListOptions{
	//     LabelSelector: "app=goldenpipe",
	// })
	// if err != nil {
	//     return nil, fmt.Errorf("failed to list VMs: %w", err)
	// }

	// Return empty list for now
	return []*types.VM{}, nil
}

// GetVM gets a specific VM
func (m *Manager) GetVM(ctx context.Context, vmName string) (*types.VM, error) {
	// TODO: Get VM when KubeVirt client is available
	return nil, fmt.Errorf("VM operations not available (KubeVirt integration pending)")
}

// DeleteVM deletes a VM
func (m *Manager) DeleteVM(ctx context.Context, vmName string) error {
	// TODO: Delete VM when KubeVirt client is available
	return fmt.Errorf("VM operations not available (KubeVirt integration pending)")
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