package deploy

import (
	"testing"

	"github.com/fleetml/fleetml/server/internal/domain"
)

func TestDevicesForStage_BasicPercentages(t *testing.T) {
	cm := &CanaryManager{}

	tests := []struct {
		total    int
		percent  int
		expected int
	}{
		{100, 5, 5},
		{100, 50, 50},
		{100, 100, 100},
		{10, 5, 1},   // ceil(0.5) = 1
		{3, 50, 2},   // ceil(1.5) = 2
		{1, 100, 1},  // single device
		{20, 10, 2},  // ceil(2.0) = 2
		{7, 33, 3},   // ceil(2.31) = 3
	}

	for _, tt := range tests {
		stage := domain.CanaryStage{Percent: tt.percent}
		got := cm.DevicesForStage(tt.total, stage)
		if got != tt.expected {
			t.Errorf("DevicesForStage(%d, %d%%): expected %d, got %d",
				tt.total, tt.percent, tt.expected, got)
		}
	}
}

func TestDevicesForStage_MinimumOne(t *testing.T) {
	cm := &CanaryManager{}

	// Even 1% of 1 device should give 1
	got := cm.DevicesForStage(1, domain.CanaryStage{Percent: 1})
	if got != 1 {
		t.Errorf("expected minimum 1 device, got %d", got)
	}
}

func TestDevicesForStage_NeverExceedsTotal(t *testing.T) {
	cm := &CanaryManager{}

	// 150% should still cap at total
	got := cm.DevicesForStage(10, domain.CanaryStage{Percent: 150})
	if got != 10 {
		t.Errorf("expected max 10 devices, got %d", got)
	}
}

func TestCanaryState_Serialization(t *testing.T) {
	state := CanaryState{
		DeploymentID: "deploy-123",
		CurrentStage: 1,
		Stages: []domain.CanaryStage{
			{Percent: 5, Duration: "5m", SuccessMetric: "error_rate < 1%"},
			{Percent: 50, Duration: "10m"},
			{Percent: 100, Duration: "15m"},
		},
		TotalDevices: 100,
	}

	if state.CurrentStage != 1 {
		t.Error("expected current stage 1")
	}
	if len(state.Stages) != 3 {
		t.Errorf("expected 3 stages, got %d", len(state.Stages))
	}
	if state.Stages[0].Percent != 5 {
		t.Errorf("expected stage 0 percent 5, got %d", state.Stages[0].Percent)
	}
}

func TestCanaryStages_TypicalConfig(t *testing.T) {
	cm := &CanaryManager{}

	// Typical 5% -> 50% -> 100% for 20 devices
	config := &domain.CanaryConfig{
		Stages: []domain.CanaryStage{
			{Percent: 5, Duration: "5m"},
			{Percent: 50, Duration: "10m"},
			{Percent: 100, Duration: "15m"},
		},
	}

	total := 20
	stage0Count := cm.DevicesForStage(total, config.Stages[0])
	stage1Count := cm.DevicesForStage(total, config.Stages[1])
	stage2Count := cm.DevicesForStage(total, config.Stages[2])

	if stage0Count != 1 {
		t.Errorf("stage 0: expected 1, got %d", stage0Count)
	}
	if stage1Count != 10 {
		t.Errorf("stage 1: expected 10, got %d", stage1Count)
	}
	if stage2Count != 20 {
		t.Errorf("stage 2: expected 20, got %d", stage2Count)
	}

	// Ensure progressive increase
	if !(stage0Count < stage1Count && stage1Count <= stage2Count) {
		t.Error("canary stages should progressively increase")
	}
}
