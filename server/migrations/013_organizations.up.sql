-- Organizations (multi-tenancy)
CREATE TABLE IF NOT EXISTS organizations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(100) UNIQUE NOT NULL,
    plan            VARCHAR(50) NOT NULL DEFAULT 'free',  -- free, starter, pro, enterprise
    device_limit    INTEGER NOT NULL DEFAULT 5,
    fleet_limit     INTEGER NOT NULL DEFAULT 1,
    log_retention_days INTEGER NOT NULL DEFAULT 3,
    features        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Subscriptions (Dodo Payments tracking)
CREATE TABLE IF NOT EXISTS subscriptions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    dodo_subscription_id VARCHAR(255) UNIQUE,
    dodo_customer_id    VARCHAR(255),
    plan                VARCHAR(50) NOT NULL DEFAULT 'free',
    status              VARCHAR(50) NOT NULL DEFAULT 'active',  -- active, on_hold, cancelled, expired
    current_period_start TIMESTAMPTZ,
    current_period_end  TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_org_id ON subscriptions (org_id);
CREATE INDEX idx_subscriptions_dodo_id ON subscriptions (dodo_subscription_id);

-- Add org_id to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id);
CREATE INDEX IF NOT EXISTS idx_users_org_id ON users (org_id);

-- Add org_id to fleets
ALTER TABLE fleets ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id);
CREATE INDEX IF NOT EXISTS idx_fleets_org_id ON fleets (org_id);

-- Add org_id to models
ALTER TABLE models ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id);
CREATE INDEX IF NOT EXISTS idx_models_org_id ON models (org_id);

-- Add org_id to devices
ALTER TABLE devices ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id);
CREATE INDEX IF NOT EXISTS idx_devices_org_id ON devices (org_id);

-- Add org_id to deployments
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id);
CREATE INDEX IF NOT EXISTS idx_deployments_org_id ON deployments (org_id);
