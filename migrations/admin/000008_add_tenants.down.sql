DROP INDEX IF EXISTS idx_merchants_tenant;
ALTER TABLE merchants DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_tenants_domain;
DROP INDEX IF EXISTS idx_tenants_slug;
DROP TABLE IF EXISTS tenants;
