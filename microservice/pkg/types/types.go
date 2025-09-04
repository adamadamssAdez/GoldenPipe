package types

import (
	"errors"
	"strings"
	"time"
)

// OSType represents the operating system type
type OSType string

const (
	OSTypeLinux   OSType = "linux"
	OSTypeWindows OSType = "windows"
)

// ImageStatus represents the status of a golden image
type ImageStatus string

const (
	ImageStatusPending  ImageStatus = "pending"
	ImageStatusCreating ImageStatus = "creating"
	ImageStatusReady    ImageStatus = "ready"
	ImageStatusFailed   ImageStatus = "failed"
	ImageStatusDeleting ImageStatus = "deleting"
)

// VMStatus represents the status of a VM
type VMStatus string

const (
	VMStatusPending  VMStatus = "pending"
	VMStatusRunning  VMStatus = "running"
	VMStatusStopped  VMStatus = "stopped"
	VMStatusFailed   VMStatus = "failed"
	VMStatusDeleting VMStatus = "deleting"
)

// CreateImageRequest represents a request to create a golden image
type CreateImageRequest struct {
	Name           string               `json:"name" binding:"required"`
	OSType         OSType               `json:"os_type" binding:"required,oneof=linux windows"`
	BaseISOURL     string               `json:"base_iso_url" binding:"required,url"`
	Customizations *ImageCustomizations `json:"customizations,omitempty"`
	StorageSize    string               `json:"storage_size,omitempty"`
	CPU            int                  `json:"cpu,omitempty"`
	Memory         string               `json:"memory,omitempty"`
	Labels         map[string]string    `json:"labels,omitempty"`
}

// ImageCustomizations represents customizations to apply to the image
type ImageCustomizations struct {
	Packages []string          `json:"packages,omitempty"`
	Scripts  []string          `json:"scripts,omitempty"`
	Files    map[string]string `json:"files,omitempty"`
	Users    []UserConfig      `json:"users,omitempty"`
	SSHKeys  []string          `json:"ssh_keys,omitempty"`
}

// UserConfig represents user configuration for the image
type UserConfig struct {
	Name     string   `json:"name"`
	Password string   `json:"password,omitempty"`
	Groups   []string `json:"groups,omitempty"`
	Sudo     bool     `json:"sudo,omitempty"`
}

// GoldenImage represents a golden image
type GoldenImage struct {
	Name           string               `json:"name"`
	OSType         OSType               `json:"os_type"`
	Status         ImageStatus          `json:"status"`
	Size           string               `json:"size,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	Labels         map[string]string    `json:"labels,omitempty"`
	Customizations *ImageCustomizations `json:"customizations,omitempty"`
	PVCName        string               `json:"pvc_name,omitempty"`
	VMName         string               `json:"vm_name,omitempty"`
}

// VM represents a virtual machine
type VM struct {
	Name      string            `json:"name"`
	ImageName string            `json:"image_name"`
	Status    VMStatus          `json:"status"`
	IP        string            `json:"ip,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Labels    map[string]string `json:"labels,omitempty"`
	CPU       int               `json:"cpu,omitempty"`
	Memory    string            `json:"memory,omitempty"`
}

// ImageStatusResponse represents the status of an image creation process
type ImageStatusResponse struct {
	Name      string              `json:"name"`
	Status    ImageStatus         `json:"status"`
	Progress  int                 `json:"progress,omitempty"`
	Message   string              `json:"message,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	Steps     []ImageCreationStep `json:"steps,omitempty"`
	Error     string              `json:"error,omitempty"`
}

// ImageCreationStep represents a step in the image creation process
type ImageCreationStep struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	StartedAt time.Time `json:"started_at,omitempty"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
}

// Metrics represents system metrics
type Metrics struct {
	TotalImages  int       `json:"total_images"`
	ActiveVMs    int       `json:"active_vms"`
	FailedImages int       `json:"failed_images"`
	StorageUsed  string    `json:"storage_used"`
	LastUpdated  time.Time `json:"last_updated"`
}

// Validate validates the CreateImageRequest
func (r *CreateImageRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}

	if r.OSType != OSTypeLinux && r.OSType != OSTypeWindows {
		return errors.New("os_type must be 'linux' or 'windows'")
	}

	if r.BaseISOURL == "" {
		return errors.New("base_iso_url is required")
	}

	if !strings.HasPrefix(r.BaseISOURL, "http://") && !strings.HasPrefix(r.BaseISOURL, "https://") {
		return errors.New("base_iso_url must be a valid HTTP/HTTPS URL")
	}

	// Validate name format (Kubernetes resource name)
	if !isValidKubernetesName(r.Name) {
		return errors.New("name must be a valid Kubernetes resource name (lowercase alphanumeric and hyphens only)")
	}

	return nil
}

// isValidKubernetesName checks if a string is a valid Kubernetes resource name
func isValidKubernetesName(name string) bool {
	if len(name) == 0 || len(name) > 253 {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}

	return true
}

// GetDefaultStorageSize returns the default storage size for an OS type
func (r *CreateImageRequest) GetDefaultStorageSize() string {
	if r.StorageSize != "" {
		return r.StorageSize
	}

	switch r.OSType {
	case OSTypeLinux:
		return "20Gi"
	case OSTypeWindows:
		return "50Gi"
	default:
		return "20Gi"
	}
}

// GetDefaultCPU returns the default CPU count for an OS type
func (r *CreateImageRequest) GetDefaultCPU() int {
	if r.CPU > 0 {
		return r.CPU
	}

	switch r.OSType {
	case OSTypeLinux:
		return 2
	case OSTypeWindows:
		return 4
	default:
		return 2
	}
}

// GetDefaultMemory returns the default memory for an OS type
func (r *CreateImageRequest) GetDefaultMemory() string {
	if r.Memory != "" {
		return r.Memory
	}

	switch r.OSType {
	case OSTypeLinux:
		return "4Gi"
	case OSTypeWindows:
		return "8Gi"
	default:
		return "4Gi"
	}
}

// String returns the string representation of OSType
func (o OSType) String() string {
	return string(o)
}

// String returns the string representation of ImageStatus
func (s ImageStatus) String() string {
	return string(s)
}

// String returns the string representation of VMStatus
func (s VMStatus) String() string {
	return string(s)
}
