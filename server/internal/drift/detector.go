package drift

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Detector performs data drift detection using PSI and KS statistics.
type Detector struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

// NewDetector creates a new drift detector.
func NewDetector(db *pgxpool.Pool, logger *zap.SugaredLogger) *Detector {
	return &Detector{db: db, logger: logger}
}

// Distribution represents a binned probability distribution.
type Distribution struct {
	Bins   []float64 `json:"bins"`
	Counts []int     `json:"counts"`
	Total  int       `json:"total"`
}

// DriftReport is the result of a drift analysis.
type DriftReport struct {
	ID                   string       `json:"id"`
	DeviceID             string       `json:"device_id"`
	ModelID              string       `json:"model_id"`
	FeatureName          string       `json:"feature_name"`
	BaselineDistribution Distribution `json:"baseline_distribution"`
	CurrentDistribution  Distribution `json:"current_distribution"`
	PSIScore             float64      `json:"psi_score"`
	KSStatistic          float64      `json:"ks_statistic"`
	KSPValue             float64      `json:"ks_p_value"`
	DriftDetected        bool         `json:"drift_detected"`
	Severity             string       `json:"severity"` // none, low, medium, high
	SampleCount          int          `json:"sample_count"`
	CreatedAt            time.Time    `json:"created_at"`
}

// PSIThresholds for drift severity.
const (
	PSIThresholdLow    = 0.1
	PSIThresholdMedium = 0.2
	PSIThresholdHigh   = 0.25
	KSPValueThreshold  = 0.05
)

// CalculatePSI computes the Population Stability Index between baseline and current distributions.
// PSI = SUM( (actual_i - expected_i) * ln(actual_i / expected_i) )
// PSI < 0.1: no significant shift
// PSI 0.1-0.25: moderate shift
// PSI > 0.25: significant shift
func CalculatePSI(baseline, current Distribution) float64 {
	if len(baseline.Counts) == 0 || len(current.Counts) == 0 {
		return 0
	}

	// Normalize to proportions
	baseProbs := normalize(baseline.Counts, baseline.Total)
	currProbs := normalize(current.Counts, current.Total)

	// Ensure same number of bins
	minLen := len(baseProbs)
	if len(currProbs) < minLen {
		minLen = len(currProbs)
	}

	psi := 0.0
	epsilon := 1e-6 // Avoid log(0)

	for i := 0; i < minLen; i++ {
		base := math.Max(baseProbs[i], epsilon)
		curr := math.Max(currProbs[i], epsilon)
		psi += (curr - base) * math.Log(curr/base)
	}

	return psi
}

// CalculateKS computes the Kolmogorov-Smirnov statistic and approximate p-value.
// Returns (D-statistic, approximate p-value).
func CalculateKS(baseline, current []float64) (float64, float64) {
	if len(baseline) == 0 || len(current) == 0 {
		return 0, 1
	}

	// Sort both samples
	b := make([]float64, len(baseline))
	c := make([]float64, len(current))
	copy(b, baseline)
	copy(c, current)
	sort.Float64s(b)
	sort.Float64s(c)

	n1 := float64(len(b))
	n2 := float64(len(c))

	// Compute D-statistic
	maxD := 0.0
	i, j := 0, 0

	for i < len(b) && j < len(c) {
		var d float64
		if b[i] <= c[j] {
			i++
			d = math.Abs(float64(i)/n1 - float64(j)/n2)
		} else {
			j++
			d = math.Abs(float64(i)/n1 - float64(j)/n2)
		}
		if d > maxD {
			maxD = d
		}
	}

	// Process remaining elements
	for i < len(b) {
		i++
		d := math.Abs(float64(i)/n1 - float64(j)/n2)
		if d > maxD {
			maxD = d
		}
	}
	for j < len(c) {
		j++
		d := math.Abs(float64(i)/n1 - float64(j)/n2)
		if d > maxD {
			maxD = d
		}
	}

	// Approximate p-value using the asymptotic formula
	en := math.Sqrt(n1 * n2 / (n1 + n2))
	pValue := ksProb(en * maxD)

	return maxD, pValue
}

// ksProb approximates the KS p-value from the D*sqrt(n) value.
func ksProb(z float64) float64 {
	if z < 0.27 {
		return 1.0
	}
	if z > 3.1 {
		return 0.0
	}

	// Marsaglia approximation
	sum := 0.0
	for k := 1; k <= 100; k++ {
		sign := 1.0
		if k%2 == 0 {
			sign = -1.0
		}
		sum += sign * math.Exp(-2*float64(k)*float64(k)*z*z)
	}
	return 2 * sum
}

// ClassifySeverity returns the drift severity based on PSI score.
func ClassifySeverity(psi float64) string {
	switch {
	case psi >= PSIThresholdHigh:
		return "high"
	case psi >= PSIThresholdMedium:
		return "medium"
	case psi >= PSIThresholdLow:
		return "low"
	default:
		return "none"
	}
}

