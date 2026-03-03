package model

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateChecksum_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "model.onnx")

	content := []byte("test model content")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	h := sha256.Sum256(content)
	checksum := "sha256:" + hex.EncodeToString(h[:])

	loader := NewLoader(dir, 3)
	if err := loader.ValidateChecksum(path, checksum); err != nil {
		t.Fatalf("expected valid checksum, got error: %v", err)
	}
}

func TestValidateChecksum_Invalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "model.onnx")

	if err := os.WriteFile(path, []byte("test model"), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := NewLoader(dir, 3)
	err := loader.ValidateChecksum(path, "sha256:0000000000000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Fatal("expected checksum mismatch error")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected 'checksum mismatch' in error, got: %v", err)
	}
}

func TestValidateChecksum_MissingFile(t *testing.T) {
	loader := NewLoader(t.TempDir(), 3)
	err := loader.ValidateChecksum("/nonexistent/path", "sha256:abc")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidateChecksum_WithoutPrefix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "model.onnx")

	content := []byte("test")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	h := sha256.Sum256(content)
	checksum := hex.EncodeToString(h[:]) // no sha256: prefix

	loader := NewLoader(dir, 3)
	if err := loader.ValidateChecksum(path, checksum); err != nil {
		t.Fatalf("expected valid checksum without prefix, got error: %v", err)
	}
}

func TestComputeChecksum(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "model.onnx")

	content := []byte("hello world")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	checksum, err := ComputeChecksum(path)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(checksum, "sha256:") {
		t.Fatalf("expected sha256: prefix, got: %s", checksum)
	}

	h := sha256.Sum256(content)
	expected := "sha256:" + hex.EncodeToString(h[:])
	if checksum != expected {
		t.Fatalf("expected %s, got %s", expected, checksum)
	}
}

func TestSaveModel(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader(dir, 3)

	content := strings.NewReader("model data")
	path, err := loader.SaveModel("test-model", "1.0", content)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("model file not created at %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "model data" {
		t.Fatalf("expected 'model data', got '%s'", data)
	}
}

func TestLoad_ModelExists(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader(dir, 3)

	// Create model file
	modelDir := filepath.Join(dir, "test-model")
	os.MkdirAll(modelDir, 0o755)
	os.WriteFile(filepath.Join(modelDir, "1.0.onnx"), []byte("data"), 0o644)

	path, err := loader.Load("test-model", "1.0")
	if err != nil {
		t.Fatal(err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}
}

func TestLoad_ModelNotExists(t *testing.T) {
	loader := NewLoader(t.TempDir(), 3)

	_, err := loader.Load("nonexistent", "1.0")
	if err == nil {
		t.Fatal("expected error for missing model")
	}
}
