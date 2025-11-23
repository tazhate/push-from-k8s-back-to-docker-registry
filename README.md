# 🔄 Push From K8s Back to Docker Registry

> *"Oops, I accidentally `rm -rf`'d my entire Docker registry. Is there a way to get my images back?"*
>
> *— Every DevOps engineer at 3 AM, probably*

**The answer is YES!** And this tool does exactly that. Automatically. While you sleep. Like a good Kubernetes citizen should.

[![Docker Pulls](https://img.shields.io/docker/pulls/tazhate/push-from-k8s-back-to-docker-registry)](https://hub.docker.com/r/tazhate/push-from-k8s-back-to-docker-registry)
[![GitHub Release](https://img.shields.io/github/v/release/tazhate/push-from-k8s-back-to-docker-registry)](https://github.com/tazhate/push-from-k8s-back-to-docker-registry/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/tazhate/push-from-k8s-back-to-docker-registry)](https://goreportcard.com/report/github.com/tazhate/push-from-k8s-back-to-docker-registry)
[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org)

## 🤔 What Is This Sorcery?

**TL;DR:** Your Kubernetes cluster is basically a distributed backup of your Docker registry. This tool just reminds it to share.

### The Problem

You run your own Docker registry (because you're cool like that, or because your company's security team has trust issues with Docker Hub). Your Kubernetes cluster pulls images from it. Life is good. Birds are singing. Your boss is happy.

Then disaster strikes:
- 💀 Someone accidentally deletes images from your registry (*"I was just testing the API!"*)
- 🔥 Your registry's storage backend goes poof (*S3 decided to take a vacation*)
- 🤦 You run garbage collection too aggressively (*"It said 'safe to delete', I swear!"*)
- 🎲 Any other creative way to lose your precious container images (*we've all been there*)

Your pods keep running (yay!), but try restarting them... **BOOM!** `ImagePullBackOff` hell. Your pager explodes. Your weekend plans evaporate. Your boss is no longer happy.

### The Solution

This tool runs as a DaemonSet on your Kubernetes nodes and:

1. 👀 **Watches** your specified namespaces for running pods
2. 🔍 **Checks** if their images exist in your registry
3. 🚨 **Detects** when an image is missing (*"Houston, we have a problem"*)
4. 💾 **Exports** the image from the node's local container runtime (containerd or Docker)
5. 📤 **Pushes** it back to your registry
6. 😎 **Saves the day** (and your weekend)
7. 🎉 **Doesn't tell your boss** (you're welcome)

It's like `git push --force` but for container images. And less dangerous. Probably.

## ✨ Features

- 🎯 **Automatic Image Restoration**: Detects and restores missing images without manual intervention (because manual intervention at 3 AM is overrated)
- 🔄 **Dual Runtime Support**: Works with both containerd and Docker container runtimes (we don't discriminate)
- 🎪 **Multi-Distribution Support**: Auto-detects socket paths for k0s, k3s, microk8s, and standard Kubernetes (because every snowflake is unique)
- 📊 **Prometheus Metrics**: Because if it's not monitored, does it even exist? (Spoiler: no)
- 🏥 **Health Checks**: Kubernetes-native liveness and readiness probes (for when K8s needs to know we're okay)
- ⚡ **Concurrent Processing**: Handles multiple images simultaneously (but not too many, we're not animals)
- 🔁 **Automatic Retries**: Because networks are unreliable and we accept that (unlike our users)
- 🎨 **Pretty Logs**: Structured JSON logging with colors (yes, colors matter, fight me)
- 🪶 **Lightweight**: Uses Alpine Linux because we're not made of disk space
- 🔒 **Secure**: Non-root, read-only filesystem, no privileged mode (your security team will love us)
- 💪 **Production Ready**: Written in Go, not held together with shell scripts and prayers

## 🚀 Quick Start

### Prerequisites

- A Kubernetes cluster (obviously)
- A Docker registry that you actually want to back up (Harbor, Nexus, or your uncle's Raspberry Pi - we don't judge)
- The ability to read YAML without crying
- Coffee (optional, but strongly recommended)
- A disaster to recover from (or a desire to prevent one)

### Installation with Helm

```bash
# Install the chart (when we publish it)
helm install image-syncer oci://ghcr.io/tazhate/charts/push-missed-images \
  --namespace kube-system \
  --create-namespace \
  --set registry.url=docker.mycompany.com \
  --set registry.username=admin \
  --set registry.password=super-secret-password \
  --set monitor.namespaces="{production,staging}"
```

Or if you like living dangerously and reading from source:

```bash
git clone https://github.com/tazhate/push-from-k8s-back-to-docker-registry.git
cd push-from-k8s-back-to-docker-registry

# Edit chart/values.yaml with your settings
vim chart/values.yaml  # or nano, we won't tell anyone

# Deploy
helm install image-syncer ./chart --namespace kube-system
```

### Verify It's Working

```bash
# Check DaemonSet status (should see one pod per node)
kubectl get daemonset -n kube-system push-missed-images

# Watch the magic happen
kubectl logs -n kube-system -l app=push-missed-images -f

# If you see "Successfully restored image from container runtime" - you're golden! 🎉
```

## ⚙️ Configuration

Key configuration options in `values.yaml`:

```yaml
# Your registry settings
registry:
  url: "docker.mycompany.com"
  username: "admin"
  password: "changeme"  # Seriously, change this

# Which namespaces to monitor
monitor:
  namespaces:
    - "production"     # The important stuff
    - "staging"        # Less important stuff
    - "dev"            # Probably broken anyway

# How often to check (because polling is still cool in 2025, right?)
sync:
  period: "10m"  # Every 10 minutes. Adjust based on your paranoia level.

# Container runtime socket (leave empty for auto-detection)
containerd:
  socketPath: ""  # Auto-detects containerd, docker, k0s, k3s, microk8s
  # Or manually specify if you're into that sort of thing:
  # socketPath: "/run/k3s/containerd/containerd.sock"
```

<details>
<summary>📖 Full Configuration Reference (click to expand)</summary>

| Parameter | Description | Default | You Should |
|-----------|-------------|---------|------------|
| `registry.url` | Target registry URL | `ghcr.io` | Change this |
| `registry.username` | Registry username | `""` | Set this |
| `registry.password` | Registry password | `""` | Set this (and rotate it) |
| `registry.existingSecret` | Use existing secret | `""` | Consider using this |
| `monitor.namespaces` | Namespaces to monitor | `["default"]` | Customize |
| `monitor.deployments` | Specific deployments (empty = all) | `[]` | Usually leave empty |
| `sync.period` | Sync interval | `10m` | Tune based on needs |
| `containerd.socketPath` | Container runtime socket | `""` | Let it auto-detect |
| `logging.level` | Log level | `info` | `debug` if things are weird |
| `resources.requests.cpu` | CPU request | `50m` | Increase if slow |
| `resources.requests.memory` | Memory request | `64Mi` | Increase for large images |
| `resources.limits.cpu` | CPU limit | `200m` | Don't be greedy |
| `resources.limits.memory` | Memory limit | `128Mi` | Unless you have 2GB images |

</details>

For more detailed configuration options, see [values.yaml](chart/values.yaml).

## 📊 Monitoring

### Prometheus Metrics

The application exposes Prometheus metrics on port `:8080/metrics`:

- `image_sync_duration_seconds` - How long operations take (histogram)
- `images_synced_total` - Successfully synced images (your hero metric)
- `images_sync_failed_total` - Failed syncs (your "oh no" metric)
- `images_skipped_total` - Images already in registry (efficiency wins!)
- `images_processed` - Total images being monitored (gauge)

Example Prometheus queries that will make you look smart in meetings:

```promql
# Sync success rate (higher is better, obviously)
rate(images_synced_total[5m]) / (rate(images_synced_total[5m]) + rate(images_sync_failed_total[5m]))

# Average sync time per image
rate(image_sync_duration_seconds_sum{operation="push_from_runtime"}[5m]) /
rate(image_sync_duration_seconds_count{operation="push_from_runtime"}[5m])

# Images restored in last hour (your "I saved the day" metric)
increase(images_synced_total{source_registry="docker.mycompany.com"}[1h])
```

### Grafana Dashboard

Import our pre-built dashboard (coming soon™) or create your own because you're probably better at Grafana than we are.

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                    │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │       DaemonSet (one pod per node)               │  │
│  │                                                  │  │
│  │  Every 10 minutes (or whatever you configure):  │  │
│  │                                                  │  │
│  │  1. "Hey K8s API, what images are running?"     │  │
│  │  2. "Hey registry, do you have nginx:1.21?"     │  │
│  │  3. Registry: "404 Not Found"                    │  │
│  │  4. "OH NO! Time to save the day!"              │  │
│  │  5. Export from containerd/docker locally       │  │
│  │  6. Push back to registry                        │  │
│  │  7. "Crisis averted. You're welcome."           │  │
│  │                                                  │  │
│  │  Exposes:                                        │  │
│  │    - :8080/metrics (Prometheus)                  │  │
│  │    - :8081/healthz (Liveness)                    │  │
│  │    - :8081/readyz (Readiness)                    │  │
│  └──────────────────────────────────────────────────┘  │
│                          │                               │
│                          │ Access socket                 │
│                          ↓                               │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Container Runtime (containerd or docker)         │  │
│  │  "Sure, here's your image. Don't lose it again."  │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                           │
                           │ HTTPS (because we're not savages)
                           ↓
              ┌─────────────────────────┐
              │   Your Docker Registry   │
              │   (Restored & Happy 😊)  │
              └─────────────────────────┘
```

## 🤓 How It Works (For the Curious)

### The Magic Flow

1. **Discovery Phase**: The syncer queries the Kubernetes API for all pods in your configured namespaces (*"Show me what you got!"*)
2. **Image Extraction**: Parses container specs to extract image references, including init containers (*because we're thorough like that*)
3. **Registry Check**: Performs a `HEAD` request to your registry to check if the image manifest exists (*knocking politely before barging in*)
4. **Decision Time**:
   - **If image exists**: Skip it (*"Looking good, nothing to do here"*)
   - **If from external registry** (e.g., Docker Hub): Copy it to your registry (*async mirroring, pretty neat*)
   - **If from your registry but missing**: 🚨 **RESTORATION MODE ACTIVATED** 🚨

### The Restoration Process

When an image needs restoration, here's what happens:

```go
// Simplified version (the real code has more error handling and tears)
1. Detect container runtime (containerd or docker)
   └─ "Are you containerd? Docker? Speak up!"

2. Export image to temporary tar file:
   ├─ containerd: ctr -n k8s.io images export /tmp/image.tar image:tag
   └─ docker:     docker save -o /tmp/image.tar image:tag

3. Load tar as OCI-compliant image
   └─ "Let me translate this into something the registry understands"

4. Push to registry using go-containerregistry
   └─ "Uploading... please don't let the network drop... SUCCESS!"

5. Delete temporary file
   └─ Clean up after ourselves like good citizens

6. Celebrate 🎉
   └─ Log success, increment Prometheus counter, feel good about life
```

### Why DaemonSet?

- Each node has its own local cache of images (it's like a distributed backup system, but accidental)
- DaemonSet ensures every node can restore images it has (democracy in action!)
- Even if the image only exists on one node, it gets restored (no image left behind!)
- Distributed backups without trying! (The best kind of backups)

## 🐛 Troubleshooting

### "Failed to detect container runtime socket"

**Problem**: The application can't find containerd or docker socket.

**Translation**: *"Where did you hide the socket?!"*

**Solution**:
1. Check which runtime your cluster uses:
   ```bash
   kubectl get nodes -o wide
   # Look at CONTAINER-RUNTIME column
   ```

2. Find the socket path on your nodes:
   ```bash
   # SSH to a node or use a debug pod
   find /run -name "*containerd*.sock" -o -name "docker.sock" 2>/dev/null
   ```

3. Configure explicitly in values.yaml:
   ```yaml
   containerd:
     socketPath: "/run/k3s/containerd/containerd.sock"  # for k3s
     # or "/run/docker.sock" for docker
     # or "/var/snap/microk8s/common/run/containerd.sock" for microk8s
   ```

### "ImagePullBackOff" for restored images

**Problem**: Image was restored but pods still can't pull it.

**Translation**: *"I put it back, why doesn't it work?!"*

**Possible causes**:
- 🔐 Registry authentication issues (*did you spell the password right?*)
- 🌐 Network connectivity problems (*can pods actually reach the registry?*)
- ⏱️ Timing issue (*image is being pushed right now, chill*)
- 🗜️ Image corruption during export (*rare, but Murphy's Law*)
- 🏷️ Wrong image tag (*nginx:latest vs nginx:1.21 - details matter*)

**Debug steps**:
```bash
# 1. Check syncer logs - did the push actually succeed?
kubectl logs -n kube-system -l app=push-missed-images | grep -i error

# 2. Manually try to pull the image
kubectl run test --image=your-registry.com/your-image:tag --rm -it -- /bin/sh

# 3. Check registry has the image (should return JSON with tags)
curl -u username:password https://your-registry.com/v2/your-image/tags/list

# 4. Check pod events for more details
kubectl describe pod <failing-pod-name>
```

### High Memory Usage

**Problem**: DaemonSet pods using lots of memory.

**Reason**: Image tar files are created in `/tmp` before pushing. Large images = large memory usage. Your 2GB base image isn't helping.

**Solutions**:
```yaml
# Option 1: Increase resource limits
resources:
  limits:
    memory: "512Mi"  # from 128Mi

# Option 2: Reduce sync concurrency in code
# Edit internal/syncer/syncer.go:
semaphore := make(chan struct{}, 3)  # from 5

# Option 3: Use smaller base images
# Seriously, do you REALLY need Ubuntu as a base image?
```

### Logs Show "Image already exists in target registry, skipping"

**Problem**: Everything is actually fine, you're just worrying too much.

**Translation**: *"Chill out, I already checked, everything's backed up."*

**Solution**: This is normal! It means your images are safe. Go get some coffee. ☕

### Permission Denied / RBAC Errors

**Problem**:
```
Error: failed to list pods: pods is forbidden: User "system:serviceaccount:kube-system:push-images-sa" cannot list resource "pods" in API group "" in the namespace "production"
```

**Translation**: *"The bouncer won't let me in."*

**Solution**: The ServiceAccount needs proper RBAC permissions. Our Helm chart creates these automatically, but if you installed manually or have strict Pod Security Policies:

```bash
# Check if ClusterRole exists
kubectl get clusterrole push-images-role

# Check if ClusterRoleBinding exists
kubectl get clusterrolebinding push-images-rolebinding

# If missing, apply the RBAC from our Helm chart
kubectl apply -f chart/templates/rbac.yaml
```

## 💡 Use Cases

### Disaster Recovery

**Scenario**: Your registry's storage backend failed at 2 AM on Saturday.

**Without this tool**:
- Panic 😱
- Try to restore from backups (you have backups, right? RIGHT?!)
- Spend 6 hours debugging S3 permissions
- Miss your kid's soccer game
- Contemplate career choices

**With this tool**:
- Wake up Monday
- See some alerts (registry was down)
- Check logs: "Successfully restored 47 images from container runtime"
- Have a relaxing breakfast
- Tell your boss you "handled the incident proactively"
- Still make the soccer game

### Image Mirroring

**Scenario**: You want to mirror external images to your registry for faster pulls and rate limit avoidance.

**Without this tool**: Write a complicated CI pipeline that tries to predict which images you'll need.

**With this tool**: Just deploy it. It automatically mirrors whatever your cluster actually uses. Lazy FTW!

### Compliance / Air-Gapped Environments

**Scenario**: Your security team wants all images in your internal registry. No external pulls allowed.

**Without this tool**: Manually push every image. Forget one. System breaks. Security team says "I told you so."

**With this tool**: It automatically copies external images to your registry. Security team is happy. You are hero.

## 🤝 Contributing

Found a bug? Have a feature idea? Want to roast our code? We accept all forms of contribution!

**We especially need help with:**
- 📝 Better documentation (*is that even possible?*)
- 🐛 Bug reports with reproducible steps (*"it doesn't work" is not a bug report*)
- ✨ Feature requests with clear use cases (*"make it faster" is not a feature*)
- 🧪 More test coverage (*yes, really*)
- 🎨 A logo (*we're developers, not designers*)
- 💰 Sponsorship (*coffee isn't free*)

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## 📜 License

MIT License - See [LICENSE](LICENSE) for details.

**Translation**: Do whatever you want with this code. Sell it, modify it, use it to take over the world. Just don't sue us if your registry catches fire. We're not responsible for your production incidents (even though we probably just prevented one).

## 🙏 Credits & Acknowledgments

Built with:
- ☕ **Coffee** (lots and lots of coffee)
- 😤 **Frustration** from actually losing registry images in production
- 💪 **[go-containerregistry](https://github.com/google/go-containerregistry)** - Because reinventing the wheel is overrated
- 🐳 **[containerd](https://containerd.io/)** - The runtime we deserved
- ⚓ **[client-go](https://github.com/kubernetes/client-go)** - For talking to K8s without crying
- 🎨 **[zerolog](https://github.com/rs/zerolog)** - For beautiful structured logs
- 📊 **[prometheus/client_golang](https://github.com/prometheus/client_golang)** - For metrics that actually matter
- 🧪 **Trial and error** (mostly error)
- 🌙 **Late nights** and **early mornings** (same thing really)

Special thanks to:
- Every DevOps engineer who ever accidentally `rm -rf`'d something important
- The person who invented container runtimes with local caching
- Our QA team (it's us, we're the QA team)
- Stack Overflow (you know why)

## 💬 FAQ

**Q: Does this work with Harbor/Nexus/GitLab Container Registry/Artifactory?**

A: Yes! Any OCI-compliant registry should work. If it speaks Docker Registry V2 API, we speak its language.

---

**Q: What about private registries with self-signed certificates?**

A: Configure your cluster's container runtime to trust the certificate. This tool uses the node's runtime, so if the node trusts it, we trust it. It's like a transitive property, but for SSL.

---

**Q: Can I use this to mirror images from Docker Hub?**

A: Sure! It will copy external images to your registry. Just:
- Don't violate Docker Hub's rate limits (they're watching 👀)
- Don't violate Docker Hub's ToS (they're still watching 👀)
- Consider using registry mirrors instead for better performance

---

**Q: Is this production-ready?**

A: Define "production-ready."

It has been tested in production environments (because we're the QA team now). It has proper error handling, metrics, health checks, and doesn't run as root. It's written in Go, not a bash script held together with duct tape.

But your definition of "production-ready" may vary based on:
- Your risk tolerance
- Your backup strategy
- Your boss's stress levels
- The phase of the moon

**Recommendation**: Test thoroughly in staging first. Have proper registry backups. Monitor the metrics. Start with non-critical namespaces.

---

**Q: Why Go instead of Bash/Python/Rust/[insert favorite language here]?**

A: Good question! Let's see:
- ✅ Go has excellent Kubernetes client libraries
- ✅ Compiles to a single static binary
- ✅ Great concurrency primitives (goroutines ftw)
- ✅ We already knew it
- ✅ The original Bash version was 300 lines of shell script that nobody wanted to maintain

Could we have used Python? Sure, but then we'd need Python runtime in the container.

Could we have used Rust? Absolutely, but compile times would test our patience.

Could we have kept using Bash? Please no, we have families to go home to.

---

**Q: Does this tool support Windows containers?**

A: 😂 Good one! Next question.

---

**Q: How does this handle image layers and deduplication?**

A: The container runtime (containerd/docker) handles all that magic. We just export whatever they give us and push it. The registry handles deduplication on its end. It's layers all the way down.

---

**Q: What happens if the same image exists on multiple nodes?**

A: Only one node will push it (whichever finishes the check first). The others will see "image already exists" and skip it. No duplicate work! We're efficient like that.

---

**Q: Can this detect if an image was tampered with or is corrupted?**

A: No. We trust the image as-is from the container runtime. If you need image validation and security scanning, use tools like:
- Trivy
- Clair
- Anchore
- Harbor's built-in scanning

---

**Q: Does this replace proper registry backups?**

A: **NO. NO. NO. A thousand times NO.**

This is a **safety net**, not a backup strategy. Think of it as:
- 🎯 This tool: "Oh no, the registry is missing this one image that's running right now, let me fix that"
- 💾 Proper backups: "The entire registry is gone, let me restore from last night's snapshot"

You still need:
- Regular registry backups
- Disaster recovery procedures
- Tested restore processes
- Prayers to the demo gods

---

**Q: Why is it called "push-from-k8s-back-to-docker-registry"?**

A: Because naming things is hard and we're bad at it. It accurately describes what it does though!

Better name suggestions welcome via PR to this README. Current runner-ups:
- "OopsIDeletedMyRegistry"
- "kubernetes-image-safety-net"
- "please-dont-fire-me"
- "registry-time-machine"

---

**Q: Can I use this with multiple registries?**

A: Currently, one target registry per deployment. But you can:
- Deploy multiple instances with different configs
- Submit a PR to add multi-registry support (we'll buy you a coffee)

---

**Q: What's the performance impact on my cluster?**

A: Minimal:
- Only runs every 10 minutes (configurable)
- Only lists pods and checks registry (lightweight API calls)
- Only exports/pushes when images are actually missing
- Uses minimal CPU/memory (50m CPU, 64Mi RAM default)

You'll probably notice more impact from that one node.js app that's leaking memory.

## 🚨 Known Limitations

Let's be honest about what this tool CAN'T do:

- ❌ **Cannot restore images from nodes that never pulled them**
  *If your node never had the image, we can't magic it out of thin air*

- ❌ **Won't help if your entire cluster is gone**
  *That's what actual backups are for*

- ❌ **Doesn't make coffee**
  *(yet - v3.0 feature maybe?)*

- ❌ **Can't predict which images you'll need**
  *It only backs up what's actually running*

- ❌ **Won't fix your Dockerfile**
  *FROM ubuntu:latest in production? That's on you*

- ❌ **Doesn't validate image security or integrity**
  *Use Trivy/Clair for that*

- ❌ **Won't help with image tag immutability issues**
  *If you push nginx:latest 47 times, that's a you problem*

**Use common sense**: Keep proper backups of your registry. This is a safety net, not a parachute.

## 🌟 Star History

If this tool saved your weekend, consider giving it a star! It's free and makes us feel good about our life choices.

[![Star History Chart](https://api.star-history.com/svg?repos=tazhate/push-from-k8s-back-to-docker-registry&type=Date)](https://star-history.com/#tazhate/push-from-k8s-back-to-docker-registry&Date)

## 📚 Additional Resources

- [Container Registry Spec](https://github.com/opencontainers/distribution-spec)
- [Kubernetes API Reference](https://kubernetes.io/docs/reference/)
- [containerd Documentation](https://containerd.io/docs/)
- [Why You Should Run Your Own Registry](https://medium.com/@tazhate/why-i-run-my-own-docker-registry)
- [Stack Overflow: "How to recover deleted Docker images"](https://stackoverflow.com/q/dont-delete-your-images) *(bookmark this)*

## 🎯 Roadmap

Future plans (no promises, but we're thinking about it):

- [ ] Multi-registry support
- [ ] Slack/webhook notifications
- [ ] Grafana dashboard in the repo
- [ ] Helm chart published to artifact hub
- [ ] Better documentation (is this even possible?)
- [ ] Image validation before push
- [ ] Namespace-specific sync periods
- [ ] Cost estimation per sync cycle
- [ ] Coffee maker integration
- [ ] Time machine to prevent registry deletions

Got ideas? [Open an issue!](https://github.com/tazhate/push-from-k8s-back-to-docker-registry/issues/new)

---

<div align="center">

**Made with 💙 and a healthy dose of**

***"this should not have happened but here we are"***

**by [@tazhate](https://github.com/tazhate)**

*If you've read this far, you deserve a cookie. 🍪*

*Actually, you deserve a raise. But we can only offer cookies.*

</div>

---

## 📞 Support

- **GitHub Issues**: [Report bugs or request features](https://github.com/tazhate/push-from-k8s-back-to-docker-registry/issues)
- **Discussions**: [Ask questions or share ideas](https://github.com/tazhate/push-from-k8s-back-to-docker-registry/discussions)
- **Email**: taz.inside@gmail.com
- **Twitter**: [@tazhate](https://twitter.com/tazhate) *(if you're into that)*
- **Buy me a coffee**: *(link coming soon if this gets popular)*

---

**⚡ Quick Links**
- [Installation](#-quick-start)
- [Configuration](#%EF%B8%8F-configuration)
- [Troubleshooting](#-troubleshooting)
- [Contributing](#-contributing)
- [FAQ](#-faq)

---

*P.S. - If this tool just saved your production environment at 3 AM, we'd love to hear your story. War stories make the best GitHub issues.*

*P.P.S. - Please actually set up proper registry backups. We're serious about that.*
