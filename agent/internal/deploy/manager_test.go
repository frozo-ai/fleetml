package deploy

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/fleetml/fleetml/agent/internal/communication"
	"github.com/fleetml/fleetml/agent/internal/device"
	"github.com/fleetml/fleetml/agent/internal/health"
	"github.com/fleetml/fleetml/agent/internal/model"
	"go.uber.org/zap"
)

// mockCommunicator records deployment status reports.
type mockCommunicator struct {
	reports []statusReport
}

type statusReport struct {
	DeviceID     string
	DeploymentID string
	State        string
	ErrMsg       string
}

func (m *mockCommunicator) Register(ctx context.Context, info *device.Info) (string, int, error) {
	return "agent-1", 30, nil
}

func (m *mockCommunicator) SendHeartbeat(ctx context.Context, deviceID, status string, system *health.SystemMetrics) ([]communication.Command, error) {
	return nil, nil
}

func (m *mockCommunicator) ReportDeploymentStatus(ctx context.Context, deviceID, deploymentID, state, errMsg string) error {
	m.reports = append(m.reports, statusReport{
		DeviceID:     deviceID,
		DeploymentID: deploymentID,
		State:        state,
		ErrMsg:       errMsg,
	})
	return nil
}

func (m *mockCommunicator) SendLogs(ctx context.Context, deviceID string, entries []communication.LogEntry) error {
	return nil
}

func (m *mockCommunicator) Close() error { return nil }

func newTestManager(t *testing.T) (*Manager, *mockCommunicator) {
	t.Helper()
	dir := t.TempDir()

	loader := model.NewLoader(dir, 5)
	swapper := model.NewHotSwapper()
	rollback := NewRollbackManager(filepath.Join(dir, "rollback"), 3)

	logger, _ := zap.NewDevelopment()
	comm := &mockCommunicator{}

	mgr := NewManager("device-1", loader, swapper, rollback, comm, logger.Sugar())
	return mgr, comm
}

func TestHandleCommand_DeployModel(t *testing.T) {
	mgr, comm := newTestManager(t)

	// Create a model file in the loader's storage dir
	modelDir := filepath.Join(mgr.loader.StorageDir(), "test-model")
	os.MkdirAll(modelDir, 0o755)
	modelPath := filepath.Join(modelDir, "v1.0.onnx")
	os.WriteFile(modelPath, []byte("model-data"), 0o644)

	payload := DeployPayload{
		DeploymentID: "deploy-1",
		ModelName:    "test-model",
		ModelVersion: "v1.0",
		Runtime:      "onnx",
	}
	payloadBytes, _ := json.Marshal(payload)

	cmd := communication.Command{
		ID:      "cmd-1",
		Type:    "deploy_model",
		Payload: payloadBytes,
	}

	mgr.HandleCommand(context.Background(), cmd)

	// Check that status was reported
	hasCompleted := false
	for _, r := range comm.reports {
		if r.DeploymentID == "deploy-1" && r.State == "completed" {
			hasCompleted = true
		}
	}
	if !hasCompleted {
		t.Error("expected completed status report")
	}

	// Check model is active
	active := mgr.swapper.Active()
	if active == nil {
		t.Fatal("expected active model after deploy")
	}
	if active.Model.Name != "test-model" {
		t.Errorf("expected model name 'test-model', got %q", active.Model.Name)
	}
}

