#!/usr/bin/env bash
# FleetML Terminal Demo Script
# ============================
# Records an animated terminal demo showing: init → deploy → status → rollback.
# Use with https://github.com/charmbracelet/vhs or asciinema for recording.
#
# Prerequisites:
#   - docker compose up -d   (infra running)
#   - fleetml CLI in PATH
#   - A sample .onnx model file at ./models/demo_model.onnx
#
# Usage:
#   ./scripts/demo.sh              # Run interactively
#   vhs scripts/demo.tape          # Record as GIF with VHS
#   asciinema rec -c ./scripts/demo.sh demo.cast  # Record with asciinema

set -euo pipefail

# --- Configuration ---
SERVER_URL="${FLEETML_SERVER:-http://localhost:8080}"
API_KEY="${FLEETML_API_KEY:-demo-api-key}"
DEVICE_ID="edge-demo-001"
FLEET_NAME="demo-fleet"
MODEL_FILE="${DEMO_MODEL:-./models/demo_model.onnx}"
PAUSE=${DEMO_PAUSE:-2}   # seconds between steps (increase for recordings)

# --- Helpers ---
banner() {
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  $1"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""
}

step() {
  echo ""
  echo "▸ $1"
  echo ""
  sleep "$PAUSE"
}

run() {
  echo "$ $*"
  eval "$@" 2>&1 || true
  sleep "$PAUSE"
}

# --- Demo Starts ---
banner "FleetML — Edge MLOps Demo"
echo "Deploy, monitor, and manage ML models on edge device fleets."
echo ""
sleep "$PAUSE"

# Step 1: Version check
step "Step 1: Check FleetML CLI version"
run fleetml version

# Step 2: Health check
step "Step 2: Verify control plane is running"
run curl -s "$SERVER_URL/api/v1/health" | python3 -m json.tool

# Step 3: Register a device (via API since agents self-register)
step "Step 3: Register an edge device"
run curl -s -X POST "$SERVER_URL/api/v1/devices" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "'{\"device_id\": \"$DEVICE_ID\", \"name\": \"Demo Jetson Orin\", \"arch\": \"arm64\", \"runtime\": \"tensorrt\", \"gpu_type\": \"jetson-orin\", \"labels\": {\"location\": \"lab\", \"env\": \"demo\"}}'" \
  | python3 -m json.tool

# Step 4: Create a fleet
step "Step 4: Create a device fleet"
FLEET_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/fleets" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$FLEET_NAME\", \"description\": \"Demo fleet for edge devices\"}")
echo "$ curl -X POST .../fleets"
echo "$FLEET_RESPONSE" | python3 -m json.tool
FLEET_ID=$(echo "$FLEET_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))" 2>/dev/null || echo "demo-fleet-id")
sleep "$PAUSE"

# Step 5: Deploy a model
step "Step 5: Deploy a model to the fleet"
if [ -f "$MODEL_FILE" ]; then
  run fleetml deploy \
    --model face-detector \
    --version v1.0 \
    --target fleet \
    --target-id "$FLEET_ID" \
    --file "$MODEL_FILE" \
    --policy rolling
else
  echo "(Skipping file upload — no model file at $MODEL_FILE)"
  run fleetml deploy \
    --model face-detector \
    --version v1.0 \
    --target fleet \
    --target-id "$FLEET_ID" \
    --policy rolling
fi

# Step 6: Check status
step "Step 6: Monitor deployment status"
run fleetml status

# Step 7: View device logs
step "Step 7: View device logs"
run fleetml logs --device "$DEVICE_ID" --tail 10

# Step 8: Deploy v2.0 with canary policy
step "Step 8: Canary deployment — roll out v2.0 gradually"
run fleetml deploy \
  --model face-detector \
  --version v2.0 \
  --target fleet \
  --target-id "$FLEET_ID" \
  --policy canary

# Step 9: Check status again
step "Step 9: Check canary deployment progress"
run fleetml status

# Step 10: Rollback
step "Step 10: Rollback to previous version"
DEPLOYMENT_ID=$(curl -s "$SERVER_URL/api/v1/deployments?status=in_progress" \
  -H "Authorization: Bearer $API_KEY" \
  | python3 -c "import sys,json; ds=json.load(sys.stdin).get('deployments',[]); print(ds[0]['id'] if ds else '')" 2>/dev/null || echo "")
if [ -n "$DEPLOYMENT_ID" ]; then
  run fleetml rollback --deployment-id "$DEPLOYMENT_ID"
else
  echo "(No active deployment to rollback — would run: fleetml rollback --deployment-id <id>)"
fi

# Step 11: Final status
step "Step 11: Verify rollback completed"
run fleetml status

# Done
banner "Demo Complete!"
echo "FleetML makes edge ML deployment simple:"
echo "  • Upload once, deploy to any hardware"
echo "  • Canary deployments with automatic rollback"
echo "  • Real-time monitoring across the fleet"
echo ""
echo "Learn more: https://github.com/frozo/fleetml"
echo ""
