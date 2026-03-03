package offline

import (
	"github.com/fleetml/fleetml/agent/internal/health"
)

// HeartbeatRecord stores a buffered heartbeat for offline sync.
type HeartbeatRecord struct {
	DeviceID  string                `json:"device_id"`
	Status    string                `json:"status"`
	System    *health.SystemMetrics `json:"system"`
	Timestamp int64                 `json:"timestamp"` // Unix timestamp
}

// MetricsStore defines the interface for offline metric storage.
type MetricsStore interface {
	// SaveHeartbeat buffers a heartbeat locally.
	SaveHeartbeat(record *HeartbeatRecord) error

	// GetBufferedHeartbeats retrieves all buffered heartbeats.
	GetBufferedHeartbeats() ([]*HeartbeatRecord, error)

	// ClearBuffer removes all buffered heartbeats.
	ClearBuffer() error

	// Close closes the store.
	Close() error
}
