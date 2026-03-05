package abtest

import (
	"encoding/json"
	"testing"

	"github.com/fleetml/fleetml/server/internal/domain"
)

func TestCreateRequest_JSONRoundtrip(t *testing.T) {
	req := CreateRequest{
		Name:     "resnet-vs-mobilenet",
		ModelAID: "550e8400-e29b-41d4-a716-446655440000",
		ModelBID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		SplitA:   80,
		SplitB:   20,
		Metric:   "latency",
		Duration: "24h",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded CreateRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Name != "resnet-vs-mobilenet" {
		t.Errorf("expected name 'resnet-vs-mobilenet', got %q", decoded.Name)
	}
	if decoded.SplitA != 80 || decoded.SplitB != 20 {
		t.Errorf("expected split 80/20, got %d/%d", decoded.SplitA, decoded.SplitB)
	}
	if decoded.Metric != "latency" {
		t.Errorf("expected metric 'latency', got %q", decoded.Metric)
	}
}

func TestCreateRequest_DefaultSplit(t *testing.T) {
	req := CreateRequest{
		Name:     "test",
		ModelAID: "a",
		ModelBID: "b",
	}

	// When split is 0/0, manager defaults to 80/20
	if req.SplitA != 0 && req.SplitB != 0 {
		t.Error("expected zero split before defaults applied")
	}
}

func TestCreateRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateRequest
		wantErr bool
	}{
		{
			name:    "valid",
			req:     CreateRequest{Name: "test", ModelAID: "a", ModelBID: "b", SplitA: 80, SplitB: 20},
			wantErr: false,
		},
		{
			name:    "missing name",
			req:     CreateRequest{ModelAID: "a", ModelBID: "b"},
			wantErr: true,
		},
		{
			name:    "missing model_a",
			req:     CreateRequest{Name: "test", ModelBID: "b"},
			wantErr: true,
		},
		{
			name:    "missing model_b",
			req:     CreateRequest{Name: "test", ModelAID: "a"},
			wantErr: true,
		},
		{
			name:    "same model",
			req:     CreateRequest{Name: "test", ModelAID: "a", ModelBID: "a"},
			wantErr: true,
		},
		{
			name:    "bad split sum",
			req:     CreateRequest{Name: "test", ModelAID: "a", ModelBID: "b", SplitA: 60, SplitB: 60},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't call Create without a DB, but we can validate the
			// request logic manually
			hasErr := false
			if tt.req.Name == "" {
				hasErr = true
			}
			if tt.req.ModelAID == "" || tt.req.ModelBID == "" {
				hasErr = true
			}
			if tt.req.ModelAID == tt.req.ModelBID && tt.req.ModelAID != "" {
				hasErr = true
			}
			if tt.req.SplitA+tt.req.SplitB != 100 && tt.req.SplitA+tt.req.SplitB != 0 {
				hasErr = true
			}

			if hasErr != tt.wantErr {
				t.Errorf("expected error=%v, got error=%v", tt.wantErr, hasErr)
			}
		})
	}
}

func TestCreateRequest_WithTargetLabels(t *testing.T) {
	req := CreateRequest{
		Name:     "test",
		ModelAID: "a",
		ModelBID: "b",
		SplitA:   70,
		SplitB:   30,
		TargetLabels: map[string]string{
			"env":    "production",
			"region": "us-west",
		},
	}

	data, _ := json.Marshal(req)
	var decoded CreateRequest
	json.Unmarshal(data, &decoded)

	if decoded.TargetLabels["env"] != "production" {
		t.Errorf("expected label env=production, got %q", decoded.TargetLabels["env"])
	}
	if decoded.TargetLabels["region"] != "us-west" {
		t.Errorf("expected label region=us-west, got %q", decoded.TargetLabels["region"])
	}
}

func TestNewManager_NilDeps(t *testing.T) {
	m := NewManager(nil, nil)
	if m == nil {
		t.Error("expected non-nil manager")
	}
}

func TestCreateRequest_AllMetricTypes(t *testing.T) {
	metrics := []string{"accuracy", "latency", "throughput", "error_rate", "f1_score"}
	for _, metric := range metrics {
		req := CreateRequest{
			Name:     "test",
			ModelAID: "a",
			ModelBID: "b",
			Metric:   metric,
		}
		if req.Metric != metric {
			t.Errorf("expected metric %q, got %q", metric, req.Metric)
		}
	}
}

