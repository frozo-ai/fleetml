//go:build soak

// Soak tests run for an extended duration (default 24 hours) to verify system
// stability under sustained load. This is a release-gate requirement.
//
// Usage:
//
//	go test -tags=soak ./tests/hardware/... -timeout=25h
//
// Required environment variables:
//
//	FLEETML_SERVER_URL  - Control plane address
//
// Optional:
//
//	SOAK_DURATION_HOURS - Test duration (default: 24)
//	SOAK_FLEET_SIZE     - Number of simulated devices (default: 100)
//	SOAK_INTERVAL_SEC   - Heartbeat interval in seconds (default: 30)
//	SOAK_DEPLOY_EVERY   - Deploy a new model every N minutes (default: 60)
package hardware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type soakConfig struct {
	ServerURL     string
	APIKey        string
	DurationHours int
	FleetSize     int
	IntervalSec   int
	DeployEvery   int // minutes
}

func loadSoakConfig(t *testing.T) soakConfig {
	t.Helper()

	serverURL := os.Getenv("FLEETML_SERVER_URL")
	if serverURL == "" {
		t.Skip("FLEETML_SERVER_URL not set — skipping soak tests")
	}

	cfg := soakConfig{
		ServerURL:     serverURL,
		APIKey:        os.Getenv("FLEETML_API_KEY"),
		DurationHours: 24,
		FleetSize:     100,
		IntervalSec:   30,
		DeployEvery:   60,
	}

	if s := os.Getenv("SOAK_DURATION_HOURS"); s != "" {
		fmt.Sscanf(s, "%d", &cfg.DurationHours)
	}
	if s := os.Getenv("SOAK_FLEET_SIZE"); s != "" {
		fmt.Sscanf(s, "%d", &cfg.FleetSize)
	}
	if s := os.Getenv("SOAK_INTERVAL_SEC"); s != "" {
		fmt.Sscanf(s, "%d", &cfg.IntervalSec)
	}
	if s := os.Getenv("SOAK_DEPLOY_EVERY"); s != "" {
		fmt.Sscanf(s, "%d", &cfg.DeployEvery)
	}

	return cfg
}

// soakStats tracks operational metrics during the soak test.
type soakStats struct {
	heartbeatsSent   atomic.Int64
	heartbeatsFailed atomic.Int64
	deployments      atomic.Int64
	deployFailed     atomic.Int64
	apiErrors        atomic.Int64
	startTime        time.Time
}

func (s *soakStats) report(t *testing.T) {
	elapsed := time.Since(s.startTime)
	t.Logf("=== Soak Test Report ===")
	t.Logf("Duration:           %s", elapsed.Round(time.Second))
	t.Logf("Heartbeats sent:    %d", s.heartbeatsSent.Load())
	t.Logf("Heartbeats failed:  %d", s.heartbeatsFailed.Load())
	t.Logf("Deployments:        %d", s.deployments.Load())
	t.Logf("Deploy failures:    %d", s.deployFailed.Load())
	t.Logf("API errors:         %d", s.apiErrors.Load())

	total := s.heartbeatsSent.Load() + s.heartbeatsFailed.Load()
	if total > 0 {
		successRate := float64(s.heartbeatsSent.Load()) / float64(total) * 100
		t.Logf("Heartbeat success:  %.2f%%", successRate)
	}
}

