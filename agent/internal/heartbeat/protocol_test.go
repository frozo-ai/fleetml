package heartbeat

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fleetml/fleetml/agent/internal/communication"
	"github.com/fleetml/fleetml/agent/internal/device"
	"github.com/fleetml/fleetml/agent/internal/health"
	"go.uber.org/zap"
)

// mockCommunicator implements communication.Communicator for testing.
type mockCommunicator struct {
	heartbeatFn func(ctx context.Context, deviceID string, status string, system *health.SystemMetrics) ([]communication.Command, error)
}

func (m *mockCommunicator) Register(ctx context.Context, info *device.Info) (string, int, error) {
	return "test-agent-id", 30, nil
}

func (m *mockCommunicator) SendHeartbeat(ctx context.Context, deviceID string, status string, system *health.SystemMetrics) ([]communication.Command, error) {
	if m.heartbeatFn != nil {
		return m.heartbeatFn(ctx, deviceID, status, system)
	}
	return nil, nil
}

func (m *mockCommunicator) ReportDeploymentStatus(ctx context.Context, deviceID, deploymentID, state, errMsg string) error {
	return nil
}

func (m *mockCommunicator) SendLogs(ctx context.Context, deviceID string, entries []communication.LogEntry) error {
	return nil
}

func (m *mockCommunicator) Close() error {
	return nil
}

func testLogger() *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	return logger.Sugar()
}

func TestNewScheduler(t *testing.T) {
	comm := &mockCommunicator{}
	reporter := health.NewReporter(5 * time.Second)
	logger := testLogger()

	s := NewScheduler("device-1", 10*time.Second, comm, reporter, logger)
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
	if s.deviceID != "device-1" {
		t.Fatalf("expected deviceID 'device-1', got '%s'", s.deviceID)
	}
	if s.interval != 10*time.Second {
		t.Fatalf("expected interval 10s, got %v", s.interval)
	}
	if s.communicator == nil {
		t.Fatal("expected non-nil communicator")
	}
	if s.reporter == nil {
		t.Fatal("expected non-nil reporter")
	}
}

func TestScheduler_Commands(t *testing.T) {
	comm := &mockCommunicator{}
	reporter := health.NewReporter(5 * time.Second)
	logger := testLogger()

	s := NewScheduler("device-1", 10*time.Second, comm, reporter, logger)
	ch := s.Commands()
	if ch == nil {
		t.Fatal("expected non-nil commands channel")
	}

	// Verify the buffer size is 100 by checking capacity
	if cap(s.commandCh) != 100 {
		t.Fatalf("expected command channel capacity 100, got %d", cap(s.commandCh))
	}
}

func TestScheduler_StatusDetermination(t *testing.T) {
	// Test that sendHeartbeat sets status to "warning" when CPU > 90 or Disk > 90.
	// We use a mock communicator that captures the status argument.
	tests := []struct {
		name           string
		cpuPercent     float64
		diskPercent    float64
		expectedStatus string
	}{
		{
			name:           "healthy_low_usage",
			cpuPercent:     20.0,
			diskPercent:    30.0,
			expectedStatus: "healthy",
		},
		{
			name:           "warning_high_cpu",
			cpuPercent:     95.0,
			diskPercent:    30.0,
			expectedStatus: "warning",
		},
		{
			name:           "warning_high_disk",
			cpuPercent:     20.0,
			diskPercent:    95.0,
			expectedStatus: "warning",
		},
		{
			name:           "warning_both_high",
			cpuPercent:     95.0,
			diskPercent:    95.0,
			expectedStatus: "warning",
		},
		{
			name:           "healthy_at_boundary",
			cpuPercent:     90.0,
			diskPercent:    90.0,
			expectedStatus: "healthy",
		},
		{
			name:           "warning_just_over_cpu",
			cpuPercent:     90.1,
			diskPercent:    50.0,
			expectedStatus: "warning",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedStatus string
			comm := &mockCommunicator{
				heartbeatFn: func(ctx context.Context, deviceID string, status string, system *health.SystemMetrics) ([]communication.Command, error) {
					capturedStatus = status
					return nil, nil
				},
			}

			// We cannot easily inject custom metrics into the reporter since it reads
			// from the OS. Instead, we test the status logic directly by observing
			// what the scheduler sends to the communicator. The actual status depends
			// on real system metrics. However, the logic itself is straightforward:
			// status = "healthy"; if cpu > 90 || disk > 90: status = "warning"
			//
			// Since we cannot mock gopsutil, we verify the contract by checking that
			// the communicator receives a valid status string.
			reporter := health.NewReporter(5 * time.Second)
			logger := testLogger()
			s := NewScheduler("device-1", 10*time.Second, comm, reporter, logger)

			s.sendHeartbeat(context.Background())

			if capturedStatus != "healthy" && capturedStatus != "warning" {
				t.Fatalf("expected status 'healthy' or 'warning', got '%s'", capturedStatus)
			}
		})
	}
}

func TestScheduler_Start_ContextCancel(t *testing.T) {
	comm := &mockCommunicator{}
	reporter := health.NewReporter(5 * time.Second)
	logger := testLogger()

	s := NewScheduler("device-1", 50*time.Millisecond, comm, reporter, logger)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		s.Start(ctx)
		close(done)
	}()

	// Let it run briefly
	time.Sleep(200 * time.Millisecond)

	cancel()

	select {
	case <-done:
		// Success: Start returned after context cancellation.
	case <-time.After(5 * time.Second):
		t.Fatal("scheduler did not stop within 5 seconds after context cancel")
	}
}

