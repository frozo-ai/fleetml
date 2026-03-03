package model

import "time"

// Runtime defines the interface for ML model inference engines.
type Runtime interface {
	Name() string
	Load(modelPath string) error
	Infer(input []byte) ([]byte, error)
	Unload() error
	IsSupported() bool
}

// Model represents a loaded ML model.
type Model struct {
	Name      string
	Version   string
	Format    string // onnx, tensorrt, openvino, tflite, snpe
	FilePath  string
	SizeBytes int64
	Checksum  string
	LoadedAt  time.Time
}

// LoadedModel wraps a Model with its Runtime for inference.
type LoadedModel struct {
	Model   *Model
	Runtime Runtime
}

// ONNXRuntime implements the Runtime interface for ONNX models.
// This is the default runtime for all platforms in MVP.
type ONNXRuntime struct {
	modelPath string
	loaded    bool
}

func NewONNXRuntime() *ONNXRuntime {
	return &ONNXRuntime{}
}

func (r *ONNXRuntime) Name() string {
	return "onnx"
}

func (r *ONNXRuntime) Load(modelPath string) error {
	r.modelPath = modelPath
	r.loaded = true
	return nil
}

func (r *ONNXRuntime) Infer(input []byte) ([]byte, error) {
	// TODO: Implement ONNX Runtime inference via CGO or subprocess
	return input, nil
}

func (r *ONNXRuntime) Unload() error {
	r.loaded = false
	r.modelPath = ""
	return nil
}

func (r *ONNXRuntime) IsSupported() bool {
	return true // ONNX is supported on all platforms
}
