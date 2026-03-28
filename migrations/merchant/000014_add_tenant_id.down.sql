DROP INDEX IF EXISTS idx_merchants_tenant;
ALTER TABLE merchants DROP COLUMN IF EXISTS tenant_id;
