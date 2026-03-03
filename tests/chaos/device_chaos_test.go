//go:build chaos

package chaos

import (
	"math/rand"
	"testing"
	"time"

	"github.com/fleetml/fleetml/simulator"
	"go.uber.org/zap"
)

// TestRandomDeviceFailures randomly kills and restores devices over time.
func TestRandomDeviceFailures(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddHeterogeneousFleet(50)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	devices := fleet.ListDevices()

	// Run 100 rounds of random chaos
	for round := 0; round < 100; round++ {
		fleet.SimulateMetrics()

		// 10% chance of killing a random device
		if rng.Float64() < 0.10 {
			idx := rng.Intn(len(devices))
			if fleet.GetDevice(devices[idx].ID).Status != "offline" {
				fleet.TakeOffline([]string{devices[idx].ID})
			}
		}

		// 20% chance of restoring a random offline device
		if rng.Float64() < 0.20 {
			for _, d := range fleet.ListDevices() {
				if d.Status == "offline" {
					fleet.BringOnline([]string{d.ID}, simulator.PredefinedNetworkProfiles["good"])
					break
				}
			}
		}
	}

	// At the end, count status distribution
	statusCount := make(map[string]int)
	for _, d := range fleet.ListDevices() {
		statusCount[d.Status]++
	}

	t.Logf("After 100 chaos rounds: %v", statusCount)

	// Should not have lost all devices
	if statusCount["healthy"]+statusCount["warning"] == 0 {
		t.Error("all devices offline after chaos — fleet should be partially available")
	}
}

// TestCPUSpike simulates all devices experiencing high CPU usage.
func TestCPUSpike(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(20, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["good"])

	// Run many ticks to potentially trigger high CPU
	warningCount := 0
	for i := 0; i < 500; i++ {
		fleet.SimulateMetrics()
		for _, d := range fleet.ListDevices() {
			if d.Status == "warning" {
				warningCount++
			}
		}
	}

	t.Logf("Warning state occurrences over 500 ticks: %d", warningCount)
	// Some devices should occasionally enter warning state
}
