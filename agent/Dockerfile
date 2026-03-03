FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY proto/ ./proto/
COPY agent/ ./agent/
WORKDIR /app/proto
RUN go mod tidy
WORKDIR /app/agent
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /fleetml-agent ./cmd/agent

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /fleetml-agent /usr/local/bin/fleetml-agent
ENTRYPOINT ["fleetml-agent"]
