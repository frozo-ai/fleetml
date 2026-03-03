package monitor

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// AlertEvaluator checks for offline devices and other alert conditions.
type AlertEvaluator struct {
	db               *pgxpool.Pool
	offlineThreshold time.Duration
	logger           *zap.SugaredLogger
}

func NewAlertEvaluator(db *pgxpool.Pool, offlineThreshold time.Duration, logger *zap.SugaredLogger) *AlertEvaluator {
	return &AlertEvaluator{
		db:               db,
		offlineThreshold: offlineThreshold,
		logger:           logger,
	}
}

// Start runs the alert evaluation loop.
func (a *AlertEvaluator) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.evaluate(ctx)
		}
	}
}

func (a *AlertEvaluator) evaluate(ctx context.Context) {
	threshold := time.Now().Add(-a.offlineThreshold)

	// Mark devices offline if no heartbeat within threshold
	result, err := a.db.Exec(ctx, `
		UPDATE devices SET status = 'offline', updated_at = NOW()
		WHERE status IN ('healthy', 'warning')
		AND last_heartbeat IS NOT NULL
		AND last_heartbeat < $1`, threshold,
	)
	if err != nil {
		a.logger.Errorw("failed to evaluate offline devices", "error", err)
		return
	}

	if result.RowsAffected() > 0 {
		a.logger.Infow("devices marked offline",
			"count", result.RowsAffected(),
			"threshold", a.offlineThreshold,
		)
	}
}
