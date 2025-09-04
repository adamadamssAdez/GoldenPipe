package kubevirt

import (
	"context"
	"fmt"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"kubevirt.io/kubevirt v1.1.0"
)

// Client wraps Kubernetes and KubeVirt clients
type Client struct {
	KubeClient     kubernetes.Interface
	KubeVirtClient kubevirt.KubevirtV1Interface
	Config         *rest.Config
}

// NewClient creates a new Kubernetes and KubeVirt client
func NewClient(kubeconfigPath string) (*Client, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		// Try in-cluster config first
		config, err = rest.InClusterConfig()
		if err != nil {
			// Fall back to default kubeconfig location
			if home := homedir.HomeDir(); home != "" {
				kubeconfigPath = filepath.Join(home, ".kube", "config")
				config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes config: %w", err)
	}

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Create KubeVirt client
	kubeVirtClient, err := kubevirt.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create KubeVirt client: %w", err)
	}

	return &Client{
		KubeClient:     kubeClient,
		KubeVirtClient: kubeVirtClient,
		Config:         config,
	}, nil
}

// HealthCheck verifies that the client can connect to the cluster
func (c *Client) HealthCheck(ctx context.Context) (bool, error) {
	_, err := c.KubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return false, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Check if KubeVirt is installed
	_, err = c.KubeVirtClient.VirtualMachines("kubevirt").List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return false, fmt.Errorf("KubeVirt not available: %w", err)
	}

	return true, nil
}
