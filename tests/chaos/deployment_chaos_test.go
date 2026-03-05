//go:build chaos

package chaos

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/fleetml/fleetml/simulator"
	"go.uber.org/zap"
)

// TestDeployDuringPartition validates deploying while part of fleet is offline.
func TestDeployDuringPartition(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddHeterogeneousFleet(40)

	devices := fleet.ListDevices()

	// Partition: take 10 offline
	offlineIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		offlineIDs[i] = devices[i].ID
	}
	fleet.TakeOffline(offlineIDs)

	// Deploy to entire fleet
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
	if onlineDeployed != 30 {
		t.Errorf("expected 30 online devices deployed, got %d", onlineDeployed)
	}

	// Heal partition
	fleet.BringOnline(offlineIDs, simulator.PredefinedNetworkProfiles["good"])
	fleet.ApplyPendingDeployments()

	// All should now have the model
	deployed := fleet.CountDeployedModel("model-v1")
	if deployed != 40 {
		t.Errorf("expected 40 deployed after heal, got %d", deployed)
	}
}

// TestRapidDeployRollback validates rapid deploy-rollback cycles.
func TestRapidDeployRollback(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(20, simulator.PredefinedProfiles["intel-nuc"], simulator.PredefinedNetworkProfiles["excellent"])

	devices := fleet.ListDevices()
	allIDs := make([]string, len(devices))
	for i, d := range devices {
		allIDs[i] = d.ID
	}

	// Rapid cycles: deploy v1 → v2 → v1 → v3 → v1
	versions := []string{"v1", "v2", "v1", "v3", "v1"}
	for _, ver := range versions {
		fleet.DeployModel(allIDs, fmt.Sprintf("model-%s", ver))
		fleet.SimulateMetrics()
	}

	// Should end on v1
	deployed := fleet.CountDeployedModel("model-v1")
	if deployed != 20 {
		t.Errorf("expected 20 on model-v1 after cycles, got %d", deployed)
	}
}

// TestDeployWithDeviceChurn validates deployment while devices join and leave.
func TestDeployWithDeviceChurn(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(30, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["good"])

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Deploy model
	devices := fleet.ListDevices()
	allIDs := make([]string, len(devices))
	for i, d := range devices {
		allIDs[i] = d.ID
	}
	fleet.DeployModel(allIDs, "model-v1")

	// Simulate churn: random devices go offline and online
	for round := 0; round < 50; round++ {
		fleet.SimulateMetrics()

		// 15% chance of random device going offline
		if rng.Float64() < 0.15 {
			onlineDevices := fleet.OnlineDevices()
			if len(onlineDevices) > 5 {
				idx := rng.Intn(len(onlineDevices))
				fleet.TakeOffline([]string{onlineDevices[idx].ID})
			}
		}

		// 25% chance of bringing a random offline device back
		if rng.Float64() < 0.25 {
			offlineDevices := fleet.OfflineDevices()
			if len(offlineDevices) > 0 {
				idx := rng.Intn(len(offlineDevices))
				fleet.BringOnline([]string{offlineDevices[idx].ID}, simulator.PredefinedNetworkProfiles["good"])
			}
		}
	}

	// Bring all back online
	offlineDevices := fleet.OfflineDevices()
	offIDs := make([]string, len(offlineDevices))
	for i, d := range offlineDevices {
		offIDs[i] = d.ID
	}
	fleet.BringOnline(offIDs, simulator.PredefinedNetworkProfiles["good"])

	// All should be online now
	for _, d := range fleet.ListDevices() {
		if d.Status == "offline" {
			t.Errorf("device %s still offline after full recovery", d.ID)
		}
	}
}

// TestSplitBrainScenario simulates a split-brain where two halves
// of the fleet receive different model versions.
func TestSplitBrainScenario(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(20, simulator.PredefinedProfiles["intel-nuc"], simulator.PredefinedNetworkProfiles["good"])

	devices := fleet.ListDevices()
	half1 := make([]string, 10)
	half2 := make([]string, 10)
	for i := 0; i < 10; i++ {
		half1[i] = devices[i].ID
		half2[i] = devices[i+10].ID
	}

	// Deploy different versions to each half
	fleet.DeployModel(half1, "model-v1")
	fleet.DeployModel(half2, "model-v2")

	// Simulate operation
	for i := 0; i < 20; i++ {
		fleet.SimulateMetrics()
	}

	// Verify split
	v1Count := fleet.CountDeployedModel("model-v1")
	v2Count := fleet.CountDeployedModel("model-v2")

	if v1Count != 10 {
		t.Errorf("expected 10 on v1, got %d", v1Count)
	}
	if v2Count != 10 {
		t.Errorf("expected 10 on v2, got %d", v2Count)
	}

	// Resolve split brain: deploy v2 to all
	allIDs := append(half1, half2...)
	fleet.DeployModel(allIDs, "model-v2")

	v2Count = fleet.CountDeployedModel("model-v2")
	if v2Count != 20 {
		t.Errorf("expected 20 on v2 after resolution, got %d", v2Count)
	}
}

// TestFleetWideRestart simulates all devices restarting simultaneously.
func TestFleetWideRestart(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddHeterogeneousFleet(50)

	devices := fleet.ListDevices()
	allIDs := make([]string, len(devices))
	for i, d := range devices {
		allIDs[i] = d.ID
	}

	// Deploy model
	fleet.DeployModel(allIDs, "model-v1")

	// Simulate fleet-wide restart
	fleet.TakeOffline(allIDs)

	offlineCount := 0
	for _, d := range fleet.ListDevices() {
		if d.Status == "offline" {
			offlineCount++
		}
	}
	if offlineCount != 50 {
		t.Errorf("expected all 50 offline, got %d", offlineCount)
	}

	// All come back
	fleet.BringOnline(allIDs, simulator.PredefinedNetworkProfiles["good"])

	for _, d := range fleet.ListDevices() {
		if d.Status == "offline" {
			t.Errorf("device %s still offline after restart", d.ID)
		}
	}

	// Model should persist after restart
	deployed := fleet.CountDeployedModel("model-v1")
	if deployed != 50 {
		t.Errorf("model should persist after restart: expected 50, got %d", deployed)
	}
}
