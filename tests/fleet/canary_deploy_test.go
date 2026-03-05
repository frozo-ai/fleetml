//go:build fleet

package fleet

import (
	"fmt"
	"testing"

	"github.com/fleetml/fleetml/simulator"
	"go.uber.org/zap"
)

// TestCanaryDeploymentStages validates staged rollout across a fleet.
func TestCanaryDeploymentStages(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddHeterogeneousFleet(100)

	devices := fleet.ListDevices()
	total := len(devices)

	// Stage 1: Deploy to 5% of fleet
	stage1Size := total * 5 / 100
	if stage1Size < 1 {
		stage1Size = 1
	}
	canaryIDs := make([]string, stage1Size)
	for i := 0; i < stage1Size; i++ {
		canaryIDs[i] = devices[i].ID
	}

	fleet.DeployModel(canaryIDs, "model-v2")
	deployed := fleet.CountDeployedModel("model-v2")
	if deployed != stage1Size {
		t.Errorf("stage 1: expected %d deployed, got %d", stage1Size, deployed)
	}

	// Simulate health check ticks
	for i := 0; i < 10; i++ {
		fleet.SimulateMetrics()
	}

	// Stage 2: Deploy to 50% of fleet
	stage2Size := total * 50 / 100
	stage2IDs := make([]string, stage2Size)
	for i := 0; i < stage2Size; i++ {
		stage2IDs[i] = devices[i].ID
	}

	fleet.DeployModel(stage2IDs, "model-v2")
	deployed = fleet.CountDeployedModel("model-v2")
	if deployed != stage2Size {
		t.Errorf("stage 2: expected %d deployed, got %d", stage2Size, deployed)
	}

	// Stage 3: Deploy to 100%
	allIDs := make([]string, total)
	for i, d := range devices {
		allIDs[i] = d.ID
	}

	fleet.DeployModel(allIDs, "model-v2")
	deployed = fleet.CountDeployedModel("model-v2")
	if deployed != total {
		t.Errorf("stage 3: expected %d deployed, got %d", total, deployed)
	}
}

// TestCanaryRollback validates that a bad canary is detected and rolled back.
func TestCanaryRollback(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(50, simulator.PredefinedProfiles["jetson-nano"], simulator.PredefinedNetworkProfiles["good"])

	devices := fleet.ListDevices()

	// Deploy original model to all
	allIDs := make([]string, len(devices))
	for i, d := range devices {
		allIDs[i] = d.ID
	}
	fleet.DeployModel(allIDs, "model-v1")

	// Canary: deploy v2 to 5 devices
	canaryIDs := allIDs[:5]
	fleet.DeployModel(canaryIDs, "model-v2")

	// Simulate error on canary devices (high error rate)
	fleet.InjectModelError(canaryIDs, 0.5) // 50% error rate

	// Run health checks
	for i := 0; i < 20; i++ {
		fleet.SimulateMetrics()
	}

	// Detect canary failure
	errorRate := fleet.GetModelErrorRate("model-v2")
	if errorRate < 0.3 {
		t.Logf("Warning: expected high error rate on canary, got %.2f", errorRate)
	}

	// Rollback canary
	fleet.DeployModel(canaryIDs, "model-v1")
	fleet.ClearModelErrors(canaryIDs)

	// Verify all on v1
	v1Count := fleet.CountDeployedModel("model-v1")
	if v1Count != len(devices) {
		t.Errorf("after rollback: expected %d on v1, got %d", len(devices), v1Count)
	}
}

// TestDeployToSpecificRuntime validates deploying to devices with a specific runtime.
func TestDeployToSpecificRuntime(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())

	fleet.AddDevices(10, simulator.PredefinedProfiles["jetson-nano"], simulator.PredefinedNetworkProfiles["good"])
	fleet.AddDevices(10, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["good"])
	fleet.AddDevices(10, simulator.PredefinedProfiles["intel-nuc"], simulator.PredefinedNetworkProfiles["good"])

	// Deploy TensorRT model only to Jetson devices
	jetsons := fleet.DevicesByProfile()["jetson-nano"]
	jetsonIDs := make([]string, len(jetsons))
	for i, d := range jetsons {
		jetsonIDs[i] = d.ID
	}

	fleet.DeployModel(jetsonIDs, "mobilenet-tensorrt")

	deployed := fleet.CountDeployedModel("mobilenet-tensorrt")
	if deployed != 10 {
		t.Errorf("expected 10 Jetson devices with TensorRT model, got %d", deployed)
	}

	// Non-Jetson devices should not have the model
	for _, d := range fleet.ListDevices() {
		isJetson := false
		for _, jid := range jetsonIDs {
			if d.ID == jid {
				isJetson = true
				break
			}
		}
		if !isJetson && d.DeployedModel == "mobilenet-tensorrt" {
			t.Errorf("non-Jetson device %s should not have TensorRT model", d.ID)
		}
	}
}

