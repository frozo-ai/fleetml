package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAPIClient(t *testing.T) {
	c := NewAPIClient("http://localhost:8080", "test-key")
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("expected baseURL http://localhost:8080, got %s", c.baseURL)
	}
	if c.apiKey != "test-key" {
		t.Errorf("expected apiKey test-key, got %s", c.apiKey)
	}
}

func TestSetToken(t *testing.T) {
	c := NewAPIClient("http://localhost:8080", "")
	c.SetToken("jwt-token-123")
	if c.token != "jwt-token-123" {
		t.Errorf("expected token jwt-token-123, got %s", c.token)
	}
}

func TestHealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/health" {
			t.Errorf("expected /api/v1/health, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
			"uptime": 3600,
		})
	}))
	defer server.Close()

	c := NewAPIClient(server.URL, "")
	result, err := c.HealthCheck()
	if err != nil {
		t.Fatalf("health check: %v", err)
	}
	if result["status"] != "healthy" {
		t.Errorf("expected status healthy, got %v", result["status"])
	}
}

func TestListDevices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/devices" {
			t.Errorf("expected /api/v1/devices, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"devices": []interface{}{},
			"total":   0,
		})
	}))
	defer server.Close()

	c := NewAPIClient(server.URL, "")
	result, err := c.ListDevices("")
	if err != nil {
		t.Fatalf("list devices: %v", err)
	}
	if result["total"] != float64(0) {
		t.Errorf("expected 0 devices, got %v", result["total"])
	}
}

func TestListDevices_WithFleetFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fleet_id") != "fleet-123" {
			t.Errorf("expected fleet_id=fleet-123, got %s", r.URL.Query().Get("fleet_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"devices": []interface{}{}, "total": 0})
	}))
	defer server.Close()

	c := NewAPIClient(server.URL, "")
	_, err := c.ListDevices("fleet-123")
	if err != nil {
		t.Fatalf("list devices with fleet: %v", err)
	}
}

func TestGetDeployment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/deployments/deploy-123" {
			t.Errorf("expected /api/v1/deployments/deploy-123, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    "deploy-123",
			"state": "rolling_out",
		})
	}))
	defer server.Close()

	c := NewAPIClient(server.URL, "")
	result, err := c.GetDeployment("deploy-123")
	if err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	if result["state"] != "rolling_out" {
		t.Errorf("expected state rolling_out, got %v", result["state"])
	}
}

func TestGetDeviceLogs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/devices/device-001/logs" {
			t.Errorf("expected device logs path, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "50" {
			t.Errorf("expected limit=50, got %s", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("since") != "1h" {
			t.Errorf("expected since=1h, got %s", r.URL.Query().Get("since"))
		}
		if r.URL.Query().Get("level") != "warn" {
			t.Errorf("expected level=warn, got %s", r.URL.Query().Get("level"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"logs": []interface{}{
				map[string]interface{}{
					"timestamp": "2026-03-03T10:00:00Z",
					"level":     "warn",
					"message":   "high CPU usage",
				},
			},
		})
	}))
	defer server.Close()

	c := NewAPIClient(server.URL, "")
	logs, err := c.GetDeviceLogs("device-001", "1h", "warn", 50)
	if err != nil {
		t.Fatalf("get device logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(logs))
	}
	if logs[0]["level"] != "warn" {
		t.Errorf("expected level warn, got %v", logs[0]["level"])
	}
}

func TestGetDeviceLogs_NoFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check no since or level params when empty
		if r.URL.Query().Get("since") != "" {
			t.Errorf("expected no since param, got %s", r.URL.Query().Get("since"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"logs": []interface{}{}})
	}))
	defer server.Close()

	c := NewAPIClient(server.URL, "")
	logs, err := c.GetDeviceLogs("device-001", "", "", 100)
	if err != nil {
		t.Fatalf("get device logs: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs, got %d", len(logs))
	}
}

func TestAuthorizationHeader_Token(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer jwt-token" {
			t.Errorf("expected Authorization: Bearer jwt-token, got %s", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer server.Close()

	c := NewAPIClient(server.URL, "api-key")
	c.SetToken("jwt-token")
	_, err := c.HealthCheck()
	if err != nil {
		t.Fatalf("health check: %v", err)
	}
}

func TestAuthorizationHeader_APIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer api-key-123" {
			t.Errorf("expected Authorization: Bearer api-key-123, got %s", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer server.Close()

	c := NewAPIClient(server.URL, "api-key-123")
	_, err := c.HealthCheck()
	if err != nil {
		t.Fatalf("health check: %v", err)
	}
}

func TestServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	c := NewAPIClient(server.URL, "")
	_, err := c.HealthCheck()
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestCreateDeployment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/deployments" {
			t.Errorf("expected /api/v1/deployments, got %s", r.URL.Path)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["model_name"] != "mobilenet" {
			t.Errorf("expected model_name mobilenet, got %s", body["model_name"])
		}
		if body["policy"] != "canary" {
			t.Errorf("expected policy canary, got %s", body["policy"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    "deploy-new-123",
			"state": "rolling_out",
		})
	}))
	defer server.Close()

	c := NewAPIClient(server.URL, "")
	result, err := c.CreateDeployment("mobilenet", "v1", "fleet", "default", "canary")
	if err != nil {
		t.Fatalf("create deployment: %v", err)
	}
	if result["id"] != "deploy-new-123" {
		t.Errorf("expected id deploy-new-123, got %v", result["id"])
	}
}

func TestRollbackDeployment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/deployments/deploy-123/rollback" {
			t.Errorf("expected rollback path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    "rollback-456",
			"state": "rolling_out",
		})
	}))
	defer server.Close()

	c := NewAPIClient(server.URL, "")
	result, err := c.RollbackDeployment("deploy-123")
	if err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if result["id"] != "rollback-456" {
		t.Errorf("expected id rollback-456, got %v", result["id"])
	}
}

func TestLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/login" {
			t.Errorf("expected /api/v1/auth/login, got %s", r.URL.Path)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["email"] != "admin@test.com" {
			t.Errorf("expected email admin@test.com, got %s", body["email"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": "jwt-token-from-server",
		})
	}))
	defer server.Close()

	c := NewAPIClient(server.URL, "")
	err := c.Login("admin@test.com", "password123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if c.token != "jwt-token-from-server" {
		t.Errorf("expected token jwt-token-from-server, got %s", c.token)
	}
}
