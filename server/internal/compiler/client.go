package compiler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for the FleetML Compiler Service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new compiler service client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute, // Compilation can be slow
		},
	}
}

// CompileRequest is the request body for the /compile endpoint.
type CompileRequest struct {
	ModelURL      string         `json:"model_url"`
	ModelID       string         `json:"model_id"`
	TargetRuntime string         `json:"target_runtime"`
	Options       map[string]any `json:"options,omitempty"`
}

// CompileResponse is the response from the /compile endpoint.
type CompileResponse struct {
	Runtime            string         `json:"runtime"`
	ArtifactURL        string         `json:"artifact_url"`
	Checksum           string         `json:"checksum"`
	FileSize           int64          `json:"file_size"`
	CompileTimeSeconds float64        `json:"compile_time_seconds"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

// Compile sends a compilation request to the compiler service.
func (c *Client) Compile(ctx context.Context, req CompileRequest) (*CompileResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/compile", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("compiler service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("compiler service returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result CompileResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// Health checks if the compiler service is healthy.
func (c *Client) Health(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("compiler service health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("compiler service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
