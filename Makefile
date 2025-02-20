# Variables
EXTENSION_NAME=kubearchinspect
VERSION=0.1.0
IMAGE_NAME=$(EXTENSION_NAME):$(VERSION)

# Build the extension
build-extension: 
	docker buildx build \
		--platform=linux/amd64,linux/arm64 \
		-t $(IMAGE_NAME) \
		--build-arg EXTENSION_NAME=$(EXTENSION_NAME) \
		.

# Install the extension
install-extension: build-extension
	docker extension install $(IMAGE_NAME)

# Update the installed extension
update-extension: build-extension
	docker extension update $(IMAGE_NAME)

# Remove the extension
remove-extension:
	docker extension rm $(IMAGE_NAME)

# Build frontend
build-frontend:
	cd ui && npm install && npm run build

# Build backend
build-backend:
	cd backend && go build -o bin/$(EXTENSION_NAME)

# Run tests
test: test-backend test-frontend

test-backend:
	cd backend && go test ./...

test-frontend:
	cd ui && npm test

# Development mode
dev:
	docker extension dev enable $(EXTENSION_NAME)

debug:
	docker extension dev debug $(EXTENSION_NAME)

# Clean up
clean:
	rm -rf ui/build backend/bin
	- docker extension rm $(IMAGE_NAME)

.PHONY: build-extension install-extension update-extension remove-extension build-frontend build-backend test test-backend test-frontend dev debug clean
