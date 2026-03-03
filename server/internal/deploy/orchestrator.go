package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fleetml/fleetml/server/internal/domain"
	"github.com/fleetml/fleetml/server/internal/fleet"
	servermodel "github.com/fleetml/fleetml/server/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Orchestrator manages deployment lifecycle.
type Orchestrator struct {
	db       *pgxpool.Pool
	fleet    *fleet.Manager
	registry *servermodel.Registry
	canary   *CanaryManager
	logger   *zap.SugaredLogger
}

func NewOrchestrator(db *pgxpool.Pool, fleetMgr *fleet.Manager, registry *servermodel.Registry, canary *CanaryManager, logger *zap.SugaredLogger) *Orchestrator {
	return &Orchestrator{
		db:       db,
		fleet:    fleetMgr,
		registry: registry,
		canary:   canary,
		logger:   logger,
	}
}

// CreateRequest contains the data needed to create a deployment.
type CreateRequest struct {
	ModelName    string            `json:"model_name"`
	ModelVersion string            `json:"model_version"`
	TargetType   string            `json:"target_type"`   // fleet, device, labels
	TargetID     string            `json:"target_id"`     // Fleet ID or device ID
	TargetLabels map[string]string `json:"target_labels"` // Label selector
	Policy       string            `json:"policy"`        // immediate, canary
	CanaryConfig *domain.CanaryConfig `json:"canary_config,omitempty"`
}

// CreateDeployment creates and starts a deployment.
func (o *Orchestrator) CreateDeployment(ctx context.Context, req CreateRequest) (*domain.Deployment, error) {
	// 1. Resolve model
	m, err := o.registry.GetModel(ctx, req.ModelName, req.ModelVersion)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	// 2. Resolve target devices
	devices, err := o.fleet.SelectDevices(ctx, req.TargetType, req.TargetID, req.TargetLabels)
	if err != nil {
		return nil, fmt.Errorf("select devices: %w", err)
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("no devices match the target")
	}

	// 3. Count online/offline
	var onlineCount, offlineCount int
	for _, d := range devices {
		if d.Status == "offline" {
			offlineCount++
		} else {
			onlineCount++
		}
	}

	policy := req.Policy
	if policy == "" {
		policy = "immediate"
	}

	var canaryJSON []byte
	if req.CanaryConfig != nil {
		canaryJSON, _ = json.Marshal(req.CanaryConfig)
	}

	var targetLabelsJSON []byte
	if req.TargetLabels != nil {
		targetLabelsJSON, _ = json.Marshal(req.TargetLabels)
	}

	// 4. Create deployment record
	now := time.Now()
	var d domain.Deployment
	err = o.db.QueryRow(ctx, `
		INSERT INTO deployments (model_id, target_type, target_fleet_id, target_labels,
			state, total_devices, queued_devices, deployment_policy, canary_config,
			started_at, created_at)
		VALUES ($1, $2, $3, $4, 'rolling_out', $5, $6, $7, $8, $9, $9)
		RETURNING id, state, total_devices, completed_devices, failed_devices, queued_devices,
			deployment_policy, started_at, created_at`,
		m.ID,
		req.TargetType,
		nilIfEmpty(req.TargetID),
		targetLabelsJSON,
		len(devices),
		offlineCount,
		policy,
		canaryJSON,
		now,
	).Scan(
		&d.ID, &d.State, &d.TotalDevices, &d.CompletedDevices,
		&d.FailedDevices, &d.QueuedDevices, &d.DeploymentPolicy,
		&d.StartedAt, &d.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create deployment: %w", err)
	}

	d.ModelID = m.ID

	// 5. Create per-device status rows
	for _, dev := range devices {
		state := "pending"
		if dev.Status == "offline" {
			state = "queued"
		}
		// For canary deployments, only the first stage's devices are "pending";
		// the rest start as "queued" and get activated stage by stage.
		if policy == "canary" && req.CanaryConfig != nil && dev.Status != "offline" {
			state = "queued"
		}

		_, err = o.db.Exec(ctx, `
			INSERT INTO deployment_device_status (deployment_id, device_id, state)
			VALUES ($1, $2, $3)`, d.ID, dev.ID, state)
		if err != nil {
			return nil, fmt.Errorf("create device status: %w", err)
		}
	}

	// 6. For canary deployments, initialize canary tracking and select first batch
	if policy == "canary" && req.CanaryConfig != nil && o.canary != nil {
		if err := o.canary.InitCanary(ctx, d.ID, req.CanaryConfig, len(devices)); err != nil {
			return nil, fmt.Errorf("init canary: %w", err)
		}

		// Select devices for stage 0
		firstStage := req.CanaryConfig.Stages[0]
		batchSize := o.canary.DevicesForStage(len(devices), firstStage)
		canaryDevices, err := o.canary.SelectCanaryDevices(ctx, d.ID, batchSize)
		if err != nil {
			return nil, fmt.Errorf("select canary devices: %w", err)
		}

		// Mark the selected devices as "pending" so they get commands
		for _, devID := range canaryDevices {
			o.db.Exec(ctx, `
				UPDATE deployment_device_status SET state = 'pending'
				WHERE deployment_id = $1 AND device_id = $2`, d.ID, devID)
		}

		if o.logger != nil {
			o.logger.Infow("canary stage 0 started",
				"deployment_id", d.ID,
				"canary_devices", len(canaryDevices),
				"stage_percent", firstStage.Percent,
			)
		}
	}

	return &d, nil
}

