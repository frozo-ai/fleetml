package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ServerAddress != "localhost:50051" {
		t.Fatalf("expected default server address 'localhost:50051', got '%s'", cfg.ServerAddress)
	}
	if cfg.HeartbeatInterval != 30*time.Second {
		t.Fatalf("expected 30s heartbeat interval, got %v", cfg.HeartbeatInterval)
	}
	if cfg.MaxModelVersions != 3 {
		t.Fatalf("expected 3 max versions, got %d", cfg.MaxModelVersions)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("expected 'info' log level, got '%s'", cfg.LogLevel)
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("FLEETML_SERVER", "myserver:50051")
	t.Setenv("DEVICE_ID", "test-device-123")
	t.Setenv("FLEETML_MODE", "test")

	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ServerAddress != "myserver:50051" {
		t.Fatalf("expected server override, got '%s'", cfg.ServerAddress)
	}
	if cfg.DeviceID != "test-device-123" {
		t.Fatalf("expected device_id override, got '%s'", cfg.DeviceID)
	}
	if cfg.Mode != "test" {
		t.Fatalf("expected mode override, got '%s'", cfg.Mode)
	}
}

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := `device_id: "file-device"
server_address: "fileserver:50051"
log_level: "debug"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.DeviceID != "file-device" {
		t.Fatalf("expected 'file-device', got '%s'", cfg.DeviceID)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("expected 'debug', got '%s'", cfg.LogLevel)
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}
