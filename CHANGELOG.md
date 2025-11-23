# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
### Changed
### Fixed
### Removed

## [2.3.0] - 2025-11-21

### Added
- Docker socket support in addition to containerd
- Auto-detection of container runtime (containerd or Docker)
- Support for multiple Kubernetes distributions (k0s, k3s, microk8s, standard)
- Automatic socket path detection for different distributions
- Manual socket path configuration option
- Runtime type logging for better debugging

### Changed
- Improved socket detection logic with fallback paths
- Enhanced error messages with tried socket paths
- Updated documentation with Docker support information

### Fixed
- Socket detection on non-standard Kubernetes distributions
- Image export working with both containerd and Docker runtimes

## [2.2.0] - 2025-11-21

### Added
- Containerd socket auto-detection
- Support for k0s, k3s, microk8s socket paths
- Volume mounting of /run and /var/snap directories
- Automatic fallback to multiple socket paths

### Changed
- Changed from single socket file mount to parent directory mounts
- Improved socket detection error messages

## [2.1.3] - 2025-11-21

### Added
- libc6-compat and libseccomp dependencies for ctr binary

### Fixed
- ctr binary execution issues in Alpine container

## [2.1.2] - 2025-11-21

### Added
- ctr binary from containerd release

### Fixed
- Missing ctr executable in container image

## [2.1.1] - 2025-11-21

### Fixed
- Containerd socket path for k0s distribution

## [2.1.0] - 2025-11-21

### Added
- Containerd socket support for image restoration
- PushImageFromContainerd function
- Image export from local containerd
- Detection of missing images from target registry

### Changed
- Enhanced SyncImage logic to detect target registry images
- Added restoration workflow for missing images

## [2.0.0] - 2025-11-21

### Added
- Complete rewrite in Go from bash
- Structured logging with zerolog
- Prometheus metrics export
- Health check endpoints (liveness and readiness)
- Concurrent image processing with semaphore
- Automatic retry logic with configurable attempts
- RBAC configuration with minimal permissions
- Helm chart for easy deployment
- Multi-namespace monitoring support
- Deployment-specific filtering option
- Graceful shutdown handling
- Configuration validation

### Changed
- Replaced docker/skopeo dependencies with go-containerregistry
- Changed from privileged container to non-root
- Moved from Alpine+bash to Alpine+Go binary
- Improved error handling and logging
- Enhanced security posture (non-root, read-only filesystem)

### Removed
- Docker daemon dependency
- Privileged mode requirement
- Bash scripts
- skopeo binary

### Security
- Non-root user (UID 65532)
- Read-only root filesystem
- No privileged mode
- Dropped all capabilities
- Seccomp profile enabled

## [1.0.0] - 2025-11-14

### Added
- Initial release with bash implementation
- Basic image syncing functionality
- Docker daemon dependency
- Single namespace support

---

## Version History Format

### Types of changes
- `Added` for new features
- `Changed` for changes in existing functionality
- `Deprecated` for soon-to-be removed features
- `Removed` for now removed features
- `Fixed` for any bug fixes
- `Security` for vulnerability fixes

[Unreleased]: https://github.com/tazhate/push-from-k8s-back-to-docker-registry/compare/v2.3.0...HEAD
[2.3.0]: https://github.com/tazhate/push-from-k8s-back-to-docker-registry/compare/v2.2.0...v2.3.0
[2.2.0]: https://github.com/tazhate/push-from-k8s-back-to-docker-registry/compare/v2.1.3...v2.2.0
[2.1.3]: https://github.com/tazhate/push-from-k8s-back-to-docker-registry/compare/v2.1.2...v2.1.3
[2.1.2]: https://github.com/tazhate/push-from-k8s-back-to-docker-registry/compare/v2.1.1...v2.1.2
[2.1.1]: https://github.com/tazhate/push-from-k8s-back-to-docker-registry/compare/v2.1.0...v2.1.1
[2.1.0]: https://github.com/tazhate/push-from-k8s-back-to-docker-registry/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/tazhate/push-from-k8s-back-to-docker-registry/compare/v1.0.0...v2.0.0
[1.0.0]: https://github.com/tazhate/push-from-k8s-back-to-docker-registry/releases/tag/v1.0.0
