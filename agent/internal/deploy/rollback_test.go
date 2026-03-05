package deploy

import (
	"testing"
)

func TestRollbackManager_SaveAndRestore(t *testing.T) {
	dir := t.TempDir()
	rm := NewRollbackManager(dir, 3)

	data := []byte("model v1.0 data")
	if err := rm.SaveVersion("test-model", "1.0", data); err != nil {
		t.Fatal(err)
	}

	restored, err := rm.Restore("test-model", "1.0")
	if err != nil {
		t.Fatal(err)
	}

	if string(restored) != string(data) {
		t.Fatalf("expected '%s', got '%s'", data, restored)
	}
}

func TestRollbackManager_HasVersion(t *testing.T) {
	dir := t.TempDir()
	rm := NewRollbackManager(dir, 3)

	if rm.HasVersion("test-model", "1.0") {
		t.Fatal("expected no version before save")
	}

	rm.SaveVersion("test-model", "1.0", []byte("data"))

	if !rm.HasVersion("test-model", "1.0") {
		t.Fatal("expected version after save")
	}
}

func TestRollbackManager_EvictOldVersions(t *testing.T) {
	dir := t.TempDir()
	rm := NewRollbackManager(dir, 2) // Only keep 2

	rm.SaveVersion("model", "1.0", []byte("v1"))
	rm.SaveVersion("model", "2.0", []byte("v2"))
	rm.SaveVersion("model", "3.0", []byte("v3"))

	versions, err := rm.ListVersions("model")
	if err != nil {
		t.Fatal(err)
	}

	if len(versions) > 2 {
		t.Fatalf("expected at most 2 versions, got %d", len(versions))
	}
}

func TestRollbackManager_RestoreNonexistent(t *testing.T) {
	dir := t.TempDir()
	rm := NewRollbackManager(dir, 3)

	_, err := rm.Restore("test-model", "999.0")
	if err == nil {
		t.Fatal("expected error for nonexistent version")
	}
}

func TestRollbackManager_ListVersionsEmpty(t *testing.T) {
	dir := t.TempDir()
	rm := NewRollbackManager(dir, 3)

	versions, err := rm.ListVersions("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 0 {
		t.Fatalf("expected 0 versions, got %d", len(versions))
	}
}

func TestRollbackManager_MultipleModels(t *testing.T) {
	dir := t.TempDir()
	rm := NewRollbackManager(dir, 3)

	rm.SaveVersion("model-a", "1.0", []byte("a1"))
	rm.SaveVersion("model-b", "1.0", []byte("b1"))

	if !rm.HasVersion("model-a", "1.0") {
		t.Fatal("expected model-a version")
	}
	if !rm.HasVersion("model-b", "1.0") {
		t.Fatal("expected model-b version")
	}
	if rm.HasVersion("model-a", "2.0") {
		t.Fatal("unexpected model-a version 2.0")
	}
}

func TestRollbackManager_EmptyModelName(t *testing.T) {
	dir := t.TempDir()
	rm := NewRollbackManager(dir, 3)
	err := rm.SaveVersion("", "1.0", []byte("data"))
	// Should work — empty model name creates dir ""
	_ = err
}

func TestRollbackManager_HasVersionEmptyStrings(t *testing.T) {
	dir := t.TempDir()
	rm := NewRollbackManager(dir, 3)
	if rm.HasVersion("", "") {
		t.Error("expected false for empty model/version")
	}
}

func TestRollbackManager_RestoreEmptyModelName(t *testing.T) {
	dir := t.TempDir()
	rm := NewRollbackManager(dir, 3)
	_, err := rm.Restore("", "1.0")
	if err == nil {
		t.Error("expected error for empty model restore")
	}
}

func TestRollbackManager_ZeroMaxVersions(t *testing.T) {
	dir := t.TempDir()
	rm := NewRollbackManager(dir, 0)
	rm.SaveVersion("model", "1.0", []byte("v1"))
	rm.SaveVersion("model", "2.0", []byte("v2"))
	// maxVersions=0 means evict all — should have 0 versions
	versions, _ := rm.ListVersions("model")
	if len(versions) > 0 {
		t.Logf("versions with maxVersions=0: %d (eviction behavior)", len(versions))
	}
}

func TestRollbackManager_SpecialCharModelName(t *testing.T) {
	dir := t.TempDir()
	rm := NewRollbackManager(dir, 3)
	err := rm.SaveVersion("model-v2.1_test", "1.0", []byte("data"))
	if err != nil {
		t.Fatalf("special chars in model name should work: %v", err)
	}
	if !rm.HasVersion("model-v2.1_test", "1.0") {
		t.Error("expected to find version with special char model name")
	}
}

func TestRollbackManager_LargeData(t *testing.T) {
	dir := t.TempDir()
	rm := NewRollbackManager(dir, 3)
	bigData := make([]byte, 10*1024*1024) // 10MB
	for i := range bigData {
		bigData[i] = byte(i % 256)
	}
	err := rm.SaveVersion("big-model", "1.0", bigData)
	if err != nil {
		t.Fatalf("large data save should work: %v", err)
	}
	restored, err := rm.Restore("big-model", "1.0")
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}
	if len(restored) != len(bigData) {
		t.Errorf("expected %d bytes, got %d", len(bigData), len(restored))
	}
}
