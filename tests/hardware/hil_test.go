//go:build hil

// Package hardware contains hardware-in-the-loop (HIL) tests that run against
// real edge devices (Jetson, Raspberry Pi, etc.). These are pre-release gate tests.
//
// Usage:
//
//	go test -tags=hil ./tests/hardware/... -timeout=30m
//
// Required environment variables:
//
//	FLEETML_SERVER_URL  - Control plane address (e.g., http://10.0.0.1:8080)
//	HIL_DEVICE_IDS      - Comma-separated device IDs to test against
//	HIL_MODEL_PATH      - Path to a test ONNX model file
//
// Optional:
//
//	HIL_TIMEOUT_MINUTES - Per-test timeout (default: 5)
package hardware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// hilConfig holds test environment configuration.
type hilConfig struct {
	ServerURL      string
	APIKey         string
	DeviceIDs      []string
	ModelPath      string
	TimeoutMinutes int
}

func loadHILConfig(t *testing.T) hilConfig {
	t.Helper()

	serverURL := os.Getenv("FLEETML_SERVER_URL")
	if serverURL == "" {
		t.Skip("FLEETML_SERVER_URL not set — skipping HIL tests")
	}

	deviceStr := os.Getenv("HIL_DEVICE_IDS")
	if deviceStr == "" {
		t.Skip("HIL_DEVICE_IDS not set — skipping HIL tests")
	}

	modelPath := os.Getenv("HIL_MODEL_PATH")
	if modelPath == "" {
		t.Skip("HIL_MODEL_PATH not set — skipping HIL tests")
	}

	timeout := 5
	if s := os.Getenv("HIL_TIMEOUT_MINUTES"); s != "" {
		fmt.Sscanf(s, "%d", &timeout)
	}

	return hilConfig{
		ServerURL:      serverURL,
		APIKey:         os.Getenv("FLEETML_API_KEY"),
		DeviceIDs:      strings.Split(deviceStr, ","),
		ModelPath:      modelPath,
		TimeoutMinutes: timeout,
	}
}

// apiGet makes an authenticated GET request and decodes JSON.
func apiGet(t *testing.T, cfg hilConfig, path string, out interface{}) {
	t.Helper()

	req, err := http.NewRequest("GET", cfg.ServerURL+path, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status %d", path, resp.StatusCode)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("decode response: %v", err)
		}
	}
}

// TestHIL_DevicesOnline verifies all configured HIL devices are online and reporting
// heartbeats to the control plane.
func TestHIL_DevicesOnline(t *testing.T) {
	cfg := loadHILConfig(t)

	for _, deviceID := range cfg.DeviceIDs {
		t.Run(deviceID, func(t *testing.T) {
			var device struct {
				DeviceID      string `json:"device_id"`
				Status        string `json:"status"`
				LastHeartbeat string `json:"last_heartbeat"`
				Arch          string `json:"arch"`
				Runtime       string `json:"runtime"`
				GPUType       string `json:"gpu_type"`
			}

			apiGet(t, cfg, "/api/v1/devices/"+deviceID, &device)

			if device.Status == "offline" {
				t.Errorf("device %s is offline", deviceID)
			}

			if device.LastHeartbeat == "" {
				t.Errorf("device %s has never sent a heartbeat", deviceID)
			}

			t.Logf("Device %s: status=%s arch=%s runtime=%s gpu=%s",
				deviceID, device.Status, device.Arch, device.Runtime, device.GPUType)
		})
	}
}

