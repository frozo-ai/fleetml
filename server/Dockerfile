FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY proto/ ./proto/
COPY server/ ./server/
WORKDIR /app/proto
RUN go mod tidy
WORKDIR /app/server
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /fleetml-server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /fleetml-server /usr/local/bin/fleetml-server
COPY server/migrations/ /migrations/
EXPOSE 8080 50051
ENTRYPOINT ["fleetml-server"]
