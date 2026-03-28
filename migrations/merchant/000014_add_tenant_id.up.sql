-- Add tenant_id for white-label multi-tenant isolation
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS tenant_id UUID;
CREATE INDEX IF NOT EXISTS idx_merchants_tenant ON merchants(tenant_id);
