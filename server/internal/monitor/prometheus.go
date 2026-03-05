package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// PrometheusExporter exposes fleet metrics on a /metrics endpoint
// in Prometheus text exposition format.
type PrometheusExporter struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger

	mu      sync.RWMutex
	metrics *fleetMetrics
}

type fleetMetrics struct {
	TotalDevices   int
	OnlineDevices  int
	OfflineDevices int
	TotalModels    int
	ActiveDeploys  int
	FailedDeploys  int
	HeartbeatCount int64
	CollectedAt    time.Time
}

// NewPrometheusExporter creates a new Prometheus exporter.
func NewPrometheusExporter(db *pgxpool.Pool, logger *zap.SugaredLogger) *PrometheusExporter {
	return &PrometheusExporter{
		db:      db,
		logger:  logger,
		metrics: &fleetMetrics{},
	}
}

// Start begins periodic metric collection in the background.
func (p *PrometheusExporter) Start(ctx context.Context, interval time.Duration) {
	// Collect immediately on startup
	p.collect(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.collect(ctx)
		}
	}
}

func (p *PrometheusExporter) collect(ctx context.Context) {
	m := &fleetMetrics{CollectedAt: time.Now()}

	// Total devices
	p.db.QueryRow(ctx, `SELECT COUNT(*) FROM devices`).Scan(&m.TotalDevices)

	// Online devices (heartbeat within last 90s)
	p.db.QueryRow(ctx, `SELECT COUNT(*) FROM devices WHERE status IN ('healthy', 'warning')`).Scan(&m.OnlineDevices)
	m.OfflineDevices = m.TotalDevices - m.OnlineDevices

	// Total models
	p.db.QueryRow(ctx, `SELECT COUNT(*) FROM models`).Scan(&m.TotalModels)

	// Active deployments
	p.db.QueryRow(ctx, `SELECT COUNT(*) FROM deployments WHERE state = 'rolling_out'`).Scan(&m.ActiveDeploys)

	// Failed deployments
	p.db.QueryRow(ctx, `SELECT COUNT(*) FROM deployments WHERE state = 'failed'`).Scan(&m.FailedDeploys)

	// Total heartbeats in last hour
	p.db.QueryRow(ctx, `SELECT COUNT(*) FROM heartbeats WHERE created_at > NOW() - INTERVAL '1 hour'`).Scan(&m.HeartbeatCount)

	p.mu.Lock()
	p.metrics = m
	p.mu.Unlock()
}

// Handler returns an HTTP handler that serves Prometheus metrics.
func (p *PrometheusExporter) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p.mu.RLock()
		m := p.metrics
		p.mu.RUnlock()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		fmt.Fprintf(w, "# HELP fleetml_devices_total Total number of registered devices.\n")
		fmt.Fprintf(w, "# TYPE fleetml_devices_total gauge\n")
		fmt.Fprintf(w, "fleetml_devices_total %d\n\n", m.TotalDevices)

		fmt.Fprintf(w, "# HELP fleetml_devices_online Number of online devices.\n")
		fmt.Fprintf(w, "# TYPE fleetml_devices_online gauge\n")
		fmt.Fprintf(w, "fleetml_devices_online %d\n\n", m.OnlineDevices)

		fmt.Fprintf(w, "# HELP fleetml_devices_offline Number of offline devices.\n")
		fmt.Fprintf(w, "# TYPE fleetml_devices_offline gauge\n")
		fmt.Fprintf(w, "fleetml_devices_offline %d\n\n", m.OfflineDevices)

		fmt.Fprintf(w, "# HELP fleetml_models_total Total number of registered models.\n")
		fmt.Fprintf(w, "# TYPE fleetml_models_total gauge\n")
		fmt.Fprintf(w, "fleetml_models_total %d\n\n", m.TotalModels)

		fmt.Fprintf(w, "# HELP fleetml_deployments_active Number of active deployments.\n")
		fmt.Fprintf(w, "# TYPE fleetml_deployments_active gauge\n")
		fmt.Fprintf(w, "fleetml_deployments_active %d\n\n", m.ActiveDeploys)

		fmt.Fprintf(w, "# HELP fleetml_deployments_failed Total number of failed deployments.\n")
		fmt.Fprintf(w, "# TYPE fleetml_deployments_failed gauge\n")
		fmt.Fprintf(w, "fleetml_deployments_failed %d\n\n", m.FailedDeploys)

		fmt.Fprintf(w, "# HELP fleetml_heartbeats_1h Number of heartbeats received in the last hour.\n")
		fmt.Fprintf(w, "# TYPE fleetml_heartbeats_1h gauge\n")
		fmt.Fprintf(w, "fleetml_heartbeats_1h %d\n\n", m.HeartbeatCount)

		fmt.Fprintf(w, "# HELP fleetml_metrics_collected_at Unix timestamp of last metric collection.\n")
		fmt.Fprintf(w, "# TYPE fleetml_metrics_collected_at gauge\n")
		fmt.Fprintf(w, "fleetml_metrics_collected_at %d\n", m.CollectedAt.Unix())
	}
}

// JSONHandler returns an HTTP handler that serves metrics as JSON (for dashboard).
func (p *PrometheusExporter) JSONHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p.mu.RLock()
		m := p.metrics
		p.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"devices_total":   m.TotalDevices,
			"devices_online":  m.OnlineDevices,
			"devices_offline": m.OfflineDevices,
			"models_total":    m.TotalModels,
			"deploys_active":  m.ActiveDeploys,
			"deploys_failed":  m.FailedDeploys,
			"heartbeats_1h":   m.HeartbeatCount,
			"collected_at":    m.CollectedAt,
		})
	}
}
