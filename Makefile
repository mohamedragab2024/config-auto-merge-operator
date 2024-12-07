# Build information
VERSION ?= 0.1.0
REGISTRY ?= mohamedragab2024
IMAGE_NAME ?= config-auto-merge-operator
IMAGE_TAG ?= $(VERSION)
IMG ?= $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=manager

# Kubernetes parameters
KUBECONFIG ?= ~/.kube/config

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

.PHONY: all build clean test coverage lint docker-build docker-push run help tidy

all: test build

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) cmd/manager/main.go

clean: ## Clean build files
	$(GOCLEAN)
	rm -rf bin/

test: ## Run unit tests
	$(GOTEST) -v ./...

coverage: ## Run tests with coverage
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

lint: ## Run linters
	golangci-lint run

tidy: ## Tidy up go.mod
	$(GOMOD) tidy

docker-build: ## Build docker image
	docker build -t $(IMG) .

docker-push: ## Push docker image
	docker push $(IMG)

run: ## Run locally
	$(GOBUILD) -o bin/$(BINARY_NAME) cmd/manager/main.go
	./bin/$(BINARY_NAME)

install: ## Install CRDs and deploy operator
	kubectl apply -f helm-chart/crds/
	helm install $(IMAGE_NAME) helm-chart/ --namespace default

uninstall: ## Uninstall operator
	helm uninstall $(IMAGE_NAME) --namespace default
	kubectl delete -f helm-chart/crds/

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

generate: ## Run code generation
	go generate ./...

# Development targets
dev-deps: ## Install development dependencies
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Docker development targets
dev-docker: ## Build and run in docker
	docker build -t $(IMG) .
	docker run --rm -it $(IMG)

# Default target
.DEFAULT_GOAL := help 
