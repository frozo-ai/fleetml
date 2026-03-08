package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	ID          int64             `json:"id,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	DeviceID    string            `json:"device_id"`
	Level       string            `json:"level"`
	Component   string            `json:"component,omitempty"`
	Message     string            `json:"message"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CPUPercent  *float64          `json:"cpu_percent,omitempty"`
	DiskPercent *float64          `json:"disk_percent,omitempty"`
	RAMMBUsed   *int              `json:"ram_mb_used,omitempty"`
	TempC       *float64          `json:"temperature_c,omitempty"`
}

// GetLogs returns device logs from the device_logs table, falling back to
// heartbeat-derived entries if no structured logs exist.
// GET /api/v1/devices/{device_id}/logs?limit=100&since=1h&level=warn&follow=true
func (h *LogsHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "device_id")

	// SSE streaming mode
	if r.URL.Query().Get("follow") == "true" {
		h.streamLogs(w, r, deviceID)
		return
	}

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 1000 {
			limit = v
		}
	}

	var sinceTime *time.Time
	if since := r.URL.Query().Get("since"); since != "" {
		if d, err := parseDuration(since); err == nil {
			t := time.Now().Add(-d)
			sinceTime = &t
		}
	}

	level := r.URL.Query().Get("level")

	// Try structured logs first
	logs, err := h.queryStructuredLogs(r.Context(), deviceID, sinceTime, level, limit)
	if err != nil {
		h.logger.Errorw("failed to query structured logs", "device_id", deviceID, "error", err)
	}

	// Fall back to heartbeat-derived logs if no structured logs found
	if len(logs) == 0 {
		logs, err = h.queryHeartbeatLogs(r.Context(), deviceID, sinceTime, level, limit)
		if err != nil {
			h.logger.Errorw("failed to query heartbeat logs", "device_id", deviceID, "error", err)
			mw.WriteInternalError(w, "failed to query logs")
			return
		}
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

// IngestLogs handles POST /api/v1/devices/{device_id}/logs — agent log ingestion.
func (h *LogsHandler) IngestLogs(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "device_id")

	var req struct {
		DeviceID string `json:"device_id"`
		Entries  []struct {
			Timestamp time.Time         `json:"timestamp"`
			Level     string            `json:"level"`
			Component string            `json:"component"`
			Message   string            `json:"message"`
			Metadata  map[string]string `json:"metadata,omitempty"`
		} `json:"entries"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.WriteBadRequest(w, "invalid request body")
		return
	}

	if len(req.Entries) == 0 {
		mw.WriteBadRequest(w, "no log entries provided")
		return
	}

	// Use URL param device_id, or fallback to body
	if deviceID == "" {
		deviceID = req.DeviceID
	}
	if deviceID == "" {
		mw.WriteBadRequest(w, "device_id is required")
		return
	}

	// Batch insert
	query := `INSERT INTO device_logs (device_id, timestamp, level, component, message, metadata) VALUES `
	args := make([]interface{}, 0, len(req.Entries)*5)
	placeholders := make([]string, 0, len(req.Entries))
	argIdx := 1

	for _, entry := range req.Entries {
		ts := entry.Timestamp
		if ts.IsZero() {
			ts = time.Now()
		}
		level := entry.Level
		if level == "" {
			level = "info"
		}
		component := entry.Component
		if component == "" {
			component = "agent"
		}

		var metadataJSON []byte
		if len(entry.Metadata) > 0 {
			metadataJSON, _ = json.Marshal(entry.Metadata)
		}

		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)",
			argIdx, argIdx+1, argIdx+2, argIdx+3, argIdx+4, argIdx+5))
		args = append(args, deviceID, ts, level, component, entry.Message, metadataJSON)
		argIdx += 6
	}

	query += strings.Join(placeholders, ", ")

	if _, err := h.db.Exec(r.Context(), query, args...); err != nil {
		h.logger.Errorw("failed to insert logs", "device_id", deviceID, "error", err)
		mw.WriteInternalError(w, "failed to store logs")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ingested": len(req.Entries),
	})
}

// streamLogs implements Server-Sent Events for real-time log streaming.
func (h *LogsHandler) streamLogs(w http.ResponseWriter, r *http.Request, deviceID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		mw.WriteInternalError(w, "streaming not supported")
		return
	}

	level := r.URL.Query().Get("level")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher.Flush()

	// Poll for new logs every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Track the last seen log ID to avoid duplicates
	var lastID int64

	// Send initial batch
	logs := h.pollNewLogs(r.Context(), deviceID, lastID, level, 50)
	for _, entry := range logs {
		data, _ := json.Marshal(entry)
		fmt.Fprintf(w, "data: %s\n\n", data)
		if entry.ID > lastID {
			lastID = entry.ID
		}
	}
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			logs := h.pollNewLogs(r.Context(), deviceID, lastID, level, 50)
			for _, entry := range logs {
				data, _ := json.Marshal(entry)
				fmt.Fprintf(w, "data: %s\n\n", data)
				if entry.ID > lastID {
					lastID = entry.ID
				}
			}
			if len(logs) > 0 {
				flusher.Flush()
			}
		}
	}
}

// pollNewLogs fetches log entries with ID > lastID.
func (h *LogsHandler) pollNewLogs(ctx context.Context, deviceID string, lastID int64, level string, limit int) []logEntry {
	query := `SELECT id, timestamp, device_id, level, component, message, metadata
		FROM device_logs WHERE device_id = $1 AND id > $2`
	args := []interface{}{deviceID, lastID}
	argIdx := 3

	if level != "" {
		query += ` AND level = $` + strconv.Itoa(argIdx)
		args = append(args, level)
		argIdx++
	}

	query += ` ORDER BY id ASC LIMIT $` + strconv.Itoa(argIdx)
	args = append(args, limit)

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var logs []logEntry
	for rows.Next() {
		var entry logEntry
		var metadataJSON []byte
		if err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.DeviceID, &entry.Level, &entry.Component, &entry.Message, &metadataJSON); err != nil {
			continue
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &entry.Metadata)
		}
		logs = append(logs, entry)
	}
	return logs
}

// queryStructuredLogs queries the device_logs table.
func (h *LogsHandler) queryStructuredLogs(ctx context.Context, deviceID string, sinceTime *time.Time, level string, limit int) ([]logEntry, error) {
	query := `SELECT id, timestamp, device_id, level, component, message, metadata
		FROM device_logs WHERE device_id = $1`
	args := []interface{}{deviceID}
	argIdx := 2

	if sinceTime != nil {
		query += ` AND timestamp >= $` + strconv.Itoa(argIdx)
		args = append(args, *sinceTime)
		argIdx++
	}

	if level != "" {
		query += ` AND level = $` + strconv.Itoa(argIdx)
		args = append(args, level)
		argIdx++
	}

	query += ` ORDER BY timestamp DESC LIMIT $` + strconv.Itoa(argIdx)
	args = append(args, limit)

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []logEntry
	for rows.Next() {
		var entry logEntry
		var metadataJSON []byte
		if err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.DeviceID, &entry.Level, &entry.Component, &entry.Message, &metadataJSON); err != nil {
			continue
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &entry.Metadata)
		}
		logs = append(logs, entry)
	}
	return logs, nil
}

// queryHeartbeatLogs falls back to heartbeat-derived logs when no structured logs exist.
func (h *LogsHandler) queryHeartbeatLogs(ctx context.Context, deviceID string, sinceTime *time.Time, level string, limit int) ([]logEntry, error) {
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

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
		entry.Component = "heartbeat"

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

		if level != "" && entry.Level != level {
			continue
		}

		logs = append(logs, entry)
	}
	return logs, nil
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
