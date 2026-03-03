//go:build fleet

package fleet

import (
	"testing"

	"github.com/fleetml/fleetml/simulator"
	"go.uber.org/zap"
)

// TestOfflineResilience simulates devices going offline and coming back.
func TestOfflineResilience(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(10, simulator.PredefinedProfiles["rpi4"], simulator.PredefinedNetworkProfiles["good"])

	devices := fleet.ListDevices()

	// Take 5 devices offline
	offlineIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		offlineIDs[i] = devices[i].ID
	}
	fleet.TakeOffline(offlineIDs)

	// Verify offline count
	offlineCount := 0
	for _, d := range fleet.ListDevices() {
		if d.Status == "offline" {
			offlineCount++
		}
	}
	if offlineCount != 5 {
		t.Errorf("expected 5 offline devices, got %d", offlineCount)
	}

	// Simulate some ticks (offline devices should not update)
	for i := 0; i < 20; i++ {
		fleet.SimulateMetrics()
	}

	// Bring devices back online
	fleet.BringOnline(offlineIDs, simulator.PredefinedNetworkProfiles["cellular"])

	// All should be healthy now
	for _, d := range fleet.ListDevices() {
		if d.Status == "offline" {
			t.Errorf("device %s should be online", d.ID)
		}
	}

	// Verify network profile changed for reconnected devices
	for _, id := range offlineIDs {
		d := fleet.GetDevice(id)
		if d.Network.Name != "cellular" {
			t.Errorf("device %s: expected 'cellular' network, got %q", id, d.Network.Name)
		}
	}
}

// TestPartialOffline simulates intermittent connectivity.
func TestPartialOffline(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(20, simulator.PredefinedProfiles["intel-nuc"], simulator.PredefinedNetworkProfiles["excellent"])

	devices := fleet.ListDevices()

	// Cycle through taking groups offline and bringing them back
	for cycle := 0; cycle < 3; cycle++ {
		start := cycle * 5
		end := start + 5
		if end > len(devices) {
			end = len(devices)
		}

		groupIDs := make([]string, 0, end-start)
		for i := start; i < end; i++ {
			groupIDs = append(groupIDs, devices[i].ID)
		}

		fleet.TakeOffline(groupIDs)

		// Run some metrics ticks while partially offline
		for i := 0; i < 5; i++ {
			fleet.SimulateMetrics()
		}

		fleet.BringOnline(groupIDs, simulator.PredefinedNetworkProfiles["good"])
	}

	// All devices should be back online
	for _, d := range fleet.ListDevices() {
		if d.Status == "offline" {
			t.Errorf("device %s still offline after all cycles", d.ID)
		}
	}
}
