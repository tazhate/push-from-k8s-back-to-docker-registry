package registry

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/rs/zerolog"
	"github.com/tazhate/push-from-k8s-back-to-docker-registry/internal/metrics"
)

const (
	containerdNamespace = "k8s.io"
)

// RuntimeType represents the container runtime type
type RuntimeType string

const (
	RuntimeContainerd RuntimeType = "containerd"
	RuntimeDocker     RuntimeType = "docker"
)

// Common socket paths for containerd and docker
var commonSocketPaths = []struct {
	path    string
	runtime RuntimeType
}{
	{"/host/run/containerd/containerd.sock", RuntimeContainerd},                   // standard containerd
	{"/host/run/k0s/containerd.sock", RuntimeContainerd},                          // k0s
	{"/host/run/k3s/containerd/containerd.sock", RuntimeContainerd},               // k3s
	{"/host/var/snap/microk8s/common/run/containerd.sock", RuntimeContainerd},     // microk8s
	{"/host/run/docker.sock", RuntimeDocker},                                      // standard docker
	{"/run/containerd/containerd.sock", RuntimeContainerd},                        // fallback without /host prefix
	{"/run/k0s/containerd.sock", RuntimeContainerd},                               // k0s fallback
	{"/run/k3s/containerd/containerd.sock", RuntimeContainerd},                    // k3s fallback
	{"/var/snap/microk8s/common/run/containerd.sock", RuntimeContainerd},          // microk8s fallback
	{"/run/docker.sock", RuntimeDocker},                                           // docker fallback
}

// DetectContainerdSocket attempts to find the container runtime socket
// If socketPath is provided and valid, returns it along with detected runtime type
// Otherwise, tries common paths and returns the first valid one
func DetectContainerdSocket(socketPath string, logger zerolog.Logger) (string, RuntimeType, error) {
	// If path is explicitly provided, validate and use it
	if socketPath != "" {
		if _, err := os.Stat(socketPath); err == nil {
			runtime := RuntimeContainerd
			if strings.Contains(socketPath, "docker.sock") {
				runtime = RuntimeDocker
			}
			logger.Info().
				Str("socket", socketPath).
				Str("runtime", string(runtime)).
				Msg("Using explicitly configured container runtime socket")
			return socketPath, runtime, nil
		}
		return "", "", fmt.Errorf("configured socket path does not exist: %s", socketPath)
	}

	// Auto-detect from common paths
	logger.Info().Msg("Auto-detecting container runtime socket path...")

	for _, socket := range commonSocketPaths {
		if _, err := os.Stat(socket.path); err == nil {
			logger.Info().
				Str("socket", socket.path).
				Str("runtime", string(socket.runtime)).
				Msg("Auto-detected container runtime socket")
			return socket.path, socket.runtime, nil
		}
	}

	var paths []string
	for _, s := range commonSocketPaths {
		paths = append(paths, s.path)
	}

	return "", "", fmt.Errorf("failed to auto-detect container runtime socket. Tried paths: %v. "+
		"Please set CONTAINERD_SOCKET_PATH environment variable or use --set containerd.socketPath in Helm",
		paths)
}

// PushImageFromContainerd exports an image from container runtime and pushes it to registry
func (c *Client) PushImageFromContainerd(ctx context.Context, imageName, targetImage, socketPath string, runtime RuntimeType) error {
	start := time.Now()
	defer func() {
		metrics.ImageSyncDuration.WithLabelValues("push_from_runtime").Observe(time.Since(start).Seconds())
	}()

	c.logger.Info().
		Str("image", imageName).
		Str("target", targetImage).
		Str("runtime", string(runtime)).
		Msg("Pushing image from container runtime to registry")

	// Export image as tar
	tmpfile := fmt.Sprintf("/tmp/image-%d.tar", time.Now().Unix())
	defer os.Remove(tmpfile)

	var cmd *exec.Cmd
	switch runtime {
	case RuntimeContainerd:
		// Use ctr to export the image
		// ctr -n k8s.io images export output.tar imageName
		cmd = exec.CommandContext(ctx, "ctr", "-n", containerdNamespace, "images", "export", tmpfile, imageName)
		cmd.Env = append(os.Environ(), fmt.Sprintf("CONTAINERD_ADDRESS=%s", socketPath))
	case RuntimeDocker:
		// Use docker to save the image
		// docker save -o output.tar imageName
		cmd = exec.CommandContext(ctx, "docker", "save", "-o", tmpfile, imageName)
		cmd.Env = append(os.Environ(), fmt.Sprintf("DOCKER_HOST=unix://%s", socketPath))
	default:
		return fmt.Errorf("unsupported runtime type: %s", runtime)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to export image from %s: %w, output: %s", runtime, err, string(output))
	}

	c.logger.Debug().
		Str("tarfile", tmpfile).
		Str("image", imageName).
		Str("runtime", string(runtime)).
		Msg("Exported image from container runtime to tar")

	// Load the tar as an image
	v1Image, err := tarball.ImageFromPath(tmpfile, nil)
	if err != nil {
		return fmt.Errorf("failed to load image from tar: %w", err)
	}

	// Push the image using crane
	err = crane.Push(v1Image, targetImage, c.options...)
	if err != nil {
		return fmt.Errorf("failed to push image to registry: %w", err)
	}

	c.logger.Info().
		Str("image", imageName).
		Str("target", targetImage).
		Str("runtime", string(runtime)).
		Msg("Successfully pushed image from container runtime to registry")

	return nil
}

// ImageExistsInContainerd checks if an image exists in local containerd
func ImageExistsInContainerd(ctx context.Context, imageName, socketPath string, logger zerolog.Logger) (bool, error) {
	// Use ctr to check if image exists
	// ctr -n k8s.io images list | grep imageName
	cmd := exec.CommandContext(ctx, "ctr", "-n", containerdNamespace, "images", "list", "--quiet")
	cmd.Env = append(os.Environ(), fmt.Sprintf("CONTAINERD_ADDRESS=%s", socketPath))

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to list images in containerd: %w", err)
	}

	// Check if imageName is in the output
	return string(output) != "", nil
}
