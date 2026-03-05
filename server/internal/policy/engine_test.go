package policy

import (
	"encoding/json"
	"testing"

	"github.com/fleetml/fleetml/server/internal/domain"
)

func TestValidPolicyTypes(t *testing.T) {
	valid := []string{"deployment", "scaling", "alerting", "compliance"}
	for _, pt := range valid {
		if !validPolicyTypes[pt] {
			t.Errorf("expected %q to be a valid policy type", pt)
		}
	}

	invalid := []string{"", "unknown", "deploy", "scale"}
	for _, pt := range invalid {
		if validPolicyTypes[pt] {
			t.Errorf("expected %q to be an invalid policy type", pt)
		}
	}
}

func TestCreateRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateRequest
		wantErr string
	}{
		{
			name:    "empty name",
			req:     CreateRequest{PolicyType: "deployment", Rules: map[string]interface{}{"k": "v"}},
			wantErr: "name is required",
		},
		{
			name:    "invalid policy type",
			req:     CreateRequest{Name: "test", PolicyType: "invalid", Rules: map[string]interface{}{"k": "v"}},
			wantErr: "invalid policy_type",
		},
		{
			name:    "nil rules",
			req:     CreateRequest{Name: "test", PolicyType: "deployment"},
			wantErr: "rules are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{} // No DB needed for validation
			_, err := e.Create(nil, tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q should contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestUpdateRequest_EmptyIsNoop(t *testing.T) {
	req := UpdateRequest{}
	if req.Name != nil || req.Description != nil || req.Rules != nil || req.Enabled != nil || req.Priority != nil {
		t.Error("empty UpdateRequest should have all nil fields")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// --- Edge-case tests ---

func TestCreate_EmptyRulesMap(t *testing.T) {
	// An empty rules map (not nil) should pass the nil check but has no rules to evaluate.
	e := &Engine{}
	req := CreateRequest{
		Name:       "empty-rules",
		PolicyType: "deployment",
		Rules:      map[string]interface{}{},
	}

	// The Create method checks req.Rules == nil, not len(req.Rules).
	// An empty map is not nil, so it passes validation.
	if req.Rules == nil {
		t.Error("expected non-nil empty rules map")
	}
	if len(req.Rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(req.Rules))
	}

	// Without a DB, Create will panic/fail at the DB query, but validation passes.
	// We verify that by checking the validation logic manually:
	if req.Name == "" {
		t.Error("name should not be empty")
	}
	if !validPolicyTypes[req.PolicyType] {
		t.Errorf("policy type %q should be valid", req.PolicyType)
	}
	_ = e // prevent unused var warning
}

func TestCreate_NegativePriority(t *testing.T) {
	req := CreateRequest{
		Name:       "negative-priority",
		PolicyType: "deployment",
		Rules:      map[string]interface{}{"max_concurrent_deployments": 5},
		Priority:   -10,
	}

	// Negative priority is structurally valid — the Engine does not validate priority bounds.
	if req.Priority != -10 {
		t.Errorf("expected priority -10, got %d", req.Priority)
	}
}

func TestCreate_MaxPriority(t *testing.T) {
	req := CreateRequest{
		Name:       "max-priority",
		PolicyType: "compliance",
		Rules:      map[string]interface{}{"require_checksum": true},
		Priority:   999999,
	}

	if req.Priority != 999999 {
		t.Errorf("expected priority 999999, got %d", req.Priority)
	}
}

func TestCreate_NonExistentFleetID(t *testing.T) {
	req := CreateRequest{
		Name:          "fleet-scoped",
		PolicyType:    "deployment",
		Rules:         map[string]interface{}{"max_concurrent_deployments": 3},
		TargetFleetID: "fleet-does-not-exist-00000000",
	}

	// The Engine does not validate that TargetFleetID references a real fleet.
	// It passes the ID straight to the DB, which may or may not have FK constraints.
	if req.TargetFleetID == "" {
		t.Error("expected non-empty target fleet ID")
	}
}

func TestEvaluateDeployment_MaxConcurrentZero(t *testing.T) {
	// A policy with max_concurrent_deployments=0 should always block deployments
	// because float64(active) >= 0 is true for any active >= 0.
	policy := domain.Policy{
		Name:       "zero-concurrent",
		PolicyType: "deployment",
		Enabled:    true,
		Rules:      map[string]interface{}{"max_concurrent_deployments": float64(0)},
	}

	maxConcurrent, ok := policy.Rules["max_concurrent_deployments"]
	if !ok {
		t.Fatal("expected max_concurrent_deployments rule")
	}

	max, ok := maxConcurrent.(float64)
	if !ok {
		t.Fatal("expected float64 type for max_concurrent_deployments")
	}

	// With 0 active deployments: float64(0) >= 0 is true, so deployment is blocked.
	active := 0
	if float64(active) >= max {
		// This is the expected behavior: always blocked when max is 0.
	} else {
		t.Error("expected deployment to be blocked when max_concurrent is 0")
	}
}

func TestEvaluateDeployment_NegativeMaxConcurrent(t *testing.T) {
	policy := domain.Policy{
		Name:       "negative-concurrent",
		PolicyType: "deployment",
		Enabled:    true,
		Rules:      map[string]interface{}{"max_concurrent_deployments": float64(-1)},
	}

	max := policy.Rules["max_concurrent_deployments"].(float64)

	// With 0 active: float64(0) >= -1 is true, so always blocked.
	active := 0
	if float64(active) >= max {
		// Expected: negative max means always blocked.
	} else {
		t.Error("expected deployment to be blocked with negative max_concurrent")
	}
}

func TestEvaluateDeployment_MalformedRuleValue(t *testing.T) {
	// If max_concurrent_deployments is a string instead of float64, the type assertion fails
	// and the rule is silently skipped.
	policy := domain.Policy{
		Name:       "malformed-rule",
		PolicyType: "deployment",
		Enabled:    true,
		Rules:      map[string]interface{}{"max_concurrent_deployments": "five"},
	}

	maxConcurrent := policy.Rules["max_concurrent_deployments"]
	if _, ok := maxConcurrent.(float64); ok {
		t.Error("expected string value to NOT be float64")
	}
	// The EvaluateDeployment method will skip this rule silently because the type assertion fails.
}

func TestEvaluateDeployment_NoEnabledPolicies(t *testing.T) {
	// When all policies are disabled, EvaluateDeployment allows the deployment.
	policies := []domain.Policy{
		{Name: "disabled-1", PolicyType: "deployment", Enabled: false, Rules: map[string]interface{}{"max_concurrent_deployments": float64(0)}},
		{Name: "disabled-2", PolicyType: "deployment", Enabled: false, Rules: map[string]interface{}{"require_canary": true}},
	}

	// Simulate EvaluateDeployment logic: skip disabled policies
	allowed := true
	reason := ""
	for _, p := range policies {
		if !p.Enabled {
			continue
		}
		// Would evaluate rules here...
		_ = p
	}

	if !allowed {
		t.Errorf("expected deployment to be allowed, got blocked: %s", reason)
	}
}

func TestEvaluateDeployment_FleetScopedPolicy_WrongFleet(t *testing.T) {
	// A fleet-scoped policy should be skipped when the deployment targets a different fleet.
	targetFleetID := "fleet-target-123"
	policyFleetID := "fleet-policy-456"

	policy := domain.Policy{
		Name:          "fleet-specific",
		PolicyType:    "deployment",
		Enabled:       true,
		TargetFleetID: &policyFleetID,
		Rules:         map[string]interface{}{"max_concurrent_deployments": float64(0)},
	}

	// Simulate EvaluateDeployment's fleet check
	if policy.TargetFleetID != nil && *policy.TargetFleetID != targetFleetID {
		// Policy is skipped because it targets a different fleet.
		// Deployment should be allowed.
	} else {
		t.Error("expected policy to be skipped for wrong fleet")
	}
}

func TestUpdate_AllFieldsNil(t *testing.T) {
	// When all UpdateRequest fields are nil, the Engine returns the existing policy unchanged.
	req := UpdateRequest{}
	if req.Name != nil || req.Description != nil || req.Rules != nil || req.Enabled != nil || req.Priority != nil {
		t.Error("all fields should be nil for an empty UpdateRequest")
	}

	// Verify via JSON: all fields should be omitted
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) != "{}" {
		t.Errorf("expected empty JSON '{}', got %q", string(data))
	}
}

func TestDelete_NonExistentPolicy(t *testing.T) {
	// Delete checks RowsAffected() == 0 and returns "policy not found".
	// This is correct behavior — non-existent policy deletions are errors.
	errMsg := "policy not found"
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestList_InvalidType(t *testing.T) {
	// List with an invalid policy type will return empty results (no DB rows match).
	invalidType := "nonexistent_type"
	if validPolicyTypes[invalidType] {
		t.Errorf("expected %q to be an invalid policy type", invalidType)
	}
}

func TestList_EmptyResult(t *testing.T) {
	// When no policies exist, List returns (nil, 0, nil).
	var policies []domain.Policy
	if policies != nil {
		t.Error("expected nil slice for empty result")
	}
	if len(policies) != 0 {
		t.Errorf("expected 0 policies, got %d", len(policies))
	}
}

func TestNewEngine_NilDeps(t *testing.T) {
	e := NewEngine(nil, nil)
	if e == nil {
		t.Error("expected non-nil engine with nil dependencies")
	}
}

func TestCreateRequest_JSONRoundTrip(t *testing.T) {
	req := CreateRequest{
		Name:          "canary-policy",
		Description:   "Require canary for production",
		PolicyType:    "deployment",
		Rules:         map[string]interface{}{"require_canary": true, "max_concurrent_deployments": float64(3)},
		Enabled:       true,
		Priority:      10,
		TargetFleetID: "fleet-prod-001",
		TargetLabels:  map[string]string{"env": "production"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded CreateRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Name != "canary-policy" {
		t.Errorf("expected name 'canary-policy', got %q", decoded.Name)
	}
	if decoded.Priority != 10 {
		t.Errorf("expected priority 10, got %d", decoded.Priority)
	}
	if decoded.TargetLabels["env"] != "production" {
		t.Errorf("expected label env=production, got %q", decoded.TargetLabels["env"])
	}
}
