package communication

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/fleetml/fleetml/agent/internal/device"
	"github.com/fleetml/fleetml/agent/internal/health"
	"github.com/fleetml/fleetml/agent/internal/offline"
	"go.uber.org/zap"
)

// fakeCommunicator simulates a primary communicator that can fail.
type fakeCommunicator struct {
	mu         sync.Mutex
	shouldFail bool
	heartbeats int
}

func (f *fakeCommunicator) Register(ctx context.Context, info *device.Info) (string, int, error) {
	return "agent-1", 30, nil
}

func (f *fakeCommunicator) SendHeartbeat(ctx context.Context, deviceID, status string, system *health.SystemMetrics) ([]Command, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.shouldFail {
		return nil, fmt.Errorf("connection refused")
	}
	f.heartbeats++
	return nil, nil
}

func (f *fakeCommunicator) ReportDeploymentStatus(ctx context.Context, deviceID, deploymentID, state, errMsg string) error {
	return nil
}

func (f *fakeCommunicator) Close() error { return nil }

func (f *fakeCommunicator) setFail(fail bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.shouldFail = fail
}

func (f *fakeCommunicator) getHeartbeats() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.heartbeats
}

// fakeStore implements MetricsStore for testing.
type fakeStore struct {
	mu      sync.Mutex
	records []*offline.HeartbeatRecord
}

func (s *fakeStore) SaveHeartbeat(record *offline.HeartbeatRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, record)
	return nil
}

func (s *fakeStore) GetBufferedHeartbeats() ([]*offline.HeartbeatRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]*offline.HeartbeatRecord, len(s.records))
	copy(result, s.records)
	return result, nil
}

func (s *fakeStore) ClearBuffer() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = nil
	return nil
}

func (s *fakeStore) Close() error { return nil }

func (s *fakeStore) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.records)
}

func TestStoreForward_SuccessPassthrough(t *testing.T) {
	primary := &fakeCommunicator{}
	store := &fakeStore{}
	logger, _ := zap.NewDevelopment()

	sf := NewStoreForwardManager(primary, store, logger.Sugar())

	_, err := sf.SendHeartbeat(context.Background(), "dev-1", "healthy", nil)
	if err != nil {
		t.Fatalf("heartbeat should succeed: %v", err)
	}

	if primary.getHeartbeats() != 1 {
		t.Errorf("expected 1 heartbeat on primary, got %d", primary.getHeartbeats())
	}
	if store.count() != 0 {
		t.Errorf("expected 0 buffered, got %d", store.count())
	}
}

func TestStoreForward_FailureBuffers(t *testing.T) {
	primary := &fakeCommunicator{shouldFail: true}
	store := &fakeStore{}
	logger, _ := zap.NewDevelopment()

	sf := NewStoreForwardManager(primary, store, logger.Sugar())

	// Should not return an error (swallowed)
	_, err := sf.SendHeartbeat(context.Background(), "dev-1", "healthy", &health.SystemMetrics{CPUPercent: 50})
	if err != nil {
		t.Fatalf("heartbeat should not return error: %v", err)
	}

	if store.count() != 1 {
		t.Errorf("expected 1 buffered heartbeat, got %d", store.count())
	}
	if !sf.IsConnected() == true {
		// After failure, should be disconnected
	}
}

func TestStoreForward_MultipleFailuresBuffer(t *testing.T) {
	primary := &fakeCommunicator{shouldFail: true}
	store := &fakeStore{}
	logger, _ := zap.NewDevelopment()

	sf := NewStoreForwardManager(primary, store, logger.Sugar())

	for i := 0; i < 5; i++ {
		sf.SendHeartbeat(context.Background(), "dev-1", "healthy", nil)
	}

	if store.count() != 5 {
		t.Errorf("expected 5 buffered heartbeats, got %d", store.count())
	}
}

func TestStoreForward_ReconnectionTriggersBulkSync(t *testing.T) {
	primary := &fakeCommunicator{shouldFail: true}
	store := &fakeStore{}
	logger, _ := zap.NewDevelopment()

	sf := NewStoreForwardManager(primary, store, logger.Sugar())

	// Buffer some heartbeats
	for i := 0; i < 3; i++ {
		sf.SendHeartbeat(context.Background(), "dev-1", "healthy", nil)
	}

	if store.count() != 3 {
		t.Fatalf("expected 3 buffered, got %d", store.count())
	}

	// Restore connection
	primary.setFail(false)

	// Next heartbeat should trigger bulk sync
	sf.SendHeartbeat(context.Background(), "dev-1", "healthy", nil)

	// Give bulk sync goroutine time to run
	time.Sleep(100 * time.Millisecond)

	// Buffer should be cleared after sync
	if store.count() != 0 {
		t.Errorf("expected buffer cleared after sync, got %d", store.count())
	}

	// Primary should have received: 1 live + 3 synced = 4
	if primary.getHeartbeats() != 4 {
		t.Errorf("expected 4 heartbeats on primary, got %d", primary.getHeartbeats())
	}
}

func TestStoreForward_IsConnected(t *testing.T) {
	primary := &fakeCommunicator{}
	store := &fakeStore{}
	logger, _ := zap.NewDevelopment()

	sf := NewStoreForwardManager(primary, store, logger.Sugar())

	if !sf.IsConnected() {
		t.Error("should be connected initially")
	}

	primary.setFail(true)
	sf.SendHeartbeat(context.Background(), "dev-1", "healthy", nil)

	if sf.IsConnected() {
		t.Error("should be disconnected after failure")
	}

	primary.setFail(false)
	sf.SendHeartbeat(context.Background(), "dev-1", "healthy", nil)

	if !sf.IsConnected() {
		t.Error("should be connected after success")
	}
}
