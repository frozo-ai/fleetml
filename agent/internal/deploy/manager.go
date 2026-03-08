package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fleetml/fleetml/agent/internal/communication"
	"github.com/fleetml/fleetml/agent/internal/model"
	"go.uber.org/zap"
)

// DeployPayload represents a deploy_model command payload.
type DeployPayload struct {
	DeploymentID     string `json:"deployment_id"`
	ModelName        string `json:"model_name"`
	ModelVersion     string `json:"model_version"`
	Runtime          string `json:"runtime"`
	ArtifactURL      string `json:"artifact_url"`
	Checksum         string `json:"checksum"`
	RollbackVersion  string `json:"rollback_version"`
	DeploymentPolicy string `json:"deployment_policy"`
}

// Manager handles model deployment commands from the control plane.
type Manager struct {
	deviceID     string
	loader       *model.Loader
	swapper      *model.HotSwapper
	rollback     *RollbackManager
	communicator communication.Communicator
	logger       *zap.SugaredLogger
}

func NewManager(
	deviceID string,
	loader *model.Loader,
	swapper *model.HotSwapper,
	rollback *RollbackManager,
	comm communication.Communicator,
	logger *zap.SugaredLogger,
) *Manager {
	return &Manager{
		deviceID:     deviceID,
		loader:       loader,
		swapper:      swapper,
		rollback:     rollback,
		communicator: comm,
		logger:       logger,
	}
}

// HandleCommand processes a command from the control plane.
func (m *Manager) HandleCommand(ctx context.Context, cmd communication.Command) {
	var err error
	switch cmd.Type {
	case "deploy_model":
		err = m.handleDeploy(ctx, cmd)
	case "rollback":
		err = m.handleRollback(ctx, cmd)
	default:
		m.logger.Warnw("unknown command type", "type", cmd.Type, "id", cmd.ID)
		return
	}

	if err != nil {
		m.logger.Errorw("command failed", "type", cmd.Type, "id", cmd.ID, "error", err)
	}
}

func (m *Manager) handleDeploy(ctx context.Context, cmd communication.Command) error {
	var payload DeployPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		return fmt.Errorf("parse deploy payload: %w", err)
	}

	m.logger.Infow("deploying model",
		"model", payload.ModelName,
		"version", payload.ModelVersion,
		"runtime", payload.Runtime,
		"deployment_id", payload.DeploymentID,
	)

	// Report downloading state
	m.reportStatus(ctx, payload.DeploymentID, "downloading", "")

	// Save current model info for rollback
	if current := m.swapper.Active(); current != nil {
		m.rollback.SaveVersion(current.Model.Name, current.Model.Version, current.Model.FilePath)
		m.logger.Infow("saved current model for rollback",
			"model", current.Model.Name,
			"version", current.Model.Version,
		)
	}

	// Try to load from local storage first
	modelPath, err := m.loader.Load(payload.ModelName, payload.ModelVersion)
	if err != nil && payload.ArtifactURL != "" {
		// Download from artifact URL with progress reporting
		m.logger.Infow("downloading model", "url", payload.ArtifactURL)

		modelPath, err = m.loader.DownloadFromURL(ctx, payload.ModelName, payload.ModelVersion, payload.ArtifactURL,
			func(p model.DownloadProgress) {
				if p.TotalBytes > 0 {
					pct := float64(p.BytesRead) / float64(p.TotalBytes) * 100
					m.logger.Debugw("download progress",
						"model", p.ModelName,
						"percent", fmt.Sprintf("%.1f%%", pct),
						"bytes", p.BytesRead,
						"total", p.TotalBytes,
					)
				}
			})
		if err != nil {
			m.reportStatus(ctx, payload.DeploymentID, "failed", err.Error())
			return fmt.Errorf("download model: %w", err)
		}
	} else if err != nil {
		m.reportStatus(ctx, payload.DeploymentID, "failed", "model not found and no artifact URL provided")
		return fmt.Errorf("model not available: %w", err)
	}

	// Validate checksum
	if payload.Checksum != "" {
		m.reportStatus(ctx, payload.DeploymentID, "verifying", "")
		if err := m.loader.ValidateChecksum(modelPath, payload.Checksum); err != nil {
			m.reportStatus(ctx, payload.DeploymentID, "failed", "checksum validation failed")
			return fmt.Errorf("checksum validation failed: %w", err)
		}
	}

	// Report loading state
	m.reportStatus(ctx, payload.DeploymentID, "loading", "")

	// Use background swap for zero-downtime deployment
	doneCh := make(chan error, 1)

	loadFn := func() (*model.LoadedModel, error) {
		runtime := model.NewONNXSubprocessRuntime()
		if err := runtime.Load(modelPath); err != nil {
			return nil, fmt.Errorf("load model into runtime: %w", err)
		}

		return &model.LoadedModel{
			Model: &model.Model{
				Name:     payload.ModelName,
				Version:  payload.ModelVersion,
				Format:   payload.Runtime,
				FilePath: modelPath,
				Checksum: payload.Checksum,
				LoadedAt: time.Now(),
			},
			Runtime: runtime,
		}, nil
	}

	// Verify function: run a test inference to validate the model is operable.
	// Sends an empty JSON object — onnx_infer generates zero-filled dummy
	// tensors for any inputs not provided, so this validates that the model
	// loads and executes without requiring knowledge of the model's schema.
	verifyFn := func(lm *model.LoadedModel) error {
		_, err := lm.Runtime.Infer([]byte("{}"))
		if err != nil {
			// If the inference helper binary is simply not installed, skip
			// verification rather than failing the deployment — the model
			// file is still valid (passed Load). This allows deployments on
			// devices that use external inference servers or custom runtimes.
			if strings.Contains(err.Error(), "helper not found") {
				m.logger.Warnw("onnx_infer helper not found, skipping verification inference",
					"model", payload.ModelName,
					"version", payload.ModelVersion,
				)
				return nil
			}
			return fmt.Errorf("test inference failed: %w", err)
		}
		return nil
	}

	m.swapper.BackgroundSwap(
		fmt.Sprintf("%s:%s", payload.ModelName, payload.ModelVersion),
		loadFn,
		verifyFn,
		doneCh,
	)

	// Wait for the swap to complete
	if err := <-doneCh; err != nil {
		m.reportStatus(ctx, payload.DeploymentID, "failed", err.Error())
		return err
	}

	m.reportStatus(ctx, payload.DeploymentID, "completed", "")
	m.logger.Infow("model deployed successfully",
		"model", payload.ModelName,
		"version", payload.ModelVersion,
		"deployment_id", payload.DeploymentID,
	)

	return nil
}

