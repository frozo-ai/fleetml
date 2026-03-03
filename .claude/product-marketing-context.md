# Product Marketing Context

*Last updated: March 2, 2026*

## Product Overview
**One-liner:** Kubernetes for edge AI models — deploy, update, and monitor ML models across thousands of heterogeneous edge devices from a single command.
**What it does:** FleetML is an open-source, chip-neutral edge MLOps platform that bridges the gap between a data scientist's notebook and a fleet of edge devices. One command deploys an ONNX model to hundreds of devices running different chips (NVIDIA, Intel, Qualcomm, Hailo, ARM), handling compilation, OTA updates, A/B testing, drift detection, and fleet monitoring — all offline-first.
**Product category:** Edge MLOps Platform / Edge AI Deployment Tool / IoT Fleet Management for ML
**Product type:** Open-source core (Apache 2.0) with cloud SaaS and enterprise tiers
**Business model:** Open-core. Free self-hosted OSS (up to 10 devices) → Cloud Free (5 devices) → Team ($29/device/month) → Scale ($19/device/month, 100+ devices) → Enterprise (custom) → OEM (white-label royalty)

## Target Audience
**Target companies:** Companies deploying AI/ML models to edge devices — manufacturing (quality inspection), smart retail (computer vision), autonomous vehicles/drones, smart agriculture, smart cities, defense/security. Company size: 10-person startups to 500+ enterprises. Stage: post-model-training, struggling with deployment at scale.
**Decision-makers:** ML Engineers, MLOps Engineers, Embedded/IoT Engineers, VP of Engineering, CTO, DevOps Engineers managing edge fleets
**Primary use case:** Deploying trained ML models (computer vision, anomaly detection, speech) to fleets of heterogeneous edge devices and managing them in production
**Jobs to be done:**
- Deploy a trained model to my entire edge fleet without SSH-ing into every device
- Update models across hundreds of devices with zero downtime and automatic rollback
- Monitor model performance, drift, and device health across my fleet from one dashboard
**Use cases:**
- Factory quality inspection: Deploy YOLOv8 defect detection to 50 Jetson devices on assembly lines
- Smart retail: Deploy customer analytics models to 200 Raspberry Pi cameras across stores
- Autonomous drones: Update navigation models OTA on a fleet of 100 drones with mixed chips
- Agriculture: Deploy crop disease detection to 500 solar-powered edge devices in remote fields (offline-first critical)

## Personas
| Persona | Cares about | Challenge | Value we promise |
|---------|-------------|-----------|------------------|
| **ML Engineer** (User) | Getting models into production fast, model accuracy in the wild | Trained a great model, stuck deploying it to actual hardware. Spends weeks per device type | "One command deploys to your entire fleet. Works on any chip." |
| **DevOps/MLOps Engineer** (Champion) | Reliability, monitoring, rollback, CI/CD for models | Managing heterogeneous fleet is manual, no unified tooling | "Fleet-wide OTA, canary deployments, auto-rollback, unified dashboard" |
| **VP Engineering / CTO** (Decision Maker) | Time-to-production, team velocity, vendor independence | 70% of edge AI projects stall in pilot, massive deployment gap | "Go from pilot to production in days, not months. No vendor lock-in." |
| **Embedded Engineer** (Technical Influencer) | Binary size, memory, performance, chip compatibility | Every chip has a different SDK, fragmented tooling | "15MB agent, 30MB RAM, compiles for your chip automatically" |

## Problems & Pain Points
**Core problem:** Deploying AI models at the edge is the #1 blocker preventing organizations from moving from pilot to production. Existing MLOps tools (MLflow, SageMaker, Vertex AI) stop at the cloud boundary — they give you a packaged model but do NOT deploy it to real edge hardware.
**Why alternatives fall short:**
- Cloud MLOps (MLflow, SageMaker) have zero edge deployment capability — they end at "here's your model file"
- Cloud vendor solutions (AWS Greengrass, Azure IoT Edge) lock you into one cloud and are not ML-focused
- Edge Impulse was chip-neutral but got acquired by Qualcomm (2024) — now Qualcomm-first
- OctoML was absorbed into NVIDIA (2023) — dead as independent
- Balena/ZEDEDA manage edge devices but have no ML awareness (no model versioning, A/B testing, drift detection)
**What it costs them:** Engineers spend 6-12 weeks per deployment manually SSH-ing into devices, writing custom scripts per chip type, building their own OTA update system. Fewer than 1/3 of organizations have fully deployed edge AI. 70% of Industry 4.0 projects stall in pilot.
**Emotional tension:** Frustration ("I built an amazing model and it's stuck in my notebook"), fear of bricking remote devices with bad updates, anxiety about managing heterogeneous hardware at scale, pressure from leadership to show production results.

