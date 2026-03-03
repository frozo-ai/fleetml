package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var serverStartTime = time.Now()

// HealthHandler handles health-related endpoints.
type HealthHandler struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewHealthHandler(db *pgxpool.Pool, logger *zap.SugaredLogger) *HealthHandler {
	return &HealthHandler{db: db, logger: logger}
}

// Health returns the server health status.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(serverStartTime).Round(time.Second)

	dbStatus := "connected"
	if err := h.db.Ping(r.Context()); err != nil {
		dbStatus = "disconnected"
	}

	overallStatus := "healthy"
	if dbStatus != "connected" {
		overallStatus = "degraded"
	}

	resp := map[string]interface{}{
		"status":   overallStatus,
		"version":  "0.1.0",
		"database": dbStatus,
		"uptime":   uptime.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	if overallStatus != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(resp)
}

// Heartbeat handles REST heartbeat fallback.
func (h *HealthHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeviceID  string `json:"device_id"`
		Timestamp string `json:"timestamp"`
		Status    string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	h.logger.Debugw("REST heartbeat received", "device_id", req.DeviceID, "status", req.Status)

	resp := map[string]interface{}{
		"commands": []interface{}{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
