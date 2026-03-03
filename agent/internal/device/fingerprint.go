package device

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type Info struct {
	DeviceID      string            `json:"device_id"`
	Arch          string            `json:"arch"`
	GPUType       string            `json:"gpu_type"`
	Runtime       string            `json:"runtime"`
	RAMMB         int               `json:"ram_mb"`
	DiskGB        int               `json:"disk_gb"`
	OS            string            `json:"os"`
	HardwareModel string            `json:"hardware_model"`
	Labels        map[string]string `json:"labels"`
}

func Fingerprint(deviceID string) (*Info, error) {
	info := &Info{
		DeviceID: deviceID,
		Arch:     runtime.GOARCH,
		Labels:   make(map[string]string),
	}

	// OS info
	hostInfo, err := host.Info()
	if err == nil {
		info.OS = fmt.Sprintf("%s %s", hostInfo.Platform, hostInfo.PlatformVersion)
		info.HardwareModel = hostInfo.KernelArch
	} else {
		info.OS = runtime.GOOS
	}

	// RAM
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		info.RAMMB = int(memInfo.Total / (1024 * 1024))
	}

	// Disk
	diskStat, err := disk.Usage("/")
	if err == nil {
		info.DiskGB = int(diskStat.Total / (1024 * 1024 * 1024))
	}

	// GPU detection
	info.GPUType = detectGPU()
	info.Runtime = selectRuntime(info.GPUType, info.Arch)

	// CPU info for hardware model
	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		info.HardwareModel = cpuInfo[0].ModelName
	}

	return info, nil
}

func detectGPU() string {
	// Check for NVIDIA GPU
	if _, err := os.Stat("/dev/nvidia0"); err == nil {
		return "nvidia"
	}

	// Check for Intel GPU
	if _, err := os.Stat("/dev/dri/renderD128"); err == nil {
		// Read driver info to distinguish Intel from others
		data, err := os.ReadFile("/sys/class/drm/card0/device/vendor")
		if err == nil {
			vendor := strings.TrimSpace(string(data))
			if vendor == "0x8086" {
				return "intel"
			}
		}
	}

	return "none"
}

func selectRuntime(gpuType, arch string) string {
	switch gpuType {
	case "nvidia":
		return "tensorrt"
	case "intel":
		return "openvino"
	}

	if arch == "arm64" || arch == "arm" {
		return "tflite"
	}

	return "onnx"
}
