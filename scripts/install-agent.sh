#!/usr/bin/env bash
# FleetML Agent Installer
# Usage: curl -sSL https://raw.githubusercontent.com/frozo-ai/fleetml/main/scripts/install-agent.sh | sh
#
# Installs the FleetML agent binary to /usr/local/bin/fleetml-agent
# Supports: Linux (amd64, arm64), macOS (amd64, arm64)

set -euo pipefail

REPO="frozo-ai/fleetml"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="fleetml-agent"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

log()  { echo -e "${GREEN}[+]${NC} $1"; }
info() { echo -e "${CYAN}[>]${NC} $1"; }
err()  { echo -e "${RED}[!]${NC} $1"; exit 1; }

echo ""
echo -e "${BOLD}FleetML Agent Installer${NC}"
echo ""

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    armv7l)  ARCH="armv6" ;;
    armv6l)  ARCH="armv6" ;;
    *)       err "Unsupported architecture: $ARCH" ;;
esac

case "$OS" in
    linux)  ;;
    darwin) ;;
    *)      err "Unsupported OS: $OS. The agent requires Linux or macOS." ;;
esac

info "Detected: ${OS}/${ARCH}"

# Get latest release tag
info "Fetching latest release..."
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name"' | head -1 | sed -E 's/.*"v?([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
    # No release yet — build from source
    info "No release found. Installing from source..."

    if ! command -v go &>/dev/null; then
        err "Go is required to install from source. Install Go from https://go.dev/dl/"
    fi

    TMPDIR=$(mktemp -d)
    trap "rm -rf $TMPDIR" EXIT

    info "Cloning repository..."
    git clone --depth 1 "https://github.com/${REPO}.git" "$TMPDIR/fleetml" 2>/dev/null

    info "Building agent..."
    cd "$TMPDIR/fleetml/agent"
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$TMPDIR/fleetml-agent-bin" ./cmd/agent

    info "Installing to ${INSTALL_DIR}..."
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMPDIR/fleetml-agent-bin" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo mv "$TMPDIR/fleetml-agent-bin" "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    log "FleetML agent installed (built from source)"
    echo ""
    echo -e "  Start the agent:"
    echo -e "    ${BOLD}export FLEETML_API_KEY=\"your-api-key\"${NC}"
    echo -e "    ${BOLD}export FLEETML_SERVER=\"your-server:50051\"${NC}"
    echo -e "    ${BOLD}fleetml-agent${NC}"
    echo ""
    exit 0
fi

# Download release binary
VERSION="$LATEST"
ASSET_NAME="agent-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${ASSET_NAME}"

info "Downloading FleetML agent v${VERSION} (${OS}/${ARCH})..."
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMPDIR/${BINARY_NAME}"; then
    err "Failed to download ${DOWNLOAD_URL}"
fi

chmod +x "$TMPDIR/${BINARY_NAME}"

# Install
info "Installing to ${INSTALL_DIR}..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMPDIR/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
else
    sudo mv "$TMPDIR/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
fi

log "FleetML agent v${VERSION} installed successfully!"
echo ""
echo -e "  ${BOLD}Quick start:${NC}"
echo -e "    export FLEETML_API_KEY=\"your-api-key\""
echo -e "    export FLEETML_SERVER=\"your-server:50051\""
echo -e "    fleetml-agent"
echo ""
echo -e "  Get your API key at: ${CYAN}https://app.fleetml.dev/dashboard/get-started${NC}"
echo ""
