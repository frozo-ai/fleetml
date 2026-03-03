package communication

import (
	"context"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ConnectionState represents the connection health state.
type ConnectionState int

const (
	Connected    ConnectionState = iota
	Disconnected
	Reconnecting
)

// ConnectionMonitor tracks connection health and manages reconnection
// with exponential backoff.
type ConnectionMonitor struct {
	mu              sync.RWMutex
	state           ConnectionState
	consecutiveFails int
	lastSuccess     time.Time
	lastAttempt     time.Time

	// Backoff config
	baseDelay time.Duration
	maxDelay  time.Duration
	maxRetries int

	logger *zap.SugaredLogger

	// Callbacks
	onDisconnect func()
	onReconnect  func()
}

// NewConnectionMonitor creates a new connection health monitor.
func NewConnectionMonitor(logger *zap.SugaredLogger) *ConnectionMonitor {
	return &ConnectionMonitor{
		state:      Connected,
		baseDelay:  1 * time.Second,
		maxDelay:   60 * time.Second,
		maxRetries: -1, // unlimited
		logger:     logger,
	}
}

// RecordSuccess records a successful communication.
func (cm *ConnectionMonitor) RecordSuccess() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	wasDisconnected := cm.state != Connected
	cm.state = Connected
	cm.consecutiveFails = 0
	cm.lastSuccess = time.Now()

	if wasDisconnected && cm.onReconnect != nil {
		go cm.onReconnect()
	}
}

// RecordFailure records a failed communication attempt.
func (cm *ConnectionMonitor) RecordFailure() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.consecutiveFails++
	cm.lastAttempt = time.Now()

	if cm.state == Connected {
		cm.state = Disconnected
		cm.logger.Warnw("connection lost",
			"consecutive_fails", cm.consecutiveFails,
		)
		if cm.onDisconnect != nil {
			go cm.onDisconnect()
		}
	}
}

// State returns the current connection state.
func (cm *ConnectionMonitor) State() ConnectionState {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.state
}

// ConsecutiveFailures returns the number of consecutive failures.
func (cm *ConnectionMonitor) ConsecutiveFailures() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.consecutiveFails
}

// NextRetryDelay calculates the next retry delay using exponential backoff.
func (cm *ConnectionMonitor) NextRetryDelay() time.Duration {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.consecutiveFails == 0 {
		return 0
	}

	delay := float64(cm.baseDelay) * math.Pow(2, float64(cm.consecutiveFails-1))
	if delay > float64(cm.maxDelay) {
		delay = float64(cm.maxDelay)
	}

	return time.Duration(delay)
}

// ShouldRetry returns whether a retry should be attempted.
func (cm *ConnectionMonitor) ShouldRetry() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.maxRetries >= 0 && cm.consecutiveFails >= cm.maxRetries {
		return false
	}

	return true
}

// TimeSinceLastSuccess returns the duration since last successful communication.
func (cm *ConnectionMonitor) TimeSinceLastSuccess() time.Duration {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.lastSuccess.IsZero() {
		return 0
	}
	return time.Since(cm.lastSuccess)
}

// OnDisconnect sets a callback for when connection is lost.
func (cm *ConnectionMonitor) OnDisconnect(fn func()) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.onDisconnect = fn
}

// OnReconnect sets a callback for when connection is restored.
func (cm *ConnectionMonitor) OnReconnect(fn func()) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.onReconnect = fn
}

// WaitForReconnect blocks until the connection is restored or context is cancelled.
func (cm *ConnectionMonitor) WaitForReconnect(ctx context.Context) error {
	for {
		if cm.State() == Connected {
			return nil
		}

		delay := cm.NextRetryDelay()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Will retry on next heartbeat cycle
		}
	}
}
