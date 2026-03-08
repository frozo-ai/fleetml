package logging

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/fleetml/fleetml/agent/internal/communication"
	"github.com/fleetml/fleetml/agent/internal/device"
	"github.com/fleetml/fleetml/agent/internal/health"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type mockComm struct {
	mu      sync.Mutex
	batches [][]communication.LogEntry
}

func (m *mockComm) Register(ctx context.Context, info *device.Info) (string, int, error) {
	return "", 0, nil
}
func (m *mockComm) SendHeartbeat(ctx context.Context, deviceID, status string, system *health.SystemMetrics) ([]communication.Command, error) {
	return nil, nil
}
func (m *mockComm) ReportDeploymentStatus(ctx context.Context, deviceID, deploymentID, state, errMsg string) error {
	return nil
}
func (m *mockComm) SendLogs(ctx context.Context, deviceID string, entries []communication.LogEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.batches = append(m.batches, entries)
	return nil
}
func (m *mockComm) Close() error { return nil }
func (m *mockComm) totalEntries() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	total := 0
	for _, b := range m.batches {
		total += len(b)
	}
	return total
}

func TestForwarder_BuffersAndFlushes(t *testing.T) {
	comm := &mockComm{}
	logger := zap.NewNop().Sugar()
	f := NewForwarder(comm, "device-1", logger)
	f.Start(50 * time.Millisecond)

	// Write some log entries
	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Now(),
		Message: "test log message",
	}
	f.Write(entry, nil)
	f.Write(entry, nil)
	f.Write(entry, nil)

	// Wait for flush
	time.Sleep(200 * time.Millisecond)
	f.Stop()

	if comm.totalEntries() != 3 {
		t.Errorf("expected 3 forwarded entries, got %d", comm.totalEntries())
	}
}

func TestForwarder_Enabled(t *testing.T) {
	comm := &mockComm{}
	f := NewForwarder(comm, "device-1", nil)

	if f.Enabled(zapcore.DebugLevel) {
		t.Error("debug should not be enabled")
	}
	if !f.Enabled(zapcore.InfoLevel) {
		t.Error("info should be enabled")
	}
	if !f.Enabled(zapcore.WarnLevel) {
		t.Error("warn should be enabled")
	}
	if !f.Enabled(zapcore.ErrorLevel) {
		t.Error("error should be enabled")
	}
}

func TestForwarder_With(t *testing.T) {
	comm := &mockComm{}
	f := NewForwarder(comm, "device-1", nil)

	core := f.With([]zapcore.Field{zap.String("component", "deploy")})
	if core == nil {
		t.Fatal("expected non-nil core from With")
	}
}

func TestForwarder_Sync(t *testing.T) {
	comm := &mockComm{}
	logger := zap.NewNop().Sugar()
	f := NewForwarder(comm, "device-1", logger)

	entry := zapcore.Entry{
		Level:   zapcore.WarnLevel,
		Time:    time.Now(),
		Message: "sync test",
	}
	f.Write(entry, nil)

	// Sync should flush immediately
	f.Sync()

	if comm.totalEntries() != 1 {
		t.Errorf("expected 1 entry after sync, got %d", comm.totalEntries())
	}
}

func TestForwarder_EmptyFlush(t *testing.T) {
	comm := &mockComm{}
	f := NewForwarder(comm, "device-1", nil)

	// Flush with nothing buffered should be a no-op
	f.flush()

	if len(comm.batches) != 0 {
		t.Error("expected no batches for empty flush")
	}
}
