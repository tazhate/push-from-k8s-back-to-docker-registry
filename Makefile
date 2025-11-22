.PHONY: help build test lint clean docker-build docker-push helm-lint helm-package run deps

# Variables
BINARY_NAME=syncer
DOCKER_REGISTRY?=ghcr.io
DOCKER_IMAGE?=$(DOCKER_REGISTRY)/tazhate/push-from-k8s-back-to-docker-registry
VERSION?=v2.0.0
GO_VERSION=1.25.4

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Download Go dependencies
	go mod download
	go mod verify

build: ## Build the Go binary
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="-w -s -X main.Version=$(VERSION)" \
		-o bin/$(BINARY_NAME) \
		./cmd/syncer

build-local: ## Build for local OS
	go build -o bin/$(BINARY_NAME) ./cmd/syncer

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linters
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	golangci-lint run ./...

fmt: ## Format Go code
	go fmt ./...
	gofmt -s -w .

vet: ## Run go vet
	go vet ./...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	docker build -f Dockerfile.new -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

docker-push: ## Push Docker image to registry
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest

docker-run: ## Run Docker container locally
	docker run --rm -it \
		-e TARGET_REGISTRY_URL=ghcr.io \
		-e TARGET_REGISTRY_USERNAME=$(REGISTRY_USERNAME) \
		-e TARGET_REGISTRY_PASSWORD=$(REGISTRY_PASSWORD) \
		-e NAMESPACES=default \
		-e SYNC_PERIOD=1m \
		-e LOG_LEVEL=debug \
		-v ~/.kube:/home/nonroot/.kube:ro \
		$(DOCKER_IMAGE):$(VERSION)

helm-lint: ## Lint Helm chart
	helm lint chart/

helm-package: ## Package Helm chart
	helm package chart/ -d dist/

helm-install: ## Install Helm chart (for testing)
	helm upgrade --install image-sync chart/ \
		--namespace kube-system \
		--create-namespace \
		--set registry.url=$(REGISTRY_URL) \
		--set registry.username=$(REGISTRY_USERNAME) \
		--set registry.password=$(REGISTRY_PASSWORD) \
		--set monitor.namespaces[0]=default \
		--set logging.level=debug

helm-uninstall: ## Uninstall Helm chart
	helm uninstall image-sync -n kube-system

run: build-local ## Run locally (requires kubeconfig and env vars)
	./bin/$(BINARY_NAME)

dev: ## Run with live reload (requires air: go install github.com/cosmtrek/air@latest)
	air

# Development helpers
.PHONY: init
init: ## Initialize development environment
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest

.PHONY: check
check: fmt vet lint test ## Run all checks

.PHONY: all
all: clean deps check build docker-build ## Build everything

# CI/CD helpers
.PHONY: ci-test
ci-test: ## Run tests in CI environment
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

.PHONY: ci-build
ci-build: deps build docker-build ## Build in CI environment

# Get current git info
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_TAG=$(shell git describe --tags --always 2>/dev/null || echo "dev")

.PHONY: version
version: ## Show version information
	@echo "Version:    $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Git Tag:    $(GIT_TAG)"
	@echo "Go Version: $(GO_VERSION)"
