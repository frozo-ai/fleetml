package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

// ONNXSubprocessRuntime implements the Runtime interface by shelling out to
// an ONNX Runtime helper process. This avoids CGO for cross-compilation.
type ONNXSubprocessRuntime struct {
	mu        sync.RWMutex
	modelPath string
	loaded    bool
	// metadata holds info parsed from the model (input/output shapes etc.)
	metadata *onnxModelMetadata
}

type onnxModelMetadata struct {
	InputNames  []string `json:"input_names"`
	OutputNames []string `json:"output_names"`
	OpsetVersion int     `json:"opset_version"`
}

func NewONNXSubprocessRuntime() *ONNXSubprocessRuntime {
	return &ONNXSubprocessRuntime{}
}

func (r *ONNXSubprocessRuntime) Name() string {
	return "onnx"
}

func (r *ONNXSubprocessRuntime) Load(modelPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verify the file exists
	info, err := os.Stat(modelPath)
	if err != nil {
		return fmt.Errorf("model file not accessible: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("model path is a directory, expected a file")
	}

	// Try to validate the model via subprocess if onnxruntime_test is available
	if helperPath, err := exec.LookPath("onnx_validate"); err == nil {
		cmd := exec.Command(helperPath, "--validate", modelPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("model validation failed: %s", string(out))
		}
		// Parse metadata from validation output
		var meta onnxModelMetadata
		if jsonErr := json.Unmarshal(out, &meta); jsonErr == nil {
			r.metadata = &meta
		}
	}

	r.modelPath = modelPath
	r.loaded = true
	return nil
}

func (r *ONNXSubprocessRuntime) Infer(input []byte) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("model not loaded")
	}

	// Look for the onnx_infer helper binary
	helperPath, err := exec.LookPath("onnx_infer")
	if err != nil {
		return nil, fmt.Errorf("onnx_infer helper not found on PATH: install onnxruntime and ensure agent/tools is on PATH")
	}

	cmd := exec.Command(helperPath, "--model", r.modelPath)
	cmd.Stdin = bytes.NewReader(input)

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("inference failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("inference subprocess error: %w", err)
	}

	return out, nil
}

func (r *ONNXSubprocessRuntime) Unload() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.loaded = false
	r.modelPath = ""
	r.metadata = nil
	return nil
}

func (r *ONNXSubprocessRuntime) IsSupported() bool {
	// ONNX Runtime is supported on all platforms
	return true
}

// ModelPath returns the current model path (empty if not loaded).
func (r *ONNXSubprocessRuntime) ModelPath() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.modelPath
}

// IsLoaded returns whether a model is currently loaded.
func (r *ONNXSubprocessRuntime) IsLoaded() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.loaded
}

// Metadata returns model metadata if available.
func (r *ONNXSubprocessRuntime) Metadata() *onnxModelMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.metadata
}
