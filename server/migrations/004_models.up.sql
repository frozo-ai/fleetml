CREATE TABLE models (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    version         VARCHAR(100) NOT NULL,
    format          VARCHAR(50) NOT NULL,

    -- Storage
    artifact_url    TEXT NOT NULL,
    artifact_size   BIGINT NOT NULL,
    checksum        VARCHAR(255) NOT NULL,

    -- Metadata
    description     TEXT,
    metadata        JSONB DEFAULT '{}',
    tags            TEXT[],

    -- Lineage
    parent_model_id UUID REFERENCES models(id),

    -- Compiled variants
    compiled_variants JSONB DEFAULT '[]',

    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by      UUID REFERENCES users(id),

    UNIQUE(name, version)
);

CREATE INDEX idx_models_name ON models(name);
CREATE INDEX idx_models_name_version ON models(name, version);
