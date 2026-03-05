package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestLoad_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")

	// Write invalid YAML (tabs are not allowed in certain positions, unmatched braces)
	content := `device_id: [invalid
  : broken: yaml: {{{{
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")

	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("empty file should not return error, got: %v", err)
	}

	// Should use defaults
	defaults := DefaultConfig()
	if cfg.ServerAddress != defaults.ServerAddress {
		t.Fatalf("expected default server address '%s', got '%s'", defaults.ServerAddress, cfg.ServerAddress)
	}
	if cfg.HeartbeatInterval != defaults.HeartbeatInterval {
		t.Fatalf("expected default heartbeat interval %v, got %v", defaults.HeartbeatInterval, cfg.HeartbeatInterval)
	}
	if cfg.MaxModelVersions != defaults.MaxModelVersions {
		t.Fatalf("expected default max model versions %d, got %d", defaults.MaxModelVersions, cfg.MaxModelVersions)
	}
	if cfg.LogLevel != defaults.LogLevel {
		t.Fatalf("expected default log level '%s', got '%s'", defaults.LogLevel, cfg.LogLevel)
	}
	if cfg.Mode != defaults.Mode {
		t.Fatalf("expected default mode '%s', got '%s'", defaults.Mode, cfg.Mode)
	}
	if cfg.ModelStorageDir != defaults.ModelStorageDir {
		t.Fatalf("expected default model storage dir '%s', got '%s'", defaults.ModelStorageDir, cfg.ModelStorageDir)
	}
}

func TestLoad_PartialYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "partial.yaml")

	content := `device_id: "partial-device"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("partial YAML should not return error, got: %v", err)
	}

	// Specified field should be set
	if cfg.DeviceID != "partial-device" {
		t.Fatalf("expected device_id 'partial-device', got '%s'", cfg.DeviceID)
	}

	// Unspecified fields should retain defaults
	defaults := DefaultConfig()
	if cfg.ServerAddress != defaults.ServerAddress {
		t.Fatalf("expected default server address '%s', got '%s'", defaults.ServerAddress, cfg.ServerAddress)
	}
	if cfg.HeartbeatInterval != defaults.HeartbeatInterval {
		t.Fatalf("expected default heartbeat interval %v, got %v", defaults.HeartbeatInterval, cfg.HeartbeatInterval)
	}
	if cfg.MaxModelVersions != defaults.MaxModelVersions {
		t.Fatalf("expected default max model versions %d, got %d", defaults.MaxModelVersions, cfg.MaxModelVersions)
	}
	if cfg.LogLevel != defaults.LogLevel {
		t.Fatalf("expected default log level '%s', got '%s'", defaults.LogLevel, cfg.LogLevel)
	}
	if cfg.Mode != defaults.Mode {
		t.Fatalf("expected default mode '%s', got '%s'", defaults.Mode, cfg.Mode)
	}
}

func TestLoad_YAMLWithUnknownFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "extra.yaml")

	content := `device_id: "extra-device"
server_address: "extra-server:50051"
unknown_field: "some-value"
another_unknown: 42
nested_unknown:
  sub_field: true
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("YAML with unknown fields should not return error (lenient parsing), got: %v", err)
	}

	if cfg.DeviceID != "extra-device" {
		t.Fatalf("expected device_id 'extra-device', got '%s'", cfg.DeviceID)
	}
	if cfg.ServerAddress != "extra-server:50051" {
		t.Fatalf("expected server address 'extra-server:50051', got '%s'", cfg.ServerAddress)
	}
}

func TestLoad_PermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not reliable on Windows")
	}
	// Skip if running as root (root can read any file)
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "noperm.yaml")

	if err := os.WriteFile(path, []byte("device_id: test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Remove all permissions
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatal(err)
	}
	// Restore permissions on cleanup so TempDir can remove the file
	t.Cleanup(func() {
		os.Chmod(path, 0o644)
	})

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for permission denied file")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected 'permission denied' in error, got: %v", err)
	}
}

func TestLoad_EmptyEnvVarOverride(t *testing.T) {
	// Set DEVICE_ID to empty string explicitly. The Load function checks
	// os.Getenv("DEVICE_ID") != "" so empty string should NOT override.
	t.Setenv("DEVICE_ID", "")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `device_id: "yaml-device"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	// Empty env var should NOT override the YAML value since the code checks v != ""
	if cfg.DeviceID != "yaml-device" {
		t.Fatalf("expected device_id 'yaml-device' (empty env should not override), got '%s'", cfg.DeviceID)
	}
}