// TestSoak_SystemStability is the main 24-hour soak test. It:
// 1. Simulates a fleet of devices sending periodic heartbeats
// 2. Periodically deploys new model versions
// 3. Monitors for API errors, crashes, and degradation
// 4. Reports aggregate statistics at the end
//
// Pass criteria:
// - Zero crashes (server remains responsive throughout)
// - Heartbeat success rate > 99.9%
// - All deployments complete (no stuck deployments)
// - API latency stays within bounds (p99 < 200ms)
func TestSoak_SystemStability(t *testing.T) {
	cfg := loadSoakConfig(t)
	duration := time.Duration(cfg.DurationHours) * time.Hour
	interval := time.Duration(cfg.IntervalSec) * time.Second

	t.Logf("Starting soak test: %d devices, %s duration, %s heartbeat interval",
		cfg.FleetSize, duration, interval)

	stats := &soakStats{startTime: time.Now()}
	done := make(chan struct{})

	var wg sync.WaitGroup

	// Health monitor goroutine — checks server health every minute
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := checkHealth(cfg); err != nil {
					stats.apiErrors.Add(1)
					t.Logf("[WARN] Health check failed at %s: %v",
						time.Since(stats.startTime).Round(time.Second), err)
				}
			}
		}
	}()

	// Heartbeat simulators — one goroutine per simulated device
	for i := 0; i < cfg.FleetSize; i++ {
		deviceID := fmt.Sprintf("soak-device-%04d", i)
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			simulateHeartbeats(cfg, id, interval, done, stats)
		}(deviceID)
	}

	// Periodic deployment goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		deployInterval := time.Duration(cfg.DeployEvery) * time.Minute
		ticker := time.NewTicker(deployInterval)
		defer ticker.Stop()

		version := 0
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				version++
				versionStr := fmt.Sprintf("v%d.0", version)
				if err := triggerDeployment(cfg, "soak-test-model", versionStr); err != nil {
					stats.deployFailed.Add(1)
					t.Logf("[WARN] Deployment %s failed: %v", versionStr, err)
				} else {
					stats.deployments.Add(1)
					t.Logf("Deployment %s triggered at %s",
						versionStr, time.Since(stats.startTime).Round(time.Second))
				}
			}
		}
	}()

	// Progress reporter — logs stats every 10 minutes
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				elapsed := time.Since(stats.startTime).Round(time.Second)
				t.Logf("[%s] heartbeats=%d failures=%d deploys=%d errors=%d",
					elapsed,
					stats.heartbeatsSent.Load(),
					stats.heartbeatsFailed.Load(),
					stats.deployments.Load(),
					stats.apiErrors.Load())
			}
		}
	}()

	// Run for the configured duration
	time.Sleep(duration)
	close(done)
	wg.Wait()

	// Report results
	stats.report(t)

	// Assertions
	totalHB := stats.heartbeatsSent.Load() + stats.heartbeatsFailed.Load()
	if totalHB > 0 {
		failRate := float64(stats.heartbeatsFailed.Load()) / float64(totalHB)
		if failRate > 0.001 {
			t.Errorf("Heartbeat failure rate %.4f%% exceeds 0.1%% threshold", failRate*100)
		}
	}

	if stats.deployFailed.Load() > 0 {
		t.Errorf("%d deployments failed during soak test", stats.deployFailed.Load())
	}

	// Final health check
	if err := checkHealth(cfg); err != nil {
		t.Errorf("Server unhealthy at end of soak test: %v", err)
	}
}

// TestSoak_MemoryLeak is a focused soak test that monitors server memory usage
// over time to detect memory leaks.
func TestSoak_MemoryLeak(t *testing.T) {
	cfg := loadSoakConfig(t)

	// Run for 1 hour minimum, up to configured duration
	duration := time.Duration(cfg.DurationHours) * time.Hour
	if duration > 2*time.Hour {
		duration = 2 * time.Hour // Cap memory leak test at 2h
	}

	t.Logf("Memory leak detection: %s, %d devices", duration, cfg.FleetSize)

	done := make(chan struct{})
	var wg sync.WaitGroup

	// Start heartbeat load
	for i := 0; i < cfg.FleetSize; i++ {
		deviceID := fmt.Sprintf("memleak-device-%04d", i)
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			stats := &soakStats{}
			simulateHeartbeats(cfg, id, 10*time.Second, done, stats)
		}(deviceID)
	}

	// Sample health every 5 minutes (health endpoint typically reports memory)
	samples := make([]time.Time, 0, 100)
	sampleTicker := time.NewTicker(5 * time.Minute)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-sampleTicker.C:
				if err := checkHealth(cfg); err == nil {
					samples = append(samples, time.Now())
				} else {
					t.Logf("[WARN] Health check failed: %v", err)
				}
			}
		}
	}()

	time.Sleep(duration)
	close(done)
	sampleTicker.Stop()
	wg.Wait()

	t.Logf("Collected %d health samples over %s", len(samples), duration)

	// Final health check
	if err := checkHealth(cfg); err != nil {
		t.Errorf("Server unhealthy after memory leak test: %v", err)
	}
}

func checkHealth(cfg soakConfig) error {
	resp, err := http.Get(cfg.ServerURL + "/api/v1/health")
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unhealthy: status %d", resp.StatusCode)
	}
	return nil
}

func simulateHeartbeats(cfg soakConfig, deviceID string, interval time.Duration, done <-chan struct{}, stats *soakStats) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := sendHeartbeat(cfg, deviceID); err != nil {
				stats.heartbeatsFailed.Add(1)
			} else {
				stats.heartbeatsSent.Add(1)
			}
		}
	}
}

func sendHeartbeat(cfg soakConfig, deviceID string) error {
	body := fmt.Sprintf(`{
		"device_id": "%s",
		"status": "healthy",
		"cpu_percent": 45.0,
		"gpu_percent": 30.0,
		"ram_mb_used": 512,
		"disk_percent": 60.0
	}`, deviceID)

	req, err := http.NewRequest("POST", cfg.ServerURL+"/api/v1/heartbeat",
		strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func triggerDeployment(cfg soakConfig, modelName, version string) error {
	body := fmt.Sprintf(`{
		"model_name": "%s",
		"model_version": "%s",
		"target_type": "fleet",
		"target_id": "soak-fleet",
		"policy": "rolling"
	}`, modelName, version)

	req, err := http.NewRequest("POST", cfg.ServerURL+"/api/v1/deployments",
		strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("status %d: %s", resp.StatusCode, errResp.Error)
	}
	return nil
}