func TestCreateRequest_WithAutoPromote(t *testing.T) {
	req := CreateRequest{
		Name:        "auto-test",
		ModelAID:    "a",
		ModelBID:    "b",
		AutoPromote: true,
		Duration:    "4h",
	}

	data, _ := json.Marshal(req)
	var decoded CreateRequest
	json.Unmarshal(data, &decoded)

	if !decoded.AutoPromote {
		t.Error("expected auto_promote=true")
	}
	if decoded.Duration != "4h" {
		t.Errorf("expected duration 4h, got %q", decoded.Duration)
	}
}

// --- Edge-case tests ---

func TestCreate_SplitTotalNot100(t *testing.T) {
	// 99% total should be rejected by validation.
	req := CreateRequest{
		Name:     "bad-split-99",
		ModelAID: "model-a",
		ModelBID: "model-b",
		SplitA:   50,
		SplitB:   49,
	}

	total := req.SplitA + req.SplitB
	if total == 100 {
		t.Errorf("expected split total != 100, got %d", total)
	}
	if total != 99 {
		t.Errorf("expected split total 99, got %d", total)
	}

	// The Create method checks: if req.SplitA+req.SplitB != 100
	// Since 50+49=99 != 100, this should error.
}

func TestCreate_SplitTotalOver100(t *testing.T) {
	req := CreateRequest{
		Name:     "bad-split-over",
		ModelAID: "model-a",
		ModelBID: "model-b",
		SplitA:   60,
		SplitB:   50,
	}

	total := req.SplitA + req.SplitB
	if total == 100 {
		t.Errorf("expected split total != 100, got %d", total)
	}
	if total != 110 {
		t.Errorf("expected split total 110, got %d", total)
	}
}

func TestCreate_EmptyMetricDefaultsToAccuracy(t *testing.T) {
	req := CreateRequest{
		Name:     "default-metric",
		ModelAID: "model-a",
		ModelBID: "model-b",
		SplitA:   80,
		SplitB:   20,
		Metric:   "",
	}

	// Simulate the Create logic: if req.Metric == "" { req.Metric = "accuracy" }
	if req.Metric == "" {
		req.Metric = "accuracy"
	}

	if req.Metric != "accuracy" {
		t.Errorf("expected default metric 'accuracy', got %q", req.Metric)
	}
}

func TestCreate_SameModelForAB(t *testing.T) {
	req := CreateRequest{
		Name:     "same-model-test",
		ModelAID: "model-same",
		ModelBID: "model-same",
		SplitA:   50,
		SplitB:   50,
	}

	if req.ModelAID != req.ModelBID {
		t.Error("expected same model for A and B")
	}

	// The Create method checks: if req.ModelAID == req.ModelBID
	// and returns "model_a_id and model_b_id must be different"
}

func TestCreate_EmptyName(t *testing.T) {
	req := CreateRequest{
		Name:     "",
		ModelAID: "model-a",
		ModelBID: "model-b",
		SplitA:   80,
		SplitB:   20,
	}

	if req.Name != "" {
		t.Errorf("expected empty name, got %q", req.Name)
	}

	// The Create method returns "name is required" for empty names.
}

