#!/bin/bash
# E2E: First-time user experience
# Tests: install CLI -> init -> deploy -> status -> rollback
set -euo pipefail

SERVER_URL="${FLEETML_SERVER_URL:-http://localhost:8080}"
CLI="${FLEETML_CLI:-./bin/fleetml}"

echo "=== FleetML E2E: First-Time User ==="

# 1. Version check
echo "Step 1: Version check"
$CLI version
echo "PASS: version"

# 2. Init
echo "Step 2: Initialize CLI"
$CLI init --server "$SERVER_URL"
echo "PASS: init"

# 3. Create a test model file
echo "Step 3: Create test model"
TMPMODEL=$(mktemp /tmp/test-model-XXXX.onnx)
dd if=/dev/urandom bs=1024 count=10 of="$TMPMODEL" 2>/dev/null
echo "PASS: created test model at $TMPMODEL"

# 4. Deploy
echo "Step 4: Deploy model"
$CLI deploy "$TMPMODEL" --name e2e-test --version 1.0 --fleet default --policy immediate
echo "PASS: deploy"

# 5. Status
echo "Step 5: Check status"
$CLI status
echo "PASS: status"

# 6. Deploy v2
echo "Step 6: Deploy v2"
dd if=/dev/urandom bs=1024 count=10 of="$TMPMODEL" 2>/dev/null
$CLI deploy "$TMPMODEL" --name e2e-test --version 2.0 --fleet default
echo "PASS: deploy v2"

# 7. Rollback
echo "Step 7: Rollback"
$CLI rollback --fleet default
echo "PASS: rollback"

# Cleanup
rm -f "$TMPMODEL"

echo ""
echo "=== ALL E2E TESTS PASSED ==="
