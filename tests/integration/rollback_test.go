//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestRollbackFlow tests: deploy v1 -> deploy v2 -> rollback -> verify v1 active.
func TestRollbackFlow(t *testing.T) {
	serverURL := os.Getenv("FLEETML_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}
	apiKey := os.Getenv("FLEETML_API_KEY")

	// 1. Deploy v1
	t.Log("Deploying v1...")
	modelData := []byte("model-v1-data")
	uploadModel(t, serverURL, apiKey, "rollback-test", "1.0", "onnx", modelData)
	deployV1 := createDeployment(t, serverURL, apiKey, "rollback-test", "1.0", "fleet", "default", "immediate")
	waitForDeployment(t, serverURL, apiKey, deployV1, 2*time.Minute)

	// 2. Deploy v2
	t.Log("Deploying v2...")
	modelDataV2 := []byte("model-v2-data")
	uploadModel(t, serverURL, apiKey, "rollback-test", "2.0", "onnx", modelDataV2)
	deployV2 := createDeployment(t, serverURL, apiKey, "rollback-test", "2.0", "fleet", "default", "immediate")
	waitForDeployment(t, serverURL, apiKey, deployV2, 2*time.Minute)

	// 3. Rollback v2
	t.Log("Initiating rollback...")
	rollbackID := rollbackDeployment(t, serverURL, apiKey, deployV2)
	t.Logf("Rollback deployment: %s", rollbackID)

	// 4. Wait for rollback to complete
	waitForDeployment(t, serverURL, apiKey, rollbackID, 2*time.Minute)

	// 5. Verify original deployment was rolled back
	status := getDeploymentStatus(t, serverURL, apiKey, deployV2)
	if status["state"] != "rolled_back" {
		t.Errorf("expected original deployment to be 'rolled_back', got %q", status["state"])
	}

	t.Log("Rollback flow completed successfully")
}

func rollbackDeployment(t *testing.T, serverURL, apiKey, deployID string) string {
	t.Helper()

	req, _ := http.NewRequest("POST",
		fmt.Sprintf("%s/api/v1/deployments/%s/rollback", serverURL, deployID),
		bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("rollback: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return fmt.Sprintf("%v", result["id"])
}

func waitForDeployment(t *testing.T, serverURL, apiKey, deployID string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status := getDeploymentStatus(t, serverURL, apiKey, deployID)
		switch status["state"] {
		case "completed", "rolled_back":
			return
		case "failed":
			t.Fatalf("deployment %s failed: %v", deployID, status["error"])
		}
		time.Sleep(5 * time.Second)
	}
	t.Fatalf("deployment %s timed out after %v", deployID, timeout)
}
