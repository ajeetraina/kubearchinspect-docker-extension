# KubeArchInspect Docker Extension

A Docker Desktop extension to inspect Kubernetes resources for ARM compatibility.

## Architecture

### Overall System Architecture
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
    
    L --> O[Kubernetes Cluster]
    
    M --> P[Deployments]
    M --> Q[StatefulSets]
    M --> R[DaemonSets]
    M --> S[CronJobs]
    M --> T[Jobs]
    M --> U[Pods]
```

### Inspection Sequence
```mermaid
sequenceDiagram
    participant U as User
    participant UI as Frontend
    participant B as Backend
    participant K as Kubernetes API
    
    U->>UI: Click Inspect Resources
    UI->>B: GET /inspect
    B->>K: List Namespaces
    K-->>B: Namespaces
    
    loop Each Namespace
        B->>K: List Resources
        K-->>B: Resources
        B->>B: Check ARM Compatibility
    end
    
    B-->>UI: Resource List
    UI->>UI: Update Statistics
    UI->>UI: Update Table
    UI-->>U: Show Results
```

### UI Component Structure
```mermaid
classDiagram
    class App {
        -resources: Resource[]
        -loading: boolean
        -error: string
        +inspectResources()
    }
    
    class Statistics {
        -resources: Resource[]
        +calculateStats()
        +renderCharts()
    }
    
    class FilterableResourceTable {
        -resources: Resource[]
        -filters: FilterState
        -sortConfig: SortConfig
        +handleSort()
        +handleFilter()
    }
    
    class ResourceTable {
        -resources: Resource[]
        -onSort: function
        +renderRow()
    }
    
    App --> Statistics
    App --> FilterableResourceTable
    FilterableResourceTable --> ResourceTable
```

## Features

- Inspects various Kubernetes resources for ARM compatibility
- Real-time statistics and visualizations
- Filterable and sortable resource table
- Export capabilities
- Detailed resource information

## Development

### Prerequisites

- Docker Desktop
- Node.js 18+
- Go 1.19+

### Building

```bash
# Build the extension
docker buildx build -t kubearchinspect:latest .

# Install the extension
docker extension install kubearchinspect:latest
```

### Development Mode

```bash
# Enable development mode
docker extension dev enable kubearchinspect

# View logs
docker extension dev debug kubearchinspect
```

## Testing

```bash
# Backend tests
go test ./...

# Frontend tests
cd ui && npm test
```

## Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests.

## License

MIT
