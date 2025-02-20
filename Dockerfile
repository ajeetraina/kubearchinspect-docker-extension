FROM --platform=$BUILDPLATFORM golang:1.19-alpine AS builder
WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Create and set the backend directory
WORKDIR /build/backend

# Copy Go module files
COPY backend/go.mod backend/go.sum ./

# Initialize the module and verify dependencies
RUN go mod download -x && \
    go mod verify

# Copy the backend code
COPY backend/ ./

# Build the backend with verbose output
RUN GOOS=linux GOARCH=amd64 go build -a -v -tags netgo -ldflags '-w -extldflags "-static"' -o /bin/backend

FROM --platform=$BUILDPLATFORM node:18.12-alpine3.16 AS client-builder
WORKDIR /ui

# Copy package files first
COPY ui/package*.json ./
RUN npm install

# Copy the rest of the UI code
COPY ui/ .
RUN npm run build

FROM alpine:3.16
LABEL org.opencontainers.image.title="kubearchinspect" \
    org.opencontainers.image.description="Docker Extension to inspect Kubernetes resources for ARM compatibility" \
    org.opencontainers.image.vendor="Ajeet Singh Raina" \
    com.docker.desktop.extension.api.version="0.3.0" \
    com.docker.extension.screenshots="" \
    com.docker.extension.detailed-description="This extension helps you identify if your Kubernetes resources are ARM compatible" \
    com.docker.extension.publisher-url="https://github.com/ajeetraina" \
    com.docker.extension.additional-urls="" \
    com.docker.extension.changelog=""

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binary, compose file and metadata
COPY --from=builder /bin/backend /
COPY docker-compose.yaml .
COPY metadata.json .
COPY --from=client-builder /ui/build ui
COPY kubearchinspect.svg .

# Set secure permissions
RUN chmod 555 /backend

ENTRYPOINT ["/backend"]