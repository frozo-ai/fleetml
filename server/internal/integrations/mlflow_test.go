package integrations

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMLflowClient_GetModel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/2.0/mlflow/registered-models/get" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("name") != "test-model" {
			t.Errorf("unexpected name param: %s", r.URL.Query().Get("name"))
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"registered_model": map[string]interface{}{
				"name": "test-model",
				"latest_versions": []map[string]interface{}{
					{"name": "test-model", "version": "3", "run_id": "abc123", "status": "READY"},
				},
			},
		})
	}))
	defer srv.Close()

	client := NewMLflowClient(srv.URL)
	model, err := client.GetModel(context.Background(), "test-model")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.Name != "test-model" {
		t.Errorf("expected name test-model, got %s", model.Name)
	}
	if len(model.LatestVersions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(model.LatestVersions))
	}
	if model.LatestVersions[0].Version != "3" {
		t.Errorf("expected version 3, got %s", model.LatestVersions[0].Version)
	}
}

func TestMLflowClient_GetModelVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"model_version": map[string]interface{}{
				"name":    "mymodel",
				"version": "2",
				"run_id":  "run-xyz",
				"source":  "s3://mlflow/models/mymodel/2",
			},
		})
	}))
	defer srv.Close()

	client := NewMLflowClient(srv.URL)
	mv, err := client.GetModelVersion(context.Background(), "mymodel", "2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mv.RunID != "run-xyz" {
		t.Errorf("expected run_id run-xyz, got %s", mv.RunID)
	}
}

func TestMLflowClient_ListArtifacts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"files": []map[string]interface{}{
				{"path": "model.onnx", "is_dir": false, "file_size": 1024},
				{"path": "config.json", "is_dir": false, "file_size": 256},
			},
		})
	}))
	defer srv.Close()

	client := NewMLflowClient(srv.URL)
	artifacts, err := client.ListArtifacts(context.Background(), "run123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(artifacts) != 2 {
		t.Errorf("expected 2 artifacts, got %d", len(artifacts))
	}
}

func TestMLflowClient_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer srv.Close()

	client := NewMLflowClient(srv.URL)
	_, err := client.GetModel(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name       string
		artifacts  []MLflowArtifact
		wantFormat string
		wantPath   string
	}{
		{
			name: "onnx model",
			artifacts: []MLflowArtifact{
				{Path: "model.onnx", IsDir: false, FileSize: 1024},
				{Path: "config.json", IsDir: false, FileSize: 100},
			},
			wantFormat: "onnx",
			wantPath:   "model.onnx",
		},
		{
			name: "pytorch model",
			artifacts: []MLflowArtifact{
				{Path: "model.pt", IsDir: false, FileSize: 2048},
			},
			wantFormat: "pytorch",
			wantPath:   "model.pt",
		},
		{
			name: "tflite model",
			artifacts: []MLflowArtifact{
				{Path: "model.tflite", IsDir: false, FileSize: 512},
			},
			wantFormat: "tflite",
			wantPath:   "model.tflite",
		},
		{
			name:       "no model files",
			artifacts:  []MLflowArtifact{{Path: "readme.md", IsDir: false}},
			wantFormat: "onnx",
			wantPath:   "model.onnx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, path := DetectFormat(tt.artifacts)
			if format != tt.wantFormat {
				t.Errorf("format = %q, want %q", format, tt.wantFormat)
			}
			if path != tt.wantPath {
				t.Errorf("path = %q, want %q", path, tt.wantPath)
			}
		})
	}
}

func TestHasExtension(t *testing.T) {
	tests := []struct {
		path string
		ext  string
		want bool
	}{
		{"model.onnx", ".onnx", true},
		{"model.pt", ".pt", true},
		{"model.onnx", ".pt", false},
		{"x", ".onnx", false},
		{"", ".onnx", false},
	}

	for _, tt := range tests {
		got := hasExtension(tt.path, tt.ext)
		if got != tt.want {
			t.Errorf("hasExtension(%q, %q) = %v, want %v", tt.path, tt.ext, got, tt.want)
		}
	}
}
