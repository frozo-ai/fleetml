# Contributing to FleetML

Thanks for your interest in contributing to FleetML! This guide covers everything you need to get started.

## Development Setup

### Prerequisites

- **Go 1.22+** — for agent, server, and CLI
- **Node.js 18+** — for the dashboard
- **Python 3.9+** — for the compiler service
- **Docker & Docker Compose** — for infrastructure (PostgreSQL, MinIO, NATS)
- **protoc + buf** — for protobuf code generation (only if modifying `.proto` files)

### Clone and Build

```bash
git clone https://github.com/fleetml/fleetml.git
cd fleetml
cp .env.example .env    # Set DB_PASSWORD, MINIO_ACCESS_KEY, etc.
make build              # Builds agent + server + cli + dashboard
```

### Start Infrastructure

```bash
docker compose up -d db minio nats   # PostgreSQL, MinIO, NATS
```

### Run the Server

```bash
cd server && go run ./cmd/server/
# REST API on :8080, gRPC on :50051, Prometheus on :9090
```

### Run the Dashboard

```bash
cd dashboard && npm install && npm run dev
# Dev server on :3000 with HMR
```

### Run the Compiler

```bash
cd compiler && pip install -r requirements.txt && python main.py
# FastAPI on :8081
```

## Project Structure

```
fleetml/
  agent/       # Edge agent (Go) — each has its own go.mod
  server/      # Control plane (Go)
  cli/         # CLI (Go)
  dashboard/   # Web UI (TypeScript/React)
  compiler/    # Model compiler (Python/FastAPI)
  proto/       # Protobuf definitions
  simulator/   # Virtual fleet simulator (Go)
  tests/       # Cross-component tests (fleet, chaos, scale, security, load, e2e)
```

Each Go component (`agent/`, `server/`, `cli/`) has its own `go.mod`. Run `go test` from within the component directory.

## Running Tests

```bash
# Unit tests (all Go components)
make test-unit

# Single package
cd server && go test -race ./internal/model/...

# Single test
cd agent && go test -race -run TestHotSwap ./internal/model/...

# Integration tests (needs Docker)
make test-integration

# Virtual fleet simulation (20 devices)
make test-fleet

# Chaos engineering
cd tests && go test -tags=chaos ./chaos/...

# Python compiler tests
cd compiler && python -m pytest tests/ -v

# Dashboard lint
cd dashboard && npm run lint

# Load tests (needs k6)
k6 run tests/load/k6-heartbeat.js
k6 run tests/load/k6-api.js
```

### Build Tags

Tests are separated by build tags:

| Tag | Purpose | When |
|-----|---------|------|
| (none) | Unit tests | Every commit |
| `integration` | Docker-dependent tests | Every PR |
| `fleet` | Virtual fleet simulator | Every PR |
| `chaos` | Fault injection | Nightly |

## Making Changes

### Branch Naming

```
feature/description    # New features
fix/description        # Bug fixes
refactor/description   # Refactoring
test/description       # Test additions
docs/description       # Documentation
```

### Commit Messages

Use conventional-style messages:

```
Add canary rollback detection for failed deployments
Fix heartbeat timeout when agent reconnects after offline
Update model registry to support compiled variants
```

Start with a verb (Add, Fix, Update, Remove, Refactor). Keep the first line under 72 characters.

### Code Style

**Go:**
- Follow standard `gofmt` formatting
- Use `golangci-lint` (run `make lint`)
- Error handling: return errors, don't panic. Use `fmt.Errorf("context: %w", err)` for wrapping
- Logging: use `zap.SugaredLogger` with structured fields (`log.Infow("msg", "key", value)`)
- Tests: table-driven tests preferred. Use `t.Fatalf` for setup failures, `t.Errorf` for assertions

**TypeScript/React:**
- Follow existing ESLint config
- Use functional components with hooks
- Use TanStack Query for API calls
- Tailwind CSS for styling

**Python:**
- Follow PEP 8
- Use type hints (compatible with Python 3.9 — use `Optional[X]` not `X | None`)
- Use Pydantic for request/response models
- Use pytest for tests

### What to Include in PRs

- Code changes
- Tests covering the changes
- Updated comments if behavior changed
- No generated files (protobuf output is committed, but re-generate if you change `.proto` files)

## Pull Request Process

1. Fork and create a branch from `main`
2. Make your changes with tests
3. Run `make lint` and `make test-unit` locally
4. Push and open a PR
5. PR must pass CI checks:
   - Lint (Go + JS)
   - Unit tests
   - Integration tests (Docker-based)
   - Virtual fleet tests
   - Security scan (`govulncheck`)
   - Build verification (amd64 + arm64)
6. One maintainer approval required for merge

## Quality Gates

PRs must meet these criteria:

- All existing tests pass
- New code has test coverage
- No `golangci-lint` or ESLint errors
- No security vulnerabilities (`govulncheck`, `npm audit`)
- Agent binary builds for both `linux/amd64` and `linux/arm64`

## Architecture Notes

**Key patterns to follow:**

- **Offline-first**: Agent must work without server connectivity. Buffer heartbeats in SQLite, queue commands in NATS
- **Zero-downtime**: Model hot-swap uses `sync/atomic.Pointer` — never block inference during updates
- **Chip-neutral**: ONNX as universal input. Compiled variants stored in `compiled_variants` JSONB column
- **Store-and-forward**: Heartbeats buffered locally during disconnection, bulk-synced on reconnect

**Important interfaces:**

| Interface | Location | Purpose |
|-----------|----------|---------|
| `BaseCompiler` | `compiler/base.py` | Implement for new hardware runtimes |
| `model.Registry` | `server/internal/model/` | Model storage and retrieval |
| `deploy.Orchestrator` | `server/internal/deploy/` | Deployment lifecycle management |
| `fleet.Manager` | `server/internal/fleet/` | Fleet and device management |

## Reporting Issues

- Use GitHub Issues
- Include: steps to reproduce, expected vs actual behavior, OS/hardware info
- For security issues, email security@fleetml.io instead of opening a public issue

## Getting Help

- Open a Discussion on GitHub
- Check existing issues and PRs first
- Look at the code — the codebase is well-structured and self-documenting

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
