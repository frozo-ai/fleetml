<p align="center">
  <h1 align="center">FleetML</h1>
  <p align="center">
    <strong>Deploy AI models to edge device fleets — one command, any chip, offline-first.</strong>
  </p>
  <p align="center">
    <a href="https://github.com/frozo-ai/fleetml/actions/workflows/ci.yml"><img src="https://github.com/frozo-ai/fleetml/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
    <a href="https://github.com/frozo-ai/fleetml/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License"></a>
    <img src="https://img.shields.io/badge/go-1.24+-00ADD8.svg" alt="Go 1.24+">
    <img src="https://img.shields.io/badge/python-3.11+-3776AB.svg" alt="Python 3.11+">
    <img src="https://img.shields.io/badge/tests-854%20passing-brightgreen.svg" alt="Tests">
    <img src="https://img.shields.io/badge/vulnerabilities-0-brightgreen.svg" alt="No vulnerabilities">
  </p>
</p>

---

FleetML is an open-source, chip-neutral edge MLOps platform. Upload an ONNX model, and FleetML compiles it for each chip type, deploys it across your fleet with canary rollouts, and monitors everything — even when devices go offline.

Think **"Kubernetes for edge AI models."**

```bash
# Deploy a model to your entire fleet in one command
fleetml deploy model.onnx --fleet production --canary 5,50,100
```

## The Problem

You trained a great model. Now you need it running on 200 devices across 3 chip types in 4 countries. Today, that means:

- SSH-ing into devices one by one
- Writing custom deployment scripts per chip (TensorRT, OpenVINO, TFLite...)
- Building your own OTA update system
- Praying nothing breaks when you push an update to a remote device

**This takes 6-12 weeks.** And it breaks at 50 devices.

## The Fix

```bash
# Clone and run — one command starts everything with 6 simulated devices
git clone https://github.com/frozo-ai/fleetml.git && cd fleetml
make quickstart
```

That's it. The quickstart script starts infrastructure (PostgreSQL, MinIO, NATS), builds the server, creates a demo user, registers 6 simulated edge devices (Jetson, Raspberry Pi, Intel NUC), uploads a sample model, and prints ready-to-use curl commands. Everything runs locally in about 60 seconds.

```bash
# Then deploy a model to your fleet
fleetml deploy defect-detector.onnx --fleet production --canary 5,50,100

# Watch it roll out
fleetml status --fleet production
```

FleetML handles compilation, OTA delivery, canary rollout, health monitoring, and automatic rollback. Your model is running on every device in minutes — not weeks.

## How It Works

```
You                    Control Plane                     Edge Devices
 │                          │                                │
 │  fleetml deploy          │                                │
 │  model.onnx ────────────►│                                │
 │                          │                                │
 │                    ┌─────┴──────┐                         │
 │                    │  Compile   │                         │
 │                    │  ONNX to:  │                         │
 │                    │  TensorRT  │                         │
 │                    │  OpenVINO  │                         │
 │                    │  TFLite    │                         │
 │                    └─────┬──────┘                         │
 │                          │                                │
 │                    ┌─────┴──────┐    gRPC + mTLS          │
 │                    │  Deploy    │◄──────────────────►┌────┴────┐
 │                    │  Canary:   │   Heartbeats       │ Agent   │
 │                    │  5% → 50% │   Commands          │ ~15MB   │
 │                    │  → 100%   │   Metrics           │ <30MB   │
 │                    └─────┬──────┘                    │  RAM    │
 │                          │                           │         │
 │  ◄───── Dashboard ──────┘                           │ Hot-swap│
 │         + Metrics                                   │ zero    │
 │         + Alerts                                    │ dropped │
 │                                                     │ infer.  │
 │                                                     └─────────┘
```

**One model in, compiled variants out, deployed everywhere, monitored always.**

## Features