func (m *Manager) handleRollback(ctx context.Context, cmd communication.Command) error {
	var payload struct {
		DeploymentID string `json:"deployment_id"`
		ToVersion    string `json:"to_version"`
		ModelName    string `json:"model_name"`
	}
	json.Unmarshal(cmd.Payload, &payload)

	// Try specific version rollback first
	if payload.ToVersion != "" && payload.ModelName != "" {
		modelPath, err := m.rollback.Restore(payload.ModelName, payload.ToVersion)
		if err == nil {
			runtime := model.NewONNXSubprocessRuntime()
			if err := runtime.Load(modelPath); err != nil {
				m.reportStatus(ctx, payload.DeploymentID, "failed", err.Error())
				return fmt.Errorf("load rollback model: %w", err)
			}

			newModel := &model.LoadedModel{
				Model: &model.Model{
					Name:     payload.ModelName,
					Version:  payload.ToVersion,
					FilePath: modelPath,
					LoadedAt: time.Now(),
				},
				Runtime: runtime,
			}

			if err := m.swapper.Swap(newModel); err != nil {
				m.reportStatus(ctx, payload.DeploymentID, "failed", err.Error())
				return fmt.Errorf("swap failed: %w", err)
			}

			m.reportStatus(ctx, payload.DeploymentID, "completed", "")
			m.logger.Infow("rollback to specific version successful",
				"model", payload.ModelName, "version", payload.ToVersion)
			return nil
		}
	}

	// Fall back to atomic pointer rollback
	if err := m.swapper.Rollback(); err != nil {
		m.reportStatus(ctx, payload.DeploymentID, "failed", err.Error())
		return fmt.Errorf("rollback failed: %w", err)
	}

	active := m.swapper.Active()
	if active != nil {
		m.logger.Infow("rollback successful",
			"model", active.Model.Name,
			"version", active.Model.Version,
		)
	}
	m.reportStatus(ctx, payload.DeploymentID, "completed", "")

	return nil
}

func (m *Manager) reportStatus(ctx context.Context, deploymentID, state, errMsg string) {
	if m.communicator == nil || deploymentID == "" {
		return
	}
	if err := m.communicator.ReportDeploymentStatus(ctx, m.deviceID, deploymentID, state, errMsg); err != nil {
		m.logger.Warnw("failed to report deployment status",
			"deployment_id", deploymentID,
			"state", state,
			"error", err,
		)
	}
}
