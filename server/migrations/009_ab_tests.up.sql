CREATE TABLE IF NOT EXISTS ab_tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    model_a_id UUID NOT NULL REFERENCES models(id),
    model_b_id UUID NOT NULL REFERENCES models(id),
    split_a INTEGER NOT NULL DEFAULT 80,
    split_b INTEGER NOT NULL DEFAULT 20,
    target_fleet_id UUID REFERENCES fleets(id),
    target_labels JSONB DEFAULT '{}',
    metric VARCHAR(100) NOT NULL DEFAULT 'accuracy',
    duration INTERVAL,
    auto_promote BOOLEAN DEFAULT false,
    state VARCHAR(50) NOT NULL DEFAULT 'pending',
    winner VARCHAR(10), -- 'a', 'b', or NULL
    model_a_metrics JSONB DEFAULT '{}',
    model_b_metrics JSONB DEFAULT '{}',
    started_at TIMESTAMPTZ,
    stopped_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    CONSTRAINT valid_split CHECK (split_a + split_b = 100),
    CONSTRAINT valid_split_range CHECK (split_a >= 0 AND split_a <= 100 AND split_b >= 0 AND split_b <= 100),
    CONSTRAINT valid_state CHECK (state IN ('pending', 'running', 'completed', 'stopped'))
);

CREATE INDEX idx_ab_tests_state ON ab_tests(state);
CREATE INDEX idx_ab_tests_created_at ON ab_tests(created_at DESC);
