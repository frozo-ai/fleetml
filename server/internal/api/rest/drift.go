package rest

import (
	"encoding/json"
	"net/http"

	"github.com/fleetml/fleetml/server/internal/drift"
	mw "github.com/fleetml/fleetml/server/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// DriftHandler handles drift detection endpoints.
type DriftHandler struct {
	detector *drift.Detector
	logger   *zap.SugaredLogger
}

// NewDriftHandler creates a new drift handler.
func NewDriftHandler(detector *drift.Detector, logger *zap.SugaredLogger) *DriftHandler {
	return &DriftHandler{detector: detector, logger: logger}
}

// Analyze handles POST /api/v1/drift/analyze — submit samples for drift analysis.
func (h *DriftHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeviceID    string    `json:"device_id"`
		ModelID     string    `json:"model_id"`
		FeatureName string    `json:"feature_name"`
		Samples     []float64 `json:"samples"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.WriteBadRequest(w, "invalid request body")
		return
	}

	if req.DeviceID == "" || req.ModelID == "" || req.FeatureName == "" {
		mw.WriteBadRequest(w, "device_id, model_id, and feature_name are required")
		return
	}
	if len(req.Samples) == 0 {
		mw.WriteBadRequest(w, "samples array is required and must not be empty")
		return
	}

	report, err := h.detector.Analyze(r.Context(), req.DeviceID, req.ModelID, req.FeatureName, req.Samples)
	if err != nil {
		h.logger.Errorw("drift analysis failed", "error", err)
		mw.WriteError(w, http.StatusBadRequest, err.Error(), "ANALYSIS_FAILED")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// SetBaseline handles POST /api/v1/drift/baselines — set baseline distribution.
func (h *DriftHandler) SetBaseline(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ModelID     string    `json:"model_id"`
		FeatureName string    `json:"feature_name"`
		Samples     []float64 `json:"samples"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.WriteBadRequest(w, "invalid request body")
		return
	}

	if req.ModelID == "" || req.FeatureName == "" || len(req.Samples) == 0 {
		mw.WriteBadRequest(w, "model_id, feature_name, and samples are required")
		return
	}

	if err := h.detector.SetBaseline(r.Context(), req.ModelID, req.FeatureName, req.Samples); err != nil {
		h.logger.Errorw("failed to set baseline", "error", err)
		mw.WriteInternalError(w, "failed to set baseline")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "baseline set"})
}

// ListReports handles GET /api/v1/drift/reports.
func (h *DriftHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	modelID := r.URL.Query().Get("model_id")
	deviceID := r.URL.Query().Get("device_id")
	driftOnly := r.URL.Query().Get("drift_only") == "true"

	reports, err := h.detector.ListReports(r.Context(), modelID, deviceID, driftOnly)
	if err != nil {
		h.logger.Errorw("failed to list drift reports", "error", err)
		mw.WriteInternalError(w, "failed to list drift reports")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reports": reports,
		"total":   len(reports),
	})
}

// GetReport handles GET /api/v1/drift/reports/{id}.
func (h *DriftHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "id")
	// Would query by ID — simplified for now
	mw.WriteNotFound(w, "drift report not found")
}
