//go:build fleet

package fleet

import (
	"os"
	"strconv"
	"testing"

	"github.com/fleetml/fleetml/simulator"
	"go.uber.org/zap"
)

// TestHeterogeneousFleet creates a mixed fleet and validates all device profiles.
func TestHeterogeneousFleet(t *testing.T) {
	fleetSize := 20
	if s := os.Getenv("FLEET_SIZE"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			fleetSize = n
		}
	}

	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddHeterogeneousFleet(fleetSize)

	if fleet.Size() != fleetSize {
		t.Fatalf("expected fleet size %d, got %d", fleetSize, fleet.Size())
	}

	// Validate all devices have valid profiles
	for _, d := range fleet.ListDevices() {
		if d.Profile.Arch == "" {
			t.Errorf("device %s has empty arch", d.ID)
		}
		if d.Profile.RAMMB == 0 {
			t.Errorf("device %s has 0 RAM", d.ID)
		}
		if d.Status != "healthy" {
			t.Errorf("device %s initial status should be healthy, got %q", d.ID, d.Status)
		}
	}

	// Validate heterogeneity
	byProfile := fleet.DevicesByProfile()
	if len(byProfile) < 3 {
		t.Errorf("expected at least 3 device profiles in heterogeneous fleet, got %d", len(byProfile))
	}

	t.Logf("Fleet created: %d devices across %d profiles", fleet.Size(), len(byProfile))
	for name, devices := range byProfile {
		t.Logf("  %s: %d devices", name, len(devices))
	}
}

// TestFleetMetricSimulation validates that metrics change realistically over time.
func TestFleetMetricSimulation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := simulator.NewFleet(logger.Sugar())
	fleet.AddDevices(10, simulator.PredefinedProfiles["jetson-nano"], simulator.PredefinedNetworkProfiles["good"])

	// Run 100 simulation ticks
	for i := 0; i < 100; i++ {
		fleet.SimulateMetrics()
	}

	for _, d := range fleet.ListDevices() {
		if d.CPUPercent < 0 || d.CPUPercent > 100 {
			t.Errorf("device %s: CPU %f out of range", d.ID, d.CPUPercent)
		}
		if d.RAMUsedMB < 0 || d.RAMUsedMB > d.Profile.RAMMB {
			t.Errorf("device %s: RAM %d out of range (max %d)", d.ID, d.RAMUsedMB, d.Profile.RAMMB)
		}
		if d.Temperature < 0 || d.Temperature > 100 {
			t.Errorf("device %s: temp %f out of range", d.ID, d.Temperature)
		}
	}
}