## Competitive Landscape
**Direct:** Edge Impulse (acquired by Qualcomm, 2024) — was the closest competitor but now Qualcomm-first, no longer chip-neutral. FleetML fills the exact vacuum they left.
**Direct:** Latent AI ($30M+ raised) — enterprise/defense-only, not open-source, not developer-friendly, not fleet orchestration.
**Secondary:** AWS IoT Greengrass / Azure IoT Edge — edge runtimes but cloud-locked, enterprise-complex, not ML-focused. Developers want open, not locked.
**Secondary:** balena / ZEDEDA — IoT fleet management but zero ML awareness. No model versioning, no drift detection, no A/B testing, no compilation.
**Indirect:** MLflow / Kubeflow — cloud MLOps, not edge. Edge is an afterthought. No fleet management, no OTA, no offline-first.
**Indirect:** "SSH + bash scripts" — the current manual approach. Works for 5 devices, breaks at 50.

## Differentiation
**Key differentiators:**
- **Chip-neutral**: Works with NVIDIA (TensorRT), Intel (OpenVINO), ARM (TFLite), Qualcomm (SNPE), Hailo — every hardware vendor benefits from promoting us
- **Open-source core**: Apache 2.0, no vendor lock-in. Developers can inspect, modify, contribute
- **Offline-first**: Edge devices operate independently when disconnected — store-and-forward, local queue, auto-sync on reconnect
- **One command deployment**: `fleetml deploy model.onnx --fleet production` → deploys to hundreds of heterogeneous devices
- **Multi-chip auto-compilation**: ONNX in → TensorRT/OpenVINO/TFLite/SNPE/Hailo out, automatically matched to device hardware
- **Lightweight agent**: <15MB binary, <30MB RAM, runs on resource-constrained devices
**How we do it differently:** FleetML lives at the intersection of "chip-neutral" and "edge-native" — a quadrant that is completely empty. Every competitor is either cloud-dependent or chip-locked.
**Why that's better:** Engineers stop maintaining per-chip deployment scripts. One workflow for all hardware. Models ship in minutes, not weeks.
**Why customers choose us:** Open source (can inspect code, no lock-in), works in 5 minutes (not 5 weeks), supports their specific hardware mix, offline-first (critical for real edge deployments).

## Objections
| Objection | Response |
|-----------|----------|
| "AWS Greengrass already does this" | Greengrass is cloud-locked (AWS only), not ML-focused, and requires significant AWS expertise. FleetML is chip-neutral, ML-first, and works with any infrastructure. |
| "We'll build it ourselves" | Most teams try this and spend 6-12 weeks building a fragile custom solution. FleetML is production-tested OSS you can deploy today and customize. |
| "Open source = no support" | FleetML Cloud and Enterprise tiers include SLA-backed support. The open-source core means you're never locked in. |
| "It's a new/small project" | We're open source — you can read every line of code. Apache 2.0 means you're never dependent on us. Community contributors add resilience. |

**Anti-persona:** Companies with a single device type running models exclusively on cloud GPUs. If you don't have edge devices or offline requirements, you don't need FleetML — use MLflow or SageMaker.

## Switching Dynamics
**Push:** "I spent 3 weeks deploying a model to 10 Jetsons manually. Now they want 200 more. I can't do this." / "Every device has a different SDK. I'm writing custom scripts for each chip." / "My devices go offline in the field and we lose everything."
**Pull:** "One command to deploy to my whole fleet." / "Works on all my chips." / "Offline-first means my remote devices actually work." / "Open source so I'm not locked in." / "Community is active and helpful."
**Habit:** "We've always SSH'd into devices." / "Our custom scripts work for now." / "We're already invested in AWS ecosystem." / "We've been doing it manually and it's 'fine'."
**Anxiety:** "What if the agent crashes and bricks our device?" (Answer: automatic rollback, keep previous model version) / "What if the project gets abandoned?" (Answer: Apache 2.0, community-driven, self-hostable) / "What if it can't handle our scale?" (Answer: tested at 1000 virtual devices, used by X companies at Y scale)

