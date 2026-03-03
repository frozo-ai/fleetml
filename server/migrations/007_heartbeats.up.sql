CREATE TABLE heartbeats (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id   UUID NOT NULL REFERENCES devices(id),

    -- System metrics
    cpu_percent     REAL,
    gpu_percent     REAL,
    ram_mb_used     INTEGER,
    disk_percent    REAL,
    temperature_c   REAL,
    uptime_hours    REAL,

    -- Model metrics (JSONB array)
    model_metrics   JSONB,

    received_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_heartbeats_device_id ON heartbeats(device_id);
CREATE INDEX idx_heartbeats_received_at ON heartbeats(received_at DESC);