// GetDeployment returns deployment status with per-device details.
func (o *Orchestrator) GetDeployment(ctx context.Context, id string) (*domain.Deployment, error) {
	var d domain.Deployment
	var canaryJSON []byte

	err := o.db.QueryRow(ctx, `
		SELECT id, model_id, target_type, state, total_devices, completed_devices,
			   failed_devices, queued_devices, deployment_policy, canary_config,
			   COALESCE(error, ''), started_at, completed_at, created_at
		FROM deployments WHERE id = $1`, id,
	).Scan(
		&d.ID, &d.ModelID, &d.TargetType, &d.State, &d.TotalDevices,
		&d.CompletedDevices, &d.FailedDevices, &d.QueuedDevices,
		&d.DeploymentPolicy, &canaryJSON, &d.Error, &d.StartedAt,
		&d.CompletedAt, &d.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("deployment %s not found", id)
		}
		return nil, fmt.Errorf("get deployment: %w", err)
	}

	if canaryJSON != nil {
		json.Unmarshal(canaryJSON, &d.CanaryConfig)
	}

	return &d, nil
}

// ListDeployments lists deployments with optional filters.
func (o *Orchestrator) ListDeployments(ctx context.Context, state, modelName string) ([]*domain.Deployment, int, error) {
	query := `SELECT d.id, d.model_id, d.target_type, d.state, d.total_devices,
			         d.completed_devices, d.failed_devices, d.queued_devices,
			         d.deployment_policy, d.started_at, d.completed_at, d.created_at
			  FROM deployments d`
	args := []interface{}{}
	where := ""
	argIdx := 1

	if state != "" {
		where += fmt.Sprintf(" WHERE d.state = $%d", argIdx)
		args = append(args, state)
		argIdx++
	}

	query += where + " ORDER BY d.created_at DESC LIMIT 50"

	rows, err := o.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list deployments: %w", err)
	}
	defer rows.Close()

	var deployments []*domain.Deployment
	for rows.Next() {
		var d domain.Deployment
		err := rows.Scan(
			&d.ID, &d.ModelID, &d.TargetType, &d.State, &d.TotalDevices,
			&d.CompletedDevices, &d.FailedDevices, &d.QueuedDevices,
			&d.DeploymentPolicy, &d.StartedAt, &d.CompletedAt, &d.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan deployment: %w", err)
		}
		deployments = append(deployments, &d)
	}

	return deployments, len(deployments), nil
}

// CancelDeployment cancels a running deployment.
func (o *Orchestrator) CancelDeployment(ctx context.Context, id string) error {
	_, err := o.db.Exec(ctx, `
		UPDATE deployments SET state = 'cancelled', completed_at = NOW()
		WHERE id = $1 AND state IN ('pending', 'rolling_out')`, id)
	if err != nil {
		return fmt.Errorf("cancel deployment: %w", err)
	}
	return nil
}

