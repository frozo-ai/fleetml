package rest

import (
	"encoding/json"
	"net/http"

	"github.com/fleetml/fleetml/server/internal/auth"
	"github.com/fleetml/fleetml/server/internal/deploy"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// DeployHandler handles deployment-related endpoints.
type DeployHandler struct {
	orchestrator *deploy.Orchestrator
	logger       *zap.SugaredLogger
}

func NewDeployHandler(orchestrator *deploy.Orchestrator, logger *zap.SugaredLogger) *DeployHandler {
	return &DeployHandler{orchestrator: orchestrator, logger: logger}
}

// Create creates a new deployment.
func (h *DeployHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req deploy.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.ModelName == "" || req.ModelVersion == "" {
		http.Error(w, `{"error":"model_name and model_version are required"}`, http.StatusBadRequest)
		return
	}

	req.OrgID = claims.OrgID
	d, err := h.orchestrator.CreateDeployment(r.Context(), req)
	if err != nil {
		h.logger.Errorw("failed to create deployment", "error", err)
		http.Error(w, `{"error":"failed to create deployment"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}

// List lists deployments.
func (h *DeployHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	deployments, total, err := h.orchestrator.ListDeployments(r.Context(),
		claims.OrgID,
		r.URL.Query().Get("state"),
		r.URL.Query().Get("model_name"),
	)
	if err != nil {
		http.Error(w, `{"error":"failed to list deployments"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deployments": deployments,
		"total":       total,
	})
}

// Get returns a single deployment with per-device status.
func (h *DeployHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")

	d, err := h.orchestrator.GetDeployment(r.Context(), id, claims.OrgID)
	if err != nil {
		http.Error(w, `{"error":"deployment not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}

// Cancel cancels a running deployment.
func (h *DeployHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")

	if err := h.orchestrator.CancelDeployment(r.Context(), claims.OrgID, id); err != nil {
		http.Error(w, `{"error":"failed to cancel deployment"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "deployment cancelled"})
}

// Rollback rolls back a deployment by creating a reverse deployment.
func (h *DeployHandler) Rollback(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")

	d, err := h.orchestrator.RollbackDeployment(r.Context(), claims.OrgID, id)
	if err != nil {
		h.logger.Errorw("rollback failed", "deployment_id", id, "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}
