.PHONY: build test lint agent server cli dashboard proto migrate-up

# Build all
build: agent server cli dashboard

# Build agent (cross-compile)
agent:
	cd agent && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../bin/agent-linux-amd64 ./cmd/agent
	cd agent && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ../bin/agent-linux-arm64 ./cmd/agent

# Build server
server:
	cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../bin/server-linux-amd64 ./cmd/server

# Build CLI
cli:
	cd cli && go build -ldflags="-s -w" -o ../bin/fleetml ./cmd/fleetml

# Build dashboard
dashboard:
	cd dashboard && npm install && npm run build

# Generate protobuf code
proto:
	cd proto && buf generate

# Run database migrations
migrate-up:
	cd server && go run ./cmd/migrate up

# Run all tests
test: test-unit test-integration

# Unit tests
test-unit:
	cd agent && go test -race -short ./...
	cd server && go test -race -short ./...
	cd cli && go test -race -short ./...

# Integration tests (requires Docker)
test-integration:
	docker compose -f docker-compose.test.yml up -d
	cd tests && go test -v -tags=integration ./integration/...
	docker compose -f docker-compose.test.yml down -v

# Virtual fleet tests
test-fleet:
	docker compose -f docker-compose.test.yml up -d
	cd tests && go test -v -tags=fleet ./fleet/... -fleet-size=20
	docker compose -f docker-compose.test.yml down -v

# Chaos tests
test-chaos:
	docker compose -f docker-compose.test.yml up -d
	cd tests && go test -v -tags=chaos ./chaos/... -timeout=60m
	docker compose -f docker-compose.test.yml down -v

# Lint
lint:
	cd agent && golangci-lint run ./...
	cd server && golangci-lint run ./...
	cd cli && golangci-lint run ./...
	cd dashboard && npm run lint

# Docker build
docker-build:
	docker build -t fleetml/agent:latest -f agent/Dockerfile .
	docker build -t fleetml/server:latest -f server/Dockerfile .
	docker build -t fleetml/dashboard:latest -f dashboard/Dockerfile .

# Start local dev environment
dev:
	docker compose up -d db minio
	cd server && go run ./cmd/server &
	cd dashboard && npm run dev &
