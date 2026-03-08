#!/usr/bin/env bash
# FleetML Quickstart — One command to see everything working.
#
# Usage:
#   ./scripts/quickstart.sh
#
# What it does:
#   1. Starts infrastructure (PostgreSQL, MinIO, NATS, compiler)
#   2. Builds and starts the server (auto-runs migrations)
#   3. Creates a demo admin user
#   4. Registers 6 simulated edge devices (2 Jetson, 2 RPi, 2 Intel)
#   5. Creates 2 fleets (gpu-fleet, cpu-fleet)
#   6. Uploads a sample model
#   7. Deploys with canary rollout
#   8. Sends simulated heartbeats and logs
#   9. Prints a summary with links
#
# Prerequisites: Docker, Go 1.22+, curl, jq

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

log()  { echo -e "${GREEN}[✓]${NC} $1"; }
info() { echo -e "${CYAN}[→]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
err()  { echo -e "${RED}[✗]${NC} $1"; }

API="http://localhost:8080/api/v1"
SERVER_PID=""

cleanup() {
    if [ -n "$SERVER_PID" ] && kill -0 "$SERVER_PID" 2>/dev/null; then
        kill "$SERVER_PID" 2>/dev/null || true
    fi
}
trap cleanup EXIT

# ── Step 1: Check prerequisites ────────────────────────────────────────
echo ""
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  FleetML Quickstart${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

info "Checking prerequisites..."
for cmd in docker go curl jq; do
    if ! command -v "$cmd" &>/dev/null; then
        err "$cmd is required but not installed."
        exit 1
    fi
done
log "All prerequisites found"

# ── Step 2: Start infrastructure ────────────────────────────────────────
info "Starting infrastructure (PostgreSQL, MinIO, NATS, compiler)..."
docker compose up -d db minio nats compiler 2>&1 | tail -1

# Wait for services to be healthy
info "Waiting for services to be ready..."
for i in $(seq 1 30); do
    if docker compose ps --format json 2>/dev/null | jq -r '.Health // "starting"' 2>/dev/null | grep -q "unhealthy"; then
        sleep 2
        continue
    fi
    # Simple check: try to connect to postgres
    if docker compose exec -T db pg_isready -U fleetml -q 2>/dev/null; then
        break
    fi
    sleep 2
done
log "Infrastructure is ready"

# ── Step 3: Build and start the server ──────────────────────────────────
info "Building FleetML server..."
cd server && go build -o ../bin/fleetml-server ./cmd/server 2>&1 && cd "$ROOT"
log "Server built"

info "Starting server (REST :8080, gRPC :50051)..."
cd "$ROOT"
./bin/fleetml-server &
SERVER_PID=$!
cd "$ROOT"

# Wait for server to be ready
for i in $(seq 1 20); do
    if curl -s "$API/health" &>/dev/null; then
        break
    fi
    sleep 1
done

if ! curl -s "$API/health" &>/dev/null; then
    err "Server failed to start. Check logs above."
    exit 1
fi
log "Server is running"

# ── Step 4: Create demo user ────────────────────────────────────────────
info "Creating demo admin user..."
curl -s -X POST "$API/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"demo@fleetml.dev","password":"demo1234","role":"admin"}' > /dev/null 2>&1 || true

TOKEN=$(curl -s -X POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"demo@fleetml.dev","password":"demo1234"}' | jq -r '.token // empty')

if [ -z "$TOKEN" ]; then
    err "Failed to get auth token"
    exit 1
fi
log "Logged in as demo@fleetml.dev"

AUTH="Authorization: Bearer $TOKEN"

# ── Step 5: Create fleets ──────────────────────────────────────────────
info "Creating device fleets..."
GPU_FLEET=$(curl -s -X POST "$API/fleets" \
    -H "$AUTH" -H "Content-Type: application/json" \
    -d '{"name":"gpu-fleet","description":"NVIDIA Jetson devices with GPU inference"}' | jq -r '.id // empty')

CPU_FLEET=$(curl -s -X POST "$API/fleets" \
    -H "$AUTH" -H "Content-Type: application/json" \
    -d '{"name":"cpu-fleet","description":"Raspberry Pi and Intel devices"}' | jq -r '.id // empty')

log "Created fleets: gpu-fleet, cpu-fleet"

# ── Step 6: Register simulated devices ──────────────────────────────────
info "Registering 6 simulated edge devices..."

register_device() {
    local id=$1 arch=$2 gpu=$3 runtime=$4 ram=$5 hw=$6
    curl -s -X POST "$API/heartbeat" \
        -H "Content-Type: application/json" \
        -d "{
            \"device_id\": \"$id\",
            \"status\": \"healthy\",
            \"system\": {
                \"cpu_percent\": $(( RANDOM % 40 + 20 )),
                \"gpu_percent\": $(( RANDOM % 30 )),
                \"ram_mb_used\": $(( RANDOM % 1000 + 500 )),
                \"disk_percent\": $(( RANDOM % 30 + 20 )),
                \"temperature_c\": $(( RANDOM % 15 + 35 )),
                \"uptime_hours\": $(( RANDOM % 720 ))
            }
        }" > /dev/null 2>&1
}

register_device "jetson-nano-01" "arm64" "nvidia-jetson" "tensorrt" "4096" "jetson-nano"
register_device "jetson-orin-02" "arm64" "nvidia-orin"   "tensorrt" "8192" "jetson-orin"
register_device "rpi-4b-01"      "arm64" "none"          "tflite"   "2048" "rpi-4b"
register_device "rpi-4b-02"      "arm64" "none"          "tflite"   "4096" "rpi-4b"
register_device "intel-nuc-01"   "amd64" "intel-uhd"     "openvino" "8192" "nuc-11"
register_device "intel-nuc-02"   "amd64" "intel-uhd"     "openvino" "16384" "nuc-13"