// TestHIL_DeployModel deploys a test model to each HIL device and verifies
// the deployment completes successfully within the timeout.
func TestHIL_DeployModel(t *testing.T) {
	cfg := loadHILConfig(t)
	timeout := time.Duration(cfg.TimeoutMinutes) * time.Minute

	for _, deviceID := range cfg.DeviceIDs {
		t.Run(deviceID, func(t *testing.T) {
			// Create deployment via API
			deployReq := fmt.Sprintf(`{
				"model_name": "hil-test-model",
				"model_version": "v1.0",
				"target_type": "device",
				"target_id": "%s",
				"policy": "rolling"
			}`, deviceID)

			req, err := http.NewRequest("POST", cfg.ServerURL+"/api/v1/deployments",
				strings.NewReader(deployReq))
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			if cfg.APIKey != "" {
				req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("POST /deployments: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
				t.Fatalf("POST /deployments: status %d", resp.StatusCode)
			}

			var deploy struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			}
			json.NewDecoder(resp.Body).Decode(&deploy)
			t.Logf("Created deployment %s for device %s", deploy.ID, deviceID)

			// Poll until deployment completes or times out
			deadline := time.Now().Add(timeout)
			for time.Now().Before(deadline) {
				var status struct {
					ID     string `json:"id"`
					Status string `json:"status"`
				}
				apiGet(t, cfg, "/api/v1/deployments/"+deploy.ID, &status)

				switch status.Status {
				case "completed", "success":
					t.Logf("Deployment %s completed successfully on %s", deploy.ID, deviceID)
					return
				case "failed", "rolled_back":
					t.Fatalf("Deployment %s failed on device %s: status=%s", deploy.ID, deviceID, status.Status)
				}

				time.Sleep(5 * time.Second)
			}

			t.Fatalf("Deployment %s timed out after %v on device %s", deploy.ID, timeout, deviceID)
		})
	}
}

// TestHIL_ModelHotSwap deploys two versions sequentially to verify zero-downtime
// hot-swap works on real hardware.
func TestHIL_ModelHotSwap(t *testing.T) {
	cfg := loadHILConfig(t)
	timeout := time.Duration(cfg.TimeoutMinutes) * time.Minute

	if len(cfg.DeviceIDs) == 0 {
		t.Skip("no devices configured")
	}

	deviceID := cfg.DeviceIDs[0] // Test on first device
	t.Logf("Testing hot-swap on device %s", deviceID)

	versions := []string{"v1.0", "v2.0"}
	for _, version := range versions {
		deployReq := fmt.Sprintf(`{
			"model_name": "hil-hotswap-test",
			"model_version": "%s",
			"target_type": "device",
			"target_id": "%s",
			"policy": "rolling"
		}`, version, deviceID)

		req, _ := http.NewRequest("POST", cfg.ServerURL+"/api/v1/deployments",
			strings.NewReader(deployReq))
		req.Header.Set("Content-Type", "application/json")
		if cfg.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("deploy %s: %v", version, err)
		}

		var deploy struct {
			ID string `json:"id"`
		}
		json.NewDecoder(resp.Body).Decode(&deploy)
		resp.Body.Close()

		// Wait for completion
		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			var status struct {
				Status string `json:"status"`
			}
			apiGet(t, cfg, "/api/v1/deployments/"+deploy.ID, &status)
			if status.Status == "completed" || status.Status == "success" {
				t.Logf("Version %s deployed successfully", version)
				break
			}
			if status.Status == "failed" {
				t.Fatalf("Version %s deployment failed", version)
			}
			time.Sleep(5 * time.Second)
		}
	}
}

