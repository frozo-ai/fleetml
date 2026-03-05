package model

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestLoader_StorageDir(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader(dir, 3)

	if got := loader.StorageDir(); got != dir {
		t.Fatalf("expected StorageDir() = %q, got %q", dir, got)
	}
}

func TestSaveModel_DirectoryCreation(t *testing.T) {
	dir := t.TempDir()
	// Use a nested subdirectory that does not exist yet as the storage dir.
	// SaveModel creates <storageDir>/<name>/ via MkdirAll.
	nestedDir := filepath.Join(dir, "deep", "nested", "storage")
	loader := NewLoader(nestedDir, 3)

	content := strings.NewReader("model bytes")
	path, err := loader.SaveModel("my-model", "2.0", content)
	if err != nil {
		t.Fatalf("SaveModel to non-existent subdirectory failed: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("model file not created at %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "model bytes" {
		t.Fatalf("expected 'model bytes', got '%s'", data)
	}
}

func TestValidateChecksum_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.onnx")

	// Create a 0-byte file
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	// Compute expected SHA-256 of empty content
	h := sha256.Sum256([]byte{})
	expectedChecksum := "sha256:" + hex.EncodeToString(h[:])

	loader := NewLoader(dir, 3)
	if err := loader.ValidateChecksum(path, expectedChecksum); err != nil {
		t.Fatalf("expected valid checksum for empty file, got error: %v", err)
	}

	// Also verify that a wrong checksum still fails
	err := loader.ValidateChecksum(path, "sha256:0000000000000000000000000000000000000000000000000000000000000001")
	if err == nil {
		t.Fatal("expected checksum mismatch error for wrong checksum on empty file")
	}
}

func TestComputeChecksum_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.onnx")

	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	checksum, err := ComputeChecksum(path)
	if err != nil {
		t.Fatalf("ComputeChecksum on empty file failed: %v", err)
	}

	if !strings.HasPrefix(checksum, "sha256:") {
		t.Fatalf("expected sha256: prefix, got: %s", checksum)
	}

	// Verify it matches the known SHA-256 of empty input
	h := sha256.Sum256([]byte{})
	expected := "sha256:" + hex.EncodeToString(h[:])
	if checksum != expected {
		t.Fatalf("expected %s, got %s", expected, checksum)
	}
}

func TestCleanupPartialDownloads_NoFiles(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader(dir, 3)

	// No .downloading files exist -- should succeed without error
	if err := loader.CleanupPartialDownloads(); err != nil {
		t.Fatalf("CleanupPartialDownloads on empty dir failed: %v", err)
	}
}

func TestCleanupPartialDownloads_RemovesDownloadingFiles(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader(dir, 3)

	// Create some .downloading files and a regular file
	modelDir := filepath.Join(dir, "test-model")
	os.MkdirAll(modelDir, 0o755)
	os.WriteFile(filepath.Join(modelDir, "1.0.onnx.downloading"), []byte("partial"), 0o644)
	os.WriteFile(filepath.Join(modelDir, "2.0.onnx.downloading"), []byte("partial"), 0o644)
	os.WriteFile(filepath.Join(modelDir, "1.0.onnx"), []byte("complete"), 0o644)

	if err := loader.CleanupPartialDownloads(); err != nil {
		t.Fatalf("CleanupPartialDownloads failed: %v", err)
	}

	// .downloading files should be gone
	if _, err := os.Stat(filepath.Join(modelDir, "1.0.onnx.downloading")); !os.IsNotExist(err) {
		t.Error("expected 1.0.onnx.downloading to be removed")
	}
	if _, err := os.Stat(filepath.Join(modelDir, "2.0.onnx.downloading")); !os.IsNotExist(err) {
		t.Error("expected 2.0.onnx.downloading to be removed")
	}

	// Regular file should remain
	if _, err := os.Stat(filepath.Join(modelDir, "1.0.onnx")); os.IsNotExist(err) {
		t.Error("expected 1.0.onnx to remain after cleanup")
	}
}

func TestDownloadFromURL_ContentLengthMismatch(t *testing.T) {
	// Server advertises Content-Length: 1000 but only sends 5 bytes
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "short")
	}))
	defer server.Close()

	dir := t.TempDir()
	loader := NewLoader(dir, 3)

	var lastProgress DownloadProgress
	progressCb := func(p DownloadProgress) {
		lastProgress = p
	}

	// The download itself may succeed or fail depending on how http.Client
	// and io.Copy handle the premature EOF. Either way the test should not panic.
	path, err := loader.DownloadFromURL(context.Background(), "mismatch-model", "1.0", server.URL+"/model.onnx", progressCb)

	if err != nil {
		// An error is acceptable here (e.g., unexpected EOF)
		if !strings.Contains(err.Error(), "EOF") && !strings.Contains(err.Error(), "unexpected") {
			// Any error is fine, we just want to ensure no panic
			t.Logf("DownloadFromURL returned error (acceptable): %v", err)
		}
		return
	}

	// If it succeeded, verify the progress reported a mismatch
	if lastProgress.TotalBytes != 1000 {
		t.Errorf("expected TotalBytes=1000 from Content-Length, got %d", lastProgress.TotalBytes)
	}
	if lastProgress.BytesRead != 5 {
		t.Errorf("expected BytesRead=5 (actual body size), got %d", lastProgress.BytesRead)
	}

	// The file should exist but be shorter than advertised
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file to exist at %s: %v", path, err)
	}
	if info.Size() != 5 {
		t.Errorf("expected file size 5, got %d", info.Size())
	}
}
