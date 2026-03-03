package simulator

import (
	"testing"

	"go.uber.org/zap"
)

func TestFleet_AddDevices(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := NewFleet(logger.Sugar())

	fleet.AddDevices(5, PredefinedProfiles["jetson-nano"], PredefinedNetworkProfiles["good"])

	if fleet.Size() != 5 {
		t.Errorf("expected fleet size 5, got %d", fleet.Size())
	}

	devices := fleet.ListDevices()
	for _, d := range devices {
		if d.Profile.Name != "jetson-nano" {
			t.Errorf("expected profile 'jetson-nano', got %q", d.Profile.Name)
		}
		if d.Status != "healthy" {
			t.Errorf("expected status 'healthy', got %q", d.Status)
		}
	}
}

func TestFleet_HeterogeneousFleet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := NewFleet(logger.Sugar())

	fleet.AddHeterogeneousFleet(20)

	if fleet.Size() != 20 {
		t.Errorf("expected fleet size 20, got %d", fleet.Size())
	}

	// Should have multiple profiles
	byProfile := fleet.DevicesByProfile()
	if len(byProfile) < 3 {
		t.Errorf("expected at least 3 different profiles, got %d", len(byProfile))
	}
}

func TestFleet_SimulateMetrics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := NewFleet(logger.Sugar())

	fleet.AddDevices(3, PredefinedProfiles["rpi4"], PredefinedNetworkProfiles["good"])

	devices := fleet.ListDevices()
	initialCPU := make(map[string]float64)
	for _, d := range devices {
		initialCPU[d.ID] = d.CPUPercent
	}

	// Simulate several rounds of metrics
	for i := 0; i < 10; i++ {
		fleet.SimulateMetrics()
	}

	// Metrics should have changed
	devices = fleet.ListDevices()
	changed := false
	for _, d := range devices {
		if d.CPUPercent != initialCPU[d.ID] {
			changed = true
			break
		}
	}
	if !changed {
		t.Error("expected metrics to change after simulation")
	}
}

func TestFleet_TakeOfflineAndBringBack(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := NewFleet(logger.Sugar())

	fleet.AddDevices(5, PredefinedProfiles["intel-nuc"], PredefinedNetworkProfiles["excellent"])

	devices := fleet.ListDevices()
	offlineIDs := []string{devices[0].ID, devices[1].ID}

	fleet.TakeOffline(offlineIDs)

	for _, id := range offlineIDs {
		d := fleet.GetDevice(id)
		if d.Status != "offline" {
			t.Errorf("expected device %s to be offline, got %q", id, d.Status)
		}
	}

	// Remaining should still be healthy
	for _, d := range fleet.ListDevices() {
		isOffline := false
		for _, oid := range offlineIDs {
			if d.ID == oid {
				isOffline = true
				break
			}
		}
		if !isOffline && d.Status != "healthy" {
			t.Errorf("expected device %s to be healthy, got %q", d.ID, d.Status)
		}
	}

	// Bring back online
	fleet.BringOnline(offlineIDs, PredefinedNetworkProfiles["good"])
	for _, id := range offlineIDs {
		d := fleet.GetDevice(id)
		if d.Status != "healthy" {
			t.Errorf("expected device %s to be healthy after coming online, got %q", id, d.Status)
		}
	}
}

func TestFleet_OfflineDevicesNotSimulated(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := NewFleet(logger.Sugar())

	fleet.AddDevices(2, PredefinedProfiles["rpi4"], PredefinedNetworkProfiles["good"])

	devices := fleet.ListDevices()
	offlineID := devices[0].ID
	fleet.TakeOffline([]string{offlineID})

	initialUptime := fleet.GetDevice(offlineID).UptimeHours

	for i := 0; i < 10; i++ {
		fleet.SimulateMetrics()
	}

	// Offline device uptime should not change
	if fleet.GetDevice(offlineID).UptimeHours != initialUptime {
		t.Error("offline device metrics should not change")
	}
}

func TestPredefinedProfiles_AllValid(t *testing.T) {
	for name, profile := range PredefinedProfiles {
		if profile.Arch == "" {
			t.Errorf("profile %s has empty arch", name)
		}
		if profile.RAMMB == 0 {
			t.Errorf("profile %s has 0 RAM", name)
		}
		if profile.DiskGB == 0 {
			t.Errorf("profile %s has 0 disk", name)
		}
		if profile.CPUCores == 0 {
			t.Errorf("profile %s has 0 CPU cores", name)
		}
	}
}

func TestPredefinedNetworkProfiles_AllValid(t *testing.T) {
	for name, profile := range PredefinedNetworkProfiles {
		if name == "offline" {
			if profile.PacketLoss != 1.0 {
				t.Errorf("offline profile should have 100%% packet loss")
			}
			continue
		}
		if profile.LatencyMs == 0 {
			t.Errorf("network profile %s has 0 latency", name)
		}
	}
}

func TestFleet_GetDevice_NonExistent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fleet := NewFleet(logger.Sugar())

	if fleet.GetDevice("nonexistent") != nil {
		t.Error("expected nil for nonexistent device")
	}
}
