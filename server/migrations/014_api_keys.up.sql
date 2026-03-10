-- Add API key to organizations for agent authentication
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS api_key VARCHAR(64) UNIQUE;

-- Create index for fast lookups
CREATE INDEX IF NOT EXISTS idx_organizations_api_key ON organizations (api_key);
