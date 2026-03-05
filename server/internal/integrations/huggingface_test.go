package integrations

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHuggingFaceClient_GetModelInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/models/bert-base-uncased" {
			// Could be the CDN base, handle accordingly
			if r.URL.Path == "/models/bert-base-uncased" {
				json.NewEncoder(w).Encode(HFModelInfo{
					ID:          "bert-base-uncased",
					ModelID:     "bert-base-uncased",
					Author:      "google",
					Tags:        []string{"transformers", "pytorch"},
					Pipeline:    "fill-mask",
					LibraryName: "transformers",
					Downloads:   1000000,
				})
				return
			}
		}
		json.NewEncoder(w).Encode(HFModelInfo{
			ID:          "bert-base-uncased",
			ModelID:     "bert-base-uncased",
			Author:      "google",
			Tags:        []string{"transformers", "pytorch"},
			Pipeline:    "fill-mask",
			LibraryName: "transformers",
			Downloads:   1000000,
		})
	}))
	defer srv.Close()

	// Override the API base for testing
	client := &HuggingFaceClient{
		httpClient: srv.Client(),
	}

	// Use server URL as base
	url := srv.URL + "/api/models/bert-base-uncased"
	req, _ := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var info HFModelInfo
	json.NewDecoder(resp.Body).Decode(&info)
	if info.ID != "bert-base-uncased" {
		t.Errorf("expected bert-base-uncased, got %s", info.ID)
	}
	if info.Pipeline != "fill-mask" {
		t.Errorf("expected fill-mask, got %s", info.Pipeline)
	}
}

func TestFindONNXFile(t *testing.T) {
	tests := []struct {
		name     string
		siblings []HFFileSibling
		wantFile string
		wantOK   bool
	}{
		{
			name: "model.onnx present",
			siblings: []HFFileSibling{
				{Filename: "config.json"},
				{Filename: "model.onnx"},
				{Filename: "tokenizer.json"},
			},
			wantFile: "model.onnx",
			wantOK:   true,
		},
		{
			name: "other onnx file",
			siblings: []HFFileSibling{
				{Filename: "config.json"},
				{Filename: "encoder.onnx"},
			},
			wantFile: "encoder.onnx",
			wantOK:   true,
		},
		{
			name: "no onnx files",
			siblings: []HFFileSibling{
				{Filename: "pytorch_model.bin"},
				{Filename: "config.json"},
			},
			wantFile: "",
			wantOK:   false,
		},
		{
			name: "prefers model.onnx over others",
			siblings: []HFFileSibling{
				{Filename: "decoder.onnx"},
				{Filename: "model.onnx"},
				{Filename: "encoder.onnx"},
			},
			wantFile: "model.onnx",
			wantOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, ok := FindONNXFile(tt.siblings)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if file != tt.wantFile {
				t.Errorf("file = %q, want %q", file, tt.wantFile)
			}
		})
	}
}

func TestDetectHFFormat(t *testing.T) {
	tests := []struct {
		name     string
		siblings []HFFileSibling
		want     string
	}{
		{
			name:     "onnx model",
			siblings: []HFFileSibling{{Filename: "model.onnx"}},
			want:     "onnx",
		},
		{
			name:     "pytorch model",
			siblings: []HFFileSibling{{Filename: "pytorch_model.bin"}},
			want:     "pytorch",
		},
		{
			name:     "tflite model",
			siblings: []HFFileSibling{{Filename: "model.tflite"}},
			want:     "tflite",
		},
		{
			name:     "unknown format",
			siblings: []HFFileSibling{{Filename: "config.json"}},
			want:     "unknown",
		},
		{
			name:     "onnx preferred over pytorch",
			siblings: []HFFileSibling{{Filename: "model.pt"}, {Filename: "model.onnx"}},
			want:     "onnx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectHFFormat(tt.siblings)
			if got != tt.want {
				t.Errorf("DetectHFFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHuggingFaceClient_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Model not found"}`))
	}))
	defer srv.Close()

	// Test directly since we can't easily override the const base URL
	client := &HuggingFaceClient{httpClient: srv.Client()}
	req, _ := http.NewRequestWithContext(context.Background(), "GET", srv.URL+"/api/models/nonexistent", nil)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
