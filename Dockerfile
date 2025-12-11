# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder
ENV CGO_ENABLED=0
WORKDIR /backend

# Install build dependencies
RUN apk add --no-cache git

# Copy Go module files
COPY backend/go.mod ./

# Download dependencies (this will regenerate go.sum)
RUN go mod download

# Copy the backend code
COPY backend/ ./

# Tidy up modules to ensure go.sum is complete
RUN go mod tidy

# Build the backend for multiple architectures
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /bin/service

# Build the frontend
FROM --platform=$BUILDPLATFORM node:18-alpine3.17 AS client-builder
WORKDIR /ui

# Copy package files and install dependencies
COPY ui/package.json ./
RUN npm install --legacy-peer-deps

# Copy UI source code
COPY ui/ ./

# Build the UI
RUN npm run build

# Final stage
FROM alpine:3.18

LABEL org.opencontainers.image.title="KubeArchInspect" \
    org.opencontainers.image.description="Docker Desktop Extension to check if container images in a Kubernetes cluster have ARM64 architecture support" \
    org.opencontainers.image.vendor="Ajeet Singh Raina" \
    com.docker.desktop.extension.api.version="0.3.4" \
    com.docker.desktop.extension.icon="https://raw.githubusercontent.com/ajeetraina/kubearchinspect-docker-extension/main/kubearchinspect.svg" \
    com.docker.extension.screenshots="[{\"alt\": \"KubeArchInspect Dashboard\", \"url\": \"https://raw.githubusercontent.com/ajeetraina/kubearchinspect-docker-extension/main/docs/screenshot.png\"}]" \
    com.docker.extension.detailed-description="KubeArchInspect is a Docker Desktop Extension that helps you identify which container images in your Kubernetes cluster have ARM64 architecture support. This is essential for planning migrations to ARM-based infrastructure like AWS Graviton or Apple Silicon." \
    com.docker.extension.publisher-url="https://github.com/ajeetraina" \
    com.docker.extension.additional-urls="[{\"title\":\"GitHub Repository\", \"url\":\"https://github.com/ajeetraina/kubearchinspect-docker-extension\"}, {\"title\":\"Original Project\", \"url\":\"https://github.com/ArmDeveloperEcosystem/kubearchinspect\"}]" \
    com.docker.extension.changelog="<p>Initial release with ARM64 compatibility checking for Kubernetes cluster images</p>" \
    com.docker.extension.categories="kubernetes,utility"

# Install runtime dependencies
RUN apk add --no-cache ca-certificates curl

# Download kubectl for ALL host platforms (these will run on the host, not in container)
# Get the stable kubectl version
RUN KUBECTL_VERSION=$(curl -L -s https://dl.k8s.io/release/stable.txt) && \
    # Linux amd64
    curl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" && \
    mkdir -p /linux && \
    chmod +x kubectl && \
    mv kubectl /linux/kubectl && \
    # macOS amd64 (Intel)
    curl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/darwin/amd64/kubectl" && \
    mkdir -p /darwin && \
    chmod +x kubectl && \
    mv kubectl /darwin/kubectl-amd64 && \
    # macOS arm64 (Apple Silicon)
    curl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/darwin/arm64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /darwin/kubectl && \
    # Windows amd64
    curl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/windows/amd64/kubectl.exe" && \
    mkdir -p /windows && \
    mv kubectl.exe /windows/kubectl.exe

# Copy backend binary
COPY --from=builder /bin/service /

# Copy metadata, icon, and docker-compose.yaml
COPY metadata.json .
COPY kubearchinspect.svg .
COPY docker-compose.yaml .

# Copy UI build output (vite config uses 'build' as outDir)
COPY --from=client-builder /ui/build ui

# Set secure permissions
RUN chmod 555 /service

# Start the HTTP server (not using Unix socket)
CMD ["/service"]
