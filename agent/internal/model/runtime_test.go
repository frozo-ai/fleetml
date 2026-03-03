package model

import "testing"

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
	r := NewONNXRuntime()

	if err := r.Load("/tmp/test-model.onnx"); err != nil {
		t.Fatalf("unexpected error loading model: %v", err)
	}

	if err := r.Unload(); err != nil {
		t.Fatalf("unexpected error unloading model: %v", err)
	}
}

func TestONNXRuntime_Infer(t *testing.T) {
	r := NewONNXRuntime()
	r.Load("/tmp/test-model.onnx")

	input := []byte("test input")
	output, err := r.Infer(input)
	if err != nil {
		t.Fatalf("unexpected error during inference: %v", err)
	}
	if output == nil {
		t.Fatal("expected non-nil output")
	}
}
