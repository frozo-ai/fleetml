package communication

import (
	"context"
	"time"

	"github.com/fleetml/fleetml/agent/internal/device"
	"github.com/fleetml/fleetml/agent/internal/health"
)

// Communicator defines the interface for agent-server communication.
type Communicator interface {
	// Register registers this device with the control plane.
	Register(ctx context.Context, info *device.Info) (agentID string, heartbeatInterval int, err error)

	// SendHeartbeat sends a heartbeat with system and model metrics.
	SendHeartbeat(ctx context.Context, deviceID string, status string, system *health.SystemMetrics) ([]Command, error)

	// ReportDeploymentStatus reports deployment progress to the server.
	ReportDeploymentStatus(ctx context.Context, deviceID, deploymentID, state, errMsg string) error

	// SendLogs sends a batch of structured log entries to the control plane.
	SendLogs(ctx context.Context, deviceID string, entries []LogEntry) error

	// Close closes the connection.
	Close() error
}

// LogEntry represents a structured log entry from the agent.
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`     // debug, info, warn, error
	Component string            `json:"component"` // agent, runtime, deploy, heartbeat
	Message   string            `json:"message"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Command represents a command from the control plane.
type Command struct {
	ID      string `json:"id"`
	Type    string `json:"type"`    // deploy_model, rollback, update_config, restart, set_ab_test
	Payload []byte `json:"payload"` // JSON-encoded payload
}
