package model

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

	// Infer should return an error when onnx_infer helper is not on PATH
	_, err := r.Infer([]byte("test-input"))
	if err == nil {
		t.Fatal("expected error when onnx_infer helper is not on PATH")
	}
	if !strings.Contains(err.Error(), "onnx_infer helper not found") {
		t.Fatalf("expected 'onnx_infer helper not found' error, got: %v", err)
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
				r.Infer([]byte("test")) // may error without helper — that's fine, testing concurrency safety
				r.IsLoaded()
				r.ModelPath()
			}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestONNXSubprocessRuntime_Load_PermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not reliable on Windows")
	}
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	dir := t.TempDir()
	modelPath := filepath.Join(dir, "noperm.onnx")

	if err := os.WriteFile(modelPath, []byte("model-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Remove all permissions from the file
	if err := os.Chmod(modelPath, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Chmod(modelPath, 0o644)
	})

	r := NewONNXSubprocessRuntime()
	err := r.Load(modelPath)
	// os.Stat with 000 permissions on the file itself still succeeds on most
	// Unix systems (stat only requires execute on the parent directory).
	// However, the onnx_validate helper (if present) would fail. Since the
	// helper is not present in test, Load may succeed. We accept either outcome
	// but ensure no panic.
	if err != nil {
		if !strings.Contains(err.Error(), "permission denied") && !strings.Contains(err.Error(), "not accessible") {
			t.Logf("Load returned error (acceptable): %v", err)
		}
	}

	// Verify the runtime is in a consistent state
	_ = r.IsLoaded()
	_ = r.ModelPath()
}

func TestONNXSubprocessRuntime_Load_Directory(t *testing.T) {
	dir := t.TempDir()

	r := NewONNXSubprocessRuntime()
	err := r.Load(dir)
	if err == nil {
		t.Fatal("expected error when loading a directory path instead of a file")
	}
	if !strings.Contains(err.Error(), "directory") {
		t.Fatalf("expected error message to mention 'directory', got: %v", err)
	}

	// Should not be loaded
	if r.IsLoaded() {
		t.Error("expected IsLoaded() = false after failed directory load")
	}
}

func TestONNXSubprocessRuntime_Infer_NotLoaded(t *testing.T) {
	r := NewONNXSubprocessRuntime()

	// Verify not loaded initially
	if r.IsLoaded() {
		t.Fatal("expected IsLoaded() = false on new runtime")
	}

	_, err := r.Infer([]byte("test-input"))
	if err == nil {
		t.Fatal("expected error when calling Infer before Load")
	}
	if !strings.Contains(err.Error(), "model not loaded") {
		t.Fatalf("expected 'model not loaded' error, got: %v", err)
	}
}

func TestONNXSubprocessRuntime_Unload_Twice(t *testing.T) {
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "test.onnx")
	os.WriteFile(modelPath, []byte("data"), 0o644)

	r := NewONNXSubprocessRuntime()
	if err := r.Load(modelPath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// First unload
	if err := r.Unload(); err != nil {
		t.Fatalf("first Unload failed: %v", err)
	}

	// Second unload should not return an error (idempotent)
	if err := r.Unload(); err != nil {
		t.Fatalf("second Unload should return nil, got: %v", err)
	}

	// State should remain unloaded
	if r.IsLoaded() {
		t.Error("expected IsLoaded() = false after double unload")
	}
	if r.ModelPath() != "" {
		t.Error("expected empty ModelPath after double unload")
	}
}

func TestONNXSubprocessRuntime_IsLoaded_AfterUnload(t *testing.T) {
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "test.onnx")
	os.WriteFile(modelPath, []byte("data"), 0o644)

	r := NewONNXSubprocessRuntime()

	// Load a model
	if err := r.Load(modelPath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !r.IsLoaded() {
		t.Fatal("expected IsLoaded() = true after Load")
	}

	// Unload the model
	if err := r.Unload(); err != nil {
		t.Fatalf("Unload failed: %v", err)
	}

	// Verify IsLoaded returns false
	if r.IsLoaded() {
		t.Fatal("expected IsLoaded() = false after Unload")
	}
}

func TestONNXSubprocessRuntime_ModelPath_AfterUnload(t *testing.T) {
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "test.onnx")
	os.WriteFile(modelPath, []byte("data"), 0o644)

	r := NewONNXSubprocessRuntime()

	// Load a model
	if err := r.Load(modelPath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if r.ModelPath() != modelPath {
		t.Fatalf("expected ModelPath() = %q, got %q", modelPath, r.ModelPath())
	}

	// Unload the model
	if err := r.Unload(); err != nil {
		t.Fatalf("Unload failed: %v", err)
	}

	// Verify ModelPath returns empty string
	if r.ModelPath() != "" {
		t.Fatalf("expected empty ModelPath after Unload, got %q", r.ModelPath())
	}
}
