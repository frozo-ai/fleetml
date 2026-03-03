# FleetML — Product Requirements Document (PRD)

## For Coding Agent Implementation

**Version:** 1.0
**Date:** February 22, 2026
**Status:** Implementation-Ready
**Target:** AI Coding Agent

---

## Table of Contents

1. [Product Overview](#1-product-overview)
2. [Monorepo Structure](#2-monorepo-structure)
3. [Technology Stack](#3-technology-stack)
4. [Database Schema](#4-database-schema)
5. [Protobuf & gRPC Service Definitions](#5-protobuf--grpc-service-definitions)
6. [Agent Architecture & Specifications](#6-agent-architecture--specifications)
7. [Control Plane Architecture & Specifications](#7-control-plane-architecture--specifications)
8. [REST API Specifications](#8-rest-api-specifications)
9. [CLI Specifications](#9-cli-specifications)
10. [Web Dashboard Specifications](#10-web-dashboard-specifications)
11. [Communication Protocols](#11-communication-protocols)
12. [Security Specifications](#12-security-specifications)
13. [Compiler Service Specifications](#13-compiler-service-specifications)
14. [Docker & Deployment Configuration](#14-docker--deployment-configuration)
15. [Testing Plan](#15-testing-plan)
16. [Performance Targets](#16-performance-targets)
17. [Development Phases](#17-development-phases)

---

## 1. Product Overview

### One-Line Description

FleetML is an open-source, chip-neutral edge MLOps platform that deploys, updates, and monitors ML models across heterogeneous edge device fleets from a single command.

### Core Value Proposition

- **One Command Deployment**: `fleetml deploy model.onnx --fleet production` deploys to hundreds of devices running different chips
- **Chip-Neutral**: NVIDIA (TensorRT), Intel (OpenVINO), ARM (TFLite), Qualcomm (SNPE), Hailo
- **Offline-First**: Edge devices operate independently when disconnected
- **Open Source**: Apache 2.0 license

### Architecture Summary

```
┌─────────────────────────────────────────────────────┐
│              FLEETML CONTROL PLANE                    │
│           (Cloud-hosted or Self-hosted)               │
│                                                       │
│  Model Registry │ Fleet Manager │ Multi-Chip Compiler │
│  Monitor/Alerts │ Policy Engine │ Data Pipeline        │
│  REST API + gRPC + Web Dashboard + CLI               │
└───────────────────────┬───────────────────────────────┘
                        │ gRPC / MQTT (encrypted)
        ┌───────────────┼───────────────┐
        ▼               ▼               ▼
  ┌──────────┐   ┌──────────┐   ┌──────────┐
  │ FleetML  │   │ FleetML  │   │ FleetML  │
  │ Agent    │   │ Agent    │   │ Agent    │
  │ ~50MB    │   │ ~50MB    │   │ ~30MB    │
  │ Jetson   │   │ Intel    │   │ RPi 5    │
  └──────────┘   └──────────┘   └──────────┘
```

---

## 2. Monorepo Structure

```
fleetml/
├── .github/
│   └── workflows/
│       └── ci.yml                    # GitHub Actions CI/CD
├── agent/                            # FleetML Edge Agent (Go)
│   ├── cmd/
│   │   └── agent/
│   │       └── main.go               # Agent entrypoint
│   ├── internal/
│   │   ├── communication/
│   │   │   ├── grpc_client.go         # gRPC client to control plane
│   │   │   ├── mqtt_client.go         # MQTT fallback client
│   │   │   └── store_forward.go       # Offline message queue
│   │   ├── model/
│   │   │   ├── loader.go              # Model loading & validation
│   │   │   ├── loader_test.go
│   │   │   ├── runtime.go             # Runtime selection (TRT/OV/TFLite/SNPE/ONNX)
│   │   │   ├── runtime_test.go
│   │   │   ├── hotswap.go             # Zero-downtime model swap
│   │   │   └── hotswap_test.go
│   │   ├── deploy/
│   │   │   ├── manager.go             # Deployment execution on agent side
│   │   │   ├── rollback.go            # Rollback manager (keep last N versions)
│   │   │   ├── rollback_test.go
│   │   │   └── abtesting.go           # A/B traffic splitting
│   │   ├── health/
│   │   │   ├── reporter.go            # System metrics collection
│   │   │   ├── reporter_test.go
│   │   │   └── gpu_detect.go          # GPU type detection
│   │   ├── drift/
│   │   │   ├── detector.go            # PSI/KS drift detection
│   │   │   ├── detector_test.go
│   │   │   └── sampler.go             # Edge data sampling for retraining
│   │   ├── heartbeat/
│   │   │   ├── protocol.go            # Heartbeat serialization & scheduling
│   │   │   └── protocol_test.go
│   │   ├── config/
│   │   │   └── config.go              # Agent configuration
│   │   ├── device/
│   │   │   └── fingerprint.go         # Hardware detection (arch, GPU, RAM, disk)
│   │   └── offline/
│   │       ├── sqlite_store.go        # Local SQLite for metrics
│   │       └── nats_queue.go          # Local NATS for command queue
│   ├── pkg/
│   │   └── version/
│   │       └── version.go
│   ├── go.mod
│   └── go.sum
├── server/                            # FleetML Control Plane (Go)
│   ├── cmd/
│   │   └── server/
│   │       └── main.go                # Server entrypoint
│   ├── internal/
│   │   ├── api/
│   │   │   ├── rest/
│   │   │   │   ├── router.go          # HTTP router setup
│   │   │   │   ├── middleware.go       # Auth, logging, CORS
│   │   │   │   ├── handlers/
│   │   │   │   │   ├── models.go      # Model CRUD handlers
│   │   │   │   │   ├── devices.go     # Device CRUD handlers
│   │   │   │   │   ├── deployments.go # Deployment handlers
│   │   │   │   │   ├── fleets.go      # Fleet/group handlers
│   │   │   │   │   ├── policies.go    # Policy CRUD handlers
│   │   │   │   │   ├── abtests.go     # A/B test handlers
│   │   │   │   │   ├── auth.go        # Authentication handlers
│   │   │   │   │   └── health.go      # Health check handler
│   │   │   │   └── dto/
│   │   │   │       ├── requests.go    # Request DTOs
│   │   │   │       └── responses.go   # Response DTOs
│   │   │   └── grpc/
│   │   │       ├── server.go          # gRPC server setup
│   │   │       └── handlers.go        # gRPC service implementations
│   │   ├── fleet/
│   │   │   ├── manager.go             # Fleet/device management
│   │   │   └── manager_test.go
│   │   ├── model/
│   │   │   ├── registry.go            # Model registry (upload, version, metadata)
│   │   │   └── registry_test.go
│   │   ├── deploy/
│   │   │   ├── orchestrator.go        # Deployment orchestration
│   │   │   ├── orchestrator_test.go
│   │   │   └── progress.go            # Deployment progress tracking
│   │   ├── policy/
│   │   │   ├── engine.go              # Policy evaluation engine
│   │   │   ├── engine_test.go
│   │   │   └── parser.go              # YAML policy parser
│   │   ├── monitor/
│   │   │   ├── alerts.go              # Alert evaluation & dispatch
│   │   │   └── metrics.go             # Metrics aggregation
│   │   ├── auth/
│   │   │   ├── jwt.go                 # JWT token generation/validation
│   │   │   ├── rbac.go                # Role-based access control
│   │   │   └── mtls.go                # Mutual TLS certificate management
│   │   ├── storage/
│   │   │   ├── s3.go                  # S3/MinIO client for model artifacts
│   │   │   └── postgres.go            # PostgreSQL connection & migrations
│   │   └── config/
│   │       └── config.go              # Server configuration
│   ├── migrations/
│   │   ├── 001_initial.up.sql
│   │   ├── 001_initial.down.sql
│   │   ├── 002_policies.up.sql
│   │   └── 002_policies.down.sql
│   ├── go.mod
│   └── go.sum
├── cli/                               # FleetML CLI (Go)
│   ├── cmd/
│   │   ├── root.go                    # Root command
│   │   ├── init.go                    # fleetml init
│   │   ├── deploy.go                  # fleetml deploy
│   │   ├── status.go                  # fleetml status
│   │   ├── logs.go                    # fleetml logs
│   │   ├── rollback.go                # fleetml rollback
│   │   ├── abtest.go                  # fleetml ab-test
│   │   └── version.go                 # fleetml version
│   ├── internal/
│   │   ├── client/
│   │   │   └── api_client.go          # HTTP client for control plane
│   │   └── output/
│   │       └── formatter.go           # Table/JSON output formatting
│   ├── go.mod
│   └── go.sum
├── dashboard/                         # Web Dashboard (React + TypeScript)
│   ├── src/
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   ├── api/
│   │   │   └── client.ts              # API client (fetch wrapper)
│   │   ├── components/
│   │   │   ├── Layout.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   ├── DeviceCard.tsx
│   │   │   ├── ModelCard.tsx
│   │   │   ├── DeploymentProgress.tsx
│   │   │   ├── MetricsChart.tsx
│   │   │   └── FleetMap.tsx
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx           # Overview page
│   │   │   ├── Devices.tsx             # Device list & detail
│   │   │   ├── Models.tsx              # Model registry
│   │   │   ├── Deployments.tsx         # Deployment history
│   │   │   ├── ABTests.tsx             # A/B test management
│   │   │   ├── Policies.tsx            # Policy management
│   │   │   └── Settings.tsx            # System settings
│   │   ├── hooks/
│   │   │   ├── useDevices.ts
│   │   │   ├── useModels.ts
│   │   │   └── useDeployments.ts
│   │   └── types/
│   │       └── index.ts                # TypeScript type definitions
│   ├── package.json
│   ├── tsconfig.json
│   ├── tailwind.config.js
│   └── vite.config.ts
├── compiler/                          # Multi-Chip Compiler Service (Python)
│   ├── main.py                        # FastAPI server
│   ├── compilers/
│   │   ├── base.py                    # Abstract compiler interface
│   │   ├── tensorrt.py                # NVIDIA TensorRT compiler
│   │   ├── openvino.py                # Intel OpenVINO compiler
│   │   ├── tflite.py                  # TensorFlow Lite compiler
│   │   ├── snpe.py                    # Qualcomm SNPE compiler
│   │   └── hailo.py                   # Hailo compiler
│   ├── Dockerfile.tensorrt
│   ├── Dockerfile.openvino
│   ├── Dockerfile.tflite
│   ├── requirements.txt
│   └── docker-compose.compiler.yml
├── proto/                             # Protocol Buffers definitions
│   ├── fleetml/v1/
│   │   ├── agent.proto                # Agent ↔ Control Plane service
│   │   ├── models.proto               # Model-related messages
│   │   ├── devices.proto              # Device-related messages
│   │   ├── deployments.proto          # Deployment-related messages
│   │   └── common.proto               # Shared message types
│   └── buf.yaml
├── simulator/                         # Virtual Fleet Simulator
│   ├── profiles.go                    # Device profiles (Jetson, RPi, NUC, etc.)
│   ├── network.go                     # Network condition simulation
│   ├── fleet.go                       # Fleet creation & management
│   ├── Dockerfile.virtual-device
│   └── docker-compose.fleet-sim.yml
├── tests/                             # Test suites
│   ├── integration/
│   │   ├── registration_test.go
│   │   ├── deploy_test.go
│   │   └── rollback_test.go
│   ├── fleet/
│   │   ├── heterogeneous_test.go
│   │   ├── offline_test.go
│   │   └── canary_test.go
│   ├── chaos/
│   │   ├── network_chaos_test.go
│   │   ├── device_chaos_test.go
│   │   └── server_chaos_test.go
│   ├── scale/
│   │   └── thousand_devices_test.go
│   ├── security/
│   │   ├── auth_test.go
│   │   └── integrity_test.go
│   ├── ml/
│   │   ├── model_formats_test.go
│   │   └── drift_test.go
│   ├── hardware/
│   │   └── real_deploy_test.go
│   ├── e2e/
│   │   ├── first_time_user.sh
│   │   ├── multi_device_fleet.sh
│   │   └── offline_resilience.sh
│   └── load/
│       └── k6-heartbeat.js
├── docs/                              # Documentation (Docusaurus or MkDocs)
│   ├── quickstart.md
│   ├── installation.md
│   ├── architecture.md
│   ├── cli-reference.md
│   └── api-reference.md
├── docker-compose.yml                 # Production self-hosted setup
├── docker-compose.test.yml            # Test environment
├── docker-compose.fleet-sim.yml       # Fleet simulator
├── Makefile                           # Build commands
├── README.md
├── LICENSE                            # Apache 2.0
├── CONTRIBUTING.md
├── CHANGELOG.md
└── CODE_OF_CONDUCT.md
```

---

## 3. Technology Stack

| Component | Technology | Version | Rationale |
|-----------|-----------|---------|-----------|
| **Agent** | Go | 1.22+ | Small binary (~10MB), cross-compilation, fast startup, low memory |
| **Control Plane API** | Go + gRPC | 1.22+ | Performance, type safety, native gRPC support |
| **REST Framework** | `net/http` + `chi` router | v5 | Lightweight, idiomatic Go |
| **gRPC** | `google.golang.org/grpc` | latest | Agent ↔ server communication |
| **Web Dashboard** | React + TypeScript + Tailwind CSS | React 18, TS 5, Tailwind 3 | Modern, component-based |
| **Dashboard Build** | Vite | 5.x | Fast build, HMR |
| **CLI** | Go + `cobra` | cobra v1.8+ | Industry standard CLI framework |
| **Database** | PostgreSQL | 16 | Reliable, well-understood |
| **DB Migrations** | `golang-migrate/migrate` | v4 | SQL migration files |
| **Model Storage** | S3-compatible (MinIO for self-host) | MinIO latest | Model artifacts (ONNX, TensorRT, etc.) |
| **Message Bus** | NATS | 2.10+ | Lightweight pub/sub, edge-friendly |
| **MQTT** | Eclipse Mosquitto | 2.0+ | Fallback for constrained devices |
| **Compiler Service** | Python + FastAPI | Python 3.11, FastAPI 0.100+ | Each compiler runs in isolated Docker container |
| **Monitoring** | Prometheus + Grafana | Latest | Industry standard |
| **Container Runtime** | Docker + Docker Compose | Latest | Self-hosted deployment |
| **Agent Local DB** | SQLite | 3.x | Offline metrics storage on device |
| **Protobuf** | Protocol Buffers v3 | proto3 | gRPC message definitions |
| **TLS** | TLS 1.3 | — | All communication encrypted |
| **Auth** | JWT + mTLS | — | API auth + agent auth |

### Go Module Dependencies (Agent)

```go
// agent/go.mod
module github.com/fleetml/fleetml/agent

go 1.22

require (
    google.golang.org/grpc v1.62.0
    google.golang.org/protobuf v1.33.0
    github.com/eclipse/paho.mqtt.golang v1.4.3
    github.com/mattn/go-sqlite3 v1.14.22
    github.com/nats-io/nats.go v1.33.0
    github.com/shirou/gopsutil/v3 v3.24.1
    github.com/NVIDIA/go-nvml v0.12.0-1
    go.uber.org/zap v1.27.0
    gopkg.in/yaml.v3 v3.0.1
)
```

### Go Module Dependencies (Server)

```go
// server/go.mod
module github.com/fleetml/fleetml/server

go 1.22

require (
    github.com/go-chi/chi/v5 v5.0.12
    github.com/go-chi/cors v1.2.1
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/golang-migrate/migrate/v4 v4.17.0
    github.com/jackc/pgx/v5 v5.5.3
    github.com/minio/minio-go/v7 v7.0.67
    google.golang.org/grpc v1.62.0
    google.golang.org/protobuf v1.33.0
    github.com/prometheus/client_golang v1.19.0
    go.uber.org/zap v1.27.0
    gopkg.in/yaml.v3 v3.0.1
    golang.org/x/crypto v0.20.0
)
```

---

## 4. Database Schema

### PostgreSQL Database: `fleetml`

#### Table: `devices`

```sql
CREATE TABLE devices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id       VARCHAR(255) NOT NULL UNIQUE,  -- Human-readable ID (e.g., "edge-042")
    name            VARCHAR(255),
    status          VARCHAR(50) NOT NULL DEFAULT 'registered', -- registered, healthy, warning, offline, decommissioned
    arch            VARCHAR(50) NOT NULL,           -- arm64, amd64
    gpu_type        VARCHAR(50),                    -- nvidia, intel, none
    runtime         VARCHAR(50),                    -- tensorrt, openvino, tflite, snpe, onnx
    ram_mb          INTEGER NOT NULL,
    disk_gb         INTEGER NOT NULL,
    os              VARCHAR(100),
    hardware_model  VARCHAR(255),                   -- "NVIDIA Jetson Orin Nano", "Raspberry Pi 5"

    -- Metadata
    labels          JSONB DEFAULT '{}',             -- {"region": "eu", "environment": "production"}
    fleet_id        UUID REFERENCES fleets(id),

    -- TLS/Auth
    certificate_fingerprint VARCHAR(255),

    -- Timestamps
    last_heartbeat  TIMESTAMP WITH TIME ZONE,
    registered_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- System metrics (latest)
    cpu_percent     REAL,
    gpu_percent     REAL,
    ram_mb_used     INTEGER,
    disk_percent    REAL,
    temperature_c   REAL,
    uptime_hours    REAL
);

CREATE INDEX idx_devices_status ON devices(status);
CREATE INDEX idx_devices_fleet_id ON devices(fleet_id);
CREATE INDEX idx_devices_labels ON devices USING GIN(labels);
CREATE INDEX idx_devices_last_heartbeat ON devices(last_heartbeat);
```

#### Table: `fleets`

```sql
CREATE TABLE fleets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL UNIQUE,      -- "production", "staging", "eu-fleet"
    description TEXT,
    labels      JSONB DEFAULT '{}',
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

#### Table: `models`

```sql
CREATE TABLE models (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,          -- "face-detect"
    version         VARCHAR(100) NOT NULL,           -- "3.2"
    format          VARCHAR(50) NOT NULL,            -- onnx, tensorrt, openvino, tflite, snpe

    -- Storage
    artifact_url    TEXT NOT NULL,                   -- S3 URL: "s3://fleetml-models/face-detect/3.2/model.onnx"
    artifact_size   BIGINT NOT NULL,                 -- Bytes
    checksum        VARCHAR(255) NOT NULL,           -- "sha256:abc123..."

    -- Metadata
    description     TEXT,
    metadata        JSONB DEFAULT '{}',              -- {"framework": "pytorch", "task": "detection", "input_shape": [1,3,640,640]}
    tags            TEXT[],                           -- ["production", "yolov8"]

    -- Lineage
    parent_model_id UUID REFERENCES models(id),      -- For tracking model derivation

    -- Compiled variants
    compiled_variants JSONB DEFAULT '[]',            -- [{"runtime": "tensorrt", "artifact_url": "...", "checksum": "..."}]

    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by      UUID REFERENCES users(id),

    UNIQUE(name, version)
);

CREATE INDEX idx_models_name ON models(name);
CREATE INDEX idx_models_name_version ON models(name, version);
```

#### Table: `deployments`

```sql
CREATE TABLE deployments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id            UUID NOT NULL REFERENCES models(id),

    -- Target
    target_type         VARCHAR(50) NOT NULL,        -- "fleet", "device", "label_selector"
    target_fleet_id     UUID REFERENCES fleets(id),
    target_device_ids   UUID[],
    target_labels       JSONB,                        -- {"region": "eu"}

    -- Status
    state               VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, rolling_out, completed, failed, rolled_back, cancelled
    total_devices       INTEGER NOT NULL DEFAULT 0,
    completed_devices   INTEGER NOT NULL DEFAULT 0,
    failed_devices      INTEGER NOT NULL DEFAULT 0,
    queued_devices      INTEGER NOT NULL DEFAULT 0,   -- Offline devices awaiting deploy

    -- Policy
    deployment_policy   VARCHAR(100) DEFAULT 'immediate', -- immediate, canary, scheduled
    canary_config       JSONB,                        -- {"stages": [{"percent": 5, "duration": "1h", "success_metric": "accuracy > 0.90"}]}
    rollback_model_id   UUID REFERENCES models(id),

    -- Metadata
    error               TEXT,
    started_at          TIMESTAMP WITH TIME ZONE,
    completed_at        TIMESTAMP WITH TIME ZONE,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by          UUID REFERENCES users(id)
);

CREATE INDEX idx_deployments_state ON deployments(state);
CREATE INDEX idx_deployments_model_id ON deployments(model_id);
CREATE INDEX idx_deployments_created_at ON deployments(created_at DESC);
```

#### Table: `deployment_device_status`

```sql
CREATE TABLE deployment_device_status (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id   UUID NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    device_id       UUID NOT NULL REFERENCES devices(id),

    state           VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, downloading, applying, completed, failed, queued
    error           TEXT,
    started_at      TIMESTAMP WITH TIME ZONE,
    completed_at    TIMESTAMP WITH TIME ZONE,

    UNIQUE(deployment_id, device_id)
);

CREATE INDEX idx_dds_deployment_id ON deployment_device_status(deployment_id);
CREATE INDEX idx_dds_device_id ON deployment_device_status(device_id);
```

#### Table: `device_models`

```sql
CREATE TABLE device_models (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id       UUID NOT NULL REFERENCES devices(id),
    model_id        UUID NOT NULL REFERENCES models(id),

    status          VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive, rollback_target
    runtime         VARCHAR(50) NOT NULL,                   -- tensorrt, openvino, tflite, onnx

    -- Inference metrics (latest)
    inference_count BIGINT DEFAULT 0,
    avg_latency_ms  REAL,
    p99_latency_ms  REAL,
    accuracy        REAL,
    drift_score     REAL,

    deployed_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(device_id, model_id)
);
```

#### Table: `ab_tests`

```sql
CREATE TABLE ab_tests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255),

    model_a_id      UUID NOT NULL REFERENCES models(id),
    model_b_id      UUID NOT NULL REFERENCES models(id),

    split_a         INTEGER NOT NULL DEFAULT 80,     -- Percentage for model A
    split_b         INTEGER NOT NULL DEFAULT 20,     -- Percentage for model B

    target_fleet_id UUID REFERENCES fleets(id),
    target_labels   JSONB,

    metrics         TEXT[] NOT NULL,                  -- ["accuracy", "latency"]
    duration        INTERVAL NOT NULL,                -- "7 days"
    auto_promote    BOOLEAN DEFAULT false,

    state           VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, running, completed, cancelled
    winner          VARCHAR(10),                     -- "a", "b", null
    results         JSONB,                            -- Aggregated results

    started_at      TIMESTAMP WITH TIME ZONE,
    ends_at         TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

#### Table: `policies`

```sql
CREATE TABLE policies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    description TEXT,

    -- Matching
    match_labels    JSONB,                           -- {"region": "eu"}
    match_hardware  JSONB,                           -- {"ram_mb": {"lt": 4096}}

    -- Rules
    rules       JSONB NOT NULL,                      -- Policy rules (see Policy Engine YAML spec)

    -- Deployment rules
    deploy_config JSONB,                             -- Canary stages, etc.

    -- Rollback triggers
    rollback_config JSONB,                           -- {"trigger": [{"metric": "accuracy", "op": "<", "value": 0.85}]}

    enabled     BOOLEAN DEFAULT true,
    priority    INTEGER DEFAULT 0,                   -- Higher = evaluated first

    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

#### Table: `users`

```sql
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) NOT NULL UNIQUE,
    name        VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    role        VARCHAR(50) NOT NULL DEFAULT 'viewer', -- admin, deployer, viewer
    api_key     VARCHAR(255) UNIQUE,

    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

#### Table: `heartbeats`

```sql
CREATE TABLE heartbeats (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id   UUID NOT NULL REFERENCES devices(id),

    -- System metrics
    cpu_percent     REAL,
    gpu_percent     REAL,
    ram_mb_used     INTEGER,
    disk_percent    REAL,
    temperature_c   REAL,
    uptime_hours    REAL,

    -- Model metrics (JSONB array)
    model_metrics   JSONB,                           -- [{"name": "face-detect", "version": "3.2", "inference_count": 12847, ...}]

    received_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Partition by time for efficient storage
-- Consider TimescaleDB extension for production
CREATE INDEX idx_heartbeats_device_id ON heartbeats(device_id);
CREATE INDEX idx_heartbeats_received_at ON heartbeats(received_at DESC);
```

#### Table: `audit_log`

```sql
CREATE TABLE audit_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action      VARCHAR(100) NOT NULL,               -- "model.deployed", "device.registered", "rollback.triggered"
    actor_id    UUID,                                 -- User or system
    actor_type  VARCHAR(50) NOT NULL,                 -- "user", "system", "policy"
    resource_type VARCHAR(50) NOT NULL,               -- "model", "device", "deployment"
    resource_id UUID,
    details     JSONB,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_action ON audit_log(action);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at DESC);
```

---

## 5. Protobuf & gRPC Service Definitions

### `proto/fleetml/v1/common.proto`

```protobuf
syntax = "proto3";
package fleetml.v1;
option go_package = "github.com/fleetml/fleetml/proto/fleetml/v1";

import "google/protobuf/timestamp.proto";

message Empty {}

message Error {
    string code = 1;
    string message = 2;
}
```

### `proto/fleetml/v1/devices.proto`

```protobuf
syntax = "proto3";
package fleetml.v1;
option go_package = "github.com/fleetml/fleetml/proto/fleetml/v1";

import "google/protobuf/timestamp.proto";

message DeviceInfo {
    string device_id = 1;
    string arch = 2;              // arm64, amd64
    string gpu_type = 3;          // nvidia, intel, none
    string runtime = 4;           // tensorrt, openvino, tflite, snpe, onnx
    int32 ram_mb = 5;
    int32 disk_gb = 6;
    string os = 7;
    string hardware_model = 8;    // "NVIDIA Jetson Orin Nano"
    map<string, string> labels = 9;
}

message SystemMetrics {
    float cpu_percent = 1;
    float gpu_percent = 2;
    int32 ram_mb_used = 3;
    float disk_percent = 4;
    float temperature_c = 5;
    float uptime_hours = 6;
}

message ModelMetrics {
    string name = 1;
    string version = 2;
    string runtime = 3;
    int64 inference_count = 4;
    float avg_latency_ms = 5;
    float p99_latency_ms = 6;
    float accuracy = 7;
    float drift_score = 8;
}
```

### `proto/fleetml/v1/agent.proto`

```protobuf
syntax = "proto3";
package fleetml.v1;
option go_package = "github.com/fleetml/fleetml/proto/fleetml/v1";

import "google/protobuf/timestamp.proto";
import "fleetml/v1/devices.proto";

// Agent ↔ Control Plane service
service AgentService {
    // Agent registration
    rpc Register(RegisterRequest) returns (RegisterResponse);

    // Heartbeat stream (bidirectional)
    rpc Heartbeat(stream HeartbeatRequest) returns (stream HeartbeatResponse);

    // Model download
    rpc GetModelArtifact(GetModelArtifactRequest) returns (stream ModelArtifactChunk);

    // Report deployment status
    rpc ReportDeploymentStatus(DeploymentStatusReport) returns (Empty);

    // Bulk sync (after offline period)
    rpc BulkSyncMetrics(stream BulkMetricsRequest) returns (BulkSyncResponse);
}

message RegisterRequest {
    DeviceInfo device_info = 1;
}

message RegisterResponse {
    string agent_id = 1;           // Server-assigned UUID
    int32 heartbeat_interval_sec = 2; // Default: 30
    string certificate = 3;        // mTLS certificate for this agent
}

message HeartbeatRequest {
    string device_id = 1;
    google.protobuf.Timestamp timestamp = 2;
    string status = 3;             // healthy, warning, error
    repeated ModelMetrics models = 4;
    SystemMetrics system = 5;
}

message HeartbeatResponse {
    repeated Command commands = 1;  // Commands to execute
}

message Command {
    string id = 1;
    string type = 2;               // deploy_model, rollback, update_config, restart, set_ab_test
    bytes payload = 3;             // JSON-encoded command payload
    google.protobuf.Timestamp issued_at = 4;
}

// Deploy model command payload (JSON in Command.payload)
message DeployModelPayload {
    string model_name = 1;
    string model_version = 2;
    string runtime = 3;            // tensorrt, openvino, tflite, snpe, onnx
    string artifact_url = 4;       // S3 URL
    string checksum = 5;           // sha256:...
    string rollback_version = 6;   // Version to rollback to on failure
    string deployment_policy = 7;  // immediate, canary
}

// A/B test command payload
message SetABTestPayload {
    string model_a_name = 1;
    string model_a_version = 2;
    string model_b_name = 3;
    string model_b_version = 4;
    int32 split_a_percent = 5;     // e.g., 80
    int32 split_b_percent = 6;     // e.g., 20
    string test_id = 7;
}

message GetModelArtifactRequest {
    string model_name = 1;
    string model_version = 2;
    string runtime = 3;
}

message ModelArtifactChunk {
    bytes data = 1;
    int64 total_size = 2;
    string checksum = 3;
}

message DeploymentStatusReport {
    string device_id = 1;
    string deployment_id = 2;
    string state = 3;              // downloading, applying, completed, failed
    string error = 4;
    google.protobuf.Timestamp timestamp = 5;
}

message BulkMetricsRequest {
    string device_id = 1;
    repeated HeartbeatRequest heartbeats = 2; // Buffered heartbeats
}

message BulkSyncResponse {
    repeated Command pending_commands = 1; // Commands queued while offline
}
```

---

## 6. Agent Architecture & Specifications

### Agent Binary

- **Language**: Go 1.22+
- **Target Binary Size**: <15MB (stripped)
- **Target Memory**: <50MB RSS at runtime
- **Cross-compilation**: `GOOS=linux GOARCH=amd64` and `GOOS=linux GOARCH=arm64`
- **Entrypoint**: `agent/cmd/agent/main.go`

### Agent Startup Flow

```
1. Load configuration (file + env vars + CLI flags)
2. Detect hardware (arch, GPU, RAM, disk, OS)
3. Initialize local SQLite store
4. Initialize local NATS queue (offline command buffer)
5. Connect to control plane via gRPC (with mTLS)
6. Register device (if first boot)
7. Start heartbeat loop (every 30s, configurable)
8. Start health reporter (system metrics collection)
9. Start drift detector (if model loaded)
10. Load active model (if any) into runtime
11. Listen for commands from control plane
12. Enter main event loop
```

### Agent Configuration (`fleetml-agent.yaml`)

```yaml
server:
  address: "control-plane.example.com:50051"
  tls:
    enabled: true
    cert_file: "/etc/fleetml/agent.crt"
    key_file: "/etc/fleetml/agent.key"
    ca_file: "/etc/fleetml/ca.crt"

device:
  id: "edge-042"                   # Auto-generated if omitted
  labels:
    region: "eu"
    environment: "production"
    building: "warehouse-3"

heartbeat:
  interval: "30s"                  # Heartbeat interval
  timeout: "10s"                   # gRPC timeout per heartbeat

model:
  storage_dir: "/var/fleetml/models"
  max_versions: 3                  # Keep last N versions for rollback
  runtime: "auto"                  # auto-detect based on hardware

offline:
  metrics_db: "/var/fleetml/metrics.db"      # SQLite path
  command_queue: "/var/fleetml/commands"      # NATS jetstream path
  max_buffer_size_mb: 100                     # Max offline buffer

drift:
  enabled: true
  method: "psi"                    # psi or ks
  threshold: 0.2
  window_size: 1000
  check_interval: "5m"

logging:
  level: "info"                    # debug, info, warn, error
  file: "/var/log/fleetml/agent.log"
```

### Agent Components — Detailed Specifications

#### 6.1 Communication Layer (`agent/internal/communication/`)

**gRPC Client** (`grpc_client.go`):
- Connect to control plane with mTLS
- Handle connection retry with exponential backoff (1s, 2s, 4s, 8s, max 60s)
- Bidirectional streaming for heartbeat
- Detect connection loss within 10 seconds
- Switch to offline mode on disconnect

**MQTT Fallback** (`mqtt_client.go`):
- Used when gRPC fails or device is too constrained
- Topics: `fleetml/{device_id}/heartbeat`, `fleetml/{device_id}/commands`, `fleetml/{device_id}/status`
- QoS 1 (at least once delivery)

**Store-and-Forward** (`store_forward.go`):
- When offline, queue heartbeats to local SQLite
- Queue commands from local NATS
- On reconnect, bulk-sync heartbeats to control plane
- Download and apply any queued commands

#### 6.2 Model Manager (`agent/internal/model/`)

**Model Loader** (`loader.go`):

```go
type ModelLoader struct {
    storageDir string
    maxVersions int
}

// Load validates and loads a model file
func (l *ModelLoader) Load(filename string) (*Model, error)

// ValidateChecksum verifies SHA-256 checksum
func (l *ModelLoader) ValidateChecksum(filename string, expectedChecksum string) error

// Download downloads model from URL with progress reporting
func (l *ModelLoader) Download(url string, checksum string) (string, error)

type Model struct {
    Name       string
    Version    string
    Format     string    // onnx, tensorrt, openvino, tflite, snpe
    FilePath   string
    SizeBytes  int64
    Checksum   string
    LoadedAt   time.Time
}
```

**Runtime Selection** (`runtime.go`):

```go
type Runtime interface {
    Name() string
    Load(modelPath string) error
    Infer(input []byte) ([]byte, error)
    Unload() error
    IsSupported() bool
}

// Implementations:
// - ONNXRuntime (default, all platforms)
// - TensorRTRuntime (NVIDIA GPUs)
// - OpenVINORuntime (Intel CPUs/GPUs)
// - TFLiteRuntime (ARM devices)
// - SNPERuntime (Qualcomm)

func DetectBestRuntime(deviceInfo DeviceInfo) Runtime
```

**Hot-Swap** (`hotswap.go`):
- Load new model while old model continues serving
- Atomic pointer swap (zero dropped inferences)
- Unload old model after swap

```go
type HotSwapper struct {
    activeModel  atomic.Pointer[LoadedModel]
    rollbackModel *LoadedModel
}

// Swap atomically switches the active model
func (h *HotSwapper) Swap(newModel *LoadedModel) error

// Rollback reverts to the previous model
func (h *HotSwapper) Rollback() error
```

#### 6.3 Health Reporter (`agent/internal/health/`)

```go
type HealthReporter struct {
    interval time.Duration
}

type SystemMetrics struct {
    CPUPercent    float64
    GPUPercent    float64   // 0 if no GPU
    RAMMBUsed     int
    DiskPercent   float64
    TemperatureC  float64
    UptimeHours   float64
}

// Collect gathers current system metrics
func (h *HealthReporter) Collect() (*SystemMetrics, error)

// GPU Detection:
// - NVIDIA: Use go-nvml library
// - Intel: Read /sys/class/drm/card0/...
// - None: Report 0
```

#### 6.4 Drift Detector (`agent/internal/drift/`)

```go
type DriftDetector struct {
    method     string     // "psi" or "ks"
    threshold  float64
    windowSize int
    baseline   []float64
    current    []float64
}

// AddBaseline adds a confidence score to baseline distribution
func (d *DriftDetector) AddBaseline(score float64)

// AddObservation adds a confidence score to current observation window
func (d *DriftDetector) AddObservation(score float64)

// IsDrifting returns true if drift exceeds threshold
func (d *DriftDetector) IsDrifting() bool

// DriftScore returns the current drift metric value
func (d *DriftDetector) DriftScore() float64

// PSI (Population Stability Index):
// PSI = Σ (actual_i - expected_i) × ln(actual_i / expected_i)
// Threshold: PSI > 0.2 = significant drift

// KS (Kolmogorov-Smirnov) Test:
// D = max|F_baseline(x) - F_current(x)|
// Threshold: D > critical value for given sample size
```

#### 6.5 Rollback Manager (`agent/internal/deploy/`)

```go
type RollbackManager struct {
    storageDir  string
    maxVersions int  // Keep last N model versions
}

// SaveVersion stores a model version for potential rollback
func (r *RollbackManager) SaveVersion(version string, modelBytes []byte) error

// HasVersion checks if a version is available for rollback
func (r *RollbackManager) HasVersion(version string) bool

// Restore restores a previous model version
func (r *RollbackManager) Restore(version string) ([]byte, error)

// Behavior:
// - Stores last N model files (default: 3)
// - When N+1 version arrives, evicts oldest
// - On failed deployment, auto-rollback to previous active version
// - On power failure during OTA, boots into last known good model
```

---

## 7. Control Plane Architecture & Specifications

### Server Startup Flow

```
1. Load configuration (file + env vars)
2. Connect to PostgreSQL, run migrations
3. Connect to S3/MinIO
4. Initialize gRPC server (port 50051)
5. Initialize REST API server (port 8080)
6. Start deployment orchestrator worker
7. Start heartbeat processor worker
8. Start alert evaluator worker
9. Start Prometheus metrics endpoint (:9090)
10. Ready to accept connections
```

### Server Configuration (`fleetml-server.yaml`)

```yaml
server:
  rest_port: 8080
  grpc_port: 50051
  tls:
    enabled: true
    cert_file: "/etc/fleetml/server.crt"
    key_file: "/etc/fleetml/server.key"
    ca_file: "/etc/fleetml/ca.crt"

database:
  url: "postgres://fleetml:password@localhost:5432/fleetml?sslmode=require"
  max_connections: 50

storage:
  type: "s3"                       # s3 or minio
  endpoint: "http://minio:9000"
  bucket: "fleetml-models"
  access_key: "minioadmin"
  secret_key: "minioadmin"
  region: "us-east-1"

auth:
  jwt_secret: "${JWT_SECRET}"
  jwt_expiry: "24h"

heartbeat:
  offline_threshold: "90s"         # Mark device offline after 3 missed heartbeats

deployment:
  default_timeout: "10m"
  concurrent_deploys_per_device: 1

logging:
  level: "info"
```

### Control Plane Components — Detailed Specifications

#### 7.1 Fleet Manager (`server/internal/fleet/`)

```go
type FleetManager struct {
    db *pgxpool.Pool
}

// RegisterDevice registers a new device
func (f *FleetManager) RegisterDevice(ctx context.Context, info DeviceInfo) (*Device, error)

// GetDevice returns device by ID
func (f *FleetManager) GetDevice(ctx context.Context, deviceID string) (*Device, error)

// ListDevices lists devices with optional filters
func (f *FleetManager) ListDevices(ctx context.Context, filter DeviceFilter) ([]*Device, error)

// UpdateDeviceStatus updates device status from heartbeat
func (f *FleetManager) UpdateDeviceStatus(ctx context.Context, deviceID string, status DeviceStatus) error

// SelectDevices selects devices matching a target (fleet, labels, device IDs)
func (f *FleetManager) SelectDevices(ctx context.Context, target DeploymentTarget) ([]*Device, error)

// CreateFleet creates a device group
func (f *FleetManager) CreateFleet(ctx context.Context, name string, labels map[string]string) (*Fleet, error)

// AssignDeviceToFleet adds a device to a fleet
func (f *FleetManager) AssignDeviceToFleet(ctx context.Context, deviceID string, fleetID string) error

type DeviceFilter struct {
    Status   string
    FleetID  string
    Labels   map[string]string
    Runtime  string
    Limit    int
    Offset   int
}
```

#### 7.2 Model Registry (`server/internal/model/`)

```go
type Registry struct {
    db      *pgxpool.Pool
    storage S3Client
}

// Upload uploads a model file and creates a registry entry
func (r *Registry) Upload(ctx context.Context, req UploadModelRequest) (*Model, error)

// GetModel returns model by name and version
func (r *Registry) GetModel(ctx context.Context, name string, version string) (*Model, error)

// ListModels lists models with optional filters
func (r *Registry) ListModels(ctx context.Context, filter ModelFilter) ([]*Model, error)

// DeleteModel soft-deletes a model (cannot delete if actively deployed)
func (r *Registry) DeleteModel(ctx context.Context, id string) error

// GetArtifactURL generates a pre-signed URL for model download
func (r *Registry) GetArtifactURL(ctx context.Context, modelID string, runtime string) (string, error)

type UploadModelRequest struct {
    Name        string
    Version     string
    Format      string
    Data        io.Reader
    Description string
    Metadata    map[string]interface{}
    Tags        []string
}
```

#### 7.3 Deployment Orchestrator (`server/internal/deploy/`)

```go
type Orchestrator struct {
    db       *pgxpool.Pool
    fleet    *FleetManager
    registry *Registry
    policy   *PolicyEngine
}

// CreateDeployment creates and starts a deployment
func (o *Orchestrator) CreateDeployment(ctx context.Context, req CreateDeploymentRequest) (*Deployment, error)

// GetDeployment returns deployment status
func (o *Orchestrator) GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error)

// CancelDeployment cancels a running deployment
func (o *Orchestrator) CancelDeployment(ctx context.Context, deploymentID string) error

// HandleDeploymentReport processes agent deployment status reports
func (o *Orchestrator) HandleDeploymentReport(ctx context.Context, report DeploymentStatusReport) error

// Deployment Flow:
// 1. Resolve target devices (fleet → device list)
// 2. Check policy rules (canary, resource-aware, etc.)
// 3. For each device:
//    a. Determine correct runtime (based on device hardware)
//    b. Check if compiled variant exists, if not trigger compilation
//    c. Generate pre-signed download URL
//    d. Queue deploy command in heartbeat response
// 4. Track progress per device
// 5. Handle failures (retry, rollback based on policy)

type CreateDeploymentRequest struct {
    ModelName    string
    ModelVersion string
    TargetType   string             // "fleet", "device", "labels"
    TargetID     string             // Fleet ID or device ID
    TargetLabels map[string]string  // Label selector
    Policy       string             // "immediate", "canary"
    CanaryConfig *CanaryConfig
}

type CanaryConfig struct {
    Stages []CanaryStage
}

type CanaryStage struct {
    Percent       int
    Duration      time.Duration
    SuccessMetric string             // "accuracy > 0.90"
}
```

#### 7.4 Policy Engine (`server/internal/policy/`)

```go
type PolicyEngine struct {
    db *pgxpool.Pool
}

// Evaluate evaluates all policies for a device/deployment
func (p *PolicyEngine) Evaluate(ctx context.Context, device *Device, deployment *Deployment) (*PolicyResult, error)

// ShouldRollback checks if metrics trigger a rollback
func (p *PolicyEngine) ShouldRollback(ctx context.Context, metrics DeviceMetrics) bool

// ParseYAML parses a YAML policy definition
func (p *PolicyEngine) ParseYAML(yamlBytes []byte) (*Policy, error)

type Policy struct {
    Name          string
    Match         PolicyMatch
    Rules         []PolicyRule
    DeployConfig  *CanaryConfig
    RollbackConfig *RollbackConfig
}

type PolicyMatch struct {
    Labels   map[string]string
    Hardware *HardwareMatch
}

type HardwareMatch struct {
    RAMMB *ComparisonRule
}

type ComparisonRule struct {
    Op    string  // "<", ">", "<=", ">=", "=="
    Value float64
}

type RollbackConfig struct {
    Triggers []RollbackTrigger
}

type RollbackTrigger struct {
    Metric string  // "accuracy", "latency_p99_ms", "error_rate"
    Op     string  // "<", ">"
    Value  float64
}
```

### Policy YAML Schema

```yaml
# Full policy YAML specification
policies:
  - name: "string"                    # Required: Policy name
    description: "string"             # Optional: Description

    match:                            # Required: Which devices this applies to
      labels:                         # Match by labels
        key: "value"
      hardware:                       # Match by hardware constraints
        ram_mb: "<4096"              # Operators: <, >, <=, >=, ==
        gpu_type: "nvidia"

    rules:                            # Optional: Deployment rules
      - model_version: ">=3.0"       # Minimum model version
      - model_format: "int8"         # Force quantized model
      - data_export: false           # Restrict data export

    deploy:                           # Optional: Deployment strategy
      stages:
        - percent: 5                 # Canary: deploy to 5% first
          duration: "1h"             # Wait 1 hour
          success_metric: "accuracy > 0.90"
        - percent: 50
          duration: "6h"
          success_metric: "accuracy > 0.92"
        - percent: 100

    rollback:                         # Optional: Auto-rollback triggers
      trigger:
        - metric: "latency_p99_ms"
          op: ">"
          value: 100
        - metric: "accuracy"
          op: "<"
          value: 0.85
        - metric: "error_rate"
          op: ">"
          value: 0.05
```

---

## 8. REST API Specifications

**Base URL**: `http://localhost:8080/api/v1`
**Authentication**: Bearer JWT token in `Authorization` header
**Content-Type**: `application/json`

### 8.1 Authentication

#### `POST /api/v1/auth/login`

Request:
```json
{
    "email": "admin@example.com",
    "password": "password"
}
```

Response `200`:
```json
{
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2026-02-23T10:30:00Z",
    "user": {
        "id": "uuid",
        "email": "admin@example.com",
        "name": "Admin",
        "role": "admin"
    }
}
```

#### `POST /api/v1/auth/register`

Request:
```json
{
    "email": "user@example.com",
    "password": "password",
    "name": "User Name"
}
```

#### `GET /api/v1/auth/me`

Returns current authenticated user info.

### 8.2 Models

#### `POST /api/v1/models` — Upload Model

Request: `multipart/form-data`
- `file`: Model file (ONNX, TFLite, etc.)
- `name`: string (required)
- `version`: string (required)
- `format`: string (required) — `onnx`, `tensorrt`, `openvino`, `tflite`, `snpe`
- `description`: string (optional)
- `metadata`: JSON string (optional)
- `tags`: comma-separated string (optional)

Response `201`:
```json
{
    "id": "uuid",
    "name": "face-detect",
    "version": "3.2",
    "format": "onnx",
    "artifact_url": "s3://fleetml-models/face-detect/3.2/model.onnx",
    "artifact_size": 12800000,
    "checksum": "sha256:abc123...",
    "description": "Face detection model",
    "metadata": {"framework": "pytorch", "task": "detection"},
    "tags": ["production", "yolov8"],
    "created_at": "2026-02-22T10:00:00Z"
}
```

#### `GET /api/v1/models` — List Models

Query params: `name`, `tags`, `limit` (default 50), `offset` (default 0)

Response `200`:
```json
{
    "models": [...],
    "total": 42,
    "limit": 50,
    "offset": 0
}
```

#### `GET /api/v1/models/{id}` — Get Model

#### `DELETE /api/v1/models/{id}` — Delete Model

Returns `409 Conflict` if model is actively deployed.

### 8.3 Devices

#### `GET /api/v1/devices` — List Devices

Query params: `status`, `fleet_id`, `label` (key=value), `runtime`, `limit`, `offset`

Response `200`:
```json
{
    "devices": [
        {
            "id": "uuid",
            "device_id": "edge-042",
            "name": "Warehouse Camera 3",
            "status": "healthy",
            "arch": "arm64",
            "gpu_type": "nvidia",
            "runtime": "tensorrt",
            "ram_mb": 8192,
            "hardware_model": "NVIDIA Jetson Orin Nano",
            "labels": {"region": "eu", "building": "warehouse-3"},
            "fleet": {"id": "uuid", "name": "production-eu"},
            "active_model": {
                "name": "face-detect",
                "version": "3.2",
                "runtime": "tensorrt",
                "inference_count": 12847,
                "avg_latency_ms": 18.3,
                "accuracy": 0.942,
                "drift_score": 0.03
            },
            "system": {
                "cpu_percent": 45,
                "gpu_percent": 72,
                "ram_mb_used": 1847,
                "disk_percent": 34,
                "temperature_c": 62
            },
            "last_heartbeat": "2026-02-22T10:30:00Z",
            "registered_at": "2026-01-15T08:00:00Z"
        }
    ],
    "total": 247,
    "limit": 50,
    "offset": 0
}
```

#### `GET /api/v1/devices/{device_id}` — Get Device Detail

#### `PATCH /api/v1/devices/{device_id}` — Update Device

Request:
```json
{
    "name": "New Name",
    "labels": {"region": "us", "environment": "staging"},
    "fleet_id": "uuid"
}
```

#### `DELETE /api/v1/devices/{device_id}` — Decommission Device

### 8.4 Fleets

#### `POST /api/v1/fleets` — Create Fleet

Request:
```json
{
    "name": "production-eu",
    "description": "European production fleet",
    "labels": {"region": "eu"}
}
```

#### `GET /api/v1/fleets` — List Fleets

#### `GET /api/v1/fleets/{id}` — Get Fleet (includes device count, active models)

#### `PATCH /api/v1/fleets/{id}` — Update Fleet

#### `DELETE /api/v1/fleets/{id}` — Delete Fleet

### 8.5 Deployments

#### `POST /api/v1/deployments` — Create Deployment

Request:
```json
{
    "model_name": "face-detect",
    "model_version": "4.0",
    "target_type": "fleet",
    "target_id": "uuid-of-fleet",
    "policy": "canary",
    "canary_config": {
        "stages": [
            {"percent": 5, "duration": "1h", "success_metric": "accuracy > 0.90"},
            {"percent": 50, "duration": "6h", "success_metric": "accuracy > 0.92"},
            {"percent": 100}
        ]
    }
}
```

Alternative target types:
```json
{
    "target_type": "device",
    "target_id": "device-id-string"
}
```
```json
{
    "target_type": "labels",
    "target_labels": {"region": "eu", "environment": "staging"}
}
```

Response `201`:
```json
{
    "id": "uuid",
    "model": {"name": "face-detect", "version": "4.0"},
    "state": "rolling_out",
    "total_devices": 247,
    "completed_devices": 0,
    "failed_devices": 0,
    "queued_devices": 5,
    "deployment_policy": "canary",
    "created_at": "2026-02-22T10:30:00Z"
}
```

#### `GET /api/v1/deployments` — List Deployments

Query params: `state`, `model_name`, `limit`, `offset`

#### `GET /api/v1/deployments/{id}` — Get Deployment Detail

Response includes per-device status:
```json
{
    "id": "uuid",
    "state": "rolling_out",
    "total_devices": 247,
    "completed_devices": 198,
    "failed_devices": 2,
    "queued_devices": 5,
    "devices": [
        {"device_id": "edge-042", "state": "completed", "completed_at": "..."},
        {"device_id": "edge-043", "state": "downloading"},
        {"device_id": "edge-099", "state": "failed", "error": "insufficient disk space"},
        {"device_id": "edge-101", "state": "queued"}
    ]
}
```

#### `POST /api/v1/deployments/{id}/cancel` — Cancel Deployment

#### `POST /api/v1/deployments/{id}/rollback` — Rollback Deployment

### 8.6 A/B Tests

#### `POST /api/v1/ab-tests` — Create A/B Test

Request:
```json
{
    "name": "v3.2 vs v4.0 accuracy test",
    "model_a": {"name": "face-detect", "version": "3.2"},
    "model_b": {"name": "face-detect", "version": "4.0"},
    "split_a": 80,
    "split_b": 20,
    "target_fleet_id": "uuid",
    "metrics": ["accuracy", "latency"],
    "duration": "7d",
    "auto_promote": true
}
```

#### `GET /api/v1/ab-tests` — List A/B Tests

#### `GET /api/v1/ab-tests/{id}` — Get A/B Test (with results)

#### `POST /api/v1/ab-tests/{id}/stop` — Stop A/B Test

### 8.7 Policies

#### `POST /api/v1/policies` — Create Policy

Request: YAML body (Content-Type: `application/yaml`) or JSON

#### `GET /api/v1/policies` — List Policies

#### `GET /api/v1/policies/{id}` — Get Policy

#### `PUT /api/v1/policies/{id}` — Update Policy

#### `DELETE /api/v1/policies/{id}` — Delete Policy

### 8.8 Heartbeat (REST fallback)

#### `POST /api/v1/heartbeat` — Submit Heartbeat

Used as REST fallback when gRPC is unavailable.

Request:
```json
{
    "device_id": "edge-042",
    "timestamp": "2026-02-22T10:30:00Z",
    "status": "healthy",
    "models": [
        {
            "name": "face-detect",
            "version": "3.2",
            "runtime": "tensorrt",
            "metrics": {
                "inference_count": 12847,
                "avg_latency_ms": 18.3,
                "p99_latency_ms": 42.1,
                "accuracy": 0.942,
                "drift_score": 0.03
            }
        }
    ],
    "system": {
        "cpu_percent": 45,
        "gpu_percent": 72,
        "ram_mb_used": 1847,
        "disk_percent": 34,
        "temperature_c": 62,
        "uptime_hours": 720
    }
}
```

Response `200`:
```json
{
    "commands": [
        {
            "id": "cmd-uuid",
            "type": "deploy_model",
            "payload": {
                "model": "face-detect",
                "version": "4.0",
                "runtime": "tensorrt",
                "artifact_url": "https://...",
                "checksum": "sha256:abc123...",
                "rollback_version": "3.2"
            }
        }
    ]
}
```

### 8.9 Health

#### `GET /api/v1/health` — Health Check

Response `200`:
```json
{
    "status": "healthy",
    "version": "0.1.0",
    "database": "connected",
    "storage": "connected",
    "uptime": "72h15m"
}
```

---

## 9. CLI Specifications

### Installation

```bash
# Via Go
go install github.com/fleetml/fleetml/cli@latest

# Via pip (Python wrapper)
pip install fleetml

# Via Docker
docker run --rm fleetml/cli deploy model.onnx
```

### CLI Configuration

```bash
# fleetml init creates ~/.fleetml/config.yaml
fleetml init
```

Config file `~/.fleetml/config.yaml`:
```yaml
server:
  address: "http://localhost:8080"
  api_key: "your-api-key"
```

### Commands

#### `fleetml init`

Initialize FleetML project. Creates config file and verifies server connectivity.

```bash
$ fleetml init
Server URL [http://localhost:8080]:
API Key: ****
✓ Connected to FleetML server v0.1.0
✓ Config saved to ~/.fleetml/config.yaml
```

#### `fleetml deploy <model_file>`

Deploy a model to devices.

```bash
$ fleetml deploy model.onnx \
    --name face-detect \
    --version 4.0 \
    --fleet production \
    --policy canary \
    --canary-stages "5:1h:accuracy>0.90,50:6h,100"
```

Flags:
- `--name` (string): Model name (inferred from filename if omitted)
- `--version` (string): Model version (auto-incremented if omitted)
- `--fleet` (string): Target fleet name
- `--device` (string): Target single device ID
- `--labels` (string): Label selector (e.g., `region=eu,env=prod`)
- `--policy` (string): Deployment policy — `immediate` (default), `canary`
- `--canary-stages` (string): Canary config — `percent:duration:metric,...`
- `--wait` (bool): Wait for deployment to complete (default: false)
- `--timeout` (duration): Wait timeout (default: 10m)

Output:
```
✓ Uploading model.onnx (12.8 MB)...           done
✓ Registered as face-detect v4.0
✓ Rolling out to 247 devices (fleet: production)
  ├── 198 NVIDIA Jetson    [████████████████████] 100%
  ├── 32 Intel NUC         [████████████████████] 100%
  ├── 12 Raspberry Pi      [████████████████████] 100%
  └── 5 offline            [queued]

✓ 242/247 deployed, 5 queued (offline)
Deployment ID: dep-abc123
Monitor: http://localhost:8080/deployments/dep-abc123
```

#### `fleetml status`

Show fleet status.

```bash
$ fleetml status
$ fleetml status --fleet production
$ fleetml status --device edge-042
$ fleetml status --format json
```

Flags:
- `--fleet` (string): Filter by fleet
- `--device` (string): Show single device detail
- `--format` (string): Output format — `table` (default), `json`, `yaml`

Output (table):
```
FLEET: production (247 devices)
STATUS: 🟢 192 Healthy  🟡 47 Warning  🔴 8 Offline

DEVICE           STATUS    MODEL              LATENCY   ACCURACY  DRIFT
edge-042         healthy   face-detect v3.2   18ms      94.2%     0.03
edge-043         healthy   face-detect v3.2   22ms      93.8%     0.05
edge-099         warning   face-detect v3.2   95ms      89.1%     0.18
edge-101         offline   face-detect v3.1   —         —         —
...
```

#### `fleetml logs <device_id>`

Stream device logs.

```bash
$ fleetml logs edge-042
$ fleetml logs edge-042 --follow
$ fleetml logs edge-042 --since 1h --level error
```

Flags:
- `--follow` / `-f` (bool): Stream logs in real time
- `--since` (duration): Show logs since (e.g., `1h`, `24h`, `7d`)
- `--level` (string): Filter by level — `debug`, `info`, `warn`, `error`
- `--limit` (int): Max lines (default: 100)

#### `fleetml rollback <device_or_fleet>`

Rollback to previous model version.

```bash
$ fleetml rollback edge-042
$ fleetml rollback --fleet production
$ fleetml rollback --fleet production --to-version 3.1
```

Flags:
- `--fleet` (string): Rollback entire fleet
- `--to-version` (string): Specific version to rollback to (default: previous)

#### `fleetml ab-test`

Run A/B test.

```bash
$ fleetml ab-test \
    --model-a face-detect:3.2 \
    --model-b face-detect:4.0 \
    --split 80/20 \
    --fleet production \
    --metric accuracy,latency \
    --duration 7d \
    --auto-promote
```

Flags:
- `--model-a` (string): Model A (name:version)
- `--model-b` (string): Model B (name:version)
- `--split` (string): Traffic split (e.g., `80/20`)
- `--fleet` (string): Target fleet
- `--metric` (string): Metrics to compare (comma-separated)
- `--duration` (duration): Test duration
- `--auto-promote` (bool): Auto-promote winner

#### `fleetml version`

Print CLI and server version.

```bash
$ fleetml version
CLI: v0.1.0
Server: v0.1.0 (http://localhost:8080)
```

---

## 10. Web Dashboard Specifications

### Tech Stack

- React 18 + TypeScript 5
- Tailwind CSS 3
- Vite 5 for build
- React Router 6 for navigation
- TanStack Query (React Query) for API state management
- Recharts for charts
- React Table for data tables

### Pages

#### 10.1 Dashboard (Overview) — `/`

Displays:
- Fleet summary cards: Total devices, Healthy, Warning, Offline
- Active model info: Name, version, accuracy, latency (P50/P95/P99), throughput
- Drift alerts: Devices showing accuracy drift
- Recent deployments: Last 5 deployments with status
- Device groups: Fleet breakdown with device count and active model

#### 10.2 Devices — `/devices`

- Searchable/filterable table of all devices
- Columns: Device ID, Name, Status, Hardware, Runtime, Active Model, Latency, Accuracy, Last Seen
- Filters: Status, Fleet, Runtime, Labels
- Click row → Device detail page

#### 10.3 Device Detail — `/devices/:id`

- Device info card (hardware, OS, labels)
- System metrics charts (CPU, GPU, RAM, Disk over time)
- Model metrics charts (Latency, Accuracy, Throughput over time)
- Deployment history for this device
- Actions: Rollback, Assign to Fleet, Update Labels

#### 10.4 Models — `/models`

- Table of all models
- Columns: Name, Version, Format, Size, Deployed To (count), Created At
- Upload model button → Upload modal
- Click row → Model detail (metadata, deployment history)

#### 10.5 Deployments — `/deployments`

- Table of all deployments
- Columns: Model, Target Fleet, Status, Progress, Created At, Duration
- Click row → Deployment detail (per-device status, progress bar)
- New Deployment button → Deployment wizard

#### 10.6 A/B Tests — `/ab-tests`

- Table of all A/B tests
- Columns: Name, Model A vs B, Split, Status, Winner, Duration
- Click row → Results detail (metric comparison charts)

#### 10.7 Policies — `/policies`

- Table of all policies
- YAML editor for creating/editing policies
- Policy evaluation preview

#### 10.8 Settings — `/settings`

- Server info
- User management (admin only)
- API key management
- Alert configuration

---

## 11. Communication Protocols

### 11.1 Heartbeat Protocol

**Direction**: Agent → Control Plane
**Interval**: Every 30 seconds (configurable)
**Transport**: gRPC bidirectional stream (primary), MQTT (fallback), REST POST (last resort)

**Heartbeat JSON Schema**:

```json
{
    "device_id": "string",
    "timestamp": "ISO8601",
    "status": "healthy | warning | error",
    "models": [
        {
            "name": "string",
            "version": "string",
            "runtime": "tensorrt | openvino | tflite | snpe | onnx",
            "metrics": {
                "inference_count": "int64",
                "avg_latency_ms": "float",
                "p99_latency_ms": "float",
                "accuracy": "float (0-1)",
                "drift_score": "float (0-1)"
            }
        }
    ],
    "system": {
        "cpu_percent": "float (0-100)",
        "gpu_percent": "float (0-100)",
        "ram_mb_used": "int",
        "disk_percent": "float (0-100)",
        "temperature_c": "float",
        "uptime_hours": "float"
    }
}
```

**Target message size**: <1KB compressed (gzip)

### 11.2 Command Protocol

**Direction**: Control Plane → Agent
**Transport**: Piggybacked on heartbeat response

**Command Types**:

| Command Type | Payload | Description |
|---|---|---|
| `deploy_model` | `DeployModelPayload` | Deploy a model to this device |
| `rollback` | `{version: "3.1"}` | Rollback to specified version |
| `set_ab_test` | `SetABTestPayload` | Configure A/B test traffic split |
| `stop_ab_test` | `{test_id: "uuid"}` | Stop A/B test |
| `update_config` | `{key: "value"}` | Update agent configuration |
| `restart` | `{}` | Restart agent |
| `decommission` | `{}` | Gracefully shut down agent |

### 11.3 Offline Resilience Protocol

```
ONLINE → OFFLINE transition:
  1. gRPC stream disconnects
  2. Agent switches to offline mode
  3. Heartbeats stored in local SQLite
  4. Commands buffered in local NATS
  5. Model inference continues uninterrupted
  6. Retry connection every 10s (exponential backoff, max 5min)

OFFLINE → ONLINE transition:
  1. gRPC reconnects
  2. Agent sends BulkSyncMetrics (all buffered heartbeats, compressed)
  3. Server responds with BulkSyncResponse (all queued commands)
  4. Agent processes queued commands (deploy, rollback, etc.)
  5. Resume normal heartbeat streaming
  6. Clear local buffers
```

---

## 12. Security Specifications

### 12.1 Transport Security

- TLS 1.3 for ALL agent ↔ control plane communication
- Minimum cipher: TLS_AES_256_GCM_SHA384
- Certificate rotation: Every 90 days

### 12.2 Agent Authentication (mTLS)

- Each agent has a unique TLS certificate
- Certificate issued during registration
- Certificate includes device_id in Subject Alternative Name (SAN)
- Control plane validates agent certificate on every connection
- Revocation via CRL or OCSP

### 12.3 API Authentication

- JWT tokens for REST API access
- API keys for programmatic access (CLI, CI/CD)
- Token expiry: 24 hours (configurable)
- Refresh tokens: 30 days

### 12.4 RBAC Roles

| Role | Permissions |
|------|------------|
| `admin` | Full access: manage users, deploy, rollback, configure |
| `deployer` | Deploy models, create A/B tests, manage policies |
| `viewer` | Read-only: view devices, models, deployments |

### 12.5 Model Integrity

- SHA-256 checksum computed on upload
- Checksum verified on agent before model load
- Tampered models are rejected with `ErrChecksumMismatch`
- All model artifacts stored encrypted at rest (S3 server-side encryption)

### 12.6 Audit Logging

Every action is logged to `audit_log` table:
- Model uploads, deployments, rollbacks
- Device registrations, decommissions
- Policy changes
- User authentication events
- API key creation/revocation

---

## 13. Compiler Service Specifications

### Architecture

The compiler service runs as a separate Python service with isolated Docker containers for each compiler backend.

### Compilation API

#### `POST /compile`

Request:
```json
{
    "model_url": "s3://fleetml-models/face-detect/3.2/model.onnx",
    "target_runtime": "tensorrt",
    "options": {
        "precision": "fp16",
        "max_batch_size": 1,
        "workspace_mb": 512
    }
}
```

Response `200`:
```json
{
    "artifact_url": "s3://fleetml-models/face-detect/3.2/model.engine",
    "artifact_size": 15200000,
    "checksum": "sha256:def456...",
    "compilation_time_sec": 12.3,
    "target_runtime": "tensorrt"
}
```

### Supported Compilers

| Runtime | Input | Output | Docker Image | GPU Required |
|---|---|---|---|---|
| TensorRT | .onnx | .engine | `fleetml/compiler-tensorrt` | Yes (NVIDIA) |
| OpenVINO | .onnx | .xml + .bin | `fleetml/compiler-openvino` | No |
| TFLite | .onnx | .tflite | `fleetml/compiler-tflite` | No |
| SNPE | .onnx | .dlc | `fleetml/compiler-snpe` | No |
| Hailo | .onnx | .hef | `fleetml/compiler-hailo` | No |

### Compiler Interface

```python
# compiler/compilers/base.py
from abc import ABC, abstractmethod

class BaseCompiler(ABC):
    @abstractmethod
    def compile(self, input_path: str, output_path: str, options: dict) -> dict:
        """Compile model. Returns metadata dict."""
        pass

    @abstractmethod
    def validate(self, model_path: str) -> bool:
        """Validate compiled model."""
        pass

    @abstractmethod
    def supported_formats(self) -> list[str]:
        """Return supported input formats."""
        pass
```

---

## 14. Docker & Deployment Configuration

### `docker-compose.yml` (Self-Hosted Production)

```yaml
version: "3.8"

services:
  server:
    image: fleetml/server:latest
    build: ./server
    ports:
      - "8080:8080"    # REST API
      - "50051:50051"  # gRPC
    environment:
      - DATABASE_URL=postgres://fleetml:${DB_PASSWORD}@db:5432/fleetml
      - S3_ENDPOINT=http://minio:9000
      - S3_ACCESS_KEY=${MINIO_ACCESS_KEY}
      - S3_SECRET_KEY=${MINIO_SECRET_KEY}
      - S3_BUCKET=fleetml-models
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      db:
        condition: service_healthy
      minio:
        condition: service_started
    restart: unless-stopped

  dashboard:
    image: fleetml/dashboard:latest
    build: ./dashboard
    ports:
      - "3000:80"
    environment:
      - VITE_API_URL=http://localhost:8080
    restart: unless-stopped

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: fleetml
      POSTGRES_USER: fleetml
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U fleetml"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_ACCESS_KEY}
      MINIO_ROOT_PASSWORD: ${MINIO_SECRET_KEY}
    volumes:
      - miniodata:/data
    restart: unless-stopped

volumes:
  pgdata:
  miniodata:
```

### `docker-compose.test.yml` (Test Environment)

```yaml
version: "3.8"

services:
  control-plane:
    build: ./server
    ports: ["8080:8080", "50051:50051"]
    environment:
      - DATABASE_URL=postgres://postgres:test@db:5432/fleetml_test
      - S3_ENDPOINT=http://minio:9000
      - S3_ACCESS_KEY=minioadmin
      - S3_SECRET_KEY=minioadmin
      - S3_BUCKET=fleetml-test
      - JWT_SECRET=test-secret
    depends_on: [db, minio]

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: fleetml_test
      POSTGRES_PASSWORD: test

  minio:
    image: minio/minio:latest
    command: server /data
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin

  agent-1:
    build:
      context: ./agent
      dockerfile: Dockerfile
    environment:
      - FLEETML_SERVER=control-plane:50051
      - DEVICE_ID=test-agent-1
      - FLEETML_MODE=test
```

### Agent Dockerfile

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY agent/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /fleetml-agent ./cmd/agent

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /fleetml-agent /usr/local/bin/fleetml-agent
ENTRYPOINT ["fleetml-agent"]
```

### Server Dockerfile

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /fleetml-server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /fleetml-server /usr/local/bin/fleetml-server
COPY server/migrations/ /migrations/
EXPOSE 8080 50051
ENTRYPOINT ["fleetml-server"]
```

### Makefile

```makefile
.PHONY: build test lint agent server cli dashboard

# Build all
build: agent server cli dashboard

# Build agent (cross-compile)
agent:
	cd agent && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../bin/agent-linux-amd64 ./cmd/agent
	cd agent && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ../bin/agent-linux-arm64 ./cmd/agent

# Build server
server:
	cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../bin/server-linux-amd64 ./cmd/server

# Build CLI
cli:
	cd cli && go build -ldflags="-s -w" -o ../bin/fleetml ./cmd/fleetml

# Build dashboard
dashboard:
	cd dashboard && npm install && npm run build

# Run all tests
test: test-unit test-integration

# Unit tests
test-unit:
	cd agent && go test -race -short ./...
	cd server && go test -race -short ./...
	cd cli && go test -race -short ./...

# Integration tests (requires Docker)
test-integration:
	docker compose -f docker-compose.test.yml up -d
	cd tests && go test -v -tags=integration ./integration/...
	docker compose -f docker-compose.test.yml down -v

# Virtual fleet tests
test-fleet:
	docker compose -f docker-compose.test.yml up -d
	cd tests && go test -v -tags=fleet ./fleet/... -fleet-size=20
	docker compose -f docker-compose.test.yml down -v

# Chaos tests
test-chaos:
	docker compose -f docker-compose.test.yml up -d
	cd tests && go test -v -tags=chaos ./chaos/... -timeout=60m
	docker compose -f docker-compose.test.yml down -v

# Lint
lint:
	cd agent && golangci-lint run ./...
	cd server && golangci-lint run ./...
	cd cli && golangci-lint run ./...
	cd dashboard && npm run lint

# Docker build
docker-build:
	docker build -t fleetml/agent:latest -f agent/Dockerfile .
	docker build -t fleetml/server:latest -f server/Dockerfile .
	docker build -t fleetml/dashboard:latest -f dashboard/Dockerfile .

# Start local dev environment
dev:
	docker compose up -d db minio
	cd server && go run ./cmd/server &
	cd dashboard && npm run dev &
```

---

## 15. Testing Plan

### 15.1 Testing Pyramid Overview

```
                      ╱╲
                     ╱  ╲
                    ╱ E2E╲           5-10 scenarios (real hardware)
                   ╱──────╲          Runs: Pre-release
                  ╱ Chaos   ╲        20-30 experiments
                 ╱ Engineer. ╲       Runs: Nightly
                ╱─────────────╲
               ╱ Virtual Fleet ╲     50+ scenarios (100-1000 virtual devices)
              ╱─────────────────╲    Runs: Every PR
             ╱   Integration     ╲   100+ tests (Docker Compose)
            ╱─────────────────────╲  Runs: Every PR
           ╱     Unit Tests        ╲ 500+ tests (pure functions)
          ╱─────────────────────────╲ Runs: Every commit
```

| Layer | Tests | Run Time | Trigger | Cost |
|---|---|---|---|---|
| Unit Tests | 500+ | 30 sec | Every commit | Free |
| Integration | 100+ | 3-5 min | Every PR | Free |
| Virtual Fleet | 50+ | 10-15 min | Every PR | Free |
| Chaos Engineering | 20-30 | 30-60 min | Nightly | Free |
| Performance/Scale | 10+ | 15-30 min | Nightly | Free |
| ML-Specific | 20+ | 5-10 min | Every PR | Free |
| Security | 15+ | 3-5 min | Every PR | Free |
| Hardware-in-Loop | 5-10 | 15-30 min | Pre-release | ₹50K one-time |
| E2E | 5-10 | 30-60 min | Pre-release | Same hardware |

### 15.2 Unit Test Coverage Map

| Component | Test File | Tests | What's Tested |
|---|---|---|---|
| Agent — Model Loader | `agent/internal/model/loader_test.go` | 15+ | Load valid ONNX, invalid checksum, corrupted model, format detection, size limits |
| Agent — Health Reporter | `agent/internal/health/reporter_test.go` | 10+ | System metrics, GPU detection (nvidia/intel/none), edge cases |
| Agent — Rollback Manager | `agent/internal/deploy/rollback_test.go` | 10+ | Version storage, eviction (keep last N), restore, disk full handling |
| Agent — Heartbeat Protocol | `agent/internal/heartbeat/protocol_test.go` | 10+ | Serialization, compression, timing, gzip encoding |
| Agent — Drift Detector | `agent/internal/drift/detector_test.go` | 15+ | PSI calculation, KS test, threshold triggers, window management |
| Server — Fleet Manager | `server/internal/fleet/manager_test.go` | 20+ | Registration, grouping, label selector, status updates |
| Server — Model Registry | `server/internal/model/registry_test.go` | 15+ | Upload, versioning, metadata, deletion guard |
| Server — Policy Engine | `server/internal/policy/engine_test.go` | 15+ | Rule evaluation, rollback triggers, canary logic, YAML parsing |
| Server — Deploy Orchestrator | `server/internal/deploy/orchestrator_test.go` | 20+ | Scheduling, progress tracking, failure handling, canary stages |
| CLI — Argument Parsing | `cli/cmd/*_test.go` | 10+ | Validation, defaults, error messages, flag parsing |

**Unit Test Examples (from Testing Strategy)**:

```go
// agent/internal/model/loader_test.go
func TestModelLoader_ValidONNX(t *testing.T) {
    loader := NewModelLoader("/tmp/models")
    model, err := loader.Load("yolov8n.onnx")
    assert.NoError(t, err)
    assert.Equal(t, "yolov8n", model.Name)
    assert.Equal(t, "onnx", model.Format)
}

func TestModelLoader_InvalidChecksum(t *testing.T) {
    loader := NewModelLoader("/tmp/models")
    err := loader.ValidateChecksum("model.onnx", "sha256:wrong_hash")
    assert.ErrorIs(t, err, ErrChecksumMismatch)
}

func TestModelLoader_CorruptedModel(t *testing.T) {
    loader := NewModelLoader("/tmp/models")
    os.WriteFile("/tmp/models/corrupt.onnx", []byte("not a model"), 0644)
    _, err := loader.Load("corrupt.onnx")
    assert.ErrorIs(t, err, ErrInvalidModel)
}

// agent/internal/deploy/rollback_test.go
func TestRollbackManager_KeepsLastN(t *testing.T) {
    rm := NewRollbackManager("/tmp/rollback", 3)
    rm.SaveVersion("v1", modelBytesV1)
    rm.SaveVersion("v2", modelBytesV2)
    rm.SaveVersion("v3", modelBytesV3)
    rm.SaveVersion("v4", modelBytesV4)
    assert.False(t, rm.HasVersion("v1")) // v1 evicted
    assert.True(t, rm.HasVersion("v4"))
}

// server/internal/policy/engine_test.go
func TestPolicyEngine_AutoRollbackTrigger(t *testing.T) {
    engine := NewPolicyEngine()
    engine.AddPolicy(Policy{
        Rollback: RollbackConfig{
            Trigger: []Condition{
                {Metric: "accuracy", Op: "<", Value: 0.85},
                {Metric: "latency_p99_ms", Op: ">", Value: 100},
            },
        },
    })
    assert.False(t, engine.ShouldRollback(Metrics{Accuracy: 0.92, LatencyP99: 45}))
    assert.True(t, engine.ShouldRollback(Metrics{Accuracy: 0.80, LatencyP99: 45}))
    assert.True(t, engine.ShouldRollback(Metrics{Accuracy: 0.92, LatencyP99: 150}))
}
```

### 15.3 Integration Tests

**Docker Compose Test Environment**: Uses `docker-compose.test.yml` (see Section 14)

**Key Integration Tests**:

```go
// tests/integration/deploy_test.go
func TestModelDeployment_HappyPath(t *testing.T) {
    ctx := setupTestEnvironment(t)
    modelID, _ := ctx.API.UploadModel("test-yolov8.onnx", testModelBytes)
    deployID, _ := ctx.API.Deploy(modelID, "test-agent-1")
    waitFor(t, 30*time.Second, func() bool {
        status, _ := ctx.API.GetDeployment(deployID)
        return status.State == "completed"
    })
    agentStatus, _ := ctx.API.GetDevice("test-agent-1")
    assert.Equal(t, modelID, agentStatus.ActiveModel.ID)
}

func TestModelRollback_AfterFailedDeploy(t *testing.T) {
    ctx := setupTestEnvironment(t)
    v1, _ := ctx.API.UploadModel("model-v1.onnx", goodModelBytes)
    ctx.API.Deploy(v1, "test-agent-1")
    waitForDeployment(t, ctx, "test-agent-1", v1)
    v2, _ := ctx.API.UploadModel("model-v2.onnx", corruptModelBytes)
    ctx.API.Deploy(v2, "test-agent-1")
    waitFor(t, 30*time.Second, func() bool {
        device, _ := ctx.API.GetDevice("test-agent-1")
        return device.ActiveModel.ID == v1
    })
}

func TestOTAUpdate_ZeroDowntime(t *testing.T) {
    // Deploy v1, start continuous inference, deploy v2 while running
    // Assert: zero dropped inferences during swap
}
```

### 15.4 Virtual Fleet Simulator

**Device Profiles**:

| Profile | Arch | RAM | GPU | Runtime | Inference (ms) | Disk |
|---|---|---|---|---|---|---|
| `jetson-orin-nano` | arm64 | 8192 MB | nvidia | tensorrt | 12ms | 64 GB |
| `rpi5-4gb` | arm64 | 4096 MB | none | tflite | 95ms | 16 GB |
| `rpi5-8gb` | arm64 | 8192 MB | none | tflite | 85ms | 32 GB |
| `intel-nuc-i5` | amd64 | 16384 MB | intel | openvino | 22ms | 256 GB |
| `generic-x86` | amd64 | 32768 MB | none | onnx | 55ms | 500 GB |

**Network Profiles**:

| Profile | Latency | Jitter | Packet Loss | Bandwidth |
|---|---|---|---|---|
| `excellent` | 5ms | 2ms | 0.01% | 100 Mbps |
| `good-wifi` | 15ms | 10ms | 0.5% | 50 Mbps |
| `poor-wifi` | 50ms | 30ms | 2.0% | 10 Mbps |
| `cellular-4g` | 80ms | 40ms | 1.5% | 20 Mbps |
| `cellular-3g` | 200ms | 100ms | 5.0% | 2 Mbps |
| `satellite` | 600ms | 200ms | 3.0% | 5 Mbps |
| `intermittent` | 100ms | 50ms | 10.0% | 5 Mbps |

Network conditions applied using Linux `tc netem` inside Docker containers.

**Fleet Scenario Tests**:

```go
func TestFleet_DeployToHeterogeneousFleet(t *testing.T)     // 20 mixed devices
func TestFleet_DeployWithOfflineDevices(t *testing.T)        // 10 devices, 3 offline
func TestFleet_CanaryWithAutoRollback(t *testing.T)          // 100 devices, bad model
```

### 15.5 Chaos Engineering

**Fault Categories**:

| Category | Faults to Test |
|---|---|
| **Network** | Complete disconnect, 50% packet loss, 2-5s latency, 56kbps throttle, DNS failure, TLS expiry |
| **Device** | Power failure (docker kill), disk full mid-download, OOM, CPU overload, clock drift |
| **Deployment** | Partial deploy (server crash at 50%), corrupted model, wrong runtime, concurrent deploys |
| **Control Plane** | Server restart mid-deploy, database failure, S3 failure, API rate limiting |

**Chaos Test Examples**:

```go
func TestChaos_NetworkPartition(t *testing.T)         // Disconnect 8/20 devices during deploy
func TestChaos_PowerFailureDuringOTA(t *testing.T)    // Kill device mid-download, verify recovery
func TestChaos_ServerCrashMidDeploy(t *testing.T)     // Kill server at 50%, verify resume
func TestChaos_DiskFull(t *testing.T)                 // Fill disk to 95%, verify graceful failure
```

### 15.6 Performance & Scale Tests

**Benchmark Targets**:

| Metric | Target |
|---|---|
| API latency p99 | <100ms |
| Agent heartbeat CPU overhead | <1% |
| Agent memory footprint | <50MB RSS |
| Agent binary size | <15MB |
| Deploy time (1 device) | <30 sec |
| Deploy time (100 devices) | <2 min |
| Deploy time (1000 devices) | <10 min |
| Heartbeat throughput (server) | 10K devices/sec |
| Dashboard load time | <2 sec |
| Heartbeat message size | <1KB compressed |
| Reconnection time | <10 sec |

**Scale Tests**:

```go
func TestScale_1000Devices(t *testing.T) {
    // 400 Jetson + 300 RPi + 200 Intel + 100 generic
    // Deploy to all, measure time and resource usage
}
```

**k6 Load Test** (`tests/load/k6-heartbeat.js`):

```javascript
export const options = {
    stages: [
        { duration: '1m', target: 100 },
        { duration: '3m', target: 500 },
        { duration: '2m', target: 1000 },
        { duration: '1m', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(99)<200'],
        http_req_failed: ['rate<0.01'],
    },
};
```

### 15.7 Security Tests

```go
func TestSecurity_UnauthenticatedAgentRejected(t *testing.T)  // Agent without cert rejected
func TestSecurity_TamperedModelRejected(t *testing.T)          // Wrong checksum = rejected
func TestSecurity_CommandInjectionPrevented(t *testing.T)      // Malicious device IDs
func TestSecurity_APIAuthRequired(t *testing.T)                // Unauthenticated = 401
func TestSecurity_RBACEnforcement(t *testing.T)                // Viewer can't deploy
```

### 15.8 ML-Specific Tests

**Model Compatibility Matrix**:

| Model | Format | Size | Test |
|---|---|---|---|
| YOLOv8n | onnx | 12.8 MB | Load + infer |
| YOLOv8s | onnx | 44 MB | Load + infer |
| ResNet50 | onnx | 97 MB | Load + infer |
| MobileNetV3 | onnx | 22 MB | Load + infer |
| EfficientNet-Lite | tflite | 18 MB | Load + infer |
| Whisper-tiny | onnx | 76 MB | Load + infer |

**Drift Detection Tests**: Verify PSI/KS detect distribution shift.

**Model Size Guardrails**: Verify large models rejected on low-RAM devices.

### 15.9 E2E Scenarios

1. **First-Time User Journey**: `pip install fleetml` → `fleetml init` → `fleetml deploy` → `fleetml status` → `fleetml rollback`
2. **Multi-Device Fleet**: Deploy to 10 devices, A/B test, detect drift, auto-rollback
3. **Offline Resilience**: Deploy, disconnect 24 hours, verify inference continues, reconnect, verify sync

### 15.10 CI/CD Pipeline (GitHub Actions)

```yaml
# .github/workflows/ci.yml
name: CI
on:
  push: { branches: [main] }
  pull_request:

jobs:
  lint-and-unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: golangci-lint run ./...
      - run: go test -race -short ./...
      - run: |
          GOOS=linux GOARCH=amd64 go build -o bin/agent-amd64 ./cmd/agent
          GOOS=linux GOARCH=arm64 go build -o bin/agent-arm64 ./cmd/agent

  integration:
    runs-on: ubuntu-latest
    needs: lint-and-unit
    if: github.event_name == 'pull_request'
    steps:
      - uses: actions/checkout@v4
      - run: docker compose -f docker-compose.test.yml up -d
      - run: go test -v -tags=integration ./tests/integration/...
      - run: go test -v -tags=fleet ./tests/fleet/... -fleet-size=20
      - if: always()
        run: docker compose -f docker-compose.test.yml down -v

  nightly:
    runs-on: ubuntu-latest
    if: github.event.schedule == 'cron(0 2 * * *)'
    steps:
      - uses: actions/checkout@v4
      - run: go test -v -tags=fleet ./tests/fleet/... -fleet-size=500 -timeout=30m
      - run: go test -v -tags=chaos ./tests/chaos/... -timeout=60m
      - run: go test -v -tags=benchmark ./tests/benchmark/... -bench=.
      - run: k6 run tests/load/k6-heartbeat.js
```

### 15.11 Quality Gates

**PR Merge Requirements**:
- All unit tests pass (500+)
- All integration tests pass
- Virtual fleet tests pass (20 devices)
- No security vulnerabilities (govulncheck)
- No lint errors
- Agent binary builds for amd64 + arm64
- Code coverage ≥ 80% on new code

**Release Candidate Criteria**:
- Nightly fleet test (500 devices) passing for 3+ consecutive nights
- Chaos engineering suite: 100% pass rate
- Performance benchmarks: no regressions > 10%
- API load test: p99 < 200ms at 1000 concurrent
- Hardware-in-loop tests pass on Jetson + RPi
- E2E scenarios pass on real hardware
- 24-hour soak test: zero agent crashes
- Security scan clean (govulncheck + trivy + npm audit)

**v0.1.0 Launch Go/No-Go**:
- [ ] Agent deploys ONNX model on x86 and arm64
- [ ] Agent survives power failure with model intact
- [ ] Agent survives network disconnect, reconnects
- [ ] OTA update with zero inference downtime
- [ ] Canary deployment detects bad model, auto-rolls back
- [ ] 100-device deployment completes in < 2 minutes
- [ ] CLI install-to-first-deploy in < 5 minutes
- [ ] Unauthenticated agents rejected
- [ ] Tampered models rejected

### 15.12 Test File Structure

```
fleetml/
├── agent/
│   ├── internal/model/loader_test.go
│   ├── internal/health/reporter_test.go
│   ├── internal/deploy/rollback_test.go
│   └── internal/heartbeat/protocol_test.go
├── server/
│   ├── internal/fleet/manager_test.go
│   ├── internal/model/registry_test.go
│   ├── internal/policy/engine_test.go
│   └── internal/deploy/orchestrator_test.go
├── cli/
│   └── cmd/deploy_test.go
├── simulator/
│   ├── profiles.go
│   ├── network.go
│   ├── fleet.go
│   └── Dockerfile.virtual-device
├── tests/
│   ├── integration/     (registration, deploy, rollback)
│   ├── fleet/           (heterogeneous, offline, canary)
│   ├── chaos/           (network, device, server chaos)
│   ├── scale/           (1000-device tests)
│   ├── security/        (auth, integrity)
│   ├── ml/              (model formats, drift)
│   ├── hardware/        (real device tests)
│   ├── e2e/             (user journey scripts)
│   └── load/            (k6 load tests)
├── docker-compose.test.yml
└── docker-compose.fleet-sim.yml
```

---

## 16. Performance Targets

| Metric | Target | How to Measure |
|---|---|---|
| Agent binary size | <15MB (stripped) | `ls -la bin/agent-linux-amd64` |
| Agent memory (idle) | <30MB RSS | `ps aux` on real hardware |
| Agent memory (active inference) | <50MB RSS | `ps aux` during inference |
| Agent heartbeat CPU | <1% | `top` on real hardware |
| Agent startup time | <2 sec | Time from process start to first heartbeat |
| Heartbeat message size | <1KB gzipped | Protocol measurement |
| API p50 latency | <20ms | k6 load test |
| API p99 latency | <100ms | k6 load test |
| Deploy 1 device | <30 sec | Virtual fleet test |
| Deploy 100 devices | <2 min | Virtual fleet test |
| Deploy 1000 devices | <10 min | Virtual fleet test |
| Heartbeat throughput | 10K devices/sec | k6 load test |
| Dashboard initial load | <2 sec | Lighthouse |
| Model download speed | Limited by network | Measure on real hardware |
| Reconnection after offline | <10 sec | Chaos test |
| Hot-swap model (zero downtime) | 0 dropped inferences | Integration test |

---

## 17. Development Phases

### Phase 1: MVP (Weeks 1-16)

**Goal**: Working OSS product that deploys ONNX models to a fleet.

| Week | Focus | Deliverables |
|---|---|---|
| 1-2 | Agent Core | Go agent: lifecycle, ONNX Runtime, device fingerprinting |
| 3-4 | Agent Communication + CLI | gRPC client, heartbeat, model download, `fleetml init/deploy/status/logs` |
| 5-6 | Control Plane | Go server, REST API, PostgreSQL, S3/MinIO, Docker Compose |
| 7-8 | Dashboard | React + TS + Tailwind: device list, model registry, deployment view |
| 9-10 | OTA Updates | Deployment orchestration, hot-swap, rollback, progress tracking |
| 11-12 | Multi-Device Testing | Test on Jetson, RPi 5, x86. Fix cross-platform issues |
| 13-14 | Documentation | Docs site, quickstart, CLI reference, API reference, 3 tutorial videos |
| 15-16 | **OSS Launch** | GitHub release, Show HN, Reddit, Twitter/X, Discord |

**MVP includes**: Agent, CLI, self-hosted control plane, basic dashboard, model versioning, OTA updates, device health monitoring.

**MVP excludes**: Multi-chip auto-compilation, A/B testing, drift detection, policy engine.

### Phase 2: Key Features (Months 5-9)

TensorRT compiler, OpenVINO compiler, A/B testing, device grouping + labels, drift detection, MLflow integration, TFLite compiler, policy engine v1.

### Phase 3: Monetization (Months 9-15)

FleetML Cloud (hosted control plane), free tier (5 devices), team features (RBAC, audit log), advanced compilers (SNPE, Hailo), enterprise features (SSO/SAML).

### Phase 4: Scale + Enterprise (Months 15-24)

EU AI Act compliance, federated learning, OEM/white-label, enterprise sales.

---

*End of FleetML PRD v1.0*

*This document contains all specifications needed for a coding agent to implement FleetML from scratch: monorepo structure, database schemas, protobuf definitions, Go interfaces, REST API contracts, CLI commands, dashboard pages, communication protocols, security requirements, Docker configurations, and a complete 9-layer testing plan with 187+ planned tests.*
