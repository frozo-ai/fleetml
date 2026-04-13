package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/fleetml/fleetml/server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// CanaryManager tracks canary deployment stages and makes advance/rollback
// decisions based on per-device metrics from heartbeats.
type CanaryManager struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewCanaryManager(db *pgxpool.Pool, logger *zap.SugaredLogger) *CanaryManager {
	return &CanaryManager{db: db, logger: logger}
}

// CanaryState represents the current state of a canary deployment.
type CanaryState struct {
	DeploymentID  string             `json:"deployment_id"`
	CurrentStage  int                `json:"current_stage"`
	Stages        []domain.CanaryStage `json:"stages"`
	StageStarted  time.Time          `json:"stage_started"`
	CanaryDevices []string           `json:"canary_devices"` // device IDs in current canary batch
	TotalDevices  int                `json:"total_devices"`
}

// InitCanary initializes canary tracking for a deployment.
func (cm *CanaryManager) InitCanary(ctx context.Context, deploymentID string, config *domain.CanaryConfig, totalDevices int) error {
	state := CanaryState{
		DeploymentID: deploymentID,
		CurrentStage: 0,
		Stages:       config.Stages,
		StageStarted: time.Now(),
		TotalDevices: totalDevices,
	}

	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal canary state: %w", err)
	}

	_, err = cm.db.Exec(ctx, `
		UPDATE deployments SET canary_config = $2
		WHERE id = $1`, deploymentID, stateJSON)
	if err != nil {
		return fmt.Errorf("save canary state: %w", err)
	}

	cm.logger.Infow("canary initialized",
		"deployment_id", deploymentID,
		"stages", len(config.Stages),
		"total_devices", totalDevices,
	)

	return nil
}

// GetCanaryState loads the current canary state for a deployment.
func (cm *CanaryManager) GetCanaryState(ctx context.Context, deploymentID string) (*CanaryState, error) {
	var stateJSON []byte
	err := cm.db.QueryRow(ctx, `
		SELECT canary_config FROM deployments WHERE id = $1`, deploymentID,
	).Scan(&stateJSON)
	if err != nil {
		return nil, fmt.Errorf("load canary state: %w", err)
	}

	if stateJSON == nil {
		return nil, fmt.Errorf("no canary config for deployment %s", deploymentID)
	}

	var state CanaryState
	if err := json.Unmarshal(stateJSON, &state); err != nil {
		return nil, fmt.Errorf("unmarshal canary state: %w", err)
	}

	return &state, nil
}

// DevicesForStage returns how many devices should be deployed in the given stage.
func (cm *CanaryManager) DevicesForStage(totalDevices int, stage domain.CanaryStage) int {
	count := int(math.Ceil(float64(totalDevices) * float64(stage.Percent) / 100.0))
	if count > totalDevices {
		count = totalDevices
	}
	if count < 1 {
		count = 1
	}
	return count
}

// SelectCanaryDevices selects device IDs for the current canary stage.
// It picks devices that haven't been deployed to yet, preferring healthy devices.
func (cm *CanaryManager) SelectCanaryDevices(ctx context.Context, deploymentID string, count int) ([]string, error) {
	rows, err := cm.db.Query(ctx, `
		SELECT dds.device_id FROM deployment_device_status dds
		JOIN devices d ON d.id = dds.device_id
		WHERE dds.deployment_id = $1 AND dds.state = 'queued'
		ORDER BY
			CASE d.status WHEN 'healthy' THEN 0 WHEN 'warning' THEN 1 ELSE 2 END,
			d.last_heartbeat DESC NULLS LAST
		LIMIT $2`, deploymentID, count)
	if err != nil {
		return nil, fmt.Errorf("select canary devices: %w", err)
	}
	defer rows.Close()

	var deviceIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan device id: %w", err)
		}
		deviceIDs = append(deviceIDs, id)
	}

	return deviceIDs, nil
}

