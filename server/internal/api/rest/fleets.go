package rest

import (
	"encoding/json"
	"net/http"

	"github.com/fleetml/fleetml/server/internal/fleet"
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

	f, err := h.fleet.CreateFleet(r.Context(), req.Name, req.Description, req.Labels)
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
	fleets, err := h.fleet.ListFleets(r.Context())
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
	id := chi.URLParam(r, "id")

	f, err := h.fleet.GetFleet(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"fleet not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
}

// Update updates a fleet.
func (h *FleetHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	f, err := h.fleet.UpdateFleet(r.Context(), id, req.Name, req.Description, req.Labels)
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
	id := chi.URLParam(r, "id")

	if err := h.fleet.DeleteFleet(r.Context(), id); err != nil {
		h.logger.Errorw("failed to delete fleet", "fleet_id", id, "error", err)
		http.Error(w, `{"error":"failed to delete fleet"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
