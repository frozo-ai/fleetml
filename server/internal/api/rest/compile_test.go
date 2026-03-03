package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fleetml/fleetml/server/internal/compiler"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func TestCompileHandler_MissingRuntime(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewCompileHandler(nil, nil, logger)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/models/test-id/compile", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-id")
	req = req.WithContext(chi.RouteContext(req.Context(), rctx))

	w := httptest.NewRecorder()
	handler.Compile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCompileHandler_InvalidBody(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := NewCompileHandler(nil, nil, logger)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/models/test-id/compile", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-id")
	req = req.WithContext(chi.RouteContext(req.Context(), rctx))

	w := httptest.NewRecorder()
	handler.Compile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCompilerClient_MockService(t *testing.T) {
	expected := compiler.CompileResponse{
		Runtime:            "mock",
		ArtifactURL:        "s3://fleetml-models/test-id/compiled/mock/model.onnx",
		Checksum:           "sha256:abc123",
		FileSize:           1024,
		CompileTimeSeconds: 0.5,
		Metadata:           map[string]any{"compiler": "mock"},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer mockServer.Close()

	client := compiler.NewClient(mockServer.URL)
	resp, err := client.Compile(context.Background(), compiler.CompileRequest{
		ModelURL:      "s3://fleetml-models/test/v1/model.onnx",
		ModelID:       "test-id",
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
}
