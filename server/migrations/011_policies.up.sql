CREATE TABLE IF NOT EXISTS policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    policy_type VARCHAR(50) NOT NULL DEFAULT 'deployment',
    rules JSONB NOT NULL DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 0,
    target_fleet_id UUID REFERENCES fleets(id),
    target_labels JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    CONSTRAINT valid_policy_type CHECK (policy_type IN ('deployment', 'scaling', 'alerting', 'compliance'))
);

CREATE INDEX idx_policies_type ON policies(policy_type);
CREATE INDEX idx_policies_enabled ON policies(enabled) WHERE enabled = true;
CREATE INDEX idx_policies_name ON policies(name);
