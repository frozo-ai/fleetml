//go:build fleet

package scale

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/fleetml/fleetml/simulator"
	"go.uber.org/zap"
)

// TestThousandDevices validates fleet creation and metric simulation at scale.
func TestThousandDevices(t *testing.T) {
	fleetSize := 1000
	if s := os.Getenv("FLEET_SIZE"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			fleetSize = n
		}
	}

	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())

	// Time fleet creation
	start := time.Now()
	fleet.AddHeterogeneousFleet(fleetSize)
	createDuration := time.Since(start)

	if fleet.Size() != fleetSize {
		t.Fatalf("expected fleet size %d, got %d", fleetSize, fleet.Size())
	}

	t.Logf("Fleet creation: %d devices in %v", fleetSize, createDuration)
	if createDuration > 5*time.Second {
		t.Errorf("fleet creation too slow: %v (target <5s)", createDuration)
	}

	// Time metric simulation (100 ticks)
	start = time.Now()
	for i := 0; i < 100; i++ {
		fleet.SimulateMetrics()
	}
	simDuration := time.Since(start)

	t.Logf("100 simulation ticks: %v (%.2fms per tick)",
		simDuration, float64(simDuration.Milliseconds())/100.0)

	// Validate all devices still healthy or warning (none invalid)
	for _, d := range fleet.ListDevices() {
		if d.Status != "healthy" && d.Status != "warning" {
			t.Errorf("device %s has unexpected status %q", d.ID, d.Status)
		}
		if d.CPUPercent < 0 || d.CPUPercent > 100 {
			t.Errorf("device %s: CPU %f out of range", d.ID, d.CPUPercent)
		}
	}

	// Profile breakdown
	byProfile := fleet.DevicesByProfile()
	t.Logf("Profile distribution across %d profiles:", len(byProfile))
	for name, devices := range byProfile {
		t.Logf("  %s: %d devices (%.1f%%)",
			name, len(devices), float64(len(devices))/float64(fleetSize)*100)
	}
}

// TestFleetOfflineRecovery validates offline/reconnect at scale.
func TestFleetOfflineRecovery(t *testing.T) {
	fleetSize := 100
	if s := os.Getenv("FLEET_SIZE"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			fleetSize = n
		}
	}

	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddHeterogeneousFleet(fleetSize)

	devices := fleet.ListDevices()

	// Take 50% offline
	offlineCount := fleetSize / 2
	offlineIDs := make([]string, offlineCount)
	for i := 0; i < offlineCount; i++ {
		offlineIDs[i] = devices[i].ID
	}

	start := time.Now()
	fleet.TakeOffline(offlineIDs)
	offlineDuration := time.Since(start)
	t.Logf("Taking %d devices offline: %v", offlineCount, offlineDuration)

	// Simulate some ticks
	for i := 0; i < 10; i++ {
		fleet.SimulateMetrics()
	}

	// Bring back online
	start = time.Now()
	fleet.BringOnline(offlineIDs, simulator.PredefinedNetworkProfiles["cellular"])
	onlineDuration := time.Since(start)
	t.Logf("Bringing %d devices online: %v", offlineCount, onlineDuration)

	// All should be back
	offlineAfter := 0
	for _, d := range fleet.ListDevices() {
		if d.Status == "offline" {
			offlineAfter++
		}
	}
	if offlineAfter != 0 {
		t.Errorf("expected 0 offline devices after recovery, got %d", offlineAfter)
	}
}