## Customer Language
**How they describe the problem:**
- "How do I get this model to 500 devices in 3 countries on 4 different chips?"
- "MLOps tools stop at the cloud boundary"
- "Deploying AI at the edge is still manual and painful"
- "70% of our edge AI projects stall in pilot because deployment is so hard"
- "Every chip has a different SDK and the fragmentation is killing us"
- "Our devices go offline for days — we need something that works without internet"
**How they describe us:**
- "Kubernetes for edge AI"
- "Like Balena but for ML models"
- "The missing layer between my notebook and my edge fleet"
- "One command to deploy everywhere"
- "Finally something chip-neutral"
**Words to use:** deploy, fleet, edge, devices, one command, chip-neutral, open source, offline-first, OTA update, monitor, rollback, zero downtime
**Words to avoid:** "cloud-native" (we're edge-native), "enterprise-grade" (we're developer-first), "AI/ML platform" (too generic), "IoT platform" (we're ML-specific, not general IoT), "solution" (developers hate this word)
**Glossary:**
| Term | Meaning |
|------|---------|
| Fleet | A group of edge devices managed together |
| Agent | The lightweight FleetML binary running on each edge device (~15MB) |
| Control Plane | The server that manages the fleet (REST API + gRPC + Dashboard) |
| OTA | Over-The-Air model update — push a new model version to devices remotely |
| Hot-swap | Replacing a running model with a new version with zero dropped inferences |
| Canary deployment | Gradually rolling out a model update (5% → 50% → 100%) with automatic rollback |
| Drift detection | Detecting when input data distribution shifts and model accuracy degrades |
| Store-and-forward | Buffering data locally when offline, syncing when reconnected |
| Hardware fingerprint | Auto-detected device capabilities (CPU, GPU, accelerators, RAM) that determine which compiled model variant to use |

## Brand Voice
**Tone:** Technical but approachable. Confident but not arrogant. Opinionated but open.
**Style:** Direct, developer-to-developer. Show code examples, not marketing slides. Prove claims with benchmarks, not buzzwords. Use "you" not "users" or "customers."
**Personality:** Builder, pragmatic, open, chip-agnostic evangelist, community-first

## Proof Points
**Metrics:**
- <15MB agent binary (stripped)
- <30MB RAM at idle
- <2 second agent startup
- Deploy to 100 devices in <2 minutes
- Deploy to 1000 devices in <10 minutes
- Zero dropped inferences during model hot-swap
- 10K heartbeats/second throughput
- p99 API latency <100ms
**Customers:** Pre-launch (building community and production users through OSS adoption)
**Testimonials:** Pre-launch (collecting from beta/alpha users)
**Value themes:**
| Theme | Proof |
|-------|-------|
| Speed | "6 minutes vs 6 weeks" — deploy to fleet in one command instead of weeks of manual work |
| Chip-neutrality | Works with NVIDIA, Intel, Qualcomm, Hailo, ARM — every vendor promotes us |
| Reliability | Offline-first, auto-rollback, canary deployments, zero-downtime hot-swap |
| Simplicity | `fleetml deploy model.onnx --fleet production` — that's it |
| Openness | Apache 2.0, no vendor lock-in, community-driven |
| Cost | India cost structure: profitable at $500K ARR where US competitors need $5M. Year 1 GTM cost: ~₹1.5 lakh |

## Goals
**Business goal:** Establish FleetML as the de facto open-source edge MLOps platform. Target: 1,000+ GitHub stars, 100+ active deployments, 5+ production users by end of Year 1. First revenue (Month 9-15) at $450K ARR Year 2.
**Conversion action:** Star the GitHub repo → Try quickstart (deploy model to one device in 5 min) → Deploy to fleet → Self-host in production → Upgrade to FleetML Cloud/Team tier
**Current metrics:** Pre-development phase. Tracking to begin with OSS launch (Month 4).
