package model

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadFromURL_Success(t *testing.T) {
	modelData := "fake-model-binary-data-for-testing"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(modelData)))
		w.Write([]byte(modelData))
	}))
	defer server.Close()

	dir := t.TempDir()
	loader := NewLoader(dir, 5)

	var lastProgress DownloadProgress
	path, err := loader.DownloadFromURL(context.Background(), "test-model", "v1.0", server.URL,
		func(p DownloadProgress) {
			lastProgress = p
		})
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	// Check file was saved
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(data) != modelData {
		t.Errorf("expected %q, got %q", modelData, string(data))
	}

	// Check progress was reported
	if lastProgress.BytesRead != int64(len(modelData)) {
		t.Errorf("expected %d bytes read, got %d", len(modelData), lastProgress.BytesRead)
	}
	if lastProgress.TotalBytes != int64(len(modelData)) {
		t.Errorf("expected %d total bytes, got %d", len(modelData), lastProgress.TotalBytes)
	}
}

func TestDownloadFromURL_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	dir := t.TempDir()
	loader := NewLoader(dir, 5)

	_, err := loader.DownloadFromURL(context.Background(), "test-model", "v1.0", server.URL, nil)
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestDownloadFromURL_Cancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow response
		w.Header().Set("Content-Length", "1000000")
		w.Write([]byte("partial"))
		w.(http.Flusher).Flush()
		// Block until test is done
		<-r.Context().Done()
	}))
	defer server.Close()

	dir := t.TempDir()
	loader := NewLoader(dir, 5)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := loader.DownloadFromURL(ctx, "test-model", "v1.0", server.URL, nil)
	if err == nil {
		t.Fatal("expected error for cancelled download")
	}
}

func TestDownloadFromURL_AtomicRename(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("model-data"))
	}))
	defer server.Close()

	dir := t.TempDir()
	loader := NewLoader(dir, 5)

	path, err := loader.DownloadFromURL(context.Background(), "my-model", "v2.0", server.URL, nil)
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	// No .downloading temp file should remain
	entries, _ := filepath.Glob(filepath.Join(dir, "my-model", "*.downloading"))
	if len(entries) > 0 {
		t.Errorf("temp .downloading file not cleaned up: %v", entries)
	}

	// Final path should exist
	if _, err := os.Stat(path); err != nil {
		t.Errorf("final model file not found: %v", err)
	}
}

func TestCleanupPartialDownloads(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader(dir, 5)

	// Create some partial download files
	modelDir := filepath.Join(dir, "test-model")
	os.MkdirAll(modelDir, 0o755)
	os.WriteFile(filepath.Join(modelDir, "v1.0.onnx"), []byte("complete"), 0o644)
	os.WriteFile(filepath.Join(modelDir, "v2.0.onnx.downloading"), []byte("partial"), 0o644)

	if err := loader.CleanupPartialDownloads(); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	// Complete file should still exist
	if _, err := os.Stat(filepath.Join(modelDir, "v1.0.onnx")); err != nil {
		t.Error("complete file was deleted")
	}

	// Partial file should be gone
	if _, err := os.Stat(filepath.Join(modelDir, "v2.0.onnx.downloading")); !os.IsNotExist(err) {
		t.Error("partial download file was not cleaned up")
	}
}

func TestDownloadFromURL_InvalidURL(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader(dir, 5)

	_, err := loader.DownloadFromURL(context.Background(), "m", "v1", "http://invalid.localhost:99999/model", nil)
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestDownloadFromURL_ProgressCallback(t *testing.T) {
	// Create a chunked response
	bigData := strings.Repeat("x", 1024*10) // 10KB
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(bigData)))
		w.Write([]byte(bigData))
	}))
	defer server.Close()

	dir := t.TempDir()
	loader := NewLoader(dir, 5)

	var progressCalls int
	_, err := loader.DownloadFromURL(context.Background(), "m", "v1", server.URL,
		func(p DownloadProgress) {
			progressCalls++
			if p.ModelName != "m" {
				t.Errorf("expected model name 'm', got %q", p.ModelName)
			}
			if p.ModelVersion != "v1" {
				t.Errorf("expected model version 'v1', got %q", p.ModelVersion)
			}
		})
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	if progressCalls == 0 {
		t.Error("expected at least one progress callback")
	}
}
