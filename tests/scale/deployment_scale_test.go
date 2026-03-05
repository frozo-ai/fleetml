//go:build fleet

package scale

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/fleetml/fleetml/simulator"
	"go.uber.org/zap"
)

// TestDeploymentAtScale validates deploying a model to 500+ devices.
func TestDeploymentAtScale(t *testing.T) {
	fleetSize := 500
	if s := os.Getenv("FLEET_SIZE"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			fleetSize = n
		}
	}

	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddHeterogeneousFleet(fleetSize)

	devices := fleet.ListDevices()
	allIDs := make([]string, len(devices))
	for i, d := range devices {
		allIDs[i] = d.ID
	}

	// Time deployment
	start := time.Now()
	fleet.DeployModel(allIDs, "model-v1")
	deployDuration := time.Since(start)

	deployed := fleet.CountDeployedModel("model-v1")
	if deployed != fleetSize {
		t.Errorf("expected %d deployed, got %d", fleetSize, deployed)
	}

	t.Logf("Deployed to %d devices in %v", fleetSize, deployDuration)

	// Deployment to 500 devices should complete in under 10 seconds
	if deployDuration > 10*time.Second {
		t.Errorf("deployment too slow: %v (target <10s for %d devices)", deployDuration, fleetSize)
	}
}

// TestCanaryAtScale validates canary rollout to 1000+ devices.
func TestCanaryAtScale(t *testing.T) {
	fleetSize := 1000
	if s := os.Getenv("FLEET_SIZE"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			fleetSize = n
		}
	}

	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddHeterogeneousFleet(fleetSize)

	devices := fleet.ListDevices()
	allIDs := make([]string, len(devices))
	for i, d := range devices {
		allIDs[i] = d.ID
	}

	// Deploy v1 to all
	fleet.DeployModel(allIDs, "model-v1")

	// Canary stages: 1% → 10% → 50% → 100%
	stages := []float64{0.01, 0.10, 0.50, 1.0}
	for _, pct := range stages {
		count := int(float64(fleetSize) * pct)
		if count < 1 {
			count = 1
		}
		stageIDs := allIDs[:count]

		start := time.Now()
		fleet.DeployModel(stageIDs, "model-v2")
		duration := time.Since(start)

		deployed := fleet.CountDeployedModel("model-v2")
		if deployed != count {
			t.Errorf("stage %.0f%%: expected %d deployed, got %d", pct*100, count, deployed)
		}

		t.Logf("Stage %.0f%% (%d devices): %v", pct*100, count, duration)
	}
}

// TestConcurrentMetricsAtScale validates metric collection at scale.
func TestConcurrentMetricsAtScale(t *testing.T) {
	fleetSize := 500
	if s := os.Getenv("FLEET_SIZE"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			fleetSize = n
		}
	}

	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddHeterogeneousFleet(fleetSize)

	// Time 1000 metric ticks
	start := time.Now()
	for i := 0; i < 1000; i++ {
		fleet.SimulateMetrics()
	}
	duration := time.Since(start)

	perTick := float64(duration.Microseconds()) / 1000.0
	t.Logf("1000 ticks for %d devices: %v (%.1fμs per tick)", fleetSize, duration, perTick)

	// Should complete in under 30 seconds
	if duration > 30*time.Second {
		t.Errorf("metric simulation too slow: %v", duration)
	}
}

// TestProfileDistributionAtScale validates heterogeneous fleet distribution at scale.
func TestProfileDistributionAtScale(t *testing.T) {
	fleetSize := 1000
	if s := os.Getenv("FLEET_SIZE"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			fleetSize = n
		}
	}

	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddHeterogeneousFleet(fleetSize)

	byProfile := fleet.DevicesByProfile()

	// Should have at least 3 different profiles
	if len(byProfile) < 3 {
		t.Errorf("expected at least 3 profiles, got %d", len(byProfile))
	}

	// Each profile should have a reasonable share
	for name, devices := range byProfile {
		pct := float64(len(devices)) / float64(fleetSize) * 100
		if pct < 1 {
			t.Errorf("profile %s has only %.1f%% of fleet — should be more balanced", name, pct)
		}
		t.Logf("  %s: %d devices (%.1f%%)", name, len(devices), pct)
	}
}

// TestMultiModelDeploymentAtScale validates deploying multiple models across a fleet.
func TestMultiModelDeploymentAtScale(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())

	// Create fleet with 4 device types, 100 each
	fleet.AddDevices(100, simulator.PredefinedProfiles["jetson-nano"], simulator.PredefinedNetworkProfiles["good"])
	fleet.AddDevices(100, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["good"])
	fleet.AddDevices(100, simulator.PredefinedProfiles["intel-nuc"], simulator.PredefinedNetworkProfiles["excellent"])
	fleet.AddDevices(100, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["cellular"])

	byProfile := fleet.DevicesByProfile()

	// Deploy different compiled variants to each profile
	for profileName, profileDevices := range byProfile {
		ids := make([]string, len(profileDevices))
		for i, d := range profileDevices {
			ids[i] = d.ID
		}
		fleet.DeployModel(ids, fmt.Sprintf("mobilenet-%s", profileName))
	}

	// Verify each profile has its correct variant
	for profileName, profileDevices := range byProfile {
		expectedModel := fmt.Sprintf("mobilenet-%s", profileName)
		count := 0
		for _, d := range profileDevices {
			device := fleet.GetDevice(d.ID)
			if device.DeployedModel == expectedModel {
				count++
			}
		}
		if count != len(profileDevices) {
			t.Errorf("profile %s: expected %d with model %q, got %d",
				profileName, len(profileDevices), expectedModel, count)
		}
	}
}
