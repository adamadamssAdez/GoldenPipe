package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/goldenpipe/microservice/pkg/types"
)

// Manager manages persistent storage for golden images
type Manager struct {
	client       kubernetes.Interface
	storageClass string
	namespace    string
}

// encryptData encrypts sensitive data for storage
func (m *Manager) encryptData(data []byte) ([]byte, error) {
	// TODO: Implement proper encryption for FedRAMP compliance
	// For now, return data as-is (placeholder for encryption)
	return data, nil
}

// decryptData decrypts sensitive data from storage
func (m *Manager) decryptData(encryptedData []byte) ([]byte, error) {
	// TODO: Implement proper decryption for FedRAMP compliance
	// For now, return data as-is (placeholder for decryption)
	return encryptedData, nil
}

// NewManager creates a new storage manager
func NewManager(client kubernetes.Interface, storageClass, namespace string) (*Manager, error) {
	return &Manager{
		client:       client,
		storageClass: storageClass,
		namespace:    namespace,
	}, nil
}

// CreateImagePVC creates a PVC for storing a golden image
func (m *Manager) CreateImagePVC(ctx context.Context, name, size string) (*corev1.PersistentVolumeClaim, error) {
	// Parse storage size
	storageSize, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, fmt.Errorf("invalid storage size %s: %w", size, err)
	}

	// Create PVC
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app":                   "goldenpipe",
				"goldenpipe.io/purpose": "golden-image",
				"goldenpipe.io/image":   name,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storageSize,
				},
			},
			StorageClassName: &m.storageClass,
		},
	}

	createdPVC, err := m.client.CoreV1().PersistentVolumeClaims(m.namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create PVC: %w", err)
	}

	klog.Infof("Created PVC %s with size %s", name, size)
	return createdPVC, nil
}

// DeletePVC deletes a PVC
func (m *Manager) DeletePVC(ctx context.Context, name string) error {
	err := m.client.CoreV1().PersistentVolumeClaims(m.namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete PVC %s: %w", name, err)
	}

	klog.Infof("Deleted PVC %s", name)
	return nil
}

// StoreImageMetadata stores golden image metadata in a ConfigMap
func (m *Manager) StoreImageMetadata(image *types.GoldenImage) error {
	// Convert image to JSON
	imageData, err := json.Marshal(image)
	if err != nil {
		return fmt.Errorf("failed to marshal image metadata: %w", err)
	}

	// Create ConfigMap for image metadata
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("golden-image-%s", image.Name),
			Namespace: m.namespace,
			Labels: map[string]string{
				"app":                   "goldenpipe",
				"goldenpipe.io/purpose": "image-metadata",
				"goldenpipe.io/image":   image.Name,
			},
		},
		Data: map[string]string{
			"metadata": string(imageData),
		},
	}

	// Create or update ConfigMap
	_, err = m.client.CoreV1().ConfigMaps(m.namespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		// If ConfigMap already exists, update it
		_, err = m.client.CoreV1().ConfigMaps(m.namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to store image metadata: %w", err)
		}
	}

	klog.Infof("Stored metadata for image %s", image.Name)
	return nil
}

// GetImageMetadata retrieves golden image metadata from a ConfigMap
func (m *Manager) GetImageMetadata(imageName string) (*types.GoldenImage, error) {
	configMapName := fmt.Sprintf("golden-image-%s", imageName)

	configMap, err := m.client.CoreV1().ConfigMaps(m.namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get image metadata: %w", err)
	}

	// Parse image metadata
	var image types.GoldenImage
	err = json.Unmarshal([]byte(configMap.Data["metadata"]), &image)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal image metadata: %w", err)
	}

	return &image, nil
}

// DeleteImageMetadata deletes golden image metadata
func (m *Manager) DeleteImageMetadata(imageName string) error {
	configMapName := fmt.Sprintf("golden-image-%s", imageName)

	err := m.client.CoreV1().ConfigMaps(m.namespace).Delete(context.TODO(), configMapName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete image metadata: %w", err)
	}

	klog.Infof("Deleted metadata for image %s", imageName)
	return nil
}

