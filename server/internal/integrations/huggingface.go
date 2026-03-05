package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	hfAPIBase = "https://huggingface.co/api"
	hfCDNBase = "https://huggingface.co"
)

// HuggingFaceClient communicates with the HuggingFace Hub API.
type HuggingFaceClient struct {
	token      string
	httpClient *http.Client
}

// NewHuggingFaceClient creates a new HuggingFace client.
func NewHuggingFaceClient(token string) *HuggingFaceClient {
	return &HuggingFaceClient{
		token: token,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Models can be large
		},
	}
}

// HFModelInfo represents metadata about a HuggingFace model.
type HFModelInfo struct {
	ID          string   `json:"id"`
	ModelID     string   `json:"modelId"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	Pipeline    string   `json:"pipeline_tag"`
	LibraryName string   `json:"library_name"`
	Downloads   int      `json:"downloads"`
	Likes       int      `json:"likes"`
}

// HFFileSibling represents a file in a HF model repo.
type HFFileSibling struct {
	Filename string `json:"rfilename"`
	Size     int64  `json:"size,omitempty"`
}

// HFImportRequest defines what to import from HuggingFace.
type HFImportRequest struct {
	RepoID      string            `json:"repo_id"`
	Revision    string            `json:"revision,omitempty"`
	Filename    string            `json:"filename,omitempty"`
	Name        string            `json:"name,omitempty"`
	Version     string            `json:"version,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// HFImportResult contains the result of a HuggingFace import.
type HFImportResult struct {
	ModelID      string `json:"model_id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Format       string `json:"format"`
	ArtifactURL  string `json:"artifact_url"`
	ArtifactSize int64  `json:"artifact_size"`
	Source       string `json:"source"`
}

// GetModelInfo retrieves model metadata from HuggingFace Hub.
func (c *HuggingFaceClient) GetModelInfo(ctx context.Context, repoID string) (*HFModelInfo, error) {
	url := fmt.Sprintf("%s/models/%s", hfAPIBase, repoID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("huggingface request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("model %q not found on HuggingFace Hub", repoID)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("huggingface error %d: %s", resp.StatusCode, string(body))
	}

	var info HFModelInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode huggingface response: %w", err)
	}

	return &info, nil
}

// ListFiles lists files in a HuggingFace model repo.
func (c *HuggingFaceClient) ListFiles(ctx context.Context, repoID, revision string) ([]HFFileSibling, error) {
	url := fmt.Sprintf("%s/models/%s", hfAPIBase, repoID)
	if revision != "" {
		url += "?revision=" + revision
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("huggingface request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("huggingface error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Siblings []HFFileSibling `json:"siblings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode huggingface response: %w", err)
	}

	return result.Siblings, nil
}

// DownloadFile downloads a file from a HuggingFace model repo.
func (c *HuggingFaceClient) DownloadFile(ctx context.Context, repoID, filename, revision string) (io.ReadCloser, int64, error) {
	rev := "main"
	if revision != "" {
		rev = revision
	}

	url := fmt.Sprintf("%s/%s/resolve/%s/%s", hfCDNBase, repoID, rev, filename)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, err
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("huggingface download failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, 0, fmt.Errorf("huggingface download error %d: %s", resp.StatusCode, string(body))
	}

	return resp.Body, resp.ContentLength, nil
}

// FindONNXFile finds the best ONNX file in a model repo.
func FindONNXFile(siblings []HFFileSibling) (string, bool) {
	// Prefer model.onnx, then any .onnx file
	for _, s := range siblings {
		if s.Filename == "model.onnx" {
			return s.Filename, true
		}
	}
	for _, s := range siblings {
		if hasExtension(s.Filename, ".onnx") {
			return s.Filename, true
		}
	}
	return "", false
}

// DetectHFFormat determines the model format from repo files.
func DetectHFFormat(siblings []HFFileSibling) string {
	for _, s := range siblings {
		if hasExtension(s.Filename, ".onnx") {
			return "onnx"
		}
	}
	for _, s := range siblings {
		if hasExtension(s.Filename, ".pt") || hasExtension(s.Filename, ".pth") || hasExtension(s.Filename, ".bin") {
			return "pytorch"
		}
	}
	for _, s := range siblings {
		if hasExtension(s.Filename, ".tflite") {
			return "tflite"
		}
	}
	return "unknown"
}

func (c *HuggingFaceClient) setAuth(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}
