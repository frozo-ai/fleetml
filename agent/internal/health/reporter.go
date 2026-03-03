package health

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemMetrics contains current system resource usage.
type SystemMetrics struct {
	CPUPercent   float64 `json:"cpu_percent"`
	GPUPercent   float64 `json:"gpu_percent"`
	RAMMBUsed    int     `json:"ram_mb_used"`
	DiskPercent  float64 `json:"disk_percent"`
	TemperatureC float64 `json:"temperature_c"`
	UptimeHours  float64 `json:"uptime_hours"`
}

// Reporter collects system health metrics.
type Reporter struct {
	interval time.Duration
}

func NewReporter(interval time.Duration) *Reporter {
	return &Reporter{interval: interval}
}

// Collect gathers current system metrics.
func (r *Reporter) Collect() (*SystemMetrics, error) {
	metrics := &SystemMetrics{}

	// CPU
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		metrics.CPUPercent = cpuPercent[0]
	}

	// RAM
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		metrics.RAMMBUsed = int(memInfo.Used / (1024 * 1024))
	}

	// Disk
	diskStat, err := disk.Usage("/")
	if err == nil {
		metrics.DiskPercent = diskStat.UsedPercent
	}

	// Uptime
	uptime, err := host.Uptime()
	if err == nil {
		metrics.UptimeHours = float64(uptime) / 3600.0
	}

	// Temperature (best effort)
	temps, err := host.SensorsTemperatures()
	if err == nil && len(temps) > 0 {
		metrics.TemperatureC = temps[0].Temperature
	}

	// GPU is 0 by default (detected separately if nvidia/intel present)

	return metrics, nil
}

// Interval returns the configured collection interval.
func (r *Reporter) Interval() time.Duration {
	return r.interval
}
