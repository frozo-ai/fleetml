package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/fleetml/fleetml/agent/internal/device"
	"github.com/fleetml/fleetml/agent/internal/health"
	pb "github.com/fleetml/fleetml/proto/gen/go/fleetml/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCClient implements the Communicator interface using gRPC.
type GRPCClient struct {
	conn       *grpc.ClientConn
	client     pb.AgentServiceClient
	stream     pb.AgentService_HeartbeatClient
	logger     *zap.SugaredLogger
	address    string
	restURL    string // HTTP base URL for REST endpoints (log ingestion)
	httpClient *http.Client
}

// NewGRPCClient creates a new gRPC client with exponential backoff.
// restURL is the HTTP base URL for REST-only endpoints (e.g. log ingestion).
func NewGRPCClient(address string, logger *zap.SugaredLogger) (*GRPCClient, error) {
	return &GRPCClient{
		address: address,
		logger:  logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// SetRESTURL sets the HTTP base URL used for REST-only endpoints like log ingestion.
func (c *GRPCClient) SetRESTURL(url string) {
	c.restURL = url
}

// Connect establishes the gRPC connection with exponential backoff.
func (c *GRPCClient) Connect(ctx context.Context) error {
	var lastErr error
	maxRetries := 10
	baseDelay := 1 * time.Second
	maxDelay := 60 * time.Second

	for i := 0; i < maxRetries; i++ {
		conn, err := grpc.Dial(c.address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			lastErr = err
			delay := time.Duration(math.Min(
				float64(baseDelay)*math.Pow(2, float64(i)),
				float64(maxDelay),
			))
			c.logger.Warnw("gRPC connection failed, retrying",
				"attempt", i+1,
				"delay", delay,
				"error", err,
			)
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		c.conn = conn
		c.client = pb.NewAgentServiceClient(conn)
		c.logger.Infow("gRPC connection established", "address", c.address)
		return nil
	}

	return fmt.Errorf("failed to connect after %d retries: %w", maxRetries, lastErr)
}

// Register registers this device with the control plane.
func (c *GRPCClient) Register(ctx context.Context, info *device.Info) (string, int, error) {
	labels := make(map[string]string)
	for k, v := range info.Labels {
		labels[k] = v
	}

	resp, err := c.client.Register(ctx, &pb.RegisterRequest{
		DeviceInfo: &pb.DeviceInfo{
			DeviceId:      info.DeviceID,
			Arch:          info.Arch,
			GpuType:       info.GPUType,
			Runtime:       info.Runtime,
			RamMb:         int32(info.RAMMB),
			DiskGb:        int32(info.DiskGB),
			Os:            info.OS,
			HardwareModel: info.HardwareModel,
			Labels:        labels,
		},
	})
	if err != nil {
		return "", 0, fmt.Errorf("register: %w", err)
	}

	c.logger.Infow("registered with control plane",
		"agent_id", resp.AgentId,
		"heartbeat_interval", resp.HeartbeatIntervalSec,
	)

	return resp.AgentId, int(resp.HeartbeatIntervalSec), nil
}

// SendHeartbeat sends a heartbeat and returns any commands.
func (c *GRPCClient) SendHeartbeat(ctx context.Context, deviceID string, status string, system *health.SystemMetrics) ([]Command, error) {
	// Use unary fallback if stream is not established
	if c.stream == nil {
		var err error
		c.stream, err = c.client.Heartbeat(ctx)
		if err != nil {
			return nil, fmt.Errorf("start heartbeat stream: %w", err)
		}
	}

	req := &pb.HeartbeatRequest{
		DeviceId:  deviceID,
		Timestamp: timestamppb.Now(),
		Status:    status,
	}

	if system != nil {
		req.System = &pb.SystemMetrics{
			CpuPercent:   float32(system.CPUPercent),
			GpuPercent:   float32(system.GPUPercent),
			RamMbUsed:    int32(system.RAMMBUsed),
			DiskPercent:  float32(system.DiskPercent),
			TemperatureC: float32(system.TemperatureC),
			UptimeHours:  float32(system.UptimeHours),
		}
	}

	if err := c.stream.Send(req); err != nil {
		c.stream = nil // Reset stream on error
		return nil, fmt.Errorf("send heartbeat: %w", err)
	}

	resp, err := c.stream.Recv()
	if err != nil {
		c.stream = nil
		return nil, fmt.Errorf("receive heartbeat response: %w", err)
	}

	var commands []Command
	for _, cmd := range resp.Commands {
		commands = append(commands, Command{
			ID:      cmd.Id,
			Type:    cmd.Type,
			Payload: cmd.Payload,
		})
	}

	return commands, nil
}

// ReportDeploymentStatus reports deployment progress.
func (c *GRPCClient) ReportDeploymentStatus(ctx context.Context, deviceID, deploymentID, state, errMsg string) error {
	_, err := c.client.ReportDeploymentStatus(ctx, &pb.DeploymentStatusReport{
		DeviceId:     deviceID,
		DeploymentId: deploymentID,
		State:        state,
		Error:        errMsg,
		Timestamp:    timestamppb.Now(),
	})
	return err
}

// SendLogs sends a batch of log entries to the server via REST API.
func (c *GRPCClient) SendLogs(ctx context.Context, deviceID string, entries []LogEntry) error {
	if c.restURL == "" {
		return fmt.Errorf("REST URL not configured for log ingestion")
	}

	payload := struct {
		DeviceID string     `json:"device_id"`
		Entries  []LogEntry `json:"entries"`
	}{
		DeviceID: deviceID,
		Entries:  entries,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal log entries: %w", err)
	}

	url := c.restURL + "/api/v1/devices/" + deviceID + "/logs"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create log request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("log ingestion failed with status %d", resp.StatusCode)
	}

	return nil
}

// Close closes the gRPC connection.
func (c *GRPCClient) Close() error {
	if c.stream != nil {
		c.stream.CloseSend()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
