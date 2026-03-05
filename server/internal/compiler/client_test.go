package compiler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestCompile_ConnectionRefused(t *testing.T) {
	client := NewClient("http://127.0.0.1:1")
	_, err := client.Compile(context.Background(), CompileRequest{
		ModelURL: "s3://bucket/key", ModelID: "m1", TargetRuntime: "mock",
	})
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestCompile_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := NewClient(server.URL)
	_, err := client.Compile(ctx, CompileRequest{ModelURL: "s3://b/k", ModelID: "m1", TargetRuntime: "mock"})
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestCompile_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{not-valid`))
	}))
	defer server.Close()
	client := NewClient(server.URL)
	_, err := client.Compile(context.Background(), CompileRequest{ModelURL: "s3://b/k", ModelID: "m1", TargetRuntime: "mock"})
	if err == nil {
		t.Fatal("expected JSON decode error")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Errorf("expected 'decode response' in error, got: %v", err)
	}
}

func TestCompile_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	client := NewClient(server.URL)
	_, err := client.Compile(context.Background(), CompileRequest{ModelURL: "s3://b/k", ModelID: "m1", TargetRuntime: "mock"})
	if err == nil {
		t.Fatal("expected error for empty response body")
	}
}

func TestCompile_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
	}))
	defer server.Close()
	client := NewClient(server.URL)
	client.httpClient.Timeout = 50 * time.Millisecond
	_, err := client.Compile(context.Background(), CompileRequest{ModelURL: "s3://b/k", ModelID: "m1", TargetRuntime: "mock"})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestCompile_NonJSONContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`plain text`))
	}))
	defer server.Close()
	client := NewClient(server.URL)
	_, err := client.Compile(context.Background(), CompileRequest{ModelURL: "s3://b/k", ModelID: "m1", TargetRuntime: "mock"})
	if err == nil {
		t.Fatal("expected error parsing non-JSON body")
	}
}

func TestCompile_500InternalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal server error`))
	}))
	defer server.Close()
	client := NewClient(server.URL)
	_, err := client.Compile(context.Background(), CompileRequest{ModelURL: "s3://b/k", ModelID: "m1", TargetRuntime: "mock"})
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected 500 in error, got: %v", err)
	}
}

func TestCompile_404NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`not found`))
	}))
	defer server.Close()
	client := NewClient(server.URL)
	_, err := client.Compile(context.Background(), CompileRequest{ModelURL: "s3://b/k", ModelID: "m1", TargetRuntime: "mock"})
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestCompile_LargePayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CompileRequest
		json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CompileResponse{Runtime: "mock"})
	}))
	defer server.Close()
	opts := make(map[string]any, 500)
	for i := 0; i < 500; i++ {
		opts[fmt.Sprintf("key-%04d", i)] = strings.Repeat("v", 100)
	}
	client := NewClient(server.URL)
	resp, err := client.Compile(context.Background(), CompileRequest{
		ModelURL: "s3://b/k", ModelID: "m1", TargetRuntime: "mock", Options: opts,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Runtime != "mock" {
		t.Errorf("expected runtime mock, got %s", resp.Runtime)
	}
}

func TestNewClient_BaseURL(t *testing.T) {
	tests := []struct{ name, url string }{
		{"standard", "http://localhost:8000"},
		{"trailing slash", "http://localhost:8000/"},
		{"empty", ""},
		{"https", "https://compiler.example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(tt.url)
			if c == nil {
				t.Fatal("expected non-nil client")
			}
			if c.baseURL != tt.url {
				t.Errorf("expected baseURL %q, got %q", tt.url, c.baseURL)
			}
			if c.httpClient == nil {
				t.Fatal("expected non-nil httpClient")
			}
			if c.httpClient.Timeout != 10*time.Minute {
				t.Errorf("expected 10m timeout, got %v", c.httpClient.Timeout)
			}
		})
	}
}

func TestHealth_ConnectionRefused(t *testing.T) {
	client := NewClient("http://127.0.0.1:1")
	err := client.Health(context.Background())
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestHealth_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := NewClient("http://localhost:8000")
	err := client.Health(ctx)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestCompileRequest_JSONSerialization(t *testing.T) {
	req := CompileRequest{
		ModelURL: "s3://bucket/model.onnx", ModelID: "abc-123",
		TargetRuntime: "tensorrt", Options: map[string]any{"precision": "fp16"},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded CompileRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.ModelURL != req.ModelURL || decoded.ModelID != req.ModelID || decoded.TargetRuntime != req.TargetRuntime {
		t.Error("field mismatch after round-trip")
	}
}

func TestCompileResponse_JSONSerialization(t *testing.T) {
	resp := CompileResponse{
		Runtime: "openvino", ArtifactURL: "s3://bucket/compiled/model.xml",
		Checksum: "sha256:deadbeef", FileSize: 999999, CompileTimeSeconds: 42.5,
		Metadata: map[string]any{"device": "CPU"},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded CompileResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Runtime != resp.Runtime || decoded.FileSize != resp.FileSize || decoded.Checksum != resp.Checksum {
		t.Error("field mismatch after round-trip")
	}
}

func TestCompile_EmptyModelID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CompileResponse{Runtime: "mock"})
	}))
	defer server.Close()
	client := NewClient(server.URL)
	resp, err := client.Compile(context.Background(), CompileRequest{ModelURL: "s3://b/k", ModelID: "", TargetRuntime: "mock"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestCompile_EmptyRuntime(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CompileResponse{Runtime: ""})
	}))
	defer server.Close()
	client := NewClient(server.URL)
	resp, err := client.Compile(context.Background(), CompileRequest{ModelURL: "s3://b/k", ModelID: "m1", TargetRuntime: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Runtime != "" {
		t.Errorf("expected empty runtime, got %q", resp.Runtime)
	}
}