func TestHandleCommand_Rollback(t *testing.T) {
	mgr, comm := newTestManager(t)

	// Set up an initial model
	initial := &model.LoadedModel{
		Model:   &model.Model{Name: "m1", Version: "1.0"},
		Runtime: model.NewONNXRuntime(),
	}
	mgr.swapper.Swap(initial)

	// Deploy v2.0
	modelDir := filepath.Join(mgr.loader.StorageDir(), "m1")
	os.MkdirAll(modelDir, 0o755)
	os.WriteFile(filepath.Join(modelDir, "v2.0.onnx"), []byte("v2-data"), 0o644)

	deployPayload := DeployPayload{
		DeploymentID: "deploy-2",
		ModelName:    "m1",
		ModelVersion: "v2.0",
		Runtime:      "onnx",
	}
	dpBytes, _ := json.Marshal(deployPayload)
	mgr.HandleCommand(context.Background(), communication.Command{
		ID:      "cmd-2",
		Type:    "deploy_model",
		Payload: dpBytes,
	})

	// Now rollback
	rollbackPayload, _ := json.Marshal(map[string]string{
		"deployment_id": "deploy-rollback",
	})
	mgr.HandleCommand(context.Background(), communication.Command{
		ID:      "cmd-3",
		Type:    "rollback",
		Payload: rollbackPayload,
	})

	// Check that rollback completed
	hasCompleted := false
	for _, r := range comm.reports {
		if r.DeploymentID == "deploy-rollback" && r.State == "completed" {
			hasCompleted = true
		}
	}
	if !hasCompleted {
		t.Error("expected completed status report for rollback")
	}
}

func TestHandleCommand_UnknownType(t *testing.T) {
	mgr, _ := newTestManager(t)

	cmd := communication.Command{
		ID:      "cmd-1",
		Type:    "unknown_command",
		Payload: []byte("{}"),
	}

	// Should not panic
	mgr.HandleCommand(context.Background(), cmd)
}

func TestHandleCommand_InvalidPayload(t *testing.T) {
	mgr, _ := newTestManager(t)

	cmd := communication.Command{
		ID:      "cmd-1",
		Type:    "deploy_model",
		Payload: []byte("invalid-json"),
	}

	// Should not crash, error is logged
	mgr.HandleCommand(context.Background(), cmd)
}

func TestHandleCommand_ModelNotAvailable(t *testing.T) {
	mgr, comm := newTestManager(t)

	payload := DeployPayload{
		DeploymentID: "deploy-1",
		ModelName:    "nonexistent",
		ModelVersion: "v1.0",
		Runtime:      "onnx",
	}
	payloadBytes, _ := json.Marshal(payload)

	cmd := communication.Command{
		ID:      "cmd-1",
		Type:    "deploy_model",
		Payload: payloadBytes,
	}

	mgr.HandleCommand(context.Background(), cmd)

	// Should report failure
	hasFailed := false
	for _, r := range comm.reports {
		if r.DeploymentID == "deploy-1" && r.State == "failed" {
			hasFailed = true
		}
	}
	if !hasFailed {
		t.Error("expected failed status report when model not available")
	}
}

func TestHandleCommand_EmptyPayload(t *testing.T) {
	mgr, _ := newTestManager(t)
	cmd := communication.Command{
		ID:      "cmd-empty",
		Type:    "deploy_model",
		Payload: []byte(""),
	}
	// Should not panic — empty payload fails JSON unmarshal
	mgr.HandleCommand(context.Background(), cmd)
}

func TestHandleCommand_NilPayload(t *testing.T) {
	mgr, _ := newTestManager(t)
	cmd := communication.Command{
		ID:      "cmd-nil",
		Type:    "deploy_model",
		Payload: nil,
	}
	// Should not panic — nil payload fails JSON unmarshal
	mgr.HandleCommand(context.Background(), cmd)
}

func TestHandleCommand_RollbackWithEmptyPayload(t *testing.T) {
	mgr, _ := newTestManager(t)
	cmd := communication.Command{
		ID:      "cmd-rb-empty",
		Type:    "rollback",
		Payload: []byte(""),
	}
	mgr.HandleCommand(context.Background(), cmd)
}

func TestHandleCommand_EmptyCommandType(t *testing.T) {
	mgr, _ := newTestManager(t)
	cmd := communication.Command{
		ID:      "cmd-1",
		Type:    "",
		Payload: []byte("{}"),
	}
	// Empty type hits default case — should not panic
	mgr.HandleCommand(context.Background(), cmd)
}
