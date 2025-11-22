package k8s

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps Kubernetes client
type Client struct {
	clientset *kubernetes.Clientset
	logger    zerolog.Logger
}

// NewClient creates a new Kubernetes client
func NewClient(logger zerolog.Logger) (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		config, err = kubeConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Client{
		clientset: clientset,
		logger:    logger,
	}, nil
}

// GetDeploymentImages returns all container images used in deployments
func (c *Client) GetDeploymentImages(ctx context.Context, namespace string, deploymentNames []string) ([]string, error) {
	var images []string
	imageSet := make(map[string]struct{})

	if len(deploymentNames) == 0 {
		// Get all deployments in the namespace
		deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list deployments in namespace %s: %w", namespace, err)
		}

		for _, dep := range deployments.Items {
			c.extractImagesFromDeployment(&dep, imageSet)
		}
	} else {
		// Get specific deployments
		for _, depName := range deploymentNames {
			dep, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, depName, metav1.GetOptions{})
			if err != nil {
				c.logger.Error().
					Err(err).
					Str("namespace", namespace).
					Str("deployment", depName).
					Msg("Failed to get deployment")
				continue
			}
			c.extractImagesFromDeployment(dep, imageSet)
		}
	}

	// Convert set to slice
	for image := range imageSet {
		images = append(images, image)
	}

	return images, nil
}

// extractImagesFromDeployment extracts all images from a deployment
func (c *Client) extractImagesFromDeployment(dep *appsv1.Deployment, imageSet map[string]struct{}) {
	// Extract from init containers
	for _, container := range dep.Spec.Template.Spec.InitContainers {
		if container.Image != "" {
			imageSet[container.Image] = struct{}{}
		}
	}

	// Extract from regular containers
	for _, container := range dep.Spec.Template.Spec.Containers {
		if container.Image != "" {
			imageSet[container.Image] = struct{}{}
		}
	}
}

// GetAllImages returns all unique images from specified namespaces and deployments
func (c *Client) GetAllImages(ctx context.Context, namespaces []string, deployments []string) ([]string, error) {
	imageSet := make(map[string]struct{})

	for _, ns := range namespaces {
		images, err := c.GetDeploymentImages(ctx, ns, deployments)
		if err != nil {
			c.logger.Error().
				Err(err).
				Str("namespace", ns).
				Msg("Failed to get images from namespace")
			continue
		}

		for _, image := range images {
			imageSet[image] = struct{}{}
		}
	}

	// Convert to slice
	var allImages []string
	for image := range imageSet {
		allImages = append(allImages, image)
	}

	c.logger.Info().
		Int("count", len(allImages)).
		Msg("Total unique images found")

	return allImages, nil
}
