DROP INDEX IF EXISTS idx_organizations_api_key;
ALTER TABLE organizations DROP COLUMN IF EXISTS api_key;
