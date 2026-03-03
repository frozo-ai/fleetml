package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// APIClient is the HTTP client for the FleetML REST API.
type APIClient struct {
	baseURL    string
	apiKey     string
	token      string
	httpClient *http.Client
}

func NewAPIClient(baseURL, apiKey string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the JWT auth token.
func (c *APIClient) SetToken(token string) {
	c.token = token
}

// Login authenticates and stores the token.
func (c *APIClient) Login(email, password string) error {
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	resp, err := c.do("POST", "/api/v1/auth/login", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	c.token = result.Token
	return nil
}

// HealthCheck checks server health.
func (c *APIClient) HealthCheck() (map[string]interface{}, error) {
	resp, err := c.do("GET", "/api/v1/health", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

// UploadModel uploads a model file.
func (c *APIClient) UploadModel(filePath, name, version, format string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	io.Copy(part, file)

	writer.WriteField("name", name)
	writer.WriteField("version", version)
	writer.WriteField("format", format)
	writer.Close()

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/models", body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

// CreateDeployment creates a new deployment.
func (c *APIClient) CreateDeployment(modelName, modelVersion, targetType, targetID, policy string) (map[string]interface{}, error) {
	body, _ := json.Marshal(map[string]string{
		"model_name":    modelName,
		"model_version": modelVersion,
		"target_type":   targetType,
		"target_id":     targetID,
		"policy":        policy,
	})

	resp, err := c.do("POST", "/api/v1/deployments", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

// ListDevices lists devices.
func (c *APIClient) ListDevices(fleet string) (map[string]interface{}, error) {
	path := "/api/v1/devices"
	if fleet != "" {
		path += "?fleet_id=" + fleet
	}

	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

// GetDeployment gets a deployment by ID.
func (c *APIClient) GetDeployment(id string) (map[string]interface{}, error) {
	resp, err := c.do("GET", "/api/v1/deployments/"+id, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

// RollbackDeployment triggers a rollback for a deployment.
func (c *APIClient) RollbackDeployment(deploymentID string) (map[string]interface{}, error) {
	resp, err := c.do("POST", "/api/v1/deployments/"+deploymentID+"/rollback", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (c *APIClient) do(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}
