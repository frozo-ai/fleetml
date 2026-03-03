# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

FleetML is an open-source, chip-neutral edge MLOps platform — "Kubernetes for edge AI models." It deploys, updates, and monitors ML models across heterogeneous edge device fleets. The project is currently **pre-development** with three planning documents serving as implementation specifications:

- **FleetML-PRD.md** — Primary implementation guide. Contains monorepo structure, database schemas, protobuf definitions, Go interfaces, REST API contracts, CLI commands, dashboard pages, communication protocols, security specs, Docker configs, and a 187+ test plan.
- **FleetML-Master-Plan.md** — Product vision, architecture, go-to-market, execution roadmap.
- **FleetML-Testing-Strategy.md** — 9-layer testing pyramid from unit tests to hardware-in-the-loop.

## Architecture (Three-Tier)

```
Control Plane (Go server, REST + gRPC)  ←→  Network (gRPC primary, MQTT fallback)  ←→  Edge Agents (Go, ~15MB binary)
```

**Key architectural decisions:**
- **Offline-first**: SQLite local store + NATS local queue on agent; bulk sync on reconnect
- **Zero-downtime model swap**: `sync/atomic.Pointer` for hot-swap during inference
- **Chip-neutral compilation**: ONNX as universal input → compile to TensorRT/OpenVINO/TFLite/SNPE/Hailo
- **Policy-driven deployments**: YAML policies define canary stages (5% → 50% → 100%), rollback triggers, hardware constraints
- **Store-and-forward**: Heartbeats buffered locally during disconnection, commands queued in NATS

## Monorepo Components

| Component | Path | Language | Framework | Purpose |
|---|---|---|---|---|
| Edge Agent | `agent/` | Go 1.22+ | gRPC, NATS, SQLite | Runs on devices; model loading, heartbeats, offline operation |
| Control Plane | `server/` | Go 1.22+ | chi v5, gRPC, pgx/v5 | REST API (port 8080), gRPC (port 50051), fleet management |
| CLI | `cli/` | Go 1.22+ | cobra v1.8+ | `fleetml init/deploy/status/logs/rollback/ab-test/version` |
| Dashboard | `dashboard/` | TypeScript/React 18 | Vite 5, Tailwind 3, TanStack Query | Web UI for fleet management |
| Compiler Service | `compiler/` | Python 3.11 | FastAPI | Multi-chip model compilation |
| Protobuf Defs | `proto/fleetml/v1/` | Protobuf v3 | buf | Agent, models, devices, deployments, common protos |
| Fleet Simulator | `simulator/` | Go | Docker Compose | Virtual fleet for testing |

Each Go component (`agent/`, `server/`, `cli/`) has its own `go.mod`. Dashboard has its own `package.json`.

## Build & Development Commands

All commands are defined in the top-level `Makefile` (see PRD section 14 for full spec).

```bash
# Build everything
make build              # agent + server + cli + dashboard

# Individual builds
make agent              # cross-compile for linux/amd64 + linux/arm64
make server
make cli
make dashboard          # npm install && npm run build

# Testing
make test               # unit + integration
make test-unit          # go test -race -short ./... (agent, server, cli)
make test-integration   # docker compose up → go test -tags=integration → down
make test-fleet         # go test -tags=fleet ./fleet/... -fleet-size=20
make test-chaos         # go test -tags=chaos ./chaos/... -timeout=60m

# Linting
make lint               # golangci-lint (Go) + npm run lint (dashboard)

# Run single Go test
cd agent && go test -race -run TestFunctionName ./internal/model/...

# Docker
make docker-build       # fleetml/agent, fleetml/server, fleetml/dashboard images

# Local dev environment
make dev                # docker compose up db minio, then server + dashboard
```

**Dashboard-specific:**
```bash
cd dashboard
npm run dev             # local dev server with HMR
npm run build           # production build
npm run lint            # ESLint
```

## Infrastructure Dependencies (via Docker Compose)

- **PostgreSQL 16** — primary database (port 5432)
- **MinIO** — S3-compatible model storage (port 9000, console 9001)
- **NATS 2.10+** — message bus + offline command queue

## Key Entry Points

- `agent/cmd/agent/main.go` — Agent: starts gRPC client, heartbeat loop, model runtime, command listener
- `server/cmd/server/main.go` — Server: starts REST (8080), gRPC (50051), Prometheus (9090)
- `cli/cmd/root.go` — CLI: cobra root with subcommands
- `dashboard/src/main.tsx` — Dashboard: React root with React Router 6
- `compiler/main.py` — Compiler: FastAPI, `POST /compile` endpoint

## Performance Targets

- Agent binary: <15MB stripped, <30MB RSS idle, <2s startup
- Heartbeat: <1KB gzipped, <1% CPU
- API: p50 <20ms, p99 <100ms
- Deploy: 1 device <30s, 100 devices <2min, 1000 devices <10min
- Hot-swap: 0 dropped inferences
- Reconnection after offline: <10s

## Development Phases

**Phase 1 (MVP, Weeks 1-16):** Agent core, gRPC comms, CLI, control plane, REST API, PostgreSQL, S3/MinIO, React dashboard, OTA updates, hot-swap, rollback. Uses ONNX Runtime only. Excludes multi-chip compilation, A/B testing, drift detection, policy engine.

**Phase 2 (Months 5-9):** TensorRT/OpenVINO/TFLite compilers, A/B testing, device grouping, drift detection, MLflow integration, policy engine v1.

## CI/CD Pipeline Tiers

- **Every commit (~2 min):** lint + unit tests + cross-compile + build
- **Every PR (~10 min):** integration tests + virtual fleet (20 devices) + security scan (govulncheck + trivy)
- **Nightly (~60 min):** fleet at scale (500 devices) + chaos engineering + benchmarks + k6 load test
- **Pre-release (manual):** hardware-in-the-loop (Jetson + RPi) + E2E scenarios + 24h soak test

**Quality gates:** 500+ unit tests pass, integration + virtual fleet pass, no govulncheck vulns, no lint errors, coverage >= 80%.

## Testing Strategy

9-layer pyramid (see FleetML-Testing-Strategy.md for full details):
1. Unit tests (500+, every commit)
2. Integration tests (100+, every PR, Docker-based)
3. Virtual fleet simulator (50+ scenarios, every PR)
4. Hardware-in-the-loop (pre-release only)
5. Chaos engineering (nightly)
6. Performance/scale (nightly)
7. ML model-specific
8. Security
9. E2E scenarios (5-10, pre-release)

Use build tags (`-tags=integration`, `-tags=fleet`, `-tags=chaos`, `-tags=benchmark`) to separate test layers.

## Auth & Security

- JWT + mTLS (TLS 1.3) for API auth
- RBAC: three roles — admin, deployer, viewer
- Model artifacts validated via SHA-256 checksums
- Agent ↔ Server communication encrypted via mTLS
