package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// RollbackManager manages model version history for rollback support.
type RollbackManager struct {
	storageDir  string
	maxVersions int
}

func NewRollbackManager(storageDir string, maxVersions int) *RollbackManager {
	return &RollbackManager{
		storageDir:  storageDir,
		maxVersions: maxVersions,
	}
}

// SaveVersion stores a model version for potential rollback.
// sourcePath is the file path of the current model to back up.
func (r *RollbackManager) SaveVersion(modelName, version string, sourcePath string) error {
	dir := filepath.Join(r.storageDir, "rollback", modelName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create rollback dir: %w", err)
	}

	// Copy model file to rollback directory
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read source model: %w", err)
	}

	path := filepath.Join(dir, version)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write rollback file: %w", err)
	}

	// Evict oldest if we exceed maxVersions
	return r.evictOldVersions(dir)
}

// HasVersion checks if a version is available for rollback.
func (r *RollbackManager) HasVersion(modelName, version string) bool {
	path := filepath.Join(r.storageDir, "rollback", modelName, version)
	_, err := os.Stat(path)
	return err == nil
}

// Restore returns the file path of a previously saved model version.
func (r *RollbackManager) Restore(modelName, version string) (string, error) {
	path := filepath.Join(r.storageDir, "rollback", modelName, version)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("rollback version not found: %w", err)
	}
	return path, nil
}

// ListVersions lists available rollback versions for a model.
func (r *RollbackManager) ListVersions(modelName string) ([]string, error) {
	dir := filepath.Join(r.storageDir, "rollback", modelName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var versions []string
	for _, e := range entries {
		if !e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	return versions, nil
}

func (r *RollbackManager) evictOldVersions(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	if len(entries) <= r.maxVersions {
		return nil
	}

	// Sort by modification time (oldest first)
	type fileInfo struct {
		name    string
		modTime int64
	}
	var files []fileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{name: e.Name(), modTime: info.ModTime().UnixNano()})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime < files[j].modTime
	})

	// Remove oldest files until we're at maxVersions
	for len(files) > r.maxVersions {
		path := filepath.Join(dir, files[0].name)
		os.Remove(path)
		files = files[1:]
	}

	return nil
}