// ListImages lists all golden images
func (m *Manager) ListImages() ([]*types.GoldenImage, error) {
	// List all ConfigMaps with image metadata
	labelSelector := labels.Set{
		"app":                   "goldenpipe",
		"goldenpipe.io/purpose": "image-metadata",
	}.AsSelector()

	configMaps, err := m.client.CoreV1().ConfigMaps(m.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list image metadata: %w", err)
	}

	var images []*types.GoldenImage
	for _, configMap := range configMaps.Items {
		// Parse image metadata
		var image types.GoldenImage
		err = json.Unmarshal([]byte(configMap.Data["metadata"]), &image)
		if err != nil {
			klog.Errorf("Failed to unmarshal image metadata for %s: %v", configMap.Name, err)
			continue
		}
		images = append(images, &image)
	}

	return images, nil
}

// GetImage gets a specific golden image
func (m *Manager) GetImage(imageName string) (*types.GoldenImage, error) {
	return m.GetImageMetadata(imageName)
}

// CreateDataVolume creates a CDI DataVolume for importing an ISO
func (m *Manager) CreateDataVolume(ctx context.Context, name, sourceURL, size string) error {
	// This would typically use the CDI client, but for now we'll create a PVC
	// that can be used with CDI DataVolumes

	// Parse storage size
	storageSize, err := resource.ParseQuantity(size)
	if err != nil {
		return fmt.Errorf("invalid storage size %s: %w", size, err)
	}

	// Create PVC for DataVolume
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app":                   "goldenpipe",
				"goldenpipe.io/purpose": "datavolume",
				"goldenpipe.io/source":  sourceURL,
			},
			Annotations: map[string]string{
				"cdi.kubevirt.io/storage.import.source": sourceURL,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storageSize,
				},
			},
			StorageClassName: &m.storageClass,
		},
	}

	_, err = m.client.CoreV1().PersistentVolumeClaims(m.namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create DataVolume PVC: %w", err)
	}

	klog.Infof("Created DataVolume PVC %s for source %s", name, sourceURL)
	return nil
}

// WaitForDataVolumeReady waits for a DataVolume to be ready
func (m *Manager) WaitForDataVolumeReady(ctx context.Context, name string, timeout time.Duration) error {
	// This would typically check CDI DataVolume status
	// For now, we'll just wait for the PVC to be bound

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for DataVolume %s to be ready", name)
			}

			pvc, err := m.client.CoreV1().PersistentVolumeClaims(m.namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Failed to get PVC %s: %v", name, err)
				continue
			}

			if pvc.Status.Phase == corev1.ClaimBound {
				klog.Infof("DataVolume PVC %s is ready", name)
				return nil
			}

			klog.Infof("DataVolume PVC %s status: %s", name, pvc.Status.Phase)
		}
	}
}

// GetStorageUsage returns storage usage information
func (m *Manager) GetStorageUsage() (map[string]string, error) {
	// List all PVCs created by GoldenPipe
	labelSelector := labels.Set{
		"app": "goldenpipe",
	}.AsSelector()

	pvcs, err := m.client.CoreV1().PersistentVolumeClaims(m.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list PVCs: %w", err)
	}

	usage := make(map[string]string)
	totalSize := resource.Quantity{}
	totalUsed := resource.Quantity{}

	for _, pvc := range pvcs.Items {
		if pvc.Status.Capacity != nil {
			if size, exists := pvc.Status.Capacity[corev1.ResourceStorage]; exists {
				totalSize.Add(size)
			}
		}

		// For bound PVCs, we can estimate usage
		if pvc.Status.Phase == corev1.ClaimBound {
			if size, exists := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; exists {
				totalUsed.Add(size)
			}
		}
	}

	usage["total_allocated"] = totalSize.String()
	usage["total_used"] = totalUsed.String()
	usage["pvc_count"] = fmt.Sprintf("%d", len(pvcs.Items))

	return usage, nil
}