// HandleDeploymentReport processes a deployment status report from an agent.
func (o *Orchestrator) HandleDeploymentReport(ctx context.Context, deviceID, deploymentID, state, errMsg string) error {
	now := time.Now()

	// Update per-device status
	var completedAt *time.Time
	if state == "completed" || state == "failed" {
		completedAt = &now
	}
	_, err := o.db.Exec(ctx, `
		UPDATE deployment_device_status
		SET state = $3, error = $4, completed_at = $5
		WHERE deployment_id = $1 AND device_id = (SELECT id FROM devices WHERE device_id = $2)`,
		deploymentID, deviceID, state, errMsg, completedAt,
	)
	if err != nil {
		return fmt.Errorf("update device status: %w", err)
	}

	// Update deployment counters
	_, err = o.db.Exec(ctx, `
		UPDATE deployments SET
			completed_devices = (SELECT COUNT(*) FROM deployment_device_status WHERE deployment_id = $1 AND state = 'completed'),
			failed_devices = (SELECT COUNT(*) FROM deployment_device_status WHERE deployment_id = $1 AND state = 'failed')
		WHERE id = $1`, deploymentID)
	if err != nil {
		return fmt.Errorf("update counters: %w", err)
	}

	// Check if deployment is done
	var total, completed, failed int
	o.db.QueryRow(ctx, `
		SELECT total_devices, completed_devices, failed_devices
		FROM deployments WHERE id = $1`, deploymentID,
	).Scan(&total, &completed, &failed)

	if completed+failed >= total {
		finalState := "completed"
		if failed > 0 {
			finalState = "failed"
		}
		o.db.Exec(ctx, `
			UPDATE deployments SET state = $2, completed_at = $3
			WHERE id = $1`, deploymentID, finalState, now)
	}

	return nil
}

// RollbackDeployment creates a rollback deployment for the given deployment ID.
func (o *Orchestrator) RollbackDeployment(ctx context.Context, deploymentID string) (*domain.Deployment, error) {
	// Get original deployment
	orig, err := o.GetDeployment(ctx, deploymentID)
	if err != nil {
		return nil, fmt.Errorf("get original deployment: %w", err)
	}

	// Find the previous model that was active before this deployment
	var prevModelID *string
	err = o.db.QueryRow(ctx, `
		SELECT model_id FROM deployments
		WHERE id != $1 AND state = 'completed'
		AND model_id != $2
		ORDER BY completed_at DESC LIMIT 1`,
		deploymentID, orig.ModelID,
	).Scan(&prevModelID)
	if err != nil || prevModelID == nil {
		return nil, fmt.Errorf("no previous deployment found to rollback to")
	}

	// Get previous model details
	var m domain.Model
	err = o.db.QueryRow(ctx, `
		SELECT id, name, version, format, artifact_url, checksum
		FROM models WHERE id = $1`, *prevModelID,
	).Scan(&m.ID, &m.Name, &m.Version, &m.Format, &m.ArtifactURL, &m.Checksum)
	if err != nil {
		return nil, fmt.Errorf("get rollback model: %w", err)
	}

	// Get devices that were part of the original deployment and completed
	rows, err := o.db.Query(ctx, `
		SELECT device_id FROM deployment_device_status
		WHERE deployment_id = $1 AND state = 'completed'`, deploymentID)
	if err != nil {
		return nil, fmt.Errorf("query deployment devices: %w", err)
	}
	defer rows.Close()

	var deviceIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		deviceIDs = append(deviceIDs, id)
	}

	if len(deviceIDs) == 0 {
		return nil, fmt.Errorf("no completed devices to rollback")
	}

	// Create rollback deployment
	now := time.Now()
	var d domain.Deployment
	err = o.db.QueryRow(ctx, `
		INSERT INTO deployments (model_id, target_type, state, total_devices,
			deployment_policy, rollback_model_id, started_at, created_at)
		VALUES ($1, 'rollback', 'rolling_out', $2, 'immediate', $3, $4, $4)
		RETURNING id, state, total_devices, completed_devices, failed_devices, queued_devices,
			deployment_policy, started_at, created_at`,
		m.ID, len(deviceIDs), orig.ModelID, now,
	).Scan(
		&d.ID, &d.State, &d.TotalDevices, &d.CompletedDevices,
		&d.FailedDevices, &d.QueuedDevices, &d.DeploymentPolicy,
		&d.StartedAt, &d.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create rollback deployment: %w", err)
	}

	d.ModelID = m.ID

	// Create per-device rows
	for _, devID := range deviceIDs {
		o.db.Exec(ctx, `
			INSERT INTO deployment_device_status (deployment_id, device_id, state)
			VALUES ($1, $2, 'pending')`, d.ID, devID)
	}

	// Mark original deployment as rolled_back
	o.db.Exec(ctx, `
		UPDATE deployments SET state = 'rolled_back'
		WHERE id = $1`, deploymentID)

	if o.logger != nil {
		o.logger.Infow("rollback deployment created",
			"original_deployment", deploymentID,
			"rollback_deployment", d.ID,
			"model", m.Name,
			"version", m.Version,
			"devices", len(deviceIDs),
		)
	}

	return &d, nil
}

