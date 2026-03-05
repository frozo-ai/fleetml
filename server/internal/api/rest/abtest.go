package rest

import (
	"encoding/json"
	"net/http"

	"github.com/fleetml/fleetml/server/internal/abtest"
	mw "github.com/fleetml/fleetml/server/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// ABTestHandler handles A/B test endpoints.
type ABTestHandler struct {
	manager *abtest.Manager
	logger  *zap.SugaredLogger
}

// NewABTestHandler creates a new A/B test handler.
func NewABTestHandler(manager *abtest.Manager, logger *zap.SugaredLogger) *ABTestHandler {
	return &ABTestHandler{manager: manager, logger: logger}
}

// Create handles POST /api/v1/ab-tests.
func (h *ABTestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req abtest.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.WriteBadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		mw.WriteBadRequest(w, "name is required")
		return
	}
	if req.ModelAID == "" || req.ModelBID == "" {
		mw.WriteBadRequest(w, "model_a_id and model_b_id are required")
		return
	}

	test, err := h.manager.Create(r.Context(), req)
	if err != nil {
		h.logger.Errorw("failed to create A/B test", "error", err)
		mw.WriteError(w, http.StatusBadRequest, err.Error(), "INVALID_REQUEST")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(test)
}

// List handles GET /api/v1/ab-tests.
func (h *ABTestHandler) List(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")

	tests, total, err := h.manager.List(r.Context(), state)
	if err != nil {
		h.logger.Errorw("failed to list A/B tests", "error", err)
		mw.WriteInternalError(w, "failed to list A/B tests")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ab_tests": tests,
		"total":    total,
	})
}

// Get handles GET /api/v1/ab-tests/{id}.
func (h *ABTestHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !mw.IsValidUUID(id) {
		mw.WriteBadRequest(w, "invalid test ID format")
		return
	}

	test, err := h.manager.Get(r.Context(), id)
	if err != nil {
		mw.WriteNotFound(w, "A/B test not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(test)
}

// Stop handles POST /api/v1/ab-tests/{id}/stop.
func (h *ABTestHandler) Stop(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !mw.IsValidUUID(id) {
		mw.WriteBadRequest(w, "invalid test ID format")
		return
	}

	var req struct {
		Winner string `json:"winner"`
	}
	json.NewDecoder(r.Body).Decode(&req) // Body is optional

	test, err := h.manager.Stop(r.Context(), id, req.Winner)
	if err != nil {
		h.logger.Errorw("failed to stop A/B test", "id", id, "error", err)
		mw.WriteError(w, http.StatusBadRequest, err.Error(), "INVALID_REQUEST")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(test)
}
