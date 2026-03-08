package model

import (
	"os"
	"path/filepath"
	"testing"
)

func TestONNXRuntime_Name(t *testing.T) {
	r := NewONNXRuntime()
	if r.Name() != "onnx" {
		t.Fatalf("expected name 'onnx', got '%s'", r.Name())
	}
}

func TestONNXRuntime_IsSupported(t *testing.T) {
	r := NewONNXRuntime()
	if !r.IsSupported() {
		t.Fatal("ONNX runtime should be supported on all platforms")
	}
}

func TestONNXRuntime_LoadAndUnload(t *testing.T) {
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "test-model.onnx")
	if err := os.WriteFile(modelPath, []byte("fake-model"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewONNXRuntime()

	if err := r.Load(modelPath); err != nil {
		t.Fatalf("unexpected error loading model: %v", err)
	}

	if err := r.Unload(); err != nil {
		t.Fatalf("unexpected error unloading model: %v", err)
	}
}

func TestONNXRuntime_Infer_NoHelper(t *testing.T) {
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "test-model.onnx")
	if err := os.WriteFile(modelPath, []byte("fake-model"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewONNXRuntime()
	r.Load(modelPath)

	// Without onnx_infer on PATH, Infer returns an error
	_, err := r.Infer([]byte("test input"))
	if err == nil {
		t.Fatal("expected error when onnx_infer helper is not on PATH")
	}
}
