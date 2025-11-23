# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /syncer \
    ./cmd/syncer

# Final stage
FROM alpine:3.22

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    containerd \
    docker-cli \
    tini

# Create symlinks for ctr
RUN ln -sf /usr/bin/ctr /usr/local/bin/ctr

# Copy binary from builder
COPY --from=builder /syncer /usr/local/bin/syncer

# Use tini as init
ENTRYPOINT ["/sbin/tini", "--"]

# Run the syncer
CMD ["/usr/local/bin/syncer"]
