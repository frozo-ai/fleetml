package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/fleetml/fleetml/server/internal/deploy"
	"github.com/fleetml/fleetml/server/internal/domain"
	"github.com/fleetml/fleetml/server/internal/fleet"
	"github.com/fleetml/fleetml/server/internal/monitor"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/fleetml/fleetml/proto/gen/go/fleetml/v1"
)

// Handler implements the AgentService gRPC service.
type Handler struct {
	pb.UnimplementedAgentServiceServer
	fleet        *fleet.Manager
	orchestrator *deploy.Orchestrator
	metrics      *monitor.MetricsProcessor
	logger       *zap.SugaredLogger
}

func NewHandler(fleetMgr *fleet.Manager, orchestrator *deploy.Orchestrator, metrics *monitor.MetricsProcessor, logger *zap.SugaredLogger) *Handler {
	return &Handler{fleet: fleetMgr, orchestrator: orchestrator, metrics: metrics, logger: logger}
}

// RegisterService registers the handler with the gRPC server.
func (h *Handler) RegisterService(s *grpc.Server) {
	pb.RegisterAgentServiceServer(s, h)
}

// Register handles device registration (implements AgentServiceServer).
func (h *Handler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.DeviceInfo == nil {
		return nil, status.Error(codes.InvalidArgument, "device_info is required")
	}

	info := req.DeviceInfo
	labels := make(map[string]string)
	for k, v := range info.Labels {
		labels[k] = v
	}

	device, err := h.fleet.RegisterDevice(ctx, &domain.Device{
		DeviceID:      info.DeviceId,
		Arch:          info.Arch,
		GPUType:       info.GpuType,
		Runtime:       info.Runtime,
		RAMMB:         int(info.RamMb),
		DiskGB:        int(info.DiskGb),
		OS:            info.Os,
		HardwareModel: info.HardwareModel,
		Labels:        labels,
	})
	if err != nil {
		h.logger.Errorw("failed to register device", "device_id", info.DeviceId, "error", err)
		return nil, status.Error(codes.Internal, "failed to register device")
	}

	h.logger.Infow("device registered",
		"device_id", info.DeviceId,
		"agent_id", device.ID,
		"arch", info.Arch,
		"gpu", info.GpuType,
	)

	return &pb.RegisterResponse{
		AgentId:              device.ID,
		HeartbeatIntervalSec: 30,
	}, nil
}

// Heartbeat handles bidirectional heartbeat streaming.
func (h *Handler) Heartbeat(stream pb.AgentService_HeartbeatServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Error(codes.Internal, "receive heartbeat: "+err.Error())
		}

		h.logger.Debugw("heartbeat received",
			"device_id", req.DeviceId,
			"status", req.Status,
		)

		// Update device status
		var cpuPct, gpuPct, diskPct, tempC, uptimeH *float64
		var ramUsed *int
		if req.System != nil {
			cpu := float64(req.System.CpuPercent)
			cpuPct = &cpu
			gpu := float64(req.System.GpuPercent)
			gpuPct = &gpu
			disk := float64(req.System.DiskPercent)
			diskPct = &disk
			temp := float64(req.System.TemperatureC)
			tempC = &temp
			uptime := float64(req.System.UptimeHours)
			uptimeH = &uptime
			ram := int(req.System.RamMbUsed)
			ramUsed = &ram
		}

		if err := h.fleet.UpdateDeviceStatus(
			stream.Context(), req.DeviceId, req.Status,
			cpuPct, gpuPct, diskPct, tempC, uptimeH, ramUsed,
		); err != nil {
			h.logger.Errorw("failed to update device status", "device_id", req.DeviceId, "error", err)
		}

		// Store heartbeat metrics in heartbeats table
		if h.metrics != nil && req.System != nil {
			if err := h.metrics.ProcessHeartbeat(
				stream.Context(), req.DeviceId,
				float64(req.System.CpuPercent), float64(req.System.GpuPercent),
				float64(req.System.DiskPercent), float64(req.System.TemperatureC),
				float64(req.System.UptimeHours), int(req.System.RamMbUsed), nil,
			); err != nil {
				h.logger.Warnw("failed to store heartbeat metrics", "device_id", req.DeviceId, "error", err)
			}
		}

		// Check for pending deployment commands
		var pbCommands []*pb.Command
		if h.orchestrator != nil {
			cmds, err := h.orchestrator.GetPendingCommands(stream.Context(), req.DeviceId)
			if err != nil {
				h.logger.Warnw("failed to get pending commands", "device_id", req.DeviceId, "error", err)
			}
			for _, cmd := range cmds {
				payloadJSON, _ := json.Marshal(cmd["payload"])
				pbCommands = append(pbCommands, &pb.Command{
					Id:       fmt.Sprintf("cmd-%s", req.DeviceId),
					Type:     cmd["type"].(string),
					Payload:  payloadJSON,
					IssuedAt: timestamppb.Now(),
				})
			}
		}

		resp := &pb.HeartbeatResponse{
			Commands: pbCommands,
		}

		if err := stream.Send(resp); err != nil {
			return status.Error(codes.Internal, "send heartbeat response: "+err.Error())
		}
	}
}

// ReportDeploymentStatus handles deployment status reports from agents.
func (h *Handler) ReportDeploymentStatus(ctx context.Context, req *pb.DeploymentStatusReport) (*pb.Empty, error) {
	h.logger.Infow("deployment status report",
		"device_id", req.DeviceId,
		"deployment_id", req.DeploymentId,
		"state", req.State,
		"error", req.Error,
	)

	if h.orchestrator != nil {
		if err := h.orchestrator.HandleDeploymentReport(ctx, req.DeviceId, req.DeploymentId, req.State, req.Error); err != nil {
			h.logger.Errorw("failed to handle deployment report",
				"device_id", req.DeviceId,
				"deployment_id", req.DeploymentId,
				"error", err,
			)
			return nil, status.Error(codes.Internal, "failed to process deployment report")
		}
	}

	return &pb.Empty{}, nil
}

// GetModelArtifact streams model data to an agent.
func (h *Handler) GetModelArtifact(req *pb.GetModelArtifactRequest, stream pb.AgentService_GetModelArtifactServer) error {
	// TODO: Implement model streaming from S3
	return status.Error(codes.Unimplemented, "GetModelArtifact not yet implemented")
}

// BulkSyncMetrics handles bulk metric sync after offline period.
func (h *Handler) BulkSyncMetrics(stream pb.AgentService_BulkSyncMetricsServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.BulkSyncResponse{
				PendingCommands: []*pb.Command{},
			})
		}
		if err != nil {
			return status.Error(codes.Internal, "receive bulk metrics: "+err.Error())
		}

		h.logger.Infow("bulk sync received",
			"device_id", req.DeviceId,
			"heartbeats", len(req.Heartbeats),
		)

		// Process each buffered heartbeat
		for _, hb := range req.Heartbeats {
			if h.metrics != nil && hb.System != nil {
				h.metrics.ProcessHeartbeat(
					stream.Context(), req.DeviceId,
					float64(hb.System.CpuPercent), float64(hb.System.GpuPercent),
					float64(hb.System.DiskPercent), float64(hb.System.TemperatureC),
					float64(hb.System.UptimeHours), int(hb.System.RamMbUsed), nil,
				)
			}
			if err := h.fleet.UpdateDeviceStatus(
				stream.Context(), req.DeviceId, hb.Status,
				nil, nil, nil, nil, nil, nil,
			); err != nil {
				h.logger.Warnw("failed to update device status from bulk sync", "error", err)
			}
		}
	}
}
