package registry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/rs/zerolog"
	"github.com/tazhate/push-from-k8s-back-to-docker-registry/internal/metrics"
)

// Client handles container registry operations
type Client struct {
	targetRegistry       string
	auth                 authn.Authenticator
	logger               zerolog.Logger
	options              []crane.Option
	containerdSocketPath string
	runtimeType          RuntimeType
}

// ImageRef represents a parsed container image reference
type ImageRef struct {
	Registry   string
	Repository string
	Tag        string
	FullRef    string
}

// NewClient creates a new registry client
func NewClient(registryURL, username, password, containerdSocketPath string, runtimeType RuntimeType, logger zerolog.Logger) (*Client, error) {
	auth := &authn.Basic{
		Username: username,
		Password: password,
	}

	options := []crane.Option{
		crane.WithAuth(auth),
		crane.WithContext(context.Background()),
	}

	return &Client{
		targetRegistry:       strings.TrimSuffix(registryURL, "/"),
		auth:                 auth,
		logger:               logger,
		options:              options,
		containerdSocketPath: containerdSocketPath,
		runtimeType:          runtimeType,
	}, nil
}

// ParseImageRef parses an image reference into components
func ParseImageRef(image string) (*ImageRef, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference: %w", err)
	}

	registry := ref.Context().RegistryStr()
	repository := ref.Context().RepositoryStr()
	tag := "latest"

	if t, ok := ref.(name.Tag); ok {
		tag = t.TagStr()
	}

	return &ImageRef{
		Registry:   registry,
		Repository: repository,
		Tag:        tag,
		FullRef:    ref.Name(),
	}, nil
}

// BuildTargetRef constructs the target image reference
func (c *Client) BuildTargetRef(sourceImage string) (string, error) {
	ref, err := ParseImageRef(sourceImage)
	if err != nil {
		return "", err
	}

	// Extract repository path (remove source registry)
	var repoPath string

	// If registry is docker.io or index.docker.io, keep the full path
	if ref.Registry == "index.docker.io" || ref.Registry == "docker.io" {
		repoPath = ref.Repository
	} else {
		// For other registries, keep only the path after registry
		repoPath = strings.TrimPrefix(ref.Repository, ref.Registry+"/")
	}

	targetImage := fmt.Sprintf("%s/%s:%s", c.targetRegistry, repoPath, ref.Tag)
	return targetImage, nil
}

// ImageExists checks if an image already exists in the target registry
func (c *Client) ImageExists(ctx context.Context, imageRef string) (bool, error) {
	start := time.Now()
	defer func() {
		metrics.ImageSyncDuration.WithLabelValues("check_exists").Observe(time.Since(start).Seconds())
	}()

	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return false, fmt.Errorf("failed to parse reference: %w", err)
	}

	_, err = remote.Head(ref, remote.WithAuth(c.auth), remote.WithContext(ctx))
	if err != nil {
		if strings.Contains(err.Error(), "MANIFEST_UNKNOWN") ||
			strings.Contains(err.Error(), "NAME_UNKNOWN") ||
			strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if image exists: %w", err)
	}

	return true, nil
}

// CopyImage copies an image from source to target registry
func (c *Client) CopyImage(ctx context.Context, sourceImage, targetImage string) error {
	start := time.Now()
	defer func() {
		metrics.ImageSyncDuration.WithLabelValues("copy").Observe(time.Since(start).Seconds())
	}()

	c.logger.Info().
		Str("source", sourceImage).
		Str("target", targetImage).
		Msg("Copying image")

	// Use crane to copy the image
	err := crane.Copy(sourceImage, targetImage, c.options...)
	if err != nil {
		return fmt.Errorf("failed to copy image: %w", err)
	}

	return nil
}

// SyncImage syncs a single image to the target registry
func (c *Client) SyncImage(ctx context.Context, sourceImage string) error {
	sourceRef, err := ParseImageRef(sourceImage)
	if err != nil {
		c.logger.Error().
			Err(err).
			Str("image", sourceImage).
			Msg("Failed to parse source image")
		return err
	}

	targetImage, err := c.BuildTargetRef(sourceImage)
	if err != nil {
		c.logger.Error().
			Err(err).
			Str("image", sourceImage).
			Msg("Failed to build target reference")
		return err
	}

	c.logger.Debug().
		Str("source", sourceImage).
		Str("target", targetImage).
		Msg("Processing image")

	// Check if image already exists in target registry
	exists, err := c.ImageExists(ctx, targetImage)
	if err != nil {
		c.logger.Warn().
			Err(err).
			Str("image", targetImage).
			Msg("Failed to check if image exists, will attempt to restore")
	} else if exists {
		c.logger.Debug().
			Str("image", targetImage).
			Msg("Image already exists in target registry, skipping")
		metrics.ImagesSkipped.WithLabelValues(sourceRef.Registry, c.targetRegistry).Inc()
		return nil
	}

	// Check if source registry is the same as target registry
	// In this case, we need to restore the image from local containerd
	isTargetRegistry := strings.Contains(sourceImage, c.targetRegistry)

	if isTargetRegistry {
		c.logger.Info().
			Str("image", sourceImage).
			Str("runtime", string(c.runtimeType)).
			Msg("Image is from target registry but missing - restoring from container runtime")

		// Try to restore from container runtime
		err = c.PushImageFromContainerd(ctx, sourceImage, targetImage, c.containerdSocketPath, c.runtimeType)
		if err != nil {
			c.logger.Error().
				Err(err).
				Str("image", sourceImage).
				Str("target", targetImage).
				Str("runtime", string(c.runtimeType)).
				Msg("Failed to restore image from container runtime")
			metrics.ImagesSyncFailed.WithLabelValues(sourceRef.Registry, c.targetRegistry, "restore_failed").Inc()
			return err
		}

		c.logger.Info().
			Str("image", sourceImage).
			Str("target", targetImage).
			Str("runtime", string(c.runtimeType)).
			Msg("Successfully restored image from container runtime")
		metrics.ImagesSynced.WithLabelValues(sourceRef.Registry, c.targetRegistry).Inc()
		return nil
	}

	// Copy image from external registry
	err = c.CopyImage(ctx, sourceImage, targetImage)
	if err != nil {
		c.logger.Error().
			Err(err).
			Str("source", sourceImage).
			Str("target", targetImage).
			Msg("Failed to copy image")
		metrics.ImagesSyncFailed.WithLabelValues(sourceRef.Registry, c.targetRegistry, "copy_failed").Inc()
		return err
	}

	c.logger.Info().
		Str("source", sourceImage).
		Str("target", targetImage).
		Msg("Successfully synced image")
	metrics.ImagesSynced.WithLabelValues(sourceRef.Registry, c.targetRegistry).Inc()

	return nil
}
