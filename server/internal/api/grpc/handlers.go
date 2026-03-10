package grpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/fleetml/fleetml/server/internal/deploy"
	"github.com/fleetml/fleetml/server/internal/domain"
	"github.com/fleetml/fleetml/server/internal/fleet"
	servermodel "github.com/fleetml/fleetml/server/internal/model"
	"github.com/fleetml/fleetml/server/internal/monitor"
	"github.com/fleetml/fleetml/server/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/fleetml/fleetml/proto/gen/go/fleetml/v1"
)

// Handler implements the AgentService gRPC service.
type Handler struct {
	pb.UnimplementedAgentServiceServer
	fleet        *fleet.Manager
	orchestrator *deploy.Orchestrator
	registry     *servermodel.Registry
	store        storage.ObjectStore
	metrics      *monitor.MetricsProcessor
	db           *pgxpool.Pool
	logger       *zap.SugaredLogger
}

func NewHandler(fleetMgr *fleet.Manager, orchestrator *deploy.Orchestrator, registry *servermodel.Registry, store storage.ObjectStore, metrics *monitor.MetricsProcessor, db *pgxpool.Pool, logger *zap.SugaredLogger) *Handler {
	return &Handler{fleet: fleetMgr, orchestrator: orchestrator, registry: registry, store: store, metrics: metrics, db: db, logger: logger}
}

// resolveOrgFromAPIKey extracts the API key from gRPC metadata and returns the org ID.
func (h *Handler) resolveOrgFromAPIKey(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("no metadata")
	}

	keys := md.Get("x-api-key")
	if len(keys) == 0 || keys[0] == "" {
		return "", fmt.Errorf("no API key provided")
	}

	apiKey := keys[0]
	var orgID string
	err := h.db.QueryRow(ctx, `SELECT id FROM organizations WHERE api_key = $1`, apiKey).Scan(&orgID)
	if err != nil {
		return "", fmt.Errorf("invalid API key")
	}

	return orgID, nil
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

	// Validate API key and resolve org
	orgID, err := h.resolveOrgFromAPIKey(ctx)
	if err != nil {
		h.logger.Warnw("device registration rejected: invalid API key", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid or missing API key — get your key at https://dashboard-production-e3e4.up.railway.app/dashboard/get-started")
	}

	// Check device limit for this org
	if h.db != nil {
		var deviceCount int
		var deviceLimit int
		h.db.QueryRow(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1`, orgID).Scan(&deviceCount)
		h.db.QueryRow(ctx, `SELECT device_limit FROM organizations WHERE id = $1`, orgID).Scan(&deviceLimit)
		if deviceLimit > 0 && deviceCount >= deviceLimit {
			return nil, status.Errorf(codes.ResourceExhausted, "device limit reached (%d/%d) — upgrade your plan", deviceCount, deviceLimit)
		}
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

	// Assign device to org
	if h.db != nil {
		h.db.Exec(ctx, `UPDATE devices SET org_id = $1 WHERE id = $2`, orgID, device.ID)
	}

	h.logger.Infow("device registered",
		"device_id", info.DeviceId,
		"agent_id", device.ID,
		"org_id", orgID,
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

// GetModelArtifact streams model data to an agent from S3 storage.
func (h *Handler) GetModelArtifact(req *pb.GetModelArtifactRequest, stream pb.AgentService_GetModelArtifactServer) error {
	if h.registry == nil || h.store == nil {
		return status.Error(codes.Unavailable, "model registry or storage not configured")
	}

	// Look up the model
	m, err := h.registry.GetModel(stream.Context(), req.ModelName, req.ModelVersion)
	if err != nil {
		return status.Errorf(codes.NotFound, "model %s:%s not found: %v", req.ModelName, req.ModelVersion, err)
	}

	// Determine the S3 key — use compiled variant if runtime is specified and available
	s3Key := fmt.Sprintf("%s/%s/model.%s", m.Name, m.Version, m.Format)
	checksum := m.Checksum

	if req.Runtime != "" && req.Runtime != "onnx" && req.Runtime != m.Format {
		variant, err := h.registry.GetVariantForRuntime(stream.Context(), m.ID, req.Runtime)
		if err == nil && variant != nil {
			s3Key = extractS3Key(variant.ArtifactURL)
			checksum = variant.Checksum
		}
		// If no variant found, fall back to base ONNX
	}

	h.logger.Infow("streaming model artifact",
		"model", req.ModelName,
		"version", req.ModelVersion,
		"runtime", req.Runtime,
		"s3_key", s3Key,
	)

	// Download from S3
	reader, err := h.store.Download(stream.Context(), s3Key)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to download model from storage: %v", err)
	}
	defer reader.Close()

	// Stream in 64KB chunks
	const chunkSize = 64 * 1024
	buf := make([]byte, chunkSize)
	hasher := sha256.New()
	var totalSent int64

	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			hasher.Write(buf[:n])
			totalSent += int64(n)

			if err := stream.Send(&pb.ModelArtifactChunk{
				Data:      buf[:n],
				TotalSize: m.ArtifactSize,
				Checksum:  checksum,
			}); err != nil {
				return status.Errorf(codes.Internal, "failed to send chunk: %v", err)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return status.Errorf(codes.Internal, "failed to read model data: %v", readErr)
		}
	}

	h.logger.Infow("model artifact streamed",
		"model", req.ModelName,
		"version", req.ModelVersion,
		"bytes_sent", totalSent,
		"checksum", "sha256:"+hex.EncodeToString(hasher.Sum(nil)),
	)

	return nil
}

// extractS3Key extracts the object key from an s3:// URL.
func extractS3Key(s3URL string) string {
	const prefix = "s3://"
	if len(s3URL) <= len(prefix) {
		return s3URL
	}
	rest := s3URL[len(prefix):]
	for i, c := range rest {
		if c == '/' {
			return rest[i+1:]
		}
	}
	return rest
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
