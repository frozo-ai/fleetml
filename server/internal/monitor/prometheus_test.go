package monitor

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPrometheusExporter_Handler_EmptyMetrics(t *testing.T) {
	exporter := &PrometheusExporter{
		metrics: &fleetMetrics{},
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	exporter.Handler()(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected text/plain content type, got %s", ct)
	}

	body := w.Body.String()
	expectedMetrics := []string{
		"fleetml_devices_total",
		"fleetml_devices_online",
		"fleetml_devices_offline",
		"fleetml_models_total",
		"fleetml_deployments_active",
		"fleetml_deployments_failed",
		"fleetml_heartbeats_1h",
		"fleetml_metrics_collected_at",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(body, metric) {
			t.Errorf("expected metric %q in output", metric)
		}
	}
}

func TestPrometheusExporter_Handler_WithMetrics(t *testing.T) {
	exporter := &PrometheusExporter{
		metrics: &fleetMetrics{
			TotalDevices:   50,
			OnlineDevices:  45,
			OfflineDevices: 5,
			TotalModels:    10,
			ActiveDeploys:  2,
			FailedDeploys:  1,
			HeartbeatCount: 500,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	exporter.Handler()(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "fleetml_devices_total 50") {
		t.Error("expected devices_total 50")
	}
	if !strings.Contains(body, "fleetml_devices_online 45") {
		t.Error("expected devices_online 45")
	}
	if !strings.Contains(body, "fleetml_devices_offline 5") {
		t.Error("expected devices_offline 5")
	}
	if !strings.Contains(body, "fleetml_models_total 10") {
		t.Error("expected models_total 10")
	}
	if !strings.Contains(body, "fleetml_deployments_active 2") {
		t.Error("expected deployments_active 2")
	}
}

func TestPrometheusExporter_Handler_HasHelp(t *testing.T) {
	exporter := &PrometheusExporter{
		metrics: &fleetMetrics{},
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	exporter.Handler()(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "# HELP fleetml_devices_total") {
		t.Error("expected HELP line for devices_total")
	}
	if !strings.Contains(body, "# TYPE fleetml_devices_total gauge") {
		t.Error("expected TYPE line for devices_total")
	}
}

func TestPrometheusExporter_JSONHandler(t *testing.T) {
	exporter := &PrometheusExporter{
		metrics: &fleetMetrics{
			TotalDevices:  25,
			OnlineDevices: 20,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics/json", nil)
	w := httptest.NewRecorder()

	exporter.JSONHandler()(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"devices_total":25`) {
		t.Error("expected devices_total:25 in JSON")
	}
}

func TestNewPrometheusExporter(t *testing.T) {
	exporter := NewPrometheusExporter(nil, nil)
	if exporter == nil {
		t.Fatal("expected non-nil exporter")
	}
	if exporter.metrics == nil {
		t.Fatal("expected non-nil initial metrics")
	}
}
