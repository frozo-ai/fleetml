package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fleetml/fleetml/server/internal/auth"
	"github.com/fleetml/fleetml/server/internal/domain"
	"github.com/fleetml/fleetml/server/internal/fleet"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// DeviceHandler handles device-related endpoints.
type DeviceHandler struct {
	fleet  *fleet.Manager
	logger *zap.SugaredLogger
}

func NewDeviceHandler(fleetMgr *fleet.Manager, logger *zap.SugaredLogger) *DeviceHandler {
	return &DeviceHandler{fleet: fleetMgr, logger: logger}
}

// List lists devices with optional filters.
func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	filter := domain.DeviceFilter{
		Status:  r.URL.Query().Get("status"),
		FleetID: r.URL.Query().Get("fleet_id"),
		Runtime: r.URL.Query().Get("runtime"),
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		filter.Limit, _ = strconv.Atoi(l)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		filter.Offset, _ = strconv.Atoi(o)
	}

	devices, total, err := h.fleet.ListDevices(r.Context(), claims.OrgID, filter)
	if err != nil {
		http.Error(w, `{"error":"failed to list devices"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"devices": devices,
		"total":   total,
		"limit":   filter.Limit,
		"offset":  filter.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Get returns a single device.
func (h *DeviceHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	deviceID := chi.URLParam(r, "device_id")

	device, err := h.fleet.GetDevice(r.Context(), claims.OrgID, deviceID)
	if err != nil {
		http.Error(w, `{"error":"device not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(device)
}

// Update updates device properties.
func (h *DeviceHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	deviceID := chi.URLParam(r, "device_id")

	var req struct {
		Name    string            `json:"name"`
		Labels  map[string]string `json:"labels"`
		FleetID string            `json:"fleet_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.FleetID != "" {
		if err := h.fleet.AssignDeviceToFleet(r.Context(), claims.OrgID, deviceID, req.FleetID); err != nil {
			http.Error(w, `{"error":"failed to assign fleet"}`, http.StatusInternalServerError)
			return
		}
	}

	device, err := h.fleet.GetDevice(r.Context(), claims.OrgID, deviceID)
	if err != nil {
		http.Error(w, `{"error":"device not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(device)
}

// Delete decommissions a device.
func (h *DeviceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || claims.OrgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	deviceID := chi.URLParam(r, "device_id")

	// Verify device belongs to org before decommissioning
	if _, err := h.fleet.GetDevice(r.Context(), claims.OrgID, deviceID); err != nil {
		http.Error(w, `{"error":"device not found"}`, http.StatusNotFound)
		return
	}

	// Set status to decommissioned
	h.fleet.UpdateDeviceStatus(r.Context(), deviceID, "decommissioned", nil, nil, nil, nil, nil, nil)

	w.WriteHeader(http.StatusNoContent)
}
