package model

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DownloadProgress reports on a model download.
type DownloadProgress struct {
	ModelName    string
	ModelVersion string
	BytesRead    int64
	TotalBytes   int64 // -1 if unknown
	StartedAt    time.Time
}

// ProgressCallback is called during model download with progress updates.
type ProgressCallback func(DownloadProgress)

// Loader handles downloading and validating model files.
type Loader struct {
	storageDir  string
	maxVersions int
	httpClient  *http.Client
}

func NewLoader(storageDir string, maxVersions int) *Loader {
	return &Loader{
		storageDir:  storageDir,
		maxVersions: maxVersions,
		httpClient: &http.Client{
			Timeout: 30 * time.Minute, // Large model downloads
		},
	}
}

// Load validates and returns a model file path.
func (l *Loader) Load(name, version string) (string, error) {
	modelPath := l.modelPath(name, version)
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return "", fmt.Errorf("model %s:%s not found at %s", name, version, modelPath)
	}
	return modelPath, nil
}

// DownloadFromURL downloads a model from a presigned S3 URL with progress
// reporting and streaming to disk. Supports cancellation via context.
func (l *Loader) DownloadFromURL(ctx context.Context, name, version, url string, progressCb ProgressCallback) (string, error) {
	dir := filepath.Join(l.storageDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create model dir: %w", err)
	}

	// Use a temp file then rename for atomicity
	tmpPath := l.modelPath(name, version) + ".downloading"
	finalPath := l.modelPath(name, version)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	f, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}

	progress := DownloadProgress{
		ModelName:    name,
		ModelVersion: version,
		TotalBytes:   resp.ContentLength,
		StartedAt:    time.Now(),
	}

	// Wrap reader with progress tracking
	reader := &progressReader{
		reader: resp.Body,
		progress: &progress,
		callback: progressCb,
	}

	_, copyErr := io.Copy(f, reader)
	f.Close()

	if copyErr != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("write model: %w", copyErr)
	}

	// Atomically rename temp file to final path
	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("rename to final path: %w", err)
	}

	return finalPath, nil
}

// progressReader wraps an io.Reader to report download progress.
type progressReader struct {
	reader   io.Reader
	progress *DownloadProgress
	callback ProgressCallback
	mu       sync.Mutex
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.mu.Lock()
		pr.progress.BytesRead += int64(n)
		if pr.callback != nil {
			pr.callback(*pr.progress)
		}
		pr.mu.Unlock()
	}
	return n, err
}

// ValidateChecksum verifies the SHA-256 checksum of a file.
func (l *Loader) ValidateChecksum(filePath string, expectedChecksum string) error {
	// Strip "sha256:" prefix if present
	expected := strings.TrimPrefix(expectedChecksum, "sha256:")

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("compute checksum: %w", err)
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if actual != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}

	return nil
}

// ComputeChecksum computes the SHA-256 checksum of a file.
func ComputeChecksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("compute checksum: %w", err)
	}

	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// SaveModel saves model data to the storage directory.
func (l *Loader) SaveModel(name, version string, data io.Reader) (string, error) {
	dir := filepath.Join(l.storageDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create model dir: %w", err)
	}

	modelPath := l.modelPath(name, version)
	f, err := os.Create(modelPath)
	if err != nil {
		return "", fmt.Errorf("create model file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, data); err != nil {
		os.Remove(modelPath)
		return "", fmt.Errorf("write model file: %w", err)
	}

	return modelPath, nil
}

func (l *Loader) modelPath(name, version string) string {
	return filepath.Join(l.storageDir, name, version+".onnx")
}

// StorageDir returns the storage directory.
func (l *Loader) StorageDir() string {
	return l.storageDir
}

// CleanupPartialDownloads removes any .downloading temp files left from
// interrupted downloads.
func (l *Loader) CleanupPartialDownloads() error {
	return filepath.Walk(l.storageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if !info.IsDir() && strings.HasSuffix(path, ".downloading") {
			os.Remove(path)
		}
		return nil
	})
}