### Deploy & Update
- **One-command deployment** — `fleetml deploy model.onnx --fleet production`
- **Multi-chip auto-compilation** — ONNX in, TensorRT / OpenVINO / TFLite out (matched to each device's hardware)
- **Canary deployments** — Progressive rollout (5% → 50% → 100%) with automatic rollback on failure
- **Zero-downtime hot-swap** — Atomic pointer swap. Zero dropped inferences during model updates
- **Automatic rollback** — Bad model? FleetML reverts to the previous version automatically

### Monitor & Manage
- **Fleet dashboard** — Real-time device health, deployment progress, model metrics
- **Drift detection** — PSI and KS statistical tests catch when input distributions shift
- **A/B testing** — Run two model versions side-by-side with configurable traffic splits
- **Policy engine** — Define deployment rules: hardware constraints, canary stages, rollback triggers

### Edge-Native
- **Offline-first** — SQLite buffer + store-and-forward. Agents work without internet, sync when reconnected
- **Lightweight agent** — <15MB binary, <30MB RAM idle, <2s startup. Runs on a Raspberry Pi Zero
- **Device grouping** — Organize devices by labels, target deployments by fleet or hardware type

### Secure
- **mTLS** — Agent ↔ Server communication encrypted with TLS 1.3
- **JWT + RBAC** — Three roles: admin, deployer, viewer
- **Model integrity** — SHA-256 checksum verification on every deployment

## Benchmarks

| Metric | Value |
|--------|-------|
| Agent binary size | **<15MB** (stripped) |
| Agent memory (idle) | **<30MB** RSS |
| Agent startup | **<2 seconds** |
| API latency | **p50 <20ms**, p99 <100ms |
| Deploy 1 device | **<30 seconds** |
| Deploy 100 devices | **<2 minutes** |
| Deploy 1,000 devices | **<10 minutes** |
| Inferences dropped during hot-swap | **0** |
| Heartbeat overhead | <1KB gzipped, <1% CPU |

## Supported Hardware

| Chip | Runtime | Status |
|------|---------|--------|
| NVIDIA Jetson (Nano, Xavier, Orin) | TensorRT | Supported |
| Intel (CPU, NCS2, Movidius) | OpenVINO | Supported |
| ARM (Raspberry Pi, Coral) | TFLite | Supported |
| Qualcomm (Snapdragon) | SNPE | Planned |
| Hailo-8 | HailoRT | Planned |

FleetML auto-detects device hardware and selects the correct compiled variant. You never write chip-specific code.

## Architecture

```
                         ┌─────────────────────────────────────────┐
                         │           Control Plane (Go)            │
                         │                                         │
  ┌──────────┐    REST   │  ┌──────────┐  ┌──────────────────────┐ │
  │  CLI     │◄─────────►│  │ REST API │  │ Deployment           │ │
  │ fleetml  │   :8080   │  │ (chi v5) │  │ Orchestrator         │ │
  └──────────┘           │  └────┬─────┘  │ - Canary rollout     │ │
                         │       │        │ - Variant selection   │ │
  ┌──────────┐    HTTP   │  ┌────┴─────┐  │ - Auto-rollback      │ │
  │Dashboard │◄─────────►│  │ Fleet    │  └──────────────────────┘ │
  │  React   │   :3000   │  │ Manager  │                           │
  └──────────┘           │  └────┬─────┘  ┌──────────────────────┐ │
                         │       │        │ Model Registry       │ │
                         │  ┌────┴─────┐  │ - ONNX upload        │ │
                         │  │PostgreSQL│  │ - Compiled variants  │ │
                         │  │  (pgx)   │  │ - S3 storage (MinIO) │ │
                         │  └──────────┘  └──────────────────────┘ │
                         │                                         │
                         │  ┌──────────┐  ┌──────────────────────┐ │
                         │  │   NATS   │  │ Compiler Service     │ │
                         │  │ JetStream│  │ (Python/FastAPI)     │ │
                         │  │ cmd queue│  │ ONNX → TensorRT/     │ │
                         │  └────┬─────┘  │ OpenVINO/TFLite      │ │
                         │       │        └──────────────────────┘ │
                         └───────┼─────────────────────────────────┘
                                 │
                    gRPC (:50051) + NATS
                    mTLS encrypted
                                 │
            ┌────────────────────┼────────────────────┐
            │                    │                    │
   ┌────────┴────────┐ ┌────────┴────────┐ ┌────────┴────────┐
   │  Jetson Nano    │ │  Raspberry Pi   │ │  Intel NUC      │
   │  Agent (~15MB)  │ │  Agent (~15MB)  │ │  Agent (~15MB)  │
   │                 │ │                 │ │                 │
   │ ┌─────────────┐ │ │ ┌─────────────┐ │ │ ┌─────────────┐ │
   │ │ TensorRT    │ │ │ │ TFLite      │ │ │ │ OpenVINO    │ │
   │ │ Runtime     │ │ │ │ Runtime     │ │ │ │ Runtime     │ │
   │ ├─────────────┤ │ │ ├─────────────┤ │ │ ├─────────────┤ │
   │ │ Hot-Swap    │ │ │ │ Hot-Swap    │ │ │ │ Hot-Swap    │ │
   │ │ (atomic ptr)│ │ │ │ (atomic ptr)│ │ │ │ (atomic ptr)│ │
   │ ├─────────────┤ │ │ ├─────────────┤ │ │ ├─────────────┤ │
   │ │ SQLite      │ │ │ │ SQLite      │ │ │ │ SQLite      │ │
   │ │ Offline Buf │ │ │ │ Offline Buf │ │ │ │ Offline Buf │ │
   │ └─────────────┘ │ │ └─────────────┘ │ │ └─────────────┘ │
   └─────────────────┘ └─────────────────┘ └─────────────────┘
```

## Compared to Alternatives

| | FleetML | AWS Greengrass | Edge Impulse | Balena | DIY Scripts |
|---|:---:|:---:|:---:|:---:|:---:|
| Open source | Apache 2.0 | No | No | Partial | N/A |
| Chip-neutral | All chips | AWS only | Qualcomm-first* | No ML | Manual |
| ML-native (model versioning, A/B, drift) | Yes | No | Yes | No | No |
| Offline-first | Yes | Partial | No | Yes | No |
| Fleet-scale OTA | Yes | Yes | No | Yes | No |
| Self-hostable | Yes | No | No | No | Yes |
| Setup time | 5 min | Days | Hours | Hours | Weeks |

*Edge Impulse was acquired by Qualcomm in 2024 and is now Qualcomm-first.

## Project Structure

```
fleetml/
├── agent/        # Edge agent (Go) — 15MB, runs on devices
├── server/       # Control plane (Go) — REST + gRPC APIs
├── cli/          # CLI tool (Go) — fleetml deploy/status/rollback
├── dashboard/    # Web UI (React/TypeScript) — fleet monitoring
├── compiler/     # Model compiler (Python/FastAPI) — ONNX → chip runtimes
├── proto/        # Protobuf definitions — agent ↔ server protocol
├── tests/        # Integration, fleet, chaos, load, security tests
└── docs/         # Documentation
```

## CLI

```bash
fleetml deploy model.onnx              # Deploy to default fleet
fleetml deploy model.onnx --fleet gpu  # Deploy to specific fleet
fleetml deploy model.onnx --canary 5,50,100  # Canary rollout
fleetml status                         # Fleet overview
fleetml status --fleet production      # Fleet-specific status
fleetml rollback --deployment abc123   # Roll back a deployment
fleetml logs --device jetson-04        # Stream device logs
fleetml ab-test create --name "v1-vs-v2" --model-a v1 --model-b v2
```

## Documentation

| Doc | Description |
|-----|-------------|
| [Quickstart](docs/quickstart.md) | Deploy your first model in 5 minutes |
| [Installation](docs/installation.md) | Production setup guide |
| [Architecture](docs/architecture.md) | System design deep dive |
| [CLI Reference](docs/cli-reference.md) | All commands and flags |
| [API Reference](docs/api-reference.md) | REST API endpoints |
| [Policies](docs/policies.md) | Deployment policy YAML format |
| [Development](docs/development.md) | Contributing and local dev setup |

## Contributing

We welcome contributions. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

```bash
# Local dev setup
git clone https://github.com/frozo-ai/fleetml.git && cd fleetml
docker compose up -d db minio nats    # Start infrastructure
cd server && go run ./cmd/server      # Start the server
cd dashboard && npm run dev           # Start the dashboard
```

**Good first issues** are tagged in the [issue tracker](https://github.com/frozo-ai/fleetml/issues).

## Roadmap

- [x] Core agent with offline-first architecture
- [x] Control plane (REST + gRPC + PostgreSQL + S3)
- [x] CLI with deploy, rollback, status, logs
- [x] React dashboard with fleet monitoring
- [x] Multi-chip compiler (TensorRT, OpenVINO, TFLite)
- [x] Canary deployments with automatic rollback
- [x] A/B testing with traffic splitting
- [x] Drift detection (PSI + KS tests)
- [x] Policy engine for deployment rules
- [x] MLflow and HuggingFace model import
- [ ] FleetML Cloud (managed SaaS)
- [ ] Qualcomm SNPE and Hailo runtime support
- [ ] Hardware-in-the-loop CI testing
- [ ] Kubernetes operator for control plane

## License

[Apache License 2.0](LICENSE) — use it freely, no vendor lock-in.

---

<p align="center">
  <strong>Built for engineers who ship models to real hardware.</strong>
  <br>
  <a href="docs/quickstart.md">Get Started</a> &nbsp;|&nbsp;
  <a href="docs/architecture.md">Architecture</a> &nbsp;|&nbsp;
  <a href="https://github.com/frozo-ai/fleetml/issues">Issues</a>
</p>