// GetPendingCommands returns deploy commands for devices with pending deployments.
// Called during heartbeat processing to piggyback commands.
// Prefers compiled variants matching the device's runtime over the base ONNX artifact.
func (o *Orchestrator) GetPendingCommands(ctx context.Context, deviceID string) ([]map[string]interface{}, error) {
	rows, err := o.db.Query(ctx, `
		SELECT dep.id, m.id, m.name, m.version, m.format, m.checksum,
			   m.compiled_variants, dep.deployment_policy, d.runtime
		FROM deployment_device_status dds
		JOIN deployments dep ON dep.id = dds.deployment_id
		JOIN models m ON m.id = dep.model_id
		JOIN devices d ON d.id = dds.device_id
		WHERE d.device_id = $1 AND dds.state = 'pending' AND dep.state = 'rolling_out'
		ORDER BY dep.created_at ASC`, deviceID)
	if err != nil {
		return nil, fmt.Errorf("query pending commands: %w", err)
	}
	defer rows.Close()

	var commands []map[string]interface{}
	for rows.Next() {
		var deployID, modelID, modelName, modelVersion, format, checksum, policy string
		var deviceRuntime string
		var variantsJSON []byte
		if err := rows.Scan(&deployID, &modelID, &modelName, &modelVersion, &format, &checksum,
			&variantsJSON, &policy, &deviceRuntime); err != nil {
			continue
		}

		// Check if a compiled variant exists for this device's runtime
		selectedRuntime := format
		selectedChecksum := checksum
		useVariant := false

		if deviceRuntime != "" && deviceRuntime != "onnx" && variantsJSON != nil {
			var variants []domain.CompiledVariant
			if err := json.Unmarshal(variantsJSON, &variants); err == nil {
				for _, v := range variants {
					if v.Runtime == deviceRuntime {
						selectedRuntime = v.Runtime
						selectedChecksum = v.Checksum
						useVariant = true
						break
					}
				}
			}
		}

		// Generate presigned download URL
		var artifactURL string
		if useVariant {
			var err error
			artifactURL, err = o.registry.GetVariantArtifactURL(ctx, modelID, deviceRuntime)
			if err != nil {
				o.logger.Warnw("failed to generate variant presigned URL, falling back to base",
					"model_id", modelID, "runtime", deviceRuntime, "error", err)
				useVariant = false
			}
		}

		if !useVariant {
			selectedRuntime = format
			selectedChecksum = checksum
			var err error
			artifactURL, err = o.registry.GetArtifactURL(ctx, modelID)
			if err != nil {
				o.logger.Warnw("failed to generate artifact URL", "model_id", modelID, "error", err)
				continue
			}
		}

		commands = append(commands, map[string]interface{}{
			"type": "deploy_model",
			"payload": map[string]string{
				"deployment_id":     deployID,
				"model_name":        modelName,
				"model_version":     modelVersion,
				"runtime":           selectedRuntime,
				"artifact_url":      artifactURL,
				"checksum":          selectedChecksum,
				"deployment_policy": policy,
			},
		})

		// Mark as downloading to prevent duplicate commands
		o.db.Exec(ctx, `
			UPDATE deployment_device_status SET state = 'downloading'
			WHERE deployment_id = $1 AND device_id = (SELECT id FROM devices WHERE device_id = $2)`,
			deployID, deviceID)
	}

	return commands, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
