package rest

import (
	"encoding/json"
	"net/http"

	"github.com/fleetml/fleetml/server/internal/integrations"
	mw "github.com/fleetml/fleetml/server/internal/middleware"
	"go.uber.org/zap"
)

// IntegrationHandler handles external model registry import endpoints.
type IntegrationHandler struct {
	service *integrations.Service
	logger  *zap.SugaredLogger
}

// NewIntegrationHandler creates a new integration handler.
func NewIntegrationHandler(service *integrations.Service, logger *zap.SugaredLogger) *IntegrationHandler {
	return &IntegrationHandler{service: service, logger: logger}
}

// ImportMLflow handles POST /api/v1/integrations/mlflow/import.
func (h *IntegrationHandler) ImportMLflow(w http.ResponseWriter, r *http.Request) {
	var req integrations.MLflowImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.WriteBadRequest(w, "invalid request body")
		return
	}

	if req.ModelName == "" {
		mw.WriteBadRequest(w, "model_name is required")
		return
	}

	result, err := h.service.ImportFromMLflow(r.Context(), req)
	if err != nil {
		h.logger.Errorw("failed to import from MLflow", "model", req.ModelName, "error", err)
		mw.WriteError(w, http.StatusBadRequest, err.Error(), "IMPORT_FAILED")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// ImportHuggingFace handles POST /api/v1/integrations/huggingface/import.
func (h *IntegrationHandler) ImportHuggingFace(w http.ResponseWriter, r *http.Request) {
	var req integrations.HFImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.WriteBadRequest(w, "invalid request body")
		return
	}

	if req.RepoID == "" {
		mw.WriteBadRequest(w, "repo_id is required")
		return
	}

	result, err := h.service.ImportFromHuggingFace(r.Context(), req)
	if err != nil {
		h.logger.Errorw("failed to import from HuggingFace", "repo_id", req.RepoID, "error", err)
		mw.WriteError(w, http.StatusBadRequest, err.Error(), "IMPORT_FAILED")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}
