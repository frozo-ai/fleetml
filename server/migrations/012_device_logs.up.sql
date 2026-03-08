CREATE TABLE IF NOT EXISTS device_logs (
    id          BIGSERIAL PRIMARY KEY,
    device_id   TEXT NOT NULL,
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    level       TEXT NOT NULL DEFAULT 'info',
    component   TEXT NOT NULL DEFAULT 'agent',
    message     TEXT NOT NULL,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_device_logs_device_id ON device_logs (device_id);
CREATE INDEX idx_device_logs_timestamp ON device_logs (timestamp DESC);
CREATE INDEX idx_device_logs_device_time ON device_logs (device_id, timestamp DESC);
CREATE INDEX idx_device_logs_level ON device_logs (level);
