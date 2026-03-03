//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestFullDeploymentFlow tests: upload model -> create deployment -> wait for completion.
func TestFullDeploymentFlow(t *testing.T) {
	serverURL := os.Getenv("FLEETML_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}
	apiKey := os.Getenv("FLEETML_API_KEY")

	// 1. Upload a model
	t.Log("Uploading test model...")
	modelData := []byte("fake-onnx-model-binary-data")
	modelID := uploadModel(t, serverURL, apiKey, "test-deploy", "1.0", "onnx", modelData)
	t.Logf("Model uploaded: %s", modelID)

	// 2. Create deployment
	t.Log("Creating deployment...")
	deployID := createDeployment(t, serverURL, apiKey, "test-deploy", "1.0", "fleet", "default", "immediate")
	t.Logf("Deployment created: %s", deployID)

	// 3. Poll for completion
	t.Log("Waiting for deployment completion...")
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		status := getDeploymentStatus(t, serverURL, apiKey, deployID)
		t.Logf("  State: %s, Completed: %v/%v", status["state"], status["completed_devices"], status["total_devices"])

		switch status["state"] {
		case "completed":
			t.Log("Deployment completed successfully!")
			return
		case "failed":
			t.Fatalf("Deployment failed: %v", status["error"])
		}
		time.Sleep(5 * time.Second)
	}
	t.Fatal("Deployment timed out")
}

func uploadModel(t *testing.T, serverURL, apiKey, name, version, format string, data []byte) string {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("name", name)
	writer.WriteField("version", version)
	writer.WriteField("format", format)

	part, _ := writer.CreateFormFile("file", name+".onnx")
	part.Write(data)
	writer.Close()

	req, _ := http.NewRequest("POST", serverURL+"/api/v1/models", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload model: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload model: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return fmt.Sprintf("%v", result["id"])
}

func createDeployment(t *testing.T, serverURL, apiKey, modelName, modelVersion, targetType, targetID, policy string) string {
	t.Helper()

	payload := map[string]string{
		"model_name":    modelName,
		"model_version": modelVersion,
		"target_type":   targetType,
		"target_id":     targetID,
		"policy":        policy,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", serverURL+"/api/v1/deployments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create deployment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("create deployment: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return fmt.Sprintf("%v", result["id"])
}

func getDeploymentStatus(t *testing.T, serverURL, apiKey, deployID string) map[string]interface{} {
	t.Helper()

	req, _ := http.NewRequest("GET", serverURL+"/api/v1/deployments/"+deployID, nil)
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}
