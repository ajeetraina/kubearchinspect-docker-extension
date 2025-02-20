# API Documentation

## Endpoints

### GET /api/v1/inspect

Inspects Kubernetes resources for ARM compatibility.

**Parameters:**
- namespace (optional): Filter by namespace
- type (optional): Filter by resource type

**Response:**
```json
{
  "resources": [
    {
      "name": "string",
      "namespace": "string",
      "kind": "string",
      "isArmCompatible": boolean,
      "image": "string"
    }
  ]
}
```

### GET /api/v1/statistics

Returns aggregated statistics about ARM compatibility.

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
  }
}
```