package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.RESTPort != 8080 {
		t.Errorf("expected REST port 8080, got %d", cfg.Server.RESTPort)
	}
	if cfg.Server.GRPCPort != 50051 {
		t.Errorf("expected gRPC port 50051, got %d", cfg.Server.GRPCPort)
	}
	if cfg.Database.MaxConnections != 50 {
		t.Errorf("expected max connections 50, got %d", cfg.Database.MaxConnections)
	}
	if cfg.Storage.Bucket != "fleetml-models" {
		t.Errorf("expected bucket 'fleetml-models', got %q", cfg.Storage.Bucket)
	}
	if cfg.Auth.JWTExpiry != 24*time.Hour {
		t.Errorf("expected JWT expiry 24h, got %v", cfg.Auth.JWTExpiry)
	}
	if cfg.Heartbeat.OfflineThreshold != 90*time.Second {
		t.Errorf("expected offline threshold 90s, got %v", cfg.Heartbeat.OfflineThreshold)
	}
	if cfg.Deploy.DefaultTimeout != 10*time.Minute {
		t.Errorf("expected deploy timeout 10m, got %v", cfg.Deploy.DefaultTimeout)
	}
	if cfg.Compiler.URL != "" {
		t.Errorf("expected empty compiler URL, got %q", cfg.Compiler.URL)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("expected log level 'info', got %q", cfg.Logging.Level)
	}
}

func TestDefaultConfig_StorageDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Storage.Type != "s3" {
		t.Errorf("expected storage type 's3', got %q", cfg.Storage.Type)
	}
	if cfg.Storage.Endpoint != "http://localhost:9000" {
		t.Errorf("expected endpoint 'http://localhost:9000', got %q", cfg.Storage.Endpoint)
	}
	if cfg.Storage.Region != "us-east-1" {
		t.Errorf("expected region 'us-east-1', got %q", cfg.Storage.Region)
	}
}

func TestDefaultConfig_TLSDisabledByDefault(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.TLS.Enabled {
		t.Error("expected TLS to be disabled by default")
	}
	if cfg.Server.TLS.CertFile != "" {
		t.Error("expected empty cert file")
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.RESTPort != 8080 {
		t.Errorf("expected default REST port, got %d", cfg.Server.RESTPort)
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoad_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")

	yaml := `
server:
  rest_port: 9090
  grpc_port: 60061
database:
  url: postgres://test:test@db:5432/test
  max_connections: 100
storage:
  bucket: test-bucket
compiler:
  url: http://compiler:8081
logging:
  level: debug
`
	if err := os.WriteFile(cfgFile, []byte(yaml), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.RESTPort != 9090 {
		t.Errorf("expected REST port 9090, got %d", cfg.Server.RESTPort)
	}
	if cfg.Server.GRPCPort != 60061 {
		t.Errorf("expected gRPC port 60061, got %d", cfg.Server.GRPCPort)
	}
	if cfg.Database.MaxConnections != 100 {
		t.Errorf("expected max connections 100, got %d", cfg.Database.MaxConnections)
	}
	if cfg.Storage.Bucket != "test-bucket" {
		t.Errorf("expected bucket 'test-bucket', got %q", cfg.Storage.Bucket)
	}
	if cfg.Compiler.URL != "http://compiler:8081" {
		t.Errorf("expected compiler URL 'http://compiler:8081', got %q", cfg.Compiler.URL)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected log level 'debug', got %q", cfg.Logging.Level)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	// Set env vars
	t.Setenv("DATABASE_URL", "postgres://env:env@envdb:5432/envdb")
	t.Setenv("S3_ENDPOINT", "http://env-minio:9000")
	t.Setenv("S3_ACCESS_KEY", "env-access")
	t.Setenv("S3_SECRET_KEY", "env-secret")
	t.Setenv("S3_BUCKET", "env-bucket")
	t.Setenv("JWT_SECRET", "env-jwt-secret")
	t.Setenv("COMPILER_URL", "http://env-compiler:8081")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Database.URL != "postgres://env:env@envdb:5432/envdb" {
		t.Errorf("expected env DATABASE_URL, got %q", cfg.Database.URL)
	}
	if cfg.Storage.Endpoint != "http://env-minio:9000" {
		t.Errorf("expected env S3_ENDPOINT, got %q", cfg.Storage.Endpoint)
	}
	if cfg.Storage.AccessKey != "env-access" {
		t.Errorf("expected env S3_ACCESS_KEY, got %q", cfg.Storage.AccessKey)
	}
	if cfg.Storage.SecretKey != "env-secret" {
		t.Errorf("expected env S3_SECRET_KEY, got %q", cfg.Storage.SecretKey)
	}
	if cfg.Storage.Bucket != "env-bucket" {
		t.Errorf("expected env S3_BUCKET, got %q", cfg.Storage.Bucket)
	}
	if cfg.Auth.JWTSecret != "env-jwt-secret" {
		t.Errorf("expected env JWT_SECRET, got %q", cfg.Auth.JWTSecret)
	}
	if cfg.Compiler.URL != "http://env-compiler:8081" {
		t.Errorf("expected env COMPILER_URL, got %q", cfg.Compiler.URL)
	}
}

func TestLoad_YAMLOverridesDefaults(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")

	yaml := `
server:
  rest_port: 3000
`
	if err := os.WriteFile(cfgFile, []byte(yaml), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// REST port should be overridden
	if cfg.Server.RESTPort != 3000 {
		t.Errorf("expected REST port 3000, got %d", cfg.Server.RESTPort)
	}
	// gRPC port should retain default
	if cfg.Server.GRPCPort != 50051 {
		t.Errorf("expected default gRPC port 50051, got %d", cfg.Server.GRPCPort)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(cfgFile, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(cfgFile)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoad_EnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")

	yaml := `
database:
  url: postgres://yaml:yaml@yamldb:5432/yaml
`
	if err := os.WriteFile(cfgFile, []byte(yaml), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Env should override YAML
	t.Setenv("DATABASE_URL", "postgres://env:env@envdb:5432/env")

	cfg, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Database.URL != "postgres://env:env@envdb:5432/env" {
		t.Errorf("expected env to override YAML, got %q", cfg.Database.URL)
	}
}

func TestDefaultConfig_DeployDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Deploy.ConcurrentDeploysPerDev != 1 {
		t.Errorf("expected 1 concurrent deploy per device, got %d", cfg.Deploy.ConcurrentDeploysPerDev)
	}
}

func TestConfigStructTypes(t *testing.T) {
	// Verify the config struct hierarchy is valid
	cfg := &Config{
		Server: ServerConfig{
			RESTPort: 8080,
			GRPCPort: 50051,
			TLS: TLSConfig{
				Enabled:  true,
				CertFile: "/path/to/cert",
				KeyFile:  "/path/to/key",
				CAFile:   "/path/to/ca",
			},
		},
		Compiler: CompilerConfig{
			URL: "http://localhost:8081",
		},
	}

	if !cfg.Server.TLS.Enabled {
		t.Error("expected TLS enabled")
	}
	if cfg.Server.TLS.CertFile != "/path/to/cert" {
		t.Errorf("expected cert file path, got %q", cfg.Server.TLS.CertFile)
	}
}
