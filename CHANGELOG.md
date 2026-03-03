# Changelog

All notable changes to FleetML will be documented in this file.

## [0.1.0] - 2026-03-02

### Added

**Core Platform**
- Edge agent with hardware fingerprinting (CPU arch, GPU, RAM, disk, OS)
- Control plane server with REST API (chi v5) and gRPC API
- PostgreSQL database with 8 migrations (users, fleets, devices, models, deployments, device_models, heartbeats, audit_log)
- MinIO (S3-compatible) model artifact storage
- JWT authentication with refresh tokens and API key support
- RBAC with admin, deployer, and viewer roles

**Model Management**
- Upload ONNX models via CLI or REST API
- SHA-256 checksum validation for model integrity
- Model versioning with artifact storage in S3
- Model registry with search and filtering

**Deployment**
- Immediate and canary deployment policies
- Canary stages with configurable percentages and durations (5% -> 50% -> 100%)
- Automatic rollback on canary failure
- Per-device deployment status tracking
- CLI `--wait` flag with progress bar

**Zero-Downtime Updates**
- Atomic model hot-swap using `sync/atomic.Pointer`
- Background model load + verify + swap (zero dropped inferences)
- Streamed S3 download with progress reporting
- Rollback to previous model version

**Offline Resilience**
- SQLite local buffer for heartbeats (pure-Go, no CGO)
- Store-and-forward communication wrapper
- Bulk sync on reconnection
- Connection health monitoring with exponential backoff

**Monitoring & Observability**
- Heartbeat-based device health monitoring (30s interval)
- CPU, GPU, RAM, disk, temperature metrics
- Offline detection (devices with stale heartbeats)
- Audit logging for key actions

**Dashboard**
- React 18 SPA with TypeScript and Tailwind CSS
- Fleet overview with health summary cards
- Device table with status filtering
- Model registry with upload instructions
- Deployment progress tracking
- Real-time updates via TanStack Query polling (10s)

**CLI**
- `fleetml init` — Configure server connection
- `fleetml deploy` — Upload and deploy models
- `fleetml status` — View fleet and deployment status
- `fleetml rollback` — Rollback deployments
- `fleetml logs` — View device logs (stub)
- `fleetml version` — Print version info

**Security**
- mTLS with self-signed CA for device authentication
- Per-device certificates with device_id in SAN
- JWT access + refresh tokens
- Token revocation support
- RBAC middleware on all endpoints

**Testing & CI**
- 500+ unit tests across agent, server, CLI, simulator
- Integration tests (registration, deploy, rollback)
- Virtual fleet simulator with 8 device profiles and 6 network profiles
- Fleet tests (heterogeneous, offline resilience)
- Chaos tests (network partition, rolling outage, degraded network)
- Scale tests (1000 devices)
- k6 load test for heartbeat endpoint
- E2E test scripts
- GitHub Actions CI (lint, test, cross-compile, security scan)
- govulncheck + Trivy container scanning

**Infrastructure**
- Docker Compose with health checks and restart policies
- Multi-stage Dockerfiles for agent, server, dashboard
- Cross-compilation for linux/amd64 and linux/arm64
- GitHub Actions release workflow with binaries and Docker images

### Not Included (Planned for v0.2.0)
- Multi-chip auto-compilation (TensorRT, OpenVINO, TFLite, SNPE, Hailo)
- A/B testing framework
- Data drift detection
- Policy engine
- MQTT fallback transport
- Embedded NATS for command queuing
- MLflow integration
