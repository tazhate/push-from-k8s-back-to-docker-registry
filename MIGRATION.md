# Migration Guide: v1 (Bash) → v2 (Go)

## Overview

Version 2 is a complete rewrite in Go with significant improvements in security, performance, and observability.

## Key Differences

| Aspect | v1 (Bash) | v2 (Go) |
|--------|-----------|---------|
| Language | Bash scripts | Go 1.25.4 |
| Base Image | Alpine 3.18 (~500MB) | Distroless (~20MB) |
| Dependencies | docker + skopeo + ctr | go-containerregistry |
| Security | Privileged, root user | Non-privileged, non-root |
| Docker Socket | Required (RW) | Not needed |
| Containerd Socket | Required (RO) | Not needed |
| Metrics | None | Prometheus metrics |
| Health Checks | None | /healthz, /readyz |
| Logging | Plain text | Structured JSON |
| ConfigMap | Bash script (174 lines) | Not needed |

## Breaking Changes

### 1. Secret Keys Changed

**v1:**
```yaml
data:
  .dockerusername: <base64>
  .dockerpassword: <base64>
```

**v2:**
```yaml
stringData:
  username: <plain-text>
  password: <plain-text>
```

### 2. Environment Variables

No changes, all env vars remain the same.

### 3. Removed Components

- ConfigMap `push-images-script` - replaced by compiled Go binary
- Docker socket volume mount - no longer needed
- Containerd socket volume mount - no longer needed
- `privileged: true` security context

### 4. New Components

- Service for metrics (`:8080/metrics`)
- ServiceMonitor (optional, for Prometheus Operator)
- Health endpoints (`:8081/healthz`, `:8081/readyz`)
- Liveness and readiness probes

## Migration Steps

### Step 1: Backup Current Configuration

```bash
# Export current values
helm get values image-sync -n kube-system > old-values.yaml

# Backup secrets
kubectl get secret registry-credentials -n kube-system -o yaml > secret-backup.yaml
```

### Step 2: Uninstall v1

```bash
helm uninstall image-sync -n kube-system
```

**Note:** The pre-delete job in v1 may fail (it has a bug), but this is expected.

### Step 3: Clean Up Old Resources (if needed)

```bash
# Remove old configmap if it still exists
kubectl delete configmap push-images-script -n kube-system --ignore-not-found

# Verify daemonset is removed
kubectl get daemonset -n kube-system push-missed-images
```

### Step 4: Update Secret Format

If you're reusing the existing secret, update its format:

```bash
# Get current credentials
USERNAME=$(kubectl get secret registry-credentials -n kube-system -o jsonpath='{.data.\.dockerusername}' | base64 -d)
PASSWORD=$(kubectl get secret registry-credentials -n kube-system -o jsonpath='{.data.\.dockerpassword}' | base64 -d)

# Delete old secret
kubectl delete secret registry-credentials -n kube-system

# Create new secret with correct keys
kubectl create secret generic registry-credentials \
  --from-literal=username="$USERNAME" \
  --from-literal=password="$PASSWORD" \
  -n kube-system
```

### Step 5: Install v2

```bash
helm install image-sync ./chart \
  --namespace kube-system \
  --set registry.url=ghcr.io \
  --set registry.existingSecret=registry-credentials \
  --set monitor.namespaces[0]=production \
  --set monitor.namespaces[1]=staging \
  --set logging.level=info
```

### Step 6: Verify Installation

```bash
# Check DaemonSet status
kubectl get daemonset -n kube-system push-missed-images

# Check pods are running
kubectl get pods -n kube-system -l app=push-missed-images

# View logs
kubectl logs -n kube-system -l app=push-missed-images -f --tail=50

# Check metrics
kubectl port-forward -n kube-system daemonset/push-missed-images 8080:8080
curl localhost:8080/metrics
```

### Step 7: Monitor

Watch the logs for the first sync cycle:

```bash
kubectl logs -n kube-system -l app=push-missed-images -f | grep "sync cycle"
```

Expected output:
```json
{"level":"info","time":1234567890,"message":"Starting sync cycle"}
{"level":"info","count":15,"time":1234567890,"message":"Found images to process"}
{"level":"info","source":"nginx:1.21","target":"ghcr.io/nginx:1.21","time":1234567890,"message":"Successfully synced image"}
{"level":"info","duration":45000,"time":1234567890,"message":"Sync cycle completed"}
```

## Rollback Plan

If you need to rollback to v1:

1. Uninstall v2:
```bash
helm uninstall image-sync -n kube-system
```

2. Restore old secret format:
```bash
kubectl apply -f secret-backup.yaml
```

3. Reinstall v1:
```bash
helm install image-sync ./chart-v1 \
  --namespace kube-system \
  -f old-values.yaml
```

## Performance Comparison

Tested with 50 images on GKE cluster:

| Metric | v1 (Bash) | v2 (Go) |
|--------|-----------|---------|
| Sync Time | ~8 minutes | ~3 minutes |
| Memory Usage | 150-200 MB | 40-60 MB |
| CPU Usage | 100-200m | 20-50m |
| Image Size | 487 MB | 19.2 MB |
| Startup Time | 15-20s | 2-3s |

## Troubleshooting

### Issue: Pods in CrashLoopBackOff

**Check logs:**
```bash
kubectl logs -n kube-system -l app=push-missed-images --previous
```

**Common causes:**
1. Wrong secret keys (`username`/`password` vs `.dockerusername`/`.dockerpassword`)
2. Missing RBAC permissions
3. Invalid registry credentials

### Issue: Images not syncing

**Check metrics:**
```bash
kubectl port-forward -n kube-system daemonset/push-missed-images 8080:8080
curl localhost:8080/metrics | grep images_sync_failed
```

**Check logs with debug level:**
```bash
helm upgrade image-sync ./chart --reuse-values --set logging.level=debug
```

### Issue: High memory usage

**Reduce concurrent processing:**

Edit `internal/syncer/syncer.go`:
```go
semaphore := make(chan struct{}, 3)  // Reduce from 5 to 3
```

**Or increase memory limits:**
```bash
helm upgrade image-sync ./chart --reuse-values \
  --set resources.limits.memory=256Mi
```

## Questions?

- GitHub Issues: https://github.com/tazhate/push-from-k8s-back-to-docker-registry/issues
- See [README.md](README.md) for full documentation
