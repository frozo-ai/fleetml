CREATE TABLE deployments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id            UUID NOT NULL REFERENCES models(id),

    -- Target
    target_type         VARCHAR(50) NOT NULL,
    target_fleet_id     UUID REFERENCES fleets(id),
    target_device_ids   UUID[],
    target_labels       JSONB,

    -- Status
    state               VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_devices       INTEGER NOT NULL DEFAULT 0,
    completed_devices   INTEGER NOT NULL DEFAULT 0,
    failed_devices      INTEGER NOT NULL DEFAULT 0,
    queued_devices      INTEGER NOT NULL DEFAULT 0,

    -- Policy
    deployment_policy   VARCHAR(100) DEFAULT 'immediate',
    canary_config       JSONB,
    rollback_model_id   UUID REFERENCES models(id),

    -- Metadata
    error               TEXT,
    started_at          TIMESTAMP WITH TIME ZONE,
    completed_at        TIMESTAMP WITH TIME ZONE,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by          UUID REFERENCES users(id)
);

CREATE INDEX idx_deployments_state ON deployments(state);
CREATE INDEX idx_deployments_model_id ON deployments(model_id);
CREATE INDEX idx_deployments_created_at ON deployments(created_at DESC);

CREATE TABLE deployment_device_status (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id   UUID NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    device_id       UUID NOT NULL REFERENCES devices(id),

    state           VARCHAR(50) NOT NULL DEFAULT 'pending',
    error           TEXT,
    started_at      TIMESTAMP WITH TIME ZONE,
    completed_at    TIMESTAMP WITH TIME ZONE,

    UNIQUE(deployment_id, device_id)
);

CREATE INDEX idx_dds_deployment_id ON deployment_device_status(deployment_id);
CREATE INDEX idx_dds_device_id ON deployment_device_status(device_id);