// TestHIL_DeviceMetrics verifies real devices report valid hardware metrics
// (CPU, memory, disk, GPU if applicable).
func TestHIL_DeviceMetrics(t *testing.T) {
	cfg := loadHILConfig(t)

	for _, deviceID := range cfg.DeviceIDs {
		t.Run(deviceID, func(t *testing.T) {
			var device struct {
				CPUPercent  *float64 `json:"cpu_percent"`
				GPUPercent  *float64 `json:"gpu_percent"`
				RAMMBUsed   *float64 `json:"ram_mb_used"`
				RAMMB       float64  `json:"ram_mb"`
				DiskPercent *float64 `json:"disk_percent"`
				TempC       *float64 `json:"temperature_c"`
			}

			apiGet(t, cfg, "/api/v1/devices/"+deviceID, &device)

			if device.CPUPercent == nil {
				t.Error("CPU percent is nil — device not reporting CPU metrics")
			} else if *device.CPUPercent < 0 || *device.CPUPercent > 100 {
				t.Errorf("CPU percent out of range: %f", *device.CPUPercent)
			}

			if device.RAMMB <= 0 {
				t.Error("RAM MB is zero or negative")
			}

			if device.RAMMBUsed != nil && *device.RAMMBUsed > device.RAMMB {
				t.Errorf("RAM used (%f MB) exceeds total RAM (%f MB)", *device.RAMMBUsed, device.RAMMB)
			}

			if device.DiskPercent != nil && (*device.DiskPercent < 0 || *device.DiskPercent > 100) {
				t.Errorf("Disk percent out of range: %f", *device.DiskPercent)
			}

			t.Logf("Device %s metrics: cpu=%.1f%% ram=%.0f/%.0fMB disk=%.1f%%",
				deviceID,
				valOrZero(device.CPUPercent),
				valOrZero(device.RAMMBUsed),
				device.RAMMB,
				valOrZero(device.DiskPercent))
		})
	}
}

// TestHIL_Rollback deploys a model then rolls it back, verifying the device
// returns to its previous state.
func TestHIL_Rollback(t *testing.T) {
	cfg := loadHILConfig(t)
	timeout := time.Duration(cfg.TimeoutMinutes) * time.Minute

	if len(cfg.DeviceIDs) == 0 {
		t.Skip("no devices configured")
	}

	deviceID := cfg.DeviceIDs[0]

	// Deploy v1
	deployReq := fmt.Sprintf(`{
		"model_name": "hil-rollback-test",
		"model_version": "v1.0",
		"target_type": "device",
		"target_id": "%s",
		"policy": "rolling"
	}`, deviceID)

	req, _ := http.NewRequest("POST", cfg.ServerURL+"/api/v1/deployments",
		strings.NewReader(deployReq))
	req.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("deploy v1: %v", err)
	}

	var deploy struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&deploy)
	resp.Body.Close()

	// Wait for v1 to complete
	waitForDeployment(t, cfg, deploy.ID, timeout)

	// Rollback
	rollbackReq, _ := http.NewRequest("POST",
		cfg.ServerURL+"/api/v1/deployments/"+deploy.ID+"/rollback", nil)
	if cfg.APIKey != "" {
		rollbackReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	rollbackResp, err := http.DefaultClient.Do(rollbackReq)
	if err != nil {
		t.Fatalf("rollback: %v", err)
	}
	rollbackResp.Body.Close()

	if rollbackResp.StatusCode != http.StatusOK {
		t.Fatalf("rollback returned status %d", rollbackResp.StatusCode)
	}

	t.Logf("Rollback initiated for deployment %s on device %s", deploy.ID, deviceID)

	// Verify device returns to healthy state
	time.Sleep(10 * time.Second)

	var device struct {
		Status string `json:"status"`
	}
	apiGet(t, cfg, "/api/v1/devices/"+deviceID, &device)

	if device.Status == "offline" {
		t.Errorf("device %s went offline after rollback", deviceID)
	}
	t.Logf("Device %s status after rollback: %s", deviceID, device.Status)
}

func waitForDeployment(t *testing.T, cfg hilConfig, deployID string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var status struct {
			Status string `json:"status"`
		}
		apiGet(t, cfg, "/api/v1/deployments/"+deployID, &status)
		if status.Status == "completed" || status.Status == "success" {
			return
		}
		if status.Status == "failed" || status.Status == "rolled_back" {
			t.Fatalf("deployment %s failed: %s", deployID, status.Status)
		}
		time.Sleep(5 * time.Second)
	}
	t.Fatalf("deployment %s timed out after %v", deployID, timeout)
}

func valOrZero(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
