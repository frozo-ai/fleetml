# FleetML Installation Guide

## Control Plane

### Docker Compose (Recommended)

```bash
git clone https://github.com/fleetml/fleetml.git
cd fleetml
cp .env.example .env

# Edit .env with your secrets
vim .env

docker compose up -d
```

### From Source

```bash
# Prerequisites
# - Go 1.22+
# - Node.js 20+
# - PostgreSQL 16
# - MinIO (or S3-compatible storage)

make build
make migrate-up

# Start server
./bin/server

# Start dashboard (development)
cd dashboard && npm run dev
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | | PostgreSQL connection string |
| `S3_ENDPOINT` | `localhost:9000` | MinIO/S3 endpoint |
| `S3_ACCESS_KEY` | | S3 access key |
| `S3_SECRET_KEY` | | S3 secret key |
| `S3_BUCKET` | `fleetml-models` | Model storage bucket |
| `JWT_SECRET` | | JWT signing secret (min 32 chars) |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

## Edge Agent

### Binary Install

```bash
# Download for your platform
curl -fsSL https://github.com/fleetml/fleetml/releases/latest/download/agent-linux-amd64 -o /usr/local/bin/fleetml-agent
chmod +x /usr/local/bin/fleetml-agent

# Configure
export FLEETML_SERVER=your-server:50051
export DEVICE_ID=$(hostname)

# Run
fleetml-agent
```

### Docker

```bash
docker run -d \
  --name fleetml-agent \
  --restart unless-stopped \
  -e FLEETML_SERVER=your-server:50051 \
  -e DEVICE_ID=$(hostname) \
  -v /var/lib/fleetml:/data \
  fleetml/agent:latest
```

### systemd Service

```ini
# /etc/systemd/system/fleetml-agent.service
[Unit]
Description=FleetML Edge Agent
After=network.target

[Service]
Type=simple
Environment=FLEETML_SERVER=your-server:50051
Environment=DEVICE_ID=%H
ExecStart=/usr/local/bin/fleetml-agent
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable --now fleetml-agent
```

### Agent Configuration

Config file: `/etc/fleetml/agent.yaml` or `~/.fleetml/agent.yaml`

```yaml
server_address: "your-server:50051"
device_id: "my-device-001"
heartbeat_interval: 30s
model_storage_dir: /var/lib/fleetml/models
max_model_versions: 3
log_level: info
```

Environment variables override config file values:
- `FLEETML_SERVER` overrides `server_address`
- `DEVICE_ID` overrides `device_id`
- `FLEETML_MODE` sets mode (`production`, `virtual`)

## CLI

### Install

```bash
# From release
curl -fsSL https://github.com/fleetml/fleetml/releases/latest/download/fleetml-$(uname -s)-$(uname -m) -o /usr/local/bin/fleetml
chmod +x /usr/local/bin/fleetml

# From source
make cli
sudo mv bin/fleetml /usr/local/bin/
```

### Configure

```bash
fleetml init --server http://your-server:8080
```

## Supported Platforms

### Agent

| Platform | Architecture | Status |
|----------|-------------|--------|
| Linux | amd64 | Supported |
| Linux | arm64 | Supported |
| Linux | armv7 | Planned |
| macOS | amd64 | Development |
| macOS | arm64 | Development |

### Server

| Platform | Architecture | Status |
|----------|-------------|--------|
| Linux | amd64 | Supported |
| Docker | any | Supported |
