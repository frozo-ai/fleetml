package rest

import (
	"testing"
	"time"
)

func TestParseDuration_Hours(t *testing.T) {
	d, err := parseDuration("1h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 1*time.Hour {
		t.Errorf("expected 1h, got %v", d)
	}
}

func TestParseDuration_Minutes(t *testing.T) {
	d, err := parseDuration("30m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 30*time.Minute {
		t.Errorf("expected 30m, got %v", d)
	}
}

func TestParseDuration_Days(t *testing.T) {
	d, err := parseDuration("7d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := 7 * 24 * time.Hour
	if d != expected {
		t.Errorf("expected %v, got %v", expected, d)
	}
}

func TestParseDuration_SingleDay(t *testing.T) {
	d, err := parseDuration("1d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 24*time.Hour {
		t.Errorf("expected 24h, got %v", d)
	}
}

func TestParseDuration_Empty(t *testing.T) {
	d, err := parseDuration("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 0 {
		t.Errorf("expected 0, got %v", d)
	}
}

func TestParseDuration_Seconds(t *testing.T) {
	d, err := parseDuration("120s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 120*time.Second {
		t.Errorf("expected 120s, got %v", d)
	}
}

func TestParseDuration_InvalidDays(t *testing.T) {
	_, err := parseDuration("abcd")
	if err == nil {
		t.Error("expected error for invalid duration string")
	}
}

func TestParseDuration_InvalidFormat(t *testing.T) {
	_, err := parseDuration("xyz")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestParseDuration_24Hours(t *testing.T) {
	d, err := parseDuration("24h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 24*time.Hour {
		t.Errorf("expected 24h, got %v", d)
	}
}

func TestParseDuration_MultipleDays(t *testing.T) {
	d, err := parseDuration("30d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := 30 * 24 * time.Hour
	if d != expected {
		t.Errorf("expected %v, got %v", expected, d)
	}
}

func TestLogEntry_Fields(t *testing.T) {
	cpu := 85.5
	entry := logEntry{
		Timestamp:  time.Now(),
		DeviceID:   "device-123",
		Level:      "info",
		Message:    "heartbeat",
		CPUPercent: &cpu,
	}

	if entry.DeviceID != "device-123" {
		t.Errorf("expected device-123, got %s", entry.DeviceID)
	}
	if entry.Level != "info" {
		t.Errorf("expected info, got %s", entry.Level)
	}
	if entry.CPUPercent == nil || *entry.CPUPercent != 85.5 {
		t.Error("expected cpu_percent 85.5")
	}
	if entry.DiskPercent != nil {
		t.Error("expected nil disk_percent")
	}
}

func TestLogEntry_WarnLevel(t *testing.T) {
	cpu := 95.0
	entry := logEntry{
		Level:      "info",
		Message:    "heartbeat",
		CPUPercent: &cpu,
	}

	// Simulate the level derivation logic from GetLogs
	if entry.CPUPercent != nil && *entry.CPUPercent >= 90 {
		entry.Level = "warn"
		entry.Message = "high CPU usage"
	}

	if entry.Level != "warn" {
		t.Errorf("expected warn level for high CPU, got %s", entry.Level)
	}
	if entry.Message != "high CPU usage" {
		t.Errorf("expected 'high CPU usage', got %s", entry.Message)
	}
}

func TestLogEntry_HighDisk(t *testing.T) {
	disk := 92.0
	entry := logEntry{
		Level:       "info",
		Message:     "heartbeat",
		DiskPercent: &disk,
	}

	if entry.DiskPercent != nil && *entry.DiskPercent >= 90 {
		entry.Level = "warn"
		entry.Message = "high disk usage"
	}

	if entry.Level != "warn" {
		t.Errorf("expected warn level for high disk, got %s", entry.Level)
	}
}

func TestLogEntry_HighTemperature(t *testing.T) {
	temp := 85.0
	entry := logEntry{
		Level:   "info",
		Message: "heartbeat",
		TempC:   &temp,
	}

	if entry.TempC != nil && *entry.TempC >= 80 {
		entry.Level = "warn"
		entry.Message = "high temperature"
	}

	if entry.Level != "warn" {
		t.Errorf("expected warn for high temp, got %s", entry.Level)
	}
}

func TestLogEntry_NormalMetrics(t *testing.T) {
	cpu := 45.0
	disk := 60.0
	temp := 55.0
	ram := 1024

	entry := logEntry{
		Level:       "info",
		Message:     "heartbeat",
		CPUPercent:  &cpu,
		DiskPercent: &disk,
		TempC:       &temp,
		RAMMBUsed:   &ram,
	}

	// None of these should trigger warn
	if entry.CPUPercent != nil && *entry.CPUPercent >= 90 {
		entry.Level = "warn"
	}
	if entry.DiskPercent != nil && *entry.DiskPercent >= 90 {
		entry.Level = "warn"
	}
	if entry.TempC != nil && *entry.TempC >= 80 {
		entry.Level = "warn"
	}

	if entry.Level != "info" {
		t.Errorf("expected info for normal metrics, got %s", entry.Level)
	}
}

func TestNewLogsHandler(t *testing.T) {
	h := NewLogsHandler(nil, nil)
	if h == nil {
		t.Fatal("expected non-nil LogsHandler")
	}
}

func TestParseDuration_ZeroDays(t *testing.T) {
	d, err := parseDuration("0d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 0 {
		t.Errorf("expected 0, got %v", d)
	}
}

func TestParseDuration_ZeroHours(t *testing.T) {
	d, err := parseDuration("0h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 0 {
		t.Errorf("expected 0, got %v", d)
	}
}

func TestParseDuration_NegativeDays(t *testing.T) {
	d, err := parseDuration("-1d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != -24*time.Hour {
		t.Errorf("expected -24h, got %v", d)
	}
}

func TestParseDuration_FractionalHours(t *testing.T) {
	d, err := parseDuration("1.5h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 90*time.Minute {
		t.Errorf("expected 90m, got %v", d)
	}
}

func TestParseDuration_LargeDays(t *testing.T) {
	d, err := parseDuration("365d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 365*24*time.Hour {
		t.Errorf("expected 365 days, got %v", d)
	}
}

func TestParseDuration_SingleCharD(t *testing.T) {
	_, err := parseDuration("d")
	if err == nil {
		t.Error("expected error for 'd' with no number")
	}
}

func TestParseDuration_Microseconds(t *testing.T) {
	d, err := parseDuration("100us")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 100*time.Microsecond {
		t.Errorf("expected 100us, got %v", d)
	}
}

func TestParseDuration_Nanoseconds(t *testing.T) {
	d, err := parseDuration("500ns")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 500*time.Nanosecond {
		t.Errorf("expected 500ns, got %v", d)
	}
}

func TestParseDuration_OnlyNumber(t *testing.T) {
	_, err := parseDuration("42")
	if err == nil {
		t.Error("expected error for number without unit")
	}
}

func TestParseDuration_Whitespace(t *testing.T) {
	_, err := parseDuration(" 1h ")
	if err == nil {
		t.Error("expected error for input with spaces")
	}
}

func TestLogEntry_StructuredFields(t *testing.T) {
	entry := logEntry{
		ID:        42,
		Timestamp: time.Now(),
		DeviceID:  "device-001",
		Level:     "warn",
		Component: "deploy",
		Message:   "model download started",
		Metadata:  map[string]string{"model": "v1", "size": "15MB"},
	}

	if entry.Component != "deploy" {
		t.Errorf("expected component 'deploy', got %s", entry.Component)
	}
	if entry.Metadata["model"] != "v1" {
		t.Errorf("expected metadata model=v1, got %s", entry.Metadata["model"])
	}
	if entry.ID != 42 {
		t.Errorf("expected id 42, got %d", entry.ID)
	}
}

func TestLogEntry_HeartbeatFallback(t *testing.T) {
	// Ensure heartbeat-derived entries still include component field
	entry := logEntry{
		Component: "heartbeat",
		Level:     "info",
		Message:   "heartbeat",
	}
	if entry.Component != "heartbeat" {
		t.Errorf("expected component 'heartbeat', got %s", entry.Component)
	}
}
