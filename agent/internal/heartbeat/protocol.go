package heartbeat

import (
	"context"
	"time"

	"github.com/fleetml/fleetml/agent/internal/communication"
	"github.com/fleetml/fleetml/agent/internal/health"
	"go.uber.org/zap"
)

// Scheduler manages periodic heartbeat sending.
type Scheduler struct {
	deviceID     string
	interval     time.Duration
	communicator communication.Communicator
	reporter     *health.Reporter
	logger       *zap.SugaredLogger
	commandCh    chan communication.Command
}

func NewScheduler(
	deviceID string,
	interval time.Duration,
	comm communication.Communicator,
	reporter *health.Reporter,
	logger *zap.SugaredLogger,
) *Scheduler {
	return &Scheduler{
		deviceID:     deviceID,
		interval:     interval,
		communicator: comm,
		reporter:     reporter,
		logger:       logger,
		commandCh:    make(chan communication.Command, 100),
	}
}

// Start begins the heartbeat loop. Blocks until context is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.logger.Infow("heartbeat scheduler started",
		"device_id", s.deviceID,
		"interval", s.interval,
	)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("heartbeat scheduler stopped")
			return
		case <-ticker.C:
			s.sendHeartbeat(ctx)
		}
	}
}

// Commands returns the channel for receiving commands from the server.
func (s *Scheduler) Commands() <-chan communication.Command {
	return s.commandCh
}

func (s *Scheduler) sendHeartbeat(ctx context.Context) {
	// Collect system metrics
	metrics, err := s.reporter.Collect()
	if err != nil {
		s.logger.Warnw("failed to collect metrics", "error", err)
		metrics = &health.SystemMetrics{} // Send empty metrics
	}

	// Determine status based on metrics
	status := "healthy"
	if metrics.CPUPercent > 90 || metrics.DiskPercent > 90 {
		status = "warning"
	}

	// Send heartbeat
	commands, err := s.communicator.SendHeartbeat(ctx, s.deviceID, status, metrics)
	if err != nil {
		s.logger.Warnw("failed to send heartbeat", "error", err)
		return
	}

	// Forward any received commands
	for _, cmd := range commands {
		select {
		case s.commandCh <- cmd:
		default:
			s.logger.Warnw("command channel full, dropping command",
				"command_id", cmd.ID,
				"type", cmd.Type,
			)
		}
	}
}
