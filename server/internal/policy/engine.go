package policy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fleetml/fleetml/server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Engine manages deployment and operational policies.
type Engine struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

// NewEngine creates a new policy engine.
func NewEngine(db *pgxpool.Pool, logger *zap.SugaredLogger) *Engine {
	return &Engine{db: db, logger: logger}
}

// CreateRequest defines the input for creating a policy.
type CreateRequest struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	PolicyType    string                 `json:"policy_type"`
	Rules         map[string]interface{} `json:"rules"`
	Enabled       bool                   `json:"enabled"`
	Priority      int                    `json:"priority"`
	TargetFleetID string                 `json:"target_fleet_id,omitempty"`
	TargetLabels  map[string]string      `json:"target_labels,omitempty"`
}

// UpdateRequest defines the input for updating a policy.
type UpdateRequest struct {
	Name        *string                 `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Rules       *map[string]interface{} `json:"rules,omitempty"`
	Enabled     *bool                   `json:"enabled,omitempty"`
	Priority    *int                    `json:"priority,omitempty"`
}

var validPolicyTypes = map[string]bool{
	"deployment": true,
	"scaling":    true,
	"alerting":   true,
	"compliance": true,
}

// Create creates a new policy.
func (e *Engine) Create(ctx context.Context, req CreateRequest) (*domain.Policy, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if !validPolicyTypes[req.PolicyType] {
		return nil, fmt.Errorf("invalid policy_type: %s (must be deployment, scaling, alerting, or compliance)", req.PolicyType)
	}
	if req.Rules == nil {
		return nil, errors.New("rules are required")
	}

	rulesJSON, _ := json.Marshal(req.Rules)
	labelsJSON, _ := json.Marshal(req.TargetLabels)

	var fleetID *string
	if req.TargetFleetID != "" {
		fleetID = &req.TargetFleetID
	}

	p := &domain.Policy{}
	err := e.db.QueryRow(ctx,
		`INSERT INTO policies (name, description, policy_type, rules, enabled, priority,
			target_fleet_id, target_labels)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, description, policy_type, enabled, priority,
			target_fleet_id, created_at, updated_at`,
		req.Name, req.Description, req.PolicyType, rulesJSON, req.Enabled, req.Priority,
		fleetID, labelsJSON,
	).Scan(
		&p.ID, &p.Name, &p.Description, &p.PolicyType, &p.Enabled, &p.Priority,
		&p.TargetFleetID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating policy: %w", err)
	}

	p.Rules = req.Rules
	p.TargetLabels = req.TargetLabels

	e.logger.Infow("policy created", "id", p.ID, "name", p.Name, "type", p.PolicyType)
	return p, nil
}

// Get retrieves a policy by ID.
func (e *Engine) Get(ctx context.Context, id string) (*domain.Policy, error) {
	p := &domain.Policy{}
	var rulesJSON, labelsJSON []byte

	err := e.db.QueryRow(ctx,
		`SELECT id, name, description, policy_type, rules, enabled, priority,
			target_fleet_id, target_labels, created_at, updated_at, created_by
		FROM policies WHERE id = $1`, id,
	).Scan(
		&p.ID, &p.Name, &p.Description, &p.PolicyType, &rulesJSON, &p.Enabled,
		&p.Priority, &p.TargetFleetID, &labelsJSON, &p.CreatedAt, &p.UpdatedAt,
		&p.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("policy not found: %w", err)
	}

	json.Unmarshal(rulesJSON, &p.Rules)
	json.Unmarshal(labelsJSON, &p.TargetLabels)

	return p, nil
}

// List returns all policies, optionally filtered by type.
func (e *Engine) List(ctx context.Context, policyType string) ([]domain.Policy, int, error) {
	query := `SELECT id, name, description, policy_type, rules, enabled, priority,
		target_fleet_id, created_at, updated_at
		FROM policies`
	args := []interface{}{}

	if policyType != "" {
		query += " WHERE policy_type = $1"
		args = append(args, policyType)
	}
	query += " ORDER BY priority DESC, created_at DESC"

	rows, err := e.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing policies: %w", err)
	}
	defer rows.Close()

	var policies []domain.Policy
	for rows.Next() {
		var p domain.Policy
		var rulesJSON []byte
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.PolicyType, &rulesJSON, &p.Enabled,
			&p.Priority, &p.TargetFleetID, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		json.Unmarshal(rulesJSON, &p.Rules)
		policies = append(policies, p)
	}

	return policies, len(policies), nil
}

// Update partially updates a policy.
func (e *Engine) Update(ctx context.Context, id string, req UpdateRequest) (*domain.Policy, error) {
	// Build dynamic update query
	updates := []string{}
	args := []interface{}{id}
	argIdx := 2

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.Rules != nil {
		rulesJSON, _ := json.Marshal(*req.Rules)
		updates = append(updates, fmt.Sprintf("rules = $%d", argIdx))
		args = append(args, rulesJSON)
		argIdx++
	}
	if req.Enabled != nil {
		updates = append(updates, fmt.Sprintf("enabled = $%d", argIdx))
		args = append(args, *req.Enabled)
		argIdx++
	}
	if req.Priority != nil {
		updates = append(updates, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, *req.Priority)
		argIdx++
	}

	if len(updates) == 0 {
		return e.Get(ctx, id)
	}

	updates = append(updates, "updated_at = NOW()")
	query := "UPDATE policies SET "
	for i, u := range updates {
		if i > 0 {
			query += ", "
		}
		query += u
	}
	query += " WHERE id = $1"

	tag, err := e.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("updating policy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, errors.New("policy not found")
	}

	return e.Get(ctx, id)
}

// Delete removes a policy.
func (e *Engine) Delete(ctx context.Context, id string) error {
	tag, err := e.db.Exec(ctx, "DELETE FROM policies WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("deleting policy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("policy not found")
	}

	e.logger.Infow("policy deleted", "id", id)
	return nil
}

// EvaluateDeployment checks if a deployment is allowed by active policies.
// Returns (allowed, reason).
func (e *Engine) EvaluateDeployment(ctx context.Context, modelID, fleetID string) (bool, string) {
	policies, _, err := e.List(ctx, "deployment")
	if err != nil {
		e.logger.Errorw("failed to list policies for evaluation", "error", err)
		return true, "" // Allow if policies can't be read
	}

	for _, p := range policies {
		if !p.Enabled {
			continue
		}

		// Check if policy targets this fleet
		if p.TargetFleetID != nil && *p.TargetFleetID != fleetID {
			continue
		}

		// Evaluate rules
		if maxConcurrent, ok := p.Rules["max_concurrent_deployments"]; ok {
			if max, ok2 := maxConcurrent.(float64); ok2 {
				var active int
				e.db.QueryRow(ctx,
					`SELECT COUNT(*) FROM deployments WHERE state IN ('pending', 'rolling_out')`,
				).Scan(&active)

				if float64(active) >= max {
					return false, fmt.Sprintf("policy %q: max concurrent deployments (%d) reached", p.Name, int(max))
				}
			}
		}

		if requireCanary, ok := p.Rules["require_canary"]; ok {
			if rc, ok2 := requireCanary.(bool); ok2 && rc {
				// Caller must use canary deployment policy
				// This is a hint — actual enforcement happens in orchestrator
			}
		}
	}

	return true, ""
}
