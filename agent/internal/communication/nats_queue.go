package communication

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// NATSCommandMessage matches the server's published format.
type NATSCommandMessage struct {
	DeviceID     string            `json:"device_id"`
	CommandID    string            `json:"command_id"`
	CommandType  string            `json:"command_type"`
	DeploymentID string            `json:"deployment_id,omitempty"`
	Payload      map[string]string `json:"payload"`
	IssuedAt     time.Time         `json:"issued_at"`
}

// NATSCommandQueue subscribes to a device-specific NATS subject and delivers
// commands via a channel. This provides real-time command delivery without
// waiting for the next heartbeat cycle.
type NATSCommandQueue struct {
	conn     *nats.Conn
	sub      *nats.Subscription
	deviceID string
	cmdCh    chan Command
	logger   *zap.SugaredLogger
}

// NewNATSCommandQueue creates a NATS-backed command queue for the given device.
func NewNATSCommandQueue(natsURL, deviceID string, cmdCh chan Command, logger *zap.SugaredLogger) (*NATSCommandQueue, error) {
	opts := []nats.Option{
		nats.Name("fleetml-agent-" + deviceID),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(-1),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if logger != nil {
				logger.Warnw("NATS disconnected", "error", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			if logger != nil {
				logger.Infow("NATS reconnected", "url", nc.ConnectedUrl())
			}
		}),
	}

	conn, err := nats.Connect(natsURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}

	q := &NATSCommandQueue{
		conn:     conn,
		deviceID: deviceID,
		cmdCh:    cmdCh,
		logger:   logger,
	}

	// Subscribe to device-specific command subject
	subject := fmt.Sprintf("fleetml.commands.%s", deviceID)
	sub, err := conn.Subscribe(subject, q.handleMessage)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("subscribe to %s: %w", subject, err)
	}
	q.sub = sub

	if logger != nil {
		logger.Infow("NATS command queue started",
			"device_id", deviceID,
			"subject", subject,
		)
	}

	return q, nil
}

func (q *NATSCommandQueue) handleMessage(msg *nats.Msg) {
	var natsCmd NATSCommandMessage
	if err := json.Unmarshal(msg.Data, &natsCmd); err != nil {
		if q.logger != nil {
			q.logger.Errorw("failed to unmarshal NATS command", "error", err)
		}
		return
	}

	// Convert to internal Command format
	payloadJSON, _ := json.Marshal(natsCmd.Payload)
	cmd := Command{
		ID:      natsCmd.CommandID,
		Type:    natsCmd.CommandType,
		Payload: payloadJSON,
	}

	if q.logger != nil {
		q.logger.Infow("received command via NATS",
			"command_id", cmd.ID,
			"type", cmd.Type,
		)
	}

	// Non-blocking send to command channel
	select {
	case q.cmdCh <- cmd:
	default:
		if q.logger != nil {
			q.logger.Warnw("command channel full, dropping NATS command",
				"command_id", cmd.ID,
			)
		}
	}
}

// IsConnected returns whether the NATS connection is active.
func (q *NATSCommandQueue) IsConnected() bool {
	return q.conn != nil && q.conn.IsConnected()
}

// Close unsubscribes and closes the NATS connection.
func (q *NATSCommandQueue) Close() error {
	if q.sub != nil {
		q.sub.Unsubscribe()
	}
	if q.conn != nil {
		q.conn.Drain()
		q.conn.Close()
	}
	return nil
}
