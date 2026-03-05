CREATE TABLE IF NOT EXISTS drift_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(255) NOT NULL,
    model_id UUID NOT NULL REFERENCES models(id),
    feature_name VARCHAR(255) NOT NULL,
    baseline_distribution JSONB NOT NULL,
    current_distribution JSONB NOT NULL,
    psi_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    ks_statistic DOUBLE PRECISION NOT NULL DEFAULT 0,
    ks_p_value DOUBLE PRECISION NOT NULL DEFAULT 1,
    drift_detected BOOLEAN NOT NULL DEFAULT false,
    severity VARCHAR(20) NOT NULL DEFAULT 'none',
    sample_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_drift_device ON drift_reports(device_id);
CREATE INDEX idx_drift_model ON drift_reports(model_id);
CREATE INDEX idx_drift_detected ON drift_reports(drift_detected) WHERE drift_detected = true;
CREATE INDEX idx_drift_created ON drift_reports(created_at DESC);

CREATE TABLE IF NOT EXISTS drift_baselines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES models(id),
    feature_name VARCHAR(255) NOT NULL,
    distribution JSONB NOT NULL,
    sample_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(model_id, feature_name)
);
