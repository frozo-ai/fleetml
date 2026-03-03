package communication

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestConnectionMonitor_InitiallyConnected(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cm := NewConnectionMonitor(logger.Sugar())

	if cm.State() != Connected {
		t.Error("expected Connected initially")
	}
	if cm.ConsecutiveFailures() != 0 {
		t.Error("expected 0 failures initially")
	}
}

func TestConnectionMonitor_RecordFailure(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cm := NewConnectionMonitor(logger.Sugar())

	cm.RecordFailure()

	if cm.State() != Disconnected {
		t.Error("expected Disconnected after failure")
	}
	if cm.ConsecutiveFailures() != 1 {
		t.Errorf("expected 1 failure, got %d", cm.ConsecutiveFailures())
	}
}

func TestConnectionMonitor_RecordSuccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cm := NewConnectionMonitor(logger.Sugar())

	cm.RecordFailure()
	cm.RecordFailure()
	cm.RecordSuccess()

	if cm.State() != Connected {
		t.Error("expected Connected after success")
	}
	if cm.ConsecutiveFailures() != 0 {
		t.Error("expected 0 failures after success")
	}
}

func TestConnectionMonitor_ExponentialBackoff(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cm := NewConnectionMonitor(logger.Sugar())

	// No failures = no delay
	if cm.NextRetryDelay() != 0 {
		t.Error("expected 0 delay with no failures")
	}

	cm.RecordFailure() // 1 failure
	d1 := cm.NextRetryDelay()
	if d1 != 1*time.Second {
		t.Errorf("expected 1s delay, got %v", d1)
	}

	cm.RecordFailure() // 2 failures
	d2 := cm.NextRetryDelay()
	if d2 != 2*time.Second {
		t.Errorf("expected 2s delay, got %v", d2)
	}

	cm.RecordFailure() // 3 failures
	d3 := cm.NextRetryDelay()
	if d3 != 4*time.Second {
		t.Errorf("expected 4s delay, got %v", d3)
	}

	// Keep failing to hit the cap
	for i := 0; i < 10; i++ {
		cm.RecordFailure()
	}
	dMax := cm.NextRetryDelay()
	if dMax != 60*time.Second {
		t.Errorf("expected max delay 60s, got %v", dMax)
	}
}

func TestConnectionMonitor_ShouldRetry(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cm := NewConnectionMonitor(logger.Sugar())

	// Unlimited retries by default
	for i := 0; i < 100; i++ {
		cm.RecordFailure()
	}
	if !cm.ShouldRetry() {
		t.Error("should always retry with unlimited retries")
	}
}

func TestConnectionMonitor_DisconnectCallback(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cm := NewConnectionMonitor(logger.Sugar())

	disconnected := make(chan struct{}, 1)
	cm.OnDisconnect(func() {
		disconnected <- struct{}{}
	})

	cm.RecordFailure()

	select {
	case <-disconnected:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Error("disconnect callback not called")
	}
}

func TestConnectionMonitor_ReconnectCallback(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cm := NewConnectionMonitor(logger.Sugar())

	reconnected := make(chan struct{}, 1)
	cm.OnReconnect(func() {
		reconnected <- struct{}{}
	})

	cm.RecordFailure()
	cm.RecordSuccess()

	select {
	case <-reconnected:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Error("reconnect callback not called")
	}
}

func TestConnectionMonitor_NoCallbackOnInitialSuccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cm := NewConnectionMonitor(logger.Sugar())

	reconnected := make(chan struct{}, 1)
	cm.OnReconnect(func() {
		reconnected <- struct{}{}
	})

	// First success should not trigger reconnect callback
	cm.RecordSuccess()

	select {
	case <-reconnected:
		t.Error("reconnect callback should not fire on initial success")
	case <-time.After(50 * time.Millisecond):
		// OK
	}
}
