# KubeArchInspect Docker Extension

A Docker Desktop extension to inspect Kubernetes resources for ARM compatibility.

## Architecture

```mermaid
graph TD
    A[Docker Desktop] --> B[KubeArchInspect Extension]
    B --> C[Frontend/UI]
    B --> D[Backend Service]
    
    C --> E[Components]
    C --> F[Services]
    
    E --> G[Statistics Dashboard]
    E --> H[Resource Table]
    E --> I[Filters]
    
    F --> J[API Client]
    F --> K[Export Service]
    
    D --> L[Kubernetes Client]
    D --> M[Resource Inspector]
    D --> N[API Server]
```

## Features

- Inspects various Kubernetes resources for ARM compatibility
- Real-time statistics and visualizations
- Filterable and sortable resource table
- Export capabilities

## Development

### Prerequisites

- Docker Desktop
- Node.js 18+
- Go 1.19+

### Building

```bash
docker buildx build -t kubearchinspect:latest .
docker extension install kubearchinspect:latest
```