# ConfigMap Auto Merge Operator

A Kubernetes operator that automatically merges multiple ConfigMaps into a single consolidated ConfigMap based on annotations. Perfect for managing distributed configurations in Kubernetes clusters.

## Overview

The ConfigMap Auto Merge Operator simplifies configuration management by:
1. Watching for ConfigMaps with specific annotations
2. Automatically merging their contents into a target ConfigMap
3. Maintaining real-time synchronization when source ConfigMaps change
4. Providing Prometheus metrics for monitoring

### How It Works

The operator uses two annotations:
- `config-merger.k8s.io/watch: "true"` - Marks a ConfigMap for watching
- `config-merger.k8s.io/target: "target-name"` - Specifies the target ConfigMap name

When a ConfigMap is annotated, the operator:
1. Detects the annotation
2. Reads the target ConfigMap name
3. Merges the data with other watched ConfigMaps
4. Creates or updates the target ConfigMap
5. Maintains synchronization as sources change

### Example Usage

```yaml
# Source ConfigMap 1
apiVersion: v1
kind: ConfigMap
metadata:
  name: app1-config
  annotations:
    config-merger.k8s.io/watch: "true"
    config-merger.k8s.io/target: "merged-config"
data:
  database.url: "postgresql://db:5432"

---
# Source ConfigMap 2
apiVersion: v1
kind: ConfigMap
metadata:
  name: app2-config
  annotations:
    config-merger.k8s.io/watch: "true"
    config-merger.k8s.io/target: "merged-config"
data:
  redis.host: "redis:6379"

---
# Resulting Merged ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: merged-config
data:
  app1-config.database.url: "postgresql://db:5432"
  app2-config.redis.host: "redis:6379"
```

## Installation

### Prerequisites
- Kubernetes 1.19+
- Helm 3.0+
- kubectl configured with cluster access

### Quick Start

```bash
# Clone the repository
git clone https://github.com/mohamedragab2024/config-auto-merge-operator.git
cd config-auto-merge-operator

# Install using Helm
helm install config-merger ./helm-chart
```

## Development

### Building from Source

```bash
# Build the binary
make build

# Run tests
make test

# Build Docker image
make docker-build

# Run locally
make run
```

### Project Structure
```
.
├── cmd/
│   └── manager/          # Main entry point
├── pkg/
│   ├── controller/       # Kubernetes controller logic
│   ├── handler/          # ConfigMap handling
│   ├── metrics/          # Prometheus metrics
│   └── utils/            # Utility functions
└── helm-chart/          # Helm deployment chart
```

## Monitoring

The operator exposes Prometheus metrics on port 8080 at `/metrics`:

| Metric | Description |
|--------|-------------|
| `configmap_operator_operations_total` | Number of operations performed |
| `configmap_operator_processing_duration_seconds` | Operation processing time |
| `configmap_operator_errors_total` | Number of errors encountered |
| `configmap_operator_merged_size_bytes` | Size of merged ConfigMaps |
| `configmap_operator_watched_configmaps` | Number of ConfigMaps being watched |

### ServiceMonitor Configuration

```yaml
serviceMonitor:
  enabled: true
  interval: 30s
  additionalLabels:
    release: prometheus
```

## Configuration

### Helm Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of operator replicas | `1` |
| `image.repository` | Image repository | `mohamedragab2024/config-auto-merge-operator` |
| `image.tag` | Image tag | `latest` |
| `resources.limits.cpu` | CPU limit | `200m` |
| `resources.limits.memory` | Memory limit | `256Mi` |
| `serviceMonitor.enabled` | Enable ServiceMonitor | `false` |

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Submit a pull request
