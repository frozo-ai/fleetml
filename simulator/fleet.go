package simulator

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
)

// VirtualDevice simulates an edge agent.
type VirtualDevice struct {
	ID             string
	Profile        DeviceProfile
	Network        NetworkProfile
	Status         string // healthy, warning, offline
	CurrentModel   string
	ModelVersion   string
	CPUPercent     float64
	RAMUsedMB      int
	DiskPercent    float64
	Temperature    float64
	UptimeHours    float64
}

// Fleet manages a collection of virtual devices for testing.
type Fleet struct {
	mu      sync.RWMutex
	devices map[string]*VirtualDevice
	logger  *zap.SugaredLogger
	rng     *rand.Rand
}

// NewFleet creates a new virtual fleet.
func NewFleet(logger *zap.SugaredLogger) *Fleet {
	return &Fleet{
		devices: make(map[string]*VirtualDevice),
		logger:  logger,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// AddDevices adds virtual devices to the fleet.
func (f *Fleet) AddDevices(count int, profile DeviceProfile, network NetworkProfile) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i := 0; i < count; i++ {
		id := fmt.Sprintf("sim-%s-%04d", profile.Name, len(f.devices)+1)
		dev := &VirtualDevice{
			ID:          id,
			Profile:     profile,
			Network:     network,
			Status:      "healthy",
			CPUPercent:  f.randomCPU(),
			RAMUsedMB:   f.randomRAM(profile.RAMMB),
			DiskPercent: f.randomDisk(),
			Temperature: f.randomTemp(),
			UptimeHours: f.rng.Float64() * 720, // 0-30 days
		}
		f.devices[id] = dev
	}

	f.logger.Infow("devices added",
		"count", count,
		"profile", profile.Name,
		"network", network.Name,
		"total_fleet_size", len(f.devices),
	)
}

// AddHeterogeneousFleet creates a mixed fleet with various device types.
func (f *Fleet) AddHeterogeneousFleet(totalDevices int) {
	distribution := map[string]float64{
		"jetson-nano":  0.20,
		"jetson-orin":  0.10,
		"rpi4":         0.25,
		"rpi5":         0.10,
		"intel-nuc":    0.15,
		"generic-x86":  0.10,
		"hailo-8":      0.05,
		"qualcomm-rb5": 0.05,
	}

	networkDist := map[string]float64{
		"excellent": 0.30,
		"good":      0.40,
		"cellular":  0.20,
		"poor":      0.10,
	}

	networks := pickNetwork(f.rng, networkDist, totalDevices)

	assigned := 0
	for profileName, ratio := range distribution {
		count := int(float64(totalDevices) * ratio)
		if assigned+count > totalDevices {
			count = totalDevices - assigned
		}

		profile := PredefinedProfiles[profileName]
		for i := 0; i < count; i++ {
			net := networks[(assigned+i)%len(networks)]
			f.AddDevices(1, profile, net)
		}
		assigned += count
	}

	// Fill remaining with generic
	for assigned < totalDevices {
		f.AddDevices(1, PredefinedProfiles["generic-x86"], PredefinedNetworkProfiles["good"])
		assigned++
	}
}

// ListDevices returns all virtual devices.
func (f *Fleet) ListDevices() []*VirtualDevice {
	f.mu.RLock()
	defer f.mu.RUnlock()

	devices := make([]*VirtualDevice, 0, len(f.devices))
	for _, d := range f.devices {
		devices = append(devices, d)
	}
	return devices
}

// GetDevice returns a specific virtual device.
func (f *Fleet) GetDevice(id string) *VirtualDevice {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.devices[id]
}

// Size returns the fleet size.
func (f *Fleet) Size() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.devices)
}

// SimulateMetrics updates device metrics with realistic jitter.
func (f *Fleet) SimulateMetrics() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, dev := range f.devices {
		if dev.Status == "offline" {
			continue
		}

		// Add realistic jitter to CPU, RAM, disk, temperature
		dev.CPUPercent += (f.rng.Float64() - 0.5) * 10
		if dev.CPUPercent < 0 {
			dev.CPUPercent = 0
		}
		if dev.CPUPercent > 100 {
			dev.CPUPercent = 100
		}

		dev.RAMUsedMB += int(f.rng.Float64()*200) - 100
		if dev.RAMUsedMB < 100 {
			dev.RAMUsedMB = 100
		}
		if dev.RAMUsedMB > dev.Profile.RAMMB {
			dev.RAMUsedMB = dev.Profile.RAMMB
		}

		dev.Temperature += (f.rng.Float64() - 0.5) * 3
		if dev.Temperature < 20 {
			dev.Temperature = 20
		}
		if dev.Temperature > 90 {
			dev.Temperature = 90
		}

		dev.UptimeHours += 1.0 / 120.0 // ~30 seconds

		// Random status changes
		if dev.CPUPercent > 90 || dev.Temperature > 80 {
			dev.Status = "warning"
		} else {
			dev.Status = "healthy"
		}
	}
}

// TakeOffline marks devices as offline to simulate network issues.
func (f *Fleet) TakeOffline(deviceIDs []string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, id := range deviceIDs {
		if dev, ok := f.devices[id]; ok {
			dev.Status = "offline"
			dev.Network = PredefinedNetworkProfiles["offline"]
		}
	}
}

// BringOnline restores devices from offline state.
func (f *Fleet) BringOnline(deviceIDs []string, network NetworkProfile) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, id := range deviceIDs {
		if dev, ok := f.devices[id]; ok {
			dev.Status = "healthy"
			dev.Network = network
		}
	}
}

// SimulateLoop runs periodic metric simulation until context is cancelled.
func (f *Fleet) SimulateLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			f.SimulateMetrics()
		}
	}
}

// DevicesByProfile returns devices grouped by profile name.
func (f *Fleet) DevicesByProfile() map[string][]*VirtualDevice {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make(map[string][]*VirtualDevice)
	for _, d := range f.devices {
		result[d.Profile.Name] = append(result[d.Profile.Name], d)
	}
	return result
}

func (f *Fleet) randomCPU() float64 {
	return 10 + f.rng.Float64()*40 // 10-50%
}

func (f *Fleet) randomRAM(totalMB int) int {
	return int(float64(totalMB) * (0.3 + f.rng.Float64()*0.4)) // 30-70%
}

func (f *Fleet) randomDisk() float64 {
	return 20 + f.rng.Float64()*50 // 20-70%
}

func (f *Fleet) randomTemp() float64 {
	return 35 + f.rng.Float64()*25 // 35-60C
}

func pickNetwork(rng *rand.Rand, dist map[string]float64, count int) []NetworkProfile {
	var result []NetworkProfile
	for name, ratio := range dist {
		n := int(float64(count) * ratio)
		for i := 0; i < n; i++ {
			result = append(result, PredefinedNetworkProfiles[name])
		}
	}
	for len(result) < count {
		result = append(result, PredefinedNetworkProfiles["good"])
	}
	// Shuffle
	rng.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	return result
}
