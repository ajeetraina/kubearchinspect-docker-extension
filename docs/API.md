# API Documentation

## Endpoints

### Resource Inspection

#### GET /api/v1/inspect
Inspects Kubernetes resources for ARM compatibility.

**Query Parameters:**
- `namespace` (optional): Filter by namespace
- `type` (optional): Filter by resource type
- `detailed` (optional): Include detailed analysis

**Response:**
```json
{
  "resources": [
    {
      "name": "string",
      "namespace": "string",
      "kind": "string",
      "isArmCompatible": boolean,
      "image": "string",
      "details": {
        "architecture": "string",
        "reasons": ["string"]
      }
    }
  ],
  "metadata": {
    "timestamp": "string",
    "clusterInfo": {}
  }
}
```

### Statistics

#### GET /api/v1/statistics
Returns aggregated statistics about ARM compatibility.

**Query Parameters:**
- `timeRange` (optional): Time range for historical data

**Response:**
```json
{
  "summary": {
    "total": number,
    "armCompatible": number,
    "nonArmCompatible": number
  },
  "byType": {
    "Deployment": {
      "total": number,
      "armCompatible": number
    }
  },
  "byNamespace": {}
}
```

### Export

#### POST /api/v1/export
Exports inspection results in various formats.

**Request Body:**
```json
{
  "format": "csv|json|xlsx",
  "filters": {
    "namespace": "string",
    "type": "string"
  }
}
```

**Response:**
File download in requested format