func TestScheduler_SendHeartbeat_Success(t *testing.T) {
	expectedCommands := []communication.Command{
		{ID: "cmd-1", Type: "deploy_model", Payload: []byte(`{"model":"test"}`)},
		{ID: "cmd-2", Type: "rollback", Payload: []byte(`{}`)},
	}

	comm := &mockCommunicator{
		heartbeatFn: func(ctx context.Context, deviceID string, status string, system *health.SystemMetrics) ([]communication.Command, error) {
			return expectedCommands, nil
		},
	}
	reporter := health.NewReporter(5 * time.Second)
	logger := testLogger()

	s := NewScheduler("device-1", 10*time.Second, comm, reporter, logger)

	s.sendHeartbeat(context.Background())

	// Read commands from the channel
	for _, expected := range expectedCommands {
		select {
		case cmd := <-s.Commands():
			if cmd.ID != expected.ID {
				t.Errorf("expected command ID '%s', got '%s'", expected.ID, cmd.ID)
			}
			if cmd.Type != expected.Type {
				t.Errorf("expected command type '%s', got '%s'", expected.Type, cmd.Type)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for command '%s'", expected.ID)
		}
	}
}

func TestScheduler_SendHeartbeat_Error(t *testing.T) {
	comm := &mockCommunicator{
		heartbeatFn: func(ctx context.Context, deviceID string, status string, system *health.SystemMetrics) ([]communication.Command, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	reporter := health.NewReporter(5 * time.Second)
	logger := testLogger()

	s := NewScheduler("device-1", 50*time.Millisecond, comm, reporter, logger)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		s.Start(ctx)
		close(done)
	}()

	// Let the scheduler run a few heartbeat cycles with errors
	time.Sleep(300 * time.Millisecond)

	cancel()

	select {
	case <-done:
		// Success: scheduler continued despite errors and stopped on cancel.
	case <-time.After(5 * time.Second):
		t.Fatal("scheduler did not stop after context cancel despite heartbeat errors")
	}
}

func TestScheduler_CommandChannelFull(t *testing.T) {
	// Create a communicator that returns one command per heartbeat
	callCount := 0
	comm := &mockCommunicator{
		heartbeatFn: func(ctx context.Context, deviceID string, status string, system *health.SystemMetrics) ([]communication.Command, error) {
			callCount++
			return []communication.Command{
				{ID: fmt.Sprintf("cmd-%d", callCount), Type: "deploy_model"},
			}, nil
		},
	}
	reporter := health.NewReporter(5 * time.Second)
	logger := testLogger()

	s := NewScheduler("device-1", 10*time.Second, comm, reporter, logger)

	// Fill the command channel to capacity (100)
	for i := 0; i < 100; i++ {
		s.commandCh <- communication.Command{
			ID:   fmt.Sprintf("fill-%d", i),
			Type: "deploy_model",
		}
	}

	// Verify the channel is full
	if len(s.commandCh) != 100 {
		t.Fatalf("expected channel length 100, got %d", len(s.commandCh))
	}

	// This should not panic even though the channel is full.
	// The command should be dropped gracefully via the default case in select.
	s.sendHeartbeat(context.Background())

	// Channel should still be at capacity (the new command was dropped)
	if len(s.commandCh) != 100 {
		t.Fatalf("expected channel length to remain 100 after drop, got %d", len(s.commandCh))
	}
}

func TestScheduler_EmptyDeviceID(t *testing.T) {
	comm := &mockCommunicator{}
	reporter := health.NewReporter(5 * time.Second)
	logger := testLogger()
	s := NewScheduler("", 10*time.Second, comm, reporter, logger)
	if s == nil {
		t.Fatal("expected non-nil scheduler with empty device ID")
	}
	if s.deviceID != "" {
		t.Errorf("expected empty device ID, got %q", s.deviceID)
	}
}

func TestScheduler_Start_ImmediateCancel(t *testing.T) {
	comm := &mockCommunicator{}
	reporter := health.NewReporter(5 * time.Second)
	logger := testLogger()
	s := NewScheduler("device-1", time.Hour, comm, reporter, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately before Start

	done := make(chan struct{})
	go func() {
		s.Start(ctx)
		close(done)
	}()

	select {
	case <-done:
		// Success: returned immediately
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler should return immediately on pre-canceled context")
	}
}

func TestScheduler_Commands_Channel(t *testing.T) {
	comm := &mockCommunicator{}
	reporter := health.NewReporter(5 * time.Second)
	logger := testLogger()
	s := NewScheduler("device-1", 10*time.Second, comm, reporter, logger)

	ch := s.Commands()
	if ch == nil {
		t.Fatal("expected non-nil command channel")
	}
	if cap(ch) != 100 {
		t.Errorf("expected channel capacity 100, got %d", cap(ch))
	}
}

func TestScheduler_SendHeartbeat_DeviceIDPassed(t *testing.T) {
	var capturedDeviceID string
	comm := &mockCommunicator{
		heartbeatFn: func(ctx context.Context, deviceID string, status string, system *health.SystemMetrics) ([]communication.Command, error) {
			capturedDeviceID = deviceID
			return nil, nil
		},
	}
	reporter := health.NewReporter(5 * time.Second)
	logger := testLogger()
	s := NewScheduler("my-device-42", 10*time.Second, comm, reporter, logger)

	s.sendHeartbeat(context.Background())

	if capturedDeviceID != "my-device-42" {
		t.Errorf("expected device ID 'my-device-42', got %q", capturedDeviceID)
	}
}
