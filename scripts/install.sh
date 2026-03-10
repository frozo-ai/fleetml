#!/usr/bin/env bash
# FleetML CLI Installer
# Usage: curl -sSL https://raw.githubusercontent.com/ashish-frozo/fleetML/main/scripts/install.sh | bash
#
# Installs the latest FleetML CLI binary to /usr/local/bin/fleetml

set -euo pipefail

REPO="ashish-frozo/fleetML"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="fleetml"

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
echo -e "${BOLD}FleetML CLI Installer${NC}"
echo ""

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)       err "Unsupported architecture: $ARCH" ;;
esac

case "$OS" in
    linux)  ;;
    darwin) ;;
    *)      err "Unsupported OS: $OS" ;;
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

    info "Building CLI..."
    cd "$TMPDIR/fleetml/cli"
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$TMPDIR/fleetml-bin" ./cmd/fleetml

    info "Installing to ${INSTALL_DIR}..."
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMPDIR/fleetml-bin" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo mv "$TMPDIR/fleetml-bin" "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    log "FleetML CLI installed (built from source)"
    echo ""
    echo -e "  Run: ${BOLD}fleetml init --cloud${NC}"
    echo ""
    exit 0
fi

# Download release binary
VERSION="$LATEST"
TARBALL="fleetml_${VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${TARBALL}"

info "Downloading FleetML CLI v${VERSION}..."
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMPDIR/$TARBALL"; then
    err "Failed to download ${DOWNLOAD_URL}"
fi

# Extract
info "Extracting..."
tar -xzf "$TMPDIR/$TARBALL" -C "$TMPDIR"

# Install
info "Installing to ${INSTALL_DIR}..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMPDIR/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
else
    sudo mv "$TMPDIR/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
fi
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

log "FleetML CLI v${VERSION} installed successfully!"
echo ""
echo -e "  Get started:"
echo -e "    ${BOLD}fleetml init --cloud${NC}    # Connect to FleetML Cloud"
echo -e "    ${BOLD}fleetml init${NC}            # Connect to self-hosted server"
echo ""
