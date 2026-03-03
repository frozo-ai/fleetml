package monitor

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// AuditLogger records audit events.
type AuditLogger struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewAuditLogger(db *pgxpool.Pool, logger *zap.SugaredLogger) *AuditLogger {
	return &AuditLogger{db: db, logger: logger}
}

// Log records an audit event.
func (a *AuditLogger) Log(ctx context.Context, action, actorID, actorType, resourceType, resourceID string, details map[string]interface{}) {
	detailsJSON, _ := json.Marshal(details)

	_, err := a.db.Exec(ctx, `
		INSERT INTO audit_log (action, actor_id, actor_type, resource_type, resource_id, details)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		action, nilIfEmpty(actorID), actorType, resourceType, nilIfEmpty(resourceID), detailsJSON,
	)
	if err != nil {
		a.logger.Errorw("failed to write audit log", "error", err, "action", action)
	}
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
