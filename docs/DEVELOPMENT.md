# Development Guide

## Setup

1. Install dependencies:
```bash
# Backend
go mod download

# Frontend
cd ui && npm install
```

2. Run in development mode:
```bash
# Backend
go run main.go

# Frontend
npm run dev
```

## Building

```bash
# Build extension
docker buildx build -t kubearchinspect:latest .

# Install extension
docker extension install kubearchinspect:latest
```

## Testing

```bash
# Backend tests
go test ./...

# Frontend tests
cd ui && npm test
```