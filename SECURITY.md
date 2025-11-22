# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 2.x.x   | :white_check_mark: |
| 1.x.x   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to: **taz.inside@gmail.com**

Please include:

- Type of vulnerability
- Full paths of source file(s) related to the vulnerability
- Location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### What to Expect

- Acknowledgment of your report within 48 hours
- Regular updates on our progress
- Credit in the security advisory (if desired)
- A fix and disclosure timeline

## Security Best Practices

When using this tool:

1. **Never commit credentials** to version control
   - Use Kubernetes Secrets
   - Consider external secret managers (Vault, AWS Secrets Manager, etc.)

2. **Use RBAC with least privilege**
   - Only grant necessary permissions
   - Review the ClusterRole regularly

3. **Enable Pod Security Standards**
   ```yaml
   apiVersion: v1
   kind: Namespace
   metadata:
     labels:
       pod-security.kubernetes.io/enforce: restricted
   ```

4. **Network Policies**
   - Restrict egress to only target registry
   - Limit API server access if possible

5. **Image Scanning**
   - Scan container images for vulnerabilities
   - Use tools like Trivy, Clair, or Anchore

6. **Audit Logs**
   - Enable Kubernetes audit logging
   - Monitor sync operations

7. **TLS/HTTPS**
   - Always use HTTPS for registry connections
   - Validate certificates properly

8. **Regular Updates**
   - Keep the tool updated to latest version
   - Monitor security advisories

## Known Security Considerations

### Socket Access

This tool requires access to the container runtime socket (containerd or Docker). This is by design to export images from local cache.

**Mitigation**:
- Socket is mounted read-only where possible
- No privileged mode required
- Runs as non-root user

### Registry Credentials

Registry credentials are stored in Kubernetes Secrets.

**Best Practices**:
- Rotate credentials regularly
- Use registry-specific service accounts
- Consider using ImagePullSecrets instead of global credentials

### RBAC Permissions

The tool requires cluster-wide read access to pods.

**Mitigation**:
- ClusterRole limits permissions to `get`, `list`, `watch` only
- No write permissions granted
- Can be scoped to specific namespaces if needed

## Security Updates

Security updates will be released as:
- Patch versions for critical vulnerabilities (e.g., 2.3.1)
- Minor versions for moderate issues (e.g., 2.4.0)

Subscribe to:
- GitHub Security Advisories
- Release notifications
- GitHub Watch → Custom → Security alerts

## Vulnerability Disclosure Timeline

1. **Day 0**: Vulnerability reported privately
2. **Day 1-2**: Initial triage and acknowledgment
3. **Day 3-7**: Develop and test fix
4. **Day 7-14**: Release patched version
5. **Day 14+**: Public disclosure

This timeline may be adjusted based on severity and complexity.

## Security Hall of Fame

We thank the following security researchers for responsibly disclosing vulnerabilities:

<!-- List will be updated as vulnerabilities are reported and fixed -->

_No vulnerabilities reported yet (knock on wood!)_

---

For general questions about security, please open a GitHub Discussion.
