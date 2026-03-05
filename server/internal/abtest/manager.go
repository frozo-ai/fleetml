package abtest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/fleetml/fleetml/server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Manager handles A/B test lifecycle.
type Manager struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

// NewManager creates a new A/B test manager.
func NewManager(db *pgxpool.Pool, logger *zap.SugaredLogger) *Manager {
	return &Manager{db: db, logger: logger}
}

// CreateRequest defines the input for creating an A/B test.
type CreateRequest struct {
	Name          string            `json:"name"`
	ModelAID      string            `json:"model_a_id"`
	ModelBID      string            `json:"model_b_id"`
	SplitA        int               `json:"split_a"`
	SplitB        int               `json:"split_b"`
	TargetFleetID string            `json:"target_fleet_id,omitempty"`
	TargetLabels  map[string]string `json:"target_labels,omitempty"`
	Metric        string            `json:"metric"`
	Duration      string            `json:"duration,omitempty"`
	AutoPromote   bool              `json:"auto_promote"`
}

// Create creates a new A/B test.
func (m *Manager) Create(ctx context.Context, req CreateRequest) (*domain.ABTest, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.ModelAID == "" || req.ModelBID == "" {
		return nil, errors.New("model_a_id and model_b_id are required")
	}
	if req.ModelAID == req.ModelBID {
		return nil, errors.New("model_a_id and model_b_id must be different")
	}

	// Default split
	if req.SplitA == 0 && req.SplitB == 0 {
		req.SplitA = 80
		req.SplitB = 20
	}
	if req.SplitA+req.SplitB != 100 {
		return nil, fmt.Errorf("split must sum to 100 (got %d + %d = %d)", req.SplitA, req.SplitB, req.SplitA+req.SplitB)
	}
	if req.Metric == "" {
		req.Metric = "accuracy"
	}

	// Verify both models exist
	for _, modelID := range []string{req.ModelAID, req.ModelBID} {
		var exists bool
		err := m.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM models WHERE id = $1)", modelID).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("checking model %s: %w", modelID, err)
		}
		if !exists {
			return nil, fmt.Errorf("model %s not found", modelID)
		}
	}

	labelsJSON, _ := json.Marshal(req.TargetLabels)

	var fleetID *string
	if req.TargetFleetID != "" {
		fleetID = &req.TargetFleetID
	}

	var duration *string
	if req.Duration != "" {
		duration = &req.Duration
	}

	test := &domain.ABTest{}
	err := m.db.QueryRow(ctx,
		`INSERT INTO ab_tests (name, model_a_id, model_b_id, split_a, split_b, target_fleet_id,
			target_labels, metric, duration, auto_promote, state)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::interval, $10, 'running')
		RETURNING id, name, model_a_id, model_b_id, split_a, split_b, target_fleet_id,
			metric, auto_promote, state, created_at`,
		req.Name, req.ModelAID, req.ModelBID, req.SplitA, req.SplitB,
		fleetID, labelsJSON, req.Metric, duration, req.AutoPromote,
	).Scan(
		&test.ID, &test.Name, &test.ModelAID, &test.ModelBID,
		&test.SplitA, &test.SplitB, &test.TargetFleetID,
		&test.Metric, &test.AutoPromote, &test.State, &test.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating ab test: %w", err)
	}

	now := time.Now()
	test.StartedAt = &now
	test.TargetLabels = req.TargetLabels

	m.logger.Infow("A/B test created",
		"id", test.ID,
		"name", test.Name,
		"model_a", req.ModelAID,
		"model_b", req.ModelBID,
		"split", fmt.Sprintf("%d/%d", req.SplitA, req.SplitB),
	)

	return test, nil
}

// Get retrieves an A/B test by ID.
func (m *Manager) Get(ctx context.Context, id string) (*domain.ABTest, error) {
	test := &domain.ABTest{}
	var labelsJSON, metricsAJSON, metricsBJSON []byte

	err := m.db.QueryRow(ctx,
		`SELECT id, name, model_a_id, model_b_id, split_a, split_b, target_fleet_id,
			target_labels, metric, auto_promote, state, winner,
			model_a_metrics, model_b_metrics, started_at, stopped_at, created_at, created_by
		FROM ab_tests WHERE id = $1`, id,
	).Scan(
		&test.ID, &test.Name, &test.ModelAID, &test.ModelBID,
		&test.SplitA, &test.SplitB, &test.TargetFleetID,
		&labelsJSON, &test.Metric, &test.AutoPromote, &test.State, &test.Winner,
		&metricsAJSON, &metricsBJSON, &test.StartedAt, &test.StoppedAt,
		&test.CreatedAt, &test.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("ab test not found: %w", err)
	}

	json.Unmarshal(labelsJSON, &test.TargetLabels)
	json.Unmarshal(metricsAJSON, &test.ModelAMetrics)
	json.Unmarshal(metricsBJSON, &test.ModelBMetrics)

	return test, nil
}

// List returns all A/B tests, optionally filtered by state.
func (m *Manager) List(ctx context.Context, state string) ([]domain.ABTest, int, error) {
	query := `SELECT id, name, model_a_id, model_b_id, split_a, split_b,
		target_fleet_id, metric, auto_promote, state, winner,
		started_at, stopped_at, created_at
		FROM ab_tests`
	args := []interface{}{}

	if state != "" {
		query += " WHERE state = $1"
		args = append(args, state)
	}
	query += " ORDER BY created_at DESC"

	rows, err := m.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing ab tests: %w", err)
	}
	defer rows.Close()

	var tests []domain.ABTest
	for rows.Next() {
		var t domain.ABTest
		err := rows.Scan(
			&t.ID, &t.Name, &t.ModelAID, &t.ModelBID,
			&t.SplitA, &t.SplitB, &t.TargetFleetID,
			&t.Metric, &t.AutoPromote, &t.State, &t.Winner,
			&t.StartedAt, &t.StoppedAt, &t.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning ab test: %w", err)
		}
		tests = append(tests, t)
	}

	return tests, len(tests), nil
}

// Stop stops a running A/B test.
func (m *Manager) Stop(ctx context.Context, id string, winner string) (*domain.ABTest, error) {
	// Validate winner
	if winner != "" && winner != "a" && winner != "b" {
		return nil, errors.New("winner must be 'a', 'b', or empty")
	}

	var winnerPtr *string
	if winner != "" {
		winnerPtr = &winner
	}

	now := time.Now()
	tag, err := m.db.Exec(ctx,
		`UPDATE ab_tests SET state = 'stopped', winner = $2, stopped_at = $3
		WHERE id = $1 AND state = 'running'`,
		id, winnerPtr, now,
	)
	if err != nil {
		return nil, fmt.Errorf("stopping ab test: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, errors.New("A/B test not found or not running")
	}

	m.logger.Infow("A/B test stopped", "id", id, "winner", winner)
	return m.Get(ctx, id)
}

// RecordMetrics updates the metrics for a specific variant in an A/B test.
func (m *Manager) RecordMetrics(ctx context.Context, id string, variant string, metrics map[string]interface{}) error {
	column := "model_a_metrics"
	if variant == "b" {
		column = "model_b_metrics"
	}

	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("marshaling metrics: %w", err)
	}

	_, err = m.db.Exec(ctx,
		fmt.Sprintf(`UPDATE ab_tests SET %s = $2 WHERE id = $1 AND state = 'running'`, column),
		id, metricsJSON,
	)
	return err
}
