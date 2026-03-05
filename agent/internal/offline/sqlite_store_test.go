package offline

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/fleetml/fleetml/agent/internal/health"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestSQLiteStore_SaveAndGet(t *testing.T) {
	store := newTestStore(t)

	record := &HeartbeatRecord{
		DeviceID:  "dev-1",
		Status:    "healthy",
		System:    &health.SystemMetrics{CPUPercent: 45.5, RAMMBUsed: 2048},
		Timestamp: time.Now().Unix(),
	}

	if err := store.SaveHeartbeat(record); err != nil {
		t.Fatalf("save heartbeat: %v", err)
	}

	records, err := store.GetBufferedHeartbeats()
	if err != nil {
		t.Fatalf("get heartbeats: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	got := records[0]
	if got.DeviceID != "dev-1" {
		t.Errorf("expected device_id 'dev-1', got %q", got.DeviceID)
	}
	if got.Status != "healthy" {
		t.Errorf("expected status 'healthy', got %q", got.Status)
	}
	if got.System == nil {
		t.Fatal("expected non-nil system metrics")
	}
	if got.System.CPUPercent != 45.5 {
		t.Errorf("expected CPU 45.5, got %f", got.System.CPUPercent)
	}
}

func TestSQLiteStore_MultipleRecords(t *testing.T) {
	store := newTestStore(t)

	now := time.Now().Unix()
	for i := 0; i < 10; i++ {
		record := &HeartbeatRecord{
			DeviceID:  "dev-1",
			Status:    "healthy",
			Timestamp: now + int64(i),
		}
		if err := store.SaveHeartbeat(record); err != nil {
			t.Fatalf("save heartbeat %d: %v", i, err)
		}
	}

	records, err := store.GetBufferedHeartbeats()
	if err != nil {
		t.Fatalf("get heartbeats: %v", err)
	}

	if len(records) != 10 {
		t.Fatalf("expected 10 records, got %d", len(records))
	}

	// Should be ordered by timestamp ASC
	for i := 1; i < len(records); i++ {
		if records[i].Timestamp < records[i-1].Timestamp {
			t.Error("records not ordered by timestamp")
		}
	}
}

func TestSQLiteStore_ClearBuffer(t *testing.T) {
	store := newTestStore(t)

	for i := 0; i < 5; i++ {
		store.SaveHeartbeat(&HeartbeatRecord{
			DeviceID:  "dev-1",
			Status:    "healthy",
			Timestamp: time.Now().Unix(),
		})
	}

	if err := store.ClearBuffer(); err != nil {
		t.Fatalf("clear buffer: %v", err)
	}

	records, _ := store.GetBufferedHeartbeats()
	if len(records) != 0 {
		t.Errorf("expected 0 records after clear, got %d", len(records))
	}
}

func TestSQLiteStore_BufferCount(t *testing.T) {
	store := newTestStore(t)

	for i := 0; i < 3; i++ {
		store.SaveHeartbeat(&HeartbeatRecord{
			DeviceID:  "dev-1",
			Status:    "healthy",
			Timestamp: time.Now().Unix(),
		})
	}

	count, err := store.BufferCount()
	if err != nil {
		t.Fatalf("buffer count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

func TestSQLiteStore_NilSystemMetrics(t *testing.T) {
	store := newTestStore(t)

	record := &HeartbeatRecord{
		DeviceID:  "dev-1",
		Status:    "healthy",
		System:    nil,
		Timestamp: time.Now().Unix(),
	}

	if err := store.SaveHeartbeat(record); err != nil {
		t.Fatalf("save heartbeat: %v", err)
	}

	records, _ := store.GetBufferedHeartbeats()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	// System should be nil since we stored nil
	if records[0].System != nil {
		t.Error("expected nil system metrics")
	}
}

func TestSQLiteStore_CommandQueue(t *testing.T) {
	store := newTestStore(t)

	// Save some commands
	if err := store.SaveCommand("cmd-1", "deploy_model", []byte(`{"model":"test"}`)); err != nil {
		t.Fatalf("save command: %v", err)
	}
	if err := store.SaveCommand("cmd-2", "rollback", []byte(`{}`)); err != nil {
		t.Fatalf("save command: %v", err)
	}

	// Get pending commands
	cmds, err := store.GetPendingCommands()
	if err != nil {
		t.Fatalf("get pending commands: %v", err)
	}
	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}

	// Mark first as processed
	if err := store.MarkCommandProcessed("cmd-1"); err != nil {
		t.Fatalf("mark processed: %v", err)
	}

	cmds, _ = store.GetPendingCommands()
	if len(cmds) != 1 {
		t.Fatalf("expected 1 pending command, got %d", len(cmds))
	}
	if cmds[0].ID != "cmd-2" {
		t.Errorf("expected cmd-2, got %s", cmds[0].ID)
	}
}

func TestSQLiteStore_DuplicateCommand(t *testing.T) {
	store := newTestStore(t)

	store.SaveCommand("cmd-1", "deploy", []byte("{}"))
	store.SaveCommand("cmd-1", "deploy", []byte("{}")) // Duplicate, should be ignored

	cmds, _ := store.GetPendingCommands()
	if len(cmds) != 1 {
		t.Errorf("expected 1 command (duplicate ignored), got %d", len(cmds))
	}
}
