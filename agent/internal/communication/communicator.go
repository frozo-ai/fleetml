package communication

import (
	"context"

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

	// Close closes the connection.
	Close() error
}

// Command represents a command from the control plane.
type Command struct {
	ID      string `json:"id"`
	Type    string `json:"type"`    // deploy_model, rollback, update_config, restart, set_ab_test
	Payload []byte `json:"payload"` // JSON-encoded payload
}
