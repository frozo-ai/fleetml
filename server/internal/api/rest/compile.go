package rest

import (
	"encoding/json"
	"net/http"

	"github.com/fleetml/fleetml/server/internal/compiler"
	"github.com/fleetml/fleetml/server/internal/domain"
	mw "github.com/fleetml/fleetml/server/internal/middleware"
	"github.com/fleetml/fleetml/server/internal/model"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// CompileHandler handles model compilation endpoints.
type CompileHandler struct {
	registry       *model.Registry
	compilerClient *compiler.Client
	logger         *zap.SugaredLogger
}

// NewCompileHandler creates a new compile handler.
func NewCompileHandler(registry *model.Registry, compilerClient *compiler.Client, logger *zap.SugaredLogger) *CompileHandler {
	return &CompileHandler{
		registry:       registry,
		compilerClient: compilerClient,
		logger:         logger,
	}
}

type compileRequest struct {
	TargetRuntime string         `json:"target_runtime"`
	Options       map[string]any `json:"options,omitempty"`
}

// Compile handles POST /api/v1/models/{id}/compile.
func (h *CompileHandler) Compile(w http.ResponseWriter, r *http.Request) {
	modelID := chi.URLParam(r, "id")

	// 1. Parse request body
	var req compileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.WriteBadRequest(w, "invalid request body")
		return
	}

	if req.TargetRuntime == "" {
		mw.WriteBadRequest(w, "target_runtime is required")
		return
	}

	// 2. Get model from registry (verify it exists and is ONNX)
	m, err := h.registry.GetModelByID(r.Context(), modelID)
	if err != nil {
		mw.WriteNotFound(w, "model not found")
		return
	}

	if m.Format != "onnx" {
		mw.WriteBadRequest(w, "only ONNX models can be compiled")
		return
	}

	// 3. Call compiler service
	compileResp, err := h.compilerClient.Compile(r.Context(), compiler.CompileRequest{
		ModelURL:      m.ArtifactURL,
		ModelID:       m.ID,
		TargetRuntime: req.TargetRuntime,
		Options:       req.Options,
	})
	if err != nil {
		h.logger.Errorw("compilation failed", "model_id", modelID, "runtime", req.TargetRuntime, "error", err)
		mw.WriteInternalError(w, "compilation failed: "+err.Error())
		return
	}

	// 4. Store compiled variant in the model record
	variant := domain.CompiledVariant{
		Runtime:     compileResp.Runtime,
		ArtifactURL: compileResp.ArtifactURL,
		Checksum:    compileResp.Checksum,
	}

	if err := h.registry.AddCompiledVariant(r.Context(), modelID, variant); err != nil {
		h.logger.Errorw("failed to store compiled variant", "model_id", modelID, "error", err)
		mw.WriteInternalError(w, "failed to store compiled variant")
		return
	}

	h.logger.Infow("model compiled successfully",
		"model_id", modelID,
		"runtime", compileResp.Runtime,
		"compile_time_s", compileResp.CompileTimeSeconds,
		"file_size", compileResp.FileSize,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(compileResp)
}
