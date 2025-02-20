FROM --platform=$BUILDPLATFORM golang:1.19-alpine AS builder
WORKDIR /backend

# Copy Go module files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the backend code
COPY backend/ .

# Build the backend
RUN CGO_ENABLED=0 go build -o /bin/backend

FROM --platform=$BUILDPLATFORM node:18.12-alpine3.16 AS client-builder
WORKDIR /ui

# Copy package files first for better caching
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

# Copy binary, compose file and metadata
COPY --from=builder /bin/backend /
COPY docker-compose.yaml .
COPY metadata.json .
COPY --from=client-builder /ui/build ui
COPY kubearchinspect.svg .

ENTRYPOINT ["/backend"]