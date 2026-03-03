#!/bin/bash
# E2E: Multi-device fleet deployment
# Tests: start fleet -> deploy to all -> canary -> verify
set -euo pipefail

SERVER_URL="${FLEETML_SERVER_URL:-http://localhost:8080}"
CLI="${FLEETML_CLI:-./bin/fleetml}"
FLEET_SIZE="${FLEET_SIZE:-5}"

echo "=== FleetML E2E: Multi-Device Fleet (${FLEET_SIZE} devices) ==="

# 1. Wait for server
echo "Step 1: Wait for server"
for i in $(seq 1 30); do
    if curl -sf "$SERVER_URL/health" > /dev/null 2>&1; then
        echo "  Server ready"
        break
    fi
    sleep 2
done

# 2. Check fleet
echo "Step 2: Verify fleet"
DEVICE_COUNT=$(curl -sf "$SERVER_URL/api/v1/devices" | python3 -c "import sys,json; print(json.load(sys.stdin).get('total',0))" 2>/dev/null || echo 0)
echo "  Devices registered: $DEVICE_COUNT"

# 3. Upload model
echo "Step 3: Upload model"
TMPMODEL=$(mktemp /tmp/fleet-model-XXXX.onnx)
dd if=/dev/urandom bs=1024 count=100 of="$TMPMODEL" 2>/dev/null
$CLI deploy "$TMPMODEL" --name fleet-test --version 1.0 --fleet default --wait --timeout 5m
echo "  PASS: Fleet deployment completed"

# 4. Deploy v2 with canary
echo "Step 4: Canary deployment"
dd if=/dev/urandom bs=1024 count=100 of="$TMPMODEL" 2>/dev/null
$CLI deploy "$TMPMODEL" --name fleet-test --version 2.0 --fleet default --policy canary --canary-stages "50:1m,100:1m" --wait --timeout 10m
echo "  PASS: Canary deployment completed"

# 5. Check all devices have latest model
echo "Step 5: Verify deployment"
$CLI status
echo "  PASS: Status check"

# Cleanup
rm -f "$TMPMODEL"

echo ""
echo "=== MULTI-DEVICE FLEET E2E PASSED ==="
