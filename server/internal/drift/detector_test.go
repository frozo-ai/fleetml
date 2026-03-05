package drift

import (
	"math"
	"testing"
)

func TestCalculatePSI_IdenticalDistributions(t *testing.T) {
	dist := Distribution{
		Bins:   []float64{0, 10, 20, 30, 40, 50},
		Counts: []int{20, 30, 25, 15, 10},
		Total:  100,
	}

	psi := CalculatePSI(dist, dist)
	if psi > 0.001 {
		t.Errorf("PSI for identical distributions should be ~0, got %f", psi)
	}
}

func TestCalculatePSI_SlightDrift(t *testing.T) {
	baseline := Distribution{
		Bins:   []float64{0, 10, 20, 30, 40, 50},
		Counts: []int{20, 30, 25, 15, 10},
		Total:  100,
	}
	current := Distribution{
		Bins:   []float64{0, 10, 20, 30, 40, 50},
		Counts: []int{18, 28, 27, 17, 10},
		Total:  100,
	}

	psi := CalculatePSI(baseline, current)
	if psi >= PSIThresholdLow {
		t.Errorf("slight drift PSI should be < %f, got %f", PSIThresholdLow, psi)
	}
}

func TestCalculatePSI_SignificantDrift(t *testing.T) {
	baseline := Distribution{
		Bins:   []float64{0, 10, 20, 30, 40, 50},
		Counts: []int{20, 30, 25, 15, 10},
		Total:  100,
	}
	current := Distribution{
		Bins:   []float64{0, 10, 20, 30, 40, 50},
		Counts: []int{5, 10, 15, 30, 40},
		Total:  100,
	}

	psi := CalculatePSI(baseline, current)
	if psi < PSIThresholdMedium {
		t.Errorf("significant drift PSI should be >= %f, got %f", PSIThresholdMedium, psi)
	}
}

func TestCalculatePSI_EmptyDistributions(t *testing.T) {
	empty := Distribution{}
	normal := Distribution{
		Bins:   []float64{0, 10},
		Counts: []int{100},
		Total:  100,
	}

	if psi := CalculatePSI(empty, normal); psi != 0 {
		t.Errorf("PSI with empty baseline should be 0, got %f", psi)
	}
	if psi := CalculatePSI(normal, empty); psi != 0 {
		t.Errorf("PSI with empty current should be 0, got %f", psi)
	}
}

func TestCalculateKS_IdenticalSamples(t *testing.T) {
	samples := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	d, p := CalculateKS(samples, samples)
	if d > 0.001 {
		t.Errorf("KS stat for identical samples should be ~0, got %f", d)
	}
	if p < 0.5 {
		t.Errorf("KS p-value for identical samples should be high, got %f", p)
	}
}

func TestCalculateKS_DifferentDistributions(t *testing.T) {
	baseline := make([]float64, 100)
	current := make([]float64, 100)

	// Baseline: normal around 50
	for i := 0; i < 100; i++ {
		baseline[i] = float64(i)
	}
	// Current: shifted to 70-170
	for i := 0; i < 100; i++ {
		current[i] = float64(i) + 70
	}

	d, p := CalculateKS(baseline, current)
	if d < 0.5 {
		t.Errorf("KS stat for shifted distributions should be high, got %f", d)
	}
	if p > KSPValueThreshold {
		t.Errorf("KS p-value for shifted distributions should be low, got %f", p)
	}
}

func TestCalculateKS_EmptySamples(t *testing.T) {
	normal := []float64{1, 2, 3}

	d, p := CalculateKS(nil, normal)
	if d != 0 || p != 1 {
		t.Errorf("KS with empty baseline: expected (0, 1), got (%f, %f)", d, p)
	}

	d, p = CalculateKS(normal, nil)
	if d != 0 || p != 1 {
		t.Errorf("KS with empty current: expected (0, 1), got (%f, %f)", d, p)
	}
}

func TestClassifySeverity(t *testing.T) {
	tests := []struct {
		psi      float64
		expected string
	}{
		{0.0, "none"},
		{0.05, "none"},
		{0.09, "none"},
		{0.10, "low"},
		{0.15, "low"},
		{0.20, "medium"},
		{0.24, "medium"},
		{0.25, "high"},
		{0.50, "high"},
		{1.0, "high"},
	}

	for _, tt := range tests {
		result := ClassifySeverity(tt.psi)
		if result != tt.expected {
			t.Errorf("PSI=%.2f: expected %q, got %q", tt.psi, tt.expected, result)
		}
	}
}