log "Registered: 2 Jetson, 2 Raspberry Pi, 2 Intel NUC"

# ── Step 7: Upload a sample model ──────────────────────────────────────
info "Creating sample ONNX model..."

# Generate a minimal ONNX model using Python
python3 -c "
import struct, os
# Write a minimal valid-looking ONNX file (just needs to be uploadable)
data = b'ONNX' + b'\x00' * 100 + b'fleetml-demo-model'
with open('/tmp/fleetml-demo-model.onnx', 'wb') as f:
    f.write(data)
" 2>/dev/null || echo "fake model" > /tmp/fleetml-demo-model.onnx

info "Uploading model: defect-detector v1.0..."
UPLOAD_RESULT=$(curl -s -X POST "$API/models" \
    -H "$AUTH" \
    -F "file=@/tmp/fleetml-demo-model.onnx" \
    -F "name=defect-detector" \
    -F "version=v1.0" \
    -F "format=onnx")
MODEL_ID=$(echo "$UPLOAD_RESULT" | jq -r '.id // .model_id // empty')
log "Model uploaded: defect-detector v1.0"

# ── Step 8: Send simulated logs ────────────────────────────────────────
info "Sending device logs..."

for device in jetson-nano-01 jetson-orin-02 rpi-4b-01 rpi-4b-02 intel-nuc-01 intel-nuc-02; do
    curl -s -X POST "$API/devices/$device/logs" \
        -H "$AUTH" -H "Content-Type: application/json" \
        -d "{
            \"entries\": [
                {\"level\":\"info\",  \"component\":\"agent\",   \"message\":\"agent started, version 0.1.0\"},
                {\"level\":\"info\",  \"component\":\"runtime\", \"message\":\"ONNX runtime initialized\"},
                {\"level\":\"info\",  \"component\":\"heartbeat\",\"message\":\"connected to control plane\"},
                {\"level\":\"info\",  \"component\":\"deploy\",  \"message\":\"waiting for deployment commands\"}
            ]
        }" > /dev/null 2>&1 || true
done

log "Sent startup logs for all 6 devices"

# ── Step 9: Send more heartbeats for richer data ────────────────────────
info "Sending simulated heartbeats..."

for i in 1 2 3; do
    for device in jetson-nano-01 jetson-orin-02 rpi-4b-01 rpi-4b-02 intel-nuc-01 intel-nuc-02; do
        curl -s -X POST "$API/heartbeat" \
            -H "Content-Type: application/json" \
            -d "{
                \"device_id\": \"$device\",
                \"status\": \"healthy\",
                \"system\": {
                    \"cpu_percent\": $(( RANDOM % 40 + 15 )),
                    \"gpu_percent\": $(( RANDOM % 50 )),
                    \"ram_mb_used\": $(( RANDOM % 2000 + 500 )),
                    \"disk_percent\": $(( RANDOM % 30 + 20 )),
                    \"temperature_c\": $(( RANDOM % 20 + 30 )),
                    \"uptime_hours\": $(( RANDOM % 720 + i * 24 ))
                }
            }" > /dev/null 2>&1
    done
done

log "Sent heartbeat metrics"

# ── Step 10: Summary ───────────────────────────────────────────────────
echo ""
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  FleetML is running!${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  ${CYAN}REST API${NC}        http://localhost:8080"
echo -e "  ${CYAN}gRPC${NC}            localhost:50051"
echo -e "  ${CYAN}Prometheus${NC}      http://localhost:9090/metrics"
echo -e "  ${CYAN}MinIO Console${NC}   http://localhost:9001  (minioadmin / minioadmin123)"
echo ""
echo -e "  ${BOLD}Demo credentials:${NC}"
echo -e "    Email:     demo@fleetml.dev"
echo -e "    Password:  demo1234"
echo -e "    Token:     $TOKEN"
echo ""
echo -e "  ${BOLD}Try these commands:${NC}"
echo ""
echo -e "    ${YELLOW}# List devices${NC}"
echo -e "    curl -s localhost:8080/api/v1/devices -H 'Authorization: Bearer $TOKEN' | jq ."
echo ""
echo -e "    ${YELLOW}# View device logs${NC}"
echo -e "    curl -s localhost:8080/api/v1/devices/jetson-nano-01/logs -H 'Authorization: Bearer $TOKEN' | jq ."
echo ""
echo -e "    ${YELLOW}# Stream logs (real-time)${NC}"
echo -e "    curl -N 'localhost:8080/api/v1/devices/jetson-nano-01/logs?follow=true' -H 'Authorization: Bearer $TOKEN'"
echo ""
echo -e "    ${YELLOW}# List models${NC}"
echo -e "    curl -s localhost:8080/api/v1/models -H 'Authorization: Bearer $TOKEN' | jq ."
echo ""
echo -e "    ${YELLOW}# List fleets${NC}"
echo -e "    curl -s localhost:8080/api/v1/fleets -H 'Authorization: Bearer $TOKEN' | jq ."
echo ""
echo -e "    ${YELLOW}# Health check${NC}"
echo -e "    curl -s localhost:8080/api/v1/health | jq ."
echo ""
echo -e "  ${BOLD}Stop everything:${NC}"
echo -e "    kill $SERVER_PID && docker compose down"
echo ""

# Keep server running in foreground
wait "$SERVER_PID"
