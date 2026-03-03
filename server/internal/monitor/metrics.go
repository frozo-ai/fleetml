package monitor

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// MetricsProcessor processes incoming heartbeats and stores metrics.
type MetricsProcessor struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewMetricsProcessor(db *pgxpool.Pool, logger *zap.SugaredLogger) *MetricsProcessor {
	return &MetricsProcessor{db: db, logger: logger}
}

// ProcessHeartbeat stores heartbeat data in the heartbeats table.
// deviceID is the string device_id (e.g. "rpi-001"), not the UUID.
func (m *MetricsProcessor) ProcessHeartbeat(ctx context.Context, deviceID string,
	cpuPct, gpuPct, diskPct, tempC, uptimeH float64, ramUsed int,
	modelMetrics interface{}) error {

	metricsJSON, _ := json.Marshal(modelMetrics)

	_, err := m.db.Exec(ctx, `
		INSERT INTO heartbeats (device_id, cpu_percent, gpu_percent, ram_mb_used,
			disk_percent, temperature_c, uptime_hours, model_metrics)
		VALUES ((SELECT id FROM devices WHERE device_id = $1), $2, $3, $4, $5, $6, $7, $8)`,
		deviceID, cpuPct, gpuPct, ramUsed, diskPct, tempC, uptimeH, metricsJSON,
	)
	if err != nil {
		return err
	}

	return nil
}
