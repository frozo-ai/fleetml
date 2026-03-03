package model

import (
	"os"
	"path/filepath"
	"testing"
)

func TestONNXSubprocessRuntime_Name(t *testing.T) {
	r := NewONNXSubprocessRuntime()
	if r.Name() != "onnx" {
		t.Errorf("expected 'onnx', got %q", r.Name())
	}
}

func TestONNXSubprocessRuntime_LoadNonexistent(t *testing.T) {
	r := NewONNXSubprocessRuntime()
	if err := r.Load("/nonexistent/model.onnx"); err == nil {
		t.Error("expected error loading nonexistent model")
	}
}

func TestONNXSubprocessRuntime_LoadDirectory(t *testing.T) {
	dir := t.TempDir()
	r := NewONNXSubprocessRuntime()
	if err := r.Load(dir); err == nil {
		t.Error("expected error loading directory as model")
	}
}

func TestONNXSubprocessRuntime_LoadAndInfer(t *testing.T) {
	// Create a temp file to act as a model
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "test.onnx")
	if err := os.WriteFile(modelPath, []byte("model-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewONNXSubprocessRuntime()
	if err := r.Load(modelPath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !r.IsLoaded() {
		t.Error("expected IsLoaded to be true")
	}

	if r.ModelPath() != modelPath {
		t.Errorf("expected model path %q, got %q", modelPath, r.ModelPath())
	}

	// Infer should work (fallback passthrough since onnx_infer helper isn't present)
	input := []byte("test-input")
	output, err := r.Infer(input)
	if err != nil {
		t.Fatalf("Infer failed: %v", err)
	}
	if string(output) != string(input) {
		t.Error("expected passthrough output in dev mode")
	}
}

func TestONNXSubprocessRuntime_InferWithoutLoad(t *testing.T) {
	r := NewONNXSubprocessRuntime()
	_, err := r.Infer([]byte("test"))
	if err == nil {
		t.Error("expected error inferring without loaded model")
	}
}

func TestONNXSubprocessRuntime_Unload(t *testing.T) {
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "test.onnx")
	os.WriteFile(modelPath, []byte("data"), 0o644)

	r := NewONNXSubprocessRuntime()
	r.Load(modelPath)

	if err := r.Unload(); err != nil {
		t.Fatalf("Unload failed: %v", err)
	}

	if r.IsLoaded() {
		t.Error("expected IsLoaded to be false after Unload")
	}
	if r.ModelPath() != "" {
		t.Error("expected empty ModelPath after Unload")
	}
	if r.Metadata() != nil {
		t.Error("expected nil Metadata after Unload")
	}
}

func TestONNXSubprocessRuntime_IsSupported(t *testing.T) {
	r := NewONNXSubprocessRuntime()
	if !r.IsSupported() {
		t.Error("ONNX runtime should be supported on all platforms")
	}
}

func TestONNXSubprocessRuntime_ConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "test.onnx")
	os.WriteFile(modelPath, []byte("data"), 0o644)

	r := NewONNXSubprocessRuntime()
	r.Load(modelPath)

	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 100; j++ {
				r.Infer([]byte("test"))
				r.IsLoaded()
				r.ModelPath()
			}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
