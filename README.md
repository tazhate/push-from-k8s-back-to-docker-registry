# Push From K8s Back to Docker Registry

> *"Oops, I accidentally deleted my Docker registry. Can I get my images back?"*
> **YES.** This tool does exactly that.

[![Docker Pulls](https://img.shields.io/docker/pulls/tazhate/push-from-k8s-back-to-docker-registry)](https://hub.docker.com/r/tazhate/push-from-k8s-back-to-docker-registry)
[![GitHub Release](https://img.shields.io/github/v/release/tazhate/push-from-k8s-back-to-docker-registry)](https://github.com/tazhate/push-from-k8s-back-to-docker-registry/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## What It Does

Your K8s nodes cache every image they pull. This tool monitors your registry and **automatically restores missing images from the node's local containerd/docker cache**.

**The Problem:** Registry dies → pods restart → `ImagePullBackOff` → your weekend dies.

**The Solution:** This DaemonSet watches your namespaces, detects missing images, exports them from local cache, and pushes them back. Crisis averted.

## Quick Start

```bash
helm install image-syncer oci://ghcr.io/tazhate/charts/push-missed-images \
  --namespace kube-system \
  --set registry.url=docker.mycompany.com \
  --set registry.username=admin \
  --set registry.password=secret \
  --set "monitor.namespaces={production,staging}"
```

Or from source:
```bash
git clone https://github.com/tazhate/push-from-k8s-back-to-docker-registry.git
helm install image-syncer ./chart --namespace kube-system
```

Verify:
```bash
kubectl logs -n kube-system -l app=push-missed-images -f
# Look for: "Successfully restored image from container runtime"
```

## Configuration

```yaml
registry:
  url: "docker.mycompany.com"
  username: "admin"
  password: "changeme"

monitor:
  namespaces: ["production", "staging"]

sync:
  period: "10m"

containerd:
  socketPath: ""  # Auto-detects k0s, k3s, microk8s, standard K8s
```

## Features

- **Auto-restore** missing images from containerd/docker
- **Dual runtime** support (containerd + Docker)
- **Multi-distro** auto-detection (k0s, k3s, microk8s)
- **Prometheus metrics** on `:8080/metrics`
- **Health checks** on `:8081/healthz` and `:8081/readyz`
- **Lightweight** Alpine-based, non-root, minimal resources

## Metrics

```promql
images_synced_total        # Successfully restored images
images_sync_failed_total   # Failed restores
images_skipped_total       # Already in registry
```

## Troubleshooting

**"Failed to detect container runtime socket"**
```bash
# Find your socket
find /run -name "*containerd*.sock" -o -name "docker.sock" 2>/dev/null
# Set explicitly in values.yaml
```

**High memory usage?**
Large images = large temp files. Increase `resources.limits.memory`.

## Limitations

- Can't restore images that were never pulled to any node
- Won't help if entire cluster is gone
- **This is a safety net, not a backup strategy** - keep proper registry backups!

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). PRs welcome!

## License

MIT - Do whatever you want, just don't sue us.

---

**Made by [@tazhate](https://github.com/tazhate)** | [Issues](https://github.com/tazhate/push-from-k8s-back-to-docker-registry/issues) | [Discussions](https://github.com/tazhate/push-from-k8s-back-to-docker-registry/discussions)
