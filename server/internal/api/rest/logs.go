package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	mw "github.com/fleetml/fleetml/server/internal/middleware"
)

// LogsHandler handles device log endpoints.
type LogsHandler struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

// NewLogsHandler creates a new logs handler.
func NewLogsHandler(db *pgxpool.Pool, logger *zap.SugaredLogger) *LogsHandler {
	return &LogsHandler{db: db, logger: logger}
}

type logEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	DeviceID    string    `json:"device_id"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	CPUPercent  *float64  `json:"cpu_percent,omitempty"`
	DiskPercent *float64  `json:"disk_percent,omitempty"`
	RAMMBUsed   *int      `json:"ram_mb_used,omitempty"`
	TempC       *float64  `json:"temperature_c,omitempty"`
}

// GetLogs returns device heartbeat history as log entries.
// GET /api/v1/devices/{device_id}/logs?limit=100&since=1h&level=warn
func (h *LogsHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "device_id")

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 1000 {
			limit = v
		}
	}

	// Parse "since" duration (e.g., "1h", "24h", "7d")
	var sinceTime *time.Time
	if since := r.URL.Query().Get("since"); since != "" {
		if d, err := parseDuration(since); err == nil {
			t := time.Now().Add(-d)
			sinceTime = &t
		}
	}

	// Build query
	query := `
		SELECT h.created_at, d.device_id, h.cpu_percent, h.disk_percent, h.ram_mb_used, h.temperature_c
		FROM heartbeats h
		JOIN devices d ON d.id = h.device_id
		WHERE d.device_id = $1`
	args := []interface{}{deviceID}
	argIdx := 2

	if sinceTime != nil {
		query += ` AND h.created_at >= $` + strconv.Itoa(argIdx)
		args = append(args, *sinceTime)
		argIdx++
	}

	query += ` ORDER BY h.created_at DESC LIMIT $` + strconv.Itoa(argIdx)
	args = append(args, limit)

	rows, err := h.db.Query(r.Context(), query, args...)
	if err != nil {
		h.logger.Errorw("failed to query logs", "device_id", deviceID, "error", err)
		mw.WriteInternalError(w, "failed to query logs")
		return
	}
	defer rows.Close()

	level := r.URL.Query().Get("level")
	var logs []logEntry
	for rows.Next() {
		var entry logEntry
		var cpuPct, diskPct, tempC *float64
		var ramUsed *int
		if err := rows.Scan(&entry.Timestamp, &entry.DeviceID, &cpuPct, &diskPct, &ramUsed, &tempC); err != nil {
			continue
		}

		entry.CPUPercent = cpuPct
		entry.DiskPercent = diskPct
		entry.RAMMBUsed = ramUsed
		entry.TempC = tempC

		// Derive level and message from metrics
		entry.Level = "info"
		entry.Message = "heartbeat"
		if cpuPct != nil && *cpuPct >= 90 {
			entry.Level = "warn"
			entry.Message = "high CPU usage"
		}
		if diskPct != nil && *diskPct >= 90 {
			entry.Level = "warn"
			entry.Message = "high disk usage"
		}
		if tempC != nil && *tempC >= 80 {
			entry.Level = "warn"
			entry.Message = "high temperature"
		}

		// Apply level filter
		if level != "" && entry.Level != level {
			continue
		}

		logs = append(logs, entry)
	}

	if logs == nil {
		logs = []logEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":      logs,
		"device_id": deviceID,
		"count":     len(logs),
	})
}

// parseDuration parses duration strings like "1h", "24h", "7d", "30m".
func parseDuration(s string) (time.Duration, error) {
	if len(s) == 0 {
		return 0, nil
	}
	last := s[len(s)-1]
	if last == 'd' {
		days, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
