# FleetML Architecture

## Overview

FleetML uses a three-tier architecture designed for edge environments where devices may have limited resources, intermittent connectivity, and heterogeneous hardware.

```
┌─────────────────────────────────────────────────────┐
│                   Control Plane                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │ REST API │  │ gRPC API │  │Dashboard │          │
│  │  :8080   │  │  :50051  │  │  :3000   │          │
│  └────┬─────┘  └────┬─────┘  └──────────┘          │
│       │              │                               │
│  ┌────┴──────────────┴────┐                         │
│  │    Server (Go)          │                         │
│  │  Fleet | Model | Deploy │                         │
│  └────┬──────────────┬────┘                         │
│       │              │                               │
│  ┌────┴────┐    ┌────┴────┐                         │
│  │PostgreSQL│    │  MinIO  │                         │
│  │  :5432  │    │  :9000  │                         │
│  └─────────┘    └─────────┘                         │
└─────────────────────────────────────────────────────┘
              │                    │
              │   gRPC / MQTT      │
              │                    │
┌─────────────┴────────────────────┴──────────────────┐
│                   Edge Devices                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │ Agent    │  │ Agent    │  │ Agent    │          │
│  │ Jetson   │  │  RPi 4   │  │ Intel   │          │
│  │  Nano    │  │          │  │  NUC    │          │
│  └──────────┘  └──────────┘  └──────────┘          │
└─────────────────────────────────────────────────────┘
```

## Components

### Edge Agent (~15MB binary)

The agent runs on each edge device and handles:

- **Hardware fingerprinting** — Detects CPU arch, GPU, RAM, disk, OS
- **Heartbeat loop** — Reports health metrics every 30s
- **Model management** — Download, validate, hot-swap models
- **Offline resilience** — SQLite buffer for metrics, NATS for command queue
- **Zero-downtime swap** — Atomic model replacement during inference

Key design: Commands are piggybacked on heartbeat responses (no server-push needed for NAT-ed devices).

### Control Plane Server

The server provides:

- **REST API** (chi v5) — CRUD for models, devices, fleets, deployments
- **gRPC API** — Agent registration, bidirectional heartbeat streaming
- **Fleet Manager** — Device grouping, label-based selection
- **Model Registry** — Upload with SHA-256 checksums, S3 storage
- **Deployment Orchestrator** — Immediate and canary deployment policies
- **Monitoring** — Heartbeat processing, offline detection, audit logging

### CLI

Cobra-based CLI for all operations:

```
fleetml init        # Configure server connection
fleetml deploy      # Upload model and deploy
fleetml status      # View fleet/device/deployment status
fleetml rollback    # Rollback a deployment
fleetml logs        # View device logs
```

### Dashboard

React 18 SPA with:
- Fleet overview with health cards
- Device table with filtering
- Model registry with upload
- Deployment progress tracking
- Real-time metrics (TanStack Query polling)

## Key Design Decisions

### Offline-First

Agents buffer heartbeats in local SQLite when disconnected. On reconnection, buffered data is bulk-synced to the server. Commands received during offline periods are queued in an embedded NATS instance.

### Zero-Downtime Model Swap

Models are loaded in a background goroutine while the current model continues serving. Once loaded and verified (test inference), an `atomic.Pointer` swap makes the transition instant — zero dropped inferences.

### Separate Go Modules

Agent, server, and CLI each have their own `go.mod` to keep the agent binary small (~15MB). The agent doesn't import pgx, chi, or minio-go.

### Commands on Heartbeat Response

Instead of maintaining persistent connections for server-push, deploy commands are included in heartbeat responses. With a 30s heartbeat interval, the maximum command delivery delay is 30s — acceptable for model deployments.

### Canary Deployments

Deployments can use a canary policy (5% → 50% → 100%) where each stage is evaluated for success metrics before advancing. Failed canary stages trigger automatic rollback.

## Communication Flow

```
Agent                          Server
  │                              │
  ├──Register(device_info)──────>│
  │<──────────(agent_id, cert)───┤
  │                              │
  ├──Heartbeat(metrics)─────────>│
  │<──────────(commands[])───────┤
  │                              │
  ├──ReportStatus(deploy_id)────>│
  │                              │
  ├──BulkSync(buffered[])───────>│  (on reconnect)
  │<──────────(ack)──────────────┤
```

## Security

- **mTLS** — TLS 1.3 with per-device client certificates (device_id in SAN)
- **JWT** — Access + refresh tokens for API auth
- **RBAC** — Admin, deployer, viewer roles
- **SHA-256** — Model artifact integrity validation
- **API Keys** — Alternative to JWT for CI/CD pipelines
