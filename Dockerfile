# Dockerfile

# Use the official Alpine 3.18 as the base image
FROM alpine:3.18

# Install runtime dependencies, containerd, skopeo, and tini
RUN apk add --no-cache \
    bash \
    curl \
    jq \
    ca-certificates \
    gnupg \
    tini \
    su-exec \
    libc6-compat \
    docker \
    skopeo \ 
    htop

# Install kubectl
RUN VERSION=$(curl -L -s https://dl.k8s.io/release/stable.txt) && \
    curl -LO "https://dl.k8s.io/release/${VERSION}/bin/linux/amd64/kubectl" && \
    install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && \
    rm kubectl

# Install crictl
RUN VERSION="v1.24.0" && \
    curl -LO https://github.com/kubernetes-sigs/cri-tools/releases/download/${VERSION}/crictl-${VERSION}-linux-amd64.tar.gz && \
    tar -zxvf crictl-${VERSION}-linux-amd64.tar.gz -C /usr/local/bin && \
    chmod +x /usr/local/bin/crictl && \
    rm crictl-${VERSION}-linux-amd64.tar.gz

# Install containerd
RUN VERSION="1.6.9" && \
    curl -LO https://github.com/containerd/containerd/releases/download/v${VERSION}/containerd-${VERSION}-linux-amd64.tar.gz && \
    rm /usr/bin/containerd && \
    tar -zxvf containerd-${VERSION}-linux-amd64.tar.gz -C /usr/local && \
    rm containerd-${VERSION}-linux-amd64.tar.gz

# Create symlinks for containerd and ctr
RUN ln -s /usr/local/bin/containerd /usr/bin/containerd && \
    ln -s /usr/local/bin/ctr /usr/bin/ctr

# Create a symlink for tini
RUN ln -s /sbin/tini /usr/bin/tini

# Verify skopeo installation
RUN skopeo --version

# Copy entrypoint script
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# Set the default command to execute the synchronization script
CMD ["/bin/bash", "/scripts/push_images.sh"]
