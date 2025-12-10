# KubeArchInspect Docker Extension Makefile

IMAGE_NAME ?= ajeetraina/kubearchinspect
TAG ?= latest
PLATFORMS ?= linux/amd64,linux/arm64

.PHONY: help build install uninstall update dev-ui dev-reset clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the extension image
	docker build --tag=$(IMAGE_NAME):$(TAG) .

build-multiarch: ## Build multi-architecture image
	docker buildx build --platform $(PLATFORMS) --tag=$(IMAGE_NAME):$(TAG) --push .

install: build ## Build and install the extension
	docker extension install $(IMAGE_NAME):$(TAG)

uninstall: ## Uninstall the extension
	docker extension uninstall $(IMAGE_NAME):$(TAG)

update: build ## Update the installed extension
	docker extension update $(IMAGE_NAME):$(TAG)

dev-ui: ## Enable UI hot-reload development
	cd ui && npm run dev &
	docker extension dev ui-source $(IMAGE_NAME):$(TAG) http://localhost:5173

dev-debug: ## Enable backend debugging
	docker extension dev debug $(IMAGE_NAME):$(TAG)

dev-reset: ## Reset development mode
	docker extension dev reset $(IMAGE_NAME):$(TAG)

ui-build: ## Build the UI only
	cd ui && npm install && npm run build

backend-build: ## Build the backend only
	cd backend && go build -o kubearchinspect .

clean: ## Clean build artifacts
	rm -rf ui/dist ui/node_modules backend/kubearchinspect

logs: ## Show extension logs
	docker extension logs $(IMAGE_NAME):$(TAG)

validate: ## Validate extension metadata
	docker extension validate $(IMAGE_NAME):$(TAG)