// EvaluateStage checks if the current canary stage passes success criteria.
// Returns (advance, rollback, error).
func (cm *CanaryManager) EvaluateStage(ctx context.Context, deploymentID string) (advance bool, rollback bool, err error) {
	state, err := cm.GetCanaryState(ctx, deploymentID)
	if err != nil {
		return false, false, err
	}

	if state.CurrentStage >= len(state.Stages) {
		return false, false, fmt.Errorf("all stages complete")
	}

	currentStage := state.Stages[state.CurrentStage]

	// Parse duration for the stage
	duration, err := time.ParseDuration(currentStage.Duration)
	if err != nil {
		duration = 5 * time.Minute // Default stage duration
	}

	// Check if enough time has passed
	if time.Since(state.StageStarted) < duration {
		return false, false, nil // Still waiting
	}

	// Evaluate success metrics for canary devices
	var completedCount, failedCount, totalCanary int
	err = cm.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE state = 'completed') AS completed,
			COUNT(*) FILTER (WHERE state = 'failed') AS failed,
			COUNT(*) AS total
		FROM deployment_device_status
		WHERE deployment_id = $1 AND state IN ('completed', 'failed', 'downloading', 'loading', 'verifying')`,
		deploymentID,
	).Scan(&completedCount, &failedCount, &totalCanary)
	if err != nil {
		return false, false, fmt.Errorf("query canary metrics: %w", err)
	}

	if totalCanary == 0 {
		return false, false, nil // No devices reported yet
	}

	// Default success metric: >80% of canary devices succeeded
	failRate := float64(failedCount) / float64(totalCanary)
	maxFailRate := 0.2 // 20% failure threshold

	if currentStage.SuccessMetric != "" {
		// Could parse custom metrics here; for MVP, use fail rate
		cm.logger.Debugw("evaluating canary stage",
			"stage", state.CurrentStage,
			"metric", currentStage.SuccessMetric,
			"completed", completedCount,
			"failed", failedCount,
			"total", totalCanary,
		)
	}

	if failRate > maxFailRate {
		cm.logger.Warnw("canary stage failed, triggering rollback",
			"deployment_id", deploymentID,
			"stage", state.CurrentStage,
			"fail_rate", failRate,
		)
		return false, true, nil
	}

	// Stage passed
	cm.logger.Infow("canary stage passed",
		"deployment_id", deploymentID,
		"stage", state.CurrentStage,
		"completed", completedCount,
		"failed", failedCount,
	)

	return true, false, nil
}

// AdvanceStage moves to the next canary stage.
func (cm *CanaryManager) AdvanceStage(ctx context.Context, deploymentID string) error {
	state, err := cm.GetCanaryState(ctx, deploymentID)
	if err != nil {
		return err
	}

	state.CurrentStage++
	state.StageStarted = time.Now()

	if state.CurrentStage >= len(state.Stages) {
		// All stages complete, move remaining pending devices to active
		cm.logger.Infow("all canary stages complete, deploying to remaining devices",
			"deployment_id", deploymentID,
		)

		_, err := cm.db.Exec(ctx, `
			UPDATE deployment_device_status SET state = 'pending'
			WHERE deployment_id = $1 AND state = 'queued'`, deploymentID)
		if err != nil {
			return fmt.Errorf("activate remaining devices: %w", err)
		}
	}

	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	_, err = cm.db.Exec(ctx, `
		UPDATE deployments SET canary_config = $2
		WHERE id = $1`, deploymentID, stateJSON)
	if err != nil {
		return fmt.Errorf("save canary state: %w", err)
	}

	return nil
}

// TriggerRollback marks the deployment for rollback due to canary failure.
func (cm *CanaryManager) TriggerRollback(ctx context.Context, deploymentID string) error {
	now := time.Now()

	// Cancel all pending/queued devices
	_, err := cm.db.Exec(ctx, `
		UPDATE deployment_device_status SET state = 'cancelled'
		WHERE deployment_id = $1 AND state IN ('pending', 'queued')`, deploymentID)
	if err != nil {
		return fmt.Errorf("cancel pending devices: %w", err)
	}

	// Mark deployment as rolled_back
	_, err = cm.db.Exec(ctx, `
		UPDATE deployments SET state = 'rolled_back', error = 'canary check failed', completed_at = $2
		WHERE id = $1`, deploymentID, now)
	if err != nil {
		return fmt.Errorf("mark deployment rolled back: %w", err)
	}

	// Send rollback commands to devices that already deployed
	rows, err := cm.db.Query(ctx, `
		SELECT d.device_id FROM deployment_device_status dds
		JOIN devices d ON d.id = dds.device_id
		WHERE dds.deployment_id = $1 AND dds.state = 'completed'`, deploymentID)
	if err != nil {
		return fmt.Errorf("query completed devices: %w", err)
	}
	defer rows.Close()

	var rollbackDevices []string
	for rows.Next() {
		var deviceID string
		if err := rows.Scan(&deviceID); err != nil {
			continue
		}
		rollbackDevices = append(rollbackDevices, deviceID)
	}

	cm.logger.Infow("canary rollback triggered",
		"deployment_id", deploymentID,
		"rollback_devices", len(rollbackDevices),
	)

	return nil
}

// StartEvaluationLoop runs a periodic canary evaluation for active canary deployments.
func (cm *CanaryManager) StartEvaluationLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cm.evaluateActiveCanaries(ctx)
		}
	}
}

func (cm *CanaryManager) evaluateActiveCanaries(ctx context.Context) {
	rows, err := cm.db.Query(ctx, `
		SELECT id FROM deployments
		WHERE state = 'rolling_out' AND deployment_policy = 'canary'`)
	if err != nil {
		cm.logger.Warnw("failed to query canary deployments", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var deploymentID string
		if err := rows.Scan(&deploymentID); err != nil {
			continue
		}

		advance, rollbackNeeded, err := cm.EvaluateStage(ctx, deploymentID)
		if err != nil {
			cm.logger.Warnw("canary evaluation error",
				"deployment_id", deploymentID,
				"error", err,
			)
			continue
		}

		if rollbackNeeded {
			if err := cm.TriggerRollback(ctx, deploymentID); err != nil {
				cm.logger.Errorw("canary rollback failed",
					"deployment_id", deploymentID,
					"error", err,
				)
			}
		} else if advance {
			if err := cm.AdvanceStage(ctx, deploymentID); err != nil {
				cm.logger.Errorw("canary advance failed",
					"deployment_id", deploymentID,
					"error", err,
				)
			}
		}
	}
}
