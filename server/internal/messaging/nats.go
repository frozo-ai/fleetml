package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// CommandMessage is published to NATS when a command is created for a device.
type CommandMessage struct {
	DeviceID     string            `json:"device_id"`
	CommandID    string            `json:"command_id"`
	CommandType  string            `json:"command_type"`
	DeploymentID string            `json:"deployment_id,omitempty"`
	Payload      map[string]string `json:"payload"`
	IssuedAt     time.Time         `json:"issued_at"`
}

// NATSClient wraps a NATS connection for publishing commands and subscribing to events.
type NATSClient struct {
	conn   *nats.Conn
	logger *zap.SugaredLogger
}

// NewNATSClient creates a new NATS client and connects to the server.
func NewNATSClient(url string, logger *zap.SugaredLogger) (*NATSClient, error) {
	opts := []nats.Option{
		nats.Name("fleetml-server"),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(-1), // infinite reconnects
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

	conn, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}

	return &NATSClient{conn: conn, logger: logger}, nil
}

// PublishCommand publishes a command to a device-specific NATS subject.
// Subject format: fleetml.commands.<device_id>
func (c *NATSClient) PublishCommand(ctx context.Context, msg CommandMessage) error {
	subject := fmt.Sprintf("fleetml.commands.%s", msg.DeviceID)

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal command: %w", err)
	}

	if err := c.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("publish command: %w", err)
	}

	if c.logger != nil {
		c.logger.Debugw("published command",
			"subject", subject,
			"command_type", msg.CommandType,
			"device_id", msg.DeviceID,
		)
	}

	return nil
}

// PublishEvent publishes a fleet event (e.g., deployment created, device offline).
// Subject format: fleetml.events.<event_type>
func (c *NATSClient) PublishEvent(ctx context.Context, eventType string, payload interface{}) error {
	subject := fmt.Sprintf("fleetml.events.%s", eventType)

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return c.conn.Publish(subject, data)
}

// SubscribeCommands subscribes to commands for a specific device.
// Returns a channel of CommandMessage.
func (c *NATSClient) SubscribeCommands(deviceID string, handler func(CommandMessage)) (*nats.Subscription, error) {
	subject := fmt.Sprintf("fleetml.commands.%s", deviceID)

	sub, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
		var cmd CommandMessage
		if err := json.Unmarshal(msg.Data, &cmd); err != nil {
			if c.logger != nil {
				c.logger.Errorw("failed to unmarshal command", "error", err)
			}
			return
		}
		handler(cmd)
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe to %s: %w", subject, err)
	}

	return sub, nil
}

// SubscribeAllCommands subscribes to all device commands (for monitoring/logging).
func (c *NATSClient) SubscribeAllCommands(handler func(CommandMessage)) (*nats.Subscription, error) {
	subject := "fleetml.commands.*"

	sub, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
		var cmd CommandMessage
		if err := json.Unmarshal(msg.Data, &cmd); err != nil {
			return
		}
		handler(cmd)
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe to %s: %w", subject, err)
	}

	return sub, nil
}

// IsConnected returns whether the NATS connection is active.
func (c *NATSClient) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}

// Close drains and closes the NATS connection.
func (c *NATSClient) Close() error {
	if c.conn != nil {
		c.conn.Drain()
		c.conn.Close()
	}
	return nil
}
