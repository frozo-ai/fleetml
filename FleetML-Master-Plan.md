# FleetML — The Edge AI Fleet Operating System

## Master Plan: Product, Technical Architecture, Go-To-Market & Execution Roadmap

**Version:** 1.0
**Date:** February 22, 2026
**Status:** Pre-Development
**Confidential**

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [The Problem](#2-the-problem)
3. [Market Opportunity](#3-market-opportunity)
4. [Competitive Landscape](#4-competitive-landscape)
5. [Product Vision & Architecture](#5-product-vision--architecture)
6. [Core Features & Capabilities](#6-core-features--capabilities)
7. [Technical Architecture](#7-technical-architecture)
8. [Development Roadmap](#8-development-roadmap)
9. [Go-To-Market Strategy](#9-go-to-market-strategy)
10. [Business Model & Pricing](#10-business-model--pricing)
11. [Financial Projections](#11-financial-projections)
12. [Team & Hiring Plan](#12-team--hiring-plan)
13. [Risk Analysis & Mitigation](#13-risk-analysis--mitigation)
14. [Success Metrics & KPIs](#14-success-metrics--kpis)
15. [Week-by-Week Execution Plan](#15-week-by-week-execution-plan)

---

## 1. Executive Summary

### One-Line Pitch

> **"Kubernetes for edge AI models — deploy, update, and monitor ML models across thousands of heterogeneous edge devices from a single dashboard."**

### The Problem in One Paragraph

Deploying AI models at the edge is the #1 blocker preventing organizations from moving from pilot to production. Independent surveys confirm fewer than one-third of organizations have fully deployed edge AI, and 70% of Industry 4.0 projects stall in pilot. The reason: existing MLOps tools (MLflow, SageMaker, Vertex AI) stop at the cloud boundary. They give you a packaged model but do NOT deploy it to real edge hardware, monitor it across a fleet, handle OTA updates offline, or manage multi-chip compilation. Engineers spend 6–12 weeks doing manually what should take 6 minutes.

### The Solution

FleetML is an **open-source, chip-neutral edge MLOps platform** that provides the missing layer between a data scientist's notebook and a fleet of edge devices. One command deploys a model to hundreds of devices running different chips (NVIDIA, Intel, Qualcomm, Hailo, ARM). FleetML handles compilation, OTA updates, A/B testing, drift detection, and fleet monitoring — all offline-first.

### Why Now

- **Edge Impulse** was acquired by Qualcomm (2024) — no longer chip-neutral
- **OctoML** was absorbed by NVIDIA (2023) — dead as independent platform
- **The throne is empty** — no independent, open-source edge MLOps platform exists
- **97% of US CIOs** have edge AI on their 2025–2026 technology roadmaps
- **4.3 billion** commercial edge-enabled IoT devices deployed globally in 2025
- **Edge AI market** reached $25.65B in 2025, projected $143B by 2034

### Why FleetML Wins

| Advantage | Detail |
|-----------|--------|
| **Zero certifications needed** | Developer tool, not compliance product. Developers judge code quality, not company location |
| **Open source core** | Apache 2.0 — builds community, eliminates vendor lock-in objection |
| **Chip-neutral** | Works with NVIDIA, Intel, Qualcomm, Hailo, ARM. Every hardware vendor promotes us |
| **10x cheaper to start** | ~₹1L Year 1 GTM vs ₹12-15L for compliance-driven products |
| **India cost advantage** | Profitable at $500K ARR where US competitors need $5M |
| **Developer-led growth** | Bottom-up adoption, no enterprise sales cycle for initial traction |

### Pain Validation Score: 42/50 (Ranked #1 across 12 evaluated ideas)

| Dimension | Score | Evidence |
|-----------|-------|----------|
| Pain Intensity | 9/10 | <1/3 of orgs have fully deployed edge AI. 70% of Industry 4.0 projects stall in pilot |
| Pain Frequency | 9/10 | Every company with edge AI devices faces fragmented SDKs daily |
| Willingness to Pay | 7/10 | 97% of CIOs have edge AI on roadmap. Category is emerging |
| Competitive Vacuum | 8/10 | Edge Impulse → Qualcomm. OctoML → NVIDIA. No neutral platform remains |
| Evidence Strength | 9/10 | Multiple peer-reviewed papers, industry surveys, conference presentations confirm this pain |

---

## 2. The Problem

### The Deployment Gap

```
DATA SCIENTIST                              EDGE DEVICE
┌──────────────────┐                       ┌──────────────────┐
│                  │                       │                  │
│  Trained great   │                       │  NVIDIA Jetson   │
│  model in        │   ????????????????    │  Raspberry Pi    │
│  PyTorch         │ ───────────────────▶  │  Intel NUC       │
│                  │   HOW DO I GET THIS   │  Qualcomm        │
│  Works great in  │   TO 500 DEVICES IN   │  Hailo           │
│  my notebook     │   3 COUNTRIES ON      │  ESP32           │
│                  │   4 DIFFERENT CHIPS?   │                  │
└──────────────────┘                       └──────────────────┘

  MLOps tools stop HERE ──┘
  (MLflow, SageMaker, Vertex AI)
```

### What Engineers Do Today (The Manual Nightmare)

**Step 1: Model Conversion (1–2 weeks)**
- Convert PyTorch/TensorFlow model to ONNX
- Then convert ONNX to chip-specific runtime:
  - NVIDIA → TensorRT
  - Intel → OpenVINO
  - Qualcomm → SNPE/QNN
  - ARM → TFLite
  - Hailo → Hailo Dataflow Compiler
- Each conversion requires different toolchain, different expertise, different bugs

**Step 2: Deployment to Devices (1–2 weeks)**
- SSH into each device or build custom deployment scripts
- Copy model files, update configurations
- Handle network failures, partial deployments
- No rollback mechanism if something breaks
- Manual testing on each device type

**Step 3: Monitoring (Ongoing, manual)**
- No fleet-wide visibility
- SSH into individual devices to check logs
- No drift detection — model degrades silently
- No alerting when devices go offline
- No performance benchmarking across fleet

**Step 4: Updates (1–2 weeks per update cycle)**
- Repeat Step 1–3 for every model update
- No A/B testing capability
- No canary deployments
- One bad model update can brick entire fleet

**Total time per deployment cycle: 4–8 weeks**
**FleetML reduces this to: 5 minutes**

### Who Feels This Pain

| Persona | Pain Level | Frequency | Current Workaround |
|---------|-----------|-----------|-------------------|
| **ML Engineer** | 🔴 Critical | Daily | Custom scripts, SSH, manual processes |
| **DevOps / MLOps Engineer** | 🔴 Critical | Daily | Ansible/Terraform hacks, not ML-aware |
| **Edge AI Team Lead** | 🟡 High | Weekly | Hiring more engineers to handle deployment |
| **VP Engineering** | 🟡 High | Monthly | Accepting 3–6 month deployment cycles |
| **CTO** | 🟠 Medium | Quarterly | Evaluating whether edge AI is "worth it" |

### Evidence of Pain (Direct Quotes from Industry)

> *"Deploying at the edge is the blocker."*
> — Edge AI and Vision Alliance, December 2025

> *"Although edge intelligence methods have been proposed... they still face multiple persistent challenges, such as large-scale model deployment."*
> — Wang et al., Mathematics 2025

> *"Traditional MLOps involves selecting model, tuning to fit device constraints, compiling for edge hardware, validating performance — a process that can take several weeks."*
> — Latent AI (claims they reduced this from 6–12 weeks to 48 hours for US Navy)

> *"Nearly half of AI PoCs [are] scrapped before production."*
> — ZEDEDA/Censuswide Industry Survey, 2025

---

## 3. Market Opportunity

### Market Sizing

| Market | Size (2025) | Projected | CAGR | Source |
|--------|-------------|-----------|------|--------|
| Edge AI | $25.65B | $143B by 2034 | 20%+ | Market.us |
| Edge Computing | $82B (2026) | $217B US | 35%+ | Bayelsawatch |
| MLOps | ~$4B | $20–30B by 2028 | 40%+ | Various |
| **Edge MLOps (overlap)** | **~$500M** | **$5–10B by 2028** | **30%+** | Estimated |

### Total Addressable Market (TAM)

- Companies deploying AI on edge devices × average spend on deployment tooling
- ~200,000 companies globally with edge AI initiatives
- Average spend $5,000–$100,000/year on deployment tooling
- **TAM: $5–10B**

### Serviceable Addressable Market (SAM)

- Companies deploying AI on 10+ edge devices
- Need fleet management + model deployment
- ~50,000 companies globally
- Average spend $5,000–$50,000/year
- **SAM: $500M–$2.5B**

### Serviceable Obtainable Market (SOM) — Year 3

- 200–500 customers × $20K average annual contract
- **SOM: $4M–$10M ARR**

### Key Market Drivers

1. **Edge AI spending surge**: 90% of enterprises increasing edge AI budgets for 2025–2026
2. **Hardware proliferation**: 4.3B edge-enabled IoT devices in 2025, growing to 4.8B+ in 2026
3. **Chip diversity explosion**: NVIDIA, Intel, Qualcomm, AMD, Hailo, Google TPU, Apple Neural Engine — each requires different compilation
4. **Regulatory push**: EU AI Act mandates audit trails, driving need for managed model lifecycle
5. **Real-time requirements**: Edge inference eliminates cloud round-trip latency (10ms vs 100ms+)
6. **Data sovereignty**: 120+ countries with data protection laws pushing processing to edge

---

## 4. Competitive Landscape

### The Competitive Vacuum is Real

| Competitor | Status | What They Do | Critical Limitation |
|-----------|--------|-------------|-------------------|
| **Edge Impulse** | Acquired by Qualcomm (2024) | tinyML development platform | Now Qualcomm-first. No longer chip-neutral |
| **OctoML** | Absorbed into NVIDIA (2023) | ML model optimization | Dead as independent. NVIDIA-only |
| **Latent AI** | Active ($30M+ raised) | Model optimization for military/defense | Enterprise-only, defense-focused. Not fleet orchestration |
| **ZEDEDA** | Active ($80M+ raised) | Edge device management | Infrastructure management, NOT ML-specific. "Edge Kubernetes" not "edge MLOps" |
| **balena** | Active ($70M+ raised) | IoT fleet management | Developer-friendly but NOT ML-aware. No model versioning, drift detection, A/B testing |
| **AWS IoT Greengrass** | Active | Edge runtime for AWS | Cloud lock-in (AWS only). Complex. Not ML-focused |
| **Azure IoT Edge** | Active | Edge runtime for Azure | Cloud lock-in (Azure only). Enterprise complexity |
| **MLflow** | Open source | Cloud MLOps lifecycle | Edge is afterthought. No fleet management, OTA, offline-first |
| **Kubeflow** | Open source | Kubernetes-native ML | Massive overhead. Not designed for resource-constrained edge |

### The Gap Nobody Has Filled

**Nobody offers a CHIP-NEUTRAL, OPEN, LIGHTWEIGHT edge MLOps platform combining:**

- ✅ Model deployment to heterogeneous hardware
- ✅ Fleet-wide OTA updates (with offline resilience)
- ✅ Model versioning + A/B testing at edge
- ✅ Drift detection + automated retraining triggers
- ✅ Multi-chip compilation (ONNX → TensorRT/OpenVINO/TFLite/SNPE in one click)
- ✅ Edge-first (not cloud-dependent)
- ✅ Simple enough for a 10-person team

**Edge Impulse WAS this. But Qualcomm killed neutrality. The throne is EMPTY.**

### Competitive Positioning Map

```
                    CLOUD-NATIVE          EDGE-NATIVE
                    (Cloud-dependent)      (Offline-first)
                   ┌──────────────────┬──────────────────┐
                   │                  │                  │
   CHIP-LOCKED     │  AWS Greengrass  │  Edge Impulse    │
   (Vendor lock-in)│  Azure IoT Edge  │  (now Qualcomm)  │
                   │  Google Coral    │                  │
                   │                  │                  │
                   ├──────────────────┼──────────────────┤
                   │                  │                  │
   CHIP-NEUTRAL    │  MLflow          │                  │
   (Open)          │  Kubeflow        │  ★ FleetML ★     │
                   │  (cloud MLOps,   │  (goes HERE)     │
                   │   not edge)      │                  │
                   │                  │                  │
                   └──────────────────┴──────────────────┘
```

### Why Hardware Vendors Will Promote FleetML (For Free)

FleetML is **"Switzerland"** — we're not aligned with any chip vendor, which means EVERY vendor benefits from promoting us:

- **NVIDIA** wants more developers buying Jetsons → FleetML makes Jetson deployment easier
- **Intel** wants more OpenVINO adoption → FleetML integrates OpenVINO natively
- **Qualcomm** wants edge AI on Snapdragon → FleetML supports SNPE
- **Hailo** is new and hungry for ecosystem → FleetML adds Hailo support = free marketing from Hailo
- **Raspberry Pi Foundation** wants ML on Pi → FleetML tutorials drive Pi purchases

Each vendor will co-market, list us in their ecosystem, and potentially give us free hardware and GPU credits.

---

## 5. Product Vision & Architecture

### Product Philosophy

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│   "Make edge AI deployment as simple as                     │
│    git push is for web deployment."                         │
│                                                             │
│   PRINCIPLES:                                               │
│                                                             │
│   1. ONE COMMAND — Deploy to entire fleet with one command  │
│   2. ANY CHIP — Works on NVIDIA, Intel, Qualcomm, ARM      │
│   3. OFFLINE-FIRST — Edge devices work without internet     │
│   4. OPEN — Apache 2.0 core, no vendor lock-in             │
│   5. OBSERVABLE — Know what every device is doing, always   │
│   6. SAFE — Rollback instantly if something goes wrong      │
│   7. LIGHTWEIGHT — Agent runs on 512MB RAM devices          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### System Architecture (High-Level)

```
┌─────────────────────────────────────────────────────────────────────┐
│                    FLEETML CONTROL PLANE                            │
│                 (Cloud-hosted or Self-hosted)                       │
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │   Model      │  │   Fleet      │  │   Multi-Chip │             │
│  │   Registry   │  │   Manager    │  │   Compiler   │             │
│  │              │  │              │  │              │             │
│  │ • Versioning │  │ • Device     │  │ • ONNX input │             │
│  │ • Metadata   │  │   inventory  │  │ • TensorRT   │             │
│  │ • Lineage    │  │ • Grouping   │  │ • OpenVINO   │             │
│  │ • Artifacts  │  │ • Policies   │  │ • TFLite     │             │
│  └──────────────┘  └──────────────┘  │ • SNPE       │             │
│                                       └──────────────┘             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │   Monitor &  │  │   Policy     │  │   Data       │             │
│  │   Alerts     │  │   Engine     │  │   Pipeline   │             │
│  │              │  │              │  │              │             │
│  │ • Health     │  │ • Deployment │  │ • Telemetry  │             │
│  │ • Performance│  │   rules      │  │ • Retraining │             │
│  │ • Drift      │  │ • Rollback   │  │   triggers   │             │
│  │ • Anomalies  │  │   triggers   │  │ • Edge data  │             │
│  └──────────────┘  └──────────────┘  │   sampling   │             │
│                                       └──────────────┘             │
│  ┌───────────────────────────────────────────────────────┐        │
│  │              WEB DASHBOARD + REST API + gRPC           │        │
│  └───────────────────────────────────────────────────────┘        │
│  ┌───────────────────────────────────────────────────────┐        │
│  │              CLI: fleetml deploy | status | rollback   │        │
│  └───────────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────────────┘
         │                    │                    │
         │    gRPC / MQTT     │                    │
         │    (encrypted)     │                    │
         ▼                    ▼                    ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  FleetML     │  │  FleetML     │  │  FleetML     │
│  Agent       │  │  Agent       │  │  Agent       │
│              │  │              │  │              │
│  NVIDIA      │  │  Intel NUC   │  │  Raspberry   │
│  Jetson Orin │  │  + OpenVINO  │  │  Pi 5        │
│  + TensorRT  │  │              │  │  + TFLite    │
│              │  │              │  │              │
│  ~50MB       │  │  ~50MB       │  │  ~30MB       │
│  footprint   │  │  footprint   │  │  footprint   │
└──────────────┘  └──────────────┘  └──────────────┘
```

### FleetML Agent (Edge Component)

```
┌─────────────────────────────────────────────────┐
│              FLEETML AGENT (~50MB)               │
│                                                  │
│  ┌──────────────────────────────────────────┐   │
│  │           Communication Layer             │   │
│  │  • gRPC client (control plane connection) │   │
│  │  • MQTT fallback (for constrained devices)│   │
│  │  • Store-and-forward (offline queue)      │   │
│  └──────────────────────────────────────────┘   │
│                                                  │
│  ┌──────────────────────────────────────────┐   │
│  │           Model Manager                   │   │
│  │  • Local model storage (versioned)        │   │
│  │  • Hot-swap model loading                 │   │
│  │  • Runtime selection (TRT/OV/TFLite/SNPE) │   │
│  │  • A/B traffic splitting                  │   │
│  └──────────────────────────────────────────┘   │
│                                                  │
│  ┌──────────────────────────────────────────┐   │
│  │           Health Reporter                 │   │
│  │  • CPU/GPU/RAM/disk monitoring            │   │
│  │  • Inference latency tracking             │   │
│  │  • Throughput counting                    │   │
│  │  • Error rate tracking                    │   │
│  └──────────────────────────────────────────┘   │
│                                                  │
│  ┌──────────────────────────────────────────┐   │
│  │           Drift Detector                  │   │
│  │  • Input distribution monitoring          │   │
│  │  • Confidence score tracking              │   │
│  │  • Statistical drift tests (PSI, KS)      │   │
│  │  • Edge data sampling for retraining      │   │
│  └──────────────────────────────────────────┘   │
│                                                  │
│  Written in: Go (primary) or Rust               │
│  Why: Small binary, cross-compilation, fast      │
│  Supported: Linux ARM64, Linux x86_64, RTOS     │
└─────────────────────────────────────────────────┘
```

---

## 6. Core Features & Capabilities

### Feature 1: One-Command Multi-Chip Deployment

```bash
$ fleetml deploy model.onnx --fleet production

✓ Analyzing model architecture...              2s
✓ Compiling for NVIDIA Jetson (TensorRT)...   12s
✓ Compiling for Intel NUC (OpenVINO)...        8s
✓ Compiling for Raspberry Pi (TFLite)...       6s
✓ Compiling for Qualcomm (SNPE)...             9s

✓ Rolling out to 247 devices...
  ├── 198 NVIDIA Jetson    [████████████████████] 100%
  ├── 32 Intel NUC         [████████████████████] 100%
  ├── 12 Raspberry Pi      [████████████████████] 100%
  └── 5 offline            [queued - deploy on reconnect]

✓ 242/247 online, 5 queued (offline)

Deployment complete in 47s.
Monitor: https://app.fleetml.io/deploy/287
```

**Today's alternative:** Engineer manually converts model for each chip, writes custom deployment scripts per device type, SSHes into devices. Takes **2–6 WEEKS**. FleetML does it in **MINUTES**.

### Feature 2: Fleet Dashboard

```
┌─────────────────────────────────────────────────────────────┐
│  FleetML Dashboard                              ☰  Settings │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  FLEET OVERVIEW                            [Deploy Model ▼] │
│  ─────────────────────                                      │
│  🟢 192 Healthy   🟡 47 Warning   🔴 8 Offline             │
│                                                             │
│  MODEL PERFORMANCE                                          │
│  ─────────────────                                          │
│  Active Model: face-detect-v3.2                             │
│  Accuracy:     94.2% ▼0.3% this week                       │
│  Latency:      P50: 18ms  P95: 42ms  P99: 89ms            │
│  Throughput:   1.2M inferences/day                         │
│  ⚠️ 3 devices showing accuracy drift                       │
│                                                             │
│  RECENT DEPLOYMENTS                                         │
│  ─────────────────                                          │
│  ✓ face-detect-v3.2  → production    2h ago    242 devices │
│  ✓ plate-read-v2.1   → staging       1d ago     12 devices │
│  ✗ face-detect-v3.3  → canary        3d ago   ROLLED BACK  │
│                                                             │
│  DEVICE GROUPS                                              │
│  ─────────────                                              │
│  production-eu    │ 156 devices │ face-detect-v3.2          │
│  production-us    │  86 devices │ face-detect-v3.2          │
│  staging          │  12 devices │ face-detect-v3.3          │
│  development      │   5 devices │ face-detect-v4.0-beta     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Feature 3: A/B Testing at the Edge

```bash
$ fleetml ab-test \
    --model-a face-detect-v3.2 \
    --model-b face-detect-v4.0 \
    --split 80/20 \
    --fleet production \
    --metric accuracy,latency \
    --duration 7d \
    --auto-promote

Running A/B test across 247 devices...
  Model A (v3.2): 198 devices (80%)
  Model B (v4.0):  49 devices (20%)

Results in 7 days → auto-promote winner if statistically significant
Monitor: https://app.fleetml.io/ab-test/42
```

### Feature 4: Drift Detection & Auto-Retrain

The FleetML agent monitors inference quality locally:

- **Input distribution changes**: lighting shifts, new object types, camera angle changes
- **Confidence score degradation**: model becomes less certain over time
- **Accuracy drop**: when ground truth labels are available

When drift is detected:
1. Alert sent to dashboard
2. Edge samples collected (privacy-safe sampling)
3. Retraining triggered (cloud or on-prem)
4. New model deployed via FleetML pipeline
5. All automatic — no human intervention needed

### Feature 5: Offline-First Operation

Edge devices lose connectivity. FleetML handles this gracefully:

- Agent operates **100% independently** when offline
- Model inference continues uninterrupted
- Model updates **queued** and applied on reconnect
- Health metrics **stored locally**, synced when back online
- No model downtime during network outages
- Configurable: "deploy update on next maintenance window"

### Feature 6: Policy Engine

Define fleet-wide rules declaratively:

```yaml
# fleetml-policy.yaml
policies:
  - name: "EU GDPR Compliance"
    match:
      labels:
        region: eu
    rules:
      - model_version: ">=3.0"  # Must run GDPR-compliant model
      - data_export: false       # No edge data leaves EU devices

  - name: "Resource-Aware Deployment"
    match:
      hardware:
        ram_mb: "<4096"
    rules:
      - model_format: "int8"    # Use quantized model for low-RAM devices

  - name: "Canary Deployment"
    deploy:
      stages:
        - percent: 5
          duration: "1h"
          success_metric: "accuracy > 0.90"
        - percent: 50
          duration: "6h"
          success_metric: "accuracy > 0.92"
        - percent: 100

  - name: "Auto-Rollback"
    rollback:
      trigger:
        - latency_p99: ">100ms"
        - accuracy: "<0.85"
        - error_rate: ">5%"
```

### Feature 7: Multi-Chip Auto-Compiler

```
                    FLEETML COMPILER PIPELINE

    ┌──────────┐
    │  ONNX    │ ──── Universal input format
    │  Model   │
    └────┬─────┘
         │
         ▼
    ┌────────────────────────────────────────────────┐
    │           FleetML Compilation Engine            │
    │                                                 │
    │  ┌──────────┐ ┌──────────┐ ┌──────────┐       │
    │  │ Optimize │ │ Quantize │ │ Validate │       │
    │  │ graph    │ │ INT8/FP16│ │ accuracy │       │
    │  └──────────┘ └──────────┘ └──────────┘       │
    └──────┬──────────┬──────────┬──────────┬────────┘
           │          │          │          │
           ▼          ▼          ▼          ▼
    ┌──────────┐ ┌─────────┐ ┌────────┐ ┌────────┐
    │ TensorRT │ │OpenVINO │ │ TFLite │ │  SNPE  │
    │ (NVIDIA) │ │ (Intel) │ │ (ARM)  │ │(Qualc.)│
    │  .engine │ │  .xml   │ │ .tflite│ │  .dlc  │
    └──────────┘ └─────────┘ └────────┘ └────────┘
```

---

## 7. Technical Architecture

### Technology Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| **Agent** | Go | Small binary size (~10MB), excellent cross-compilation, fast startup, low memory |
| **Control Plane API** | Go + gRPC | Performance, type safety, native gRPC support |
| **Web Dashboard** | React + TypeScript + Tailwind | Modern, component-based, fast development |
| **CLI** | Go (also: Python wrapper via `pip install fleetml`) | Single binary distribution + Python ecosystem access |
| **Model Registry** | PostgreSQL + S3-compatible storage | Metadata in Postgres, model artifacts in S3/MinIO |
| **Message Bus** | NATS / MQTT | Lightweight, edge-friendly pub/sub |
| **Compiler Service** | Python + Docker containers | Each compiler (TensorRT, OpenVINO, etc.) runs in isolated container |
| **Monitoring** | Prometheus + Grafana (or built-in) | Industry standard, easy to integrate |
| **Database** | PostgreSQL | Reliable, well-understood, works at scale |
| **Container Orchestration** | Docker Compose (small) / Kubernetes (large) | Familiar to DevOps teams |

### Agent Communication Protocol

```
DEVICE → CONTROL PLANE (Heartbeat)
──────────────────────────────────
Every 30s (configurable):
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

CONTROL PLANE → DEVICE (Commands)
──────────────────────────────────
{
  "command": "deploy_model",
  "payload": {
    "model": "face-detect",
    "version": "4.0",
    "runtime": "tensorrt",
    "artifact_url": "s3://models/face-detect-v4.0-tensorrt.engine",
    "checksum": "sha256:abc123...",
    "rollback_version": "3.2",
    "deployment_policy": "canary_5_percent"
  }
}
```

### Offline Resilience Architecture

```
ONLINE MODE:
  Agent ←──gRPC──→ Control Plane (real-time sync)

OFFLINE MODE:
  Agent runs independently:
  ├── Inference continues with current model
  ├── Metrics stored in local SQLite
  ├── Commands queued in local NATS
  └── On reconnect: bulk sync (metrics up, commands down)

RECOVERY:
  1. Agent detects connectivity restored
  2. Uploads buffered metrics (compressed)
  3. Downloads queued commands
  4. Applies pending model updates
  5. Resumes real-time heartbeat
```

### Security Architecture

| Layer | Protection |
|-------|-----------|
| **Transport** | TLS 1.3 for all agent ↔ control plane communication |
| **Authentication** | mTLS (mutual TLS) — each agent has unique certificate |
| **Model Integrity** | SHA-256 checksums for all model artifacts |
| **API Access** | API keys + JWT tokens with RBAC |
| **Secrets** | Environment variables or Vault integration |
| **Audit** | All deployments, rollbacks, and access logged immutably |

---

## 8. Development Roadmap

### Phase 1: MVP — "Deploy with One Command" (Month 1–4)

**Goal:** Working open-source product that deploys ONNX models to a fleet of devices

| Week | Milestone | Deliverable |
|------|-----------|-------------|
| 1–2 | Agent v0.1 | Go agent that runs on Linux x86/ARM, receives model files, loads into ONNX Runtime |
| 3–4 | CLI v0.1 | `fleetml init`, `fleetml deploy`, `fleetml status`, `fleetml rollback` |
| 5–6 | Control Plane v0.1 | Docker-based server with model registry, device inventory, REST API |
| 7–8 | Web Dashboard v0.1 | Basic React dashboard: device list, model versions, deployment history |
| 9–10 | OTA Updates | Over-the-air model updates with checksum validation and rollback |
| 11–12 | Multi-device testing | Test with Jetson Orin Nano, Raspberry Pi 5, Intel NUC |
| 13–14 | Documentation + Demo | Full docs site, quickstart guide, 3 tutorial videos |
| 15–16 | **Open Source Launch** | GitHub release, Hacker News Show HN, Reddit launch |

**MVP Feature Set:**
- ✅ FleetML Agent (ONNX Runtime, Go)
- ✅ CLI (deploy, status, rollback, logs)
- ✅ Self-hosted control plane (Docker)
- ✅ Basic web dashboard
- ✅ Model versioning
- ✅ OTA updates
- ✅ Device health monitoring
- ❌ Multi-chip auto-compilation (manual ONNX for now)
- ❌ A/B testing
- ❌ Drift detection
- ❌ Policy engine

### Phase 2: Community Growth + Key Features (Month 5–9)

| Month | Feature | Impact |
|-------|---------|--------|
| 5 | **TensorRT compiler** | NVIDIA Jetson native performance (2–5x faster than ONNX) |
| 5 | **OpenVINO compiler** | Intel hardware native performance |
| 6 | **A/B testing** | Split traffic between model versions at edge |
| 6 | **Device grouping + labels** | Organize fleet by region, hardware, environment |
| 7 | **Drift detection** | Statistical monitoring of model performance degradation |
| 7 | **MLflow integration** | One-click: MLflow model → FleetML deployment |
| 8 | **Hugging Face integration** | Deploy any HF model to edge fleet |
| 8 | **TFLite compiler** | Raspberry Pi and ARM native performance |
| 9 | **Policy engine v1** | Canary deployments, auto-rollback rules |

### Phase 3: Monetization — FleetML Cloud (Month 9–15)

| Month | Milestone | Detail |
|-------|-----------|--------|
| 9–10 | FleetML Cloud beta | Hosted control plane, free tier (5 devices) |
| 10–11 | Team features | RBAC, deployment approvals, audit log |
| 11–12 | Advanced compiler | SNPE (Qualcomm), Hailo, quantization optimization |
| 12–13 | Enterprise features | SSO/SAML, SLA, dedicated support |
| 13–14 | First paying customers | Target: 20 paying customers |
| 14–15 | Partner integrations | NVIDIA Inception, Intel Edge, hardware co-marketing |

### Phase 4: Scale + Enterprise (Month 15–24)

| Month | Milestone | Detail |
|-------|-----------|--------|
| 15–18 | EU AI Act compliance layer | Audit trail, model documentation, risk classification |
| 18–20 | Federated learning | Train models across fleet without centralizing data |
| 20–22 | OEM/white-label | Other platforms embed FleetML |
| 22–24 | Enterprise sales | Dedicated sales team for 500+ device deployments |

### Long-Term Vision (Year 3+)

```
YEAR 1: FleetML Core (Open Source)
  └── "Deploy edge AI models with one command"

YEAR 2: FleetML Cloud (SaaS)
  └── "Manage your edge AI fleet from anywhere"

YEAR 3: FleetML Sovereign (Enterprise)
  └── "Compliant edge AI across jurisdictions"
      ├── EU AI Act audit trails
      ├── Data sovereignty enforcement
      ├── Federated learning across borders
      └── This is the Sovereign Edge AI Platform

ONE CODEBASE → THREE PRODUCTS → COMPOUND GROWTH
```

---

## 9. Go-To-Market Strategy

### GTM Philosophy: Developer-Led Growth (DLG)

```
FLEETML GTM IS COMPLETELY DIFFERENT FROM TRADITIONAL SAAS:

Traditional Enterprise SaaS:          FleetML (Developer-Led):
├── Hire sales team                   ├── Build great product
├── Buy LinkedIn Sales Navigator      ├── Open source it
├── Cold email 1000 CISOs             ├── Write tutorials
├── 4-12 week sales cycle             ├── Developer tries it in 5 min
├── Procurement, legal, security      ├── Works → tells manager
├── $50K+ deal size                   ├── Team adopts it
└── 12 months to first revenue        ├── Company pays for cloud/support
                                      └── 3-6 months to first revenue

DEVELOPERS DON'T NEED CERTIFICATIONS. THEY NEED:
✅ Great documentation
✅ Works in 5 minutes
✅ Open source (can inspect code)
✅ Active community
✅ Solves a real pain they feel daily
```

### Phase 1: Build in Public (Month 1–4)

**Strategy:** Build the product in front of an audience. Don't launch on Day 1 — build community interest BEFORE you have a product.

**Weekly "Build in Public" Content:**

| Week | Content | Platform |
|------|---------|----------|
| 1 | "I'm building an open-source edge MLOps platform. Here's why." | Blog + Twitter/X + LinkedIn |
| 2 | "Deploying YOLOv8 to 5 Jetsons — manual way vs FleetML" | YouTube + Blog |
| 3 | "How I built the FleetML agent in Go" | Hacker News + Dev.to |
| 4 | "FleetML v0.1 architecture deep-dive" | Blog + Twitter/X thread |
| 5 | "OTA model updates without SSH — how FleetML does it" | YouTube + Blog |
| 6 | "Why 70% of edge AI projects fail (and how to fix it)" | LinkedIn article |
| 7 | "FleetML devlog #7 — the offline-first challenge" | Twitter/X thread |
| 8+ | Continue weekly devlogs | Rotating platforms |

**Platform Priority:**
1. **GitHub** — The product lives here
2. **Twitter/X** — ML community is extremely active here
3. **LinkedIn** — Engineering managers and VPs
4. **Hacker News** — For launch moments
5. **Reddit** — r/MachineLearning, r/embedded, r/IOT, r/EdgeComputing
6. **YouTube** — Tutorials and demos
7. **Dev.to / Hashnode** — Long-form technical posts
8. **Discord** — Community hub

### Phase 2: Open Source Launch (Month 4)

**Launch Checklist:**

- [ ] Clean README with animated demo GIF
- [ ] "Get started in 5 minutes" quickstart
- [ ] Architecture diagram
- [ ] Supported hardware list
- [ ] Contributing guide (CONTRIBUTING.md)
- [ ] Apache 2.0 license
- [ ] CI/CD badges (builds passing, tests)
- [ ] Documentation site (fleetml.io/docs)
- [ ] 3 tutorial videos on YouTube
- [ ] Discord server with channels

**Launch Day Sequence:**

| Time | Action |
|------|--------|
| Morning | Push v0.1 tag to GitHub |
| Morning | Post "Show HN: FleetML — deploy edge AI models with one command" |
| Morning | Twitter/X thread with demo GIF |
| Midday | LinkedIn post (longer form, business angle) |
| Midday | Reddit posts (r/MachineLearning, r/embedded) |
| Evening | Dev.to article (tutorial format) |

**Target: 300–500 GitHub stars in first week**

### Phase 3: Community Growth (Month 4–9)

**Strategy 1: Integration Partnerships (Zero Cost)**

| Integration | Value Proposition |
|-------------|-------------------|
| **MLflow → FleetML** | "Train in MLflow, deploy to edge with FleetML" |
| **Hugging Face → FleetML** | "Pick any HF model, deploy to your edge fleet" |
| **Weights & Biases → FleetML** | "Track experiments in W&B, deploy winners to edge" |
| **Docker → FleetML** | FleetML agent runs as Docker container |
| **ONNX (native)** | Any ONNX model works out of the box |

Each integration = joint blog post + exposure to their community + "Works with FleetML" in their docs.

**Strategy 2: Hardware Vendor Co-Marketing (Zero Cost)**

| Vendor | Action | Benefit |
|--------|--------|---------|
| **NVIDIA** | "Deploy to Jetson Orin with FleetML" tutorial → submit to NVIDIA Developer Blog. Apply for NVIDIA Inception Program (free GPU credits + co-marketing) | They want more Jetson developers |
| **Intel** | "FleetML + OpenVINO: Edge AI deployment made easy" → Intel Edge AI community | They want more OpenVINO adoption |
| **Hailo** | First-class Hailo-8 support → Hailo actively promotes ecosystem tools | Hailo is new, hungry for ecosystem |
| **Raspberry Pi** | "Deploy ML to your Pi fleet with FleetML" tutorial → RPi community shares widely | RPi is #1 dev platform |

**Strategy 3: Content Marketing Engine**

**Tutorial Series: "Edge AI Deployment Cookbook"**

- "Deploy YOLOv8 to 10 Jetson devices in 5 minutes"
- "Object detection on Raspberry Pi 5 + Hailo at 30fps"
- "OTA model updates for your edge fleet"
- "A/B test ML models across your edge devices"
- "Monitor model drift on edge devices"
- "Deploy Whisper (speech-to-text) to edge"
- "Run Llama 3.2 on Jetson Orin with FleetML"
- "Multi-chip deployment: one model, five architectures"

Each tutorial = Blog post (SEO) + YouTube video (discovery) + GitHub repo (credibility) + Tweet thread (viral reach)

**Target Keywords (low competition, high intent):**

| Keyword | Monthly Searches | Competition |
|---------|-----------------|-------------|
| "deploy model to jetson" | ~500 | Very Low |
| "edge AI deployment tool" | ~300 | Low |
| "OTA update ML model" | ~200 | Very Low |
| "manage edge AI devices" | ~400 | Low |
| "mlops edge devices" | ~350 | Low |
| "deploy onnx to raspberry pi" | ~250 | Very Low |
| "fleet management edge AI" | ~150 | Very Low |

These are TINY keywords today — easy to rank #1. As edge AI grows, FleetML owns these search terms.

**Strategy 4: Community Building**

**Discord Server Structure:**
- `#general` — Chat
- `#help` — Support questions
- `#showcase` — Users share their deployments
- `#feature-requests` — Community-driven roadmap
- `#hardware` — Device-specific discussion (jetson, rpi, intel, hailo)
- `#contributors` — OSS contribution coordination

**Weekly Community Call (30 min):**
- Demo new features
- User showcase (someone demos their edge AI setup)
- Q&A

**"FleetML Champions" Program:**
- Top contributors get early access to features
- Name in CONTRIBUTORS.md
- FleetML swag (stickers, t-shirts)
- They become evangelists inside their companies

**Strategy 5: Conference Talks (Zero to Low Cost)**

| Event | Type | Cost |
|-------|------|------|
| MLOps Community meetups | Virtual, global | Free |
| PyData meetups | Every major city | Free |
| NVIDIA GTC | Online, free to present | Free |
| KubeCon + CloudNativeCon | If accepted, expenses covered | Free |
| Edge AI Summit | Apply as speaker | Free |
| Local ML meetups (Bangalore, Pune, Mumbai) | In-person | ₹1–5K travel |

**Talk Titles:**
- "Why 70% of Edge AI Projects Fail (and How to Fix It)"
- "From Jupyter Notebook to 500 Edge Devices in 5 Minutes"
- "The Missing Layer in Edge MLOps"
- "I Deployed AI to 100 Raspberry Pis — Here's What I Learned"

### Phase 4: First Revenue (Month 9–15)

**Signals That You're Ready to Monetize:**
- [ ] 1,000+ GitHub stars
- [ ] 100+ active deployments (self-hosted)
- [ ] 5+ companies using in production
- [ ] Users ASKING for hosted version
- [ ] Users ASKING for premium features
- [ ] At least 3 companies managing 50+ devices

**Monetization Trigger: Users saying "I'd pay for..."**
- Multi-chip auto-compilation
- Hosted dashboard (don't want to self-host)
- Team collaboration / RBAC
- Priority support
- SLA guarantees

---

## 10. Business Model & Pricing

### Pricing Tiers

| Tier | Price | Target | What's Included |
|------|-------|--------|-----------------|
| **Open Source** | Free forever | Individual developers, prototyping | Agent + CLI + self-hosted server. Up to 10 devices. Community support. Manual compilation |
| **Cloud Free** | $0/month | Startups evaluating | Cloud dashboard. Up to 5 devices. Basic monitoring |
| **Team** | $29/device/month | Small teams (10–100 devices) | Multi-chip auto-compiler. A/B testing. Drift detection. Email support |
| **Scale** | $19/device/month | Growing teams (100–500 devices) | Everything in Team + Policy engine. Team roles. Priority support |
| **Enterprise** | Custom ($10–15/device/month) | Large deployments (500+ devices) | Self-hosted option. SSO/SAML. SLA. Dedicated support. Custom integrations. EU AI Act audit trail |
| **OEM** | Custom royalty | Platform companies | White-label FleetML into their product. Per-device royalty or annual platform license |

### Revenue Composition (Target at Scale)

| Revenue Source | % of Total | Example |
|----------------|-----------|---------|
| SaaS subscriptions (Team + Scale) | 50% | 100 customers × 50 devices × $24 avg = $120K MRR |
| Enterprise contracts | 30% | 10 contracts × $100K/year = $1M ARR |
| OEM/white-label | 15% | 3 partners × $200K/year = $600K ARR |
| Professional services | 5% | Custom integrations, training |

### Unit Economics

| Metric | Target |
|--------|--------|
| **Customer Acquisition Cost (CAC)** | $500–1,500 (organic/content-driven) |
| **Average Revenue Per Account (ARPA)** | $1,500–3,000/month |
| **Lifetime Value (LTV)** | $36,000–72,000 (24-month avg life) |
| **LTV:CAC Ratio** | 24:1 to 48:1 |
| **Gross Margin** | 80%+ (SaaS) |
| **Payback Period** | <2 months |

---

## 11. Financial Projections

### Year 1: Investment Phase (₹0 Revenue)

| Item | Monthly | Annual |
|------|---------|--------|
| Cloud hosting (control plane) | ₹3,000 | ₹36,000 |
| Domain + website | ₹500 | ₹6,000 |
| GitHub Pro | ₹800 | ₹9,600 |
| Dev hardware (one-time) | — | ₹50,000 |
| Docs/website hosting | ₹0 | ₹0 |
| Design (logo, website) | — | ₹15,000 |
| Stickers/swag | — | ₹5,000 |
| Conference travel (India) | — | ₹15,000 |
| Miscellaneous | ₹1,000 | ₹12,000 |
| **TOTAL YEAR 1** | | **~₹1.5 lakh** |

**Milestones:** 1,000+ GitHub stars, 100+ active deployments, 5–10 production users

### Year 2: First Revenue ($450K ARR)

| Metric | Projection |
|--------|-----------|
| Paying customers | 50 |
| Average devices per customer | 30 |
| Average price per device/month | $25 |
| MRR | $37,500 |
| **ARR** | **$450,000 (~₹3.7 Cr)** |
| Infrastructure cost | ₹15 lakh/year |
| Team (2–3 engineers) | ₹25 lakh/year |
| **Net margin** | **~60%** |

### Year 3: Scale ($4M ARR)

| Metric | Projection |
|--------|-----------|
| Paying customers | 200 |
| Average devices per customer | 75 |
| Average price per device/month | $22 |
| MRR | $330,000 |
| **ARR** | **$4M (~₹33 Cr)** |
| Enterprise contracts (3–5) | $500K–$1M additional |
| Team (8–12 people) | ₹1 Cr/year |
| **Net margin** | **~50%** |

### Year 5: Market Leadership ($15–25M ARR)

| Metric | Projection |
|--------|-----------|
| Paying customers | 500+ |
| Total managed devices | 100,000+ |
| Revenue mix | 50% SaaS + 30% Enterprise + 20% OEM |
| **ARR** | **$15–25M** |
| Team | 30–50 people |
| **Valuation range** | **$150–250M (10x ARR)** |

---

## 12. Team & Hiring Plan

### Founder (Month 1–6): Solo

The founder wears all hats initially:
- Architecture & core agent development (Go)
- CLI & control plane development
- Content marketing (blog posts, tutorials)
- Community management (Discord, GitHub)
- Hardware testing (Jetson, RPi)

### Phase 1 Team (Month 6–12): 3 People

| Role | Focus | Cost (India) |
|------|-------|-------------|
| Founder | Architecture, strategy, community, sales | — |
| Backend Engineer | Control plane, API, compiler service | ₹8–15L/year |
| Frontend Engineer | Dashboard, developer experience | ₹6–12L/year |

### Phase 2 Team (Month 12–18): 5–7 People

| Role | Focus | Cost |
|------|-------|------|
| +DevRel / Community Manager | Content, community, talks, partnerships | ₹6–10L/year |
| +Edge Platform Engineer | Agent optimization, hardware support | ₹10–18L/year |
| +ML Compiler Engineer | Multi-chip compilation optimization | ₹10–18L/year |

### Phase 3 Team (Month 18–24): 8–12 People

Add: Sales Engineer, Customer Success, Additional Backend/Frontend engineers

### Key Hiring Principles

1. **India-first**: All engineering can be done from India at 70–80% lower cost than US/EU
2. **Open source contributors first**: Hire from the FleetML community — they already know and love the product
3. **Generalists over specialists**: Small team needs people who can wear multiple hats
4. **Remote-first**: Access talent across India (Bangalore, Pune, Hyderabad, Delhi NCR)

---

## 13. Risk Analysis & Mitigation

### Risk Matrix

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| **Cloud vendor builds this** (AWS/Azure/GCP add edge MLOps) | High | High | Open source + chip-neutral positioning. Cloud vendors will ALWAYS favor their own chips and services. FleetML is the neutral alternative |
| **NVIDIA builds native fleet management** | Medium | High | NVIDIA will always be NVIDIA-first. FleetML works with ALL chips. Multi-vendor customers need neutral platform |
| **Slow open source adoption** | Medium | Medium | "Build in public" creates audience before product launches. Hardware vendor co-marketing provides distribution. Focus on top-3 use cases with killer tutorials |
| **Monetization takes longer than expected** | High | Medium | Keep burn rate extremely low (₹1.5L/year Year 1). Can sustain for years without revenue. Consulting/services as bridge revenue if needed |
| **Security vulnerability in agent** | Low | High | Security-first architecture (mTLS, signed artifacts). Active vulnerability disclosure program. Regular security audits |
| **Hardware fragmentation too complex** | Medium | Medium | Start with top 3 platforms (Jetson, RPi, Intel NUC). Add hardware support based on community demand. Community contributors add hardware support |
| **Key person risk (founder)** | High | High | Document everything. Build community of contributors. Open source means project survives any individual |
| **Competitor raises $50M+** | Low | Medium | Open source moat — hard to compete against free. Community lock-in. India cost structure = sustainable at small scale |

### Why "Big Company Builds This" is Not a Death Sentence

When AWS launched IoT Greengrass, it didn't kill balena (now worth $300M+).

Why?
- AWS is always AWS-locked → customers with multi-cloud or on-prem needs can't use it
- Enterprise tools are complex → FleetML is simple enough for a 5-person team
- Open source creates switching costs → once you build on FleetML, you don't want to rewrite

The same logic applies to FleetML vs any cloud vendor edge tool.

---

## 14. Success Metrics & KPIs

### Product Metrics

| Metric | Month 4 (Launch) | Month 9 | Month 15 | Month 24 |
|--------|------------------|---------|----------|----------|
| GitHub stars | 300 | 1,500 | 5,000 | 15,000 |
| Active deployments | 20 | 150 | 500 | 2,000 |
| Devices under management | 100 | 1,000 | 5,000 | 25,000 |
| Community members (Discord) | 50 | 300 | 1,000 | 3,000 |
| Contributors | 5 | 20 | 50 | 100 |
| Supported hardware platforms | 3 | 5 | 8 | 12 |
| Integrations | 2 | 5 | 10 | 15 |

### Business Metrics

| Metric | Month 9 | Month 15 | Month 24 |
|--------|---------|----------|----------|
| Paying customers | 0 | 20 | 50+ |
| MRR | $0 | $15,000 | $37,500+ |
| ARR | $0 | $180,000 | $450,000+ |
| Trial → Paid conversion | — | 15% | 25% |
| Monthly churn | — | <5% | <3% |
| NPS | — | >50 | >60 |
| CAC | — | $1,000 | $500 |
| LTV:CAC | — | 10x | 24x |

### Content & Community Metrics

| Metric | Month 4 | Month 9 | Month 15 |
|--------|---------|---------|----------|
| Blog posts published | 12 | 30 | 50 |
| YouTube videos | 5 | 15 | 30 |
| Website visitors/month | 2,000 | 8,000 | 20,000 |
| Newsletter subscribers | 200 | 1,000 | 3,000 |
| Conference talks given | 2 | 5 | 10 |
| Twitter/X followers | 500 | 2,000 | 5,000 |

---

## 15. Week-by-Week Execution Plan

### Week 1: Foundation

- [ ] Register fleetml.io domain
- [ ] Create GitHub org: github.com/fleetml
- [ ] Set up project structure (monorepo)
- [ ] Initialize Go module for agent
- [ ] Write first "Building FleetML" blog post
- [ ] Post on Twitter/X: "I'm building an open-source edge MLOps platform"
- [ ] Apply for NVIDIA Inception Program
- [ ] Order hardware: 1x Jetson Orin Nano, 2x RPi 5

### Week 2: Agent Core

- [ ] FleetML agent: basic process lifecycle (start, stop, health check)
- [ ] ONNX Runtime integration (load and run model)
- [ ] Device fingerprinting (detect hardware, GPU, memory)
- [ ] Post devlog #2 on Twitter/X

### Week 3: Agent Communication

- [ ] gRPC client in agent (connect to control plane)
- [ ] Heartbeat protocol (device → server every 30s)
- [ ] Model download from URL with checksum validation
- [ ] Blog post: "Why I chose Go for an edge AI agent"

### Week 4: CLI v0.1

- [ ] `fleetml init` — Initialize a fleet project
- [ ] `fleetml deploy <model>` — Push model to devices
- [ ] `fleetml status` — Show fleet health
- [ ] `fleetml logs <device>` — Stream device logs
- [ ] Python wrapper: `pip install fleetml`

### Week 5–6: Control Plane v0.1

- [ ] Go server with REST API
- [ ] PostgreSQL for device registry + model metadata
- [ ] S3/MinIO for model artifact storage
- [ ] Docker Compose for easy self-hosting
- [ ] Device registration endpoint
- [ ] Model upload + versioning

### Week 7–8: Dashboard v0.1

- [ ] React + TypeScript + Tailwind project setup
- [ ] Device list view (status, hardware, current model)
- [ ] Model registry view (versions, deployment history)
- [ ] Deployment view (progress, rollback button)
- [ ] YouTube video: "FleetML dashboard walkthrough"

### Week 9–10: OTA Updates

- [ ] Server-side deployment orchestration
- [ ] Agent-side model hot-swap (zero-downtime update)
- [ ] Rollback mechanism (revert to previous version)
- [ ] Deployment progress tracking
- [ ] Blog post: "OTA model updates without SSH"

### Week 11–12: Multi-Device Testing

- [ ] Test full flow on Jetson Orin Nano
- [ ] Test full flow on Raspberry Pi 5
- [ ] Test full flow on Intel NUC (or x86 Linux box)
- [ ] Performance benchmarking: deployment time, inference latency
- [ ] Fix cross-platform issues
- [ ] YouTube video: "Deploying YOLOv8 to 3 different devices with FleetML"

### Week 13–14: Documentation

- [ ] Documentation site (using Docusaurus or MkDocs)
- [ ] Installation guide (per device type)
- [ ] Quickstart tutorial (5-minute guide)
- [ ] CLI reference
- [ ] API reference
- [ ] Architecture guide
- [ ] Contributing guide
- [ ] 3 end-to-end tutorial videos

### Week 15–16: OPEN SOURCE LAUNCH 🚀

- [ ] Clean up code, add tests, add CI/CD
- [ ] Write comprehensive README with GIF
- [ ] Create CHANGELOG, CODE_OF_CONDUCT, LICENSE
- [ ] Set up Discord server
- [ ] Launch on Hacker News (Show HN)
- [ ] Launch on Reddit (r/MachineLearning, r/embedded)
- [ ] Launch on Twitter/X (thread with demo)
- [ ] Launch on LinkedIn (business angle)
- [ ] Launch on Dev.to (tutorial format)
- [ ] Email list announcement
- [ ] **Target: 300+ GitHub stars in first week**

### Month 5 Onwards: See Development Roadmap (Section 8)

---

## Appendix A: Comparison With Alternatives

### FleetML vs Manual Deployment

| Capability | Manual | FleetML |
|-----------|--------|---------|
| Deploy to 100 devices | 2–6 weeks | 5 minutes |
| Multi-chip compilation | Per-chip toolchain knowledge | One command |
| Rollback bad deployment | SSH into each device | One command |
| Monitor fleet health | SSH + custom scripts | Real-time dashboard |
| A/B test models | Not feasible | Built-in |
| Drift detection | Not done | Automatic |
| Offline handling | Devices break | Seamless |

### FleetML vs Cloud MLOps (MLflow/SageMaker)

| Capability | MLflow/SageMaker | FleetML |
|-----------|-----------------|---------|
| Train models | ✅ | ❌ (Use MLflow/SageMaker) |
| Track experiments | ✅ | ❌ (Use MLflow/W&B) |
| Deploy to cloud | ✅ | ❌ (Not our focus) |
| **Deploy to edge** | **❌** | **✅** |
| **Fleet management** | **❌** | **✅** |
| **OTA updates** | **❌** | **✅** |
| **Multi-chip compilation** | **❌** | **✅** |
| **Offline-first** | **❌** | **✅** |
| **Edge drift detection** | **❌** | **✅** |

FleetML is **complementary** to cloud MLOps, not competitive. Train in MLflow → Deploy with FleetML.

### FleetML vs balena

| Capability | balena | FleetML |
|-----------|--------|---------|
| Container deployment to fleet | ✅ | ✅ |
| Device management | ✅ | ✅ |
| OTA updates | ✅ | ✅ |
| **ML model versioning** | **❌** | **✅** |
| **Multi-chip ML compilation** | **❌** | **✅** |
| **A/B testing models** | **❌** | **✅** |
| **Drift detection** | **❌** | **✅** |
| **Inference monitoring** | **❌** | **✅** |
| **ML-specific policies** | **❌** | **✅** |

balena is great for general IoT fleet management. FleetML is **ML-aware** — it understands models, not just containers.

---

## Appendix B: Use Case Examples

### Use Case 1: Retail Analytics

A retail chain deploys computer vision cameras across 200 stores for customer counting, heatmapping, and shelf analytics. Each store has 3–5 edge devices (Jetson Orin Nano). With FleetML, they deploy updated models to all 800 devices in minutes, A/B test new models at 10% of stores first, and auto-rollback if accuracy drops.

### Use Case 2: Manufacturing Quality Inspection

A factory runs visual inspection AI on 50 production lines, each with a different camera angle and lighting condition. Models drift as products change. FleetML detects drift, triggers retraining, and deploys updated models — all without stopping production.

### Use Case 3: Smart Agriculture

A precision farming company deploys crop disease detection models on drones and field cameras across 1,000 farms. Devices are often offline (rural connectivity). FleetML's offline-first architecture ensures models run regardless of connectivity, with updates applied on next sync.

### Use Case 4: Autonomous Vehicles / Robotics

A robotics company manages a fleet of 500 delivery robots, each running perception, navigation, and planning models. FleetML handles model versioning across the fleet, canary deployments of new perception models, and real-time monitoring of model performance across all robots.

### Use Case 5: Privacy-Compliant Video Surveillance

A security company deploys face anonymization models across customer sites. FleetML deploys models, ensures compliance model versions are running in EU locations (via policy engine), and provides audit trails for regulatory inspections. *This is the EdgeGuard vertical — proving FleetML in action.*

---

## Appendix C: Technology Decisions & Rationale

| Decision | Choice | Why | Alternatives Considered |
|----------|--------|-----|----------------------|
| Agent language | Go | Small binaries, cross-compilation, fast, mature ecosystem | Rust (steeper learning curve, slower iteration), Python (too heavy for edge), C++ (development speed) |
| Control plane | Go + PostgreSQL | Performance, reliability, ecosystem | Node.js (less suitable for systems), Python (GIL limitations) |
| Communication | gRPC + MQTT fallback | gRPC for normal operations (efficient, typed), MQTT for extremely constrained devices | REST (too chatty for edge), WebSocket (less typed) |
| Model format | ONNX as universal input | Industry standard, every framework exports to ONNX | Direct PyTorch/TF (too many formats) |
| License | Apache 2.0 | Most permissive, enterprise-friendly, allows commercial use | MIT (less explicit patent grant), AGPL (scares enterprises), BSL (limits adoption) |
| Dashboard | React + TypeScript | Fast development, large talent pool, component ecosystem | Vue (smaller ecosystem), Svelte (smaller talent pool) |

---

## Appendix D: Glossary

| Term | Definition |
|------|-----------|
| **Edge AI** | Running AI/ML models on local devices (not in the cloud) for real-time inference |
| **OTA** | Over-The-Air — updating software/models on devices remotely without physical access |
| **Fleet** | A group of edge devices managed as a unit |
| **ONNX** | Open Neural Network Exchange — universal model format supported by all major ML frameworks |
| **TensorRT** | NVIDIA's inference optimization library for NVIDIA GPUs |
| **OpenVINO** | Intel's inference optimization toolkit for Intel hardware |
| **TFLite** | TensorFlow Lite — Google's lightweight inference runtime for mobile and ARM devices |
| **SNPE** | Snapdragon Neural Processing Engine — Qualcomm's inference runtime |
| **Model Drift** | Gradual degradation of model performance due to changing input data distributions |
| **Canary Deployment** | Rolling out changes to a small subset of devices before full fleet deployment |
| **A/B Testing** | Running two model versions simultaneously to compare performance |
| **DLG** | Developer-Led Growth — go-to-market strategy where developers adopt bottom-up |
| **ARR** | Annual Recurring Revenue |
| **MRR** | Monthly Recurring Revenue |

---

*This document is a living plan. Updated as FleetML evolves from idea to product to platform.*

**Next Steps:**
1. Execute Week 1 action items (domain, GitHub, hardware order)
2. Begin agent development in Go
3. Start "Build in Public" content on Twitter/X and LinkedIn
4. Apply for NVIDIA Inception Program

---

**FleetML — Deploy edge AI models with one command.**

*github.com/fleetml | fleetml.io | @fleetml*