func TestBuildDistribution_AutoBins(t *testing.T) {
	samples := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	dist := BuildDistribution(samples, nil)

	if len(dist.Bins) != 11 { // 10 bins + 1 edge
		t.Errorf("expected 11 bin edges, got %d", len(dist.Bins))
	}
	if len(dist.Counts) != 10 {
		t.Errorf("expected 10 count bins, got %d", len(dist.Counts))
	}
	if dist.Total != 10 {
		t.Errorf("expected total 10, got %d", dist.Total)
	}

	// Sum of counts should equal total
	sum := 0
	for _, c := range dist.Counts {
		sum += c
	}
	if sum != 10 {
		t.Errorf("sum of counts should be 10, got %d", sum)
	}
}

func TestBuildDistribution_WithBins(t *testing.T) {
	samples := []float64{1, 5, 10, 15, 20, 25, 30}
	bins := []float64{0, 10, 20, 30}

	dist := BuildDistribution(samples, bins)

	if len(dist.Counts) != 3 {
		t.Errorf("expected 3 bins, got %d", len(dist.Counts))
	}

	// Bin [0, 10): 1, 5 = 2
	// Bin [10, 20): 10, 15 = 2
	// Bin [20, 30]: 20, 25, 30 = 3
	expected := []int{2, 2, 3}
	for i, c := range dist.Counts {
		if c != expected[i] {
			t.Errorf("bin %d: expected %d, got %d", i, expected[i], c)
		}
	}
}

func TestBuildDistribution_Empty(t *testing.T) {
	dist := BuildDistribution(nil, nil)
	if dist.Total != 0 {
		t.Errorf("expected total 0, got %d", dist.Total)
	}
}

func TestBuildDistribution_SingleValue(t *testing.T) {
	dist := BuildDistribution([]float64{5, 5, 5, 5}, nil)
	if dist.Total != 4 {
		t.Errorf("expected total 4, got %d", dist.Total)
	}
}

func TestKsProb_Boundaries(t *testing.T) {
	// Very small z → p-value ≈ 1
	p := ksProb(0.1)
	if p < 0.99 {
		t.Errorf("expected p ≈ 1 for z=0.1, got %f", p)
	}

	// Very large z → p-value ≈ 0
	p = ksProb(5.0)
	if p > 0.01 {
		t.Errorf("expected p ≈ 0 for z=5.0, got %f", p)
	}
}

func TestNormalize(t *testing.T) {
	counts := []int{20, 30, 50}
	probs := normalize(counts, 100)

	if math.Abs(probs[0]-0.2) > 1e-10 {
		t.Errorf("expected 0.2, got %f", probs[0])
	}
	if math.Abs(probs[1]-0.3) > 1e-10 {
		t.Errorf("expected 0.3, got %f", probs[1])
	}
	if math.Abs(probs[2]-0.5) > 1e-10 {
		t.Errorf("expected 0.5, got %f", probs[2])
	}
}

func TestReconstructSamples(t *testing.T) {
	dist := Distribution{
		Bins:   []float64{0, 10, 20},
		Counts: []int{3, 2},
		Total:  5,
	}

	samples := reconstructSamples(dist)
	if len(samples) != 5 {
		t.Errorf("expected 5 samples, got %d", len(samples))
	}

	// First 3 should be midpoint of [0, 10) = 5
	for i := 0; i < 3; i++ {
		if samples[i] != 5 {
			t.Errorf("sample %d: expected 5, got %f", i, samples[i])
		}
	}
	// Last 2 should be midpoint of [10, 20) = 15
	for i := 3; i < 5; i++ {
		if samples[i] != 15 {
			t.Errorf("sample %d: expected 15, got %f", i, samples[i])
		}
	}
}

func TestNewDetector_NilDeps(t *testing.T) {
	d := NewDetector(nil, nil)
	if d == nil {
		t.Error("expected non-nil detector")
	}
}

func TestPSI_Symmetry(t *testing.T) {
	// PSI is not symmetric, but should be non-negative in both directions
	d1 := Distribution{
		Bins:   []float64{0, 10, 20, 30},
		Counts: []int{40, 35, 25},
		Total:  100,
	}
	d2 := Distribution{
		Bins:   []float64{0, 10, 20, 30},
		Counts: []int{25, 35, 40},
		Total:  100,
	}

	psi1 := CalculatePSI(d1, d2)
	psi2 := CalculatePSI(d2, d1)

	if psi1 < 0 {
		t.Errorf("PSI should be non-negative, got %f", psi1)
	}
	if psi2 < 0 {
		t.Errorf("PSI should be non-negative, got %f", psi2)
	}
}

func TestDistribution_JSONRoundtrip(t *testing.T) {
	dist := Distribution{
		Bins:   []float64{0, 10, 20, 30},
		Counts: []int{40, 35, 25},
		Total:  100,
	}

	data, err := json.Marshal(dist)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Distribution
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Total != 100 {
		t.Errorf("expected total 100, got %d", decoded.Total)
	}
	if len(decoded.Bins) != 4 {
		t.Errorf("expected 4 bins, got %d", len(decoded.Bins))
	}
}
