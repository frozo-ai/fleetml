CREATE TABLE devices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id       VARCHAR(255) NOT NULL UNIQUE,
    name            VARCHAR(255),
    status          VARCHAR(50) NOT NULL DEFAULT 'registered', -- registered, healthy, warning, offline, decommissioned
    arch            VARCHAR(50) NOT NULL,
    gpu_type        VARCHAR(50),
    runtime         VARCHAR(50),
    ram_mb          INTEGER NOT NULL,
    disk_gb         INTEGER NOT NULL,
    os              VARCHAR(100),
    hardware_model  VARCHAR(255),

    -- Metadata
    labels          JSONB DEFAULT '{}',
    fleet_id        UUID REFERENCES fleets(id),

    -- TLS/Auth
    certificate_fingerprint VARCHAR(255),

    -- Timestamps
    last_heartbeat  TIMESTAMP WITH TIME ZONE,
    registered_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- System metrics (latest)
    cpu_percent     REAL,
    gpu_percent     REAL,
    ram_mb_used     INTEGER,
    disk_percent    REAL,
    temperature_c   REAL,
    uptime_hours    REAL
);

CREATE INDEX idx_devices_status ON devices(status);
CREATE INDEX idx_devices_fleet_id ON devices(fleet_id);
CREATE INDEX idx_devices_labels ON devices USING GIN(labels);
CREATE INDEX idx_devices_last_heartbeat ON devices(last_heartbeat);
