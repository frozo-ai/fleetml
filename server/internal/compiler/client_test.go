package compiler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompile_Success(t *testing.T) {
	expected := CompileResponse{
		Runtime:            "mock",
		ArtifactURL:        "s3://fleetml-models/test-id/compiled/mock/model.onnx",
		Checksum:           "sha256:abc123",
		FileSize:           1024,
		CompileTimeSeconds: 0.5,
		Metadata:           map[string]any{"compiler": "mock"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/compile" {
			t.Errorf("expected /compile, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type")
		}

		var req CompileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.ModelID != "test-model-id" {
			t.Errorf("expected model_id test-model-id, got %s", req.ModelID)
		}
		if req.TargetRuntime != "mock" {
			t.Errorf("expected target_runtime mock, got %s", req.TargetRuntime)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resp, err := client.Compile(context.Background(), CompileRequest{
		ModelURL:      "s3://fleetml-models/test/v1/model.onnx",
		ModelID:       "test-model-id",
		TargetRuntime: "mock",
	})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	if resp.Runtime != "mock" {
		t.Errorf("expected runtime mock, got %s", resp.Runtime)
	}
	if resp.ArtifactURL != expected.ArtifactURL {
		t.Errorf("expected artifact_url %s, got %s", expected.ArtifactURL, resp.ArtifactURL)
	}
	if resp.Checksum != expected.Checksum {
		t.Errorf("expected checksum %s, got %s", expected.Checksum, resp.Checksum)
	}
	if resp.FileSize != expected.FileSize {
		t.Errorf("expected file_size %d, got %d", expected.FileSize, resp.FileSize)
	}
}

func TestCompile_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"detail":"Unsupported runtime"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.Compile(context.Background(), CompileRequest{
		ModelURL:      "s3://bucket/key",
		ModelID:       "id",
		TargetRuntime: "unknown",
	})
	if err == nil {
		t.Fatal("expected error for bad request")
	}
}

func TestHealth_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("expected /health, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.Health(context.Background()); err != nil {
		t.Fatalf("health check: %v", err)
	}
}

func TestHealth_Unhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.Health(context.Background()); err == nil {
		t.Fatal("expected error for unhealthy service")
	}
}
