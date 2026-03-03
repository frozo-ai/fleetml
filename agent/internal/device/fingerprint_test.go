package device

import (
	"runtime"
	"testing"
)

func TestFingerprint_Basic(t *testing.T) {
	info, err := Fingerprint("test-device-001")
	if err != nil {
		t.Fatal(err)
	}

	if info.DeviceID != "test-device-001" {
		t.Fatalf("expected device ID 'test-device-001', got '%s'", info.DeviceID)
	}

	if info.Arch != runtime.GOARCH {
		t.Fatalf("expected arch '%s', got '%s'", runtime.GOARCH, info.Arch)
	}

	if info.RAMMB <= 0 {
		t.Fatalf("expected positive RAM, got %d", info.RAMMB)
	}

	if info.DiskGB <= 0 {
		t.Fatalf("expected positive disk, got %d", info.DiskGB)
	}

	if info.OS == "" {
		t.Fatal("expected non-empty OS")
	}

	if info.Labels == nil {
		t.Fatal("expected non-nil labels map")
	}
}

func TestFingerprint_RuntimeSelection(t *testing.T) {
	info, err := Fingerprint("test-device")
	if err != nil {
		t.Fatal(err)
	}

	// Runtime should be one of the valid options
	validRuntimes := map[string]bool{
		"onnx":      true,
		"tensorrt":  true,
		"openvino":  true,
		"tflite":    true,
	}

	if !validRuntimes[info.Runtime] {
		t.Fatalf("unexpected runtime '%s'", info.Runtime)
	}
}

func TestSelectRuntime(t *testing.T) {
	tests := []struct {
		gpu     string
		arch    string
		want    string
	}{
		{"nvidia", "amd64", "tensorrt"},
		{"nvidia", "arm64", "tensorrt"},
		{"intel", "amd64", "openvino"},
		{"none", "arm64", "tflite"},
		{"none", "arm", "tflite"},
		{"none", "amd64", "onnx"},
	}

	for _, tt := range tests {
		got := selectRuntime(tt.gpu, tt.arch)
		if got != tt.want {
			t.Errorf("selectRuntime(%q, %q) = %q, want %q", tt.gpu, tt.arch, got, tt.want)
		}
	}
}

func TestDetectGPU(t *testing.T) {
	gpu := detectGPU()
	// Just verify it returns a valid value
	validTypes := map[string]bool{
		"nvidia": true,
		"intel":  true,
		"none":   true,
	}
	if !validTypes[gpu] {
		t.Fatalf("unexpected GPU type '%s'", gpu)
	}
}
