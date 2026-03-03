//go:build chaos

package chaos

import (
	"testing"
	"time"

	"github.com/fleetml/fleetml/simulator"
	"go.uber.org/zap"
)

// TestNetworkPartition simulates a network split where half the fleet
// becomes unreachable, then recovers.
func TestNetworkPartition(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddHeterogeneousFleet(50)

	devices := fleet.ListDevices()

	// Partition: take first half offline
	partitionIDs := make([]string, len(devices)/2)
	for i := 0; i < len(partitionIDs); i++ {
		partitionIDs[i] = devices[i].ID
	}

	t.Logf("Partitioning %d/%d devices", len(partitionIDs), len(devices))
	fleet.TakeOffline(partitionIDs)

	// Simulate 60 seconds of operation (2 ticks per second)
	for i := 0; i < 120; i++ {
		fleet.SimulateMetrics()
	}

	// Verify online devices are still healthy
	for _, d := range fleet.ListDevices() {
		if d.Status == "offline" {
			continue
		}
		if d.CPUPercent < 0 || d.CPUPercent > 100 {
			t.Errorf("online device %s has invalid CPU: %f", d.ID, d.CPUPercent)
		}
	}

	// Heal partition
	t.Log("Healing network partition")
	fleet.BringOnline(partitionIDs, simulator.PredefinedNetworkProfiles["good"])

	// Verify recovery
	offlineCount := 0
	for _, d := range fleet.ListDevices() {
		if d.Status == "offline" {
			offlineCount++
		}
	}
	if offlineCount != 0 {
		t.Errorf("expected 0 offline after heal, got %d", offlineCount)
	}
}

// TestRollingOutage simulates devices going offline one by one and
// recovering after a delay.
func TestRollingOutage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(20, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["cellular"])

	devices := fleet.ListDevices()

	// Take devices offline one at a time, bring back after 5 ticks
	for i, d := range devices {
		fleet.TakeOffline([]string{d.ID})
		fleet.SimulateMetrics()

		// Bring back earlier devices
		if i >= 5 {
			fleet.BringOnline([]string{devices[i-5].ID}, simulator.PredefinedNetworkProfiles["good"])
		}
	}

	// Bring remaining back
	for i := len(devices) - 5; i < len(devices); i++ {
		fleet.BringOnline([]string{devices[i].ID}, simulator.PredefinedNetworkProfiles["good"])
	}

	offlineCount := 0
	for _, d := range fleet.ListDevices() {
		if d.Status == "offline" {
			offlineCount++
		}
	}
	if offlineCount != 0 {
		t.Errorf("expected all devices online after rolling recovery, got %d offline", offlineCount)
	}
}

// TestDegradedNetwork simulates all devices switching to poor network conditions.
func TestDegradedNetwork(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(30, simulator.PredefinedProfiles["intel-nuc"], simulator.PredefinedNetworkProfiles["excellent"])

	// Degrade all networks to "poor"
	devices := fleet.ListDevices()
	allIDs := make([]string, len(devices))
	for i, d := range devices {
		allIDs[i] = d.ID
	}

	// Simulate poor network by taking offline then bringing back with poor profile
	fleet.TakeOffline(allIDs)
	fleet.BringOnline(allIDs, simulator.PredefinedNetworkProfiles["poor"])

	// All should still be online (poor != offline)
	for _, d := range fleet.ListDevices() {
		if d.Status == "offline" {
			t.Errorf("device %s should be online with poor network", d.ID)
		}
		if d.Network.Name != "poor" {
			t.Errorf("device %s: expected 'poor' network, got %q", d.ID, d.Network.Name)
		}
	}

	// Simulate some operation
	start := time.Now()
	for i := 0; i < 50; i++ {
		fleet.SimulateMetrics()
	}
	t.Logf("50 ticks with degraded network: %v", time.Since(start))
}
