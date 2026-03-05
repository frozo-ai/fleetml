package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MLflowClient communicates with an MLflow Tracking Server.
type MLflowClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewMLflowClient creates a new MLflow client.
func NewMLflowClient(baseURL string) *MLflowClient {
	return &MLflowClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MLflowModel represents a registered model in MLflow.
type MLflowModel struct {
	Name              string `json:"name"`
	LatestVersions    []MLflowModelVersion `json:"latest_versions"`
}

// MLflowModelVersion represents a model version in MLflow.
type MLflowModelVersion struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	Source          string `json:"source"`
	RunID           string `json:"run_id"`
	Status          string `json:"status"`
	Description     string `json:"description"`
	CreationTimestamp int64 `json:"creation_timestamp"`
}

// MLflowArtifact represents a downloadable artifact.
type MLflowArtifact struct {
	Path     string `json:"path"`
	IsDir    bool   `json:"is_dir"`
	FileSize int64  `json:"file_size"`
}

// MLflowImportRequest defines what to import from MLflow.
type MLflowImportRequest struct {
	ModelName    string            `json:"model_name"`
	Version      string            `json:"version,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	Description  string            `json:"description,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// MLflowImportResult contains the result of an import.
type MLflowImportResult struct {
	ModelID      string `json:"model_id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Format       string `json:"format"`
	ArtifactURL  string `json:"artifact_url"`
	ArtifactSize int64  `json:"artifact_size"`
	Source       string `json:"source"`
}

// GetModel retrieves a registered model from MLflow.
func (c *MLflowClient) GetModel(ctx context.Context, name string) (*MLflowModel, error) {
	url := fmt.Sprintf("%s/api/2.0/mlflow/registered-models/get?name=%s", c.baseURL, name)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mlflow request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mlflow error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		RegisteredModel MLflowModel `json:"registered_model"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode mlflow response: %w", err)
	}

	return &result.RegisteredModel, nil
}

// GetModelVersion retrieves a specific model version from MLflow.
func (c *MLflowClient) GetModelVersion(ctx context.Context, name, version string) (*MLflowModelVersion, error) {
	url := fmt.Sprintf("%s/api/2.0/mlflow/model-versions/get?name=%s&version=%s", c.baseURL, name, version)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mlflow request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mlflow error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ModelVersion MLflowModelVersion `json:"model_version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode mlflow response: %w", err)
	}

	return &result.ModelVersion, nil
}

// DownloadArtifact downloads a model artifact from MLflow.
func (c *MLflowClient) DownloadArtifact(ctx context.Context, runID, artifactPath string) (io.ReadCloser, int64, error) {
	url := fmt.Sprintf("%s/api/2.0/mlflow/artifacts/get?run_id=%s&path=%s", c.baseURL, runID, artifactPath)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("mlflow download failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, 0, fmt.Errorf("mlflow download error %d: %s", resp.StatusCode, string(body))
	}

	return resp.Body, resp.ContentLength, nil
}

// ListArtifacts lists artifacts for a run.
func (c *MLflowClient) ListArtifacts(ctx context.Context, runID, path string) ([]MLflowArtifact, error) {
	url := fmt.Sprintf("%s/api/2.0/mlflow/artifacts/list?run_id=%s", c.baseURL, runID)
	if path != "" {
		url += "&path=" + path
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mlflow request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mlflow error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Files []MLflowArtifact `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode mlflow response: %w", err)
	}

	return result.Files, nil
}

// Health checks if the MLflow server is reachable.
func (c *MLflowClient) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mlflow not reachable: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// DetectFormat inspects artifact list and returns the model format.
func DetectFormat(artifacts []MLflowArtifact) (string, string) {
	for _, a := range artifacts {
		switch {
		case !a.IsDir && hasExtension(a.Path, ".onnx"):
			return "onnx", a.Path
		case !a.IsDir && hasExtension(a.Path, ".pt") || hasExtension(a.Path, ".pth"):
			return "pytorch", a.Path
		case !a.IsDir && hasExtension(a.Path, ".pb"):
			return "tensorflow", a.Path
		case !a.IsDir && hasExtension(a.Path, ".tflite"):
			return "tflite", a.Path
		}
	}
	// Default to ONNX directory convention
	return "onnx", "model.onnx"
}

func hasExtension(path, ext string) bool {
	if len(path) < len(ext) {
		return false
	}
	return path[len(path)-len(ext):] == ext
}
