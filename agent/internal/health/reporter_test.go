package health

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewReporter(t *testing.T) {
	interval := 10 * time.Second
	r := NewReporter(interval)
	if r == nil {
		t.Fatal("expected non-nil reporter")
	}
	if r.interval != interval {
		t.Fatalf("expected interval %v, got %v", interval, r.interval)
	}
}

func TestReporter_Interval(t *testing.T) {
	interval := 45 * time.Second
	r := NewReporter(interval)
	if got := r.Interval(); got != interval {
		t.Fatalf("expected Interval() = %v, got %v", interval, got)
	}
}

func TestReporter_Collect(t *testing.T) {
	r := NewReporter(5 * time.Second)
	metrics, err := r.Collect()
	if err != nil {
		t.Fatalf("Collect() returned error: %v", err)
	}
	if metrics == nil {
		t.Fatal("Collect() returned nil metrics")
	}

	if metrics.CPUPercent < 0 {
		t.Errorf("CPUPercent should be >= 0, got %f", metrics.CPUPercent)
	}
	if metrics.RAMMBUsed < 0 {
		t.Errorf("RAMMBUsed should be >= 0, got %d", metrics.RAMMBUsed)
	}
	if metrics.DiskPercent < 0 || metrics.DiskPercent > 100 {
		t.Errorf("DiskPercent should be between 0 and 100, got %f", metrics.DiskPercent)
	}
	if metrics.UptimeHours < 0 {
		t.Errorf("UptimeHours should be >= 0, got %f", metrics.UptimeHours)
	}
}

func TestSystemMetrics_JSONRoundTrip(t *testing.T) {
	original := SystemMetrics{
		CPUPercent:   42.5,
		GPUPercent:   10.0,
		RAMMBUsed:    2048,
		DiskPercent:  65.3,
		TemperatureC: 55.0,
		UptimeHours:  123.45,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded SystemMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.CPUPercent != original.CPUPercent {
		t.Errorf("CPUPercent: expected %f, got %f", original.CPUPercent, decoded.CPUPercent)
	}
	if decoded.GPUPercent != original.GPUPercent {
		t.Errorf("GPUPercent: expected %f, got %f", original.GPUPercent, decoded.GPUPercent)
	}
	if decoded.RAMMBUsed != original.RAMMBUsed {
		t.Errorf("RAMMBUsed: expected %d, got %d", original.RAMMBUsed, decoded.RAMMBUsed)
	}
	if decoded.DiskPercent != original.DiskPercent {
		t.Errorf("DiskPercent: expected %f, got %f", original.DiskPercent, decoded.DiskPercent)
	}
	if decoded.TemperatureC != original.TemperatureC {
		t.Errorf("TemperatureC: expected %f, got %f", original.TemperatureC, decoded.TemperatureC)
	}
	if decoded.UptimeHours != original.UptimeHours {
		t.Errorf("UptimeHours: expected %f, got %f", original.UptimeHours, decoded.UptimeHours)
	}
}

func TestSystemMetrics_DefaultValues(t *testing.T) {
	var m SystemMetrics

	if m.CPUPercent != 0 {
		t.Errorf("expected default CPUPercent 0, got %f", m.CPUPercent)
	}
	if m.GPUPercent != 0 {
		t.Errorf("expected default GPUPercent 0, got %f", m.GPUPercent)
	}
	if m.RAMMBUsed != 0 {
		t.Errorf("expected default RAMMBUsed 0, got %d", m.RAMMBUsed)
	}
	if m.DiskPercent != 0 {
		t.Errorf("expected default DiskPercent 0, got %f", m.DiskPercent)
	}
	if m.TemperatureC != 0 {
		t.Errorf("expected default TemperatureC 0, got %f", m.TemperatureC)
	}
	if m.UptimeHours != 0 {
		t.Errorf("expected default UptimeHours 0, got %f", m.UptimeHours)
	}
}

func TestReporter_Collect_CPURange(t *testing.T) {
	r := NewReporter(5 * time.Second)
	metrics, err := r.Collect()
	if err != nil {
		t.Fatalf("Collect() returned error: %v", err)
	}

	if metrics.CPUPercent < 0 || metrics.CPUPercent > 100 {
		t.Errorf("CPUPercent should be between 0 and 100, got %f", metrics.CPUPercent)
	}
}

func TestReporter_Collect_DiskRange(t *testing.T) {
	r := NewReporter(5 * time.Second)
	metrics, err := r.Collect()
	if err != nil {
		t.Fatalf("Collect() returned error: %v", err)
	}

	if metrics.DiskPercent < 0 || metrics.DiskPercent > 100 {
		t.Errorf("DiskPercent should be between 0 and 100, got %f", metrics.DiskPercent)
	}
}

func TestReporter_Collect_MultipleCollections(t *testing.T) {
	r := NewReporter(5 * time.Second)

	metrics1, err := r.Collect()
	if err != nil {
		t.Fatalf("first Collect() returned error: %v", err)
	}
	if metrics1 == nil {
		t.Fatal("first Collect() returned nil")
	}

	metrics2, err := r.Collect()
	if err != nil {
		t.Fatalf("second Collect() returned error: %v", err)
	}
	if metrics2 == nil {
		t.Fatal("second Collect() returned nil")
	}

	// Both should have valid uptime (non-negative)
	if metrics1.UptimeHours < 0 {
		t.Errorf("first collection UptimeHours should be >= 0, got %f", metrics1.UptimeHours)
	}
	if metrics2.UptimeHours < 0 {
		t.Errorf("second collection UptimeHours should be >= 0, got %f", metrics2.UptimeHours)
	}

	// Both should have valid RAM readings
	if metrics1.RAMMBUsed < 0 {
		t.Errorf("first collection RAMMBUsed should be >= 0, got %d", metrics1.RAMMBUsed)
	}
	if metrics2.RAMMBUsed < 0 {
		t.Errorf("second collection RAMMBUsed should be >= 0, got %d", metrics2.RAMMBUsed)
	}
}

func TestReporter_ZeroInterval(t *testing.T) {
	r := NewReporter(0)
	if r == nil {
		t.Fatal("expected non-nil reporter with zero interval")
	}
}

func TestReporter_NegativeInterval(t *testing.T) {
	r := NewReporter(-1 * time.Second)
	if r == nil {
		t.Fatal("expected non-nil reporter with negative interval")
	}
}

func TestReporter_Collect_ReturnsNonNil(t *testing.T) {
	r := NewReporter(5 * time.Second)
	metrics, err := r.Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}
	if metrics == nil {
		t.Fatal("expected non-nil metrics")
	}
}

func TestSystemMetrics_JSONMarshalEmpty(t *testing.T) {
	m := SystemMetrics{}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal empty metrics: %v", err)
	}
	var decoded SystemMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal empty metrics: %v", err)
	}
	if decoded.CPUPercent != 0 || decoded.RAMMBUsed != 0 {
		t.Error("zero-value metrics should round-trip cleanly")
	}
}

func TestReporter_Collect_UptimePositive(t *testing.T) {
	r := NewReporter(5 * time.Second)
	metrics, err := r.Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}
	if metrics.UptimeHours < 0 {
		t.Errorf("uptime should be non-negative, got %f", metrics.UptimeHours)
	}
}

func TestReporter_Collect_RAMNonNegative(t *testing.T) {
	r := NewReporter(5 * time.Second)
	metrics, err := r.Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}
	if metrics.RAMMBUsed < 0 {
		t.Errorf("RAM should be non-negative, got %d", metrics.RAMMBUsed)
	}
}
