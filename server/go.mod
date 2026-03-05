module github.com/fleetml/fleetml/server

go 1.24

require (
	github.com/fleetml/fleetml/proto v0.0.0
	github.com/go-chi/chi/v5 v5.0.12
	github.com/go-chi/cors v1.2.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/golang-migrate/migrate/v4 v4.17.0
	github.com/jackc/pgx/v5 v5.5.3
	github.com/minio/minio-go/v7 v7.0.67
	go.opentelemetry.io/otel v1.24.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.24.0
	go.opentelemetry.io/otel/sdk v1.24.0
	go.opentelemetry.io/otel/trace v1.24.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.20.0
	google.golang.org/grpc v1.62.0
	google.golang.org/protobuf v1.33.0
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/fleetml/fleetml/proto => ../proto
