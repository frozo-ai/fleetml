# FleetML Development Guide

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker & Docker Compose
- `buf` CLI (for protobuf codegen)
- `golangci-lint`

## Repository Structure

```
fleetml/
├── agent/          # Edge agent (Go)
├── server/         # Control plane (Go)
├── cli/            # CLI tool (Go)
├── dashboard/      # Web UI (React/TS)
├── compiler/       # Model compiler (Python)
├── proto/          # Protobuf definitions
├── simulator/      # Virtual fleet simulator
├── tests/          # Integration, fleet, security, chaos tests
├── docs/           # Documentation
├── Makefile        # Build commands
└── docker-compose.yml
```

## Getting Started

```bash
# Clone and setup
git clone https://github.com/fleetml/fleetml.git
cd fleetml

# Start infrastructure
docker compose up -d db minio

# Generate protobuf code
make proto

# Build everything
make build

# Run tests
make test-unit
```

## Development Workflow

### Running Locally

```bash
# Terminal 1: Server
cd server && go run ./cmd/server

# Terminal 2: Dashboard
cd dashboard && npm run dev

# Terminal 3: Agent (optional)
cd agent && go run ./cmd/agent
```

### Running Tests

```bash
# Unit tests (fast, no docker needed)
make test-unit

# Single package
cd agent && go test -race -run TestHotSwap ./internal/model/...

# Integration tests (requires docker)
make test-integration

# Fleet simulation
make test-fleet

# All tests
make test
```

### Build Tags

Tests are organized by build tags:

| Tag | Command | Description |
|-----|---------|-------------|
| (none) | `go test ./...` | Unit tests only |
| `integration` | `go test -tags=integration` | Docker-based integration |
| `fleet` | `go test -tags=fleet` | Virtual fleet scenarios |
| `chaos` | `go test -tags=chaos` | Chaos engineering |
| `benchmark` | `go test -tags=benchmark -bench .` | Performance benchmarks |

### Protobuf Changes

```bash
# Edit proto files in proto/fleetml/v1/
# Then regenerate:
make proto

# Lint:
buf lint
```

### Database Migrations

```bash
# Create a new migration
# File: server/migrations/NNN_description.up.sql

# Apply migrations
make migrate-up

# Rollback
make migrate-down
```

## Code Standards

### Go

- Run `golangci-lint run` before committing
- Use `go test -race` for all tests
- Follow standard Go project layout
- Separate modules for agent/server/CLI

### TypeScript

- Run `npm run lint` before committing
- Use strict TypeScript
- Functional components with hooks
- TanStack Query for server state

## CI Pipeline

Every commit runs:
1. `golangci-lint` + `npm run lint`
2. Unit tests with `-race`
3. Cross-compilation check (linux/amd64 + linux/arm64)
4. Docker build verification

PRs additionally run:
5. Integration tests (Docker Compose)
6. Virtual fleet tests (20 devices)
7. Security scan (`govulncheck`, `trivy`)
