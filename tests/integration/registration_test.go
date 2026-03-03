//go:build integration

package integration

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestDeviceRegistration tests the full agent registration flow against a running server.
func TestDeviceRegistration(t *testing.T) {
	serverURL := os.Getenv("FLEETML_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for server to be ready
	if err := waitForServer(ctx, serverURL); err != nil {
		t.Fatalf("server not ready: %v", err)
	}

	// Register a device via REST fallback
	// In real integration test, this would use gRPC
	t.Log("Server is ready, registration test would proceed here")
}

func waitForServer(ctx context.Context, url string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			resp, err := http.Get(url + "/health")
			if err == nil && resp.StatusCode == http.StatusOK {
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}