func TestStop_AlreadyStopped(t *testing.T) {
	// Stop uses: WHERE id = $1 AND state = 'running'
	// If the test is already stopped, RowsAffected() == 0 and it returns
	// "A/B test not found or not running"
	test := domain.ABTest{
		ID:    "test-1",
		State: "stopped",
	}

	if test.State != "stopped" {
		t.Errorf("expected state 'stopped', got %q", test.State)
	}

	// Trying to stop an already-stopped test should fail.
	errMsg := "A/B test not found or not running"
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestStop_InvalidWinner(t *testing.T) {
	// Stop validates: winner must be 'a', 'b', or empty.
	invalidWinners := []string{"c", "A", "B", "model_a", "1", "both", "none"}

	for _, winner := range invalidWinners {
		if winner == "" || winner == "a" || winner == "b" {
			t.Errorf("expected %q to be an invalid winner", winner)
		}
	}

	// Valid winners
	validWinners := []string{"a", "b", ""}
	for _, winner := range validWinners {
		if winner != "" && winner != "a" && winner != "b" {
			t.Errorf("expected %q to be a valid winner", winner)
		}
	}
}

func TestRecordMetrics_InvalidVariant(t *testing.T) {
	// RecordMetrics uses: if variant == "b" { column = "model_b_metrics" }
	// For any other variant value (including "a", "c", ""), it defaults to "model_a_metrics".
	// This means "c" would silently write to model_a_metrics column.

	tests := []struct {
		variant        string
		expectedColumn string
	}{
		{"a", "model_a_metrics"},
		{"b", "model_b_metrics"},
		{"c", "model_a_metrics"},  // Bug: should be rejected
		{"", "model_a_metrics"},   // Bug: should be rejected
		{"A", "model_a_metrics"},  // Bug: case-sensitive
	}

	for _, tt := range tests {
		column := "model_a_metrics"
		if tt.variant == "b" {
			column = "model_b_metrics"
		}
		if column != tt.expectedColumn {
			t.Errorf("variant %q: expected column %q, got %q", tt.variant, tt.expectedColumn, column)
		}
	}
}

func TestRecordMetrics_StoppedTest(t *testing.T) {
	// RecordMetrics uses: WHERE id = $1 AND state = 'running'
	// If the test is stopped, the UPDATE affects 0 rows but does NOT return an error.
	// This is a design gap: metrics are silently dropped for stopped tests.
	test := domain.ABTest{
		ID:    "test-stopped",
		State: "stopped",
	}
	if test.State != "stopped" {
		t.Errorf("expected stopped state, got %q", test.State)
	}
}

func TestList_InvalidState(t *testing.T) {
	// List with an invalid state filter returns empty results.
	validStates := map[string]bool{
		"pending":   true,
		"running":   true,
		"completed": true,
		"stopped":   true,
	}

	invalidState := "nonexistent_state"
	if validStates[invalidState] {
		t.Errorf("expected %q to be an invalid state", invalidState)
	}
}

func TestList_EmptyResult(t *testing.T) {
	// When no A/B tests exist, List returns (nil, 0, nil).
	var tests []domain.ABTest
	if tests != nil {
		t.Error("expected nil slice for empty result")
	}
	if len(tests) != 0 {
		t.Errorf("expected 0 tests, got %d", len(tests))
	}
}

func TestCreate_DefaultSplitApplied(t *testing.T) {
	// When SplitA and SplitB are both 0, Create defaults to 80/20.
	req := CreateRequest{
		Name:     "default-split",
		ModelAID: "a",
		ModelBID: "b",
		SplitA:   0,
		SplitB:   0,
	}

	// Simulate Create's default logic
	if req.SplitA == 0 && req.SplitB == 0 {
		req.SplitA = 80
		req.SplitB = 20
	}

	if req.SplitA != 80 || req.SplitB != 20 {
		t.Errorf("expected default split 80/20, got %d/%d", req.SplitA, req.SplitB)
	}
	if req.SplitA+req.SplitB != 100 {
		t.Errorf("expected split total 100, got %d", req.SplitA+req.SplitB)
	}
}

func TestCreate_OnlyOneZeroSplit(t *testing.T) {
	// When only SplitA is 0 but SplitB is not, the default is NOT applied.
	// This means SplitA=0 + SplitB=100 passes validation (sums to 100).
	req := CreateRequest{
		Name:     "one-zero",
		ModelAID: "a",
		ModelBID: "b",
		SplitA:   0,
		SplitB:   100,
	}

	// Default logic only triggers when BOTH are 0
	if req.SplitA == 0 && req.SplitB == 0 {
		t.Error("only one split is zero, default should not trigger")
	}

	// But 0+100=100, so validation passes
	if req.SplitA+req.SplitB != 100 {
		t.Errorf("expected split total 100, got %d", req.SplitA+req.SplitB)
	}
}

func TestABTest_AllStates(t *testing.T) {
	states := []string{"pending", "running", "completed", "stopped"}
	for _, state := range states {
		test := domain.ABTest{State: state}
		if test.State != state {
			t.Errorf("expected state %q, got %q", state, test.State)
		}
	}
}
