package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDevice_JSONMarshalling(t *testing.T) {
	now := time.Now()
	d := Device{
		ID:            "uuid-123",
		DeviceID:      "jetson-001",
		Name:          "Jetson Nano #1",
		Status:        "healthy",
		Arch:          "arm64",
		GPUType:       "nvidia-jetson",
		Runtime:       "tensorrt",
		RAMMB:         4096,
		DiskGB:        64,
		OS:            "linux",
		HardwareModel: "jetson-nano",
		Labels:        map[string]string{"zone": "factory-1"},
		RegisteredAt:  now,
		UpdatedAt:     now,
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal device: %v", err)
	}

	var decoded Device
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal device: %v", err)
	}

	if decoded.DeviceID != "jetson-001" {
		t.Errorf("expected device_id jetson-001, got %s", decoded.DeviceID)
	}
	if decoded.Runtime != "tensorrt" {
		t.Errorf("expected runtime tensorrt, got %s", decoded.Runtime)
	}
	if decoded.Labels["zone"] != "factory-1" {
		t.Errorf("expected label zone=factory-1, got %s", decoded.Labels["zone"])
	}
}

func TestDevice_OptionalFields(t *testing.T) {
	d := Device{
		ID:       "uuid-123",
		DeviceID: "rpi-001",
		Status:   "registered",
	}

	if d.FleetID != nil {
		t.Error("expected nil FleetID")
	}
	if d.LastHeartbeat != nil {
		t.Error("expected nil LastHeartbeat")
	}
	if d.CPUPercent != nil {
		t.Error("expected nil CPUPercent")
	}
}

func TestDevice_StatusValues(t *testing.T) {
	validStatuses := []string{"registered", "healthy", "warning", "offline", "decommissioned"}
	for _, s := range validStatuses {
		d := Device{Status: s}
		if d.Status != s {
			t.Errorf("expected status %s, got %s", s, d.Status)
		}
	}
}

func TestModel_JSONMarshalling(t *testing.T) {
	m := Model{
		ID:           "model-uuid",
		Name:         "mobilenet",
		Version:      "v2.1",
		Format:       "onnx",
		ArtifactURL:  "s3://fleetml-models/mobilenet/v2.1/model.onnx",
		ArtifactSize: 13456789,
		Checksum:     "sha256:abcdef1234",
		Description:  "MobileNet V2 for object detection",
		Tags:         []string{"production", "arm64"},
		Metadata:     map[string]interface{}{"framework": "pytorch"},
		CreatedAt:    time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal model: %v", err)
	}

	var decoded Model
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal model: %v", err)
	}

	if decoded.Name != "mobilenet" {
		t.Errorf("expected name mobilenet, got %s", decoded.Name)
	}
	if decoded.Format != "onnx" {
		t.Errorf("expected format onnx, got %s", decoded.Format)
	}
	if len(decoded.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(decoded.Tags))
	}
}

func TestModel_CompiledVariants(t *testing.T) {
	m := Model{
		Name:    "resnet50",
		Version: "v1",
		Format:  "onnx",
		CompiledVariants: []CompiledVariant{
			{Runtime: "tensorrt", ArtifactURL: "s3://bucket/resnet50/compiled/tensorrt/model.trt", Checksum: "sha256:trt123"},
			{Runtime: "openvino", ArtifactURL: "s3://bucket/resnet50/compiled/openvino/model.xml", Checksum: "sha256:ov456"},
			{Runtime: "tflite", ArtifactURL: "s3://bucket/resnet50/compiled/tflite/model.tflite", Checksum: "sha256:tfl789"},
		},
	}

	if len(m.CompiledVariants) != 3 {
		t.Fatalf("expected 3 variants, got %d", len(m.CompiledVariants))
	}

	// Verify we can find TensorRT variant
	found := false
	for _, v := range m.CompiledVariants {
		if v.Runtime == "tensorrt" {
			found = true
			if v.Checksum != "sha256:trt123" {
				t.Errorf("expected tensorrt checksum sha256:trt123, got %s", v.Checksum)
			}
		}
	}
	if !found {
		t.Error("tensorrt variant not found")
	}
}

func TestCompiledVariant_JSONMarshalling(t *testing.T) {
	v := CompiledVariant{
		Runtime:     "tensorrt",
		ArtifactURL: "s3://bucket/compiled/model.trt",
		Checksum:    "sha256:abc123",
	}

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal variant: %v", err)
	}

	var decoded CompiledVariant
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal variant: %v", err)
	}

	if decoded.Runtime != "tensorrt" {
		t.Errorf("expected runtime tensorrt, got %s", decoded.Runtime)
	}
	if decoded.Checksum != "sha256:abc123" {
		t.Errorf("expected checksum sha256:abc123, got %s", decoded.Checksum)
	}
}

