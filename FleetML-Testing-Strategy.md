# FleetML — Complete Testing Strategy

## How to Test an Edge MLOps Platform Without 500 Devices

**Version:** 1.0  |  **Date:** February 22, 2026  |  **Companion to:** FleetML Master Plan v1.0

---

## Table of Contents

1. [Testing Philosophy](#1-testing-philosophy)
2. [Testing Pyramid](#2-testing-pyramid-for-fleetml)
3. [Unit Tests](#3-layer-1--unit-tests)
4. [Integration Tests](#4-layer-2--integration-tests)
5. [Virtual Fleet Simulator](#5-layer-3--virtual-fleet-simulator)
6. [Hardware-in-the-Loop](#6-layer-4--hardware-in-the-loop-testing)
7. [Chaos Engineering](#7-layer-5--chaos-engineering)
8. [Performance & Scale Testing](#8-layer-6--performance--scale-testing)
9. [ML Model-Specific Testing](#9-layer-7--ml-model-specific-testing)
10. [Security Testing](#10-layer-8--security-testing)
11. [End-to-End Scenarios](#11-layer-9--end-to-end-e2e-scenarios)
12. [CI/CD Pipeline Design](#12-cicd-pipeline-design)
13. [Testing Infrastructure & Budget](#13-testing-infrastructure--budget)
14. [Test Execution Timeline](#14-test-execution-timeline)
15. [Quality Gates & Release Criteria](#15-quality-gates--release-criteria)

---

## 1. Testing Philosophy

### The Core Challenge

FleetML operates across **three very different environments simultaneously**:

```
┌───────────────────────────────────────────────────────────┐
│                                                           │
│  CONTROL PLANE           NETWORK              EDGE        │
│  (Cloud/Server)          (Unreliable)         (Device)    │
│                                                           │
│  ┌──────────┐       ┌──────────────┐     ┌──────────┐   │
│  │ Standard │ ←───→ │ Lossy, slow, │ ←─→ │ Resource │   │
│  │ server   │       │ drops,       │     │ limited, │   │
│  │ reliable │       │ offline      │     │ varied   │   │
│  │ infra    │       │              │     │ hardware │   │
│  └──────────┘       └──────────────┘     └──────────┘   │
│                                                           │
│  ✅ Easy to test     ⚠️ Hard to test   ❌ Hardest       │
│                                                           │
└───────────────────────────────────────────────────────────┘
```

Traditional SaaS testing only covers the control plane. FleetML must test **ALL THREE** — including the messy, unreliable middle.

### Testing Principles

1. **Test without hardware first** — 90% of FleetML can be tested on a laptop using virtual devices
2. **Simulate the fleet, test the edge** — Use Docker containers as virtual edge devices
3. **Chaos is not optional** — Network failures, device crashes, partial deployments are NORMAL at the edge
4. **Every merge to main runs the full suite** — Automated CI catches regressions before humans
5. **Hardware tests are the LAST step** — Real devices validate what virtual tests already proved
6. **Test the sad paths more than happy paths** — What happens when 40% of your fleet goes offline during an OTA update?

### What Makes Edge Testing Different

| Cloud SaaS Testing | Edge / FleetML Testing |
|--------------------|-----------------------|
| Server is always online | Devices go offline randomly |
| Homogeneous infrastructure | 5+ chip architectures |
| Network is fast and reliable | Network is slow, lossy, intermittent |
| One deployment target | Hundreds of heterogeneous targets |
| Rollback = redeploy container | Rollback = push update to offline devices |
| Logs are centralized | Logs are on-device, may be unreachable |
| State is in database | State split across devices + server |

---

## 2. Testing Pyramid for FleetML

```
                          ╱╲
                         ╱  ╲
                        ╱ E2E╲          ← 5-10 scenarios
                       ╱ Real ╲           Real hardware
                      ╱Hardware╲          Runs: Pre-release
                     ╱──────────╲
                    ╱   Chaos    ╲       ← 20-30 experiments
                   ╱  Engineering ╲        Network/device faults
                  ╱────────────────╲       Runs: Nightly
                 ╱   Virtual Fleet  ╲    ← 50+ scenarios
                ╱     Simulator      ╲     100-1000 virtual devices
               ╱──────────────────────╲    Runs: Every PR
              ╱      Integration       ╲  ← 100+ tests
             ╱     Tests (Docker)       ╲   Agent ↔ Server
            ╱────────────────────────────╲  Runs: Every PR
           ╱          Unit Tests          ╲ ← 500+ tests
          ╱     Agent, CLI, Server, API    ╲  Pure functions
         ╱──────────────────────────────────╲ Runs: Every commit
```

| Layer | Tests | Run Time | Runs When | Cost |
|-------|-------|----------|-----------|------|
| Unit Tests | 500+ | 30 seconds | Every commit | ₹0 |
| Integration | 100+ | 3-5 minutes | Every PR | ₹0 |
| Virtual Fleet | 50+ | 10-15 minutes | Every PR | ₹0 |
| Chaos Engineering | 20-30 | 30-60 minutes | Nightly | ₹0 |
| Hardware-in-Loop | 5-10 | 15-30 minutes | Pre-release | ₹50K one-time |
| E2E Real Devices | 5-10 | 30-60 minutes | Pre-release | Same hardware |

---

## 3. Layer 1 — Unit Tests

Pure function tests. No network, no Docker, no hardware. Run in milliseconds.

### Agent Unit Tests (Go)

```go
// agent/model/loader_test.go

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
```

```go
// agent/deploy/rollback_test.go

func TestRollbackManager_KeepsLastN(t *testing.T) {
    rm := NewRollbackManager("/tmp/rollback", 3)
    rm.SaveVersion("v1", modelBytesV1)
    rm.SaveVersion("v2", modelBytesV2)
    rm.SaveVersion("v3", modelBytesV3)
    rm.SaveVersion("v4", modelBytesV4)
    assert.False(t, rm.HasVersion("v1")) // v1 evicted
    assert.True(t, rm.HasVersion("v4"))
}
```

```go
// server/policy/engine_test.go

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

### Unit Test Coverage Map

| Component | Tests | What's Tested |
|-----------|-------|--------------|
| Agent — Model Loader | 15+ | Load, validate, corrupt, format detection |
| Agent — Health Reporter | 10+ | System metrics, GPU detection, edge cases |
| Agent — Rollback Manager | 10+ | Version storage, eviction, restore |
| Agent — Heartbeat Protocol | 10+ | Serialization, compression, timing |
| Agent — Drift Detector | 15+ | PSI/KS tests, threshold triggers |
| Server — Fleet Manager | 20+ | Registration, grouping, selection |
| Server — Model Registry | 15+ | Upload, versioning, metadata |
| Server — Policy Engine | 15+ | Rule evaluation, rollback triggers, canary logic |
| Server — Deployment Orchestrator | 20+ | Scheduling, progress, failure handling |
| CLI — Argument Parsing | 10+ | Validation, defaults, error messages |
| **Total** | **~150+** | |

---

## 4. Layer 2 — Integration Tests

Test component interactions using Docker. No real hardware needed.

### Docker Compose Test Environment

```yaml
# docker-compose.test.yml
services:
  control-plane:
    build: ./server
    ports: ["8080:8080", "50051:50051"]
    environment:
      - DATABASE_URL=postgres://postgres:test@db:5432/fleetml_test
      - S3_ENDPOINT=http://minio:9000
    depends_on: [db, minio]

  db:
    image: postgres:16
    environment:
      POSTGRES_DB: fleetml_test
      POSTGRES_PASSWORD: test

  minio:
    image: minio/minio
    command: server /data

  agent-1:
    build: ./agent
    environment:
      - FLEETML_SERVER=control-plane:50051
      - DEVICE_ID=test-agent-1
```

### Key Integration Tests

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

    // Deploy v1 (good), then v2 (corrupted)
    v1, _ := ctx.API.UploadModel("model-v1.onnx", goodModelBytes)
    ctx.API.Deploy(v1, "test-agent-1")
    waitForDeployment(t, ctx, "test-agent-1", v1)

    v2, _ := ctx.API.UploadModel("model-v2.onnx", corruptModelBytes)
    ctx.API.Deploy(v2, "test-agent-1")

    // Agent should auto-rollback to v1
    waitFor(t, 30*time.Second, func() bool {
        device, _ := ctx.API.GetDevice("test-agent-1")
        return device.ActiveModel.ID == v1
    })
}

func TestOTAUpdate_ZeroDowntime(t *testing.T) {
    ctx := setupTestEnvironment(t)
    v1, _ := ctx.API.UploadModel("v1.onnx", modelV1Bytes)
    ctx.API.Deploy(v1, "test-agent-1")
    waitForDeployment(t, ctx, "test-agent-1", v1)

    // Start inference in background
    failures := atomic.Int64{}
    go func() {
        for i := 0; i < 100; i++ {
            _, err := ctx.Agent.Infer("test-agent-1", testImage)
            if err != nil { failures.Add(1) }
            time.Sleep(50 * time.Millisecond)
        }
    }()

    // Deploy v2 while inference is running
    v2, _ := ctx.API.UploadModel("v2.onnx", modelV2Bytes)
    ctx.API.Deploy(v2, "test-agent-1")
    waitForDeployment(t, ctx, "test-agent-1", v2)

    assert.Equal(t, int64(0), failures.Load()) // Zero dropped inferences
}
```

---

## 5. Layer 3 — Virtual Fleet Simulator

**The most critical and unique testing layer.** Simulates 10–1000 edge devices using Docker containers, no real hardware.

### Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    VIRTUAL FLEET SIMULATOR                    │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              FleetML Control Plane (Docker)           │   │
│  └──────────────────────┬───────────────────────────────┘   │
│                         │ gRPC (with simulated network)      │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐     ┌──────────┐  │
│  │ Virtual  │ │ Virtual  │ │ Virtual  │ ... │ Virtual  │  │
│  │ Device 1 │ │ Device 2 │ │ Device 3 │     │Device N  │  │
│  │ Jetson   │ │ RPi 5    │ │ Intel    │     │ Jetson   │  │
│  │ Orin     │ │ 4GB RAM  │ │ NUC i5   │     │ Nano     │  │
│  │ TensorRT │ │ TFLite   │ │ OpenVINO │     │ ONNX     │  │
│  └──────────┘ └──────────┘ └──────────┘     └──────────┘  │
│                                                              │
│  Each virtual device:                                        │
│  • Runs the REAL FleetML agent binary                       │
│  • Fakes hardware detection (reports Jetson but runs x86)   │
│  • Simulates inference metrics (latency, accuracy)          │
│  • Uses `tc netem` for real network condition simulation     │
│  • Can be killed (docker kill) = power failure              │
│  • Can be disconnected (docker network disconnect) = offline│
└──────────────────────────────────────────────────────────────┘
```

### Device Profile Simulation

```go
// simulator/profiles.go

var Profiles = map[string]DeviceProfile{
    "jetson-orin-nano": {
        Name: "NVIDIA Jetson Orin Nano", Arch: "arm64",
        RAMMB: 8192, GPUType: "nvidia", Runtime: "tensorrt",
        InferenceMs: 12.0, DiskGB: 64,
    },
    "rpi5-4gb": {
        Name: "Raspberry Pi 5 4GB", Arch: "arm64",
        RAMMB: 4096, GPUType: "none", Runtime: "tflite",
        InferenceMs: 95.0, DiskGB: 16,
    },
    "intel-nuc-i5": {
        Name: "Intel NUC i5", Arch: "amd64",
        RAMMB: 16384, GPUType: "intel", Runtime: "openvino",
        InferenceMs: 22.0, DiskGB: 256,
    },
    "rpi5-8gb": {
        Name: "Raspberry Pi 5 8GB", Arch: "arm64",
        RAMMB: 8192, GPUType: "none", Runtime: "tflite",
        InferenceMs: 85.0, DiskGB: 32,
    },
    "generic-x86": {
        Name: "Generic x86 Server", Arch: "amd64",
        RAMMB: 32768, GPUType: "none", Runtime: "onnx",
        InferenceMs: 55.0, DiskGB: 500,
    },
}
```

### Network Condition Simulation

```go
// simulator/network.go

var NetworkProfiles = map[string]NetworkProfile{
    "excellent":    {LatencyMs: 5,   JitterMs: 2,   PacketLossPct: 0.01, BandwidthKbps: 100000},
    "good-wifi":    {LatencyMs: 15,  JitterMs: 10,  PacketLossPct: 0.5,  BandwidthKbps: 50000},
    "poor-wifi":    {LatencyMs: 50,  JitterMs: 30,  PacketLossPct: 2.0,  BandwidthKbps: 10000},
    "cellular-4g":  {LatencyMs: 80,  JitterMs: 40,  PacketLossPct: 1.5,  BandwidthKbps: 20000},
    "cellular-3g":  {LatencyMs: 200, JitterMs: 100, PacketLossPct: 5.0,  BandwidthKbps: 2000},
    "satellite":    {LatencyMs: 600, JitterMs: 200, PacketLossPct: 3.0,  BandwidthKbps: 5000},
    "intermittent": {LatencyMs: 100, JitterMs: 50,  PacketLossPct: 10.0, BandwidthKbps: 5000,
                     DisconnectProbPerMin: 0.30},
}

// Applied using Linux tc (traffic control) inside each Docker container
func ApplyNetworkProfile(containerID string, p NetworkProfile) error {
    cmd := fmt.Sprintf(
        "tc qdisc add dev eth0 root netem delay %dms %dms loss %.2f%% rate %dkbit",
        p.LatencyMs, p.JitterMs, p.PacketLossPct, p.BandwidthKbps,
    )
    return dockerExec(containerID, cmd)
}
```

### Fleet Scenario Tests

```go
func TestFleet_DeployToHeterogeneousFleet(t *testing.T) {
    fleet := simulator.CreateFleet([]simulator.FleetSpec{
        {Count: 8, Profile: "jetson-orin-nano", Network: "good-wifi"},
        {Count: 6, Profile: "rpi5-4gb",         Network: "poor-wifi"},
        {Count: 4, Profile: "intel-nuc-i5",     Network: "excellent"},
        {Count: 2, Profile: "rpi5-8gb",         Network: "cellular-4g"},
    })
    defer fleet.Teardown()

    waitFor(t, 30*time.Second, func() bool { return fleet.OnlineCount() == 20 })
    deployID := fleet.Deploy("yolov8n.onnx", "all")
    waitFor(t, 120*time.Second, func() bool {
        return fleet.GetDeployment(deployID).Completed == 20
    })
    for _, device := range fleet.Devices() {
        assert.Equal(t, "yolov8n.onnx", device.ActiveModel())
    }
}

func TestFleet_DeployWithOfflineDevices(t *testing.T) {
    fleet := simulator.CreateFleet([]simulator.FleetSpec{
        {Count: 10, Profile: "jetson-orin-nano", Network: "good-wifi"},
    })
    defer fleet.Teardown()

    fleet.TakeOffline("device-3", "device-5", "device-8")
    deployID := fleet.Deploy("model-v2.onnx", "all")

    waitFor(t, 60*time.Second, func() bool {
        s := fleet.GetDeployment(deployID)
        return s.Completed == 7 && s.Queued == 3
    })

    fleet.BringOnline("device-3", "device-5", "device-8")
    waitFor(t, 60*time.Second, func() bool {
        return fleet.GetDeployment(deployID).Completed == 10
    })
}

func TestFleet_CanaryWithAutoRollback(t *testing.T) {
    fleet := simulator.CreateFleet([]simulator.FleetSpec{
        {Count: 100, Profile: "jetson-orin-nano", Network: "good-wifi"},
    })
    defer fleet.Teardown()

    fleet.Deploy("model-v1.onnx", "all")
    waitForFleetDeploy(t, fleet, 100)

    fleet.SetSimulatedAccuracy("model-v2-bad.onnx", 0.72) // Below 0.85 threshold
    deployID := fleet.DeployCanary("model-v2-bad.onnx", "all", CanaryPolicy{
        Stages: []CanaryStage{
            {Percent: 5, Duration: 30 * time.Second, SuccessMetric: "accuracy > 0.85"},
            {Percent: 50, Duration: 60 * time.Second},
            {Percent: 100},
        },
    })

    waitFor(t, 90*time.Second, func() bool {
        return fleet.GetDeployment(deployID).State == "rolled_back"
    })

    for _, device := range fleet.Devices() {
        assert.Equal(t, "model-v1.onnx", device.ActiveModel())
    }
}
```

### Virtual Device Dockerfile

```dockerfile
FROM golang:1.22-alpine AS builder
COPY agent/ /app/agent/
WORKDIR /app/agent
RUN CGO_ENABLED=0 go build -o /fleetml-agent ./cmd/agent

FROM alpine:3.19
RUN apk add --no-cache iproute2  # For tc (traffic control)
COPY --from=builder /fleetml-agent /usr/local/bin/fleetml-agent
ENV FLEETML_MODE=virtual
ENTRYPOINT ["fleetml-agent", "--virtual"]
```

```bash
# Spin up 100 virtual devices
docker compose -f docker-compose.fleet-sim.yml up --scale virtual-device=100
```

---

## 6. Layer 4 — Hardware-in-the-Loop Testing

Real hardware, real models, real inference. Validates that virtual tests translate to reality.

### Hardware Lab (₹50K budget)

```
┌──────────────────────────────────────────────────┐
│           FLEETML HARDWARE TEST LAB              │
│                                                  │
│  ┌──────────────┐  ┌──────────────┐             │
│  │ Jetson Orin  │  │ Raspberry    │             │
│  │ Nano (8GB)   │  │ Pi 5 (8GB)   │             │
│  │ ₹15,000      │  │ ₹5,000       │             │
│  └──────┬───────┘  └──────┬───────┘             │
│         └───────┬──────────┘                     │
│                 │ WiFi / Ethernet                 │
│                 ▼                                 │
│  ┌──────────────────────────────────┐           │
│  │ Developer Laptop / Desktop       │           │
│  │ (Control plane + x86 device +    │           │
│  │  CI runner for hardware tests)   │           │
│  └──────────────────────────────────┘           │
│                                                  │
│  Optional: RPi 4 (₹3K) + Intel NUC (₹15K)     │
│  TOTAL: ₹20K minimum, ₹50K ideal               │
└──────────────────────────────────────────────────┘
```

### Hardware Tests

```go
func TestHardware_JetsonDeployYOLOv8(t *testing.T) {
    if !isHardwareAvailable("jetson-orin-nano") {
        t.Skip("Jetson not available")
    }
    ctx := setupHardwareTest(t)
    deployID, _ := ctx.API.Deploy("yolov8n.onnx", "jetson-lab-1")
    waitForDeployment(t, ctx, deployID)

    result, err := ctx.Device.Infer("jetson-lab-1", testPersonImage)
    assert.NoError(t, err)
    assert.Greater(t, len(result.Detections), 0)
    assert.Less(t, result.LatencyMs, 50.0) // <50ms on Jetson Orin Nano
}

func TestHardware_RPiDeployTFLite(t *testing.T) {
    if !isHardwareAvailable("rpi5") { t.Skip("RPi 5 not available") }
    ctx := setupHardwareTest(t)
    ctx.API.Deploy("yolov8n.tflite", "rpi-lab-1")

    result, _ := ctx.Device.Infer("rpi-lab-1", testPersonImage)
    assert.Less(t, result.LatencyMs, 200.0) // RPi is slower
}

func TestHardware_ZeroDowntimeOTA(t *testing.T) {
    ctx := setupHardwareTest(t)
    ctx.API.Deploy("model-v1.onnx", "jetson-lab-1")

    failures := atomic.Int64{}
    ctx.RunContinuousInference("jetson-lab-1", &failures)

    ctx.API.Deploy("model-v2.onnx", "jetson-lab-1")
    waitForDeployment(t, ctx, "jetson-lab-1")
    time.Sleep(5 * time.Second)

    assert.Equal(t, int64(0), failures.Load()) // ZERO failures during hot-swap
}
```

### When to Run

- **NOT on every PR** — hardware is shared, tests are slow
- **Nightly:** Full suite against `main`
- **Pre-release:** Full suite + 24-hour soak test
- **After architecture changes:** Manual trigger

---

## 7. Layer 5 — Chaos Engineering

Deliberately break things to verify FleetML handles real-world failures gracefully.

### Fault Categories

| Category | Faults |
|----------|--------|
| **Network** | Complete disconnect, intermittent (50% loss), extreme latency (2-5s), bandwidth throttle (56kbps), DNS failure, TLS expiry, network partition |
| **Device** | Power failure (docker kill), disk full mid-download, OOM, CPU overload, clock drift, corrupted filesystem |
| **Deployment** | Partial deploy (server crash at 50%), corrupted model, wrong runtime for device, model OOM, concurrent deploys, rollback-during-rollback |
| **Control Plane** | Server restart mid-deploy, database failure, S3/MinIO failure, API rate limiting |

### Chaos Tests

```go
func TestChaos_NetworkPartition(t *testing.T) {
    fleet := simulator.CreateFleet([]simulator.FleetSpec{
        {Count: 20, Profile: "jetson-orin-nano", Network: "good-wifi"},
    })
    defer fleet.Teardown()

    fleet.Deploy("model-v1.onnx", "all")
    waitForFleetDeploy(t, fleet, 20)

    // CHAOS: Disconnect 8 random devices
    partitioned := fleet.RandomDevices(8)
    fleet.DisconnectNetwork(partitioned...)

    deployID := fleet.Deploy("model-v2.onnx", "all")
    waitFor(t, 60*time.Second, func() bool {
        s := fleet.GetDeployment(deployID)
        return s.Completed == 12 && s.Queued == 8
    })

    // Partitioned devices still running v1 (not crashed!)
    for _, d := range partitioned {
        assert.Equal(t, "healthy", fleet.Device(d).LocalStatus())
    }

    // Reconnect → auto-deploy pending updates
    fleet.ReconnectNetwork(partitioned...)
    waitFor(t, 120*time.Second, func() bool {
        return fleet.GetDeployment(deployID).Completed == 20
    })
}

func TestChaos_PowerFailureDuringOTA(t *testing.T) {
    fleet := simulator.CreateFleet([]simulator.FleetSpec{
        {Count: 1, Profile: "jetson-orin-nano", Network: "good-wifi"},
    })
    defer fleet.Teardown()

    fleet.Deploy("model-v1.onnx", "all")
    waitForFleetDeploy(t, fleet, 1)

    fleet.Deploy("model-v2.onnx", "all")
    time.Sleep(500 * time.Millisecond) // Mid-download

    fleet.Kill("device-0")     // No graceful shutdown
    time.Sleep(2 * time.Second)
    fleet.Restart("device-0")

    // Comes back on v1 (safe), then auto-retries v2
    waitFor(t, 30*time.Second, func() bool {
        return fleet.Device("device-0").Status() == "healthy"
    })
    assert.Equal(t, "model-v1.onnx", fleet.Device("device-0").ActiveModel())

    waitFor(t, 60*time.Second, func() bool {
        return fleet.Device("device-0").ActiveModel() == "model-v2.onnx"
    })
}

func TestChaos_ServerCrashMidDeploy(t *testing.T) {
    fleet := simulator.CreateFleet([]simulator.FleetSpec{
        {Count: 50, Profile: "jetson-orin-nano", Network: "good-wifi"},
    })
    defer fleet.Teardown()

    deployID := fleet.Deploy("model.onnx", "all")
    waitFor(t, 30*time.Second, func() bool {
        return fleet.GetDeployment(deployID).Completed >= 25
    })

    fleet.KillControlPlane()
    time.Sleep(5 * time.Second)

    // Already-deployed devices still running fine
    for _, d := range fleet.Devices() {
        if d.ActiveModel() == "model.onnx" {
            assert.Equal(t, "healthy", d.LocalStatus())
        }
    }

    fleet.RestartControlPlane()
    // Deployment RESUMES from where it left off
    waitFor(t, 120*time.Second, func() bool {
        return fleet.GetDeployment(deployID).Completed == 50
    })
}

func TestChaos_DiskFull(t *testing.T) {
    fleet := simulator.CreateFleet([]simulator.FleetSpec{
        {Count: 1, Profile: "rpi5-4gb", Network: "good-wifi"},
    })
    defer fleet.Teardown()

    fleet.FillDisk("device-0", 95)
    deployID := fleet.Deploy("large-model.onnx", "all")

    waitFor(t, 30*time.Second, func() bool {
        return fleet.GetDeployment(deployID).State == "failed"
    })
    assert.Equal(t, "healthy", fleet.Device("device-0").Status()) // Agent didn't crash
    assert.Contains(t, fleet.GetDeployment(deployID).Error, "insufficient disk space")
}
```

---

## 8. Layer 6 — Performance & Scale Testing

### Benchmark Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| API latency p99 | <100ms | k6 load test |
| Agent heartbeat CPU overhead | <1% | Real hardware profile |
| Agent memory footprint | <50MB RSS | Real hardware monitor |
| Agent binary size | <15MB | Build artifact |
| Deploy time (1 device) | <30 sec | Virtual fleet |
| Deploy time (100 devices) | <2 min | Virtual fleet |
| Deploy time (1000 devices) | <10 min | Virtual fleet |
| Heartbeat throughput (server) | 10K devices/sec | Load test |
| Dashboard load time | <2 sec | Lighthouse |
| Heartbeat message size | <1KB compressed | Protocol measure |
| Reconnection time | <10 sec | Chaos test |

### 1000-Device Scale Test

```go
func TestScale_1000Devices(t *testing.T) {
    if testing.Short() { t.Skip("Skipping scale test") }

    fleet := simulator.CreateFleet([]simulator.FleetSpec{
        {Count: 400, Profile: "jetson-orin-nano", Network: "good-wifi"},
        {Count: 300, Profile: "rpi5-4gb",         Network: "poor-wifi"},
        {Count: 200, Profile: "intel-nuc-i5",     Network: "excellent"},
        {Count: 100, Profile: "generic-x86",      Network: "cellular-4g"},
    })
    defer fleet.Teardown()

    start := time.Now()
    waitFor(t, 60*time.Second, func() bool { return fleet.OnlineCount() == 1000 })
    t.Logf("1000 devices registered in %v", time.Since(start))

    deployStart := time.Now()
    deployID := fleet.Deploy("model.onnx", "all")
    waitFor(t, 10*time.Minute, func() bool {
        return fleet.GetDeployment(deployID).Completed == 1000
    })
    t.Logf("1000-device deployment in %v", time.Since(deployStart))

    metrics := fleet.ControlPlaneMetrics()
    assert.Less(t, metrics.CPUPercent, 80.0)
    assert.Less(t, metrics.RAMMB, 2048)
}
```

### API Load Test (k6)

```javascript
// tests/load/k6-heartbeat.js
import http from 'k6/http';
import { check } from 'k6';

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

export default function () {
    const res = http.post('http://localhost:8080/api/v1/heartbeat',
        JSON.stringify({
            device_id: `device-${__VU}`,
            status: 'healthy',
            metrics: { cpu: Math.random() * 100, ram_mb: 2048 },
        }),
        { headers: { 'Content-Type': 'application/json' } }
    );
    check(res, { 'status 200': (r) => r.status === 200 });
}
```

---

## 9. Layer 7 — ML Model-Specific Testing

### Model Compatibility Matrix

```go
var testModels = []struct {
    Name   string; Format string; Size int64; File string
}{
    {"YOLOv8n",          "onnx",   12_800_000, "testdata/yolov8n.onnx"},
    {"YOLOv8s",          "onnx",   44_000_000, "testdata/yolov8s.onnx"},
    {"ResNet50",         "onnx",   97_000_000, "testdata/resnet50.onnx"},
    {"MobileNetV3",      "onnx",   22_000_000, "testdata/mobilenetv3.onnx"},
    {"EfficientNet-Lite","tflite", 18_000_000, "testdata/efficientnet.tflite"},
    {"Whisper-tiny",     "onnx",   76_000_000, "testdata/whisper-tiny.onnx"},
}

func TestModelFormats_LoadAndInfer(t *testing.T) {
    for _, model := range testModels {
        t.Run(model.Name, func(t *testing.T) {
            agent := setupTestAgent(t)
            err := agent.LoadModel(model.File)
            assert.NoError(t, err)
            output, err := agent.Infer(testInput)
            assert.NoError(t, err)
            assert.NotNil(t, output)
        })
    }
}
```

### Drift Detection Tests

```go
func TestDriftDetector_DetectsDistributionShift(t *testing.T) {
    detector := drift.NewDetector(drift.Config{
        Method: "psi", Threshold: 0.2, WindowSize: 1000,
    })

    // Feed baseline
    for i := 0; i < 1000; i++ { detector.AddBaseline(normalConfidenceScore()) }

    // Same distribution → no drift
    for i := 0; i < 1000; i++ { detector.AddObservation(normalConfidenceScore()) }
    assert.False(t, detector.IsDrifting())

    // Shifted distribution → drift detected
    for i := 0; i < 1000; i++ { detector.AddObservation(shiftedConfidenceScore()) }
    assert.True(t, detector.IsDrifting())
    assert.Greater(t, detector.DriftScore(), 0.2)
}
```

### Model Size Guardrails

```go
func TestModelSize_FitsDeviceRAM(t *testing.T) {
    tests := []struct {
        Model   string; ModelSizeMB int; DeviceRAMMB int; ShouldFit bool
    }{
        {"yolov8n.onnx", 12, 4096, true},     // 12MB on 4GB RPi → OK
        {"yolov8x.onnx", 260, 4096, false},   // 260MB on 4GB RPi → rejected
        {"yolov8x.onnx", 260, 8192, true},    // 260MB on 8GB Jetson → OK
        {"resnet152.onnx", 450, 4096, false},  // 450MB on 4GB → rejected
    }

    for _, tt := range tests {
        t.Run(fmt.Sprintf("%s_on_%dMB", tt.Model, tt.DeviceRAMMB), func(t *testing.T) {
            fits := agent.ModelFitsDevice(tt.ModelSizeMB, tt.DeviceRAMMB)
            assert.Equal(t, tt.ShouldFit, fits)
        })
    }
}
```

---

## 10. Layer 8 — Security Testing

### Security Test Matrix

| Test | What It Validates |
|------|------------------|
| mTLS enforcement | Agent without valid cert cannot connect |
| Model integrity | Tampered model (wrong checksum) rejected |
| API authentication | Unauthenticated calls return 401 |
| RBAC enforcement | Viewer role cannot trigger deployments |
| Command injection | Malicious device IDs don't execute shell |
| Replay attack | Old signed deploy commands can't be replayed |
| Agent impersonation | One agent can't report as another |
| Secrets in transit | No plaintext secrets in gRPC/HTTP |
| Dependency vulns | `govulncheck` + `npm audit` clean |

### Security Tests

```go
func TestSecurity_UnauthenticatedAgentRejected(t *testing.T) {
    ctx := setupTestEnvironment(t)
    conn, _ := grpc.Dial(ctx.ServerAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    client := pb.NewFleetMLClient(conn)

    _, err := client.Register(context.Background(), &pb.RegisterRequest{
        DeviceID: "rogue-device",
    })
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "authentication")
}

func TestSecurity_TamperedModelRejected(t *testing.T) {
    ctx := setupTestEnvironment(t)
    modelID, _ := ctx.API.UploadModel("model.onnx", legitimateModelBytes)

    ctx.TamperModelArtifact(modelID, append(legitimateModelBytes, byte(0xFF)))
    deployID, _ := ctx.API.Deploy(modelID, "test-agent-1")

    waitFor(t, 30*time.Second, func() bool {
        return ctx.API.GetDeployment(deployID).State == "failed"
    })
    assert.Contains(t, ctx.API.GetDeployment(deployID).Error, "checksum mismatch")
}

func TestSecurity_CommandInjectionPrevented(t *testing.T) {
    ctx := setupTestEnvironment(t)
    // Device ID with shell injection attempt
    _, err := ctx.API.RegisterDevice("device-$(rm -rf /)")
    assert.Error(t, err) // Should be rejected at validation layer

    _, err = ctx.API.RegisterDevice("device; cat /etc/passwd")
    assert.Error(t, err)
}
```

---

## 11. Layer 9 — End-to-End (E2E) Scenarios

Full user journey tests, from `pip install fleetml` to deployed model on real hardware.

### E2E Scenario 1: First-Time User Journey

```bash
#!/bin/bash
# tests/e2e/first_time_user.sh
set -e

echo "=== E2E: First-Time User ==="

# Install CLI
pip install fleetml
fleetml version | grep -q "v0.1"

# Start control plane
docker run -d --name fleetml-server -p 8080:8080 -p 50051:50051 fleetml/server:latest
sleep 5

# Register a virtual device
docker run -d --name test-device \
  -e FLEETML_SERVER=host.docker.internal:50051 \
  -e DEVICE_ID=my-first-device \
  fleetml/agent:latest
sleep 10

# Verify device shows up
fleetml status | grep -q "my-first-device"
fleetml status | grep -q "healthy"

# Deploy a model
fleetml deploy testdata/yolov8n.onnx --device my-first-device
sleep 15

# Verify deployment
fleetml status --device my-first-device | grep -q "yolov8n.onnx"
fleetml status --device my-first-device | grep -q "running"

# Rollback
fleetml rollback my-first-device
sleep 10
fleetml status --device my-first-device | grep -q "no model"

echo "=== PASSED ==="
docker rm -f fleetml-server test-device
```

### E2E Scenario 2: Multi-Device Fleet

Tests: deploy to 10 devices, A/B test, detect drift, auto-rollback across heterogeneous fleet

### E2E Scenario 3: Offline Resilience

Tests: deploy to device, disconnect it, verify inference continues for 24 hours, reconnect, verify metrics sync back

---

## 12. CI/CD Pipeline Design

```
┌───────────────────────────────────────────────────────────────┐
│                  CI/CD PIPELINE (GitHub Actions)              │
│                                                              │
│  ON EVERY COMMIT (~2 min):                                   │
│  ├── Lint (golangci-lint, eslint)                    30 sec  │
│  ├── Unit Tests (go test -short ./...)               45 sec  │
│  ├── Build Agent (cross-compile arm64+amd64)         60 sec  │
│  └── Build Server + Dashboard                        90 sec  │
│                                                              │
│  ON EVERY PR (~10 min):                                      │
│  ├── Integration Tests (Docker Compose)              3 min   │
│  ├── Virtual Fleet Tests (20 devices)                5 min   │
│  └── Security Scan (govulncheck + trivy)             2 min   │
│                                                              │
│  NIGHTLY (~60 min):                                          │
│  ├── Virtual Fleet at Scale (500 devices)            15 min  │
│  ├── Chaos Engineering Suite                         30 min  │
│  ├── Performance Benchmarks                          10 min  │
│  └── API Load Test (k6)                              8 min   │
│                                                              │
│  PRE-RELEASE (manual trigger):                               │
│  ├── Hardware-in-Loop Tests                          30 min  │
│  ├── E2E Scenarios (real hardware)                   45 min  │
│  └── 24-hour Soak Test                               24 hrs  │
│                                                              │
└───────────────────────────────────────────────────────────────┘
```

### GitHub Actions Config

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
      - name: Upload Benchmark Results
        uses: benchmark-action/github-action-benchmark@v1
        with:
          tool: 'go'
          output-file-path: benchmark-results.txt
```

---

## 13. Testing Infrastructure & Budget

### Cost Summary

| Item | One-Time Cost | Monthly Cost | Purpose |
|------|--------------|-------------|---------|
| Jetson Orin Nano (8GB) | ₹15,000 | ₹0 | TensorRT edge testing |
| Raspberry Pi 5 (8GB) | ₹5,000 | ₹0 | ARM/TFLite edge testing |
| SD Cards, cables, power | ₹3,000 | ₹0 | Lab accessories |
| GitHub Actions (free tier) | ₹0 | ₹0 | CI/CD for public repo |
| GitHub Actions (if private) | ₹0 | ~₹1,500 | 3000 min/month |
| Docker Hub (free tier) | ₹0 | ₹0 | Container registry |
| **TOTAL** | **₹23,000** | **₹0-1,500** | |

### Optional Upgrades (Phase 2)

| Item | Cost | When |
|------|------|------|
| Intel NUC (used) | ₹15,000 | When first OpenVINO customer |
| Raspberry Pi 4 (4GB) | ₹3,000 | For low-RAM testing |
| USB WiFi adapter (for network chaos on real hardware) | ₹1,000 | When testing real network failures |
| Cloud CI runners (larger) | ₹3,000/mo | When 1000-device tests exceed free tier |

### What Runs Where

| Test Layer | Where It Runs | Hardware Needed |
|-----------|--------------|-----------------|
| Unit Tests | GitHub Actions (free) | None |
| Integration | GitHub Actions (free) | None (Docker) |
| Virtual Fleet (20 devices) | GitHub Actions (free) | None (Docker) |
| Virtual Fleet (500+ devices) | GitHub Actions (large runner) or dev laptop | None (Docker) |
| Chaos Engineering | GitHub Actions or dev laptop | None (Docker) |
| API Load Test | GitHub Actions or dev laptop | None |
| Hardware-in-Loop | Self-hosted runner (dev laptop) | Jetson + RPi |
| E2E Real Hardware | Self-hosted runner (dev laptop) | Jetson + RPi |
| 24hr Soak Test | Dev laptop (overnight) | Jetson + RPi |

---

## 14. Test Execution Timeline

### Aligned with FleetML Development Phases

#### Week 1-2: Foundation (Agent + Server)
```
Tests to write:
├── Unit tests for agent model loader         (15 tests)
├── Unit tests for agent heartbeat            (10 tests)
├── Unit tests for server fleet manager       (20 tests)
├── Unit tests for server model registry      (15 tests)
├── Integration: agent registration           (3 tests)
├── Integration: model upload + deploy        (5 tests)
└── Docker Compose test environment           (infra)

Total: ~68 tests | CI pipeline: commit + PR gates
```

#### Week 3-4: Fleet Features
```
Tests to write:
├── Unit tests for deployment orchestrator    (20 tests)
├── Unit tests for policy engine              (15 tests)
├── Unit tests for rollback manager           (10 tests)
├── Virtual fleet simulator (core framework)  (infra)
├── Fleet tests: deploy to 20 devices         (5 tests)
├── Fleet tests: offline devices              (3 tests)
├── Fleet tests: canary deployment            (3 tests)
├── Integration: rollback flow                (5 tests)
└── Integration: OTA zero-downtime            (3 tests)

Total: ~64 tests | CI: nightly fleet tests added
```

#### Week 5-6: Chaos + Scale
```
Tests to write:
├── Chaos: network partition                  (3 tests)
├── Chaos: power failure during OTA           (2 tests)
├── Chaos: server crash mid-deploy            (2 tests)
├── Chaos: disk full                          (2 tests)
├── Scale: 100-device deploy                  (2 tests)
├── Scale: 500-device deploy                  (1 test)
├── Scale: 1000-device deploy                 (1 test)
├── API load test (k6)                        (1 suite)
├── Performance benchmarks                    (5 benchmarks)
└── Hardware-in-loop setup + first tests      (5 tests)

Total: ~24 tests + infra | CI: chaos + scale nightly
```

#### Week 7-8: Security + E2E
```
Tests to write:
├── Security: mTLS enforcement                (3 tests)
├── Security: model integrity                 (3 tests)
├── Security: API auth + RBAC                 (5 tests)
├── Security: command injection               (3 tests)
├── ML: drift detection                       (5 tests)
├── ML: model compatibility matrix            (6 tests)
├── E2E: first-time user journey              (1 scenario)
├── E2E: multi-device fleet                   (1 scenario)
├── E2E: offline resilience                   (1 scenario)
└── 24-hour soak test                         (1 scenario)

Total: ~28 tests | CI: full pipeline complete
```

### Cumulative Test Count

| End of Week | Unit | Integration | Fleet | Chaos | Scale | Security | E2E | **TOTAL** |
|-------------|------|-------------|-------|-------|-------|----------|-----|-----------|
| Week 2 | 60 | 8 | 0 | 0 | 0 | 0 | 0 | **68** |
| Week 4 | 105 | 16 | 11 | 0 | 0 | 0 | 0 | **132** |
| Week 6 | 115 | 16 | 11 | 9 | 4 | 0 | 0 | **155** |
| Week 8 | 130 | 16 | 11 | 9 | 4 | 14 | 3 | **187** |

---

## 15. Quality Gates & Release Criteria

### PR Merge Requirements (Every PR)

```
✅ All unit tests pass (500+ tests)
✅ All integration tests pass
✅ Virtual fleet tests pass (20 devices)
✅ No new security vulnerabilities (govulncheck)
✅ No lint errors
✅ Agent binary builds for amd64 + arm64
✅ Code coverage ≥ 80% on new code
```

### Release Candidate Criteria

```
✅ All PR requirements met
✅ Nightly fleet test (500 devices) passing for 3+ consecutive nights
✅ Chaos engineering suite: 100% pass rate
✅ Performance benchmarks: no regressions > 10%
✅ API load test: p99 < 200ms at 1000 concurrent connections
✅ Hardware-in-loop tests pass on Jetson + RPi
✅ E2E scenarios pass on real hardware
✅ 24-hour soak test: zero agent crashes, zero dropped deployments
✅ Security scan clean (govulncheck + trivy + npm audit)
```

### Go/No-Go Checklist for v0.1.0 Launch

```
MUST HAVE:
□ Agent deploys ONNX model on x86 and arm64             (integration test)
□ Agent survives power failure with model intact          (chaos test)
□ Agent survives network disconnect, reconnects           (chaos test)
□ OTA update completes with zero inference downtime       (hardware test)
□ Canary deployment detects bad model, auto-rolls back    (fleet test)
□ 100-device deployment completes in < 2 minutes          (scale test)
□ CLI install-to-first-deploy in < 5 minutes              (E2E test)
□ Unauthenticated agents are rejected                     (security test)
□ Tampered models are rejected                            (security test)

NICE TO HAVE:
□ 1000-device deployment completes in < 10 minutes
□ Drift detection triggers alert within 60 seconds
□ Dashboard loads in < 2 seconds
□ Agent memory < 50MB on Raspberry Pi
```

---

## Appendix A: Tools Reference

| Tool | Purpose | Used In |
|------|---------|---------|
| `go test` | Unit + integration tests | All Go tests |
| `Docker Compose` | Test environment orchestration | Integration, Fleet, Chaos |
| `tc netem` | Network condition simulation | Virtual fleet, Chaos |
| `docker kill/disconnect` | Device failure simulation | Chaos |
| `k6` | API load testing | Performance |
| `golangci-lint` | Go code quality | CI |
| `govulncheck` | Go vulnerability scan | Security |
| `trivy` | Container vulnerability scan | Security |
| `Lighthouse` | Dashboard performance | Performance |
| `github-action-benchmark` | Performance regression tracking | CI |

## Appendix B: Test File Structure

```
fleetml/
├── agent/
│   ├── model/loader_test.go
│   ├── health/reporter_test.go
│   ├── deploy/rollback_test.go
│   └── heartbeat/protocol_test.go
├── server/
│   ├── fleet/manager_test.go
│   ├── model/registry_test.go
│   ├── policy/engine_test.go
│   └── deploy/orchestrator_test.go
├── cli/
│   └── cmd/deploy_test.go
├── simulator/
│   ├── profiles.go
│   ├── network.go
│   ├── fleet.go
│   └── Dockerfile.virtual-device
├── tests/
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
├── docker-compose.test.yml
├── docker-compose.fleet-sim.yml
└── .github/workflows/ci.yml
```

---

*End of FleetML Testing Strategy v1.0*
*Total planned tests: 187+ across 9 layers*
*Zero hardware needed for 90% of testing*
