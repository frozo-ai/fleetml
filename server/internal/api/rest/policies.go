package rest

import (
	"encoding/json"
	"net/http"

	mw "github.com/fleetml/fleetml/server/internal/middleware"
	"github.com/fleetml/fleetml/server/internal/policy"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// PolicyHandler handles policy endpoints.
type PolicyHandler struct {
	engine *policy.Engine
	logger *zap.SugaredLogger
}

// NewPolicyHandler creates a new policy handler.
func NewPolicyHandler(engine *policy.Engine, logger *zap.SugaredLogger) *PolicyHandler {
	return &PolicyHandler{engine: engine, logger: logger}
}

// Create handles POST /api/v1/policies.
func (h *PolicyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req policy.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.WriteBadRequest(w, "invalid request body")
		return
	}

	p, err := h.engine.Create(r.Context(), req)
	if err != nil {
		h.logger.Errorw("failed to create policy", "error", err)
		mw.WriteError(w, http.StatusBadRequest, err.Error(), "INVALID_REQUEST")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

// List handles GET /api/v1/policies.
func (h *PolicyHandler) List(w http.ResponseWriter, r *http.Request) {
	policyType := r.URL.Query().Get("type")

	policies, total, err := h.engine.List(r.Context(), policyType)
	if err != nil {
		h.logger.Errorw("failed to list policies", "error", err)
		mw.WriteInternalError(w, "failed to list policies")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"policies": policies,
		"total":    total,
	})
}

// Get handles GET /api/v1/policies/{id}.
func (h *PolicyHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !mw.IsValidUUID(id) {
		mw.WriteBadRequest(w, "invalid policy ID format")
		return
	}

	p, err := h.engine.Get(r.Context(), id)
	if err != nil {
		mw.WriteNotFound(w, "policy not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// Update handles PATCH /api/v1/policies/{id}.
func (h *PolicyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !mw.IsValidUUID(id) {
		mw.WriteBadRequest(w, "invalid policy ID format")
		return
	}

	var req policy.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.WriteBadRequest(w, "invalid request body")
		return
	}

	p, err := h.engine.Update(r.Context(), id, req)
	if err != nil {
		h.logger.Errorw("failed to update policy", "id", id, "error", err)
		mw.WriteError(w, http.StatusBadRequest, err.Error(), "INVALID_REQUEST")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// Delete handles DELETE /api/v1/policies/{id}.
func (h *PolicyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !mw.IsValidUUID(id) {
		mw.WriteBadRequest(w, "invalid policy ID format")
		return
	}

	if err := h.engine.Delete(r.Context(), id); err != nil {
		h.logger.Errorw("failed to delete policy", "id", id, "error", err)
		mw.WriteNotFound(w, "policy not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
