//go:build chaos

package chaos

import (
	"testing"

	"github.com/fleetml/fleetml/simulator"
	"go.uber.org/zap"
)

// TestDiskFullScenario simulates devices running out of disk space.
func TestDiskFullScenario(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(20, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["good"])

	devices := fleet.ListDevices()
	targetIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		targetIDs[i] = devices[i].ID
	}

	// Inject disk full condition
	fleet.InjectDiskFull(targetIDs)

	for i := 0; i < 50; i++ {
		fleet.SimulateMetrics()
	}

	// Verify disk-full devices are in warning/error state
	for _, id := range targetIDs {
		d := fleet.GetDevice(id)
		if d.DiskPercent < 90 {
			t.Errorf("device %s: expected disk >= 90%%, got %.1f%%", id, d.DiskPercent)
		}
	}

	// Non-affected devices should be normal
	for _, d := range fleet.ListDevices() {
		affected := false
		for _, id := range targetIDs {
			if d.ID == id {
				affected = true
				break
			}
		}
		if !affected && d.DiskPercent > 90 {
			t.Errorf("unaffected device %s has high disk: %.1f%%", d.ID, d.DiskPercent)
		}
	}
}

// TestOOMScenario simulates devices running out of memory.
func TestOOMScenario(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(10, simulator.PredefinedProfiles["jetson-nano"], simulator.PredefinedNetworkProfiles["good"])

	devices := fleet.ListDevices()

	// Inject high memory usage on all devices
	allIDs := make([]string, len(devices))
	for i, d := range devices {
		allIDs[i] = d.ID
	}
	fleet.InjectHighMemory(allIDs, 95) // 95% memory usage

	for i := 0; i < 20; i++ {
		fleet.SimulateMetrics()
	}

	for _, d := range fleet.ListDevices() {
		memPercent := float64(d.RAMUsedMB) / float64(d.Profile.RAMMB) * 100
		if memPercent < 80 {
			t.Errorf("device %s: expected high memory, got %.1f%%", d.ID, memPercent)
		}
	}
}

// TestThermalThrottling simulates devices overheating.
func TestThermalThrottling(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(10, simulator.PredefinedProfiles["jetson-nano"], simulator.PredefinedNetworkProfiles["good"])

	devices := fleet.ListDevices()
	allIDs := make([]string, len(devices))
	for i, d := range devices {
		allIDs[i] = d.ID
	}

	// Inject high temperature
	fleet.InjectHighTemperature(allIDs, 85.0)

	for i := 0; i < 30; i++ {
		fleet.SimulateMetrics()
	}

	warningCount := 0
	for _, d := range fleet.ListDevices() {
		if d.Temperature > 80 {
			warningCount++
		}
	}

	if warningCount == 0 {
		t.Error("expected some devices with high temperature")
	}
	t.Logf("%d/%d devices in thermal warning", warningCount, len(devices))
}

// TestCorruptedModelDeployment simulates deploying a corrupted model.
func TestCorruptedModelDeployment(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(10, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["good"])

	devices := fleet.ListDevices()
	allIDs := make([]string, len(devices))
	for i, d := range devices {
		allIDs[i] = d.ID
	}

	// Deploy valid model first
	fleet.DeployModel(allIDs, "model-v1")

	// Attempt to deploy corrupted model
	fleet.DeployCorruptedModel(allIDs, "model-v2-corrupt")

	// Devices should reject and stay on v1
	v1Count := fleet.CountDeployedModel("model-v1")
	corruptCount := fleet.CountDeployedModel("model-v2-corrupt")

	if corruptCount > 0 {
		t.Errorf("corrupted model should be rejected, but %d devices accepted it", corruptCount)
	}
	if v1Count != 10 {
		t.Errorf("expected 10 devices on v1 after corrupt rejection, got %d", v1Count)
	}
}

// TestGPUFailure simulates GPU failure on compute devices.
func TestGPUFailure(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(10, simulator.PredefinedProfiles["jetson-nano"], simulator.PredefinedNetworkProfiles["good"])

	devices := fleet.ListDevices()
	failIDs := []string{devices[0].ID, devices[1].ID}

	fleet.InjectGPUFailure(failIDs)

	for i := 0; i < 30; i++ {
		fleet.SimulateMetrics()
	}

	// Devices with GPU failure should show 0 GPU utilization or error status
	for _, id := range failIDs {
		d := fleet.GetDevice(id)
		if d.GPUPercent > 0 {
			t.Logf("device %s: GPU should be 0 after failure, got %.1f%%", id, d.GPUPercent)
		}
	}
}

// TestSimultaneousFailures simulates multiple failure types at once.
func TestSimultaneousFailures(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddHeterogeneousFleet(30)

	devices := fleet.ListDevices()

	// Different failures on different groups
	fleet.InjectDiskFull([]string{devices[0].ID, devices[1].ID})
	fleet.InjectHighMemory([]string{devices[5].ID, devices[6].ID}, 95)
	fleet.InjectHighTemperature([]string{devices[10].ID, devices[11].ID}, 90)
	fleet.TakeOffline([]string{devices[15].ID, devices[16].ID})

	for i := 0; i < 50; i++ {
		fleet.SimulateMetrics()
	}

	// Count status distribution
	statusCount := make(map[string]int)
	for _, d := range fleet.ListDevices() {
		statusCount[d.Status]++
	}

	t.Logf("Status distribution after multi-failure: %v", statusCount)

	// Most devices should still be operational
	healthy := statusCount["healthy"] + statusCount["warning"]
	if healthy < 20 {
		t.Errorf("expected at least 20 operational devices, got %d", healthy)
	}
}
