FROM bash:latest

RUN apk add --no-cache curl jq && \
    curl -LO https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.24.0/crictl-v1.24.0-linux-amd64.tar.gz && \
    tar -zxvf crictl-v1.24.0-linux-amd64.tar.gz -C /usr/local/bin && \
    chmod +x /usr/local/bin/crictl

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

ENTRYPOINT ["/bin/bash"]