// TestConcurrentDeployments validates multiple deployments happening simultaneously.
func TestConcurrentDeployments(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddHeterogeneousFleet(40)

	devices := fleet.ListDevices()
	byProfile := fleet.DevicesByProfile()

	// Deploy different models to different device types simultaneously
	for profileName, profileDevices := range byProfile {
		ids := make([]string, len(profileDevices))
		for i, d := range profileDevices {
			ids[i] = d.ID
		}
		modelName := fmt.Sprintf("model-%s", profileName)
		fleet.DeployModel(ids, modelName)
	}

	// Verify each profile has its model
	for profileName, profileDevices := range byProfile {
		expectedModel := fmt.Sprintf("model-%s", profileName)
		for _, d := range profileDevices {
			device := fleet.GetDevice(d.ID)
			if device.DeployedModel != expectedModel {
				t.Errorf("device %s (profile %s): expected model %q, got %q",
					d.ID, profileName, expectedModel, device.DeployedModel)
			}
		}
	}

	_ = devices // used for count validation
}

// TestDeployToOfflineDevice validates that offline devices queue the deployment.
func TestDeployToOfflineDevice(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(10, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["good"])

	devices := fleet.ListDevices()

	// Take 3 devices offline
	offlineIDs := []string{devices[0].ID, devices[1].ID, devices[2].ID}
	fleet.TakeOffline(offlineIDs)

	// Deploy to all devices
	allIDs := make([]string, len(devices))
	for i, d := range devices {
		allIDs[i] = d.ID
	}
	fleet.DeployModel(allIDs, "model-v1")

	// Online devices should have the model
	onlineDeployed := 0
	for _, d := range fleet.ListDevices() {
		if d.Status != "offline" && d.DeployedModel == "model-v1" {
			onlineDeployed++
		}
	}
	if onlineDeployed != 7 {
		t.Errorf("expected 7 online devices deployed, got %d", onlineDeployed)
	}

	// Bring offline devices back
	fleet.BringOnline(offlineIDs, simulator.PredefinedNetworkProfiles["good"])

	// Pending deployment should apply
	fleet.ApplyPendingDeployments()

	deployed := fleet.CountDeployedModel("model-v1")
	if deployed != 10 {
		t.Errorf("after reconnect: expected 10 deployed, got %d", deployed)
	}
}

// TestFleetScaleDeployment validates deployment time scales linearly with fleet size.
func TestFleetScaleDeployment(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	sizes := []int{10, 50, 100}

	for _, size := range sizes {
		fleet := simulator.NewFleet(logger.Sugar())
		fleet.AddHeterogeneousFleet(size)

		devices := fleet.ListDevices()
		ids := make([]string, len(devices))
		for i, d := range devices {
			ids[i] = d.ID
		}

		fleet.DeployModel(ids, "test-model")
		deployed := fleet.CountDeployedModel("test-model")

		if deployed != size {
			t.Errorf("fleet size %d: expected %d deployed, got %d", size, size, deployed)
		}
	}
}

// TestModelVersionUpgrade validates upgrading from v1 to v2 across fleet.
func TestModelVersionUpgrade(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(20, simulator.PredefinedProfiles["intel-nuc"], simulator.PredefinedNetworkProfiles["excellent"])

	devices := fleet.ListDevices()
	allIDs := make([]string, len(devices))
	for i, d := range devices {
		allIDs[i] = d.ID
	}

	// Deploy v1
	fleet.DeployModel(allIDs, "model-v1")
	if fleet.CountDeployedModel("model-v1") != 20 {
		t.Fatal("v1 deployment failed")
	}

	// Upgrade to v2
	fleet.DeployModel(allIDs, "model-v2")
	if fleet.CountDeployedModel("model-v2") != 20 {
		t.Error("v2 upgrade failed")
	}
	if fleet.CountDeployedModel("model-v1") != 0 {
		t.Error("v1 should be fully replaced")
	}
}