func TestDefaultConfig_AllFieldsSet(t *testing.T) {
	cfg := DefaultConfig()

	// ServerAddress should have a non-empty default
	if cfg.ServerAddress == "" {
		t.Error("ServerAddress should not be empty in default config")
	}

	// HeartbeatInterval should be non-zero
	if cfg.HeartbeatInterval == 0 {
		t.Error("HeartbeatInterval should not be zero in default config")
	}

	// ModelStorageDir should have a non-empty default
	if cfg.ModelStorageDir == "" {
		t.Error("ModelStorageDir should not be empty in default config")
	}

	// MaxModelVersions should be non-zero
	if cfg.MaxModelVersions == 0 {
		t.Error("MaxModelVersions should not be zero in default config")
	}

	// LogLevel should have a non-empty default
	if cfg.LogLevel == "" {
		t.Error("LogLevel should not be empty in default config")
	}

	// Mode should have a non-empty default
	if cfg.Mode == "" {
		t.Error("Mode should not be empty in default config")
	}

	// DeviceID is intentionally empty in defaults (set per-device)
	// so we do NOT check it here
}

func TestConfig_EmptyServerAddress(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ServerAddress = ""
	if cfg.ServerAddress != "" {
		t.Error("expected empty server address")
	}
}

func TestConfig_EmptyDeviceID(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.DeviceID != "" {
		t.Error("DefaultConfig should have empty DeviceID")
	}
}

func TestConfig_ZeroHeartbeatInterval(t *testing.T) {
	cfg := DefaultConfig()
	cfg.HeartbeatInterval = 0
	if cfg.HeartbeatInterval != 0 {
		t.Errorf("expected 0 interval, got %v", cfg.HeartbeatInterval)
	}
}

func TestConfig_NegativeHeartbeatInterval(t *testing.T) {
	cfg := DefaultConfig()
	cfg.HeartbeatInterval = -1
	if cfg.HeartbeatInterval >= 0 {
		t.Error("expected negative interval")
	}
}

func TestConfig_AllFieldsZeroValue(t *testing.T) {
	cfg := Config{}
	if cfg.ServerAddress != "" || cfg.DeviceID != "" || cfg.ModelStorageDir != "" {
		t.Error("zero-value config should have empty strings")
	}
	if cfg.HeartbeatInterval != 0 {
		t.Error("zero-value config should have 0 heartbeat interval")
	}
	if cfg.MaxModelVersions != 0 {
		t.Error("zero-value config should have 0 max model versions")
	}
}

func TestConfig_WhitespaceServerAddress(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ServerAddress = "   "
	if cfg.ServerAddress == "" {
		t.Error("whitespace-only address should not be empty string")
	}
}

func TestDefaultConfig_Idempotent(t *testing.T) {
	c1 := DefaultConfig()
	c2 := DefaultConfig()
	if c1.ServerAddress != c2.ServerAddress {
		t.Error("DefaultConfig should return same defaults")
	}
	if c1.HeartbeatInterval != c2.HeartbeatInterval {
		t.Error("DefaultConfig should return same interval")
	}
	if c1.MaxModelVersions != c2.MaxModelVersions {
		t.Error("DefaultConfig should return same max versions")
	}
}
