# FleetML Quickstart

Get from zero to deploying an ONNX model in under 5 minutes.

## Prerequisites

- Docker & Docker Compose
- Go 1.22+ (for building from source)
- An ONNX model file

## 1. Start the Control Plane

```bash
git clone https://github.com/fleetml/fleetml.git
cd fleetml
cp .env.example .env
docker compose up -d
```

This starts:
- **Server** — REST API on `:8080`, gRPC on `:50051`
- **Dashboard** — Web UI on `:3000`
- **PostgreSQL** — Database on `:5432`
- **MinIO** — Model storage on `:9000` (console `:9001`)

## 2. Install the CLI

```bash
# From source
make cli
sudo mv bin/fleetml /usr/local/bin/

# Or download a release
curl -fsSL https://github.com/fleetml/fleetml/releases/latest/download/fleetml-$(uname -s)-$(uname -m) -o /usr/local/bin/fleetml
chmod +x /usr/local/bin/fleetml
```

## 3. Initialize the CLI

```bash
fleetml init --server http://localhost:8080
```

## 4. Deploy a Model

```bash
# Upload and deploy in one command
fleetml deploy my-model.onnx --fleet default --wait

# Or step by step:
fleetml deploy my-model.onnx --name my-model --version 1.0 --fleet default
fleetml status
```

## 5. Start an Edge Agent

On your edge device:

```bash
# Download the agent binary
curl -fsSL https://github.com/fleetml/fleetml/releases/latest/download/agent-linux-$(uname -m) -o /usr/local/bin/fleetml-agent
chmod +x /usr/local/bin/fleetml-agent

# Start the agent
FLEETML_SERVER=your-server:50051 fleetml-agent
```

Or use Docker:

```bash
docker run -d \
  -e FLEETML_SERVER=your-server:50051 \
  fleetml/agent:latest
```

## 6. View the Dashboard

Open `http://localhost:3000` to see:
- Fleet overview with device health
- Model registry
- Deployment progress
- Device metrics

## Next Steps

- [Installation Guide](installation.md) — Production setup
- [Architecture](architecture.md) — System design
- [CLI Reference](cli-reference.md) — All commands
- [API Reference](api-reference.md) — REST API docs
