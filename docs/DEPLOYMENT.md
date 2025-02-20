# Deployment Guide

## Prerequisites

- Docker Desktop 4.12.0 or later
- Kubernetes cluster access configured
- Node.js 18.x for development
- Go 1.19.x for development

## Installation

### From Docker Hub

```bash
docker extension install ajeetraina/kubearchinspect:latest
```

### From Source

1. Clone the repository:
```bash
git clone https://github.com/ajeetraina/kubearchinspect-docker-extension.git
cd kubearchinspect-docker-extension
```

2. Build the extension:
```bash
docker buildx build -t kubearchinspect:latest .
```

3. Install the extension:
```bash
docker extension install kubearchinspect:latest
```

## Development Setup

1. Enable development mode:
```bash
docker extension dev enable kubearchinspect
```

2. Start the development server:
```bash
# Terminal 1 - Backend
cd backend
go run main.go

# Terminal 2 - Frontend
cd ui
npm install
npm run dev
```

3. Debug logs:
```bash
docker extension dev debug kubearchinspect
```

## Usage Examples

1. Basic resource inspection:
```bash
# Via UI
1. Open Docker Desktop
2. Go to Extensions
3. Click on KubeArchInspect
4. Click 'Inspect Resources'

# Via CLI
curl http://localhost:8080/api/v1/inspect
```

2. Filtered inspection:
```bash
# Via UI
1. Use the namespace dropdown
2. Select resource types
3. Apply ARM compatibility filter

# Via CLI
curl http://localhost:8080/api/v1/inspect?namespace=default&type=Deployment
```

3. Export results:
```bash
# Via UI
1. Click Export button
2. Select format (CSV/JSON)

# Via CLI
curl -X POST http://localhost:8080/api/v1/export \
  -H 'Content-Type: application/json' \
  -d '{"format":"csv"}' \
  --output results.csv
```
