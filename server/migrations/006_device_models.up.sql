CREATE TABLE device_models (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id       UUID NOT NULL REFERENCES devices(id),
    model_id        UUID NOT NULL REFERENCES models(id),

    status          VARCHAR(50) NOT NULL DEFAULT 'active',
    runtime         VARCHAR(50) NOT NULL,

    -- Inference metrics (latest)
    inference_count BIGINT DEFAULT 0,
    avg_latency_ms  REAL,
    p99_latency_ms  REAL,
    accuracy        REAL,
    drift_score     REAL,

    deployed_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(device_id, model_id)
);