func TestCompiledVariant_SliceJSONRoundtrip(t *testing.T) {
	variants := []CompiledVariant{
		{Runtime: "mock", ArtifactURL: "s3://b/mock.onnx", Checksum: "sha256:m1"},
		{Runtime: "tensorrt", ArtifactURL: "s3://b/model.trt", Checksum: "sha256:t2"},
	}

	data, err := json.Marshal(variants)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded []CompiledVariant
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(decoded))
	}
	if decoded[0].Runtime != "mock" {
		t.Errorf("expected first variant runtime 'mock', got %s", decoded[0].Runtime)
	}
}

func TestDeployment_Fields(t *testing.T) {
	now := time.Now()
	d := Deployment{
		ID:               "deploy-123",
		ModelID:          "model-456",
		TargetType:       "fleet",
		State:            "rolling_out",
		TotalDevices:     10,
		CompletedDevices: 3,
		FailedDevices:    1,
		QueuedDevices:    2,
		DeploymentPolicy: "canary",
		StartedAt:        &now,
		CreatedAt:        now,
	}

	if d.TotalDevices != 10 {
		t.Errorf("expected 10 total devices, got %d", d.TotalDevices)
	}
	if d.CompletedDevices+d.FailedDevices+d.QueuedDevices > d.TotalDevices {
		t.Error("sum of completed+failed+queued exceeds total")
	}
}

func TestDeployment_States(t *testing.T) {
	validStates := []string{"pending", "rolling_out", "completed", "failed", "rolled_back", "cancelled"}
	for _, s := range validStates {
		d := Deployment{State: s}
		if d.State != s {
			t.Errorf("expected state %s, got %s", s, d.State)
		}
	}
}

func TestCanaryConfig_Stages(t *testing.T) {
	cfg := CanaryConfig{
		Stages: []CanaryStage{
			{Percent: 5, Duration: "5m", SuccessMetric: "error_rate < 1%"},
			{Percent: 50, Duration: "10m", SuccessMetric: "latency_p99 < 100ms"},
			{Percent: 100, Duration: "15m"},
		},
	}

	if len(cfg.Stages) != 3 {
		t.Fatalf("expected 3 stages, got %d", len(cfg.Stages))
	}

	// Verify progressive increase
	for i := 1; i < len(cfg.Stages); i++ {
		if cfg.Stages[i].Percent <= cfg.Stages[i-1].Percent {
			t.Errorf("stage %d percent (%d) should be > stage %d percent (%d)",
				i, cfg.Stages[i].Percent, i-1, cfg.Stages[i-1].Percent)
		}
	}
}

func TestCanaryConfig_JSONMarshalling(t *testing.T) {
	cfg := CanaryConfig{
		Stages: []CanaryStage{
			{Percent: 5, Duration: "5m"},
			{Percent: 100, Duration: "30m"},
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded CanaryConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded.Stages) != 2 {
		t.Fatalf("expected 2 stages, got %d", len(decoded.Stages))
	}
	if decoded.Stages[0].Percent != 5 {
		t.Errorf("expected first stage 5%%, got %d%%", decoded.Stages[0].Percent)
	}
}

func TestUser_PasswordNotInJSON(t *testing.T) {
	u := User{
		ID:           "user-123",
		Email:        "admin@fleetml.io",
		Name:         "Admin",
		PasswordHash: "hashed-secret",
		Role:         "admin",
	}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal user: %v", err)
	}

	// PasswordHash should NOT appear in JSON (tagged with json:"-")
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := m["PasswordHash"]; ok {
		t.Error("PasswordHash should not be in JSON output")
	}
	if _, ok := m["password_hash"]; ok {
		t.Error("password_hash should not be in JSON output")
	}
}

func TestUser_Roles(t *testing.T) {
	validRoles := []string{"admin", "deployer", "viewer"}
	for _, role := range validRoles {
		u := User{Role: role}
		if u.Role != role {
			t.Errorf("expected role %s, got %s", role, u.Role)
		}
	}
}

func TestDeviceFilter_Defaults(t *testing.T) {
	f := DeviceFilter{}
	if f.Limit != 0 {
		t.Errorf("expected default limit 0, got %d", f.Limit)
	}
	if f.Status != "" {
		t.Errorf("expected empty status, got %q", f.Status)
	}
}

func TestModelFilter_Defaults(t *testing.T) {
	f := ModelFilter{}
	if f.Limit != 0 {
		t.Errorf("expected default limit 0, got %d", f.Limit)
	}
	if f.Name != "" {
		t.Errorf("expected empty name, got %q", f.Name)
	}
}

func TestFleet_JSONMarshalling(t *testing.T) {
	f := Fleet{
		ID:          "fleet-123",
		Name:        "factory-floor",
		Description: "All factory floor devices",
		Labels:      map[string]string{"env": "production"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("marshal fleet: %v", err)
	}

	var decoded Fleet
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal fleet: %v", err)
	}

	if decoded.Name != "factory-floor" {
		t.Errorf("expected name factory-floor, got %s", decoded.Name)
	}
	if decoded.Labels["env"] != "production" {
		t.Errorf("expected label env=production")
	}
}
