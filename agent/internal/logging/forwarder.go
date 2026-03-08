package logging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fleetml/fleetml/agent/internal/communication"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Forwarder buffers structured log entries and periodically sends them
// to the control plane. It implements zapcore.Core so it can be attached
// to the agent's zap logger as a tee — all log output is forwarded.
type Forwarder struct {
	comm     communication.Communicator
	deviceID string
	logger   *zap.SugaredLogger

	mu      sync.Mutex
	buffer  []communication.LogEntry
	maxBuf  int
	fields  []zapcore.Field
	flushCh chan struct{}
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// NewForwarder creates a log forwarder that batches and sends agent logs.
func NewForwarder(comm communication.Communicator, deviceID string, logger *zap.SugaredLogger) *Forwarder {
	f := &Forwarder{
		comm:     comm,
		deviceID: deviceID,
		logger:   logger,
		buffer:   make([]communication.LogEntry, 0, 64),
		maxBuf:   256,
		flushCh:  make(chan struct{}, 1),
		stopCh:   make(chan struct{}),
	}
	return f
}

// Start begins the background flush loop.
func (f *Forwarder) Start(interval time.Duration) {
	f.wg.Add(1)
	go f.flushLoop(interval)
}

// Stop flushes remaining logs and stops the forwarder.
func (f *Forwarder) Stop() {
	close(f.stopCh)
	f.wg.Wait()
	f.flush() // final flush
}

// Enabled implements zapcore.Core.
func (f *Forwarder) Enabled(lvl zapcore.Level) bool {
	return lvl >= zapcore.InfoLevel
}

// With implements zapcore.Core.
func (f *Forwarder) With(fields []zapcore.Field) zapcore.Core {
	clone := &Forwarder{
		comm:     f.comm,
		deviceID: f.deviceID,
		logger:   f.logger,
		buffer:   f.buffer,
		maxBuf:   f.maxBuf,
		flushCh:  f.flushCh,
		stopCh:   f.stopCh,
		fields:   append(f.fields[:len(f.fields):len(f.fields)], fields...),
	}
	return clone
}

// Check implements zapcore.Core.
func (f *Forwarder) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if f.Enabled(entry.Level) {
		return ce.AddCore(entry, f)
	}
	return ce
}

// Write implements zapcore.Core — buffers the log entry for forwarding.
func (f *Forwarder) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Extract component from fields if present
	component := "agent"
	metadata := make(map[string]string)

	enc := zapcore.NewMapObjectEncoder()
	for _, field := range append(f.fields, fields...) {
		field.AddTo(enc)
	}
	for k, v := range enc.Fields {
		if k == "component" {
			if s, ok := v.(string); ok {
				component = s
				continue
			}
		}
		metadata[k] = fmt.Sprint(v)
	}

	logEntry := communication.LogEntry{
		Timestamp: entry.Time,
		Level:     entry.Level.String(),
		Component: component,
		Message:   entry.Message,
	}
	if len(metadata) > 0 {
		logEntry.Metadata = metadata
	}

	f.mu.Lock()
	f.buffer = append(f.buffer, logEntry)
	shouldFlush := len(f.buffer) >= f.maxBuf
	f.mu.Unlock()

	if shouldFlush {
		select {
		case f.flushCh <- struct{}{}:
		default:
		}
	}

	return nil
}

// Sync implements zapcore.Core.
func (f *Forwarder) Sync() error {
	f.flush()
	return nil
}

func (f *Forwarder) flushLoop(interval time.Duration) {
	defer f.wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			f.flush()
		case <-f.flushCh:
			f.flush()
		case <-f.stopCh:
			return
		}
	}
}

func (f *Forwarder) flush() {
	f.mu.Lock()
	if len(f.buffer) == 0 {
		f.mu.Unlock()
		return
	}
	batch := f.buffer
	f.buffer = make([]communication.LogEntry, 0, 64)
	f.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := f.comm.SendLogs(ctx, f.deviceID, batch); err != nil {
		// Log forwarding is best-effort — don't crash
		if f.logger != nil {
			f.logger.Debugw("failed to forward logs", "entries", len(batch), "error", err)
		}
	}
}
