package simulator

// DeviceProfile defines a virtual device's characteristics for simulation.
type DeviceProfile struct {
	Name          string  `json:"name"`
	Arch          string  `json:"arch"`   // amd64, arm64, armv7
	GPUType       string  `json:"gpu_type"`
	Runtime       string  `json:"runtime"`
	RAMMB         int     `json:"ram_mb"`
	DiskGB        int     `json:"disk_gb"`
	OS            string  `json:"os"`
	HardwareModel string  `json:"hardware_model"`
	CPUCores      int     `json:"cpu_cores"`
	InferenceMs   float64 `json:"inference_ms"` // Simulated inference latency
}

// NetworkProfile defines simulated network conditions.
type NetworkProfile struct {
	Name        string  `json:"name"`
	LatencyMs   int     `json:"latency_ms"`
	JitterMs    int     `json:"jitter_ms"`
	PacketLoss  float64 `json:"packet_loss"` // 0.0 - 1.0
	BandwidthKB int     `json:"bandwidth_kb"`
}

// PredefinedProfiles provides common device configurations.
var PredefinedProfiles = map[string]DeviceProfile{
	"jetson-nano": {
		Name:          "jetson-nano",
		Arch:          "arm64",
		GPUType:       "nvidia-tegra",
		Runtime:       "tensorrt",
		RAMMB:         4096,
		DiskGB:        32,
		OS:            "linux",
		HardwareModel: "NVIDIA Jetson Nano",
		CPUCores:      4,
		InferenceMs:   15,
	},
	"jetson-orin": {
		Name:          "jetson-orin",
		Arch:          "arm64",
		GPUType:       "nvidia-ampere",
		Runtime:       "tensorrt",
		RAMMB:         16384,
		DiskGB:        64,
		OS:            "linux",
		HardwareModel: "NVIDIA Jetson Orin",
		CPUCores:      12,
		InferenceMs:   5,
	},
	"rpi4": {
		Name:          "rpi4",
		Arch:          "arm64",
		GPUType:       "",
		Runtime:       "tflite",
		RAMMB:         4096,
		DiskGB:        32,
		OS:            "linux",
		HardwareModel: "Raspberry Pi 4 Model B",
		CPUCores:      4,
		InferenceMs:   50,
	},
	"rpi5": {
		Name:          "rpi5",
		Arch:          "arm64",
		GPUType:       "",
		Runtime:       "tflite",
		RAMMB:         8192,
		DiskGB:        64,
		OS:            "linux",
		HardwareModel: "Raspberry Pi 5",
		CPUCores:      4,
		InferenceMs:   30,
	},
	"intel-nuc": {
		Name:          "intel-nuc",
		Arch:          "amd64",
		GPUType:       "intel-uhd",
		Runtime:       "openvino",
		RAMMB:         16384,
		DiskGB:        256,
		OS:            "linux",
		HardwareModel: "Intel NUC 13 Pro",
		CPUCores:      8,
		InferenceMs:   10,
	},
	"generic-x86": {
		Name:          "generic-x86",
		Arch:          "amd64",
		GPUType:       "",
		Runtime:       "onnx",
		RAMMB:         8192,
		DiskGB:        128,
		OS:            "linux",
		HardwareModel: "Generic x86",
		CPUCores:      4,
		InferenceMs:   25,
	},
	"hailo-8": {
		Name:          "hailo-8",
		Arch:          "arm64",
		GPUType:       "hailo-8",
		Runtime:       "hailo",
		RAMMB:         4096,
		DiskGB:        32,
		OS:            "linux",
		HardwareModel: "Hailo-8 Accelerator",
		CPUCores:      4,
		InferenceMs:   8,
	},
	"qualcomm-rb5": {
		Name:          "qualcomm-rb5",
		Arch:          "arm64",
		GPUType:       "adreno-650",
		Runtime:       "snpe",
		RAMMB:         8192,
		DiskGB:        128,
		OS:            "linux",
		HardwareModel: "Qualcomm RB5",
		CPUCores:      8,
		InferenceMs:   12,
	},
}

// PredefinedNetworkProfiles provides common network conditions.
var PredefinedNetworkProfiles = map[string]NetworkProfile{
	"excellent": {
		Name:        "excellent",
		LatencyMs:   5,
		JitterMs:    1,
		PacketLoss:  0.0,
		BandwidthKB: 100000,
	},
	"good": {
		Name:        "good",
		LatencyMs:   20,
		JitterMs:    5,
		PacketLoss:  0.001,
		BandwidthKB: 10000,
	},
	"cellular": {
		Name:        "cellular",
		LatencyMs:   100,
		JitterMs:    30,
		PacketLoss:  0.02,
		BandwidthKB: 2000,
	},
	"poor": {
		Name:        "poor",
		LatencyMs:   500,
		JitterMs:    100,
		PacketLoss:  0.05,
		BandwidthKB: 500,
	},
	"satellite": {
		Name:        "satellite",
		LatencyMs:   600,
		JitterMs:    50,
		PacketLoss:  0.03,
		BandwidthKB: 1000,
	},
	"offline": {
		Name:        "offline",
		LatencyMs:   0,
		JitterMs:    0,
		PacketLoss:  1.0,
		BandwidthKB: 0,
	},
}
