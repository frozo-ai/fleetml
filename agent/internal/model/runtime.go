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
// Delegates to ONNXSubprocessRuntime for actual inference via onnx_infer helper.
// Falls back to pass-through when the helper binary is not available (dev/test mode).
type ONNXRuntime struct {
	delegate *ONNXSubprocessRuntime
}

func NewONNXRuntime() *ONNXRuntime {
	return &ONNXRuntime{
		delegate: NewONNXSubprocessRuntime(),
	}
}

func (r *ONNXRuntime) Name() string {
	return "onnx"
}

func (r *ONNXRuntime) Load(modelPath string) error {
	return r.delegate.Load(modelPath)
}

func (r *ONNXRuntime) Infer(input []byte) ([]byte, error) {
	return r.delegate.Infer(input)
}

func (r *ONNXRuntime) Unload() error {
	return r.delegate.Unload()
}

func (r *ONNXRuntime) IsSupported() bool {
	return true // ONNX is supported on all platforms
}