// Analyze performs drift detection for a feature and stores the report.
func (d *Detector) Analyze(ctx context.Context, deviceID, modelID, featureName string, currentSamples []float64) (*DriftReport, error) {
	// Get baseline distribution
	var baselineJSON []byte
	err := d.db.QueryRow(ctx,
		`SELECT distribution FROM drift_baselines WHERE model_id = $1 AND feature_name = $2`,
		modelID, featureName,
	).Scan(&baselineJSON)
	if err != nil {
		return nil, fmt.Errorf("baseline not found for model %s feature %s: %w", modelID, featureName, err)
	}

	var baseline Distribution
	if err := json.Unmarshal(baselineJSON, &baseline); err != nil {
		return nil, fmt.Errorf("parsing baseline: %w", err)
	}

	// Build current distribution using baseline's bins
	current := BuildDistribution(currentSamples, baseline.Bins)

	// Calculate metrics
	psi := CalculatePSI(baseline, current)

	// For KS test, we need raw samples — reconstruct approximate baseline samples
	baselineSamples := reconstructSamples(baseline)
	ksStat, ksPValue := CalculateKS(baselineSamples, currentSamples)

	severity := ClassifySeverity(psi)
	driftDetected := psi >= PSIThresholdLow || ksPValue < KSPValueThreshold

	// Store report
	baseJSON, _ := json.Marshal(baseline)
	currJSON, _ := json.Marshal(current)

	report := &DriftReport{
		DeviceID:             deviceID,
		ModelID:              modelID,
		FeatureName:          featureName,
		BaselineDistribution: baseline,
		CurrentDistribution:  current,
		PSIScore:             psi,
		KSStatistic:          ksStat,
		KSPValue:             ksPValue,
		DriftDetected:        driftDetected,
		Severity:             severity,
		SampleCount:          len(currentSamples),
	}

	err = d.db.QueryRow(ctx,
		`INSERT INTO drift_reports (device_id, model_id, feature_name,
			baseline_distribution, current_distribution, psi_score,
			ks_statistic, ks_p_value, drift_detected, severity, sample_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at`,
		deviceID, modelID, featureName, baseJSON, currJSON,
		psi, ksStat, ksPValue, driftDetected, severity, len(currentSamples),
	).Scan(&report.ID, &report.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("storing drift report: %w", err)
	}

	if driftDetected {
		d.logger.Warnw("drift detected",
			"device", deviceID,
			"model", modelID,
			"feature", featureName,
			"psi", psi,
			"severity", severity,
		)
	}

	return report, nil
}

// SetBaseline stores or updates the baseline distribution for a model feature.
func (d *Detector) SetBaseline(ctx context.Context, modelID, featureName string, samples []float64) error {
	dist := BuildDistribution(samples, nil)
	distJSON, _ := json.Marshal(dist)

	_, err := d.db.Exec(ctx,
		`INSERT INTO drift_baselines (model_id, feature_name, distribution, sample_count)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (model_id, feature_name)
		DO UPDATE SET distribution = $3, sample_count = $4, updated_at = NOW()`,
		modelID, featureName, distJSON, len(samples),
	)
	return err
}

// ListReports returns drift reports, optionally filtered.
func (d *Detector) ListReports(ctx context.Context, modelID, deviceID string, driftOnly bool) ([]DriftReport, error) {
	query := `SELECT id, device_id, model_id, feature_name, psi_score,
		ks_statistic, ks_p_value, drift_detected, severity, sample_count, created_at
		FROM drift_reports WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if modelID != "" {
		query += fmt.Sprintf(" AND model_id = $%d", argIdx)
		args = append(args, modelID)
		argIdx++
	}
	if deviceID != "" {
		query += fmt.Sprintf(" AND device_id = $%d", argIdx)
		args = append(args, deviceID)
		argIdx++
	}
	if driftOnly {
		query += " AND drift_detected = true"
	}
	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := d.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []DriftReport
	for rows.Next() {
		var r DriftReport
		err := rows.Scan(
			&r.ID, &r.DeviceID, &r.ModelID, &r.FeatureName,
			&r.PSIScore, &r.KSStatistic, &r.KSPValue,
			&r.DriftDetected, &r.Severity, &r.SampleCount, &r.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}

	return reports, nil
}

// BuildDistribution creates a histogram from raw samples.
// If bins is nil, auto-creates 10 equal-width bins.
func BuildDistribution(samples []float64, bins []float64) Distribution {
	if len(samples) == 0 {
		return Distribution{}
	}

	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	sort.Float64s(sorted)

	if bins == nil || len(bins) < 2 {
		// Auto-create 10 bins
		min, max := sorted[0], sorted[len(sorted)-1]
		if min == max {
			max = min + 1
		}
		numBins := 10
		binWidth := (max - min) / float64(numBins)
		bins = make([]float64, numBins+1)
		for i := 0; i <= numBins; i++ {
			bins[i] = min + float64(i)*binWidth
		}
	}

	counts := make([]int, len(bins)-1)
	for _, v := range samples {
		for b := 0; b < len(bins)-1; b++ {
			if v >= bins[b] && (v < bins[b+1] || b == len(bins)-2) {
				counts[b]++
				break
			}
		}
	}

	return Distribution{
		Bins:   bins,
		Counts: counts,
		Total:  len(samples),
	}
}

func normalize(counts []int, total int) []float64 {
	probs := make([]float64, len(counts))
	for i, c := range counts {
		probs[i] = float64(c) / float64(total)
	}
	return probs
}

func reconstructSamples(d Distribution) []float64 {
	var samples []float64
	for i, count := range d.Counts {
		if i+1 < len(d.Bins) {
			mid := (d.Bins[i] + d.Bins[i+1]) / 2
			for j := 0; j < count; j++ {
				samples = append(samples, mid)
			}
		}
	}
	return samples
}
