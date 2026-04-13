package rest

import (
	"encoding/json"
	"net/http"

	"github.com/fleetml/fleetml/server/internal/auth"
	"github.com/fleetml/fleetml/server/internal/domain"
	"github.com/fleetml/fleetml/server/internal/fleet"
	mw "github.com/fleetml/fleetml/server/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// FleetHandler handles fleet-related endpoints.
type FleetHandler struct {
	fleet  *fleet.Manager
	logger *zap.SugaredLogger
}

func NewFleetHandler(fleetMgr *fleet.Manager, logger *zap.SugaredLogger) *FleetHandler {
	return &FleetHandler{fleet: fleetMgr, logger: logger}
}

// Create creates a new fleet.
func (h *FleetHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Labels      map[string]string `json:"labels"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	f, err := h.fleet.CreateFleet(r.Context(), claims.OrgID, req.Name, req.Description, req.Labels)
	if err != nil {
		http.Error(w, `{"error":"failed to create fleet"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(f)
}

// List lists all fleets.
func (h *FleetHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	fleets, err := h.fleet.ListFleets(r.Context(), claims.OrgID)
	if err != nil {
		http.Error(w, `{"error":"failed to list fleets"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"fleets": fleets,
	})
}

// Get returns a single fleet.
func (h *FleetHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")

	f, err := h.fleet.GetFleet(r.Context(), claims.OrgID, id)
	if err != nil {
		http.Error(w, `{"error":"fleet not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
}

// Update updates a fleet.
func (h *FleetHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")

	var req struct {
		Name        *string           `json:"name"`
		Description *string           `json:"description"`
		Labels      map[string]string `json:"labels"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	f, err := h.fleet.UpdateFleet(r.Context(), claims.OrgID, id, req.Name, req.Description, req.Labels)
	if err != nil {
		h.logger.Errorw("failed to update fleet", "fleet_id", id, "error", err)
		http.Error(w, `{"error":"failed to update fleet"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
}

// Delete deletes a fleet.
func (h *FleetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")

	if err := h.fleet.DeleteFleet(r.Context(), claims.OrgID, id); err != nil {
		h.logger.Errorw("failed to delete fleet", "fleet_id", id, "error", err)
		http.Error(w, `{"error":"failed to delete fleet"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Stats returns aggregated statistics for a fleet.
func (h *FleetHandler) Stats(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if !mw.IsValidUUID(id) {
		mw.WriteBadRequest(w, "invalid fleet ID format")
		return
	}

	stats, err := h.fleet.GetFleetStats(r.Context(), claims.OrgID, id)
	if err != nil {
		h.logger.Errorw("failed to get fleet stats", "fleet_id", id, "error", err)
		mw.WriteInternalError(w, "failed to get fleet stats")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ListDevices returns all devices in a fleet.
func (h *FleetHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if !mw.IsValidUUID(id) {
		mw.WriteBadRequest(w, "invalid fleet ID format")
		return
	}

	devices, total, err := h.fleet.ListDevices(r.Context(), claims.OrgID, domain.DeviceFilter{FleetID: id, Limit: 1000})
	if err != nil {
		h.logger.Errorw("failed to list fleet devices", "fleet_id", id, "error", err)
		mw.WriteInternalError(w, "failed to list fleet devices")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"devices": devices,
		"total":   total,
	})
}

// BulkAssign assigns devices matching labels to this fleet.
func (h *FleetHandler) BulkAssign(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if !mw.IsValidUUID(id) {
		mw.WriteBadRequest(w, "invalid fleet ID format")
		return
	}

	var req struct {
		Labels map[string]string `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.WriteBadRequest(w, "invalid request body")
		return
	}
	if len(req.Labels) == 0 {
		mw.WriteBadRequest(w, "labels are required for bulk assignment")
		return
	}

	count, err := h.fleet.BulkAssignByLabels(r.Context(), claims.OrgID, id, req.Labels)
	if err != nil {
		h.logger.Errorw("failed to bulk assign devices", "fleet_id", id, "error", err)
		mw.WriteInternalError(w, "failed to bulk assign devices")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"assigned": count,
	})
}
