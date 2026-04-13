package communication

import (
	"context"
	"sync"
	"time"

	"github.com/fleetml/fleetml/agent/internal/device"
	"github.com/fleetml/fleetml/agent/internal/health"
	"github.com/fleetml/fleetml/agent/internal/offline"
	"go.uber.org/zap"
)

// StoreForwardManager wraps a Communicator with offline buffering.
// When the primary connection fails, heartbeats are buffered locally.
// On reconnection, buffered data is bulk-synced to the server.
type StoreForwardManager struct {
	primary Communicator
	store   offline.MetricsStore
	logger  *zap.SugaredLogger

	mu        sync.RWMutex
	connected bool

	// Bulk sync config
	syncBatchSize int
	syncInterval  time.Duration
}

// NewStoreForwardManager creates a new store-and-forward wrapper.
func NewStoreForwardManager(
	primary Communicator,
	store offline.MetricsStore,
	logger *zap.SugaredLogger,
) *StoreForwardManager {
	return &StoreForwardManager{
		primary:       primary,
		store:         store,
		logger:        logger,
		connected:     true,
		syncBatchSize: 100,
		syncInterval:  30 * time.Second,
	}
}

// Register delegates to the primary communicator.
func (sf *StoreForwardManager) Register(ctx context.Context, info *device.Info) (string, int, error) {
	agentID, interval, err := sf.primary.Register(ctx, info)
	if err != nil {
		sf.setConnected(false)
		return "", 0, err
	}
	sf.setConnected(true)
	return agentID, interval, nil
}

// SendHeartbeat tries the primary communicator; on failure, buffers locally.
func (sf *StoreForwardManager) SendHeartbeat(ctx context.Context, deviceID, status string, system *health.SystemMetrics) ([]Command, error) {
	// Try primary
	commands, err := sf.primary.SendHeartbeat(ctx, deviceID, status, system)
	if err == nil {
		if !sf.isConnected() {
			sf.setConnected(true)
			sf.logger.Infow("connection restored, starting bulk sync")
			go sf.bulkSync(ctx, deviceID)
		}
		return commands, nil
	}

	// Primary failed, buffer locally
	sf.setConnected(false)
	sf.logger.Warnw("connection lost, buffering heartbeat",
		"device_id", deviceID,
		"error", err,
	)

	record := &offline.HeartbeatRecord{
		DeviceID:  deviceID,
		Status:    status,
		System:    system,
		Timestamp: time.Now().Unix(),
	}

	if storeErr := sf.store.SaveHeartbeat(record); storeErr != nil {
		sf.logger.Errorw("failed to buffer heartbeat",
			"error", storeErr,
		)
	}

	return nil, nil // Swallow error to prevent heartbeat loop from failing
}

// ReportDeploymentStatus delegates to primary; no buffering for status reports.
func (sf *StoreForwardManager) ReportDeploymentStatus(ctx context.Context, deviceID, deploymentID, state, errMsg string) error {
	return sf.primary.ReportDeploymentStatus(ctx, deviceID, deploymentID, state, errMsg)
}

// SendLogs delegates to primary. Logs are best-effort — dropped if offline.
func (sf *StoreForwardManager) SendLogs(ctx context.Context, deviceID string, entries []LogEntry) error {
	if !sf.isConnected() {
		sf.logger.Debugw("offline, dropping log batch", "entries", len(entries))
		return nil
	}
	return sf.primary.SendLogs(ctx, deviceID, entries)
}

// Close closes both the primary communicator and the local store.
func (sf *StoreForwardManager) Close() error {
	if err := sf.primary.Close(); err != nil {
		sf.logger.Warnw("error closing primary communicator", "error", err)
	}
	return sf.store.Close()
}

// IsConnected returns whether the primary connection is active.
func (sf *StoreForwardManager) IsConnected() bool {
	return sf.isConnected()
}

// BufferedCount returns the number of buffered heartbeats.
func (sf *StoreForwardManager) BufferedCount() int {
	count, err := sf.store.BufferCount()
	if err != nil {
		return -1
	}
	return count
}

// bulkSync sends all buffered heartbeats to the server in batches.
func (sf *StoreForwardManager) bulkSync(ctx context.Context, deviceID string) {
	records, err := sf.store.GetBufferedHeartbeats()
	if err != nil {
		sf.logger.Errorw("failed to get buffered heartbeats", "error", err)
		return
	}

	if len(records) == 0 {
		return
	}

	// Record the latest timestamp we're syncing — only clear up to this point
	var maxTimestamp int64
	for _, r := range records {
		if r.Timestamp > maxTimestamp {
			maxTimestamp = r.Timestamp
		}
	}

	sf.logger.Infow("starting bulk sync",
		"buffered_records", len(records),
		"device_id", deviceID,
	)

	synced := 0
	for i := 0; i < len(records); i += sf.syncBatchSize {
		end := i + sf.syncBatchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		for _, r := range batch {
			_, err := sf.primary.SendHeartbeat(ctx, r.DeviceID, r.Status, r.System)
			if err != nil {
				sf.logger.Warnw("bulk sync interrupted",
					"synced", synced,
					"remaining", len(records)-synced,
					"error", err,
				)
				return
			}
			synced++
		}
	}

	// Only clear records up to the timestamp we synced — new ones are preserved
	if err := sf.store.ClearBufferBefore(maxTimestamp); err != nil {
		sf.logger.Errorw("failed to clear buffer after sync", "error", err)
		return
	}

	sf.logger.Infow("bulk sync completed",
		"synced_records", synced,
	)
}

func (sf *StoreForwardManager) isConnected() bool {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.connected
}

func (sf *StoreForwardManager) setConnected(connected bool) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.connected = connected
}
