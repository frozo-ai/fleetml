//go:build fleet

package fleet

import (
	"testing"

	"github.com/fleetml/fleetml/simulator"
	"go.uber.org/zap"
)

// TestNetworkProfileTransitions validates changing network conditions.
func TestNetworkProfileTransitions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(10, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["excellent"])

	devices := fleet.ListDevices()
	allIDs := make([]string, len(devices))
	for i, d := range devices {
		allIDs[i] = d.ID
	}

	// Cycle through all network profiles
	profiles := []string{"excellent", "good", "cellular", "poor"}
	for _, profile := range profiles {
		fleet.TakeOffline(allIDs)
		fleet.BringOnline(allIDs, simulator.PredefinedNetworkProfiles[profile])

		for _, d := range fleet.ListDevices() {
			if d.Network.Name != profile {
				t.Errorf("expected network %q, got %q for device %s", profile, d.Network.Name, d.ID)
			}
		}
	}
}

// TestMixedNetworkFleet validates a fleet with diverse network conditions.
func TestMixedNetworkFleet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())

	fleet.AddDevices(5, simulator.PredefinedProfiles["intel-nuc"], simulator.PredefinedNetworkProfiles["excellent"])
	fleet.AddDevices(5, simulator.PredefinedProfiles["jetson-nano"], simulator.PredefinedNetworkProfiles["good"])
	fleet.AddDevices(5, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["cellular"])
	fleet.AddDevices(5, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["poor"])

	if fleet.Size() != 20 {
		t.Fatalf("expected fleet size 20, got %d", fleet.Size())
	}

	// Run simulation — all should remain operational
	for i := 0; i < 50; i++ {
		fleet.SimulateMetrics()
	}

	offlineCount := 0
	for _, d := range fleet.ListDevices() {
		if d.Status == "offline" {
			offlineCount++
		}
	}
	if offlineCount > 0 {
		t.Errorf("expected 0 offline devices in mixed network fleet, got %d", offlineCount)
	}
}

// TestCellularToWifi validates transitioning from cellular to wifi.
func TestCellularToWifi(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(10, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["cellular"])

	devices := fleet.ListDevices()

	// Verify all on cellular
	for _, d := range devices {
		if d.Network.Name != "cellular" {
			t.Errorf("expected cellular, got %q", d.Network.Name)
		}
	}

	// Transition to good wifi
	allIDs := make([]string, len(devices))
	for i, d := range devices {
		allIDs[i] = d.ID
	}
	fleet.TakeOffline(allIDs)
	fleet.BringOnline(allIDs, simulator.PredefinedNetworkProfiles["good"])

	for _, d := range fleet.ListDevices() {
		if d.Network.Name != "good" {
			t.Errorf("expected good, got %q after transition", d.Network.Name)
		}
	}
}

// TestFleetGroupingByNetwork validates grouping devices by network profile.
func TestFleetGroupingByNetwork(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())

	fleet.AddDevices(3, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["excellent"])
	fleet.AddDevices(3, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["poor"])

	byNetwork := fleet.DevicesByNetwork()
	if len(byNetwork) < 2 {
		t.Errorf("expected at least 2 network groups, got %d", len(byNetwork))
	}

	if len(byNetwork["excellent"]) != 3 {
		t.Errorf("expected 3 excellent devices, got %d", len(byNetwork["excellent"]))
	}
	if len(byNetwork["poor"]) != 3 {
		t.Errorf("expected 3 poor devices, got %d", len(byNetwork["poor"]))
	}
}

// TestFleetHeartbeatWithNetworkJitter validates heartbeat behavior under jitter.
func TestFleetHeartbeatWithNetworkJitter(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(20, simulator.PredefinedProfiles["jetson-nano"], simulator.PredefinedNetworkProfiles["poor"])

	// Run 200 ticks — with poor network, some heartbeats may be delayed
	for i := 0; i < 200; i++ {
		fleet.SimulateMetrics()
	}

	// All devices should still report valid metrics
	for _, d := range fleet.ListDevices() {
		if d.CPUPercent < 0 || d.CPUPercent > 100 {
			t.Errorf("device %s: invalid CPU %f", d.ID, d.CPUPercent)
		}
	}
}
