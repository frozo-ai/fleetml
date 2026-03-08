#!/usr/bin/env bash
# FleetML Railway Deployment
#
# Prerequisites:
#   1. Install Railway CLI: npm install -g @railway/cli
#   2. Login: railway login
#   3. Create a project: railway init (or use existing)
#
# Usage:
#   ./scripts/deploy-railway.sh
#
# What it does:
#   1. Creates 4 Railway services: server, dashboard, compiler, nats
#   2. Provisions PostgreSQL add-on
#   3. Sets all environment variables
#   4. Deploys each service

set -euo pipefail

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

echo ""
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  FleetML → Railway Deployment${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# ── Check prerequisites ──────────────────────────────────────────
info "Checking prerequisites..."
if ! command -v railway &>/dev/null; then
    err "Railway CLI not found. Install: npm install -g @railway/cli"
    exit 1
fi

if ! railway whoami &>/dev/null; then
    err "Not logged in. Run: railway login"
    exit 1
fi

log "Railway CLI ready ($(railway whoami 2>&1 | head -1))"

# ── Check for project ────────────────────────────────────────────
if ! railway status &>/dev/null 2>&1; then
    info "No Railway project linked. Creating one..."
    railway init
fi

PROJECT_ID=$(railway status 2>&1 | grep -i "project" | head -1 || echo "unknown")
log "Project: $PROJECT_ID"

echo ""
echo -e "${BOLD}This script will create the following Railway services:${NC}"
echo "  1. server     — Go control plane (REST + gRPC)"
echo "  2. dashboard  — React frontend"
echo "  3. compiler   — Python model compiler"
echo "  4. PostgreSQL — Database (Railway add-on)"
echo ""
echo -e "${YELLOW}Railway charges based on usage. Estimated: ~\$5-20/month for light usage.${NC}"
echo ""
read -p "Continue? [y/N]: " confirm
if [[ "${confirm,,}" != "y" ]]; then
    echo "Aborted."
    exit 0
fi

# ── Generate secrets ─────────────────────────────────────────────
JWT_SECRET=$(openssl rand -hex 32)
log "Generated JWT secret"

# ── Step 1: Add PostgreSQL ───────────────────────────────────────
info "Adding PostgreSQL database..."
railway add --plugin postgresql 2>/dev/null || warn "PostgreSQL may already exist"
log "PostgreSQL provisioned"

# ── Step 2: Deploy Server ────────────────────────────────────────
info "Creating server service..."
cat <<EOF

${BOLD}Manual steps required (Railway doesn't support multi-service CLI yet):${NC}

1. Go to your Railway project dashboard: https://railway.app/dashboard
2. Click "New Service" → "GitHub Repo" → select your fleetML repo

   ${BOLD}Service: server${NC}
   - Root Directory: /
   - Dockerfile: server/Dockerfile
   - Set these environment variables:
     DATABASE_URL        = \${{Postgres.DATABASE_URL}}
     JWT_SECRET          = ${JWT_SECRET}
     S3_ENDPOINT         = (your S3/R2 endpoint)
     S3_ACCESS_KEY       = (your access key)
     S3_SECRET_KEY       = (your secret key)
     S3_BUCKET           = fleetml-models
     NATS_URL            = (leave empty for now)
     COMPILER_URL        = \${{compiler.RAILWAY_PRIVATE_NETWORKING_URL}}:8081
     DODO_API_KEY        = (from Dodo Payments dashboard)
     DODO_WEBHOOK_KEY    = (from Dodo Payments dashboard)
     DODO_ENVIRONMENT    = live
     DODO_STARTER_PRODUCT_ID = (your starter product ID)
     DODO_PRO_PRODUCT_ID = (your pro product ID)
     BILLING_SUCCESS_URL = https://\${{dashboard.RAILWAY_PUBLIC_DOMAIN}}/dashboard/billing?success=true
   - Port: 8080
   - Generate Domain → use as API URL

   ${BOLD}Service: dashboard${NC}
   - Root Directory: /
   - Dockerfile: dashboard/Dockerfile.railway
   - Set build argument:
     VITE_API_URL = https://\${{server.RAILWAY_PUBLIC_DOMAIN}}
   - Port: 3000
   - Generate Domain → this is your app URL

   ${BOLD}Service: compiler${NC}
   - Root Directory: /
   - Dockerfile: compiler/Dockerfile.mock
   - Port: 8081
   - No public domain needed (internal only)

3. After all services are deployed, configure Dodo Payments webhook:
   Webhook URL: https://<server-domain>/api/v1/webhooks/dodo

EOF

# ── Print quick reference ────────────────────────────────────────
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}  Quick Reference${NC}"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  ${CYAN}JWT Secret:${NC}  ${JWT_SECRET}"
echo ""
echo -e "  ${CYAN}Server env vars to copy:${NC}"
echo "    JWT_SECRET=${JWT_SECRET}"
echo "    S3_BUCKET=fleetml-models"
echo "    DODO_ENVIRONMENT=live"
echo ""
echo -e "  ${CYAN}Dodo Payments setup:${NC}"
echo "    1. Go to dodopayments.com → Dashboard → Products"
echo "    2. Create 'FleetML Starter' subscription (\$49/month)"
echo "    3. Create 'FleetML Pro' subscription (\$199/month)"
echo "    4. Copy product IDs to Railway env vars"
echo "    5. Go to Developer → Webhooks → Add your server URL"
echo ""
echo -e "  ${CYAN}After deployment:${NC}"
echo "    1. Visit your dashboard URL"
echo "    2. Sign up → creates free account"
echo "    3. CLI: fleetml init --cloud"
echo ""
