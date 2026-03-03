# FleetML

**Kubernetes for Edge AI Models** — Deploy, update, and monitor ML models across heterogeneous edge device fleets.

FleetML is an open-source, chip-neutral edge MLOps platform that manages the full lifecycle of ML models on edge devices: from uploading ONNX models to zero-downtime OTA updates with canary deployments and automatic rollback.

## Features

- **Chip-neutral** — ONNX as universal input; supports Jetson, Raspberry Pi, Intel NUC, Hailo, Qualcomm
- **Zero-downtime model swap** — Atomic pointer swap ensures zero dropped inferences during updates
- **Canary deployments** — Progressive rollout (5% -> 50% -> 100%) with automatic rollback on failure
- **Offline-first** — SQLite buffer + store-and-forward; agents survive network disconnects
- **Fleet management** — Group devices by labels, target deployments by fleet or label selector
- **Real-time monitoring** — Dashboard with device health, metrics, and deployment progress
- **Secure** — mTLS, JWT auth, RBAC (admin/deployer/viewer), SHA-256 model integrity

## Quick Start

```bash
# Start the control plane
git clone https://github.com/fleetml/fleetml.git && cd fleetml
cp .env.example .env
docker compose up -d

# Deploy a model
fleetml deploy my-model.onnx --fleet default --wait
```

See the [Quickstart Guide](docs/quickstart.md) for the full walkthrough.

## Architecture

```
Control Plane (Go)          Edge Devices (Go, ~15MB)
┌────────────────┐          ┌──────────────────┐
│  REST + gRPC   │◄────────►│  Agent           │
│  Dashboard     │  gRPC/   │  ONNX Runtime    │
│  PostgreSQL    │  MQTT    │  SQLite Buffer   │
│  MinIO (S3)    │          │  Hot-Swap Engine  │
└────────────────┘          └──────────────────┘
```

## Components

| Component | Description |
|-----------|-------------|
| `agent/` | Edge agent — hardware detection, heartbeats, model loading, offline resilience |
| `server/` | Control plane — REST/gRPC APIs, fleet management, deployment orchestration |
| `cli/` | CLI — `fleetml init/deploy/status/rollback/logs` |
| `dashboard/` | Web UI — React dashboard with fleet overview, metrics, deployments |
| `compiler/` | Model compiler service (Python/FastAPI) |
| `simulator/` | Virtual fleet simulator for testing |
| `proto/` | Protobuf definitions for agent-server communication |

## CLI Commands

```bash
fleetml init                    # Configure server connection
fleetml deploy model.onnx      # Upload and deploy a model
fleetml status                  # View fleet status
fleetml rollback --deployment X # Rollback a deployment
fleetml logs --device X         # View device logs
```

## Development

```bash
# Build everything
make build

# Run tests
make test-unit          # Unit tests
make test-integration   # Integration tests (needs Docker)
make test-fleet         # Virtual fleet (20 devices)

# Lint
make lint
```

See [Development Guide](docs/development.md) for details.

## Documentation

- [Quickstart](docs/quickstart.md)
- [Installation](docs/installation.md)
- [Architecture](docs/architecture.md)
- [CLI Reference](docs/cli-reference.md)
- [API Reference](docs/api-reference.md)
- [Development](docs/development.md)

## Performance Targets

| Metric | Target |
|--------|--------|
| Agent binary | <15MB stripped |
| Agent memory | <30MB RSS idle |
| Agent startup | <2s |
| Heartbeat overhead | <1KB gzipped, <1% CPU |
| API latency | p50 <20ms, p99 <100ms |
| Deploy 1 device | <30s |
| Deploy 100 devices | <2min |
| Hot-swap | 0 dropped inferences |

## License

Apache License 2.0 — see [LICENSE](LICENSE).
